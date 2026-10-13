# Transactional Event serving — revised finite-first plan

Plan only against exact owned successor **01e2d5470941a49240fab7820a6dea1b6deb291e**.
Original71549f2 proposal preserved as `transaction-event-serving-plan-pre-council.md`.
External review: `transaction-event-serving-claude-plan-review.txt`. This revision
independently dispositions its claims; no production code or public claim changes.
The owned successor has native review; root owns its remaining integration gates.

## First slice and explicit boundary

Implement a **finite JSON Event reader** over the existing same-owner committed
groups, exposed through the private Transactional service/HTTP fixture for Scope
`events/` and canonical Bead/Link `?view=events`. Start with the actual fixed
all-visible installation (now including owned Links), all retained groups and no
pruning. This is useful cold-client delivery, not a complete Transactional profile.
Keep SSE transport as the next separately reviewed slice. Do not advertise SSE HEAD
before the corresponding GET representation is supported. No optional History,
Snapshot/changefeed serving, erasure administration, compiler/Type lifecycle,
public admission or RU owner changes.

Reuse actual owned facts as immutable input; do not reimplement their writer or
Event cap. `data.ownedLink` passes through unchanged. Scope Event rendering must be
the same rendering later used inside served change groups; raw internal group IDs
are not a separate public history contract. Current `inspect()` is a diagnostic
surface, not a serving API.

## Normative source and cursor correction

- Spec `#events-and-event-sources`5361–5373: each source has its own identity;
  deterministic projection from committed groups; stable ordinal; ID/cursor stable
  for checkpoint, ordinal and projection.
- `#event-replay-and-live-observation`6184–6195: after is **the ID of an Event in that
  Event Source**; omitted starts earliest retained; exclusive continuation; next null
  at current end; IDs/order have meaning only within their source.
- Owned fact order5480–5487: ordinals are never renumbered by projection. That says
  nothing about requiring identical opaque ID bytes across different sources.
- Version erasure6450–6457 separately preserves a cursor whose formerly observable
  Event was withheld. Never-member unrelated facts and formerly-member withheld
  facts are not interchangeable. This slice performs no withholding, but must not
  bake in an assumption that every valid future cursor still has a visible body.

**Reject F1's mandatory fixture encoding and F4's global-ordinal admission.** The
fixture `ckpt-43_0` spelling illustrates one permitted encoding; the protocol test
checks source equality and ordinal increase, not universal byte encoding. Skipping
whole unrelated groups after a valid issued cursor proves filtering, not validity
of an arbitrary cursor naming a fact absent from that source. A source-specific
encoding is permitted. The spec does not mandate detectable refusal of an identical
string previously obtained elsewhere when that string also names a local Event.

Recommended private codec choice: bounded readable position/ordinal plus a fixed
SHA-256 base64url digest over an unambiguous serialization of epoch, view, canonical
source and position/ordinal. Do not concatenate/base64 an unbounded URL. Full digest
is fixed43characters; current safe-integer pos-N plus ordinal stays far below256.
Recompute from requested source/context, then verify retained fact/source membership;
do not treat digest parsing as authorization. This choice makes cross-source tokens
unequal/refusable, which is an implementation consequence, not a newly normative
cross-source error rule. `genesis` is reserved and never an Event ID (6564–6576);
Event parsing must not reuse minimum-position parsing that accepts genesis.

No stored ID rewrite, sidecar table or format change is needed: render source-local
IDs deterministically from existing immutable facts. Preserve original ordinals,
transaction IDs and payload bytes/values. A later changefeed renderer must use this
same Scope rendering, not expose old internal pos-N-ordinal tokens inconsistently.
Stable restarts reproduce IDs. No current public Event cohort is relabeled.

## Same-group endpoint proof: no cross-group index

At01e2, transaction-batch.ts643–663 normalizes final changes;666–697 emits created,
updated and incidence facts. For a Link updated in a group:

1. If it survives, the group's final Link upsert includes immutable endpoints.
2. If deleted later in that group, its final change is a tombstone, but its unlinked
   facts carry both endpoints. For Bead projection at least one endpoint is local,
   so at least one such fact exists; external-only Links need no Bead projection.
3. If created, updated then deleted in the group, the created fact also carries both
   endpoints. This holds with owned source Events in the same group.

Independent emitted-owner probe passed all four owned/unowned × update-delete/
create-update-delete cases: final Link absent/tombstoned, update delta lacks endpoints,
exact endpoint witness remains in the same group. Script and results:
`event-plan-revision/same-group-endpoints.mjs/.log/.exit` (exit0), against built01e2.
No source mutation. Add this invariant as a permanent regression in implementation.

