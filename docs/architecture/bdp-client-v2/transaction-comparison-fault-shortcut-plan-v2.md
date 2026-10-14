# Preserve unexpected owner faults before Transactional comparison (plan v2)

Status: revised proposal for independent closure after native + Claude council; **no source authoring or execution authorization**. Root accepted review findings1–8 with the clarifications recorded in the revision map. Original plan and reviews remain frozen. This supports the sole plan, `/Users/dbox/repos/wt/vickie/bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md`, [Autonomous profile completion / current priority](https://github.com/donnabox/beads/blob/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/design.md). T3 serving correctness takes precedence over another partial evidence row.

Source basis: frozen `70ac666002708202d2f93195eed90dfe8541300d`, tree `1353e8f4fb0b1149b3f4ff158335804bf3f1cecc`, `/Users/dbox/repos/wt/vickie/vickie-comparison-timeout-review-repair`. Final authoring parent remains separately assigned **verified main after PR120**. Root may explicitly re-anchor later; this plan does not assume PR120 merged, invent a current main SHA or change that dependency. Preserve this candidate and all historical reports. Re-anchor the five owned paths and read-only callers before drafting; do not rebase a frozen evidence source.

I read the completed independent [fault-shortcut audit](transaction-comparison-fault-shortcut-audit.md), SHA256 `1e16ad91aa31d5a68ab64314fa0d4373daae08b920137bafffd59f69f735c02f`. Its concrete reachability and authority conclusions agree with this source inspection. No tests, builds, service runs, packages, installs or product execution were performed for this revision. Only plan artifacts were written; the required remote session guard was the sole network access.

## Settled behavior and defect

`docs/specs/bdp.md:1956–1960` requires unexpected internal faults to produce bodyless500 without a BDP Problem or internal details. `:4588–4592` repeats that exception for Transactional mutation targets. The normal finite-comparison503 laws at3994–4026 and unknown-key admission unavailability at4517–4520 do not reclassify a retained internal exception. No product ruling, new error code, context requirement, minimum policy, Event question or schema change is needed.

`transaction-store.ts:1340–1359` catches real maintenance failure, retains its cause, cancels its interval, closes SQLite and sets closed. `enter` at341–345 correctly prioritizes ownerFailure and maintenanceFailure over ordinary closure. But `comparisonEpoch` at1724 returns undefined immediately for every closed owner. `transaction-comparison.ts:52–55` then treats that as benign stopped comparison; observed submission maps it to contextless comparison-refusal, and HTTP serves retryable503. A new valid authorized direct POST is sufficient; it needs neither an unresolved reservation nor deadline timing. The correctly guarded final deadline observation at1726–1740 is never reached.

## Minimal production change and scope

Change only comparisonEpoch's benign-close predicate to require closed **and no retained ownerFailure or maintenanceFailure**. All other states fall through to the existing `enter(() => metadata("epoch"))`. Prefer the explicit short block mirroring the adjacent deadline guard. Do not add a catch, new owner API, callback, context synthesis, metadata read after closure, store format or broad lifecycle helper.

Exactly five owned files proposed:

1. `packages/server/src/transaction-store.ts`: the guard only.
2. `packages/server/src/transaction-store.test.ts`: extend three existing close/fault cases to cover comparisonEpoch.
3. `packages/server/src/transaction-submission.test.ts`: observed/unobserved propagation and normal-close exact-shape cases.
4. `packages/server/src/transaction-http.test.ts`: one emitted-runtime maintenance-fault scenario using existing source-only seams.
5. `docs/design/transaction-implementation-progress.md`: explicit historical residual/successor correction and evidence limits.

No edits to transaction-comparison/submission/service/HTTP implementation, runtime/configuration, evaluator, schema, discovery, public exports, installed process, fixtures or conformance catalog. Any inability to exercise the existing seams is a preflight obstacle, not permission to add an injection API.

## Caller and response boundaries

Production search finds one comparisonEpoch invocation: `transaction-comparison.ts:52`. `transaction-submission.ts:116–132` supplies either the actual store or an observed adapter; both delegate to the same method. Both must now reject on an internal fault; unobserved **normal/healthy** return shapes stay unchanged. Explicit normal closure still returns undefined, so existing benign stopped503 behavior remains. Defensive changed-epoch ordering, deadline counting and same-epoch retry law remain untouched.

The caller is Transactional-only. RU development and public Read do not use transaction-store/comparison. Dependency and sequence-alias tests call the private comparison function directly, so include one exact existing regression each. Transactional singleton/batch direct mutations share submission; the sequence executor uses owner entry directly and already propagates sticky faults. No Type/History/alias normalization or creator-fact behavior changes.

Receiving precedes owner comparison: submission's captured-body preparation/key validation/authentication at80–115 remains; earlier400/401/403/415-style refusals are not changed to500. The new wire proof uses a valid authorized request with acceptable JSON. For the proposed valid POST, fault propagation reaches **the existing HTTP handler catch**, `transaction-http.ts:584–617`, specifically empty500 at614 (helper80–88), marker at616 and POST vary at617. The runtime fallback at `transaction-runtime.ts:253–260` is explicitly excluded from this proof: the handler catches this owner exception and returns normally. Pin response status500, body exactly empty, `content-length: 0`, `cache-control: private, no-store`, `x-bdp-development: non-attesting`, and **`vary: Accept`**. Require absent content-type (therefore no application/problem+json), Retry-After, Location, development-limit and all three BDP context fields. Internal sentinel details must be absent from body and every actual response header value. Transport-added headers need not form an invented closed header set. Do not synthesize a fault context or accept an otherwise sanitized fallback response missing the handler's vary marker.

## Deterministic source observations

### Owner and submission

Extend the existing store cases `returns a stopped comparison observation for a normally closed owner`, `preserves a fenced owner's comparison observation failure`, and `preserves background maintenance failure in a closed comparison owner` to assert comparisonEpoch behavior at the same relevant point. Normal close returns undefined; closed fencing throws the same retained failure; actual timer maintenance failure throws the maintenance wrapper with its original cause. Preserve every current deadline-observation assertion. Existing fault hook,25ms fake timer and fixture cleanup suffice.

Add a two-row observed false/true submission case using a **plain real owner** with the existing before-maintenance-commit fault hook and25ms maintenance interval. Create no dependency/sequence reservation, call neither `dependencies()` nor `cancelSequence()`, and do not borrow their teardown from neighboring tests. Trigger the actual live maintenance callback, restore real timers, then submit a fresh valid request. Expect rejection retaining the maintenance wrapper and original cause, not either refused shape.

The owner is frozen: do **not** use `vi.spyOn(f.store, ...)`. Follow existing submission-test spread-adapter precedent: `admit = vi.fn(f.store.admit)`, `admitObserved = vi.fn(f.store.admitObserved)`, then `{ ...f.store, admit, admitObserved }`, preserving the existing typed overload boundary if needed. Leave scopeEpoch, comparisonEpoch and fault state delegated unchanged to the actual owner closures. Assert zero calls to both admission delegates; the failure must originate at early epoch observation. The only owner teardown for these new fault variants is safe fixture store.close, plus timer restoration and directory cleanup; no faulted-owner entry during finally may obscure the primary assertion. Preserve primary and cleanup errors together. No synthetic stopped result, new production seam, extra test or workload is introduced.

Add two normal-close observed false/true cases pinning existing exact legacy refused versus comparison-refusal/owner-stopped shapes with no context; no callback or wait runs. These prove the narrowness of the fix rather than assuming the typed union preserves runtime behavior.

### Actual emitted HTTP

The existing HTTP fixture at `transaction-http.test.ts:76–175` calls the real emitted runtime, accepts `TransactionRuntimeTestSeams`, and supports a clock callback plus maintenance interval (`transaction-runtime.ts:51–65,319–327`). The fixture spreads `{ clock: () => time, ...seams }` at91 and provides retainedReceiptInventory at109–124. Use these existing options, not a new fault field:

1. Fake only setInterval/clearInterval, leaving network timers and setImmediate real. Start one disposable runtime with maintenanceIntervalMs25 and a supplied clock returning **literal1000** unless a local throw flag is set. This callback replaces the fixture's internal clock, so `f.setTime` is inert and **must not be used**. Startup and the already-planned first seed POST must succeed, with200 receipt and independently authored seed facts. Parse and retain that actual200 receipt for the later inventory equality; do not fabricate its identity or full receipt body.

   Name this same seed round trip **“selective interval/fetch compatibility”**: completion under selective fake interval timers is the runtime go/no-go for the otherwise unverified bundled-fetch timer interaction. It is inside the one planned new HTTP case and its first of two requests, **not an extra test, listener, HTTP request or preflight service run**. Static preflight names its assertion, existing deadline and stop result; build precedes later execution. If the actual round trip fails or times out under approved execution, retain that attempt and stop for the specific compatibility obstacle. Do not add sleeps, widen timers, run a substitute request or introduce an installed fault seam.
2. Set the flag only while advancing the fake interval25ms. The actual owner maintenance callback calls instant at store295–299; that clock throws a unique internal sentinel. Assert exactly one injected call and timer cancellation. Reset flag and restore real timers before the next network request. This is a real callback/exception/closed-owner path, not a manufactured response. Store-level tests separately exercise the before-maintenance-commit hook; this HTTP test does not claim arbitrary disk-error coverage.
3. A second valid authorized direct batch POST with a fresh key must produce the exact sanitized500 properties above. No unresolved dependency or comparison wait is needed.
4. Close through the existing fixture and call `retainedReceiptInventory()` **once**, including after the deliberately failing status assertion in the old-guard control. Require exactly `{ principal: "p", key: <authored seed key>, receipt: <actual parsed seed200 body> }` as its sole row. This opens one separate read-only SQLite connection after the owner closes and closes that connection in finally; it does not reopen a Transactional owner or listener. No retry receipt or failed-key capture. Collect primary status/assertion failure and close/inventory failures together, recording actual status and inventory/close completion so the derivative's named503 result cannot mask cleanup failure. This is receipt inventory evidence, not a complete graph-state/replica proof.

Exactly two new HTTP requests, one listener lifetime, one disposable store directory and one Transactional owner open, plus one later read-only inventory connection, zero installed children. Capture release/close outcome and preserve primary+cleanup errors. Restore timers even if startup or assertions fail; drain/close before directory removal. Do not modify historical global teardown. Static draft preflight records selective interval/fetch compatibility as unproved; if approved execution cannot drive the emitted interval while the planned seed fetch completes, stop that attempt and return its specific obstacle; do not fall back to sleeps or a new installed injection option.

## Selection, counterfactual and proposed execution bounds

Prospective14 source cases:

- Three extended owner cases above.
- Four new submission variants (fault × observed/unobserved; normal close × observed/unobserved).
- Existing submission `keeps full-owner unobserved comparison exact and excludes a caller callback` and `distinguishes a stopped final observation at the reached comparison deadline`.
- Existing service `uses the original finite comparison budget while an earlier sequence dependency is unresolved`.
- Existing dependency `proves static mismatch immediately, retains no timeout outcome and keeps the owner`.
- Existing alias `uses one bounded comparison deadline and leaves the owner's reservation intact`.
- New HTTP maintenance-fault case plus existing `keeps a live listener from serving through an owner fenced by injected commit uncertainty`.

The existing HTTP fencing regression has three requests and one reopen (two listener lifetimes). Thus proposed source positives total5 HTTP requests/three listener lifetimes. A single isolated baseline derivative restores only comparisonEpoch's old unconditional closed shortcut and runs the new HTTP case: after seed and actual callback failure, the second response becomes503. Require the named `maintenance failure is internal500` assertion, actual503 and retained cleanup proof; reject generic setup/timeouts as sensitivity evidence. Proposed control2 HTTP requests/one listener. Production checkout stays frozen; derivative is separate and restored. No additional mutants are presumed necessary.

### Static owner-open and directory inventory

“Store” is not an interchangeable count of directories, owner instances and SQLite inspection handles. Source inspection of the exact selected fixtures/cases at70ac gives this planned ledger (not executed results):

| Selection | Executions | Disposable store directories | Transactional owner opens | Extra read-only inventory connections | HTTP / listener lifetimes |
|---|---:|---:|---:|---:|---:|
| Three extended owner cases |3|3|3|0|0 /0 |
| Four new submission variants |4|4|4|0|0 /0 |
| Two existing submission regressions |2|2|2|0|0 /0 |
| Existing service finite-budget case |1|1|1|0|0 /0 |
| Existing dependency case |1|1|1|0|0 /0 |
| Existing alias case |1|1|1|0|0 /0 |
| New emitted HTTP case |1|1|1|1|2 /1 |
| Existing emitted HTTP fencing case |1|1|2|0|3 /2 |
| **14 positive cases** |**14**|**14**|**15**|**1**|**5 /3** |
| One old-guard derivative of new HTTP case |1|1|1|1|2 /1 |
| **Source plus control total** |**15**|**15**|**16**|**2**|**7 /4** |

The existing fencing test at `transaction-http.test.ts:867–890` calls f.reopen once in its original directory; its two owner opens are not two directories. No other selected existing case reopens: store fixtures at63–90; submission17–49 with existing selected cases1105–1177; service54–111 and409–526; dependency44–69 and197–222; alias40–64 and358–385. The proposed four new submission variants each use one plain fixture. The two read-only inventory handles come from new HTTP positive/control and are separate from the16 owner opens; together they make18 transaction-file SQLite connection lifetimes, not18 stores. Temporary build/package directories, if later authorized, are separate and not silently included in these source-store counts.

Draft preflight must resolve exact parameter-expanded titles, request bytes and route inventory, source hashes, timer/control selectors, fault-call counts and concrete primary-plus-cleanup finally paths. Reconcile the implementation to this static **14 positives +1 derivative /15 store directories /16 owner opens /2 inventory connections /7 HTTP requests /4 listener lifetimes /zero installed children** ledger before execution. A source change requiring additional opens, directories, cases or requests returns to root; the former loose24-store ceiling is removed. maxWorkers1–2, existing test/hook deadlines unchanged,120s command supervision; no unfiltered service/HTTP/full suite and no additional compatibility probe. No measured success or cleanup is inferred from this static count.

After green source/static controls and a clean signed source freeze, propose only the existing installed **direct refusal14/1** regression on fresh actual packages, with its unchanged request/scenario/outer/cleanup bounds. This exercises the changed installed direct-call path in its ordinary healthy configuration; it does **not** induce a maintenance fault or qualify default installed fault sensitivity. No new installed test, clock option, child command or public flag. No sequence installed rerun is proposed: its unaffected execution path is covered by the bounded source service/alias regressions. Exact pack/install/request/PID/config/payload attribution must be measured in the later draft preflight. Root may omit or amend this ordinary regression during plan acceptance; it is not authorized now.

## Done-when and history

Before source authoring: independent closure of this revision, root acceptance and verified-main-after-PR120 parent/ownership release (or later explicit root re-anchor). Before source execution: normal-format draft and measured preflight acceptance. No authoring follows merely from this revised plan. Build before emitted-runtime tests; preserve every failed attempt and do not automatically rerun unexpected failures.

Delivery predicates: only the five approved files change; normal closed behavior stays exact; both retained fault terms reach existing enter; observed and unobserved submission reject before admission; actual emitted wire is sanitized bodyless500 without retry/context/detail; old guard reliably produces named503 failure with cleanup; selected dependency/alias/healthy deadline regressions pass; static/Read gates pass; source-bound package regression (if accepted) is clearly separate from fault sensitivity; all owned processes are absent.

The current progress text at1781–1784 is historically accurate for30035/70ac. Reword it to say that those snapshots left the early shortcut unrepaired and add a separately dated/source-bound successor note once this slice is verified. Do not retroactively change their reports, prior18/five or27/two evidence, or claim this repair was part of PR120. No public profile activation/advertisement, broad receiving guarantee, live epoch rotation, Event/minimum ruling, automatic recovery policy or full conformance claim.
