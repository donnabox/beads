# Independent atomic store repair review — integration held

Reviewed frozen `d0020005ff01d6b444a916b83adf5fd4e376bb77` against published `3a8f9d553a2944bc7d1d7812a284f0160dcbe798` in isolated clean detached checkout `/Users/dbox/repos/wt/vickie/bdp-transaction-atomic-review`. No source edits, commits, merge or outward actions.

## Required bounded repair

P2: Subsequent fenced entries lose the original failure after compensating cleanup fails. `transaction-store.ts:164` initializes ownerFailure with ??=. When cleanup's atomic() fails, it first fences the cleanup error. `cleanupThenThrow` then assembles original + cleanup at208 and calls fence(combined) at209, but ??= ignores the richer cause. The immediate thrown AggregateError contains both; the persistent fence thrown by later entries contains only cleanup. This violates the stated sticky diagnostic/cause-preservation intent, though the connection remains safely fenced and this is not an observed data-corruption path.

External repro: `/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-atomic-cause-repro.mjs`, output transaction-atomic-cause-repro.log. It imports the exact built store, injects an after-admission application error followed by a BEGIN failure during compensation, then inspects the same owner. Immediate message traversal includes original application cause and cleanup SQLite cause; subsequent traversal includes only cleanup SQLite cause. Prototype injection is restored in finally; store closed and disposable directory removed.

Recommended fix: retain a new fence wrapper whose cause is the combined aggregate, and add subsequent-entry cause traversal assertions for admission, execution and sequence cleanup paths. Do not mutate the earlier wrapper's cause in place: the aggregate already references that wrapper and such a mutation would create a cycle. Root accepted this finding and requested author repair/re-review before integration.

## Other inspected behavior

Before/after-COMMIT faults retain safe uncertainty: guarded rollback only when SQLite reports an active transaction, sticky fencing, no compensating retraction through that fence. Reopen determines the durable receipt/effects/group outcome; the after-COMMIT case replays the completed receipt rather than executes again. Failed rollback and engine-ended transactions preserve the immediate causes and fence. Real SQLite constraint rollback remains retryable, as do existing ordinary transient/application paths.

Inspection/alias reads, receipts, maintenance, admission, execution and retraction enter the fenced owner. Active sequence stepping/cancellation also reach the guard. Completed/already-cancelled cancelSequence can return zero before enter (:624); this only reports an already-complete local no-op and does not read SQLite, admit or mutate work. It is not a state-access bypass. Startup and maintenance use the same atomic helper; background maintenance's existing catch stops the timer and closes the connection. No broader real storage-fault or close-failure qualification is established here. The new prototype injections restore their patches in finally.

## Independent evidence

Frozen install and build passed on pinned Node24.16.0/pnpm11.20.0. Then 361 tests across transaction-store (including all 10 new fault cases), transaction-sequence, recovery-store, member-executor and sequence-executor passed in4.62 seconds. Verbose/JSON/exit outputs: transaction-atomic-review-tests.*. Despite green suite, the external sticky-cause repro remains a confirmed gap.

Per root instruction, no integration/check/smoke expansion after the finding. Await corrected frozen head and independent re-review. No physical disk, power-loss, filesystem-fault, full-profile or HTTP qualification claim.
