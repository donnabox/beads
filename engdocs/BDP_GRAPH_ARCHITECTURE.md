# BDP graph store — architecture and design

**Status:** Draft v4 (W-arch, after three council rounds) — feat/bead-graph
**Date:** 2026-09-02
**Companion:** `BDP_BEAD_GRAPH_PLAN.md` (the plan and its twelve rulings) and
`BDP_GRAPH_CLI_AND_STORAGE_SPEC.md` (the detailed CLI and storage-interface
changes). This document is the *shape*: what the pieces are, where they live,
how a request flows, and where this design corrects the plan after the
tree's own conventions were read closely. Three three-reviewer councils
(Claude, Codex, Gemini) reviewed v1–v3; §2 records what changed and which
changes amend a ruling rather than a mechanism. Round 3 converged on one
area — the authority mechanism — and v4 rebuilds it on two precedents the
tree already has: the ephemeral, dolt-ignored `leases` table (bd-lrgn1) and
Dolt's non-fast-forward push refusal.

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
same `route()` middleware; `bd bdp serve` is a thin command over that server
that *requires* a Scope, and `bd serve` mounts the same rows when a Scope URL
is configured. Every operation is **one role call, one transaction**, and
that transaction asserts a **store-owned authority witness** — loaded by the
accessor from a clone-local, lock-protected file that no clone, pull, or
database restore can reproduce, and checked against a hash-chained ledger
head — so a clone, a restore, a copied file, or a promotion elsewhere is
refused by the store itself, and no caller can supply the witness. The
CLI's graph verbs live under `bd bdp …` and reach the same roles through the
same accessors, or — when a workspace is wired as a client — speak BDP to
a designated server.

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
| `GraphCapable` as a *separate optional interface*, resolved by targeted decorator peels (`graphsource` package) | **Role accessors on `storage.Storage`** — the house rule: "a new capability gets a new role interface and a new accessor" — *if A8 is ruled*; otherwise the optional-interface shape returns with the costs stated under A8 | Every accessor lives on `Storage` (28 today); `DoltStorage` embeds it; both decorators embed `DoltStorage`. An accessor added to `Storage` *compiles* through every decorator by promotion — and promotion is exactly what the reflection census in `role_accessor_decorator_test.go` (and its telemetry twin) rejects: each decorator must **declare** the accessor. An *optional* interface is promoted the same way, **unwrapped**, and no census sees it — the telemetry bypass round 1 objected to. |
| A `ResolveGraphReadSource` policy that "re-applies telemetry" after peeling | Nothing — **the accessors carry the layers** | This is what accessors are *for* in this tree; a resolver that peels and re-wraps duplicates that machinery. |
| `backend/types.go` aliases for every `graphops` type, enforced by `TestPublicSurfaceComplete` | **No aliases.** `graphops` is a public root package, like `issueops` | The completeness guard demands aliases only for types under `internal/` (`backend/completeness_test.go`); `backend/types.go` states `issueops` is deliberately not re-exported. |
| `internal/graphapi` as a separate meaning-function package | **The laws live in `graphops`** as pure functions beside the values they govern | A `graphops` constructor calling `internal/graphapi` would break the leaf's import rule (stdlib + `beadserrors` only, the `memoryops` precedent) and cycle. |
| A `ReadSnapshot` handed from a resolver to a handler; a caller-supplied expectation (v2/v3) | **One role call, one transaction, asserting a store-owned witness** (amendment A1 for the ruling-9 wording) | Nothing escapes a role; the authority check and the read share a snapshot; and a witness the *caller* supplies is forgeable from the replicated rows (round 3, Codex Critical) — so the accessor loads it. |

### 2b. Proposed ruling amendments (pending operator ruling)

