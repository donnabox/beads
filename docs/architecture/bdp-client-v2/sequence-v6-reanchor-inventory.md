# Sequence v6 mechanical reanchor inventory

No new semantic blocker or four-path byte drift was found against actual main0d63. The accepted sequence plan remains authoring-closed. This is a read-only inventory for a future reanchor after the depth repair merges; it is not a new plan, review council, allowance or execution release.

Accepted v6 `591b1088e101bdb12d90e3366eea1ec2c605777a59e37ca804f68062da5e4a9e`; companion `15e050224f9bc5e1f2cee34d74023b2711adfb8e6651c337ca4f1cb55c2cf694`; root review `8afa160d01575e0fc5980a6042b51b9c52e4cbe112f7aa3d709139ba2dc30333` (`PLAN_ACCEPTED_AUTHORING_STILL_CLOSED`). Frozen originals are unchanged.

Committed identity: `0d63cf59d447eca2e7ff8b2e0e6a5f31513b9a3a 80a6fe0605698e18f9504432407e7943e1f665b4 77f4b89c117e46f0e73476f69b4c8b377306452b 1f9cb46c42fa69a09ac847f57e47bc96390aea6f`. Actual PR130 postmerge record `aeb03629fd82d95ee80c5d42deab0777210b457f622a22e5630dad4ed0f91f1e` verifies the same tree and ordered parents. No future depth commit/tree/CI/merge is known or inferred.

## Four owned files and historical preservation

| Path | Bytes | SHA256 at0d63 (also9a42 and77f4) |
|---|---:|---|
| `packages/conformance/test-support/transactional/manifest-v1.json` | 738624 | `29729151cf6da641545910ec3880f9609c13b319af7bbe99d3d2ad47833dab96` |
| `packages/conformance/test-support/transactional/coverage-v1.json` | 108139 | `4619813e9446e0e714a64c11eac184531950f57720e32c7ef414c7e3342121af` |
| `packages/conformance/src/transactional-cumulative.test.ts` | 101554 | `22a568820ab2abc187d37c27658fdcf6e5ad39e74e2172374e3658c180f7668e` |
| `apps/bdptest/src/transactional-matrix.test.ts` | 111388 | `6c0c240afef61831d0573775ea35b0c04b116a55da1d3fb40127b9ede68e939c` |

All four current working files also equal these committed bytes. Comparing PR126 merge88c through PR130: only the app changed among these four, by PR128’s already accepted cancellation timer repair; PR127, PR129 and PR130 introduce no additional four-file drift. v6 already binds that app6c0c successor. Depth successor manifest `5fbe5c9b0c57f2c253fb16d114675ccb4929a9e4e18b21a1efbf1ca18385d778` contains eight disjoint authored paths, so its current dirty source does not change any sequence-owned byte. This does not predict its future accepted/merged tree.

Independently decoded raw slices use the same object boundaries as cumulative objectSlices: opening `{` through matching `}`, excluding inter-object whitespace/separators. All14+14 raw slices match committed9a42 exactly; retaining original prefix/separators/suffix reconstructs both complete current files byte-for-byte. Their old object order, minimum at0, pagination at10, cross-carrier at12 and failed-retained at13 are unchanged. The proposed sequence row remains append-at14.

