# Initial Scope Event SSE opaque minimum — bounded plan

PLAN ONLY. Basis: clean `bbfa5b2cac0e381b497384e3a8e29cd937ea2738`, tree `da812ea5a1da8f8fb5be9fd2856d312c316236a3`, direct parent `5ceb213b632e17f28714684efc5e35a62b282bf6`. PR161 exact CI39044573943 remains pending at the root preparation handoff: no authoring base or execution is released. Root must independently verify exact green basis before authoring. Native plan review, actual Claude plan review and root findings disposition precede source authoring. This packet continues the sole design plan, not another tracker. All counts below are source-derived successful-path workload; none is an observed result or completion percentage.

## Selected behavior and normative boundary

A negotiated SSE read of the canonical Scope `events/` Source, with no URL `after` key and no Last-Event-ID occurrence, uses the existing finite minimum wait for a lexical unknown minimum in the actual current epoch/view. If eligibility stays unproved it returns contextual503 catch-up-timeout, application/problem+json and Retry-After:1 before successful SSE headers, packet construction or stream-slot admission. HEAD returns the corresponding metadata and no body. No token-derived freshness or position ordering is introduced; `unknown` and `pos-01` remain opaque identities.

Pinned monolithic specification `docs/specs/bdp.md` SHA67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8:5820–5832 lexical receiving/HEAD,5841–5851 remaining-read eligibility/wait/route/problem law,5874–5884 minimum before conditional and HEAD no-stream,6257–6263 cursor-free Event start,6325–6356 SSE same Source and Last-Event-ID precedence. This is a bounded reference implementation choice under the general remaining-read law. The narrower current Bead/Link Resource MUST does not explicitly name Scope SSE. Cursor-free start at the earliest retained Event does not invent retained-continuation precedence.

Keep defaults/direct calls strict unless the explicit new private intent is supplied and structurally eligible. Keep URL after (even empty/malformed), Last-Event-ID (even empty/repeated raw fields), reconnect-invalid, Bead/Link Sources and foreign-context opaque minima outside this extension. Known minimum, source identity/missing handling, cursor errors, projection/corruption, erasure, pump, conditional handling, capacity and lifecycle retain their existing order. Foreign opaque remains an unsupported private gap, not a new normative409/410 decision. S1/G2 and public discovery/profile admission remain held; no Types, History, retention redesign, parser change, main work or Update rewrite.

## Exactly six owned files, three production seams

