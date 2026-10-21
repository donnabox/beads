# Profile/write integration diagnosis

Candidate: `29eb76fd8d6a841e69df54c0a88bd22477ae9465`, branch `codex/vickie-bdp-profile-write-candidate`, clean local checkout. Runtime base: `21ba42bf28b483a5ee22c77b6b28b67d464f3045`. Ordinary editorial merge: `2432e84434130feeddfe2b22957bcdc3977dae21`, integrating reviewed `e9361d77b862aa0e6ddfa19cf272225f62c32d35`. Nothing pushed.

## Bounded integration

Two files changed after the merge: `packages/conformance/development/read-update-catalog-v1.json` (22 of 96 citations migrated), and `packages/conformance/src/profile-catalog-boundaries.test.ts` (explicit diagnostic catalog included in citation/profile audit). No production artifact admission change. Nine diagnostic scenarios, 57 HTTP actions, 36 bounded/45 omitted of 81 mapping preserved. Manifest, fixture, coverage inventory, schema copies and all non-requirement catalog fields unchanged from runtime base. Citation mapping is in `profile-write-citation-mappings.json` beside this report.

## Verification

Pinned Node24.16.0/pnpm11.20.0. Frozen install, build, typecheck, dependency boundaries, whole-repo formatting, installed executable smoke, and non-attesting development smoke all passed. Full Vitest with maxWorkers=2: 114 files (113 passed/1 failed), 7678 tests (7674 passed/3 failed/1 skipped), 62.11 seconds. Logs and JSON are `profile-write-*` in this directory. `git diff --check` passed. No source changes after that suite except committing the tested bytes.

## Three exact failures

All in `scripts/read-cohort-evidence.test.mjs`, describe `Read schema projection drift over the committed cohort`:

1. Line348: verifies the sealed cohort against the committed bundle's Read projection.
2. Line390: tolerates top-level bundle metadata drift: description, title, $id, and $schema are outside the digest.
3. Line415: tolerates a definition the seal does not name, added by a later profile.

Each assertion expects `verifyReadCohortEvidence(committedInput(bundle))` not to throw. Each throws exactly `ReadCohortVerificationError: current catalog bytes drift for 'bdptest': re-seal required`. The positive projection cases share current catalog inputs (lines324–332), so all hit the same independent byte-binding gate.

`pnpm evidence:verify` independently exits1 with that same error; see `profile-write-evidence-verify.log`. This is not only a test-fixture issue. The production verifier reads the current catalog in `scripts/read-cohort-evidence.mjs:359`, and `packages/conformance/src/cohort-verification.ts:383–400` compares exact current catalog/manifest/fixture bytes to every sealed segment. `packages/conformance/matrices/README.md:730` and `docs/design/read-projection-gate.md:26` explicitly require current byte binding and new observations even for assertion-only changes. CI runs this verifier at `.github/workflows/ci.yml:77`. Current qualification therefore requires genuine requalification/reseal under the existing rule. Changing tests to read an old catalog would not resolve the production gate.

## Pins and hashes

Cohort run head: `de99030d13a57e77f4f2660b91ea3c39b82823cb`. Latest evidence commit: `a7edb0c2c5c8e799a82f05cfbbb00b8aa0795d3c`. Current sealed artifact SHA256: `7ef1a79470ebf638539412d9c79756ffffbc1a9a128a33a880f27a57b538413c`; recorded constant `7ef1a79470ebf638539412d9c79756ffffbc1a9a`. Artifact and constant are byte-identical to runtime base and editorial input; no relabeling performed.

All four cohort segments (two targets, in-process and packaged) bind the old catalog hash below. Runtime base also has those old catalog bytes. Editorial input and candidate have the current catalog hash. Manifest, fixtures and schema remain identical across run head, runtime base, editorial input and candidate.

| File | Cohort run-head SHA256 | Candidate SHA256 |
|---|---|---|
| `packages/conformance/catalog/read-v1.json` | `f338675885de28e01de3134a210920b8f10c02655c8a93837bedca281baa782a` | `73e3d50901c570bd066e31efc2484f47e51e5a886bfd9c1af0c5cc370a231aae` |
| `packages/conformance/matrices/read-v1.json` | `ca175737db61eba6e013f429248088ed05ccae31e26f1eb4e30539bd2dfd62a1` | `ca175737db61eba6e013f429248088ed05ccae31e26f1eb4e30539bd2dfd62a1` |
| `packages/conformance/fixtures/read-reference-v1.json` | `2eb36cb0b1336cd86bf350983d20e6c6eca94fcb9b66ec3e9c8d3d2a19db3247` | `2eb36cb0b1336cd86bf350983d20e6c6eca94fcb9b66ec3e9c8d3d2a19db3247` |
| `packages/conformance/fixtures/read-bdpbd-v1.json` | `03ecf4809a6056171ac0dbb00da4ce6ab31fbb7bec7e32253092169dd675f97f` | `03ecf4809a6056171ac0dbb00da4ce6ab31fbb7bec7e32253092169dd675f97f` |
| `schemas/bdp-v0.schema.json` | `bc9d5f65d6907bdf0702ebb0587d521dc443ba643aa56fb235414e97943dd994` | `bc9d5f65d6907bdf0702ebb0587d521dc443ba643aa56fb235414e97943dd994` |

The unchanged sealed Read schema projection is `2db57a82d9bf99f3fb5d18646e8bd6c7646590d1e806aa266de8af415ee83545`; no schema drift caused these failures. Historical evidence remains historical evidence for its recorded inputs. This candidate is not all-green and has no current qualification claim. No seal, constants, evidence checks, runtime semantics, or schema were changed in this integration slice. The coordinator owns any further qualification work or decision.
