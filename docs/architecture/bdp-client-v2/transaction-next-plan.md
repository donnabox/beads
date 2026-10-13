# Next Transactional vertical slice: current point/receipt minimum reads

Initial read-only inventory against frozen8844fe738d7457d6c9d3ea71b0b8d0fa778eecde;
its parent b1de5f3 contains T2b. Root subsequently approved implementing the clear
canonical complete-tuple point/receipt subset, with no inferred continuation or
partial-header semantics. Sole fork design remains plan of record.

## Existing executable primitives and missing contracts

- transaction-store.ts:135–141 persists Resource rows, identity tombstones,
  aliases, receipts/creator facts and position-keyed immutable group JSON. Its
  executeClaim atomically writes Resource effects, one effective group/events,
  terminal receipt and head. No-effect/failure creates no group. Epoch/view are
  fixed owner metadata; same-entry withRead/readReceipt/observed admission and
  execution capture actual epoch/view/position. readContext validates pos-N.
- transaction-read.ts composes actual owner configuration with AuthorityReadPlane;
  current and retained-page original contexts are separately validated. It has
  no minimum-position decision/wait hook.
- authority-read.ts:observe enters the owner, captures authorization and context,
  then materializes. finish resolves retained pagination after that entry.
  read-pagination.ts retains stable query pages with view/epoch/expiry and bounded
  capacity; these are not a complete Scope snapshot or replay checkpoint.
- transaction-service.ts tracks accepted work and lifecycle; receipt lookup is
  principal-bound, current retention/context in one entry. Same-entry mutation
  delivery is fixed. No receipt pages or asynchronous minimum-read loop exists.
- transaction-http.ts now has correct method shape gate, actual Read/receipt
 200/304/412 context and bodyless internal development gaps for all minimum
 controls. Cold discovery/Directory are intentionally absent.
- transaction-batch.ts prepares/evaluates six Resource and two alias mutations,
  canonical sequence dependencies, final state changes and ordered Events. Set
  operation schemas exist, but actual set selection/cardinality/evaluation do not.
- Change groups and Events exist durably only as write-owner output. inspect()
  is diagnostic, not a safe replication reader. There is no replay floor,
  checkpoint resolver/view+epoch binding, finite group reader, live/SSE lifecycle,
  event-source cursor/index or retention/projection serving contract.
- First-class Scope Snapshot has no manifest/handle/two-stream storage or lifetime
  manager, no checkpoint binding and no promised replay rendezvous. Reusing two
  independently paginated collections would violate bdp.md5734–5753,5825–5957.
  Existing pos-N group spellings alone cannot distinguish a supplied after cursor
  from a prior epoch; exposed replication checkpoints need an actual binding.

Therefore a Snapshot HTTP endpoint is larger than a barrier slice: it requires
manifest identity/refetch/HEAD lifecycle, joint consistent Bead+Link capture,
bounded independent streams, expiry/service-failure distinction and promised
replay continuation. Event/changefeed HTTP is likewise not just serializing
inspect(). None should be advertised from existing diagnostic primitives.

## Approved smallest complete subset

Existing visible canonical Bead/Link Resource and properties GET/HEAD, plus receipt-root-instance
GET/HEAD, with a complete canonical supported minimum tuple. No continuation,
collection/bootstrap, alias redirect, Type artifact or replication serving claim.
Supported current point requests and receipt instances should fully honor:

1. raw target/method and existing authentication handling;
2. receipt principal/epoch nondisclosure before inspecting supplied history
   controls (bdp.md4618–4624);
3. immutable epoch/view/minimum tuple checked against the actual read entry;
4. foreign epoch ->410 cursor-expired, foreign view ->409 foreign-view; lower
   current position waits up to one finite original deadline, then503
   catch-up-timeout with Retry-After; a concurrent actual commit can satisfy it;
5. capture representation and serving context from the same successful entry,
   then existing HTTP conditions/HEAD. No bodyless304/412 may bypass the barrier.

Normative basis:bdp.md3664–3705 epoch/order/view bindings;5757–5763 strict prefixes;
5766–5807 fields/minimum/conditional ordering;5280–5296 catch-up-timeout;
6563–6584 history-token character profile. Existing protocol HTTP fixture checks
at packages/protocol/src/transactional-http-contracts.test.ts:275–295 already
encode foreign epoch/view and actual exhausted-wait witness requirements.

