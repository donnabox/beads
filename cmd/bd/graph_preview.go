package main

// The disposable C0 preview is an explicit subset of the graph proposal. Its
// admission runs before legacy opening: unsupported commands cannot accidentally
// mutate Issue tables outside canonical graph transactions.
import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/steveyegge/beads/cmd/bd/setup"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/config"
	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/graphpatch"
	"github.com/steveyegge/beads/internal/memoryapi"
	"github.com/steveyegge/beads/internal/migration"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/workapi"
	publicops "github.com/steveyegge/beads/issueops"
)

const graphPreviewMarker = "graph-preview-format"

// The marker records the fresh workspace format. Earlier preview workspaces
// are not upgraded in place. Changing installed Types changes this marker.
const (
	graphPreviewGenerationPrefix = "link-preview-v"
	graphPreviewGeneration       = "link-preview-v11\n"
)

var graphPreviewSupportedGenerations = []string{graphPreviewGeneration}

var graphPreviewActive bool
var graphPreviewStructuredErrors bool
var graphPreviewDir string
var graphPreviewConfig *configfile.Config
var graphPreviewTypeGeneration int

func graphPreviewLinkNamespaceCommand(cmd *cobra.Command) bool {
	switch cmd {
	case linkCreateCmd, linkListCmd, linkShowCmd, linkUpdateCmd, linkDeleteCmd, linkVersionsCmd, linkCompareCmd, linkCountCmd:
		return true
	default:
		return false
	}
}

// Admission pins the installed Type generation before any graph command runs.
func graphPreviewIssueTypeURL(scope string) string {
	return graphstore.IssueTypeURL(scope)
}

func graphPreviewBlockingTypeURL(scope string) string {
	return graphstore.DependencyTypeURL(scope)
}

