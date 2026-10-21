# Supporting plan — bounded Read+Update artifact segmentation

**Proposal for council/root review; no implementation.** This supports the sole plan at `/Users/dbox/repos/wt/vickie/bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md`, Autonomous profile completion/U3. It creates no competing authority or normative BDP change. Remote session guard passed before inspection. No outward posts or product-source edits.

Source inspected: frozen `579019b676568345f93bb424b2fa6f5774211df2`, parent8e12162, with root's separately reviewed comparator6f3e3bd treated as an integration input, not an already tested merged tree. Independently counted579:96,446 nodes,3,601,051 bytes,49 authored rows; comparator adds662 nodes, giving97,108 and2,892 remaining under100,000. Old C draft remains at `read-update-reference-evidence/deferred-c-draft`; its3154 compressed nodes cannot fit that combined manifest. The original100,400-node parser failure remains a failure before target execution. Public admission, schema/catalog/spec semantics and historical evidence remain unchanged.

## Recommendation and alternatives

Use **multiple ordinary version2 manifests**, each bound by the existing source-private cumulative RU factory, with a small source-private artifact-set index and validator. Execute them through unchanged runner calls and retain the separate genuine reports. Add one private progress record that binds the set, configuration and report ownership; it is not a ConformanceRunResult or a qualifying cohort artifact.

| Option | Assessment |
| --- | --- |
| Normal v2 shards plus bounded private index | Recommended. Existing parse limits, fixture-oracle validation, factory WeakSet provenance, exact execution and per-report digests remain intact. New logic only owns set completeness/uniqueness and real-run association. |
| Add `imports`, file references or segments to v2/v3 executable format | Unnecessary first step. `executable-manifest.ts:293` closes root keys to version/catalog/scenarios. Imports would require new parser, resolution/security/cycle rules, digest semantics and every downstream consumer to understand expansion. No current row needs cross-file action execution. |
| Parse shards then fabricate one combined manifest/bundle/report | Reject. It sidesteps the aggregate parser boundary, cannot truthfully use a single source-byte manifest digest, bypasses WeakSet ownership if structurally forged, and blurs actual run boundaries. |
| More constant deduplication/cap increase | Not the needed mechanism. Existing equivalent fixturePointer compression remains useful; limits must not silently grow and required cases cannot be trimmed to fit. |

This is a transport-test artifact layout change, not protocol batching, target configuration or public evidence admission.

## Existing seams and constraints

- `artifact-bundle.ts:105–124` creates a cumulative RU bundle only from a canonical factory-loaded baseline RU selection. `154–250` copies source bytes, parses each bounded source, requires v2/read-update-v1, checks every row against the complete canonical required set, binds fixture IDs/oracles and hashes exact bytes. Bundles are frozen and registered in the factory WeakSet (`87`, `250`, `380`). Retain this path unchanged for every shard.
- Limits at `artifact-bundle.ts:17–23,394–469`: private cumulative RU manifest4MiB; ordinary fixture/catalog1MiB;100,000 nodes, depth128 and10,000 children per container. TX remains1MiB. These are per-source limits, not permission to create an unbounded list of sources.
- `runner.ts:340–410` already accepts `scenarioSelection`, preserves canonical catalog order and records that selection, artifact digests and exact configuration digest in each v4 report. It always says claimEligiblefalse. `resultsById` is local to one call; `2781+` refuses dependencies that did not pass in that call. Do not seed it with a previous report.
- Actual current prerequisite graph: seven inherited rows depend on `read.discovery.scope-service-desc` (canonical navigation, Types inventory, both collections, GET/HEAD,405,cache/CORS). They all stay in the original shard. No new C prerequisite crosses a shard boundary.
- `cumulative-artifact.test.ts:109–204` currently assumes one manifest when checking canonical order, complete coverage bindings and the reviewed whole-ID list. This is the principal static consumer to update, retaining the same canonical requirement/coverage authority.
- `profile-contract-plan.ts:215–365` plans one manifest against a factory-bound served configuration. Reuse once per shard; absent rows in one shard are not global omissions. No merged synthetic manifest is needed for planning.
- `read-update-matrix.test.ts:146–309` derives routes from actual manifest actions plus the operation routes, captures installed metadata, binds fixture bytes and checks the planning digest. `326+` verifies each actual scenario instance against the same served baseline. `765+` produces a private report. Current `uncoveredIds` describes reviewed whole coverage, not fresh execution success; the new set summary must keep those two notions distinct.
- `cohort-artifact.ts` is the historical two-target Read projection with legacy report, package and redaction contracts. It is **not** a suitable template for relabeling RU results. Do not edit it, the Read seal/constants, verifier, public report schema, public exports, diagnostic RU or TX carrier.

