# BDP event stream, journaling, and batch payload audit

Date: 2026-10-08. Read-only architecture research; no Preview 2 source, spec, test, database, or runtime changes.

Authority baseline: `gastownhall/bdp@182f1fcf8a01d896976bff3c9e3fb87c596c6ca6`. Existing CLI comparison: `gastownhall/beads@5126de8c1a02d01db2481e4fd5e78dd41010e25f`. Full baseline metadata is in `baselines.json` beside this report.

## Answer

**BDP Events are useful committed-fact log entries, but the Event Source stream alone is not a durable replication or recovery log. The protocol already has the more suitable journal unit: a complete Scope change group, consumed after a transaction-consistent snapshot.** A change group supplies final postimages/tombstones for state, ordered Events for application meaning, erasure records, and a contiguous checkpoint chain. Use those existing contracts before inventing a second journal format. This is an architectural recommendation based on the current specification, not evidence of a running Transactional implementation. [Events][events] [Groups][groups] [Snapshots][snapshots]

**An Event and a batch operation share much of their domain data, especially the exact Property Change representation, but they are not interchangeable payloads.** A batch requests work; Events describe committed effects. Operations can fan out into several facts, produce no fact, or fail. The Event envelope drops request labels, selection intent, explicit guards and the idempotency key, while adding allocated IDs, revisions, transaction identity and observation metadata. Clients never submit Events as mutations. [Batch][batch] [Event wire contract][event-wire]

There are five different meanings of “log”:

| Intended use | Events alone | Existing supported route / gap |
|---|---|---|
| Show an activity timeline of committed, visible changes | Yes, within source retention and authorization | Resource/Scope Event Sources; do not claim attempted-command or authenticated-actor coverage |
| Incrementally advance an already-held resource | Conditionally | Correct previous revision, all needed deltas, relevant Type information, current view, and erasure integration are required; a mismatch requires reread/snapshot |
| Maintain a durable materialized replica | No | Snapshot + complete change groups, atomically applied with a durable checkpoint; this is explicitly the replication contract |
| Recover exact original commands or replay batch intent | No | Retain normalized requests and outcomes separately if needed; event-to-command conversion cannot recover missing intent |
| Restore a full provider/store, including historical and administrative state | No, even with groups | Provider backup/restore also needs Type installation, aliases, authorization policy, identity non-reuse, receipt/idempotency state, private storage state, and any retained history required by the restore target |

## What the current contract actually gives us

- **Commit integrity:** resource state, change group and Mutation Receipt commit atomically. Operations execute in declaration order against staged state; every operation leaves valid staged state. Failed/rolled-back transactions emit no Events. All-no-op or zero-effect transactions have a receipt but no group or new position. An update followed by its reverse is two transitions and two Events even if final properties equal the pre-transaction value. [Transactions][transactions] [Revisions][revisions] [Groups][groups]
- **Ordering:** Scope positions form one total authority order within an epoch. Positions are opaque equality tokens, not sortable counters. `previousPosition` lets a group consumer detect gaps/reordering. Event IDs are source-local; stable group ordinals preserve operation order but can have legitimate gaps from projection or erasure. Neither wall-clock time nor lexical ID ordering is the authority order. [Scope history][history] [Groups][groups] [Event wire contract][event-wire]
- **Framing:** one complete group per finite-page member or SSE message; never split a group. Consumers apply the entire group and advance the durable checkpoint atomically. A set of individual Events sharing `transaction` is not an end-of-transaction framing protocol. Event Sources have no group `previousPosition`, erasure array, complete-state array or group completion marker. [Changefeed wire contract][feed]
- **Authorization:** this is a journal of one authorized projection. A hidden transaction still produces an identifier-free projection advance in the changefeed, so contiguous catch-up is possible without hidden identifiers. A grant, revocation or policy change rotates the view token and requires a fresh snapshot; there are no required incremental authorization-policy Events. A projection tombstone means a resource left the projection, not necessarily that the underlying object was deleted. [Authorization][authorization] [Groups][groups]
- **Retention/reconnect:** snapshots anchor a precise exclusive changefeed checkpoint. Snapshot expiry bounds bootstrap/replay availability, subject to the erasure override. A too-old, foreign-view or foreign-epoch checkpoint fails explicitly and requires a new snapshot. An SSE library's remembered delivery ID is not proof of durable local application: reconnect must use the applied checkpoint. A snapshot recovers current state, not all missed historical activity. [Snapshots][snapshots] [Feed][feed]
- **Erasure:** Event Sources do not deliver erasure records. Erasure may remove previously exposed Event content and leave ordinal gaps, including an erasure-only group with zero Events and zero live-state changes. Persistent Event consumers claiming protocol-backed erasure handling must integrate the changefeed and snapshot erasure ledger. An erasure fence expires earlier checkpoints/snapshots for affected views; even an archive has cleanup duties. The ledger survives epoch changes, view rotation and restore. “Immutable committed fact” therefore does not mean a forever-readable append-only byte archive. [Erasure][erasure] [Persistent consumers][persistent]
- **Audit identity:** BDP expressly provides no authority-attested actor. Optional carried attribution can be claimed/unknown. History's `changeContext` does improve a journal: native created/updated version events carry authority-observed commit time and operation-local agent/message states; it does not attest an actor or restore original request syntax. Plain deletion creates no version/context record; owned-Link deletion supplies context on the new source version. No-ops preserve old context. [Actor attribution][actor] [History context][context]
- **Integrity:** no public cryptographic group digest is required. The chain fences continuity/history replacement under the protocol; it is not an authenticated tamper-evident audit chain. An integrity extension is permitted. [Groups][groups]

