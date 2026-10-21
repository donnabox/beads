# Transactional discovery minimum — private service preparation

Supporting bounded implementation plan under the sole Vickie design plan. This extracts only the settled internal minimum-observation preparation from transaction-discovery-serving-plan-v2.md. Root and independent council approval are required before authoring. No source changes or tests were performed to produce this plan.

## Exact base, outcome and exclusions

Base: ordinary integrated `9b95df5dd866dbaf84150404c4db86ac8054b724`. The private accessor is in **packages/server/src/transaction-service.ts**, not service.ts. Its exact base SHA256 is `02b5278ab88a476bd98e0ae40d1f6cbfe7ed3f03208a59f57d133d0f75e341e3`. Inspection used Git objects for9b95 and the a53 callback-repair checkout; the production service is identical at these pins. The9b95 service test additionally includes the reviewed validation-limit literal, and transaction-discovery.ts includes the reviewed fixed-selector coherence/diagnostic projection. Preserve those integrated changes and the C3 callback receiver/getter-once tests.

Extend the existing private `discoveryObserved` accessor to accept an optional canonical minimum tuple and optional AbortSignal. A successful minimum-bearing call returns an immutable discovery candidate, directory and context from one eligible owner observation. A behind caller uses the existing finite policy; nothing is served or advertised. This is useful preparation for a future honest cold-client wire surface, not completion of that surface.

Exactly three repository files are in scope:

1. `packages/server/src/transaction-service.ts` — existing accessor and the existing consistency-check import.
2. `packages/server/src/transaction-service.test.ts` — direct real-owner minimum/lifecycle tests, preserving all existing C3 tests.
3. `docs/design/transaction-implementation-progress.md` — narrow internal-observation paragraph, distinguishing this preparation from unchanged wire refusal and remaining qualification.

No HTTP/runtime/store/parser/schema/protocol/index/export/fixture/manifest/catalog/evidence-file changes. No Type/incident/alias/current-missing Resource work. No route, principal, new option, public API, new timer, new barrier or alternative profile. No installed fixture backdoor or IPC accessor. Existing Scope, descriptor and directory wire discovery refusals remain executable; public admission/claims remain closed. Receiving malformed wire controls, retained surfaces and broader required profile gaps remain separate.

## Existing source seams and law

The internal accessor at transaction-service.ts237–254 already runs under `run`, captures exactly three context values in one `owner.withRead`, preserves entry-refused cause, and invokes `assembleTransactionDiscovery` after the owner releases. `run` at125–141 registers the whole promise in active work. `close` at528–546 rejects new work, drains admitted finite work, then closes Read cursors and the owner. There is no separate discovery lifetime to add.

`minimumReads` at107 is already a frozen copy of the existing finite policy. `readAtMinimum` at289–305 and fresh Snapshot handling at306–329 demonstrate reuse of `snapshotTransactionMinimum` and `waitForTransactionMinimum`. transaction-consistency.ts defines canonical tuple capture, exact epoch/view/position comparison and a one-deadline wait that retries only `TransactionMinimumFailure("behind")`; other failures escape. The default is1000ms, with at-most5ms scheduled sleeps. This is a policy bound, not a hard scheduler/latency guarantee.

`owner.withRead` at transaction-store.ts884–962 rejects reentry/closed/fenced entries and releases synchronous callbacks even when they throw. It does not permit an async callback. Its context is the actual readContext from that entry. No copying of Resource bodies or creation/expiry of a Snapshot is needed for discovery.

The existing strict-read/context/minimum law (bdp.md5758–5799) supports comparison against the actual observation; conditional behavior is outside this service-only change. No new HTTP precedence or error mapping is decided here. The accessor has no principal argument; it remains an internal fixed-policy service method. Future HTTP must authenticate before calling it and cannot infer authorization from these tests.

## Implementation contract

Use the existing signature location: `discoveryObserved(supplied?: TransactionMinimum, signal?: AbortSignal)`.

