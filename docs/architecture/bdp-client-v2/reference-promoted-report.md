# Reviewed reference A/B coverage promotion

Frozen clean source **9dd30b8d5961a68250fa3f61ec257ed3d9c83e70**, tree **f14ca53d69c7d1ec9d5a44cc8bcf7be4cf1da5b3**, branch `codex/vickie-reference-integration`. Ordinary direct successor of `eecfbe56dba0695a490c0e719a5a084d9653548e`. No push, PR/bus mutation, runtime, manifest or fixture change.

Root explicitly accepted whole coverage for the exact selected catalog clauses after `reference-whole-coverage-audit.md`, `reference-target-pin-native-review.md`, and `reference-repair-final-dispositions.md`. The promotion is limited to `read-update.singleton.no-local-bindings` and `read-update.alias.non-canonical-target`. Unknown-subject C is unchanged and remains separate.

## Exact changes and preservation

Three files, 10 insertions/8 deletions: the two approved coverage rows change only `wholeContract` and `reason`; the independently maintained reviewed-ID test allowlist adds A/B in canonical order; four matrix expectations change 101→99. Every other coverage row and every assertion binding is byte-equivalent as parsed data. Top-level `claimEligible` remains false. Entire manifest and fixture bytes equal eec. Runtime/configuration and all other source files are unchanged.

Coverage is **30/129 whole**, **99 uncovered**, **49 authored rows**. A covers singleton @name carrier rejection in the admitted bare/pinned positions, no-effect afterstate and same-key corrected positives. B covers alias/Link/external target categories and both Link endpoint categories with exact diagnostic Problems, including independent admitted pinned source and pinned target plus bare endpoint controls. These claims cover the current selected clauses, not all nearby normative paragraphs or the full profile. No public write admission or conformance qualification follows.

Manifest stays 3,626,922 bytes / 97,116 nodes, digest `c86b33535f25a1d117ce021a3bb70164f5635bdaee50758e3428c130ccc3227f`. Fixture stays 26,370 bytes, digest `f5f38235dfd28cf06e760688e2054d3e29d8d64b08c4c729baa94f380f1f7a63`. Structural cap remains 100,000 nodes; headroom 2,884. See `preservation-audit.json`.

## Separate actual runs

| Evidence source | Installed four-file RU suite | Cumulative tests | Whole / uncovered |
| --- | --- | --- | --- |
| Prior target comparator repair `eecfbe56dba0695a490c0e719a5a084d9653548e` | 173 passed | 6 passed | 28 / 101 |
| New promotion `9dd30b8d5961a68250fa3f61ec257ed3d9c83e70` | 173 passed | 6 passed | 30 / 99 |

Prior evidence was read from `reference-target-pin/packed/`, preserved untouched and not rerun or relabeled. Its completion/source-after files identify clean eec. New evidence is in this report's `packed/`, generated with a copied original runner, a new frozen-head file, new tarballs and new isolated install. Both source records and their complete safe configuration bindings are separately represented in `summary.json`.

Pinned Node 24.16.0 / pnpm 11.20.0. Before commit: typecheck, clean build, 6 cumulative tests, boundaries, format, lint passed. On committed exact source: clean build, all static checks, 6 cumulative tests, packing/install and the four-file installed RU suite passed. No failures, skips or timeouts in the requested tests; no full 8k repeat for this evidence-only delta. No active hooks were bypassed.

Fresh new baseline has 49 passing authored scenarios and 85 explicit missing/unplanned harness-error scenarios, **985** completed raw serializer observations and **876** reported exchanges, plus 44 setup and 21 native Read observations. These are distinct layers, not additive counts. Focused A/B has 2 sessions, 82 raw observations, 89 reported exchanges, both pass. The positive reports contain no injected counterfactuals; actual negative-control tests are included in the test results.

## Exact new bindings

| Binding | Value |
| --- | --- |
| Exact configuration digest | `3224c31c83988dc0606cb76ee1f4ba8e6394159bdf7daaccc301461316cc2477` |
| Live configuration identity | `1720b63b49a7be3299d4b2865de4db9fcbe47560f217a32eea6397db7ea54ca5` |
| Installed payload | `78a80950acd6c1981d6829fc4d59c05eaafc65e0d68cf20d3d84f61f377acb99` |
| Installed artifact | `development-schema5-48667497d6e13064316d0364cf8f49dfc247ab9189306985a357d7b558333dfe` |
| Policy | `dbe8dbd5430f26e19e754d38167cd8084a32d702c695550e2ee0063dc2912430` |
| Authentication | `66941efc779f4f2431d65fdd5b2387b2af67bd7c1559e09dac032f82259ea16e` |
| Cumulative catalog | `e26d3fc758a34ebb0b5028be15d28e3e09db87dc7984bba7bad6a00b26329aa0` |
| Manifest | `c86b33535f25a1d117ce021a3bb70164f5635bdaee50758e3428c130ccc3227f` |
| Fixture | `f5f38235dfd28cf06e760688e2054d3e29d8d64b08c4c729baa94f380f1f7a63` |
| protocol.tgz SHA-256 | `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c` |
| server.tgz SHA-256 | `aa52a497f921dc00f4b66e94cf31843d0c91613019ddbb9292a9319d1ba8115b` |
| installed-reference-process.js SHA-256 | `3eb9a6bc8df1237d78ba4aae7e95589df6fb7f4781f572e04adda583f359c71d` |
| package.json SHA-256 | `d583f4158c3a393923c37d70a61212f6a716c951a5f3be7ed9bbe3aa8846ca2e` |

Safe `summary.json` lists five constituent catalog hashes, actual package bytes/hashes, method/URL route counts and separate predecessor bindings; no raw credentials. Canonical Scope remains `https://reference.invalid/acme/`, profile read-update, History false, fixture reference-read-update-schema-v1. The TX entry hash in package inventory does not imply a TX test run.

## Cleanup and limits

Every subprocess completed naturally under bounded process-group supervision. Final head/cleanliness and after-process inventory were checked; no owned installed child remains. Raw command/test logs, JSON results, stage exit/elapsed records and copied runner are retained outside the checkout. Unrelated processes were left alone. Source is frozen. Root owns exact-head CI, council disposition and publication. No existing evidence seal was edited or relabeled.
