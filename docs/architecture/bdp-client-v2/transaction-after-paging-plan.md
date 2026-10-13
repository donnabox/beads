# Next Transactional serving slice after durable receipt paging

Read-only proposal, 2026-10-13. Inspected immutable paging head
`0ff180a531b3c9ae594e0a104321e80744de38ed` in
`/Users/dbox/repos/wt/vickie/vickie-transaction-receipt-pages`. Root reports its
ordinary integration into `bf462023`; that integration and the ongoing paging
reviews remain predecessors, not review results asserted by this memo. No source
edits, tests, commits, public claims, or outward writes were made for this plan.

Authority remains the sole fork design's **Autonomous profile completion** T2/T3
plan and its `profile-completion-required-contract-matrix.md`. That matrix is an
inventory at its stated older pin, not a current pass count. This memo is a bounded
implementation proposal, not another plan of record.

## Recommended unit

Implement **updateWhere and deleteWhere through the existing atomic owner and
actual private HTTP singleton/batch routes**, for the owner's current fixed
all-visible, schema-free, unowned installation. Preserve the existing sequence
vocabulary: sets are not sequence members. This completes one useful serving
unit, including selection, effects, retained results, paging, retry and restart.
It does not complete Transactional discovery or qualify an installed profile.

**Boundary requiring review before implementation:** the exact error/admission
phase for grammar-invalid set Selectors and Selector complexity is not settled by
the carrier schema alone. See the dedicated admission section below: spec
4490–4539 versus the nonempty-string selector schema and closed receipt-code
list at 5297–5332. Root will review the normative reading before approving that
boundary; the proposed mapping below is not a ruling.

Paging now removes the result-size obstacle. Both set verbs reuse the same
selection and singleton expansion seam; splitting them would leave almost the
same plumbing unfinished. The first milestone can be preparation/owner tests,
but the delivered unit must include actual HTTP and page-following evidence.

Snapshot/changefeed are the next larger serving family, not a cheap alias for
`inspect()` or retained Read query pages. They need a replay floor, epoch/view
bound checkpoint lifecycle, joint snapshot capture, retained handles, finite
readers and SSE. Completing them together with sets would obscure the smaller
atomic mutation review. Remaining minimum-read combinations still include
unresolved semantics and should not be silently settled as implementation work.

## Gap inventory after paging

| Required family | Current concrete primitive | Remaining work |
| --- | --- | --- |
| Atomic mutations, identity, receipts | One durable owner; direct/sequence admission; original alias witnesses; canonical sequence dependencies; fenced recovery; same-entry observations; durable bounded receipt pages | Set preparation/evaluation/routes; installed contracts and policy integration; applicable variable-view receipt projection; Resource-alias sequence representation |
| Set mutation and expansion | Normative schemas, pure Read Selector, singleton Resource evaluator, complete receipt storage/page plan | Both set verbs, staged canonical selection, cardinality, flat ordered results, finite Event expansion, actual set-produced receipt pages |
| Read serving consistency | Same-owner context, retained query snapshots, visible point and receipt/page complete canonical minimum tuples, genesis | Absent/hidden point ordering, query continuation minima, unsupported/partial tuples and opaque unknown-token disposition; route coverage beyond bounded subset |
| Snapshot | Same-entry read facade and private query pagination | First-class manifest and joint Bead/Link generation, checkpoint/replay rendezvous, immutable typed streams, handle retention/expiry, mandatory erasures array |
| Changefeed | Durable atomic groups with predecessor, final postimages/tombstones, Events, empty erasures | Replay retention/floor, explicit start intent, epoch/view-bound checkpoints, finite pages/SSE, reconnect, projection advance and minimum/HEAD behavior |
| Event Sources | Correct ordered generic facts; Link subject/source separation and endpoint/self-Link facts | Scope/Resource finite and SSE readers, source-local cursors, history expiry, deletion/withheld cursor behavior |
| Configuration-sensitive laws | Fixed all-visible unowned owner; shared Resource evaluator exists; Update installation work is separate | Reviewed installed-contract/config composition; applicable schema/owned/hidden policy paths; erasure and restore applicability/evidence. No new Types or optional History in this slice |
| Cold installed service and qualification | Source-private runtime and canonical route fixture; no incomplete descriptor | Complete truthful Transactional Directory/discovery, installed selection, cumulative actual target-bound observations and evidence. No profile claim from a component test count |

