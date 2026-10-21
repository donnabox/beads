# After metadata closure: bounded R evidence and conditional B/N acceptance

Revised draft; original c027010fc81f0ba18f4fcd1c009cc38b1a18f7a3d3b2bb3054c343d205837fed is preserved unchanged. Dispositions are in `read-update-after-metadata-plan-dispositions.md`.

Supporting plan only under the sole fork design.md. No implementation authorization, source edits, tests, network, servers, credentials or outward actions. Do not touch active `update_implementation` work. Implementation must await its B metadata successor being frozen, independently reviewed and accepted by root, with genuine clean-source both-shard evidence. No successor SHA is invented here.

Basis: frozen06b0e04a921c8501e36a8a68480cee72dc4c54b3 at `vickie-numeric-admission-review-repair`; whole-clause audit `read-update-numeric-admission-whole-audit.md` (root pin2782cfe6). Exact measured metadata candidate is frozen4d19a6a8a749b8f21cf96070a66e1c85f1528f1b, ordinarily integrated at8873facfe0d2e530c51953275eb2643ace4e12b7. Metadata LOW repair is active from8873 in `vickie-metadata-review-repair`; **root has not assigned the eventual authoring base**. Require that repair’s frozen acceptance and final exact-source evidence before root supplies the ordinary successor parent. A whole is already accepted; `read-update-numeric-whole-root-acceptance.md` now conditionally accepts B and N in the next reviewed coverage successor. Do not invent a repair SHA or reset inherited decisions to06b.

## Recommendation

A cohesive R extension appears feasible without a runtime/parser/time-budget increase: rewrite one existing R numeric subflow per kind to carry nested values, insert eight metadata-refusal/afterstate actions reusing existing corrected requests, and add eight schema-enforcement comparator actions. R becomes88 actions; N stays52. R stays **wholeContract:false** pending a separate audit of all current selected clauses. Do not change the catalog/spec selection to hide the broad fourth clause.

Include **only the conditionally root-approved B and N whole-credit changes**, with exact reason/flag/canonical allowlist/count updates after metadata council closure and final-source fresh both-shard evidence. N needs no new cases; B’s accepted81 actions already close its missing plane. B/N raw scenario bytes and all their existing controls are immutable in this slice. R stays partial and may not borrow B/N executions.

The tightest fences are the preceding B slice's full-file growth assertions, not the larger R/N/global parser caps. The plan below keeps both. Authoring must stop if actual expansion does not fit; it may not silently re-anchor or weaken the earlier fences. Root can then choose a separately reviewed split, not an automatic cap increase.

## Required contract and R-local observations

Catalog `packages/conformance/catalog/numeric-model-v1.json:38–68` contains five selected requirements, unchanged:

- Exact decimal round-trip admission and named admitted/refused spellings (`bdp.md:694–723`). Existing integer/precision/overflow classes and corrected raw literals remain.
- Properties supplied whole or via Property Change, at any depth, with validation-failed/named diagnostic and no change (`bdp.md:2522–2529`, PROTO-016). R currently has only shallow bad numbers; add real nested representatives in both forms.
- Literal fourth selectedText: **Type and properties-schema constraints, and admissibility of every number literal** (`bdp.md:2658–2659`). Retain it. Add R-local metadata whole/change refusals and a representative inherited effective-schema failure/admitted comparator. Wire-schema validation is not treated as enforcement of an effective properties schema.
- PROTO-016's nested refusal repeats the properties obligation; do not import every neighboring policy/authorization/concurrency clause or every JSON Schema keyword.

N's assigned four clauses and current positive/negative evidence are mapped in the audit. Finite depth16, object/array placement, both delivery forms and both Resource kinds, exact named diagnostics, lawful absence of unapplied-patch instanceLocation, metadata contrast and positive corrections support N for the configured target. No universal mathematical proof, sequence/replay test or every depth is newly demanded by that selected contract.

## R change 1: nested precision inputs, no extra actions

Deliberately revise only `nr-{bead,link}-precision-*` document/patch values and dependent expected records: replace the scalar `properties.value` with `{items:[{"a/b~c":<precision literal>}]}`. Whole refusal retains0.12345678901234567891; correction retains0.125. Patch replaces `/value` with the same object/array shape containing refused literal, then corrected0.25. Preserve all admitted spelling companions1.0/-0.0/1e300/expanded1e300 and all current afterstate/endpoint assertions.

