# Native review — T2a private Transactional service

**No actionable findings or blockers identified in this bounded increment.**

Reviewed frozen `bddc4486fe39676bb84d63e05dda22f1bfa77a1a` against its sole parent `1949f25cbd0e1c639a26e482f4a82763acf8aa20`, read-only checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-service`. The complete six-file delta includes the private service and tests, atomic receipt-context read and store tests, implementation progress and smoke inventory. The author tree remained clean and unmodified.

## Conclusions

- **One owner and explicit lifetime:** successful service construction owns shutdown under `ownerLifetime: 'service'`; failed construction leaves the supplied owner with the caller. The service uses one owner and existing Read plane, submission/comparison policy and sequence stepper. Accepted asynchronous work enters the active set before execution. Close refuses new calls, drains accepted work, closes the Read plane and then the database. Previously acquired Read facets pass through the same service admission check. Detached delivery does not cancel accepted mutations.
- **Scheduling and failure cleanup:** sequence turns use the existing scheduler contract. Inline or repeated callbacks do not double-execute members. Scheduler failure aborts only the attempt's remaining reservations, preserving committed prefix and foreign pending work. Reservation-cleanup failure fences service entry and retains original and cleanup causes without introducing an error-cause cycle. Shutdown attempts owner cleanup even when cursor/Read cleanup fails; repeated close and the `closed` promise retain the same outcome.
- **Atomic receipt observation:** `readReceipt` performs principal-bound lookup, any lawful expiry update, detached/deep-frozen receipt capture, and current epoch/view/head selection within one guarded SQLite transaction. It reuses existing receipt retention and parsing behavior. The service does not pair an older execute result with a later diagnostic position: direct delivery re-observes the current retained receipt and context together. Historical result postimages and original execution positions remain unchanged as the current serving position advances.
- **Identity and disclosure:** receipt access uses trusted principal input, independently of actor attribution. Unauthenticated service reads refuse before lookup. Foreign-principal, unknown, root, retracted and forgotten receipt lookups have the same absent result without exposing context through the service response. This owner remains explicitly fixed all-visible; the slice does not claim variable-grant or erasure projection support.
- **Retention and recovery:** completed compact identities remain after detail expiry; expired failed receipts are forgotten. Pending duplicates do not execute twice. Abrupt process exit tests recover admitted/tentative work or preserve delivered terminal state, and retry produces exactly one group. This is process/SQL fault evidence, not disk or power-loss qualification.
- **Inline bound:** the service requires advertised `page.maximumItems >= 32` because this supported owner can emit at most 32 singleton result entries. The guard and maximum batch test conform to the existing receipt inline-result counting rule; they do not assert a protocol-wide minimum. No receipt page URL is manufactured and no complete-response byte limit is advertised.

The current spec's Mutation Receipt responses and current-view delivery clauses (`docs/specs/bdp.md`, approximately 5090–5240), principal/non-disclosure ordering, and accepted-sequence execution rules were checked. HTTP methods/headers/conditions/minimum-position barriers, sequence serving context, durable receipt paging, variable policy/erasure, set expansion and replication paths remain future work. Their absence is not newly presented as full profile support here.

## Independent validation

Created an exact-commit private archive in `/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-service-native/source`, with private workspace package builds and reused external dependencies. Explicit Node24.16.0 build completed before emitted child-process tests. No source branch was modified.

- PASS build (`transaction-service-native/build.log`).
- PASS **198 tests / 5 files** (`transaction-service-native/tests.log`): transaction-service, transaction-store, transaction-sequence, transaction-read and transaction-submission. Includes same-owner afterstate, cross-carrier keys, pending/principal rules, retained/current receipt distinction, atomic SQLite entry assertions, scheduling/cleanup faults, finite comparison, drain and three actual abrupt subprocess exits.
- PASS **21 tests / 1 file** (`transaction-service-native/native-probes.log`): nineteen service tests repeated plus two independent private-archive probes. Repeated tests are not additive unique coverage.
- PASS dependency boundaries for all 11 packages (`transaction-service-native/boundaries.log`).

Additional probes:

1. Expire a real completed receipt, inject an error **after the actual receipt-read COMMIT**, and verify the read throws/fences rather than delivering a guessed observation. Close/reopen recovers the committed compact expired receipt at the unchanged actual position; exactly one group remains.
2. Inject a cursor cleanup error after actual cursor release and verify service shutdown still closes the owned database, rejects with the cleanup cause, exposes the same rejection through `closed`, and returns the identical promise on repeated `close()`.

No test failures, timeouts or interrupted runs occurred in this independent review. The separately known pre-build type-only `dist` import problem is addressed by root's disjoint clean-checkout fix and is not classified as a new T2a runtime issue. This reviewer did not rerun a clean-checkout typecheck, full suite, packed smoke, combined integration tree or fresh conformance cohort here.

## Integration and evidence boundary

The new service module remains absent from the public root, with additive emitted-file inventory and package deep-import denial coverage. No HTTP adapter, protocol schema, storage/identity format, Read observation evidence artifact or public profile advertisement changes occur in this delta. Author-reported broader checks are separately recorded in `transaction-service/report.md`; this review does not relabel them or the historical Read cohort as new serving qualification.

Root may integrate the reviewed service seam and the separate type-only test repair. T2b HTTP work must continue to obtain receipt bytes and current serving context through this atomic owner read, preserve accepted-work ownership, and keep unfinished protocol paths explicit. No outward writes or author-source edits were made.
