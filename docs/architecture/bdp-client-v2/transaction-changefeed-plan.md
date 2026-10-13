# G1: finite whole-group Transactional changefeed

Read-only implementation proposal, 2026-10-13. Original inventory basis was **c0d67388cd849410b38ec7ecb3a328771431dbfa**, already containing owned candidate01e2. Implementation predecessor is now frozen finite Event **8569b85d785f96325ca948580106f454c48ed034**, which includes the mandatory materialization predecessor **076201849f98eeab71bad5c8404fb693a36c6474**. The old c0 counts are not the G1 admission basis. This supplements the sole fork design's T2/T3 completion plan and `transaction-required-contract-burndown.md`; it is not a second authority or qualification report. No runtime, schema, catalog, evidence or frozen checkout is changed by this memo. No new tests were run for planning.

## Bounded outcome and prerequisites

Implement actual finite JSON GET/HEAD of canonical `changes/` over the existing authoritative `groups` table, with whole groups, explicit starting intent, fenced checkpoints, current same-entry context, canonical minimum controls and native conditional semantics. Keep the actual fixed all-visible policy, retained complete history and empty erasure ledger. One complete group per page is an implementation bound, not a new advertised protocol limit. Do not implement SSE, Snapshot creation/streams, policy/view mutation, erasure administration, optional History, set HTTP admission or discovery/profile advertisement here.

Predecessor8569 includes076 and is frozen; root authorized an isolated ordinary successor. Native/external Event reviews remain active: integrate any frozen repairs normally before G1 readiness. Owned transitions are already integrated at c0. The frozen Event source remains immutable; G1 owns successor transaction-store/service/http and reuses transaction-events. Coordinate later review repairs through root, not simultaneous edits to the frozen tree. The materialization lane establishes the admissible committed-group premise; do not impose retrospective transaction or group refusal under a new response budget.

The replication oracle here starts from genesis on an empty graph, whose only administrative data is the same supplied Type inventory. It is not a Snapshot implementation or evidence for bootstrap from an arbitrary existing graph. Alias locators and Type installation are not fabricated as graph StateChange records.

## Exact contract anchors

References below are to c0 `docs/specs/bdp.md` and normative `schemas/bdp-v0.schema.json`, not illustrative fixtures.

| Source | Required behavior relevant to G1 |
| --- | --- |
|3652–3706, Scope history and authorization fences| Stable epoch on ordinary restart, view/epoch-bound history tokens, equality-only client positions, immediately preceding group position; aliases occupy no position. |
|3738–3753, discovery field meanings| minimumReplayPosition is oldest legal exclusive cursor; this slice retains genesis and does not advertise discovery. |
|5642–5733, Change groups and replication| Exactly one atomic group per Event-inducing transaction; none for failed/no-effect; deterministic projection; final postimages/tombstones; owned source/Link agreement (mandatory negative rejection additionally at5946–5954); stable ordinal sequence; no rejection of a valid transaction solely for derived-view size. |
|5750–5824, strict reads/HTTP| Same-prefix context, canonical minimum tuple wait/refuse, epoch/view/minimum/expiry before conditional shortcut; current fields on304/412; HEAD creates no mutation and acknowledges no application. |
|5959–6129, Scope changefeed| Exactly one start intent, exclusive continuation, genesis or start=now, same-entry observed head, all arrays, whole groups, no silent replay-window clamp, durable checkpoint only after atomic application. |
|6536–6586, validators/tokens| Finite feed may omit validators; token alphabet/length, reserved genesis, positions distinct from revisions. |
|1059 and3831, limits| page.defaultItems/maximumItems count Resource records and receipt result entries. Neither clause defines group-count or Event-count semantics. |
|schema $defs.changeGroup/stateChange/changefeedPage| Authoritative required members, discriminator shapes, absolute next URL/null, eventCount/array facts additionally asserted semantically. |

