# Private transaction materialization budget — revised proposal

Plan only. No production budget, admission rule, schema, public advertisement, or
source file changed. Original proposal is preserved byte-for-byte as
`transaction-materialization-plan.original.md` (SHA-256
`97f07f2bc25191dcf2c6d9ff5640161d5e4d2a53c698f29423a6aa5e57919ac8`).
This revision uses actual owned implementation **01e2d5470941a49240fab7820a6dea1b6deb291e**
in `vickie-transaction-owned`, coordinated with its author; it does not assume the
older prospective Event shape. Root must accept this revision before implementation.

## Recommendation

Use the existing private count vocabulary with **M=256 cumulative matched Resources
and E=256 induced Events**, retaining O=32 explicit operations, Q=65536 raw request
bytes and R=262144 complete Resource representation bytes. The two independent
counts bound different artifacts: no-op result volume needs M; change-group fanout
needs E. No byte refusal is relabeled as a Resource count.

Under the explicit prerequisites below, conservative whole JSON bounds are:

- completed receipt: **173,346,816 bytes**;
- effective group: **305,397,760 bytes**;
- actual pinned V8 string ceiling: **536,870,888 code units**;
- SQLite MAX_LENGTH: **1,000,000,000 bytes**.

The group has about 43% headroom below the first string ceiling. These are bounds,
not desirable target allocations and not a heap-safety proof. The recommendation is
conditional on bounded real workload measurements. M32/E4096 has a tighter corrected
group bound of169,345,024 and receipt39,391,232, but unnecessarily rejects tiny/no-op
sets above32. M256/E128 is a useful fallback (group154,402,816) if measurements reject
E256: it preserves no-op capacity but halves Event-heavy work. Do not choose the
highest count that barely fits a string ceiling.

M256/E256 includes every count combination accepted under the previous proposed
M32/E4096: at most M+O=64 invocations, each at most4 Events, gives at most256 Events.
This argument overcounts set carrier entries, which only makes it safer. Thus this
revision preserves that proposal's worst-case owned atomic-work envelope while
allowing up to8x the matches in less Event-intensive workloads. It does NOT preserve
all workloads of the current4096/4096 implementation; lowering either cap is an
intentional private reference-implementation capacity change.

Retain the original private installation condition **S<=65536**, where S is encoded
JSON length of Scope+`events/`. This is a chosen implementation condition, not an
inferred BDP URL limit. A4096 Scope cap is unnecessary to establish the recommended
bound, and the compiler's4096 diagnostic-Type-ID bound is not authority for limiting
Scope URLs. The alternative is quantified below, without recommending a second
unneeded narrowing.

## Source basis and closed scope

Exact line references below are against owned01e2d547:

- `transaction-batch.ts:54–68`: current4096/4096 counts,32 operations,Q/properties/
  diagnostic limits. `:568–585`: Event capacity checked before push; fixed envelope.
- `transaction-batch.ts:626–653`: evaluator changes/deletions become final Resource
  images or identity tombstones in a Map; repeated versions deduplicate only here.
  `:655–698`: direct lifecycle and local endpoint facts. `:700–740`: owned source
  Event, with exact create/update/delete variants.
- `transaction-batch.ts:779–838`: materialize entire candidate collection, visibility
  filter, selector, sort, cardinality, cumulative match charge, then member work.
  Sets forward properties patches only; explicit operations may also patch metadata.
- `transaction-store.ts:210–218`: exact installation configuration; `:457–464`
  compatibility comparison; `:923–930` safe-integer position allocation;
  `:1024–1029` actual evaluator limits; `:1043–1101` receipt/group persistence.
- `resource-evaluator.ts:97–108`: representation/properties bounds currently OPTIONAL.
  `:1048–1059` checks complete subject and source body JSON before writes.
  `:1077–1101` result carries direct record/tombstone plus source ID/revision, never
  a second source body. `:1145–1162` source body embeds owned Link records.
- `transaction-store.ts:1549–1565` inspect() materializes all resources/groups.
  `recovery-store.ts:189–204,580–598` is a different RU schema/format/owner.

