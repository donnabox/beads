# Private Transactional cold navigation: next implementation proposal

Planning only, subordinate to the sole fork `bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md`, “Autonomous profile completion” (lines7–51). No implementation or activation release is implied.

## Conclusion

The missing navigation adapter is small and technically independent of public admission: the existing owner already builds truthful current discovery and the complete twelve-target directory, including finite minimum waits. Public `admitReadServerProfile` must remain Read-only. However, serving all three requested routes cannot literally be “without profile advertisement”: the required `bdp.json` body contains `profile:"transactional"`. A private listener and `x-bdp-development: non-attesting` do not exempt that wire claim from the cumulative behavior it asserts. Qualification evidence and mandatory behavior are separate gates.

A bounded adapter can be authored and validated as an **inactive internal projection**, while the existing external refusal remains. That would prove serialization/owner/HTTP projection, not actual cold navigation. The recommended scope below is the eventual actual route implementation; root should not commission another internal builder or a switch that silently bypasses admission merely to obtain a successful cold walk. If staging now is useful, keep the route-activation patch separately reviewable and leave real runtime requests returning their current refusal until its applicability/behavior blockers are dispositioned. Do not add a public factory, alternate profile, special test endpoint or runtime activation flag. The source-only projection can be tested directly over the actual owner without representing it as a live cold client.

The bounded inspection found **no additional, unambiguously prescribed minimum/visibility/depth repair** that can honestly replace the unresolved cases below. This is not a finding that all remaining requirements are implemented. It means the existing text and accepted rulings do not justify inventing another response policy to make this navigation patch appear activation-ready.

## Actual basis and authority

Actual merged main is `5d7fa3f049fa0b4db3aa8dae20e20d8e47fb3854`, tree `057f879132074166f2c0f643a8b45bbd9dd47c6f`, ordered parents `9bf601bfb570adae97ce8cdc99caf2cc7a2b34c1` and `8833c9242b7788e843e9eb97a2c99b2277fe75ef`. The inspected clean checkout remains at8833; its **entire tree equals5d7**, independently checked by Git. Source hashes below are taken from committed5d7, not a fabricated checkout move.

Read the sole design, current live ledger, accepted cold-client-v2, C3 native review, the existing `transaction-cold-discovery-serving-next-plan.md` and `transaction-discovery-serving-plan-v2.md`, plus the post-PR130/sequence serving audits. The older activation plans expressly withhold even private schema-valid discovery while known applicable mandatory gaps remain. Their completed prerequisites must be removed from the backlog, not implemented twice.

Normative distinctions:

- Spec73–95: Transactional is the complete cumulative contract. Scope is one advertised profile at a time.
- Spec914–978: Scope GET supplies the real service-desc Link; `profile` is mandatory and names the highest supported cumulative profile. A Read descriptor cannot legally contain the Transactional operation/history members.
- Spec3730–3799: Transactional discovery requires actual epoch/view/head/replay floor and all navigation members; it is not an out-of-band inventory. Schema availability alone is not implementation evidence.
- Spec2066–2119 and4321–4345: directory targets must be followed and depend on the advertised profile; only Transactional includes batch.
- Spec5788–5848: ordinary reads use the current transaction-consistent prefix; minimum and eligibility checks precede conditional shortcuts, and observed context remains on304/412.
- Sole design T2/T3 versus T4/P: implement serving first, qualify the actual target and enable public selection separately. This separation does not waive missing mandatory wire behavior.

## Exact obstruction and existing reusable seams

`transaction-http.ts:570–578` returns bodyless500 with `x-bdp-development-limit: transactional-discovery` for Scope pathname, exact `bdp.json`, and exact `operations/`. The generic Read hook also refuses discovery at195/228. The preceding minimum-route filter560–569 would reject these targets even if only the final refusal were removed. `read-http-routing.ts` already knows the inherited204 service-desc convention, but its Scope response has no Transactional owner observation; reusing it unchanged would lose the required tuple.