Targeted rows: transactional.changefeed.group, wire-form, start-intent, owned-link-agreement, consistency-fields, catch-up-timeout, replay-window (supported retained/fence cases), owned-closure; shared http.token-profile and http.conditional-reads. This is evidence preparation, not a declaration that whole rows have passed. SSE/reconnect, projection-advance under a genuinely hidden transaction, Snapshot rendezvous and erasure obligations remain incomplete.

## Existing owner versus missing seam

Current transaction-store.ts has `groups(position INTEGER PRIMARY KEY,body TEXT)`, committed in the graph/receipt transaction; group fields include raw checkpoint=position=`pos-N`, previousPosition, final changes, ordered Events, empty erasures. `readContext()` reads persisted position plus actual epoch/view while the owner's synchronous entry is held. `withRead` is a lifetime-checked Resource facade, not a group API. `inspect().groups` materializes all history and is explicitly diagnostic: never use it for serving.

Frozen8569 at `vickie-transaction-events-finite` provides `readEvents(sourceUrl, after?, minimum?)`, reading bounded rows one at a time under `enter`, and `renderTransactionEvent(context, source, checkpoint, event)` in transaction-events.ts. Its renderer currently accepts the stored pos-N checkpoint and emits epoch/view/source-tagged IDs. These signatures were inspected at exact clean8569. Its renderer has separate fixed43-character epoch/view/source tags and uses stored pos-N as input. Its source-derived TRANSACTION_EVENT_GROUP_BYTES already implements the accepted closed-shape stored bound; reuse that bound, not a duplicate stale constant.

Reuse the Event renderer with canonical `${scope}events/` and the original stored group position/checkpoint. Scope Event page facts and the corresponding group's Events must be byte-identical, including ID, source, ordinal, time, transaction and data. Do not replace the renderer's input with a new outward checkpoint spelling on one path only. The public group checkpoint can be a distinct deterministic encoding of the same unique stored group position; that does not make raw pos-N a public replay checkpoint.

Extract only a narrow owner-internal `groupAtOrAfter(position, capturedHead)` SQL helper if it avoids repeating the Event lane's one-row query. Keep it inside transaction-store.ts; no independent log, exported iterator or callback with SQLite across await. G1 needs exact adjacency checks, not merely “some later row”: if the expected retained group is missing, fail closed as owner corruption/unavailable state, never silently skip. Reuse the owner's `enter` fence/reentrancy/shutdown discipline. No mutation, receipt or identity machinery changes.

## Checkpoint representation

Use a private deterministic checkpoint codec shared later by Snapshot. Suggested bounded encoding: `cg_<43-char epoch/scope tag>_<43-char view tag>_<positive internal position>`, where each tag is SHA256/base64url over an unambiguous tagged tuple (canonical Scope+epoch; canonical Scope+epoch+view). Fixed separators have an exact parser; decimal position is private implementation detail, safe integer checked. Total stays below256; no base64 of arbitrarily long URLs. Tokens are not credentials or proof of issuance.

`genesis` is the reserved position-zero checkpoint and is never minted for a later group. Outward position, previousPosition, observedHeadPosition and BDP-Scope-Position preserve pos-N; the genesis position is pos-0. Thus an empty genesis page has after=genesis and observedHeadPosition=pos-0. These are distinct token roles; a replica uses equality only and never parses/arithmetic-orders pos-N. The inherited minimum reader accepts genesis as its established zero alias; G1 does not remove that behavior or emit genesis as a position. Current raw `pos-N` group.checkpoint is not accepted as a client checkpoint merely because it is stored internally. A presented positive checkpoint must match the captured epoch/view tags and a real retained row whose position and predecessor agree. Do not accept a future/nonexistent index by clamping to head. At all-history-retained configuration, a valid current checkpoint is never expired by time; normal receipt maintenance does not prune it.

