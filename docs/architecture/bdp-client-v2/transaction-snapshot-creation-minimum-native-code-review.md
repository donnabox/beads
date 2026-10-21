# Fresh Snapshot creation minimum — native code review

## Findings

**No actionable findings in the reviewed nine-file delta.** The implementation follows the accepted fresh-creation-only plan, preserves the retained/malformed private exclusions and keeps public discovery/profile admission closed. This is bounded code/evidence review, not full Snapshot or Transactional qualification.

## Exact scope

- Frozen source: `587a7eeaa9663ab3d516a45375b3dbf0e9fedca9`, tree `492fe6c84793a69a695d4847f3f9986d32cdeaa9`.
- Direct base: `82a0a0261ba042be3721a16bdccc37915a0e5fa1`, the attribution-only alias HTTP successor specified by root. Frozen checkout `/Users/dbox/repos/wt/vickie/vickie-tx-snapshot-minimum` remained clean at the same HEAD throughout review.
- Full nine-file delta inspected: three production files (`transaction-store.ts`, `transaction-service.ts`, `transaction-http.ts`), four server test files, installed Snapshot test and progress document. No unrelated runtime, schema, format, admission, or evaluator change.
- Authority: accepted Snapshot plan v2 `2fb0a12ea6061f40033f51a5f7c52a104248e3b65eef296d7bfbd090b42fa9d0`, its root acceptance, actual normative strict-read/minimum/conditional clauses (`docs/specs/bdp.md:5758–5811`) and existing HTTP semantics. No new universal minimum-versus-media ordering is inferred.

## Production and test assessment

| Area | Observed implementation and evidence |
|---|---|
| Same-entry eligibility, no allocation | `transaction-store.ts:1428–1436` snapshots the optional minimum, rejects retained-kind misuse, obtains current context and checks eligibility before expiry/preflight/copy/allocation. The minimum error is outside the SnapshotFailure-only catch and cannot become a Snapshot success/problem of another category. Existing atomic copy/rollback/fencing remains unchanged. Owner tests repeat failed attempts, inspect all three Snapshot tables and show expired details remain untouched by ineligible requests. |
| Immutable controls and finite wait | `transaction-service.ts:252–275` freezes the minimum and create/preflight kind synchronously before run/retry. Both paths use the existing single-deadline wait helper with the actual signal; only behind is retried. No SQLite entry spans a wait. Service tests mutate both caller tuple and request during a controlled wait, commit a real mutation independently, and assert the resulting actual manifest/context. They also show close drains pending delivery, pre-abort makes zero owner attempts, timeout preserves actual context and retained misuse is rejected. |
| HTTP parsing and boundary | `transaction-http.ts:378–437` retains authentication and closed query parsing, accepts canonical complete controls only for fresh creation, keeps partial/malformed/retained controls explicitly unsupported, and forwards the minimum/signal into both capture and preflight. The fresh foreign epoch/view outcomes use current context rather than echoing supplied context. Outer error mapping excludes raw behind; service wait supplies the normative catch-up-timeout error. |
| Terminal preflight | The existing pure JSON/no-validator candidate only selects the path. Authoritative owner preflight is still required for HEAD/non200 candidates; the branch returns immediately afterward, preventing a second wait/capture. Tests pin one service call for HEAD/304/412/406 at eligible and future bounds. Existing owner-before406 ordering is preserved; it is not asserted as universal protocol law. |
| Cancellation and lifecycle | Actual Node disconnect tests now include Snapshot GET and HEAD, synchronize on an armed wait and real server abort, release the controlled wait and await runtime close. At the fixed head a missed cancellation would cause another wait; exactly one is observed. This supports stopping subsequent attempts, not instantaneous interruption of synchronous copying or a pending bounded sleep. Existing finite-service drain owns the promise. |
| Retained/no-minimum behavior | The installed small cohort compares prior retained bytes before/after live mutation and reopen. The larger existing fixture sends a complete tuple to an actual issued continuation URL and observes the retained minimum gap, then completes its unchanged no-minimum page/convergence checks. No retained-anchor interpretation is implemented. |
| Independent state/context oracles | Actual authored Bead/Link literals, fixed private positions, exact epoch/view headers, schema-valid manifests/timeout Problems, foreign-context refusals, later fresh capture and distinct handle, and full retained body equality distinguish meaningful neighboring outcomes. HEAD/304/412 have exact context and no body checks. Direct table observations complement HTTP capacity controls; an incidental503 cannot satisfy the catch-up-timeout oracle. |
| Transport/provenance | The app adapter routes the view header through declared credential resolution and verifies the actual exact-request header. It retains canonical Scope/Host, closed routing, response-octet/wire-digest observations and actual installed binding across reopen. Completed-response records are explicitly completed responses, not a claim that failed attempts are serialized. New child-exit recording adds evidence without changing launcher/runtime behavior. |

