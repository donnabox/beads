# Conditional Read representation repair — proposed, not authored

## Decision

Use an explicit, source-private **Transactional execution parsing context** selected by the existing canonical cumulative Transactional bundle factory and the validated private profile planner. Keep the row's `requiredProfile:"read"` and `read.http.conditional-reads` identity unchanged. The inherited requirement identifies what is being demonstrated; the validated cumulative target profile determines which setup and interleaved requests can be executed. Standalone Read still cannot author a successful mutation or an exact request body.

This is a necessary conformance implementation repair. The current frozen four-file draft does not parse; no passing execution is claimed. Preserve that draft, both actual review inputs and every prior measurement. No product semantics or user ruling is needed. Authoring requires root plan acceptance and the end/disposition of current immutable reviews.

## Source-established blockers and rejected alternatives

1. executable-manifest.ts840–853 restricts HTTP actions using the row's requirement profile;1003–1004 separately rejects its raw body. Both seed POST and later update POST are rejected. artifact-bundle.ts223 must still enforce exact canonical ID→requiredProfile correspondence, so changing the row to Transactional is not valid.
2. `prerequisiteScenario` is only an attribution/failure dependency, not another execution profile. parseActions761–771 still parses every action under the containing row's profile and requires prerequisites to be a prefix. runner.ts2842–2858 requires an earlier catalog requirement. The new canonical row is index0 and the second mutation is action12, after owned read assertions. This cannot express the intended after-mutation observation without changing the prerequisite model, ordering or lifecycle.
3. `ScenarioSetup` is closed fixture/requires metadata (executable-manifest.ts575–584), not a mutation script. Hiding setup or update calls in a harness would remove them from the exact ledger; an arbitrary actionExecutor would add a new unreviewed transport route rather than use the existing raw capability.
4. Cross-credit from another TX row would require changing the canonical one-row/one-binding preservation contract in transactional-cumulative.test.ts110–120, force shared scenario state across isolated sessions, and/or alter existing parent objects. It is a larger repair and would no longer preserve this requested independent Read scenario.
5. `negativeMethodProbe` means a negative POST405 observation, not successful setup. Do not apply it, change the requirement's profile, manufacture discovery, or change public admission.

## Narrow typed seam

Proposed new internal module export (not an existing API):

`parseTransactionalExecutableScenarioManifest(value: unknown, sourceLabel?: string): ExecutableScenarioManifest`

Keep the existing public `parseExecutableScenarioManifest(value, sourceLabel?)` and JSON loader signatures and behavior unchanged. Factor their existing body into a private core with an optional literal execution profile `"transactional"`; do not add a user JSON field, public options flag, general capability switch or schema version. The internal entry accepts only version2/catalogId transactional-v1. Malformed inputs still use the existing ManifestValidationError/path machinery. It does not create a bundle, execute anything, or grant admission.

Thread the explicit context through parseScenario into both parseActions/parseHttpAction and parseRequests/parseRequestInput so the two accepted v2 representations cannot disagree. Retain requiredProfile verbatim in parsed metadata and every existing unknown-field, accessor, body encoding, header secrecy/reserved-header, capture URL, conditional-reference and raw-target validation. Only unmarked successful mutation actions gain the cumulative Transactional request capability. Do not replace declared requiredProfile with effectiveProfile around the entire Read guard. In BOTH requests and actions forms, keep these independent predicates: (1) declared Read plus GET/HEAD/OPTIONS plus negativeMethodProbe:true is rejected even under private Transactional execution; (2) declared Read plus negativeMethodProbe:true plus any raw body is rejected even under private Transactional execution; (3) strict/default declared Read rejects an unmarked POST and all exact bodies; (4) private Transactional context can permit an UNMARKED POST with a valid raw body. A declared Read marked POST without a body retains its existing negative-probe meaning. Keep marker-type validation and other method restrictions unchanged. A test that only admits the new POST is insufficient: add discriminating rejects for both marker cases in both representations. No body/header/parser size or runtime/time limit changes.

