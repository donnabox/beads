# Native source audit of command-minimum evidence plan Claude findings

No confirmed production defect or need for a new runner/harness interface was found in F1/F2/F5/F6/F7/F8. These findings identify missing proof in the supplied excerpts. Close them by recording the source facts below and making the small existing-test pins/count clarifications explicit in the successor plan. Existing first-row receipt-corruption retargeting remains necessary, as already planned. No existing control needs to be dropped or newly scenario-filtered.

Basis: clean source102e140fdd8ba27bcff7d80320c94ec7ace57dc3/tree40782d97c9764bc19ae78c46d1d39475e3db317a. Root's integrated main has the same inspected tree; this audit's direct filesystem basis is102e140f. Frozen plan-v3 SHAe5c977762c38746bc5cbf33cefb9652a9c303fe073341520d6c52a370b8d6e50; actual Claude retry stdout SHA540a7599109b8b6d79eebd17f67c1f1d603912610d048af4e588c8079615e769. Reviewer did not author the future row. Only source/text/JSON/git/hash reads and this guarded report write were performed; no tests, imports, builds, services or package execution. Line anchors below are relative to this checkout.

## F1: cumulative applicability — closed by existing implementation

`packages/conformance/src/selection.ts:6–12` implements `profileIncludes` as `claimedRank >= requiredRank`; :55–63 selects each included, non-retired row. `packages/protocol/src/index.ts:54–58` orders read, read-update, transactional. `runner.ts:355,367–378` uses that selection for metadata/applicable IDs and optional exact scenarioSelection. `runner.ts:527` separately checks metadata/manifest requiredProfile equality. Thus a canonical requiredProfile read-update row is applicable under a transactional run without relabeling it. The catalog mapping already pins inheritance/no retirement; the new manifest must retain requiredProfile read-update.

Small plan pin: within the existing revised static mapping/ledger callbacks, keep exact inherited membership and requiredProfile; within the existing positive callback assert `result.selectedScenarioIds` contains the new ID and the row passes. The selected controls already pin exact top-level `scenarioSelection` and `selectedScenarioIds`. No new static callback or applicability implementation is needed. Unfiltered selectedScenarioIds contains all256 selected catalog IDs, not only14 implemented rows.

## F2: complete52-invocation inventory and hook collision audit

App file is `apps/bdptest/src/transactional-matrix.test.ts` (abbreviated A). All observe call sites and literal it.each arrays were read. The52 expanded calls divide into45 without an afterExchange callback and7 with one. The common wrapper A:267–284 may deny particular receipt URLs; its predicate is separately audited below. Nothing keys an old control on a generic first create/batch/sequence/update/collection/Event request.

| Call site / expanded group | Count | afterExchange predicate / state and new-row effect |
|---|---:|---|
| A:334–338 full positive |1|crossCarrierWire; waits for exact carrier-create sentinel, described below. New minimum-* row safely precedes it; +11wire/+1child.|
| A:566–582 changed ready fields |5|No hook. runtimeConfigurationIdentity/artifactIdentity/policyIdentity/authenticationIdentity/authorityMode rewritten on every row launch; before-action refusal,0wire, +1attempted child each.|
| A:600–605 altered receipt binding/method/credential/body-assertion |4|No hook. Must replace positional `value.scenarios[0]` with ID transactional.batch.atomic-commit as already planned; +11/+1.|
| A:800–817 ordinary Resource capture/follow |2|No hook; manifest mutations locate atomic-commit by ID; +11/+1.|
| A:851–858 real-wire budget abort |1|Unconditional first actual response: require nonempty wire, set seen, wait for abort. Intentionally now fires on m-create-a POST instead of old anonymous GET. Still one first wire; no extra completed row.|
| A:888–891 pre-aborted |1|No hook. signal.throwIfAborted at A:118 precedes bundle/bootstrap,0launch/0wire.|
| A:925–928 duplicate-binding preflight |1|No hook. Explicit atomic-commit a-point capture mutation rejected by bundle construction A:124–130 before bootstrap,0launch/0wire.|
| A:941–942 global invalid receipt credentials |1|No hook. Common resolver denies every URL beginning scope+receipts/. New row reaches action9 and fails its real dynamic GET401: +9wire/+1child; matrix continues to old rows.|
| A:959–966 child-kill plus abort |1|Unconditional first actual response: kill children.at(-1), abort controller, throw original. Intentionally first new POST; original failure/cleanup assertions remain.|
| A:987–999 wrong set matched/order/index/poststate |4|No hook; exact set.singleton-targets row/action IDs; +11/+1.|
| A:1026–1027 cross-wired set receipt |1|No hook; exact set.singleton-targets/set-delete-facts; +11/+1.|
| A:1058–1067 omitted update-where/delete-where route |2|Hook merely increments exchanges; preflight rejects before hook or row launch. Existing zero-wire/one-bootstrap checks remain.|
| A:1083–1086 denied set-update/set-delete receipt |2|No hook; common wrapper predicate is route+status+receipt key, described below; no new-row match; +11/+1.|
| A:1111–1155 set effects |7|No hook; each tuple supplies exact row/action/assertion (second postimage, rollback metadata, receipt order, Event order, zero results, zero Events, no effect); +11/+1.|
| A:1178–1189 pagination |6|No hook; exact receipt.pagination/action/assertion; +11/+1.|
| A:1217–1236 completed conflict |5|No hook; exact idempotency.conflict/action/assertion; +11/+1.|
| A:1337–1343 no-op |3|No hook; exact no-op row/action; +11/+1.|
| A:1422–1427,1668–1673 deletion controls |3|No hook; three applyDeletionControl variants locate deletion-safety by ID; +11/+1.|
| A:1772,2078–2086 cross-carrier controls |2|crossCarrierWire(control); same key-scoped sentinel; +11/+1.|
| **Total** |**52**|**40 ordinary full traversals +1 global receipt denial +5 ready mutations +2 aborts +4 prelaunch/preflight.**|

