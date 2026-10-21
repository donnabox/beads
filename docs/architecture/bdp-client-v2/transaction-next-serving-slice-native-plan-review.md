# Next serving slice — native plan review

## Finding

### P3 — Specify a retrying positive comparator that distinguishes captured options from rereading caller options

**Plan location:** `transaction-next-serving-slice-plan.md:46`, together with immutable-per-attempt requirement at line24 and done-when3 at line56.

The abort half of the proposed held-wait test is useful cancellation coverage, but it does not distinguish the old service's mutable receiving input from the proposed capture. Existing `transaction-service.ts:309–315` evaluates `receiving?.signal` once when calling the wait helper. After a wait, `waitForTransactionMinimum` checks that original signal before calling the next attempt. Thus mutating the caller object and then aborting the original signal produces no second owner entry even with the old service branch. The plan's "otherwise identical non-aborted call" does not specify what makes the replacement signal distinguishable: replacing it with another active signal could still let the old implementation pass.

**Minimal fix:** make that already-planned successful comparator explicit: leave the original signal active, replace the caller's signal with a pre-aborted signal while the first wait is held, perform the independent Resource commit that reaches the original minimum, then release the wait. Require success at the new eligible context under the original signal and expected second owner entry. The old service would pass the replacement aborted signal to the next facet attempt and fail; the captured implementation succeeds. Keep the separate original-abort/no-next-entry case. This adds no installed request, production API, wire semantics or extra framework; it specifies the existing source-test comparator rather than expanding the slice.

No other actionable plan finding. The underlying four-file repair is feasible using existing receiving capture and finite-wait mechanisms. This report is not implementation authorization.

## Exact review basis

Plan SHA256 **d245c319cd9f5e236a31afd6a7fe46e579896afe635d009a2aab6d1d832af676**. Inspected source at frozen alias repair **3959a25e70bc6dc9d143abfead187ccaf84ca11e**, tree **7d6212ea04046d7c45a337b293a13a482a397caf**, and its actual authority receiving, service run/close and consistency-wait functions. Main/integration baseline described by the plan is not treated as including alias repair prematurely. The separate native closure of3959 is now complete; root still owns its remaining council/integration and assignment of an exact implementation base.

Read-only inspection only: no tests/builds/services/network/source edits, no implementation or publication. This report is the only new file from this plan review.

## Feasibility, phase order and bounds

The inherited perform spread and unchecked service signal access are real source-level analogues of the alias defect. The helper already rejects precisely authority-invalid option/signal shapes, retains non-enumerable signal and strips/replaces only opaque caller condition. Reusing it avoids adding receiving policy or a public API. A source-private rename, if chosen, need not alter behavior or package exports.

Preserve existing phase boundaries while applying it: service `run` checks admission before capture; direct perform keeps the current supported-request predicate before authority delivery; for valid supported requests, capture-failure delegation leaves authority guard/receiving order intact. Direct closed delivery remains an observed rejected Promise, while service closed admission remains its existing synchronous throw. The plan's reference to existing lifecycle behavior should be implemented with those distinct expectations, not normalized into one API contract. No malformed-request predicate redesign is needed.

The immutable minimum remains captured when constructing the facet. Every successful capture must yield detached options reused on retries; caller condition cannot lower the bound. Existing helper failure delegates only known-invalid inputs, which authority rejects before observation; it must not accidentally create an unconditioned successful fallback. No owner entry spans a wait, and ordinary authority behavior remains untouched.

The four named source/test files contain the necessary seams. Existing real-owner counters, controlled waits and lifecycle fixtures can exercise the proposed descriptors, signal mutation and reentrant cases without new production hooks. Preserve cooperative cancellation limits: abort cannot force an unresolved custom wait to settle or preempt synchronous SQLite work. Use the existing controlled-clock source-test approach when holding waits.

Fresh existing collection38/1 and Type/incident20/1 plus targeted Update2 observations are honest combined-runtime regression evidence, not hostile-JavaScript input proof. Alias18/1 can be separately counted if the helper is renamed as proposed. No fixture/cap/deadline/route/advertisement expansion is justified or required. Public discovery and complete cold-client readiness remain outside this correction.
