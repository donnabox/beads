# P7 full independent native mechanics review

**Static mechanics accepted; no concrete findings. Execution remains CLOSED.** This review covers the complete other-authored P7 helpers and their data/source/provenance integration, not just the eight-array patch. No reviewed helper or schema was imported, evaluated or executed.

The reviewed [preparation index](/Users/dbox/repos/wt/vickie/bdp-integration-logs/tx-discovery-validation-preparation-7/preparation-index.json) is `ccb92eaccffa287acc18b392ac62405f6b6ce7d908af2a438e6d5be553fc6843`. Implementation index: `8d623d949f221cdedf52114a7a8bca5c3784cf2fdcab4ad971cbd9abbac90106`. Current source binding: `9bb16e9c984f7c259c6f4cb42ddad2886234771a90f4421c3fddec57c5079a41`; current schema: `c247f1b13bb475c9e29fe5c2d0f8b466ef0c086545c364bfa712a563caede6c9`.

## Exact data correction

The changed field-occurrence arrays follow the actual source's request assembly at666–670: explicit fields first, content-type when a body exists, then idempotency-key when supplied. `d3Command`1073–1076 supplies body and `cold-${id}`; the sequence direct call1377 supplies body without an outer key; replay1455–1458 reuses the original batch key; minimum1660–1662 supplies its fixed key/body. The report records those actual outgoing arrays at711–724. These are request occurrences, not inferred response headers.

| Stage / ordinal | ID | Exact expected occurrence sequence | Evidence |
| --- | --- | --- | --- |
| installed03 /7 | transactional-carriers-batch-seed | content-type JSON, idempotency-key cold-transactional-carriers-batch-seed | Measured failed report + source |
| installed03 /10 | sequence-turns | content-type JSON | Measured failed report + source |
| installed03 /13 | transactional-carriers-set-update | content-type JSON, idempotency-key cold-transactional-carriers-set-update | Measured failed report + source |
| installed03 /16 | transactional-carriers-set-delete | content-type JSON, idempotency-key cold-transactional-carriers-set-delete | Measured failed report + source |
| installed03 /19 | replay-batch | content-type JSON, original cold-transactional-carriers-batch-seed key | Measured failed report + source |
| installed06 /4 | minimum-seed | content-type JSON, idempotency-key d3-minimum-seed | Static; unattempted |
| installed08 /4 | minimum-seed | Same ordered pair | Static; unattempted |
| installed10 /4 | minimum-seed | Same ordered pair | Static; unattempted |

Here “JSON” is the exact primitive value `application/json`. All eight old arrays were empty. Replacing only these eight arrays with their old values reproduces **P6 schema bytes exactly**, SHA `b74e093741780341c0094aebb4d629582e9abd2532052bd15eab8051b5f256e0`. P6's complete installed02 table, including its prior eleven corrections, remains identical. Every other schema field, strict equality/order/type rule, header/current-tuple substitution, status/URL/body/lifecycle rule and selection stays unchanged. `audit.py:131` still requires full exact occurrence/order/value equality; no comparison is loosened.

The five measured arrays agree with the actual retained26-row failed installed03 report. This does not make that callback pass: it remains failed at Snapshot target confinement, ending with snapshot-root, and audit03 was not attempted. Exact body/manifest ID/token remains UNKNOWN; fingerprints are not inverted. The other three arrays are static predictions for unattempted callbacks, not observed runtime evidence.

## Full helper and pipeline assessment

I read all five complete current Python files (run80 lines, gate137, audit339, packages116, guard20) and the full current data/control contracts. All five are byte- and AST-identical to P6. The wrapped implementation manifest has a `files` mapping of1952 primitive SHA strings, matching `gate.py:101`; this is not a flat-map/schema mismatch. Eleven precise self-referential controls/deltas are excluded, while the complete preparation index binds every other path.

The52 stages and16 immediate test/audit pairs remain identical after inverse replacement of the exact P7 namespace with P6. All selected full names are escaped and anchored, files exactly mapped, maxWorkers1 unchanged; absent/duplicate/unselected executed callbacks fail the audit. Pure-data counts independently rederive21 source callbacks,20 construction baselines,96 started direct calls versus90 retained responses,14 raw wire responses and9 zero-wire clones. Eleven installed commands retain190 scheduled,187 completed,3 expected aborts,186 exact and4 raw requests; cold43/28, legacy owner12, legacy suites32/9, pure schema1/zero-wire and separate installed schema14 retain their distinct contracts. Counts remain prospective for P7, not new results or one pooled configuration.

The full audit still requires typed closed nested grammars, exact finite ordered IDs/methods/URLs/credentials/transports/occurrences, actual context values, source response versus raw-wire distinctions, exact legacy semantics and current external installed proof. It rejects unhandled runtime diagnostics, duplicate/missing reports and cleanup/terminal failures. Direct reader position0 remains a source projection, not a wire claim. Raw pending and Host witnesses retain their actual availability, ordering and lifecycle checks.

