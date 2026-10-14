# Internal discovery minimum service — frozen implementation

Source **2bd484367f63c7b8cc25e144baec983c5e579798**, tree **20a494de5555b39ece1e3eb09e9dd2ad37cbd554**, direct parent **26d6cd1b947db8c00e24015771506d26098a293e**. Checkout `/Users/dbox/repos/wt/vickie/vickie-discovery-minimum-service`; branch `codex/vickie-discovery-minimum-service`. Exactly three approved files changed: transaction-service.ts, transaction-service.test.ts and the progress document. Accepted plan SHA256 ee9de18ed49dcb296a91816ee0156e76d9b4a4afca1a5d4fc1519d37655279a6; root acceptance governs.

## Result and boundaries

The existing private discoveryObserved method accepts an optional minimum and signal. It snapshots canonical controls once; each synchronous owner entry checks eligibility and copies the same current epoch/view/head. The existing finite wait owns one service lifetime and retries only behind. Assembly/parsing happens after the entry releases and, on the minimum path, after the wait succeeds. Entry refusal/cause, foreign/expired context, timeout, abort and parser failures remain distinct. Close drains admitted work before owner shutdown.

The no-minimum path preserves synchronous entry timing and its signal remains inert, explicitly documented and tested. No HTTP handler invokes this accessor. Existing discovery/minimum-route refusals, public exports/admission, catalog flags, stores, parser/schema and runtime configuration are unchanged. The missing-Resource callback lane remains independent. No new protocol law or profile claim is introduced.

All preexisting transaction-service test bodies are byte-identical to the parent; only imports were added before the unchanged body, followed by the new suite. `preservation.json` records this check. In particular C3 receipt-root/HTTP200/exact-receipt/pos-2 assertions, replay/expiry/reopen, limit literal and callback receiver/getter-once tests are preserved. PR107 runtime-default alignment tests and repaired discovery assembly files are outside this commit.

## Completed validation

Pinned Node24.16.0/pnpm11.20.0; exact commands, durations and exits are retained in validate.py, checks*.py and exits.json.

- Frozen-lockfile install, corrected prebuild typecheck and build: PASS.
- Focused service, consistency, discovery, Read, HTTP and emitted installed-contract regressions: **183 tests passed,6 files**,0 failed/skipped,3.609s. This includes11 new service cases. It is not a full workspace run.
- Format, lint, dependency boundaries and actual existing Read-seal verification: PASS. No evidence/constants edits or hook bypass.
- Clean source committed/frozen, followed by a new build before direct emitted execution: PASS.

The source cases cover synchronous no-min behavior, equal/earlier/genesis eligibility, post-release parsing, frozen output, immutable caller controls, independent catch-up commit and concurrent observation, both foreign reasons and subsequent owner reuse, deterministic behind timeout and no allocation, pre-/pending abort, close/drain, malformed input, real reentry/closed owner, delegated provider/fenced-style cause preservation, and parser fault recovery. Every held wait uses a controlled clock and cleanup release. Delegated refusal is a fault discriminator, not an actual disk-failure recovery claim.

The source HTTP regression pins minimum-bearing discovery refusal. The existing emitted runtime test sends ordinary authenticated discovery requests only—six before and six after reopen—inside its larger runtime/receipt/restart test. Neither is the separate packed12/1 fixture (six500,three401,write/read/retry200). No new packed fixture was run or credited with this internal path.

## Sensitivity controls

Run in an independent control checkout, never in frozen author source. Full changed source copies, hashes, exact test names/failures and red exits are preserved in sensitivity.json and mutant logs.

1. Omitted in-entry minimum check: both foreign controls and the fixed-head timeout oracle fail (three named assertion failures across two bounded runs). The timeout test fails because the expected error is absent, not because the test times out.
2. Mutable caller tuple retained instead of captured: the paused catch-up case fails when changed foreign controls affect retry (one named failure).
3. Assembly performed while the owner entry is active: the three equal/earlier/genesis tests fail through the real-owner parse-after-release oracle (three named failures). This counterfactual deliberately inserts the real assembly before owner release; it does not rely on a fabricated parser error.

All red runs exited1 as expected without timeouts. Restored unmodified source passed all11 new cases; the later omitted-check/behind run also restored exact source bytes, and the control checkout is clean. These are separate source sensitivity observations, not installed qualification or extra HTTP calls.

## Actual direct emitted-module proof

`emitted.mjs` imports the actual frozen build's private service/store/consistency/read-request modules by file URL. No emitted bytes, module loader or API was changed. One Node process, one disposable SQLite store, zero listeners and **zero HTTP requests**. Real pagination callbacks are supplied; minimumReads uses budgetMs2, tick++ and immediate synthetic wait. Process bound10s; measured0.147s.

Four calls yielded:

| Call | Observed result | Incremental owner entries |
|---|---|---|
| Ordinary | Frozen candidate/context at actual pos-0 |1|
| Satisfied minimum | Frozen candidate/context at actual pos-0 |1|
| Fixed-head future pos-1 | Actual instanceof TransactionMinimumFailure, catch-up-timeout, exact pos-0 context |2|
| Pre-aborted minimum | Actual instanceof ScopeServerOperationAbortedError |0|

Measured totals: **4 calls,4 owner entries,1 synthetic wait,2 candidates,2 exact errors**. The <=10 entry guard passed. `emitted-summary.json` retains actual module byte hashes, observations and process identity. SourceCommit is derived from the checked frozen checkout, not target attestation. This is direct emitted-module evidence, not installed package/public API/HTTP/cold-client evidence.

**Root-approved measurement clarification:** the emitted probe does not instrument the immutable ESM assembler import. Exact assembly call/no-call assertions come independently from source spies: one per successful ordinary/eligible case and zero for foreign/timeout/abort cases. They are not relabelled as emitted instrumentation or combined into a measured emitted assembly count. Root expressly approved preserving real emitted bytes rather than adding a hook/loader solely for that count.

## Preserved failures and corrections

- Initial typecheck exited1 because the new test helper passed minimumReads:undefined with exactOptionalPropertyTypes. The helper now defaults to an empty policy object; production was unaffected. Original typecheck.log is preserved.
- The first retry script accidentally renamed the pnpm task to nonexistent typecheck-fixed while attempting to give the log a unique label. It exited1 without running TypeScript; typecheck-fixed.log and checks-resume.py are preserved. checks-corrected.py ran the actual typecheck command successfully before build/tests. No failed result was called passing.
- An intermediate progress message misstated the new-case count as13. Actual structured output is11; the full focused count183 was always correct. Root was notified and all final reports use11.

## Durability, cleanup and remaining gates

The ordinary commit includes the Agent-Signature attribution trailer. Guarded push rechecked the remote Vickie session and published only the new branch. Independent ls-remote verifies exact2bd4843; remote-verification.json records it. No PR/main change or readiness action was performed.

Emitted process87053 exited0 and is absent; its store was removed after service/owner close. Path-bound OS inspection found no remaining Node test/probe processes for the author/control/report paths. Author and restored control checkouts are clean. Source.json records exact source/tree/parent and three file hashes; summary.json combines observed results, failures, sensitivity evidence, approved measurement clarification and cleanup. No shared DB or unrelated process was touched.

Source remains frozen for independent code council. Root owns code acceptance, CI/PR publication, ordinary integration and any future activation. This work completes only the accepted internal service increment; complete Transactional discovery and qualification remain separate.
