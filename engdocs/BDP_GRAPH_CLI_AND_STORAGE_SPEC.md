# BDP graph store — CLI and storage-interface changes, in detail

**Status:** Draft v5 (W-arch, after four council rounds) — feat/bead-graph
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
more sites than the obvious ones — `effectiveRootStorePolicy(cmd.Name(), …)`,
`runsPostCommandMaintenance(cmd.Name(), …)` (`cmdName == serveCmdName`),
`isReadOnlyCommand` (`readOnlyCommands`, which `context_cmd.go` mutates at
init), `shouldAutoPruneEventsJournal` in `cmd/bd/events_journal.go`, the
`cmd.Name() != "import" && != "setup"` branches, `workspace_gate.go`,
`main_errors.go` — and the tree records that same-named subcommands
collide. Leaf names under `bd bdp` **may** coincide with those lists
(`bd bdp bead list` is a `list`; `bd bdp serve` is a `serve`); the
coincidence never governs. One `commandPolicy(*cobra.Command)` keyed by
`CommandPath()` is consulted first at each site for any path under
`bd bdp` (the only `CommandPath()`-keyed map in the tree today is
`help_supplements.go`; `CheckMigrationFreeze`'s use is an error label, so
this is a new seam), with an exhaustive Cobra-tree test that walks the
`bd bdp` subtree **paired with a source scan** that fails on any new
`cmd.Name()` consumer not routed through the policy.

| Verb | Local store | Maintenance | Note |
| --- | --- | --- | --- |
| `bd bdp bead get\|list`, `link get\|list`, `types [get]`, `status` | opens **read-only** when `bdp.client: store`; **skipped entirely** when `bdp.client: server` (no schema check, no auto-start, no migration prompt) | no | `bdp.client` is read from `config.local.yaml`/`config.yaml` before any store opens |
| `bd bdp serve` | serve's own classification (A3) | no | |
| `bd bdp client` | none (writes `config.local.yaml`) | no | |
| `bd bdp types install`, `promote`, `restore`, `ledger snapshot\|apply` | opens writable, always local, under the **workspace exclusive gate** (`internal/workspacegate`, the `bd backup restore` precedent) | no (admin verbs commit — and where hazard R applies, publish — explicitly) | |

### A1. `bd init` — graph store initialization (ruling 12)

`bd init` initializes the graph store with everything else it initializes,
against the normalized storage interfaces:

1. Runs the graph migrations (Part B4): the replicated series in the normal
   migration series; the dolt-ignored `graph_authority_lease` table arrives
   through the tree's three-part mechanism — its name in the canonical
   `doltIgnorePatterns` list (seeded by `MigrateUp` before either series), a
   main-series migration that creates it for existing workspaces, and the
   ignored-series twin for fresh clones (the 0055 / `ignored/0012` shape;
   hygiene check D demands the twin).
2. **Installs no descriptors.** There is no ledger before mint, and a
   descriptor row needs the provenance a Scope supplies, so the built-in
   catalog is installed by `Mint` (A3) inside its fenced transaction; later
   catalog changes go through `bd bdp types install` (A4). `bd init` on any
   clone therefore keeps working — re-init is the documented repair path —
   and installs nothing. A provider answering `*storage.ErrUnsupported`
   from the migrations' capability probe is skipped **silently at the
   default verbosity** (a debug-level line only), so gate output is
   byte-identical; any other error fails init like any other step.
3. Writes **no Scope identity** and no witness file; nothing is written to
   `metadata.json` for the graph (A6).
