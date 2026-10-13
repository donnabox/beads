# G1: finite whole-group Transactional changefeed

Read-only implementation proposal, 2026-10-13. Exact inspected integration is **c0d67388cd849410b38ec7ecb3a328771431dbfa**, which already contains owned candidate01e2. This supplements the sole fork design's T2/T3 completion plan and `transaction-required-contract-burndown.md`; it is not a second authority or qualification report. No runtime, schema, catalog, evidence or frozen checkout is changed by this memo. No new tests were run for planning.

## Bounded outcome and prerequisites

Implement actual finite JSON GET/HEAD of canonical `changes/` over the existing authoritative `groups` table, with whole groups, explicit starting intent, fenced checkpoints, current same-entry context, canonical minimum controls and native conditional semantics. Keep the actual fixed all-visible policy, retained complete history and empty erasure ledger. One complete group per page is an implementation bound, not a new advertised protocol limit. Do not implement SSE, Snapshot creation/streams, policy/view mutation, erasure administration, optional History, set HTTP admission or discovery/profile advertisement here.

Predecessors: ordinary-integrate the independently reviewed finite Event lane and accepted materialization/installed-configuration lane before authoring shared files. Owned transitions are already integrated at c0. Event lane currently writes transaction-store/service/http and a new transaction-events module; do not edit those concurrently. The materialization lane establishes the admissible committed-group premise; do not impose retrospective transaction or group refusal under a new response budget.

The replication oracle here starts from genesis on an empty graph, whose only administrative data is the same supplied Type inventory. It is not a Snapshot implementation or evidence for bootstrap from an arbitrary existing graph. Alias locators and Type installation are not fabricated as graph StateChange records.

## Exact contract anchors

References below are to c0 `docs/specs/bdp.md` and normative `schemas/bdp-v0.schema.json`, not illustrative fixtures.

| Source | Required behavior relevant to G1 |
| --- | --- |
|3652–3706, Scope history and authorization fences| Stable epoch on ordinary restart, view/epoch-bound history tokens, equality-only client positions, immediately preceding group position; aliases occupy no position. |
|3738–3753, discovery field meanings| minimumReplayPosition is oldest legal exclusive cursor; this slice retains genesis and does not advertise discovery. |
|5642–5733, Change groups and replication| Exactly one atomic group per Event-inducing transaction; none for failed/no-effect; deterministic projection; final postimages/tombstones; owned source/Link agreement; stable ordinal sequence; no rejection of a valid transaction solely for derived-view size. |
|5750–5824, strict reads/HTTP| Same-prefix context, canonical minimum tuple wait/refuse, epoch/view/minimum/expiry before conditional shortcut; current fields on304/412; HEAD creates no mutation and acknowledges no application. |
|5959–6129, Scope changefeed| Exactly one start intent, exclusive continuation, genesis or start=now, same-entry observed head, all arrays, whole groups, no silent replay-window clamp, durable checkpoint only after atomic application. |
|6536–6586, validators/tokens| Finite feed may omit validators; token alphabet/length, reserved genesis, positions distinct from revisions. |
|1059 and3831, limits| page.defaultItems/maximumItems count Resource records and receipt result entries. Neither clause defines group-count or Event-count semantics. |
|schema $defs.changeGroup/stateChange/changefeedPage| Authoritative required members, discriminator shapes, absolute next URL/null, eventCount/array facts additionally asserted semantically. |

Targeted rows: transactional.changefeed.group, wire-form, start-intent, owned-link-agreement, consistency-fields, catch-up-timeout, replay-window (supported retained/fence cases), owned-closure; shared http.token-profile and http.conditional-reads. This is evidence preparation, not a declaration that whole rows have passed. SSE/reconnect, projection-advance under a genuinely hidden transaction, Snapshot rendezvous and erasure obligations remain incomplete.

## Existing owner versus missing seam

