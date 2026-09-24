# Disposable graph CLI preview

This is the first installed vertical slice for [delivery plan review #18](https://github.com/donnabox/beads/pull/18). It is not complete C0 qualification or complete Memory support. The original three-working-day attempt began September 23, 2026 at 07:01 PDT. September 23's command demonstration commitment was missed during a capacity interruption; installed command evidence began September 24.

## Run it

Build through the canonical Makefile, installing into an isolated directory. The qualified integration baseline is intentionally pinned, so use `make install-force INSTALL_DIR=/absolute/preview/bin` if the normal install freshness check rejects this branch. This does not waive build or signing checks.

From a fresh disposable working directory:

```sh
bd init --graph-mode link --scope-url https://example.invalid/disposable-demo
bd remember 'First durable graph record.' --id beads/plan --title Plan
bd show beads/plan --json
# Every invocation is a new process; repeat after exiting the shell if desired.
bd show beads/plan --json
bd status --graph --json
```

Use an operator-selected Scope URL; the placeholder above is for a disposable demonstration. The existing Scope normalization rule adds a trailing slash. Initialization reports the normalized identity. No HTTP service needs to run at that URL for local use.

For an ordinary, externally managed Dolt SQL server, add `--server --external --server-host 127.0.0.1 --server-port PORT` to init. The server must be a disposable instance owned by the test operator. Normal initialization creates the requested database, installs the standard Beads schema and preview graph tables, and publishes readiness. No SQL seed script or separate bootstrap command is used.

`remember` in this preview requires an explicit canonical `--id` and a nonempty `--title`. Its one argument is always the body, including an empty body. Reusing an allocated ID refuses. `show` returns the complete current record and verifies that its retained snapshot exists and matches. Other graph-workspace commands refuse before opening the legacy Issue store. Existing dependency-mode workspaces retain their original routing.

## Current boundaries

- Specialized Issue and Dependency integration is planned, not implemented. This preview creates non-Issue Memory records only. No Issue record is copied into a second editable graph representation.
- The catalog contains one experimental Scope-local Memory descriptor, as permitted for the minimal checkpoint in plan §5. Complete built-in catalog installation belongs to W1/M1; it is not an additional prerequisite invented for the one-Memory C0 transcript. The descriptor and schema are provisional; these disposable workspaces carry no migration, movement, backup or recovery compatibility promise.
- Each create atomically changes the shared coordination cell, allocates the canonical path, writes current payload and retains complete accepted preview state. A typed server serialization rejection is a conflict; transport/commit uncertainty is not a confirmed rollback and is never replayed automatically.
- Retained version tokens are preview tokens scoped to the canonical record. This is recording groundwork, not public BDP/Jim History interoperability. Historical reads, comparison, restoration, removals and change feeds remain unavailable.
- Links, Issue workflows in graph mode, aliases, updates, migration and BDP serving are not yet available. Open deletion, repinning, nominal-Type and metadata decisions are not settled by this implementation.
- An interrupted init leaves a marked, incomplete disposable workspace. It refuses ordinary writes or reinitialization. The failed workspace/database must be inspected and explicitly discarded; there is no automatic adoption or repair of an existing database. This fence is not delivery of W1's recoverable-bootstrap exit.
- Local metadata and database identity must agree. Copying or moving the workspace does not transfer authority. Arbitrary SQL writers and hostile metadata modifications are outside this preview's controlled-writer contract.

## Evidence

Run `python3 scripts/graph-c0-smoke.py --bd /absolute/installed/bd --output-dir /new/receipt-directory`. For shared server add `--server-port PORT --server-root /absolute/disposable/server/root`. The harness starts actual CLI processes, retains raw stdout/stderr and exit codes, and never seeds schema. A smoke pass explicitly reports `c0_qualified: false`: accepted atomicity and failed/cancelled/uncertain-writer observations are separate C0 obligations. Full catalog, recoverable bootstrap, rich History and four-way OS/backend qualification remain the plan's broader W1/M1 obligations. Storage tests separately cover rollback after every write stage, cancellation before commit, forced ordinary-server overlap, and a real lost COMMIT response. The latter returns an unknown outcome to the client while a fresh connection recovers the complete committed record; it is not proof of every disconnect timing or physical worker termination.

The binary currently inherits the pinned donnabox/mysql client replacement in go.mod. Using a released Dolt server is not a claim of a stock-client build. No maintained engine fork is introduced for this slice.
