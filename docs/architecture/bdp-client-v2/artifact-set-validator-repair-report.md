# Private artifact-set validator repair

Frozen head `175b98b40d7f683a7f171e821407f8c561a2ec68`, branch `codex/vickie-artifact-set-validator-repair`, checkout `/Users/dbox/repos/wt/vickie/vickie-artifact-set-validator-repair`. Direct ordinary successor of `3ae9849854e731049782d3962ebf0b2421b7ceb0`; original PR91 tree untouched. Origin is GitHub. Clean source, no push/bus/launch update. Only private helper and its test changed; no app migration, C restoration, runtime, runner, schema, catalog, public boundary or coverage decisions changed.

## Review dispositions

- **F1:** Reject owned scenario setup or applicability requirements absent from the bound shared fixture before assigning ownership. Negative cases cover both gates/both shards and the inapplicable missing-fixturePointer bypass. Actual shipped 47-scenario corpus is accepted with exact coverage and unchanged prerequisite group: test splits its independent final row into the second artifact; this is validator construction only, not execution evidence.
- **F2:** Safe source capture assigns fixed index/fixture/coverage labels or validated shard ID/source. Errors from existing hardened copy retain the cause and bounded error text under that role. JSON parse labels distinguish either manifest. No caller byteLength/getter precheck; continuation hostile byte properties/iterator remain uninvoked. Arbitrary caller labels do not determine attribution. Existing shape validation and every-source-before-factory capture order remain intact.
- **F3:** Dotted IDs stay legal. Duplicate joined action/assertion identifiers within a row are refused. Two legal alternative splits with equal serialized coverage remain accepted but have different exact manifest and set digests. Code comments and tests state that manifest plus coverage jointly bind structure; coverage alone is not a structural identity. No concrete same-artifact ambiguity remains in this joined list because duplicates are rejected.
- **F4:** An owned whole=true value is accepted as inert authored data, changes coverage/set digests and never changes claimEligible=false or ownership/bundle facts. The independent reviewed-ID gate is untouched. This branch still carries historical 28 whole; later 9dd's accepted 30 whole belongs to ordinary integration, not a promotion here.
- **F5:** Whitespace-only reasons refused. Continuation now participates in exact-byte identity and byte/node ceiling tests, plus detached capture with hostile byte properties. Dynamic imports inspect actual named exports of both public index and testing module; ownData/set constructors remain absent. Retained explicit nonempty-binding check is cheap defensive intent. No error-class refactor or broader framework.

## Verification

Pinned Node24.16.0/pnpm11.20.0, maxWorkers2. External logs record command exits and timings.

- Frozen install passed.
- Red run before helper repair: **9 failed / 62 passed** across 71 tests. Failures reproduced missing capability, source attribution, duplicate joined binding and blank-reason acceptance. New dotted-split/whole-inert/corpus/boundary checks already passed on predecessor, as intended. `red.log/json/exit` preserved.
- Initial green focused run: **71/71**, 0.586s Vitest duration.
- Final regression after adding continuation node-ceiling parity: **417/417 in 6 files**, 1.09s Vitest duration. Includes 71 private set tests plus artifact-bundle, cumulative-artifact, profile-contract-plan, runner and transactional-cumulative. Full commands/results in `checks.py`, `regressions.log/json` and exit JSON files.
- Typecheck, clean build, dependency boundaries, changed-file format, lint and git diff checks passed. Lint reports 26 warnings / zero errors, including existing test-builder style; no suppression or hook bypass.
- Historical Read verifier passed unchanged: 89 rows, 47 definitions / 46 reachable, existing constant and artifact. No write-profile evidence reseal.
- No timed-out/interrupted run, installed authority or owned service child was created. Final process snapshot shows no matching owned Node/Vitest process; source clean after commit.

## Exact measurements and limits

`source-measurements.json` binds byte/node counts and hashes independently to both commits. JSON node convention counts each scalar/container once; counts are inputs, not allocation/heap bounds.

| Source | Manifest bytes / nodes | Fixture bytes / nodes | Coverage bytes / nodes | Authored / whole / required |
|---|---:|---:|---:|---:|
| Frozen3ae (this repair's base) | 3,328,407 / 89,024 | 15,888 / 423 | 528,050 / 10,996 | 47 / 28 / 129 |
| Later accepted9dd (read-only measurement) | 3,626,922 / 97,116 | 26,370 / 687 | 573,722 / 11,931 | 49 / 30 / 129 |

The permanent actual-corpus construction test uses this branch's 3ae corpus. Reading/counting9dd is not running this repaired validator against9dd and not installed compatibility proof. No C draft, merged fixture, report association, per-shard route/configuration binding, abort/drain, fresh packed target, aggregate memory/time or full workspace suite was executed. Those remain the approved later app slice after ordinary integration. All current parser bounds, genuine factories and public admission remain unchanged.
