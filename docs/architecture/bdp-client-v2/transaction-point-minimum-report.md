# Frozen private current-point/receipt minimum-read slice

- Commit: `d3252087a31529cf97b5bfcfe2c402f02c694b91`
- Parent: `7bfba525bbb66e64647110ac336f76b164b41942`
- Checkout: `/Users/dbox/repos/wt/vickie/vickie-transaction-point-minimum`
- Branch: `codex/vickie-transaction-point-minimum`
- Origin: `https://github.com/gastownhall/bdp.git`
- Clean after commit. Local only; no outward writes or profile/evidence promotion.
- Updated plan/dispositions: `../transaction-next-plan.md` (all 14 external plan
  findings plus four inherited repair-review findings explicitly dispositioned).

## Implemented

Complete supported epoch/view/minimum tuples on existing visible canonical Bead
and Link records, properties views and receipt instances. Immutable tuple snapshot;
authority-only safe integer position comparison; original finite monotonic wait
budget with bounded 5 ms sleep outside the owner. Real concurrent writes can satisfy
waits. Ordinary faults retain their original identity and only typed behind
observations become catch-up-timeout after exhausting the budget.

Point rows and current source context are captured in one exclusive synchronous
owner entry, without pretending that entry adds a SQLite BEGIN. The generic Read
plane has a fixed private condition callback and per-request data slot, no public
Read protocol concepts or request callbacks. Receipt lookup/retention/context and
minimum decision share the actual owner transaction. Receipt nondisclosure precedes
minimum diagnosis. Missing/expired/foreign receipts return 404 without context;
unauthenticated requests return 401 with challenge and no context.

Actual 409/410/503 bodies are checked against canonical transactionalProblem.
Their response headers use the final actual attempt context, including when real
commits advance the head but do not satisfy the minimum. GET/HEAD/304/412 cannot
bypass the check. HEAD waits identically and suppresses only the final body.

Stable shared Read router receives the actual Request; a pre-route classifier
rejects controls on alias, probe, discovery, collection, malformed and unsupported
routes. No mutable global last-position/control variable. Service tracks receipt
and point waits and drains finite waits on close. The source-private point facet
accepts an AbortSignal, but transaction-http does not pass request.signal and the
runtime does not connect disconnects to an abort controller. Actual HTTP point
and receipt disconnects therefore both drain the remaining finite wait budget
and can delay shutdown by that amount. No HTTP disconnect cancellation claim. Existing accepted mutation drain semantics are unchanged.

Inherited command-control fix now authenticates/authorizes before the private gap
without entering key admission. Tests inspect the disposable durable receipt table
after owner shutdown, then reopen it, proving no key row was captured for either
unauthenticated or read-only requests. Batch depth marker is now pinned. Documented
actual authority-form CONNECT 400 versus origin-form method probes 405, and uniform
non-operation wrong-method Allow based on shape regardless of existence.

## Exact file scope

- `authority-read.ts`: fixed private source condition and receiving slot.
- `transaction-consistency.ts` + tests: tuple snapshot, authority comparison,
  typed observations, bounded wait.
- `transaction-read.ts` + tests: same-plane minimum facet and visible-point seam.
- `transaction-store.ts`: optional typed receipt minimum check within existing read
  entry; reused captured context for receipt integrity/projection. No format/write
  identity/mutation changes.
- `transaction-service.ts` + tests: tracked asynchronous receipt/point waits.
- `transaction-http.ts` + tests: approved route classification, envelopes/context,
  condition ordering, auth-first command control gap and actual HTTP matrix.
- `transaction-runtime.ts`: source-private deterministic minimum-wait test seam.
- `scripts/smoke-executables.mjs`: private emitted module inventory and root-negative
  imports; existing export map/deep import checks unchanged.
- `docs/design/transaction-implementation-progress.md`: limits and ownership facts.

## Verification

Pinned Node 24.16.0, pnpm 11.20.0; build ran before emitted HTTP/process tests.

- PASS frozen install (initial checkout), build (`build-repairs.log`).
- PASS final 13-file regression matrix: 498 tests, no skips (`regressions-final.log`).
  Covers all transaction suites, AuthorityRead, Read lifecycle and pagination,
  including unchanged recovery/fencing/fault/reopen tests.
- PASS final HTTP 30 tests (`http-final.log`, also in 498).
- PASS typecheck, dependency boundaries, format check, git diff --check.
- PASS lint, with 3 existing warnings + 1 existing info in untouched files
  (numeric-model, read-update-transport, installed-schema-evaluator,
  read-sequence-lifecycle). No new diagnostics from changed files.
- PASS installed-package smoke (`smoke-final.log`) with new private helper
  inventory/root-negative and existing denied deep imports.
- PASS development smoke (`smoke-development.log`), claimEligible:false.
- PASS unchanged 89-row Read evidence verification (`evidence.log`); no reseal.

Earlier attempts: a test incorrectly assumed point reads use SQLite transactions
and was corrected to observe owner reentrancy fencing; a properties URL was
corrected to `?view=properties`. The first broad run had 497 passes / 1 failure because a
new test opened the exclusively locked database. Corrected the test to close,
inspect and reopen the disposable owner; final 498 passed. These are test corrections,
not claims that interrupted/failed runs passed. Full workspace suite not rerun by
this worker; root owns broader integration verification. No tests/processes were
terminated, and completed runs left no owned runtime process.

## Deliberately unqualified

No full barrier/profile claim. Absent or hidden Resource ordering versus minimum
is unruled: these combinations return the explicit bodyless private visibility
500 gap without adopting either 404-first or barrier-first behavior. Only this
fixed all-visible owner is tested; no changing-policy qualification.

Continuation minimum versus original snapshot anchor/current serving context,
partial/malformed/duplicate tuple policy, bare epoch/view controls and unknown
opaque positions retain named development gaps. No inferred normative 400/503.
No CORS configuration or replica-routing capability was added; Origin remains
refused and this owner uses the permitted local wait strategy. Receipt paging,
set operations, first-class replication Snapshot, events/changefeed/replay and
schema-valid cold Transactional discovery remain required future slices.

No shared evaluator, Update runtime/config, schema/spec, Type work, evidence,
public exports or durable format changes. Frozen b1/8844/7bf remain untouched.