`transaction-service.ts:253–278` already exposes source-private `discoveryObserved(minimum?, signal?)`: snapshot controls, one guarded owner observation per attempt, check the supplied minimum against that entry, finite existing wait, then parse/assemble after releasing the reader. It returns descriptor, directory and the same frozen current context. It supports lifetime closure/fencing and does not need a second store read or new clock. Existing service tests1151–1702 cover this assembly/minimum/lifecycle seam.

`transaction-discovery.ts:14–69` binds the same twelve routes the HTTP handler dispatches. It produces16 required members plus `limits`, actual page/Selector/diagnostic bounds and no-pruning `minimumReplayPosition: pos-0`; it does not invent a receipt retention duration or replay checkpoint. The schema parsers already exist. No new schema/export/configuration/store mechanism is needed.

`transaction-runtime.ts:125–178` validates raw target/Host, dispatches method failures, authenticates captured credentials, rejects Origin, and calls the handler with the real disconnect signal. `transaction-http.ts:389` rejects anonymous reads before owner observation. Read-only authenticated credentials may discover; mutation permission is irrelevant. This private configuration has a fixed all-visible serving view: do not manufacture one view per principal or use actor attribution as authentication.

`context()` at HTTP113–118 and the normal handler tail628–634 supply the three fields and `applyReadHttpSemantics` once. `read-http.ts:10–46` handles JSON Accept/Vary,304/412 and HEAD representation length; runtime delivery suppresses the actual HEAD body. Public `index.ts:288–306` must continue rejecting non-Read selection; no new admission receipt or claimEligible change.

## Smallest eventual route change and test ownership

Production scope is **transaction-http.ts only**: recognize the three discovery routes after existing method/auth checks and before the generic minimum-route exclusion; preserve the inherited Scope-path query convention and exact descriptor/directory URL matching. Parse supported complete controls once and invoke `discoveryObserved` once. Scope returns204, service-desc Link and explicit private,no-store; descriptor/directory use the already parsed values and JSON response helper with the same observation. Preserve unsupported tuple behavior, canonical Host/Scope identity, ordinary auth/method ordering, sanitized faults and the normal finite-response tail. Do not derive ETags from headPosition or authorize a later request with an earlier probe observation. Audit the two fallback refusal sites explicitly rather than simply deleting every development refusal.

Smallest actual activation footprint is five paths:

1. `packages/server/src/transaction-http.ts` — route projection and accurate module comment.
2. `packages/server/src/transaction-http.test.ts` — replace the three-route refusal portion of the existing canonical-route callback; add bounded navigation/eligibility/HEAD/auth tests.
3. `packages/server/src/transaction-installed-contracts.test.ts` — reconcile only the existing discoveryBoundary helper, called before/after reopen in the callback at378; retain its compiled Types/receipt/sequence/restart assertions.
4. `apps/bdptest/src/transactional-discovery-installed.test.ts` — replace its specifically named private-refusal witness with the cold walk; preserve its historical evidence externally.
5. `docs/design/transaction-implementation-progress.md` — implemented/private/qualified distinctions and actual validation attribution.

The runtime header comment currently says no incomplete discovery claim; that remains a valid constraint. If root elects to amend it when activation becomes truthful, declare that sixth comment-only path explicitly. Service/discovery/store/shared Read/schema/public admission files need no production edits for the three-route adapter. Select existing owner observation tests as regressions, not additional copied implementations.

Observable source acceptance: all three GET/HEAD routes; exact Link and literal twelve-entry map independent of production constants; descriptor fields equal its own tuple; readonly and anonymous credentials;405/Allow/Origin/JSON406 with Scope204 distinction; satisfied/foreign-view/foreign-epoch/future minima before304/412; same-entry context advancing after an actual commit; disconnect and close/fault cleanup. An emitted restore-refusal derivative should fail on the **first actual Scope exchange**, after checking actual500/bodyless/development marker, never continue with out-of-band targets. Separate source-only missing-context/incorrect-map sensitivity can use existing small fixture methods if needed, but must be measured before execution.