Only stores newly provisioned under, or exactly reopened with, the accepted fixed
budget tuple are covered. All Resource rows must have been written through this
owner/evaluator under R; no arbitrary database injection, old-budget records or
cross-store resource import are included. The owner has no Resource seed/import
option. Fixed UUID revisions/transactions, bounded key/position/time encodings and
current closed outcome shapes are also prerequisites.

## Normalize patch bytes explicitly

Lengths here are compact JSON UTF-8 bytes. This bounds UTF-16 code units of the
well-formed encoded JSON too. Let **N=7Q=458752** bound normalization of request
JSON values. A raw4-byte `1e20` becomes21 output bytes; Q is not a normalized patch
bound. Finite JSON.stringify number tokens have at most25 characters. Tokens of at
least4 input bytes expand by at most25/4<7; length1–2 do not expand; length3 has at
most10 output characters. Strings/keys of accepted well-formed JSON do not grow
relative to their valid original JSON encoding; whitespace removal helps.

Both patch lists of one invocation together are subsets of that one carrier and
therefore <=N **combined**, not N for each list. The normalized carrier partitions
into these two array values plus other fields; the sum of the two array serializations
cannot exceed its <=7Q encoding. Array/member names, separating commas and any
wrapper duplication are paid by A. For an owned update the two Events together
therefore charge at most2N, never4N. The source Event outer provenance is charged
to the source R and its nested Link provenance to the separate Link R. Charging N
once per Event is deliberately loose. A set repeats its
one properties patch across matches, so a carrier-wide N must not be charged only
once to the entire transaction. An owned update duplicates the admitted patch in
the Link Event and source Event. A patch may add a large temporary value and remove
it before the final record, so its bytes are not bounded by the postimage R.

The review's E*(2R+S) Event formula does not account for these patch bytes; it is
not an established bound. This revision does not claim a constructed real input
violates that formula; it replaces its unsupported premise with an explicit N term.

The revised probe uses the actual emitted protocol decoder/admission/encoder.
Escaped lone surrogates are **rejected** by that decoder, rather than being admitted
as the review's JSON.parse experiment assumed; both high/low cases are recorded.
Accepted pair/exponent cases satisfy the7x bound. This is supplementary evidence;
the token argument, not sampling alone, establishes N.

## Exact Event and invocation partitions

Let A=8192 fixed structural bytes per invocation and H=1048576 per document.
A covers at most4 closed Event envelopes,2 change wrappers,result punctuation and
fixed revision/transaction/time/position fields. H covers document envelope,
fixed receipt URI suffix/page envelope and bounded failure diagnostics. Variable
Scope, Resource fields, patches, operation names and diagnostic bytes are accounted
for explicitly; neither allowance licenses arbitrary extensions.

An Event's subject/subjectType and ordinary lifecycle data draw disjoint fields
from the same serialized Resource: id/type/properties/metadata/revision/provenance/
endpoints. They are not separately charged R each. Fixed renamed keys belong to A.
The incidence Event duplicates Link id/type in its data and envelope, so <=2R.
The source Event combines source identity/provenance fields with Link fields; <=2R
conservatively, even though creation's Link is also embedded in the source body.
Previous bodies never appear: only previousRevision/sourceRevision are serialized.
Adding a preimage body to ANY Event/change/result invalidates this proof and requires
an explicit new calculation. Capturing a prior row for evaluation is not permission
to serialize it.

Per-Event variable contributions (plus S per Event and fixed overhead in A):

| Event | Resource/patch contribution |
|---|---:|
| Direct created | R |
| Direct deleted | R (identity subset of old live record) |
| Direct updated | R + N |
| Endpoint linked/unlinked | 2R |
| Owning source updated for create/delete | 2R |
| Owning source updated for update | 2R + N |

To include change records rigorously, partition by invocation. Let t be the number
of local endpoint facts; t=0..2 for unowned Links, **t=1..2 for owned Links** because
the source must be local and live. Self-Links retain both endpoint facts (t=2).
No-op and alias work produce zero Events/changes, but retain result entries.

