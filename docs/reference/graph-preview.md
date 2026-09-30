# Mixed Memory and Issue graph preview

This integration checkpoint supports a bounded Memory/Issue workflow in a
**fresh, explicitly selected graph workspace**. Existing ordinary Issue
workspaces continue using their existing commands and storage. The generic
model, Type names and result shapes remain an experimental preview.

Start in a new directory with no `.beads` directory:

```sh
git init
bd init --graph-mode link --scope-url https://example.org/team/ \
  --skip-hooks --skip-agents --non-interactive
bd remember 'The release uses the integration branch.' \
  --id beads/plan --title 'Release plan' --json
bd show beads/plan --json
# A separate invocation reopens the same stored Memory.
bd show https://example.org/team/beads/plan --json
bd status --graph --json
```

The Scope URL establishes local identity; initialization does not publish a
web server at that address. Local `beads/PATH`, `links/PATH` and their exact
Scope URLs identify records. Aliases and foreign Scope URLs are unavailable.
Creation uses an explicit canonical Bead path; an allocated identity cannot
be reused for a different record.

For an ordinary external Dolt SQL server, add these options to `bd init`:

```sh
--server --external --server-host 127.0.0.1 --server-port 3306 --server-user root
```

Use the normal connection configuration appropriate to your server. Embedded
and server modes use the existing driver and standard schema initialization;
there is no manual schema seeding step. Graph initialization records and checks
workspace identity and refuses existing workspaces, mismatched bindings and
incomplete initialization. This preview does not adopt an existing Issue
database. Different-database provisioning on Dolt 2.1.8 must be serialized.

## Instructions for agents

