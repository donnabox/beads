# BDP graph store — architecture and design

**Status:** Draft v5 (W-arch, after four council rounds) — feat/bead-graph
**Date:** 2026-09-02
**Companion:** `BDP_BEAD_GRAPH_PLAN.md` (the plan and its twelve rulings) and
`BDP_GRAPH_CLI_AND_STORAGE_SPEC.md` (the detailed CLI and storage-interface
changes). This document is the *shape*: what the pieces are, where they live,
how a request flows, and where this design corrects the plan after the
tree's own conventions were read closely. Four three-reviewer councils
(Claude, Codex, Gemini) reviewed v1–v4; §2 records what changed and which
changes amend a ruling rather than a mechanism. Rounds 3 and 4 converged on
one area — the authority mechanism — and v5 states it on the precedents the
tree already has: the journal's single-row sequence counter, the ephemeral
`leases` table, the scoped `doltAddAndCommit`, the workspace exclusive gate,
and Dolt's non-fast-forward push refusal.

## 1. The one-paragraph version

The graph store is a new **plane** beside issues and memories: a public leaf
contract package (`graphops`) declaring the value types, the laws, and six
role interfaces, reached through **role accessors on `storage.Storage`**
(`BeadGraph*`, to keep clear of the issue-graph `GraphCounter()`) exactly as
`issueops` and `memoryops` are — declared explicitly by every decorator
(promotion is the failure mode the censuses catch), wrapped by telemetry,
recursed unwrapped by the hook layer — with one shared transaction-level
body under `internal/storage` taking a `DBTX`-shaped runner so that both
Dolt stores *and* the unit-of-work leg call the same code, proven by
`backend/conformance` role contracts wired on all three legs and guarded by
the existing coverage gates. BDP is served by the existing
`internal/httpapi` server as a **conditional second route table** behind the
same `route()` middleware — always from the **unit-of-work leg**, because
`bd serve` refuses embedded Dolt permanently, so every serving workspace is
a SQL-server workspace. `bd bdp serve` is the strict command over that
server: it mints and it requires a Scope this workspace holds; `bd serve`
mounts the same rows when it holds an already-minted Scope and never
refuses on account of the graph. Every operation is **one role call, one
transaction**, asserting a **store-owned authority witness** — loaded by the
accessor from a clone-local file bound to this installation, checked
against a hash-chained ledger head and a fencing cell the mutation itself
must update — so a clone, a restore, a copied file, or a promotion
elsewhere is refused by the store, and no caller can supply the witness.
Fences compose by hazard: a shared database gets a lease; a configured
remote gets fetch → ancestor check → scoped commit → push. The CLI's graph
verbs live under `bd bdp …` and reach the same roles through the same
accessors, or — when a workspace is wired as a client — speak BDP to a
designated server.

## 2. Corrections to the plan, and proposed ruling amendments (read this first)

Two kinds of change are recorded here and must not be confused. A
**mechanism correction** replaces something the plan's §4 *proposed* with
the house idiom that already solves the problem; the ruling it serves is
unchanged. A **ruling amendment** changes text the operator ratified in §9;
it is *proposed* here and takes effect only when ruled. Nothing below is
settled until ruled, and the spec marks every A8-dependent sentence as such.

### 2a. Mechanism corrections (no ruling changes)

