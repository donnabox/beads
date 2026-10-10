# Beads 1.3 with generic events versus current BDP

2026-10-09. Supporting analysis for [design.md](design.md), requested by Donna. This compares a hypothetical modification of shipped 1.3 with the current normative BDP contract. It does not claim the modifications exist, accept release scope, or change either specification. CLI rationalization is excluded.

## Baselines and assumptions

Beads baseline: v1.3.0, `f45b249ce6b40ba62aecc03949e6371e8f7c79d8`. BDP `main`, freshly checked: `182f1fcf8a01d896976bff3c9e3fb87c596c6ca6`. The left column below is called **modified 1.3** and assumes:

1. Closing is a generic Bead update of status/closure properties. There is no `close` event type.
2. Comments are logically part of their parent Bead's state, retaining comment identity, author, text, time and provenance. Appending one is a generic Bead update. This is the inline collection scenario from our discussion, not a new decision about physical tables or paging. A renamed record that still omits comment state would not satisfy this assumption.
3. Links have stable, nonreused lifetime identities, including delete/recreate handling. The adapter distinguishes creation from metadata update, and produces the five generic verbs with endpoint facts as described in the [worked model](five-verb-event-model.md). This gives the proposed mapping the benefit of all its intended semantics, not just new names.

Everything else remains 1.3 unless explicitly marked as further work. In particular, these assumptions do **not** quietly add transaction envelopes, Resource revisions, snapshots, epochs, authorization projections, complete write coverage, or BDP JSON encoding. The resulting record schema has not been specified; schematic examples below are semantic traces, not invented wire contracts.

The BDP comparison is its **Transactional profile**. Read and Read+Update do not expose BDP Events and need not supply replication. Also distinguish two existing BDP surfaces: individual **Events** for observation, and **Scope change groups** for atomic replica advancement. They derive from the same committed groups, not two independently authored logs. [BDP model][events] [Profiles and replay][delivery]

## Finding

With those assumptions, **the graph vocabulary can match BDP exactly**. The remaining differences are substantial and mostly concern payload semantics, reliable replay, and which state is represented. Modified 1.3 would still be a stream of captured Issue postimages and adapted graph facts. Current BDP offers revisioned semantic deltas and, in its changefeed, final canonical state within an explicit atomic/recovery contract.

The journal already has a useful foundation: transactional capture, commit-ordered sequence allocation, durable replay, polling/SSE, and explicit truncation. Those should be reused. Full postimages are not a replication defect; BDP's replica surface uses them too. Missing commands and before-images are not the obstacle to forward reconstruction.

## 1. Mutation meaning, side by side

| Action | Modified 1.3 | Current BDP | Remaining difference |
|---|---|---|---|
| Create Bead | `created` with resulting Issue state, including initial comment state under the assumption. | `created`: immutable subject ID/Type, initial properties, initial revision; optional version attribution/context. Group also has canonical postimage. | Canonical record, revision, metadata and encoding still need mapping. |
| Change title/status/labels | `updated`, retaining the existing full resulting Issue image unless a further payload change is chosen. | `updated` with predecessor/new revisions and committed Property Change. Group has final postimage. | A postimage is sufficient to replace known state but is not the BDP Event delta contract. |
| Close/reopen | Ordinary `updated`; Issue-specific status and closure fields are data. | Ordinary `updated`; protocol has no knowledge of closure. | Vocabulary resolved. All changed closure fields must be captured, not merely the status string. |
| Append comment | Ordinary parent `updated`; comment object survives in the logical parent state. | Ordinary property update; a JSON Patch `add` at `/comments/-` can express append. | Owner revision, canonical reads and size/representation policy must agree. BDP imposes no native Comment concept. |
| Create Link | `created L`, plus `linked L` at each represented in-Scope endpoint. | Same facts; Link lifecycle first, source endpoint before target. | Link's complete canonical properties, revisions and exact references are further work. |
| Refresh Link metadata | `updated L`, retaining lifetime ID. | `updated L` with predecessor/new revisions and patch; incident Bead Event Sources also observe the Link update. | Need state-change/no-op detection and revisioned payload, not just re-label every `dep_add`. |
| Delete Link | `deleted L`, plus endpoint `unlinked` facts. | Same facts; deletion carries final live revision; graph facts carry both endpoint references and Link ID/Type. | Final revision and exact endpoint facts must survive deletion. |
| Delete then recreate same pair | Retire L1 and allocate L2. | New immutable Resource identity; old canonical ID never reassigned, even after epoch change. | Assumed solved for Links; the guarantee must be durable and shared with canonical reads for BDP. |
| Parallel same-pair Links | Legacy producer cannot create them, but generic event schema can represent them. | Independent Link identities represent each Link. | No stream incompatibility. Current store remains a restricted graph; this does not require a CLI change. |
| Semantic no-op | 1.3 emits `dep_add` even for accepted same-type re-adds; unchanged logical state must be detected by the adapter if normalizing. | Equal resulting properties preserve revision and emit no Event. Failed/rolled-back work emits none. | ID/verb changes alone do not establish BDP no-op semantics. |
| Owned Link mutation | No assumed owned-Link projection/version rule. | Additionally updates the owning source Bead's revision and inline owned set; its `updated` Event follows Link/endpoint facts. | Conditional requirement if the chosen model uses BDP Owned Links; ordinary dependency does not automatically mean owned Link. |

