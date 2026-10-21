# Next bounded Transactional conformance slice: matching updates with no effect

Status: revised v2 proposed plan for council and root acceptance; no implementation authorized by this document. This is a subordinate execution plan under the sole design plan’s U3/T4 lanes, not a competing runtime roadmap.

## Recommendation and exact basis

Add one **partial** installed row, `transactional.batch.no-op-entries`, to the existing cumulative Transactional evidence. Exercise actual matching Bead and unowned Link updates whose properties and metadata remain equal, then an effectful update comparator. Prove completed ordered results at the retained full records, unchanged current position, no `effectPosition`, no additional Scope Events, and no additional change group.

Inspected frozen source `/Users/dbox/repos/wt/vickie/vickie-current-minimum-receiving`, commit `5d854a1da21a377f03c2d4cd523b10f1e7659cea`, tree `75a9db868de253a92fa4e08495d86ec0f6ed9ac6`. Its pending external review/CI is not treated as completed. Root must assign a reviewed implementation parent and reconcile any relevant drift before authoring. All citations below refer to this frozen checkout.

The serving audit `transaction-next-unblocked-serving-audit.md` leaves explicit route decisions pending. This proposal uses existing finite JSON routes, with no minimum headers, SSE reconnect choice, incident-Event projection, discovery response, or retained-Snapshot interpretation. It adds no runtime mechanism or protocol law. The current private configured observation target remains private; it does not emit a fabricated discovery/profile assertion. No public admission or qualification is proposed.

### Selected obligation and remaining coverage

Catalog `packages/conformance/catalog/transactional-v1.json:246–263` selects these exact clauses:

1. `docs/specs/bdp.md#change-groups-and-replication`: “A failed or admitted no-effect mutation produces no group and no new position. Its receipt reports the current `requiredPosition` and omits `effectPosition`.”
2. `docs/specs/bdp.md#mutation-receipt-responses`: “a transaction all of whose operations are no-ops is an admitted no-effect mutation: it completes, produces no group, and its receipt omits `effectPosition`.”

The surrounding current text at `bdp.md:5141–5157` requires an `updated` entry with the complete postimage at the retained revision. Use this as a concrete independent oracle; do not alter or narrow the catalog’s selected text.

Existing installed `transactional.set.zero-match` checks zero matched set operations. It does not execute matching Resource updates with equal postimages. `transaction-store.test.ts:507–539` already proves all-no-op Bead/Link behavior in process, but cannot substitute for target-bound installed observations. Runtime `transaction-batch.ts:710–711` suppresses transition recording for unchanged evaluator results; `transaction-store.ts:1264–1281` inserts a group/advances position only for an effectful result and constructs the no-effect receipt. This is testing existing behavior, not implementing speculative semantics.

Keep `wholeContract:false` for the new row and all ten existing rows, and `claimEligible:false`. The new row witnesses the all-no-op success partition on one fixed all-visible installation. It does **not** cover the selected first clause’s failed-mutation partition, every other no-effect category, guarded races, owned-Link source revisions, mixed effect/no-effect transactions, principal/view changes, expiry, or restart retention. Existing observations in other rows are preserved, not borrowed to promote this row. Eleven partial required rows would leave239 required plus five optional rows unimplemented; the canonical250/255 selection remains unchanged.

## Existing seams and fixture

Use the unchanged716-byte `reference-transactional-v1` fixture and existing Task/Relationship descriptors in `apps/bdptest/src/transactional-matrix.test.ts:46–77`. Relationship is unowned. Its properties schema accepts `title:string`; metadata can use ordinary object values. Seed exactly two Beads `beads/a`, `beads/b` and one Link `links/l`, with canonical local endpoints a→b, titles `a`, `b`, `l`, and explicit metadata `{}`. No new Type, fixture binding, capability, server option, policy, or database access.

The two update operations target a and l. Their no-op bodies each replace `/title` with its existing string and replace the metadata root with `{}`. The later comparator uses the same two update operation forms but changes titles to `a-changed` / `l-changed` and metadata to `{tag:"changed"}`. No `expectedRevision` is injected into raw text: this is the unconditional no-op partition, not revision-guard coverage.

Raw HTTP capture permits URLs only (`packages/conformance/src/runner.ts:2467–2485`). Capture each POST’s receipt `/id` through that existing path. Use the existing authenticated `transactional.receipt-read` action to capture full Resource records, revision strings and transaction tokens from the actually validated receipt. No raw body/array/UUID capture, new capture grammar, or public helper is needed. Before any later equality, independently assert the seed’s complete result order, literal IDs/types/properties/metadata/endpoints and current schema. Preserve the existing receipt source/principal/session validation.

