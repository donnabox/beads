# BDP in beads: the bead-graph plan

**Status:** Draft v8 — feat/bead-graph (nine adversarial review rounds:
1–7 on the whole plan, verdict SOUND at round 7; 8–9 focused on the
storage-interfaces addition; v6 withdrew the Issue projection from v0 on
review-round-5 counterexamples)
**Date:** 2026-09-02 (v1: 2026-08-31)
**Owners:** Donna Box (ruling), janet (drafting/implementation)
**References:** the BDP spec (gastownhall/bdp `docs/specs/bdp.md`), beads#6051,
this repo's `backend/` conformance surface, `engdocs/PROJECT_CHARTER.md`.

## 0. The BDP pin, and the spec-first dependency

This plan targets the BDP spec **as of the owned-Links rulings**:
**BDP commit `aee075f5`** (the gastownhall/bdp PR #17 merge, 2026-09-01),
schema bundle `schemas/bdp-v0.schema.json` at that commit, Read conformance
matrix `packages/conformance/matrices/read-v1.json` at that commit
(38 scenarios).
**No implementation phase begins until the pin is written here.** BDP remains
a draft; "matrix green" exits below mean green against the *pinned* matrix,
re-pinned deliberately, never against a moving `main`.

Model laws this plan builds to (all normative at the pin):

- A Link is first-class; its `id`, `type`, `source`, `target`, and pin are
  immutable; repoint/re-pin is delete-and-create.
- **Owned Links:** a Bead Type may own outgoing Link Types; each owned Link —
  target, pin, and properties — is part of the source Bead's versioned state;
  every mutation of an owned Link (create, delete, property update) versions
  the source. The record read serves an `ownedLinks` member: complete Link
  records, keyed by owned Link Type URL, ascending code-unit id order.
  Unowned incident Links are a view and version nothing.
- Revisions are opaque and equality-only. **Each surviving Resource whose
  state actually changes receives a fresh opaque revision; a semantic no-op
  (RFC 6902 §4.6 value comparison) changes none; and A→B→A is three
  distinct revisions — a reverse transition never reuses one.** Deletion
  mints nothing for the deleted Resource — its result reports the deleted
  identity (the final live revision is Transactional `DeletedData` Event
  material, not a deletion-result member); an owned-Link deletion result
  additionally reports the owning source's fresh revision.
- References: the URI is identity; a pin is provenance, echoed byte-identical,
  equality-only, never validated in v0. In-Scope and external references have
  different canonicalization duties — the Go model captures them as a sum,
  not a naked pair.
- Authorization views are closed projections, closed over owned Links.
- The Read problem table is closed vocabulary (including `resource-pruned`
  and `resource-erased` — merged in gastownhall/bdp#16).

## 1. Goal and constraints

Implement BDP natively in this repo: a first-class bead/link graph the
protocol (and eventually the CLI) is implemented in terms of, beside — not
instead of — the existing Issue/Dependency machinery.

Hard constraints, in priority order:

1. **Zero compatibility degradation, defined precisely:** *same-version*
   legacy behavior is byte-identical — every existing CLI verb, JSONL
   interchange shape, journal record, sync path, and out-of-tree `backend/`
   implementation behaves exactly as before on the same binary. Schema
   migrations keep their existing version discipline (an upgraded database
   is "ahead" of an older binary, which refuses to open it — that is the
   *current* contract, and this plan does not promise more than the repo
   already does). Mixed-version rollout beyond that is out of scope unless
   ruled otherwise.
2. **BDP fidelity** at the pin, provable by the pinned matrix under its
   packaged/public-boundary and self-certified/in-process provenance
   split.
3. **One seam per axis.** Two different seams exist and must not be
   conflated: an outer **authority seam** (which workspace/store, which
   authorization view, which read snapshot — `ScopeResolver`) selects
   exactly one scope; the **representation seam** (native vs projected —
   `unionscope`) is C-lane future work now that the v0 projection is
   withdrawn (§5) — in v0 the resolver fronts the native store directly.
   Call sites see `graphops` interfaces; composites are the only
   switches.

## 2. What exists (survey — corrected after adversarial verification)

- **Storage contract:** `internal/storage/storage.go` defines `Storage` as
  **28 role accessors** (IssueLifecycle, IssueReader, IssueClaimer,
  ReadyClaimer, BatchCloser, BatchCreator, DependencyEditor, Commenter,
  Counter, Memories, …), each returning a role interface. Implementations:
  `*dolt.DoltStore` under decorator chains (`hook_*.go`, telemetry) plus a
  distinct **UOW/provider** chain. Documented policy: adding a required
  `Storage` method breaks out-of-tree implementers.
- **Optional-capability idiom — with teeth:** bare type assertions on a
  decorated store are a KNOWN BUG CLASS here: decorators embed an interface,
  so methods outside it are not promoted; the repo requires `UnwrapStore`
  peeling before optional assertions (`hook_decorator.go:160`) and carries a
  regression test for exactly the silent-skip failure
  (`cmd/bd/vc_recompute_test.go`). `bd serve` assembles a deliberately
  narrow role source (drops hooks, keeps telemetry) — full unwrapping there
  would change semantics.
- **Backends actually in tree:** server Dolt, embedded Dolt, and the
  UOW/proxied path. **SQLite is gone** (`cmd/bd/backend_support.go`);
  `backend/conformance` exists but `RunAll` openly does not exercise version
  control, sync, or federation families. `backend/` has a completeness guard
  requiring public aliases for every internal type reachable through the
  contract.
- **Read paths are not snapshots, and not even read-only:** role reads open
  a transaction per call (`dolt/store.go withReadTx`); issue paging is
  offset-oriented and finalized above storage; and "ready" reads run an
  advisory WRITE (waking expired defers) on server, embedded, and UOW paths;
  even `OpenForReadOnlyCommand` returns a writable store.
- **Data model:** `Issue` (`internal/types/types.go:17–190`) is very wide
  (well over 40 exported fields), `Metadata json.RawMessage` as extension
  point, `RowVersion int64` equality-only with documented partial coverage.
  `Dependency`: surrogate `ID` populated only by some read paths, endpoints,
  `Type`, `Metadata` as **string**, `ThreadID`, **no revision**;
  `depid.New(issueID, target)` keys on endpoints only — no type — so at most
  one edge per (source, target) pair, and delete/recreate would reuse the
  projected URL.
- **v0 REST precedent is narrow:** detail is the only **read**
  representation carrying `revision` (`GET /v0/beads/issues/{id}`);
  list/JSONL explicitly forbid it, while guarded **mutation** responses
  (update, close) also return their resulting revision. The
  HTTP layer maps ordinary typed errors to generated Problem DTOs centrally
  (`internal/httpapi/problem.go`) — the house pattern this plan follows.
- **Auth today:** bearer tokens grant the whole surface and carry no user
  identity or scopes (`internal/httpapi/auth.go`); routes run one
  table-driven middleware path (auth, project identity, deadlines,
  concurrency). Any BDP serving must preserve those semantics regardless of
  listener choice.
- **Wisps:** same `Issue` struct routed to a second table by storage-class
  flags; detail assembly carries `isWisp` explicitly and resolves
  issue-then-wisp; the two tables share one logical ID space enforced by
  transactional sibling checks (`cross_table_id_collision_test.go`). Wisps
  are **private/transient: excluded from export and federation by default** —
  so they are a policy decision for BDP serving, not a free rider.
- **The type shoehorn precedent:** `IssueType` is an open vocabulary
  (decision, message, molecule, gate, event, plus `types.custom`);
  `issue_type NOT IN` filters exist. Non-task semantics on the Issue chassis
  is proven practice.
- **Memories:** `memoryops` is a separate key/value plane with its own role.
  **Operator ruling 2026-09-01 (external to this tree): the memory system is
  legacy** — successor is Memory-typed Beads on this graph; no new
  investment. Note `Memories()` remains wired in `bd serve` today; retiring
  it is future work outside this plan.
- **Charter tension, named honestly:** `engdocs/PROJECT_CHARTER.md` frames
  beads as a focused issue tracker and prefers metadata over new schema for
  extension concepts. A native generic graph is a product-scope expansion.
  **P-1 decision #1 is an explicit charter ADR** — this plan does not
  proceed on implication.
- (Correction from v1: `internal/storage/domain` is UOW-specific machinery
  over `types.Issue` and `.beads`-directory concerns, not a generic domain
  landing zone. The graph package lands as its own leaf package.)

## 2b. Where BDP and the Issue/Dependency stack disagree

The conflict inventory, consolidated. Each row is a law of the pinned BDP
spec set against verified current behavior; the last column says what the
conflict costs. This section is why the v0 projection was withdrawn (§5)
and is the requirements list for any C-lane path.

| Area | BDP law (pinned) | Current Issue/Dependency stack | Consequence |
| --- | --- | --- | --- |
| Revision coverage | Every record read serves an opaque, equality-only revision; every state-changing operation on a surviving Resource mints a fresh one; a semantic no-op mints none | `RowVersion` has documented partial coverage (direct-UPDATE text paths bypass it); label writes touch only the labels table; revision is served on the detail read and mutation responses only — list/JSONL forbid it | No existing token can stand in for a BDP revision |
| Reverse transitions | A→B→A is three distinct revisions | `updated_at` is second-precision `DATETIME` with documented same-second ties; `bd import --allow-stale` restores old rows including timestamps | State-derived revisions are impossible (r5 blocker) |
| Out-of-band writes | (Implementation constraint, not spec text: BDP's revision/identity laws presuppose the authority observes every mutation) | `bd sql` permits arbitrary direct SQL; compaction rewrites text in place; backup/restore resurrects historical state | No complete mutation feed exists; only funnel (C1/C3) or storage-level observation (C2) can close it |
| Identity non-reuse | Committed Resource URLs are never reassigned — surviving deletion and epoch changes | Same-ID delete/recreate is permitted; import UPSERTs over existing IDs and accepts caller-supplied historical `created_at`; rename (delete+create) can A→B→A-reactivate an ID | Legacy IDs cannot be BDP URLs without a durable allocation/tombstone mechanism the stack lacks |
| ID grammar | Creation-time canonical IDs, multi-segment supported, reject-don't-trim | Configurable prefix grammar + adaptive-length collision-probability IDs; validation checks prefix shape, not BDP path grammar | Eligibility/surrogate policy required before any legacy ID is served (C lane) |
| Type system | One immutable nominal declared Type per Resource; descriptors with `conformsTo`; a Type describes beads or links, never both | `issue_type` is an ordinarily mutable column; open string vocabulary via `types.custom`; no descriptors, no hierarchy | Type immutability is violated by ordinary updates (r5); descriptor catalog must be built |
| Edge multiplicity | Links are first-class; no uniqueness constraint on (type, source, target) | `depid.New(issueID, target)` — at most ONE edge per (source, target) pair, type excluded from the key | Dependencies structurally cannot represent BDP Links (S2 killer #1) |
| Edge versioning | Every Link carries its own revision; owned-Link mutations version the source | Dependencies carry no revision; dependency edits never touch the source Issue's `row_lock`; `Metadata` is a `string`, surrogate `ID` populated only on some read paths | S2 killers #2 and #3; no owned-links concept exists |
| Snapshot reads | Collection cursors continue ONE logical projected snapshot across requests, bound to an authorization view | Per-call read transactions (`withReadTx`); offset pagination finalized above storage; "read-only" paths write (defer-wake); `OpenForReadOnlyCommand` returns a writable store | BDP Read semantics need a new snapshot port; existing role readers cannot serve it |
| Authorization | Per-request Authorization View — a closed projection, closed over owned Links; uniform 404 nondisclosure | Bearer token grants the whole surface; no identity, no scopes, no view concept | View mapping is a P-1 design, not a translation |
| Deletion lifecycle (Read profile) | Logical identity non-reuse survives deletion; `resource-pruned`/`resource-erased` disclosure vocabulary on reads | Deletion frees the ID for reuse; no disclosure vocabulary | The gone-family Read contract must be built native |
| Deletion lifecycle (Transactional) | Deletion results report deleted identity; tombstones and erasure records propagate on the changefeed | No tombstones; no erasure machinery | Transactional-profile obligations; arrive with P3 writes |
| Changefeed (Transactional profile) | Change Groups at Scope positions, projection advances, erasure records, no-Event erasures | The journal has a frozen vocabulary limited to Issue/Dependency/Comment payloads, emitted structurally inside issue mutations | Frozen journal stays untouched; the graph gets its own changefeed (§4 matrix) |
| Content model | One authored JSON properties OBJECT per Resource, schema-validated per Type | Typed columns (status enums, priority int, timestamps) plus a `Metadata` blob | The column→properties mapping is a design artifact (the C lane inherits §5's table) |
| Multi-writer history (Transactional/history contract) | One serialized serving authority per Scope; independently writable replicas and multi-authority merge of one Scope history are excluded | Independently writable Dolt clones merging later is a normal workflow | Decision 9: authority rule; foreign-clone graph merges out-of-contract |

## 3. Thrust 1 — the abstract data model (Go)

**What Go gives us:** structurally-satisfied interfaces — Java-interface
shape, Rust-trait spirit (implicit satisfaction, declared where consumed),
capability discovery by type assertion, composition by embedding. The repo's
role-accessor style is already exactly this idiom; `graphops` speaks it.

Two layers, strictly separated (review Blocker 1/High 6):

- **Wire DTOs are generated from the pinned BDP schema bundle** — the
  protocol layer serializes those, and only the BDP handler maps domain
  errors to generated Problem records.
- **Domain values are immutable and JSON-faithful:**

```go
package graph

// Properties is an immutable authored JSON OBJECT value (BDP properties
// are objects, not arbitrary documents): backed by copied raw bytes;
// rejects duplicate keys; preserves numbers (no float64 laundering);
// deterministic encoding; and provides the RFC 6902 §4.6 semantic-equality
// check that gates revision minting (a no-op write MUST NOT mint).
type Properties struct{ /* unexported: raw []byte + parsed index */ }

// Ref is a sum, not a naked pair: an in-Scope reference (canonicalized at
// admission, resolvable locally) or an external one (opaque, preserved
// byte-identically). Both may carry an equality-only pin.
type Ref struct{ /* unexported discriminant; constructors enforce */ }

type Revision string // opaque, equality-only

type Bead struct { /* unexported fields; accessors */ }
// ID, Type (immutable), Revision, Properties.
// NOTE: ownedLinks is not physically duplicated inside the immutable
// authored Bead value — but it IS semantically covered Bead state. The
// port assembles it from the Links themselves, in the same snapshot, and
// EVERY record projection — singleton, collection item, selection item —
// returns a BeadRecord carrying the complete ownedLinks expansion, one
// entry per declared owned type, empty entries included. (Coverage —
// owned-Link mutations versioning the source — is a storage-transaction
// law, not a struct field.)

type Link struct { /* unexported; ID, Type, Revision, Source, Target Ref, Properties */ }

type TypeDescriptor struct { /* ID, Name, Describes, ConformsTo, PropertiesSchema, OwnsOutgoing{Label,Max}, endpoint constraints */ }
```

Laws in the package, tested once: canonical-ID grammar (reject-don't-trim),
canonical-URI ordering, owned-Link trigger law (as invariant checks the
storage transactions call), semantic no-op equality. BDP problem-code
constants live in the generated protocol layer, NOT here — the domain stays
transport-neutral and speaks typed Go errors.

```go
package graphops

// Scope answers Reads within ONE resolved (workspace, authorization view,
// read snapshot). Errors are ordinary typed Go errors — transport-neutral;
// the BDP handler maps them to Problems.
type Scope interface {
    // BeadRecord = Bead + its complete ownedLinks expansion, assembled in
    // the SAME snapshot. Collections return records too — for a Bead whose
    // Type owns, the member is never elided from any projection (absent
    // exactly when the Type owns nothing). Acceptance also asserts each
    // inlined Link's type equals its entry key and source equals the
    // containing Bead — mirroring the BDP parser's contextual laws.
    // Plan-owned collection tests enforce all of it (the external matrix
    // validates the schema but does not force the optional member's
    // presence on every item).
    Bead(ctx context.Context, id string) (graph.BeadRecord, error)
    Link(ctx context.Context, id string) (graph.Link, error)
    Beads(ctx context.Context, q CollectionQuery) (Page[graph.BeadRecord], error)
    Links(ctx context.Context, q CollectionQuery) (Page[graph.Link], error)
    IncidentLinks(ctx context.Context, bead string, d Direction) (Page[graph.Link], error)
    Types(ctx context.Context) ([]graph.TypeDescriptor, error)
}
```

### The layering, in one picture

```text
bd serve (HTTP/BDP)              generated DTOs; error→Problem mapping
      │
ScopeResolver                    ← OUTER authority seam: picks workspace/store,
      │                            authorization view, and ONE ReadSnapshot
graphops.Scope (per snapshot)    ← the "trait"
      ├─ nativestore             ← v0: S1 tables, the only realization
      ├─ (unionscope + issueproj)← C-lane future, when Issues move native
      └─ (tests, CLI later)
```

**ReadSnapshot is a first-class port**: one SQL transaction (or UOW unit)
backs all reads a request makes against the native store. BDP cursors must
continue one logical projected snapshot; per-call `withReadTx` role reads
cannot provide that, so graph reads run their own snapshot-scoped queries.
Cursors bind (snapshot, view). Graph reads never call readiness roles and
never use writable "read-only command" opens (the defer-wake write is
exactly what a Read surface must not trigger). (Union cursors with per-leg
continuation are C-lane machinery, recorded in §5's historical record.)

## 4. Thrust 2 — storage: additive, capability-resolved, no breaking change

Not a bare assertion (review Blocker 3). One resolution policy with a
**typed** source — no `any` — and, per round 3, NO `UnwrapStore`:
`UnwrapStore` peels every `Unwrapper` including telemetry, which is exactly
what `bd serve` refuses ("never storage.UnwrapStore" — it performs one
concrete hook peel and keeps telemetry, `cmd/bd/serve.go:671`). Resolution
follows that model — targeted single-layer peels, each named in the result:

```go
// GraphReadSource is what a plumbing stack must yield to serve graph
// reads: the snapshot opener plus the telemetry it must retain.
// (Full contract in "The storage interfaces, concretely" below:
// ErrGraphUnsupported = absence; any other error = operational failure.)
func ResolveGraphReadSource(s storage.Storage) (GraphReadSource, error)
func ResolveGraphReadSourceFromUOW(p uow.UnitOfWorkProvider) (GraphReadSource, error)
```

With regression tests mirroring `vc_recompute_test.go` for: hook+telemetry
chains (asserting telemetry RETAINED while exactly the hook layer peels),
the notifying UOW provider, and `bd serve`'s narrow role source.
(The resolver PAIR — `ResolveGraphReadSource` for the store arm,
`ResolveGraphReadSourceFromUOW` for the provider arm — is the name; P1
uses both.)
P-legs are the real ones — with the transport distinction the tree
enforces: **`bd serve` refuses embedded Dolt permanently** (its commit
protocol cannot satisfy the server's per-request atomicity contract,
`cmd/bd/serve.go:546`); it serves from server-Dolt/UOW **and registered
backends' store sources** (`serve.go:563`) — transport claims scope to what
serve actually accepts, not to a two-item list.
**Embedded Dolt is a `graphops` storage-contract conformance leg, not a
`bd serve` transport leg**; serving BDP from an embedded workspace would
require a separately ruled read-only listener with its own snapshot
contract, out of scope here. (SQLite removed from the plan.) If out-of-tree
backends may implement the capability, `graphops` types get public aliases
to satisfy `backend/`'s completeness guard; until then the capability is
documented in-tree-only.

Native persistence (substrate S1): new `beads`/`links` tables in the normal
migration series, **plus the Type Descriptor store**: descriptors are
persisted rows (not compiled-in Go values), because every Read Scope must
advertise `types/` and mutation authorities must retain the pinned
descriptor contract closure. That means: descriptor bootstrap at graph
initialization, an operator installation mechanism for new Types, closure
validation with fingerprint retention on install, and acceptance coverage
in each phase (P1 persistence + serving, P2 the pinned Type scenarios,
P3 write-time contract validation). Revision minting gated on semantic
change (no-op preserves revision); owned-Link version coupling enforced in
the write transaction.
Where this plan says "ledger" it now means exactly one thing: the NATIVE
allocation/tombstone table written inside native write transactions (the
journal-counter pattern the tree already demonstrates). The deleted
read-time revision ledger does not return; no projection ledger exists
because no projection exists.

### The storage interfaces, concretely

How the graph attaches to the existing storage architecture, member by
member — and what changes where:

**What exists (verified):**

```text
storage.Storage (interface, 28 role accessors)      ← contract; adding a
  IssueLifecycle() / IssueReader() / ... / Memories()   required method BREAKS
      ▲ implemented by                                   out-of-tree stores
*dolt.DoltStore (concrete)
      ▲ wrapped by (each embeds + forwards the interface)
HookFiringStore → telemetry.Storage → ...           ← methods OUTSIDE the
      ▲ or, separately                                   embedded interface
uow.UnitOfWorkProvider (RunTxRead/RunTx)                 are NOT promoted
      ▲ consumed by
cmd/bd/serve role-source table                      ← one concrete hook peel,
                                                      telemetry KEPT, never
                                                      storage.UnwrapStore
```

**What is added (and precisely what is not):**

1. **`storage.Storage` does not change.** No new required method — the
   documented breaking-change policy holds. The graph capability is a
   separate, optional interface:

   ```go
   package graphops
   type Store = GraphReadSource // v0 alias: the read surface IS the
                                // store; write roles widen it in P3

   package storage
   type GraphCapable interface {
       BeadGraph() (graphops.Store, error) // error = operational failure,
   }                                       // never "unsupported"
   ```

   (`graphops` owns `Store`, `GraphReadSource`, `ReadSnapshot`, and
   `Scope`; the UOW adapter satisfies them by constructing a
   `ReadSnapshot` over the transaction a direct `NewUOW` owns, answering
   `graphops.Scope` queries from that one transaction.)

   `*dolt.DoltStore` implements it concretely. Exposure policy, settled:
   **in-tree-only for v0** — out-of-tree implementation arrives only when
   the contract is deliberately opened (claim 6), not by accident of an
   exported interface. "Unsupported" and "broken" stay distinct all the
   way up: resolution returns `(GraphReadSource, error)` with a sentinel
   `ErrGraphUnsupported`; any other error is an operational failure and
   must not be collapsed into absence.

2. **Decorators do not forward it.** Forwarding through every wrapper is
   the failure mode the tree already documents. Instead, resolution does
   what `bd serve` already does — targeted peels, telemetry retained:

   ```go
   func ResolveGraphReadSource(s storage.Storage) (GraphReadSource, error)
   // Peels the known hook layer, then — because telemetry's wrapper
   // embeds the statically typed DoltStorage, so an inner BeadGraph
   // method is NOT promoted through it — inspects THROUGH telemetry,
   // asserts GraphCapable on the inner store, and explicitly re-wraps
   // the returned graph source in the telemetry layer it peeled. Never
   // storage.UnwrapStore. The result names every peeled/rewrapped
   // layer; ErrGraphUnsupported means absence, anything else is failure.
   func ResolveGraphReadSourceFromUOW(p uow.UnitOfWorkProvider) (GraphReadSource, error)
   // The UOW access path CANNOT ride RunTxRead — it closes its unit of
   // work before returning. OpenSnapshot instead takes ownership of a
   // direct NewUOW; ReadSnapshot.Close performs the rollback/close.
   ```

   Graph reads carry the same telemetry issue reads carry — by explicit
   re-wrap, not by promotion.

3. **The one genuinely new storage primitive: the snapshot lease.**
   Existing read helpers are per-call — `withReadTx` opens and closes a
   transaction inside each role call, which is exactly why they cannot
   serve BDP snapshot semantics. `GraphReadSource` therefore exposes:

   ```go
   package graphops // owns Store, GraphReadSource, ReadSnapshot, Scope

   type GraphReadSource interface {
       OpenSnapshot(ctx context.Context) (ReadSnapshot, error)
   }
   type ReadSnapshot interface {
       Scope                    // unqualified: same package; all reads
       Close(ctx context.Context) error // answer from ONE transaction
   }
   ```

   The RESOLVERS live in `internal/storage` (not `graphops`): they take
   `storage.Storage` / `uow.UnitOfWorkProvider` and return
   `graphops.GraphReadSource` — storage imports graphops, never the
   reverse, so `storage.GraphCapable` creates no import cycle.

   A snapshot is request-scoped, and **`ScopeResolver` is its one owner**:
   it selects workspace, authorization view, and opens the snapshot; the
   handler receives and uses it; the resolver closes it when the response
   is written. This is a new *lifetime* discipline, not a new engine
   feature — the same Dolt/SQL transaction machinery `withReadTx` uses,
   held open for the request instead of per call. Two P1 design items
   with tests: pool interaction (a leaked snapshot must not pin a
   connection indefinitely), and **detached, bounded close** — rollback
   must run on a fresh bounded context, never the request context, which
   may already be cancelled (the tree documents that rolling back on a
   cancelled context burns the pinned connection).

4. **Schema machinery: normal series, no special cases.** Graph tables
   (beads, links, type descriptors, allocation/tombstone ledger) are
   ordinary migrations in the existing series, subject to the existing
   version gate (older binary refuses newer DB — §1's contract). No
   changes to the migration framework itself.

5. **`bd serve` gets a separate OPTIONAL graph field, not a role-table
   entry.** The existing role binding table is deliberately mandatory —
   it aborts on any binding error, and the HTTP layer rejects partial
   role sets — so BDP cannot join it as one more ordinary binding.
   Instead: an optional graph source on the server config, populated at
   assembly by the source-appropriate resolver (store arm or UOW provider
   arm), with a conditional route-registration seam. `ErrGraphUnsupported` leaves BDP routes
   unregistered (existing serve behavior exactly as before); an
   operational error still aborts; and capability-present-but-graph-
   uninitialized is a THIRD state with its own explicit representation,
   surfaced per whichever branch of decision 12 is ruled (honest empty
   Scope, or no routes until `bd graph init`) — never conflated with
   absence. The optional field is populated via the source-appropriate
   resolver (`ResolveGraphReadSource` for the store arm,
   `ResolveGraphReadSourceFromUOW` for the provider arm — `serve.go`
   assembles from both).

6. **`backend/` (out-of-tree implementers): additive, and precise about
   the existing machinery.** Today `RunAll` never exercises the optional
   capability families — `RunUnsupportedContract` proves their typed
   refusals instead. The graph contract therefore arrives as its own
   suite beside them (the P1 graph-storage conformance suite), with a
   refusal contract for non-capable stores. If and when the capability
   is opened out-of-tree, `GraphCapable` becomes a new ROOT of the
   public completeness guard (not merely a set of aliases) — until then
   it is in-tree-only, per claim 1.

7. **`issueops`, the journal, sync, and every legacy role: untouched.**
   The graph store is a sibling under the same DoltStore, not a layer
   over the issue roles — with the projection withdrawn, nothing in the
   graph path calls them at all.

### Replication participation matrix (review High 4, corrected round 2)

Each row is policy decided in P-1, not discovered in CI. "Byte-identical
legacy behavior" scopes to **legacy-only data and operations**; rows are
split by topology where the tree differs:

| Surface | Legacy behavior (verified) | Native graph policy (proposed) |
| --- | --- | --- |
| Dolt push/pull (server + embedded) | rows travel; embedded pushes directly | graph rows travel identically |
| Merge settlement | `versioncontrolops/mergesettle.go` already settles metadata, dependencies, migrations, config, issues, labels, comments, and events, with seven-table FK-cascade repair — conflict dispatch is a hard-coded switch, separate from an always-considered FK-repair pass; NOTE `MergeWithStrategy` returns early on clean merges and plain `Merge` bypasses settlement entirely | native graph settlement must be **centralized so every merge entry point runs it** (clean-merge early-returns and plain `Merge` included — enumerate or funnel the routes): identity/endpoint integrity, dangling-Link detection, owned-Link invariant validation. A pass can reject or quarantine invalid imported state; it CANNOT serialize two independently accepted `max`-violating writes after the fact — hence decision 9: BDP writes flow through one serving authority per Scope, and foreign-clone merges of graph tables are out-of-contract (quarantined on detection) |
| Federation type-filtering | **server-topology-specific**; deletes `issues` rows by type | graph tables get their own filter hook per topology; filtering one endpoint must also drop/deny the Link (never emit a dangling edge) |
| Journal (frozen v0 vocabulary) | Issue/Dependency/Comment payloads only | **native graph events are excluded**; a separate graph changefeed carries them; the frozen vocabulary is not extended |
| Export/JSONL (contract class) | contractual shapes | graph gets its own export lane; legacy shapes untouched |
| Backup | whole-database state (a different contract class from export) | graph tables ride along by construction |
| Wisps | private/transient; excluded from export/federation by default | **P-1 policy decision** — excluded from BDP serving in v0 (proposed) |

## 5. Thrust 3 — Issues/Dependencies beside the graph

### The seam decision (HISTORICAL — C-lane record; superseded for v0 by
the withdrawal below)

What follows is preserved as design input for the C lane, not v0 scope.
Option A (projection port + union composite) over B (peer interfaces on the
structs), C (storage unification now), D (call-site switches) — with the
review's correction adopted: the union is the *representation* seam only,
subordinate to the outer authority seam, snapshot-scoped, and:

- **Duplicate full Resource ID across legs is an integrity error, never
  precedence.** The v1 "native first, then legacy" shadowing is withdrawn.
- **Namespace AND ledger — they answer different laws (round-2
  correction):** a reserved native namespace prevents *collisions*; it does
  nothing for *lifetime identity non-reuse* (a deleted projected Issue ID
  must never be reassigned — BDP's no-URL-reassignment law survives
  deletion and epoch changes). So: namespace disjointness for allocation,
  PLUS a durable allocation/tombstone guarantee behind every exposed URL,
  native and projected — covering legacy import's same-ID UPSERT and
  Dependency delete/recreate reuse. And an **eligibility policy for legacy
  IDs** (P-1): an issue ID that is not a canonical BDP path segment is
  omitted, mapped to a stable surrogate, or fails Scope projection — ruled,
  not improvised (current validation checks prefix shape, not BDP grammar).
- **Cross-realization Links** (C-lane decision when projection returns):
  if a native Link may target a projected Issue, every legacy deletion
  needs a graph coordinator hook (else dangling edges); if forbidden, the
  Type constraints must say so. Moot in v0 — nothing legacy is served.
- Multi-repo routing stays where it is — `ScopeResolver` wraps the existing
  owning-store resolution; the union never spans stores.

### The substrate decision (SETTLED for v0: S1)

S1 (new tables) is the v0 substrate — with the projection withdrawn there
is nothing for a chassis substrate to buy: S2's free-rider argument was
chassis sync for *legacy interop*, and v0 has none. The historical
scorecard stands as C-lane input: S2 fails conformance on three laws
(`depid` admits one edge per endpoint pair — no type, no multiplicity;
Dependencies carry no revision; dependency edits don't version the
source); S2-lite was the fallback only while a projection existed.

### The Issue projection is withdrawn from v0 (round-5 conclusion)

The conflict inventory in §2b is the full map; rounds 3–5 tested every
read-side fidelity mechanism against it, and each fell to tree
counterexamples:

- **State tuples** (r3): label add→remove recreates the tuple; direct text
  restoration moves only second-granularity `updated_at`.
- **Read-time witness ledgers** (r4): witness reads, not transitions —
  legacy A→B→A between reads serves the old revision.
- **Complete-representation state hashes** (r5): `updated_at` is
  second-precision `DATETIME` and the tree documents same-second ties, so
  A→B→A within one second reuses the hash; `bd import --allow-stale`
  deliberately restores old rows including `updated_at`; `bd sql` permits
  arbitrary direct writes no witness can enumerate; merge resolution and
  backup restore reinstate historical rows.
- **Birth-identity URLs** (r5): same-second recreation collides; import
  accepts caller-supplied historical `created_at`; import-over-existing
  does not converge `created_at` across replicas; rename A→B→A reactivates
  the original URL; database restore resurrects old tuples. BDP's
  never-reassign law cannot be met.
- **Type immutability** (r5): legacy `issue_type` is ordinarily mutable;
  BDP declared Types are immutable.

Mutation-time witnessing — the only remaining mechanism — would require
instrumenting every legacy write path including `bd sql`, which is
arbitrary SQL and cannot be completely instrumented even in principle.
The conclusion is structural, not incremental: **a store that permits
timestamp ties, stale restores, arbitrary SQL, and identity resurrection
cannot be projected into BDP's revision and identity laws by any read-side
mechanism.** So v0 withholds the projection:

- The v0 BDP Scope serves **native beads and links only**.
- Issues keep their existing surfaces (CLI, REST v0, JSONL) untouched.
- Issues join the graph when storage unification (Option C) moves them
  into the native store — where operation-local revisions and durable
  identity are properties of the write path, not reconstructions. The
  union composite and this section's counterexample record are the design
  input for that future lane.
- The `unionscope`/`issueproj` machinery drops out of v0 scope; the
  authority seam (`ScopeResolver`: workspace, view, one ReadSnapshot) and
  `GraphReadSource` remain — they serve the native store.

### The C lane: paths to Issues/Dependencies on the graph (informative)

Operator direction (2026-09-02): the withdrawal stands for v0, and the C
lane should be sketched now. The round-3–5 counterexamples fix the shape of
the solution space: **uniform versioning requires that every mutation
either funnels through one write path or is completely observed at the
storage layer.** Read-side reconstruction is proven impossible. That yields
three paths:

- **C1 — Funnel with a legacy compat shim** (the operator's sketch): Issues
  and Dependencies become native beads/links; the legacy surfaces (CLI
  verbs, REST v0, JSONL, journal, sync) are reimplemented as a compat
  adapter OVER the graph store, reproducing legacy behavior byte-for-byte.
  Versioning is uniform because every mutation goes through the native
  write path. The crux is what "keep the current code path" means at the
  storage layer: if legacy code keeps writing legacy TABLES, the bypasses
  persist and uniformity fails; so C1 means legacy *behavior* preserved
  over graph *storage* — and the hard cases are exactly the round-5
  killers re-specified deliberately: `bd sql` (verification task: whether
  Dolt supports legacy-compatible updatable views that can route DML
  through graph revision/tombstone semantics — else direct SQL is
  re-scoped), import `--allow-stale` (becomes a versioned operation),
  backup/restore (restores history, not just state; decision 11). The
  wisp precedent (plane routing at storage, surfaces unchanged) is the
  house style for this move.
- **C2 — Observe: a complete mutation feed at the storage layer**: legacy
  tables remain the record for legacy surfaces; a storage-level observer —
  DB triggers on the legacy tables feeding a transition log in the same
  transaction — would give the graph per-operation revisions and
  tombstones without touching legacy code paths, and would be the only
  observation variant that survives round 5 IF the following verification
  tasks all pass (none is established fact): (a) Dolt trigger availability
  and transactional semantics; (b) whether direct `bd sql` DML actually
  fires them (noting `bd sql` is unavailable in embedded mode and runs via
  direct SQL-server or proxied-server paths — and UOW is an access path,
  not a third storage engine); (c) trigger-row behavior under replication,
  merge, and restore; (d) Scope URL/epoch handling after restore
  (decision 11).
- **C3 — Cutover**: one-time migration, graph store becomes the only
  store, legacy tables dropped or frozen read-only. Maximum uniformity,
  no dual bookkeeping, maximum one-shot risk; `bd sql` compat ends or
  becomes views. Realistic only after C1's compat adapter exists and has
  soaked — C3 is C1 minus the legacy tables, not an alternative to it.

Sequencing implication: C1 and C3 share the compat-adapter investment; C2
is the only path that leaves legacy storage untouched. A future C-lane
ruling chooses funnel (C1→C3) vs observe (C2) — and none of it blocks or
changes v0.

## 6. Addressing

One workspace = one Scope. Scope URL scheme is a P-1 decision (config key
vs derived). Native IDs mint under BDP creation-time rules (supplied
multi-segment or generated flat, reject-don't-trim) against the native
allocation/tombstone ledger. No legacy IDs are served in v0.

## 7. Phasing (re-sequenced per review; each phase exits green)

- **P-1 — Decisions and pins (no code):** charter ADR; ratify the
  projection withdrawal (v0 Scope = native only); Scope URL/identity;
  native allocation/tombstone design; serving authority; replication
  matrix rows; auth-view mapping for bearer-token reality;
  listener/authority-semantics choice. (The BDP pin is already written in
  §0.) *Exit: every row ruled by Donna, recorded in this doc.*
- **P0 — Contracts:** generated wire DTOs from the pinned schema; immutable
  domain values (`Properties`, `Ref` sum, records); pure validators; typed
  error vocabulary. *Exit: model laws 100% table-tested; DTO round-trip
  against pinned schema fixtures.*
- **P1 — Native read storage (S1):** tables + migrations (descriptor
  store included); typed snapshot-source resolution (`GraphReadSource`)
  with single-request snapshot consistency and the zero-legacy-writes
  regression (defer-wake); the resolver pair across the storage legs —
  `ResolveGraphReadSource` for server/embedded Dolt (embedded as
  storage-contract leg), `ResolveGraphReadSourceFromUOW` for the UOW
  access path — with decorator regression tests; an internal, non-BDP
  bootstrap/fixture write API (the only writer until P3) enforcing the
  allocation/tombstone ledger; replication-matrix gates. *Exit: a NEW
  graph-storage conformance suite — defined in this phase under
  `backend/conformance`, enumerating the storage contract (snapshot
  isolation, ordering, ledger enforcement, descriptor persistence) —
  green on all legs (descriptor persistence AND inventory serving
  included); cross-request cursor stability is explicitly NOT a P1
  claim.*
- **P2 — Protocol Read over the native store:** snapshot-bound cursors —
  including the **cross-request continuation mechanism** BDP requires
  (later requests continue the same selected set, projection, and
  revisions): a durable snapshot registry, materialized result sets, or
  Dolt `AS OF` identity surviving through cursor expiry — chosen by ADR in
  this phase; BDP handler through the existing middleware path
  (auth/project/deadline semantics preserved); run the **pinned** external
  BDP Read matrix as a target. *Exit: the pinned matrix green with its
  own provenance split honored — packaged rows proven at the packaged
  public boundary, self-certified in-process rows via the in-process lane
  and labeled as such (the pinned artifact is explicit that they are not
  black-box conformance) — plus a beads-owned public-boundary
  cursor-stability test across requests.*
- **P3 — Writes and CLI, gated on its own ADR AND on upstream spec
  artifacts:** the pinned write-profile envelope (an owned-Link mutation's
  result must also report the source Bead's resulting revision), the
  owned-Link Event delta, AND the later-profile artifacts BDP itself marks
  pending — the sequence/idempotency envelope schemas, problem rows, and
  conformance artifacts for the write profiles. Profiles are **Scope-wide**
  (uniformity law), and the Event-delta gate binds exactly the profile
  that has Events: a Scope containing owning Types cannot advertise the
  **Transactional** profile until the owned-Link Event delta exists —
  Read+Update has no Events and is not blocked by it — and this is a
  Scope-level gate, not a per-Type advertisement. Write tests require: create/property-update
  mint fresh Link AND source revisions; **deletion mints nothing for the
  deleted Link** — its result reports the deleted identity plus the
  source's fresh revision (the final live revision is Transactional
  `DeletedData` Event material, not a result member); target revision
  unchanged throughout; both surviving revisions preserved on semantic
  no-op. Then tombstones,
  endpoint constraints, replication of writes; only then
  `bd bead`/`bd link` verbs. *Exit: the pinned write-profile conformance
  artifacts green when they exist upstream, plus beads-owned transaction,
  identity/non-reuse, deletion-result, installed-Type-contract
  validation, and replication tests at the public boundary.*

## 8. Process

All work on `feat/bead-graph`; slices land by PR with adversarial
convergence; the feature branch merges to main only at phase exits behind a
differential gate proving same-version legacy behavior unchanged. Spec
changes go to gastownhall/bdp first — this plan's §0 pin is the enforcement
of that.

## 9. Decisions requested from Donna (the P-1 list)

1. **Charter ADR:** does beads core own a general BDP graph (v0: native
   beads/links beside untouched Issues), or is the work deferred entirely?
2. **Substrate:** settled for v0 — S1 (§5); ratify.
3. **Native allocation/tombstone design** (§4): the native write path owns
   a durable allocation/tombstone table (journal-counter pattern);
   committed URLs are never reassigned. (Namespace-vs-issue-grammar and
   legacy-ID eligibility move to the C lane with the projection.)
4. *(Moved to the C lane — cross-realization Links cannot arise in v0;
   see §5's historical record.)*
5. **Ratify the projection withdrawal** (v0 Scope serves native beads and
   links only; Issues and Dependencies join via the C lane) — §5's
   round-5 counterexample record is the evidence base.
6. **Wisps**: moot for v0 serving (nothing legacy is served); recorded as
   a C-lane visibility decision.
7. **Scope URL scheme**; and the listener question restated properly: not
   "same port or new," but which choice preserves today's bearer-auth and
   project-identity semantics unchanged while carrying the P-1
   authorization-view mapping (no view concept exists today — §2).
8. **Journal/changefeed:** ratify "frozen v0 journal untouched; separate
   graph changefeed."
9. **Serving authority per Scope** (round-4): BDP excludes independently
   writable replicas and multi-authority merge from one Scope history.
   Rule one of: a single serialized serving authority per canonical Scope
   URL (proposed), or distinct Scope URLs for independently writable
   clones.
10. *(Absorbed into decision 5 — the round-5 review closed this fork:
   withholding is the only conformant option.)*
11. **Restore vs identity** (round-6): a whole-database restore of an older
   backup forgets later allocations, permitting URL reuse under the same
   Scope URL — which BDP forbids. Rule one of: durable allocation
   preservation across restore (the ledger survives outside the restored
   state), or canonical Scope-URL rotation on any identity-losing restore.
   An epoch change alone is insufficient.
12. **Empty-at-birth Scope policy** (round-6): for an existing legacy-only
   workspace, either advertise an honest EMPTY native Scope (stable URL;
   required `beads/`, `links/`, `types/` all present and empty), or
   advertise no BDP Scope until explicit graph initialization
   (`bd graph init`, proposed). Either way: tests prove legacy Issues
   never leak into native inventories, and a registered backend without
   `GraphReadSource` keeps its existing `bd serve` behavior — BDP routes
   absent, never a startup failure.
