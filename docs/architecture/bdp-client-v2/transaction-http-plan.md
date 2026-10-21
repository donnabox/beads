# T2b bounded private Transactional HTTP

Base bddc4486fe39676bb84d63e05dda22f1bfa77a1a plus ordinary merge of exact GitHub-fetched PR86 03b7f435ef1a2c433dcf37b4910d5b7c75a91ed2. Checkout /Users/dbox/repos/wt/vickie/vickie-transaction-http; branch codex/vickie-transaction-http. Previous T2a checkout remains frozen. This is an implementation slice under the sole Vickie fork design plan, not a competing plan of record.

## Scope and file ownership

- New packages/server/src/transaction-http.ts: source-private request/response projection over one TransactionService; canonical operation and receipt paths, supported canonical Read routes using reviewed read-http-routing/read-http helpers, direct and sequence receiving boundaries, trusted authorization argument independent of attribution.
- New transaction-runtime.ts: source-private loopback Node listener around that handler/service; finite body/raw-target bounds, trusted credentials, explicit owner provisioning and graceful accepted-work drain. No public startup/configuration/export option.
- New tests for HTTP and runtime, with source-only test-support process launcher if required. Tests must use actual one-owner stores and real loopback traffic, not mocked receipt/position responses.
- Narrow transaction-service.ts addition only if needed to expose trusted preparation/Scope facts and a current sequence observation; no new owner/key policy. Any store write or shared authority-read/pagination seam requires root coordination before editing.
- Progress note and smoke private emitted inventory/deep-import/root-negative checks.
- Do not edit P1a's development-runtime, development-authority, installed factories/configuration, shared Read routing, U2 contracts, evaluator, protocol/schema or public exports.

## Required wire matrix and bounded limits

Current bdp.md 4485ff orders direct requests: bounds, carrier/media/key syntax, principal, authoritative key comparison, fresh-key capacity, response negotiation. Existing submitTransaction provides the key/admission portion. The HTTP receiver must not pre-negotiate a duplicate ahead of conflict or bind any refused key. Mutation HTTP conditions do not guard commands. Sequence has one whole-carrier syntax/preflight before admission, per-member keys and results; no carrier Idempotency-Key or sequence-level receipt/202. Existing prepared sequence/store preserve all-or-none key reservations and separate ordered commits.

Current 4600ff: receipt GET/HEAD200 and auth401; uniform404 for foreign/unknown/retracted/forgotten/prior-epoch identities; receipt root GET/HEAD404 and other methods405; command targets allow POST only; every receipt uses private,no-store, no validators and current serving epoch/view/position, not retained execution facts. Pending duplicate may return202 with Location/Retry-After; terminal failed receipt remains200. HEAD and conditional decisions reuse existing applyReadHttpSemantics after ordinary access/query checks. No synthetic ETag or Last-Modified is added to receipts.

Current 4618ff and5758ff require authentication then receipt nondisclosure, then minimum context/wait, then page expiry/representation; all read context/erasure/expiry checks precede304/412. T2b will not claim minimum-position support: any supplied minimum-context controls fail closed as source-private unavailable behavior, after receipt nondisclosure on receipt routes and before a successful/conditional ordinary Read result. This is not a qualified catch-up wait or substitute conformance result. Immediate T3 must implement the narrow same-entry precondition/barrier seam, current epoch/view and original retained-page context ordering, positive wait/routing outcomes, foreign-view/cursor-expired/catch-up-timeout failures and304/412 current context. No authority-read/pagination edits are authorized here. No field will be silently ignored to return stale200/304.

The current TX directory schema requires updateWhere/deleteWhere, and TX discovery requires snapshot/changes/events/minimumReplayPosition. Those implementations are absent. Root confirmed: no schema-invalid TX descriptor, no invented development navigation protocol, no fake directory links. The source-private fixture returns canonical route URLs/operation inventory out of band to tests. Scope/service-description/Operation Directory discovery remains unimplemented and explicitly fails closed. Canonical Resource/operation/receipt URLs retain the eventual protocol route shape. Cold schema-valid TX discovery is an unfinished done predicate for the overall workstream, not waived by this increment.

Fixed all-visible, schema-free/unowned policy remains explicit. Service requires page.maximumItems>=32 only because supported inline receipts can have32 results and no durable pages exist; no protocol minimum is inferred. Set operations and sequence Resource-alias operands remain unsupported before effects. Raw unsupported development categories must not be mislabeled protocol malformed syntax. Receipt paging, variable grants/erasure, installed contracts, required Snapshot/changefeeds and full discovery/qualification remain later work.