| Plan §4 mechanism | Replaced by | Why |
| --- | --- | --- |
| `GraphCapable` as a *separate optional interface*, resolved by targeted decorator peels (`graphsource` package) | **Role accessors on `storage.Storage`** — the house rule — *if A8 option A is ruled*; otherwise the optional-interface shape returns with the costs stated under A8 | Every accessor lives on `Storage` (28 today); `DoltStorage` embeds it; both decorators embed `DoltStorage`. An accessor added to `Storage` *compiles* through every decorator by promotion — and promotion is exactly what the reflection census in `role_accessor_decorator_test.go` (and its telemetry twin) rejects: each decorator must **declare** the accessor. An *optional* interface is promoted the same way, **unwrapped**, and no census sees it — the telemetry bypass round 1 objected to. |
| A `ResolveGraphReadSource` policy that "re-applies telemetry" after peeling | Nothing — **the accessors carry the layers** | This is what accessors are *for* in this tree. |
| `backend/types.go` aliases for every `graphops` type | **No aliases.** `graphops` is a public root package, like `issueops` | The completeness guard demands aliases only for types under `internal/`; `backend/types.go` states `issueops` is deliberately not re-exported. |
| `internal/graphapi` as a separate meaning-function package | **The laws live in `graphops`** | A `graphops` constructor calling `internal/graphapi` would break the leaf's import rule and cycle. |
| A `ReadSnapshot` handed from a resolver to a handler; a caller-supplied expectation (v2/v3) | **One role call, one transaction, asserting a store-owned witness** (amendment A1 for the ruling-9 wording) | Nothing escapes a role; the authority check and the read share a snapshot; a witness the *caller* supplies is forgeable from the replicated rows. |
| `SELECT … FOR UPDATE` on the Scope row to serialize the ledger (v4) | **A single-row sequence counter incremented with `UPDATE`** — the journal's `bd_events_seq` precedent | Dolt has no row locks; `FOR UPDATE` / `SKIP LOCKED` are parse-only no-ops (the tree says so in five places). Contention on the counter row is a commit-time serialization loser that `withRetryTx`/`RunTxResult` replay. |

### 2b. Proposed ruling amendments (pending operator ruling)

