# Startup focused-source failure diagnosis

The smallest contract-preserving correction is **test-only**: replace refusal-loop row10's unreferenced schema fixture with the existing explicit `artifacts()` bundle, placing the missing external `$ref` in its actually referenced `schemas/item`. No loader, RU policy, aggregate inventory or compiler change is supported by this failure. The one-hunk recommendation and inverse are inert artifacts here; nothing was applied or run.

The actual validation4 failure remains43 passed /1 failed /389 skipped. The failing callback is `delegates RU credential, policy, provisioning, artifact, locality and aggregate refusals`. Its `failure()` helper throws its own `expected local preparation failure` at test216, called from loop541. That proves an awaited loader call resolved in the refusal loop, not that a compiler threw the wrong error. The test does not log its patch ordinal. **Row10 is the first predicted successful case from the exact code, rather than a directly recorded runtime row.** `fixture-predictions.json` distinguishes all13 source-derived predictions from actual observations.

## Exact basis and limits

This read-only diagnosis binds frozen startup full12patch `ecdc9d4ae144a4a8d24a60b4cba758decbffbe5362b3cc9a8afe9e9dc71ea38f`, source identity `7e856365e701e96ddf233e13d6883c6d74464dd755c0ff59f1b185c37c7b19db`, all682 source hashes and pinned spec67772742. The fixture test is5b9de797, and the regex correction remains unchanged. The actual failed command, JSON/log/exit/process samples and closed gate are copied, alongside root's preserved failure packet. No incoming main/spec/proposals or separate source basis is adopted.

The existing root-run red signal is the captured `focused-source.started.json` argv and its failed JSON/log. This diagnosis uses the diagnosing-bugs skill's separation of evidence/prediction, with live reproduction/instrumentation/fix phases excluded by root's explicit no-runtime/no-probe/no-rerun/source-edit instruction. There is no new pass/fail probe, compiler execution or verified-green claim.

## Why row10 resolves under the existing contract

`reference-deployment.test.ts:510–517` supplies `contracts.descriptors=[]` and one local schema at `schemas/bad` with `{"$ref":"https://missing.test/schema"}`. All preceding directory/create/credential/policy cases have explicit rejecting branches.

1. Loader254–302 decodes/admit-bounds the envelope and delegates the immutable RU projection to `captureInstalledReferenceConfiguration`; it adds no independent schema policy. The accepted plan3/contract3 require precisely this delegation.
2. Installed capture135–147 interprets empty descriptors as the shipped complete reference Type closure. `reference-domain.ts:3–73` contains8 Bead Types and5 Link Types. None supplies `propertiesSchema`.
3. `compileDevelopmentContracts:74–105` ingests all artifacts, then calls `compilePrivateSchemaEvaluator` only inside `if (descriptor.propertiesSchema !== undefined)`. **For this fixture it calls the schema evaluator zero times.** Artifact ingestion357–383 parses the schema's embedded JSON, while405–424 checks Type dependencies/conformance, not JSON Schema `$ref` closure. The standalone schema is valid JSON and satisfies byte/count/local hosting constraints.
4. The installed capture's Scope-local schema hosting151–158 accepts this retrieval URI; aggregate[] passes; identities/capture are produced. Loader303–313 registers and returns a legitimate inert handle. `failure()` therefore reaches its own throw216.

This is narrower than saying every unused schema always bypasses validation. Once any descriptor supplies a propertiesSchema, evaluator180–201 builds the graph from the full bundle. `installed-schema-graph.ts:186–191` indexes all submitted schema documents,498–511 resolves their references, and444–446 refuses a missing resource. The failing fixture selects no evaluator entry at all. The recommended replacement deliberately activates that existing path.

## Per-fixture prediction

