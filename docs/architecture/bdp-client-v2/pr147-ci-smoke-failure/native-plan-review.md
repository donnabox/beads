# Native plan review: PR147 packed-server smoke inventory

**Approve the bounded four-entry plan. No blocker found.** Add exactly json-value-equality.d.ts, json-value-equality.d.ts.map, json-value-equality.js and json-value-equality.js.map to expectedServerOutput in scripts/smoke-executables.mjs. Preserve the exact installed-output equality check, test/test-support exclusion, package export object fence, consumer resolution checks and every later smoke observation. No runtime, build configuration, package export or module-removal change is justified.

## Diagnosis and source evidence

The saved CI log at exact ddfb5f43 records pnpm smoke reaching installedServerOutput after pack/install, then rejecting missing=[] and precisely this four-file unexpected set at job-full.log:6547. This is a real packed-consumer inventory failure, not a TypeScript compilation failure or missing runtime dependency. The root reports the full tests passed in CI38950645644; this review does not claim to have rerun them. I verified the saved log hash and inspected the clean exact-head checkout.

json-value-equality.ts is intended runtime source. resource-evaluator.ts:1 and transaction-snapshot.ts:1 directly import its jsonValueEqual implementation through ./json-value-equality.js. Removing/excluding it would break the emitted internal dependency. packages/server/tsconfig.json includes src/**/*.ts and excludes src/**/*.test.ts; shared tsconfig.base.json enables declaration, declarationMap and sourceMap, so each included TS source emits precisely the expected JS, JS map, declaration and declaration map. packages/server/package.json includes dist in packed files. CI independently observed those four actual packed files.

A standard-library-only static inventory calculation derives 256 filenames from the 64 non-test server TS modules using those four configured suffixes. The current expectedServerOutput has 252 names, no extraneous names, and differs by exactly this quartet. Adding exactly four entries makes the explicitly authored inventory agree; do not replace it with a dynamic source-derived whitelist in production smoke. This derivation is a diagnosis check, not a newly executed build or smoke pass.

## Parallel inventories and exports

Repository searches found one explicit complete packed-server filename inventory: expectedServerOutput. No second hard-coded module list needs this addition. installed-payload.ts enumerates actual installed dist runtime bytes dynamically; it will include json-value-equality.js without an inventory edit and intentionally ignores declaration/map files as non-runtime data. The build/test artifact and package-root files fences remain unchanged.

The export map remains exactly '.' and './development', checked by assertServerExports in scripts/smoke-safety.mjs using deep equality. Neither index.ts nor development.ts names/re-exports json-value-equality/jsonValueEqual. The existing transaction-snapshot.test.ts:674–680 explicitly asserts this private module is absent from those public source entry points and the export map has only the two approved entries. Therefore acknowledging an existing internal packed module does not create a public package subpath or root export.

The smoke loop dynamically probing transaction-* and installed-schema-* paths does not itself enumerate json-value-equality, and its explicit private-specifier list also omits that name. This is not a blocking parallel-inventory omission: the exact export-object equality fence forbids any new direct/dist/wildcard subpath for it, and the existing source export test covers its root/development non-reexport. Adding extra probes would be optional defense-in-depth, not required for this four-entry repair. Direct filesystem access to shipped implementation bytes is distinct from a supported public package export; this plan changes neither.

## Recommended bounded verification after authoring/review release

1. Review the exact successor diff: only four string entries in expectedServerOutput; unchanged exact set comparison and export/assertion/cleanup code. Bind the successor source hashes and HEAD, retaining CI38950645644 as prior evidence rather than attributing its tests to new source.
2. Ensure all dist artifacts used by smoke have a fresh or positively verified source binding for this successor/runtime baseline. pnpm smoke does not build; it packs existing dist. If fresh build is needed, root should authorize it as a separately bounded prerequisite, not silently reuse unknown artifacts.
3. Run the existing full pnpm smoke command once through the existing reviewed 120-second isolated-process-group supervisor pattern using pinned Node24.16/pnpm11.20. Give it a fresh output/temp directory, source-before/after pins and explicit npm_execpath via pnpm. Preserve built-in30-second command limits and all later checks. No retry, broad suppression, whitelist wildcard or partial extraction of only the repaired set check.
4. Require exit0, unchanged sources, actual successful completion of the later package/export/CLI/installed-reference smoke path, and no surviving owned processes or forced cleanup. Any timeout/intervention is failure even if the quartet check passed. Preserve the actual failure trace and investigate a distinct next failure without relaxing the original boundaries. Use only isolated smoke-owned temporary paths; do not supply Janet/shared external workspaces or promote/release anything.

The quartet-only edit does not justify rerunning the whole multi-minute test suite locally; the targeted complete smoke is the relevant integration check, with exact-head CI still required by root before branch integration/publication. This report is plan review, not validation release.

## Limits

Read-only diagnosis/native review. The captured CI failure supplied the existing red signal under the diagnosing-bugs workflow; the assigned no-test/no-build constraint intentionally defers local reproduction/green verification to root's released validation. No source edit, import, test, build, install, process launch, commit, push or network publication. Event and conditional frozen source/review packets were not touched. Only this guarded external report is written.

## Input hashes

- /Users/dbox/repos/wt/vickie/bdp-integration-logs/pr147-ci-smoke-failure/job-full.log: `dad60bb785b95bbdc2fe9f41dffd1ef3a735849219203ed89c663db33c217aa0`
- /Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving/scripts/smoke-executables.mjs: `96701a2e4a7535e44ad4490a109d8e4c93925b4702b5453f3c55abe689a88060`
- /Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving/scripts/smoke-safety.mjs: `8aa35543d3236187aa377ad01d2255ecc6abcab3a81665aa6498baff2bf73803`
- /Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving/packages/server/package.json: `110e00c403c667d8c848c6280b2e98c47f98f134158a621967d889914d44632e`
- /Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving/packages/server/tsconfig.json: `da054bc9239dfe20febb2fd9255e4263781627499af7eccc3c40a5c80526b1da`
- /Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving/tsconfig.base.json: `6c70be275f9f7d5de533cbd3f2454ae7cb84bc765ab84f4f199889580e18e152`
- /Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving/packages/server/src/json-value-equality.ts: `f93e0bc3a090eeda9d40ec873c0024da2fbc368df5babe4f7d18cfa9b4780a66`
- /Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving/packages/server/src/resource-evaluator.ts: `6784629841a92baaeed648862cfa84d071356ef82e6d29250b254a9eddf793a2`
- /Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving/packages/server/src/transaction-snapshot.ts: `dac7c2f121cfb6bae557b3917f5ec52e47a358ab39da4062f6a5c26f5b9fbd83`
- /Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving/packages/server/src/transaction-snapshot.test.ts: `b6c31d8b66650ec0668c85d1fb803dda1072ef463989299ed377dd18dcbfee98`
- /Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving/packages/server/src/installed-payload.ts: `d1095a8fa641c77ddef32b1d432273572f46b14499c526dbb0a85cfde82113f9`
- /Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving/packages/server/src/index.ts: `34f9d9de8b48852c64919e69ce93f149e9cdae5142c15c75e4c35f08f15efc44`
- /Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving/packages/server/src/development.ts: `a3f994bd610f17821c7d51189104308abef6196bb12b565b268a0ffc96b04f39`