## Executable done-when

1. A loopback private runtime with one SQLite owner accepts real singleton/batch/sequence requests at canonical URLs, serves canonical afterstate and receipt reads from the same service, closes cleanly and reopens exact retained key/receipt state. No public config/profile option starts it.
2. Direct precedence tests combine oversized/malformed/media/auth/key conflict/capacity/Accept cases; inspect actual retained state/receipt identity to prove refused keys are not captured and equal retries do not execute twice. Sequence tests prove carrier syntax/key rejection before effects, member ordered independent commits, canonical @name dependency and exact-key replay across carriers.
3. Receipt GET/HEAD tests prove current response context versus unchanged original body after later mutation/deletion; no validator; applicable304/412 and HEAD metadata/bodylessness; unknown/foreign/prior receipt paths share404 and cannot disclose minimum-context information. Receipt root/method tables match supported paths.
4. Unsupported minimum controls cannot yield200/304; test them against successful Resource/collection/receipt paths, expired cursor and foreign receipt cases. Report these as refusal safety only, not barrier qualification. Conditional request fields on POST cannot prevent execution.
5. Actual network response disconnect does not cancel admitted work; graceful runtime close drains accepted work and store reopen proves terminal outcome. Bounded process launcher/restart tests use emitted build; label process exit/SQL injection honestly, no power-loss claim.
6. Build first; focused HTTP/service/store/Read/runtime and inherited regression tests with2workers; typecheck/boundaries/format/lint, installed/development smokes and unchanged evidence verification. No test processes remain. Freeze exact clean commit, external report and next gaps for independent review; root publishes.

## Accepted review dispositions (supersede interim response choices above)

Canonical identity is an optional explicit validated private Scope configuration,
independent of loopback binding; absence is only a disposable same-port fixture.
Actual tests use a canonical HTTPS identity and different ephemeral bindings,
without rewriting any retained URL. Public installed configuration is untouched.
Unsupported valid set/depth/alias-sequence and minimum-control categories return
bodyless internal 500 with a development marker, not invented malformed/limit or
temporary/catch-up outcomes. Schema-valid discovery remains absent with the same
honest internal gap. Genuine finite unresolved comparison maps to 503
 temporarily-unavailable and Retry-After. Context on implemented successful
Read/receipt/304/412 paths is implemented now; T3 only owns remaining barrier,
expiry/retained-context ordering and complete required consistency behavior.

External 16-finding disposition:
1. Fixed canonical identity separated from listener; no fixed-port law adopted.
2. Unsupported valid request -> bodyless500, no key/effects; set targets absent.
3. Actual Read/receipt200/304/412 current context covered now.
4. No fabricated catch-up/temporary proof. All supplied controls refused as an
   explicit private incomplete behavior, including already-satisfied minima.
5. POST never enters Read conditional handling; receipt delivery preserved.
6. Target method routing precedes generic Read and receipt existence checks.
7. Close stops delivery/admission and drains accepted work; injected uncertain
   commit fences owner, live routes fail bodyless500, reopen recovers pending.
8. Body cap shared; target cap8192 applies to POST, not a blanket Read413.
9. No blanket rule omitting context from every error adopted. Existing Read
   observations retain their actual context; unauthorized receipt nondisclosure
   exposes none. T3 must complete ordering, not add diagnostic inspect reads.
10. Actual unresolved dependency comparison503, no disclosed receipt/second
    execution; later equal completion200 verified.
11. Stray sequence carrier key rejected400 before member admission.
12. Prior-epoch guard tested by persisted-fact injection, no rotation claim.
13. Reuses target normalization, Read semantics, acceptsJson and body cap.
    Node listener shell remains deliberately separate from concurrent P1a work.
14. sequence/sequenceObserved share one runner and lifecycle task semantics.
15. Same-owner afterstate uses exclusive SQLite authority, not a replica barrier.
16. Original submissions terminal;202 exercised by equal pending member retry.

T2a external findings F1/F2 repaired by same-entry execution/replay observations,
no delivery clock reread or receipt-ID scan; GET scan remains a performance gap.
F3 enforces serving position>=required/effect. F4 guards stored prior epoch. F5
sequenceObserved is implemented. F6 scheduler eventual-delivery precondition is
explicit. F7 uses one unchanged cap and actual inline result guard. F8 new guard
failures retain receipt ID/key; the old delivery-unavailable path is removed.
