# Positive HTTP receipt-page recovery witness

Frozen clean **d9949a03607af6060ad0dde9891fa40492008b91**, direct child of
**992ccb1bc3174a0e277ce5e861213c5292e8f668**.
Checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-page-witness`, branch
`codex/vickie-transaction-page-witness`, GitHub origin configured. No outward writes.

Only `packages/server/src/transaction-runtime.test.ts` changed (21 added/14 removed
lines). No runtime, store, helper, schema, public profile or unrelated cleanup edit.
Addresses the inherited Low in `transaction-recovery-witness/claude-review.txt`.

The durable-pages case now requires a nonempty continuation URL, so its page branch
cannot silently skip. Both actual page fetches require HTTP200. The first page must
name the submitted receipt and contain created entries at the expected indices2/3;
the reopened page must equal that positively established page. The existing real
child crash, canonical `/stable/pages/` Scope and changed-listener-port assertions
remain unchanged.

Verification used pinned Node24.16.0 (recorded in node-version.log) / pnpm11.20.0,
frozen install and a build before the emitted-child tests. Every command has a log
and explicit `.exit` in this directory:

- Counterfactual: temporarily suffix both page URLs with `-missing`, run the pages=true
  case through the real HTTP service, then restore the exact authored test. Actual404
  fails the first status200 assertion at line93. **1 expected failure/1 filtered skip**,
  counterfactual.exit=1. This is a wrong-request control, not a changed authority.
- Restored runtime suite: **2 passed /1 file**, no skips,754ms; green.exit=0.
- Install, build, typecheck, scoped format/lint, diff-check and local commit each exit0.

No other suites or smokes rerun for this one-file test change. No interrupted tests,
bypassed hooks, leaked disposable stores or test-owned child processes remain. All
predecessor trees remain frozen. This establishes the specific positive page witness;
it does not add profile qualification or power-loss durability claims.
