# Next Transactional slice: completed-key conflict evidence — version 2

Revision basis: original c2a6a8e8 is preserved unchanged. This version dispositions all seven numbered external findings (despite the review introduction saying six) and the native key-count correction; source remains main627/0359.

Status: supporting proposal for root/native/Claude review; no implementation authorized by this document. Sole authority remains the fork `docs/architecture/bdp-client-v2/design.md`, Autonomous profile completion T2/T4. This proposal adds one partial installed observation, not a profile claim or a new protocol rule.

## Source basis and recommendation

Verified main commit `627b560c0daf1a76b292615718c86670255ac9af` has tree `6a5337cf9548cfe340593a3858213593e4aca1d9`. Local frozen checkout `/Users/dbox/repos/wt/vickie/vickie-receipt-page-review-repair` is clean at `0359f1cae3f36f600b1b375ea276ec7d60bf689b`, whose tree is identical. All source inspections below therefore describe exact main627 contents; no network refresh, source edit, build, test, server, or reviewer process was started for this plan.

Recommend one new partial canonical row: `transactional.idempotency.conflict`. Exercise an already completed batch key, present a genuinely different valid mutation under that key, prove direct409 and no graph/Event changes, then retrieve the original exact outcome. Repeat conflict and replay after ordinary same-epoch restart. This is a small existing-contract gap with a strong independent oracle and no new harness capability.

Do not combine cross-carrier, concurrent join, duration, expiry, erasure, Types or discovery work. Cross-carrier is valuable but would require expanding current fixture/route authority to a singleton or sequence and defining sequence projection witnesses. Receipt page-restart already has substantial incidental evidence, but claiming its complete row also requires lifecycle/projection branches not proved by the current fixed-view fixture. Validator omission largely has incidental assertions already; simply assigning a new row to those assertions would add less new behavioral coverage than the proposed completed-key conflict witness.

## Existing evidence and actual missing coverage

Main627 has nine authored TX rows, all `wholeContract:false`; 139 actions; 250 required plus five optional selected catalog rows. Existing `transactional.idempotency.failed-retained` has eleven actions. Its `conflict` action reuses an earlier failed key and asserts409 plus exact request metadata, but does not assert the direct problem's code/schema or restart a completed key. Preserve that row byte-for-byte.

Owner and HTTP behavior already exists:

- `packages/server/src/transaction-store.ts:1030–1076` compares retained key identities and returns existing/conflict without a second execution. `:1372–1393` performs comparison/admission inside the owner transaction and applies receipt negotiation only to an established existing result.
- `packages/server/src/transaction-submission.ts:100–122` uses authoritative observed admission and maps conflict to409/idempotency-conflict.
- `packages/server/src/transaction-http.ts:345–367` maps that result to the direct conflict problem, independently of a receipt response.
- `packages/server/src/transaction-http.test.ts:228–248` already checks conflict-before-negotiation at the runtime seam. `transaction-store.test.ts` covers retained identity and reopen. These are implementation evidence, not the missing installed canonical row.
- `transactional-observation.ts:303–306` validates a root read against its exact originally observed POST body. The existing private reader permits JSON body/Resource captures; ordinary HTTP captures remain URL-only.
- The installed app already provides batch POST, `beads/` collection GET and Scope `events/` GET, source-bound credentials, current-session routing, and `transactional.reopen`. No fixture or route additions are needed.

This proposal addresses an installed evidence gap. It does not assume a runtime defect or authorize runtime repair if a new test fails.

## Canonical selected laws

Catalog: `packages/conformance/catalog/transactional-v1.json`, canonical index44, id `transactional.idempotency.conflict`, requiredProfile transactional, kind normative. Pin both exact requirement objects, not just matching an anchor substring:

1. `docs/specs/bdp.md#mutation-transactions`: “Reusing the key for a different semantic request is an idempotency conflict.” Current normative location3888–3889.
2. Same source/anchor: “A request with the same key and a different normalized request is refused before admission with `409` `idempotency-conflict` for as long as the key is bound”. Current location3931–3936 extends this to the rest of the epoch for completed outcomes, even after detail expiry; failed receipt retention is different.

Supporting existing law, without adding another credited row:

- `bdp.md:3875–3880`: key identity is canonical Scope + epoch + authenticated principal, with normalized semantic request.
- `:3929–3939`: equal later requests retrieve the retained receipt without re-execution; epoch changes create a new namespace, authorization-view changes do not.
- `#batch-operation-target`, `:4507–4532`: comparison precedes new admission controls/response negotiation; different semantics produce409 and leave the earlier outcome unaffected. HTTP-native406 applies to unacceptable negotiation in the appropriate later branch. Direct refusals do not create a new receipt.
- `#read-update-problem-details` table at3421, inherited by TX, plus `#transactional-problem-details` around5301: idempotency-conflict uses409, conflict problem type and retry never. Validate actual body with `#/$defs/transactionalProblem`.

The first conflicting presentation uses a valid JSON request and valid authenticated mutating principal but `Accept: text/plain`. Expect409 problem, not406. The second presentation after reopen uses Accept application/json. This is existing prescribed precedence, not a negative-Selector or malformed-body phase ruling.

## One twelve-action row

Insert at authored canonical position9 of10 (zero-based8), immediately after receipt.pagination and before failed-retained. Promote pagination to a pinned original by adding its exact parent manifest and coverage raw-span pins, then preserve all nine old objects per artifact. The existing guard currently covers eight, not nine. Every new action belongs only to the new conflict row; no new catalog entry or whole-row flag.

Use one ordinary unowned Task Bead `beads/a` with properties `{title:"original"}` and metadata `{tag:"conflict"}`. Changed request is a valid create of `beads/b` with the same Type/properties/metadata but the original key. An explicit different unused identity makes erroneous execution visibly create a second Bead; it cannot hide behind identity-already-taken for a. Both requests are one-operation batch requests, and each body uses fixed source literals. One original key, e.g. `completed-conflict`, reused byte-exactly for all five presentations: three original-semantic POSTs (c-commit, c-retry, c-retry-reopened) and two conflicting POSTs (c-conflict, c-conflict-reopened); no `@name`, selectors, pins or exotic numeric semantics.

| # | Action | Actual operation and essential oracle |
|---|---|---|
| 1 | `c-commit` | POST batch creates a;200/completed receipt schema, one created entry, exact id/Type/properties/metadata, effectPosition exists, next:null. Capture only root URL `c-root` from actual `/id`. |
| 2 | `c-original` | Private GET c-root;200/complete body and current-context presence. Capture whole `/body` as `c-body` and whole `/body/results/0/resource` as `c-record`. Require literal created record and one-entry shape. |
| 3 | `c-conflict` | POST valid create-b under original key, Accept text/plain. Exact raw request observations;409, application/problem+json, transactionalProblem schema, conflict type, code idempotency-conflict, retry never. No captures. |
| 4 | `c-state` | GET beads/?limit=32;200/`#/$defs/currentBeadCollection`,next:null; exact ordered one-record tuple with literal a/Type/properties/metadata; `/items/0` equals private c-record. Thus b is absent and a including revision/provenance remains unchanged. |
| 5 | `c-events` | GET events/;200/eventPage schema,next:null; exact one Event tuple created/a/ordinal0. Current epoch/view/position equals c-commit's response, with explicit presence assertions on both comparison sides. |
| 6 | `c-retry` | POST original create-a bytes and original key, Accept application/json.200/completed schema; whole response via json-pointer `""` equals captured c-body, and id equals c-root. No new capture or private route grant. A mistaken re-execution would fail because a already exists. |
| 7 | `c-reopen` | Existing private transactional.reopen; require reopened:true. Same persisted directory/config/artifacts, create:false through existing launcher. |
| 8 | `c-conflict-reopened` | POST create-b under original key after reopen, Accept application/json. Same409 problem oracles. |
| 9 | `c-retry-reopened` | POST exact original create-a with original key. Same exact body and id replay oracles as6, including original revision/effect position/transaction/creation facts. |
| 10 | `reopened-conflict-root` | Private GET original pre-reopen c-root grant; whole body equals c-body. Explicit current-context presence. App proves served generation > original source generation, same logical config identities, exact observed target, completed wire. |
| 11 | `c-state-reopened` | Repeat `#/$defs/currentBeadCollection`, exact one-record collection and full `/items/0` equals c-record. No whole HTTP array capture. |
| 12 | `c-events-reopened` | Repeat exact one-created-Event/no-next and current context equality with c-commit, including presence. No event-group or erasure inference beyond this fixed all-visible one-commit fixture. |