| Index / row | Manifest raw bytes / SHA256 | Coverage raw bytes / SHA256 |
|---|---|---|
| 0 `read-update.http.command-minimum-fields` | 54412 / `f04c001617ccad14ef67a82586ea5f61ec918bc2545bb041fafc21dee07b5d7e` | 8888 / `47914ed67c7b06e721e74134e0c633f3e1a7d461cb50ea200ff1874686885c2a` |
| 1 `transactional.batch.atomic-commit` | 14936 / `0dfd3b9ed02d01f76ecac08c00b009b4bb9f03e57f7e15a2266eed82fc144747` | 2383 / `ffd77a9d34ac41a6e1f960671bea6880eb5b2ab7b11b3af4ccf0177363d7f4d7` |
| 2 `transactional.batch.rollback` | 16319 / `be24067358af7b5acf7bc1beae2eb6534cbc0effa6f65406ab61b139310bfc37` | 2857 / `9e4f5cdb45f0133f1c6d23caa402701981db2e6f4198e23c667579622ace6769` |
| 3 `transactional.batch.deletion-safety` | 131808 / `3e66e7c95aec963cc80e076249a5870b952765b2e98c1023ea148f8f628e6aeb` | 17375 / `ac27a329a1cc013e6a0e24e0f5cad6b5b4af6c128ad65302d030422a3c7e3c0a` |
| 4 `transactional.batch.no-op-entries` | 81047 / `9307b230eb053131e9127cbbb5e6285305b79d6794edd22f5432a07fc74d9faa` | 10587 / `383de4e0cb9b3bad310b9db7ff9009cb59fac55966f0309bc71c5d007bea6c1c` |
| 5 `transactional.set.mutation` | 71739 / `6e7280a9ebbbc3faa1da989015cc802f82cd51b17a9cea11e7cc91392bef49d6` | 11676 / `058147f8194f091ee5a7ef3d4fd378f0324f6570049df777f993acb5f967d540` |
| 6 `transactional.set.zero-match` | 36055 / `35aa23bd32539869ea3db41e1c9f3e81a3217e2eb983e764e860a6e00c9eb2c9` | 6520 / `4e79e18a5e4c2735b5c686b9ad610dedc890386bfb06584f98268a421bbb832b` |
| 7 `transactional.set.cardinality` | 28546 / `205af1c54d62efb09c9eadf20147a1730909e788c0d4763ce9718b8648f24ede` | 5800 / `2a255ed4d364c435c9f985d7a770386bc7b20b2865887b9c8cc2d6cc5f0dd116` |
| 8 `transactional.set.singleton-targets` | 40018 / `561b8ebc21106628918111332a9ab7b47be7350935925ed1e76a6cb5bed576b9` | 7314 / `0c9008bac3e2c047d9aa8d47d343b7ed936c9c9352bb140c6422127c71b156b1` |
| 9 `transactional.receipt.readable` | 12545 / `ddb97e450e233b941f7d905d49f48648d9965ed1b361dcdc5dee6f810038a1e0` | 2142 / `684eba200912ccb0932c0a8286e350a6e545262a205b3ea3df47a12fc6dd18ec` |
| 10 `transactional.receipt.pagination` | 138750 / `c531c260349347cdc3e54ba68282f1a266a5fa207d3775660615f836a4768782` | 13057 / `0735edbca2fbf384f9f8147140a353dc084d9ad85facaddf2f5073f56a7256ee` |
| 11 `transactional.idempotency.conflict` | 29772 / `ba93aea884ae462533d02695642f03a2ca854f5adabeb41867254216dabb2106` | 5188 / `de9b758f8ac4df7db5597b1eb2b402a116720becc30a41e7198f78f244358ec0` |
| 12 `transactional.idempotency.cross-carrier` | 64106 / `4ad8c05fe5fbb4682809d4481c93e2684fe30fbbb8962b4da498590637f8f35b` | 11196 / `c5e6076a2104c06f88fac88c77214522cab457faf925fe03881dff9778c35245` |
| 13 `transactional.idempotency.failed-retained` | 18404 / `a867694406dde07da71ac70197c60c478dd0fb3344cf2c8160795d1aa62d3378` | 2917 / `63918cbb495a504fd2043f52b60e7cd8ef1518b8022a96edb0d02660c7df1a8e` |

Current JSON still has14rows,210actions,2294assertions. Whole-file equality preserves the accepted738624disk/750430normal/366116compact/19570node and108139coverage measurements; those normalization/node figures are retained measurements, not newly executed product measurements. Historical raw pins56318c56…/dafedd035… and older layers must remain historical. Future append must strip the exact new tail first and reconstruct these immediate-parent bytes before older withoutMinimumRow/cross layers. No new tail exists yet.

## Interfaces, callbacks and relevant merged changes

All eight exact static titles in the companion each occur once in current committed source. Cumulative objectSlices669ff/assertOriginalObjects828/preservation836, pagination1007 with index10, withoutMinimumRow2408 and minimum preservation2724 retain their current anchors. App observe98, onObserved327, positive title353, summary589, cancellation first-wire option895, static callbacks1321/1607, crossCarrierWire1911 and commandMinimumWire2323 are unchanged. Cross finish1976–1981 remains pure; its post-prefix forbidden sequence-route detector remains intact. The15-route digest02dd7d9f… remains at394/1357.

No added observe call or row traversal comes from PR127–130. Static text has21observe call sites (excluding its declaration); loops/table expansion are why the accepted invocation model is54, not21. Byte identity preserves the reviewed literal inventories (40mutated+14unchanged; proposed41+14 and reachability axes). This task did not re-execute or invent an expanded callback count. The accepted ten actions, four cohorts623expected/632maxHTTP/71children, eight static callbacks and all guards/timeouts remain unchanged; current whole-job totals include unrelated newer tests and are not the app model.