## Operation-to-Event matrix

The eight-operation batch union is closed. `putAlias`/`deleteAlias` are excluded. `changeContext` columns below refer to the History amendment, not a new proposal. [Operation schema][operation-schema] [Event schema][event-schema]

| Requested operation | Committed facts | Fields shared or transformed | Information unavailable from the facts alone |
|---|---|---|---|
| `createBead` | `created` | Requested `type` → `subjectType`; requested/allocated ID → canonical `subject`; initial `properties`; carried attribution; generated revision and version context | Whether ID was supplied or allocated; `name` label; omitted/default spellings; Type-derived empty `ownedLinks` keys |
| `updateBeadProperties` | `updated`, unless no-op | Target → `subject`; same committed `change` representation; `previousRevision`, new `revision`, attribution/context | Whether `expectedRevision` was supplied and its request spelling; prior property values not necessarily in patch; no-op attempts; whole resulting properties |
| `deleteBead` | `deleted` | Canonical subject/type, final live revision | Properties/before-image; request guard; deletion attribution/context (no newly minted Bead version) |
| `createLink` | Link `created`; `linked` at each in-Scope endpoint; source Bead `updated` if owned | Canonical Link identity/type/properties/endpoints; stored pins preserved; owned source delta includes created Link record | Local labels/relative endpoint spellings; whether ID allocated; original creation's operation index; derived facts are not additional requested mutations |
| `updateLinkProperties` | Link `updated`; source Bead `updated` if owned; incident projections expose Link update | Link patch and revisions; owned delta repeats Link's identity/patch/revisions/context | Caller guard, prior values not carried in patch; deduplication required across projections; no-op attempts |
| `deleteLink` | Link `deleted`; `unlinked` at in-Scope endpoints; source Bead `updated` if owned | Deletion identity; graph facts carry typed Link reference and both endpoints; owned delta carries deleted Link identity | Link properties and caller guard; attribution/context only on source's newly minted version where applicable |
| `updateWhere` | Singleton-equivalent Events for each affected Resource, canonical-ID order | The applied Property Change appears on changed Resources; attribution/context fan out | Original Selector/collection and selected-but-no-op resources; cannot distinguish one set operation from several explicit operations |
| `deleteWhere` | Singleton-equivalent deletion/graph/owned facts per selected Resource | Final identity/revision, graph endpoints where applicable | Original Selector, selected-set intent, request operation boundary; zero-match set operations leave no Event |
| `putAlias` / `deleteAlias` (outside batch) | **None**; no group or Scope position | Receipt can record alias mutation outcome | Alias table is absent from snapshots/changefeed; offline alias reconstruction is unavailable by explicit contract |
| Failed batch / admitted no-op batch | **None**; receipt has disposition and required position | Receipt records outcome; effect position absent | Attempt, reason for refusal, original request and no-op operations cannot be recovered from Events |

Facts are not an injective encoding of commands: two distinct legal requests can induce the same fact sequence. Therefore there is no general inverse Event→original batch transformation. [Event semantics][events] [Aliases][aliases] [Receipts][receipts]

### Field-by-field batch versus journal record

