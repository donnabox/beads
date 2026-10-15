# Native plan review — missing Transactional Event context

**No blocking finding.** The frozen plan describes a narrow, source-supported serving repair and a discriminating bounded validation set. One Low citation correction is recorded below; it does not change implementation, requests, selection or acceptance scope. This is plan review, not implementation approval or execution evidence.

## Exact binding

- Plan: `transaction-event-missing-context-plan.md`, SHA256 **cfeb672740e393b71376b4d7cf39989d7fd8ea75f097195279e2525844db8c9f**, fully read after author freeze.
- Ledger: `transaction-event-missing-context-plan-ledger.json`, SHA256 **457c0c33d0558cb92c4bbdd0ae060f81446656c6562ef59c5eb0e394b7f6ce18**, fully inspected.
- Initial audit: `transaction-after-sequence-serving-next-audit.md`, SHA256 **f56f5cdfe285c5adc6ed4e12a737c520c6ed0177200918222e3f3fcc502ea8be**.
- Source: `/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence`, clean **88c9408e908590ab1a4c64c9b9dee1a8b705999c**, tree **b2d9f2c18502e0778c51123be8b4564af052d803**, reused branch `codex/vickie-event-missing-context`.
- Independently verified all four existing owned-file hashes against the ledger and absence of the proposed new test file. The two production files are transaction-store.ts and transaction-http.ts; service source is not included.
- No product imports, tests, build, services, install, packaging, source/plan edits or Claude invocation. Artifact write is session-guarded. Root retains authoring/execution/publication authority.

## Low N1 — precise HEAD framing attribution

Location: plan “Fresh installed witness,” sentence attributing HEAD representation-length computation to runtime deliver87–94.

The exact intended oracle is valid, but the actual handler path computes the HEAD representation length in **packages/server/src/read-http.ts41–46**, then removes the body. **transaction-http.ts620–624** applies that helper. Consequently, **transaction-runtime.ts87–94** receives no HEAD body and preserves the already-set length; its final HEAD suppression is defensive. For GET, runtime deliver computes the serialized body length.

Smallest fix: cite both stages and state “Read semantics computes HEAD length before stripping the body; the runtime preserves it and emits no HEAD body.” No test or request change is needed. The existing source Event HTTP test at transaction-events.test.ts480–484 already uses the same corresponding-GET UTF-8 length oracle.

## Settled obligation and source seam

The normative basis is concrete. docs/specs/bdp.md5661–5668 requires the same404 resource-not-found for unknown/out-of-projection Event identities.5800–5808 supplies the three Scope context fields, and5830–5835 requires the projected-prefix context actually observed by every Scope-bounded read. A canonical authorized unknown Resource Event source reaches an owner observation today, so attaching that observed Scope tuple does not disclose whether a hidden Resource exists.

The two missing branches are exactly **transaction-store.ts1565 and1648**. Each currently queries identity within synchronous enter and returns frozen missing without context. **enter348–359** already rejects failed/fenced/closed/reentrant owners. **readContext590–598** reads the current metadata position and freezes epoch/view/position. Calling it only within the existing missing branch produces same-entry context without a second service/HTTP read, cached receipt context or broader metadata traversal.

The two explicit private result unions need the context member. **transaction-service.ts390–423** infers and forwards both results through generic waitForTransactionMinimum; the latter returns the original successful attempt value (transaction-consistency.ts85–125). No separate service API change is necessary. Search of actual readEvents/beginEventScan callers found no competing typed missing-result producer requiring another source path. The three exact missing-only expectations identified in the plan are real and need reconciliation.

HTTP already has the desired projection seam: **transactionHttpProblem107–112** takes optional context, and **context114–118** sets the exact fields. Change only missing returns at498/512. Successful branches and event-stream construction remain untouched.

