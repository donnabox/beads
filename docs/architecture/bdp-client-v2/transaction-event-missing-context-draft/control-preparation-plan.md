# Concrete emitted-control copy and resolution plan

Preparation only. No build, import, test, copy, symlink, mutation or runtime was performed for this report. Source is88c9408e plus the reviewed redirect successor: bindingacf190e742dfb669fe5607085a3ac60c5ecb66f4f86b1f05025424a7873615b0; full patche0c25b193ef1191d10916728c54eb6a894c5f29992e0da768bd585ef089e98e3. The accepted control remains11 HTTP/one in-process owner/listener/directory/zero product children.

Use one disposable mirror at:
`/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-event-missing-context-draft/control/copy`
(C below), with positive workspace W at:
`/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence`.

## Copy arrangement

After separately authorized positive setup/build/checks, freeze a manifest of the positive input and build files. Refuse an already-used C; do not overwrite a failed attempt.

1. Mirror the actual working bytes of all git-tracked files plus the one approved new installed-test file, retaining relative paths. Do not copy .git or inherited runtime/temp/output directories. This simple mirror avoids inventing a trimmed TypeScript import closure or a new Vitest config. Current read-only inventory is664 tracked files/19,018,519 bytes plus the new15,956-byte installed test; recalculate at preparation, do not treat these as a future build inventory. No installed test is selected by the control.
2. Copy the complete positive `packages/*/dist` trees for all eight library packages after the build. Retain package manifests, schemas, source and test-support files from step1. Do not copy application binaries to run them. The actual built member list/hashes do not exist yet; capture them after the authorized build, never predeclare hashes.
3. Provide the already-installed dependency graph without running a second installation:
   - C/node_modules may be a link to the verified W/node_modules development dependency graph, allowing the pinned pnpm/Vitest tooling to run from cwd C.
   - For each copied package, create its own node_modules directory layout from the actual positive package links. Workspace `@bdp/*` links must point to **copied C/packages/<package>**, not the original W package; do not symlink the entire package node_modules directory back to W.
   - Non-workspace dependencies may link to the same verified real third-party package directories used by W. Record link text, realpath, package version/package.json hash and complete consumed dependency-tree inventory. Their transitive pnpm store links remain the original bound dependency graph.
   - Stop on an unexpected link/layout rather than guessing. At this inspection W has **no node_modules or dist directories**; these are explicit future setup/build prerequisites, not currently verified links.
4. Apply only the control.json replacement to C/packages/server/dist/transaction-http.js. Demand exactly two occurrences of `transactionHttpProblem("resource-not-found", observed.context)`; replace with the one-argument call. Record before/after hash and exact two-hunk diff. No TypeScript source mutant, rebuild, source-map rewrite or positive-workspace write.

The source-map files remain copied historical positive build files; generated JS is the declared mutation authority. If reporter mapping points into unchanged TS, adjudicate the exact named assertions and JS mutation record, not a claim that the TS source itself was changed.

## Why this resolves the intended modules

- C/vitest.config.ts computes workspaceRoot from its own import.meta.url. It scans all eight C/packages directories, reads their package.json files, and aliases @bdp libraries to **C source entries**; /testing and server/development aliases follow the same copied layout. This is why the mirror retains all package source/test-support, not only the selected test.
- setupFiles remains `./packages/server/test-support/read-conformance-mock.ts`. Its mock and mock-state imports are copied relative source paths. No custom config or setup omission is needed.
- The selected C/packages/server/src/transaction-events.test.ts has the explicit dynamic URL `new URL("../dist/transaction-runtime.js", import.meta.url).href`. Resolving it as a file URL yields **C/packages/server/dist/transaction-runtime.js**.
- The emitted runtime's relative `./transaction-http.js` import and runtime-configuration's same relative import resolve to **C/packages/server/dist/transaction-http.js**, the mutated file. Bare @bdp aliases do not rewrite those explicit relative file imports.
- Under Vitest processing, bare workspace imports may use the existing copied-source aliases. If an emitted dependency is instead resolved through native package exports, the copied @bdp links select copied package manifests/dist. Both remain byte-bound to the same positive inputs; do not falsely label this source-consumer control as packed-consumer qualification.
- Third-party modules (Vitest/Ajv and the server's fast-uri, plus protocol dependencies) use the unchanged recorded dependency graph. Native node: modules use the pinned Node process.

## Pre-dispatch checks and one invocation

The author's existing wrapper should record, without importing application modules:
- Positive source/build manifest and equality of every copied regular file before mutation.
- Realpath containment of the source test, its dynamic runtime URL, and every emitted relative runtime/HTTP edge inside C. Inspect the actual built import text; fail if the expected edge changed.
- Copied Vitest config/setup bytes and alias targets. Verify each workspace alias entry/testing target exists under C; no symlink escapes for workspace source/dist.
- Root/package dependency link manifest and real third-party resolutions, with positive/copied byte equality. Native and Vitest alias resolution are distinct inventories, not interchangeable claims.
- Exactly one changed JS file/two replacement sites after copying; source test hash remains0152561cdf096a48b0b01ec0e0d8bc81208a905bb5bf71dad411694387b58b27.
- No pre-existing run output or prior attempt marker. Clear inherited BDP overrides. Use C as cwd and the exact anchored control title already in commands.json; no installed-entry override or product child.

Run only the accepted source HTTP callback once under the existing120s supervisor. Its positive result was obtained separately; the derivative must produce one failed case with exactly eight named tuple failures,11 completed responses with200/200/eight404/401, and no other assertion, timeout or unhandled failure. Require the genuine source report's cleanup/deadline facts plus independent temp/process checks. The report's failureCount excludes soft matcher failures and is not a verdict.

Do not add a positive control rerun, standalone runtime import probe, installed mutant or extra request to prove graph identity. The exact copied paths/import text, actual dependency bindings and expected branch-specific eight-failure result jointly establish the control. Preserve all graph manifests, logs/results and cleanup evidence; remove only owned disposable runtime/copy resources after evidence capture according to the wrapper's accepted cleanup policy. No new council or generic harness is needed for this arrangement; root still owns execution release.