func init() {
	rootCmd.PersistentFlags().String("graph-mode", "", "Assert workspace format: dependency or link (init selects format)")
	initCmd.Flags().String("scope-url", "", "Permanent operator-selected Scope URL for a fresh disposable graph preview")
	rememberCmd.Flags().String("id", "", "Literal public Memory ID; creates or updates the same identity (--key is compatible; omitted ID derives from content in graph mode)")
	rememberCmd.Flags().String("title", "", "Memory title (defaults to a short body summary on create; updates preserve omitted fields; graph preview only)")
	rememberCmd.Flags().String("metadata", "", "Merge a JSON object into an existing Memory, or set initial metadata on create (graph preview only)")
	rememberCmd.Flags().String("replace-metadata", "", "Replace all Memory metadata with a JSON object on an existing-ID edit (graph preview only)")
	rememberCmd.Flags().String("properties", "", "Initialize or merge named Memory properties from a JSON object, @file, or @- (graph preview only)")
	rememberCmd.Flags().String("replace-properties", "", "Replace all properties of an existing Memory selected by --id or --update (graph preview only)")
	rememberCmd.Flags().StringArray("set-metadata", nil, "Set Memory metadata key=value (repeatable; graph preview only)")
	rememberCmd.Flags().StringArray("unset-metadata", nil, "Remove Memory metadata key (repeatable; graph preview only)")
	rememberCmd.Flags().String("update", "", "Existing canonical Memory selector to update (graph preview only)")
	rememberCmd.Flags().Bool("create-only", false, "Require a new Memory at --id or --key; refuse an allocated ID (graph preview only)")
	rememberCmd.Flags().String("if-revision", "", "Require this observed Memory revision for an existing-ID update (graph preview only)")
	rememberCmd.Flags().String("body-file", "", "Read graph Memory body from this UTF-8 file (preview: at most 1 MiB)")
	rememberCmd.Flags().Bool("stdin", false, "Read graph Memory body from stdin (preview: at most 1 MiB)")
	// delete already has upstream's --if-revision (as do update, close and
	// assign), which link mode reads through cmd.Flags(); registering it again
	// panics at init. forget has no upstream flag, so it keeps the preview's own.
	forgetCmd.Flags().String("if-revision", "", "Require this observed Memory revision (graph preview only)")
	statusCmd.Flags().Bool("graph", false, "Report graph preview capabilities")
	memoriesCmd.Flags().Bool("all", false, "Return all matching graph Memory summaries within preview bounds")
	memoriesCmd.Flags().Bool("details", false, "Include Memory versions, attribution and owned-Link counts")
	memoriesCmd.Flags().String("format", "table", "Graph Memory summaries: table or records-json (preview only)")
	recallCmd.Flags().String("version", "", "Recall an exact retained Memory version token (graph preview only)")
	showCmd.Flags().String("version", "", "Read an exact retained version token (graph preview only)")
	createCmd.Flags().String("properties", "", "Initialize installed Bead Type properties from a JSON object, @file, or @- (graph preview only)")
	linkCmd.Flags().String("id", "", "New informational Link ID or links/PATH (bare ID is shorthand for links/ID)")
	linkCmd.Flags().String("properties", "", "Informational Link properties as JSON, @file, or @- (graph preview only)")
	linkCmd.Flags().String("metadata", "", "Initial informational Link metadata as a JSON object or @file (graph preview only)")
	linkCmd.Flags().String("source-revision", "", "Source endpoint revision: exact token, current or unpinned; owned source pins refuse (distinct from --if-source-revision)")
	linkCmd.Flags().String("target-revision", "", "Target endpoint revision: exact token, current or unpinned (informational Links only)")
	updateCmd.Flags().String("source-revision", "", "Re-pin an informational Link source to an exact token or current, or unpin with unpinned; owned source pins refuse")
	updateCmd.Flags().String("target-revision", "", "Re-pin an informational Link target to an exact token or current, or unpin with unpinned")
	updateCmd.Flags().String("patch", "", "Apply ordered Issue, Memory or informational Link property operations from JSON, @file, or @- (graph preview only)")
	updateCmd.Flags().String("properties", "", "Shallowly merge named Issue, Memory or informational Link properties from JSON, @file, or @- (graph preview only)")
	updateCmd.Flags().String("replace-properties", "", "Replace all writable Issue, Memory or informational Link properties with a JSON object, @file, or @- (graph preview only)")
	updateCmd.Flags().String("replace-metadata", "", "Replace all Resource metadata with a JSON object or @file (graph preview only)")
	// update.go registers upstream's --if-revision (the decimal compare-and-swap
	// outside link mode); link mode reads that same flag as a graph token.
	updateCmd.Flags().String("if-source-revision", "", "Require this observed source revision for an experimental owned Link")
	linkCmd.Flags().String("if-source-revision", "", "Require this observed source revision for an experimental owned Link")
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

// graphFailureOutput, when set, receives graphFailure diagnostics instead of
// stderr. Only Claude hook admission sets it, to reword its refusals.
var graphFailureOutput io.Writer

// A canonical command may supply its error contract during workspace admission.
// The scoped caller restores this hook before command execution.
var graphFailureReporter func(code, message string, status int) error

func graphFailure(code, message string, exit int) error {
	if graphFailureReporter != nil {
		return graphFailureReporter(code, message, exit)
	}
	out := io.Writer(os.Stderr)
	if graphFailureOutput != nil {
		out = graphFailureOutput
	}
	if jsonOutput || graphPreviewStructuredErrors {
		_ = json.NewEncoder(out).Encode(map[string]any{"code": code, "message": message, "retryable": false})
	} else {
		fmt.Fprintf(out, "%s: %s\n", code, message) //nolint:gosec // G705: stderr, not a browser context
	}
	return &exitError{Code: exit}
}

// graphPreviewGenerationSupported is the one admission predicate. Both gates in
// admitGraphPreview use it, so a marker is never accepted at one and refused at
// the other.
func graphPreviewGenerationSupported(marker []byte) bool {
	return slices.Contains(graphPreviewSupportedGenerations, string(marker))
}

// graphPreviewGenerationDigits returns the generation number of a marker whose
// exact bytes are the generation prefix, a decimal number with no leading zero
// and one newline.
func graphPreviewGenerationDigits(marker []byte) (string, bool) {
	digits, ok := bytes.CutPrefix(marker, []byte(graphPreviewGenerationPrefix))
	if !ok {
		return "", false
	}
	digits, ok = bytes.CutSuffix(digits, []byte("\n"))
	if !ok || len(digits) == 0 || (digits[0] == '0' && len(digits) > 1) {
		return "", false
	}
	for _, c := range digits {
		if c < '0' || c > '9' {
			return "", false
		}
	}
	return string(digits), true
}

// graphPreviewGenerationNewer reports whether marker is a well-formed generation
// above the one this bd writes. A malformed marker is not newer: it keeps the
// generic refusal instead of earning a claim about a future format.
func graphPreviewGenerationNewer(marker []byte) bool {
	digits, ok := graphPreviewGenerationDigits(marker)
	if !ok {
		return false
	}
	current, _ := graphPreviewGenerationDigits([]byte(graphPreviewGeneration))
	// Without leading zeros a longer number is larger, and equal lengths order as text.
	return len(digits) > len(current) || (len(digits) == len(current) && digits > current)
}

func admitGraphPreview(cmd *cobra.Command) (handled bool, admissionErr error) {
	var hookDiagnostic *bytes.Buffer
	if cmd == claudeHookCmd {
		hookDiagnostic = new(bytes.Buffer)
		graphFailureOutput = hookDiagnostic
	}
	defer func() {
		if hookDiagnostic != nil {
			graphFailureOutput = nil
			if admissionErr != nil {
				handled, admissionErr = true, graphClaudeHookAdmissionWarning(hookDiagnostic.String(), admissionErr)
			}
		}
		if handled {
			// Graph refusals already emit a complete diagnostic. Suppress Cobra's
			// duplicate error and usage output, including for deferred commands.
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
		}
	}()
	// Explicit graph list and discovery formats select structured admission errors.
	// Reset this state for every command; ordinary output policy is unchanged.
	format, _ := cmd.Flags().GetString("format")
	graphPreviewStructuredErrors = (cmd == memoriesCmd && format == "records-json") || (cmd == listCmd && (format == "records-json" || strings.EqualFold(format, "json"))) || ((cmd == searchCmd || cmd == countCmd || cmd == queryCmd) && cmd.Flags().Changed("format"))
	graphPreviewActive, graphPreviewConfig, graphPreviewDir, graphPreviewTypeGeneration = false, nil, "", 0
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
				if (cmd == graphCmd && graphGenericFlagsChanged(cmd)) || (cmd == importCmd && cmd.Flags().Changed("source-prefix")) {
					return true, graphFailure("graph_not_initialized", err.Error(), 5)
				}
				return false, nil
			}
			return true, graphFailure("graph_not_initialized", err.Error(), 5)
		}
	}
	mode, err := cfg.GetGraphMode()
	if err != nil {
		return true, graphFailure("capability_unavailable", err.Error(), 5)
	}
	if markerPresent && graphPreviewGenerationNewer(marker) {
		return true, graphFailure("graph_not_initialized", fmt.Sprintf("graph_mode workspace format %s is newer than this bd supports (%s); upgrade bd; no database was opened",
			strings.TrimSuffix(string(marker), "\n"), strings.TrimSuffix(graphPreviewGeneration, "\n")), 5)
	}
	if markerPresent && (mode != "link" || !graphPreviewGenerationSupported(marker)) {
		return true, graphFailure("graph_not_initialized", "graph_mode marker and metadata disagree; automatic recovery is not supported", 5)
	}
	if cmd == initCmd && requested == "link" {
		if dir == "" {
			return true, graphFailure("graph_not_initialized", "no selected workspace", 5)
		}
		if _, err := os.Lstat(dir); !os.IsNotExist(err) {
			return true, graphFailure("graph_not_initialized", "graph_mode link initialization requires a fresh .beads directory; existing or incomplete stores are never replaced", 5)
		}
		graphPreviewActive, graphPreviewDir, graphPreviewTypeGeneration = true, dir, graphstore.TypeGeneration
		return true, configureGraphPreview(cmd)
	}
	if requested != "" && requested != mode {
		return true, graphFailure("not_authority", "graph_mode assertion does not match persisted workspace format", 5)
	}
	if mode != "link" {
		if cmd == importCmd && cmd.Flags().Changed("source-prefix") {
			return true, graphFailure("capability_unavailable", "--source-prefix requires a workspace initialized with graph_mode link", 5)
		}
		if cmd == doneTodoCmd && cmd.Flags().Changed("if-revision") {
			return true, graphFailure("capability_unavailable", "todo done --if-revision requires an experimental graph workspace", 5)
		}
		if (cmd == deferCmd || cmd == undeferCmd) && cmd.Flags().Changed("if-revision") {
			return true, graphFailure("capability_unavailable", "graph deferral guards require a workspace initialized with graph_mode link", 5)
		}
		if cmd == graphCmd && graphGenericFlagsChanged(cmd) {
			return true, graphFailure("capability_unavailable", "generic traversal options require an experimental graph workspace", 5)
		}
		if cmd == listCmd && format == "records-json" {
			return true, graphFailure("capability_unavailable", "records-json listing requires an experimental graph workspace", 5)
		}
		if cmd == listCmd && cmd.Flags().Changed("type") {
			selector, _ := cmd.Flags().GetString("type")
			if graphPreviewNominalSelector(selector) {
				return true, graphFailure("capability_unavailable", "graph Bead Types require a workspace initialized with graph_mode link", 5)
			}
		}
		// delete's --if-revision is upstream's compare-and-swap here, not a graph
		// guard: only forget's --if-revision is graph-only.
		if cmd == forgetCmd && cmd.Flags().Changed("if-revision") {
			return true, graphFailure("capability_unavailable", "Memory deletion guards require a workspace initialized with graph_mode link", 5)
		}
		if cmd == memoriesCmd && (cmd.Flags().Changed("all") || cmd.Flags().Changed("details") || (cmd.Flags().Changed("format") && !strings.EqualFold(format, "json"))) {
			return true, graphFailure("capability_unavailable", "Memory discovery options require an experimental graph workspace", 5)
		}
		if cmd == showCmd && cmd.Flags().Changed("format") {
			return true, graphFailure("capability_unavailable", "--format requires an experimental graph workspace", 5)
		}
		if (cmd == showCmd || cmd == recallCmd) && cmd.Flags().Changed("version") {
			return true, graphFailure("capability_unavailable", "--version requires an experimental graph workspace", 5)
		}
		if cmd == graphCompareCmd || graphPreviewLinkNamespaceCommand(cmd) || graphPreviewMemoryNamespaceCommand(cmd) {
			return true, graphFailure("capability_unavailable", "this command requires an experimental graph workspace", 5)
		}
		if cmd == createCmd && (cmd.Flags().Changed("properties") || graphPreviewNominalTypeFlagChanged(cmd)) {
			return true, graphFailure("capability_unavailable", "generic create options require a workspace initialized with graph_mode link", 5)
		}
		if cmd == linkCmd && (cmd.Flags().Changed("id") || cmd.Flags().Changed("properties") || cmd.Flags().Changed("metadata") || cmd.Flags().Changed("source-revision") || cmd.Flags().Changed("target-revision") || cmd.Flags().Changed("if-source-revision")) {
			return true, graphFailure("capability_unavailable", "generic Link options require a workspace initialized with graph_mode link", 5)
		}
		if cmd == depAddCmd && cmd.Flags().Changed("id") {
			return true, graphFailure("capability_unavailable", "--id requires a workspace initialized with graph_mode link", 5)
		}
		// --if-revision is absent on purpose: outside link mode it is upstream's
		// compare-and-swap on a decimal bead revision, not a graph option.
		if cmd == updateCmd && (graphPreviewLinkPinFlagsChanged(cmd) || cmd.Flags().Changed("patch") || cmd.Flags().Changed("properties") || cmd.Flags().Changed("if-source-revision")) {
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
	if cfg == nil || !graphPreviewGenerationSupported(marker) || !cfg.GraphReady || cfg.GraphSchemaVersion != graphstore.SchemaVersion || cfg.GraphScopeURL == "" || cfg.GraphAuthorityID == "" || cfg.GraphWorkspace == "" || cfg.DoltDatabase == "" {
		return true, graphFailure("graph_not_initialized", "graph_mode link metadata is missing, incomplete, or unsupported; no database was opened", 5)
	}
	real, err := filepath.EvalSymlinks(dir)
	if err != nil || real != cfg.GraphWorkspace {
		return true, graphFailure("not_authority", "graph_mode workspace binding differs; copied/moved workspaces cannot claim this authority", 5)
	}
	// These relationship commands are not complete graph CLI journeys yet.
	// Refuse before opening the store: a Type mapping alone does not make
	// epic projections or edge-plus-close History atomic. Native children
	// now has its own checked read route in childrenCmd.
	switch cmd {
	case duplicateCmd:
		return true, graphFailure("capability_unavailable", "graph duplicate requires one atomic duplicates edge and native close reason; this graph CLI does not provide that journey yet", 5)
	case supersedeCmd:
		return true, graphFailure("capability_unavailable", "graph supersede requires one atomic supersedes edge and native close reason; this graph CLI does not provide that journey yet", 5)
	}
	if !graphPreviewMemoryNamespaceCommand(cmd) && !graphPreviewLinkNamespaceCommand(cmd) && cmd != importCmd && cmd != setupCmd && cmd != claudeHookCmd && cmd != memoriesCmd && cmd != recallCmd && cmd != graphCompareCmd && cmd != listCmd && cmd != childrenCmd && cmd != epicStatusCmd && cmd != closeEligibleEpicsCmd && cmd != blockedCmd && cmd != graphCmd && cmd != graphCheckCmd && cmd != rememberCmd && cmd != createCmd && cmd != showCmd && cmd != statusCmd && cmd != stateCmd && cmd != stateListCmd && cmd != depCmd && cmd != depAddCmd && cmd != depRemoveCmd && cmd != depListCmd && cmd != depTreeCmd && cmd != depCyclesCmd && cmd != linkCmd && cmd != closeCmd && cmd != reopenCmd && cmd != unclaimCmd && cmd != heartbeatCmd && cmd != reclaimCmd && cmd != mergeSlotCreateCmd && cmd != mergeSlotCheckCmd && cmd != mergeSlotAcquireCmd && cmd != mergeSlotReleaseCmd && cmd != commentsCmd && cmd != deferCmd && cmd != undeferCmd && cmd != readyCmd && cmd != updateCmd && cmd != assignCmd && cmd != priorityCmd && cmd != noteCmd && cmd != todoCmd && cmd != addTodoCmd && cmd != listTodosCmd && cmd != doneTodoCmd && cmd != serveCmd && cmd != deleteCmd && cmd != forgetCmd && cmd != typesCmd && cmd != versionsCmd && cmd != historyCmd && cmd != searchCmd && cmd != countCmd && cmd != queryCmd && cmd != quickCmd && cmd != statusesCmd && cmd != commentCmd && cmd != commentsAddCmd && cmd != editCmd {
		// COUPLING: admitting versionsCmd and historyCmd here is only safe
		// because each has an early `if graphPreviewActive` dispatch to
		// runGraphPreviewVersions. Admission suppresses legacy store opening,
		// so admitting a command WITHOUT its dispatch makes it panic on a nil
		// store rather than refuse. See the note in history.go.
		return true, graphFailure("capability_unavailable", "this graph preview supports import of legacy data into fresh workspaces, remember, memories, recall, versions (and history as its alias here), compare, create, q, show, update, edit, assign, priority, note, comment, comments add, todo, delete, forget, dep add/remove/list/tree/cycles and dep --blocks, link and link create/list/show/update/delete/versions/compare/count, close, reopen, unclaim, heartbeat, reclaim, merge-slot create/check/acquire/release, state read, defer, undefer, ready, list/--format records-json and Issue-only list --json, children, search, count, query, blocked, graph ISSUE, graph check and graph --view generic, types, status and status --graph, statuses, project-local setup claude, claude-hook stop and shared-server serve; this command has not opened the legacy store", 5)
	}
	graphPreviewActive, graphPreviewConfig, graphPreviewDir = true, cfg, dir
	graphPreviewTypeGeneration = graphstore.TypeGeneration
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
	skipAgents, _ := cmd.Flags().GetBool("skip-agents")
	skipHooks, _ := cmd.Flags().GetBool("skip-hooks")
	workspace := filepath.Dir(graphPreviewDir)
	guidance, err := prepareGraphPreviewAgentInstructions(workspace, config.SafeAgentsFile(), skipAgents, graphPreviewInitAgentsRemedy)
	if err != nil {
		return graphFailure("graph_not_initialized", "agent guidance was not installed; no graph database was initialized: "+err.Error(), 5)
	}
	if !skipAgents && !skipHooks {
		if err := setup.PreflightGraphClaudeStop(workspace); err != nil {
			return graphFailure("graph_not_initialized", "Claude Stop hook preflight failed; no graph database was initialized: "+err.Error(), 5)
		}
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
	if err := guidance.install(); err != nil {
		return graphFailure("graph_not_initialized", "database initialized but agent guidance installation failed; workspace remains incomplete: "+err.Error(), 5)
	}
	if !skipAgents && !skipHooks {
		if err := setup.InstallGraphClaudeStopOnInit(workspace); err != nil {
			// Storage and guidance already exist. Refusing here would leave a
			// fenced workspace that neither init nor setup can repair, so treat
			// the hook as ordinary init treats Claude setup: warn and publish.
			fmt.Fprintf(os.Stderr, "Warning: graph workspace initialized without the Claude Stop hook: %v. Run `bd setup claude` after fixing this to add the Stop hook.\n", err) //nolint:gosec // G705: stderr, not a browser context
		}
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
		Binding:        graphstore.Binding{WorkspaceID: cfg.GraphWorkspace, ScopeURL: cfg.GraphScopeURL, AuthorityID: cfg.GraphAuthorityID, SchemaVersion: cfg.GraphSchemaVersion},
		TypeGeneration: graphPreviewTypeGeneration, AssessmentsEnabled: cfg.GraphAssessmentsEnabled,
		ServerHost: cfg.DoltServerHost, ServerPort: cfg.DoltServerPort, ServerUser: cfg.DoltServerUser, ServerPassword: password, ServerSocket: cfg.DoltServerSocket, ServerTLS: cfg.DoltServerTLS}
}

func withGraphStore(fn func(context.Context, *graphstore.Store) (any, string, error)) error {
	return withGraphStoreOutput(fn, func(result any, human string) error {
		return graphPrint(result, human, quietFlag)
	})
}

// Output starts only after the operation and ordinary store cleanup succeed.
func withGraphStoreOutput(fn func(context.Context, *graphstore.Store) (any, string, error), output func(any, string) error) error {
	return withGraphStoreBudget(30*time.Second, fn, output)
}

func withGraphStoreBudget(budget time.Duration, fn func(context.Context, *graphstore.Store) (any, string, error), output func(any, string) error) error {
	return withGraphStoreErrorMapper(budget, fn, output, graphStorageError)
}

// The mapper receives the raw open or joined operation/cleanup error before
// any diagnostic is emitted. Ordinary graph routes keep graphStorageError.
func withGraphStoreErrorMapper(budget time.Duration, fn func(context.Context, *graphstore.Store) (any, string, error), output func(any, string) error, mapError func(error) error) error {
	return withGraphStoreStageErrorMappers(budget, fn, output, mapError, mapError)
}

// Open failures and failures after a successful open can need different
// fallback categories. Recognized storage errors keep their shared mapping.
func withGraphStoreStageErrorMappers(budget time.Duration, fn func(context.Context, *graphstore.Store) (any, string, error), output func(any, string) error, mapOpenError, mapOperationError func(error) error) error {
	ctx, cancel := context.WithTimeout(getRootContext(), budget)
	defer cancel()
	s, err := graphstore.OpenExisting(ctx, graphOptions(graphPreviewConfig))
	if err != nil {
		return mapOpenError(err)
	}
	defer func() { _ = s.Close() }() // Panic fallback; ordinary cleanup is checked below.
	result, human, opErr := fn(ctx, s)
	closeErr := s.Close()
	if opErr != nil {
		return mapOperationError(errors.Join(opErr, closeErr))
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
	if err := graphPreviewFlags(cmd, "id", "key", "title", "body-file", "stdin", "update", "create-only", "if-revision", "metadata", "replace-metadata", "set-metadata", "unset-metadata", "properties", "replace-properties"); err != nil {
		return err
	}
	if cmd.Flags().Changed("id") && cmd.Flags().Changed("key") {
		return graphFailure("invalid_selector", "choose --id or --key, not both", 2)
	}
	if cmd.Flags().Changed("replace-properties") {
		return runGraphPreviewRememberPropertiesReplacement(cmd, args)
	}
	explicit := cmd.Flags().Changed("id") || cmd.Flags().Changed("key")
	if cmd.Flags().Changed("update") {
		if explicit {
			return graphFailure("invalid_selector", "choose --update or explicit --id/--key, not both", 2)
		}
		return runGraphPreviewRememberUpdate(cmd, args)
	}
	title, body, err := graphPreviewRememberPatchInput(cmd, args)
	if err != nil {
		return err
	}
	publicID := ""
	if explicit {
		name := "id"
		if cmd.Flags().Changed("key") {
			name = "key"
		}
		publicID, _ = cmd.Flags().GetString(name)
		if _, err := memoryapi.ValidateKey(publicID); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
	} else {
		if body == nil {
			return graphFailure("invalid_properties", "remember without an explicit ID requires content", 2)
		}
		if name, ok := matchesKnownCommand(cmd, *body); ok {
			return graphFailure("invalid_properties", fmt.Sprintf("%q looks like a command, not content; use an explicit --key to store it", name), 2)
		}
		publicID, err = memoryapi.ResolveKey("", *body)
		if err != nil {
			return graphFailure("invalid_properties", err.Error(), 2)
		}
		if publicID == *body {
			return withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
				path, err := store.ResolveExactAllocatedResourcePath(ctx, publicID)
				if err != nil {
					return nil, "", fmt.Errorf("%w: bare key-like content requires an explicit --key to write", err)
				}
				value, err := store.Show(ctx, path)
				return nil, value.Properties.Body, err
			}, func(_ any, body string) error { _, err := fmt.Fprintln(cmd.OutOrStdout(), body); return err })
		}
	}
	if cmd.Flags().Changed("create-only") {
		enabled, _ := cmd.Flags().GetBool("create-only")
		if !enabled || !explicit || cmd.Flags().Changed("if-revision") {
			return graphFailure("invalid_selector", "--create-only requires explicit --id/--key and cannot combine with write guards", 2)
		}
		if body == nil {
			return graphFailure("invalid_properties", "Memory creation requires a body", 2)
		}
		if title != nil && strings.TrimSpace(*title) == "" {
			return graphFailure("invalid_properties", "an explicit creation title must be nonempty", 2)
		}
		createTitle := graphPreviewMemoryTitle(*body)
		if title != nil {
			createTitle = *title
		}
		metadata, err := graphPreviewMetadataCreate(cmd)
		if err != nil {
			return err
		}
		return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
			r, err := store.Create(ctx, graphstore.CreateRequest{PublicID: publicID, Title: createTitle, Body: *body, Actor: getActorWithGit(), Metadata: metadata})
			return r, fmt.Sprintf("Created %s", publicID), err
		})
	}
	return runGraphPreviewRememberUpsert(cmd, publicID, title, body)
}

