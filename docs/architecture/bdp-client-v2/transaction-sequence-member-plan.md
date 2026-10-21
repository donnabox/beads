# Sequence member transactions: bounded implementation and evidence plan

## Deferred status — serving rulings take priority

Root reported the operator's new decisions while this plan was being finalized: invalid Selector syntax is400 before key capture with no receipt; a valid over-capacity fresh request is413 with an unused key; equal retained retry/409 conflict precedes capacity checks. Mutation POST ignores the three Read minimum fields, while receipt GET still enforces them. Those concrete serving fixes supersede this evidence slice. **This sequence plan is preserved as deferred only: no authoring, council review or execution release is requested now.** Planning below records the inspected c02 basis and proposed future workload; it is not an active assignment. Root must refresh predecessor, measurements, inventory and scope before any later resumption.

## Status and recommendation

Planning only. This document proposes one partial cumulative row for already implemented sequence serving, not a production repair or complete Transactional qualification. Root must obtain native Codex plus external Claude council, resolve findings, and authorize the measured draft separately. No source authoring, guard changes, tests, build, install, pack, service, CI submission, push or new operator question is authorized by this plan.

Pending predecessor is **c02ccc3706f2e258d380805162a8c47dbb73f4e1**, tree **088290751951c353797b984f690b7d9fd96deb80**, inspected clean at `/Users/dbox/repos/wt/vickie/vickie-cross-carrier-draft`. PR122 CI/merge is root-owned and was pending at planning; this is not a claim that c02 is merged or green. Before authoring, root must verify the actual ordinary merged parent/tree and reanchor the plan, inventory and baseline hashes. Preserve ancestry; never rewrite the frozen predecessor. If the four owned files or relevant runtime/configuration changed, reevaluate rather than assume tree equivalence.

The prior audit `transaction-after-cross-carrier-next-audit.md` (SHA256 d2ab3e95f4bb796561d583408c07690ab15326386c721221f5218837b7552e3c) found no new safely independent serving repair in its bounded surfaces. This row fills a required evidence gap: successful members surrounding a terminal failed member commit separately, have their own retained receipts, and replay without effects. An answered serving decision takes priority before authoring/execution. At initial inspection four decisions were excluded: metadata-only incident Link Event membership; retained Snapshot/continuation minimum comparison S1-Q1; Selector admission status/key capture; command minimum applicability. The first two remain pending in this plan; root subsequently conveyed the latter two rulings recorded above. Their implementation is separate and higher priority. This sequence row does not implement or test those newly resolved laws.

## Contract and implementation basis

All anchors are relative to c02. `docs/specs/bdp.md:2718–2759` requires ordered, partially committing sequence execution, continued execution after failure and an outer200 envelope. `:3919–3923` gives equivalent singleton, one-operation batch and member identity; `:4046–4060` requires each member's own one-operation transaction/receipt/group under the unchanged envelope. Catalog `packages/conformance/catalog/transactional-v1.json:866–876` names `transactional.sequence.member-transactions`, whose selected text also requires declaration-order admission of unknown keys with pending receipts.

The terminal row below does **not** prove that all keys were simultaneously durably pending at initial admission. Coverage must explicitly remain `wholeContract:false`; cumulative `claimEligible:false` remains exact. Neither required-row presence nor passing diagnostic counts imply supported profiles.

Serving already exists: `packages/server/src/transaction-store.ts:1754–1777` admits unknown member claims, `:1825` onward processes turns, `:1239–1249` rolls back only failed member effects and retains a failed receipt at previous position, and `:1251–1280` creates separate successful receipt/group outcomes. `transaction-sequence.ts:32–76` projects original results/problems with current index/name. `transaction-http.ts:304–336` serves it. The Read Problem table at spec1899 fixes resource-not-found to not-found/404/after-state-change; transaction-batch.ts:984 retains its one-operation index0, and sequence projection rewrites that to current index1. Component sequence tests at `transaction-sequence.test.ts:180–218` and installed-contract precedent at `transaction-installed-contracts.test.ts:476–489` establish existing implementation intent, not fresh evidence here.

## Owned scope and preservation

