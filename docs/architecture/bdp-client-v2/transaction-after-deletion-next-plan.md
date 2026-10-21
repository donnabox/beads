# After deletion: retained cross-carrier Transactional outcomes

## Recommendation and authority

Propose one bounded **executable required-contract proof**, `transactional.idempotency.cross-carrier`: demonstrate that an equivalent singleton, one-operation batch and sequence member reuse one retained creation, then that sequence projection returns the retained old Resource while a fresh transaction advances current state. This is practical retry behavior for a client changing delivery carrier. It is not a newly diagnosed runtime defect or a substitute public-discovery implementation.

Supporting plan only, requiring root/native/Claude review before authoring. Sole authority remains `/Users/dbox/repos/wt/vickie/bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md`, [sole design plan](https://github.com/donnabox/beads/blob/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/design.md), Current priority and Autonomous profile completion (T2/T4). Read the live priority, `transaction-next-unblocked-serving-audit.md` and `transaction-consistency-rulings-audit.md`.

Inspected source is clean `39c160d0d35294003f400cef8d5ab5c6ec9aec9c`, tree `75e8fb97979e304bbe6fa24b7a570d9fa01efbbb`, at `vickie-deletion-safety-schema-repair`; runtime/spec/catalog are unchanged from verified main5995. Original main-equivalent checkout `vickie-noop-conformance` is8d66855/treea1ae75b822e1d7fadd739d818d09a25fb788e380. No execution or source modification was performed for this plan. PR115 CI and separately frozen SSEd534 remain independent gates; neither is assumed merged or qualified.

## Why this slice rather than a new serving patch

The current explicit receiving gaps still match the earlier audit: HTTP discovery remains deliberately unavailable; incomplete/unknown minimum tuples, minimum fields on commands, retained Snapshot/collection minimum interpretation, missing/hidden plus incompatible minimum precedence, and embedded set Selector admission phase require recorded contract decisions. Metadata-only incident-Link Event membership remains unresolved. The only newly settled gap is saved-header precedence for start=now, already owned by the SSE lane; do not duplicate it or touch its eight files.

Current-read receiving, Type/incident/alias minimums, compatible404 and fresh Snapshot minimums are already implemented. `transaction-http.ts:295–349,380–386,528–544` and the projection/minimum paths substantiate the residual boundaries. The read-only inventory does not establish that every uninspected contract is complete, nor prove that all small hardening work is blocked. It supplies no new settled production defect worth authoring ahead of this proof.

The current installed alias fixture already exercises turn-time alias retry/restart (`apps/bdptest/src/transactional-alias-installed.test.ts:300–426`), and Snapshot fixture already reconstructs a finite graph from a real checkpoint (`transactional-snapshot-installed.test.ts:450–590`). Recreating either as a new ad hoc helper adds less value. The cumulative manifest currently has no sequence action or cross-carrier row. The proposed row uses the existing raw HTTP and receipt-observation interfaces, with no launcher, runner, schema or fixture extension.

## Exact law and available implementation

- `docs/specs/bdp.md:3899–3922`, Mutation Transactions: all carriers use the same Scope/epoch/principal key namespace; carrier and name are delivery metadata; a one-operation batch, equivalent singleton and equivalent sequence member present one semantic request and retrieve the same receipt. Existing required catalog ID `transactional.idempotency.cross-carrier` quotes the latter sentence. Catalog position46 determines insertion order; preserve surrounding conflict/failed-retained rows.
- `bdp.md:3924–3935`: later identical requests return retained outcomes and never execute again; conflicting normalized requests are refused before admission.
- `bdp.md:4050–4088`: sequence returns its unchanged result envelope, not a receipt; available completed detail projects its single entry with the **present** member index/name. Use this as oracle authority, not automatic whole-credit for the separate sequence rows.
- `transaction-http.ts:303–332` validates sequence carrier/auth/limits and invokes `sequenceObserved`; lines333–360 invoke the same service submission machinery for batch/singleton. No top-level Idempotency-Key on sequence; per-member keys only.
- `transaction-sequence.ts:32–74` projects the retained receipt entry and replaces index/name with current presentation. Current source supports the intended completed paths; no pending-template, alias, clock or policy seam is needed.
- The existing client receipt action captures complete observed JSON after an actual root GET (see current conflict row `c-original`), while raw HTTP captures are URL-only. Reuse that supported distinction; do not attempt a raw Resource/body capture or broaden the DSL.

All new coverage remains `wholeContract:false`, global `claimEligible:false`. Explicit limits: same principal/view/epoch, explicit-ID ordinary Bead, completed available detail, two mutations, no pending/concurrency, failure retention, alias/local reference, allocated ID, expiry, restart, erasure or all-operation/all-origin-carrier crossproduct. Existing unknown-key admission/order and sequence per-member transaction laws are not fully qualified by terminal observations.

## Proposed twelve-action row

Use only the existing Task Type and a single explicit `beads/a`. Initial properties exactly `{title:"carrier-original"}`; fresh update replaces title with `"carrier-current"`. Distinct literal keys `carrier-create` and `carrier-update`. No Type or fixture changes, no actor/provenance proposal.

|#|Action|Discriminating observation|
|---:|---|---|
|1|POST `operations/create-bead`, key carrier-create|Completed single-result creation receipt, position/effect/required pos-1, exact literal Resource facts, schema and root URL capture.|
|2|Existing receipt-read client action on observed root|Actual root GET; pin the same literal creation/status/index/context facts, capture complete receipt body and Resource for later equality. No expected value derived solely from the response.|
|3|POST one-operation batch with the equivalent create and same key|Complete receipt body/root equals action2, no extra effect; current context pos-1.|
|4|POST sequence with one equivalent create, same member key, name `first`|200 transactionalSequenceResponse, exactly one created result/index0/namefirst, Resource equals action2; no receipt member.|
|5|POST one-operation batch update with fresh key carrier-update|Completed update, pos-2, exact current title, old revision differs from current revision where existing assertion forms permit; capture new root URL.|
|6|Receipt-read client action on new root|Actual current receipt GET; independently pin current title/kind/identity/index/outcome/context, capture complete body and new Resource.|
|7|POST sequence of retained create(name `again`) then retained update|200; exactly two results in declaration order: old created Resource/index0/nameagain, current updated Resource/index1/no operationName. Both Resources deep-equal their validated receipt captures; still pos-2, no added receipt envelope.|
|8|POST singleton update-bead with equivalent carrier-update request|Complete body/root equals action6; sequence index1 must not overwrite the retained receipt's index0.|
|9|POST singleton original create with carrier-create|Complete old receipt equals action2 although current serving header is pos-2.|
|10|POST one-operation original create batch with carrier-create|Same old receipt/root, current pos-2.|
|11|GET existing `beads/?limit=32` route|Exactly one row/id and next:null; full Resource equals action6; independent title/metadata/type/id facts prevent a mutually wrong baseline.|
|12|GET existing finite `events/` route|Exactly creation and update facts, ordered contexts/subjects/ordinals, current pos-2 and next:null; literal creation and property patch payload facts. No third Event from replays.|

Actions2/6 are client actions, all others raw HTTP. All12 produce one actual HTTP exchange; no reopen. Parent total187actions/180exchanges becomes199/192. Existing receipt-target adapter alone handles dynamically observed URLs; raw targets add only `operations/create-bead`, `operations/update-bead`, `operations/sequence` via `{binding:"scope",path:...}`. Derive and independently pin the complete canonical method/URL inventory and digest, do not merely amend a hash produced by the same helper.

Every response must have a matching normative schema where available, correct status/content type, authored request/body assertions, and explicit required context presence before context equality is credited. All complete receipt comparisons use `exists:true` plus existing equalsBinding, with separately authored mandatory literal fields. Resource revision/opaque receipt identity are observed values, not invented literals. Initial source references are captured only by the existing supported URL capture, then followed by a real receipt GET. Pin both client IDs separately from all12 exchange IDs, with every assertion/capture and terminal row pass. No new postwire transformation or helper framework.

## Two meaningful installed controls and static guards

1. **Semantic mismatch on batch delivery.** Change only action3's title under the same retained key to `carrier-conflict`. Update its exact authored request oracle and add independent actual409/transactionalProblem/idempotency-conflict/never/context assertions. Require those pass while the old expected completed receipt/body assertions fail by their named IDs. This distinguishes equivalence from blanket same-key reuse, before any later fresh mutation. Do not accept any arbitrary row failure.
2. **Unbound sequence key is not a retry.** Change only action4's member key to `carrier-create-new`, keeping identity/body/name. Require actual200 sequence envelope with an identity-taken member problem at index0/namefirst and would-be status409/retry `never`, as explicitly specified by the problem table at `bdp.md:3416` (identity remains occupied even after deletion; see3477ff). Preserve request/schema/context assertions; add exact required Problem member assertions while permitting normative optional fields. Require intended old successful-projection/Resource-equality assertions fail, and all newly authored refusal assertions pass. This discriminates a genuinely retained projection from an unconditional successful response. No new same-key receipt or GET is inferred for this control.

The exact failure-ID lists and derivative normal-form size/node/assertion deltas must be published at stage1. A stronger third installed control is not assumed necessary: static weakening controls must separately remove full receipt equality, change current name/index and replace old Resource equality with current Resource equality, and fail the independent row guard by named cause. Positive old/current divergence plus full current-state/Event oracles is the independent data discriminator.

## Four owned files and preservation

1. `packages/conformance/test-support/transactional/manifest-v1.json`: one new existing-ID row only.
2. `packages/conformance/test-support/transactional/coverage-v1.json`: one false whole flag and exact ordered bindings/limitations.
3. `packages/conformance/src/transactional-cumulative.test.ts`: independent content/capture/mirror/route/preservation/budget pins, meaningful weakening controls; preserve schema resolver regression.
4. `apps/bdptest/src/transactional-matrix.test.ts`: route/digest/counters/wire filters and two bounded controls using existing machinery; no launch/executor changes.

Preserve all24 current raw row objects (12manifest+12coverage), including deletion333 and no-op223 assertions, byte-for-byte; reconstruct full parent documents after removing only the new row. Fixture/catalog/schema/runtime bytes remain unchanged. The four paths are disjoint from active SSE's eight owned files. Root supplies final ordinary parent after deletion/inflight integration; re-anchor exact identities rather than assuming39c is main. No shared worktree or outward action authorized by this proposal.

## Preflight gate, limits and execution

Measured39c parent M620094B/N16373/C88043B,187actions/1872assertions. Runtime artifact admission remains1MiB per TX artifact and100000nodes; these are not the authoring allowance. Proposed conservative stage1 stops: <=100000B manifest growth, <=2600nodes, <=18000B coverage growth, <=300new canonical assertions, exactly12actions. Absolute candidate stops720094B/18973nodes/106043coverage. Derivatives must also fit those absolute stops; no weakening or cap increase to fit. These are planning ceilings, not measured final content. Stage1 must report normal-form actual M/N/C/A, all24 raw hashes, full-parent reconstruction, independent route preimage/digest, per-action capture/assertion lists and derivative failure lists. Commit binding actual-plus-small-margin guards only after root accepts measurements; do not silently reuse deletion's older growth ceiling.

After approval and clean source freeze: pinned Node24.16/pnpm11.20; build before emitted tests; focused cumulative/app-static, type/format/lint/boundaries/Read seal; fresh exact-source pack/install positive plus two controls. No new test timeout. Existing per-exchange5s/cohort90s/outer120s/focused30s remain. New positive21children (20scenario/reopen sessions plus bootstrap),192HTTP; each derivative at most21/192. Total three planned installed cohorts <=63children/576HTTP, early-stop actual counts recorded. Existing six deletion/no-op control request bodies and assertion lists stay untouched; because the new row is catalog-ordered after those rows, their reached prefixes stay unchanged. Static guards must verify this; do not mechanically inflate their observed exact counts. If a shared harness edit actually affects them, seek root-approved targeted rerun rather than asserting unchanged behavior. No full workspace/full historical cohort repetition.

Report clean source/tree/parent, tar and installed payload identity, exact configuration/policy/auth/fixture/catalog/route identities, genuine complete result files/hashes, pass/expected-failure IDs, timings, all child shutdown and retained failure attempts. Source provenance remains launcher/reviewer binding, not target self-attestation. No runtime behavior change is predicted and no public discovery/profile claim follows.

## Machine-checkable done-when / stop conditions

- Existing required ID selected; one new authored row, whole false; all old24 objects and fixture/runtime/catalog exact.
- Exactly12new actions/199total,192positiveHTTP, two actual receipt GETs; three closed carriers share exact retained bodies/resources while only the fresh update advances to pos-2.
- Current sequence presentation independently pins names/indexes; both old and new receipts survive it unchanged.
- Complete current collection and exactly two finite Events exclude replay side effects.
- Schema resolution, independent oracle and weakening guards pass; both derivative results show precisely intended semantic distinctions, not missing-route/schema/setup failures.
- Measured source and all derivative budgets accepted before installed execution; actual source-bound installed runs and child cleanup complete.
- Stop for DSL/fixture/route binding limitation, runtime defect, unknown normative disposition, source/count overflow, or lost predecessor assertions. Do not widen runtime scope or eliminate proof to stay busy.
