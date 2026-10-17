# Expanded lint diagnostic — original validation remains failed

One authorized exact lint invocation with only --max-diagnostics=1000 added completed raw1/0.453s. Actual source unchanged; no forced cleanup/group survivors. Full raw output reports1 error,165 warnings,9 infos and no diagnostic truncation. No fixes or qualification rerun occurred.

The sole actual error is packages/server/src/transaction-service.test.ts:4267:26, lint/correctness/noUnsafeFinally, at `if (!outcome.ok) throw outcome.error` in the new current-observation drain callback's finally block. This throw is caught locally and pushed to failures, but the rule disallows it in finally.

Minimal proposed correction, not yet authored: replace the false-outcome throw with `if (!outcome.ok) { failures.push(outcome.error); } else { ... }`, enclosing only the existing postCloseProbes increment, returned owner refusal assertion and directory removal in else. Retain the same surrounding try/catch for timeout/probe/removal exceptions. This preserves exactly one recording of an unsuccessful close, skips probe/removal on that outcome, retains successful-path counts, one1000ms close bound and final aggregate/report order. No suppression, warning cleanup, additional retry or deadline change. Root source-only release and native delta closure are required before any correction/validation.

Original capped raw log6a3d744f remains unchanged and its failed stage is not replaced by this diagnostic. No build/source/installed stage ran.
