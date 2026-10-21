# Graph CLI guide

> **New workspaces only.** The graph preview cannot be enabled on an existing
> `.beads` project or used to migrate its Issue database. Start in a new
> directory without `.beads`; `bd init --graph-mode link` refuses an existing
> or incomplete workspace rather than replacing its data. Keep ordinary
> projects on their existing format. Do not copy a `.beads` directory into a
> graph workspace as an upgrade path.

Graph workspaces are not migrated between graph schema versions. If a
workspace's `graph_schema_version` in `.beads/metadata.json` is not this
build's `7`, `bd` refuses it with `graph_not_initialized` and does not open
its database. Preserve its data and create a fresh workspace. The selected Memory Type catalog must match the build; an older catalog is
refused even when the storage schema version matches. See
[Discover installed Types](#discover-installed-types).

Build `bd` from the [integration branch](https://github.com/versioned-beads/beads/tree/integration),
not a released binary, and initialize a new project explicitly:

```sh
mkdir graph-demo && cd graph-demo
git init
bd init --graph-mode link --scope-url https://example.org/team/ \
  --skip-hooks --skip-agents --non-interactive
bd status --graph
```

The Scope URL establishes canonical identity; this command does not start an
HTTP server at that address. The example skips agent-file and hook setup to
keep the CLI exercise isolated. See the
[Graph CLI Specification (Draft)](/reference/graph-cli-specification-draft) for supported
embedded and external Dolt modes, agent setup, limits, and refusals.

This page is the evolving **task-oriented CLI guide**. The graph preview
technical reference is the evolving, detailed command matrix and contract;
neither page is a frozen release note. The blog post explains the model and
links to these pages for commands that may change after publication. Ordinary
Beads workspaces keep their existing Issue and key/value-memory commands.
The technical reference also carries the [proposed next-milestone command contract](/reference/graph-cli-specification-draft#proposed-command-contract-for-the-next-milestone);
its **NYI** rows describe proposed behavior, not commands available in this
build. Use the current command matrix there when trying the preview.

The coordination and blocking-view walkthroughs below describe a **local,
unqualified combined-tail candidate**. They are for a disposable workspace
created from that candidate, not a claim that these commands have landed in the
fork integration branch or a released binary. The draft specification keeps
the older pinned integration ledger separate from this candidate inventory.

The selected Memory Type examples and catalog below describe the Preview 3
candidate in [PR130](https://github.com/donnabox/beads/pull/130). Check the
[release ledger](https://github.com/gastownhall/beads/issues/7656) for its actual
integration and qualification status before choosing a build.

In a graph workspace, a Bead is an Issue or a Memory. Its canonical identity
is under `beads/`. In CLI arguments, `policy` means `beads/policy`; a Link
still needs an explicit `links/ID` where a command accepts either kind of
resource. The shorthand does not change stored IDs or HTTP URLs. Omit `--id`
when creating a Bead to allocate an ID, or supply one to use it exactly.
`bd create` refuses a duplicate ID; `bd remember --id ID` updates an existing
Memory at that ID unless `--create-only` is supplied. The examples below use
explicit IDs only so later commands are easy to follow.

## Create Beads

`bd remember` stores a Memory from text. It preserves the full body. Without
`--title`, it makes a title from the first nonempty body line (collapsed
whitespace, at most 80 Unicode characters including an ellipsis) on creation.
An explicit `--title` is used as supplied. It can also read a body from
`--body-file PATH` or `--stdin` instead of the positional text. With `--id`,
an unused ID creates and an existing Memory updates in place, matching
ordinary `bd remember --key` behavior. `--create-only` with `--id` refuses
any previously allocated ID, including one that has been deleted.

```sh
bd remember 'Code flow policy: changes land on integration.' --id policy \
  --metadata '{"team":"docs"}'
bd remember 'Keep review branches until their changes land.' --title 'Review branch policy'
```

`bd create` creates an Issue by default. `--type task` or `--type bug`
selects an existing Issue classification. To select a nominal Bead Type,
pass its printed `types/NAME` selector or full local Type URL to `--type`.
This distinguishes a Memory Decision from an Issue classification named
`decision`. `--properties JSON` supplies the selected Type's properties.
The [Memory Type CLI facade](/reference/bdp/cli/types/memory/memory) gives
property, retyping and refusal details for the Preview 3 candidate.

When an Issue is created without `--id`, its canonical `beads/` ID uses the
Issue ID allocated under the workspace prefix set by `bd init --prefix`. For
example, a generated `mp-53` Issue is `beads/mp-53` locally. An explicitly
supplied `--id` remains the requested canonical path. Generated Memory IDs
remain random `beads/` paths; generated Link IDs remain under `links/`.

```sh
bd create 'Move the release branch' --id work --type task
bd create --type types/memory \
  --properties '{"title":"Code flow policy","body":"Keep the old policy as a versioned Memory."}'
```

Use `bd types` in the selected workspace to discover its installed
`types/NAME` IDs. Use the printed `types/NAME` selector or full local Type URL with `--type`;
bare names retain their existing Issue or Dependency classification meaning. Do not assume a
Type shown in another workspace is installed in this one.

## Find and read Beads

| Command | Graph workspace behavior |
| --- | --- |
| `bd memories [SEARCH]` | List current Memory title/body summaries. Use `--all` for all matches within the preview's bounds, `--details` for saved version and attribution, or `--format records-json` for structured summaries. |
| `bd recall ID` | Print **one** Memory's body followed by a newline, as ordinary `bd recall` does. It does not enumerate Memories. |
| `bd show ID --json` | A current Issue uses the ordinary `bd show` JSON detail array, including labels, dependency/comment counts, Scope-relative Bead IDs that round-trip to graph commands, and the graph revision token. Memory and Link still use the preview complete-record envelope. |
| `bd show ID --format graph-json` | Read the preview complete-record envelope explicitly, including for an Issue or an exact `--version` read. This is a preview record, not a BDP Resource representation. |
| `bd list` or `bd list --format records-json` | Without an Issue filter, list every current Memory and Issue the ordinary `bd list` would show, newest recorded change first; closed and pinned Issues need `--all`. Use `--type types/NAME` to narrow by exact declared nominal Type. An Issue filter, or a matching directory label, switches to the Issue-only query described below and says so. |
| `bd statuses [--json]` | List built-in Issue statuses and any custom statuses resolved for this workspace, using the ordinary command's human and JSON shapes. This is Issue status discovery, not Bead Type discovery. |

```sh
bd memories --all
bd memories 'code flow' --details
bd recall policy
bd show work --json
bd show work --format graph-json
bd list --all
bd list --json --all
bd list --format records-json --type types/memory --all
bd statuses --json
```

Without an Issue filter, `bd list` reads one bounded snapshot of every
current Bead. Each human row shows the Bead's local ID (such as
`beads/work`) and its kind (`Memory` or `Issue`); an Issue row also shows its
status and priority, then the title. Closed and pinned Issues are hidden
unless you pass `--all`, which also removes the row limit. `--limit` sets a
visible prefix and `hasMore` indicates that the prefix omitted matches.
It is not a continuation cursor. A positive `BEADS_MAX_ROWS` refuses a page of
more Beads than the cap. The BDP HTTP `beads/` collection provides
pagination in an ordinary shared-server graph workspace. Follow every
response's `next` URL until it is `null`, or use the
[public Python read example](https://github.com/versioned-beads/beads/blob/integration/examples/bdp-read/read_beads.py), which
follows those pages.
BDP Read returns the common `metadata` object on current Issues, Memories,
informational Links and owned Links, using `{}` when no keys have been set.
An older retained record can omit `metadata` without rewriting its history.

The Issue filters are `--status` (or `--state`), `--type`, `--title`,
`--title-contains`, `--priority`, `--priority-min`, `--priority-max`,
`--assignee`, `--no-assignee`, `--label`, `--label-any`, `--exclude-label`,
`--pinned`, `--no-pinned`, `--due-before`, `--due-after`, `--overdue`,
`--sort` and `--reverse`. Supplying an Issue filter switches `bd list` to the existing Issue query,
even when an empty value such as `--assignee=` adds no restriction.
`--type types/NAME` or a full Type URL instead selects a nominal Type;
a bare `--type task` remains an Issue classification filter. A configured `directory.labels` entry that matches the current
directory switches it too. That query lists Issues only, never Memories. It
omits closed and pinned Issues unless `--all` or a filter selects them. A line
under the header says Memories are not listed and names the option that
selected the query. A directory label alone never refuses a nominal `--type`; it
is not applied to a Memory. Its human rows add status and priority, for example
`"https://example.org/team/beads/work" "open" P2 "Move the release branch"`.
An Issue filter cannot be combined with a non-Issue nominal `--type`.

`bd list --json` returns the ordinary Issue-only array, including the
familiar Issue field names and Scope-relative IDs. Memories are omitted from
that shape. Use `--format records-json` to read all Bead Types as graph
records; the formats serve different clients and are not interchangeable.

## Quick task Issues

`bd todo` is a shortcut for regular task Issues in a graph workspace. Add
creates an open task at P2 by default and assigns it to the current actor.
The bare command and `bd todo list` show open tasks; `bd todo list --all`
includes completed tasks. JSON uses the ordinary todo Issue, array and
completion shapes with Scope-relative `beads/PATH` IDs.

```sh
bd todo add 'Write the migration tests' --description 'Cover old data'
bd todo list --all
bd todo done beads/TEAM-123 --reason 'Tests landed'
```

`bd todo done` accepts any Issue, including one whose type is not `task`.
Without a revision guard, it tries each supplied ID in order, reports a
failed target on stderr and still returns the successfully closed IDs with
exit status zero. For one Issue, `--if-revision TOKEN` from `bd show --json`
refuses a stale close before changing that Issue. An already closed Issue is a
no-op and retains its original close reason and version.

## Issue coordination

`bd heartbeat ISSUE` refreshes the current holder's in-progress lease without
creating an Issue version or History event. `bd reclaim` recovers expired
Issue leases through the native Issue writer; its `--older-than`, `--id`,
assignee, label and replica selectors retain their ordinary meanings. Both
commands accept local Issue IDs in a graph workspace. `bd state ISSUE
DIMENSION` reads a state label, and `bd state list ISSUE` reads all state
labels. `bd set-state` is still unavailable. Bare `bd status` reports the
ordinary Issue summary; `bd status --graph` reports graph capabilities.

`bd merge-slot create`, `check`, `acquire` and `release` use an Issue-only
coordination slot. Acquire accepts `--holder` and `--wait`; release accepts
`--holder`. The graph route preserves the ordinary flat JSON reports and
holder/queue behavior. The broader `bd gate` workflow remains unavailable.

In a fresh disposable graph workspace initialized as above, the following
sequence creates an Issue, claims it as one actor and refreshes that actor's
lease. `bd state` reads labels only; it does not create state. The slot is a
separate Issue-only coordination primitive, so create it once per workspace.

```sh
bd create 'Review the release' --id lease-demo --type task
bd update lease-demo --claim --actor demo
bd heartbeat lease-demo --actor demo
bd state lease-demo patrol
bd state list lease-demo --json
bd status
bd status --graph --json
bd merge-slot create --actor demo
bd merge-slot check --json
bd merge-slot acquire --holder demo --actor demo
bd merge-slot check --json
bd merge-slot release --holder demo --actor demo
```

An unset state dimension prints a missing-value message (or JSON `null`).
Heartbeat requires the current in-progress holder and writes no Issue version
or History event. To exercise recovery, let a claimed Issue's lease actually
expire first, then run `bd reclaim --older-than 0s --id lease-demo --json` in
that same disposable workspace. An active lease is not reclaimed; a successful
reclaim uses the native Issue writer and records the graph projection. The
default grace period is longer than zero. Scope flags include repeatable
`--id`, `--assignee`, `--label`, `--label-any`, `--exclude-label`, and
`--any-replica`; use the latter only when recovering a lease from a departed
replica. `bd set-state` and `bd gate` still refuse in graph mode.

## Blocking dependency views

`bd dep list ISSUE`, `bd dep tree ISSUE`, `bd dep cycles`, `bd graph ISSUE`
and `bd graph check` read the current blocking Dependency graph. The tree
supports `--direction down|up|both`; its both-direction human view labels
dependencies and dependents separately. `bd graph check --json` returns the
ordinary flat check object and local Issue IDs in its `cycles`. The existing
`bd graph ISSUE --view generic` still traverses generic Links.

These views read a complete bounded graph snapshot. `bd graph ISSUE` and
`bd dep tree ISSUE` refuse a positive `--max-rows` cap that would make their
blocking answer incomplete.
`dep list/tree/cycles --json` and `graph ISSUE --json` refuse until their
ordinary JSON shapes can be preserved; use the human view or the explicit
generic graph JSON view. In the local relationship-journeys candidate, fresh
v7 workspaces support `bd children PARENT` and `bd children PARENT --json` for
immediate native parent-child Links, including closed children. The JSON result
retains the ordinary Issue array shape; an empty family is `[]`. A dotted Issue
ID or blocking dependency alone does not make a child. `bd epic status` reports
progress from those same checked relationships, and `bd epic close-eligible`
closes an epic only when every child has a completing close reason. An explicit
`--actor` controls the close writer and History attribution; any failed close
makes the command exit nonzero, with successful sibling IDs in the diagnostic.
In a freshly initialized v7 workspace, this separate local
`create --parent PARENT` candidate uses the native Issue
writer to allocate a child ID and parent-child Dependency, then retains the
matching graph Issue and Link in the same transaction. It inherits the parent's
labels unless `--no-inherit-labels` is set. Pinned v5/v6 workspaces refuse these
new relationship journeys before opening the store. `children --pretty`,
duplicate and supersede still refuse. These candidates have not completed
release qualification or landed in the integration branch.

This disposable fixture shows the direction of a blocking edge: `plan`
depends on `tests`. The human views use local Issue IDs. `graph check --json`
returns a flat object containing `clean`, `cycles` and
`summary.cycle_count`; cycle members are local Issue IDs. It exits nonzero
when it finds a cycle.

```sh
bd create 'Prepare the release' --id plan --type task
bd create 'Write release tests' --id tests --type task
bd dep add plan tests
bd dep list plan
bd dep tree plan --direction down
bd dep tree tests --direction up
bd dep tree plan --direction both
bd dep cycles
bd graph plan
bd graph check --json
bd graph plan --view generic --direction out --depth 2 --json
```

`dep list --direction up` lists dependents instead; `dep tree` also accepts
`--max-depth` and deprecated `--reverse` (equivalent to `--direction up`).
`dep cycles --limit N` caps displayed cycles, not the count reported.
`dep list/tree/cycles --json` and `graph ISSUE --json` refuse because the
ordinary JSON shapes are not yet preserved. A positive `--max-rows` or
`BEADS_MAX_ROWS` also refuses the full blocking graph/tree snapshot rather
than returning a partial answer. Generic traversal is a separate explicit
Link view with its own bounds and JSON shape.

## Update and delete Beads

For a Memory, `bd remember --id ID` changes only the fields you supply if
that ID already exists. Omitted title or body stays unchanged, and the current
revision is accepted by default. `--update ID` remains an existing-only
spelling; it refuses a missing Memory. A title-only update needs no body
argument. Add `--if-revision TOKEN` to either spelling to reject a stale edit.

```sh
bd remember 'Changes now land on the release branch.' --id policy
bd remember --update policy --title 'Current code flow policy'
```

`--properties JSON` initializes writable Type properties on `bd create` and
`bd remember`, and shallowly merges named top-level properties on `bd update`
or an existing-ID `bd remember`. It works for the installed Issue and Memory
Bead Types and informational Link Types. Omitted keys survive; `{}` is a no-op.
Use `bd update ID --replace-properties JSON` to replace the whole properties
object, including clearing omitted optional fields. The replacement must still
satisfy the selected Type: `{}` clears informational Link properties, while an
Issue still needs its required title and a Memory still needs title and body.
`bd remember --id EXISTING --replace-properties JSON` and `bd remember
--update EXISTING --replace-properties JSON` use the same existing-only Memory
replacement; use `--properties` to initialize a new Memory.
Replacement cannot combine with `--properties`, `--patch`, or Issue field flags.
Issue property edits still use the native Issue writer and History recorder.
Use `bd update ID --patch '[{"op":"replace","path":"/name","value":...}]'`
for ordered property operations, including removal. A patch of an Issue may
change only its writable scalar properties. A stale `--if-revision TOKEN`
refuses a write; omission accepts the current state. See
[Versioning and History](#versioning-and-history) for exact-version reads.

This changes the earlier graph-preview `bd update LINK --properties '{}'`
behavior: an empty object now preserves the existing `note` instead of
clearing it. Remove the note explicitly with
`bd update LINK --patch '[{"op":"remove","path":"/note"}]'` or replace the
whole property object with
`bd update LINK --patch '[{"op":"replace","path":"","value":{}}]'`.

Issue, Memory and informational Link records also have a separate open
metadata object. The default is `{}`. Supply `--metadata JSON` when creating
one, or on `bd update` to merge the JSON object's top-level keys. To set a
typed value or remove a key, use repeatable `--set-metadata KEY=VALUE` and
`--unset-metadata KEY`; those two flags can combine, and unsetting wins.
`--metadata` cannot combine with set or unset. Metadata and an admitted
property or Issue field edit commit together under the same revision choice.
Use `--replace-metadata JSON` on `bd update` or `bd remember` to replace the
whole metadata object; `{}` clears it. It cannot combine with `--metadata`,
`--set-metadata`, or `--unset-metadata`.
Omitting metadata leaves it unchanged; a metadata-only edit is also valid.
An identical edit retains the current revision. Existing-ID `bd remember`
accepts `--metadata`, `--set-metadata`, or `--unset-metadata` with or without a
body and applies them to that Memory. On creation, these flags form its initial
metadata object.

```sh
bd update policy --metadata '{"reviewed":true}'
bd update policy --unset-metadata team
bd update policy --replace-metadata '{}'
bd update links/policy-work --replace-properties '{}'
```

Creating, updating, showing or closing a graph Issue records it as the last
touched Issue. An Issue claimed by `bd ready --claim` becomes last touched as
well; showing a Memory or Link does not replace that marker. At an interactive
terminal, `bd update` and `bd close` may omit the Issue ID and use this marker.
Scripts and agent sessions must supply an ID unless they explicitly set
`BD_LAST_TOUCHED_FALLBACK=1`. `BD_LAST_TOUCHED_FALLBACK=0` disables the
fallback, including at a terminal. `--readonly` does not write the marker.

```sh
bd update work --title 'Move the release branch after review'
bd update policy --properties '{"body":"Land reviewed changes on integration."}'
bd update work --properties '{"priority":1,"description":"Review the release branch."}'
bd update work --patch '[{"op":"replace","path":"/description","value":"Reviewed."}]'
```

`bd close ID...` and `bd reopen ID...` accept one or more local Issues. A
single `--reason` applies to every close target, repeated reasons map to IDs
in input order, and `--reason-file PATH` preserves the file's literal text.
The `bd done ID MESSAGE` alias accepts a trailing reason. A batch keeps
successful changes when another target fails: successful records go to stdout,
per-target failures go to stderr, and the command exits nonzero. Reopen
reports already-open Issues without adding a version. `--force` overrides the
native pinned, holder, blocker and open-child close policies; it does not make
an already-closed Issue change again. `--session ID` records the closing
session, with `CLAUDE_SESSION_ID` as its fallback. On a single Issue,
`--if-revision TOKEN` from `bd show ID --json` refuses a stale close, including
an already-closed retry. It cannot be combined with multiple IDs,
`--suggest-next` or `--claim-next`; omission accepts the current revision.
On a single Issue,
`--suggest-next` reads Issues that this close newly unblocks; it does not claim
them. `--claim-next` claims the highest-priority ready Issue in the same
transaction as the close, only when at least one target actually closes. A
mixed batch may keep a close and its claim while reporting another target's
failure; retrying an already-closed batch makes no new claim or version.
Gate evaluation, `--continue`, molecule advancement and remote routing remain
pending.

```sh
bd close work review --reason 'Finished the work' --reason 'Review complete'
bd close blocked-work --force --session review-session --reason 'Override approved'
bd close blocker --suggest-next --reason 'Dependency finished'
bd close blocker --claim-next --reason 'Dependency finished and next work claimed'
bd close work --if-revision "$revision" --reason 'Finished the work'
bd reopen work review --reason 'Follow-up needed'
```

`bd delete ID` previews deletion of one unreferenced Memory or Issue without
changing storage; `--force` applies it. `bd forget ID` applies Memory deletion
directly and does not accept Issues. Applying either command requires an
explicit write choice, shown here with `--unconditional`. Any live incoming,
outgoing or self-Link makes deletion refuse; `--force` does not cascade.
Memory refusal names the first blocking Link ID and the total count. Issue
refusal lists the incident Link IDs; unlink them explicitly before retrying
with a fresh revision.

```sh
bd remember 'Temporary note' --id scratch
bd delete scratch
bd delete scratch --force --unconditional
# Or, for another unreferenced Memory:
bd remember 'Another temporary note' --id other-scratch
bd forget other-scratch --unconditional
# An unreferenced Issue follows the same preview and guarded apply shape:
bd create 'Temporary task' --id temp-work
bd delete temp-work
bd delete temp-work --force --unconditional
```

Deletion removes current Bead state but reserves its ID and retains prior
snapshots. It does not create a graph deletion version or promise erasure or
restore. Native Issue deletion records its ordinary delete journal event.

Use `bd defer ID...` to set Issues aside and `bd undefer ID...` to return
deferred Issues to open. An undated defer stays in the icebox until undeferred.
`--until` accepts the same date and relative-time forms as ordinary `bd`;
the next `bd ready` after that time wakes the Issue, records a new version,
and clears its defer date. `--reason` appends a line to the Issue's notes.
Assigned Issues retain their assignee through defer and undefer. Neither
command requires a revision flag, but `--if-revision` checks the observed
revision for a single Issue. A repeated dateless defer or undefer is a no-op
when no date or reason changes. `bd undefer` also clears a stale defer date
without changing a non-deferred status.

```sh
bd defer work --until tomorrow --reason 'Waiting on review'
bd ready                       # Wakes work once its defer date has passed
bd undefer work                 # Or restore it explicitly
bd defer work another-work      # Set multiple Issues aside indefinitely
bd show work --json
bd undefer work --if-revision REVISION_FROM_SHOW
```

The graph preview can narrow `bd ready` by priority, Issue classification,
assignee, label (including glob or regex), top-level metadata, excluded Issue
type, or defer state and can sort or limit the result. Listing and atomic
`--claim` use the same native filter. `bd blocked`
accepts the ordinary label, label-any, and exclude-label filters. Both use
the native blocker-aware predicate, then return canonical graph Issue IDs.
An explicitly supplied blank label filter is an error, so a typo cannot turn
an intended narrow query into the whole work queue.

```sh
bd ready --priority 2 --type task --limit 10 --json
bd ready --label release --sort priority
bd ready --metadata-field team=delivery --exclude-type bug --json
bd ready --claim --label release --actor rig.agent --json
bd blocked --label release --json
```

`ready --claim --json` returns a one-element array of complete graph Issue
records, or an empty array when no matching ready Issue exists. An empty claim
does not create a claim lease or revision. As with ordinary `ready`, the lazy
defer-wake before selection may separately version an expired deferred Issue.
The remaining ordinary ready modes and blocked `--parent` remain in
the [draft CLI specification](/reference/graph-cli-specification-draft)
and refuse in this graph slice until their supporting graph behavior is ready.

## Edit Issue notes

An Issue's notes can carry progress across sessions. `--append-notes` adds a
line without discarding earlier notes. `--notes` sets a nonempty value; if it
would replace different, nonempty notes, add `--force` to make that intent
explicit. To erase notes, use `--clear-notes`: `--notes=` is refused, even with
`--force`, because an accidentally empty shell value could erase the record.

```sh
bd show work --json                     # Read the current revision.
bd update work --append-notes 'Review started.' --if-revision REVISION_FROM_SHOW
bd update work --notes 'Revised handoff' --force --if-revision NEW_REVISION
bd update work --clear-notes --if-revision LATEST_REVISION
```

`bd assign ID PERSON`, `bd priority ID LEVEL` and `bd note ID TEXT` use the
same native Issue writer as `bd update`. They accept a current Issue, preserve
other fields, and reject Memory and Link targets. Assign accepts an optional
`--if-revision` graph token and preserves the native live-claim transfer
fence; `--force` is explicit when that fence must be bypassed. The note command
also accepts its ordinary `--stdin` and `--file` text sources. These shortcuts
keep their ordinary Issue-shaped `--json` response; use
`bd show ID --format graph-json` when the complete graph Resource envelope is
needed.

```sh
bd priority work 1
bd note work 'Checklist complete.'
bd assign work alice
```

An Issue edit accepts the current state by default. A stale `--if-revision`
refuses even with `--force`; force only authorizes the notes overwrite. An
identical value or a clear of already-empty notes leaves the revision alone.
Use `bd versions work` and `bd show work --version TOKEN` to inspect prior
snapshots after an accepted change.

## Create, inspect, edit and remove Links

Use `bd types` to find installed Link Types. Select a concrete nominal Type
with `--type types/NAME` or its full local Type URL. `References` can connect
a Memory to a Memory or Issue; its optional common properties are `title`,
`body` and `status`. An explicit `--id ID` is shorthand for `--id links/ID`;
omit it to allocate an ID. Type and endpoint identities remain fixed during
a properties edit.

```sh
bd link policy work --type types/references \
  --id policy-work --properties '{"body":"work follows this policy"}' \
  --metadata '{"origin":"manual"}'
bd link list --incident-to policy
bd show links/policy-work --json
bd update links/policy-work --properties '{"body":"reviewed policy"}'
bd unlink policy-work --unconditional
```

`bd link list --incident-to ID` lists current incident Links with Type first and in-Scope IDs
relative to the Scope URL; JSON retains canonical URLs. `bd link SOURCE TARGET`
without `--type` is an Issue dependency shorthand. Direct References Links
are unowned. Other selected families have source ownership: Memory owns
Subject To and Superseded By, Constraint owns Requires, and Evaluation owns
Conformance claims. The selected nominal subtypes inherit those rules.
Changing an owned Link changes its owner; its target does not change merely
because it is linked. See the [Memory Type CLI facade](/reference/bdp/cli/types/memory/memory)
for constraints and exact Evaluation subject rules.

Link updates accept the current Link by default; `--if-revision TOKEN`
rejects a stale Link. Removal needs `--if-revision TOKEN` or `--unconditional`
and retains the identity and prior snapshots. Native Dependency remains
restricted to live Issues and retains its existing blocking and unlink rules.

For two live Issues, `bd dep add BLOCKED BLOCKER` creates a blocking Link.
The native `bd dep add BLOCKED --blocked-by BLOCKER` and `--depends-on BLOCKER`
spellings create the same edge. On one edge, `--no-cycle-check` suppresses only
the post-add warning; it does not bypass the writer's cycle check.
`bd dep BLOCKER --blocks BLOCKED` is the same operation with the arguments in
the opposite order. Add `--id links/ID` to `bd dep add` when the Link needs a
chosen local identity; repeating the same pair with that ID or no ID is a
no-op, while a different ID refuses. `bd dep remove BLOCKED BLOCKER` (also
`bd dep rm`) removes the blocking pair through the same Issue writer as
canonical `bd unlink links/ID`. An absent pair reports `not_found` without
adding Issue History or a Link tombstone; an actual removal records one native
dependency event and source version. The target Issue stays unchanged.
Bare native Issue IDs are resolved within the same removal transaction when
their graph paths differ. A bare ID that also names a different Bead refuses
as ambiguous; use `beads/PATH` or its full canonical URL to select explicitly.
`bd unlink` continues to require an exact Link ID and revision guard for a
blocking Dependency. `bd dep relate` and `bd dep unrelate` are outside this
graph preview because the ordinary command owns two directed relationships,
not one Link.

## Claim and release Issue work

`bd update ID --claim` atomically claims a live Issue for the current actor.
`bd unclaim ID...` releases assigned open or in-progress Issues. By default
the current actor must hold each claim. Each successful release clears its
assignee, lease, and started time, returns it to open, and retains one native
Issue version and one graph version. A second unclaim refuses because no claim
remains; it does not create another version. A batch attempts each ID and exits
nonzero if any release fails.

```sh
bd update work --claim --actor rig.agent
bd unclaim work --actor rig.agent --reason 'Handing this back'
bd comments work              # The reason is a native Issue comment
bd update work --claim --actor another.agent
bd unclaim work --if-assignee another.agent --actor supervisor
```

`--if-assignee HOLDER` releases only while that holder remains assigned; a
mismatch leaves the Issue and its lease untouched. `--force` bypasses the
holder check for an abandoned claim, but still uses the native row
compare-and-swap. The two flags cannot be combined. `--reason TEXT` appends a
native Issue comment after the release; if that separate append fails, the
release remains committed and the CLI warns. Comments are a separate feed,
outside retained Issue state, and do not mint another Issue version.

Add comments with either familiar command, then list the feed:

```sh
bd comment work 'Ready for review'
bd comments add work 'Reviewer confirmed' --author alice
bd comments work --json
```

`bd comment` accepts positional text, `--stdin`, or `--file`; `bd comments add`
accepts positional text or `--file` and its ordinary `--author` override. Both
return the stored native comment with `--json`, using the Scope-relative Issue
ID in `issue_id`. `bd comments ID --json` retains the ordinary comment-array
shape. Comment writes leave the Issue revision and retained version list
unchanged.

## Versioning and History

The [revision-aware CLI reference speclet](/reference/bdp/cli/references)
defines this candidate's `ID@TOKEN` and `'[URI]@TOKEN'` inputs, exact `show`
reads, and refusal of unsupported endpoint pins. Use a full canonical URI for
an existing ID containing raw `@`. JSON and stored identities keep revision
separate. These additions require candidate qualification before release.


**Revision** names the current state of one Bead or Link. A complete graph
record has one opaque `revision` field. Use it with `--if-revision` when a
write must apply only to the state you read. A different current revision
means the state changed and the guarded write refuses. **Version** names a
retained state of that Resource. `bd versions` rows and discovery summaries
call its exact-read address `version`; use that token with `--version`, or as
an operand to `bd compare`, to retrieve or compare the state later. A current
record's `revision` is also its retained-state address. These names express
two uses of one opaque token, not separate counters. The token encodes no
time or order and is not a Dolt commit ID.

Save a token from a record or `bd memories --details` to read that exact
retained state later. Version reads do not depend on the record still being
current. `bd compare` compares two **chosen** retained versions; the token
order you give it sets the comparison direction. Compare does not discover
their chronological order.

```sh
bd memories 'code flow' --details
bd recall policy --version SAVED_TOKEN
bd show policy --version SAVED_TOKEN --json
bd compare policy --from FIRST_TOKEN --to SECOND_TOKEN --json
bd versions policy
bd history policy  # same listing in a graph workspace
```

| Flag | Current use |
| --- | --- |
| `--version TOKEN` | Select one exact retained state for `bd show`, or one retained Memory body for `bd recall`; use a saved `version` token. |
| `--from TOKEN --to TOKEN` | Select the two complete states for `bd compare`. |
| `--if-revision TOKEN` | On a supported write, refuse if the current record no longer has the saved `revision`. In an ordinary workspace, `bd update` and `bd delete` instead read it as the legacy compare-and-swap on a decimal bead `revision`. |
| `--unconditional` | On deletion and Link removal, explicitly accept the current record without an expected revision. Graph deferral also accepts it as an explicit spelling of its current-state default. Ordinary edits no longer use this flag. |
| `--if-source-revision TOKEN` | On a Memory-owned Link write, optionally require the source Memory's observed revision; otherwise that source defaults to unconditional acceptance. |

`bd remember` with an existing `--id` or `--update`, `bd update`, and `bd link
update` default to accepting the current state. Memory deletion and Link
removal still have their command-specific guard
requirements; consult the [Graph CLI Specification (Draft)](/reference/graph-cli-specification-draft) before
automating them. A semantic no-op retains the existing revision.

Structured graph records carry an `attribution` object only when a principal
was recorded. This applies to reads such as `bd show`, `bd compare`,
`bd memories --format records-json` and `bd list --format records-json`, and to
mutation results from `bd remember`, `bd create`, `bd update`, `bd link`,
`bd unlink` and `bd delete`, including nested replaced-source records.
Its `basis` is `writer-supplied` when that version's writer
supplied the actor, or `unknown` when the relationship to the version cannot
be established. `basis` describes the source of the value; it does not assert
that an authority authenticated the actor. The stored snapshot format retains
its older `status` field, while carried CLI record and BDP Read projections use `basis`.

`bd versions ID` lists a Memory, Issue or Link's versions newest first in a
graph workspace. `bd history ID` is an alias there; in an ordinary workspace,
`bd history` retains its Dolt-commit meaning. Each graph row includes an
opaque `version` token, a store-local `local_revision`, a display `change_at`
time, and attribution. The `bd versions --json` row's scalar `attribution` is
the native per-version status label (`claimed`, `unknown`, or empty), not the
carried-attribution object above. Use the **token** for `show --version` or `compare`, never
the local revision number. Human output calls the number `REV`, matching
ordinary `bd versions`; graph JSON uses `local_revision`. It orders versions
within this store and is not a stable cross-clone address. `change_at` is not
the ordering authority. A removed Link still lists its prior citable versions;
the private deletion marker is not a Resource version and is omitted. Memory
deletion also adds no version. There is still no
BDP HTTP History, as-of selection or restoration. `bd status --graph` reports
`versionList: true`; `historyExact: false` refers to the unavailable HTTP
History profile. The [technical reference](/reference/graph-cli-specification-draft#list-a-resources-versions)
details ordering and refusal behavior.

### Canonical retained-reader output

In a graph workspace, `--output-format` opts retained readers into the canonical
output contract. Existing `--json` and `--format` behavior remains unchanged.
Do not combine either legacy flag with `--output-format`, even when both request
JSON; the command refuses the ambiguous contract choice.

| Commands | Canonical formats |
| --- | --- |
| `versions`, `history`, `link versions` | `default` (table), `table`, `detail`, `json` |
| `compare`, `link compare` | `default` (detail), `detail`, `json` |

```bash
bd versions policy --output-format json
bd link versions links/related --output-format table
bd versions policy --output-format json --include /change_at
bd compare policy --from FIRST_TOKEN --to SECOND_TOKEN --output-format json
bd link compare links/related --from FIRST_TOKEN --to SECOND_TOKEN \
  --output-format json --omit envelope
```

Canonical JSON uses `schema_version: 2`, an exact `bdp_spec` identifier, and
`result_kind: "version_collection"` or `"comparison"`. These result kinds
identify CLI-defined structures; they do not claim to implement BDP HTTP History.
`--omit envelope` emits the native result object alone.

Version projections apply to each summary row. An explicit `--include` selects
optional fields; the exact `version` token always remains. The enclosing
`resource` and `kind` values remain present, and human output includes them as
subject context. The native `resource` value is Scope-relative, so interpreting
its subject and version tokens requires the originating workspace Scope.
`--omit /version` refuses. Projection never fetches full
Resource snapshots to fill absent summary fields.

Version output returns the provider's complete bounded result or refuses. It
does not silently truncate at a default page size, invent continuation cursors,
or support JSONL or paging flags. Comparison output preserves the chosen operand
order, exact tokens, differences and coverage. Repeated `--from` or `--to` values
must agree; conflicting values refuse before reading. Its operands are references, so
data `--include` and `--omit` masks refuse; `--omit envelope` remains supported.

An ordinary JSON read failure emits a complete failure envelope and a diagnostic
on stderr, with a nonzero exit status. Raw JSON and human failures use stderr and
the exit status without adding a result object. Output failures after a write
has started do not append a second document.

## Discover installed Types

`bd types` lists the selected workspace's public Bead and Link Types.
Use a printed `types/NAME` selector or full local Type URL with `--type`.
Bare Issue and Dependency classifications keep their existing meaning.

The Preview 3 catalog includes nine Memory Bead Types: Memory, Constraint,
Policy, Decision, Procedure, Requirement, Scenario, Specification and
Evaluation. Its nine Link Types are References, Superseded By, Requires,
Defines, Subject To, Conformance, Addresses, Satisfies and Violates. Native
Issue and Dependency remain available. Conformance is abstract: it appears
in inventory and is a valid exact Type-filter selector, but direct creation
is refused. Exact filters do not promise descendant expansion.

Type, BeadType, LinkType, ConformsTo and Refines are intrinsic and unavailable
as public Type selections or filters. Active `example-*` and `preview-*`
nominal registrations are removed; existing configured native Issue
classification names are preserved. No command installs the retired examples
or upgrades an earlier catalog in place.

```sh
bd types
bd types --details
bd types --json
```

All three formats use the same public visibility rules. `--details` may show
a partial description; it does not promise a complete HTTP Type descriptor.
The Preview 3 `--json` shape has no Preview 3-to-Preview 4 compatibility
guarantee. See the [Memory Type CLI facade](/reference/bdp/cli/types/memory/memory)
for the exact typedef source and candidate qualification boundary. Ordinary
workspaces retain their existing Issue classification inventory.

## Import an ordinary export

The independent legacy import workstream adds bounded `bd import` for a fresh graph workspace. See the [legacy import contract](/reference/graph-legacy-import) for supported data, identity mappings and refusals. This candidate capability is pending release integration; graph-format import follows graph export.

## Canonical Resource output

In a graph workspace, opt into the canonical output contract with
`--output-format`. Root `show` accepts a Bead or Link identity; `link show`
accepts a Link identity. Both use the same BDP Resource representation:

```sh
bd show beads/plan --output-format json
bd link show links/note --output-format json
bd show beads/plan@REVISION --output-format json
bd show beads/plan --output-format detail --omit /metadata
bd show beads/plan --output-format json --include /properties/title
bd show beads/plan --output-format json --omit envelope
```

`default` and `detail` show labeled fields; `json` emits a schema-version-2
CLI envelope with the exact BDP contract identifier and Resource in `data`.
Current records include their inline owned Links. Reading or displaying them
never retrieves target Resources. Exact retained reads use the retained state,
including its owned Links, without filling fields from current state.

JSON Pointer masks are repeatable. Includes select optional fields while retaining
identity, Type and revision; Link endpoints are mandatory too. Omission wins.
Nested paths are supported within `properties` and `metadata`; other top-level
sections, including `ownedLinks`, are selected or omitted as whole values.
Omitting mandatory context refuses the request. Masks do not filter which
Resources exist or change the selected graph.

`--omit envelope` emits the selected Resource directly. A failed raw read emits
no stdout; always check the command's exit status. Framed JSON failures contain
`outcome: "failure"` and an error origin/problem wrapper. Diagnostics use stderr.

Existing `--json` and `--format` behavior remains available. Combining either
explicit flag with `--output-format` refuses, including `--json=false`. These
existing command paths require `--output-format` when using `--include` or
`--omit`. This opt-in currently requires a graph workspace and exactly one
Resource per `show` invocation.

### Canonical collections

Use `bd list` for current Beads and `bd link list` for current Links. Canonical
collections retain the public BDP records, including inline owned Links, without
retrieving endpoint Resources:

```sh
bd list --output-format json --unpaged
bd list --output-format table --unpaged --include /properties/title
bd link list --output-format json --unpaged --source beads/plan
bd link list --output-format jsonl --incident-to beads/plan --direction in
bd link list --output-format json --unpaged --omit envelope
```

`--type` selects an exact canonical absolute HTTP(S) Type URI;
`--conforms-to` requires the same URI grammar and also admits Types conforming
to that Type. Use the full Type identity, for example
`https://example.test/project/types/note`. Canonical selection does not resolve
Issue classifications such as `bug`, bare names, `types/NAME`, or Dependency
aliases; those remain available on legacy paths. Link selection accepts
`--source`, `--target`, and `--incident-to`.
An incident selection can use `--direction in`, `out`, or `both`. These filters
combine with AND. Canonical collection selection includes every current matching
Bead, without a default filter for Issue status. It refuses Issue workflow filters, including `--all`, and
does not accept the textual query language. Legacy `bd list --all` retains its
selection meaning; it is not a pagination flag.

`default` and `table` display generic identity columns; `detail` displays each
selected record. The same JSON Pointer masks select optional fields for every
format. Framed JSON puts the native `{items, next}` collection in `data`, with
completion information alongside it. `--omit envelope` returns that native
collection directly.

JSONL defaults to the complete selection. Its first record is a header, each
Resource has an item record, and its last record reports completion or failure.
With `--omit envelope`, JSONL contains only Resource records: no header, terminal,
or error record. Check the exit status even when some records were written.

The current local provider supports a complete selection from one snapshot,
with limits of 1,000 live Resources and 16 MiB of current-state inventory.
Table, detail and JSON default to a page boundary of 50; use `--unpaged` for a
larger complete selection. `--page-size` requests a per-page delivery boundary;
`--limit` requests a total delivery bound, with `0` meaning unlimited. This complete-result provider can
satisfy a requested bound only when the whole selection fits. It does not yet
support returning a smaller subset: if the selection exceeds the requested page
or total bound, the adapter refuses with `capability_unavailable` and emits no
items after reading the selection. In particular, the default JSON/table/detail
invocation refuses for 51 or more matching records. Use `--unpaged --limit 0` or
narrow the selection. The failure identifies the unsatisfied bound; it does not
return a truncated collection as exhausted.
`--continue` also refuses because this local provider cannot retain a cursor
across CLI invocations. Persistent paging is not yet available on these paths.

These collection controls require `--output-format`, except the existing root
`list --limit` flag, whose legacy behavior is unchanged. Combining canonical
output with explicit `--json` or `--format` refuses. The Issue/dependency commands
and Memory facades keep their existing output contracts.
