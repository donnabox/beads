# Command-minimum catalog follow-up audit

The new command-minimum obligation is currently **unmapped**, not proved by the cumulative catalogs. This is the deferred Claude F4 boundary, not a defect in the accepted serving repair and not a reason to block or displace the Selector serving fix. Mapping and applicable executable evidence must precede any public cumulative qualification.

Source inspected: clean `f2ac3f980fc8d35a3b485db04b813ad1b569d94d`, tree `74327f052cef6026b3ea10e186523ace9c87a128`. No tests, builds, services, installs, product edits or publication occurred.

## Law and current neighboring rows

`docs/specs/bdp.md:5792–5800` under `#transactional-http-consistency` requires mutation POSTs to ignore all three read-consistency request fields irrespective of presence/completeness/value, without affecting identity; ordinary syntax/bounds, authorization, idempotency, carrier validation and expectedRevision remain. It expressly covers RU and TX and preserves later receipt GET/HEAD minimum behavior. RU's cross-reference is at `bdp.md:3578–3580`, `#mutation-command-preconditions`.

There is no catalog requirement ID selecting this clause. Exact nearby IDs do not establish it:

| Existing ID | Source/citation and current evidence |
|---|---|
| `transactional.changefeed.consistency-fields` | TX catalog:1172; cites only “It never reports success with an older position.” at `#transactional-http-consistency`. Spec table:6936. This is read/context law, not the POST exception. No binding in the private TX manifest/coverage. |
| `transactional.changefeed.catch-up-timeout` | TX catalog:1185; cites `#transactional-problem-details` read timeout. It does not require ignoring fields on commands. |
| `transactional.http.conditional-reads`, `transactional.http.head-parity` | TX catalog:2040/2068; cite read refusal/conditional/HEAD rules. Preserve these independently. |
| `read-update.idempotency.semantic-identity` | RU catalog:1292; `#idempotency-keys` citations cover record normalization/guards, not the new request-field exception. RU coverage:12394 is wholeContract:false, without bindings. |
| `read-update.idempotency.cross-carrier` | RU catalog:1328; shared carrier namespace/identity citations. RU manifest:112592 has 40 actions; coverage:12408 remains wholeContract:false. |
| `transactional.idempotency.cross-carrier` | TX catalog:735, `#mutation-transactions`; spec table:6905. Retires only `read-update.singleton.idempotent-retry`. Private TX manifest:19721 has 12 actions and coverage:1948 is wholeContract:false. These existing cross-carrier observations do not carry copied minimum fields. |
| `read-update.singleton.expected-revision`, `read-update.singleton.idempotency-conflict` | RU catalog:377/480; revision and retained conflict laws remain separately relevant. RU coverage:4080/6505 says wholeContract:true for those existing cited contracts; that credit cannot expand to the new clause. |

No `bdp-minimum-scope-position` request occurs in either approved RU executable shard or the private TX executable manifest. Searches of catalogs, their requirement citations and both coverage inventories find no command-ignore mapping. Citation checks can pass because the old selected quotation is unchanged; spec/catalog mirror tests compare enumerated tables, not every prose MUST. Relevant tests: `read-update-catalog.test.ts:151` and `transactional-catalog.test.ts:207`.

## Smallest mapping follow-up

Recommend one new normative **Read+Update** row, inherited without retirement by Transactional. For example, proposed ID `read-update.http.command-minimum-fields` (not an existing ID), with the exact new paragraph as its authority and the RU cross-reference as supporting citation. Its title should state that copied read-consistency fields do not gate mutation POSTs or change command identity. Include the preserved admission/read boundaries in the selected requirements; do not silently credit those from a status-only check.

This is sufficient as a shared requirement because both profiles explicitly obey the same rule. TX must observe its additional batch carrier when executing the inherited obligation; it does not need a duplicate normative ID or a retirement that contradicts inheritance. Do not repurpose or broaden the read-oriented consistency-fields row, nor attach the new clause to an already whole-qualified existing row without new evidence.

The minimal metadata seam is the RU catalog, matching RU spec-table row, and the inheritance audit in `docs/design/profile-contract-planning.md`. `profile-contract-plan.test.ts:224–243` requires that audit to list every RU row in catalog order as Inherited/Retired. Cumulative selection already combines core catalogs and selects preserved inherited rows (`profile-contract-selection.ts:9,57–104`; `transactional-catalog.test.ts:169–184`). Add a focused citation/inheritance presence check so the new obligation cannot vanish while all old mirror tests remain green. Reconcile any affected coverage inventories/digests honestly; initially mark the new obligation omitted/unproved as required by the existing artifact schema. Do not relabel old catalog digests or historical reports.

## Evidence follow-up, kept separate from mapping

Current private tests are useful implementation evidence: HTTP A–E at `transaction-http.test.ts:1137,1193,1405,1529,1783`; emitted RU case at `installed-reference.test.ts:552`; fresh packaged TX case at `transactional-collection-minimum-installed.test.ts:1173`. Root's source/control results and fresh TX18/one-child artifacts remain attributed to their actual runs. They are not existing runner assertion bindings for the new catalog ID.

The smallest subsequent executable work is to port the relevant bounded observations into the existing RU shard/private TX cumulative machinery under the new inherited ID, using supported transport/assertion bindings and a separately measured request budget. Required distinctions are absence/partial/empty/malformed/foreign/future tuples across applicable singleton/sequence and TX batch carriers; same-key unchanged-semantic replay; changed-body key conflict; authorization/revision/carrier admission unchanged; subsequent receipt GET/HEAD minimum enforcement. Preserve concrete full results/afterstate and the named refusal sensitivity control. Source fetch accepting empty values is not independent wire proof: retain actual exact-serializer request-header evidence for empty fields, as Claude F3 and the final draft disposition require.

Do not assume the existing RU6/TX18 cases alone prove every operation-family/field-class partition. Reuse their established oracles and declare remaining partitions uncovered until a reviewed equivalence partition or additional bounded observation justifies credit. Keep wholeContract:false and claimEligible:false for partial bindings. No broad diagnostic matrix, new runtime seam, production edit, public discovery activation or pending minimum/Event ruling is needed to record this settled obligation.

Search scope: the frozen spec and profile-contract planning document; RU/TX catalog JSON; both approved RU manifests, RU coverage and artifact-set selection; private TX manifest/coverage; cumulative catalog selection/mirror tests; the new source/RU/TX tests; and full Claude command code review F4/F3. This audit proposes follow-up scope only; root remains CI/merge and prioritization owner.