Only these four existing files may change after separate release:

1. `packages/conformance/test-support/transactional/manifest-v1.json`
2. `packages/conformance/test-support/transactional/coverage-v1.json`
3. `packages/conformance/src/transactional-cumulative.test.ts`
4. `apps/bdptest/src/transactional-matrix.test.ts`

No production, catalog, fixture, schema, API/export, runner, loader, application configuration, CI configuration, timing/fault/clock or generic assertion vocabulary changes. No new afterExchange option or callback argument. Preserve all13 old raw manifest row objects and13 old coverage row objects byte-for-byte, their relative order, complete old assertion sets and immutable historical comments. Pin26 raw objects; removing only the new object from each artifact must reconstruct the complete exact parent bytes. Pin a separately computed canonical whole-new-row hash and explicit action/assertion/capture sets. Do not silently regenerate old JSON formatting or minify to fit a fence.

Append the new row in catalog-compatible order. Existing catalog/manifest membership stays250 required/255 total, with14 authored partial rows after this change. Missing catalog plans remain missing and claims remain false. Existing reopened action IDs and13 reopen observations remain unchanged.

## Literal requests and fixture

Use the existing fresh per-row fixture: Scope `https://reference.invalid/acme/`, Task Type `https://reference.invalid/acme/types/task`, allowed `{title: string}` properties, empty stored metadata, no Links or aliases. The source fixture schema and descriptor are at matrix test50–73. Keys below are local to the fresh row. Explicit IDs avoid allocation/local-binding semantics. All raw mutation bodies use exactly the shown property order under JSON.stringify; content-type and accept are application/json. The outer sequence has **no** Idempotency-Key. Equivalent batch requests have only the indicated top-level key. Request-authored headers/body and actual exact write state are asserted.

Define literal operation objects:

```json
{"operation":"createBead","id":"beads/a","type":"https://reference.invalid/acme/types/task","properties":{"title":"member-first"}}
{"operation":"deleteBead","bead":"beads/missing"}
{"operation":"createBead","id":"beads/b","type":"https://reference.invalid/acme/types/task","properties":{"title":"member-last"}}
```

First and replay sequence body:

```json
{"operations":[{"operation":"createBead","id":"beads/a","type":"https://reference.invalid/acme/types/task","properties":{"title":"member-first"},"idempotencyKey":"member-a"},{"operation":"deleteBead","bead":"beads/missing","idempotencyKey":"member-fail"},{"operation":"createBead","id":"beads/b","type":"https://reference.invalid/acme/types/task","properties":{"title":"member-last"},"idempotencyKey":"member-b"}]}
```

The three recovery batch bodies are `{"operations":[<the corresponding operation object above>]}` under keys member-a/member-fail/member-b, omitting member key and name inside the operation. They are the same semantic request; no field is changed to obtain a different outcome. No name fields anywhere in this row.

## Exact eleven-action ledger and positive oracle

There are8 ordinary HTTP actions observed by afterExchange and3 existing client receipt GET actions, yielding11 actual HTTP exchanges. Captures listed here are from actual parsed responses; a new binding is available only to subsequent actions (`runner.ts:1435–1471` assertions precede captures). No same-action equalsBinding dependence.

