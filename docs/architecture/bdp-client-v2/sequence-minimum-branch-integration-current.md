# Current sequence/minimum branch integration map

The original integration plan is preserved. This successor binds clean minimum commit 5ca22ec4e710cc4417a80d27bff69e1d8424c115 (ordinary child of2e1676, tree 2747663746fde13b8b0b8c7ef3c78f7d8ca9f2a8) and the applied sequence correction binding b937a736c8b4b16c10346b0f8d998640fe31415a6ee79a10188945cb4324c81a on dirtydf80.

All two minimum repair blobs and all seven sequence blobs match their exact bindings. The union remains20 paths: incoming minimum11 plus inherited PR140five, sequence7, three overlaps. No paths, request budgets or validation selections were added by refreshing these hashes.

All three shared composed hashes are unchanged from the original map. Only the three sole-sequence correction hashes change: transaction-batch.ts, transaction-set.test.ts and transactional-sequence-depth-installed.test.ts. The intended merge remains root-owned ordinary minimum-working-branch into sequence-working-branch after exact author commit/review disposition; no main prerequisite or chasing.

Computed in memory only. No source files, Git index/objects, branches or bound review inputs changed; no merge/build/tests/imports/services/reviewer invocation occurred. Actual future sequence/merge commit and tree remain unfilled. At integration, compare all20 hashes below and preserve every outside-union base entry/mode. Preserve both full EOF appends in minimum-then-sequence order, rather than taking one side wholesale. Prior branch attribution and failures remain historical.

Machine-readable companion: sequence-minimum-branch-integration-current.json, SHA 6180b009216b3be97fe37862c51d97c664407c9b0a5a11acffe030bd5096f050.

| Path | Intended SHA-256 | Origin |
|---|---|---|
| `apps/bdptest/src/read-update-matrix.test.ts` | `376709ecdcf356a98c5bf7770e6b0a637ab6e06cf9a9f7ed68c2e5c15da9dd36` | minimum-commit |
| `apps/bdptest/src/transactional-collection-minimum-installed.test.ts` | `615caab0e5a846c546e718f65c4041c5ca41bef59df1d56a6dea0e42638b0174` | minimum-commit |
| `apps/bdptest/src/transactional-minimum-current-resource-installed.test.ts` | `21146871e6b5bfa3d9e96ff2db7c4897dd62fac026de538981832de295628f15` | minimum-commit |
| `apps/bdptest/src/transactional-sequence-depth-installed.test.ts` | `8ad2c796bfb22375584fc6b348d494ece512490f8a59ef72a74fb34de3f2fca1` | sequence-working-source |
| `docs/design/transaction-implementation-progress.md` | `9668b52c1d5cfdcc7f2a326cafe3468d87de6cca14438b73ae88816587b99c28` | combined, unchanged |
| `docs/specs/bdp.md` | `67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8` | minimum-commit |
| `packages/conformance/matrices/read-update-v1.json` | `10a7c6f97a280cf88e58d3aa2039fe580904b74bfffcebda4747fc3c1a34b583` | minimum-commit |
| `packages/conformance/src/profile-contract-plan.test.ts` | `3fec4ca8c6eee7a20bbcdd1aa53ef2a015989087ba2f8d6b37fd997dcae7dd64` | minimum-commit |
| `packages/conformance/src/read-update-artifact-set.test.ts` | `e5c8903119b7c336dcc1e1479378e22029aa4d36978817557e461ea723991379` | minimum-commit |
| `packages/conformance/test-support/read-update-coverage-v1.json` | `287c2a258afd04a95417f5de416c81bd8159c95762d71d568a64f0b402159f0f` | minimum-commit |
| `packages/server/src/transaction-batch.ts` | `0f67235a7fdadf3ae342695b6f8eee795ad51b70e6507dcb7d70ec8848ba9836` | sequence-working-source |
| `packages/server/src/transaction-consistency.test.ts` | `904d1f74ff06237da41c45f3dc1d4727a1387dc760950860ef548db51e7f02c4` | minimum-commit |
| `packages/server/src/transaction-consistency.ts` | `ad51311d6098db32fd47849e0b78d3118162cdf1e80ee9ef4d46324bfa9f436f` | minimum-commit |
| `packages/server/src/transaction-http.test.ts` | `f4353454c8f8a9dc42367395804bbe6309f7bcc51dba2b91b3176975d08b5cce` | combined, unchanged |
| `packages/server/src/transaction-http.ts` | `4143e8a36021d60d1deba4649b1de5dfbcb3ee2d630486749932eccc1374c602` | combined, unchanged |
| `packages/server/src/transaction-read.test.ts` | `58e745c2ade0804bf0ab0230c34ff5e522c28725fa77ecb9303dbe972162e0f2` | minimum-commit |
| `packages/server/src/transaction-read.ts` | `7e2cc29b668bb80a3ba2948e90511518a98a7ff798b1dbfb8c6f51f0c0bd2a6f` | minimum-commit |
| `packages/server/src/transaction-sequence.test.ts` | `687f1a9c10277d178db187b2c8efefe07d099755c473db4c2e7d87e4d7be6a4b` | sequence-working-source |
| `packages/server/src/transaction-service.ts` | `c700d5719918b60c3239a336eef02c9dff6717faca698777aba01181ebdcd278` | minimum-commit |
| `packages/server/src/transaction-set.test.ts` | `494f7536fac5fd452521902301d943865e40b20895fa16e9c0a3beb1b4481dc9` | sequence-working-source |
