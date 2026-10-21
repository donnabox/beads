# Next bounded Update evidence: numeric admission and diagnostic locations

Plan revision only; no source edits, test executions or implementation authorization. Inspected frozen **9c7b643d22562ebabd8d9b51a24157b66f3e5c89** at `/Users/dbox/repos/wt/vickie/vickie-read-update-attribution-review-repair`, the reviewed PR94 repair whose direct parent is integration2bc2a94214954e3e03261d5e8b0dd2baac2a7268. The sole plan remains the fork design document. Implementation must await root/council acceptance and a root-supplied ordinary descendant preserving main and reviewed9c; this document does not claim9c merged or replace its historical reports. If the accepted parent changes these anchors or old rows, reconcile explicitly before authoring.

Original plan preserved byte-for-byte in `read-update-after-attribution-next-plan-original-116be54.md` (SHA256116be54fe2306c7aaa779cc5b9a13c2d2f5010120613f63e6b33e1b87207285a). Per-finding dispositions are appended below.
## Recommendation and remaining inventory

Add exactly two independent continuation scenarios, initially **wholeContract:false**:

1. `read-update.numeric-model.round-trip-admission`
2. `read-update.numeric-model.nested-refusal`

They are presently omitted; locate them by their exact IDs in `read-update-coverage-v1.json` (line positions moved with9c bindings). The current canonical inventory is129 required,52 authored,30 whole,99 without whole credit,77 unplanned. These two rows would yield54 authored/75 unplanned, with whole/uncovered counts unchanged unless root separately promotes already-reviewed rows. Never overwrite such intervening promotions. No extra IDs receive indirect credit.

These observations address silent data corruption at a generic client/store boundary. The preceding numeric no-op row only refuses one nested patch integer per kind; it does not independently cover whole-document admission, precision and range classes, deep placement or document-relative diagnostics. The proposed rows exercise real existing routes and require no new policy, Type, runtime hook, clock or product decision.

Other useful omissions were considered. Authorization-view/hidden-subject cases need a separately bound credential/policy fixture. Concurrent/in-progress, retention and crash rows need installed process/scheduling observations, not serial manifest assertions or qualification clocks. Alias allocation cannot prove collision handling by sampling. HTTP method/cache rows are useful but would add less new data-integrity evidence than this existing gap. Selector complexity, metadata-only incident-Link projection, Snapshot, Types and optional History remain outside this slice.

## Exact law and implementation basis at9c

| Row | Catalog and normative clauses | Finite proposed observation and limit |
| --- | --- | --- |
| round-trip-admission | `catalog/numeric-model-v1.json:38–68`; `docs/specs/bdp.md:694–723` Revisions, particularly709–717; `:2522–2529` Numeric admission; `:2658–2659` Validation and results; `:6650` conformance table; `docs/design/requirements.md:109–115` PROTO-016 | Three concrete inadmissible classes, both Resource kinds, both create-whole and update-patch carriers, paired admitted corrections and complete observed state. This is not exhaustive over all binary64 boundaries or literals, and does not by itself close the catalog's universal “at any depth” wording. |
| nested-refusal | `catalog/numeric-model-v1.json:71–97`; `docs/specs/bdp.md:2522–2529`; `:3452–3473` Read+Update problem details; `:6651`; PROTO-016 | Whole properties/metadata and unapplied properties/metadata patches, finite depths and escaped keys/array positions, exact configured diagnostics and no mutation. “Any depth” is sampled, not an unbounded recursion proof. Depth16 is bounded here by actual64KiB body/properties capacities and the128-level artifact-parser cap, not a nonexistent receiver document-depth ceiling. N3/N4 metadata refusal is a configured adjunct inferred from the general number model at891–903; its diagnostic document label and optional location rules are directly specified at3452–3473. Optional patch instanceLocation omission is this implementation's observed choice, not a universal requirement to omit. |

Supporting context: every BDP JSON text follows this number model (`bdp.md:891–903`); metadata is not Type-schema validated, but is still BDP JSON. This supports the configured metadata-refusal expectation as an inference; the properties-scoped Numeric admission MUST and PROTO-016 do not become new quoted metadata mandates. If metadata behavior disagrees, audit the existing general-number and diagnostic semantics with the runtime/spec owners before deciding whether this is an implementation defect or a specification ambiguity; do not silently weaken the oracle or declare a new law. String/object syntax failures are distinct from inadmissible numbers (`:2518–2520`). The numeric refusal is `validation-failed`/422/`retry:never`, not400 malformed-request (`:3413` and the complete table). A failed admitted operation can retain its disposition: all corrected requests here use **fresh keys**, not a claim that numeric refusal leaves a key unclaimed.