| # | Action | Actual request | Required result and capture |
|---:|---|---|---|
|1|s-sequence|POST operations/sequence, literal three-member body|200 #/$defs/transactionalSequenceResponse; exactly3 results with ordered indices0,1,2. First/last created complete Bead records with independent fixed IDs/types/properties/metadata and real nonempty revisions. Middle flat resource-not-found Problem, type `https://github.com/gastownhall/bdp/problems/not-found`, status404,retry after-state-change; no outcome/resource/deleted. No member names or receipt-envelope root/status fields at envelope level. Capture full body (`s-sequence-body`), successful records/revisions (`s-first-record`, `s-first-revision`, `s-last-record`, `s-last-revision`). Current serving position pos-2; real nonempty epoch/view.|
|2|s-first-receipt|POST operations/batch, first create, key member-a|200 mutationReceipt completed/available, exact key, requiredPosition=effectPosition=pos-1, next null, exactly one created result at internal index0 with no operationName. Complete resource equals first captured record. Capture actual root /id, full receipt body, /transaction as s-first-root/body/transaction.|
|3|s-first-facts|client transactional.receipt-read GET s-first-root|Real200 body present, full /body equals captured first receipt, proper content-type/context. Existing client output schema.|
|4|s-failed-receipt|POST operations/batch, missing delete, key member-fail|200 mutationReceipt failed/available; exact key, requiredPosition pos-1, no effectPosition/results/next; problem resource-not-found404-after-state-change with internal operationIndex0. No retained current sequence index1/name. The failed receipt is terminal under its key even though the Problem's normative retry value is after-state-change; never relabel it retry never. Capture root/body/transaction as s-failed-root/body/transaction.|
|5|s-failed-facts|client receipt GET s-failed-root|Real200, complete /body equals the actual recovered failed receipt, proper media/context.|
|6|s-last-receipt|POST operations/batch, last create, key member-b|200 completed/available, exact key, requiredPosition=effectPosition=pos-2,next null, one internal index0 created result with complete last captured record, no name. Capture s-last-root/body/transaction.|
|7|s-last-facts|client receipt GET s-last-root|Real200; complete /body equals the recovered last receipt; proper media/context.|
|8|s-replay|POST operations/sequence, identical body|200; full body equals s-sequence-body. No new allocation, effect or presentation shift.|
|9|s-current|GET beads/?limit=32|Complete exactly2-element collection a,b in the existing canonical order; full records equal captured first/last records and independently fixed literal semantic fields. next null.|
|10|s-events|GET events/|Exactly2 created Events, ordered for a then b, each ordinal0, correct complete created record/revision and the corresponding captured transaction identity; no failed-member Event. Match the existing Scope Event representation/schema rather than inventing another envelope.|
|11|s-groups|GET changes/?after=genesis|Exactly2 groups, pos-1 then pos-2, previousPosition pos-0 then pos-1, eventCount1 each, projectionAdvance false, erasures empty; each group's sole created Event and changes entry point to that member's actual record/revision and transaction. observedHead pos-2,next null. Match existing finite-changefeed shape.|

All ordinary responses include existing media/schema/context presence checks; positions are fixture-specific literal owner tokens, not a generic inference from opaque tokens. After action1 all responses serve current pos-2 even when their retained receipt's required/effect position is pos-1. Use metadata-equals only for equal-status current serving contexts; never compare retained position fields as though they were current headers. Client receipt actions assert their actual current context against the existing context capture/equals mechanism and full body equality. Binding receipt /id is accepted only from the real recovery receipt and uses the existing private receipt authority path (`runner.ts:850–950`); an ordinary Bead /id does not authorize it.

Use existing json-pointer/equalsBinding and ordered json-array-tuples for exact complete counts/fields. For mixed sequence outcomes, one tuple projects indices only and separate pointer assertions check each outcome; do not project an absent outcome/resource from the middle Problem into a mixed tuple. Schema validity alone is not a semantic oracle. Full captured-object equality is accompanied by fixed independent literals so a consistently wrong ID/type/body cannot pass.

The existing afterExchange hook reads actual unredacted bodyText and headers to check all3 recovered root strings and all3 transaction strings are nonempty and pairwise distinct, successful revisions match the initial sequence records, and subsequent response bodies match the prior actual values. Distinctness is over actual values, never report-redacted digests or placeholders. Receipt GETs do not enter this hook; their manifest/client assertions supply their proof. The hook does not fabricate a response, replace an executor, or grant authority.

## Real-wire order and old sentinel composition

`runner.ts:353–393` iterates catalog order. scenarioSelection filters membership and cannot reorder it. Merely moving a manifest row before cross-carrier therefore cannot solve hook composition. Changing catalog order or extending afterExchange with a scenario argument would expand scope and is rejected here.

Existing crossCarrierWire remains active after its expected prefix and rejects later create-bead/sequence/update-bead URLs and its two keys (`transactional-matrix.test.ts:1854` onward). A legitimate new later sequence would now trip that guard. Preserve its complete previous prefix assertions, finish(), and all checks over intervening failed-retained traffic.