A pure typed intent distinguishes `{after: checkpoint}` and `{start: "now"}`. No HTTP token decoder reaches mutation admission. A future Snapshot must use this same codec for its anchor; no Snapshot handle is allocated here. Existing groups need no durable rewrite or format bump for deterministic serving. If the Event predecessor chose a different frozen context-tag codec, reuse its narrow tagged hashing utility only when this avoids divergence without generalizing a framework.

## Request execution and original start anchor

1. Apply existing canonical target/method and trusted transport authentication rules. `changes/` remains a Resource shape with GET/HEAD; unsupported methods use bodyless405/Allow, independent of database contents. Existing disabled-CORS Origin handling remains. Credentials never come from actor or a checkpoint.
2. Parse exactly one starting intent. `after=genesis` is complete retained history; after a real group is exclusive. `start=now` captures the authority head checkpoint **when admitting this request**, not on every retry.
3. For start=now, obtain an immutable private resolved anchor under the owner entry once. A later minimum wait must retain that original anchor: if new commits make the minimum attainable, they belong to this request's future suffix and must be returned. Capturing the anchor after the wait, or renewing it on each attempt, would skip those groups. No durable handle/pin is created. A direct after cursor already supplies its anchor.
4. Each finite page attempt captures current epoch/view/head once inside the same authoritative owner entry, checks the anchor/fences and supported minimum, loads at most the next whole group plus a cheap existence/boundary check, and builds a detached deeply frozen page and current context there. Query uses `position <= capturedHead`. No later inspect, second read for headers or mutable last-position variable.
5. Reuse waitForTransactionMinimum for behind only: original monotonic finite deadline, bounded5ms backoff, no SQLite entry across await, context from the actual terminal attempt for409/410/503. Use the service's tracked run/drain lifetime. Existing HTTP disconnect behavior drains bounded waits; do not claim new AbortSignal propagation or socket cancellation.
6. Serialize JSON using the existing lossless storage/wire discipline. Apply ordinary conditional/HEAD handling only after the ordinary response, fences and minimum decision. Omit ETag/Last-Modified rather than fabricate them from checkpoint. `If-None-Match:*`/`If-Match:*` still follow actual target semantics;304/412 keep current context and never represent applied progress.

The separately captured start anchor is request intent, not the page's serving context. The returned page/groups and observedHeadPosition are all from the final same-entry observation. For an immediate start=now with no intervening commit the result is empty at that anchor. For after at current head, an empty page remains a usable continuation point after later commits.

## Finite response, navigation and materialization

Page fields are exactly scope, scopeEpoch, authorizationView, after, observedHeadPosition, groups, next. after preserves the requested checkpoint, or the captured start=now head checkpoint. observedHeadPosition and BDP-Scope-Position report the current captured prefix, which may be later than the last group returned; clients must advance only to delivered/applied group.checkpoint, never to observedHeadPosition.

Return at most one complete group. This simplifies a real finite vertical slice without importing Resource/receipt page limits into group semantics. next is absolute canonical `${scope}changes/?after=<last-returned-checkpoint>` when another group exists within the captured head; otherwise null. The next request is a fresh strict suffix read, not a retained Resource-query snapshot. It can observe newly appended groups. When next is null a client can later resume with its last durable checkpoint (or page.after if none applied). Never use an Event ID or HTTP conditional result as that cursor.

Serve stored final StateChanges unchanged apart from validation/deep freeze; preserve retained pre-metadata representation where present. Validate scope/epoch/view/position/previousPosition, projectionAdvance===false, transaction, at least one nonempty changes/erasures/events array, eventCount and exact zero-based contiguous ordinals0..eventCount-1 in this all-visible configuration; reject inconsistent persisted data rather than repairing it on read. changes order is nonsemantic, but preserve existing stable serialization; repeated updates/reversal yield all original Events and only the final StateChange. A create/delete in one group must preserve the recorded final tombstone and intermediate facts. Empty erasures is the actual supported administrative configuration, not a reason to remove the array.

