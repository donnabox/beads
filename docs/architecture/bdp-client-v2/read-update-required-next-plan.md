# Next installed Read+Update evidence: singleton admission boundaries

Plan only for council; no implementation or coverage promotion authorized by this document. Sole plan remains the fork design document. Inspected exact integrated `4a8607ac2131aae4706d00684a2f3f1185cb37e9` in `/Users/dbox/repos/wt/vickie/vickie-retry-runtime-integration`, the proposed next authoring basis once root’s full validation passes. Its delta from `a899872efcccc1e7871746c8c8cfb0eeefdc8db1` affects TX configuration/runtime and related tests; the cited RU receiver/preflight, manifest, fixture and app harness bytes are unchanged. All source references below are to4a8607a. Prior a899 packed timing and payload results are historical comparison only, not integrated-source verification. Root’s fresh integrated pack must establish the new payload/baseline.

## Inventory and choice

Actual coverage has **129 required IDs,45 authored observations,26 whole contracts,19 authored partial contracts and84 omitted rows;103 uncovered**. The raw omission count is not a list of missing runtime features. Key/media receiving behavior is already implemented but lacks row-local required-contract evidence. The app test at `apps/bdptest/src/read-update-matrix.test.ts:1331` injects a missing-key case into create-bead; that ad hoc observation does not bind the missing-key required ID and does not cover malformed/repeated keys.

Propose **two substantive admission rows**, using real packed installed service, actual raw requests, immediate afterstate and corrected retries. They answer whether a client may safely fix malformed HTTP input without an accidental mutation or captured key. No Selector complexity, metadata-only incident Link, Types, provenance, optional History, IAM or protocol decision is needed.

| Required ID | Exact catalog requirement and source | Proposed credit boundary |
| --- | --- | --- |
| `read-update.singleton.idempotency-key-required` | Catalog `#idempotency-keys`: key outside profile, omitted singleton field or repeated field is `malformed-request` before execution. Spec3171–3179 defines ASCII `[A-Za-z0-9_-]{1,256}` and rejection rather than choosing an occurrence. | Whole-contract **candidate** for this selected requirement after the corpus and independent oracles below pass and citation audit agrees. No credit for namespace isolation, semantic identity, concurrency or retention. |
| `read-update.singleton.unsupported-media-type` | Catalog `#readupdate-problem-details`: mutation target with non-`application/json` or undeclared media type is `unsupported-media-type`; parameters ignored. Spec3432–3434; negotiation distinction3587–3592. | Whole-contract **candidate** for this selected requirement, including positive parameter-ignored behavior and the sequence target because the quoted requirement says mutation target. No Accept/replay precedence claim. |

No flags change during planning. Author initially keeps both partial until the cited-section review accepts completeness; even two accepted promotions leave101 required IDs uncovered and `claimEligible:false`.

## Existing implementation and harness, no new mechanism

- `packages/server/src/development-runtime.ts:78–83` inspects original header multiplicity, case-insensitively, and refuses repeated fields;443–466 checks request media, missing/stray key, then calls the actual mutable owner. It consumes all request bytes before admission. `packages/protocol/src/read-update-preflight.ts:39–54` validates key grammar before preparing a singleton.
- Existing manifest v2 HTTP action supports exact repeated header lines, credentialRef and authored UTF-8/base64 body (`packages/conformance/src/executable-manifest.ts:48–80,917–985`). Use those actions, not a new client operation or fetch that coalesces repeated headers. Header source forbids secrets/transport-owned fields; credentials remain the operator handle.
- Existing app route assembly (`read-update-matrix.test.ts:172–250`) admits exact manifest HTTP routes and all nine POST routes. Raw negative probes may target `{binding:"operations",path:"create-bead"}` as existing carrier probes do. First assert the observed directory value matches each exact fixture route being tested; do not pretend that a hardcoded path is arbitrary cold-client navigation. Corrected client actions follow the actual captured directory target.
- Existing `installed.exact-request` executes successful corrections and exact Bead GETs. Existing raw HTTP GET actions can observe bounded Bead/Link/alias collections. No runner, executor, public exports or installed configuration seam is needed.
- Reuse existing response schema wrappers without weakening normative definitions. Expected status/Problem fields are independent authored literals; state expected bodies derive from earlier successful captures or predetermined fixture values, never from the response currently asserted.

