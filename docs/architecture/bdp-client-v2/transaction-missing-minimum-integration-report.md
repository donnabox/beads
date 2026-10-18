# Discovery-limit / compatible-missing integration

Frozen source **4697a737b1446021d920f201552cb27e5c207cc5**, tree **3669fb366dbc7a6113eccecc37aa65e9bd5431bc**. Checkout `/Users/dbox/repos/wt/vickie/vickie-missing-minimum-integration`, branch `codex/vickie-missing-minimum-integration`.

Ordinary merge parents, in order:

- PR107 repair `26d6cd1b947db8c00e24015771506d26098a293e`.
- Compatible-missing minimum `7791aed7394e30fef2d2c1ed4fed3f43504e42ba`.

Common base `9b95df5dd866dbaf84150404c4db86ac8054b724`. Both author checkouts were preserved. The merge was entirely automatic and conflict-free, including the shared progress document. Independently recomputed `git merge-tree --write-tree` returned the exact committed tree. Seven non-document donor files are byte-identical to their respective donor commits; only the progress document combines both branches. Its PR107 limits/optional-group wording and both missing-minimum scope/Event convention edits are preserved. `donor-preservation.json` records per-file donor/hash and both document comparisons. No integration-specific production or test implementation was added.

The ordinary merge commit carries an Agent-Signature attribution trailer. Guarded push verified the existing remote Vickie role claim, then created only `codex/vickie-missing-minimum-integration`. `git ls-remote` independently confirms exact4697a737 at that ref; `remote-verification.json` retains the result. No PR107 advancement, PR creation, main merge, readiness change or public claim was performed.

## Validation actually completed

Pinned Node24.16.0/pnpm11.20.0; saved `validate.py`, `checks.py`, `pack.py`, `pack-resume.py` and `exits.json` contain exact commands, limits, exits and durations. No hook was skipped. No full workspace suite was run.

- Frozen-lockfile install, clean, prebuild typecheck, build: PASS.
- Combined affected service/discovery/runtime-configuration/Read/HTTP/consistency/protocol-discovery/installed-contract tests: **256 passed,8 files**, no failed or skipped cases. Stage3.865s; structured `affected.json` retained.
- Format check, lint, dependency boundaries and actual existing Read-seal verification: PASS. No sealed evidence/constants changed. Existing configured formatting exclusions remain exclusions, not newly formatted artifacts.
- After committing clean source: fresh build, actual protocol/server pack and offline install: PASS. Installed test entry points point into this run's new install directory, with exact source declarations and clean-source assertions.

| Fresh installed cohort | Actual result | Attribution |
|---|---|---|
| Compatible missing minimum |1 test passed,10 completed requests,1 child|New missing404/current-context path and preserved final behind gap; status sequence[200,200,404,404,404,404,404,404,404,500] (eight reads total: seven404 and one500). Exact response sequence is retained in summary.json.|
| Existing collection minimum |1 test passed,38 completed requests,1 separate child|Unchanged original scenario, including actual waits and position-oracle corruption controls; not counted as part of the new10.|
| Discovery boundary |1 test passed,12 completed requests,1 separate child|Six authenticated refusal500, three anonymous401, real batch/afterstate/retry200. Confirms wire refusal, not execution of internal discovery assembly.|
| Retained retry/current position Update test |1 test passed;2 selected catalog scenarios,50 raw HTTP observations,2 scenario sessions|sequence.idempotent-retry15 and idempotency.cross-carrier35.|
| Reference syntax/category Update test |1 test passed;2 selected catalog scenarios,82 raw HTTP observations,2 scenario sessions|singleton.no-local-bindings53 and alias.non-canonical-target29.|

The first two TX tests ran together in5.269s; discovery stage0.855s; the two Update tests stage2.665s. Update deliberately filtered200 other tests. Thus “two Update tests” represents four selected catalog scenarios; no broader catalog pass is implied. Both Update reports have zero setup/nativeRead observations and record raw-http1-serializer as the source of all132 observations. They retain actual catalog/manifest/fixture/configuration bindings and complete scenario results separately. All claimEligible flags remain false.

## Preserved collector failure

After all three TX cohorts had already returned exit0, the initial pack script's post-processing assertion failed: it selected every JSON line carrying an observation field, including six diagnostic/control lines from the original collection scenario, then incorrectly expected exactly three reports. This was a **collector failure**, not a source, test or transport failure. Original `pack.py`/`pack-attempt-original.py`, test logs and `collector-failure.txt` preserve it. `pack-resume.py` narrowed selection to the existing private-installed report labels, validated the original completed cohorts and then ran the still-pending Update tests. No passed cohort was rerun or relabelled. The summary explicitly records collector exit1 separately from all executed validation-stage exits0.

## Source and payload bindings

- Protocol tarball SHA256: `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227`.
- Server tarball SHA256: `762c5ca78708c5b88b6bc00917d825ba77d09811eeb567babbb9ead49b5a3815`.
- Actual installed server payload identity across these TX/RU runs: `e8f8920fe29e39f94ac06cee7c81882bcb701e2c7796928101be6fab48e3984e`.

`source.json` records exact source/tree/parents and all eight changed-file hashes versus common base. `summary.json` preserves the three distinct TX runtime configuration identities, actual installed contracts/policy/auth/fixture bindings, exact completed response hashes, tarball hashes, validation exits, and both separately hashed RU report references/selection counts. SourceCommit is a launcher declaration checked against Git, not independent target attestation. Per-scenario configuration identity may differ; no single universal configuration claim is made. Historical donor evidence was neither edited nor reused as current execution.

## Cleanup and remaining gates

TX children77571,77582,77595 all exited0 and are absent. Both RU reports assert cleanupComplete; path-bound OS process inspection found no remaining transaction-installed-process or installed-reference-process under this run's install directory. Clean worktree confirmed after all runs and remote verification. `cleanup.json` retains exact checks. Temporary test databases were owned by the fixtures; no shared DB or unrelated process was touched. Pack/install/report artifacts remain outside the source checkout for inspection.

No new semantic issue was found by integration. External council on the donor, root final review, exact publication/CI decisions and any later profile activation remain separate gates. This report establishes the stated combined focused and fresh installed evidence only.