In artifact-bundle.ts, after the caller's existing assertCanonicalProfileContractSelection and exact baseline Transactional check, choose the internal entry only when `cumulative?.profile === "transactional"`. All ordinary/public, development and cumulative Read+Update paths continue using the strict public parser in this repair. Do not generalize RU implicitly. Keep the subsequent exact ID/profile requirement check, version/catalog binding, copied source digests, deepFreeze and WeakSet brand unchanged.

The private planner is a second required caller: packages/conformance/test-support/profile-contract-plan.ts245–251 currently invokes the strict parser directly. Use the new entry there only after its existing canonical selection check, bound configuration/fixture check, matching selection/profile/History check and exact execution-configuration identity check (225–241), and only for baseline Transactional version2/catalogId transactional-v1. The input is already copied and bounded JSON. Preserve strict parsing and the existing carrierBlocker behavior for other profile/History/carrier combinations; do not turn a historically unsupported carrier into a newly accepted context or arbitrarily change it into a parser error. Preserve required ID/profile validation at283–293, fixture validation, digest, all missing obligations, ready-unexecuted state and claimEligible:false. The trusted selection/configuration selects capability, never a catalogId alone. Add targeted cumulative tests exercising this real planner, including wrong/forged configuration, profile and History refusals and unchanged unsupported-carrier reporting.

This adds no root/testing package export. Existing index.ts96–99 exports only the public parser/loader; leave them as-is. A directly parsed object cannot become a factory-branded bundle. runner.ts348–355 rejects a TX-branded bundle in the public runner and requires the private Transactional runner/profile;2792–2820 requires actual matching exact transport Scope/profile. These guards remain untouched. No runner.ts production change is planned.

## Indexed fixture repair and caller fanout

runner.test.ts6060–6070 expands eight cases, all constructing receipt/page probes from `manifest.scenarios[0]` at6167. Canonical insertion changed that unrelated template to Read. Even if the new context makes it parse, retaining that accidental coupling would disguise the fixture's meaning. Select the existing `transactional.receipt.pagination` row by exact ID; assert it exists and has requiredProfile transactional, then replace only its synthetic actions as before. Its observations already select results by `row.id` (6263), so no index-based assertion rewrite is needed. Preserve all eight titles, wire schedules, cancellation and authority checks. This selection does not mutate the canonical JSON parent row.

The frozen callsites.txt is a repository-wide source reference inventory, not an assertion that each call executes the TX fixture. Its test-support/profile-contract-plan.ts call was present but missed in the original analysis; this revision explicitly closes that substantive omission. A separate test-support-callsites.txt records that caller family. Public parser/loader consumers remain unchanged. Direct TX-manifest references are runner.test.ts6078, transactional-binding.test.ts29 and profile-contract-plan.test.ts302, in addition to the author helpers that load it by assembled paths. transactional-binding.test.ts39 uses index0 only as a scenario/session token; it has no synthetic mutation template tied to that row. Verify its existing three socket cases; no edit proposed. profile-contract-plan.test.ts selects its command-minimum row by ID, so preserve it. The new conditional-specific index0 checks intentionally assert catalog order and remain. Recheck all relevant callsites during implementation; do not bulk-rewrite unrelated tests.

## Proposed ownership: nine files total

Retain the four original owned files, whose two JSON fixtures stay byte-identical to the current frozen draft:

- packages/conformance/test-support/transactional/manifest-v1.json
- packages/conformance/test-support/transactional/coverage-v1.json
- packages/conformance/src/transactional-cumulative.test.ts (add bundle/identity/firewall assertions here)
- apps/bdptest/src/transactional-matrix.test.ts

Add only:

- packages/conformance/test-support/profile-contract-plan.ts — same validated baseline TX parsing context, preserving other carrier behavior

- packages/conformance/src/executable-manifest.ts — internal parsing context
- packages/conformance/src/artifact-bundle.ts — factory selection of that context
- packages/conformance/src/executable-manifest.test.ts — parser safety cases
- packages/conformance/src/runner.test.ts — stable-ID receipt fixture and runner firewall assertion if not already sufficient