Sources: [1.3 record, dependency upsert and coverage][journal], [dependency metadata writer][deps], [BDP Event expansion and payloads][events], [Property Change/no-op rules][patch], [identity][identity].

BDP endpoint facts do not by themselves advance the endpoint Bead's revision. Their subject is the Link. A self-Link creates two endpoint facts, distinguished by `endpoint: source|target`; out-of-Scope endpoints have no local Bead Event Source. This is generic graph behavior, not dependency-specific behavior. Owned-Link versioning is a distinct rule, included here only for accurate comparison, not to reopen Trish's design scope.

## 2. Every published field has a different job

The shipped 1.3 envelope is `seq, ts, op, issue_id, actor?, issue, dep?, comment?`. Close/comment/Link normalization necessarily changes some of these fields, but no final replacement envelope has been agreed. This table does not pretend otherwise. [Encoder][encoder]

| Information | Modified 1.3 inherits or assumes | BDP Event | BDP change group |
|---|---|---|---|
| Delivery identity | Integer `seq` belongs to this clone/history. One source row may expand to several generic facts, requiring distinct stable projected IDs. | Opaque source-local `id`, stable from group checkpoint, authority ordinal and source projection. | Opaque `checkpoint` resumes after a whole group. |
| Subject identity | Issue ID; new lifetime ID for Link subjects. | Absolute canonical Resource `subject` URL and immutable `subjectType`. | Each state entry identifies Bead/Link kind and carries canonical identity. |
| Event kind | Assumed five generic verbs. | `type`: `created`, `updated`, `deleted`, `linked`, `unlinked`. | Group is transport/control framing, not a sixth graph Event type. |
| Event Source | Clone journal; no assumed standard per-Resource Event Source. | `source` URL: Scope source or Bead/Link projection. IDs are meaningful only within that source. | One Scope order, projected for an Authorization View. |
| Order within transaction | Flat `seq`; no published transaction membership or ordinal. | `transaction`, stable zero-based `ordinal`, possibly gaps after projection. | `position`, `previousPosition`, ordered `events`; `eventCount` counts served entries. |
| Time | `ts`: UTC insertion time inside transaction. | RFC 3339 `time`; not a substitute for transaction/order fields. | Position chain, not timestamp grouping, establishes order. |
| Resource revision | Internal Issue RowVersion is omitted from journal Issue JSON. No assumed Link revision wire contract. | Create: new revision. Update: predecessor/new revisions. Delete: final live revision. | Final projected revision in every postimage/tombstone. |
| Resulting state | Full Issue row plus labels/readiness; assumed comment state integrated. Link state needs mapping beyond old `dep` tuple. | Initial properties only on `created`; `updated` carries a delta. | Complete final canonical postimage per affected live Resource. |
| Delta | Not recorded as a generic Property Change. May be derived from known prior/final state for a new projection contract. | Ordered JSON Patch using only `add`, `replace`, `remove`, or one owned-Link delta. | Events preserve ordered transitions; `changes` supplies normalized final state. |
| Actor/provenance | Optional event-level actor string; comment author/provenance retained as data. | Optional per-version `attribution {principal,status}`; status `claimed` or `unknown`, not authenticated identity. | Canonical postimages preserve version attribution. |
| History context | No equivalent closed version envelope assumed. | With advertised History, native created/updated versions carry `changeContext` as well. | Complete native version postimages carry the same context under History. |
| History/view fence | No epoch or Authorization View token. | Event cursor is bound to its source/view retention contract. | Explicit `scopeEpoch`, `authorizationView`, checkpoint and position chain. |
| Removal from replica view | No separate concept beyond underlying delete. | No false `deleted` Event for mere view departure. | Identity-bearing tombstone can mean deletion **or** projection departure. |
| Historical erasure | Prefix retention pruning only; no per-version erasure signal. | Event Sources do not deliver erasure records. | `erasures[]` plus snapshot ledger reconcile retained version content. |

