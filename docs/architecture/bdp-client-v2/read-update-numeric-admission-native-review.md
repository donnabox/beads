# Native review: numeric admission/location evidence

## Findings

**No actionable findings in the frozen five-file candidate.** Reviewed `131fb9874ae53f94ea7c15d5b77bd6caa4a704e9`, direct parent `2a40740cb98545f4e83166c8e58e2aee0e587a65`, against the shared council prompt and accepted plan `0af9d4d447b0a571d702c960de4e48ff2925074a4af89a535b0077dc1c285a8e`. Both new rows remain partial/wholeContract:false; existing30 whole decisions and public claim boundary are unchanged.

I inspected the full five-file diff, using structured action/assertion inspection and an independent exact-decimal state audit for the large manifest/coverage JSON. Source author checkout `/Users/dbox/repos/wt/vickie/vickie-read-update-numeric-admission` stayed immutable. Tests/build ran only in detached isolated `/Users/dbox/repos/wt/vickie/vickie-read-update-numeric-admission-native`.

## Evidence correctness

- **Literals and phase discrimination:** R submits both Resource kinds through whole creates and properties patches using exact UTF8 integer9007199254740993, precision0.12345678901234567891 and overflow1e309. Corrected creates use…992/0.125/1e300; corrected patches use…994/0.25/1e299, so properties really change while companion metadata changes. Admitted1.0/-0.0/exponent and301-digit expanded values remain authored raw strings. No JSON number parsing is used to construct the wire exchange. Existing raw serializer body/write assertions and observed source/write metadata remain intact.
- **State oracles:** Every refused create has an independent fixed-ID404; every refused patch has full-record equality to a previously independently checked client GET capture. Those captures include literal identity/Type/documents, no provenance/owned-plane and endpoints where applicable; opaque revision is retained in the full record. Successful raw replies and later GETs independently pin intended values. Source/target records are checked after each Link refusal and finally. No failing response defines its own afterstate oracle, no raw object capture is introduced and no response agreement substitutes for literal postimages.
- **Diagnostics:** N's two whole-properties errors identify shallow and16-level escaped-key locations; unapplied patches report original operation-input pointers in messages without claiming a postimage instanceLocation. Metadata whole/patch cases use the configured metadata-prefixed message/document label; Type/schemaLocation blame is absent. Exact Problem comparisons cover the configured fixture choice without implying optional fields must be absent on every implementation. Numeric metadata refusal remains the explicitly disclosed general-model inference. Both finite rows remain false rather than claiming arbitrary depth, every number, sequence/replay, mutable policy or universal diagnostics coverage.
- **Sensitivity controls:** Six raw controls target exact scenario/key/method/operation route; request-number changes one actual token before transport, while response substitutions follow genuine exchange and recompute bodyOctets. Three client controls use the existing executed-input-equality seam and act only after observation. Each requires one application and exact named failed assertions. The request-token control naturally can cause later state-dependent row failures; its named exchange requires status/authored-body/body-schema/complete-problem, and no whole-row singleton-failure claim is made. All nine ran independently and passed their intended discrimination checks.
- **Closed catalog ownership:** The local numeric catalog test permits only the approved three numeric IDs in the continuation, with read-update-v1/manifestVersion2 required. Read companion ownership and standalone numeric-catalog executable denial remain. Each actual new row injected under an unapproved filename or into intact Read is rejected on that exact row ID; wrong catalog/version under the approved filename is also rejected. No loader, normative catalog, protocol/schema or runtime semantics change.

## Preservation, source audit and bounds

The independent out-of-tree `read-update-numeric-admission-native/audit.py` passed. It checks exact prior A/C/B raw blocks; baseline/index bytes; every nonselected coverage row; preservation of prior fixture bindings/values; unique new keys; exact ordered coverage bindings/false flags; full modeled literal postimages, original-input diagnostics and capture-based unchanged state across every new action. Its Decimal-based finite corpus walk preserves the bad literals rather than rounding through a host JSON float. The model does not invent opaque revisions or claim to prove the target runtime; it checks the independence/correctness of the authored state oracle.270 concrete literal/state assertions were compared, across20 expected numeric refusals and14 modeled Resources.

Measured source matches the author ledger exactly:

