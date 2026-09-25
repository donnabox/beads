# Disposable graph CLI preview

This branch extends the first installed vertical slice for [delivery plan review #18](https://github.com/donnabox/beads/pull/18). It adds experimental Issue create/read and blocking-Dependency workflow adapters described in [the Issue preview notes](GRAPH_ISSUE_ADAPTER_PREVIEW.md) and [workflow notes](GRAPH_DEPENDENCY_WORKFLOW_PREVIEW.md). It is not complete Memory or Issue workflow support. The original three-working-day attempt began September 23, 2026 at 07:01 PDT. September 23's command demonstration commitment was missed during a capacity interruption; installed command evidence began September 24.

## Run it

Build through the canonical Makefile, installing into an isolated directory. The qualified integration baseline is intentionally pinned, so use `make install-force INSTALL_DIR=/absolute/preview/bin` if the normal install freshness check rejects this branch. This does not waive build or signing checks.

From a fresh disposable working directory:

```sh
bd init --graph-mode link --scope-url https://example.invalid/disposable-demo
bd remember 'First durable graph record.' --id beads/plan --title Plan
bd show beads/plan --json
# Every invocation is a new process; repeat after exiting the shell if desired.
bd show beads/plan --json
bd create "Fix deployment" --id beads/work --description "A real Issue" --type bug --labels demo
bd show beads/work --json
bd status --graph --json
```

Use an operator-selected Scope URL; the placeholder above is for a disposable demonstration. The existing Scope normalization rule adds a trailing slash. Initialization reports the normalized identity. No HTTP service needs to run at that URL for local use.

For an ordinary, externally managed Dolt SQL server, add `--server --external --server-host 127.0.0.1 --server-port PORT` to init. The server must be a disposable instance owned by the test operator. Normal initialization creates the requested database, installs the standard Beads schema and preview graph tables, and publishes readiness. No SQL seed script or separate bootstrap command is used.

`remember` in this preview requires an explicit canonical `--id` and a nonempty `--title`. Its one argument is always the body, including an empty body. Reusing an allocated ID refuses. `show` returns the complete current record and verifies that its retained snapshot exists and matches. The experimental `create` route accepts a canonical `--id` and creates a durable Issue through the existing Issue writer. The bounded workflow adds `dep add`, `link`, single-Issue `close`, and unfiltered `ready`. Other graph-workspace commands refuse before opening the legacy Issue store. Existing dependency-mode workspaces retain their original routing.

The preview uses the exact persisted storage route. Select the workspace with the working directory, `--directory`, or `BEADS_DIR`. Its `.beads/.env` supplies policy and static credentials with shell values taking precedence. Unsupported backend values, database selectors, redirects and conflicting endpoint/data-directory assertions refuse before opening storage; the preview does not silently ignore them or start another server. Server passwords may come from `BEADS_DOLT_PASSWORD` or the existing endpoint-keyed credentials file. Credential commands are explicitly unsupported. An explicit init `--server-user` wins over environment defaults; later opens honor the static environment user. Quiet init suppresses human output while retaining explicit JSON output.

## Current boundaries

- Specialized Issue create/read, local blocking Dependencies, close and ready are experimental on this branch. General Issue updates and the remaining workflows are not implemented. Issue payloads remain authoritative in the existing normalized Issue tables; no second editable graph representation is created.
- The C0 catalog contained one experimental Scope-local Memory descriptor, as permitted for the minimal checkpoint in plan §5. Complete built-in catalog installation belongs to W1/M1; it is not an additional prerequisite invented for the one-Memory C0 transcript. This branch also installs experimental Issue and blocking-Dependency descriptors. All descriptors and the schema are provisional; these disposable workspaces carry no migration, movement, backup or recovery compatibility promise.
- Each create atomically changes the shared coordination cell, allocates the canonical path, writes current payload and retains complete accepted preview state. A typed server serialization rejection is a conflict; transport/commit uncertainty is not a confirmed rollback and is never replayed automatically.
- Retained version tokens are preview tokens scoped to the canonical record. This is recording groundwork, not public BDP/Jim History interoperability. Historical reads, comparison, restoration, removals and change feeds remain unavailable.
- Informational Links, Issue workflows beyond create/read/blocking Dependency/close/ready in graph mode, aliases, general updates, migration and BDP serving are not yet available. Open deletion, repinning, nominal-Type and metadata decisions are not settled by this implementation.
- An interrupted init leaves a marked, incomplete disposable workspace. It refuses ordinary writes or reinitialization. The failed workspace/database must be inspected and explicitly discarded; there is no automatic adoption or repair of an existing database. This fence is not delivery of W1's recoverable-bootstrap exit.
- Local metadata and database identity must agree. Copying or moving the workspace does not transfer authority. Arbitrary SQL writers and hostile metadata modifications are outside this preview's controlled-writer contract.

## Evidence

Run `python3 scripts/graph-c0-smoke.py --bd /absolute/installed/bd --output-dir /new/receipt-directory`. Add `--dependency-workflow` to also exercise the installed two-Issue workflow. For shared server add `--server-port PORT --server-root /absolute/disposable/server/root`. The harness starts actual CLI processes, retains raw stdout/stderr and exit codes, and never seeds schema. A smoke pass explicitly reports `c0_qualified: false`: accepted atomicity and failed/cancelled/uncertain-writer observations are separate C0 obligations. Full catalog, recoverable bootstrap, rich History and four-way OS/backend qualification remain the plan's broader W1/M1 obligations. Storage tests separately cover rollback after every write stage, cancellation before commit, forced ordinary-server overlap, and a real lost COMMIT response. The latter returns an unknown outcome to the client while a fresh connection recovers the complete committed record; it is not proof of every disconnect timing or physical worker termination.

The binary currently inherits the pinned donnabox/mysql client replacement in go.mod. Using a released Dolt server is not a claim of a stock-client build. No maintained engine fork is introduced for this slice.
