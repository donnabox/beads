# After no-op evidence: operation-time Bead deletion safety — v4 draft

Proposed next U3/T4 slice only. This separate v4 draft incorporates root-accepted native P3, the completed external review with root dispositions F1–F9, and the native-v3 mechanical receipt-pagination index correction. Original f99e40, v2 66814de1 and v3 f8bb4652 remain frozen. V4 awaits independent native closure and final root acceptance. No implementation or claim promotion is authorized. PR114 is merged as 5995f6b08ebd045e4343059c2da1cffd38b09b80; root assigns that base upon final acceptance, not before. The sole plan remains `bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md`.

## Recommendation and inspected basis

Add one partial required row, **`transactional.batch.deletion-safety`**. Use two opposite unowned Links between two Beads. Prove that removing only the outgoing Link cannot permit deletion while the incoming Link remains, and vice versa; earlier updates/Link deletions roll back, and a later Link-deletion suffix cannot rescue the failed batch. Then delete both Links before deleting the Bead and prove the ordered transaction succeeds.

This tests practical graph teardown and transaction ordering, including destructive-prefix rollback. It is distinct from the current no-op and completed-key-conflict rows. No product ruling or new runtime, schema, DSL or fixture configuration is required by this fixed visible/unowned partition.

Read-only basis: `/Users/dbox/repos/wt/vickie/vickie-noop-conformance`, frozen **bec97a7a00b8df663a62004e1ee640303d8527fa**, tree **f173c70f0806b72fa75075b1cdb1ce35d0c14bf0**, clean. Read the sole plan's live ledger/decision section, `transaction-next-unblocked-serving-audit.md`, actual catalog/spec/evaluator/store/HTTP test/runner paths, current artifacts and `transaction-noop-conformance/report.md`. Pending SSE/minimum/incident-view questions remain untouched; no new serving audit or public-discovery workaround is proposed.

Current merged parent is **5995f6b08ebd045e4343059c2da1cffd38b09b80**, tree **a1ae75b822e1d7fadd739d818d09a25fb788e380**, per root's verified PR114 merge. That tree equals candidate8d66855ec4e26d3af0b3d7d22d621e1f37005294. The latter's sole delta from bec97 is four added/two removed comment lines in `transactional-cumulative.test.ts`; executable assertions, runtime, manifest and coverage are unchanged. Existing installed evidence remains attributed to bec97. Root assigns merged5995 as the implementation base only upon final plan acceptance. This planning pass uses local8d668 source without fetching or relabelling historical execution.

Alternatives considered:

| Candidate | Assessment |
|---|---|
|Deletion safety at each operation|Recommended: settled law, practical ordering/rollback discriminator, existing batch/receipt/current-read/Event/changefeed seams, one additional already-implemented collection URL.|
|Batch local-reference/semantic-identity normalization|Useful later, but the existing atomic row already exercises a local creator/update/endpoint. A material next partition needs careful pinned-reference/retained-identity scope and is less immediately valuable than destructive rollback.|
|Sequence failure/dependent projections|Useful implemented behavior, but sequence result projections are not canonical batch receipt grants in the present private observation helper. A full receipt-retention claim would require additional authority design; do not disguise that as a four-file evidence-only slice.|
|Revision/aggregate races, pending recovery, hidden policy|High value but not deterministically expressible as the selected serial fixed-fixture row without new scheduling or policy fixtures. Leave explicit; do not invent concurrency or authorization evidence.|

## Exact law and coverage gap

Catalog `packages/conformance/catalog/transactional-v1.json` row `transactional.batch.deletion-safety` selects `docs/specs/bdp.md#explicit-bead-operations`:

> `DeleteBead` fails if any live Link is incident upon (attached to) the Bead when the operation is reached.