func runGraphPreviewShow(cmd *cobra.Command, args []string) error {
	return runGraphPreviewShowKind(cmd, args, false)
}

func runGraphPreviewShowKind(cmd *cobra.Command, args []string, linkOnly bool) error {
	if canonicalOutputSelected(cmd) {
		return runCanonicalShow(cmd, args, linkOnly)
	}
	if err := graphPreviewFlags(cmd, "id", "current", "refs", "children", "watch", "version", "format", "include-dependents", "include-comments", "brief-deps", "short", "long", "local-time", "comments-tail"); err != nil {
		return err
	}
	idFlags, _ := cmd.Flags().GetStringArray("id")
	args = append(args, idFlags...)
	currentMode, _ := cmd.Flags().GetBool("current")
	if currentMode {
		if len(args) != 0 {
			return graphFailure("invalid_selector", "--current cannot be combined with explicit Issue IDs", 2)
		}
		currentID, err := graphPreviewCurrentIssueID(getActorWithGit())
		if err != nil {
			return err
		}
		if currentID == "" {
			return graphFailure("not_found", "no current Issue found (no in-progress, hooked, or recently touched Issues)", 3)
		}
		args = []string{currentID}
	}
	shortMode, _ := cmd.Flags().GetBool("short")
	longMode, _ := cmd.Flags().GetBool("long")
	localTime, _ := cmd.Flags().GetBool("local-time")
	commentsTail, _ := cmd.Flags().GetInt("comments-tail")
	if err := validateCommentsTail(commentsTail); err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString("format")
	if format != "" && format != "graph-json" {
		return graphFailure("invalid_selector", "graph show --format must be graph-json", 2)
	}
	if format != "" && (cmd.Flags().Changed("include-dependents") || cmd.Flags().Changed("include-comments") || cmd.Flags().Changed("brief-deps")) {
		return graphFailure("capability_unavailable", "Issue detail flags require legacy Issue --json output", 5)
	}
	if format != "" && (cmd.Flags().Changed("short") || cmd.Flags().Changed("long") || cmd.Flags().Changed("local-time") || cmd.Flags().Changed("comments-tail")) {
		return graphFailure("capability_unavailable", "Issue display flags require the familiar Issue output format", 5)
	}
	watchMode, _ := cmd.Flags().GetBool("watch")
	if watchMode {
		return runGraphPreviewShowWatch(cmd, args, format, commentsTail)
	}
	showRefs, _ := cmd.Flags().GetBool("refs")
	showChildren, _ := cmd.Flags().GetBool("children")
	if showRefs || showChildren {
		return runGraphPreviewShowRelated(cmd, args, showRefs, shortMode, format)
	}
	if len(args) > 1 {
		return runGraphPreviewShowMany(cmd, args, shortMode, longMode, localTime, commentsTail, format)
	}
	if len(args) != 1 {
		return graphFailure("invalid_selector", "graph show requires one local public ID or canonical Scope URL; use a positional ID or --id", 2)
	}
	var literal string
	var err error
	if cmd.Flags().Changed("id") {
		literal, err = graphPreviewResolveLiteralArgument(args[0])
	} else {
		literal, err = graphPreviewResolveMemoryArgument(args[0])
	}
	if err != nil {
		return err
	}
	ref, err := graphPreviewParseReference(literal)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	version, _ := cmd.Flags().GetString("version")
	versioned := cmd.Flags().Changed("version")
	if ref.Explicit {
		if versioned {
			return graphFailure("invalid_selector", "choose ID@revision or --version, not both", 2)
		}
		if ref.Revision != "current" && ref.Revision != "unpinned" {
			version, versioned = ref.Revision, true
		}
	}
	if versioned && (cmd.Flags().Changed("include-dependents") || cmd.Flags().Changed("include-comments") || cmd.Flags().Changed("brief-deps")) {
		return graphFailure("capability_unavailable", "Issue detail flags cannot describe an exact retained version", 5)
	}
	if versioned && (cmd.Flags().Changed("short") || cmd.Flags().Changed("long") || cmd.Flags().Changed("local-time") || cmd.Flags().Changed("comments-tail")) {
		return graphFailure("capability_unavailable", "Issue display flags require a current Issue", 5)
	}
	if versioned && (version == "" || !utf8.ValidString(version) || len(version) > graphstore.PreviewVersionTokenLimit) {
		return graphFailure("invalid_selector", "--version requires a nonempty UTF-8 token of at most 4096 bytes", 2)
	}
	path, localID, err := graphPreviewResourceReadSelector(ref.Identity)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if versioned && localID != "" {
		return graphFailure("capability_unavailable", "retained versions require an exact canonical Scope URL", 5)
	}
	var shownIssue, legacyIssue, legacyShort bool
	err = withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		var r any
		var details *types.IssueDetails
		var err error
		if versioned {
			r, err = s.ReadVersion(ctx, path, version)
		} else {
			includeDependents, _ := cmd.Flags().GetBool("include-dependents")
			includeComments, _ := cmd.Flags().GetBool("include-comments")
			briefDeps, _ := cmd.Flags().GetBool("brief-deps")
			if !jsonOutput && format == "" {
				includeDependents, includeComments = true, true
			}
			path, r, details, err = s.ReadResourceWithIssueDetailsSelected(ctx, path, localID, workapi.DetailOptions{IncludeDependents: includeDependents, IncludeComments: includeComments, BriefDeps: briefDeps})
		}
		if err != nil {
			return nil, "", err
		}
		if linkOnly {
			if _, ok := r.(graphstore.LinkRecord); !ok {
				return nil, "", graphFailure("capability_unavailable", "link show requires a Link Resource", 5)
			}
		}
		_, shownIssue = r.(graphstore.IssueRecord)
		if shortMode && shownIssue && !versioned && format == "" {
			legacyShort = true
			issue := *r.(graphstore.IssueRecord).Properties
			issue.ID = r.(graphstore.IssueRecord).Properties.ID
			return r, formatShortIssue(&issue), nil
		}
		if shownIssue && jsonOutput && format == "" && !versioned {
			legacyIssue = true
			return []*types.IssueDetails{details}, "", nil
		}
		if shownIssue && !jsonOutput && format == "" && !versioned {
			return details, renderGraphIssueDetails(details, graphIssueDetailOptions{Long: longMode, LocalTime: localTime, CommentsTail: commentsTail}), nil
		}
		if !shownIssue && (cmd.Flags().Changed("include-dependents") || cmd.Flags().Changed("include-comments") || cmd.Flags().Changed("brief-deps") || cmd.Flags().Changed("short") || cmd.Flags().Changed("long") || cmd.Flags().Changed("local-time") || cmd.Flags().Changed("comments-tail")) {
			return nil, "", graphFailure("capability_unavailable", "Issue detail and display flags require an Issue", 5)
		}
		humanRecord := r
		if memory, ok := r.(graphstore.Record); ok && memory.PublicID != "" {
			memory.ID = memory.PublicID
			humanRecord = memory
		}
		projected, err := graphProjectCompleteRecords(humanRecord)
		if err != nil {
			return nil, "", err
		}
		data, err := json.MarshalIndent(projected, "", "  ")
		if err != nil {
			return nil, "", err
		}
		return r, string(data), nil
	}, func(result any, human string) error {
		if legacyShort {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), human)
			return err
		}
		if legacyIssue {
			return outputJSON(result)
		}
		if format == "graph-json" {
			return graphPrintTo(cmd.OutOrStdout(), result, human, quietFlag, true)
		}
		return graphPrint(result, human, quietFlag)
	})
	if err == nil && shownIssue {
		graphPreviewStoreLastTouchedPath(path)
	}
	return err
}

