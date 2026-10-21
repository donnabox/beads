# Ordinary Read+Update next-serving audit

**No concrete new ordinary Read+Update serving defect or necessary missing targeted regression was established in this bounded pass.** I do not recommend creating another Update implementation/test slice merely to increase diagnostic coverage. Keep the independently established Transactional fresh-depth work moving; it is not an ordinary Update rewrite.

## Actual source and scope

Reviewed clean workspace `/Users/dbox/repos/wt/vickie/vickie-update-cold-directory`, HEAD `3bb2f2449ce345e118d04c07eadd777363add544`. Its tree `69497d4ffc2089ce56790a9a257ecdff2eb0b47b` equals the locally available actual verified main merge `9df574a9ac980272fa31be9f443a771233b69c6a`. No pending minimum or sequence changes are treated as landed.

Read the sole design plan's active objective/U1–U3 and current live ledger, then the prior `update-after-pr131-audit.md`. The six ordinary serving modules and four existing regression files pinned by that earlier audit are byte-identical today. I reread the load-bearing receiver, owner, retained-disclosure and relevant test paths, rather than inferring current correctness from the historical report alone. Shared `authority-read.ts`, `read-http.ts` and `read-pagination.ts` have no diff between that prior18104 base and actual9df574. Newer command-minimum/cold-directory evidence is considered separately.

This was source inspection and data-only hashing, not a test run, fresh installed observation, exhaustive conformance review, or qualification decision. I previously authored/reviewed related evidence preparation; that overlap is disclosed. No product files, branches or bound review inputs were changed.

## Boundaries checked

| Boundary | Actual behavior and existing targeted proof |
|---|---|
| Cold operation navigation and ordinary current state | `development-runtime.ts:335–414` routes canonical operations separately from composed reads; `development-authority.ts:296–344` uses the ordinary owner for both. Existing `development-discovery.test.ts:121` navigates all eight singleton targets plus sequence;631 exercises the SDK without supplied operation URLs. The newer matrix code1580–1623 and1680–1705 records real raw attempts/completions and client observations separately, joins their ordered mixed ledger, checks actual directory contents and the missing-member counterfactual. The ledger records delivered PR138 installed evidence, not a new observation made here. |
| Commands ignore minimum fields without acquiring Transactional semantics | Spec3561–3588 explicitly excludes durable receipts/Events/Scope context and requires the three minimum fields to be ignored on mutation POSTs. Runtime444–474 passes only operation/body/key/Accept into the ordinary owner; no minimum tuple enters identity or admission. Current matrix oracle1010–1200 checks exact authored raw headers and serializer completion, full ordinary schemas/postimages/replay equality, no TX context/receipt location, and the genuine negative assertion-input change. PR140 already supplies the newly targeted11-response witness and2-response negative; it remains partial qualification. |
| Revision concurrency, retry and afterstate | Resource evaluator1015–1027 checks permission then expectedRevision before writes. `read-update-guarded-overlap-installed.test.ts:460–550` requires two client requests outstanding before first settlement, one200 winner and one exact409 Problem, both full schemas, current winner state, exact replay of both dispositions and final unchanged state. It does not claim physical simultaneous server execution. No new overlap scenario is justified merely by the current TX work. |
| Receiving/negotiation vs retained identity | Runtime authenticates before body/key processing; ordinary owner prepares and submits through its lifecycle. `member-executor.ts:720–780` compares retained identity before returning its saved disposition, applies current disclosure only to an equal retained Resource result, and does not rerun the mutation. `development-discovery.test.ts:272–342` already checks cross-carrier keys, retained stale-revision failure, changed-body conflict, independent tail, unknown-key syntax/media/body/Accept refusals, equal replay406 and changed-identity409 despite unacceptable Accept. |
| Trusted principal and current disclosure | `development-authority.ts:166–180` captures policy by trusted principal; writer actor metadata does not select it. `development-contracts.test.ts:281–335` checks forged actor, independent principal key behavior, hidden404, retained Resource403 without disclosure, restored visibility yielding the original postimage, and retained failure despite changed permissions. The next callback336+ covers hiding an owned Link. `retained-disclosure.ts:18–45` deliberately overlays the saved Bead for retained postimage visibility; it does not incorrectly require the live subject to exist. No broader ACL/Types prerequisite is inferred. |
| Read methods, conditions and lifecycle | Shared `read-http.ts:12–45` applies negotiation/conditions after selection and preserves HEAD representation length. `development-discovery.test.ts:231–270` covers methods/Allow, HEAD, ETag304,412 and406. Its480–563 callback parks the real sequence scheduler, reads a committed prefix, disconnects the client, drains admitted work, reopens and checks exact saved results. These directly address ordinary sequence partial commit/disconnect semantics without TX receipts or rollback assumptions. |