| # | Ruling | Earlier drafts said | v5 proposes | Evidence |
| --- | --- | --- | --- | --- |
| A1 | **9** ("the snapshot lease" in the obligation list) | dropped silently (v1); caller-supplied expectation (v2–v3) | **The per-call transaction is the v0 lease and asserts a store-owned authority witness:** the accessor loads the witness from the clone-local file and the body checks it *inside the same transaction* — Scope row identity, the hash-chained ledger head (exact-prefix identity), the fencing cell (a mutation must *update* it and see one affected row; a read may select it), and the graph-state version (per-table hashes, not the working-set hash, which every heartbeat moves). Public request types carry **no** authority fields. Cross-request continuation is P2's cursor ADR; the cursor type is opaque from P1. Ruling 9's obligation reads "single-transaction operations under a store-asserted authority witness". | Rounds 1, 3, 4. |
| A2 | **7b + 12** (listener; "`bd serve` creates the Scope on first serve") | a sibling server (v1); `bd serve` mints (v2–v4) | **BDP rows mount inside `internal/httpapi`** as a conditional second route table behind the same `route()`. **Only `bd bdp serve` mints** (once the Scope URL is a tracked project fact — A6 — a plain `bd serve` on any clone would otherwise become the authority by being first). `bd bdp serve` inherits `errServeReadonly` wholesale and refuses without a held Scope; **`bd serve` converts every graph failure — capability, identity, fence — into "rows absent + notice" and keeps the legacy surface up**; it is byte-identical when no URL is configured. **Every serving workspace is a SQL-server workspace** (`errServeEmbedded` is permanent), so the serving leg is the unit-of-work provider, and the fences live there. | Round 2 (`Capabilities()` derives from `routeTable`); round 3 (tracked URL); round 4 (mint trigger; UOW leg). |
| A3 | **12** / §4 lifecycle (`bd bead`, `bd link`, `bd graph …`) | verbs "reserved now" | **Everything under `bd bdp …`**, with a `CommandPath()`-keyed policy authoritative at every leaf-name call site (spec Part A names them and pairs the Cobra-walk test with a source scan for `cmd.Name()` consumers). | `bd link`, `bd graph`, `bd restore`, `bd promote` exist; leaf-name policy inheritance. |
| A4 | §3 layering | values moved to `graphops` silently | **Record it:** values, laws, and roles in public `graphops`; accessors named `BeadGraph*`. | Package layout is public API. |
| A5 | **11** (ledger "restorable independently … its own migration") | dolt-ignored table (v2, disproved by probe); project-id-stamped file with a scalar mark (v3); hostname-keyed file (v4) | **(i) The clone-local half is `.beads/graph-authority.local.json`**, bound to an **installation key** (a random id generated once under the user's config dir — the user-global precedent — plus the canonical `.beads` path; never the hostname, which the tree's own `NodeID` note refuses for containers, DHCP, and shared servers). Written only by the witness manager under a bounded exclusive lock with monotone read-modify-write and directory fsync; **transitions (mint, promote, rotate, ledger apply) are two-phase**: a pending record before the fenced transaction, finalized after it commits and publishes, recovered on the next load. The manager **ensures** the ignore entry and that the path is not git-tracked before the first write; an already-tracked witness is a hard doctor error. **(ii) The ledger is an append-only, hash-chained event table** (`mint`, `install`, `update`, `promote`, `rotate`, `allocate`, `tombstone`, `refuse_url`) with a single-row sequence counter; the witness records the head `{seq, hash}`, so a restore to any earlier point or a different history is refused. **No ledger event exists before mint**: the built-in catalog is installed inside the fenced Mint transaction. Snapshots are manifest-bound contiguous ranges; `LedgerApply` and `bd bdp restore` have their own recovery predicate (exempt from the head check they exist to repair). **(iii)** providers declare `LedgerDurability`; `bd bdp restore` rotates unless continuity is shown. Residuals, stated: a whole-directory filesystem snapshot rewinds file and state together; acknowledged-but-unwitnessed writes before a crash are a P3 write-profile obligation. | Rounds 2–4. |
| A6 | §4 lifecycle (metadata.json; `BD_BDP_TOKEN`; `BDP_SCOPE_URL`) | both files; env token; all in `config.yaml` (v3) | **`bdp.scope_url` is a project fact in tracked `config.yaml`** (yaml-only; `BDP_SCOPE_URL` read *first*, then `BD_BDP_SCOPE_URL` — the `BEADS_ACTOR` precedent; **refused by `bd config set` once minted**: the URL changes only through `bd bdp promote --rotate-url` / `bd bdp restore`, which update the Scope row and `config.yaml` as one recoverable transition). **`bdp.client`, `bdp.server`, `bdp.insecure_http` are per-workspace and live in `config.local.yaml`** — the tree's machine-specific override that viper merges over `config.yaml` — written by `bd init --bdp-server` and `bd bdp client`; generic `bd config set` refuses them. `config.local.yaml`, the witness, and its lock join all three doctor lists. No env-carried token; no token key in config; `bdp.client` blocked from env. | Rounds 2–4. |
| A7 | **9** (promotion "explicit and epoch-rotating") — *new* | v4 partitioned fences by storage mode | **Fences compose by hazard, and every graph mutation is fenced inside its transaction.** *Hazard S — more than one workspace on one database* (every SQL-server topology): a dolt-ignored **`graph_authority_lease`** row holding `{installation key, process nonce, epoch, expires_at}` (the `leases`/`RunTxEphemeral` precedent); every mutation `UPDATE`s it `WHERE holder = self AND epoch = self` and requires one affected row — a takeover is then a serialization loser, replayed into a refusal; expiry fails closed; renewal is ephemeral (no history). *Hazard R — a remote is configured* (push/pull replication): `Mint`/`Promote`/`Rotate`/`LedgerApply` are one gated operation — `DOLT_FETCH`; require the remote-tracking HEAD to be an **ancestor** of local HEAD (`DOLT_MERGE_BASE`; unpushed issue-plane commits do not block); the fenced transaction; a **scoped** commit (`DOLT_ADD` of the graph tables only, then `DOLT_COMMIT` — the `doltAddAndCommit` precedent, never `-Am`); `DOLT_PUSH` with a **typed non-fast-forward classifier** (pinned by a real-Dolt test): on non-FF, fetch and re-read the remote Scope row — changed `(authority_id, epoch)` → reset **to the recorded pre-operation local HEAD** and refuse; unchanged (issue-plane divergence only) → "sync required", retryable; any other failure → keep the local commit, touch nothing. A serving process fetches on `bdp.authority_heartbeat`, compares `(authority_id, epoch)`, and fails closed after `bdp.authority_heartbeat_grace`. P3 writes on hazard R are **push-on-commit**. Both hazards → both fences. *Neither hazard* (embedded, no remote: a CLI-only workspace) — **not promotable in place**: a copy of such a database is an unrelated second copy, so `bd bdp promote` there requires `--rotate-url` (a new Scope). Force-push routes bypass the fence as operator acts, like `bd sql`. All of this runs under the **workspace exclusive gate** (`internal/workspacegate`, the `bd backup restore` precedent) and on the **unit-of-work leg** (the procedures exist on `doltVersionControlSQLRepository`) as well as the store legs. | Round 4 (Codex C1–C3, H7; Claude H1, M1, M2). |
| A8 | plan §1 constraint #1 ("zero compatibility degradation") and ruling 12 — *new* | v2–v3 treated the break as settled | **Two options; the docs do not pre-empt.** **A (recommended):** constraint #1 is scoped to *behavior* — every in-tree topology and existing workspace is byte-identical in gate (non-TTY) output (`bd init`'s skip notice is debug-level only; migration progress lines print to a TTY as they do for every migration); out-of-tree `backend/` implementers take the source break `storage.go` already declares (six one-line `ErrUnsupported` stubs; the joint `ReadyClaimer`/`BatchCloser` CHANGELOG entry is the precedent). Under A the compiler catches every type *compiled as* a `Storage`; the censuses cover decorators and facades. **B:** an optional `BeadGraphCapable` interface — no source break, but embedding promotes it unwrapped, so each decorator implements it explicitly, a capability census is written, and every consumer needs a resolver. | `storage.go` contract text; `backend.Storage = storage.Storage`. |

