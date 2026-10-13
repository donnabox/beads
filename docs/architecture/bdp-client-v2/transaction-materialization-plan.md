# Private transaction materialization budget — proposed bounded follow-up

Read-only plan against frozen **064db8a99629122f130a91ee8b5189432ffe97e9** in
`vickie-transaction-set-bounds` (parent8d64fcf). Prospective owned transitions are
based on `transaction-owned-transitions-plan.md`, not implemented evidence. No
source, schema, public profile or limit advertisement changed. Root/council must
approve scope before implementation.

## Recommendation

Start the reference fixture at **32 cumulative matched Resources per transaction**,
keeping the existing32 explicit-operation,4096 Event,65536 request-byte and262144
per-Resource representation bounds. Count repeat matches and no-ops, preserve
cardinality-before-budget ordering, and retain the existing413/limit-exceeded/
transaction.matchedResources failure with no invented operation index.

A fixed count of128 can satisfy a conservative whole-string proof below, but its
505,675,776-byte prospective group bound is too close to the actual536,870,888-code-
unit V8 ceiling to be a comfortable reference budget.64 is intermediate;32 gives
more headroom while matching the existing explicit-operation capacity. This is a
reference implementation operational tradeoff, not a BDP product maximum.32 is a
proposal, not a committed limit or public claim.

Two small installation conditions are necessary for an unconditional bound over
new records: bound the configured Scope's encoded overhead, and prevent silently
opening stores whose retained data was admitted under the prior larger budget.
Lowering only the batch.ts constant is insufficient to assert a new bound for all
served records. Do not redesign storage merely to avoid those private conditions.

## Source facts and counterexamples

At frozen064:

- transaction-batch.ts54–67:32 operations,4096 induced Events,4096 cumulative matches,
 65536-byte carrier,65536 properties,524288 diagnostic bytes/16 diagnostics.
- transaction-batch.ts707–750: each valid set selects once at its staged turn,
 applies cardinality first, charges selected.length including repeated/no-op matches,
 then emits one matched entry and invokes the existing evaluator once per Resource.
- transaction-store.ts747–755 bounds idempotency keys to256 ASCII characters;
 principal strings are nonempty/well-formed but have no length bound.
- transaction-store.ts944–956 allocates UUID identities/revisions.989–993 supplies
 the actual262144 representation cap; resource-evaluator.ts1048–1100 checks both
 materialized subject and source before writes. Deletion does not cascade; a Bead
 with incident Links is refused (1028–1029). Alias mutations have no Resource Events.
- transaction-batch.ts544–570,624–657 constructs current lifecycle/incidence Events.
 Current unowned Link creation/deletion produces at most3 facts, update1. The owned
 plan adds at most one source updated Event and one source Resource: at most4 facts
 and2 changed/deleted Resource identities per invocation, not an arbitrary cascade.
- transaction-store.ts1012–1070 serializes one complete terminal receipt and one
 complete effective group before persistence. Receipt paging does not remove that
 work:481–501 stores ranges over the existing complete result list.
- protocol/json-values.ts95–141 snapshots values, collects chunks and joins one
 complete string. There is no implicit output cap, streaming or byte-limit refusal.
- transaction-store.ts175–196 accepts canonical Scope without a length bound and
 persists exact Scope in configuration;848–860 includes its receipt URL even for
 zero-match work. Thus a count bound alone is not a bound over arbitrary installation
 metadata. The current configuration lacks the matched cap and representation tuple.
- transaction-store.ts819–850 compares retained keys before fresh admission; replay
 must not be reevaluated against a newly lowered match cap.

The prior emitted probe built an exactly262144-byte Resource through ordinary
mutations (largest request60138 bytes), then persisted/reopened/replayed a32-match
no-op receipt of8,392,367 bytes.4096 such postimages alone project1,073,741,824 bytes.
This is a reachable input family; no GiB allocation/failure was attempted. See
`transaction-set-bounds/receipt-size-probe.mjs/.log`. Pinned Node24.16.0 reports
MAX_STRING_LENGTH536870888; its SQLite reports MAX_LENGTH1000000000. The initial
4096 cap does not close that materialization gap.

