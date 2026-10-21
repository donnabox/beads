# Next Transactional evidence slice: complete set effects and zero matches

Supporting plan only. Sole plan of record remains the fork's `docs/architecture/bdp-client-v2/design.md`. Read-only audit of frozen `9eb9339a0b25a5ce9c16e90c18010874bab8d571` at `/Users/dbox/repos/wt/vickie/vickie-transaction-set-conformance`; clean source unchanged. No tests, child authorities or source changes were run for this audit. Existing packed evidence below was read, not relabeled as a reviewer run. Implementation needs root acceptance and an exact integrated predecessor after current council dispositions.

## Recommendation and exact law

Add two small, still-partial required-row observations: `transactional.set.mutation`, then `transactional.set.zero-match`, in canonical catalog order. The useful increment is a complete two-member update/delete plus a late-selected-member rollback and actual zero-match receipts/Events. Use only ordinary Beads with the already installed Task Type; no new descriptors or Link ownership. Do not add duplicate scenarios merely to claim another row.

The current normative catalog states:

| Row | Existing source | Obligation observed |
| --- | --- | --- |
| `transactional.set.mutation` | `packages/conformance/catalog/transactional-v1.json:295`, spec `#set-mutation-objects` and `#mutation-receipt-responses` (bdp.md:4866,4905,5160–5167) | Entire matched set mutates atomically or fails; flat matched/count followed by one result per selected canonical ID in code-unit order. |
| `transactional.set.zero-match` | catalog:313, spec `#events-and-event-sources` (5586) and `#mutation-receipt-responses` | Zero produces no Events and exactly its matched/count0 result, nothing else. |

The mutation row's receipt order will also be corroborated by actual Scope Events in this slice, but **no credit is added to `transactional.event.set-expansion-order`** (catalog:1068). Its law also covers Link singleton expansion/incident facts; the narrow Bead witness does not justify an extra credited row without its own deliberate plan. `transactional.set.attribution-fanout` (catalog:362, spec:4250–4256) specifically includes fresh owned-source versions and deletion's lack of a new deleted version. Current app fixture uses unowned Task/Relationship descriptors, so that is a subsequent ownership-fixture slice rather than silently adding Type design here.

Existing evidence is complementary, not borrowed: six current rows observe atomic commit/rollback, cardinality, singleton set targets, readable receipts and retained failures. `packages/server/src/transaction-set.test.ts:125,161,184,216,235` already implements component witnesses for ordering, zero matches/no-op, provenance, late selected failure and Link expansion. Their passes do not substitute for installed conformance observations.

## Implementable seam and bounded scenarios

Reuse the existing raw HTTP actions and `transactional.receipt-read` dynamic observation. Existing exact canonical `batch`, `update-where`, `delete-where` receipt-authority suffixes suffice; no change to `transactional-observation.ts`, runner grammar, IPC, launcher or execution binding. Ordinary HTTP only captures actual URL values; full Resource/revision/transaction captures come from the existing client receipt-read projection. Every captured record must first have its authored ID/Type/properties/metadata pinned independently, then opaque revision fields can be reused to compare complete records. Point-GET affordance shape must follow the existing exact-record assertions, not a new projection.

### M — complete set, rollback, order and deletion

At most three live Beads, all existing Task Type. Seed in a deliberately noncanonical order (b,c,a), with literal IDs a/b/c, distinguishable metadata, a.properties={title:"A"}, b.properties={}, and c.properties={title:"untouched"}. The installed Task schema allows absent title; confirm that existing descriptor fact before authoring. No incidental Links. Seed receipt/GETs independently pin all fields; capture original full records/revisions from the actual validated receipt. Observe the exact three seed Events, next:null.

