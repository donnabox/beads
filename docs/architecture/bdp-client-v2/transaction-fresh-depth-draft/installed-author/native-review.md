# Native review — three installed fresh-depth witnesses

One actionable inherited cleanup defect (N1); no additional defect found in the changed request/response oracles. This is a read-only source review, not an execution or packaged-qualification result. The six concurrently authored source-test paths were outside this review.

## N1 — preserve the direct cohort store when pre-ready stop is unproven

In apps/bdptest/src/transactional-depth-key-order-installed.test.ts:652–656, `child === undefined` is treated as proof that the process stopped. The assignment at365 occurs only after launchInstalledTransaction resolves ready. The launcher (scripts/transaction-installed-observation.mjs:39–49,100–108) can instead throw after its bounded startup/cleanup attempt without returning a handle. In that case this callback removes the directory even though it cannot prove the spawned child stopped. The original error still fails the test, so this is not a false positive; it can erase failure evidence or the store while startup cleanup remains uncertain.

Use the already-reviewed sequence/new-cohort pattern: track launchAttempted immediately before the launcher call, compute stopKnown only from no launch attempt or a returned child's terminal exit/signal, retain the directory on unknown stop, and report that state. Do not add requests, retries, children, or a second startup probe. The parent accepted this finding and will preserve the frozen review inputs until the active actual Claude review is terminal.

## Checked behavior

- The selected three callbacks contain exactly13+8+13 planned HTTP exchanges and one launcher invocation each:34 HTTP,3 children/owners/listeners/directories, no reopen/retry/polling. Request body totals are3270+3179+169821=176270B; maxima621/680/48363B. These are source-derived planned counts, not observed runtime measurements.
- Direct batch/singleton retained-conflict and full shallow replay/point assertions remain. Fresh deep creates now advance positions2/4; changed shallow reuse conflicts and the final four-Resource collection includes both deep creations. The accepted two Accept changes are explicit, not silent preservation claims.
- Sequence retained conflict/replay/point assertions remain. New deep execution returns the full named created result atpos2; changed-body reuse is a200sequence containing the exact409 member Problem. Both unacceptable-Accept406 checks retain empty framing, current context, and absence of a depth-limit header.
- New13 uses independent bracket literals and explicit-stack equality for all depth12000 Resource, receipt, Event and Snapshot payloads; no deep value is passed to recursive Vitest equality or native JSON.stringify. Its shallow schema/error/header/report operations do not traverse the deep payload recursively. The old depth41 checks remain bounded ordinary matchers.
- Full receipt checks retain dynamically validated producer IDs/transactions/expiry, exact operation indexes/outcomes/positions and required fields. No-op keeps the original Link/source revisions and lacks effectPosition; exact replay preserves the entire original receipt. Failed late operation is a200failed receipt with operationIndex1 and exact validation diagnostic, then identical failed replay. Current Link and complete Event list independently exclude staged rollback effects.
- transaction-batch.ts675–685,774–785,805–845 and919–944 justify the expected Link Event ordinals2/0/0, properties versus metadata change spelling, owned source revisions and matched-result prefix. transaction-events.ts250–270 selects the Link's lifecycle events without treating this Link-view metadata case as the pending incident-Bead rule. The full Snapshot checks all A/B/C and L/L2 records plus both owned copies, with owner A at the last Link-induced source revision.
- Request13 uses the actual schema-valid request2 receipt.id, then verifies exact Scope receipts prefix plus UUIDv4 suffix. That suffix restriction excludes credentials/query/fragment/encoded path separators. The dynamic route is added only after validation and its actual target bytes/hash are retained; the returned original receipt is checked under currentpos5 headers.
- All exchanges use genuine raw HTTP serialization, canonical Host while dialing the observed loopback listener, complete-write/body digest/length checks, safe selected response headers, and source/entry identity. The binding helper independently checks all12 operation targets and canonical installed metadata. New13 requires clean/external entry when the override is supplied. The reused tests' development fallback is not packaged proof: an outer fresh-archive/consumer/eight-package and explicit entry gate remains required.
- New13 and sequence8 abort/drain with finite work/request/cleanup deadlines and retain unknown-stop stores. The launcher close promise is memoized; abort and explicit close do not spawn duplicate shutdown work. New13 additionally rejects abnormal child exits. Direct13's unchanged pre-ready exception is N1 above.
- No claim of successful all-workspace typecheck: the author retained its failed check caused by concurrently authored source tests. Formatting/source inventory are author evidence only. No runtime, build, package, control, or review CLI was run by this review.

## Source binding

Read source-binding.json, full installed draft.patch, report.md, request-inventory.json and selection.json against accepted plan-v2/ledger context. HEAD remained77b4cf980ec91f66c52a554c3dd8c9de5aee859f. All three actual files match the frozen snapshot/hash at the closing check; the full patch remains6f688147f553a5f43d8ab36c860430bf56eb4d257cfaf3be926d14466c868149.

| File | SHA256 |
|---|---|
| transactional-depth-key-order-installed.test.ts | 2ecfc14d4aff48f9dc1d497cb744a29f7d457179f35d7659f2322068e5b63303 |
| transactional-sequence-depth-installed.test.ts | 163549168b98d6f2cdbb111bee840ee4fc1ccaf16b1a9800c5ea721a87df5e1d |
| transactional-fresh-depth-installed.test.ts | d54acc614e6e372fa1e0dfd0a909ed8719d8a4afbd2d93b81116a47a0dba590c |

No product or frozen packet inputs were edited. This report supplements the separate actual Claude code review; it does not replace it or release execution.