| Effective invocation | Event count e | Events + assigned changes, excluding A |
|---|---:|---:|
| Bead create/delete | 1 | 2R + S |
| Unowned Resource update | 1 | 2R + N + S |
| Unowned Link create/delete | 1+t | (2+2t)R + (1+t)S |
| Owned Link create/delete | 2+t | (5+2t)R + (2+t)S |
| Owned Link update | 2 | 5R + 2N + 2S |

The first four rows also conservatively cover tombstones and omitted remote
incidence facts. Every row is <= **e*(2.5R+N+S)**. For owned create/delete, dividing
(5+2t) by(2+t) gives at most7/3<2.5; owned update gives exactly2.5R+N before S.
At most2 changed/deleted identities per invocation remain a load-bearing invariant;
there is no delete cascade. The final Map can only reduce the sum of assigned
change bytes. This uses exact shapes, not merely the looser #changes<=#Events fact.

Let K=M+O (overcounts set operation headers as extra invocations). Then:

```
Event-limited group <= E*(2.5R+N+S) + K*A + H
Invocation-limited group <= K*(max(9R+4S, 5R+2N+2S)+A) + H
Group <= min(those two bounds)
Receipt <= K*(2R+Q+A) + O*128 + S + H
```

Receipt derivation: direct Resource/tombstone <=R; source identity/revision or alias
extra fields fit another R; operationName <=Q; O matched-summary entries fit128
bytes each. Repeated no-ops carry full records and are counted, despite no Events.
Aliases currently occur only in singleton/sequence members, not batches; their
locator/target strings fit this overbound. Failed receipts have no partial results,
diagnostics<=524288 plus name<=Q and Scope<=S, inside the separate H allowance with
headroom. Expired details keep smaller allocation identity lists. None of these
claims treats an unbounded trusted principal as part of a bounded JSON outcome.

## Recomputed alternatives

Use original S=Q=65536. All values below are whole compact JSON byte upper bounds,
not measurements. Existing32 operations remain fixed in every alternative.

| M | E | Group bound | Receipt bound |
|---:|---:|---:|---:|
|32|4096|169,345,024|39,391,232|
|64|256|253,493,248|58,527,744|
|128|256|304,349,184|96,800,768|
|256|128|154,402,816|173,346,816|
|256|256|305,397,760|173,346,816|
|256|384|456,392,704|173,346,816|
|256|512|607,387,648|173,346,816|

E512 is not established safe by this proof; the upper bound exceeding a ceiling is
not proof that a real reachable group fails. E384 has inadequate comfortable
headroom for this fixture. At S4096, E256/M256 becomes289,669,120 for groups and
173,285,376 for receipts: it helps but does not justify copying the review's
205,783,040 result, which omitted patches. Exact arithmetic and every partition
inequality are executable in `transaction-materialization-revised-proof.mjs`.

Sequences still contain at most32 singleton operations, never nested set batches;
they do not multiply a256-match receipt by32. Receipt pages contain subsets of a
previously bounded complete receipt plus bounded next URI. Whole histories or pages
that aggregate multiple groups need their own bound; this is not their qualification.

## Required compiler and installation prerequisites

A serialized tuple without enforcement is insufficient. The actual evaluator
currently accepts omission of propertiesBytes/representationBytes. Proposed scope
must close this before relying on R:

1. Make those two fields required finite positive bounds at the evaluation boundary;
   use the single immutable private tuple in the actual TX owner call site and its
   identity. Compiler-negative tests must fail when either field is omitted, and
   real subject/source overflow tests must fail without effects. Diagnostic and
   patch limits retain their existing semantics. Updating shared evaluator typings
   requires a bounded call-site/test audit and root/owner coordination; no silent
   unbounded fallback. A narrower TX-only required subtype is an alternative if
   shared optionality is intentionally retained, but then claims must stay strictly
   about that typed TX owner and include proof every TX construction uses it.
