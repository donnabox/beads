# Native independent review: inherited Read on installed Read+Update

Reviewed exact frozen `c5627933dc932040a90aea463eb179dad00c8afa` against `68db54ac5d2971aaf02d10b829e2757d37349311`. Author checkout `/Users/dbox/repos/wt/vickie/vickie-read-update-read-coverage` untouched. Fresh detached review clone `/Users/dbox/repos/wt/vickie/bdp-read-update-inherited-native-review`. Read author `read-update-inherited-report.md` and external `read-update-executable-claude-review.txt`; focus is successor coverage, not already integrated runtime ancestry.

## Required repairs

### P2 — Bead collection whole-contract claim lacks complete per-identity record oracles

Locations: `packages/conformance/matrices/read-update-v1.json:1123–1152`, inventory `test-support/read-update-coverage-v1.json:202–205`; pinned whole list `src/cumulative-artifact.test.ts:155` (paths relative to packages/conformance).

The cited collection contract (`docs/specs/bdp.md:1509`, with record definition at1245) returns complete Resource records in canonical order. The matrix checks schema, next:null, an unordered set of titles, and a separate ordered array of IDs. It never associates each title with its ID or checks that record's Type, complete open properties or metadata against independent input. This is weaker than the report's “exact independent contents” and the wholeContract:true assessment.

Independent real-HTTP counterfactual: seed the actual authority through actual writes, then proxy just the collection response. Keep canonical IDs/revisions/order, swap title a↔z, drop the stored open property, substitute types/bug and replace metadata with an unstored object. The actual exact HTTP executor receives a self-consistent wire response (real status/headers/body octets); every authored assertion passes. No store bytes or setup outcomes are fabricated. `read.collections.beads` stays pass.

Impact: schema validity plus two projections does not establish complete stored records; a conforming-looking but wrong collection can be counted as whole. Fix with independent per-record tuples at least id/type/properties/metadata, retaining order/schema/revision validation; add a response-only negative control with swapped/missing values. Alternatively mark this row incomplete and adjust the reviewed list/uncovered number. Do not invent fixed revision oracles.

### P2 — Type inventory whole-contract claim observes IDs and shape, not actual summaries

Locations: `packages/conformance/matrices/read-update-v1.json:539–570`, inventory `test-support/read-update-coverage-v1.json:129–132`; pinned whole list `src/cumulative-artifact.test.ts:153`.

The cited Type inventory clause (`docs/specs/bdp.md:1090–1100`) describes summaries of advertised descriptors with exactly id/name/describes. Schema validation establishes fields and allowed kinds, but the independent expectations contain only IDs and order. Neither name nor describes is compared with the known fixture summary.

Independent real-HTTP counterfactual: proxy the real types/ response, preserve all13 IDs/order and next:null, change every name to “wrong inventory name” and every describes to bead, including actual Link Types. `read.types.inventory` still passes all assertions. Closed individual descriptor checks do not repair the missing inventory-to-descriptor value correspondence.

Impact: materially incorrect Type summaries are reported as whole-contract observation. Add a static independently declared id/name/describes tuple oracle and a wrong-name/wrong-kind response negative control, or keep this row incomplete. Together these two issues mean the current17/112 headline is mechanically reproduced but not yet supported by the claimed oracles; without added observations the conservative count would be15/114.

Exact reproducible external probe and output: `read-update-inherited-native/oracle-probe.mjs` and `.log`. It runs the real seeded authority behind a disposable HTTP proxy; request serializer/response integrity checks remain enabled. Early probe drafting omitted required target label, then attempted inconsistent decoded-body replacement which the runner correctly rejected by body-octet validation. The final retained probe sends the counterfactual over a real socket and both rows pass. These are assertion-quality findings, not observed bugs in the server implementation.

## All17 whole-contract assessments

Assessment is for the cited row clauses on this fixed configuration, not universal proof or public profile eligibility. Separate continuation, authorization, ownership and History requirements remain separately counted.