The paging candidate fixes root/page retention ordering and supports bounded
inline results; its large-set catalog witness is still missing until an actual
set creates those entries. Its actual HTTP disconnect waits drain the remaining
finite budget; only the internal point facet has AbortSignal support. Preserve
that distinction in later documentation.

## Settled contracts and source anchors

All line references below are to `docs/specs/bdp.md` at the inspected head.

- Selection, 825–889: bounded JSONPath grammar; canonical complete candidates;
  ordinary literal comparison, with no reinterpretation of local IDs, aliases or
  `@name` inside Selector strings. Reuse `read-selector.ts:selectReadResources`.
- Set mutation, 4215–4285: properties-only update; metadata unchanged; selection
  at this operation's serialization turn, seeing preceding staged effects;
  inclusive cardinality, entire transaction failure outside bounds; provenance
  fans out as the corresponding singleton operations. No expected-member guard.
- Admission, 4490–4539: bounds, whole-carrier syntax, principal, existing key,
  unknown-key capacity, negotiation, durable admission. Equal duplicates do not
  reselect or consume fresh execution capacity. Failures before admission create
  neither key nor receipt. Do not introduce Read+Update key behavior.
- Set wire objects, 4837–4918: exact collections, closed request/member shapes,
  singleton bodies omit discriminator, no set pagination, visible match count,
  matched nonwritable Resource fails rather than being filtered out. Global graph
  constraints still inspect authoritative state. Current installation is all-visible.
- Receipt expansion, 5141–5187: one matched/count entry followed by one update or
  deletion per selected Resource, canonical code-unit ID order, original outer
  operationIndex on every entry. No set name field is admitted by the schemas.
  Zero-match still has one result. Every entry consumes one page slot. Retained
  intermediate postimages differ legitimately from the group's final postimages.
- Failure framing, 5247–5263 and 5287–5332: cardinality-violated is a 409 conflict
  with after-state-change, retained inside a failed receipt; ordinary operation
  failure has operationIndex. A genuinely transaction-wide limit omits it.
  Limit failures after admission use receipt limit-exceeded, never partial success.
- Event expansion, 5578–5594: singleton-equivalent facts in canonical ID order;
  no facts for zero matches or no-op Resources; one transaction identity; all
  facts atomic. A finite per-transaction induced-Event maximum is mandatory.
- Catalog: `transactional.set.mutation`, `.zero-match`, `.cardinality`,
  `.singleton-targets`, `.attribution-fanout`, `transactional.event.set-expansion-order`,
  and `transactional.receipt.pagination`, plus inherited atomicity/replay/recovery.
  Owned-source fanout is not satisfied by unowned-only tests.

## Proposed representation and file scope

1. **Private preparation, not a broader RU operation type.** Add a small
   `packages/server/src/transaction-set.ts` and tests for the two set records.
   Widen the private PreparedTransaction operation union in `transaction-batch.ts`
   from its current `PreparedReadUpdateCarrier`-only representation. Keep original
   branded RU operation objects for existing members; use the existing lossless
   decoder and numeric admission utilities for sets. Preserve whole-batch label
   validation for Resource members and original outer indices when set members
   are interleaved. Do not forge RU parser brands or admit sets into RU/sequence.
   Reuse existing patch/provenance validation; if its inaccessible private seam
   requires a protocol helper extraction, request root scope before editing that
   shared package. No parallel patch interpreter or public parser API by accident.

2. **Identity before evaluation.** The private union carries original lossless
   input through normalization. Selector strings remain exact ordinary strings;
   object key order and mathematically equal number spellings have existing
   semantic equality, arrays retain order, omitted bounds remain absent. Empty
   provenance follows existing normalization. Singleton and one-operation batch
   representations compare equally. Set witness entries have no alias slots;
   preserve strict witness validation rather than treating unknown kinds as sets.
   Shared semantic-identity and RU machinery should not require edits.

3. **Same-owner staged selection.** Add a lifetime-checked collection enumeration
   capability to TransactionMutationContext in `transaction-store.ts`, using the
   current SQLite transaction's rows. Convert to canonical records, project with
   the same operation policy, evaluate the existing Selector, sort with canonical
   code-unit ordering, check cardinality, and freeze that operation's selected IDs
   before applying updates. Never select using withRead(), public paginated GET,
   a later inspect(), or a second owner. Earlier batch operations are visible;
   changing the selection predicate during the expansion does not reselect.

