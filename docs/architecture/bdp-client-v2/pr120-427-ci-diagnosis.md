# PR120 427cfef CI diagnosis

**The actionable failure is the outer 15-minute job deadline.** The retained check-run annotation explicitly says “The job has exceeded the maximum execution time of 15m0s.” There is no observed test failure before cancellation. This is incomplete, unsuccessful CI, not a passing suite or proof the original W1b failure is fixed. I recommend revising and reviewing the existing bounded CI-headroom proposal before another run, rather than spending another unchanged run solely hoping for a faster execution episode. No source or per-test repair is justified by this attempt's evidence.

## Exact source and result

PR120 run38620066822/job116044976221 reports head `427cfef6decbb422b393024e36218affc26007fa`. Local checkout is clean at tree `9075897867d88344a8477e1f4a8e1a00a50f1228`. The workflow actually checked out PR merge `088fec9496fc11ee00ac4ee19b5757fe2daf7d4d`, merging that head into main `7e46cdc93b2fd02094ba7e15e57e743d2ef81daa` (raw log419–422). Installed sourceCommit fields truthfully name that merge checkout; they are not standalone427 attestations. The one-file427 delta from70ac remains diagnostic messages only.

Job metadata: start19:00:56Z, tests19:04:27Z, cancellation19:16:10Z, job complete19:16:15Z. All setup/static/build steps succeeded. Test step was cancelled; e2e:ready:preflight, smoke and Read-cohort evidence gate were skipped. Fifteen minutes is configured at `.github/workflows/ci.yml:19`; cancellation/cleanup delivery after the deadline does not grant usable extra budget.

From the full ANSI-stripped log I count **9255 completed test pass rows**, one skip, zero failed rows and no terminal Vitest summary. Counting rule: lines containing ` ✓ (apps|packages|scripts)/... > `; this excludes the pnpm supply-chain check and includes multiline test-title starts. A strict duration-at-end parser finds9244 because11 titles continue on later lines. This parser also reproduces the earlier70ac official9255-pass total. These are reported completed rows, not an official final summary or proof all172 files completed. The existing skipped demo-read case appears at10367. The expected unchanged total from70ac is9257 cases; one final Transactional case lacks a completion row.

Concrete observations:

- W1b link-metadata passed2630ms at19:07:15.475Z, raw4857. Its three unchanged assertions passed; diagnostic text therefore was not emitted. There is no new failing report from which to diagnose70ac's counter0. Local diagnostic36 exchanges and the tiny successor's two selected local passes remain separate evidence.
- Existing diagnostic projection/secrecy unit passed3ms at19:09:06.570Z, raw6311. This does not independently exercise every failure rendering in CI.
- Repaired service finite-budget test passed19ms, raw2640.
- Final reported pass was Transactional deletion outgoing-order,12397ms,19:15:58.796Z, raw11004; cancellation follows at11005,19:16:10.822Z. Incoming-order immediately before it passed12411ms. No failure trace, assertion failure, individual-test timeout, or full-suite summary appears. The final committed-prefix-state case has no completion row; neither its success nor failure may be inferred.

## Timing and comparison

| Run/source | Setup before tests | Test step / suite | Same56 completed matrix cases |
| --- | ---: | --- | ---: |
| Current427 |211s, including168s Go install |703s until cancelled; suite incomplete |658.206s |
|70ac first CI |Raw tests begin18:31:19.691Z |485.16s Vitest; real W1b failure |451.450s |
|30035 first CI |Raw tests begin18:05:58.717Z |425.44s Vitest; real service assertion/cleanup failure |394.156s |
|PR119 c555 successful retry |156s, including123s Go install |410s step /409.53s Vitest; all final gates complete |382.556s |

