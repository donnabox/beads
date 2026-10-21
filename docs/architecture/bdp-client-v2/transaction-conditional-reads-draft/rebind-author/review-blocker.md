# Frozen author snapshot blocked by native review

Root relayed a concrete code-review finding after author packet completion. A textual replacement intended for the new conditional finish join also changed the old assertPrincipalPositive helper at transactional-matrix.test.ts:4066. It now looks up principal assertions with conditionalActionIds[i], whose IDs differ. This is an authoring regression and blocks execution.

Read-only confirmation found exactly two conditionalActionIds[i] sites: the erroneous principal helper and the intended conditional finish join. The principal helper must restore its original mapping (or equivalent principal-owned ID lookup), but root explicitly requested no source edits until complete native findings and a correction release. The frozen source binding remains bf626872ba83bfafab4b957632247684d1169e0b9e3fbf6c627f8c4a644413d6; original author report, all measurements and failed candidate remain preserved. No product tests or imports ran.

The initial author report's preservation checks apply to parent raw objects, historical derivative transformations, and observation routing; they are not proof that every old assertion helper is byte-identical. This finding establishes that the old principal assertion helper was inadvertently changed. Root/native/Claude approval and validation remain outstanding.
