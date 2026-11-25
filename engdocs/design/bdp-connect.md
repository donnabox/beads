# `bd connect`: native vs BDP backend selection (Track A design)

Status: **DRAFT / WIP** — Track A of the BDP 0.2.3-rc2 phase. Nothing here is
wired yet; nothing merges without Donna. Validation target is the rc2 clean-room
server `bdp-cleanroom-v023 @ 5fba7d23`, **not** `bd serve` (which has no BDP
surface yet — that interop is a Track B dependency).

## The problem Britt framed

`bd connect` has two separable concerns, and conflating them is the hazard:

1. **RemoteScope conformance** — proving the target actually speaks rc2 BDP
   (protocol version, the mandatory discovery floor). A pure, side-effect-free
   check.
2. **Workspace attachment** — persisting the target and *selecting the backend
   for every later command*. This is what makes `bd <cmd>` in the workspace
   route to one store or another.

A **handshake-only** change (do the BDP check, but persist the status quo) leaves
`metadata.json backend="http"` → later commands still dial native `/v0/beads`.
So the design must reify the native-vs-BDP choice end-to-end, not just at connect.

## Current architecture (what the choice has to thread through)

- **`bd connect [url]`** (`cmd/bd/connect.go`): validates the URL, runs the
  native `httpclient.Handshake` (api_version + wire revision + project id), then
  persists via `bdhttp.Attach` (`backend/http/attach.go:56`): writes the sidecar
  `.beads/http_target.json` (`httpclient.Target`) **and** sets
  `metadata.json backend="http"`.
- **Backend selection for later commands**: `openDoltStoreFromConfig`
  (`cmd/bd/store_factory.go:225`) branches on
  `backends.Lookup(cfg.GetBackend())` — the **`backend` name in metadata.json**
  is the whole switch. `"http"` is registered (`cmd/bd/http_backend.go:12` →
  `httpclient.Register`) with `Open = NewFromConfig`, which reads the sidecar and
  dials a `wire.Client`. The read-only twin is `store_factory.go:380`.
- **The `/v0/beads` assumption is compiled in**: `internal/httpclient/wire/paths.go`
  (`/v0/beads/...`) + `wire.APIVersion="v0"`, joined onto the base URL at
  `wire/client.go:446`. Not configurable — the `"http"` backend *is* native.
- **Interface impedance**: native backends satisfy the fat, issue-centric
  `storage.DoltStorage` (`internal/storage/storage.go:787`, ~16 sub-interfaces:
  `CreateIssue`, `ClaimIssue`, `ListReadyWork`, dependency/version-control ops…).
  The BDP client `bdp.Scope` (`internal/bdp/scope.go`) is a narrow, generic REST
  contract (get/list/create/put/patch/delete/… over `ResourceID`/`Representation`).
  **Not a drop-in.** cmd/bd never calls the store directly — it uses issueops
  **role accessors** (`store.IssueReader()`, `store.Counter()`, …).

## Design

### 1. How connect decides native vs BDP

**Recommend: an explicit `--protocol bdp` flag for v0** (default = native), with
**autodetect as a documented follow-up**. Rationale:

- Explicit is deterministic and testable; the operator connecting to a BDP server
  knows it. No ambiguous probing, no double-handshake latency/failure modes.
- Autodetect is *spec-sanctioned* and can come later: a BDP server answers
  `GET /` with a `ServiceDescriptor` and serves `/sys/capabilities`; a native
  server does not. So connect could probe `GET /`, and if it parses as a BDP
  `ServiceDescriptor` with a compatible `version`, treat it as BDP, else fall
  back to the native handshake. Defer until there's a UX need.

### 2. The conformance handshake (SEPARATED from attachment)

A new `bdpclient` entry point — `Discover(ctx) (ServiceDescriptor, Capabilities, error)`
(or `RemoteScope.Handshake`) — that:

- `GET /` → `ServiceDescriptor`; negotiate `version` by `major.minor` (we speak
  0.2.x; refuse an incompatible major/minor).
- `GET /sys/capabilities` → `Capabilities`; confirm the mandatory floor (`get`
  present, `/sys/capabilities` readable).
- Returns the negotiated descriptor + capabilities; **writes nothing**. Mirrors
  the native `httpclient.Handshake` discipline: nothing persists unless it passes.

