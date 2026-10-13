# Next bounded Update evidence: numeric admission and diagnostic locations

Plan only; no source edits or new test executions. Inspected clean frozen **28a2f3799edd9a642086702453bfb7c8d8597f5b** at `/Users/dbox/repos/wt/vickie/vickie-read-update-attribution-numeric`. The sole plan remains the fork design document. Root reports PR93 merged as92c188794edd62016b3633d7de8543f8e1a08f76; that main ancestry must be preserved by ordinary integration with reviewed28a and any council repair. Implementation starts only from a root-supplied frozen integrated successor, not by relabeling28a or assuming its outstanding external review is complete.

## Recommendation and remaining inventory

Add exactly two independent continuation scenarios, initially **wholeContract:false**:

1. `read-update.numeric-model.round-trip-admission`
2. `read-update.numeric-model.nested-refusal`

They are presently omitted (`read-update-coverage-v1.json:13720,13727`). The current canonical inventory is129 required,52 authored,30 whole,99 without whole credit,77 unplanned. These two rows would yield54 authored/75 unplanned, with whole/uncovered counts unchanged unless root separately promotes already-reviewed rows. Never overwrite such intervening promotions. No extra IDs receive indirect credit.

These observations address silent data corruption at a generic client/store boundary. The preceding numeric no-op row only refuses one nested patch integer per kind; it does not independently cover whole-document admission, precision and range classes, deep placement or document-relative diagnostics. The proposed rows exercise real existing routes and require no new policy, Type, runtime hook, clock or product decision.

Other useful omissions were considered. Authorization-view/hidden-subject cases need a separately bound credential/policy fixture. Concurrent/in-progress, retention and crash rows need installed process/scheduling observations, not serial manifest assertions or qualification clocks. Alias allocation cannot prove collision handling by sampling. HTTP method/cache rows are useful but would add less new data-integrity evidence than this existing gap. Selector complexity, metadata-only incident-Link projection, Snapshot, Types and optional History remain outside this slice.

## Exact law and implementation basis at28a

| Row | Catalog and normative clauses | Finite proposed observation and limit |
| --- | --- | --- |
| round-trip-admission | `catalog/numeric-model-v1.json:38–68`; `docs/specs/bdp.md:694–723` Revisions, particularly709–717; `:2522–2529` Numeric admission; `:2658–2659` Validation and results; `:6650` conformance table; `docs/design/requirements.md:109–115` PROTO-016 | Three concrete inadmissible classes, both Resource kinds, both create-whole and update-patch carriers, paired admitted corrections and complete observed state. This is not exhaustive over all binary64 boundaries or literals, and does not by itself close the catalog's universal “at any depth” wording. |
| nested-refusal | `catalog/numeric-model-v1.json:71–97`; `docs/specs/bdp.md:2522–2529`; `:3452–3473` Read+Update problem details; `:6651`; PROTO-016 | Whole properties/metadata and unapplied properties/metadata patches, finite depths and escaped keys/array positions, exact configured diagnostics and no mutation. “Any depth” is sampled below actual receiver bounds, not an unbounded recursion proof. Optional patch instanceLocation omission is this implementation's observed choice, not a universal requirement to omit. |

Supporting context: every BDP JSON text follows this number model (`bdp.md:891–903`); metadata is not Type-schema validated, but is still BDP JSON. String/object syntax failures are distinct from inadmissible numbers (`:2518–2520`). The numeric refusal is `validation-failed`/422/`retry:never`, not400 malformed-request (`:3413` and the complete table). A failed admitted operation can retain its disposition: all corrected requests here use **fresh keys**, not a claim that numeric refusal leaves a key unclaimed.

