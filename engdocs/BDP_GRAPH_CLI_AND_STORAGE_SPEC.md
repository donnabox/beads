# BDP graph store — CLI and storage-interface changes, in detail

**Status:** Draft v4 (W-arch, after three council rounds) — feat/bead-graph
**Date:** 2026-09-02
**Companions:** `BDP_BEAD_GRAPH_PLAN.md` (rulings), `BDP_GRAPH_ARCHITECTURE.md`
(shape; its §2b lists the proposed ruling amendments A1–A8 this spec
assumes — **none is ruled yet**, and every sentence that depends on A8's
option A says so). This document is the *diff*: every command, flag, config
key, interface member, package, migration, and gate the graph work adds or
touches — and, just as precisely, what it does not touch (Part C) and what
it changes that an earlier draft claimed it did not (Part C2). Phase
markers follow the plan's §7: **P0** contracts and wire, **P1** storage,
**P2** serving, **P3** writes.

## Part A — CLI

All graph-store commands live under one root, **`bd bdp`** (amendment A3):
`bd link`, `bd graph`, `bd restore`, and `bd promote` are existing verbs
with positional arguments, and the plan's constraint #1 forbids changing
them. The differential gate gains one row per legacy form of each (Part B7).

**Root store policy is keyed by command path for this subtree, and it is
authoritative.** The root command classifies commands by *leaf name* at
several call sites — `effectiveRootStorePolicy(cmd.Name(), …)`,
`runsPostCommandMaintenance(cmd.Name(), …)` (`cmdName == serveCmdName`),
`isReadOnlyCommand` (`readOnlyCommands` contains `list`),
`shouldAutoPruneEventsJournal` in `cmd/bd/events_journal.go`, and the
`cmd.Name() != "import"` branches — and the tree records that same-named
subcommands collide. Leaf names under `bd bdp` **may** coincide with those
lists (`bd bdp bead list` is a `list`; `bd bdp serve` is a `serve`); the
coincidence never governs. A `bdpRootPolicy` table keyed by
`CommandPath()` (the `CheckMigrationFreeze(cmd.CommandPath()…)` precedent)
is consulted first at each of those call sites for any path under
`bd bdp`; the implementation shape is one `commandPolicy(*cobra.Command)`
that every consumer asks, with an exhaustive Cobra-tree test that walks the
`bd bdp` subtree and fails on an unlisted leaf or a duplicate leaf name
resolving differently from its path.

| Verb | Local store | Maintenance | Note |
| --- | --- | --- | --- |
| `bd bdp bead get\|list`, `link get\|list`, `types`, `status` | opens **read-only** when `bdp.client: store`; **skipped entirely** when `bdp.client: server` (no schema check, no auto-start, no migration prompt — client verbs never touch the local database) | no | `bdp.client` is read from `config.local.yaml`/`config.yaml` before any store opens — the reason it is yaml-class (A6) |
| `bd bdp serve` | serve's own classification (A3) | no | |
| `bd bdp client` | none (writes `config.local.yaml`) | no | |
| `bd bdp promote`, `restore`, `ledger snapshot\|apply` | opens writable, always local | no (admin verbs commit — and where the topology requires, push — explicitly) | |

### A1. `bd init` — graph store initialization (ruling 12)

`bd init` initializes the graph store with everything else it initializes,
against the normalized storage interfaces:

1. Runs the graph migrations (Part B4): the replicated series in the normal
   migration series, and the dolt-ignored `graph_authority_lease` table in
   the `migrations/ignored/` series with its `dolt_ignore` entry registered
   before creation (the 0055 `leases` precedent).
2. **Bootstraps the built-in Type Descriptor catalog through
   `BeadGraphTypeInstaller()`** — idempotent and fingerprint-keyed, so
   re-init converges; W3 supplies the inventory, and until then the catalog
   is honestly empty. **Only when this workspace may:** on an unminted store
   (no Scope row — nothing to assert), or on the authority workspace (a
   consistent witness). On a non-authority clone of a served project the
   bootstrap is **skipped** and init succeeds — the catalog is the
   authority's and this clone converges by pull; re-init is the documented
   repair path (it is what appends new `.beads/.gitignore` patterns) and
   must keep working on every clone. A provider answering
   `*storage.ErrUnsupported` also skips it and init succeeds, **silently
   at the default verbosity** (a debug-level line only) so legacy output is
   byte-identical; any other error fails init like any other bootstrap
   step.
3. Writes **no Scope identity** and no witness file — a workspace has a
   graph store from init but not a Scope; the Scope is minted on first serve
   (A3). Nothing is written to `metadata.json` for the graph (A6).
4. Ensures the workspace's `.beads/.gitignore` carries the witness-file
   entry (`EnsureGitignoreForBeadsDir`; the template, `requiredPatterns`,
   and `trackedRuntimePatterns` lists all gain it — B4).

