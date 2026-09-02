# BDP graph store — architecture and design

**Status:** Draft v1 (W-arch) — feat/bead-graph
**Date:** 2026-09-02
**Companion:** `BDP_BEAD_GRAPH_PLAN.md` (the plan and its twelve rulings) and
`BDP_GRAPH_CLI_AND_STORAGE_SPEC.md` (the detailed CLI and storage-interface
changes). This document is the *shape*: what the pieces are, where they live,
how a request flows, and — importantly — where this design corrects the plan
after the tree's own conventions were read closely.

## 1. The one-paragraph version

The graph store is a new **plane** beside issues and memories: a public leaf
contract package (`graphops`) declaring the value types and role interfaces,
reached through **role accessors on `storage.Storage`** exactly as
`issueops` and `memoryops` are — promoted through every decorator by
interface embedding, wrapped by telemetry, recursed unwrapped by the hook
layer — with one shared transaction-level body under `internal/storage`
that both Dolt stores call and a unit-of-work leg beside it, proven by a
`backend/conformance` role contract wired on all three legs and guarded by
the existing coverage gates. `bd bdp-serve` takes those roles from beneath
the hook layer and serves BDP over them; CLI graph verbs reach the same
roles through the same accessors, or — when a workspace is `bd init
--bdp-server`'d — speak BDP to a designated server instead. Every ruling in
the plan survives; three mechanisms in the plan's §4 are replaced by the
house idiom that already solves their problem.

## 2. Corrections to the plan (read this first)

The plan's storage section was written before the repo's role machinery was
read in depth. Three of its mechanisms are superseded — not because they
were wrong in the abstract, but because the tree already has a stronger,
enforced answer to each:

| Plan §4 mechanism | Replaced by | Why |
| --- | --- | --- |
| `GraphCapable` as a *separate optional interface*, resolved by targeted decorator peels (`graphsource` package) | **Role accessors on `storage.Storage`** (`GraphReader()`, `GraphTypes()`), the house rule: "a new capability gets a new role interface and a new accessor" | Decorators embed the `DoltStorage` interface, so an accessor added to it is **promoted through every decorator automatically**; the reflection census in `role_accessor_decorator_test.go` puts it under test the moment it joins the interface. A separate interface is exactly the shape the census warns about — an accessor "in the same shape" that "every role simply stops being asked about." The `graphsource` package existed only to work around non-promotion; it is unnecessary. The cost is the documented one: a required accessor is a **breaking change for out-of-tree backends**, called out in `CHANGELOG.md` — which is also how `Memories()` landed. |
| The **snapshot lease** (`OpenSnapshot` → request-scoped transaction object) | **One role call = one transaction**, the house transaction model ("a request and a transaction are the same span"); the BDP handler makes **one role call per HTTP request** | Reads that must see one snapshot are written as `…InTx` bodies under `internal/storage/<plane>` with each store's accessor wrapping `withReadTx`/`withConn` — the `Relations`/`EdgeReader`/`CycleDetector` pattern. Assembling a Bead record with its complete `ownedLinks` in one call is exactly that shape. No transaction object escapes a role. Cross-*request* cursor continuation remains P2's ADR (registry / materialized sets / `AS OF`), unchanged. |
| A `ResolveGraphReadSource` policy that "re-applies telemetry" after peeling | Nothing — **the accessors carry the layers**: telemetry wraps in its own accessor (`internal/telemetry/graph_reader.go`), the hook layer declares the accessor and recurses unwrapped (a read role), and `bd bdp-serve` takes roles from beneath the hook layer with the same one-line peel `bd serve` already does | This is what accessors are *for* in this tree ("the accessor is where each storage decorator adds its layer"); a resolver that peels and re-wraps duplicates that machinery and is the shape the `cmd-bd-role-constructors` depguard rule exists to forbid. |

What the corrections do **not** change: ruling 9's level — the authority is
the graph store as reached through the normalized storage abstraction, on
any provider; the graph contract is **public in `backend/` from P1**
(aliases, completeness-guard roots, conformance family); Dolt is the
reference realization. The house idiom is the *mechanism* by which that
abstraction is realized; ruling 9 is the *policy* it serves.

## 3. Packages and their imports