This is the "RemoteScope conformance" half — independently runnable and testable
against `cleanroom-v023`, with no workspace side effects. It is also the first
real consumer of the **live `/sys/capabilities` fetch** (today `RemoteScope.Capabilities`
returns a static advert) — so that opaque item comes in now, because connect needs it.

### 3. What connect persists (attachment)

Symmetric to the native two-file write, with a **BDP discriminator**:

- `metadata.json`: **`backend: "bdp"`** — a new registered backend name. This is
  the one bit the selection switch keys on.
- A sidecar (recommend a parallel **`.beads/bdp_target.json`**, not reusing
  `http_target.json`, because the fields differ): `url`, `expect_project_id`, the
  **negotiated protocol version** (e.g. `"0.2.3"`), `ca_file`, `allow_plaintext`.

### 4. The selection branch

Register a **`"bdp"` backend** in `internal/storage/backends` (parallel to
`httpclient.Register`), wired from a `cmd/bd` `init()` like `http_backend.go:12`.
Its `Open` reads the sidecar and constructs the Scope→Store adapter (§5) over
`bdpclient.RemoteScope`. The native-vs-BDP branch then **falls out for free** at
`store_factory.go:225` / `:380` (`backends.Lookup("bdp")`). Registration is
mandatory: without it, `GetBackend()` silently falls back to `"dolt"`
(`configfile.go:333`) and mis-routes.

### 5. The Scope→Store adapter (the hard part)

`bdp.Scope` must present to cmd/bd as a store. **Do not target the full
`DoltStorage`** — target the **issueops role accessors cmd/bd actually consumes**
(`IssueReader`, `Lifecycle`, `Counter`, …), and generate a refusing shell
(`unsupported*`) for the Dolt-specific surface (version-control, compaction, GC —
no BDP analog), exactly as the native http client does today.

- CRUD slice first: role reads/writes → `Scope.Get/List/Create/Put/Patch/Delete`
  over generic `ResourceID`/`Representation`, decoding the issue body from/into
  the bead `body`.
- Map the error models at the boundary: BDP `FaultError`/`Problem` → beads
  sentinels / `*ErrUnsupported`.
- **Lifecycle & dependency ops (claim/close/heartbeat, add-dependency, trees)**
  have no generic Scope verb — they map to BDP `call` (operations) or to
  body-field patches. **This is the open rc2 question (below); deferred until
  ruled.** Until then the adapter refuses them with `ErrUnsupported`.

## Validation plan (against `cleanroom-v023 @ 5fba7d23`)

The clean-room serves the rc2 surface (`route.rs`: `GET /` → ServiceDescriptor,
`GET /sys/capabilities`, action-fence `POST /-/create`, `GET /-/list`,
`GET /{id}`, …). Known **client-wire alignment** work before green:

- **Create path**: `RemoteScope` does `POST /` today; rc2/clean-room use
  **`POST /-/create`** (bare `GET /` is now the ServiceDescriptor). Fix in the
  wire layer.
- **`/sys/capabilities` wire shape**: confirm the clean-room serializes the rc2
  nested `{core, scope, packs}` object vs the spike's `Vec<PrimitiveCapabilities>`
  array; absorb any divergence in the `bdpclient/wire.go` DTO layer, keeping the
  contract spec-faithful.
- Re-point the existing smoke test at a `cleanroom-v023` server and extend it to
  the discovery handshake.

## Open question routed to Lydia (rc2 ambiguity)

**How do beads domain lifecycle/dependency operations map onto BDP Core?** The
CRUD verbs are unambiguous, but *claim / close / heartbeat / add-dependency /
dependency-tree* have no generic Scope primitive. Candidates: (a) BDP `call` to
intrinsic operations addressed by element path — core.lumen §call hints
`issue/claim` as exactly such a path; (b) `patch` of body status/field; (c) a
Types/pack-layer concern outside Core entirely. This gates the lifecycle half of
the adapter (§5); the CRUD slice proceeds without it.

## Milestones

- **M-A1** — `bdpclient` discovery handshake + live `/sys/capabilities` fetch
  (unambiguous; starts now, testable against `cleanroom-v023`).
- **M-A2** — `connect --protocol bdp` persist (backend name + sidecar) + register
  the `"bdp"` backend + CRUD Scope→Store adapter slice → a read command routes
  end-to-end through RemoteScope to `cleanroom-v023`.
- **M-A3** — lifecycle/dependency mapping (post-Lydia ruling).
- **M-A4** — `bd connect` ↔ `bd serve` interop (blocked on Track B's Go BDP
  server adapter).
