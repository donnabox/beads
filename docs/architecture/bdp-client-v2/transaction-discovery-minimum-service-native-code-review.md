# Discovery minimum service — native code review

## Findings

**No actionable findings.** The frozen implementation follows the accepted internal-only service plan, preserves the existing no-minimum behavior and does not activate discovery or public profile admission.

Reviewed source **2bd484367f63c7b8cc25e144baec983c5e579798**, tree **20a494de5555b39ece1e3eb09e9dd2ad37cbd554**, against direct parent **26d6cd1b947db8c00e24015771506d26098a293e** in `/Users/dbox/repos/wt/vickie/vickie-discovery-minimum-service`. Git status was clean. Exactly the three approved files differ: transaction-service.ts, transaction-service.test.ts and the progress document. Accepted plan SHA256 is `ee9de18ed49dcb296a91816ee0156e76d9b4a4afca1a5d4fc1519d37655279a6`; root acceptance and the final author implementation report were read.

## Production assessment

- **Same-entry eligibility and one callback:** transaction-service.ts:239–252 snapshots controls once, enters the actual owner once per attempt, compares against that callback's context, and copies/freezes precisely the same epoch/view/position. There is no second head lookup, retained reader or asynchronous owner callback. Existing transaction-store.ts:883–959 releases the callback through its synchronous owner lifetime even on throws.
- **One finite service lifetime:** lines 262–264 put the complete wait and final assembly inside the existing run registration. The unchanged consistency helper establishes one deadline and retries only an actual behind failure; foreign context, provider/refusal and parser errors do not become retryable. No timer, retry policy or close mechanism was added.
- **Release before assembly:** lines 254–264 assemble the successful captured context after owner release and outside the retry callback. Parsing cannot be retried as a behind observation, and a later independent commit cannot refresh only one half of the returned candidate/context. The owner-refusal wrapper preserves its cause at line 251.
- **No-minimum compatibility:** line 261 retains synchronous observation and assembly inside the initial async function invocation, before its promise delivery. The new optional signal and minimum clock are deliberately inert on this branch. Valid minimum calls still pass through the existing service acceptance check; malformed tuple capture follows the accepted Snapshot-style ordering rather than introducing a universal precedence claim.
- **Cancellation/drain:** the unchanged finite helper checks abort before attempts and between waits, not during synchronous work or by adding a post-success cancellation barrier. The existing active set registers the entire call before start; close stops new work and drains admitted work before closing Read and the owner. Source tests distinguish this from immediate interruption.

No production edits to HTTP, runtime, store, shared consistency, schemas, catalogs, public exports or configuration were found. Captured limits and pagination receiver behavior remain untouched.

## Test and evidence assessment

I independently compared the predecessor and successor test text: all existing test-body bytes, including receipt-route, replay/expiry/reopen, limit and callback-receiver tests, are preserved. Only required imports precede those bytes; the new suite is appended.

The 11 new cases use real owners, controlled clocks and deferred waits. They cover synchronous no-minimum timing; equal/earlier/genesis eligibility; post-release parser access and immutable output; caller-tuple mutation during catch-up; a concurrent satisfied observation and genuine independent commit; foreign epoch/view and owner reuse; fixed-head timeout with no owner state change; pre-/pending abort; close/drain; malformed input, real reentry/closed owner, delegated provider/refusal causes and parser recovery. Deferred releases are registered after service cleanup so reverse cleanup releases held waits before draining the service. The catch-up entry total correctly includes the independent satisfied observation (three entries total, two for the waiting call).

Inspected the actual author JSON and logs rather than treating the report headline as proof:

- Focused run: **183 passed, 0 failed/skipped across six files** — service55, consistency2, discovery23, Read28, HTTP62, emitted installed-contract13.
- Restored focused new suite: **11 passed**, 44 deliberately unselected existing tests.
- Omitted-check mutations: two foreign assertions fail, plus a separate fixed-head timeout assertion fails because no error is returned. These are named assertion failures rather than timeout-only evidence.
- Mutable-control mutation: one catch-up assertion fails.
- In-entry assembly mutation: all three eligible-prefix cases fail through actual owner reentry protection.
- Corrected typecheck/build, format/lint/boundaries and existing Read evidence verification have recorded successful exits. The original exactOptionalPropertyTypes failure and the mistyped follow-up pnpm task remain separately recorded as failures.

I did not rerun tests or run fixtures concurrently. These execution counts are inspected author evidence, not independent execution claims. Independent checks here recomputed the three source hashes against source.json and all four emitted-module hashes against emitted-summary.json; all match the clean frozen checkout/build.

## Emitted and installed boundaries

The retained emitted script imports actual private built modules without a loader hook, creates one disposable owner and executes exactly four calls. Its observed totals are **4 entries, 1 synthetic wait, 2 returned candidates and 2 exact error-class outcomes**, with no listeners or HTTP requests. Cleanup records show the process exited and its temporary store was removed.

The approved clarification is accurately maintained: immutable ESM assembly was not instrumented in the emitted probe. Exact assembly call/no-call assertions are separate source-spy evidence; they are not combined into an emitted assembly count. The direct probe is not package/public-API/HTTP discovery qualification.

The existing source HTTP test continues to cover minimum-bearing discovery refusal; the unchanged emitted runtime regression covers ordinary authenticated refusals. The separate packed12/1 boundary scenario was not rerun or credited with exercising the new accessor. No new-path installed wire evidence is claimed, as the service accessor intentionally has no HTTP caller.

## Disposition and review limits

Clean native code review for this bounded internal increment. Root still owns external council, CI, ordinary integration and any later activation decision. Missing-Resource callback work remains independent. This report grants no public profile or whole-contract conformance claim.

Review was read-only: complete three-file diff, accepted plan/root acceptance, final author report, directly relevant service/owner/consistency paths, structured test/mutation outputs, emitted script and source/build identity checks. No tests/builds/network/source edits or publication were performed. Only this external review report was written.
