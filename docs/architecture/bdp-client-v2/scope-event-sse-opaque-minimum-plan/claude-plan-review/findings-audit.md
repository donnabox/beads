# Independent native audit of Scope Event SSE Claude plan findings

Verdict: root's corrected dispositions close the concrete findings without a new product decision. This is findings disposition, not approval of an uninspected successor plan or authored code. The original reviewed plan has real F1/F2 blockers; they must be corrected before authoring. Root authorized the seventh, test-only file. No product or validation execution occurred in this audit.

## Concrete dispositions

- **F1 accept root alternative, retain seed.** transaction-batch.ts:771–804 emits three created Events and two linked endpoint Events for a/b/l; the fixture Link type is not owning. Scope projection includes all five. receiptMaximumItems2 therefore means exactly three finite pages 2+2+1, two actual continuation requests, and five positive complete SSE frames. D becomes31 requests,200x6; other statuses unchanged (503x10,500x8,400x3,404/401/304/412 each1). A's one-item packet equality remains unchanged; its prose must say five-Event group. Do not optimize the fixture to preserve mistaken counts.
- **F2 accept exact pair-only correction and select existing callback.** transaction-events.test.ts:1530 is the newly eligible canonical Scope SSE request. Change expected500 to503/catch-up-timeout/observedtrue. The existing helper checks exact GET Problem, empty HEAD, representation length/type parity, actual context, Retry-After1 and private,no-store. The generic 500 private-gap comment remains correct for excluded rows; there is no adjacent SSE-specific comment requiring replacement. Narrow seventh-file exception is sufficient. The callback has one seed POST and one finite GET, followed by15 GET/HEAD pairs:32 HTTP, **one commit, not Claude's proposed two**. Preserve dispatch order/count and its actual existing report fields calls/productChildren/directoryAbsent/failureCount. The report does not independently enumerate32 exchanges or literal media; do not invent those observations.
- **F3 accept src class identity.** A/B/C must use ./transaction-service.js and ./transaction-consistency.js with ownerFixture's ./transaction-store.js. The waiter catches instanceof identities, so a dist service plus src owner would skip the intended retry. D and the existing finite Scope callback continue to use emitted runtime; a fresh bound build remains prerequisite to collection.
- **F4 accept callback-local close-settlement flag.** service.closed is a promise, not a boolean. Attach observers to the actual close promise immediately and record local flag state. Preserve the explicit synchronous-instant limitation; it is not a scheduling-sensitive proof that early closure could never occur. Await and inspect actual completion during drain.
- **F5 accept C-only close counter.** C's spread delegate wraps beginEventScan and close, forwarding to real owner. B wraps only beginEventScan. Fixture cleanup closes the real owner directly and does not increment the delegate counter; count1 is coherent with service.close's memoization and actual owner.close call.
- **F6 accept canonical validation then relative dial.** eventContinuation creates an absolute canonical URL under the fixture origin; f.request accepts a scope-relative target. Validate each returned canonical next against the preceding page's actual final Event, then dial its scope-relative suffix against the actual listener. Two continuation requests are now required. Do not dial the nonresolvable canonical origin or silently skip canonical checks.
- **F7 accept citation corrections.** beginEvents starts539; beginEventScan1766–1808 excludes advanceEventScan1809; lexical receiving citation begins5819.
- **F8 accept only root's additional corrected wording.** progress:2060 belongs to **finite changefeed start-now**, so its SSE exclusion remains true for changefeed SSE. A new append documents canonical Scope Event SSE independently; it may reject a broader/misapplied all-SSE inference, but must not say the changefeed SSE limitation itself is superseded. Keep2060 and the unrelated receipt paragraph1880 byte-preserved and explicitly retain changefeed SSE/Bead/Link/reconnect/foreign exclusions. This corrects the review's imprecise attribution without changing production scope.
- **F9 defer optional observer expansion.** A row1 remains honestly false/null/unobserved; rows2–13 zero preparations and row14 positive remain the planned measured sensitivity. Preparations are not executions, rows scanned, or allocation census.
- **F10 defer optional regex cosmetic change.** Retain the known working no-u/v Vitest pattern; adding the exact fourth legacy title must still be verified against actual selected names during preparation/execution. No regex redesign is required by this finding.
- **F11 preserve historical pending-CI prose.** The separate exact-head acceptance updates operational basis; old preparation history must not be relabeled as already green at its creation.

## Independently reconciled prospective workload

| Dimension | Corrected source | Installed | Combined |
|---|---:|---:|---:|
| Selected callbacks |8 (4 new +4 legacy)|2|10|
| HTTP dispatches |108 =31+28+13+4+32|11 =5+6|119|
| Stores |8|2|10|
| Owner lifetimes |8|3|11|
| Listener lifetimes |5|not recast as source listeners|not asserted|