| Claimed ID | Citation / observed discriminator | Native assessment |
|---|---|---|
| read.discovery.scope-service-desc | Scope discovery: successful GET plus actual parsed service-desc relation capture | Supported bounded row |
| read.discovery.canonical-navigation | Scope discovery: each discovered beads/links/types URL equals canonical Scope resolution; wrong-navigation control fails | Supported |
| read.resource.link-record | Resource records: closed record schema and independent id/type/source/target/properties/metadata; actual opaque revision required by schema | Supported for present unpinned fixture |
| read.resource.properties-view | Resource views: exact independent full open properties for both kinds | Supported |
| read.resource.default-bounded | Resource views: real incident Links exist but default Bead response omits links plane | Supported for non-owning Type fixture |
| read.types.inventory | Types: complete exact summaries | Repair above; IDs/order/schema alone insufficient |
| read.types.descriptor | Types: closed descriptor/endpoint shape plus identity for all13 actual hosted Types | Supported for cited closed-shape clause |
| read.collections.beads | Collection/record clauses: complete records and canonical order | Repair above; unrelated projections miss per-record contents |
| read.collections.links | Collection/record clauses: complete schema plus independent id/type/source/target/properties/metadata tuples and order | Supported; no fabricated fixed revisions |
| read.collections.invalid-query | Collection clause: unsupported parameter and repeated limit/selector produce real400 request/invalid-parameter/never | Supported representative error contract |
| read.http.get-head | Problem details cumulative method support: real Scope/discovery/collections/records/properties/link view/Type inventory/descriptors GET and HEAD | Supported method-support row; conditional/cursor parity remains separately incomplete |
| read.http.method-405 | Problem details: unsupported method at Scope has405/exactAllow/body absence | Supported on CORS-disabled Read endpoint; mutation routes not wrongly constrained |
| read.http.cache-cors | HTTP cache clause: actual authorization-dependent collections/records/views/missing cache checks, disabled-CORS Origin/OPTIONS | Supported for current disabled-CORS configuration; no enabled-CORS claim |
| read-update.singleton.create-bead | Explicit operations/results: explicit and omitted ID creates, captured complete postimage agrees with real GET | Inherited68 bounded assessment retained; no successor change to operation semantics |
| read-update.singleton.delete-bead | Results/deletion/problems: incident-Link refusal + unchanged state, deletion final live revision, no resource field, three missing views | Inherited bounded assessment retained |
| read-update.singleton.expected-revision | Guards/problems: matching, stale update and deletion, unchanged state, omitted guard and successful matching deletion | Inherited bounded assessment retained |
| read-update.singleton.idempotency-conflict | Duplicate-key clause: conflict, unchanged current record, original retained disposition replayed unchanged | Inherited bounded assessment retained |

The four authored inherited rows marked incomplete remain correctly bounded: Bead record untouched-property preservation after update; Accept auth/non-disclosure precedence; conditional auth/cursor cases; HEAD auth/cursor parity. No observation is inferred merely from another row passing.

## Counts, fixture and transport

Independent JSON recount:129 required inventory rows,24 authored,17 flags true,105 required entirely omitted;26 inherited Read IDs entirely omitted. Real unfiltered run reproduced134 applicable scenarios with24 pass/110 not-implemented harness errors,24 sessions,44 client exchanges,44 actual seed writes and21 native Read observations; reported uncovered112 is the flag-derived count above. Catalog/manifest/fixture and exact-configuration digests reproduce the author report exactly; sourceDirty:false and claimEligible:false.

Setup writes two Beads and two Links in reversed canonical order through real credentialed HTTP. Setup checks declared values rather than generating oracle values from returned server records. Schemas plus static fixture oracles are sound where they compare complete values. Missing seed/control failures are meaningful. The two findings are specific missing value projections, not an assertion that all seed/oracle independence is circular.

The native header adapter retains one exact executor and finite route/credential boundary. Raw requests delegate unchanged; semantic requests permit only GET/HEAD and named Read fields, require own string data descriptors, reject casing collisions, accessors/unresolved values/control bytes, and carry original target/signal/raw target into exact serialization. Its operator credential selection is explicitly this fixture's binding, not general authorization support. Actual socket test proves strong/weak conditional header serialization and HEAD bodylessness. Runner captures/response assertions remain authoritative. No new semantic table or expected-result synthesis was introduced. No additional current adapter correctness finding.

## External A–H successor dispositions

- A: mechanically fixed—explicit17-ID list and exact112 uncovered pin prevent silent flag drift. Semantic completeness still needs the two repairs above; a pin is not proof of its content.
- B/C: inherited68 fix preserved—update-bead incomplete for metadata/effective schema; no-op returned record equality and response-only corrupt-revision control retained.
- D: renamed guard-binding-agreement accurately denotes request-object binding; no wire claim remains. Actual serializer digest/octet observations and stale-guard effects are separate evidence.
- E: bootstrap documented as target-module self-consistency; installed payload/source association explicitly reviewer/launcher provenance. Per-scenario identity consistency remains substantive. No claim of independent attestation.
- F: structural/conditional planner readiness remains explicitly non-admitting; actual exact harness gates still precede execution. No new execution admission modeled or claimed; accepted planning limitation rather than new code closure.
- G: three formerly missing carrier cells added; version named in documentation; v1 non-Read refusal comment added. No observed regression.
- H: returned plan digest compared to bundle/report; exact existing configuration digest pinned. Credential content remains excluded from that projection and separately bound in private auth configuration, explicitly documented.

## Independent validation / provenance

Pinned Node24.16.0/pnpm11.20.0. Fresh isolated frozen install and build PASS before tests.90 tests/7files PASS,12.23s,maxWorkers2: cumulative-artifact, observed-binding-assertion, profile-contract-plan, read-update-catalog, read-update-exact-client, installed-read-fixture, read-update-matrix. Logs/structured report in `read-update-inherited-native/`. Actual progress report reproduced the author catalogDigest03ca39f2…, manifestDigestbc3fd575…, fixtureDigest819d6452…, exactConfigurationDigestae2f29b8… . Fresh emitted payload used; did not independently repeat author packed install or886-test run.

No repo/source/outward edits, no full suite, no new public qualification/evidence claim. Author and independent clone clean at completion. Disposable probe stores/listeners/processes closed; source frozen at exact c562793. Recommend hold whole-contract count acceptance for the two bounded oracle repairs; remaining architecture/admission boundaries are intact.
