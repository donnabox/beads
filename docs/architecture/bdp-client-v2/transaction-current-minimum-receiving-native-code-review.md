# Current minimum receiving — native code review

No actionable finding in the complete four-file change at **5d854a1da21a377f03c2d4cd523b10f1e7659cea**, tree **75a9db868de253a92fa4e08495d86ec0f6ed9ac6**, direct parent **7171251d282769c1b9e2a39aed63e16f8b24ed96**. Source implements the accepted v3 receiving correction without expanding protocol or request scope.

## Exact scope and source reasoning

Read the complete diff in frozen `/Users/dbox/repos/wt/vickie/vickie-current-minimum-receiving`, accepted plan-v3d04250fa and dispositions845ce082, and surrounding authority option capture/delivery, minimum wait and service lifecycle. Working tree is clean; changes are exactly transaction-read.ts/.test.ts and transaction-service.ts/.test.ts. Read-only inspection and external report writing only; no test/build/server/network/source/outward operation was performed.

The helper rename from captureTransactionAliasDeliveryOptions to captureTransactionReadDeliveryOptions changes no capture behavior. It still rejects proxies before reflection, reads own data descriptors without invoking accessors, retains non-enumerable signal by value, rejects unsupported fields/prototypes/signal shapes and freezes detached options. The same authority predicate validates signals. No public root/development export or package entry is added.

Minimum perform preserves the supported-request predicate first, then captures options. On success it injects the existing immutable minimum into detached data. On capture failure it delegates the original invalid options through authority's observed Promise. Every helper rejection class is also an authority receiving rejection, including non-enumerable unknown fields (authority republishes own data descriptors as enumerable before its key check). That fallback cannot reach an unconditioned successful observation. Closed/reentrant authority guards retain precedence over receiving validation for supported literal requests. Ordinary facets are untouched.

Service perform captures inside unchanged run admission, so closed-service calls retain their synchronous admission throw. It uses captured options for every attempt and captured signal for the one finite wait lifetime. Failed capture delegates to the rejecting facet without dereferencing caller signal. No new wait, renewed deadline or owner-held asynchronous operation is introduced. Every retry invokes the normal facet again; the inherited source condition checks the actual same-entry context before release. Caller condition cannot lower the configured minimum.

Alias branches differ only in the private helper identifier/type-reference rename; receiving behavior remains the independently reviewed alias repair. No HTTP routing, request predicate, schema, fixture, catalog, configuration, public capability or other bare-signal entrypoint changes occur.

## Meaningful test inspection

The new tests use valid supported Resource literals and exercise null/scalar/array/nonplain/proxy/revoked/accessor/symbol/unknown input, including named non-enumerable unknown-key and invalid-signal cases. Both direct/service modes assert observed rejection, zero owner reads and no option getter/trap invocation. Non-enumerable pre-aborted signal refuses before entry; null-prototype non-enumerable valid signal succeeds. Opaque caller sourceCondition is overridden, lower caller bound cannot replace future configured minimum, and nested/closed cases distinguish direct Promise from synchronous service admission.

The held-wait tests use a controlled clock and one held wait. The positive comparator keeps the original signal active, replaces caller signal with pre-aborted signal, commits a real Resource effect while the wait is held, then succeeds at pos-2 with exactly two owner entries. That tests per-attempt capture; old code would reread the aborted replacement. The negative comparator aborts the original, observes close/drain waiting, releases the wait and requires no second owner read plus closed owner. It is correctly cancellation coverage rather than the sole capture discriminator. Successful independent submission during the held wait substantiates that no Read owner entry spans the await. Cleanup releases the held wait before service teardown.

## Completed author evidence inspected; not independently rerun

At review time completed result JSON/logs and summary/source/cleanup records were available; the final prose author report had not yet been finalized. This is a source-code review with the following inspected executions, not a claim of independently executed qualification.

- Final receiving6pass/117filtered; focused187pass; final typecheck/build/format/lint/boundaries/Read seal/diff checks exit0.
- Exact disposable facet counterfactual fails the direct non-enumerable pre-aborted-signal test because it unexpectedly resolves successfully. Exact service counterfactual fails both unchecked-getter and original-active/replacement-preaborted successful-retry tests (the latter aborts). Their restored candidate tests pass. Counterfactuals are separate source derivatives; their exit1 records are expected, not target-run successes.
- Fresh installed minimum3pass/onefiltered: separate collection38/1, Type/incident20/1 and alias18/1 scenarios, each bound to5d854 with clean source and closed child. Alias regression is appropriate because the shared private helper was renamed. Compatible-missing was intentionally unselected, not reported as freshly executed.
- Two targeted installed Update positives pass/200filtered. These are ordinary-wire regressions, not hostile-JavaScript receiving evidence.

Independently rehashed all four source files and both actual tarballs against their source/summary values; all match. Fresh served payload **535612f787a2c3e855cab57801599c86faba1ed78e5952534d70c036e64f442a** is source-bound. Cleanup records all three minimum child PIDs absent and source clean. No reviewer-owned child processes were started.

## Limits

Source inspection found no actionable defect within the accepted four-file scope. Completion of the final author report, external council, exact integration/CI and root readiness decisions remain separate gates. This review does not claim deferred request-accessor handling, other bare-signal APIs, hidden policy, full receiving safety across the product or public Transactional/discovery readiness. Existing cooperative cancellation limitations remain unchanged.
