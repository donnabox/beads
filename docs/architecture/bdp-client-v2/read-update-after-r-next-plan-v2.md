# After R: schema PropertyChange closure plan — council revision 2

**Plan only. Add eight actual observations to R: a schema-invalid PropertyChange, full unchanged GET, one-leaf admitted correction and independent current-state GET for each Resource kind. Keep R partial until fresh source-bound execution and independent clause review close the gap.** The earlier zero-action recommendation missed a distinct mutation form. Do not narrow the selected text or its per-operation resulting-state preamble.

This supporting plan remains under the sole fork design.md. It authorizes no implementation, promotion, product change or outward action. Root and council must accept the explicitly proposed authoring-budget amendments before implementation. No tests, installed processes, network actions or product-source changes were performed to prepare it.

## Preserved review and exact basis

Preserve `read-update-after-r-next-plan.md`, SHA256 `1bb7009c7fa172ceb861cc30268ccdcfacff00b8fcb00676a395da64f649799a`, and `read-update-after-r-claude-plan-review.txt` unchanged. That first revised plan remains preserved at SHA256 `df80a835ee8755c91a5a97674d20be0f3a2c3ab76953dcce12f88fbb9d9ab3f7`; its review is `read-update-after-r-revised-claude-review.txt`. All seven new dispositions are in `read-update-after-r-next-plan-v2-dispositions.md`.

Basis is accepted final R source **e08673227362119dfafe7a8eabe38a0a11125104**, tree `9e4c20413724218f0210db9f97a668cc34104e5a`. During this read-only work root ordinarily merged main into that checkout: its current head **b14b91b762f9d2e5f9eb3102fec70e02a0879d45** has exactly the same tree. `read-update-after-r-revised-basis.json` records the actual read head, artifact bytes, hashes and raw scenario spans. Implementation requires a separately assigned final accepted ordinary parent; do not invent that future SHA or assume an intervening source change is harmless.

`numeric-r-council-integration/report.md` and `ru-safe-summary.json` record exact e086 fresh packing: 27 selected tests passed, genuine two-shard54-row observation, 129 required/75 omitted,33 whole/96 not whole, R partial, all public claims false. Aggregate16.585611125s left73.414388875s of90; the27-test command took30.419s. Historical4578's209-test invocation took130.544s under240. These are measured histories, not guaranteed future durations. Earlier failures remain failures, not qualification.

## Selected contract and the precise gap

The full current R selection in **`packages/conformance/catalog/numeric-model-v1.json:38–68`** remains byte-identical:

| Selected clause | Existing R-local evidence retained | This increment |
| --- | --- | --- |
| `docs/specs/bdp.md:709–717`: exact-decimal nearest-binary64/shortest-round-trip admissibility | Shallow and nested refusal/correction partitions across both kinds and whole/change inputs | No replacement or narrowing |
| `bdp.md:714–717`:1e300/expanded,-0.0,…993,twenty-significant-digit examples | Existing raw spellings and independent admitted values | Preserve exact raw bytes |
| `bdp.md:2522–2529`: inadmissible properties whole/change at any depth; no changes; named diagnostic | Existing numeric refusal and full before/after records, including mixed metadata cases | Preserve every witness |
| `bdp.md:2658–2659`, within **2648–2664**: Type/properties-schema constraints and every numeric literal; validate **each operation's resulting state** | Metadata whole/change is observed. Child-declared inherited-schema validation is observed on create only | Add the missing resulting-schema validation on PropertyChange for both Bead and Link |
| `docs/design/requirements.md:109–115`, PROTO-016 | Existing row-local shallow/nested whole/change refusal evidence | No change |

A runtime that validates schemas only on create can currently pass R. No amount of numeric-input diagnostic evidence addresses this resulting-state gap. Conversely, we do not import every JSON Schema keyword, Type lifecycle, policy, endpoint rule, History or Transactional behavior. Neighboring `read-update.validation.type-contract` and `read-update.validation.diagnostics` are at **`packages/conformance/catalog/read-update-v1.json:1532–1572`**; this increment gives neither row credit and borrows neither row's execution.

