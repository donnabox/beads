# Actual aggregate HTTP count audit

Confirmed2314 actual completed HTTP observations; native21 must not be added again. Author self-audit disclosed; root owns acceptance. No execution or source change.

|Shard|Runner exchanges|Client observations|Setup|Actual HTTP|Native subset already in exchanges|
|---|---:|---:|---:|---:|---:|
|baseline|897|996|44|1937|21|
|reference-continuation|135|242|0|377|0|
|total|1032|1238|44|2314|21|

Source: `apps/bdptest/test-support/installed-read-fixture.ts:13–58` sends exactly one execute call for a legacy Read request, records its same response, and returns that response to the runner. It is the `execute` adapter passed at app1687–1707; exact raw actions instead take `exactExecution.execute`. Runner `packages/conformance/src/runner.ts:1210–1234,1356–1471` stores that same operation in result.exchanges. Consequently nativeReadObservations is an additional view of21 of897 baseline exchanges, not21 extra HTTP requests. I joined every actual native record by scenario/method/url/status in order to the exact21 authored HTTP actions lacking raw; all21 tuples matched. There are no continuation native records.

`read-update-exact-client.ts:124–172` performs one actual execute per client action, invokes observe once and returns the same output. Actual baseline996 and continuation242 client observations equal the corresponding runner actions counts; they are disjoint from runner HTTP exchanges. `installed-read-fixture.ts:61–139` performs one request per seed record:11 seeded scenarios×4 records=44, recorded separately and absent from portable runner actions/exchanges. Canonical51 baseline rows contain1893actions (897HTTP+996client); continuation5 rows377 (135HTTP+242client). Thus2270 authored+44setup=2314, with no double count or hidden bootstrap HTTP. Bootstrap configuration is IPC, not HTTP.

Actual sessions51+5 plus two per-shard bootstrap children imply58 aggregate children by the inspected lifecycle, not by summing process samples. Positive11 and negative2 complete the planned2327 expected observed requests; maximum2336 remains prospective failure envelope, not actual traffic. This count uses completed retained observations and source single-dispatch paths, not an independent packet-capture meter or universal no-network proof. Root separately audited stage exits, payload/source identities, negative assertion and112sampled PIDs; this note does not relabel those outcomes.

Exact retained shard evidence:
- `/Users/dbox/repos/wt/vickie/bdp-integration-logs/read-update-command-minimum-draft/installed-preparation/run/aggregate.json.baseline.json` SHA`6f1bcf7b113ac89f8f279b80544175550ba53e33cdf99135516d41e3576c3797`; HTTP1937.
- `/Users/dbox/repos/wt/vickie/bdp-integration-logs/read-update-command-minimum-draft/installed-preparation/run/aggregate.json.reference-continuation.json` SHA`0967a9468b256e6f2a5eedd642cb879a754c2f1164d2fe44a0362fc197b32426`; HTTP377.
