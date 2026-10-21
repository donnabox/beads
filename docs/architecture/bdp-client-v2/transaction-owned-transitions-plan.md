# Next bounded TX slice: owned-Link transitions

Read-only plan against frozen cbecfeac9b2279e7424d2499d558e976ff3f9dcf in
vickie-transaction-contracts. Sole fork design remains authority. No source edits,
tests, new protocol rulings or implementation authorization are implied by this memo.
Implement only after root/council scope, from the reviewed integrated successor.

## Recommendation and prerequisites

Complete the existing private atomic owner's owned-Link transition projection,
reusing Resource evaluator behavior. This is one small vertical seam: permit the
installed owning descriptors, retain the evaluator's subject/source result, capture
both Resources' final group state, and append the mandated source Event. Exercise
singletons, batches and the existing private valid-set core. No new HTTP set route,
profile/discovery claim, policy configuration, Type lifecycle or compiler work.

Predecessors: reviewed compiled-contract cbecfea; reviewed root recovery-low5e;
set repair 064db8a99629122f130a91ee8b5189432ffe97e9 from update_implementation
(failure type preserves required Problem members, cardinality before matched-count
charge). Its external review is pending; root must integrate/accept it first. Their ordinary integration
should precede authoring because both sets and owned transitions meet in batch.ts.
The set author confirmed current accounting implements matchedResources and
inducedEvents only, NOT mutatedResources. Do not invent or conflate a byte budget.
RU artifact-retention work owns its store/installation files; no overlap is needed.

## Existing laws and actual implementation

Normative source: docs/specs/bdp.md at cbecfea, with schema bundle authoritative for
wire shapes (not the illustrative Markdown type sketches).

- 411–504, Owned Links: ownership is the source Bead Type/Link Type pair, explicit
  or wildcard; first-class and inline Link records agree exactly. Explicit empty
  buckets persist, wildcard-only buckets exist only when present, never a literal
  wildcard key. Source versions, target does not (self-Link is still one source).
  Visibility is closed over owned Links and in-Scope targets.
- 2154–2166: expectedRevision guards the addressed Resource's current revision.
  sourceRevision is an OUTPUT, not a second invented source CAS input. Link pins
  are references, not such a guard. Shared evaluator already supplies CAS ordering.
- 3858–3878: every operation sees prior staged effects, leaves valid staged state,
  and all graph/receipt/group effects commit atomically in Scope order.
- 4250–4278,4907–4918: sets select at their operation turn, over the whole visible
  canonical collection; provenance fans out per selected Resource and owning
  source; nonwritable selected Resources fail the whole operation. Global graph
  constraints use complete authority state. Cardinality is selected Resources.
- 5140–5168: results retain per-operation intermediate postimages; owned operations
  carry source+sourceRevision even on deletion/no-op. All-noop receipts omit effect
  position and produce no group. Set matched entry precedes canonical-ID results.
- 5437–5508: source updated Event uses ownedLink, never direct change/metadataChange.
  Creation embeds the complete new Link; update embeds id/type plus exactly the
  Link lifecycle delta; deletion embeds final live Link identity, not properties.
  Event-level revisions/provenance are the source's; nested revisions/provenance
  are the Link's. New Link and source versions share the operation's attribution.
- 5477–5495: one operation's facts order lifecycle, graph source then target, owning
  source updated last. Self-Link retains both endpoint facts. No-op emits none.
- 5695–5715,5943–5957: normalize group changes to each Resource's FINAL postimage or
  tombstone, include Link and owning source, enforce inline/first-class agreement.
- 5194–5230: retained receipt records project under current visibility including owned
  closure. Current owner is fixed all-visible; enabling owning Types must not pretend
  it adds variable grants/withholding/erasure or full retained-view machinery.
- 5250–5269: count/Event cap failures retain transaction-wide failure framing; no byte
  count substitution for matched or mutated Resources.

Actual source:
- resource-evaluator.ts:570–601 exports isResourceVisible: transitive owned/endpoint
  closure, cycle guarded. 854–974 computes ownership/no-op and sourceRecipe;
  998–1050 materializes full subject/source once, policy checks both, then CAS;
  1051–1100 writes both only after every check. ResourceEvaluation returns outcome,
  changed local IDs and deleted local IDs. sourceRecipe is private, not a new API.
