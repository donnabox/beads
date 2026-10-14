# Deletion-safety plan v3 — native closure review

## Remaining finding

**1. P3 — One affected structural index is still missing from the complete re-pin inventory (external F1).** `packages/conformance/src/transactional-cumulative.test.ts:930` currently asserts `manifest.scenarios.indexOf(row)` equals8 for `transactional.receipt.pagination`. Inserting deletion-safety after rollback at index2 shifts that unchanged row to index9. V3's “Complete aggregate re-pin inventory” omits this assertion, while asking authors to preserve prior row-local pins. Literal adherence leaves a deterministic failing cumulative test.

**Fix:** explicitly include line930's8→9 scenario-index adaptation. Preserve the pagination row's18 actions, raw bytes, semantic assertions and all existing observations. This is a structural insertion consequence within an already owned file, not new scope, semantics or budget. No implementation/test run is necessary to establish the index arithmetic.

All other reviewed corrections are closed at plan level. Recommend final root disposition after recording this small inventory addition; this report does not approve implementation. The retained native P3 and the new control3 are feasible as written.

## Exact basis

- V3 plan SHA256 `f8bb465200455bdb1d57d07a614a0fc079c149dab48be2cd746acd4d55246ed7`.
- V3 dispositions SHA256 `fe281738ab8667bd042f9d844d8540769b2c6327daa1a83080d46359458a59ae`.
- Completed external original review SHA256 `f40ff6bb58c8fdf0a9d900c2b509be0e2a7d4113da89e56f2cd8b2d341688762`.
- Inspected local source `/Users/dbox/repos/wt/vickie/vickie-noop-conformance`, HEAD8d66855ec4e26d3af0b3d7d22d621e1f37005294/treea1ae75b822e1d7fadd739d818d09a25fb788e380, unchanged executable/runtime/artifact basis from bec97. Root supplies verified merged5995f6b08ebd045e4343059c2da1cffd38b09b80 with this tree; no fetch was performed.

## Dispositions checked

|Item|Closure|
|---|---|
|Native P3|Closed. Controls1/2 explicitly require independent passing completed-result/updated-b/seed-backed deletion/pos2 checks while retaining the named failed-disposition assertion. Request expectations describe the swapped input. Additional old failed-receipt assertion failures are classified; later GETs are not claimed.|
|External F1|Original old-control bound failures are corrected in the plan:19/159→20/180, all three affected controls rerun. Existing route/digest/pass/session/size/count inventories are named. One additional structural index remains omitted as finding1 above.|
|F2|Closed. Control3 actually changes wire state and reaches the next real GET; detailed reasoning below. The old expectation-only case is retained only as static weakening.|
|F3|Closed with root's explicit decision. Source uses one `(source=? OR target=?)` query and common refusal/rollback path (`transaction-store.ts:1148–1152`). V3 accurately justifies the two directions as normative partitions, not independent implementation paths. The optional single-direction arithmetic21→16 is correct; no fabricated measured1× budget.|
|F4|Closed. Exact404 currentpos2 remains mandatory and explicitly unmeasured; Link collection shape and opposite-Link setup are also named risks. No fallback weakening is silently authorized.|
|F5|Closed. Exact five-field source/type/subject/subjectType/ordinal projection is stated.|
|F6|Closed. Operation maximum32 is distinguished from configured receiptMaximumItems/default32 used for finite Event paging, including its byte-bound dependency and no-override evidence requirement. Source `transactionReceiptMaximum()` and finite Event loop support this distinction.|
|F7|Closed. Staged deletions never committed; no restoration Event or compensating write is implied.|
|F8|Closed. Canonical IDs and metadata/title values consistently identify which incident direction survives. Prior artifact bytes remain preserved.|
|F9|Closed. Atomic deletion clause/rollback ranges and self-Link versus distinct-endpoint precedent are accurately scoped.|

## Control3 execution feasibility

Removing delete-a produces `[update b, delete l, delete r]`. Both Links can be deleted while both Beads remain; the unowned fixture needs no source revision side effects. b becomes the prefix postimage and the transaction commits atpos2. The derivative independently authors this exact request and a complete passing POST oracle; it replaces failed-disposition, Problem/absence and pos1/baseline-equality checks rather than merely appending incompatible positives.

`runner.ts:1440–1470` evaluates every assertion on an exchange and aborts the row if any fails. V3 therefore correctly requires ALL POST assertions, including authored request-body checks and mandatory context presence, to pass. The existing next Bead collection GET can then run. Its canonical seed-b full-record binding and literal metadata{} inventory both contradict the genuinely changed served b. Both named failures must be observed on actual200; additional context/revision failures are honestly classified. V3 requires all preceding exchanges to pass and that GET to be the final observed exchange. It does not claim the later Link collection GET ran. No extra request, route, capture, raw-response helper or DSL is needed.

Existing shapes support the intended assertions: updated b has `resource`; deleted entries have `deleted.resourceKind` and `deleted.resource` (id/type/revision). Earlier validated seed captures are available. No unknown output value needs to be captured from a deliberately failing POST. Positive derivative assertions are separate from canonical A/coverage and must be measured; their count is not assumed free.

## Scope, bounds and limits

Canonical21 new/187 total actions and180 full-cohort exchanges remain intact. Twelve partial rows,19 scenario/reopen sessions and20children per cohort reconcile. The seven targeted cohorts (positive + three new + three adapted old controls) have maxima140children/1260exchanges; the169/169/174 old-control figures are explicitly arithmetic rather than executed claims. Per-cohort time/body/loader caps do not increase. Absolute authoring stops and measured-root gate remain, including separate derivative size/assertion ledgers. Whole/public flags remain false and pending product rulings remain untouched.

Read-only plan/review/source reads, Git identity inspection and SHA checks only. No tests, builds, installs, network, source edits, services or outward actions. The only written file is this report; no owned process. Actual candidate sizes, installed observations, final assigned base, implementation council and root acceptance remain future gates.
