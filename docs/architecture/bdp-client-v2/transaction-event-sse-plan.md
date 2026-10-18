# G2b: bounded Scope/Resource Event Source SSE

Council proposal, 2026-10-13. Read-only planning complete; **no implementation authorization or profile qualification**. The sole fork design remains the plan of record. Root must assign a reviewed integrated base before authoring. Frozen G2a `b64d7e46652739c6436259a8c8ed4c5e8049a077` remains unchanged; its outstanding external review may require a predecessor repair.

Inspected checkout: `/Users/dbox/repos/wt/vickie/vickie-changefeed-sse`, branch `codex/vickie-changefeed-sse`, clean at b64. Root notes: `transaction-event-sse-design-notes.md` (0944 references reconciled with b64). Previous paused WIP preserved byte-identically as `transaction-event-sse-plan-paused-deb205.md`, SHA256 `deb205bd0004caee44ab0e0536f7743c8081c3f5f73f828406090eadc315aee8`. Resume remote guard still reports session `2026-10-08T20:26:35Z-81e5`; captured in `transaction-event-sse-resume-guard.md`.

## Objective and boundary

Serve Scope `events/` and canonical Bead/Link `?view=events` as SSE over the existing one-owner projections. Reuse G2a negotiation, Node disconnect signal, stream capacity/lifecycle, write/drain pump and installed recorder. Add only private bounded scan progress so idle narrow Sources do not repeatedly traverse unrelated history. No new persistent log, wire cursor, subscription endpoint, selector behavior, storage format, public export or CLI/profile flag.

The deliverable is a private fixed-view serving slice. Erasure withholding/publication fences, variable grants, pruning/expired-history policy, set HTTP admission and full discovery remain separate required work. Metadata-only incident-Link membership remains the existing explicit unresolved behavior, not a new G2b ruling.

## Normative basis and current implementation

Citations below refer to b64 `docs/specs/bdp.md`; anchors are the durable references.

