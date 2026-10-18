# Proposed generic BDP additions for an efficient Issues client

Working proposal, 2026-10-08. This is design input, not an approved protocol amendment. Baseline: BDP `182f1fcf8a01d896976bff3c9e3fb87c596c6ca6`. The audited Issue behavior and immutable source/test evidence are in [issue-cli-audit.md](audit-20261008/issue-cli-audit.md). No Issue verbs, provider-loaded packs, arbitrary SQL, or server-side user code are proposed.

## Recommendation

Add **one bounded query algebra**, usable for coherent reads, transaction assertions, and ordered selection for mutation. Preserve existing BDP batch atomicity. Add guarded partial-success groups only as a separate optional capability if compatibility requires them. Keep advanced traversal and text indexing capability-gated. Packs declare the generic capabilities and scale envelope they require; “conforming BDP” alone does not promise efficient execution of arbitrary workloads.

This addresses three distinct costs: discovering a small relevant result, validating a decision against concurrent changes, and selecting the row to mutate without a client round trip. A local replica remains valuable for offline use, fuzzy comparisons, and repeated graph analysis; it should not be mandatory just to return ten ready tasks from a remote service.

The current spec deliberately excludes joins, traversal, projection, aggregation and string functions, fixes collection ordering, and restricts batch bindings to creation identities. These proposals therefore need a versioned extension with discoverable schemas, not undocumented query parameters added to v0. [Current selection](https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L781-L842), [collections](https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L5367-L5455), [bindings](https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L970-L1043).

## The small shared algebra

Use a structured, versioned JSON AST rather than expanding JSONPath into a programming language. Existing Selectors remain supported and can be translated into its scalar subset. A request names one Scope, one root collection, a predicate, an ordered list of sort keys, selected fields, and a result bound. Every pagination continuation preserves the same query and read handle.

| Construct | Proposed contract |
|---|---|
| Scalar predicate | Existing equality, ordered comparison, Boolean composition and existence, with literal parameters and singular property paths. Define missing, null and type mismatch explicitly. |
| Text contains | Exact Unicode scalar-sequence substring matching. Optional case transform must have one protocol-pinned Unicode algorithm/version, independent of database locale; no authority-selected collation. No implicit accent stripping, stemming, fuzzy matching or normalization. A separate explicitly named normalized mode may be added later. |
| Membership | `anyOf`/`allOf` over an array of scalar values, exact literal equality; object key existence for map representations. Duplicate operands do not alter truth; any-of-empty is false and all-of-empty true. No arbitrary nested JSON subquery. |
| Relationship predicate | `existsLink` with direction, Link type/conformance filter, Link-property predicate and optional opposite-endpoint Bead predicate. `not` expresses anti-join; bounded nesting composes two-hop conditions. References compare by URI regardless of provenance pin, following BDP endpoint-filter semantics. |
| Ordering | Singular scalar property keys, explicit ascending/descending and null/missing placement, followed by mandatory canonical-ID tie-break. Cross-type sorting follows a specified total order or is rejected; never provider collation defaults. |
| Projection | Requested property paths plus mandatory identity/type/revision. Missing projected properties stay absent. Responses explicitly identify a projection, so clients cannot treat them as complete Resource records or replica postimages. |
| Existence/count | `exists` and exact `count` of selected roots; count before pagination. A projected `countLinks` uses the same bounded anchored relationship predicate for per-root counts. Relationship traversal must not multiply root rows. Estimated counts require a different, optional response shape. |

No arbitrary cross-product joins, path-to-path arithmetic, grouping language, relevance engine, or general functions are required. Relationship predicates are anchored to the candidate Resource, not unrestricted relational algebra. This is reusable for inventory with unavailable components, documents awaiting approvals, assets lacking valid certificates, and customer accounts with overdue invoices.

For today’s hybrid ready ordering, the client can issue two disjoint queries under one handle: recent candidates ordered by priority/creation, then older candidates ordered by creation. It fills the requested limit from the first bucket before reading the second. That avoids adding arbitrary conditional order expressions merely for an Issue policy.

**Optional traversal:** `reachable` starts from explicit seeds, follows specified Link types/direction, returns distinct identities, and requires limits on depth, examined vertices/edges and result count. Separate traversal predicates from output predicates so filtering out a closed intermediate node does not accidentally prevent reaching its descendants. Specify inclusion of seeds, cycles and deduplication. No path enumeration initially. If the requested answer needs exploring beyond a bound, fail with an explicit incomplete-query problem; do not answer “unreachable” after silently stopping. Positive existence may stop once proved; a negative answer needs exhaustion within the budget. This capability supports dependency descendants, folder trees and bill-of-material expansion. Full Issues compatibility may require it even though minimal BDP query providers can omit it.

