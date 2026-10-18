# Native review: Snapshot runtime binding delta

**No actionable findings in the four-file binding delta. Accepted within this scope.**

Frozen source `5208c621ac810a16477d589585a89be03e4eb632`, compared with merge base `9b877f8be5d2c6281695fd701c0c6f59587f4c7d`, in `/Users/dbox/repos/wt/vickie/vickie-snapshot-runtime-integration`. Read the root author report. Source inspection is read-only; independent clean build/tests/packaged smoke used `/Users/dbox/repos/wt/vickie/vickie-snapshot-binding-native-review`. Both checkouts remain clean. No external review used.

## Binding and boundaries

- `transaction-store.ts:539` exposes `scopeSnapshot` through the owner's immutable read configuration, using the same compiled `TRANSACTION_SNAPSHOT_LIMITS` and `TRANSACTION_SNAPSHOT_TOKEN_POLICY` as the Snapshot owner. The limits and containing tuple are frozen. Existing persisted compatibility construction at lines245–255 uses these same constants; startup compares the stored configuration before returning the live owner. The delta does not create an override or derive authority from caller-supplied compatibility JSON.
- `transaction-runtime-configuration.ts:89` incorporates the Scope Snapshot tuple into recursively frozen captured settings, separate from the query-pagination cursor settings. The existing runtime identity hashes the entire captured settings object. The unchanged pure dependency test visits every default leaf, so it now includes all Snapshot limits and token policy.
- `transaction-runtime-configuration.ts:327` checks full structural equality against the actual owner's tuple before returning metadata or reaching the injected-seam early return. The runtime obtains that argument from `owner.readConfiguration()` (`transaction-runtime.ts:230`); it is not an external untrusted assertion of identity. Lifetime or policy disagreement throws instead of producing a misleading identity.
- The new test opens a real SQLite owner, cross-checks parsed stored compatibility tuple, frozen runtime facts and digest construction, distinguishes Snapshot lifetime from cursor TTL, and negatively tests altered lifetime and token policy. Existing emitted runtime tests exercise the ordinary construction/cleanup path. No new Scope Snapshot HTTP route, public admission, Q1 comparison behavior or conformance claim appears.
- `scripts/smoke-executables.mjs:243` inventories all four emitted Snapshot files. Lines465–466 explicitly require `ERR_PACKAGE_PATH_NOT_EXPORTED` for bare and dist package specifiers; the existing transaction-module loop also covers the dist path. `packages/server/package.json` remains unchanged with only root/development public exports. Installed package smoke passes with the new inventory and denial checks.

## Independent validation

Pinned Node24.16.0 / pnpm11.20.0, frozen dependency install:

- PASS clean pre-build typecheck.
- PASS build (including clean).
- PASS156 tests /3 files in4.48s: `transaction-runtime-configuration.test.ts`, `transaction-runtime.test.ts`, `transaction-store.test.ts`. Includes actual emitted runtime lifecycle/reopen tests, binding/mismatch tests and existing store checks.
- PASS `pnpm smoke`: packs/installs the actual packages, verifies emitted inventory and private import denials, exercises public Read admission and fail-closed profile guards, and closes observed installed processes.
- The first direct `node scripts/smoke-executables.mjs` invocation failed before allocation because it requires pnpm-provided `npm_execpath`; retained as `smoke.log`. Corrected canonical `pnpm smoke` completed exit0 (`smoke-pnpm.log`). This is an invocation correction, not a source repair or ignored test failure.
- No test-owned process remains; no source edits, commits, pushes or other outward actions.

Logs: `/Users/dbox/repos/wt/vickie/bdp-integration-logs/snapshot-runtime-binding-native/` (`install.log`, `typecheck.log`, `build.log`, `focus.log/json`, `smoke.log`, `smoke-pnpm.log`).

The underlying Snapshot capture/page/expiry/corruption semantics and their separately assigned core repairs were deliberately not re-reviewed. This acceptance does not supersede those findings or qualify Snapshot HTTP/profile behavior. Full repository regression, development smoke and static checks were not independently repeated; the author's256-test/check/smoke results remain separately attributed author evidence.

## Addendum: independent metadata-preimage omission mutant

After recording the native review, inspected the external binding review's claim that filtering `scopeSnapshot` from the actual metadata hash input would escape the new test and instead be detected by the pure leaf test. **That concrete claim is disproved.**

In the disposable review clone only, replaced the `settings` property of the actual `runtimeConfigurationIdentity` preimage with `Object.fromEntries(Object.entries(settings).filter(([key]) => key !== "scopeSnapshot"))`. Left defaults, returned captured settings, equality guard, digest helper and tests unchanged. Ran exactly the new owner tuple test and the pre-existing default-leaf test.

- Mutant: exit1, **owner tuple test fails at test line425**, while the pure leaf test passes (1 failed /1 passed /19 filtered,0.551s).
- The owner test independently hashes complete `captured.settings`, so its expected `ff437e2bef53649c227691a8d663aed52d2b1224a30216290c683e3f71312aff` differs from filtered metadata's `b7d9fe44fd125fbbe5b44cff5fa1c105620deb7456384c30f3aa45bc4ebe72ad`.
- Restored exact original file bytes in a `finally` block; both tests then pass (2 passed /19 filtered). Review clone `git status --short` is empty. No frozen author source changed.

The pure leaf test proves sensitivity of the digest helper when given modified defaults, not that the actual metadata caller includes those defaults. The new owner test supplies the missing caller-preimage check. No source/test addition is needed for this hypothesized regression.

Evidence in `snapshot-runtime-binding-native/`: `metadata-filter-mutant.patch`, `metadata-filter-mutant.log/json`, `metadata-filter-mutant-exit.json`, and `metadata-filter-restored.log/json`. This is one deliberate rejected mutant, not a product failure or broader mutation-testing claim. Optional defensive guard commentary and packaging history do not change the original no-finding result.
