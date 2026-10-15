# Sequence member transactions — proposed plan v2

**Planning only; ten-action narrowing is proposed, not accepted authoring.** Preserve original transaction-sequence-member-plan.md (SHA c9886e67e5abc5926329c5a621d29111d6b595d8d1a57bd49e28ab4e3ba9a33b). This is a supporting T2/T4 slice under the sole plan `/Users/dbox/repos/wt/vickie/bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md`, Autonomous profile completion. Root/native/Claude review, actual accepted parent and measured execution releases remain required. No product edits or execution are authorized.

## Priority and source chronology

Verified main is453b6dab386746bb03aac2c0a293572b8a71c969/tree40782d97c9764bc19ae78c46d1d39475e3db317a. The inspected pending command-minimum snapshot is full patch fab20b6534f645fbc705b4b9b7800394eca71eee59e8c53ccd6db5c320d7699d, preserved at transaction-command-minimum-evidence-draft/header-repair/draft.patch. It is not a commit, merge or executed evidence. Its five changed paths contain no production changes. An Event-ID oracle successor is now being authored separately; neither that mutable source nor its future evidence is silently incorporated here. Reanchor to its actual accepted ordinary merged parent before authoring.

Settled serving work still outranks evidence. Bounded inspection found no new safely independent production repair: command-minimum ignore and Selector admission are merged; maintenance fault propagation remains correct (transaction-store.ts1747–1759); prior current-read/alias/schema/SSE fixes remain integrated. Actual known gaps remain metadata-only incident Link membership (transaction-events.ts257–264), retained Snapshot/continuation minima (transaction-http.ts406–411), and public discovery/profile activation (HTTP195/228/567). The first two await existing decisions; activation must not advertise an incomplete profile. Do not infer new semantics from private unsupported tuple branches, introduce optional History/Types work, or ask another product question. This is bounded triage, not a universal defect-free claim.

An incidental pending-draft oracle issue was reported to root during this inspection: stored Event IDs pos-N-0 are not the served Event IDs. HTTP488–499→service.events→store.readEvents→projectTransactionEventGroup→renderTransactionEvent replaces IDs with ev_… (transaction-events.ts82–104,267). Root assigned a separate repair. This plan follows actual projection semantics and waits for that predecessor's disposition; no production defect or validation success is claimed.

## Law, scope and explicit limits

Spec docs/specs/bdp.md2720–2765 requires ordered, separately atomic, partially committing members, continuation after failure and outer200;4057–4080 gives each member its own transaction/pending and terminal receipt/group. Catalog transactional-v1.json866–876 owns transactional.sequence.member-transactions. Store1777ff admits claims,1848ff executes turns,1258–1295 retains failed receipts at previous position and commits successful groups separately. Sequence projection32–76 rewrites operationIndex/name for the current envelope. Resource-not-found remains404/after-state-change (spec1901), including a terminal retained failed key.

Prove terminal success/failure/success, independently retained receipts and no-effect replay. Do not claim simultaneous initial pending admission, interleaving/exclusive ownership, crash recovery, expiry/erasure, all operation kinds or whole clause coverage. wholeContract:false and claimEligible:false remain. Required251/unobserved251/report256 remain; anticipated authored14→15 and unplanned237→236.

Own only transactional manifest-v1.json, coverage-v1.json, transactional-cumulative.test.ts and transactional-matrix.test.ts. No fixture/runtime/runner/schema/catalog/spec/config/export changes. Preserve anticipated14old raw manifest objects plus14old coverage objects byte-for-byte and complete parent reconstruction. Preserve command-minimum mapping and all historical tests/oracles. Use catalog-compatible order, not reordered catalog/selection to evade observers.

## Proposed ten-action narrowing

The old s-groups claim was impossible in one GET: readChanges1691–1742 retrieves start+1, emits one group and a next link when head is later. retainedGroupReader uses SQL LIMIT1 at604; event scan maximumGroupsPerObservation=1 is fixed at transaction-events.ts20–23/store76. Finite Changes accepts only after/start (transaction-changefeed.ts91–112), so invented limit32 is also invalid. Drop only s-groups, rather than adding a continuation/route or claiming both groups/next:null. Distinct receipt transactions/effect positions and complete Events support terminal independence, but do not enumerate both Changes groups. Root must explicitly accept this narrower credit.

