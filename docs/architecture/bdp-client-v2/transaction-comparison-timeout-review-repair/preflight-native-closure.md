# Comparison review repair: amended independent native preflight closure

**Clean bounded static preflight: no blocker or additional actionable finding.** The four-file successor fixes the genuine stale service oracle and its assertion-failure cleanup path while retaining the accepted F1/F2/F3 repair. The enumerated workload is suitable for root execution release. This is not an execution authorization, test pass, full code/evidence acceptance or CI waiver.

## Exact basis and preservation

Reviewed the stable draft in `/Users/dbox/repos/wt/vickie/vickie-comparison-timeout-review-repair`, HEAD `30035b88380b7cbb3abd5359787db236b70345c7`, tree `cdee4bc7e31511cde094d17c045ea618163b6d7a`, against its preflight, bounds, source diff, controls, preservation manifest and the actual CI failure. Exactly four paths change, 159 additions/21 deletions: store, submission and service tests plus progress prose. There are no production changes.

Independently verified:

- preflight.md: `a3f89340417f2aeff2f07c2d9374d8fc5300e7437de5f5f2af7143fb968d3639`.
- preflight-bounds.json: `711b47035b179faf83ef8a14631aa871478bd505e9746186e49b4a5da96ef9df`.
- source.diff: `17eab65f6c95fa6e49f848de23dd7d342682c8b1dfaadb39c7f3bd402aa913a4`.
- preservation.json: `7a47c3c3c0af523e2286b976d614f6759bdbfa98a5eca3e08101427ed3ee7b9e`.

Live diff bytes match the artifact. All four changed hashes and lengths match bounds. All 659 other tracked files match their HEAD blobs and manifest hashes, including production, packages, installed tests, workflow and configuration. The original three-file closure is preserved at `initial-preflight/preflight-native-closure.md`, SHA `1f42b8dc4331c718b27fa9cc0183a66f464479455fdc120e2f5db421b7d3411d`. Its three reviewed file hashes remain unchanged; this successor replaces only its current applicability to execution release.

## CI amendment: correct contract and cleanup

The actual full CI log at lines 11008–11058 records the service expectation mismatch, a subsequent 10000 ms afterEach timeout, 9253 passed/1 failed/1 skipped, and exit 1. It is a genuine failure, not a timeout-only retry candidate. Production service submission already deliberately calls submitTransaction with observed=true (transaction-service.ts:499–515). The returned comparison-refusal/deadline/current context therefore agrees with the accepted implementation; no hidden production defect or reason to restore the legacy result was found.

The service test at lines 439–486 captures a trusted same-owner context before admitting the sequence and independently expects pos-0. Its strict named oracle checks the complete observed refusal, including reason, epoch, authorization view and position. The original two waits, empty pre-turn graph, two scheduled executions, retained completed retry at pos-2 and exactly two committed groups remain. No receipt is fabricated and no additional graph mutation is introduced.

The immediately attached settled handler observes both job outcomes before any assertion can abandon delivery. The finally path releases at most the two carrier members, awaits the captured outcome, then closes the service. The manual scheduler removes each callback before invocation and flushes the same two microtasks used by the original passing completion path. Under the old-shape control, the first assertion fails before either turn, so the same two-member release is reachable. Under the positive path, both turns have already been consumed, so cleanup executes zero turns.

Primary assertion, scheduler, job and close errors are collected rather than replacing one another; AggregateError retains them for the pinned Vitest reporter's existing aggregate-error flattening. This does not claim arbitrary scheduler faults can never hang: unchanged Vitest deadlines and the bounded supervisor still govern unexpected nonsettlement. The accepted deterministic control tests the concrete old assertion-failure path. Existing afterEach performs idempotent service/owner close and directory removal after the explicit drain. The settled discriminated outcome, unknown error list and captured context introduce no evident TypeScript mismatch. The mutant retains use of the initial context through the pos-0 precondition, so it introduces no unused-variable obstacle.

## Controls and measured bounds

The prior F1/F2 typing, cleanup and guard sensitivity findings remain closed as documented in the preserved original closure. F1 covers the closed fenced and background-maintenance failure terms of the correct production guard; the two guard controls reach their named assertions without earlier masking. F2 remains the narrow submission-only stopped observation double with exact contextless shape and finally cancellation. Progress prose only reflows historical evidence.

Independently reconstructed the added service old-shape mutant. Patch SHA `be09970db8718cdf4d6bf6fd0c31c857ad7a8e25d7f7735c7508724a7addc63a` changes only the expected object at the named strict assertion. Original test SHA `03f4e5c53cc8717a9818de97dd7a4c606cb1db6018844481e7cd43a18d55dd20`; reconstructed mutant SHA `1193c43ed5fec1f7937be7d1cdfb4fb5833246f349a35a7bdf7fb101cb6fa631` matches bounds. Cleanup code remains identical. It must fail specifically at `observed service comparison deadline shape`; compiler/setup failures or a hook timeout cannot satisfy the control.

Exact selected workload is four positive cases and three controlled case executions: seven temporary stores at most; two in-process service instances total (one positive, one derivative); zero HTTP requests, listeners, external services or installed children. There are two admitted sequence members per service run. The maintenance cases retain one fake timer and one 25 ms advance each. No new sleep, retry, unfiltered suite or installed rerun is authorized by these bounds. MaxWorkers1, existing Vitest deadlines and proposed 120-second per-command supervision remain unchanged.

Adjudication must parse exactly one actual cleanup record from each service run. Both require drained=true, closed=true and pendingCallbacks=0; positive requires cleanupTurns=0 and the old-shape derivative requires cleanupTurns=2. Missing/duplicate records, mismatched fields or unrelated failure cannot pass adjudication. These values are predictions to verify after release, not observations from this static review. Absence of a hook timeout alone is insufficient cleanup evidence.

No tests, builds, compiler, install, services, network or source writes were performed. Only this closure artifact was written. Frozen repair/evidence review and exact-head CI remain root gates; original installed evidence and unresolved minimum/Event/public-admission scope remain unchanged.