No new flag is required. **Registered backends:** `bd init` refuses to
provision them today (`cmd/bd/init.go`: "can only open an existing
workspace"); their own workspace-creation path owes the same obligations,
proven by the conformance family (B7). Under A8 option A, a backend that
implements the six accessors as `ErrUnsupported` stubs keeps every existing
behavior; under option B nothing changes for it.

### A2. Client wiring — `bd init --bdp-server <url>` and `bd bdp client` (ruling 12, third command; amendment A6)

One more `bd init` target, beside `--server`, `--shared-server`,
`--proxied-server`, `--team-server`, and `--backend`. Every existing target
selects the provider or topology that realizes the storage interfaces — a
choice *below* the normalized abstraction. This one reroutes *above* it, at
the CLI: the `bd bdp` read verbs (A5) become a BDP client of the designated
server instead of opening a store.

**Two files, by what the key is.** `config.yaml` is git-tracked by default
(the tree's own gitignore template says so), so a key that differs per
workspace cannot live there — a teammate's `bd init --bdp-server` would
otherwise flip the authority's own workspace into client mode on the next
pull. The tree already has the answer: **`config.local.yaml`**, merged by
viper over `config.yaml` for "machine-specific settings without polluting
tracked config", untracked by convention (the template gains it if absent).

| Key | File | Values | Notes |
| --- | --- | --- | --- |
| `bdp.scope_url` | `config.yaml` (tracked; yaml-only) | absolute URL | a **project** fact (ruling 7a): what the authority mints/serves; every clone may know it; yaml-only so it never rides the Dolt `config` table |
| `bdp.authority_heartbeat` | `config.yaml` | duration (default `30s`) | remote-backed embedded topology only (A7); refused as `0` there |
| `bdp.authority_heartbeat_grace` | `config.yaml` | count (default `3`) | missed fetches before the server fails closed |
| `bdp.client` | **`config.local.yaml`** | `store` (default) \| `server` | per-workspace; explicit, never inferred from a URL; **not settable from env** (`blockedEnvVars`, the `backend` precedent) |
| `bdp.server` | `config.local.yaml` | absolute URL | the Scope URL of the designated server; `https` required unless loopback or `bdp.insecure_http: true` |
| `bdp.insecure_http` | `config.local.yaml` | bool (default false) | the named waiver; `bd bdp status` reports it |

**Writers.** `bd init --bdp-server <url>` and **`bd bdp client server
--server <url> [--insecure-http]`** / **`bd bdp client store`** write the
per-workspace keys to `config.local.yaml` through one shared writer (the
explicit lifecycle verb round 2 asked for). Generic `bd config set` accepts
`bdp.scope_url` and the heartbeat keys (yaml-only routing to `config.yaml`)
and **refuses** `bdp.client` / `bdp.server` / `bdp.insecure_http` with the
guidance "use `bd bdp client`" — so the tracked file can never receive a
per-workspace key by accident. There is deliberately **no token key in
config** (every key containing `token` is a secret to `IsSecretKey`, and the
tracked-config guard refuses it; other trackers keep tokens in yaml-only
config and accept that — BDP chooses a file). Mechanics: `bdp.` joins the
yaml-only prefix list in `internal/config/yaml_config.go` and
`recognizedConfigPrefixes` in `cmd/bd/config.go`; a `localOnlyKeys` class
names the three per-workspace keys; `validateYamlConfigValue` gains entries
for the enum, the URLs (absolute, scheme check), the duration, and the
count.

**Environment.** Viper binds config keys as `BD_<KEY>` (`SetEnvPrefix("BD")`,
dots to underscores): `BD_BDP_SERVER`, `BD_BDP_INSECURE_HTTP`,
`BD_BDP_AUTHORITY_HEARTBEAT` exist without new code; `BD_BDP_CLIENT` is
**blocked**. Ruling 7a spells the Scope URL variable `BDP_SCOPE_URL`, so it
is bound explicitly — `BindEnv("bdp.scope_url", "BDP_SCOPE_URL",
"BD_BDP_SCOPE_URL")`, the tree's `BEADS_IDENTITY` precedent — and both
spellings work. Hand-read connection/auth environment is `BEADS_*`, so the
client token file is **`BEADS_BDP_TOKEN_FILE`**. There is **no
`BD_BDP_TOKEN`**: an env-carried secret is inherited by every child process
(git hooks, dolt subprocesses) — the reason `remote_migrate_gate.go` keeps a
programmatic twin of `BD_ALLOW_REMOTE_MIGRATE`.

**Credential lookup** (client): `BEADS_BDP_TOKEN_FILE` > credentials file
section **`[bdp <origin><scope-path>]`** with `token=` (distinct from the
Dolt `[host:port]` `password=` sections — ruling 7a allows several
path-distinguished Scopes on one host). Transport: no redirects followed;
`Authorization` is sent only to the configured origin.

**Precedence, per command** (a table test pins it):

| Command | Resolution |
| --- | --- |
| `bd init --bdp-server <url>` | flag > `BD_BDP_SERVER` env > existing `config.local.yaml`; writes `bdp.client: server` and `bdp.server` to `config.local.yaml`; touches nothing else |
| `bd bdp client …` | writes `config.local.yaml`; effective for the next command |
| `bd config set bdp.scope_url` | validates, writes `config.yaml` (yaml-only routing) |
| every `bd bdp` verb | env (where permitted) > `config.local.yaml` > `config.yaml`; `bdp.client` (local file only) decides the route, `bdp.server` the target |
| `bd bdp status` | prints the resolved route, target, token source (never the token), `insecure_http`, and — on the store route — the identity state row (A3) |

**In `client: server` mode** the verbs' `openBeadGraph*()` accessors return
a BDP-client realization of the same `graphops` role interfaces (A5), so a
verb cannot tell which route it took. A workspace in this mode still has a
local graph store (A1 ran); `bd bdp serve` there **refuses** unless
`--serve-local-store` is passed. Issue verbs are unaffected in either mode.

### A3. `bd bdp serve` — serving a Scope (rulings 7b, 9, 12; amendments A1, A2, A7)

`bd bdp serve` is a **thin command over the existing `internal/httpapi`
server**: the same storage classification, the same hook peel, the same
`httpapi.Config` — plus a populated `Graph` field. It differs from
`bd serve` in one policy: it **requires a Scope this workspace holds**
(exit 2 otherwise), where `bd serve` mounts the BDP rows only when
`bdp.scope_url` is configured **and this workspace holds the authority**,
keeps the legacy surface up with the rows absent (and a notice) when it
does not, and serves exactly what it serves today when no URL is
configured. Because it builds the same server — every server built from
`routeTable` advertises `issues.claim` — it **inherits `errServeReadonly`
wholesale**: `--readonly` is refused ahead of the workspace, from v0. W2
decides whether `bd bdp serve` survives as the strict alias; nothing needs
folding later because there is one server.

```text
bd bdp serve [--addr IP:PORT]            default 127.0.0.1:0 (ephemeral), numeric IP only
             [--allow-non-loopback]      requires --auth-token-file (or --insecure-no-auth)
             [--auth-token-file PATH]    bearer tokens, one per line, hot-reloaded
                                         (env BEADS_SERVE_TOKEN_FILE — serve's own; no new env)
             [--insecure-no-auth]
             [--allowed-host NAME]...    extra exact Host allowlist entries
             [--scope-url URL]           first serve: mints under it; later: must equal the
                                         persisted Scope URL or refuse
             [--serve-local-store]       permit serving in a client: server workspace
```

No `--dev-local-test`: ruling 7a's `local-test` development mode is the BDP
reference server's; bd's tests configure a real `bdp.scope_url`.

The five posture flags are `bd serve`'s **variables**, not copies, with the
same validators, and `serveListen` builds the server. Behavior, in order:

1. **Classification is `serveDatabaseSource`, verbatim:** registered
   backend → store source; embedded Dolt → **refused, permanently**
   (`errServeEmbedded`); otherwise the unit-of-work provider — BDP's primary
   production path. The provider is built with the workspace's `beadsDir`
   (as today), which is how its accessors reach the witness manager.
2. **Roles from the same source.** Store arm: `serveIssueRoles`' one peel,
   and the graph roles are taken **off the same `src` value in the same
   function**. Provider arm: the provider beneath `uow.UnwrapProvider`
   carries the six accessors like every other role, `GraphConfig.Reader/
   Types` are nil, and `checkDatabaseSource`'s exactly-one-source rule
   extends to them ("same source" is structural; `checkDatabaseSource`
   verifies completeness and hook posture, not provenance).
3. **Capability probe, all-or-nothing:** the four read-path accessors are
   taken; any `*storage.ErrUnsupported` → `bd bdp serve` **exit 2, typed**;
   `bd serve` → BDP rows absent, legacy routes exactly as before. Any
   *other* error aborts startup as serve's role binding does.
4. **Identity** — the state table, split by command (each row a contract
   case, B7). `bd serve` **never refuses** on account of the graph: every
   refusal row is "legacy up, no BDP rows, notice" for it.

   | Persisted Scope | Configured URL | Witness (`.beads/graph-authority.local.json`) | `bd bdp serve` | `bd serve` |
   | --- | --- | --- | --- | --- |
   | none | none | — | exit 2 (7a: an explicit URL is required) | no BDP rows, silent |
   | none | set | — | **first serve: `Mint`, fenced per topology (A7)** — shared-server: take the lease + INSERT the Scope row + `mint` event; remote-backed: fetch → compare → commit → push (non-fast-forward → undo, "pull first"); then write the witness; serve honestly empty | same — the first `bd serve` under a configured URL mints (ruling 12) |
   | present | none | any | exit 2 | no BDP rows; notice names the persisted Scope |
   | present | **different** | any | refuse: configured ≠ persisted | no BDP rows; notice: configured URL contradicts the store |
   | present, matches | matches | **absent** (a clone; a pull into a fresh directory; a directory copied to another path or host — workspace key mismatch) | refuse `ErrNotAuthority`; guidance: `bd bdp promote` | no BDP rows; notice |
   | present, matches | matches | present, `(authority_id, epoch)` **stale** (promoted elsewhere) or lease not held (shared-server) | refuse; same guidance | no BDP rows; notice |
   | present, matches | matches | present, ledger head `{seq, hash}` **not in the store's ledger** (a database restore, or a different history) | refuse `ErrStateRewound`; guidance: `bd bdp restore` | no BDP rows; notice |
   | present, matches | matches | present, `unverified` set (`bd backup restore` ran) | refuse until `bd bdp restore` clears it | no BDP rows; notice |
   | present, matches | matches | present, consistent | serve; every operation re-asserts the witness in its transaction (A1) | serve BDP rows |

5. **Host policy:** the Scope URL's host joins the allowlist automatically;
   the listener is plaintext behind TLS termination.
6. **Fencing while serving (A7):** shared-server topology — the lease is
   renewed on `bdp.authority_heartbeat` through the ephemeral commit form
   (no history) and asserted inside every transaction; losing it fails
   closed. Remote-backed embedded topology — the server fetches the remote
   tracking ref on the heartbeat, compares `(authority_id, epoch)`, and
   after `bdp.authority_heartbeat_grace` missed fetches, or any higher
   epoch, stops serving BDP rows with `ErrNotAuthority` (`bd serve` keeps
   the legacy surface; `bd bdp serve` exits 3) and logs the promoting
   authority. No-remote embedded — nothing to watch.
7. **Mounting:** `serveListen(opts, httpapi.Config{…, Graph:
   &httpapi.GraphConfig{…}})`. Liveness stays `GET /healthz`; BDP readiness
   is a real discovery read.
8. **Lifecycle:** excluded from the post-command maintenance net by the
   `bdpRootPolicy` table (Part A), not by sharing the leaf name `serve`; the
   events-journal maintenance ticker applies as for `bd serve`.

### A4. `bd bdp promote` / `bd bdp restore` / `bd bdp ledger` (rulings 9, 11; amendments A5, A7)

These are the **only** reachers of `BeadGraphAdmin()`; the server assembly
has no field for it (architecture §4). They run under the workspace's
exclusive advisory lock.

- **`bd bdp promote`** — make *this workspace* the authority for the Scope
  it carries, fenced per topology (A7):
  - *shared-server:* `Promote` takes the lease if expired (or with
    `--steal`, operator-confirmed, which also revokes the holder's row) and
    CASes the epoch (`epoch = epoch + 1 WHERE epoch = <read>`; a lost race
    is a typed refusal) with a `promote` ledger event, in one transaction;
    then writes the witness.
  - *remote-backed embedded:* one provider operation
    (`PublishAuthorityTransition`): `DOLT_FETCH`; require local HEAD ==
    remote-tracking HEAD (`DOLT_HASHOF`); CAS + `promote` event;
    `DOLT_COMMIT` with a named message; `DOLT_PUSH`. A non-fast-forward
    push means another promotion landed first: `DOLT_RESET --hard
    <remote ref>`, no witness, refusal "pull first"; if the reset itself
    fails the command reports the divergent commit hash and the operator
    resets by hand (stated, not hidden). The superseded authority's next
    fetch (heartbeat) or pull refuses it.
  - *no remote:* the CAS alone.
  `--rotate-url <new>` additionally rotates the URL when history diverged
  (operator-confirmed).
- **`bd bdp restore`** — runs after a database restore. `bd backup restore`
  (`runBackupRestore`, after a successful `RestoreDatabase`) calls
  `Admin.MarkUnverified`; independently of that flag, the ledger-head check
  (A3 row 7) refuses a store whose ledger does not contain the witness's
  `{seq, hash}`. `bd bdp restore` branches on the provider's declaration:

  | `LedgerDurability` | Meaning | `bd bdp restore` does |
  | --- | --- | --- |
  | `in-state` (Dolt, v0) | the ledger is ordinary versioned state | requires a ledger snapshot whose range reaches the witness's `{seq, hash}` (`--ledger <file>`, applied first) to show continuity; **otherwise rotates** the Scope URL (operator supplies the new one) and epoch, appends the `refuse_url` event, and re-grants the witness |
  | `independent` | the provider keeps the ledger outside the restored state (contract case) | re-validates and re-grants; no rotation |
  | `none` | no ledger | always rotates |

  Never silent. Residuals, stated: a whole-directory filesystem snapshot
  rewinds witness and state together and is undetectable in-band
  (operators who take such snapshots schedule `bd bdp ledger snapshot`
  beside them); a witness file copied to the *same path on the same host*
  is that workspace; a `dolt` CLI or `bd sql` restore sets no flag — the
  ledger-head check is what catches it.
- **`bd bdp ledger snapshot [--from SEQ] [--to SEQ]`** / **`bd bdp ledger
  apply <file>`** — the continuity lane ruling 11 needs: the hash-chained
  `graph_ledger_events` as JSONL under a **manifest** `{scope_url,
  authority lineage (every mint/promote/rotate event's authority_id and
  epoch), first_seq, last_seq, prev_hash of first, head_hash}`. Apply
  validates lineage against the Scope row, contiguity against the store's
  current head (`prev_hash` must equal it; a gap or a foreign lineage is
  refused), appends, and re-derives `graph_allocations` and
  `graph_scope_history`. It never rewinds a tombstone or a refusal.

### A5. Graph verbs (v0 reads; P3 writes; names fixed now)

```text
bd bdp bead get <path> | list [--type URL] [--after CURSOR] [--limit N]
bd bdp link get <path> | list [--type URL] [--source PATH] [--target REF] [--after CURSOR] [--limit N]
bd bdp types [get <url>]
bd bdp status
bd bdp client store | server --server URL [--insecure-http]
bd bdp serve | promote | restore | ledger snapshot|apply
(P3)  bd bdp bead create|update|delete ; bd bdp link create|update|delete
```

Each verb reaches its role through an accessor and nothing else, on
whichever route the workspace is on — the `cmd/bd/label.go` pattern plus
the third route, with a route-fork test in the shape of
`cmd/bd/vc_recompute_test.go`:

```go
func openBeadGraphReader() (graphops.Reader, error) {
    if bdpClientMode() == "server" {        // config.local.yaml / config.yaml only, per A2
        return bdpclient.Reader(bdpClientConfig())          // internal/bdpclient
    }
    if usesProxiedServer() { return proxiedBeadGraphReader() }   // the UOW provider; no new HTTP route
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

## Part B — Storage interfaces

### B1. Accessors on `storage.Storage` (architecture §2a; amendments A4, A8 — option A wording)

Named **`BeadGraph*`**: `Storage.GraphCounter()` already exists and counts
the *issue* dependency graph.

```go
// BeadGraphReader returns the guarded graph-read surface for this store:
// Bead and Link records (a Bead with its complete, grouped, bounded
// ownedLinks, assembled in the call's one transaction), keyset selections
// in code-unit order under an opaque cursor, and incident Links. Every
// method asserts this workspace's authority witness inside that
// transaction; no request carries authority. Reads fire no hooks; the
// hook decorator recurses.
BeadGraphReader() (graphops.Reader, error)

// BeadGraphTypes returns the Type Descriptor inventory (ordered, bounded,
// keyed); protected like every read.
BeadGraphTypes() (graphops.DescriptorReader, error)

// BeadGraphTypeInstaller returns the descriptor install/converge role used
// by bd init's bootstrap and the conformance fixture. Protected once a
// Scope exists; every install is a ledger event.
BeadGraphTypeInstaller() (graphops.TypeInstaller, error)

// BeadGraphIdentityReader returns the Scope row, this workspace's witness
// claim, and the provider's LedgerDurability declaration. Exempt: it
// reports state.
BeadGraphIdentityReader() (graphops.IdentityReader, error)

// BeadGraphBootstrapper returns the one-time mint, fenced per topology.
// The server assembly may hold it on the first-serve path and nothing else.
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
`configStore` stub, and every other `Storage` implementer — which, under
option A, **the compiler enumerates** (a required method fails every
implementer that lacks it); the three censuses (B5) cover the decorators
and facades. **This is the source break `storage.go` declares for
out-of-tree backends** (A8 option A): six one-line `ErrUnsupported` stubs,
called out in CHANGELOG as the `ReadyClaimer` accessor's entry was ("any
external type that *implements* it … must add the method to compile").
Under option B, B1 becomes a `BeadGraphCapable` interface plus explicit
decorator implementations, a capability census, and a resolver; the rest
of this spec is unchanged.

### B2. The `graphops` leaf (public, repo root; amendment A4)

```go
package graphops   // imports: stdlib, beadserrors — nothing else

// ---- requests and results: NO authority fields anywhere (the witness is the store's)
type Cursor string            // OPAQUE: store-produced; binds Scope URL, epoch, selection hash, last path;
                              // P2 adds snapshot identity inside it — no public interface change
type BeadRequest        struct{ Path string }
type LinkRequest        struct{ Path string }
type BeadSelectRequest  struct{ TypeURL string; After Cursor; Limit int }
type LinkSelectRequest  struct{ TypeURL string; SourcePath string; Target *Ref; After Cursor; Limit int }
type IncidentRequest    struct{ Path string; Direction Direction /* In | Out | Both */; After Cursor; Limit int }
type DescriptorRequest  struct{ URL string }
type InstallRequest     struct{ Descriptors []TypeDescriptor }

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
    Descriptors(ctx) ([]TypeDescriptor, error)            // ordered by URL; bounded by MaxCatalog
    Descriptor(ctx, DescriptorRequest) (TypeDescriptor, error)
}
type TypeInstaller interface {
    Install(ctx, InstallRequest) (InstallResult, error)   // idempotent by fingerprint; closure validated; an owning
}                                                          // declaration without Max → ErrValidation; over MaxCatalog → ErrValidation
type IdentityReader interface {
    Read(ctx) (ScopeIdentity, error)                      // Scope row + witness claim (Held bool, Epoch, LedgerSeq, Unverified)
    LedgerDurability(ctx) (LedgerDurability, error)       // in-state | independent | none (ruling 11)
}
type ScopeBootstrapper interface {
    Mint(ctx, MintRequest) (ScopeIdentity, error)         // fenced per topology; singleton row + mint event + witness
}
type Admin interface {
    Promote(ctx, PromoteRequest) (ScopeIdentity, error)   // fenced per topology; CAS + promote event + witness (+ push)
    Rotate(ctx, RotateRequest) (ScopeIdentity, error)     // new URL; refuse_url event
    LedgerSnapshot(ctx, LedgerRange) (LedgerManifest, []LedgerEvent, error)
    LedgerApply(ctx, LedgerManifest, []LedgerEvent) (LedgerApplyResult, error)
    MarkUnverified(ctx) error
    ClearUnverified(ctx) error
}
```

**Bounds.** `MaxExpandedRows` (10,000) caps `len(page) + Σ ownedLinks` for
any read; the batched owned-links query carries **`LIMIT (MaxExpandedRows −
rows already materialized) + 1`**, so the combined bound holds before any
over-bound row is decoded, and `ErrRepresentationTooLarge` is typed;
descriptor decoding is capped the same way. Owning Types **must** declare
`Max`; the installer refuses one that does not. Reads are **five
statements, no per-row statements** — Scope row + state version; ledger
head; the row or page; the descriptors for the Types on the page (cached by
fingerprint); the batched owned-links query — and the contract's
statement-count case pins exactly that.

**Values** (`Bead`, `Link`, `Ref`, `Properties`, `Revision`, `Attribution`,
`TypeDescriptor`, `OwnedLinkDecl`, `ScopeIdentity`, `LedgerEvent`,
`LedgerManifest`) have unexported fields and constructors that enforce the
laws in `laws.go`. `Properties` is the immutable raw-JSON object value from
the plan; its canonical bytes are what B4 stores. `Ref` is a sum: in-Scope
(`Path`) or external (`URL`), with the discriminant stored. `Revision` is
128 bits from `crypto/rand`, lower-hex, minted by the write body. A
`LedgerEvent` carries `{seq, kind, path|scope_url, authority_id, epoch, at,
prev_hash, hash}` with `hash = sha256(canonical(event without hash))`.

**Errors:** `beadserrors` (stdlib-only) declares the new sentinels
(`ErrNoScope`, `ErrScopeExists`, `ErrNotAuthority`, `ErrStateRewound`,
`ErrURLReused`, `ErrRepresentationTooLarge`, `ErrNotServedYet`) and the
typed `GoneError{Path, State}` (state `pruned` | `erased`); `graphops`
aliases them so `errors.Is` crosses the module boundary.

### B3. Bodies, the witness, and legs

`internal/storage/graphops` bodies take **`DBTX`** (the `issueops.DBTX`
shape, declared locally), which `*sql.Tx` and `domain/db.Runner` both
satisfy, **and the witness** the accessor loaded:

```go
func ReadBeadInTx(ctx, tx DBTX, w authority.Witness, req graphops.BeadRequest) (graphops.BeadRecord, error)
```

`internal/storage/authority` is the witness manager (architecture §3):
`Load(beadsDir)` under a shared advisory lock on `.beads/graph-authority.lock`
(the `internal/lockfile` primitives); `Advance(beadsDir, fn)` under the
exclusive lock — read, apply, **never let `LedgerSeq` decrease**, write
through `internal/atomicfile`, fsync the directory; refuses to write unless
`.beads/.gitignore` carries the entry. A missing or unparsable file is
`ErrNotAuthority`; a workspace-key mismatch is `ErrNotAuthority` with the
copied-file guidance. **Order of a mutation:** DB transaction commits (and
the push lands, where the topology requires) *before* the witness advances,
so the file is never ahead of durable state; a process dying between the
two leaves the witness behind, which the next assertion tolerates (the
store's head is at or past the witness's) and the next successful mutation
catches up.

| Leg | Files | Body |
| --- | --- | --- |
| server Dolt | `internal/storage/dolt/beadgraph_*.go` | accessor loads the witness, wraps the body in `withReadTx` (reads) / `withRetryTx` + `DOLT_COMMIT` (install, mint, promote, rotate, apply); `PublishAuthorityTransition` for the remote-backed fence |
| embedded Dolt | `internal/storage/embeddeddolt/beadgraph_*.go` (`//go:build cgo`) | same body, `withConn` |
| unit of work | `internal/storage/domain/beadgraph.go` (`BeadGraphUseCase`), `internal/storage/domain/db/beadgraph.go` (over `db.Runner`, delegating to the bodies — the `IssueUseCase()` → `CompareAndSetMetadataKeyInTx` shape), `internal/storage/uow/beadgraph_*.go` (`uow.UnitOfWork.BeadGraphUseCase()`; provider accessors via `RunTxRead`; mutations via `RunTxResult` with commit messages `bdp: mint scope <url>` / `bdp: promote epoch <n>` / `bdp: install types <fingerprint>` / `bdp: ledger apply <from>-<to>`; lease heartbeats via `RunTxEphemeral` — no history) | **same body** |

Contract headers say **"one reading plus an engine check"**. Every
protected body begins with `assertAuthorityInTx(ctx, tx, w)`: Scope row
identity; ledger head `{seq, hash}` present (exact prefix) and `MAX(seq) >=
w.LedgerSeq`; on the shared-server topology the `graph_authority_lease`
row (holder `w.WorkspaceKey`, unexpired); state version (`storage.StateHasher`
— Dolt's `DOLT_HASHOF`) equal to `w.StateVersion`, else `ValidateStateInTx`
under **singleflight** (one validation per state change, not one per
concurrent read) and a witness update on success. Exempt: `IdentityReader`
(reports state); `Admin.MarkUnverified`/`ClearUnverified` (witness-file
operations under the exclusive lock); `Install` on an unminted store.
`SeedBeadInTx`/`SeedLinkInTx` (P1 fixture writer) allocate through the
ledger like a real write; every call site is a `_test.go` file
(source-scan test) and the package is denied to everything but the three
legs and `domain/db` (depguard, B6).

### B4. Schema (migrations; frozen once merged)

Rules the tree enforces and this schema obeys: migrations are **frozen once
merged** and content-hashed across clones (`check-migration-hygiene.sh`
check C); **no `NOW()`/`UUID()`/`RAND()`** in migration SQL (check B) — every
timestamp and id is set in Go; real-Dolt tests for anything a `sqlmock` echo
cannot exercise; DDL is not transactional across statements, so each
`CREATE` is guarded and resumable. **Seven replicated tables in five
files** — `NNNN_beadgraph_scope.up.sql` (scope + history),
`NNNN_beadgraph_types.up.sql`, `NNNN_beadgraph_beads.up.sql`,
`NNNN_beadgraph_links.up.sql`, `NNNN_beadgraph_ledger.up.sql` (events +
allocations) — plus **one ignored-series file**,
`ignored/NNNN_beadgraph_authority_lease.up.sql`, with its `dolt_ignore`
registration in the main series first (the 0055 `leases` shape).

**Collation.** Dolt's default collation is already binary
(`utf8mb4_0900_bin` — probed in round 2), and no migration in the tree
declares one. Every identifier column below still carries **`CHARACTER SET
utf8mb4 COLLATE utf8mb4_bin`** explicitly (written `BIN`), as the defense
for providers whose default is case-insensitive and against a future
default change — with a contract case asserting two paths differing only
in case are distinct and sort by code unit.

**Identity is Scope-relative.** Rows store the canonical Scope-relative
`path`; the absolute URL is `scope_url + path`, computed at the boundary,
so a URL rotation rewrites no rows. External references keep their
absolute URL with a discriminant.

**JSON is bytes.** `properties` and `descriptor` are canonical JSON bytes in
`LONGBLOB`, never the engine `JSON` type: the tree measured that column
decoding through float64 (`1.0`→`1`, integers past 2^53 rounded, `1e300`
expanded — `internal/storage/issueops/metadata_cas.go`; `-0.0`→`0` per the
role guide's own measurement note). Round-trip tests cover each case. Size
limit 1 MiB per value.

**Provenance on every row.** `last_authority_id` / `last_epoch` are stamped
by every mutation, so the validator can identify a foreign *update*.

| Table | Columns (type; nullability) | Keys / constraints |
| --- | --- | --- |
| `graph_scope` | `id TINYINT NOT NULL` (always 1), `scope_url VARCHAR(2048) BIN NOT NULL`, `authority_id CHAR(32) NOT NULL`, `epoch BIGINT UNSIGNED NOT NULL`, `minted_at DATETIME(6) NOT NULL` | `PRIMARY KEY (id)`, `CHECK (id = 1)` — singleton; also the row every ledger-event append locks (`SELECT … FOR UPDATE`) to serialize `seq` |
| `graph_scope_history` | `scope_url VARCHAR(2048) BIN NOT NULL`, `refused_seq BIGINT UNSIGNED NOT NULL`, `refused_at DATETIME(6) NOT NULL`, `reason VARCHAR(64) NOT NULL` | `PRIMARY KEY (scope_url)`; derived from `refuse_url` events |
| `graph_type_descriptors` | `url VARCHAR(2048) BIN NOT NULL`, `descriptor LONGBLOB NOT NULL`, `fingerprint CHAR(64) NOT NULL`, `installed_seq BIGINT UNSIGNED NOT NULL`, `installed_at DATETIME(6) NOT NULL`, `last_authority_id CHAR(32) NOT NULL`, `last_epoch BIGINT UNSIGNED NOT NULL` | `PRIMARY KEY (url)`; `UNIQUE (fingerprint)` |
| `graph_beads` | `path VARCHAR(1024) BIN NOT NULL`, `type_url VARCHAR(2048) BIN NOT NULL`, `revision CHAR(32) NOT NULL`, `attribution_principal VARCHAR(512) NULL`, `attribution_status ENUM('claimed','unknown') NULL`, `properties LONGBLOB NOT NULL`, `last_authority_id CHAR(32) NOT NULL`, `last_epoch BIGINT UNSIGNED NOT NULL`, `created_at DATETIME(6) NOT NULL`, `updated_at DATETIME(6) NOT NULL` | `PRIMARY KEY (path)`; `INDEX (type_url, path)`; `FOREIGN KEY (type_url) REFERENCES graph_type_descriptors(url)`; attribution columns both NULL or both set |
| `graph_links` | `path VARCHAR(1024) BIN NOT NULL`, `type_url … BIN NOT NULL`, `revision CHAR(32) NOT NULL`, `source_path VARCHAR(1024) BIN NOT NULL`, `source_pin CHAR(32) NULL`, `target_kind ENUM('in','ext') NOT NULL`, `target_path VARCHAR(1024) BIN NULL`, `target_url VARCHAR(2048) BIN NULL`, `target_pin CHAR(32) NULL`, `attribution_*`, `properties LONGBLOB NOT NULL`, `last_authority_id`, `last_epoch`, timestamps | `PRIMARY KEY (path)`; `INDEX (source_path, type_url, path)`, `INDEX (target_path, path)`; `FOREIGN KEY (source_path) REFERENCES graph_beads(path)`; `CHECK` exactly one of `target_path`/`target_url` per `target_kind`; **no** uniqueness on (type, source, target) — multiplicity is law |
| `graph_ledger_events` | `seq BIGINT UNSIGNED NOT NULL`, `kind ENUM('mint','install','promote','rotate','allocate','tombstone','refuse_url') NOT NULL`, `path VARCHAR(1024) BIN NULL`, `scope_url VARCHAR(2048) BIN NULL`, `resource_kind ENUM('bead','link') NULL`, `state ENUM('pruned','erased') NULL`, `fingerprint CHAR(64) NULL`, `authority_id CHAR(32) NOT NULL`, `epoch BIGINT UNSIGNED NOT NULL`, `at DATETIME(6) NOT NULL`, `prev_hash CHAR(64) NOT NULL`, `hash CHAR(64) NOT NULL` | `PRIMARY KEY (seq)` — **append-only, hash-chained**; `UNIQUE (hash)`; `INDEX (path, seq)`; `CHECK` per kind on which payload columns are set. Every v0 mutation is an event, so the witness's head covers every graph-state rewind |
| `graph_allocations` | `path VARCHAR(1024) BIN NOT NULL`, `resource_kind ENUM('bead','link') NOT NULL`, `birth_seq BIGINT UNSIGNED NOT NULL`, `birth_authority_id CHAR(32) NOT NULL`, `birth_epoch BIGINT UNSIGNED NOT NULL`, `state ENUM('live','pruned','erased') NOT NULL`, `tombstone_seq BIGINT UNSIGNED NULL` | `PRIMARY KEY (path)` — the O(1)/O(log n) ID test (ruling 3); **derived state**, re-derivable from the events |
| `graph_authority_lease` (**ignored series**, dolt-ignored, never replicates) | `id TINYINT NOT NULL`, `holder_workspace_key CHAR(64) NOT NULL`, `epoch BIGINT UNSIGNED NOT NULL`, `granted_at DATETIME(6) NOT NULL`, `expires_at DATETIME(6) NOT NULL`, `heartbeat_at DATETIME(6) NOT NULL` | `PRIMARY KEY (id)`, `CHECK (id = 1)`; the shared-server fence (A7); heartbeats through the ephemeral commit form, never a Dolt commit — the `leases` (bd-lrgn1) precedent |

`updated_at` is protocol-irrelevant bookkeeping. Bead `type_url` and Link
`source_path`/`target_*` are immutable after insert (enforced in the write
body and checked by the validator).

**The witness file: `.beads/graph-authority.local.json`.**
`{workspace_key, scope_url, authority_id, epoch, ledger_seq, ledger_hash,
state_version, unverified, granted_at}` where `workspace_key =
sha256(hostname ":" realpath(.beads))` — **not** the project id, which
every clone of a project shares (`resolveInitProjectID` adopts the
database's `_project_id`). Written only by the manager (B3) under the
exclusive lock, and only once the workspace's `.beads/.gitignore` carries
the entry: the doctor's template, `requiredPatterns` (appended to existing
workspaces by `bd init`/doctor via `EnsureGitignoreForBeadsDir`), and
`trackedRuntimePatterns` (the list that catches an already-committed copy)
all gain `graph-authority.local.json` and `graph-authority.lock`. What
each operation does to it: `git clone` / `dolt clone` — absent; pull —
untouched (neither transfers nor removes it); `bd backup restore` /
`DOLT_BACKUP` restore — present, and the ledger head it names is no longer
in the store → `ErrStateRewound`; directory copy to another path or host —
present, workspace key mismatch → `ErrNotAuthority`; directory copy to the
same path on the same host — that *is* the workspace.

**Cross-repo coupling (bts).** `DoltTeamServer` workspaces refuse to open
when `current < latest` with "ask your operator to run `bts migrate`" and
**no `BD_IGNORE_SCHEMA_SKEW` hatch** — a **numeric-version** comparison
only. The coupling is therefore a cross-repository **release-parity gate**:
bts must ship byte-identical copies of the five main-series files and the
ignored-series file (the embedded content hashes in
`migration_content_hashes.go` are what a bts-side parity test compares).
The migration PR is sequenced with bts and says so; the remote-migrate gate
(#4259) forces migrate-vs-adopt on every remote-backed workspace at upgrade.

### B5. Decorators, censuses, and every embedding surface

- `internal/storage/hook_beadgraph_*.go` (six files): declared, recurse
  **unwrapped** — the hook vocabulary hands scripts an *issue*; a graph
  hook vocabulary is a separate proposal. `storage.RoleFiresHooks` is a
  type switch over hook wrappers, so an unwrapped role needs no entry; a
  test asserts each graph role answers `false`. Added to
  `role_accessor_decorator_test.go`'s wrapped/unwrapped table.
- `internal/telemetry/beadgraph_*.go`: every method spanned `storage.op` /
  `storage.done`; the **telemetry census** gains the classification.
- **Three censuses**, each must learn `graphops`: the storage reflection
  census (`roleAccessorNamesOf`), the telemetry census, and the conformance
  package's **source-parsed** `facadePackages` map
  (`backend/conformance/role_coverage_scan_test.go`) — without the third,
  `TestRoleFacadeCensusAgreesWithReflection`
  (`backend/conformance/role_coverage_gate_test.go`) fails.
- `internal/storage/uow/notifying.go`: explicit accessors built from *this*
  provider (parity test); `internal/httpapi/claim.go`'s `timedProvider`;
  `cmd/bd/serve.go`'s `serveRoleSource` and its non-embedding test stubs;
  `internal/jira/tracker_test.go`'s `configStore` stub — every surface that
  embeds the store or a provider, enumerated by
  `grep -l 'func (.*) Memories()'` at implementation and listed in the PR;
  under option A the compiler finds the rest.

### B6. `backend/` public surface and depguard

- **No aliases** (amendment A4): `graphops` is public and imported
  directly, like `issueops`. `TestPublicSurfaceComplete` stays green
  *because* no `internal/` type is reachable from the new accessors — a
  test asserts that.
- `backend/backend.go`'s "minimal external backend" example gains the six
  accessors as `ErrUnsupported` stubs (option A); the **CHANGELOG entry**
  follows the `ReadyClaimer` precedent's wording.
- `.golangci.yml` gains a **new, stricter** rule (cmd/bd imports the
  `issueops` tx-body package in fourteen files today, so this is not the
  existing convention): `internal/storage/graphops` is importable only by
  `internal/storage/{dolt,embeddeddolt,domain/db}` and its own tests, with a
  mutation test that removes the entry and expects the lint to pass.

### B7. Conformance

- Families: `beadgraph_reader_contract.go`, `beadgraph_types_contract.go`
  (reader + installer), `beadgraph_identity_contract.go` (reader,
  bootstrapper, admin), each citing the leaf doc by line.
- `RoleContractBundle` gains six factory fields **and** their rows in
  `role_bundle_cases.go`; `BeadGraphFixture` carries the seed hook and a
  `Witness` hook (a temp workspace directory standing in for `.beads/`).
- Wirings in `internal/storage/{dolt,embeddeddolt,uow}/beadgraph_*_contract_test.go`;
  the leg registry (`internal/storage/contract_leg_registry_test.go`) and
  `TestEveryLegWiresEveryRoleContract` see them; both coverage gates apply.
- Non-capable stores answer `*storage.ErrUnsupported{Op: "<accessor
  name>"}` — the six strings pinned — proven per accessor.
- Cases the councils asked for by name: a clone produced by push/pull
  refuses; a **`DOLT_BACKUP` restore of an authority** refuses
  (`ErrStateRewound`, with at least the `mint` event past the backup point
  — every v0 mutation is an event); a **copied witness file** in another
  directory refuses; two clones minting before either pulls (remote-backed:
  the second push is non-fast-forward and undone); concurrent first-serve
  mint on one database (one wins); promotion race (one CAS wins; the
  loser's push is non-fast-forward and reset); heartbeat detects a higher
  `(authority_id, epoch)` and fails closed after the grace; shared-server
  lease expiry refuses in-transaction; rotation refuses the old URL;
  case-differing paths distinct and code-unit ordered; ownedLinks
  completeness incl. **empty groups** and the bound under `LIMIT
  remaining+1`; keyset continuation inside one transaction; gone-family;
  a promote in another process is honored by the next read (witness
  reload); descriptor read on a non-authority clone refuses; install after
  mint on a non-authority clone refuses and `bd init` skips it; ledger
  snapshot/apply round trip, gap refusal, foreign-lineage refusal; the
  installer refuses an owning declaration without `Max`; statement count =
  5; validator singleflight under concurrent reads after a pull.
- Differential-gate rows: every legacy form of `bd link`, `bd graph`,
  `bd graph check`, `bd restore`, `bd promote` parses and behaves as before;
  `bd init` output on a non-capable backend is byte-identical.

### B8. `httpapi` integration and the pinned wire (amendments A2, A7)

- **`httpapi.Config.Graph *GraphConfig`** — `{Reader graphops.Reader; Types
  graphops.DescriptorReader; ScopeURL string; Fence FenceSource}` on the
  store arm; on the provider arm `Reader`/`Types` are nil and the provider's
  own accessors answer per request through `timedProvider`. `Fence` is the
  topology's watcher (lease renewal or remote fetch). There is deliberately
  **no admin or installer field**: a test asserts the struct cannot carry
  one; `checkDatabaseSource`'s exactly-one-source rule extends to the graph
  fields.
- **`bdpRouteTable`** (`internal/httpapi/bdp_routes.go`) — **P2**: rows in
  the same `route` shape (`authExempt: false`; `projectExempt: false` — an
  absent stamp passes; semaphore taken), registered by `Server.handler()`
  **only when `cfg.Graph != nil`**, each wrapped by the same `s.route(rt)`.
  First rows: discovery, `types/`, one Bead, one Link; **collection rows
  wait for P2's cursor ADR**. No token added to
  `ContextResponse.capabilities` in v0; `TestSpecRouteParity` compares only
  `routeTable`; a sibling parity test compares `bdpRouteTable` against the
  pinned schema's path grammar.
- **Posture parity test:** one refusal matrix drives a legacy row and a BDP
  row and asserts identical status and log shape.
- **Handler = serializer** (`bdp_handlers.go`); typed graph errors → BDP
  Problem records (`bdp_problem.go`), here and only here.
- **Wire — P0, not yet in the tree:** `internal/httpapi/bdpwire/schema/bdp-v0.schema.json`
  **will be vendored** with a `PROVENANCE` file (upstream repo, commit — the
  plan's §0 pin — and sha256); `make bdp-gen` runs a **pinned**
  JSON-Schema→Go generator (fallback recorded at P0: hand-written DTOs
  validated against the schema); `make bdp-check` regenerates and diffs, and
  `scripts/ci/pr-policy.sh` runs it beside `make api-check`. The artifact's
  presence and hash are P0's exit gate.

## Part C — What does not change

`storage.Storage`'s existing 28 accessors and every `issueops` role; the
journal's frozen vocabulary; `openapi.v0.yaml` and `TestSpecRouteParity`;
`bd serve` on a workspace with **no** `bdp.scope_url` (byte-identical) and
on a non-authority clone (legacy surface up, rows absent); every legacy CLI
verb (differential gate rows in B7); JSONL export shapes; `metadata.json`'s
schema; `bd init` output on non-capable backends.

## Part C2 — What changes that an earlier draft claimed did not

- **Merge, pull, and sync.** Every `DOLT_PULL`/`DOLT_MERGE` route (the four
  `versioncontrolops` functions; `doltCLIPull`, `Pull`, `PullRemote`,
  `pullTransport`/`pullWithAutoResolve` in `internal/storage/dolt/store.go`;
  the UOW leg's `DoltRemoteUseCase`; embedded federation sync; the
  remote-migrate gate's fast-forward `DOLT_MERGE`) can change graph state
  outside the roles, so the **state-change validator** runs on every
  observed state-version change at assertion time (B3) and refuses a
  foreign-authority or invalid delta; a superseded clone that pulls resets
  to the remote. The replication/merge ADR specifies its rules before the
  migrations land.
- **Federation.** Graph tables ride filtered pushes **unfiltered, by
  decision, in v0**; the lease table never replicates.
- **`bd sql`, raw SQL, and force-push.** Out of contract for graph tables;
  the validator catches what the next state-version change reveals; the
  enforcement-boundary ruling decides the rest before P3.
- **`bd backup restore`** calls `Admin.MarkUnverified` after a successful
  `RestoreDatabase`.
- **Root store policy** gains the CommandPath-keyed table (Part A) and the
  `commandPolicy` consolidation.
- **`bd init`** gains the `.beads/.gitignore` entry and the
  bootstrap-only-when-permitted rule (silent at default verbosity).

## Part D — Open implementation questions (not rulings)

1. `MaxExpandedRows`, `MaxCatalog`, page default/max, the 1 MiB value
   limit, the lease TTL, the heartbeat default and grace are proposed
   numbers; the leaf doc fixes them at P1 with the rationale.
2. Whether the P1 fixture writer stays fixture-only through P3 or becomes
   the internal half of `graphops.Writer`.
3. Generator choice for `bdpwire` — recorded at P0 with the provenance file.
4. Whether `bd bdp serve` remains after W2 (default: the strict alias).
5. Whether the Dolt realization adds an ancestry belt to the witness check
   (`DOLT_MERGE_BASE(state_version, HEAD) == state_version`) — cheap, and it
   detects rewinds of non-graph state too.

## Part E — Proposed ruling amendments this spec assumes (pending)

A1 store-owned witness asserted in every transaction is the v0 lease
(ruling 9); A2 BDP rows inside `httpapi`, `bd bdp serve` a thin strict
command inheriting `errServeReadonly`, `bd serve` never refusing on account
of the graph (rulings 7b/12); A3 `bd bdp …` namespace with an
authoritative CommandPath-keyed policy; A4 values, laws, roles in public
`graphops`, `BeadGraph*` accessors, no `backend/` aliases; A5 ruling 11's
mechanism = workspace-keyed witness file + hash-chained ledger events +
`bd bdp ledger snapshot|apply` manifests + provider `LedgerDurability`;
A6 `bdp.scope_url` tracked, per-workspace keys in `config.local.yaml` via
`bd bdp client`, `BDP_SCOPE_URL` honored, no env token, `bdp.client`
blocked from env; A7 fencing per topology (shared-server lease; remote fetch
→ compare → commit → push; push-on-commit for P3 writes); A8 two options
for constraint #1. Plus two decisions the plan does not yet hold: the
out-of-role DML enforcement boundary, and the replication/merge ADR as a
P1 gate. Full text: architecture §2b.
