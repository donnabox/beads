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

## Independent native Codex closure review

Reviewed exact frozen `85c661543d4ccb68c1b426bf4355762bcd3f841c` against its sole parent `4a6c020296aac7e7331af8f52e527365786d83cd`. The entire five-file delta was inspected. Author checkout remained clean and unmodified. Native review and observable test discipline were applied; this is a bounded repair review, not a repeated whole-runtime or public-profile qualification.

**No new actionable findings or blockers in this delta.**

1. **Medium first-diagnostic overflow: closed.** `transaction-batch.ts:40–48,454–465` defines one frozen set of private bounds and checks the existing feasibility floor at initialization. The numeric emitter and Resource evaluator (`transaction-store.ts:658–663`) consume the same diagnostic count/byte values. The 64 KiB body bound is enforced before parse/key access in `transaction-submission.ts:61` and in all three preparation paths; the properties cap remains 64 KiB. The correction enlarges the budget rather than catching/swallowing an unrepresentable first diagnostic. Encoded messages, document markers, and create locations continue to count toward the budget. Exact-bound and one-byte-over controls now exercise 512 KiB, independently of the separate sixteen-entry truncation control.
2. **Low dead callback location: closed.** Both shipped callbacks remove only the unused `instanceLocation` assignment. The shared formatter still derives create document-relative locations and omits optional locations for unapplied updates. Request-location text and metadata discrimination remain intact.
3. **Low embedder callback type/schema-location concern: explicitly deferred, not resolved.** The preexisting custom callback behavior is outside this repair and absent from both shipped defaults. This review does not approve it for a general-purpose public write profile.

### Independent verification

A private `git archive` of the exact commit was built with Node 24.16.0. Tests and extra probes were written only into `/Users/dbox/repos/wt/vickie/bdp-integration-logs/diagnostic-budget-native/source`; no author source was changed.

- PASS: TypeScript build (`build.log`, exit 0).
- PASS: 171 tests across transaction-store, transaction-submission, and read-update-runtime-limits (`tests.log`). This includes 16 additional native raw `submitTransaction` probes at exactly **65,536 UTF-8 bytes**, covering batch/singleton × properties/metadata × create/update Bead/Link. Every case returns a retained `validation-failed` receipt with the exact first escaped diagnostic, serialized diagnostics over 65,536 bytes and at most 524,288 bytes, and no false truncation. Create locations remain document-relative; update locations are absent and request pointers appear in messages.
- Those raw probes check state equality across resources, groups, identities, and position, including rollback of an effectful batch prefix and suppression of the tail. They check replay before and after store close/reopen, equivalent exponent-case spelling replay, and a changed exponent's identity conflict. A 65,537-byte variant is refused with 413 before any key-admission call.
- PASS: a follow-up 32-test submission run (`raw-submit-tests.log`) includes an additional accumulation probe: a legal sub-64-KiB update would grow properties from about 40 KiB to about 80 KiB. It returns retained `limit-exceeded`, rolls back the prefix, preserves state, and replays after reopening. The 32-test run repeats the submission subset; counts are not additive unique coverage.
- PASS: dependency-boundary check for all 11 workspace packages (`boundaries.log`). Static local import traversal finds no back-edge from the newly imported limits module to transaction-batch (`import-graph.log`). Package export map/public declarations are unchanged; the constant remains an internal module export.

An initial extra-probe run used the incorrect success status `succeeded` instead of the actual `completed`, so its 16 probes failed before exercising the repair. That harness mistake was corrected only in the private archive; the first log is preserved as `probe-harness-setup-error.log` and is not counted as product evidence. Subsequent runs completed cleanly. No full suite or fresh Read cohort was run by this reviewer.

### Evidence and scope integrity

The five-file delta contains no Read schema, observer, catalog, manifest, evidence artifact, profile advertisement, identity codec, or private-store format change. Therefore this review does not relabel the historical Read observations as observations of 85c6615. Existing source/evidence ancestry remains the parent's responsibility for final integration. Build and tests above independently verify this repair only; worker-reported broader checks are separately recorded earlier in this report.