Spec2303–2305 also says incident Links must be explicitly removed before the Bead. This is an **operation-time** test, not a check of the transaction's intended final graph. Existing runtime `resource-evaluator.ts:1023–1029` checks revision and then `tx.incidentLinks(id).length`; the transaction runs against staged state. `transaction-store.ts:1239–1248` rolls effects back and constructs a failed receipt. Batch error conversion preserves `incident-links-exist` (`transaction-batch.ts:507–552`); its failure is a receipt Problem with would-be status409, not an outer409 response (`bdp.md:5299–5315`). No new precedence is inferred.

Current installed `transactional.batch.rollback` fails a later revision guard after updates/creation, with no Link deletion. Existing set rows cover set effects/cardinality and deletions, but do not bind this required row or test both incident directions with only one of two opposite Links removed. In-process precedent includes incomplete-cascade refusal/rollback (`transaction-http.test.ts:1771–1808`), explicit **self-Link**-before-Bead success and tombstones (`transaction-store.test.ts:625–670`), and rollback of a preceding Link deletion on a later guard failure (`:671–696`). These are source support, not installed observations under the new catalog row. The self-Link precedent does not establish this distinct-endpoint partition; its seven deletion facts are derived from the per-Link source/target emission in `transaction-batch.ts:756–769`. The adjacent spec2305–2306 sequence-interleaving sentence is outside this atomic-batch row.

Keep this row and all eleven predecessor rows `wholeContract:false`, overall `claimEligible:false`. It observes two local visible unowned bare-reference directions and ordered success, not hidden incidence/nondisclosure, owned Links, self-Links, pinned/external endpoints, concurrent incidence insertion, alias locators, retention/restart, or all optional configurations. No credit is borrowed from other rows. If successful,12 partial required rows leave238 required plus5 optional unimplemented; canonical250/255 selection is unchanged.

## Fixture, mutation inputs and independent oracles

Use unchanged716-byte `reference-transactional-v1` and the existing installed Task/Relationship descriptors. Relationship is unowned, properties permit title:string. Scenario setup is ordinary HTTP mutation, not a fixture/config expansion:

- Seed a and b with titles `a`,`b`, metadata `{}`.
- Seed l:a→b and r:b→a with titles `l`,`r`, metadata `{}`. All IDs are literal canonical Scope-local identities; no aliases, labels, pins, selectors or minimum headers.
- Incoming-Link-survives refusal batch: update b to title `incoming-survives-prefix` and metadata `{phase:"incoming-survives-prefix"}`; delete l; delete a; delete r. Expected failure at operationIndex2, while r remains incident. The later r deletion must not rescue the batch.
- Outgoing-Link-survives refusal batch: update b analogously to `outgoing-survives-prefix`; delete r; delete a; delete l. Same failure index2, while l remains incident.
- Ordered success batch: delete l; delete r; delete a. Bead b remains exactly its original full record. All operations omit guards; no raw interpolation of captured revisions is needed.

Distinct literal idempotency keys for the four POSTs, same authenticated principal. Each failed batch must return actual HTTP200 with `mutationReceipt` schema, status `failed`, detail `available`, requiredPosition pos-1; no effectPosition/results/next. Problem is exactly code `incident-links-exist`, type `https://github.com/gastownhall/bdp/problems/conflict`, status409, retry `after-state-change`, operationIndex2. No unsupported failure-envelope or direct-status inference.

The seed receipt's actual private GET supplies full a,b,l,r records, revisions needed for deleted identities/Events, and seed transaction token. First validate the raw seed receipt's literal IDs/types/properties/metadata/endpoints and exact four-created result order. Raw HTTP captures remain URLs only; existing client receipt capture handles records/tokens. The final successful receipt is likewise read through the current private receipt helper to capture its transaction token after literal deleted-result validation. No raw body/array capture or new helper.

Before success, complete current collections must contain exactly a,b and l,r in canonical order, `next:null`, and every complete item must equal its independently validated seed capture, including revisions/provenance. Explicit literal tuple inventories also retain IDs/properties/metadata, so a same-ID metadata drift control can check a substantive rollback fact. Strict `currentBeadCollection`/`currentLinkCollection` schemas apply. Failed prefixes must leave full b unchanged and leave the staged-deleted Link unchanged because its deletion never committed; surviving suffix Link and Bead a must also remain.

