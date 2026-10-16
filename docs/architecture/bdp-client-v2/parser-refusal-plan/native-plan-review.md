# Native parser-refusal plan review

Verdict: one concrete blocking fixture/oracle mismatch in C; the production reordering and bounded refusal scope are otherwise sound. Read-only inspection of immutable e34 snapshots; no tests, imports, services, source/worktree edits or main access.

## N1 — blocker: valid incident-Link control is not pre-owner500

C creates an empty store but expects beads/a?view=links GET/HEAD to return private500 without owner entry. requestFromUrl produces an initial bead-links request. currentReadMinimumKind returns missing (transaction-read.ts:99), so it passes the strict unknown guard in transaction-http.ts:584–588 and reaches the authority. checkSourceCondition permits resource-not-found for non-strict kind (transaction-read.ts:211–212); the missing anchor therefore returns contextual404 after owner observation. This matches spec5833–5839. With an existing anchor this path still enters the owner before refusing unsupported unknown-minimum eligibility, so merely seeding the anchor cannot fix the no-owner oracle.

Smallest recommended correction: replace C's valid beads/a?view=links control with valid types/task GET/HEAD. Type Resource requests classify strict regardless of existence and the unchanged early guard returns bodyless500/minimum-position without owner entry. Together with types/ this preserves four valid excluded-route controls, all16 no-owner calls and the67 total. Keep the existing missing/hidden incident-Link coverage separately selected; do not invent new incident success/error policy or add seed writes for this refusal slice. Update plan wording precisely before author release.

## N2 — required validation ordering fact, correctly anticipated by this plan

transaction-http.test.ts:18–20 imports ../dist/transaction-runtime.js through an explicit file URL at module evaluation. Workspace aliases only map package specifiers; they do not replace this import. A handler-only callback can use source modules, but collecting the suite still loads emitted runtime; the existing collection-unknown callback actually uses fixture/startTransactionRuntime. Therefore require a fresh bound build before collecting the HTTP suite, with current source/dependency provenance. The plan's emission/dependency warning is correct. The earlier collection validation preparation's source-only-before-build explanation was incomplete and has been reported to root; do not run that predecessor ordering as current candidate evidence. No new production behavior is requested by this finding.

## Scope and precedence verified

The proposed move returns only an already-created parser400 before semantic unknown-minimum support classification. It does not parse twice, widen successful routes, compare foreign epoch/view, visit retained state, allocate continuations or change S1/G2 semantics. Existing authentication and complete lexical tuple validation remain earlier (transaction-http.ts:403–405); earlier Event/receipt/alias/snapshot/changefeed routes remain intact. The ordinary Read router also returns parser problems before perform (read-http-routing.ts:39–40), supporting absent/current/unknown parity and no-owner expectations on A/B's invalid requests.

Query-bearing continuation-shaped requests can return parser400 before any retained-token lookup: requestFromUrl validates allowed/repeated parameters and limits before constructing continuation requests (read-request.ts:145–181). Returning that refusal does not implement a valid continuation minimum. The spec's unsupported/repeated collection parameter400 rule explicitly includes cursor among collection parameters (spec1539–1555). Valid continuation handling remains strict/excluded. Missing canonical resources with invalid query likewise fail ordinary query validation first, as spec5838 allows; valid missing resource behavior remains404. No foreign-context collision is being chosen: proposed fixtures use actual current context.

Outer applyReadHttpSemantics negotiates/conditions only successful200 JSON or204; existing400/401/500 refusals remain errors, with HEAD suppression and metadata handling (read-http.ts:10–44). Invalid Accept and preconditions cannot turn these parser400s into406/304/412. Expected HEAD metadata in direct handler checks is implementation evidence; installed plan correctly does not require optional HEAD media. Parser404 remains in the old support guard; this plan does not claim to repair it normatively.

## Workload and coverage

A36+setup1, B12+setup1, C16+setup1 =67 handler invocations over3 stores and no product child. The recommended C substitution preserves those counts. Installed25 =setup1+opaque12+absent/current4+noncollection6+precedence2; one child/store, no new waits. This is a justified cross-section of parser branches plus auth/lexical/conditional refusal ordering, not a new router redesign. Existing regression selections have additional workload and must be inventoried from authored bytes before execution;67 does not include them.

Optional small strengthening: substitute one already represented B query case with beads/?cursor=one&cursor=two GET/HEAD to directly show parser rejection of a continuation-shaped target without visiting retained state. This can preserve counts and does not need a real token. Structural ordering and existing repeated-parameter tests already support the behavior, so this is not a second blocker and must not become valid retained minimum work.

Authoring remains contingent on the reviewed/integrated collection successor and a fresh binding of the actual HTTP block. Update its two parser500 sentinels only; preserve valid503, materialized400/413, hidden404 and excluded-route assertions. Exact source scope3 files, no spec/classifier/grammar/launcher/public activation changes.

Reviewed plan SHA256 `9a3251349989ceaeb0be9190ea729b1ae82f00679e7ff0f660140b6abec5ca45`. Snapshot commit `e34a06d9033b26488be4a03ff124b4c8665f2c68`. Every snapshot hash below was verified before and after review preparation:

- `docs/specs/bdp.md` `67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8`
- `packages/server/src/transaction-http.ts` `81c130dd18f604ab1d13c03a2f52a98e11a6988f45d62af98b4b8c7499d14529`
- `packages/server/src/read-request.ts` `57675afc647df3e8ad1f26a40aef72584aa388ade6af4a465e3ffef16670f6a2`
- `packages/server/src/read-http.ts` `31934105eff6fff8ded8ce2ca9e3c19028991ca7c3d5f51b848004177e8c9268`
- `packages/server/src/read-http-routing.ts` `d9344b87838f8c71381ccf9efe28b4f023df45503570beb8fb4293d553caca42`
- `packages/server/src/transaction-read.ts` `7e2cc29b668bb80a3ba2948e90511518a98a7ff798b1dbfb8c6f51f0c0bd2a6f`
- `packages/server/src/transaction-consistency.ts` `ad51311d6098db32fd47849e0b78d3118162cdf1e80ee9ef4d46324bfa9f436f`
- `packages/server/src/transaction-service.ts` `b6f9584d67992ddd3a08304a6b76563318252c316b09afd991e49a3304d3af05`
- `packages/server/src/transaction-http.test.ts` `9eb452b8e9d80b53f28cd61b40385886c3a61a992aadb1c9c201d9a188081cce`
- `packages/server/src/read-request.test.ts` `1b4450b2e37f2bf716f1b2462ccfa76a88458fa4bbabe1b0435ae7e0d61ccdfa`
- `apps/bdptest/src/transactional-collection-minimum-installed.test.ts` `615caab0e5a846c546e718f65c4041c5ca41bef59df1d56a6dea0e42638b0174`
- `apps/bdptest/test-support/transactional-execution-binding.ts` `cde506c72a8710eafe58522900f6db70ed3701cf333560252dc8e7c5e706cf89`
- `scripts/transaction-installed-observation.mjs` `6f772f3776d3c3effa788676d853cc2932a6ee59ad63ac3338ad1b487408914b`
