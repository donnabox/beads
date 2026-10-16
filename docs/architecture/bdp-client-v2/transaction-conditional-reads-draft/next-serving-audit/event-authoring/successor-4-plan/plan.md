# Successor 4 plan — three actual lint errors only, CLOSED

The preserved successor-3 lint log reports three errors: noAssignInExpressions at lines 432 and 444, and useIterableCallbackReturn at line 861. Typecheck and format-check completed successfully before lint failed. These three constructs belong to the new Event restart callback; none appears in the committed HEAD historical test. They predate the successor-3 report-local const correction. This is the first reached lint stage for this draft, not a regression introduced by that const.

After root accepts this plan and separately releases authoring once the active review's frozen inputs can change, make exactly three local edits:

1. In closeStream, replace the return-wrapped nullish assignment with the same nullish assignment as a statement, followed immediately by returning the memoized property:

```ts
owned.closing ??= track(
  owned.recorder.close().then(() => {
    owned.drained = true;
  }),
);
return owned.closing;
```

2. In closeSession, make the identical syntactic transformation: `session.closing ??= track(handle.close().then(...));` as a statement followed immediately by `return session.closing;`. Preserve the entire existing then body byte-for-byte: capture actual exit code/signal, reject any nonzero/non-null exit, then mark closed. Preserve the prior missing-handle rejection and stopUnknown behavior.

3. Give the observedFrames forEach an explicit void block body, without changing iteration or assertions:

```ts
observedFrames.forEach((frame) => {
  sameFrame(frame, e1);
});
```

Both close holders are private ordinary objects with plain closing properties, not accessors or proxies. Nullish assignment still reads the same holder/property, evaluates the same RHS only when absent, tracks the same promise, writes it once and returns that same promise. No await, task/microtask scheduling point, early memoization or new asynchronous boundary is introduced between assignment and return. Existing fulfilled or rejected closing promises are reused exactly as before, so repeated calls do not retry close. A synchronous RHS exception still exits before assignment/return. This does not claim to redesign the existing pre-assignment synchronous reentrancy behavior: both versions evaluate recorder.close/handle.close before setting closing. The actual recorder close destroys its socket and awaits its closed promise; the launcher close already memoizes its own shutdown promise. Neither invokes this callback's closeStream/closeSession synchronously as a callback. The owned reference remains captured before closing, preserving holder identity.

The forEach body still calls sameFrame exactly once per observed frame, in the same order, synchronously. sameFrame currently has no explicit return; its assertion failures still propagate immediately. The discarded concise-arrow return had no consumer. This is not a switch to async iteration or a weaker frame predicate.

Preserve module setup, historical callback, seed, report-local const, redaction, source attribution, controls, all success/failure checks, five-request/two-child workload, shared cleanup and 45/10/60/120 deadlines. No changes to launcher or recorder helpers. No warning sweep: the 15 noNonNullAssertion warnings are recorded but excluded from this bounded correction. Do not apply an automated unsafe fix or suppression.

Freeze a separate before/after source, correction/full patch and binding after authoring. Native closure must establish only these three edits (formatter wrapping if needed), unchanged historical prefix and seed, and unchanged lifecycle/assertion bodies. Keep successor-3 static results and all earlier evidence immutable. Future validation needs a separately rebound closed packet and new explicit release; no retry, source edit, formatter, helper execution, test or build is authorized by this plan. Actual fresh static checks remain required; no pass is predicted here.
