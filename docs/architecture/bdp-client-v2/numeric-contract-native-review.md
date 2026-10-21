# Native review — numeric-admission diagnostic contract

Exact candidate `4f5f5b1c201f79411c6b88a076dae9c1053c8282` against its parent `b225fe5b4bcf42d3ac2116a1659412e4eec6efe3`, read-only source `/Users/dbox/repos/wt/vickie/vickie-numeric-admission-contract`. Full two-file delta reviewed. Author tree remained clean.

## Finding

1. **Medium — update the existing diagnostic regression expectations and byte boundaries before landing.**
   - **Locations:** `packages/server/src/member-identity.test.ts:1327`, `:1460`, `:1486–1508`; the changed behavior originates in `member-numeric-diagnostics.ts:24–29,44–46`.
   - **Evidence:** independent exact-commit private build passed, but running `member-identity.test.ts` and `member-executor.test.ts` produced **6 failed / 217 passed**. Both Bead and Link variants fail in each of three groups: paired create location normalization, unapplied properties patch locations/document relabeling, and exact numeric diagnostic budgeting. These failures reproduce without source edits.
   - **Cause and impact:** those tests still require callback-supplied `type` and `schemaLocation` on properties numeric failures and compute byte thresholds from that larger obsolete payload. The implementation now correctly strips them, so the frozen candidate leaves the directly related regression suite red and the byte-boundary expectations wrong. This is test migration fallout, not evidence that the new runtime behavior violates the spec.
   - **Fix:** keep callback inputs containing the bogus fields as negative coverage, but expect their absence after normalization. Recompute each exact byte threshold from the actual stripped diagnostic, preserving exact-bound/one-byte-under controls, caller nonmutation, create document locations, update location omission, document anti-relabeling and truncation assertions. Run the affected identity suite with the durable tests.

## Runtime assessment

The helper now matches `docs/specs/bdp.md:3459`: failures outside effective Type validation must omit `type` and `schemaLocation`. Both numeric properties create and update retain the caller's message while removing the unsupported Type claims; metadata already did so. Creates retain normalized document-relative pointers; unapplied updates retain pointer omission. The helper is shared by the ordinary member and private Transactional paths, and the shipped callbacks did not supply these Type fields, so this does not change their normal wire output.

All eight new durable member cases pass in the independent run: create/update Bead/Link × properties/metadata, contract evaluation must not run, state remains unchanged, refused create IDs remain absent, and exact failures replay after close/reopen without running the numeric callback again. No new schema/store format or public API is introduced. No additional runtime issue identified.

## Independent validation and limits

Private archive `/Users/dbox/repos/wt/vickie/bdp-integration-logs/u2-numeric-native/numeric`, explicit Node24.16.0, private workspace build with existing external dependencies. Build ran before tests.

- PASS build: `u2-numeric-native/numeric-build.log`.
- FAIL focused tests: **6 failed / 217 passed, 2 files**, `u2-numeric-native/numeric-tests.log`. Exact command: `node node_modules/vitest/vitest.mjs run packages/server/src/member-identity.test.ts packages/server/src/member-executor.test.ts --reporter=verbose --maxWorkers=2` using the pinned Node executable.

No full suite, new evidence generation, or outward/source writes. This candidate needs the bounded test correction above before native closure.

## Native closure — legacy diagnostic assertions

**Finding 1 is CLOSED at `c245cddeb2e0ee64ce58d1b27a9e9f6990df5cd7`. No new findings.** Verified that this is a descendant of frozen `4f5f5b1c201f79411c6b88a076dae9c1053c8282` and its entire delta from that candidate is the existing `member-identity.test.ts`. Inspected the full delta read-only in `/Users/dbox/repos/wt/vickie/vickie-numeric-contract-closure`; no runtime source changes or author edits.

The corrected tests still pass deliberately incorrect Type/schema fields, caller locations and metadata relabeling through the callback. Only the expected normalized payload omits those fields. Exact serialized-byte budgets now derive from that smaller actual payload. Create escaped/multibyte pointers, empty-key pointers, original callback nonmutation, exact-bound admission, one-byte-under first-diagnostic refusal, truncation, update pointer omission and anti-relabeling assertions remain meaningful. The change does not merely remove the failing boundary checks.

Independent verification used the prior exact 4f5 private archive plus the sole changed file read directly from c245; that produces the reviewed tracked source tree while preserving the original failure log. Explicit Node24.16.0 build completed before process tests.

- PASS build (`u2-numeric-native/numeric-closure-build.log`).
- PASS **366 tests / 4 files** (`u2-numeric-native/numeric-closure-tests.log`): member identity, member executor, Transactional store and Transactional submission. All six previously failing identity cases pass; durable retained/refusal/reopen and existing Transactional numeric-budget regressions remain green.

Root's separate initial missing-dist run is not counted as evidence here. No full suite or fresh conformance evidence was run by this reviewer. This addendum closes the test-migration blocker and leaves the original implementation assessment unchanged.
