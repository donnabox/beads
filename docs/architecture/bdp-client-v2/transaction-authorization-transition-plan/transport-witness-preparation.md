# T1–T4 transport witness preparation (not execution)

This is a concrete test-authoring interface proposal for the accepted successor, not a source freeze or additional council. Production is concurrently being authored; all reference pins below come from committed67a2842, not moving worktree bytes. Product files unchanged by this preparation. Exact expanded test registrations and implementation hashes remain the final measured-draft gate.

## Interface coordination

Production author confirmed drafted private shapes: owner.authorizationSnapshot() returns frozen {revision:1,bindingKind:"scope",generation:0|1,authorizationView,readPolicy:"all-visible"|"empty",canMutate}; owner.replaceAuthorizationPolicy() refuses active claims/sequence attempts; owner.acquireAuthorizationLease(expectedContext) returns {signal,release()} after serialized equality check. Service/runtime replaceAuthorizationPolicy(signal) pauses/drains finite mutations then replaces. service.stream(start,suppliedSignal,expectedContext) acquires before invoking start. These are trusted source-private calls, never HTTP inputs. Verify exact final signatures with author before test edits. T1/T2 use runtime's trusted replacement and returned snapshot, not an exposed owner/database backdoor. T3/T4 use real owner/service fixtures. No extra production accessors are requested.

## Shared fixture and transport

Use canonical Scope https://event-sse.invalid/acme/ and Task type https://event-sse.invalid/acme/types/task, with existing Event fixture's minimal Task descriptor (no ownsOutgoing declaration). One fresh owner/directory/listener for each HTTP case. Important: existing runtimeFixture at transaction-event-sse.test.ts191+ overrides receiptMaximumItems=2. Do NOT reuse that override for T1. Dedicated local setup must retain normal32; this also controls finite Event page length. Snapshot page32 stays unchanged. No Types/config limit inflation. Runtime credentials are the existing private random token/principal p/canMutate true; never persist token-bearing configuration.

Finite requests dial runtime.listenUrl with canonical pathname+query and explicit Authorization, Accept:application/json (application/problem+json accepted for failures), POST Content-Type:application/json and literal Idempotency-Key. Request bodies below are keyless batch envelopes. Use a local bounded Node request/stream reader or the existing bounded raw finite transport pattern; preserve actual headers, no redirects/retries, timeout5000ms, decoded body cap32768B enforced while consuming (cancel/destroy and await close on excess). Null bodies fail JSON prerequisites; do not coerce them to empty objects. The fixture's unbounded fetch/text helper is not the new finite reader.

Live requests reuse recordRawSse: genuine Node raw header arrays, Host derived from listener, Authorization, Accept:text/event-stream, Connection:close; max2MiB wire/16KiB headers and5000ms socket timeout already implemented. Register rejection handlers immediately; headers alone are not an SSE Event observation. Record stream closure separately from a completed finite response. No automatic finite() page loop: it would add unbudgeted HTTP.

Proposed per-callback bounds to freeze with authored ledger: T1/T2 test30s inclusive of local cleanup reserve5s; T3/T4 test5s with explicit deferred barriers and no polling sleeps. All commands retain120s outer. Each await uses existing abortable transport or bounded barrier; final cleanup attempts every registered action and preserves primary plus cleanup errors. These are proposed test ceilings, not a production lifetime change.

## T1: narrows current Transactional visibility and fences issued handles after policy publication

Exactly17 requests, one owner/directory/listener, zero product children. No background probes.

