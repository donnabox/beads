# After no-op evidence: operation-time Bead deletion safety

Proposed next U3/T4 slice only. No implementation or claim promotion is authorized. Keep this independent of PR114's pending council/CI; root must assign its accepted successor/merged tree before authoring. The sole plan remains `bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md`.

## Recommendation and inspected basis

Add one partial required row, **`transactional.batch.deletion-safety`**. Use two opposite unowned Links between two Beads. Prove that removing only the outgoing Link cannot permit deletion while the incoming Link remains, and vice versa; earlier updates/Link deletions roll back, and a later Link-deletion suffix cannot rescue the failed batch. Then delete both Links before deleting the Bead and prove the ordered transaction succeeds.

This tests practical graph teardown and transaction ordering, including destructive-prefix rollback. It is distinct from the current no-op and completed-key-conflict rows. No product ruling or new runtime, schema, DSL or fixture configuration is required by this fixed visible/unowned partition.

Read-only basis: `/Users/dbox/repos/wt/vickie/vickie-noop-conformance`, frozen **bec97a7a00b8df663a62004e1ee640303d8527fa**, tree **f173c70f0806b72fa75075b1cdb1ce35d0c14bf0**, clean. Read the sole plan's live ledger/decision section, `transaction-next-unblocked-serving-audit.md`, actual catalog/spec/evaluator/store/HTTP test/runner paths, current artifacts and `transaction-noop-conformance/report.md`. Pending SSE/minimum/incident-view questions remain untouched; no new serving audit or public-discovery workaround is proposed.

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

Spec2303–2306 also says incident Links must be explicitly removed before the Bead. This is an **operation-time** test, not a check of the transaction's intended final graph. Existing runtime `resource-evaluator.ts:1023–1029` checks revision and then `tx.incidentLinks(id).length`; the transaction runs against staged state. `transaction-store.ts:1236–1247` rolls effects back and constructs a failed receipt. Batch error conversion preserves `incident-links-exist` (`transaction-batch.ts:507–552`); its failure is a receipt Problem with would-be status409, not an outer409 response (`bdp.md:5299–5315`). No new precedence is inferred.

Current installed `transactional.batch.rollback` fails a later revision guard after updates/creation, with no Link deletion. Existing set rows cover set effects/cardinality and deletions, but do not bind this required row or test both incident directions with only one of two opposite Links removed. In-process precedent includes incomplete-cascade refusal/rollback (`transaction-http.test.ts:1771–1808`), explicit Link-before-Bead success and tombstones (`transaction-store.test.ts:625–670`), and rollback of a preceding Link deletion on a later guard failure (`:671–696`). These are source support, not installed observations under the new catalog row.

Keep this row and all eleven predecessor rows `wholeContract:false`, overall `claimEligible:false`. It observes two local visible unowned bare-reference directions and ordered success, not hidden incidence/nondisclosure, owned Links, self-Links, pinned/external endpoints, concurrent incidence insertion, alias locators, retention/restart, or all optional configurations. No credit is borrowed from other rows. If successful,12 partial required rows leave238 required plus5 optional unimplemented; canonical250/255 selection is unchanged.

## Fixture, mutation inputs and independent oracles

Use unchanged716-byte `reference-transactional-v1` and the existing installed Task/Relationship descriptors. Relationship is unowned, properties permit title:string. Scenario setup is ordinary HTTP mutation, not a fixture/config expansion:

- Seed a and b with titles `a`,`b`, metadata `{}`.
- Seed l:a→b and r:b→a with titles `l`,`r`, metadata `{}`. All IDs are literal canonical Scope-local identities; no aliases, labels, pins, selectors or minimum headers.
- Incoming-only refusal batch: update b to title `incoming-prefix` and metadata `{phase:"incoming-prefix"}`; delete l; delete a; delete r. Expected failure at operationIndex2, while r remains incident. The later r deletion must not rescue the batch.
- Outgoing-only refusal batch: update b analogously to `outgoing-prefix`; delete r; delete a; delete l. Same failure index2, while l remains incident.
- Ordered success batch: delete l; delete r; delete a. Bead b remains exactly its original full record. All operations omit guards; no raw interpolation of captured revisions is needed.

Distinct literal idempotency keys for the four POSTs, same authenticated principal. Each failed batch must return actual HTTP200 with `mutationReceipt` schema, status `failed`, detail `available`, requiredPosition pos-1; no effectPosition/results/next. Problem is exactly code `incident-links-exist`, type `https://github.com/gastownhall/bdp/problems/conflict`, status409, retry `after-state-change`, operationIndex2. No unsupported failure-envelope or direct-status inference.

The seed receipt's actual private GET supplies full a,b,l,r records, revisions needed for deleted identities/Events, and seed transaction token. First validate the raw seed receipt's literal IDs/types/properties/metadata/endpoints and exact four-created result order. Raw HTTP captures remain URLs only; existing client receipt capture handles records/tokens. The final successful receipt is likewise read through the current private receipt helper to capture its transaction token after literal deleted-result validation. No raw body/array capture or new helper.