4. **One Resource application/fact path.** Refactor only the private batch loop's
   existing per-Resource execution and fact recording into a shared local helper.
   Sets invoke it within the existing effects savepoint, not store.execute or an
   independently committing loop. updateWhere.change maps to propertiesChange;
   metadata remains untouched. Emit matched first, then each selected outcome with
   the outer operationIndex. Reuse evaluator failures, revision/provenance/no-op
   rules, incident-Link deletion checks, Link endpoint facts and final changes map.
   The existing store retains the failed receipt while rolling back all graph,
   identity, group and page effects. No shared Resource evaluator changes expected. Widen the private BatchEvaluation
   problem type to the actual receipt vocabulary (including cardinality-violated)
   and optional operationIndex for transaction-wide failures; do not weaken the
   shared RU Problem type.

5. **Real expansion bound.** Enforce one finite induced-Event cap across the whole
   transaction, including ordinary Resource operations and set expansions. The
   value is private installation policy, not a new protocol constant. Use a
   detached immutable configuration/constant and deterministic small test bound;
   fail atomically with the existing receipt code/limit. Do not omit facts or stop
   at a page. Do not advertise duration or other optional limits without enforcing
   them. Selection and complete retained result materialization remain linear in
   the selected population in this implementation; no production-scale streaming
   or concurrency guarantee is implied. An additional matched-resource bound is
   legitimate but should be explicit scope, not invented silently during coding.

6. **Actual serving.** Widen `transaction-submission.ts`'s private singleton union,
   and add update-where/delete-where to `transaction-http.ts`'s existing operation
   table so raw method routing and POST dispatch stay one authority. Service/store
   observation and receipt paging remain the only output path. Minimal runtime
   test seam for Event bound only if needed; keep immutable receiving controls,
   trusted credentials, fixed Scope/listener separation and existing conditional
   handling. No Directory/profile advertisement until all required endpoints exist.

Expected edits: transaction-batch/store/submission/http, new transaction-set helper
and tests, focused transaction store/submission/http/receipt-page tests, progress
note, and private module installed-smoke allowlist/denied imports. Minimal service
or runtime changes only for truthful shared bound configuration if necessary.
No authority-read/pagination, Update runtime, shared evaluator, schema, normative
spec, conformance evidence, public export or Types changes. Root owns any shared
parser/configuration extraction and integration with the installed Update factory.

The persisted witness vocabulary gains new kinds. Before enabling them, explicitly
choose compatibility: either prove old-format readers safely refuse new state or
use the existing private-format/policy refusal mechanism. Do not silently open a
new set-containing database in an old owner. Any approved format bump must refuse
old files byte-identically before recovery/maintenance; no migration or reseed.

## Admission issues to settle before calling the negative surface complete

There is one narrow clause-to-confirm question, not an excuse to stop the settled
core: invalid/unsupported Selector grammar and Selector complexity in a set body.
The set schema only requires a nonempty string. The inherited Read implementation
(`packages/server/src/read-request.ts:954–964`) maps syntax/unsupported grammar to invalid-parameter and complexity to
limit-exceeded. However, the closed receipt code list excludes invalid-parameter,
so copying Read error mapping after admission would be wrong.

Recommended reading for root review: grammar-invalid set Selectors fail whole-
carrier syntax before consulting the key, using malformed-request; execution
work bounds follow key comparison and yield the existing permanent failed receipt
where applicable. Confirm the exact mapping/timing for selector.bytes/depth/nodes
against existing rulings before implementing those negative routes. Do not silently
reinterpret receiptProblem or use a generic transient failure for a permanent
Selector error. Valid bounded selectors, cardinality and atomic expansion do not
need a new product choice. The unrelated absent-point/continuation minimum questions
remain out of scope. Cardinality min > max need not become a new malformed rule:
its schema permits both nonnegative bounds and no count satisfies that constraint.

Numeric shape validation must not replace all number literals with zero when
checking cardinality: negative/fractional bounds are invalid syntax, while the
actual literal must survive until normal numeric admission/equality. Do not impose
Number.MAX_SAFE_INTEGER as a new protocol cardinality rule; use existing exact
number semantics, and ensure diagnostics retain actual submitted locations.

