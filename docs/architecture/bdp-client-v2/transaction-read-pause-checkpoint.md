# Private same-owner Read slice — frozen for review

Head: c4fce0b44b297dc2cbc022c4dd84485eda861a4e
Base: cd22ac3e388be3baffaf1d3a4caac7c4bd288502
Checkout: /Users/dbox/repos/wt/vickie/bdp-transaction-read
Branch: codex/vickie-bdp-transaction-read
Origin: https://github.com/gastownhall/bdp.git

Clean local commit; no push, PR, bus or launch writes. Operator pause reached after this bounded slice. No source-position/HTTP follow-on started. No runs interrupted or stopped; test/smoke processes completed normally.

## Change

AuthorityReadPlane now requests only resource/resources/incidentLinks/alias/policy, keeping RU store/key semantics intact. The Transactional store provides a synchronous expiring reader under its existing exclusive connection/entry guard. Its row snapshots and collections are frozen; facade methods expire on return/failure, async/thenable callbacks are refused with rejections observed, and reentrant read/write/close is refused. Sticky owner and maintenance failures refuse entry; storage-failure causes survive the Read plane wrapper. Actual immutable owner installation supplies Type descriptors and fixed configuration. No separate configuration inventory or fake RU key outcome exists.

The new private transaction-read composition uses the existing Read plane for Resource/properties/revisions, collections/selectors, incident Links, Types, aliases and retained pagination. Closing the plane disposes its cursors but leaves the caller-owned write store open. The nine files are the agreed authority-read/test, read-sequence-lifecycle test annotations, transaction-store/test, transaction-read/test, progress note and executable smoke allowlist.

## Verification

Pinned Node 24.16.0 / pnpm 11.20.0; frozen install.

PASS final 698 tests / 12 files, two workers, 4.61s: transaction-read, transaction-store, transaction-dependencies, transaction-sequence, transaction-submission, authority-read, read-sequence-lifecycle, recovery-store, sequence-executor, member-executor, read-pagination, read-selector. Log regressions-final.log. Earlier 444 tests / six files also passed (regressions.log).

PASS build, typecheck, dependency boundaries, format, lint (existing 3 warnings / 1 info), git diff --check. Individual logs in this directory. No lint baseline increase.

PASS installed-package smoke, including unchanged exports/denied transaction deep imports and exact private output inventory. PASS development installed-package smoke through actual SDK Fetch transport/restart; it remains non-attesting. Logs smoke.log and smoke-development.log.

PASS evidence:verify: unchanged Read constant 6142cf73deae687bc3e9668cc014ea408fa02843; run 4f0ed57906525c36bc55b343b77e0f07edbb7553; evidence 72bdaad02b6fb2b803d4af7a8c4f90ada2f0d2d5; schema projection 2db57a82d9bf99f3fb5d18646e8bd6c7646590d1e806aa266de8af415ee83545; 89 rows. No reseal.

Initial new-test runs exposed test fixture mistakes (pagination capacity, mutation request field shapes, expecting the underlying cause as the wrapper message); corrected before final run. They did not require runtime semantics changes. The final matrix proves canonical reads before/after actual commits, failed-batch rollback, matching Resource revisions, same-owner alias state across restart/repoint/deletion/dangling targets, stable pagination across later commits and expired cursors on close/reopen, callback expiration/reentry/async refusal, and injected before/after-COMMIT uncertainty with cause preservation and reopen recovery. Injection tests do not prove real disk/power-loss behavior.

## Remaining / explicit limits

Private schema-free/unowned fixed all-visible configuration only. No new public API, HTTP listener, profile advertisement, schema/spec/format change, receipt projection or migration. Source-position propagation and continuation source-context handling remain a separately scoped next slice requiring authorization; current private observations are not a Transactional HTTP serving contract. Minimum-position barriers, receipt reads, replication Snapshots/changefeed and broader qualification remain absent. No source code modified outside agreed scope.

Plan saved beside this report as read-composition-plan.md. Root owns independent review/publication and pause checkpoint. Update worker's smoke allowlist additions should merge additively; it owns routing/development files untouched here.
