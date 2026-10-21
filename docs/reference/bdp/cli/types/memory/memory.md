---
title: "Memory CLI specification"
description: "Preview 3 design for Memory relationships, requirements and explicit applicability."
---

Use Memory relationships to find related knowledge, follow replacements, and
read the requirements governing work. The proposed `bd memory` command group
provides those focused operations over the shared Bead and Link graph.

This is an implementation-pending design specification. The availability
notice below states the target policy; implementation and qualification are
tracked separately.

## Preview 3 availability

Assessment authoring is disabled by default in Preview 3. `bd types` omits
Evaluation, Conformance, Addresses, Satisfies and Violates while the switch is
off. Creating or editing those objects, restoring Evaluation content from a
retained revision, recovering a deleted Evaluation, and changing claim endpoint
pins return capability unavailable. An existing object or installed definition
does not bypass this restriction.

The definitions, identities and retained data remain intact. Existing objects
can still be inspected and deleted; valid retyping out of Evaluation follows
the ordinary incident-Link constraints. Other Memory Types, Constraint/Requires,
Memory body editing, and Issue/Dependency behavior remain available.

The existing workspace metadata configuration accepts
`"graph_assessments_enabled": true` to enable the retained assessment model for
controlled experimental use. An omitted or false value keeps it off. The same
setting controls Type availability and authoring; it changes neither the stored
Type installation nor the workspace schema/generation. The assessment examples
below describe that enabled model, rather than the default Preview 3 surface.


## Status and scope

Status: **accepted CLI design; implementation and qualification pending**.
The command names and semantics below are the Preview 3 design target selected
on October 21, 2026. They are not an inventory of commands in an installed
build. Every invocation and output example on this page is a proposed
acceptance example, not a captured execution.

This page extends the existing generic Type and CRUD facade. It introduces
focused `bd memory` operations without replacing `bd create`, `bd update`,
`bd show`, `bd list`, `bd delete` or `bd link` operations. Existing
`remember`, `recall`, `memories` and `forget` retain their separate contracts.
This design adds no public query language or Type-management command tree.

Preview 3 focuses on Memory relationships and explicit obligations.
Assessment execution and an assessment CLI are deferred. The Evaluation and
Conformance-family design remains preserved separately below; its presence
is not a claim of Preview 3 authoring availability. Availability policy and
its runtime qualification are tracked separately from this command design.