Every ordinary HTTP action includes status/schema where applicable, exact authored header/body observations and explicitly specified credential. Every dynamic read checks status/bodyPresent and epoch/view/position presence, plus current private/no-store/validator omissions where the existing reader pattern applies. Whole body equality uses existing json-pointer equalsBinding; its binding is a client JSON capture, not an attempted HTTP array capture. Original successful POST and subsequent receipt wire remain independently recorded.

Do not require receipt-style context headers on the direct409 problem unless a separate actual law requires them; this plan relies on schema/problem facts and following observed reads. Absence of a returned receipt is not proof of all internal bookkeeping. Evidence for no visible effect comes from exact collection/Resource/Event observations, unchanged serving position, original outcome replay and reopen.

## Negative controls and independent static guards

Five controls, each modifying one declared expected fact and executing the real installed row, without fabricating a response:

1. Change c-conflict expected code to malformed-request. Require actual409/complete wire and exactly code assertion failure.
2. Change c-conflict-reopened status expectation to200. Actual409 must fail exactly that status assertion; this discriminates the after-restart path.
3. Change c-state's expected one-record tuple from a to b. Actual single a must fail exactly ordered-state; keep full record-equality assertion unchanged.
4. On c-retry-reopened, replace whole-body equalsBinding with literal `{}`. Actual200 retained receipt must fail exactly retained-body, proving the replay comparison is active. This is oracle sensitivity, not an injected server corruption claim.
5. Change c-events-reopened expected Event tuple to created/b/0. Actual created/a/0 must fail exactly ordered-facts; this protects the no-new-effect witness.

Static guards independently construct all source literals, exact12-action ID/order ledger, changed request identity b versus original a, the exact five-key/body mapping (three original create-a bodies, two create-b bodies) and original retry bytes, both409 shape assertions, unacceptable first Accept vs JSON later, exact single Resource/Event tuples, original body/record bindings, and whole-body equality on both replay POSTs and reopened root. Presence assertions precede every context comparison. Pin both exact canonical source/anchor/selectedText requirements. Pin coverage's exact ordered action:assertion mirror separately. Do not regenerate expected literal oracles from manifest values.

A source-static semantic drift control should change the conflicting request to create a (the original semantics) and fail the explicit authored request guard; do not pretend that changing idempotency semantics is a server fault. Existing whole predecessor object hash tests remain authoritative.

## Files and overlap

Exactly four source files proposed:

1. `packages/conformance/test-support/transactional/manifest-v1.json` — one12-action row.
2. `packages/conformance/test-support/transactional/coverage-v1.json` — one false row, exact bindings and limited reason containing the mandatory literal “Whole contract remains unqualified” plus the complete applicability exclusions below.
3. `packages/conformance/src/transactional-cumulative.test.ts` — append pagination to the existing eight original-object pins in both artifacts, thereby preserving nine; add new static pins/count/budget/clause guards. Retitle the existing preservation test at785 from “adding one bounded partial receipt page row” to the completed-key conflict slice.
4. `apps/bdptest/src/transactional-matrix.test.ts` — ten-row/151-action/18-child ledger, one new reopened ID, row-specific byte bounds and five controls. Retitle the installed positive at329 to “ten partial Transactional contracts” and update its source-bound summary label consistently.

