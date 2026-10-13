# Native closure: internal payload symlinks

Frozen `4cf6acd03774fd90e629ba63516ff1365865b00e`, direct successor of reviewed `39a76748329fa88eee6cd1ca31b9501c5ef7f395`. Independent detached checkout `/Users/dbox/repos/wt/vickie/bdp-update-payload-symlink-native-review`. Two-file source/test delta reviewed; no source/outward edits.

**P2 closed.** The actual factory now rejects symbolic links within traversed package payloads, checks a directly read package manifest with lstat, and checks explicitly traversed dist/schema directory roots before reading. Ordinary Node-resolved dependency roots remain supported. This is a bounded refusal policy rather than new symlink-following semantics.

Reran the original isolated actual-fast-uri reproduction, with only successor checkout and refusal/no-store expectations changed. The linked actual-utils.cjs still proves it was loaded (stderr `NATIVE_CHANGED_UTILS_EXECUTED`), then the factory rejects before qualification readiness and before creating the configured store (`storeExists:false`). This intentionally does **not** claim prevention of module initialization execution: ESM/CJS loading occurs before the factory guard. The fix closes misleading identity/readiness admission, as requested.

Maintained real-process tests independently pass for file, directory and manifest symlinks. Each verifies startup refusal, absent store and rebindable listener port. Positive real-process tests keep the protocol package-root symlink and real pnpm transitive dependencies, prove normal readiness/reopen, and retain regular-file fast-uri byte-change identity discrimination. No shared package bytes were changed.

## Verification

Pinned Node24.16.0 / pnpm11.20.0. Frozen install and build pass. **13 tests passed in one focused file,4.52s,maxWorkers2** (`apps/bdptest/src/installed-reference.test.ts`). Original external probe passes with expected refusal. Logs `update-payload-symlink-native-{install,build,focused,repro}.log/.exit`, structured `update-payload-symlink-native-focused.json`, exact probe `update-payload-symlink-native-repro.mjs`, all in the same external log directory. No interrupted runs or source changes; both checkouts clean.

No new concern in this bounded delta. Broader54 tests/checks/packed smokes/evidence remain author results, not rerun here. No claim of sandboxing, concurrent filesystem mutation defense, general package attestation, public Write admission or full-profile qualification. Node/deployment/environment-injected execution remain outside the declared candidate payload identity. Prior review's symlink finding is closed; unrelated external review stream retains its own scope.