Two decisions the plan does not yet contain, surfaced for ruling rather
than designed around:

- **Enforcement boundary for out-of-role writes.** `bd sql`, the proxied
  `RawSQLUseCase`, force-push, and a merge can change graph tables without
  allocation, authority, revision, or owned-Link coupling checks. v5's
  position (§7): out of contract; the **state-change validator** rejects
  invalid or foreign-authority graph state; DB-privilege or trigger
  enforcement is a C-lane verification task. To be ruled before P3.
- **Replication/merge ADR as a P1 gate.** The merge entry points are not
  four Go functions: `mergesettle.go` exports seven, `fastforward.go` and
  `automerge.go` more; `CALL DOLT_PULL` merges inside Dolt on every pull
  route; the UOW leg's `doltVersionControlSQLRepository` calls
  `DOLT_MERGE`/`DOLT_PULL`/`DOLT_PUSH` directly; embedded federation sync
  fetches and merges; the remote-migrate gate does a fast-forward
  `DOLT_MERGE`. A wrapper cannot see a server-side merge, so the validator
  runs **on every observed graph-state-version change**: the witness
  records the ordered `DOLT_HASHOF_TABLE()` hashes of the graph tables and
  the HEAD commit; a body that sees a different version returns
  `ErrStateChanged` *without* validating inside the held transaction; the
  accessor then validates under singleflight in its own transaction (an
  ancestry check `DOLT_MERGE_BASE(recorded HEAD, HEAD)` catches rewinds a
  ledger-independent provider would miss; row provenance
  `last_authority_id`/`last_epoch` identifies foreign *updates*), advances
  the witness, and retries once. A superseded clone that pulls resets to
  the remote; push-on-commit means none of its discarded events were
  acknowledged. Prefer refusal of foreign-authority deltas over invented
  merge rules. Lands before the graph migrations.

What none of this changes: ruling 9's level — the authority is the graph
store as reached through the normalized storage abstraction, on any
provider; Dolt is the reference realization; the CLI verbs and the BDP
handler are both clients of that abstraction.

## 3. Packages and their imports

