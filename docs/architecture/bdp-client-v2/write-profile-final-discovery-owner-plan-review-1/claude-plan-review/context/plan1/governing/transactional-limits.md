### Transactional limits

- `transaction.operations`, `transaction.examinedResources`,
  `transaction.matchedResources`, `transaction.mutatedResources`, and
  `transaction.inducedEvents` are counts, while `transaction.duration` is an
  ISO 8601 duration; and
- `retention.idempotency`, `retention.receipt`,
  `retention.maximumSnapshotLifetime`, and `retention.replay` are ISO 8601
  durations.

`retention.idempotency` is the minimum interval for which an authority
retains an idempotency-key disposition after its terminal outcome; the
Read+Update profile binds it under [Outcome retention](#outcome-retention).
The bundle's profile discovery definitions admit only the groups a profile
exposes: the Read discovery document's `limits` is `advertisedLimits`,
which admits no `validation` group; the Read+Update discovery document's
`limits` is `readUpdateAdvertisedLimits`, a closed definition of its own
that shares every limit primitive with `advertisedLimits`, restates the
`page`, `request`, `resource`, `selector`, `patch`, and `sequence` groups
unchanged, carries the `validation` group, and rejects the `transaction`
group and the Transactional `retention.receipt` and `retention.replay`
members, while keeping `retention.idempotency` and the pagination
`retention.maximumSnapshotLifetime`. The Transactional discovery
document's `limits` is `transactionalAdvertisedLimits`, a closed
definition of its own on the same primitives: it carries the `validation`
group as well, since a Transactional authority advertises the same bound,
admits the `transaction` group and the `retention.receipt`,
`retention.maximumSnapshotLifetime`, and `retention.replay` members, and
rejects `retention.idempotency` under the paragraph below.

On a Transactional Scope a key's disposition is retained by its receipt — a
completed transaction's for the rest of the epoch and a failed
transaction's for at least `retention.receipt` — so `retention.idempotency`
is a Read+Update-only member: a Transactional discovery document MUST NOT
advertise it, `retention.receipt` bounds how long detailed outcomes and
failed receipts remain, and the bundle's `transactionalAdvertisedLimits`
rejects the member. `page.defaultItems` and `page.maximumItems` also count
the result entries of a Mutation Receipt and its pages, every entry
counting as one.

