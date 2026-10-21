# PR108 repair native closure

No actionable findings in the bounded repair. The accepted delivery-validation regression, durable evidence reference, plain missing GET/HEAD witnesses, and same-owner-entry assertion are closed at the reviewed source. This is a source/evidence review, not profile qualification or a merge decision.

## Exact source and scope

- Checkout: `/Users/dbox/repos/wt/vickie/vickie-missing-minimum-review-repair`.
- Head: `dfafd82290fb6f06db1e01522a3684a21d9813a7`.
- Direct parent: `4697a737b1446021d920f201552cb27e5c207cc5`.
- Tree: `cfd5313612249cb48211cbea7d3fe46e2edb5daf`.
- Working tree clean when inspected. Entire four-file delta inspected: 187 additions / 17 removals in transaction-read.ts, transaction-read.test.ts, transaction-http.test.ts, and the progress document. No changes to authority-read.ts, minimumFor, HTTP production code, schemas, fixtures, or installed test source.
- Read the root code dispositions, repair report and source-bound summary; compared the helper against the actual authority data capture, delivery, lifecycle, and error handling.

## Closure reasoning

1. **Descriptor-safe narrowing and original validation.** `transaction-read.ts:40–70` rejects proxy/revoked-proxy containers before reflective access, requires the same plain-object/prototype/data-descriptor shape as `authority-read.ts:297–314`, inspects nonenumerable own fields, and rejects symbols/accessors/unknown fields. It retains the own signal value without truthiness coercion; null, false, zero and empty strings consequently reach the original AbortSignal validator and fail. An own data sourceCondition is intentionally omitted without traversing its value. In particular, a revoked proxy held as that ignored value is inert data. No accepted ordinary call can supply a source condition to the owner callback.

2. **Promise and lifecycle precedence.** `transaction-read.ts:145–172` routes helper rejection back through the original authority facet with the original invalid options. The authority creates and observes its delivery Promise before checking state (`authority-read.ts:1025–1053`); its guard checks reentrancy and then closed state (`965–968`), before receiving validation. Invalid options therefore retain the original lifecycle precedence and asynchronous rejection channel. The helper performs no caller getter/proxy trap, so its preliminary inspection cannot reenter or alter lifecycle state. The original facet catches receiving/work faults inside deliver, so normal rejected deliveries do not enter the wrapper's synchronous fallback or cause a second owner read. Original failure-before-abort precedence, listener removal and busy reset remain in the untouched authority implementation (`1093–1124`). Source reasoning establishes the reentrant path; the added test directly demonstrates the closed path, without claiming a new reentrant runtime probe.

3. **Fallback cannot leak an ordinary condition into execution.** A helper-rejected container also fails the authority's same receiving boundary before observation (or fails an earlier lifecycle guard). A valid container always takes the stripped signal-only path. The explicit minimumFor path and its captured tuple remain byte-unchanged by this repair; no expansion of ordinary minimum authority is introduced.

4. **Meaningful regression assertions.** Both ordinary perform and resolveAlias exercise malformed shape/data, proxy/accessor nonexecution, invalid falsy signals, valid null-prototype/nonenumerable data, ignored opaque condition, Promise delivery, and closed-state precedence (`transaction-read.test.ts:1041` onward). The compatible-missing matrix now counts precisely one actual owner entry for each call as well as retaining the existing lookup guards/context assertions (`868–925`). The new plain GET/HEAD pair has no incidental conditional headers: it checks resource-not-found, all three context headers, no-store/no ETag, Problem media, and HEAD body/length behavior (`transaction-http.test.ts:1254–1274`). The progress evidence reference is now a commit-pinned repository document; this review did not perform a network retrieval of it.

5. **Bounded disposition preserved.** No revision regex, new missing-versus-foreign ordering law, new HTTP context fault mechanism, or public profile claim was added. The previously declared ambiguous missing/consistency cases remain gaps. Existing broader minimumFor receiving behavior is not redesigned by this ordinary-facet repair.

## Evidence inspected, not rerun

Independent read-only commands: git status/rev-parse, full git diff, rg/sed/nl source inspection, and Python SHA-256/JSON metadata inspection. All completed successfully. No tests, builds, package installs, servers, credentials, network operations, or source writes were run for this review. Only this review document was written outside source checkouts.

All four source-file hashes in summary.json match actual frozen source bytes. The author records final 258 tests in eight files, typecheck/build/boundaries/format/lint/Read evidence verification successful, followed by a frozen build and fresh package tests. Earlier two typecheck failures remain visible and are followed by corrected final checks; they are not erased or treated as passes.

The actual packaged summary binds both observations to dfafd822, clean source, and separate closed exit-zero children: existing collection minimum 38 exchanges / one child (86823), compatible missing minimum 10 exchanges / one child (86834). Both are passed and claimEligible remains false. The two targeted Update scenarios also have an exit-zero packed run. Cleanup records both children absent and no remaining installed children. This reviewer launched no processes requiring cleanup. Prior source mutants are predecessor evidence, not newly executed repair controls.

Inspected artifact hashes:

- report.md: `6f99bca395f19371cc3510b8b6801aa6e8fd23450446c2a3425649c99bed8da3`.
- source.json: `11abe328c35813a1fac8c3c99b5e3f3e474afec41087d5e76c9819b91185d815`.
- summary.json: `c232fc8894d307cb717fcbf0dad2a578009e09bbe02b02db41bc66a552f99c62`.
- cleanup.json: `66bc1f23bbb0f0d41c2f6f39f22bbd63799c468746d3d990e15d5e13aa69fad8`.

Root retains responsibility for exact CI, publication and final integration gates. This review closes the specified repair only, without duplicating the original whole-branch council or installed qualification.
