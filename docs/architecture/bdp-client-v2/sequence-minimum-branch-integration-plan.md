# Sequence/minimum working-branch integration map

This is a read-only plan for merging the reviewed **minimum working branch into the sequence working branch** after their pending reviews and exact-source local commits. It does not wait for, chase, or merge into main. Root owns both commits, the ordinary branch merge and validation release. The editor reconciles broader main/spec changes separately under the operator's latest direction.

No Git merge, index/object write, source edit, imports, build, test, service or review invocation occurred. The prospective combined hashes below were calculated entirely in memory from line edits against the common base; they are not a claimed Git merge result or a committed tree.

## Frozen inputs and ancestry

- Common base / sequence HEAD: `df80f8cb86b7763bc7bb005492e43d8840527045`, tree `25e9d4aa926ea728f5e356365cb1ffadba269221`.
- MW: `/Users/dbox/repos/wt/vickie/vickie-minimum-current-resource`, HEAD `2e1676caa0f5861d16c5741fbd412f6c3ef7c269`, tree `d909acfdbc65e111855233ff86a3db6edfedd49e`, with exactly the two accepted test repairs dirty.
- Minimum final repair binding: `transaction-minimum-next-serving-plan/authoring-draft/ci-expectation-repair/final-source-binding.json`, SHA `bf1c223dede6029eb28a5ba4e264922a2b16f567e4549ecf6ef85cceee5c8e20`. Both actual file hashes match; the exact current dirty patch matches its recorded SHA `c105a4c1ed07241e5d10ff9cd095bd1e9fbd78fe5f4128844842c2aaba497bf1`.
- SW: `/Users/dbox/repos/wt/vickie/vickie-sequence-depth-retained`, dirty seven-file sequence draft at the common base.
- Sequence lint-successor binding: `sequence-depth-retained-member-draft/lint-repair/source-binding.json`, SHA `9eb076e7fee0b300010a4edac01f55fecd7ab6f1d27955931e7791556d2080ec`. All seven actual hashes match. Preserve the two cleanup-only lint repairs and their original failure attribution.

MW already contains the five PR140 paths. They are part of the incoming branch history and must not be dropped just because this merge is for Transactional work. No new fetch/main dependency is proposed. Spec bytes must be exactly MW's current `docs/specs/bdp.md`, not a newly fetched editorial version.

## Footprint and deterministic composition

The incoming branch differs from df80 in **16 paths**: minimum's ten original paths, one additional installed-test path from the two-test expectation repair (the other repaired path was already owned), plus five disjoint PR140 paths. Sequence owns seven paths. Exactly three paths overlap, so the intended union is **20 paths**.

| Shared path | Actual changes against df80 | Required combined result |
|---|---|---|
| `packages/server/src/transaction-http.ts` | Minimum adds lexical/current minimum imports and route classification/catches, changing alias/read handling around base212/233 and575–641. Sequence changes its preparation import at42 and only sequence dispatch at328–341. | Keep both sets. Sequence uses receiving preparation, rejects unacceptable Accept before admission, catches only the sequence execution development-limit into its existing private marker. Ordinary current read routing retains missing-first and unknown-current bounded-wait behavior. These are separate branches, not competing replacements. |
| `packages/server/src/transaction-http.test.ts` | Minimum changes existing callbacks around1784–3132 and appends63 lines at base EOF4794. Sequence appends396 lines at that same EOF. | Preserve every minimum existing-test edit and **both whole appended callbacks**, ordered minimum then sequence for deterministic bytes. Retain the sequence failure collector/cleanup-only lint repair. Do not resolve the EOF conflict by selecting either side wholesale. |
| `docs/design/transaction-implementation-progress.md` | Both append at base EOF1994: minimum29 lines, sequence3 lines. | Preserve minimum checkpoint first, then sequence checkpoint. These paragraphs retain their historical evidence attribution. Any later factual validation update must be a separately recorded edit, not invented into this composition. |

Using exact original lines and that explicit append order gives these prospective hashes:

