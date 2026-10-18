# Native review — private set core integrated with recovery

Reviewed frozen `fccb951ce0a470526a22db9762236935f78ec5e8` relative to recovery
`270a1216cb918a4194f3b195db6c8fcbecf0c040`. Author checkout:
`/Users/dbox/repos/wt/vickie/vickie-transaction-set-integration`.
Independent build/review clone: `/Users/dbox/repos/wt/vickie/bdp-set-native-review`.
Both source trees are clean; author HEAD was verified at entry and completion.

**Disposition: no actionable finding in the bounded private valid-set core.**
This is blocker-closure/implementation review, not public Transactional profile
qualification. No source edits, commits, network publishing, or shared databases.

## Review conclusions

- **Shared Resource path:** `transaction-batch.ts:515–611` extracts the existing
  mutation/result/fact handling with the intended no-op `continue` → helper return
  adjustment. Ordinary numeric admission, local-reference substitution, admission-time
  alias witnesses, and alias-only singleton path stay outside/in front of that helper.
  A real emitted-owner probe exercised all six Resource operations, checked exact
  normalized identity equality between old and widened preparation, and replayed
  the old request through widened preparation without executing it again.
- **Full outer carrier:** `transaction-batch.ts:118–157` uses a Resource-only projection
  solely for existing syntax validation, then restores the full ordered operation
  union. No filtered count reaches owner admission. Normalization, creator indices,
  operation indices and witness arrays traverse that outer union; set witness slots
  have no alias fields. Tests prove a set before an allocated creator, renamed labels,
  a set before an alias operand, alias repointing between admission and execution,
  retained original target, original outer result index on a page, and canonical retry
  equivalence after reopen. Operation-count refusal tests inspect empty durable receipt
  storage and correct/reuse the original key.
- **Lossless cardinality and identity:** `transaction-set.ts:28–65` checks original
  number-literal sign/coefficient/fraction/exponent before the RU syntax placeholder
  path. It accepts mathematical zero and exact nonnegative integers without adding a
  safe-integer cap. Original literals remain identity inputs and numeric-admission
  inputs. Both canonical bundle operation roots are cross-checked; tests include tiny
  negative/positive fractions and equivalent exponent spellings. Schema projection
  cannot silently turn these bounds into zero. Patch diagnostics retain `/change`.
- **Private preparation immutability:** the public-text-to-private-prepared path uses
  the deep-frozen lossless decoder, frozen operation array and branded prepared wrapper.
  The independent emitted probe directly attempted to replace a nested patch value
  and append an operation: both threw. The syntax-only set placeholder never enters
  durable identity or execution. HTTP still invokes ordinary preparation, which refuses
  sets; the new core does not settle negative Selector wire semantics.
- **Serialization and frozen selection:** `transaction-batch.ts:648–702` reads the staged
  full collection, shapes canonical Resources, filters visibility before selection,
  captures/sorts the resulting array once, checks cardinality, emits one matched entry,
  then expands in canonical ID order through the ordinary evaluator. Later set operations
  read the updated staged state. Independent probe admits a set, commits another matching
  Resource before execution, and proves both are selected; selection is not frozen at
  admission. Existing tests cover preceding staged writes, selection-changing patches,
  zero matches, no-ops, metadata preservation and unowned provenance.
- **Atomic facts and recovery:** the new store enumeration at `transaction-store.ts:903`
  shares the same lifetime-checked mutation context and SQLite savepoint as ordinary
  mutations. Rows/array are frozen and capability use after the turn fails. Failed
  cardinality, late patch failure, induced-Event refusal and injected terminal failures
  retract graph changes; permanent failures retain their disposition. Set pages retain
  original postimages through deletion/reopen; uncertain COMMIT retains its receipt/page
  chain. The focused run also reran both real process-death hot-journal tests.
- **Event bound:** the bound is one total across ordinary and set expansions, enforced
  at fact append (`transaction-batch.ts:498`), with rollback on the private sentinel.
  The evaluator override can only lower the real 4096 bound. No-op results consume no
  Event slot; self-Link source/target facts consume separate slots. The real-cap probe
  seeds 1366 self-Links: deleting all would create 4098 Events, so it fails atomically,
  retaining the failed receipt across reopen. Deleting 1365 Links plus one ordinary Bead
  update succeeds with exactly 4096 Events. No lowered limit or mocked evaluator was
  used for this probe. Failure correctly omits a fabricated operationIndex.
- **Recovery integration:** relative to 270a, the only store changes are format
  `bdp-private-transaction-set-core-7` and staged enumeration. The hot-journal repair,
  compatibility-before-BDP-maintenance ordering, receipt ordering and owner fencing
  remain intact. Format rejection is intentional rather than an implicit migration.

## External plan-review dispositions

Read `transaction-after-paging-claude-plan-review.txt` and the author disposition.
F1/F5 are resolved by the actual full outer carrier rather than redundant counts or
index maps. F4 has exact-literal validation plus bundle cross-checks. F6's format bump
is present; its old read-only-preflight recommendation is superseded by the reviewed
real hot-journal recovery defect. F10 is a private evaluator-only lowering seam, not
an installation setting; actual cap behavior was independently exercised. F9 uses
canonical bundle roots for validation. F2 remains a normative decision; F3's proposed
all-errors-to-400 shortcut was deliberately not adopted. F8/F11 remain route-activation
follow-ups. F7's owned-source and selected-but-nonwritable branches remain explicitly
unqualified under this fixed unowned/all-visible/all-writable owner.

## Independent executable evidence

Pinned Node **24.16.0**, pnpm **11.20.0**. Frozen install and build completed before
emitted/process tests. Full logs/scripts are in this directory's `native/` child.

- PASS build (`build.log`).
- PASS **424 tests / 7 files**, 3.78 seconds (`tests.log`, `tests.json`):
  read-selector 63; resource-evaluator 167; transaction-dependencies 25;
  transaction-receipt-pages 12; transaction-recovery 2; transaction-set 23;
  transaction-store 132. The command additionally named a nonexistent
  `transaction-batch.test.ts`; Vitest selected the seven existing files above, and no
  separate batch-test file is claimed as evidence.
- PASS four independent emitted-owner observations (`probe.mjs`, `probe.log`): Resource
  path identity/replay, preparation immutability plus selection-time visibility, real
  4098-Event rollback/restart replay, real 4096-Event success.
- PASS typecheck, dependency boundaries, scoped format and lint (`checks.log`), and
  `git diff --check`.

Two initial reviewer-probe assertions were corrected before the final clean pass:
Node deep equality distinguished in-memory null-prototype JSON objects from parsed
receipt JSON (comparison now uses their JSON values), and inspect() returns canonical
Resources rather than storage rows with `kind` (Link count now uses its actual Type).
These were reviewer-oracle mistakes, not runtime failures. The first correction does
not assert raw response byte equality. Each attempt used finally cleanup.

No interrupted test runs or test-owned processes remain. Disposable probe stores were
removed. No author source was modified. Did not repeat the author's full 545-test run,
packed/development smokes, 89-row Read evidence verification or old-writer byte-equality
probe; those remain author evidence, not independent results here.

## Boundaries retained

No public export/profile/discovery/admission activation is claimed. Negative Selector
syntax/complexity wire ordering, HTTP set targets, owned/schema/policy integration,
other optional transaction limits, replication and whole-profile conformance remain
outside this slice. Valid fixed-bound private selectors are the supported input set.
Process-death tests establish the exercised SQLite recovery cases, not arbitrary
power-loss or failover durability. No speculative product rule was used as current law.
