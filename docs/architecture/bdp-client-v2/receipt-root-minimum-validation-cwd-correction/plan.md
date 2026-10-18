# Closed receipt validation cwd correction plan

The finalized packet has one blocking command-directory error: `consumer-install.cwd` names the validation packet, although `packages.py` creates the external manifest and pnpm workspace overrides under `E/consumer` (`E=P/run`). `run.py` passes `spec["cwd"]` directly to subprocess.Popen. There is no `--dir` argument or gate override. Consequently installation would not target the consumer that later provenance checks inspect. No stage has been run by this preparation worker.

Proposed only executable change after full native findings and root acceptance: change `commands.json` → `commands.consumer-install.cwd` from `/Users/dbox/repos/wt/vickie/bdp-integration-logs/receipt-root-minimum-validation` to `/Users/dbox/repos/wt/vickie/bdp-integration-logs/receipt-root-minimum-validation/run/consumer`. Preserve argv, all21 stage names/order, selectors/counts, package versions, deadlines, no-retry supervision, helpers and source bytes.

Before this later mutation, preserve the entire currently finalized packet in an explicit separate sibling directory `receipt-root-minimum-validation-before-cwd-correction`, including current index, closed gate and any completed native report. Verify every archived indexed file and record all additional archived file hashes. Keep its existing `initial-closed-preparation` unchanged. This plan is outside the frozen packet; it does not itself copy or mutate that packet.

After accepted correction, record the exact one-value commands delta and unchanged helper/source pins in a separate correction report; regenerate adaptation/finalization metadata and the active implementation index, preserving predecessor versions in the archive. Rebind the still-closed gate to the new index. Require independent native/root inspection before any release. No tests, installs, builds or helper invocation are needed for the correction review. No source or actual CODE review context changes.

The prior proven Snapshot packet already had the correct `/run/consumer` cwd. Broad normalization of command directories while adapting the receipt packet erased this required suffix; that was my preparation error. Absolute helper paths are safe with packet cwd because they derive P/W/E from their own files. A package-manager command does not share that property: its cwd selects the package manifest/lock and must be checked against the producer and consumer of those files. Preserve that known distinction during adaptations rather than generalizing helper cwd rules.

All21 cwd/argv pairs were read directly. Workspace install/typecheck/build/pack commands run in the source workspace; formatter/linter configuration and files resolve there; Vitest source/installed selectors run there. Helper commands use absolute scripts in the packet and derive all artifact paths internally. `node-version` has no path-dependent output. The only functional cwd mismatch found is consumer-install. The inherited consumer manifest name `snapshot-opaque-minimum-external-consumer` is unused descriptive metadata, not an installation target or identity predicate; it is recorded here and intentionally excluded from this one-value correction.

| Stage | Current cwd | Assessment |
| --- | --- | --- |
| `workspace-install` | `/Users/dbox/repos/wt/vickie/vickie-receipt-root-minimum` | workspace-relative command/source collection |
| `bind-workspace-tools` | `/Users/dbox/repos/wt/vickie/bdp-integration-logs/receipt-root-minimum-validation` | absolute helper; internal P/W/E paths |
| `typecheck` | `/Users/dbox/repos/wt/vickie/vickie-receipt-root-minimum` | workspace-relative command/source collection |
| `format-check` | `/Users/dbox/repos/wt/vickie/vickie-receipt-root-minimum` | workspace-relative command/source collection |
| `lint` | `/Users/dbox/repos/wt/vickie/vickie-receipt-root-minimum` | workspace-relative command/source collection |
| `fresh-build` | `/Users/dbox/repos/wt/vickie/vickie-receipt-root-minimum` | workspace-relative command/source collection |
| `bind-fresh-build` | `/Users/dbox/repos/wt/vickie/bdp-integration-logs/receipt-root-minimum-validation` | absolute helper; internal P/W/E paths |
| `focused-source` | `/Users/dbox/repos/wt/vickie/vickie-receipt-root-minimum` | workspace-relative command/source collection |
| `review-focused-source` | `/Users/dbox/repos/wt/vickie/bdp-integration-logs/receipt-root-minimum-validation` | absolute helper; internal P/W/E paths |
| `node-version` | `/Users/dbox/repos/wt/vickie/vickie-receipt-root-minimum` | workspace-relative command/source collection |
| `prepare-package-directory` | `/Users/dbox/repos/wt/vickie/bdp-integration-logs/receipt-root-minimum-validation` | absolute helper; internal P/W/E paths |
| `pack-protocol` | `/Users/dbox/repos/wt/vickie/vickie-receipt-root-minimum` | workspace-relative command/source collection |
| `pack-server` | `/Users/dbox/repos/wt/vickie/vickie-receipt-root-minimum` | workspace-relative command/source collection |
| `verify-archives` | `/Users/dbox/repos/wt/vickie/bdp-integration-logs/receipt-root-minimum-validation` | absolute helper; internal P/W/E paths |
| `prepare-consumer` | `/Users/dbox/repos/wt/vickie/bdp-integration-logs/receipt-root-minimum-validation` | absolute helper; internal P/W/E paths |
| `consumer-install` | `/Users/dbox/repos/wt/vickie/bdp-integration-logs/receipt-root-minimum-validation` | BLOCKER: package-manager cwd must be generated run/consumer |
| `verify-installed-consumer` | `/Users/dbox/repos/wt/vickie/bdp-integration-logs/receipt-root-minimum-validation` | absolute helper; internal P/W/E paths |
| `receipt-history-external` | `/Users/dbox/repos/wt/vickie/vickie-receipt-root-minimum` | workspace-relative command/source collection |
| `review-receipt-history-external` | `/Users/dbox/repos/wt/vickie/bdp-integration-logs/receipt-root-minimum-validation` | absolute helper; internal P/W/E paths |
| `receipt-root-external` | `/Users/dbox/repos/wt/vickie/vickie-receipt-root-minimum` | workspace-relative command/source collection |
| `review-receipt-root-external` | `/Users/dbox/repos/wt/vickie/bdp-integration-logs/receipt-root-minimum-validation` | absolute helper; internal P/W/E paths |

This plan authorizes no execution and no packet mutation. All release flags remain false; source and ongoing actual Claude CODE inputs stay frozen.