The real path already exists: `read-update-values.ts:366–374` admits numbers over the operation input; `json-admission.ts:212–217` compares exact decimals to the shortest finite binary64 serialization, and`:323–345` walks occurrences and formats bounded diagnostics. `member-executor.ts:607–610` creates the real validation-failed disposition before Resource evaluation on numeric failure. `member-numeric-diagnostics.ts:13–31` removes the whole-create document prefix for instanceLocation and labels metadata;`:35–51` leaves unapplied patch instanceLocation absent and labels metadata without attributing the failure to a Type. `development-authority.ts:230–238` supplies the actual diagnostic messages. These are source support, not a substitute for new HTTP evidence.

These initial rows use singleton carriers only; sequence-member numeric validation, retained-failure replay, concurrency and alternate authorization views remain explicit unobserved variants. Do not promote whole merely because all planned singleton examples pass.

No setup change is needed: `scripts/installed-reference-observation.mjs` supplies the same operator credential/policy and empty aggregate-policy override. `installed-schema-cohort.ts:60–74` adds its existing schema Types without changing shipped Task/relates. Task owns no Links, relates accepts the row's in-Scope Task endpoints, and the existing configuration imposes no endpoint-multiplicity override; separate Link IDs permit the three same-pair Links in rowR. Actual served descriptor identity/kind/endpoints are asserted before relying on them. No owned-Link claims belong here.

## Shared execution/oracle rules

Use the current two-artifact index, same genuine factory and shared fixture. Append these rows after existing numeric no-op in canonical catalog order. Keep all baseline49 blocks, C and the two28a blocks byte-identical; preserve their coverage, controls, schema references and outcomes. New rows own literal IDs with `nr-` and `nn-` prefixes and dedicated fixture point-GET bindings. No collection path fallback, allocated-ID sampling, new route wildcard or cross-row setup dependency.

Each row begins its own existing six-action Scope→discovery→directory/captures→Task/relates descriptor navigation. Check closed actual discovery, directory and descriptor schemas plus independently authored descriptor identity/kind/endpoint facts. Check advertised diagnostic count32/bytes524288 where two diagnostics are expected; these tiny cases deliberately stay below both. This is not a truncation or diagnostic-limit qualification row.

Every numeric create/update—negative and corrected positive—uses existing `family:http`, `raw.body.encoding:"utf8"` and explicitly authored literal bytes. No request is passed through JSON.parse→JSON.stringify, including overflow1e309 and underflow1e-400. Every raw mutation carries a fresh distinct literal key and `request-authored-body`/`request-write-state:"complete"`; normative mutation/Problem schema checks remain separate. The actual report must project source=`raw-http1-serializer`, complete write and actual byte count, while retaining current raw bodyDigest withholding. No synthetic-manifest approval is enabled.

Use existing client GETs to capture full records only after independent literal identity, Type, properties, metadata, endpoints, provenance absence and closed record schema assertions. All corrected numeric postimages come from independent authored expected JSON. A later failing request never defines its own expected afterstate. Opaque revisions are carried only for complete-record equality. Raw object captures remain forbidden; complete client capture followed by `json-pointer` `equalsBinding` is the approved seam. Preserve `exists:true`.

The exact configured Problem oracle is a whole object, not a permissive subset: family URI, status422, codevalidation-failed, retrynever, exact diagnostics and absent incidental title/instance/extensions. This is an exact fixture expectation, not a new prohibition on other conforming Problem extensions. No Type/schemaLocation blame for numeric admission. Whole properties create: message `unsupported number`, document omitted, document-relative instanceLocation. Whole metadata create: `metadata: unsupported number`, documentmetadata, document-relative location. Patch messages identify the actual operation-input pointer; metadata patch has documentmetadata; both omit instanceLocation in this implementation. No request pointer masquerades as a document pointer.

## R — round-trip admission,72 planned actions

Eight fixed Resources at most: two Task endpoints plus three Task subjects and three unowned relates Links (`nr-<kind>-integer|precision|overflow`). Use the same two endpoints for all three Links. Capture endpoint records independently once and check their complete preservation.