Whole-group materialization depends on the reviewed installed count/property/diagnostic tuple. Derive the served bound from its maximum raw committed-group representation **plus** Event rewriting overhead: canonical Scope Event source's actual JSON-escaped length, each bounded256-char ID, outward checkpoint expansion, and page envelope's actual bounded tokens/Scope URL. Do not apply the Event page byte cap or Resource representationBytes to a group. Count caps are not byte caps. Construct/measure an actual boundary group and prove it is returned whole; there must be no path silently slicing it or returning next past an undelivered group. The pinned calculation below establishes that premise for076/8569. It is not a new protocol byte limit or a justification to retroactively refuse committed transactions.

## Error decisions: laws versus reference conventions

| Input/condition | G1 handling and authority |
| --- | --- |
|Unauthenticated actual GET/HEAD|Existing401/challenge; no history context or row observation. Token possession does not authorize.|
|Both/none starting intents; duplicate/unsupported query; empty after; start other than now; malformed token|Reject. Exactly-one/error is normative5959ff, but the exact generic feed parameter mapping is not separately specified. Root permits the documented reference convention400 invalid-parameter, matching the Event lane; do not call the collection-specific1550 clause a universal feed law.|
|Well-formed current-context but never-issued/nonexistent/future checkpoint; Event ID supplied as group cursor|Reference convention400 invalid-parameter, no silent clamp. This does not prove expiry, and the documented convention must remain distinct from mandatory retained-history refusal.|
|Recognized checkpoint context from another epoch/Scope or Authorization View|Explicit fence before success/conditional;410 cursor-expired for epoch/unavailable context and409 foreign-view for view, following existing consistency conventions/laws. Tags are diagnostic, not issuance proof; a forged foreign tag can receive the same nondisclosing context refusal.|
|Actual retained gap/corrupt group|Fail closed private owner/service error; do not mislabel internal corruption as client400 or rewrite the log. No pruning path is introduced. A future legitimate pruned-window case must implement410/rebootstrap and minimumReplayPosition honestly.|
|Complete supported minimum tuple behind|Wait original finite budget then503 catch-up-timeout with actual last context; already-satisfied succeeds. Known wrong epoch/view410/409. Schema-validate bodies.|
|Partial/malformed/unsupported minimum tuple|Existing explicit private bodyless development gap after authentication; this plan does not decide the still-unruled broad parser semantics.|
|Only text/event-stream acceptable|Private finite slice406, no SSE claim. When JSON is allowed choose it; acceptable SSE HEAD is not advertised before corresponding GET. SSE Last-Event-ID precedence remains future work.|

For Last-Event-ID on finite JSON GET/HEAD, use ordinary unknown-header handling: ignore it completely. It neither supplies an omitted start intent nor overrides an explicit one; malformed or foreign header bytes do not become cursor failures on this finite path. Explicit intent remains mandatory. Frozen8569 routes observations through applyReadHttpSemantics, which returns406 for JSON when Accept excludes JSON and does not interpret Last-Event-ID. Reuse that path; SSE support and its header precedence remain expressly incomplete. Unknown query mapping above is a documented reference convention approved for planning, not a normative spec change. No parameter ambiguity blocks the settled valid-intent owner core.

## File ownership and implementation sequence

1. New private `transaction-changefeed.ts`: typed intent/checkpoint codec, page/group projection using Event renderer, integrity assertions and pure tests. Scope-local URL parsing, no new public protocol parser/export or second renderer. Add private build-output/deep-import smoke entries only.
2. `transaction-store.ts`: narrow group-row helper reuse; one-time start-head capture and same-entry finite page observation. Preserve current entry fence, immutable returned data, complete-group existence validation. No callback injection, format, compiler, evaluator, admission, receipt-expiry or retention change.
3. `transaction-service.ts`: authenticated finite read facade with start anchor fixed before wait, shared minimum wait, tracked lifecycle; no second scheduler/owner.
4. `transaction-http.ts`: route canonical changes before generic Read/minimum route guard, existing request auth/negotiation/error/context/conditional helpers. `transaction-runtime.ts` only if fixture navigation needs the real private changes URL; no listener architecture or installed profile factory change.
5. Focused `transaction-changefeed.test.ts`, store/service/actual HTTP integration tests as necessary; one small executable replica-application oracle in test support. `docs/design/transaction-implementation-progress.md` and private smoke allowlist. No catalog/evidence relabel or normative edits.

