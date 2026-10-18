# Native closure: conditional principal-helper correction

**N1 closed. No further finding in this one-line correction.** The complete assertPrincipalPositive helper is byte-identical to committed HEAD 1ba65f2f192cc84e44eb09ef89b9693ec6cdaf05, including its original positional assertion lookup and pre-existing exact principal action-order check. The helper no longer consults conditionalActionIds.

Reversing only the corrected line inside that helper reconstructs the exact app SHA256 from the frozen rebind-author review. The other three owned source files remain byte-identical to that review. Current git diff matches the successor patch digest. This verifies the claimed bounded correction rather than merely observing that the bad expression disappeared. Prior broader native review remains applicable with its sole N1 resolved. This is code closure only; actual Claude disposition and root-controlled validation remain separate requirements.

No tests, imports, builds or source edits. Guard verified the same role session before writing this report.

Successor full patch SHA256: bc7e620ea97a37933901ee4bf2e2d818685178bd974a262e88b025e513e2a60a

Principal helper SHA256: 09162bdefa50d64de65900956848ac22e8db8d46bff5ff40ac1650030ca793e0

- `packages/conformance/test-support/transactional/manifest-v1.json`: `60c527f6a5673161eacf022c0015d09bea2743d153b8b0d0c86160b76fafeb68`
- `packages/conformance/test-support/transactional/coverage-v1.json`: `33032fe07eb06d2a8968fcd8e5a3cf93cb4637c551868c22da9ea0c4edd587a3`
- `packages/conformance/src/transactional-cumulative.test.ts`: `9f70f8a24ea358b2cd613fd392be68057abd7b1552ec8cc6c4bc1a3241de69f6`
- `apps/bdptest/src/transactional-matrix.test.ts`: `136de49e859546e081c39410f1d3ee6fc01e32faf1879084718ad67b553d62ec`
