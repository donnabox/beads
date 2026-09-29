package main

// The disposable C0 preview is an explicit subset of the graph proposal. Its
// admission runs before legacy opening: unsupported commands cannot accidentally
// mutate Issue tables outside canonical graph transactions.
import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/config"
	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/migration"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

const graphPreviewMarker = "graph-preview-format"
const graphPreviewGeneration = "link-preview-v5\n"

var graphPreviewActive bool
var graphPreviewStructuredErrors bool
var graphPreviewDir string
var graphPreviewConfig *configfile.Config

func init() {
	rootCmd.PersistentFlags().String("graph-mode", "", "Assert workspace format: dependency or link (init selects format)")
	initCmd.Flags().String("scope-url", "", "Permanent operator-selected Scope URL for a fresh disposable graph preview")
	rememberCmd.Flags().String("id", "", "New canonical beads/PATH in a graph preview")
	rememberCmd.Flags().String("title", "", "Memory title (required on create; --update preserves omitted title/body; graph preview only)")
	rememberCmd.Flags().String("update", "", "Existing canonical Memory selector to update (graph preview only)")
	rememberCmd.Flags().String("if-revision", "", "Require this observed Memory revision for --update (graph preview only)")
	rememberCmd.Flags().Bool("unconditional", false, "Replace only supplied Memory fields without a revision guard; requires --update (graph preview only)")
	rememberCmd.Flags().String("body-file", "", "Read graph Memory body from this UTF-8 file (preview: at most 1 MiB)")
	rememberCmd.Flags().Bool("stdin", false, "Read graph Memory body from stdin (preview: at most 1 MiB)")
	statusCmd.Flags().Bool("graph", false, "Report graph preview capabilities")
	linkCmd.Flags().String("resource-type", "", "Installed experimental Link Type URL (graph preview only)")
	linkCmd.Flags().String("id", "", "New canonical links/PATH for an informational graph Link")
	linkCmd.Flags().String("properties", "", "Informational Link properties as JSON, @file, or @- (graph preview only)")
	updateCmd.Flags().String("properties", "", "Replace Memory or informational Link properties from JSON, @file, or @- (graph preview only)")
	updateCmd.Flags().String("if-revision", "", "Require this observed experimental Resource revision")
	updateCmd.Flags().Bool("unconditional", false, "Explicitly accept the current experimental Resource state")
	updateCmd.Flags().String("if-source-revision", "", "Require this observed source revision for an experimental owned Link")
	updateCmd.Flags().Bool("unconditional-source", false, "Explicitly accept the current source state for an experimental owned Link")
	linkCmd.Flags().String("if-source-revision", "", "Require this observed source revision for an experimental owned Link")
	linkCmd.Flags().Bool("unconditional-source", false, "Explicitly accept the current source state for an experimental owned Link")
}