The real path already exists: `read-update-values.ts:366–374` admits numbers over the operation input; `json-admission.ts:212–217` compares exact decimals to the shortest finite binary64 serialization, and`:323–345` walks occurrences and formats bounded diagnostics. `member-executor.ts:607–610` creates the real validation-failed disposition before Resource evaluation on numeric failure. `member-numeric-diagnostics.ts:13–31` removes the whole-create document prefix for instanceLocation and labels metadata;`:35–51` leaves unapplied patch instanceLocation absent and labels metadata without attributing the failure to a Type. `development-authority.ts:230–238` supplies the actual diagnostic messages. These are source support, not a substitute for new HTTP evidence.

These initial rows use singleton carriers only; sequence-member numeric validation, retained-failure replay, concurrency and alternate authorization views remain explicit unobserved variants. Do not promote whole merely because all planned singleton examples pass.

No setup change is needed: `scripts/installed-reference-observation.mjs` supplies the same operator credential/policy and empty aggregate-policy override. `installed-schema-cohort.ts:60–74` adds its existing schema Types without changing shipped Task/relates. Task owns no Links, relates accepts the row's in-Scope Task endpoints, and the existing configuration imposes no endpoint-multiplicity override; separate Link IDs permit the three same-pair Links in rowR. Actual served descriptor identity/kind/endpoints are asserted before relying on them. No owned-Link claims belong here.

## Shared execution/oracle rules

Use the current two-artifact index, same genuine factory and shared fixture. Append these rows after existing numeric no-op in canonical catalog order. Keep all baseline49 blocks, C and the two reviewed9c blocks byte-identical; preserve their coverage, controls, schema references and outcomes. New rows own literal IDs with `nr-` and `nn-` prefixes and dedicated fixture point-GET bindings. No collection path fallback, allocated-ID sampling, new route wildcard or cross-row setup dependency.

Each row begins its own existing six-action Scope→discovery→directory/captures→Task/relates descriptor navigation. Check closed actual discovery, directory and descriptor schemas plus independently authored descriptor identity/kind/endpoint facts. Check advertised diagnostic count32/bytes524288 where two diagnostics are expected; these tiny cases deliberately stay below both. This is not a truncation or diagnostic-limit qualification row.

Every numeric create/update—negative and corrected positive—uses existing `family:http`, `raw.body.encoding:"utf8"` and explicitly authored literal bytes. No request is passed through JSON.parse→JSON.stringify, including overflow1e309 and underflow1e-400. Every raw mutation carries a fresh distinct literal key and `request-authored-body`/`request-write-state:"complete"`; normative mutation/Problem schema checks remain separate. The actual report must project source=`raw-http1-serializer`, complete write and actual byte count, while retaining current raw bodyDigest withholding. No synthetic-manifest approval is enabled.

Use existing client GETs to capture full records only after independent literal identity, Type, properties, metadata, endpoints, provenance absence and closed record schema assertions. All corrected numeric postimages come from independent authored expected JSON. A later failing request never defines its own expected afterstate. Opaque revisions are carried only for complete-record equality. Raw object captures remain forbidden; complete client capture followed by `json-pointer` `equalsBinding` is the approved seam. Preserve `exists:true`.

The exact configured Problem oracle is a whole object, not a permissive subset: family URI, status422, codevalidation-failed, retrynever, exact diagnostics and absent incidental title/instance/extensions. This is an exact fixture expectation, not a new prohibition on other conforming Problem extensions. No Type/schemaLocation blame for numeric admission. Whole properties create: message `unsupported number`, document omitted, document-relative instanceLocation. Whole metadata create: `metadata: unsupported number`, documentmetadata, document-relative location. Patch messages identify the actual operation-input pointer; metadata patch has documentmetadata; both omit instanceLocation in this implementation. No request pointer masquerades as a document pointer.

## R — round-trip admission,72 planned actions

Eight fixed Resources at most: two Task endpoints plus three Task subjects and three unowned relates Links (`nr-<kind>-integer|precision|overflow`). Use the same two endpoints for all three Links. Capture endpoint records independently once and check their complete preservation.

