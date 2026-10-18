# Native V3 static outcome and helper-script audit

No blocking finding. Read-only reconciliation of retained actual outcomes and narrow gate/helper-review source changes. No rerun, product import, build, service, installation or source edit by this reviewer. Installed results do not yet follow from these static results.

## Actual source and stages

Actual working diff remains06cf7eb8ca3d4e0a4caaeee81ee5688a9e44f24895809414b164833068d207bb against HEAD453b6dab386746bb03aac2c0a293572b8a71c969. All12 current file hashes match prepared-source.json, the author90ced9c6 binding and every started/exit record. All nine actual argv arrays exactly match their declared command entries; their starting diff hashes match the frozen successor. Source remained unchanged.

|Stage|Actual exit/supervisor|Elapsed seconds|
|---|---|---:|
|offline-setup|0/0|0.449|
|typecheck|0/0|1.165|
|build|0/0|1.166|
|format-check|0/0|0.559|
|lint|0/0|0.659|
|boundaries|0/0|0.753|
|read-seal|0/0|0.556|
|static-cases|0/0|1.852|
|build-inventory|0/0|0.238|

Every exit record has timeout=false, forcedCleanup=false, empty errors/remainingAfterCleanup/sampledPidsStillPresent/runtimeDirectoriesStillPresent. Build log records clean followed by tsc -b; typecheck uses tsconfig.check.json. Format checked439 files, no fixes, one existing oversized3.5MiB RU matrix warning. Lint checked439, no fixes,765 warnings/48 infos, exit0. These warnings are disclosed rather than misrepresented as a warning-free run. Boundaries checked11 packages. Read seal verified89 historical rows.

Vitest reports success:true,30 passed,0 failed,601 pending/skipped,631 total. I compared all executed path/fullName pairs against commands.json: exact30, with no extra executed case or missing selected case. Counts are executable-manifest7, http-executor12, runner3, cumulative5, profile-contract1, app2. Six files; JSON numPassedTestSuites22 includes nested suite accounting and is not22files. Source workload remains the selected single loopback/listener fixture, zero product children/stores; this is source-case accounting, not independent network capture.

## Build inventory and cleanup

Recomputed all312 build inventory file hashes (2413225 bytes), no mismatch. Reconstructed the inventory path set independently: LICENSE, protocol/server package.json, protocol dist+schemas and server dist; exact set, no symlink. This is the prospective package-output inventory, not a claim about every workspace build artifact or already produced tar archives.

Independent ps inspection at2026-10-15T07:51:50.515447Z found none of the29 sampled PIDs and none of groups33582,33616,33683,33939,33998,34060,34118,34227,34307. Current attempt temp directories contain only node-compile-cache for eight stages; build-inventory.tmp is empty. The short-lived extra static temporary entry recorded by sampling is absent. No owned runtime directory remains. Sampling cannot prove that an unsampled detached process never existed.

The original admission-failure exit1 artifact is still unchanged (SHA4ad0910028cc5f56cf9180002bdfba4e3f0d65401d07ec20bfc3a492b290a611). Successful V2 remains distinct predecessor evidence. V3 is now actual successor evidence, not a relabeling of either prior run.

Primary actual hashes:
- static-cases.results.json1240e387629e298a7fb9164f57412bc99dcc25b7e311965aadb3e3fe3b5c273c
- static-cases.exit.jsoned9799e3e8900f91088096e3c9461bd524acdacab3ffbb6d2c17ca4d170abcfe
- build-output-inventory.json7bde8c94657b1b48d7000e7c96bc71d9af11d34f1a93f0b55a9f7ef93ea9878d

## Narrow installed-preparation source audit

Read gate.py0996ba64e3fa88fb718044c6fd8ce79a02d10b4d8f721ea962ad18bb6a402b02 and helper-review.py2ffc5cf9ed82866ccdc60b94cee1281c309702487e0912d6db1b0c6c111956a1 in full, comparing the declared V2 delta and the actual helper/binding source.

The gate requires exact12 source hashes/footprint, pinned preparation, approved command IDs, source/build binding, clean ordinary final commit for installed phases, installed entry/lock/source provenance and explicit dependency completion. Helpers receive only installed entry/source overrides; no matrix output override. The order is verify-installed, collection helper+review, Snapshot helper+review, then three matrix cohorts. Existing supervisor is hash-pinned, not copied or expanded.

The helper adjudicator's assumptions match actual source:
- collection emits exactly one private-installed-compatible-missing-minimum record with closed, childPid/childExitCode, bodySha256 and wireSha256; its method/URL/status order is2POST200, seven404 observations, final500 behind observation (10total).
- Snapshot emits private-installed-snapshot with one child record, wireSha256 and no bodySha256 requirement; selected case is GET200 then GET503 (2total), no reopen.
- both log outcome/aborted/claimEligible/workspaceDirty, source label, entry/hash and binding.served fields. installedTransactionBinding supplies canonical authority, runtime digest and served source/payload/contracts/policy/authentication/fixture digests. fixtureDigest is SHA256 of exact fixture bytes (profile-contract-plan.ts:157), so the adjudicator's direct byte hash is correct.
- exact selected path/title and one-pass results are checked; duplicate/missing observations fail. No body or header is fabricated from fingerprints. Actual source declarations are not treated as child-attested Git.
- successful review requires reported child exits and an independent PID absence check; completed() additionally checks supervisor cleanup records. The copied genuine observation is written exclusively after successful reconciliation, preserving the original log.

No bad actual-record field or fixture-hash assumption found. These scripts have only been read here, not invoked. Full installed dependency closure/provenance and root adjudication remain required before claiming same fresh payload. The planned helper total is12HTTP/2children/listeners/store-owner lifetimes/dirs, operator-only; it adds to the three unchanged matrix cohorts208expected/225maxHTTP/26children, giving220expected/237maxHTTP/28children. Static one-loopback request is separate. Existing time/request/test bounds remain unchanged. Exact-head CI and actual supplemental council remain separate gates; no whole-profile/public qualification.