Current transaction-store.ts has `groups(position INTEGER PRIMARY KEY,body TEXT)`, committed in the graph/receipt transaction; group fields include raw checkpoint=position=`pos-N`, previousPosition, final changes, ordered Events, empty erasures. `readContext()` reads persisted position plus actual epoch/view while the owner's synchronous entry is held. `withRead` is a lifetime-checked Resource facade, not a group API. `inspect().groups` materializes all history and is explicitly diagnostic: never use it for serving.

The active Event lane at `vickie-transaction-events-finite` presently adds `readEvents(sourceUrl, after?, minimum?)`, reading bounded rows one at a time under `enter`, and `renderTransactionEvent(context, source, checkpoint, event)` in transaction-events.ts. Its renderer currently accepts the stored pos-N checkpoint and emits epoch/view/source-tagged IDs. These are provisional WIP signatures, not a frozen predecessor claim. Root relayed the author's shared-rendering commitment; finalize signatures only after that lane freezes.

Reuse the Event renderer with canonical `${scope}events/` and the original stored group position/checkpoint. Scope Event page facts and the corresponding group's Events must be byte-identical, including ID, source, ordinal, time, transaction and data. Do not replace the renderer's input with a new outward checkpoint spelling on one path only. The public group checkpoint can be a distinct deterministic encoding of the same unique stored group position; that does not make raw pos-N a public replay checkpoint.

Extract only a narrow owner-internal `groupAtOrAfter(position, capturedHead)` SQL helper if it avoids repeating the Event lane's one-row query. Keep it inside transaction-store.ts; no independent log, exported iterator or callback with SQLite across await. G1 needs exact adjacency checks, not merely “some later row”: if the expected retained group is missing, fail closed as owner corruption/unavailable state, never silently skip. Reuse the owner's `enter` fence/reentrancy/shutdown discipline. No mutation, receipt or identity machinery changes.

## Checkpoint representation

Use a private deterministic checkpoint codec shared later by Snapshot. Suggested bounded encoding: `cg_<43-char epoch/scope tag>_<43-char view tag>_<positive internal position>`, where each tag is SHA256/base64url over an unambiguous tagged tuple (canonical Scope+epoch; canonical Scope+epoch+view). Fixed separators have an exact parser; decimal position is private implementation detail, safe integer checked. Total stays below256; no base64 of arbitrarily long URLs. Tokens are not credentials or proof of issuance.

`genesis` is the reserved position-zero checkpoint and is never minted for a later group. Current raw `pos-N` group.checkpoint is not accepted as a client checkpoint merely because it is stored internally. A presented positive checkpoint must match the captured epoch/view tags and a real retained row whose position and predecessor agree. Do not accept a future/nonexistent index by clamping to head. At all-history-retained configuration, a valid current checkpoint is never expired by time; normal receipt maintenance does not prune it.

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

Return at most one complete group. This simplifies a real finite vertical slice without importing Resource/receipt page limits into group semantics. next is canonical `changes/?after=<last-returned-checkpoint>` when another group exists within the captured head; otherwise null. The next request is a fresh strict suffix read, not a retained Resource-query snapshot. It can observe newly appended groups. When next is null a client can later resume with its last durable checkpoint (or page.after if none applied). Never use an Event ID or HTTP conditional result as that cursor.

Serve stored final StateChanges unchanged apart from validation/deep freeze; preserve retained pre-metadata representation where present. Validate scope/epoch/view/position/previousPosition, transaction, eventCount and ordinal ordering; reject inconsistent persisted data rather than repairing it on read. changes order is nonsemantic, but preserve existing stable serialization; repeated updates/reversal yield all original Events and only the final StateChange. A create/delete in one group must preserve the recorded final tombstone and intermediate facts. Empty erasures is the actual supported administrative configuration, not a reason to remove the array.

Whole-group materialization depends on the reviewed installed count/property/diagnostic tuple. Derive the served bound from its maximum raw committed-group representation **plus** Event rewriting overhead: canonical Scope Event source's actual JSON-escaped length, each bounded256-char ID, outward checkpoint expansion, and page envelope's actual bounded tokens/Scope URL. Do not apply the Event page byte cap or Resource representationBytes to a group. Count caps are not byte caps. Construct/measure an actual boundary group and prove it is returned whole; there must be no path silently slicing it or returning next past an undelivered group. If the predecessor tuple cannot bound the serving envelope, report that before implementation readiness; do not solve it by retroactively refusing committed transactions or adding an unadvertised truncation policy.

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

