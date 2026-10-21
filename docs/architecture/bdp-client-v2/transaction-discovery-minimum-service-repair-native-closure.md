# Discovery minimum service repair — independent native closure

## Findings

**No actionable findings remain for the four accepted external code-review items.** Frozen repair **739022601416346fa68f8f027f9b1c62d8dd2b1b**, tree **7940c8e7723ae236f6a0047a916cce979e333269**, direct parent **2bd484367f63c7b8cc25e144baec983c5e579798**. Checkout is clean. The complete diff contains only transaction-service.test.ts and the progress document; production service, consistency, store and all other production bytes are unchanged.

Read the completed external review, root code dispositions, exact successor diff and final repair report/summary. This reviewer did not author this repair.

## Four-finding disposition

1. **Bounded held waits:** both affected catch-up and pending-abort tests now use synthetic budget100 and `now: () => tick++`. The held wait still prevents retry until explicit release, but a faulty catch-up/abort escape after release reaches a finite deadline instead of starving Vitest's macrotask timeout forever. Intended one-wait/two-attempt behavior is unchanged; cleanup still registers release before draining the service in reverse order.
2. **Meaningful close pending assertion:** the drain test awaits `setImmediate` while its wait remains held before checking close is pending and the owner remains accessible. This allows a wrongly non-draining close to complete before the assertions. Its finally block releases the gate and advances the clock; later assertions still require the correct timeout and actual owner shutdown.
3. **Durable report citation:** the bare machine-local report path is replaced by a commit-pinned fork report URL at4925b1b2d37de792b6ec35f89f29ba23fb1210d9; the duplicate blank line is removed. The text retains the distinction between source/direct-emitted evidence and installed HTTP discovery. I checked the literal diff, without network access; root's publication record supplies the URL's remote provenance.
4. **Explicit refusal and prefix oracles:** the test now pins reentry attempts and zero waits/assembly, installs a shared wait spy on the delegated service, resets the construction attempt counter, and requires one attempt per delegated throw/refusal and closed-owner failure. The delegated thrown cause retains identity; the refusal wrapper retains its expected message/cause. Successful catch-up checks exactly **two** assemblies in orderpos-1 thenpos-2, with the latter argument identical to the waiting result's frozen context. Two is correct because the test intentionally performs an independent satisfied observation while the original call waits; the external suggestion of one would reject the accepted behavior. No failed attempt is assembled.

These changes strengthen existing test evidence without modifying any assertion of protocol semantics, deadline policy, runtime behavior or public capability.

## Evidence and limits

Independently matched both changed file hashes to source.json and confirmed the two-file-only Git delta. Inspected the actual focused JSON: **183 passed,0 failed,0 skipped,6 files**. The recorded pinned-toolchain install, formatting, prebuild typecheck, build, lint, boundary and Read-seal commands all exit0. This is inspection of the author's execution evidence, not a reviewer rerun.

The repair report correctly makes **no fresh packed or direct-emitted claim**. Original2bd sensitivity and emitted evidence retain their original source identity; production is unchanged. No production repair is inferred from these test/doc findings, and no installed/public discovery qualification is granted.

Independent review activity was read-only: no tests, builds, fixture execution, network, source changes or publication. Only this external closure report was written. Root retains exact-head CI, ordinary integration and merge-tree verification gates.