Add exactly one literal exact-HTTP route: `GET https://reference.invalid/acme/changes/?after=genesis`. Its parser and HTTP path already exist (`transaction-changefeed.ts:87–108`, `transaction-http.ts:427–451`); explicitly send `Accept: application/json`. The current owner serves **one group per page** (`transaction-store.ts:1683–1710`). Read this route only while the head is pos-1, before and after the no-op. Both reads then have exactly one group and `next:null`. Do not claim a later pos-2 read contains two groups or add a dynamic continuation grant. The post-comparator finite Scope Event read can hold all seven facts under the existing32-item page bound (`transaction-store.ts:1547–1589`).

## Exact new action ledger

One isolated scenario session, no reopen. Every action makes one HTTP exchange, including the three private receipt client actions. Names below are proposed stable action IDs.

| # | Action | Family / request | Required discriminator |
|---:|---|---|---|
|1|`n-seed`|raw POST batch, key `noop-seed`|Completed, three `created` results in a,b,l order; exact literal Resource data; required/effect pos-1; available detail and `next:null`; capture receipt URL.|
|2|`n-seed-facts`|client receipt GET|Actual200/body present/context; independently asserted seed facts; capture a,b,l full records, a/l revisions, seed transaction token.|
|3|`n-seed-groups`|raw GET changes after genesis|Current `changefeedPage` schema, head pos-1, exactly one pos-1 group, previous pos-0, `projectionAdvance:false`, empty erasures, transaction equals seed token, exact three upsert identities and their captured full records, five Events and explicit `eventCount:5`, `next:null`.|
|4|`n-noop`|raw POST batch, key `noop-equal`|Completed, exactly two ordered `updated` entries at indices0/1; each full Resource equals its earlier seed capture; required pos-1; `effectPosition` absent; `next:null`; capture receipt URL.|
|5|`n-noop-facts`|client receipt GET|Same completed/no-effect/result facts on actual retained root; both full records equal seed captures, position pos-1. No repeated execution/restart claim.|
|6|`n-noop-beads`|raw GET beads/?limit=32|Strict current collection schema; exactly a,b in order, each entire item equals its seed capture, `next:null`; current pos-1.|
|7|`n-noop-link`|raw GET links/l|Strict current Link schema; entire record equals seed l, current pos-1.|
|8|`n-noop-events`|raw GET events/|Strict Event page; exactly the five seed facts in order, `next:null`, current pos-1; no updated fact.|
|9|`n-noop-groups`|raw GET changes after genesis|Repeat the independent singleton-group/seed transaction/postimages/Events and explicit `eventCount:5` oracle from action3, `observedHeadPosition:pos-1` and `next:null`. Do not compare merely an opaque checkpoint or infer no group from Event count.|
|10|`n-changed`|raw POST batch, key `noop-changed`|Same update forms with changed values; completed ordered updated a/l records containing both intended property and metadata changes; required/effect pos-2; `next:null`; capture receipt URL.|
|11|`n-changed-facts`|client receipt GET|Validate changed literal data before capturing full a/l records, revisions and transaction token; context pos-2.|
|12|`n-changed-beads`|raw GET beads/?limit=32|Exactly a,b, `next:null`; a equals changed capture, b still equals seed b.|
|13|`n-changed-link`|raw GET links/l|Full l equals changed capture, including metadata and endpoints.|
|14|`n-changed-events`|raw GET events/|Exactly seven ordered facts, `next:null`, current pos-2: original five followed by updated a and updated l, ordinals0/1 in the new transaction. Check previous/new revisions against their receipt captures and exact property/metadata deltas.|

The seed Event sequence is **five**, not three: created a ordinal0, created b ordinal1, created l ordinal2, linked l/source ordinal3, linked l/target ordinal4 (`transaction-batch.ts:739–776`). Scope source and subjects are literal; endpoints on the two linked facts are independently distinguished. Group changes contain three Resource upserts, not five. The no-op must preserve this exact inventory. Later updates add two facts, not linked/unlinked incidence events. All queries are Scope-level; the pending incident metadata membership choice is irrelevant.

Use existing `json-array-tuples` with an exact ordered expected array to prove full cardinality; `json-pointer` with `exists:true` and `equalsBinding` compares complete independently validated records; `exists:false` proves effectPosition absence. No invented `notEqualsBinding` operator. Use current schemas (`mutationReceipt`, `currentBeadCollection`, `currentLinkRecord`, `eventPage`, `changefeedPage`) on actual raw response bodies. The effectful comparator proves updates are honored through intended fields, position and Events; it is not a separate general revision-law claim.

Each raw action retains request-authored headers/body assertions. Every contextual response has explicit epoch/view/position presence, and exact position where known. Put presence assertions before any response-metadata equality. Compare same-status200 responses only; use seed as the pos-1 baseline and changed POST as pos-2 baseline. The private client envelope uses `/headers/...`; do not assume raw report headers expose its private context. Existing DSL validation requires prior captures and forbids normalized equalsBinding comparisons (`executable-manifest.ts:1988–1992,2534–2539`).