Proposed smallest explicit adjustment for council: a test-local composed wire ledger, using the existing callback unchanged. Until a handoff, forward every ordinary response to crossCarrierWire except the exact first new sequence request. Recognize that request only by exact sequence URL, expected raw request-body digest, no top-level key, body-present and complete write state; call cross.finish() **before** accepting it. Wrong-body, premature or incomplete requests still hit/reject against the old checks. Then transfer to the eight-action member ledger: sequence, batchA,batchFail,batchB,sequenceReplay,collection,events,changes. Compare every method afterward in the existing report (HttpExchangeResponse itself has no method), and compare URL/body digest/key/write state/status directly in the actual callback. New-member key strings must never appear before the handoff. After handoff, old carrier keys and old create/update routes remain forbidden; only the exact two new sequence requests at their ledger positions are allowed. No blanket suppression of sequence-URL checks. End checks require the entire new ledger (or exactly1 in its wire control), reject extra new requests, and preserve old finish() checks.

Apply this composition to the existing positive, both existing cross-carrier controls and the new member wire control. Existing cross controls still stop their own row at the exact previous prefixes and then run the new row normally. The handoff does not excuse a missing old prefix or a late old carrier operation. All other observe invocations retain their current callbacks. Extend an existing static app case with data-only callback observations proving premature/wrong-body/extra/old-key sequences reject and only exact completed-ledger handoff succeeds; these are test-local observer checks, not claimed protocol evidence or new registered test cases. If this cannot be achieved without weakening an existing detector, stop for root/council; do not add a general exemption or a new runner seam.

This adjustment is explicitly measured/reviewed. It is not a source-authoring authorization.

## One negative actual-wire control

Add one registered app case, with one observe invocation. Mutate only first sequence member1's delete target from `beads/missing` to `beads/a`; all other raw request bytes/literal fields are unchanged. This actual request creates a, deletes it successfully, then creates b. Expected actual response is outer200, three successful results, middle outcome deleted with the full just-created Resource, and current pos-3. No fake response or changed expected value alone qualifies as a wire control.

Retain these seven baseline assertions in the first action so they fail, in the authored order: `serving-position`, `middle-type`, `middle-code`, `middle-status`, `middle-retry`, `middle-no-outcome`, `middle-no-deleted`. Preserve passing `middle-no-resource`, first/last created literals, indices, schema/media, authored raw bytes and nonempty epoch/view. Do not add a baseline whole-serving-context equality at action1 that would create an unplanned eighth failure.

Add independent control-positive assertions `control-position`=pos-3; `control-middle-outcome`=deleted; complete middle deleted kind/id/type/properties/metadata literals; and actual-wire deleted revision equals the just-created first member revision. That last equality occurs in the hook on the same actual body, not through an unavailable same-action capture. The hook expects actual control position pos-3 and success; it must not abort before the manifest can retain all seven intended failures. Require exact negative and positive ID sets, not only count. The static case loads the derived bundle, resolves every json-schema reference through the existing validator, verifies canonical-vs-derived **same-compact-form** digest difference, exact mutation preimage, complete retained baseline assertions and independently authored additions.

The new row stops after its first HTTP exchange; later actions are absent. This is a row assertion failure, not matrix cancellation. All13 old authored rows must pass. Expected complete matrix cost is192+1=193 HTTP and22 children (one new row plus predecessor bootstrap/sessions). Require exact one-request row prefix, zero client actions in this row, real200/write-complete, exact failed IDs, all independent success assertions and all child exits. Save the full genuine `{sourceCommit,entry,...observed}` report using the existing optional `BDP_TRANSACTIONAL_OBSERVATION_OUTPUT` pattern and a supervisor-selected distinct output path; no new env API is needed. Failure to obtain the expected shape is an unexpected failure, not permission to edit the oracle to match.

## Measurements before any execution or guard amendment

