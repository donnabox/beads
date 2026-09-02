# BDP graph store — architecture and design

**Status:** Draft v3 (W-arch, after two council rounds) — feat/bead-graph
**Date:** 2026-09-02
**Companion:** `BDP_BEAD_GRAPH_PLAN.md` (the plan and its twelve rulings) and
`BDP_GRAPH_CLI_AND_STORAGE_SPEC.md` (the detailed CLI and storage-interface
changes). This document is the *shape*: what the pieces are, where they live,
how a request flows, and where this design corrects the plan after the
tree's own conventions were read closely. Two three-reviewer councils
(Claude, Codex, Gemini) reviewed v1 and v2; §2 records what changed and
which changes amend a ruling rather than a mechanism.

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
is configured. Every read is **one role call, one transaction**, and that
transaction asserts the authority expectation the process started with —
the replicated Scope row, plus a **clone-local authority file** that neither
clone nor pull carries and whose ledger high-water mark a database restore
cannot satisfy — so a clone, a restore, or a promotion elsewhere is refused
by the store, not by a startup check. The CLI's graph verbs live under
`bd bdp …` and reach the same roles through the same accessors, or — when a
workspace is `bd init --bdp-server`'d — speak BDP to a designated server.

## 2. Corrections to the plan, and proposed ruling amendments (read this first)

Two kinds of change are recorded here and must not be confused. A
**mechanism correction** replaces something the plan's §4 *proposed* with
the house idiom that already solves the problem; the ruling it serves is
unchanged. A **ruling amendment** changes text the operator ratified in §9;
it is *proposed* here and takes effect only when ruled. Round 1 found v1
blurring the two; round 2 found v2 restating A5 on a premise a live Dolt
probe disproved (a backup restore carries the working set, so a dolt-ignored
table survives it). v3's amendments are stated with their evidence.

### 2a. Mechanism corrections (no ruling changes)

| Plan §4 mechanism | Replaced by | Why |
| --- | --- | --- |
| `GraphCapable` as a *separate optional interface*, resolved by targeted decorator peels (`graphsource` package) | **Role accessors on `storage.Storage`** — the house rule: "a new capability gets a new role interface and a new accessor" | Every accessor lives on `Storage` (28 today); `DoltStorage` embeds it; both decorators embed `DoltStorage`. An accessor added to `Storage` *compiles* through every decorator by promotion — and promotion is exactly what the reflection census in `role_accessor_decorator_test.go` (and its telemetry twin) rejects: each decorator must **declare** the accessor. A separate interface is the shape the census cannot see. The cost is the one `storage.go` itself documents: "adding a required method is a breaking change for out-of-tree implementations; such additions are called out in CHANGELOG.md" — see amendment A8, because the plan's constraint #1 promised more than that. |
| A `ResolveGraphReadSource` policy that "re-applies telemetry" after peeling | Nothing — **the accessors carry the layers** | This is what accessors are *for* in this tree; a resolver that peels and re-wraps duplicates that machinery. |
| `backend/types.go` aliases for every `graphops` type, enforced by `TestPublicSurfaceComplete` | **No aliases.** `graphops` is a public root package, like `issueops` | The completeness guard demands aliases only for types under `internal/` (`backend/completeness_test.go`); `backend/types.go` states `issueops` is deliberately not re-exported. Ruling 9's "public in `backend/` from P1" is satisfied by the package being public and the conformance family living in `backend/conformance`. |
| `internal/graphapi` as a separate meaning-function package | **The laws live in `graphops`** as pure functions beside the values they govern | A `graphops` constructor calling `internal/graphapi` would break the leaf's import rule (stdlib + `beadserrors` only, the `memoryops` precedent) and cycle. |
| A `ReadSnapshot` handed from a resolver to a handler | **One role call, one transaction, carrying an `AuthorityExpectation`** (amendment A1 for the ruling-9 wording) | Nothing escapes a role; the authority check and the read share a snapshot. |

### 2b. Proposed ruling amendments (pending operator ruling)