Replace the exact scalar precision corpus pins at `read-update-matrix.test.ts:4104–4107` with full contextual `"value":{"items":[{"a/b~c":<literal>}]}` structure pins (inside `properties` for create or the `/value` patch operation). Require the correct bad/create/patch literal at that leaf; a bare numeric substring is insufficient. Integer and overflow retain their existing scalar `"value":<literal>` pins.

Exact whole diagnostic location becomes `/value/items/0/a~1b~0c`; unapplied-patch message names `/propertiesChange/0/value/items/0/a~1b~0c`, with no instanceLocation for this configured implementation. Existing integer/overflow flows remain shallow. Independent static checks pin raw nested structure, escaped key and bad/good literal spellings; a local in-memory shortened-subtree mutation must fail that pin. Do not parse an inadmissible number to derive expectations. All prior IDs/actions remain, but this R block's bytes are intentionally changed and old R evidence is not relabeled.

## R change 2: whole metadata and metadataChange, eight extra actions

Use each existing integer subject; no additional Resource or route. Insert rather than append, so existing successful corrections remain the positive comparators:

1. After the current integer properties-create refusal/absence and before its corrected create, add a raw create with **admitted properties** (the corrected…992 corpus), metadata `{phase:"created",number:1e-400}` and a fresh key. Require actual422, authored bytes/complete write, normative Problem plus independent complete configured metadata diagnostic at `/number`; follow with the existing-kind point404. This new absence GET is its own action.
2. Change the existing corrected create metadata to `{phase:"created",number:0}`, with its existing fresh key. Bad/good metadata pair changes only the inadmissible leaf and key. Update independent raw-success/GET document facts and all later preimage-based oracles that carry this record.
3. Preserve existing integer properties-patch refusal with its full unchanged GET. Immediately before the existing corrected properties patch, add another raw patch: properties.value is the admitted **effectful…994**, metadataChange replaces phase with `updated` and number with0.12345678901234567891. Require exact configured422 metadata Problem naming `/metadataChange/1/value`, no postimage instanceLocation, and full independent unchanged GET equals the validated created preimage. Fresh key.
4. Existing corrected patch keeps properties…994 and phaseupdated and adds the same number replacement with0.25 under its existing fresh key; independently assert both effectful documents in response/GET. This distinguishes dropping propertiesChange, rejecting all metadata, and partial commit.

Two raw refusals plus two client reads per kind=8 extra actions. All existing malformed-numeric properties requests retain their bad literals and fresh keys. Full source keys remain unique within R and disjoint from other rows except intentional unrelated retries; neither a refusal nor a correction reuses a key. Update exact metadata expectations, not just response status. Metadata admissibility is required by R's actual unqualified fourth clause; exact diagnostic prose/optional omission choice remains target-specific.

### Explicit reused-comparator map and body-pair guard

For each kind, declare these exact pairings in the app’s R corpus guard rather than relying solely on `-refused`→`-corrected` suffix substitution:

| Refusal | Reused admitted action | Sole body difference allowed |
| --- | --- | --- |
| existing `nr-<kind>-integer-create-refused` | existing `nr-<kind>-integer-create-corrected` | properties.value …993→…992; **both** bodies carry valid metadata `{phase:"created",number:0}` |
| new `nr-<kind>-integer-meta-create-refused` | same existing `nr-<kind>-integer-create-corrected` | metadata.number1e-400→0; **both** bodies carry admitted properties…992 |
| existing `nr-<kind>-integer-patch-refused` | existing `nr-<kind>-integer-patch-corrected` | propertiesChange value …993→…994; **both** have phaseupdated and valid metadata number0.25 |
| new `nr-<kind>-integer-meta-patch-refused` | same existing `nr-<kind>-integer-patch-corrected` | metadataChange[1].value0.12345678901234567891→0.25; **both** have admitted effectful properties…994 |

These updates deliberately add the valid companion metadata to the original properties-refusal requests too, preserving their otherwise-identical corrected comparison. Every old bad properties literal and fresh key remains. Each corrected action executes once after both independent refusals/afterstates; mapping it twice does not create two observations or keys.

