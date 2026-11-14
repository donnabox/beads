# Missing-prefix error contract reconciliation

The owner-run red-capable command was `scripts/test.sh -v -run '^TestLegacyImport' ./internal/storage/graphstore` on frozen2b, both embedded and ordinary shared-server Dolt2.1.8. Its log/result/source identities and nine raw hashes are retained. It fails exactly four missing-prefix apply/dry-run leaves plus three parents, no skips. Installed workflow families passed both modes at2b; parser/mapper were not run. This is NOT a successor PASS. No local rerun is authorized; Janet owns every runtime lane.

Ranked explanations reconciled by source:

1. Incorrect cross-layer SQL sentinel expectation: confirmed source path and captured error. Import first calls checkBinding; schema.go279–281 uses issueops.ReadConfigPrefix, whose missing row is deliberately mapped to storage.ErrNotInitialized and then represented as ErrInvalidStore with an Issue-configuration diagnostic. No sql.ErrNoRows reaches the import-specific reader on this failure path.
2. Broken import-specific inspection wrapping: not the failing path. legacyTargetInspectionError uses `%w` for raw SQL causes; the unchanged TestLegacyImportInterruptedInspection separately requires both ErrInvalidStore/sql.ErrNoRows for a direct prefix inspection. Keep that control.
3. Global binding cause-chain policy defect: not established by these four failures. Changing checkBinding or issueops.ReadConfigPrefix would broaden all graph callers and the established missing-configuration contract. No such change is made.

Repair only the existing real-engine missing-prefix matrix. Require ErrInvalidStore, Issue-configuration and missing issue_prefix diagnostics, zero result, no write-stage callback, and unchanged logical state after restoring only the deliberately removed prerequisite. Preserve apply/dry-run, both modes, operation contexts, cleanup and all remaining fault/uncertainty assertions. Save and restore the prior stage hook normally and in cleanup. This is an oracle reconciliation plus earlier-refusal regression, not weaker admission or a new error ABI. Future additional typed causes may be preserved without violating this test; no negative SQL-sentinel assertion is added.

Only one existing test file changes. Production, shared binding/issueops/error mappers/transaction policy, installed workflows, BUILD and dependencies remain byte-identical2b. Fresh source council and Janet-only exact-successor rerun remain mandatory; previous installed PASS does not qualify changed source. No PR86 retarget/main/default/release promotion, in-place migration or production input.