## Small first artifact set

Initial delivery has exactly **two** shards, with a private hard maximum of two for this first slice. This allows at most8MiB/200,000 parsed manifest nodes across sources, each still individually subject to4MiB/100,000. A third shard requires an explicit reviewed bound change; no arbitrary growth loop is introduced. This is enough to unblock C and continue filling the second source; it does not promise all remaining requirements fit forever.

1. Baseline shard: retain the combined reviewed579+6f3 manifest at `matrices/read-update-v1.json`. Preserve every existing scenario block byte-for-byte. Do not repartition its dependency group or change existing assertions/whole flags.
2. Reference continuation shard: new `matrices/read-update-reference-continuation-v1.json`, ordinary `{manifestVersion:2,catalogId:"read-update-v1",scenarios:[...]}`. First contains the preserved C scenario only, revalidated against the new exact source. C remains wholefalse because invisible-subject coverage is absent.
3. One shared fixture source, extended only with C's authored fixed bindings and oracle constants. Preserve existing values; identity-preserving new names resolve any collision. The fixture digest changes, so both shards need fresh observations. It is not permissible to attach the old baseline report to the new fixture digest. The preserved full A+B+C draft fixture is28,753 bytes/742 nodes versus current26,370/687 (only2,383 bytes/55 nodes growth); remeasure the exact merged union rather than copying unrelated draft content. Keep its independent1MiB/100,000-node boundary and unresolved-pointer/collision controls. Sharding manifests must not become an unbounded shared-oracle escape hatch.
4. Keep one complete coverage inventory over canonical required IDs. C changes omitted→bounded observation with exact ordered assertion bindings and wholefalse only after actual execution. No duplicate shard-local coverage authorities. Shard ownership is derived from manifests, not another manually maintained list of requirement IDs.

A small new private index (suggested `test-support/read-update-artifact-set-v1.json`) declares version1 and the ordered two shard IDs/source filenames. Source locations are closed repository-relative JSON names, no URLs, absolute paths, `..`, optional globbing, import recursion or runtime discovery. It is an artifact-set index, **not a new executable-manifest version**. The loader resolves only these named bytes; the validator itself does no filesystem/network I/O. Shared fixture/coverage are explicit captured source inputs rather than shard overrides.

Bound the index to16KiB and two entries, using the existing bounded source/JSON helpers. Bound coverage and shared fixture independently at the ordinary1MiB/100,000-node limits. Preserve all existing node/depth/container controls. No new opt-in flag widens other factories. Record actual summed source bytes/nodes and peak resident test memory in the implementation report; worst-case JSON heap is not claimed merely from input-byte bounds.

## Private validation and identity contract

Add a narrowly scoped source-only helper (suggested `packages/conformance/test-support/read-update-artifact-set.ts`, plus tests), not a public package export. It composes existing factories rather than implementing a second manifest parser. If a set object is consumed later, retain a private factory brand and immutable detached data; structural lookalikes must not acquire evidence authority.

Inputs: canonical RU/no-History selection, exact index bytes, exact shared fixture bytes, exact coverage bytes, exact declared shard source bytes. Capture inputs once; reject proxies/accessors or use the existing hardened data-boundary helper through an internal export. Never validate a caller object and reread it later. Do not duplicate unsafe/coercive digest checks. Reuse `copySourceBytes`, bounded JSON and existing fixture/manifest validation; source descriptors and keyed source maps require closed, detached data.

Validation before launching a child:

- Index IDs/source names unique and valid; required declared files present; no unlisted supplied file, duplicate or empty shard.
- Each shard independently becomes a genuine `createCumulativeReadUpdateArtifactBundle` result for the **same canonical full selection** and shared fixture. Same catalog digest map, v2, RU profile, no History. Each shard's scenario order equals canonical order restricted to its membership.
- Each scenario has exactly one shard owner. Duplicate ownership is rejected even for byte-identical rows; no last-wins merge or double counting.
- Coverage IDs equal the complete canonical selected required-ID list in order, including inherited/companion obligations after retirements. Every bounded-observation row has exactly one owned scenario and its full ordered action/assertion binding; every omitted row has no scenario, empty bindings and wholefalse. No unknown/retired row, orphan coverage binding, omitted-but-executed row or executable row absent from coverage.
- Every prerequisite is earlier and owned in the same shard. Reject cross-shard references explicitly. Do not duplicate prerequisite scenarios or treat another run's success as an executable dependency. Any future need for shared dependency execution is a separate plan.
- Shared fixture identity/capabilities and every applicable fixturePointer remain validated by each ordinary bundle. No target-produced value becomes an authored oracle through this layout.

Compute—not caller-declare—a detached canonical set descriptor and SHA256 from index digest, canonical exact catalog-file digest map, coverage digest, shared fixture digest, and ordered shard entries `{id, manifestDigest, scenarioIds}`. Expose per-shard bytes/digests/budgets in the private progress record. A set digest is additive provenance, never a substitute manifest digest. Filenames are diagnostic locations; exact contents determine identity. Preserve original report/artifact digests in every association.

The full required inventory is always canonical. For each required ID, the set gives either one owned plan or an explicit absence. Changing the index or coverage to omit work changes the set binding and cannot yield a full claim.

## Actual execution and reporting

Make the smallest app-test adapter change: parameterize the existing `run` path with an already validated shard source/bundle and common fixture input. Preserve existing single-manifest countercontrol calls and assertions; do not duplicate the installed launcher/exact client/HTTP harness. New controls can select their owning shard. Extract a helper only where necessary to reuse actual execution; no broad harness refactor.

For each shard run, pass **exact owned scenario IDs** through the existing scenarioSelection, rather than running every canonical requirement once per shard and accumulating duplicate missing-plan rows. The canonical full catalog still resides in each bundle. `runConformanceMatrix` records the precise subset and validates prerequisites normally. Catalog order remains within each shard. Cross-shard scheduling order is only index order, never a normative operation ordering assertion.

Run the two shards sequentially under one existing90-second cooperative aggregate budget and120-second outer backstop; retain5-second exchanges and30-second focused tests. Do not reset90 seconds per shard or launch concurrent authorities to hide runtime cost. Check abort before each shard/start and after return before writing a passed progress record; use existing bounded bootstrap/prepare/cleanup behavior and drain failure paths. Record any unrun shard after failure/abort as unobserved, never as absent-but-passed. A slow/failing run fails qualification; measure before proposing scheduling changes.

One shared served-configuration baseline must be checked against **every actual scenario instance in both shards**, reusing `bindServedProfileConfiguration`/`assertSameServedProfileConfiguration`. Bind Scope/profile/History/sourceCommit, installed payload/contract/policy/auth identities and the shared fixture digest/capabilities. Also compare the actually observed runtime configuration identity captured today. Stable canonical Scope does not require sharing a mutable database: existing per-scenario fresh authorities and physical listener ports remain legitimate, with actual instance receipts retained. Port/directory are not configuration identity fields (`installed-reference-configuration.ts:175–185`). Do not mix separately observed policy/schema cohorts into this same-config set.

Exact route tables are derived **per shard** from that shard's actual actions and necessary directory operations, using the existing gate. Their digests will generally differ. Bind each report to its own captured exactConfigurationDigest; do not force a common union wildcard allowlist or relabel route digests. Same served configuration and same exact client route configuration are distinct facts. Known dynamic Resource captures stay within the existing route validation.

The new private progress record contains:

- literal claimEligiblefalse, set/index/catalog/fixture/coverage digests, exact source and measured installed configuration baseline;
- per-shard source/digest, requested IDs, actual selected IDs, actual report digest/reference, exact route digest, actual instance receipts, outcome/abort/cleanup state;
- all canonical required IDs, each mapped to its single actual report row or an explicit unobserved/unplanned reason;
- distinct counts for authored, observed pass/fail/not-run, reviewed whole, and currently uncovered. Whole credit needs both a reviewed binding and an actual passing row at the bound source/configuration. Partial rows do not become whole merely because their shard passes.

Do **not** flatten these into a fabricated ConformanceRunResult, add a public report version, or copy old observations into a new run. Exact child reports remain the evidence. The private summary can point into them and machine-check uniqueness/completeness; it cannot manufacture successes or applicability. Unknown/missing reports fail set completion. Countercontrols remain diagnostic runs with their own changed source/fixture digest, not mixed into the positive progress record.

