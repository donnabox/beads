# Worked model: five generic event verbs

2026-10-09. Proposed worked model for Donna's D03 ruling: zero domain-specific event types. This is a design example, not an accepted wire amendment or implementation instruction. The five verbs are BDP's existing `created`, `updated`, `deleted`, `linked`, and `unlinked`. Examples provisionally put comments in the parent Bead's logical properties; that data-model choice remains open.

## Meaning of each verb

| Verb | Meaning | Subject/payload |
|---|---|---|
| `created` | A Bead or Link began to exist. | Resource identity and Type, initial properties/revision; Link endpoints when applicable. |
| `updated` | Existing Bead or Link state changed. | Resource identity/Type, predecessor and successor revisions, committed property change or an owned-Link transition. |
| `deleted` | A Bead or Link ceased to exist. | Identity/Type and final live revision; no old body required to remove current state. |
| `linked` | A Link became incident upon an in-Scope Bead. | Link identity/Type, endpoints and endpoint role. |
| `unlinked` | A Link ceased to be incident upon an in-Scope Bead. | Link identity/Type, endpoints and endpoint role. |

`created`/`updated`/`deleted` apply equally to Beads and Links. `linked`/`unlinked` let a subscriber to a Bead observe its incident relationships. They are induced facts from the same operation, not additional mutation commands to replay. Their Event subject is the Link; the endpoint-specific Event Source/context tells the consumer where it became incident. A domain's Bead/Link Types and property names remain data; the event verb never becomes `issue_closed`, `comment_added`, or `memory_updated`.

This follows the current [generic Event definitions and Link expansion](https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1701-L1940). Ordinary incident facts do not advance endpoint Bead revisions. BDP owned Links additionally advance their owning source and emit its generic `updated` fact; no new verb is needed. That existing rule is preserved without deciding Trish's Type design here.

## Mapping application actions

| Application action | Generic committed fact(s) | Where the domain meaning lives |
|---|---|---|
| Create issue, memory, document or person | Bead `created` | Bead Type and initial properties. Legacy KV-backed memories first need a graph representation. |
| Edit text, priority, labels or metadata | Bead `updated` | Changed property paths and values. |
| Close, reopen, defer, wake, assign or release work | Bead `updated` | Status/assignment/time properties chosen by the pack. Lease Links, if chosen, use ordinary Link lifecycle too. |
| Add a structured comment | Parent Bead `updated` | Append a comment object containing its existing identity, author, text and time. |
| Record an audit note | Parent Bead `updated` if the note is retained in its logical annotation collection | Provenance such as `source: audit` is data, not an event kind. A separate audit resource model is another generic alternative. |
| Edit/delete a comment, if supported | Parent Bead `updated` | Change/remove the selected comment object; not deletion of the parent. No new comment-edit API is implied. |
| Add dependency/relationship | Link `created` and `linked` at each in-Scope endpoint | Link Type, endpoints and properties. |
| Change dependency metadata | Link `updated` | Metadata/properties change; no new incidence, so no `linked`/`unlinked`. |
| Remove dependency/relationship | Link `deleted` and `unlinked` at each in-Scope endpoint | Link identity and endpoints suffice. |
| Change a relationship's endpoints or immutable Type | Delete old Link/create new Link under current BDP | Corresponding lifecycle and endpoint facts; not a property update of immutable structure. |
| Persist a changed readiness projection | Bead `updated` | `is_blocked` or other chosen property. If computed only on read, there is no stored change or event. |
| Delete a Bead | Bead `deleted`, plus facts for any Links actually deleted in the transaction | The deletion policy decides which graph mutations are allowed/required; events report their effects. |
| Compact/restore text or import data through the authority | `updated`, `created`, or `deleted` for the actual Resource transitions, plus Link facts as needed | Operation provenance may be data. Bypass writes must be captured or invalidate continuity. |
| Bulk operation | The same singleton facts for each affected Resource, grouped by committed transaction | No `batch`, `bulk_close` or `import` event verb. |
| Failed transaction or semantic no-op | No state-change Events | Failure/no-effect outcomes belong to responses/receipts. Same-value metadata refresh should not invent a Link transition. |

These are model mappings, not assertions that every current command is already implemented over generic BDP. Identity rename, transient heartbeat state, side collections, and backend-native administration retain their migration decisions. A display-name update is a property change; changing a canonical nonreusable identity cannot masquerade as one.

