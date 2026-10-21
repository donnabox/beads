# Guarded-batch reanchor and authoring checklist

Authoring remains **closed**. Root plan acceptance root-plan-acceptance.json accepts plan c2c4642c / ledger5aa8f750 / dispositions512904a1; it explicitly requires actual PR134 merge/tree verification and reanchor. This checklist adds no plan cycle, source scope or validation permission.

## Verified candidate comparison

Planning basis:98c600767e8e91a79e109cb741beb91e23b9bebe/treefde859381752d15c8caabcbc946d3c2e384b9cf6. Read-only comparison used exact committed blobs at candidatea18fa993703185a108df4ad75b6d142072ba8736/tree592c5e6e5c0c790b6e2690495b94ac48705b9069; no future merged SHA is invented.

All38 meaningful source paths in source-binding.json were compared. **37 remain byte-identical** to their planning pins. The sole mismatch is the expected append in docs/design/transaction-implementation-progress.md: planning SHAfd69135a88ee2802eb9ce55d2c2f3771d4a16228c0e306202c9a8778f0bdcb3f → candidate21069d8d8eb5d5048a1dbcb38871ea0f8e72b463e9456633b12b4860e024a728. The append is the historical same-key dispatch draft paragraph; preserve it verbatim, including its historical unexecuted attribution. No runtime/schema/helper/spec dependency mismatch found. These38 pins are meaningful review inputs, not exhaustive dependency closure.

The additional implementation precedent is the final committed apps/bdptest/src/transactional-same-key-overlap-installed.test.ts at candidatea18, SHAb13688d6b02c5901e68251620ab6d7e00f9f675af4025319817f9e8cc6c9bca3. Its final reporting branch (:932–965) rejects both initial file-report overflow and overflow after reserializing the final stdout report: set reportOverflow, addFailure, emit only bounded safe fallback, and then throw preserved failure(s). Reuse this corrected behavior rather than the earlier e6a52fa4 pattern. Do not import another test or copy its whole scenario; borrow only necessary local transport/lifecycle/report practices. Existing actual same-key6+2 results are precedent, never guarded-batch evidence.

## Root reanchor before any authoring

1. Verify actual PR134 exact candidate/council/green CI and ordinary merge externally. Record the real main SHA, tree and ordered parents. When the incoming base has no extra changes, the expected merged tree is candidate592c5e6e5c0c790b6e2690495b94ac48705b9069; require actual tree equality, not merely a branch name. If other main changes exist, root must inspect the actual combined tree and relevant dependencies before release. Do not fabricate the future merge commit or bypass integration on this checklist.
2. Compare all38 source paths against the actual merged tree. Expected reanchor is37 unchanged pins plus the documented progress append. Verify the final same-key precedent still hashesb13688d6. Any additional semantic/test/helper/spec mismatch goes to root before dependent authoring; mechanical unchanged-source reanchor alone needs no new plan cycle.
3. Root supplies a clean isolated actual-merged-parent checkout/branch and exclusive ownership release. Preserve the accepted planning/source bindings as historical; write a new actual-parent binding rather than modifying those records.

## Exact prospective ownership

- New apps/bdptest/src/transactional-guarded-batch-installed.test.ts.
- Append only docs/design/transaction-implementation-progress.md with factual **unexecuted** private-evidence prose at the actual parent; retain all existing history.

No production, fixture, schema, catalog, manifest, coverage, shared helper, scheduler or public-admission changes. One registration: `preserves one guarded batch winner and rolls back the losing marker through installed HTTP`. No new metadata-only case or independent framework.

## Draft acceptance and measurements

- Keep one fresh configuration per mode, actual served identities and eight-package provenance before execution; exact descriptor170B/schema166B and five configuration keys. Secrets never enter reports/stdout. Default positive; only explicit unguarded-contenders alternative.
- Seven positive requests, all200; control three completed200 responses then the sole named `guarded-batch:one-retained-revision-failure`. Combined10HTTP/two scheduled children/owners/listeners/stores; automatic retries0, deliberate retained replay1, no bootstrap/continuation/reopen. At most two outstanding pair requests and two live sockets; independent client milestones do not prove server admission overlap.
- Construct both guarded bodies once from the same schema/literal-validated seed revision. Pin343+r per guarded body/replay, r<=128, largest471B; seed172B; controls323B each; positive maximum1585B, control818B, combined2403B. Record actual runtime hashes and exact replay body/key equality; the reference UUID is never sent. Control deletes only both zero-based operationIndex1 expectedRevision members.
- Preserve full seed/winner/loser Resource and receipt oracles; failed embedded Problem exactly conflict/revision-mismatch/409/after-state-change/index1, no results/next/effectPosition; retained replay and issued-root full body equality; current serving context distinct from retained history. Require complete two-Resource canonical-order collection and three exact Events with transaction linkage, full data and independently projected private IDs. No loser marker/Event, no no-op ambiguity; don't claim reservation release or crash durability.
- Retain both control successes with distinct transactions/receipts, effect positions2/3, full own postimages and all prerequisites before named failure. Generic exit1, pending-only outcome, wrong schema/status, missing actual report or cleanup/report error never discriminates.
- Keep45s work,5s exchange,hard10s local cleanup race/5s request drain,60s test and120s command. Helper nested close-failure wait is not extra passing allowance or proof of actual process stop. Aggregate original/cleanup/report failures, handle late rejections, remove store only with known stop or no launch, and retain unknown pre-ready state for external supervision. Expected control failure follows completed pair responses; only remaining owned work/sockets are aborted/destroyed.
- Enforce streaming32768B response ceiling,16384B headers,65536B request/report limits, no redirects, exact canonical Host/query grant, exclusive outside-store output paths separately per cohort. Overflow/write failure must make the case nonpassing even when fallback stdout fits. Preserve allowlisted phase/mode/failure/prerequisite/count facts; raw response fingerprints are not full-wire retention.
- Freeze normally formatted complete two-file patch and actual source hashes, exact expanded selector/argv, literal/body measurement, configuration/route and lifecycle inventory, report safe-field contract and control prerequisites. No tests/build/imports/packages/services/commit/push follow automatically from authoring. Root reviews measured draft and separate code council/validation release. Final package gates require actual clean source/build/archive/consumer/member/payload equality; no workspace fallback or prior-runtime reattribution.

No concrete source blocker beyond the expected historical progress append was found. Await root's actual merged-parent release.
