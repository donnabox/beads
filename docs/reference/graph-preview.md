# Memory graph preview: create and read

This integration checkpoint supports creating and reading non-Issue Memory
Beads in a **fresh, explicitly selected graph workspace**. It is an experimental
subset of the graph delivery plan. Existing ordinary Issue workspaces continue
using their existing commands and storage.

Start in a new directory with no `.beads` directory:

```sh
git init
bd init --graph-mode link --scope-url https://example.org/team/ \
  --skip-hooks --skip-agents --non-interactive
bd remember 'The release uses the integration branch.' \
  --id beads/plan --title 'Release plan' --json
bd show beads/plan --json
# A separate invocation reopens the same stored Memory.
bd show https://example.org/team/beads/plan --json
bd status --graph --json
```

The Scope URL establishes local identity; this command does not publish a web
server at that address. `remember` also accepts explicit `--body-file PATH` or
`--stdin` input. These input sources are mutually exclusive. Empty text is a
present Memory body.

For an ordinary external Dolt SQL server, add these options to `bd init`:

```sh
--server --external --server-host 127.0.0.1 --server-port 3306 --server-user root
```

Use the normal connection configuration appropriate to your server. Embedded
and server modes use the existing driver and standard schema initialization;
there is no manual schema seeding step. Graph initialization records and checks
workspace identity and refuses existing workspaces, mismatched bindings and
incomplete initialization. This preview does not adopt an existing Issue
database. Different-database provisioning on Dolt 2.1.8 must be serialized.

## Current limits

This checkpoint admits `init`, `remember`, current `show`, and `status --graph`.
Other commands in a graph workspace refuse before ordinary storage operations.
Memory updates and deletion, Issues in graph workspaces, Link mutations,
collections, exact-version reads and BDP HTTP serving arrive in subsequent
integration increments. CLI JSON here is **not** the promised BDP script API.

Memory creation atomically writes its current payload and private retained
snapshot. Those snapshots are not a claim of complete native History, Dolt
HEAD/sync durability, public BDP Write, or recovery support. The current read
acquisition budget can refuse an oversized workspace; this checkpoint does not
promise that every successful create preserves that aggregate read budget.
`bd status --graph` reports the implemented capabilities explicitly.

## Required regression gate

The PR workflow's `Graph C0 / real engines and installed CLI` job is required by
the existing CI gate. It verifies source and CLI artifact provenance, executes
all graphstore tests with both engines available, and runs the graph config and
CLI admission tests without skips. It then runs installed initialization,
Memory creation, fresh-process reads and status against both engines.

The runner checks discovered source tests against compiled tests and requires
every started test and subtest to pass. Missing services, omitted tests, skips,
unclean process shutdown or mismatched artifacts fail the job. Evidence is
uploaded on success and failure. Normal Issue regression suites remain required.
