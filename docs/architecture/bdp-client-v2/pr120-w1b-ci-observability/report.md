# W1b assertion diagnostic preservation

Frozen clean ordinary successor `427cfef6decbb422b393024e36218affc26007fa`, tree `9075897867d88344a8477e1f4a8e1a00a50f1228`, parent `70ac666002708202d2f93195eed90dfe8541300d`, branch `codex/vickie-w1b-ci-observability`. One file changes: `apps/bdptest/src/read-update-matrix.test.ts`, four additions/two deletions. All other 662 tracked files have identical Git object/mode entries to the parent. Production, configuration, package manifests, helper implementation, assertion values, inputs, controls and timeout are unchanged. Commit includes the requested Agent-Signature trailer. No push, PR update or CI retry performed.

The W1b parameterized case now passes the existing retainedRetryDiagnostic projection to all three existing assertions. Its projected failure position/status/route fields omit body, headers and error text and redact known fixture credentials. This makes another failure actionable; it does not identify or repair the original CI failure's cause, which remains unresolved.

## Verification

Exactly two cases ran once on the reviewed source bytes before the commit, with an anchored two-title selector and maxWorkers=1. The W1b link-metadata actual HTTP case passed in 709.812 ms; existing diagnostic projection/secrecy unit passed in 0.781 ms. Reporter totals: 2 passed, 0 failed, 200 skipped by selection. Vitest reported 1.22 s; supervised command elapsed 2.316 s. Successful W1b assertions establish counterfactualApplications=1, failed scenario and exactly the intended owned-metadata.metadata assertion. No full-suite claim.

Offline frozen dependency setup, explicit fresh build, typecheck, touched-file formatting, touched-file lint and diff whitespace check all exited 0. Lint retained 226 warnings/8 infos from the existing large test file; no fixes or unrelated edits were made. All six supervised commands completed without timeout or forced termination. Full logs, JSON reporter and command/exit/timing records are retained here.

Workload is expected 36/max 63 HTTP from the unchanged action sequence, not a freshly retained per-request count. No further per-request instrumentation was authorized or added. Two installed server child PIDs 4967 and 4971 were observed under test process group 4904; test close assertions enforce exit 0 or SIGTERM. The prior instrumented diagnostic's measured 36 exchanges/two children remains separate under `pr120-w1b-diagnostic`; it is not relabelled as this run. This verification uses the newly built local emitted entry, not a newly packed or installed distribution. Entry hash in source.json identifies only that file, not the complete served dependency payload.

## Cleanup and limits

Supervisor sampled processes and temporary roots every 100 ms. It observed one `bdp-cumulative-ru-g2D56r` root; the root was absent afterward, as were all 27 recorded PIDs and all six owned process groups. No matching temporary roots remained. Successful execution also exercised unchanged child-close assertions and the run's finally cleanup. Process sampling can miss short-lived processes; absence of recorded PIDs and root inventory is the retained external observation, not an exhaustive per-request lifecycle report. Configured workload remains two store directories, two child listener lifetimes and the existing availablePort probe.

`source.json`, `summary.json`, `cleanup.json`, `selected.json`, six `.exit.json` records and their full `.log`/`.processes.log` files carry provenance. The preflight and source.diff remain unchanged. Native/external review, exact-head CI and any outward action remain root-owned gates.