## Proposed finite cold walk once activation is accepted

One fresh installed fixture, fixed non-root canonical Scope, operator plus readonly credential, one exact local self-contained Type/schema artifact. The client starts with canonical Scope, credential and physical dial mapping only. No `child.ready.targets`, synthetic descriptor, guessed operation URL, alias fabrication or network schema fetch is navigation authority. Ready metadata remains provenance, not discovery.

Proposed17 exchanges in one child: (1) Scope GET; (2) follow its service-desc; (3) follow descriptor.operations; (4) follow descriptor.types and require a complete one-Type inventory; (5) follow that actual Type ID; (6) follow its actual local propertiesSchema and compare exact captured bytes; (7) resolve directory.createBead relative to the returned directory URL and POST a schema-valid singleton (no operation field); (8) GET returned Resource ID; (9) GET actual receipt.id; (10) GET descriptor.events and assert the single complete transaction-linked group; (11) refetch descriptor and require head advancement with unchanged navigation; (12–14) HEAD Scope/descriptor/directory; (15–17) anonymous GET of those three with401 and no Link/body discovery/context leak. Readonly can perform the initial walk; use the writer only for7 and normal subsequent reads. Full original receipt body remains distinct from each current serving tuple.

These are **proposed counts**, not frozen executable evidence. A restore-refusal control is one request/one separate child; expected total18 requests/two children, no automatic retries/pagination/reopen. A first-source plan must specify exact titles, bodies, raw schema byte budget, emitted graph identity and all source-regression fanout before release. Existing source restart callback supplies a separate regression; do not hide it in17. Preserve inherited per-exchange/cohort/test/supervisor limits, streaming receive caps and safe report allowlists; independently verify child exit/PID/store absence. Fresh protocol/server archives and external consumer must verify source/archive/installed members, eight-package closure and actual payload/entry/configuration binding. No workspace fallback or old evidence relabeling.

## What still blocks truthful activation; no new ruling invented

The following are actual code boundaries, not counts of unobserved catalog rows:

- **S1-Q1 retained Snapshot/continuation minimum comparison:** HTTP406–411 and transaction-read36–60 still exclude retained requests with minima. The original anchor versus current serving comparison is the already pending decision; this report neither chooses it nor asks it again.
- **G2-Q1 metadata-only incident-Link Events:** transaction-events257–264 explicitly refuses that projection. Fixed all-visible credentials do not remove the trigger; an applicable configuration must implement the settled outcome after the existing ruling. A descriptor advertising Events is not proof this case works.
- **Complete minimum plus missing/hidden target collision:** transaction-read139–159 converts behind/foreign-view/old-epoch plus materialized404 into a private visibility gap. Compatible/satisfied404 is already repaired. General read text does not supply the receipt-specific nondisclosure ordering; importing it here would invent a precedence decision. Preserve the previously documented residual and consult any existing root disposition, not a new question from this audit.
- **Malformed/incomplete/repeated/unknown read controls:** transaction-consistency20–57 accepts the supported closed tuple; HTTP preserves explicit refusal otherwise. Mutation POST ignore semantics are already repaired and do not settle these read error codes. No defaulting unknown position to current, blanket400 or fabricated expiry is justified by this inspection.
- **Fresh private depth/composition limits:** store1083 and HTTP bounded composition branches remain private refusals. The retained-key-before-depth defect is already fixed. Do not undo finite safety bounds or relabel fresh private500 as a normative413 without a source-grounded policy/phase disposition; this audit established no new erroneous retained-key branch.

The old prerequisites for diagnostic disclosure, discovery minimum observation, initial collections/Type/incident reads, aliases, compatible missing reads, command minimum fields, Selector grammar/caps, sequence aliases, start=now reconnect and observed error contexts have been implemented. They are not new work. Public cumulative qualification remains a later separate evidence/admission gate even after functionality is correct.

