# Preserve unexpected owner faults before Transactional comparison

Status: proposal for native + Claude plan council; no authoring or execution authority implied. This supports the sole plan, `/Users/dbox/repos/wt/vickie/bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md`, [Autonomous profile completion / current priority](https://github.com/donnabox/beads/blob/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/design.md). T3 serving correctness takes precedence over another partial evidence row.

Source basis: frozen `70ac666002708202d2f93195eed90dfe8541300d`, tree `1353e8f4fb0b1149b3f4ff158335804bf3f1cecc`, `/Users/dbox/repos/wt/vickie/vickie-comparison-timeout-review-repair`. Final authoring parent must be separately assigned verified main after PR120. Preserve this candidate and all historical reports. Re-anchor the five owned paths and read-only callers before drafting; do not rebase a frozen evidence source.

I read the completed independent [fault-shortcut audit](transaction-comparison-fault-shortcut-audit.md), SHA256 `1e16ad91aa31d5a68ab64314fa0d4373daae08b920137bafffd59f69f735c02f`. Its concrete reachability and authority conclusions agree with this source inspection. No tests, builds, service runs, packages, installs or network operations were performed for this plan.

## Settled behavior and defect

`docs/specs/bdp.md:1956–1960` requires unexpected internal faults to produce bodyless500 without a BDP Problem or internal details. `:4588–4592` repeats that exception for Transactional mutation targets. The normal finite-comparison503 laws at3994–4026 and unknown-key admission unavailability at4517–4520 do not reclassify a retained internal exception. No product ruling, new error code, context requirement, minimum policy, Event question or schema change is needed.

`transaction-store.ts:1340–1359` catches real maintenance failure, retains its cause, cancels its interval, closes SQLite and sets closed. `enter` at341–345 correctly prioritizes ownerFailure and maintenanceFailure over ordinary closure. But `comparisonEpoch` at1724 returns undefined immediately for every closed owner. `transaction-comparison.ts:52–56` then treats that as benign stopped comparison; observed submission maps it to contextless comparison-refusal, and HTTP serves retryable503. A new valid authorized direct POST is sufficient; it needs neither an unresolved reservation nor deadline timing. The correctly guarded final deadline observation at1726–1741 is never reached.

## Minimal production change and scope

Change only comparisonEpoch's benign-close predicate to require closed **and no retained ownerFailure or maintenanceFailure**. All other states fall through to the existing `enter(() => metadata("epoch"))`. Prefer the explicit short block mirroring the adjacent deadline guard. Do not add a catch, new owner API, callback, context synthesis, metadata read after closure, store format or broad lifecycle helper.

Exactly five owned files proposed:

1. `packages/server/src/transaction-store.ts`: the guard only.
2. `packages/server/src/transaction-store.test.ts`: extend three existing close/fault cases to cover comparisonEpoch.
3. `packages/server/src/transaction-submission.test.ts`: observed/unobserved propagation and normal-close exact-shape cases.
4. `packages/server/src/transaction-http.test.ts`: one emitted-runtime maintenance-fault scenario using existing source-only seams.
5. `docs/design/transaction-implementation-progress.md`: explicit historical residual/successor correction and evidence limits.

No edits to transaction-comparison/submission/service/HTTP implementation, runtime/configuration, evaluator, schema, discovery, public exports, installed process, fixtures or conformance catalog. Any inability to exercise the existing seams is a preflight obstacle, not permission to add an injection API.

## Caller and response boundaries

Production search finds one comparisonEpoch invocation: `transaction-comparison.ts:52`. `transaction-submission.ts:116–132` supplies either the actual store or an observed adapter; both delegate to the same method. Both must now reject on an internal fault; unobserved **normal/healthy** return shapes stay unchanged. Explicit normal closure still returns undefined, so existing benign stopped503 behavior remains. Defensive changed-epoch ordering, deadline counting and same-epoch retry law remain untouched.

The caller is Transactional-only. RU development and public Read do not use transaction-store/comparison. Dependency and sequence-alias tests call the private comparison function directly, so include one exact existing regression each. Transactional singleton/batch direct mutations share submission; the sequence executor uses owner entry directly and already propagates sticky faults. No Type/History/alias normalization or creator-fact behavior changes.

Receiving precedes owner comparison: submission's captured-body preparation/key validation/authentication at80–115 remains; earlier400/401/403/415-style refusals are not changed to500. The new wire proof uses a valid authorized request with acceptable JSON. Fault propagation reaches existing `transaction-http.ts:614` empty500 (helper80–88, marker at616), or runtime's sanitized fallback at253–260. Assert500, zero body/content-length0, absent Problem content-type, Retry-After, Location, development-limit and all three BDP context fields. Do not invent context for an internal fault. Non-attesting marker and private,no-store remain truthful. Internal sentinel details must be absent from response body/headers.

## Deterministic source observations

### Owner and submission

Extend the existing store cases `returns a stopped comparison observation for a normally closed owner`, `preserves a fenced owner's comparison observation failure`, and `preserves background maintenance failure in a closed comparison owner` to assert comparisonEpoch behavior at the same relevant point. Normal close returns undefined; closed fencing throws the same retained failure; actual timer maintenance failure throws the maintenance wrapper with its original cause. Preserve every current deadline-observation assertion. Existing fault hook,25ms fake timer and fixture cleanup suffice.

Add a two-row observed false/true submission case using a real owner with the existing before-maintenance-commit fault hook. Trigger the live maintenance callback, restore real timers, then submit a fresh valid request. Expect rejection retaining the fault (not either refused shape). Spy only on existing admission functions to assert zero admission calls; the fault must arise at the early epoch check. No synthetic owner-stopped outcome or new runtime seam.

Add two normal-close observed false/true cases pinning existing exact legacy refused versus comparison-refusal/owner-stopped shapes with no context; no callback or wait runs. These prove the narrowness of the fix rather than assuming the typed union preserves runtime behavior.

### Actual emitted HTTP

The existing HTTP fixture at `transaction-http.test.ts:84–193` calls the real emitted runtime, accepts `TransactionRuntimeTestSeams`, and supports a clock callback plus maintenance interval (`transaction-runtime.ts:51–66,319–327`). Use these existing options, not a new fault field:

1. Fake only setInterval/clearInterval, leaving network timers and setImmediate real. Start one disposable runtime with maintenanceIntervalMs25 and a clock returning the normal fixture value unless a local throw flag is set. Startup and a valid seed batch POST must succeed (200 receipt, exact expected seed facts).
2. Set the flag only while advancing the fake interval25ms. The actual owner maintenance callback calls instant at store295–299; that clock throws a unique internal sentinel. Assert exactly one injected call and timer cancellation. Reset flag and restore real timers before the next network request. This is a real callback/exception/closed-owner path, not a manufactured response. Store-level tests separately exercise the before-maintenance-commit hook; this HTTP test does not claim arbitrary disk-error coverage.
3. A second valid authorized direct batch POST with a fresh key must produce the exact sanitized500 properties above. No unresolved dependency or comparison wait is needed.
4. Close through the existing fixture and inspect `retainedReceiptInventory()` read-only: only the original seed receipt remains, with exact principal/key/receipt equality. No retry receipt or failed-key capture. This is receipt inventory evidence, not a complete graph-state/replica proof.

Exactly two new HTTP requests, one listener lifetime, one disposable store, zero installed children. Capture release/close outcome and preserve primary+cleanup errors. Restore timers even if startup or assertions fail; drain/close before directory removal. Do not modify historical global teardown. If selective fake timers cannot drive the existing emitted interval without affecting network behavior, stop at draft preflight and return the actual obstacle; do not fall back to sleeps or a new installed injection option.

## Selection, counterfactual and proposed execution bounds

Prospective14 source cases:

- Three extended owner cases above.
- Four new submission variants (fault × observed/unobserved; normal close × observed/unobserved).
- Existing submission `keeps full-owner unobserved comparison exact and excludes a caller callback` and `distinguishes a stopped final observation at the reached comparison deadline`.
- Existing service `uses the original finite comparison budget while an earlier sequence dependency is unresolved`.
- Existing dependency `proves static mismatch immediately, retains no timeout outcome and keeps the owner`.
- Existing alias `uses one bounded comparison deadline and leaves the owner's reservation intact`.
- New HTTP maintenance-fault case plus existing `keeps a live listener from serving through an owner fenced by injected commit uncertainty`.

The existing HTTP fencing regression has three requests and one reopen (two listener lifetimes). Thus proposed source positives total5 HTTP requests/three listener lifetimes. A single isolated baseline derivative restores only comparisonEpoch's old unconditional closed shortcut and runs the new HTTP case: after seed and actual callback failure, the second response becomes503. Require the named `maintenance failure is internal500` assertion, actual503 and retained cleanup proof; reject generic setup/timeouts as sensitivity evidence. Proposed control2 HTTP requests/one listener. Production checkout stays frozen; derivative is separate and restored. No additional mutants are presumed necessary.

Draft preflight must resolve exact parameter-expanded titles, request bytes and route inventory, source hashes, timer/control selectors, fault call counts, all temporary stores, and cleanup paths before execution. Provisional ceiling is14 positives+one derivative, at most24 temporary stores (exact count must replace this conservative bound),7 HTTP requests/four listener lifetimes/zero installed children for source/control combined. maxWorkers1–2, existing test/hook deadlines unchanged,120s command supervision; no unfiltered service/HTTP/full suite. Scope/semantics or measured-count overruns return to root, not silent omission.

After green source/static controls and a clean signed source freeze, propose only the existing installed **direct refusal14/1** regression on fresh actual packages, with its unchanged request/scenario/outer/cleanup bounds. This exercises the changed installed direct-call path in its ordinary healthy configuration; it does **not** induce a maintenance fault or qualify default installed fault sensitivity. No new installed test, clock option, child command or public flag. No sequence installed rerun is proposed: its unaffected execution path is covered by the bounded source service/alias regressions. Exact pack/install/request/PID/config/payload attribution must be measured in the later draft preflight. Root may omit or amend this ordinary regression during plan acceptance; it is not authorized now.

## Done-when and history

Before source execution: native+Claude plan review, root parent/ownership release, then normal-format draft and measured preflight acceptance. Build before emitted-runtime tests; preserve every failed attempt and do not automatically rerun unexpected failures.

Delivery predicates: only the five approved files change; normal closed behavior stays exact; both retained fault terms reach existing enter; observed and unobserved submission reject before admission; actual emitted wire is sanitized bodyless500 without retry/context/detail; old guard reliably produces named503 failure with cleanup; selected dependency/alias/healthy deadline regressions pass; static/Read gates pass; source-bound package regression (if accepted) is clearly separate from fault sensitivity; all owned processes are absent.

The current progress text at1781–1784 is historically accurate for30035/70ac. Reword it to say that those snapshots left the early shortcut unrepaired and add a separately dated/source-bound successor note once this slice is verified. Do not retroactively change their reports, prior18/five or27/two evidence, or claim this repair was part of PR120. No public profile activation/advertisement, broad receiving guarantee, live epoch rotation, Event/minimum ruling, automatic recovery policy or full conformance claim.
