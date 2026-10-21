# PR65 native bounded repair review

**One actionable P2 finding remains; the current-source projection pin is a separately acknowledged final-integration gate.** Reviewed exact `11fe83926bace2c2a89a9b6f97cde7b02ce67229` against `9da223a` in `/Users/dbox/repos/wt/vickie/bdp-revision-provenance-repairs`: 12 changed files, 63 additions/34 deletions. No author-checkout edits or outward actions.

## Finding

1. **P2 — `packages/protocol/src/history-values.ts:62`.** The exported output type still accepts an empty object: `{ readonly actor?: RevisionActor } & RevisionContext` combines an optional actor with the branch forbidding all context fields. Consequently `const invalidOutput: RevisionProvenance = {}` compiles, while the same commit's output schema and `parseRevisionProvenance({})` reject it through `minProperties: 1`. This leaves the repaired public TypeScript model inconsistent with the output wire contract and lets callers construct an output record the parser will refuse. Require an actor in the context-absent branch, keep actor optional in the complete-context branch, and add an `expectTypeOf<{}>().not.toExtend<RevisionProvenance>()` assertion. Do not tighten `RevisionProvenanceInput`: empty input remains valid and normalizes to omission. A reviewer-only probe in the private archive independently compiled the invalid empty output with exit 0 (`native-empty-output-probe.log`). Parent was notified.

## Verified repair scope

- The output schema's nonempty constraint is mirrored exactly and applies through the shared provenance definition. Actor-only, complete-context-only and actor-plus-complete-context shapes remain admitted. Context remains all-or-none and old/dual top-level envelopes remain rejected. Input schema is unchanged and explicitly accepts `{}`.
- The TypeScript change correctly rejects partial context groups and retains the complete group's existing value-state encodings. The remaining empty-output issue above is confined to the context-absent type branch.
- The historical packet restores the exact old `attribution.principal/status` bytes for its two altered vectors. Independently recomputed all seven packet SHA-256 values over the displayed serialized lines: every value matches. This preserves historical vectors; it does not reintroduce old wire aliases into current schema or parsers.
- History input fixture descriptions now match the approved optional actor input and caller-timestamp prohibition, with an actual actor-bearing positive fixture. Catalog/spec corrections describe no-op retention of the entire envelope and preserve catalog IDs and profile scope.
- The member-identity change is commentary only and now accurately points to recovery format 3 as the already-reviewed refusal boundary; it does not change request identity or retained results.
- Refreshing the bundle hash while retaining `schema-resource-index-2` is justified for installed graph interpretation: only Resource/History/Event output provenance validation changes. Installed Type schema graph admission, decoding, resolution and budgets are untouched. The graph pin check passed independently. This is not a claim that the public Read projection remains unchanged.

## Independent tests and open integration gate

Node v24.16.0, exact-commit external archive `native-repairs-source/`.

- Required typecheck passed on the unmodified archive (`native-repairs-typecheck.log`). The separate reviewer-only empty-output probe also passed, demonstrating Finding 1.
- `native-repairs-tests.log`: **157 passed, 2 failed, 7 files** — schema bundle, History wire, Read projection, Read catalog, Read+Update catalog, Transactional catalog and installed schema graph policy.
- Both failures are the known current-source Read projection regression pin: expected `387cd60fc2a253885ae7cedbb9ee491a3dfd832af21e8c093664047f13c41dac`, observed `bd7da0fc3f8082e0bfb6f4969c04409f2a610e12408dbc6a0bd14b4b351127cc`. Parent explicitly reserves this pin update for the final combined PR64+PR65 source, whose ultimate digest may differ. It must pass before final qualification; it is not confused with historical evidence that intentionally awaits a fresh cohort.

This report does not claim the entire tree is green, a full test suite ran, or fresh evidence was sealed. After the type-branch repair, a bounded verification can close Finding 1; final projection and cohort checks remain with the integration owner.

## Type finding closure at `85beebd37c553df830da117ef6d7ad0891030030`

Reviewed the two-file child repair: `RevisionProvenance` now has an actor-required/context-absent branch and an actor-optional/complete-context branch. Its type regression tests reject `{}` and partial context while admitting actor-only and both complete-context variants. Input remains unchanged. In the external archive, the previously accepted reviewer probe now fails exactly with TS2322 (`native-empty-output-probe-fixed.log`); after removing that reviewer-only probe, required typecheck passes (`native-output-fixed-typecheck.log`) and all **14 schema-bundle tests pass** (`native-output-fixed-tests.log`). Finding 1 is closed. The previously acknowledged final combined-source projection pin/evidence gate remains separate.