func runGraphPreviewStatus(cmd *cobra.Command) error {
	if err := graphPreviewFlags(cmd, "graph"); err != nil {
		return err
	}
	enabled, _ := cmd.Flags().GetBool("graph")
	if !enabled {
		return runGraphPreviewIssueStatus()
	}
	return withGraphStore(func(_ context.Context, _ *graphstore.Store) (any, string, error) {
		return map[string]any{"scope": graphPreviewConfig.GraphScopeURL, "backend": graphPreviewConfig.DoltMode, "preview": true,
				"limits": map[string]int{"issueListOutputBytes": graphIssueListOutputLimit, "beadListOutputBytes": graphIssueListOutputLimit,
					"beadListInventoryResources": graphstore.PreviewSnapshotLimit,
					"issueBlockedOutputBytes":    graphIssueBlockedOutputLimit, "issueBlockedInventoryResources": graphstore.PreviewSnapshotLimit,
					"genericTraversalOutputBytes": graphGenericOutputLimit, "genericTraversalDepth": graphGenericBound,
					"genericTraversalNodes": graphGenericBound, "genericTraversalLinks": graphGenericBound, "genericTraversalInventoryResources": graphstore.PreviewSnapshotLimit,
					"memoryBodyInputBytes": graphPreviewMemoryBodyLimit, "memoryOwnedLinks": graphstore.PreviewOwnedLinkLimit,
					"issueOwnedLinks": graphstore.PreviewOwnedLinkLimit, "currentReadBytes": graphstore.PreviewCurrentReadByteLimit,
					"linkPropertiesInputBytes": graphPreviewPropertiesLimit, "memoryPropertiesInputBytes": graphPreviewPropertiesLimit,
					"memoryPatchInputBytes":      graphpatch.MaxInputBytes,
					"memoryPatchOperations":      graphpatch.MaxOperations,
					"memoryPatchPointerBytes":    graphpatch.MaxPointerBytes,
					"memoryPatchDepth":           graphpatch.MaxDepth,
					"memoryPatchDocumentBytes":   graphpatch.MaxDocumentBytes,
					"memoryPatchEvaluationBytes": graphpatch.MaxEvaluationBytes,
					"memoryPatchPointerSegments": graphpatch.MaxPointerSegments,
					"linkPatchInputBytes":        graphpatch.MaxInputBytes,
					"linkPatchOperations":        graphpatch.MaxOperations,
					"linkPatchPointerBytes":      graphpatch.MaxPointerBytes,
					"linkPatchDepth":             graphpatch.MaxDepth,
					"linkPatchDocumentBytes":     graphpatch.MaxDocumentBytes,
					"linkPatchEvaluationBytes":   graphpatch.MaxEvaluationBytes,
					"linkPatchPointerSegments":   graphpatch.MaxPointerSegments,
					"incidentLinks":              graphstore.PreviewIncidentLinkLimit, "versionTokenBytes": graphstore.PreviewVersionTokenLimit,
					"memoryDiscoveryDefaultMatches": graphMemoryDiscoveryDefaultLimit, "memoryDiscoveryOutputBytes": graphMemoryDiscoveryOutputLimit,
					"memoryDiscoveryQueryBytes": graphMemoryDiscoveryQueryLimit, "memoryDiscoveryExcerptCodePoints": graphMemoryDiscoveryExcerptLimit},
				"capabilities": map[string]bool{
					"legacyImport": true, "graphImport": false, "graphExport": false,
					"memoryCreate": true, "memoryRead": true, "memoryBodyRecall": true, "memoryJSONRecall": false,
					"memoryDiscovery": true, "memoryDiscoveryPagination": false, "memoryBodyFileInput": true, "memoryBodyStdinInput": true,
					"memoryPropertiesUpdate": true, "memorySelectedUpdate": true, "memorySelectedUpdateCurrentByDefault": true,
					"memoryOverwriteDisclosure": true, "issueCreate": true, "issueCreateAuthorship": true, "issueTextUpdate": true, "issuePriorityUpdate": true, "issueAssigneeUpdate": true,
					"issueCreateFields": true, "issueInitialNotes": true, "issueNotesAppend": true, "issueNotesReplace": true, "issueNotesClear": true,
					"issueEstimateUpdate": true, "issueReferenceUpdate": true, "issuePropertiesUpdate": true, "issuePropertiesPatch": true,
					"memoryUnreferencedDelete": true, "issueUnreferencedDelete": true,
					"informationalLink": true, "blockingDependency": true, "blockingDependencyShorthand": true, "blockingDependencyExplicitID": true, "linkPropertiesUpdate": true, "linkEndpointPins": true, "linkEndpointRepin": true, "ownedSourceEndpointPins": false,
					"linkUnlink": true, "blockingDependencyUnlink": true, "incidentLinks": true, "ownedLinks": true,
					"issueClose": true, "issueReopen": true, "issueDatelessDeferral": true, "issueDatedDeferral": true, "issueReady": true, "genericRead": true,
					"memory": false, "memoryDelete": false, "memoryPropertiesPatch": true, "linkPropertiesPatch": true,
					"issueList": true, "beadList": true, "beadTypeFilter": true, "issueBlocked": true, "genericTraversal": true,
					"issueListTree": false, "issueListLegacyJSON": true, "issueAssigneeFilter": true, "issueDueDate": true, "issueDueFilter": true, "issueClaim": true, "issueUnclaim": true, "issueTodo": true, "issueHeartbeat": true, "issueReclaim": true, "issueStateRead": true, "issueStateSet": false, "issueStatus": true, "issueGate": false, "issueMergeSlot": true, "issueWorkflows": false,
					"blockingDependencyPairUnlink": false, "bdpRead": graphPreviewConfig.DoltMode == configfile.DoltModeServer, "historyExact": false, "versionList": true, "exactVersionRead": true, "exactVersionCompare": true,
					"requestStatus": false, "backupContinuity": false}},
			"Mixed graph preview: Memory create/read, typed Issue and Memory property initialization and shallow property merge with optional revision comparison and actual predecessor disclosure, unreferenced Memory and Issue deletion with read-only preview and retained identity/snapshots, Issue create/read including initial fields, notes, due date and ordinary creator/owner defaults, inline Issue title/description/design/acceptance, priority, estimate, external/spec references, due date and non-claim assignee edits with optional revision comparison, standalone Issue assign, priority and note shortcuts, native Issue comments, quick capture and editor-backed Issue edits, bounded Issue search/count/query with ordinary JSON, and task-Issue todo add/list/done, atomic Issue claims with five-minute leases, node-local history-free heartbeat, native stale-lease reclaim, native-policy unclaim (holder, force or conditional, with optional reason and multiple IDs), state-label read, durable Issue status summary and merge-slot create/check/acquire/release. Issue notes append, guarded replacement with explicit overwrite intent and deliberate clear, read-only comments, informational Links with local endpoint pins, same-ID re-pin, shallow property merge and guarded unlink (owned-source pins remain unavailable), blocking Dependencies with canonical-ID unlink or native pair-level dep remove, incident Links, and Issue close/reopen/defer/undefer/ready with dated wake are available. dep BLOCKER --blocks BLOCKED is available, and dep add accepts an explicit Link ID. Human blocking dep list/tree/cycles and flat graph check are available within their bounded snapshot and format guards. Bounded current all-Bead listing with nominal Bead Type filtering and Issue-only legacy list JSON are available; Issue-specific filters retain the native Issue query and due/assignee filters. Native immediate children has human and ordinary JSON views; tree formatting is not yet available. Fresh graph workspaces admit atomic create --parent with a native dotted child ID and parent-child Link, plus epic status and close-eligible (including dry-run) from native parent-child Links. Complete dependency-blocked inspection and bounded current typed Bead graph traversal with Link-Type unions/refinements, canonical JSON and ordinary output formats are available. Memory discovery returns complete bounded title/body search summaries; current and exact retained body-only recall, show --version, explicit-version compare and ordered local bd versions/bd history listing are available; common metadata creation, optional-revision update, exact retained read and comparison are available for admitted Beads and informational Links. Ordered Issue, Memory and informational Link property patches are available with optional resource/source comparisons. Full Memory, linked Memory deletion, gate, set-state, duplicate/supersede journeys, HTTP History and recovery remain unavailable. BDP Read serving is available only on ordinary shared-server Dolt; embedded serving, HTTP writes and aliases remain unavailable.", nil
	})
}