Build a temporary per-group subject→endpoint lookup from upserts, created facts and
incidences; check consistency, preserve stored pinned References, and discard after
projection. Cost is O(group changes + Events), not literally O(1) without indexing.
No cross-group scan/index, new log or writer mutation. Existing group byte-materialization
limits remain root's separate work; count bounds are not a byte-memory guarantee.

## Bounded implementation sequence

1. New source-private `transaction-events.ts`: strict source/query/cursor primitives,
   same-group projection, finite page assembly. Scope includes all visible retained
   facts; Link includes its own lifecycle; Bead includes own lifecycle, applicable
   endpoint incidence (self-Link two facts), and incident Link updates according to
   the metadata disposition below. Never copy incident Link created/deleted lifecycle
   into a Bead source. Retain owned source updated as its own subject's lifecycle.
2. `transaction-store.ts`: narrow guarded finite observation method over durable groups
   and retained identities, returning detached frozen page plus current readContext.
   No `inspect().groups` all-log copy; fetch bounded groups incrementally up to a
   selected prefix under one entry. No SQLite entry survives await. Current subject
   absence does not discard its retained source. Receipt expiry already leaves groups
   alone; add a cheap regression, not a new sweep mechanism.
3. `transaction-service.ts`: finite method wraps ordinary owner observation; reuse
   checkTransactionMinimum/waitForTransactionMinimum's one original finite deadline.
   Authentication remains HTTP authority-owned; fixed all-visible projection is not
   principal-owned receipt lookup. No new live consumer registry in this finite slice.
4. `transaction-http.ts`: detect exact Scope events/ and canonical bead/link view=events
   **after authentication and before receipt/general minimum/read-request dispatch**.
   The lower-profile Read parser must continue rejecting view=events. Return
   jsonResponse+current context, then existing JSON negotiation/conditional/HEAD helper.
   Do not set ETag or Last-Modified. Current Origin refusal remains; no CORS claim.
   SSE-only Accept receives ordinary bodyless406 in this finite-only private slice.
5. Freeze with focused real HTTP and owner tests, no full-profile admission/evidence
   claim. No change to transaction-batch writer, shared public HttpResponse, protocol
   schema, fixture Type installation, compiled capture or private store format expected.

Bound pages by existing fixed page configuration; do not invent an Event `limit`
query parameter merely because collections have one. Finite Event pages may split a
committed group's Events: atomic visibility means only committed facts escape, not
one group per Event page. next names the last returned Event only when continuation
exists; valid after-last yields empty/current-end and can later observe appended facts.
Use scan/lookahead to establish filtered current end without fabricating a cursor.

## Questions that need disposition versus ordinary choices

**Actual prose inconsistency — incident metadata-only updates.**5553–5554 says
incident Link mutable *properties* changes;5377–5378 and5617–5620 define properties
or common metadata as one updated transition. Neither proves conclusively that the
narrow incident rule includes metadata-only transitions. Do not silently apply F13's
recommended broadening. Root should reconcile the incident sentence with the generic
transition rule (recommend properties or common metadata), or explicitly choose the
narrow projection. Scope/Link sources and own-source ownedLink Events are unambiguous;
projection core and ordinary property-update tests can proceed independently. Do not
claim complete Bead Event serving until this edge has an authoritative disposition.

**Unknown cursor code — real underspecification.** A well-formed token that was never
issued/names no source Event is not a demonstrated expired retained cursor. Neither
6184 nor the closed problem table uniquely assigns its failure. Malformed grammar,
genesis, wrong source digest, fabricated ordinal, former epoch and former view must
be separate inputs; do not label them all expired. Recommend root approve bounded
invalid-parameter for malformed/never-issued tokens, preserving normative foreign-view/
cursor-expired where a known context permits that diagnosis. This is a proposed
refusal convention, not existing mandatory mapping. No new error code. Known issued
continuations and omitted-start finite core are independent of that error decision.

**Minimum versus unknown source — precedence not uniquely stated.**1980ff and5804ff
require disclosure/cursor/minimum checks before304, but do not order a simultaneous
unknown/undisclosable source and incompatible/future minimum. Existing receipt code
returns uniform404 first; current Resource read explicitly records a visibility gap.
Recommend same authenticated source404-first convention for Event routes, with retained
deleted sources classified as available before minimum checks. This does not widen
visibility or return a representation below the minimum. Root can adopt this bounded
precedence; do not cite it as already universally mandated or propagate the Resource
read gap500 by accident. No question about subject deletion itself:5630ff already
permits retained Event Sources independently of current Resource liveness.

**SSE conditionals — no new product ruling required.**1980–2008 makes GET/HEAD
preconditions apply to existing representations, including absent validators;5804ff
and6536–6561 explicitly apply inherited rules to Scope/Resource Events and SSE HEAD.
No media-type exception exists. F12's fixture omission is a missing implementation/test,
not permission to skip preconditions or evidence that law is undecided. SSE must apply
ordinary304/412 before opening, after ordinary refusals/minimum checks.