After success, Bead collection contains only original b, Link collection is empty, both `next:null`; GET a is404. Successful receipt contains exactly three deleted entries at indices0/1/2: l,r,a, correct resourceKind and canonical id/type/final revision matching the seed captures, with required/effect pos-2. No opaque revision guessed.

### Scope Events and change groups

Seed has **eight** ordered Events: created a0, created b1, created l2, linked l/source3, linked l/target4, created r5, linked r/source6, linked r/target7. Four Resource upserts, eventCount8. Both failed batches leave that exact Scope inventory and pos-1 unchanged; no prefix update/deleted/unlinked fact escapes.

Ordered successful deletion adds seven facts: deleted l0, unlinked l/source1, unlinked l/target2, deleted r3, unlinked r/source4, unlinked r/target5, deleted a6. Final Scope Event page contains15 facts, `next:null`; check complete ordered five-field tuples in exact projection order **source/type/subject/subjectType/ordinal** (pointers `/source`, `/type`, `/subject`, `/subjectType`, `/ordinal`), endpoint-specific unlink data, canonical identities, deletion revisions and final transaction token against the receipt. This fits the existing32-item finite page and avoids pending incident-source projection semantics.

Existing finite `changes/?after=genesis` serves one group/page. At seed and both failures require exactly seed group pos-1/previouspos-0, headpos-1, next:null, seed transaction, four captured upserts, eight exact Events and eventCount8. After real successful deletion, same route still returns the retained seed group, headpos-2 and a non-null next URL. Validate the latter with an ordinary scoped-URL `/next` capture, never fetch it. Do not falsely expect two groups/page or claim observation of the deletion group's tombstone payload: its effect is independently checked via receipt/current state/Events, while the final genesis read demonstrates the group/position advance. This is the same already-reviewed capture mechanism, not new authority.

## Exact21-action ledger

|#|ID|Action / decisive evidence|
|---:|---|---|
|1|d-seed|POST batch creates a,b,l,r, pos-1; validate four created entries, capture root URL.|
|2|d-seed-facts|Receipt-client GET; validate and capture four records/revisions and transaction token.|
|3|d-seed-events|GET events/; exact eight seed facts and no next.|
|4|d-seed-groups|GET changes/?after=genesis; complete seed group/headpos-1/next:null.|
|5|d-incoming-survives|POST incomplete teardown with l deletion; failed receipt/problem/index2, pos-1.|
|6|d-incoming-survives-beads|GET beads/?limit=32; full a,b unchanged, literal b metadata{}, no next.|
|7|d-incoming-survives-links|GET links/?limit=32; full l,r unchanged; staged deletion never committed, no next.|
|8|d-incoming-survives-events|GET events/; exact original eight, pos-1.|
|9|d-incoming-survives-groups|GET changes after genesis; exact original group, headpos-1/next:null.|
|10|d-outgoing-survives|POST incomplete teardown with r deletion; failed receipt/problem/index2, pos-1.|
|11|d-outgoing-survives-beads|GET complete Bead collection; full a,b unchanged.|
|12|d-outgoing-survives-links|GET complete Link collection; full l,r unchanged.|
|13|d-outgoing-survives-events|GET events/; exact original eight.|
|14|d-outgoing-survives-groups|GET genesis changes; exact original group/headpos-1/next:null.|
|15|d-delete|POST delete l,r,a in that order; completed ordered deleted identities, pos-2, capture receipt URL.|
|16|d-delete-facts|Receipt-client GET; independently validate final result/position facts, capture final transaction token.|
|17|d-final-beads|GET Bead collection; only full original b, no next.|
|18|d-final-links|GET Link collection; exactly empty items, no next.|
|19|d-deleted-a|GET beads/a;404, authored request and present current context pos-2.|
|20|d-final-events|GET events/; complete15-fact sequence, deleted identities/revisions/transaction, no next.|
|21|d-final-groups|GET genesis changes; original group/headpos-2, validate scoped non-null next URL without following.|

