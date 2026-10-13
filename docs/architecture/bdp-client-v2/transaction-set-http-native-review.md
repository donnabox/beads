# Native S2a implementation review

## Finding

**P2 — installed owned-Link positive oracle accepts an entirely suppressed mutation.** `apps/bdptest/src/transactional-set-installed.test.ts:322–330` compares the receipt Resource with its later GET and the receipt sourceRevision with the later owner GET. It never requires the requested literal `properties.n: 1`, a changed Link revision, a changed owning Bead revision, or an effectPosition for the first effectful command. Therefore mutually consistent stale outputs pass. The fabricated `sourceRevision: "wrong"` negative only demonstrates equality checking, not that the source was actually versioned.

Independent discrimination proof: after a clean emitted-source baseline passed, I changed exactly one expression in the isolated review clone's ignored `packages/server/dist/transaction-batch.js`. Only updateWhere on `/links/owned` with requested value1 was rewritten to the existing value0 before the evaluator. The entire installed test still passed (exit0). Actual captured original-request receipt showed `properties: {"n":0}`, no effectPosition and `requiredPosition: "pos-5"`; the unmodified baseline showed `n:1`, effectPosition `pos-6`. This violates the positive witness intended by accepted plan done-when8/11 while preserving the test's current comparisons. It is an evidence sufficiency defect, **not an observed defect in the submitted runtime**.

Minimal repair: capture the owner and Link immediately before the update; require exact intended properties `n:1`, unchanged metadata, changed Link/owner revisions and an effectPosition on the first successful receipt; retain the existing Resource/GET and sourceRevision/owner equality checks, repeated no-op checks and rollback checks. Keep the existing wrong-sourceRevision negative; add a bounded control demonstrating the literal/transition assertions reject stale observed state. No new runtime or protocol semantics are needed.

Exact counterprobe, originals, normal/mutated transcripts and compact receipt proof are under `transaction-set-http-native/`: `owned-noop-mutant.txt`, `transaction-batch.js.original`, `installed-baseline.log`, `installed-owned-noop-mutant.log`, `owned-oracle-proof.json`. The mutant is not packaged evidence and must not be interpreted as clean-source qualification merely because the ignored-dist mutation leaves git status clean. Its payload digest differs, as recorded. Parent accepted this bounded finding while the report was being finalized.

## Exact source and scope

Frozen head `2f7a1641c613680c5b4dfe056b933bfa6af869dd`, direct parent `96b00e0be63994b79025fc84ba82a36a55da0bb7`. Author checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-set-http` remained unchanged and clean. Independent detached shared clone `/Users/dbox/repos/wt/vickie/vickie-transaction-set-http-native-review` is at the same head. Read the entire eleven-file delta (1146 additions/44 deletions), accepted revised plan (SHA256 `798d21ebabc7377e97adf6f242b25d652b67638be2cd86704a632aef2d57aedc`), dispositions and root acceptance.

Production review covered mixed ordinary/set preparation and singleton wrapper depth; original actual byte admission; complete known-shape/label/key checks before unsupported Selector refusal; authorization before private gap; twelve exact private HTTP targets and method/query classification; numeric admission versus cardinality and cardinality-before-match capacity; staged selection and owned mutations; original-carrier numeric/cardinality/patch failure pointers and retained cross-carrier identity; shared compare/admit path and actual disconnect/reopen test. All changed app binding, installed witness, server tests and progress-document claims were read. No protocol/schema/catalog, public exports, store format, Selector parser or public activation change is present in this delta.

No additional actionable source defect found in this bounded review. Existing tests discriminate late malformed shapes after unsupported Selector, same-key correction/no admission, invalid versus admissible large numbers, singleton/batch pointer retention and depth/byte boundaries, staged multi-set behavior, pinned-endpoint incidence rollback and actual post-admission disconnect/replay. The owned installed finding above remains separate from these passing source/component checks.

## Independent commands and results

Pinned Node `v24.16.0`; pnpm `11.20.0` via `/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs`, with `/Users/dbox/.nvm/versions/node/v24.16.0/bin` first on PATH. All logs are outside the source checkouts.

- Remote session guard via `guarded-action.py /usr/bin/true`: exit0 before review work.
- Frozen dependency installation: exit0 (`transaction-set-http-native/install.log`).
- `pnpm build`: exit0 (`build.log`).
- `pnpm exec vitest run packages/server/src/transaction-submission.test.ts packages/server/src/transaction-set.test.ts packages/server/src/transaction-owned.test.ts packages/server/src/transaction-http.test.ts apps/bdptest/src/transactional-binding.test.ts --maxWorkers=2 --reporter=verbose`: **141 tests /5 files pass**, exit0,2.17s (`focused.log`).
- `pnpm exec vitest run apps/bdptest/src/transactional-set-installed.test.ts --maxWorkers=2 --reporter=verbose`: **1 test pass**, exit0, unmodified emitted-source actual HTTP baseline (`installed-baseline.log`). This independently runs the disposable fixture against built emitted files, not a new packed-cohort claim.
- Same one-test command after the explicitly isolated emitted-runtime countermutation: **1 test still passes**, exit0,814ms (`installed-owned-noop-mutant.log`). This is the demonstrated oracle failure, not successful behavior evidence.

The only modified emitted file was restored byte-for-byte from the saved original; SHA256 `10edbb4cf0e196cd000c4a19900875e46832fbb9a2b39bd4797f15a88375a1ff` verified. The intentional mutant SHA256 was `de93ffcfd9e96db43c8f96176fc00c95d565c3107296d29146f1799ea91b14a8`. Both source checkout statuses remained clean at exact2f7a164. Both installed runs completed their child.close/finally cleanup; process inspection afterwards found no review child or bdp-set-installed process. No outward writes or author-source edits occurred.

## Limits and disposition

This is source review plus the independent focused checks above, not a duplicate full repository run, not verification of the author's fresh tarball cohort, and not public Transactional qualification. The supplied author report's614 subsystem, packed TX/RU and static-gate results remain author evidence; I did not relabel them as my runs. Explicit unsupported negative Selector timing/complexity Q1/Q2, materialized unbounded candidate scan, cold discovery, fixed-policy limits and other recorded private-profile gaps remain open as planned. No new law for those cases was inferred. Root owns review disposition and any subsequent source repair/publication.
