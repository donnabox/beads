# Next required Transactional slice after minimum reads

Read-only plan against frozen `d3252087a31529cf97b5bfcfe2c402f02c694b91`.
The sole fork design remains the plan of record. No source edits or outward
writes occurred for this inventory. Implementation awaits root scope review.

## Recommendation

Implement durable receipt pagination, end to end through the existing private
HTTP receiver, before set mutations. This is independently exercisable using
existing explicit batches: configure a result page bound of 2, commit 5 existing
Resource operations, and recover all 5 results through one receipt and its pages.
It removes the artificial `page.maximumItems >= 32` service restriction while
preserving the existing 32-operation receiving bound. It does not require a
Selector, new mutation semantics, or a new public profile advertisement.

The completed unit is durable result pagination for the current fixed all-visible,
schema-free/unowned supported configuration. The normative catalog's large-set
pagination witness remains open until actual set expansion exists. Variable grant
projection and erasure are separate prerequisites to complete profile support;
this plan does not call those behaviors satisfied by an all-visible fixture.

## Required-contract gap inventory

The normative catalog has 133 scenarios in
`packages/conformance/catalog/transactional-v1.json`. These are obligations, not
133 executable passes. Current private tests implement substantial atomic write,
identity, recovery, receipt and fixed-view Read behavior. The remaining groups are:

| Group | Existing primitive | Remaining required implementation / evidence |
| --- | --- | --- |
| Receipt pagination | Durable complete result array, terminal expiry, same-entry observation, root HTTP reads | Durable page identity and boundaries, page reader/HTTP, current context and complete page refusal order, restart witness; catalog receipt.pagination/page-expiry/page-restart |
| Sets | Normative request schemas; Read Selector evaluator; singleton Resource evaluator; atomic owner | Set preparation and semantic identity, staged projected candidate enumeration, cardinality and complete-set mutation, canonical ordered result/Event expansion, set singleton routes and finite expansion limits; catalog set.* and event.set-expansion-order |
| Replication Snapshot | Stable in-memory Read query pagination and same-entry rows/context | First-class manifest, one jointly captured Bead/Link snapshot, durable/retained handles, independent streams, expiry, exact checkpoint rendezvous with replay; catalog snapshot.* |
| Changefeed | Durable ordered group JSON including final postimages and Events | Bound epoch/view checkpoints, replay floor/retention, finite pages, SSE and reconnect, start intent, minimum/conditional/HEAD semantics, projection advance and failure lifecycle; catalog changefeed.* |
| Event Sources | Committed graph-generic facts with correct Link subjects/endpoints and self-Link facts | Source-specific finite/SSE readers, source cursor binding, replay/history expiry and withheld-cursor behavior; catalog event.history-expired/cursor-after-withheld and HTTP rows |
| Current authorization / erasure | Fixed all-visible view and principal-bound receipt access | Shared variable-view projection across receipt/sequence/read/replication paths, owned closure, version-erasure ledger and invalidation where applicable; catalog receipt.reauthorization*, sequence.withheld*, erasure.*. This is required Transactional behavior where applicable, not optional History. No design expansion in this slice. |
| Remaining sequence references | Durable canonical @name reservations and alias witnesses on direct carriers | Resource-alias member-turn resolution/admission representation; retain separately documented wording question, do not infer a ruling |
| Full Read consistency surface | d325 visible point and receipt-root complete canonical tuple subset | Absent/hidden point ordering, retained query continuation minimum meaning, unsupported tuple/unknown-token policy and remaining routes. Those unsettled combinations are excluded here. |
| Installed contract integration | Existing shared Resource evaluator; separate Update installed-contract work | Integrate reviewed installed configuration into TX without copying policy/compiler authority; root coordinates, no Type/provenance changes here |
| Cold discovery / qualification | Canonical target fixture, source-private operations and catalog schemas | Schema-valid truthful Transactional discovery, complete Directory, installed runtime configuration and exact executable conformance matrix once required routes exist. No fabricated descriptor now. |
| History replacement / optional claims | Ordinary same-epoch restart recovery exists | Operational restore/new-epoch fencing and evidence for cross-epoch tokens; optional History features remain excluded. |

The current bounded unowned six-Resource owner has at most a finite number of
facts per operation and a 32-operation cap. Set expansion removes that implicit
bound, so a real enforced induced-Event budget is a prerequisite to enabling sets.
Do not accidentally characterize an advertised transaction duration as mandatory
when none is advertised; its corresponding behavior must be implemented if chosen.

## Why paging before sets

Settled paging clauses:

- `docs/specs/bdp.md:3831–3833`: page limits count every receipt result entry.
- `4952–4979`: same durable receipt identity/disposition; inline or bounded prefix;
  complete results, no truncation or second result abstraction.
- `5067–5073`: complete result and pages created atomically; no recomputation.
- `5115–5142`: detail expiry, full allocation tombstone, failed receipt forgetting.
- `5167–5187`: prefix, absolute next URL, immutable unsplit entries, restart-stable
  page identities/boundaries, current projection, service failure distinct from expiry.
