# Native code review — direct Transactional refusal context

## Findings

**P3 — missing-header matcher loses the intended negative-control label.** `apps/bdptest/src/transactional-direct-refusal-installed.test.ts:223–226` passes `response.headers[name]?.length` to `toBeGreaterThan(0)`. For an absent header Vitest rejects undefined with a generic TypeError before rendering the supplied diagnostic. The actual predecessor run therefore fails the right source predicate, but `transaction-direct-refusal-context/packed.py` subsequently fails its required custom-label check and does not produce its normal negative summary. Impact is diagnostic/control classification, not a production refusal bug or false positive. A bounded future correction can use a boolean/nonempty presence assertion that accepts missing input and preserves the label; retain the exact tuple comparison separately. The report must preserve the failed postprocessor rather than describe it as passed. The author's final report does that.

**No Medium-or-higher finding or additional production defect identified.** This is the seven-file implementation review, not approval of broad response-context completeness or public profile qualification. The actual installed negative can be accepted as independently attributable sensitivity evidence for this exact frozen source without pretending its label collector passed; the evidence and limitations are detailed below. No replacement installed run was performed or requested by this review.

## Exact source and reviewed contract

Checkout `/Users/dbox/repos/wt/vickie/vickie-direct-refusal-context` is clean at `6ce9cb4e30d4afa7ac26552490619358fb7044fe`, tree `4984b474155560429a318bd5459593a6c53ad6bd`, direct parent `8c78cba5d2be5d5612c6a9324ee36ed690fdeb19`. Verified at entry and completion. Full diff: seven files,880 insertions/15 deletions. Production changes are limited to transaction-store.ts, transaction-submission.ts and transaction-http.ts; other changes append to their three existing tests and add one installed test. Independently recomputed all seven source hashes against source.json and proved the complete old bytes of each existing test file remain its prefix.

Reviewed accepted plan v2 SHA256 `16917c36972aefa6601f8af602b15da2c2ce24b1ec72e01022fe7ffda37714f3`, dispositions `6fc8bfb3e2324ca2a3ea09e8ffe1123ed9667f399a65ddbc372afaccc76e8d7e`, root plan acceptance, preflight and measured execution release. Source/spec inspection confirms the general Scope response fields at `docs/specs/bdp.md:5766–5774` and ordered admission at4490–4530. No spec, schema, catalog, configuration, launcher or public-admission edits are present.

## Production review

- `transaction-store.ts:1384–1410`: the selected refusal tuple is captured by existing readContext inside the same atomic callback that made the admission decision. The fresh conflict/operation-limit/not-acceptable kinds and retained duplicate406 are all observed-gated. Position is read from actual metadata; the tuple is frozen. No follow-up owner entry, receipt-effectPosition substitution or observation on ordinary unobserved/owned admissions is introduced. Existing operation-limit-without-controls development behavior remains before the new branch.
- `transaction-submission.ts:120–153`: required context is checked only for the three observed refusal kinds; translations preserve it. Comparison returns the final decision intact, so a dependency wait and intervening commit cannot reuse a prior attempt's context. Legacy unobserved shapes and comparison-timeout503 stay separate. Optional type members do not force a context onto unrelated refusals.
- `transaction-http.ts:347–379`: conflict retains409/idempotency-conflict/retry-never, operation-limit requires both code and limit=transaction.operations, and both direct406 responses remain bodyless. Existing context() adds only the three fields. Earlier request-too-large413, authentication, malformed input, minimum development gaps, sequence paths and comparison503 are not accidentally wrapped by status or a blanket fallback. Existing Vary/cache behavior is preserved. This bounded selection does not waive still-unimplemented response families.

The full tests include ordinary and provable unresolved-template conflicts, pending/retained406, frozen returned tuples after a later commit, unchanged owner inspection and retained receipt, fresh-key reuse, capacity-before-Accept, duplicate bypass of fresh capacity, and final comparison-entry context after unrelated work. The SQLite spy inspects the position binding and isTransaction at the owner test, avoiding the legitimate external comparisonEpoch read.

## Independent examination of recorded validation

No tests/builds/installations were run by this reviewer. I inspected actual author exit JSON, Vitest JSON, logs, derivative patches, source and package records; these are author executions independently checked, not new native executions.