## Coherent reads without an interactive transaction

Extend the existing snapshot model with a **queryable read handle**: canonical Scope, epoch, authorization view, immutable position and fixed expiry. Queries against the handle see that same state across both collections. Creating the handle may accompany the first query; an explicit bundle of independent queries can share it in one request. This is a read-only retained view, not `begin/commit` and not a lock held by the client.

Existing full snapshots remain the replication bootstrap. Projected query results do not replace them. A minimum checkpoint still means “at least this fresh,” not “read exactly this old state.” Expired handles, epoch/view changes or erased required history fail explicitly. A provider can retain immutable rows/manifests or use another finite snapshot representation; PostgreSQL need not hold a connection and transaction open for every idle client. [Current snapshot contract](https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L2035-L2065).

## Assertions that protect decisions, including phantoms

Introduce `assertQuery` inside a Transactional batch. It evaluates the shared algebra against the transaction’s current staged state and supports either:

- an expected cardinality, including zero; or
- an opaque **observation token** returned by an earlier query, meaning the same ordered selected identities and revisions still result.

The token binds the normalized query, semantic parameters, Scope/epoch/view, selection bound, ordering, result identities/revisions and expiry. Projection does not weaken revision comparison. An observation of a bounded top-K guards that exact ordered top-K, not the entire unseen population. An unbounded observed set must have been read completely before a usable exact-set token is issued. Tokens carry no authority to read or mutate.

Cardinality assertions are best for semantic checks such as “no open child exists.” Observed-set assertions are best when a client calculated heterogeneous edits or a digest from retrieved records. Checking only count cannot validate an observed membership set. Checking only selected row revisions misses newly appearing matches; the authority must re-evaluate membership/order or prove an equivalent result, including phantoms.

All successful checks must hold at the batch’s logical serialization point, with each check evaluated after preceding staged operations. Concurrent writes may cause serialization retry; they must never turn a stale check into a successful commit. Do **not** re-evaluate every guard against final post-state: a valid “currently unassigned” precondition may intentionally be falsified by the later claim in that same batch. PostgreSQL serializable transactions, predicate locks, constraint/generation rows or a serialized Scope writer are implementation choices. A token compresses request bytes; it does not make validation free.

This replaces mutation-shaped no-op guard tricks, grants read authorization to assertions rather than requiring write permission on every inspected Resource, and enables a client to guard multiple query results in one batch. It does not enforce policy on a client that deliberately omits the guard. Permanent global invariants remain a separate generic constraint/authorization question.

## Select and mutate at the same serialization point

Add bounded `queryApply`: select roots with predicate/order/take, bind each selected root, then apply a fixed generic operation template per root. The selected identity list is frozen when the operation is reached; later template executions do not repeatedly select from a shrinking pool. Entire operation and surrounding batch remain atomic. A matched-set cardinality constraint and a post-take selected-count constraint are distinct fields; `take:1` chooses one from many, whereas `{max:1}` on the full match set would refuse many.

For the minimal version, templates may bind selected identity/revision and named creations. Scalar assignment values may be literals, selected singular-path values, or a narrowly defined `coalesce` of those values for initialization (skip missing/null, with no other implicit coercion). The JSON carrier must distinguish value expressions from literal JSON, so strings such as `@foo` never become accidental code. No arithmetic, network access, arbitrary result paths, recursion or unbounded loops. This is enough to retain an existing started timestamp or initialize it from a client-supplied timestamp. More complex string append/merge can keep using guarded read/modify/write.

A take-one claim can patch the chosen Resource and create its lease/event Links in the same template; no match produces an explicit zero-selection success with no effects. Return chosen identities, final projected records and changed counts in the receipt. Bindings remain transaction-local and kind-checked. Preserve idempotency across ambiguous delivery: retry returns the same winner/result, not a newly selected task.

Do not implement this by PostgreSQL `SKIP LOCKED` unless the requested semantics expressly allow skipping locked earlier candidates. Strict first-in-order selection and opportunistic queue throughput are different promises. Start with strict semantics and measure contention before adding an optional relaxed mode.

Generic uses include reserving the earliest available appointment, allocating a warehouse item, leasing an unprocessed document, and assigning the oldest eligible job.

## Partial success: optional guarded groups, never weakened batch

Existing multi-ID close can refuse some targets yet commit other closes and claim-next in one transaction. To preserve that, optionally add **flat guarded groups**. Each group has a query precondition and ordinary operations. False precondition records `skipped` and has no effects; true runs the operations. A later group may depend on the bounded summary “any earlier named group changed state” before selecting next work. No nesting or general exception handling.

