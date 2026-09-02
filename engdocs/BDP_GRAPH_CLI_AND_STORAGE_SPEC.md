# BDP graph store — CLI and storage-interface changes, in detail

**Status:** Draft v2 (W-arch, post-council) — feat/bead-graph
**Date:** 2026-09-02
**Companions:** `BDP_BEAD_GRAPH_PLAN.md` (rulings), `BDP_GRAPH_ARCHITECTURE.md`
(shape; its §2b lists the six proposed ruling amendments A1–A6 this spec
assumes). This document is the *diff*: every command, flag, config key,
interface member, package, migration, and gate the graph work adds or
touches — and, just as precisely, what it does not touch (Part C) and what
it changes that v1 claimed it did not (Part C2).

## Part A — CLI

All graph-store commands live under one root, **`bd bdp`** (amendment A3):
`bd link`, `bd graph`, `bd restore`, and `bd promote` are existing verbs
with positional arguments, and the plan's constraint #1 forbids changing
them. The differential gate gains one row per legacy form of each (Part B7).

### A1. `bd init` — graph store initialization (ruling 12)

`bd init` initializes the graph store with everything else it initializes,
against the normalized storage interfaces:

1. Runs the graph migrations: the **replicated** series (Part B4) in the
   normal migration series, and the **clone-local** `graph_authority_local`
   table in the `migrations/ignored/` series with its `dolt_ignore` entry
   registered before creation (the 0028/0019 precedent).
2. **Bootstraps the built-in Type Descriptor catalog through
   `GraphTypeInstaller()`** — idempotent and fingerprint-keyed, so re-init
   converges; W3 supplies the inventory, and until then the catalog is
   honestly empty. A provider answering `*storage.ErrUnsupported` here makes
   `bd init` **skip the bootstrap with a notice and succeed** — never a
   failed init (the honest-absence half of ruling 12). Any other error fails
   init as it would for any other bootstrap step.
3. Writes **no Scope identity** — a workspace has a graph store from init but
   not a Scope; the Scope is minted on first serve (A3). Nothing is written
   to `metadata.json` for the graph (amendment A6).