| Case | Authored refused literal | Admitted replacement | Distinction |
| --- | --- | --- | --- |
| integer |9007199254740993|9007199254740992| Rounding a distinct exact integer is not admission. This is not a rule forbidding every integer above2^53; the preceding row already distinguishes admitted9007199254740994. |
| precision |0.12345678901234567891|0.125| Twenty significant digits in this specific non-round-tripping value; not a universal digit-count rule. |
| overflow |1e309|1e300| Valid JSON numeric syntax with a nonfinite binary64 conversion is a422 admission failure, not malformed JSON. |

For each case and each kind:

1. Raw whole create with `properties.value` equal to the refused literal, the four admissible spelling fields listed below already present, and otherwise valid metadata. Expect the exact configured422 Problem naming `/value`; fixed-ID GET404 proves absence.
2. Correct **only the numeric literal and key**, preserving route/body fields, then raw create must succeed. Its complete literal postimage and independent canonical GET must agree. Every corrected whole properties document also includes the admitted spelling corpus `one:1.0`, `zero:-0.0`, `large:1e300`, and `expanded:` literal1 followed by300zeroes. Expected semantic values are independently authored1,0,1e300,1e300.
3. On that now-live subject, raw properties patch replaces `/value` with the original refused literal, accompanied by an otherwise-valid metadata change. Expect exact422 diagnostic message with input pointer `/propertiesChange/0/value`, no document-instance pointer/Type blame; independent GET must equal the complete prior record, proving the companion metadata change did not commit.
4. Correct only the literal and key, keeping the companion metadata change; raw update and independent GET must show the exact authored properties and changed metadata. This is a useful positive discriminator against reject-all-numbers or ignoring all input. It is not another whole no-op/idempotency claim.

Action ledger:6navigation +4endpoint create/read +6whole cycles×4 +6patch cycles×4 +12endpoint GETs (both endpoints immediately after each of the six Link refusals) +2final endpoint GETs =**72**. New afterstates follow every refusal and correction. Exactly29 raw exchanges (5navigation +24mutations) and43 client exchanges are planned; all current explicit IDs and8 fixture bindings. No alias or deletion is needed. Remaining unrelated Resources are outside the subject/endpoints “nothing changed” observation; do not infer a global store audit from a handful of point GETs.

## N — nested refusal and diagnostic locations,52 planned actions

Six fixed Resources at most: two Task endpoints plus one properties-case and one metadata-case subject per kind. Both kinds execute the same four cases:

1. **Whole properties:** an initially empty `queue`, one inadmissible integer at `/shallow` and another at a specifically authored object depth16 under `/deep/.../a~1b~0c` (raw key `a/b~c`). Expect both configured diagnostics, in this fixture's known evaluation order, with correct document-relative RFC6901 escaping and no Type/schemaLocation. Fixed-ID404, then the same body with both literals replaced by independently chosen admitted values under a fresh key; complete raw postimage and independent GET.
2. **Unapplied properties patch:** on that corrected Resource, append an object containing a nested array/escaped key at `/queue/-` (the valid initial properties include an empty queue). Inadmissible leaf9007199254740993 at a depth4 input subtree. Companion metadata change is otherwise valid. Expect the exact `/propertiesChange/0/value/...` input pointer in the message and no instanceLocation; full GET unchanged. Correct only the numeric leaf and key; the expected appended value and metadata are independently authored and GET-checked.
3. **Whole metadata:** separate fixed-ID create with valid properties, an initially empty metadata queue and an inadmissible underflow1e-400 in an object/array at depth3 in metadata. Expect documentmetadata and the relative metadata instanceLocation; fixed-ID404. Correct only that literal to0 and key; raw response and independent full GET prove the numeric/document distinction. Nonzero exact underflow is refused, independently of the admitted negative-zero corpus inR.
4. **Unapplied metadata patch:** append a nested object/array with an escaped member and inadmissible twenty-digit decimal to a metadata queue established by the valid create. Companion properties marker changes otherwise validly. Expect documentmetadata, `metadata: ... /metadataChange/0/value/...` and no postimage instanceLocation. GET unchanged, then correct only numeric leaf and key and check both resulting documents.