Sources: [1.3 encoder][encoder], [RowVersion omission][rowversion], [BDP Event wire fields][delivery], [group schema][groupwire], [attribution][attribution], [History context][context].

Two mapping traps matter. First, legacy `issue_type` cannot silently become BDP's immutable Resource Type if the legacy field remains mutable; an Issue classification can instead remain an ordinary property. Second, the old dependency payload only carries kind, target and metadata (source is outer `issue_id`). Existing dependency storage also has creation fields and `thread_id`; assigning a Link ID does not magically make those fields present in the old stream. Choose the canonical property set and capture every promised field. [Dependency storage shape][depshape]

BDP History's `changeContext` is capability-scoped, not an unconditional new base field. It contains `committedAt`, `agent`, and `message` with explicit present/absent/undetermined states. Native versions in one transaction share the authority-observed commit instant. A 1.3 row's insertion timestamp cannot be relabeled as that exact instant, and missing legacy context must not be fabricated. Context is separate from actor attribution.

## 3. Replay and recovery guarantees

| Guarantee | Modified 1.3 | Current BDP Transactional contract | Consumer consequence |
|---|---|---|---|
| Atomic capture | Covered mutation and journal rows share the storage transaction. Counter changes roll back too. | State, change group and mutation receipt commit atomically. | Reuse the journal's capture foundation; it is already stronger than best-effort notifications. |
| Commit order | Gapless, increasing sequence during an intact local history; rollback burns no number. | One Scope position per event-producing transaction; linked predecessor positions. | Both order committed work. BDP exposes the transaction-level chain. |
| Atomic delivery/application | Individual rows; page limit or dropped connection may separate one transaction's records. | JSON pages and SSE messages never split groups; consumer applies complete group and checkpoint atomically. | Modified 1.3 cannot promise observers never see half a transaction. |
| Multiple writes to same object | May retain successive images; no final-per-transaction normalization contract. | Ordered Events retain each transition; `changes` has at most one final entry per Resource. | An activity log and replica need different views of the same commit. |
| Reconnect/dedup | Consumer persists processed `seq`; SSE Last-Event-ID takes precedence. | Event IDs resume observation; replica resumes from durable applied group checkpoint, not merely last received SSE ID. | Delta append must not run twice. Automatic transport resume alone is not atomic durable application. |
| Bootstrap | Export/full-read advice; no protocol that couples the baseline to an exact journal continuation point. | Consistent Bead+Link snapshot at one epoch/view/position; precise exclusive checkpoint; retain continuation through advertised expiry. | Fix the baseline/head race before claiming a correct cold-start replica. |
| Retention | Prefix prune; default retain last 7 days or newest 100,000 rows, whichever keeps more. Typed truncation/HTTP 410; SSE truncation closes stream. | Explicit expiry/foreign-context failures; snapshot recovery. Global retention window, not per-consumer pins. | Both allow bounded history; BDP is not a forever journal or consumer-ack retention system. |
| Reset/source switch | Clone/branch local. In-place counter/history reset can make an old cursor above head look caught up. | Foreign epoch/checkpoint must fail explicitly. | Needs history identity, not just a larger integer. |
| Capture completeness | Opt-in at every writer; sync/merge, raw SQL, compact/restore bypasses remain. | Every relevant visible transition must be represented or continuity invalidated/recovery required. | Gapless numbers do not prove complete state coverage. |
| Authorization and projection | Shared journal credential; no assumed per-client graph projection protocol. | View-bound projection, explicit state entry/exit; grant/revocation rotates view and requires fresh snapshot. | Cannot treat resource invisibility as resource deletion. |
| Hidden transaction progress | No corresponding contract. | Identifier-free projection advance at same Scope position when transaction has no visible effect. | Replica can prove catch-up without receiving hidden Resource/transaction identities. |
| Erasure | Old content survives edit/delete until pruning; pruning is not a subject/version erasure protocol. | Projected erasure records, permanent ledger, old checkpoint expiry and snapshot reconciliation; Event-only acquisition insufficient. | Ordinary delete and historical content removal are distinct. |
| Freshness proof | `head` shows journal drain progress, not a strict-read or complete-source guarantee. | Strict reads report epoch/view/position; minimum checkpoint requires replica to catch up, route, wait or fail. | Empty tail does not mean “all canonical state is current” in 1.3. |
| Bounds | HTTP rows/page bounded; no BDP semantic expansion or projection-limit contract. | Finite Event expansion bound; reject oversized mutation before commit. Oversized derived view transition triggers view reset, not rejection of valid authority mutation. | Generic Link expansion and large transactions need explicit admission/framing limits. |