## Ownership, preservation and budgets requiring acceptance

Exactly four source files:

1. `packages/conformance/test-support/transactional/manifest-v1.json` — insert the one row after rollback and before set.mutation, in catalog order.
2. `packages/conformance/test-support/transactional/coverage-v1.json` — one false partial row, honest reason including “Whole contract remains unqualified”, exhaustive assertion mirror.
3. `packages/conformance/src/transactional-cumulative.test.ts` — ten-parent-object preservation pins, independent new assertion/capture/selected-clause pins, exact ledger and approved growth guards, weakening controls.
4. `apps/bdptest/src/transactional-matrix.test.ts` — new literal route, eleven-row ledger, exact new routing digest, new response bounds and installed counterfactuals.

No catalog/spec/schema/fixture/runtime/DSL/exports/sealed Read evidence changes. Preserve all ten previous manifest and coverage raw objects byte-for-byte, not merely their IDs. The existing nine-object pin lists must gain the completed-key-conflict row:

| Artifact raw row | Bytes | SHA256 |
|---|---:|---|
|manifest conflict|29772|`ba93aea884ae462533d02695642f03a2ca854f5adabeb41867254216dabb2106`|
|coverage conflict|5188|`de9b758f8ac4df7db5597b1eb2b402a116720becc30a41e7198f78f244358ec0`|

Current whole files are manifest407227B/10891 value nodes/SHA `7f55695401b815534f17fc69dc924ee0c0920fb6c9175529709a8e60afda5786` and coverage60069B/SHA `4c84318e57c74a928edcaead5490bc647a67f2e3ba79872f1babe7ac3b6916ec`. The actual fixture hash remains `3a7899ced771d4880066d98b0000c53f801ece53458502cc4f3bc641146c9a85`.

**Explicit authoring-budget amendment, not a runtime-cap increase:** current prior-slice conjunctive guards leave only2990 manifest bytes,391 value nodes and2998 coverage bytes. They cannot accommodate this row. Council/root must accept the following proposed ceilings before implementation. They are conservative planning bounds, not a measured candidate: the existing17-action zero-match row is36055B/1028nodes/142assertions; this14-action row adds Link records and finite-group oracles while omitting reopen. Allow at most220 new assertions,64KiB new manifest bytes,2200 new value nodes and16KiB coverage growth. Preserve normal formatting; do not compress to evade counts. Before tests, measure the actual candidate and stop on overrun rather than dropping an oracle or raising a fence automatically.

| Live guard at transactional-cumulative.test.ts:835–852 | Proposed accepted successor |
|---|---|
|151 total actions|165 exactly; new row14 exactly|
|manifest−377449<=32768 and manifest<=413696|manifest−407227<=65536 and manifest<=472763|
|coverage−54875<=8192 and coverage<=65536|coverage−60069<=16384 and coverage<=76453|
|nodes−10082<=1200 and nodes<=11300|nodes−10891<=2200 and nodes<=13091|

Both sides of each existing conjunctive guard must be reconciled visibly; preserve historical values in reports, not as impossible live inequalities. Pin all220-or-fewer new assertions by meaningful expected parameters, not just generated IDs. Retain all old per-row byte limits and context assertions. Existing TX loader1MiB/depth128/100000nodes/container10000 limits are unchanged (`artifact-bundle.ts:17–23,159–164`); the4MiB exception belongs only to RU. No segmentation or loader change is needed.

The new row has11 raw HTTP actions (three POST/eight GET), three receipt-client GETs and zero lifecycle actions. Total artifact actions165 =112 raw HTTP +46 receipt-client GETs +seven reopen actions; **158 actual exchanges**, versus parent144. Positive sessions18/children19 including bootstrap, all closed. Existing13 reopened receipt observations are unchanged. Exact response/action ledgers must distinguish network exchanges from reopen actions and bootstrap launches.

New POSTs must be <=2048 authored bytes and all new decoded responses <=32768B, borrowing the existing M/Z wire bounds. The exact HTTP allowlist becomes11 routes: the existing ten plus changes/?after=genesis, inserted at its first occurrence after the existing events/ route. Independently recompute the exactConfigurationDigest from the actual canonical full configuration; do not reuse the parent `63421454…` digest or manufacture an expected value from the changed report. Record both canonical preimage and digest. Runtime/fixture configuration and contract identities are measured afresh rather than presumed equal.

Keep63s positive runtime,90s cohort cancellation,120s outer test,5s exchange and10s prepare/cleanup bounds. Parent PR110 integration’s actual positive3.605s is feasibility context only, not new-source evidence or a guarantee. One positive and three new controls are separate bounded cohorts; at most19 children/158 exchanges per cohort, at most76 children/632 exchanges across those four runs. Failed actions may reduce actual counts; report actual outcomes and cleanup, not an invented identical count. No full workspace suite or new broad installed cohort is needed by default.