Keep the current numbers-stripped body equality assertion (`read-update-matrix.test.ts:4064–4066`) for every mapped pair as well as the stronger single-leaf check below. All raw POST refusal IDs, including all12 old R refusals, the4 new metadata refusals and2 new schema refusals, must equal the map’s18 distinct keys exactly; reject an unmapped refusal or duplicate key. Preserve every old refusal-to-corrected association. Reusing a corrected value is allowed, losing a refusal is not. N retains its current suffix mapping. This deliberately generalizes lookup authority rather than evading the `-refused` convention.

Assert each mapped pair exists exactly once, is raw POST to the same target with identical headers except its unique Idempotency-Key, and has the exact authored body equality after replacing **only the one explicitly identified numeric member literal**. Match contextual raw substrings at the named member, require one substitution, and never parse/round the inadmissible number. Do not drop the old semantic guard, ignore metadata, normalize every number indiscriminately, or accept arbitrary extra body differences. All preserved integer/overflow associations and revised precision pairings retain corresponding exact declared differences; nested precision pairs preserve their full object/array shape and only replace their scalar leaf. N’s existing pair guard/inputs are unchanged. The explicit map includes the new schema n:-1→0 pairings as well. Missing/duplicate pair, wrong route/key or an unintended other-document difference must fail a named static assertion. Include a bounded in-memory companion-metadata mutation that proves this guard detects it.

## R change 3: installed inherited-schema comparator, eight extra actions

Use unchanged `apps/bdptest/test-support/installed-schema-cohort.ts`: existing schema-bead-base/child and schema-link-base/child. Base requires integer n>=0; child requires boolean enabled and detail when enabled=true. No Type creation/update/lifecycle, schema installation change or new runtime policy.

Do not add descriptor or hosted schema GETs: the selected clause requires enforcement, and the deterministic refused/corrected/afterstate observations below directly test it. Installed schema documents/Type identities are independently authored from the unchanged reviewed cohort source, not fetched into an oracle. Hosted schema GETs are explicitly development-marked; this plan neither uses nor credits that surface. Descriptor retrieval remains its separate existing row. Child Link endpoint rules are in-Scope Task and external:none; reuse R’s independently validated nr-source/nr-target. No additional fixture binding or `schemas` route resolver is needed.

For each kind add four raw actions (8 total):

- create fixed new ID `beads/nr-schema` or `links/nr-schema`, declare that kind's existing child Type, properties `{n:-1,enabled:true,detail:"ok"}`, valid metadata; fresh key. All numbers are admissible; only inherited minimum fails. Require422/validation-failed, actual authored body/complete write, and independently authored diagnostic fields proving base-Type/schema minimum enforcement: owning Type URI, document-relative `/n`, schemaLocation ending `schemas/schema-<kind>-base#/properties/n/minimum`. Pin the complete configured Problem using the deterministic source-derived diagnostic `{message:"Schema assertion failed: minimum", type:"<scope>types/schema-<kind>-base", schemaLocation:"<scope>schemas/schema-<kind>-base#/properties/n/minimum", instanceLocation:"/n"}` (`installed-schema-evaluator.ts:1119–1137`, `1077`; `resource-evaluator.ts:192–210`, `277–282`). Never derive expected bytes from the target. No new nondeterminism is presumed.
- raw point GET fixed ID→404 using existing scope binding plus fixed path; no new fixture binding.
- raw corrected create differs only n:-1→0 and fresh key, remains child-declared and satisfies its own enabled/detail rule. Require200/created and independent full semantic identity/Type/properties/metadata/no-provenance/Link endpoints; closed normative result schema.
- raw canonical GET validates normative current record plus the same independent facts. No raw object capture or invented opaque revision; response-to-GET opaque-version linkage is explicitly unobserved for these raw comparators, as for existing raw R corrections.

This is representative effective-schema enforcement, not all Type constraints/keywords or all write carriers. R stays partial until independent selected-clause review determines whether any concrete conjunct still lacks evidence. No automatic whole promotion is promised. Move the existing final two Task endpoint GETs after these schema actions, preserving their assertions, so endpoints are checked after the entire new row. No extra endpoint Resources.

## Execution, scope and sensitivity budget

