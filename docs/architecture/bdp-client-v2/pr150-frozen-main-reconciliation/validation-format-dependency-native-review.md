# Native diagnosis and dependency-continuation plan review

No actionable plan blocker found. The proposed isolated optional-dependency repair and explicit successful-prefix reuse are proportionate. This approves the bounded preparation approach, not an unimplemented gate delta, package resolution, probe result or execution release.

Reviewed frozen plan `validation-format-dependency-plan.md` SHA256 `4f0968449d722702f57f9062ec2ff7f9e6962c2cf32884259dfa651717693958`, root diagnosis `04a93521ed82705d539a5b0fdbbcba759045a32346845363f38809bade6a1910`, failed terminal record/log and installed Python source/metadata. No Python package imports, CLI/test/helper executions or environment/source mutations were performed.

## Narrow cause verified from source

`scripts/check_type_examples.py:38` constructs Draft202012Validator with FormatChecker; its due_at negative case at205 expects `bad-date` to fail. The Issue schema declares due_at as string-or-null with format date-time. Installed jsonschema4.26.0 `_format.py:395–402` puts the rfc3339_validator import and date-time registration inside suppress(ImportError). Its FormatChecker.check returns without error for an unregistered format (around131). Installed METADATA lists rfc3339-validator only for format/format-nongpl extras; the isolated site-packages contains neither provider module nor distribution. These facts independently explain the observed empty error list. Root's recorded live probe additionally reports no registered checker/importable provider; this review did not rerun it.

The current script, Issue schema and domain-case fixture are each byte-identical to exact incoming8d6671ed19b0c0406428ccc2d2161768d01e8b9b. Their SHA256 values are respectively c096eec47f4fa8e9d7248a6e57829b5c4e14f5eb5fa5b626673bf28e1f21e06e,7e7628dcd6303127ac216780b86ed0c9c9a1a04dac968808662e52c41f4367a7,e829b81585c6c9d1bb9fb65604a9dc54c2de45c820cfb8a5c388b935fb11c305. No product or Types/schema repair follows from this evidence.

Actual type-examples raw1/accepted=false,0.710s, unchanged source/no forced cleanup/no reported survivors remains failed. Its log SHA256fd09239d4797357ab8c2c651c18139340bbe9085ce6d66c07301ede33037e7f9 and exit3d3654bf7eabe241b9eb27026d276cb9fc71137f30256af0c9ba3d133b49fa52 match. An optional-dependency fix is not yet a passing example check.

## Environment and probe scope

The plan preserves the old environment and every prior tool binding. Its new isolated environment retains the six actual metadata versions: attrs26.1.0,jsonschema4.26.0,jsonschema-specifications2025.9.1,referencing0.37.0,rpds-py2026.9.1,pip26.1.1. Adding only the date-time provider and its needed six dependency avoids unrelated format extras/upgrades. Actual added versions/archive identities and complete resulting tool/environment hashes remain required before runtime release; no version or success is presumed here.

The positive/negative probe is discriminating: require registration, then verify both FormatChecker and Draft202012Validator accept the valid timestamp and reject bad-date. It is separately recorded as a prerequisite, followed by the unchanged complete example script once. Provisioning/probe/script retain the planned120s/TERM115 limits and no automatic retry. No schema weakening or substitution for the example script is proposed.

## Explicit reuse and continuation

Independently checked all11 predecessor prefix exits as accepted raw0 with unchanged source, no forced cleanup/survivors and matching logs. None of their commands uses the old optional jsonschema environment: ten use Node/pnpm, and bind-build uses the pinned base Python/standard-library helper. This supports explicit reuse after unchanged source/lock/Node/pnpm/bd/build/tool verification, without repeating the full suite solely for the isolated formatter dependency.

The plan correctly keeps those11 records at their actual origins, including original started/exit/log/process hashes and root acceptance; no copied outcomes, generalized fallback, or failed type-example reuse. Original build-artifact evidence remains bound and rechecked against current output. There are14 fresh logical stages from type-examples through packed-affected-5 (the wording 'remaining13' means after the fresh type-example stage), plus separately bounded environment provisioning/probe. This retains the25-stage logical chain.

Actual implementation must still be inspected before release: exact11-name origin map, old build-record origin, immutable original tool attribution and fresh future Python binding, all subsequent names confined to new run evidence, closed root/consumer slots, unchanged packed selector/entry distinctions, raw0 evidence rule and actual lock/provenance/callback acceptance. The plan explicitly requires those controls; this is not a request for a wider framework or extra test surface.

All five evidence pins listed in the plan matched during review. Both historical failures remain immutable and separately identified. No product/public capability claim, main/source change, or automatic execution is authorized by this report.