- **No supplied minimum:** preserve the current `run(async () => ...)` synchronous observation/assembly behavior and result/error shape. It performs one owner entry and parses only afterward. As with the existing no-minimum Snapshot path, this path does not observe the newly available signal and does not claim cancellation of synchronous work. Document and test that limitation rather than silently changing old behavior. Existing zero-argument callers are unchanged.
- **Supplied minimum:** snapshot/validate the tuple synchronously at service entry, before the first asynchronous retry. Caller mutation cannot retarget a pending call. Follow the existing Snapshot convention for synchronous malformed-argument rejection; do not invent precedence between malformed input and closing service. For valid input the existing run/check rejects new work when closing/closed.
- One registered run encompasses the complete finite wait **and** final assembly. Each attempt calls owner.withRead exactly once. Inside its synchronous callback, compare the captured tuple to that entry's actual context using `checkTransactionMinimum`, then copy/freeze the same three fields. Do not compare an earlier discovery result, call owner.inspect/readConfiguration for a second head, or carry a reader across await.
- Preserve the existing entry-refused Error and its cause. It is not behind and must not retry. Actual owner/provider failures propagate without being converted into a consistency failure or visibility gap.
- Pass the attempt to the existing `waitForTransactionMinimum` with captured minimumReads and supplied signal. Its returned successful observation is assembled only **after** the wait returns and the owner entry has released. Keeping assembly outside the retry callback also makes parser failures categorically nonretrying. An await between capture and assembly is acceptable: body and returned context describe the captured eligible prefix, even if a later independent commit occurs before delivery. Do not refresh only the body or context.
- Reuse the current assembler and deeply immutable parser outputs; preserve scope/operation map, replay floor and captured limits. There is no second builder or extra context parser. Return exactly the existing `{ discovery, directory, context }` shape.
- Cancellation with a minimum follows the existing helper: pre-abort rejects before any attempt; abort during a wait is observed at its next loop check, not by interrupting synchronous callbacks or an already-running sleep. Do not add a post-observation abort barrier. Close drains admitted waits rather than independently aborting them; caller abort can end its own wait. No fresh deadline on retry.

A small local synchronous observation helper may avoid duplicating the old capture/error handling inside the existing accessor. It must not broaden scope or change the no-minimum execution timing. No changes to the consistency helper are authorized.

## Meaningful tests and controls

Extend the existing “internal discovery owner observation” suite and reuse the real fixture, deferred promises, assembly spy and existing minimumReads test seam. Count owner entries after service construction so constructor validation is not mistaken for an attempt. No elapsed-time races or sleeps are needed for direct tests.

| Case | Exact evidence required |
|---|---|
| No minimum compatibility | Existing candidate/context/directory exactness, deep immutability, single entry and parse-after-release tests remain green. A pre-aborted signal with undefined minimum still returns the ordinary observation; a throwing minimumReads.now spy is untouched. This is a documented synchronous limitation, not newly cancellable behavior. |
| Satisfied equal/earlier | Fresh pos-0 and actual post-commit pos-1 observations. Equal pos-1 and earlier recognized pos-0/genesis controls return exact current pos-1 and matching descriptor head/epoch/view. One entry, one parse, zero waits. Context/directory/discovery remain frozen. Use actual epoch/view, not guessed opaque tokens. |
| Catch-up and immutable controls | At head1 start a bound2 observation, pause in injected wait, mutate caller tuple fields, and complete an independent service.submit to head2 while waiting. Release; exact returned context and descriptor head are2 under the originally captured epoch/view, with exactly two attempts and one wait. An independently admitted satisfied call can complete while the first waits. The successful commit plus parse-time owner access prove no owner entry spans the wait or assembly. |
| Final-prefix coherence | Use an assembly spy after owner release to establish a later independent owner observation/commit if needed without changing the captured context; at minimum assert the assembler receives exactly the successful attempt's frozen context and output uses that context, never the first failed attempt. Do not require response head to equal a newer separately observed head. |
| Foreign epoch/view | Well-shaped foreign tokens fail with exact TransactionMinimumFailure reason cursor-expired/foreign-view and the actual current context. Exactly one entry, zero waits, zero assembly. Do not assert HTTP statuses here. |
| Fixed-head timeout | Fake monotonic clock plus bounded deterministic waits leaves head fixed. Assert exact catch-up-timeout and last observed context, measured attempt/wait counts from that clock schedule, zero assembly and no allocation/state change. The unchanged helper owns the deadline; test does not fabricate timeout by catching an unwaited behind error. |
| Pre-abort and pending abort | Pre-aborted minimum-bearing call performs zero entries/parses/waits and yields existing ScopeServerOperationAbortedError. For pending abort, pause after the first behind attempt, abort, then release the wait; no second owner attempt or parse occurs. Do not claim instant cancellation inside a wait. |
| Drain/close | Start a deterministic behind call, then close. New valid calls throw closing/closed; close remains pending while the admitted wait is held. Release to timeout or caller abort; assert original result, closed promise completion and owner closed only afterward. Retain idempotent close behavior. |
| Reentry/closed/fenced entry | Existing real-owner reentry and closed-owner controls remain. Repeat minimum-bearing refusal: exact existing wrapper error/cause, zero waits/parses. A delegating owner returning the existing entry-refused shape may test cause identity for a fenced-style refusal, labelled delegated fault—not a real disk-failure/recovery experiment. No new storage fault hook. |
| Invalid input/provider/parser | Malformed tuple is rejected before owner entry. Actual/delegated provider error preserves identity and is not retried. Existing assembly/parser failure remains distinct; for a satisfied minimum it occurs after one released entry, with zero waits, and a later valid call/close remains usable. Test-induced parser failure does not poison the real owner. |

