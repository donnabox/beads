# Receipt source validation outcome

All ten released stages completed once with exit0: offline setup, typecheck, build, format, lint, boundaries, Read seal, and the three exact source selections. Offline setup reported lockfile current/already up to date. No retry, timeout, forced cleanup, unhandled error or unexpected executed test occurred.

Exact source cases:3 owner +5 service +6 HTTP =14 passed;143 other cases skipped. Full names and original raw/JSON hashes are retained in source-outcome.json. The source-derived selected workload is68 HTTP,16 owner opens/14 directories/6 listeners/2 readonly inventories; those modeled lifetime totals are not independently instrumented counts.

The new HTTP witness independently logged13attempts and13completed responses:200,200, six404,200, two404, two401. Eight context tuples passed; first headpos1, laterpos2;401has no tuple. HEAD bodies were0bytes with GET representation length preserved. Seed/page/advance bodies990/334/664bytes; missing Problem bodies134bytes and anonymous GET136bytes. Full body/schema oracles ran in memory; retained trace is safe status/URL/method/selected headers/body lengths, not complete body/wire retention. The other55HTTP remains a source-derived count from selected completed paths.

Both new local cleanup callbacks completed. All supervisors recorded no residual group, sampled PID or runtime directory. An independent final process check found none of the recorded PIDs; each owned temporary root contains only node-compile-cache. Those caches are preserved and are not service/store leaks. Sampling does not prove every short-lived process or directory was observed.

All eight frozen source hashes and complete diff remain identical to bindingedc5c98c/diff837bacdc on parent015cac1a. No source edits, commit, push, package or installed service occurred during validation. Control and installed gates remain closed. Actual Claude review is separate and remains a required root gate; these bounded passes do not claim full CI or public qualification.