## Counterfactuals and verification

Use the existing source-only `observe({manifest(...)})` control seam; do not replace transport responses or alter a production package:

1. In the no-op POST only, change a’s `/title` value to a different valid string. Require an actual completed200 wire exchange and failure of the named `n-noop:bead-retained-record` oracle. The legitimate result is now effectful, so other no-effect/state/Event assertions may also fail. Do not require a singleton row failure or call a valid200 an invalid disposition.
2. Independently change only l’s metadata replacement to `{tag:"different"}`. Require actual200 and the named `n-noop:link-retained-record` failure. This discriminates the Link/metadata side without a grammar/schema failure.
3. Leave requests unchanged. In the same third control, change the `n-noop-groups` expected group-position tuple from `[["pos-1"]]` to `[["pos-1"],["pos-2"]]` AND its expected `observedHeadPosition` from `pos-1` to `pos-2`. Require an actual valid200 finite page and both named assertion failures: group inventory and observed head. The two-group tuple is a live cardinality assertion check only: this owner always serves one group per page, so it cannot itself prove absence of a later group. The positive head=pos-1 and next=null oracles, pinned independently in static expected-content guards, establish that distinction. Do not add a cohort/action/route, change pagination or infer no group from receipt/Event counts alone.

All control reports must show actual `claimEligible:false`, distinct manifest digests and completed authored wire where specified. Independently pin exact baseline record comparisons, no-effect absence, required/current positions, complete result/group/Event cardinality, explicit group eventCount=5 in both group reads, changefeed next:null and observedHeadPosition=pos-1, schema refs, mandatory context presence and captures. Local static weakening controls remove one full-record equality, the effectPosition-absence assertion, one context-presence assertion, and alter a complete group tuple while keeping IDs: each must fail its independent expected-content guard. Preserve all original tests, including the five conflict controls; run relevant affected tests without automatically repeating every historical packed control.

After authoring and measured preflight, use pinned Node24.16.0/pnpm11.20.0. Run the affected cumulative artifact/static tests, typecheck/build, format/lint/boundaries and Read evidence verification. Freeze a clean source commit before a fresh actual package/install. Run the existing packed TX positive and these three controls against that exact installation; archive exact commands/exits, source/tree, tar/payload/contracts/configuration/route/fixture/artifact digests, action/assertion mappings, child launch/exit ledger and final process cleanup. Do not relabel the existing5d854/8d238 packages or reports. The configured target may prove these selected operations while still lacking discovery/public admission; that limitation must remain in the report.

Done when all four changed files remain within accepted bounds, all ten parent raw objects/fixture/runtime bytes are preserved, the new partial row passes its actual positive and three named discriminators, all required oracles have live successful assertion IDs, source-bound reports and cleanup are complete, and independent council/root dispositions are resolved. A new required helper, route beyond the single finite changes target, unexpected runtime defect, body/capture limitation or budget overrun is a stop-and-report condition, not implicit expansion. Public/whole-contract qualification remains a separate unfinished objective.

## Audit performed for this proposal

Read-only source/catalog/DSL/runtime/fixture and prior accepted evidence inspection, plus local JSON counting and cryptographic file/raw-object hashing. No test, build, service, network, credential, Git mutation or source edit was performed. Only this external plan was written. Initial inspection verified clean5d854/tree75a9. At final read-only verification the shared checkout had advanced independently to clean `c534fcab840ef2e3753ab179feae10d3ce6431e5`, tree `48537438975a4c56ee1a5bb6056013bbddd38e03`; its entire delta from5d854 is72 added lines in `packages/server/src/transaction-service.test.ts`. No file used for this conformance plan or quoted source seam changed. This plan remains based on5d854 and does not review or claim acceptance of that separate test repair. The plan deliberately excludes a final pos-2 changefeed read because the actual owner serves one group per page; existing receipt-only dynamic grants must not be repurposed for a changefeed continuation. No product ruling is needed for the described fixed all-no-op/effectful partitions; the proposed artifact-authoring budget is an explicit review gate.

## V2 native review dispositions

Original plan SHA2561ce661b78ee37a2c22a27239af6ea318c1942149dbd828d0b9d64de6dcce4127 remains frozen for the external review. Native P3 findings are both accepted: the third existing control now perturbs group count AND observed head and requires both exact failures, with the count claim limited to cardinality; both positive group reads explicitly assert eventCount5 and pin that scalar independently. No new request/action/cohort/route or cap increase. These assertions must fit the proposed220-assertion authoring ceiling; actual measurement remains required. External findings/root acceptance remain pending.