Current R72actions=29raw/43client,24raw POST; new R88=41raw/47client,32raw POST. N remains52; R+N ceiling must be explicitly authorized as140 (R=88/N=52), reconciling the prior72/52/124 ledger and the exact co-located pins enumerated below. A/B ceilings/counts/control tables remain exactly the accepted metadata predecessor. New R maximum10 Resources (existing8+2schema); two new point GET routes only (`beads/nr-schema`, `links/nr-schema`); continuation route inventory49→51 if predecessor still49. No descriptor/schema routes are added. Derive exact ordered routes independently; do not infer the digest from the runner. Baseline exact-route digest unchanged; continuation exact-executor digest changes. Common fixture/configuration/policy/authentication stays identical; route partition identity is not conflated with common served identity.

Three new bounded R sensitivity cases, separate from existing R/N9 and A/B controls:

1. Corrupt only the nested whole-properties diagnostic pointer after its real422; require exactly complete-problem failure.
2. Remove metadata document label from the new metadata refusal; require exactly complete-problem failure.
3. Corrupt the inherited-schema diagnostic's schemaLocation to the child schema or wrong keyword while preserving status and schema-valid shape; require exactly the named schema-location/full-Problem assertion(s), fixed after authoring. One application each.

Use existing closed app-local raw response-control selection; add only the specific R action/key/URL/mode cases needed, recompute substituted bodyOctets, retain genuine request observations. Put these3 new controls in a separate named R-extension table; preserve existing `numericAdmissionRawControls.length===6`, `numericAdmissionClientControls.length===3`, and combined<=9 assertions unchanged. Existing request-number controls must still fail their original four assertion IDs on the original intended action; later dependent failures may also occur, and are not a reason to weaken that discriminator. The client afterstate and metadata/precision existing controls remain valid and unchanged; the added metadata preimages may require corresponding independent expected values, not relaxed failure sets. Pin actual table lengths: oldR/N9 plusnew3<=12; oldA/B plusmetadata-closure controls retain their exact accepted ceiling. No framework/helper/runner/API change. Static nested-span/control checks are not additional installed scenario sessions.

### Exact ledger and source-cost reconciliation

At `read-update-matrix.test.ts:4011–4124`, update R’s exact action count72→88 and combined action ceiling124→140; N remains exactly52. The ordered created-Resource list retains the original eight IDs and appends `beads/nr-schema`, then `links/nr-schema` in that actual creation order. The existing global R/N fresh-key uniqueness assertion remains; R raw requests stay UTF-8 and each<4096bytes with exactly one authored-body assertion and complete-write. Keep combined assertion ceiling1300. Do not weaken fixed raw/corrected pair checks, nested lexical checks or admitted1.0/-0.0/exponent spellings. Update R-only source-span pin to an explicit changed-action ledger; all other scenario hashes stay fixed. Explicit Rfalse/Ntrue and Btrue expectations are conditional on the named root acceptance gates, not derived from counts.

The proposed **98 new assertion bindings** are auditable before authoring:

| Per-kind added action | Assertions | Independent purpose |
| --- | ---: | --- |
| `nr-<kind>-integer-meta-create-refused` | 5 | status/authored-body/complete-write/body-schema/complete-problem |
| `nr-<kind>-integer-meta-absent` | 4 | client status/written/body-schema/not-found |
| `nr-<kind>-integer-meta-patch-refused` | 5 | same five raw refusal assertions |
| `nr-<kind>-integer-meta-unchanged` | 4 | client status/written/body-schema/full validated preimage |
| `nr-<kind>-schema-create-refused` | 5 | same five raw refusal assertions, complete deterministic schema Problem |
| `nr-<kind>-schema-absent` | 3 | raw status/body-schema/not-found |
| `nr-<kind>-schema-create-corrected` | 13 bead /15 Link | raw status/authored-body/complete-write/body-schema/outcome; independent identity/Type/properties/metadata; owned-plane/provenance/source/sourceRevision absence; Link endpoint pair |
| `nr-<kind>-schema-read` | 8 bead /10 Link | raw status/body-schema; independent identity/Type/properties/metadata; owned-plane/provenance absence; Link endpoint pair |