| Case | Exact fixture | Predicted existing behavior |
|---|---|---|
|1|Relative directory|contract refusal: absolute provisioning path required.|
|2|create is string|contract refusal: boolean required.|
|3|Empty credentials|contract refusal: at least one credential required.|
|4|Duplicate token|contract refusal: token uniqueness, independent of principal duplication.|
|5|Short token|contract refusal: existing token lexical/length rule.|
|6|RU credential adds canMutate|contract refusal: exact token/principal shape; this is not the RU access-policy write flag.|
|7|Only writer policy, reader credential retained|contract refusal: capture policy for every credential principal; reader has none.|
|8|Unknown writer policy field|contract refusal: access policy has exactly create/write/hidden/readOnly.|
|9|Malformed descriptor payload|contract refusal from artifact JSON parsing.|
|10|Empty descriptor list, standalone dangling-reference schema|**Preparation succeeds**, because substituted reference Types have no schema entry to invoke the evaluator; first predicted wrong refusal expectation.|
|11|Foreign Type descriptor|contract refusal from existing RU Scope-local hosted-Type constraint; predicted unreachable after10.|
|12|Aggregate policy names types/absent|contract refusal: actual captured link Type inventory has no such member; predicted unreachable after10.|
|13|Explicit owner owns a missing Link Type, after the loop|contract refusal from existing Type dependency checks; predicted unreachable after10.|

These are static predictions. The actual console/JSON show only the callback failure, so cases1–9 and non-reachability of11–13 are inferred from the sequential code and identified first legal fixture. No test-level row result has been manufactured.

## Policy, aggregate and production distinctions

A complete finite RU policy with `create:false`/`write:false` for reader, legal hidden/readOnly lists and a corresponding principal is accepted existing policy. It should stay legal. `createDevelopmentAccessPolicy`181–243 validates shape/flags/local IDs, and installed capture119–122 checks credential-to-policy presence. Missing reader policy is invalid; a complete read-only policy is not. Duplicate principals with distinct tokens are also already accepted and covered by a passing positive callback. Do not tighten these contracts to make this loop pass.

`maximumEndpointMultiplicity` is captured against the actual compiled descriptor inventory (`development-contracts.ts247–273`). Default remapping supplies13 local Types, with Link suffixes relationship/dependency/blocks/relates/cites, not absent. The existing invalid aggregate fixture therefore remains a proper refusal. The positive explicit `artifacts()` edge policy test304–325 passed in the actual run. Neither adding a fictitious absent Type nor dropping aggregate validation is justified. Pinned spec647–692 describes positive maximum multiplicity and leaves its administrative mechanism outside BDP;802–823 describes authorization projection. Neither creates a new loader validation rule for every uploaded unused schema.

The captured evidence supports a misplaced negative fixture, not a demonstrated production defect in loader delegation. Adding blanket schema validation/preflight for every unused RU artifact, changing default reference Types, or imposing a new finite-policy restriction would alter the established capture contract beyond this repair. TX's existing explicit runtime preflight does not justify copying extra policy into RU. No new BDP normative decision is needed for the recommended fixture correction.

## Recommended bounded correction and later check

Keep the same callback/loop/refusal helper. Replace only row10 contracts by the already valid `artifacts()` descriptors/schema closure, replacing its first schema with local `schemas/item` whose `$id` is that same retrieval URI and whose `$ref` names the absent external resource. The base descriptor already points to `schemas/item`; owner/edge dependencies remain complete. No network fetch is introduced. Expected chain is graph missing-resource → evaluator refusal → existing compiler exception → loader fixed contract error. Preserve error redaction/cause omission, owner/listener tripwires, no-store proof, callback selection44, all other fixtures and budgets.

`proposed-test-only.patch`, its inverse and `proposed-correction.json` bind before/proposed-after hashes. They are recommendations, not an applied successor. Root/native precision acceptance and a separate exact source release precede authoring. Later validation must use a separately named fresh successor with source/selection/report bindings updated and existing stop/acceptance rules; do not retry or overwrite this failed attempt. All25 stages remain governed by root's fresh-pipeline process. The predicted correction has not been formatted, typechecked, linted, built, tested or reviewed here.

The runner exit is1/unaccepted, source unchanged,6.28395s elapsed, no forced cleanup, no reported errors/remaining sampled processes/runtime directories. Stage10 failed; the15 later stages have no started records. These are copied runner observations, not a new process census or cleanup action. Gate stays closed. Public profile startup/admission, intended-listener lifecycle and installed loader proof remain unqualified by this failed source prefix.
