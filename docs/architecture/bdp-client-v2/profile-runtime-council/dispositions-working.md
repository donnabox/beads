# Working council dispositions (not final)

Core72bdaad02b6fb2b803d4af7a8c4f90ada2f0d2d5 vs a9bcf9ed1d869aaba2cb49fe9cca10e3b8d52a94. Alias delta59915ea916a210f5ce36611eb6d799f7c3124056. CI budget21654be3cae51d50e75e817ef8ac57735b8ee4ea separately native-only. Later e538e23 changes outside these scopes until supplemental review.

## Claude01

F1: Accept editorial clarity gap, downgrade High to low/medium: old all-three optional historicalResolution sentence removed, table itself preexisting omission. Present optional History says cumulative profiles inherit capability, and schema explicitly admits it in allthree definitions (1135,2758,5048). No actual schema admission regression. Add optional row and name Scope-history fences precisely. Parent assigned separate reviewable repair.
F2: Real preexisting tension: no-op successful key tombstones grow without Resource effects, contradicting storage-bound rationale. Do NOT accept proposed forgetting behavior: rerunning the formerly no-op patch after later state changes can create effects. Product/retention decision deferred; no runtime change authorized by council.
F3: Confirmed regression-guard gap: owner0 permissively accepts every shared/endmatter anchor, including Transactional rows for Read scenario. Parent assigned ownership/allowlist repair; no current catalog wrong citation established.
F4: Confirmed guard fragility: explicit anchors unindexed, whitespace slugger diverges from rendered Part heading. Citation validator may separately reject invalid anchors; no current normative loss proven. Parent assigned aligned anchor parser checks.
F5: Baseline alias equivalence oracle: unfiltered lookup is explicit normative design, current disclosure guards do not hide equality bit. Worth product/security documentation decision; do not silently change comparison or add new authority behavior.
F6: Baseline changeContext admission phrasing. Schema admits deleteLink changeContext independently of Type; inert input still participates in explicit mutation semantic identity. No new runtime bug proved; optional future clarity, not unauthorized semantic change.
F7: Reject missing fresh seal as current defect. This map explicitly describes earlier editorial pass baselinee936; packet03 current artifact and source pin establish fresh Read seal. No entire-branch validation inferred from earlier pass record.
F8: Historical implementation slice doc wording could clarify internal lifecycle versus then-absent exposed sequence receiver. Current runtime later added receiver; no current capability regression.
F9: Reject skip-if-absent proposal: explicit development citation audit is intentional and missing expected file should fail. No production admission granted by metadata test.
Questions: schema historicalResolution resolved above; requiredProfile validated by catalog loader; other omitted-context questions are not findings. Read alias rules resolved by present PartI: Aliases254–269 and Alias resolution1401 onward; no forward PartII dependency required.

## Gemini01

Discovery historicalResolution table ambiguity agrees with Claude01 F1. Table wording itself baseline; restore explicit optional capability clarity. TX patch/sequence limits question resolved by actual transactionalAdvertisedLimits schema. Historical docs read as current state rejected using current runtime/artifact.

## Gemini02

Pinned update/delete crash: false positive/unreachable. durableResourceReference schema1222 is string; updateBeadRequest1501 and analogous fields require it; carrier read-update-values204–216 validates before admission. Pin objects are permitted for endpoint Reference fields, not Resource subject locator strings. Do not widen accepted subject syntax to fix impossible input.
Q1 null alias: capturedMemberAlias member-identity482–494 converts null to undefined.
Q2 named noncreator: sequence schema closes name to createBead/createLink; parser carrier validation precedes dependency execution. No completed-prefix missing creator path for valid carrier.
Coverage: provider explicitly truncated transaction-store480–808 and transaction-sequence; allthree receive02b supplement (87KB) to close exact missing source scope.

## Gemini03

Accept low development routing defect: POST unknown /development path returns405 Allow GET,HEAD atdevelopment-runtime209; intended404 branch213 unreachable. No public BDP claim affected; parent notified.
Questions: request-too-large is supported Read problem code; protocol/schema internal faults intentionally500; cohort serializer needs omitted context but no identified mismatch. Full evidence exchanges omitted, summary only; no manual exchange-level review claim.

## Claude04