Before success, complete current collections must contain exactly a,b and l,r in canonical order, `next:null`, and every complete item must equal its independently validated seed capture, including revisions/provenance. Explicit literal tuple inventories also retain IDs/properties/metadata, so a same-ID metadata drift control can check a substantive rollback fact. Strict `currentBeadCollection`/`currentLinkCollection` schemas apply. Failed prefixes must leave full b unchanged and restore the deleted Link; surviving suffix Link and Bead a must also remain.

After success, Bead collection contains only original b, Link collection is empty, both `next:null`; GET a is404. Successful receipt contains exactly three deleted entries at indices0/1/2: l,r,a, correct resourceKind and canonical id/type/final revision matching the seed captures, with required/effect pos-2. No opaque revision guessed.

### Scope Events and change groups

Seed has **eight** ordered Events: created a0, created b1, created l2, linked l/source3, linked l/target4, created r5, linked r/source6, linked r/target7. Four Resource upserts, eventCount8. Both failed batches leave that exact Scope inventory and pos-1 unchanged; no prefix update/deleted/unlinked fact escapes.

Ordered successful deletion adds seven facts: deleted l0, unlinked l/source1, unlinked l/target2, deleted r3, unlinked r/source4, unlinked r/target5, deleted a6. Final Scope Event page contains15 facts, `next:null`; check complete ordered type/source/subject/ordinal tuples, endpoint-specific unlink data, canonical identities, deletion revisions and final transaction token against the receipt. This fits the existing32-item finite page and avoids pending incident-source projection semantics.

Existing finite `changes/?after=genesis` serves one group/page. At seed and both failures require exactly seed group pos-1/previouspos-0, headpos-1, next:null, seed transaction, four captured upserts, eight exact Events and eventCount8. After real successful deletion, same route still returns the retained seed group, headpos-2 and a non-null next URL. Validate the latter with an ordinary scoped-URL `/next` capture, never fetch it. Do not falsely expect two groups/page or claim observation of the deletion group's tombstone payload: its effect is independently checked via receipt/current state/Events, while the final genesis read demonstrates the group/position advance. This is the same already-reviewed capture mechanism, not new authority.

## Exact21-action ledger

|#|ID|Action / decisive evidence|
|---:|---|---|
|1|d-seed|POST batch creates a,b,l,r, pos-1; validate four created entries, capture root URL.|
|2|d-seed-facts|Receipt-client GET; validate and capture four records/revisions and transaction token.|
|3|d-seed-events|GET events/; exact eight seed facts and no next.|
|4|d-seed-groups|GET changes/?after=genesis; complete seed group/headpos-1/next:null.|
|5|d-incoming|POST incomplete teardown with l deletion; failed receipt/problem/index2, pos-1.|
|6|d-incoming-beads|GET beads/?limit=32; full a,b unchanged, literal b metadata{}, no next.|
|7|d-incoming-links|GET links/?limit=32; full l,r restored/unchanged, no next.|
|8|d-incoming-events|GET events/; exact original eight, pos-1.|
|9|d-incoming-groups|GET changes after genesis; exact original group, headpos-1/next:null.|
|10|d-outgoing|POST incomplete teardown with r deletion; failed receipt/problem/index2, pos-1.|
|11|d-outgoing-beads|GET complete Bead collection; full a,b unchanged.|
|12|d-outgoing-links|GET complete Link collection; full l,r unchanged.|
|13|d-outgoing-events|GET events/; exact original eight.|
|14|d-outgoing-groups|GET genesis changes; exact original group/headpos-1/next:null.|
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

Preserve all eleven parent manifest/coverage raw row objects and the fixture byte-for-byte; preserve all existing assertion IDs/pins/controls, including no-op223 assertions and successful final URL capture. Add no-op to the current ten-original-object guard:

|Raw no-op object|Bytes|SHA256|
|---|---:|---|
|manifest|81047|`9307b230eb053131e9127cbbb5e6285305b79d6794edd22f5432a07fc74d9faa`|
|coverage|10587|`383de4e0cb9b3bad310b9db7ff9009cb59fac55966f0309bc71c5d007bea6c1c`|

Parent manifest488280B/13019nodes/SHA0358cbf7416321bb911e990259236f89bc2720cbc29d037581f542080d9f32ab; coverage70662B/SHA8a770b4371689a3d9d5f84ecfc7a1d9b72715c81e617cf0ea5ae04e55c66098e. Fixture716B/SHA3a7899ced771d4880066d98b0000c53f801ece53458502cc4f3bc641146c9a85. These are inspected candidate bytes, not approval of PR114 or a promise they survive its review. Rebase/integration drift requires explicit reconciliation, not refreshing pins blindly.

