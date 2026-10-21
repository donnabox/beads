# Deletion-safety evidence: stage1 measurement gate

**Complete source candidate; STOPPED before installed execution.** Frozen ordinary local commit `2e742307845f52c23c4a16e6150fcf2201713fb5`, tree `dffd239a283a071b5c31f6d365a8d85b38d21b25`, direct parent `5995f6b08ebd045e4343059c2da1cffd38b09b80`. Checkout `/Users/dbox/repos/wt/vickie/vickie-deletion-safety-conformance`, branch `codex/vickie-deletion-safety-conformance`, clean. Exactly the four authorized evidence files changed. No push, PR, bus, main, runtime, protocol, schema, catalog, fixture or public-flag changes. No installed process/cohort was launched.

Authority: accepted v4 plan SHA `c4d8a96c63656ad813d601a05f8cf6c5f89eaf7bc9d039e62f32d34b1faee837` and root acceptance `transaction-after-noop-root-acceptance.md`. Root must accept these actual measurements before fresh execution. This stage is not an installed result or a qualification claim.

## Canonical source measurement

|Measure|Actual|Growth from merged parent|Planning growth stop|Absolute planning stop|
|---|---:|---:|---:|---:|
|Manifest bytes M|620090|131810|162106|650386|
|Manifest JSON nodes N|16373|3354|4256|17275|
|Coverage bytes C|88043|17381|21186|91848|
|New authored assertions A|333|333|446|—|

Manifest SHA256 `97d90c968b107b2ca1fe2191c7d2c3be168c7c52bc530580f9d8b6278b272453`; coverage SHA256 `dafedd035ca20f684559c739ebf2e6155f8446ce8062c3fa60ff86967b259bbd`. These are repository-normal Biome formatting, not compact source JSON. The explicit two-space JSON derivative renderings below are conservative size measurements; actual existing runner derivative bytes use JSON.stringify as before, recorded separately rather than used to fit a source budget.

Proposed binding growth guards: manifest135906 from488280; nodes3418 from13019; coverage18405 from70662. They leave actual+margin4096B/64nodes/1024B, within unchanged planning stops. The independent absolute envelopes remain650386/17275/91848; for this baseline the growth guards are tighter. Source comments call these proposed stage1 margins pending root acceptance. Canonical exact counts:21 new/187 total actions,333 new/1872 total authored assertions. Prior no-op15actions/223assertions remain unchanged. Full per-action assertion/capture ledger and ordered coverage mirror are in [preflight-ledger.json](preflight-ledger.json).

## Preservation and authored behavior

All eleven prior manifest objects and eleven coverage objects are byte-identical to parent. Removing the new row plus its separator reconstructs each entire parent artifact exactly. The machine ledger includes all22 hashes/lengths. Fixture stays716 bytes/SHA `3a7899ced771d4880066d98b0000c53f801ece53458502cc4f3bc641146c9a85`. All twelve coverage rows remain partial; claimEligible false. Canonical250 required/255 total selection remains unchanged, with238 required and5 optional unplanned.

The row independently seeds two Beads and opposite unowned Links. Both direction-specific failed batches attempt an update prefix, one Link deletion, Bead deletion and a later Link-deletion suffix. Exact failed receipt/Problem/index2 and full unchanged Resources, metadata, Event inventory and genesis group are pinned. The ordered successful batch removes both Links before the Bead; deleted identities/revisions, surviving full b, empty Links, exact404pos2,15 Event facts and retained genesis group with advanced head are pinned. Final scoped next URL capture is required and never followed. The new source oracle reconstructs all bodies/assertions/captures from separate constants, and three weakening tests check unchanged-Link equality, Problem index and rollback metadata. Source setup/applicability/cleanup are also pinned.

Aggregate bookkeeping is explicitly adapted:12 expected positive rows,19 scenario/reopen sessions,20 children including bootstrap;132 raw+48 client=180 actual exchanges and7 reopens make187 actions. New row21 exchanges includes exactly2 receipt clients and19 raw actions. Existing three no-op mutations/oracles are unchanged, with aggregate maxima adapted to20children/180exchanges.

## Derivative measurements