// No home-directory or other repository fallback: an incomplete local graph
// workspace must never resolve to an unrelated store. Existing legacy discovery
// is unchanged when this gate does not find a graph workspace.
func graphCandidateDir(cmd *cobra.Command) (string, error) {
	if cmd.Root().PersistentFlags().Changed("db") || os.Getenv("BEADS_DB") != "" || os.Getenv("BD_DB") != "" {
		return selectedNoDBBeadsDir(cmd), nil
	}
	if d := os.Getenv("BEADS_DIR"); d != "" {
		return filepath.Abs(d)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if cmd == initCmd {
		return filepath.Join(cwd, ".beads"), nil
	}
	for d := cwd; ; d = filepath.Dir(d) {
		candidate := filepath.Join(d, ".beads")
		if _, err := os.Lstat(candidate); err == nil {
			return candidate, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		if _, err := os.Lstat(filepath.Join(d, ".git")); err == nil {
			break
		}
		if filepath.Dir(d) == d {
			break
		}
	}
	return "", nil
}

func graphFailure(code, message string, exit int) error {
	if jsonOutput || graphPreviewStructuredErrors {
		_ = json.NewEncoder(os.Stderr).Encode(map[string]any{"code": code, "message": message, "retryable": false})
	} else {
		fmt.Fprintf(os.Stderr, "%s: %s\n", code, message) //nolint:gosec // G705: stderr, not a browser context
	}
	return &exitError{Code: exit}
}

func admitGraphPreview(cmd *cobra.Command) (handled bool, admissionErr error) {
	defer func() {
		if handled {
			// Graph refusals already emit a complete diagnostic. Suppress Cobra's
			// duplicate error and usage output, including for deferred commands.
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
		}
	}()
	// Reset graph admission state before every command. C0 uses ordinary --json
	// output selection; later collection formats are not activated.
	graphPreviewStructuredErrors = false
	graphPreviewActive, graphPreviewConfig, graphPreviewDir = false, nil, ""
	flag, _ := cmd.Flags().GetString("graph-mode")
	env := os.Getenv("BD_GRAPH_MODE")
	if flag != "" && env != "" && flag != env {
		return true, graphFailure("invalid_selector", "--graph-mode and BD_GRAPH_MODE disagree", 2)
	}
	requested := flag
	if requested == "" {
		requested = env
	}
	if requested != "" && requested != "link" && requested != "dependency" {
		return true, graphFailure("invalid_selector", "unknown graph_mode; expected dependency or link", 2)
	}
	dir, err := graphCandidateDir(cmd)
	if err != nil {
		return true, graphFailure("graph_not_initialized", err.Error(), 5)
	}
	var cfg *configfile.Config
	var marker []byte
	markerPresent := false
	if dir != "" {
		marker, err = os.ReadFile(filepath.Join(dir, graphPreviewMarker)) // #nosec G304 -- fixed format sentinel in the explicitly selected workspace
		markerPresent = err == nil
		if err != nil && !os.IsNotExist(err) {
			return true, graphFailure("graph_not_initialized", err.Error(), 5)
		}
		cfg, err = configfile.LoadForDiscovery(dir)
		if err != nil {
			// Without a graph sentinel or assertion, preserve the legacy
			// command's corrupt-metadata refusal and its diagnostic context.
			// A damaged graph workspace must still fail before legacy opening.
			if !markerPresent && requested != "link" {
				return false, nil
			}
			return true, graphFailure("graph_not_initialized", err.Error(), 5)
		}
	}
	mode, err := cfg.GetGraphMode()
	if err != nil {
		return true, graphFailure("capability_unavailable", err.Error(), 5)
	}
	if markerPresent && (mode != "link" || string(marker) != graphPreviewGeneration) {
		return true, graphFailure("graph_not_initialized", "graph_mode marker and metadata disagree; automatic recovery is not supported", 5)
	}
	if cmd == initCmd && requested == "link" {
		if dir == "" {
			return true, graphFailure("graph_not_initialized", "no selected workspace", 5)
		}
		if _, err := os.Lstat(dir); !os.IsNotExist(err) {
			return true, graphFailure("graph_not_initialized", "graph_mode link initialization requires a fresh .beads directory; existing or incomplete stores are never replaced", 5)
		}
		graphPreviewActive, graphPreviewDir = true, dir
		return true, configureGraphPreview(cmd)
	}
	if requested != "" && requested != mode {
		return true, graphFailure("not_authority", "graph_mode assertion does not match persisted workspace format", 5)
	}
	if mode != "link" {
		if cmd == graphUnlinkCmd || cmd == graphLinksCmd {
			return true, graphFailure("capability_unavailable", "this command requires an experimental graph workspace", 5)
		}
		if cmd == linkCmd && (cmd.Flags().Changed("resource-type") || cmd.Flags().Changed("id") || cmd.Flags().Changed("properties") || cmd.Flags().Changed("if-source-revision") || cmd.Flags().Changed("unconditional-source")) {
			return true, graphFailure("capability_unavailable", "generic Link options require a workspace initialized with graph_mode link", 5)
		}
		if cmd == updateCmd && (cmd.Flags().Changed("properties") || cmd.Flags().Changed("if-revision") || cmd.Flags().Changed("unconditional") || cmd.Flags().Changed("if-source-revision") || cmd.Flags().Changed("unconditional-source")) {
			return true, graphFailure("capability_unavailable", "generic update options require a workspace initialized with graph_mode link", 5)
		}
		if (cmd == rememberCmd && rememberGraphFlagsChanged(cmd)) || (cmd == initCmd && cmd.Flags().Changed("scope-url")) || (cmd == statusCmd && cmd.Flags().Changed("graph")) {
			return true, graphFailure("capability_unavailable", "these graph options require a workspace initialized with graph_mode link", 5)
		}
		return false, nil
	}
	if err := requireDoltBackend(cfg, dir); err != nil {
		return true, graphFailure("graph_not_initialized", "graph_mode link: "+err.Error(), 5)
	}
	if cfg == nil || string(marker) != graphPreviewGeneration || !cfg.GraphReady || cfg.GraphSchemaVersion != graphstore.SchemaVersion || cfg.GraphScopeURL == "" || cfg.GraphAuthorityID == "" || cfg.GraphWorkspace == "" || cfg.DoltDatabase == "" {
		return true, graphFailure("graph_not_initialized", "graph_mode link metadata is missing, incomplete, or unsupported; no database was opened", 5)
	}
	real, err := filepath.EvalSymlinks(dir)
	if err != nil || real != cfg.GraphWorkspace {
		return true, graphFailure("not_authority", "graph_mode workspace binding differs; copied/moved workspaces cannot claim this authority", 5)
	}
	if cmd != rememberCmd && cmd != createCmd && cmd != showCmd && cmd != statusCmd && cmd != depAddCmd && cmd != linkCmd && cmd != closeCmd && cmd != reopenCmd && cmd != readyCmd && cmd != updateCmd && cmd != graphUnlinkCmd && cmd != graphLinksCmd {
		return true, graphFailure("capability_unavailable", "this graph preview supports remember, create, show, update, dep add, link, links, unlink, close, reopen, ready and status --graph; this command has not opened the legacy store", 5)
	}
	if cmd == statusCmd {
		enabled, _ := cmd.Flags().GetBool("graph")
		if !enabled {
			return true, graphFailure("capability_unavailable", "use status --graph for the graph preview", 5)
		}
	}
	graphPreviewActive, graphPreviewConfig, graphPreviewDir = true, cfg, dir
	if err := configureGraphPreview(cmd); err != nil {
		return true, err
	}
	if cfg.DoltMode == configfile.DoltModeServer {
		cfg.DoltServerUser = cfg.GetDoltServerUser()
	}
	return true, validateGraphPreviewRoute(cfg)
}

// Refuse flags whose semantics this slice does not implement, rather than
// silently pretending a legacy option was honored by the generic route.
func graphPreviewFlags(cmd *cobra.Command, allowed ...string) error {
	set := map[string]bool{"json": true, "graph-mode": true, "actor": true, "quiet": true, "no-color": true, "directory": true, "readonly": true}
	for _, name := range allowed {
		set[name] = true
	}
	var bad string
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if !set[f.Name] && bad == "" {
			bad = f.Name
		}
	})
	if bad != "" {
		return graphFailure("capability_unavailable", "graph preview does not implement --"+bad, 5)
	}
	return nil
}

