# Frozen recovery-low successor

Head: `5e7041bcd64c58d84ff671e6cc3a2bb9a6c8b9ee`
Parent: `2fa2d0291d0a4dec95bc8b4e864c8f30c00b74b8`
Checkout: `/Users/dbox/repos/wt/vickie/vickie-transaction-recovery-low`
Branch: `codex/vickie-transaction-recovery-low`
Origin: `https://github.com/gastownhall/bdp.git`
Clean, frozen, no outward writes. No previous review source altered.

## Review dispositions

- F1 fixed. The missing-plan branch uses disclosedReceipt without retention and
  parses only the completed-status invariant. A disclosed completed parent missing
  its durable plan remains a storage-integrity failure. Failed parents return
  missing without expiry. New real HTTP GET/HEAD matrix covers ordinary, unsupported,
  behind and foreign-view controls for original/foreign principals; all 404, no
  serving context, and the expired-eligible stored failed receipt remains identical.
  Existing long-interval seam isolates independent sweeps. The red run returned the
  correct 404 but exposed an empty receipts table instead of the retained failed row.
- F2 fixed evidence. Existing actual after-terminal SIGKILL/hot-journal test now
  opens with incompatible page maximum and expiry-eligible clock, asserts installation
  mismatch, then correctly reopens. recoveredPending=1, all 60 receipt chains remain
  available and 60 groups/180 Resources persist. This distinguishes journal recovery
  from prohibited BDP recovery/maintenance. No byte-identity claim for hot recovery.
- F3 deferred broad tagged-union redesign as directed. The current `kind in observed`
  check is only within supplied===unsupported, not ordinary minimum reads; the
  review's predicted ordinary-minimum misrouting does not describe this control flow.
  No unnecessary type redesign or claim of future-proofing introduced.
- F4 already closed by parent2fa: private test-only maintenance interval and real
  delayed request witness. Inherited unchanged, regression suite includes it.
- F5 fixed guard/evidence. Check parsed parentId receipt namespace before plan query.
  Test both unrelated namespace and collection-shaped receipts/pages/token, with
  ordinary/unsupported controls, and assert no receipt_pages query. Red run showed
  the old collection-shaped parent reaching that query. Service prefix checks are
  explicitly documented as defense in depth; result-size bounds remain independent.

Six files: store (10-line delta), service (3 comment lines), HTTP test, receipt-page
unit test, recovery test, progress note. No format/schema/wire/profile/policy changes,
no new preflight. Existing format7 and all recovery/fencing semantics remain.

## Verification

Pinned Node24.16.0/pnpm11.20.0, frozen install. Build before emitted HTTP/process tests.
- RED: 2 selected new tests failed, 50 filtered; `red.log`.
- PASS: build, then 215 tests/5 files: transaction-http, receipt-pages, recovery,
  store and service. Both actual hot-journal process-death tests passed. Duration3.73s.
- PASS: typecheck, boundaries, format check, lint (existing3 warnings/1info), both
  installed executable/development smokes, existing89-row Read evidence verification.
- PASS: git diff --check. Clean commit. No hooks bypassed, no interrupted runs or
  test-owned processes left. Logs alongside this report.

No full-workspace rerun: root handles combined coverage. These tests establish
application-crash/SQLite rollback behavior, not power-loss/filesystem durability.
No profile readiness or new evidence cohort claimed.

## Next plan

`../transaction-after-core-plan.md` revised per root: existing factory does not
persist the bundle. Approved minimal immutable captured-artifact view on branded
DevelopmentContracts plus exact bounded capture in locked TX installation metadata,
unchanged identity, offline reconstruction, prior-format refusal if needed, schema-free
fallback and owned refusal retained. RU durability is a separate follow-up. No
contract-runtime implementation started; awaiting council/root scope.
