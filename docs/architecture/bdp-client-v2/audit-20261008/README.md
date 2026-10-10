# Existing CLI over BDP: feasibility, efficiency and journal audit

2026-10-08 · Vickie · Design evidence for [Beads #7403](https://github.com/gastownhall/beads/issues/7403).

**Requirement from Donna:** a conforming provider with the required generic BDP capabilities must support the Issues pack without Issue-specific provider code. Where the existing protocol cannot do that efficiently or faithfully, examine the protocol and client architecture. No private Issue endpoint is an acceptable escape hatch.

**Finding:** generic BDP already supports most bounded mutations, especially when compiled into Transactional batches. The current protocol does not yet establish an efficient, faithful implementation of the entire CLI. The largest gaps are query expressiveness/ordering, graph-condition concurrency, several legacy semantic mismatches, and engine-specific administration. This audit identifies those gaps without prescribing Issue-specific provider logic or changing the specification.

## Evidence and navigation

| Artifact | Contents |
|---|---|
| [Command crosswalk](command-crosswalk.csv) | All 274 current-main registry paths plus five Preview additions; architectural disposition, source, direct flags and report pointer for every path |
| [Issues and workflows](issue-cli-audit.md) | CRUD, ready/claim/close, dependencies, labels/comments, formulas/molecules, gates, leases, deletion, identity, duplicates, query language and long-tail commands; source/test contracts, generic request plans and costs |
| [Other CLI families](other-cli-audit.md) | Memory, generic graph, History, import/export/backup, federation, configuration, workspaces, integrations and provider administration |
| [Supporting families](supporting-cli-audit.md) | Human queue, provenance, KV, interaction audit, ping/mail/rules/upgrade, output and argument obligations |
| [Separate journaling report](event-log-audit.md) | Log suitability, operation/Event/receipt field matrices, exact fixture example, loss/recovery cases, existing bd batch and legacy journal comparison |
| [Verification](verification.json) | Citation bounds and remote Git blob verification; [artifact hashes](artifact-manifest.json) |
| [Inventory summary](inventory-summary.json) | Exact declaration counts and scope limits; raw AST metadata is in cli-ast-main.json and cli-ast-preview.json |

Pinned source baselines: `gastownhall/beads@5126de8c1a02d01db2481e4fd5e78dd41010e25f`, `donnabox/beads@6c33ea17af3bffbdea46a5ac04715c6dccd7c9bf` (Preview 2), and `gastownhall/bdp@182f1fcf8a01d896976bff3c9e3fb87c596c6ca6`. Full baseline metadata: [baselines.json](baselines.json). Runtime source remains unchanged.

The inventory counted 3,453 top-level test declarations in 733 cmd/bd files on main and 3,621 in 799 files on Preview; 1,798 and 2,077 flag-declaration calls respectively. These are **static inventory counts**, not executed tests, unique effective flags, behavior coverage percentages or assertions that every flag combination was manually audited. Registry paths include some parent commands and omit alias expansion; raw Cobra declarations and registration edges provide the second inventory axis. Five factory/constant/alias paths use registry evidence in the CSV. Selected deeper storage/role tests were also inspected and cited. Exact legacy help/output/error parity remains future execution qualification.

## What maps well today

- Scalar create/update/reopen, exact field filters and independent metadata-key patches.
- Create a Bead and its Links in one batch using local identity labels; formula expansion is client work followed by a generic graph batch.
- Claim a known candidate with expectedRevision or an exactly-one conditional set mutation. One winner is achievable without an Issue endpoint.
- Separate comment/provenance records, small merge-slot state machines, bounded graph transforms, logical import chunks and coherent export snapshots.
- Packs own validation, display, argument processing, lifecycle policy and external integration. No BDP protocol verb is needed for every command name.

The full command contract needs **Transactional**, with **History** when retained-version features are promised. Read-only or Read+Update providers cannot claim support for the whole Issues pack. Capability requirements and refusal behavior must be explicit; this is a generic capability contract, not permission for domain plugins.

## Protocol/client pressure points

| Priority | Existing behavior | Current generic route and problem | Design question |
|---|---|---|---|
| P0 | `ready`, priority/date ordering, search, counts, parent/dependency filters | Coarse scalar filter, page complete records, collect Links, evaluate/sort locally. No caller-selected order, text functions, joins/traversal, projection or aggregation. Worst-case candidate/graph download for a few results. | Which generic query facilities are required for cold clients, and which workloads justify a synchronized local index? |
| P0 | Close only when children/blockers allow it; cycle-safe edits; dry-run-derived mutations | Read graph, calculate write plan, guard relevant revisions and membership, commit batch. A target revision alone misses graph phantoms. A whole-Scope snapshot supplies read consistency but not commit-time protection. | Define compact generic read/predicate assertions and consistent selective reads; measure the bounded existing guard strategy first. |
| P0 | `ready --claim` and close-plus-claim-next | Known candidate CAS gives one owner. It does not select the best eligible candidate at commit. Set cardinality max=1 rejects two matches; it does not choose one. | Is guarded client selection/retry the accepted semantic contract, or is generic ordered conditional selection necessary? |
| P1 | Multi-ID close with per-item refusal, successful closes and claim-next in one transaction | Current BDP batch is all-or-nothing; preflight can select successes, but racing policy changes cause retry. Separate calls change atomicity. | Preserve via bounded guarded preflight/conditions, or explicitly approve changed command semantics. |
| P1 | Large cascade, formula graph, label propagation and import | Batches/set mutation are powerful but finite; induced events and selected resources also consume limits. Chunking can change atomicity. | Advertise/discover usable limits; distinguish atomic jobs from resumable chunks; fail explicitly when the promised unit cannot fit. |
| P1 | `ready` wakes dated deferrals; heartbeat avoids durable history; promotion keeps ID | Ready currently writes, timestamps use a clock policy, and ordinary BDP mutations generate versions. Separate operational Scope loses claim+lease atomicity. | Specify clock, permission and generic retention/volatility behavior; do not hide an Issue lease table in the provider. |
| P1 | Rename IDs, history ordinals, native branch/SQL/backup | Canonical BDP identity is immutable and never reused; History is not Dolt commit history; engine administration has no generic equivalent. | Separate mutable human IDs from canonical identity; explicitly retain/redesign provider administration in the side-by-side story. |

These findings are backed by immutable source/spec/test citations in the family reports. Priorities are this audit's recommendations, not accepted product rulings.

### Transactions help with graph conditions, but need care

A zero-match `deleteWhere` with cardinality `max:0` is a possible generic absence assertion: any match aborts before deletion, no match changes nothing. A same-value update with expectedRevision can guard a known dependency. These are inferences from the specified semantics, and their batch shapes validate against the schema; no live provider has been tested executing them here.

An exact old member count is insufficient: one old Link can be replaced by a new Link without changing count. Enumeration/exclusion plus revision guards can protect bounded graphs. However, guards grow with the read set, hit Selector/operation limits, and no-op updates require write permission on dependencies that the caller may only need to read. Hidden state is outside the caller's authorized view. Source-owned Links can improve the aggregate revision boundary for a chosen representation, but do not guard every target status or transitive dependency.

A shared generation Resource is another generic cooperative-client strategy. It creates a contention hot spot and requires every writer to participate. Decide separately whether pack policy is a cooperating-client promise or an authority-enforced invariant. Neither interpretation authorizes Issue-specific server code.

## Executed checks and measured example

**219 existing BDP tests passed in three files**, using the pinned lockfile and Node 24.16.0 in an isolated source snapshot:

- 61 Read Selector implementation tests: [selector-existing-tests.log](selector-existing-tests.log).
- 158 Transactional wire/catalog tests: [event-log-existing-tests.log](event-log-existing-tests.log).

**20 additional probes passed:** 12 in [client-feasibility-probes.mjs](client-feasibility-probes.mjs) and eight in [event-log-probes.py](event-log-probes.py). Logs: [client probes](client-feasibility-probes.log), [journal probes](event-log-probes.log).

The client probes use the actual Selector implementation and JSON Schema, plus explicitly labeled small CAS/predicate models. They validate four [batch request recipes](batch-recipes.json), reject unsupported batch query/partial-success shapes, demonstrate stale graph reads and membership-count holes, and exercise Selector limit failures. The journal probes reconstruct the pinned fixture by final-state groups and by a seeded partial Event reducer, and test missing intent, predecessor mismatch, projection gaps and erasure-only groups.

A synthetic 50,000-record corpus with roughly 1 KiB of text per record contains one substring match at the end of URI order. The coarse-filter/full-scan fallback consumes **59,300,007 bytes of resource JSON and 500 pages at page size 100** to find that one result. This is measured synthetic payload size plus derived page count, **not measured HTTP traffic, latency, provider throughput or a lower bound on every possible client algorithm**. Exact cost data: [client-query-cost.json](client-query-cost.json). A warm replica can avoid repeating that transfer but introduces bootstrap, disk, freshness, erasure, authorization-view and recovery costs.

**There is no end-to-end BDP-provider qualification result.** The checked-in server admits Read only. Transactional wire fixtures explicitly disclaim behavior evidence, and its conformance catalog has no executable manifest. We did not run the existing Go CLI/database suites; they exercise its current implementation rather than a BDP client, and no new Rust client exists. Passing the checks above establishes parser/schema/fixture behavior and stated model properties, not provider atomicity, crash recovery or workload performance.

## Proposed qualification gates

Keep these as design requirements until an implementation is separately authorized:

1. Run the same Issues pack against two generic Transactional providers, with no Issue routes or server domain module. Capture every HTTP operation to prove the boundary.
2. Port existing semantic scenarios: create-with-links, exactly-one claim among eight clients, successful concurrent notes/metadata preservation, close/child and cycle races, conditional unclaim, duplicate merge, dated defer wake, promotion, partial-close policy and large import failure recovery. Preserve CLI JSON/error/exit contracts or record an approved change per case.
3. Measure cold and warm ready/search/graph workloads at 50k/100k and 1M/2M Beads/Links. Record bytes, requests, latency, local storage, catch-up traffic and conflict retries. Donna must set acceptance budgets; this audit does not invent them.
4. Exercise authorization-view changes, hidden dependencies, no-write access to read dependencies, stale revisions, idempotent retries, expired receipts, negotiated bounds and oversized induced Event groups.
5. Qualify snapshot→changefeed crash/restart at every group boundary, duplicate/gap/reorder, epoch/view rotation, retention expiry and erasure. Keep command intent audit and full provider restore separate.

The next architecture decision should settle the cold-client query/read-consistency contract and the bounded transaction guard model. A local index is useful, but requiring full-Scope replication for an ordinary first `ready` or substring search should be an explicit, measured choice.

## Separate journaling answer

**For an activity timeline, Events are useful. For a durable state journal, use snapshot + complete change groups.** Groups provide final records/tombstones, ordered Events, erasure records and checkpoint continuity, and must be applied atomically with the durable cursor. Event-only consumers lack complete framing, erasure notifications and some authority state. Full provider backup needs additional administrative/historical data even beyond change groups.

**Batch requests and Events share data but are not interchangeable.** The checked-in worked example is three requested operations → six Events → three final postimages. A Link create generates endpoint facts and possibly a source-owned version change. Guards, labels, selectors, no-op attempts and failed commands cannot be recovered from those effects. Consequently, translating each Event back into a batch operation can duplicate derived work and cannot reconstruct intent.

The [separate report](event-log-audit.md) compares the generic BDP batch, existing line-oriented `bd batch`, the legacy Issue batch API and Beads journal rows field by field. The raw legacy journal's postimages are closer to change-group `changes` than to BDP update deltas.

No Preview 2 source, specification, test, branch, playground or database was modified. Type lifecycle remains Trish's work; this audit identifies interface consequences without ruling on it. No V2 implementation or merge occurred.
