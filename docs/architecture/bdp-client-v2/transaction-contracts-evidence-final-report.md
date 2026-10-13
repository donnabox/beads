# Compiled-contract evidence follow-up

Frozen head: `fc75dc5e83e8b87db22f28e7137f57ee6204f3df`
Parent: `7b6f8bcbe7a99776e674b05686080d607c98c775`
Branch: `codex/vickie-contracts-evidence-final`
Checkout: `/Users/dbox/repos/wt/vickie/vickie-contracts-evidence-final`
Origin: `https://github.com/gastownhall/bdp.git`
Working tree clean. Local commit only; no outward writes.

## Changes and dispositions

Three repository files changed, with no runtime changes:

- `apps/bdptest/src/development-contracts.test.ts`: mutate both schema and descriptor input buffers before verifying captured bytes; clarify that the pinned historical identity binds artifacts and all three ceiling sets. Future intentional input changes require a justified new witness.
- `packages/server/src/transaction-installed-contracts.test.ts`: the corresponding TX descriptor-buffer mutation; exact complete failed-batch and failed-set receipt replay equality; exact full Type GET equality before and after restart; exact inventory summaries and full GET equality for both listed IDs.
- `docs/design/transaction-implementation-progress.md`: describe precisely those assertions and their bounds.

The external review suggestion to place full descriptors in the Type inventory was rejected. Current spec `docs/specs/bdp.md:1092–1100` requires exactly `id`, `name`, and `describes`; `$defs/typeSummary` closes extra properties, and the actual reader deliberately projects that summary. The attempted full-descriptor assertion failed against correct runtime output (1 failed, 26 passed); `inventory-schema-counterexample.log/.json/.exit` preserve this counterexample. The approved repair instead checks the exact summary array derived from independently captured installed descriptor inputs, follows both IDs, and compares the full fetched documents. This is a review correction, not a product or protocol change.

## Historical identity reproduction

The external `../transaction-contracts-evidence/prior-identity.mjs` now computes the source-diff result with `git diff --exit-code fccb951 3e2b878 -- <exact 34 relevant files>` rather than reporting a hardcoded boolean. The set includes compiler and installed-schema source, protocol production source, schema documents, and package/build pins. It excludes unrelated runtime and test changes. `predecessor-source.diff` now records the exact command and actual exit zero. Historical original script/log copies were preserved.

The probe verifies the clean frozen predecessor checkout at `fccb951ce0a470526a22db9762236935f78ec5e8`, verifies ancestry, imports its existing emitted compiler (not rebuilt or modified), records compiler SHA-256 `abb874ef02bad361482d9a72543500886a6e8fb47f0d6a8b85f9e805103ba788`, and reproduces `development-schema5-894c98da1156e66e72d7f5b4321365b9496ec68afea377dc1861bdaaa2cadf55`. This is bounded historical reproduction, not an artifact attestation. Corrected probe exit: 0. `broad-directory-control.log` records why the review's proposed broad directory comparison would include unrelated changes.

## Verification

Pinned Node 24.16.0 and pnpm 11.20.0. External logs are in this report's directory.

- Frozen install, typecheck, build, boundaries, formatting and lint: exit 0.
- Build preceded emitted HTTP tests.
- Focused Vitest: **27 passed, 2 files, zero failed/skipped**, 1.86 seconds. Files: `apps/bdptest/src/development-contracts.test.ts` and `packages/server/src/transaction-installed-contracts.test.ts`. This includes actual emitted HTTP serving and durable restart assertions.
- Final typecheck, format and lint: each exit 0.
- `git diff --check`: passed; committed working tree clean.
- An initial duplicate test-local constant caused TypeScript errors; it was corrected before build, and the failed log is retained as `typecheck-initial.log/.exit`.

No full-suite, separate transaction-runtime suite, packed smoke, or evidence verification was repeated for this tests/docs-only follow-up. No production profile qualification, evidence reseal, public advertisement, store-format change, or materialization-budget change is claimed. The owned-Link author’s separate checkout and source were untouched.
