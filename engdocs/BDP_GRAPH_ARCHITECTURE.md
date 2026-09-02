# BDP graph store — architecture and design

**Status:** Draft v2 (W-arch, post-council) — feat/bead-graph
**Date:** 2026-09-02
**Companion:** `BDP_BEAD_GRAPH_PLAN.md` (the plan and its twelve rulings) and
`BDP_GRAPH_CLI_AND_STORAGE_SPEC.md` (the detailed CLI and storage-interface
changes). This document is the *shape*: what the pieces are, where they live,
how a request flows, and where this design corrects the plan after the
tree's own conventions were read closely. v2 answers a three-reviewer
council (Claude, Codex, Gemini) on v1; §2 records what changed and which
changes amend a ruling rather than a mechanism.

## 1. The one-paragraph version

The graph store is a new **plane** beside issues and memories: a public leaf
contract package (`graphops`) declaring the value types, the laws, and six
role interfaces, reached through **role accessors on `storage.Storage`**
exactly as `issueops` and `memoryops` are — declared explicitly by every
decorator (promotion is the failure mode the censuses catch), wrapped by
telemetry, recursed unwrapped by the hook layer — with one shared
transaction-level body under `internal/storage` taking a `DBTX`-shaped
runner so that both Dolt stores *and* the unit-of-work leg call the same
code, proven by `backend/conformance` role contracts wired on all three legs
and guarded by the existing coverage gates. BDP is served by the existing
`internal/httpapi` server as a **conditional second route table** behind the
same `route()` middleware (deadline, bearer, project stamp, semaphore);
`bd bdp serve` is a thin command over that server that *requires* a Scope,
and `bd serve` mounts the same rows when a Scope URL is configured. Every
read is **one role call, one transaction**, and that transaction asserts the
authority expectation the server was started with — the Scope row that
replicates plus the clone-local authority half that does not — so a clone,
a restore, or a promotion elsewhere is refused by the store, not by a
startup check. The CLI's graph verbs live under `bd bdp …` and reach the
same roles through the same accessors, or — when a workspace is `bd init
--bdp-server`'d — speak BDP to a designated server instead.

## 2. Corrections to the plan, and proposed ruling amendments (read this first)