// Configuration is bound during admission, before any graph route opens storage.
func graphPreviewWritePolicy() error {
	if readonlyMode {
		return graphFailure("permission_denied", "read-only invocation cannot mutate graph state", 5)
	}
	// Use the target's vendor-neutral lookup against both the caller and the
	// selected graph workspace. Its environment override and fail-closed lookup
	// errors match ordinary command admission; no legacy discovery is needed.
	if freeze := migration.Find(graphPreviewDir); freeze.Frozen() {
		if freeze.Err != nil {
			return graphFailure("permission_denied", "cannot determine migration freeze state; graph writes are blocked: "+freeze.Err.Error(), 5)
		}
		return graphFailure("permission_denied", "workspace is frozen for migration; graph writes are blocked by "+freeze.Path, 5)
	}
	return nil
}

func runGraphPreviewInit(cmd *cobra.Command) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "scope-url", "prefix", "server", "external", "server-host", "server-port", "server-user", "server-socket", "server-tls", "database", "skip-hooks", "skip-agents", "non-interactive"); err != nil {
		return err
	}
	scope, _ := cmd.Flags().GetString("scope-url")
	normalized, err := graph.NormalizeScopeURL(scope)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	scope = normalized
	if readonlyMode {
		return graphFailure("permission_denied", "read-only invocation cannot initialize", 5)
	}
	if cmd.Root().PersistentFlags().Changed("db") || os.Getenv("BEADS_DB") != "" || os.Getenv("BD_DB") != "" || os.Getenv("BEADS_DIR") != "" {
		return graphFailure("capability_unavailable", "graph preview init requires the current workspace without database-directory overrides", 5)
	}

	server, _ := cmd.Flags().GetBool("server")
	external, _ := cmd.Flags().GetBool("external")
	if server != external {
		return graphFailure("capability_unavailable", "shared-server preview requires --server --external; embedded uses neither", 5)
	}
	if !server {
		for _, name := range []string{"server-host", "server-port", "server-user", "server-socket", "server-tls"} {
			if cmd.Flags().Changed(name) {
				return graphFailure("invalid_selector", "--"+name+" requires --server --external", 2)
			}
		}
	}
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		return err
	}
	id := hex.EncodeToString(token)
	cfg := &configfile.Config{Backend: "dolt", GraphMode: "link", GraphScopeURL: scope, GraphAuthorityID: id, GraphSchemaVersion: graphstore.SchemaVersion, DoltMode: "embedded", DoltDatabase: "beads_graph_" + id, ProjectID: id}
	if server {
		cfg.DoltMode = "server"
		cfg.DoltServerHost, _ = cmd.Flags().GetString("server-host")
		if cfg.DoltServerHost == "" {
			cfg.DoltServerHost = "127.0.0.1"
		}
		cfg.DoltServerPort, _ = cmd.Flags().GetInt("server-port")
		if cfg.DoltServerPort == 0 {
			cfg.DoltServerPort = 3307
		}
		cfg.DoltServerUser, _ = cmd.Flags().GetString("server-user")
		if !cmd.Flags().Changed("server-user") {
			cfg.DoltServerUser = cfg.GetDoltServerUser()
		}
		if cfg.DoltServerUser == "" {
			cfg.DoltServerUser = "root"
		}
		cfg.DoltServerSocket, _ = cmd.Flags().GetString("server-socket")
		cfg.DoltServerTLS, _ = cmd.Flags().GetBool("server-tls")
	}
	if cmd.Flags().Changed("database") {
		cfg.DoltDatabase, _ = cmd.Flags().GetString("database")
	}
	if err := validateGraphPreviewRoute(cfg); err != nil {
		return err
	}
	if err := os.Mkdir(graphPreviewDir, 0o700); err != nil {
		return graphFailure("graph_not_initialized", err.Error(), 5)
	}
	real, err := filepath.EvalSymlinks(graphPreviewDir)
	if err != nil {
		return err
	}
	cfg.GraphWorkspace = real
	graphPreviewDir = real
	// Incomplete bootstrap remains fenced even if metadata is later lost. This is
	// a format sentinel, not another database or authoritative content copy.
	if err := os.WriteFile(filepath.Join(real, graphPreviewMarker), []byte(graphPreviewGeneration), 0o600); err != nil {
		return err
	}
	if err := cfg.Save(real); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(getRootContext(), 2*time.Minute)
	defer cancel()
	options := graphOptions(cfg)
	prefix, _ := cmd.Flags().GetString("prefix")
	if prefix == "" {
		prefix = config.GetString("issue-prefix")
	}
	if prefix == "" {
		prefix = filepath.Base(filepath.Dir(real))
	}
	options.IssuePrefix = normalizeIssuePrefix(prefix)
	if err := graphstore.Init(ctx, options); err != nil {
		return graphStorageError(err)
	}
	cfg.GraphReady = true
	if err := cfg.Save(real); err != nil {
		return graphFailure("graph_not_initialized", "database initialized but local readiness publication failed: "+err.Error(), 5)
	}
	graphPreviewConfig = cfg
	quiet, _ := cmd.Flags().GetBool("quiet")
	return graphPrint(map[string]any{"scope": scope, "backend": cfg.DoltMode, "preview": true, "memoryComplete": false}, "Initialized disposable mixed graph preview. Use status --graph for supported commands; no migration or recovery compatibility is promised.", quiet)
}

