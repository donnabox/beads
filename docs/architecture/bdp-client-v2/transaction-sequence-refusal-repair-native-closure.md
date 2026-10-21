# Native closure: PR118 sequence-refusal review repair

## Findings and conclusion

**No new actionable findings. All four accepted corrections are closed by the bounded test/documentation successor.** Production semantics and installed scenario bytes are unchanged. This is correction closure, not fresh installed proof, successor CI approval, merge authorization or profile qualification.

## Exact basis

Read-only checkout `/Users/dbox/repos/wt/vickie/vickie-sequence-refusal-review-repair`:

- HEAD `4aa2d0be950f157f3f0617278dd37c40f41ec035`.
- Tree `5afa00950cb042b931450bb7545369c4274ef8ba`.
- Direct ordinary parent `90cc834c3672b02840c68dcd7abb35794f038887`.
- Clean working tree observed. Complete diff: two files, 76 insertions / 35 deletions; SHA256 `4bc38b495ffb7aab7f949756f98c07996334eba6cb14371b847e3107853d0f47`.

Read the root code dispositions and full external review SHA256 `9095995242a1a534e5d342fd912c11c5d19cca4904238d5a498846acded19fa8`, complete parent-to-successor diff, surrounding production routing and test implementations, and the repair report/source/summary/preflight plus actual focused and typing logs. No tests, builds, installs, services, network actions or product changes were performed by this reviewer. The only written artifact is this report.

## Correction dispositions

1. **Scope documentation — closed.** `docs/design/transaction-implementation-progress.md:1721–1731` now separates ordering preservation from tuple attachment and limits attachment to operation-count 413 and supported-carrier preadmission 406. It explicitly lists carrier-key/malformed 400, 401, 403, 415 and composition/minimum development 500 absences as retained behavior and remaining gaps. It expressly denies normative exemptions or a ruling that unauthenticated errors must disclose context. No new wire rule or production behavior is inferred.

2. **Preparation before negotiation — closed.** `packages/server/src/transaction-service.test.ts:1968–2015` constructs a depth-40 properties value with valid sequence shape and unacceptable `Accept: text/plain`, then uses the real HTTP handler. It requires the unsupported-transaction-composition development marker and bodyless 500, no content type, Content-Length 0, all three context headers absent and zero refusal-observation calls. The table leaves owner position at zero. Actual production `transaction-http.ts:323–334` prepares before negotiation: moving negotiation first would produce 406 and call the observation, failing these checks. This is a meaningful ordering discriminator within the existing seven-row table, not another installed scenario.

3. **Both observation-fault sites — closed.** `transaction-service.test.ts:1914–1957` parameterizes the existing failure case over 33 and one valid members, with unacceptable Accept in both. The 33-member carrier reaches the over-count observation at `transaction-http.ts:316–322`; the one-member carrier completes preparation and reaches the negotiation observation at332–334. Closing the real owner makes both fail through the existing sanitized handler path. Both require status 500, absent body/content type/all three context headers, zero content length and the non-attesting marker. Original reentrant and failed-owner cause assertions remain. A fallback to contextless 413 or 406 at either site would fail status assertions. These are source-handler observations; no claim of a newly executed installed owner fault follows.

4. **Execution spies — closed.** `transaction-service.test.ts:1859–1902` adds `vi.fn(f.owner.execute)` and `vi.fn(f.owner.executeObserved)`, installs both delegates into the service owner, clears their history after the two seed submissions, and asserts neither was called during the refusal observation. They are delegates to the actual methods, not inert substitutes. The overloaded execute cast changes TypeScript's view of the spy only; it does not replace implementation or argument forwarding. Existing single synchronous entry, unchanged store, immutable tuple and other no-admission/no-sequence assertions remain. The later advance still verifies current context progression separately.

## Independent preservation checks

Compared full parent and successor Git trees, rather than relying only on the author's summary: both have 660 tracked files; exactly the owned service test and progress document differ. The other **658 files are byte-identical**. Independently recomputed both changed-file hashes and every one of the **57 installed-named source/support hashes** in `source.json`; all agree and match the parent. In particular, transactional runtime, service, HTTP, store, protocol/configuration and installed refusal scenarios did not change.

Changed hashes:

- Service test: `fb189cce0c1680fcf7b8d9e82fed90acd410994f735c60fc044bf4dd27cadec8`.
- Progress document: `e8558dde9d8ad83585ab4dbb9b74e43d93bb3b97e5246d537d02cc2e73ff1c51`.

There is no separate preservation.json; the independently verified preservation inventory is in source.json and the two-file Git delta.

## Recorded execution checked, not rerun

`focused.json` reports exactly **63 total / 63 passed / 0 failed / 0 pending**, one file at the assigned repair checkout. All four affected test results appear as passed: the no-execution observation, 33-member fault, one-member fault, and exclusion table. `focused.log` agrees (Vitest duration 1.62 seconds); `focused.exit.json` records exit 0 and supervisor elapsed 3.014 seconds. The preflight expected 63 from the parent's 62 because one case became two. The command selects only `packages/server/src/transaction-service.test.ts`, maxWorkers 2, verbose plus JSON reporters.

The initial typecheck is honestly retained as exit 1: TS2345 reports Vitest's final-overload execute spy cannot satisfy the overloaded owner interface. The final source applies the same transparent explicit-cast idiom as the pre-existing admit delegate; corrected typecheck exits 0. Recorded build completes before the focused test run. Build, corrected typing, format, lint, boundaries and Read evidence verification each have exit 0. Lint success is not a claim of zero warnings. No failed typing attempt is reclassified as successful verification.

No installed cohort was rerun for this test/doc-only repair. The prior 373 focused source tests, five derivative selections, installed 13 + 14 + 3 exchanges across three children, and exact-parent CI remain historical **90cc834** evidence. Entry-file equality alone is not payload identity; this closure does not relabel those observations as execution at 4aa2d0. Root retains fresh successor CI, publication and merged-tree verification gates.
