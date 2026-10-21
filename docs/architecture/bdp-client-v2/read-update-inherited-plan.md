# Inherited Read observations on the installed Read+Update candidate

## Source and boundary

New isolated checkout `/Users/dbox/repos/wt/vickie/vickie-read-update-inherited`, branch `codex/vickie-read-update-inherited`, GitHub origin `https://github.com/gastownhall/bdp.git`, starts at frozen `31007d84a91964786cdc1424e4a55d2af83204c6`. The reviewed checkout remains untouched. This is a plan only: no implementation, installation, tests, public claims or outward writes yet.

Keep the installed factory's existing Scope, profile, shipped Type contracts, policy and authentication configuration. Expand only the captured fixture bytes and its actual public-HTTP setup; the resulting fixture digest/configuration binding changes honestly. Existing RU scenarios remain isolated and unseeded; only the new Read scenarios receive the fixture's declared graph. Every installed session still binds the same configured factory before any setup or observation. No store inspection/mutation, evaluator hooks, custom Type contracts, mocked HTTP, arbitrary lifecycle actions or secondary requirement list.

The canonical source currently selects 43 inherited normative Read IDs among 129 cumulative requirements. Source evidence: `packages/conformance/src/profile-contract-selection.ts`, `test-support/read-update-coverage-v1.json`, `catalog/read-v1.json`. The existing Read manifest is `matrices/read-v1.json`; its fixture is `fixtures/read-reference-v1.json`. These remain unchanged and continue qualifying the original Read target only.

## Immediate prerequisite: reviewed executor repair

Before expansion, close the native310 no-op-response assertion gap and correct update-bead whole-contract overstatement in a bounded successor. Await the final native findings and preserve310 unchanged for the external review. The inherited Read work below depends on that corrected source, not the historical five-whole-contract/124-uncovered claim.

## Bounded scope: 17 additional inherited rows

Two reviewable commits are proposed; the second can proceed independently of unrelated reviewer repairs once the first passes.

| Slice | Exact IDs (prefix `read.`) | Reuse and changes |
|---|---|---|
| A: cold/static Read (6) | `discovery.scope-service-desc`, `discovery.canonical-navigation`, `types.inventory`, `types.descriptor`, `collections.invalid-query`, `http.method-405` | Reuse native request/assertion forms. Replace fixture identity/capabilities honestly; inventory expectations use the 13 shipped Scope-local Types. Descriptor case uses actual native GETs plus the normative closed descriptor schema instead of the old external publisher action. Strengthen 405 with body-absence; test unsupported methods only on Read endpoints, never require 405 from supported mutation routes. |
| B: seeded Read and HTTP semantics (11) | `resource.bead-record`, `resource.link-record`, `resource.properties-view`, `resource.default-bounded`, `collections.beads`, `collections.links`, `http.get-head`, `http.cache-cors`, `http.accept-negotiation`, `http.conditional-reads`, `http.head-conditional-parity` | Reuse existing status/schema/header/ETag/conditional request assertions and independent fixture oracles. Seed two Beads and two Links through actual create operations in reverse canonical ID order; independent native reads validate complete records, open properties, metadata, exact properties views and collection order. Use non-owning task + relates Types so existing configured contracts suffice. Preserve the complete Accept/conditional/HEAD request matrix where valid, including status/error precedence and actual observed validators. |

Target IDs are a temporary work list, not a replacement required set. The existing canonical selection and single machine coverage inventory remain authoritative. Each proposed row is marked whole-contract only after auditing all cited clauses and its actual observations; otherwise it stays explicitly incomplete. There is no target percentage or automatic reduction to 107 uncovered: actual completed contracts determine the count.

### Discovery and prerequisite ordering