No fixture, configuration identity, route authority, testing export, public runner, helper/runner implementation, schema, server/runtime, Type, catalog, Read evidence, or RU file changes. This does not overlap Update's schema PropertyChange fixture/app/coverage work. If main advances before authoring, root supplies exact ordinary base and verifies these four sources before implementation. Preserve prior source/evidence; integrate other lanes ordinarily, not by refreshing historical hashes.

## Feasibility and budget proposal requiring root acceptance

Current manifest377449B/10082 value nodes; coverage54875B; fixture716B unchanged SHA3a7899ced771d4880066d98b0000c53f801ece53458502cc4f3bc641146c9a85. Current live authoring fences require four explicit reconciliations: conjunctive manifest inequalities at cumulative.test.ts824–825 leave508B effective headroom (377957B ceiling); node cap10200 at838 leaves118 nodes; coverage41812+16KiB at826 leaves3321B; exact139 actions at823 must become151. The new row cannot honestly fit those unchanged live guards. Request a deliberate authoring-only successor budget; do not silently increase existing numbers:

- New row exactly12 actions; total151. Ten partial rows; all ten whole flags false. Existing139 actions/raw spans unchanged.
- New manifest growth from exact main627 bytes <=32KiB; absolute total <=404KiB (413696B). These are conjunctive. Standard repository formatting only.
- New row/manifest growth <=1200 authoring value nodes; total <=11300. Existing runtime `OBSERVED_JSON_MAX_CONTAINER_ENTRIES` and all observation/transport limits remain unchanged; authoring counts are separate.
- New coverage growth from54875 <=8KiB, absolute <=64KiB; preserve all nine existing coverage spans. Actual action/assertion mirror count measured after authoring, not guessed.
- Fixture, route inventory, and exactConfigurationDigest `634214546089490d235fa404166dc868a0feb5dc2a8af2a566a6ad33664b4b1a` unchanged. Historical source/result hashes remain historical.
- New row <=2KiB raw authored POST and <=8KiB decoded response per exchange. One visible live Resource/one Event; changed candidate b must never commit. Existing M/Z2KiB/32KiB and receipt16KiB/32KiB guards untouched. No materialization/count/runtime setting change.
- One additional scenario plus its reopen:17 scenario/reopen sessions + bootstrap =18 children, all closed. Thirteen dynamic reopened observations, adding only reopened-conflict-root to existing12. No lost HEAD witness.
- Preserve90s cohort,120s outer,5s wire,10s prepare/cleanup and63s positive-runtime guard. Last exact0359 positive was3.373s, but that is historical feasibility evidence, not a time guarantee for the new source. Five controls run separate bounded cohorts; every run owns/drains its children.

Projection feasibility is source-grounded: the existing11-action failed-retained row is exactly18,404B for its pinned raw object span (transactional-cumulative.test.ts731),511 nodes,64 assertions; this proposed12-action row has more shape/equality assertions but no32-item arrays or Resource fanout. The32KiB/1200-node allowance is a conservative proposed ceiling, not a measured generated candidate. Before execution, author must generate normally formatted candidate and measure bytes/nodes/coverage/exact actions; report any overrun and stop rather than dropping required oracles or auto-raising caps.

### Exact preservation and live-fence replacement checklist

Before any authoring, independently extract exact object spans from the assigned parent with the existing objectSlices semantics. For unchanged main627 bytes, add these ninth pins to originalObjectPins, leaving all eight existing hashes/lengths untouched:

| Artifact | Row | Exact raw bytes | SHA256 |
|---|---|---:|---|
| manifest | transactional.receipt.pagination |138750|c531c260349347cdc3e54ba68282f1a266a5fa207d3775660615f836a4768782|
| coverage | transactional.receipt.pagination |13057|0735edbca2fbf384f9f8147140a353dc084d9ad85facaddf2f5073f56a7256ee|

These spans were independently re-read and hashed during this plan-only revision. Parent drift requires root reconciliation, not refreshing the pins blindly. A preservation guard must visibly iterate nine entries per artifact; add a pagination-specific drift discriminator in addition to the existing older-row counterexample.

Replace the following live global authoring guards only after root accepts this budget. Retain old values as historical context if useful; do not leave obsolete inequalities active and do not relabel old evidence.

| Current guard / source line | Proposed replacement | Effective maximum |
|---|---|---:|
| exact139 actions (:823) | exact151, with exact12 new |151|
| manifest−238693<=136KiB (:824) | manifest−377449<=32KiB |410217B|
| manifest<=370KiB (:825) | manifest<=404KiB |413696B; the growth guard is tighter|
| coverage−41812<=16KiB (:826) | coverage−54875<=8KiB **and** coverage<=64KiB |63067B; the growth guard is tighter|
| nodes<=10200 (:838) | nodes−10082<=1200 **and** nodes<=11300 |11282; the growth guard is tighter|

Fixture716B/hash, runtime limits, old per-row wire/raw guards and all deadline guards remain active unchanged. Explicitly inspect all live assertions in these owned files for previous total-row/action/session values; update only the documented successor ledgers. Standard formatting and actual candidate measurement remain gates before tests.

## Acceptance and validation when separately authorized

1. Root accepts this scoped plan and authoring budgets after council; no human protocol decision is currently needed.
2. Four-file ordinary successor from assigned verified base; pagination promoted to a pinned original in both artifacts, all nine old row objects/fixture/routes/config identity unchanged; all whole/profile flags false.
3. Standard formatted candidate meets every declared authoring/action/response projection fence before tests; no hidden helper requirement.
4. Static guards pin exact semantics and both selected clauses; focused cumulative/helper/catalog tests and type/build/boundaries/format/lint plus original Read seal pass.
5. Fresh packed target source/hash/config identities are recorded. Ten positive rows/151 actions complete; all18 children closed; all13 reopened dynamic observations have correct original source/current generation and context checks. Same old255 result mappings,250 required/five optional;245 rows unplanned and honestly unobserved (240 required plus five optional).
6. Both actual conflicts are409 idempotency-conflict, including first unacceptable Accept; no additional Resource/Event or serving-position change; both original replay POSTs and reopened root equal original body; current Resource equals original captured record across reopen.
7. All five negative controls fail their intended named assertion on complete actual wire; target failure is not misclassified as a harness failure.
8. Fresh genuine RU positive aggregate may be run at ordinary integration to prove combined packaging; do not repeat or relabel broad RU negative cohorts merely for this TX artifact-only change. No8k suite claimed unless root runs it.
9. Exact source is frozen before independent native/Claude code review; root owns publication/integration/CI. No full-row or profile eligibility promotion in this slice.

## Applicability and remaining gaps

Evidence is for one authenticated principal, fixed all-visible view, one canonical Scope/epoch, available completed detail, ordinary explicit-ID create batch and ordinary same-epoch process restart. It does not establish all-epoch duration, completed detail expiry, failed-receipt retention transitions, restore/new epochs, another principal, view change, pending/unresolved keys, concurrency, cross-carrier/sequence equivalence, all operation kinds, rate/capacity precedence, or internal absence of an undisclosed receipt row. No optional History or erasure behavior is inferred. These limits must appear in coverage.reason; the whole conflict row remains false.

No unavoidable human decision was found. Existing law and current private seams suffice. If actual installed evidence exposes a defect or an unexpected required route/helper, stop that portion, preserve failure and request root scope rather than changing protocol/runtime in this evidence slice.
