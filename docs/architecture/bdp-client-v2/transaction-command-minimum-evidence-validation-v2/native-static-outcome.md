# Native audit of completed V2 static outcomes

No blocker found in the retained nine-stage static evidence. The exact30 selected path/full-name pairs all passed on reviewed ten-file source fc4fd747379917531c4d64d3fe52603337e72ec3f78e09aedcab8e9a8cc0b4b7. This audit is read-only reconciliation, not a rerun and not installed/profile qualification.

Source manifest612cbb471df504b3a0440e11f15a29911cf59aeeae89247836f04d5b19dc580b; HEAD453b6dab386746bb03aac2c0a293572b8a71c969/tree40782d97c9764bc19ae78c46d1d39475e3db317a. Independently rehashed all10 source files and the full working diff; all match. Every stage start/exit remains bound to that source. The original five files retain their ea04 bytes.

## Completed stages

All stages have actual exit0/supervisorExit0, sourceUnchanged:true, no timeout, forced cleanup, recorded errors, retained sampled PID or owned runtime directory. Each was bounded by120seconds, with no repeated stage record.

| Stage | Actual elapsed seconds | Result |
|---|---:|---|
| offline-setup | 0.442 | exit0 |
| typecheck | 1.069 | exit0 |
| build | 1.069 | exit0 |
| format-check | 0.572 | exit0 |
| lint | 0.657 | exit0 |
| boundaries | 0.744 | exit0 |
| read-seal | 0.544 | exit0 |
| static-cases | 1.823 | exit0 |
| build-inventory | 0.238 | exit0 |

Actual argv use pinned Node24.16.0/pnpm11.20.0. Offline setup used --offline --frozen-lockfile. The build log explicitly runs clean then tsc -b; typecheck is tsc -p tsconfig.check.json. Boundaries reports11 workspace packages valid. Read seal verifies the historical89-row Read artifact; it is not a new Transactional observation.

Format checked439 files with no fixes and one configured-size warning for the3.5MiB RU matrix. Lint checked439 files with no fixes,769 warnings and48 infos (diagnostic display capped), exit0. This is successful configured validation, not a warning-free claim.

## Exact test outcome

The JSON and verbose log agree:30 passed,0 failed,601 skipped, six test files; no reported unhandled errors. JSON nested-suite counters are not a count of test files. The30 actual non-skipped (path,fullName,status) triples equal the reviewed selection exactly, with no missing or extra executed callback. Distribution: executable-manifest7, http-executor12, runner3, cumulative5, profile-contract-plan1, app2. The app callbacks actually ran in V2.

The real wire regression passed its own exact received head/body, received=1, resolver-once and complete-write assertions. That provides one loopback exchange/listener witness. Other selected callbacks are CPU/mock/configuration paths, with zero installed/product children or stores by inspected source. No independent network trace or global child census is implied. The fixture's tracked listener/socket cleanup completed without a hook failure.

- static-cases.results.json SHA256 e99ebc85b4be2ca379d3cb3038cf90ffadb1468b6b1e766e690121fc814d9599
- static-cases.exit.json SHA256 a986b8f0924d7fe038c5af2f9f047a4257bb9387fe73bd5b6d6c3b3ea3ce3990
- reviewed selection.json SHA25617859e4aff9ab0863077df90dab022dfeeef9a80f6625a113398d47b68ac8865

## Build inventory and cleanup

Independently enumerated the inventory scope and rehashed every file, rather than executing build-inventory.py:312 files,2413225bytes, no symlink entries, exact path/hash equality. Scope is LICENSE; protocol package.json,60dist files and one schema; server package.json and248dist files. This is a package-input inventory, not a claim to inventory every workspace output or a fresh tar/consumer provenance claim. Inventory source hashes and its30 staticSelected triples also match.

build-output-inventory.json SHA2562b872047cac1cacaf9ec5bbf5cbbe082e4a639418c008ff2a83bcf2fece8e8ff.
Separately read, without importing, current emitted conformance classifiers: executable-manifest.js SHA25662e113f98f7df97f87685d792ac045a25f2cb56c3dc2579429832e0fbc75976d; http-executor.js SHA25625c702652cadf6be40f915d9190a5a92734fd21d8dc3112c570174a6082ea82c. Both contain the reviewed exact-name predicate.

Independent ps snapshot at2026-10-15T07:29:23.385465Z found none of29 retained sampled PIDs and no members of nine recorded process groups24572,24679,24754,24878,24941,25001,25144,25250,25332. This supplements retained supervisor cleanup records; short-lived or detached unsampled descendants remain the sampling limitation, not an asserted impossibility.

Independent current directory listing: build-inventory.tmp empty; the other eight stage TMP roots contain only node-compile-cache. No owned runtime roots remain. Cache directories are deliberately retained and classified as build/runtime caches, not stores or leaks; nothing was deleted by this reviewer.