No new flag is required. **Registered backends:** `bd init` refuses to
provision them today (`cmd/bd/init.go`: "can only open an existing
workspace"); their own workspace-creation path owes the same two
obligations (schema + descriptor bootstrap), proven by the conformance
family (B7), and a registered workspace whose provider lacks the capability
answers `ErrUnsupported` from every graph accessor.

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
| `bdp.client` | `store` (default) \| `server` | explicit mode, never inferred from a URL's presence |
| `bdp.server` | absolute URL | the Scope URL of the designated server; `https` required unless the host is loopback or `bdp.insecure_http: true` is set explicitly |
| `bdp.insecure_http` | bool (default false) | the named waiver for a plaintext non-loopback target; refused silently nowhere — `bd bdp status` reports it |
| `bdp.scope_url` | absolute URL | what this workspace's server mints/serves (ruling 7a) — a *server-side* key, yaml-only for the same reason: it must not replicate through the Dolt `config` table |
| `bdp.token_file` | path | client bearer token file (one token); alternative to the credentials file |

Mechanics: `bdp.` joins the yaml-only prefix list in
`internal/config/yaml_config.go` and `recognizedConfigPrefixes` in
`cmd/bd/config.go`; `validateYamlConfigValue` gains entries for
`bdp.client` (enum) and the two URLs (absolute, scheme check). No token key
ever lives in config (`IsSecretKey` would flag it; the rule is stated here
so nobody tries).

**Environment.** The tree has two conventions and this spec follows both
exactly: viper-bound config keys are reachable as `BD_<KEY>` automatically
(`SetEnvPrefix("BD")`, dots to underscores), so `BD_BDP_SERVER`,
`BD_BDP_CLIENT`, `BD_BDP_SCOPE_URL` exist without new code; hand-read
connection/auth environment is `BEADS_*` (`BEADS_SERVE_TOKEN_FILE`,
`BEADS_CREDENTIALS_FILE`, `BEADS_DOLT_PASSWORD`), so the client token file is
`BEADS_BDP_TOKEN_FILE`. There is deliberately **no `BD_BDP_TOKEN`**: an
env-carried secret is inherited by every child process (git hooks, dolt
subprocesses) — the same reason `remote_migrate_gate.go` keeps a
programmatic twin of `BD_ALLOW_REMOTE_MIGRATE`.

**Credential lookup** (client): `BEADS_BDP_TOKEN_FILE` > `bdp.token_file` >
credentials file section **`[bdp <origin><scope-path>]`** with `token=`
(distinct from the Dolt `[host:port]` `password=` sections — ruling 7a
allows several path-distinguished Scopes on one host, so a host-keyed
section would send one Scope's token to another). Transport rules: no
redirects followed (a 3xx is an error naming the location, never
re-sent with the credential); `Authorization` is sent only to the
configured origin.

**Precedence, per command** (a table test pins it — `bdp.*` gets its own
order rather than an inherited claim):

| Command | Resolution |
| --- | --- |
| `bd init --bdp-server <url>` | flag > `BD_BDP_*` env > existing `config.yaml`; writes `bdp.client: server` and `bdp.server` to `config.yaml`; touches nothing else |
| `bd config set bdp.<key>` | validates, writes `config.yaml` (yaml-only routing) — effective for the next command, no other file to outrank it |
| every `bd bdp` verb | `BD_BDP_*` env > `config.yaml`; `bdp.client` decides the route, `bdp.server` the target |
| `bd bdp status` | prints the resolved route, target, token source (never the token), and whether `insecure_http` is in force |

**In `client: server` mode** the verbs' `openGraph*()` accessors return a
BDP-client realization of the same `graphops` role interfaces (A5), so a
verb cannot tell which route it took. A workspace in this mode still has a
local graph store (A1 ran); `bd bdp serve` there **refuses** unless
`--serve-local-store` is passed, so an operator cannot serve a dead local
copy beside the designated server by accident. Issue verbs are unaffected
in either mode.

### A3. `bd bdp serve` — serving a Scope (rulings 7b, 9, 12; amendments A1, A2)

`bd bdp serve` is a **thin command over the existing `internal/httpapi`
server**: the same storage classification, the same hook peel, the same
`httpapi.Config` — plus a populated `Graph` field. It differs from
`bd serve` in one policy: it **requires a Scope** (exit 2 otherwise), where
`bd serve` mounts the BDP rows only when `bdp.scope_url` is configured and
serves exactly what it serves today when it is not (ruling 12: "no URL → no
BDP routes"). W2 decides whether `bd bdp serve` survives as the strict alias
once every configured workspace serves BDP from `bd serve`; nothing needs
folding later because there is one server.

```text
bd bdp serve [--addr IP:PORT]            default 127.0.0.1:0 (ephemeral), numeric IP only
             [--allow-non-loopback]      requires --auth-token-file (or --insecure-no-auth)
             [--auth-token-file PATH]    bearer tokens, one per line, hot-reloaded
                                         (env BEADS_SERVE_TOKEN_FILE — serve's own; no new env)
             [--insecure-no-auth]
             [--allowed-host NAME]...    extra exact Host allowlist entries
             [--scope-url URL]           see "identity" below
             [--dev-local-test]          ephemeral in-memory identity; never persisted
             [--serve-local-store]       permit serving in a client: server workspace
```

The five posture flags are `bd serve`'s **variables**, not copies: the
command reuses `serveAddr`/`serveAllowNonLoopback`/… and the same
validators (`ValidateBindAddr`, `ValidateAuthPosture`,
`ValidateAllowedHost`, `NewTokenFileAuth`), and `serveListen` builds the
server. Behavior, in order:

1. **Classification is `serveDatabaseSource`, verbatim:** registered
   backend → store source; embedded Dolt → **refused, permanently**, with
   serve's typed `errServeEmbedded` (its commit protocol cannot meet the
   per-request atomicity contract); otherwise the unit-of-work provider —
   which is therefore BDP's primary production path (architecture §5).
2. **Roles from beneath the hook layer** — `serveIssueRoles`' one peel. For
   the graph roles this is a no-op today (every graph hook wrapper recurses
   unwrapped, B5); it is kept for uniformity with serve and as the guard a
   future graph hook vocabulary would need.
3. **Capability probe, all-or-nothing:** `GraphReader()`, `GraphTypes()`,
   `GraphIdentityReader()`, `GraphScopeBootstrapper()` are taken from the
   source; any `*storage.ErrUnsupported` → **exit 2, typed** ("this
   backend does not implement the graph contract"); any *other* error is
   an operational failure and aborts startup as serve's role binding does
   (unsupported and broken never collapse into each other). `bd serve`
   under the same probe: `ErrUnsupported` → BDP rows absent, legacy
   routes exactly as before; operational error → startup failure.
4. **Identity** — the state table (each row a contract case, B7):

   | Persisted Scope | Configured URL | This clone's authority half | Result |
   | --- | --- | --- | --- |
   | none | none | — | `bd bdp serve`: exit 2 (ruling 7a: an explicit URL is required); `bd serve`: no BDP rows |
   | none | `bdp.scope_url` or `--scope-url` | — | **first serve:** `Mint` under that URL (persisted — a `--scope-url` used here *is* the identity from now on) → grant this clone the authority half → serve honestly empty. Under `--readonly`: **refuse** with the `ErrNoScope` guidance (a mint is a write; serve's own `errServeReadonly` posture) |
   | present | absent or **different** | any | refuse: "configured URL ≠ persisted Scope" (a `--scope-url` on a minted store must equal the persisted URL) |
   | present, matches | matches | **absent** (a clone, or a restored database — the ignored table arrives empty) | refuse with `ErrNotAuthority` and the guidance: `bd bdp promote` (take over) or `bd bdp restore` (after a restore) |
   | present, matches | matches | present, epoch **stale** (promoted elsewhere, pulled) | refuse, same guidance |
   | present, matches | matches | present, epoch matches | serve; the expectation `{ScopeURL, AuthorityID, Epoch}` is handed to `httpapi.GraphConfig.Expect` and **re-asserted inside every read transaction** (amendment A1) |
   | any | any | `unverified` flag set (written by `bd restore`, A4) | refuse until `bd bdp restore` clears it |

   `--dev-local-test` (ruling 7a's development mode): a `local-test` URL
   derived from the listener, served from an **in-memory** identity that is
   never written; it refuses to start if a persisted Scope exists (dev mode
   is for empty stores) and refuses `--allow-non-loopback`. (Council
   disagreement resolved for 7a: Gemini's "persist an ephemeral epoch" would
   make the derived URL an identity; 7a says never.)
5. **Host policy:** the Scope URL's host joins the allowlist automatically
   (it is operator-configured identity), because BDP clients dial the Scope
   URL, and the rebinding defense otherwise refuses every real request. The
   listener is plaintext behind TLS termination; the Scope URL's `https`
   scheme never matches the socket, and the doc says so where the flag is.
6. **Mounting:** `serveListen(opts, httpapi.Config{…, Graph:
   &httpapi.GraphConfig{Reader, Types, Expect, ScopeURL}})`. `Listen`'s
   `checkDatabaseSource` extends to the graph roles: same source as the
   issue roles, never hook-firing. Liveness stays `GET /healthz`; BDP
   readiness is a real discovery read.
7. **Lifecycle:** `bd bdp serve` joins `runsPostCommandMaintenance`'s
   exclusion by name (a server is not a command; no auto-commit/push/export
   from a signal handler) and serve's events-journal maintenance ticker
   applies as it does to `bd serve`.

`--readonly` + minted Scope serves (v0 has no graph writes); `--readonly` +
unminted refuses (row 2). When P3 adds writes, `bd bdp serve` adopts
serve's whole-surface refusal for `--readonly`, stated in P3's slice.

### A4. `bd bdp promote` / `bd bdp restore` / `bd bdp ledger` (rulings 9, 11; amendment A5)

These are the **only** reachers of `GraphAuthorityRotator()`; the server
assembly has no field for it (architecture §4).

- **`bd bdp promote`** — make *this clone* the authority for the Scope it
  carries: a CAS on the replicated Scope row (`epoch = epoch + 1 WHERE
  epoch = <read>`; a lost race is a typed refusal, retry by re-reading) and
  a write of the clone-local authority half `{authority_id, epoch}`. The
  old authority is fenced **by the expectation check**: its next read after
  it pulls the new epoch fails `assertAuthorityInTx`; until it pulls, it
  serves a history that is now a fork — which is why `promote` prints the
  old authority's id and the operator's obligation to stop it (v0 has no
  network fencing; the ADR in architecture §7 owns "refuse foreign
  deltas"). `--rotate-url <new>` additionally rotates the URL when history
  diverged (operator-confirmed).
- **`bd bdp restore`** — runs after a database restore. `bd restore`
  (`cmd/bd/backup_restore.go`) writes the clone-local **`unverified`**
  flag; `bd bdp serve` refuses while it is set (A3 row 8). `bd bdp restore`
  then branches on the provider's declaration:

  | `LedgerDurability` (declared by `GraphIdentityReader`) | Meaning | `bd bdp restore` does |
  | --- | --- | --- |
  | `in-state` (Dolt, v0) | the ledger is ordinary versioned state; a restore carries it only to the restored point | requires a ledger export newer than the restored watermark (`--ledger <file>`) to show continuity; **otherwise rotates** the Scope URL (operator supplies the new one) and epoch, records the old URL refused-forever, and re-grants the authority half |
  | `independent` | the provider keeps the ledger outside the restored state and proves it (contract case) | re-validates and re-grants; no rotation |
  | `none` | the provider keeps no ledger | always rotates |

  Never silent: every branch prints what it did and why.
- **`bd bdp ledger export [--since SEQ]`** / **`bd bdp ledger import
  <file>`** — the continuity lane ruling 11 needs and v1 lacked: the
  allocation/tombstone ledger as JSONL with a monotonically increasing
  `seq` watermark, and the refused-URL history. Import is
  append-if-absent, keyed by `(path, seq)`; it never rewinds a tombstone.
  Operators who want restore-without-rotation schedule the export beside
  their backups; a restore with a newer export than the restored watermark
  proves continuity.

### A5. Graph verbs (v0 reads; P3 writes; names fixed now)

```text
bd bdp bead get <path> | list [--type URL] [--after CURSOR] [--limit N]
bd bdp link get <path> | list [--type URL] [--source PATH] [--target REF] [--after CURSOR] [--limit N]
bd bdp types [get <url>]
bd bdp status
bd bdp serve | promote | restore | ledger export|import
(P3)  bd bdp bead create|update|delete ; bd bdp link create|update|delete
```

Each verb reaches its role through an accessor and nothing else, on
whichever route the workspace is on — the `cmd/bd/label.go` pattern plus
the third route, with a route-fork test in the shape of
`cmd/bd/vc_recompute_test.go`:

```go
func openGraphReader() (graphops.Reader, error) {
    if bdpClientMode() == "server" {        // config.yaml / BD_BDP_* per A2
        return bdpclient.GraphReader(bdpClientConfig())   // internal/bdpclient
    }
    if usesProxiedServer() { return proxiedGraphReader() }
    return store.GraphReader()
}
```

`internal/bdpclient` speaks the pinned wire (`bdpwire`, B8) and implements
`graphops.Reader` and `graphops.DescriptorReader` over it; a BDP Problem
maps back to the same typed graph errors the store bodies raise (round-trip
test: every Problem in the pinned vocabulary → `errors.Is` holds across the
fork, gone-family included). Pagination through the client is
cursor-faithful: the CLI never re-sorts.

## Part B — Storage interfaces

### B1. Accessors on `storage.Storage` (the house rule — architecture §2a)

```go
// GraphReader returns the guarded graph-read surface for this store: Bead
// and Link records (a Bead with its complete, bounded ownedLinks, assembled
// in the call's one transaction), keyset selections in code-unit order, and
// incident Links. Every method asserts the caller's AuthorityExpectation
// inside that transaction. Reads fire no hooks; the hook decorator recurses.
GraphReader() (graphops.Reader, error)

// GraphTypes returns the Type Descriptor inventory (ordered, bounded, keyed).
GraphTypes() (graphops.DescriptorReader, error)

// GraphTypeInstaller returns the descriptor install/converge role used by
// bd init's bootstrap and the conformance fixture. A caller entitled to read
// the catalog is not thereby entitled to change it.
GraphTypeInstaller() (graphops.TypeInstaller, error)

// GraphIdentityReader returns the Scope row, this clone's authority half,
// and the provider's LedgerDurability declaration.
GraphIdentityReader() (graphops.IdentityReader, error)

// GraphScopeBootstrapper returns the one-time mint. The server assembly may
// hold it on the first-serve path and nothing else.
GraphScopeBootstrapper() (graphops.ScopeBootstrapper, error)

// GraphAuthorityRotator returns promote/rotate. Reached only by the
// bd bdp promote|restore composition root; never by a server.
GraphAuthorityRotator() (graphops.AuthorityRotator, error)

// P3: GraphWriter() (graphops.Writer, error)
```

Added to **`storage.Storage`** (verified: all 28 accessors live there;
`DoltStorage` embeds it; both decorators embed `DoltStorage`). **Every
decorator and provider wrapper declares each accessor explicitly** — hook
layer, telemetry, the notifying UOW provider, `timedProvider` in `httpapi`,
`serveRoleSource` and its test stubs, `internal/jira/tracker_test.go`'s
`configStore` stub, and every other `Storage` implementer the censuses
enumerate. Promotion through the embedded interface compiles and is the
failure mode; the three censuses (B5) fail the build that relies on it.
**This is a breaking change for out-of-tree backends.** `backend/backend.go`
promises a CHANGELOG call-out; this slice writes it (there is no `Memories()`
precedent entry to copy — v1's claim was wrong).

### B2. The `graphops` leaf (public, repo root; amendment A4)

```go
package graphops   // imports: stdlib, beadserrors — nothing else

// ---- requests and results (the leaf doc specifies each field; contract cases cite it by line)
type AuthorityExpectation struct{ ScopeURL string; AuthorityID string; Epoch uint64 } // zero value = "no expectation" is REFUSED by every body
type BeadRequest   struct{ Path string; Expect AuthorityExpectation }
type LinkRequest   struct{ Path string; Expect AuthorityExpectation }
type SelectRequest struct{
    TypeURL string      // optional exact filter
    After   string      // keyset cursor: the last path of the previous page ("" = start)
    Limit   int         // default 50, max 200; > max → ErrValidation
    Expect  AuthorityExpectation
}
type IncidentRequest struct{ Path string; Direction Direction /* In | Out | Both */; After string; Limit int; Expect AuthorityExpectation }
type BeadRecord struct{ Bead Bead; OwnedLinks []Link }   // complete per declared owned Type; ascending code-unit order by Link path;
                                                         // a declared owned Type with no Links is PRESENT as an empty entry
type BeadPage struct{ Items []BeadRecord; Next string }  // Next == "" means exhausted; items carry ownedLinks (page × ΣMax bound)
type LinkPage struct{ Items []Link; Next string }

type Reader interface {
    Bead(ctx, BeadRequest) (BeadRecord, error)
    Link(ctx, LinkRequest) (Link, error)
    Beads(ctx, SelectRequest) (BeadPage, error)       // WHERE path > After ORDER BY path LIMIT n, on the binary-collated column
    Links(ctx, SelectRequest) (LinkPage, error)
    IncidentLinks(ctx, IncidentRequest) (LinkPage, error) // two index scans (source, target) merged in code-unit order
}
type DescriptorReader interface {
    Descriptors(ctx) ([]TypeDescriptor, error)        // ordered by URL; bounded by MaxCatalog (1,000)
    Descriptor(ctx, url string) (TypeDescriptor, error)
}
type TypeInstaller interface {
    Install(ctx, InstallRequest) (InstallResult, error) // idempotent by fingerprint; closure validated; an owning
}                                                        // declaration without Max → ErrValidation; over MaxCatalog → ErrValidation
type IdentityReader interface {
    Read(ctx) (ScopeIdentity, error)                  // Scope row + this clone's authority half (Granted bool, Epoch); ErrNoScope when never minted
    LedgerDurability(ctx) (LedgerDurability, error)   // in-state | independent | none — the provider's declaration (ruling 11)
}
type ScopeBootstrapper interface {
    Mint(ctx, MintRequest) (ScopeIdentity, error)     // INSERT into the singleton row; loses the race → ErrScopeExists
}
type AuthorityRotator interface {
    Promote(ctx, PromoteRequest) (ScopeIdentity, error) // CAS on epoch; writes this clone's authority half
    Rotate(ctx, RotateRequest) (ScopeIdentity, error)   // new URL; old URL → refused-forever history
}
```

**Bounds.** `MaxExpandedRows` (10,000) caps `len(page) + Σ ownedLinks` for
any read; over it → `ErrRepresentationTooLarge` (typed; the handler maps it
to the pinned Problem). Owning Types **must** declare `Max`; the installer
refuses one that does not (W3's generator emits it). Reads are **one page
query plus one batched owned-links query** (`source_path IN (page) AND
type_url IN (owned types)`), grouped in Go — never N+1 inside the held
transaction; a query-count test pins it.

**Values** (`Bead`, `Link`, `Ref`, `Properties`, `Revision`, `Attribution`,
`TypeDescriptor`, `OwnedLinkDecl`, `ScopeIdentity`) have unexported fields
and constructors that enforce the laws in `laws.go` (same package).
`Properties` is the immutable raw-JSON object value from the plan
(duplicate-key rejection, number preservation, deterministic encoding, RFC
6902 §4.6 equality); its canonical bytes are what B4 stores. `Ref` is a
sum: in-Scope (`Path`) or external (`URL`), with the discriminant stored
(B4) so rotation never confuses the two. `Revision` is 128 bits from
`crypto/rand`, lower-hex; minted by the write body, unique per resource by
construction (a collision on re-mint is retried, never returned).

**Errors:** `beadserrors` (stdlib-only) declares the new sentinels
(`ErrNoScope`, `ErrScopeExists`, `ErrNotAuthority`, `ErrURLReused`,
`ErrRepresentationTooLarge`) and the typed `GoneError{Path, State}` (state
`pruned` | `erased`, from the ledger's tombstone row — what the handler
needs to emit `resource-pruned` / `resource-erased`); `graphops` aliases
them so `errors.Is` crosses the module boundary.

### B3. Bodies and legs

`internal/storage/graphops` bodies take **`DBTX`** (the `issueops.DBTX`
shape, declared locally: `ExecContext`/`QueryContext`/`QueryRowContext`),
which `*sql.Tx` and `domain/db.Runner` both satisfy — the decision the
guide's mechanical test gives at *design* time (the `MetadataCAS`/
`TreeWalker` precedent), not "per method at implementation time" (v1).

| Leg | Files | Body |
| --- | --- | --- |
| server Dolt | `internal/storage/dolt/graph_reader.go`, `graph_types.go`, `graph_identity.go` | accessor wraps the body in `withReadTx` (reads) / `withRetryTx` (install, mint, promote, rotate) |
| embedded Dolt | `internal/storage/embeddeddolt/graph_*.go` (`//go:build cgo`) | same body, `withConn` |
| unit of work | `internal/storage/domain/graph.go` (`GraphRepository` interface), `internal/storage/domain/db/graph_repository.go` (over `db.Runner`, delegating to the bodies), `internal/storage/uow/graph_*.go` (`uow.UnitOfWork.Graph()`; provider accessors via `RunTxRead`; mint/promote via `RunTxResult` with commit messages `bdp: mint scope <url>` / `bdp: promote epoch <n>`; a no-op result commits nothing) | **same body** |

Contract headers therefore say **"one reading plus an engine check"** for
every graph role. Every read body begins with `assertAuthorityInTx(ctx, tx,
req.Expect)`: it reads the singleton Scope row and the clone-local authority
row in the same transaction and returns `ErrNotAuthority` on any mismatch
(a zero expectation is refused, so no caller can opt out). `SeedBeadInTx`/
`SeedLinkInTx` (P1 fixture writer) allocate through the ledger like a real
write; they are reachable only from `backend/conformance`'s `GraphFixture`
hook.

### B4. Schema (migrations; replicated vs clone-local; frozen once merged)

Rules the tree enforces and this schema obeys: migrations are **frozen once
merged** and content-hashed across clones (`check-migration-hygiene.sh`
check C); **no `NOW()`/`UUID()`/`RAND()`** in migration SQL (check B) — every
timestamp and id is set in Go; a clone-local table lives in the
`migrations/ignored/` series with its `dolt_ignore` entry before creation
(0028); real-Dolt tests for anything a `sqlmock` echo cannot exercise
(`BEADS_TEST_EMBEDDED_DOLT=1 -tags cgo`); DDL is not transactional across
statements, so each `CREATE` is guarded and resumable.

**Collation.** No migration in the tree declares one; Dolt's default
`utf8mb4` collation is case- and accent-insensitive, which would make
`ORDER BY path` not code-unit order and `Foo`/`foo` one primary key. Every
identifier column below is **`VARCHAR(n) CHARACTER SET utf8mb4 COLLATE
utf8mb4_bin`** (written `BIN` in the table), with a contract case asserting
two paths differing only in case are distinct and sort by code unit.

**Identity is Scope-relative.** Rows store the canonical Scope-relative
`path`; the absolute URL is `scope_url + path`, computed at the boundary
from the live Scope row — so a URL rotation (ruling 11) rewrites no rows,
and code-unit order of paths equals code-unit order of absolute URLs under
one Scope prefix. External references keep their absolute URL with a
discriminant.

**JSON is bytes.** `properties` and `descriptor` are canonical JSON bytes in
`LONGBLOB`, never the engine `JSON` type: the tree measured that column
decoding through float64 (`1.0`→`1`, integers past 2^53 rounded, `-0.0`→`0`,
`1e300` expanded — `metadata_cas.go`, guide §"a precision the substrate does
not keep"), which would break the no-op gate and descriptor fingerprints.
Round-trip tests cover each measured case. Size limit 1 MiB per value
(`ErrValidation` above it).

| Table | Series | Columns (type; nullability) | Keys / constraints |
| --- | --- | --- | --- |
| `graph_scope` | replicated | `id TINYINT NOT NULL` (always 1), `scope_url VARCHAR(2048) BIN NOT NULL`, `authority_id CHAR(32) NOT NULL`, `epoch BIGINT UNSIGNED NOT NULL`, `minted_at DATETIME(6) NOT NULL` | `PRIMARY KEY (id)`, `CHECK (id = 1)` — **singleton**: `Mint` is an INSERT that loses the race |
| `graph_scope_history` | replicated | `scope_url VARCHAR(2048) BIN NOT NULL`, `refused_at DATETIME(6) NOT NULL`, `reason VARCHAR(64) NOT NULL` | `PRIMARY KEY (scope_url)` — refused-forever URLs (also carried by the ledger lane) |
| `graph_authority_local` | **ignored** (clone-local) | `id TINYINT NOT NULL`, `authority_id CHAR(32) NOT NULL`, `epoch BIGINT UNSIGNED NOT NULL`, `granted_at DATETIME(6) NOT NULL`, `unverified TINYINT(1) NOT NULL DEFAULT 0` | `PRIMARY KEY (id)`, `CHECK (id = 1)`; `dolt_ignore` entry precedes creation — a clone or a restore **arrives without this row** |
| `graph_beads` | replicated | `path VARCHAR(1024) BIN NOT NULL`, `type_url VARCHAR(2048) BIN NOT NULL`, `revision CHAR(32) NOT NULL`, `attribution_principal VARCHAR(512) NULL`, `attribution_status ENUM('claimed','unknown') NULL`, `properties LONGBLOB NOT NULL`, `created_at DATETIME(6) NOT NULL`, `updated_at DATETIME(6) NOT NULL` | `PRIMARY KEY (path)`; `INDEX (type_url, path)`; `FOREIGN KEY (type_url) REFERENCES graph_type_descriptors(url)`; attribution columns both NULL or both set (`CHECK`) |
| `graph_links` | replicated | `path VARCHAR(1024) BIN NOT NULL`, `type_url … BIN NOT NULL`, `revision CHAR(32) NOT NULL`, `source_path VARCHAR(1024) BIN NOT NULL`, `source_pin CHAR(32) NULL`, `target_kind ENUM('in','ext') NOT NULL`, `target_path VARCHAR(1024) BIN NULL`, `target_url VARCHAR(2048) BIN NULL`, `target_pin CHAR(32) NULL`, `attribution_*`, `properties LONGBLOB NOT NULL`, timestamps | `PRIMARY KEY (path)`; `INDEX (source_path, type_url, path)` (ownedLinks assembly), `INDEX (target_path, path)` (incident); `FOREIGN KEY (source_path) REFERENCES graph_beads(path)` (a Link's source is in-Scope — law); `CHECK` exactly one of `target_path`/`target_url` per `target_kind`; **no** uniqueness on (type, source, target) — multiplicity is law |
| `graph_type_descriptors` | replicated | `url VARCHAR(2048) BIN NOT NULL`, `descriptor LONGBLOB NOT NULL`, `fingerprint CHAR(64) NOT NULL`, `installed_at DATETIME(6) NOT NULL` | `PRIMARY KEY (url)`; `UNIQUE (fingerprint)` |
| `graph_allocations` | replicated | `path VARCHAR(1024) BIN NOT NULL`, `kind ENUM('bead','link') NOT NULL`, `seq BIGINT UNSIGNED NOT NULL`, `birth_authority_id CHAR(32) NOT NULL`, `birth_epoch BIGINT UNSIGNED NOT NULL`, `allocated_at DATETIME(6) NOT NULL`, `state ENUM('live','pruned','erased') NOT NULL`, `tombstoned_seq BIGINT UNSIGNED NULL`, `tombstoned_at DATETIME(6) NULL` | `PRIMARY KEY (path)` — the O(1)/O(log n) ID test (ruling 3); `UNIQUE (seq)`; insert-once, **update-on-tombstone** (v1's "append-only" was wrong: the append-only property lives in the export lane's `seq` watermark, A4) |

`updated_at` is protocol-irrelevant bookkeeping; revisions are minted by
the write body, never derived from timestamps. The bead `type_url` and the
link `source_path`/`target_*` columns are immutable after insert (enforced in
the write body; P3 adds a trigger-free check via the post-merge validator).

**Cross-repo coupling (bts).** `DoltTeamServer` workspaces refuse to open
when `current < latest` with "ask your operator to run `bts migrate`" and
**no `BD_IGNORE_SCHEMA_SKEW` hatch** (`internal/storage/uow/team_server_schema.go`).
Six new migrations (five replicated, one ignored) mean every bts-managed
workspace refuses the next bd release until bts ships content-hash-identical
copies. The migration PR is sequenced with bts and says so in its
description; the remote-migrate gate (#4259) forces migrate-vs-adopt on
every remote-backed workspace at upgrade as it does for any migration.

### B5. Decorators, censuses, and every embedding surface

- `internal/storage/hook_graph_*.go` (six files): declared, recurse
  **unwrapped** — the hook vocabulary (`on_create/on_update/on_close`)
  hands scripts an *issue*; a graph hook vocabulary is a separate proposal.
  `storage.RoleFiresHooks` gains a `false` row per graph role (so
  `checkDatabaseSource` can classify them). Added to
  `role_accessor_decorator_test.go`'s wrapped/unwrapped table and to the
  storage file's pass-through paragraph.
- `internal/telemetry/graph_*.go`: every method spanned `storage.op` /
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
  provider (parity test), `internal/httpapi/claim.go`'s `timedProvider`
  (per-request accessors so units of work land in `uow_ms`), `cmd/bd/serve.go`'s
  `serveRoleSource` and its non-embedding test stubs,
  `internal/jira/tracker_test.go`'s `configStore` stub — every surface that
  embeds the store or a provider, enumerated by `grep -l 'func (.*) Memories()'`
  at implementation and listed in the slice's PR.

### B6. `backend/` public surface and depguard

- **No aliases** (amendment A4; architecture §2a): `graphops` is public and
  imported directly, like `issueops`. `TestPublicSurfaceComplete` stays
  green *because* no `internal/` type is reachable from the new accessors —
  a test asserts that (a graph accessor leaking an internal type would fail
  it, which is the guard).
- `backend/backend.go`'s "minimal external backend" example gains the six
  accessors; its stability note's promised **CHANGELOG entry** is written
  (`BREAKING (published backend package): six graph accessors are required
  on Storage`).
- `.golangci.yml` `cmd-bd-role-constructors` gains an explicit deny for
  `github.com/steveyegge/beads/internal/storage/graphops` ("a command
  reaches graph roles through the store accessors, never the tx-level
  bodies"), with a mutation test that removes the entry and expects the
  lint to pass — proving the entry is load-bearing.

### B7. Conformance

- Families: `graph_reader_contract.go`, `graph_types_contract.go`
  (reader + installer), `graph_identity_contract.go` (reader,
  bootstrapper, rotator), each citing the leaf doc by line.
- `RoleContractBundle` gains `GraphReader`, `GraphTypes`,
  `GraphTypeInstaller`, `GraphIdentityReader`, `GraphScopeBootstrapper`,
  `GraphAuthorityRotator` factory fields **and** their rows in
  `role_bundle_cases.go` (`TestRoleContractCasesMatchTheBundleFields`);
  `GraphFixture` carries the seed hook (`SeedBead`/`SeedLink` → the
  ledger-enforcing InTx writers).
- Wirings in `internal/storage/{dolt,embeddeddolt,uow}/graph_*_contract_test.go`;
  the leg registry (`contract_leg_registry_test.go`) and
  `TestEveryLegWiresEveryRoleContract` (`leg_contract_wiring_test.go`) both
  see them; both coverage gates apply (`TestEveryRoleMethodHasAContractCase`,
  `unwiredContractEntrypoints`); waivers only with a reason, only until wired.
- Non-capable stores answer `*storage.ErrUnsupported{Op: "<accessor
  name>"}` — `RunUnsupportedContract` compares `Op` to the method name
  exactly, so the six strings are pinned — proven per accessor.
- Cases the council asked for by name: a clone produced by push/pull
  refuses (`ErrNotAuthority`); concurrent first-serve mint (one wins);
  rotation refuses the old URL; case-differing paths are distinct and
  code-unit ordered; ownedLinks completeness incl. declared-empty entries
  and the page × ΣMax bound with `ErrRepresentationTooLarge`; keyset
  continuation inside one transaction; gone-family from tombstone state;
  expectation mismatch mid-process (promote elsewhere between two reads);
  `LedgerDurability` answer per provider; installer refuses an owning
  declaration without `Max`; query-count = 2 for a page read.
- Differential-gate rows: every legacy form of `bd link`, `bd graph`,
  `bd graph check`, `bd restore`, `bd promote` parses and behaves as before.

### B8. `httpapi` integration and the pinned wire (amendment A2)

- **`httpapi.Config.Graph *GraphConfig`** — `{Reader graphops.Reader; Types
  graphops.DescriptorReader; Expect graphops.AuthorityExpectation; ScopeURL
  string}`; nil = no BDP rows. On the provider arm the roles are rebuilt per
  request through `timedProvider` like every other role. `checkDatabaseSource`
  refuses a graph role from a different source than the issue roles, and a
  hook-firing one. There is deliberately **no rotator field**: a test asserts
  the struct cannot carry one.
- **`bdpRouteTable`** (`internal/httpapi/bdp_routes.go`): rows in the same
  `route` shape (op, method, pattern, `projectExempt: false` — an absent
  stamp passes and BDP clients never send one; `authExempt: false`;
  semaphore taken), registered by `Server.handler()` **only when
  `cfg.Graph != nil`**, each wrapped by the same `s.route(rt)` — so
  deadline, bearer-before-semaphore, stamp-behind-auth, and the never-log-
  the-token rule apply by construction. v0 P1 rows: discovery, `types/`,
  one Bead, one Link; **collection rows wait for P2's cursor ADR**
  (architecture §7). BDP rows contribute **no token** to
  `ContextResponse.capabilities` in v0 (BDP has its own discovery; W2 may
  add a behavior token); `TestSpecRouteParity` continues to compare only
  `routeTable` against `openapi.v0.yaml`, and a sibling parity test compares
  `bdpRouteTable` against the pinned schema's path grammar.
- **Posture parity test:** one refusal matrix (missing / malformed / unknown
  token, bad `Host`, saturated semaphore, deadline) drives a legacy row and
  a BDP row and asserts identical status and log shape.
- **Handler = serializer** (`bdp_handlers.go`): `graphops` records → `bdpwire`
  DTOs; typed graph errors → BDP Problem records (`bdp_problem.go`), the
  closed vocabulary, here and only here.
- **Wire:** `internal/httpapi/bdpwire/schema/bdp-v0.schema.json` is
  **vendored** with a `PROVENANCE` file (upstream repo, commit — the plan's
  §0 pin — and sha256); `make bdp-gen` runs a **pinned** JSON-Schema→Go
  generator (version recorded in `Makefile`; if the pinned schema exceeds
  what it can express, P0 records the fallback: hand-written DTOs validated
  against the schema in tests); `make bdp-check` regenerates and diffs, and
  `scripts/ci/pr-policy.sh` runs it beside `make api-check`. The OpenAPI
  generator and document are untouched.

## Part C — What does not change

`storage.Storage`'s existing 28 accessors and every `issueops` role; the
journal's frozen vocabulary; `openapi.v0.yaml` and `TestSpecRouteParity`;
`bd serve` on a workspace with **no** `bdp.scope_url` (no rows, no
capability change, byte-identical); every legacy CLI verb (differential
gate rows in B7); JSONL export shapes; `metadata.json`'s schema.

## Part C2 — What changes that v1 claimed did not

- **Merge and sync entry points.** `MergeAndSettle`,
  `MergeAndSettleWithStrategy`, `MergeWithStrategy`, and plain `Merge` all
  funnel through a **post-merge graph validator** that rejects (rolls back)
  a graph delta failing the invariants or carrying a foreign
  `birth_authority_id`/epoch; clean-merge early returns included. The
  replication/merge ADR (architecture §7) specifies it before the
  migrations land.
- **Federation.** Graph tables ride filtered pushes **unfiltered, by
  decision, in v0** (identity, descriptors, allocations, beads, links); a
  per-topology filter hook is a later ruling, and a filter that drops an
  endpoint must drop the Link (never a dangling edge).
- **`bd sql` and raw SQL.** Documented as **out of contract** for graph
  tables (the command already says it bypasses storage); the validator above
  catches what merges import, not what a local `bd sql` writes — the
  enforcement-boundary ruling (architecture §2b) decides whether DB
  privileges or triggers close that gap before P3.
- **`bd restore`** writes the clone-local `unverified` flag (A4).

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

## Part E — Proposed ruling amendments this spec assumes (pending)

A1 per-call transaction with in-transaction authority expectation is the v0
lease (ruling 9); A2 BDP rows mount inside `httpapi`, `bd bdp serve` is a
thin strict command (rulings 7b/12); A3 `bd bdp …` namespace; A4 values,
laws, and roles in public `graphops`, no `backend/` aliases; A5 ruling 11's
mechanism = ignored-table authority half + ledger export/import lane +
provider `LedgerDurability` declaration; A6 `config.yaml` as the single
local source for `bdp.*`. Plus two decisions the plan does not yet hold: the
out-of-role DML enforcement boundary, and the replication/merge ADR as a P1
gate. Full text: architecture §2b.