Canonical predecessor: manifest disk684206 bytes, normal `JSON.stringify(value,null,2)+'\n'`695598, compact execution339575,18107 nodes,199 actions,2102 assertions; coverage disk/normal99245. Manifest SHA256726fc623bbb115a0f39e0bb3aaa586ea37254522b254e55e555ea95587f14cb9; coverage SHA256a3a15b5143c118b24965698ec10289821664c90e935820b51c8cf76c05bdb1df. Nodes count each scalar/object/array recursively including root. Disk, normal and compact are different measurements. Reverify JavaScript serialization; do not rely solely on Python output equivalence.

Current live fences remain unchanged during measured drafting: canonical disk688302, normal699694,nodes18171; coverage disk/normal100269; all derivatives normal705812,nodes18327. Historical delta anchors are488280 manifest disk/70662 coverage/13019 nodes, with existing deltas200022/29607/5152. Do not silently run tests against an unreviewed raised fence or inherit the full planning envelope as the final guard.

Proposed **joint measurement stop ceilings for council**, not guaranteed maxima or executable guards: exactly11 new actions, at most260 new assertions; growth no more than80000 disk AND normal bytes,2200 nodes,15000 coverage disk AND normal bytes. Thus canonical ceilings764206 disk/775598 normal/20307 nodes/114245 coverage,210 total actions and at most2362 total assertions. These are a bounded authoring envelope informed by the preceding12-action230-assertion cross-carrier row (~64KB/1734 nodes), with room for three receipt comparisons and complete groups. They are not a license to add observations or inflate vocabulary. Report actual exact action/assertion/capture counts and byte deltas; if any simultaneous bound fails, stop rather than minify or silently increase it.

Inventory every actual transformed manifest:39 mutated invocations plus14 unchanged, retaining each derivative file, transformation/source location, normal/compact hashes, nodes, action/assertion counts, schema result and admission/HTTP behavior. Include the existing duplicate-binding replacement; it mutates a-point capture to commit-receipt and is rejected before bootstrap, not an unchanged manifest. Existing maximum derivative normal701716 and nodes18263 are overhead6118/+156 relative to canonical. Proposed derivative measurement ceilings are canonical planning maximum plus those existing overheads:781716 normal and20463 nodes, **both binding**. Measure the new control's true added assertions/serialization and allold transforms; if its overhead or any other transform breaches a joint stop, ask root to narrow/review rather than assume a fresh reserve. No claim that byte maxima and node maxima occur in the same derivative.

Preserve exact route first-occurrence inventory15 routes and its preimage, digest **a9a17908b6fa87674d5f573b584f4a1ce1d41d7909fadf48b65e9b05397ddfd4**. New actions use only existing POST batch/sequence and GET beads/?limit=32,events/,changes/?after=genesis; client roots use the existing dynamic receipt authority. No fixture/configuration bytes change. Assert each new ordinary request<=2048 and response<=8192 bytes, matching existing bounded-row practice; stop if a measured valid complete payload exceeds it. Runtime request/body/header/loader/node and timeout bounds are unchanged.

Freeze measured draft, four-file hashes, complete diff, raw parent reconstruction, literal request digests, all derivative manifests, exact full-control success/failure sets and source-derived cost ledger for native+Claude review. Only root after measured acceptance may authorize narrowly amended duplicate disk/normal/node/delta guards with small measured margins, all sites reconciled. Add new-row presence-before-equality checks to existing filters; retain complete old row-order assertions and append the new ID rather than replacing old lists. Two existing app static cases and39 cumulative cases remain the planned static selection; add assertions within them, not new registrations.

## Complete invocation inventory and recurring cost model

This inventory was derived from c02 `transactional-matrix.test.ts` observe call sites/it.each arrays and row-stop behavior, including N1 duplicate-binding correction. It is not an assumed multiplier over the positive. Existing20 lexical sites expand to52 observe invocations;38 mutated manifests+14 unchanged. Seven other expanded app cases do not call observe. Proposed one control makes21 lexical sites/53 invocations,39 mutated+14 unchanged and**60 expanded app cases**, not56. Existing cumulative39 cases are unchanged.

Counts below are HTTP/children per invocation; metadata child counts are conservative at-most because readiness mismatch can stop before a later child starts. Ordinary row assertion failure permits later rows; abort, pre-abort, invalid bundle and absent routes stop earlier as shown.

