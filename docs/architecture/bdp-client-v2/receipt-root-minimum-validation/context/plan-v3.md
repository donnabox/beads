# Receipt-root opaque minima — concrete plan v3 (precision successor)

**CLOSED, preparation only.** Root/native review and actual Claude PLAN council precede authoring. No source mutation, execution or review launch is authorized. This precision successor incorporates actual Claude PLAN findings1–11 and preserves its context-access limitation12; native exact-delta closure is still required. This incorporates critique `477641ce82f1538a11a70500d81666de85cf22be227d4d998ae40252921a7e6f`. Clean Snapshot successor08812c6c752c06cd0edf1af2878b2367692e5b32, tree9b6f50b9a6b9eea953a6407599090c259cf97331, is now the exact planning/authoring base: all679 source bytes equal the earlier dirty successor1 map, and exactCI39012707852 completed success/all27steps. Root still owns separate authoring release and must recheck these pins at release; do not substitute a branch tip or fetch newer main.

## Behavior and six-file boundary

An actually disclosed, principal-owned receipt ROOT with matching current epoch/view and lexical opaque minimum enters the existing finite waiter and ends503 catch-up-timeout if eligibility remains unproved. Spec5841–5851 permits this reference choice under the general remaining-read law; the narrower Resource MUST is not represented as a receipt-specific MUST. Spec5850–5851 explicitly limits that narrower rule for retained continuations/Snapshots and successful incident-Link lists. The positive receipt choice rests on the general remaining-read law plus explicit subsequent receipt controls, not on silence in that caveat. Spec5853–5857 explicitly includes subsequent receipt GET/HEAD,4997–5005 protects principal-bound recovery,5876–5883 requires minimum-before-shortcut and HEAD parity. Receipt nondisclosure derives from principal binding/current behavior, not a claim receipts are enumerated in the5834 Resource list.

Only these files may change:

- packages/server/src/transaction-http.ts
- packages/server/src/transaction-service.ts
- packages/server/src/transaction-store.ts
- packages/server/src/transaction-service.test.ts
- packages/server/src/transaction-http.test.ts
- apps/bdptest/src/transactional-minimum-receiving-installed.test.ts

No shared consistency helper, store-test fixture, runtime/schema/configuration/discovery/admission changes. Existing Scope/Bead/Link finite Event and fresh Snapshot behavior stays intact. Receipt pages, changefeed/reconnect, retained Snapshot/S1, incident-Link/G2, foreign opaque policies, History/Types/provenance remain excluded.

HTTP retains transport/method/auth and lexical receiving guards. Split the currently shared const minimum at transaction-http.ts551–553 into branch-local computations. Inside the page branch keep transactionMinimumHeaders(request.headers) only when hasMinimumControls(request), otherwise undefined, and preserve minimum ?? (hasMinimumControls(request) ? "unsupported" : undefined). After that branch returns, root alone uses currentReadMinimumHeaders(request.headers) with the same has-controls/undefined fence and existing minimum ?? "unsupported" call. Do not replace the shared parser in place, which would relax page input and lose its private sentinel. No new root URL classifier: a page-shaped identity passed directly to root lookup remains absent. Conditional/media handling follows the service decision.

Service receiptAtMinimum preserves synchronous run() check/registration, then unauthenticated refusal inside its async body, captured principal, explicit unsupported-sentinel branch, then snapshotCurrentReadMinimum(supplied) capture before waiter policy/clock. Admitted malformed tuples reject their promises; closed admission throws synchronously first. Capture once before retries. Do not move capture outside run() or into first attempt by analogy with Events. Explicit unsupported delegates once unchanged. No new AbortSignal API; receiptAtMinimum and its HTTP caller deliberately retain their preexisting lack of signal forwarding. receiptPage also lacks signal forwarding and remains unchanged; do not call the root the only such path. The finite registered lifetime drains on close, which is not disconnect cancellation.