| # | Ruling | Earlier drafts said | v4 proposes | Evidence |
| --- | --- | --- | --- | --- |
| A1 | **9** ("the snapshot lease" in the obligation list) | dropped the lease silently (v1); caller-supplied expectation (v2–v3) | **The per-call transaction is the v0 lease and asserts a store-owned authority witness:** the accessor loads the witness from the clone-local file under a shared advisory lock and the body checks it *inside the same transaction* — Scope row identity, hash-chained ledger head (exact-prefix identity, not a scalar high-water mark), and the provider's state version. Public request types carry **no** authority fields. Cross-request continuation is P2's cursor ADR; the cursor type is opaque from P1. Ruling 9's obligation reads "single-transaction operations under a store-asserted authority witness". | Round 1: a startup-only check serves a superseded authority; round 3: a caller-supplied expectation is forgeable from replicated rows. |
| A2 | **7b + 12** (listener; `bd bdp-serve`) | a sibling server (v1) | **BDP rows mount inside `internal/httpapi`** as a conditional second route table behind the same `route()`; **`bd bdp serve`** is a thin command over the same server that refuses without a Scope and inherits `errServeReadonly` wholesale; `bd serve` mounts the rows when a Scope URL is configured *and this workspace holds the authority*, keeps the legacy surface up (rows absent, a notice) when it does not, and is byte-identical when no URL is configured. | `route()` and the host policy are private; `Capabilities()` derives from `routeTable` (round 2, all three); a tracked Scope URL reaches every clone, so a non-authority clone's `bd serve` must not refuse (round 3, Claude High). |
| A3 | **12** / §4 lifecycle (`bd bead`, `bd link`, `bd graph …`) | verbs "reserved now" | **Everything under `bd bdp …`**, and the root command's policy for that subtree is **keyed by `CommandPath()`** and authoritative at every leaf-name call site (`effectiveRootStorePolicy`, `runsPostCommandMaintenance`, `isReadOnlyCommand`, `shouldAutoPruneEventsJournal`, the `cmd.Name() != "import"` branches) — the `CheckMigrationFreeze(cmd.CommandPath()…)` precedent. Leaf names such as `list` may coincide with `readOnlyCommands`; the table, not the coincidence, governs. | `bd link`, `bd graph`, `bd restore`, `bd promote` exist; leaf-name policy inheritance (rounds 2–3). |
| A4 | §3 layering | values moved to `graphops` silently | **Record it:** values, laws, and roles in public `graphops`; `internal/graphapi` not created; accessors named `BeadGraph*`. | Package layout is public API. |
| A5 | **11** (ledger "restorable independently … its own migration") | v2: dolt-ignored table (disproved: `DOLT_BACKUP` restore carries the working set); v3: a project-id-stamped file with a scalar high-water mark (project ids are shared by every clone; a scalar mark cannot prove prefix identity) | **(i) The clone-local half is `.beads/graph-authority.local.json`**, bound to a **workspace key** (`sha256(hostname ":" realpath(.beads))`, not the project id), written only under an exclusive advisory lock with read-modify-write (monotone), fsync of file and directory, and only after the ignore entry exists in the workspace's `.beads/.gitignore` (the tree's template / `requiredPatterns` / `trackedRuntimePatterns` lists). Neither `git clone`, `dolt clone`, nor a pull carries it; a directory copy to another path or host fails the workspace key; a database restore keeps it — and then fails **(ii)**. **(ii) The ledger is an append-only, hash-chained event table** (`mint`, `install`, `promote`, `rotate`, `allocate`, `tombstone`, `refuse_url`; every v0 mutation is an event) and the witness records the head `{seq, hash}`: the in-transaction check requires the store's event at that seq to carry that hash — exact-prefix identity — so a restore to any earlier point, or a different history, is refused (`ErrStateRewound`). Snapshots (`bd bdp ledger snapshot|apply`) are manifest-bound contiguous ranges. **(iii)** providers declare `LedgerDurability`; `bd bdp restore` rotates unless continuity is shown. Residuals, stated: a whole-directory filesystem snapshot rewinds file and state together; a witness file copied to the same path on the same host *is* that workspace. | Round 2 probe; round 3 (Codex H5/H6, Claude M7/M8/L11, Gemini Critical). |
| A6 | §4 lifecycle (metadata.json; `BD_BDP_TOKEN`; `BDP_SCOPE_URL`) | both files; env token; v3: everything in `config.yaml` | **`bdp.scope_url` is a project fact and lives in tracked `config.yaml`** (yaml-only; `BDP_SCOPE_URL` honored via `BindEnv`, plus `BD_BDP_SCOPE_URL`). **`bdp.client`, `bdp.server`, `bdp.insecure_http` are per-workspace and live in `config.local.yaml`** — the tree's machine-specific override that viper merges over `config.yaml` and that stays untracked — written by `bd init --bdp-server` and by a dedicated `bd bdp client` verb (generic `bd config set` refuses them with that guidance). No env-carried token; no token key in config; `bdp.client` blocked from env like `backend`. Precedence is env (where permitted) > `config.local.yaml` > `config.yaml`. | Round 3 (Claude High): `config.yaml` is git-tracked by default, so a teammate's `bd init --bdp-server` would flip the authority's own workspace into client mode. |
| A7 | **9** (promotion "explicit and epoch-rotating") — *new* | v3: CAS + push after commit; heartbeat on epoch alone; "no remote means no second clone" | **Fencing is per topology, and every graph mutation is fenced inside its transaction:** **(1) shared SQL server** (`--server`, `--shared-server`, `--proxied-server`, `--team-server`: one database, many workspaces) — the database is the arbiter: a dolt-ignored **`graph_authority_lease`** row (holder workspace key, epoch, `expires_at`, heartbeat — the `leases`/`RunTxEphemeral` precedent, no history) is asserted inside every protected transaction; renewal is an ephemeral write; expiry is fail-closed; promotion takes an expired lease (or `--steal`, operator-confirmed) and CASes the epoch. **(2) embedded/direct Dolt with a remote** — the remote is the arbiter: `Mint` and `Promote` are one provider operation, *fetch → require local HEAD == remote-tracking HEAD → CAS + ledger event → `DOLT_COMMIT` → push*; a non-fast-forward push resets to the remote ref, removes the witness, and refuses ("pull first"). A serving process fetches on `bdp.authority_heartbeat`, compares `(authority_id, epoch)`, and fails closed after `bdp.authority_heartbeat_grace` missed fetches. P3 writes on this topology are **push-on-commit**: the response is sent only after the push lands; a refused push resets the commit and answers `ErrNotAuthority` — no acknowledged write is ever orphaned. Reads inside the heartbeat window are the "stale-authority reads" ruling 9 already defers. **(3) embedded, no remote** — one copy; the CAS alone; directory copies are operator-made second Scopes (ruling 7a: one Scope per workspace). Force-push routes (`bd dolt push --force`, `PushRemote(force)`) bypass the fence as operator acts, like `bd sql`. | Round 3 (Codex Critical 2, H3; Claude M3–M5; Gemini L4); the `leases` table (migration 0055) and `RunTxEphemeral` precedents. |
| A8 | plan §1 constraint #1 ("zero compatibility degradation") and ruling 12 — *new* | v2–v3 treated the break as settled | **Two options for the ruling; the docs do not pre-empt it.** **Option A (recommended):** scope constraint #1 to *behavior* — every in-tree topology and existing workspace is byte-identical (`bd init`'s skip notice is debug-level only); out-of-tree `backend/` implementers take the source break `storage.go` already declares (six one-line `ErrUnsupported` stubs; the CHANGELOG precedent exists — the `ReadyClaimer` entry: "any external type that *implements* it … must add the method to compile"). Under A the *compiler* finds every implementer; the censuses cover decorators and facades. **Option B:** an optional `BeadGraphCapable` interface — no source break, but embedding promotes it unwrapped, so each decorator must implement it explicitly, a capability census must be written, and every consumer needs a resolver. Ruling 12's sentence becomes "a provider *implementing the stubs* without the capability keeps existing `bd serve` behavior" under A. | `storage.go` contract text; `backend.Storage = storage.Storage`; CHANGELOG. |

