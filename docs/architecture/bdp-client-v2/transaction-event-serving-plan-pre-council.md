# Transactional Event serving: bounded T3 implementation plan

Plan only. Basis: `71549f2610d61e6562b2c703d96e4c25d1093492` in
`vickie-installed-profiles-integration`, inspected 2026-10-13. Root advanced through
RU artifact/test merges during inspection; no changes were made to any checkout.
This plan addresses required finite and SSE **Scope and Resource Event Sources**.
It does not implement optional revision History, Snapshot or lossless changefeed
serving, erasure administration, owned-Type lifecycle, or public profile admission.
No tests were run: this is a source/spec planning audit, not qualification.

## Recommendation and authority boundary

Add one source-private Event observation facet to the existing Transactional
owner, one service lifecycle adapter, and finite/SSE HTTP projections in its
existing private runtime. Start with its fixed all-visible, unowned configuration,
including compiled unowned Types. Keep all committed groups for this first slice;
receipt expiry must never delete Event facts or their continuation identity.
This provides complete finite/live delivery for that supported configuration,
not a complete Transactional profile or universal authorization/erasure coverage.

The owner remains the sole authority for committed prefix, epoch/view, durable
facts, Resource identity and fencing. The service owns cancellation, waiting and
shutdown. HTTP owns negotiation/framing/backpressure, not a second event log.
Do not expose `inspect()` as an implementation port or derive current history from
receipt results. An Event Source survives subject deletion independently of the
current Resource read path.

## Existing law and reusable code

Normative references in `docs/specs/bdp.md` at the basis:

- `#events-and-event-sources` (5350 onward): deterministic group projections,
  source-local IDs, five verbs, operation order, all-or-nothing commit visibility,
  incident and self-Link facts, finite induced-Event bound.
- `#event-replay-and-live-observation` (6130 onward): Scope `events/` and Resource
  `?view=events`, exclusive source-local `after`, omitted start at earliest retained,
  finite `{source,events,next}`, SSE one complete Event per message, and live
  `Last-Event-ID` precedence. This is **not** changefeed `start=now` or mandatory
  explicit genesis intent, and an Event message is **not** a complete change group.
- Scope history/authorization fences (3652–3706): epoch/view-bound history tokens,
  stable normal restart, equality-only client positions; view token is not a credential.
- Transactional consistency/conditional reads (5760–5830) and finite-feed validators
  (6536–6561): same-entry current context, minimum-before-conditionals, HEAD decisions
  and bodylessness, finite validator omission, SSE `no-store, no-transform`.
- Deleted Resource Events (5630–5642): retained source may remain available; unknown
  or undisclosable source stays 404; 410 event-history-expired is narrowly authorized.
- Erasure rules (6430–6460): withholding leaves ordinal gaps and valid exclusive
  cursors; deleted lineage facts survive. This is a future integration constraint,
  not permission to invent an Event erasure record.

Existing seams, all under `packages/server/src/`:

- `transaction-store.ts:1059–1079` commits group plus graph/receipt in one atomic
  owner transaction; `groups(position,body)` is durable. No-op, alias-only and failed
  mutations do not invent groups. Existing guarded entry/fence and `readContext()`
  must also govern Event reads; ordinary restart keeps the same epoch/view.
- `transaction-batch.ts:435–446,543–556` defines ordered Event facts, with current
  Scope IDs `pos-N-ordinal`; `:644–659` supplies linked/unlinked endpoint facts.
- `transaction-read.ts` and `authority-read.ts` demonstrate detached same-entry
  observations, source conditions and continuation context. Reuse the ownership
  discipline; do not route Events through the current Resource-existence check.
- `transaction-consistency.ts` provides checked epoch/view/minimum comparison and
  bounded retry with an original deadline. Reuse for initial minimum admission;
  live SSE waiting is separately cancellable, not repeated catch-up-timeout polling.
- `transaction-service.ts` tracks in-flight work and drains before owner close.
  Extend it for live consumers; otherwise a never-ending stream prevents shutdown.
- `transaction-http.ts` and `transaction-runtime.ts` already own exact Scope routing,
  raw target/Host/auth/Origin gates, method/HEAD handling and non-attesting markers.
  Keep these gates ahead of Event observation, just as for other private routes.
- `read-http.ts` has conditional and JSON negotiation mechanics, but selects only
  JSON. `HttpResponse` (`index.ts:348`) and all current Node delivery paths serialize
  a finite body and call `end()`. There is **no reusable SSE server/stream writer**.
  Reuse small conditional/media parsing logic where sound; introduce a private
  streaming response branch rather than widening the public Read export or making
  an async iterator masquerade as JSON.