1. `packages/server/src/transaction-http.ts` (Event branch509–559). Retain early auth/lexical receiving and transactionEventRequest parsing before owner admission. Compute selected media and raw reconnect fields before selecting the minimum helper. The new predicate is negotiated `media === "sse"`, canonical Scope Source, `eventRequest.after === undefined`, no `url.searchParams.has("after")`, no `request.headers.has("last-event-id")`, and zero raw reconnect fields. Empty/repeated raw values are occurrences, not absence. Reuse the existing finite JSON predicate unchanged. With hasMinimumControls, use currentReadMinimumHeaders for the new selected path or the existing finite selection; otherwise use the existing strict helper. Preserve undefined/no-controls versus unsupported sentinel. Pass a final optional intent `"initial-scope-sse"` to beginEvents only when selected. Raw reconnect array remains authoritative for repeated-SSE and header precedence; ordinary unsupported Accept (`selectChangefeedMedia` may return undefined) remains excluded from this new SSE predicate and follows existing finite/negotiation behavior. No minimum/conditional/capacity order changes outside selection.
2. `packages/server/src/transaction-service.ts` (beginEvents540–557). Append sixth argument `admission?: "initial-scope-sse"`; existing five-argument calls stay strict. Eligibility additionally requires exact canonical `new URL("events/", scope).href`, after undefined and invalidReconnect false. Capture selected tuple with snapshotCurrentReadMinimum, otherwise snapshotTransactionMinimum; preserve undefined/sentinel and synchronous capture BEFORE run/check/waiter. Capture once, not inside attempts. Forward explicit intent only for eligible calls to owner; each retry invokes the same owner method and recaptures current context. Do not copy finite events' distinct first-attempt capture ordering. Existing waiter, signal, deadline and close ownership remain unchanged. No new shared helper or public option.
3. `packages/server/src/transaction-store.ts` (beginEventScan1766–1810). Append fifth optional literal intent. Preserve initial control validation and eventSource parsing before minimum capture. Select only matching intent, actual parsed canonical Scope Source, after undefined and invalidReconnect false. Lexically snapshot only selected minima; default/excluded calls stay strict. Inside enter preserve missing Source before unsupported sentinel, then current context, then selected checkCurrentReadMinimum(minimum,context,true) or original checkTransactionMinimum, then invalidReconnect/cursor validation and finally eventPacket. No scan/packet is constructed on unproved failure. Do not change advanceEventScan, cursor membership, group reader, source projection, packet ownership or stream pump. An invalid/nonmatching optional intent falls through the strict path rather than granting broader behavior.
4. `packages/server/src/transaction-event-sse.test.ts`: append A–D below. Preserve the original entire file prefix, old callbacks and helpers. Local imports needed by new cases may be dynamic inside appended callbacks/helper code. A may load test modules asynchronously, then perform its owner checks synchronously. Add only a local helper for the new real-owner service fixtures if useful; do not rewrite existing fixtures.
5. `apps/bdptest/src/transactional-event-sse-installed.test.ts`: preserve the entire original file prefix and both historical callbacks; append new E. Reuse existing module configuration/artifacts/launcher/binding and implement bounded local ownership for E, without changing shared helpers. Default development fallback must still execute normally; paired external overrides require an actual outside-worktree realpath and separately verified provenance.
6. `docs/design/transaction-implementation-progress.md`: append one short paragraph describing this fresh Scope SSE subset as a draft until validation, explicitly superseding only its prior unsupported-minimum limitation. Preserve historical prose/evidence and remaining reconnect/Resource/foreign/S1/G2/public limits. No retroactive historical evidence claims.

New production paths use the existing consistency imports and private types; no seventh file. Original file bytes/hashes for both test prefixes and the documentation prefix are pinned in preservation.json. Original source snapshots accompany this packet.

## New source callbacks and exact workload

All new source callbacks are top-level `it` in transaction-event-sse.test.ts. Four cases, four fresh stores; no parameterized hidden expansion. Existing default5s callback bounds remain. Held-latch cases use a1000ms local entered-versus-settlement guard to trigger failure cleanup, not a product timing assertion; test clock below is independent and deterministic. No owner entry survives an await.

### A — gates initial Scope Event SSE opaque minima before scan packets

Use existing ownerFixture, seed one real batch creating a,b and one Link l from a to b (one commit, three operations). Wrap real beginEventScan with a counter. Fourteen calls in exact order:

|#|Source/controls|Expected|
|--|--|--|
|1|Scope, explicit intent, no minimum|packet; capture pos1 context and valid first Event ID|
|2|Scope selected, matching unknown|TransactionMinimumUnproved with actual pos1 context|
|3|Scope selected, matching pos-01|same unproved/context|
|4|Scope selected, known pos2|TransactionMinimumFailure behind/context|
|5|Scope selected, unknown + foreign epoch|TransactionMinimumUnsupported|
|6|Scope selected, unknown + foreign view|TransactionMinimumUnsupported|
|7|Scope default/no intent, unknown|TypeError unsupported minimum tuple|
|8|Scope explicit intent but after=valid ID, unknown|same strict TypeError|
|9|Scope explicit intent but invalidReconnect=true, unknown|same strict TypeError|
|10|existing Bead Source with intent, unknown|same strict TypeError|
|11|existing Link Source with intent, unknown|same strict TypeError|
|12|missing Bead Source, explicit unsupported sentinel|missing with actual context, preserving missing-before-sentinel|
|13|Scope selected, malformed empty minimum field|TypeError before owner admission|
|14|Scope selected, known current pos1|packet; complete same Event data as call1|