For whole creates, all other document fields are identical across the refusal/correction pair. For patches, the other document's would-be change is retained in both requests, so a premature partial commit is observable. Do not conflate document depth with JSON Pointer patch-path depth: `/queue/-` is shallow while the admitted value is nested. These are finite depth1/3/4/16 cases, not a parser-maximum test.

Ledger:6navigation +4endpoint create/read +8cycles×4 +8endpoint GETs (both endpoints after each of four Link refusals) +2final endpoint GETs =**52**. Exactly21 raw exchanges (5navigation +16mutations),31 client exchanges and6 new fixture bindings are planned. Complete prior records are refreshed after each corrected positive before the next refusal.

## Nine bounded negative controls

Reuse the existing app-only client `RetryControl` with exact clientAction.input equality and the narrow exact raw wrapper at `read-update-matrix.test.ts:396–438`. Extend its closed case mapping only for named new row/action/key targets; no helper/executor/runner changes or configurable target framework. Actual exchange occurs before response controls. Preserve existing controls byte-for-byte where possible and at least behavior/expected failure sets exactly.

1. R: replace the integer refusal's raw token with its admissible neighbor before transmission. The authored-body assertion and expected422/result oracles must fail; assert one substitution and the exact intended assertion set.
2. R: keep422/family/code/retry but corrupt its diagnostic member pointer; whole Problem equality fails.
3. R: corrupt a corrected positive numeric postimage while preserving identity/revision; literal properties equality fails.
4. R: corrupt the GET after a refused combined patch so the otherwise-valid metadata change appears committed; whole prior-record equality fails.
5. N: prefix a whole-properties instanceLocation with `/properties`; complete Problem equality fails.
6. N: add a request-body-shaped instanceLocation to an unapplied patch diagnostic; exact fixture Problem equality fails (this detects this configured behavior's regression, not all conforming optional pointer choices).
7. N: remove the metadata document label while preserving the diagnostic message/status; complete Problem equality fails.
8. N: omit the deep second diagnostic from the two-leaf whole-document refusal; complete Problem equality fails.
9. N: corrupt one nested field of the independent unchanged GET after a refused metadata patch; complete prior-record equality fails.

All controls must fail their intended assertion IDs, not merely any schema/harness failure, and apply exactly once. Successful corrections are the complementary reject-all/ignore-input discriminator. A wrong numeric expected oracle derived from the target is not allowed. Control-only response substitution is explicitly not a wire-transparent faulty authority claim.

## Measured headroom and proposed stop thresholds

Measured28a artifacts and fresh packed evidence (`read-update-attribution-numeric/source-audit.json`, `packed-summary.json`):

| Artifact | Bytes / nodes | Remaining existing cap |
| --- | --- | --- |
| frozen baseline manifest |3,626,922 /97,116|567,382bytes /2,884nodes; **do not append here** |
| continuation |521,921 /13,945|3,672,383bytes /86,055nodes |
| shared fixture |29,055 /750|1,019,521bytes /99,250nodes |
| coverage |646,402 /13,461|402,174bytes /86,539nodes |
| index |242 /9|unchanged; no new shard or fields |

Actual two-shard aggregate17.819s,72.180s remaining within90s. Old A/B positive single-row observations took0.647s/0.616s; the whole five-file241-test invocation was123.76s. These are historical measurements, not a forecast guarantee for added rows. The whole-run supervisor and each in-test deadline are distinct. Keep90s aggregate,120s outer aggregate test,5s exchange and30s focused tests. Retain the existing240s five-file process-supervision ceiling; do not enlarge it if the added controls expose overload. The prior1,241,456KiB peak is whole descendant-tree RSS, not an owner or parser-memory bound.

Proposed static thresholds before running: R<=72/N<=52, combined<=124 actions; <=9new controls; <=8/6 row Resources; each actual/authored request<4KiB; incremental continuation<=450KiB/16,000nodes, fixture<=8KiB, coverage<=65KiB. Estimated independent assertions <=1,300 combined, not a minimum to pad. These are planning ceilings, not measured authored sizes. Measure after concrete authoring and **stop/report** before dropping cases, weakening independent GETs, compression or cap/time changes. Expected route count is49 (current35 plus14 literal point GETs); raw create operations already belong to the mandatory nine carriers, but their first occurrence can change route order. Independently enumerate the exact final method/URL order and pin its freshly calculated digest; do not just append a count to the prior digest. Baseline route pin stays unchanged.

The app file is already large; this scope adds data/closed test cases only, no new generic executor abstraction or configuration surface. If concrete authoring cannot fit these bounds or needs a helper seam, report it for separate scope review.

## Owned files, validation and done-when

Exactly four files proposed: continuation manifest, `fixtures/read-update-reference-v1.json`, `test-support/read-update-coverage-v1.json`, and `apps/bdptest/src/read-update-matrix.test.ts`. Root owns eventual whole-list promotion. No runtime/server, schemas, normative catalogs, runner, executor, index, helper, TX inventories, Type artifacts or public entrypoints. Update continuation selected-ID/count/route assertions for five canonical rows while preserving old row bytes and C/small-set behavior. Common fixture digest changes require a genuine fresh execution of **both** shards; never assemble a pass from old baseline and new continuation reports.

Machine-evaluable completion predicates:

- Source parent is the approved ordinary integration; only the four authorized paths differ. Baseline manifest hash equals28a'sc86b3353…; raw C, attribution and no-op blocks each equal the accepted parent. Existing fixture values and nonselected coverage rows are semantically identical; new fixture bindings have no collisions.
- Genuine artifact-set factory accepts the unchanged129 ordered required inventory. Each new ID is owned once in continuation; every coverage binding equals its row's exact ordered assertion inventory and is nonempty. Both new whole flags false. With no independent promotion, counts are54authored/30whole/99uncovered/75unplanned.
- Static ledger meets all action/resource/key/body/source-growth caps. All numeric mutations are raw UTF-8, use unique fresh keys, contain exact intended lexical tokens, and assert authored body/complete write. Every refusal has its independent afterstate; every negative is paired with a fresh-key admitted correction differing only in the intended literal(s) and key.
- Actual installed R/N rows pass all authored assertions with expected raw/client exchange counts; no public setup replay or cross-row dependency substitutes for row-local setup. Per-row full record baselines were independently checked before capture. All actual bodies<4KiB and raw digest projection remains withheld.
- Nine controls apply once, fail exact intended assertion sets, and leave no child. New whole-record or diagnostic corruption cannot hide behind assertion-free manifest metadata.
- Frozen-lock install, pinned Node24.16.0/pnpm11.20.0, typecheck/clean build before process tests, focused new/C/association controls, applicable artifact regressions/static/boundaries/bothsmokes/Readseal complete. Fresh clean packed five-file cohort uses exact frozen head and fresh payload/config/fixture/route identities; no8k workspace rerun required for this evidence-only slice. Record all failures and explicit exits.
- Both genuine per-shard results persist with separate digests and exact source/config association; required omission inventory remains complete. Cooperative abort/failure retention and cleanup controls continue passing. Preserve90/120/5/30 bounds, no token in reports, no success summary for failed/incomplete execution, no owned process remaining.

No true product decision or known missing runtime capability was found for this bounded proposal. Any concrete contrary HTTP behavior discovered later is a runtime-owner finding, not permission to change the oracle, runtime or law. Council should approve the finite partitions and metadata diagnostic adjunct; neither arbitrary-depth nor every-number universality is claimed by initial false flags.