|#|Action|Required actual result|
|---:|---|---|
|1|s-sequence|POSTsequence, no outer key: createa/member-a; deletemissing/member-fail; createb/member-b.200 sequence schema, exact indices0/1/2 and created/known404Problem/created, currentpos2. Independently validate complete fixed Resource fields, absence facts, Problem and context before baseline.|
|2|s-first-receipt|POSTbatch equivalent createa/member-a.200 completed/available, required/effectpos1,nextnull, internalindex0, Resource equals initiala. Capture only actual/id to s-first-root.|
|3|s-first-facts|Real client receipt GET, currentpos2/historicalreceiptpos1; helper full-original equality and actual source/session/authority bindings.|
|4|s-failed-receipt|POSTbatch equivalent missingdelete/member-fail.200 failed/available, requiredpos1,noeffect/results/next; known404/after-state-change Problem has internalindex0 rather than sequenceindex1. Capture only/id to s-failed-root.|
|5|s-failed-facts|Real receipt GET with complete original failure retained and successful actual client/internal assertions.|
|6|s-last-receipt|POSTbatch equivalent createb/member-b.200 completed/available, required/effectpos2,nextnull, internalindex0 and initialb Resource. Capture only/id to s-last-root.|
|7|s-last-facts|Real receipt GET, full checked originalB retained/currentpos2.|
|8|s-replay|POST identical sequence, full parsed body equals independently checked initial envelope/currentpos2.|
|9|s-current|GETbeads/?limit=32, exacta/b ordered collection/full original records/nextnull/currentpos2.|
|10|s-events|GETevents/, exactly2created Events with ordinal0, correct subjects/types, full Resource data/revisions and corresponding receipt transactions, nextnull/currentpos2; no failed-member Event.|

Scope https://reference.invalid/acme/, Type types/task, titles member-first/member-last, empty metadata, no names/owned/provenance input. All three minimum headers absent. Literal fixture positions are not inferred semantics for opaque tokens. Receipt roots and transaction strings must each be pairwise distinct actual values. Initial and replay full bodies must have independently pinned known fields and absences, not schema validity alone.

Event IDs must use the actual served projection: independently derive the fixture ev_ epoch/view/source hash plus pos1/pos2/ordinal0 from captured actual context, never import production renderer as its own oracle or use stored pos-N-0. Full Event data/transaction must match the checked create/receipt. Finite Event read scans groups under its existing item/byte bounds (store1543–1635); exactly2/nextnull remains an actual assertion. No complete Changes claim remains.

## Raw capture, helper and hook contract

Apply accepted v5 correction/root dispositions994c9655. Raw captures permit scoped HTTP URLs only (runner2468–2487): retain three useful receipt roots, remove arbitrary body/revision/transaction captures. App uses existing schema validator plus independent literal/context/absence checks before cloning baselines. A sole literal replay-pair ledger drives s-replay→s-sequence whole-JSON equality; performed comparisons are recorded only after success and finish requires exact per-mode inventory. Static remove/redirect checks use named sequence-member-replay-ledger, no source regex or fake runner evidence.

Private transactional-observation.ts206–244 stores original raw receipt and303–307 compares actual GET before projection. For all3client actions pin no executionError, all authored assertions pass, sourceReceiptObservation.action to correct recovery POST, /id/root/operator/digest/session, and actual-status/actual-request/actual-body internal successes. Internal IDs are excluded from coverage mirror. Same-session equality is this fixture only. Full body is NOT in ObservedAction; do not inspect nonexistent report values. O headers follow helper empty-authored-header exact validation, not redacted values. Hook checks URL/key/bodydigest/writeState/order; HttpExchangeResponse has no method, so postrun additionally binds recorded method/action IDs. Coverage explicitly separates manifest assertions, app equality and helper equality. Programmatic JSON captures deliberately unused.

## Composition and one real control

Existing crossCarrierWire after its2/3/10prefix forbids later sequence URLs. Preserve that detector. In fullpositive and the two old cross controls, narrowly recognize the exact new first sequence URL+bodydigest+no outerkey+completewrite; require cross.finish() before handing it to the new seven-raw-response ledger. Recognize the identical replay only at its exact later ledger position. Forward every other response to the unchanged old sentinel, including old keys/routes after handoff; wrong/premature/extra sequences fail. Member finish checks all7ordinaryresponses, not the3internalGETs. Existing commandMinimumWire still sees all responses and finishes unchanged. Selectednewcontrol uses only its selected row's observer. Add data-only handoff/replay-ledger weakening assertions inside existing app static callback, no registration/runtime request.