The unchanged installed cohort (`apps/bdptest/test-support/installed-schema-cohort.ts:12–53`) gives each child Type a base requiring integer n>=0 and child rules requiring enabled plus detail when enabled=true. Both R schema resources already have n=0,enabled=true,detail="ok", metadata.phase="created". The Link connects the existing Task endpoints and is unowned. `packages/server/src/resource-evaluator.ts:250–296,798–812` applies properties/metadata patches and reports effective-contract diagnostics: line812 validates the post-patch document, and line281 identifies the diagnostic Type as the declaring base Type. Existing baseline `schema-update-bead-base` demonstrates the implementation seam; its outcome is implementation evidence only, not R credit.

## Exact observation sequence and existing capture seam

For each kind, keep the existing create-refusal/absence/create-correction sequence. Deliberately alter only two existing actions per kind:

1. `nr-<kind>-schema-create-corrected` retains every request byte/assertion and adds a capture of `/resource/id` to `nr-<kind>-schema.url`. This is a confined URL capture, supported by `runner.ts:2464–2486`; it is not a forbidden raw object capture.
2. Convert the existing `nr-<kind>-schema-read` into `installed.exact-request` client GET targeting that captured URL. Preserve all independent status/schema/identity/Type/properties/metadata/absence/Link-endpoint assertions with client `/body` paths, add `written`, and capture the full `/body` as `nr-<kind>-schema.before`. The runner captures only after those assertions pass (`runner.ts:1110–1129`). The configured exact-client adapter still sends a real credentialed HTTP request. No new assertion or transport API is needed.

Then append these four actions immediately after that anchored preimage:

| Action suffix after `nr-<kind>-schema` | Transport and independent oracle | Bead/Link assertion count |
| --- | --- | --- |
| `-patch-refused` | Raw POST `operations/update-<kind>`: resource ID plus propertiesChange replace `/n` with -1 and metadataChange replace `/phase` with "updated". Require actual422, authored-body agreement, complete write, normative Problem schema and full independently authored configured Problem |5/5|
| `-patch-unchanged` | Client GET of the same captured URL:200, complete write, current-record schema, **entire body equals validated preimage**, including properties, metadata, revision, Type and Link endpoints |4/4|
| `-patch-corrected` | Raw POST to the same operation and Resource; exactly -1→1 is changed in the body, with a fresh distinct Idempotency-Key. Require200, authored body, complete write, normative result schema, outcome updated and independent identity/Type/full properties/full metadata; Link additionally has both independently expected endpoint tuples. Both corrected responses retain all four absence assertions: `no-owned-plane`, `no-provenance`, `no-source`, `no-source-revision` |13/15|
| `-updated-read` | Separate client GET of the same captured URL:200, complete write, normative record schema, independently authored identity/Type/properties `{n:1,enabled:true,detail:"ok"}` and metadata `{phase:"updated"}`, owned-plane/provenance absence, plus both Link endpoints |9/11|

The bad -1 is numerically admissible and satisfies integer type; it violates only the inherited minimum. The child constraints remain satisfied. Exact configured diagnostic: message `Schema assertion failed: minimum`, base Type `types/schema-<kind>-base`, schemaLocation `schemas/schema-<kind>-base#/properties/n/minimum`, instanceLocation `/n`, and the existing validation Problem type/status/code/retry. This is a failing **resulting document** location; do not substitute the request patch pointer or omit it by copying numeric-input diagnostics. Equality here intentionally pins this configured target's complete Problem, not a new closed-world protocol schema.

Raw refused/corrected bodies are157/156UTF-8 bytes, e.g. `{ "bead": "beads/nr-schema", "propertiesChange": [{"op":"replace","path":"/n","value":-1}], "metadataChange": [{"op":"replace","path":"/phase","value":"updated"}] }` without display whitespace in the actual authored body. No type, endpoint, resource, route or fixture additions. Final existing Task source/target full-record checks remain after these actions.