After call1 install a callback-local DatabaseSync.prepare observer for group queries (`FROM groups`), restoring it in finally. Calls2–13 must not construct a group reader; call14 must construct one. Record per-call actual outcome/group-query count and contexts where provided. Do not assert an exact implementation-dependent positive query count, or invent a WeakMap scan-allocation census. Throwing before eventPacket and zero group queries provide distinct source/test support for the no-packet-before-eligibility assertion. Actual begin calls14, waits0, HTTP0, seed commits1.

### B — captures initial Scope Event SSE minima once and recaptures context on retry

Seed a (one commit) before handing the real owner to createTransactionService. Use a callback-local spread delegate preserving scope, close, readConfiguration and all other methods; wrap only beginEventScan to record actual inputs/outcomes/context. No spying on a frozen owner. Local Read options must match ownerFixture's receiptMaximumItems1 (default/maximum page items1); use existing validated option shapes. Budget2, controlled clock initially0, counted now() and a single held wait. Minimum tuple context comes from the real seed receipt, not a synthetic read observation.

First an open-service malformed selected call must throw synchronously before any owner attempt/clock. Start selected unknown with a caller-owned mutable tuple and supplied AbortController. Attach fulfillment/rejection observer immediately, then race wait-entered against premature settlement and the local guard. At the held wait mutate all caller tuple fields; submit a second real batch creating b through service.submit (not direct owner access after lifetime transfer), prove completed status, set clock2 and release. The second owner attempt sees the original captured tuple and current pos2 context; final catch-up-timeout carries pos2. A known current selected call then succeeds with a packet; original/second attempt inputs remain identical and no packet result occurred during the unproved attempts. Close service. A malformed selected post-close call still throws TypeError synchronously before lifecycle check; a well-formed known selected post-close call throws closing/closed synchronously without owner entry.

Exactly five service.beginEvents invocations (open malformed, unknown, positive known, closed malformed, closed known), three delegated owner attempts, one wait, four minimum-clock calls (3 timeout +1 positive), two commits (seed and one service submission), one store. Additional submit calls are not beginEvents counts. Record actual attempts and current contexts rather than deriving them from the final503 alone. Every failure path must unconditionally abort the supplied controller BEFORE release of the held latch, then observe/drain the operation and memoized service close, aggregating primary/cleanup errors. On normal success abort in finally is cleanup after settlement, not the cause of the timeout.

### C — drains an aborted initial Scope Event SSE minimum wait before owner close

New seeded a store and separate service/delegate, clock0, budget2 and held wait. Observe the selected unknown operation immediately; race entered against premature settlement/guard. Start service.close while wait is held and observe that promise immediately. A synchronous `closed === false` assertion merely records that instant; it is not an independent proof of deferred closure. Abort the supplied signal, then release. Await the aborted operation and successful service/real-owner close; count actual delegate close once. Post-close valid selected service call must fail before owner attempt. Separately make one raw owner no-minimum probe and require closed-owner refusal (not counted as delegated attempt).

Two service.beginEvents invocations; one delegated beginEventScan attempt plus one explicitly separate raw post-close probe; one wait; two minimum-clock calls; one seed commit/store. Every finally path aborts before releasing and drains both observed promises, preserving failures. No sleep-based assertion or extra product timeout is added.

### D — serves initial Scope Event SSE opaque minima before conditional admission

Use existing real emitted runtimeFixture (25ms minimum budget, shared stream capacity1). One actual batch creates a,b,l, then a single finite Scope Event GET returns all three Events and next:null. Use callback-local request wrappers/counts, not edits to shared helpers. Finite setup must use a single request rather than an unbounded pagination loop. Derive context from real finite response, cursor from first Event, and validate expected subjects/ordinals/data. All subsequent rows are raw SSE requests via existing recorder except setup; maximum/final count29. Literal table is authoritative:

|#|Method/target|Controls and expected status|
|--|--|--|
|1|POST operations/batch|seed;200 completed|
|2|GET events/ JSON|finite baseline;200,3 Events,next:null|
|3|HEAD events/ SSE|known current minimum;200,no body|
|4–5|GET,HEAD events/ SSE|matching unknown;503/503|
|6–7|GET,HEAD events/ SSE|matching pos-01;503/503|
|8–9|GET,HEAD events/ SSE|unknown + If-None-Match:*;503/503,not304|
|10–11|GET,HEAD events/ SSE|unknown + If-Match:"wrong";503/503,not412|
|12–13|GET,HEAD events/ SSE|incomplete minimum tuple;400/400,no context|
|14–15|GET,HEAD events/ SSE|known future pos2;503/503|
|16|GET events/ SSE|unknown,foreign epoch;unchanged private500|
|17|GET events/ SSE|unknown,foreign view;unchanged private500|
|18|GET events/?after=validID SSE|unknown;unchanged private500|
|19|GET events/?after= SSE|unknown;parser400,no context before owner|
|20|GET events/ SSE|unknown + Last-Event-ID validID;private500|
|21|GET events/ SSE|unknown + Last-Event-ID empty;private500 (minimum sentinel precedes cursor check)|
|22|GET events/ SSE|unknown + two raw valid Last-Event-ID fields;private500,not newly waitable|
|23|GET beads/a?view=events SSE|unknown;private500|
|24|GET links/l?view=events SSE|unknown;private500|
|25|GET beads/missing?view=events SSE|unknown;404 actual context before sentinel|
|26|GET events/ SSE unauthenticated|unknown;401,no context|
|27|HEAD events/ SSE|known current + If-None-Match:*;304,no body|
|28|HEAD events/ SSE|known current + If-Match:"wrong";412,no body|
|29|GET events/ SSE|known current;200 live exact baseline frames,then deliberate close|

Status totals:200x4,503x10,500x8,400x3,404x1,401x1,304x1,412x1. All ten503s require exact current context, code catch-up-timeout, retry after-delay, literal Retry-After1 and application/problem+json. All paired HEADs have zero body and matching GET representation Content-Length/context/problem metadata. No timeout timing assertion or retry-count estimate. Selected unknown cannot succeed while context remains matching and eligibility unproved; abort/context change follow their existing paths.

For the positive row29: await ready; require status200, SSE media, no-store/no-transform, actual full context, no ETag/content-length, then await exactly the three complete id frames; compare id/type and decoded full Event to finite baseline. Immediately before intentional close, require signal not aborted, recorder not destroyed and elapsed below its5000ms bound. Close/drain then assert exact complete frames still equal baseline, with no extra IDs. `close()` alone does not surface stored recorder failure and is not a success oracle. Following successful admission proves no persistent slot leak, not a measured census of all internal allocation attempts. No extra dispatches for negative controls: local mutated-frame identity assertion must fail without network. Cleanup every recorder and runtime on assertion failure; no shared helper changes.

## Exact legacy selection and aggregate bounds

Select the following three unchanged source callbacks, and no entire suite:
- `Event reconnect syntax, minimum ordering and exact context codes are preserved over HTTP`:28 HTTP = seed1 + finite1 + initial HEAD1 +10 rows*2 methods20 +3 header-override HEADs + JSON ignored-header1 +unauth HEAD1. One store/listener.
- `conditions, HEAD,406 and shared cross-kind capacity preserve initial context without a stream slot`:13 HTTP = seed1 +2 iterations*(active1 +refused1 +four HEADs). One store/listener. Existing known condition/406/capacity coverage stays distinct from new unknown-minimum assertions.
- `disconnect during an Event minimum wait releases the request without capturing a stream slot`:4 HTTP =seed,HEAD,waiting request,positive live request. One store/listener. Existing3ms delay and narrower historical assertions are retained and attributed as such.