`read.discovery.document` is retired and must never appear in the new manifest or be reintroduced into selection. Canonical merged order places inherited Read before `read-update.discovery.document`; simply replacing an old prerequisite with that RU ID would create a late dependency. Keep the canonical catalog and required-ID order unchanged. Reused inherited requests instead perform a local successful discovery GET, validate `#/$defs/readUpdateDiscovery`, and capture their needed navigation within their own scenario. Retain valid earlier `read.discovery.scope-service-desc` prerequisites. Remove only the retired cross-row discovery prerequisite, replacing it with explicit local RU validation; do not strip the actual discovery request or claim that another row passed. Each scenario then has a complete observable setup without a forward edge. Test that neither retired nor forward discovery prerequisites enter the new manifest.

### Fixture and exact execution edits

Extend `read-update-reference-v1.json` with a bounded named seed graph and static oracles: Beads `beads/read-a`, `beads/read-z`; Links `links/read-a`, `links/read-z`; local task/relates Type references; known open properties and metadata; expected sorted collection IDs and complete properties. Preserve existing seven RU scenario states by seeding only new scenario IDs declared to need this fixture setup. Add named fixture bindings for resource/link/missing IDs and the shipped Type URLs. Expected revisions are never fabricated: conditional rows capture real ETags, and full-record expectations use actual operation capture only where required while static properties/metadata expectations remain independent input oracles.

Use existing exact HTTP machinery to perform setup create requests, recording each actual request/response/write observation with a setup label. Fail preparation before assertions if a seed create fails, returns a mismatched identity or yields an unexpected postimage. Preserve credential-handle redaction. No test-only direct storage seeding. Native Read assertions then issue independent actual GET/HEAD. The fixed seed identities permit a finite exact route inventory; query variants/methods remain explicit. Enumerate routes from the authored request targets and fixture bindings once using existing manifest/target helpers where available; do not invent a wildcard route or independently recompute exact identity. Preserve actual target/configuration comparison before all HTTP. This proposal requires no extension of dynamic captured Link routing or the client input language.

## Explicitly uncovered inherited obligations (26)

- `discovery.optional-limits`: fixed installed limits do not provide the omission/capacity variants exercised by the controlled Read action. Can observe declaration later but not claim whole contract from it.
- `resource.external-endpoint`: existing action relies on controlled external Type publisher and pin-faithful endpoints/owned closure; later real public writes can replace some legs.
- `resource.not-found-nondisclosure`: unknown/deleted can be produced publicly, but same-configuration non-visible authorization needs a separately captured policy fixture; do not claim from deletion alone.
- `collections.structural-predicates`, `collections.selector`, `collections.pagination-snapshot`, `collections.cursor-errors`, `bead-links.directions`, `bead-links.pagination`: require a richer graph, continuation capture and actual concurrent mutation or view transitions. Direction/predicate cases are the next uncomplicated extension; snapshot/expiry need separate lifecycle design.
- `http.problem-table`, `http.internal-fault`, `http.raw-request-targets`, `scope.restore-identity`: controlled failures, exact raw malformed target routes, and restore are outside the minimal fixture. No fake fault or reset outcome.
- `alias.resolution`, `alias.absent`: installed RU advertises aliases; resolution can use public putAlias later. Alias-absent does not apply to that advertised configuration, but remains unresolved/uncovered in this slice until the applicability decision is represented and reviewed; never auto-N/A.
- `disclosure.gone`, `disclosure.gate`: no public pruning/erasure setup in the present installed candidate; unknown/deleted is not proof of authorized disclosure behavior.
- `owned.closure`, `attribution.carried`, `attribution.link`, `owned-wildcard.declaration`, `owned-wildcard.max-required`, `owned-wildcard.explicit-max-bounded`, `owned-wildcard.present-entries`, `owned-wildcard.closure`: require declared owned/attribution fixture states and sometimes contract/policy variants; none is inferred from empty/default data.
- `numeric-model.declared-token-model`: needs actual numeric-model/unsupported-token adversarial inputs; copied simple integer fixtures do not establish it.

The five applicable diagnostics and other RU obligations remain explicitly missing in the unfiltered run. History and Types lifecycle are not added as prerequisites.

## Highest-value next Update reuse (not implemented in this slice)

