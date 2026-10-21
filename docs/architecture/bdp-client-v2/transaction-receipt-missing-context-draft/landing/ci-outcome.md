# PR128 exact-head CI failure

Run38739669955/job116437776256 failed on3968081f8231566908e69045424b067c3c36dc40. Sole failure: apps/bdptest/src/transactional-matrix.test.ts:902, “aborts the actual cohort budget after real wire and drains all installed children”,2087ms. Expected seen=true, actual false. The prior assertion found “installed cohort deadline” in the error causes. All later published/duration/children-drained assertions were not reached, so this failed callback does not prove its cleanup oracle passed.

Source shows observe arms its2000ms test-selected timer at lines152–155 before bootstrap launch/close, route setup and scenario launch; seen is set at885 only after the first returned ordinary exchange has nonempty wire bytes. The result is consistent with pre-wire startup/setup/dispatch consuming the deadline, but no per-stage observation identifies the exact delay. It is not a Vitest20-second timeout or proof of a receipt semantic regression. The entire failing matrix file is byte-identical to base015cac. No timing change or rerun was performed.

Suite9299passed/1failed/1skipped,9301total,174passedfiles/1failedfile,783.14seconds. Job996seconds. Setup/install/format/lint/boundaries/typecheck/build succeeded. Full pnpm test failed; ready preflight, smoke and Read-cohort evidence gate were skipped, as were setup-go/setup-node post steps; remaining post steps completed successfully. No downstream success or headroom claim is made.

The corrected minimum callback passed51ms; new source receipt HTTP case99ms, owner page-context case22ms and installed receipt callback1134ms passed in this failed overall run. These are exact CI callback observations, not an overall qualification.

Actual checkout log pins syntheticde34f7479eba629bef73a979b12e5d6b9773ddba. Remote commit read confirms ordered parents[015cac,3968081] and treea597518b equal candidate. Raw full log retained unchanged: SHA2560d858f06611714eb1ef95cc23541f32d60005a36a5bb4e492d02fb83e9833d35,3329440bytes. Full job/run/PR snapshots and90second polling history preserved. Landing remains blocked on exact-head green CI; root owns diagnosis, any repair and publication.
