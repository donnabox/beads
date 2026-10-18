# Native closure — command minimum implementation plan v2

**All nine external findings and native N1 are closed at the plan level. No new actionable finding.** The corrected seven-file proposal is coherent against the actual clean main checkout. This closes the specific revision issues; it does not authorize authoring or execution, certify a future draft, or replace measured preflight and exact-head review.

## Exact inputs and source

- Plan `transaction-command-minimum-implementation-plan-v2.md`: SHA256 `f8824197e765e8228679044ab32fd7a54ebd26ec3a892614fdd9b5a15cf07c82`.
- Dispositions `transaction-command-minimum-plan-v2-dispositions.md`: SHA256 `53af77f99ae45a9178089c2912c2f5806df1520734e8c8201f36c9af36a7a427`.
- Completed Claude report: SHA256 `8b17956af2e85a15da106b4c13ea3d28ba49c189ade549cc03112a4940f3b6ab`.
- Original v1 remains unchanged at `00ffec1ab3146fa62601a66c95faf1eafae9986d0bad3070b599dbd1599b8ab6`. The earlier native no-blocker review is historical and does not negate the subsequently found fixture and regression defects.
- Actual checkout `/Users/dbox/repos/wt/vickie/vickie-command-minimum`: independently verified clean HEAD `122800511ee6306e4105e1f73bf3658f35c90a14`, tree `088290751951c353797b984f690b7d9fd96deb80`.

Read the complete v2 plan/dispositions and original council findings, and inspected the source supporting the corrections. The operator ruling remains unchanged: ignore the three copied minimum fields on mutation POSTs; normal command admission and subsequent receipt minimum checks remain operative.

## Specific closure

| Finding | Independent closure |
| --- | --- |
| F1: stale service exclusion row | `transaction-service.test.ts:2017–2072` has seven original rows. Removing only the minimum-position row at2040 leaves **six**, not five: prohibited carrier key, malformed JSON, unauthenticated, forbidden, media, unsupported composition. V2 owns this seventh file and selects the exact existing case. Keeping context absence, zero refusal-observer calls after each row and final position0 is valid because the remaining routes refuse before sequence execution. The old minimum row would execute after the production deletion, so its removal is necessary and now explicit. |
| F2: invalid singleton member | V2 explicitly omits `operation` from every singleton body, including A/D and installed/RU create/update. Batch/sequence members retain the discriminator. This matches schema-backed carrier validation in `read-update-values.ts:203–216` and the existing body-override refusal test at `transaction-http.test.ts:2080–2087`. It corrects previously masked fixtures without changing syntax/auth order or parser semantics. Actual bodies and bytes must still be verified in draft preflight. |
| F3: installed page maximum | Steps1/18 now use the same `beads/?limit=32` URL. `transaction-runtime-configuration.ts:95–96,334–338` supplies default2/maximum32, and the installed test already rejects limit33 at551. Three final Beads fit this page. The nine-URL/eighteen-request count is unchanged. |
| F4: B pagination | B's Beads and Links reads now both request limit32. Four expected Beads fit one page, so exact a/b/c/e and next:null is feasible with the existing fixture. No continuation request is hidden. |
| F5: field vocabulary and applicability | The full rule moves to the field definitions under Transactional HTTP consistency. The field-name-free cross-reference in Mutation command preconditions explicitly preserves the generic operator ruling for Read+Update as well. It introduces no RU response tokens, durable receipts or Scope-history capability. This is the root-accepted placement/applicability disposition, not adoption of the reviewer's suggestion to narrow the ruling. |
| F6: catalog citation | V2 preserves the exact sentence `It never reports success with an older position.` inside the same anchored section and changes adjacent request-side wording. The actual catalog at1179–1181 selects that sentence. Q selects the existing citation-validation test, so the normative amendment is checked without editing catalog source or adding an eighth owned file. |
| F7: progress prose | Both current command-refusal claims at660–661 and1727 are now explicit edit sites, with nearby1730 qualified as read-control continuity. Unsupported-composition500 and historical evidence remain distinct. |
| F8: targeted regression gate | S and Q are exact existing selected cases, not whole service/catalog file execution. They cover the two omitted CI-sensitive areas while preserving the finite gate. |
| F9: anchors and selector scope | V2 corrects the handler/RU/spec anchors. Existing S is inside `describe("nonexecuting sequence refusal observation")` at1905 and Q inside `describe("draft Transactional catalog")` at54. Their full anchored patterns include those prefixes. A–E and the new RU/TX cases are explicitly required at module scope. Draft title expansion remains required before execution. |
| Native N1: auxiliary listener | V2 counts RU's `availablePort()` listener separately from its serving child/owner. The helper at `scripts/installed-reference-observation.mjs:10–16` actually listens and awaits close. S adds an owner but no listener. Final source/control totals therefore correctly converge to15 owner opens and15 listener lifetimes, with different composition. |