| Field/concern | Batch request | Event | Change group / receipt |
|---|---|---|---|
| Scope | Discovered target URL | Source and canonical subject URLs | Epoch/view/order explicitly in group; receipt includes epoch/view |
| Retry identity | One HTTP `Idempotency-Key` for whole batch | Absent | Receipt contains key/outcome; group transaction ID is a different token |
| Requested operation | Eight-value `operation` discriminator | Five-value fact `type` | Receipt preserves ordered operation results; group preserves ordered facts |
| Batch-local identity | Optional creation `name`, references such as `@newTask` | Canonical allocated identities only | Creation result maps label/identity; group final identities only |
| Properties | Full creation properties or update `change` | Initial properties on creation; delta on update; no body on delete | Group has complete final live record or tombstone |
| Revision | Optional optimistic `expectedRevision` | Factual predecessor/new revision, or last live revision on deletion | Group preserves final projected revision; receipt records version at each result |
| Selection | `collection` + Selector for set operations | Expanded per-resource facts | Receipt set results/matched count; no original Selector in group |
| Actor | Optional carried attribution on applicable operations; authenticated caller is outside payload | Carried attribution where applicable, not attested caller | Same limitation; private provider audit can be separate |
| Context | Optional operation-local `agent`/`message` input | Version's resolved state envelope, including commit-time state, when History applies | Complete native History records carry the envelope; no-op retains old metadata |
| Transaction boundary | One ordered `operations` array | Shared transaction ID, but individual delivery | Group is the complete atomic application frame; receipt maps mutation to `effectPosition` |
| Operation boundaries | Explicit array indexes | No `operationIndex` or label; `ordinal` counts facts | Ordered receipt results retain operation indexes, but normalized group state does not |
| Before-image/undo | Not generally present; BDP Property Change excludes JSON Patch `test` | Not generally present | Final postimages are not automatic inverse/undo data |
| Erasure | No generic erasure operation in the eight-operation union | No erasure Event type | Group + snapshot ledger provide erasure obligations |

`previousRevision` is **not evidence that the caller supplied `expectedRevision`**. It describes the real predecessor after serialization. Likewise `changeContext.agent` is not authenticated principal identity. [Schema][event-schema] [Context][context]

## Worked example from checked-in fixtures

The actual fixture's first batch requests three operations:

```json
{
  "operations": [
    {"name":"decision","operation":"createBead","id":"beads/dec-9","type":"https://work.example/types/decision","properties":{"title":"Adopt owned Links","status":"proposed"},"attribution":{"principal":"agent:planner","status":"claimed"}},
    {"name":"cite","operation":"createLink","type":"https://work.example/types/cites","source":"@decision","target":"beads/task-42","properties":{"role":"evidence"},"attribution":{"principal":"agent:planner","status":"claimed"}},
    {"operation":"updateBeadProperties","bead":"beads/task-42","expectedRevision":"task-42-r7","change":[{"op":"replace","path":"/status","value":"cited"}]}
  ]
}
```

With HTTP key `client-key-0001`, it commits transaction `txn-0a1b`, position `pos-43`, after `pos-42`. The single group contains:

| Ordinal | Event | Why |
|---|---|---|
| 0 | `created` decision `dec-9`, revision `dec-9-r1` | First requested operation |
| 1 | `created` Link `9c1e`, revision `9c1e-r1` | Second operation; authority allocated Link identity |
| 2 | `linked`, source endpoint | Derived graph fact |
| 3 | `linked`, target endpoint | Derived graph fact |
| 4 | `updated` decision `dec-9`, r1→r2, `ownedLink.created` | Existing owned-Link rule versions source; no separate user update |
| 5 | `updated` task `task-42`, r7→r8, status patch | Third requested operation |

Its `changes` array has **three** final postimages: decision at r2 with the Link inline, first-class Link at r1, task at r8. The created intermediate decision version r1 is in Events/results, not an extra final-state entry. A consumer of the decision's Resource Event Source sees only ordinals **0, 2, 4**. That gap is expected, not lost transport.

The example proves why flattening each Event into a batch operation is wrong: the two `linked` facts are not two extra Link creations, and the owned source `updated` fact must not be applied as another user mutation. Snapshot state at pos42 + atomic pos43 `changes` reaches the intended graph without interpreting these derived facts. A seeded Event reducer can also reach it, but only with prior state and explicit Type knowledge, and it must reject a wrong previous revision. [Batch fixture][batch-fixture] [Changefeed fixture][feed-fixture] [Event fixture][events-fixture]

Exact request, snapshot and complete group are saved in `event-log-worked-example.json`. The fixture predates History context examples and does not claim advertised History; absence of `changeContext` here must not be generalized to native History version events.

## Irreversibility and recovery counterexamples