Every raw action pins authored request headers/body and appropriate response schema/status; collection/group/Event oracles prove complete cardinality, not selected indices alone. Mandatory epoch/view/position presence precedes metadata equality. Compare200 contexts to their seed or successful POST baseline; the404 action uses explicit known position/presence, not status-folding equality against200. No receipt position substitutes for independently served read context.

The new row has19 raw HTTP exchanges (4POST/15GET) and2 receipt-client GETs; zero reopen. Total187 actions =132raw+48receipt-client+7reopen, **180 actual exchanges**. Positive12 pass rows,19 scenario/reopen sessions,20 children including bootstrap. Client-only `row.actions` must contain exactly the2 client IDs; `row.exchanges` contains21 ordered IDs. Do not repeat PR114's fixed bookkeeping error by treating client actions as all exchanges. Built-in client wire assertions are reported separately from authored assertion totals.

## Four-file ownership and preservation

Only the current Transactional manifest, coverage, `packages/conformance/src/transactional-cumulative.test.ts`, and `apps/bdptest/src/transactional-matrix.test.ts`. Insert row after rollback and before no-op in catalog order. Exactly one additional literal allowlist route: GET `https://reference.invalid/acme/links/?limit=32`. The existing changes route first occurs in this new row but keeps its relative place. Independent12-route order: beads/a GET; batch POST; links/l GET; beads collection GET; events GET; beads/new GET; changes genesis GET; **links collection GET**; beads/b GET; beads/c GET; update-where POST; delete-where POST. New route is eighth (zero-based7), no dynamic fetch. Recompute full canonical configuration digest/preimage; never reuse parent90619905….

Preserve all eleven parent manifest/coverage raw row objects and the fixture byte-for-byte; preserve all existing semantic assertion IDs, row-local pins and control mutations, including no-op223 assertions and successful final URL capture. Explicitly adapt aggregate bookkeeping/size/configuration pins listed below; preserving control semantics does not mean preserving obsolete19-child/159-exchange aggregate bounds. Add no-op to the current ten-original-object guard:

|Raw no-op object|Bytes|SHA256|
|---|---:|---|
|manifest|81047|`9307b230eb053131e9127cbbb5e6285305b79d6794edd22f5432a07fc74d9faa`|
|coverage|10587|`383de4e0cb9b3bad310b9db7ff9009cb59fac55966f0309bc71c5d007bea6c1c`|

Parent manifest488280B/13019nodes/SHA0358cbf7416321bb911e990259236f89bc2720cbc29d037581f542080d9f32ab; coverage70662B/SHA8a770b4371689a3d9d5f84ecfc7a1d9b72715c81e617cf0ea5ae04e55c66098e. Fixture716B/SHA3a7899ced771d4880066d98b0000c53f801ece53458502cc4f3bc641146c9a85. These inspected bytes now match the root-verified PR114 merged tree; they remain preservation pins, not evidence from a newly run merged-source cohort. Rebase/integration drift requires explicit reconciliation, not refreshing pins blindly.

## Complete aggregate re-pin inventory

Line anchors below are inspected8d668 source, byte-identical to merged5995 outside the same documented comment repair; reanchor by symbol at authoring. Preserve semantic oracles, update only explicitly affected aggregate pins and add the new row's independent pins.

