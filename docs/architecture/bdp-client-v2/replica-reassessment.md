# Reassessment: can BDP Events reconstruct a replica?

2026-10-08. Analysis of `gastownhall/bdp@182f1fcf8a01d896976bff3c9e3fb87c596c6ca6`. This supplements, and deliberately corrects the scope of, the earlier audit. No upstream changes or runtime implementation are made by this document.

## Corrected answer

**Yes: committed lifecycle events can reconstruct ordinary Bead/Link state. BDP already carries most of the necessary data.** Given a correct starting state, complete applicable events in order, and the necessary record metadata, creates insert records, updates apply deltas, and deletes remove records. Full postimages on every update are a choice, not an inherent requirement of replication.

My earlier answer put too much weight on information needed to recover commands, audit requests, undo mutations, or restore a whole provider. **Guards, labels, original selectors, request idempotency keys and before-images are not necessary for ordinary forward replication.** Their omission does not show that an event stream cannot maintain a replica. The replica applies facts that already committed; it does not reissue original commands through mutation APIs.

The precise limitation is narrower: **the current Event Source contract is an observation projection, without all the initialization, membership-transition, transaction-delivery and erasure semantics required for the strongest replica claim.** These are fixable protocol choices. The spec's existing change-group envelope solves much of that problem; its duplicated final-state postimages are not the only possible solution. [Event definition][events] [Event Source contract][event-source] [Change groups][groups]

## Keep four requirements separate

| Requirement | What it needs | What it does not need |
|---|---|---|
| Current authorized Bead/Link copy | Correct base; complete changes to visible state; deterministic forward application | Original commands, caller guards, failed attempts, old property values that have been overwritten |
| Atomic, live, resumable replica | Above, plus transaction completion, snapshot/stream rendezvous, recovery fences, durable progress and freshness semantics | Physical store layout or exact original batch syntax |
| Historical replica / full provider backup | Retained old versions as required; erasure reconciliation; additional provider state for a full backup | Cannot be inferred merely from a current-state replica claim |
| Command/audit journal | Requested intent, attempted/refused operations and identity claims appropriate to that audit | These are additional requirements, not prerequisites for forward state replication |

An ordinary replica need not contain aliases, Type installation administration, provider configuration or receipt tables unless we explicitly include those in the replication target. Their absence from Events should not be used to defeat a claim about authorized current Beads and Links. Current BDP itself defines replication as an authority snapshot/changefeed feeding a cache or materialized replica, not an independently writable cloned authority. [Scope history][history]

## The existing event payload is already close

- `created` provides subject identity/type, initial properties, revision, attribution/context where applicable, and Link endpoints.
- `updated` provides actual predecessor revision, new revision, and the Property Change or owned-Link delta. The replica needs its held previous state, not the caller's original `expectedRevision`.
- `deleted` provides enough identity to remove the local record. Deleted properties are needed for undo/history, not deletion from the current map.
- Owned-Link source revision changes are explicitly represented. The spec actually supplies a forward-application algorithm. Derived `linked`/`unlinked` notifications must not be mistaken for additional requested writes.
- No-op and failed transactions can be absent without harming current state: they change nothing. A transaction that changes a value and then changes it back still produces both transitions, as required for ordered history. [Payload and reducer][payload] [Revisions][revisions]

Our existing fixture probe already demonstrates this: snapshot + event reducer reaches the same final records as snapshot + group postimages for the example with three operations and six events. It requires a known base revision and known owned-Type keys. That is positive evidence for event reconstruction over the illustrated transitions, not a proof of a complete generic client or a running provider.

## Concrete limitations of the present Event Source contract

### 1. Visible-state membership is not the same as resource lifecycle

The specification explicitly permits a resource to leave a projection within a stable Authorization View, and supplies a projection tombstone for that case. Separately, it exposes an `updated` Event only if its references are visible in **both** the pre-state and post-state. [Projection tombstones][membership] [Event visibility][event-source]

