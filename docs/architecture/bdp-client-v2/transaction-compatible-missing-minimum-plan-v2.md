# Compatible missing-Resource minimum — implementation plan v2

Supporting plan under the sole Vickie design plan. Root-proposed council dispositions are in `transaction-compatible-missing-minimum-review-dispositions.md`; independent closure and root dispatch remain required. No source change, test run, serving activation or new protocol law is authorized by writing this revision. Original plan SHA256 `07fa069b8519c6fe06d74cc236b3d3e941962771140c35485a777208ea92d65b` is preserved.

## Scope, basis and affirmative contract

Repair the current development500 for a missing/deleted canonical Bead or Link, including `view=properties`, when the complete supplied minimum is compatible and already satisfied. Return the existing observed404 with the same-entry current context. Preserve missing+behind/foreign gaps, unsupported receiving controls and all HTTP route predicates.

Source basis: main `aebf656506d187ed226302b68b526d34088537c6`; inspected source also in frozen C3 repair `a53f4bbf573c64fe900de373b3fef75170002f7c`, `/Users/dbox/repos/wt/vickie/vickie-tx-discovery-repair`. `transaction-read.ts` SHA256 remains `27e57473295d494e172107a5ea0f939e39429d66cbb6196f065d107b0dde7e1e`. Parent will assign an accepted ordinary integrated base after C3/projection; recheck named seams and preserve all ancestor changes before authoring. Do not edit either frozen review source. Fresh Snapshot creation minimum support already landed in aebf; this is its independent serving follow-on.

Affirmative normative basis in `docs/specs/bdp.md` at aebf:

- **Reads after deletion:1445–1450** requires uniform resource-not-found404 for unknown/deleted/non-visible ordinary Resource/properties reads. History-specific optional410 is not used by this configured target.
- **Transactional Resource read context:5793–5799** requires every Transactional Scope-bounded read to report its actual observed prefix context.
- **Strict-read paragraph:5758–5765**, **Transactional HTTP consistency:5766–5787**, and **conditional reads:5801–5811** require eligible observations and fences before success shortcuts.

The selected interpretation, expressly accepted in the prior consistency audit, is that the consistency enumeration includes a **Problem representation of an already-satisfied observed prefix**. Combined with the affirmative404/context rules, the current header-free development500 is the defect; this is not merely permission to prefer404. There is no failing consistency condition left to arbitrate in this narrow case. No new product ruling is proposed here. Simultaneous absent/hidden and behind/foreign failures remain unresolved for these Resource routes.

Do not imply a uniform ordering across this owner. Unknown Resource **Event Sources** already use a documented local404-before-minimum convention without context (`transaction-implementation-progress.md`, Event projection paragraph around1058; HTTP Event route). Unknown **receipt pages** similarly retain404 without context under arbitrary controls (`transaction-http.test.ts`, “keeps failed parent rows unchanged for nonexistent page reads with any controls”). This patch does not copy or alter those route-specific conventions. Reconcile the existing minimum-subset prose around661–671 with the Event paragraph: compatible current Resource/properties404 now carries observed context; incompatible missing Resource requests retain their private gap; Event/receipt conventions remain separately documented, not promoted to a normative global arbitration rule. Fixed all-visible policy does not establish hidden-policy behavior.

## Source mechanism and exact production edits

Only `packages/server/src/transaction-read.ts` changes in production. Existing `authority-read.ts:977–998` captures, materializes and checks in one owner entry; `observed():509–515` keeps the observation on the Problem, and `finish():852ff` preserves it. HTTP `readRoute.perform` already attaches that observation's current context. No second read, fabricated context or HTTP production edit is needed.

1. **Close the private ordinary-facet escape.** In composed `readFor().perform` and `.resolveAlias` (`transaction-read.ts:93–100`), forward only the caller's optional `signal`; do not forward `sourceCondition`. This is an explicit **narrowing of the internal ordinary Read facet**: callers that previously injected a condition through generic options will now receive ordinary no-minimum behavior. Only `minimumFor` may attach a source condition. No repository call site relies on the ordinary forwarding; HTTP minimum routing already selects `minimumFor`. Do not claim all in-process behavior is unchanged, change the generic authority API, add a new option or widen any HTTP predicate. Preserve signal forwarding and cancellation.
2. In `checkSourceCondition` (`:56–69`), preserve the early materialized invalid-parameter/limit-exceeded return first. **After that return**, snapshot/validate the tuple and parse `transactionReadContexts(observation).current` **outside** the special translation catch.
3. If the materialized refusal is resource-not-found, run `checkTransactionMinimum` with those validated values. On success, explicitly `return`, allowing the original observed404 through without a second checker call.
4. Catch only `error instanceof TransactionMinimumFailure` with reason exactly `behind`, `foreign-view` or `cursor-expired`; translate only those collisions to the existing `TransactionMinimumVisibilityGap`. Rethrow every other error unchanged. The reason allowlist defensively excludes catch-up-timeout, which the synchronous checker cannot produce. This impossible-producer branch is deliberately not given a complicated module-mocking test. No duck typing or broad catch.
5. Other results use the normal checker unchanged. Present-target behind reads continue to wait. Missing+behind stays a distinct Gap, so `waitForTransactionMinimum` (retry loop:106–120) does not poll it. Unknown spelling/partial tuples remain rejected at the existing receiving/facet boundary.