- `5189–5239`: validator omission, shared projection, immutable execution facts
  versus the serving observation.
- `4590–4624`: exact status/method table and explicit receipt/page ordering:
  authentication, principal/epoch nondisclosure, minimum context/wait, page expiry,
  representation. This is specific receipt-page authority; it does not resolve the
  separate retained query-page minimum question.

By contrast, sets touch `transaction-batch.ts`'s PreparedReadUpdateCarrier
representation and singleton-only normalization, its per-operation expansion loop,
and the store's Resource transaction interface. Read selector reuse is real
(`read-selector.ts:selectReadResources`), but selection must see staged canonical
postimages at the operation turn, not a previous public Read page. Required rules
are `bdp.md:4240–4280`, `4904–4918`, `5158–5172`, `5578–5594`. Set input identity
must retain selector text and normalized defaults/lossless numeric equality. A
narrow set implementation can follow paging with no protocol changes, but it is
not the smaller next unit.

## Concrete representation and seam

Retain the full immutable receipt facts in the existing receipt row. Do not replace
that internal array with the first wire page: `expireRow` must still extract every
creation allocation (including tail pages), `creatorFact` needs exact creation
facts, and the sequence projector expects the complete singleton fact. Current
call sites: store `expireRow` around328; `creatorFact` around610; execute terminal
receipt around884; observed admission around1004; sequence `project` around55.

Add a narrow durable page-plan relation owned by the same store. Recommended
shape: opaque page token, parent receipt ID and principal/key, start index,
exclusive end index, and next token. Index parent principal/key plus start index
for the first/next boundary lookup; pass the already-authoritative principal/key
through observed execution/admission rather than adding an ID scan to delivery. The relation stores boundaries/identity, not duplicate Resource
postimages. Create all required page records in the same SQLite terminal commit
as full receipt facts/effects/group. Their tokens become durable only if that
commit succeeds. Root projection is a pure prefix over retained facts plus the
first page URL; each page projects its recorded slice through the same fixed-view
receipt projection. Root prefix length comes from the first recorded page boundary,
not a new current default. Receipts with no page records retain their full array. No in-memory cursor cache or second store.

Use an authority-chosen positive page-entry bound, fixed at store construction and
checked against the service's actual `page.maximumItems`. Record immutable boundaries
per receipt, so same-epoch reopen never derives them from a new current default.
Recommend persisting the configured bound in the private installation fingerprint
and refusing mismatched startup configuration before recovery/maintenance. The
private format changes; root must approve the usual byte-identical refusal of old
formats, with no migration/reseed. This is a private representation choice, not a
new wire semantic or public configuration field.

Opaque page URLs can use the existing canonical receipt namespace, e.g.
`receipts/{receipt-token}/pages/{page-token}`. The spec does not prescribe the page
suffix. The exact generated URL is a navigation link, not permission: lookup always
requires current authenticated principal plus matching parent and epoch. Unknown
page token/parent/principal/prior epoch returns uniform404. Do not infer receipt
existence or expiry from parsing a guessed URL.

Retain compact page-plan identity/boundaries alongside completed receipt tombstones
for the epoch, without retaining erased/expired payloads. This lets an actually
issued page URL return410 after detail expiry. A retracted/pending/failed receipt
never acquires successful-result pages. Forgetting failed receipts cannot orphan
pages. All raw SQL failure paths retain existing atomic rollback/fencing/cause
semantics; this slice must not bypass the owner helpers.

Add one typed `readReceiptPage` observation entry. Lookup/authenticated principal
and epoch nondisclosure, current context, accepted minimum tuple/wait decision and
page expiry/body selection occur through that entry. Wait outside the entry and
retry actual lookup; never hold a page body across a wait or attach later context.
A retained expired completed root still establishes the principal-bound identity;
therefore a foreign minimum must precede page expiry, exactly as 4618 requires.
Root expired receipts continue to return200 with full allocated identities.

Reuse `observeReceipt`'s current integrity checks and projection for initial
mutation, duplicate and root reads; add prefix navigation there while preserving
internal full facts. Do not add a second clock observation between terminal commit
and the first delivery (the reviewed T2a failure). Pages use the same projection,
current context and finite wait machinery. Singletons still fit at bound>=1, so
existing cross-carrier and sequence projections remain unchanged. No sequence
paging and no set-operation changes in this slice.

## Proposed exact file ownership

- `transaction-store.ts`, `transaction-store.test.ts`: page relation, private format
  and installation check, atomic terminal page allocation, internal facts versus
  wire prefix projection, same-entry page lookup and retention behavior.
- New source-private `transaction-receipt-pages.ts` and tests only if useful for
  deep, pure page partition/URL validation; no generic pagination framework. One
  module may own the bounded representation policy, while SQLite ownership stays
  in transaction-store.
- `transaction-service.ts`, `transaction-service.test.ts`: remove >=32 restriction;
  validate supported bound compatibility; typed page results and tracked async page
  minimum waits, shared shutdown behavior. Preserve sync receipt API.