```text
graphops/                        PUBLIC LEAF (sibling of issueops/, memoryops/)
  ├─ types.go                    Bead, Link, Ref (in-Scope|external), Properties, Revision,
  │                              Attribution, TypeDescriptor, OwnedLinkDecl, OwnedLinkGroup,
  │                              ScopeIdentity, LedgerEvent, LedgerManifest, Cursor (opaque)
  ├─ laws.go                     pure functions: canonical-ID grammar (reject, never trim),
  │                              code-unit ordering, JSON canonicalization, RFC 6902 §4.6
  │                              equality (the no-op gate), Scope-URL validation, ledger
  │                              event hashing
  ├─ reader.go                   Reader: Bead, Link, Beads, Links, IncidentLinks
  ├─ types_role.go               DescriptorReader: Descriptors, Descriptor
  │                              TypeInstaller: Install (idempotent, fingerprint-keyed)
  ├─ identity.go                 IdentityReader: Read, LedgerDurability
  │                              ScopeBootstrapper: Mint (once; installs the built-in catalog)
  │                              Admin: Promote, Rotate, LedgerSnapshot, LedgerApply,
  │                                     MarkUnverified, ClearUnverified
  ├─ writer.go                   (P3) Writer — born whole with the write-profile ADR
  └─ errors.go                   sentinels ALIASED from beadserrors: ErrNotFound, ErrValidation,
                                 GoneError{Path, State}, ErrNoScope, ErrScopeExists,
                                 ErrNotAuthority, ErrStateRewound, ErrStateChanged,
                                 ErrSyncRequired, ErrURLReused, ErrRepresentationTooLarge,
                                 ErrNotServedYet
  NO authority fields on any request type — the witness is the store's.
  imports: stdlib + beadserrors ONLY (the memoryops precedent)

internal/storage/authority/      THE WITNESS MANAGER (clone-local half; no SQL)
  Witness{InstallationKey, ScopeURL, AuthorityID, Epoch, LedgerSeq, LedgerHash,
          StateVersion, StateCommit, Unverified, GrantedAt, Pending *Transition}
  installation id: random, generated once at os.UserConfigDir()/bd/installation-id
    (the user-global config precedent); InstallationKey = sha256(id ":" realpath(.beads))
  Load(beadsDir): a plain read (atomic rename makes it complete or absent);
    a Pending record triggers recovery (below) before any assertion
  Advance(beadsDir, fn): bounded exclusive flock on .beads/graph-authority.lock
    (internal/lockfile; poll with timeout, the workspacegate precedent) → read → fn →
    LedgerSeq never decreases → atomicfile write → fsync the directory
  Begin(transition) / Finalize / Abandon: the two-phase record for mint/promote/
    rotate/ledger-apply; preflight ENSURES the .beads/.gitignore entry
    (EnsureGitignoreForBeadsDir) and that the path is not git-tracked (git ls-files)
  Recovery on Load with Pending: ask the store (Scope row, ledger head, and — hazard R —
    the remote-tracking ref) whether the transition committed → Finalize, else Abandon

internal/storage/graphops/       TX-LEVEL SHARED BODY — all three legs call it
  type DBTX interface { ExecContext; QueryContext; QueryRowContext }
    (the issueops.DBTX shape; *sql.Tx and domain/db.Runner both satisfy it)
  assertAuthorityInTx(ctx, tx, w, mutating bool) — first statements of every protected
    body: Scope row == (w.ScopeURL, w.AuthorityID, w.Epoch); ledger event at w.LedgerSeq
    has hash w.LedgerHash and MAX(seq) >= w.LedgerSeq; hazard S: the lease row — SELECT for
    a read, UPDATE … WHERE holder=self AND epoch=self with one affected row for a mutation;
    graph-state version (per-table hashes) == w.StateVersion else ErrStateChanged
  nextLedgerSeqInTx: UPDATE graph_ledger_seq SET next = next + 1 WHERE id = 0 (the
    bd_events_seq precedent) — contention is a serialization loser, replayed
  ReadBeadInTx / ReadLinkInTx / SelectBeadsInTx / SelectLinksInTx / IncidentLinksInTx /
  DescriptorsInTx / DescriptorInTx / InstallDescriptorsInTx / ReadIdentityInTx /
  MintScopeInTx (Scope row + mint event + catalog install events, one transaction) /
  PromoteInTx / RotateInTx (refuse_url + rotate, one transaction) / LedgerSnapshotInTx /
  LedgerApplyInTx (own recovery predicate) / ValidateStateInTx / graphStateVersionInTx
  SeedBeadInTx / SeedLinkInTx — the P1 fixture writer, ledger-enforcing; call sites
    are _test.go files only (a source-scan test enforces it)
  No exported constructor. Inside the charter's Storage Boundary. A NEW, stricter
  depguard rule denies this package to everything except the three legs and domain/db
  (cmd/bd imports the issueops tx-body package directly today; this rule is stricter).

internal/storage/dolt/beadgraph_*.go, internal/storage/embeddeddolt/beadgraph_*.go
  accessors (CLI paths; embedded never serves): load the witness, wrap the body in
  withReadTx / withRetryTx (server) or withConn (embedded, //go:build cgo); mutations
  commit SCOPED (DOLT_ADD graph tables + DOLT_COMMIT, the doltAddAndCommit precedent);
  hazard R transitions run the gated publish sequence; nil receiver → ErrUnsupported

internal/storage/domain/beadgraph.go + internal/storage/domain/db/beadgraph.go
  BeadGraphUseCase over a db.Runner, delegating to the InTx bodies — the
  IssueUseCase() → CompareAndSetMetadataKeyInTx precedent; the publish sequence uses
  the DOLT_FETCH/DOLT_PUSH procedures already on doltVersionControlSQLRepository
internal/storage/uow/beadgraph_*.go        THE SERVING LEG
  uow.UnitOfWork gains BeadGraphUseCase(); doltSQLProvider gains a beadsDir field
  (newSQLServerUOWProvider receives it today and drops it) and timedProvider forwards
  it; the provider's six accessors: RunTxRead for reads; RunTxResult with a commit
  message for mutations (scoped staging); RunTxEphemeral for lease renewal (no history —
  the bd-lrgn1 form). The notifying provider declares each explicitly; parity test.

internal/storage/storage.go        + six BeadGraph* accessors (one line each)
internal/storage/hook_beadgraph_*.go   declared; recurse UNWRAPPED (no graph hook vocabulary)
internal/telemetry/beadgraph_*.go  every method spanned storage.op / storage.done
backend/conformance/beadgraph_*_contract.go   role contracts; RoleContractBundle fields;
                                   role_bundle_cases.go rows; wirings on all three legs

internal/httpapi/bdp_routes.go     bdpRouteTable — conditional rows behind route() (P2);
                                   handler() reads cfg.Graph — a first for that function —
                                   and TestSpecRouteParity keeps excluding the rows
internal/httpapi/bdp_handlers.go   handler = serializer over graphops roles
internal/httpapi/bdp_problem.go    typed graph errors → BDP Problem records, here only
internal/httpapi/bdpwire/          GENERATED DTOs from the vendored, pinned bdp-v0 schema — P0
internal/bdpclient/                graphops.Reader/DescriptorReader over the wire

cmd/bd/bdp.go                      `bd bdp` root; subcommands in cmd/bd/bdp_*.go;
                                   bdpRootPolicy keyed by CommandPath()
cmd/bd/bdp_serve.go                thin: serveDatabaseSource + serveIssueRoles + graph
                                   roles from the same source + httpapi.Config{Graph: …}
cmd/bd/backup_restore.go           runBackupRestore → Admin.MarkUnverified (no-op without a
                                   witness) after RestoreDatabase, before the commit
```

