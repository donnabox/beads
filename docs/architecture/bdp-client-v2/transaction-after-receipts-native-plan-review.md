# Native council review: completed-key conflict evidence plan

## Findings first

**No High or Medium finding. One Low wording correction; the proposed four-file,12-action partial observation is feasible on the inspected source.** Implementation still requires root approval and actual normally formatted budget measurements; this review does not confer row or profile credit.

### [P3] The key is presented five times, not four

- **Location:** `transaction-after-receipts-next-plan.md:48` (“all four presentations”), versus its action table52–60.
- **Evidence/impact:** `c-commit`, `c-conflict`, `c-retry`, `c-conflict-reopened`, and `c-retry-reopened` are five POST presentations. The first and both retries use original create-a bytes; the two conflicts use create-b bytes. A literal four-presentation guard derived from the prose would omit one key/body comparison even though the12-action table is correct.
- **Fix:** say “five presentations: three original-semantic POSTs and two conflicting POSTs,” and pin the complete five-key/body mapping. No additional action, runtime behavior or budget is needed.

## Exact basis and method

Reviewed plan SHA256 **c2a6a8e892c70593f96b257aa33f2acb5138fe23ed672b5f7ca8d86e1ac40663**. Read frozen `/Users/dbox/repos/wt/vickie/vickie-receipt-page-review-repair`, HEAD **0359f1cae3f36f600b1b375ea276ec7d60bf689b**, clean, tree **6a5337cf9548cfe340593a3858213593e4aca1d9**, matching the supplied main627b560c0daf1a76b292615718c86670255ac9af tree. No external council report was read.

Checks were source inspection (`git rev-parse`, `git status --porcelain`, `rg`, `sed`), SHA256 of the plan, and Python JSON reads/counts. No tests, build, server, credentials, network, source edit, branch change or outward action. Only this report was written. Some initial searches used nonexistent guessed helper/report paths and returned no file; corrected source paths were then inspected. The plan's3.373s historical positive time was not independently recovered from its report path in this review; it remains author-reported historical evidence, not a verified new execution or safety guarantee.

## Normative and runtime feasibility

`packages/conformance/catalog/transactional-v1.json` contains the proposed normative conflict row at zero-based44, with exactly the two selected source/anchor/text objects quoted by the plan. Canonical authored order places it after receipt.pagination and before failed-retained. Its full selected duration law extends through the rest of the epoch for completed outcomes, including expired detail (`docs/specs/bdp.md:3929–3939`). The plan correctly retains `wholeContract:false`, limits evidence to available completed detail and ordinary same-epoch reopen, and excludes expiry/epoch/principal/view/concurrency/cross-carrier branches. This is not a blanket whole-row claim.

The first unacceptable Accept is a useful prescribed branch: `bdp.md:4507–4532` orders key comparison ahead of negotiation for different semantics. `packages/server/src/transaction-store.ts:1030–1076` compares retained normalized identities; `:1372–1393` negotiates an **equal** existing receipt but passes conflict through. `transaction-submission.ts:100–122` maps the comparison conflict, and `transaction-http.ts:345–367` emits409 conflict type/code/retry never before its not-acceptable branch is relevant. `transaction-http.test.ts:228–248` already demonstrates conflict-before-negotiation at the implementation seam. This supports feasibility; it is not borrowed installed-row credit.

Create-b differs in an explicit, unused ID while preserving the admitted create-a Type/documents. It is a valid discriminating semantic change: an erroneous second execution would create another Bead rather than disappearing behind a duplicate-a identity failure. The unchanged Task configuration provides the same admissible content, and no new Type/Selector/alias/ownership rule is required. Both complete semantic identities and all five Idempotency-Key presentations should be independently literal-pinned.

## Capture, route and authority checks

The proposed `/id` HTTP capture is permitted. `packages/conformance/src/transactional-observation.ts:166–246` grants receipt-read authority only from an actual canonical mutation-source200 receipt, checks wire/body/schema agreement, binds credentials/configuration/session, and refuses reuse. Capturing only once on `c-commit` matters: retries must not overwrite that authority. Their `captures:[]` and the old `c-root` binding make the reopened read an actual cross-generation observation of the original receipt.

The private `transactional.receipt-read` root path compares the actual GET body to its originally recorded POST receipt (`transactional-observation.ts:275–306`), then the ordinary client JSON capture seam can retain `/body` and `/body/results/0/resource`. HTTP capture remains URL-only; no whole HTTP array capture or new framework is needed. Root-pointer `json-pointer` equality against the earlier client body works for the replay POSTs, and `/items/0` equality works for the current Resource. Neither grants extra route authority.

`transactional-observation.ts:120–163` validates a physical session transition, increasing generation with unchanged runtime/configuration identity. `apps/bdptest/src/transactional-matrix.test.ts:359–397` already binds every `reopened-` observation to original source generation, exact receipt/page URL, current configuration and context presence. Appending `reopened-conflict-root` to that list retains these checks and the existing HEAD case. One new scenario plus one reopen gives15→17 scenario/reopen sessions and16→18 children including bootstrap;12→13 reopened dynamic observations is correct.

