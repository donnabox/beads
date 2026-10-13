# Native review — profile contract council repair

Reviewed candidate `28b5da2092e57cac53fd67fab38ab9e95b64df28` against direct parent `9ea408958e08be38cc4e586a4846caf3d6ef6874` in `vickie-profile-contract-council-repairs`. The author checkout was clean and remained untouched. Read the author repair report and all eleven findings in `profile-contract-claude-review.txt`. Review covers all nine changed files and their existing artifact, runner, fixture, catalog-selection and package-export boundaries.

## Verdict

The original eleven findings are addressed within the intended source-only planning scope. One new Medium finding remains in the prerequisite repair; close it before treating this repair as complete. No public Read artifact/runner admission guard was weakened, and no RU/TX qualification or new evidence claim was introduced.

## New finding

1. **Medium — exponential prerequisite evaluation blocks the synchronous planner on a small valid manifest.** `packages/conformance/test-support/profile-contract-plan.ts:388–400` recursively recomputes each dependency without memoizing its readiness. When row n depends on rows n−1 and n−2, the same completed subgraphs are visited repeatedly, including again from each requirement. The earlier-in-catalog constraint prevents cycles but does not prevent this shared DAG. A valid manifest containing the first 32 canonical Read obligations, each with up to two prerequisite HTTP actions and one own HTTP action, is only **23,568 bytes** and took **4,979 ms** to plan under Node 24.16.0; all 32 rows were correctly classified `ready-unexecuted`. The same run measured 28 rows / 20,611 bytes at 664 ms. A preceding independent run measured 16/20/24/28 rows at 4/13/79/650 ms. Thus byte/depth parsing limits do not bound this synchronous repeated work adequately. This is a source-only planner availability/performance defect, not a claimed public network vulnerability.

   **Fix:** compute dependency readiness once per scenario (memoized DFS, or a catalog-order pass using the existing earlier-than constraint), retaining rejection/blocking for absent, retired, late, unready and cyclic dependencies. Add a shared-DAG regression that exercises the full Read obligation set without relying on a fragile tight timing threshold. The probe uses the actual canonical loader, actual fixture bytes and existing executable-manifest parser; no production source was patched.

   **Reproduction:** private archive at `profile-contract-native/source`, test `packages/conformance/src/native-prerequisite-probe.test.ts`; run `/Users/dbox/.nvm/versions/node/v24.16.0/bin/node node_modules/vitest/vitest.mjs run packages/conformance/src/native-prerequisite-probe.test.ts --reporter=verbose --maxWorkers=1`. Logs: `profile-contract-native/prerequisite-probe.log` and `profile-contract-native/prerequisite-probe-32.log`. The initial probe used invalid action IDs and failed at manifest parsing; `prerequisite-probe-invalid-id.log` preserves that harness error separately. Corrected probes use valid hyphenated IDs and pass validation. Larger sizes were not timed; no extrapolated duration is presented as observed evidence.

## Original council disposition

1. **Closed/preserved:** own-data snapshots and primitive binding fields reject proxies, getters, boxed/coercible identities. Fixture and manifest wrappers now receive equivalent checks; detached immutable bindings remain factory branded.
2. **Closed:** normative RU index explicitly mirrors `read-update.discovery.document` retiring `read.discovery.document`; TX explicitly inherits that retirement in addition to the twelve direct TX→RU replacements.
3. **Closed:** both replacement discovery rows retain citations for optional-contract semantics and omitted `order` → `canonical-uri`, with individual removal counterfactuals. Contradicted closed-profile shapes remain retired.
4. **Closed:** unsupported catalog/profile carriers receive `carrier-unsupported`; successful parsing alone is not readiness. This conservative source-only planner does not admit a new writer bundle.
5. **Functionally closed, new performance finding above:** actual parsed dependencies now block missing, retired, unready, late and cyclic prerequisite chains. Blocked IDs remain inspectable even when carrier status takes precedence.
6. **Closed:** ordinary selection loads real catalog files. Injected negative-control inspection returns unbranded selections that cannot enter planning.
7. **Closed:** actual directory JSON filenames must exactly match the explicit core/companion/optional classification.
8. **Closed:** fixture ID, digest and capabilities derive from captured fixture bytes through the existing parser; independent claims are rejected. Fixture oracle references are validated against the captured fixture.
9. **Closed:** canonical Scope and fixture identifier/capability checks reuse existing runner/artifact helpers, eliminating the divergent predicates.
10. **Closed:** intrinsic byte copying and bounded UTF-8/JSON parsing are reused for fixture and manifest. Getter/proxy/array-like/coercion and buffer-mutation controls remain meaningful.
11. **Closed:** the 81-row editorial audit is checked against exact catalog row order, retirement disposition and replacement IDs; preserved discovery citation controls are explicit.

