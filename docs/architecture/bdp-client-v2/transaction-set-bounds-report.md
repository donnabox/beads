# Set bounds and failure-shape follow-up

Frozen clean commit **064db8a99629122f130a91ee8b5189432ffe97e9**, branch
`codex/vickie-transaction-set-bounds`, checkout
`/Users/dbox/repos/wt/vickie/vickie-transaction-set-bounds`, parent
**8d64fcf7fdaa34642fadb3c2c0b8f4729c7778df**. Origin is
`https://github.com/gastownhall/bdp.git`. No outward writes. Prior sources untouched.

Exactly three files changed: `packages/server/src/transaction-batch.ts`,
`packages/server/src/transaction-set.test.ts`, and
`docs/design/transaction-implementation-progress.md`. No protocol, store, evaluator,
HTTP, configuration, schema, profile, or artifact changes.

## Review dispositions

- **F1:** Replace the mapped Omit branch with an explicit local required Problem
  shape: type, numeric status, typed retry, and typed diagnostics. Both branches
  use it; operation-local index remains required. Shared RFC extension types remain
  unchanged. ReadUpdateProblem does explicitly redeclare type as string (correcting
  my preliminary message); its optional status still requires a narrow local check
  when constructing required receipt Problems. Existing evaluator type/status/retry
  are checked before spreading unchanged values; nothing is defaulted. Missing fields
  throw an internal integrity error, roll back effects, and leave no invalid terminal
  receipt. Three real owner/evaluator malformed-output probes verify missing type,
  status or retry and successful valid failure execution with the same key after
  reopen retracts the interrupted pending claim. Compiler negatives cover each
  omitted required field, invalid retry, and string status. This does not seal RFC
  extension names or provide a comprehensive runtime Problem validator; the owning
  evaluator/parser remains responsible for other field semantics.
- **F2:** Cardinality guard precedes charging the matched budget. A selection of three
  with max=1 and evaluator cap=2 returns operation-local 409/index1 rather than
  transaction-wide 413. A preceding create rolls back, the failed receipt survives
  reopen, a permissive max=3 still gets the cap refusal, and reducing the live set
  enables a fresh key while the original failed key remains retained.
- **F5:** Prose states deletion provenance is inert for both supported collections:
  Beads and unowned Links. Owned-Link transitions remain unsupported in this base.
- **F7:** Permanent real default-bound test seeds 128 Beads, runs 32 no-op sets and
  observes 4096 matches / 4128 result entries / no effects. A 129th Bead produces
  atomic overflow refusal. Both completed and failed dispositions replay after
  durable reopen. No exported constant or raised/lowered cap is used in this test.
- **F4/F6 deferred:** Documentation distinguishes the shared visibility predicate
  from full checkedGraph validation, and records repeated endpoint reads plus
  complete candidate scans. No caching/integrity redesign in this slice.
- **F3 remains a materialization delivery gap:** no byte condition mislabeled as a
  Resource-count limit, and no new failure semantics. Separate design needed.

## Actual receipt-size measurement versus projection

`receipt-size-probe.mjs` uses the emitted owner, ordinary admitted create/update
requests and disposable SQLite. It builds an exactly **262144-byte Resource** by
appending metadata in separate mutations (largest request **60138 bytes**, below
65536). Metadata is governed by representation size; the properties-only bound does
not stop this construction. Thirty-two no-op matches on this single Resource produce
an **8392367-byte receipt**, **262196 bytes** for one indexed Resource result entry,
which is persisted and successfully replayed after reopen. This is measured.

**4096 × 262144 = 1073741824 bytes** for Resource postimages alone is an arithmetic
projection, before result/envelope overhead. No GiB allocation or failing persistence
was attempted. The same installed runtime reports V8 MAX_STRING_LENGTH **536870888**
(code units) and SQLite MAX_LENGTH **1000000000** (bytes). These measured records are
ASCII, so those units coincide for this input family. Already 2048 maximum-sized
postimages exceed the V8 string limit before envelope overhead; the count cap admits
up to twice that. Paged output does not avoid complete receipt JSON construction,
storage and reparsing. The observation proves the present request/Resource/count
bounds do not guarantee deliverable receipts, not that a GiB write was tested.

## Verification

Node **24.16.0**, pnpm **11.20.0**. Each named command has its own `.log` and `.exit`
under this directory; regressions also have machine-readable JSON and exact command.

- Frozen install, final typecheck, build, dependency boundaries, touched-file
  format/lint, and git diff check: PASS. No hook bypass.
- Initial guard red: **1 failed /37 passed**, exact expected409 vs actual413.
- Compiler red: **5 expected compilation errors** for formerly accepted malformed
  required-field shapes.
- First green attempt: **3 malformed-fixture failures /38 passed** because an empty
  propertiesChange is not legal syntax. Fixed fixture to an ordinary add patch;
  no runtime change was required for those failures.
- Corrected set suite: **41 passed**, 849ms total.
- Default-cap mutant (4096→4095 only): **1 expected failure /40 filtered skips**;
  actual boundary receipt was failed instead of completed. Mutation restored before
  final build and regressions.
- Final regressions: **554 passed /14 files**, no skips, **6.03s**. All transaction
  test files plus Resource and alias evaluator tests; includes actual emitted HTTP
  changed-port/page recovery and real hot-journal crash recovery.
- Both `smoke` (packed installed lifecycle) and `smoke:development`: PASS.
- Emitted receipt-size probe and durable replay: PASS. Runtime materialization-limit
  introspection recorded separately in `materialization-limits.log`.

No full repository suite or public profile qualification claimed. No interrupted
runs or remaining test-owned processes; disposable fixture/probe stores cleaned.
HTTP set admission remains closed. Current fixed cap remains4096; candidate scanning,
large receipt materialization, owned-source fanout and broader policy/configuration
work remain outside this repair.