The fixed exact route list in `transactional-matrix.test.ts:180–192` already includes batch POST, `beads/?limit=32`, and `events/`. No new ordinary GET to b is needed: the one-item complete collection proves absence. Existing dynamic receipt routing handles the opaque root. Fixture bytes/route order/configuration digest can therefore remain unchanged. Keep the literal digest guard and independently compare actual source/installed identities rather than blindly refreshing it.

## No-effects and comparison oracles

The combination is stronger than a status-only conflict test:

- One-entry created receipt with independently literal identity/Type/properties/metadata anchors `c-record`; private root read additionally checks equality with actual original POST bytes.
- Collection schema plus `next:null` and exact one-element ordered tuple establish completeness for this one-Bead fixture. Full `/items/0 == c-record` includes opaque revision and optional record members. It detects a changed original or an extra b without claiming arbitrary database absence.
- The Scope Event page's exact one-event tuple and `next:null` independently detect execution effects; equal epoch/view/position before/after conflict supplies another no-effect witness.
- Same-key original replay and reopened root equality prove the retained original outcome is still available, not replaced by the conflicting submission. Explicitly declared a makes accidental replay execution fail visibly.
- Final collection/Event checks occur after both reopened conflict and reopened original retry, covering both refusal and replay effects.

`runner.ts:2179–2199` implements response-metadata comparison, including equal status. Both compared original commit and Event GET are200, so it fits. It can otherwise accept absent==absent; the plan explicitly requires each context header's presence on **both sides**, which is necessary and must remain a static pin. No context-header requirement is invented for the direct409 Problem. No internal absence of a hidden receipt record is claimed.

Raw/authored request observations and actual independent expected values are appropriate. Exact configured Problem assertions must distinguish protocol requirements from optional fields; the proposed code/type/retry/status/schema checks do so without inventing closed-object equality or a new Problem tuple.

## Five negative controls

All five are feasible with the current app `observe({manifest})` seam (`transactional-matrix.test.ts:91–131`): create a fresh validated source-private bundle with one altered assertion expectation, then execute real installed HTTP. They change no target response. Existing tests at858–902 demonstrate the pattern.

Controls1/2 distinguish the two actual409 branches;3 protects collection identity/cardinality;4 protects entire replay body comparison;5 protects exact Event facts. For4, **remove** `equalsBinding` when adding `equals:{}` so the mutually exclusive assertion form remains valid. The word “replace” in the plan already requires this; no new seam is needed. Require the indicated exact single assertion failure, complete actual wire, actual expected status and cleanup. Retain those control results as failed expected counterfactuals, not successful qualification. No new helper/runtime fault injection is needed.

## Source guards and budgets to reconcile explicitly during authoring

The requested successor budgets are coherent but remain proposed maxima, not a measured candidate. Independent current counts reproduce377449manifest bytes/10082nodes and54875coverage bytes. The effective requested conjunctive ceilings are:

| Measure | Current | Effective proposed ceiling | Remaining growth |
|---|---:|---:|---:|
| Manifest bytes |377449|410217 (=current+32768, tighter than413696)|32768|
| Manifest nodes |10082|11282 (=current+1200, tighter than11300)|1200|
| Coverage bytes |54875|63067 (=current+8192, tighter than65536)|8192|
| Authored actions |139|151|12|
| Authored rows |9|10|1|

When implementing approved budgets, explicitly replace the obsolete global authoring guards in `packages/conformance/src/transactional-cumulative.test.ts:823–838`:139 exact actions; manifest−238693<=136KiB; absolute370KiB; coverage−41812<=16KiB; nodes<=10200. Preserve their reviewed values as historical comments if helpful, not live inequalities that inevitably reject the new row. The plan acknowledges replacement of current508B headroom; this is an implementation reconciliation checklist, not an extra semantic finding. Keep all runtime1MiB TX source/observation limits and declared fixture limits unchanged.

Current original-object pins at `transactional-cumulative.test.ts:690–777` cover **eight** pre-pagination rows. To implement the plan's promise to preserve all **nine**, add exact parent pagination manifest/coverage raw-object pins while retaining all eight existing hashes/lengths. Existing canonical mirror test at76–113 derives each complete action:assertion sequence, requires every `wholeContract:false`, and requires the reason phrase “Whole contract remains unqualified”; retain it and include the scoped applicability explanation. Exact clause objects and literal expectations remain separately authored, not derived from the manifest.

Update only owned app exact9→10 pass/session/reopened ledgers and summary label. The250 required/five optional mapping stays255;10 passes leave245 unimplemented catalog mappings, comprising240 required plus5 optional. The unchanged90s cohort,120s outer,5s exchange,10s preparation/cleanup and63s positive guard remain meaningful hard stops. Measure the normally formatted candidate before tests; stop on overrun rather than remove required oracles. A fresh measured run, not the historical3.373s, is the completion criterion.

## Conclusion and limits

Native plan review supports this bounded partial slice after the Low key-count correction and root approval of the already explicit authoring budgets. No unavoidable product decision, new runtime feature, private authority expansion, RU edit or framework change was found. The author must still satisfy exact generated-source guards and fresh installed positive/negative observations. No test execution, execution timing, new receipt behavior, whole-contract closure or public profile readiness is certified by this read-only review.