## Service case and finite ledger

S makes six direct `handler.handle` calls using six POST Request/body inputs: four valid texts under differing header/auth conditions, one malformed text, one unsupported-depth carrier. These are **not TCP HTTP exchanges**. The source fixture at `transaction-service.test.ts:54–116` opens one owner/store and constructs one service through f.open(); handler construction adds no listener or child. `transaction-service.ts:102` reads owner configuration in construction; the test's final inspect is retained. This is correctly described as no admission/member execution, not zero owner access. The existing reversed cleanup list attempts service.close, the idempotent owner.close, then removal; close attempts are not new owner opens. V2 does not miscount ordinary read-plane construction or silently refactor historical hooks.

Q reads/parses catalogs and validates spec citations, with no service, store, listener, child or request. Its actual full title matches the proposed selector at `transactional-catalog.test.ts:54,66`.

Recomputed the planned ledger:

| Workload | Executions | HTTP expected / maximum | Stores | Owners | Listeners | Children |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| A–E |5|80 / 80|5|11|11|0|
| Two source derivatives |2|3 / 33|2|2|2|0|
| RU emitted regression |1|6 / 6|1|1|2|1|
| S |1|0 / 0|1|1|0|0|
| Q |1|0 / 0|0|0|0|0|
| **Source/control total** |**10**|**89 / 119**|**9**|**15**|**15**|**1**|
| Separate fresh installed TX |1|18 / 18|1|1|1|1|

The source/control total is eight positive case executions plus two intended-failure executions. Six read-only SQLite handles remain confined to D. Fourteen serving listener lifetimes plus the RU probe make fifteen; S contributes the fifteenth owner without a listener. Method arithmetic also reconciles: source/control expected60 POST +28 GET +1 HEAD; maximum77 POST +40 GET +2 HEAD. Fresh TX remains8 POST +9 GET +1 HEAD. Optional combined totals107/137 requests do not permit extra requests in a cohort.

The two source controls remain narrowly bound: old command refusal restored only for A, and receipt-root minimum dispatch bypassed only for E. Rebuilds are required because the HTTP source tests import emitted runtime code. Each must fail at its specified actual response oracle, preserve primary/cleanup evidence and stop on a different outcome. The existing source/installed limits, strict receipt-root grant, authorization-view resolver channel, safe fingerprint-report limitation and independent process/directory cleanup checks are retained.

## Scope and remaining obligations

The production edit remains deletion of only the command-specific minimum refusal block. The seventh file is a test-only stale-row correction; no service production change, new schema/catalog/fixture, public profile activation, read-minimum reinterpretation or Selector-law implementation is introduced. Remaining minimum-position assertions inspected for this correction are read paths and are not candidates for bulk removal.

This is sufficient closure of the specified F1–F9/N1 revisions. The actual seven-file draft, exact payload/schema/body ledger, selector expansion, derivative patches, cleanup instrumentation and measured bounds still require root preflight release before execution. There is no compiled, test-pass or installed claim at this planning stage.

Reviewer performed source/document/hash inspection and this guarded report write only. No tests, builds, services, installation, source edits or product/remote mutation. The only network operation was the explicitly permitted read-only role-session check required by guarded-action.py.
