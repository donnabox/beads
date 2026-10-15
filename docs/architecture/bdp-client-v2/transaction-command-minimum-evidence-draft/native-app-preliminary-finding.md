# Preliminary native app finding at frozen f72c4f2c

**N1 — Medium: genuine completed runner report is lost on outer cleanup failure.**

Source: apps/bdptest/src/transactional-matrix.test.ts:308–334, positive writer:350–365, selected-control writer:2720–2729. Accepted v4 plan:102 requires the genuine returned report to be captured before assertions or cleanup failure can erase it.

observe assigns the real returned result to observed, then enters finally and awaits target.close. A failure there is added to cleanupFailures and causes AggregateError before observe returns. Both new output writers are caller-side after await observe. They preserve reports before supplemental app adjudication but cannot preserve the already-completed report on this cleanup path. The failure itself is correctly retained; the evidence artifact is not.

Smallest app-only repair: optional onObserved callback using a local observed-result shape/type, invoked immediately after observed assignment inside the existing try and before finally. Positive/control closures write the exact existing safe sourceCommit/entry/observed/supplemental object there. Preserve existing cleanup and return/error behavior. Callback write failures enter the existing primary-error path and do not bypass finally. Do not manufacture a result when the runner itself throws before returning one. No executor, runtime, runner API, request, child, deadline or generic framework expansion is needed.

Root raised this path; native independently traced and confirmed it. No source was edited and no execution occurred. Other reviewed app logic has no concrete finding so far; final report remains pending this disposition and accepted fence/source rebinding. The pre-fence Claude prompt packet is preparation only, not launched and not a binding for a repaired successor.
