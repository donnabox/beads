# PR127 evaluator-work callback timeout: bounded proposal

**Recommend only an explicit 10,000ms timeout on the one existing callback, after native/actual Claude review.** Preserve its full input, work ceiling, assertions and cleanup. This accommodates the observed CI margin problem without asserting that CPU contention is a proven root cause or that the failed CI has been repaired by an isolated pass. No source edit or execution was performed for this proposal.

Source: clean `3df82ac480f50c8f486b4e456336366e7acbdf87`, direct parent `88c9408e908590ab1a4c64c9b9dee1a8b705999c`. Sole proposed path: `apps/bdptest/src/development-contracts.test.ts`, callback at628:

`keeps an actual evaluator work refusal local and permits correction using the same unclaimed key`

## Exact proposed edit

Add this short comment immediately before the callback:

```ts
// Exercises the full evaluator work ceiling directly and through HTTP;
// allow CI contention headroom without changing that ceiling.
```

Change only this callback's final `});` to `}, 10000);` (normal formatting may wrap the existing call). No global Vitest timeout, hook timeout, CI job deadline, evaluator budget, request timeout, fixture input or retry change. This is a test wall-clock ceiling, not a product performance guarantee. Receipt source authoring remains closed.

## Evidence and uncertainty

- Original PR127 CI `38728209051` failed this callback at the default5,000ms (reported5,029ms). `../landing/ci-full.log` retains the failure and stack at628; suite duration755.84s. Root reports job960s, so this is not the whole-job timeout. No new CI retry is proposed as a substitute for a repair.
- Previous PR126 raw CI `transaction-command-minimum-evidence-draft/landing-successor/ci-full.log:5508` records the identical callback passing in4,430ms, only570ms below the default threshold.
- Root's one exact unchanged callback run is preserved in `targeted-start.json`, `targeted-exit.json`, `targeted.log`, `targeted.json`. It passed1, skipped14, callback813ms, Vitest duration1.15s; supervisor elapsed2.2838637500535697s/exit0/no timeout. The log also records pnpm's already-up-to-date check, not a source change. Root's temporary snapshot contains only Node compile cache, no fixture store. This proposal does not rerun it.
- The wide duration difference and prior near-threshold pass support inadequate callback headroom under the observed full-suite environment. They do **not** isolate CPU contention, filesystem scheduling, compilation or another transient source of delay, nor prove every future run completes within10s. The original CI remains failed. A reviewed successor needs fresh exact-head CI.

## Actual callback work and cleanup

At630–653, the fixture uses4,000 distinct integers with `uniqueItems:true` and full `EVALUATOR_CEILINGS`, explicitly asserting the direct evaluator returns instance `limit-work`. `installed-schema-evaluator.ts:597–605` compares array pairs; `installed-schema-value.ts:13,60–75` enforces the16,777,216 logical work ceiling. The same oversized properties are then submitted over real HTTP, where `development-contracts.ts:97–112` invokes the same full-ceiling evaluator and converts refusal into an internal failure. The callback requires sanitized bodyless500, verifies no Resource, corrects input under the **same key**, requires200, and verifies complete afterstate. Neither reducing the array nor lowering the private ceiling is equivalent coverage.

Exactly **seven HTTP exchanges** on the successful path, one in-process service/listener/owner/store directory, zero installed product children:

| Ordinal | Source | Request / expected result |
|---|---|---|
|1|fixture149|Scope probe GET → discovery Link (normal204)|
|2|fixture152|discovered descriptor GET → parsed discovery200|
|3|fixture153–155|discovered operations directory GET → parsed directory200|
|4|callback655|createBead with4,000-item properties/key work → bodyless500|
|5|callback658|GET beads/work →404|
|6|callback659–663|createBead corrected values[1,2], same key →200|
|7|callback665|GET beads/work →complete corrected Resource equality|

