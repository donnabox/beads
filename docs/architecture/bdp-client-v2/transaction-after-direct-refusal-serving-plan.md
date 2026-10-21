# Next settled serving slice: context on sequence carrier refusals

## Recommendation and status

Repair **missing Transactional Scope response context on the two existing authenticated sequence refusals: operation-count413 and pre-admission bodyless406**. This is a concrete wire defect under the same settled general response-field law as the direct-refusal work. It does not depend on incident-Link metadata membership, retained minimum interpretation, or a new admission rule. It is not an evidence-only addition and does not make discovery advertisement ready.

Plan only under the sole fork `docs/architecture/bdp-client-v2/design.md`. No authoring permission or acceptance is implied. Root must assign the reviewed direct-refusal successor/main basis after that council closes, since both slices touch transaction-http.ts. Preserve its selected409/transaction.operations413/fresh+retained406 changes and evidence. Do not quietly append sequence changes to frozen6ce9.

## Exact basis and current implementation

Read-only clean `/Users/dbox/repos/wt/vickie/vickie-start-now-resume` HEAD `8c78cba5d2be5d5612c6a9324ee36ed690fdeb19`, tree `f2bc6fd3750bb7b453984f9927f75e64554c777a`. Root reports verified main `f76bf119f531c6b44d1d3bc43531a4a8da27c4d0` has this identical tree; this audit used the local frozen tree and did not independently fetch main. Separately reviewed candidate6ce9 changes direct submit refusals, **not** the sequence branch below.

Read the sole plan's autonomous T2/T3 priorities, current consistency-rulings audit and discovery-serving v2 plan, plus the previous next-unblocked-serving audit. Their older status inventories must not be treated as current code: current canonical/Type/incident/alias and compatible-missing minimums, fresh Snapshot minimums, internal discovery minimum observation, and the now-ruled Scope start-now reconnect path are implemented at this basis. In particular, the old audit's start-now pending question is resolved, implemented and merged according to root; do not ask it again.

Verified remaining sequence paths at this source:

- `packages/server/src/transaction-http.ts:303–315` rejects a carrier-level Idempotency-Key, parses the shared carrier, then authenticates/authorizes.
- `:316–317` returns direct limit-exceeded with limit=sequence.operations for more than32 members, with **no context observation or context() wrapper**.
- `:318–324` prepares the existing supported sequence, preserving its explicit development-composition gap.
- `:326` returns bodyless406 for unacceptable JSON response media, again **without context**.
- `:327–328` executes only the accepted path with service.sequenceObserved and attaches its actual context. No refused request should be sent through that executing method merely to obtain headers.
- `transaction-service.ts:528–543` supplies successful sequence context from owner.withRead after ordered execution. `:130–144` provides the existing finite-work registration/close mechanism. `transaction-store.ts:883–958` already supplies a synchronous owner read entry with context and explicit entry-refused result; no new store method is needed.
- Existing `transaction-http.test.ts:421–443` verifies sequence syntax/key/negotiation rejection and a reused member key, but does not assert refusal context. Retain those assertions and add the missing context/state distinctions.

This is a bounded newly identified omission, not a claim to have exhaustively inventoried every required error response. The previous explicit-gap audit did not cover every direct refusal path.

## Settled normative contract

All anchors are `docs/specs/bdp.md` in the8c tree:

| Law | Required behavior and boundary |
|---|---|
|5766–5774, Transactional HTTP consistency|Scope-bounded responses carry epoch, authorization view and Scope position. It is not limited to receipt success or direct singleton/batch requests. The selected authenticated known-Scope refusals therefore need the tuple.|
|3157–3168, shared sequence response envelope|Carrier failures including excessive sequence member count are direct problems and execute nothing. Unexpected internal faults retain the existing bodyless500 rule.|
|3579–3592, mutation response negotiation|Unacceptable successful response media yields bodyless406 before durable admission/key binding/state change. Ordinary authentication/authorization/admission refusal ordering remains. A bodyless406 still carries required response headers.|
|4046–4056, Transactional sequences|The shared sequence envelope/carrier remains; member dispositions become one-operation transactions. It does not become one batch receipt or gain a carrier-level key.|
|4588–4599, Transactional mutation surface|Sequence has its own direct413 and shared406; no sequence-level202 receipt. Keep these exact existing statuses and no new code.|
|6973, transactional.http.accept-refusal|Corroborates negotiation before admission while ordinary refusal precedence remains. This slice does not qualify the entire catalog row.|