F1/F1b accept bounded observation gap: status-only intermediate alias mutations do not directly prove state preconditions for recorded-target/pin/miss retained-replay legs. Global no-op-all-tests claim overstated because fresh-locator probes discriminate some no-ops. No runtime defect established. Parent assigned direct state readbacks.
F2 accept cheap observed-sample noncollision assertion; deterministic collision skip remains explicitly unqualified. Existing count is bounded mapping, not full coverage. Parent assigned allocated ID != reserved alias segment check.
F3 accept naming precision: new URL(id).href===id is URL serialization stability, not complete BDP safe-segment validation. Parent assigned honest name/limit.
F4 reject semantic conflict: Link endpoint identity/reference denotation versus idempotency semantic request identity are distinct; explicit opaque pin byte comparison governs retry. No new behavior decision.
F5 reject demand for category-specific diagnostic string: fixed message names canonical Bead-target constraint; normative text does not prescribe distinct category wording. Limit already fixed development message, not authority-wide wording.
F6 incomplete omitted-row context: existing inventory and remaining semantic-idempotency bullets retain open work. Optional prose improvement only.
F7 low traceability completeness suggestions: add defining alias-path-taken and retained-disposition clauses where useful; broad citation alone not promise every clause proved. Parent may strengthen alongside witness edits.
F8 reject unsolicited removal of unused helper branch (YAGNI decision excluded).
F9 placeholder intentional; missing binding fails against nonexistent target, not silent success. No defect.
F10 scope canonical trailing slash required; invalid scope makes explicit inScope fail. No reachable bug.
Questions: full delta exactsixfiles, helper uses SDK transport, controlled assertion evaluator previously independently tested269tests. All23catalog noncitationfields sameauthorhead0965971b; old17plans unchanged. Numerical counts checked natively23plans221actions705assertions48bounded33omitted81total. No runtime/schema/evidence changes in delta.

## Gemini04

Reported new URL identity failure as uncaught process crash: false impact. runner.ts860–915 wraps action promise, records redacted executionError and propagates ScenarioActionExecutionFailure into scenario failure. Bad receiver IDs fail the diagnostic, not silently pass or crash process. A false-valued URL-serialization observation could improve failure precision, but no correctness blocker. Provider independently accepted explicit deterministic-collision/hidden-subject/retention exclusions and current diagnostic-only claims. It did not identify Claude’s alias precondition gap; native accepts narrower gap as useful strengthening, not consensus runtime defect. Full sixfilediff and selected clauses reported read; no truncation.

## TX supplement: completion and dispositions

Both external reviewers completed the full packet. Gemini's successful direct-answer retry is `02b-tx-tail-gemini-retry-gemini.stdout.txt`; first empty/permission-denied attempt remains excluded. Claude is `02b-tx-tail-claude.stdout.txt`. Each reports spec 3793–4099, store 1–808, sequence 1–136, batch 1–538, with no truncation. Native reviewed the same source.

### Gemini 02b

1. Reject stop-on-member-failure change. Ordered sequences intentionally continue after member problems. Private stepper requires caller execution or explicit abort; production `runTransactionSequence` owns the loop and continues regardless of response consumption. Dropping the private stepper is an embedder misuse/liveness qualification limit, not evidence that public client disconnect strands keys.
2. Reject compulsory comparison wait for already-normalized pending receipt. The quoted wait concerns unresolved dependency reservations, explicitly absent from this frozen private slice. Equal normalized requests can receive the pending receipt under the explicit duplicate rule.
3. `retract` leaves its claim in a WeakMap. Later execution is correctly fenced by missing owner row; weak keys do not leak by themselves. Deleting the weak entry would improve consistency/error precision, but no committed-state or ownership violation exists.
Questions: evaluator stores local source/target indices (resource-evaluator1045–1046) and queries local IDs; incident lookup consistent. Link-scoped Event subject versus Bead projection is separate; see full Event source rules. No new behavior accepted from omitted context.

### Claude 02b

A1: Private split admission/execute and manual stepper permit caller-held pending claims indefinitely; public scheduling/duration guarantees are not implemented/claimed by this slice. Record as a prerequisite before exposing a public Transactional receiver, not an automatic timer-expiry design or proven current public defect. Startup retraction and explicit abort are already tested. A finite advertised duration cannot be inferred from this private API.
A2: Potential wording/intent question sent to parent: turn-time classification can admit a formerly known foreign key after failed detail is forgotten or another owner retracts. Existing test transaction-sequence.test511 explicitly expects failed-detail forgetting then execution at turn. Do not silently freeze all projections at carrier admission based on this review; distinguish general turn-time state classification from direct unresolved comparison wait/retry. Resolved by inherited spec3234–3239 and3255–3272: initial member-turn classification sees current unknown state; no direct comparison wait/retry occurs. See sequence-turn-admission-disposition.md.
A3: Reject silent quarantine/continue proposal. Fail-closed maintenance rollback/fencing is intentional; alias targets are canonical strings (alias-evaluator167–181), not pin objects. A corrupt retained result is an invariant violation, not a supported result shape. No reachable legitimate row causing timer failure demonstrated.
A4: False positive. Spec4160–4168 explicitly says alias effects occupy no position/Event/group and receipt may omit effectPosition. Locator-only semantics are deliberate.
A5: False positive. Spec5050–5076 retains identities for every creation operation, exactly the vanished created results; supplied IDs still allocate a durable Resource identity. Same semantic key ensures presented creation kind matches retained creation. Successful creates return created; shape failures remain invariants, not omission fallback.
A6/A7/A10: No current defect established; self-check wording, fixed private limits duplication, and diagnostic labelling are low-priority maintainability nits. No unsolicited refactor/API expansion authorized.
A8: No reachable timestamp mismatch in this slice: recordChangeContext=false; evaluator samples commit time only when a native metadata recipe requires it (resource-evaluator978–986). Future context support must obey the existing attribution contract.
A9: Valid low operational diagnostic issue: cleanup failure in executeClaim catch can mask the original exception, unlike the sequence AggregateError path. State remains fenced/rolled back; preserve both causes in a later bounded repair if desired, not a readiness blocker.
A11: Empty-array changeContext rejected by closed input schema before semantic normalization; no valid request collision.
Omitted context questions: semantic-identity sorts object names and canonicalizes numeric values; endpoint records canonicalize URIs while local indices use local IDs; updateLink does not mutate immutable endpoints; alias target validation rejects alias chains; carrier schema rejects duplicate keys/noncreator names/creation id @ forms. Private modules are absent from package barrel and write claims remain off. No whole-branch or full-profile claim from this packet.