1. Submit update-where selecting a and b using the already exercised bounded valid ID-disjunction Selector. Its replace-/title patch is valid for a but fails for b's missing path. Observe the real failed receipt at operationIndex0, schema-valid existing patch failure, zero partial results, and its dynamic retained read. Use the actual existing normative patch failure mapping; do not invent a new Problem code or unsupported-selector classification. GET all three original records unchanged and assert the exact seed Event inventory/position unchanged. This discriminates a service that mutates a then fails b without rolling back.
2. With a fresh key, submit valid update-where for the same a/b population using an ordinary supported Property Change that sets the complete properties object to {title:"changed"}. Assert completed receipt, exactly matched2 then updated a,b in literal ascending code-unit order, operationIndex0, next:null, complete independently expected properties and each original metadata. GET both independently validated postimages and the unchanged nonmember c.
3. Observe Scope Events via existing literal `/events/`. Assert exact seed prefix plus exactly two updated subjects a,b in canonical expansion order, no c update, shared transaction identity captured from this command's actual receipt, and revisions tied to its validated postimages. Schema-check the Event page, next:null. Do not require arbitrary opaque Event IDs, timestamps or implementation-selected patch spelling. Event deltas must be present/schema-valid; this is not full replay/delta qualification.
4. Delete the same two selected Beads with a fresh delete-where key. Assert matched2 followed by deleted a,b, final live revisions equal the observed update postimages, next:null. Both point GETs404; c remains unchanged. Scope Events have exactly the prior prefix plus deleted a,b under this receipt's transaction identity. Read retained update and delete receipts via their own actual original POST binding; no fabricated URL from keys. Reopen once, reread both receipts with exact retained-body equality and advancing physical generation, and verify final point state.

Cap this row at32 actions. Maximum three matched/live Resources (actual selected count2), three seed Events plus2update+2delete=7 total; largest receipt4 entries for seed,3 for sets, all inline. One failing patch contributes no Event. No receipt paging, races or reader interleaving is claimed. Observing complete committed state/rollback does not prove all possible concurrent atomic visibility.

### Z — nonempty graph, empty selection

Separate fresh authority with one independently pinned Task a. Seed, capture full record, and observe exactly its one created Event. Submit valid update-where with a literal absent-ID predicate and a normal valid properties patch. Observe completed receipt with the exact one-element results array [{operationIndex:0,outcome:"matched",count:0}], next:null, no effectPosition, and serving epoch/view/position equal the seed/current point. Dynamic receipt GET validates original wire/body provenance. GET a unchanged; Scope Event inventory remains exactly one, no extra subject/fact.

Submit a fresh-key delete-where with the same empty selection. Require the same one-entry result and no-effect/position/Event invariants; independently reread a. Reopen once and reread both zero-match receipts and unchanged a/Event inventory. The nonempty starting graph is intentional: a globally broken no-op implementation cannot receive credit merely from an empty store. Positive mutation behavior is independently exercised by M, not assumed from setup.

Cap this row at16 actions, one live Resource/one Event, each set receipt one result. Use existing `response-metadata-equals` (as rollback Events already do) for unchanged position/epoch/view and existing absent-member assertions for effectPosition. No numeric position spelling or invented monotonic counter is hard-coded. This does not conflate zero-match with all-selected-no-op; the latter remains unobserved separately.

## Targeted controls and evidence honesty

Use existing negative-only manifest expectation controls in `apps/bdptest/src/transactional-matrix.test.ts:820` rather than a new response substitution hook (`afterExchange` is a void gate, not a transformer). Each control changes one independent expectation and must fail precisely that assertion while retaining actual successful wire evidence:

1. Wrong second-member expected postimage (complete-set state witness).
2. Wrong failed-operation prior record/metadata expectation (rollback afterstate witness).
3. Reversed literal receipt a/b order, and separately reversed Event subject order (each exact ordered oracle must fail).
4. A zero receipt expected to contain an extra result/count1 (exact array-size/content witness).
5. Zero Event expected to contain an extra updated fact (no-Event inventory witness).

At most six controls, no broad negative framework. Pin the literal expected subject/result tuples independently of runtime sorting and manifest-derived route lists. Controls test assertion discrimination, not real server fault injection; don't label mutated expectations as corrupted server output. Retain inherited wrong-source/wire/credential/firewall/cleanup controls unchanged. Actual receipt binding/reopen provenance remains attached to each own POST; a valid cross-wired binding is not inherently unauthorized, only wrong for the expected result.

## Files, preservation and counts

Exactly four owned source files proposed:

- `packages/conformance/test-support/transactional/manifest-v1.json`: append/insert two scenarios in canonical catalog order.
- `packages/conformance/test-support/transactional/coverage-v1.json`: exactly their ordered exhaustive assertion bindings plus clause-specific reasons. Both wholeContract:false; reasons include “Whole contract remains unqualified”.
- `packages/conformance/src/transactional-cumulative.test.ts`: extend exact substring pins from four to all six predecessor scenarios/coverage objects, independent clause map, canonical eight-row inventory, explicit source-growth ceiling and existing denial controls.
- `apps/bdptest/src/transactional-matrix.test.ts`: eight-row count/session/route inventory, expected outputs, six targeted controls, safe actual evidence summary.