| Case | Authored refused literal | Whole-create correction | Effectful patch correction | Distinction |
| --- | --- | --- | --- | --- |
| integer |9007199254740993|9007199254740992|9007199254740994| Rounding a distinct exact integer is not admission. This is not a rule forbidding every integer above2^53; the preceding row already distinguishes admitted9007199254740994. |
| precision |0.12345678901234567891|0.125|0.25| Twenty significant digits in this specific non-round-tripping value; not a universal digit-count rule. |
| overflow |1e309|1e300|1e299| Valid JSON numeric syntax with a nonfinite binary64 conversion is a422 admission failure, not malformed JSON. |

For each case and each kind:

1. Raw whole create with `properties.value` equal to the refused literal, the four admissible spelling fields listed below already present, and otherwise valid metadata. Expect the exact configured422 Problem naming `/value`; fixed-ID GET404 proves absence.
2. Correct **only the numeric literal and key**, preserving route/body fields, then raw create must succeed. Its complete literal postimage and independent canonical GET must agree. Every corrected whole properties document also includes the admitted spelling corpus `one:1.0`, `zero:-0.0`, `large:1e300`, and `expanded:` literal1 followed by300zeroes. Expected semantic values are independently authored1,0,1e300,1e300.
3. On that now-live subject, raw properties patch replaces `/value` with the original refused literal, accompanied by an otherwise-valid metadata change. Expect exact422 diagnostic message with input pointer `/propertiesChange/0/value`, no document-instance pointer/Type blame; independent GET must equal the complete prior record, proving the companion metadata change did not commit.
4. Correct only the literal and key, keeping the companion metadata change, but use the distinct **effectful patch correction** in the table:9007199254740994,0.25 or1e299. Raw update and independent GET must show `/value` changed from the whole-create correction AND the independently authored changed metadata. A receiver that drops propertiesChange while applying metadataChange must fail. This also distinguishes reject-all-numbers or ignoring all input. It is not another whole no-op/idempotency claim.

Action ledger:6navigation +4endpoint create/read +6whole cycles×4 +6patch cycles×4 +12endpoint GETs (both endpoints immediately after each of the six Link refusals) +2final endpoint GETs =**72**. New afterstates follow every refusal and correction. Exactly29 raw exchanges (5navigation +24mutations) and43 client exchanges are planned; all current explicit IDs and8 fixture bindings. No alias or deletion is needed. Remaining unrelated Resources are outside the subject/endpoints “nothing changed” observation; do not infer a global store audit from a handful of point GETs.

## N — nested refusal and diagnostic locations,52 planned actions

Six fixed Resources at most: two Task endpoints plus one properties-case and one metadata-case subject per kind. Both kinds execute the same four cases:

1. **Whole properties:** an initially empty `queue`, one inadmissible integer at `/shallow` and another at a specifically authored object depth16 under `/deep/.../a~1b~0c` (raw key `a/b~c`). Expect both configured diagnostics, in this fixture's known evaluation order, with correct document-relative RFC6901 escaping and no Type/schemaLocation. Fixed-ID404, then the same body with both literals replaced by independently chosen admitted values under a fresh key; complete raw postimage and independent GET.
2. **Unapplied properties patch:** on that corrected Resource, append an object containing a nested array/escaped key at `/queue/-` (the valid initial properties include an empty queue). Inadmissible leaf9007199254740993 at a depth4 input subtree. Companion metadata change is otherwise valid. Expect the exact `/propertiesChange/0/value/...` input pointer in the message and no instanceLocation; full GET unchanged. Correct only the numeric leaf and key; the expected appended value and metadata are independently authored and GET-checked.
3. **Whole metadata (configured adjunct):** separate fixed-ID create with valid properties, an initially empty metadata queue and an inadmissible underflow1e-400 in an object/array at depth3 in metadata. Expect documentmetadata and the relative metadata instanceLocation; fixed-ID404. Correct only that literal to0 and key; raw response and independent full GET prove the numeric/document distinction. Nonzero exact underflow is refused, independently of the admitted negative-zero corpus inR.
4. **Unapplied metadata patch (configured adjunct):** append a nested object/array with an escaped member and inadmissible twenty-digit decimal to a metadata queue established by the valid create. Companion properties marker changes otherwise validly. Expect documentmetadata, `metadata: ... /metadataChange/0/value/...` and no postimage instanceLocation. GET unchanged, then correct only numeric leaf and key and check both resulting documents.

