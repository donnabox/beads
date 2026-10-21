# Read+Update admission coverage promotion

Source: `d9a22c312b5cc0a46a68852f2be6511fb1b39e14`, ordinary direct child of frozen `6f3e3bd02a09844422c5a00651cb14c90290d55b`, branch `codex/vickie-read-update-admission-promoted`. Isolated checkout: `/Users/dbox/repos/wt/vickie/vickie-read-update-admission-promoted`.

## Decision and preservation

Only `read-update.singleton.idempotency-key-required` and `read-update.singleton.unsupported-media-type` gain `wholeContract: true`. Their reasons bound credit to the reviewed selected clauses and describe the accepted class/carrier equivalence partition and all nine bare media comparators. The exact reviewed list in `packages/conformance/src/cumulative-artifact.test.ts` gains these IDs in catalog order. All three uncovered-count assertions in `apps/bdptest/src/read-update-matrix.test.ts` change from 103 to 101.

This gives 28 whole rows / 129 required, 47 authored rows, and 101 uncovered required rows. These are diagnostic coverage counts, not public profile qualification. `claimEligible` remains false. All 47 scenario bytes are unchanged from 6f3; runtime, routes, fixtures, all other coverage rows, bindings, and historical evidence are unchanged. The entire source diff is three files, 9 additions and 7 deletions. `audit.py` and `source-ledger.json` record the preservation check.

The accepted native review is `../read-update-admission-comparators-native-review.md`. It includes the actual emitted-runtime differential gate probe: a Delete Alias handler incorrectly requiring a Content-Type parameter passed the predecessor media scenario but failed the new bare comparator at `media-route-delete-alias-corrected-bare` with 415 and the exact status/schema/outcome/alias failed assertions. The counterprobe was not rerun or relabeled as new observations here. The existing nine comparators and unchanged matrix execute freshly below.

## Validation

Pinned Node 24.16.0 / pnpm 11.20.0, maxWorkers=2. Frozen install and build completed. All 123 targeted tests in the cumulative binding and RU matrix files passed (82.29 seconds). Typecheck, boundaries, format (including explicit 4 MiB manifest check), lint, and git diff whitespace checks all exited zero. Lint retains existing warnings; no warning-free claim. No full repository suite was requested or run for this three-file coverage-only change.

## Fresh installed observation

`pack-executor.py` built new local package archives from the exact clean committed source, installed them into a new disposable package directory, and ran the four installed RU test files. All **152 tests / four files passed**, exit zero, 82.57 seconds. The fresh 47-session aggregate completed in 14.384 seconds; the focused two-row observation completed in 1.437 seconds. Full raw diagnostic reports: `packed/progress.json` and `packed/admission.json`; compact result: `packed-summary.json`.

The key row passed 196 raw HTTP exchanges + 88 client actions with 2,303 assertions. The media row passed 149 raw HTTP exchanges + 34 client actions with 1,341 assertions. Raw and client categories remain separate. Complete aggregate observations, setup, native Read, source/configuration binding, and exact archive digests are recorded in the accompanying JSON. These are fresh diagnostic observations, not a reseal or public conformance qualification.

The source was clean at exact commit before and after the run. No timeout occurred and no test-owned installed child remains (`cleanup.json`, process snapshot). No source modification followed validation.

## Exact bindings


- source: `d9a22c312b5cc0a46a68852f2be6511fb1b39e14`
- installed payload: `819c09e3ade4bbc533c118a956af23e22d44d5b7f87a73aef3d391a98dc11093`
- installed contracts: `48667497d6e13064316d0364cf8f49dfc247ab9189306985a357d7b558333dfe`
- fixture: `64814924334103e9375c043de472f8bf15c8b71a267a176e06cd0443b8249751`
- policy: `dbe8dbd5430f26e19e754d38167cd8084a32d702c695550e2ee0063dc2912430`
- authentication: `66941efc779f4f2431d65fdd5b2387b2af67bd7c1559e09dac032f82259ea16e`
- catalog: `e26d3fc758a34ebb0b5028be15d28e3e09db87dc7984bba7bad6a00b26329aa0`
- manifest: `0c6e1c9e9853e82e714e919ba23fa33dbf267c8a36ae36aa9457c1a9d3efc64a`
- exact configuration: `a8d5687519da4040d2f23989f86f9bc804d30ac6c91a85074397eaf7c92de712`

No public admission, historical evidence replacement, or qualification claim is made. Root integration and external review remain separate gates.

## Publication

The remote role guard passed immediately before push. Source branch `codex/vickie-read-update-admission-promoted` was pushed successfully; an independent `git ls-remote` read confirmed exact `d9a22c312b5cc0a46a68852f2be6511fb1b39e14`. No PR, bus, or launch mutation was made.