|File / current lines|Required update|
|---|---|
|`apps/bdptest/src/transactional-matrix.test.ts:184–196`|Insert Link-collection route at index7, keep other relative order; twelve routes.|
|`:330`, `:477`|Positive title and summary label eleven→twelve partial rows.|
|`:339`, `:341–342`|Pass rows11→12; scenario/reopen sessions18→19; children19→20, including explanatory comment.|
|`:347`, `:1216–1237`|Recompute independent full route/configuration digest and canonical preimage; both digest assertions use newly measured digest.|
|`:407–418`|Include deletion-safety in existing2KiB POST/32KiB response row filter; retain exact limits.|
|`:473`|Positive observed exchange count159→180; separate client action count2 from new row21 exchanges.|
|`:1264–1341`, especially`:1312`,`:1314`|Existing bead-property, link-metadata and group-inventory-and-head control mutations/IDs/oracles unchanged. Change aggregate child maximum19→20 and exchange maximum159→180; execute each once on final source. Expected arithmetic is169/169/174 exchanges if only their existing stop points apply, not evidence until observed.|
|`packages/conformance/src/transactional-cumulative.test.ts:189`|Unplanned required239→238; canonical250 required/255 total unchanged.|
|`:691–796`, `:1317`|Add exact parent no-op raw manifest/coverage object pins to prior ten; preservation table count10→11 (22objects).|
|`:805`,`:812–824`|Update structural-test wording as needed; insert deletion-safety after rollback before no-op in full ordered ID list.|
|`:826–838`,`:847`|Add row action21, preserve prior row counts, total166→187.|
|`:930–931`|Receipt-pagination structural scenario-index pin8→9 because deletion-safety is inserted at index2. Preserve the receipt row’s18 actions, raw manifest/coverage bytes and all semantic assertions; this is only an index shift.|
|`:846–856`,`:873–874`|Replace both measured delta and absolute size/node guards explicitly using root-accepted measured M/N/C and margins within unchanged planning stops; correct accompanying comments.|
|`:857–861`|Total canonical authored assertions1539→1539+A, measured A≤446; derivative additions excluded and separately accounted.|
|New row-local/source guards in same files|Exact21-action bodies/assertions/captures, five-field Event tuples, Problem and receipt shapes, all header presence pins, final URL capture, controls and false flags.|

No unrelated receipt-pagination count18, original no-op15/223, fixture pin, prior semantic assertion IDs, catalog clause or runtime limit changes. Existing no-op control bounds are an explicit bookkeeping adaptation; their semantic mutations remain unchanged.

## Measurement gates and performance budget

Use actual dense analogues, not actions-only guesses: rollback9actions/59assertions is16319B/452nodes; set-mutation30actions/279assertions is71739B/2024nodes; no-op15actions/223assertions is81047B/2128nodes. Repeated literal group/complete-record assertion text dominates the new row; its21 actions, four Resources and two refusal partitions make a tiny incremental allowance implausible. PR114's original64KiB/220 estimate failed and was explicitly amended; do not claim it passed or repeat it as a reliable predictor.

Proposed **planning stop**, subject to council/root: at most twice the measured no-op insertion cost—manifest growth162106B, nodes4256, coverage21186B, new assertions446. This deliberately budgets two **normative spec partitions for arbitrary authorities**, plus an ordered comparator; it is not predicted consumption or a final committed growth guard. Absolute ceilings are650386B/17275nodes/91848coverage. No runtime loader cap changes: TX remains1MiB/100000nodes/depth128/container10000. No fixture segmentation required.

Both directions currently traverse the same `source=? OR target=?` incidence query (`transaction-store.ts:1148–1152`), evaluator failure and savepoint rollback. They are not independent implementation paths. An arbitrary authority or a future source/target regression can nevertheless fail asymmetrically. Root explicitly chooses both required-law direction partitions within the existing2× planning stop. A single-direction variant would remove the second refusal POST and its four after-state reads (21→16 actions); it could reduce authoring size, but neither the review's approximate13-action suggestion nor a1× size ceiling is measured or adopted. This tradeoff does not require new semantics or a further product choice.

Before installed testing, submit the actual normally formatted draft M/N/C/A, exact21-action assertion/capture/mirror ledger and22 parent raw-object hashes for root measurement acceptance. Separately submit each control derivative’s exact added assertion IDs/count, total assertion count, byte/node sizes and artifact digest; derivative-only positive checks are not canonical row assertions and do not enter canonical A or its coverage mirror. All derivatives must remain within the existing runtime/body/deadline limits and the stated absolute planning ceilings; any overrun requires disposition before execution. Committed measured margins must be declared and <=4096B/64nodes/1024coverage, with exact A and1539+A assertions, exact21/187 actions and unchanged parent no-op223. Retain independent absolute planning ceilings; shrink margins if necessary. Replace both old live growth and absolute guards visibly (current manifest−407227<=85147 / <=499387; nodes−10891<=2192 / <=13091; coverage−60069<=11615 / <=76453). Exceeding a planning stop requires an explicit amendment before execution, not dropping an oracle, compressed formatting or silent growth.