These are same-name completed matrix cases. Matrix source, workflow, Vitest config and package.json are byte-identical between c555 and427, verified by an empty Git diff. Current matrix has56/57 completions, compared with57 in each earlier run. Current summed56 time is1.458×70ac and1.721×green c555; median individual nonzero-duration ratios are1.486 and1.739 respectively. These sums describe sequential work within that file, not extra elapsed time to add to the suite. Full-suite aggregate test durations likewise overlap and must not be added to wall time. W1b itself is2630ms now versus1435ms failed70ac,1604ms passed30035 and1719ms passedc555; different failure positions make the failed1435ms unsuitable as equivalent-work throughput evidence.

Setup211s plus already-consumed703s test-step time is914s, with tests unfinished and all downstream gates still required. The nominal900s allowance leaves only689s for tests and every downstream gate after that setup. PR119's measured downstream step-end→job-complete cost was12s; an explicit conservative30s final-gates/cleanup allowance would leave659s for tests. These are accounting bounds, not a prediction that current cancellation occurred at exactly900s.

The logs show continuing completed work until cancellation and a slower unchanged matrix; they do not establish CPU, I/O, host contention, or a particular runner defect as the cause. No new work was added to the long matrix by427. A hang is not supported by the available progress, but interrupted final-case cleanup and outcome are unobserved.

## Safe next action

`transaction-ci-headroom-plan.md` already proposes one workflow-only literal change,15→20 minutes. `transaction-ci-headroom-held-disposition.md` says it is unimplemented and not preapproved: required Claude corrections1–3 must be incorporated before future use. Preserve that requirement. Update its evidence to include the identical-matrix successful baseline, the present slower episode, and both normal/slow partition regimes. Do not attribute most of the time increase to newly added work or treat an old cancellation as a completed suite.

A1200s outer allowance with current211s setup and proposed30s downstream reserve leaves959s for the suite:256s above this attempt's703s already consumed. Using the historical worst231s setup from the held plan leaves939s,236s above the consumed portion. Neither is a guarantee because this suite did not finish. For scale only, applying the observed56-case ratio1.458 to70ac's485.16s suite gives707.4s; with231s setup and30s reserve,968.4s total and231.6s remaining under1200. This is a cross-run estimate, not an observed full run, and the ratio is from the matrix rather than an entire-suite timing guarantee. The successful c555 run independently demonstrates a much faster unchanged matrix regime.

Thus a separately reviewed, bounded outer-headroom amendment is evidence-supported and preferable now to another unmodified lottery against observed timing variance. It must retain every command, full test inventory, existing test/HTTP/child/cohort deadlines, preflight, smoke, Read gate and required-check identity. No partition, skipped rows, caching/build provenance change or broad workload expansion is needed for that smallest proposal. Root decides acceptance; this audit authorizes no mutation or retry. Require a subsequent exact-head run to complete all gates. Stop and diagnose any actual failure, repeated cancellation, missing cleanup or changed inventory; no automatic timeout growth. An unchanged retry could pass, but would not resolve the demonstrated margin problem. Nothing here closes the unknown original70ac W1b cause.

## Evidence and audit limits

Inputs under the integration logs directory:

- Current full log SHA256 `752187cef60ddcf05d3149c7f67a9611962126ac58ca81d4a0be63924538c935`.
- Current status `0ab28254e8e414f2d0fbba8b1e909a7fa888e5fa8a7d6cf059b880da62340f56` and annotations `6e5d1267716d31073330ff299a82994f3b9523312b7655ca93bdef1d198fd159`.
-70ac full log `070cd1f21125d65b97d043488e34621328c0c2f2f4d44536eda4898c96225c31`.
-30035 full log `e6850dab246f2bd6708f0b2f31000041a1ff5bf47346fe67fff17e973c0c00a7`.
-PR119 successful retry log `40e297e7e94d2133cadc7f07142ac170b6860d9ef06fcf4bd6476d30eb86798b`, with transaction-cold-schema-ci-success.json metadata.

Read-only local Git/log analysis only. No tests, builds, services, install, network requests, retry, source/workflow edit or other outward action. This report is the only new artifact. No CI cleanup proof is inferred from runner orphan-cleanup text or cancelled job status.