- `transaction-http.ts`, `transaction-http.test.ts`: distinguish root versus known
  page targets, uniform method gate, route through actual service, existing Problem
  envelope/context and conditional/HEAD handling. No shared Read router edits.
- `transaction-runtime.ts`: source-private fixture page-bound seam only; same owner
  construction and canonical Scope/listener separation. No public flags/options.
- `transaction-runtime.test.ts` or existing test-support process fixture: same-epoch
  process loss/reopen preserves previously issued page URLs and boundaries.
- `transaction-sequence.test.ts` / `transaction-dependencies.test.ts`: regressions
  for singleton projection/creator binding after pagination and detail expiry; no
  sequence mechanism edit unless a concrete coupling is found and coordinated.
- Private emitted helper allowlist/root-negative/deep-import smoke assertions if
  a helper module is introduced; progress note with bounded claims.

No shared RU evaluator/config/HTTP-routing, selector compiler, normative spec/schema,
public exports, catalog/evidence relabeling, Type/provenance or owned-Link changes.
Root coordinates transaction-runtime overlap with any TX installed-contract work.
Base should be the reviewed successor of d325, including any review fixes, not an
unreviewed work-in-progress. Do not cherry-pick away prerequisite recovery history.

## Done-when / executable acceptance

1. Actual HTTP with page bound 2: one explicit 5-operation batch produces two inline
   entries, then pages of2 and1, final next:null. Every entry exactly once in order;
   all root/page bodies validate the existing mutationReceipt/mutationReceiptPage
   schema. One effect group/terminal receipt, not one commit per page.
2. Exact-fit/all-inline, bound 1 singleton and alias, no-op and failed transactions
   preserve shapes and no unnecessary page links. An entry is never split. Root
   and page rows encode original intermediate postimages even if later operations
   in the same batch or later transactions change/delete those Resources.
3. Initial POST, equal retry, independent root GET and HEAD have the same receipt
   identity, disposition, original entry boundaries/next URLs. A post-commit clock
   jump does not insert a second expiry lookup before first delivery. Replays after
   normal later expiry return the compact expired root and never execute again.
4. Late-operation failure, injected terminal/page persistence fault and failed
   COMMIT produce no orphan page, partial group or partially admitted effect.
   Existing uncertain-commit/sticky fence/reopen fault matrix remains green.
5. Issued page URLs survive ordinary process restart and actual abrupt post-commit
   process exit in the same epoch. No fresh URL, changed boundaries or410 caused
   solely by reopen. A changed canonical listener address does not rewrite links.
6. Expire detail: root 200 expired retains creation allocations from all pages;
   authorized known page 410 cursor-expired; unknown/foreign/prior-epoch page 404;
   unauthenticated 401+challenge. No result data retained in compact page rows.
7. Page refusal order: method shape first; auth then nondisclosure; then supported
   complete minimum tuple (foreign view 409, foreign epoch 410, real timeout 503);
   then actual detail expiry 410; then representation/conditional/HEAD. Hidden
   receipt pages never disclose context or expiry. Use actual owner states and
   canonical transactionalProblem validation, not a mocked status table.
8. Page 200/304/412 and minimum refusals carry the actual final entry context while
   retained execution facts remain original; omit ETag/Last-Modified, retain
   private,no-store. GET and HEAD parity, no unnecessary body on HEAD, Retry-After
   on real timeout; no fake context or minimum claims on unsupported tuple forms.
9. Fixed-view projection helper is shared across root/page delivery; no bypass of
   present principal/epoch checks. Variable grant/erasure scenarios remain explicit
   future gaps rather than fake all-visible passes.
10. Close/disconnect behavior stated exactly: tracked finite waits drain; source
    facet AbortSignal support does not imply HTTP disconnect propagation. No new
    cancellation subsystem in this slice.
11. Startup with prior private format or incompatible bound refuses before mutation,
    byte-identical file; compatible reopen preserves facts. Public imports/config
    unchanged, helper inaccessible through package deep imports.
12. Pinned build before emitted tests; focused owner/service/HTTP/page + all TX and
    Read composition regressions, typecheck/boundaries/format/lint, both smokes and
    unchanged Read evidence. Exact clean commit/report for independent review.

## Decisions and remaining unknowns

No product ruling is needed for the scoped paging behavior: page receipt ordering
and current observation are explicit. Root must approve the private storage/page
bound representation and file ownership before implementation. That is code design,
not a proposed protocol amendment.

Avoid reusing Read query cursors: their memory lifetime/restart semantics do not
satisfy receipt pages. Avoid treating the page helper as full result authorization:
variable grants and erasure remain required follow-up work. No catalog-wide claim
until actual set-result expansion exercises pagination and all applicable views.
Unknown minimum tuple handling stays exactly the private gap from d325; no new
normative response invented for partial tuples. Absent-point and query-continuation
questions do not affect the explicit receipt-page ordering contract.

After this slice, the next natural implementation is atomic set operations using
existing Selector/Resource evaluators, then Snapshot/checkpoint + replay as one
coherent bounded design. Root can schedule installed-contract integration in
parallel only when transaction-store ownership is coordinated.
