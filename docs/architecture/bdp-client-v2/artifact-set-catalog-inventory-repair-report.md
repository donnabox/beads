# CI catalog-inventory repair

Frozen **87abe223dd27519d900fa015845ee3d90c97819a**, tree **cade4fe992e6bb93852440b805848458a5afaf23**, direct ordinary successor of **4819068d0201fb3d1fe32dd6adaf8625ef9ba666**. Isolated checkout `/Users/dbox/repos/wt/vickie/vickie-artifact-set-catalog-inventory-repair`, branch `codex/vickie-artifact-set-catalog-inventory-repair`. Original 481 remains unchanged. No push or other outward writes.

## Diagnosis and narrow fix

PR93 CI 38441826564 failed exactly one test because the catalog-inventory test recognized only `read-update-v1.json`. The already-reviewed artifact-set index explicitly includes `read-update-reference-continuation-v1.json`. The test therefore rejected an approved private shard as though it were an unapproved matrix. This is a stale test inventory, not a runtime admission defect.

Reproduced unchanged 481 using pinned Node 24.16.0 / pnpm 11.20.0: catalog file **5 passed, 1 failed**, exact same continuation-name assertion at line 92 (`red.py`, `red.json`, `red.log`). The single test file is now updated; no artifact, factory, runner, runtime, schema, public export or seal change.

The allowlist is the two explicit reviewed filenames, independent of manifest tags. The test also pins the artifact-set index's exact version, shard IDs, sources and order to that closed list, requires both files to exist, and retains version 2/catalog-ID/cumulative-required-ID checks for the two approved entries. All other manifest files retain both former denial checks: they cannot claim the RU catalog or contain any RU-required ID. Existing sealed Read catalog collision and normative composition tests remain.

Two permanent countercontrols use the actual continuation manifest: an unapproved filename with its real RU catalog must fail; a Read filename with a forged Read catalog tag must still fail on the contained RU row. Thus admitting the continuation is not a generic profile/catalog exemption, and changing the index alone cannot silently authorize another shard. Existing artifact-set factory and public-admission regression files remain untouched and are rerun.

## Verification and evidence limits

Pinned final run: **424 passed across 7 files**, no failures/skips, including 7 catalog tests and targeted artifact-set, artifact-bundle, cumulative-artifact, profile-contract-plan, runner and Transactional-cumulative tests. Frozen install, typecheck, clean build, boundaries, full format/lint, unchanged Read evidence verifier and diff-check pass. Exact commands/results: `check.py`, `results.json`, `regressions.json`. Initial formatting-only failure was corrected with the existing formatter; its output remains under `initial-check/`. No assertion was weakened to address that formatting failure.

Only `packages/conformance/src/read-update-catalog.test.ts` differs. `summary.json` records identical runtime/protocol/catalog/matrix/fixture/private-artifact/app/script trees. No evidence artifact was regenerated; prior 481 installed 223 and CI 8828-pass/1-fail results retain their own source and outcome. This targeted repair still requires root's new exact-head CI; it does not turn the prior failed CI green retrospectively.

Source is clean/frozen, no owned Node child remains (`cleanup.json`), and no hook was bypassed. Prevention: catalog inventory must be reviewed when the explicit artifact-set shard inventory changes; this test now pins that relationship while preserving closed-world denial.

Preserved installed evidence: [481 integration report](/Users/dbox/repos/wt/vickie/bdp-integration-logs/artifact-set-execution-main-integration/report.md), [exact source/configuration/file bindings](/Users/dbox/repos/wt/vickie/bdp-integration-logs/artifact-set-execution-main-integration/summary.json). Its fresh packed 223-pass cohort ran on 481, not this test-only repair. No fresh packed repeat was performed.
