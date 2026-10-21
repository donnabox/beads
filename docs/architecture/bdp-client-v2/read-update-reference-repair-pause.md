# Paused on operator request

Checkout: `/Users/dbox/repos/wt/vickie/vickie-read-update-reference-repair`
Branch: `codex/vickie-read-update-reference-repair`
HEAD: `579019b676568345f93bb424b2fa6f5774211df2` (base; repair is uncommitted WIP). Original579 checkout is untouched. No push, PR or outward coordination changes.

Three dirty files preserved: app Read+Update matrix test, RU manifest, RU coverage inventory. Exact current file hashes are in `paused-files.json`. Changes reuse successful B seed alias with an absolute canonical target and seed Link with an opaque pinned source, pin the complete normalized endpoint, add named identity/stale-record collection controls and shape/category countercontrols, and clarify bounded coverage reason. Both A/B whole flags remain false.

Completed evidence: frozen install exit0, initial typecheck exit0, build exit0; deliberate comparator red test failed against predecessor local alias spelling as expected (`red-comparators.log/json/exit`). After changes, focused two-file run completed **21 tests passed / 121 filtered skips**,12.33seconds, exit0 (`focused.log/json/exit`). Actual new collection controls failed exactly their intended assertions and therefore passed their negative-control tests. No active test existed when pause arrived; process inspection afterward found no test-owned child, so no run was terminated.

Outstanding: external `audit.py` failed its original47 raw-scenario-byte preservation assertion. The first serialization pass had escaped non-ASCII old literals; an ensure_ascii=False rewrite was applied, but the audit still failed and has **not been diagnosed or fixed**. Do not claim raw-byte preservation or final source/byte/node ledger. Resume by comparing raw old blocks, restoring unrelated formatting/spelling while retaining only B changes, then rerun audit. The focused run passed despite this bookkeeping/source-preservation defect. Preserve the failed audit as a real incomplete validation, not a product failure.

No final type/static checks, full affected142-test run, fresh packed171-test run, final budget ledger, clean commit, freeze or independent review occurred. No full repository suite was planned. Historical reserve1500 remains historical; actual comparator reserve662 is recorded in dispositions, but final repair capacity remains unverified pending audit. Next work is only this bounded repair after operator resumes; no new slice started.
