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
		{rememberCmd, `Create a Memory with bd remember 'Policy text' [--id policy]
[--title 'Policy']. An omitted ID is generated; an omitted title summarizes the
body. Explicit duplicate IDs fail. Bare policy means canonical beads/policy.
Graph Memories use canonical IDs, not legacy keys.

Update an existing Memory with:
  bd remember 'Revised policy' --update policy
  bd remember --update policy --title 'New title'
Omitted fields remain unchanged. Use --body-file PATH or --stdin instead of
positional body text. --update is required for an existing Memory; --id is
creation-only. The current revision is accepted by default; add
--if-revision TOKEN to reject a stale update. Read the token with
bd show policy --json. --unconditional explicitly selects the default.`},
		{memoriesCmd, `Search Memory titles and bodies with bd memories [SEARCH].
Use --all for a complete bounded result, --details for version/Link counts,
or --format records-json for machine-readable summaries. --json is unavailable;
use bd recall ID for the exact body.`},
		{recallCmd, `Use bd recall ID for one exact Memory body. --version TOKEN
selects a retained body; use bd show ID --json for the record. Graph
recall does not accept legacy keys or --json.`},
		{createCmd, `Create an Issue by default, with an optional --id ID.
Use --bead-type types/preview-memory-v2 to create a Memory instead:
  bd create --bead-type types/preview-memory-v2 --id policy --body 'Code flow policy'
Use bd types to see Bead Types installed in this workspace. Both types/NAME
and full local Type URLs work. --type remains the Issue classification
(for example task or bug), not the Bead Type. Memory creation
accepts body/description/message and an optional title; unsupported Issue-only
fields refuse.`},
		{showCmd, `Use bd show ID (equivalent to beads/ID) for a current Memory or
Issue; use bd show links/ID for a Link. --version TOKEN selects one exact
retained record; this is not
an ordered history listing. --json returns the experimental graph record.`},
		{updateCmd, `Use bd remember --update ID for selected Memory title/body
edits. For complete Memory or informational Link property replacement, use:
  bd update policy --properties '{"title":"Policy","body":"Text"}' --if-revision TOKEN
--patch applies ordered property operations. Generic updates require
--if-revision TOKEN or --unconditional. Informational Links owned by a Memory
may also use --if-source-revision TOKEN; without it, the current source is
accepted. Blocking Dependency properties are not editable here. Issue scalar
edits and standalone --claim are separate graph operations.`},
		{deleteCmd, `For an unreferenced Memory, bd delete ID previews the
deletion without writing. Apply with --force and either --if-revision TOKEN or
--unconditional. Referenced Memories refuse; graph deletion does not cascade.`},
		{forgetCmd, `Use bd forget ID to delete one unreferenced Memory now.
Supply --if-revision TOKEN or --unconditional. Canonical IDs are retained and
incident Links prevent deletion; no cascade is performed.`},
		{depAddCmd, `Use bd dep add issue blocker for a blocking
Dependency between two live Issues. Memory endpoints, remote routing and bulk
dependency flags are unavailable in this preview.`},
		{linkCmd, `Without --link-type, bd link SOURCE TARGET creates the ordinary
blocking Dependency between two live Issues. For a Memory or Issue endpoint,
choose an installed informational Type, for example on a fresh workspace:
  bd link policy work --link-type types/example-cites
Use bd types to see Link Types installed in this workspace. --link-type
accepts types/NAME or a full local Type URL. An optional --id
selects links/PATH; --properties supplies informational Link properties.
Memory-owned Links accept the current source by default, or use
--if-source-revision TOKEN to reject a stale source. --unconditional-source
explicitly selects the default. Blocking Types require Issue endpoints.`},
		{closeCmd, `Close one live Issue by ID or beads/ID. Batch, force and
remote-routing forms are unavailable in this graph preview.`},
		{reopenCmd, `Reopen one closed Issue by ID or beads/ID, optionally
with --reason. Batch and remote-routing forms are unavailable.`},
		{readyCmd, `Show current ready Issues with no graph-specific filters.
This graph preview refuses positive BEADS_MAX_ROWS instead of truncating.`},
		{listCmd, `List current Beads of all installed Bead Types. Human output
is flat by default; --format records-json returns complete records. Use
--bead-type types/NAME to narrow by Bead Type.
Issue-specific filters such as status, --type, priority, assignee, labels,
pinned and due dates select Issues through the existing Issue query; they
cannot be combined with a non-Issue --bead-type. The unfiltered mixed list
uses a bounded complete current snapshot, not a continuation cursor. Tree
output and legacy --json are unavailable.`},
		{blockedCmd, `Show current dependency-blocked Issues with canonical
blocker IDs. Filters and positive BEADS_MAX_ROWS are unavailable.`},
		{graphCmd, `Use bd graph ID --view generic to traverse the current
local graph. Control direction, depth, node and Link bounds with --direction,
--depth, --max-nodes and --max-links. Legacy visualization modes are unavailable.`},
		{statusCmd, `Use bd status --graph to inspect this workspace's supported
graph capabilities and limits. Ordinary issue-statistics mode is unavailable.`},
		{typesCmd, `Use bd types to list the Bead and Link Types actually installed
in this workspace, grouped by category. The displayed types/NAME IDs are
accepted by --bead-type and --link-type. Add --details to show each complete
persisted Type descriptor. --json returns the full descriptors as structured
data. Legacy Issue classifications such as task and bug belong to --type;
--sections is unavailable in graph preview workspaces.`},
	} {
		entry.cmd.Long += "\n\nGraph preview workspaces:\n" + entry.text
	}
}