**Negotiation ties — ordinary implementation choice.** Missing Accept/default or an
equal preference permitting both can choose JSON. `text/*` permits SSE, **not JSON**;
reject F11's suggested text/*→JSON rule. Respect specificity/q=0 and select among
supported representations. Pure media selection can precede execution; final406
must preserve ordinary refusal precedence. No normative latency bypass of minimum
failures is inferred merely to avoid spending a catch-up wait.

## Focused finite done predicates

- Exact Scope, Link and Bead projected pages for ordinary/owned create/update/delete,
  direct source update, self-Link source/target incidence, external and pinned endpoint;
  owned delta untouched, source/subject/ordinal distinction explicit. Assert two
  subjects for an owning Bead's incident Link update plus its own ownedLink update.
- Same-group endpoint invariant above at actual writer boundary; deleted historical
  source/restart/page stability; no current-Link-table dependence. Failed/rolled-back,
  no-op and alias-only mutations create no fabricated Event/position.
- Page within/across groups, gaps from unrelated facts/groups, empty source/current
  end, append after end, repeat/restart. Cursor from that source excludes exactly its
  Event. Never-member fact is not upgraded to a valid cursor by global ordinal; no
  erasure-withheld case is claimed without the separate erasure implementation.
- All response Events pass schema plus literal semantic oracles. Actual cold HTTP
  reader accumulates same independently expected facts. Owner observation provides
  one projected prefix/current headers; lower-profile view=events remains invalid.
- Current/ahead minimum success/timeout, wrong epoch/view, ordinary source404 and
  chosen unknown-token convention tested independently and together after disposition.
  Authentication/Origin/method checks and HEAD bodylessness preserved. Existing JSON
  conditional helper covers absent-validator *, specific and precedence cases after
  admission; finite responses omit validators. Receipt expiry leaves Event replay intact.
- Fenced/closed owner cannot serve; no read capability outlives entry. Service finite
  waits drain under existing budget. No mutation or pending-key admission by observation.

## Deferred SSE engineering, not part of finite freeze

New private discriminated stream response, route exit avoiding JSON-only handler
semantics, Node incremental writer honoring drain, complete frame serialization,
Last-Event-ID validation before headers, and replay→follow race test are real work.
Existing server.timeout10s plus timeout socket.destroy requires explicit per-stream
policy; do not change ordinary receiving/header limits globally. Idle survival,
keepalives if chosen, slow clients, socket close and closeAllConnections need tests.
Service must cancel registered streams before active-work drain; signal propagation
from response close is currently absent. Stream failures must not fabricate retained
mutation failures; owner failure still terminates all affected observations.

Reject F7's claim that timer polling necessarily interrupts a busy owner: same-thread
synchronous entry cannot be interrupted by another JS timer. Reentrant callback access
is an invariant problem, not automatically retryable contention. Schedule observation
only after owner exit; bounded cancellable polling or post-commit wakeup is a design
choice, not law. External SQLite access is not converted into an automatic retry loop.
No stream implementation may hold owner entry across await/backpressure.

## External finding dispositions

F1 reject mandatory literal IDs; F2 accept bounded digest clarification, reject claim
that a codec must concatenate all input; F3 correct overstatement; F4 reject global
never-member admission; F5/F6 accept SSE timeout/cancellation gaps, defer to SSE slice;
F7 reject inevitable timer contention/retry prescription, retain owner-fence tests;
F8 accept same-group simplification with four-case proof; F9 accept exact route branch;
F10 accept stream/JSON distinction, reject unsupported HEAD-SSE now as a free shortcut;
F11 accept selection/refusal separation, reject text/*→JSON; F12 resolve from existing
GET/HEAD laws; F13 actual metadata consistency question; F14 bounded precedence
proposal; F15 mapping proposal requires explicit disposition; F16 separate parser;
F17 no general token-validator refactor needed for finite slice, small private reuse
only if actual duplication justifies it; F18 add row distinctions; F19 regression-only;
F20 corrected spans/Version-erasure attribution.

Catalog references include transactional.event.{fact-order,graph-facts,self-link,
set-expansion-order,owned-link-delta,owned-link-update-delta,owned-link-deleted-identity}
and transactional.http.{token-profile,accept-refusal,method-405,problem-contexts,
head-parity,conditional-reads,finite-validator-omission}. Writer rows are inherited
integration observations, not newly implemented by the reader. history-expired and
cursor-after-withheld remain unqualified. Event SSE must never count as changefeed.sse
or changefeed.sse-reconnect; SSE disconnect tests must not count as
transactional.http.disconnect-admitted (mutation survival). No claimed complete
public profile follows from finite Event delivery.