Source `packages/conformance/development/read-update-manifest-v1.json` contains concrete real-runtime probes, but its action executor is different and its diagnostic result is not installed evidence. Reuse request payloads and independent assertion logic, not outcomes:

1. `development.read-update.documents` (5 actions) → singleton document merge/replace, patch atomicity and validation. Port open properties/metadata cases to actual installed exact HTTP with unchanged-state GET after failure and explicit no-op revision checks.
2. `development.read-update.links-delete` (11 actions) plus `pinned-references` (17) → create/update/delete Link, endpoint echo and metadata/properties preservation. Needs a narrowly approved canonical captured-Link GET route, symmetric with current Bead handling, with wrong-kind/foreign-Scope negative controls; do not generalize routing first.
3. `development.read-update.sequence-partial` (4 actions), `contextual-local-kind` (4), `contextual-durable-kind` (12), `carrier-rejection` (6) → a first real sequence success/failure/skip prefix with public afterstate and invalid-envelope no-effect. Add the already advertised exact sequence route; use literal member labels/static seed IDs initially, no body template framework. Audit full selected normative rows before marking complete.
4. `result-headers` (2), `raw-method` (3), `raw-media-json` (10), `raw-admission` (12), `raw-accept-keys` (17) → low-fixture-cost observable HTTP/error laws. Existing exact executor is reusable; every rejection needs proper precedence and unchanged-state checks.

Process-disconnect/drain/crash/pending/retention cases require lifecycle seams unavailable in this candidate's current harness. Do not copy their controlled diagnostic capabilities into installed fixtures.

## File ownership and minimal edit set

This lane owns only the new successor's following files for this slice:

- `packages/conformance/matrices/read-update-v1.json`: append inherited scenarios in canonical order; preserve original Read matrix.
- `packages/conformance/fixtures/read-update-reference-v1.json`: captured seed inputs, bindings and independent oracles.
- `packages/conformance/test-support/read-update-coverage-v1.json`: same canonical requirement inventory, exact assertions and remaining limits.
- `apps/bdptest/src/read-update-matrix.test.ts`: actual seeding, finite exact routes, progress/source/config binding and counterfactual probes. Extract narrowly to `apps/bdptest/test-support/installed-read-fixture.ts` if needed to keep the test readable.
- `packages/conformance/src/cumulative-artifact.test.ts`: coverage/order/prerequisite integrity and unchanged factory guards.
- `docs/design/read-update-executable-progress.md`: update observed row table and limits.

No edits to server/evaluator/configuration implementation, shared runner/assertion/parser/factory semantics, canonical selector, public exports, normative spec/schema, original Read evidence or diagnostics. Parent/review workers own repairs to those shared source files. If reviews require a prerequisite fix, root supplies an exact frozen successor for an ordinary merge before final validation; this lane does not edit the frozen310 checkout.

## Done when and validation

1. New scenarios run unfiltered against the same actual packed installed RU candidate, with explicit source/payload/contracts/policy/auth/fixture/exact-transport identities and `claimEligible:false`.
2. Retired discovery is absent; canonical order unchanged; every reused discovery prerequisite is earlier or replaced with local RU schema validation; original Read/development bundles remain unchanged.
3. Native Read and HTTP assertions observe actual seeded state, validators and afterstate. Counterfactual missing/modified seed, incorrect property/collection oracle, incorrect Type identity and changed conditional/header expectation fail; unsupported setup never passes/N/A.
4. Coverage retains every canonical required ID exactly once and reports all unresolved conditions honestly. Authored probe pass is distinct from whole-contract completion.
5. Pinned Node24.16/pnpm11.20 frozen install, clean → typecheck → build, full conformance plus relevant bdptest installed/Read/RU regressions, boundaries, format/lint, unchanged Read evidence verification, and a fresh packed clean-source observation all complete. Save structured/verbose logs and inspect child cleanup.
6. Freeze separate source commits, exact progress record and concise report; no public admission, profile evidence seal or outward write. Any actual product defect is reported with its failing request rather than silently weakening the row.