Two decisions the plan does not yet contain, surfaced for ruling rather
than designed around:

- **Enforcement boundary for out-of-role writes.** `bd sql`, the proxied
  `RawSQLUseCase`, force-push, and a merge can change graph tables without
  allocation, authority, revision, or owned-Link coupling checks. v4's
  position (§7): out of contract; the **state-change validator** (below)
  rejects invalid or foreign-authority graph state; DB-privilege or trigger
  enforcement is a C-lane verification task. To be ruled before P3.
- **Replication/merge ADR as a P1 gate.** The merge entry points are not
  four Go functions: `CALL DOLT_PULL` merges inside Dolt on every pull
  route, the UOW leg's `DoltRemoteUseCase` calls `DOLT_MERGE`/`DOLT_PULL`
  directly, embedded federation sync fetches and merges, and the
  remote-migrate gate does a fast-forward `DOLT_MERGE`. A wrapper cannot
  see a server-side merge, so the validator runs **on every observed
  state-version change** (`storage.StateHasher` — Dolt's `DOLT_HASHOF` —
  recorded in the witness; a mismatch at assertion time runs the validator
  under singleflight before the operation is answered), with every row
  carrying `last_authority_id`/`last_epoch` provenance so foreign *updates*,
  not just foreign births, are identifiable. A superseded clone that pulls
  resets to the remote (its unpublished, unacknowledged events are
  discarded — push-on-commit means none were acknowledged). Prefer refusal
  of foreign-authority deltas over invented merge rules. Lands before the
  graph migrations.

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
  │                              ScopeBootstrapper: Mint (once; fenced per topology)
  │                              Admin: Promote, Rotate, LedgerSnapshot, LedgerApply,
  │                                     MarkUnverified, ClearUnverified
  ├─ writer.go                   (P3) Writer — born whole with the write-profile ADR
  └─ errors.go                   sentinels ALIASED from beadserrors: ErrNotFound, ErrValidation,
                                 GoneError{Path, State}, ErrNoScope, ErrScopeExists,
                                 ErrNotAuthority, ErrStateRewound, ErrURLReused,
                                 ErrRepresentationTooLarge, ErrNotServedYet
  NO authority fields on any request type — the witness is the store's.
  imports: stdlib + beadserrors ONLY (the memoryops precedent)