Owner readReceipt preserves principal/id validation and data-only capture before guarded atomic entry. Specifically replace snapshotTransactionMinimum(minimum) with snapshotCurrentReadMinimum(minimum) in its condition capture (transaction-store.ts959–962). Within entry: principal disclosure; explicit own unsupported result BEFORE readContext; same-entry context; missing/foreign-principal absence; semantic minimum; expireRow(...instant()); projection. Each retry rediscovers the receipt. At transaction-store.ts970–971 replace checkTransactionMinimum(condition, context) with checkCurrentReadMinimum(condition, context, true), retaining the existing condition !== undefined && condition !== "unsupported" guard. Literal true selects Unproved for an unrecognized position with matching actual epoch/view. The owner remains directly callable and may throw that synchronous observation; receiptAtMinimum supplies the finite retry/deadline, not an owner loop. Known comparison remains. Foreign PRINCIPAL or missing identity returns contextual404 before comparison; only a disclosed owned root with unknown position plus foreign EPOCH or VIEW preserves private unsupported500. No early expiry/cached disclosure or receipt-detail change. Direct opaque calls become admissible where strict capture previously threw even for missing/foreign-principal identities; this deliberate direct API change is distinct from existing wire-sentinel nondisclosure.

## Exact selected source work

selection-v2.json pins12 final callback titles, their files/suites, source-derived workloads and prospective counters. New callbacks do not yet exist. Counts are not runtime observations or completion percentages. A/B and the existing receipt-specific tests share established semantics; no duplicate general token/parser suite is added.

### A: captures opaque receipt root minima through real owner retries

Use existing service fixture, real owner, batch/create/requireReceipt and overloaded delegating readReceipt-wrapper pattern from the receipt-absence callback. Record actual delegated returns or thrown TransactionMinimumUnproved/context and rethrow the identical error. This observes real owner Unproved directly; service timeout alone is not that oracle.

One store; actual seed submit creates a. budget3 with now:()=>tick++ and callback-local finite wait. During first wait only, mutate the caller tuple from unknown to eligible pos-2 and submit create(later) with a distinct key. Captured unknown stays unproved, actual contexts pos-1/pos-2/pos-2. No later wait performs a write. Then read with pos-01; known eligible pos-2; explicit unsupported; other principal; nonexistent root; unauthenticated access; malformed authenticated tuple. Exactly8 root-service invocations with3/3/1/1/1/1/0/0 live attempts: **10 root attempts,4 waits,11 minimum-clock calls,2 submits**. Unknown and pos-01 timeout with actual context; sentinel remains unsupported without context; known returns actual receipt/current context; foreign-principal/missing404, anonymous401, malformed rejected before owner. This callback does not itself claim retention preservation.

### B: drains an opaque receipt root wait before closing its owner

One seeded receipt/store. budget2/tick++ and one held wait gate produce exactly2 live owner attempts,1 wait,3 clock calls. Attach settlement observers immediately. Start opaque read, await first real wait, call close(), verify it has not settled; new root call synchronously refuses with no owner attempt/clock. Release gate; read times out and close completes. Probe f.owner.readReceipt directly after close once, bypassing the counting delegate handed to the service (separate closed probe, not a live attempt), and require the closed-store throw. Always release gate in finally and drain observed outcomes before fixture cleanup. No invented AbortSignal/disconnect claim.

### R: retain one shared failed-retention semantic body

Existing checks minimum before read-triggered expiry but honors independently forgotten receipts (%s) becomes four explicit labeled rows in the same body. Preserve final historical titles (false)/(true), forget flags and pos-1; add (opaque false)/(opaque true) with unknown. Use label-first tuples exactly [["false", false, "pos-1"], ["true", true, "pos-1"], ["opaque false", false, "unknown"], ["opaque true", true, "unknown"]] as const, retain the existing (%s) title format, and accept callback arguments (_label, forget, minimumPosition). This preserves the two historical rendered titles byte-for-byte; no copied body. The prior prose did not mandate object rows, so this is explicit spelling rather than a claim title changes were inevitable. Each row uses one real failed submit, other-principal lookup whose tuple also carries a foreign epoch (404 nondisclosure before comparison), then own root read.3 live attempts and1 wait per row. Minimum-clock calls are4 for no-forget and3 for forget (second own attempt returns404 without another failure-clock sample). Four stores. Existing controlled wait advances real fixture time; only forget rows call owner.maintain(). No-forget times out before read-triggered disposal; independently forgotten receipt returns404. Keep asserted waits and outcomes.

