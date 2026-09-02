# BDP graph store — CLI and storage-interface changes, in detail

**Status:** Draft v1 (W-arch) — feat/bead-graph
**Date:** 2026-09-02
**Companions:** `BDP_BEAD_GRAPH_PLAN.md` (rulings), `BDP_GRAPH_ARCHITECTURE.md`
(shape). This document is the *diff*: every command, flag, config key,
interface member, package, migration, and gate the graph work adds or
touches — and, just as precisely, what it does not touch.

## Part A — CLI

### A1. `bd init` — graph store initialization (ruling 12)

`bd init` initializes the graph store with everything else it initializes,
against the normalized storage interfaces (any provider):

- Runs the graph migrations in the normal series (Part B4).
- Bootstraps the built-in Type Descriptor catalog (W3 supplies the
  inventory; until then, an empty catalog is honest).
- Writes **no Scope identity** — a workspace has a graph store from init but
  not a Scope; the Scope is minted on first serve (A3).

No new flag is required for this. Registered backends that provision their
own workspaces (`backend/` doc: "bd init does not provision registered
backends") run their own graph bootstrap through the same
`graphops`-contract obligations, proven by the conformance family (B7).

### A2. `bd init --bdp-server <url>` — the client reroute (ruling 12, third command)

One more `bd init` target, beside `--server`, `--shared-server`,
`--proxied-server`, `--team-server`, and `--backend`. Every existing target
selects the provider or topology that realizes the storage interfaces — a
choice *below* the normalized abstraction. This one reroutes *above* it, at
the CLI: the graph verbs (A5) become a BDP client of the designated server
instead of opening a store.

| Surface | Spelling | Notes |
| --- | --- | --- |
| init flag | `--bdp-server <url>` | validated as an absolute http(s) URL |
| config.yaml | `bdp.client: store \| server` (default `store`), `bdp.server: <url>` | `bd config set bdp.server https://…` sets both (client → server) |
| metadata.json | `bdp_client`, `bdp_server` | written by init like `dolt_mode`; read by every command that forks on it |
| env | `BD_BDP_SERVER` (overrides `bdp.server`; implies `client: server`), `BD_BDP_TOKEN`, `BD_BDP_TOKEN_FILE` | token precedence: `BD_BDP_TOKEN` > token file > credentials-file section keyed by the server host (the existing `[host:port]` password pattern) |
| precedence | flag > env > metadata.json > config.yaml | identical to `dolt.mode` |

`bdp.client` is an explicit mode, never inferred from the presence of a
URL. Issue verbs are unaffected in either mode. In `server` mode the graph
verbs' `openGraph*()` accessors return a BDP-client realization of the
same `graphops` role interfaces (A5), so a verb cannot tell which route it
took — the memoryops "the accessor IS the API on both seams" rule.

### A3. `bd bdp-serve` — the isolated bootstrap server (ruling 12; W2 folds it into `bd serve` later)

```text
bd bdp-serve [--addr IP:PORT]            default 127.0.0.1:0 (ephemeral), numeric IP only
             [--allow-non-loopback]      requires --auth-token-file (or --insecure-no-auth)
             [--auth-token-file PATH]    bearer tokens, one per line, hot-reloaded
                                         (env BD_BDP_SERVE_TOKEN_FILE — Part D)
             [--insecure-no-auth]
             [--allowed-host NAME]...    exact Host allowlist entries
             [--scope-url URL]           overrides bdp.scope_url for this run (not persisted)
             [--dev-local-test]          permit a derived `local-test` Scope URL (never persisted)
```

Every flag mirrors `bd serve`'s posture and reuses the same exported
validators (`httpapi.ValidateBindAddr`, `ValidateAuthPosture`,
`ValidateAllowedHost`, `NewTokenFileAuth`). Behavior, in order:

1. **Storage classification is serve's, verbatim:** registered backend →
   its store source; embedded Dolt → **refused, permanently**, with the
   same typed `*storage.ErrUnsupported{Op: "bdp-serve", Backend:
   "embedded-dolt"}` (its commit protocol cannot meet the per-request
   atomicity contract); otherwise the Dolt provider path.
2. **Roles are taken from beneath the hook layer** — the one concrete peel
   `bd serve` does — so no hook script fires per request; telemetry stays.
3. **Scope identity:** requires `bdp.scope_url` (config/metadata) or
   `--scope-url`, validated by `graphapi.ValidateScopeURL` (BDP's startup
   contract: absolute, dialable, no reserved `local-test` segment unless
   `--dev-local-test`). Absent → refuse to start (exit 2, mirroring BDP's
   production mode). On **first serve** it mints the identity row (Scope
   URL, random authority id, epoch) through `graphops.Identity` (B2) and
   serves the Scope honestly empty. On later serves it reads the row and
   **refuses** if the configured URL does not match the persisted one, or
   if this store instance is not the marked authority (ruling 9) —
   promotion is `bd graph promote` (A4), explicit and epoch-rotating.
4. Listens via `internal/bdpapi` (Part B8). Liveness probe `GET /healthz`
   as in serve; readiness = a real discovery read.

`--readonly` refuses to start for the same reason `bd serve` does once
writes exist (P3); until then a read-only graph server is honest and
allowed — stated so the refusal arrives with the writes, not before.

### A4. `bd graph restore` / `bd graph promote` (rulings 11, 9)

- `bd graph restore` runs after a database restore: if the allocation
  ledger is intact and declared restore-surviving by the provider, it
  re-validates identity and continues; otherwise it **rotates** the Scope
  URL (operator supplies the new one) and epoch and records the old URL as
  refused-forever. Never silent.
- `bd graph promote` makes this store instance the authority for its Scope
  (a clone taking over): epoch rotates; URL rotates if history diverged
  (operator-confirmed); the old authority's marker is superseded.

### A5. Graph verbs (P3 — names reserved now, semantics with the write ADR)

```text
bd bead get|list|create|update|delete
bd link get|list|create|update|delete
bd graph types|restore|promote
```

Each verb reaches its role through an accessor and nothing else, on
whichever route the workspace is on — the `cmd/bd/label.go` pattern, plus
the third route:

```go
func openGraphReader() (graphops.Reader, error) {
    if bdpClientMode() == "server" {        // metadata/config/env, precedence per A2
        return bdpclient.GraphReader(bdpServerConfig())   // internal/bdpclient
    }
    if usesProxiedServer() { return proxiedGraphReader() }
    return store.GraphReader()
}
```

`internal/bdpclient` speaks the pinned wire (generated DTOs, B8) and
implements `graphops.Reader` over it; a BDP Problem maps back to the same
typed graph errors the store bodies raise, so `errors.Is` holds across the
route fork. `.golangci.yml` gains `cmd-bd-role-constructors` entries only
if a constructor exists to deny (the tx-level body has none).

## Part B — Storage interfaces

### B1. Accessors on `storage.Storage` (the house rule — architecture §2)

```go
// GraphReader returns the guarded graph-read surface for this store: Bead
// and Link records (a Bead with its complete ownedLinks, assembled in the
// call's one transaction), collection selections in canonical-URI order,
// and incident Links. Reads fire no hooks; the hook decorator recurses.
GraphReader() (graphops.Reader, error)

// GraphTypes returns the Type Descriptor inventory role.
GraphTypes() (graphops.Types, error)

// GraphIdentity returns the Scope-identity pair: the read half every
// serve consults, and the one-time write half only bd bdp-serve's first
// serve and bd graph promote/restore may reach (Bootstrapper/InitVerifier
// split — a caller entitled to read is not thereby entitled to mint).
GraphIdentity() (graphops.Identity, error)

// P3: GraphWriter() (graphops.Writer, error)
```

Added to the interface the decorators embed (`storage.DoltStorage`;
confirm at implementation which of `Storage`/`DoltStorage` is the embedded
one and place the accessors there). **This is a breaking change for
out-of-tree backends**, called out in `CHANGELOG.md` exactly as
`Memories()` was; `backend/` is EXPERIMENTAL and this is its documented
change class.

### B2. The `graphops` leaf (public, repo root)

```go
package graphops   // imports: stdlib, beadserrors — nothing else

type Reader interface {
    Bead(ctx, BeadRequest) (BeadRecord, error)        // BeadRecord = Bead + OwnedLinks (complete, ordered)
    Link(ctx, LinkRequest) (Link, error)
    Beads(ctx, SelectRequest) (BeadPage, error)       // canonical-URI order; bounded; single-request snapshot
    Links(ctx, SelectRequest) (LinkPage, error)
    IncidentLinks(ctx, IncidentRequest) (LinkPage, error)
}
type Types interface {
    Descriptors(ctx) ([]TypeDescriptor, error)
    // P3: Install(ctx, InstallRequest) (InstallResult, error) — closure validation, fingerprint retention
}
type Identity interface {
    Read(ctx) (ScopeIdentity, error)                  // ErrNoScope when never minted
    Mint(ctx, MintRequest) (ScopeIdentity, error)     // once; ErrScopeExists after
    Rotate(ctx, RotateRequest) (ScopeIdentity, error) // promote/restore paths only
}
```

Value types (`Bead`, `Link`, `Ref`, `Properties`, `Revision`,
`Attribution`, `TypeDescriptor`, `OwnedLinkDecl`, `ScopeIdentity`) live
here with unexported fields and constructors that enforce the laws
(`graphapi` supplies the checks); `Properties` is the immutable raw-JSON
object value from the plan (duplicate-key rejection, number preservation,
deterministic encoding, RFC 6902 §4.6 equality). Errors: aliases of
`beadserrors` sentinels — `ErrNotFound`, `ErrValidation`, plus new
`ErrNoScope`, `ErrScopeExists`, `ErrNotAuthority`, `ErrURLReused` declared
in `beadserrors` (stdlib-only) and aliased here, so `errors.Is` crosses
the module boundary.

### B3. Bodies and legs

| Leg | File | Body |
| --- | --- | --- |
| server Dolt | `internal/storage/dolt/graph_reader.go` (+ `graph_types.go`, `graph_identity.go`) | accessor wraps `internal/storage/graphops.*InTx` in `withReadTx` (writes: `withRetryTx`) |
| embedded Dolt | `internal/storage/embeddeddolt/graph_reader.go` … (`//go:build cgo`) | same body, `withConn` |
| unit of work | `internal/storage/uow/graph_reader.go` … (`GraphReaderSource`, `GraphTypesSource`, `GraphIdentitySource`) | reaches the same `InTx` body **if** every function it calls takes an interface `domain/db.Runner` satisfies — the mechanical test from `ADDING_AN_ISSUEOPS_ROLE.md`; otherwise its own body, and the contract header says which |

`internal/storage/graphops` (tx-level, no exported constructor):
`ReadBeadInTx`, `ReadLinkInTx`, `SelectBeadsInTx`, `SelectLinksInTx`,
`IncidentLinksInTx`, `DescriptorsInTx`, `ReadIdentityInTx`,
`MintIdentityInTx`, `RotateIdentityInTx`; P3 adds the write bodies with
the no-op gate, owned-Link source versioning, and ledger consultation.
**One transaction per role call** is the contract; the ownedLinks
assembly and the descriptor lookup happen inside it.

### B4. Schema (migrations, normal series, existing version gate)

| Table | Columns (essentials) | Keys / indexes |
| --- | --- | --- |
| `graph_beads` | `url` (canonical, PK), `path` (local id), `type_url`, `revision`, `attribution_principal`, `attribution_status`, `properties` (JSON), `created_at`, `updated_at` | PK `url`; index `(type_url, url)` for typed selections in canonical order |
| `graph_links` | `url` PK, `path`, `type_url`, `revision`, `source_uri`, `source_pin`, `target_uri`, `target_pin`, `attribution_*`, `properties`, timestamps | indexes `(source_uri, type_url, url)` and `(target_uri, url)` — ownedLinks assembly and incident reads are keyed scans; **no** uniqueness on (type, source, target) — multiplicity is law |
| `graph_type_descriptors` | `url` PK, `descriptor` (JSON), `fingerprint`, `installed_at` | — |
| `graph_allocations` | `url` PK, `kind` (bead/link), `birth_authority_id`, `epoch`, `allocated_at`, `tombstoned_at` | PK is the O(1)/O(log n) ID test (ruling 3); append-only; its own migration so a restore can carry it independently (ruling 11) |
| `graph_identity` | `scope_url` PK, `authority_id`, `epoch`, `minted_at`, `superseded_at` | one live row per store |

`updated_at` here is protocol-irrelevant bookkeeping; revisions are
minted by the write body (P3), never derived from timestamps.

### B5. Decorators

- `internal/storage/hook_graph_reader.go`, `hook_graph_types.go`,
  `hook_graph_identity.go`: declared, recurse **unwrapped** — the hook
  vocabulary (`on_create/on_update/on_close`) hands scripts an *issue*; a
  graph hook vocabulary is a separate proposal (the Sweeper/Memories
  argument). Added to `role_accessor_decorator_test.go`'s wrapped/unwrapped
  table and to the storage file's pass-through paragraph.
- `internal/telemetry/graph_*.go`: every method spanned with `storage.op`
  / `storage.done`.
- `roleAccessorNamesOf` gains the `graphops` facade package in its
  classification map — otherwise the census reports the new accessors as
  *unclassified* and fails, which is the guard working.

### B6. `backend/` public surface

`backend/types.go` gains aliases for every `graphops` type reachable from
the accessors (`TestPublicSurfaceComplete` fails without them — that is
the mechanism, not a checklist). `backend/backend.go`'s "minimal external
backend" example gains the three accessors. Public from P1 (ruling 9).

### B7. Conformance

- `backend/conformance/graph_reader_contract.go`, `graph_types_contract.go`,
  `graph_identity_contract.go`: cases asserting what the leaf docs promise,
  by line — ownedLinks completeness/order/bound and declared-empty
  entries, canonical-URI ordering, not-found sentinels, identity mint-once
  / rotate semantics, authority refusal, ledger non-reuse (rulings 3/9/11).
  Vote count stated at the top (two shared-body legs + UOW).
- `RoleContractBundle` gains `GraphReader`, `GraphTypes`, `GraphIdentity`
  factory fields; wirings in `internal/storage/{dolt,embeddeddolt,uow}/
  graph_*_contract_test.go`.
- Both coverage gates apply (`TestEveryRoleMethodHasAContractCase`,
  `unwiredContractEntrypoints`); waivers only with a reason and only until
  wired.
- Non-capable stores answer `*storage.ErrUnsupported{Op: "GraphReader"}`
  and prove it with `RunUnsupportedContract` (allowlist entry) — the
  honest-absence half of ruling 12.

### B8. `internal/bdpapi` and DTO generation

- Wire DTOs generated from the **pinned** `bdp-v0.schema.json` by a new
  `make bdp-gen` (JSON Schema → Go), with a `bdp-check` drift gate in the
  PR policy job beside `api-check`. BDP's wire is defined by that schema,
  not by `openapi.v0.yaml`; the two generators stay separate (W2 decides
  whether/how they converge).
- Handler = serializer: `graphops` records → DTOs; typed graph errors →
  BDP Problem records (the closed vocabulary) here and only here.
- Posture reused from `httpapi` exports: bearer auth, bind validation,
  Host allowlist, request deadline, semaphore.

## Part C — What does not change

`storage.Storage`'s existing 28 accessors and every `issueops` role; the
journal's frozen vocabulary; sync, merge settlement (graph settlement is a
*new* always-run pass over B4's tables only), federation filtering (graph
tables get their own hook), `httpapi`'s route table and OpenAPI-first rule
(until W2); `bd serve` (until W2); every legacy CLI verb; JSONL export
shapes.

## Part D — Open implementation questions (not rulings)

1. Whether the UOW leg reaches the shared `InTx` bodies through the domain
   repository or needs its own body — answered per method by the Runner
   test at implementation time and recorded in each contract header.
2. Whether `graphops.Types.Install` (P3) is a write role or a
   `Bootstrapper`-style one-time role — depends on W3's catalog lifecycle.
3. Env-var names: `BD_BDP_SERVE_TOKEN_FILE` (server) and `BD_BDP_TOKEN` /
   `BD_BDP_TOKEN_FILE` (client) proposed; and whether the credentials file
   gains a `[bdp host]` section or reuses `[host:port]`.