4. Ensures the workspace's `.beads/.gitignore` carries the three new
   entries — `config.local.yaml`, `graph-authority.local.json`,
   `graph-authority.lock` — through `EnsureGitignoreForBeadsDir`; the
   template, `requiredPatterns`, and `trackedRuntimePatterns` all gain them
   (the lock is also covered by the runtime `*.lock` wildcard). Init paths
   that bypass that call today (`--init-if-missing`, an external
   `BEADS_DIR`, `--proxied-server`'s own writer) are not relied on: the
   witness manager **ensures** the entries itself before its first write
   (B3).

No new flag is required. **Registered backends:** `bd init` refuses to
provision them today; their own workspace-creation path owes the same
obligations, proven by the conformance family (B7). Under A8 option A, a
backend that implements the six accessors as `ErrUnsupported` stubs keeps
every existing behavior; under option B nothing changes for it.

### A2. Client wiring — `bd init --bdp-server <url>` and `bd bdp client` (ruling 12, third command; amendment A6)

One more `bd init` target, beside `--server`, `--shared-server`,
`--proxied-server`, `--team-server`, and `--backend`. Every existing target
selects the provider or topology that realizes the storage interfaces — a
choice *below* the normalized abstraction. This one reroutes *above* it, at
the CLI: the `bd bdp` read verbs (A5) become a BDP client of the designated
server instead of opening a store.

**Two files, by what the key is.** `config.yaml` is git-tracked by default
(the tree's own gitignore template says so), so a key that differs per
workspace cannot live there. The tree already has the answer:
**`config.local.yaml`**, merged by viper over `config.yaml` for
"machine-specific settings without polluting tracked config" — merged only
for the project `.beads` and only when a `config.yaml` exists, and
`projectConfigPathFromLoadedState` hard-requires the basename
`config.yaml`, so the writer below gets its own path plumbing.

| Key | File | Values | Notes |
| --- | --- | --- | --- |
| `bdp.scope_url` | `config.yaml` (tracked; yaml-only) | absolute URL | a **project** fact (ruling 7a): what the authority mints/serves; every clone may know it; **settable by `bd config set` only while unminted** — once minted, the URL changes only through `bd bdp promote --rotate-url` / `bd bdp restore`, which update the Scope row and this file as one two-phase transition |
| `bdp.authority_heartbeat` | `config.yaml` | duration (default `30s`) | hazard R (A7); `0` refused there |
| `bdp.authority_heartbeat_grace` | `config.yaml` | count (default `3`) | missed fetches before the server fails closed |
| `bdp.lease_ttl` | `config.yaml` | duration (default `30s`, renewed every third) | hazard S (A7) |
| `bdp.client` | **`config.local.yaml`** | `store` (default) \| `server` | per-workspace; explicit; **not settable from env** (`blockedEnvVars`, the `backend` precedent) |
| `bdp.server` | `config.local.yaml` | absolute URL | the Scope URL of the designated server; `https` required unless loopback or `bdp.insecure_http: true` |
| `bdp.insecure_http` | `config.local.yaml` | bool (default false) | the named waiver; `bd bdp status` reports it |

**Writers.** `bd init --bdp-server <url>` and **`bd bdp client server
--server <url> [--insecure-http]`** / **`bd bdp client store`** write the
per-workspace keys to `config.local.yaml` through one shared writer.
Generic `bd config set` accepts `bdp.scope_url` (unminted only) and the
heartbeat/lease keys (yaml-only routing to `config.yaml`) and **refuses**
`bdp.client` / `bdp.server` / `bdp.insecure_http` with the guidance "use
`bd bdp client`". There is deliberately **no token key in config** (every
key containing `token` is a secret to `IsSecretKey`, and the tracked-config
guard refuses it; other trackers keep tokens in yaml-only config and accept
that — BDP chooses a file). Mechanics: `bdp.` joins the yaml-only prefix
list and `recognizedConfigPrefixes`; a `localOnlyKeys` class names the
three per-workspace keys; `validateYamlConfigValue` gains entries for the
enum, the URLs, the durations, and the count.

**Environment.** Viper binds config keys as `BD_<KEY>` (`SetEnvPrefix("BD")`):
`BD_BDP_SERVER`, `BD_BDP_INSECURE_HTTP`, `BD_BDP_AUTHORITY_HEARTBEAT`
exist without new code; `BD_BDP_CLIENT` is **blocked**. Ruling 7a spells
the Scope URL variable `BDP_SCOPE_URL`; because viper consults
`AutomaticEnv` before a `BindEnv` list (the GH#4645 `BD_ACTOR` precedent),
`BDP_SCOPE_URL` is **read first, explicitly** (the `BEADS_ACTOR` shape),
then `BD_BDP_SCOPE_URL`. Hand-read auth environment is `BEADS_*`, so the
client token file is **`BEADS_BDP_TOKEN_FILE`**. There is **no
`BD_BDP_TOKEN`** (child-process inheritance — the `remote_migrate_gate`
precedent).

**Credential lookup** (client): `BEADS_BDP_TOKEN_FILE` > credentials file
section **`[bdp <origin><scope-path>]`** with `token=`. Transport: no
redirects followed; `Authorization` is sent only to the configured origin.

**Precedence, per command** (a table test pins it):

| Command | Resolution |
| --- | --- |
| `bd init --bdp-server <url>` | flag > `BD_BDP_SERVER` env > existing `config.local.yaml`; writes `bdp.client: server` and `bdp.server` to `config.local.yaml`; touches nothing else |
| `bd bdp client …` | writes `config.local.yaml`; effective for the next command |
| `bd config set bdp.scope_url` | unminted: validates, writes `config.yaml`; minted: refused with the rotate guidance |
| every `bd bdp` verb | env (where permitted) > `config.local.yaml` > `config.yaml`; `bdp.client` (local file only) decides the route |
| `bd bdp status` | prints the resolved route, target, token source (never the token), `insecure_http`, and — on the store route — the identity state row (A3) |

**In `client: server` mode** the verbs' `openBeadGraph*()` accessors return
a BDP-client realization of the same `graphops` role interfaces (A5). A
workspace in this mode still has a local graph store; `bd bdp serve` there
**refuses** unless `--serve-local-store` is passed. Issue verbs are
unaffected in either mode.

### A3. `bd bdp serve` — serving a Scope (rulings 7b, 9, 12; amendments A1, A2, A7)

**Every serving workspace is a SQL-server workspace.** `bd serve` refuses
embedded Dolt permanently (`errServeEmbedded`), and every Dolt-server
topology serves from the unit-of-work provider — so the serving leg is the
UOW leg, and every fence below is implemented there (the store legs carry
the same code for the CLI admin verbs). Embedded workspaces are CLI-only
(A5) and promotable only through a remote (A4).

`bd bdp serve` is a **thin command over the existing `internal/httpapi`
server**: the same storage classification, the same hook peel, the same
`httpapi.Config` — plus a populated `Graph` field. Two policies differ from
`bd serve`: it **requires a Scope this workspace holds** (exit 2 otherwise)
and **it is the only command that mints** (amendment A2: with the Scope URL
a tracked project fact, a plain `bd serve` on any clone would otherwise
become the authority by being first). `bd serve` mounts the rows only when
it holds an already-minted Scope, **converts every graph failure —
capability, identity, fence, heartbeat — into "rows absent + notice" and
keeps the legacy surface up**, and serves exactly what it serves today when
no URL is configured. Because it builds the same server — every server
built from `routeTable` advertises `issues.claim` — it **inherits
`errServeReadonly` wholesale**. W2 decides whether `bd bdp serve` survives
as the strict alias (default: yes; it is the minting path).

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

Behavior, in order:

1. **Classification is `serveDatabaseSource`, verbatim:** registered
   backend → store source; embedded Dolt → **refused, permanently**;
   otherwise the unit-of-work provider. `doltSQLProvider` gains a `beadsDir`
   field (`newSQLServerUOWProvider` receives it today and drops it) and
   `timedProvider` forwards it, so the provider's accessors reach the
   witness manager.
2. **Roles from the same source.** Store arm: `serveIssueRoles`' one peel,
   graph roles off the same `src`. Provider arm: the provider beneath
   `uow.UnwrapProvider` carries the six accessors; `GraphConfig.Reader/
   Types` are nil; `checkDatabaseSource`'s exactly-one-source rule extends
   to them.
3. **Capability probe, all-or-nothing:** the four read-path accessors; any
   `*storage.ErrUnsupported` → `bd bdp serve` **exit 2, typed**; `bd serve`
   → rows absent. Any *other* error: `bd bdp serve` aborts as serve's role
   binding does; `bd serve` → rows absent + notice.
4. **Identity** — the state table, split by command (each row a contract
   case, B7). `bd serve` **never refuses and never mints**.

   | Persisted Scope | Configured URL | Witness (`.beads/graph-authority.local.json`) | `bd bdp serve` | `bd serve` |
   | --- | --- | --- | --- | --- |
   | none | none | — | exit 2 (7a: an explicit URL is required) | no BDP rows, silent |
   | none | set | — | **first serve: `Mint`** (A4's two-phase, fenced sequence; the built-in catalog is installed inside it); then serve honestly empty | no BDP rows; notice "unminted; run `bd bdp serve`" |
   | present | none | any | exit 2 | no BDP rows; notice names the persisted Scope |
   | present | **different** | any | refuse: configured ≠ persisted | no BDP rows; notice |
   | present, matches | matches | **absent** (a clone; a pull into a fresh directory; a directory copied elsewhere — installation key mismatch) | refuse `ErrNotAuthority`; guidance: `bd bdp promote` | no BDP rows; notice |
   | present, matches | matches | present, **pending transition** | run recovery first (B3); then re-evaluate | same |
   | present, matches | matches | present, `(authority_id, epoch)` **stale**, or (hazard S) the lease is held by another holder | refuse; guidance: `bd bdp promote` | no BDP rows; notice |
   | present, matches | matches | present, ledger head `{seq, hash}` **not in the store's ledger** (a database restore, or a different history) | refuse `ErrStateRewound`; guidance: `bd bdp restore` | no BDP rows; notice |
   | present, matches | matches | present, `unverified` set (`bd backup restore` ran) | refuse until `bd bdp restore` clears it | no BDP rows; notice |
   | present, matches | matches | present, consistent | take/renew the lease (hazard S); serve; every operation re-asserts the witness in its transaction (A1) | serve BDP rows |

5. **Host policy:** the Scope URL's host joins the allowlist automatically;
   the listener is plaintext behind TLS termination.
6. **Fencing while serving (A7):** hazard S — the lease is renewed every
   third of `bdp.lease_ttl` through `RunTxEphemeral` (no history) and
   asserted inside every transaction (a read selects it; a mutation
   `UPDATE`s it with one affected row required); losing it fails closed.
   Hazard R — the server fetches the remote tracking ref on
   `bdp.authority_heartbeat`, compares `(authority_id, epoch)`, and after
   `bdp.authority_heartbeat_grace` missed fetches, or any change, stops
   serving BDP rows with `ErrNotAuthority` (`bd serve` keeps the legacy
   surface; `bd bdp serve` exits 3). Both hazards → both.
7. **Mounting:** `serveListen(opts, httpapi.Config{…, Graph: …})`. Liveness
   stays `GET /healthz`; BDP readiness is a real discovery read.
8. **Lifecycle:** excluded from the post-command maintenance net by
   `commandPolicy` (Part A), not by sharing the leaf name `serve`; the
   events-journal maintenance ticker applies as for `bd serve`.

### A4. `bd bdp promote` / `bd bdp restore` / `bd bdp ledger` / `bd bdp types install` (rulings 9, 11; amendments A5, A7)

These are the **only** reachers of `BeadGraphAdmin()` and
`BeadGraphTypeInstaller()`; the server assembly has no field for either.
They run under the **workspace exclusive gate**. Every transition is
**two-phase** (B3): `Begin` writes a pending record to the witness file
(after ensuring the ignore entries and taking the lock), the fenced
transaction commits and — where hazard R applies — publishes, then
`Finalize` writes the new witness; a crash in between is recovered on the
next `Load` by asking the store (and, on hazard R, the remote-tracking
ref) whether the transition committed.

- **`Mint`** (`bd bdp serve`'s first serve): precondition *no Scope row*;
  one transaction: INSERT the singleton Scope row, seed
  `graph_ledger_seq`, append `mint`, install the built-in catalog with its
  `install` events (W3 supplies it; empty until then); hazard S: take the
  lease; scoped commit; hazard R: the publish sequence below.
- **`bd bdp promote`** — make *this workspace* the authority for the Scope
  it carries: precondition a consistent Scope row. Hazard S: take the lease
  if expired (or `--steal`, operator-confirmed, which revokes the holder's
  row) and CAS the epoch (`epoch = epoch + 1 WHERE epoch = <read>`; a lost
  race is a serialization loser, replayed into a typed refusal) with a
  `promote` event. Hazard R — the **publish sequence**: record the local
  HEAD; `DOLT_FETCH`; require the remote-tracking HEAD to be an **ancestor**
  of local HEAD (`DOLT_MERGE_BASE`); the fenced transaction; **scoped
  commit** (`DOLT_ADD` of the graph tables, then `DOLT_COMMIT` with a named
  message — never `-Am`, which sweeps the working set); `DOLT_PUSH`. Push
  outcomes, classified by a typed `IsNonFastForward` (pinned by a
  real-Dolt test — `pushToRemote` wraps every failure untyped today and the
  only classifier in the tree matches "no common ancestor"): **non-FF** →
  fetch and re-read the remote Scope row: `(authority_id, epoch)` changed →
  `DOLT_RESET --hard <recorded local HEAD>` (only this operation's commit;
  never the remote ref), `Abandon`, refuse "another promotion landed; pull
  first"; unchanged (issue-plane divergence only) → keep the commit,
  `ErrSyncRequired`, retryable after `bd dolt pull`; **any other failure**
  → keep the commit, touch nothing, report. Both hazards → both.
  **Neither hazard** (embedded, no remote): **not promotable in place** — a
  copy of such a database is an unrelated second copy nobody can observe,
  so `bd bdp promote` there requires `--rotate-url <new>` (a new Scope).
  `--rotate-url` on any topology additionally rotates the URL (operator-
  confirmed) and updates tracked `config.yaml` in the same transition.
- **`bd bdp restore`** — runs after a database restore. `bd backup restore`
  (`runBackupRestore`; also reached by `bd bootstrap`'s restore action)
  calls `Admin.MarkUnverified` after `RestoreDatabase` and before its own
  commit — a **no-op without a witness**. Independently of that flag, the
  ledger-head check (A3 row 8) refuses a store whose ledger does not
  contain the witness's `{seq, hash}`. `bd bdp restore` is the recovery
  path and is therefore **exempt from the head check**: its precondition is
  lineage match under the exclusive gate, and it branches on the provider's
  declaration:

  | `LedgerDurability` | Meaning | `bd bdp restore` does |
  | --- | --- | --- |
  | `in-state` (Dolt, v0) | the ledger is ordinary versioned state | requires a ledger snapshot whose range reaches the witness's `{seq, hash}` (`--ledger <file>`, applied through `LedgerApply`'s recovery predicate) to show continuity; **otherwise rotates** the Scope URL (operator supplies the new one) and epoch, appends `refuse_url` + `rotate`, updates `config.yaml`, and regrants the witness |
  | `independent` | the provider keeps the ledger outside the restored state (contract case) | re-validates (the ancestry check, B3) and regrants; no rotation |
  | `none` | no ledger | always rotates |

  Never silent. Residuals, stated: a whole-directory filesystem snapshot
  rewinds witness and state together; a `dolt` CLI or `bd sql` restore sets
  no flag — the ledger-head check is what catches it.
- **`bd bdp ledger snapshot [--from SEQ] [--to SEQ]`** / **`bd bdp ledger
  apply <file>`** — the hash-chained `graph_ledger_events` as JSONL under a
  **manifest** `{scope_url, authority lineage, first_seq, last_seq,
  prev_hash of first, head_hash}`. `LedgerApply`'s **recovery predicate**
  (exempt from the head check): exclusive gate; lineage matches the Scope
  row; the store's current head equals the manifest's predecessor; every
  event's `hash` verifies; then append, re-derive `graph_allocations` and
  `graph_scope_history`, and regrant the witness. A gap or a foreign
  lineage is refused. It never rewinds a tombstone or a refusal.
- **`bd bdp types install <file>`** — the post-mint catalog change (P1;
  W3 emits the file): a fenced mutation with `install` events, idempotent
  by fingerprint, closure-validated; an owning declaration without `Max` is
  refused.

### A5. Graph verbs (v0 reads; P3 writes; names fixed now)

```text
bd bdp bead get <path> | list [--type URL] [--after CURSOR] [--limit N]
bd bdp link get <path> | list [--type URL] [--source PATH] [--target REF] [--after CURSOR] [--limit N]
bd bdp types [get <url>] | types install <file>
bd bdp status
bd bdp client store | server --server URL [--insecure-http]
bd bdp serve | promote [--rotate-url URL] [--steal] | restore [--ledger FILE] | ledger snapshot|apply
(P3)  bd bdp bead create|update|delete ; bd bdp link create|update|delete
```

Each verb reaches its role through an accessor and nothing else, on
whichever route the workspace is on — the `cmd/bd/label.go` pattern plus
the third route, with a route-fork test in the shape of the
`*_proxied_integration_test.go` / `*_embedded_test.go` pairs (e.g.
`label_proxied_integration_test.go`):

```go
func openBeadGraphReader() (graphops.Reader, error) {
    if bdpClientMode() == "server" {        // config.local.yaml / config.yaml only, per A2
        return bdpclient.Reader(bdpClientConfig())          // internal/bdpclient
    }
    if usesProxiedServer() { return proxiedBeadGraphReader() }   // the UOW provider; no new HTTP route
    return store.BeadGraphReader()
}
```

CLI reads on the authority workspace read the authority's own state; on
any other workspace they refuse (`ErrNotAuthority`) — there is no
"replica read" in v0 (ruling 9 defers replica labeling to BDP).
`internal/bdpclient` speaks the pinned wire and implements `graphops.Reader`
and `graphops.DescriptorReader`; a BDP Problem maps back to the same typed
errors (round-trip test). **Until P2's cursor ADR lands there are no
collection routes**: on the client route `bead list` / `link list` answer
`ErrNotServedYet`; on the store route they work from P1.

## Part B — Storage interfaces

### B1. Accessors on `storage.Storage` (architecture §2a; amendments A4, A8 — option A wording)

Named **`BeadGraph*`**: `Storage.GraphCounter()` already exists and counts
the *issue* dependency graph.

```go
// BeadGraphReader returns the guarded graph-read surface for this store:
// Bead and Link records (a Bead with its complete, grouped, bounded
// ownedLinks, assembled in the call's one transaction), keyset selections
// in code-unit order under an opaque cursor, and incident Links. The
// returned role is manager-backed: every call reloads this workspace's
// authority witness and asserts it inside its transaction; no request
// carries authority. Reads fire no hooks; the hook decorator recurses.
BeadGraphReader() (graphops.Reader, error)

// BeadGraphTypes returns the Type Descriptor inventory (ordered, bounded,
// keyed); protected like every read.
BeadGraphTypes() (graphops.DescriptorReader, error)

// BeadGraphTypeInstaller returns the post-mint descriptor install/converge
// role: a fenced mutation with ledger events, reached only by
// bd bdp types install and the conformance fixture.
BeadGraphTypeInstaller() (graphops.TypeInstaller, error)

// BeadGraphIdentityReader returns the Scope row, this workspace's witness
// claim, and the provider's LedgerDurability declaration. Exempt: it
// reports state.
BeadGraphIdentityReader() (graphops.IdentityReader, error)

// BeadGraphBootstrapper returns the one-time mint (two-phase, fenced per
// hazard; installs the built-in catalog). Held only by bd bdp serve's
// first-serve path.
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
`configStore` stub, and every other type compiled as a `Storage` — which,
under option A, **the compiler catches** (a required method fails every
such type); the three censuses (B5) cover the decorators and facades.
**This is the source break `storage.go` declares for out-of-tree backends**
(A8 option A): six one-line `ErrUnsupported` stubs, called out in CHANGELOG
as the joint `ReadyClaimer`/`BatchCloser` entry was ("must add both methods
to compile"). Under option B, B1 becomes a `BeadGraphCapable` interface plus
explicit decorator implementations, a capability census, and a resolver.

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
type BeadPage struct{ Items []BeadRecord; Next Cursor }
type LinkPage struct{ Items []Link; Next Cursor }

type Reader interface {
    Bead(ctx, BeadRequest) (BeadRecord, error)
    Link(ctx, LinkRequest) (Link, error)
    Beads(ctx, BeadSelectRequest) (BeadPage, error)       // WHERE path > last ORDER BY path LIMIT n, binary-collated column
    Links(ctx, LinkSelectRequest) (LinkPage, error)
    IncidentLinks(ctx, IncidentRequest) (LinkPage, error) // one UNION over (source, target) indexes, ordered, limited
}
type DescriptorReader interface {
    Descriptors(ctx) ([]TypeDescriptor, error)            // ordered by URL; bounded by MaxCatalog
    Descriptor(ctx, DescriptorRequest) (TypeDescriptor, error)
}
type TypeInstaller interface {
    Install(ctx, InstallRequest) (InstallResult, error)   // post-mint; fenced; idempotent by fingerprint; closure validated
}
type IdentityReader interface {
    Read(ctx) (ScopeIdentity, error)                      // Scope row + witness claim (Held, Epoch, LedgerSeq, Unverified, Pending)
    LedgerDurability(ctx) (LedgerDurability, error)       // in-state | independent | none (ruling 11)
}
type ScopeBootstrapper interface {
    Mint(ctx, MintRequest) (ScopeIdentity, error)         // two-phase; fenced per hazard; catalog installed inside
}
type Admin interface {
    Promote(ctx, PromoteRequest) (ScopeIdentity, error)   // two-phase; fenced per hazard; RotateURL optional
    Rotate(ctx, RotateRequest) (ScopeIdentity, error)     // refuse_url(old) + rotate(new), one transaction; config.yaml updated
    LedgerSnapshot(ctx, LedgerRange) (LedgerManifest, []LedgerEvent, error)
    LedgerApply(ctx, LedgerManifest, []LedgerEvent) (LedgerApplyResult, error)   // recovery predicate; regrants
    MarkUnverified(ctx) error                             // no-op without a witness
    ClearUnverified(ctx) error
}
```

**Bounds.** `MaxExpandedRows` (10,000) caps `len(page) + Σ ownedLinks` for
any read; the batched owned-links query carries **`LIMIT (MaxExpandedRows −
rows already materialized) + 1`**; descriptor decoding is capped the same
way; `ErrRepresentationTooLarge` is typed. Owning Types **must** declare
`Max`.

**Statement budgets, per role method** (the contract pins each; never a
per-row statement; a validation run has its own budget and happens outside
the read's transaction):

| Method | Statements | Composition |
| --- | --- | --- |
| `Bead` | ≤ 7 (7 cold, 6 on a descriptor-cache hit) | Scope row; ledger head; lease (hazard S); state version; row; descriptors; batched owned links |
| `Beads` page | ≤ 7 | same, with the page in place of the row |
| `Link`, `IncidentLinks`, `Links` page | ≤ 5 | Scope row; ledger head; lease; state version; the query (one `UNION` for incident) |
| `Descriptors`, `Descriptor` | ≤ 5 | Scope row; ledger head; lease; state version; catalog |
| `IdentityReader.Read` | 3 | Scope row; ledger head; lease |

**Values** (`Bead`, `Link`, `Ref`, `Properties`, `Revision`, `Attribution`,
`TypeDescriptor`, `OwnedLinkDecl`, `ScopeIdentity`, `LedgerEvent`,
`LedgerManifest`) have unexported fields and constructors that enforce the
laws in `laws.go`. `Properties` is the immutable raw-JSON object value from
the plan; its canonical bytes are what B4 stores. `Ref` is a sum. `Revision`
is 128 bits from `crypto/rand`, lower-hex. A `LedgerEvent` carries `{seq,
kind, path|scope_url, revision?, fingerprint?, authority_id, epoch, at,
prev_hash, hash}` with `hash = sha256(canonical(event without hash))`.

**Errors:** `beadserrors` (stdlib-only) declares the new sentinels
(`ErrNoScope`, `ErrScopeExists`, `ErrNotAuthority`, `ErrStateRewound`,
`ErrStateChanged`, `ErrSyncRequired`, `ErrURLReused`,
`ErrRepresentationTooLarge`, `ErrNotServedYet`) and the typed
`GoneError{Path, State}`; `graphops` aliases them so `errors.Is` crosses the
module boundary.

### B3. Bodies, the witness manager, and legs

`internal/storage/graphops` bodies take **`DBTX`** (the `issueops.DBTX`
shape, declared locally), which `*sql.Tx` and `domain/db.Runner` both
satisfy, **and the witness** the accessor loaded:

```go
func ReadBeadInTx(ctx, tx DBTX, w authority.Witness, req graphops.BeadRequest) (graphops.BeadRecord, error)
```

**`internal/storage/authority` — the witness manager** (no SQL):

- **Installation key.** A random id generated once at
  `os.UserConfigDir()/bd/installation-id` (the user-global config precedent
  in `yaml_config.go`) — never the hostname, which the tree's `NodeID`
  documentation refuses for containers, DHCP/macOS, and shared servers;
  `InstallationKey = sha256(id ":" realpath(.beads))`. A workspace moved on
  the same machine changes its key and needs a `bd bdp promote` (stated in
  the refusal's guidance as "moved or copied"; the shared-server lease and
  the remote fence — not the key — are what catch a same-path copy on
  another machine).
- **`Load`** is a plain read: atomic rename makes the file complete or
  absent, so no shared lock is taken (it would be released before the
  transaction anyway). A pending transition record triggers **recovery**
  before any assertion: the store (Scope row, ledger head) and, on hazard
  R, the remote-tracking ref answer whether the transition committed;
  `Finalize` or `Abandon` accordingly.
- **`Advance`** takes the exclusive lock on `.beads/graph-authority.lock`
  with a **bounded poll** (`internal/lockfile` has no timeout API; the
  `workspacegate` gate's poll is the precedent; both `ErrLocked` and
  `ErrLockBusy` sentinels are honored), reads, applies, **never lets
  `LedgerSeq` decrease**, writes through `internal/atomicfile` (which fsyncs
  the file, not the directory), and fsyncs the directory.
- **`Begin` / `Finalize` / `Abandon`** — the two-phase record for mint,
  promote, rotate, and ledger apply. `Begin`'s preflight **ensures** the
  three `.beads/.gitignore` entries (`EnsureGitignoreForBeadsDir`) and
  refuses if the witness path is git-tracked (`git ls-files
  --error-unmatch`); an already-tracked witness is also a hard doctor
  error (`trackedRuntimePatterns`).
- **Order of a mutation:** the DB transaction commits (and publishes, on
  hazard R) *before* the witness advances, so the file is never ahead of
  durable state; a crash between the two leaves the witness behind, which
  the next assertion tolerates and the next mutation catches up.
  Transitions are the exception — they change identity — hence the pending
  record. Residual (stated, P3): a write acknowledged before the crash and
  never witnessed is invisible to a restore that lands before the next
  advance.

| Leg | Files | Body |
| --- | --- | --- |
| server Dolt (CLI) | `internal/storage/dolt/beadgraph_*.go` | accessor loads the witness, wraps the body in `withReadTx` (reads) / `withRetryTx` (mutations) with a **scoped** commit (`DOLT_ADD` graph tables + `DOLT_COMMIT`); hazard R transitions run the publish sequence |
| embedded Dolt (CLI only) | `internal/storage/embeddeddolt/beadgraph_*.go` (`//go:build cgo`) | same body, `withConn` |
| unit of work (**the serving leg**) | `internal/storage/domain/beadgraph.go` (`BeadGraphUseCase`), `internal/storage/domain/db/beadgraph.go` (over `db.Runner`; the publish sequence uses the `DOLT_FETCH`/`DOLT_PUSH` procedures already on `doltVersionControlSQLRepository`), `internal/storage/uow/beadgraph_*.go` (`uow.UnitOfWork.BeadGraphUseCase()`; `RunTxRead`; `RunTxResult` with commit messages `bdp: mint scope <url>` / `bdp: promote epoch <n>` / `bdp: install types <fingerprint>` / `bdp: ledger apply <from>-<to>` and scoped staging; `RunTxEphemeral` for lease renewal) | **same body** |

Contract headers say **"one reading plus an engine check"**. Every
protected body begins with `assertAuthorityInTx(ctx, tx, w, mutating)`:
Scope row identity; ledger head `{seq, hash}` present (exact prefix) and
`MAX(seq) >= w.LedgerSeq`; on hazard S the `graph_authority_lease` row —
**a read `SELECT`s it; a mutation `UPDATE`s `heartbeat_at … WHERE id = 1
AND holder = self AND epoch = self` and requires one affected row** (the
heartbeat-upsert shape in `issueops/lease.go`), so a takeover between a
SELECT and a commit is a serialization loser that `withRetryTx`/
`RunTxResult` replay into a refusal; the **graph-state version** — the
ordered `DOLT_HASHOF_TABLE()` hashes of the seven graph tables (not
`DOLT_HASHOF_DB()`, which every ephemeral write, including this design's
own lease heartbeat, moves) — equal to `w.StateVersion`, else the body
returns **`ErrStateChanged` without validating inside the held
transaction**; the accessor then runs `ValidateStateInTx` under
**singleflight** in its own transaction (ancestry check
`DOLT_MERGE_BASE(w.StateCommit, HEAD)` so a rewind a ledger-independent
provider would miss is refused; provenance identifies foreign updates),
advances the witness, and retries once. Providers without table hashes
declare a `GraphStateVersion` capability or fail closed. Exempt from the
head check, with their own preconditions (architecture §4): `Mint`,
`Promote`, `Rotate`, `LedgerApply`, `IdentityReader`, the witness-file
operations. `SeedBeadInTx`/`SeedLinkInTx` (P1 fixture writer) allocate
through the ledger like a real write; every call site is a `_test.go` file.

### B4. Schema (migrations; frozen once merged)

Rules the tree enforces: migrations are **frozen once merged** — hygiene
check C forbids editing a shipped file (a git-diff check), and the runtime
`content_skew.go` compares `schema_migrations.content_hash` across clones;
**no `NOW()`/`UUID()`/`RAND()`** in migration SQL (check B) — timestamps and
ids come from Go; real-Dolt tests for anything a `sqlmock` echo cannot
exercise; DDL is not transactional across statements, so each `CREATE` is
guarded and resumable. **Eight replicated tables in five files** —
`NNNN_beadgraph_scope.up.sql` (scope, history), `NNNN_beadgraph_types.up.sql`,
`NNNN_beadgraph_beads.up.sql`, `NNNN_beadgraph_links.up.sql`,
`NNNN_beadgraph_ledger.up.sql` (events, counter, allocations) — plus the
lease's three parts: its name in `doltIgnorePatterns`, a main-series
`NNNN_beadgraph_authority_lease.up.sql` that creates it for existing
workspaces (the 0055 `__temp__` + conditional `RENAME` shape), and
`ignored/NNNN_beadgraph_authority_lease.up.sql` for fresh clones (check D).
The lease is reached exactly as `issueops/lease.go` reaches `leases`: on
the default branch's working set (branch-qualified sessions do not see it).

**Collation.** Dolt's default collation is already binary
(`utf8mb4_0900_bin` — probed), and no migration in the tree declares one.
Every identifier column below still carries **`CHARACTER SET utf8mb4
COLLATE utf8mb4_bin`** (written `BIN`) as the defense for providers whose
default is case-insensitive, with a contract case.

**Identity is Scope-relative.** Rows store the canonical Scope-relative
`path`; the absolute URL is `scope_url + path`, computed at the boundary,
so a URL rotation rewrites no rows.

**JSON is bytes.** `properties` and `descriptor` are canonical JSON bytes in
`LONGBLOB`, never the engine `JSON` type (the tree measured `1.0`→`1`,
integers past 2^53 rounded, `1e300` expanded in `metadata_cas.go`;
`-0.0`→`0` per the role guide). Size limit 1 MiB per value.

**Provenance on every mutable row.** `last_authority_id` / `last_epoch` are
stamped by every mutation on descriptors, beads, links, **and
allocations**, so the validator can identify a foreign *update*.

| Table | Columns (type; nullability) | Keys / constraints |
| --- | --- | --- |
| `graph_scope` | `id TINYINT NOT NULL` (always 1), `scope_url VARCHAR(2048) BIN NOT NULL`, `authority_id CHAR(32) NOT NULL`, `epoch BIGINT UNSIGNED NOT NULL`, `minted_at DATETIME(6) NOT NULL` | `PRIMARY KEY (id)`, `CHECK (id = 1)` — singleton |
| `graph_scope_history` | `scope_url VARCHAR(2048) BIN NOT NULL`, `refused_seq BIGINT UNSIGNED NOT NULL`, `refused_at DATETIME(6) NOT NULL`, `reason VARCHAR(64) NOT NULL` | `PRIMARY KEY (scope_url)`; derived from `refuse_url` events |
| `graph_type_descriptors` | `url VARCHAR(2048) BIN NOT NULL`, `descriptor LONGBLOB NOT NULL`, `fingerprint CHAR(64) NOT NULL`, `installed_seq BIGINT UNSIGNED NOT NULL`, `installed_at DATETIME(6) NOT NULL`, `last_authority_id CHAR(32) NOT NULL`, `last_epoch BIGINT UNSIGNED NOT NULL` | `PRIMARY KEY (url)`; `UNIQUE (fingerprint)` |
| `graph_beads` | `path VARCHAR(1024) BIN NOT NULL`, `type_url VARCHAR(2048) BIN NOT NULL`, `revision CHAR(32) NOT NULL`, `attribution_principal VARCHAR(512) NULL`, `attribution_status ENUM('claimed','unknown') NULL`, `properties LONGBLOB NOT NULL`, `last_authority_id CHAR(32) NOT NULL`, `last_epoch BIGINT UNSIGNED NOT NULL`, `created_at DATETIME(6) NOT NULL`, `updated_at DATETIME(6) NOT NULL` | `PRIMARY KEY (path)`; `INDEX (type_url, path)`; `FOREIGN KEY (type_url) REFERENCES graph_type_descriptors(url)`; attribution columns both NULL or both set |
| `graph_links` | `path VARCHAR(1024) BIN NOT NULL`, `type_url … BIN NOT NULL`, `revision CHAR(32) NOT NULL`, `source_path VARCHAR(1024) BIN NOT NULL`, `source_pin CHAR(32) NULL`, `target_kind ENUM('in','ext') NOT NULL`, `target_path VARCHAR(1024) BIN NULL`, `target_url VARCHAR(2048) BIN NULL`, `target_pin CHAR(32) NULL`, `attribution_*`, `properties LONGBLOB NOT NULL`, `last_authority_id`, `last_epoch`, timestamps | `PRIMARY KEY (path)`; `INDEX (source_path, type_url, path)`, `INDEX (target_path, path)`; `FOREIGN KEY (source_path) REFERENCES graph_beads(path)`; `CHECK` exactly one of `target_path`/`target_url` per `target_kind`; **no** uniqueness on (type, source, target) |
| `graph_ledger_seq` | `id TINYINT NOT NULL` (always 0), `next BIGINT UNSIGNED NOT NULL` | `PRIMARY KEY (id)` — **the single-row sequence counter** (the `bd_events_seq` precedent): `UPDATE … SET next = next + 1 WHERE id = 0` inside the mutation's transaction; two allocators contend on the row and only one commit order survives; a rolled-back transaction burns no seq; seeded by `Mint` |
| `graph_ledger_events` | `seq BIGINT UNSIGNED NOT NULL`, `kind ENUM('mint','install','update','promote','rotate','allocate','tombstone','refuse_url') NOT NULL`, `path VARCHAR(1024) BIN NULL`, `scope_url VARCHAR(2048) BIN NULL`, `resource_kind ENUM('bead','link') NULL`, `revision CHAR(32) NULL`, `state ENUM('pruned','erased') NULL`, `fingerprint CHAR(64) NULL`, `authority_id CHAR(32) NOT NULL`, `epoch BIGINT UNSIGNED NOT NULL`, `at DATETIME(6) NOT NULL`, `prev_hash CHAR(64) NOT NULL`, `hash CHAR(64) NOT NULL` | `PRIMARY KEY (seq)` — **append-only, hash-chained**; `UNIQUE (hash)`; `INDEX (path, seq)`; `CHECK` per kind. **Every mutation is an event**, so the witness's head covers every graph-state change |
| `graph_allocations` | `path VARCHAR(1024) BIN NOT NULL`, `resource_kind ENUM('bead','link') NOT NULL`, `birth_seq BIGINT UNSIGNED NOT NULL`, `birth_authority_id CHAR(32) NOT NULL`, `birth_epoch BIGINT UNSIGNED NOT NULL`, `state ENUM('live','pruned','erased') NOT NULL`, `tombstone_seq BIGINT UNSIGNED NULL`, `last_authority_id CHAR(32) NOT NULL`, `last_epoch BIGINT UNSIGNED NOT NULL` | `PRIMARY KEY (path)` — the O(1)/O(log n) ID test (ruling 3); **derived state**, re-derivable from the events |
| `graph_authority_lease` (**dolt-ignored**, never replicates) | `id TINYINT NOT NULL`, `holder_installation_key CHAR(64) NOT NULL`, `holder_nonce CHAR(32) NOT NULL` (per process), `epoch BIGINT UNSIGNED NOT NULL`, `granted_at DATETIME(6) NOT NULL`, `expires_at DATETIME(6) NOT NULL`, `heartbeat_at DATETIME(6) NOT NULL` | `PRIMARY KEY (id)`, `CHECK (id = 1)`; the hazard-S fence (A7); renewals through the ephemeral commit form — the `leases` (bd-lrgn1) precedent |

`updated_at` is protocol-irrelevant bookkeeping. Bead `type_url` and Link
`source_path`/`target_*` are immutable after insert.

**The witness file: `.beads/graph-authority.local.json`.**
`{installation_key, scope_url, authority_id, epoch, ledger_seq,
ledger_hash, state_version, state_commit, unverified, granted_at,
pending?}`. Written only by the manager (B3). What each operation does to
it: `git clone` / `dolt clone` — absent; pull — untouched; `bd backup
restore` / `DOLT_BACKUP` restore — present, and the ledger head it names is
no longer in the store → `ErrStateRewound`; directory copy to another
path or machine — present, installation key mismatch → `ErrNotAuthority`;
a copy to the same path on another machine — the key matches, and the
lease (hazard S) or the remote fence (hazard R) is what refuses it; an
embedded no-remote copy is the unrelated second copy A7 declares
non-promotable in place.

**Cross-repo coupling (bts).** `DoltTeamServer` workspaces refuse to open
when `current < latest` with **no `BD_IGNORE_SCHEMA_SKEW` hatch** — a
**numeric-version** comparison only. The coupling is a cross-repository
**release-parity gate**: bts must ship byte-identical copies of the six
main-series files and the ignored-series file (the `schema_migrations.content_hash`
values `migration_content_hashes.go` reads are what a bts-side parity test
compares). The migration PR is sequenced with bts; the remote-migrate gate
(#4259) forces migrate-vs-adopt on every remote-backed workspace at upgrade.

### B5. Decorators, censuses, and every embedding surface

- `internal/storage/hook_beadgraph_*.go` (six files): declared, recurse
  **unwrapped**; `storage.RoleFiresHooks` is a type switch over hook
  wrappers, so an unwrapped role needs no entry; a test asserts each graph
  role answers `false`. Added to `role_accessor_decorator_test.go`'s table.
- `internal/telemetry/beadgraph_*.go`: every method spanned; the
  **telemetry census** gains the classification.
- **Three censuses**, each must learn `graphops`: the storage reflection
  census, the telemetry census, and the conformance package's
  **source-parsed** `facadePackages` map (`role_coverage_scan_test.go`) —
  without the third, `TestRoleFacadeCensusAgreesWithReflection`
  (`role_coverage_gate_test.go`) fails.
- `internal/storage/uow/notifying.go` (explicit accessors, parity test);
  `internal/httpapi/claim.go`'s `timedProvider` (forwarding `beadsDir`);
  `cmd/bd/serve.go`'s `serveRoleSource` and its stubs;
  `internal/jira/tracker_test.go`'s `configStore` stub — every surface that
  embeds the store or a provider, enumerated by
  `grep -l 'func (.*) Memories()'` at implementation; under option A the
  compiler finds every other type compiled as a `Storage`.

### B6. `backend/` public surface and depguard

- **No aliases** (amendment A4). `TestPublicSurfaceComplete` stays green
  *because* no `internal/` type is reachable from the new accessors — a
  test asserts that.
- `backend/backend.go`'s "minimal external backend" example gains the six
  accessors as `ErrUnsupported` stubs (option A); the **CHANGELOG entry**
  follows the joint `ReadyClaimer`/`BatchCloser` entry's wording.
- `.golangci.yml` gains a **new, stricter** rule (cmd/bd imports the
  `issueops` tx-body package directly today, so this is not the existing
  convention): `internal/storage/graphops` is importable only by
  `internal/storage/{dolt,embeddeddolt,domain/db}` and its own tests. A
  mutation test **deletes the deny entry and asserts that a fixture
  violating it then passes lint** — which proves the entry is what fails
  the violation.

### B7. Conformance

- Families: `beadgraph_reader_contract.go`, `beadgraph_types_contract.go`
  (reader + installer), `beadgraph_identity_contract.go` (reader,
  bootstrapper, admin), each citing the leaf doc by line.
- `RoleContractBundle` gains six factory fields **and** their rows in
  `role_bundle_cases.go`; `BeadGraphFixture` carries the seed hook, a
  `Witness` hook (a temp workspace directory and installation id standing
  in for `.beads/` and the user config dir), and a `Remote` hook (a temp
  Dolt remote for hazard-R cases).
- Wirings on all three legs; the leg registry
  (`internal/storage/contract_leg_registry_test.go`) and
  `TestEveryLegWiresEveryRoleContract` see them; both coverage gates apply.
- Non-capable stores answer `*storage.ErrUnsupported{Op: "<accessor
  name>"}` — the six strings pinned — proven per accessor.
- Cases the councils asked for by name: a clone produced by push/pull
  refuses; a **`DOLT_BACKUP` restore of an authority** refuses
  (`ErrStateRewound`); a **copied witness** in another directory refuses;
  a **lease takeover between a mutation's SELECT and its commit** is a
  serialization loser that refuses; two clones minting before either pulls
  (hazard R: the second push is non-FF and undone); concurrent mint on one
  database (one wins); a promotion race (one CAS wins; the loser's push is
  non-FF, reset to its recorded pre-op HEAD, issue-plane commits kept); a
  **non-FF caused by issue-plane divergence alone** answers
  `ErrSyncRequired` and keeps the commit; a **network failure on push**
  keeps the commit and touches nothing; heartbeat detects a changed
  `(authority_id, epoch)` and fails closed after the grace; a **pending
  transition** is recovered on the next load (both outcomes); rotation
  refuses the old URL and updates `config.yaml`; case-differing paths
  distinct and code-unit ordered; ownedLinks completeness incl. **empty
  groups** and the bound under `LIMIT remaining+1`; keyset continuation
  inside one transaction; gone-family; a promote in another process is
  honored by the next read; descriptor read on a non-authority clone
  refuses; `bd init` re-run on any clone succeeds and installs nothing;
  `Mint` installs the built-in catalog with `install` events; ledger
  snapshot/apply round trip, gap refusal, foreign-lineage refusal,
  recovery-predicate regrant; the installer refuses an owning declaration
  without `Max`; statement budgets per method; `ErrStateChanged` triggers
  one validation under concurrent reads and the reads retry once; a
  heartbeat does **not** change the graph-state version.
- Differential-gate rows: every legacy form of `bd link`, `bd graph`,
  `bd graph check`, `bd restore`, `bd promote` parses and behaves as before;
  `bd init` gate output on a non-capable backend is byte-identical;
  `bd serve` without a Scope URL is byte-identical.

### B8. `httpapi` integration and the pinned wire (amendments A2, A7)

- **`httpapi.Config.Graph *GraphConfig`** — `{Reader graphops.Reader; Types
  graphops.DescriptorReader; ScopeURL string; Fence FenceSource}` on the
  store arm; on the provider arm `Reader`/`Types` are nil and the provider's
  own accessors answer per request through `timedProvider`. `Fence` is the
  hazard watcher (lease renewal and/or remote fetch). No admin or installer
  field (a test asserts it); `checkDatabaseSource`'s exactly-one-source rule
  extends to the graph fields.
- **`bdpRouteTable`** (`internal/httpapi/bdp_routes.go`) — **P2**: rows in
  the same `route` shape, registered by `Server.handler()` **only when
  `cfg.Graph != nil`** — `handler()` reads no `Config` today, so this is a
  first, and `TestSpecRouteParity`'s `(&Server{}).handler()` keeps excluding
  the rows — each wrapped by the same `s.route(rt)`. First rows: discovery,
  `types/`, one Bead, one Link; **collection rows wait for the cursor ADR**.
  No capability token in v0; a sibling parity test compares `bdpRouteTable`
  against the pinned schema's path grammar.
- **Posture parity test:** one refusal matrix drives a legacy row and a BDP
  row and asserts identical status and log shape.
- **Handler = serializer**; typed graph errors → BDP Problem records
  (`bdp_problem.go`), here and only here.
- **Wire — P0, not yet in the tree:** `internal/httpapi/bdpwire/schema/bdp-v0.schema.json`
  **will be vendored** with a `PROVENANCE` file (upstream repo, commit — the
  plan's §0 pin — and sha256); `make bdp-gen` runs a **pinned** JSON-Schema→Go
  generator (fallback recorded at P0: hand-written DTOs validated against
  the schema); `make bdp-check` regenerates and diffs, and
  `scripts/ci/pr-policy.sh` runs it beside `make api-check`.

## Part C — What does not change

`storage.Storage`'s existing 28 accessors and every `issueops` role; the
journal's frozen vocabulary; `openapi.v0.yaml` and `TestSpecRouteParity`;
`bd serve` on a workspace with **no** `bdp.scope_url` (byte-identical) and
on any workspace that does not hold the authority (legacy surface up, rows
absent); every legacy CLI verb (differential gate rows); JSONL export
shapes; `metadata.json`'s schema; `bd init` gate output on non-capable
backends.

## Part C2 — What changes that an earlier draft claimed did not

- **Merge, pull, and sync.** Every `DOLT_PULL`/`DOLT_MERGE` route (the
  `versioncontrolops` exports — seven in `mergesettle.go`, plus
  `fastforward.go` and `automerge.go`; `doltCLIPull`, `Pull`, `PullRemote`,
  `pullTransport`/`pullWithAutoResolve` in `internal/storage/dolt/store.go`;
  the UOW leg's `doltVersionControlSQLRepository`; embedded federation sync;
  the remote-migrate gate's fast-forward `DOLT_MERGE`) can change graph
  state outside the roles, so the **state-change validator** runs on every
  observed graph-state-version change (B3) and refuses a foreign-authority
  or invalid delta; a superseded clone that pulls resets to the remote. The
  replication/merge ADR specifies its rules before the migrations land.
- **Federation.** Graph tables ride filtered pushes **unfiltered, by
  decision, in v0**; the lease table never replicates.
- **`bd sql`, raw SQL, and force-push.** Out of contract for graph tables;
  the enforcement-boundary ruling decides the rest before P3.
- **`bd backup restore`** calls `Admin.MarkUnverified` after
  `RestoreDatabase`, before its commit (no-op without a witness).
- **Root store policy** gains `commandPolicy` (Part A).
- **`bd init`** gains the three `.beads/.gitignore` entries and installs no
  descriptors (the catalog moves to `Mint`).
- **`bd config set bdp.scope_url`** is refused once minted.

## Part D — Open implementation questions (not rulings)

1. `MaxExpandedRows`, `MaxCatalog`, page default/max, the 1 MiB value
   limit, the lease TTL, the heartbeat default and grace are proposed
   numbers; the leaf doc fixes them at P1 with the rationale.
2. Whether the P1 fixture writer stays fixture-only through P3 or becomes
   the internal half of `graphops.Writer`.
3. Generator choice for `bdpwire` — recorded at P0 with the provenance file.
4. Whether `bd bdp serve` remains after W2 (default: the strict alias and
   the only minting path).
5. Whether hazard-R publication should run on a temporary branch to keep
   the working set untouched on failure (the scoped commit + recorded-HEAD
   reset is the v0 answer).

## Part E — Proposed ruling amendments this spec assumes (pending)

A1 store-owned witness asserted in every transaction is the v0 lease
(ruling 9); A2 BDP rows inside `httpapi`, `bd bdp serve` the strict command
and the only minting path, `bd serve` never refusing on account of the
graph, the serving leg is the UOW leg (rulings 7b/12); A3 `bd bdp …`
namespace with an authoritative CommandPath-keyed policy; A4 values, laws,
roles in public `graphops`, `BeadGraph*` accessors, no `backend/` aliases;
A5 ruling 11's mechanism = installation-keyed witness file with two-phase
transitions + hash-chained ledger events with a sequence counter + manifest
snapshots with a recovery predicate + provider `LedgerDurability`; A6
`bdp.scope_url` tracked (refused by `config set` once minted), per-workspace
keys in `config.local.yaml` via `bd bdp client`, `BDP_SCOPE_URL` first, no
env token, `bdp.client` blocked from env; A7 fences composed by hazard
(shared-database lease with a mutating check; remote fetch → ancestor →
scoped commit → typed push; push-on-commit for P3; no-remote embedded not
promotable in place); A8 two options for constraint #1. Plus two decisions
the plan does not yet hold: the out-of-role DML enforcement boundary, and
the replication/merge ADR as a P1 gate. Full text: architecture §2b.
