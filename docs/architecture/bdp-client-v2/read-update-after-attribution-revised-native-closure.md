# Native closure: revised next Update evidence plan

## Disposition

**Clean scoped closure; no actionable findings.** The revised plan addresses all seven original external findings and the intervening9c numeric-catalog ownership dependency. It is feasible as a bounded evidence-only implementation plan with both new wholeContract flags false. This review does not authorize implementation, promote coverage, claim PR94 CI/merge completion or qualify the profile; root supplies the accepted ordinary implementation parent.

Exact revised plan: `read-update-after-attribution-next-plan-revised.md`, SHA256 `0af9d4d447b0a571d702c960de4e48ff2925074a4af89a535b0077dc1c285a8e`. Original retained plan: `read-update-after-attribution-next-plan-original-116be54.md`, SHA256 `116be54fe2306c7aaa779cc5b9a13c2d2f5010120613f63e6b33e1b87207285a`. Source reviewed read-only: `9c7b643d22562ebabd8d9b51a24157b66f3e5c89` in `/Users/dbox/repos/wt/vickie/vickie-read-update-attribution-review-repair`, clean before and after inspection. Original external review: `read-update-after-attribution-claude-plan-review.txt`. I read the revised plan end-to-end, original plan and original seven findings; all measurements below are independently computed from9c, not copied from the old28a review.

## Evidence and finding closure

| Finding / concern | Independent source check and disposition |
| --- | --- |
| Coverage growth anchor | Current coverage is648157bytes/13494nodes. Proposed648157+65*1024=714717 is explicit and below the unchanged1048576-byte parser cap. The old590847+65536 whole-file pin at `read-update-matrix.test.ts` is explicitly replaced rather than accidentally retained or silently weakened. Nonselected rows remain semantically frozen. Closed. |
| Continuation growth | Current535052bytes/14286nodes. Explicit byte ceiling995852 and node ceiling26286; retained historical27158 total node bound is872 higher. Both are below immutable4194304/100000 factory caps. Fixture29055+8192=37247 is stricter than old28753+32768=61521. These are separate planning pins, not runtime/parser-cap changes. Closed. |
| Action/control accounting | Current old A/B rows are75+73=148 under the existing150 pin;13 client plus3 raw controls are bounded by16. Revised R/N have separate72+52=124 and6 raw+3 client=9 tables and source ledgers. They do not enter old arrays/loops. Closed. |
| Effectful properties comparator | Each R create correction differs from its patch correction:9007199254740992→9007199254740994,0.125→0.25,1e300→1e299. Both literal properties and companion metadata must change in response and independent GET. This detects dropping only propertiesChange; refusal/correction still changes only the numeric token and key. Closed. |
| Raw versus client controls | `read-update-matrix.test.ts:397–440` exact wrapper receives raw requests/responses; client actionExecutor around455–500 separately identifies the client action and verifies execution.input. Revised raw1/2/5/6/7/8 and client3/4/9 split matches those seams. UTF8 bodyOctets recomputation, one application and exact intended assertion IDs are required. Existing controls remain intact; no public/helper/executor seam change. Closed. |
| Depth justification | Actual64KiB request/properties and artifact depth128 limits are identified. There is no receiver document-depth maximum;16 is a finite sampled document depth, not patch-path/Selector depth. No arbitrary-depth qualification is claimed. Closed. |
| Metadata law versus configured behavior | General model at bdp.md891–903 supports the inference; properties-only numeric MUST at2522–2529/PROTO-016 is not misquoted as a new metadata clause. Problem diagnostics3452–3473 directly govern document labeling and optional instanceLocation. Metadata refusal/omitted patch location are explicitly configured adjunct observations, with an owner/spec audit if behavior differs. Closed. |
| Five-file catalog ownership | `numeric-model-catalog.test.ts` currently permits only same-value-no-op in the continuation and the Read companion in read-v1; all other file/row pairs are false and standalone numeric-model-v1 executable binding forbidden. The plan names this fifth file, exactly two additions in the approved continuation, and denial tests for wrong filename, Read ownership and catalog/version identity. Existing requiredProfile selection remains. This is necessary closed-world test maintenance, not a production loader or normative catalog change. Closed. |

## Preservation and source measurements

Independent Python byte/node/SHA256 calculations reproduce every revised-plan baseline:

| Artifact | Bytes | JSON value nodes | SHA256 |
| --- | ---: | ---: | --- |
| baseline manifest |3626922|97116|c86b33535f25a1d117ce021a3bb70164f5635bdaee50758e3428c130ccc3227f|
| continuation |535052|14286|69a8387f0459b35638f805e38bb9309ba585c956ed8e1a92bf4719801fd48aed|
| fixture |29055|750|07bbe646a5ce8436e8a746fadd22a9906013f0c28e28f608dcd9a0a84aa7b234|
| coverage |648157|13494|46d55f94e2726f50de91f857bfb7dee5f09f6b6bb08281ecece669b707aeced9|
| index |242|9|7277980894dd0955a8fcd22ba5d2845b0a6801b7e7e13a937018e11d3a65de37|

Raw four-space scenario block measurements also match: attribution229073bytes/hash1b81ae03b7aea896732e001291bb2ef92edd6bf5586c13a7e91afaf563f3e769; C114099/hash39eb2fe22d25eada2c43490f37b6cb5d66950bd90042af1ae954036426e07f2a; no-op191793/hash368c2d3d2aead864bfe925f8777b55ce9b09a107c81429820ac390d6d0a8550d. The revised plan freezes these blocks and the entire baseline; whole continuation/fixture/coverage hashes correctly identify starting bytes rather than impossible final-file invariants.

Coverage has129 rows and30 whole, and both proposed rows currently have empty bindings/false flags. Canonical order places them after no-op.52→54 authored and77→75 unplanned follows two additions, while30whole/99without-whole stays unchanged absent separately accepted promotion. The index has no scenario inventory to edit. Plan instructions explicitly require reconciling an intervening accepted parent/promotion rather than overwriting it.

## Semantics, executability and bounds

Checked catalog numeric-model-v1 lines38–97 against bdp.md694–723,891–903,2518–2529,3452–3473 and requirements.md109–115. Checked runtime `json-admission.ts:212–217`, `read-update-values.ts` admission, `member-numeric-diagnostics.ts:13–51`, `member-executor.ts` numeric failure and `development-authority.ts:230–238`. These support exact-decimal classification, whole-create document-relative escaped pointers, patch input-pointer messages with no invented postimage pointer, metadata document labels, absence of Type blame and configured422 diagnostics. Fresh keys after an admitted numeric failure are appropriate; no unsupported unclaimed-key/retry assertion is added.

R ledger:6navigation+4endpoint create/read+24whole-cycle+24patch-cycle+12post-Link-refusal endpoint reads+2final endpoints=72 (29 raw/43 client),8 Resources. N:6+4+32+8+2=52 (21 raw/31 client),6 Resources.14 new fixed GET bindings explain expected35→49 continuation routes; the plan requires actual ordered enumeration/new digest rather than count-only acceptance. Body estimates remain plausible below4KiB even with the301-digit literal and finite16-level objects, but final measured values are an implementation stop gate, not a proven measurement today.

The target uses32/524288 diagnostic bounds; these two-diagnostic cases are well below them and claim no truncation coverage. Existing empty endpoint-multiplicity policy and shipped Task/relates descriptors permit the proposed separate same-pair Links; withSchemaCohort appends its descriptors rather than changing those types. Full postimages are independently checked before client-only capture, so later unchanged-record equality is not self-derived from the failed request. New raw mutations preserve lexical tokens and authored/write assertions; no raw object-capture extension is requested.

Source app90s aggregate/120s outer/5s exchange/30s focused budgets and recorded240s supervision are preserved. The cited9c packed summary reports242 tests,127.3556s and1309296KiB descendant-tree peak; these remain historical author measurements, not my new execution or a performance guarantee. Two genuinely fresh shards are required after fixture/coverage changes, with separate digests and exact associations. Finite singleton partitions, inferred metadata adjunct, sequence/retry/concurrency and universal any-depth limits are accurately distinguished; no new protocol law or whole credit is implied.

## Scope and verification limits

Read-only source/plan/hash/ledger inspection only; no tests, builds, installed processes, credentials, production modifications or outward writes were needed or performed. Session guard succeeded. No helper/framework expansion or unresolved product choice is required for the proposed finite cases. Authoring still must prove actual counts, allowance fit, negative-control assertion sets and fresh installed behavior; if it cannot, the explicit stop/report instructions apply. Root acceptance and implementation dispatch remain separate.
