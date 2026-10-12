# PR83 diagnostic first-entry feasibility repair

Frozen clean commit: `85c661543d4ccb68c1b426bf4355762bcd3f841c`, direct child of published `4a6c020`, in `/Users/dbox/repos/wt/vickie/bdp-pr83-council-repairs`, branch `codex/vickie-bdp-pr83-council-repairs`.

## Actual finding and repair

Claude correction finding1 reproduced before the implementation change:16 tests failed with JsonDiagnosticBudgetError, covering batch/singleton × properties/metadata × create/update Bead/Link. Each request is exactly65536 UTF-8 bytes. A long slash-containing JSON member name requires pointer escaping, making the first serialized diagnostic exceed the former65536-byte cap.

One immutable private Transactional bounds record now defines requestBodyBytes65536, propertiesBytes65536, diagnosticBytes524288 and diagnosticCount16. Existing snapshotReadUpdateRuntimeLimits validates this record at module initialization; its minimum first-entry bound is283906 bytes. Batch/singleton/sequence parsing, submission body admission, numeric admission and store Resource evaluation use the same applicable fields. No protocol field, advertised profile, request cap, property cap, schema, identity codec or private storage format changes. Existing retained failures remain replayable verbatim; the larger bound does not reinterpret their results.

Removed the dead instanceLocation property from both shipped reference formatter callbacks. The shared formatter derives create locations and omits unapplied-update locations, unchanged.

The private bounds live in transaction-batch.ts and are imported only by internal store/submission/tests; no package public API export was added. The new dependency on read-update-runtime-limits leads through the RU evaluator/executor graph, with no Transactional back-edge.

## Tests and evidence

- RED:16 failing new legal-maximum tests before source repair,107 other owner cases filtered. `pr83-budget-red.log/.exit` (exit1).
- GREEN:314 tests in7 files,2.91s, maxWorkers2. Includes16 exact65536-byte requests with first diagnostics larger than65536; each becomes a retained validation-failed receipt with exact message/document/location, valid wire schema, unchanged state, batch prefix rollback/tail suppression, and identical replay after reopen.
- Existing byte controls adapted to actual524288 budget: metadata create and properties update each retain16 diagnostics at the exact bound; increasing the shared parent by one byte leaves15, proving byte truncation; separate17-entry count test remains. All raw requests remain within65536.
- PASS:build, typecheck, dependency boundaries, format, lint, development smoke and installed smoke.
- PASS:evidence:verify, preserving actual Read source e5f7d0c/evidence dbaf63a/constant674982be2123b9b7aae630bf79c3f774122900cd. No evidence edit or new Read observations; this is a write-path repair plus removal of unreachable callback fields. No claim that historical cohort observed this later implementation.
- PASS:git diff --check; clean frozen tree. No full-suite rerun for this bounded worker repair. No interrupted/timed-out run counted as pass.

Logs under `/Users/dbox/repos/wt/vickie/bdp-integration-logs/pr83-budget-*`; structured focused result `pr83-budget-tests.json`.

## Disposition and limits

External medium1 and low2 addressed. Low3 (embedder-supplied type/schemaLocation retained on numeric property failures) is intentionally unchanged: preexisting callback behavior, absent from the two shipped formatters, and root explicitly deferred a broader contract repair. Track before general-purpose public write qualification; do not treat it as resolved.

No external writes. Root owns native/external re-review, publication, exact-head CI and any further qualification. Existing partial/private profile limits remain unchanged.
