# Combined profile/write successor verification

Frozen head: `4f0ed57906525c36bc55b343b77e0f07edbb7553`.
Branch: `codex/vickie-bdp-profile-write-successor`.
Checkout: `/Users/dbox/repos/wt/vickie/bdp-profile-write-successor`.
Worker froze a clean local candidate. Root subsequently pushed this exact head; no further implementation edits after freeze.

## Sources and integration

All reviewed inputs are ancestors by ordinary merges:

- Existing candidate `29eb76fd8d6a841e69df54c0a88bd22477ae9465`.
- Runtime integration `0ccdc9bd8acd808cbddd103119c679a3f2dd42ae`.
- Process qualification `5f07a51c0bc0e32652b02ab16cdb2fd4c7f76bc3`.
- Present-contract editorial `5d7694e8be2f5f4ca72856b288dac0ecd383cd2f`.
- Transactional sequence `4a9137e01817712aedce0b780abb39d02cf1470c`.

No merge conflicts or shared-store repairs were needed. The sole integration edit is `packages/conformance/development/read-update-catalog-v1.json`: five mutation problem citations move from Read to Read+Update; one excerpt drops the reviewed editorial amendment tag. Six changed lines; exact mappings are in `successor-citation-mappings.json` beside this report.

All diagnostic IDs, metadata, requirement counts and asserted behaviors are unchanged. The inventory remains 17 scenarios, 132 HTTP/lifecycle actions, 44 bounded and 37 omitted rows out of 81. Diagnostic observations remain non-attesting; production admission is unchanged.

## Verification

Pinned Node 24.16.0/pnpm 11.20.0. Full logs and JSON are `successor-*` in this directory.

| Check | Result |
| --- | --- |
| Frozen install | PASS |
| Build | PASS |
| Typecheck | PASS |
| Dependency boundaries | PASS |
| Whole-repository formatting | PASS, including final bytes |
| Full Vitest, maxWorkers=2 | 115 files: 114 passed, 1 failed; 7,727 tests: 7,723 passed, 3 failed, 1 skipped; 73.95 seconds |
| Installed executable smoke | PASS |
| Development smoke | PASS, non-attesting |
| Final citation/profile/development matrix checks | 3 files, 21 tests PASS |
| git diff --check | PASS |
| Production Read evidence verifier | Expected FAIL, exit 1, current catalog drift |

The final 21-test check followed restoration of original JSON Unicode escape spelling; parsed catalog data was unchanged. All six meaningful repairs were present during the full run. No runtime or schema edit followed that run.

## Known failures and unchanged evidence

Exactly these three tests failed:

- `Read schema projection drift over the committed cohort verifies the sealed cohort against the committed bundle's Read projection`
- `Read schema projection drift over the committed cohort tolerates top-level bundle metadata drift: description, title, $id, and $schema are outside the digest`
- `Read schema projection drift over the committed cohort tolerates a definition the seal does not name, added by a later profile`

Each full log error is exactly `ReadCohortVerificationError: current catalog bytes drift for 'bdptest': re-seal required`. Independent `pnpm evidence:verify` on the clean committed head returns the same error and exit 1. There were no other failures.

- Sealed catalog SHA256: `f338675885de28e01de3134a210920b8f10c02655c8a93837bedca281baa782a`.
- Current Read catalog SHA256: `e89ed2ca0c674352be76d67535a116b2a208808738b51935af721151c39366f1`.
- Both schema copies remain byte-identical to the original candidate, SHA256 `bc9d5f65d6907bdf0702ebb0587d521dc443ba643aa56fb235414e97943dd994`.
- Sealed artifact remains unchanged: `7ef1a79470ebf638539412d9c79756ffffbc1a9a128a33a880f27a57b538413c`.
- Evidence-constant source remains unchanged: `3978f873ce349db99e54b275fddf55acf092899d36c2c555f31a11148511945d`.

No evidence generation, relabeling, constant modification, schema/runtime semantics repair, or Janet asset mutation occurred. This candidate is frozen for independent source review, not all-green or newly qualified. Fresh Read qualification is a separate next task after this combined head is reviewed. Root subsequently reported the pinned baseline independently verified all 27 observations; this integration worker did not rerun that baseline.
