package main

import "github.com/spf13/cobra"

// Cobra renders help before graph workspace admission. Keep the ordinary help
// intact and add an explicitly scoped section for commands reused by the graph
// preview, so their legacy descriptions do not mislead graph operators.
func init() {
	for _, entry := range []struct {
		cmd  *cobra.Command
		text string
	}{
		{initCmd, `Use bd init --graph-mode link --scope-url URL in a fresh workspace.
The Scope URL names local identities; it does not start a web server.
--server --external selects an ordinary shared Dolt server; otherwise storage
is embedded. Existing .beads directories are never adopted or overwritten.`},
		{rememberCmd, `Create a Memory with bd remember 'Policy text' [--id beads/policy]
[--title 'Policy']. An omitted ID is generated; an omitted title summarizes the
body. Explicit duplicate IDs fail. Graph Memories use canonical IDs, not keys.

Update an existing Memory with:
  bd remember 'Revised policy' --update beads/policy
  bd remember --update beads/policy --title 'New title'
Omitted fields remain unchanged. Use --body-file PATH or --stdin instead of
positional body text. --update is required for an existing Memory; --id is
creation-only. The current revision is accepted by default; add
--if-revision TOKEN to reject a stale update. Read the token with
bd show beads/policy --json. --unconditional explicitly selects the default.`},
		{memoriesCmd, `Search Memory titles and bodies with bd memories [SEARCH].
Use --all for a complete bounded result, --details for version/Link counts,
or --format records-json for machine-readable summaries. --json is unavailable;
use bd recall beads/ID for the exact body.`},
		{recallCmd, `Use bd recall beads/ID for one exact Memory body. --version TOKEN
selects a retained body; use bd show beads/ID --json for the record. Graph
recall does not accept legacy keys or --json.`},
		{createCmd, `Create an Issue by default, with an optional --id beads/PATH.
Use --bead-type types/preview-memory-v2 to create a Memory instead:
  bd create --bead-type types/preview-memory-v2 --body 'Code flow policy'
Both Type names and full local Type URLs work. --type remains the Issue
classification (for example task or bug), not the Bead Type. Memory creation
accepts body/description/message and an optional title; unsupported Issue-only
fields refuse.`},
		{showCmd, `Use bd show beads/ID or bd show links/ID for a current Memory,
Issue, or Link. --version TOKEN selects one exact retained record; this is not
an ordered history listing. --json returns the experimental graph record.`},
		{updateCmd, `Use bd remember --update beads/ID for selected Memory title/body
edits. For complete Memory or informational Link property replacement, use:
  bd update beads/policy --properties '{"title":"Policy","body":"Text"}' --if-revision TOKEN
--patch applies ordered property operations. Generic updates require
--if-revision TOKEN or --unconditional. Informational Links owned by a Memory
may also use --if-source-revision TOKEN; without it, the current source is
accepted. Blocking Dependency properties are not editable here. Issue scalar
edits and standalone --claim are separate graph operations.`},
		{deleteCmd, `For an unreferenced Memory, bd delete beads/ID previews the
deletion without writing. Apply with --force and either --if-revision TOKEN or
--unconditional. Referenced Memories refuse; graph deletion does not cascade.`},
		{forgetCmd, `Use bd forget beads/ID to delete one unreferenced Memory now.
Supply --if-revision TOKEN or --unconditional. Canonical IDs are retained and
incident Links prevent deletion; no cascade is performed.`},
		{depAddCmd, `Use bd dep add beads/issue beads/blocker for a blocking
Dependency between two live Issues. Memory endpoints, remote routing and bulk
dependency flags are unavailable in this preview.`},
		{linkCmd, `Without --link-type, bd link SOURCE TARGET creates the ordinary
blocking Dependency between two live Issues. For a Memory or Issue endpoint,
choose an installed informational Type, for example on a fresh workspace:
  bd link beads/policy beads/work --link-type types/example-cites
--link-type accepts types/NAME or a full local Type URL. An optional --id
selects links/PATH; --properties supplies informational Link properties.
Memory-owned Links accept the current source by default, or use
--if-source-revision TOKEN to reject a stale source. --unconditional-source
explicitly selects the default. Blocking Types require Issue endpoints.`},
		{closeCmd, `Close one live Issue by canonical beads/ID. Batch, force and
remote-routing forms are unavailable in this graph preview.`},
		{reopenCmd, `Reopen one closed Issue by canonical beads/ID, optionally
with --reason. Batch and remote-routing forms are unavailable.`},
		{readyCmd, `Show current ready Issues with no graph-specific filters.
This graph preview refuses positive BEADS_MAX_ROWS instead of truncating.`},
		{listCmd, `List current Issues with --flat or --format records-json.
Supported graph filters include status, type, title, priority, assignee,
labels, pinned and due dates. Tree output and legacy --json are unavailable.`},
		{blockedCmd, `Show current dependency-blocked Issues with canonical
blocker IDs. Filters and positive BEADS_MAX_ROWS are unavailable.`},
		{graphCmd, `Use bd graph beads/ID --view generic to traverse the current
local graph. Control direction, depth, node and Link bounds with --direction,
--depth, --max-nodes and --max-links. Legacy visualization modes are unavailable.`},
		{statusCmd, `Use bd status --graph to inspect this workspace's supported
graph capabilities and limits. Ordinary issue-statistics mode is unavailable.`},
	} {
		entry.cmd.Long += "\n\nGraph preview workspaces:\n" + entry.text
	}
}