Thus the fixture has three navigation GETs, then **two POSTs and two GETs**, not four further reads. Counts are source-derived; the isolated run does not emit an intercepted seven-response trace. Compilation/direct evaluator calls are CPU work, not hidden HTTP requests.

Fixture allocates and registers one `bdp-contracts-` directory at110–112, starts and registers one server at114–122. Existing afterEach29–32 closes registered servers, then removes directories. The fixture has no product subprocess or explicit restart. Its existing cleanup is sequential and can skip removal if close throws; do not silently refactor that separate limitation in this timeout patch. The original local exit snapshot shows no retained fixture directory. `fetch` uses manual redirect handling, and the product listener's existing10s request/header/socket bounds remain unchanged. A callback timeout is not itself cancellation of all pending work; review final exact-head cleanup rather than inferring it from a larger timeout.

## Alternatives and why this edit is the smallest

An unchanged retry would supply another sample without correcting the already observed570ms headroom. Lowering the work limit or shrinking the array weakens the production-ceiling refusal/key-correction oracle. Splitting direct and HTTP portions adds a registration and changes fixture/evidence structure; avoiding the direct evaluation removes the independent reason check. Global worker reductions or timeout increases affect unrelated tests and CI throughput. Profiling a full CI environment could refine causality but is not necessary to propose this narrow, explicit per-test ceiling; no throughput or root-cause claim is made.

For eventual validation, root may select only the exact callback once on the reviewed successor (existing7HTTP/one owner/listener/dir), format/diff checks, and fresh exact-head full CI. No run is authorized by this document. Product/source byte preservation can retain historical Event installed proof under its original source attribution; no new package/installed Event run is justified solely by this timeout change.

## Independently checked unchanged source inputs

Read-only byte comparison against parent88c found equality for every file below, including the entire test and its runtime fixture. SHA256 values are current file bytes, not executed-build provenance. Event source changes therefore did not edit these evaluator/store/helper inputs; this alone does not prove absence of system-wide interaction.

| Path | SHA256 (identical to parent) |
|---|---|
| `apps/bdptest/src/development-contracts.test.ts` | `722728cac9fb3d85282c0c03aebdd81afc8a7ef0dae7f414ebc67635c7d90992` |
| `apps/bdptest/test-support/development-discovery.ts` | `c5188d67dec391d8b7bb9f6f7b81d8ed52d67432df97cde0b19b08dc2f0f3c56` |
| `packages/server/src/installed-schema-evaluator.ts` | `20662b8fb5457d4233e2cc22e35152aa43a2033b7db3568d42847a8752774544` |
| `packages/server/src/installed-schema-value.ts` | `ba14d1a1ef75b13ac14ea20c3066b3f8b742a05d50ac8899fd7b456fa94b067c` |
| `packages/server/src/installed-schema-shape.ts` | `ca08b4f31ed8a77ea92257bc5154afbf158edb7bfc2b4666b159d2b35d338b57` |
| `packages/server/src/development-contracts.ts` | `182a1a6d06b09f08ca3c83ad0fb27a09a419ad3e99729528381f210973b2c40a` |
| `packages/server/src/development-runtime.ts` | `37e222682645e9d8f2d5fcf62de208640fb6216c9e9f2f279ab6017e8b9d6ca9` |
| `packages/server/src/development-authority.ts` | `d49f89da345234f7eaee83dd872a5c2d7a1a8d38ffc794e5e69d333976d22d49` |
| `packages/server/src/recovery-store.ts` | `48b7901e48998fe29598168a527213ebfb49e1d6e5dc5c2eb6494c862b5de9f9` |
| `packages/server/src/member-executor.ts` | `cfce8e561f9315948c6b3b3def25893a7dfde78e2cc338fa142b859ae209771f` |
| `packages/server/src/sequence-executor.ts` | `bea7623cf47521329fa0b3f65def22bbb0a970b066609f8cba4eac17d2181a27` |