## Preserved predecessor failure and limits

The original validation directory still contains the failed seven-selection attempt with1pass/4fail/94skipped and app0registered results. Its exit1 and original five-file source attribution remain intact:
- predecessor static-cases.exit.json SHA2564ad0910028cc5f56cf9180002bdfba4e3f0d65401d07ec20bfc3a492b290a611
- predecessor static-cases.results.json SHA256c38e3a3ef8a477b250a69979dcad57f2d91ad45d742d004769a4571e74f32376

V2 is a repaired-source validation, not an unchanged retry or retroactive pass for that failure. This audit does not claim the three installed cohorts, packed payload, exact-head CI, Claude review, merge or public qualification. Root owns those remaining gates. No tests/build/services/imports/source changes were performed by this reviewer.

## Audited executed path/full-name pairs

- apps/bdptest/src/transactional-matrix.test.ts — pins deletion-safety derivative carriers, complete positives, and measured source bounds: passed
- apps/bdptest/src/transactional-matrix.test.ts — pins the matching no-op installed route digest and fifteen-action ledger: passed
- packages/conformance/src/executable-manifest.test.ts — executable scenario manifest admits normalized authorization-view context without changing lowercase rules: passed
- packages/conformance/src/executable-manifest.test.ts — executable scenario manifest rejects authority-changing paths, secret-like headers, and control values: passed
- packages/conformance/src/executable-manifest.test.ts — executable scenario manifest rejects credentials and noncanonical header names: passed
- packages/conformance/src/executable-manifest.test.ts — manifest v2 exact HTTP inputs admits literal authorization-view context through actions=false: passed
- packages/conformance/src/executable-manifest.test.ts — manifest v2 exact HTTP inputs admits literal authorization-view context through actions=true: passed
- packages/conformance/src/executable-manifest.test.ts — manifest v2 exact HTTP inputs rejects closed-form violations in actions=false: passed
- packages/conformance/src/executable-manifest.test.ts — manifest v2 exact HTTP inputs rejects closed-form violations in actions=true: passed
- packages/conformance/src/http-executor.test.ts — exact HTTP request execution keeps authorization-view lookalikes protected in parser and serializer: passed
- packages/conformance/src/http-executor.test.ts — exact HTTP request execution rejects authored protected field AUTHORIZATION before resolving credentials: passed
- packages/conformance/src/http-executor.test.ts — exact HTTP request execution rejects authored protected field Authorization before resolving credentials: passed
- packages/conformance/src/http-executor.test.ts — exact HTTP request execution rejects authored protected field Content-Length before resolving credentials: passed
- packages/conformance/src/http-executor.test.ts — exact HTTP request execution rejects authored protected field Cookie before resolving credentials: passed
- packages/conformance/src/http-executor.test.ts — exact HTTP request execution rejects authored protected field Expect before resolving credentials: passed
- packages/conformance/src/http-executor.test.ts — exact HTTP request execution rejects authored protected field Host before resolving credentials: passed
- packages/conformance/src/http-executor.test.ts — exact HTTP request execution rejects authored protected field Proxy-Connection before resolving credentials: passed
- packages/conformance/src/http-executor.test.ts — exact HTTP request execution rejects authored protected field Transfer-Encoding before resolving credentials: passed
- packages/conformance/src/http-executor.test.ts — exact HTTP request execution rejects authored protected field X-Auth-Token before resolving credentials: passed
- packages/conformance/src/http-executor.test.ts — exact HTTP request execution rejects authorization-view credential handles before resolution or dialing: passed
- packages/conformance/src/http-executor.test.ts — exact HTTP request execution writes authorization-view context separately from credentials: passed
- packages/conformance/src/profile-contract-plan.test.ts — canonical cumulative profile contracts inherits partial TX command-minimum evidence while preserving RU omission: passed
- packages/conformance/src/runner.test.ts — black-box conformance runner does not reflect credential-shaped response header values or names: passed
- packages/conformance/src/runner.test.ts — version 2 exact HTTP receiving accepts registered credential decoration and authored defaults without resolving credentials again: passed
- packages/conformance/src/runner.test.ts — version 2 exact HTTP receiving keeps admitted authorization-view values out of report projections: passed
- packages/conformance/src/transactional-cumulative.test.ts — binds the canonical cumulative TX selection and exact source digests without qualification: passed
- packages/conformance/src/transactional-cumulative.test.ts — copies source bytes before binding and exposes all missing obligations without executing them: passed
- packages/conformance/src/transactional-cumulative.test.ts — pins the inherited command-minimum manifest facts, read-header absence and all thirteen prior objects: passed
- packages/conformance/src/transactional-cumulative.test.ts — preserves every original TX scenario and coverage object while adding one bounded partial cross-carrier row: passed
- packages/conformance/src/transactional-cumulative.test.ts — rejects noncanonical selections, History, wrong factories, retired and misprofiled rows: passed