Only a false declared business guard is skippable. Malformed requests, authorization failures, schema violations, limits and execution errors fail the whole batch. All applied groups commit together; nothing commits until the surrounding transaction commits. Guard failures may return a caller-specified safe reason code, never hidden Resource details. This preserves the meaning of batch and avoids turning every error into partial success. [Current all-or-nothing contract](https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L4419-L4430).

If this capability is omitted, a client can preflight and retry one all-or-nothing batch or use separately committed per-target batches. Neither should claim exact parity with all partial-success transactional behaviors. Prototype the required command cases before standardizing guarded groups; this is the largest control-flow addition and should not block the query/assertion work.

## Authorization, indexes and bounded failure

Reads, joins and counts operate entirely within one Authorization View. A Link to a hidden in-Scope endpoint does not join. No count or assertion may reveal hidden rows. Therefore “no visible blocker” is not “no blocker exists.” For commands requiring authoritative graph completeness, require a principal granted the necessary complete Scope/subgraph view, or a separately defined conservative completeness certificate. For the first implementation, a whole-Scope read grant is the simplest auditable requirement. Do not condition an incomplete-visibility refusal on whether a hidden blocker happens to exist; that would create an oracle. Cross-Scope joins and atomicity remain excluded. [Current view semantics](https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L904-L953).

PostgreSQL can compile this bounded algebra to parameterized predicates, `EXISTS`/`NOT EXISTS`, indexed endpoint lookups, ordered scans and optional bounded recursive CTEs. Providers need indexes on frequently queried scalar paths, type/status/order combinations, both Link endpoints, and declared text modes. JSON storage alone is not an efficiency guarantee. Exact substring search may still scan; trigram indexes help many distributions but do not bound all workloads. Broad anti-joins and reachability may explore most of the graph.

Discovery advertises supported algebra/version/text semantics and limits: AST nodes/depth, join depth, selected/projected bytes, examined roots/edges, traversal depth, take, template expansion, operation count, duration and read/token retention. Unsupported capability fails before execution. Exceeded budget fails explicitly, atomically for mutations, with no silently truncated predicate, partial count or falsely complete result. Read pagination is allowed; mutation paging across transactions never restores whole-job atomicity. A useful failure distinguishes unsupported query, expired observation, failed assertion, incomplete visibility and exceeded execution budget, with safe retry guidance.

An optional generic index-administration interface can follow later. First agree required query shapes and run provider-specific indexed benchmarks; do not standardize a cost promise based on a storage implementation detail.

## Three worked command journeys and acceptance budgets

**1. `ready --limit 10` with hybrid ordering.** At a captured cutoff, a pre-read `updateWhere` wakes dated expired deferrals (existing CLI behavior). Obtain query handle and recent-bucket page using scalar predicates plus `not existsLink` for open blockers/children; fetch older bucket only if needed. Project display fields and required counts. Target: 2–3 application requests after discovery, at most O(K) selected Issue payload plus explicitly requested counts, no whole-Scope download. One additional request is acceptable if the provider separates handle creation. These are acceptance targets under supported indexed shapes, not universal latency guarantees; defer sweep or graph budget failure must remain visible.

**2. `ready --claim`.** One batch performs expired-defer wake if required, then `queryApply(take:1)` over the full ready predicate/order. Template sets holder/status, preserves-or-initializes started_at, and creates/updates generic lease state. For hybrid order, two disjoint query applications need a bounded “only if first selected zero” control; initially compile this through optional guarded groups or use a separate recent-bucket probe plus assertion/retry. Target for simple priority order: one mutation request and receipt, one chosen row, no N-candidate client reads. With 32 contending clients, prove no double winner and measure retries/throughput; do not promise wait-free strict ordering. Lease no-history semantics are still a separate retention decision.

**3. `close A B --claim-next`.** Two guarded groups test each target’s children/blocker policy, patch eligible targets with expected revision, and record skipped outcomes for others. A final group runs queryApply only if a close changed state. All closes and the selected next claim commit once; repeated idempotency key returns the same outcomes. Target: one batch request when IDs/state-independent patch data are known, O(named targets + returned next task) payload. Without guarded-groups support, explicitly choose preflight/retry or weaker multi-transaction behavior. Concurrent child attachment, blocker reopen and candidate insertion are mandatory phantom tests.

## Proposed sequencing

1. Query scalar/text/membership semantics, ordering/projection/counts, anchored relationship predicates and read handles; differential tests across two providers.
2. Transactional assertQuery and queryApply; concurrency tests including empty-result phantoms, queue contenders and authorization/view changes.
3. Optional bounded reachability and guarded groups, driven by concrete compatibility tests and measured workloads.

Ship an Issues pack manifest declaring this generic profile. Reuse the same conformance fixtures for a non-Issue domain. That is the strongest practical proof that the provider implements BDP semantics rather than hiding an Issue runtime behind a new endpoint.
