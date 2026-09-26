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
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

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
var graphPreviewDir string
var graphPreviewConfig *configfile.Config

func init() {
	rootCmd.PersistentFlags().String("graph-mode", "", "Assert workspace format: dependency or link (init selects format)")
	initCmd.Flags().String("scope-url", "", "Permanent operator-selected Scope URL for a fresh disposable graph preview")
	rememberCmd.Flags().String("id", "", "New canonical beads/PATH in a graph preview")
	rememberCmd.Flags().String("title", "", "Title of a new graph Memory")
	statusCmd.Flags().Bool("graph", false, "Report graph preview capabilities")
	showCmd.Flags().String("version", "", "Read an exact retained version token (graph preview only)")
	linkCmd.Flags().String("resource-type", "", "Installed experimental Link Type URL (graph preview only)")
	linkCmd.Flags().String("id", "", "New canonical links/PATH for an informational graph Link")
	linkCmd.Flags().String("properties", "", "Informational Link properties as JSON, @file, or @- (graph preview only)")
	updateCmd.Flags().String("properties", "", "Replace informational Link properties from JSON, @file, or @- (graph preview only)")
	updateCmd.Flags().String("if-revision", "", "Require this observed experimental Link revision")
	updateCmd.Flags().Bool("unconditional", false, "Explicitly accept the current experimental Link state")
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
	if jsonOutput {
		_ = json.NewEncoder(os.Stderr).Encode(map[string]any{"code": code, "message": message, "retryable": false})
	} else {
		fmt.Fprintf(os.Stderr, "%s: %s\n", code, message)
	}
	return &exitError{Code: exit}
}

func admitGraphPreview(cmd *cobra.Command) (bool, error) {
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
		if cmd == showCmd && cmd.Flags().Changed("version") {
			return true, graphFailure("capability_unavailable", "--version requires an experimental graph workspace", 5)
		}
		if cmd == graphUnlinkCmd || cmd == graphLinksCmd {
			return true, graphFailure("capability_unavailable", "this command requires an experimental graph workspace", 5)
		}
		if cmd == linkCmd && (cmd.Flags().Changed("resource-type") || cmd.Flags().Changed("id") || cmd.Flags().Changed("properties") || cmd.Flags().Changed("if-source-revision") || cmd.Flags().Changed("unconditional-source")) {
			return true, graphFailure("capability_unavailable", "generic Link options require a workspace initialized with graph_mode link", 5)
		}
		if cmd == updateCmd && (cmd.Flags().Changed("properties") || cmd.Flags().Changed("if-revision") || cmd.Flags().Changed("unconditional") || cmd.Flags().Changed("if-source-revision") || cmd.Flags().Changed("unconditional-source")) {
			return true, graphFailure("capability_unavailable", "generic update options require a workspace initialized with graph_mode link", 5)
		}
		if (cmd == rememberCmd && (cmd.Flags().Changed("id") || cmd.Flags().Changed("title"))) || (cmd == initCmd && cmd.Flags().Changed("scope-url")) {
			return true, graphFailure("capability_unavailable", "these graph options require a workspace initialized with graph_mode link", 5)
		}
		return false, nil
	}
	if err := requireDoltBackend(cfg); err != nil {
		return true, graphFailure("graph_not_initialized", "graph_mode link: "+err.Error(), 5)
	}
	if cfg == nil || string(marker) != graphPreviewGeneration || !cfg.GraphReady || cfg.GraphSchemaVersion != graphstore.SchemaVersion || cfg.GraphScopeURL == "" || cfg.GraphAuthorityID == "" || cfg.GraphWorkspace == "" || cfg.DoltDatabase == "" {
		return true, graphFailure("graph_not_initialized", "graph_mode link metadata is missing, incomplete, or unsupported; no database was opened", 5)
	}
	real, err := filepath.EvalSymlinks(dir)
	if err != nil || real != cfg.GraphWorkspace {
		return true, graphFailure("not_authority", "graph_mode workspace binding differs; copied/moved workspaces cannot claim this authority", 5)
	}
	if cmd != rememberCmd && cmd != createCmd && cmd != showCmd && cmd != statusCmd && cmd != depAddCmd && cmd != linkCmd && cmd != closeCmd && cmd != readyCmd && cmd != updateCmd && cmd != graphUnlinkCmd && cmd != graphLinksCmd && cmd != serveCmd {
		return true, graphFailure("capability_unavailable", "graph_mode link preview supports create, remember, show, dep add, link, update links/PATH, unlink, links, close, ready, status --graph, and shared-server serve; this command has not opened the legacy store", 5)
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
	if migration.IsFrozen(findTownRoot()) {
		return graphFailure("permission_denied", "town is frozen for migration; graph writes are blocked", 5)
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
	return graphPrint(map[string]any{"scope": scope, "backend": cfg.DoltMode, "preview": true, "memoryComplete": false}, "Initialized disposable graph preview; Memory, Issue and Link operations are experimental. No migration or recovery compatibility is promised.", quiet)
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
	return graphPrint(result, human, quietFlag)
}

func runGraphPreviewRemember(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "id", "title"); err != nil {
		return err
	}
	if readonlyMode {
		return graphFailure("permission_denied", "read-only invocation cannot remember", 5)
	}
	path, _ := cmd.Flags().GetString("id")
	title, _ := cmd.Flags().GetString("title")
	if path == "" {
		return graphFailure("invalid_selector", "this preview requires an explicit --id beads/PATH", 2)
	}
	if err := graph.ValidateBeadPath(path); err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if len(args) != 1 {
		return graphFailure("invalid_properties", "remember requires exactly one body argument", 2)
	}
	if strings.TrimSpace(title) == "" {
		return graphFailure("invalid_properties", "this preview requires an explicit nonempty --title", 2)
	}
	return withGraphStore(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		r, err := s.Create(ctx, graphstore.CreateRequest{Path: path, Title: title, Body: args[0], Actor: getActorWithGit()})
		if err != nil {
			return nil, "", err
		}
		return r, fmt.Sprintf("Created %s\n", path), nil
	})
}

