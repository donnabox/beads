# Beads canonical CLI and convenience aliases

Status: design proposal based on the agreed generic-command direction.
Date: 2026-10-09.

This document specifies a target CLI, not the capabilities of an existing
release. It replaces the command-design alternatives in
[the research comparison](link-cli-namespace-variations.md); that document
retains the detailed source audit. Type representation and evolution remain
separate design decisions. No implementation or database migration is implied.

## 1. Framing the problem

### 1.1 The CLI must generalize beyond Issue-specific commands

Beads currently exposes related operations through several different shapes:

- Issue-oriented verbs such as `create`, `show`, `update`, and `delete`.
- A dependency namespace, `dep`, with its own verbs and shorthand.
- Relationship verbs and collections: `link`, `links`, and `unlink`.
- Memory conveniences: `remember`, `memories`, `recall`, and `forget`.
- Graph queries, history, and workflow operations with separate conventions.

Preview2 introduces independently identified, typed, versioned Links alongside
Issue and Memory Beads. User-defined Bead and Link Types will extend this
model. Type definitions may themselves become Beads. A CLI whose generic
operations assume Issues, Memories, or a fixed inventory of Types will require
new command-specific implementations for every extension.

The problem is therefore broader than renaming `dep` to `link`. We need one
operation model that can create, discover, inspect, change, and delete Resources
regardless of which built-in or user-defined Type describes them.

### 1.2 Design direction

The canonical interface uses generic verbs:

```text
bd create
bd list
bd show
bd update
bd delete
bd versions
bd compare
```

Command interpretation follows three rules:

1. **Creation is selected by Type:** `--bead-type` or `--link-type`.
2. **Listing is selected by filters:** kind, Type, endpoints, metadata,
   properties, and applicable domain filters.
3. **Existing-resource operations are selected by identity:** the ID selects
   the Resource kind, and the stored Resource determines its Type.

Topology operations use `bd graph traverse`, `bd graph tree`, and
`bd graph cycles`. They operate on graphs rather than individual Resources.

Existing commands can remain as aliases or compatibility adapters. `dep` is
never the canonical Resource namespace. Aliases are useful interfaces, not
temporary mistakes that must all be removed. They must share operation services
instead of becoming independent implementations of the data model.

### 1.3 Terminology

| Term | Meaning in this design |
| --- | --- |
| Resource | A Bead or Link addressed by canonical identity. |
| Kind | Structural category: Bead or Link. A Link has source and target endpoints. |
| Type | A nominal definition controlling a Resource's validated properties and applicable behavior. |
| Type definition | The object describing a Type. Its representation may become a Bead; this proposal does not require a third Resource kind. |
| Properties | The Resource's Type-defined data. |
| Issue classification | The legacy Issue `--type` value, stored as `properties.issue_type` on an Issue Bead. |
| Dependency classification | The legacy Dependency `--type` value, stored as `properties.dep_type` on a Dependency Link. |
| Metadata | A separate open JSON object for annotations. Metadata does not redefine Type properties or create edges. |
| Revision | The opaque token used to guard a change to the current Resource state. |
| Version | An opaque address for a retained Resource state. A token does not encode chronological order. |

A Type definition describing Links may itself be a Bead. Its kind and the
kind of its instances are different facts. If Types become Beads, their
definitions themselves use generic Bead CRUD and filtering, selected by the
appropriate metatype. Instances of the described Types retain their declared
Bead or Link kind. No concrete metatype ID is invented here.

Nominal Type and legacy classification are orthogonal. An Issue's nominal
Type is Issue; `task` or `bug` is its `issue_type` property. A Dependency's
nominal Type is Dependency; `blocks`, `tracks`, or `parent-child` is its
`dep_type` property. These properties apply only to Issues and Dependencies,
respectively. Adding a classification value does not install a nominal Type.

### 1.4 Compatibility is more than spelling

