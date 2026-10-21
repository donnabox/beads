# Actual run stopped on focused-source failure

Execution stopped at the first failed stage. No source correction, retry, package stage, consumer installation, installed callback, or required check was executed by this worker after the failure. The original two bootstrap records remain explicitly inherited prerequisites, not fresh successor runs.

| Fresh stage | Raw exit | Accepted | Seconds |
| --- | --- | --- | --- |
| typecheck | 0 | True | 6.847 |
| format-check | 0 | True | 5.815 |
| lint | 0 | True | 5.754 |
| fresh-build | 0 | True | 6.894 |
| bind-fresh-build | 0 | True | 12.219 |
| focused-source | 1 | False | 9.152 |

Lint passed with 87 warnings and 4 informational diagnostics; no zero-warning claim. Fresh build performed its existing clean step before compilation. Build inventory contains 476 artifacts; SHA256 `bc12210f12ce40f8aa3d258a7dbe0cd9119f3add9af2814b4da60fa5c5c2674f`.

The selected source run reported 7 passing tests, 1 failure, and 57 skipped tests. The failed callback was `captures initial Scope Event SSE minima once and recaptures context on retry`. Its `advance()` fixture passes receiving.maximumOperations=100 at transaction-event-sse.test.ts:1363; transaction-batch.ts defines the private maximum as32, so snapshotTransactionReceivingControls throws `TypeError: valid private receiving controls required` before the intended mutation. This is a test-fixture receiving-control defect, not evidence that opaque-minimum retry failed. Proposed bounded correction for root disposition: use maximumOperations=1 for this one-operation request. No correction has been applied.

The failed capture report truthfully records one owner attempt, one wait, two clock calls, zero intervening commits, owner close once, closed=true and failureCount=1. It does not establish the intended second attempt or recaptured context. The other new owner and abort callbacks passed; D passed31requests, three pages of2+2+1 Events, proven live readiness and five exact frames before/after intentional drain. Existing finite32 and the three historical SSE callbacks passed. These raw observations are not a successful aggregate source audit; review-focused-source was not run.

The focused-source supervisor recorded sourceUnchanged=true, forcedCleanup=false, errors=[], remainingProcesses=[], runtimeDirectoriesStillPresent=[]. Sampled PIDs93760,93769,93789 were absent at completion. No consumer lock or archive proof was produced; subsequent stages remain unattempted.

Failure evidence: `run/focused-source.exit.json`, `run/focused-source.log`, `run/focused-source.results.json`. Raw log SHA256 `b3e378f44b7025f68187bf9b7c37eb8a8b709dcafafc9868914dfc5206761560`. Source identity `41d71ce7bc02fc4ee8e729e443490d94ee1b0c69469dfc7ddfea24469a5ffaad`.