The unchanged runner records actual raw exit, group samples, seen PIDs, forced cleanup/errors/remaining processes, runtime directories and final source/build/consumer checks. Its sampled process observations are not a continuous ownership or universal file-descriptor guarantee. No D4 watchdog/marker/whole-clock model is imported into this separate lane. Existing120outer/TERM115/no automatic retry remains; root actual terminal/process acceptance is separate. Source default5s and authored3work/4inclusive cleanup remain; installed request5/work45/cleanup10/callback60 remain; inherited schema90work/120callback remains distinct. Root's explicit correction of the prior native prose budget shorthand is retained.

Fresh actual version/offline install/workspace-tool binding precedes static checks/build. Fresh build inventory/root acceptance precedes full package checks; tar membership/source bytes and only the known workspace metadata rewrite are checked. Fresh external consumer install, actual generated lock/manifest/settings root acceptance, complete installed member/production graph/realpath/payload/entry/inventory proof and root acceptance precede installed callbacks. Actual future tool/build/archive/consumer/proof identities are unresolved. Static candidate realpaths/hashes and current source lock were verified without invoking those tools.

The [full mechanics assessment](/Users/dbox/repos/wt/vickie/bdp-integration-logs/tx-discovery-validation-preparation7-native-mechanics-closure-1/full-mechanics-assessment.json) records fourteen areas, including provenance, lifecycle, bounded supervision, secrecy and qualification limits. No caller boolean, fabricated tuple, new wire semantic, public admission or whole Write credit is introduced. Raw evidence and secret-committing digests remain protected. Missing administrative evidence is not N/A; historical Read checks do not qualify Write.

## Current source versus history

All679 current source files were rehashed. The current canonical source identity is `a61cd3fbf9fac68e3aad857a24a2456dd16b83048067eef85348d7091c0aa08e`, with six owned paths and673 byte-equal cce base paths. Only transactional-discovery-installed.test.ts differs from old P6's9b7 candidate; the other678 prior candidate files remain exact. The one accepted call changes `d3Safe(manifest.id)` to `d3Safe(manifest.id, scope, true)` at1480. The helper's8192byte/origin/Scope-path/userinfo/fragment rules remain unchanged; this call permits the existing Snapshot query form and dispatches no extra request.

Root's actual Git receipt binds cce HEAD/tree675 and full six-file binary `eed91e5f1abc48ef92f31d92690772c8a6489a48dd9e5e07af66de9c392712ab`; this reviewer invoked no Git command. Source author268/bindingb8a, native0f1 and rootd310 form the pinned source-only chain. Original actual CODE read older bytes; it is never attributed to these future corrections. Unchanged audit-contract descriptive prose retains its historical9b7 binary reference; executable current binding/current-source-chain explicitly binda61. That old descriptive field is not used as a future source gate or current pass.

Historical preservation is exact and bounded:1312 prepared P6 paths,134 direct actual files/29 exits/2 archives, plus3 consumer metadata files and the separately retained original active-gate bytes. The original P6 preparation index still describes its original closed active gate1d02. Its copied consumed active gate738a is the single explicit exception; both versions are retained. No other prepared path differs. The29 exits independently show28 accepted stages and one failed installed03, with23 later unattempted; they remain old9b7 evidence. Earlier history remains inside the exact P6 prepared snapshot.

The old870-file/17-link consumer stays IN PLACE; only package.json, pnpm-lock.yaml and pnpm-workspace.yaml are copied. No consumer body, new consumer proof, complete fresh consumer revalidation or historical result reuse is claimed. No new run directory exists.

## Verification and disposition

Own bounded JSON/hash/AST-as-data/difflib work verified all1963 physical/1962 indexed files,1952 wrapped implementation hashes/exact11 exclusions,1858 copied-origin pairs,5817 pins and the full exact mirror. All38 declared forward/inverse pairs were recomputed and replayed:27 nonempty edits and11 explicit identity pairs. All current source/base/protected maps, candidate tool realpaths and history subset hashes passed. No helper/schema/control/product was imported/evaluated/executed. Two reviewer data-inspection shape/path mistakes are preserved as such, with no helper or product effect.

[verification.json](/Users/dbox/repos/wt/vickie/bdp-integration-logs/tx-discovery-validation-preparation7-native-mechanics-closure-1/verification.json) contains the exact checks; [findings.json](/Users/dbox/repos/wt/vickie/bdp-integration-logs/tx-discovery-validation-preparation7-native-mechanics-closure-1/findings.json) records zero findings. **This is static mechanics acceptance only.** Both gates remain false, approved stages and bootstrap reuse empty, all52 new stages fresh, and current source-bound root/native execution dispositions and measured proof barriers unresolved. Root must independently inspect, durably accept/release the exact bytes and control the actual attempt. This review changes no packet, gate, source, review history or execution authority.