Thus exactly seven selected source callbacks/fresh stores, four listener lifetimes and74 HTTP (new29 +legacy45). New owner/service A–C only:18 live/delegated owner begin calls (A14+B3+C1), one separate raw post-close owner probe,7 service.beginEvents invocations (B5+C2),2 waits,6 minimum-clock calls,4 commits/3 stores. D and old callbacks perform real additional runtime owner work, not inferred into those counters. Source total8 seed/mutation commits. Four new observation records only; old callbacks have no newly invented reports. A–C's measurements cannot be labeled full-run owner counts.

## E: exact six-request installed witness, separately supervised historical witness

Append `observes genuine installed initial Scope Event SSE opaque minimum refusal`. Unique label `private-installed-scope-event-sse-opaque-minimum`. One child/store, one batch creating only a, default installed budget1000ms unchanged. Exact requests:
1. POST operations/batch200; validate actual mutationReceipt/currentBeadRecord and effect/required position/context.
2. GET events/ Accept:application/json200; validate eventPage and exactly one complete created Event matching actual receipt, next:null; obtain actual context/baseline.
3. HEAD events/ Accept:text/event-stream +known current minimum200; SSE metadata/context,zero body,no stream frames.
4. GET events/ SSE +matching unknown +If-None-Match:*503; validate transactionalProblem,code/retry/context,media and Retry-After1.
5. HEAD same request503;zero body, exact GET4 Content-Length/context/problem metadata parity.
6. GET events/ SSE +known current minimum200; validate status/media/cache/context/readiness, exactly one baseline id/type/complete Event frame, synchronous nonaborted/notdestroyed check immediately before deliberate close, and exact frame equality after drain.

One callback-local finite raw executor for requests1–5, allowlist POST batch and GET/HEAD canonical events/,5s request,131072 response bytes,16384 request headers,65536 request body, no redirects. Request6 uses recordRawSse with5000ms and canonical Host. A common pre-dispatch counter refuses a seventh request and records actual method/target/ordinal for both transports; five settled finite responses plus one bounded nonterminal stream, not six completed finite exchanges. No pagination/schema/receipt fetch/polling request. Schema compilation uses local pinned bundle and existing formats. Reuse actual configured artifact/schema and installed binding; public capability discovery remains unavailable and is not bypassed for qualification. Required overrides are both-or-neither; default local fallback executes and is labeled development-local. External evidence requires true outside-worktree realpath plus independently bound package/consumer/source identities; declared source string and witness-file fingerprint alone do not prove installed provenance.

Keep the existing restart callback separately selected: `installed Event cursors survive restart with Last-Event-ID precedence and source binding`, existing label `private-event-restart-cursor`, five HTTP (four finite plus one stream),two sequential children/one store with clean shutdown before reopen. Run it then E, each followed by its immediate report audit; three installed children/two stores/11 HTTP overall. Preserve its original45s/10s/60s/5s bounds and exact local controls. The earlier15-second sparse Scope/Bead/Link witness remains byte-preserved and unselected in this bounded local run; ordinary full CI still exercises its unchanged default selection. No claim it was rerun locally.

E owns45s work,5s finite/recorder bounds, one10s total cleanup envelope,60s callback and later120s supervisor (TERM115/KILL120). Default1000ms minimum waits leave qualitative request budget margin; no approximately200-attempt claim or wall-clock success threshold. Attach rejection observers immediately to launcher, finite requests, stream closure and child closure before races. Stop issuing work at deadline. On failure initiate stream stop, abort work, observe pending operations and close owned child within the same cleanup envelope. Record actual PID, exit and close/drain events, ownership return, exit code/signal, pending counts and directory removal. If pre-ready ownership is unknown or cleanup expires, preserve directory and fail; observe late rejection/returned ownership without asserting cleanup happened. Existing launcher may have additional internal kill/drain phases; do not await them beyond the outer reporting deadline or fabricate survivor absence. Aggregate primary and cleanup failures. Normal removal requires verified clean closure and drained owned operations. A local pending-executor count is not a socket-close inventory; raw recorder closure and child exit/drain are separately observed. Redact token/credentials in every error/report.