| Unchanged relevant source versus reviewed9a42 | SHA256 at0d63 |
|---|---|
| `packages/conformance/src/runner.ts` | `ac20c7fbfe15a5e42a6559e89cd3467562efe9bc74d462478414414ed3efede1` |
| `packages/conformance/src/transactional-observation.ts` | `603fcc1240e43da3c477f42b8b942141c06c989a48a2a88099fd5d90e283159f` |
| `packages/server/src/transaction-sequence.ts` | `019e02f2f392047d568d7464b1c5af0482d6d743ed09a6d0aa4edab151607cbe` |
| `packages/server/src/transaction-events.ts` | `c6052433662fcb2d0a202ad4e408a24d4c43f980313fa89659501b4475790db5` |
| `packages/server/src/resource-evaluator.ts` | `89606787e455f4ed555b807d91c5331a376431807a9a581de1f94e538fcab4ff` |
| `apps/bdptest/test-support/transactional-execution-binding.ts` | `cde506c72a8710eafe58522900f6db70ed3701cf333560252dc8e7c5e706cf89` |
| `packages/conformance/test-support/transactional/fixture-v1.json` | `3a7899ced771d4880066d98b0000c53f801ece53458502cc4f3bc641146c9a85` |
| `packages/conformance/catalog/transactional-v1.json` | `c9125f914129c71f01cefc5c54b32db80637d37b57e321571c96e81936c31eb4` |
| `docs/specs/bdp.md` | `f036103ffdf5f89fa3a2033570e1e987c4cd203cd80b1c633cb44615123f892e` |

PR127 Event404 and PR128 missing-receipt context changes are already within the v6 reviewed9a42 baseline. PR129 adds an ordinary Update witness outside this scope. PR130 changes only invalid changefeed parameter error context in transaction-changefeed.ts, transaction-store.ts and transaction-http.ts, plus tests/progress: no sequence request, successful receipt, finite Event projection or capture/helper branch is changed. Its parser gains7lines (old changefeed91–112 references shift to98–119); the one-group finite Changes behavior remains, and v6 deliberately has no s-groups action. Header/key routing before HTTP609 is unaffected; the catch block after609 shrinks3lines.

The pending depth repair changes transaction-batch/store/submission (plus tests/progress/new installed test), not runner/capture/helper/sequence projection/Event rendering. It introduces a private receiving depth fact for direct carriers and compares retained keys before fresh refusal; the v6 shallow fixture bodies retain their accepted semantics. Current candidate offsets move batch code after preparation by48lines and store code after admission by9lines; these are provisional dirty-candidate offsets only. Recompute actual line anchors from the final merged depth tree, rather than replacing v6 citations with unmerged numbers. The complete accepted header ledger, only three scoped URL captures, helper original-wire receipt equality, independent Event-ID projection, seven named negative failures and exact ordered handoff remain required.

## Minimal substitutions at the eventual reanchor

1. Update v6 chronology/context and companion currentContextCommit/currentContextTree/contextOrderedParents/contextProof: parameter merge is now actual0d63/tree80a6 with parents77f4+1f9, verified in the bound postmerge record. Do not keep “parameter merge PENDING”. Preserve reviewedBaseCommit9a42 and all original council/source/timing attribution.
2. Add depth as the new outstanding predecessor. Keep actualAuthoringParent/actualAuthoringTree null and authoringReleased:false until its exact green CI, ordinary actual merge/tree/parents and current branch/source are supplied by root. Then bind those actual identities, rerun this four-file/28-object comparison against the actual merged tree, and reconcile any unexpected delta before authoring.
3. Carry all four current hashes unchanged if that final comparison confirms them; preserve all raw objects and full-parent byte pins above. No fixture/body/header/action/route/assertion/guard or callback-count substitution is needed on current evidence.
4. Refresh source citations only where actual merged code moved: changefeed parsing after PR130, then depth batch/store/submission anchors. Sequence normalization/projection, Event rendering, raw captures/receipt helper and the four owned file anchors currently do not drift.
5. Record newer CI only as separate historical observations if useful: PR130 run38755921518/job116494181912 passed9303+1skip/177files,475.31s suite/630s job with all downstream gates. This is not a guarantee of capacity for404additional expected wires/48children and is not evidence for unimplemented sequence rows. Keep all v6 simultaneous draft stops and no allowance/timeout increase.

Unresolved gates remain actual depth code council/control/installed/CI/ordinary merge and final reanchor, then explicit root authoring release; later measured source/derivative fences, bounded source/cohorts and final code/CI/merge gates remain as accepted. No public profile claim follows. No tests, build, install, service, product imports, new review, worktree, source edit or coordination mutation occurred. Only this external inventory was written after the session guard.
