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

## Supported graph commands

| Command | Admitted scope and flags |
|---|---|
| `remember BODY --id beads/PATH --title TITLE` | Memory creation. An explicit `--body-file PATH` or `--stdin` replaces the positional body source. These sources are mutually exclusive; empty text is present content. |
| `remember --update BEAD` | Change only supplied `--title` and/or one explicit body source, preserving omitted fields inside the transaction. Requires `--if-revision TOKEN` or `--unconditional`. |
| `update BEAD --properties JSON` | Replace a Memory's complete properties with exactly the `title` and `body` strings. Requires `--if-revision TOKEN` or `--unconditional`. |
| `delete BEAD` | Read-only preview of deleting one unreferenced Memory. `--force` applies and requires `--if-revision TOKEN` or `--unconditional`. A preview needs no guard but checks any supplied guard. |
| `forget BEAD` | Apply the same unreferenced Memory deletion immediately, with `--if-revision TOKEN` or `--unconditional`. |
| `create TITLE --id beads/PATH` | Create an Issue. Allows `--title`, inline `--description`/`--body`/`--message`, `--type`, `--priority`, `--labels`/`--label`. Existing classification rules apply. Initial status is open. Ordinary creator identity and git-email Owner defaults are included in the Issue data. |
| `update BEAD` with Issue text flags | Inline `--title`, `--description`/`--body`/`--message`, `--design`, `--acceptance` only. Requires `--if-revision TOKEN` or `--unconditional`. Description aliases must agree. Files/stdin and other Issue fields are unavailable. |
| `show RESOURCE` | Current Memory, Issue or Link. No exact-version or chronological History option. |
| `link SOURCE TARGET --resource-type TYPE` | Use an exact installed Link Type URL. Informational Links permit `--id links/PATH`, `--properties JSON` and source guards. Memory sources own informational Links; Issue sources do not. |
| `dep add SOURCE TARGET` or `link SOURCE TARGET` | A local blocking Dependency between Issues, using the ordinary default `blocks` type. No bulk, remote, routing or bypass flags. |
| `update LINK --properties JSON` | Replace all informational Link properties. Requires a Link guard and, for a Memory-owned Link, a source guard. Blocking Dependency properties are not editable here. |
| `links BEAD` | Complete bounded current incident Links, with optional `--direction in\|out\|both` and exact `--resource-type TYPE` filter. No pagination. |
| `unlink LINK` | Remove one informational Link or blocking Dependency by canonical ID. Requires a Link guard and, for an owned Link, a source guard. |
| `unlink SOURCE TARGET --resource-type TYPE` | Remove an unambiguous informational Link with the same guards. Multiple matches refuse and report candidate IDs. Blocking Dependency pair removal is unavailable. |
| `close BEAD` | Close one Issue through the existing Issue policy, optionally with ordinary reason aliases. No force or batch operations. |
| `reopen BEAD` | Reopen one Issue, optionally with `--reason`. |
| `ready` | Unfiltered current ready Issues through ordinary readiness rules. No list filters, output limit or configured positive `BEADS_MAX_ROWS`. |
| `list --flat` or `list --format records-json` | Current complete Issue records with status/type/title/priority/label/pinned filters and explicit limited-page `hasMore`. Tree and legacy JSON remain unavailable. |
| `blocked` | Complete native dependency-blocked Issue view with canonical blocker IDs. No filters or positive `BEADS_MAX_ROWS`. |
| `graph BEAD --view generic` | Current local summary traversal with `--direction in\|out\|both`, `--depth`, `--max-nodes` and `--max-links`. |
| `status --graph` | Report only the capabilities and bounds admitted by this checkpoint. |
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
range, label/label-any/exclude-label and pinned/no-pinned filters. Sorting accepts
priority, created, updated, title, status or type; reverse is supported. Explicit
limit wins over `--all` and configured limits. The result contains complete
canonical Issue records in `items` and a truthful `hasMore` boolean. It is a new
read each time, not a snapshot cursor or BDP continuation. Returned records and
the extra probe record are validated before trimming the page.

Use explicit `--flat` for quoted human summaries or `--format records-json` for
the experimental graph envelope. Bare tree, `--json`, `--format json`, watch,
readiness, parent/ID/routing/offset selectors, repeated status/state/type filters
and supplied-empty labels refuse. Assignee/unassigned and due/overdue filters
also remain unavailable: positive graph fixtures depend on later writer slices.
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

## Limits and remaining work

Linked Memory deletion, Issue deletion, claims, assignment edits and filters,
estimate/reference/date edits and due filters, notes changes, label mutation,
ordered property patches and full Memory remain unavailable. Issue creation can set initial
labels; that does not adopt a label-editing contract. These restrictions apply
to graph workspaces; ordinary Issue workspaces keep their existing behavior.

BDP HTTP Read is available on ordinary shared-server Dolt as described below.
CLI JSON is a separate command output format. No public History route is
enabled by the internal retained-state readers used in tests.

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
