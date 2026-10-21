# PR113 second-attempt cancellation test: native closure

**No actionable finding. The accepted external F1 coverage gap is closed by this bounded test-only repair.** This is independent source/evidence inspection, not a newly executed test run or CI/merge approval.

Reviewed clean head `c534fcab840ef2e3753ab179feae10d3ce6431e5`, tree `48537438975a4c56ee1a5bb6056013bbddd38e03`, direct parent `5d854a1da21a377f03c2d4cd523b10f1e7659cea`, at `/Users/dbox/repos/wt/vickie/vickie-current-minimum-receiving`. The complete delta is72 added lines in `packages/server/src/transaction-service.test.ts`; all657 other tracked files are identical to the parent. Independently removing precisely the added test from the current file reproduces the complete parent test bytes. Current test-file SHA256 is `4b437a056e18c7a8942604fb539a46bae0e5779dd541dfb6c9b336bc93e83118`.

## Why the new witness discriminates the missing path

At service.test.ts:1658–1667 the wrapper enters the real owner's `withRead`, forwards both reader and context to the actual authority callback, asserts the original signal is active on entry2, then aborts it **after that callback returns and before the wrapped owner read returns**. No caller signal property is shadowed. The first read is genuinely behind: the fixture seeds pos-1, requests pos-2, reaches the controlled wait with exactly one entry and an active signal, and an independent committed create advances the owner to pos-2 before releasing the wait (:1679–1706). The controlled clock avoids timing dependence.

The authority has already passed its pre-work signal check and installed the abort listener (`authority-read.ts:1068–1078`). It must sample abort after the work (`:1104–1107`). Assertions at service.test.ts:1708–1713 require the specific `ScopeServerOperationAbortedError`, no delivered result, the original controller aborted, exactly two owner reads and one wait. The signal is checked active while waiting, after the genuine advance, and immediately before the second callback. This does not merely retest caller-option immutability or wait-loop cancellation.

Production `transaction-service.ts:319–323` still passes the captured options into each `facet.perform` and the same signal separately into the wait loop. The recorded derivative changes only the former to `undefined`. In `transaction-consistency.ts:106–109`, the loop checks the active signal before the second attempt and returns its successful result without an additional post-attempt check. Consequently, dropping attempt-level options loses the abort sampled during owner work; the new test's error assertion must distinguish that mutant. The supplied one-line patch matches this exact path, leaving wait-loop signal handling intact.

## Evidence inspected, not rerun

Read the full original external review and root dispositions, author report/source/summary/counterfactual records, exact patch, actual new/focused/mutant JSON results and failure log, final typecheck log, current authority/service/wait source, and package/build configuration. Independent read-only hashes agree:

| Artifact | SHA256 |
|---|---|
|repair report.md|`c44f03abdfeb565576d7d557c1fb68fc9570ee93792e37fcc5e46a3085ba36b6`|
|source.json|`e9c1875369a2468b9327ed2a8a9dd46ad17593f12af9c4f5eeab235e247eb6b6`|
|summary.json|`501141baae886070edf5cdb46a2fda080892d5045efe169f0d09f2421ad2c348`|
|counterfactual.json|`91fa8b13808550e206b5fa0948dd5ff3eb724c9e34c50239671dbceb12613bb0`|
|signal-dropping-mutant.patch|`aaaae2d4145f487e0cce86f2cfd4d5affa6278fb71ee00ee0952fdfa95c76098`|

The actual stored results show the new test1 pass/0 fail/58 filtered; affected read/service/HTTP cohort188 pass/0 fail/0 pending; and the exact mutant0 pass/1 fail/58 filtered. The mutant fails at line1709 with `expected undefined to be an instance of ScopeServerOperationAbortedError`. This is the intended oracle, not a setup/build/timeout failure. Recorded pinned Node24.16.0/pnpm11.20.0 commands show successful final typecheck/format/diff checks and successful mutant install/build followed by expected test exit1. The report honestly preserves the initial missing-context-argument type error and accidental nonexistent `corrected-typecheck` command, followed by the correct successful command. Filtered tests are not counted as passes.

## Preservation and scope

Package configuration exports `dist`, and `packages/server/tsconfig.json` excludes `src/**/*.test.ts` from emission. All production/build/package inputs are identical to5d854. Independently read/hash/listed the historical protocol/server tarballs:63/246 files respectively, no changed service test, hashes `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227` and `ebe17f4c52c685c29444252df996996d5e5683512c4599b9288946617d38fb3e`. This supports package-input preservation, not a claim of new c534 installed execution. The report correctly keeps all previous installed observations and payload `535612f787a2c3e855cab57801599c86faba1ed78e5952534d70c036e64f442a` attributed to5d854 and explicitly sets freshInstalledExecution false.

Root-deferred external F2–F4 remain deferred: matching rejection predicates are not changed, ordinary-facet Promise behavior is not changed, and helper extraction is not added. This report does not reopen those latent/out-of-scope matters or claim them repaired. Original native5d854 review remains the production-code review; this closure covers only the added test and its evidence.

The derivative directory is actually absent; recorded cleanup reports normal removal and no owned mutant processes. I launched no test/build/service and therefore own no such process. Git remained clean. Only this external closure report was written; no source, network or outward action occurred. Exact c534 CI38561881669 and any PR113 merge remain root's separate gates; no public/whole-profile claim is made.