- Existing resource-evaluator.test.ts:326ff,1251–1397,1565ff already demonstrates
  metadata-only fanout, source-only versioning, provenance/no-op, wildcard bounds,
  sorting/self-Link, hidden-target closure, source write refusal and deep equality.
- transaction-batch.ts:514–617 currently applies that evaluator, then throws whenever
  outcome.sourceRevision exists, records only direct subject in changes, and emits
  only Link lifecycle/incidence facts. This is the main integration gap.
- At cbecfea only, batch.ts:648–660 still uses direct policy.canRead. That branch
  predates integrated a94 visibility work: root8d/current637 already use existing
  isResourceVisible closure. The implementation predecessor must preserve that
  accepted predicate; this is a regression/reuse obligation, NOT a new missing fix.
- transaction-store.ts:183–197 rejects owning descriptors. The same installation
  already drives Read and evaluator registry; remove only owned refusal when the
  batch seam is complete. Its terminal path1040ff already commits graph, group,
  receipt and page plan together; it requires no second authority or write loop.
- store.ts:601ff projects receipts under its fixed all-visible policy. No current
  mutable policy or history-erasure config is claimed. Read plane already handles
  owning records and uses the owner configuration; no Read engine changes needed.

## Smallest implementation

1. Keep public and private request syntax/identity unchanged. Use existing evaluator
   once per explicit or selected Resource. Capture before/after write facts through
   a LOCAL wrapper of its existing tx.putResource inside applyResource: before the
   first write of each ID, read that existing row from the SAME staged transaction;
   forward the original write unchanged. On evaluator success, read the changed IDs'
   materialized postimages, using existing readStoredResource. No exposure of
   sourceRecipe, duplicate endpoint/ownership resolver, clock call, revision mint,
   nested execute, outside callback or later owner.inspect.

   This wrapper solves the crucial previous source revision problem without guessing
   from a later final state: after evaluator returns, the source has already changed.
   Do not retrieve its preimage then. Keep captured rows local to that operation;
   freeze/parse through existing helpers before storing Events. No await/capability
   escape. Existing outcome.resource/before supplies direct subject lifecycle facts.

2. Remove sourceRevision rejection. Push evaluator outcome intact (including unchanged
   sourceRevision on no-op), preserve label binding. If changed/deleted are empty,
   return with no Events or changes. For an effectful owned outcome, verify matching
   same-turn source before/postimage, Bead kind and outcome.sourceRevision. These
   are internal invariants, not invented request errors.

3. Evaluator changed/deleted IDs are Scope-relative paths, used ONLY for same-turn
   tx.resource(localId) lookup. The changes Map is ALWAYS keyed by the absolute
   canonical readStoredResource(row,scope).id or outcome.deleted.resource.id. Never
   mix local paths and canonical keys: each Resource must normalize to one entry.
   Identify the owning source local path from validated outcome.source and Scope.
   For the direct deletion, use the before snapshot already captured before evaluator
   execution and verify its canonical identity/revision against outcome.deleted;
   the post-deletion resource lookup is necessarily absent. Reuse captured before
   rows for other invariants, refusing internal inconsistency. Later mutations/deletion
   overwrite that Resource's map entry, while earlier receipt results and Events
   stay immutable. Do not synthesize source updates from final transaction state.
   Changes array order is not semantic; no requirement to reorder its Map insertion.

4. Derive the Link lifecycle delta once. Construct the nested ownedLink delta
   FIELD BY FIELD (id,type,previousRevision,revision, optional change/metadataChange/
   revisionProvenance), not by unconstrained spreading of Event data. At least one
   patch field must exist for an effectful update: accepted update inputs have no
   whole-value properties/metadata member, and no-patch inputs are no-ops that exit
   before Event emission. Assert this invariant, including the source updated data's
   exclusion of change/metadataChange alongside ownedLink. After existing incidence facts,
   append source updated with source previousRevision/new revision, ownedLink,
   and source provenance. Creation uses full Link postimage; deletion uses identity.
   Preserve absence/removal of provenance rather than copying an earlier version.
   Reuse scopeEvent so the new source fact counts against the same4096 Event cap;
   lifecycle + two endpoint facts + source updated is four facts for local creation/
   deletion, including self-Link; update is two. No new Event types or log.