Sources: [capture/counter][writer], [1.3 journal coverage/retention][journal], [HTTP reader][http], [BDP groups][groups], [snapshots and strict reads][snapshots], [group wire/recovery][groupwire], [erasure][erasure].

The transport overlap is real, but the interfaces differ:

| Surface | Modified 1.3 inherits | BDP today |
|---|---|---|
| Finite replay | `GET /v0/beads/events?since=N&limit=M`; required integer `since`; default 1,000/max 10,000 rows; records and head/floor progress. | Discovered Scope `changes/` uses explicit `after` checkpoint or `start=now`; Event sources use opaque Event `after`, omitted to start at earliest retained Event. |
| Live replay | Separate `/v0/beads/events:watch` endpoint; individual record per SSE message. | Same discovered resource with `Accept: text/event-stream`; changefeed emits complete `change-group` messages, Event source emits individual generic Events. |
| Reconnection | Numeric Last-Event-ID overrides `since`. | Source-appropriate opaque Last-Event-ID overrides `after`; replica must supply its durable applied group checkpoint. |
| Per-resource observation | No BDP Resource/Event-Source identity contract assumed. | Bead/Link canonical URL with `view=events`, plus Scope `events/`. |

These are HTTP differences, not a CLI redesign proposal. The two systems already share the broad polling/SSE mechanism. Neither promises transport exactly-once delivery. [1.3 transport][journal] [BDP delivery][delivery] [Changefeed start][feedstart]

Lease heartbeat is deliberately outside the journal, even though Issue snapshots may contain sampled lease fields. A replica needs an explicit state boundary: do not promise live lease authority from those samples. This is separate from the generic Event vocabulary. [Coverage test explanation][heartbeat]

## 4. The same transaction rendered both ways

Assume one actual storage transaction closes A, appends comment C to A, and changes ordinary Link L's metadata. This example excludes unrelated derived readiness changes. If the real operation changes another Resource, that effect must also be captured. Assume the adapter has the needed starting state; IDs/revisions below are symbolic, and field paths are illustrative.

Modified 1.3, keeping its postimage orientation:

```text
seq 101: updated A; full A after close
seq 102: updated A; full A after comment C
seq 103: updated L; resulting Link state
```

Those rows only become readable after commit. But a page ending at 101 gives no evidence that 102 and 103 belong to that same transaction. No amount of timestamp matching can make that inference reliable. If the comment update instead carries only an append delta, replay must deduplicate that delta; the “full images are safe to replace twice” rationale no longer applies to it.

Current BDP, semantic shorthand:

```text
ChangeGroup: epoch E, view V, position P, previousPosition Q,
             checkpoint K, transaction T, projectionAdvance false
  events (eventCount = 3):
    ordinal 0: updated A, a7 -> a8, replace /status and other closure fields
    ordinal 1: updated A, a8 -> a9, add /comments/- = C
    ordinal 2: updated L, l3 -> l4, change Link metadata properties
  changes:
    upsert A: complete canonical final state at a9
    upsert L: complete canonical final state at l4
  erasures: []
```

The replica installs both postimages and K atomically. An application may inspect all three Events. The replica must not install A at a9 and then append C again by applying the Event to that final image. Current BDP deliberately carries both semantic transitions and final state in the same committed group. Removing or deriving one representation is a future design choice, not today's contract.

No-op suppression is tested per operation against its immediately preceding staged state. Changing a property and changing it back in the same transaction is two real BDP transitions, with two Events and a final newer revision; it is not a no-op merely because the final properties equal the original properties. A final-postimage-only diff cannot recover that ordered history.

