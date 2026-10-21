# PR113 review repair: second-attempt signal delivery

Frozen `c534fcab840ef2e3753ab179feae10d3ce6431e5`, tree `48537438975a4c56ee1a5bb6056013bbddd38e03`, ordinary parent `5d854a1da21a377f03c2d4cd523b10f1e7659cea`. Checkout `/Users/dbox/repos/wt/vickie/vickie-current-minimum-receiving`, branch `codex/vickie-current-minimum-receiving`. Clean and guarded-pushed; remote branch independently verified at the same commit. Agent-Signature trailer present. Only `packages/server/src/transaction-service.test.ts` changes (+72 lines); no production changes or new installed execution.

## Findings dispositions

1. **Closed:** one bounded real-owner witness proves the captured signal reaches the retried authority attempt. It seeds pos-1, enters a genuine behind wait for pos-2, commits an independent create to pos-2, confirms the original signal remains active, and releases the wait. Inside the second actual owner read, after the callback returns, it aborts the original controller. Delivery must reject with `ScopeServerOperationAbortedError`, not return the successful read. Exact entry/wait assertions require two reads and one wait. No shadowing signal property or new production seam is used.
2. **Deferred by root:** fallback defensive hardening is latent; capture/authority rejection sets match today. This repair does not broaden production scope.
3. **Deferred by root:** ordinary-facet synchronous-fallthrough concern is unreachable under the existing observed Promise authority contract and outside this slice.
4. **Deferred by root:** low-cost helper duplication is retained to avoid unrelated refactoring.

## Checks and actual failures

Pinned Node24.16.0/pnpm11.20.0. Exact commands, exits, durations and logs are retained in `exits.json` and named log files. The new witness passes; **188 tests across read/service/HTTP pass**, with zero failed/skipped tests in that cohort. Typecheck, format and diff checks pass. The test-only candidate bytes were validated then committed without edits.

Two preparation failures are retained: initial `typecheck.log` reported the callback's missing second context argument; this test wrapper was corrected to forward both reader and context. A script-label replacement accidentally invoked nonexistent `pnpm corrected-typecheck`, preserved in `corrected-typecheck.log`; `final-typecheck.log` records the correct successful command. These are preparation corrections, not production failures. No broad suite was run.

## Exact counterfactual

A disposable detached derivative of frozen `c534fcab840ef2e3753ab179feae10d3ce6431e5` changes only the service attempt from `facet.perform(request, captured)` to `facet.perform(request, undefined)`. The wait loop still receives `captured?.signal`. Frozen install/build succeed. The new test then fails exactly at the expected abort-error oracle: `expected undefined to be an instance of ScopeServerOperationAbortedError` (one failed test, zero passed selected tests). This distinguishes loss of attempt-level cancellation despite preserved wait-loop signal handling.

Patch SHA256 `aaaae2d4145f487e0cce86f2cfd4d5affa6278fb71ee00ee0952fdfa95c76098`; exact patch and failure are in `signal-dropping-mutant.patch`, `mutant-test.json/log`, and `counterfactual.json`. The derivative was restored clean and removed normally; candidate source stayed frozen. No mutant child process remains.

## Production/evidence preservation

`source.json` verifies the single test-file delta and byte-identical original test content around the new case. Every other tracked path, including all production/package/build inputs, remains identical to `5d854a1da21a377f03c2d4cd523b10f1e7659cea`. Historical protocol/server packs omit this test, and their retained hashes/file counts are recorded. No new pack or installed identity is asserted for this successor.

The actual installed evidence remains attributed to **5d854a1** at [original report](../transaction-current-minimum-receiving-implementation/report.md): collection38 measured exchanges/1child, Type20/1, alias18/1, two RU positives/four rows, payload `535612f787a2c3e855cab57801599c86faba1ed78e5952534d70c036e64f442a`. Those historical runs are not relabeled as execution at this test-only successor. Original report SHA256 `fc352b263d3a79f34166956ba26e13c8fd65bfc8bc58a0373ff7096772c658c4`.

No runtime/schema/HTTP/options/limit/public-admission changes, installed request additions, PR/main/bus/launch edits, or qualification claims. Source is clean, commands finished, derivative removed, owned processes absent. Root retains independent closure and exact CI/merge decisions.
