# Comparison timeout context: native frozen code and evidence review

## Findings

**No remaining actionable findings.** The frozen implementation matches the accepted bounded contract, and completed source/control/installed evidence supports the stated slice. Preflight control reachability and exact-optional typing corrections are closed in the actual executed bytes. The final summary's inherited executed:false labels were corrected before signoff; independently verified that its only correction is five executed:true fields, with original summary preserved and no rerun/source change.

This is an independent frozen code/evidence review, not a substitution of earlier preflight for code council. Root retains external review, exact CI, publication/integration and cumulative public-admission gates. Default installed comparison-timeout sensitivity remains unobserved.

## Exact review basis

Checkout `/Users/dbox/repos/wt/vickie/vickie-comparison-timeout-context` is clean at30035b88380b7cbb3abd5359787db236b70345c7, treecdee4bc7e31511cde094d17c045ea618163b6d7a, ordinary parentc555fe01df5cd20454fb0d663431df816c1270aa. Full parent-to-HEAD diff: eight files,557 additions/35 deletions. Independently verified final file hashes against source.json; all four production and three test files equal accepted preflight bytes. Only the already-owned progress document changed after validation, accurately replacing draft claims with bounded results. Parent/main tree equivalence is integration context supplied by root, not rewritten execution ancestry.

Read actual full diff and relevant surrounding store/service/runtime, accepted v2 plan/dispositions/closure, root authoring/measured acceptance, original/corrected preflights, exact controls and executed-source snapshots, source JSON/logs/exits, build/static logs, both packed cohorts, supervising scripts, final report/source/summary/cleanup and root independent audit. Reviewer performed only read-only source/artifact inspection, hashes, in-memory patch reconstruction and archive/payload arithmetic. No product code, compiler, test, build, service, installation or network operation was executed by this reviewer.

Final author artifacts independently hashed:

- report.md:f79e1f5468c492f57985e4b9f2e2077d3a67af9dc50fbe74d1382291f6cbc9f4.
- source.json:562896d01e09f070ee741c40b781973971310443a09fc35f9509507fdca172e5.
- summary.json:368397ba571a32f249b3b7cc0346a310febc993c8d78722d33a1ee90ccb0d4cc.
- cleanup.json:62618dee6b325d3b404b4e0c12b6b1d1f18f39725e6c22bd2e3abd05d636d128.

Summary source/check/packed records match their underlying files; its cleanup list matches cleanup.json children. The corrected execution flags preserve original52beb6516fc5471ffac1846bb8090f8cf0476143dab5ebb41f306d54aca2995f, with a separate correction record.

## Production and normative correspondence

General spec5768–5774 requires context on Transactional Scope-bounded responses. Applying it to the authenticated known-Scope healthy comparison deadline is consistent with the existing3994–4026 laws: one finite deadline, no waiting transaction, no receipt disclosure/second mutation/reservation retraction,503 temporarily-unavailable/after-delay and unchanged epoch-stop behavior. Unexpected internal fault handling remains the bodyless500 exception at4588–4592. No unresolved minimum or Event decision is needed for this selected repair.

`transaction-comparison.ts:49–76` preserves epoch-before-deadline order and one original deadline. Only the healthy observed deadline invokes the bound callback, synchronously with no intervening await; waiting stays outside owner/database entry. Unobserved full-store callers return the original bare temporary result. Observed deadline and owner-stopped are distinct private kinds.

`transaction-store.ts:1725–1741` performs final persisted-epoch comparison and current-position capture inside one enter plus existing atomic boundary. Context is freeze-copied and cannot become a cached earlier receipt/startup tuple. It performs no key lookup, reservation change or admission. Normal closed returns stopped; reentrancy/fencing/metadata failures that reach the operation propagate through established mechanisms. The earlier comparisonEpoch closed shortcut can still mask background maintenance failure; that real pre-existing residual is explicitly unrepaired. Exclusive owner lifetime has no supported live epoch rotation, so the changed-epoch test is correctly defensive.

`transaction-submission.ts:71–78,123–134,161–187` explicitly excludes observeDeadline from inherited caller options and overwrites even untyped extra input after spreading options. The private method-or-undefined annotation resolves exactOptionalPropertyTypes without exposing service/runtime configuration. Observed refusal propagation does not add enumerable fields to the unobserved submission result. Existing active-service work retains the original promise; no post-submit observation or detached owner work is introduced.

`transaction-http.ts:355–361` emits current context only for the observed deadline, requires its tuple, and lets a missing tuple take the existing sanitized internal-fault path. Owner-stopped remains distinguishable and bare503. Canonical Problem, retry, Retry-After and previously repaired direct/sequence/schema paths remain intact. Admission ordering and earlier unrelated exclusions are preserved, not newly declared normative exemptions. No package export, runtime/service configuration, store format, evaluator, parser, schema/catalog or public advertisement changed.

## Source proof and controls

Actual selected results: store6, submission5, HTTP2, dependency3, alias2 =18 passed,287 filtered,zero failed. These are explicit literal-title selections, not complete suites. Build preceded emitted runtime consumers. Typecheck/build/format/lint/boundaries and historical Read verification have actual exit0; lint retains666 warnings/43 infos. Read verification retains its own historical run/seal identity. No broad workspace/matrix suite or local smoke is claimed.