internal/storage/authority/      THE WITNESS MANAGER (clone-local half; no SQL)
  Witness{WorkspaceKey, ScopeURL, AuthorityID, Epoch, LedgerSeq, LedgerHash,
          StateVersion, Unverified, GrantedAt}
  Manager(beadsDir): Load() under shared flock (.beads/graph-authority.lock, the
  internal/lockfile primitives); Advance(fn) under exclusive flock: read → fn →
  monotone (seq never decreases) → atomicfile write → fsync dir. Refuses to write
  unless .beads/.gitignore carries the entry (EnsureGitignoreForBeadsDir). The
  file is .beads/graph-authority.local.json. Used by every leg's accessor.

internal/storage/graphops/       TX-LEVEL SHARED BODY — all three legs call it
  type DBTX interface { ExecContext; QueryContext; QueryRowContext }
    (the issueops.DBTX shape; *sql.Tx and domain/db.Runner both satisfy it)
  assertAuthorityInTx(ctx, tx, w authority.Witness) — first statements of every
    protected body: Scope row == (w.ScopeURL, w.AuthorityID, w.Epoch); ledger
    event at w.LedgerSeq has hash w.LedgerHash and MAX(seq) >= w.LedgerSeq;
    shared-server topology: graph_authority_lease holder == w.WorkspaceKey and
    unexpired; state version == w.StateVersion else ValidateStateInTx (singleflight)
  ReadBeadInTx / ReadLinkInTx / SelectBeadsInTx / SelectLinksInTx / IncidentLinksInTx /
  DescriptorsInTx / DescriptorInTx / InstallDescriptorInTx / ReadIdentityInTx /
  MintScopeInTx / PromoteInTx / RotateInTx / LedgerSnapshotInTx / LedgerApplyInTx /
  ValidateStateInTx / appendLedgerEventInTx (seq under SELECT … FOR UPDATE on graph_scope)
  SeedBeadInTx / SeedLinkInTx — the P1 fixture writer, ledger-enforcing; call sites
    are _test.go files only (a source-scan test enforces it)
  No exported constructor. Inside the charter's Storage Boundary. A NEW, stricter
  depguard rule (cmd/bd imports the issueops tx-body package in fourteen files
  today) denies this package to everything except the three legs and domain/db.