- Protocol bundle `$defs/event` and `$defs/eventPage` already exist. Protocol
  `transactional-wire.test.ts` and `transactional-http-contracts.test.ts` provide
  wire/example assertions, not executable evidence that this runtime serves them.

## Smallest implementation sequence

1. **Durable projection and token contract.** Add a private Event module for source
   identity, token construction/validation, finite projection and framing. Capture
   facts and current context under one owner entry; freeze detached outputs. Scope
   order is committed group order then stored ordinal. Resource projection retains
   original ordinals and transaction identity, rewrites source and source-local ID,
   and filters facts without renumbering or committing another log.

   Link sources contain their lifecycle facts only. Bead sources contain their own
   lifecycle, the applicable endpoint-specific linked/unlinked facts (two for a
   self-Link), and updates of incident Links. Link lifecycle facts are not copied
   indiscriminately into each Bead source. An incident Link update is the same
   semantic group fact projected into the Bead source, not an extra mutation or
   Bead revision. Metadata/provenance remain whatever the retained Event contains.

   **Required implementation dependency:** historical incident membership cannot
   use only today's Link table. An updated Link Event lacks endpoints, and the Link
   can subsequently be deleted. Recover immutable endpoints from retained creation
   facts (a bounded per-entry lookup/index over durable groups is a valid first
   implementation), including create/update/delete within one transaction. Test
   after restart/deletion. If adding an index, it is a projection of group facts,
   committed atomically or deterministically rebuilt; it is never another authority.

   Existing `pos-N-ordinal` tokens alone do not fence epoch/view/source. Use one
   deterministic private codec over epoch, view, source, checkpoint and ordinal,
   obeying the 256-character token profile. Verify a candidate names an actual
   retained position belonging to this source; token shape alone cannot authorize
   replay. The Scope Event projection must be the same one later used by a served
   change group. Do not silently change persisted historical IDs and then claim
   byte-identical replay. If the selected implementation changes stored group/ID
   semantics, version the private configuration/format and refuse old stores before
   BDP maintenance; no migration/reseed. Derived serving IDs need no second log.

2. **Owner read operation.** Add finite observation with source, optional exclusive
   cursor and bounded page count. Perform source disclosure, cursor ownership,
   minimum and prefix selection under the same guarded entry; never hold SQLite
   across await/network writes. Return current context plus detached Event page.
   `next` uses the last returned Event only when more matching retained Events exist;
   null means current end, not permanent end. A valid after-last cursor returns empty
   current-end page and remains usable when later transactions commit. One transaction
   may span finite Event pages: visibility is atomic commit, not a prohibition on
   Event-page boundaries. Keep configured page/memory limits honest; do not collect
   the entire log into `inspect().groups` on every request.

3. **Live service observation.** Start replay and follow from the same selected
   cursor/prefix without a race between replay end and subscription. A small
   owner-aware wakeup or cancellable bounded polling loop is an implementation
   choice; do not keep an unbounded per-subscriber Event queue. Every next observation
   goes through the owner fence/current policy context. A commit during replay or
   just before waiting must be seen once on the connection, or on exclusive reconnect.
   Register a stream lifetime cancellation handle, stop it on socket close/abort,
   and cancel before draining service shutdown. Respect backpressure and release all
   handlers/timers on finish, failure and owner fencing. No new wire error message
   after SSE headers: close on subsequent failure and let reconnect run normal
   admission. A frame written to a socket is not evidence of consumer application.

4. **Actual private HTTP.** Route exact Scope `events/` and canonical Bead/Link
   `view=events` independently of current subject existence. Accept finite JSON or
   SSE on the same source, negotiate honestly, apply live Last-Event-ID override,
   and emit complete `id`, `event`, `data`, blank-line frames. Preserve query/header
   token validation and current context on successful/conditional responses. HEAD
   does the same admission and returns metadata immediately without opening a stream.
   Omit finite ETag/Last-Modified; apply existing preconditions only after all source,
   cursor and barrier decisions. Keep current Origin refusal (no CORS support claim).
   Source-private operation URLs are test-discovered via fixture wiring; do not fake
   full Transactional discovery while Snapshot/changefeed remain absent.

## Executable done predicates

Use disposable owners and actual Node HTTP; add focused Event store/projection,
service and HTTP tests. Existing owner/recovery/minimum tests remain regression
neighbors, not replacements for live observations.

