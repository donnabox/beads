# Transactional no-op plan v3 — review dispositions

Plan-only revision; no implementation, tests, builds, services, network, credentials, BDP source changes or outward actions. Root's requested bounded corrections are incorporated for independent native closure. Implementation remains held.

Exact v3 plan: `transaction-next-conformance-plan-v3.md`, SHA256 **8d5b9baab8c98067729a297d6bdf10ab03b1ae9ce9717b810d4b24b7df29f495**.

Preserved reviewed inputs:

- Original plan `transaction-next-conformance-plan.md`: `1ce661b78ee37a2c22a27239af6ea318c1942149dbd828d0b9d64de6dcce4127`.
- V2 plan: `38d0b851848e5d0006ec972d4ea7c806c64ef857b2e6843314793e02cc602ea8`.
- Original native review: `392474a3f41df38eb02ceb4376520b254c98101d69db1bb4304887eaf2d57d0e`.
- V2 native closure: `69cbd6955a73c1451e6459131d2c457554f0c3b2e4e804bc77bb722f61e17079`.
- Completed external original review: `b25abbc1f9c365bfdbb112133709325bc1e31170b825f0c06c5cbd95badc9c9f`.

Read-only source basis is c534fcab840ef2e3753ab179feae10d3ce6431e5/tree48537438975a4c56ee1a5bb6056013bbddd38e03. Root has verified merged main ea584103f0dc90ae365970bb672f24bedf318e42 has the same tree and assigns it as future authoring base. This checkout does not have that merge object locally; no fetch was performed. The plan clearly distinguishes the inspected tree from the root-supplied main identity and requires author preflight verification.

## External findings

**F1 High — accepted with a stronger existing non-null validation than the review's suggested presence assertion.** Add action15 `n-changed-groups`: the same literal GET changes/?after=genesis after the genuine effectful batch. Require schema, current/headpos-2, one seed group at pos-1/previouspos-0, original seed transaction/full postimages, five Events and eventCount5. This page must have a non-null continuation, while earlier reads have headpos-1/next:null. It is a real changed-state comparator, not a claim that the owner returns two groups per page.

The external proposal `exists:true` alone would accept null. Root correctly rejected that as sufficient. Existing raw capture `{binding:"n-changed-next",from:{kind:"json-pointer",pointer:"/next"}}` validates an absolute scoped HTTP(S) URL and is intentionally never fetched. I inspected the actual seams: runner1465–1502 runs captures after assertions and exact-wire checks; captureValue2467–2499 rejects null/non-string/non-URL/out-of-scope values; manifest795–808/1510–1562 and runner3090–3166 permit an unused capture while rejecting duplicates/unknown references. The Transactional receipt registration gate at transactional-observation174–197 returns for the GET source, so this ordinary URL capture grants no receipt/page authority. Static pins preserve the capture; successful action completion is the non-null observation. A failed capture is an observation/action failure, not a fictional named assertion failure. No new DSL/helper/route/dynamic grant. The first two controls abort their scenario at action4, so v3 no longer implies they reach later state/Event/group observations.

**F2 High — accepted as staged, measured bounds; no silent220→240 increase.** Retain64KiB manifest/2200 value nodes/16KiB coverage/220 assertions as preflight stops only. Before installed tests, author submits actual normally formatted candidate M/N/C/A and exact action/assertion/capture/coverage ledgers for root acceptance. Committed measured growth caps use actual candidate plus declared margins no greater than8KiB/128nodes/2KiB; new/total assertion counts are exact A and1316+A, actions15/166. The complete required row must fit220; otherwise stop and report. The review's extrapolated120–240 range is not an approved increase. All transport/runtime/deadline caps are unchanged.

**F3 Medium — accepted with indexing made unambiguous.** The new route follows beads/new, precedes beads/b: **seventh entry, one-based index7 / JavaScript zero-based index6**. Existing ten routes retain relative order. The independently authored11-entry sequence and full canonical digest preimage are required; do not sort routes or copy an observed digest into its own oracle. This corrects the original placement after events/ without silently skipping beads/new.

**F4 Medium — accepted; correct the external arithmetic.** Explicitly change positive pass count10→11 and include the new row in the existing2KiB POST/32KiB response filter (or equally strict separate loop). All15 row exchanges are covered, including three receipt-client GETs. The actual row contains12 raw requests (3POST+9GET) plus3clientGET =15, not the review's suggested13. Full manifest becomes166 actions =113raw+46receiptGET+7reopen =159 exchanges;18 scenario/reopen sessions and19 children including bootstrap. Four complete positive/control cohorts are bounded by636 exchanges/76 children, with actual failed-run counts separately reported. No additional cohort or process model is introduced.

**F5 Low — accepted with truthful purpose.** The committed measured-delta guards and fixed absolute preflight ceilings serve different purposes. Small measured margins bind the accepted artifact; absolute limits preserve the initial ceiling if a later baseline/delta changes. The absolute guard can be redundant for a given measured candidate and is explicitly described that way, not represented as a second independently binding inequality. Margins shrink when necessary to stay below the fixed ceilings472763B/13091nodes/76453B. Both old live conjuncts are visibly replaced; no obsolete smaller inequality remains by accident.

**F6 Low — accepted, with exact precedent characterization.** Correct source ranges to transaction-store.test507–538, transaction-changefeed87–106 and transaction-batch738–771. The owner test's Bead operation is guarded (`expectedRevision` at515), while its Link metadata operation is unguarded; the new installed batch is entirely unconditional. That existing test supports the underlying mechanics but does not itself prove this exact installed unconditional partition. No unsupported claim that every unconditional path lacks any other in-process coverage is added.

## Native v2 findings retained

1. The third existing control still changes both the group-position tuple and observed-head expectation, requiring both named failures on a real200 page. It remains explicitly a cardinality/head sensitivity control, not a standalone proof of no later group. The new successful action15 supplies the separate actual advancement comparator; no fourth control is added.
2. All three group reads now explicitly pin eventCount5 alongside the exact five-fact Event array. The third read is still the seed group, so eventCount is5 even though the Scope head and full Event source have advanced topos-2/seven facts. The new assertions remain inside the hard220 preflight threshold.

## Unchanged boundaries and pending gate

One new partial catalog row, four evidence/test files, unchanged ten prior raw manifest/coverage objects,716-byte fixture, catalog/spec/runtime/schema/DSL and whole/public flags. Root-selected sole U3/T4 plan remains authority. No pending SSE/minimum/incident rule is resolved or bypassed. All known body, runtime parser,63s positive/90s cohort/120s outer/5s exchange/10s lifecycle limits remain. The only added request versus v2 is the final15th GET to an already planned route. Root/native closure must precede implementation; root acceptance of actual measured candidate guards must precede installed testing. These documents contain planned observations, not executed proof.