## Conservative whole-string calculation

All lengths below are compact JSON UTF-8 bytes, not JS object sizes. A byte bound
also bounds UTF-16 code units for the well-formed encoded JSON, so it is sufficient
for the observed V8 string ceiling as well as SQLite's TEXT byte ceiling.

Let:

- M = cumulative set matches; O=32; K=M+O. Each match invokes evaluation once; every
 explicit operation invokes at most once. K deliberately overcounts set carrier
 entries as if they were additional singletons. No-op invocations remain counted.
 There are at mostO extra matched-summary entries.
- R=262144 for each complete current/new Resource, including metadata, owned inline
 records, references and attribution. Tombstone identity is a subset of a live
 record. New installs contain only records written under that bound; corrupted rows
 and external database edits are integrity faults outside this proof.
- Q=65536 raw carrier bytes. S is encoded JSON byte length of Scope+"events/", and
 the proposed private installation precondition is S<=Q. Receipt URL suffix adds
 only a fixed number of bytes covered below. This bounds actual configured Scope,
 including reopen, not merely a short test URL.
- N=7Q bounds reserialized admitted request data. Do NOT assume N=Q:1e20 is4 wire
 bytes and21 JSON output bytes. JSON.stringify finite-number tokens are at most25
 characters. Input number tokens of at least4 bytes therefore expand by at most
 25/4<7; one/two-byte tokens do not expand; three-byte tokens expand to at most10
 characters (e.g.9e9), also<7. Ordinary JSON string/key escaping does not enlarge a
 valid original token's UTF-8 encoding; whitespace removal reduces it. Punctuation
 does not expand. Thus7Q safely covers all normalized patch data in one carrier.
 Scope-relative identity substitution is NOT covered by N; Resource fields below
 cover canonical references separately. Patch payload values are not @ bindings.
- A=8192 fixed structural bytes per invocation, plus H=1048576 per whole document.
 These deliberately loose allowances cover JSON keys/punctuation, four Event
 envelopes, two change wrappers, UUID revisions/transaction, timestamps, integer
 ordinals/positions, result tags, stable view/epoch and256-byte key. Variable Scope,
 Resource data, operation name, patch payload and diagnostics are accounted for
 separately. Neither allowance licenses arbitrary new extension fields.

`transaction-materialization-ratios.mjs/.log` exercises the emitted decoder,
admission and encoder on exponent/negative-zero/Unicode/control-character examples;
observed1e20 ratio5.25 and9e9 ratio3.333. The general7x bound is the token argument
above, not an empirical claim inferred from a few examples. Initial script invocation
omitted its required diagnostic-budget argument; corrected invocation passed. No
source tests or suite were run for this read-only plan.

Conservative per-invocation group contributions, including up to2 final changes:

1. Owned creation/deletion is bounded by **11R+4S+A**. Bound Link lifecycle Event by
 2R+S, each of two graph facts by2R+S, source updated by3R+S (source identity and
 attribution, nested complete Link record), and final changes by2R. Deletion is
 smaller. Include previous UUID revisions and all fixed keys in A. Creation emits
 no user-supplied patch arrays.
2. Owned update is bounded by **7R+2S+2N+A**: Link lifecycle<=2R+S+N; source update
 <=3R+S+N; final changes<=2R. Only two Events; no graph facts on update. N includes
 both propertiesChange and metadataChange together. This intentionally counts the
 same input patch twice when the source Event embeds the Link delta.
3. Unowned transitions fit those envelopes. No-op/alias mutations contribute no
 group. Each invocation may count the same final Resource in this bound even though
 the actual changes Map keeps only its last postimage, so deduplication only helps.
4. A receipt result is bounded by **2R+Q+A**: direct Resource/tombstone, possible
 source/sourceRevision, and operationName from the carrier. Alias outcomes (Scope+
 path and target identity) fit the same bound. No-op receipts retain full postimages
 and obey exactly this bound. Each matched entry is below128 bytes. Add S+H once for
 the receipt envelope. Failed receipts have no partial results: diagnostic bytes
 <=524288, name<=Q, Scope<=S, fixed fields withinH, far below the successful bound.

