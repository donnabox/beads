# Independent hot-journal witness repair

Frozen clean **992ccb1bc3174a0e277ce5e861213c5292e8f668**, direct child of
**5e7041bcd64c58d84ff671e6cc3a2bb9a6c8b9ee**.
Checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-recovery-witness`, branch
`codex/vickie-transaction-recovery-witness`, GitHub origin verified locally.
No outward writes. All prior review trees, including set repair a94ab98, stay frozen.

Only `packages/server/src/transaction-recovery.test.ts` and its narrow progress-note
paragraph changed. No runtime, protocol, store, schema, public admission or invariant
changes. Informational F4–F6 remain deferred as directed.

## Repair and preserved evidence

After the actual terminal child dies inside execution and the original rollback journal
has the hot magic header, copy both SQLite database and journal to a sibling filename.
Run the incompatible receipt-page-bound open against that copy at an expiry-eligible
clock. Require the installation-mismatch error and require the copied journal to be
absent afterward: this directly observes SQLite consuming it during recovery.

Before the valid open, independently require the original journal still to have its hot
header. Compatible opens then validate **both** branches: original still-hot crash image
and the already-recovered refused copy. Each must recover one pending admission, expose
60 committed groups/180 Resources, and serve all60 original available receipt/page chains.
Thus the copy does not lose the no-BDP-cleanup/retention witness while the original again
proves direct valid-open recovery from a hot image. The maintenance crash case retains its
separate original-hot-image path (zero pending admissions).

Both opens use finally-close; the existing per-test temp-directory cleanup also removes
the sibling database. Process-crash evidence does not imply power-loss durability.

## Red → green

Before changing the refused-open target, add the assertion that the original hot journal
still exists immediately before the compatible open. Against the old test flow the
terminal case fails with ENOENT because refusal consumed that journal; maintenance passes.
`red.log/.json` records1 failed/1 passed and `red.exit` is1. Then make the two-copy flow
above; both recovery cases pass in the focused regression suite.

## Verification, with explicit exits

Pinned Node24.16.0 / pnpm11.20.0. Each verification step has its own `.log` and `.exit`
in this report directory. Expected red is the sole nonzero exit.

| Step | Exit | Result |
| --- | --- | --- |
| frozen install | 0 | completed |
| initial build | 0 | completed before emitted crash tests |
| intentional red | 1 | terminal witness fails as described |
| format-write | 0 | test formatting only |
| typecheck | 0 | explicit recorded compiler completion |
| final build | 0 | completed before final emitted/process suite |
| regressions | 0 | **217 tests /6 files**, no skips,4.02s |
| dependency boundaries | 0 |11 packages valid |
| scoped format-check | 0 | passed |
| scoped lint | 0 | passed |
| diff-check | 0 | passed |
| local commit | 0 | clean frozen commit above |

Regression files: transaction-recovery, transaction-runtime, transaction-store,
transaction-receipt-pages, transaction-service and transaction-http. In particular,
transaction-runtime's actual emitted child / HTTP path runs both durable-pages variants
and exercises canonical Scope `/stable/pages/`, successful page reads, abrupt process
loss and changed-listener-port reopen. This closes the review's selection gap.

No interrupted runs, bypassed hooks or test-owned subprocesses remain. No full workspace
suite, packaging smoke or sealed evidence verification was rerun for this test/doc-only
change; the exact six-file process/owner/HTTP regression above is the verification scope.
This does not award profile qualification or alter public readiness claims.