A Link creation similarly expands to three Events for two distinct in-Scope endpoints: one `created` and two `linked`. It need not change either endpoint Bead's Resource revision. If that Link is owned, the owning source adds its own `updated` Event and postimage. Consequently **command count, journal-row count, Event count and final-state-entry count are different quantities**.

## 5. What can consumers do?

| Consumer job | Modified 1.3 | BDP individual Events | BDP Scope changefeed + snapshot |
|---|---|---|---|
| Show activity or trigger an index refresh | Yes, for covered mutations and retained history. | Yes; generic Resource/Scope sources and ordered semantic facts. | Yes; Events are included, although group delivery is more than a simple observer needs. |
| Maintain a best-effort local index | Yes, with known baseline and explicit resync handling. | Yes, with required starting state and revision checks; re-read when not positioned to apply a delta. | Yes, with stronger complete-view guarantees. |
| Reconstruct state from complete known history | In principle for the state actually captured; our changes retain comments/Link identity. Actual bypasses and bootstrap limit the guarantee. | Possible under constrained complete-history/known-schema assumptions, but not the general protocol-backed replica contract. | Explicitly supported current authorized state reconstruction. |
| Expose only whole committed transactions | No published grouping contract. | Transaction labels/order exist, but individual delivery is observational. | Yes, complete group application. |
| Prove a cold-start copy has no bootstrap race | No. | No independent snapshot rendezvous/replica contract. | Yes. |
| Track authorized projection entry/exit correctly | No generic contract. | Lifecycle facts alone are insufficient: appearance in a view need not mean creation. | Yes, postimages/tombstones plus view reset. |
| Reconcile historical erasure obligations | No subject/version mechanism. | No, it must integrate changefeed and snapshot ledger. | Yes, under the defined acquisition/recovery contract. |
| Restore every original command, failed attempt or provider configuration | No. | No. | No; this is not the purpose of current-state replication. |

Even BDP's own five-verb Event Source is therefore **not currently specified as sufficient for a general replica**. That observation does not prove Events cannot be made sufficient. It identifies exactly what an event-first redesign must retain: projection transitions, initialization information, atomic grouping, progress, recovery and erasure. No Issue-specific event kind is needed for any of those requirements. [Current separation][delivery]

## 6. Recommendation and remaining decisions

Use the five-verb model as the common semantic vocabulary. Then make explicit choices about the remaining work rather than calling all of it “event normalization.”

1. **Publish a coherent generic record model.** Canonical Bead/Link identities, immutable protocol Types, complete property mapping, Resource revisions and a comment representation shared by reads, snapshot and stream. Suppress semantic no-ops. Preserve known provenance and truthful absence.
2. **Add the replica foundation.** Durable complete commit framing, epoch/checkpoint chain, complete capture or explicit invalidation, consistent snapshot/checkpoint rendezvous, atomic consumer application and replay deduplication. These are necessary even for a single unfiltered workspace with no sophisticated authorization policy.
3. **Choose payloads explicitly.** A full-postimage journal can replicate efficiently enough in some workloads; a small patch avoids retransmitting large descriptions/comment threads but demands proper base revisions and deduplication. Current BDP requires both ordered deltas and final postimages in groups. Initially sharing capture and deriving both views is lower risk than discarding either without evidence; measure bytes and workload before changing the protocol.
4. **Account for the full BDP contract where advertised.** View projection/progress/reset, erasure acquisition and recovery, strict-read freshness and bounds. A single all-visible view simplifies implementation but does not excuse silent gaps or ambiguous cursors. Owned-Link and History payload requirements apply when those features are used/advertised; they do not force Issue dependencies to become owned or reopen Type lifecycle design.

These are recommendations for discussion, not release 1.4 commitments. The first three assumed changes solve domain vocabulary and Link lifetime representation. They do not already solve the rest of this list. Conversely, we should not demand original commands, arbitrary event sourcing of provider storage, or perpetual retention: current BDP requires none of those for forward replication.

## Reassessment: useful journal information as protocol input

Donna's principle is to take information exposed by working 1.3.1 seriously as a candidate for generic BDP when it benefits implementation efficiency or convenience. Do not dismiss it as legacy baggage merely because the current protocol omits it. This is an accepted evaluation criterion; the proposals below remain open. Here, working 1.3.1 means 1.3 plus the proposed close/comment changes, not a claim about the already published version tag.