Dependency direction, enforced: `cmd/bd → graphops, storage accessors,
bdpclient`; `internal/httpapi → graphops, bdpwire`; `internal/storage/* →
graphops, authority`; `graphops → beadserrors, stdlib`.

## 4. Roles — how many, and why

The house test: a role is a **different question**, born whole with the
methods that are shapes of one question; and *can one caller be entitled
to the read and not the write?* — if yes, two roles. Six, each behind its
own `BeadGraph*` accessor. **Authority is never a parameter:** every
protected operation is asserted by the store from the witness its accessor
loaded. **Transitions produce the witness** and therefore carry their own
preconditions instead of the head check — named per method.

- **`graphops.Reader`** — "what is in this Scope, as BDP sees it": one
  record by path (a Bead with its complete, bounded, **grouped**
  `ownedLinks` — the pinned schema keys the member by Link Type URL, so the
  Go shape is `[]OwnedLinkGroup{TypeURL, Links}`), a keyset-paged selection
  under an opaque cursor, incident Links. Protected.
- **`graphops.DescriptorReader`** — the ordered catalog and a keyed lookup.
  Protected. Bounded.
- **`graphops.TypeInstaller`** — install/converge descriptors after mint,
  keyed by fingerprint, with closure validation; every install is a ledger
  event and a fenced mutation. **P1** (the conformance fixture and `bd bdp
  types install` need it). **Before mint there is no install**: the
  built-in catalog is installed by `Mint`.
