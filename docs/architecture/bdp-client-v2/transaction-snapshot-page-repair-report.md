# Snapshot page-boundary repair

Frozen source **a0a068c689f5588701c2b7a350bb677dc4f114ae**, direct successor of reviewed **b9afb271ef089b0ed204532fe7b1f153813450b3**. Clean isolated checkout `/Users/dbox/repos/wt/vickie/vickie-snapshot-page-boundary-repair`, branch `codex/vickie-snapshot-page-boundary-repair`; origin `https://github.com/gastownhall/bdp.git`. Parent author and native/external review checkouts remain untouched. No outward action. Diff: `transaction-snapshot-page-repair.patch`.

## Native F1: closed by bounded source correction

Only the packing calculation changes: framing now reserves `pageToken(seed,id,kind,ordinal + 1)`, the exact continuation a nonterminal page emits. Reserving that URL even for a terminal page is conservative because actual `next: null` is smaller. This covers decimal-width transitions without a guessed allowance. The checked 4 MiB page limit, existing complete-page validation, 32-item limit, R, total capture/live capacities, token policy, store schema/format and immutable stored ranges are unchanged. Previously persisted ranges are not repacked or migrated.

The permanent regression exercises **both 9→10 and 99→100**. Offline complete-record capacity fixtures have 305 and 3,185 records respectively, each <=R and aggregate below the owner's 32 MiB bound. They are explicitly **not mutation-admission evidence**. Nine or99 count-limited pages precede a near-byte-cap page. Sixteen records would fit exactly using the current token but exceed the cap by precisely one byte with the emitted next token. The repaired owner chooses15, puts the remaining records in the successor, and every serialized page remains <=4 MiB. Each fixture's complete Bead stream is followed through termination; the entire Bead ID inventory is independently compared and the same retained capture is refetched/retraversed after reopen.

Red control on unchanged parent production code: **2 failed /27 intentionally skipped**, each with `Snapshot integrity: serialized page bound`, reproducing the native finding at both transitions (`snapshot-page-repair-red.log/.json`). Green affected suite: **277 passed /7 files,17.62s** (`snapshot-page-repair-green.log/.json`): Snapshot, store, Events, changefeed, materialization, HTTP and consistency.

## Native evidence Low: description corrected

The permanent abrupt-exit test is renamed to the observation it establishes: a tentative handle is not issued after reopening, and its graph survives. It does **not** inspect a pre-reopen rollback-journal header or dirty database pages, so it does not independently prove actual hot-journal recovery. The prior `transaction-snapshot-core-report.md` is preserved as historical; its stronger journal-recovery statement is superseded by this narrower account. No added hot-journal proof is claimed. The separate actual parent-build compatibility/unchanged-file-byte proof remains valid and distinct.

After that test-title-only correction, final typecheck and the affected named controls passed: **3 passed /26 intentionally skipped,3.15s** (`snapshot-page-repair-final-controls.log/.json`). These run the abrupt-exit/reopen observation and both decimal-boundary cases. All assertion bodies in the seven-file green run are unchanged by the title correction.

## Checks and limits

Node24.16.0/pnpm11.20.0, frozen-lock install. Clean→pre-build typecheck→build PASS (`snapshot-page-repair-final-build.log`). Boundaries, full format, lint, diff check, unchanged89-row Read evidence verification PASS (`snapshot-page-repair-final-checks.log`); lint has71 warnings/1info, exit0. All test-owned subprocesses are synchronously reaped and fixtures removed. No full repository suite or new packed cohort was run.

No HTTP/service/runtime/evaluator/catalog/schema/public-admission changes. This is owner page-packing correctness, not Snapshot protocol qualification. Creation conditional preflight, authorization and expected-failure context pairing remain S1b; Q1 retained-minimum comparison remains unresolved. Existing immutable malformed captures are not repaired; this correction prevents newly captured ranges from making the width-transition mistake. Root owns integration/publication.
