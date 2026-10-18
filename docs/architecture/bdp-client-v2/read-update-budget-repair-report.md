# Cumulative test scheduling budget repair

Frozen source **55045386812ab3b14042ec573024a90a1b38a829**, direct child of75f279cec24ef8d7eeae0fde56120ddab82ec907. Isolated checkout `vickie-read-update-budget-repair`; branch `codex/vickie-read-update-budget-repair`; origin gastownhall/bdp. One test file only,55insertions/7deletions. No runtime or evidence/admission edits.

Observed CI38374131399 at637e0ed (`pr88-637e0ed-ci-failed.log`) had8254pass/1fail/1skip; sole failure was the cumulative matrix exceeding30000ms. It executes40 fresh authority scenarios,609current client observations plus setup/native traffic. Linux selected cases took about1second each; one aggregate30second budget was inconsistent with that work. The fullCI took131.90seconds. This is scheduling evidence, not an authority timeout or new protocol limit.

The single full cumulative test now has a finite90000ms cooperative timer and120000ms outer Vitest ceiling, reserving30seconds for teardown/reporting. Existing inner HTTP5000ms, launcher10000ms and every selected-row test30000ms remain unchanged. No global timeout or suite configuration change. Signal is passed into the existing runConformanceMatrix API, which relays it to scenario preparation and exchanges, marks canceled/unstarted outcomes unsuccessful, and cleans sessions through the raw target. The outer run's existing finally closes the target and removes the store; the new timer is cleared in finally on success/failure. Actual child exit is asserted after every bootstrap/session close.

A pre-aborted negative starts/closes the bounded bootstrap, proves no scenario session/observation executes and no pass is awarded; actual ChildProcess exit is asserted. This exercises cancellation wiring and no live bootstrap leak without a flaky small wall-clock threshold. It does not claim an induced90second production stall test. Active scenario cleanup continues to use the existing runner/target implementation and its tests; no lifecycle redesign.

Validation: clean→prebuild typecheck→build PASS.58matrix tests PASS39.76s before the final child-exit assertion; final main+preabort tests with that assertion PASS2tests/56filtered10.50s. Format/lint (warnings only), dependency boundaries, diff check and unchanged89-row Read evidence PASS. Logs `read-update-budget-{build,tests,final-build,final-tests,checks}`. Fresh frozen packed matrix results follow. This local run does not replace the required exact LinuxCI rerun.


Frozen packed verification: **58tests/1file PASS39.25s**, `read-update-installed-5504538/tests.log/.json`. Actual progress binds source55045386812ab3b14042ec573024a90a1b38a829 with sourceDirty=false;40pass/94missing,101uncovered, claimEligible=false. Main observed40sessions and the cancellation negative passed with actual child exit asserted. Exact manifest/fixture/configuration digests are unchanged from75f:
```json
{
  "artifacts": {
    "catalogDigest": "e26d3fc758a34ebb0b5028be15d28e3e09db87dc7984bba7bad6a00b26329aa0",
    "manifestDigest": "5bc99fb31c6c4157d21a2f11a381bce312c6494eb7952207115d9d97702184c8",
    "fixtureDigest": "9bf5fcf2f4216480985d10b15bba3881fa357f973115834bafbd3c696610b11c"
  },
  "exactConfigurationDigest": "89d471885903a3e1dda4939f54993a88696e1b8c336dfef287a0c28966c47d0a"
}
```