func graphOptions(cfg *configfile.Config) graphstore.Options {
	password := os.Getenv("BEADS_DOLT_PASSWORD")
	if cfg.DoltMode == configfile.DoltModeServer && password == "" {
		password = configfile.LookupCredentialsPassword(cfg.DoltServerHost, cfg.DoltServerPort)
	}
	return graphstore.Options{Backend: cfg.DoltMode, DataDir: filepath.Join(graphPreviewDir, "embeddeddolt"), Database: cfg.DoltDatabase, Branch: "main",
		Binding:    graphstore.Binding{WorkspaceID: cfg.GraphWorkspace, ScopeURL: cfg.GraphScopeURL, AuthorityID: cfg.GraphAuthorityID, SchemaVersion: cfg.GraphSchemaVersion},
		ServerHost: cfg.DoltServerHost, ServerPort: cfg.DoltServerPort, ServerUser: cfg.DoltServerUser, ServerPassword: password, ServerSocket: cfg.DoltServerSocket, ServerTLS: cfg.DoltServerTLS}
}

func withGraphStore(fn func(context.Context, *graphstore.Store) (any, string, error)) error {
	return withGraphStoreOutput(fn, func(result any, human string) error {
		return graphPrint(result, human, quietFlag)
	})
}

// Output starts only after the operation and ordinary store cleanup succeed.
func withGraphStoreOutput(fn func(context.Context, *graphstore.Store) (any, string, error), output func(any, string) error) error {
	ctx, cancel := context.WithTimeout(getRootContext(), 30*time.Second)
	defer cancel()
	s, err := graphstore.OpenExisting(ctx, graphOptions(graphPreviewConfig))
	if err != nil {
		return graphStorageError(err)
	}
	defer func() { _ = s.Close() }() // Panic fallback; ordinary cleanup is checked below.
	result, human, opErr := fn(ctx, s)
	closeErr := s.Close()
	if opErr != nil {
		return graphStorageError(errors.Join(opErr, closeErr))
	}
	if closeErr != nil {
		return graphFailure("route_unavailable", "operation completed but store cleanup failed: "+closeErr.Error(), 5)
	}
	return output(result, human)
}