For whole creates, all other document fields are identical across the refusal/correction pair. For patches, the other document's would-be change is retained in both requests, so a premature partial commit is observable. Do not conflate document depth with JSON Pointer patch-path depth: `/queue/-` is shallow while the admitted value is nested. These are finite depth1/3/4/16 cases, not a parser-maximum test.

Ledger:6navigation +4endpoint create/read +8cycles×4 +8endpoint GETs (both endpoints after each of four Link refusals) +2final endpoint GETs =**52**. Exactly21 raw exchanges (5navigation +16mutations),31 client exchanges and6 new fixture bindings are planned. Complete prior records are refreshed after each corrected positive before the next refusal.

## Nine bounded negative controls

Use two explicitly separate app-only control tables: exact raw controls **1,2,5,6,7,8** wrap `executeExact`; client controls **3,4,9** use `RetryControl` inside actionExecutor and assert selected clientAction.input equals execution.input. Raw mutation Problems never pass through the client executor. Extend only the closed named new row/action/key mapping; no helper/executor/runner changes or configurable targets. All response controls act after the actual exchange and preserve its authentic wire observation. When changing raw response bodyText, recompute bodyOctets from UTF-8 byte length so the intended Problem assertion, rather than a stale byte-count artifact, is what fails. Control1 modifies the request before transport and preserves the original authored-body oracle; its actual serializer metadata must describe the sent request. Pin each exact intended failure set after inspecting the real row assertion shape; no “any failure passes” test.

The existing thirteen attribution/numeric client controls and three raw controls stay in their current arrays with their <=16 bound. New R/N controls use their own two tables with lengths6+3=9 and their own <=9 assertion; never append them to or increase the old bound. Preserve existing controls and failure sets.

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

Measured9c artifacts and fresh packed evidence (`read-update-attribution-review-repair/source-audit.json`, `packed-summary.json`):

| Artifact | Current bytes / nodes | Immutable factory/parser caps |
| --- | --- | --- |
| frozen baseline |3,626,922 /97,116|4,194,304bytes /100,000nodes; do not append |
| continuation |535,052 /14,286|4,194,304bytes /100,000nodes |
| shared fixture |29,055 /750|1,048,576bytes /100,000nodes |
| coverage |648,157 /13,494|1,048,576bytes /100,000nodes |
| index |242 /9|unchanged; no new shard or fields |

The immutable factory/parser/depth caps above do not change. Separately, existing app growth assertions are per-slice planning pins, and currently bind the *entire* continuation/coverage files to pre-A/B anchors. This revision explicitly replaces their whole-file growth scope with reviewed9c anchors, while independently freezing every old scenario and existing per-row/action/control limits. No static assertion is simply deleted without a replacement.

| Named proposed app pin | Exact anchor at9c | New-slice allowance | Final ceiling |
| --- | --- | --- | --- |
| NUMERIC_ADMISSION_CONTINUATION_BYTES_BASE |535052|450*1024bytes|995852bytes |
| NUMERIC_ADMISSION_CONTINUATION_NODES_BASE |14286|12000nodes|26286nodes |
| NUMERIC_ADMISSION_COVERAGE_BYTES_BASE |648157|65*1024bytes|714717bytes |
| NUMERIC_ADMISSION_FIXTURE_BYTES_BASE |29055|8*1024bytes|37247bytes |

Concretely, replace `bytes.length-114182<=650*1024`, `nodes(manifest)-3158<=24000`, and `coverageBytes.length-590847<=65536` in the old combined test with an explicitly named new-slice source-budget test using the four anchors above. Retain an additional explicit whole continuation node ceiling **27158** (=3158+24000), so the new14286+12000=26286 ceiling is stricter, leaving872 nodes under that historical total. Re-anchor the fixture growth assertion from28753/32768 to29055/8192; this is also stricter in total. The baseline whole-file hash remains unchanged. Preserve old per-row assertions rather than accidentally applying their old growth anchors to the new entire artifact.