## Executable done-when predicates

1. Both singleton targets and mixed explicit/set batches execute through real HTTP
   on one owner, returning schema-valid receipts and same-entry context. No route
   advertises an incomplete profile. Sets remain rejected in sequences before any
   member key or prefix effect.
2. Create/update before a set changes that set's matches; a later set sees earlier
   set effects. At least one test changes its own Selector property and proves
   every originally selected Resource is handled once. Canonical absolute identity
   literals match; local/alias/@ strings are not rewritten.
3. Bead and Link updates/deletions produce matched then per-Resource outcomes in
   canonical code-unit ID order, independent of insertion order. Metadata survives
   property updates; retained no-op revision/provenance is unchanged; real versions
   receive supplied provenance. Zero-match and all-noop transactions have no group
   or effectPosition, including mixed explicit no-op members.
4. Inclusive cardinality boundaries pass; misses fail with the schema-valid retained
   cardinality receipt, correct operationIndex/pointer where locatable, and no
   graph/group/Event/page effects. Retry with the same key returns original failure
   even after population changes; a new key may succeed.
5. A late patch test failure, incident-Link conflict, invalid numeric mutation or
   last-operation failure rolls back every prior selected Resource and earlier
   batch operation. deleteWhere(links) then deleteBead succeeds; the reverse order
   fails atomically. No identity allocation survives a rolled-back creation.
6. Explicit and set updates share exact Resource fact semantics. Set Link deletion
   includes subject Link identity and both endpoint facts, including a self-Link;
   facts have original order, ordinals and transaction identity. Repeated mutation
   or reversal preserves intermediate receipt facts but only final group state.
7. Finite Event cap exact-bound succeeds and bound+1 fails before commit. A mixed
   batch counts all its facts. No partially persisted effects/group/pages; retained
   failure and retry obey existing owner laws. Transaction-wide failure framing
   omits a fabricated operationIndex.
8. One singleton set and equivalent one-op batch share key identity and disposition;
   numeric spelling variants compare correctly, Selector string differences
   conflict, and changing population never changes an equal retained retry.
   Pending presentation joins rather than running a second selection. Existing
   Resource/alias witness and sequence regressions remain intact.
9. Set results cross at least three receipt pages with a small actual shared page
   maximum. Count the matched entry in the cap, follow every next URL, reopen on a
   different listener port with stable Scope, and recover byte-equivalent recorded
   boundaries/postimages. Later updates/deletions do not recompute old results.
   Existing expiry/principal/epoch/minimum ordering remains unchanged.
10. Body/schema/key/media/auth/negotiation and unknown-key capacity tests inspect
    actual owner state, not mocked decisions. Include malformed later set members,
    duplicate decoded names, invalid collection/name/discriminator/cardinality,
    invalid patch pointer, read-only/unauthenticated requests and conflicting keys.
    Resolve the Selector negative mapping above before marking these complete.
11. Inject after-effects, after-terminal and uncertain-COMMIT faults into actual
    set executions. Preserve sticky fencing, cause chains, restart recovery and
    recorded page plans; no claim of disk/power-loss testing. If compatibility
    changes, prove byte-identical refusal before maintenance and valid same-version
    reopen/pending recovery.
12. Pinned Node24.16/pnpm11.20 build first, focused set plus all Transactional owner,
    admission, paging and runtime regressions; selector/Resource evaluator/Read
    regressions where touched; typecheck/format/lint/boundaries; both installed
    smokes; exact unchanged Read evidence validation. Record real counts and gaps,
    with no reseal or profile admission. Root performs independent review and the
    integrated broader suite before publishing.

## Execution order after authorization

1. Start a new isolated successor from root's reviewed paging integration and
   required review repairs; leave 0ff and review checkouts immutable.
2. Land private preparation/identity and the staged collection seam with owner
   tests, keeping HTTP routes closed until atomic and failure tests pass.
3. Add shared expansion/fact recording, finite cap, full result/page/recovery tests.
4. Wire the two existing normative singleton targets and mixed batches, run real
   HTTP/restart/negative matrix and package boundaries, freeze/report exact head.
5. Root reviews/integrates. Separately scope installed-contract composition, then
   Snapshot/changefeed retention/checkpoint serving. Do not fold unresolved alias,
   conditional policy, optional History or new Type work into this unit.
