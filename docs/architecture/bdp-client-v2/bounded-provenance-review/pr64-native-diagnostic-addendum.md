# PR64 native diagnostics repair review

**No actionable findings or blockers in the PR64 repair delta.** Exact reviewed commit: `aeb19286269f922bc9f514bba39703c3cce67608`, direct child of `c1712f2d26f889f723ffb79c01146c87883a65df`, in `/Users/dbox/repos/wt/vickie/bdp-pr64-diagnostic-document-repair`. Scope: 13 files, 336 additions/28 deletions. No author-checkout changes or outward actions.

## Verified behavior

- The mirrored `validationDiagnostic` schema and TypeScript type add only optional `document`, closed to `properties` or `metadata`. Omission continues to mean properties. Unknown values, null and invalid JSON pointers are rejected; old omission-form diagnostics remain accepted. There is no new request member or request-identity change.
- Create numeric diagnostics derive a document-relative pointer from the actual lossless input location. Metadata creates force `document: metadata`, identify metadata in the message and strip property-Type/schema fields, regardless of the supplied formatter. Property creates retain the existing omission shape.
- Update numeric diagnostics recognize the actual `/metadataChange/` input path and force metadata attribution while omitting the unapplied target pointer. Property update formatting keeps its existing caller convention and cannot be relabeled metadata by the callback.
- Metadata patch failures for root removal/replacement, missing targets/parents and invalid array indices identify metadata and retain the relevant document-relative pointer. Property failures retain their prior representation. A failed metadata patch leaves both documents, writes and revision allocations unchanged.
- Both numeric admission and evaluator accumulation snapshot the closed diagnostic fields including `document` before measuring the complete UTF-8 JSON-array cost. Exact-boundary acceptance, first-entry-too-large refusal, count truncation and retained discriminator behavior are covered. Mandatory metadata information is not appended after budget accounting or dropped to fit a bound.
- Create/update numeric failures on both Beads and Links happen before effects, retain the complete metadata diagnostic and replay it byte-equivalently at the stored-disposition level after reopening. Legacy diagnostics remain parseable, and existing retained-result behavior is unchanged.
- Graph policy pin refresh is justified: the installed-schema graph uses unchanged lossless decoding, not the changed numeric diagnostic formatter; the added validation-diagnostic member does not change Type-descriptor graph interpretation. Keeping `schema-resource-index-2` does not certify a fresh Read cohort. Pin hashes and unchanged Read projection passed independently.

## Independent validation

Node v24.16.0; external exact-commit archive at `native-diagnostic-source/`. Tests used source aliases and did not modify the author checkout.

- `native-diagnostic-tests.log`: **551 passed, 5 files** — member identity, member executor, Resource evaluator, Read+Update wire and numeric JSON admission.
- `native-diagnostic-pins-tests.log`: **23 passed, 3 files** — installed-schema graph policy, Read schema projection and schema bundle.
- Total: **574 tests passed in 8 files**. This review does not claim the author's full build/smoke/evidence runs as independently executed.

## Separate downstream integration requirement

When this repair is incorporated into PR83, adapt the private transaction numeric formatter too. Frozen PR83 source `8c78d3e198561c242286a185cd8cb0424f9b5b83`, `packages/server/src/transaction-batch.ts:439–442`, directly invokes `admitReadUpdateOperationNumbers` with `instanceLocation: pointer` and no discriminator. It bypasses the repaired member preparation formatter. A plain merge would retain ambiguous metadata numeric errors there and operation-relative create locations. Add create/update metadata and exact diagnostic-budget checks with that adaptation. This module is absent from the PR64 repair tree, so the observation is an integration dependency, not a finding against this bounded commit. Parent was notified immediately.
