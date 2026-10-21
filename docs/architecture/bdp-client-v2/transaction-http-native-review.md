# Native council review — T2b private Transactional HTTP

Candidate **b1de5f342fb8c8b90c6909458018b892534863f0**, compared with T2a **bddc4486fe39676bb84d63e05dda22f1bfa77a1a**, in frozen `vickie-transaction-http`. Reviewed the complete 20-file delta, the ordinary merge of PR86 **03b7f435ef1a2c433dcf37b4910d5b7c75a91ed2**, `transaction-http/report.md`, all 16 dispositions in `transaction-http/plan.md`, and F1–F8 in `transaction-service-claude-review.txt`. Author checkout was clean and remained unchanged.

## Verdict

One Medium method-rejection finding remains. The same-entry receipt/context repair closes the earlier service findings within the declared fixed-policy scope. No additional finding in the reviewed receipt, retention, comparison, execution or lifetime paths. This is a private incomplete implementation; the review does not qualify Transactional or waive the explicitly unfinished required contracts.

## Finding

1. **Medium — method rejection is bypassed by authentication and Fetch construction.**

   - `packages/server/src/transaction-http.ts:272–273`: non-operation routes test authentication before rejecting unsupported methods. An actual unauthenticated `POST /acme/receipts/unknown` returns **401**, an `unauthenticated` BDP Problem, and no `Allow`; it should return bodyless **405** with `Allow: GET, HEAD`. The same branch covers the receipt root and Operation Directory. The independent loop stopped at the first failing receipt assertion, so the other two paths are code-traced rather than separately observed in that probe.
   - `packages/server/src/transaction-runtime.ts:138`: the Node receiver builds a Fetch `Request` before the method router can run. Fetch rejects `TRACE`; an actual authenticated raw `TRACE /acme/operations/batch` therefore reaches the catch and returns bodyless **500**, without `Allow`, instead of **405 / Allow: POST**. This is an ordinary unsupported method, not an unexpected storage/application fault.

   **Basis:** `docs/specs/bdp.md:4599` requires command targets to reject every other method with bodyless405/Allow POST; lines4601 and4604–4605 specify the corresponding Directory/receipt method rejections. The authentication-first receipt ordering at line4618 applies to receipt/page **reads**. Preserve that ordering for GET/HEAD; rejecting a wrong method does not disclose receipt existence or authorize a read.

   **Impact:** the new actual HTTP boundary does not implement its claimed target/method table for unauthenticated callers or methods forbidden by Fetch, despite the existing method test passing. No mutation or receipt disclosure was observed in these probes.

   **Fix:** classify the canonical target and reject unsupported methods before Fetch construction and before ordinary read authentication. Retain raw-target safety checks and the existing GET/HEAD authentication → nondisclosure → minimum-control/conditional ordering. Add actual transport tests covering unauthenticated receipt/root/Directory wrong methods and raw TRACE on command/Read targets, with exact bodylessness and Allow assertions.

   **Reproduction:** private exact-source archive `transaction-http-native/source`, appended tests beginning `native:` in `packages/server/src/transaction-http.test.ts`. Command: `/Users/dbox/.nvm/versions/node/v24.16.0/bin/node node_modules/vitest/vitest.mjs run packages/server/src/transaction-http.test.ts -t 'native:' --reporter=verbose --maxWorkers=1`. Both probes fail their expected405 assertions; observed responses and stacks are preserved in `transaction-http-native/native-method-probes.log`. No author source was edited. Root and the successor author received the exact paths during review.

## T2a finding dispositions

- **F1 closed:** `executeObserved` constructs the receipt/context pair inside execute's original SQLite transaction; `admitObserved` does the same for equal replay inside admission. Service submission delivers that pair directly. No second retention clock evaluation occurs before first delivery. Existing clock-jump tests independently passed for failed and completed outcomes; failed first delivery retains its Problem and successful delivery retains results, while later receipt GET applies expiry normally.
- **F2 closed on mutation/replay delivery:** indexed principal/key admission produces replay and execute produces fresh delivery directly. The test that makes the receipt-ID scan throw still passes for both fresh and equal replay. Receipt GET's principal scan remains explicitly disclosed and was not claimed repaired.
- **F3 closed:** `observeReceipt` checks canonical/safe required/effect positions against current serving position, rejects an epoch mismatch, preserves receipt ID/key/field in the failure cause, and freezes the receipt graph. Persisted-fact position mismatch tests pass. This does not claim replica catch-up support.
- **F4 closed in current scope:** receipt reads suppress rows whose stored epoch differs from the owner. Persisted prior-epoch injection verifies nondisclosure; no epoch-rotation implementation is inferred.
- **F5 closed in fixed all-visible scope:** `sequence` and `sequenceObserved` share the same runner/lifecycle. Observed sequence output pairs immutable original member facts with current context captured through one final guarded owner read. Failed context capture does not emit a guessed observation. This does not implement mutable grants, erasure or barriers.
- **F6 disposed explicitly:** private scheduler documentation requires eventual delivery of uncancelled turns; no invented timeout changes execution or draining semantics.
- **F7 closed:** one private operation cap is shared; service checks actual available completed receipt result length against both that cap and configured page.maximumItems, and requires next:null. Unsupported sets and durable receipt paging remain separate work.
- **F8 closed:** obsolete delivery-unavailable re-read path is removed; new observation/bound failures retain receipt ID and key in their cause.