## A. Key admission corpus and state proof

1. Start from ordinary discovered Scope/directory/Type fixture. Create a small explicitly identified graph and alias through real mutation commands. Capture complete baseline current records and complete collections, with `next:null` so equality cannot hide a later page. Use no seed/store injection. Choose a fixture small enough for one actual page.
2. Exercise missing key against all eight singleton targets using otherwise valid operation records. Prepared updates/deletes/alias changes must be real effects if executed; creates use distinct absent IDs. Prove the applicable state/collection equals its prior full value immediately after each refusal, including absence of proposed creates and no changed revision on prepared targets.
3. On create-bead, malformed field corpus: empty;257ASCII characters; quoted token; internal space and tab; invalid punctuation/comma-combined value; non-ASCII Latin-1 byte permitted by the existing raw transport. Do not treat HTTP field-value surrounding OWS as a token grammar witness—the HTTP parser may trim framing OWS. Confirm authored exact line/value metadata and actual request body byte count.
4. Repeated identical and repeated distinct valid keys, with a mixed-case field-name variant, must all refuse400 `malformed-request`. Both distinct occurrences need a recovery witness: submit each key separately with a **different** valid create record/identity and require success plus exact GET. This distinguishes neither-occurrence-captured from choosing/reserving either one. The original refused identity must remain absent. Identical-key repetition similarly uses a changed corrected record with that exact key.
5. Positive boundary controls: one-character and256-character valid keys plus alphabet representatives execute successfully with exact final GET; this prevents a reject-all implementation passing the negative corpus. Missing-key correction uses a newly supplied key; do not call that same-key proof. Invalid token itself cannot be reused lawfully; corrected spelling is only a no-normalization/capture observation when a specific plausible normalized valid spelling is deliberately tested. Do not imply it proves every normalization is absent.
6. Every negative requires actual HTTP400, normative Problem schema/type/code/status/retry and full deterministic fixture Problem equality. Successful writes require exact outcome/body shape and observable postimage. Initial and final collection proof excludes unnoticed allocations within this small observed graph.

## B. Media admission corpus and state proof

1. Valid JSON operation bytes but absent Content-Type, `text/plain`, and `application/problem+json` yield415 `unsupported-media-type`; give each probe a valid fresh key except sequence, which correctly has only member keys. Assert before/after complete current state and proposed identity absence.
2. Missing Content-Type against each of eight singletons and sequence ensures every mutation route applies the common gate. A sequence body should contain at least two independent would-be effects so “no prefix executes” is directly visible. Reuse the small prepared graph, not a separate authority implementation.
3. For each independent media refusal family, resend the **same valid key with a changed valid body** and correct media. Require fresh success with changed expected state, not the original refused body or retained failure. Sequence reuses original member keys with changed member inputs; all succeed. These controls prove the refusal did not capture a durable disposition under those keys.
4. `application/json; charset=iso-8859-1` over ordinary UTF-8 JSON bytes and a second irrelevant parameter spelling succeed; parameters do not switch the decoder. Include a non-ASCII JSON string in a permitted metadata value and assert exact UTF-8 semantic value after GET. Do not fabricate alternate byte encodings or infer additional content-encoding behavior. Ordinary `application/json` success remains the baseline.
5. Refusal assertions use actual status415, exact fixture Problem and schema; parameter acceptance requires actual success and correct current state. Keep Accept fixed to an acceptable value so request-content behavior is not conflated with response negotiation or retained-key admission ordering.

## Independent controls and evidence fidelity