2. Tuple contains M,E,O,Q,properties cap,R,diagnostic bounds and the chosen Scope
   condition. Freeze, consume and persist it in existing exact configuration, rather
   than duplicating numeric literals. Scope check before provisioning; persisted
   tuple comparison before any BDP maintenance. No new public switch or limit field.
3. Old missing/different tuple refuses compatibility; no migration or historical
   receipt rewrite. Current public metadata does not advertise M/E, but private
   configuration/qualification identity necessarily changes. Preserve old artifacts.
4. Equal retained keys are compared/replayed before fresh capacity admission; never
   recompute them under a lowered count. Conflict remains conflict. Count failure
   remains a retained failed receipt until existing forgetting; fresh keys follow
   the new installation. Do not place operational caps in semantic request identity.
5. SQLite hot-journal recovery must occur before reading configuration; do NOT move
   tuple refusal ahead of SQLite recovery. The required boundary is before BDP
   pending retraction/creator cleanup/retention sweep, preserving the reviewed
   transaction recovery repair and actual refusal/valid-reopen witnesses.

`recovery-store.ts` is explicitly OUT of the TX tuple scope: it is the distinct
Read+Update owner/schema (key_state, installed_types, policy, format3), not a loader
or reseeder of TX receipts/groups. TX owns its separate resources/receipts/groups
schema/configuration and has no RU import route. Wrong-owner files must continue
to refuse; an actual future migration/import seam would need the same R/retention
proof before it exists. No claim about all RU stores follows from this TX plan.

The output-string proof is also not a complete SQLite-row proof. Raw trusted
principal strings are presently nonempty but unbounded and stored beside receipt
and witnesses. Identity/witness/configuration fields need their own bounds before
claiming every row fits MAX_LENGTH. Existing fixed fixture principals are small.
Do not invent an IAM policy or conceal this distinction. Private principal bounds
are a separate small choice if root wants the stronger full-row guarantee.

## Capacity, failure and memory consequences

The two counts are cumulative: repeated matches/no-ops consume M, each actual fact
consumes E. Both failures use existing413/limit-exceeded with transaction.matchedResources
or transaction.inducedEvents, no fabricated operationIndex. Cardinality runs first
for the current set. For20+20 matches M32 fails after the second selection, whereas
M256 does not; if the supplied max100 is satisfied it does not suppress cumulative
capacity failure. The absence of a failing operation index makes smaller limits
harder to diagnose. Do NOT add a pre-admission cardinality.max<=M restriction: a
max100 request matching5 is valid today; rejecting it changes existing behavior.

At E256, up to256 ordinary one-Event updates,128 owned two-Event updates, or64 owned
four-Event create/deletes fit the Event count when no other work consumes it; the
32-operation limit still applies, so larger totals require supported sets. Actual
remaining capacity depends on other operations and cardinality. Nonmutating work
can reach M256 without Events. Atomic chunking remains the caller's decision and
loses cross-chunk atomicity; never silently split and claim one transaction.

No count bounds total authority memory. Before either guard, selection currently
materializes `resources(kind)`, JSON-parses every Resource, filters transitive
visibility, retains selector output and sorts it. Peak candidate memory scales with
C*R for collection size C (plus representation/object overhead), up to32 times per
batch, not a streaming scan. Event checks occur during fanout but cannot repair this
prior whole-collection materialization. Serializer snapshots/chunks/strings, Event
arrays, receipts, staged rows and SQLite bindings can coexist. Dense tiny-field
objects have large overhead; no fixed text-byte multiple proves heap safety.

Stress must NOT call inspect(): that helper parses all Resources AND all historical
groups and changes the workload being measured. Use emitted owner/HTTP point reads,
receipts/pages and size-only observation at the single current-group serialization
boundary if necessary. No full history fetch; no synthetic group or DB injection.
The owned author's stopped earlier workload is inconclusive (graph shape and
redundant historical AJV work both changed); do not attribute a diagnosed runtime
performance defect to that observation.

## Bounded implementation and acceptance proposal

