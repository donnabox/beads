# Minimum receiving source-v3 outcome audit

No discrepancy found in completed source-v3 outcomes. All13 stages have raw/supervisor exit0, unchanged eleven-file source hashes, no timeout/forced cleanup/runtime-root residue, and exact recorded argv matching the released manifest. This read-only evidence audit ran no tests, builds, product imports or services and changed no source.

Frozen dirty base remains5d7fa3f049fa0b4db3aa8dae20e20d8e47fb3854; source binding b7082c68869b1cecd715ed8d6e0ef88c6ba064a2e698beddf02844ed4dc252fb and full tracked+new-file patch7a6cc4e63f72d3920f04e5e8e03ff3915ed7488cb93d324ea22374fecff0f375 match current bytes. It has not been reattributed to the later principal merge.

Structured Vitest results and raw logs agree: **18 selected cases passed / 147 skipped**. Selections are consistency3, HTTP8, Snapshot1, schema1, changefeed1 and catalog/spec4; all executed full names and files match the released selectors, with no extra executed case or reported runtime error.

Actual HTTP log records corroborate the new receiving boundary:33 direct handler calls with malformedEntries empty and absent-minimum readReceipt entry; six raw responses are400, HEAD has zero body bytes, and all six lack the development marker; the emitted-listener positive returns400 malformed-request/retry never with marker null. The source command ledger's131HTTP/13owners/12listeners/11directories is a static model across selected callbacks, not independently observed aggregate request instrumentation. Direct handler calls are distinct from HTTP. No product child is planned by these source callbacks.

Format succeeds with one oversized JSON warning. Lint succeeds with867warnings/54infos. Boundaries validates11 workspace packages; Read seal emits evidence.verified with89rows. These passes do not mean warning-free source or complete profile qualification.

All60 sampled PIDs are independently absent in a current process listing. All selected test temp roots are gone; only seven Node compile-cache roots remain under the tool stages. No copied/mutant/installed phase was executed by this audit. Group sampling has the usual short-lived/detached-process observation limitation.

Original run and run-v2 failures and logs remain separate immutable historical attempts; source-v3 does not relabel them. Code council, negative control, fresh installed92/six-child validation, actual integration/commit and exact-head CI remain separate root-owned gates.

| Stage | Raw / supervisor | Seconds |
|---|---|---|
| boundaries | 0 / 0 | 1.284 |
| build | 0 / 0 | 1.428 |
| changefeed-positive | 0 / 0 | 1.263 |
| consistency-positive | 0 / 0 | 0.879 |
| format-check | 0 / 0 | 1.041 |
| http-positive | 0 / 0 | 2.477 |
| lint | 0 / 0 | 1.099 |
| offline-setup | 0 / 0 | 0.762 |
| read-seal | 0 / 0 | 0.892 |
| schema-positive | 0 / 0 | 1.079 |
| snapshot-positive | 0 / 0 | 1.270 |
| spec-static | 0 / 0 | 1.761 |
| typecheck | 0 / 0 | 1.491 |

Result/log pins:
- changefeed-positive.results.json: 40bfd05c8c03ebcbbcff4a18ce1e46aeb1acea53b44c041b3749706a6d320c2f
- consistency-positive.results.json: 6fbe391752e8d4d3da8b924acdbf6176900630c51793316fbaa119b40aa6c726
- http-positive.results.json: 0f0ea02bb23a993bedd65e2bed43eb82bd966dc03e76c070e3acc66ff837e998
- schema-positive.results.json: 911731a4f5aa978b84875564e2b5a819b612d45359df5cc45a681bbb24f7409b
- snapshot-positive.results.json: 5eaa1d21a5716432adae3bae2a4ac360102314471bbb915b1776ffeb02230eb8
- spec-static.results.json: 3bf69c378e01a69151be4c0880774f233beb6ccb771f77e5c7436fb71de33f7e
- http-positive.log: 9b4cd8b7288806df7acb9356e310bbaace069de698a4bcddad6eef29989bbf55
- lint.log: 855f133fc8d24123943fbb438aa6fce2438cf8cc60b68fdb0e3a114b6eb41913
