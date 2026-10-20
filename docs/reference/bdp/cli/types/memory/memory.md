---
title: "Memory Type CLI facade"
description: "Candidate contract for using the selected Memory Types through existing bd commands."
---

Memory Types let you store constraints and assessments alongside ordinary work.
This facade describes how the selected Types fit the existing `bd` commands.

## Status and scope

Status: **draft CLI contract for the Preview 3 candidate**. The selected Type
semantics are accepted design. The complete command integration and its
qualification are pending; these examples are acceptance cases, not transcripts
from a released build.

The inspected integration source is
[Beads `5b8cdefe`](https://github.com/donnabox/beads/commit/5b8cdefe055b8f08cf9eb385e946396c1b5799d7).
The new catalog, validation and retyping candidate is separate from that source.
This page is a preview specification and does not describe the pinned stable
release. It does not change the documentation release pin.

This facade covers `bd types`, typed creation and updates, showing, listing,
Type filtering, deletion, and existing Link commands. This facade introduces no
separate `bd memory` command. Query and list design remains Melinda's separate
lane. A `bd type` management tree is not required for this facade.

| Area | Candidate status | Remaining acceptance |
| --- | --- | --- |
| Selected catalog and reusable validation | Authored; focused pure checks passed on an exact candidate | Installed provider and combined product checks |
| Properties, Memory-family retyping, ownership and exact Evaluation subjects | Authored; selected storage checks passed | Complete focused embedded and server cases |
| `bd types` provider-driven inventory | Authored separately | Combined installed CLI and workspace checks |
| Typed create/update/show/list/filter/delete routing | Integration work pending | Command cases below on the qualified combined source |
| Integration and release | Pending | Reviewed exact source, required checks and approved integration landing |

These states describe the candidate, not feature availability in the integration
source above. The [release ledger](https://github.com/gastownhall/beads/issues/7656)
records exact qualification and landing evidence.

## Requires

This CLI facade requires the
[Memory Bead and Link Type specification at BDP `bbd97f4`](https://github.com/gastownhall/bdp/blob/bbd97f4b57ddbcea8ee5ca38adc2ec7d0a784f00/specs/types/memory/memory.md).
The exact source revision is
`bbd97f4b57ddbcea8ee5ca38adc2ec7d0a784f00`; its SHA256 is
`25905e496f36ab7f95046c5c98ed681aeb9aa0bb1c832d63beb6038120bd0ac5`.
That specification defines the Type meanings, properties, endpoint restrictions
and ownership rules. This page defines the Beads command mapping.

The dependency is the whole typedef specification. It does not implicitly
require the BDP Types profile or introduce a protocol change. Future source
changes require deliberate reconciliation and qualification; following a moving
branch does not change the contract tested for this candidate.

## Type inventory and selection

`bd types`, `bd types --details` and `bd types --json` enumerate the selected
workspace's public installed Types. They use the same visibility rules.
Existing Issue classifications and Dependency behavior remain available.

The selected Memory Bead Types are Memory, Constraint, Policy, Decision,
Procedure, Requirement, Scenario, Specification and Evaluation. The selected
Link Types are References, Superseded By, Requires, Defines, Subject To,
Conformance, Addresses, Satisfies and Violates. Conformance is abstract: it can
appear in the inventory, but cannot be used to create a
direct unspecialized Link.

Type, BeadType, LinkType, ConformsTo and Refines are intrinsic implementation
Types. They are absent from public inventories and invalid as user Type
selections or filters. Active `example-*` and `preview-*` registrations are
removed rather than merely hidden from the printed list.

Use the Type selector reported for the selected workspace. Examples below use
variables for those values; they do not prescribe canonical Type identities or
construct a Type URL from its display name. In particular, select the nominal
Memory Decision Type explicitly when distinguishing it from an existing Issue
classification named `decision`.

`--details` may show a partial Type description. The Preview 3 `bd types --json`
shape has no Preview 3-to-Preview 4 compatibility guarantee. That exception
applies to this inventory output only.

## Creation and property editing

`bd create --type TYPE --properties JSON` selects a concrete installed Bead
Type and validates its complete properties. Type-specific properties survive
storage, subsequent edits and readback. Memory titles are nonblank; fields with
other required shapes follow the typedef specification. Native Issue fields
retain their existing domain validation.

For Memory-family records, `bd update ID --properties JSON` edits the supplied
top-level members and preserves omitted properties. `--replace-properties`
supplies a complete replacement properties object; `--patch` applies the
existing ordered property-patch format. Their existing exclusivity rules still
apply. All forms validate the complete result before publication.

Unsupported authored `applicability` values are refused in this candidate. The
shared spec reserves the concept without selecting an executable value format
or evaluator. This limitation does not define new predicate syntax.

## Retyping through ordinary update

`bd update ID --type TYPE` is the command mapping for explicit Type changes.
Omitting a properties edit preserves all current properties. Explicit property
edits may accompany the change; the complete resulting record must be valid.
No values are automatically converted or dropped.

Retyping uses the command's ordinary revision-guard behavior. This facade adds
no separate mandatory retyping guard. An explicit stale guard refuses the
operation. Examples use `--if-revision` so their intended predecessor is clear.

The complete prospective graph must remain valid. If the Type change makes an
existing incoming or outgoing Link invalid, the operation refuses atomically:
no Type, property, Link, owned snapshot or retained history change is published.
Repair the Links separately, then retry against the current revision.

Earlier retained revisions keep their original declared Types and properties.
They are not reinterpreted using the current Type. A new Type selection must
also preserve the selected ownership and final-head integrity rules.

The current implementation assignment covers Memory-family transitions. Generic
retyping is the broader design intent; conversion between native Issue backing
and generic Memory backing is not implemented by this candidate.

## Evaluation subjects and claims

An Evaluation is a Memory subtype representing an assessment. Its `subjects`
property is a nonempty array of exact Bead references, each with `uri` and
`revision`. Floating subject references are refused. The current candidate
admits canonical local Bead identities with an available retained revision;
foreign-subject admission is not implemented.

The Evaluation owns its Conformance-family claims. Each claim concerns the
Evaluation's whole subject group collectively. Adding a claim revises its
Evaluation owner, while assessed subjects and the target Constraint remain
unchanged. Claims do not select per-claim subsets or automatically distribute
a group's satisfaction claim to each member.

Use the existing Link commands with the selected concrete Link Type. A plain
Memory or Issue is not a substitute source for an Evaluation-owned claim.
Subject To remains available for native Issues with the selected source
ownership behavior; ordinary Issue payload edits keep native validation.

## Showing, listing, filtering and deletion

`bd show` and the existing list/Type-filter surfaces must expose the selected
Memory subtypes without collapsing them to one Memory shape. Reads preserve
the authored properties and actual declared Type. Type filters use the
selected command's nominal Type semantics, with intrinsic selections refused.
This facade does not define a new query grammar or change Issue workflow
classification.

`bd delete` and the Memory-only `bd forget` retain their existing preview,
apply and revision-guard rules. Deletion refuses incident Links and does not
cascade through them. Retained history remains available under the existing
history contract; `forget` is not retained-version erasure. Special Evaluation
cleanup and automatic cascade behavior are not selected here.

Use a workspace initialized for the matching candidate catalog. Automatic
conversion of an earlier graph catalog is not supplied. An incompatible
workspace must be refused explicitly rather than silently adopting new Type
definitions or reinterpreting existing records.

## Acceptance examples

Run these against a disposable workspace using the combined candidate build.
Set `POLICY_TYPE`, `SPECIFICATION_TYPE` and `EVALUATION_TYPE`
from its public Type inventory. Set `RULE`, `RULE_REV`,
`SUBJECT` and `SUBJECT_REV` from actual records returned by that workspace.
These variables stand for observed identities and revision tokens.

```sh
bd types
bd types --details
bd types --json
```

All three formats expose the same public selected Type set. Intrinsics and
active example/preview Types are absent. Abstract Conformance is distinguishable
from concrete claim Types.

```sh
bd create --type "$POLICY_TYPE" \
  --properties '{"title":"Retain evidence","body":"Keep the run report","status":"draft"}'
bd update "$RULE" --type "$SPECIFICATION_TYPE" --if-revision "$RULE_REV"
bd show "$RULE" --json
```

A valid retype preserves the authored properties and identity, produces the
new declared Type and revision, and leaves the earlier exact revision intact.
An existing Link that would become invalid instead causes complete refusal.

```sh
bd create --type "$EVALUATION_TYPE" --properties "$EXACT_EVALUATION_PROPERTIES"
```

`EXACT_EVALUATION_PROPERTIES` is an object with a nonblank title, a body and
`subjects` containing the observed `SUBJECT` and `SUBJECT_REV` as
`{"uri": SUBJECT, "revision": SUBJECT_REV}`. Construct valid JSON from those
values. Omitting a revision, selecting an unknown/unavailable revision, or
using an empty subject list must refuse without creating an Evaluation.

| Case | Observable acceptance |
| --- | --- |
| Edit one property, leaving an extension object untouched | The extension survives current and retained readback without numeric conversion |
| Retype and explicitly edit properties together | Admit the complete valid result or publish no change |
| Supply a stale explicit revision guard | Refuse without overwriting the current record |
| Create an abstract Conformance Link or a claim from an ordinary Memory | Refuse without revising any owner or endpoint |
| Create a valid Evaluation-owned Satisfies Link | Revise only the Link and its Evaluation owner; preserve the assessed group and target |
| Exceed a selected Link-family limit, create a forbidden cycle or identity self-Link | Refuse atomically according to the typedef rules |
| Edit an ordinary Issue in a mixed workspace | Preserve native Issue semantics without unrelated Memory payload admission or budget scans |
| Use an incompatible earlier catalog | Refuse explicitly; do not migrate or reinterpret it implicitly |

Qualification must cover installed commands on both embedded and fresh server
workspaces, retained readback, refusal atomicity, and the exact combined source.
Pure validation checks alone do not close this facade's command acceptance.