Metadata36 +schema62=98. Existing R521 +98=619; N373 unchanged; combined992<1300. This assumes changed old actions retain their assertion cardinality; any extra needed oracle must be reconciled rather than silently added. Schema raw GET lacks capture, so the disclosed opaque-version linkage residual remains. All required bindings must still equal the exact flattened action/assertion inventory.

Read-only arithmetic using these exact new IDs and existing assertion suffixes, rendered as the coverage file’s12-space-indented JSON string lines plus comma/newline, gives **5460 added binding bytes**. That leaves **1708bytes within the unchanged7168byte coverage increment for the net R/B/N reason changes, two flags, and any other coverage syntax changes**. Existing R action/assertion IDs remain unchanged, so rewriting their expectations adds no old binding-line growth. Reasons must be concise but traceable; the full clause audit stays in logs. This is a concrete planned binding calculation, not a generated-source fit claim. Measure complete actual coverage against both7168increment and710902absolute fences; reason growth beyond1708 or any other expansion consumes the same budget. No implicit reserve beyond those figures. Raw manifest bytes/nodes retain the conservative56KiB/1400 estimates and must be measured before execution; no generated-source claim is made here.

The genuine positive artifact set still has54 scenario sessions,2shards,129required/75unplanned. Three additional control runs each use one bounded existing R scenario/isolated target, not extra required rows. R focused30s,5s exchanges,90s aggregate,120s outer,240s packed supervisor unchanged. Frozen4d aggregate21.146170417s had68.853829583s remaining; its204-test selected packed invocation took154.117s. Integrated8873 genuine RU aggregate19.49269275s had70.50730725s remaining. These are historical exact-source evidence, not the pending repair’s timings. Accepted metadata parent's exact passing margin is the execution baseline. Stop/report any timeout or inadequate measured margin; do not extend deadlines.

## Real source headroom, including prior-slice fences

Frozen4d metadata source/audit measures continuation844551bytes/22146nodes (SHA d6ad2f7abf0556227a20c786bd88d30c2c6f67a5d922c6fc3fc5168678fceec5), coverage703415bytes (SHA8274385bff08618e15ccbdb863127f5a472606970c14c735f4d9805f9ff01954), fixture29707bytes unchanged. These artifacts remain the reconciliation baseline for the pending LOW repair; remeasure the actual assigned parent. Against larger R/N ceilings995852/26286/714717/37247, room is151301 continuation bytes/4140nodes/11302coveragebytes/7540fixturebytes.

But accepted metadata closure also asserts growth from131 anchors823730/21614/698614 with deltas81920/2000/12288. Its tighter full-file ceilings are **905650bytes/23614nodes/710902coveragebytes**. Current remaining space there is **61099bytes/1468nodes/7487coveragebytes**. Keep these gates; merely quoting the larger R/N ceilings would be misleading.

Propose new-parent-anchored incremental fences **<=56KiB continuation,<=1400nodes,<=7KiB coverage including R plus B/N reasons/bindings/promotions, zero fixture/index growth,<=98 added assertion bindings**. At measured4d values:<=901895bytes/23546nodes/710583coveragebytes, leaving3755bytes/68nodes/319coveragebytes within the tighter metadata fences. The revised98-binding layout above fits the planned coverage increment on paper; final source bytes/nodes are still ungenerated. Coverage reasons are the tightest remaining planning constraint. Use concise new action IDs (schema/refusal IDs may be short without losing meaning) and canonical normal formatting; do not minify or remove oracles. Existing R precision/metadata edits' net growth counts in these increments too. Do not reintroduce the dropped schema/descriptor reads or delete oracles to fit. No assertion scheme extension is needed.

These are planning estimates, not already-generated measurements. Before implementation root must accept exact increment fences and pending-parent measurements. If normal authored source exceeds even one fence, **stop before execution and report measured need**; do not assume a cap adjustment. A fallback is a separately approved two-step R extension, but this plan does not authorize that split or reserve extra space. No file/parser/node/container/depth limit increases.

## Four-file ownership and preservation

Only continuation manifest, RU coverage JSON, app RU matrix test, cumulative-artifact.test.ts. Fixture, index, baseline, numeric catalog test, normative catalog/spec, schemas, installed schema cohort, runtime/runner/transport, TX and B author files remain untouched by this planner and are read-only during this follow-on except the four approved successor paths.