**Computed state correction.** Current `is_blocked` is producer-maintained and explicitly published in full journal Issue images. Maintenance emits ordinary updates when affected values change. The JSON field uses `omitempty`, so a false value is absent from the replacement image. Readers need not recalculate the dependency graph. Generic BDP can publish that value as an ordinary property change; selecting the Issue-aware writer of that change is a separate architecture decision, not an Event-model incompatibility. It is not a promise that every ready-query condition is embodied in `is_blocked`. [Capture and hydration][writer] [Field representation](https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/types/types.go#L37-L41)

**Owned Links and complete images.** Current BDP already supplies both an owned Link's final state/tombstone and its owning source's final postimage in a change group. Individual Events remain deltas. A proposal to make individual observations self-contained must consider owned Link and owner images as well as ordinary relationships. Distinguish a related source image included for convenience from a source version that truly changed. Avoid blindly repeating large owner records in every derived endpoint fact; payload selection/placement remains open.

**Deleted content.** Favor making the final pre-delete canonical record available to a stream consumer for both Beads and Links. This generalizes 1.3's removed dependency metadata and helps stateless cleanup, routing and inspection. The decision between always including it and a guaranteed requested enrichment is open. It must be captured before deletion, correspond to the final live revision, respect pre-state visibility and existing retained-content erasure obligations. This is not yet a normative requirement.

**Successful no-op observations.** Donna challenged the absence of no-op records. Emitting evidence of an accepted no-effect operation can be useful and can preserve state correctness. Recommend supporting that observation without minting a Resource revision, changing version attribution, creating adjacency facts, or inventing an owned-source transition. One candidate is an explicitly marked no-effect `updated` record with unchanged predecessor/current revision and empty effect delta; admitting it would deliberately broaden the current Event definition. Whether it belongs there or in generic operation outcomes in the same journal remains open. A new operation and a retry under the same idempotency identity must be distinguished; retries must not create duplicate original-operation evidence. Define completion, durable replay position and receipt linkage for a no-op-only transaction. Today's no-group/no-position rule would need revision if such records use the changefeed. “No state change” does not mean “no useful observation,” but existing consumers cannot retain an assumption that every update Event advances revision.

**Operation attribution.** Recommend a generic envelope describing attribution of the operation being observed, including delete and successful no-op. Keep it distinct from attribution of an existing Resource version: deleting Alice's version by Bob should preserve both facts without rewriting Alice's version. Specify truthful absence and claimed/unknown provenance; do not infer authenticated identity. This addresses an information gap with value beyond Issues. No field name or actor-attestation scheme is accepted yet.

**Cursors.** Do not preserve exact legacy row numbering unless an actual consumer requires old checkpoints to continue. Native BDP consumers need stable replay identity, ordering and durable checkpoints, not one sequence number for each historical Beads writer call. If legacy and enriched streams coexist, identify them as separate projections with their own cursor contracts, or supply an explicit migration mapping. Details are deferred; the representation difference is not inherently a generic-model blocker.

## Verification

Fresh remote checks confirmed the v1.3.0 annotated tag's target and current BDP main. Ten Beads source files were re-read through the GitHub API and compared byte-for-byte with the pinned local source copies. BDP was independently reviewed against the current remote specification, including the History amendment omitted by some earlier base Event examples. Existing source fixtures/tests were read as evidence; no runtime tests, database writes, implementation, release qualification or normative specification edits were performed.

[journal]: https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/docs/reference/events-journal.md
[encoder]: https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/eventsjournal/record.go#L32-L67
[writer]: https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/storage/issueops/journal.go
[deps]: https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/storage/issueops/dependencies.go#L260-L278
[depshape]: https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/types/types.go#L1104-L1123
[rowversion]: https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/types/types.go#L76-L99
[http]: https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/httpapi/events.go
[heartbeat]: https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/storage/issueops/journal_completeness_test.go#L199-L215
[events]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1701-L1940
[groups]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1942-L2034
[snapshots]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L2043-L2065
[delivery]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L6121-L6275
[groupwire]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L6037-L6119
[patch]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L4912-L4955
[identity]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L617-L640
[attribution]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L844-L900
[context]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L3358-L3410
[erasure]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L5737-L5949
[feedstart]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L5950-L5967