| Source site/category | multiplicity | c02 | proposed |
|---|---:|---|---|
|334 positive|1|192/21|203/22|
|565 metadata mismatches|5|0/at-most14|0/at-most15|
|600 receipt binding/method/credential|3|187/21|198/22|
|600 receipt body assertion|1|188/21|199/22|
|800 ordinary Resource capture|1|192/21|203/22|
|814 forbidden ordinary receipt authority|1|192/21|203/22|
|850 actual budget abort|1|1/2|1/2|
|888 pre-abort|1|0/0|0/0|
|925 duplicate-binding invalid bundle|1|0/0|0/0|
|941 all receipt GETs denied|1|27/14|30/15|
|959 abort with cleanup failure|1|1/2|1/2|
|987 set expected matched/order/index|3|177/20|188/21|
|987 set poststate|1|180/20|191/21|
|1026 cross-wired set receipt|1|184/20|195/21|
|1058 omitted set routes|2|0/1|0/1|
|1083 denied set-update/set-delete|2|178/20;184/20|189/21;195/21|
|1105 effects controls|7|179,172,176,181,181,184,181 /20 each|190,183,187,192,192,195,192 /21 each|
|1179 pagination controls|6|180,179,181,181,188,181 /20 each|191,190,192,192,199,192 /21 each|
|1214 conflict controls|5|184/20;188/21;185/20;189/21;192/21|195/21;199/22;196/21;200/22;203/22|
|1338 no-op controls|3|181/21;181/21;186/21|192/22;192/22;197/22|
|1668 deletion controls|3|176/21;181/21;177/21|187/22;192/22;188/22|
|2078 cross-carrier controls|2|183/21;184/21|194/22;195/22|
|new member successful-delete wire control|1|absent|193/22|

Reconciliation: predecessor52 model7342HTTP/908children. Forty existing invocations reach the whole new row: +440HTTP/+40children. Five metadata mismatches add noHTTP and at most5children. Denied-all row reaches first new receipt GET at action3: +3HTTP/+1child. Remaining6 invocations add neither. Existing52 therefore7785HTTP/954children; add new193/22 control => **7978HTTP/at-most976children** across53. These are planned ordinary/control paths, not a promise that an unexpected failure cannot terminate sooner. No rerun allowance is hidden in them.

The seven non-observe cases remain: bootstrap metadata rejection, permission-swapped configurations, inert import/invalidIPC configuration, IPC-disconnect/reopen, startup cancellation, and two static app cases. The permission test has its separate existing4HTTP/two-child workload; non-observe startup/import/reopen processes must be inventoried separately in the execution preflight, not falsely included in976. No new non-observe process work is proposed.

Explicitly amend all eight old bounded control total checks (three no-op,three deletion,two cross) and positive session/child/HTTP/action counts; their original negative prefixes/failure IDs and other-row-success assertions remain exact. Existing cases without explicit fixed total still incur the modeled new row. Source-derived inventory must be recomputed at actual parent and final authored source; unexpected count drift stops release.

## Finite staged validation proposal