## Claude core02 (full packet completed, no truncation)

A1 repeats private pending lifetime point from02b. Public scheduling/duration ownership is a prerequisite before claiming a complete Transactional service; a deliberately split private owner API is not proof of a client-disconnect bug.
A2: Valid operational hardening concern beyond the low diagnostic issue: transaction-store atomic uses unconditional rollback and lacks recovery-store's guarded rollback/sticky fence after uncertain I/O or commit failure. Ordinary injected callback faults roll back correctly; no actual SQLite fault reproduction or committed-state violation supplied. Record before broad durability qualification. Do not claim general power-loss safety.
A3: Valid construction-time configuration validation gap: store minimumRetentionMs7days and executor minimum/actual1day each initialize, then fail per request safely with retention-too-short. Current development wiring matches24h. Candidate has no such configured outage; expose/cross-check floor in a bounded later hardening slice.
A4: Fixed private evaluator/store values match16 diagnostics and65536bytes. Arbitrary changed caller limits could drift, but there is no current supported path with divergent values. Record limitation before configurable/public use, no present over-budget receipt demonstrated.
A5: Expiry boundary inference not established: duration D naturally completes at t+D, while Transactional expiresAt wording promises through its explicit instant. No source requires a duration interval inclusive of endpoint; changing RU comparison would be a policy clarification, not automatic bug fix.
A6/A8: Invalid context array and duplicate member keys fail carrier schema/parser before admission, as previously resolved.
A7: False depth calculation: batch and sequence roots both put operations at depth2, singleton begins at2. Reserializing a member reconstructs the same wrapper rather than nesting it one layer deeper.
Q1: Admission-time alias capture for a direct TX request is explicit normalized identity state; execution uses the same witness. Review does not prove nonserializability merely because an overlapping alias repoint completes first. Alias state is outside Resource history, and sequence turn-time alias support is explicitly excluded. No new resolution-timing behavior accepted without complete contract proof.
Q2: Resolved by resource-evaluator615–624 loadVisible and703–711 endpoint handling: every local Link endpoint passes visibility before creation. isResourceVisible548–574 checks current Resource, owned Links and endpoints. A withheld creator binding is not permission to disclose a hidden endpoint. Fixed all-visible private service still does not qualify the full mutable-policy matrix.
Q3/Q6: Parser rejects noncreator names/forward bindings and removes operation/idempotencyKey/name from op.input; synthetic keys never enter identity/store.
Q4: Resolved by first-turn classification clauses, confirmed parent source review. See sequence-turn-admission-disposition.md.
Q5: Numeric admission and retained startup gate use the same serialized diagnostics list budget under existing producer tests; no concrete divergent byte accounting identified.

## Gemini06 — final repairs

Provider completed all supplied diff hunks and extra full files; no truncation. It explicitly confirms prior alias-witness and ownership/parser gaps are closed.

1. Progress document253/797 counts describe the earlier pending/interleaving slice. Current README and manifest are264/827; a current combined-count note would reduce ambiguity, but rewriting the historical validation as if it had run827 would be wrong. Gemini's approximate826 arithmetic is itself incorrect; native counted827. Low documentation clarity, not stale current artifact or attestation failure.
2. Multiple stacked explicit anchors before a heading are not supported by the documented immediate-anchor subset; earlier anchor fails citation validation. No such source exists in current spec. This is a future feature request, not a permissive boundary bypass or present regression.
3. Reject wrong-root allocation claim. The shared uniqueness namespace concerns alias path segments, so allocated Bead identity must differ from beads/reserved when alias/reserved is live. Comparing a Bead URL directly to alias/reserved would always differ and fail to test the requirement.
4. CRLF concern disproved with exact expressions on pinned Node24.16.0: heading matched Heading and explicit anchor matched old for CRLF input. JavaScript multiline boundaries recognize CR as a line terminator.
