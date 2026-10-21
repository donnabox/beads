# Depth code review findings1,2,6 — source audit

No actionable blocker was established for these assigned findings. Keep the existing strict assertions. This closes supplied-context gaps with actual source; it is not an installed execution or final council disposition.

## 1. beadRecord/currentBeadRecord: definition exists; fixture oracle closes the difference

`schemas/bdp-v0.schema.json:774–815` defines beadRecord, so the proposed missing-ref/module-evaluation failure is disproved. It requires id/type/revision/properties; metadata is allowed but optional. `:881–910` defines currentBeadRecord as beadRecord plus required metadata and stronger current-Link validation when links/ownedLinks are present. They are not generally equivalent schemas.

The actual installed fixture has a stronger independent whole-record oracle: `apps/bdptest/src/transactional-depth-key-order-installed.test.ts:482–506` validates the receipt and requires its Resource to strictly equal exactly id/type/nonempty actual revision/properties:{title:name}/metadata:{}, with no links or ownedLinks. At573–577 each point response is status200, application/json, checked current context, beadRecord-valid, then strictly equal to that already independently checked Resource. Thus missing metadata, changed metadata, extra links/ownedLinks/provenance, changed properties or revision cannot pass. The final currentBeadCollection schema at581–592 additionally validates all four current Resources and strictly equals savedResources. There is one point-read callsite inside the two-carrier loop, not two distinct code callsites.

Disposition recommendation: F1 is not a blocker or an effective weakening for this exact fixture. Do not infer general beadRecord=currentBeadRecord. No assertion removal, fallback or source correction is necessary to close the stated concern; root owns any optional schema-name alignment.

## 2. Header and absent-body contracts are exact, not ambiguous

- `packages/server/src/transaction-http.ts:619–620` applies non-attesting to every handled response and sets `vary: Accept` on every POST after both route returns and caught errors. Therefore private bodyless500 and409 pass through the same POST decorator. There are ten POSTs in this witness, not six.
- Cache control comes from all actual finite families used here: empty500 at HTTP79–87 sets private,no-store; problem at89–105 delegates to jsonResponse; receiptResponse175–190 delegates to jsonResponse; readRoute.perform219–239 delegates successful point/collection responses to scopeDataResponse. `read-http-routing.ts:44–75` sets `cache-control: private, no-store` in jsonResponse and preserves it in scopeDataResponse. GET/HEAD decoration at HTTP626–629 uses `read-http.ts:10–48`, which clones/preserves existing headers and does not replace cache-control. The separate SSE no-store,no-transform path is not selected by any of these13 application/json/text/plain requests.
- `packages/conformance/src/http-executor.ts:48–50` declares exact request bodyOctets as required number; actual serializer construction at756–764 sets `bodyPresent: body !== undefined`, `bodyOctets: body?.length ?? 0`, and omits bodyDigest only when body is absent. Absent GET therefore correctly has false/0/undefined. The installed assertions at440–447 pin that exact convention and also require serializer source/writeState. Using `bodyOctets ?? 0` would silently accept a missing required field and weaken this oracle; it should not be applied.

Disposition recommendation: all three alleged blockers are missing packet context, now resolved by source. Keep the blanket checks for this finite witness and the strict zero assertion; no response-table expansion or permissive normalization is needed.

## 6. Exports and completed-attempt cleanup

`packages/server/package.json:10–19` exposes only the root and ./development entrypoints. Full export-declaration search of index.ts (17–24, public declarations throughout,1491–1497) and development.ts (6–10) shows no wildcard re-export and no transaction-batch reference or new accessor/receiving export. Both new names remain internal module exports; no package-export surface was added. This is source/package-map inspection, not an executed package import test.

`transaction-submission.test.ts:1682–1735` executes the pending claim, clears pending, completes both sequence turns, checks saved replay, and retains attempt for unconditional finally cleanup. `transaction-store.ts:1342–1346` explicitly accepts a genuine completed attempt: after brand/state and non-reentrancy checks, `state.cancelled || state.next === state.sequence.members.length` returns0. Successful nextSequence increments next, removes creator facts at terminal completion, and clears busy in finally (`:1969–1990`). The successful path is therefore a documented no-op branch, not accidental tolerance inferred from a passing run. Keeping the handle also guarantees cleanup if an assertion or turn fails early. Clearing it after the completed checks would be optional style only; there is no demonstrated cleanup defect or required repair.

The other informational F6 notes do not contradict this audit: exact class/message is retained despite a different throw site, and internal no-controls depth protection is defensive for genuinely strict sequence members. Root retains their final disposition.

## Basis and limits

Read-only file/search/JSON inspection; no product import, schema compilation, tests, service, build, source change or new Claude review. The review itself was read in full; only findings1,2 and6 export/cleanup were assigned for this audit. Current eight source hashes still match the frozen lint-repair manifest. Installed execution and its independent package/provenance/cleanup gate remain necessary; source proof is not substituted for that result.

Frozen source manifest `5fbe5c9b0c57f2c253fb16d114675ccb4929a9e4e18b21a1efbf1ca18385d778`; full patch `b404308f83bc8faf8fc94be8883daed0454c051f86353a97f903dcf3597e83df`; actual Claude stdout `d0ca412ec71bc03160af98fbaefcd7e426ce5e268cee1cf954c4e9b7a5566dd7`.

| Source | SHA256 |
|---|---|
| `schemas/bdp-v0.schema.json` | `0baa5d5ef67144f8e6513b4cde31168effc8df6d52ef3141b0f144b87f8dc227` |
| `packages/server/src/transaction-http.ts` | `b51d02801e5924190aba77c377e538db0b8619149d6184e19cb56d440f40c2cf` |
| `packages/server/src/read-http-routing.ts` | `d9344b87838f8c71381ccf9efe28b4f023df45503570beb8fb4293d553caca42` |
| `packages/server/src/read-http.ts` | `31934105eff6fff8ded8ce2ca9e3c19028991ca7c3d5f51b848004177e8c9268` |
| `packages/conformance/src/http-executor.ts` | `770a8405ae1532484b5a2658619c067832644947888f95e0c6575002376f1b5d` |
| `packages/server/package.json` | `110e00c403c667d8c848c6280b2e98c47f98f134158a621967d889914d44632e` |
| `packages/server/src/index.ts` | `34f9d9de8b48852c64919e69ce93f149e9cdae5142c15c75e4c35f08f15efc44` |
| `packages/server/src/development.ts` | `a3f994bd610f17821c7d51189104308abef6196bb12b565b268a0ffc96b04f39` |
| `packages/server/src/transaction-store.ts` | `fcd6882b29e538db45867c894c06dfbaa6ca588db0c2e403ef1b7abb14f6c706` |
| `packages/server/src/transaction-submission.test.ts` | `86342bc18126965f380d0e09218481eceac4a8396afb4543dd7e11d464fd1976` |
| `apps/bdptest/src/transactional-depth-key-order-installed.test.ts` | `48c6a84c84e470c7ce79f4d978bb09e2099daa5df4b4a6ef8629e1b7bf1b1700` |