- HTTP production: `4143e8a36021d60d1deba4649b1de5dfbcb3ee2d630486749932eccc1374c602` (28,669 bytes).
- HTTP tests: `f4353454c8f8a9dc42367395804bbe6309f7bcc51dba2b91b3176975d08b5cce` (200,546 bytes).
- Progress document: `9668b52c1d5cfdcc7f2a326cafe3468d87de6cca14438b73ae88816587b99c28` (135,887 bytes).

All other17 paths must equal their sole owner's frozen bytes. The algorithm applied nonoverlapping base-line edits from both sides, permitting only the two same-position EOF insertions in minimum-then-sequence order. There was no overlapping replacement interval. Git may still present textual conflicts because of context/EOF; the hashes specify the intended result, not a prediction that its automatic merge must succeed.

The minimum `transaction-read.test.ts` must retain its already-integrated authorization `originalReaders` facade and the new exact invalid-input expectations for accessor/inherited/proxy/revoked input. SW does not modify that file. Likewise preserve the minimum installed repair's GET404/full Problem/context/length and alias HEAD404/empty-body/representation-length assertions. The sequence preparation must not revert minimum's frozen own-data request classification.

## Combined validation recommendation, still closed

1. After natural review completion/root dispositions, commit the two exact frozen source states in their own workspaces; ordinary merge the resulting minimum branch commit into the sequence branch. Bind actual ordered parents, union20 paths and the three combined hashes above. Do not reset/rebase/cherry-pick away branch history. Candidate IDs remain unknown until those commits exist.
2. Run the normal pinned typecheck/build and relevant formatting/lint/boundary/seal checks once on the combined source. A fresh emitted runtime is necessary: source HTTP callbacks import dist, and neither historical branch build contains the combined HTTP module. Preserve original failed lint/CI and all prior source/package attribution; do not relabel them as combined proof.
3. Run sequence's exact existing18-case selection (2 preparation,13 owner sequence,3 HTTP) against that build. This is the accepted slice's final combined-source proof, not an invented expansion. Its21 source HTTP exchanges exercise the changed dispatch/preparation and receiving failures.
4. Add minimum's existing five HTTP selectors from source-validation-v3, **because this is the shared production and test file being composed**. They exercise epoch/view/timeout vs conditions, missing-first404, Type/incident exclusions, alias context/refusals and the new unknown-current/hidden routing. Existing ledger is172 HTTP/5 owners/listeners. Together with the sequence selection this is23 callbacks/193 planned source HTTP, not executed outcomes.
5. Fresh external-package proof should use one new combined archive/consumer graph: the sequence installed positive (8 HTTP/one child) plus minimum's existing32-response positive (one child). The latter already uses a real four-member sequence seed and covers current Resource/hidden/alias/incident-anchor minimum behavior through the combined dispatcher. Thus40 positive HTTP/two children give direct cross-seam evidence without repeating unrelated RU aggregates.
6. Preserve sequence's already-planned separate exact emitted strict-preparation control (2 expected/11 maximum HTTP, one source owner/listener, zero product children), with its sole named retained-member-conflict failure and actual500/empty/no-context prerequisites. Its derivative must be made from the **combined** emitted graph, not the old df80 copy. This control is part of the accepted sequence slice, not a new minimum control.
7. Minimum's prior five consistency tests, five owner-read tests, eleven repaired malformed-Type variants, and two repaired installed callbacks (10+18 HTTP) remain valid historical evidence under their precise original bindings: sequence edits none of their implementation files. Repeating all of them is not necessary solely for this disjoint integration. If root elects to claim those exact callbacks against the combined payload, run them explicitly and account for their additional work; do not infer new passes. The new minimum32-positive covers both repaired route families. Likewise no need to rerun the minimum two omission controls just to claim an unchanged oracle; their actual earlier control outcomes remain attributed to their old graph. Root may retain a stricter already-released final gate, but no such release is created here.

