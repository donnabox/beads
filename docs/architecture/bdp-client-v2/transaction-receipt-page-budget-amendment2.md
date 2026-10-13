# Receipt pagination amendment2 — existing URL-only capture grammar

Log-only proposal; no repository installation or fence change. Frozen helper43e0aff remains unchanged. Ordinary evidence successor is mutable with the previously approved candidate. Initial emitted-source installed run failed honestly at action11 after all earlier page/root/HEAD oracles passed; exact artifacts preserved in evidence-positive-draft.log/json and extracted draft-pagination-failure.json.

## Cause and smallest remedy

runner.ts captureValue (around2470) returns only absolute HTTP(S) URLs for HTTP captures; arbitrary JSON captures are available only to programmatic client actions. The reviewed plan's p-later-state HTTP /items capture was therefore impossible. Do not extend public capture grammar, add a private collection action, weaken the final-state oracle, or add requests. Remove that single capture. Replace p-final's unavailable equalsBinding with the identical independent literal32-record ordered tuple already checked by p-later-state. This preserves independent semantic Type/properties/metadata/ordered identity and count checks after real reopen, but does NOT preserve the planned full-items-envelope equality for opaque revisions or additional provenance fields. That stronger check never ran successfully: it was an unimplemented plan assumption, not a previously passing oracle. Disclose this residual explicitly; do not call literal tuples identical/full-envelope preservation. Earlier immutable root/page JSON equality still uses supported client captures and remains exact, including opaque revisions/provenance in those retained receipts. The remaining gap concerns current Resource collection envelopes after reopen, not recorded receipt facts. Current collection schema still validates each returned Resource. No helper code change is needed.

## Exact standard-formatted measurements

| Metric | Approved candidate | Corrected candidate | Delta / proposed fence |
|---|---:|---:|---|
| Manifest bytes |366621|377077|+10456; propose total370KiB=378880,1803headroom|
| Growth from238693 |127928|138384|propose136KiB=139264,880headroom|
| Authoring JSON-value nodes |9843|10071|+228; propose10200,129headroom|
| Coverage bytes |54804|54804|original assertion ID retained; +16KiB from41812 unchanged|
| Actions |18new/139total|18new/139total|unchanged, no contingency|
| Resources/sessions/children |32/15/16|32/15/16|unchanged|

Corrected manifest SHA256 `fb0a71cb5f568a3c65bbae35f2f0946710e5906088dc11dd31fe88c6bbdbcc29` at manifest-final-literal.formatted.json. Corrected coverage SHA256 `f21d2822b789ff3e03c633b89c2412ee3bb2e7b38ece62e8eed0923d930f9172` at coverage-final-literal.candidate.json. Narrow artifact diff saved in amendment2-manifest.diff. Original reviewed candidates remain unchanged. Final loopback source still needs all8 old raw object preservation verification after installation; no old object is intentionally changed by this candidate.

The10200 fence is strictly authoring test JSON-node accounting. Runtime OBSERVED_JSON_MAX_CONTAINER_ENTRIES=10000 is unchanged; it applies separately to each actual response. No parser, transport, owner, mutation, raw-body, response, wire or deadline setting changes.

## Actual wire observations, bounds unchanged

Existing failed draft already observed seed POST raw5104B (p-update157,p-later152), below16KiB. Responses observed before the capture failure: seed8985B, rootseed8985, beforecollection6903, update8898, rootupdate8898, page382, HEAD0/0, changedcollection6935, laterreceipt737, latercollection6933. Max actual8985B remains below32KiB. Corrected final request itself is unchanged GET beads/?limit=32 after the same reopen, and assertions do not change wire requests; this is not a claim that the unexecuted final/reopen responses already passed. A complete fresh final-source packed run must verify every response bound and cleanup.90scohort/120souter/5swire/10sprepare+cleanup unchanged.

Root/native review must approve this narrow correction before installation/fence edits. If declined, checkpoint as incomplete; do not introduce URL-capture semantics or silently drop final-state evidence.

## Native clarification: existing grammar cannot retain collection opaque envelopes here

ScenarioCapture on HTTP accepts only response-url/header-link/json-pointer results that are absolute URLs (runner.ts captureValue). json-pointer equalsBinding can compare arbitrary captured client values, but these collection reads are HTTP and have no permitted client projection. json-equals accepts literal/fixture values, not an earlier HTTP body. response-metadata-equals compares only status/named headers; header equalsResponse is ETag-specific and cannot attest individual opaque Resource revisions/provenance. An optional collection validator would not be equivalent to full-envelope equality or a portable required check. With unchanged18 actions, current private receipt-read action and no new grammar/runtime/fixture, there is no existing supported assertion that captures/recompares the entire current collection's opaque fields. A separate actual-wire app callback witness would be additional non-DSL evidence, not a substitute silently relabeled as this canonical assertion; it is not proposed here.

Scope disposition requested: retain independent semantic current-state/schema/count/order checks and exact recorded root/page body retention, explicitly leave current collection opaque revision/provenance equality across reopen unqualified. Whole row remains partial. No new capture semantics, helper or extra request is requested.

## External final dispositions

- Preserve p-final assertion ID `exact-later-items` while replacing its kind/payload with the complete literal tuple oracle. The coverage assertion mirror therefore remains exactly the original54804-byte candidate, including `p-final:exact-later-items`; no stale binding. Corrected manifest377077bytes/growth138384/10071nodes fits the proposed136KiB/370KiB/10200 fences (880growth/1803total bytes and129nodes headroom).
- Thin authoring headroom is intentional for fixed18new/139total actions; no contingency or runtime limit change.
- Final collection after reopen remains unexecuted in the failed draft. Fresh source-frozen packed execution must prove the ordered oracle; if it fails, diagnose rather than relax ordering.
- URL-only capture restriction is specifically HTTP captureValue. Private client captures support arbitrary JSON, so retained root/page whole-body equality remains exact. The separate current collection opaque revision/provenance residual remains explicit and unqualified.
- External final text was read by root from session83904 while process completion was pending; no successful exit is inferred here. Root acceptance remains required before source adoption.