The new12-request HTTP test establishes a genuine unresolved original sequence, commits independent Bead b while comparison waits and then advances the private clock to the original deadline. It requires canonical503/no Location/currentpos1 with epoch/view unchanged from an actual initialpos0 observation. Before original release, complete graph pages and exact one-fact Event content establish no timeout effect. After finally drains the original sequence, validated replay and finalpos3, full graph equality and five exact ordered Event facts exclude duplicate execution. Durable readonly principal/key/receipt inventory separately proves only creator/dependent/independent rows and full retained replay equality; graph inspection is not treated as receipt absence.

The Event substitution is valid: finite events() traverses retained groups under item/byte bounds, while changes/stream packets have different single-group behavior. Five small facts fit32 items with next:null. Actual unowned self-Link effects are created plus source/target linked facts, yielding0→1→5. This does not settle metadata-only Event membership or prove change-group pagination. Dynamic revision/provenance/allocated Link identity are validated observed values; expected semantic contents/order are independently pinned.

Owner tests establish one BEGIN IMMEDIATE/COMMIT, exactly two metadata prepares, immutable context after progress, final epoch mismatch, normal close, reentrancy, fencing and metadata-fault recovery. Submission tests establish observed-only callback invocation/current context and untyped callback exclusion. Existing dependency/alias tests preserve deadline/retraction behavior. Their direct comparison assertions use toEqual; strict undefined-field sensitivity is specifically the submission-boundary test, not a broader direct-comparison claim.

Independently reconstructed every derivative and matched executed-source hashes. All five builds exit0; all five selected tests exit1 with exactly one intended assertion error, no extra cleanup error; adjudications pass:

|Control|Actual discriminating failure|HTTP path count|
|---|---|---:|
|remove-deadline-wrapper|required epoch presence false versustrue|4|
|stale-deadline-position|pos0 versus independently expectedpos1|4|
|lose-required-tuple|sanitized500 versus healthy503|4|
|leak-unobserved-fields|strict submission result sees enumerable undefined context|0|
|omit-final-epoch-check|observed result versus owner-stopped|0|

Source transport is16/two listeners:12 dispatch/completion assertions in the rewritten scenario plus the unchanged four-request path. Controls use12/max36 across three listeners. Control counts follow deterministic paths to the retained named assertion, not a separate raw wire recorder. HTTP finally releases held callbacks and awaits/validates the original sequence before propagating errors; actual Vitest results retain the intended inner assertions after AggregateError flattening. Generic failure/cleanup labels are not the oracle. Source12-request bodies remain277B maximum/629B aggregate. The response estimate39936B/479232B remains prospective, not measured wire/RSS or an enforced Fetch cap.

## Installed evidence, binding and cleanup

After freeze the supervising script rebuilt/packed/installed actual30035 and ran only unchanged direct14 and sequence13 cohorts. Actual result files pass one direct case and two sequence-file cases (one confinement case makes zero requests). Exactly27 requests/two children, no retry/replacement/predecessor timeout cohort. Direct observations retain409/406/413 context atpos2 plus key reuse throughpos4; sequence retains initial/refreshed413/406 contexts atpos2/pos4, immutable receipt root, complete graph/Event state and key reuse. Both claimEligible values remain false. These are ordinary regressions, not default installed timeout observations.

Independently checked source/entry/tar hashes and every archive member against installed files:63 protocol plus250 server members. Independently recomputed the full14-visit runtime dependency closure, matching both actual served payload digests:

- payload:2c62c66724c2c48fcd1de19360abbb90a3d5bb7c37257a095af752ba9e0b322d.
- server tar:b9ebfae7b6fd2f5602688fadf0110da3749d5af012fdf4599867c6af2f84bf35.
- protocol tar:4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227.
- entry:bcf153170f953f9e28c6c89f1153be907f64df289e64f3ad929cad4a4a54c47c.

Entry equality with earlier cohorts reflects unchanged bootstrap, not equal payload. Both actual contracts/fixture digests match; fresh authority/configuration/policy/authentication identities remain distinct. Source attribution is supervisor-verified checkout/package binding and a caller label, not child attestation. The report now states this precisely.

Read all27 recorded response observations and matched them to their completed cohorts/postprocessors, expected route/status/context sequence and unchanged source assertions. Direct request bodies total5518B/max4616B; sequence11970B/max5728B. All retained byte counts fit existing caps:65536 request,16384 headers,131072 response,8192 URI. Existing5s exchange/90s scenario/120s test/10s close limits remain; supervising150s command cap does not widen inner contracts. Evidence retains body/wire fingerprints and contexts, not complete wire transcripts.

PIDs82412/82445 each record normal exit0/no signal and temporary-directory absence. Author postprocessors check actual PID absence; root-packed-audit independently confirms both child PIDs absent and recomputes the same package closure. Directory absence is asserted by the installed tests; the postprocessor validates that report, without a separately retained path inspection. Packed supervisors are recorded absent. Disposable review directories remain intentionally as evidence, with no claimed running service. This distinguishes process absence from exit-field inference.

## Limits and conclusion

No additional execution or source correction is requested by this review. The implementation and evidence support fresh context at the selected healthy observed deadline and preservation of surrounding behavior. They do not repair the maintenance-fault masking residual, prove ordinary installed timeout sensitivity, add live epoch rotation, decide minimum/Event questions, enable cold discovery or establish cumulative Transactional/public admission. External council and exact-head CI remain root-owned independent gates.
