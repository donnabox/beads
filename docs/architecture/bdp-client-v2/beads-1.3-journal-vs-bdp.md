# Beads 1.3 journal versus BDP Events and replication

2026-10-09. Research and recommendations for Donna, requested after Chris raised the overlap. This is supporting research for [the fork design](design.md), not a release commitment or normative protocol amendment. The bounded-expression interview question remains open; this research does not accept it.

## Recommendation

**Converge the transactionally captured facts, while preserving the distinct consumer contracts.** Beads 1.3 already has much of the machinery we want: durable records in the mutation transaction, commit-ordered replay, resumable reads, explicit truncation and full resulting issue state. Its closest BDP counterpart is the **Scope changefeed**, especially its state-change/postimage side. Calling it equivalent to BDP's individual Event feed would hide important differences.

For 1.4, prioritize a versioned journal contract with transaction completion, history identity/reset detection, and a consistent snapshot/checkpoint handoff. Preserve the 1.3 reader surface during migration. Capture enough generic Bead/Link effects at the commit boundary to support both the legacy Issues projection and BDP's semantic Events/change groups. Do not discard full postimages merely to make the payload look like BDP Events; whether deltas become the canonical transfer representation remains an open, measurable design choice.

Do not build two independently authored logs, require an event-sourced database, or assume an adapter can recover all BDP semantics from old journal rows. BDP already permits one atomically committed group containing Events and state changes. A relational provider can commit state and the durable record/outbox together. If an outbox later publishes the feed asynchronously, it must preserve commit order, completeness and replay/freshness guarantees; ordinary eventual notification is insufficient for the advertised replication contract.

## Sources and scope

- Beads **v1.3.0**, released 2026-09-15, tag commit `f45b249ce6b40ba62aecc03949e6371e8f7c79d8`: shipped reference, encoder, SQL writer and golden records. [Release][release] [Journal guide][journal] [Encoder][encoder] [Writer][writer] [Fixture][fixture]
- Beads v1.3.1: `c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c`; current upstream main examined at `d15958a5eb88fc32e3a47680ccc017ff47b561ac`. Six core files (guide, writer, encoder, journal reader/pruner, HTTP reader and SSE reader) are byte-identical across these three baselines. Surrounding write paths have changed, as described below. Pending PRs are separately identified; they are not shipped behavior. The [independent source audit](beads-1.3-journal-source-audit.md) records those comparisons.
- BDP current main verified at `182f1fcf8a01d896976bff3c9e3fb87c596c6ca6`. This is a specification comparison, not a claim that a released provider implements the entire Transactional profile. [BDP Events][bdp-events] [BDP groups][bdp-groups]

The replica target here is current authorized Bead/Link state. Restoring provider configuration, complete historical versions, failed requests, or every original command is a different requirement. Lack of original commands and before-images does not prevent forward state reconstruction.

## What overlaps and what differs

| Aspect | Beads 1.3 journal | BDP contract | Consequence |
|---|---|---|---|
| Purpose | Machine replay of this clone's covered Issue mutations; separate from audit history and hooks | Semantic observation Events plus a lossless Scope replication changefeed | Compare all three, not just similarly named endpoints. |
| Commit capture | Journal row and mutation share a transaction; sequence allocation rolls back with it | State, change group and mutation receipt commit atomically | Strong common foundation; preserve it. |
| Record vocabulary | `create`, `update`, `close`, `delete`, `dep_add`, `dep_remove`, `comment` | Generic `created`, `updated`, `deleted`, `linked`, `unlinked` on Beads/Links | Domain operations require mapping, not renaming. |
| Payload | Full issue row/labels/readiness image; separate dependency/comment payloads; null issue on delete | Events carry committed deltas; groups also carry final canonical Resource postimages/tombstones | Beads is closer to group state changes, but has Issue-specific side records and no group envelope. |
| Transaction delivery | Flat sequence of rows; no transaction ID, ordinal/count or end marker in the published record | Groups are delivered and applied whole; Events carry transaction and ordinal but individual Event delivery is observational | A consumer cannot recover exact transaction boundaries from Beads timestamps or adjacency. |
| Identity/order | Clone-local integer `seq`; `issue_id`; bare sequence is tied to source history | Canonical Resource IDs/types/revisions; opaque source IDs/cursors; Scope epoch, view and position fences | Add history identity; do not expose a Dolt counter as a universal BDP cursor. |
| Bootstrap | Guide says baseline from export/full read and follow the journal; no coupled snapshot/checkpoint contract | Transaction-consistent Bead+Link snapshot with exact exclusive checkpoint and retained continuation | Needs a race-free bootstrap, not just an inexpensive head query. |
| Coverage | Opt-in at each writer; direct covered mutations only; sync, raw SQL, compact/restore bypass it | Transactional authority must account for every relevant visible transition, or invalidate continuity | Strict replication cannot tolerate silent bypasses or disabled writers. |
| Projection | Whole journal under a shared server credential | Authorization View, projection entry/removal via state changes, hidden-transaction progress, view reset | Existing Issue rows alone cannot implement arbitrary authorized BDP views. |
| Retention/recovery | Prefix pruning; floor/head; typed truncation; default 7 days or newest 100,000 rows, whichever retains more | Expired/foreign epoch/view checkpoints explicitly fail; snapshot recovery; erasure reconciliation | Reuse explicit failure discipline, extend identity and recovery semantics. |
| Historical erasure | Old full issue images remain after ordinary edit/delete until pruning | Separate version erasure duties, records and snapshot ledger | Ordinary deletion is not historical erasure; do not claim equivalence. |
| Actor | Optional actor string, including caller-supplied comment author | Carried per-version `attribution` with `claimed`/`unknown`; no generic authority-attested actor | Preserve provenance without upgrading a string into authenticated evidence. |