func graphPrint(result any, human string, quiet bool) error {
	return graphPrintTo(os.Stdout, result, human, quiet, jsonOutput) //nolint:forbidigo // Graph preview output predates the command-writer rule; withGraphStore callers have no command writer to route through yet.
}

func graphPrintTo(out io.Writer, result any, human string, quiet, structured bool) error {
	if structured {
		projected, err := graphProjectCompleteRecords(result)
		if err != nil {
			return err
		}
		return json.NewEncoder(out).Encode(map[string]any{"schemaVersion": 1, "preview": true, "result": projected})
	}
	if !quiet {
		_, err := fmt.Fprintln(out, human)
		return err
	}
	return nil
}

func graphStorageError(err error) error {
	return graphStorageErrorWithFallback(err, "graph_not_initialized")
}

func graphStorageErrorWithFallback(err error, fallback string) error {
	// A failed rollback may join uncertainty with a definite storage refusal.
	// No other classification may hide the inspect-before-retry boundary.
	if errors.Is(err, graphstore.ErrOutcomeUnknown) {
		return graphFailure("outcome_unknown", err.Error()+"; do not automatically replay; inspect the canonical ID before deciding the next action", 6)
	}
	// A graph route may validate a selected Type only after reading it inside
	// withGraphStore. graphFailure has already reported this typed CLI refusal.
	var reported *exitError
	if errors.As(err, &reported) {
		return reported
	}
	var ambiguous *graphstore.ErrAmbiguousLink
	if errors.As(err, &ambiguous) {
		if jsonOutput {
			_ = json.NewEncoder(os.Stderr).Encode(map[string]any{"code": "ambiguous_link", "message": err.Error(), "retryable": false, "candidateIDs": ambiguous.CandidateIDs})
			return &exitError{Code: 4}
		}
		return graphFailure("ambiguous_link", err.Error()+": "+strings.Join(ambiguous.CandidateIDs, ", "), 4)
	}
	switch {
	case errors.Is(err, errGraphListSelector):
		return graphFailure("invalid_selector", err.Error(), 2)
	case errors.Is(err, graphstore.ErrVersionUnknown):
		return graphFailure("revision_unknown", err.Error(), 3)
	case errors.Is(err, graphstore.ErrIncidentLinkConstraint):
		return graphFailure("constraint_violation", err.Error(), 4)
	case errors.Is(err, publicops.ErrDependentsOutsideRequest):
		return graphFailure("constraint_violation", err.Error(), 4)
	case errors.Is(err, graphstore.ErrGone):
		return graphFailure("gone", err.Error(), 3)
	case errors.Is(err, graphstore.ErrCapabilityUnavailable), errors.Is(err, graphstore.ErrLimitExceeded):
		return graphFailure("capability_unavailable", err.Error(), 5)
	case errors.Is(err, graphstore.ErrInvalidStore):
		return graphFailure("invalid_store", err.Error(), 5)
	case errors.Is(err, graph.ErrValidation):
		return graphFailure("invalid_properties", err.Error(), 2)
	case errors.Is(err, storage.ErrValidation), errors.Is(err, publicops.ErrValidation):
		return graphFailure("invalid_properties", err.Error(), 2)
	case errors.Is(err, publicops.ErrNotesOverwrite):
		return graphFailure("notes_overwrite_refused", err.Error()+"; use --force to replace existing notes or --append-notes to preserve them", 4)
	case errors.Is(err, storage.ErrCloseBlocked), errors.Is(err, storage.ErrCloseOpenChildren), errors.Is(err, storage.ErrAlreadyClaimed), errors.Is(err, storage.ErrNotClaimable), errors.Is(err, storage.ErrNotOwner), errors.Is(err, storage.ErrAssigneeMismatch), errors.Is(err, publicops.ErrNotClaimed), errors.Is(err, publicops.ErrNotReleasable):
		return graphFailure("constraint_violation", err.Error(), 4)
	case errors.Is(err, graphstore.ErrAlreadyExists), errors.Is(err, storage.ErrAlreadyExists), errors.Is(err, publicops.ErrAlreadyExists):
		return graphFailure("identity_reserved", err.Error(), 4)
	case errors.Is(err, graphstore.ErrConflict):
		return graphFailure("revision_conflict", err.Error(), 4)
	case errors.Is(err, graphstore.ErrNotFound):
		return graphFailure("not_found", err.Error(), 3)
	default:
		return graphFailure(fallback, err.Error(), 5)
	}
}
