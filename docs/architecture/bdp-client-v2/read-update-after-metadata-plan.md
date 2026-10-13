# After metadata closure: bounded R evidence and proposed N acceptance

Supporting plan only under the sole fork design.md. No implementation authorization, source edits, tests, network, servers, credentials or outward actions. Do not touch active `update_implementation` work. Implementation must await its B metadata successor being frozen, independently reviewed and accepted by root, with genuine clean-source both-shard evidence. No successor SHA is invented here.

Basis: frozen06b0e04a921c8501e36a8a68480cee72dc4c54b3 at `vickie-numeric-admission-review-repair`; whole-clause audit `read-update-numeric-admission-whole-audit.md` (root pin2782cfe6). Read-only active predecessor context: `vickie-numeric-metadata-closure` at06b with four dirty authorized files. Its current measurement is provisional, not evidence. Its A whole-credit decision is separately approved; B's ultimate flag remains root-owned. Preserve every accepted predecessor decision without resetting counts to06b.

## Recommendation

A cohesive R extension appears feasible without a runtime/parser/time-budget increase: rewrite one existing R numeric subflow per kind to carry nested values, insert eight metadata-refusal/afterstate actions reusing existing corrected requests, and add sixteen schema-descriptor/schema/comparator actions. R becomes96 actions; N stays52. R stays **wholeContract:false** pending a separate audit of all current selected clauses. Do not change the catalog/spec selection to hide the broad fourth clause.

Propose **N-only new whole-credit promotion** based on the accepted audit's four-clause map, subject to root approval and final-source fresh both-shard evidence. N needs no new execution cases. This does not promote R or borrow B/N executions to satisfy R.

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

Exact whole diagnostic location becomes `/value/items/0/a~1b~0c`; unapplied-patch message names `/propertiesChange/0/value/items/0/a~1b~0c`, with no instanceLocation for this configured implementation. Existing integer/overflow flows remain shallow. Independent static checks pin raw nested structure, escaped key and bad/good literal spellings; a local in-memory shortened-subtree mutation must fail that pin. Do not parse an inadmissible number to derive expectations. All prior IDs/actions remain, but this R block's bytes are intentionally changed and old R evidence is not relabeled.

## R change 2: whole metadata and metadataChange, eight extra actions

Use each existing integer subject; no additional Resource or route. Insert rather than append, so existing successful corrections remain the positive comparators:

1. After the current integer properties-create refusal/absence and before its corrected create, add a raw create with **admitted properties** (the corrected…992 corpus), metadata `{phase:"created",number:1e-400}` and a fresh key. Require actual422, authored bytes/complete write, normative Problem plus independent complete configured metadata diagnostic at `/number`; follow with the existing-kind point404. This new absence GET is its own action.
2. Change the existing corrected create metadata to `{phase:"created",number:0}`, with its existing fresh key. Bad/good metadata pair changes only the inadmissible leaf and key. Update independent raw-success/GET document facts and all later preimage-based oracles that carry this record.
3. Preserve existing integer properties-patch refusal with its full unchanged GET. Immediately before the existing corrected properties patch, add another raw patch: properties.value is the admitted **effectful…994**, metadataChange replaces phase with `updated` and number with0.12345678901234567891. Require exact configured422 metadata Problem naming `/metadataChange/1/value`, no postimage instanceLocation, and full independent unchanged GET equals the validated created preimage. Fresh key.
4. Existing corrected patch keeps properties…994 and phaseupdated and adds the same number replacement with0.25 under its existing fresh key; independently assert both effectful documents in response/GET. This distinguishes dropping propertiesChange, rejecting all metadata, and partial commit.

Two raw refusals plus two client reads per kind=8 extra actions. All existing malformed-numeric properties requests retain their bad literals and fresh keys. Full source keys remain unique within R and disjoint from other rows except intentional unrelated retries; neither a refusal nor a correction reuses a key. Update exact metadata expectations, not just response status. Metadata admissibility is required by R's actual unqualified fourth clause; exact diagnostic prose/optional omission choice remains target-specific.

## R change 3: installed inherited-schema comparator, sixteen extra actions

Use unchanged `apps/bdptest/test-support/installed-schema-cohort.ts`: existing schema-bead-base/child and schema-link-base/child. Base requires integer n>=0; child requires boolean enabled and detail when enabled=true. No Type creation/update/lifecycle, schema installation change or new runtime policy.

Within R, GET all four descriptors and all four referenced schema documents over the actual Scope-bound exact raw HTTP transport (8 actions). Independently assert descriptor identity/name/describes/conformsTo/propertiesSchema plus full schema JSON (including identical retrieval/$id URI). Child Link endpoints require in-Scope Task endpoints and external:none; reuse R's validated nr-source/nr-target. Actual inherited base/child documents are R-local observations, not a pass borrowed from the existing baseline schema row.

