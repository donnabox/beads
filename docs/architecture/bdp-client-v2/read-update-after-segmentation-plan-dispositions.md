# Next Update evidence plan — council dispositions

Original reviewed plan is preserved byte-for-byte: `read-update-after-segmentation-plan.md`, SHA256 **f16a1e163e2ef55b228431fda42e280ef7cb8e32a89afb9b651572bfcd068a70**. Revised separate plan: `read-update-after-segmentation-plan-revised.md`, SHA256 **f262ee95350555a2b151a2c37cfc1c7c046753e4bd1eaec4de208dcc9093f664**.

Normative/runtime basis b919ab37326e259ff30304ab0dd642562a21ea40. Step-two harness inspection/probe basis 7d256e57114b2ddc920eb01388381e86fa2b8285. Both author checkouts remain immutable; product code unchanged. Implementation still requires root approval and the final repaired/integrated step-two parent. This disposition does not accept 7d's outstanding failure-retention gap or create a new implementation lane.

| External finding | Disposition and concrete resolution |
| --- | --- |
| F1 High: whole-record equality impossible | **Reject impossibility; accept need to name the seam.** `json-pointer` accepts empty pointer and `equalsBinding`. `runner.ts:2704–2705` returns the whole JSON body, and `:2231–2241` deep-compares it to the earlier arbitrary JSON binding. Capture a client GET's `/body` or mutation's `/body/resource`; compare later HTTP at `pointer:""` and client output at `/body` or `/body/resource`. Direct HTTP capture is URL-only (`:2464–2480`), so it must not be used for an object. A four-case actual v2 parser/runner probe passes the unchanged whole record and detects metadata, nested provenance and extra-field differences. No field enumeration, weakened oracle, json-equals extension, or normalization needed. |
| F2 High: digest control/report exposure inaccurate | **Accept and correct.** Each authored raw numeric request uses existing `request-authored-body` plus complete `request-write-state`. `runner.ts:1910–1934` computes the expected length/digest from authored bytes and checks the actual serializer observation. There is no authored expected-digest field. A test-only wrapper changes the real exchange's observed digest to another valid SHA256 to fail the named assertion. Report projection withholds raw digest absent the synthetic approval; this slice does not set that approval. Source is a checked/projected observation, not a new manifest assertion kind. Citations and evidence claim now distinguish checks from redacted output. |
| F3 High: continuation coverage breaks static inventory test | **Already resolved by inherited step two, verify before implementation.** 7d `cumulative-artifact.test.ts:103–126` builds the accepted set and flattens both shard manifests; `:187–203` checks status/bindings against that union. Native 7d review independently passed the 440-test/7-file focused suite including this test and real C inventory. No duplicate migration is needed. Preserve this generalization in the final repaired parent; root still owns any eventual exact whole-list promotion. |
| F4 Medium: no-Event clause approval omitted | **Accept.** Both new-row profile prefixes will observe actual closed `#/$defs/readUpdateDiscovery`, `profile:read-update`, and absent `/transactional`, `/events`, `/receipts`, `/changes`, `/snapshot`; retain direct result observations. Explicitly ask council to accept these as the selected RU profile-context evidence. No internal journal absence, TX no-event proof, or automatic whole credit; keep false if the applicable selected clause remains unobserved. |
| F5 Medium: reused guard keys could test conflict instead of CAS | **Accept.** Every mutation/no-op/guard/comparator/restore in both rows has a fresh literal key, including old/current guards. Count keys in the per-row action ledger. No deliberate retained retry is part of these candidates. |
| F6 Low: uncovered terminology | **Accept.** b919: 129 required = 30 whole + 19 observed partial + 80 omitted; “99” means rows without a whole flag. 7d adds C only: 30 whole + 20 partial + 79 omitted. Do not describe all 99 as unobserved. |
| F7 Low: index schema already known | **Accept.** Index is `{version,shards:[{id,source}]}` only, with no digest fields. No index edit or filename change in this slice; derived identity updates come from source bytes. |
| F8 Low: citation boundaries | **Accept.** Corrected both-document comparison to 862–866, evaluator test spans to 409–438, 1184–1227 and 1252–1315, raw executor observation range to 632–651/753–760, and actual runner authored-byte assertion versus projection references. |
| F9 Low: per-resource fixture bindings | **Accept.** Each row-owned Resource has a dedicated fixture binding, including views. Collection bindings remain reserved for collection queries and their complete-page oracles. No base-plus-path shortcut for Resource GETs. |
| F10 Low: denominator/report terminology | **Accept precision, retain real wrapper distinction.** The set owns one complete 129-ID coverage inventory. Genuine ConformanceRunResult objects retain their own requested/selected/results lists; baseline is unfiltered, continuation is its exact owned set. The app wrappers do carry full common requiredIds as association metadata and per-shard plannedIds. The revised plan names these layers separately instead of claiming both genuine runner reports select 129 IDs. |

The independent original native review had no actionable findings and accepted the finite numeric/provenance scope as a plan, not executable evidence. Its feasibility reasoning is retained; F1/F2 are now more concrete, and the RU no-Events/no-receipts interpretation is an explicit council acceptance condition. No novel product semantics are requested.

## Bounded seam probe

Disposable source copy: `/Users/dbox/repos/wt/vickie/vickie-artifact-set-execution-native` at 7d. No server, network, broad suite, or installed target was started. The probe appended four tests using existing runner-test helpers, parsed actual v2 manifests, created normal artifact bundles, and invoked the exact v2 runner with synthetic response observations. Each case first captured a complete client-action record and compared client postimage and HTTP root against it. Positive same-record passed; wrong metadata, changed nested provenance message, and extra top-level member each produced the intended `complete-root` assertion failure, making the corresponding sensitivity test pass.

Pinned command:

```sh
PATH=/Users/dbox/.nvm/versions/node/v24.16.0/bin:$PATH \
node /Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs exec vitest run \
  packages/conformance/src/runner.test.ts \
  -t 'native plan root-pointer binding seam' --maxWorkers=1 \
  --reporter=verbose --reporter=json \
  --outputFile.json=/Users/dbox/repos/wt/vickie/bdp-integration-logs/read-update-after-segmentation-seam-probe/v2-result.json
```

Actual final exit **0**, **4 passed**, **0 failed**, **246 unrelated tests unselected**. Logs/code/results: `read-update-after-segmentation-seam-probe/{probe-addition.ts,v2-run.log,v2-result.json,evidence.json}`. This is assertion-seam proof only, not evidence of new protocol behavior.

The first probe incorrectly used a direct HTTP object capture and failed with “capture 'prior-record' did not produce an absolute HTTP(S) URL.” Preserved `initial.log`/`initial.json` and debug log show that failure. Correcting the probe to the existing client capture arm then passed the shared runner comparison, followed by the final explicit v2 run above. No failure is relabeled as success. Original disposable test bytes were restored and Git status is clean; frozen b919/7d author source was never changed.

## Implementation gate

Proceed only after root accepts this revised plan and supplies the repaired step-two parent. Recheck unchanged seams and inventory generalization there, measure exact continuation/C/fixture budgets, preserve all 49 old blocks plus C and its false flag, and keep both candidate flags false until row-local observations and their targeted controls survive council. All prior parser/time/public admission limits remain unchanged. No outward writes occurred.
