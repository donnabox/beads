# Artifact-set validator — step1 frozen checkpoint

Head **3ae9849854e731049782d3962ebf0b2421b7ceb0**, direct child of approved main **1f9d7dbc401f6a265cf1ce0bdfea0fe493cd79a5**. Checkout `/Users/dbox/repos/wt/vickie/vickie-artifact-set-validator`, branch `codex/vickie-artifact-set-validator`, origin `https://github.com/gastownhall/bdp.git`. Remote main and role session2026-10-08T20:26:35Z-81e5 were checked before work. Clean local commit; **not pushed**, root owns publication/integration. Later ordinary reference integration b204 has not been merged into this isolated step1.

Authority: revised segmentation plan d5377fb47a7c64f424c77889a5e777c1cbcb36121a85c36bcc6394ad54ccf610 and root's step1-only acceptance. Sole plan remains design.md.

## Delivered scope

Three files only:

- `packages/conformance/test-support/read-update-artifact-set.ts`: source-private closed two-source index/factory, canonical ownership and coverage validation, immutable branded result, detached exact-byte copy-out for future planner use. No filesystem/network I/O.
- `packages/conformance/src/read-update-artifact-set.test.ts`:56 tests using tiny index/fixture/coverage/manifest inputs and bounded parser-limit fixtures.
- `packages/conformance/test-support/profile-contract-plan.ts`: **one approved export keyword only** on existing ownData; its body/behavior is unchanged. Imported directly by the new private helper, not reexported through testing/public entrypoints.

No app adapter, real index pointing to C, actual fixture migration, coverage promotion, runner/executable-format change, runtime/store/schema/catalog/public export change, or historical Read reseal. Test index bytes deliberately name tiny synthetic sources. Actual two-source index and C remain later steps.

## Boundary and identity

The index is version1 with exactly two nonempty shards, unique bounded IDs and flat repository-relative `matrices/<name>.json` filenames. Absolute/URL/traversal/recursive imports and unlisted/missing supplied sources are not accepted. Index parsing and source descriptors use existing hardened ownData and copied intrinsic Uint8Array bytes, rejecting accessors/proxies without invoking getters. All caller bytes are captured before bundle construction; copy-out returns fresh byte arrays and cannot mutate retained identity.

Every shard independently enters the existing cumulative Read+Update factory with the same genuine full canonical RU/no-History selection and shared fixture. Factory manifest parsing, fixture-pointer checks, fixture identity, exact catalog map and WeakSet provenance remain intact. Scenarios must be canonical-order nonempty members with exactly one owner. Every prerequisite, including requests, HTTP, client and lifecycle actions, must be an earlier implemented row in the same shard. Cross-shard, missing and later prerequisites fail before any potential execution.

Coverage must contain every canonical required ID once in exact order. Authored rows require exactly one binding with the complete nonempty ordered action/assertion list; requests use their real assertions rather than a vacuous actions fallback. Omitted rows must have empty bindings and wholefalse. Parsed whole flags are retained as **authored review decisions only**; this validator grants no whole credit or execution/qualification claim. Each requirement is exposed with one owner or explicit null. Independent reviewed-whole-ID gate remains unchanged.

Computed SHA256 set identity binds exact index, coverage, shared fixture, canonical catalog digest map and ordered shard IDs/manifest digests/scenario IDs. Renaming a source changes index/set identity but leaves its unchanged manifest digest intact. Each genuine child bundle remains separate; no synthetic aggregate manifest or report is constructed. Structural clones of bundles/sets do not acquire brands. Outputs and inventory are deeply frozen; exact source bytes are retained privately and returned only as fresh copies.

## Limits and accounting

Unchanged existing per-source boundaries: each manifest4MiB/100,000nodes, fixture1MiB/100,000nodes, coverage1MiB/100,000nodes, depth128 and container10,000. New index16KiB and exactly2entries, also parsed through existing depth/node/container checks. Shard count cannot grow silently. Unique captured manifest inputs are at most8MiB/200,000nodes; fixture+coverage each retain their own1MiB cap. Output sourceBytes counts **only unique captured inputs**, not total allocation.

Tests accept two separate60,000-node manifests (120,000 total) and reject a single100,001-node manifest. Exact byte ceilings accept boundary and reject +1 for index, fixture, coverage and manifest. Fixture/coverage depth/container/node overflow controls also fail at existing parser gates. No merged manifest or increased per-source cap is used.

Fixed parse/copy multiplicity still matters: captured source snapshots are retained privately; the existing factory separately copies/parses each manifest, serializes/copies/parses the combined catalog twice, and copies/parses the shared fixture twice. Canonical selection was loaded separately. Index and coverage parse once here; source copy-out may allocate additional detached bytes. Planner/configuration parsing, retained actual HTTP observations and reporting have not yet been introduced. No aggregate heap bound or installed-process RSS claim is made from source caps. Later app migration must measure actual aggregate resident memory/report bytes/time under the accepted plan.

## Verification

Pinned Node24.16.0/pnpm11.20.0, maxWorkers2. Frozen install completed. Final typecheck, clean build, dependency boundaries, format, changed-file lint and diff whitespace checks all exit0. Lint reports21 warnings (test any/non-null assertions and validated-path non-null assertions), no errors; not a warning-free claim.

- New validator tests: **56/56 pass**,418ms including startup, first focused run (`initial.log/json/exit`).
- New plus unchanged artifact-bundle, cumulative-artifact, profile-contract-plan, runner and transactional-cumulative regressions: **402/402 tests across6files pass**,970ms (`tests.log/json/exit`).
- Unchanged historical Read verifier: **evidence.verified**,89rows,47 projection definitions/46reachable (`evidence.log`, exit0).

Initial and second typecheck attempts found only new test-builder typing issues (unused import and union property narrowing). They are fixed; failed logs are retained as `typecheck-initial.*` and `typecheck.*`, followed by passing `typecheck-final.*`. No test was interrupted or timed out. No service/installed target was launched by this step. No full repository suite, installed pack run or combined baseline+C observation is claimed.

## Remaining work

Root/native/external review of this frozen helper first. Then ordinary merge the approved reference integration; remeasure its actual manifest/fixture/coverage union and preserve whole28/101 decisions. Only a later authorized step should adapt the app, bind actual per-shard routes/configuration/cleanup, keep one90s aggregate controller, preserve baseline missing-plan witnesses, restore C's exact independent oracles, and execute fresh installed evidence. Report association, second-shard readiness mismatch, abort/drain and successful-summary publication tests belong to that later app slice. Existing runner prerequisite-failure/filter regressions passed. The full plan’s dedicated dependent-only scenarioSelection/not-run witness is not newly added here; retain it for the later execution slice. This step adds explicit preflight closure without modifying the runner.
