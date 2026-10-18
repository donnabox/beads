# Next unblocked Transactional serving work — bounded audit

## Recommendation

**No further independent, advertisement-blocking runtime repair is justified by this bounded inventory without resolving an existing contract question.** The previously identified settled current-read repairs are implemented. Finish the separately owned current-`perform` receiving repair and its review; do not duplicate it. Keep discovery HTTP advertisement off. The smallest next concrete wire repair is **Scope changefeed SSE reconnect using `start=now` plus `Last-Event-ID`**, but it remains conditional on the already pending precedence ruling. This is not authorization to infer that ruling or to substitute internal preparation for serving progress.

Scope of this conclusion: the actual explicit gaps in the requested minimum/current-view/Event/discovery inventory, not a claim that every mandatory profile obligation has been audited or that cumulative qualification is complete. Additional executable conformance work can proceed under its own accepted scope; it is not a newly found runtime-serving blocker.

## Exact basis and inspection

Read-only clean `/Users/dbox/repos/wt/vickie/vickie-alias-minimum-integration`: **7171251d282769c1b9e2a39aed63e16f8b24ed96**, tree **76931ffeede5532cf11bd742d41dac9cdf5e7f01**, identical tree to root-reported merged main **d968d0ee7df4b9510944fb1cbf76f8355d4e1a5c**. Read the sole fork `docs/architecture/bdp-client-v2/design.md` (live ledger plus current priority), consistency-rulings audit, discovery-serving v2 plan, metadata Event and start-now assessments. All code/spec pointers below are repository-relative at717. No network, builds, tests, servers, source changes or outward actions. Local Git/primary-source inspection only. The in-flight5d854 receiving source is separately owned, not incorporated or assumed qualified here.

## What is already implemented

| Surface | Verified implementation and relevant contract |
|---|---|
| Current Bead/Link points/properties, local Type points, initial Bead/Link/Type collections and incident-Link pages | `packages/server/src/transaction-read.ts:30–58` includes precisely these routes and excludes continuations. `minimumFor` attaches same-entry source conditions; service `readAtMinimum` supplies existing finite waits. Spec `docs/specs/bdp.md:5758–5786,5793–5807` requires current coherent context/minimum before conditional shortcuts. The separate receiving repair is the remaining owned work here, not a missing route to rediscover. |
| Current aliases and compatible missing targets | Read alias `resolveAlias` and HTTP `:195–209,525–536` use an alias-specific minimum path before canonical parsing. Compatible missing source-condition checks return normally (`transaction-read.ts:139–160`); simultaneous missing/behind/foreign remains explicit. Alias semantics and receiving repair are in this source. |
| Fresh Snapshot minimums | HTTP `:380–425`, service `snapshot` `:331–354`, owner `observeSnapshot` `transaction-store.ts:1428–1469`: current owner context checked before fresh capture/preflight, finite wait/signal, nonallocating conditional/HEAD path. Spec `:5829–5843,5915,5805–5813`. Retained handles/pages are intentionally separate. |
| Receipt/page current-serving minimums | HTTP `:498–523`; service `receiptPage`/`receiptAtMinimum` `:422–495`; owner receipt read checks. The spec explicitly resolves this domain: `:5233–5244` distinguishes current serving observation from immutable execution facts. It cannot automatically settle Snapshot/collection anchors. |
| Finite/SSE Event and Scope changefeed minimums | HTTP `:427–496`, service `changes/events/beginEvents`, owner `readEvents` `transaction-store.ts:1508ff`: same observed context, actual cursor/minimum validation and finite wait; separate explicit gaps below. Do not describe all Event minima as absent. |
| Internal discovery candidate, limits, directory and minimum observation | `transaction-discovery.ts:24–66` parses the actual fixed twelve-map, pos-0 replay floor and enforcement-bound page/Selector/diagnostic limits. `transaction-service.ts:239–265` captures/checks context in one owner entry and assembles after release, with finite minimum wait. This was missing in the older serving plan table and is now done. HTTP `:538–544` still intentionally refuses Scope/descriptor/directory. |

There is no need for another internal discovery observer or another current-route predicate expansion. Advertising `profile:transactional` would assert the complete cumulative contract (`bdp.md:73–86`), even on a private listener; `claimEligible:false` is not a substitute profile.

## Actual remaining explicit gaps and decision boundaries