No new minimum barrier or collision precedence is introduced. Work only with an authenticated authorized principal, valid supported carrier and no minimum fields. Preserve existing size/media/syntax/auth and unsupported-control/composition behavior. The fixed all-visible reference view is the exercised policy; do not claim hidden-policy nondisclosure coverage.

## Minimal implementation boundary

Use a **narrow source-private service observation for these sequence refusals**, under existing run() lifetime and exactly one owner.withRead. It may return an immutable `{ context }` or a typed selected refusal plus context; choose the smallest existing-style representation. The fixed refusal outcome derives from immutable already-validated carrier length/media inputs, not mutable graph/key state. Capture the context at the serving observation; never synthesize it from a prior successful receipt or cached discovery. Do not call discoveryObserved (which unnecessarily assembles/validates discovery), executeSequence, admitSequence, a mutation submission, or an arbitrary callback framework.

The HTTP branch retains its current order. At the existing over-count branch, obtain the narrow observation, then wrap the existing sequence.operations Problem. At the existing negotiation branch, obtain it and wrap empty406. Unsupported controls/composition and early syntax/auth refusals remain where they are. No generic “all errors” postprocessing, blanket413 wrapper, new owner read for unobserved service.sequence, or changed admission law. The current order keeps over-count413 ahead of unacceptable Accept.

Return the same frozen three-field tuple from the single synchronous entry. This does **not** require a new SQLite write transaction: these decisions do not inspect keys or mutate state, and existing withRead is the owner's synchronous consistency boundary. Do not claim it is the atomic key-admission transaction used by the direct-refusal repair. Keep the owner entry synchronous and released before any await/transport delivery. Closed/reentrant/failed owner and validation faults cannot produce invented context: preserve existing service/HTTP internal-fault behavior, with cause retained internally and sanitized500 on wire. No wait, timer, arbitrary signal policy or command minimum support is added by this nonwaiting observation.

Proposed maximum six owned files:

1. `packages/server/src/transaction-service.ts` — narrow private observation/lifetime method.
2. `packages/server/src/transaction-service.test.ts` — real owner context, entry/immutability/lifecycle/no-admission tests.
3. `packages/server/src/transaction-http.ts` — only the two refusal wrappers.
4. `packages/server/src/transaction-http.test.ts` — context/precedence/no-effects and existing regression preservation.
5. New `apps/bdptest/src/transactional-sequence-refusal-installed.test.ts` — bounded actual installed proof using existing launcher/raw transport/binding conventions.
6. `docs/design/transaction-implementation-progress.md` — exact scope and remaining gaps, if needed.

No store, evaluator, sequence execution, parser, protocol export, schema, catalog, fixture contract, runtime configuration or public-admission changes. A needed additional file or serving semantic change returns to root before implementation.

## Source tests and discriminators

- Seed two genuine separate commits. Both valid33-member rejection and valid one-member406 return actual currentpos-2 and seed-derived opaque epoch/view; header presence is a boolean/nonempty check before equality, avoiding the recently observed numeric-matcher diagnostic loss. Repeat after an unrelated commit to catch cached context; earlier returned tuple stays frozen.
- Service test records exactly one successful owner entry, no admitSequence/nextSequence/execute calls and unchanged owner inspection. Tests for closed/reentrant entry preserve failure rather than fabricate headers. Existing run registration/close behavior remains; no artificial asynchronous owner callback or long wait is needed.
- HTTP:33 syntactically valid keyed members plus unacceptable Accept still return413/sequence.operations. A supported one-member sequence with unacceptable Accept returns bodyless406, no Content-Type. Both retain actual context, no-store and Vary handling. Verify canonical Problem shape on413. Wrong media, malformed body, carrier key, unauthenticated/forbidden and unsupported minimum controls keep their existing paths without calling the new observation; do not infer new disclosure precedence for them.
- Independently inspect unchanged complete graph/Events/receipts after refusal. Reuse each refused request's selected member key with a **different valid request identity** and observe success, proving no retained key capture rather than merely equal replay. Source inspection can check all33 rejected keys; installed proof need not duplicate33 mutations.
- Two isolated source controls are sufficient: omit the413/406 context wrapper (run the named cases separately so each is reached), and substitute a stalepos-1 tuple after the second commit. Each must fail the named presence/value assertion, not setup or an arbitrary matcher error. An entry/no-admission spy catches obtaining context via execution. These are source sensitivities, not installed target observations.