Minimal counterexample: a replica holds visible Bead A. An allowed stable-view projected transition removes A while A remains live in the authority. There is no resource `deleted` fact; an `updated` fact that is not visible in both states is filtered out. A consumer restricted to the five existing fact types can retain A incorrectly. The change group's projection tombstone removes it.

The reverse case needs an entry image: a previously hidden, already-existing Bead becomes visible. There need not be a `created` fact, and an ordinary property patch would not provide the hidden fields the replica never held. A projection-entry record with the full visible resource solves this; replaying unseen historical state is neither necessary nor appropriate.

This is a missing projected-state transition contract, not an argument for full postimages on every update. The spec does not fully enumerate which policy/data transitions keep a view token stable; the counterexample relies on its explicit admission of stable-view removals, rather than pretending all authorization changes behave that way. Grants/revocations rotate the token and already require a new snapshot. A simpler alternative is to require reset/resnapshot for **every** visibility-membership change, with a potentially substantial efficiency cost. [Authorization views][authorization]

### 2. Snapshot bootstrap has no Event Source rendezvous contract

A snapshot gives a Scope checkpoint. The changefeed accepts precisely that checkpoint as its exclusive continuation. Event Sources instead accept an Event ID local to that source. The spec does not define passing a snapshot checkpoint to the Event Source, nor a translation from the snapshot position to an Event cursor. An Event ID's stable derivation from checkpoint/ordinal does not let clients decode or synthesize opaque tokens. [Snapshot contract][snapshots] [Event cursors][event-source]

Minimal operational case: a snapshot is installed at position P while writes continue. To start delta application, the client needs “all effects after P, none silently missed.” The currently specified Event request cannot directly express that handoff. Starting at the earliest retained event may overlap the snapshot; starting observation after installing it may miss the interval. A subscribe/buffer/read reconciliation algorithm could be designed, but that is additional semantics and complexity, not a rendezvous guarantee already stated.

A finite retained event history also need not contain the creation of a still-live resource. Therefore “rebuild from an empty map using whatever events remain” is not universally supported. This is a bootstrap/retention issue shared by any finite log, not a deficiency of deltas. A snapshot plus an aligned cursor fixes it without update postimages.

### 3. Atomic live delivery needs a completion boundary

Suppose a transaction changes A and B. The SSE Event Source sends one Event per message. After receiving A's event, the client cannot tell whether it has received the whole transaction or B's event is still in flight. A shared transaction ID identifies membership but not completion; ordinals can legitimately have gaps. Publishing immediately can expose half a committed transaction. Waiting for the next transaction stalls the final transaction on a quiet stream. [Event SSE][event-source]

This does **not** prove that atomic reconstruction through Event reads is impossible. A finite read reaching `next: null` means it reached the Source's current end; with the after-commit visibility rule, a client can use that fact and additional reads to finish a transaction. It is a more complicated polling/reconciliation design. A complete transaction frame or an explicit completion marker supplies the missing live-delivery boundary directly. Existing changefeed pages/SSE messages never split a group. [Changefeed framing][feed]

Likewise, hidden no-effect transactions need not be reproduced to keep the visible graph correct. Projection advances and head positions are useful for proving contiguous progress and satisfying minimum-position reads. Their absence from individual Events is a progress/freshness-contract difference, not missing visible properties.

### 4. Empty owned-Link keys are metadata, not a fatal information gap

A newly created Bead's Event omits the empty `ownedLinks` key set. The spec says those keys come from its Type Descriptor and that the record/snapshot is authoritative for them. Two replicas without that metadata could disagree on the exact canonical representation, although both know there are no owned Links. [Owned creation][payload]

A correct descriptor cache or one canonical read resolves it. Alternatively, make creation/entry events carry the complete initial canonical record, including empty keys. This is a bounded metadata or payload decision; it does not justify requiring full records after every subsequent patch. Existing source updates already carry their owned-Link deltas and fresh revision.

### 5. Erasure matters, but the claim must be scoped accurately

