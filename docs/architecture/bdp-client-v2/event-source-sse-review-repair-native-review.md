# Native closure review — Event Source SSE council repairs

## Findings and disposition

**No actionable findings. F1–F3 are closed by the frozen five-file repair; F4 retains the explicitly accepted structural evidence limit.** This is focused repair closure, not new profile qualification or approval of unfinished set/Selector work. Root still owns exact-head CI, fresh packed evidence acceptance and PR readiness.

Exact target: `310871d49c4bb9adb1177a318e06fde29b29cf96`, direct parent `6a13489d0de673fd59c50b6be1d65e93f3b4be6a`. Author checkout `/Users/dbox/repos/wt/vickie/vickie-event-source-review-repair` was read only and clean. The entire five-file, 129-addition/2-deletion delta was inspected against the root dispositions, original Claude report and earlier native review.

- **F1: actual structural bound.** `packages/server/src/transaction-store.ts:71–75` fails module initialization if the shared group-step constant differs from the algorithm's one-group structure. This preserves the existing one-group algorithm instead of adding an unnecessary multi-group scanner. `transaction-events.test.ts:826–839` now derives per-entry body-read and total observation expectations from the shared value. The new mismatched-module test at 971–990 checks refusal when the constant is2. The author's retained `group-guard-before.log` shows that exact test resolving unexpectedly before the guard and failing at the rejection assertion. The repaired test passes independently.
- **F2: shared cadence binding.** The configuration comment names `eventStream` as owner scan/packet bounds and `changefeedStream` as shared cadence/capacity. The new test at `transaction-event-sse.test.ts:881–912` injects a distinguishable frozen captured settings object and observes that exact object arriving at the real Event pump. It does not infer binding from elapsed wall-clock comments. I independently changed only disposable emitted code to pass literal default cadence rather than captured settings: the named test failed at line909, showing17/31 expected versus100/5000 observed. Restoring emitted bytes made it pass. This closes propagation coverage without claiming a different production timer configuration or independently testing GC.
- **F3: actual valid-cursor suffix.** Both new GET cases at `transaction-event-sse.test.ts:847–879` use valid differently positioned query/header cursors. A subsequently committed marker bounds observation; the assertions check complete subject order, prefix Event IDs and counts through that marker. This avoids an empty-suffix claim based only on HEAD or a short sleep. The retained `cursor-precedence-inverted.log` shows both directions fail at line868 when query precedence is inverted. Both repaired cases pass independently.
- **F4: structural reference release.** The repair documentation accurately limits evidence to the inspected reference chains: owner tokens contain no group body; runtime clears `selected` after transfer; synchronous adapters return the pump Promise; the pump clears its initial argument and delivered packet before subsequent waits/pulls. Those production assignments are unchanged by this delta. No test here establishes garbage collection, a release deadline, maximum heap usage, or absence of all possible engine retention. The explicit limit is consistent with the root disposition; no WeakRef/forced-GC requirement has been added.

The only production behavior change is the guard against a future mismatched structural constant. Current constant1, packet limits, route selection, replay, finite paging, cancellation, shared capacity and timer values remain unchanged. No Selector, public activation, schema, catalog or evidence source was edited.

## Independent checks

Disposable detached checkout: `/Users/dbox/repos/wt/vickie/vickie-event-source-review-repair-native`, exact310871d. Logs and probe script: `/Users/dbox/repos/wt/vickie/bdp-integration-logs/event-source-sse-review-repair-native/`.

Pinned command prefix: `PATH=/Users/dbox/.nvm/versions/node/v24.16.0/bin:$PATH /Users/dbox/.nvm/versions/node/v24.16.0/bin/node /Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs`.

- Session guard: exit0.
- `install --frozen-lockfile`: exit0, `install.log`.
- `build`: exit0, `build.log`.
- `exec vitest run packages/server/src/transaction-event-sse.test.ts packages/server/src/transaction-events.test.ts packages/server/src/transaction-runtime-configuration.test.ts packages/server/src/transaction-sse.test.ts --maxWorkers=2 --reporter=verbose`: exit0, **98 tests /4 files**, 3.76s, `focused.log`.
- `git diff --check HEAD^ HEAD`: exit0.
- Independent generated-code cadence counterfactual: `cadence-probe.py`; targeted test exit1 as expected (`cadence-ignored-mutant.log`), byte-restored targeted test exit0 (`cadence-restored.log`). `cadence-probe.json` records the exact command and restoration SHA-256 `6ee5e5fb25b3d7b5bda875c773c5c488fca313c9a2309cdacce8f7c7c3878ba2`. No tracked source file was changed.

All commands completed; no review-owned test/Node children remain, no forced cleanup was needed, and the isolated tracked tree is clean. The original author tree was not touched.

The author's repair report was not yet present on the initial read, and its packed run was still active. This review therefore makes no claim to have validated completion of that run or exact-head CI. It does not repeat the full server/workspace/packed cohorts. Source closure is supported by the complete bounded diff, retained red evidence and the independent checks above; final integration/evidence gates remain separate.