1. **A guard disappears.** Updating the same current revision with or without `expectedRevision` can produce the same `updated` fact. A log cannot recover the caller's concurrency intent from the predecessor revision.
2. **A set operation disappears into expansion.** `updateWhere` selecting 100 resources and a batch of 100 explicit updates can yield the same resource facts. Replaying the original Selector tomorrow could change a different set; replaying enumerated facts is a different contract.
3. **No-op and failure attempts disappear.** A wrong-revision batch or an unchanged update has no Events. A command/security audit that must include failed attempts needs another source.
4. **Deletion is not undo data.** A deleted Event carries final revision, not the deleted properties. A historical read may help only if that body is retained and authorized; no permanent before-image guarantee follows from Events.
5. **Patch inversion is not assured.** Replacing `/title` carries the new title; removing a property carries its path. Without held prior state or a retained historical body, neither yields an inverse patch. Do not promise undo from deltas alone.
6. **Erasure is invisible to an Event-only listener.** Erasing an old version can create a group containing just an erasure record. An Event-only archive would keep forbidden old content indefinitely unless it separately reconciles the ledger.
7. **A replica is not the entire authority.** Hidden resources, aliases and policy changes are not reconstructible from a visible Event journal. Snapshots contain current Beads/Links, not the full Type installation closure, alias index or receipt/idempotency tables. Type installation is outside the client protocol; recording descriptor versions/packages remains a restore concern, not a request to redesign Types now. [Types boundary][types]
8. **Reissuing derived writes cannot preserve history.** Mutation inputs have no revision/position/time setter. Replaying creates with explicit IDs does not preserve authority revisions, original epoch or receipts; deleted canonical identities cannot be reused. Pins may preserve old revision text while a newly built authority has no matching historical body. Exact history import requires a separate defined restore contract. [Revisions][revisions] [Operation schema][operation-schema] [History][history]
9. **An epoch change is not continuation.** Restores may preserve canonical URLs, but previous history tokens are fenced. Prior-epoch idempotency keys become unbound and can execute anew. Never treat an archived key or Event cursor as global replay protection. [History][history]

## Which “batch command” are we comparing?

There are at least three distinct existing surfaces; their names should not be conflated:

1. **BDP Transactional `operations/batch`** is the generic eight-operation JSON envelope above. It is the natural Issues-pack compilation target when atomic generic mutations suffice. Read+Update's `sequence` is explicitly non-atomic and is not BDP batch.
2. **Current `bd batch`** is a line-oriented narrow command runner: `close`, `update`, `create`, `dep add`, `dep remove`. It shares a storage transaction and rolls back on the first failing line. It has command-specific policy differences—its help explicitly says `update status=closed` checks closure policy while `close <id>` does not. Events cannot reconstruct which of those two command intents produced a similar state. This surface accepts no `apply` subcommand in its parser; comments elsewhere using “bd batch apply” should not be interpreted as the BDP endpoint. [CLI batch][cli-batch]
3. **Existing `POST /v0/beads/issues:batchApply`** accepts an issue-specific JSON plan, `{actor, force_id_prefix, items, provenance, skip_per_edge_cycle_check}`. Its tagged items are `create`, `update`, `close`, `dep_add`, with local key references, issue patches, guards, force flags and end-of-plan domain validation. It is an atomic heterogeneous plan, not the same payload as BDP batch or Events. `bd create --graph` uses this family of compiled issue plans. Its actor is explicitly caller-asserted provenance, not authentication. [HTTP batch apply][legacy-batch]

The existing Beads journal is also distinct. `journalops.Row` contains `Seq, TS, Op, IssueID, Actor, IssueJSON, DepJSON, CommentJSON`; the issueops seam emits post-mutation snapshots in the mutation transaction. Seven engine operations become six public operations: comment rows advance the source cursor without a separate wire event. Its row shape has no transaction/group identity. In shape, the raw journal's postimages are closer to BDP `changes` than BDP `updated` deltas. It therefore must not be used as proof that the proposed BDP stream already provides equivalent semantics. [Journal row][legacy-journal] [Journal emission][legacy-emission]

## Practical architecture choices