internal/storage/dolt/beadgraph_*.go, internal/storage/embeddeddolt/beadgraph_*.go
  accessors: load the witness (authority.Manager over the store's beadsDir), wrap
  the body in withReadTx / withRetryTx (server) or withConn (embedded, //go:build
  cgo); mutations DOLT_COMMIT with a named message; Mint/Promote on a remote-backed
  workspace run the fetch → compare → commit → push sequence (PublishAuthority);
  nil receiver → *storage.ErrUnsupported

internal/storage/domain/beadgraph.go + internal/storage/domain/db/beadgraph.go
  BeadGraphUseCase over a db.Runner, delegating to the InTx bodies — the
  IssueUseCase() → CompareAndSetMetadataKeyInTx precedent
internal/storage/uow/beadgraph_*.go
  uow.UnitOfWork gains BeadGraphUseCase(); the provider (built with the workspace's
  beadsDir, as bd serve builds it) gains the six accessors: RunTxRead for reads;
  RunTxResult with a commit message for install/mint/promote/rotate/apply;
  RunTxEphemeral for lease heartbeats (no history — the bd-lrgn1 form). The
  notifying provider declares each explicitly; the parity test covers them.

internal/storage/storage.go        + six BeadGraph* accessors (one line each)
internal/storage/hook_beadgraph_*.go   declared; recurse UNWRAPPED (no graph hook vocabulary)
internal/telemetry/beadgraph_*.go  every method spanned storage.op / storage.done
backend/conformance/beadgraph_*_contract.go   role contracts; RoleContractBundle fields;
                                   role_bundle_cases.go rows; wirings on all three legs

internal/httpapi/bdp_routes.go     bdpRouteTable — conditional rows behind route() (P2)
internal/httpapi/bdp_handlers.go   handler = serializer over graphops roles
internal/httpapi/bdp_problem.go    typed graph errors → BDP Problem records, here only
internal/httpapi/bdpwire/          GENERATED DTOs from the vendored, pinned bdp-v0 schema
                                   (+ schema/ with PROVENANCE: upstream commit, sha256) — P0
internal/bdpclient/                graphops.Reader/DescriptorReader over the wire
                                   (Problem → the same typed errors; errors.Is holds)

cmd/bd/bdp.go                      `bd bdp` root; subcommands in cmd/bd/bdp_*.go;
                                   bdpRootPolicy keyed by CommandPath()
cmd/bd/bdp_serve.go                thin: serveDatabaseSource + serveIssueRoles + graph
                                   roles from the same src + httpapi.Config{Graph: …}
cmd/bd/backup_restore.go           runBackupRestore → Admin.MarkUnverified after a
                                   successful RestoreDatabase
```

Dependency direction, enforced: `cmd/bd → graphops, storage accessors,
bdpclient`; `internal/httpapi → graphops, bdpwire`; `internal/storage/* →
graphops, authority`; `graphops → beadserrors, stdlib`. `.golangci.yml`
gains the explicit deny for `internal/storage/graphops` — the
`cmd-bd-role-constructors` rule matches by package import, not by
constructor symbol.

## 4. Roles — how many, and why

The house test: a role is a **different question**, born whole with the
methods that are shapes of one question; and *can one caller be entitled
to the read and not the write?* — if yes, two roles. Six, each behind its
own `BeadGraph*` accessor. **Authority is never a parameter:** every
protected operation is asserted by the store from the witness its accessor
loaded; the exemptions are named.

- **`graphops.Reader`** — "what is in this Scope, as BDP sees it": one
  record by path (a Bead with its complete, bounded, **grouped**
  `ownedLinks` — the pinned schema keys the member by Link Type URL, so the
  Go shape is `[]OwnedLinkGroup{TypeURL, Links}` in code-unit order, an
  owned Type with no Links present as an empty group), a keyset-paged
  selection under an opaque cursor, incident Links. Protected.
- **`graphops.DescriptorReader`** — "what Types does this Scope
  advertise": the ordered catalog and a keyed lookup. Protected (a stale
  clone must not serve `types/` either). Bounded.
- **`graphops.TypeInstaller`** — install/converge descriptors, keyed by
  fingerprint, with closure validation; every install is a ledger event.
  **P1**: `bd init`'s bootstrap and the conformance fixture need it. Before
  a Scope exists nothing can be asserted; once a Scope row exists the
  install is protected like every mutation — so `bd init` bootstraps only
  on an unminted store or on the authority workspace, and skips elsewhere.
  Refuses an owning declaration without a `Max`.
- **`graphops.IdentityReader`** — the Scope row, the witness's claim, and
  the provider's `LedgerDurability`. **Exempt** (it reports state; `bd bdp
  status` and the serve gate consult it).
- **`graphops.ScopeBootstrapper`** — `Mint`, once, **fenced per topology**
  (A7): INSERT into the singleton Scope row + the `mint` ledger event + the
  witness; on a remote-backed workspace it is the fetch → compare → commit
  → push sequence, and a non-fast-forward push undoes it. The only mutation
  the server assembly may hold, and only on the first-serve path.
- **`graphops.Admin`** — `Promote` (fenced like `Mint`), `Rotate` (new URL;
  `refuse_url` event), `LedgerSnapshot`/`LedgerApply` (the continuity lane;
  `Apply` validates the manifest's lineage and contiguity),
  `MarkUnverified`/`ClearUnverified` (witness-file operations). **Authorized
  by being the local administrative composition root** — a process with
  write access to the workspace under the exclusive lock — never by a
  network caller; reached only by the `bd bdp promote|restore|ledger` verbs
  and `bd backup restore`; `httpapi.GraphConfig` has no field for it
  (compile-time, tested).
- **`graphops.Writer`** (P3, not now) — born whole with the write-profile
  ADR (W1 upstream); per-token authorization classes and push-on-commit
  precede it.

The fixture writer (`SeedBeadInTx`/`SeedLinkInTx`) is deliberately not a
role: reachable only from `backend/conformance`'s `BeadGraphFixture` hook
(which also stands in for the witness file), and every call site is a
`_test.go` file (source-scan test).

## 5. A read, end to end

```text
client ──GET /acme/beads/x──▶ internal/httpapi (bd serve | bd bdp serve)
   │ route(): deadline → bearer (before the semaphore) → Bd-Project-Id stamp
   │          (absent = pass; BDP clients never send it) → database slot
   ▼
bdp handler: path grammar (graphops laws) → ONE role call
   ▼
graphops.Reader.Bead(ctx, BeadRequest{Path})          ← no authority fields
   ▼ telemetry span (hook layer absent: taken from beneath it)
BeadGraphReader (accessor): w := authority.Load(beadsDir)   ← shared flock; absent → ErrNotAuthority
   ▼ withReadTx ─▶ storage/graphops.ReadBeadInTx(ctx, tx, w, req)
   │ stmt 1  Scope row + state version: (scope_url, authority_id, epoch) == w;
   │         StateHasher == w.StateVersion else ValidateStateInTx (singleflight)
   │ stmt 2  ledger head: event at w.LedgerSeq has hash w.LedgerHash; MAX(seq) >= w.LedgerSeq
   │         [shared-server topology: + graph_authority_lease holder == w.WorkspaceKey, unexpired]
   │         → ErrNotAuthority / ErrStateRewound — a clone, a restore, a copied file, a
   │           promotion elsewhere: all fail HERE, in the store
   │ stmt 3  bead row
   │ stmt 4  descriptor(s) for the Bead's Type (owned-Type declarations; cached by fingerprint)
   │ stmt 5  ONE batched owned-links query, LIMIT (bound − rows so far) + 1
   ▼
graphops.BeadRecord {Bead, OwnedLinks []OwnedLinkGroup}  ← complete, grouped, ordered,
   ▼                                                        bounded (ErrRepresentationTooLarge)
bdp handler: bdpwire DTO ← record; typed error → BDP Problem (bdp_problem.go); JSON out
```

Five statements per read, single-resource or page, **no per-row
statements** — the contract's statement-count case pins that. The witness
is reloaded per operation under the shared lock (cheap: one small file), so
a `bd bdp promote` in another process is honored by the next read. On a
remote-backed workspace the server also fetches on the heartbeat and fails
closed after the grace (A7). That is the v0 "lease" (A1).

On Dolt-server workspaces `bd serve` answers from the **unit-of-work
provider**, so the UOW leg is the *primary* production path for BDP; the
`DBTX`-shaped bodies are what make it the same code, and the provider is
built with the workspace's `beadsDir`, as `bd serve` builds it today, so
the same witness manager serves that leg.

A write (P3) follows the same path through `graphops.Writer`: the body
asserts the witness, runs the no-op gate, records attribution per version,
versions the source on owned-Link mutation, appends the ledger event (seq
allocated under the Scope row's lock), stamps `last_authority_id`/
`last_epoch`; the accessor commits (`DOLT_COMMIT`), **pushes when the
topology requires it**, and only then advances the witness (seq, hash,
state version) under the exclusive lock — DB first, file second, so the
file is never ahead of durable state.

## 6. Where the twelve rulings land

| Ruling | Lands in |
| --- | --- |
| 1 charter (core; amendment after working bits) | `engdocs/PROJECT_CHARTER.md` edit rides the first merged slice |
| 2 substrate S1 | `internal/storage/schema/migrations/NNNN_beadgraph_*.up.sql` (replicated: scope, scope history, descriptors, beads, links, ledger events, allocations) + `migrations/ignored/NNNN_beadgraph_authority_lease.up.sql` (the shared-server lease; the 0055/`leases` precedent); the clone-local witness is the `.beads/` file (A5) |
| 3 allocation ledger (keyed O(1)/O(log n)) | `graph_allocations` PK on the Scope-relative path (derived state); `graph_ledger_events` is the hash-chained record |
| 5 withdrawal | nothing projects Issues; `graphops` imports no `internal/types` — structural |
| 7a Scope URL | `graphops` Scope-URL law; `bdp.scope_url` in tracked `config.yaml` (`BDP_SCOPE_URL` honored); singleton Scope row; **no dev-mode derivation in bd** (BDP's `local-test` mode belongs to the reference server; bd tests configure a real URL) |
| 7b listener | BDP rows behind `httpapi.route()` — same posture by construction (A2) |
| 8 changefeed | P3: a `graphops.Changefeed` role over the graph's own log; the frozen v0 journal untouched |
| 9 authority | accessors = the normalized abstraction; store-owned witness asserted in every transaction (A1); fencing per topology (A7); contract cases incl. a push/pull-produced clone, a `DOLT_BACKUP` restore, and a copied witness file all refusing |
| 11 restore vs identity | hash-chained ledger head in the witness (exact-prefix rewind detection); `bd bdp ledger snapshot|apply` with manifests; `LedgerDurability`; `bd bdp restore` rotates unless continuity shown (A5) |
| 12 store/Scope/client | `bd init` migrations + descriptor bootstrap (unminted or authority only); mint on first serve, fenced; `bd bdp client` / `bd init --bdp-server` → `config.local.yaml` (A6); A8 options |
| 6 wisps | not served; C-lane visibility decision recorded in the plan |

## 7. What is deliberately not designed here

- Cross-request cursor continuation (P2 ADR): `Cursor` is opaque from P1;
  **no BDP collection routes ship before the ADR** — P2's first rows are
  discovery and single-resource reads.
- The write profiles' wire (W1 upstream), `graphops.Writer`'s request
  shapes, per-token authorization classes, and the push-on-commit latency
  budget.
- The replication/merge ADR (§2b): route inventory, the validator's rules,
  foreign-delta refusal, reset-to-remote for superseded clones, federation
  policy (v0 default: replicated tables travel unfiltered, by decision).
  Precedes the migrations.
- The enforcement boundary for out-of-role DML beyond the validator.
- Whether `bd bdp serve` remains after W2 as the strict alias of `bd serve`
  (default: yes).
- Type generation from the bead-type inventory (W3) — it feeds
  `TypeInstaller`'s bootstrap catalog; it does not change the role.
