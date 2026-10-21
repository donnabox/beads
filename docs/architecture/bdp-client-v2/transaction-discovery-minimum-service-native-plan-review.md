# Native plan review — private discovery minimum service preparation

**No actionable findings.** The proposed three-file slice is feasible under the current service/owner/consistency contract, with the stated internal-only and evidence boundaries. Recommend implementation within the written scope; this review is not activation or implementation approval by itself.

## Reviewed basis

Plan SHA256 `7b8132c751647ff49fa06e69b31cd98c598bbe809f0d54fa9a7e5930e9130d40` reproduced. Read-only inspection used clean frozen `/Users/dbox/repos/wt/vickie/vickie-discovery-limits-integration`, HEAD `9b95df5dd866dbaf84150404c4db86ac8054b724`. Actual transaction-service.ts SHA256 is the plan's `02b5278ab88a476bd98e0ae40d1f6cbfe7ed3f03208a59f57d133d0f75e341e3`. The separately frozen compatible-missing checkout was not modified.

## Correctness and lifecycle assessment

- **One registered operation:** `transaction-service.ts:125–139` registers before starting work and retains the promise until fulfillment/rejection. Wrapping both finite wait and final assembly in this one run ensures `close():528–548` drains them before closing Read state and the owner. It is important that assembly remain inside that registered promise, as the plan explicitly requires. Closing rejects new valid calls while admitted calls may finish.
- **Same-entry observation:** current accessor `:237–252` captures three immutable fields through withRead and then assembles after release. `transaction-store.ts:883–958` gives exactly the intended synchronous entry/cause boundary, rejects async callback results and expires its reader in finally. Putting the existing comparison inside that callback gives a coherent eligible prefix without a second head lookup. Returning the captured prefix after a later independent commit remains internally coherent; requiring it to equal a newer observation would be an incorrect oracle.
- **Minimum and failure handling:** `transaction-consistency.ts:19–42` supplies detached canonical tuple validation; `:66–78` supplies exact epoch/view/position comparisons. `:85–121` creates one deadline and retries only actual behind failures, preserving foreign/context/provider errors. No extra parser, retry loop or error translation is needed. Existing entry-refused wrapping/cause must stay unchanged. Parser/assembly failures after the successful wait are nonretrying by construction.
- **Compatibility:** the no-minimum branch can preserve its synchronous capture/assembly-before-return behavior inside run's async callback, without consulting minimumReads or the newly accepted signal. The plan explicitly discloses that limitation. Existing zero-argument callers, captured limits and pagination callback receiver semantics remain unchanged. The minimum branch's synchronous input validation mirrors fresh Snapshot handling at `transaction-service.ts:305–325`; the plan correctly avoids inventing malformed-input versus closing precedence.
- **Cancellation:** the helper checks abort before an attempt, after a held wait resumes on the next loop iteration. It does not interrupt synchronous work or add a post-success abort gate. Its clock is initialized before the first abort check; the plan requires zero entries/parses/waits for pre-abort, not zero clock reads, so its oracle is accurate. Close draining rather than separately cancelling admitted waits is consistent with the existing service.

## Test feasibility and discriminating value

The real fixture and source-private seams already support the proposed cases without extra production files. Existing service tests at approximately `:709–750` use a paused injected wait, independent service.submit, mutable caller controls and a concurrent satisfied read. `:811–866` demonstrates pending work plus close with release in finally. Discovery tests at `:1085–1147` already count withRead entries after construction, prove parse-after-release using a reentrancy-protected real owner API, and distinguish entry refusal from parser failure. Their preservation is correctly required.

Equal/earlier, foreign, timeout, pre-abort/pending abort and close counts can be deterministic using the existing fake monotonic clock and deferred wait. The plan confines delegated fenced/provider failures to labelled test faults instead of claiming disk recovery. Its three proposed source mutants distinguish meaningful obligations: missing comparison, mutable retained controls and parsing inside the guarded entry. No extra generic fault framework is necessary. Retain the stated short mutant bound and named assertion evidence; a stalled mutant is not interchangeable with the expected assertion failure.

The source test requirements do not depend on the concurrently implemented missing-Resource behavior: this accessor observes Scope context and assembles discovery, without Resource lookups or visibility translation. Shared service test additions still need ordinary ancestry preservation at integration, as the plan states.

## Evidence boundary and scope

`transaction-runtime.ts:350–378` returns scope/listenUrl/configurationMetadata/targets/close, not its service object. `transaction-http.ts:525–546` still rejects minimum-bearing discovery through the subset gate and ordinary discovery through the explicit discovery gap. The current emitted installed-contract test imports dist runtime and checks the actual discovery refusal before and after mutations (`transaction-installed-contracts.test.ts:404–418,489`). It is useful regression coverage, not new-accessor coverage.

The proposed direct emitted script is therefore an honest additional execution layer: actual dist service/store modules, a disposable owner, matching literal captured configuration, four bounded cases, one process and no listener. The existing build layout supports those private emitted files; no public export or IPC seam is needed. Its hashes/source binding and cleanup must be retained. Not requiring a new installed HTTP test for an intentionally unreachable private accessor is appropriate. Future serving/advertisement remains a separate review and wire-proof gate.

Exactly the three named files suffice: transaction-service.ts, its test, and progress documentation. The existing fixed-policy/no-principal accessor remains internal; these tests do not establish caller authorization or any public cold-client behavior. No protocol choice, store change, schema change, public profile claim or whole-contract qualification follows from this preparation.

## Review activity

No tests, builds, servers, network requests, repository edits or outward actions were performed. Only this external review report was written. The frozen source remained clean.