|Control|Added IDs|Removed IDs|Net assertions|Total assertions|2-space bytes / nodes|Existing execution bytes|
|---|---:|---:|---:|---:|---:|---:|
|incoming-order|23|0|23|1895|636790 / 16529|311314|
|outgoing-order|23|0|23|1895|636790 / 16529|311314|
|committed-prefix-state|19|8|11|1883|633776 / 16450|309751|

All derivative nodes/normal bytes fit the absolute planning stops. The machine ledger records every added/removed/assertion ID and both normal/execution manifest digests; the external derivative JSON files are exact generated variants of the app source function, not installed responses. Both successful-order controls retain the expected-failed disposition check, add independently passing completed/order/identity/revision/updated-b/required-effect-currentpos2 assertions, and classify the seven correlated original failed-envelope failures. Their exact modified input is authored from constants, with request-body checks required to pass.

Control3 removes all eight incompatible POST facts: serving-context, disposition, detail, required-position, no-effect, no-results, no-next and problem. It preserves seven status/schema/presence/request checks and adds nineteen positives, net11. Every POST assertion must pass; every preceding exchange/client assertion must pass; the existing next d-incoming-survives-beads GET must actually occur and be final. That unchanged GET must fail both complete-b and ordered-identities on actual changed state. No later Link GET is claimed. Its canonical GET object is deep-equal pinned in the static derivative test. No request/action/route is added by any derivative.

## Route and configuration boundaries

Exactly one new authored route: GET https://reference.invalid/acme/links/?limit=32, eighth in the twelve-route inventory. Both source digest pins use `1ba7cbbffc4fb40df2c3f12640bc636e70076cfa25bdd6952b83ec1bb1fe0035`. [route-preimage.json](route-preimage.json) contains the exact canonical five-field preimage (newline excluded from hash), independently authored then checked against first occurrence across actual actions. Credentials/ports are not included or invented.

Two separate source-configured caps are32: transaction maximumOperations (transaction-batch.ts:63) and finite Event receiptMaximumItems (transaction-store.ts:217, captured transaction-runtime-configuration.ts:294, used for Event count transaction-store.ts:1575). The unchanged app supplies no override. Stage1 records this source binding only; actual installed config/payload identities await execution. Four-operation carriers fit the first;15 Event facts and next:null depend on the latter and byte bound. Existing2KiB row POST/32KiB response checks include the new row. Runtime/parser/transport bounds are unchanged:63s positive observer,90s cohort,120s outer,5s exchange,10s lifecycle. Seven eventual cohorts remain max140children/1260exchanges. There have been zero installed attempts in this stage.

## Local validation and preserved preparation failures

Pinned Node24.16.0/pnpm11.20.0 frozen install completed. PASS: typecheck, build,38 cumulative tests,2 app-static tests (55 installed/other cases intentionally filtered), format, lint, boundaries, unchanged Read evidence verification and diff check. Static derivative construction invokes the genuine artifact factory but starts no service or process. All command exits/timings are retained in exits.json; JSON test outputs in cumulative.json/app-static.json.

First static/type/build attempt passed, but full format check rejected non-normal formatting in the newly inserted JSON. Its logs and pre-format artifact remain under first-validation/. Repository formatter then changed only new row whitespace; independent22-object preservation and complete parent reconstruction still pass. M decreased623056→620090; N/C/A unchanged; proposal measurements were updated explicitly. No semantic oracle was removed to fit. One exploratory source lookup used an absent historical action ID and raised StopIteration; it made no source change and was immediately replaced by actual inventory inspection. No failed installed execution or hidden retry exists.

## Remaining gate

Await root measured acceptance before installed package execution. Then the authorized next step is exact-source fresh pack with positive + three new controls + three adapted no-op controls, keeping all actual failures and cleanup. Runtime/fixture shape risks (deleted404context, opposite-Link creation, complete Link collection equality) remain unobserved in this new row until then; no weakening is authorized. No owned test/installed child remains. No further authoring is active.

Machine ledger SHA256 `548de5dae7b4f98a2d9a7d1f1f6e7ff008b2c745ba9af1e4253f2e7695d4929c`. Candidate source hashes and exact commands are in that ledger. Full implementation council and root integration remain separate later gates.