## Recommendation and limits

There is no independently established small serving correction to assign from these paths now. Further combinations (for example another malformed-minimum variant paired with a retained conflict) would be optional extensions of already separate receiving/identity proofs, not evidence of a missing production branch. I have not promoted existing tests to a whole-profile claim or treated a historical full-CI pass as new per-response evidence.

The sole plan still requires target-bound cumulative qualification and truthful cold profile selection/activation; existing private/non-attesting launch behavior cannot simply be relabeled. Those remain delivery constraints, not a newly diagnosed ordinary serving bug. Optional Types work, S1 retained-Snapshot minima, G2 incident-Link Event membership, public admission and unrelated History/provenance were not reopened. In particular, the fresh-depth Snapshot comparison finding belongs to the separately scoped Transactional path; ordinary Update has no Snapshot obligation introduced by this audit.

## Exact source pins

- `packages/server/src/development-runtime.ts`: `37e222682645e9d8f2d5fcf62de208640fb6216c9e9f2f279ab6017e8b9d6ca9`
- `packages/server/src/development-authority.ts`: `d49f89da345234f7eaee83dd872a5c2d7a1a8d38ffc794e5e69d333976d22d49`
- `packages/server/src/sequence-executor.ts`: `bea7623cf47521329fa0b3f65def22bbb0a970b066609f8cba4eac17d2181a27`
- `packages/server/src/member-executor.ts`: `cfce8e561f9315948c6b3b3def25893a7dfde78e2cc338fa142b859ae209771f`
- `packages/server/src/resource-evaluator.ts`: `89606787e455f4ed555b807d91c5331a376431807a9a581de1f94e538fcab4ff`
- `packages/server/src/retained-disclosure.ts`: `68bd3e3d0e531738aae65081d8d2532750e6f135436c2177d30678ca384b3f36`
- `packages/server/src/authority-read.ts`: `fd6554c4ee0d08b997fb3622c2f21697f02d58b838f4ae0de10aca4707f67734`
- `packages/server/src/read-http.ts`: `31934105eff6fff8ded8ce2ca9e3c19028991ca7c3d5f51b848004177e8c9268`
- `packages/server/src/read-pagination.ts`: `ee2462c355aec5b5c4f5fbd7b323ca6d814591c0a46c809ebb4636596e4f52f8`
- `apps/bdptest/src/development-discovery.test.ts`: `88d2bcb716ef34110d20e944548e8fc74d4eb5d706377a1853030d68c40703d3`
- `apps/bdptest/src/development-contracts.test.ts`: `c27a4ff9012c85af17bd38e4f06a3b65d26a75a949e57a572a4382fd5bf6e522`
- `apps/bdptest/src/read-update-guarded-overlap-installed.test.ts`: `767ed468ac29347100203dbb9ff5f435ca2dfaebcad98c61158e5ed59d63a1ed`
- `apps/bdptest/src/read-update-matrix.test.ts`: `376709ecdcf356a98c5bf7770e6b0a637ab6e06cf9a9f7ed68c2e5c15da9dd36`
- `docs/specs/bdp.md`: `63e93be562334c84c1e967da2fce212e8ef515aa3c9448924c5f9a272ea7b80b`
