# C1 repair native closure

## Findings and disposition

No new actionable findings. All four accepted findings are closed by this bounded doc/test successor. Production behavior is unchanged; this review does not grant full-profile qualification or substitute for root's ordinary integration and CI.

Reviewed `20d61119b5faf7251a5d285c5c34a1049866ffe5`, tree `3de0c4b5589423d01eb0dcb7df7ad150d184273d`, against direct parent `71c7aa6f9d81ea32965a1c5d5a15fd4b89bb092c`. The full delta is exactly three files, 162 additions/18 deletions: the installed collection-minimum test, direct transaction-read test, and transaction implementation progress document. Frozen checkout `/Users/dbox/repos/wt/vickie/vickie-tx-collection-minimum` remained clean and unchanged. Read prior native/external reviews, the accepted repair scope, author report, pack supervisor, actual logs, and packed summary.

## Closure mapping

1. **External1, stale progress claims — closed.** `docs/design/transaction-implementation-progress.md:623–671` now includes initial Bead/Link collections, identifies named control-refusal precedence and pre-retention minimum checks, and describes actual Request-signal forwarding. It distinguishes next-loop observation of cancellation from synchronous materialization/current sleep, while leaving receipt and changes/events behavior unchanged. Named unsupported routes and unresolved continuation/visibility ordering remain explicit. The text matches the already accepted production source rather than claiming broader cancellation or minimum support.

2. **External2, refusal nonallocation — closed.** `packages/server/src/transaction-read.test.ts:606–636` installs an independent token-generation counter and checks over-limit, malformed Selector, and oversized Selector against a behind minimum. Each must resolve to its named Problem and leave the counter zero. This measures the existing pagination allocation seam, not an inferred body field. Existing HTTP tests separately retain status/code/no-wait checks and parser-invalid controls. The new test supplies the missing allocation assertion; the prior review's suggestion that all existing refusal tests would pass a generic 503 regression is not required for accepting this repair.

3. **External3, accepted filters and direct empty collections — closed.** Direct tests at `transaction-read.test.ts:638–698` cover both empty kinds at pos-0 and behind pos-1, plus matching and nonmatching source filters with satisfied and behind minima. Installed targets at `transactional-collection-minimum-installed.test.ts:428` include a valid Type filter. The fixture now gives the two existing Links different sources; the source=a query at `:533–539` must return only the independently identified Link l. The preceding complete collection assertion pins ordered IDs and titles, so `slice(0,1)` does not make arbitrary returned membership acceptable. The valid nonmatching Type query at `:541–547` requires an empty Bead page; dropping that filter returns three known Beads and fails. These are meaningful bounded filter witnesses, not a claim to test every supported filter combination.

4. **NativeP3, exact current-position oracle — closed.** The common helper at `transactional-collection-minimum-installed.test.ts:389–397` requires exact expected current position. Genesis pins pos-0. Stable seeded GET/HEAD/304/412/timeout/foreign-epoch/view responses use the seed's current position; retained continuation deliberately uses the later mutation's position at `:583`. Controls at `:458–479` clone actual already-checked HEAD responses for each collection, substitute syntactically valid pos-999, and require the named exact-position assertion to throw. These are honestly labelled assertion-input sensitivity controls, not extra wire requests or modified target observations. Actual packed logs record both controls rejected with pos-1 versus pos-999.

## Independent checks

Applied observable-test-runs. Created an exact git archive under `transaction-collection-minimum-repair-native/source`, copied emitted outputs, linked existing tooling dependencies, and redirected Vitest cache to the disposable review directory. No build or source write occurred in the author's checkout.

Pinned Node24.16 command from the disposable source:

```text
node node_modules/vitest/vitest.mjs run packages/server/src/transaction-read.test.ts packages/server/src/transaction-http.test.ts --maxWorkers=1 --reporter=verbose
```

Actual **exit 0**, **84 tests passed / 2 files**, **2.53 seconds**. Log: `transaction-collection-minimum-repair-native/focused.log`. An initial tool invocation used the not-yet-created disposable working directory and failed before any command or test started; setup was then run from the existing log directory. That tool setup failure is not a test failure or passing evidence.

No broad suite, build, or installed child run was independently duplicated. The direct run completed; no review-owned child process remains.

## Fresh packed evidence inspected

`transaction-collection-minimum-repair/pack.py` explicitly requires exact clean20d611 before building/packing and rechecks cleanliness afterward. `packed-summary.json` binds exact source/tree/direct parent, actual tarballs, installed payload/configuration/policy/authentication/fixture identities, operation targets, and safe completed response hashes. SourceCommit is the launcher's checked source label, not independent target source attestation.

Independently recomputed actual tarball SHA256 values and matched the summary:

- protocol: `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`
- server: `c6ea9254b268f25099d0e2e96100e6ffeabb900ede4b997c032c118cb2d3b369`

These equal the predecessor packages because this successor changes only documentation/tests; fresh execution still binds the new test source. Inspected actual positive and negative logs, not just the narrative report:

- Genuine positive: exit0, one test passed, 38 completed requests, both position controls rejected, clean source, child closed/exit0. Supervisor duration4.975s; reported test timing approximately4.65s is a different measurement.
- Separate deliberately mutated installed copy: exit1, expected actual304 versus expected503 assertion failure after9 completed requests. Its payload digest differs, outcome is failed, and it is never counted as source-matching target qualification.
- Total47 completed requests/two sequential children, below unchanged80/three limits. Positive maximum body2342 bytes is below32768. Summary reports zero remaining children. All recorded build/static/smoke/Read-seal stages exit0; those are inspected author results, not independent reruns.

The two context controls prove assertion sensitivity only; the unchanged full installed requests establish actual target observations. Historical71c evidence is preserved, and all current claimEligible fields remain false. No schema, catalog, matrix, configuration, production runtime, or public admission change is included.
