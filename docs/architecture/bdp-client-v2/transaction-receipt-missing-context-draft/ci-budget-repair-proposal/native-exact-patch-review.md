# Native exact-patch review — existing cohort deadline regression

No actionable correctness finding in the applied one-file repair. Review is source-only; no test, build, import, service or Claude invocation was performed here. Actual selected validation and exact successor CI remain separate gates.

Bound source: clean committed base3968081f8231566908e69045424b067c3c36dc40 plus only apps/bdptest/src/transactional-matrix.test.ts dirty. Actual test hash6c0c240afef61831d0573775ea35b0c04b116a55da1d3fb40127b9ede68e939c equals proposed.ts. Preserved before.ts exactly equals that committed file (e319109ab0e9f7258e6918ae789ed7a5672eb7aabd0ec8725ae38e3634164e68). Correction patchfb4a530df7fea2b2afd7661b5ed655865908fbcc6e7ba6da296bea29c0867c5b matches the inspected change. Proposal/binding status text remains its historical pre-application record, not a claim that the current worktree is clean.

## Actual repair and caller audit

The private observe option is optional and validated before directory/child creation (lines98–130). Only the existing post-wire deadline callback supplies it (894–895). Searching every observe/budget/afterExchange caller in this file found no second opt-in. Its file-local scope is not an exported runtime API. All absent-option callers retain the same90,000ms default or existing external abort behavior; they allocate no second timer. The original abort-plus-cleanup and pre-aborted cases still follow their previous signal paths.

Lines162–164 preserve the initial timer, with the identical installed cohort deadline reason moved into one local callback. It starts before bootstrap launch/close, scenario preparation and first dispatch. The repaired case supplies10,000ms overall, replacing its former2,000ms startup-sensitive deadline; this is an explicit test-local increase, not a hidden claim that its former timing is preserved. No production, request, global test or CI timer changes.

At278–282, only a returned actual target.execute response can arm the optional100ms timer, and it is armed before the afterExchange hook. The timer handle is retained rather than reset, so subsequent exchanges cannot renew it. The original10,000ms timer remains armed; the earlier deadline wins and AbortController retains its first reason. This is actual timer expiry, not a fabricated response/manual control abort. The callback immediately requires genuine nonempty wireResponseBytes and sets seen before awaiting abort. A failed wire assertion is still a failed test rather than successful timeout evidence.

Runner.ts1193–1262 wraps the entire exact execute promise, including afterExchange, in its unchanged5,000ms request deadline. Merely raising the original timer to10s would let that request deadline preempt the held hook. The additional100ms first-response deadline ordinarily avoids that masking while allowing startup/preparation time. A response arriving close to the5s request boundary can still lose that race; event-loop scheduling and cleanup duration are not guaranteed. The diagnosis acknowledges this limitation, and the code does not accept an ordinary request timeout as a substitute for the required cohort-deadline cause.

Finally337–341 clears both handles regardless of target-close success, preserves primary/cleanup aggregation and removes the owned root. The second timer is not left active on successful or failed absent/opt-in observations. Native Node clearTimeout permits the undefined handle when no response was received.

## Oracles and scope preserved

The existing callback still checks the cohort-deadline error cause, seen=true, no successful publication, elapsed<15s, at least bootstrap+scenario children all exited/signalled, and no late publication after30ms. Its20s Vitest timeout remains. No assertion/input/route/body/schema/profile/coverage rows were removed or weakened. The real-wire hook blocks immediately at the first ordinary exchange, so its intended source-derived workload stays one HTTP exchange/two product children; no request/restart/case was added. This is static accounting, not a new execution record or guarantee that every failed attempt reached the wire.

The prior CI failure remains precisely a seen=false failure before the successful-wire marker under the old2s timer. The supplied diagnosis does not identify which startup/preparation stage consumed the time or prove CPU contention, a product fault, a cleanup leak, or the eventual successor's CI outcome. Cleanup assertions after seen were not reached in that original failure and cannot be retroactively claimed.

Only a test helper and its one existing case change. Production source/build inputs remain unchanged by the Git diff; this review did not independently inventory generated outputs or relabel earlier installed evidence. The superseded rearm proposal remains preserved and is not the applied patch. No additional required case or scope expansion is identified from this bounded source review.