5. Preserve the accepted root8d/current637 set selection predicate using existing
   exported isResourceVisible(tx,record,scope,policy), before Selector/cardinality/
   matched counting; no replacement or newly claimed fix is needed.
   Reuse the SAME applyResource for each frozen canonical-ID selection. Each owned
   Link versions the source separately, including multiple selected Links of one
   owner. Subsequent operation sees all staged state. A selector that ceases to match
   after an earlier selected change does not rescan/redefine its frozen member set.
   Preserve repaired cardinality/cap ordering and failure type, not an alternative
   set parser/limit path. HTTP sets stay closed.

6. Enable owning installed descriptors only after these seams pass. A compiled
   owning fixture must bundle every explicitly owned Link Type descriptor (and its
   referenced endpoint Types) and obey existing capture dependency/max limits;
   do not diagnose an incomplete fixture as a batch failure. Preserve schema-
   backed compiled registry, descriptor exactness, runtime limits and pinned capture.
   Schema-free owning fixtures can use the existing descriptor evaluator fallback;
   no reason to introduce a second ownership rule. Format8 and identities need not
   change: no durable representation changes, prior valid installations had no owned
   state, and changed Type installation already refuses before maintenance. If a
   real representation change is found, coordinate explicitly instead of bumping by
   habit or silently accepting older formats.

## Visibility and scope boundary

The serving owner remains fixed all-visible and empty aggregate. Integration must
prove correct complete inline owned state in actual Read/receipt output. Add bounded
batch/evaluator tests using its existing policy interface (or the existing test-only
spy pattern on evaluator options under a disposable owner) for hidden closure and
source-write refusal. Do not add mutable installed grants, project retained receipts
through an ad hoc filter, or use these unit tests to claim full authority visibility.

The closure helper handles transitive ownership and cycles; reuse it for set
candidates. A readable-but-nonwritable selected Link, or an otherwise writable Link
whose source cannot be written, is a failure after selection, not filtered out.
Hidden endpoints/owned closure must not contribute to visible cardinality or expose
IDs in diagnostics. Global authoritative constraints/deletion safety remain evaluator
checks against full state; no authorization-based filtering of incident index.

## File scope and ownership

Expected edits: transaction-batch.ts; transaction-store.ts (owned installation guard
only); new transaction-owned.test.ts; narrow installed-contract test update moving the
owned case OUT of the shared negative/no-file-created test into its own positive
fixture with a distinct path and explicit close; transaction-implementation-progress.md.
Actual emitted HTTP evidence may live in the new test via existing runtime fixture.
No transaction-runtime/service/receipt-page/Read engine/public export changes expected.
The surviving missing-compiled-schema guard error must drop its stale "and unowned"
wording. Preserve every other refusal and the negative fixture's no-file invariant.
No shared Resource evaluator or compiler edits: sourceRecipe remains private, existing
isResourceVisible is imported. A small source-private fact helper only if it materially
reduces duplication; any new module needs smoke inventory/denied deep-import evidence.
Root coordinates after set repair. RU retention has no source overlap.

## Executable done predicates

1. Creation, property update, metadata-only update, combined update and deletion of
   owned Link produce the correct two-Resource transition. Ordinary target stays
   byte-identical; self-Link versions its source once. Inline and first-class Link
   records agree member-for-member, explicit/wildcard bucket rules unchanged.
2. Event payloads pass normative bundle and semantic assertions: Link subject on
   lifecycle/incidence; source subject on final owned updated; group source remains
   Scope events/; exact ordinals; source-before-target self-Link facts; nested update
   delta excludes full properties and contains at least one patch field; the source
   updated shape excludes direct patch fields alongside ownedLink; deletion excludes
   payload; attribution agrees.
3. Two owned Links on one source in one batch, and direct source update interleaved
   with Link updates, form consecutive source revision chains per operation. Receipt
   carries intermediate sourceRevision; group carries final source only. Reverse an
   earlier property change and verify intermediate Events remain, even if final data
   equals initial data. No revision or provenance collapsed/reused.
4. No-op Link update retains both Resource revisions and provenance and reports the
   unchanged sourceRevision. All-noop batch yields no group/effectPosition. Mixed
   no-op/effectful transitions emit only the effectful facts; each receipt remains.
5. Create source/target with @labels, create/update/delete owned Link, then delete
   now-empty source: final group tombstones/upserts agree; initial/intermediate facts
   persist. Bead deletion while any incident Link exists still fails atomically.
