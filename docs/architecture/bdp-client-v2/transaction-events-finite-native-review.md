# Finite Transactional Event native review

Reviewed head: `8569b85d785f96325ca948580106f454c48ed034` against `076201849f98eeab71bad5c8404fb693a36c6474`.
Author checkout: `/Users/dbox/repos/wt/vickie/vickie-transaction-events-finite` (frozen).
Independent detached review checkout: `/Users/dbox/repos/wt/vickie/vickie-events-native-review` (clean at reviewed head).

## Findings

No actionable defect found in the authorized finite Event scope. This is a bounded implementation review, not public profile qualification. In particular, the known metadata-only incident-Link projection case remains an explicit unresolved gap; SSE is not implemented or claimed.

## Inspected guarantees

- **Projection and identity:** Scope, Bead and Link views use the same persisted group facts with source-specific IDs. The source classifier and cursor source tag prevent an ID from another source being accepted. Ordinals preserve original group positions; self-Link source and target incidence remain distinct. Endpoint witnesses come from final Link upserts and lifecycle/incidence facts, with consistency checks. Deleted Resources retain Event-source identity.
- **Cursor context:** Separate epoch, authorization-view and source tags classify expired epoch and foreign view independently. A continuation must also correspond to an actual projected member, not merely an in-range global ordinal. These IDs are context-bearing cursors, not cryptographic proof of issuance.
- **Stored group and page bounds:** `transaction-store.ts` checks SQL `length(CAST(body AS BLOB))` before fetching/parsing the group body. The bound is derived from the actual immutable runtime tuple and is 253,493,248 bytes. Traversal selects one position at a time beneath one captured head, retains a bounded page plus one group, and checks actual serialized page bytes including the actual continuation URL. The first-Event feasibility bound is 1,778,688 bytes, below the 4 MiB page budget. This does not prove SQLite internal allocation bounds, a general heap bound, or a bounded total history scan.
- **Page and HTTP coherence:** Pages split within groups without splitting Events. Continuations identify the last returned Event only when more exist. The minimum-position waiter is outside SQLite ownership and keeps one original deadline. Ordinary cursor and minimum checks precede conditional handling. GET/HEAD share the Read response helper, carry current consistency context and private/no-store caching, and omit ETag/Last-Modified. The absence of validators is honest; wildcard conditionals must not hide minimum/context failures. SSE-only Accept is refused rather than advertised.
- **Lifecycle and independence:** Service work participates in close/drain accounting. Cancellation is effective between synchronous attempts, not an interrupt of a running synchronous history scan. Receipt compaction/forgetting does not erase persisted groups. Restart preserves exact Event IDs and payloads. This does not qualify administrative Event erasure, live streaming or immediate disconnect cancellation.

Relevant implementation reviewed: `transaction-events.ts`, its tests, `transaction-store.ts` Event accessor (approximately lines 1237–1351), `transaction-service.ts`, `transaction-http.ts`, service tests, package-output allowlist and bounded progress documentation. Normative Event/projection, pagination, conditional and minimum-position clauses were checked against the frozen spec, including the sections around lines 5360, 5470, 5530, 5750–5820, 6175 and 6536–6565.

## Independent execution

Pinned Node 24.16.0 and pnpm 11.20.0; disposable stores only. Frozen install, build and typecheck each exited 0. Focused tests: **89 passed, 6 files, 0 failures** (3.40 seconds), covering `transaction-events`, `transaction-service`, `transaction-http`, `transaction-consistency`, `transaction-recovery` and `transaction-runtime`. No full-suite or smoke rerun is claimed here; the author's separate report records those runs.

Two additional emitted-runtime probes passed:

1. **Cold actual HTTP and restart:** Two transactions create/update/delete three Beads, an ordinary Link and a self-Link. Literal independent expected Event sequences produce Scope 18 Events/9 pages, deleted Bead 10 Events/5 pages, and deleted Link 3 Events/2 pages. The probe checks every page's source, count, byte bound, continuation cursor and headers; deleted current Resource 404 versus available Event source; empty continuation after the last Event; HEAD; wrong epoch/view/source before wildcard conditional handling; and exact equality of every page after close/reopen. Output records `exactRestartPages:true` and `validatorsOmitted:true`.
2. **External-source/local-target Link:** Update and delete a legal external-source Link. Its local target receives exactly created/linked/updated/unlinked while an unrelated Bead's page stays exactly unchanged. Scope and Link views remain readable.

A rejected hypothesis is retained for audit: a Link with *both* endpoints outside the Scope might have lacked a local incidence witness during update/delete. The real evaluator refuses its creation with 422 (`a Link must have at least one in-Scope endpoint`, `resource-evaluator.ts` around lines 745–750), so that proposed counterexample is unreachable in this owner. The initial failed probe is not counted as a product failure or a passing test. The corrected legal-Link probe passed.

Evidence directory: `/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-events-finite-native/`. It contains install/build/typecheck logs and exit files; `tests.log`, `tests.json`, `tests.exit`; cold and external-Link probe scripts/logs/exit files; and the rejected-hypothesis log. Both source checkouts remain unchanged. No outward writes, shared database use, source edits or public conformance claims were made.

## Remaining boundaries

Metadata-only incident projection is still deliberately unresolved. Finite delivery is not SSE, retained Event paging is not full History, context tags are not signatures, and proof of a maximum serialized group is not proof of comfortable heap usage for all graph shapes. Synchronous scanning can traverse many groups before filling a source-filtered page. Those limits must remain explicit in downstream claims; none is silently closed by these 89 tests or the additional probes.