func runGraphPreviewRemember(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "id", "title", "body-file", "stdin", "update", "if-revision", "unconditional"); err != nil {
		return err
	}
	if cmd.Flags().Changed("update") {
		if cmd.Flags().Changed("id") {
			return graphFailure("invalid_selector", "remember requires either --id for creation or --update for an existing Memory, not both", 2)
		}
		return runGraphPreviewRememberUpdate(cmd, args)
	}
	if cmd.Flags().Changed("if-revision") || cmd.Flags().Changed("unconditional") {
		return graphFailure("capability_unavailable", "remember write guards require --update; creation does not accept them", 5)
	}
	path, _ := cmd.Flags().GetString("id")
	title, _ := cmd.Flags().GetString("title")
	if path == "" {
		return graphFailure("invalid_selector", "this preview requires an explicit --id beads/PATH", 2)
	}
	if err := graph.ValidateBeadPath(path); err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if strings.TrimSpace(title) == "" {
		return graphFailure("invalid_properties", "this preview requires an explicit nonempty --title", 2)
	}
	body, err := graphPreviewRememberBody(cmd, args)
	if err != nil {
		return err
	}
	return withGraphStore(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		r, err := s.Create(ctx, graphstore.CreateRequest{Path: path, Title: title, Body: body, Actor: getActorWithGit()})
		if err != nil {
			return nil, "", err
		}
		return r, fmt.Sprintf("Created %s\n", path), nil
	})
}

func runGraphPreviewShow(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd); err != nil {
		return err
	}
	if len(args) != 1 {
		return graphFailure("invalid_selector", "graph show requires one canonical beads/PATH or links/PATH", 2)
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	return withGraphStore(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		r, err := s.Read(ctx, path)
		if err != nil {
			return nil, "", err
		}
		data, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return nil, "", err
		}
		return r, string(data), nil
	})
}