Freeze old block hashes (raw bytes from four-space opening brace through closing brace):
- attribution9c:229073bytes, SHA256`1b81ae03b7aea896732e001291bb2ef92edd6bf5586c13a7e91afaf563f3e769`;
- restored C:114099bytes, SHA256`39eb2fe22d25eada2c43490f37b6cb5d66950bd90042af1ae954036426e07f2a`;
- same-value-no-op:191793bytes, SHA256`368c2d3d2aead864bfe925f8777b55ce9b09a107c81429820ac390d6d0a8550d`.
Baseline manifest SHA256 remains`c86b33535f25a1d117ce021a3bb70164f5635bdaee50758e3428c130ccc3227f`. Current continuation whole SHA256`69a8387f0459b35638f805e38bb9309ba585c956ed8e1a92bf4719801fd48aed`, coverage`46d55f94e2726f50de91f857bfb7dee5f09f6b6bb08281ecece669b707aeced9`, and fixture`07bbe646a5ce8436e8a746fadd22a9906013f0c28e28f608dcd9a0a84aa7b234` identify the reviewed starting bytes, not expected final whole-file identities. New final hashes/counts must be recorded before freeze and fresh execution.

Existing A/B tables remain A75/B73,combined148<=150, each<=75; existing16 controls remain<=16. R/N get a separate table R<=72,N<=52,combined<=124 and6raw+3client controls<=9. Their resources<=8/6, actual/authored requests<4KiB and estimated assertions<=1300 combined remain. Do not conflate the global artifact action limit with either slice's authoring ledger. The final source ledger enumerates all old and new groups so no growth falls outside an asserted budget.

Actual9c two-shard aggregate20.735s,69.265s remaining within90s. Whole five-file242-test invocation127.36s and1,309,296KiB supervised descendant-tree peak are historical measurements, not a forecast or owner heap bound. Keep90s aggregate,120s outer aggregate test,5s exchange,30s focused and240s whole-process supervision. Do not enlarge deadlines if the added controls overload them.

Measure concrete source/action/assertion/body/route counts before freeze; **stop/report** before dropping cases, weakening GETs, compressing away expected facts or exceeding the stated source allowances. Expected continuation route count49 (current35+14literal GET routes); independently enumerate exact final method/URL order and pin the freshly calculated digest. Baseline route pin stays unchanged. Update both the independent literal inventory and derived-digest assertion; no count-only replacement. Shared fixture and coverage digests change genuinely, so rerun both shards against the same fresh frozen source/configuration.
The app file is already large; this scope adds data/closed test cases only, no new generic executor abstraction or configuration surface. If concrete authoring cannot fit these bounds or needs a helper seam, report it for separate scope review.

## Owned files, validation and done-when

Exactly five files proposed: continuation manifest, `fixtures/read-update-reference-v1.json`, `test-support/read-update-coverage-v1.json`, `apps/bdptest/src/read-update-matrix.test.ts`, and `packages/conformance/src/numeric-model-catalog.test.ts`. Root owns eventual whole-list promotion. No runtime/server, schemas, normative catalogs, runner, executor, index, helper, TX inventories, Type artifacts or public entrypoints. Update continuation selected-ID/count/route assertions for five canonical rows while preserving old row bytes and C/small-set behavior. The numeric catalog test at9c explicitly permits only same-value-no-op in the continuation artifact: add exactly round-trip-admission and nested-refusal to that filename’s approved list. Keep the Read companion confined to read-v1, the standalone numeric catalog forbidden as executable identity, and all other manifest/row pairs denied. Add targeted denial controls for the new rows under an unapproved filename and under read-v1, and require the approved continuation to retain catalogId read-update-v1/manifestVersion2 so a changed catalog binding is refused. Preserve existing requiredProfile selection tests (the manifest has no separate profile field). Reuse local assertion logic in this test file rather than widen production loaders. Run the numeric/Read+Update/owned-wildcard/Transactional catalog tests so stale closed-world assumptions are caught locally. Common fixture digest changes require a genuine fresh execution of **both** shards; never assemble a pass from old baseline and new continuation reports.

Machine-evaluable completion predicates:

- Source parent is the approved ordinary integration; only the five authorized paths differ. Baseline manifest hash equals9c'sc86b3353…; raw C, attribution and no-op blocks each equal the accepted parent. Existing fixture values and nonselected coverage rows are semantically identical; new fixture bindings have no collisions.
- Genuine artifact-set factory accepts the unchanged129 ordered required inventory. Each new ID is owned once in continuation; every coverage binding equals its row's exact ordered assertion inventory and is nonempty. Both new whole flags false. With no independent promotion, counts are54authored/30whole/99uncovered/75unplanned.
- Static ledger meets the named9c-anchor source-growth assertions and immutable factory/parser caps separately; old blocks/hashes and A/B75/73/148/16 tables are unchanged, R/N72/52/124/9 have independent tables. Actual final file bytes/nodes/hashes are recorded before clean freeze. Numeric catalog permits the two new rows only in the approved continuation and the explicit denial controls pass. The remaining action/resource/key/body caps hold. All numeric mutations are raw UTF-8, use unique fresh keys, contain exact intended lexical tokens, and assert authored body/complete write. Every refusal has its independent afterstate; every negative is paired with a fresh-key admitted correction differing only in the intended literal(s) and key.
- Actual installed R/N rows pass all authored assertions with expected raw/client exchange counts; no public setup replay or cross-row dependency substitutes for row-local setup. Per-row full record baselines were independently checked before capture. All actual bodies<4KiB and raw digest projection remains withheld.
- Nine controls apply once, fail exact intended assertion sets, and leave no child. New whole-record or diagnostic corruption cannot hide behind assertion-free manifest metadata.
- Frozen-lock install, pinned Node24.16.0/pnpm11.20.0, typecheck/clean build before process tests, focused new/C/association controls, applicable artifact regressions/static/boundaries/bothsmokes/Readseal complete. Fresh clean packed five-file cohort uses exact frozen head and fresh payload/config/fixture/route identities; no8k workspace rerun required for this evidence-only slice. Record all failures and explicit exits.
- Both genuine per-shard results persist with separate digests and exact source/config association; required omission inventory remains complete. Cooperative abort/failure retention and cleanup controls continue passing. Preserve90/120/5/30 bounds, no token in reports, no success summary for failed/incomplete execution, no owned process remaining.

No true product decision or known missing runtime capability was found for this bounded proposal. Any concrete contrary properties-admission behavior discovered later requires a runtime-owner investigation. For inferred metadata-refusal behavior, first audit existing general-number semantics and configured diagnostic handling with the appropriate owners; the discrepancy may be an implementation issue or a specification ambiguity. Neither permits silently weakening the oracle, editing runtime or inventing new normative law within this evidence slice. Council should approve the finite partitions and metadata diagnostic adjunct; neither arbitrary-depth nor every-number universality is claimed by initial false flags.


## Review dispositions (all seven external findings, plus intervening9c scope)

1. **HIGH coverage anchor — accepted.** Original plan confused factory headroom with the exhausted old per-slice growth assertion. Re-anchor full coverage648157+66560=714717, preserve old rows, and bind the new budget test explicitly. Immutable1MiB parser cap unchanged.
2. **HIGH continuation growth — accepted.** Re-anchor bytes535052+460800=995852 and nodes14286+12000=26286; retain historical total27158 node ceiling. This is the root-approved planning allowance, not a parser/runtime-cap increase. Exact final measurement required before freeze.
3. **MEDIUM action/control pins — accepted.** Keep old A75/B73/148 and16 controls intact. Add separate R72/N52/124 and9-control ledgers; both sets remain visible in the final static inventory. Do not enlarge old constants or put new cases in old arrays.
4. **MEDIUM properties correction — accepted.** R create remains...992,0.125,1e300; R correction patches use...994,0.25,1e299. Both literal properties and companion metadata must change in raw response and independent GET. This specifically detects dropped properties patches.
5. **LOW control seam — accepted.** Raw1,2,5,6,7,8; client3,4,9. Closed named mapping only, exact executed-input/target/key checks, genuine wire observation before response substitution, recomputed UTF-8 bodyOctets, exact intended failure IDs and one application. No runner/helper change.
6. **LOW depth claim — accepted.** No receiver document-depth maximum exists; finite depth16 is a sample under64KiB request/properties and128artifact-depth caps. Patch pointer depth and selector depth are unrelated and are not cited as document-depth bounds.
7. **LOW metadata obligation — accepted with explicit distinction.** Refusal for N3/N4 is a configured adjunct inferred from general number semantics, not a quoted expansion of the properties MUST. Diagnostic metadata label/relative or omitted locations are directly specified. A contrary observation triggers existing-semantics audit rather than automatic blame or oracle weakening.
8. **Intervening9c numeric catalog ownership — accepted root addition.** Five-file scope now names the exact approved-list extension plus negative filename/profile/catalog ownership controls. The external report’s four-file feasibility finding described28a and is historical, not valid authority to omit this new test dependency.

No source edits or tests were performed for this revision. The prior plan and review remain historical. Root/council acceptance and an accepted ordinary parent are still required before implementation.
