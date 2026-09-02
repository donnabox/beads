# BDP graph store — CLI and storage-interface changes, in detail

**Status:** Draft v3 (W-arch, after two council rounds) — feat/bead-graph
**Date:** 2026-09-02
**Companions:** `BDP_BEAD_GRAPH_PLAN.md` (rulings), `BDP_GRAPH_ARCHITECTURE.md`
(shape; its §2b lists the proposed ruling amendments A1–A8 this spec
assumes). This document is the *diff*: every command, flag, config key,
interface member, package, migration, and gate the graph work adds or
touches — and, just as precisely, what it does not touch (Part C) and what
it changes that an earlier draft claimed it did not (Part C2). Phase
markers follow the plan's §7: **P0** contracts and wire, **P1** storage,
**P2** serving, **P3** writes.

## Part A — CLI

All graph-store commands live under one root, **`bd bdp`** (amendment A3):
`bd link`, `bd graph`, `bd restore`, and `bd promote` are existing verbs
with positional arguments, and the plan's constraint #1 forbids changing
them. The differential gate gains one row per legacy form of each (Part B7).

**Root store policy is keyed by command path for this subtree.** The root
command classifies commands by *leaf name* (`readOnlyCommands`,
`serveCmdName`, the `cmd.Name() != "import"` branches in `cmd/bd/main.go`),
and the tree already records that same-named subcommands collide. So: no
`bd bdp` leaf is named `import`, `export`, or anything in
`readOnlyCommands`; a `bdpRootPolicy` table keyed by `CommandPath()` says,
per verb, whether the local store opens, in which mode, and whether the
post-command maintenance net runs; a test walks the `bd bdp` subtree and
fails on an unlisted leaf. `bd bdp serve` is excluded from maintenance by
that table, not by sharing the leaf name `serve`.

| Verb | Local store | Maintenance | Note |
| --- | --- | --- | --- |
| `bd bdp bead get\|list`, `link get\|list`, `types`, `status` | opens **read-only** when `bdp.client: store`; **skipped entirely** when `bdp.client: server` (no schema check, no auto-start, no migration prompt — the client verbs never touch the local database) | no | `bdp.client` is yaml-only, so it is readable before any store opens — the reason A6 makes it yaml-only |
| `bd bdp serve` | serve's own classification (A3) | no | |
| `bd bdp promote`, `restore`, `ledger snapshot\|apply` | opens writable, always local | no (admin verbs commit their own transactions and push explicitly) | |

### A1. `bd init` — graph store initialization (ruling 12)

`bd init` initializes the graph store with everything else it initializes,
against the normalized storage interfaces:

1. Runs the graph migrations (Part B4) in the normal series. There is **no**
   ignored-series table: the clone-local half is a file (A5, below).
2. **Bootstraps the built-in Type Descriptor catalog through
   `BeadGraphTypeInstaller()`** — idempotent and fingerprint-keyed, so
   re-init converges; W3 supplies the inventory, and until then the catalog
   is honestly empty. Before a Scope exists the install carries an empty
   expectation (there is nothing to assert). A provider answering
   `*storage.ErrUnsupported` makes `bd init` **skip the bootstrap with a
   notice and succeed** — never a failed init (the honest-absence half of
   ruling 12); any other error fails init like any other bootstrap step.
3. Writes **no Scope identity** and no authority file — a workspace has a
   graph store from init but not a Scope; the Scope is minted on first serve
   (A3). Nothing is written to `metadata.json` for the graph (A6).