For each kind add four raw actions (8 total):

- create fixed new ID `beads/nr-schema` or `links/nr-schema`, declare that kind's existing child Type, properties `{n:-1,enabled:true,detail:"ok"}`, valid metadata; fresh key. All numbers are admissible; only inherited minimum fails. Require422/validation-failed, actual authored body/complete write, and independently authored diagnostic fields proving base-Type/schema minimum enforcement: owning Type URI, document-relative `/n`, schemaLocation ending `schemas/schema-<kind>-base#/properties/n/minimum`. Pin the full configured diagnostic/problem shape from the reviewed schema evaluator behavior before freeze; do not derive expected bytes from the target. If the existing formatter yields nondeterministic detail, report it rather than weakening to code-only.
- raw point GET fixed ID→404 using existing scope binding plus fixed path; no new fixture binding.
- raw corrected create differs only n:-1→0 and fresh key, remains child-declared and satisfies its own enabled/detail rule. Require200/created and independent full semantic identity/Type/properties/metadata/no-provenance/Link endpoints; closed normative result schema.
- raw canonical GET validates normative current record plus the same independent facts. No raw object capture or invented opaque revision; response-to-GET opaque-version linkage is explicitly unobserved for these raw comparators, as for existing raw R corrections.

This is representative effective-schema enforcement, not all Type constraints/keywords or all write carriers. R stays partial until independent selected-clause review determines whether any concrete conjunct still lacks evidence. No automatic whole promotion is promised. Move the existing final two Task endpoint GETs after these schema actions, preserving their assertions, so endpoints are checked after the entire new row. No extra endpoint Resources.

## Execution, scope and sensitivity budget

Current R72actions=29raw/43client,24raw POST; new R96=49raw/47client,32raw POST. N remains52; R+N ceiling must be explicitly authorized as148 (R<=96/N<=52), replacing only the prior72/52/124 planning ledger. A/B ceilings/counts/control tables remain exactly the accepted metadata predecessor. New R maximum10 Resources (existing8+2schema); two new point GET routes plus8 descriptor/schema GET routes; continuation route inventory49→59 if predecessor still49. Derive exact ordered routes independently; do not infer the digest from the runner. Baseline exact-route digest unchanged; continuation exact-executor digest changes. Common fixture/configuration/policy/authentication stays identical; route partition identity is not conflated with common served identity.

Three new bounded R sensitivity cases, separate from existing R/N9 and A/B controls:

1. Corrupt only the nested whole-properties diagnostic pointer after its real422; require exactly complete-problem failure.
2. Remove metadata document label from the new metadata refusal; require exactly complete-problem failure.
3. Corrupt the inherited-schema diagnostic's schemaLocation to the child schema or wrong keyword while preserving status and schema-valid shape; require exactly the named schema-location/full-Problem assertion(s), fixed after authoring. One application each.

Use existing closed app-local raw response-control selection; add only the specific R action/key/URL/mode cases needed, recompute substituted bodyOctets, retain genuine request observations. The client afterstate and metadata/precision existing controls remain valid and unchanged; the added metadata preimages may require corresponding independent expected values, not relaxed failure sets. Pin actual table lengths: oldR/N9 plusnew3<=12; oldA/B plusmetadata-closure controls retain their exact accepted ceiling. No framework/helper/runner/API change. Static nested-span/control checks are not additional installed scenario sessions.

The genuine positive artifact set still has54 scenario sessions,2shards,129required/75unplanned. Three additional control runs each use one bounded existing R scenario/isolated target, not extra required rows. R focused30s,5s exchanges,90s aggregate,120s outer,240s packed supervisor unchanged. Frozen06b aggregate18.790s had71.210s remaining; this is historical support only. Accepted metadata parent's exact passing margin is the execution baseline. Stop/report any timeout or inadequate measured margin; do not extend deadlines.

## Real source headroom, including prior-slice fences

Read-only active metadata author source/audit currently measures continuation844551bytes/22146nodes (SHA d6ad2f7abf0556227a20c786bd88d30c2c6f67a5d922c6fc3fc5168678fceec5), coverage703313bytes (SHA10a1954c48b67e90a1afd3936b61f6586f557bd38cbe810d8ac07ac1f26a7cc1), fixture29707bytes unchanged. These are provisional until frozen. Against larger R/N ceilings995852/26286/714717/37247, room is151301 continuation bytes/4140nodes/11404coveragebytes/7540fixturebytes.

