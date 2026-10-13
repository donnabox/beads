# B numeric metadata whole-clause audit

Recommendation: the current B scenario supports all five of its selected obligations. The earlier audit's specific missing metadata-admissibility plane is now covered. No additional runtime case or assertion is necessary to close this selected row on the accepted finite-representative evidence standard. This is an audit recommendation for root's explicit decision; B remains `wholeContract:false`, and no coverage/source file was edited.

## Basis and evidence boundary

Read-only inspection of clean frozen `4d19a6a8a749b8f21cf96070a66e1c85f1528f1b` (tree `b14bcaaa0882824f782858c2a4f67562408ff204`) at `vickie-numeric-metadata-closure`, its actual manifest, catalog, coverage and countercontrols, the earlier `read-update-attribution-numeric-whole-audit.md`, and final installed reports. No tests, implementation, network, credentials or outward actions were performed for this audit. Writing this report and parsing existing JSON are inspection, not a new qualification run.

Normative source: `packages/conformance/catalog/numeric-model-v1.json:5–34`; `docs/specs/bdp.md` anchors `#revisions`, `#mutation-revisions-and-guards`, `#readupdate-consistency`, `#property-change-values`; `docs/design/requirements.md#protocol-requirements` (PROTO016). The Property Change paragraph explicitly applies the same comparison/admissibility definition to properties and metadata. No R/N executions or neighboring whole-row laws are borrowed below.

Authoring source: `packages/conformance/matrices/read-update-reference-continuation-v1.json:11605` (B), with new Bead closure beginning17985 and Link closure18301; coverage begins `packages/conformance/test-support/read-update-coverage-v1.json:13207`. App source `apps/bdptest/src/read-update-matrix.test.ts:3444` records positive B; numeric controls start3571 and3632; metadata controls4209–4318. Original B73 actions remain, followed by exactly8 actions/66 assertions. The complete81-action index below accounts for every action, including setup and auxiliary guards.

## Complete selected-clause map

