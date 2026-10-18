# Native closure review — artifact-set failure retention

## Finding

**P2 — preserve the original cancellation alongside a returned-evidence I/O failure.** At `apps/bdptest/src/read-update-matrix.test.ts:855–860`, `prior` selects `execution.error` whenever `observe()` rejected, even when the shared signal is already aborted. A write failure in the new `lifecycle.returned` callback (`:830`, wired at `:912–914`) makes `observe()` reject after the real runner has returned its aborted result. The original `signal.reason` is then omitted from the aggregate error, and `runShard`'s persistence wrapper (`:924–934`) only nests these replacement errors. This contradicts the repair's explicit original-failure preservation guarantee.

Independent actual-response probe: complete the small baseline, then abort continuation after its first real client response and make the evidence directory read-only using POSIX permissions. The writer's returned snapshot, cleanup metadata and failed-attempt writes raise EACCES. Recursive inspection of the thrown error/cause/AggregateError graph finds three I/O errors and wrappers, **but not the original AbortSignal.reason**. The completed baseline file remains byte-identical, no success summary exists, and every acquired child is exited. Thus this is a cause-preservation defect, not recurrence of the original report-loss path, false qualification, or a child leak.

Minimal fix: when cancellation and an observation/persistence error coexist, include the original signal reason along with the execution and cleanup/persistence errors, preserving both rather than choosing one. Do not serialize their raw messages into safe failure evidence or weaken any failure status. Add the bounded combined abort-plus-I/O regression; the existing separate abort/cleanup and atomic-replacement controls do not exercise this combination.

## Closure of the original P2

**The original child-report retention finding is closed.** `run()` now delivers the authentic runner result to the writer before final cleanup and before `runShard` checks cancellation. The writer detaches that result, stores it without manufacturing a new ConformanceRunResult, and independently updates cleanup state. A failed or aborted continuation retains the completed baseline; a returned failed/partial continuation remains available. Startup/no-result paths carry safe failed/unobserved status without invented reports. Association and successful summary remain behind all old assertions, cleanup and claim=false gates.

## Source and requirements

Frozen source **b569524dfe1826b0b43ac77afc1accf362d906db**, direct parent **7d256e57114b2ddc920eb01388381e86fa2b8285**, author checkout `/Users/dbox/repos/wt/vickie/vickie-artifact-set-failure-retention`. Read author report `artifact-set-failure-retention/report.md`, prior native P2 and accepted step-two failure-retention contract. Scope is the three-file repair only; no external 7d follow-up findings are adjudicated here. Remote session guard passed. Original source remained immutable and clean; probes/tests used `/Users/dbox/repos/wt/vickie/vickie-artifact-set-failure-retention-native`.

## Verified properties and limits

- Safe wrappers retain explicit known fields and genuine results/observations, exclude arbitrary wrapper exception/stderr/resolver fields, and check actual configured credentials in literal strings, object keys and the supported base64/percent spellings. A credential-bearing genuine result is withheld as a whole, not sanitized into purportedly authentic evidence. This is not an arbitrary encoding or arbitrary-secret detector claim.
- Child snapshots are detached before later caller mutation. Cleanup is pending before final close, then complete or failed; the result itself is unchanged. Returned evidence may truthfully exist while the overall attempt fails.
- Atomic temporary-file replacement preserves previous complete bytes if an update fails. Fresh output paths prevent earlier successful summaries from being mistaken for current attempts. POSIX permission-denial countercontrol ran on this non-root Mac; no power-loss durability claim is made.
- The ordinary error-plus-cleanup-fault branch retains both causes; the new finding concerns the narrower cancellation-plus-returned-persistence-error combination. No raw I/O exception is written to the safe attempt ledger by this code.
- Shared 90-second cooperative budget, 120-second outer bound, 5-second exchange and 30-second focused limits are unchanged. The timer now sits inside a cleanup-protected execution path after set/writer initialization. Bootstrap launch remains bounded/awaited, with acquired children closed after cancellation. No abandoning Promise.race was added.
- Manifests, fixtures, coverage, catalogs, source-only admission guards and runtime are outside this three-file diff. No new whole flag or public admission. Source/configuration association from the predecessor is preserved; this review does not repeat the whole segmentation audit or grant final readiness.

## Independent execution

Pinned Node 24.16.0 / pnpm 11.20.0. Clean disposable source:

1. Frozen install, pre-build typecheck, build, boundaries, format, lint and unchanged Read seal: all exit **0**. Six conformance files plus association/writer unit file: **446 passed / 7 files, zero failures or skips**. Exact commands and exits: `artifact-set-failure-retention-native/run.py`, `results.json`, `regressions.json`.
2. Actual focused retention controls plus private persistence controls: **12 passed / 2 files, 171 unrelated tests unselected**, exit **0**. Filter `private attempt persistence|retains actual artifact|retains actual returned` over `apps/bdptest/src/read-update-artifact-execution.test.ts` and `read-update-matrix.test.ts`, `--maxWorkers=1 --reporter=verbose --reporter=json`. This includes real startup/abort/assertion/cleanup discriminators, credential withholding, and atomic replacement failure. Results: `retention-focused.json`, `retention-focused.log`.
3. Independent combined-failure probe: **one expected red test / 154 unrelated tests unselected**, actual exit **1**. It injects cancellation and filesystem permission denial through existing test seams after a real continuation response. No runtime substitution. Probe code/log/results: `abort-io-probe.ts`, `abort-io-probe.log`, `abort-io-probe.json`. Logged facts: `includesOriginal:false`, `preservedBaseline:true`, `successExists:false`, `childrenDrained:true`. The failed probe is retained as failure evidence, not qualification.

All artifacts above are under `bdp-integration-logs/artifact-set-failure-retention-native/`. The disposable test file was restored exactly; both author and review checkouts are clean, no owned child remains. No full suite/fresh packed run was repeated; author's separate packed 218-pass result is read as author evidence only. No outward writes or source commits occurred.

**Disposition:** original P2 closed; repair the one narrow original-cause preservation issue above before claiming complete failure-retention closure. This report is not final step-two readiness.