- **`graphops.IdentityReader`** — the Scope row, the witness's claim, and
  the provider's `LedgerDurability`. **Exempt** (it reports state).
- **`graphops.ScopeBootstrapper`** — `Mint`, once, fenced per hazard (A7),
  two-phase (A5): precondition *no Scope row*; INSERT the singleton Scope
  row, seed the sequence counter, append the `mint` event, install the
  built-in catalog with its `install` events — one transaction, one scoped
  commit, published where hazard R applies; then the witness is finalized.
  The only mutation `bd bdp serve` may hold, on its first-serve path.
- **`graphops.Admin`** — `Promote` (precondition: Scope row CAS + the
  fence; produces the witness), `Rotate` (precondition: Scope-row lineage
  under the exclusive gate; `refuse_url(old)` + `rotate(new)` in one
  transaction; updates tracked `config.yaml` in the same two-phase
  transition), `LedgerSnapshot`, `LedgerApply` (precondition: lineage match,
  store head == manifest predecessor, chain verified — **the head check is
  deliberately skipped**, it is what this repairs; produces the witness),
  `MarkUnverified`/`ClearUnverified` (witness-file operations). Authorized
  by being the local administrative composition root under the exclusive
  gate; reached only by the `bd bdp promote|restore|ledger|types` verbs and
  `bd backup restore`; `httpapi.GraphConfig` has no field for it.
- **`graphops.Writer`** (P3, not now) — born whole with the write-profile
  ADR (W1 upstream); per-token authorization classes and push-on-commit
  precede it.

The fixture writer (`SeedBeadInTx`/`SeedLinkInTx`) is deliberately not a
role: reachable only from `backend/conformance`'s `BeadGraphFixture` hook
(which also stands in for the witness file and the installation id), and
every call site is a `_test.go` file (source-scan test).

## 5. A read, end to end

```text
client ──GET /acme/beads/x──▶ internal/httpapi (bd serve | bd bdp serve; UOW leg)
   │ route(): deadline → bearer (before the semaphore) → Bd-Project-Id stamp
   │          (absent = pass; BDP clients never send it) → database slot
   ▼
bdp handler: path grammar (graphops laws) → ONE role call
   ▼
graphops.Reader.Bead(ctx, BeadRequest{Path})          ← no authority fields
   ▼ telemetry span (hook layer absent: taken from beneath it)
BeadGraphReader (manager-backed role): w := authority.Load(beadsDir)   ← per call;
   ▼ RunTxRead ─▶ storage/graphops.ReadBeadInTx(ctx, runner, w, req)     absent → ErrNotAuthority
   │ stmt 1  Scope row: (scope_url, authority_id, epoch) == w
   │ stmt 2  ledger head: event at w.LedgerSeq has hash w.LedgerHash; MAX(seq) >= w.LedgerSeq
   │ stmt 3  hazard S: lease row holder == w (SELECT — a read)
   │ stmt 4  graph-state version: ordered DOLT_HASHOF_TABLE() of the graph tables == w.StateVersion
   │         → else ErrStateChanged: the body returns WITHOUT validating in this transaction;
   │           the accessor validates under singleflight in its own transaction, advances the
   │           witness, and retries the read once
   │         (a clone, a restore, a copied file, a promotion elsewhere, a foreign delta: all
   │          fail HERE, in the store)
   │ stmt 5  bead row
   │ stmt 6  descriptors for the Bead's Type (cached by fingerprint → a cache hit skips it)
   │ stmt 7  ONE batched owned-links query, LIMIT (bound − rows so far) + 1
   ▼
graphops.BeadRecord {Bead, OwnedLinks []OwnedLinkGroup}  ← complete, grouped, ordered, bounded
   ▼
bdp handler: bdpwire DTO ← record; typed error → BDP Problem (bdp_problem.go); JSON out
```