Sources: Beads [record contract and coverage][journal], [SQL capture][writer], [published encoder][encoder]; BDP [payloads][bdp-events], [atomic groups and snapshots][bdp-groups], [Event delivery][bdp-feed], [attribution][attribution].

## Concrete mappings

| Beads journal operation | Plausible BDP representation | Missing decision/information |
|---|---|---|
| `create` | Issue Bead `created`; canonical state in group `changes` | Stable canonical ID, Type and revision mapping. |
| `update` / `close` | Bead `updated`; close is a status-property change | Journal carries resulting image, not the committed patch or predecessor revision. Comparing known images can produce a state-equivalent delta, but need not recover the original semantic operation sequence. |
| `delete` | Bead `deleted` plus projected tombstone | Final live revision and transaction framing; renamed legacy IDs require explicit identity mapping. |
| `dep_add` | Link `created` or `updated`, with endpoint graph facts when newly incident | Existing Beads accepts metadata-refresh re-adds: this is an upsert, not invariably a new Link. Link identity/revisions and prior existence matter. |
| `dep_remove` | Link `deleted` plus endpoint `unlinked` facts | Dependency tuple is not yet a canonical Link identity. |
| `comment` | Generic resource mutation(s) under the eventual Issues-pack comment model | Model is not decided here; preserve structured versus audit provenance. Do not invent a sixth generic BDP Event called comment. |
| Derived `is_blocked` update | Explicit Bead update if the chosen model persists it, or a derived client/query result | Preserve current command behavior while deciding whether this materialized property belongs in the new model. |

One dependency mutation may induce a Link lifecycle fact and multiple endpoint facts in BDP. A batch may produce many Events but one final state image per affected Resource. Thus neither journal-row count nor BDP Event count is the batch-operation count. A replay consumer applies committed effects; it should not resubmit these records as batch commands and accidentally re-run derived effects. [Beads dependency contract][journal] [BDP Link expansion and normalization][bdp-events] [BDP groups][bdp-groups]

## The differences that affect correctness

### 1. Atomic capture is not atomic replay

Consider a transaction that changes two issues. Both journal rows become visible after commit, but the HTTP `limit` or a dropped stream can separate them. The 1.3 record has no group completion information. A consumer applying each row immediately can expose half the transaction; equal timestamps do not identify a transaction. BDP's complete change-group frame directly supplies that boundary. The problem is framing, not missing original command intent. [Record][encoder] [HTTP paging][http] [BDP framing][bdp-feed]

For new 1.4 history, capture a transaction identity plus ordinal and a reliable completion boundary, or persist/publish complete frames. A transaction ID alone is not enough: the last transaction must complete even when the stream goes quiet. Old rows cannot safely be grouped retroactively by timestamp. Define a new history boundary and baseline for the stronger contract.

### 2. A baseline and a separately sampled head can lose updates

If a client reads A at state A1, a mutation produces journal sequence 101 with A2, and the client then samples head 101, installing A1 and following after 101 misses the update. Reading a head first and buffering all subsequent effects can support a reconciliation scheme, but it needs a specified complete algorithm, retention and concurrent-pagination behavior. The current export/full-read advice does not provide BDP's snapshot guarantee.

