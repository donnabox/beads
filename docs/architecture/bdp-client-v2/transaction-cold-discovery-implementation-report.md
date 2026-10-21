# C3 internal discovery assembly — frozen author report

Recorded 2026-10-14T09:27:51.350345+00:00. This is bounded author evidence, not independent review or public qualification.

## Source and authority

- Frozen source **b20c92914b48e66b33b66e9e7434dbc3467fad10**, tree **6a84e1cc2916ee5fb4cc49ad7cfd2a65562ef260**; ordinary parent **aebf656506d187ed226302b68b526d34088537c6**.
- Checkout `/Users/dbox/repos/wt/vickie/vickie-tx-discovery-assembly`, branch `vickie-tx-discovery-assembly`; clean after frozen pack/install execution. Commit includes the Agent-Signature trailer.
- Accepted v3 plan `../transaction-cold-discovery-next-plan-v3.md`, SHA `1ff0c0cdd954249bb2ad743614c7aa7e4b4d9dce89298e4954d78d9b34e6a0ae`; root acceptance `../transaction-cold-discovery-v3-root-acceptance.md`.
- Started at frozen Snapshot587a. Root explicitly authorized current main integration: saved WIP, fetched and verified main aebf656, fast-forwarded ordinarily, reapplied WIP without conflicts. Existing alias/Snapshot repairs preserved. No rebase, cherry-pick, or authored resolution.
- Fourteen changed files; per-file byte/SHA bindings in `source.json`:
- `apps/bdptest/src/transactional-discovery-installed.test.ts`
- `docs/design/transaction-implementation-progress.md`
- `packages/protocol/src/index.ts`
- `packages/protocol/src/transaction-discovery-values.test.ts`
- `packages/protocol/src/transaction-discovery-values.ts`
- `packages/server/src/transaction-discovery.test.ts`
- `packages/server/src/transaction-discovery.ts`
- `packages/server/src/transaction-http.test.ts`
- `packages/server/src/transaction-http.ts`
- `packages/server/src/transaction-installed-contracts.test.ts`
- `packages/server/src/transaction-read.test.ts`
- `packages/server/src/transaction-service.test.ts`
- `packages/server/src/transaction-service.ts`
- `scripts/smoke-executables.mjs`

## Implementation and boundaries

Two protocol parsers/types validate the existing canonical Transactional discovery/directory schemas after a detached frozen JSON capture. Canonical Scope, nine navigation URLs and optional multiplicity Type IDs use existing grammar helpers. No schema or normative amendment.

The service captures the five nested advertised page/Selector scalars plus private TTL and existing enforcement options before validation. The Read plane, receipt size check and candidate projection use that capture. `discoveryObserved()` copies epoch/view/head in exactly one real synchronous owner entry. Assembly and both schema parsers run after the gate releases. Owner refusals and protocol validation errors remain distinguishable; no reader escapes.

The private candidate has exactly sixteen required members plus limits (17 keys), fixed current replay floor `pos-0`, and the shared twelve-operation directory. HTTP constructors reject a different Scope. The directory is shared with actual target construction; literal independent tests compare all twelve routes. Private module inventory/deep-import denial remains intact; only the two approved protocol parsers/types become root exports.

**Not wire-ready:** the candidate projects only page/Selector limits. C4 must add all required enforced-limit advertisement, including bounded validation diagnostics/diagnosticBytes, before any activation. Existing public admission stays closed. Scope/descriptor/directory routes still refuse; there is no public cold-client discovery claim, optional History/Types work, new store law or format, C4 activation, or fixture pretending to be discovery.

## Executable evidence

Pinned Node24.16.0 / pnpm11.20.0. `exits.json` preserves every command, exit and elapsed time; verbose/JSON logs remain alongside it.