## Measurement gates and performance budget

Use actual dense analogues, not actions-only guesses: rollback9actions/59assertions is16319B/452nodes; set-mutation30actions/279assertions is71739B/2024nodes; no-op15actions/223assertions is81047B/2128nodes. Repeated literal group/complete-record assertion text dominates the new row; its21 actions, four Resources and two refusal partitions make a tiny incremental allowance implausible. PR114's original64KiB/220 estimate failed and was explicitly amended; do not claim it passed or repeat it as a reliable predictor.

Proposed **planning stop**, subject to council/root: at most twice the measured no-op insertion cost—manifest growth162106B, nodes4256, coverage21186B, new assertions446. This deliberately budgets two destructive-refusal direction partitions plus an ordered comparator; it is not predicted consumption or a final committed growth guard. Absolute ceilings are650386B/17275nodes/91848coverage. No runtime loader cap changes: TX remains1MiB/100000nodes/depth128/container10000. No fixture segmentation required.

Before installed testing, submit the actual normally formatted draft M/N/C/A, exact21-action assertion/capture/mirror ledger and22 parent raw-object hashes for root measurement acceptance. Committed measured margins must be declared and <=4096B/64nodes/1024coverage, with exact A and1539+A assertions, exact21/187 actions and unchanged parent no-op223. Retain independent absolute planning ceilings; shrink margins if necessary. Replace both old live growth and absolute guards visibly (current manifest−407227<=85147 / <=499387; nodes−10891<=2192 / <=13091; coverage−60069<=11615 / <=76453). Exceeding a planning stop requires an explicit amendment before execution, not dropping an oracle, compressed formatting or silent growth.

Add the new row to the actual2KiB authored POST/32KiB decoded-response filter. Maximum4operations per new batch and15 finite Events fit installed32 limits. Keep63s positive-runtime/90s cohort/120s outer/5s exchange/10s prepare-cleanup. Current bec97 positive observer3.934s/19children is historical feasibility context only. New positive + three controls: at most20children/180exchanges each,80children/720exchanges across four cohorts. A failed control stops its own scenario at the failed action; later rows may still run. Report actual per-run ledger and cleanup. No clocks, races, restart or long retention wait are needed; wire/page sizes still require actual measured validation.

## Three meaningful installed controls and static guards

1. For d-incoming only, swap its last two operations: delete r before delete a, retaining the update-b and delete-l prefix. The actual transaction now completes. Require its actual HTTP200/schema-valid wire and failure of the named expected-failed disposition oracle, with coherent completed results/currentpos-2. This is not a400/invalid-input substitute. The row stops at this action; do not claim later rollback observations ran in this mutant.
2. Independently do the symmetric swap for d-outgoing. It also completes, falsifying the expected failure while testing the other incidence direction. Same action/route/key count in a fresh isolated scenario, distinct derivative manifest.
3. Leave wire inputs unchanged; alter only the d-incoming-beads literal inventory expectation for b's metadata from{} to the actual tentative prefix value `{phase:"incoming-prefix"}`, preserving ID, type/order/count and other fields. Require actual200 complete collection and the named inventory assertion failure. This demonstrates a same-identity state oracle catches the rolled-back prefix instead of merely checking IDs/status. It is an expectation sensitivity control, not an injected production rollback failure.

Static guards independently pin complete new assertions/captures/request bodies, exact operation ordering and Problem index/code/type/status, full-record equalities, complete Event/group arrays/scalars, headers/schemas, final capture, and raw-object preservation. Include a weakening case that deletes the Link-restoration full-record assertion and one that changes the Problem index while preserving IDs; both must fail independent expected-content pins. All controls keep claimEligiblefalse and distinct artifact identities. Do not conflate authored assertions with built-in receipt-client checks or capture completion.

## Done when and remaining uncertainty

After separate authorization: affected cumulative/app static checks, pinned Node24.16.0/pnpm11.20.0 type/build/format/lint/boundaries/Read-seal, exact source freeze, then genuinely fresh packed TX positive + three controls from that frozen source. Record source/tree, tar/payload/contracts/configuration/routes/catalog/artifact/fixture digests, actual requests and required assertion/capture outcomes, all child exits and no owned processes. No full workspace suite or repeat of every historical installed cohort absent a failure. Runtime bytes should remain identical; do not relabel historical packs/runs even if payload hash is equal.

Done when measured bounds and all old bytes are preserved, actual source-bound positive and controls discriminate the two operation-time failures and ordered success, every child is drained, and native+Claude/root dispositions are resolved. Public discovery/qualification stays closed. Actual final size/counts and current-collection/group wire sizes are unmeasured until implementation; any unexpected fixture/helper/runtime need or limit overrun stops that portion for root disposition. Planning performed no tests, builds, services, credentials, network, Git/source mutations or outward actions; only this recommendation was written. No new tracker or product choice.