The semantic comparator map becomes20 entries, preserving every previous18 entry. New pairs require the exact contextual substitution `"path":"/n","value":-1`→`"path":"/n","value":1`; identical target and all non-key headers; equal numbers-stripped bodies; one changed leaf; fresh unique keys equal to the action IDs; refusal before correction. Preserve the repaired duplicate-comparator error discriminator and original numeric/nested companion guards.

## Independent countercontrols

Keep every inherited control definition unchanged. Add one separate four-entry schema-update table using only current source-private app seams:

- Actual Bead update422: replace the diagnostic's base schemaLocation with the child location. Exactly `complete-problem` must fail, with one mutation and the actual original422 recorded. Reuse the existing raw `schema-location` seam, selecting the new update action/operation.
- Actual Link unchanged GET200 after refusal: corrupt `/body/metadata/phase` to "updated". Exactly `postimage` must fail. The preimage was independently validated before capture.
- Actual Bead corrected-state GET200: corrupt `/body/properties/n` from1 to0. Exactly `properties` must fail.
- Actual Link corrected-state GET200: corrupt `/body/metadata/phase` from "updated" to "created". Exactly `metadata` must fail.

The last three use existing `RetryControl` action/scenario/input-bound actual-client observation mutation. Original observations must first have the expected200/values; count exactly one application; verify no unrelated failed assertion and complete cleanup. These mutations prove assertion sensitivity after real HTTP execution, not transparent packet forwarding or target mutations. They do not substitute for successful positive runs. Static guards additionally reject an omitted/duplicate new pair or a comparator whose metadata companion differs; no extra installed sessions for those guards.

## Exact projected budget and explicit amendment request

`after-r-schema-budget-v2.py` is a **planning-only arithmetic projection**, reading the accepted source and emitting metrics, not an executable manifest. It changes no product file and executes no tests. `read-update-after-r-schema-budget-v2.json` records the final eight-action candidate. The original six-action/raw-GET calculations and first reviewed candidate remain preserved; only this version2 candidate is proposed. Its `convertedActions` and `insertedActions` ledgers expose every byte/node contribution, including insertion separators, plus the exact derived binding split. Its JSON object/array/value node count matches the existing source-test definition. Normal expanded action formatting is used, and untouched source spans remain untouched. The eventual formatter's actual bytes must be measured again before execution.

| Quantity | Accepted e086 | Candidate | Effective inherited ceiling / remaining |
| --- | ---: | ---: | --- |
| Continuation bytes |878548|**899400** (+20852)|901895;2495 spare |
| Continuation JSON nodes |22972|**23502** (+530)|23546;44 spare |
| Coverage bytes |708995|**712792** (+3797)|710583; **over2209** before amendment |
| New mirrored binding lines |—|**68 /3786bytes**|Exact count and serialized byte sum must be pinned |
| R actions/assertions |88/619|**96/687**|Deliberately reconcile reviewed count pins |
| R raw/client observations |41/47|**43/53**|36 raw POSTs; existing32+4 |
| R+N actions/assertions |140/992|**148/1060**|**Action ceiling140→148 explicitly proposed**; assertion ceiling1300 unchanged (240 spare); N52/373 unchanged |
| R Resources/routes |10/unchanged|10/unchanged|No additions |
| Fixture bytes/nodes |29707/764|unchanged|37247bytes and existing parser limits |

Candidate reason, included in those bytes: “Partial pending council: packages/conformance/catalog/numeric-model-v1.json:38-68; row-local whole/change numeric, metadata and inherited-schema refusals, full unchanged preimages, one-leaf corrections and independent current-state GETs. Configured Read+Update finite partitions; raw correction response-to-GET revision linkage unobserved; public admission closed.”

**Request exactly three authoring-policy amendments, not runtime/parser changes:**

- Newest coverage fence703415+7168=710583 becomes703415+9728=**713143**: explicit **+2560bytes**, leaving351bytes after the complete candidate. The prior +2048 proposal fails by161bytes after restoring all eight absence assertions.
- Older metadata coverage fence698614+12288=710902 becomes698614+14336=**712950**: explicit **+2048bytes**, leaving158bytes. Broad714717 coverage cap remains unchanged. These are the smallest512-byte-rounded increases above measured overages2209/1890, with modest normal-format headroom; actual formatting must still pass before execution.
- R+N action ceiling in `read-update-matrix.test.ts:4336` changes **140→148**, exactly eight additional actions. The assertion ceiling1300 at line4337 remains unchanged: candidate1060 leaves240. No other action/resource/session ceiling increases.

