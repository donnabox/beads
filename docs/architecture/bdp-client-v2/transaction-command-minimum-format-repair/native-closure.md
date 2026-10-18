# Native formatting-only closure

No actionable findings. The candidate preserves behavior and all previously reviewed command-minimum source coverage. It is suitable for root's bounded formatting application only after the existing external-review hold is released; this closure does not authorize that application or test execution.

Read the complete correction.patch, report, ast-equality.mjs/json and summary. Independently verified every artifact hash, input/output source hash, and exact unified-diff reconstruction from the retained copies. Every input copy equals the corresponding still-frozen checkout file; actual checkout diff remains `8f0123c9484a74a5c4995dea84833561056f603c8da64d5f9139107b85d44df8`.

The three changes are line wrapping, indentation, trailing commas and formatting of an existing parenthesized TypeScript cast. The large installed-TX hunk is the same withInstalled callback reindented when its argument list becomes multiline; its undefined signal and observation label remain in the same argument positions. No title, fixture, request, expected result, assertion, route, timeout, control diagnostic or cleanup action changes. A–E/RU/S/Q selection, source/control89 expected/max119 exchanges and separate freshTX18 remain unchanged.

The retained parse-only equality program parses both complete files as TypeScript modules, compares full normalized JSON strings, and exits nonzero on inequality. It removes positions/comments/extra parser metadata, retains literal values/operators and TemplateElement raw/cooked values, and does not import product modules. The saved result has three equal AST pairs. Manual patch inspection also checks the details omitted by normalization: no comment/directive was added or removed and there is no semantic parenthesization change. This is static formatting evidence, not a runtime pass; I did not rerun the parser or formatter.

Bindings:
- correction.patch: `5d5104f9ada2652d9dd821792d59027bc0ea374d25626db33eeed10e1cace888`.
- ast-equality.mjs: `56c72757b40fa4f13be374a17f6522075c48ee9a48d0066f1b94acb2e61798e3`.
- ast-equality.json: `3c6a41735fa5a659249cf4a696d8b93c662d39633ebfdca710f12b91959a712f`.
- report.md: `996f7f0664b694590cd869e606dd2355d28ab49586400f716e62ad3eff5804df`.
- summary.json: `e78064e9ca369fc7fbdc70da66cd8b0721934edb3ecf9077ca872a725867efc5`.

| Path | Reviewed input SHA256 | Candidate output SHA256 |
|---|---|---|
| `apps/bdptest/src/installed-reference.test.ts` | `447a8b5fb3ee03ec4f2a894b02a1bb0fdb74ccdea0ad35576da78942373e0cae` | `2c87160890b1c8aed2ada117a1aa293fe7fcfa85db794a3d9ef6514a39d199d1` |
| `apps/bdptest/src/transactional-collection-minimum-installed.test.ts` | `5349dab51055d29d30bb80f286cb30e69947882bd07ed1eaa831b26fd0567174` | `a3f772525ee1c87f3c488e37914af8114421dff5cc003d81640182d6f7a05903` |
| `packages/server/src/transaction-http.test.ts` | `b4ed09796b4dc28a51d9cc73c04a677059434a606f09dddffa46eefaa432ad8e` | `8f66df9f9cf891383a19a67fb50c85708eaae055c59a661777b7110060b0c56f` |

No checkout edits, source authoring, tests/build/install/services or new council were performed. Only this closure report was written after the role-session guard. Initial format failure and all frozen original evidence remain preserved. After any application, source/diff/manifests and gates need their actual successor hashes rebound by the owner; old byte hashes must not be relabelled as the formatted successor.
