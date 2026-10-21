#### Alias targets

The Read+Update profile defines two alias targets, `put-alias` and
`delete-alias`, keyed by alias path beneath the fixed `alias/` root defined
under [Aliases](#aliases); the Transactional profile inherits both.
Each is a singleton target under
[Operation Directory and singleton targets](#operation-directory-and-singleton-targets)
and a sequence member with its own `idempotencyKey`; the operation
discriminators are `putAlias` and `deleteAlias`.

A **put** creates an alias or repoints an existing one to exactly one
canonical in-Scope Bead URL; repointing is the same operation, not a second
one. Its record carries `alias`, the alias named by its local spelling
`alias/{alias-path}` or by its absolute alias URL, and `target`, one
canonical in-Scope Bead reference — a durable local ID, an absolute canonical
Bead URL, or, in a sequence, a `@name` bound by an earlier creation of a Bead
in the same sequence. A **delete** removes the alias; its record carries
`alias` alone, and the path is reusable afterwards, as [Aliases](#aliases)
says. Neither record carries `expectedRevision`, since an alias has no
revision, nor `revisionProvenance`, since attribution is per version and
alias mutation mints none; neither binds a `name`, since neither creates a
Resource.

```json
{
  "alias": "alias/adr/sequence-envelopes",
  "target": "beads/decision-7"
}
```

The alias path uses the local-ID grammar: one or more safe segments,
compared exactly as local-ID segments are. `alias` is resolved against the
canonical Scope URL like a durable reference, and it never accepts
`@name`; a spelling that violates the grammar is carrier syntax rejected
before execution with `malformed-request`, and a well-formed spelling
that is not beneath the `alias/` root names no alias and fails with
`resource-not-found` when the member is reached, as a wrong-root subject
reference does under [Read+Update problem details](#readupdate-problem-details). Canonical Bead
segments and alias paths share one
uniqueness namespace in the Scope, a store invariant the authority
enforces when the member is reached: a put whose alias path is the
`{id-path}` of a canonical Bead URL ever committed in the logical Scope —
a deleted one included, since canonical segments are never released —
fails with `identity-taken`, and a Bead creation that supplies an `id`
whose `{id-path}` is a live alias path fails with `alias-path-taken`, a
condition that clears when the alias is deleted. An authority never
allocates a Bead id whose segment is a live alias path. Link segments and
alias paths coexist: `links/foo` and `alias/foo` do not collide, and a
Link creation may supply an `id` whose `{id-path}` is a live alias path.
An alias path in use as an alias is not taken for a put, which repoints
it. Alias operations follow the model's check order under
[Explicit alias operations](#explicit-alias-operations), identifier
uniqueness first, so a put whose path is taken and whose target is
unknown answers the uniqueness fault.
A put whose `target` is a canonical Bead reference naming a Bead that does
not exist or is not visible in the request's Authorization View fails with
`resource-not-found`, and a delete of an unknown alias fails with
`resource-not-found`: aliases are not an enumeration oracle. A put whose
`target` is not a canonical Bead reference — an alias, absolute or local,
a Link, or an external URI — fails with `validation-failed`, carrying one
diagnostic that names the cause: an alias targets a canonical Bead URL
only, so no chain is ever admitted, exactly as
[Alias resolution](#alias-resolution) never follows one. A put or delete
the principal may not perform fails with `forbidden`. Alias operations
are authorized as mutations of the Beads they touch: a put requires that
the principal may write the proposed target Bead, and a repoint or a
delete additionally that it may write the alias's current target; when
the current target is not visible to the principal, the alias itself is
`resource-not-found`, disclosing nothing.
These alias mutation authorization rules are distinct from retry identity
comparison within a principal's key namespace under
[Idempotency keys](#idempotency-keys); comparison grants no permission to
perform an alias mutation or disclose a retained Resource.

Alias mutation mints no version: an alias is a locator, not part of any
Bead's durable state, so the target Bead's revision is unchanged by a put
or a delete, and a repoint changes the revision of neither the former nor
the new target. Aliases are not members of the Bead record or of its
`properties`, carry no revision, and are not Resources.

Every successful alias mutation produces one **alias result**, closed:

```text
AliasResult {
  outcome: created | updated | deleted
  alias             // the absolute alias URL
  target?           // created, updated: the absolute canonical target Bead URL
}
```

A put reports `created` when the alias path was not in use as an alias and
`updated` when it repointed an existing alias — including a put whose
`target` the alias already had, which changes nothing and reports
`updated`; both carry `alias`, the absolute alias URL, `alias/{alias-path}`
resolved against the canonical Scope URL, and `target`, the absolute
canonical Bead URL the alias now resolves to. A delete reports `deleted`
with `alias` and no `target`. The outcome vocabulary is the mutation
result's own. A singleton alias target returns the alias result as the
body of a `200 OK` response, exactly as a mutation result is returned
under
[Operation Directory and singleton targets](#operation-directory-and-singleton-targets);
a sequence member's entry is the alias result plus `operationIndex`, and
never `operationName`. The bundle defines `aliasResult`,
`putAliasRequest`, `deleteAliasRequest`, `sequencePutAlias`,
`sequenceDeleteAlias`, and `sequenceMemberAliasResult`.

Alias members are sequence members under every rule of this section: keys
are claimed at admission, carrier discipline and static reference checks
apply unchanged, and a put's `target` may name a `@name` bound by an
earlier Bead creation in the same sequence — creating the Bead and then
binding its alias is one sequence — resolving to the identity the
creation allocated, fresh, retained, or expired. A `@name` bound by a Link
creation is of the wrong Resource kind and is rejected before execution.
An alias member's semantic identity is its operation kind plus its
normalized record, under [Idempotency keys](#idempotency-keys): `alias`
canonicalized to the absolute alias URL and `target` resolved to the
canonical Bead URL or to the identity its creating member bound. The same
key with the same semantic identity returns the retained disposition, as
for every singleton, and the same key with a different identity is
`idempotency-conflict`. An alias disposition is retained, replayed, and
tombstoned exactly as a Resource mutation's is: a put or delete commits
state, so its tombstone outlives the retention interval under
[Outcome retention](#outcome-retention), and an alias result discloses no
Resource record, so it is returned as retained, as a `deleted` identity
is. Because every Read+Update Scope offers the alias targets, a
Read+Update authority serves alias resolution and advertises `aliases` in
its discovery document; `aliases` is therefore a required member of the
Read+Update and Transactional discovery documents under
[Scope discovery and human documentation](#scope-discovery-and-human-documentation),
and the bundle's `readUpdateDiscovery` requires it.

An alias spelling — `alias/{alias-path}` or the absolute alias URL — is
admitted wherever a canonical in-Scope Bead reference is: as the `bead`
subject of an update or a deletion and as a Link endpoint `source` or
`target`, bare or as the `uri` of a Pinned Reference, in a singleton and
in a sequence member alike. The authority resolves it to the alias's
current target when the member is reached — exactly as a `@name` binding
is resolved when its member is reached, so an alias put earlier in the
same sequence is what a later member observes — and stores and serves the
canonical Bead URL, as [Aliases](#aliases) requires: a reference resolved
through an alias does not follow a later repoint. An alias spelling that
names no live alias fails with `resource-not-found`, under the same
non-disclosure rule as an unknown Bead, and a `link` subject spelled by
alias is of the wrong kind and fails with `resource-not-found` too, since
an alias resolves to a Bead only. A put's own `target` admits no alias
spelling: an alias target is `validation-failed`, as above. The semantic
identity of a member that spelled a reference by alias records the
resolution rather than the spelling, under
[Idempotency keys](#idempotency-keys).