| # | Ruling | v1/v2 said | v3 proposes | Evidence |
| --- | --- | --- | --- | --- |
| A1 | **9** ("the snapshot lease" in the obligation list) | dropped the lease silently (v1) | **The per-call transaction is the v0 lease and carries the authority check:** every role call takes an `AuthorityExpectation` `{ScopeURL, AuthorityID, Epoch, LedgerHighWater}` that the body asserts *inside the same transaction* as the read — descriptor reads and installs included. Cross-request continuation is P2's cursor ADR; the cursor type is opaque from P1 so P2 can bind snapshot identity into it without a public interface change. Ruling 9's obligation reads "single-transaction operations under an asserted authority expectation". | A startup-only check split from the read serves a superseded authority the moment a promotion lands elsewhere (round 1, all three). |
| A2 | **7b + 12** (listener; `bd bdp-serve`) | a sibling server (v1) | **BDP rows mount inside `internal/httpapi`** as a conditional second route table behind the same `route()`; **`bd bdp serve`** is a thin command over the same server that refuses without a Scope and inherits `errServeReadonly` wholesale (it mounts the issue routes, so serve's whole-surface refusal applies); `bd serve` mounts the rows when a Scope URL is configured and is byte-identical otherwise. | `route()`'s ordering and the host policy are private (round 1); `Capabilities()` derives from `routeTable`, so any server built from it advertises `issues.claim` (round 2, all three). |
| A3 | **12** / §4 lifecycle (`bd bead`, `bd link`, `bd graph …`) | verbs "reserved now" | **Everything under `bd bdp …`**, classified by **command path**, not leaf name: `readOnlyCommands`, `serveCmdName`, and three `cmd.Name() != "import"` branches key on the leaf, so `bd bdp` leaves avoid `import`/`export`/`serve`-by-coincidence and get an explicit CommandPath-keyed policy (spec A5). | `bd link`, `bd graph`, `bd restore`, `bd promote` exist; leaf-name policy inheritance (round 2). |
| A4 | §3 layering | values moved to `graphops` silently | **Record it:** values, laws, and roles in public `graphops`; `internal/graphapi` not created; accessors named `BeadGraph*`. | Package layout is public API. |
| A5 | **11** (ledger "restorable independently … its own migration") | v2: dolt-ignored authority table "arrives absent after a restore" | **(i) The clone-local authority half is a file, `.beads/graph-authority.local.json`** (gitignored; the resolved-port-file precedent), carrying `{authority_id, epoch, ledger_high_water, state_version}`. A `dolt clone`/`git clone`/pull **arrives without it**; a **database restore keeps it** but rewinds the ledger, so the in-transaction check `max(ledger seq) >= file.ledger_high_water` fails — structural detection, not a flag. **(ii)** the ledger is an **append-only event table** with one global sequence (allocate / tombstone / refuse-URL), exported and imported as contiguous, Scope-bound ranges (`bd bdp ledger snapshot|apply`). **(iii)** providers declare `LedgerDurability`; `bd bdp restore` rotates unless continuity is shown. Residual, stated: a whole-directory filesystem snapshot rewinds file and state together and is undetectable in-band; the ledger lane is the answer there. | Live probe (round 2): `DOLT_BACKUP` restore carries a dolt-ignored table and its uncommitted row; `dolt clone` does not. |
| A6 | §4 lifecycle (metadata.json; `BD_BDP_TOKEN`; `BDP_SCOPE_URL`) | both files; env token | **`config.yaml` is the single local source for `bdp.*`** (yaml-only, validated); nothing in metadata.json. **`BDP_SCOPE_URL` is honored** as ruling 7a spells it (`BindEnv("bdp.scope_url", "BDP_SCOPE_URL", "BD_BDP_SCOPE_URL")` — the tree's `BEADS_IDENTITY` precedent). **No env-carried token** (child-process inheritance; the `remote_migrate_gate` precedent) and no `bdp.token_file` key (every key containing `token` is a secret to `IsSecretKey`, and the tracked-config guard refuses it): token file via `BEADS_BDP_TOKEN_FILE` or the credentials file. **`bdp.client` is blocked from env** like `backend` (`blockedEnvVars`: a silent reroute hazard). | Round 1 precedence contradiction; round 2 env-spelling and secret-pattern findings. |
| A7 | **9** (promotion "explicit and epoch-rotating") — *new* | v2: "the old authority keeps serving until it pulls" | **Promotion fences at the remote, and the server watches the remote.** `bd bdp promote` performs the epoch CAS, writes the local file, and **pushes**; a non-fast-forward push means someone else promoted first and the command fails. The superseded authority cannot land writes (its push is non-fast-forward — Dolt's refusal is the fence for history), and it stops *serving* within one heartbeat: a serving process on a remote-backed workspace fetches the remote tracking ref every `bdp.authority_heartbeat` (default 30s) and refuses with `ErrNotAuthority` when the remote Scope epoch exceeds its own. Reads served inside that window are the "stale-authority reads" ruling 9 already defers to BDP's replica-labeling note. Workspaces with no remote have no second clone to promote. | Round 2 (Codex Critical): an operator warning is not fencing. |
| A8 | plan §1 constraint #1 ("zero compatibility degradation") and ruling 12 ("a provider without the capability keeps existing `bd serve` behavior") — *new* | v2 called the break "documented" | **Scope constraint #1 to behavior:** every in-tree topology and every existing workspace behaves identically; **out-of-tree `backend/` implementers take the source break** `storage.go` already declares (six one-line `ErrUnsupported` stubs), called out in CHANGELOG. The alternative — an optional interface with a decorator-aware resolver — is what round 1 rejected: the censuses cannot see it and the decorators do not promote it. Ruling 12's sentence becomes "a provider *implementing the stubs* without the capability keeps existing `bd serve` behavior". | `storage.go` contract text; `backend.Storage = storage.Storage`. |

Two decisions the plan does not yet contain, surfaced for ruling rather
than designed around:

- **Enforcement boundary for out-of-role writes.** `bd sql`, the proxied
  `RawSQLUseCase`, and a merge can change graph tables without allocation,
  authority, revision, or owned-Link coupling checks. v3's position (§7):
  out of contract; a **state-change validator** (below) rejects invalid or
  foreign-authority graph state; DB-privilege or trigger enforcement is a
  C-lane verification task. To be ruled before P3.
- **Replication/merge ADR as a P1 gate.** The merge entry points are not
  four Go functions: `CALL DOLT_PULL` merges inside Dolt on every pull
  route, the UOW leg's `DoltRemoteUseCase` calls `DOLT_MERGE`/`DOLT_PULL`
  directly, embedded federation sync fetches and merges, and the
  remote-migrate gate does a fast-forward `DOLT_MERGE`. A wrapper cannot
  see a server-side merge, so the validator runs **on every observed
  state-version change** (the local file records the `state_version` — the
  Dolt HEAD hash — it last validated; a mismatch at authority-assertion time
  runs the validator before the read is answered), with every row carrying
  `last_authority_id`/`last_epoch` provenance so foreign *updates*, not just
  foreign births, are identifiable. Prefer refusal of foreign-authority
  deltas over invented merge rules. Lands before the graph migrations.

What none of this changes: ruling 9's level — the authority is the graph
store as reached through the normalized storage abstraction, on any
provider; Dolt is the reference realization; the CLI verbs and the BDP
handler are both clients of that abstraction.

## 3. Packages and their imports

```text
graphops/                        PUBLIC LEAF (sibling of issueops/, memoryops/)
  ├─ types.go                    Bead, Link, Ref (in-Scope|external), Properties, Revision,
  │                              Attribution, TypeDescriptor, OwnedLinkDecl, OwnedLinkGroup,
  │                              ScopeIdentity, AuthorityExpectation, Cursor (opaque) — value
  │                              types with unexported fields and law-enforcing constructors
  ├─ laws.go                     pure functions: canonical-ID grammar (reject, never trim),
  │                              code-unit ordering, JSON canonicalization, RFC 6902 §4.6
  │                              equality (the no-op gate), Scope-URL validation
  ├─ reader.go                   Reader: Bead, Link, Beads, Links, IncidentLinks
  ├─ types_role.go               DescriptorReader: Descriptors, Descriptor
  │                              TypeInstaller: Install (idempotent, fingerprint-keyed)
  ├─ identity.go                 IdentityReader: Read, LedgerDurability
  │                              ScopeBootstrapper: Mint (once)
  │                              Admin: Promote, Rotate, LedgerSnapshot, LedgerApply,
  │                                     MarkUnverified, ClearUnverified
  ├─ writer.go                   (P3) Writer — born whole with the write-profile ADR
  └─ errors.go                   sentinels ALIASED from beadserrors: ErrNotFound, ErrValidation,
                                 GoneError{Path, State}, ErrNoScope, ErrScopeExists,
                                 ErrNotAuthority, ErrStateRewound, ErrURLReused,
                                 ErrRepresentationTooLarge, ErrNotServedYet
  imports: stdlib + beadserrors ONLY — no internal/types (the memoryops precedent)

internal/storage/graphops/       TX-LEVEL SHARED BODY — all three legs call it
  type DBTX interface { ExecContext; QueryContext; QueryRowContext }
    (the issueops.DBTX shape; *sql.Tx and domain/db.Runner both satisfy it)
  assertAuthorityInTx (first statement of every body: Scope row epoch/id == Expect;
    max(ledger seq) >= Expect.LedgerHighWater else ErrStateRewound)
  ReadBeadInTx / ReadLinkInTx / SelectBeadsInTx / SelectLinksInTx / IncidentLinksInTx /
  DescriptorsInTx / DescriptorInTx / InstallDescriptorInTx / ReadIdentityInTx /
  MintScopeInTx / PromoteInTx / RotateInTx / LedgerSnapshotInTx / LedgerApplyInTx /
  ValidateStateInTx (the state-change validator)
  SeedBeadInTx / SeedLinkInTx — the P1 fixture writer, ledger-enforcing; call sites
    are _test.go files only (a source-scan test enforces it)
  No exported constructor. Inside the charter's Storage Boundary. A NEW, stricter
  depguard rule (cmd/bd imports the issueops tx-body package in five files today)
  denies this package to everything except the three legs and domain/db.

internal/storage/dolt/beadgraph_*.go, internal/storage/embeddeddolt/beadgraph_*.go
  accessors wrapping the bodies in withReadTx / withRetryTx (server) and withConn
  (embedded, //go:build cgo); nil receiver → *storage.ErrUnsupported

internal/storage/domain/beadgraph.go + internal/storage/domain/db/beadgraph.go
  BeadGraphUseCase over a db.Runner, delegating to the InTx bodies — the
  MetadataCAS/TreeWalker precedent (uow.Tx → IssueUseCase() → CompareAndSetMetadataKeyInTx)
internal/storage/uow/beadgraph_*.go
  uow.UnitOfWork gains BeadGraphUseCase(); the provider gains the six accessors
  (RunTxRead for reads; RunTxResult with a commit message for install/mint/promote/
  rotate/apply — a no-op result commits nothing). The notifying provider declares
  each explicitly and the parity test covers them.

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
                                   CommandPath-keyed root store policy for the subtree
cmd/bd/bdp_serve.go                thin: serveDatabaseSource + serveIssueRoles + graph
                                   roles from the same src + httpapi.Config{Graph: …}
cmd/bd/backup_restore.go           runBackupRestore → Admin.MarkUnverified after a
                                   successful RestoreDatabase
```

Dependency direction, enforced: `cmd/bd → graphops, storage accessors,
bdpclient`; `internal/httpapi → graphops, bdpwire`; `internal/storage/* →
graphops`; `graphops → beadserrors, stdlib`. `.golangci.yml` gains the
explicit deny for `internal/storage/graphops` described above — the
`cmd-bd-role-constructors` rule matches by package import, not by
constructor symbol, so "no constructor to deny" (v1) was not a guard.

## 4. Roles — how many, and why

The house test: a role is a **different question**, born whole with the
methods that are shapes of one question; and *can one caller be entitled
to the read and not the write?* — if yes, two roles. Six, each behind its
own `BeadGraph*` accessor:

- **`graphops.Reader`** — "what is in this Scope, as BDP sees it": one
  record by path (a Bead with its complete, bounded, **grouped**
  `ownedLinks` — the pinned schema keys the member by Link Type URL, so the
  Go shape is `[]OwnedLinkGroup{TypeURL, Links}` in code-unit order, an
  owned Type with no Links present as an empty group), a keyset-paged
  selection under an opaque cursor, incident Links. Every method takes the
  `AuthorityExpectation` and asserts it in-transaction.
- **`graphops.DescriptorReader`** — "what Types does this Scope
  advertise": the ordered catalog and a keyed lookup, **under the same
  expectation** (a stale clone must not serve `types/` either). Bounded.
- **`graphops.TypeInstaller`** — install/converge descriptors, keyed by
  fingerprint, with closure validation. **P1**: `bd init`'s bootstrap and
  the conformance fixture need it. Before a Scope exists the expectation
  is empty by necessity; **once a Scope row exists, `Install` requires the
  matching expectation** like every other mutation. Refuses an owning
  declaration without a `Max`.
- **`graphops.IdentityReader`** — the Scope row, the local authority file's
  claim, and the provider's `LedgerDurability` declaration.
- **`graphops.ScopeBootstrapper`** — `Mint`, once: INSERT into the
  singleton Scope row (loses the race → `ErrScopeExists`), the first
  ledger event, and the local file. The only write the server assembly may
  hold, and only on the first-serve path.
- **`graphops.Admin`** — `Promote` (epoch CAS + local file + push),
  `Rotate` (new URL; old URL → refuse event), `LedgerSnapshot`/`LedgerApply`
  (the continuity lane), `MarkUnverified`/`ClearUnverified`. Reached only
  by the `bd bdp promote|restore|ledger` verbs and `bd backup restore` — an
  offline administrative composition root; `httpapi.GraphConfig` has no
  field for it, so a server *cannot* hold it (compile-time, tested).
- **`graphops.Writer`** (P3, not now) — born whole with the write-profile
  ADR (W1 upstream); per-token authorization classes precede it.

The fixture writer (`SeedBeadInTx`/`SeedLinkInTx`) is deliberately not a
role: reachable only from `backend/conformance`'s `BeadGraphFixture` hook,
and every call site is a `_test.go` file (source-scan test).

## 5. A read, end to end

```text
client ──GET /acme/beads/x──▶ internal/httpapi (bd serve | bd bdp serve)
   │ route(): deadline → bearer (before the semaphore) → Bd-Project-Id stamp
   │          (absent = pass; BDP clients never send it) → database slot
   ▼
bdp handler: path grammar (graphops laws) → ONE role call
   ▼
graphops.Reader.Bead(ctx, BeadRequest{Path, Expect})     Expect = the identity this
   ▼ telemetry span (hook layer absent: taken from beneath it)   process started with
BeadGraphReader → withReadTx ─▶ storage/graphops.ReadBeadInTx(ctx, tx, req)
   │ stmt 1  assertAuthorityInTx: graph_scope(url, epoch, authority_id) == Expect
   │         AND max(graph_ledger_events.seq) >= Expect.LedgerHighWater
   │         → ErrNotAuthority / ErrStateRewound (a clone, a restore, a promotion
   │           elsewhere that was pulled — all fail HERE, mid-process)
   │         [state_version changed since last validation → ValidateStateInTx first]
   │ stmt 2  bead row       stmt 3  ONE batched owned-links query, LIMIT bound+1
   │ stmt 4  catalog fingerprint (descriptor decode cached by fingerprint)
   ▼
graphops.BeadRecord {Bead, OwnedLinks []OwnedLinkGroup}  ← complete, grouped, ordered,
   ▼                                                        bounded (ErrRepresentationTooLarge)
bdp handler: bdpwire DTO ← record; typed error → BDP Problem (bdp_problem.go); JSON out
```

Four statements per single-resource read, **no per-row statements** — the
contract's query-count case pins that, not "2". The server obtains `Expect`
once at startup through `IdentityReader` and the local file (and mints
through `ScopeBootstrapper` on the first-serve path); it never caches the
*answer* — every operation re-asserts it in the store. On remote-backed
workspaces the server also fetches the remote tracking ref on the
`authority_heartbeat` and stops serving when the remote epoch is higher
(amendment A7). That is the v0 "lease" (amendment A1).

On Dolt-server workspaces `bd serve` answers from the **unit-of-work
provider**, so the UOW leg is the *primary* production path for BDP; the
`DBTX`-shaped bodies are what make it the same code.

A write (P3) follows the same path through `graphops.Writer`, whose body
asserts the expectation, runs the no-op gate, records attribution per
version, versions the source on owned-Link mutation, appends the ledger
event, stamps `last_authority_id`/`last_epoch`, and advances the local
file's `ledger_high_water` after commit — all inside the one write
transaction, on whichever provider realizes the accessor.

## 6. Where the twelve rulings land

| Ruling | Lands in |
| --- | --- |
| 1 charter (core; amendment after working bits) | `engdocs/PROJECT_CHARTER.md` edit rides the first merged slice |
| 2 substrate S1 | `internal/storage/schema/migrations/NNNN_beadgraph_*.up.sql` (replicated: scope, scope history, beads, links, descriptors, ledger events, allocations); no ignored-series table — the clone-local half is the `.beads/` file (A5) |
| 3 allocation ledger (keyed O(1)/O(log n)) | `graph_allocations` PK on the Scope-relative path (derived state); `graph_ledger_events` is the append-only record |
| 5 withdrawal | nothing projects Issues; `graphops` imports no `internal/types` — structural |
| 7a Scope URL | `graphops` Scope-URL law; `bdp.scope_url` (yaml-only; `BDP_SCOPE_URL` honored); singleton Scope row; **no dev-mode derivation in bd** — BDP's `local-test` development mode belongs to the reference server; bd tests configure a real URL |
| 7b listener | BDP rows behind `httpapi.route()` — same posture by construction (A2) |
| 8 changefeed | P3: a `graphops.Changefeed` role over the graph's own log; the frozen v0 journal untouched |
| 9 authority | accessors = the normalized abstraction; Scope row + local file + in-transaction expectation (A1); promotion CAS fenced at the remote with a heartbeat (A7); contract cases incl. a push/pull-produced clone refusing and a `DOLT_BACKUP` restore of an authority refusing |
| 11 restore vs identity | local file high-water mark (rewind detection); append-only ledger events + `bd bdp ledger snapshot|apply`; `LedgerDurability`; `bd bdp restore` rotates unless continuity shown (A5) |
| 12 store/Scope/client | `bd init` migrations + descriptor bootstrap via `TypeInstaller`; mint on first serve; `bd init --bdp-server` → config.yaml (A6); out-of-tree stubs (A8) |
| 6 wisps | not served; C-lane visibility decision recorded in the plan |

## 7. What is deliberately not designed here

- Cross-request cursor continuation (P2 ADR): the `Cursor` type is opaque
  from P1 so P2 binds snapshot identity into it without an interface break;
  **no BDP collection routes ship before the ADR** — P2's first rows are
  discovery and single-resource reads.
- The write profiles' wire (W1 upstream), `graphops.Writer`'s request
  shapes, and per-token authorization classes (read / write / admin).
- The replication/merge ADR (§2b): route inventory, the state-change
  validator's rules, foreign-delta refusal, federation policy (v0 default:
  replicated tables travel unfiltered, by decision). Precedes the migrations.
- The enforcement boundary for out-of-role DML beyond the validator.
- Whether `bd bdp serve` remains after W2 as the strict alias of `bd serve`
  (default: yes).
- Type generation from the bead-type inventory (W3) — it feeds
  `TypeInstaller`'s bootstrap catalog; it does not change the role.