Authentication at **HTTP389** precedes Event owner dispatch; the anonymous401 remains contextless. Source grammar and tuple snapshot validation remain where they currently run. Missing identity stays before unsupported-minimum refusal, semantic minimum checking, cursor checking and invalidReconnect. The final plan explicitly avoids claiming malformed API inputs bypass their existing validation. This preserves current semantics and leaves unresolved Event membership and retained Snapshot minima outside scope.

## Oracles, routing and workload

The source selection is exactly eight intended callbacks: three existing named cases plus four expanded owner cases and one emitted-runtime HTTP case. The existing titles were checked in transaction-events.test.ts. The current owner fixture creates one directory/owner per call, without an implicit reopen; the first selected existing case commits twice, the other two once. The planned four new cases each compare genuine missing context atpos-0 and after one independent commit atpos-1, retain the old immutable value, and exercise unknown-before-unsupported/future minimum/cursor paths. These checks distinguish stale or mutable context from the requested observed value. The plan correctly limits the concurrency claim: source structure establishes the same-entry property; these tests are not a general linearizability proof.

The HTTP control is discriminating: keep the new owner context and remove only the two emitted projection arguments. Hard404/body/framing/authorization prerequisites precede the eight named soft tuple assertions. Eight soft failures permit both JSON and initial-SSE branches to be exercised in one11-request run; final Vitest results, not local catch state, determine failure. Retained actual response observations and cleanup are required even when the matcher failures are expected. Binding the copied emitted graph and the dynamic import path is an explicit prerequisite; a wrong graph, generic exit1, timeout or unexpected failure is not accepted.

The corrected seed target is **https://reference.invalid/acme/operations/batch**, matching transaction-http.ts167–174 and transaction-discovery.ts19. The ledger now pins HTTP idempotency-key **missing-event-context-seed**, Content-Type and Accept application/json; the batch body correctly contains no member key. I independently recomputed its152 UTF-8 bytes and SHA256 **e09f4e2f71b146c5de095f3a841a016c5ad64fe64759e7f217c4d6768fc0eae9**.

The installed ledger has11 requests/four distinct URLs: seed, known Scope Event page, eight Bead/Link × JSON/SSE × GET/HEAD missing reads, and one anonymous read. The final plan now correctly separates full Problem checks for GET from empty-body/representation-metadata checks for HEAD. The current Problem builder produces private,no-store JSON, resource-not-found/not-found/404/after-state-change; it does not emit retry-after for that retry disposition. On404, no stream response is constructed. Full bounded response completion tests that behavior without opening a successful live stream.

Model totals are internally consistent:
- Positive source:8 cases,8 owners/directories,1 listener,11 HTTP,0 product children.
- Emitted control:1 case,1 owner/directory/listener,11 HTTP,0 product children.
- Fresh installed:1 case,1 owner/directory/listener,11 HTTP,1 product child.
- Combined:10 case executions,10 owner lifetimes/directories,3 listeners,33 expected/max HTTP,1 product child.

Four owner cases plus one source HTTP case and one installed case add six full-suite callbacks, with two new network callbacks. The inherited matrices are not selected or edited here. Measured draft must still verify expanded names, every dispatch/counter, body/response/config budgets and emitted-control binding before execution.

## Lifecycle and evidence limits

The plan requires a build before emitted imports, bounded45s HTTP/installed operation budgets within60s tests and120s supervisors,5s exchanges, no retries or replacement child, and independent process/temp absence. Local cleanup preserves primary failures, attempts close and removal independently, and does not infer success from soft assertions. Existing unrelated hooks are not refactored.

Fresh packed execution remains a separate phase after clean-source freeze and package/member/dependency verification. Declared source is correctly described as externally verified attribution, not child attestation. Runtime payload, entry, configuration and actual response evidence remain separate bindings. The plan does not fabricate ConformanceRunResult bodies or claim that retained fingerprints can reconstruct checks performed only in process.

Proceeding to a measured draft would remain root's decision after council. No plan review result here grants public profile admission, whole-clause qualification, runtime budget expansion or source execution.
