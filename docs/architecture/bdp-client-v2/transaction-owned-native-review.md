# Independent native review: owned Transactional transitions

**Bounded pass; no blocking correctness finding.** Frozen source
`01e2d5470941a49240fab7820a6dea1b6deb291e` versus
`71549f2610d61e6562b2c703d96e4c25d1093492`, author checkout
`vickie-transaction-owned`. Independent checkout `bdp-transaction-owned-native`.
No source edits, outward writes or external implementation findings consulted.
Read the accepted owned-transitions plan, author report, normative Event/owned
clauses and actual implementation/tests.

## Ancestry and scope

Reviewed27bad2badc765fc02aeec9456e1fc5f9dd49ea5c is an ancestor through ordinary
merged3008c7, not a rewritten dependency. The incoming transaction-set.test.ts is
byte-identical to27bad2; closed receipt Problem codes/statuses and cardinality-before-
matched-cap remain in batch.ts. These are inherited reviewed changes, not attributed
to owned authoring. Owned runtime changes are batch projection plus removal of the
owned installation guard, not shared evaluator, HTTP, service, identity or schema work.

## Correctness assessment

- **Same-turn source state:** applyResource's synchronous putResource wrapper reads
  and parses each prior row before forwarding its first write, local to one evaluator
  invocation. It does not expose the evaluator's recipe, allocate another revision,
  read later final owner state or call the evaluator twice. Source identity/kind/new
  revision are checked against the outcome before emitting its Event. Captured full
  preimages remain local; only previousRevision is published. Error paths do not
  attach preimage bodies.
- **Final versus intermediate facts:** evaluator changed/deleted paths are local
  lookup keys; change Map keys are canonical record URLs. Repeated changes overwrite
  only final postimages/tombstones. Deleted direct subject's pre-evaluation identity
  is checked against outcome.deleted; no deleted-row postimage lookup is attempted.
  Each operation still contributes its immutable intermediate receipt result/Event
  revision chain, including reversal and direct source updates interleaved with
  two owned Links. No collapse based on equal final properties.
- **Event contract:** spec5437–5508 requires Link lifecycle first, graph incidences
  source then target, source updated last; implementation follows that order and uses
  the same cap/ordinal emitter. Created nested Link is its complete new record;
  updated is constructed field-by-field with Link revision pair, actual admitted
  patches and new provenance, excluding whole properties; deleted is identity only.
  Source data carries its own revision pair/provenance plus ownedLink, without direct
  change/metadataChange. Attribution removal is not replaced by prior attribution.
  Self-Link yields two endpoint facts and one source transition. No-op exits before
  changes/Events while retaining the evaluator's source/sourceRevision outcome.
- **Atomic ownership and limits:** graph/effects/receipt/page plan/group remain in the
  existing owner transaction. An extra owned source fact can fail the existing total
  inducedEvents bound and rolls back both Resources. The actual4096 test uses1024
  small self-Link deletes (4096facts/2048changes);1025 fails transaction-wide and
  replays after restart. This is real default count evidence, not a smaller override
  mislabeled as4096 or a byte/high-fanout performance claim.
- **Closure and authorization:** no new ownership/visibility implementation is added.
  Existing evaluator source-write check and isResourceVisible transitive/cycle closure
  remain; selected readable but unwritable targets fail rather than disappear from
  cardinality. Policy-spy tests are correctly disclosed as interface evidence.
  Actual owner stays fixed all-visible; no retained mutable-grant projection claim.
- **Installation/recovery:** remaining missing-compiler/brand/module/Scope mismatch
  guards still refuse before provisioning. Complete compiled owning installation is
  moved to its own positive fixture/file and closed. Format8 unchanged is justified:
  prior accepted installations could not contain owning descriptors; immutable exact
  descriptor/capture identity still gates reopen. No old-store reinterpretation is
  introduced. Actual emitted HTTP checks compiled inherited-schema rejection after
  an earlier staged owned update, identical beforestate, later state change and
  retained failed replay across reopen. Child SIGKILL controls cover after-effects
  and after-terminal before transaction commit, not real power loss.

## Independent verification

Pinned Node24.16.0/pnpm11.20.0, frozen-lockfile install, completed build before emitted
HTTP/process tests and independent typecheck: PASS.

Focused suites: transaction-owned, transaction-store, transaction-set,
transaction-installed-contracts. **213tests/4files PASS in4.05seconds**, maxWorkers2.
All15 new owned cases passed, including actual compiled HTTP, two SIGKILL positions
and real4096 boundary (2.149seconds). Logs/JSON/exit0 under
`bdp-integration-logs/transaction-owned-native/`. Author source remains clean.

No duplicate full workspace run, packed installed cohort, Event-serving test or
public qualification performed. Remaining byte-materialization and mutatedResources
accounting, mutable grants/erasure, HTTP set admission, Snapshot/changefeed/Event
serving and broader minimum controls remain explicitly outside this acceptance.
The author's stopped high-fanout attempt remains inconclusive; this successful
small-per-source workload neither diagnoses nor resolves its cost.