This closes the whole minimum-read contract for that named subset only. It does
not make the overall Read surface or Transactional profile qualified.

## Proposed seam and file ownership

New private transaction-consistency.ts/test: validate/snapshot immutable supported
triples; compare owner pos-N using authority arithmetic, never instruct clients to
order tokens; typed satisfied/foreign-epoch/foreign-view/behind decisions. Finite
monotonic deadline/wait stays outside SQLite. No schema/public limit added.

transaction-store.ts/test: narrow typed receipt-observation attempt accepts the
immutable tuple; checks principal/epoch nondisclosure first and checks minimum
against readContext within that same owner entry before returning any body.
Waiting attempts return no receipt body and capture no mutation/key. Retry the
actual lookup each turn; never keep a stale receipt across an async wait. Existing
readReceipt and mutation observed paths stay compatible.

transaction-read.ts/test + authority-read.ts/test: a source-private per-request
condition seam in the existing plane, evaluated after captured authorization and
after point materialization in its one owner entry but before delivery. An absent
Resource result with controls remains an explicit private visibility gap, avoiding
an unruled 404-versus-barrier order. Constructor supplies fixed trusted
condition evaluator; receiving passes immutable data only, never a caller callback.
TX boundary owns strong tuple validation and typed behind/refusal errors. Read/RU
calls with no source condition keep exact behavior; no TX concepts become required
Read protocol fields. A source-private generic control slot/hook is sufficient;
no second plane per request, global pending condition or last-position variable.
Tests must preserve current public import/declaration boundary.

transaction-service.ts/test: one tracked bounded retry operation outside SQLite,
reusing the same Read plane/store. Retry only typed behind outcomes; preserve
ordinary owner fencing/diagnostic causes, original deadline, shutdown and aborted
GET delivery behavior. No mutation scheduling changes.