Select preserves current receipt absence context across ordinary and minimum service entrypoints unchanged:1 store,2 submits,4 root+4 page observations,0 waits,5 waiter clock calls. It preserves old contextual absence, anonymous no-entry and strict page controls. Its counters are source-derived here, not newly measured.

### D: serves opaque receipt root minima before HTTP shortcuts with private fences

Existing emitted HTTP fixture,1 store, budget2/tick++/resolved wait. Exactly10 network requests:

|#|Request|Expected|
|---|---|---|
|1|real seed batch POST|200 receipt/context|
|2|own root GET unknown, If-None-Match:*|503 catch-up-timeout|
|3|own root HEAD unknown, failing If-Match|503 empty|
|4|own root GET pos-01, Accept:image/png|503, not406|
|5|own root GET unknown, foreign epoch|preserved excluded private gap500|
|6|own root HEAD unknown, foreign view|preserved excluded private gap500 empty|
|7|own root GET partial minimum tuple|preserved shared receiving fence400/no context|
|8|other principal root HEAD unknown|404 current context/empty|
|9|missing root GET unknown|404 current context|
|10|anonymous root GET unknown|preserved shared auth fence401/no context|

Use actual seed epoch/view/current position. Check Problem schema/exact code/retry,503 private no-store/Retry-After1 and actual context; no private limit on normative responses. Non-attesting marker remains. Private500 stays bodyless/contextless. Rows5/6 preserve the explicitly excluded foreign-epoch/view opaque implementation gap; they do not attest normative409/410 or settle that policy. Rows7/10 stop at shared pre-route receiving/auth fences, not receipt-owner observations. For already paired rows2/3, require GET content-length equals actual serialized503 bytes, HEAD content-length equals that positive GET length, and HEAD has zero body octets. This is a wire assertion using the emitted runtime, not a promise the direct handler GET object already has content-length; no additional requests. Record ordered method/status/redacted route class. Three deterministic waits; HTTP observations do not expose owner attempt counts.

Select four existing HTTP regressions:

- authenticates and hides receipts before minimum-control refusal or conditional shortcuts:7 requests POST1/GET5/HEAD1.
- keeps receipt minimum enforcement after commands ignore copied controls:6 POST1/GET4/HEAD1.
- checks a stored failed receipt minimum before read-triggered retention disposal over HTTP:3 POST1/GET2.
- refuses unsupported root controls after nondisclosure but before read-triggered retention:10 POST2/GET4/HEAD4, unchanged title/loop scope. Add deterministic budget2/tick++/resolved-wait seam, retaining maintenanceIntervalMs60000 and real1100ms sweep-isolation delay. Four own results now503 with actual context/Problem/Retry-After1, other-principal four remain404, HEAD empty. Preserve final retainedReceipts() close/read-without-reopen equality [failed,completed]. This is the concrete no-disposal oracle. Four default1s waits plus1100ms would exceed default5s callback; fixture seam is required, no deadline widening. Other three callbacks remain unchanged.

Aggregate: **12 source callbacks/12 stores/36 HTTP requests (POST6/GET21/HEAD9)**. Seven service callbacks derive28 live root+4 page observations,9 controlled waits,33 minimum-clock calls,9 submits, plus1 real closed-owner probe. Do not call those source-derived aggregates new measured evidence or infer owner attempts from wire records.

## Installed historical22 and new E5

Historical rejects malformed read minimum tuples before observation while preserving receipt nondisclosure retains its exact22 requests,1 child/store, seed32 resources, real updateWhere/page, raw receiving matrix, allowlist, report schema and lifecycle. In own-unknown root/page loop only root changes to checked503 catch-up-timeout/after-delay/current context/private no-store/Retry-After1; page remains bodyless500/private minimum-position/no context. Other-principal root/page404 and malformed receiving controls unchanged. Do not reroute page, add requests or relabel historical evidence. Existing45s/10s/60s work/cleanup/callback bounds and5s request bound remain.

