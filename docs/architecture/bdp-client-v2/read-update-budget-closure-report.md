# Installed cumulative budget council closure

Frozen source **7d2d23c99e9904203ef880d8319d30d54f7117b8**, direct child of a0b44bef83308cecf57d96065df191d71249119e. Checkout `/Users/dbox/repos/wt/vickie/vickie-read-update-budget-closure`, branch `codex/vickie-read-update-budget-closure`, GitHub origin gastownhall/bdp. One test file only, +22/-3. Frozen550 and a0 remain unchanged. No outward actions or runtime changes.

The three Target2 LOWs in `receipt-and-budget-followup-claude-review.txt` are addressed:

1. Awaiting child exit no longer suffices: `closeInstalledChild` requires normal exit0 or SIGTERM. Small assertion-input sensitivity controls reject settled SIGKILL and exit1. These controls are explicitly test-only; every actual HTTP scenario still exercises real child closure.
2. Pre-aborted run requires exactly harness-error/not-run and the stopped-before-run reason, zero sessions, zero observations, and no eligible claim.
3. After run returns and its partial report is saved, `controller.signal.throwIfAborted()` throws the original abort reason before scenario assertions. The existing timer remains cleared in finally. Inner HTTP5s, launch10s, per-control30s, aggregate cooperative90s and outer120s limits remain unchanged. No artificial90s stall test was added.

Actual Linux evidence, rather than a timing prediction: root CI38376385086 at b54392340de34c6c3e8aa07d959d7b6458338261 logged the cumulative matrix at43,479ms, and 8,285tests/141files passed with one optional skip, whole suite179.41s (`pr88-b543923-ci.log`). This closes the previously unknown Linux duration for that historical40-scenario configuration; it is not a guarantee about future fixtures.

Verification under Node24.16.0/pnpm11.20.0:
- Frozen install; clean→prebuild typecheck→build PASS (`budget-closure-build.log`).
- Actual matrix plus all focused controls: **66tests/onefile PASS57.50s** (`budget-closure-tests.log/.json`).
- Format, lint (existing warnings), dependency boundaries11packages, diff whitespace and unchanged89-row Read evidence PASS. No broad full-suite claim for this successor.
- Fresh packed protocol/server installed outside checkout: **66tests/onefile PASS46.89s** (`read-update-installed-7d2d23c/tests.log`, `results.json`). Actual progress source is exact7d2d23c, sourceDirty:false, counterfactualApplications:0, 41sessions. All41 authored scenarios passed; uncovered required IDs remain explicit and claimEligible:false. Coverage remains129required/28proposedwhole/101uncovered, not qualification.
- No matching test-owned installed child remained after completion; working tree clean.

Fresh progress artifact: `read-update-installed-7d2d23c/progress.json`. Manifest bab6a5341d4488fdf8fbcbb3ebbf4d96c7fe856fc8574b9be3d0126c271de145; fixture9bf5fcf2f4216480985d10b15bba3881fa357f973115834bafbd3c696610b11c; exact HTTP configbe0e83fb1663e2fd91b6c5b90b5b18e3ba0b65887b18479b880d9edec88290f9; installed payload0f914e6d688078dab562bc84a092fd143370abd4000f8ce5c6b0d0635c24ac26. No fixture, route, manifest, catalog, public admission or stored evidence changes. Source provenance is launcher/reviewer verified, not target attestation.

Independent native retention/contracts review is separately complete in `retention-contracts-followup-native-review.md` (no actionable finding; RU88focused tests, TX27focused tests). Next effective-schema cohort remains plan-only in `read-update-effective-schema-plan.md`.
