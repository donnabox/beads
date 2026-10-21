# Comparison-timeout review repair: final bounded report

Frozen clean commit `70ac666002708202d2f93195eed90dfe8541300d`, tree `1353e8f4fb0b1149b3f4ff158335804bf3f1cecc`, ordinary parent `30035b88380b7cbb3abd5359787db236b70345c7`. Checkout `/Users/dbox/repos/wt/vickie/vickie-comparison-timeout-review-repair`, branch `codex/vickie-comparison-timeout-review-repair`. Four authorized test/doc paths,159 insertions/21deletions. Local commit passed the remote session guard and includes `Agent-Signature: Vickie (Codex)`. No worker push, PR change, main merge or coordination mutation.

## Changes and dispositions

F1 closed: the existing commit-uncertainty fencing case now explicitly closes the owner, then requires the same failure on direct deadline observation. A new existing-fixture fake-timer case triggers actual background maintenance failure at25ms; it verifies timer cancellation, the maintenance-failure exception and its original cause after the owner is closed. Production already had the correct two fault guards; this is permanent coverage, not a production fault fix.

F2 closed narrowly: a submission-only owner double returns owner-stopped at a reached deadline. The case checks strict contextless comparison-refusal, one wait, one admission and one final callback with actual Scope epoch; its finally cancels both pending members. No HTTP stub or supported live epoch transition is introduced.

F3 closed: reflow the existing progress paragraph. F4 informational, no source action. Existing18-test/five-control and27-request/two-child evidence remains historical30035 evidence.

CI amendment: original30035 full CI had a real stale observed-service expectation and then a10s afterEach timeout because the assertion skipped release of two queued turns. The immutable original log remains `../transaction-comparison-timeout-ci-first-full.log` (bound by preflight). The one authorized service test now expects strict comparison-refusal/deadline with the actual trusted owner epoch/view and independently expected pos-0. Original two comparison waits, empty pre-turn graph, retained completed retry/pos-2 and exactly two groups remain. Its finally releases at most the two admitted callbacks, drains the captured sequence outcome and closes the service; primary and cleanup failures are retained together. Fixed-field cleanup JSON provides a direct positive and failure-path witness. No timeout increase or runtime change.

Read-only search is retained in `stale-shape-search.txt`. No other observed-service legacy-shape expectation was identified. Unobserved submission/dependency/alias assertions remain correct; unrelated Snapshot/Event/read/client unavailable assertions were not edited. This search is not a full-suite result.

## Draft lineage and actual failure

The original three-file draft/preflight and native closure are preserved in `initial-preflight/`. Root added the service file after the real CI failure. Accepted four-file preflight SHA `a3f89340417f2aeff2f07c2d9374d8fc5300e7437de5f5f2af7143fb968d3639`, bounds `711b47035b179faf83ef8a14631aa871478bd505e9746186e49b4a5da96ef9df`, diff `17eab65f6c95fa6e49f848de23dd7d342682c8b1dfaadb39c7f3bd402aa913a4`; native closure `09b1d054ca9a65253609aa07cec4cb7fa6ba6bb319106a7a102a577faf4e6bc3` and `root-measured-acceptance.md` released execution.

Initial four positives and three intended failures completed. Whole lint then exited1 with one hidden error; an owned-three-file diagnostic check identified `noUnsafeFinally` on an immediately caught throw inside the service finally. Both failed lint logs/exits are retained (`lint.*`, `lint-diagnosis.*`). Root authorized exactly one equivalent line correction: append the already captured outcome error directly to failures, instead of throwing and catching it. Pre-correction source/diff/measurements are preserved in `prelint-correction/`; `lint-correction.json` records source/mutant hashes and authorized affected repeats. No runtime, assertion, deadline or semantic change occurred.

The service positive and its old-shape derivative were rerun on corrected bytes; unchanged store/submission evidence was not repeated or relabeled as a fresh whole-source run. Total actual case executions **nine**: five positive executions covering four distinct cases, four expected-failure executions covering three distinct controls. This includes the approved two affected repeats beyond the original seven. Total temporary stores9, in-process service instances4, HTTP requests0, listeners0, installed children0. No automatic retry or broad suite.

## Verification

Pinned Node24.16.0/pnpm11.20.0, maxWorkers1,120s command supervisors, unchanged inner test deadlines. Exact argv/cwd/start/end/exit records are `*.exit.json`; source results and raw logs remain adjacent.