The [historical candidate ledger](#historical-candidate-ledger) retains the
previous facade's exact source and evidence scope. It does not qualify the
new commands or presentation. This page is a preview specification, not the
pinned stable release documentation; it does not change the release pin.

## Requires

This CLI facade requires the
[Memory Bead and Link Type specification at BDP `afec9117`](https://github.com/gastownhall/bdp/blob/afec9117914bd1c1fe0027c2067e4d89677e40a3/specs/types/memory/memory.md).
The exact source revision is
`afec9117914bd1c1fe0027c2067e4d89677e40a3`; its SHA256 is
`1f97f8a80f725b2a44a0c715b4ccf09918f02af2abb9885797c3395255dc9e34`.
That specification defines the Type meanings, properties, endpoint restrictions
and ownership rules. This page defines the Beads command mapping.

The dependency is the whole typedef specification. It does not implicitly
require the BDP Types profile or introduce a protocol change. Future source
changes require deliberate reconciliation and qualification; following a moving
branch does not change the contract tested for this candidate.

## Read commands

Each read command accepts one or more roots. All roots and their results come
from one consistent, bounded read view. Each root has its own hop budget.
Calculate each root's graph in that view, then aggregate those graphs; do not
re-induce extra Links solely across the union of reached Beads. Results retain
the roots, their membership, frontiers and supporting relationship paths. A shared
Constraint is one Bead; discovering it from two subjects does not transfer
one subject's obligations to the other.

| Invocation | Link selection | Direction | Default depth |
| --- | --- | --- | --- |
| `bd memory successor ID...` | Superseded By family | Outgoing | 1000 |
| `bd memory supersedes ID...` | Superseded By family | Incoming | 1 |
| `bd memory references ID...` | References family | Outgoing | 1 |
| `bd memory referenced-by ID...` | References family | Incoming | 1 |
| `bd memory requirements ID...` | Requires family, including Defines | Outgoing | 1000 |
| `bd memory required-by ID...` | Requires family, including Defines | Incoming | 1 |
| `bd memory definition ID...` | Defines family | Outgoing | 1000 |
| `bd memory defined-by ID...` | Defines family | Incoming | 1 |
| `bd memory constraints SUBJECT...` | Subject To once, then Requires family | Staged | Bounded closure |
| `bd memory subjects CONSTRAINT...` | Subject To family | Incoming | 1 |

`ID...`, `SUBJECT...` and `CONSTRAINT...` mean one or more arguments, not literal
selector syntax. Each command's roots must meet its Type contract. Successor
and supersedes roots are Memories. Requirements, required-by, definition,
defined-by and subjects roots are Constraints. References roots and Subject
To sources can be any eligible Beads, including native Issues. Endpoint eligibility still requires actual
support in the selected workspace.

Every family selection includes refinements. Display the actual declared
Types, not the selected base Type. A depth of 1000 is a bound, not a promise
of unlimited traversal. `--depth` adjusts ordinary family traversal extent. `subjects` remains a
direct declaration lookup; deeper Subject To traversal is not part of its
initial contract. The staged query's stage-specific depth controls and the
remaining flag combinations are not selected by this document.

Ordinary family traversals use the shared graph semantics: reach Beads in the
selected direction and depth, then retain all matching Links between reached
Beads. This includes cross-links, parallel Links and cycle-closing Links within
each root's result. For a diamond A-to-B, A-to-C, B-to-D, C-to-D, querying
roots A and B at depth one includes AB, AC and BD, but not CD: neither
individual root reaches both C and D within its own budget.
The staged `constraints` query and direct `subjects` lookup have the narrower
edge scopes described below. They do not admit unrelated Links solely because
their endpoints are present.

### References and succession

`references` and `referenced-by` show outgoing and incoming relationships.
The default is direct neighbors. Increasing depth expands the graph; cycles
terminate through visited-identity tracking and remain visible in the result.
Parallel Links retain their separate identities.

`successor` follows replacement relationships forward. It returns the
replacement graph, including intermediate Memories, rather than only the
terminal Memory. `supersedes` asks what the input directly replaced; deeper
reverse traversal may branch because multiple older Memories can share one
replacement. Superseded By permits at most one outgoing successor per Memory
and prohibits cycles. A reader must not mistake a traversal boundary for a
terminal successor.

All reads retain the starting roots when no selected relationship exists.
This is zero-hop inclusion, not a self-Link or a supersession claim.

Proposed output:

```text
$ bd memory successor original

original [memory] Original guidance
└── superseded-by → revised [memory] Revised guidance
    └── superseded-by → current [memory] Current guidance

3 Beads, 2 Links
```

```text
$ bd memory references alpha --depth 1000

alpha [memory] Alpha
└── references → beta [memory] Beta
    └── references → gamma [memory] Gamma
        └── references → alpha (cycle)

3 Beads, 3 Links
```

These examples elide display prefixes and Link IDs for readability. Canonical
identity is unchanged; the final Link-ID display controls remain deferred.

### Requirements and definitions

A Constraint states something that must be true or must be done when it
applies. Requires identifies another Constraint necessary to meeting it.
Defines refines Requires and additionally records the defining parent.
Meeting the referenced Constraints does not by itself establish that the
source Constraint's own conditions were met.

`requirements` follows both relationships. `definition` follows only Defines,
so it shows the defining subtree without external Requires dependencies.
The reverse commands start with direct relationships. These queries discover
structure and obligations; they do not decide satisfaction, approval or
compliance.

Proposed output:

```text
$ bd memory requirements privacy

privacy [policy] Protect customer data
├── defines → access [policy] Limit data access
│   ├── defines → authorize [requirement] Authorize every export
│   └── defines → denied [scenario] Reject an unauthorized export
└── defines → retention [policy] Keep data only as needed
    └── requires → erase [requirement] Remove expired exports

6 Beads, 5 Links
```

`bd memory definition privacy` returns the five Beads and four Defines Links,
excluding `erase`. The external requirement remains necessary in the
`requirements` result without becoming part of Privacy's defining subtree.
A Scenario selected through Requires or Defines is a Constraint, not an
optional illustration. Its actual Type label makes the distinction visible.

The proposed presentation places Defines children before other Requires
relationships at each Constraint. It keeps supporting paths intact rather
than moving all Scenarios or Requirements into separate global groups.
Optional grouped summaries, body display and detailed Link properties remain
future presentation choices.

### Explicit applicability and subjects

`constraints` begins with each original subject's explicit outgoing Subject
To declarations. It then follows outgoing Requires-family Links, including
Defines, from those target Constraints. Subject To is not transitive.

For example, an Issue Subject To a Specification inherits the Specification's
Requires dependencies. It does not inherit another Subject To declaration
made by that Specification. Nor does it inherit a defining parent's rules
merely because an applicable Policy is that parent's child.

Proposed output:

```text
$ bd memory constraints export-work

export-work [issue] Implement CSV export
└── subject-to → access [policy] Limit data access
    ├── defines → authorize [requirement] Authorize every export
    └── defines → denied [scenario] Reject an unauthorized export

4 Beads, 3 Links
Discovery: explicit Subject To declarations and their Requires dependencies
```

The result includes only Links admitted by that staged path: Subject To once
from an original subject, followed by Requires-family Links. It must not
include a secondary Subject To Link merely because both endpoints happened
to be reached. With multiple subjects, preserve the originating subject for
each admitted path and obligation, even when expansion shares records.

`subjects` reports only direct incoming Subject To declarations whose target
is the requested Constraint. It does not include secondary Subject To Links
between those subjects, repeatedly traverse Subject To, or perform reverse
Requires traversal to discover indirectly bound subjects. With multiple roots,
aggregate these per-root declaration graphs and preserve their membership.

Proposed output:

```text
$ bd memory subjects access

access [policy] Limit data access
└── subject-to ← export-work [issue] Implement CSV export

2 Beads, 1 Link
```

Preview 3 discovers explicit applicability and the obligations derived from
Requires/Defines. It does not discover every implicitly applicable Policy.
Executable applicability rules and their evaluation are deferred; an
`applicability` string or Policy body must not silently become a predicate.
No inferred stored Subject To Link or fabricated Link identity is created.

## Link authoring commands

| Invocation | Created relationship |
| --- | --- |
| `bd memory reference A --to B` | A References B |
| `bd memory supersede OLD --by NEW` | OLD Superseded By NEW |
| `bd memory requires A --to B` | A Requires B |
| `bd memory defines A --to B` | A Defines B |
| `bd memory subject-to SUBJECT --to CONSTRAINT` | SUBJECT Subject To CONSTRAINT |

Each command selects its concrete installed Link Type and uses the common
Link authoring operation. Source ownership, endpoint eligibility, arity,
cycle checks, atomicity and revision guards follow the Type and shared Link
contracts. These facades do not bypass validation or introduce a separate
allocator.

Known properties have named `--title`, `--body` and `--status` flags.
`--properties` supplies other declared properties and `--metadata` carries
informational metadata. A property supplied both through a named flag and
`--properties` is refused; there is no silent precedence. Omitted properties
remain absent. Status is an optional nonblank open string, not an automatic
workflow, dispatch, visibility or approval state.

Use shared `--id` and applicable revision guards. Shared `bd create` and
`bd link create` own automatic ID generation. Confirmation must include the
created Link identity. Exact CRUD confirmation formatting remains under the
common output contract, not this facade.

Proposed authoring example:

```text
$ bd memory reference design --to context --title "Background" --body "Reasoning behind the design" --status reviewed --metadata '{"importer":"migration-1"}'
Created reference-id: design → context
```

Repeated References creation without an explicit identity creates distinct
Links, including for equal endpoints. Explicit-ID retry and conflict behavior
follows the common contract; endpoint equality is not automatic deduplication.

`supersede` is create-only. If OLD already has an outgoing Superseded By-family
Link, refuse even when the requested NEW is the same. Do not retarget, replay
as a no-op, or silently replace a successor. Explicit changes use ordinary
guarded Link operations. OLD owns the new relationship; recording it revises
OLD's owned-Link state, not NEW. No automatic status change or deletion occurs.

Proposed confirmation, showing the required identity:

```text
$ bd memory supersede original --by revised --id replacement-id
Created replacement-id: original → revised
```

## Presentation, JSON and bounded results

The default presentation is the compact relationship tree shared
with generic `bd graph`; no `--graph` opt-in is required. See
[shared graph presentation](/reference/graph-cli-specification-draft#memory-and-generic-graph-presentation-target)
for cycles, shared descendants, cross-links, authored arrows and limits.
That is a target presentation, not the current candidate's flat Bead/Link
listing or the unchanged native Dependency renderer.

The [common selection and output contract](/reference/bdp/cli/selection-and-output)
owns canonical records, projections, JSON framing, completeness, errors and
writer behavior. This facade does not define a second writer, an independent
JSON schema version or private partial-result rules. Canonical JSON preserves
identities, Types, endpoints and graph context; human display prefix elision
must not rewrite those values.

Multi-root graph framing must be reconciled with the common output adapter
before implementation can claim support. Preserve explicit roots and
per-root relationship membership in both human and machine forms; silently
flattening those distinctions is not acceptable.

Reads disclose truncation and unresolved scope under the common contract.
A depth boundary cannot establish a terminal replacement, complete obligation
discovery, or satisfaction. Resource/acquisition budget exhaustion and an
invalid or unavailable requested root refuse the aggregate operation rather
than silently returning a successful partial result. Type-specific human labels
must not invent Issue readiness, execution layers or workflow status for
Memory records. Native `--graph-mode dependency` behavior remains unchanged.

## Generic Type and CRUD mapping

The following mapping remains shared generic CLI behavior. It is not a claim
that the new `bd memory` facades have shipped. Use Type selectors reported by
the selected workspace rather than constructing identities from display
names. Select the nominal Memory Decision Type explicitly when distinguishing
it from an existing Issue classification named `decision`.

`bd types`, `bd types --details` and `bd types --json` follow the selected
workspace's inventory and availability policy. Existing native Issue
classifications and Dependency behavior remain available. The intrinsic
Type, BeadType, LinkType, ConformsTo and Refines Types are absent from public
inventories and are not user authoring or filter selections. Active
`example-*` and `preview-*` nominal Type registrations are removed, while
existing configured native Issue classifications with those prefixes remain
valid. Inventory admission does not promise that an exact nominal Type filter
expands to refinements; the family reads above explicitly request family
semantics. `--details` may show a partial Type description.

Memory, Constraint, Policy, Decision, Procedure, Requirement, Scenario and
Specification form the non-assessment Bead catalog. References, Superseded By,
Requires, Defines and Subject To form the corresponding Link catalog.
Evaluation and Conformance-family definitions are retained as deferred design;
listing definitions or recognizing existing records does not imply permission
to author them.

The Preview 3 `bd types --json` inventory shape has no Preview 3-to-Preview 4
compatibility guarantee. That exception applies only to inventory output.

### Creation and property editing

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

### Retyping through ordinary update

`bd update ID --type TYPE` is the command mapping for explicit Type changes.
Omitting a properties edit preserves all current properties. Explicit property
edits may accompany the change; the complete resulting record must be valid.
No values are automatically converted or dropped.

Retyping uses the command's ordinary revision-guard behavior. This facade adds
no separate mandatory retyping guard. An explicit stale guard refuses the
operation. Use `--if-revision` to select an explicit predecessor.

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

### Showing, listing and deletion

Shared `bd show` and list/Type-filter surfaces preserve authored properties
and actual declared Types rather than collapsing Memory subtypes to a single
shape. This design does not introduce a query grammar or change native Issue
workflow classification.

Use the existing generic deletion, Link cleanup and revision-guard contract.
The Memory facade introduces no independent cascade, erasure or restore
policy. `forget` is not retained-version erasure. The historical incident-Link
refusal in the ledger's source is evidence about that source, not an override
of a separately selected deletion policy.

Use a workspace initialized for the matching catalog. This facade supplies
no automatic catalog conversion or migration; incompatible state must not
be silently reinterpreted.

## Deferred Evaluation and Conformance design

The following material is preserved from the earlier facade to keep the
selected Type semantics and prior candidate work reviewable. It is not part
of the Preview 3 Memory command surface. No `assess` or claim-authoring facade
is introduced here. Revisit the design using experience from practical agent
skills rather than treating this hierarchy as a required next-wave deliverable.

### Evaluation subjects and claims

An Evaluation is a Memory subtype representing an assessment. Its `subjects`
property is a nonempty array of exact Bead references, each with `uri` and
`revision`. Floating subject references are refused. The current candidate
admits canonical local Bead identities with an available retained revision;
foreign-subject admission is not implemented.

The first subject group recorded while an identity is an Evaluation remains
bound to that identity for its lifetime. Retyping it away from Evaluation and
back does not reset the group; a different group requires a new Evaluation.
An ordinary Memory extension named `subjects` before its first Evaluation
revision does not establish this binding. The exact typedef input above
includes this accepted lifetime rule. The [decision and implementation receipt](https://github.com/donnabox/agent-coordination/issues/1#issuecomment-6282549426)
records the ruling and its separately qualified product evidence.

The Evaluation owns its Conformance-family claims. Each claim concerns the
Evaluation's whole subject group collectively. Adding a claim revises its
Evaluation owner, while assessed subjects and the target Constraint remain
unchanged. Claims do not select per-claim subsets or automatically distribute
a group's satisfaction claim to each member.

Use the existing Link commands with the selected concrete Link Type. A plain
Memory or Issue is not a substitute source for an Evaluation-owned claim.
Subject To remains available for native Issues with the selected source
ownership behavior; ordinary Issue payload edits keep native validation.

### Deferred acceptance material

Future assessment qualification includes exact subject admission; lifetime
subject-group preservation through retyping; refusal of abstract Conformance
creation or claims from ordinary Memories; and revision of only the Evaluation
owner when adding a concrete claim. These are design acceptance cases, not
executed tests for the command set on this page. They must not be used as
Preview 3 availability evidence.

## Historical candidate ledger

The previous facade inspected
[Beads `04996869`](https://github.com/donnabox/beads/commit/0499686941260268ae4520c996ab2603f81248a6).
Its separately authored catalog, validation and retyping candidate had the
following reported scope. These retained descriptions do not qualify this
expanded CLI design or any later combined source.

| Area | Candidate status | Remaining acceptance |
| --- | --- | --- |
| Selected catalog and reusable validation | Authored; focused pure checks passed on an exact candidate | Installed provider and combined product checks |
| Properties, Memory-family retyping, ownership and exact Evaluation subjects | Authored; targeted cases passed on both backends | Broader affected backend and combined checks |
| `bd types` provider-driven inventory | Authored separately | Combined installed CLI and workspace checks |
| Typed create/update/show/list/filter/delete routing | Authored; targeted installed command cases passed | Broader affected CLI checks and integration landing |
| Integration and release | Pending | Reviewed exact source, required checks and approved integration landing |

These states describe the candidate, not feature availability in the integration
source above. The [release ledger](https://github.com/gastownhall/beads/issues/7656)
records exact qualification and landing evidence.

### Historical acceptance scope

The earlier acceptance material covered property preservation without numeric
conversion; complete-result validation on retyping; stale-guard refusal;
Evaluation subject lifetime binding; claim-owner revision isolation;
Link-family limits and forbidden cycles; native Issue edits in a mixed
workspace; and explicit refusal of incompatible earlier catalogs. Its
source's deletion route refused incident Links until explicit cleanup.

Qualification of any new combined candidate still requires installed commands
on embedded and fresh server workspaces, current and retained readback,
refusal atomicity and exact-source evidence. Pure validators, authored
examples and earlier candidate receipts do not qualify these new commands.