Erasing historical version A/r1 while current state is A/r2 can produce no Event and no live-state change. **That alone is not a counterexample to correctness of a current-only graph replica that retains no r1 content.** My earlier emphasis could imply otherwise.

It is a counterexample to an Event-only *persistent history/archive* consumer: that consumer might keep r1 forever without receiving an erasure instruction. Even a current-state replica may retain logs, snapshots, old indexes or caches that need cleanup. The present contract explicitly requires persistent Event consumers claiming protocol-backed erasure handling to integrate the changefeed and snapshot ledger. Live-version erasure additionally has successor/deletion rules; its ordinary successor Event is deliberately a content-free root replacement where appropriate, so the protocol already considers delta reconstruction. [Live erasure][erasure] [Persistent consumer duty][persistent]

An event-first protocol can carry erasure records in the same canonical framed stream, plus a durable ledger in bootstrap snapshots. There is no inherent requirement for a second independently authored log.

## Minimal event-first architecture

**Retain the existing snapshot and transaction envelope; make its single ordered transition list sufficient for reconstruction, and derive other views from it.** This is a design option requiring a specification change, not a claim that today's Event endpoint already provides it.

1. **One committed frame.** Keep Scope epoch, Authorization View, position, previous position, durable checkpoint and optional transaction identity. A frame is delivered/applied completely; checkpoint persistence is atomic with application. An empty projection-advance frame can represent hidden work without leaking identifiers. Keep existing expiry, reconnect and view/epoch reset rules.
2. **Reconstruction-complete entries.** Existing lifecycle entries remain the normal path: complete creation record; revision-guarded committed delta; identity-only deletion. Add explicit projection entry/exit semantics, with a full canonical entry image and a removal identity. Those facts describe view membership, not false resource creation/deletion. Alternatively, specify reset on every membership change.
3. **Metadata at the right point.** Include initial owned-set keys in creation/entry images, or require a version-stable metadata resolution contract. Continue carrying version attribution/changeContext. Do not add request guards, labels or before-images merely for replication.
4. **Erasure and progress in the same stream.** Include erasure instructions and existing fences; bootstrap carries the accumulated ledger. Resource Event Sources become filtered convenience views of the canonical frames, with their narrower guarantees stated honestly.
5. **One source of truth.** Define one externally observable commit contract, with deterministic equivalence between event-folded state and reads/snapshots. A provider may update relational state and write the corresponding committed transition record atomically in one transaction; event-sourced storage internals are not required. Notification projections and optional postimages must agree with that commit, rather than forming another separately accepted mutation history.

Existing BDP already says Events and state changes live in **one atomically committed change group**; it does not require two independent logs. The design choice is therefore whether each group transmits both ordered deltas and normalized postimages, or whether the canonical ordered list is sufficient and the postimage representation becomes derived/optional. Removing wire duplication can reduce bandwidth for small patches to large records; complete postimages simplify consumer recovery and normalization. Neither choice should be defended as mathematically necessary without measuring the workloads. [Groups][groups]

## Recommendation and evidence limits

Reframe the conclusion as: **“The Event payloads can drive replication. We need to finish the stream's replica contract, chiefly projected membership, bootstrap alignment, transaction completion, progress and erasure.”** Current BDP's complete groups provide those guarantees today in the specification; an event-first revision can preserve them without independent duplicated truth.

For qualification, require snapshot + event-fold equivalence across ordinary mutations, owned-Link transitions, multiple writes to one resource, membership entry/exit, truncated delivery/reconnect, view/epoch reset and erasure. Also test atomic publication and quiet-stream completion. No new runtime tests were run for this reassessment. The prior eight fixture probes and 158 artifact tests remain limited evidence; they are not Transactional provider conformance.

[events]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1701-L1797
[payload]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1816-L1877
[groups]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1942-L2034
[event-source]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L6127-L6268
[history]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1198-L1245
[revisions]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L697-L742
[membership]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L6037-L6050
[authorization]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L912-L941
[snapshots]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L2043-L2065
[feed]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L6092-L6117
[erasure]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L5800-L5856
[persistent]: https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L6219-L6234