| Gap | Current source / law | Can existing text alone select the repair? |
|---|---|---|
| Incident Bead Events for metadata-only Link updates | `transaction-events.ts:259–264` throws `incident-link-metadata`. Spec `:5547–5554` names incident Link mutable **properties**; updated payload/metadata laws separately permit metadata-only transitions. | **No new resolution found.** The pending include/exclude decision remains. Payload-schema permission does not decide projection membership. Scope/Link facts already exist; no second mutation or synthetic patch should be invented. |
| Scope SSE `start=now` + reconnect header | HTTP `:438–439` refuses before existing header→after mapping `:440–443`. Spec `:5965–5975` defines starting intents; `:6112–6124` explicitly says header overrides **after**, not start. | **Pending narrow ruling.** Recommended resume-after-header remains sensible; the Resource EventSource clause belongs to a different cursor domain and does not silently extend this rule. |
| Retained Snapshot/collection minima | HTTP `:385–386` and Read predicate reject retained minima; owner `:1430–1431` permits only fresh capture/preflight. Spec `:5899–5907,5923–5935` freezes old anchor/state; `:5783–5786` speaks of a representation at/after minimum. | **Anchor versus current-serving comparison remains unresolved.** Receipt-specific wording is not universal. A minimum already below both anchors is an intersection case, but a special partial branch would not close this blocker; no such branch is proposed merely to generate activity. |
| Partial/repeated/malformed tuples and unrecognized positions | `transaction-consistency.ts:20–53` accepts one complete private tuple/position grammar; HTTP preserves an unsupported gap. | **Error/applicability selection still missing.** Schema wireToken syntax does not determine raw repetition, omitted-field behavior or unknown-token diagnosis. Known future pos-N already has a real finite catch-up path. |
| Minimum headers on mutations | HTTP `:295–301` refuses after principal check. Spec `:4490–4530` enumerates admission; `:5776–5786` is read consistency, with no command barrier. | **Do not invent a mutation barrier.** Ignore versus reject/applicability needs an existing or new narrow ruling; absence of a speculative barrier is not itself a normative defect. |
| Missing/hidden plus incompatible minimum | `transaction-read.ts:139–160` retains visibility gap. Resource unknown/deleted/unreadable404 and minimum laws do not order this collision. | **Pending precedence choice.** Compatible/satisfied404 has already been implemented. Fixed all-visible behavior cannot qualify hidden-policy precedence. |
| Embedded set Selector refusal/admission phase | `transaction-set.ts:132–143` converts Selector parser/complexity errors to the explicit development gap; HTTP `:322–323,348–349`. | **Existing set-admission question remains.** Read query errors cannot simply be promoted to preadmission syntax or retained execution errors without reconciling the ordered key/admission law `:4490–4530`. |

The two Event questions do not themselves block unrelated code. Here, however, the other inventoried route closures also reach already documented choices; they are not safely unblocked just because they are unrelated to Events.

One small observed lifecycle asymmetry is that receipt/page minimum waits do not receive the actual Request signal (`transaction-http.ts:503–517`, service `:422–495`), whereas current reads/Events do. This may merit bounded cancellation hardening, but those reads already enforce the minimum with a finite budget and return the specified result. I found no source-law basis to call this an unimplemented wire requirement blocking discovery. It is not recommended as a substitute for closing the substantive gaps, nor expanded into a receiving audit.

## Smallest next slice once the existing ruling is issued

**Conditional proposal: Scope changefeed reconnect header replaces an original start=now intent.** This is the smallest production boundary and gives a cold client a practical lossless reconnect improvement. Root must first confirm that interpretation (and authorize its normative wording/citation alignment). If the decision is to reject the combination instead, it must supply the actual status/code; do not implement the recommended path by default.

Runtime ownership: only `packages/server/src/transaction-http.ts` remove the specific private start/header refusal after the ruling. Keep exact URL grammar, authentication, raw repeated-header rejection, finite JSON ignoring the header, and the existing one-header `{after: checkpoint}` path. No store, cursor, Event projection, pump, fixture, public admission or limit change.

Tests: `packages/server/src/transaction-sse.test.ts` and existing `apps/bdptest/src/transactional-sse-installed.test.ts`; progress documentation only if dispatched. Replace the old500 expectation (`transaction-sse.test.ts:274`) and exercise existing actual stream/finite helpers. A separately accepted normative edit must clarify Scope precedence without changing Resource EventSource grammar. No catalog whole-credit assumption.

Bounded executable done-when:

1. On a start-now stream, receive a genuine completeG1; disconnect, commitG2, reconnect the same URL with G1's genuine checkpoint. First observed group isG2, not a new-head empty stream. CommitG3 and observe it next. Compare full groups with independent finite-afterG1 results and literal intended mutations, excludingG1 replay.
2. Reconnect after genuine latest checkpoint and after actual store reopen; preserve complete group equality/checkpoint bytes. A counterimplementation ignoring the header or re-anchoring must fail, not merely return200.
3. Header checkpoint survives a controlled minimum wait. Existing header-free start-now first-anchor test `:351–379` remains unchanged. Existing effective-after cursor tests `:579ff` remain active.
4. Repeated/empty/foreign checkpoint, unauthenticated, HEAD/conditional and finite-JSON header ignorance follow the existing after-header behavior. Invalid URL intents are never rescued by the header. No new precedence choice is smuggled into this repair.
5. Real disconnect/close cleans stream slots and all installed children. Use existing deadlines and the smallest explicit installed exchange ledger; measure authoring/execution budgets before implementation, without raising caps. Exact clean source, payload and per-configuration bindings retained. Full profile claims stay false.

Until that ruling or another documented choice is resolved, report the serving lane as blocked on the precise contract decisions after the currently owned receiving repair completes. Do not create a new runtime abstraction, fake discovery capability or hidden semantics merely to keep the lane busy.
