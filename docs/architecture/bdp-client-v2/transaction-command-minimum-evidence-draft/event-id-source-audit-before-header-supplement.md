# Native Event-ID source audit

**Confirmed High stale-oracle defect in the new minimum row.** The literal pos-1-0/pos-2-0/pos-3-0 values are stored group Event IDs, not the IDs actually served by GET events/. Both the manifest ordered-facts assertion and the app's exact Event expectation would reject valid current serving. Earlier native clean review missed this distinction; this audit supersedes that conclusion on Event ID correctness. No test or runtime execution was performed.

## Exact source and path

Base453b6dab386746bb03aac2c0a293572b8a71c969. Frozen app421eeff168aa54b4d97893c81ebe1b9f790f9f4cf2e015a9fc6aac552b0aafb0; manifest a48612a0935e7fe557cd3f75af281cc87522c339c3f09c0fa6aff6ccd1671446. Source remained unchanged for this audit.

- transaction-batch.ts:644–659 creates stored scope-group Events with id equal checkpoint-plus-hyphen-ordinal.
- transaction-http.ts:492 calls service.events for the actual source request; transaction-service.ts:390–400 calls owner.readEvents.
- transaction-store.ts:1543–1601 captures current context, scans retained groups and calls eventProjection;:655 calls projectTransactionEventGroup.
- transaction-events.ts:267 pushes renderTransactionEvent(context, source.url, group.checkpoint, event). At:73–79 each tagged field is SHA256 of JSON.stringify([tag,value]), encoded base64url. At:96–100 rendering overwrites source and id. Its exact served ID is ev_<epoch hash>_<view hash>_<source hash>_<checkpoint>_<ordinal>.
- The renderer explicitly describes the spelling as private, not a fixture law. This is a private installed-test expectation, not a new normative Event-ID format.
- schemas/bdp-v0.schema.json event:3075 requires id and uses wireToken. Both spellings can pass that syntax; schema validation does not establish the correct served identity.
- New m-events ordered-facts projects /id against three stored-form literals. cumulative:2694–2709 independently pins the same wrong literals. App:2510 repeats them before completing the actual Event observation. The hook throws first; the runner's named manifest mismatch is not reached with a complete positive observation.
- Inspection of every existing /events tuple assertion found this literal /id column only in the new minimum row. Do not rewrite historical26 objects or broaden this repair to unrelated prior Event witnesses.

## Narrow existing-vocabulary repair

1. Keep the existing m-events ordered-facts assertion ID and exact three ordered tuples, but remove only its /id projection and the corresponding stored-form values. Retain source/type/subject/subjectType/ordinal plus all existing properties/metadata/provenance/next checks. The schema still requires IDs. The runner's existing ordered tuple evaluation at2360–2386 enforces exact cardinality/order.
2. In the actual app hook, independently compute each complete expected served ID from the already checked epoch/view, the literal Scope events URL, literal pos-1/pos-2/pos-3 checkpoints and ordinal0, with explicit tag strings epoch/view/source and SHA256(JSON.stringify([tag,value])).digest(base64url). Compare actual event.id exactly to that independently constructed string. Do not import the production renderer, derive expected checkpoint/ordinal/source from the Event under test, or use returned event.id as its own expectation. Existing createHash is sufficient.
3. Preserve complete Event.data/revision and transaction correlation to independently validated A/B receipts, all Resource/replay equality, exact Event identities/source/type/subject/subjectType/ordinal/count, and next:null. No raw-body or opaque revision capture is introduced.
4. Update the cumulative independently authored tuple ledger and terminal row digest to match. Keep coverage's exact manifest assertion mirror; it can remain192 IDs because the existing assertion is adjusted rather than added/removed. Clarify coverage/report attribution: exact served Event IDs are app-supplemental private source-bound checks, while manifest tuples establish literal semantic inventory. Do not pretend manifest alone now proves those exact opaque IDs.
5. Remeasure all affected canonical/derivative bytes/nodes/hash bindings before root accepts a final successor. Preserve the failed unexecuted snapshot and historical bases. No extra GET/action/static registration/runner API/runtime change. The local11actions/208expected225maximumHTTP/26children and time limits need not expand.

The existing hook context comparisons permit this without exposing report bodies: the same actual raw responses provide epoch/view, validated against earlier raw observations. This does not add raw-to-client action9 context equality or alter any unresolved Event membership ruling.

## Recommendation

Stop validation release until this bounded source correction and its exact revised measurements/code review are accepted. The current actual header supplement remains a review of its immutable header-repair snapshot; do not relabel it as review of a later Event correction. Root owns authoring/review disposition. No source edits, builds, imports, services or new reviewer launch occurred here.