| # | Request | Required independent oracle |
|---|---|---|
|1|POST operations/batch, key auth-transition-seed-32, body first seed below|200 mutationReceipt, completed/available,32 created results, exact absolute IDs/type/properties n0/metadata{}, nonempty revisions; save full receipt R1/context C1. required/effectPosition=pos-1, next:null.|
|2|POST same target, key auth-transition-seed-1, second seed below|200 completed receipt R2, one created b32; same epoch/view, required/effectPosition=pos-2; current context C2.|
|3|GET beads/b00|200 currentBeadRecord; full Resource equals R1 first postimage (current record links allowance handled by schema/source, do not invent ownedLinks).|
|4|GET beads/?limit=1|200 currentBeadCollection; one actual known record and non-null next; save actual next unchanged.|
|5|GET snapshot|200 snapshotManifest; epoch/view=C2, scopePosition=pos-2, positive-position checkpoint not genesis; beads first32 records and actual non-null beads.next, links empty, erasures empty. Save actual id/checkpoint/next. Manifest already contains firstpage; no extra pageGET.|
|6|GET events/ Accept JSON|200 eventPage;32 visible create Events at normal cap32 and actual next non-null; validate exact source/subject/transaction linkage against R1. Save actual final Eventid. This is the first page, NOT a complete33-Event history claim.|
|7|GET changes/?start=now Accept SSE|200 SSE headers/context=C2; await actual service stream registration by transport readiness (start acquires before writing headers). Keep socket open. No emitted Event/Changes checkpoint claim from this empty caught-up stream.|
|—|Trusted runtime.replaceAuthorizationPolicy(bounded signal)|Zero HTTP. Named authorization-transition:generation-and-lease: returned generation1/readPolicy empty/canMutate false/newview distinct from C2; real old Changes socket closes without client initiating abort. The control retains generation0 and uncancelled lease; bounded failure must be recorded before ordinary cleanup closes it.|
|8|GET beads/b00|404 resource-not-found Problem, no Resource identity/body disclosure, current epoch/newview/pos-2 headers.|
|9|GET beads/?limit=1|200 empty currentBeadCollection, items[]/next:null, current context.|
|10|GET snapshot|200 new manifest, same epoch/pos-2, newview, beads/links empty, erasures[]; actual new id/checkpoint validated.|
|11|GET saved request5 manifest.id|409 foreign-view full transactionalProblem/current context, no retained Resource body.|
|12|GET saved request5 beads.next|409 foreign-view same; URL is genuinely issued, not constructed.|
|13|GET saved request4 next|409 foreign-view; actual retained Read continuation.|
|14|GET changes/?start=now; Last-Event-ID=request5 checkpoint; Accept SSE|409 foreign-view finite Problem/current context, no stream frames; genuine Changes checkpoint from Snapshot manifest, NOT an Event ID or an SSE-emitted checkpoint. Header precedence over start=now exercised.|
|15|GET events/?after=<encoded actual request6 last Eventid>, Accept JSON|409 foreign-view finite Problem/current context.|
|16|GET exact R1.id|200 mutationReceipt completed/detail withheld. Same original id/key/transaction/epoch/view/requiredPosition/effectPosition/expiresAt; remove results,next,allocated. HTTP current headers are newview/pos-2, body historical R1view/pos-1. Full projected receipt equality, not only absence of a known Bead string.|
|17|Exact request1 POST replay, identical key/body|403 forbidden, no receipt success. Observe private state/position through trusted fixture if needed at zeroHTTP; no effect/group/revision mint. Current context if established by final authorized-refusal path; do not require an invented unauthenticated context.|

All successful/Problem bodies validate against actual schema before equality assertions: mutationReceipt, currentBeadRecord, currentBeadCollection, snapshotManifest, eventPage, transactionalProblem. Compare bdp-scope-epoch/bdp-authorization-view/bdp-scope-position exactly; context on GET is a serving snapshot, retained receipt body is execution history. No minimum headers or expired handles are introduced, keeping pending collision rulings out.

Finite Resource equality uses independently constructed literal identity/type/properties/metadata and receipt-derived revision, then checks postimage equality. Event schema/full facts likewise bind actual receipt transaction and created Resource; opaque IDs are saved/compared, never invented. Scope Event pagination is intentionally not exhausted. Snapshot first32 prove actual continuation; failure prevents later fabricated URL usage.

## T2: cancels an active Event stream after policy publication

Exactly4 requests: (1) batch create a,b with key auth-transition-event-seed; (2) GET events/ JSON, exactly2 Events/next:null with receipt linkage; (3) GET events/ SSE withoutafter, wait until actual id frames match the known finite Events and actual stream is registered; invoke trusted replacement at zeroHTTP, await server-side closure without first aborting client; (4) GET events/ JSON returns {source:Scope events/,events:[],next:null} under newview. Preserve epoch/position and original frame observations. Old already-delivered frames are allowed; no instantaneous network-recall assertion. This case proves real lifecycle termination and fresh empty projection, not deterministic queued-frame containment (T3 owns that).

## T3: queued-frame cancellation (zero HTTP)

Real owner seed a,b in one transaction; beginEventScan supplies an actual two-Event packet. Acquire old expected-context lease (or service.stream with that context), feed packet to existing pumpTransactionEvents and Sink. Sink's first write appends exact transactionEventFrame(packet.events[0]), resolves a dedicated first-write deferred, and returns false. Attach pump rejection observation immediately. After first-write barrier, assert pulls0/chunks1, perform real owner policy replacement, assert lease.signal.aborted, then emit drain and await pump terminal outcome. Assert chunks remain exactly the original first frame; no second Event, heartbeat or owner pull; all sink drain/close/error listeners removed, lease released. Returning false means first frame was accepted; do not assert it was recalled. No timer race or sleep is the causal barrier. Only explicit abort reason may explain expected termination; unrelated rejection fails.

## T4: handler-selected stale packet versus atomic lease registration (zero HTTP)

