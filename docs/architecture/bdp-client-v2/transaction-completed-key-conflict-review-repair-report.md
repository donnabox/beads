# PR110 completed-key-conflict native review repair

Frozen ordinary successor **b3b9327c262ba16eb6b8f994303f3a08f098c7a6**, tree **b7c42f6fc27e19db0d25b20c356f6ed458505e33**, direct parent **b13055049ed7c1bd09ad2d7af29e0a38a81001ea**. Checkout `/Users/dbox/repos/wt/vickie/vickie-completed-key-conflict-review-repair`, branch `codex/vickie-completed-key-conflict-review-repair`. Agent-Signature attribution trailer is present. Remote-session guard accepted session `2026-10-08T20:26:35Z-81e5`; own branch push completed and `git ls-remote` returned the exact head. No PR/main/bus/launch mutation.

## Disposition

Native P3 is addressed in exactly one file: `packages/conformance/src/transactional-cumulative.test.ts` (+191 lines). The independently authored guard pins the complete assertion arrays for all twelve completed-conflict actions (99 existing assertions, including request-wire assertions), plus every exact capture array. These include full retained Resource comparisons; complete state and Event tuples; both Problem media/schema/type/code/status/retry assertions; original body/record captures; three retained-body equalities and replay-root equality; and mandatory context presence on original receipt, both Event reads and both client receipt reads.

The expected literals and projection paths are authored in the test, not computed from the manifest under test. Small local constructors express existing assertion grammar; no new executor/helper framework or production API was added. Two static weakening controls remove `equalsBinding` from each retained-record assertion while retaining its ID, pointer and presence check. Both fail the same full-array guard at their named action. These are test-source regression controls, not additional installed requests or a sixth installed sensitivity control.

Root explicitly accepted preserving the current conflict-row order. Its Event presence assertions follow metadata equality, but every declared assertion must pass; the new guard pins those mandatory presence assertions and both comparison inputs. Reusing the M/Z presence-before-equality ordering convention blindly would change artifact bytes without strengthening the observation. The existing M/Z guard remains unchanged.

## Source and evidence preservation

`source.json` independently checks Git's complete changed-file list is exactly the single test file, and compares the manifest, coverage and fixture bytes to the frozen parent. Thus all nine old objects, the new row, every binding order, all false whole flags, runtime/package source, installed test, schema/catalog, deadlines, source caps and request counts are unchanged. Manifest407227bytes, coverage60059bytes and fixture716bytes remain exact. The new row remains12actions/99assertions; aggregate151actions. Hashes are in `source.json`.

No new packing or installed run was performed or claimed. The actual parent packed positive and five controls remain bound to b13055049ed7c1bd09ad2d7af29e0a38a81001ea, documented in [the original frozen report](../transaction-completed-key-conflict-implementation/report.md). This repair changes test guards only; it does not relabel that historical execution or assert unmeasured tarball hashes. No whole-contract or public-profile qualification follows from this repair.

## Validation

Pinned Node24.16.0 / pnpm11.20.0; frozen dependency install in the assigned checkout succeeded. Commands and actual exit codes/durations are preserved in `exits.json` and `summary.json`.

- PASS: clean pre-build typecheck and emitted build; final typecheck/build after the test-authoring correction.
- PASS: focused cumulative **28/28**, including both weakening controls. Repeated once on the clean frozen commit: **28/28**, 0 failures (`frozen-cumulative.json`).
- PASS: format, lint, dependency boundaries, Read evidence seal and `git diff --check`. Lint exit0 includes the existing repository warnings/info; no claim of warning-free lint.
- Not run: full workspace suite, unchanged app installed cohort, new tarball packing. Their scope did not change.

The initial cumulative attempt recorded27pass/1fail because the independent expected objects used `ref` instead of the existing `schema` field and string projections instead of `{pointer}` projections. The source test was corrected to the existing grammar; the unchanged manifest was never weakened. Initial `cumulative.log/json` remain separate from final and frozen passes.

The first preparation command omitted its working directory and failed before opening the test file. Its following dependency command ran in the default formula-language checkout and generated an untracked lockfile. After inspecting the exact timestamp/content, only that generated lockfile was removed; unrelated files were preserved. `preparation-note.md` and the initial install log preserve this operational mistake. The dependency install was then run successfully in the assigned worktree.

## Cleanup and limits

Checkout is clean. All owned validation processes exited; no test/build/installed children remain. No server or shared store was started by this repair beyond bounded local cumulative test fixtures, which completed normally. No source mutation followed freezing. Parent owns independent external review disposition, PR110 integration and readiness; this report closes only the accepted native P3 repair scope.