## Comments as an embedded collection

A possible parent property is:

```json
{
  "comments": [
    {
      "id": "c123",
      "author": "Donna",
      "text": "Recheck this after the rollout.",
      "created_at": "2026-10-09T23:00:00Z",
      "source": "structured"
    }
  ]
}
```

Appending this object could produce a generic committed Property Change:

```json
[
  {
    "op": "add",
    "path": "/comments/-",
    "value": {
      "id": "c124",
      "author": "Chris",
      "text": "Confirmed against the new deployment.",
      "created_at": "2026-10-09T23:05:00Z",
      "source": "structured"
    }
  }
]
```

The Event is `updated`, subject is the parent Bead, and its revision advances. The patch does not repeat all previous comments. A receiver applies it only at the expected predecessor revision and persists its replay checkpoint with state. Replaying the append blindly twice duplicates the comment; generic stream deduplication and mutation-request idempotency remain necessary. Stable comment IDs also preserve identity for import, selection and future edits, independently of array position.

Physical storage can remain separate tables. Logical reads, snapshots and events must expose the same collection state; a page fetched later without a pinned revision is not automatically part of the earlier snapshot. Large threads, bounded append, parent contention and revision-bound pagination need qualification before choosing this model. Making comments separate Beads/Links is an alternative if those costs justify it, not required for generic events.

## Relationship examples

For an ordinary Link L from A to B, with both endpoints in Scope:

```text
create L:
  created   subject=L
  linked    subject=L, endpoint=source, observed at A
  linked    subject=L, endpoint=target, observed at B

change L.metadata:
  updated   subject=L

remove L:
  deleted   subject=L
  unlinked  subject=L, endpoint=source, observed at A
  unlinked  subject=L, endpoint=target, observed at B
```

All facts induced by one operation belong to the same transaction. A replica inserts/patches/removes L from its lifecycle facts and uses endpoint facts for adjacency/observation; it must not create/delete the Link twice. A self-Link still has two endpoint-role facts. An out-of-Scope endpoint does not get a Bead-scoped Event Source. Owned Links add the existing source-state update described above.

## Transaction example

Suppose one atomic batch closes issue A, appends a comment to A, and changes metadata on existing Link L. With the inline-comment model and an ordinary (not owned) Link:

```text
one complete transaction group, position P, previousPosition Q
  ordinal 0: updated A, revision a7 -> a8, status/closure property patch
  ordinal 1: updated A, revision a8 -> a9, append comment object
  ordinal 2: updated L, revision l3 -> l4, metadata property patch
end of group
```

There is no `close`, `comment`, or `dep_updated` Event. A delta consumer stages all three and publishes them atomically. Under current BDP, the group's `changes` additionally contains A's final a9 postimage and L's final l4 postimage: two normalized state entries, three ordered Events. A replica using `changes` must not then reapply the Event patches to the already-final state. Whether postimages become optional/derived remains open in the event-first proposal.

Multiple commands need not equal multiple transactions, and one legacy command can commit more than one transaction (including a later derived-state settlement). The group follows the actual authority commit boundary, not an inferred command boundary. [Current group normalization and framing](https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L1942-L2034)

## Five graph verbs do not remove recovery requirements

- **Transactions** are complete group envelopes around Events; no transaction event verb is necessary.
- **Epochs and checkpoints** identify the history and replay position in the envelope. A reset rejects old continuation and requires a new baseline; it is not a domain mutation event.
- **Snapshots** install canonical state at a known checkpoint, then replay later groups. They are not synthetic creation events for resources that already existed.
- **Projection changes and erasure** still need the current change-group/snapshot contract. A Resource can leave a view without being deleted, so do not synthesize a false `deleted` fact. With exactly these five graph verbs, retain generic projected-state changes, erasure records and progress in the envelope, or explicitly choose reset/resnapshot for unsupported transitions. Five verb names alone do not make individual Event Sources sufficient for every replica obligation.

The worked model deliberately keeps current BDP's group `changes`/`erasures` and snapshot rules while examining the mutation vocabulary. Folding every recovery/projection transition into one event-only stream remains a separate design decision. D03 rules out domain-specific event types; it does not require falsifying graph lifecycle facts or dropping generic replication control information. [Snapshot contract](https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L2043-L2065) [Projection and Event visibility](https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L6037-L6234)

No runtime implementation, tests or normative protocol changes accompany this model.
