# Current alias minimum — native code review

## Finding

### P2 — New minimum alias delivery bypasses the authority data-only options boundary

**Locations:** `packages/server/src/transaction-read.ts:215–220` and the related new service forwarding at `transaction-service.ts:313–319`.

The new alias facet builds `{ ...options, sourceCondition: minimum }` before entering the underlying authority facet. Object spread invokes enumerable getters/proxy reflection and copies only enumerable own properties. The underlying authority deliberately rejects proxies/accessors without invoking them, accepts own data descriptors regardless of enumerability, and returns early receiving failures through an observed Promise (`authority-read.ts:296–313,1026–1078`). The ordinary alias wrapper already preserves that boundary through `ordinaryDeliveryOptions` and its authority fallback.

Two concrete source-derived counterexamples, not executed during this read-only review:

1. With a live alias and satisfied minimum, pass `Object.defineProperty({}, "signal", { value: AbortSignal.abort() })` directly to `plane.minimumFor(...).resolveAlias(...)`. The own signal is non-enumerable by default. Authority `ownData` would retain it and reject before the owner entry; the new spread drops it and supplies only sourceCondition, allowing the read. This is a valid plain data options object, not merely a hostile invalid-input case. The service wrapper's separate signal read masks this particular loss on service calls; it does not fix the direct facet.
2. Pass an enumerable signal getter that increments a counter and throws a sentinel, or a Proxy with an ownKeys trap. The spread executes it before authority validation; a throwing getter escapes synchronously rather than producing the authority invalid-input Promise. On the service path, `receiving?.signal` itself also invokes the getter before the facet can validate it. Invalid `null`/empty-array values can likewise be laundered into an accepted sourceCondition-only object by spread.

**Impact:** the newly supported private minimum-alias API can ignore valid direct cancellation and execute caller code before its existing receiving/lifecycle boundary. The HTTP callback supplies a normal `{signal: Request.signal}` data object, so this is not evidence of a remote routing exploit or a failure of the measured installed18-request path. Minimum `perform` has an inherited analogous pattern; this finding is scoped to the new alias support and does not assert that the entire issue originated here.

**Minimal correction:** capture alias delivery options with descriptor-safe validation (proxy check before reflection, own data signal retained even when non-enumerable), then inject the immutable minimum. Preserve authority Promise delivery and closed/reentrant error precedence for invalid input; do not turn malformed input into an ordinary successful alias read. The service must not read an unchecked signal getter while selecting its wait signal. Add bounded direct/service controls for non-enumerable pre-aborted signal, accessor/proxy non-invocation and invalid-option Promise/error precedence, alongside valid signal and existing HTTP controls. Reuse the reviewed ordinary receiving logic where appropriate without widening alias into the canonical Resource predicate.

No other actionable finding identified in this bounded review. Root disposition and repair review are required before claiming this source is ready.

## Exact review target and coverage

Frozen **961125320766b3f02c3ff73ad6029f11117fedae**, tree **dc117ebe883a19fa51632bb808f4be8f4f14b030**, direct parent **2a78699d8e7c5d9f744d583776ee54ac9056b28d**. Checkout `/Users/dbox/repos/wt/vickie/vickie-alias-minimum` was clean. Inspected the full seven-file diff, accepted plan-v2, complete author report and source/summary/preservation evidence, plus surrounding authority delivery/materialization, consistency wait, service lifetime and HTTP routing code. No tests/builds/services/network calls/source edits or outward actions were performed. Only this report was written.

The supported data-object path correctly forwards a captured minimum into the existing same-owner observation: alias mapping, target existence/visibility and captured context are read in one authority entry; sourceCondition is checked before releasing the result. Service retries call resolveAlias anew after each finite wait. Alias-only repoint retains pos-1 and does not revise either Bead; a real Resource update is needed to reach pos-2. Controlled source tests distinguish repoint-only waiting from actual progress, and compare both complete target records before/after repoint.

HTTP retains method/auth/alias classification and rejects malformed/partial tuple shapes as before. Well-formed aliases bypass `requestFromUrl` and stay structurally outside the Type/Resource predicate. Success307 remains bodyless with canonical Location and current context; conditions/Accept do not fabricate an alias ETag or JSON response. Minimum failures precede Location release. Compatible absent/deleted/dangling cases use observed404; incompatible missing stays the disclosed visibility gap. No hidden-policy rule, target unreadability proof or retained-locator semantics was added. Ordinary alias signal forwarding is an intentional behavior change, accurately disclosed.

Inspected abort/drain tests: pre-aborted HTTP requests with/without minima assert zero SQL and zero owner entries; admitted wait is included in service close lifetime and rejects after cancellation. These test conventional enumerable signal objects and therefore do not cover the finding above. Provider/context/stored-target faults remain integrity/errors, not disguised404.

## Evidence and preservation checks

Independently rehashed all seven source files against `source.json`; every digest matches. Re-read Git parent and verified the installed test's entire prior35483-byte prefix is byte-identical, SHA256 **81a619fd9127694b36c9045b8adfc20de89926b7a41e55d2f667bc72e7af640f**. Thus earlier scenarios and shared no-ETag context helper were preserved. The new canonical Bead follow uses a separate context/actual-revision ETag oracle rather than relaxing that helper. No schema/catalog/runtime launcher/store/public-export/fixture changes appear in the seven-file diff.

Inspected author executions, not independently rerun:

- Affected source cohort508pass across9files; separate actual app disconnect file3pass. Frozen alias-focused7pass/2files with117 filtered. Clean typecheck/build, format/lint/boundaries/Read seal/diff check exit0.
- Actual packed tests4pass: separate collection38/1, compatible-missing10/1, Type/incident20/1, alias18/1. The alias completed-response ledger matches the planned eighteen method/URL/status entries and seven distinct URLs exactly. Seed and alias mutations establish pos-1/no-alias-effectPosition; redirects show a then b at the same position; separate b GET checks the complete seeded postimage and exact strong ETag; future/view/epoch and missing/auth/query refusals match the declared boundaries.
- Targeted installed Update2pass/200filtered. No combined-store86-request claim.
- Disposable source controls report intended failures when removing alias sourceCondition and caching the old target, followed by restored-source2pass. The initial pnpm dependency-layout rejection is correctly excluded from semantic evidence. Installed response-copy Location and position corruptions invoke the same positive redirect oracle without adding wire requests. They are explicitly assertion-sensitivity controls, not live faulty-server observations.
- Original literal ownedLinks omission and TypeScript wrapper failure are retained and explained; no runtime fault is disguised as a passing attempt.

Fresh protocol/server tarball hashes independently match the summary: protocol **4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227**, server **cfe3e338cef8b9186c45559d4d7cfab4c4d3f0b02b8f1be0ca845c0be14fd725**. Actual served payload **c4a7f1c3ffd0721de9e0db26516933bf57b49183f9c42461283dce97bd709a15** is consistently bound to the frozen-source runs; source metadata is not claimed as target attestation. Report SHA256 **8c716c323e92063c4a97fd0c83975631a1f3ccea48e1a27f5be0709755e536fb**; summary SHA256 **0a72d957bfbdaf8ce9341a4369afc77793ca73c94433222035779b0254562743**.

## Limits and cleanup

Review evidence is source inspection plus verification of completed author logs, not an independently executed reproduction. Author cleanup records all four installed children exited0/absent and the source clean; this reviewer launched no owned processes requiring cleanup. No whole-profile/public claim follows. The all-visible policy cannot demonstrate unreadable targets, and live-future503 versus missing-future500 remains explicitly unqualified. Root owns disposition, independent external review and integration.