## Bounded installed proof proposal

One new scenario, **11 exchanges/one child**, no reopen. Fixed Task schema and all-visible installed policy; use actual shipped32 sequence limit. Four static routes (singleton-create, sequence, Beads collection, Scope Events) plus one validated observed receipt URL. No hidden polling or Type fetch. Schema-check actual receipts, collections, Events and direct Problem; independently authored IDs/properties/metadata plus real opaque revision/transaction equality. Keep old direct-refusal14/1 scenario separate and unchanged.

|#|Exchange and oracle|
|---|---|
|1–2|Two successful singleton creates a,b under distinct keys:pos-1 thenpos-2; retain complete first receipt and current literal Resources.|
|3|33 valid explicitly identified sequence creates, distinct member keys, unacceptable Accept, no carrier key:413/sequence.operations and all current fields atpos-2.|
|4|One valid supported create sequence member c under fresh key, unacceptable Accept:bodyless406/currentpos-2.|
|5|GET genuine first receipt root:body remains exactly original; serving contextpos-2.|
|6–7|Complete current Beads and finite Events pages:only a,b; next:null; literal facts and real receipt relationships, currentpos-2.|
|8|Singleton create c with the refused406 member key but a different valid properties value:completedpos-3, no conflict.|
|9|Singleton create d with the first refused413 member key and different valid ID/body:completedpos-4, no conflict.|
|10–11|Complete Beads and finite Events pages:exactly a,b,c,d and independently expected corrected values, next:null/currentpos-4.|

Use unchanged5s exchange/90s aggregate/120s outer/10s child-close conventions and current request/header/response limits; do not raise them. Before execution measure actual33-member bytes and route/body hashes, source footprint, response estimates and exact tests. Root accepts that measured ledger. Suggested author validation: new positive11/1, unchanged direct-refusal regression14/1 after production changes, and one separately packed predecessor negative stopping at request3=3/1: **28 expected exchanges/three children**, ceiling36/three (two full11-request new-scenario runs plus the14-request regression). No replacement run after an unexpected failure without disposition. Negative source must be the assigned accepted parent lacking this sequence repair, not assumed to be8c after integration.

Record exact clean source/tree/parent, actual installed tar/payload/configuration/policy/auth/fixture identities and expected route authority, actual attempted/completed counts and cleanup. Retained artifacts must accurately describe fingerprints versus full wire bytes. The negative must fail specifically at a missing sequence-refusal context assertion after its two seeds, with genuine413 and clean shutdown. Preserve failures; do not repeat the direct-refusal collector's label mistake or treat exit1 alone as proof.

## Done-when and blocked alternatives

Root/native+Claude council accepts the final plan and assigns the reviewed parent; the six-file implementation passes focused service/HTTP/sequence/receiving and direct-refusal regressions plus type/build/static/boundary/Read checks. Actual frozen installed observations and controls meet measured budgets and have no child leaks. Independent code review closes concrete findings. This fixes two actual response surfaces; catalog whole flags, discovery, public profile selection and qualification remain unchanged.

Do not substitute these unresolved alternatives:

- Incident metadata-only Link Event membership still needs its existing operator ruling.
- Retained Snapshot/collection minimum anchor comparison, malformed/partial/unknown minimum controls, mutation minimum applicability, incompatible missing/hidden precedence and set-Selector phase remain the precisely documented questions; this plan chooses none.
- Comparison-loop timeout503 is a real adjacent response-context concern, but its refusal can arise outside the final owner entry and can include epoch/lifecycle changes. It needs a separate bounded observation/disclosure analysis; do not reuse a stale last-attempt tuple or add a blanket fallback here.
- HTTP discovery activation still asserts the complete cumulative profile; a new internal observation or two repaired errors cannot authorize it.

The old serving audit's “no independent repair identified” conclusion was scoped to its explicit minimum/Event/discovery inventory. These two demonstrably missing sequence context fields are an independent settled-contract serving increment outside that inventory. No extra product decision is identified for this narrow case. This plan neither asserts all other requirements are blocked nor invents work solely to fill time.

Audit used local Git/source/report reads only; no tests, builds, services, network, source edits or outward actions. Only this supporting plan was written. It is not another plan of record or authoring authorization.
