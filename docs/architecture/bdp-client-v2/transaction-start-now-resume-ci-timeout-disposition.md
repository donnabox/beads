# PR116 combined-head CI timeout disposition

Exact8c78cba5d2be5d5612c6a9324ee36ed690fdeb19 CI38587341773/job115933355104 attempt1 was cancelled by the15-minute job deadline, independently confirmed by the GitHub annotation. Raw job logs show the full Vitest result:168files passed,9216tests passed,1optional skip,696.14s. The last deletion control completed at15:29:14 immediately before cancellation. No failing-test marker was found. This is not green CI: preflight, smoke and Read evidence gate were skipped and must still pass in the required workflow.

Go source installation took168seconds, and test execution began at15:17:37 after job start15:14:06. Setup plus the measured test duration exhausted the whole-job deadline. Prior PR115 raw-log diagnosis already identified broad runner-throughput variation and the sequential cumulative-matrix tail; that earlier successful retry does not erase this observed headroom issue. No deadlock or specific production defect is demonstrated by this run.

Root authorizes one unchanged-head retry of this combined candidate, preserving the original cancellation and full raw logs. This bounded retry must pass all steps before readiness; it does not reinterpret passed test output as a successful workflow. No source, test, request/cohort/runtime deadline or job limit changes. If the same exhaustion repeats on this head, stop automatic retries and require a measured, separately reviewed CI scheduling/job-budget repair. Independent source-only direct-refusal drafting remains permitted; no new product semantics depend on this infrastructure failure.

Raw log SHA256: d0c70ea7d8c621df281e69cf9f96bc742201dfde6ee36bd0ba9626a35ab767b6