func runGraphPreviewStatus(cmd *cobra.Command) error {
	if err := graphPreviewFlags(cmd, "graph"); err != nil {
		return err
	}
	return withGraphStore(func(_ context.Context, _ *graphstore.Store) (any, string, error) {
		return map[string]any{"scope": graphPreviewConfig.GraphScopeURL, "backend": graphPreviewConfig.DoltMode, "preview": true,
				"limits": map[string]int{"memoryBodyInputBytes": graphPreviewMemoryBodyLimit, "memoryOwnedLinks": graphstore.PreviewOwnedLinkLimit,
					"issueOwnedLinks": graphstore.PreviewOwnedLinkLimit, "currentReadBytes": graphstore.PreviewCurrentReadByteLimit,
					"linkPropertiesInputBytes": graphPreviewPropertiesLimit, "memoryPropertiesInputBytes": graphPreviewPropertiesLimit,
					"incidentLinks": graphstore.PreviewIncidentLinkLimit, "versionTokenBytes": graphstore.PreviewVersionTokenLimit},
				"capabilities": map[string]bool{
					"memoryCreate": true, "memoryRead": true, "memoryBodyFileInput": true, "memoryBodyStdinInput": true,
					"memoryPropertiesUpdate": true, "memorySelectedUpdate": true, "memorySelectedUpdateUnconditional": true,
					"memoryOverwriteDisclosure": true, "issueCreate": true, "issueCreateAuthorship": true, "issueTextUpdate": true,
					"informationalLink": true, "blockingDependency": true, "linkPropertiesUpdate": true,
					"linkUnlink": true, "blockingDependencyUnlink": true, "incidentLinks": true, "ownedLinks": true,
					"issueClose": true, "issueReopen": true, "issueReady": true, "genericRead": true,
					"memory": false, "memoryDelete": false, "memoryPropertiesPatch": false, "linkPropertiesPatch": false,
					"issueList": false, "issueBlocked": false, "issueClaim": false, "issueWorkflows": false,
					"blockingDependencyPairUnlink": false, "bdpRead": false, "historyExact": false, "exactVersionRead": false,
					"requestStatus": false, "backupContinuity": false}},
			"Mixed graph preview: Memory create/read, guarded complete title/body replacement and selected remember updates, actual predecessor disclosure for unconditional Memory writes, basic Issue create/read with ordinary creator/owner defaults, guarded inline Issue title/description/design/acceptance edits, informational Links with property replacement and guarded unlink, blocking Dependencies with canonical-ID unlink, incident Links, and Issue close/reopen/ready. Full Memory, deletion, Issue list/blocked and later fields, ordered patches, public History, BDP serving, adoption and recovery remain unavailable.", nil
	})
}

func graphPrint(result any, human string, quiet bool) error {
	return graphPrintTo(os.Stdout, result, human, quiet, jsonOutput)
}

func graphPrintTo(out io.Writer, result any, human string, quiet, structured bool) error {
	if structured {
		return json.NewEncoder(out).Encode(map[string]any{"schemaVersion": 1, "preview": true, "result": result})
	}
	if !quiet {
		_, err := fmt.Fprintln(out, human)
		return err
	}
	return nil
}

func graphStorageError(err error) error {
	var ambiguous *graphstore.ErrAmbiguousLink
	if errors.As(err, &ambiguous) {
		if jsonOutput {
			_ = json.NewEncoder(os.Stderr).Encode(map[string]any{"code": "ambiguous_link", "message": err.Error(), "retryable": false, "candidateIDs": ambiguous.CandidateIDs})
			return &exitError{Code: 4}
		}
		return graphFailure("ambiguous_link", err.Error()+": "+strings.Join(ambiguous.CandidateIDs, ", "), 4)
	}
	switch {
	case errors.Is(err, graphstore.ErrGone):
		return graphFailure("gone", err.Error(), 3)
	case errors.Is(err, graphstore.ErrCapabilityUnavailable), errors.Is(err, graphstore.ErrLimitExceeded):
		return graphFailure("capability_unavailable", err.Error(), 5)
	case errors.Is(err, graphstore.ErrOutcomeUnknown):
		return graphFailure("outcome_unknown", err.Error()+"; do not automatically replay; inspect the canonical ID before deciding the next action", 6)
	case errors.Is(err, graph.ErrValidation):
		return graphFailure("invalid_properties", err.Error(), 2)
	case errors.Is(err, storage.ErrValidation):
		return graphFailure("invalid_properties", err.Error(), 2)
	case errors.Is(err, storage.ErrCloseBlocked), errors.Is(err, storage.ErrCloseOpenChildren), errors.Is(err, storage.ErrAlreadyClaimed), errors.Is(err, storage.ErrNotClaimable):
		return graphFailure("constraint_violation", err.Error(), 4)
	case errors.Is(err, graphstore.ErrAlreadyExists):
		return graphFailure("identity_reserved", err.Error(), 4)
	case errors.Is(err, graphstore.ErrConflict):
		return graphFailure("revision_conflict", err.Error(), 4)
	case errors.Is(err, graphstore.ErrNotFound):
		return graphFailure("not_found", err.Error(), 3)
	default:
		return graphFailure("graph_not_initialized", err.Error(), 5)
	}
}
