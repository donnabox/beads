# PR83 native diagnostic integration addendum

**No actionable findings or blockers in this bounded integration delta.** Reviewed frozen `65ff2f00419d7aa94d84d56a7127d172d894b5c4` against `55ed77e3035647f7ab32030978eba6b01a0dc993`, using exact committed source from `/Users/dbox/repos/wt/vickie/bdp-provenance-runtime-integration`. Scope: 19 files, 563 additions/61 deletions, including the already-reviewed PR64 coherence/diagnostic repairs and the private Transactional formatter integration. Later author-head changes are excluded.

## Reviewed integration

- `member-numeric-diagnostics.ts` factors the reviewed PR64 formatter into a private helper, preserving its create-document detection, document-relative create pointers, metadata update omission of unapplied target pointers, stripping of metadata Type/schema fields, properties omission default, and pre-accounting discriminator/message normalization. It has no store, clock or public export. RU member preparation and private Transactional evaluation now both call it.
- `transaction-batch.ts` invokes the helper before numeric admission with the existing 16-entry/65536-byte limits. This closes the earlier integration dependency: Transactional metadata numeric errors no longer bypass mandatory document attribution. Property-update pointer formatting keeps its inherited caller convention; this bounded repair does not claim a whole-profile diagnostic-location redesign.
- The eight operation/document tests exercise actual admitted transactions, with an effectful prefix before the inadmissible number and a create tail after it. Both Beads and Links, create and update, properties and metadata are covered. They verify failure index/name, complete diagnostic shape and receipt schema, unchanged graph/groups/identities/position, durable exact receipt replay after reopen, and conflict for a different inadmissible literal under the retained key.
- The exact-byte test uses a repeated long parent path to reach a 65536-byte diagnostic list without violating request size limits; the third error proves truncation. Increasing the path forces the second diagnostic to be omitted. Both paths preserve the metadata discriminator, and the exact retained receipt survives restart. The count test retains sixteen complete diagnostics and marks the seventeenth omitted.
- Manual spec resolutions retain the four-part editorial layout and already-reviewed provenance vocabulary while adding the metadata diagnostic contract and coherent UpdateWhere, representation-bound, retained upsert and historical digest wording. The owned-Link updated-delta catalog citation uses `revisionProvenance`, not the PR64-only `attribution` name. Citation and profile-boundary tests pass.
- The mirrored schema diff against the integration baseline is only the optional closed diagnostic `document` field. Existing common provenance definitions are preserved. Graph interpretation hashes match their intended source; the review note accurately distinguishes diagnostic formatting from installed Type graph interpretation and from a fresh Read seal.
- The new private helper's four compiled output files are added to the exact installed package inventory. Public server exports and profile advertisement do not change. No source migration or retained-record rewrite is introduced.

## Independent validation

External `git archive` snapshot at `native-diagnostic-source/`, with local workspace package links and private build outputs. Node v24.16.0. Author checkout untouched.

- Private build `tsc -b`: exit 0, `native-diagnostic-build.log`.
- Required typecheck: exit 0, `native-diagnostic-typecheck.log`.
- **484 tests passed in 7 files**, `native-diagnostic-integration-tests.log`: transaction-store (including all ten new cases and existing crash tests), member-identity, member-executor, installed-schema graph policy, Transactional catalog, profile citation boundaries, Read+Update wire.

This is not a full-suite or installed smoke execution claim; package inventory was inspected and private build verified output production. Fresh final-source cohort qualification and subsequent Read/output-envelope repairs remain separate integration gates. No source edits, pushes, bus/PR messages or profile claims were made.
