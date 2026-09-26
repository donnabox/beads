# BDP current-record projection — preparation for serving

This increment follows qualified [PR #23](https://github.com/donnabox/beads/pull/23)
at `9129dfd08690099a0fe2b8d8182322b6912d1587` and the
[fork delivery plan](https://github.com/donnabox/beads/pull/18).
It adds a storage-to-wire reader, not a new CLI command or an HTTP service.
`bd serve` remains unavailable for graph workspaces. The September 23, 2026
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
corruption also remain distinct internal errors. The eventual HTTP adapter must
apply the public problem table: ordinary reads of deleted Resources use
`resource-not-found`; they must not publish the private CLI `gone` envelope or
pretend deletion is erasure.

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
`parseBeadRecord`, `parseLinkRecord` and `parseTypeDescriptor`.

Those are persistence and wire-format checks, not an installed CLI/HTTP demo.
The minimum Read profile still requires Scope discovery, Resource and property
views, installed Type inventory, filtered collections with bounded Selectors,
incident views, pagination and continuation semantics, HTTP method/media and
conditional behavior, and a real independent-client round trip. No Read profile
or History capability is advertised by this package. Private preview Memory is
still incomplete relative to the canonical Memory proposal.

Read-only adoption preflight is a separate gate before M2 breadth; this reader
does not prove migration, external-blocker compatibility or backup continuity.