Total selected execution:9 callbacks,85 HTTP (74 source+5 historical+6 new installed),9 stores and10 owner lifetimes (7 source +3 installed). Three installed processes are sequential. No new full local suite or unrelated control cohort.

## Observation contracts (actual measurements only)

A `scope-event-sse-minimum-owner`: reads14,seedCommits1,cases array14 `{label,outcome,groupQueries,context?}`; packet rows and failure contexts asserted, positive query counts not fixed.
B `scope-event-sse-minimum-capture`: serviceCalls5,ownerAttempts3,waits1,clockCalls4,seedCommits1,interveningCommits1,attempts array3 with captured tuple/outcome/current context,final timeout pos2,closed:true. Record malformed/closed probe outcomes separately.
C `scope-event-sse-minimum-abort`: serviceCalls2,ownerAttempts1,rawPostCloseProbes1,waits1,clockCalls2,seedCommits1,operationOutcome:aborted,ownerCloseCalls1,closed:true; synchronous close observation explicitly limited.
D `scope-event-sse-minimum-http`: requests29,ordered records29 `{ordinal,method,target,status,bodyBytes,context?}`,elapsedMs,positive stream evidence `{ready,status,headers,framesBeforeClose,framesAfterDrain,liveBeforeClose,drained}`,one seed commit. Report only safe response fields; no credential. Header/body schema assertions stay in callback, not inferred from counts.
E label above: claimEligible:false,attribution,declaredSourceCommit,entry/hash,binding,limited witnessSource fingerprint scope,requests array6,completedFiniteResponses array5 with real headers/body/wire hashes,baselineEvents array1,sse `{status,headers,completeObservedFrames,liveBeforeClose,drained,wireBytes,wireSha256}`,one session with actual lifecycle,deadline values,pendingOperationsAtReport,directory/directoryRemoved and safe failures. Report success only after all assertions/cleanup. Unknown/failure report preserves available facts rather than filling missing success fields. Exact authored schema must be frozen before validation adapters bind it. No transport request counters replace semantic assertions.

## Review, validation and provenance sequence

This packet grants no execution. After root/native/actual Claude PLAN dispositions and green base release: source authoring on isolated checkout; freeze six-file diff/all source679, callback selection, schema and prefix preservation; independent native and actual Claude CODE review. Then root may release fresh dependency bootstrap/tool capture if needed; static typecheck,format,lint; exact fresh build+artifact binding BEFORE any collection of transaction-event-sse.test.ts (top-level dynamic dist runtime import). `pnpm build` already runs clean then tsc -b; no duplicate clean. Run only seven selected source callbacks via exact anchored title regex and audit their four new reports plus real verbose selected names. Fresh package archives/external consumer, actual lock acceptance, payload/manifest/source identity and installed proof precede separately supervised historical restart5/audit then E6/audit. Default fallback remains supported and full CI retains historical selection; no duplicate local default E is proposed merely for bookkeeping.

Prospective commands: pnpm typecheck; pnpm format:check; pnpm lint; pnpm build; pnpm exec vitest run packages/server/src/transaction-event-sse.test.ts --reporter=verbose -t <anchored union in selection.json>; separate pnpm exec vitest run apps/bdptest/src/transactional-event-sse-installed.test.ts --reporter=verbose -t <each exact title>, under the established root-released private/tmp supervisor and actual external overrides after provenance acceptance. Packaging/binding/audits use the established helpers, not hand-invented installed overrides. Preserve120s/TERM115 boundaries, no retries, stop first failure and keep failed packets. Required repository boundary/preflight/evidence and exact-head CI remain root's separate workflow. Final commit/publication only after actual validation and council; no release/Preview2 or main merge.