Use the larger group coefficient for every K invocation, although mixed operations
cannot simultaneously maximize all terms. This yields:

| M | K | Whole group, create/delete bound | Whole group, update bound | Whole completed receipt bound |
|---:|---:|---:|---:|---:|
|32|64|202,899,456|186,122,240|39,391,232|
|64|96|303,824,896|278,659,072|58,527,744|
|128|160|505,675,776|463,732,736|96,800,768|
|256|288|909,377,536|833,880,064|173,346,816|

Formulas: `K*(11R+4Q+A)+H`, `K*(7R+16Q+A)+H`, and
`K*(2R+Q+A)+32*128+Q+H`. Numeric values use the maximal permittedS=Q.
128 passes this deliberately conservative string test;256 does not. This does NOT
prove256 actually produces a failing group: correlation between inline source and
Link sizes, carrier size and reused identifiers may support a tighter proof later.
There is no reason to pursue that optimization before reference delivery.

Existing sequences contain at most32 singleton RU-kind members, not nested set
batches. Their combined response therefore contains at most32 single-Resource/alias
receipts, not32 M-sized batch receipts. Preserve that restriction; nesting batches
or set carriers in sequences would require redoing the aggregate response bound.
Receipt page output contains a subset of an already-bounded receipt plus bounded
Scope-derived next URI; continuation never recomputes mutable effects. Future pages
that aggregate multiple complete change groups need their own output budget; this
per-group proof is not that qualification.

## Why even32 is not a heap proof

Two hundred MB is a coarse worst-case group bound, not a desirable fixture size.
JSON snapshots, admitted values, staged row copies, changes, Event arrays, serializer
chunks, joined strings, SQLite bindings and response Buffers can coexist. JavaScript
object/node overhead depends on shape and is not bounded by a small fixed multiple
of text bytes. Many tiny fields can be worse than one large string. Candidate scans
remain independent of M and can visit an arbitrarily large collection before the
match guard. Owned inline state amplifies copies even when the final group dedupes.

Consequently the claim should be: a small fixed budget removes the demonstrated
platform string-limit exposure for newly admitted records under the stated
installation; measured stress determines whether it is practical. It is NOT
"cannot OOM", a total authority memory bound, a scan bound, or complete TX readiness.
The existing32 explicit operations already consume part of the conservative bound;
reducing M cannot erase that baseline. Prefer32 initially; do not promote128 simply
because it barely fits an engine constant. If real maximum-shape tests are still
impractical at32, reduce the private budget further or take the physical alternative
below—do not catch allocation failure and invent a count-limit Problem.

## Installation, retention and identity consequences

1. Use one immutable source-private budget tuple for existing count/request/Resource
 bounds and startup proof, consumed by evaluator/owner. No options, flags, public
 exports, wire fields or new byte-limit vocabulary. Validate S<=Q BEFORE provisioning
 or maintenance. Persist/bind the tuple in existing exact installation configuration
 alongside Scope. The configured Scope is already persisted; reopen must validate
 that same value and tuple. Fixed owner/view/UUID formats remain proof assumptions.
2. Reuse the current incompatible-installation refusal for an older configuration
 without this tuple. Do not silently interpret old4096-budget stores as the new
 installation. Keep old files/artifacts historical and use a new disposable store;
 no migration/data loss is authorized. Private compatibility identity change does
 not change semantic request equality policy or normative wire format.
3. If compatibility with old files is required instead, it needs an explicit legacy
 carve-out: old receipt/group bytes remain replayable at their original size and are
 outside the new bound, or a reviewed physical migration. Lowering a live limit may
 affect fresh keys but must never turn an already-retained equal key into a new limit
 refusal or recompute its result. Retained conflict still wins. Failed limit outcomes
 stay bound until existing failed-key forgetting rules apply; fresh keys then observe
 current state/current capacity. No cap value is added to canonical request identity.
