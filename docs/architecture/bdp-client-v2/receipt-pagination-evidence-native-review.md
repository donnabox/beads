# Native source review: receipt-pagination evidence

## Finding

**Low — only one of the two required literal-state oracles is statically pinned.** `packages/conformance/src/transactional-cumulative.test.ts:892–909` checks that `p-later-state` has no captures, then independently verifies the full expected32 tuples only for `p-final`. Amendment2 explicitly requires both the before-reopen and final literal oracles to be independently pinned. Both current manifest arrays are correct, so this is not a present runtime/state failure. However, a later edit can weaken or truncate the earlier tuple expectation without this new guard detecting the change; that earlier observation is the independent evidence that live state has advanced before receipt retention is tested. Apply the same independently constructed expected32 tuples, projections and ordered-array checks to `p-later-state`'s `ordered-state` assertion as well as `p-final`'s `exact-later-items`. Do not merely compare the two manifest copies to each other.

No other actionable source finding in this four-file diff. Final installed validation was still running during review; no completion or readiness claim is made.

## Exact source and reviewed scope

Frozen `df51dfcd8d68e3d0a4e91c7be285c3ffd669c3f0` in `/Users/dbox/repos/wt/vickie/vickie-receipt-page-evidence`, reviewed against helper `55299682`. Source remained clean at the start and end. Exactly four evidence files changed: app transactional-matrix test, conformance transactional-cumulative test, transactional manifest and coverage. Read the full test diffs, all new action/assertion shapes and literal result/state contents, accepted plan612beec5, original budget acceptance and amendment2 root acceptance. The previously reviewed helper is not reclassified as part of this evidence diff.

## Oracle and scope assessment

The18-action row seeds32 Task Beads in reverse canonical order, independently checks the seed receipts and complete initial collection, and updates all32. Inline mutation results require one matched-count32 plus31 updated entries in canonical p00–p30 order, with every identity/Type/properties/metadata asserted. The exact32-entry outcome array and absence at results32 prevent an extra inline result. The observed continuation page independently requires exactly one updated p31 result, the same parent receipt and terminal next:null. Thus matched consumes an entry slot; the witness would not pass if the32 updates fit alongside an uncounted matched entry.

The root/page capture chain preserves actual observations rather than expected synthetic envelopes. Canonical mutation POST discloses the separately named root and page URLs; private root GET is checked against the original whole POST body by the reviewed helper before the manifest captures the root body and opaque transaction/positions. The first page GET has independent parent/result/postimage/terminal assertions before capturing the whole page and p31 Resource. Later receipt GETs and reopened GETs compare whole root/page JSON, including embedded opaque fields; page history still says changed after a later real mutation sets live p31 to later. The current pre-later collection also equals the recorded full p31 Resource, adding a complete recorded-result/current-state comparison at that point.

The final current collection now uses the approved complete literal32-tuple semantic oracle, with exact count/order, identity, Type, properties, metadata, schema and terminal next:null. No impossible HTTP array capture remains. This deliberately does not prove current-collection opaque revision/provenance continuity across reopen; root amendment2 explicitly accepts that residual. Full recorded receipt root/page equality remains intact. The current semantic-state distinction is enough for this partial pagination witness and is not represented as whole-row credit.

The app checks the exact12 reopened dynamic exchanges, including reopened page HEAD. Every serving generation must advance beyond its original grant while runtime and configuration identities remain equal. Each request target discriminates root versus recorded page. Required context fields are checked against successful actual client-projection assertions, not unavailable public-report headers. Bodyless HEAD remains checked. All new private receipt reads have current context, private/no-store and absent ETag/Last-Modified assertions.

Five new controls deliberately alter one expected fact after actual successful execution: inline boundary, page member, parent, historical title and terminal next. Each requires a completed exact request, HTTP200 and exactly its named failed assertion; client projection results are used for private page actions and raw results for p-update. These are expectation-sensitivity controls, not service-fault injection claims. Existing controls are unchanged.

Coverage remains false and explicitly limits evidence to source-defined no-override32, matched+updated outcomes and ordinary fixed-view reopen. Advertised page.maximumItems, other outcome classes, policy/erasure/retention/failover remain unobserved. No separate page-restart or public-profile credit is added.

## Independent read-only checks

Using local git reads and Python JSON/raw-span/hash inspection (no execution tests):

- All eight predecessor manifest objects and all eight coverage objects are byte-identical to55299682. Only pagination is inserted, at canonical position8of9.
- Manifest377077 bytes, SHA256 `fb0a71cb5f568a3c65bbae35f2f0946710e5906088dc11dd31fe88c6bbdbcc29`;10071 value nodes;18 new-row actions/139 total. Coverage54804 bytes, SHA256 `f21d2822b789ff3e03c633b89c2412ee3bb2e7b38ece62e8eed0923d930f9172`.
- All307 new assertion IDs exactly match the ordered `action:assertion` coverage mirror; all TX whole flags remain false. The preserved final assertion ID avoids the earlier mirror mismatch.
- Fixture remains716 bytes and SHA256 `3a7899ced771d4880066d98b0000c53f801ece53458502cc4f3bc641146c9a85`, byte-identical to parent. No route/config/runtime source changed.
- Independently checked every p00–p30 inline identity/Type/document fact and all32 tuple values in p-before, p-changed, p-later-state and p-final. Both current later arrays match the independent expected semantic state.
- Raw seed is5104 bytes. Source caps match explicit amendment2:136KiB growth from238693 (880 bytes remain),370KiB absolute (1803 remain),10200 nodes (129 remain),16KiB coverage growth (3392 remain). Existing M/Z2KiB request/32KiB response guards are unchanged; new receipt row uses16KiB request/32KiB response guard. Execution/deadline limits are not raised.

## Execution boundary

Reviewer ran no tests, builds, servers or network actions and edited no source. Author validation was active: saved stage records showed exit0 through source/focused/pack/install checks; installed-tx was still running. Existing `tx-positive.json` already named df51 and showed transactional.receipt.pagination passing, but it was not treated as the completed suite, final cleanup proof or final report. The result includes255 catalog mappings; unplanned rows must remain honestly unobserved. Final installed nine-row/five-control and genuine RU evidence, all process cleanup, external council and root readiness disposition remain separate gates.