| Item | Actual |
| --- | --- |
| R |72actions,521assertions,29raw/43client,24raw POSTs,8Resources,max raw body566bytes |
| N |52actions,373assertions,21raw/31client,16raw POSTs,6Resources,max raw body373bytes |
| New keys |44 unique mutation keys, including4 endpoint setup creates |
| Continuation |823730bytes/21614nodes; +288678/+7328 within reviewed450KiB/12000 growth and27158 historical total-node ceiling |
| Coverage |698614bytes/14394nodes; +50457bytes within65KiB growth |
| Fixture |29707bytes/764nodes;14 new fixed bindings/+652bytes, no prior value changed |
| Inventory |129required,54authored,30whole,99without-whole,75omitted |

New whole-file source-budget pins use accepted9c anchors535052/14286/648157/29055. Old A75/B73/148 and16 controls remain separate from R72/N52/124 and9 controls; no runtime/parser/deadline cap increased. Continuation source SHA256 `923443a20223cd84abc144f3dce4858fd0b5d18f2cea4a23a9a61a274a236267`; coverage `14d856bc135c0c93f6ee74741530a38fb7ad5c1bd523bc97438bdf5714817ac3`; fixture `f89932d68e3bf5867d79eb3de680af4bc67b761b3c3bce5da127f6ca75e27765`. Baseline49 and unchanged index remain exact parent bytes. The explicit49-route ordered inventory accounts for raw create routes moving ahead of mandatory fallback routes; it does not merely change a count or retain the obsolete digest.

## Independent executed checks

Pinned Node24.16.0/pnpm11.20.0. Frozen install and clean build: exit0. Then:

```text
pnpm exec vitest run apps/bdptest/src/read-update-matrix.test.ts packages/conformance/src/numeric-model-catalog.test.ts --testNamePattern='numeric admission|draft numeric-model catalog' --maxWorkers=2 --reporter=verbose
```

**19 tests pass /2 files**, exit0,7.74s.174 other tests in the selected files were filtered out (not passed or optional skips). This covers two actual emitted-server positive rows, all nine negative controls, source ledger and seven numeric catalog checks. Source audit exits0 and git diff check exits0. Logs are `read-update-numeric-admission-native/{install.log,build.log,focused.log,audit.py,audit-result.json}`. No independent full suite or duplicate packed run was requested or performed.

## Completed author packed evidence, separately inspected

While review ran, the exact frozen packed cohort completed. I inspected its actual reports/exits and independently rehashed both written shard reports and both tarballs; this is verification of author artifacts, not a claim I executed that cohort.

- Author five-file packed invocation:254tests/5files,0fail/0skip, exit0,143.635s (240s supervisor preserved); recorded descendant-tree peak1338384KiB is not a per-owner memory bound. Build/pack/install source logs remain separate.
- `packed/source.json` identifies clean exact131fb987 and actual installed `node_modules/@bdp/server/dist/installed-reference-process.js`. Tarball hashes match protocol `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c` and server `e9fd483e3f8ba2c4fdf0f1a64ce1b94a7990d3fb5f69bdeb5b998132e37b1e7b`.
- Aggregate records sourceDirty:false, claimEligible:false, setDigest `050018ff135dce188b22dfbbb98cdaff8e991a71894f432e663ffc88c5665938`, exact source-matching coverage/fixture/index digests, runtime configuration `1720b63b49a7be3299d4b2865de4db9fcbe47560f217a32eea6397db7ea54ca5`. Baseline49 and continuation5 both freshly pass, complete within90s aggregate (last recorded elapsed19.836s).
- Actual written baseline report hash `d85c6e3b967b870b52fc12e18b20605248ab129ddc008f61059dcb87b40ae6e1`/4725650bytes and continuation report hash `88a976c94559462e05de9fd7ee78c8acb2f6479b9f94578571d4dc988e8f3aee`/940760bytes independently match the files.
- Separate R/N reports each name exact clean131fb987 and the installed entry, claimEligible:false/cleanupComplete:true. R has29raw+43client/521 passing assertions; N21+31/373. I checked every recorded assertion in these two positive reports passes. Their actual raw serializer/write information and withheld digest projection are preserved in the reports.

The earlier unsupported raw-capture authoring attempt, system-sleep timeout and dirty-source awake comparisons remain historical failures/development evidence in the author report. None is counted as the frozen packed pass. I did not independently diagnose host sleep or repeat the comparison; final packed success stands on its own source-bound evidence.

## Limits and cleanup

Both author and review checkouts are clean at exact131fb987; no test-owned installed-reference process remained on inspection. No author-source edits, runtime/source mutation, outward posts, pushes or main integration. This is finite evidence review with both new whole flags false, not profile qualification or acceptance of the later attribution/B metadata plan. Main's separate TX changes and any future ordinary integration require their own exact-source checks. External council and root disposition remain separate.
