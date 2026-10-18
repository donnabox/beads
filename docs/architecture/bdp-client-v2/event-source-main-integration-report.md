# Provisional Event SSE integration validation

Frozen source **6a13489d0de673fd59c50b6be1d65e93f3b4be6a**, tree **28eea2abfc5e53399c9a92dd7da44eb74b33d418**, checkout `/Users/dbox/repos/wt/vickie/vickie-event-source-main-integration`, branch `codex/vickie-event-source-main-integration`. Ordinary parents: candidate `33ebad3ec539852ff46990a3647ca8195276a9f8` and main `f31424cdfc57871925a04659208686bac7bddfed`. No source changes or outward writes.

Remote launch session guard remained `2026-10-08T20:26:35Z-81e5`. Independently reproduced root's source audit: exactly **13 files** differ from main; every differing file is byte-identical to frozen candidate 33ebad3. No conflict resolution introduces another source delta. Existing timeouts/caps remain unchanged. This is validation of a provisional integration, not acceptance: native candidate council completed without findings; external Claude39205 was pending at assignment. Root owns final external disposition and publication.

## Completed inherited validation

Before packing, inspected root's `run.py`, `current.json`, `results.json`, `complete.json`, process state and exact source. All eight stages had exited 0 and the recorded process had ended; no build was duplicated or overlapped. Frozen install, typecheck, clean build, boundaries, format, lint and Read evidence verification passed. Root's server plus six conformance-file run passed **5849/5849** tests, no failures/skips. These are actual integration-head results, not predecessor results relabeled.

## Fresh installed validation

Copied bounded process-group supervisor into `packed/validate.py`, repinned source/path, packed protocol+server from the completed exact build, and installed into a new isolated `packed/install/`. Pinned Node 24.16.0 / pnpm 11.20.0. Each stage checks exact HEAD and clean status. No full 8k suite was repeated.

| Actual installed cohort | Passed | Failed | Skipped |
| --- | ---: | ---: | ---: |
| event-sse | 1 | 0 | 0 |
| changefeed-sse | 1 | 0 | 0 |
| finite-tx-snapshot | 30 | 0 | 0 |
| packed-ru | 173 | 0 | 0 |

Commands are preserved in `packed/validate.py`; exact exits/elapsed durations are in `packed/stages.json`, with verbose and JSON Vitest outputs per cohort. Roots: `transactional-event-sse-installed.test.ts`, `transactional-sse-installed.test.ts`, `transactional-matrix.test.ts` + `transactional-snapshot-installed.test.ts`, and the four-file RU matrix/exact-client/development-server/installed-read-oracles cohort.

Event SSE observes Scope/Bead/Link replay and sparse Bead continuation in four actual wire recordings. Idle **15199.46ms**, **2** heartbeat comments, then **1** relevant live Event frame. Existing changefeed SSE idle **15192.21ms**, **2** comments and **2** group frames. These are observed timings on pinned Node, not a universal latency or capacity guarantee. Actual settings are retained in the safe summary, including 100ms polling, 5000ms heartbeat, one stream, one group per Event observation and 4MiB packet cap.

Finite TX report: `{'harness-error': 251, 'pass': 4}`, with 5 sessions. Only the four authored rows pass; missing-plan rows remain explicit, no whole-profile claim. Snapshot records retain passed/aborted outcome, exact source/artifact binding and completed-exchange counts separately.

Current-main RU report: **49** authored sessions, **985** raw client observations, **876** reported exchanges, **44** setup and **21** native Read observations. Scenario states `{'pass': 49, 'harness-error': 85}`. Coverage remains **30/129 whole**, **99 uncovered**; these counts and observation layers are not additive. Public claim eligibility remains false in every cohort.

## Actual binding and package identity

| Binding | Value |
| --- | --- |
| Event source revision | `6a13489d0de673fd59c50b6be1d65e93f3b4be6a` |
| TX payload | `9add08f913ba0bb5960926988b181987a6ddc150a4840d16008ec7b036a6b16a` |
| TX contracts | `826a867731bf28d606110d97443772081cbcba3986c1a5b6ced66082c6bfb094` |
| TX runtime configuration | `5fce76b1c68d9c17680e3fa4b7119fa39ff4f8aab42c3872a61860ba3c2faa03` |
| TX policy | `bae3e7241816f6f3c54857f59011a85131580f3b0aaa20f4a4612dcde74427b4` |
| TX authentication | `fb3ab3f71c46b031185b399378ca236c110f064dc86d77804d170729e03b5156` |
| RU configuration digest | `3224c31c83988dc0606cb76ee1f4ba8e6394159bdf7daaccc301461316cc2477` |
| RU catalog | `e26d3fc758a34ebb0b5028be15d28e3e09db87dc7984bba7bad6a00b26329aa0` |
| RU manifest | `c86b33535f25a1d117ce021a3bb70164f5635bdaee50758e3428c130ccc3227f` |
| RU fixture | `f5f38235dfd28cf06e760688e2054d3e29d8d64b08c4c729baa94f380f1f7a63` |
| protocol.tgz SHA-256 | `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c` |
| server.tgz SHA-256 | `83c9cb7a02a512cc6e108766b5d7775d9331b21185067b95a123a5885e021bc8` |
| transaction-installed-process.js SHA-256 | `bcf153170f953f9e28c6c89f1153be907f64df289e64f3ad929cad4a4a54c47c` |
| installed-reference-process.js SHA-256 | `3eb9a6bc8df1237d78ba4aae7e95589df6fb7f4781f572e04adda583f359c71d` |

`summary.json` retains source audit, package byte lengths/hashes, complete actual settings, separate TX/SSE/Snapshot/RU bindings, canonical operation targets and RU method/URL counts. It whitelists metadata and excludes raw credentials, response bodies and wire encodings. Original test logs retain their source-owned wire observations; no credential dump was added.

## Cleanup and limits

All owned commands finished naturally under bounded supervision; no timeout or interruption counted as success. Final process inventory found no owned installed Node children; exact HEAD and clean source were checked at completion. No unrelated processes were terminated. No seals, source flags or profile gates were changed.

Outstanding external council and root disposition remain independent gates. These installed observations do not establish complete Transactional or Read+Update qualification, optional History, pending metadata-only incident-Link policy, Snapshot minimum semantics, release promotion or public write admission. This report does not claim an entire 8k suite or new Read cohort generation.