## Guard and citation boundary assessment

The only `artifact-bundle.ts` and `runner.ts` changes are internal named exports and comments on existing helper functions; their bodies and artifact/runner admission branches are unchanged. Package-root exports and the package exports map are unchanged. The source-only planner is still outside those public entry points and always returns `claimEligible: false`; every selected obligation remains unobserved, including ready rows. Read's sealed artifact, matrix and evidence constant are untouched by this delta.

Catalog changes are confined to the two discovery replacement rows and add the shared clauses lost at retirement. The normative prose mirrors the already-declared retirement chain; it does not introduce a new Type rule, retire another Type/profile obligation, or expand advertised profiles. Existing inherited Type, numeric and owned-companion selection remains in the canonical cumulative catalog; companion duplicates must match exactly. History remains an explicit optional catalog. Catalog/profile and citation tests passed. This assessment does not claim that a future cumulative RU/TX executable manifest has been implemented or qualified.

## Independent verification

- Exact candidate archived outside the author checkout, with workspace links pointing to private packages; Node **24.16.0**, repository pnpm **11.20.0** dependency tree.
- Private workspace TypeScript build passed before tests (`profile-contract-native/build.log`).
- Complete `packages/conformance/src` run: **783 tests / 22 files passed**, including artifact, runner, catalog, citation, selection and all 53 planner tests (`profile-contract-native/conformance-tests.log`).
- Standalone valid shared-DAG probe passed its semantic assertions and reproduced the slowdown as reported above.
- Author reports 816 tests / 25 files and unchanged sealed-evidence verification; these are author evidence, not additional independently executed tests here. No full workspace suite, process smokes, evidence regeneration or served-profile qualification was performed for this narrow review.

No source edits, outward actions or new evidence claims.


## Native closure addendum — prerequisite DAG repair

**Finding 1 closed at `1f2d75fbaa67f631e72bd79384c829d7588b7277`.** Verified clean frozen checkout `vickie-profile-contract-dag-repair`, direct parent `28b5da2092e57cac53fd67fab38ab9e95b64df28`, and read `profile-contract-dag-repair-report.md`. Exactly two files change: the source-only planner and its tests. No new finding in this bounded successor.

The recursive traversal is replaced by a single pass over canonical catalog order. Each selected row stores one boolean readiness result; each prerequisite edge consults that result. Because valid dependencies must be strictly earlier in that same order, this preserves the previous predicate while removing repeated shared-subgraph traversal. Missing/retired IDs have no true stored result, missing setup/applicability and unsupported carriers store false, and late/self/cyclic edges fail the existing earlier-than rule. Consumers of blocked dependencies remain blocked. Requirements still expose immediate blocker IDs, carrier status retains display precedence, and all requirements remain unobserved with `claimEligible: false`.

The new deterministic regression covers every canonical Read obligation with the two-predecessor DAG. Its Map-lookup budget fails fast rather than using a timing threshold; the spy restores in `finally`, and the test also asserts readiness, empty blockers, complete unobserved IDs and false claim eligibility. Author red/green evidence is recorded separately in the author report; I inspected the regression and independently ran the green source.

Independent exact-source archive: `profile-contract-native/source-dag`. Pinned Node **24.16.0**, existing pnpm **11.20.0** dependency tree. Private TypeScript build passed (`profile-contract-native/dag-build.log`). Relevant suite and unchanged native probe: **55 tests / 2 files passed** (`profile-contract-native/dag-native-focused.log`), comprising all 54 planner tests plus the native probe. Shared workspace dependency implementations are unchanged by this two-file successor; the repaired planner and tests are loaded directly from the new archive.

The exact prior probe and input construction were copied unchanged: **28 rows / 20,611 bytes: 2 ms** (previous 664 ms); **32 rows / 23,568 bytes: 1 ms** (previous 4,979 ms). Both still classify every supplied row as ready-unexecuted. These are observed rounded timings, not a performance guarantee. The full-obligation deterministic regression also passed. Existing focused tests cover missing/unready/cyclic prerequisites, real retired discovery prerequisites, carrier precedence, unsupported versions/profiles and binding/byte hardening.

No author checkout edits, outward actions, repeated full workspace suite or served RU/TX qualification. The original eleven closures and Read admission/citation boundary assessment remain unchanged. This addendum supersedes the earlier open Medium verdict solely for candidate `1f2d75fbaa67f631e72bd79384c829d7588b7277`.
