# CLOSED Python format-dependency continuation plan

Preparation only; root approval is required before provisioning or execution. Candidate d7197e97a29bff8f075028419afad05388e2e655/tree630a94d0ca1545f898cb8a8b614d4b7d2dd6d454 remains unchanged. No product/Types edits, test changes, new source review or full-suite rerun is proposed.

The preserved type-examples attempt returned1 at `Issue due_at` with no validation errors for `bad-date`. Installed jsonschema4.26.0 registers date-time only when its optional rfc3339_validator import succeeds; METADATA declares rfc3339-validator under the format extras. Root observed no registered date-time checker and no provider package. Read-only inspection confirms that mechanism. The example checker and Issue schema are byte-identical to frozen main8d6671. This is an isolated environment dependency gap, not a product/schema defect. Root evidence is [type-example-dependency-diagnosis.json](type-example-dependency-diagnosis.json); the original failed log/exit and old environment must remain immutable.

## Environment correction

1. Provision a new isolated `tools/python-format-venv` using the same pinned base Python. Keep existing versions: attrs26.1.0, jsonschema4.26.0, jsonschema-specifications2025.9.1, referencing0.37.0, rpds-py2026.9.1 and pip26.1.1. Add only rfc3339-validator and its declared six dependency if needed. Pin actual resolved added versions/archive identities before runtime; record resulting package metadata, interpreter and environment file hashes. Do not install the entire format extra or modify the old venv. Reject unintended upgrades.
2. Run a bounded diagnostic in that new environment: require date-time registration and verify both FormatChecker and Draft202012Validator with a string/date-time schema accept `2026-10-17T12:00:00Z` and reject `bad-date`. The negative check detects silently missing enforcement. Record actual provider identity and outcomes. This is a tool prerequisite, not a pass for the example script.
3. Run the unchanged `scripts/check_type_examples.py` once using the new interpreter. Preserve its actual output and shape-only meaning. Keep its120s/TERM115 bound. Provisioning and probe are separately bounded120s/TERM115, no automatic retries or silent widening. Any failure stops for root disposition.

## Preserve and reuse the successful prefix explicitly

Prepare a small CLOSED sibling continuation after approval; do not change `validation-private-tmp-successor` or its run. Retain the original25-stage logical sequence and all existing source/build/process/provenance gates, deadlines, required CLI matrix, raw0 evidence rule, packed selectors and callback entry distinctions.

Reuse exactly11 accepted stages from that predecessor: dependencies, format-check, lint, check-boundaries, typecheck, build, bind-build, full-test, e2e-ready-preflight, smoke and evidence-verify. The9375-test full-suite result remains attributed to its actual prior run. Bind root's actual acceptance record. Do not copy or synthesize fresh stage results, and never reuse the failed type-examples attempt.

The only gate adaptation needed is an explicit11-name prefix-origin map: `complete()` reads the original pinned started/exit/log/process records for those names and existing source/cleanup acceptance checks remain in force. All later names resolve only to the sibling's fresh run directory. `verify_build()` reads the original pinned build-artifacts record and checks the current emitted inventory through the existing gate. No general fallback or new review framework. Verify unchanged candidate, lock, Node/pnpm/bd and build inputs before reuse; drift stops continuation.

The new tool binding preserves original prefix tool attribution while binding the new Python environment for future stages. Only type-examples' interpreter changes semantically; remaining command edits are sibling helper/output paths. Continue once from type-examples through the remaining13 stages, with fresh packages, consumer lock/provenance and five separate installed callbacks. Root lock acceptance and immediate callback audits remain required. Runtime temp remains task-owned `/private/tmp`.

New gate starts rootExecutionReleased=false, approvedStages empty and consumer pins null. Root/native review the minimal helper delta, actual tool probe and explicit origin bindings before release. Final reporting must distinguish11 reused stages and14 new attempts in the25-stage chain, with provisioning/probe separate. Neither failed attempt becomes a pass retrospectively. Public capability claims, source and main remain unchanged.

## Inspection performed

This plan involved file/metadata/Git-object reads only, no package imports or runtime/provisioning. All701 candidate source hashes remain unchanged. Each of the11 prefix exit records was checked accepted with unchanged source, no forced cleanup/survivors and matching log hash. Detailed runtime acceptance remains root-owned.

Evidence pins:

- `type-example-dependency-diagnosis.json`: `04a93521ed82705d539a5b0fdbbcba759045a32346845363f38809bade6a1910`
- `validation-private-tmp-successor/run/type-examples.exit.json`: `3d3654bf7eabe241b9eb27026d276cb9fc71137f30256af0c9ba3d133b49fa52`
- `validation-private-tmp-successor/run/type-examples.log`: `fd09239d4797357ab8c2c651c18139340bbe9085ce6d66c07301ede33037e7f9`
- `validation-private-tmp-successor/run/full-test.exit.json`: `894289211d4f34596ccb0ef2d0e2de369928614ceff6f580d0bf3d8fa1aa1869`
- `validation-private-tmp-successor/run/build-artifacts.json`: `8c53e912f90406831be452eb79fcde95bb5ecdb2641298c6a1e43f98d3708c54`
