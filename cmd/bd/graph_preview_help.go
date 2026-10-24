package main

import "github.com/spf13/cobra"

// Cobra renders help before graph workspace admission. Keep the ordinary help
// intact and add an explicitly scoped section for commands reused by the graph
// preview, so their legacy descriptions do not mislead graph operators.
func init() {
	todoCmd.Long = `Graph preview workspaces:
bd todo add TITLE creates an open task Issue at P2 by default, assigned to
the current actor. bd todo and bd todo list show open task Issues;
bd todo list --all includes completed task Issues. bd todo done ID... closes
any Issue, as in ordinary workspaces, with reason Completed unless --reason
is supplied. Each unguarded ID is attempted independently; failed targets
are reported to stderr while successful closes are returned.
For one Issue, --if-revision TOKEN refuses a stale completion. List and add
JSON keep the familiar Issue shape with canonical beads/PATH IDs.

Ordinary Issue workspaces:
` + todoCmd.Long
	// These commands have long ordinary-mode descriptions. Put the graph
	// distinction first so help does not lead graph users to turn on legacy
	// recording or expect Dolt commits and unsupported flags.
	versionsCmd.Short = "List retained graph Resource versions or recorded ordinary Bead versions"
	versionsCmd.Long = `Graph preview workspaces:
Use bd versions ID to list one Memory, Issue or Link's retained versions newest
first. Bare ID means beads/ID; use links/PATH for a Link. Every new graph
Resource has a creation version; do not enable ordinary versioned-history
recording for this command. Cite the opaque version token with bd show ID
--version TOKEN or bd compare ID --from TOKEN --to TOKEN. The store-local
local_revision only orders versions within this store. --json names that field
local_revision, matching ordinary bd versions. BDP HTTP History is unavailable.
A removed Link lists only its prior citable versions; deletion adds no version.

Ordinary Issue workspaces:
` + versionsCmd.Long
	historyCmd.Short = "Show graph Resource versions or ordinary Issue commit history"
	historyCmd.Long = `Graph preview workspaces:
bd history ID is an alias for bd versions ID. It lists retained Memory, Issue
or Link versions newest first; bare ID means beads/ID and Link IDs use
links/PATH. Use the opaque token for exact reads, not local_revision.
--limit and --events are not supported by the graph alias. BDP HTTP History
is not available.

Ordinary Issue workspaces:
` + historyCmd.Long
	deferCmd.Long = `Graph preview workspaces:
Use bd defer ID... to defer one or more Issues. --until accepts ordinary bd
date and relative-time forms; an undated defer stays deferred until restored.
--reason appends to Issue notes. A later bd ready wakes due Issues, recording
one native and graph version per changed Issue. A repeated dateless defer is a
no-op when neither date nor notes change. Use --if-revision TOKEN for a
single-Issue stale-write check. Without it, the current revision is accepted.

Ordinary Issue workspaces:
` + deferCmd.Long
	undeferCmd.Long = `Graph preview workspaces:
Use bd undefer ID... to restore one or more deferred Issues to open. It also
clears a stale defer date on a non-deferred Issue. A repeated undefer is a
no-op if nothing changes. Use --if-revision TOKEN for a single-Issue
stale-write check. Without it, the current revision is accepted.

Ordinary Issue workspaces:
` + undeferCmd.Long
	quickCmd.Long = `Graph preview workspaces:
Use bd q TITLE to create one Issue and print only its Scope-relative ID.
Pass -- before a title beginning with '-', for example bd q -- -follow-up.
--parent is unavailable here; the graph parent-child journey
is available through bd create on fresh graph workspaces.

Ordinary Issue workspaces:
` + quickCmd.Long
	for _, entry := range []struct {
		cmd  *cobra.Command
		text string
	}{
		{initCmd, `Use bd init --graph-mode link --scope-url URL in a fresh workspace.
The Scope URL names local identities; it does not start a web server.
--server --external selects an ordinary shared Dolt server; otherwise storage
is embedded. Existing .beads directories are never adopted or overwritten.`},
		{rememberCmd, `Store a Memory with bd remember 'Policy text' [--id policy]
[--title 'Policy'] [--metadata '{"team":"docs"}']. Alternatively,
--properties '{"title":"Policy"}' initializes the typed Memory property;
do not supply the same field with both --properties and a shorthand flag.
Metadata is a JSON object separate from Memory title/body properties. An omitted ID is generated; an omitted creation title
summarizes the body. With --id, an unused ID creates and an existing Memory
updates in place, as ordinary bd remember --key does. Bare policy means
canonical beads/policy. Graph Memories use canonical IDs, not legacy keys.

To refuse any previously allocated ID, add --create-only with --id.
To require an existing Memory, keep using --update instead of --id:
  bd remember 'Revised policy' --id policy
  bd remember 'Revised policy' --update policy
  bd remember --update policy --title 'New title'
--update retains literal Memory-ID semantics, including raw @; it does not
parse revision suffixes. The selected bd update targets refuse revision suffixes.
Omitted fields remain unchanged on update. Use --body-file PATH or --stdin
instead of positional body text. An existing-ID update accepts the current
revision by default; add --if-revision TOKEN to reject a stale update. Read
the token with bd show policy --json. --replace-properties JSON replaces all
Memory properties where the Type permits it, and --replace-metadata JSON
replaces all metadata; {} clears the selected document.`},
		{memoriesCmd, `Search Memory titles and bodies with bd memories [SEARCH].
Use --all for a complete bounded result, --details for version/Link counts,
or --format records-json for machine-readable summaries. --json is unavailable;
use bd recall ID for the exact body.`},
		{recallCmd, `Use bd recall ID for one exact Memory body. --version TOKEN
selects a retained body; use bd show ID --json for the record. Graph
recall does not accept legacy keys or --json.`},
		{createCmd, `Create an Issue by default, with an optional --id ID.
Without --id, the canonical beads/ path uses the Issue ID allocated by the
ordinary writer, including this workspace's configured prefix and ID mode.
An explicit --id ID keeps the entire spelling as the native Issue ID, even
when it contains slashes, and maps it to beads/ID. It must match the
workspace's configured Issue prefix unless --force overrides that shipping
check. Use --graph-path beads/PATH to select only a graph path and generate
the native Issue ID separately. --id and --graph-path cannot be combined.
On a fresh graph workspace, --parent ID creates a dotted child ID
and its parent-child Link in one transaction. The parent must be a local Issue;
the child inherits its labels unless --no-inherit-labels is set. --parent and
--id or --graph-path cannot be combined with --parent.
--metadata accepts one JSON object on Issue or Memory creation; omitted
metadata is {}. Metadata is separate from Type-validated properties.
--properties JSON initializes the selected Type's writable fields, such as
Issue description or Memory body. It may accompany shorthand flags only when
they supply different fields; duplicate fields refuse.
Use --type types/memory to create a Memory instead:
  bd create --type types/memory --id policy --body 'Code flow policy'
Use bd types to see Bead Types installed in this workspace. Use types/NAME
or a full local Type URL for nominal Types. Bare names select Issue
classifications (for example task or bug); the CLI selects types/issue for them. Memory creation
accepts body/description/message and an optional title; unsupported Issue-only
fields refuse. Fresh workspaces create types/issue Issues.`},
		{showCmd, `Use bd show ID or bd show --id ID (equivalent to beads/ID) for a current Memory or
Issue; use bd show links/ID for a Link. Multiple positional/--id operands
show current Issues in the supplied order, with ordinary Issue JSON or human
details. Memory, Link, exact-version and graph-json reads remain single-ID.
--refs shows incoming Issue references, grouped by relationship; --children
shows only parent-child references. Both accept current Issue IDs and retain
the familiar human view or Issue-ID-keyed JSON object.
--watch follows one current Issue, redrawing its human detail view when that
Issue changes; stop it with Ctrl+C.
--current selects the actor's first in-progress Issue, then hooked Issue, then
the last-touched Issue; it cannot combine with an explicit ID.
ID@N selects the stable store-local revision number N. ID@=TOKEN or
'[URL]@=TOKEN' selects one exact retained record; --version TOKEN remains
an exact-token alternative (not both). @current reads current. Unwrapped URL forms containing :// are
literal, including raw @; use the full canonical Scope URL to escape local
@ shorthand. Colon-only forms such as plan:v1@=TOKEN use local syntax.
Bracketed operands require canonical HTTP(S) Scope URLs, not opaque URIs.
Suffixes are rejected on the selected update and deletion targets; use
--if-revision for write comparison. --version treats current and unpinned as exact tokens.
Endpoint pins remain unavailable. This is not an ordered history listing.
A current Issue's
--json uses the ordinary detail array with Scope-relative Bead IDs and its
graph revision token. Memory, Link, and exact --version reads use the
experimental graph envelope; --format graph-json explicitly requests that
envelope for a current Issue too. graph-json is not a BDP Resource format.
Showing an Issue makes it the last-touched Issue for interactive update/close;
showing a Memory or Link does not.`},
		{updateCmd, `Use bd remember --update ID for selected Memory title/body
edits. --properties JSON shallowly merges named top-level properties on an
Issue, Memory, or informational Link, preserving omitted keys. The same flag
initializes properties on bd create and bd remember. For example:
  bd update policy --properties '{"body":"Revised text"}' --if-revision TOKEN
An empty merge object is a no-op. --replace-properties JSON replaces all
writable properties, and --replace-metadata JSON replaces all metadata; {}
clears the selected document where the Type allows it. The replacement flags
cannot combine with merge, patch, or key edits for the same document, but
property and metadata changes can be made atomically. --patch applies ordered add/replace/remove property
operations, including on writable native Issue scalar properties.
--metadata merges a JSON object's top-level keys; --set-metadata KEY=VALUE
sets one typed JSON value and --unset-metadata KEY removes one key. Set and
unset may combine (unset wins); --metadata cannot combine with either.
Metadata may accompany a property or Issue scalar edit atomically, and a
metadata-only update uses the same Resource and owning-source guards.
Generic updates accept the current state when --if-revision is omitted.
Informational Links owned by a Memory may also use --if-source-revision TOKEN;
without it, the current source is
accepted. --source-revision TOKEN and --target-revision TOKEN re-pin an
informational Link without changing its ID or endpoint identities. Use current
to resolve an exact pin inside the transaction or unpinned to float it; omission
preserves the existing pin. Pins can combine atomically with properties and
metadata. Owned-source pins remain unavailable. Native Dependency properties and metadata are not editable here. Issue scalar
edits and standalone --claim are separate graph operations. Without an ID,
interactive Issue update uses the last-touched Issue; scripts require an ID
unless BD_LAST_TOUCHED_FALLBACK=1 explicitly enables the fallback.`},
		{deleteCmd, `bd delete ID previews a Memory or Issue without writing.
Multiple Issue IDs and --from-file select a checked Issue set; --cascade
includes dependent Issues. Apply with --force, or keep any delete request
read-only with --dry-run. --if-revision TOKEN optionally rejects stale
single-Issue state. Referenced Memories refuse and remain single-ID;
dotted native Issue IDs refuse set deletion pending citation qualification.`},
		{forgetCmd, `Use bd forget ID to delete one unreferenced Memory now.
Add --if-revision TOKEN to reject stale state. Canonical IDs are retained and
incident Links prevent deletion; no cascade is performed.`},
		{depCmd, `Use bd dep BLOCKER --blocks BLOCKED to create a blocking
Dependency between two live Issues. Bare bd dep prints help; the ordinary
--no-cycle-check option is unavailable in graph preview workspaces.`},
		{depAddCmd, `Use bd dep add issue blocker for a native
Dependency between two live Issues, optionally with --id links/ID.
--type KIND sets its dep_type property; the default is blocks. Any nonempty
kind up to 32 bytes is accepted. Built-in blocks and parent-child retain
their native scheduling or hierarchy behavior; unknown kinds do not block.
The ordinary --blocked-by blocker and --depends-on blocker spellings replace
the second positional argument. For a single edge, --no-cycle-check only
suppresses a post-add warning; the domain writer still validates cycles.
bd dep blocker --blocks issue is the same operation with reversed arguments.
Memory endpoints, remote routing and bulk dependency flags are unavailable
in this preview. Native dep operands retain literal Issue-ID semantics,
including @; bd link parses the selected endpoint revision syntax, but exact
pins and @current are supported for local informational Links. Native
Dependency endpoints remain floating.`},
		{depRemoveCmd, `Use bd dep remove issue blocker (or bd dep rm) to
remove one local blocking Dependency by its Issue pair. An absent pair is a
successful no-op and reports status not_found with --json. A real removal
records one native dependency_removed event and one source Issue version;
the target Issue is unchanged. Bare native Issue IDs are accepted when their
graph path is unambiguous; use explicit beads/PATH if a different Bead claims
that bare name. These native pair operands retain literal Issue-ID semantics,
including @. For Link-ID removal use bd link delete ID --force.`},
		{linkCmd, `The canonical graph Link namespace is bd link
create/list/show/update/delete/versions/compare/count. Create requires
--type and SOURCE TARGET (or SOURCE --to TARGET, or --from SOURCE --to TARGET); list and count query complete current
Link collections within the preview's advertised bounds. Delete previews
by default and --force applies the checked mutation.

Without --type, bd link SOURCE TARGET creates the ordinary
blocking Dependency between two live Issues. In graph workspaces this form
also parses the selected revision syntax for informational Links; exact local
pins and @current are supported, while native Dependency endpoints remain
floating. bd dep add/remove keep literal Issue IDs, so the input
syntaxes are not interchangeable. To create any native Dependency
kind through the canonical namespace, use
  bd link create --type blocks --from ISSUE --to ISSUE
Only built-in Dependency kinds are accepted. The resulting Link has a dep_type property.
For a Memory or Issue endpoint,
choose an installed informational Type, for example on a fresh workspace:
  bd link policy work --type types/references
Use bd types to see Link Types installed in this workspace. --type
accepts a built-in Dependency kind, types/NAME or a full local Type URL. An optional --id
selects a bare Link ID or links/PATH; --properties supplies informational Link properties.
Memory-owned Links accept the current source by default, or use
--if-source-revision TOKEN to reject a stale source. Endpoint @unpinned suffixes
or --source-revision unpinned/--target-revision unpinned keep endpoints floating.
Exact endpoint revisions persist on local informational Links; current resolves
an exact pin inside the writer transaction. Owned-source pins remain unavailable.
Native Dependency endpoints remain floating. Endpoint selectors are separate
from --if-source-revision. --metadata JSON supplies
an initial open metadata object for informational Links; native Dependencies
do not accept it. The single Type types/dependency
requires Issue endpoints. Add --if-source-revision TOKEN to reject a stale source.`},
		{closeCmd, `Close one or more local Issues by ID or beads/ID. One --reason
applies to all IDs; repeat it once per ID for positional reasons, or use
--reason-file PATH for literal file content. The done alias accepts a trailing
positional reason. A batch reports successful Issues on stdout and per-Issue
failures on stderr, then exits nonzero if any failed. --force bypasses pinned,
holder, blocker and open-child policy; --session (or CLAUDE_SESSION_ID) records
the closing session. On one Issue, --if-revision TOKEN from bd show --json
rejects a stale close, including an already-closed retry. It cannot be used
with multiple IDs, --suggest-next or --claim-next. Without it, close accepts
the current revision. On one Issue, --suggest-next lists Issues that closing it
newly unblocks, without claiming them. --claim-next atomically claims the
highest-priority ready Issue when at least one close lands; an already-closed
retry earns no new claim. A missing ID aborts a batch before any close or
claim; other per-Issue policy refusals may keep successful closes and a claim.
Without an ID, interactive close uses the
last-touched Issue; scripts require an ID unless BD_LAST_TOUCHED_FALLBACK=1
explicitly enables the fallback. Gate evaluation, --continue, molecule
advancement and remote-routing forms remain unavailable.`},
		{reopenCmd, `Reopen one or more local closed Issues by ID or beads/ID,
optionally with --reason. A batch reports changed Issues on stdout and
per-Issue failures on stderr, then exits nonzero if any failed. Already-open
Issues remain unchanged. Remote-routing forms remain unavailable.`},
		{unclaimCmd, `Use bd unclaim ID... to release one or more assigned open or
in-progress Issues. By default only the current holder may release a claim.
--force bypasses holder authorization but still respects the native row CAS;
--if-assignee HOLDER releases only while that holder remains assigned. The two
flags cannot be combined. --reason TEXT adds a native Issue comment after a
successful release; bd comments ID reads it. The release clears the assignee,
lease and started time, returns the Issue to open, and records one Issue and
graph version. A repeated release refuses without creating a version.`},
		{commentsCmd, `Use bd comments ID to read the native comment feed for a
graph Issue. bd comment ID TEXT and bd comments add ID TEXT each append one
comment; --json uses the ordinary comment or comment-array shape. bd unclaim
--reason also appends there. Comments are outside the retained Issue snapshot
and do not mint an Issue version.`},
		{readyCmd, `Show current ready Issues with admitted priority, type, label,
assignee, sort and deferred-state filters. --claim atomically claims the first
matching ready Issue. This graph preview refuses positive BEADS_MAX_ROWS
instead of truncating; parent, molecule, ephemeral and metadata filters remain
unfinished.`},
		{listCmd, `Without Issue filters, bd list reads one bounded snapshot of current Beads
of all installed Bead Types and lists every Memory and every Issue the
ordinary bd list would show, newest recorded change first. Closed and pinned
Issues are hidden unless --all is given; --all also removes the row limit.
Each human row shows the local ID and kind; an Issue row also shows its
status and priority. Use --type types/NAME to narrow by Bead Type.
--limit returns a prefix with hasMore, not a continuation cursor. A positive
BEADS_MAX_ROWS refuses a page of more Beads than that.
Issue-specific filters (--status, --type, --title, --title-contains,
--priority, --priority-min, --priority-max, --assignee, --no-assignee,
--label, --label-any, --exclude-label, --pinned, --no-pinned, --due-before,
--due-after, --overdue, --sort, --reverse) or a matching configured
directory label switch to the existing Issue query: Issues only, closed and
pinned Issues omitted unless --all or a filter selects them, quoted rows
with status and priority, and a line under the header saying Memories are
not listed. An Issue filter cannot be combined with a non-Issue --type;
a configured directory label alone does not refuse one, it is simply not
applied to Memories. Tree output is unavailable; --tree=false selects the
same flat output as --flat. --json and --format json
return the ordinary Issue-only array; use --format records-json for generic
Bead records.`},
		{searchCmd, `Search current graph Issues by title or canonical local Bead ID.
Results include closed Issues unless --status narrows them. --json returns the
ordinary counted Issue row array with Scope-relative IDs. Supported filters:
--status, --assignee, --type, --label, --label-any, --priority-min,
--priority-max, --desc-contains, --notes-contains, --external-contains,
--empty-description, --no-assignee, --no-labels, created/updated/closed date
bounds, --metadata-field, --has-metadata-key, --limit, --sort, --reverse, and
--long. Only built-in status names are admitted here.`},
		{countCmd, `Count current durable graph Issues, including closed Issues by default.
The ordinary scalar and --by-status/priority/type/assignee/label JSON shapes
are preserved. Supported filters: --status, --priority, --assignee, --type,
--label, --label-any, --title, --title-contains, --desc-contains,
--notes-contains, date bounds, --empty-description, --no-assignee, --no-labels,
--priority-min/max, --id, --parent, --no-parent, --exclude-type and
--exclude-status. IDs and parents accept local graph Issue paths or bare IDs;
an edge-less dotted child matches --parent by ID prefix and --no-parent as a
top-level Issue. Excluded statuses are limited to built-in names. Ephemeral,
infrastructure and metadata count filters are unavailable in graph preview.`},
		{queryCmd, `Query current durable graph Issues with ordinary field comparisons,
AND, OR, NOT, dates and canonical local Bead IDs. Closed Issues are excluded
unless --all or an explicit status comparison includes them. --json returns
the ordinary counted Issue row array. --limit, --offset, --sort, --reverse,
--long and --parse-only work. A parent=ID comparison works in an AND chain;
OR and NOT parent forms follow native query refusals. mol_type and ephemeral
predicates are unavailable in graph preview.`},
		{blockedCmd, `Show current dependency-blocked Issues with canonical
blocker IDs. Filters and positive BEADS_MAX_ROWS are unavailable.`},
		{graphCmd, `Use bd graph BEAD or bd graph --all in a link workspace.
The default Link selector is Dependency, including its refinements, exactly
as --link-type Dependency. Repeat --link-type NAME for an OR union, or use
--all-link-types. Excluded Link Types cannot bridge the traversal.
One --direction in|out|both (default both) and --depth (default 1000) apply
to a named-root selection. --all refuses explicit --depth/--direction. Depth counts Link hops to select Beads; every matching
Link between selected Beads is displayed, including parallels and cycles.
A frontier explicitly reports depth truncation. Node/Link bounds refuse on
exhaustion. Acquisition remains a complete bounded current inventory (at most
1000 Resources / 16 MiB; serialized output at most 16 MiB), not scalable or arbitrary historical traversal.
Authored endpoint revision pins remain visible; adjacent Beads are current
records and are not silently substituted for those pinned revisions.
--all with the Dependency selector retains open Issues, including isolated
ones. Other selectors show matching components/endpoints, excluding unrelated
isolates. A named root stays visible without matching Links. --open remains unavailable in link mode; dependency mode retains its existing view. --json carries complete canonical Beads/Links
and separate traversal metadata; --dot, --html, --compact and --box also work.
--view generic is optional/deprecated. bd graph check retains its existing
blocking-dependency consistency report.`},
		{statusCmd, `Use bd status for ordinary Issue statistics in this workspace.
Use bd status --graph to inspect graph capabilities and limits.`},
		{typesCmd, `Use bd types to list the Bead and Link Types actually installed
in this workspace, grouped by category. The displayed types/NAME IDs are
accepted by --type on the matching command. Add --details to show each complete
persisted Type descriptor. --json returns the full descriptors as structured
data. Built-in Issue and Dependency values such as task and blocks also work
with --type, but are not separate installed graph Types in this preview;
--sections is unavailable in graph preview workspaces.`},
	} {
		entry.cmd.Long += "\n\nGraph preview workspaces:\n" + entry.text
	}
}