| Selected obligation (complete text) | B's own evidence and discrimination | Assessment |
| --- | --- | --- |
| “Two JSON numbers are the same value if and only if the decimal values their literals denote are equal, whatever their spelling: `1.0`, `1`, and `1e0` are one value, and `9007199254740993` and `9007199254740992` are two.” | Independently pinned initial Bead/Link records; raw `num-{subject,link}-decimal` and `-exponent` preserve literal spellings before parsing, including nested reordered objects/arrays, negative zero and large exponent spelling. Mutation Resource and following GET equal the full initial Resource. `-inadmissible` refuses …993 rather than rounding it into admitted …992; `-safe`/`-distinct` independently pin admitted …992/…994 and changed state. Metadata closure now tests the same …993/…992 distinction in the other plane. Raw authored-body/write assertions and digest/different-number countercontrols discriminate a normalized or altered request. | Supported. The inadmissible …993 case is a refusal witness, not an impossible successful stored comparison. |
| “If applying an update leaves both `properties` and common `metadata` equal, under the JSON value-comparison rules of RFC 6902 Section 4.6, to their values immediately before that operation, then the operation retains the existing revision.” | `-decimal`, `-exponent`, `-metadata-noop`, `-combined-noop`, and their GETs compare entire Resource against independently validated prior capture for both Bead and unowned Link. Complete equality includes revision and provenance, not merely properties. `-current-noop` additionally exercises a current revision guard. Real properties/metadata changes, stale/current guards and restore distinguish no-op from unconditional no-effect behavior. Countercontrols alter an actual no-op response revision, following GET envelope and real-change postimage; only the named assertions fail. | Supported for properties-only, metadata-only and combined equal updates. |
| “Read+Update has no durable Mutation Receipts and no BDP Events.” | B's own `discovery` schema/profile and explicit absent transactional/events/receipts/changes/snapshot assertions; its own closed Read+Update operation directory schema; all successful writes use closed ordinary Mutation Result schemas, and failures closed Read+Update Problems. | Supported as the configured protocol contract/capability boundary. This does not purport to inspect internal SQL or forbid an implementation's private journal/idempotency records. No invented endpoint probes are needed. |
| “Number equality in that comparison, and the admissibility of every number literal a change carries, are defined under [Revisions](#revisions).” | Equality witnesses above cover both planes. Old `-inadmissible` plus `-after-inadmissible` reject a nested bad properties literal while preserving a companion metadata change's preimage. New actions73–80 reverse the planes: inadmissible metadata literal alongside effectful properties patch, exact422 diagnostic, complete unchanged GET, corrected admitted literal and independently asserted actual response/GET state. Both Resource kinds execute this. | Supported. This closes the exact missing conjunct from the prior audit. |
| “Number equality under the no-op law is over exact decimal values.” | Same row-local raw lexical and no-op comparisons as the first two clauses; raw distinct-value control prevents evidence credit from following only an authored object after binary64 normalization. | Supported; no additional independent law imported. |

## The new eight actions close the earlier finding

For both `subject` and `link`, `-metadata-closure-refused` sends a fresh-key raw update carrying `propertiesChange` `/zero`→7 and `metadataChange` `/number`→9007199254740993. It asserts actual complete-write/authored bytes,422, the Problem schema and the complete `validation-failed` Problem with `document:metadata` and the request-location message. It does not fabricate a document-relative `instanceLocation`.

`-metadata-closure-unchanged` compares the complete GET Resource to that kind's preceding `-distinct-read`, which itself has literal/schema assertions for identity, type, properties, metadata and actor (plus endpoints for Link). Thus the bad metadata does not silently round, and the valid companion properties update does not partially commit; revision and provenance remain unchanged too.

`-metadata-closure-corrected` changes the inadmissible literal to9007199254740992 with a new key. The positive is not a no-op: prior zero0 becomes7 and prior metadata number1 becomes9007199254740992. Its response and following `-metadata-closure-read` independently assert complete semantic fields, fixed identity/type/actor, no owned plane, and Link source/target. This supplies the admitted comparator and detects arbitrary refusal of metadata numerics. No new Resources/routes are required.

Two new countercontrols supplement seven preexisting B controls. One changes the unchanged GET's properties.zero to7 and must fail only `postimage`. The other substitutes plausible success after the real metadata422 and must fail only `status`, `body-schema`, `complete-problem`. Its returned negative run explicitly verifies all ten prior `num-subject-distinct-read` assertions passed and all other client/raw assertions passed. A pre-assertion capture alone is not its validity argument. These countercontrols are observation-corruption checks following actual service execution, not injected runtime failures.

The original seven B controls remain: raw digest; actual different-number request; plausible success replacing the properties refusal; changed no-op revision; changed no-op GET envelope; prior state replacing a real change; metadata side effect replacing the unchanged-after-properties-refusal GET. A's nine controls and R/N's nine controls are not B credit.

## Actual final installed evidence read

From `read-update-numeric-metadata-closure/packed/`:

- `attribution-numeric.numeric.json`: SHA256 `b560c0a1278c614fe5a711ae8b42ffa91bcfcae7f328b95a1408da99d9f06993`,256293 bytes. B `pass`,19 raw exchanges +62 client actions,599 assertions, every assertion passed; cleanup complete and claimEligible false.
- `artifact-set-progress.json.reference-continuation.json`: SHA256 `5588cd1b761e8576ced50b1e54dda02de4b1fbc1ec2fca45d9ec2d0efbee7529`,964026 bytes. Genuine shard B independently records the same81/599 all-pass evidence. This is not a fabricated aggregation of isolated positives.
- `matrix.json`:204/204 across the selected three-file cohort, including all seven B controls and both new controls, plus the authored-preservation test. Author's final supervisor exit0; no failures/skips/filters in that cohort. Prior build/type/format authoring failures remain disclosed in the author's report, not recast as final passing executions.
- Source4d19; installed payload `1b12629fe54c39c8f7864dbb13bdba16bcbe4298b0a5bd420fa19daaa4fb7ec2`; contracts `48667497d6e13064316d0364cf8f49dfc247ab9189306985a357d7b558333dfe`; policy `dbe8dbd5430f26e19e754d38167cd8084a32d702c695550e2ee0063dc2912430`; authentication `66941efc779f4f2431d65fdd5b2387b2af67bd7c1559e09dac032f82259ea16e`; fixture `f89932d68e3bf5867d79eb3de680af4bc67b761b3c3bce5da127f6ca75e27765`. Scope `https://reference.invalid/acme/`, Read+Update, History false.
- Genuine aggregate21.146170417s under90s; B14 raw POSTs/max526 bytes. Global counts remain129 required,54 authored,31 whole,98 uncovered,75 omitted. The observed suite/report does not automatically promote B.

## Residuals and smallest next action

The corrected metadata raw mutation response's opaque revision is not captured and compared to the following GET revision. Both representations satisfy their schemas and independently assert semantic fields; the missing cross-response opaque-token comparison remains honestly disclosed. It is not a missing conjunct of these selected clauses: that comparator is an effectful update, while the selected revision law requires retaining revision when both planes are equal. B's actual equal-update witnesses already compare full Resource, including revision, across response and GET. Broader result/read revision coherence is not silently added to this row.

Finite admitted/inadmissible lexical partitions, one configured installed service, fixed Types and Bead/unowned Link witnesses are representative conformance evidence, not exhaustive proof over all decimals, arbitrary nesting/Types, owned-link fanout, History or internal durable journals. R/N remain separate and uncredited. Existing source/evidence identities remain pinned; no prior report is relabeled.

Smallest next action: root may explicitly approve B whole-row credit, then make the bounded coverage/reason/independent-allowlist/count changes under its normal validation process. No extra runtime case is identified here. Until that decision, B stays false and all current counts remain unchanged; a hypothetical single-row promotion would be32 whole/97 uncovered, not public or whole-profile qualification.

## All81 authored actions accounted for

Zero-based index matches the manifest. Raw/client report containers separate these actions, but their IDs and assertion counts reconcile to this one authored row. Setup0–11, Bead witnesses12–41, Link witnesses42–72, metadata closure73–80.

| Index | Action | Transport | Assertions |
| --- | --- | --- | --- |
| 0 | `scope` | http | 2 |
| 1 | `discovery` | http | 8 |
| 2 | `directory` | http | 2 |
| 3 | `directory-captures` | client | 1 |
| 4 | `task` | http | 4 |
| 5 | `relates` | http | 6 |
| 6 | `num-create-subject` | client | 12 |
| 7 | `num-subject-initial` | client | 10 |
| 8 | `num-create-target` | client | 12 |
| 9 | `num-target-initial` | client | 10 |
| 10 | `num-create-link` | client | 14 |
| 11 | `num-link-initial` | client | 12 |
| 12 | `num-subject-decimal` | http | 6 |
| 13 | `num-subject-decimal-read` | client | 4 |
| 14 | `num-subject-exponent` | http | 6 |
| 15 | `num-subject-exponent-read` | client | 4 |
| 16 | `num-subject-metadata-noop` | http | 6 |
| 17 | `num-subject-metadata-noop-read` | client | 4 |
| 18 | `num-subject-combined-noop` | http | 6 |
| 19 | `num-subject-combined-noop-read` | client | 4 |
| 20 | `num-subject-current-noop` | client | 7 |
| 21 | `num-subject-properties-change` | client | 12 |
| 22 | `num-subject-properties-read` | client | 10 |
| 23 | `num-subject-metadata-change` | client | 12 |
| 24 | `num-subject-metadata-read` | client | 10 |
| 25 | `num-subject-stale` | client | 4 |
| 26 | `num-subject-stale-read` | client | 4 |
| 27 | `num-subject-current` | client | 7 |
| 28 | `num-subject-restore` | client | 12 |
| 29 | `num-subject-restored-read` | client | 10 |
| 30 | `num-subject-restore-stale` | client | 4 |
| 31 | `num-subject-restore-stale-read` | client | 4 |
| 32 | `num-subject-inadmissible` | http | 5 |
| 33 | `num-subject-after-inadmissible` | client | 4 |
| 34 | `num-subject-safe` | client | 12 |
| 35 | `num-subject-safe-read` | client | 10 |
| 36 | `num-subject-distinct` | client | 12 |
| 37 | `num-subject-distinct-read` | client | 10 |
| 38 | `num-subject-distinct-stale` | client | 4 |
| 39 | `num-subject-distinct-stale-read` | client | 4 |
| 40 | `num-subject-distinct-current` | client | 7 |
| 41 | `num-subject-target-unchanged` | client | 4 |
| 42 | `num-link-decimal` | http | 6 |
| 43 | `num-link-decimal-read` | client | 4 |
| 44 | `num-link-exponent` | http | 6 |
| 45 | `num-link-exponent-read` | client | 4 |
| 46 | `num-link-metadata-noop` | http | 6 |
| 47 | `num-link-metadata-noop-read` | client | 4 |
| 48 | `num-link-combined-noop` | http | 6 |
| 49 | `num-link-combined-noop-read` | client | 4 |
| 50 | `num-link-current-noop` | client | 7 |
| 51 | `num-link-properties-change` | client | 14 |
| 52 | `num-link-properties-read` | client | 12 |
| 53 | `num-link-metadata-change` | client | 14 |
| 54 | `num-link-metadata-read` | client | 12 |
| 55 | `num-link-stale` | client | 4 |
| 56 | `num-link-stale-read` | client | 4 |
| 57 | `num-link-current` | client | 7 |
| 58 | `num-link-restore` | client | 14 |
| 59 | `num-link-restored-read` | client | 12 |
| 60 | `num-link-restore-stale` | client | 4 |
| 61 | `num-link-restore-stale-read` | client | 4 |
| 62 | `num-link-inadmissible` | http | 5 |
| 63 | `num-link-after-inadmissible` | client | 4 |
| 64 | `num-link-safe` | client | 14 |
| 65 | `num-link-safe-read` | client | 12 |
| 66 | `num-link-distinct` | client | 14 |
| 67 | `num-link-distinct-read` | client | 12 |
| 68 | `num-link-distinct-stale` | client | 4 |
| 69 | `num-link-distinct-stale-read` | client | 4 |
| 70 | `num-link-distinct-current` | client | 7 |
| 71 | `num-link-target-unchanged` | client | 4 |
| 72 | `num-link-source-unchanged` | client | 4 |
| 73 | `num-subject-metadata-closure-refused` | http | 5 |
| 74 | `num-subject-metadata-closure-unchanged` | client | 4 |
| 75 | `num-subject-metadata-closure-corrected` | http | 13 |
| 76 | `num-subject-metadata-closure-read` | client | 9 |
| 77 | `num-link-metadata-closure-refused` | http | 5 |
| 78 | `num-link-metadata-closure-unchanged` | client | 4 |
| 79 | `num-link-metadata-closure-corrected` | http | 15 |
| 80 | `num-link-metadata-closure-read` | client | 11 |
