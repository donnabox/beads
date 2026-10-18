# Legacy-import Preview 3 preparation — October 18, 2026

Current follow-up product: `c57725789230945a48ae111b75d70e14068c1fa5`; latest83-case evidence and review status are in followup/diagnostics/README.md. The original source snapshot below remains historical.

Preparation source: `89ca1e3124059c9f292fd362b93413b6b683dc8a`, tree `96d4c0ae35c76cc9654497db83716e8822e811e7`. Parent merge `67896392784918326c775e3cae9c4558520062de` preserves original PR86 source `ca4d72a9ea2590925e5876912df54a78559a580d` and Janet-qualified integration source `b9eebc31f8e0f21b9a6154652476ed87027e9cad`. This independent preparation branch does not retarget PR86 or move any integration/release ref. Janet explicitly selected b9e for isolated preparation; her newer candidate owns the engine slot.

## Changes and review disposition

- Preserve expanded integration routes; adapt installed Issue show to graph-json and comments to current owner/projection. General dependency source is byte-identical to b9e; importer retains its own local/prefix validation. Expanded producer workflow remains an inactive patch, not an executed fixture claim.
- Confirmed predecessor High: retained minted-address registry and changed epoch state could alias imported identities. Refuse `epoch_minted_addresses`; allow absent/pristine epoch1 singleton but refuse changed epoch/markers. Schema inventory also found retained `interactions`, now refused. Infrastructure binding/configuration remains permitted.
- Confirmed Medium: unsupported-record fixtures previously failed prefix admission first. Valid prefixed IDs and cause/message assertions now reach intended state/dependency branches.
- Confirmed Medium: freshness-query errors need named plane and underlying cause. Focused SQL-boundary tests cover table, sequence, address, interaction, epoch and config failures. Missing issue_prefix cases assert ErrInvalidStore + sql.ErrNoRows, zero result and full rollback state for apply/dry-run; those engine cases are authored and UNRUN.
- Confirmed Medium: final current-read validation could mislabel SQL/cancellation failures as a size overflow. Use neutral validation context and preserve wrapped causes; late cancellation cases reject misleading size text.
- Confirmed Medium: dry-run executes staged writes, so failed rollback must preserve ErrOutcomeUnknown. Private mayWrite/commit policy split retains ordinary withTx callers. Fourteen strict SQL protocol cases exercise staged cleanup, callback/cancel causes, ErrTxDone, old reader/writer and commit classification, single callback and connection release. Janet inspected/accepted this narrow local repair; real-driver proof remains pending.
- Confirmed Medium: all operation/reopen/inspection contexts now get independent bounded budgets rather than inheriting initialization's cumulative deadline.
- Confirmed Low: validate RawMessage metadata original UTF-8 before cloning/canonicalization; tests cover malformed key/value bytes.
- Optional SQL character-length clarity: no defect; kv.memory. is ASCII and existing rune count matches LEFT character count. No behavior change for this nit.
- MEM-POL-001 boundary: graph-only --source-prefix (including explicit empty) refuses before ordinary storage open; corrupt-metadata and default omission have focused tests.

Raw completed external predecessor report/exit/hash and reviewed manifest are retained. Fresh native repaired-source and generated BUILD reviews have no actionable findings. The full external review of89ca1e312 completed exit0 with follow-up findings. Source41bbc5591 repairs/dispositions are in followup/README.md; its completed review led to the narrow latest sourceffc139a05 under followup/diagnostics/README.md, whose external review completed; finalc57725789 direct file-only review is also complete without actionable findings. No repaired-source pass is inherited. Gemini excluded. Prior source review never qualifies changed code.

## Measured checks

Canonical scripts/test.sh, no-engine environment `BEADS_TEST_SHARED_SERVER=0 BEADS_TEST_ENV_RUN_DOLT=0 BEADS_TEST_PREBUILT_TEST_BINARY=''`:

- Focused graphstore + CLI selection in focused-runner.log: 66 test/subtest pass records, zero failures/skips; exit0. This includes real code pre-open admission and SQL protocol mocks, not a real graph engine.
- Parser package: 36 pass records, zero failures/skips; exit0. Parser source is unchanged on preparation merge/repair; source hashes bind this evidence.
- Pinned affected-package lint: golangci-lint v2.10.1, GOTOOLCHAIN=go1.26.5, `run --new-from-rev=67896392784918326c775e3cae9c4558520062de --build-tags=gms_pure_go --timeout=5m ./internal/storage/graphstore ./cmd/bd`: 0issues, exit0. At execution HEAD was this merge.
- `python3 tools/bazel/go_srcs.py --check` initially found five stale managed blocks after integration. Ran generator; final check exit0. Before/after logs retained. Hand-maintained SQL mock source/dependency entries are included.
- gofmt, git diff --check and inactive completeness-workflow git apply --check pass.

The new base requires Bazel as final gate. bazel is unavailable on PATH; make bazel-sync Gazelle/mod tidy and full Bazel/nogo/guards remain OPEN. Managed hooks invoke ambient bd and/or broad engine checks; they were bypassed to honor Janet's engine reservation and no ambient-store mutation. Direct focused checks do not replace hooks/full CI/runtime gates. No CLI/producer DB command or engine ran; only compilation/prebuilding for selected pure checks.

## Remaining release gates

Genuine expanded ordinary bd producer corpus and receipts, activating expanded workflow/runfiles, settled/documented key/unsupported relationship/custom-workflow/comment dedup semantics, full exact-source quality checks, actual installed embedded/shared-server qualification with zero required skips, exact source/build/engine identities and owned cleanup. No readiness, merge, default flip or release admission claim. Earlier installed evidence qualifies older source only.

Freshness is a fail-closed inventory of retained resource/history/identity planes, not proof that a workspace has never executed any operation. Inventory details are in freshness-inventory.md.

Source council is complete for currentc57725789 with findings dispositioned; runtime gates remain open. A shared Git-metadata incident required isolated recovery in legacy-import-preview3-recovery; see git-state-incident-2026-10-18/README.md. Do not use affected shared Git bindings until owner audit.