- Focused JSON reports362 passed/0 failed across the seven named files: store, submission, HTTP, dependencies, sequence-alias, set and owned. Actual command exit0. Old exact-shape assertions were not weakened.
- First typecheck exited1 for the new spy's TypeScript overload comparison. The retained two-line typing correction assigns the argument to unknown; typecheck2 exited0. Build, format, lint, boundaries and Read evidence verification exited0. The failed attempt remains recorded; lint success does not mean warning-free.
- Four deliberate source derivatives built, then exited1 at intended selected assertions: missing HTTP context → required epoch presence; stale retained position → pos-1 versus pos-2; capture outside atomic → isTransaction false; unobserved leak → original406 and operation-limit exact-shape failures. I inspected each source.patch and actual result.json. These source derivatives are not installed cohorts.
- Fresh positive package run used clean6ce9 source and actual installed payload digest `ced4ef93640ae2dcd67e88a5c7684bab802e9aa73eba974fab81d30ff3e56d19`. Vitest reports one passed test/exit0; its ledger records14 attempted and14 completed requests, one child exit0 and directoryAbsent=true. All five selected direct refusals carry the same epoch/view and currentpos-2 after independent seed commits atpos-1/2. The retained receipt GET remains body-equal to seed while served atpos-2. Collection and Events GETs prove the two known creations and no refusal effects; successful reuse of both previously refused keys producespos-3/4 and four final creations. Refusal Problems use canonical schema checks;406 checks body text, actual zero octets and absent Content-Type.

The installed test validates authored IDs, Type, literal properties and empty metadata independently, then compares full collection Resource records and Event opaque revision/transaction relationships to actual validated receipts. It does not invent opaque revisions or derive expected literal state from the returned collection. The four static routes plus one validated observed receipt URL are exact per-exchange transport routes. The test keeps non-attesting response/ready markers and claimEligible=false. Bound requests65536/header16384/response131072 bytes and5s/90s/120s deadlines match the release. Actual largest request4616 bytes, total positive request bodies5518; observed largest response2182. No extra request/session was added.

## Original payload negative: actual attribution, not arbitrary failure

The separate target package is original8c, payload `a4f4d446ff0e506a329fc19d8f0dbac9a81e325e80750a30936a31bb9e57f6b6`; the harness is clean6ce9. Its two seed responses pass the schema, literal Resource, receipt position and context checks before request3. Its ledger then records singleton-conflict409 with context={}, after which the Vitest stack points to the exact presence matcher at226, called from the singleton-conflict check at318. The thrown message is `TypeError: actual value must be number or bigint, received "undefined"`. It is the missing epoch value reaching the intended matcher, not an arbitrary timeout/setup/parse exception. The subsequent equality and Problem-body checks are not claimed executed on that negative response.

The recorded child6047 exits0 without signal; directoryAbsent=true and attempted/completed count is3. Positive child5945 also exits0 without signal; author cleanup.json records both PIDs absent. Thus positive14/1 plus negative3/1 =17 requests/two children, with no replacement run. Both logs explicitly retain completed-response fingerprints and context only; raw body/wire bytes are inspected in memory but not retained as complete transcripts.

The postprocessor's explicit label assertion still failed after those cleanup checks; normal `packed-negative/summary.json` is absent. Author negative-adjudication.json and final report accurately distinguish that collection failure from the actual missing-header sensitivity. Do not relabel the command as green or accept any unrelated failure merely because exit1 was expected. The exact stack, preceding successful checks, observed409/context{}, and cleanup are the sufficient bounded grounds for this native attribution. Future stable-label repair is Low as above; root owns its disposition and any validation authorization.

The entry module hash is identical between packages because the changed production dependencies are other modules; the distinct full payload digests and source declarations are retained. SourceCommit is launcher/reviewer provenance, not independent target attestation. Random credential/policy identities differ per child as recorded; no credential bytes are retained.

## Evidence fingerprints and limits

Relative to `transaction-direct-refusal-context/`:

- source.json: `bae1fd6057c4349107ab7769ecebb6bd39e299a2a3991fa3dbbd91e4bd016b60`
- summary.json: `a18b9683ee4cd4afaff9ce04e77df2c035c8b1b771fa038cf0c9f0ed2111c4fc`
- positive installed.json: `2a5095a2e2b969602e4728e3a12a1e5e1767bd10e33e96e74fa0e0dff69bfb95`
- positive installed.log: `caa750ac3986bfe823148f44482bbde1918b0e572f3b3420d404efe98afbf1cd`
- negative installed.json: `66e7ebffa6dd6ec311402c3e1770bec908627e9cc97c77c372bf877540df2e60`
- negative installed.log: `bc119469c88cedcbb9f22f54315e43a9fd5f3c4cbfbbdfbf8a049c96ba37734d`
- negative source.json: `9bf3ffedd0c42b47d479035d4bb5b3f487cca5fb2efdb534eacdf7b30a6459eb`
- cleanup.json: `9b61316bea20925f04f1ae919fba9fd82853eacca2a8532d77e70f14467acec0`

The author's final report was read when available before this review completed. Review commands were read-only Git diff/status/rev-parse, source reads and Python JSON/hash/prefix comparisons. One initial abbreviated8c78 diff lookup was ambiguous and failed; it was repeated with the unique8c78cba5 parent and all seven diffs were read. No source edits, new tests, builds, installs, network operations, child services or outward changes were performed. Only this report is written. Full CI, independent external council and root publication gates remain separate; this is no whole-profile qualification claim.