Validation of corrupt source context now changes the missing-target error flavor intentionally: previously the early visibility Gap masked malformed context; now its integrity error propagates, yielding the existing sanitized500 **without** the visibility marker. This is a disclosed correctness delta, not a new target refusal code. No condition is parsed on the preserved early-control-return path.

## Five owned files and concrete test obligations

1. `packages/server/src/transaction-read.ts` — callback and ordinary-facet option narrowing above.
2. `packages/server/src/transaction-read.test.ts` — actual-owner matrix and sensitivity tests.
3. `packages/server/src/transaction-http.test.ts` — compatible/deleted/property/conditional/HEAD observations and preserved collisions.
4. `apps/bdptest/src/transactional-collection-minimum-installed.test.ts` — separate ten-request scenario and narrow label/counter support.
5. `docs/design/transaction-implementation-progress.md` — reconcile both existing paragraphs and report precise scope/evidence.

Source tests use real create/delete commits and the existing guarded owner fixture. Cover absent/deleted × Bead/Link × record/properties × equal/earlier satisfied bound; assert actual `phase:observed`404 and exact current context. The small HTTP matrix covers GET/HEAD, Problem schema/required fields (extensions allowed), private,no-store, no Resource ETag, and If-None-Match:* / mismatching If-Match still404. Anonymous401 remains context-free. Present-target current/behind/foreign behavior is a regression, not changed law.

Split the old “does not invent absent Resource versus minimum ordering” test (aebf line1190; use symbol on the assigned successor): satisfied case now404/context; behind, foreign epoch and foreign view remain bodyless500 with minimum-position-visibility and no context. Unrecognized spelling keeps the distinct minimum-position receiving marker. Use the existing injected wait counter to prove missing+behind never polls; no flaky elapsed-time assertion.

For ordinary-facet narrowing, inject a valid **behind** sourceCondition into ordinary `readFor().perform` and `.resolveAlias` and prove ordinary results are unaffected; a satisfied condition alone would not discriminate accidental honoring. Separately verify the same supplied bound through minimumFor still rejects/waits according to the supported request, out-of-subset minimumFor remains rejected, and an AbortSignal still reaches the ordinary facet. This is internal seam testing, not a wire-route expansion claim.

Integrity test: use a delegating real-owner wrapper that substitutes **string** `scopePosition:"pos-02"`, with otherwise valid epoch/view. Generic authority capture accepts strings, so this input reaches the Transactional callback and fails its exact grammar. Assert neither returned404 nor TransactionMinimumVisibilityGap. A non-string failure alone is insufficient because it would fail before this callback. Preserve exact arbitrary/provider error propagation through an existing delegating owner failure; do not add a runtime injection option. Check owner remains reusable after failures and same-entry access remains guarded.

Bounded source sensitivity controls, recorded separately from installed proof: (a) restore only the old unconditional missing Gap → compatible case fails; (b) replace missing branch with unconditional return → behind/foreign cases fail; (c) move the Transactional context parse into an overly broad visibility catch → pos-02 discriminator fails. Run only the focused relevant tests for each disposable source mutant and retain red exits plus the unmodified green result. No mutation of the frozen author source, no extra installed mutant package, no mock framework for unreachable catch-up-timeout.

## Installed proof: exact10 exchanges,1 child,9 URLs

Reuse the existing fixture/artifacts and exact raw transport. The Task Type **owns** outgoing Relationship Links: owned create/delete changes the source Bead's revision as well as the Link. Keep these semantics; do not change the shared fixture. Maximum live state2 Beads+1 Link. Two fresh-key actual batches have3 then2 members, within32. No DB seeding, reopening or extra GET. History is false and policy all-visible.