I also checked the existing exact local schema adapter: it deliberately declines every supplied minimum (`transaction-schema-http.ts:42–46`, service schemaObserved541–548), and its regression232 explicitly pins that refusal. This is a useful follow-up applicability audit, **not a newly proved mandatory repair here**: Type `propertiesSchema` is an ordinary Web retrieval URL (spec1200), and URL containment plus current context alone should not silently settle which external/schema documents are BDP Scope-bounded read representations. Do not expand this task into Types or a general resolver to force a new priority.

Thus the actionable next choice is a small, inactive navigation projection/route patch for review, or closure of an existing applicable behavior disposition. There is no safe basis in this focused inspection to activate an incomplete profile or invent another serving fix merely to avoid evidence work. Authoring/validation/public activation remain root-owned gates; this report does not alter the sole plan.

## Inspection and source bindings

No tests, product imports, builds, services, reviewer subprocesses, product/coordination edits or public changes ran. The only write was this report after the existing remote session guard passed. Prior reports are historical inputs, not current runtime proof.

- `docs/specs/bdp.md`: `f036103ffdf5f89fa3a2033570e1e987c4cd203cd80b1c633cb44615123f892e`
- `packages/server/src/index.ts`: `34f9d9de8b48852c64919e69ce93f149e9cdae5142c15c75e4c35f08f15efc44`
- `packages/server/src/transaction-discovery.ts`: `bb8055e2089f640a5d2071e389ac2011d08c038427265b50ab3d03c02047cfef`
- `packages/server/src/transaction-service.ts`: `53f655faf494209996aac2e0f93a4c2c785e471f66aac6af9ea41f92c9cfbe00`
- `packages/server/src/transaction-http.ts`: `b51d02801e5924190aba77c377e538db0b8619149d6184e19cb56d440f40c2cf`
- `packages/server/src/transaction-runtime.ts`: `77954f0a85029d76443feac54cd53cda4339ffc126660ace5a9f5af92466c573`
- `packages/server/src/read-http-routing.ts`: `d9344b87838f8c71381ccf9efe28b4f023df45503570beb8fb4293d553caca42`
- `packages/server/src/read-http.ts`: `31934105eff6fff8ded8ce2ca9e3c19028991ca7c3d5f51b848004177e8c9268`
- `packages/server/src/transaction-read.ts`: `c3b486baf6b9e48d4a716d36b4162c2ba1f0c7672a6f48988192e2591eda7910`
- `packages/server/src/transaction-consistency.ts`: `1507f760c790c7232f4e200faa45fdcd3c508a3fbd02909c4c1caaba7a70b6c7`
- `packages/server/src/transaction-schema-http.ts`: `1082e8483848794971eae3eb8dfb3b7006749fc4e3b9c17c2cca1adbd94a6130`
- `packages/server/src/transaction-events.ts`: `c6052433662fcb2d0a202ad4e408a24d4c43f980313fa89659501b4475790db5`
- `packages/server/src/transaction-store.ts`: `fcd6882b29e538db45867c894c06dfbaa6ca588db0c2e403ef1b7abb14f6c706`
- `packages/server/src/transaction-http.test.ts`: `1c9cd7c7c85a5f6a58253f4178e40453a2443a218408eeba9abfddde4bf7948d`
- `packages/server/src/transaction-service.test.ts`: `c61ba7254ab1cf39abf85e6020e4f846ffb684740076de39f43fe12d509093b9`
- `packages/server/src/transaction-installed-contracts.test.ts`: `f2dd0993c4fa61742435d3bedb101afb021ca0f1b5904147ab367ba243286bd1`
- `apps/bdptest/src/transactional-discovery-installed.test.ts`: `7cf8822108489f3fb2f9c5ca1d97c6d7b832acca95a2f7d891c3413404d90c49`
- Sole design observation: `/Users/dbox/repos/wt/vickie/bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md`, SHA256 `e6434930a4d978a423963a4313ec50083924be1599eb0fbf0b3b340ea386637d`.