## Independently executed checks

Created an exact `git archive` of the frozen commit under `transaction-snapshot-creation-minimum-native/source`; dependency links point to existing installed dependencies, while root Vitest caches stay in the disposable archive. No install/build or author-output mutation.

Command, working directory that archive:

```text
/Users/dbox/.nvm/versions/node/v24.16.0/bin/node node_modules/vitest/vitest.mjs run packages/server/src/transaction-service.test.ts packages/server/src/transaction-snapshot.test.ts -t 'fresh Snapshot minimum service|checks each fresh Snapshot minimum|does not expire retained Snapshot details' --maxWorkers=1 --reporter=verbose --reporter=json --outputFile.json=/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-snapshot-creation-minimum-native/focused.json
```

**Exit0:7 passed,0 failed,79 unselected/skipped across2 files.** Logs: `transaction-snapshot-creation-minimum-native/focused.log` and `focused.json`. These are targeted source tests, not independently repeated packed or full-suite results. The initial combined setup/test tool invocation failed before process creation because its intended archive working directory did not yet exist; setup was then run from an existing directory and the recorded test completed normally. No failed test is hidden by that setup correction.

## Fresh author evidence inspected, not independently rerun

During review, `transaction-snapshot-creation-minimum-implementation/packed-summary.json` became available, SHA256 `f3237923d2b8ff57db8adec2e85d03c7e9a1a9623420ab18b984247e874a78ed`. It binds the exact clean positive source/tree above. Checked its commands/exits, machine-readable test reports, actual cohort ledgers and derivative patch; independently recomputed both positive package hashes:

- protocol tarball: `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`.
- server tarball: `e63f7509f403140c009ad26b31ac72e15e720d2dfc3e851793b81f8166c6b2a0`.
- actual positive installed payload: `c9be37147a7afe3f047a9a647effaae1abc0c49db80343e1651dd02f0ecee7b0`.
- retained installed entry: `transaction-snapshot-creation-minimum-implementation/install/node_modules/@bdp/server/dist/transaction-installed-process.js`; entry SHA256 `bcf153170f953f9e28c6c89f1153be907f64df289e64f3ad929cad4a4a54c47c`.

The author's final source regression reports190 tests passed across5 files (exit0,6.672s). Its actual packed positive run reports3 selected tests passed (exit0,8.189s):89 exchanges/2 children for page/convergence; capacity10+29/3 children; fresh minima31/2 children. This is159 positive exchanges and7 child sessions. The separate derivative is clean commit `55d9ffc433b7014c72f4d2ad624f005eb4acb3f3`, with only the fresh owner eligibility check bypassed; installed payload `dc0b68e5ebca6aae58a34881c4490a245fc75a312f7289738d94b281e1ee213e` differs. Its2 exchanges/1 child produce the intended **exit1** failure, `expected 200 to be 503`, at the same `snapshotTimeout` assertion. It is recorded as failed derivative evidence, never a positive qualification result.

Totals reconcile exactly to accepted161 exchanges/8 sequential child sessions, of which33/3 are new evidence. All recorded children exited0; their PIDs were no longer present when checked. SourceCommit is launcher/source-bound metadata, not independent target attestation. Per-session policy/auth/config identities remain genuine captured values, with binding equality across each reopen rather than fabricated equality across separate random-credential fixtures. `claimEligible` remains false throughout.

## Limits and cleanup

No broad full suite, packaging rerun or redundant installed cohort was started by this reviewer. No owner process remains from the targeted tests; no source, branch, flags, report history or outward state was changed. The frozen source stayed clean. The inspected fresh artifacts supersede the earlier “packing pending” state for this review, while remaining clearly attributed to the author's actual execution. Root retains integration/publication decisions and any final inherited council gates.