Real owner seed a,b; select actual Event packet/context while all-visible. Publish policy before calling service.stream(start,callerSignal,oldContext). start has independent counters for headers/write/pull and would consume the actual packet; require stale-context rejection before start invocation, therefore all counters0. This tests actual service/owner seam, not a mock policy result. Then capture current new context and acquire one fresh lease through same service entry. Its start resolves a registered barrier and awaits signal; caller abort/release settles it cleanly. New lease acceptance is not Event/body disclosure and adds no HTTP. Do not add a second owner or turn this into a test of arbitrary fabricated expected-context objects. Ensure old rejection handler attached immediately and release/abort every lease in finally.

## Cleanup and measurable limits

Snapshot cleanup-registration boundary BEFORE startup. Local finally closes each raw socket, aborts any bounded administration/lease work, awaits observed tasks, then runtime/service/owner close and only then removes owned directory; individually catch failures and aggregate with primary. Failed startup with unknown owner/listener stop preserves directory and reports uncertainty rather than assuming null means stopped. Existing shared afterEach can remain a fallback but must not short-circuit other callbacks on one error. Assert no firstStreamFault for expected authorization cancellation. T1/T2=21positiveHTTP/2owners/2dirs/2listeners; T3/T4 add2owners/2dirs/zeroHTTP/listeners. No installedchild or public-admission credit. Control stays separate7expected/17max, not part of this21positive count. Runtime controlgraph and freeze still root-gated.

## Exact seed body construction and data-only measurements

UTF-8 compact JSON key order is operations then each operation/id/type/properties; properties has only n:0. IDs b00…b31, b32, a/b as listed. Keys remain headers, never body members. T1 replay repeats first body byte-for-byte. These are proposed immutable test literal recipes, not actual wire observations.

- T1 first seed: 3728 bytes; SHA256 `a778feb6f8559743228db36139ccc653370c43a30cdcf3d2017db2e85349b054`.

- T1 second seed: 132 bytes; SHA256 `8f22a1dee8d5c16750f476c2c0f4883ca91c257014e07b9efef9d2e678a98220`.

- T2/T3/T4 two-Event seed: 244 bytes; SHA256 `77b717d875eb96e4f16756f40001700f9cf4848be09a7acfb2c7b44ccc36f072`.

Two-Event exact body:
```json
{"operations":[{"operation":"createBead","id":"beads/a","type":"https://event-sse.invalid/acme/types/task","properties":{"n":0}},{"operation":"createBead","id":"beads/b","type":"https://event-sse.invalid/acme/types/task","properties":{"n":0}}]}
```

Reference pins from committed base (production under authoring is not claimed frozen):
- `packages/server/src/transaction-event-sse.test.ts`: `69f58153b93947f028c37aebce2f57b0a25f725aacb9a1891c7b276b1c9e44c0`
- `packages/server/test-support/raw-sse-recorder.ts`: `09ff8343ac1a5dbad77eaf24fc2745cc28ea870ada6b05dc4ea1644c0ad5ef81`
- `packages/server/src/transaction-sse.ts`: `1a414cb252b1eb2b5c3695d16aecb59382cbd417876837351e1d80b91b45784c`
- `packages/server/src/transaction-runtime.ts`: `77954f0a85029d76443feac54cd53cda4339ffc126660ace5a9f5af92466c573`
- `packages/server/src/transaction-service.ts`: `53f655faf494209996aac2e0f93a4c2c785e471f66aac6af9ea41f92c9cfbe00`
- `packages/server/src/transaction-store.ts`: `fcd6882b29e538db45867c894c06dfbaa6ca588db0c2e403ef1b7abb14f6c706`
- `packages/server/src/transaction-http.ts`: `ebb3550110bd6fda638572966e5faa05c15f062f37a10947f6e442fc7d4570f4`
- `packages/server/src/transaction-snapshot-http.test.ts`: `61023b5f24396fe484c7820c706b560cb87abd7a3e16670f5a0944fbefd44d8a`
- `schemas/bdp-v0.schema.json`: `0baa5d5ef67144f8e6513b4cde31168effc8df6d52ef3141b0f144b87f8dc227`

Plan SHA256 `8f74cee8992d8d5b7c0c46cd640fe30971acb7022ebf19611dece0565cd2a521`; root authoring release SHA256 `fd8751aaa7929f71f17fa87f014275cf620a537c38722d910d87940ff0fd7226`. Raw recorder cap and cancellation: test-support/raw-sse-recorder.ts; deterministic Sink/pump pattern: transaction-event-sse.test.ts37–119; real listener fixture191–281; normal Snapshot schema/setup: transaction-snapshot-http.test.ts47–115. Existing pump write checks signal before write and removes drain/close/error/abort listeners (transaction-sse.ts60–100). Final code must remeasure exact file/title/body/count inventory before runtime release.
