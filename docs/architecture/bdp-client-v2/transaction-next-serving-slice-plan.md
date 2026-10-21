# Next serving slice: preserve receiving on current minimum Reads

**Recommendation, not implementation approval:** after the alias receiving repair clears council, reuse its descriptor-safe capture for the existing minimum `perform` path. This is a concrete serving implementation correction, not another diagnostic row. It does not enable a new endpoint or clear the discovery advertisement gate.

Supporting plan under the sole fork `design.md` (“Autonomous profile completion” and “Current priority”). Read-only inspection: main0432d48db8f340016a000d633d799aebbd26db3f, via integration8d23862 whose complete server/protocol source equals that main; frozen alias repair3959a25e70bc6dc9d143abfead187ccaf84ca11e read only for its existing helper and ownership. No tests/builds/network/source changes. The alias repair is not presumed reviewed or merged.

## Why this slice, and what remains blocked

The accepted discovery-serving v2 plan's current Type/incident/missing-target and internal discovery-minimum prerequisites are now on main; alias minimum receiving is actively repaired. The remaining refused wire cases identified there still involve explicit receiving/applicability, retained-anchor, missing-target/foreign-minimum, Selector admission or existing Event decisions. Discovery wire activation remains held because it would advertise the cumulative profile. This audit found no basis to turn one of those refusals into new law. It is not an exhaustive proof that all remaining required behavior is complete.

There is a separate, demonstrable receiving defect left on both main and the frozen alias repair:

- Main `packages/server/src/transaction-read.ts:206–213` (`3959:225–232`) spreads caller options into the minimum `perform` call. A valid own **non-enumerable** pre-aborted signal is dropped, while an enumerable signal getter or Proxy reflection can execute before the authority receives the input. Null/array option shapes can be laundered by spread.
- Main `transaction-service.ts:305–311` (`3959:309–315`) reads `receiving?.signal` before descriptor-safe validation. That can invoke a caller getter while selecting the finite-wait signal.
- The existing authority contract explicitly inspects plain own data descriptors, rejects Proxy/accessor input without invoking it, retains non-enumerable own values, and delivers early failures through its observed Promise after lifecycle checks (`authority-read.ts:296–313,1026–1078`). The native alias review documents the same distinction and explicitly identifies the inherited `perform` analogue.
- Actual HTTP `transaction-http.ts:214–219` supplies a normal `{signal: request.signal}` object. Thus this is a private serving-composition boundary defect, **not evidence of a remote exploit or failure of the existing installed wire observations**. Do not pretend a new HTTP query can express JavaScript property descriptors.

## Authority and boundaries

Normative wire law remains unchanged: `docs/specs/bdp.md:5758–5764` requires a transaction-consistent current prefix and explicit treatment when behind; `:5776–5786` defines the complete minimum and eligible/wait/failure alternatives; `:5791–5798` requires truthful read context; `:5805–5807` requires consistency checks before conditional shortcuts. These clauses justify retaining the current owner/minimum mechanism; they do **not** prescribe JavaScript options syntax or a new cancellation Problem. Descriptor safety and Promise/lifecycle precedence come from the existing authority API above, not invented protocol text.

Receiving: valid request stays within the current Resource/properties/initial collection/Type/incident predicate. Capture only existing delivery options; validate signal as the existing authority does. Preserve own non-enumerable signal and ignore/replace only the caller's opaque sourceCondition with the already-captured required minimum. Invalid options must go through the authority rejection path, never an ordinary success fallback or a raw getter read.

Consistency: no change to `snapshotTransactionMinimum`, owner-entry observation, visibility gaps, finite policy/deadline, conditional ordering or source context. Use the captured immutable options for **every** attempt and the captured signal for the wait loop. Abort during a held behind wait must prevent the next owner attempt once that wait is released; do not claim interruption of synchronous work or a custom wait Promise that has not settled (`transaction-consistency.ts:101–118`).

Response: preserve the authority's observed Promise, invalid-input/closed/reentrant precedence and `ScopeServerOperationAbortedError`. No new status/code, Problem schema, error sanitization, public export, listener configuration or advertisement. Valid wire response bytes remain the same.

## Exact ownership and dependency

Four proposed source/test files only:

1. `packages/server/src/transaction-read.ts`: apply the alias repair's existing capture/fallback pattern to minimum `perform` before injecting its own immutable minimum. Reuse rather than duplicate `captureTransactionAliasDeliveryOptions` (`3959:97–114`). A narrow rename to `captureTransactionReadDeliveryOptions`, including its existing alias call sites, is reasonable only after the reviewed repair is inherited; its capture behavior must remain unchanged. Do not edit authority-read.ts or broaden request classification.
2. `packages/server/src/transaction-service.ts`: mirror the repaired alias branch (`3959:317–331`) for `perform`, selecting wait signal only from validated capture. Invalid receiving options delegate once to the minimum facet so the authority chooses the receiving/lifecycle failure. Do not change alias semantics, receipt waits, Snapshot, Events or discovery.
3. `packages/server/src/transaction-read.test.ts`: direct facet receiving discriminators and valid-current read regressions.
4. `packages/server/src/transaction-service.test.ts`: capture/wait/lifetime discriminators through real owner and service.

These files overlap the active alias lane; **do not dispatch simultaneous writers or recreate its helper**. Wait for root's reviewed frozen alias successor and ordinary integrated base. Then this slice owns only the `perform` branches plus necessary helper-name references, leaving the reviewed alias branches behaviorally intact. It is disjoint from PR110's conformance artifacts and all Type lifecycle, runtime configuration, HTTP routing and public admission work. If a new file/export/authority change is needed, return the concrete gap to root before expanding scope.

## Discriminating tests and bounded execution

Use real disposable owner fixtures and existing counter/wait seams; no new production test seam.

- Direct facet: current eligible Resource request plus non-enumerable own pre-aborted signal must return the observed rejected Promise and enter no owner read. An ordinary active signal remains accepted. Also exercise a null-prototype data object.
- Direct and service: enumerable/non-enumerable signal getter, Proxy options with observable traps, null/array/non-plain/unknown/symbol options and invalid signal shapes must not be laundered. Assert no getter/trap invocation where authority currently promises it, no owner entry, Promise rejection and the existing named failure. Use valid literal requests to isolate delivery receiving; malformed request/predicate redesign is outside this slice.
- Lifecycle: invoke a previously obtained facet after close, and during a controlled reentrant owner callback, using invalid receiving options. Require the authority's existing closed/reentrant precedence rather than a capture exception or accidental successful ordinary read.
- Service capture: start a real behind minimum with a held finite wait; mutate caller options to a different signal/sourceCondition after admission. Abort the originally captured signal, release the held wait and require no second owner attempt. An independent valid commit can satisfy an otherwise identical non-aborted call. Close/drain remains bounded and releases the owner.
- Preserve minimum enforcement with caller `sourceCondition` supplied, including non-enumerable data: the caller cannot replace the configured minimum. Existing compatible404, Type locality, initial incident parsing, foreign/behind and conditional tests continue passing.
- Two disposable source counterfactuals: restore only the old facet spread (non-enumerable/getter control must fail), then restore only unchecked service signal selection (service getter/capture control must fail). Restore candidate bytes and prove clean source afterward. No installed mutant needed for descriptor-only inputs.

Validation after review-authorized implementation: pinned clean typecheck/build; focused read/service tests plus existing HTTP current-minimum/disconnect tests; format/lint/boundaries/Read seal. Fresh packed current-collection38/1 and Type/incident20/1 scenarios, plus two targeted RU positives, use unchanged existing fixtures/deadlines. They are combined-runtime regression evidence, not proof of hostile JavaScript receiving input. The alias18/1 regression may be added only as a separately counted existing scenario if the helper is renamed; no new request/control or broad full suite. Keep each fixture's actual timeout rather than calling one aggregate budget universal.

## Machine-checkable done when

1. Changed paths are exactly the four authorized paths; reviewed alias helper behavior, all HTTP/artifact/schema/config bytes and current route predicate are unchanged except an approved private helper rename.
2. Every new receiving/lifecycle control passes against the candidate; each bounded old-code counterfactual fails its intended test. No getters/traps or owner calls occur in cases where the current authority forbids them.
3. Held-wait counters prove one original finite lifetime, captured original signal, no post-abort owner attempt and successful cleanup; the real commit comparator proves the eligible path remains usable.
4. All named checks and existing packed regressions pass with exact clean source/tree, actual new package/config/payload identities, per-scenario counts, unchanged limits and no owned children. No historical evidence is relabeled or whole-profile claim added.
5. Native plus Claude review and exact-head CI close before root integration/readiness. This plan itself grants no authoring authority.

No unavoidable product decision is needed for this correction. It improves the supported serving composition while the remaining wire decisions stay with root/operator; it must not be described as completing the cold-client milestone by itself.