4. The proof is about receipt/group JSON strings. SQLite also sizes complete rows:
 raw trusted principal currently has no bound and is stored beside receipt/identity/
 witnesses. A blanket whole-database-row-fit claim would additionally require a
 compact validated principal/configuration boundary and bounds for identity/witness
 encodings. Existing installed fixture principals are small; do not pretend the
 public protocol limits that hidden trusted string. If root wants the broader row
 guarantee in this slice, validate the captured principal identifiers using the same
 privateQ budget and account for identity/witness overhead, without inventing IAM.

The customer-visible tradeoff is explicit: a fresh set selecting33 matches will
atomically fail under proposedM32 even when each Resource is tiny;32 repeated matches
of the same Resource consume the same budget as32 distinct matches. There is no
silent first-page mutation. Splitting work trades away cross-chunk atomicity; clients
that require a larger atomic set need a differently provisioned implementation,
not an automatic retry that hides this distinction. Cardinality failure still wins
when the client's own guard is violated. The4096 Event cap remains a true upper bound
but may be unreachable at the current lower invocation budget; do not claim that
its boundary was exercised without a lower Event-cap test seam or a valid future
fanout path.

## Candidate implementation scope and done predicates

After council acceptance, coordinate transaction-batch/store with the owned-Link
worker. Proposed scope: existing private budget constants, owner configuration and
Scope startup check, set/store tests, narrow progress doc. Preserve exact shared
contract capture; no schema, public admission, HTTP sets, Type lifecycle or new
storage framework. A source-private helper only if it reduces duplicated constants;
no public option to raise the budget.

Done when all of these are executable or precisely source-derived:

1. One fixed tuple is consumed by preparation/evaluation/owner representation limits
 and installation identity. A changed/missing tuple or invalid Scope refuses before
 provisioning/maintenance; reopening the valid exact tuple succeeds. Old-source
 fixture proves compatibility refusal, not reinterpretation of historical receipts.
2. Actual M matches succeed and M+1 refuse atomically; repeated no-ops count. Preserve
 the guard-vs-cap discriminator, rollback preceding effects, lack of fabricated
 operationIndex, and retained equality/conflict/fresh-key behavior through reopen.
 The current4096 regression becomes historical evidence, not a false current limit.
3. Build R-sized metadata through ordinary sub-Q requests, then execute M repeated
 no-ops; measure actual receipt encoding, replay, observation and pages through
 emitted owner/HTTP. No DB injection, limit raising or fabricated Resource result.
4.32 explicit operations remain supported. Exercise mixed sets/singletons, maximum
 name/key bytes, alias put/delete sequence, exponent patches, escaped strings,
 unchanged before/after state on limits, and failed diagnostic bound. Every counted
 invocation is covered; no reliance solely on returned result order.
5. Owned integration establishes2-Resource/4-Event maxima, before/after Resource
 checks and exact payload formulas with real fixtures before claiming that portion
 implemented. Test source-inline size nearR, repeated source versions, self-Link,
 two endpoint facts and no-op sourceRevision. Record maximum actual group/receipt
 strings and peak RSS for large strings AND dense small-field payloads in bounded
 child processes. Fix an execution-time/memory qualification budget before the run;
 abort means inconclusive, not safe. Do not allocate the conservative500MB128 bound
 merely to demonstrate arithmetic.
6. Compiler/build/boundaries, focused transaction/evaluator/receipt-page/recovery
 regressions, and existing packed/development smokes pass. Frozen Read evidence is
 neither resealed nor relabeled. Report exact Node/SQLite limits and bound arithmetic.

## Smallest fallback under existing wire

If count reduction plus explicit private installation conditions is unacceptable,
reuse existing receipt page planning but store immutable result entries separately
from receipt header/core; likewise store each Event/change separately from group
header under the SAME SQLite transaction and ownership fence. Reads assemble only
one existing bounded wire page. Keep stable receipt/page identities, exact retained
order/postimages, crash atomicity and all existing receipt/Problem shapes. This
changes private physical format and recovery/expiry/indexing; it is a separate
reviewed slice, not a quick constant fix. It still needs bounded page materialization
and does not by itself make an arbitrarily large unpaged group response practical.
Do not map a serializer/SQLite allocation exception to matchedResources or
mutatedResources: those fields are counts. No new error vocabulary is proposed.
