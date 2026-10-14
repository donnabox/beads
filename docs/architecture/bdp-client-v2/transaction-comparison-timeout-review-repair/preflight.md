# Comparison timeout review repair — unexecuted measured preflight

Draft in `/Users/dbox/repos/wt/vickie/vickie-comparison-timeout-review-repair` on `codex/vickie-comparison-timeout-review-repair`, ordinary HEAD `30035b88380b7cbb3abd5359787db236b70345c7` (tree `cdee4bc7e31511cde094d17c045ea618163b6d7a`). Original implementation checkout remains frozen. No commit, build, tests, install, services, package execution, push or network action occurred during this draft. `git diff --check` is a static whitespace check only.

## Dispositions and source

- F1 accepted: extend the existing fenced case to close and assert the same failure still throws. Add a real existing-fixture background maintenance callback failure with interval25ms and fake time advance25ms; verify timer cancellation, direct observation exception message and original cause. This covers the two fault terms of the normal-closed shortcut. Production is already correct; no fault-handling implementation changes.
- F2 accepted narrowly: a submission-only private owner double returns owner-stopped at an actually reached finite deadline. Pin exact contextless comparison-refusal, one wait, one admission and exactly one final callback with the expected epoch. Cancel both existing pending members in finally. No HTTP stub, live epoch-change claim, service seam or runtime change.
- F3 reflow only the existing final progress paragraph. Historical18selected/five-mutant claims remain attributed to frozen30035, not these unexecuted additions.
- F4 informational; no action.

Only three authorized paths changed, 83 insertions /5deletions. Exact file hashes and byte lengths are in `preflight-bounds.json`; patch is `source.diff` SHA `edda7cd65cd321e4ee637130217f8c9a99847de2262a264107cd43fb049727bf`. All 660 other tracked files are independently byte-compared against30035 in `preservation.json`, including production and installed scenarios. No package/installed rerun is proposed for this test/doc-only delta.

## Proposed selection and control (not executed)

Three positive cases: two store titles and one submission title, exact argv in bounds. Two cases are new; fenced case is extended. No HTTP/listeners/services/installed children. The store cases use at most two disposable stores; submission uses one, with no admitted effects executed.

One isolated derivative changes only `if (closed && !ownerFailure && !maintenanceFailure)` to `if (closed)` in transaction-store.ts. It runs the two F1 cases separately, expected one named assertion failure in each: `closed fenced deadline preserves failure` and `closed maintenance deadline preserves failure`. Both must fail at those assertions, not a compiler/setup/cleanup failure. Patch `4dea38875ef699137a64719f8b99a4772af53fb8a86fff95481d9404acb375d8` is external only; production checkout remains unchanged. Two derivative temporary stores total; no HTTP or child services. No additional F2 mutant is proposed.

Maintenance case restores real timers in nested finally even if owner close fails. Existing fixture cleanup independently closes the store idempotently and removes its directory after each case. The new maintenance callback itself cancels its timer and closes its connection; test asserts zero remaining fake timers before direct observation. The fenced case restores its SQLite exec spy before its close assertion. The submission finally cancels its two owned reservations; existing fixture cleanup closes and removes the disposable store even after assertion failure. Do not broaden historical cleanup hooks.

After root release, proposed commands use Node24.16.0/pnpm11.20.0: offline frozen install; typecheck; build before any emitted consumer; exact selected Vitest argv from bounds with verbose+JSON; isolated mutant build and two selected failures; format check, lint, dependency boundaries, historical Read evidence verification, diff check. Logs/exits stay outside source. Per-command supervision proposed120s; normal Vitest test deadlines unchanged; maxWorkers1. No unfiltered store/submission/full/matrix suite. Total positive/control case executions five, temporary stores at most five, timers one per maintenance case, no real-time delay. Process monitoring/owned cleanup follows observable-test-runs.

Execution remains held for root review of this preflight. No new installed evidence or public/default-timeout qualification is claimed.
