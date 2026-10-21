# C1 native code review

## Findings

No High/Medium or production correctness defect found in the five-file C1 delta. One bounded Low test-strengthening finding:

**P3 — pin known current position on installed HEAD/conditional/incompatible-context responses.** `apps/bdptest/src/transactional-collection-minimum-installed.test.ts:389–393` checks epoch/view equality but checks position only against a regex. The successful HEAD and304/412 callers at450–462 and incompatible-context callers at486–502 do not add an exact position comparison. No writer advances the target during this phase, and the seed response already supplies its current position. A stale or invented syntactically valid position on those responses could therefore pass despite the accepted requirement that they retain actual current context. The ordinary positive GET and timeout paths correctly add equality at449/471. Add the same equality for the fixed-head cases (and pos-0 for empty genesis if strengthening the helper), with a narrow response-context corruption control. Keep continuation checks using the later mutation's context rather than forcing all responses to the seed position. This is an evidence limitation, not an observed runtime defect or a basis for whole-profile claims.

Recommend accepting the production implementation; address this small evidence refinement before describing installed conditional/HEAD observations as checking the exact current position. No broader C1 redesign indicated.

## Exact reviewed scope

- Author source: `/Users/dbox/repos/wt/vickie/vickie-tx-collection-minimum`.
- Frozen commit `71c7aa6f9d81ea32965a1c5d5a15fd4b89bb092c`, tree `5f4f278136b1ac4fb33cfc4711b0cb111949d025`.
- Compared with main `581bad0de6f5eb098b920147f6a7e09809c553ea`, tree22e448.
- Exactly five changed files: production `transaction-read.ts`, `transaction-http.ts`; direct tests `transaction-read.test.ts`, `transaction-http.test.ts`; new installed collection-minimum test. No store/evaluator/configuration/spec/schema/catalog/public export or admission changes.
- Reviewed accepted plan4252bf3e… and root acceptance, actual full diff, relevant unchanged owner/materialization/wait/projection boundaries, author frozen inventory. Frozen source was clean before and after review.

## Implementation assessment

**Route scope:** `isInitialTransactionCollection` is the single private helper used by HTTP and facet. It requires collection beads/links and no continuation. The previous point/properties predicate is preserved. Types, incident views, missing anchors and continuation negatives remain pinned; the old beads negative was deliberately moved into positive coverage. HTTP parsing still enforces canonical Scope/query grammar before admission.

**Refusal precedence:** the callback returns only for named invalid-parameter/limit-exceeded. Under the newly allowed initial collection path these arise from the existing materializer control wrapper before graph materialization. No representation is returned on that branch, no arbitrary store/provider exception is swallowed, and resource-not-found retains the existing visibility gap. Existing Read400/413 classification is tested with behind minima and a zero-wait oracle. The generic sourceCondition callback has no route argument, but its current admitted route set and actual refusal provenance support the narrow policy; there is no unrelated public API expansion.

**Entry, wait and retention:** valid data still materializes and checks minimum inside the same owner entry; finish/firstPageWithContext remains afterward. The direct small-capacity test makes four behind attempts and observes zero token allocations, then successfully creates a page. It also compares retained page data/snapshot context against a later current serving context. Deterministic HTTP tests commit through the actual server during the wait hook and require both literal changed data and the satisfying position. No mutable last-position field or secondary reader was added.

**Cancellation:** HTTP forwards the real Request signal on both minimum and ordinary reads. Existing service/facet checks consume it; synchronous owner work is not claimed to be preemptible. Direct pre-aborted handler Requests prove no owner SQL, while real Node disconnect tests synchronize on the server AbortController calling the actual abort implementation before releasing their wait gate. They assert only one wait and close the runtime. Independent removal/restoration confirmed these four tests are forwarding-sensitive, not client-only cancellation checks.

**HTTP/context:** minimum checking precedes final conditional semantics, HEAD body stripping and pagination retention as intended. No collection ETag was introduced. The installed test checks actual wire response presence, canonical Host, literal IDs/properties/type/metadata, record revision presence, timeout/foreign-context statuses and Problems, nondisclosure and retained pagination. The P3 above narrows what its HEAD/conditional position checks currently prove.

**Identity and boundaries:** installed execution uses ordinary plain configuration and the existing launcher/binding; source declares itself non-attesting, with actual payload/configuration binding. The view header is supplied through the existing declared credential-resolver field rather than weakening exact transport rules. Routes are scoped exact per-request targets, no public wildcard. Data derives from a three-Bead/two-Link fixture. Safe report contains hashes/completed response summaries and explicit closure, not credentials. No matrix or whole-contract flag changes.

## Independent verification

Used the observable-test-runs skill. Created an isolated git archive of71c7 under `L/transaction-collection-minimum-native/source`; copied emitted package output and used read-only tooling dependency links. Redirected Vitest cache into the review directory. No author source, build output or packaging work was edited. No builds or installed child cohort was independently rerun.

1. Initial pinned pnpm11.20 invocation exited1 before test discovery: its dependency-status check attempted an install and refused modules purge without a TTY. This is retained as setup failure in `focused.log`, not test evidence. No retry with purge/CI/install authorization was made.
2. Direct pinned Node24.16 invocation of existing Vitest ran `packages/server/src/transaction-read.test.ts packages/server/src/transaction-http.test.ts --maxWorkers=1 --reporter=verbose`: **exit0,80 tests/2files,2.53s**. Log `focused-direct.log`.
3. In the disposable copy only, removed the HTTP signal forwarding from source and emitted JS. Ran the four tests matching `pre-aborted server Request|synchronized actual Node disconnect`: **exit1,4 assertion failures,76 skipped,2.90s**. Pre-aborted requests returned503 instead of aborting; disconnect wait counts exceeded1. Log `no-signal.log`.
4. Restored both disposable files exactly and reran the same filter: **exit0,4 passed,76 skipped,839ms**. Log `restored.log`.

The copied/restored emitted HTTP and Read modules were byte-compared against the frozen author's current output after testing; both match. Review runner processes completed; no review-owned processes remain. No full suite, new packed execution, or public qualification is inferred from these focused tests.

## Author packed evidence now available

During review, `L/transaction-collection-minimum/packed-71c7aa6f/summary.json` and actual positive log became available. They record exact71c7/tree5f4f, clean source, packed protocol/server digests, positive exit0 (4.944s), and source-condition-removal installed counterfactual exit1 as expected. The genuine positive report records36 exchanges, child95140 closed with exit0, workspaceDirty=false and claimEligible=false. These are inspected author artifacts, distinct from the independent direct tests above. The installed mutant is reported as a failed discriminator, never passing target qualification. Static/smoke/Read-seal stages in that summary report exit0; this review did not rerun them.

No outward actions or frozen-source edits were performed. Read/Update/TX public admission and historical evidence remain unchanged.