Keep all existing per-command/work/cleanup/wire limits, child/store ownership and no-retry rules. This proposal does not substitute for unfinished review, public qualification, or a separately requested whole-suite check. It does not authorize execution. No new main merge or publication gate is introduced.

## Exact union file hashes

For shared files, the listed value is the prospective in-memory composition above; for all other files it is the current frozen sole-owner byte hash. These pins are sufficient to reject lost PR140 paths, omitted repairs or unintended source churn at the later branch merge.
- `apps/bdptest/src/read-update-matrix.test.ts`: `376709ecdcf356a98c5bf7770e6b0a637ab6e06cf9a9f7ed68c2e5c15da9dd36` (minimum)
- `apps/bdptest/src/transactional-collection-minimum-installed.test.ts`: `615caab0e5a846c546e718f65c4041c5ca41bef59df1d56a6dea0e42638b0174` (minimum)
- `apps/bdptest/src/transactional-minimum-current-resource-installed.test.ts`: `21146871e6b5bfa3d9e96ff2db7c4897dd62fac026de538981832de295628f15` (minimum)
- `apps/bdptest/src/transactional-sequence-depth-installed.test.ts`: `d451340062915a31f79c1d02371bdbff13fb404967fe563e0484c43f218dbcf9` (sequence)
- `docs/design/transaction-implementation-progress.md`: `9668b52c1d5cfdcc7f2a326cafe3468d87de6cca14438b73ae88816587b99c28` (combined)
- `docs/specs/bdp.md`: `67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8` (minimum)
- `packages/conformance/matrices/read-update-v1.json`: `10a7c6f97a280cf88e58d3aa2039fe580904b74bfffcebda4747fc3c1a34b583` (minimum)
- `packages/conformance/src/profile-contract-plan.test.ts`: `3fec4ca8c6eee7a20bbcdd1aa53ef2a015989087ba2f8d6b37fd997dcae7dd64` (minimum)
- `packages/conformance/src/read-update-artifact-set.test.ts`: `e5c8903119b7c336dcc1e1479378e22029aa4d36978817557e461ea723991379` (minimum)
- `packages/conformance/test-support/read-update-coverage-v1.json`: `287c2a258afd04a95417f5de416c81bd8159c95762d71d568a64f0b402159f0f` (minimum)
- `packages/server/src/transaction-batch.ts`: `95f3ce3957503cba240f18e911fe3d2c8027b4accbba8f1128ea47e3d2b0c362` (sequence)
- `packages/server/src/transaction-consistency.test.ts`: `904d1f74ff06237da41c45f3dc1d4727a1387dc760950860ef548db51e7f02c4` (minimum)
- `packages/server/src/transaction-consistency.ts`: `ad51311d6098db32fd47849e0b78d3118162cdf1e80ee9ef4d46324bfa9f436f` (minimum)
- `packages/server/src/transaction-http.test.ts`: `f4353454c8f8a9dc42367395804bbe6309f7bcc51dba2b91b3176975d08b5cce` (combined)
- `packages/server/src/transaction-http.ts`: `4143e8a36021d60d1deba4649b1de5dfbcb3ee2d630486749932eccc1374c602` (combined)
- `packages/server/src/transaction-read.test.ts`: `58e745c2ade0804bf0ab0230c34ff5e522c28725fa77ecb9303dbe972162e0f2` (minimum)
- `packages/server/src/transaction-read.ts`: `7e2cc29b668bb80a3ba2948e90511518a98a7ff798b1dbfb8c6f51f0c0bd2a6f` (minimum)
- `packages/server/src/transaction-sequence.test.ts`: `687f1a9c10277d178db187b2c8efefe07d099755c473db4c2e7d87e4d7be6a4b` (sequence)
- `packages/server/src/transaction-service.ts`: `c700d5719918b60c3239a336eef02c9dff6717faca698777aba01181ebdcd278` (minimum)
- `packages/server/src/transaction-set.test.ts`: `a8a7d5b127866c2b1a60012e6c1f864b8e58bd524f6c520604e2b860b58d491e` (sequence)