Omit `--skip-agents` when initializing a graph workspace to install a managed
graph instruction block in `AGENTS.md` (or the configured agents filename).
The block incorporates Stephanie Jarmak's durable-memory guidance from
[integration PR43](https://github.com/versioned-beads/beads/pull/43), with
commands supported by this graph preview:

```sh
bd remember "Use UTC for timestamps" --id beads/time-policy --title "Timestamp policy"
bd recall beads/time-policy
bd memories timestamps --format records-json
bd status --graph
```

Choose a distinct canonical ID for each new fact. These commands demonstrate
storage and retrieval across invocations; installing instructions does not
guarantee that an agent will decide to save a fact.

Graph initialization preserves surrounding user-authored text and an existing
`CLAUDE.md` import of `@AGENTS.md`. An existing minimal managed block is replaced
with graph instructions; pass `--skip-agents` to preserve that block unchanged.
It installs no agent hooks or separate Claude instructions. Use your agent's existing support for the configured instruction
file. The ordinary `bd prime` and `bd setup` commands remain unavailable in
graph mode.

Existing full or unknown managed profiles, malformed or duplicate managed
blocks, and symlink or nonregular targets refuse before database initialization.
Keep those files and pass `--skip-agents` to initialize without changing them.
If instruction publication fails after database initialization, the workspace
remains incomplete and fenced; this preview provides no repair command.

The ordinary minimal profile also includes Stephanie's exact durable-memory
contribution, which changes its managed content hash. Existing refresh policy
is unchanged; this addition does not independently rewrite instruction files.

## Supported graph commands

| Command | Admitted scope and flags |
|---|---|
| `remember BODY --id beads/PATH --title TITLE` | Memory creation. An explicit `--body-file PATH` or `--stdin` replaces the positional body source. These sources are mutually exclusive; empty text is present content. |
| `remember --update BEAD` | Change only supplied `--title` and/or one explicit body source, preserving omitted fields inside the transaction. Requires `--if-revision TOKEN` or `--unconditional`. |
| `memories [SEARCH]` | Complete bounded Memory title/body search summaries. Supports `--all`, `--details` and `--format table\|records-json`; legacy `--json` refuses. |
| `recall BEAD` | Stream one Memory's exact body bytes. Optional `--version TOKEN` selects a retained body. `--quiet` does not suppress content; `--json` refuses. |
| `update BEAD --properties JSON` | Replace a Memory's complete properties with exactly the `title` and `body` strings. Requires `--if-revision TOKEN` or `--unconditional`. |
| `update RESOURCE --patch JSON` | Apply ordered `add`, `replace`, and `remove` property operations to one Memory or informational Link. Accepts literal JSON, `@file`, or explicit `@-` stdin. Requires a Resource guard and a separate source guard for a Memory-owned Link. |
| `delete BEAD` | Read-only preview of deleting one unreferenced Memory. `--force` applies and requires `--if-revision TOKEN` or `--unconditional`. A preview needs no guard but checks any supplied guard. |
| `forget BEAD` | Apply the same unreferenced Memory deletion immediately, with `--if-revision TOKEN` or `--unconditional`. |
| `create TITLE --id beads/PATH` | Create an Issue. Allows `--title`, inline `--description`/`--body`/`--message`, `--type`, `--priority`, `--labels`/`--label`, inline `--design`, `--acceptance`, `--assignee`, `--estimate`, `--external-ref`, `--spec-id`, `--due` and initial `--notes`. Existing classification rules apply. Initial status is open. Ordinary creator identity and git-email Owner defaults are included in the Issue data. |
| `update BEAD` with Issue scalar flags | Inline `--title`, `--description`/`--body`/`--message`, `--design`, `--acceptance`, `--priority`, non-claim `--assignee`, `--estimate`, `--external-ref`, `--spec-id`, `--due` and literal `--append-notes`. Requires `--if-revision TOKEN` or `--unconditional`. Description aliases must agree. Files/stdin and other Issue fields are unavailable. |
| `update BEAD --claim` | Atomically claim one Issue for the current actor using the native writer. Standalone `--claim=true` only; no other edits or revision/force guard. Repeating the same actor is a no-op and does not renew its five-minute lease. |
| `show RESOURCE` | Current Memory, Issue or Link; optional `--version TOKEN` selects an exact retained record. No chronological History option. |
| `compare RESOURCE --from TOKEN --to TOKEN` | Compare two complete retained preview versions of one Memory, Issue or Link. Explicit tokens determine direction, not chronology. |
| `link SOURCE TARGET --resource-type TYPE` | Use an exact installed Link Type URL. Informational Links permit `--id links/PATH`, `--properties JSON` and source guards. Memory sources own informational Links; Issue sources do not. |
| `dep add SOURCE TARGET` or `link SOURCE TARGET` | A local blocking Dependency between Issues, using the ordinary default `blocks` type. No bulk, remote, routing or bypass flags. |
| `update LINK --properties JSON` | Replace all informational Link properties. Requires a Link guard and, for a Memory-owned Link, a source guard. Blocking Dependency properties are not editable here. |
| `links BEAD` | Complete bounded current incident Links, with optional `--direction in\|out\|both` and exact `--resource-type TYPE` filter. No pagination. |
| `unlink LINK` | Remove one informational Link or blocking Dependency by canonical ID. Requires a Link guard and, for an owned Link, a source guard. |
| `unlink SOURCE TARGET --resource-type TYPE` | Remove an unambiguous informational Link with the same guards. Multiple matches refuse and report candidate IDs. Blocking Dependency pair removal is unavailable. |
| `close BEAD` | Close one Issue through the existing Issue policy, optionally with ordinary reason aliases. No force or batch operations. |
| `reopen BEAD` | Reopen one Issue, optionally with `--reason`. |
| `ready` | Unfiltered current ready Issues through ordinary readiness rules. No list filters, output limit or configured positive `BEADS_MAX_ROWS`. |
| `list --flat` or `list --format records-json` | Current complete Issue records with status/type/title/priority/assignee/label/pinned and `--due-before`/`--due-after`/`--overdue` filters and explicit limited-page `hasMore`. Tree and legacy JSON remain unavailable. |
| `blocked` | Complete native dependency-blocked Issue view with canonical blocker IDs. No filters or positive `BEADS_MAX_ROWS`. |
| `graph BEAD --view generic` | Current local summary traversal with `--direction in\|out\|both`, `--depth`, `--max-nodes` and `--max-links`. |
| `status --graph` | Report the capabilities and bounds admitted by this checkpoint, including initial Issue fields/notes, append-only notes, estimate/reference edits and due-date authoring/filtering. |
| `serve --readonly --addr HOST:PORT` | BDP Read over HTTP for an ordinary shared-server graph workspace. Existing token-file authentication, Host controls and non-loopback opt-in apply. Embedded serving is refused. |

Commands accept the common graph controls `--json`, `--graph-mode`, `--actor`,
`--quiet`, `--no-color`, `--directory` and `--readonly` where applicable. A
read-only invocation cannot write. Unsupported command options refuse rather
than falling through to ordinary storage operations.

The installed informational Type is
`https://example.org/team/types/preview-related-v2` for the example Scope;
the blocking Type is `https://example.org/team/types/preview-blocks-v1`.
Arbitrary Type installation and friendly generic Type names are not available.
For example:

```sh
bd remember 'Context for the plan.' --id beads/context --title Context
bd create 'Ship the release' --id beads/work --priority 1
bd link beads/plan beads/context \
  --resource-type https://example.org/team/types/preview-related-v2 \
  --id links/context --properties '{"note":"background"}' \
  --unconditional-source
bd link beads/plan beads/work \
  --resource-type https://example.org/team/types/preview-related-v2 \
  --id links/work --unconditional-source
bd links beads/plan --json
bd update links/context --properties '{"note":"revised background"}' \
  --unconditional --unconditional-source
bd unlink links/context --unconditional --unconditional-source
```

`--unconditional` explicitly accepts the current record; use an observed
`--if-revision TOKEN` to reject a stale write. Source guards are
`--if-source-revision TOKEN` or `--unconditional-source`. Memory-owned Link
writes require a source guard as well as the Link guard where applicable.
An unconditional changed Memory write, including an owned-Link change,
discloses the actual replaced version and attribution. Guarded writes and
semantic no-ops do not claim an unconditional overwrite. Targets do not acquire
new versions merely because another Bead links to them.

Unlink removes current Link state while reserving its identity and retaining
private prior snapshots. It does not erase a Memory or promise irreversible
erasure, restoration or identifier reuse.

## Ordered property changes

`--patch` applies a nonempty ordered array to the actual guarded predecessor
inside the existing write transaction. It cannot be combined with
`--properties`, selected Memory fields or Issue update flags. For example:

```sh
bd update beads/plan --if-revision OBSERVED_MEMORY_REVISION --patch \
  '[{"op":"replace","path":"/title","value":"Release plan"},{"op":"replace","path":"/body","value":"Updated context."}]' --json
bd update links/context --if-revision OBSERVED_LINK_REVISION \
  --if-source-revision OBSERVED_MEMORY_REVISION \
  --patch '[{"op":"add","path":"/note","value":"Updated background"}]' --json
bd update beads/plan --if-revision OBSERVED_MEMORY_REVISION --patch @changes.json
```

Use fresh observed tokens for each changed write. `--unconditional` accepts
the current Resource; for a Memory-owned Link, `--unconditional-source` is a
separate decision. An Issue-source informational Link does not require a
source guard, but any supplied source guard is checked. Patching that Link
does not change its Issue source. Issue properties and blocking Dependency
properties cannot be patched through this route.

Operation objects contain only `op`, `path`, and, for add/replace, `value`.
Pointers use JSON Pointer escapes (`~0` for `~`, `~1` for `/`); an empty
pointer selects the root. Add inserts array items or replaces an object
member, and `-` appends to an array. Replace/remove require an existing
target; missing parents are never synthesized. Operations run in order and
may temporarily introduce arrays or other JSON values. The final Memory
properties must contain exactly the `title` and `body` strings; final Link
properties permit only an optional string `note`. Empty strings are present
values, distinct from absent members or null.

All operations succeed together or leave the stored state unchanged. A
same-value or reversing patch is a semantic no-op: the existing version,
revision and attribution remain unchanged, with no replacement disclosure.
Stale guards still refuse before no-op evaluation. A changed unconditional
write discloses the actual replaced Memory state, including when the change
is to a Memory-owned Link. Prior complete states remain available through
the exact retained reads described below. Deleted Memories and removed Links
cannot be revived by patching.

Input and each working properties document are limited to 1 MiB, with at
most 256 operations, 4,096 pointer bytes, and 64 pointer segments/nesting
levels. A cumulative 16 MiB evaluation-byte limit bounds repeated document
work; it is not an exact heap or timing bound. Duplicate members, invalid
Unicode, inexact unsupported numbers, extra operation members and unsupported
operations refuse. Files/stdin are read only when explicitly selected and
after readonly/freeze, selector and local guard admission. Input/parse errors
report `invalid_properties`; stale guards report `revision_conflict`.

Changed patches must also fit the complete current-read acquisition budget,
including retained final heads of deleted Memories. Failure rolls back the
mutation and retained records. Existing replacement/selected-field writers
keep their established admission policy. A no-op writes nothing and does not
repair an oversized workspace; a properties document already above the
patch limit cannot use this route to shrink itself. Runtime bound refusals
report `capability_unavailable`. These bounded CLI changes do not enable a
public BDP Write profile, History ordering or durable request outcomes.

## Find and inspect retained Memory

`memories` searches current Memory titles and bodies using literal Unicode case
folding. It reads one checked current snapshot before filtering. The default
must match at most 50 Memories; a larger result refuses instead of returning a
partial page. Narrow the search or use `--all` for all matching summaries within
the same acquisition and output bounds. `--details` adds owned-Link counts.

```sh
bd memories release
bd memories release --details --format records-json
# Copy a selected item's canonical id and version from the summary.
bd recall beads/plan --version SAVED_TOKEN
bd show beads/plan --version SAVED_TOKEN --json
bd compare beads/plan --from EARLIER_SELECTED_TOKEN --to OTHER_SELECTED_TOKEN --json
```

Summaries include identity, title, saved version, attribution, matched fields
and a body excerpt when the body matches. Search input is limited to 4,096 UTF-8
bytes, excerpts to 160 code points and final output to 1 MiB. A short matching
body can appear in full as its excerpt. Results are sorted by canonical ID,
marked complete and have no continuation. The existing 1,000 live Resource and
16 MiB acquisition bounds apply before filtering, including retained deleted
Memory heads counted by the current reader.

Use `--format records-json` for experimental summaries; `--json` has no settled
graph Memory-list mapping and refuses. Explicit `--format table` or
`--format records-json` takes precedence over ambient JSON configuration.
Ordinary KV-memory workspaces retain their existing output and `--format json`
alias. BDP Read remains the protocol interface for scripts.

`recall` writes the selected body's exact bytes with no envelope, added newline
or quiet suppression. Empty content succeeds with zero output bytes. Without
`--version`, it reads current Memory; with a token from a saved result it reads
that retained body even after edits or deletion. Graph JSON recall remains
unavailable because the complete Memory representation is unresolved.

`show --version` also supports retained Issue and Link records, including their
saved owned Links. Tokens are opaque nonempty UTF-8 strings of at most 4,096
bytes. Local ordinals, timestamps, full version URLs and as-of selection are
not interpreted. An unknown or other-resource token returns `revision_unknown`;
a missing subject returns `not_found`. A removed Link's prior live versions
remain readable; its private removal token returns `gone`. A deleted Memory
retains its final live head without inventing a deletion version.

`compare` reads both operands in one authority-checked transaction. Equal tokens
still require a valid retained record. Either unavailable or corrupt operand
fails the whole operation. Differences include complete property values and
owned Links matched by canonical identity, with explicit presence so absence,
null and empty content remain distinct. Arrays retain order; object member
order is ignored; JSON numbers are compared without float rounding. Reversing
the tokens reverses the comparison. The selected versions and attribution are
context rather than property differences. Immutable identity or endpoint
mismatches refuse.

Comparison output is provisional JSON, indented for human output. Its
`compared` and `unsupported` fields describe the projection: common metadata and
Memory inception/derivation are unavailable. Issue comparison uses the native
retained durable projection; current row-lock and content-hash fields are not
reconstructed. Each retained snapshot acquisition has a 16 MiB input budget. A deleted
Memory also validates its final live snapshot before reading an older selected
snapshot, so a distinct pair can acquire up to 64 MiB in total (up to 32 MiB
for equal older tokens, which are resolved once). This excludes decoding/output
overhead and is not a total-heap or rendered-output cap. These read commands work under `--readonly` and migration freeze. They
do not enable HTTP History, authoritative ordering, restoration or a stable
public comparison contract; `historyExact` remains false.

## Unreferenced Memory deletion

`delete` previews one current Memory without changing it, including under
`--readonly`. Apply with `delete --force` or `forget`; both require the existing
revision guard or an explicit unconditional choice. The `result.memory` in
both preview and apply is the checked final live Memory, including its existing
revision and attribution. `result.preview` distinguishes the read-only preview;
`result.deleted` is true only after successful apply. The outer `preview: true`
continues to identify the experimental graph format, including applied writes.

```sh
bd remember '' --id beads/scratch --title Scratch --json
bd delete beads/scratch --readonly --json
# Supply the revision observed in result.memory.revision from the preview.
bd delete beads/scratch --force --if-revision OBSERVED_REVISION --json
# Or apply to another unreferenced Memory with explicit unconditional intent:
bd remember 'Temporary note' --id beads/another-scratch --title Scratch
bd forget beads/another-scratch --unconditional --json
```

An empty body is present content and can be deleted. Successful deletion removes
current Memory state while preserving the canonical ID reservation and immutable
prior snapshots. A new `show` reports `gone`; recreating the same ID reports
`identity_reserved`. Repeated deletion refuses. Deletion does not create a new
live revision or a deletion version, and the result does not claim a deletion
timestamp or actor event. Internal snapshot retention does not expose public
History or restoration.

Every live incident Link causes both preview and apply to refuse with
`deletion_policy_unresolved`, including incoming, outgoing and self-Links.
Explicitly unlinking each Link with its normal guards allows a later deletion;
these are separate transactions, and a concurrently added Link can still cause
deletion to refuse. `--force` does not cascade. Batch selectors, `--from-file`,
`--cascade`, `--dry-run`, erasure and Issue deletion are unavailable in this
graph slice. Ordinary Issue deletion and key/value `forget` remain unchanged.
`status --graph` advertises `memoryUnreferencedDelete`; general `memoryDelete`
remains false because linked Memory deletion policy is unresolved.

## Read-only Issue queries and traversal

```sh
bd list --format records-json --limit 1
bd list --flat --all --sort title
bd blocked --readonly --json
bd graph beads/plan --view generic --direction out --depth 2 --json
```

Issue listing reuses the existing native query, configuration and limit policy.
It accepts status (or state), type, title/title-contains, priority and priority
range, assignee/no-assignee, label/label-any/exclude-label, pinned/no-pinned and
due-before/due-after/overdue filters. Sorting accepts
priority, created, updated, title, status or type; reverse is supported. Explicit
limit wins over `--all` and configured limits. The result contains complete
canonical Issue records in `items` and a truthful `hasMore` boolean. It is a new
read each time, not a snapshot cursor or BDP continuation. Returned records and
the extra probe record are validated before trimming the page.

Use explicit `--flat` for quoted human summaries or `--format records-json` for
the experimental graph envelope. Bare tree, `--json`, `--format json`, watch,
readiness, parent/ID/routing/offset selectors, repeated status/state/type/assignee filters
and supplied-empty labels refuse. Defer filters remain unavailable; due selection
does not schedule or wake work.
This does not change ordinary list parsing or implement contributor-owned filter
unions. Records-json selects structured errors and overrides ambient human
format selection; explicit `--json` still refuses. BDP Read remains the documented
script interface below; these CLI result shapes are experimental.

`blocked` follows the existing native dependency-blocked query. It is not the
complement of ready, nor a query for every manually blocked or deferred Issue.
The result is an array of objects with complete `issue` records and canonical
`blockedBy` IDs, sorted by canonical ID. Empty is `[]`. It accepts no positional
selectors or filter flags, even explicit empty values. A positive
`BEADS_MAX_ROWS` refuses; zero or unset permits the complete bounded view.
No query wakes deferred work, repairs blocked state, creates versions or opens
the ordinary store. Readonly and migration freeze permit these reads.

Generic traversal reads one checked current snapshot. Nodes expose only ID,
Type, title, version and attribution; Links expose ID, Type, source, target,
version and attribution. Memory bodies, Issue long text and Link properties
are omitted. Directions are in/out/both; depth defaults to 1 and accepts 0..1000.
The root counts as one node. Default node/Link caps are 100/200, each accepting
1..1000. Distinct Link identities survive repeated-node visits, including diamonds
and self-Links. `frontier` identifies reached nodes with eligible unexpanded
Links; `complete` says whether that frontier is empty. Exceeding a node/Link cap
refuses rather than silently truncating. Foreign roots, external endpoints,
custom Types, legacy graph options and positive `BEADS_MAX_ROWS` refuse.

All three views retain the shared whole-workspace 1000-live-Resource and 16 MiB
acquisition bounds, including unrelated content and retained final heads of
deleted Memories. List config acquisition/resolved policy is additionally bounded
to 64 KiB and 256 rows/entries; database failures refuse instead of silently
choosing YAML fallback. List and blocked output have a 16 MiB bound; traversal
summary output has a 1 MiB bound. These are operational bounds, not heap guarantees.
Output is prepared before emission; physical output failures can still occur
after some bytes have been written.

A valid unreferenced Memory deletion does not invalidate an Issue query. Fresh
views exclude it, while its identity and final live retained state remain reserved.
Traversal cannot emit a deleted endpoint; a removed root is absent from its live
snapshot. Corrupt deleted allocations still refuse, including an invalid retained
head or an inconsistent current payload/Link. No new tombstone representation,
public History surface or external traversal contract is introduced.

## Priority and assignment

```sh
bd show beads/work --json
bd update beads/work --priority P0 --assignee alice --if-revision OBSERVED_REVISION --json
bd list --assignee alice --format records-json --all
bd update beads/work --assignee= --if-revision NEW_REVISION --json
bd list --no-assignee --format records-json --all
```

Use the opaque revision from the preceding `show` or accepted mutation. The
revision covers the Issue and its owned blocking Dependencies. Priority uses
ordinary P0–P4 meanings; omission preserves it, and zero explicitly sets P0.
Assignment preserves the literal value, including whitespace; an explicit empty
value clears it. Text, priority and assignment can be combined in one guarded
operation and one retained version. A matching same-value edit is a no-op;
a stale revision still refuses, even if every requested value is already current.

Assignment does not claim work, start a lease or change status. The ordinary
active-holder fence still applies to guarded and unconditional transfers. An
authorized transfer or clear removes the prior lease; unrelated edits preserve
it. Clearing an in-progress assignment does not reopen the Issue. Graph claim is
available only in the standalone form below; heartbeat and reclaim remain unavailable. Changes made through an external lease
writer have no supported graph-preview repair path.

Assignee listing reuses native SQL comparison/collation, not actor identity
normalization. `--no-assignee` includes native empty/NULL assignments. An empty
`--assignee=` supplies no assignee restriction; combining a nonempty assignee
with `--no-assignee` intersects the filters and returns no matches. Listing does
not modify state or claim work. Its output remains the experimental CLI envelope;
BDP HTTP Read is the script interface and exposes the current Issue properties.

## Append-only Issue progress

```sh
bd show beads/work --json
bd update beads/work --append-notes 'Finished the first pass.' --if-revision OBSERVED_REVISION --json
```

The native transaction appends to the current notes. Omission preserves notes;
empty on empty is a no-op, while empty on nonempty appends one newline. Text is
literal, including whitespace, Unicode, CRLF, `-` and `@file` markers; this flag
does not read files or stdin. Repeating nonempty text appends it again. A stale
revision refuses even when the append would otherwise be a no-op. An unknown
commit outcome must be inspected; never automatically replay an append.

Append can accompany supported scalar edits in one guarded update and one
retained version. Notes-only edits preserve owned Dependencies, claim lease
fields and closed state. Assignment mixed with append retains the native holder
fence and lease rules. A changed request supplying append intent must leave the
workspace within its existing read acquisition budget, including its new retained
snapshot, or every effect rolls back. This conservative postcondition also applies
to an empty append combined with a changing scalar. Scalar-only repair paths and
true no-ops keep their existing behavior. No new per-input notes limit is added.

Replacement/clear, force overwrite, and combined claim-plus-append remain
unavailable. Contributor notes replacement safeguards are a separate review gate.
CLI current/exact reads and BDP Read expose the accepted notes; no HTTP write or
public History contract is added.

## Standalone Issue claim

```sh
bd update beads/work --claim --actor alice --json
```

The native claim writer decides eligibility, including configured active statuses
and literal claim-pool aliases. A successful claim sets the assignee and
`in_progress` status, preserves an existing start time, and records one native
version and one complete graph version. It does not require the Issue to be ready.
A foreign holder or non-claimable status is refused without mutation.

The native lease lasts five minutes. Repeating the claim as the same actor
(including equivalent actor spelling) is a no-op, including the lease timestamps.
This preview does not provide heartbeat, renewal, reclaim, unclaim or lease repair.
An external writer that changes the lease bypasses the saved graph projection;
current graph reads and writes then refuse that mismatch. Previously retained
snapshots remain readable. Do not use external lease commands on these preview
Issues expecting the graph to repair them.

`--claim=false`, combined claim plus scalar/property edits, revision guards and
unconditional/force claim flags are refused. A lost commit reply reports an
unknown outcome; callers must inspect rather than automatically retry the write.

## Issue due dates

```sh
bd create 'Prepare the review' --id beads/review --due '2030-01-02T10:00:00Z' --json
bd update beads/review --due '+30min' --if-revision OBSERVED_REVISION --json
bd list --format records-json --due-after '2030-01-01' --due-before '2030-01-03'
bd list --format records-json --overdue --all
bd update beads/review --due= --if-revision OBSERVED_REVISION --json
```

Create and guarded update reuse the ordinary date parser, including `+30min`,
`+6h`, date-only and offset-bearing timestamps. `m` means calendar months;
`min` means minutes. Timezone-less input uses the invoking process's local zone.
Omission preserves the existing due date; explicit empty update clears it.
An empty create value leaves the date absent. Due edits may accompany other
admitted scalar edits or notes append in the same native transaction/version;
standalone claim still cannot be combined with an edit.

This private preview retains the existing whole-second SQL representation:
writes normalize to UTC and round to the nearest second, with half-second ties
forward. Filter cutoffs normalize to UTC and truncate fractions before native
strict `<` / `>` comparisons. A due time of `12:00:00Z` therefore does not match
a before-cutoff of `12:00:00.900Z`. Dates must remain within years 1..9999 after
normalization. These are provisional compatibility rules, not a durable BDP
precision contract. The CLI's 4096-byte UTF-8 date-input bound is likewise a
private preview admission limit.

The ordinary filters intersect, including existing status/type/limit policy.
Overdue means a non-null due date before the native UTC query clock, excluding
closed Issues even with `--all`; it neither changes status nor schedules work.
After normalization, an unchanged due value or repeated clear is a no-op, while
a stale revision still refuses. Current and exact retained reads preserve the
accepted instant; due writes retain existing links, lease and lifecycle state.
No defer, recurrence, mandatory deadline or scheduler behavior is added.

## Limits and remaining work

Linked Memory deletion and Issue deletion,
defer/scheduling, notes replacement/clear, label mutation,
and full Memory remain unavailable. Issue creation can set initial
labels; that does not adopt a label-editing contract. These restrictions apply
to graph workspaces; ordinary Issue workspaces keep their existing behavior.

BDP HTTP Read is available on ordinary shared-server Dolt as described below.
CLI JSON is a separate command output format. No public History route is
enabled by the exact retained CLI reads or experimental comparison.

Writes retain their complete accepted state within the existing transaction.
Those snapshots are not a claim of complete native History, Dolt HEAD/sync
durability, public BDP Write or recovery support. The current read acquisition
budget can refuse an oversized workspace; this checkpoint does not promise
that every successful create preserves that aggregate budget. `status --graph`
reports the current bounds.

## BDP Read from scripts

A graph workspace on ordinary shared-server Dolt can expose its current records
through BDP HTTP. The protocol and independent public client are pinned to
`gastownhall/bdp@53bdbd03136875f952af184fce7b3c7af8f74e96`. This serves the Read
profile; it does not enable HTTP writes or History.

Choose a stable, reachable HTTP address **before** initializing a new workspace.
The persisted Scope URL supplies record identities and the HTTP path. It does
not change when an allowed Host alias or reverse proxy reaches the listener.
For a disposable local example, with an ordinary Dolt server already listening
on port 3306, use a new directory:

```sh
git init
bd init --graph-mode link --scope-url http://127.0.0.1:8765/demo/ \
  --server --external --server-host 127.0.0.1 --server-port 3306 \
  --server-user root --skip-hooks --skip-agents --non-interactive
bd remember 'Why we chose this design.' --id beads/plan --title Plan
bd serve --readonly --addr 127.0.0.1:8765
```

From another terminal, read actual BDP representations:

```sh
curl --fail-with-body -i http://127.0.0.1:8765/demo/
curl --fail-with-body http://127.0.0.1:8765/demo/bdp.json
curl --fail-with-body http://127.0.0.1:8765/demo/beads/plan
curl --fail-with-body 'http://127.0.0.1:8765/demo/beads/plan?view=properties'
curl --fail-with-body 'http://127.0.0.1:8765/demo/beads/?limit=1'
curl --fail-with-body http://127.0.0.1:8765/demo/types/preview-memory-v2
curl --fail-with-body 'http://127.0.0.1:8765/demo/beads/plan?view=links&direction=both'
```

The Scope root returns a `service-desc` Link to `bdp.json`. Collections expose
`items` and `next`; enumerate by following each complete `next` URL until it is
null. A first page alone is not the full inventory. Bead, Link and Type
collections are available at `beads/`, `links/` and `types/`. Bead/Link
collections support `type`, `conformsTo` and bounded `selector` filters; Link
collections also support `source`, `target` and `endpoint`. Incident HTTP
directions are `inbound`, `outbound` and `both`. `include=links` returns a Bead
and its first incident Link page from one snapshot.

GET and HEAD use canonical current identities. Resource/properties responses
carry revision ETags. Authentication, authority, query and storage validation
precede conditional responses: an invalid cursor or missing Resource cannot
become a successful 304. No Last-Modified timestamp is invented. Historical
version queries, aliases, HTTP mutation, legacy `/v0` and `/healthz` routes are
unavailable. Use an authenticated Scope or discovery request for readiness.

A collection continuation retains its original records and owned Link state
even if a later CLI write changes fresh reads. Cursors expire after five
minutes and do not survive service restart. This is retained pagination, not
chronological History. Stop the service before restoring storage; same-binding
hot restore and out-of-band SQL are unsupported.

The [standard-library Python example](https://github.com/versioned-beads/beads/tree/integration/examples/bdp-read)
follows every continuation and can also read one canonical Bead. From this
repository's root, with the example listener running:

```sh
python3 examples/bdp-read/read_beads.py --scope http://127.0.0.1:8765/demo/ --limit 1
python3 examples/bdp-read/read_beads.py --scope http://127.0.0.1:8765/demo/ \
  --id http://127.0.0.1:8765/demo/beads/plan
```

For an authenticated listener, supply its token through the `BDP_TOKEN`
environment variable. The example reads BDP HTTP directly and prints results
only after the requested read or full enumeration succeeds.

### HTTP access and limits

The existing `serve` controls apply: `--auth-token-file`, `--allowed-host`,
`--allow-non-loopback` and explicit `--insecure-no-auth`. The default listener
is loopback; its existing local trust policy permits no-token access there.
For a token-protected listener, provide a bearer token on every request,
including continuations. Token-file rotation keeps the existing last-good-file
reload policy. Every accepted token sees the complete workspace; separate
per-user authorization views are not implemented. Responses use
`Cache-Control: private, no-store`. The server provides no TLS or CORS support.

The store stays open until HTTP requests have drained. Serving opens only an
already initialized, matching ordinary shared-server graph authority; it does
not create a workspace or enable legacy writers. Embedded CLI operations remain
supported, but embedded HTTP serving refuses before opening storage or binding
a listener.

Page size defaults to 100 and is capped at 1,000. The current inventory is
limited to 1,000 live Resources and a conservative 16 MiB persisted-byte
acquisition budget before filtering. That budget can refuse a small or exact
read because other current data is oversized. HTTP representations are limited
to 8 MiB and retained snapshots to 7 MiB; at most 32 snapshots, 4,096
continuation positions and 32 MiB total are retained. Request targets are capped
at 32 KiB, Selectors at 16 KiB, depth 256 and 2,048 nodes. Capacity pressure
refuses new snapshots without evicting valid continuations. These are preview
bounds, not a guarantee about Go heap usage or production migration.

## Regression coverage

The required Graph C0 gate checks source/CLI provenance, storage tests on both
engines, graph configuration and CLI admission with no required skips. Its
installed initialization and fresh-process read checks remain in place. The
mixed workflow adds real CLI coverage for Memory edits, all three informational
endpoint combinations, Link replacement/unlink, blocking readiness and Issue
close/reopen/text editing. Storage tests verify retention, guards, no-ops,
concurrency, rollback and uncertain commit outcomes separately. Ordinary Issue
regression suites remain required. Qualification of any new combined commit
requires the actual destination CI results, not this document alone.

BDP coverage includes storage-to-wire projection on both real engines, bounded
selectors and retained pages, plus an ordinary-server HTTP listener exercising
authentication, token rotation, Host identity, authority loss, conditionals and
concurrent aggregate reads. The independent public-client capture uses the
installed CLI as its writer and real HTTP as its reader. Aggregate/replay
checks use real fetch plus the pinned public parsers where the public client
has no corresponding API. Combined-source qualification must include these
roots and captures without required skips; internal projection tests alone
are not installed HTTP proof.

The unreferenced deletion workflow adds normal installed CLI initialization on
both engines, empty-body preview/apply, stale and missing guard refusals,
read-only admission, fresh-process absence, ID nonreuse, incident-Link refusal
and explicit unlink before deletion. These tests do not implement or qualify
the unresolved linked-deletion policy.

## Issue authoring fields

```sh
bd create 'Implement the plan' --id beads/work --design 'Approach' --acceptance 'Done when verified' --assignee alice --estimate 45 --external-ref 'tracker #42' --spec-id 'spec/section' --notes 'Initial context' --json
bd update beads/work --estimate 0 --external-ref= --spec-id= --if-revision OBSERVED_REVISION --json
bd update beads/work --append-notes 'First pass complete.' --if-revision NEW_REVISION --json
```

Initial fields are part of one native create and retained record. Initial assignment
is not a claim and does not create a lease. Existing Owner and CreatedBy defaults
still apply, including the ordinary git-email Owner data already disclosed above.
Omitted estimates remain absent; explicit zero is present. Estimates use the
existing nonnegative signed SQL INT range; clearing an estimate to NULL is not
admitted. External and spec references are literal strings, not fetched URLs or
alternate Bead identities. The existing columns allow 255 and 1024 Unicode
codepoints respectively. Empty external reference on update clears it to NULL;
empty spec ID clears it to an empty string. CLI create omits an empty external
reference; the Go API preserves an explicitly supplied empty pointer.

Design, acceptance and initial notes are literal UTF-8, including whitespace and
line endings. Changed initial fields must survive storage hydration exactly.
Nonempty initial notes also require the completed current and retained state to
fit the existing workspace read budget before commit. Later notes changes use
append only; replacement and clear remain unavailable. Guarded scalar edits keep
owned Links and unrelated fields unchanged; an identical scalar edit records no
new version. Current CLI records, exact retained records and BDP Read carry these
fields without a new wire representation or HTTP write operation.