Add only bounded sensitivity controls that distinguish the new obligations. In disposable source copies outside the frozen author checkout: (1) omit the in-entry check → behind and foreign tests fail; (2) reuse mutable supplied tuple instead of its snapshot → paused-mutation catch-up test fails; (3) assemble inside withRead → existing real-owner parse-after-release oracle fails. Record named assertion/red exit and restore unmodified green source. No mutant installation, new production hook, or broad full-suite repetition. If a mutant hangs, enforce a short outer test-run bound and clean it; a timeout alone is weaker than the intended named assertion and must be reported as such.

## Source/emitted proof and installed limitation

The existing installed runtime returns scope/listenUrl/configurationMetadata/targets/close (transaction-runtime.ts350–378); it does not expose its service object. The actual HTTP route at transaction-http.ts533–549 refuses minimum-bearing discovery through the existing route gap and ordinary discovery through transactional-discovery. The12-request installed discovery fixture observes those refusals plus a real write/read/retry; it cannot invoke the new internal accessor.

Therefore **no new-path installed HTTP proof is claimed or required for this internal slice**. Do not add an endpoint, runtime return value, IPC command, index export or direct unexported-package import disguised as a cold-client observation. Later activation requires its own fresh wire evidence. Historical12/1 proof remains bound to its source; an unchanged emitted refusal regression establishes only continued refusal.

After typecheck/build, run the existing emitted transaction-installed-contracts test as the unchanged wire-boundary regression. For the new path, retain a small external review script beside the implementation report that imports the actual built private dist/transaction-service.js and dist/transaction-store.js by absolute file URL, creates a disposable own database with the existing literal service fixture configuration, and exercises: ordinary pos-0; satisfied pos-0; fixed-head future pos-1 timeout using a deterministic finite policy; pre-aborted minimum. Assert matching candidate/context on positives and exact error class/reason/context on negatives. Close the service and remove the database in finally. This is explicitly **direct emitted-module evidence**, not installed package/API/HTTP evidence. It adds no repository file or API. Use one Node process, zero listeners/HTTP requests, one temporary store, at most10 owner observation attempts, and a10-second process bound; all cases should complete without wall-clock waiting. Save script, source/tree, emitted module hashes, output/exit and cleanup proof. If the existing build/import layout prevents this, report the concrete issue rather than expanding the API.

No actual packed run is necessary merely to repeat the old refusal path. Root can request a separately labelled regression pack if integration evidence calls for it; it earns no evidence for the new accessor. Source tests plus direct emitted proof are the explicit completion evidence for this stage.

## Validation and done-when

Use pinned Node24.16.0/pnpm11.20.0 in a new isolated checkout from exact9b95 or a root-assigned ordinary successor. Before editing, record exact base/tree and reanchor named seams. Shared service.test.ts contains reviewed C3 limit and callback repairs; preserve them byte-for-byte outside added/necessary local assertions. Current compatible-missing work owns transaction-read.ts and separate tests; this plan neither depends on nor duplicates that implementation. Any ordinary integration conflict is reviewed explicitly.

Run typecheck, build before emitted tests, focused transaction-service.test.ts and transaction-consistency.test.ts (maxWorkers2), then unchanged transaction-discovery.test.ts and transaction-installed-contracts.test.ts as applicable regression coverage. Run format/lint/boundaries and the repository's required evidence verification without editing historical evidence. No automatic full-workspace test rerun. Keep actual commands/exits and identify source versus emitted results. Use observable test-run discipline and cleanup; all test databases are disposable, no shared database.

Done when:

1. Exactly the three approved files change; no executable HTTP/public/export/schema/catalog/fixture changes.
2. The new optional minimum path has same-entry comparison/capture, immutable controls, one finite deadline, release before waits/parsing, preserved causes and lifecycle/cancellation behavior proven by the stated tests/controls.
3. Zero-argument/no-minimum behavior and C3 captured limits/callback semantics remain intact; wire discovery is still refused.
4. Clean frozen successor, scoped green checks and named sensitivity reds are recorded. Direct emitted proof is labelled accurately with hashes and cleanup; no claim of installed discovery or qualification is made.
5. Native and external code review plus root disposition complete before any subsequent activation work.

Stop for a production file beyond transaction-service.ts, new runtime options, changed wire behavior, a new protocol choice, need to alter the shared consistency/store helper, or inability to retain lifecycle/context correctness inside this scope. Root owns outward publication/integration. This plan introduces no product decision; it stages existing semantics while cumulative wire-advertisement gates remain closed.