Add the new row to the actual2KiB authored POST/32KiB decoded-response filter. Two distinct bounds happen to be32 in the installed configuration: operation cardinality is `TRANSACTION_RUNTIME_LIMITS.maximumOperations=32` (`transaction-batch.ts:63`); finite Event item cardinality uses `receiptMaximumItems` (`transaction-store.ts:1575`), whose default32 comes through `transactionReceiptMaximum()` (`:217`) and is captured by runtime configuration (`transaction-runtime-configuration.ts:294`). Record the installed fixture/configuration binding proving no override. Four operations fit the first bound;15 Events and `next:null` depend on the second bound plus the existing byte bound. Do not treat these separately configurable limits as one cap. Keep63s positive-runtime/90s cohort/120s outer/5s exchange/10s prepare-cleanup. Current bec97 positive observer3.934s/19children is historical feasibility context only. Run the new positive + three new controls **and the three existing no-op controls once each**, because their aggregate bounds change: seven cohorts, at most20children/180exchanges each, **140children/1260exchanges maximum**. This is a targeted rerun of affected controls, not a broad historical cohort sweep. A failed control stops its own scenario at the failed action; later rows may still run. Report actual per-run ledger and cleanup. No clocks, races, restart or long retention wait are needed; wire/page sizes still require actual measured validation.

## Three meaningful installed controls and static guards

1. For d-incoming-survives only, swap its last two operations: delete r before delete a, retaining the update-b and delete-l prefix. The actual transaction now completes. In this derivative manifest only, retain the named intentionally failing expected-failed disposition assertion and add named positive assertions on that same actual exchange: HTTP200 and receipt schema; disposition `completed`; exact four-result count and order **updated b, deleted l, deleted r, deleted a**; literal result indices, resource kinds, IDs and types; updated b properties/title `incoming-survives-prefix` and metadata `{phase:"incoming-survives-prefix"}`; deleted identities/revisions against the independently validated seed captures; requiredPosition and effectPosition pos-2; and independently present current response epoch/view/position headers with current position pos-2. Require every added positive assertion ID to pass while the named expected-failed assertion fails. An expected-failed assertion failure or a200/schema pass alone is insufficient evidence of these completed facts.
2. Independently apply the symmetric swap to d-outgoing-survives. Add the corresponding derivative-only positive checks for **updated b, deleted r, deleted l, deleted a**, updated b title `outgoing-survives-prefix` and metadata `{phase:"outgoing-survives-prefix"}`, the appropriate seed-backed deleted identities, and required/effect/current pos-2. Again every added positive check must pass alongside the intentionally failing expected-failed disposition check. Each control uses a fresh isolated scenario and distinct derivative manifest, with unchanged action/route/key counts. These are lawful successful-order comparators, not400/invalid-input substitutes.

For controls1/2, the existing runner evaluates all assertions on the returned exchange before aborting the row (`runner.ts:1440–1470`). Independently authored request-body expectations must describe the swapped input; preserve the named intentional disposition failure, not an accidentally stale request-body oracle. Use that existing manifest assertion seam; no response interceptor, raw-body capture, helper, DSL or later GET is introduced. The original failed-receipt assertions may additionally fail on completed results, positions or their required absence; classify those expected correlated failures explicitly instead of requiring a singleton whole-row failure. Require the exact named positive checks above independently; do not infer their values from failure messages or the report’s redacted response summary. The row stops at the changed POST, so no later rollback observation is claimed for either derivative. Record each derivative’s assertions/counts separately from canonical A/1539+A and from built-in client checks. The canonical row remains21actions, each derivative retains the same21-action manifest, and all per-cohort maximum request/process bounds remain unchanged. The overall validation ledger separately includes the three affected old controls, for seven cohorts total.