func runGraphPreviewShow(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "version"); err != nil {
		return err
	}
	if len(args) != 1 {
		return graphFailure("invalid_selector", "graph show requires one canonical beads/PATH or links/PATH", 2)
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	version, _ := cmd.Flags().GetString("version")
	versioned := cmd.Flags().Changed("version")
	if versioned && (version == "" || !utf8.ValidString(version) || len(version) > graphstore.PreviewVersionTokenLimit) {
		return graphFailure("invalid_selector", "--version requires a nonempty UTF-8 token of at most 4096 bytes", 2)
	}
	return withGraphStore(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		var r any
		var err error
		if versioned {
			r, err = s.ReadVersion(ctx, path, version)
		} else {
			r, err = s.Read(ctx, path)
		}
		if err != nil {
			return nil, "", err
		}
		// Human output includes the complete record in the deliberately small preview.
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
		return map[string]any{"scope": graphPreviewConfig.GraphScopeURL, "backend": graphPreviewConfig.DoltMode, "preview": true, "limits": map[string]int{"issueOwnedLinks": graphstore.PreviewOwnedLinkLimit, "memoryOwnedLinks": graphstore.PreviewOwnedLinkLimit, "linkPropertiesInputBytes": graphPreviewPropertiesLimit, "incidentLinks": graphstore.PreviewIncidentLinkLimit, "currentReadBytes": graphstore.PreviewCurrentReadByteLimit, "versionTokenBytes": graphstore.PreviewVersionTokenLimit}, "capabilities": map[string]bool{"memoryCreate": true, "issueCreate": true, "issueWorkflows": false, "blockingDependency": true, "informationalLink": true, "linkPropertiesUpdate": true, "linkUnlink": true, "blockingDependencyUnlink": false, "incidentLinks": true, "issueClose": true, "issueReady": true, "genericRead": true, "bdpRead": graphPreviewConfig.DoltMode == configfile.DoltModeServer, "memory": false, "historyExact": false, "exactVersionRead": true, "ownedLinks": true, "requestStatus": false, "backupContinuity": false}}, "Graph preview: Memory and Issue create/read, exact retained Bead/Link reads via show --version TOKEN, local blocking Dependencies, informational Links and guarded Link-property replacement/unlink, incident Link listing, Issue close and ready. Full Memory, public History, blocking Dependency unlink, remaining Issue workflows and recovery remain unavailable. BDP Read serving is available only on ordinary shared-server Dolt; embedded serving, HTTP writes and aliases remain unavailable.", nil
	})
}

func graphPrint(result any, human string, quiet bool) error {
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"schemaVersion": 1, "preview": true, "result": result})
	}
	if !quiet {
		fmt.Fprintln(os.Stdout, human)
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
	case errors.Is(err, graphstore.ErrVersionUnknown):
		return graphFailure("revision_unknown", err.Error(), 3)
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
	case errors.Is(err, storage.ErrCloseBlocked), errors.Is(err, storage.ErrCloseOpenChildren):
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
