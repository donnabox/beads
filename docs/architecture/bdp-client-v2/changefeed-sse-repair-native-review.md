# Native closure review — G2a repair

**Accepted F1–F6 repairs are closed in this bounded review. No actionable regression found.** Full profile qualification, fresh packaged evidence and G2b remain separate gates.

Frozen source: `b14fb2875447a2bfd9de9fa725add412ba28ad5c`, direct child of previously reviewed `b64d7e46652739c6436259a8c8ed4c5e8049a077`. Independent detached checkout: `/Users/dbox/repos/wt/vickie/vickie-changefeed-sse-repair-native-review`. Reviewed all nine changed paths against `changefeed-sse-root-dispositions.md` and the complete `changefeed-sse-claude-resumed-review.txt`. Remote guard passed first. No author/source branch, outward post or public claim was changed.

## Dispositions

| Finding | Closure and actual evidence |
| --- | --- |
| F1 timer bounds and relationship | `transaction-sse.ts:15–24,88` rejects nonpositive, noninteger, nonfinite and oversized Node timer values before pumping. `transaction-runtime-configuration.ts:294–303` validates unchanged installed defaults and derives live availability for short, unattestable transport test seams. `transaction-runtime.ts:166–175` refuses selected live delivery before registration/headers when the seam cannot sustain it; finite/HEAD remain usable. Focused tests pin invalid values and the strict 4999/5000/5001 boundary without the rejected factor-of-two rule. Existing service validation still bounds stream-count validity. No deadline widening. |
| F2 retained genuine fault | `transaction-runtime.ts:194–200` records the first unexpected post-header cause while the stream signal is still live, **before** response destruction triggers disconnect. `:339` includes it in the existing close AggregateError after cleanup. The real injected-store-failure test passed exact original-error identity and one status/no error Event. An additional native two-stream/two-distinct-fault probe passed: both connections close correctly, but shutdown contains exactly the first original cause. Retention is one bounded slot. Normal disconnect, service shutdown and pre-abort tests still close cleanly. This corrects the private diagnostic limitation noted in the prior native report. |
| F3 test dependency edge | Raw recorder relocation is byte-identical: SHA256 `09ff8343ac1a5dbad77eaf24fc2745cc28ea870ada6b05dc4ea1644c0ad5ef81`. Server tests use their own relative test-support path; the app uses existing source-only `@bdp/server/testing`. `testing.ts:429` re-exports that helper. Public package exports and production inventory are unchanged. Typecheck and boundary checks passed. |
| F4 pre-abort guard | `transaction-runtime.ts:204–210` explicitly checks the signal before generic finite500. The discriminating registration-race test injects abort between selection and `service.stream`, observes zero wire bytes and clean fixture shutdown. No new abort-class or wire error semantics were added. |
| F5 raw finite/SSE oracle | The source replay test now retains finite response text and compares each complete SSE data payload directly against it (`transaction-sse.test.ts:168–180`), while preserving structural/order/count checks. The fixture includes numeric-like property names. The app observation similarly compares its first SSE payload against actual finite raw body text (`transactional-sse-installed.test.ts:174–175`). Neither byte check reparses/reserializes its finite operand. Both tests passed. |
| F6 module prose | The runtime and configuration docblocks are restored before imports. Index's new helper import is moved into its existing import section; that file did not have the alleged leading module docblock to restore. No behavioral change. |

F7 remains deliberately deferred: selected SSE `start=now` plus Last-Event-ID retains the private gap, now documenting the automatic-reconnect consequence. No invented 4xx mapping, browser support claim, erasure semantics, or G2b work entered this repair.

## Independent execution

Pinned Node **24.16.0**, pnpm **11.20.0**, maxWorkers **2**.

- Frozen install and clean build: exit 0.
- Focused replay/stream/service/configuration/Read HTTP/narrated protocol/actual-default idle run: **324 tests / 7 files passed**, exit 0, **18.23 seconds**.
- Additional finite Event/runtime/Snapshot HTTP regressions: **39 tests / 3 files passed**, exit 0, **2.23 seconds**.
- Native two-fault retention probe: **1 passed**, 25 copied tests deliberately unselected, **0.763 seconds**. It injected separate real owner-read exceptions after each stream's first complete frame and asserted shutdown retained only the original first cause. Its disposable added test was removed; exact probe source/log retained outside the checkout.
- Five emitted-pump probes passed: drain advances the checkpoint once; no read-ahead or heartbeat accumulation while blocked; error/close preserve termination and remove listeners; pre-abort publishes nothing; abort during an outstanding read prevents publishing that result.
- Typecheck and dependency-boundary checks: exit 0.

Logs/scripts/JSON: `/Users/dbox/repos/wt/vickie/bdp-integration-logs/changefeed-sse-repair-native/`. `focused.json`, `finite.json`, `first-fault-probe.log`, `native-sse-repair-probe.test.ts`, `pump-probes.mjs`, and `pump-probes.json` preserve actual evidence. The focused 15-second test used this clean checkout's emitted child; it was not relabeled as a new installed-tarball cohort. Its workspace was clean during observation; the disposable native probe was added only after that run completed.

The final checkout is clean at exact b14; the temporary probe is removed, no owned Node processes remain, and cleanup/process snapshots are saved. No timeout, forced interruption, deadline widening or retry occurred. No replay, current-context/header, conditional/HEAD or finite-path regression was observed.

## Limits

This is a native repair review, not a repetition of the author's 5,406-server regression run or ongoing fresh tarball evidence. I did not run the full workspace suite or smokes. Five pump probes use minimal synthetic pages solely to discriminate control flow; real replay/partial-prefix/owned agreement is covered by the source tests. Erasure/View publication fences, Event Source SSE, remaining set/discovery work and whole-profile qualification remain explicit gaps. No production qualification or public claim promotion is made.