No catalogs/schema/protocol/server/runner production/export/admission/Read evidence artifact edits. No preserved32 raw object or parent reconstruction changes. Existing14actions/198assertions,14positive+4control HTTP/fourchildren, canonical footprint and42 derivative data remain unchanged by this seam. Additional CPU and existing eight-case synthetic HTTP regression work is separate validation, not silently added to installed cohort totals. Freeze exact expanded test selections and their observed-source request bounds before execution.

## Required discriminating safety checks

Use bounded source fixtures; test both actions and requests forms where affected.

- In both requests/actions forms, explicitly reject declared-Read GET/HEAD/OPTIONS with a probe marker and declared-Read marked POST with a raw body even in the private TX parser. A marked no-body POST keeps its old interpretation. The unchanged public parser and JSON loader reject this genuine Read row's unmarked POST and raw body, including when its untrusted catalogId says transactional-v1. Existing Read negative405 probe remains accepted only under its existing rules. GET/HEAD/OPTIONS and all raw field prohibitions remain unchanged.
- The internal context accepts the same valid v2 raw POST bytes while preserving `requiredProfile:"read"`; rejects wrong version/catalog, unknown fields, invalid raw encodings, secret/executor-owned headers and invalid markers through existing validation. No context is inferred from caller-controlled catalog text.
- Actual canonical cumulative TX factory AND validated private planner accept all current rows and exact bytes; inherited Read remains Read. Existing cumulative callbacks at185–200 and310–320 must exercise the repaired planner, preserving unobserved requirement accounting and transportAdmission without qualification. Wrong-profile row, forged/noncanonical selection, History, ordinary/development/RU factory misuse still fail. The public runner refuses the TX bundle before prepare/execute; private runner rejects mismatched Read profile/configuration before dispatch. Keep claimEligible:false.
- Preserve full parent bytes/32 raw objects and all42 derivative identities. Run the current targeted static selection, affected parser/bundle checks, the eight receipt authority cases and existing three binding socket cases. Actual native/Claude code review must include this added parser/bundle boundary before installed release.
- No automatic retries or test edits to manufacture a passing admission outcome. A new runtime failure is independent evidence requiring diagnosis.

## Read cohort seal/source projection impact

Do not relabel or regenerate old Read evidence as a side effect. scripts/generate-read-cohort.mts332–343 records schema/validator/runner/executor provenance; this patch changes executable-manifest.ts/artifact-bundle.ts, neither a current input pin in that table. cohort-verification.ts303–320 checks digest shapes and the sealed Read schema projection;383–404 compares current Read catalog, Read manifest and per-target fixture bytes. schema-read-projection.ts1–60 defines the sealed schema-definition projection, not a source-code projection. All those Read artifacts/roots/definitions and runner bytes are proposed unchanged. scripts/read-cohort-evidence.mjs50–61/130–164 gates committed evidence inputs and the historical run→evidence ancestry delta; it is not an assertion that current parser source equals the old packaged payload.

Therefore the existing verifier is not mechanically invalidated by this source-private-only change, provided the public strict path remains equivalent. That is not a fresh current-payload Read observation. Preserve the original historical source/runner/payload attribution, run ordinary Read parser regressions and the existing Read seal gate after a correct build, and stop if the gate detects real input/projection drift. Do not patch evidence constants or comparison logic. If implementation expands the public Read behavior or changes a sealed input, reassess qualification explicitly rather than using this conclusion.

## Review and execution boundary

Planning only. Frozen source binding b630e46d and ongoing reviews are untouched. This report supersedes only the claim that the present inherited row already has a usable parser representation; all other supported capture corrections and bounded workload remain. Root should accept this narrow representation plan, then authorize a separately preserved successor and its expanded code council/finite validation. No actual parser invocation, tests, build, imports, services, source edits, commit or push occurred in this investigation.
