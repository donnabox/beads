# Native independent review: profile contract groundwork

Reviewed frozen `abdd1d4af45cd02ad1238b311462a1b154e44b79` against `caf5ca75a4e9b9c7865e61e97f849ab736c21e64`. Author checkout `/Users/dbox/repos/wt/vickie/vickie-profile-contract-selection` was read-only. Tests used detached isolated clone `/Users/dbox/repos/wt/vickie/bdp-profile-contract-native-review`; both trees clean at completion.

## Blocking findings

### P2 — repeated input reads defeat closed, validated configuration snapshot

`packages/conformance/test-support/profile-contract-plan.ts:140–175`, particularly final spread at175. Validation reads live fields/keys, then `{...input}` reads them again. A sourceCommit getter returning40 hex characters on the validation read and `invalid` on the second read produces a frozen accepted binding with `sourceCommit: "invalid"`. A proxy returning the eleven allowed keys to initial Object.keys and adding an enumerable `token` key for spread produces a binding containing that extra field. Thus neither validated primitives nor closed non-secret membership are guaranteed. The same pattern affects capability array reads and may freeze caller-owned values instead of a detached snapshot.

Fix: reject proxies/accessors and malformed non-plain shapes at the boundary as directed by root; capture allowed primitive data once into a detached explicit closed record, copy capabilities without live custom iteration/aliasing, then validate/freeze that captured record. Do not reread live input after validation. Add changing-accessor, ownKeys proxy, input alias mutation and capability-array controls.

### P2 — digest regex coercion can hide different configuration identities

Same file158–166 and equality181–184. Neither sourceCommit nor SHA256 fields check typeof string before regex.test. Two object-valued policyDigest fields with inherited toString methods yielding distinct valid64-character hashes both pass. canonicalJson subsequently serializes each as `{}`, so assertSameServedProfileConfiguration accepts their combination. This directly defeats the exact installed-policy/configuration separation claimed by the planner, even without a getter/proxy.

Fix: require primitive strings for every commit/digest before regex/coercion, and ensure only validated detached primitives reach canonicalJson. Add boxed/coercible object and unequal-digest regression tests.

Both defects are in new source-private planning code; no current public conformance admission bypass is claimed. They nevertheless block using this configuration gate for the proposed same-target served runner.

## Reproduction

Source `/Users/dbox/repos/wt/vickie/bdp-integration-logs/profile-contract-native-repro.test.ts`; log `profile-contract-native-repro.log` and exit file alongside. The file was temporarily placed at packages/conformance/src/profile-contract-native-repro.test.ts in the isolated clone, run with pinned Vitest maxWorkers1, then moved externally. To repeat, copy it to that location in an isolated checkout of abdd1d4 and run that single test file.

Three repro assertions intentionally assert the observed faulty acceptance; all3 pass: ACCESSOR_ACCEPTED_INVALID invalid2, PROXY_EXTRA_FIELD includes token, COERCED_DIFFERENT_DIGESTS_ACCEPTED distinct all1/all2 hashes. These are evidence of defects, not desired acceptance tests. Author should turn them into rejection controls.

## Catalog and inheritance review

- Independent parse/dedup/retirement arithmetic gives44 normative Read,129 cumulative RU,250 TX. Read has49 total rows with5 diagnostics; companion duplicate copies are exact. No diagnostic coverage inflation found.
- New RU discovery retirement correctly excludes closed Read-only discovery at RU and carries its exclusion into TX even when TX retires RU discovery. Twelve TX→RU retirements plus the inherited RU→Read edge produce13 excluded IDs, not13 direct TX retirements.
- The audit table lists each of81 RU IDs exactly once,12 retired and69 inherited, with no ID omission/duplication. Its preserved sequence200/member-envelope/separate-commit/no-isolation interpretation is supported by existing Mutation Transactions text (4048 onward), dependent/pending handling4098 onward and status table4598. No additional semantic rule is required. Existing explicit receipt/error overrides provide the envelope context for inherited operation laws. This bounded review did not independently implement or qualify all81 behaviors.
- Removing the no-receipts/no-Events descriptive citation from owned-Link no-op avoids importing a lower-profile-only fact; revision/source/provenance requirements remain. TX no-op Event obligations remain separate.
- Missing executable rows remain unplanned; missing setup/applicability remain separate and unobserved rather than automatic N/A. Every result claimEligible:false. Optional History remains opt-in and diagnostics cannot become required rows. Factory selections are frozen/branded. The injected readSource boundary is explicitly trusted rather than attested, consistent with source-private scope.
- Existing Read manifest, sealed evidence and development diagnostic bundle are unmodified. The focused artifact/catalog/runner regression tests pass. No other blocker identified in this bounded pass.

## Validation

Pinned Node24.16.0/pnpm11.20.0, frozen install and build passed in isolated clone. Focused serial suite:375 passed in11 files,9.19s. Exact list matches author report: profile-contract-plan/catalog/read-update-catalog/transactional-catalog/artifact-bundle/numeric-model-catalog/owned-wildcard-catalog/history-catalog/profile-catalog-boundaries/development-read-update/runner.

Structured result: `profile-contract-native-focused.json`; verbose log/exit alongside. No broad duplicate suite, actual served qualification, evidence generation or source/outward mutation. Root owns routing these findings and independent re-review after repair.


## Independent closure at 9ea4089

Reviewed exact successor `9ea408958e08be38cc4e586a4846caf3d6ef6874`, direct child of preserved `abdd1d4af45cd02ad1238b311462a1b154e44b79`. Reviewed three-file repair diff and author report `profile-contract-selection-repair-report.md`. Frozen author checkout untouched; isolated review clone updated to detached successor.

**Both P2 findings CLOSED.**

- Configuration records reject proxies (including revoked proxies) before reflection, reject custom prototypes, and inspect own descriptors once. Only exact enumerable own data members are captured. Accessors are rejected without reading them; output uses explicit captured primitives rather than spreading caller input.
- Commit/digest/Scope/fixture values must be primitive strings before regex or URL construction. No coercible object reaches equality. Profile/History checks remain strict scalar checks.
- Capability arrays reject proxy/subclass/accessor/sparse/custom-iterator/extra-member forms, capture only own data values, require primitive nonempty strings, and sort/freeze a newly allocated array. Neither caller record nor original array is frozen or aliased in the result. Null-prototype plain records remain accepted. No prototype or caller hook is used in the accepted value traversal.
- Independent versions of the exact three original reproducers now require rejection and verify zero getter/ownKeys/toString calls. All three pass. Persisted source: `profile-contract-native-closure.test.ts` in this log directory; run it from packages/conformance/src in an isolated clone, as with the original reproducer. The original faulty-acceptance repro remains unchanged separately.

Independent validation: build PASS; focused serial suite397 tests/12 files PASS in4.31s (394 maintained tests across original11 files plus3 independent closure tests); typecheck PASS. Structured log `profile-contract-native-repair-focused.json`, verbose/exit companions; build/type logs share `profile-contract-native-repair-*`. No broad suite, new evidence generation or public qualification claimed. Isolated scratch test removed to external logs; author and review checkouts clean.

Verdict: native closure approved for the two reported binding defects. Original catalog/inheritance assessment and source-private, non-attesting limits stand. External Claude supplement is independent and remains root-owned.