| Contract | Exact source / consequence |
|---|---|
| Deterministic projections, not separate logs | [Events and Event Sources](docs/specs/bdp.md#events-and-event-sources), lines5350–5373: source-local identity derives from group checkpoint/ordinal/projection; an index/materialization may preserve those facts. |
| Membership and self-Link | lines5547–5565: Link lifecycle; Bead lifecycle/incidence/incident Link properties updates; self-Link has two endpoint facts. Projection must retain original group ordinals, not renumber a filtered Source. |
| Deleted/unknown Source | [Events after Resource deletion](docs/specs/bdp.md#events-after-resource-deletion), lines5629–5640: retained subject history may outlive the Resource; unknown/out-of-view404; history-expired410 requires authorization and disclosure policy. |
| Initial cursor and ordering | [Event replay and live observation](docs/specs/bdp.md#event-replay-and-live-observation), lines6130–6190: after is exclusive, source-local; omission means earliest retained; finite next is after last returned Event and null at current end. |
| Complete framing | lines6251–6285: id/type repeat full JSON Event, complete data record, blank-line delimiter; comments/retry optional; Last-Event-ID overrides original after for live delivery. |
| Observation is not replica/erasure proof | lines6225–6250: mutation Events appear after one commit, but individual delivery is application observation; atomic replication uses groups, erasure reconciliation uses changefeed/snapshot ledger. |
| Initial consistency and HEAD | [Transactional HTTP consistency](docs/specs/bdp.md#transactional-http-consistency), lines5766–5813: current same-entry epoch/view/position; minimum and expiry/erasure decisions before conditional shortcuts; HEAD ends without frames. |
| Validators and cache | [Receipt and finite-feed HTTP validators](docs/specs/bdp.md#receipt-and-finite-feed-http-validators), lines6536–6559: finite Event pages omit ETag/Last-Modified; SSE no per-group validator; GET/HEAD/405, conditions/current context; no-store,no-transform for SSE. |
| Token distinction | [Event-ID and checkpoint character profile](docs/specs/bdp.md#event-id-and-checkpoint-character-profile), lines6563–6576: restricted ASCII, case-sensitive, max256; genesis is never an Event ID. |

Current owner `transaction-store.ts:1370–1470` checks retained source identity, minimum/context, then actual cursor membership; scans committed groups through `retainedGroupReader` (`:560`), projects through `projectTransactionEventGroup`, and bounds finite pages to installed receiptMaximumItems (default32) and4MiB. Statements are prepared once per owner observation, length checked before body parsing, and no iterator escapes.

The current finite API is correct for a finite read but inefficient as a polling primitive: calling `service.events(source,lastEmittedID)` repeatedly rechecks that Event's group and every later unrelated group. After an idle sparse prefix of N groups, each poll costs O(N) again. A Resource Source is not assumed empty before its creation Event; the sparse test begins after that real Event. The Scope Source can be empty at a fresh head0.

## Exact private owner seam

Add a stream-specific begin/advance pair on the existing owner, with source-private result types in `transaction-store.ts`. Working names below are implementation names, not wire schema:

- `beginEventScan(sourceUrl, effectiveAfter?, minimum?, invalidReconnect?)`: one guarded entry; returns missing/unsupported as today, or `{context, events, scan, caughtUp}`.
- `advanceEventScan(scan)`: one guarded entry; returns the next bounded `{context, events, scan, caughtUp}`. It accepts no caller-chosen source, numeric position or public Event ID.
- `scan` is an opaque immutable owner-issued object authenticated by an owner-local WeakMap. Its inaccessible state contains canonical source descriptor, epoch/view, next group position, exclusive ordinal within that group (or -1 before its first Event), and observed head. The map holds no strongly enumerated list of tokens. Tokens from another/reopened owner, fabricated tokens, or closed/fenced owners are refused. There is no durable token, table, schema or migration.

`begin` performs the existing source identity/minimum/cursor checks **inside one owner entry**, including actual projected membership for supplied after. It returns no scan capability when those checks fail or wait. A malformed/repeated raw reconnect field must use this same ordering, rather than introducing a pre-owner unknown-source disclosure decision. Recommended minimal seam is an explicit private `invalidReconnect` flag: after known-source and minimum checks, throw the existing `TransactionEventParameterError`; it is never a public token. Preserve existing URL/query syntax rejection in `transactionEventRequest` before owner access. Do not convert unknown Source404 into a cursor400 merely because a reconnect header is malformed.

`advance` authenticates the capability and rechecks retained source/context under `enter`. The current implementation has fixed epoch/view and no pruning; a head less than the capability's last observed head is an internal integrity fault. Changed context invalidates progress rather than silently continuing under a new projection. A reconnect is a new begin and always validates its public ID; the private fast path is never reachable from a request parameter.

No callback, statement, parsed full group, iterator or active SQLite transaction lives in the capability. No mutable global last position, diagnostic inspect adapter or second authority is introduced. Fixed view is the identity boundary, not a fabricated per-principal Event namespace.

### Bounded scan algorithm and no-skip invariant

Use at most **one group per owner entry**. This is a private scheduling/materialization bound, not a new protocol limit or public page guarantee.

1. Capture current context/head and validate access/cursor/progress.
2. If next position is beyond captured head, return an empty caught-up packet without selecting/parsing a group body.
3. Otherwise fetch exactly the next expected retained group, with the existing length/context/framing checks. Missing/gapped rows in the known retained prefix remain internal faults; an unissued initial public cursor retains its current invalid-parameter convention.
4. Project the whole group using the existing renderer. Consider source Events in stable ordinal order strictly after the capability's ordinal. Never renumber filtered ordinals.
5. Collect Events under both existing installed item count and finite Event page byte bounds. If another relevant Event would exceed either bound, return progress at the **same group after the last included ordinal**. Do not advance beyond that unreturned Event. A first Event that cannot fit remains the existing admitted-bound invariant fault; no partial Event or new refusal code.
6. Only after all remaining relevant Events in that group have been considered may progress become next group/before-first ordinal. This is also legal when there were no matching Events, because this entry proved that prefix empty for the bound Source/context.
7. `caughtUp` means the progress has traversed the captured head. It does not promise there was no commit after the entry returned. The next advance observes a fresh head, so such a commit is read normally.

Initial actual-cursor membership and remaining Events of its group are handled from the same parsed projection, not two parses. A page boundary before self-Link's second endpoint fact stays in the group. An empty scan never emits an invented Event/head/checkpoint ID. Last public Event ID and private inspected prefix are intentionally different concepts.

Extract only the smallest per-entry projection/collection/byte-accounting helper needed to share this logic with finite readEvents. Finite calls keep scanning until their current item/byte/end rule decides next/null; stream calls stop after one group. Preserve finite lookahead: exactly filling the item count is not automatically a non-null next when no later matching Event exists. Do not replace finite behavior with the stream's caughtUp flag or expose that flag on the wire.

Previously traversed immutable history is not repeatedly rehashed on idle polls. This optimization relies on the current owner's append-only committed groups; every newly inspected group retains existing validation. Arbitrary out-of-owner edits to an already traversed old row are not a newly promised live detection mechanism. Future pruning/erasure will need its separate invalidation/fence contract before this fast path can be qualified there.

## Service, pump and transport composition

### One pump, two payload adapters

Refactor `transaction-sse.ts` narrowly into one private packet/write/drain loop plus existing changefeed and new Event adapters. No generic subscription framework.

A packet carries bounded ordered payloads, the prospective next private continuation, and whether its captured prefix is exhausted. The loop:

1. Encodes/writes one complete frame at a time. Event frames are exactly `id: <event.id>`, `event: <event.type>`, `data: <stringifyJsonValue(event)>`, blank line.
2. On write(false), await drain/error/close/abort before another frame, comment or owner pull. Comments use the same write path; elapsed heartbeat ticks never enqueue behind a blocked frame.
3. Adopt the prospective continuation only after handling all frames in the packet. If disconnected partway, discard the packet/progress; reconnect uses a public ID and a fresh owner validation.
4. An empty packet with remaining historical work advances only private progress and yields one cancellable Node turn before another bounded owner step. It does **not** sleep100ms for every historical unrelated group. At the captured end use the existing100ms poll. Heartbeats still use the existing5s monotonic interval and never carry IDs. Yield between bounded backlog packets so timers/disconnects can run.
5. Abort/fault propagation, post-header socket closure and listener cleanup stay in the reviewed G2a path. The global private one-live-stream cap covers both changefeed and Event streams, not one slot per Source.

Keep `pumpTransactionChanges` as the thin existing typed adapter so G2a tests remain meaningful. Add `pumpTransactionEvents` (or an equally small private adapter), no duplicated loop. The stream-specific `service.beginEvents`/`nextEvents` wrappers register each finite owner observation with the existing `run`/minimum machinery; the infinite pump belongs only to `service.stream`. Service.close rejects new work, cancels streams, drains accepted finite operations, and closes the same owner. No minimum deadline restarts on an initial wait; after it is satisfied the fixed epoch/view and monotonic head preserve that minimum during continuation.

### Initial HTTP decisions

- Negotiate through the reviewed media helper; JSON remains default/tie winner. The finite JSON branch ignores all Last-Event-ID values and uses existing `service.events` unchanged.
- For selected Event SSE, a single Last-Event-ID overrides after. No after plus a valid header is supported because after omission is already legal. No header/no after starts earliest retained. `genesis` is invalid as an Event ID, unlike the changefeed's reserved initial checkpoint.
- Preserve source/method/auth/query/minimum/cursor ordering and existing reference error conventions. No new error code. Malformed, empty, repeated, source-mismatched, unissued, stale and foreign-view cursor cases get discriminating tests; underlying never-issued400 convention is not presented as a universal normative ruling.
- Initial packet/context comes from the owner entry before negotiation refusal or conditions. HEAD/304/412 create no subscription/slot and drop any ephemeral scan capability. Unknown Source returns404 before contextual owner checks as the finite path does; the deleted Resource Source remains available while identity/history is retained.
- Successful SSE uses the G2a context/cache/marker/Vary headers, without Content-Length/ETag/Last-Modified. Scope and Resource routes use the same rendering/stream lifetime. No separate Event endpoint or per-group URL.
- Original query grammar still must be valid even when a selected header overrides its cursor value; a syntactically valid overridden query is not looked up in the owner. Read selectors and non-Event routes are untouched.
- Metadata-only incident membership: if the initial bounded projection encounters the existing gap, preserve private failure before headers. If a later step encounters it after headers, close through the existing fault path; no skipped fact, invented Event or second HTTP response. Scope/Link projection remains available. This does not qualify the missing Bead membership contract.

## Memory and work accounting

Reuse `TRANSACTION_EVENT_PAGE_BYTES = 4,194,304` and the existing installed item cap; do not retain a full history array or a parsed group in the scan token. A packet's encoded aggregate Event content stays below that page budget. Source-private Event frame bound is at most page bytes+512 = **4,194,816**, because id<=256, type<=8 and fixed framing overhead is under512. This is below pinned Node24 MAX_STRING_LENGTH and uses the existing first-page feasibility proof (1,778,688 bytes), not the much larger changefeed envelope as a shortcut.

During an owner entry, one admitted group can still be materialized up to the existing `TRANSACTION_EVENT_GROUP_BYTES` bound169,345,024 before filtering. That is inherited projection cost, not a4MiB total-memory claim. The packet keeps only selected Event records afterward; the pump retains one bounded packet and one encoded pending frame. Encoded-byte bounds are not JavaScript heap/RSS or latency bounds. No large proof allocation, count/byte substitution or post-commit rejection is introduced.

The group-step limit and packet-byte bound should be named source-private constants in `transaction-events.ts`, consumed by owner and runtime configuration metadata rather than duplicated. Extend captured settings with these actual Event stream mechanics; preserve poll/heartbeat/live-slot settings and all four Node timers. There is no installed user-configurable switch. Source-private small fixture limits may test boundaries with metadata suppressed, using existing page-cap seams where possible.

Sparse-work oracle: seed one Source Event and N unrelated groups, begin after that real Event, count group-body reads through the existing DatabaseSync seam while advancing. Each newly traversed unrelated group is inspected once; after caught-up, repeated idle advances select **zero historical bodies**. Adding one unrelated group adds one body inspection, and a later relevant mutation is delivered once. A large multi-Event group may be reparsed for each bounded packet, but that work is bounded by its event count/byte budget and never rewalks the unrelated prefix. No timing-only performance assertion.

## Exact proposed ownership

| Files | Bounded change |
|---|---|
| `packages/server/src/transaction-store.ts` | Private begin/advance capability, shared finite/stream per-entry scan helper. No write path, config/storage format, tables or evaluator change. |
| `transaction-events.ts` | Event packet/frame/group-step constants and only minimal private types/helpers; existing ID/projection/membership semantics unchanged. |
| `transaction-service.ts` | Finite scan observation wrappers using existing tracking/minimum lifecycle. Reuse one stream registry. |
| `transaction-sse.ts` | Single cancellable packet pump + changefeed/Event adapters; preserve G2a boundaries. |
| `transaction-http.ts`, `transaction-runtime.ts` | Event SSE discriminated result/delivery using existing raw headers/auth/Node bridge. No extra HTTP parser or public interface. |
| `transaction-runtime-configuration.ts` | Actual Event stream mechanics included in immutable settings/hash; timers unchanged. |
| Existing corresponding tests; new `transaction-event-sse.test.ts` | Owner scan/protocol/socket/fault/backpressure tests. Keep finite Event/Read/RU/Snapshot/G2a suites. |
| `apps/bdptest/src/transactional-event-sse-installed.test.ts` | Actual packed installed observation using existing launcher/binding and raw SSE recorder. No H1b runner/finite recorder edits. |
| `apps/bdptest/test-support/raw-sse-recorder.ts` | Only if a bounded observation helper addition is needed; no emulated target responses or terminal finite-success label. Prefer reuse unchanged. |
| `docs/design/transaction-implementation-progress.md` | Executable boundary, measured evidence and explicit gaps. |

No new production module is currently necessary, so no smoke inventory addition is planned. If extraction genuinely needs another emitted module, root must approve its inventory and denied bare/dist imports before editing. `read-http.ts`, `read-request.ts`, public exports, protocol/schema/catalog/fixtures, mutation/evaluator modules, Type and selector work stay untouched. Shared file ownership must be serialized with G2a repairs or Snapshot follow-ups assigned by root.

## Implementation sequence after council approval

1. **Owner core:** begin/advance capability and shared finite scanner, deterministic sparse-work and same-group boundary tests; preserve finite byte/order/next/error tests. No route activation yet.
2. **Shared pump:** Event adapter and bounded backlog yielding, packet-progress adoption, abort/drain/capacity/close/fault tests, unchanged G2a tests. Include metadata settings binding.
3. **HTTP + installed proof:** Scope/Bead/Link SSE route, precise initial/header/HEAD error matrix, actual packed recorder observations, regressions and clean freeze. Root reviews/publishes; no source claim before this evidence.

## Executable done-when predicates

1. Scope, Bead and Link actual SSE frames equal independently observed finite JSON Events in source order; complete id/type/data and stable ordinals/transaction/source fields, real subsequent mutation, no no-op Event.
2. No-after replay, exclusive valid after, Last-Event-ID override with/without query, cross-source rejection, genesis rejection as Event ID, empty/repeated/malformed/stale/foreign/unissued ID and JSON-header ignorance all have discriminating cases. Real mixed-Accept winners included.
3. Self-Link's two endpoint facts and multiple relevant Events from one group survive item and byte packet boundaries. A reconnect between two such ordinals returns the second. Filters never renumber original ordinals.
4. Retained deleted Bead/Link Sources stream their prior/deletion Events while canonical Resource GET stays404; unknown Source404 precedes owner minimum/cursor decisions. No history-expired policy is fabricated.
5. Private capability from a different owner/source/lifetime, fabricated token, closed/fenced owner, regressed head and incompatible context are refused; no scan token or synthetic head Event appears on wire. A source step cannot be pointed at a caller-supplied numeric position.
6. Sparse/empty Source inspection counts prove no repeated historical body scan during idle polling. New irrelevant/relevant groups between observation and next step are neither fabricated nor skipped. One owner entry reads at most one group.
7. Initial minimum wait uses one finite deadline and same-entry context; HEAD/304/412 finish without slot/frame/cursor advancement. Authentication, raw target/method, Origin and metadata-only gap behavior remain explicit.
8. Only one bounded packet and one pending frame are held by the pump; a blocked frame prevents all further owner reads/comments. Multiple heartbeat ticks cannot queue writes; SQLite is released so a concurrent actual mutation commits. Drain resumes at the exact next ordinal.
9. Actual half-frame disconnect and reconnect use the last fully accepted application Event ID. Pre-abort, disconnect during empty backlog/minimum wait/drain, idle and backpressured service close release slot/listeners/owner; already accepted finite mutation still completes and survives reopen.
10. Post-header projection/store/transport fault closes the stream without a second status/new error Event. Metadata-only incident update remains refused for Bead source while Scope/Link facts remain available, both initial and live-after-header cases. Preserve original cleanup causes and no credential disclosure.
11. Exact installed tarballs serve Scope/Bead/Link streams over canonical Host/physical port0 with CSPRNG credentials and fresh validated payload/artifact/configuration identity. Retain actual wire bytes/hash and complete-frame boundaries; no source-only result is labeled installed proof.
12. Installed sparse Bead source survives at least15 seconds idle under actual unchanged timers, optional5s comments with no ID movement, then a relevant mutation arrives. Recorder25s/cohort45s/cleanup10s remain separate from finite5s executor; no timeout widening. Remaining installed source cases use bounded shorter observations, sequentially respecting the shared one-stream slot.
13. Arithmetic tests pin Event frame/page/group bounds and admitted numeric/UTF8/newline serialization; no large allocation needed for proof. Sparse-count performance and byte bounds are described separately from heap/latency promises.
14. Pinned Node24.16/pnpm11.20 clean build, type/boundaries/format/lint, public installed/development smokes, preflight/unchanged Read seal, full affected finite Events/changefeed/G2a/Read/RU/Snapshot regressions, narrated HTTP contracts, fresh installed cohorts and no lingering owned processes. Exact clean source/report for native and Claude review; no public profile or catalog advancement.

## Decisions and limits

No new product decision is required to implement this bounded transport/scan slice. The known incident-Link metadata membership question remains pending and explicit. Erasure/revocation and retention expiry are not solved by scan capabilities; those mechanisms must later invalidate/withhold/fence before publication according to their own laws. Current configured owner cannot trigger variable-view/pruning/erasure transitions, so stale/foreign cursor tests do not qualify such runtime transitions.

The plan does not claim browser EventSource/CORS interoperability, durable replica application from individual Events, full mutable policy, History or Type lifecycle support. No normative edits are proposed. Council should review the private capability lifetime, finite helper equivalence, header nondisclosure ordering and packet adoption invariant before root authorizes code.
