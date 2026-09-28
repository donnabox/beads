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

## Limits and remaining work

Memory deletion, `list`, `blocked`, claims, assignment edits, estimate/reference/
date edits, notes changes, label mutation, ordered property patches, generic
traversal and full Memory remain unavailable. Issue creation can set initial
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