1. **Materialized-state journal:** make a Rust client library consumer for snapshot + changefeed. Persist `(scope URL, epoch, authorization view, applied position/checkpoint)` with atomic group application. Validate owned inline/first-class agreement; separate ordered fact notification from normalized state application; deduplicate complete checkpoints; recover by snapshot on gap, view change, epoch change or erasure expiry.
2. **Activity journal:** expose Event projections as a convenience built on the same committed groups. Preserve source-local identities and original ordinals. If the client persists event content, integrate erasure ledger cleanup and record gaps/rebootstrap honestly. Do not make event retention imply security audit or forever replay.
3. **Command audit, if required:** retain client-side requested command + compiled BDP request + resolved receipt linkage as a separate concern. A client log covers that client only; an authority-wide log must be provider-owned and commit-linked. Decide whether the requirement includes attempted/refused requests, authenticated principal, supplied guards, set-selection intent and command/pack version before proposing wire extensions. Receipts help connect request and group but expire/redact; they are not a permanent full-request audit export.
4. **Provider restore:** keep physical/logical backup and authority metadata preservation with the provider. Current BDP supports a read/materialized replica, not independently writable replay, multi-authority merge or full store import. If portable full-fidelity backup becomes a requirement, specify its contents and preservation semantics explicitly; do not overload batch until those choices are made.

The implementation cost question is also real: a Link creation can produce three Events, or four for an owned Link, in addition to final postimages; a self-Link still emits two endpoint facts. Set mutations expand per selected Resource; authorities must bound induced Events and reject the entire oversized transaction. Chunking reduces pressure but loses whole-job atomicity. For a compact state journal, store final `changes`; for intermediate version/activity history, retain `events` too, subject to erasure/retention. Measure both byte volume and operation count. [Expansion and limits][events] [Set bounds][set-bounds]

## Verification performed and limits

- **PASS — 158 existing tests, two files:** `packages/protocol/src/transactional-wire.test.ts` and `packages/conformance/src/transactional-catalog.test.ts`, run with Node 24.16.0 and the exact locked dependencies in the isolated BDP snapshot. Complete output: `event-log-existing-tests.log`.
- **PASS — eight additional evidence probes:** `python3 event-log-probes.py`; complete output `event-log-probes.log`. They check the pinned fixture's 3→6→3 mapping, snapshot/group versus seeded-event reconstruction, rejection of a wrong predecessor, legitimate projection gaps, erasure-only groups, missing inverse-request fields and deletion before-images. The small reducer intentionally supports only the selected fixture, not all BDP operations.
- **No provider Transactional conformance was run or established.** The existing wire test explicitly calls its fixtures narrated examples, not behavior evidence. The catalog test asserts that no executable manifest names the Transactional catalog. The checked-in server's admission boundary only admits Read. Passing these tests proves artifact consistency and the stated small fixture exercises, not runtime atomicity, throughput, streaming recovery, erasure races, crash durability or operational correctness. [Wire-test caveat][wire-tests] [Catalog caveat][catalog-tests] [Server profile][server-profile]

[events]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1701-L1940
[groups]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1942-L2034
[snapshots]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L2036-L2068
[transactions]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1246-L1289
[history]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1198-L1245
[revisions]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L697-L767
[batch]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L4419-L4526
[event-wire]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L6119-L6273
[feed]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L6040-L6117
[authorization]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L904-L948
[erasure]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L5860-L5952
[persistent]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L6219-L6234
[actor]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L844-L901
[context]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L3360-L3407
[operation-schema]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/schemas/bdp-v0.schema.json#L1260-L1380
[event-schema]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/schemas/bdp-v0.schema.json#L2751-L3058
[aliases]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1572-L1585
[receipts]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1663-L1698
[types]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L220-L238
[set-bounds]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1635-L1652
[batch-fixture]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/fixtures/transactional/batch.json
[feed-fixture]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/fixtures/transactional/changefeed.json
[events-fixture]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/fixtures/transactional/events.json
[cli-batch]: https://github.com/gastownhall/beads/blob/5126de8c1a02d01db2481e4fd5e78dd41010e25f/cmd/bd/batch.go#L20-L84
[legacy-batch]: https://github.com/gastownhall/beads/blob/5126de8c1a02d01db2481e4fd5e78dd41010e25f/internal/httpapi/batch_apply.go#L18-L145
[legacy-journal]: https://github.com/gastownhall/beads/blob/5126de8c1a02d01db2481e4fd5e78dd41010e25f/journalops/journal.go#L5-L33
[legacy-emission]: https://github.com/gastownhall/beads/blob/5126de8c1a02d01db2481e4fd5e78dd41010e25f/internal/storage/issueops/journal.go#L18-L85
[wire-tests]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/packages/protocol/src/transactional-wire.test.ts#L12-L25
[catalog-tests]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/packages/conformance/src/transactional-catalog.test.ts#L17-L26
[server-profile]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/packages/server/src/index.ts#L269-L325
