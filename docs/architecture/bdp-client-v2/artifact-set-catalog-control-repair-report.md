# Catalog denial control order independence

Frozen **2ec7eb5c22ebf00c045154cccbcf0f7a425a0ae2**, tree **aa765230defcd8ce8a0bfd1ed746cb92c6859072**, direct ordinary successor of **87abe223dd27519d900fa015845ee3d90c97819a**. Separate checkout `/Users/dbox/repos/wt/vickie/vickie-artifact-set-catalog-control-repair`; original 87 checkout remains frozen.

Only `packages/conformance/src/read-update-catalog.test.ts` changes, 9 insertions/1 deletion. The forged-Read negative fixture now selects the actual `read-update.alias.unknown-subject` row from the approved continuation, asserts exactly one exists, and supplies only that row. Its exact denial message is therefore independent of other existing or future earlier catalog rows. The missing/duplicate-C assertion also prevents an empty or ambiguous fixture from yielding a misleading control.

The real inventory test still inspects the complete manifests; only the isolated negative-control input is narrowed. The unapproved-filename control, closed two-shard list/index checks, sealed Read and public boundaries are unchanged. No manifest, fixture, coverage, runtime, factory, parser limit or evidence change. An artificial preceding-row mutation is unnecessary for this fix: the exact-ID filter excludes every non-C row before the negative call.

Pinned Node 24.16.0 / pnpm 11.20.0: frozen install, typecheck, clean build, boundaries, format, lint and **7 catalog tests passed**, zero failures/skips. `check.py`, `results.json`, `regressions.json` preserve commands and exits. No broad suite or packed repeat; original 87's 424 targeted passes/Read seal and [481 installed 223 evidence](/Users/dbox/repos/wt/vickie/bdp-integration-logs/artifact-set-execution-main-integration/report.md) remain separate source-bound results. Root must run new exact-head CI; earlier failed CI is not relabeled.

Final source clean; no source-dependent child service was launched by this catalog-only run. No push/outward change. Native reviewer and root were notified of this final exact head.