Recommend a snapshot of all mirrored records and side collections at one checkpoint, with continuation retained for the snapshot lifetime. An inexpensive `events head` is useful for observation and lag monitoring, but does not make an arbitrary export atomic. [Beads bootstrap advice][journal] [BDP snapshot][bdp-snapshot] [Pending head command][head-pr]

### 3. A gapless sequence can still omit state changes

The guide explicitly excludes sync/merge, raw SQL, compaction and restore. Journaling can also be disabled on one writer while another publishes the feed. None of these omissions has to leave an integer hole. The guide also warns that journal recreation/counter restoration can make an old cursor appear caught up indefinitely. A per-URL checkpoint is insufficient when history changes behind the same URL. [Coverage and resets][journal]

Heartbeat is another explicit exemption: it is transient lease liveness even though issue snapshots may contain sampled lease fields. A journal mirror must not treat those samples as current lease authority. The replica target must explicitly include or exclude operational lease state. [Heartbeat exemption](https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/storage/issueops/journal_completeness_test.go#L199-L215)

For a strict profile, either journal the relevant committed effects, disallow bypasses while that profile is active, or rotate an epoch and require resnapshot. Enforce this at the authority/provider boundary; a per-process preference is not a completeness guarantee. A non-replicating legacy mode can remain opt-in with its existing limits.

### 4. Comments expose a real contract inconsistency

The guide and writer comments describe a six-operation downstream event vocabulary that skips `comment`, saying its effect is already present in issue snapshots. Yet the guide says snapshots do not inline comments, and the shipped encoder and golden fixture explicitly expose a separate `comment` payload. The actual journal can carry comment content; a six-operation projector that drops it cannot reconstruct comments. This does not establish a bug in the shipped HTTP read, which publishes the payload. It identifies unsafe guidance for a future projector. [Guide][journal] [Writer vocabulary and hydration][writer] [Encoder][encoder] [Golden comment rows][fixture]

Fix the wording and pin a comment-preserving mirror contract before reusing this feed for BDP. The pending shape-adaptive PR can drop unavailable `comment_json` with a warning; that fallback must not silently satisfy a strict complete-replay capability. Distinguish missing attribution from missing comment state. [Adaptive PR][adaptive-pr]

## Suggested 1.4 scope

1. **Preserve compatibility and name guarantees.** Keep the shipped JSONL and `/v0/beads/events` contract; introduce a versioned/enabled stronger stream rather than silently changing operation meanings. State which clients are observers and which can maintain a complete replica. Capture format/version and capability information explicitly.
2. **Add the replication foundation first.** History/source identity and epoch; transaction-complete replay; consistent snapshot plus checkpoint; durable client checkpoint guidance; explicit errors/reset on unsupported continuation. Do not reuse old cursors or infer lost boundaries.
3. **Close or fence coverage gaps.** Require complete capture for every writer behind a replicating authority. Treat import, sync/merge, SQL, compaction/restore and activation changes as explicit policy cases. Preserve all replay state, including comments, dependencies and derived effects promised by the selected model.
4. **Share capture across surfaces.** Produce domain-independent resource effects and necessary semantic facts in the transaction, then project the legacy Issues journal and BDP surfaces. Do not try to synthesize full BDP identity, exact semantic history and transaction boundaries from already-lossy old rows. The reusable provider library can offer encoders/reducers/conformance cases; the store provider still owns the endpoint and atomic commit integration.
5. **Keep payload optimization separate.** Preserve postimages initially where they simplify compatibility; measure full images versus deltas on large descriptions and small updates before removing them. Do not expand 1.4 into a forced rewrite of the current store. Full BDP projection/erasure/receipt guarantees are required before claiming Transactional BDP, even if a narrower improved journal ships first.

If release capacity is limited, fix misleading completeness claims and ship a deliberately bounded journal contract first. Do not declare the existing event endpoint BDP-compatible merely because it uses SSE and similarly named operations.

## Changes around the unchanged journal

Beads 1.3.1 adds post-commit rechecks of derived blocked state. One user command can therefore lead to an original transaction and a subsequent settling transaction; command identity and transaction identity must remain distinct. Current main adds version-history bookkeeping and label-rename coverage, but the journal envelope remains unchanged. Its issue JSON omits the internal `RowVersion`; new version-history records are not automatically BDP revision facts in the journal. [Independent audit and source anchors](beads-1.3-journal-source-audit.md#changes-since-130-and-pending-work)

## Existing work to reuse

At inspection, [PR7211][adaptive-pr] and [PR7213][converge-pr] are open. They address old journal column shapes, migration convergence and diagnostics. PR7213's body explicitly defers outbox/epoch work to later items in a referenced plan; that is evidence of related planning, not proof those items are implemented. [PR7144][head-pr] is also open and adds a cheap head/floor command. Coordinate the stronger framing/bootstrap/epoch contract with those owners instead of creating a competing journal migration.

This review does not modify these PRs or Janet's Preview 2 artifacts. Its recommendations belong to the fork design until Donna decides release scope.

## Clarification: zero domain events, comments and Link updates

Donna ruled on 2026-10-09 that the generic stream has zero domain-specific event types (D03). The earlier examples of a comment Bead were alternatives, not a requirement or accepted data model. Likewise preserving the current journal verbatim is a compatibility option needing an explicit migration decision, not a reason to retain domain events in the generic stream.

Current structured comments are records with `id`, `issue_id`, `author`, `text` and `created_at`. They have string identities, though not canonical Bead identities. They are not a string array. The journal additionally distinguishes structured and audit provenance. [Shipped Comment shape](https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/types/types.go#L1593-L1604)

If the chosen logical model embeds a collection of comment objects in its parent Bead, appending one can be a generic update of that collection. A committed patch could append an object at `/comments/-` rather than retransmitting the whole collection. Exact wire paths are illustrative. This requires canonical reads/snapshots to include or consistently expose the collection and the parent version to advance with it; renaming today's `comment` record to `update` alone does not supply those semantics. Keep comment IDs and attribution/time metadata as data. Do not silently replace object identity with an unstable array offset or lose retry deduplication. Efficient append, large-thread reads and concurrent writes remain design choices.

Current dependencies **do** carry metadata: the shipped Dependency includes a JSON-encoded `metadata` string, creation attribution/time and a thread ID. A same-type dependency re-add updates stored metadata and emits the existing `dep_add` journal record as an upsert. Therefore current link-like data has mutable state that a generic stream must represent. Recommend generic `updated` with a Link subject rather than adding `dep_updated`. Creating/removing a relationship still produces generic Link lifecycle/endpoint facts as defined by the chosen contract. Were a particular Link Type immutable and propertyless, it would simply never need a property-update event. [Shipped Dependency shape](https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/types/types.go#L1104-L1123) [Metadata refresh and journal emission](https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/storage/issueops/dependencies.go#L260-L278)

The close/update asymmetry is also in the shipped writers: dedicated close emits journal `close`; the generic update writer emits journal `update` even when the audit event names closure. Normalize the generic stream to the committed state transition, independently of the command entry point. [Close](https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/storage/issueops/close.go#L374-L389) [Update](https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/storage/issueops/update.go#L540-L546)

## Validation needed before promising replication

- A transaction changing A and B, interrupted after the first transport fragment, is never published half-applied; the quiet final transaction completes.
- Snapshot plus replay equals canonical state under concurrent writers, including dependency metadata refresh, comments, cascades, renames and readiness changes.
- A disabled writer or any bypass cannot silently change an authority that advertises complete replay.
- Reset, restore, source swap, branch change, pruning and malformed/foreign cursors either replay safely or fail explicitly.
- New readers detect old/incomplete payload shape; old readers retain their published contract.
- BDP adapters preserve canonical identity/revisions, authorization-view transitions, erasure and idempotent mutation outcomes; replica replay emits no new authority mutations.

The initial comparison verified 20 immutable remote source files and 45 file/line references plus relative file links. The later clarification additionally inspected the shipped Comment/Dependency shapes and close/update/metadata write paths. This was source/spec research and an independent code-reading pass. No new runtime tests, database mutations, release qualification, implementation or merge were performed. Existing test source and golden fixtures are evidence of intended behavior, not fresh test results.

[release]: https://github.com/gastownhall/beads/releases/tag/v1.3.0
[journal]: https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/docs/reference/events-journal.md
[encoder]: https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/eventsjournal/record.go#L32-L67
[writer]: https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/storage/issueops/journal.go
[fixture]: https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/cmd/bd/testdata/events_journal_records.jsonl
[http]: https://github.com/gastownhall/beads/blob/f45b249ce6b40ba62aecc03949e6371e8f7c79d8/internal/httpapi/events.go#L95-L160
[bdp-events]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1701-L1940
[bdp-groups]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1942-L2065
[bdp-snapshot]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L2043-L2065
[bdp-feed]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L6037-L6268
[attribution]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L844-L900
[adaptive-pr]: https://github.com/gastownhall/beads/pull/7211
[converge-pr]: https://github.com/gastownhall/beads/pull/7213
[head-pr]: https://github.com/gastownhall/beads/pull/7144
