# Installed half-open CONNECT regression

Frozen commit `94c863b` on `codex/vickie-installed-connect-regression`, checkout `/Users/dbox/repos/wt/vickie/vickie-installed-connect-regression`, based on exact `b50c5e8d91336b087966cd3374de7e72513d6906`. GitHub origin configured; no push or other outward action. Base checkout unchanged. Clean working tree after commit.

Only three files changed:
- `apps/bdptest/src/installed-reference.test.ts`: permanent real emitted installed-child regression.
- `scripts/installed-reference-observation.mjs`: reset-loop comment now claims continued service only.
- `docs/design/update-implementation-progress.md`: explicitly labels captured installed schemas.

The new test creates a Bead, connects an `allowHalfOpen` TCP peer, pauses it, sends CONNECT, and waits for actual refusal bytes buffered on the peer without consuming them. It verifies the client has sent no FIN before and after child shutdown. The real launcher close must return code 0/no signal within two seconds; the same directory then reopens with create:false and returns the exact stored Bead. Cleanup destroys the test peer before awaiting child cleanup, including the deliberate negative path. No sleeps, fake lifecycle outcomes, runtime edits, public/configuration changes, conformance or transaction files.

## Verification

Pinned Node 24.16.0 / pnpm 11.20.0.

- PASS frozen install; clean, pre-build typecheck, build (`install.log`, `build.log`).
- PASS focused installed-reference and public node-http-listener suites: **30 tests / 2 files**, 5.04s (`tests.log`, `tests.json`). This includes the real emitted installed child and public sibling shutdown behavior.
- EXPECTED FAILURE: removed only `() => socket.destroy()` from the emitted installed runtime's CONNECT end callback. New test fails after two seconds with **installed shutdown blocked by non-reading CONNECT peer**, not launcher SIGKILL or a generic outer timeout (`mutant.log`). One failure / 13 filtered skips. Emitted bytes restored in Python finally; no source mutation.
- PASS restored new test: **1 test**, 408ms test time (`restored.log`).
- PASS scoped Biome format and lint, dependency boundaries (`format-lint.log`), `git diff --check`.
- PASS packed executable smoke, including ordinary public negative admission and existing installed candidate observations (`packed.log`). The new half-open test itself runs against the emitted actual process entry; the existing packed smoke retains its independent CONNECT/reset observations and is not claimed as an execution of this new test.
- An optional `biome check` invocation rejected import organization in the two existing files (`checks.log`). The repository's separate format/lint commands passed; no unrelated import sorting applied. No hooks skipped.

Process inspection after mutant, restored run and packed smoke found no test-owned children remaining. No full suite requested or run. This pins the current installed shutdown behavior and actual reopen, not every kernel socket error timing or a new profile qualification claim.