For Last-Event-ID on a finite JSON request, no SSE precedence is inferred. Recommended bounded initial handling is an explicit private unsupported-control gap, never silently override after or create a second start intent. Exact JSON semantics are not stated by the SSE reconnect clause; list as a remaining parameter-specific question for council. Unknown query mapping above is a documented reference convention approved for planning, not a normative spec change. No parameter ambiguity blocks the settled valid-intent owner core.

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
3. Inject client failure before atomic publication: graph+durable checkpoint remain unchanged; reconnect from old checkpoint replays whole group, applies once. Duplicate delivery after publication is detected by checkpoint; wrong previousPosition, reorder/gap, wrong epoch/view and conflicting duplicate refuse without partial state. This is a test oracle, not a new production client library.
4. Singleton, multi-operation batch, repeated update/reversal, create-update-delete, deletion/tombstones, self-Link incidence and owned transitions all preserve ordered facts versus normalized final state. Relevant Scope Event pages and group.events are byte-identical; no source preimage body is invented.
5. Head may be later than the delivered group. Client advances only to applied group checkpoint; next/after never skips undispatched groups. Empty caught-up page, genesis-empty, after-current and subsequent append work. start=now skips old groups but retains its original anchor across minimum wait and includes intervening commits.
6. Long configured canonical source URL, maximum safe supported ordinal/position spelling, token case and restart determinism exercise codec bounds. Event cursor/raw pos-N/never-issued group token cannot masquerade as a delivered group checkpoint. Real old-epoch and distinct-view test contexts trigger explicit fences without minting a new policy configuration in production.
7. Current page body/observedHeadPosition/headers are from one owner observation. Foreign/expired/behind/timeout wins before304;401 has no context;409/410/503 use actual final attempt context. GET/HEAD parity, no frames/no mutation onHEAD,304/412 no checkpoint advancement.
8. Wrong/missing/repeated intents and documented reference conventions have actual HTTP regressions; supported minimum success/wait/timeout/error responses validate against transactionalProblem. Unsupported minimum/Last-Event-ID controls remain explicit gaps rather than ignored input.
9. Ordinary restart preserves epoch/checkpoint/Event IDs and page suffix. Existing actual hot-journal recovery tests remain green; no read-only preflight reintroduced. After owner fencing or close no new feed can serve; service drain preserves tracked finite work, no leaked iterator or live reader escapes.
10. Boundary committed group returns whole under the accepted materialization tuple including public Event source/ID expansion. Serving loads at most the chosen group plus bounded row metadata, not all groups. Deep freeze prevents consumer mutation. Tampered position/previousPosition/eventCount/ordinal fails closed, never repaired or silently skipped.
11. Build before process tests; pinned Node24.16/pnpm11.20; focused meaningful suites(maxWorkers2), typecheck/boundaries/format/lint, both installed smokes and unchanged historical evidence verifier. Record exact head/logs/counts and all untested branches; no historical evidence reseal.
12. Progress report still lists SSE, Snapshot/bootstrap, hidden projection/view rotation, erasure, remaining broad minimum forms, full cold discovery and installed cumulative TX qualification as missing. G1 never changes the public profile/export map or claims those obligations from finite-feed tests.

## Council questions that remain bounded

- Confirm the documented feed-specific400 reference conventions and finite-JSON Last-Event-ID unsupported boundary; neither requires rewriting the protocol to implement valid inputs.
- Confirm start=now one-time capture before a minimum wait; this follows admission-time anchor wording and needs a discriminating test, not a second durable cursor system.
- Confirm exact frozen Event renderer/helper signatures and materialization formula on the integration base. These are active lane dependencies, not product decisions.
- The existing no-pruning/all-visible fixture cannot qualify actual expiry, erasure or hidden projection paths. Their absence stays in the burn-down; do not synthesize support or application evidence from standalone JSON examples.