1. Create/update/delete Bead and Link, both local endpoints, external endpoint,
   self-Link, pinned endpoint, properties and metadata changes; independently
   assert exact Scope and Resource lists, ordinals, transaction, payload and source.
   Include deleted Link historical updates on each Bead source; no synthetic Bead
   revision. No-op/alias/failed/late-rollback transactions yield no Event or position.
   Applicable set core ordering remains canonical URI; HTTP set admission stays
   governed by its separate unfinished slice.
2. Paginate across and inside groups, repeat pages, empty/current-end reads, then
   append; restart with different listener and stable canonical Scope reproduces
   IDs/order/next and retained deleted-subject history. Expire receipt detail and
   prove Events unchanged. Scope token on Resource, one Resource token on another,
   foreign epoch/view, malformed/genesis Event token and fabricated ordinal fail
   explicitly without disclosure, fallback-to-start or silent skipping.
3. Cold HTTP finite reader accumulates exact expected Events; SSE reader parses
   actual frames and compares the same facts/IDs. Conflicting query versus
   Last-Event-ID proves live header precedence. Disconnect mid-frame and reconnect
   after the last completely parsed Event; prove no missed committed Event. A
   controlled commit at replay/wait boundary is the race counterexample.
4. Initial minimum ahead waits then succeeds at a current owner context, or reaches
   existing catch-up-timeout; epoch/view failures beat conditional success. Unknown
   source still nondisclosing. Reproduce HEAD finite/SSE, Accept exclusions/bodyless
   406, method 405, If-Match/If-None-Match behavior without validators, forbidden
   Origin, unauthorized requests, and no body/stream on HEAD/304/412.
5. Actual slow socket/backpressure, disconnect, concurrent streams and shutdown
   terminate boundedly without blocking mutations, leaking timers/listeners, or
   reading a closed owner. Inject owner fence and rollback; no uncommitted Events
   escape. Restart after committed crash gives the same complete facts. Assert
   source handlers do not bypass the same-owner context/fence.
6. Old-format/config refusal, if introduced, leaves normal incompatible store bytes
   and pending claims unchanged before BDP cleanup. Preserve legitimate SQLite hot
   journal recovery and distinguish that from authority maintenance.

## Gaps, decisions and honest claim boundary

No product ruling is needed to implement the supported all-visible/no-pruning
slice. Full Transactional acceptance still requires policy-dependent pre/post
visibility, owned-Link closure/facts, erasure withholding and stable withheld-event
cursors, replay retention/fences, Snapshot/changefeed serving and full discovery.
Keep these named as unqualified; successful unowned tests do not close those rows.
All-visible projection is shared across authenticated principals; do not incorrectly
apply principal-owned receipt lookup to Scope Event visibility.

No time-based Event pruning is needed for this slice: preserve all existing groups.
Do not borrow receipt expiry, collection cursor TTL or changefeed minimumReplayPosition
as an invented Event expiry rule. If pruning is added later, authorized
`event-history-expired` versus cursor expiry and retained deletion identity need
explicit tests. This first slice cannot claim those expiry cases by manufacturing
receipts or directly deleting database rows.

The spec clearly requires explicit cross-source/epoch/view cursor refusal and the
shared token grammar. It does not spell out every precedence combination (for
example malformed/unknown source cursor versus a simultaneous unsupported minimum).
Reuse established problem categories and existing source-disclosure rules; record
an unresolved exact combination if implementation needs one, rather than inventing
new error codes or claiming a generic all-input ordering theorem. Pure syntax and
unknown retained identity are distinct test inputs.

Relevant catalog rows include `transactional.event.graph-facts`, `.self-link`,
`.set-expansion-order`, `.history-expired`, `.cursor-after-withheld`, and shared
`transactional.http.finite-validator-omission`/HEAD/conditional rows. Existing
`transactional.changefeed.sse` and `.sse-reconnect` require **group** delivery and
must not be marked covered by Event SSE tests. New tests can cite Event prose for
currently unbound serving clauses without changing public coverage/evidence claims.

Likely files: new `transaction-events.ts` and focused tests; guarded methods in
`transaction-store.ts`; lifecycle methods in `transaction-service.ts`; private
response union/route in `transaction-http.ts`; streaming delivery/cancellation in
`transaction-runtime.ts`. Only extract a small shared HTTP helper if justified by
actual JSON/SSE negotiation/conditional duplication. No public Read response/API,
protocol schema, RU owner, compiler, installed artifact capture or Type lifecycle
changes are required. Root assigns the single TX owner writer and integration;
this report authorizes no code changes.
