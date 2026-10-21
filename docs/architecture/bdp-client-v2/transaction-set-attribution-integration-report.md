# Transactional set / RU attribution integration validation

Frozen clean source `10517d59b431df4772897ae015bef54e1828fb63`, tree `a8397bbe945d2b18152c90060304aad84ac7891e` in `/Users/dbox/repos/wt/vickie/vickie-transaction-set-attribution-integration`, branch `codex/vickie-transaction-set-attribution-integration`.

Ordinary parents: `3964b18dafcbbc3598c5ab21c550b73ccb0746cf` and actual fetched remote main `2a40740cb98545f4e83166c8e58e2aee0e587a65` (remote tree `d779f7f321bc1244dc09528eba4f2c4b0c3c35af`). No conflicts or authored resolution. All 11 candidate-delta files byte-match3964; all five main-only files byte-match main. Per-file blob proof is in source.json. Both source workspaces remain untouched. No source, runtime, artifact, policy, capability or profile edits were made beyond the ordinary merge.

## Completed validation

Pinned Node24.16.0 / pnpm11.20.0. Frozen install, typecheck, clean build, dependency boundaries, format, lint, Read evidence verification and both smokes all exited0. Relevant server Transactional, app association and conformance tests: **1,085 passed**. Supplemental catalog controls: **21 passed across three actual files** (numeric-model, Read+Update, Transactional). The runner also supplied a nonexistent read-catalog.test.ts filter, which selected no fourth file; no extra Read-catalog coverage is claimed. Read verifier preserved the existing89-row evidence and constant `674982be2123b9b7aae630bf79c3f774122900cd`; it did not generate or relabel a cohort. Explicit commands live in run.py and packed/validate.py; exit/duration records in results.json and packed/stages.json. All substantive commands exited0.

Fresh protocol/server tarballs were built from this clean exact source, installed into a new isolated directory using npm --ignore-scripts, then exercised through actual installed process entrypoints:

- Set installed test: **1 passed**, 33 receipt entries over pages32+1, 96 Events over32+32+32, one change group, five negative oracles.
- Existing installed Transactional cohort: **35 passed in five files**, including finite receipts/point/Events/Snapshot and both actual15-second SSE tests. Wall33.086s. The cumulative result itself remains four passing planned rows plus251 missing-plan harness-error rows; this is no full-profile claim.
- RU aggregate plus attribution/numeric positives, all16 countercontrols and static source check: **20 passed**,154 other tests intentionally filtered, no failure; wall30.578s. The complete genuine two-shard aggregate completed in18.297s of the unchanged90-second budget, observing52 pass /77 not-implemented. All129 required mappings resolve to actual result rows. Both wrapper-byte hashes, canonical result hashes/lengths and cleanup/source fields independently verified (packed/independent-report-audit.json). Whole coverage remains30/129,99 uncovered; new attribution/numeric rows remain partial.
- Attribution row:6 raw exchanges +69 client actions,663 assertions. Numeric row:15 raw +58 client actions,533 assertions,10 real raw mutations with complete serializer write state and withheld raw body digest. Numeric largest raw body526bytes. Current sanitized per-client observations do not expose bodyOctets; no client-byte maximum is inferred.

Full8k workspace suite was deliberately not repeated. This is the requested affected/cross-profile validation, with no evidence regeneration, broad RU-negative repeat or public eligibility promotion.

## Actual installed binding

Both profiles observed payload `1b12629fe54c39c8f7864dbb13bdba16bcbe4298b0a5bd420fa19daaa4fb7ec2`. Protocol tarball SHA256 `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`; server `80f0ab8fa061480be11428b72b110b0f2f8185d467f561f6d30912f2e268562c`. These bytes also equal3964's previously packed runtime, but all observations here were freshly run and bound to10517d5; predecessor evidence is not relabeled.

RU contracts `48667497d6e13064316d0364cf8f49dfc247ab9189306985a357d7b558333dfe`, fixture `07bbe646a5ce8436e8a746fadd22a9906013f0c28e28f608dcd9a0a84aa7b234`, runtime configuration `1720b63b49a7be3299d4b2865de4db9fcbe47560f217a32eea6397db7ea54ca5`, set `5c40fa9f8b6b3fedf70e37a24ceb82315cc27ee1e9fa1aeefea0c591befc5e00`, coverage `46d55f94e2726f50de91f857bfb7dee5f09f6b6bb08281ecece669b707aeced9`. Scope is `https://reference.invalid/acme/`, profile read-update, history:false. Actual bootstrap and per-session policy/authentication/configuration hashes, routes, manifest/fixture/catalog hashes and exact source sizes are in safe summary.json. RU baseline route digest `3224c31c83988dc0606cb76ee1f4ba8e6394159bdf7daaccc301461316cc2477` and continuation `6713ab2c41b09869b664f372de1ef1390c00cfd24b154ea717f7ee6d62c35059`. The main-added fixture and attribution/numeric observations were exercised against the changed Transactional runtime payload.

TX cumulative runtime configuration `06b14416526dbbcdc63e922ac9afc9ca9fb65bf1a56c19145e20ebfa634c4640`, contracts `c67d12264b97e16c76bf68f850eaf394f61896bb72491e06603ef1d6debb335b`. Its separate set fixture binding is retained in summary.json; these are separate actual configurations, not conflated. No credentials or raw body bytes are copied into the safe summary.

## Cleanup, limitations and review scope

Source remained clean at the exact frozen commit before/after every stage and packing. No test-owned Node children or pending report files remain; cleanup-processes.txt and summary.json preserve the check. No timeout or test failure occurred. An optional summary extraction initially assumed bodyOctets on sanitized client observations and raised KeyError; it was corrected to record absence/null, without rerunning tests or changing source.

Parent reports external3964 runtime closure completed without a runtime regression and is handling a separate Low docs-only supersession follow-up. That future source is not this observed head. Integration validation does not extend predecessor council coverage or claim full Read+Update/Transactional admission. No push, PR, bus, launch, main, shared database, Types or CLI mutation was performed.
