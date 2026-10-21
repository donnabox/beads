# CLOSED lint-only successor plan

Trigger: accepted typecheck/format, followed by raw1 lint on frozen author successor-2. Exact lint log SHA256 `d33d74a8112151ff71d2d977a9cc4c53cd08c3f878e15e625bb5eba886e03c2c`. Sole error is noUnsafeFinally at `apps/bdptest/src/transactional-event-missing-context-installed.test.ts:592`; warnings19/info1 remain out of scope. This is a read-only plan, not source authoring or validation. Root requests native review before separate authoring release.

## Exact replacement

Within the existing finally/nested try only, replace:

```typescript
        if (deadlineExceeded || child.child.exitCode !== 0 || child.child.signalCode !== null)
          throw new Error("installed finite Event child close was not clean and timely");
        closed = true;
        if (directory) rmSync(directory, { recursive: true, force: true });
```

with:

```typescript
        if (deadlineExceeded || child.child.exitCode !== 0 || child.child.signalCode !== null) {
          failures.push(new Error("installed finite Event child close was not clean and timely"));
        } else {
          closed = true;
          if (directory) rmSync(directory, { recursive: true, force: true });
        }
```

This keeps the same condition and same Error message. Previously, an unclean/tardy close allocated that error, threw it to the immediately enclosing local catch and appended it to failures. The replacement appends the same error directly and does not enter the else branch. Therefore closed stays false, directory removal is skipped and subsequent final aggregation still fails. The deadline flag and failure metadata remain accurate; late exit cannot restore success.

For a clean timely close, the condition is false and the else performs the original assignments/removal in original order. The local catch stays intact for rejected close/race and rmSync exceptions. If rmSync throws after closed=true, it still records a failure exactly as before. Retain the outer timer cleanup, deadlineReason deduplication, ownership guard, report, and final throw/AggregateError unchanged. No return/break/new timer or suppression is introduced. The helper close promise, observed late rejection and shared expiresAt deadline are untouched.

One file, one existing new-callback block. No production, historical callback, request/response predicate, credential, allowlist, sequence, assertion, number of requests/children/stores, deadline or callback-name change. No new test: existing callback assertions plus fresh lint and real installed close outcome cover this mechanical cleanup-preserving correction. Stop if the diff exceeds this block or tooling requests additional semantic work.

## Freeze and next validation preparation

After native plan acceptance and root authoring release, freeze a new source successor with before/after/fullpatch/correction hashes and explicit unchanged historical/production/workload evidence. Preserve this failed lint and the prior failed format separately. Prepare a fresh closed validation sibling, retaining exact13 source plus2 installed selectors,91+14 HTTP,15 stores/two sequential children and all existing budgets. Explicitly reuse original successful bootstrap from source6d547d6c with unchanged17 dependency inputs/11 tool hashes; do not reinstall or claim it ran on the new source. Actual source mapping must be regenerated from final hashes, not assumed.

Fresh static3 and the existing nineteen-stage chain require root release; no retries in current output directories. Build and runtime remain closed until source/native/council disposition is recorded. The completed Claude review remains attached to its original source; the tiny correction needs explicit delta disposition. No broader lint cleanup or weakening to obtain a green result.