The live old16-`addedIds` ledger naturally becomes **100 bindings/5546bytes** (the original98/5460 plus both new schema-read `written` assertions,2/86). The new eight-action `schemaPatchIds` ledger has **66 bindings/3700bytes** after restoring the quartet on both corrected responses. Reconcile exact guards to those derived sets; do not pretend the old live ledger remains98. The new bindings relative to the parent are68/3786 in total. Reason accounting becomes `coverage.length - 703415 - 5546 - 3700 <= 1708`; candidate remainder131. The parent98 IDs remain a preserved subset, but the live100 set is the current assertion target.

All other inherited fences remain unchanged: newest continuation901895/23546; older metadata905650/23614 (candidate leaves6250bytes/112nodes); original numeric limits995852/26286 (96452bytes/2784nodes); additional27158node guard; broader coverage714717 (1925bytes spare); fixture37247 (7540spare). Baseline manifest stays3626922bytes/97116nodes. The two ordinary manifests remain separately below4MiB/100000nodes; no parser depth/container allowance changes. No minification, shortened opaque action names, removed old witnesses or new shard is proposed to avoid the measured coverage overrun.

These three amendments change **test authoring-growth policy only**, not a public format, parser, runtime, server policy, resource limit, timeout or admission gate. Root requested measuring this variant rather than weakening current-state observation. They still need explicit council/root approval. Without that approval, stop with R partial; do not implement the eight-action candidate under failing old fences.

## Auditable source projection

All figures below are UTF-8 bytes with normal expanded action formatting. Added-action byte counts include the10-byte comma/newline/eight-space insertion separator; converted byte deltas compare exact parent spans. The machine ledger records old/new absolute sizes as well.

| Action (`nr-` prefix) | Byte delta | Node delta | Assertion delta |
| --- | ---: | ---: | ---: |
| bead-schema-create-corrected capture |208|5|0|
| bead-schema-read conversion |550|12|1|
| link-schema-create-corrected capture |208|5|0|
| link-schema-read conversion |560|12|1|
| bead-schema-patch-refused |2239|51|5|
| bead-schema-patch-unchanged |1163|32|4|
| bead-schema-patch-corrected |3410|89|13|
| bead-schema-updated-read |2347|64|9|
| link-schema-patch-refused |2239|51|5|
| link-schema-patch-unchanged |1163|32|4|
| link-schema-patch-corrected |3918|101|15|
| link-schema-updated-read |2847|76|11|
| **Total** |**20852**|**530**|**68**|

Conversions total1526bytes/34nodes/2assertions; additions19326bytes/496nodes/66assertions. Thus878548+20852=899400 and22972+530=23502. Coverage increment3786 new-binding bytes+11 reason bytes=3797;708995+3797=712792. The plan review's suggested683 assertions/64 new bindings was arithmetically inconsistent: four absences restored on **each** of two responses add8, so679→**687**,60→**68**; corrected counts9/11→**13/15**. No product projection artifact was written or executed.

## Ownership, preservation and exact bindings

Execution-only increment needs three product files:

1. `packages/conformance/matrices/read-update-reference-continuation-v1.json`: R only; four existing action conversions/captures plus eight additions.
2. `packages/conformance/test-support/read-update-coverage-v1.json`: only R's ordered complete binding list and honest reason; whole flag staysfalse.
3. `apps/bdptest/src/read-update-matrix.test.ts`: new control table and static preservation/budget/map/count guards; preserve root's diagnostics and all existing controls.

`packages/conformance/src/cumulative-artifact.test.ts` is reserved as the fourth path **only if root later explicitly authorizes R promotion**. No current execution change needs its whole-ID allowlist changed. Catalogs/specs/fixtures/index/runtime/executor/runner/schema cohort and TX files are outside ownership.