In the eventual successor deliberately modify R's row only; A,C,B,N raw blocks equal accepted metadata parent. Preserve all54 scenario IDs/order/ownership; preserve R integer/overflow bad/admitted numeric corpus and shallow placement (the four integer refusals deliberately gain matching valid companion metadata as specified above), all existing controls and all pre-existing coverage rows except R bindings/reason and conditionally approved B/N reasons/flags. Reconcile R/N action and false-flag pins, R raw exchange/POST counts, corpus pins and route inventory with exact frozen parent. Do not retain R's whole old-block hash pretending it did not change: replace that R-only preservation pin with exact unchanged-action/subflow pins and explicit new/rewritten-action ledger. Robust source-span extraction handles final N; N hash stays exact. Preserve prior B old73+8/final-block pins and metadata closure growth/assertion guards unchanged.

Add new named parent-source byte/node/coverage bases to app static tests alongside every prior fence. Fixture/index/baseline digests remain identical. Record full catalog digest map unchanged, independent ordered route digest and changed continuation/coverage/set identities. Verify new raw literals/source bytes, exact fresh keys and complete ordered bindings mechanically. No coverage association relabeling.

## Conditional B/N promotion and executable done-when

`read-update-numeric-whole-root-acceptance.md` conditionally approves B and N whole credit after current metadata code council closes and final-source genuine both-shard evidence passes. B audit is a0adb7f6b064c5f9bc2fe4b8818eefdd17a3a60e38f5c8ef87146b6c64532567; N audit is2782cfe6765be3eb5eb6ee81037cf15cf0f3dc2b4b167e1e418c32483f0cc99d. N reason cites its four selected obligations, actual N diagnostic/positive-comparator IDs, configured location omission and finite partitions. B reason cites its five selected obligations,81 row-local actions and newly closed metadata plane, and keeps the raw effectful response-to-GET opaque-version residual disclosed. Root ruled that residual outside B’s selected no-op clauses; do not falsely claim it was observed. Neither reason borrows R evidence or asserts public qualification.

Preserve A credit. Insert B and N once in the canonical whole-ID allowlist, never derive whole from test passes. Expected successor **129 required/54 authored/33 whole/96 not whole/75 omitted** from current31. If root’s assigned parent already contains either approved promotion, preserve it idempotently and add only the missing accepted ID; do not double-count. R remainsfalse. App R/N loop’s prior shared whole=false assertion becomes explicit Rfalse/Ntrue only after the acceptance gates; B’s corresponding false assertion/count expectations likewise follow the two named decisions. No other flag or scenario/control byte may change. The7KiB coverage-growth ceiling includes both final reasons and flags.

Completion requires: accepted ordinary metadata parent and clean passing predecessor evidence; exact four-file diff; machine-checked preservation/clauses/bounds/routes; real R positive row and3new+all inherited controls; meaningful schema and nested rejection/correction afterstates; pinned pre-build typecheck/build, conformance/catalog/static/boundaries/format/lint/smokes/Readseal; final clean frozen **fresh both-shard packed run including the B/N coverage promotion**, exact genuine written child reports and required omission inventory; source/config/payload/route/fixture/coverage digests; no credential leaks, success on failed attempt, or owned-child leaks. Retain all failed attempts honestly. Independent council reviews actual source and each proposed whole-credit change. If either promotion is incorporated only after an evidence run, coverage-only successor needs its own fresh final-source both-shard observation rather than relabeling the old report.

No new product decision is required for the input semantics. Budget acceptance, exact pending-parent reconciliation, metadata council closure and final-source evidence remain root/council gates; the B/N decision is conditional rather than automatically effective. This bounded slice can improve every identified R evidence gap without changing law, but it does not promise that one successful finite run automatically closes R's entire broad catalog selection.

## External council status

External c027 plan review completed and was read in full: `read-update-after-metadata-claude-plan-review.txt`. B1–B3/H1–H3/M1–M4 are dispositioned individually in the companion document. Root directed dropping eight retrieval actions and preserving the comparator invariant through a complete explicit map rather than adding four duplicate corrected POSTs. Neither the original plan nor external output was edited. This revised plan still awaits root/council acceptance; no implementation or promotion occurred.