| Run | Outcome | Filtered | Attribution |
| --- | --- | ---: | --- |
| positive-1 | 2 store cases passed |138| Accepted draft; file bytes equal final commit |
| positive-2 | 1 submission case passed |36| Accepted draft; file bytes equal final commit |
| positive-3 | 1 service case passed |64| Retained pre-lint-correction draft |
| service-positive-final | 1 service case passed |64| Corrected file bytes committed unchanged |

Initial positives ran18:20:03–18:20:07Z. Final corrected service positive ran18:22:31–18:22:32Z. No failed positive cases or hook timeouts occurred locally. This does not erase the earlier full CI failure.

The closed-guard derivative changes only `if (closed && !ownerFailure && !maintenanceFailure)` to `if (closed)`. Each selected F1 case built successfully and exited1 at its own intended assertion: `closed fenced deadline preserves failure` and `closed maintenance deadline preserves failure`. Actual stacks/adjudication are retained under `controls/closed-guard/`. Both failed assertions directly distinguish the benign stopped result from the required exception.

The service derivative changes only the expected observed shape back to legacy refused. Initial and corrected variants each built and failed at `observed service comparison deadline shape`. The intended inner assertion is retained, not merely a generic AggregateError label. Actual cleanup JSON is exactly present once for each run: positive has drained:true,closed:true,cleanupTurns:0,pendingCallbacks:0; derivative has drained:true,closed:true,cleanupTurns:2,pendingCallbacks:0. Neither derivative timed out its hook. Final records: `service-positive-final-cleanup.json` and `control-service-final-adjudication.json`; earlier records remain separate. The final derivative ran18:22:46–18:22:47Z.

Final offline setup, typecheck, build, format, lint, dependency boundaries, historical Read evidence verification and diff whitespace check passed. Final lint retains666 warnings/43infos; it is not warning-free. Controls' nonzero exits are expected sensitivity evidence, not ordinary passing tests. The full workspace suite was not rerun and no new CI claim is made.

## Source and package preservation

`source-preservation-final.json` independently byte-compares **659 other tracked files** against30035. All match, including production modules, package inputs, installed scenarios, protocol/schema/runtime/configuration and public surfaces. Only the four accepted test/doc paths differ. All selected store/submission bytes match accepted preflight; final service differs solely by the separately authorized lint correction.

`package-preservation.json` compares the rebuilt local outputs read-only with the actual original30035 archives. Protocol62/63 regular-file members and server249/250 members match byte-for-byte, including root LICENSE, all emitted files/maps/declarations and schemas. The remaining package.json per archive is compared decoded: protocol equals directly; server equals after the existing pack-time workspace:*→0.0.0 dependency substitution. Package source inputs are unchanged. The initial simple package-path comparison and its LICENSE/manifest differences are preserved in `package-preservation-prelint-repair.json`; the final report explicitly resolves root LICENSE location and metadata normalization.

Original archives: protocol `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227`, server `b9ebfae7b6fd2f5602688fadf0110da3749d5af012fdf4599867c6af2f84bf35`. No new pack/install/installed child execution was performed for this successor. The old27/two installed regression remains bound to30035 and payload2c62c667; this is rebuilt-file preservation, not a new tar-byte identity or installed execution claim.

## Cleanup and bounds

`cleanup.json` records actual service cleanup objects and independently checks every recorded command PID absent. No owned test/build/service process remains. Both derivative source edits were restored after execution. Existing fixture teardown completed and removed disposable directories; individual directory paths were not externally retained for a second independent filesystem check. Fake timers are restored in finally, the maintenance timer is asserted cancelled, and the fenced spy is restored before close. Derivative checkout/build artifacts remain review material, not active processes.

Default installed timeout sensitivity, live epoch rotation, earlier maintenance-failure comparisonEpoch shortcut and public profile qualification remain outside scope. `claimEligible:false`; no schema or protocol law change. Root owns final native closure, publication and new repaired-head CI.

Final artifacts:

- source.json: `68524854724fb5b0da026be9c5dec32ea0e722698b621d27ffc1a8450c73e58b`
- summary.json: `cf1cfd25bb95b97983d08f070236f13a287a812ad843d32fdc13d993ac756cf9`
- cleanup.json: `f9187884b3fcb0fbd8335fd461d2a1d5b194fc6cd8a6158464731ee3e26dff2d`
- package-preservation.json: `b8a1abff6a0534214a165e3124089557abad2d217957d92163569399672552de`
- source-preservation-final.json: `351c4604c218bccb73eaa17e69e5c5684c84d85b077be66781a895fd192c8132`