No new flag is required. **Registered backends:** `bd init` refuses to
provision them today (`cmd/bd/init.go`: "can only open an existing
workspace"); their own workspace-creation path owes the same two
obligations (schema + descriptor bootstrap), proven by the conformance
family (B7). A backend that implements the six accessors as
`ErrUnsupported` stubs (A8) keeps every existing behavior.

### A2. `bd init --bdp-server <url>` — the client reroute (ruling 12, third command)

One more `bd init` target, beside `--server`, `--shared-server`,
`--proxied-server`, `--team-server`, and `--backend`. Every existing target
selects the provider or topology that realizes the storage interfaces — a
choice *below* the normalized abstraction. This one reroutes *above* it, at
the CLI: the `bd bdp` read verbs (A5) become a BDP client of the designated
server instead of opening a store.

**One local source of truth: `config.yaml`** (amendment A6). The keys are
yaml-only (read before any store opens — the GH#536 class
`config.YamlOnlyKeys` exists to prevent), validated at `config set` time,
and never persisted to `metadata.json`.

| Key (config.yaml) | Values | Notes |
| --- | --- | --- |
| `bdp.client` | `store` (default) \| `server` | explicit mode, never inferred from a URL's presence; **not settable from env** (`blockedEnvVars`, the `backend` precedent: a silent reroute is the hazard) |
| `bdp.server` | absolute URL | the Scope URL of the designated server; `https` required unless the host is loopback or `bdp.insecure_http: true` is set explicitly |
| `bdp.insecure_http` | bool (default false) | the named waiver for a plaintext non-loopback target; `bd bdp status` reports it |
| `bdp.scope_url` | absolute URL | what this workspace's server mints/serves (ruling 7a); a *server-side* key, yaml-only for the same reason: it must not replicate through the Dolt `config` table |
| `bdp.authority_heartbeat` | duration (default `30s`, `0` disables) | how often a serving process fetches the remote tracking ref to detect a promotion elsewhere (A7); ignored on workspaces with no remote |

There is deliberately **no token key in config**: every key containing
`token` is a secret to `IsSecretKey`, and the tracked-config guard refuses
it (other trackers keep tokens in yaml-only config *and* accept that; BDP
chooses a file). Mechanics: `bdp.` joins the yaml-only prefix list in
`internal/config/yaml_config.go` and `recognizedConfigPrefixes` in
`cmd/bd/config.go`; `validateYamlConfigValue` gains entries for
`bdp.client` (enum), the two URLs (absolute, scheme check), and the
duration.

**Environment.** Viper binds config keys as `BD_<KEY>` automatically
(`SetEnvPrefix("BD")`, dots to underscores): `BD_BDP_SERVER`,
`BD_BDP_INSECURE_HTTP`, `BD_BDP_AUTHORITY_HEARTBEAT` exist without new
code. Ruling 7a spells the Scope URL variable `BDP_SCOPE_URL`, so it is
bound explicitly — `BindEnv("bdp.scope_url", "BDP_SCOPE_URL",
"BD_BDP_SCOPE_URL")`, the tree's `BEADS_IDENTITY` precedent — and both
spellings work. Hand-read connection/auth environment is `BEADS_*`
(`BEADS_SERVE_TOKEN_FILE`, `BEADS_CREDENTIALS_FILE`, `BEADS_DOLT_PASSWORD`),
so the client token file is **`BEADS_BDP_TOKEN_FILE`**. There is **no
`BD_BDP_TOKEN`**: an env-carried secret is inherited by every child process
(git hooks, dolt subprocesses) — the reason `remote_migrate_gate.go` keeps a
programmatic twin of `BD_ALLOW_REMOTE_MIGRATE`.

**Credential lookup** (client): `BEADS_BDP_TOKEN_FILE` > credentials file
section **`[bdp <origin><scope-path>]`** with `token=` (distinct from the
Dolt `[host:port]` `password=` sections — ruling 7a allows several
path-distinguished Scopes on one host, so a host-keyed section would send
one Scope's token to another). Transport: no redirects followed (a 3xx is
an error naming the location, never re-sent with the credential);
`Authorization` is sent only to the configured origin.

**Precedence, per command** (a table test pins it):

| Command | Resolution |
| --- | --- |
| `bd init --bdp-server <url>` | flag > `BD_BDP_SERVER` env > existing `config.yaml`; writes `bdp.client: server` and `bdp.server` to `config.yaml`; touches nothing else |
| `bd config set bdp.<key>` | validates, writes `config.yaml` (yaml-only routing) — effective for the next command; no other file can outrank it |
| every `bd bdp` verb | env (where permitted) > `config.yaml`; `bdp.client` decides the route (from config.yaml only), `bdp.server` the target |
| `bd bdp status` | prints the resolved route, target, token source (never the token), `insecure_http`, and — on the store route — the identity state row (A3) |

**In `client: server` mode** the verbs' `openBeadGraph*()` accessors return
a BDP-client realization of the same `graphops` role interfaces (A5), so a
verb cannot tell which route it took. A workspace in this mode still has a
local graph store (A1 ran); `bd bdp serve` there **refuses** unless
`--serve-local-store` is passed, so an operator cannot serve a dead local
copy beside the designated server by accident. Issue verbs are unaffected
in either mode.

### A3. `bd bdp serve` — serving a Scope (rulings 7b, 9, 12; amendments A1, A2, A7)

`bd bdp serve` is a **thin command over the existing `internal/httpapi`
server**: the same storage classification, the same hook peel, the same
`httpapi.Config` — plus a populated `Graph` field. It differs from
`bd serve` in one policy: it **requires a Scope** (exit 2 otherwise), where
`bd serve` mounts the BDP rows only when `bdp.scope_url` is configured and
serves exactly what it serves today when it is not (ruling 12: "no URL → no
BDP routes"). Because it builds the same server — every server built from
`routeTable` advertises `issues.claim` (`Capabilities()` derives from the
table) — it **inherits `errServeReadonly` wholesale**: `--readonly` is
refused ahead of the workspace, exactly as `bd serve` refuses it, from v0.
W2 decides whether `bd bdp serve` survives as the strict alias; nothing
needs folding later because there is one server.

```text
bd bdp serve [--addr IP:PORT]            default 127.0.0.1:0 (ephemeral), numeric IP only
             [--allow-non-loopback]      requires --auth-token-file (or --insecure-no-auth)
             [--auth-token-file PATH]    bearer tokens, one per line, hot-reloaded
                                         (env BEADS_SERVE_TOKEN_FILE — serve's own; no new env)
             [--insecure-no-auth]
             [--allowed-host NAME]...    extra exact Host allowlist entries
             [--scope-url URL]           first serve: mints under it (persisted); later: must
                                         equal the persisted Scope URL or refuse
             [--serve-local-store]       permit serving in a client: server workspace
```

No `--dev-local-test`: ruling 7a's `local-test` development mode is the BDP
reference server's; a derived URL served from an in-memory identity could
not satisfy the in-transaction assertion, and persisting it would make it
an identity (both councils, opposite ways). bd's tests configure a real
`bdp.scope_url`.

The five posture flags are `bd serve`'s **variables**, not copies: the
command reuses `serveAddr`/`serveAllowNonLoopback`/… and the same
validators (`ValidateBindAddr`, `ValidateAuthPosture`,
`ValidateAllowedHost`, `NewTokenFileAuth`), and `serveListen` builds the
server. Behavior, in order:

1. **Classification is `serveDatabaseSource`, verbatim:** registered
   backend → store source; embedded Dolt → **refused, permanently**, with
   serve's typed `errServeEmbedded`; otherwise the unit-of-work provider —
   BDP's primary production path (architecture §5).
2. **Roles from beneath the hook layer** — `serveIssueRoles`' one peel, and
   the graph roles are taken **off the same `src` value in the same
   function**, so "same source" is structural (`checkDatabaseSource`
   verifies completeness and hook posture, not provenance — v2 overclaimed).
   For the graph roles the peel is a no-op today (every graph hook wrapper
   recurses unwrapped, B5); kept for uniformity and as the guard a future
   graph hook vocabulary would need.
3. **Capability probe, all-or-nothing:** `BeadGraphReader()`,
   `BeadGraphTypes()`, `BeadGraphIdentityReader()`,
   `BeadGraphBootstrapper()`; any `*storage.ErrUnsupported` → **exit 2,
   typed**; any *other* error aborts startup as serve's role binding does.
   `bd serve` under the same probe: `ErrUnsupported` → BDP rows absent,
   legacy routes exactly as before; operational error → startup failure.
4. **Identity** — the state table, **split by command** (each row a
   contract case, B7). `bd serve` with no `bdp.scope_url` configured is
   always the first column's answer: **no BDP rows, never a refusal**,
   whatever the store holds (ruling 12; Part C).

   | Persisted Scope | Configured URL | Local authority file | `bd bdp serve` | `bd serve` |
   | --- | --- | --- | --- | --- |
   | none | none | — | exit 2 (7a: an explicit URL is required) | no BDP rows |
   | none | set | — | **first serve:** `Mint` (Scope row + first ledger event) → write the local file `{authority_id, epoch: 1, ledger_high_water: 1, state_version}` → serve honestly empty | same |
   | present | none | any | exit 2 | no BDP rows (a notice names the persisted Scope) |
   | present | **different** | any | refuse: configured ≠ persisted (a `--scope-url` on a minted store must equal it) | refuse (a configured URL that contradicts the store is an operator error, not "no URL") |
   | present, matches | matches | **absent** (a clone, or a pull into a fresh directory) | refuse `ErrNotAuthority`; guidance: `bd bdp promote` | same |
   | present, matches | matches | present, `epoch` **stale** (promoted elsewhere, pulled) | refuse; same guidance | same |
   | present, matches | matches | present, `ledger_high_water` **above** the store's max seq (a database restore rewound the state) | refuse `ErrStateRewound`; guidance: `bd bdp restore` | same |
   | present, matches | matches | present, `unverified` set (`bd backup restore` ran) | refuse until `bd bdp restore` clears it | same |
   | present, matches | matches | present, consistent | serve; `Expect = {ScopeURL, AuthorityID, Epoch, LedgerHighWater}` is handed to `httpapi.GraphConfig.Expect` and **re-asserted inside every transaction** (A1) | same |

5. **Host policy:** the Scope URL's host joins the allowlist automatically
   (it is operator-configured identity), because BDP clients dial the Scope
   URL and the rebinding defense otherwise refuses every real request. The
   listener is plaintext behind TLS termination; the Scope URL's `https`
   scheme never matches the socket.
6. **Heartbeat (A7):** on a remote-backed workspace the server fetches the
   remote tracking ref every `bdp.authority_heartbeat` and reads the remote
   Scope row's epoch; higher than its own → it stops serving BDP rows with
   `ErrNotAuthority` (the process stays up for the legacy surface; `bd bdp
   serve` exits 3) and logs the promoting authority id.
7. **Mounting:** `serveListen(opts, httpapi.Config{…, Graph:
   &httpapi.GraphConfig{Reader, Types, Expect, ScopeURL}})`. Liveness stays
   `GET /healthz`; BDP readiness is a real discovery read.
8. **Lifecycle:** excluded from the post-command maintenance net by the
   `bd bdp` policy table (A, above) — not by sharing the leaf name `serve`
   with `serveCmdName` — and serve's events-journal maintenance ticker
   applies as it does to `bd serve`.

### A4. `bd bdp promote` / `bd bdp restore` / `bd bdp ledger` (rulings 9, 11; amendments A5, A7)

These are the **only** reachers of `BeadGraphAdmin()`; the server assembly
has no field for it (architecture §4).

- **`bd bdp promote`** — make *this clone* the authority for the Scope it
  carries, fenced at the remote (A7): (1) `Promote` CASes the replicated
  Scope row (`epoch = epoch + 1 WHERE epoch = <read>`; a lost race is a
  typed refusal — re-read and retry) and appends a ledger event; (2) writes
  the local file `{authority_id: new, epoch, ledger_high_water: max seq,
  state_version: HEAD}`; (3) **pushes** — a non-fast-forward push means
  another promotion landed first: the command fails, the local file is
  removed, and the operator re-reads. The superseded authority cannot land
  writes (its push is non-fast-forward) and stops serving within one
  heartbeat. `--rotate-url <new>` additionally rotates the URL when history
  diverged (operator-confirmed). Workspaces with no remote: the CAS alone
  (there is no second clone).
- **`bd bdp restore`** — runs after a database restore. `bd backup restore`
  (`cmd/bd/backup_restore.go`, `runBackupRestore`, after a successful
  `RestoreDatabase`) calls `Admin.MarkUnverified`; and independently of
  that flag the rewind check (A3 row 7) refuses a store whose ledger is
  behind the local file. `bd bdp restore` branches on the provider's
  declaration:

  | `LedgerDurability` (declared by `BeadGraphIdentityReader`) | Meaning | `bd bdp restore` does |
  | --- | --- | --- |
  | `in-state` (Dolt, v0) | the ledger is ordinary versioned state; a restore carries it only to the restored point | requires a ledger snapshot whose range reaches the local file's `ledger_high_water` (`--ledger <file>`, applied first) to show continuity; **otherwise rotates** the Scope URL (operator supplies the new one) and epoch, appends the refuse-URL event, and re-grants the local file |
  | `independent` | the provider keeps the ledger outside the restored state and proves it (contract case) | re-validates and re-grants; no rotation |
  | `none` | the provider keeps no ledger | always rotates |

  Never silent: every branch prints what it did and why. Residual, stated:
  a whole-directory filesystem snapshot rewinds file and state together and
  is undetectable in-band; operators who take such snapshots schedule
  `bd bdp ledger snapshot` beside them.
- **`bd bdp ledger snapshot [--from SEQ] [--to SEQ]`** / **`bd bdp ledger
  apply <file>`** — the continuity lane ruling 11 needs: the append-only
  `graph_ledger_events` as JSONL, **bound to the Scope** (`scope_url`,
  `authority_id` lineage) and to a **contiguous `[from, to]` range**. Apply
  validates provenance and contiguity against the store's current max seq
  (a gap or a foreign lineage is refused), appends the missing events, and
  re-derives `graph_allocations` from them. It never rewinds a tombstone or
  a refusal.

### A5. Graph verbs (v0 reads; P3 writes; names fixed now)

```text
bd bdp bead get <path> | list [--type URL] [--after CURSOR] [--limit N]
bd bdp link get <path> | list [--type URL] [--source PATH] [--target REF] [--after CURSOR] [--limit N]
bd bdp types [get <url>]
bd bdp status
bd bdp serve | promote | restore | ledger snapshot|apply
(P3)  bd bdp bead create|update|delete ; bd bdp link create|update|delete
```

Each verb reaches its role through an accessor and nothing else, on
whichever route the workspace is on — the `cmd/bd/label.go` pattern plus
the third route, with a route-fork test in the shape of
`cmd/bd/vc_recompute_test.go`:

```go
func openBeadGraphReader() (graphops.Reader, error) {
    if bdpClientMode() == "server" {        // config.yaml only, per A2
        return bdpclient.Reader(bdpClientConfig())          // internal/bdpclient
    }
    if usesProxiedServer() { return proxiedBeadGraphReader() }
    return store.BeadGraphReader()
}
```

`internal/bdpclient` speaks the pinned wire (`bdpwire`, B8) and implements
`graphops.Reader` and `graphops.DescriptorReader` over it; a BDP Problem
maps back to the same typed graph errors the store bodies raise (round-trip
test: every Problem in the pinned vocabulary → `errors.Is` holds across the
fork, gone-family included). **Until P2's cursor ADR lands there are no
collection routes**, so on the client route `bead list` / `link list`
answer the typed `ErrNotServedYet`; on the store route they work from P1.
Pagination through the client is cursor-faithful: the CLI never re-sorts.

## Part B — Storage interfaces

### B1. Accessors on `storage.Storage` (the house rule — architecture §2a; amendments A4, A8)

Named **`BeadGraph*`**: `Storage.GraphCounter()` already exists and counts
the *issue* dependency graph; six bare `Graph*` accessors beside it would
invite exactly the confusion the plan's terminology block exists to prevent.

```go
// BeadGraphReader returns the guarded graph-read surface for this store:
// Bead and Link records (a Bead with its complete, grouped, bounded
// ownedLinks, assembled in the call's one transaction), keyset selections
// in code-unit order under an opaque cursor, and incident Links. Every
// method asserts the caller's AuthorityExpectation inside that transaction.
// Reads fire no hooks; the hook decorator recurses.
BeadGraphReader() (graphops.Reader, error)

// BeadGraphTypes returns the Type Descriptor inventory (ordered, bounded,
// keyed), under the same expectation.
BeadGraphTypes() (graphops.DescriptorReader, error)

// BeadGraphTypeInstaller returns the descriptor install/converge role used
// by bd init's bootstrap and the conformance fixture. A caller entitled to
// read the catalog is not thereby entitled to change it.
BeadGraphTypeInstaller() (graphops.TypeInstaller, error)

// BeadGraphIdentityReader returns the Scope row, this clone's authority
// claim, and the provider's LedgerDurability declaration.
BeadGraphIdentityReader() (graphops.IdentityReader, error)

// BeadGraphBootstrapper returns the one-time mint. The server assembly may
// hold it on the first-serve path and nothing else.
BeadGraphBootstrapper() (graphops.ScopeBootstrapper, error)

// BeadGraphAdmin returns promote/rotate/ledger/unverified. Reached only by
// the bd bdp admin verbs and bd backup restore; never by a server.
BeadGraphAdmin() (graphops.Admin, error)

// P3: BeadGraphWriter() (graphops.Writer, error)
```

Added to **`storage.Storage`** (verified: all 28 accessors live there;
`DoltStorage` embeds it; both decorators embed `DoltStorage`). **Every
decorator and provider wrapper declares each accessor explicitly** — hook
layer, telemetry, the notifying UOW provider, `timedProvider` in `httpapi`,
`serveRoleSource` and its test stubs, `internal/jira/tracker_test.go`'s
`configStore` stub, and every other `Storage` implementer the censuses
enumerate. Promotion through the embedded interface compiles and is the
failure mode; the three censuses (B5) fail the build that relies on it.
**This is the source break `storage.go` declares for out-of-tree
backends** (amendment A8): six one-line `ErrUnsupported` stubs, called out
in CHANGELOG by this slice (no earlier accessor wrote the entry
`backend/backend.go` promises).

### B2. The `graphops` leaf (public, repo root; amendment A4)

```go
package graphops   // imports: stdlib, beadserrors — nothing else

// ---- requests and results (the leaf doc specifies each field; contract cases cite it by line)
type AuthorityExpectation struct {
    ScopeURL        string
    AuthorityID     string
    Epoch           uint64
    LedgerHighWater uint64   // max ledger seq this process has observed; the store's must be >= it
}                             // the zero value is accepted ONLY while no Scope row exists (pre-mint)
type Cursor string            // OPAQUE: store-produced; binds Scope URL, epoch, selection hash, last path;
                              // P2 adds snapshot identity inside it — no public interface change
type BeadRequest   struct{ Path string; Expect AuthorityExpectation }
type LinkRequest   struct{ Path string; Expect AuthorityExpectation }
type BeadSelectRequest struct{ TypeURL string; After Cursor; Limit int; Expect AuthorityExpectation }
type LinkSelectRequest struct{ TypeURL string; SourcePath string; Target *Ref; After Cursor; Limit int; Expect AuthorityExpectation }
type IncidentRequest   struct{ Path string; Direction Direction /* In | Out | Both */; After Cursor; Limit int; Expect AuthorityExpectation }
type DescriptorsRequest struct{ Expect AuthorityExpectation }
type DescriptorRequest  struct{ URL string; Expect AuthorityExpectation }
type InstallRequest     struct{ Descriptors []TypeDescriptor; Expect AuthorityExpectation } // Expect required once a Scope exists

type OwnedLinkGroup struct{ TypeURL string; Links []Link }   // the pinned schema keys ownedLinks by Link Type URL
type BeadRecord struct{ Bead Bead; OwnedLinks []OwnedLinkGroup }  // groups in code-unit order of TypeURL; Links in code-unit
                                                                   // order of path; an owned Type with no Links is an EMPTY group
type BeadPage struct{ Items []BeadRecord; Next Cursor }   // Next == "" means exhausted; items carry ownedLinks (bounded)
type LinkPage struct{ Items []Link; Next Cursor }

type Reader interface {
    Bead(ctx, BeadRequest) (BeadRecord, error)
    Link(ctx, LinkRequest) (Link, error)
    Beads(ctx, BeadSelectRequest) (BeadPage, error)       // WHERE path > last ORDER BY path LIMIT n, binary-collated column
    Links(ctx, LinkSelectRequest) (LinkPage, error)
    IncidentLinks(ctx, IncidentRequest) (LinkPage, error) // two index scans (source, target) merged in code-unit order
}
type DescriptorReader interface {
    Descriptors(ctx, DescriptorsRequest) ([]TypeDescriptor, error)   // ordered by URL; bounded by MaxCatalog
    Descriptor(ctx, DescriptorRequest) (TypeDescriptor, error)
}
type TypeInstaller interface {
    Install(ctx, InstallRequest) (InstallResult, error)   // idempotent by fingerprint; closure validated; an owning
}                                                          // declaration without Max → ErrValidation; over MaxCatalog → ErrValidation
type IdentityReader interface {
    Read(ctx) (ScopeIdentity, error)                      // Scope row + the local file's claim (Granted bool, Epoch, HighWater)
    LedgerDurability(ctx) (LedgerDurability, error)       // in-state | independent | none (ruling 11)
}
type ScopeBootstrapper interface {
    Mint(ctx, MintRequest) (ScopeIdentity, error)         // INSERT into the singleton row + first ledger event + local file
}
type Admin interface {
    Promote(ctx, PromoteRequest) (ScopeIdentity, error)   // epoch CAS + ledger event + local file; the CLI pushes after
    Rotate(ctx, RotateRequest) (ScopeIdentity, error)     // new URL; refuse-URL event
    LedgerSnapshot(ctx, LedgerRange) ([]LedgerEvent, error)
    LedgerApply(ctx, []LedgerEvent) (LedgerApplyResult, error)
    MarkUnverified(ctx) error
    ClearUnverified(ctx) error
}
```

**Bounds.** `MaxExpandedRows` (10,000) caps `len(page) + Σ ownedLinks` for
any read; the batched owned-links query carries **`LIMIT bound+1`** so an
over-bound (or locally corrupted) store never materializes more than the
bound before `ErrRepresentationTooLarge` is returned; descriptor decoding is
capped the same way. Owning Types **must** declare `Max`; the installer
refuses one that does not (W3's generator emits it). Reads are **one page
query plus one batched owned-links query** (`source_path IN (page) AND
type_url IN (owned types)`), grouped in Go — no per-row statements; the
contract's query-count case pins **four statements per single-resource
read** (authority + high-water, row, links, catalog fingerprint) and five
per page.

**Values** (`Bead`, `Link`, `Ref`, `Properties`, `Revision`, `Attribution`,
`TypeDescriptor`, `OwnedLinkDecl`, `ScopeIdentity`, `LedgerEvent`) have
unexported fields and constructors that enforce the laws in `laws.go` (same
package). `Properties` is the immutable raw-JSON object value from the plan
(duplicate-key rejection, number preservation, deterministic encoding, RFC
6902 §4.6 equality); its canonical bytes are what B4 stores. `Ref` is a
sum: in-Scope (`Path`) or external (`URL`), with the discriminant stored
(B4). `Revision` is 128 bits from `crypto/rand`, lower-hex; minted by the
write body, unique per resource (a collision on re-mint is retried, never
returned).

**Errors:** `beadserrors` (stdlib-only) declares the new sentinels
(`ErrNoScope`, `ErrScopeExists`, `ErrNotAuthority`, `ErrStateRewound`,
`ErrURLReused`, `ErrRepresentationTooLarge`, `ErrNotServedYet`) and the
typed `GoneError{Path, State}` (state `pruned` | `erased`, from the
allocation's tombstone state — what the handler needs to emit
`resource-pruned` / `resource-erased`); `graphops` aliases them so
`errors.Is` crosses the module boundary.

### B3. Bodies and legs

`internal/storage/graphops` bodies take **`DBTX`** (the `issueops.DBTX`
shape, declared locally: `ExecContext`/`QueryContext`/`QueryRowContext`),
which `*sql.Tx` and `domain/db.Runner` both satisfy — the decision the
guide's mechanical test gives at *design* time (the `MetadataCAS`/
`TreeWalker` precedent), not "per method at implementation time".

| Leg | Files | Body |
| --- | --- | --- |
| server Dolt | `internal/storage/dolt/beadgraph_reader.go`, `beadgraph_types.go`, `beadgraph_identity.go` | accessor wraps the body in `withReadTx` (reads) / `withRetryTx` (install, mint, promote, rotate, apply) |
| embedded Dolt | `internal/storage/embeddeddolt/beadgraph_*.go` (`//go:build cgo`) | same body, `withConn` |
| unit of work | `internal/storage/domain/beadgraph.go` (`BeadGraphUseCase` interface), `internal/storage/domain/db/beadgraph.go` (over `db.Runner`, delegating to the bodies — the `IssueUseCase()` → `CompareAndSetMetadataKeyInTx` shape), `internal/storage/uow/beadgraph_*.go` (`uow.UnitOfWork.BeadGraphUseCase()`; provider accessors via `RunTxRead`; mutations via `RunTxResult` with commit messages `bdp: mint scope <url>` / `bdp: promote epoch <n>` / `bdp: install types <fingerprint>` / `bdp: ledger apply <from>-<to>`; a no-op result commits nothing) | **same body** |

Contract headers therefore say **"one reading plus an engine check"** for
every graph role. Every body begins with `assertAuthorityInTx(ctx, tx,
req.Expect)`: it reads the singleton Scope row and `MAX(seq)` from the
ledger in the same transaction and returns `ErrNotAuthority` on an
epoch/id mismatch or `ErrStateRewound` when the store's high-water mark is
below the expectation's (a zero expectation is accepted only while no Scope
row exists). When the process's recorded `state_version` differs from the
store's (Dolt: HEAD hash; other providers: their state token), the body
runs `ValidateStateInTx` first — the replication/merge ADR's validator —
and refuses on a foreign-authority or invalid delta. `SeedBeadInTx`/
`SeedLinkInTx` (P1 fixture writer) allocate through the ledger like a real
write; every call site is a `_test.go` file (source-scan test) and the
package is denied to everything but the three legs and `domain/db`
(depguard, B6).

### B4. Schema (migrations; frozen once merged)

Rules the tree enforces and this schema obeys: migrations are **frozen once
merged** and content-hashed across clones (`check-migration-hygiene.sh`
check C); **no `NOW()`/`UUID()`/`RAND()`** in migration SQL (check B) — every
timestamp and id is set in Go; real-Dolt tests for anything a `sqlmock` echo
cannot exercise (`BEADS_TEST_EMBEDDED_DOLT=1 -tags cgo`); DDL is not
transactional across statements, so each `CREATE` is guarded and resumable.
**Seven replicated tables, no ignored-series table** (the clone-local half
is the `.beads/` file). Migration files: `NNNN_beadgraph_scope.up.sql`,
`NNNN_beadgraph_types.up.sql`, `NNNN_beadgraph_beads.up.sql`,
`NNNN_beadgraph_links.up.sql`, `NNNN_beadgraph_ledger.up.sql` (events +
allocations) — five files, each with its `.down.sql`.

**Collation.** Dolt's default collation is already binary
(`utf8mb4_0900_bin`: `Foo`/`foo` are distinct keys and `ORDER BY` is
code-unit — probed in round 2), and no migration in the tree declares one.
Every identifier column below still carries **`CHARACTER SET utf8mb4
COLLATE utf8mb4_bin`** explicitly (written `BIN` in the table), as the
defense for providers whose default is case-insensitive and against a
future default change — with a contract case asserting two paths differing
only in case are distinct and sort by code unit.

**Identity is Scope-relative.** Rows store the canonical Scope-relative
`path`; the absolute URL is `scope_url + path`, computed at the boundary
from the live Scope row — so a URL rotation (ruling 11) rewrites no rows,
and code-unit order of paths equals code-unit order of absolute URLs under
one Scope prefix. External references keep their absolute URL with a
discriminant.

**JSON is bytes.** `properties` and `descriptor` are canonical JSON bytes in
`LONGBLOB`, never the engine `JSON` type: the tree measured that column
decoding through float64 (`1.0`→`1`, integers past 2^53 rounded, `-0.0`→`0`,
`1e300` expanded — `internal/storage/issueops/metadata_cas.go`, guide §"a
precision the substrate does not keep"). Round-trip tests cover each
measured case. Size limit 1 MiB per value (`ErrValidation` above it).

**Provenance on every row.** `last_authority_id` / `last_epoch` are stamped
by every mutation, so the state-change validator can identify a foreign
*update*, not only a foreign birth (round 2).

| Table | Columns (type; nullability) | Keys / constraints |
| --- | --- | --- |
| `graph_scope` | `id TINYINT NOT NULL` (always 1), `scope_url VARCHAR(2048) BIN NOT NULL`, `authority_id CHAR(32) NOT NULL`, `epoch BIGINT UNSIGNED NOT NULL`, `minted_at DATETIME(6) NOT NULL` | `PRIMARY KEY (id)`, `CHECK (id = 1)` — **singleton**: `Mint` is an INSERT that loses the race |
| `graph_scope_history` | `scope_url VARCHAR(2048) BIN NOT NULL`, `refused_seq BIGINT UNSIGNED NOT NULL`, `refused_at DATETIME(6) NOT NULL`, `reason VARCHAR(64) NOT NULL` | `PRIMARY KEY (scope_url)`; derived from refuse-URL ledger events |
| `graph_type_descriptors` | `url VARCHAR(2048) BIN NOT NULL`, `descriptor LONGBLOB NOT NULL`, `fingerprint CHAR(64) NOT NULL`, `installed_at DATETIME(6) NOT NULL`, `last_authority_id CHAR(32) NOT NULL`, `last_epoch BIGINT UNSIGNED NOT NULL` | `PRIMARY KEY (url)`; `UNIQUE (fingerprint)` |
| `graph_beads` | `path VARCHAR(1024) BIN NOT NULL`, `type_url VARCHAR(2048) BIN NOT NULL`, `revision CHAR(32) NOT NULL`, `attribution_principal VARCHAR(512) NULL`, `attribution_status ENUM('claimed','unknown') NULL`, `properties LONGBLOB NOT NULL`, `last_authority_id CHAR(32) NOT NULL`, `last_epoch BIGINT UNSIGNED NOT NULL`, `created_at DATETIME(6) NOT NULL`, `updated_at DATETIME(6) NOT NULL` | `PRIMARY KEY (path)`; `INDEX (type_url, path)`; `FOREIGN KEY (type_url) REFERENCES graph_type_descriptors(url)`; attribution columns both NULL or both set (`CHECK`) |
| `graph_links` | `path VARCHAR(1024) BIN NOT NULL`, `type_url … BIN NOT NULL`, `revision CHAR(32) NOT NULL`, `source_path VARCHAR(1024) BIN NOT NULL`, `source_pin CHAR(32) NULL`, `target_kind ENUM('in','ext') NOT NULL`, `target_path VARCHAR(1024) BIN NULL`, `target_url VARCHAR(2048) BIN NULL`, `target_pin CHAR(32) NULL`, `attribution_*`, `properties LONGBLOB NOT NULL`, `last_authority_id`, `last_epoch`, timestamps | `PRIMARY KEY (path)`; `INDEX (source_path, type_url, path)` (ownedLinks assembly), `INDEX (target_path, path)` (incident); `FOREIGN KEY (source_path) REFERENCES graph_beads(path)` (a Link's source is in-Scope — law); `CHECK` exactly one of `target_path`/`target_url` per `target_kind`; **no** uniqueness on (type, source, target) — multiplicity is law |
| `graph_ledger_events` | `seq BIGINT UNSIGNED NOT NULL`, `kind ENUM('allocate','tombstone','refuse_url') NOT NULL`, `path VARCHAR(1024) BIN NULL`, `scope_url VARCHAR(2048) BIN NULL`, `resource_kind ENUM('bead','link') NULL`, `state ENUM('pruned','erased') NULL`, `authority_id CHAR(32) NOT NULL`, `epoch BIGINT UNSIGNED NOT NULL`, `at DATETIME(6) NOT NULL` | `PRIMARY KEY (seq)` — **append-only**, one global sequence allocated in the write transaction; `INDEX (path, seq)`; `CHECK` per kind on which of `path`/`scope_url` is set |
| `graph_allocations` | `path VARCHAR(1024) BIN NOT NULL`, `resource_kind ENUM('bead','link') NOT NULL`, `birth_seq BIGINT UNSIGNED NOT NULL`, `birth_authority_id CHAR(32) NOT NULL`, `birth_epoch BIGINT UNSIGNED NOT NULL`, `state ENUM('live','pruned','erased') NOT NULL`, `tombstone_seq BIGINT UNSIGNED NULL` | `PRIMARY KEY (path)` — the O(1)/O(log n) ID test (ruling 3); **derived state** re-derivable from the events (insert-once, update-on-tombstone; the append-only property lives in the events table) |

`updated_at` is protocol-irrelevant bookkeeping; revisions are minted by
the write body, never derived from timestamps. Bead `type_url` and Link
`source_path`/`target_*` are immutable after insert (enforced in the write
body and checked by the validator).

**The clone-local half: `.beads/graph-authority.local.json`.** Gitignored
(the resolved-port-file precedent; `.beads/.gitignore` lists it),
project-id-stamped so a file copied between workspaces is refused:
`{project_id, scope_url, authority_id, epoch, ledger_high_water,
state_version, unverified, granted_at}`. Written atomically (temp file +
rename) by `Mint`, `Promote`, `Rotate`, `LedgerApply`, and — after commit —
by every write transaction (P3) to advance `ledger_high_water` (writers take
the max; last-writer-wins is safe because the value only grows).
Neither `dolt clone` nor `git clone` nor a pull carries it; a `DOLT_BACKUP`
restore leaves it in place with a high-water mark the restored ledger
cannot meet (A3 row 7).

**Cross-repo coupling (bts).** `DoltTeamServer` workspaces refuse to open
when `current < latest` with "ask your operator to run `bts migrate`" and
**no `BD_IGNORE_SCHEMA_SKEW` hatch** (`internal/storage/uow/team_server_schema.go`)
— a **numeric-version** comparison only: equal version numbers with
different SQL pass silently. So the coupling is a cross-repository
**release-parity gate**, not a runtime check: bts must ship byte-identical
copies of the five files (the embedded content hashes in
`migration_content_hashes.go` are what a bts-side parity test compares),
and every bts-managed workspace refuses the next bd release until it does.
The migration PR is sequenced with bts and says so; the remote-migrate gate
(#4259) forces migrate-vs-adopt on every remote-backed workspace at upgrade
as it does for any migration.

### B5. Decorators, censuses, and every embedding surface

- `internal/storage/hook_beadgraph_*.go` (six files): declared, recurse
  **unwrapped** — the hook vocabulary (`on_create/on_update/on_close`)
  hands scripts an *issue*; a graph hook vocabulary is a separate proposal.
  `storage.RoleFiresHooks` is a type switch over hook wrappers, so an
  unwrapped role needs no entry; a test asserts each graph role answers
  `false`. Added to `role_accessor_decorator_test.go`'s wrapped/unwrapped
  table and to the storage file's pass-through paragraph.
- `internal/telemetry/beadgraph_*.go`: every method spanned `storage.op` /
  `storage.done`; the **telemetry census**
  (`internal/telemetry/role_accessor_decorator_test.go`) gains the
  classification.
- **Three censuses**, each must learn `graphops`: the storage reflection
  census (`roleAccessorNamesOf`), the telemetry census, and the conformance
  package's **source-parsed** `facadePackages` map
  (`backend/conformance/role_coverage_scan_test.go`) — without the third,
  `TestRoleFacadeCensusAgreesWithReflection` fails, which is the guard
  working.
- `internal/storage/uow/notifying.go`: explicit accessors built from *this*
  provider (parity test); `internal/httpapi/claim.go`'s `timedProvider`
  (per-request accessors so units of work land in `uow_ms`); `cmd/bd/serve.go`'s
  `serveRoleSource` and its non-embedding test stubs;
  `internal/jira/tracker_test.go`'s `configStore` stub — every surface that
  embeds the store or a provider, enumerated by `grep -l 'func (.*) Memories()'`
  at implementation and listed in the slice's PR.

### B6. `backend/` public surface and depguard

- **No aliases** (amendment A4): `graphops` is public and imported
  directly, like `issueops`. `TestPublicSurfaceComplete` stays green
  *because* no `internal/` type is reachable from the new accessors — a
  test asserts that.
- `backend/backend.go`'s "minimal external backend" example gains the six
  accessors as `ErrUnsupported` stubs; its stability note's promised
  **CHANGELOG entry** is written (`BREAKING (published backend package):
  six BeadGraph accessors are required on Storage; stub them with
  ErrUnsupported to keep existing behavior`).
- `.golangci.yml` gains a **new, stricter** rule (cmd/bd imports the
  `issueops` tx-body package in five files today, so this is not the
  existing convention): `internal/storage/graphops` is importable only by
  `internal/storage/{dolt,embeddeddolt,domain/db}` and its own tests, with a
  mutation test that removes the entry and expects the lint to pass —
  proving the entry is load-bearing.

### B7. Conformance

- Families: `beadgraph_reader_contract.go`, `beadgraph_types_contract.go`
  (reader + installer), `beadgraph_identity_contract.go` (reader,
  bootstrapper, admin), each citing the leaf doc by line.
- `RoleContractBundle` gains `BeadGraphReader`, `BeadGraphTypes`,
  `BeadGraphTypeInstaller`, `BeadGraphIdentityReader`,
  `BeadGraphBootstrapper`, `BeadGraphAdmin` factory fields **and** their
  rows in `role_bundle_cases.go` (`TestRoleContractCasesMatchTheBundleFields`);
  `BeadGraphFixture` carries the seed hook (`SeedBead`/`SeedLink` → the
  ledger-enforcing InTx writers) and a `LocalAuthority` hook so the fixture
  can stand in for the `.beads/` file.
- Wirings in `internal/storage/{dolt,embeddeddolt,uow}/beadgraph_*_contract_test.go`;
  the leg registry (`internal/storage/contract_leg_registry_test.go`) and
  `TestEveryLegWiresEveryRoleContract`
  (`internal/storage/leg_contract_wiring_test.go`) both see them; both
  coverage gates apply (`TestEveryRoleMethodHasAContractCase`,
  `unwiredContractEntrypoints`); waivers only with a reason, only until wired.
- Non-capable stores answer `*storage.ErrUnsupported{Op: "<accessor
  name>"}` — `RunUnsupportedContract` compares `Op` to the method name
  exactly, so the six strings are pinned — proven per accessor.
- Cases the councils asked for by name: a clone produced by push/pull
  refuses (`ErrNotAuthority`); a **`DOLT_BACKUP` restore of an authority
  clone refuses** (`ErrStateRewound`); concurrent first-serve mint (one
  wins); promotion race (one CAS wins; the loser's push is non-fast-forward);
  heartbeat detects a higher remote epoch; rotation refuses the old URL;
  case-differing paths distinct and code-unit ordered; ownedLinks
  completeness incl. **empty groups** and the bound with
  `ErrRepresentationTooLarge` under `LIMIT bound+1`; keyset continuation
  inside one transaction with an opaque cursor; gone-family from tombstone
  state; expectation mismatch mid-process (promote elsewhere between two
  reads); descriptor read under a stale expectation refuses; install after
  mint without the expectation refuses; `LedgerDurability` answer per
  provider; ledger snapshot/apply round trip, gap refusal, foreign-lineage
  refusal; installer refuses an owning declaration without `Max`;
  statement-count = 4 for a single-resource read.
- Differential-gate rows: every legacy form of `bd link`, `bd graph`,
  `bd graph check`, `bd restore`, `bd promote` parses and behaves as before.

### B8. `httpapi` integration and the pinned wire (amendments A2, A7)

- **`httpapi.Config.Graph *GraphConfig`** — `{Reader graphops.Reader; Types
  graphops.DescriptorReader; Expect graphops.AuthorityExpectation; ScopeURL
  string; Heartbeat HeartbeatSource}`; nil = no BDP rows. On the provider
  arm the roles are rebuilt per request through `timedProvider` like every
  other role. There is deliberately **no admin or installer field**: a test
  asserts the struct cannot carry one.
- **`bdpRouteTable`** (`internal/httpapi/bdp_routes.go`) — **P2**: rows in
  the same `route` shape (op, method, pattern; `authExempt: false`;
  `projectExempt: false` — an absent stamp passes and BDP clients never send
  one; semaphore taken), registered by `Server.handler()` **only when
  `cfg.Graph != nil`**, each wrapped by the same `s.route(rt)` — so
  deadline, bearer-before-semaphore, stamp-behind-auth, and the never-log-
  the-token rule apply by construction. First rows: discovery, `types/`,
  one Bead, one Link; **collection rows wait for P2's cursor ADR**. BDP
  rows contribute **no token** to `ContextResponse.capabilities` in v0
  (BDP has its own discovery; W2 may add a behavior token);
  `TestSpecRouteParity` continues to compare only `routeTable` against
  `openapi.v0.yaml`, and a sibling parity test compares `bdpRouteTable`
  against the pinned schema's path grammar.
- **Posture parity test:** one refusal matrix (missing / malformed / unknown
  token, bad `Host`, saturated semaphore, deadline) drives a legacy row and
  a BDP row and asserts identical status and log shape.
- **Handler = serializer** (`bdp_handlers.go`): `graphops` records → `bdpwire`
  DTOs; typed graph errors → BDP Problem records (`bdp_problem.go`), the
  closed vocabulary, here and only here.
- **Wire — P0, not yet in the tree:** `internal/httpapi/bdpwire/schema/bdp-v0.schema.json`
  **will be vendored** with a `PROVENANCE` file (upstream repo, commit — the
  plan's §0 pin — and sha256); `make bdp-gen` runs a **pinned**
  JSON-Schema→Go generator (version recorded in `Makefile`; if the pinned
  schema exceeds what it can express, P0 records the fallback: hand-written
  DTOs validated against the schema in tests); `make bdp-check` regenerates
  and diffs, and `scripts/ci/pr-policy.sh` runs it beside `make api-check`.
  The artifact's presence and hash are P0's exit gate. The OpenAPI generator
  and document are untouched.

## Part C — What does not change

`storage.Storage`'s existing 28 accessors and every `issueops` role; the
journal's frozen vocabulary; `openapi.v0.yaml` and `TestSpecRouteParity`;
`bd serve` on a workspace with **no** `bdp.scope_url` (no rows, no
capability change, no startup refusal whatever the store holds —
byte-identical); every legacy CLI verb (differential gate rows in B7);
JSONL export shapes; `metadata.json`'s schema.

## Part C2 — What changes that an earlier draft claimed did not

- **Merge, pull, and sync.** The entry points are not four Go functions:
  `MergeAndSettle` / `MergeAndSettleWithStrategy` / `MergeWithStrategy` /
  `Merge` in `versioncontrolops`; `CALL DOLT_PULL` inside every pull route
  (`doltCLIPull`, `Pull`, `PullRemote`, `pullTransport`/`pullWithAutoResolve`
  in `internal/storage/dolt/store.go`); the UOW leg's `DoltRemoteUseCase`
  (`DOLT_MERGE`/`DOLT_PULL` directly); embedded federation sync; the
  remote-migrate gate's fast-forward `DOLT_MERGE`. A Go wrapper cannot see a
  server-side merge, so the **state-change validator** runs on every
  observed `state_version` change at authority-assertion time (B3) and
  refuses a foreign-authority or invalid delta before any read or write is
  answered. The replication/merge ADR (architecture §2b) specifies its
  rules before the migrations land.
- **Federation.** Graph tables ride filtered pushes **unfiltered, by
  decision, in v0**; a per-topology filter hook is a later ruling, and a
  filter that drops an endpoint must drop the Link (never a dangling edge).
- **`bd sql` and raw SQL.** Documented as **out of contract** for graph
  tables (the command already says it bypasses storage); the validator
  catches what merges import and what a local `bd sql` leaves inconsistent
  at the next state-version change; the enforcement-boundary ruling
  (architecture §2b) decides whether DB privileges or triggers close the
  rest before P3.
- **`bd backup restore`** calls `Admin.MarkUnverified` after a successful
  `RestoreDatabase` (A4).
- **Root store policy** gains a CommandPath-keyed table for the `bd bdp`
  subtree (Part A).

## Part D — Open implementation questions (not rulings)

1. `MaxExpandedRows`, `MaxCatalog`, page default/max, and the 1 MiB value
   limit are proposed numbers; the leaf doc fixes them at P1 with the
   rationale, and the conformance cases cite them.
2. Whether the P1 fixture writer stays fixture-only through P3 or becomes
   the internal half of `graphops.Writer` (W1 decides the wire; the body
   is shared either way).
3. Generator choice for `bdpwire` (pinned JSON-Schema→Go tool vs
   hand-written DTOs under schema validation) — recorded at P0 with the
   provenance file.
4. Whether `bd bdp serve` remains after W2 (default: the strict alias).
5. Whether the heartbeat should also gate `bd bdp` admin verbs (promote
   already pushes; a stale `bd bdp status` is informational).

## Part E — Proposed ruling amendments this spec assumes (pending)

A1 per-call transaction with in-transaction authority expectation is the v0
lease (ruling 9); A2 BDP rows inside `httpapi`, `bd bdp serve` a thin strict
command inheriting `errServeReadonly` (rulings 7b/12); A3 `bd bdp …`
namespace with CommandPath-keyed policy; A4 values, laws, roles in public
`graphops`, `BeadGraph*` accessors, no `backend/` aliases; A5 ruling 11's
mechanism = `.beads/` authority file with a ledger high-water mark +
append-only ledger events + `bd bdp ledger snapshot|apply` + provider
`LedgerDurability`; A6 `config.yaml` single source, `BDP_SCOPE_URL`
honored, no env token, `bdp.client` blocked from env; A7 promotion fenced at
the remote with a heartbeat; A8 constraint #1 scoped to behavior, out-of-tree
stubs. Plus two decisions the plan does not yet hold: the out-of-role DML
enforcement boundary, and the replication/merge ADR as a P1 gate. Full text:
architecture §2b.