Wait for Event/materialization freezes, ordinary-integrate their exact heads and inspect shared-file overlap before any implementation. If a helper extraction would alter the Event policy or shared Read/runtime surface, return to root for ownership scoping; do not broaden silently.

## Executable done-when

1. Actual finite GET from genesis returns all committed graph history one whole group per page; JSON validates as changefeedPage/changeGroup. No effect from failed/no-op/alias-only mutation appears as a new group. Absolute next is followed as supplied.
2. A test replica starts empty, applies each complete group's final postimages/tombstones into a staged map and atomically publishes graph+checkpoint. Exact canonical Resource state/revisions/metadata/provenance and owned inline/first-class equality match authority reads at the same captured final position. Compare state, not Event count.
3. Inject client failure before atomic publication: graph+durable checkpoint remain unchanged; reconnect from old checkpoint replays whole group, applies once. Duplicate delivery after publication is detected by checkpoint; wrong previousPosition, reorder/gap, wrong epoch/view and conflicting duplicate refuse without partial state. A source inline ownedLinks record disagreeing member-for-member with the corresponding Link upsert or tombstone must also refuse as an authority fault with neither partial graph publication nor checkpoint advance (5946–5954). This is a test oracle, not a new production client library.
4. Singleton, multi-operation batch, repeated update/reversal, create-update-delete, deletion/tombstones, self-Link incidence and owned transitions all preserve ordered facts versus normalized final state. Relevant Scope Event pages and group.events are byte-identical; no source preimage body is invented.
5. Head may be later than the delivered group. Client advances only to applied group checkpoint; next/after never skips undispatched groups. Empty caught-up page, genesis-empty, after-current and subsequent append work. start=now skips old groups but retains its original anchor across minimum wait and includes intervening commits.
6. Long configured canonical source URL, maximum safe supported ordinal/position spelling, token case and restart determinism exercise codec bounds. Event cursor/raw pos-N/never-issued group token cannot masquerade as a delivered group checkpoint. Real old-epoch and distinct-view test contexts trigger explicit fences without minting a new policy configuration in production.
7. Current page body/observedHeadPosition/headers are from one owner observation. Foreign/expired/behind/timeout wins before304;401 has no context;409/410/503 use actual final attempt context. GET/HEAD parity, no frames/no mutation onHEAD,304/412 no checkpoint advancement.
8. Wrong/missing/repeated intents and documented reference conventions have actual HTTP regressions; supported minimum success/wait/timeout/error responses validate against transactionalProblem. Unsupported minimum controls remain explicit gaps; Last-Event-ID is ignored on finite JSON, tested both with a valid explicit intent and without one.
9. Ordinary restart preserves epoch/checkpoint/Event IDs and page suffix. Existing actual hot-journal recovery tests remain green; no read-only preflight reintroduced. After owner fencing or close no new feed can serve; service drain preserves tracked finite work, no leaked iterator or live reader escapes.
10. Boundary committed group returns whole under the accepted materialization tuple including public Event source/ID expansion. Serving loads at most the chosen group plus bounded row metadata, not all groups. Deep freeze prevents consumer mutation. Tampered position/previousPosition/eventCount/ordinal fails closed, never repaired or silently skipped.
11. Build before process tests; pinned Node24.16/pnpm11.20; focused meaningful suites(maxWorkers2), typecheck/boundaries/format/lint, both installed smokes and unchanged historical evidence verifier. Record exact head/logs/counts and all untested branches; no historical evidence reseal.
12. Progress report still lists SSE, Snapshot/bootstrap, hidden projection/view rotation, erasure, remaining broad minimum forms, full cold discovery (including changes navigation, headPosition and minimumReplayPosition) and installed cumulative TX qualification as missing. G1 never changes the public profile/export map or claims those obligations from finite-feed tests.