The source audit is pinned to upstream stable
[v1.3.1, c1c4b642](https://github.com/gastownhall/beads/tree/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c)
and the user-selected fork's
[preview2-integration, ff25e824](https://github.com/donnabox/beads/tree/ff25e824d80d4c1a52370c539ff8c767f1410de6).
These are different from the installed Homebrew 1.0.5 binary and the local
checkout's HEAD used for some initial exploration.

The following differences must survive normalization or be changed explicitly:

- Ordinary dependencies allow one edge per ordered endpoint pair. Preview2
  informational Links have their own IDs and permit parallel Links. Graph
  blocking Dependencies still allow one live Link per pair.
- `bd link --type` accepts custom strings in v1.3.1. `bd dep add --type`
  restricts types and normalizes dependency aliases. They are not equivalent
  handlers despite the help text's shorthand claim.
- Ordinary dependency classifications are strings; preview2 also has installed
  nominal Type descriptors. The agreed target models legacy Issue and Dependency
  `--type` values as `issue_type` and `dep_type` properties on their respective
  nominal Types. This requires an explicit data/model transition; it is not a
  claim that the pinned preview already implements that representation.
- `dep list` and `links` have different default directions and result shapes.
- Link removal, Bead deletion, Memory upsert, Issue closure and deduplication
  have different semantics. They cannot all be implemented as textual aliases
  for one generic mutation.

Related contributor work: [upstream PR #5648](https://github.com/gastownhall/beads/pull/5648)
proposes unifying legacy classification validation across link, batch and the
create form. This documentation proposal does not replace that implementation
work; its behavior comparisons remain pinned to the revisions above.

Sources: [legacy type validation](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/create_deps.go#L159),
[legacy link](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/link.go#L13),
[ordinary edge identity](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/internal/storage/issueops/dependencies.go#L258),
[graph Link lifecycle](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/cmd/bd/graph_preview_lifecycle.go#L13).

### 1.5 Boundaries

This design does not settle Type installation, metatype identity, schema
evolution, migration of existing instances, or whether Type deletion is allowed
while instances exist. Sharing the verbs `update` and `delete` does not decide
those policies.

It does not promise migration from ordinary Issue workspaces to graph
workspaces, arbitrary remote or pinned-version endpoints, erasure, restoration,
or support for graph Types that have no writer yet. Unsupported operations
must refuse explicitly, without falling through to a different storage mode.

## 2. Specification of the canonical CLI

### 2.1 Shared identity and Type rules

Bead IDs use `beads/PATH`; Link IDs use `links/PATH`. Exact canonical URLs in
the selected Scope are also valid where the implementation supports them.
Existing bare-ID Bead shorthand may remain. A generic command never probes
both namespaces to guess which Resource the caller intended.

Type selectors name an installed, usable Type. The selector's descriptor must
describe the requested kind. The current `types/NAME` shorthand and exact local
Type URL remain valid; representing definitions as Beads later requires an
explicit identity/alias decision, not silently rewriting those existing URLs.

ID, kind and nominal Type remain stable across ordinary updates. A Link's
source and target also remain stable. Type versioning is expected to introduce
explicit nominal Type updates later; this proposal retains the current
restriction until that design defines the operation and migration rules.
Changing identity or endpoints likewise requires a separate operation or a
new Resource, not a properties edit.

User-defined Types use the same operation paths as built-ins. Their properties
are validated by their descriptors; generic authoring must not assume Issue
fields, Memory title/body, or the preview's optional Link `note` property.

### 2.2 Create

```sh
bd create --bead-type TYPE --properties JSON [--metadata JSON] [--id ID]
bd create --link-type TYPE --source BEAD --target BEAD \
  [--properties JSON] [--metadata JSON] [--id ID]
```

| Input | Required behavior |
| --- | --- |
| `--bead-type` | Select Bead creation; reject `--source` or `--target`, including explicitly empty values. |
| `--link-type` | Select Link creation; require both endpoints and validate them against the descriptor. |
| Both selectors | Reject before writing. |
| Neither selector | Preserve default Issue creation as a convenience binding to the Issue Type. A Dependency alias may supply its nominal Type binding. Endpoints alone never infer a Link Type. |
| `--id` omitted | Allocate and reserve an ID in the selected kind's namespace. |
| `--id` supplied | Use exactly that identity or refuse; a deleted, reserved identity is not reusable. |
| Properties omitted | Use an empty object only if the Type permits it; otherwise report the required data. |
| Metadata omitted | Use the empty metadata object. |

JSON inputs support inline JSON, `@FILE`, and `@-` for stdin. Supplied objects
are validated before writing. Create never means upsert. Type-specific fields
and whole-document `--properties` cannot provide competing values for the
same properties document.

Endpoint order is explicit. For a Dependency with `dep_type=blocks`, SOURCE
depends on TARGET. For `tracks`, the tracker is SOURCE; for `parent-child`,
the child is SOURCE. The Dependency domain interprets these properties. Other
nominal Link Types define their own semantics. Generic parsing never reverses
operands based on the English meaning of a Type or classification name.

The pinned graph blocking route does not accept arbitrary properties,
metadata or a caller-selected Link ID. The target Dependency model needs a
writer that admits and validates `dep_type` and enforces its domain policy.
Renaming the command cannot make that capability available. Source ownership
and revision rules apply as specified in §2.6.

Legacy `--type` is a contextual property convenience. It can supply the Issue
or Dependency nominal Type binding when the command context already identifies
that domain; its value never determines the Resource kind:

| Context | Meaning of `--type VALUE` |
| --- | --- |
| Default `bd create`, or explicit Issue Type | Bind Issue if needed; set `properties.issue_type=VALUE`. |
| Explicit Dependency Link Type, or a Dependency convenience such as `bd dep add` or legacy `bd link A B` | Bind Dependency if needed; set `properties.dep_type=VALUE`. |
| Any other explicit nominal Type | Reject this domain shorthand. |

Thus `bd create --type task` remains valid, and `bd dep add A B --type tracks`
supplies the Dependency Type and endpoints. Generic Link creation still uses
`--link-type`; `bd create --type tracks --source A --target B` is rejected
because the default context is Issue. Open classification vocabularies make
guessing kind from values unsafe. Conflicting shorthand and explicit property
values are errors. Update resolves the same shorthand from the existing
Resource's nominal Type without changing it.

### 2.3 List: collection selection

```sh
bd list
bd list --kind bead
bd list --kind link
bd list --kind all
bd list --bead-type TYPE
bd list --link-type TYPE
```

| Selector | Meaning |
| --- | --- |
| No kind or nominal Type | Existing default Bead view. Preserve the documented default Issue visibility rules. |
| `--kind bead` | Bead collection. |
| `--kind link` | Current Link collection across Link Types. No endpoint anchor is required. |
| `--kind all` | Mixed current Bead and Link collection. |
| `--bead-type TYPE` | Beads of that nominal Type; implies Bead kind. |
| `--link-type TYPE` | Links of that nominal Type; implies Link kind. |

Reject both nominal Type selectors together. An explicit kind must agree with
the implied kind; `--kind all` plus a single-kind Type selector is rejected
rather than given an undocumented union meaning.

`--kind link` with no other filters means the workspace's current Links, not
Links belonging to an implicit last-touched Bead. No matches is a successful
empty result. Listing excludes removed Resources and retained historical
versions; use version operations for retained state.

`--kind all` is useful for Resource inventory, metadata audits and generic
clients. It does not mean unlimited output, historical records, or automatic
inclusion of every endpoint mentioned by a Link. Kind selection, visibility
and result limits are independent. Existing `--all` retains its documented
visibility/limit role; it does not become a synonym for `--kind all`.

Unanchored Link and mixed listing require proper collection queries. They
must not be implemented by repeatedly enumerating incident Links for every
Bead and hoping to deduplicate the results.

### 2.4 List: endpoint, metadata and property filters

```sh
bd list --kind link --source A
bd list --kind link --target B
bd list --kind link --incident-to A
bd list --link-type TYPE --source A --target B
bd list --kind all --metadata-field team=compiler
bd list --kind link --has-metadata-key reviewed_by
bd list --bead-type TYPE --property-field severity=high
bd list --link-type TYPE --has-property-key note
```

Endpoint filters require Link selection through `--kind link` or
`--link-type`. `--source A --target B` means AND. `--incident-to A` means
source A OR target A. Initially reject combining `--incident-to` with either
source or target; additional set expressions belong in the predicate syntax.

Metadata and properties remain distinct filter namespaces. Repeatable
`--metadata-field KEY=VALUE` and `--property-field KEY=VALUE` provide equality
filters. Existence filters distinguish absent fields from explicit JSON null.
Independent filters combine with AND. Repeating a key with conflicting values
must not silently replace the earlier predicate.

The existing metadata equality flags retain their accepted value conventions.
Typed comparison is available through the predicate form:

```sh
bd list --kind all \
  --where 'metadata.team = "compiler" AND properties.risk >= 3'
```

`--where` reuses and extends the `bd query` parser and evaluation model. It
supports explicit AND, OR, NOT and grouping, typed literals, comparisons and
field-existence tests. Strings require quoting when needed by the grammar;
JSON string `"3"` and numeric `3` must not silently compare as the same value.

The shared predicate model exposes canonical `source` and `target` identity
fields alongside `kind`, `type`, `metadata` and `properties`. Endpoint literals
are resolved and normalized using the same rules as endpoint flags. Here
`type` means nominal Type; legacy classifications use `properties.issue_type`
or `properties.dep_type`.

Before implementation, specify nested-path and literal-key escaping against
the existing metadata-key grammar. For example, a literal key containing a dot
must not silently acquire nested-object meaning. The examples here establish
the namespace and boolean model, not an incompatible replacement parser.

Across heterogeneous Types, a missing field fails positive comparisons; it
does not make the command invalid. An explicit presence/null predicate is used
when that distinction matters. Malformed predicates and unsupported operators
are errors. When an explicitly selected Type makes a property/operator
combination invalid, reject it rather than return a misleading empty result.

Current v1.3.1 supports ordinary metadata equality/key-existence filters, but
preview2 graph listing does not yet admit them. General property filters and
cross-kind predicates are target capabilities.
[Existing filters](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/list.go#L517),
[query evaluator](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/internal/query/evaluator.go#L499),
[preview allowlist](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/cmd/bd/graph_preview_list.go#L85).

### 2.5 List: domain shortcuts and presentation

Issue filters such as status, priority, assignee, labels and due dates remain
conveniences over the same query machinery. On an otherwise default Bead list,
an Issue-specific flag may select the existing Issue view, with the existing
disclosure. An explicitly selected Link, mixed collection, or incompatible
Bead Type rejects that shorthand instead of silently hiding other Resources.
Generic metadata/property predicates can filter a mixed collection.

`--type VALUE` follows the same domain binding rule as creation: the default
Bead view binds Issue and filters `properties.issue_type`; `--kind link`
binds Dependency and filters `properties.dep_type`. Explicit Issue or
Dependency Type selectors allow the corresponding shorthand. Other nominal
Types reject it. `--kind all --type VALUE` is ambiguous and is rejected;
mixed queries use explicit property predicates. For example:

```sh
bd list --type bug
bd list --kind link --type tracks
bd list --kind link --property-field dep_type=tracks
```

The second command selects Dependency Links explicitly through the shorthand;
the third uses a generic property filter and adds no nominal Type binding.

Filtering occurs before ordering and limiting. Ordinary flags and `--where`
must lower to one query model, not separate evaluators with subtly different
truth rules. Type-aware search and count adapters use that model as well.

Canonical results must have a stable record shape independent of filters or
row count. Human mixed output includes ID, kind and nominal Type, with a
Type-aware summary. It must not require every record to have title, status
or priority. Kind/Type can be CLI projection fields without changing the BDP
Resource schema.

Use deterministic ordering with canonical ID as a tie-breaker. Preserve legacy
default ordering on compatibility paths. Generic Link/mixed listing defaults
to ID order until a common ordering contract is defined. Issue-only sorts on
an incompatible collection are rejected.

Normal row limits apply; `--limit 0` explicitly requests an unbounded result
subject to advertised safety bounds. A bounded collection either exposes a
usable continuation or explicitly reports truncation. A `hasMore` boolean
without a cursor must never be described as pagination. Snapshot/continuation
semantics must be specified before promising a complete mutable-store walk.

Canonical JSON uses one documented collection contract, including items and
completeness/continuation information; exact envelope fields must be pinned
with conformance fixtures before implementation. Compatibility adapters may
preserve their existing JSON shapes. A changed contract at an existing root
spelling such as `bd list --json` requires an announced release boundary:
an alias cannot simultaneously preserve old bytes and publish new ones at the
same spelling. Existing graph `--format records-json` remains an explicit
compatibility format until that boundary is chosen.

### 2.6 Show, update and delete

```sh
bd show RESOURCE_ID [--version TOKEN] [--json]
bd update RESOURCE_ID --properties JSON [metadata edits] [--if-revision TOKEN]
bd update RESOURCE_ID --patch JSON [metadata edits] [--if-revision TOKEN]
bd update RESOURCE_ID --metadata JSON [--if-revision TOKEN]
bd delete RESOURCE_ID
bd delete RESOURCE_ID --force [--if-revision TOKEN]
```

`--if-revision TOKEN` opts into a revision precondition. Omission accepts the
current revision; there is no `--unconditional` flag in the target CLI. This
changes the preview's requirement to choose an explicit guard mode. A stale
precondition refuses the entire mutation, including an otherwise identical edit.
Accepted semantic no-ops retain the revision. Metadata and property changes
requested together commit atomically.

`--metadata` merges top-level keys; repeatable `--set-metadata` and
`--unset-metadata` provide typed key edits. Set and unset can combine, with
unset last; they cannot combine with the metadata-merge spelling.

> **Pending properties-edit design (Janet).** The requested direction is a
> properties-edit family symmetric with the metadata merge/set/unset family,
> plus `--patch`. Keep this sidebar until Janet's updated design is available.
> Exact flag names, merge versus replacement semantics, and combinations/order
> with `--patch` remain open here. The previous draft's whole-object replacement
> rule is not the settled target contract. The examples above name the entry
> points without deciding those details. Creation still validates the complete
> initial properties object. Every update validates the resulting object before
> committing properties and metadata together.

Owned Link mutations can also affect the source Bead's version. On every
owned-Link route, omission of `--if-source-revision` means **use the source
Bead's current state**. This applies uniformly to creation, update and deletion,
including generic blocking creation and all convenience adapters. The public
`--unconditional-source` flag is removed; accepting the current source requires
no flag.

When supplied, `--if-source-revision TOKEN` is checked atomically alongside
any Link `--if-revision` precondition, within the same operation transaction.
Either stale value refuses the entire operation, even a would-be semantic
no-op. Omitting one precondition does not disable the other. Creation has no
existing Link revision to guard, but still checks a supplied source revision.

Omission must record the actual source predecessor used by the operation and
preserve existing ownership, attribution and exactly-once rules. A changed
write accepting the current source retains the existing predecessor and
overwrite-attribution disclosure; an accepted semantic no-op creates no new
version and claims no overwrite. Accepting the current source does not permit
a separate unchecked read followed by a write, or duplicate source-version
or attribution effects.

Ownership comes from the installed descriptor. The target is not versioned
merely because it is linked. These defaults change only source-revision
acceptance, not which Bead owns a Link or which operations affect that owner.

This resolves the preview inconsistency: Memory-owned informational Links
already default to current-source acceptance, while generic blocking creation
requires an explicit source choice in the inspected Preview 2 contract. The
proposal replaces that distinction with the uniform rule above.
[Preview 2 source-guard contract](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/docs/reference/graph-cli-specification-draft.md).

**Proposed delete policy:** canonical `delete` previews by default for both
kinds; `--force` applies the deletion using the selected guard. This extends
the existing Bead-delete convention consistently to Links. It is a proposed
resolution of the earlier delete-policy question, not existing Link behavior.
Legacy `unlink` and `forget` adapters can preserve immediate application.

Force confirms application; it does not bypass revision checks, Type policy,
or incident-Link constraints. Deletion removes current state, reserves the
identity and retains supported prior versions. It does not mean erasure,
cascade or a promise of restoration. Preview2 Bead deletion refuses incident
Links; preserve that rule until a separate deletion-policy design changes it.

For ordinary dependencies without public Link Resource identity, retain pair
selection in compatibility adapters initially. A canonical ID-oriented
show/update/delete surface needs an explicit ordinary-to-Resource identity
mapping before it can be advertised. Do not guess a synthetic `links/ID`.

### 2.7 Versions and comparison

```sh
bd versions RESOURCE_ID
bd show RESOURCE_ID --version TOKEN
bd compare RESOURCE_ID --from TOKEN --to TOKEN
```

These work by identity and supported history capability, including when the
current Resource has been removed. Versions lists retained citable states in
the store's defined order; opaque tokens themselves are not sortable counters.
Compare uses the caller's explicit operand order.

Private deletion markers are not Resource versions. A missing Resource and a
deleted Resource with retained history must not be conflated. This design
does not equate graph Resource history with ordinary Dolt commit history.

### 2.8 Graph operations

```sh
bd graph traverse BEAD --direction out --depth 3
bd graph tree BEAD --direction out --link-type TYPE
bd graph cycles
bd graph cycles --link-type TYPE
bd graph cycles --root BEAD --metadata-field team=compiler
bd graph cycles --link-type TYPE --property-field dep_type=blocks
```

| Operation | Contract |
| --- | --- |
| `traverse` | Return the reachable selected subgraph from a root, with explicit direction/depth and node/Link bounds. |
| `tree` | Render a tree projection of a selected traversal, marking revisits, parallel Links and cycles. The underlying graph need not be a tree. |
| `cycles` | Discover cycles in selected edges. No root means workspace-wide; a root selects its directed reachable subgraph. |

All three operations share Link selection with `bd list`. They differ in what
they compute over that selection: reachability, a tree projection, or cycles.
They remain under `graph` because they compute topology instead of returning
the selected Resource rows. In the last example, TYPE must be the installed
Dependency Type; `dep_type` is its ordinary property, not a nominal Type name.

| List selector/filter | Fit for graph operations |
| --- | --- |
| `--link-type TYPE` | Select edges by nominal Link Type, including user-defined Types. |
| `--metadata-field`, `--has-metadata-key` | Select edges by Link metadata, using the same equality/existence semantics. |
| `--property-field`, `--has-property-key` | Select edges by Link properties, with the same Type validation and missing-field rules. |
| `--where EXPRESSION` | Use the shared predicate model over Links, including boolean combinations of Type, endpoint, metadata and property conditions. |
| `--source`, `--target`, `--incident-to` | Restrict eligible edges globally, using the list rules. These can cut paths; they do not designate a traversal root. |
| `--kind` | Link kind is implicit. An explicit `--kind link` is redundant but valid; Bead/all selections are rejected. |
| `--bead-type` and Issue field shortcuts | No unqualified node filter. Future node-selection and output-filter syntax must distinguish pruning traversal from hiding displayed nodes. |
| Sorting and row limits | Apply to presentation only, where supported. They must never silently truncate the input to reachability or cycle analysis. |
| Legacy `--all` | Does not select an edge domain or remove algorithm bounds. Legacy visibility behavior belongs to the relevant adapter. |

Edge filters combine with AND, with explicit OR/NOT/grouping in `--where`.
They use the shared fields and identity rules in §2.4. Selecting a Type or a
Link property does not implicitly filter endpoint Bead properties.

Selection precedes traversal and cycle analysis. The root limits reachability
within that filtered graph. For example, `--incident-to A` admits only edges
that touch A; a traversal rooted at A may follow a longer path when it does
not impose that filter. The selected subgraph is read consistently.

Directions are `in`, `out`, and `both`. New traverse/tree commands default to
`out`; compatibility adapters retain their old defaults. Depth defaults to
one expansion step; callers can request greater depth. Depth zero selects the
root only. Cycles has no implicit one-step depth cutoff. A rooted cycle query
examines directed reachability from the root; cycle analysis always uses stored
edge direction and never invents reverse edges.

Node/Link bounds must be explicit in help and completeness output. An output
limit may reduce the displayed report but cannot change what was analyzed.
If an analysis bound prevents completion, report that fact and the scanned
scope; never report the selected graph as acyclic on the basis of a partial
scan. A completed cycle query succeeds whether it found zero or more cycles.
Compatibility adapters preserve any legacy exit convention.

There is no canonical `--edge-set` flag. With no filters, cycles examines all
current Links. Cycles are graph facts, not automatically domain errors:
informational cycles can be valid. Dependency-specific scheduling selection
belongs to the `bd dep cycles` adapter. It binds the Dependency nominal Type
and translates the domain policy into predicates over `properties.dep_type`.
The inspected default scheduling set is `blocks`, `conditional-blocks`, and
`parent-child`; it is not the set of every readiness-affecting classification.
The domain service supplies this policy, rather than hardcoding it in generic
CLI parsing. Section 3.2 gives the translation and explains the additional
`--include-tracks` behavior.

### 2.9 Implementation and remaining decisions

Canonical commands and adapters share Type resolution, validation, query,
mutation, history and graph services. They must not invoke the CLI recursively
or duplicate storage writers. Register separate Cobra command instances for
separate paths. Preview2's command-identity admission gate must deliberately
admit new routes; unrecognized graph operations cannot fall into legacy storage.

Before implementation, settle these bounded details without reopening the
generic-command direction:

- Confirm the proposed uniform delete preview/apply convention.
- Pin predicate escaping, typed literals and missing/null truth tables.
- Pin canonical JSON envelopes, pagination consistency and the release boundary
  for root-command output changes.
- Define ordinary dependency Resource identity and the migration to Issue and
  Dependency nominal Type bindings with `issue_type`/`dep_type` properties.
- Reconcile the properties-edit sidebar with Janet's updated design.
- Define Type authoring, installation, evolution and metatype representation
  in the Type design; reuse this CLI's verbs once those services exist.

Acceptance coverage must include both workspace modes, alias defaults/output,
Type validation, directional semantics, ambiguous pairs, bulk atomicity,
multi-edge identity, stale Link/source guards, no-ops, retained history,
mixed-resource filters and incomplete graph scans. Owned-Link coverage must
exercise omitted and supplied source preconditions on every create/update/delete
route and adapter, independently of Link preconditions; verify atomic refusal
on either stale token (including no-ops), actual source-predecessor recording,
and preservation of ownership, attribution and exactly-once effects. The source
audit verifies existing behavior; this document does not claim those target checks have run.

## 3. Detailed convenience aliases

### 3.1 What “alias” means

Three forms are supported:

| Form | Purpose |
| --- | --- |
| Simple alias | Another spelling for the same operation and result contract. |
| Compatibility adapter | Translates arguments, supplies historical defaults, preserves applicable legacy validation/output and calls shared services. |
| Domain convenience | Performs a meaningful workflow, possibly composing multiple changes atomically. It must disclose effects beyond Resource CRUD. |

Aliases may remain indefinitely. Document their canonical operation and any
extra behavior in help. They cannot silently fall back across workspace modes,
invent missing Type mappings, or replace unsupported functionality with a
partial approximation. Conflicting explicit options are errors, not silently
overridden defaults.

The following mappings specify intent. Legacy classification strings become
properties on Issue or Dependency instances; they do not become nominal Type
URLs. Resolve the nominal Type binding separately.
Where Resource identity or Type services do not yet exist, the adapter shares
the appropriate domain service and reports the capability boundary honestly.

### 3.2 Existing relationship commands

The translations below specify the target canonical calls. `D` stands for the
workspace's installed Dependency nominal Type selector, not a new literal Type
ID. `A` and `B` stand for normalized Bead identities; `L` is a resolved canonical
Link identity. `N` is the legacy maximum depth (default 50). Replace these
placeholders with the resolved values before executing a command. Other
explicit compatible options are forwarded; conflicting options are rejected.

For arbitrary `--type T`, the adapter validates/normalizes T under its existing
policy, then JSON-encodes it as `{"dep_type":T}`. It never substitutes an
unescaped string into JSON. Concrete examples below use `tracks` and `blocks`.

| Existing command | Exact canonical syntax after binding/resolution | Adapter behavior beyond the canonical call |
| --- | --- | --- |
| `bd link A B` | `bd create --link-type D --source A --target B --properties '{"dep_type":"blocks"}'` | Preserve historical blocking default and the applicable workspace capability checks. |
| `bd link A B --type tracks` | `bd create --link-type D --source A --target B --properties '{"dep_type":"tracks"}'` | The legacy link path retains its custom-classification policy. Other admitted values populate the same property. |
| `bd link A B --link-type TYPE` | `bd create --link-type TYPE --source A --target B` | Forward supported properties, metadata and ID options. Forward an optional `--if-source-revision`; omission accepts the current source under §2.6. |
| `bd dep add A B` | `bd create --link-type D --source A --target B --properties '{"dep_type":"blocks"}'` | B blocks A; preserve dependency validation and rendering. |
| `bd dep add A B --type tracks` | `bd create --link-type D --source A --target B --properties '{"dep_type":"tracks"}'` | Preserve the dep path's validated vocabulary and normalization before populating the property. |
| `bd dep add A --blocked-by B` / `--depends-on B` | `bd create --link-type D --source A --target B --properties '{"dep_type":"blocks"}'` | Preserve the ordered pair; normalize documented blocking aliases. |
| `bd dep A --blocks B` / `-b B` | `bd create --link-type D --source B --target A --properties '{"dep_type":"blocks"}'` | Reverse the operands: B depends on A. |
| `bd dep add --file FILE` / `--file -` | No single canonical CLI invocation yet; normalize each entry to the corresponding `bd create --link-type D --source A --target B --properties JSON` operation inside the bulk service. | Preserve JSONL grammar, same-store requirements, transaction semantics and final cycle checks. Do not execute independent create commands. See §3.9 for the separate batch interface. |
| `bd dep list A` | `bd list --link-type D --source A` | Default down→out; project related Issues into the legacy single-ID output. |
| `bd dep list A --direction up` | `bd list --link-type D --target A` | Up→in; retain the legacy projection. |
| `bd dep list A B` | `bd list --link-type D --where 'source = "A" OR source = "B"'` | A/B inside the predicate stand for canonical endpoint IDs. Extend the OR for more IDs; retain multi-ID raw-edge output. Up uses target instead of source. |
| `bd dep list A --type tracks` | `bd list --link-type D --source A --property-field dep_type=tracks` | Filter the Dependency classification, not nominal Type. |
| `bd links A` | `bd list --kind link --incident-to A` | Preserve default direction both and the existing result shape. |
| `bd links A --direction out` / `in` | `bd list --kind link --source A` / `bd list --kind link --target A` | Preserve direction and rendering. |
| `bd dep remove A B` / `rm A B` | Resolve with `bd list --link-type D --source A --target B`; apply `bd delete L --force`. | Pair resolution and deletion share the mutation service. Preserve supported dangling/external target handling and ordinary-mode pair deletion where Resource IDs do not exist. |
| `bd unlink LINK` | `bd delete links/LINK --force` for a bare ID; `bd delete L --force` for a qualified identity. | Bare ID means Link here. Forward supplied Link/source revision preconditions; omission of the source precondition accepts its current state under §2.6. |
| `bd unlink A B --link-type TYPE` | Resolve with `bd list --link-type TYPE --source A --target B`; apply `bd delete L --force`. | Require exactly one informational Link; refuse zero/ambiguous matches. Blocking graph pair deletion remains unsupported until its domain service supports it. |
| `bd dep relate A B` | `bd create --link-type D --source A --target B --properties '{"dep_type":"relates-to"}'` and `bd create --link-type D --source B --target A --properties '{"dep_type":"relates-to"}'` | Two directional Dependencies, with the workflow's transaction contract; not one nominal `related` Link. |
| `bd dep unrelate A B` | Resolve using `bd list --link-type D --property-field dep_type=relates-to --where '(source = "A" AND target = "B") OR (source = "B" AND target = "A")'`; apply `bd delete L --force` to each resolved identity. | Select both ordered pairs and validate classification. Do not reproduce untyped deletion of an unrelated edge; disclose that correction. |
| `bd dep tree A` | `bd graph tree A --link-type D --direction out --depth N` | Translate down→out, up→in, both→both and max-depth→depth. Preserve legacy status/output filtering, Mermaid rendering and deprecated flag handling without changing traversal semantics. |
| `bd dep cycles` | `bd graph cycles --link-type D --where 'properties.dep_type = "blocks" OR properties.dep_type = "conditional-blocks" OR properties.dep_type = "parent-child"'` | Resolve the scheduling predicate from domain policy. Preserve output and exit conventions. |
| `bd dep cycles --include-tracks` | `bd graph cycles --link-type D --where 'properties.dep_type = "blocks" OR properties.dep_type = "conditional-blocks" OR properties.dep_type = "parent-child" OR properties.dep_type = "tracks"'` plus the domain cycle-report rule below. | This edge selection alone is not a complete translation of the legacy report. |

The query-then-delete entries describe resolution and mutation, not safe
multi-process shell recipes. The adapter checks identity, classification and
pair cardinality at the write boundary and applies supplied revision
preconditions. It cannot select the first of several matches. Where ordinary
Dependencies have no public Link ID, it continues using the equivalent domain
pair service until Resource identity mapping exists.

Bidirectional operations compose two mutations and must preserve and document
their real transaction/partial-failure contract. A table containing two calls
does not promise that two independent CLI invocations are atomic. Graph support
requires the underlying writer and admission changes.

`--include-tracks` widens the legacy scheduling report while excluding loops
made only of tracks edges. Its existing reporting algorithm and projection
remain in the domain service; merely admitting tracks into a generic edge
predicate does not reproduce that result contract. The legacy `--limit` caps
the rendered cycle list, retains the total count, and does not truncate JSON.
These are adapter presentation rules, not bounds on graph analysis.
[Preview cycle options](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/cmd/bd/dep.go).

Ordinary `link --type` and `dep add --type` retain their different validation
policies unless explicitly changed. Their translated values are properties of
the Dependency nominal Type. A custom `blocked-by` classification admitted by
legacy link must not silently acquire scheduling semantics merely because the
command is renamed. Nominal Type binding and classification validation are
separate steps; installed-Type capability restrictions still apply.

### 3.3 Existing generic reads and graph views

In the remaining tables, `I` and `M` are the workspace's Issue and Memory
nominal Type selectors; `K` is the eventual metatype selector, if Type
definitions become Beads. `JSON`/`@FILE` contains the validated properties
prepared by the adapter; `EXPRESSION` is its normalized predicate, including
legacy visibility rules. These are placeholders, not new literal Type IDs.
“Retain” means the domain command remains a useful interface because no
single generic CRUD command expresses its full behavior. It does not make
`dep` a canonical namespace. Compound calls describe shared-service operations,
not independent shell commands with an implied transaction.

| Existing command | Canonical spelling / operation composition | Adapter behavior |
| --- | --- | --- |
| `bd show/update/delete ID` | `bd show ID`; `bd update ID` with the supplied edits; `bd delete ID` (or `--force` to apply). | Canonical spelling. Keep default Bead shorthand and applicable Issue field conveniences. Output/guard changes at the same spelling require a declared compatibility boundary. |
| `bd versions ID` | `bd versions ID`. | Canonical retained-version operation where supported. Ordinary and graph history backends remain distinct. |
| `bd history ID` in graph mode | `bd versions ID`. | Alias to versions, preserving its accepted flags and output. |
| `bd history ID` in ordinary mode | Retain `bd history ID`; no equivalent retained-Resource-version call. | Retain ordinary commit-history semantics; do not relabel Dolt commits as Resource versions. |
| `bd graph A --view generic` | `bd graph traverse A --direction both --depth N`, forwarding the existing depth and bounds. | Traverse adapter retaining preview2's default direction both, depth/bounds and output. |
| Ordinary `bd graph A` / `--all` | Retain `bd graph A` / `bd graph --all`; the dependency/epic view has no single generic rendering equivalent. | Retain existing dependency/epic view and rendering options, including dot/HTML/compact modes. Reuse graph services without silently changing view semantics. |
| `bd graph check` | The `bd graph cycles --link-type D --where ...` scheduling predicate from §3.2, plus the check adapter's verdict/exit projection. | Compatibility scheduling-cycle check using the same domain predicate translation as `bd dep cycles`; preserve its check result/exit behavior. The inspected implementation does not supply every extra check its help wording suggests. |

### 3.4 Query, search, count and work views

| Existing command | Canonical spelling / operation composition | Adapter behavior |
| --- | --- | --- |
| `bd query EXPRESSION` | `bd list --bead-type I --where EXPRESSION` after grammar and default-visibility normalization. | Generic selection/predicate service underlying `list --where`; retain the legacy Issue grammar, default exclusion of closed Issues, date expressions, field aliases, parse-only mode and output. Explicit generic selection can expose the new Resource query contract. |
| `bd search TEXT` | Retain `bd search TEXT`; an exact resolved identity can use `bd show beads/ID`, but generic text-search/prefix syntax is not specified here. | Query convenience with its existing ID-like exact/prefix path versus title-text path. Preserve all-status search by default; it is not simply default `list --title`. Description search stays explicit. |
| `bd count [filters]` | Retain `bd count [filters]`; generic aggregation syntax is not specified by CRUD. | Aggregation over the same selection model, independent of list page limits. Preserve its Issue-only, all-status default and `--by-*` grouping semantics; never count only the first returned page. Generic Resource counting is an extension. |
| `bd children PARENT` | Select edges with `bd list --link-type D --target PARENT --property-field dep_type=parent-child`; resolve/project each source with `bd show beads/ID`. | Issue child query / parent-child relationship view, preserving its all-status and presentation defaults. Return child Beads, not raw Links. |
| `bd ready` | Retain `bd ready` and its claim options; no single generic list/update equivalent. | Domain readiness query using actual scheduling semantics. A claim option remains an atomic selection-and-claim workflow, not list followed by update. |
| `bd blocked` | Retain `bd blocked`; no equivalent stored-status predicate. | Domain blocker query, not `status=blocked`; dependent Issues can still have stored status open. |
| `bd stale`, `bd orphans`, `bd lint` | Retain `bd stale`, `bd orphans`, and `bd lint`; their full analyses have no specified single generic equivalent. | Keep their domain analysis and defaults. Reuse query facilities where applicable; do not claim they are simple field-filter aliases without proving equivalence. |

Preview2 graph mode currently refuses query/search/count despite those commands
being registered in the binary. Exposing their generic forms requires query
and aggregation implementation and admission work.

### 3.5 Memory conveniences

| Existing command | Canonical spelling / operation composition | Adapter behavior |
| --- | --- | --- |
| `bd remember TEXT` | `bd create --bead-type M --properties JSON` with the derived Memory properties. | Create a Memory using its Type-specific body/title binding. Preserve existing title derivation and body-input options. |
| `bd remember --id ID ...` in graph mode | Create branch: `bd create --bead-type M --id ID --properties JSON`; update branch: `bd update beads/ID --patch JSON`. Branch selection and edits occur atomically in the upsert service. | Create if unused; update the existing Memory otherwise. Preserve omitted fields, guard defaults, and refusal for deleted/non-Memory identities. This is a deliberate upsert convenience, not generic create behavior. |
| `bd remember --create-only --id ID ...` | `bd create --bead-type M --id ID --properties JSON`. | Generic create semantics: refuse any allocated identity. |
| `bd remember --update ID ...` | `bd update beads/ID --patch JSON` after the Memory-only check; patch details await §2.6. | Existing-Memory-only patch; never create a missing Resource. |
| `bd memories [SEARCH]` | `bd list --bead-type M`; retain the search adapter until generic title/body-search syntax is specified. | Memory-Type collection with existing title/body search, summary, details and output conventions. |
| `bd recall ID [--version TOKEN]` | `bd show beads/ID [--version TOKEN]` followed by the Memory-body projection. | Read a Memory and emit its exact body bytes, unaffected by quiet. Preserve the no-added-newline contract and preview refusal of JSON; it is not ordinary show output. |
| `bd forget ID` | `bd delete beads/ID --force` after the Memory-only check; forward revision preconditions. | Memory-only deletion convenience preserving immediate apply and graph guard/incident-Link rules. Reject other Bead Types. |

Ordinary memories are key/value-backed and do not become graph Beads merely
because these commands are aliases. Preserve their key addressing and backend
contracts until a separate migration exists, including `remember --key` and
the existing bare-key recall behavior. The graph upsert must resolve
identity and apply atomically through shared services; a shell-style
“show, then create or update” sequence is not sufficient.

### 3.6 Issue and compound workflows

| Existing command or flag | Canonical spelling / operation composition | Adapter behavior |
| --- | --- | --- |
| `bd create TITLE`, existing `new` alias, `q TITLE` | `bd create --bead-type I --properties JSON` with TITLE and the Issue defaults encoded in JSON. | Default Issue-Type convenience. Preserve q's ID-only result and accepted inputs. |
| `bd create --parent P` | `bd create --bead-type I --properties JSON`, then `bd create --link-type D --source NEW --target P --properties '{"dep_type":"parent-child"}'` within the compound service. | Atomic Issue creation plus structural relationship and applicable inheritance/ID policy. |
| `bd create --deps SPEC` | `bd create --bead-type I --properties JSON` plus one `bd create --link-type D --source A --target B --properties JSON` per normalized dependency, in the same compound service. | Atomic Issue-and-edge creation. Preserve historical parsing: explicit `blocks:B` reverses direction, while bare B and `blocked-by:B` mean the new Issue depends on B. |
| `bd create --waits-for ...` / graph-plan creation | Retain `bd create --waits-for ...` and graph-plan creation; no single generic equivalent or complete canonical plan-input format is specified. | Compound domain creation including edge metadata and scheduling policy; preserve atomicity. |
| `bd update --parent P` / `--parent ''` | Resolve with `bd list --link-type D --source ID --property-field dep_type=parent-child`; delete the old L with `bd delete L --force`; when P is nonempty, create with `bd create --link-type D --source ID --target P --properties '{"dep_type":"parent-child"}'`. | Reparent or clear parent; replace/remove the appropriate old edges instead of merely adding another. |
| `bd close` / `done`, `reopen`, `defer`, `undefer` | Retain `bd close ID`, `bd reopen ID`, `bd defer ID`, and `bd undefer ID` with their domain options; no equivalent generic property-only edit. | Preserve Issue lifecycle rules, reason handling, events, batching, no-ops and any next-work effects. Not simple property assignment. |
| `bd update --claim`, `bd unclaim` | Retain `bd update ID --claim` and `bd unclaim ID`; no equivalent unguarded property edit. | Preserve atomic claim/lease/holder policy and concurrency checks. |
| `bd assign` | `bd update ID --assignee USER` through the existing Issue field convenience; generic property-edit spelling awaits §2.6. | Preserve assignment/update-assignee behavior; assigning an Issue is not automatically a lease-bearing claim. |
| `bd duplicate A --of B` | Retain `bd duplicate A --of B`; relation creation and closure compose a domain workflow. | Add the appropriate relation and close A; preserve backend restrictions and existing transaction behavior. |
| `bd supersede OLD --with NEW` | Retain `bd supersede OLD --with NEW`; Link creation and closure require the domain workflow. | OLD→NEW relationship plus closure of OLD. Do not reverse the existing stored direction. |
| `bd duplicates --auto-merge` | Retain `bd duplicates --auto-merge`; no single generic mutation equivalent. | Preserve child movement, issue closure and its actual relationship Type; it is not an alias for one Link create. |
| `bd label`, `tag`, `comment`, `comments`, `note`, `edit` | Retain these command families. Property edits may use `bd update ID` once §2.6 is settled; feeds and editor interactions have no generic CRUD equivalent specified here. | Retain their property, set, append/feed or editor semantics. Use shared services where equivalent; comments may live outside Resource snapshots and edit may invoke an editor. |
| Gate creation, molecule bonding/cooking, graph apply, imports and sync | Retain their existing command paths; no blanket generic CRUD translation. | Domain workflows that maintain Links internally. They do not need to move into a Link namespace. |

When a workflow composes operations, preserve its real atomicity/partial-failure
contract; do not improve or weaken it accidentally under an alias label.
Current graph admission gaps remain explicit. This design does not assert that
all ordinary Issue conveniences already work in preview2.

### 3.7 Optional noun-scoped conveniences

These can be added for discoverability without becoming the canonical model:

| Convenience family | Canonical spelling / operation composition | Adapter behavior |
| --- | --- | --- |
| `bd bead create/list/show/update/delete` | `bd create --bead-type TYPE`; `bd list --kind bead`; `bd show beads/ID`; `bd update beads/ID`; `bd delete beads/ID`, with the supplied payload/options. | Restrict generic operations to Bead kind. |
| `bd link create/list/show/update/delete` | `bd create --link-type TYPE --source A --target B`; `bd list --kind link`; `bd show links/ID`; `bd update links/ID`; `bd delete links/ID`, with the supplied payload/options. | Restrict to Link kind; creation accepts positional endpoints; Link-only IDs may be bare. `bd link A B` remains the legacy parent shorthand. |
| `bd issue create/list/show/update/delete` | `bd create --bead-type I`; `bd list --bead-type I`; `bd show beads/ID`; `bd update beads/ID`; `bd delete beads/ID`, with payload/options and an Issue-only check. | Bind the installed Issue Type and admit Issue field conveniences. |
| `bd memory create/list/show/update/delete` | `bd create --bead-type M`; `bd list --bead-type M`; `bd show beads/ID`; `bd update beads/ID`; `bd delete beads/ID`, with payload/options and a Memory-only check. | Bind the installed Memory Type. These generic noun aliases do not silently inherit remember's upsert or recall's exact-body output. |
| `bd type create/list/show/update/delete` | If definitions become Beads: `bd create --bead-type K`; `bd list --bead-type K`; `bd show beads/ID`; `bd update beads/ID`; `bd delete beads/ID`, subject to Type lifecycle policy. | Semantic Type-definition view over the eventual Type services; if definitions are Beads, use the metatype and generic Bead operations. |

Type bindings must be resolved from the workspace's supported Type model, not
hardcoded example preview Type IDs. A requested binding that is absent fails
clearly. Type lifecycle verbs remain unavailable until their services exist.

For a parent that supports both a shorthand and subcommands, verb names and
aliases take precedence before ID resolution. Qualified IDs avoid collisions.
Do not guess that a token is an ID merely because a matching Bead exists.

User-defined Types need no generated top-level nouns. Operators can compose
their own shell conveniences around the canonical commands. A configurable
alias registry is a possible later feature, not a requirement of this design.


### 3.8 Type discovery conveniences

| Existing or proposed convenience | Canonical spelling | Adapter behavior |
| --- | --- | --- |
| Ordinary `bd types` | Retain `bd types`; no generic Type-definition query equivalent. | Lists Issue classifications, not nominal Types. |
| Graph `bd types` | Retain `bd types` until Type-definition representation is settled. | Lists installed nominal Type descriptors. |
| Future `bd type list` | `bd list --bead-type K` if definitions become Beads. | K must resolve to the supported metatype; installation state and lifecycle policy still belong to the Type service. |

`bd types` currently means different things by workspace: Issue classifications
in ordinary mode, installed Bead/Link descriptors in graph mode. Preserve that
behavior explicitly. A future `bd type list` can provide a Type-definition
view once the representation is resolved; it must not silently recast task/bug
classifications as nominal Types or treat every Bead as a Type definition.

### 3.9 Batch operations: a separate parser and transaction boundary

> **Open for the next design session.** The handling of `bd batch` is deferred.
> This subsection records the inspected behavior and constraints; it does not
> select a future batch language, translation strategy, or rollout plan.

`bd batch` is an existing write-oriented script runner in both inspected
sources. It reads commands from stdin or `-f/--file FILE`, parses its own
line-oriented language, and executes them through a shared transaction. A
successful batch makes one Dolt commit; any execution error rolls back the
whole batch and reports the failing line. It was introduced to reduce the
write amplification from scripts that invoke `bd` repeatedly.

The current grammar is deliberately small:

```text
close ID [reason...]
update ID KEY=VALUE [KEY=VALUE ...]
create TYPE PRIORITY TITLE...
dep add FROM_ID TO_ID [TYPE]
dep remove FROM_ID TO_ID
```

`dep rm` is also accepted. Blank lines and lines beginning with `#` are
ignored. Tokens are separated by whitespace; double quotes allow spaces,
with escaped double quotes and backslashes inside quoted strings. This is
not a shell: it does not evaluate variables, pipes or command substitutions,
and input lines do not start with `bd`.

For example, using IDs already present in the selected ordinary workspace:

```sh
bd batch <<'BATCH'
# These changes succeed or roll back together.
update bd-1 status=in_progress assignee=alex
dep add bd-1 bd-2 blocks
create task 2 "Review the dependency cleanup"
BATCH
```

The positional TYPE in `create` is an Issue classification; the optional TYPE
in `dep add` is a Dependency classification and defaults to `blocks`. In the
target model they populate `issue_type` and `dep_type`, respectively, while
binding the appropriate nominal Type. The batch dependency parser currently
uses its own validity check, admitting custom nonempty classifications rather
than reusing the top-level dep command's whitelist and alias normalization.
That difference must be preserved or intentionally changed.

| Existing batch feature | Actual scope and limitation |
| --- | --- |
| `update` keys | Only `status`, `priority`, `title`, `assignee`, and `force`. Other keys fail. `force` is a policy override, not a stored field. |
| `--dry-run` | Parses and echoes input without executing it. It does not prove that identities, operation arguments, revision preconditions or domain rules will pass at execution. |
| `-m` / `--message` | Supplies the single Dolt commit message. |
| `--json` | Reports operation results or a batch error. It does not change the input grammar into JSON. |
| `create` output | Reports the created identity; the current grammar has no placeholder/reference mechanism for later lines to use that generated identity. |
| Closing | The documented current `close ID` path and `update ID status=closed` path differ: update enforces open-child/live-blocker policy unless forced; close does not apply that policy. Alias normalization must not silently erase this difference. |
| Other CLI commands/flags | Not automatically accepted. Reads, complex creation, graph Resource flags and newly registered Cobra aliases do not become batch syntax. |

The parser dispatches directly to transaction operations rather than recursively
invoking top-level Cobra handlers. Consequently, renaming or aliasing `dep`
at the root does not affect `dep add/remove` inside batch input. Any decision to translate existing batch scripts would need to address this
parser explicitly. Whether to preserve its language, extend it, or introduce
canonical operation records remains open, including how Type selectors,
properties, metadata and optional preconditions would be expressed.

Any extension must retain one transaction, rollback on failure, and the domain
validation required by the admitted operations. It must not spawn a CLI
subprocess per line. Do not assume batch has the same final whole-graph cycle
validation as `dep add --file`; those are distinct runners and must be audited
separately. Graph-mode admission and transaction support must be established
before promising generic Bead/Link batching.

Sources: [stable batch parser and runner](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/batch.go),
[preview integration batch parser and runner](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/cmd/bd/batch.go).

### 3.10 Compatibility rollout

Implementation rollout must update help, completion, examples, prompts,
provider scripts and structured-output consumers together. Every compatibility
adapter needs a recorded contract covering operand direction, defaults,
validation, output, errors, mutation effects and workspace capabilities.

Existing aliases should not produce noisy deprecation warnings merely because
they are aliases. New documentation should teach the generic canonical verbs
first, then show the useful conveniences and their additional behavior.

The compatibility goal is preservation of useful existing workflows, not
preservation of known unsafe behavior. Explicitly disclose exceptions such as
typed unrelate removal, removal of `--unconditional` with optional
`--if-revision`, removal of `--unconditional-source` with current-source
acceptance by default on every owned-Link route, and any root-command JSON
contract change. No alias can make an unsupported storage capability exist or
resolve an unsettled Type lifecycle policy by implication.