One new registered negative selects ONLY this row. Change middle delete missing→a. Actual200 then has three successes/currentpos3. Keep exact seven baseline failing IDs: serving-position,middle-type,middle-code,middle-status,middle-retry,middle-no-outcome,middle-no-deleted. Keep middle-no-resource, first/last literals, indices/schema/media/request checks passing. Add independent actual-success assertions pos3/deletedkind/id/type/properties/metadata and app deletedrevision equals newly created firstrevision. Known-control hook must not mask manifest failures. Demand oneexchange/no clientaction/exact named failure set/all independent positive assertions/no transport,harness,execution or cleanup errors. Save genuine returned report before cleanup/postchecks. Static derivative bundle/schema/mutation checks precede execution. No response replacement or runtime seam.

## Measurements, finite validation and CI

Exact body companion: sequence414B, batchA147B, failed66B, batchB146B, replay414B:1187total/largest414. Negative408B. No body-size limit increase. Actual response bytes remain to be measured under unchanged target caps.

Select six existing static callbacks: the four cumulative and two app callbacks in the companion JSON. Extend existing assertions, no new static registration or unfiltered cumulative file/listeners. Top-level modules/schema/fixture initialization remain real CPU work. Normal offline setup/type/build/format/lint/boundaries/Read stages require separate release.

Proposed four installed cohorts once after exact clean frozen-source fresh archives/consumer provenance:

|Cohort|ExpectedHTTP|MaxHTTP|Children|
|---|---:|---:|---:|
|fullpositive15rows|213|213|23|
|selectednewdeletecontrol|1|10|2|
|existingcrossbatchcontrol|204|204|23|
|existingcrosssequencecontrol|205|205|23|
|TOTAL|623|632|71|

Old cross controls run because composition changes. Existing no-op/deletion/cross count checks increment10 while inputs/negative prefixes remain exact. No local13cohort expansion or broad suite. Unchanged90scohort/120stest-supervisor/positive63selapsed/5sexchange/10sprepare-cleanup; sequential maxWorkers1. Unexpected failure stops dependent stages, no hidden retries or replacement children.

Recurring source inventory: main52observe calls; pendingminimum54 (40mutated/14unchanged); proposed55 (41mutated/14unchanged),62expanded app cases.40existing full traversals add400HTTP/40children; ready5 add0HTTP/5children; global receipt denial reaches newaction3 adding3HTTP/1child; old2selected minimum controls and other6abort/preflight add0. Newselectedcontrol adds1expected/max10HTTP/2children. Increment404expected/max420HTTP/48children. Prior pending7796expected/958children→8200expected/1006children; conservative prior maximum7815+420=8235HTTP. This excludes separately inventoried unchanged non-observe startup/permission cases. Arithmetic is source-modelled, not execution.

Actual PR125 main CI38683826258 is750.80ssuite/969sjob under1200s:231sheadroom. This predates pendingminimum cost, NOT its future CI headroom. Require actual accepted predecessor CI/tail durations before release; extra48starts/404wires have no invented throughput guarantee. No job/test/runtime limits increase; defer or review a smaller slice if insufficient.

Anticipated canonical738783disk/750589normal/366164compact/19575nodes/210actions/2294assertions, coverage108091; mergedmain remains684206/695598/18107/199/2102,coverage99245. PendingEvent repair can change these numbers: remeasure, never carry them forward as observed final. Current anticipated sourceauthoring fences742879disk/754685normal/19639nodes/109115coverage and derivative760803/19795 have only margins. A new row will exceed them: this plan grants NO allowance change. A future measured draft must stop before tests for root's exact sourceauthoring amendment or refusal; runtime1MiB/100000nodes and all time/wire limits stay fixed.

Proposed drafting stop envelope retains original <=80000new disk AND normal bytes/2200nodes/15000coveragebytes/260assertions. Relative to the inspected pending snapshot, preliminary absolute maxima818783disk/830589normal/21775nodes/123091coverage/220actions/2554assertions; all41changed derivatives must fit joint836707normal/21931nodes including old6118B/156node overhead. These are unaccepted planning stops, not permission to raise guards/minify/drop oracles. Preserve first-use15route set/order/digest02dd7d9f35e7482c07ce1550a6572232cc48ba6858cd66f8dfc8f5384f3628be only after independently recomputing insertion; no new route is needed.

Done when actual parent/four-path scope and28oldobjects proven; tenactions/threelegalcaptures/exactmirror/falsecredit; full validated baselines/equalities and correct failureindex1versus0; sixstatic cases plus named weakenings; four actual cohorts within632/71 with exact failures; freshsource/archive/consumer/payload/config/report binding and independent PID/directory absence; native+Claude closure, unchanged-bound exactheadCI and root ordinary merge verification. No profile qualification or public activation follows.