```text
graphops/                        PUBLIC LEAF CONTRACT (sibling of issueops/, memoryops/)
  ├─ types.go                    Bead, Link, Ref (in-Scope|external sum), Properties,
  │                              TypeDescriptor, Revision, Attribution — value types
  ├─ reader.go                   type Reader interface { Bead, Link, Beads, Links,
  │                              IncidentLinks } + request/result types
  ├─ types_role.go               type Types interface { Descriptors } (+ install, P3)
  └─ errors.go                   sentinels ALIASED from beadserrors (errors.Is across
                                 the module boundary)
  imports: stdlib + beadserrors ONLY — no internal/types (nothing here is
  issue-shaped; the memoryops precedent, stated the same way)

internal/graphapi/               MEANING FUNCTIONS (pure; importable by cmd/bd)
  canonical-ID grammar (reject-don't-trim), canonical-URI ordering, JSON
  object canonicalization + RFC 6902 §4.6 equality (the no-op gate),
  Scope URL validation mirroring BDP's startup contract. No bodies, no
  constructors — the memoryapi shape.

internal/storage/graphops/       TX-LEVEL SHARED BODY (both Dolt stores call it)
  ReadBeadInTx / ReadLinkInTx / SelectBeadsInTx / SelectLinksInTx /
  IncidentLinksInTx / DescriptorsInTx — every function takes *sql.Tx; no
  exported constructor, nothing for depguard to deny. ownedLinks is
  assembled HERE, inside the caller's one transaction.
  Inside the charter's Storage Boundary (only internal/storage/** may
  touch the engine — .golangci.yml dolt-storage-boundary).

internal/storage/dolt/graph_reader.go, internal/storage/embeddeddolt/graph_reader.go
  five-line accessors wrapping the body in withReadTx / withConn (embedded
  carries //go:build cgo); nil receiver → *storage.ErrUnsupported.

internal/storage/uow/graph_reader.go
  GraphReaderSource { GraphReader() (graphops.Reader, error) } on the
  provider; the UOW leg reaches the same InTx body through the domain
  repository IF every function it calls takes an interface Runner
  satisfies — decided per method at implementation time and STATED in the
  contract header (vote count: two legs share the body → "two readings
  plus an engine check", never "three backends agree").

internal/storage/storage.go        + GraphReader(), GraphTypes() accessors (one line
                                     each, with the spec-grade doc the file uses)
internal/storage/hook_graph_reader.go   declares the accessor, recurses UNWRAPPED (read role)
internal/telemetry/graph_reader.go      wraps every method with storage.op / storage.done
backend/types.go                        public aliases for every graphops type the
                                        accessors reach (completeness guard root)
backend/conformance/graph_reader_contract.go   the role contract; RoleContractBundle
                                        gains GraphReader/GraphTypes fields; wirings in
                                        internal/storage/{dolt,embeddeddolt,uow}

internal/bdpapi/                 THE BDP HTTP SURFACE (sibling of internal/httpapi)
  generated wire DTOs from the pinned bdp-v0 schema; handler = serializer
  over graphops roles; typed graph errors → BDP Problem records mapped
  here and only here; reuses httpapi's exported pieces (TokenFileAuth,
  ValidateBindAddr, ValidateAuthPosture, ValidateAllowedHost) for identical
  bearer/bind/host posture.

cmd/bd/bdp_serve.go              `bd bdp-serve`: resolves the workspace, takes roles
                                 from BENEATH the hook layer (serve's one-line peel),
                                 mints the Scope on first serve under bdp.scope_url,
                                 listens via internal/bdpapi
cmd/bd/graph_*.go                graph verbs (P3): reach roles via openGraphReader()
                                 — the direct/proxied/BDP-client fork lives there
```

Dependency direction, enforced: `cmd/bd → graphops, graphapi, storage
accessors`; `internal/bdpapi → graphops, graphapi`; `internal/storage/* →
graphops`; `graphops → beadserrors, stdlib`. `.golangci.yml`'s
`cmd-bd-role-constructors` deny list gains no entry for the tx-level body
(no constructor exists to deny), exactly as for `memoryops`.

## 4. Roles — how many, and why

The house test: a role is a **different question**, born whole with the
methods that are shapes of one question, never appended to later.

- **`graphops.Reader`** — "what is in this Scope, as BDP sees it": one
  record by canonical ID (Bead with its complete `ownedLinks`, or Link),
  a collection selection (canonical-URI order, bounded page), and
  incident Links for a Bead. These are shapes of one question (Read
  projection of the graph) the way `Reader.Ready/List/Get` are; they share
  the snapshot-per-call model and the same refusal vocabulary.