## Council questions that remain bounded

- The400 invalid-parameter mapping is settled as a supported local reference convention, with alias/version/query-table analogies as support rather than a universal new rule. No human ruling is required. Finite JSON ignores Last-Event-ID; SSE precedence is out of slice.
- Confirm start=now one-time capture before a minimum wait; this follows admission-time anchor wording and needs a discriminating test, not a second durable cursor system.
- Exact8569 signatures and076 source-derived tuple are now pinned. Any review repairs still require ordinary integration. No product decision is needed for this seam.
- The existing no-pruning/all-visible fixture cannot qualify actual expiry, erasure or hidden projection paths. Their absence stays in the burn-down; do not synthesize support or application evidence from standalone JSON examples.


## Council dispositions and source-derived serving bound

The pre-council memo is preserved as transaction-changefeed-plan-pre-council.md.
H1 correctly demonstrates that old c0 capacity could not establish this premise;
076 was mandatory precisely to replace that tuple. At frozen8569, actual installed
limits are M=64 cumulative matches, E=256 Events, O=32 operations, R=262144 complete
Resource bytes, Q=65536 request bytes and S=65536 encoded Scope Event-source bytes.
These enforced count and representation premises—not an asserted new byte cap—give
N=7Q, A=8192, H=1048576, K=M+O and the already-reviewed stored bound:

```text
G = min(E*(2.5R+N+S)+K*A+H, K*(max(9R+4S,5R+2N+2S)+A)+H)
  = 253,493,248 compact JSON UTF-8 bytes
```

Use the frozen Event module's source-derived G. Add an intentionally conservative
allowance even where stored source/id already paid those bytes: every rewritten
Event adds at most S+258+32 bytes (encoded Scope source, encoded256-char ID and
structural slack). The one-group page plus headers adds2S+12*258+4096: two bounded
Scope/absolute-next spellings, checkpoint/context tokens and fixed HTTP/envelope
fields. Scope is shorter than its events/ URL; percent encoding is not introduced
by ASCII checkpoint tokens. No caller-controlled arbitrary header is reflected.
The stored group already includes positions and original facts; this ADDITION
therefore overcounts rather than relying on subtracting their old spelling.

```text
rewritten Event allowance = 256*(65536+258+32) = 16,851,456
page/header allowance = 2*65536+12*258+4096 = 138,264
served one-group bound = 270,482,968 bytes
pinned Node24.16.0 MAX_STRING_LENGTH = 536,870,888 UTF-16 code units
headroom = 266,387,920
```

Well-formed compact JSON UTF-8 bytes upper-bound its UTF-16 code-unit count. The
proof was executed by transaction-changefeed-bound.mjs against exact clean8569;
transaction-changefeed-bound.log records the source tuple and actual Node24
constant. It allocates no large Resource/group/string. This does not prove total
heap/RSS safety or concurrent multi-request capacity; cloned parsed graphs,
serialization temporaries and long histories remain distinct costs. The owner
still selects one whole group, not an all-history array. The bounded test plan
uses a real256-small-Event boundary and representative large admissible Resource
inputs, not a near512MiB synthetic proof allocation.

M1 adds negative owned inline/first-class disagreement (including tombstone) to
the replica oracle. M2 ignores Last-Event-ID only on finite JSON, without supplying
or overriding required intent. L1 states pos-N versus genesis and equality-only
consumer logic; L2 directly validates non-advance/nonempty; L3 validates exact
all-visible ordinal contiguity, with any future gap relaxation owned by hidden
projection work; L4 makes next explicitly absolute; L5 removes an unnecessary
human decision on the existing local400 convention. Current discovery membership,
headPosition/minimumReplayPosition, Snapshot, SSE, actual replay pruning, hidden
projection and erasure remain incomplete—none is certified by this finite slice.