Machine checks must prove: all non-R scenario raw spans identical; entire baseline, fixture, index and all catalogs identical; all84 untouched R actions identical; only the named four old actions changed; eight new actions and68 new assertion bindings exactly match the manifest; all old assertion IDs remain; every old request body/key unchanged; old18 map entries preserved plus exactly2; all old control tables unchanged. Replace the live R tuple value at `read-update-matrix.test.ts:4207` with the independently computed new entire R span digest. Retain reviewed-parent12982a61faf2f6710ad165d69022ba2e8c9cc083e599118910a173c52fb9aa6c as a **non-asserting adjacent comment**, never as a second live expectation for the changed span. Use the robust final-block source-span extractor; no blind hash refresh.

Introduce exactly eight ordered `schemaPatchIds` and exclude them from the old `unchanged` partition: `!changedIds.includes(id) && !addedIds.includes(id) && !schemaPatchIds.includes(id)`. Preserve the existing24 changed IDs,16 added IDs, literal48 unchanged IDs and digest3378857a12c4632d66354e62f044dc5db27708133d5d1fdb3f522a0323d1b5e5. The four converted actions remain in old `addedIds`; the eight new actions must never fall into the48 unchanged set. Pin the100/5546 and66/3700 ledgers above independently, their order and total. Reconcile R exact pins88→96,619→687, raw/client41/47→43/53, raw POST32→36, map18→20, and combined actions140→148/assertions992→1060. Preserve1300 assertion ceiling and all unrelated count guards.


Route inventory should remain identical: existing raw `schema-absent` GETs already install both schema Resource routes; both update operations already exist. Derive and assert the full exact transport configuration including Scope/profile/routes/header/body limits, not routes alone. Expected continuation digest remains272c31e6…; baseline3224c31c…. If order/digest actually differs, inspect and report the exact preimage rather than refreshing a literal blindly. Common runtime identity1720b63… and payload1b12629… should remain unchanged because no runtime/configuration input changes; verify actual installed output. Manifest/coverage/set/result digests and clean launcher source pin legitimately change and must bind each new report. Git source is launcher metadata, not target attestation.

## Residual, timing and completion gates

Raw authored-byte corrections still lack explicit opaque **response revision→subsequent GET revision** equality. The new independent GETs establish stored semantic facts, not that token relation. In contrast, existing client Task endpoint creates do capture the response Resource and compare the subsequent whole record including revision; failed-update full preimage comparisons also include revision. Those distinctions are explicit. The raw revision linkage residual is outside R's current five selected clauses, as council confirmed; the schema PropertyChange gap is inside and cannot be dismissed as a residual.

No aggregate sessions are added: the existing R scenario performs eight more requests in its existing isolated service session. Four focused controls add at most four scenario+bootstrap pairs. Keep5s request,30s focused,90s aggregate,120s outer and240s supervisor bounds; retain both-shard common cancellation/cleanup and real failure reports. Historical e08673.414s aggregate headroom is substantial but not a forecast. Measure final run and stop/report on deadline or child-cleanup failure; do not raise deadlines.

Done when: root/council accept this plan and the three explicit authoring amendments; an exact accepted ordinary parent is assigned; generated source passes **every** old/unamended and explicitly amended fence before execution; all preservation/map/mirror/route/configuration guards pass; clean pre-build typecheck/build/static/boundary/Read-seal checks pass; positive R execution records96actions/687assertions and correct current-state facts; all four new and inherited controls fail only their intended assertions; fresh clean-source packing runs both genuine shards with54authored/129required, preserved75omissions, truthful33whole/96notwhole and Rfalse; genuine child reports/digests/cleanup are retained; independent code/selected-clause council accepts the result.

Only then may root separately authorize a concise coverage-only R promotion to34whole/95notwhole and the fourth-file allowlist change, followed by fresh evidence bound to that final coverage/source. No automatic promotion or public conformance claim is part of this implementation proposal. There is no unresolved protocol/product ruling: the remaining decisions are bounded authoring-budget acceptance and evidence/council closure.