Two kinds of change are recorded here and must not be confused. A
**mechanism correction** replaces something the plan's §4 *proposed* with
the house idiom that already solves the problem; the ruling it serves is
unchanged. A **ruling amendment** changes text the operator ratified in §9;
it is *proposed* here and takes effect only when ruled. v1 blurred the two
(the council's first finding, unanimously); v2 keeps them apart.

### 2a. Mechanism corrections (no ruling changes)

| Plan §4 mechanism | Replaced by | Why |
| --- | --- | --- |
| `GraphCapable` as a *separate optional interface*, resolved by targeted decorator peels (`graphsource` package) | **Role accessors on `storage.Storage`** — the house rule: "a new capability gets a new role interface and a new accessor" | Every accessor lives on `Storage` (28 today); `DoltStorage` embeds it; both decorators embed `DoltStorage`. An accessor added to `Storage` therefore *compiles* through every decorator by promotion — and promotion is exactly what the reflection census in `role_accessor_decorator_test.go` (and its telemetry twin) rejects: each decorator must **declare** the accessor and say what it does. A separate interface is the shape the census cannot see. The cost is the documented one: a required accessor is a **breaking change for out-of-tree backends**. `backend/backend.go`'s stability note promises a CHANGELOG call-out for such additions; no previous accessor actually wrote one (v1 cited a `Memories()` entry that does not exist), so this slice writes the first. |
| A `ResolveGraphReadSource` policy that "re-applies telemetry" after peeling | Nothing — **the accessors carry the layers**: telemetry wraps in its own accessor file, the hook layer declares the accessor and recurses unwrapped, and the server takes roles from beneath the hook layer with the one-line peel `bd serve` already does | This is what accessors are *for* in this tree; a resolver that peels and re-wraps duplicates that machinery. |
| `backend/types.go` aliases for every `graphops` type, enforced by `TestPublicSurfaceComplete` | **No aliases.** `graphops` is a public root package, like `issueops`; external backends import it directly | The completeness guard demands aliases only for types under `internal/` (`backend/completeness_test.go`); public `issueops` request/result types are deliberately *not* re-exported. Ruling 9's "public in `backend/` from P1" is satisfied by the package being public and by the conformance family living in `backend/conformance`. |
| `internal/graphapi` as a separate meaning-function package | **The laws live in `graphops`** as pure functions beside the values they govern (plan §3: "laws in the package, tested once") | A `graphops` constructor that called `internal/graphapi` would break the leaf's import rule (stdlib + `beadserrors` only) and, if `graphapi` validated `graphops` values, cycle. `cmd/bd` and `httpapi` import the public leaf for the grammar and canonicalization. |

### 2b. Proposed ruling amendments (pending operator ruling)

| # | Ruling | v1 said | v2 proposes | Why |
| --- | --- | --- | --- | --- |
| A1 | **9** (obligations list names "the snapshot lease") | dropped the lease as a mechanism swap | **The per-call transaction is the v0 lease**, and it carries the authority check: every read role call takes an `AuthorityExpectation` that the body asserts *inside the same transaction* as the read. Cross-request continuation is P2's cursor ADR and may reintroduce a request-scoped snapshot there. Ruling 9's obligation list reads "single-transaction reads under an asserted authority expectation" in place of "the snapshot lease". | A ruled obligation cannot be swapped silently; and a startup-only authority check split from the read serves a superseded authority the moment a promotion lands elsewhere (all three reviewers). |
| A2 | **7b + 12** (listener; `bd bdp-serve` isolated bootstrap) | a sibling server (`internal/bdpapi`) reusing four exported validators | **BDP rows mount inside `internal/httpapi`** as a second, conditional route table behind the same `route()` wrapper; **`bd bdp serve`** is a thin command over the same server that refuses to start without a Scope; `bd serve` mounts the rows when a Scope URL is configured (ruling 12's literal text). Isolation is preserved as *a conditional table plus a config field*: a `bd serve` with no Scope configured is byte-identical. | The posture that matters — rebinding defense, bearer-before-semaphore ordering, project stamp behind the auth gate, deadline, never-log-the-token — is **private** to `httpapi`; a sibling reimplements exactly the pieces easiest to get wrong (Codex Critical, Claude High). |
| A3 | **12** + plan §4 lifecycle (`bd bead`, `bd link`, `bd graph …`) | verbs "reserved now" | **Everything under `bd bdp …`**: `bd bdp bead`, `bd bdp link`, `bd bdp types`, `bd bdp status`, `bd bdp promote`, `bd bdp restore`, `bd bdp ledger`, `bd bdp serve`. | `bd link <id1> <id2>`, `bd graph [issue-id]` (with `graph check`), `bd restore <issue-id>`, `bd promote <wisp-id>` all exist; the plan's constraint #1 forbids changing them. |
| A4 | plan §3 layering (values in `graph`, roles in `graphops`) | moved values into `graphops` without saying so | **Record it:** values, laws, and roles all live in public `graphops`; `internal/graphapi` is not created. | Package layout is public API (Codex M21). |
| A5 | **11** (ledger "restorable independently … its own migration") | its own migration | **A migration is a table definition, not durability.** Ruling 11's mechanism is: (i) the clone-local authority half lives in a dolt-ignored table, so a restore or a clone *arrives without authority* and must be re-granted explicitly; (ii) a **ledger export/import lane** (`bd bdp ledger export|import`, sequence-watermarked) is how continuity is carried across a restore; (iii) the provider **declares** `LedgerDurability` on the identity reader, and `bd bdp restore` rotates unless continuity is shown. | Codex Critical 3, Claude 2.5/5.3, Gemini L7. |
| A6 | plan §4 lifecycle (`metadata.json` `bdp_client`/`bdp_server`; "precedence identical to `dolt.mode`") | both metadata.json and config.yaml | **config.yaml is the single local source** for `bdp.*` client routing (yaml-only, validated); nothing is written to metadata.json; durable Scope identity lives only in the graph store. | metadata.json outranking a later `bd config set bdp.server` forever (Claude 1.3, Codex M16); the tree's `dolt.mode` resolution is not the four-step chain v1 claimed. |

Two further council findings are **decisions the plan does not yet contain**
and are surfaced for ruling rather than designed around:

- **Enforcement boundary for out-of-role writes** (Codex Critical 4). `bd
  sql`, the proxied `RawSQLUseCase`, and a Dolt merge can change graph tables
  without allocation, authority, revision, or owned-Link coupling checks.
  v2's position (§7): graph-table DML outside the roles is **out of
  contract**; a **post-merge validator** rejects and rolls back graph deltas
  that fail the invariants or arrive from a foreign authority; DB-privilege
  or trigger enforcement is a C-lane verification task, not v0. To be ruled
  before P3.
- **Replication/merge ADR** (Codex High 13, Claude 5.5). Graph settlement is
  not an "always-run pass that changes nothing": `MergeAndSettle`,
  `MergeAndSettleWithStrategy`, `MergeWithStrategy`, and plain `Merge` are
  the entry points, and a single-authority history wants *refusal* of
  foreign-authority deltas rather than invented merge rules. The ADR lands
  before the graph migrations do (P1 gate).

What none of this changes: ruling 9's level — the authority is the graph
store as reached through the normalized storage abstraction, on any
provider; Dolt is the reference realization; the CLI verbs and the BDP
handler are both clients of that abstraction.

## 3. Packages and their imports

```text
graphops/                        PUBLIC LEAF (sibling of issueops/, memoryops/)
  ├─ types.go                    Bead, Link, Ref (in-Scope|external), Properties,
  │                              Revision, Attribution, TypeDescriptor, OwnedLinkDecl,
  │                              ScopeIdentity, AuthorityExpectation — value types with
  │                              unexported fields and law-enforcing constructors
  ├─ laws.go                     pure functions: canonical-ID grammar (reject, never
  │                              trim), code-unit ordering, JSON object canonicalization,
  │                              RFC 6902 §4.6 equality (the no-op gate), Scope-URL
  │                              validation mirroring BDP's startup contract
  ├─ reader.go                   Reader: Bead, Link, Beads, Links, IncidentLinks
  ├─ types_role.go               DescriptorReader: Descriptors, Descriptor
  │                              TypeInstaller: Install (idempotent, fingerprint-keyed)
  ├─ identity.go                 IdentityReader: Read, LedgerDurability
  │                              ScopeBootstrapper: Mint (once)
  │                              AuthorityRotator: Promote, Rotate
  ├─ writer.go                   (P3) Writer — born whole with the write-profile ADR
  └─ errors.go                   sentinels ALIASED from beadserrors (errors.Is across
                                 the module boundary): ErrNotFound, ErrValidation,
                                 ErrGone{Path, State}, ErrNoScope, ErrScopeExists,
                                 ErrNotAuthority, ErrURLReused, ErrRepresentationTooLarge
  imports: stdlib + beadserrors ONLY — no internal/types (nothing here is
  issue-shaped; the memoryops precedent, stated the same way)

internal/storage/graphops/       TX-LEVEL SHARED BODY — all three legs call it
  type DBTX interface { ExecContext; QueryContext; QueryRowContext }
    (the issueops.DBTX shape; *sql.Tx and domain/db.Runner both satisfy it)
  ReadBeadInTx / ReadLinkInTx / SelectBeadsInTx / SelectLinksInTx /
  IncidentLinksInTx / DescriptorsInTx / DescriptorInTx / InstallDescriptorInTx /
  ReadIdentityInTx / MintScopeInTx / PromoteInTx / RotateInTx /
  assertAuthorityInTx (called first by every reader body)
  SeedBeadInTx / SeedLinkInTx — the P1 fixture writer, ledger-enforcing,
    reachable ONLY through the conformance fixture hook (no accessor)
  No exported constructor. Inside the charter's Storage Boundary; cmd/bd
  is DENIED this package by an explicit depguard entry (Part B6).

internal/storage/dolt/graph_*.go, internal/storage/embeddeddolt/graph_*.go
  accessors wrapping the bodies in withReadTx / withRetryTx (server) and
  withConn (embedded, //go:build cgo); nil receiver → *storage.ErrUnsupported

internal/storage/domain/graph.go + internal/storage/domain/db/graph_repository.go
  GraphRepository over a db.Runner, delegating to the InTx bodies — the
  MetadataCAS/TreeWalker precedent for reaching a shared body from the
  unit-of-work leg
internal/storage/uow/graph_*.go
  uow.UnitOfWork gains Graph() domain.GraphRepository; the provider gains
  GraphReader()/GraphTypes()/GraphIdentityReader()/GraphScopeBootstrapper()
  accessors (RunTxRead for reads; RunTxResult with a commit message for
  mint/promote — a no-op result commits nothing). The notifying provider
  declares each explicitly.

internal/storage/storage.go        + six accessors (one line each, spec-grade doc)
internal/storage/hook_graph_*.go   declared; recurse UNWRAPPED (no graph hook vocabulary)
internal/telemetry/graph_*.go      every method spanned storage.op / storage.done
backend/conformance/graph_*_contract.go   role contracts; RoleContractBundle fields;
                                   role_bundle_cases.go rows; wirings on all three legs

internal/httpapi/bdp_routes.go     bdpRouteTable — conditional rows behind route()
internal/httpapi/bdp_handlers.go   handler = serializer over graphops roles
internal/httpapi/bdp_problem.go    typed graph errors → BDP Problem records, here only
internal/httpapi/bdpwire/          GENERATED DTOs from the vendored, pinned bdp-v0 schema
                                   (+ schema/ with provenance: upstream commit, sha256)
internal/bdpclient/                graphops.Reader/DescriptorReader over the wire
                                   (Problem → the same typed errors; errors.Is holds)

cmd/bd/bdp.go                      `bd bdp` root; subcommands in cmd/bd/bdp_*.go
cmd/bd/bdp_serve.go                thin: serveDatabaseSource + serveIssueRoles +
                                   graph roles + httpapi.Config{Graph: …}
```

Dependency direction, enforced: `cmd/bd → graphops, storage accessors,
bdpclient`; `internal/httpapi → graphops, bdpwire`; `internal/storage/* →
graphops`; `graphops → beadserrors, stdlib`. `.golangci.yml` gains an
explicit `cmd-bd-role-constructors` deny entry for
`internal/storage/graphops` — the rule matches by package import, not by
constructor symbol, so "no constructor to deny" (v1) was not a guard.

## 4. Roles — how many, and why

The house test: a role is a **different question**, born whole with the
methods that are shapes of one question; and *can one caller be entitled
to the read and not the write?* — if yes, two roles. v1 stated the test and
then put `Read`/`Mint`/`Rotate` on one interface (all three reviewers).
v2's six, each behind its own accessor:

- **`graphops.Reader`** — "what is in this Scope, as BDP sees it": one
  record by path (a Bead with its complete, bounded `ownedLinks`, or a
  Link), a keyset-paged selection in code-unit order, incident Links. Every
  method takes the `AuthorityExpectation` and asserts it in-transaction.
- **`graphops.DescriptorReader`** — "what Types does this Scope
  advertise": the ordered catalog and a keyed lookup. Bounded (the
  installer refuses a catalog past the bound).
- **`graphops.TypeInstaller`** — install/converge descriptors, keyed by
  fingerprint, with closure validation. **P1**, not P3: `bd init`'s
  descriptor bootstrap needs a writer (v1 had none), and the conformance
  fixture needs the same one. Refuses an *owning* declaration without a
  `Max`.
- **`graphops.IdentityReader`** — the Scope row, the clone-local authority
  half, and the provider's `LedgerDurability` declaration. What every serve
  and every `bd bdp status` consults.
- **`graphops.ScopeBootstrapper`** — `Mint`, once: INSERT into the
  singleton Scope row (loses the race, `ErrScopeExists`) and grant this
  clone the authority half. The only write the server assembly may hold,
  and only on the first-serve path.
- **`graphops.AuthorityRotator`** — `Promote` (CAS on the epoch; writes
  this clone's authority half) and `Rotate` (new URL; old URL recorded
  refused-forever). Reached only by `bd bdp promote|restore` — an offline
  administrative composition root; `httpapi.GraphConfig` has no field for
  it, so the server *cannot* hold it (compile-time, tested).
- **`graphops.Writer`** (P3, not now) — born whole with the write-profile
  ADR (W1 upstream).

The fixture writer (`SeedBeadInTx`/`SeedLinkInTx`) is deliberately not a
role: it is reachable only through the conformance fixture hook (the
`MemoriesFixture.SetConfig` shape), so nothing outside a contract test can
write beads before P3.

## 5. A read, end to end

```text
client ──GET /acme/beads/x──▶ internal/httpapi (bd serve | bd bdp serve)
   │ route(): deadline → bearer (before the semaphore) → Bd-Project-Id stamp
   │          (absent = pass; BDP clients never send it) → database slot
   ▼
bdp handler: path grammar (graphops laws) → ONE role call
   ▼
graphops.Reader.Bead(ctx, BeadRequest{Path, Expect})     Expect = the identity the
   ▼ telemetry span (hook layer absent: taken from beneath it)   server started with
dolt.GraphReader → withReadTx ─▶ storage/graphops.ReadBeadInTx(ctx, tx, req)
   │ assertAuthorityInTx: graph_scope(url, epoch) == Expect AND
   │                      graph_authority_local(authority_id, epoch) == Expect
   │                      → else ErrNotAuthority (a promote elsewhere + pull, a
   │                        restore, or a clone all fail HERE, mid-process)
   │ bead row + ONE batched owned-links query + descriptor lookup (fingerprint-
   │ cached; re-read in-tx only when the catalog fingerprint changed)
   ▼
graphops.BeadRecord {Bead, OwnedLinks}     ← complete, ordered, bounded (page × ΣMax;
   ▼                                          over the bound → ErrRepresentationTooLarge)
bdp handler: bdpwire DTO ← record; typed error → BDP Problem (bdp_problem.go); JSON out
```

The server obtains `Expect` once at startup through `IdentityReader` (and
mints through `ScopeBootstrapper` on the first-serve path); it never caches
the *answer* — every read re-asserts it in the store. That is the v0
"lease" (amendment A1): one transaction per role call, authority and data
from the same snapshot, nothing escaping a role.

On Dolt-server workspaces `bd serve` answers from the **unit-of-work
provider**, so the UOW leg is the *primary* production path for BDP, not a
third vote; the `DBTX`-shaped bodies are what make it the same code.

A write (P3) follows the same path through `graphops.Writer`, whose body
asserts the expectation, runs the no-op gate, records attribution per
version, versions the source on owned-Link mutation, and consults the
ledger — all inside the one write transaction, on whichever provider
realizes the accessor.

## 6. Where the twelve rulings land

| Ruling | Lands in |
| --- | --- |
| 1 charter (core; amendment after working bits) | `PROJECT_CHARTER.md` edit rides the first merged slice; nothing here depends on it |
| 2 substrate S1 | `internal/storage/schema/migrations/NNNN_graph_*.up.sql` (replicated: scope, scope history, beads, links, descriptors, allocations) + `migrations/ignored/NNNN_graph_authority_local.up.sql` (clone-local) |
| 3 allocation ledger (keyed O(1)/O(log n)) | `graph_allocations` PK on the Scope-relative path; `ledger.go` InTx |
| 5 withdrawal | nothing projects Issues; `graphops` imports no `internal/types` — structural |
| 7a Scope URL | `graphops` Scope-URL law; `bdp.scope_url` (yaml-only); singleton Scope row; dev mode never persists |
| 7b listener | BDP rows behind `httpapi.route()` — same posture by construction again (A2) |
| 8 changefeed | P3: a `graphops.Changefeed` role over the graph's own log; the frozen v0 journal untouched |
| 9 authority | accessors = the normalized abstraction; replicated Scope row + clone-local authority half + in-transaction expectation (A1); promotion CAS; contract cases incl. a push/pull-produced clone refusing |
| 11 restore vs identity | dolt-ignored authority half (arrives absent after restore/clone); ledger export/import lane; `LedgerDurability` declaration; `bd bdp restore` rotates unless continuity shown (A5) |
| 12 store/Scope/client | `bd init` migrations + descriptor bootstrap via `TypeInstaller`; mint on first serve (`bd serve` when configured / `bd bdp serve` always); `bd init --bdp-server` → config.yaml (A6) |
| 6 wisps | not served; C-lane visibility decision recorded in the plan |

## 7. What is deliberately not designed here

- Cross-request cursor continuation (P2 ADR) — and therefore **no BDP
  collection routes ship before it lands**; v0 P1 serves single-resource
  reads and discovery only.
- The write profiles' wire (W1 upstream), and therefore `graphops.Writer`'s
  exact request shapes; per-token authorization classes (read / write /
  admin) are a W1 dependency and precede P3.
- The replication/merge ADR (foreign-authority delta refusal; funnel through
  every merge entry point; federation policy for graph tables — v0 default:
  replicated tables travel unfiltered, by decision). Precedes the migrations.
- The enforcement boundary for out-of-role DML (`bd sql`, raw SQL, merge)
  beyond the post-merge validator — a ruling, then a C-lane verification task.
- Whether `bd bdp serve` survives W2 as an alias of `bd serve` once every
  workspace with a Scope URL serves BDP from `bd serve` (default: it does,
  as the strict form).
- Type generation from the bead-type inventory (W3) — it feeds
  `TypeInstaller`'s bootstrap catalog; it does not change the role.