But accepted metadata closure also asserts growth from131 anchors823730/21614/698614 with deltas81920/2000/12288. Its tighter full-file ceilings are **905650bytes/23614nodes/710902coveragebytes**. Current remaining space there is **61099bytes/1468nodes/7589coveragebytes**. Keep these gates; merely quoting the larger R/N ceilings would be misleading.

Propose new-parent-anchored incremental fences **<=56KiB continuation,<=1400nodes,<=7KiB coverage including reasons/bindings, zero fixture/index growth,<=128 added assertion bindings**. At provisional values:<=901895bytes/23546nodes/710481coveragebytes, leaving3755bytes/68nodes/421coveragebytes within the tighter metadata fences. This is feasible but tight, especially nodes/reason text. Use concise new action IDs (schema/refusal IDs may be short without losing meaning) and canonical normal formatting; do not minify or remove oracles. Existing R precision/metadata edits' net growth counts in these increments too. R new schema reads can use exact authored whole-document equality to avoid redundant per-leaf assertions while retaining complete independent semantics.

These are planning estimates, not already-generated measurements. Before implementation root must accept exact increment fences and pending-parent measurements. If normal authored source exceeds even one fence, **stop before execution and report measured need**; do not assume a cap adjustment. A fallback is a separately approved two-step R extension, but this plan does not authorize that split or reserve extra space. No file/parser/node/container/depth limit increases.

## Four-file ownership and preservation

Only continuation manifest, RU coverage JSON, app RU matrix test, cumulative-artifact.test.ts. Fixture, index, baseline, numeric catalog test, normative catalog/spec, schemas, installed schema cohort, runtime/runner/transport, TX and B author files remain untouched by this planner and are read-only during this follow-on except the four approved successor paths.

In the eventual successor deliberately modify R's row only; A,C,B,N raw blocks equal accepted metadata parent. Preserve all54 scenario IDs/order/ownership; preserve R existing shallow integer/overflow and admitted corpus, all existing controls and all pre-existing coverage rows except R bindings/reason and explicitly approved N reason/flag. Reconcile R/N action pins, R raw exchange/POST counts, corpus pins and route inventory with exact frozen parent. Do not retain R's whole old-block hash pretending it did not change: replace that R-only preservation pin with exact unchanged-action/subflow pins and explicit new/rewritten-action ledger. Robust source-span extraction handles final N; N hash stays exact. Preserve prior B old73+8/final-block pins and metadata closure growth/assertion guards unchanged.

Add new named parent-source byte/node/coverage bases to app static tests alongside every prior fence. Fixture/index/baseline digests remain identical. Record full catalog digest map unchanged, independent ordered route digest and changed continuation/coverage/set identities. Verify new raw literals/source bytes, exact fresh keys and complete ordered bindings mechanically. No coverage association relabeling.

## N promotion and executable done-when

Root has not approved N promotion. Proposed durable reason cites the four exact selected obligations, actual N IDs/diagnostics/positive comparators and audit2782cfe6 with its full final hash verified at implementation. State configured instanceLocation omission, finite depths/numbers, no borrowed B/R credit and no public qualification. N raw row/action/assertion bytes stay unchanged.

Read accepted parent's actual reviewed-whole IDs, then insert **only N** in canonical catalog order after root approval. A remains accepted there; B remains whatever root has separately accepted. With parent A31/Bfalse, N gives32whole/97notwhole; if B also accepted before freeze, N gives33/96. In general finalWhole=parentWhole+1, finalNotWhole=129-finalWhole; authored54/unplanned75 stay. Never overwrite an intervening approved whole list. R remainsfalse.

Completion requires: accepted ordinary metadata parent and clean passing predecessor evidence; exact four-file diff; machine-checked preservation/clauses/bounds/routes; real R positive row and3new+all inherited controls; meaningful schema and nested rejection/correction afterstates; pinned pre-build typecheck/build, conformance/catalog/static/boundaries/format/lint/smokes/Readseal; final clean frozen **fresh both-shard packed run even for N-only coverage promotion**, exact genuine written child reports and required omission inventory; source/config/payload/route/fixture/coverage digests; no credential leaks, success on failed attempt, or owned-child leaks. Retain all failed attempts honestly. Independent council reviews actual source and each proposed whole-credit change. If N is approved only after an evidence run, coverage-only successor needs its own fresh final-source both-shard observation rather than relabeling the old report.

No new product decision is required for the input semantics. Budget acceptance, exact pending-parent reconciliation and N promotion are root/council gates. This bounded slice can improve every identified R evidence gap without changing law, but it does not promise that one successful finite run automatically closes R's entire broad catalog selection.