New E title: observes installed opaque receipt root timeout with principal nondisclosure. Reuse module artifacts/constants/schema validators; executor/lifetime/report remain callback-local. Same descriptor/schema (properties n:number), two random writer/other credentials. Actual seed creates beads/receipt-root with {n:0} under unique key. Validate receipt schema/id/requiredPosition/effectPosition and context. Exactly5 requests:

1. Seed batch POST200.
2. Own issued root GET unknown plus If-None-Match:* →503.
3. Own root HEAD unknown plus failing If-Match →503 empty.
4. Other principal root GET unknown →404 contextual.
5. Explicit nonexistent root GET unknown →404 contextual.

Minimum tokens and root URL come from actual seed reply. Grant only configured batch, returned root and explicit nonexistent root. Increment local dispatch counter BEFORE transport, enforce maximum5 and exact final5. Assert wire method/path/headers and response schema/context/cache/body semantics in callback; exact503 Retry-After1, optional validators absent, no private-limit marker. For the existing paired E requests2/3, assert GET content-length equals actual serialized503 bytes, HEAD has the same positive content-length and zero response body octets; use this pair, not extra requests. Keep non-attesting marker. Default1s bounded waits mean two actual timeout requests, not an asserted internal attempt count.

One child/store;5s request,45s work,10s combined transport+child cleanup,60s callback. A future separately reviewed/released local external validation-stage runner enforces120s nominal/TERM115; that is not an in-repository test, CI or product-code timeout guarantee, and the runner is not executed or authored by this plan. Bound response bytes as the existing executor does. Observe all rejection/late-cleanup outcomes and aggregate failures. Abort/destroy/drain owned requests; require timely child close, actual exit0/null signal before clean flag and deletion. Preserve directory on failed/unknown/pre-ready ownership; record actual PID/exit/signal/directory state. Cleanup deadline is failure even if child later exits. Do not rewrite historical22 cleanup as unrelated scope. New report claimEligible:false with actual installed binding and source/payload attribution; credentials/raw bodies stay unlogged.

Installed total **27 requests in2 callbacks/2 sequential children/2 stores** (22+5), immediate audit per callback. All selected stores14. Default local dist stays supported; installed evidence requires fresh external packed consumer, not fake override/old build.

## Gates and preservation

The native critique's five points are incorporated: finite existing retention budget, exact service and owner precedence, delegated real-owner observation, closed-store retention proof, and separate bounded history/new evidence. No new product decision or cancellation API is needed. Historical title containing unsupported is preserved as an evidence identifier; its assertions/documentation must clearly distinguish preserved sentinel/page fences from superseding owned lexical-root behavior.

Before authoring recheck the bound clean08812c6c successor and exactCI record; the clean-base rebind found zero intervening source-byte differences. Any later differences require a fresh audit/rebind. Freeze six-file source/diff/selection, obtain native and actual Claude CODE review. Future validation uses fresh measured tools/static checks, fresh exact build/bind BEFORE collecting HTTP suite (dynamic emitted import), exact12 source callbacks, fresh archives/external consumer, root acceptance of actual lock/inventory, historical22/audit then E5/audit. The future local external validation runner uses120s stages/TERM115 (not a CI or code guarantee), no retries, actual source/build/consumer/process/store checks. No full-suite local rerun merely for bookkeeping. Commit only after accepted local/runtime/installed outcomes, then exact-head CI; ordinary working-branch integration, no main/release/profile advertisement. Current Snapshot source remains untouched throughout this planning task.


## Actual review attribution and future prompt paths

Claude PLAN completed on the predecessor clean-base plan, exit0/1123.968s. It read the complete inline plan/selection and actual repository source/spec. The prompt's relative context location was not discoverable from stdin/cwd; actual Glob attempts failed. It did NOT read the copied design/matrix/rebind/CI artifacts or independently establish679-byte equality/tree/CI pins. Root/native/runner separately verified those facts. Preserve the original prompt, stream, verdict and this limitation; no retrospective access claim or rerun solely to restate pin evidence.

Future review prompts must provide the absolute artifact paths listed in review-context-paths.json, including this exact plan and immutable spec/design/CI/context copies, plus the actual repository cwd. Do not say merely beside this prompt or context/. This file is plan precision, not a newly completed Claude review of the successor; full implemented CODE review remains required.
