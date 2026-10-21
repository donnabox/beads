# Native final source/evidence closure audit — PR64 and PR65

**No new source or evidence findings in the audited scope. All prior native findings are closed.** PR65's serial full-suite rerun was still in progress at this audit; its earlier full-suite failures and focused successful rerun are preserved below. This report does not qualify PR83's separate running cohort or advertise any public write profile.

## Exact source and seals

- PR64 source: `3deafd650958fe0845efcdfe37c868f98b7ec652` in `bdp-pr64-freeze-repair`. Its parent seal is `d653124d21f7137e36aa3660229bc997e7408f0c`, over actual observed source `2a4ede9b17b29c9686dbfd8ae4ee32b9a6adecac`.
- PR65 seal: `c8c8c2d2ab14cd93b677a87c19903eff519a79c7` over actual observed source `a68db90f995c1d3317c19acb0b6dcdf63cd21b0e`. During the audit the checkout advanced to `c352b372761d35425b04afff22f2e5b6d544ba9e`, an ordinary merge carrying only the same PR64 anti-relabel test strengthening. The diff from the seal is solely `packages/server/src/member-identity.test.ts` (9 additions/9 deletions); no runtime, schema, observer, catalog or evidence bytes changed.
- Independently confirmed each run-head-to-seal diff contains exactly the cohort artifact and shipped evidence constant. PR64's subsequent child is test-only; no evidence-affecting source was relabeled as observed after the run.

## Closure checks

1. **Anti-relabel test (Claude N2).** `3deafd6` retains the positive `ok:false`/expected diagnostic assertion and separately asserts the diagnostic has no `document` property. Unlike the prior subset match, it fails if the formatter's injected `document: metadata` leaks through on a properties update. The preceding match prevents the conditional assertion being skipped by an admitted result. The production stripping behavior is unchanged. Root's `diagnostic-test-strengthening.log` records 95 passing member-identity tests; this audit inspected the assertion and its unchanged implementation rather than repeating them.
2. **PR64 Read repair.** Strict current-schema matrix checks, raw transport observer, finite harness binding and metadata selector remain present. Current Read matrices contain 25 strict current-record/collection schema assertions and zero legacy record/collection assertion references. The one-line generator type error remains fixed.
3. **PR65 merged intentions.** Reviewed `d684862` and `a68db90` integration deltas and their final source: PR64 metadata diagnostics/selector/coherence text coexist with common `revisionProvenance`; the owned-Link citation uses the complete provenance vocabulary. Output `minProperties:1`, the corrected nonempty output TypeScript union, permissive empty input, and the recovery-format explanation are preserved. The output type equals the reviewed `85beebd` implementation. The schema diff from that repair adds only the reviewed diagnostic document discriminator. No old attribution wire alias was reintroduced.
4. **Projection gate.** Final direct roots are 16; the sealed set remains 47 definitions with 46 reachable. PR65's regression pin now records the independently observed `bd7da0fc3f8082e0bfb6f4969c04409f2a610e12408dbc6a0bd14b4b351127cc`, closing the previously deferred current-source pin mismatch. This pin update is accompanied by a real fresh cohort, not a rewrite of old observations.
5. **Prior native findings.** Read generator typecheck closed at `3e7b672`; nonempty provenance output type closed at `85beebd`; private Transactional metadata numeric formatting integration closed at reviewed `65ff2f0`. Earlier frozen defaults, Event folding, repair citations and recovery-store boundary findings remain closed. This does not assert that the PR83 cohort has completed.

## Independent evidence verification

Invoked the evidence verifier's exported `main()` explicitly with Node v24.16.0, read-only against each checkout, and retained full result logs. Both exit 0 with `evidence.verified`:

- PR64 `native-final-evidence-verify.log`: constant `e425bb5c8256e306b251800ba1ba738a4f82a1ed`, run head `2a4ede9…`, evidence commit `d653124…`, projection `26db67b2f8611a641df6b44bc42959bf047a9efc8b2d5a7d665fc3dd47d631a7`.
- PR65 `../pr65-review-repairs/native-final-evidence-verify.log`: constant `81e9df94e47cce08d4a505d85406560c5e8bb54c`, run head `a68db90…`, evidence commit `c8c8c2d…`, projection `bd7da0fc…1127cc`.

Both sealed artifacts have 89 passing target-row observations: bdptest 47 (36 packaged, 11 in-process) and bdpbd 42 (34 packaged, 8 in-process). All segment outcomes are pass, with explicit capability-based non-applicability outside those observations. These are not 89 wholly black-box rows. The PR64 actual generation log is `repaired-final-cohort.log`, which matches the final constant/run head; its earlier `final-cohort.log` describes an earlier candidate and is not substituted for final evidence. PR65 `final-cohort.log` matches its final constant/run head.

## Root validation evidence inspected

PR64 `final-tests.log`: 108 files, **7617 passed, 1 skipped**. Build, typecheck, boundary, formatting, smoke and seven E2E preflight checks have successful result logs. Lint completes with three warnings and one informational notice, not a warning-free claim. The later test-only child has its separate 95-test, typecheck and format logs.

PR65 `sealed-tests.log`: **7618 passed, 2 failed, 1 skipped**. Both failures are in unchanged adapter-bd child-process lifecycle tests, with ENOENT reading their PID fixture files. No causal conclusion is drawn from the symptom. `process-recheck.log` reruns that file and the strengthened diagnostic test with one worker: **138 passed, 2 files**. `final-tests-serial.log` was still running when inspected. Build/typecheck/boundary/format/smoke/seven preflight checks pass; lint has the same warnings/info. The original failure log remains evidence and is not represented as a green full suite.

The inherited properties-update numeric-location issue (Claude N1) stays explicitly deferred before public Update advertisement; polish items N4–N6 were not promoted into this development landing. The stale-evidence concern N3 is closed by the source-matched fresh observations above. No new mechanism, full write-profile requirement or external action was introduced by this audit.

## Subsequent validation completion

The preserved PR65 serial rerun subsequently completed: `../pr65-review-repairs/final-tests-serial.log` records **7620 passed, 1 skipped, 108 files** (143.31 seconds). No test weakening or runtime correction intervened. The original two ENOENT failures and focused 138-test pass remain retained. Final PR65 narration-only `074a6e3` and integrated PR83 `2c8b63c` are audited separately in `../provenance-runtime-integration/native-final-source-audit.md`; that later report independently verifies PR83's completed `fccaa537` seal and does not rely on this earlier report's pending-cohort status.