No implementation is authorized by this document alone. After root acceptance,
coordinate budget/batch/store and the minimal shared evaluator typing audit. Do not
edit schemas, HTTP set admission, Type lifecycle, public evidence gates or Read
artifacts. Done when:

1. One consumed immutable tuple and required compiler bounds enforce the premises;
   missing/changed tuple refuses before BDP maintenance, after necessary SQLite
   recovery. Valid exact reopen succeeds; old fixtures remain historical.
2. Actual M256/257 repeated-no-op match boundary and E256/257 effective Event boundary
   prove atomic rollback, the exact retained limit Problem and no fabricated index.
   Include cardinality/cap coexistence, no-op count, mixed work, same-key replay,
   changed-key conflict and fresh-key recovery before/after reopen. Existing4096
   evidence is historical, not relabeled as current.
3. Each owned shape has exact Event field assertions: local/remote endpoint cases,
   self-Link both facts, source event last, at most2 Resource effects, no preimage
   body anywhere, no-op keeps both revisions but emits nothing. Source inline state
   and direct postimages both enforce R. Patch tests include transient large values,
   exponent normalization, metadata-only explicit update and properties-only sets.
4. Through normal sub-Q mutations construct near-R records; measure M no-op receipt
   storage/replay/pages and representative E-bounded group serialization without
   inspect(). Include32 explicit operations, names/keys near bounds, aliases, repeated
   source versions, dense arrays/objects and sparse large-string records. Use separate bounded child
   processes and pre-agreed memory/time abort thresholds; record RSS and actual byte
   counts. Abort means inconclusive, never proof of safety. Do not allocate the
   conservative305MB bound merely to validate arithmetic.
5. Stress/recovery fixtures operate on disposable stores, prove clean close and
   reopen exact bytes, and do not scan unrelated historical groups. Configuration
   refusal/hot-journal ordering stays covered by existing recovery witnesses.
6. Focused evaluator/owned/set/receipt/page/recovery checks, compiler/build/boundaries
   and applicable smokes pass. Report precise artifacts, configuration and runtime;
   no broad profile readiness or no-OOM claim. If E256 is impractical, evaluate E128
   with the same M256 before dropping matches; if the candidate scan dominates,
   neither knob solves it and a separately reviewed physical/iterator slice is needed.

The existing-wire fallback remains immutable result/Event/change rows under the
same fenced SQLite transaction, with existing receipt page planning and bounded
page assembly. It requires private format/recovery changes and cannot be smuggled
into a constant repair. No serializer/SQLite exception may become a fake count
limit; no new byte vocabulary or semantics is proposed here.

## Review dispositions and evidence

1. Independent Event knob: accepted, with corrected patch-aware arithmetic.
2. Optional R enforcement: accepted as a required compiler/owner premise, with
   shared-file scope explicitly requiring coordination.
3. Preimage bodies: explicit forbidden proof assumption and regression predicate.
4. Partition: now derived field-by-field and invocation-by-invocation, including
   incidence duplication and owned source patches.
5. Diagnosability: named; no new cardinality admission rule accepted.
6. Scope4096: evaluated, not inferred from unrelated diagnostic limits; original
   private65536 condition already suffices and avoids unnecessary narrowing.
7. Selection materialization: explicit C*R/whole-collection limit of the proposal.
8. inspect(): forbidden in memory/size stress, not merely discouraged.
9. Recovery scope: RU owner excluded by schema/path; TX recovery ordering preserved.
10. Lone surrogates: actual receiver decoder rejects them; recorded separately.

Read-only source inspection and small proof/encoding probe only; no suite, build,
load test, code mutation or outward write. Probe exit0 on Node24.16.0; actual
MAX_STRING_LENGTH536870888 and SQLite MAX_LENGTH1000000000 recorded in
`transaction-materialization-revised-proof.log`. Its initial run assumed lone
surrogates admitted and failed; preserved as `*-initial.log`, corrected to assert
actual refusal. The original shell exit capture also hit zsh's read-only `status`
variable; rerun uses a dedicated variable and records explicit exit. Neither
failure is represented as successful validation. No source checkout was modified.
