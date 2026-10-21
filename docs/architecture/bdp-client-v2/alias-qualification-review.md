Worker verification report; root subsequently published the reviewed commit as a draft.

# Independent Update alias and pinned-reference diagnostic review

Reviewed frozen head `0965971b855bd7af0753bc40487a0e0efac1f0eb` against `5f07a51c0bc0e32652b02ab16cdb2fd4c7f76bc3` in isolated detached clone `/Users/dbox/repos/wt/vickie/bdp-update-alias-review`. Source checkout remains clean. No runtime, schema, store, exports, sealed evidence, or public artifacts changed by this review.

## Result

No actionable defect found in the bounded change. The assertions support the stated development-only observations; they do not qualify normative rows or admit production write artifacts.

The six added plans contain meaningful resulting-state checks, rather than relying only on response status:

- Allocation: read-back of the allocated Bead and submitted properties, canonical/in-Scope/single-child identity observations, rejection of explicit creation at a live alias path, and continued use of that alias. This does not force an allocator collision, and both README and coverage explicitly retain that limit.
- Invalid alias targets: rejected alias, Link, and external targets leave the proposed alias absent. An admitted failed key stays failed after a separate corrected request; changed identity conflicts. A Link through the corrected alias verifies its canonical target.
- Unknown subjects: failed alias puts/deletes remain retained after their formerly missing subjects become live. New keys can succeed, and Link/Bead reads prove that replay did not remove newly reachable state.
- Recorded targets: retained Link creation preserves the original canonical target across alias repoint, deletion, reassignment, and cross-slot use of the original locator. Fresh retry locators are resolved anew after repoint/delete and can become equivalent again after restoration. The original Link is read back.
- Recorded misses: retained misses stay misses after assignment and reassignment; a different missing locator or canonical target conflicts. Both Beads survive and a fresh Link proves current alias resolution.
- Pins: local references canonicalize while opaque revision strings and external URI spelling remain exact. Byte-distinct Unicode revision strings, removing a pin, and changing external URI case conflict. Read-back preserves the original Link; sequence-local creator references replay after local label changes.

Relevant plan starts: `packages/conformance/development/read-update-manifest-v1.json:6669,6902,7438,7807,8628,9046`. The helper changes at `apps/bdptest/src/development-matrix.test.ts:57–120` expose actual transport observations and preserve pin objects when expanding local URI references. Normative checks used frozen spec clauses at `docs/specs/bdp.md:642–644`, `2950–2980`, `4015–4019`, and `4263–4290`.

## Preservation and counts

Independently counted 23 scenarios, 221 actions, 705 assertions; all 81 coverage rows remain, with 48 bounded observations and 33 omitted. All previous 17 manifest scenarios and 17 diagnostic catalog rows are byte-equivalent as parsed JSON; no IDs removed. Only eight coverage rows changed, adding observations or changing four omitted rows to explicitly bounded observations. Existing normative catalog remains unchanged.

## Independent validation

Pinned Node 24.16.0 / pnpm 11.20.0. Frozen dependency install and build passed. Targeted Vitest run passed **269 tests in six files**, with `--maxWorkers=2`:

- `apps/bdptest/src/development-matrix.test.ts`
- `packages/conformance/src/development-read-update.test.ts`
- `packages/conformance/src/read-update-catalog.test.ts`
- `packages/server/src/alias-evaluator.test.ts`
- `packages/server/src/member-executor.test.ts`
- `packages/server/src/member-identity.test.ts`

`git diff --check` passed. Logs and JSON are alongside this report as `alias-review-install.log`, `alias-review-build.log`, `alias-review-tests.log`, and `alias-review-tests.json`.

## Limits

This was a bounded source and targeted-test review, not a full suite or production qualification. Deterministic allocator collision skipping remains untested. Fixed all-visible authorization cannot prove hidden-subject behavior. Retention/expiry, disclosure changes, and internal retained metadata remain outside these new observations. Pinned endpoints do not establish historical revision resolution or external fetching. Helper identity checks observe the emitted URL's canonical serialization and path shape, not an exhaustive allocator grammar proof. The frozen combined successor and its separate Read reseal were untouched.