1. Planning council and root acceptance; then isolated ordinary parent-child **measurement-only** draft under four-file ownership. No sourceexecution yet. Review all measured artifacts/diff and new hook composition before final guard edits.
2. Root may release one bounded offline frozen setup if needed, fresh build and relevant unchanged type/format/lint/boundaries/Read gates. Run39 cumulative static cases plus the two existing app static cases once (no installed cohort yet). Preserve their actual expanded titles/counts and prior source-only HTTP/listener workload; no new service cases are added. Existing predecessor static selection modeled14 sourceHTTP/max24 and12 listeners,0 installed children: verify against actual selection rather than treating this as a new entitlement.
3. Clean frozen source with factual doc/evidence checkpoint outside the four-file payload as root directs; separately authorized fresh packaging: same package closure and exact tar/installed/emitted/source comparison used by c02, actual entry realpath, saved source commit/tree, manifest/fixture/schema/config digests, allowlisted BDP overrides only, readiness payload and environment binding. No stale dist, unbuilt artifact or historical receipt relabeling. Only after provenance passes may installed cohorts begin.
4. Proposed installed selection is **13 cohorts once each**, maxWorkers1 and isolated saved supervisor/log/report per cohort: positive; new wire control; existing three no-op,three deletion,two cross controls; existing ordinary capture and forbidden-ordinary-receipt controls; existing denied-all-receipts case. The first10 include every changed explicit total/sentinel and total1933HTTP/220children. The last3 cover the affected unchanged30-second cases and new denied-client prefix (+203/22,+203/22,+30/15), giving **2369 expectedHTTP/279children**. A crude uniform ceiling13*203=2639HTTP and13*22=286children is only an upper supervision check; exact per-case counts remain mandatory and cannot be traded. Do not rerun a failing cohort automatically. No full historical local app file, optional profile, sequence whole-suite or extra healthy probe is included.
5. Preserve full genuine output reports for every selected cohort, exact selected titles, statuses, actual attempted requests, all assertion IDs, context/artifact/session binding, duration, readiness PID, child exit/signal and cleanup. Every owned process absent and private directory removed on both pass/failure; no shared database use. Failure stops the dependent stage for diagnosis. Root decides corrected-source release after new review.
6. Exact final-head full CI remains required, with all60 app cases and all workspace gates. Native+Claude code findings disposed, scope/ancestry reviewed, root merges and verifies actual merged tree. No release, Preview2 or profile claim follows.

Timing is unchanged: observe budget90s, existing positive63s, request5s and existing preparation/cleanup10s bounds; old selected30s tests stay30s. Existing individual control test/outer supervision120s remains. Do not increase limits to fit this plan. Authoring has no runtime grant until root releases a measured final command list. Parent c02 local positive took4.717s for192HTTP/21children on one machine, but that is not a bound for this row or CI. Existing c02 selected9-cohort evidence1641HTTP/189children is historical only.

Estimated incremental full-file time is approximately45–80s on a responsive host and70–120s on a slower CI worker:40 ordinary extra rows (~0.7–1.3s each), five preparation-only attempts, one3-request denied prefix and one new complete negative cohort (~15–30s). These are estimates with unmeasured startup/CPU/contention cost, not guaranteed maxima. Obtain real PR122 full-job baseline before execution planning: predecessor PR120941s/PR121817s under the current1200s job bound are context, not c02 timing. The candidate must finish actual full CI within unchanged1200s and each case's unchanged limit. Record full-job elapsed/headroom and test-level tail durations, not just a passing63s positive. If PR122 already leaves insufficient margin for the upper incremental estimate, stop before authoring or choose a narrower reviewed slice; no CI budget inflation.

## Narrowing and limits

Primary recommendation remains11 actions because three real receipt GETs connect each recovered member root to installed authority-bearing retrieval, including failed receipt retention. If measured size or CI capacity makes this too broad, one explicit council alternative is8 actions omitting those3 client GETs while preserving recovery batch receipt/body/identity and separate Event/group proof. That would only prove returned retained receipts, not their actual GET linkage. It saves115 modeled full-file HTTP (7863 instead of7978) but **no child launches**, so is unlikely to solve startup-dominated CI cost. Its denied-all case reaches all8 new ordinary requests rather than action3, changing the oracle/count; it requires its own reviewed measurement/inventory and must not be adopted silently. Another split into more independently spawned rows would increase setup cost and is not recommended merely to reduce JSON per row.

No interleaving injection, concurrent pending claim, all-keys pending-state proof, admission crash recovery, expiry, reauthorization/erasure, all operation kinds, alternate schemas, metadata-only Links, command minima, Selector semantics or retained Snapshot behavior is claimed. No extra tests are requested for those exclusions. If existing DSL cannot express the listed full fields/relations, stop for scope review rather than adding a generic assertion or fixture hook. This plan introduces no new product law.

## Planning verification

Only read-only local inspection and this guarded report write were performed. No source mutation, runtime/test/build/typecheck/install/pack/service/network publication occurred. The report intentionally proposes future bounds and gates; none of the new request counts, control outcomes or timing estimates are labeled observed. Root owns session verification and any outward work, code authoring and execution releases.
