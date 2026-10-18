# C1 initial collection minimum implementation report

## Frozen source

- Checkout: `/Users/dbox/repos/wt/vickie/vickie-tx-collection-minimum`.
- Branch: `vickie-tx-collection-minimum`.
- Commit: `71c7aa6f9d81ea32965a1c5d5a15fd4b89bb092c`.
- Tree: `5f4f278136b1ac4fb33cfc4711b0cb111949d025`.
- Ordinary parent: verified main581bad0de6f5eb098b920147f6a7e09809c553ea. Working tree remains clean and immutable after freeze. No push/PR/bus/main actions by this worker; root owns publication/council.
- Scope: five files. Production only transaction-read.ts and transaction-http.ts; affected transaction-read.test.ts/transaction-http.test.ts; new apps/bdptest/src/transactional-collection-minimum-installed.test.ts. No store/evaluator/AuthorityReadPlane/runtime/configuration/schema/spec/fixture/catalog/public exports or flags edited.

## Result and limits

Initial current Bead/Link collection GET/HEAD now honor the existing complete canonical epoch/view/minimum tuple. One shared private predicate excludes Types, incident-Link collections, continuations, aliases and discovery. The same owner entry supplies records and context; named materialized request-control400 invalid-parameter /413 limit-exceeded survives instead of becoming a timeout. Valid reads check their minimum before retention or conditional shortcuts. Existing request.signal now reaches the Read facet/service wait, including the previous point path. No collection ETag, new wait policy or new profile claim.

This closes explicit minimum-control refusal, not a previously stale ordinary-read defect. Ordinary same-owner collections were already current. Existing unresolved partial/unknown tuple, retained-page/Snapshot minimum, absent-point visibility, alias/discovery and wider profile obligations remain. Successful initial continuation captures original data while its later no-minimum response reports current context. Synchronous materialization and current bounded sleep are not preempted on abort; the next observed abort stops further attempts. Collection scans remain unbounded by page size; no performance claim is added.

## Verification, accurately separated

Pinned Node24.16.0/pnpm11.20.0; frozen install exit0. Build preceded emitted/server process testing and was repeated cleanly for final package production.

- Pre-freeze focused six-file run:263 pass, covering shared AuthorityReadPlane, Transactional Read/HTTP/service/consistency and preliminary emitted-server installed test. This was a dirty author checkout and is not clean tarball evidence.
- Restored direct Read/HTTP after counterfactuals:80 pass. Eight new direct cases cover controlled behind→real HTTP mutation→new data, zero cursor retention across refused attempts/small capacity, actual pre-aborted server Requests with no owner SQL, and actual Node disconnect synchronized through server AbortController before releasing the wait. Both collection and existing point paths are exercised.
- Typecheck:pass; final frozen clean build:pass; boundaries/format:pass; lint:exit0 with existing warnings/information (not claimed warning-free); both installed executable/development smokes:pass; sealed historical Read evidence verification:pass without relabeling.
- Fresh clean tarball cohort:one test passes,36 actual raw requests, one installed child. Four no-progress cases (two collection kinds × GET/HEAD) use the real1000ms budget. Context errors GET/HEAD, conditionals304/412, named control failures, unauthenticated nondisclosure, literal properties/metadata/types/record revisions and immutable continuation rows with current serving context are observed. No fabricated behind-attempt notification or installed timing seam is used.
- Packed negative:separate disposable installed copy bypasses the actual source-condition call; the original source/tarballs/positive installation are untouched. It fails at the intended behind GET status assertion:304 conditional shortcut instead of503 catch-up-timeout.9 completed raw requests; child closed exit0. This is a deliberately mutated binary derived from71c7, not a source-matching implementation or a conformance result.

Direct semantic counterfactuals (all restored before final build): minimum bypass causes3 intended assertion failures; missing signal forwarding causes4 intended failures; removing named refusal preservation causes1 status failure. The first minimum-bypass TS edit had a trailing-comma transform failure; that direct Read failure is not credited. A valid Array.of bypass rerun produces the intended3 semantic failures. Original logs remain available.

## Exact installed identities and measured budget

- Positive installed payload: `9d5c6b45fe54d863b89c47d4fe1d779bf1a2a8f57a103ffd6845c49b7151c637`.
- Protocol tarball SHA256: `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`.
- Server tarball SHA256: `c6ea9254b268f25099d0e2e96100e6ffeabb900ede4b997c032c118cb2d3b369`.
- Counterfactual payload: `93b0ad19c3fc9ccf24e7ad8afe8e4f705e683299eb22865caf71fdda1d402e7b` (different by design); original/modified module digests are in packed-71c7aa6f/counterfactual.json.
- Positive request count36; negative9; actual combined45/two children, below frozen80/three ceiling. Planned counterfactual prefix11 was conservative; actual first failing prefix is9. No extra Resources:3Beads+2Links per installed session.
- Maximum positive response body `2342` bytes under32KiB. Request headers16KiB/body64KiB, wire5s, lifecycle10s, aggregate90s and outer120s unchanged.
- Positive timeout elapsed milliseconds: 1005.02, 1003.4, 1002.46, 1002.98. These measure the finite real wait; no stronger SLA is claimed.
- Positive child `95140` and negative child `95165` each reported closed=true/exit0; ps confirms neither remains. All worker-owned test/executor commands completed; no interrupted/timed-out run, no killed process needed.

## Failures retained and corrected without weakening an oracle

1. Initial direct fixture mistakes used wrong mutation members (id/change instead of bead|link/propertiesChange), incompatible service page maximum and too-small capacity relative to its position bound. Corrected tests only; initial5 failures preserved in direct-first/direct-debug logs. Restored production then passes80.
2. First installed attempt encountered the exact serializer's existing conservative secret-name policy on bdp-authorization-view. Test now uses its existing declared credential-resolver field for that observed view; no firewall or API changed. Second trial exposed a remaining wrong authored update member and was corrected; neither failed trial is counted as positive evidence.
3. First minimum counterfactual's TS trailing comma and a subsequent external mutation-script pattern mismatch are preserved, with all files restored in finally. Only intended semantic failures are credited.
4. Final package automation initially checked for the wrong negative text (200 versus503), while the actual expected failure was304 versus503 because the test deliberately supplied If-None-Match:*. Original pack-progress AssertionError and executor remain. Separate finish-validation.py validates that actual failure and performs remaining checks; no rerun or weakening of the positive/test/source was needed.

## Evidence locations and handoff

All logs: `L/transaction-collection-minimum/`, where L=/Users/dbox/repos/wt/vickie/bdp-integration-logs.

- `frozen-inventory.md`: bounded pre-execution inventory, plus explicit measured prefix correction.
- `focused-green.log`, `restored-green.log`, `typecheck-final.log`: author validation.
- `no-minimum-valid-red.log`, `no-signal-red.log`, `no-refusal-red.log`: discriminating direct failures; earlier inconclusive attempts separately preserved.
- `packed-71c7aa6f/`: clean positive package, deliberately distinct counterfactual copy, exact test output, exit/timing records, static/smoke/Readseal results and summary.
- `safe-summary.json`: source/tree/tarball/payload identities, actual contexts/configuration associations, response hashes/timings/counts and cleanup; no tokens/raw authorization headers.

The clean frozen source is ready for independent code council/root integration. Public admission and every whole-profile qualification flag remain unchanged/false; no CI/council completion is inferred from author evidence. Parent can ordinary-integrate the frozen HTTP seam into the alias lane; this worker makes no further source edits.