Source commits are9: A1+B2+C1+D1 plus four legacy callbacks each1. A–C subset remains18 delegated/live begin calls, one separate raw closed probe,7 service.beginEvents calls,2 waits,6 clock calls,4 commits,3 stores. Runtime owner work from D/legacy is outside these counters. There are four new source observations plus the unchanged finite-event-minimum-http observation; the three old SSE callbacks gain no invented reports. Installed remains three sequential children/two stores. Dispatch totals include finite and stream transports and are planned successful-path counts, not measured evidence or full profile qualification.

The three original legacy source counts were inspected directly:28=seed+finite+initialHEAD+20 pair requests+3 overrides+JSON+unauth;13=seed+2*(active+refused+4HEAD);4=seed+HEAD+waiting+positive. Added finite callback starts one runtime/listener/store and one committed seed. Source/default deadlines and installed45/5/10/60/120 limits need no change.

## Evidence and limits

Read full final-review and frozen context plan/selection/workload; inspected relevant source ranges listed above and the three legacy callbacks. Checked clean actual base HEAD. Conditional reviewer owns review completion/provenance audit; this report does not duplicate or claim its terminal-stream/process verification. No runtime, formatter, compiler, test, helper import, source mutation, or external publication. Root's latest F8 correction is included. Successor plan/index/selection must separately encode these decisions; this report does not pre-approve unseen bytes.

Actual inspected input hashes:
- `/Users/dbox/repos/wt/vickie/bdp-integration-logs/scope-event-sse-opaque-minimum-plan/claude-plan-review/final-review.txt` — `5a938579df557b03efb96aeb3c14835787b197e3889d4c5107f3bfb2c9caf7af`
- `/Users/dbox/repos/wt/vickie/bdp-integration-logs/scope-event-sse-opaque-minimum-plan/claude-plan-review/context/plan.md` — `bf2f84fe9e7e358b743cb5f363a53943bac12a00070258e3a480477cd4e0eb7b`
- `/Users/dbox/repos/wt/vickie/bdp-integration-logs/scope-event-sse-opaque-minimum-plan/claude-plan-review/context/selection.json` — `a95473c59618021d4309786fe335754484001ea361a447ef43f680ee85b3445e`
- `/Users/dbox/repos/wt/vickie/bdp-integration-logs/scope-event-sse-opaque-minimum-plan/claude-plan-review/context/workload.json` — `1968c6409adda8abb4062728d7df2203a22925fd86e48ebea6488e178d4a54f5`
- `/Users/dbox/repos/wt/vickie/vickie-changefeed-opaque-minimum/packages/server/src/transaction-batch.ts` — `830ad57a680db7148fc75cd9cc10cadfc6a812448fd1704c1b0f47b13611808d`
- `/Users/dbox/repos/wt/vickie/vickie-changefeed-opaque-minimum/packages/server/src/transaction-event-sse.test.ts` — `de14cbbbe935205c5ec43ef6cf6a055d3f1ca2b9632ee684c4c13fbfe9eed042`
- `/Users/dbox/repos/wt/vickie/vickie-changefeed-opaque-minimum/packages/server/src/transaction-events.test.ts` — `4f439f1c7b02dbd1e5a65fcc675d24386cf1901a9e3f7ba1819cfc1d47317555`
- `/Users/dbox/repos/wt/vickie/vickie-changefeed-opaque-minimum/packages/server/src/transaction-events.ts` — `6c5ab4482f0b00110a71a96b16d44ca5fe713cb5f397ff78778cc6646f632a3a`
- `/Users/dbox/repos/wt/vickie/vickie-changefeed-opaque-minimum/packages/server/src/transaction-consistency.ts` — `ad51311d6098db32fd47849e0b78d3118162cdf1e80ee9ef4d46324bfa9f436f`
- `/Users/dbox/repos/wt/vickie/vickie-changefeed-opaque-minimum/packages/server/src/transaction-service.ts` — `759dcf5bb5d2f84a59dd5c4daa7877d470a3fbe9b0776678326a99baa2bc48b4`
- `/Users/dbox/repos/wt/vickie/vickie-changefeed-opaque-minimum/packages/server/src/transaction-store.ts` — `54605c6f0cd26ce4e9ebf6ee4b9dab8c556aa52fc7f8e602324809c86dd2c6fd`
- `/Users/dbox/repos/wt/vickie/vickie-changefeed-opaque-minimum/docs/design/transaction-implementation-progress.md` — `c12343963d0f3b3500f67c66fe55279111b3441f2ee33d2ee25ba4842c4522e6`
