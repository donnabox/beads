# Independent dependency integration review

Frozen head: `c7a16fe70160d3961ecd70c5294eb525b5cae7a8`; parents `486ad62d7c6ecb8e36dad4e596b1597fde425cdd` and `0fdaa7b50f926aef3e6f9725deea368e63ecf45b`.
Isolated checkout: `/Users/dbox/repos/wt/vickie/bdp-transaction-dependencies-integration-review`.

No new integration blocker found. The sole sequence-admission conflict retains cleanupThenThrow and unconditional in-memory claim disposal, with exact-owner reservation deletions and attempt-scoped creator_facts deletion inside one compensating atomic transaction. Existing uncertain-commit fencing prevents speculative cleanup. Creator facts join successful creator effects/receipt/group in the same atomic transaction; startup retracts pending reservations and facts together. Private dependencies-4 / transaction-dependencies-3 configuration refusal remains before recovery. Reviewed fencing wrapper and owner entry guards survive the merge.

Independent validation: frozen install/build, 384 tests in six files (transaction-store, transaction-sequence, transaction-dependencies, recovery-store, member-executor, sequence-executor), typecheck, boundaries, format, diff check and unchanged Read evidence verification all passed. Tests include before/after COMMIT injection, rollback and cleanup faults, later-access causes, retained/expired creator identity, failed creator forgetting/reuse, pending overlap, abrupt restart and byte-identical prior-format refusal. Duration 5.26 seconds. Logs: `transaction-dependencies-independent-*` alongside this report. No source edits or outward writes.

Known pre-existing nextSequence cancellation cause-loss path remains in this exact head and is being repaired separately at 05189ef; this is not a c7 regression. No full suite, physical storage-fault qualification or full-profile admission claim. Both smokes will be rerun on the authorized combined candidate.
