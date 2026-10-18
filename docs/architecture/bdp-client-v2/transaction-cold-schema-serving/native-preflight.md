# Native static review — local schema measured preflight

**No Medium-or-higher blocker found in this bounded preflight. The proposed execution is sufficiently enumerated for root's release decision, subject to the Low cleanup item and the final installed postprocessing conditions below.** This is not code-council approval, a passing test result, parent-CI approval or public qualification.

## Basis and independent verification

Read-only inspection of `/Users/dbox/repos/wt/vickie/vickie-local-schema-serving`: HEAD4aa2d0be950f157f3f0617278dd37c40f41ec035, tree5afa00950cb042b931450bb7545369c4274ef8ba, nine uncommitted authorized paths. Reviewed the full actual source/test/documentation diff, accepted plan c464ddba, dispositions72d3a92b, closurea356a0cd and root's four-output packaging amendment. No source mutation, tests, build, service, install, network or derivative execution occurred.

Independently recomputed all nine current and parent file hashes/lengths and regenerated the complete unified diff in memory; it exactly equals saved source.diff. Confirmed1073 additions/5 deletions. Applied each of the five unified control patches only to strings in memory and verified each supplied mutant hash. Recomputed all14 request body hashes/lengths, nine-route preimage/digest, all three captured artifact hashes/byte/base64 lengths and memory arithmetic.

Reviewed artifact identities:

- preflight.md:9c92b1ba854586937308834d3b03ee6c6c9ec7fe34cdbf4e93fbcf1a6c5e32ad
- preflight-bounds.json:151e5b13108a46647686cc9cebb7b9a7db61c350c59f2d7c4edd5d98c993194a
- source.diff:dabf5835eb91610ee3c8c32411e56f5e98c0a7a0648cb2fb534e015b14a56d84

Compilation remains unexecuted. The service spies retain explicit overload casts, and the installed harness imports emitted packages, making the declared typecheck/build order necessary. The heterogeneous authentication-header fixture at adapter-test247–256 is a typing spot to check under strict/exactOptionalPropertyTypes; a typed Record<string,string>[] fixture is the narrow repair if the compiler rejects its inferred optional keys. This is a prospective compilation check, not an observed compiler failure or an additional wire case.

The nine files match the accepted eight plus scripts/smoke-executables.mjs solely adding four private output filenames. HTTP changes only the export keyword on hasMinimumControls; the existing implementation and handler are unchanged. Service adds only schemaObserved; existing methods and the complete old service-test prefix are preserved. No package public export, configuration, shared Read/store/compiler/catalog or launcher modification is present. The exact-output inventory amendment is necessary packaging work; no local smoke workload is silently included.

## Low finding: source cleanup can skip the directory and mock restoration

`packages/server/src/transaction-schema-http.test.ts:27–30` removes the entire cleanup list, then awaits each callback without collecting failures. Runtime close is registered after directory removal, so reverse iteration closes first. If close rejects, directory removal is skipped, the queued callbacks are lost, and vi.restoreAllMocks is skipped. This is a failure-path test-lifecycle weakness, not a production schema defect or evidence that the expected controls will leak.

Small correction: in the new test's afterEach, attempt every registered cleanup callback while collecting errors, restore mocks in a finally block, then surface the aggregate cleanup failure. Keep the test/assertion failure visible in Vitest. No HTTP/listener/callback count needs to increase. The service file's similar historical cleanup convention is not a request for unrelated refactoring. Root may disposition this Low before execution; if changed, regenerate the exact file/diff/control/preflight hashes before freezing.

## Production preflight checks

The adapter's closed lookup comes from genuine immutable DevelopmentContracts capture. Exact schema retrieval URIs must match the canonical local schemas/ namespace and exclude query/fragment delimiters and userinfo. No canonical-ID alias, external resolver, additional startup policy or Resource shadowing is introduced. Compatibility cases retain valid and invalid offline schema evaluation under differing$id, outside-namespace and query-bearing capture identities.

The runtime invokes the adapter only after existing raw-authority/method/body/closing/Origin gates. The adapter requires authenticated GET/HEAD, no minimum controls via the exact existing helper, and an exact manifest hit; excluded requests have no owner observation and retain ordinary handler behavior. The shared HTTP selector needs representation metadata, not a parsed schema object, so the metadata-only candidate preserves negotiation and conditional decisions without another schema copy. One schemaObserved call uses existing run/withRead, returns an immutable actual tuple and releases the entry synchronously. Selected200 alone supplies raw text to existing deliver; its destroyed guard, marker and HEAD strip remain shared. No stale tuple is substituted for an observation fault.