Exact cross-carrier hook A:1837–1909: initial `active=false`; return while `response.url !== scope+operations/create-bead || key !== carrier-create`; only that conjunction activates. New keys minimum-a/minimum-b/minimum-c/minimum-stale and the changed-title reuse of minimum-b cannot activate it. Once active, compare the ordered request URL/key/body digest/status/context against requests[index]. Its maximum is2 for batch-semantic-conflict,3 for sequence-fresh-key,10 positive. At completion it forbids carrier-create/carrier-update keys and create-bead/sequence/update-bead routes. The new row is **before** activation, not after completion: these strict completed-hook guards need no weakening. Collection/Event requests from the new row likewise arrive while inactive and are ignored. `finish()` still demands activation and exact maximum. There is no ordinal consumption before activation.

Common denial wrapper A:269–282: only when invalidReceiptCredentials is a string, status200, URL equals baseline updateWhere or deleteWhere, and parsed receipt.idempotencyKey equals set-update or set-delete does it add receipt.id to deniedReceiptUrls. The new row uses neither set route, so there is no collision even if entity IDs overlap. Resolver A:225–230 denies URLs in this set or, in global mode, receipt-prefix URLs. New-row internal receipt GET bypasses the outer afterExchange completely: runner.ts:898–920 creates its own raw executor from active session and calls it directly; :929–934 sets original credential and empty authored headers. It still uses active.resolveCredentials, so global denial applies correctly.

All source-targeted manifest mutations outside the one positional receipt test already use exact row IDs; existing action names such as m-update do not select another row. Scope/session isolation creates fresh per-scenario directories A:239–249. No new sentinel handoff, generic hook filtering, old-control selection refactor, or runtime change is warranted. Keep the planned existing-row fixture retargets and counts, and bind all11 actual returned new-row exchange IDs/order. Do not equate the10 outer raw-hook observations with11 exchanges.

## F5: ready-metadata refusal and exact CI increments

Binding file `apps/bdptest/test-support/transactional-execution-binding.ts:113–180`: launch child; parse/validate ready; compare served configuration and runtime identity **before** active session/dial route publication; on any mismatch close child and rethrow. authorityMode fails installedTransactionBinding:26–35; runtime identity mismatches:122–127; artifact/policy/auth changes fail served configuration equality (:121; profile-contract-plan.ts:171–178). The launcher itself is IPC, not a request. No HTTP executor is reachable on these preparation failures.

Runner catches generic settled preparation errors as harness-error without a stop category (runner.ts:784–787); `mustStopMatrixAfter`:448–475 returns false for category undefined, unless cleanup failed. Thus all14 authored rows are attempted after insertion, each refusal before actions/reopens, and normal clean refusal has15children including bootstrap versus14previously: exactly+1 per ready-control, zero new wire. Abnormal abort/cleanup failure stops earlier; it does not create extra row wire. Existing A:584–587 asserts the entire cohort exchange list is empty.

