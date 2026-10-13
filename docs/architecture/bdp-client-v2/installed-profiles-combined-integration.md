# Combined private write-runtime integration

Exact head `69c1e95493924f3aae3d962252cdb3b02d3a95e8`, draftPR88, codex/vickie-installed-profiles, remote verified. Ordinary ancestry: c104e35 (PR86+PR87+11c) → f85a135 (merge installed4cf) → d9b981760184d836e4f1fe43debeb16b2e298fe3 (test-only clean typechecking repair) →69c1e95 (merge Transactional7bf). Both new merges clean; no rebase/cherry-pick/force push.

CI38355238048 atc104 failed pre-build typecheck because installed startup tests used generated declaration imports. Clean local reproduction also found the same mistake in successor closuretest. Fixd9 changes only those two tests: dynamic URL-based imports still execute emitted JS; typeof source imports typecheck before dist exists. Clean→typecheck→build now passes. No workflow relaxation.

Exact final-tree full suite8,081passed/129files/one optional skip; format/lint(existing warnings)/boundaries pass; both installed/development smokes,7preflight tests and unchanged89-row Read evidence pass. Logs installed-profiles-combined-*.log. ExactCI38356316908 SUCCESS, full log pr88-69c-ci.log.

Native closes installed4cf symlink identity refusal; independent13tests and original actual-code probe. This is readiness/identity admission, not sandboxing: dependency code may initialize before factory checks. Externaldea closes startup/import/owner limits with7,924tests plus independent probes. New dea low/info followups do not reopen those findings; forged argv by host is outside bare-import guarantee, packed inertness separately verified by reviewer, future tuple key collisions/dead alias cleanup remain optional followups.

External7bf closes F-A-D with49tests/~250HTTPprobes; E/F/G/H explicitly retained. New command minimum-control auth ordering repair is assigned in the point-minimum successor; keep401/403 before privategap without capturing key/effect. CONNECT authority-form400 is a declared fixture limitation, origin-form405 tests do not imply support; non-operation Allow is shape-derived for nondisclosure; depth marker assertion is being pinned. No full Transactional qualification.

Actual cumulative Update executor310 is separately frozen/reviewed, NOT merged here. Native77tests and real unfiltered run confirm observation wiring but found missing no-op response equality and overclaimed update-bead whole-contract coverage. Reported124uncovered is unaccepted pending correction to at least125 or actual additional cases. No public profile activation, main merge or Preview2 promotion.
