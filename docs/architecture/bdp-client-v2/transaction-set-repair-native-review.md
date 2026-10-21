# Private set-core repair: independent native review

**Verdict: bounded pass; F1–F5 closed as scoped. No blocking finding.** Exact frozen `a94ab9885762385cc823aecd7071bda14217d4b2`, parent `3e2b8784daefd5893f990785cfe91f1d872a2ffe`. Author source remained read-only; independent clone `bdp-transaction-set-repair-native` remains clean. No source/outward/evidence changes.

Read the author report, four-file diff, and prior review at `transaction-set-integration/claude-review.txt` (the root-level filename supplied in the handoff was absent). Reviewed preparation, expansion, visibility, error projection and existing owner rollback integration. This is private valid-set core evidence; HTTP set admission remains closed.

## Dispositions

- **F1: closed.** `transaction-set.ts:69–94` selects the existing matching update singleton syntax root; both delete collections use deleteLink's shared revisionProvenanceInput syntax. The synthetic Link subject is only in the discarded syntax-check request: the returned set input, normalization, identity and evaluator retain the actual set operation/collection. This accepts the bundle's deletion provenance without a second parser. The canonical-schema cross product covers both operations, both collections, cardinality and valid/invalid provenance variants. Actual owner tests execute all four operation/collection combinations; updates observe expected actor and value-state message, metadata retention and changed postimage. Deletes retain the subject's final live revision and tombstone identity, mint no subject version, and preserve surviving unowned endpoints. Bead-delete provenance remains inert. No owned-source fanout claim follows.
- **F2: closed within the disclosed bound.** `transaction-batch.ts:56,507–515,682–688` enforces an independent default4096 total matched-Resource cap across set operations, including repeated matches and no-ops, before adding the offending set's result entries. The lower-only evaluator seam rejects invalid/raised values. Overflow becomes a retained whole-transaction413 limit-exceeded with transaction.matchedResources and no operation index/name or partial result prefix. Existing owner savepoint handling rolls back prior staged effects. Lower-cap tests prove prior creation rollback and durable failed replay; an independent emitted-owner probe verifies the actual unmodified4096 boundary and overflow (below). Enumeration, visibility traversal and selection still precede the cap and scan/materialize the collection; whole-collection work/memory and receipt-byte bounds are explicitly not claimed. Event cap remains separate and unchanged.
- **F3: closed as an evaluator repair.** Candidate filtering now uses existing isResourceVisible with the same transaction, Scope and policy before Selector evaluation/cardinality. Its closure checks local endpoint Beads, unlike the old bare canRead. Both update/delete probes narrow per-turn readability while individual Links remain readable; the hidden-endpoint Link is absent from matches/results/mutations, and only the visible Link's Events appear. The owner installation remains fixed all-visible: mocked evaluator-option probes are correctly labeled, not installed variable-policy qualification. Owned/hidden-policy/aggregate completeness remains outside scope.
- **F4: closed.** BatchEvaluation now requires operationIndex on operation-local failures and allows omission only for the two implemented transaction-wide limit names, explicitly forbidding operationName there. Negative and positive type probes pass independent typecheck. No public receipt schema/union was widened.
- **F5: closed.** Module comment precedes the contiguous import block; constants/classes follow imports.

No runtime HTTP handler, owner format/configuration, compiler, public schema or package export changed. The actual shared selector, canonical ordering, staged evaluation and owner atomic rollback remain in use; no broader Selector refusal policy or product decision was inferred.

## Independent validation

Pinned Node24.16.0 / pnpm11.20.0; frozen install and completed build before tests. Independent typecheck passed, including the failure-union compile-time assertions.

**304 tests /4 actual files passed in2.12s**, maxWorkers2:

- packages/server/src/transaction-set.test.ts
- packages/server/src/transaction-http.test.ts
- packages/server/src/resource-evaluator.test.ts
- packages/server/src/read-selector.test.ts

The command also included transaction-batch.test.ts, which does not exist and selected no additional file; no separate batch-test pass is claimed. Set tests execute the actual batch evaluator/owner. Build, typecheck and test exits are0; logs `native-{clone,install,build,typecheck,tests}.log`, exit files, and tests JSON are alongside this report.

Additional independent emitted-owner probe `native-boundary-probe.mjs` / `.log` / `.exit`:

1.128 real Beads ×32 no-op sets =4096 matches completes with4128 result entries, no effectPosition, unchanged complete graph/group inspection.
2.Adding a129th Bead makes the same32 sets4128 matches: retained atomic match-limit failure, no partial results, complete graph/group state unchanged.
3.Close/reopen and exact-key replay returns the identical failure without re-execution.
4.Ordinary prepareTransaction still refuses a set operation; the private preparer alone admits the core test.

Probe passed with no lowered cap or evaluator mock; disposable store removed. Independent checkout clean after all checks.

## Limits

Did not independently repeat the author's558-test/14-file run, smokes, evidence verifier, or prior-format6 reverse opener probe. The latter was read and is correctly distinguished from the independent default-boundary observation above; this commit does not change persisted format/identity. Root owns combined qualification. No physical durability, whole-scan resource bound, public set capability, owned-Type integration, variable-policy installation or profile completion is claimed.