transaction-http.ts/test + transaction-runtime.ts/test: snapshot raw header tuples
before awaiting; use point/receipt service path only for approved routes; translate
existing problem envelope shapes (catch-up-timeout is a Transactional direct code,
not readProblem's narrower union), preserve current response context and HEAD/
conditional processing. The stable router receives the actual Request; a pre-route classification gate
rejects alias/probe/discovery/collection/malformed routes with supplied controls;
never mutable cross-request state. Add private emitted helper allowlist/denied
import smoke entries and progress note.

Shared ownership: root must coordinate authority-read.ts with Update/P1; do not
edit development-runtime/authority/config/contracts, evaluator, schemas, Types,
public index/exports or evidence. Minimal read seams only; no store write/format
changes. Implementation is in an isolated successor from7bfba525; root integration/publication only.

## Whole-subset done predicates and executable matrix

- Immediate satisfied current/older min returns actual representation/current
  context; concurrent real mutation advances a waiting read to the required
  position. It can never return a success context below min.
- Real future min exhausts a finite original monotonic budget before503; no guessed
  catch-up-timeout on unsupported controls, no deadline restart per attempt. No
  SQLite transaction held during await; writes remain usable during the wait.
- Epoch mismatch410 and view mismatch409 precede conditions. Fixed all-visible
  owner supplies real context; fault inject only to test failures, not claim grant
  rotation support. Receipt foreign/unknown/prior-epoch remains404 before those
  differences; unauthenticated receipt read remains401.
- Resource/properties and receipt GET+HEAD cover200/304/412 with same current
  context. Failed/expired retained receipt outcomes remain truthful after a wait;
  no expired failed receipt leaks because it was cached from an earlier attempt.
- Parallel callers with different tuples cannot share/overwrite control state;
  early abort and close while waiting release tracked work within the finite bound.
  Owner fencing preserves original/cleanup causes and never returns guessed data.
- No supplied tuple changes a key, mutation receipt or graph state. No no-op/event
  position advancement. Existing private method/POST admission matrix unchanged.
- Existing continuation/no-control behavior and immutable retained page tests pass.
  Incomplete/unsupported header combinations remain explicitly unqualified; no
  complete barrier/profile or public discovery claim.
- Build first; affected owner/Read/service/HTTP and full transaction regressions,
  typecheck/boundaries/format/lint, both smokes and unchanged Read evidence. Local
  clean commit and external exact logs for independent review.

## Unresolved questions to keep outside this implementation

1. A retained query page was anchored at pos1; current serving entry is pos3; the
   client supplies valid minimum pos2. Does the minimum constrain the original
   immutable page anchor or current serving observation? If the former, which
   existing refusal is required for an otherwise-live old page? Inherited stable
   pagination and strict/minimum-prefix clauses do not state this combination.
   Do not relabel old rows as newer or silently rematerialize the cursor.
2. A minimum field with absent epoch/view, repeated tokens, or malformed token
   spelling: character profile is explicit, but receipt response table4602 omits
   direct400 and no tuple-level malformed policy was located. Bare epoch/view
   without minimum likewise has no explicit decision. Inventory general header
   clauses before proposing any normative response; retain private development
   gap for unresolved forms.
3. A valid opaque token from another implementation/unknown same-epoch position
   is not necessarily a numeric pos-N. This owner can interpret only tokens it
   issued. The exact unknown-position refusal is not explicit; do not invent a
   lexical comparison or call inability to decode an actual exhausted wait.

4. An absent or hidden Resource is requested with an incompatible epoch/view or
   future minimum. Does Resource 404 precede the barrier, or vice versa? Receipt
   nondisclosure has an explicit order; current point Resource clauses do not.
   This implementation leaves that combination unqualified and returns only an
   internal bodyless development-gap response, never an invented wire rule.

These questions do not block the approved canonical complete tuple point/receipt
subset. They do block claiming complete minimum-position support across the full
surface. No new product semantics are adopted by this plan.


## External plan review dispositions (14 findings)

Implementation base is 7bfba525, with the command-control and exact operation
routing repairs already present. This plan does not promote narrated protocol
fixtures into executable evidence.

1. Implemented final-attempt captured context on 409/410/503; receipt 401/404 has
   no context. Tests advance the actual owner during an exhausted wait.
2. Corrected predecessor: 7bf includes POST no-key/effect control refusals.
3. Stable shared Read router receives Request; preclassification gates all
   unsupported targets before alias/probe/discovery execution.
4. Actual HTTP Problem bodies validate against canonical transactionalProblem;
   no public protocol parser/type added for this private receiver.
5. Reviewer suggestion of tuple-first on absent Resources is not adopted.
   Parent confirmed genuine missing ordering rule; explicit visibility gap.
6. Inventoried transaction-comparison.ts lines41–54: performance.now, original
   deadline, bounded5ms delay. Minimum waiting uses the same pattern with a
   separate typed behind outcome and finite/nondecreasing clock validation;
   receipt comparison/retraction/key policy is not copied or changed.
7. Existing synchronous receipt method preserved; additive tracked async method
   drains on close. Receipt delivery cancellation does not cancel the finite wait;
   point AbortSignal cancels delivery. No indefinite default scheduling.
8. GET and HEAD use identical waits, final context and Retry-After; HEAD strips
   the body only after selection/refusal.
9. Receipt attempts release SQLite before bounded5ms sleep; no busy-loop polling.
10. Fixture conditions are narrated precedent only. Current executable tests use
    real owner/HTTP/commit/retention facts; no receipt-page qualification.
11. Partial/duplicate/malformed tuples remain internal gaps, not normative400.
    Exhaustive receipt error table alone does not establish a tuple policy.
12. No replica-routing configuration exists, so local waiting is used. Origin
    is refused and CORS disabled; no claim of enabled-CORS header coverage.
13. Point attempts use the existing synchronous exclusive owner mutex; receipt
    attempts use atomic SQLite. No claim of a new SQL transaction on each point.
14. Updated POST inventory to7bf. Earlier tests supplying only minimum position
    still exercise an unsupported partial tuple, and remain bodyless500 tests.

Additional inherited repair review: command controls authenticate/authorize before
privategap, with actual receipt-table empty checks; batch depth marker asserted.
Actual authority-form CONNECT receives400 from raw-target validation; origin-form
405 probes are not claims about standard CONNECT. Non-operation Allow is based
on target shape regardless of existence, preserving nondisclosure.
