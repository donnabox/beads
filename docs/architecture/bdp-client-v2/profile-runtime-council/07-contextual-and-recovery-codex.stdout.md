# Native review: contextual diagnostics and transaction failure handling

Exact head: `486ad62d7c6ecb8e36dad4e596b1597fde425cdd`; base: `3a8f9d553a2944bc7d1d7812a284f0160dcbe798`.

## Finding

**Low — later sequence cancellation fence loses the original turn exception.** `packages/server/src/transaction-store.ts:786–791` aggregates the turn failure with `cancelSequence` failure for the immediate caller, but leaves the stored owner fence containing only the cancellation failure. The earlier cleanup helper correctly replaces the fence with a new acyclic wrapper; this separate catch does not use it. A two-member sequence, ordinary first-turn `after-effects` failure, successful per-claim retraction, then failed cancellation `BEGIN IMMEDIATE` reproduces it. Initial error traversal contains both causes; later `inspect()` contains only cancellation. Data access remains fenced: this is diagnostic loss, not corruption or a proven durability violation.

External disposable reproduction, pinned Node 24.16.0: `07-sequence-cancel-cause-repro.mjs`, output `07-sequence-cancel-cause-repro.stdout.txt`. It imports the candidate's built modules, allocates a fresh temporary database, restores the injected prototype hook and removes its database. Exit 0. Parent has assigned the narrow follow-up. Reuse the acyclic cleanup wrapper and assert later-entry traversal after sequence cancellation failure.

## Checked without additional findings

Read the full six-file delta, the complete ten added storage failure tests, both added scenario bodies and all 80 new assertions, coverage mappings and cited clauses; checked complete failure/cancellation control paths in current store source. The review packet index records additional exact source ranges. Initial tool output containing formatted scenario JSON was truncated; both scenarios were reread in compact form in full before this report.

The two diagnostic plans distinguish malformed local binding (whole carrier rejects, no prefix, same keys reusable after correction) from wrong-kind durable subject (member fails, independent tail commits, retry retains selected result fields, changed semantic request conflicts, fresh key succeeds). A no-op guarded by the initially captured Link revision witnesses unchanged revision. These assertions support the bounded claims; they do not claim all contextual combinations or whole-response replay equivalence. Recounted current inventory: 27 plans, 280 actions, 907 assertions; coverage 55 bounded observations / 26 omitted out of 81.

Storage changes fence uncertain COMMIT outcomes and unproven rollback, preserve ordinary clean rollback retry, skip unsafe compensation after fencing, and replace wrappers to avoid recursive cause cycles in the repaired direct/admission cleanup paths. New tests cover before/after COMMIT, rollback failure, engine rollback, direct and sequence admission cleanup, and an actual SQLite constraint. The distinct sequence-cancellation path above remains.

## Limits

No broad test suite was rerun by this council; parent's 375-test/check/smoke/evidence validation is separate evidence. Fault injection is not live filesystem failure or power-loss qualification. Public Transactional ownership/deadlines and wire admission are outside this packet. No source/evidence edits or outward writes were made. No readiness or entire-branch coverage claim.
