# Native closure review — Snapshot aggregate deadline

**ACCEPT: prior P2 closed. No new actionable finding in this two-file repair.**

Frozen source `ffe188a3405881e9488b4e7c997f557d39829fc5`, direct parent `6413cbc2842d898e523869d56c8286e87f358804`, author checkout `/Users/dbox/repos/wt/vickie/vickie-snapshot-deadline-repair`. Scope only `apps/bdptest/src/transactional-snapshot-installed.test.ts` and its dedicated observations doc. Previously accepted serving and deferred P3 failed-attempt diagnostics were not reopened.

## Closure

The capacity test now owns one `withDeadline` around both calibration and observation. Its default remains45 seconds. Both `withInstalled` calls receive that exact signal; neither starts another timer. An already-aborted signal is checked before a fixture starts and again before provisioning. The same signal reaches the installed launcher, every exact raw exchange, reopen launcher, finish checks and final ledger state. Single-fixture use still gets one default45-second deadline. The timer is cleared through `finishObservation` on success or failure after nested fixture cleanup. Request5-second and outer60-second limits are unchanged; inherited finite child shutdown is unchanged. There is no deadline enlargement or production edit.

Two authored scaled controls pin no second fixture after aggregate expiry and two sequential observations consuming one budget, with cleanup and `[passed, failed]` outcome distinction. A first completed fixture's passed ledger is explicitly not represented as whole-test success.

An independent real-process probe used a3-second aggregate deadline, rather than another60-second failure run. It launched two actual installed fixtures, made actual HTTP requests in each, reopened the second child, and asserted that each fixture received the identical signal. At shared expiry the next exchange rejected; cleanup completed; the thrown error was exactly the shared abort reason. Ledgers were `[[passed,false],[failed,true]]`. A process-table check found no children from this disposable build after return. This directly exercises propagation and drain beyond the authored fake-timer/helper controls.

## Verification

Disposable clone `/Users/dbox/repos/wt/vickie/vickie-snapshot-deadline-native-review`, exact frozen source before probe-only edits. Pinned Node24.16.0 / pnpm11.20.0. Frozen-lock install, pre-build typecheck and build all exit0 under `set -e`.

- PASS: original installed Snapshot file, **8/8 tests**,1.48 seconds, exit0. This includes both actual service tests and the two new scaled controls.
- PASS: independent real-process aggregate/reopen cancellation probe, **1/1**, eight unrelated tests filtered,3.34 seconds, exit0.
- Probe edits only in disposable checkout; author source remains clean and unchanged. No test-owned child remains.

Evidence directory `snapshot-deadline-native/`: `install.log`, `typecheck.log`, `build.log`, `focused.log/json`, `probe.log/json`, `native-probe.patch`. The probe reports `native-aggregate-real-child-closure` with shared signal, actual reopen, post-abort rejection, truthful ledger states and empty child list.

No fresh packed run/full-suite claim is made by this review. Root owns final combined validation and publication. No public claims, flags or serving semantics changed.