The plan's deferred continuation question can be replaced by these facts. Existing ordinary40 traversals add440wire, global denial adds9, selected new controls add5, total **454 expected incremental HTTP**. Existing ready5 add0; abort2 and preflight4 add0 incremental wire. Conservative per-row wire ceiling remains **41×11+22=473**, with both new selected rows allowed their full11 actions as a joint upper fence. Child increment remains **40+1+5+4=50 maximum**, under unchanged clean lifecycle/stop rules. These are static upper/accounting predictions; actual records remain required and deadlines are not guaranteed by arithmetic. Full suite950-ish seconds is not a new deadline allowance.

## F6: bootstrap dispatch is zero, not a hidden HTTP probe

A:158–165 launches bootstrap, consumes ready metadata with installedTransactionBinding, and closes it before route construction; no request API is called. `scripts/transaction-installed-observation.mjs:4–91` launches with inherited IPC, sends configuration, waits for ready, then sends close; no fetch/socket HTTP exchange. `packages/server/src/transaction-installed-process.ts:46–60` starts candidate and publishes ready, waits for stop, closes. transaction-installed.ts:5–13 captures config/computes local payload/starts runtime; transaction-runtime.ts:305–339 binds listener and opens local owner/service, not a self-request. Starting a listening socket is not an HTTP exchange. The bootstrap B is therefore **0 authored/dispatched HTTP** on this inspected path; no additional actual probe is needed merely to establish this source fact.

Retain exact planned full positive203wire/22children; selected controls4/2 and1/2:208expected/26children,225maximum/26children. Count bootstrap separately in child ledger and keep actual returned exchange counts. Source proof of zero bootstrap HTTP is not a child-attested provenance claim or a new measured run.

## F7: exact unobserved count

`packages/conformance/test-support/profile-contract-plan.ts:335` derives requirements from selection.required; :369 uses `requirements.map(({id})=>id)` for unobservedIds irrespective of whether authored. `transactional-cumulative.test.ts:196–197` already asserts exact selected.required IDs and new ID membership. Thus **251→251 unobserved**, **13→14 ready-unexecuted**, **238→237 unplanned**. Runtime report catalog count256 includes5diagnostics and is distinct. Add these literal deltas to plan; retain the existing exact-array assertion, optionally pair it with251 in that same callback. No extra execution or semantic change.

## F8: unused routes and readonly POST

runner.ts:1600–1650 preflights only selected exact actions, calls prepareExactRequest and validates each known resolved target. `http-executor.ts:548–565` uses `routes.some(route.url===target.href && route.method===method)`, not route/request-set equality. A selected one-row control may retain all15 valid canonical grants; unused routes do not fail preflight. Existing missing-used-route control proves the opposite direction, not an equality requirement.

A:136–137 configures reader with same principal and canMutate:false; A:219–230 declares reader handle and maps every non-anonymous/non-operator credentialRef to that reader token, with no method or URL restriction except the separately configured receipt-denial modes. So POST create-bead with explicit credentialRef reader resolves normally and reaches the real server; it does not become a0wire credential harness error. Request materialization runner.ts:1671–1673 checks the declared handle; reader is declared. The planned actual403 negative observation remains necessary proof, not established merely by this audit.

Smallest pins: existing new ledger callback pins readonly derivative action1 credentialRef reader and canonical grants unchanged; the new existing planned control verifies actual one response403, authored request pass and named receipt-success failures. The batch control verifies4responses/conflict. Pin both top-level selections and exact child counts as planned. No new helper, additional callback, filtered bundle, route permission or source path is required.

## Disposition and bounded successor edits

F1/F2/F5/F6/F7/F8 are closed as source-proof gaps subject to recording this inventory and the planned actual-result pins. The only source edits arising from these checks fit the already owned five paths: new manifest/coverage row; cumulative ledger and counts; matrix option pass-through/new positive+control assertions and existing exact-ID retarget/count adjustments; inherited mapping test. Do not edit selection.ts/runner.ts/harness/launcher/runtime or change the52 old selections. No broadened semantics, timeout, parser ceiling, public admission or qualification claim.

Root/author separately owns F3 mechanical merged-parent reanchor and F4 explicit O headers on9–11. Keep original reviews and source attributions. Root's measurement and code gates still precede execution; this report does not release authoring or tests.
