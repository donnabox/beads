# RU retained-artifact review follow-up

Frozen clean source `6413f29fa1053f5a9118c43ba4f00d7b7a51e1f2`, direct successor of
`8e3781ca78bc69571f6baa0e27d8d1f593986694`. Checkout
`/Users/dbox/repos/wt/vickie/vickie-read-update-retention-followup`, branch
`codex/vickie-read-update-retention-followup`; GitHub origin. No push/outward action.

Seven files, 80 insertions/6 deletions. No shared accessor, TX store, protocol,
profile, installed artifact or sealed evidence changes. Only behavioral correction:
a missing mandatory contractCapture row now reports integrity / missing store
metadata: contractCapture. A present mismatching row remains store-mismatch.

## External finding dispositions

1. Medium duplication: per root, RU ceiling comment explicitly references TX
   captureMaximumBytes and the exact differing length gates. Deferred consolidation
   is tracked in docs/design/update-implementation-progress.md. No shared constant
   or TX refactor during parallel work; duplication itself deliberately remains.
2. Missing capture: corrected and discriminated from mismatch. Both configured and
   built-in missing-row controls failed before the two-line implementation change.
3. Caller-side bound: accepted disclosed residual; direct recovery-store callers
   remain responsible for bounded expected text. No low-level API expansion.
4. Permanent predecessor-shaped fixture: format3 plus absent capture row, retained
   disposition and pending claim. Reopen asserts exact format-refusal reason/message
   (not new integrity gate); main file byte-identical and key rows unchanged.
   Explicit creation cannot overwrite it (EEXIST), also byte-identical.
5. member-identity stale comment now names format4.
6. STATUS qualifies ordinary-file byte preservation versus necessary SQLite hot-
   journal recovery before readable metadata/BDP refusal.
7. All 14 corruption variants now assert exact reason and gate message; inventory,
   present capture mismatch and missing capture can no longer satisfy each other's
   oracle. Existing byte/pending preservation assertions remain.

## Verification

Pinned Node24.16.0 / pnpm11.20.0. Frozen-lockfile install and build passed.
Red run: 2 expected missing-capture failures, 1 filtered test (`red.log`, exit1).
Initial green run: 186 passed, one new test's explicit-create oracle incorrectly
expected store-mismatch rather than existing EEXIST. Corrected that test only.
Final `recovery-store`, `development-retention`, `member-identity`: **187 passed,
3 files**, 1.72 seconds, `tests-final.log`/`.json`/`.exit` (0). Includes real hot-
journal and startup preservation regressions in recovery-store.
Typecheck, dependency boundaries, format (including final rerun), lint and existing
Read evidence verification passed; logs alongside this report. Git diff check clean.
Commit signed off; no installed Git hooks were bypassed. No source changes after
freeze. No full workspace suite, new packed cohort or public reseal run: this is a
bounded diagnostic/test/documentation follow-up, not renewed profile qualification.