## Bounded ownership and integration

Allowed after approval:

- New private set helper + focused tests + index; source-only `test-support/testing.ts` reexport if needed.
- `cumulative-artifact.test.ts`: reuse set validation for canonical union/coverage checks, preserve independent public Read/development/TX budget/brand tests and the exact reviewed whole-ID gate.
- `apps/bdptest/src/read-update-matrix.test.ts`: narrow shard source adapter, aggregate execution/progress association, existing and new controls. A small private app helper may be extracted only to avoid copying the current run path.
- New C manifest; shared reference fixture additions; C coverage row only. Existing combined baseline manifest and old coverage entries unchanged.
- Dedicated implementation/progress note. Root's external packed supervisor may need new private record output handling; no new public runner/report format.

Do not edit `runner.ts`, `executable-manifest.ts`, public `index.ts`, package exports, normative schemas/catalog/spec, runtime/store/evaluator, TX/Snapshot/SSE, historical `cohort-artifact.ts`/evidence/constant. `artifact-bundle.ts` should need no behavior change; any proposed helper visibility change requires root review first. `profile-contract-plan.ts` remains single-source planning; aggregate its existing per-shard outputs by ownership rather than teaching it an import language.

Suggested commits: (1) private set validation/negative controls using tiny test-only manifests; (2) narrow existing installed adapter and genuine segmented progress; (3) restore C exact preserved scenario/oracles/controls, retaining wholefalse. Root picks the exact ordinary merged579+6f3 base and owns publication. Preserve all frozen ancestors and historical logs. Promotion of A/B/admission is a separate reviewed input, not a side effect of segmentation.

## Negative controls and done-when

1. Factory tests reject forged/trimmed canonical selection, wrong profile/History/v1, ordinary public factory use, structural set/bundle clones, duplicate source/row, missing shard, unlisted shard, excess shard count, unknown/retired row and extra/missing coverage binding. Input mutation after construction cannot change contents or digests; accessor/proxy controls execute no getter.
2. Exact boundary tests prove each manifest still rejects4MiB+1 and100,001 nodes and shared fixture/coverage retain1MiB limits; ordinary Read/development/TX stay unchanged. Two valid shards may exceed100,000 total **only as separate artifacts**, never by constructing a merged manifest. Test set/index bound too.
3. Missing inherited obligation remains visible even if every authored shard passes. Duplicate requirement cannot count twice. Cross-shard prerequisite fails preflight; in-shard failed prerequisite still blocks dependent execution through unchanged runner. Current seven-edge group stays intact.
4. One-byte manifest/fixture/coverage/index mutation changes the proper binding or is rejected. Report source/manifest/fixture/catalog/route/config mismatch, duplicate report, missing report and failed required row cannot be accepted as complete set observation. Fixture oracle corruption still fails the named actual assertion.
5. Actual installed baseline plus restored C execute under the same captured configuration. Reintroduce C's preserved now-satisfiable retained404, fresh-key success, alias-delete retry and independent tail/afterstate controls; no old draft success substitutes for these new observations. C remains partial with invisible-subject gaps named.
6. Short cooperative abort tests cover before second-shard startup and during second-shard real work, with drained children and failed/unobserved status; no per-shard timer reset, blanket timeout increase or abandoning Promise.race.
7. Clean pre-build typecheck, build, focused artifact/runner-regression/app tests, boundaries/format/lint, then fresh packed baseline+C runs with exact source/payload/config/index/per-shard report digests. Existing Read evidence verifier remains unchanged. Root runs combined CI/smokes as appropriate; no claim from an interrupted or parser-rejected run.

Done when these predicates pass, source/byte/node/route counts and failures are recorded, all original scenarios/assertions are unchanged, C is genuinely observed yet partial, and independent council accepts the source. This unblocks continued bounded authoring in a second artifact; it does not close any unimplemented protocol surface or qualify Read+Update.

No product ruling is required for this source-private layout. Main risks are incomplete union accounting, duplicate credit, crossing prerequisite/run boundaries, stale report/config reuse, uncontrolled aggregate memory/time, and accidental widening of public formats. The proposal addresses these with explicit ownership, existing factory/runner calls, bounded set size and fresh observations rather than a new execution framework.