- Existing raw-action sensitivity idiom can change one refusal's expected status/code and prove the named assertion fails. More usefully, replace a duplicate-key raw request with a single valid key while keeping its refusal oracle: actual200/changed state must make it fail. Replace a257-character key with256 to cross the actual boundary. Change a parameter-accepted request's media to `text/plain` while leaving success/state expectations intact: actual415 must fail.
- Use existing post-wire client control style for a corrected successful result replaced by a conflict and a later GET replaced by its earlier captured state. Require exactly one application and its named immediate assertion failure; preserve the newly added executed-input deep-equality guard.
- No need to modify the shared raw executor or add raw response corruption machinery. Raw HTTP cases use actual input perturbations and independently authored oracles; exact wire observations remain real. No passing synthetic response, Type lifecycle or fixture mutation hook.
- Bind new fixed GETs as named fixture entries only if needed for existing client afterstate controls. Update **both** ordered route digest pins using actual route insertion/dedup order, never a set-only comparison. Keep catalog ordering, all129 requirements and all existing45 scenarios. Capture new manifest/fixture/route digests and exact clean source in a fresh packed run; unchanged payload may remain byte-identical but must be measured.

## Files and non-overlap

Expected owned files only:

1. `packages/conformance/matrices/read-update-v1.json` — two row-local scenarios and fixed assertions/captures.
2. `packages/conformance/test-support/read-update-coverage-v1.json` — exact bindings, explicit credit limits; promotions only after review.
3. `apps/bdptest/src/read-update-matrix.test.ts` — bounded raw input counterfactuals, existing client post-wire controls, count/route pins and focused tests.
4. `packages/conformance/fixtures/read-update-reference-v1.json` — minimal named fixed read routes if needed; no new installed contract/policy.

No modifications to catalog, schemas, runner/executor, H1/TX harness, M1 runtime, stores/evaluators, shared Read admission or public profile gates. If actual repeated-header transport or configuration cannot express a case, report the exact limitation before widening scope. Source inspection currently reveals no such blocker.

## Done when / budgets

- Both actual rows execute from fresh disposable packed installed authorities; every negative has state proof and every claimed noncapture has a discriminating corrected retry.
- Full declared malformed/repeated/key-boundary and media/parameter corpus is accounted for, not skipped. Complete collections explicitly terminate. Controls fail the intended immediate oracles on real wire/input perturbations.
- Independently audit the two catalog-selected requirements and whole-versus-partial disposition. No credit spills into other required IDs. If any cited obligation remains unobserved, preserve partial flags and name it.
- Build before process tests; frozen Node24.16/pnpm11.20; focused rows, cumulative binding tests and fresh packed four-file suite. Compare a899’s historical45-row aggregate12.751s with root’s fresh4a8607a baseline; do not relabel the former as evidence for the latter. Keep per-exchange5s, aggregate90s/cooperative cancellation and120s outer budget. If approaching them, report measurements rather than silently trim requirements or widen deadlines.
- Clean local commit, exact source/configuration/payload binding, explicit exits and child/store cleanup; external report states all omissions. Root owns council/integration/publication.

## Closest subsequent candidates and actual constraints

- `singleton.no-local-bindings`: current bundle and parser support bare/pinned refusal; a separate small row can cover reference-bearing singleton fields and corrected same-key success. It is deliberately queued to keep this slice focused on HTTP admission.
- `singleton.method-405`: current installed CORS-disabled method branch is available, but selected catalog requirement also contains the CORS-enabled exception. Do not grant whole credit merely for all current405s; current configuration is not an enabled-CORS fixture.
- `discovery.limits` and `operation-directory`: genuine runtime support exists; whole credit needs truthful limit applicability and directory semantics, not just schema/successful nine-target navigation. Existing cross-row observations are not row-local coverage.
- `expected-revision-race`, pending/concurrent retry, durable boundary and retention rows need concurrent/process/reopen/controlled-retention observations, not sequential request substitutions. They are not absent solely because a row is omitted; avoid adding hooks under this four-file slice.
- Hidden-subject, authorization-view and other-principal rows need actual supported policy/credential fixture diversity. Do not weaken current fixed-principal evidence or infer them from operator200s.
- Request-too-large is observed within carrier-rejection but not its own row; a separate row could reuse actual request bound and afterstate. Lower priority than missing key/media semantics; no count-only promotion.

There is no identified product question for the proposed two rows. Product-dependent Selector/incident-Link issues remain outside this plan.