Statement budgets are **per role method** (spec B2): a cold `Bead` is
seven, warm six; `Link` five; `Descriptors` five; never a per-row
statement. The witness is reloaded per call (a plain file read), so a
`bd bdp promote` in another process is honored by the next read. On hazard
R the server also fetches on the heartbeat and fails closed after the
grace. That is the v0 "lease" (A1).

A write (P3): the body asserts the witness **mutatingly** (the lease
`UPDATE` with one affected row on hazard S), runs the no-op gate, records
attribution per version, versions the source on owned-Link mutation, takes
the next `seq` from the counter, appends the event, stamps
`last_authority_id`/`last_epoch`; the accessor commits scoped, **pushes
when hazard R applies**, and only then advances the witness under the
exclusive lock — DB first, file second.

## 6. Where the twelve rulings land

| Ruling | Lands in |
| --- | --- |
| 1 charter (core; amendment after working bits) | `engdocs/PROJECT_CHARTER.md` edit rides the first merged slice |
| 2 substrate S1 | replicated migrations (scope, scope history, descriptors, beads, links, ledger events + counter, allocations) + the lease: its name in `doltIgnorePatterns`, a main-series creation migration for existing workspaces, and the ignored-series twin for fresh clones (the 0055/`ignored/0012` shape; hygiene check D) |
| 3 allocation ledger (keyed O(1)/O(log n)) | `graph_allocations` PK on the Scope-relative path (derived state); `graph_ledger_events` is the hash-chained record; `graph_ledger_seq` the counter |
| 5 withdrawal | nothing projects Issues; `graphops` imports no `internal/types` — structural |
| 7a Scope URL | `graphops` Scope-URL law; `bdp.scope_url` in tracked `config.yaml` (`BDP_SCOPE_URL` first); singleton Scope row; no dev-mode derivation in bd |
| 7b listener | BDP rows behind `httpapi.route()` — same posture by construction (A2) |
| 8 changefeed | P3: a `graphops.Changefeed` role over the graph's own log; the frozen v0 journal untouched |
| 9 authority | accessors = the normalized abstraction; store-owned witness asserted in every transaction (A1); fences composed by hazard, on the serving leg (A7); contract cases incl. a push/pull-produced clone, a `DOLT_BACKUP` restore, a copied witness, a lease takeover mid-transaction, and a non-FF push all refusing |
| 11 restore vs identity | hash-chained ledger head in the witness; two-phase transitions; `bd bdp ledger snapshot|apply` with manifests and a recovery predicate; `LedgerDurability`; `bd bdp restore` rotates unless continuity shown (A5) |
| 12 store/Scope/client | `bd init` migrations only; `bd bdp serve` mints (catalog inside the mint); `bd bdp client` / `bd init --bdp-server` → `config.local.yaml` (A6); A8 options |
| 6 wisps | not served; C-lane visibility decision recorded in the plan |

## 7. What is deliberately not designed here

- Cross-request cursor continuation (P2 ADR): `Cursor` is opaque from P1;
  **no BDP collection routes ship before the ADR**.
- The write profiles' wire (W1 upstream), `graphops.Writer`'s request
  shapes, per-token authorization classes, the push-on-commit latency
  budget, and the acknowledged-but-unwitnessed-write window.
- The replication/merge ADR (§2b): route inventory, the validator's rules,
  foreign-delta refusal, reset-to-remote for superseded clones, federation
  policy (v0 default: replicated tables travel unfiltered, by decision).
  Precedes the migrations.
- The enforcement boundary for out-of-role DML beyond the validator.
- Whether `bd bdp serve` remains after W2 as the strict alias of `bd serve`
  (default: yes — it is the only minting path).
- Type generation from the bead-type inventory (W3) — it feeds the built-in
  catalog `Mint` installs; it does not change the role.