|#|Request|Required row-local observation|
|---|---|---|
|1|Batch: create Beads a,b, then Link l from a to b.|Completed available receipt with exactly3 **per-operation results**, indices0–2, exact outcomes and authored canonical IDs/types/properties/endpoints. Link result includes source=a and sourceRevision. First Bead create result is its per-step image, not final a. Actual epoch/view and position captured; independently expected effectPosition=requiredPosition=current position=pos-1.|
|2|Fresh-key batch: delete l, then delete a.|Exactly2 per-operation deletion results, indices0–1. Link deleted identity/type/revision equals its created result; owned deletion source=a with a new sourceRevision. Deleted a revision equals that immediately preceding sourceRevision. Actual position2 is independently expected for second effective commit; effectPosition=requiredPosition=current position=pos-2.|
|3|GET never-authored Bead record; earlier bound from#1.|Canonical404 Problem/current pos-2 context.|
|4|HEAD never-authored Bead properties; equal bound#2.|404/bodyless/current context.|
|5|GET never-authored Link record, If-None-Match:*, equal bound.|404 not304; actual Problem/current context.|
|6|HEAD never-authored Link properties; earlier bound.|404/bodyless/current context.|
|7|GET deleted a record; equal bound.|Canonical404/current context.|
|8|GET deleted a properties; earlier bound.|Canonical404/current context.|
|9|GET deleted l record, mismatching If-Match; equal bound.|404 not412; actual Problem/current context.|
|10|GET deleted l properties; recognized future pos-3 with actual epoch/view.|Bodyless500, exact minimum-position-visibility marker, no Scope context. No waiting; discriminates blanket404 while preserving existing collision handling.|

Assert owner revision relations without invented tokens: #1 Link sourceRevision differs from a's create revision; #2 Link deletion sourceRevision differs from #1's sourceRevision; #2 deleted a revision equals #2 Link deletion sourceRevision. Every observed revision must be a valid nonempty token under the existing schema. Batch2 may use the actually captured Link revision as its expectedRevision but must not guard a with its stale creation revision. The ownership update is mandatory in both batches, not merely possible. b is target-only and not mutated by these commands; this ledger does not claim an independent final b GET. No exact changes/Event count is asserted, and no compact receipt allocated field is expected. Existing schema and per-operation outcomes supply the receipt shape.

Each batch advances Scope position once despite multiple Resource revisions. Requests use captured epoch/view/positions, with literal pos-1/pos-2 serving as independent implementation expectations for the fresh store's two effects. Row10's recognized next position is deliberately a negative input, not purported issued history. Deleted resources are proven created then deleted by successful per-step results; never-authored IDs occur in neither batch. Source tests cover the positive deleted-Link-properties case and broader method/collision matrix not observed by installed row10. No hidden-policy or complete minimum qualification claim.

## Fixture bookkeeping, budgets and evidence

Minimal helper edits inside owned file4: thread a **defaulted label** through withInstalled→start→report, default exactly the existing `private-installed-collection-minimum`; new scenario uses its own distinct label. Existing scenario output schema/label and request behavior stay unchanged (natural source/config/session values remain run-specific). In the new scenario wrap **every actual exchange locally**, including both batch writes: use the existing exchange directly with fresh explicit keys or ensure the local wrapper sees write calls. Assert <=10 during the scenario and exact10 on success. Do not wrap only GETs while write() calls escape counting. Reuse the helper's safe completedResponses report; its general80-response ceiling remains unchanged. Per-start state and distinct labels keep ledgers separate; failed runs retain actual counts and failure, never fabricated missing responses.

Preserve5s request,90s cooperative aggregate,120s outer,10s lifecycle,32768 response-body,16384 header and65536 request-body limits. No cap/timeout changes, added framework or public format. Actual installed proof <=10/1; source mutants add no installed calls. Do not conflate the separately rerun old collection scenario with the new count.

Run pinned Node24.16/pnpm11.20 clean prebuild typecheck→build→focused Read/HTTP and relevant unchanged service/consistency regressions, format/lint/boundaries. Then freeze clean ordinary source and run genuine retained build/pack/offline install with both installed entry/source variables set, targeted new10/1 test and existing scenario as a separately accounted regression if root requests it. Preserve source/tree/tarball/payload/config/fixture identity, exact safe response hashes, commands/exits/timing and child cleanup. No old source evidence relabeled, no automatic broad suite, no additional forced mutant installation.

## Done when / stop boundaries

- Exactly five approved files differ from root-assigned accepted base; both C3/projection lineage and existing old scenario remain preserved.
- Compatible observed404 has actual same-entry context; missing collisions/unknown receiving controls retain their old gaps; ordinary-facet sourceCondition is explicitly ignored while signal works; corrupt callback context escapes as integrity500 without visibility marker.
- Green focused tests and meaningful three source-mutant reds recorded; impossible checker catch-up-timeout branch disclosed as defensive/unexercised.
- Fresh exact-source10/1 has all receipt-step/revision and status/context oracles above; no child remains and no failed/aborted run counted pass.
- Progress documentation explains route-specific conventions without claiming owner-wide arbitration; public discovery/admission/catalog/coverage flags unchanged.

Stop for additional production/shared files, route widening, new law, altered original installed behavior, lost source context, extra installed calls/children or a budget overrun. This plan explicitly accepts the bounded internal ordinary-facet narrowing and corrupt-context marker correction; no other semantic expansion is authorized. Planning ran no tests, builds, servers, source edits, network or outward publication.