6. Stale Link expectedRevision after another owned mutation fails against staged Link
   revision; stale Bead expectedRevision after a Link-induced source version fails
   and rolls back the whole batch. A source change alone does not invent a failure
   of a current Link guard. Separate admitted keys execute in one owner order.
7. Private set updates/deletes several owned Links in canonical ID order, emits matched
   selected count only, versions their common source once per effective member, and
   retains per-member sourceRevision. No-op members have no source fact. Later set or
   explicit operation sees final staged owned state. No HTTP set qualification.
8. Wildcard/explicit ownership maximum, inherited schema failure, source write denial,
   stale guard or final-operation failure rolls back Link/source versions, identities,
   groups and page plans, while retaining original failed receipt. Equal retry after
   changed graph state replays it; crash/reopen preserves disposition/atomic effects.
9. Lower evaluator-only Event cap fails exactly when the additional source fact crosses
   the bound, yielding transaction-wide inducedEvents limit failure with no partial
   graph or group. Matched cap/cardinality predecessor tests remain green. No claim
   of currently absent mutatedResources or byte-materialization budgeting.
10. Hidden-target transitive/cyclic ownership closure is tested at batch/set seam;
    visible matched count excludes hidden closure, read/write distinction enforced,
    no hidden-ID leak. All-visible real owner Type/Resource/receipt Reads show complete
    owned state; no mutable policy/retained withholding readiness claim.
11. Low page maximum splits owned results while retaining original sourceRevision and
    original source/Link postimages; same-key singleton/one-op batch replay and restart
    preserve facts without minting another source version or new Event.
12. Pinned build before process tests; all TX and shared Resource evaluator/Read owned
    regressions, new focused tests, types/boundaries/format/lint/both smokes/unchanged
    evidence. Commit/freeze exact candidate for independent review, no outward writes.

## Gaps/questions, without new policy

No unavoidable product decision identified for these owned transitions: exact source
Event shape/order, versioning, provenance, final group normalization and visibility
closure are already explicit. The capture-before-write seam is an implementation
choice; council should verify it retains each operation's actual prior source state.

Two limits remain independent and must stay visible: aggregate receipt/group byte
materialization can still be large despite existing per-Resource and count caps;
mutatedResources accounting is not implemented. Do not silently turn a count into a
byte cap. Root's separate bounded storage/receipt plan precedes public profile claims.

Mutable authorization projection/erasure, Snapshot/changefeed serving, schema hosting,
HTTP set refusal policy and incomplete minimum controls remain out of this slice.
They are not resolved by stored owned Events or by tests of the policy helper.


Root review correction: cbecfea's parent3e predates integrated a94 visibility work;
root8d/current637 already fixed set closure. Plan updated to preserve/regress it.
Capture-before-put was accepted as a reasonable local seam; direct deleted-subject
identity tracking still needs explicit checks (existing direct before snapshot, not
an absent post-deletion lookup). External council remains pending; no implementation.


## External council dispositions (review of root637, before required064)

1. Cardinality/cap HIGH is stale relative to explicitly required064. Root directed
   preserving064's deliberate cardinality-before-cap correction and its discriminating
   guarded409-versus-transactionwide413 regression, not reverting to a94's old ordering.
   Exact predecessor064db8a99629122f130a91ee8b5189432ffe97e9 awaits external review and
   root integration. No owned implementation before that gate.
2. Canonical/local Map-key warning accepted explicitly in step3. Local changed/deleted
   IDs are lookup keys only; absolute canonical Resource IDs key every state change.
3. Visibility fix already integrated at a94/root8d: reuse/regression, no new fix claim.
4. Owned positive installation gets a separate fixture path and close; negative
   provisioning test preserves all refusal/no-file checks, and guard wording updates.
5. Compiled owned fixture includes complete Link/endpoint descriptor dependencies.
6. Nested delta built field-by-field with nonempty patch invariant and source-shape
   exclusivity asserted; no future whole-value update semantics inferred.

Council confirms same-transaction putResource preimage timing and direct deleted Link
before snapshot are sufficient; no sourceRecipe API needed. Fact ordering/provenance,
no-op receipt handling, Format8 compatibility and unchanged receipt/page/service paths
are confirmed. No new product decision identified. This remains plan-only.
