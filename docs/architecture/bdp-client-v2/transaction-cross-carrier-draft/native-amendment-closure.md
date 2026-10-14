# Native closure — staged guard and report amendment

**No findings.** The staged successor is exactly the root-accepted measured guard amendment plus the required test-local full-result writer. It closes native N2 at the source-design level. It has not been applied, compiled or executed; root still owns application after the external review and the later execution gate.

Reviewed `/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-cross-carrier-draft/guard-and-report-amendment.patch`, SHA256 `83cf3edd53e4c584afe159ccb43f96c836e3bcbd7a56b080fa8070528bac5122`, its binding JSON SHA256 `c21be575f499d26b3af21864090d3320b59c55f9a3b9f12c7777a4069859dd42`, complete before/after copies, and `root-measured-disposition.md` SHA256 `b28db1b7ca095daf685a32483a8f62b1b8f10929f58ba61d9602e58905ea652a`. Basis remains uncommitted four-file measurement draft on parent `eff171839fe4f37638284a4c3346428979d26b5d` in `/Users/dbox/repos/wt/vickie/vickie-cross-carrier-draft`.

## Independent exact-diff checks

Recomputed all binding hashes. Both staged before files equal the actual frozen checkout bytes. Reconstructed the complete unified diff from staged before/after text and compared it byte-for-byte with the retained patch. Staging has exactly those four copies, representing two amended source files: cumulative test15 additions/8 deletions, app test12 additions/5 deletions, total27 additions/13 deletions. Reverified all four original checkout source hashes remain unchanged. No repository edit or patch application occurred.

| Staged after file | Bytes | SHA256 |
| --- | ---: | --- |
| packages/conformance/src/transactional-cumulative.test.ts |89341|cfb0cbb6d534d813b7b4364d86ba805319d82d3499a567c2bc0f4de9060c9972|
| apps/bdptest/src/transactional-matrix.test.ts |83085|08d64942a5e9014d64b15a404f44ef77337cd576ed302c0391dd4dc22822c35f|

Manifest SHA256 remains `726fc623bbb115a0f39e0bb3aaa586ea37254522b254e55e555ea95587f14cb9`; coverage remains `a3a15b5143c118b24965698ec10289821664c90e935820b51c8cf76c05bdb1df`. This closure does not replace the original source identity with staged after hashes.

## Guard closure

All proposed numbers match the root disposition and accepted v3 joint ceilings:

| Form | Measured | Staged ceiling | Margin | v3 ceiling |
| --- | ---: | ---: | ---: | ---: |
| Canonical manifest disk bytes |684206|688302|4096|720094|
| Canonical manifest normal stringify bytes |695598|699694|4096|730676|
| Canonical nodes |18107|18171|64|18973|
| Coverage disk and normal stringify bytes |99245|100269|1024|106043|
| Maximum derivative normal stringify bytes |701716|705812|4096|736794|
| Maximum derivative nodes |18263|18327|64|19129|

Delta arithmetic is exact:488280+200022=688302;70662+29607=100269;13019+5152=18171. The cumulative guard at staged lines908–936 changes both delta/absolute pairs and independently adds `JSON.stringify(parsedValue, null, 2) + "\n"` byte fences for the manifest and coverage. Manifest disk bytes, normal bytes and compact execution bytes remain distinct. Both app derivative locations use705812/18327, on existing normal-form serialization and actual node walks. No stale smaller active duplicate remains at the identified guard sites; historical comments preserve the predecessor measurements.

Exactly199 canonical actions and2102 assertions remain pinned; no authored request, capture, success assertion, control transformation, named failure set or row/prefix expectation changes. N1's preserved-original supplement already accounts for38 mutated manifests+14 unchanged invocations; its extra invalid-input derivative is below the same maximum. The amendment neither adds cases nor claims new measurements.

## Full-result writer closure

The only added non-guard executable statements are the optional writer at the new control tail. It follows the existing no-op/deletion pattern exactly: if `BDP_TRANSACTIONAL_CROSS_CARRIER_CONTROL_OUTPUT` is set, `writeFileSync` writes `<prefix>-<control>.json` containing `JSON.stringify({ sourceCommit, entry, ...observed }, null, 2) + "\n"`. The existing import is reused. The closed control-name union gives distinct batch-semantic-conflict and sequence-fresh-key files.

`observed` is the genuine returned cohort result already used by the assertions. The writer retains its baseline, session and complete exchange/client assertion records together with sourceCommit and entry; it does not synthesize values, re-fetch, add a capture option, alter transport or launch another child. It sits after the existing complete assertion checks and console summary. Consequently, a thrown earlier assertion can still leave no full report, just as with the predecessor writers; the existing execution plan correctly requires retaining available logs and stopping rather than inventing a result. A write failure remains a failed test. This staged change closes the missing successful-control report path, not every possible failure-report path.

Before installed execution the external supervisor must validate/preserve this exact prefix under its owned evidence directory and bind both output paths. No source change beyond this staged writer is implied. Full reports remain ordinary test output, not child-attested source or qualification.

## Preserved boundary

The diff does not change any runtime loader/body/header/node limit, request/cohort/test deadline, setup/cleanup logic, test count, selected-case workload, fixture, catalog, schema, public API, installed payload or CI allowance. Existing runtime and other four-file draft semantics remain those reviewed in native-measured-preflight.md. Final after bytes still require authorized application/source binding, normal static checks and the separately bounded validation; no execution success is inferred here.

Only read-only file inspection, hash/diff reconstruction and arithmetic were performed. No tests, typecheck, build, install, pack, service, network action, application of the patch or repository mutation occurred. This report is the only review artifact written.