Leave fixture-v1.json **byte-identical**: scope/batch/set targets already exist, existing a/b URIs are reused in independent fresh scenarios and c is a literal relative path. Preserve all six predecessor manifest and coverage object substrings byte-for-byte, including both recent set rows; preserve existing assertions and fixture/configuration semantics. No production, schema, catalog, spec, Type descriptor, RU artifact, evidence-seal, public export or admission edits. If another file is actually necessary, report the seam before expanding scope.

Canonical order becomes atomic-commit, rollback, set.mutation, set.zero-match, set.cardinality, set.singleton-targets, receipt.readable, failed-retained. Existing index-sensitive atomic-commit stays first. Required denominator250/five optional/applicable255 remain unchanged. Implemented8, missing planned required242; no full claims or whole promotions. Sessions9→13 (two fresh scenario sessions +two reopens); separate bootstrap means total children10→14. Existing five reopened receipt GET IDs remain pinned; add exactly four explicit reopened-M-update/M-delete/Z-update/Z-delete GET IDs and generation/source checks, for9 receipt exchanges, not9 reopen processes.

## Measured budget and proposed planning ceilings

Read-only measurements at9eb: manifest130887bytes/3647JSON-value nodes; coverage23604/541; fixture721/15. Six scenarios contain74 actions. Actual parser ceiling remains1MiB for private TX; remaining manifest byte headroom is917689. `artifact-bundle.ts:17,163,394` applies the1,048,576-byte default and selects the4MiB exception only for Read+Update; the existing cumulative test also asserts TX1MiB denial. This plan uses one TX manifest and does not assume the RU artifact-set executor supports TX sharding. Existing prior-slice planning test128KiB leaves only185bytes, so silently preserving that ceiling is impossible. Request root/council approval to explicitly re-anchor only that planning test to **224KiB total manifest**, <=7000nodes, incremental <=96KiB. Coverage growth <=24KiB; fixture0growth. These are bounded authoring budgets, not parser/transport limit changes. Old raw-object pins prevent using extra room to rewrite inherited observations.

At most48 new actions =>122 total; each row<=32/16, six controls. Largest requests remain bounded tiny Selector/property JSON; pin each new raw body<=2KiB, each finite response<=32KiB on this seven-Event fixture, far below existing receiver/1MiB observation caps. If the existing schema/runtime response exceeds that proposed planning bound, report actual bytes before adjusting; don't truncate evidence.

Existing exact packed9eb log reports the successful six-row cohort2570.915ms (test2574ms),74actions,10closed children; this is inspected author evidence, not a new timing run. It leaves87.429s against90s on that observed host, not a universal performance promise. The full44-test TX/SSE cohort45.52s is a different denominator and is not treated as one observation's elapsed time. Keep90s observation/120s outer/5s exchange/10s prepare+cleanup limits; final exact installed positive must remain<=63s for the previously accepted30% headroom target. Any timeout, cap excess or missing independent assertions is a root escalation, never authorization to enlarge execution deadlines or drop a row.

## Done predicates and open limits

Done when accepted predecessor is ordinarily integrated; both new rows run through the real branded installed private runner; all named receipt/state/Event/reopen assertions and six countercontrols discriminate; all old source objects/denial gates remain pinned; exact8/250 partial coverage and255 actual result inventory/omissions are honest; type/build/focused catalog/cumulative/app/receipt-observation checks and Read verifier pass; actual clean-source fresh packed TX cohort plus appropriate RU cross-profile positive confirms source/payload/contracts/config/fixture/routes/report identities; all owned children drain. Record original failures and actual phase durations without relabeling earlier evidence.

No unavoidable product ruling blocks this properties-only Bead subset. Negative/complex Selector policy, Snapshot retained-prefix minimum Q1, metadata-only incident-Link Events, authorization variation, owned attribution fanout, full Link Event expansion, receipt pagination, broader stress/retention and whole-profile qualification stay outside it. Runtime defects found by the new observations are reported to root/runtime owner rather than repaired in this evidence lane. No new semantics are inferred from current implementation behavior.
