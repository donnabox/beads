# S2a private valid set HTTP: frozen implementation and evidence

Frozen source: **2f7a1641c613680c5b4dfe056b933bfa6af869dd**, tree `8f62d9c3ed05b6b6d11a7cbb8ccace85e08969f6`, direct ordinary parent `96b00e0be63994b79025fc84ba82a36a55da0bb7`. Checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-set-http`; branch `codex/vickie-transaction-set-http`; origin remains GitHub. Source is clean and immutable for native/Claude review. No push, PR, bus, launch, main merge, public profile activation or conformance credit was performed. Root owns integration/publication.

Authority: root acceptance of supporting plan SHA `798d21ebabc7377e97adf6f242b25d652b67638be2cd86704a632aef2d57aedc`; sole plan remains the fork design document's Autonomous profile completion. Remote launch session was checked before work and matched `2026-10-08T20:26:35Z-81e5`.

## Result and exact scope

Valid supported `updateWhere` and `deleteWhere` now reach the existing private atomic owner through singleton POST targets and mixed batches. They share the existing Resource evaluator, key namespace, receipt paging, intermediate Events and final group persistence. There is no second evaluator or owner, no storage format change, and no new normative rule for invalid Selectors.

Eleven files changed:

- `packages/server/src/transaction-set.ts`: separate closed shape validation from candidate-independent supported Selector validation, so every member's unambiguous shape can be checked before a private gap.
- `transaction-batch.ts`: one mixed preparation path, private singleton union, wrapper-equivalent depth, original-carrier numeric/cardinality/Property Change problem pointers. Internal singleton wrapping does not count synthetic bytes against the actual request cap.
- `transaction-submission.ts`: complete known syntax/key checks, principal refusal, then unsupported composition gap before any key access; normal authoritative compare/admit remains unchanged.
- `transaction-http.ts`: two canonical target entries and typed dispatch through the existing route/method table.
- `transaction-set.test.ts`, `transaction-submission.test.ts`, `transaction-http.test.ts`: meaningful core/admission/actual HTTP regressions described below.
- `apps/bdptest/test-support/transactional-execution-binding.ts` and `apps/bdptest/src/transactional-binding.test.ts`: exact twelve-target inventory and corresponding ready fixture. Exact equality remains enforced.
- New `apps/bdptest/src/transactional-set-installed.test.ts`: actual raw installed observations, receipt/Event pagination, owned effects, rollback, reopen and four failing counterfactual expectation controls.
- `docs/design/transaction-implementation-progress.md`: supported subset and remaining limitations, preserving historical evidence.

No shared evaluator/store/Selector implementation, conformance mutation-source authority, public export, schema/catalog, Snapshot/SSE implementation or other lane was edited. The old unused prepareSetCore helper was removed instead of retaining a competing unused entry point; all preparation uses the same private machinery.

## Behaviors proved

1. Actual canonical set target GET/HEAD/DELETE refuses with bodyless405/Allow POST, including unauthenticated requests. Unknown/noncanonical/query-bearing operation paths retain existing routing behavior. Cold discovery remains a development gap.
2. Unsupported Selector singleton and mixed batches receive401/403 before the private gap when principal authorization fails; no development-limit header leaks through that branch. Later malformed members and invalid key grammar still produce400 before principal. Actual receipt table checks, absent Resources and corrected same-key execution prove no partial admission. Supported callers still receive the explicit bodyless development500 for unresolved Selector categories.
3. Cardinality shape400, retained numeric422, and retained cardinality409 are distinct.9007199254740993 is the numeric refusal;1e30 and1e+30 are admissible/equal numeric identities. Large min on a small population yields409, large max permits selection, and min5/max2 yields409. Cardinality remains before cumulative matched-cap charging.
4. Every outer set counts before negotiation. Actual owner tests prove fresh capacity refusal captures no key; exact pending and completed sets bypass fresh operation capacity. HTTP conflict wins over unacceptable Accept; an equal retained replay still negotiates its response.
5. Numeric/cardinality failures carry original singleton or `/operations/{index}` locations. Cross-carrier replay/reopen preserves those original receipt facts. Property Change failures use the unique original change/path only when existing diagnostic evidence supports it; repeated target paths and afterstate validation point to `/change`. Final Resource diagnostic instanceLocation retains its own meaning. No inferred patch index or shared evaluator change.
6. Equivalent depth32 operation content succeeds through singleton and one-op batch; one level deeper remains a private refusal. A newly added exact64KiB singleton test failed before the internal-wrapper byte correction and passes afterward. The actual larger batch body still receives413. Logs `actual-bytes-before-fix.log` and `actual-bytes-after-fix.log` preserve the discriminating red/green result.
7. Later sets observe preceding staged creations and set updates; late patch/schema/incident-Link/guard failures roll back all earlier effects. Existing owned visibility, source revision, cardinality/cap, numeric identity, no-op, fault/recovery and label-index regressions remain green.
8. A bare-URI endpoint Selector matches only the bare Link, not a Pinned Reference object. The actual HTTP cascade test pins cardinality1, tentatively deletes that bare Link, fails guarded Bead deletion because the pinned Link remains, and proves both Links/Bead plus Scope position unchanged. No implicit pin stripping was added.
9. An actual client is destroyed only after SQLite commits a pending set receipt (existing DatabaseSync prototype observation seam). The set completes and its exact original disposition is replayed after listener close/reopen. This is observed client cancellation, not a real power-loss claim. Existing fault and killed-process recovery tests also pass.
10. Packed installed service uses compiled retained contracts and canonical Host with physical port0. Thirty-two unowned Link deletions produce33 receipt entries over pages32+1,96 Scope Events over pages32+32+32, and exactly one complete change group containing32 tombstones. The test follows actual receipt/Event next URLs and excludes seed Events using the actual last seed Event ID. The Event cap is not treated as a group-page cap.
11. Installed owned Link updates preserve sourceRevision agreement with actual source Bead Reads; subsequent no-op retains it and omits effectPosition. A compiled-schema failure rolls back an earlier tentative Bead and both owned Link/source changes. Retry after changed population does not reselect; original completed and failed receipts/pages survive reopen with create:false.
12. Independent installed expectations reject an omitted receipt result, omitted Event, reordered Event block, and wrong owned sourceRevision. Successful responses also require all three actual current-context headers. Receipt roots/pages, Event pages and change groups are checked against the normative schema bundle.

## Validation, commands and limitations of evidence

Pinned Node24.16.0/pnpm11.20.0. Build precedes emitted-module/process tests; Vitest maxWorkers2. Logs remain outside the checkout under this report directory.

- **PASS614 tests /25 files on frozen2f7a164**: `vitest run packages/server/src/transaction-*.test.ts apps/bdptest/src/transactional-binding.test.ts apps/bdptest/src/transactional-set-installed.test.ts --maxWorkers=2 --reporter=verbose`. Log `frozen-2f7a1641-regressions.log`, exit0,14.93s. Includes Event/changefeed/Snapshot/Read, atomic fault/recovery, sequence, service, receipt and materialization regressions. This is a focused subsystem run, not the full repository8k suite.
- **PASS** clean build, typecheck, dependency boundaries, formatting, lint, installed package smoke, development smoke and unchanged Read evidence verifier. `check-exits.json` contains all eight exit0 results, with individual `final-*.log` files. Formatting reports one existing oversized-file warning; lint reports240 warnings/eight infos and no error. These are not warning-free claims. Fresh frozen packing rebuilt exact source successfully again.
- **PASS1 new packed set test**: `packed-2f7a1641/set.log`; complete clean-source observation extracted to `set-observation.json` with raw actual wire responses, canonical target URLs, body assertions, binding and counts.557ms test duration; exact log is authority.
- **PASS35 existing TX app tests /5 files** with the freshly installed entry, including actual Event SSE/changefeed SSE/Snapshot and retained-receipt harness observations plus local binding tests: `packed-2f7a1641/tx-regressions.log`,32.88s. Not every test in these files is itself a distinct wire witness.
- **PASS173 RU tests /4 files** against the fresh installed reference entry: read-update matrix, exact-client, development-server and installed-read-oracles; `packed-2f7a1641/ru-matrix.log`,96.53s. Read seal was verified without regeneration or relabeling.

`packed-2f7a1641/exits.json` records build, both packs, install and all three cohorts at0. `pack-executor.py` pins exact clean HEAD before packing. Earlier `installed-*.log` WIP observations explicitly record workspaceDirty=true and are not the frozen tarball evidence. Initial authoring failures (test fixture assumptions, one unused import after cleanup, and the deliberately red actual-byte regression) are retained separately; none is presented as a passing frozen run.

Fresh tarballs:

- protocol SHA256 `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`
- server SHA256 `c89ff8bc08a262a4609ef126e169ee7bda9ff5ae7901e33c57a3683a97b30428`

New set observation binding:

- source `2f7a1641c613680c5b4dfe056b933bfa6af869dd`, workspaceDirtyfalse
- installed payload `312ffe5038aefa399096870568a26013fc9c2dfa55168fff700af2a2c9ff5ce6`
- captured contract closure `95b4ef854865acafc04caa438285dd52b7955fa99e6b241397660d6f262b4935`
- runtime configuration `1df00e341a34f4f3e57796c06e7c6e0ee93bc317c323bad7f2f1945087306ef1`
- claimEligiblefalse; generated CSPRNG credentials are not included in logs. Exact policy/authentication identities and initial/reopened metadata are in the observation JSON.

## Remaining boundaries and handoff

Q1/Q2 negative Selector syntax/complexity admission semantics remain unresolved and explicitly private. Candidate enumeration still materializes the current staged collection per set;32 matched Resources is not a scan/CPU bound. No frozen transaction-wide candidate cache is introduced. Installed policy remains the existing fixed supported policy; component visibility tests do not qualify mutable grants. New request-pointer work is bounded to set failures with known location; it is not a global audit/repair of every ordinary Resource Problem.

The ordinary conformance mutation receipt-source allowlist was deliberately untouched; no set ordinary-runner/catalog credit follows from raw installed proof. Public discovery, public write-profile admission, known metadata-only Event projection gaps, erasure/publication fences and remaining consistency obligations stay unqualified. No Type/History/provenance redesign.

All owned subprocess runs completed with explicit success exits; final process inspection found no remaining test/installed children, only the inspection command itself. Disposable fixture databases were removed by test cleanup. No run was killed or timed out. Worktree is clean, frozen, and ready for the root's independent native+Claude code council. Final integration must preserve ordinary ancestry with any concurrently merged app-only PR93 work; this lane made no outward changes.