3. **Wire-state rollback discriminator:** in a third distinct derivative remove only `delete a` from d-incoming-survives, yielding `[update b, delete l, delete r]`. Replace the authored request-body expectation with independently authored literal constants for the new three-operation input; do not derive expected values from the observed response or mechanically accept whatever request was sent. Audit and replace EVERY incompatible original failed-receipt oracle (failed disposition, Problem/detail members where incompatible, effect/results absence, required/current pos-1 and baseline metadata equality) with named positive assertions for actual HTTP200/schema, completed disposition, exactly three ordered results **updated b, deleted l, deleted r**, exact result identities/kinds/indices/types, b's `incoming-survives-prefix` properties/metadata, seed-backed deleted Link identities/revisions, and required/effect/current pos-2 with required context presence. ALL assertions on this POST must pass, allowing the existing next action d-incoming-survives-beads to execute. Keep that GET's canonical full seed-b binding and literal tuple metadata expectations unchanged. On actual200 complete Bead collection, both named `complete-b` and tuple-inventory assertions must fail because the served b truly changed; classify any additional revision/context failures honestly. Verify every preceding exchange passes, that the expected GET is reached, and that it is the row's final observed exchange. Do not claim d-incoming-survives-links or any later observation ran. No action/request/route/DSL is added.

Control3 is an actual wire-input mutation, unlike the earlier expectation-only proposal. Retain the latter only as a static weakening case: changing expected b metadata from{} to the prefix value must fail independent expected-content pins. Record control3's replaced/added/removed assertion IDs, net and total assertions, sizes and digest separately before execution. Controls1/2 retain their intentional failing POST disposition plus passing positives; control3 intentionally has no failing POST assertion, so it can reach the failing state GET.

Static guards independently pin complete new assertions/captures/request bodies, exact operation ordering and Problem index/code/type/status, full-record equalities, complete Event/group arrays/scalars, headers/schemas, final capture, and raw-object preservation. Include a weakening case that deletes the unchanged-Link full-record assertion and one that changes the Problem index while preserving IDs; both must fail independent expected-content pins. All controls keep claimEligiblefalse and distinct artifact identities. Do not conflate authored assertions with built-in receipt-client checks or capture completion.

## Done when and remaining uncertainty

After separate authorization: affected cumulative/app static checks, pinned Node24.16.0/pnpm11.20.0 type/build/format/lint/boundaries/Read-seal, exact source freeze, then genuinely fresh packed TX positive + three new controls + the three adapted existing no-op controls from that frozen source (seven cohorts). Record source/tree, tar/payload/contracts/configuration/routes/catalog/artifact/fixture digests, actual requests and required assertion/capture outcomes (including separately counted derivative-only positive assertions and the named expected failures), all child exits and no owned processes. No full workspace suite or repeat of every historical installed cohort absent a failure. Runtime bytes should remain identical; do not relabel historical packs/runs even if payload hash is equal.

Done when measured bounds and all old bytes are preserved, actual source-bound positive and controls discriminate the two operation-time failures and ordered success, every child is drained, and native+Claude/root dispositions are resolved. Public discovery/qualification stays closed. Actual final size/counts and current-collection/group wire sizes are unmeasured until implementation. Explicit measurement risks are exact pos-2 context on the deleted-Bead404, identity of collection Link items versus validated receipt Link records, and actual creation of two opposite Relationship Links. Existing source supports those oracles (404 observation: `authority-read.ts:797`, HTTP context: `transaction-http.ts:235–237`, preserved Problem response: `read-http.ts:10–39`), but no new installed observation is claimed. If an actual mismatch occurs, diagnose/fix it within authorized scope or stop for root disposition; there is no automatic presence-only weakening or erasure of the exactpos-2 oracle; any unexpected fixture/helper/runtime need or limit overrun stops that portion for root disposition. Planning performed no tests, builds, services, credentials, network, Git/source mutations or outward actions; only the requested successor plan and dispositions were produced. No new tracker or product choice.