## HTTP and integration assessment

The sixteen planned dispositions were checked against implementation and existing tests. Stable canonical identity is independent of listener port; unsupported valid categories fail before capture/effects as documented development gaps; implemented Read/receipt/304/412 paths retain current observation headers; supplied minimum controls cannot return unchecked success; POST conditionals do not strip terminal receipts; accepted execution survives disconnect and runtime close drains it; raw body/target limits and actual finite-comparison503 are separate; sequence carrier keys are rejected; actor attribution does not become the trusted principal; prior-epoch and pending-duplicate handling remain explicit. Disposition6 (complete method routing) is the exception reported above. Deliberate listener duplication is a source-private composition choice, not an expanded public interface.

The ordinary PR86 ancestor is present. Comparing the merged candidate with exact03b confirms no difference in the inherited public index/shared Read-routing/development composition or pagination implementation/type-boundary test files. The original public pagination surface and runtime package-root negative test remain intact. New HTTP/runtime helpers are absent from package-root exports; smoke inventory additions retain private deep-import/root rejection checks. No protocol/schema/profile admission or evidence artifact change occurs in the T2b-specific work.

Required barriers, sets, replication endpoints, full discovery/Directory, installed contracts, variable grants/erasure and durable receipt pages remain unfinished. Bodyless500 development refusals are not normative barrier responses or profile evidence. Those disclosed omissions are not newly invented blockers for this bounded increment.

## Independent verification

Private exact-source archive under `transaction-http-native/source`; Node **24.16.0**, pnpm **11.20.0** dependency tree, local workspace dependencies point to private built packages. Build completed before emitted/process tests (`transaction-http-native/build.log`).

- **506 existing tests /11 files passed**, covering transaction store/submission/sequence/service/HTTP/runtime/Read, AuthorityRead, HTTP semantics, sequence lifecycle and emitted public pagination negatives (`transaction-http-native/focused-tests.log`). The runtime test performs actual child SIGKILL after acknowledged commit and exact-key/receipt recovery; this is a process-loss witness, not power-loss qualification.
- **46 additional existing tests /2 files passed**, shared pagination plus merged development-discovery HTTP tests (`transaction-http-native/merge-read-tests.log`). Total independent existing tests: **552 /13 files**.
- **Two independent method probes failed as described**, without modifying production code. Their failures are findings, not baseline test failures.
- Inspected author-reported565/13 and all-check/smoke results; did not independently rerun the full workspace, all smokes or evidence verification. No new attestation or blanket qualification claim.

No author source edits, outward actions or test-owned ongoing processes.


## Closure addendum — isolated method repair

**Finding 1 closed at `8844fe738d7457d6c9d3ea71b0b8d0fa778eecde`**, direct child of reviewed `b1de5f342fb8c8b90c6909458018b892534863f0`, in frozen `vickie-transaction-http-methods`. Read the author method repair report and inspected the full three-file delta. No new finding in this bounded repair.

A shared private `methodResponse` classifies target shape without receipt lookup or principal inspection. The raw listener calls it after raw path/Host validation and before Fetch Request construction or body consumption; the HTTP handler uses the same gate. Wrong-method connections close without consuming a body. GET/HEAD proceed to the unchanged authentication and nondisclosure path. Node's separate CONNECT event now dispatches through that same raw gate using a ServerResponse and closes after refusal; no tunnel is established. Canonical-path raw TRACE and CONNECT therefore reach target-specific405 without passing through Fetch's forbidden-method constructor. Scope/raw-target safety checks remain ahead of that dispatch.

Independent private archive: `transaction-http-native/source-methods`; Node **24.16.0**, repository pnpm **11.20.0** dependency tree, private workspace packages. Build passed before emitted/process tests (`transaction-http-native/methods-build.log`). **49 tests /3 files passed** (`transaction-http-native/methods-focused.log`): all46 author HTTP/runtime/service tests plus the two unchanged native reproductions and a native TRACK parser observation.

Observed closure of the original probes:

- Unauthenticated POST to `receipts/unknown`, `receipts/`, and `operations/`: **405**, `Allow: GET, HEAD`, empty body. All three paths were observed on the repaired candidate.
- Authenticated raw TRACE to `operations/batch`: **405**, `Allow: POST`, empty body.
- The actual raw-socket regression passed60 combinations of PUT/DELETE/OPTIONS/TRACE/CONNECT, authenticated and unauthenticated, across command targets, known/unknown receipts, receipt root and Directory. It asserts bodylessness, target Allow, no authentication/context disclosure; additional POST refusals and unauthenticated GET/HEAD401 checks pass. The existing foreign-principal receipt and conditional/minimum tests also pass, preserving GET/HEAD authentication → nondisclosure ordering.
- The unchanged accepted-execution/disconnect/drain and emitted child restart tests passed in the same focused run.

**Explicit transport limit:** native raw TRACK still returns **400 with empty body** from Node24's HTTP parser before application dispatch. Independently observed in the added probe; this is not a405 result and not claimed as arbitrary-method conformance. The application repair covers methods delivered by Node, including its separate CONNECT event. No parser replacement, full profile qualification or broader transport completion is inferred.

No owner, receipt, retention, schema, public API or admission semantics change in this successor. No author checkout edits or outward actions. This addendum supersedes the earlier open Medium verdict for exact `8844fe738d7457d6c9d3ea71b0b8d0fa778eecde`; the prior T2b scope limits remain.