- **`graphops.Types`** — "what Types does this Scope advertise": the
  descriptor inventory, and (P3) installation with closure validation and
  fingerprint retention. A different question with a different lifetime
  (descriptors change by operator action, not by writes), and P3's
  install is a write the Reader must not be able to reach — the
  `Bootstrapper`/`InitVerifier` argument: *can one caller be entitled to
  the read and not the write?* Yes.
- **`graphops.Writer`** (P3, not now) — create/update/delete for graph
  beads and links with the owned-Link version coupling and the
  allocation/tombstone ledger inside the write transaction. Born whole in
  P3 with the write-profile ADR; not declared before its semantics are.

Deliberately *not* a role: the authority marker and Scope identity. They
are workspace identity written once (`bd bdp-serve`'s first-serve mint)
and read on every serve — the `Bootstrapper`/`InitVerifier` split applies:
`graphops.Types` does not carry them; a small `graphops.Identity` pair
(read role + one-time write role) is the honest shape, decided in the CLI
and storage spec.

## 5. A read, end to end

```text
client ──GET /acme/beads/x──▶ bd bdp-serve (internal/bdpapi)
   │ bearer + Host + project stamp + deadline + semaphore   (httpapi's posture, reused)
   │ ScopeResolver: workspace → Scope URL + authority marker check (ruling 9/12)
   ▼
graphops.Reader.Bead(ctx, {ID})            ← ONE role call per request
   ▼ (telemetry span) → (hook layer absent: taken from beneath it)
dolt.GraphReader → withReadTx ─▶ storage/graphops.ReadBeadInTx(tx, id)
   │ one transaction: bead row + its owned Links + descriptor lookup
   ▼
graphops.BeadRecord {Bead, OwnedLinks}     ← the covered member, assembled in-snapshot
   ▼
internal/bdpapi: generated DTO ← record; typed error → BDP Problem; JSON out
```

A write (P3) follows the same path through `graphops.Writer`, whose body
runs the no-op gate (`graphapi` equality) before minting, records
attribution per version, versions the source on owned-Link mutation, and
consults the allocation/tombstone ledger — all inside the one write
transaction, on whichever provider realizes the accessor.

## 6. Where the twelve rulings land

| Ruling | Lands in |
| --- | --- |
| 1 charter (core; amendment after working bits) | `PROJECT_CHARTER.md` edit rides the first merged slice; nothing here depends on it |
| 2 substrate S1 | `internal/storage/schema/migrations/NNNN_graph_*.up.sql` — beads, links, type descriptors, allocation ledger, scope identity |
| 3 allocation ledger (keyed O(1)/O(log n)) | `internal/storage/graphops/ledger.go` (InTx), keyed by canonical URL |
| 5 withdrawal | nothing projects Issues; `graphops` imports no `internal/types` — the boundary is structural |
| 7a Scope URL | `internal/graphapi/scopeurl.go` (validator mirroring BDP's startup contract); `bdp.scope_url` config; identity row |
| 7b listener | `internal/bdpapi` reuses httpapi's exported auth/bind/host posture; v0 view = whole Scope per bearer token |
| 8 changefeed | P3: a `graphops.Changefeed` role over the graph's own log; the frozen v0 journal untouched |
| 9 authority | the accessors ARE the normalized abstraction; marker/serialization/refusal are contract cases in `backend/conformance` |
| 11 restore vs identity | ledger table restorable independently (append-only, its own migration); `bd graph restore` rotation path in the CLI spec |
| 12 store/Scope/client | `bd init` migrations; `bd bdp-serve` mint-on-first-serve; `bd init --bdp-server` reroute (CLI spec §) |
| 6 wisps | not served; C-lane visibility decision recorded in the plan |

## 7. What is deliberately not designed here

- Cross-request cursor continuation (P2 ADR).
- The write profiles' wire (W1 upstream), and therefore `graphops.Writer`'s
  exact request shapes.
- `bd serve` integration (W2): this document gives `bd bdp-serve` its own
  command and its own `internal/bdpapi` server; folding BDP routes into
  `httpapi`'s table is W2's design, gated on the OpenAPI-first rule that
  table lives under (BDP's wire is defined by the pinned JSON Schema, not
  by `openapi.v0.yaml`, so the two surfaces cannot share a generator
  without a ruling).
- Type generation from the bead-type inventory (W3) — it feeds
  `graphops.Types` bootstrap; it does not change the role.