The added await on adapter selection also occurs for exclusions, but runtime close drains pending handlers before service.close; this inspection found no resulting closed-owner race requiring a change. Final lifecycle review remains part of code council after execution.

## Source budgets and controls

The eight adapter cases expand as1 unit +1 selection +1 gate +3 compatibility +1 disconnect +1 fault. The selected service regex matches the two new cases, not the63 historical cases. Source total10 cases; seven runtime listeners/stores plus two service owners. HTTP counts are8+16+5+3+3+2+1=38. Gate count16 is11 minimum rows,3 unauthenticated selector/conditional rows,1 Origin and1 method-precedence request. The aborted disconnect request is counted. No subprocess, implicit fetch or installed cohort is created by these source cases. The separate installed confinement test has no HTTP.

Callback accounting matches the preflight distinction:10 actual successful schema callbacks across runtime fixtures, one injected throwing callback, two successful unit callbacks; service cases have two successful schema entries, one comparison discovery entry, two refused schema attempts, one outer reentrant callback and no entry after service closure. Three service seed/advance submissions are direct, not hidden wire requests. Twelve explicit contract captures are fixture setup, not network reads. Build-before-emitted-consumers is explicit.

All five control selectors reach their intended assertions without an earlier unrelated assertion:

| Control | First intended discriminator | Expected wire attempts |
|---|---|---:|
|remove-handling|known root status on404|1|
|reserialize|exact captured schema bytes in unit case|0|
|omit-transitive|referenced document status after root passes|0|
|stale-context|current schema position after real commit|8|
|minimum-bypass|preserved minimum refusal before unit exclusion assertion|1|

Thus expected derivative10 HTTP/three listeners is correct. Hard fallback32 HTTP/three listeners is8+8+16 if the selected runtime fixtures unexpectedly complete; the two unit controls contribute no wire calls and at most four supplied unit callbacks. These are prospective bounds, not successful sensitivities. Compilation/setup/timeout failures must remain failures, not accepted negative controls.

## Installed bounds, independent oracles and lifecycle

The actual fixture implements14 calls with nine distinct method/URL identities. Its first two Type reads validate the unchanged normative schemas and exact independent descriptor/inventory; observed Type/root/ref grants must equal the separately configured identities. Root$id and actual relative$ref resolution are explicitly checked before granting the reference GET. The control route set does not grant arbitrary returned URLs. Conditional, HEAD, auth, method, unknown and withheld-discovery cases precede one real schema-valid write and a final unchanged-byte/current-position read. No receipt fetch, pagination, reopen or external resolver occurs.

Only request13 has a body:98 UTF8 bytes, aggregate98. Captures are164+174+118=456 bytes, with612 retained base64 characters. The adapter adds two immutable strings:292 UTF8-equivalent bytes,582 UTF16 content bytes, one transient decode buffer at most174 bytes for this fixture. It adds no parsed schema copy. Existing compiler objects, string overhead, concurrent encoding and socket buffers remain excluded from that retention estimate as stated; no RSS claim follows.

The predecessor should reach actual200/200/404, with the third response recorded before the named `known installed schema status` assertion fails. That status assertion accepts a numeric404 and does not have the prior undefined-number matcher defect. Positive14/one child plus negative3/one gives17 expected requests/two children, hard28/two. The14-call guard is explicit. Five-second exchange,90-second cohort,120-second outer,10-second launcher close and declared byte caps are unchanged.

Installed cleanup collects close/removal failures, clears the cohort timer, records exit/signal and directory absence. The existing launcher independently requires exit0/no signal and bounds cleanup. **The test itself does not probe PID absence**; preflight's promised independent PID-absence check therefore belongs in the final source-bound postprocessor/root cleanup record, alongside named negative failure and actual200/200/404. Do not report that check as already implemented or passed solely from session.exit. Final postprocessing is also required to bind actual archives/configuration/source, count failed startup children and verify no replacement run; those scripts/results do not yet exist in this draft. Their absence is consistent with staged execution, not evidence of completion.

Root can release only the enumerated source/static/control workload and separately frozen installed commands within these bounds, after dispositioning the Low. Root's subsequently read transaction-sequence-refusal-ci-retry-disposition.md explicitly permits this bounded validation without waiting for the parent merge; it preserves the separate exact-head PR118 CI/merge gate. No smoke/broad workspace matrix, hidden install cohort or retry is included. All eventual evidence must retain the actual frozen successor and separately packed4aa predecessor identities; source drafting on4aa does not imply its CI or merge succeeded.