- `combined-tests`: **213/213**, seven files, actual service/owner, schema parsers, HTTP/emitted contract routes, and emitted child boundary fixture; 2.83s Vitest duration.
- `read-ru-regressions`: **246/246**, four files (Read values, RU values, Read HTTP, shared Read/sequence lifecycle); 0.853s duration.
- Typecheck, clean build, boundaries, format, lint, unchanged Read evidence verification, both executable/development smokes: exit0. Lint reports 506 warnings/30 infos repository-wide; exit0 is not a warning-free claim. Smoke checks installed protocol parsers, private server exports/deep imports and unchanged public admission failure.
- Frozen clean build, protocol/server packs, isolated offline npm install: exit0. Retained tarballs/install under `packs/` and `install/`; SHA bindings in `summary.json`.
- **Fresh packed TX:** one test, 12 actual raw exchanges, one child, canonical configured Scope `https://reference.invalid/acme/` on separate loopback transport. Six authenticated GET/HEAD discovery refusals500; three anonymous GET refusals401 without context/discovery targets; batch200, independent Bead GET200, exact retained retry200. All three success positions pos-1. 0.525s runner duration, 0.214s case. Enforced budgets remain45s cohort/60s outer/5s exchange/10s launcher; ceiling16 exchanges/one child. Actual12/one.
- Fresh packed targeted RU: **2 passing**, 200 deliberately unselected; retained cross-carrier retry and reference/category observations against the same packed payload. No full-profile/full-workspace assertion.
- Independent negative oracles cover stale/foreign owner values and roots, optional History insertion, wrong limit projection, required/malformed schema envelopes, canonical grammar, mutable caller configuration, owner refusal, schema error, Scope mismatch and closed lifecycle. Complete genesis replay survives completed receipt detail expiry and durable reopen; exact retained receipt compactness is separately asserted.

## Exact installed bindings

- Payload `da8ccc6f5c5f36074f4f56e15e787705073335e996aa347e2f02f4363f1b9272`.
- Contracts `184330942591336c6ad84ef3d32e2704f67006b8ff64882a300488ab4698608a`.
- Runtime configuration `7335b088e633ec32982476e99bd60f24ed105e089a26fe2115b1af3610671ff4`.
- Policy `7abde996888c210dd40d671a76f7cba6a6ba560da39a36508ea5d76f91da8bb7`; authentication `5a4ebded36163c7c39f5de35bef70d39faf9f112a0db4bdd2fb6a07eed64d34f`.
- Fixture `3a7899ced771d4880066d98b0000c53f801ece53458502cc4f3bc641146c9a85`; profile transactional, historyfalse, claimEligiblefalse.
- Source is caller-declared launcher metadata independently pinned by `pack.py` to clean HEAD before/after; target payload/config bindings come from actual ready capture. No historical execution is relabeled.

## Preserved preparation failures

1. `typecheck-initial`: two new test ReadProblem union-narrowing errors; corrected with existing isReadProblem helper. No production failure.
2. `core-tests`: 136pass/1fail; fixture wrongly expected two groups in one changefeed page. Corrected to follow actual bounded continuation pages and independently pin positions1/2 and floor0.
3. `typecheck-affected`: exact raw transport requires a default credential handle; supplied explicit anonymous empty-header handle and writer handle, without broadening target routes or executor.
4. `affected2-tests`: 212pass/1fail; test assumed receipt expiry at exact expiresAt. Owner promises through that instant; moved clock to strictly after it.
5. `affected3-tests`: 212pass/1fail; test assumed completed receipt disappears after expiry. Existing owner correctly retains completed compact identity/detailexpired. Corrected exact compact receipt observation; original groups still replay unchanged.

All original logs remain. These are author fixture corrections, not claimed runtime defects or waived failures.

## Cleanup and next step

Packed child PID 57850 exited0, process absence checked in fixture and independently after return. Temp stores removed by fixture. Retained install/tarballs are evidence, not live processes. No active owned test/build/launcher remains after supervisor completion. Saved integration stash is a recovery artifact only; current work is committed and clean. `cleanup.json` records final process check.

Root owns independent council, PR/integration/readiness. This completes internal C3 assembly only. C4 serving/complete required limits and its admission/evidence gate remain separately authorized future work.

Guarded branch push completed and exact remote ref independently verified at b20c92914b48e66b33b66e9e7434dbc3467fad10. See `session-before-push.json` and `remote-verification.json`. No PR, main merge, bus or launch modification performed.
