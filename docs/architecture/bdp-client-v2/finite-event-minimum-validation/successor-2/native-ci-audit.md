# Native exact-head CI audit — PR154

No blocking discrepancy found in the captured successful CI run. This is an audit of recorded evidence and local Git objects; no tests or product commands were rerun and no source was modified.

## Identity and scope

- Run 38995336477 / verify job 117286523844 completed successfully at 2026-10-17T02:50:57Z; all 27 recorded steps are completed/success.
- PR154 captured head is `f6204255d53ef65d4aea56ff5ccc86d72e8cc7b0`, working-branch base `codex/vickie-parser-refusal`. Local Git independently confirms tree `f5a20fe30562b11887d11a1100dbbd473843de7e` and sole parent `b76c3b7df77ea9b33611bc9027db651833e20abb`.
- Exact parent diff is six files: transaction-http.ts, transaction-service.ts, transaction-store.ts, transaction-events.test.ts, transaction-service.test.ts under packages/server/src, and apps/bdptest/src/transactional-event-missing-context-installed.test.ts. No additional integration or main merge is represented.
- Raw log independently rehashed: 3,582,932 bytes, SHA256 `c59b07c295771471b84d50b34f8f14a5a5fd1b5fbb568a923b4b273578abc9e4`. Captured metadata SHA256 `b023ad0fee4c9f15247a45e93682bc086ced5bcba67f838bbb54a0f600a490f0`.

## Actual logged execution

The raw log independently shows 186 test files passed, 9,375 tests passed and one skipped (9,376 total). The separate ready-preflight run shows seven tests passed. Format, lint, boundaries, typecheck, build, test, preflight, smoke and historical Read evidence steps all succeeded. Counts are execution totals, not functionality or profile qualification percentages.

Every one of the ten new selected callbacks has exactly one passing test line: A owner gate (line3371), B captured tuple (2485), C cancellation/drain (2486), D finite HTTP (3375), E installed minimum (6131), and five F preflight cases (2487–2491). Line references count raw-log split lines, without changing retained bytes.

The nine source observation records agree with the selected workload: owner8 reads/one seed; capture4 reads/two calls/two waits/two real probes/no group reads/times[0,0,5,10]; cancel/drain5 reads/four calls/three waits/three probes/no group reads/times[0,0,5,10,15]/one post-close refusal; five preflight cases each one call, zero owner reads/waits, clock counts1/0/1/1/1; emitted HTTP32 calls/no product children/directory absent/zero failures.

The new installed record (line6129) parses as JSON and declares exact f620, clean workspace, entry hash `bcf153170f953f9e28c6c89f1153be907f64df289e64f3ad929cad4a4a54c47c`, canonical authority, three attempted/exchanged responses, outcome passed. Observed sequence is GET200, GET503, HEAD503. The refusal GET has127 bytes; HEAD has zero bytes with paired content-length127. Both refusal responses carry application/problem+json, private/no-store, retry-after1, matching current epoch/view/pos-0 and non-attesting development metadata. Child24499 reports close attempted and completed, exit0, null signal, no deadline, directory absent, ownership closed and zero failures. claimEligible remains false. This full-suite callback uses the logged workspace dist entry; it must not be relabeled as the separately validated external package consumer.

The preserved historical missing-context callback also has a passing line (6126). Its raw record (6124) visibly declares f620/clean, attempted11, passed, closed, child24463 exit0/null signal/no deadline, directory absent and zero failures. However GitHub secret masking damages JSON string escaping in a WWW-Authenticate field. Therefore this audit does not claim full JSON parsing or independent revalidation of every historical response/header from that masked line. The unaffected metadata fields and actual passing callback are sufficient for this bounded CI audit; the preserved local external report remains separate evidence.

## Smoke and historical evidence

The smoke observation (line11372) reports payloadIdentity `b9d1d50e595213b55bfb4a0da7ea536f3dd625239f7b09c7f6cf301b010243ed`, matching the independently audited local installed payload. allNineTargets is true and claimEligible false. It is a separate smoke observation, not the same identity field as the private transactional callback's served installedPayloadDigest.

The Read gate command permits exit2, but the actual log contains `evidence.verified` (line11382), not just a tolerated command status. It verifies docs/design/evidence/read-cohort/read-v1.json with historical runHead `e5f7d0c43c151426350ceededdc77fe7e372db9b`, evidenceCommit `dbaf63a24219e1ea3f80394746853598b0d4464a`, projection `bd7da0fc3f8082e0bfb6f4969c04409f2a610e12408dbc6a0bd14b4b351127cc`, 47 definitions/46 reachable/89 rows. Those historical identities are not relabeled as new f620 conformance evidence.

## Limits

The captured API response is the basis for remote exact-head status; local Git independently corroborates the commit/tree/parent/scope. Remote process absence cannot be freshly inspected from this machine. Specific callback exit/store cleanup assertions are logged and passed; Actions' generic orphan cleanup line is not a complete remote PID census. No claim of complete Transactional/write profile delivery, public discovery qualification, S1/G2 closure, release or promotion follows from this CI success.
