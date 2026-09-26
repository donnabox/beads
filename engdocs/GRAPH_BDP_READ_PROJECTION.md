# BDP current-record projection

This increment follows qualified [PR #23](https://github.com/donnabox/beads/pull/23)
at `9129dfd08690099a0fe2b8d8182322b6912d1587` and the
[fork delivery plan](https://github.com/donnabox/beads/pull/18).
It introduced the storage-to-wire reader. The subsequent
[HTTP preview](GRAPH_BDP_READ_HTTP.md) now composes it into installed shared-server
`bd serve`; this document describes the underlying projection modules. The September 23, 2026
07:01 PDT attempt clock is unchanged.

## What the reader does

`internal/httpapi/graphread.Reader.Resource` reads a canonical local Bead or
Link through `graphstore.Read`, then returns a public `bdpwire` record:

- Canonical IDs, declared Types and opaque revisions remain unchanged.
- Complete properties remain intact, including the Issue domain's own ID
  inside its properties; that field is distinct from the canonical Resource ID.
- Owned Link records come from the same storage snapshot as their source.
  They are grouped by Type and ordered by canonical ID. An empty wildcard
  owner carries `ownedLinks: {}`; an explicit owner carries its declared empty
  groups. No `"*"` group is invented.
- A recorded actor becomes the carried claimed principal without renaming.
  An absent actor yields no attribution. Private timestamps, version aliases,
  `owned` arrays and other preview envelope members do not appear on the wire.
- `Reader.Type` returns the persisted installed descriptor after binding,
  canonical-byte and fingerprint checks. It never reconstructs a replacement
  contract under the same Type ID or fetches a remote descriptor.

Type reads can follow Resource reads because this preview admits only immutable
installed contracts and checks those contracts on every operation. This is not
permission to use separate reads if future mutable installation rules weaken
that invariant. The caller owns the store lifetime and must check close errors.

The underlying `graphstore.ReadType` distinguishes an unknown valid Type path
from a corrupt or missing required installation. Resource absence, deletion and
corruption also remain distinct internal errors. The HTTP adapter must
apply the public problem table: ordinary reads of deleted Resources use
`resource-not-found`; they must not publish the private CLI `gone` envelope or
pretend deletion is erasure.

## Current inventory for collection reads

`graphstore.CurrentSnapshot` captures current live Resources, the four installed
Type descriptors and the controlled-writer token in one read transaction.
`graphread.Reader.Inventory` projects only that captured state; unlike separate
exact reads it needs no later descriptor lookups. Beads, Links and Types are
ordered by canonical URL code units. Owned Links remain complete parts of their
source Bead, with the same contents as the corresponding current Link record.
Removed Links are excluded from the current inventory.

The inventory is an internal input for selection and pagination. It is not a
BDP response, retained snapshot handle, Scope epoch or public cursor. Its token
supports equality-only invalidation across controlled writes. Reads and reopen
leave the token unchanged; successful creates, property updates and unlink
change it. It cannot prove safety against out-of-band SQL changes or restoration.

The isolation test commits a writer after establishing the reader transaction,
then verifies that the reader still sees its original records and token while
a new transaction sees the write. The shared-server test uses separate store
connections. The embedded test admits two SQL sessions on one connector solely
for that test; normal embedded stores retain their one-connection configuration.
It does not claim concurrent independent embedded processes are supported.

The subsequent HTTP slice adds a conservative 16 MiB current-workspace read
acquisition budget before payload decoding, including exact Resource reads; see
[HTTP limits](GRAPH_BDP_READ_HTTP.md#implemented-surface-and-limits).
This preview also refuses inventories above 1,000 live Resources with an explicit
limit error and no partial result. The bound applies before filtering and is
not a page size. It does not truncate or silently skip malformed entries.

## Selection and retained pages

The internal reader now composes collection selection with bounded, process-local
pagination. `CompileCollection` validates the entire query before a store read:
unknown parameters, repeated parameters, malformed limits, noncanonical local
identities and unsupported collection predicates fail. `CompileIncident` covers
the separate Bead Link view, including its inbound/outbound/both directions.
Selection applies all predicates before paging and orders complete records by
canonical URI code units. Type inventory entries are summaries; exact Type reads
still return descriptors. Effective conformance includes the declared Type itself
and captured ancestors, never a remote fetch. The current four installed preview
Types have no parent contracts.

Structural endpoint predicates normalize canonical local Bead IDs and compare
reference URIs without pins. Selectors compare stored JSON values exactly:
`@.source == "https://example.test/beads/a"` does not match a pinned object,
while the structural `source` predicate matches its URI. Selector parsing and
evaluation are iterative, with explicit source-byte, AST-depth and AST-node
bounds. They implement the subset and classified errors of the pinned public
BDP implementation, including existence, scalar comparisons and Boolean
composition. Response-only revision, attribution and owned-state members are
excluded from the candidate root. Diagnostic offsets are UTF-8 byte offsets.

`Pagination` copies the selected JSON bytes and reserves every future cursor
position before returning the first page. A later write cannot change that
snapshot's membership, revisions or owned Link properties. Replaying a cursor
returns the same page and next URL; it remains valid through the original expiry,
including after the final page. Capacity pressure refuses a new snapshot instead
of evicting an active one. Expired/unknown cursors, changed authority epochs,
foreign authorization views and changed collection projections fail explicitly.
Concurrent calls are synchronized. Cleanup is lazy and bounded; `Close` releases
all retained state. Process restart loses cursors and requires a fresh first read.

Provisional internal defaults are 100 items per page, maximum 1,000; a five-minute
cursor lifetime; 32 retained snapshots; 4,096 total continuation positions and
999 per snapshot; 8 MiB serialized JSON per snapshot and 32 MiB total; and 32 KiB
per continuation URL or authority context. A caller supplies Selector limits
explicitly; current integration tests use 16 KiB, depth 256 and 2,048 nodes.
These settings are reversible implementation bounds, not an adopted public
configuration contract. The storage inventory still refuses over 1,000 live
Resources before filtering.

These APIs require an authority-derived authorization projection and epoch.
They do not implement either one, and the writer token is not an epoch. The
real-engine tests supply explicit test identities and demonstrate paging across
a committed Link update on both backends. The [HTTP composition root](GRAPH_BDP_READ_HTTP.md) authenticates before selection,
checks authority on continuation, and uses a process generation for its one
full-workspace view. Restore requires stopping and restarting serving. These
internal modules alone deliver neither HTTP nor public History.

## Verification boundary

Run the real storage-to-wire sequence with:

```sh
BEADS_GRAPH_TEST_SERVER_PORT=PORT TEST_VERBOSE=1 ./scripts/test.sh \
  ./internal/httpapi/graphread ./internal/httpapi/bdpwire
BEADS_GRAPH_TEST_SERVER_PORT=PORT TEST_VERBOSE=1 \
  TEST_RUN='TestReadInstalledTypes|TestReadTypeRefusesInvalidInstallation' \
  ./scripts/test.sh ./internal/storage/graphstore
```

The first sequence initializes ordinary stores, writes Issue/Memory/Link data
through the existing APIs, verifies exact and owned projections after updates
and unlink, and reopens the store. No SQL payload or mock store is seeded. The
test log emits `WIRE_RECEIPT` JSON for independent protocol validation.
The public TypeScript parsers at the exact Beads wire pin
`53bdbd03136875f952af184fce7b3c7af8f74e96` accept these receipts through
`parseBeadRecord`, `parseLinkRecord` and `parseTypeDescriptor`. The extended
sequence selects Memory Beads, pages them across a real Link-property update,
replays the continuation, and emits Bead/Link/Type collection receipts for the
public `parseBeadCollection`, `parseLinkCollection` and `parseTypeInventory`
parsers. Focused Selector differential tests compare the Go implementation with
the pinned TypeScript evaluator; pagination tests also cover concurrent replay,
caller mutation, expiry, capacity refusal and view/epoch/projection fences.

Those checks establish persistence and wire format. The separate
[installed HTTP capture](GRAPH_BDP_READ_HTTP.md#reproducible-independent-client-proof)
now covers discovery, selection, authorization, HTTP behavior and a real public
client. This internal package itself advertises no profile or History capability. Private preview Memory is
still incomplete relative to the canonical Memory proposal.

Read-only adoption preflight is a separate gate before M2 breadth; this reader
does not prove migration, external-blocker compatibility or backup continuity.

## Shared-server provisioning limitation

The first Linux qualification exposed a Dolt 2.1.8 catalog race when separate
test packages initialized different databases concurrently. A transaction that
predates another connection's `CREATE DATABASE` can fail an
`INFORMATION_SCHEMA.COLUMNS` query with `could not resolve initial root for
database ...`, even when its query filters for its own database. A direct
two-connection reproduction failed three out of three times; creating the
databases before starting the transaction succeeded three out of three times.
No database deletion or graph record operation is required to reproduce it.

The qualification workflow runs test packages serially on the shared server
(`GO_TEST_PKG_PARALLEL=1`) so unrelated fixture provisioning does not overlap.
The existing explicit transaction concurrency, stale-writer, cancellation and
uncertain-commit tests remain enabled. This is test isolation, not an engine
fix or proof that concurrent provisioning of different databases is safe.
Deployments on this release must serialize such provisioning; lifting that
limitation needs a qualified engine release. No application lock, retry or
storage-engine workaround is introduced here.
