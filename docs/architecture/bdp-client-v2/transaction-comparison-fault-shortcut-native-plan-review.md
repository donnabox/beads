# Comparison fault shortcut: independent native plan review

**No blocking finding. One Low drafting precision below.** The plan is implementable within the five proposed paths and existing private seams. It follows settled fault semantics, reaches the actual installed runtime implementation through the existing emitted HTTP fixture, and keeps that controlled source evidence separate from ordinary installed regression. Root still owns plan acceptance, parent assignment and measured execution release.

Reviewed plan SHA `ce2cd4b77843d13d963c8000f535129961dce14f69c6fbb8a865ad6bf6624af0` against frozen `70ac666002708202d2f93195eed90dfe8541300d`, tree `1353e8f4fb0b1149b3f4ff158335804bf3f1cecc`, in `/Users/dbox/repos/wt/vickie/vickie-comparison-timeout-review-repair`. Read actual store/comparison/submission/service/HTTP/runtime/configuration and selected fixtures/cases. No implementation or test execution was performed.

## Finding N1 — Low: specify delegating spies for the frozen owner

Location: plan “Owner and submission,” instruction to spy on existing admission functions. Source: transaction-store.ts:1432 returns an Object.freeze owner; transaction-submission.test.ts:1028, 1072 and 1112 already use vi.fn delegates on a spread adapter. Direct vi.spyOn(f.store, "admit") or vi.spyOn(f.store, "admitObserved") cannot replace frozen properties and would fail in setup rather than prove zero admission.

Concrete fix: say to wrap the actual admit/admitObserved functions with vi.fn on the existing spread adapter, leaving comparisonEpoch and all fault state delegated to the real owner. Preserve overload typing using the existing typed boundary pattern if needed. This needs no production seam, extra case or budget increase. The plan's word “spy” need not imply the invalid API, so this is a precision improvement rather than a blocker.

## Law, status path and callers

The minimal predicate change is correct. transaction-store.ts:341–345 prioritizes retained ownerFailure/maintenanceFailure; the old closed check at 1724 suppresses them. Restricting that return to normal closure mirrors the adjacent final-observation guard at 1731 and lets the existing exception path handle faults. It adds no storage entry after closure, because enter throws before metadata is read.

Spec 1956–1960 explicitly requires bodyless 500 with no Problem/internal details; 4588–4592 carries that rule onto Transactional mutation targets. Ordinary deadline/epoch-stop 503 and admission capacity 503 are distinct. HTTP's existing empty(500) at 614 supplies content-length 0/private,no-store and no Problem type; the marker is added at 616. The runtime fallback is separately sanitized. No new context rule or minimum/Event ruling is needed, and absence of a fabricated fault tuple is consistent with this narrow existing error path.

The single production comparisonEpoch caller at transaction-comparison.ts:52 reaches it before deadline/admission, and both observed and unobserved submission adapters delegate to it. The behavioral change for both is intentional fault rejection; healthy legacy shapes remain unchanged. Normal explicit closure still returns undefined. Source request preparation, key checks and authentication in transaction-submission.ts:80–115 remain ahead of comparison; HTTP media/method routing remains ahead of submission. Thus the guard does not override earlier malformed/auth/media refusals. Sequence execution and Read owner entry already propagate faults; the plan correctly avoids implementation changes there.

## Fault injection is reachable and appropriately scoped

The HTTP fixture at transaction-http.test.ts:84–193 loads the emitted runtime and passes the existing clock and maintenanceIntervalMs options. Runtime opens the real store with those values at 319–327. Store's interval at 1340–1359 invokes maintain; maintain enters atomic and evaluates instant at 512–514; instant calls the captured clock at 295–299. Throwing only while the fake interval advances therefore raises the sentinel inside a real maintenance transaction. With ordinary successful rollback this retains maintenanceFailure, cancels the interval, closes SQLite and sets closed. It does not manufacture an owner-stopped result or HTTP response.

Faking only setInterval/clearInterval is materially narrower than freezing network scheduling: fetch timers, setImmediate and comparison's clock remain real. The store is the only production user of setInterval in this server source. Seed startup and receipt delivery occur with the flag false; the flag is true only during the synchronous 25 ms advance, then reset before the second POST. Exactly one throwing clock invocation and cancelled maintenance timer distinguish the intended state from a setup fault. The second fresh-key request tests the early epoch shortcut; unresolved dependencies and default timeout timing are unnecessary.

This is a controlled emitted-runtime observation, not deterministic default-installed disk-failure qualification. The plan states that limit correctly. It also requires stopping at preflight if selective timers prove incompatible instead of inventing a runtime or installed injection API.

## Oracles, control and cleanup

The three existing store tests exist at 3267, 3289 and 3320. Extending them without removing final-observation checks covers normal closed behavior and both sticky fault terms. Four new parameter-expanded submission variants cover fault propagation and normal-close exact results, with zero admission/wait/observation activity where required. The real maintenance fault must be distinguished by original cause rather than accepting any rejection.

The new HTTP case's seed receipt, named bodyless-500 assertion, absent retry/Problem/context/internal details and final exact retained receipt inventory provide complementary oracles. Inventory is correctly labelled receipt evidence rather than proof of every graph invariant. No guessed receipt URL or new capture route is needed. The old-guard derivative should pass setup/timer assertions and reach the named second-response status failure with actual 503. Its successful build, exact executed patch, named failure, actual status and cleanup must all be retained; a generic failure cannot satisfy it.

The plan explicitly requires primary and cleanup errors to survive together, real timers restored on startup/assertion failure, and runtime drain/close before directory removal without modifying historical global teardown. Draft preflight must make that finally structure concrete, including the control failure path. Current fixture retainedReceiptInventory closes runtime and its read-only database in finally; fixture teardown removes the directory. Those facilities support the proposal. No extra execution is needed to decide plan feasibility.

## Selection and bounds

All six named existing regressions beyond the three owner cases exist: two submission cases, the finite-budget service case, dependency mismatch/timeout case, alias bounded-deadline case, and existing HTTP fencing case (with the new HTTP case completing the proposed group). The arithmetic is 3 extended owner + 4 new submission + 2 existing submission + 1 service + 1 dependency + 1 alias + 2 HTTP = 14 positive executions. New variant titles and exact parameter expansion remain a measured-preflight requirement.

Existing HTTP fencing at 867–890 issues three requests and reopens once, hence two listener lifetimes. New HTTP positive adds two requests/one lifetime; its derivative adds two/one. Combined proposal is seven HTTP requests/four listener lifetimes/zero installed children. The 24-store ceiling is explicitly conservative and must be replaced by measured fixture accounting before execution; no inconsistency in the upper bound was found. Counts include no implicit full suite. Existing test deadlines and 120-second supervision remain intact.

The optional direct-refusal 14-request/one-child installed regression is proportionate to a production direct-call guard change, provided root explicitly accepts it and later preflight binds exact routes/packages/configuration/cleanup. It is a healthy regression, not fault sensitivity. Omitting sequence installed rerun is justified by the unchanged execution path and selected source service/alias regressions. Neither optional execution nor public admission is authorized by this review.

Progress history must retain the original 30035/70ac residual claim and add a separately source-bound successor statement, as planned. No source changes, tests, builds, installs, services or network calls were performed. Only this review artifact was written.
