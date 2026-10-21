# Reference evidence ordinary integration

Validated and frozen clean **b20450950af93f882e4d3a1dafc547cc6f303ac7**, branch `codex/vickie-reference-integration`, checkout `/Users/dbox/repos/wt/vickie/vickie-reference-integration`.

Ordinary parents: main `1f9d7dbc401f6a265cf1ce0bdfea0fe493cd79a5` and reviewed reference `88e428ce5b24a666cf2a27f0cbb34cd64dfcf7e5`. No push, main mutation, protocol or runtime edit. Remote session guard was read before work: `2026-10-08T20:26:35Z-81e5`.

## Merge review

Only four evidence/test files differ from main: RU matrix test, RU fixture, RU manifest, RU coverage inventory. The sole conflicted test hunk preserves main's media ledger **183 actions / 149 raw / 34 client** and the reference ledgers **99/46/53** and **72/43/29**. The newly added uncovered count is **101**, reflecting main's existing **28/129** whole rows. All main admission promotion flags remain; reference A/B remain false.

Automated raw-span audit proves all **47 existing scenario blocks equal main byte-for-byte** and both added blocks equal reference88e byte-for-byte. The entire fixture equals reference88e. Every coverage row equals its intended parent. The manifest has **3,626,754 bytes and 97,112 nodes**, with **2,888 nodes** of unchanged structural-cap headroom. No cap or semantic adjustment was made. Test additions, fixture-pointer oracle resolution and failure-aware live-identity accounting match the reviewed reference parent; main's paired-media and exact countercontrol assertions remain intact.

Server, protocol, client, configuration and scripts tree identities equal main. `merge-audit.json` and `runtime-trees.json` contain exact audit data. No additional integration-specific count repair was needed beyond root's resolved 103→101 count.

## Validation

Pinned Node **24.16.0**, pnpm **11.20.0**. Frozen lockfile install, typecheck, clean build, boundaries, format and lint passed. Precommit affected matrix+cumulative tests: **143 passed**, zero failures/skips. Merge committed with sign-off; no active hooks were installed or bypassed.

The committed clean source was rebuilt before packaging. Fresh protocol and server tarballs were installed into a new isolated `packed/install/`; the actual installed RU entry was used for the four-file suite. All committed-source static checks passed again, and cumulative artifact tests passed **6/6**. Installed four-file tests passed **172/172**, no failures/skips. Full 8k runtime suite was not repeated because runtime trees are unchanged; prior main/30351 CI remains evidence for its own source, not this commit.

Fresh baseline: **49** sessions, **985** completed raw HTTP serializer observations, **876** reported exchanges, **44** setup observations and **21** native Read observations. These are distinct layers, not additive counts. Scenario results: `{'pass': 49, 'harness-error': 85}`. Unsupported/unplanned harness-error rows remain explicit. Required129, planned49, whole28, uncovered101; **claimEligible false**.

The focused reference report has **2** sessions, **82** raw observations, **89** reported exchanges and `{'pass': 2}`. The admission report separately has **2** sessions, **122** raw observations, **345** reported exchanges and `{'pass': 2}`. These positive reports have counterfactualApplications0; actual negative controls are recorded in the passing test results, not relabeled baseline successes.

## Exact installed bindings

Canonical Scope `https://reference.invalid/acme/`, profile `read-update`, History false, fixture `reference-read-update-schema-v1`. Five catalog source digests, actual package hashes, and every method/URL route count are in credential-free `summary.json`.

| Binding | Actual value |
| --- | --- |
| Source | `b20450950af93f882e4d3a1dafc547cc6f303ac7` |
| Configuration digest | `3224c31c83988dc0606cb76ee1f4ba8e6394159bdf7daaccc301461316cc2477` |
| Live configuration identity | `1720b63b49a7be3299d4b2865de4db9fcbe47560f217a32eea6397db7ea54ca5` |
| Payload digest | `78a80950acd6c1981d6829fc4d59c05eaafc65e0d68cf20d3d84f61f377acb99` |
| Installed artifact identity | `development-schema5-48667497d6e13064316d0364cf8f49dfc247ab9189306985a357d7b558333dfe` |
| Policy identity | `dbe8dbd5430f26e19e754d38167cd8084a32d702c695550e2ee0063dc2912430` |
| Authentication identity | `66941efc779f4f2431d65fdd5b2387b2af67bd7c1559e09dac032f82259ea16e` |
| Catalog digest | `e26d3fc758a34ebb0b5028be15d28e3e09db87dc7984bba7bad6a00b26329aa0` |
| Manifest digest | `d40a9a3311c11d58d74841ddd57d0f07e69759fb5ba7e51fa371e98382fd42bb` |
| Fixture digest | `f5f38235dfd28cf06e760688e2054d3e29d8d64b08c4c729baa94f380f1f7a63` |
| protocol.tgz SHA-256 | `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c` |
| server.tgz SHA-256 | `aa52a497f921dc00f4b66e94cf31843d0c91613019ddbb9292a9319d1ba8115b` |
| installed-reference-process.js SHA-256 | `3eb9a6bc8df1237d78ba4aae7e95589df6fb7f4781f572e04adda583f359c71d` |
| package.json SHA-256 | `d583f4158c3a393923c37d70a61212f6a716c951a5f3be7ed9bbe3aa8846ca2e` |

## Evidence and cleanup

`precommit.py`, `packed/run.py`, `audit.py` and `summarize.py` reproduce the commands and audit. Logs include verbose and JSON test results, command exit status/elapsed time, source checks and process inventories. Every command completed naturally under bounded process-group supervision. No owned installed child remains. Source was clean at final check and is frozen; unrelated pre-existing processes were left alone. No credentials appear in the whitelist summary/report; tests retain their credential-leak assertions.

No whole reference-row credit, public write-profile admission, profile-wide qualification, fresh Read seal or release claim is made. Root owns external council disposition, exact-head CI and publication.
