# Beads relationship CLI: two namespace designs

The command-design discussion in this research note is superseded by
[Beads canonical CLI and convenience aliases](canonical-cli.md). Retain this
file as the source audit and record of explored alternatives; use the new
document for the current proposal.

Research and design discussion, 2026-10-09. This is a proposal, not a change to
the CLI or an amendment to the graph/BDP specifications.

## Sources and scope

The comparison pins two sources rather than relying on the locally installed
binary or a moving upstream main:

- Current upstream stable release: **v1.3.1**, commit
  `c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c`, released September 30.
  [Release](https://github.com/gastownhall/beads/releases/tag/v1.3.1).
- User-selected fork branch: **donnabox/beads release/preview2-integration**,
  commit `ff25e824d80d4c1a52370c539ff8c767f1410de6`.
  [Pinned source](https://github.com/donnabox/beads/tree/ff25e824d80d4c1a52370c539ff8c767f1410de6).
- The fork's preview2-candidate was also compared, at
  `4cac507fd5c55cf632fee7ca52df0d26befd4159`; integration is the primary source.
  The installed Homebrew bd 1.0.5 and local checkout HEAD 96b879a27490 are not
  the release baselines for this document.

Research used command registrations, admission gates, handlers, storage
contracts, tests, and documentation. Neither release binary was built or
installed, and no live database was changed. Some preview specification rows
are pinned to an older build and still mark implemented behavior NYI; code and
tests at the selected commit determine the implemented inventory below.

“All link operations” means standalone relationship creation, inspection,
editing, removal, history and relationship queries. Issue lifecycle operations
and atomic compound workflows can still maintain edges internally. For
example, marking an Issue duplicate also closes it; that is not interchangeable
with creating a `duplicates` edge. This boundary applies to both variations.

## Accepted design constraints

- Existing top-level commands may remain as compatibility aliases/adapters to
  the generic operations; they do not need to be removed after a transition.
- `dep` is never the canonical relationship namespace in either variation.
- User-defined Bead and Link Types must use the same operation vocabulary as
  built-in Types. The design also covers CRUD on Type definitions themselves.
- Type definitions may become Beads. That representation is not settled; CLI
  design must not require Types to be a permanently separate Resource kind.
- “All under link” means all canonical standalone relationship operations live
  there, not that every compatibility spelling must disappear.
- The variation without a Link namespace uses generic Resource operations.
  `bd link` may survive only as a legacy creation alias, with no canonical
  operation family beneath it. A literal removal is optional, not required.

## Current direction under discussion

The operator is leaning toward generic canonical verbs. Creation selects its
Resource kind through `--bead-type` or `--link-type`; bead creation rejects
`--source` and `--target`. Show/update/delete dispatch by ID space. Listing
should extend the existing filtering model. The noun-scoped variation remains
below for comparison; it is not the current preferred direction.

## What exists now

| Capability | Upstream v1.3.1, ordinary workspace | Preview2 integration, graph workspace |
| --- | --- | --- |
| Create | `bd link A B`; `bd dep add A B`; `bd dep BLOCKER --blocks BLOCKED` | `bd link SOURCE TARGET [--link-type TYPE]`; `bd dep add SOURCE TARGET` for blocking Issue dependencies only |
| List | `bd dep list ID...`, outgoing by default; `--direction up` for incoming | `bd links BEAD`, both directions by default; `--direction in/out/both`, optional Link Type filter |
| Inspect one relationship | No standalone Link-resource show command; dependency views expose edges or related Issues | `bd show links/ID [--version TOKEN]` |
| Edit properties or metadata | No standalone general dependency update command | `bd update links/ID --properties JSON`, `--patch JSON`, or metadata edits; informational Links only |
| Remove | `bd dep remove A B`, alias `rm` | `bd unlink LINK`, or an unambiguous informational pair plus `--link-type`; blocking removal requires Link ID |
| Bidirectional convenience | `bd dep relate A B` / `unrelate A B` | Not admitted in graph mode |
| History / compare | Issue history and database history; no equivalent public Link-resource history | `bd versions links/ID`, `bd history links/ID`, `bd compare links/ID --from TOKEN --to TOKEN` |
| Tree / cycles / graph | `bd dep tree`, `bd dep cycles`, `bd graph`, `bd graph check` | `bd graph BEAD --view generic`; ordinary dep tree/cycles/list/remove are not admitted |
| Type discovery | Built-in dependency constants and flag help; `bd types` lists Issue classifications | `bd types [--details --json]` discovers installed Bead and Link descriptors |

Preview2 still preserves the ordinary behavior in ordinary workspaces. A graph
workspace activates a separate admitted subset; the presence of a Cobra command
or flag in help does not mean graph mode accepts it.

Sources: [release dep registrations](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/dep.go#L1529),
[preview admission](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/cmd/bd/graph_preview.go#L332),
[preview list/unlink](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/cmd/bd/graph_preview_lifecycle.go#L13),
[preview update dispatch](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/cmd/bd/graph_preview_memory.go#L13),
[guide](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/docs/reference/graph-cli.md).

### Differences hidden by the existing names

1. **`link` is not an exact alias for `dep add`.** In v1.3.1, `link --type`
   accepts arbitrary nonempty strings up to 32 bytes. `dep add --type` and
   `create --deps` accept only known types and normalize `blocked-by` and
   `depends-on` to `blocks`. `link` does not normalize those strings, so a
   custom edge spelled `blocked-by` does not acquire blocking semantics.
   `dep add` also supports external references, bulk input and target-resolution
   forms that `link` does not. The earlier conversation's shorthand description
   must therefore be qualified.
   [link](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/link.go#L13),
   [type policy](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/create_deps.go#L159).
2. **Ordinary edges and graph Links have different identities.** Ordinary
   dependencies allow one edge per ordered pair A→B; another type on that pair
   conflicts. Re-adding the same type can refresh edge metadata. Graph
   informational Links have independent, permanently allocated `links/PATH`
   identities and allow parallel Links, including the same endpoints and Type.
   Pair removal must refuse ambiguity and report matching Link IDs.
   Graph blocking Dependencies still have one live Link per ordered pair;
   reasserting that pair is a no-op that retains its current Link ID. Recreating
   the pair after unlink allocates a new Link ID, without recycling the old one.
   [ordinary identity](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/internal/storage/issueops/dependencies.go#L258),
   [graph multiedge descriptor](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/internal/storage/graphstore/informational_types.go#L56),
   [pair selection](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/internal/storage/graphstore/link_lifecycle.go#L229).
3. **Custom legacy names are not arbitrary graph Types.** Preview graph mode
   accepts blocking Issue dependencies or its installed informational Types:
   `types/preview-related-v2`, `types/example-follows`, `types/example-cites`.
   The examples exist only in workspaces initialized with them. The blocking
   Type is `types/preview-blocks-v1`. There is no Type-installation CLI, and
   graph `--type tracks`, `--type parent-child`, and unknown names refuse.
   Renaming the command does not implement these missing capabilities.
   [graph dependency admission](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/cmd/bd/graph_preview_workflow.go#L75),
   [installed Types](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/internal/storage/graphstore/informational_types.go#L16).
4. **Direction is semantic.** For a `blocks` edge A→B, A depends on B.
   For `parent-child`, CHILD→PARENT. For `tracks`, TRACKER→TRACKED.
   However, `create --deps blocks:B` reverses the usual operands: B depends on
   the newly created Issue. Bare B and `blocked-by:B` keep the opposite direction.
   This compatibility grammar must not become the new canonical grammar.
   [create dependency parsing](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/create_deps.go#L130).
5. **`related` is not `relates-to`.** `link A B -t related` makes one edge;
   `dep relate A B` makes two `relates-to` edges. Current `dep unrelate` removes
   both ordered pairs without checking their stored types. New canonical
   removal must check the selected type and must not remove a blocking edge
   merely because it occupies the same pair. The old help examples incorrectly
   omit `dep`; the commands are actually registered under `dep`.
   [implementation](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/relate.go#L44).
6. **List output changes with argument count today.** Ordinary single-ID
   `dep list --json` returns related Issue records; outgoing multi-ID listing
   returns raw edge records. Some external targets cannot appear in the joined
   single-ID view. Preview `links` returns Link records. New canonical list
   output should have one stable edge-record contract, not inherit that switch.
   [list implementation](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/dep.go#L822).
7. **Graph writes preserve ownership and history.** Informational Link edits
   change the source Memory's version when its Type owns that Link; informational
   Links from the installed Issue Type are not owned by that Issue. Blocking
   dependencies are owned by the source Issue. Link update/removal needs its
   own revision choice; blocking unlink additionally needs a source-revision
   choice. Removed Link IDs remain reserved and prior snapshots remain citable.
   Integration omits private deletion markers from public version listings.
   There is no general restore, erase, endpoint retargeting or Type-change
   operation. `bd delete links/ID` is not the implemented unlink spelling.
   [guards](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/cmd/bd/graph_preview_links.go#L57),
   [unlink](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/cmd/bd/graph_preview_lifecycle.go#L108),
   [versions](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/cmd/bd/graph_preview_versions.go#L109).

The ordinary release's 19 built-in type names are `blocks`, `parent-child`,
`conditional-blocks`, `waits-for`, `related`, `discovered-from`, `replies-to`,
`relates-to`, `duplicates`, `supersedes`, `authored-by`, `assigned-to`,
`approved-by`, `attests`, `tracks`, `until`, `caused-by`, `validates`, and
`delegated-from`. Readiness recognizes the first four; the static scheduling
cycle set is only `blocks`, `conditional-blocks`, and `parent-child`.
`parent-child` is structural and can propagate readiness constraints, rather
than simply meaning “finish the parent before starting its child.”
[Type semantics](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/internal/types/types.go#L1243).

The current graph informational Types allow only an optional string `note` in
their closed properties object; their metadata is a separate open JSON object.
They allow live local endpoints only, including informational self-Links, but
not pinned historical or foreign-Scope endpoint references. Blocking Links
require distinct live Issues. The current blocking CLI route does not accept
`--id`, `--properties`, or `--metadata`; presenting those on a common create
command must not imply every Type supports them.
[Informational descriptor](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/internal/storage/graphstore/informational_types.go#L56),
[blocking CLI](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/cmd/bd/graph_preview_workflow.go#L75).

## User-defined Types and Types as Beads

The design now needs two levels: operations on instances of a Type and
operations on the Type definition itself. Both use `create / list / show /
update / delete`, plus version discovery and comparison where supported.
These are future design requirements, not capabilities of the pinned preview.

Keep three concepts distinct:

- Resource kind determines structural inputs: a Link has source and target;
  a Bead does not acquire endpoints simply because its Type is user-defined.
- A nominal Type defines the Resource's validated properties and behavior.
  User-defined Types participate in the same generic operation path as built-ins.
- A Type definition is the object describing such a Type. If represented as a
  Bead, its own Type is a metatype; a definition describing a Link can itself
  still be a Bead. Its kind and the kind of its instances are different facts.

In the noun-scoped variation, the natural extension is:

```text
bd bead create/list/show/update/delete ...
bd link create/list/show/update/delete ...
bd type create/list/show/update/delete ...
```

The slash-separated forms above denote the common verb set, not literal CLI
syntax. `bd bead` is an extension of the broader normalization, not a command
implemented by this document. Existing top-level Bead commands can remain
aliases, just as `dep` remains an alias for Link operations. `bd type` is a
semantic view of Type definitions; it need not imply a third physical kind.
If Types become Beads, type commands share the underlying Bead operation
services and apply the appropriate definition validation and type filter.

In the generic Resource variation, instances use the existing design's generic
verbs with explicit selectors:

```text
bd create --bead-type TYPE_ID --properties JSON
bd list --bead-type TYPE_ID
bd show beads/ID
bd update beads/ID --properties JSON [guards]
bd delete beads/ID [guards]

bd create --link-type TYPE_ID --source A --target B --properties JSON
bd list --links --incident-to A --link-type TYPE_ID
bd show links/ID
bd update links/ID --properties JSON [guards]
bd delete links/ID [guards]
```

If Types become Beads, a definition can use this same generic Bead CRUD with
the metatype's ID. Do not invent that metatype ID, relocate existing `types/`
identities, or settle discovery/installation semantics in this CLI proposal.
A `bd type` convenience view can remain available while that representation
evolves. Its generic creation mapping is conditional on the eventual Type
representation; it is not currently implemented.

Generic authoring must validate against the selected descriptor, rather than
hardcode the current Issue/Memory writers or the preview's optional `note`
property. Typed conveniences may delegate to those services. New user-defined
Types should not require new compiled top-level command names.

Sharing verbs does not imply unrestricted schema mutation: `type update`
must obey the chosen Type evolution/versioning contract, and `type delete`
must obey the chosen rules for existing instances and references. Whether an
update replaces, versions, or requires a new Type identity remains a separate
Type-design decision. Current immutable preview descriptors do not decide it.

This makes the core implementation common to both variations: shared Resource
operations with Type-specific validation and policy. The choice of Link-noun
commands versus generic Resource verbs becomes a presentation decision.

## Shared proposed contract

Namespace placement should not change the graph model. Both variations use
`create / list / show / update / delete`. Variation 1 scopes them to the Link
noun; variation 2 applies generic Resource verbs to explicit Link selectors. These are proposed spellings, not commands already
available in the inspected builds.

- Creation always names SOURCE then TARGET. Preserve the existing default
  `blocks` behavior for compatibility, and document that TARGET blocks SOURCE.
  Use explicit types in examples involving informational Links.
- Keep `--type` for ordinary dependency vocabulary and `--link-type` for graph
  installed descriptors during the normalization. Never silently reinterpret
  a custom legacy name as a nominal graph Type; reject both flags together.
- Graph single-Link operations accept bare Link IDs in a Link-only command,
  or `links/PATH`/the exact local canonical Link URL. Endpoint arguments are
  Bead selectors. A command determines the namespace, never database probing.
- Ordinary show/delete select an ordered endpoint pair; graph show/update/delete
  select Link IDs. Retain graph informational pair deletion only with an
  explicit Type and exactly one match. Blocking graph deletion stays ID-only.
- Canonical list uses `--direction in|out|both`, default `both`, and returns
  edge records with an invariant shape for one or multiple anchors. Translate
  old `down→out`, `up→in` only in compatibility entry points. Do not silently
  change their defaults or JSON contracts.
  The current graph `bd links` accepts exactly one Bead. The generic Resource
  variation proposes an unanchored Link collection listing as well as endpoint
  filters; that is additional implementation, not a wrapper around the current
  incident-Link query. If multiple anchors are later accepted, evaluate
  direction relative to each anchor, union the matching edges, and emit each
  Link ID once (ordinary edges deduplicate by store plus ordered endpoint pair).
- Graph update edits admitted informational properties/metadata, not identity,
  endpoints, Type or blocking behavior. Preserve atomic property+metadata writes,
  revision checks, source ownership, no-op behavior and retained snapshots.
  Ordinary general update/history are capability gaps, not effects supplied by
  a namespace rename; do not advertise them as working there.
- Informational source acceptance defaults to unconditional; callers can opt
  into `--if-source-revision`. Update/delete still require their own Link guard
  choice. Explicit creation with `--link-type types/preview-blocks-v1` requires
  a source guard choice; the existing Issue convenience route without
  `--link-type` accepts the current source. Preserve that distinction rather
  than making a command move silently change concurrency behavior.
- For the first implementation, canonical `delete` delegates to existing unlink
  semantics, with explicit revision choices in graph mode. It is not erasure,
  restore, cascade or the existing Bead-delete preview/`--force` operation.
- Preserve distinctions between scheduling-cycle checks and arbitrary graph
  cycles. Informational graph cycles need not be errors. Graph mode's missing
  scheduling tree/cycle features remain missing until implemented.

### Variation 1: all standalone relationship operations under `bd link`

```text
bd link create SOURCE TARGET [--type NAME | --link-type TYPE] [options]
bd link list BEAD... [--direction in|out|both] [type filter]
bd link show LINK [--version TOKEN]
bd link update LINK [--properties JSON | --patch JSON] [metadata edits] [guards]
bd link delete LINK [guards]
bd link versions LINK
bd link compare LINK --from TOKEN --to TOKEN
bd link relate A B
bd link unrelate A B
bd link tree BEAD [options]
bd link cycles
bd link graph BEAD [options]
```

`show SOURCE TARGET` and `delete SOURCE TARGET` are the ordinary pair-selector
forms; graph pair deletion retains the narrower rule above. `relate/unrelate`
preserve their explicit bidirectional intent; they are not aliases for one
directed `related` edge. In graph mode those conveniences require additional
implementation and Type semantics, so they must not be assumed available.

Use `create`, `delete`, `show`, `versions` as documented names. `add`, `remove`
and `history` can be noun-local aliases. The existing `bd link A B` shorthand
can remain inside this namespace: Cobra selects a verb or alias first, then
falls back to the parent's two-endpoint legacy handler. Keep that handler's
type behavior during migration rather than silently redirecting it to a new
validator. Qualified endpoint IDs
avoid bare-ID/verb collisions. No one-argument implicit show is proposed.

Existing `bd dep`, `bd links`, `bd unlink`, graph views, and generic Link
resource reads/edits may remain as compatibility aliases/adapters. The canonical
help, examples and implementation ownership live under `bd link`; aliases share
those operation services rather than becoming independent relationship APIs.
Compatibility adapters may preserve old flag defaults and output shapes.
Generic Bead commands may still display relationships as part of a Bead view.

Examples (proposed):

```sh
# Ordinary workspace
bd link create bd-convoy bd-task --type tracks
bd link create bd-child bd-parent --type parent-child
bd link list bd-task --direction in
bd link delete bd-child bd-parent

# Preview graph workspace
bd link create policy work --link-type types/example-cites --id policy-work
bd link show policy-work --json
bd link update policy-work --properties '{"note":"reviewed"}' --unconditional
bd link delete policy-work --unconditional
bd link versions policy-work
```

**Strengths:** Link matches preview2's actual Resource name, includes both
scheduling and informational relationships, and gives discoverable noun-first
CRUD. The existing create shorthand survives in its original namespace.

**Costs:** Moves the established `dep` family and preview's generic Link read/
update/history paths. Requires explicit compatibility handling rather than
simply assigning Cobra aliases to semantically different handlers.

### Variation 2: generic Resource operations; no canonical `bd link` family

This is the current working direction. The canonical operation vocabulary is
shared across all built-in and user-defined Resources. `dep` and other
conveniences remain adapters to it.

#### Creation selects kind through Type

```text
bd create --bead-type TYPE --properties JSON [--id ID]
bd create --link-type TYPE --source BEAD --target BEAD [--properties JSON] [--id ID]
```

Exactly one Type selector is permitted. `--bead-type` selects Bead creation and
rejects either endpoint flag, even if explicitly supplied empty. `--link-type`
selects Link creation and requires both endpoints. The descriptor must have
the matching kind, and supplied properties and endpoints must validate against
it before a write. No additional `--link` or `--kind` creation selector is
needed. Supplying endpoints alone must not silently choose a default Link Type.

For compatibility, omitting both Type selectors can preserve today's default
Issue creation. Legacy `--type task` remains an Issue classification; it must
not be reinterpreted as a nominal Type selector. Link creation always explicitly
names its Type in the canonical syntax; existing `bd link A B` / `bd dep add`
can supply the historical blocking default in their adapters.

#### Listing extends the filtering model

Proposed selector and filter forms:

```text
bd list                                      # Existing default: Beads
bd list --bead-type TYPE                      # Beads of this nominal Type
bd list --kind link                          # All current Links
bd list --link-type TYPE                      # Links of this nominal Type
bd list --kind link --source BEAD             # Outgoing Links
bd list --kind link --target BEAD             # Incoming Links
bd list --kind link --incident-to BEAD        # Either endpoint
bd list --link-type TYPE --source A --target B
bd list --kind all                           # Explicit mixed Resource collection
```

`--kind bead|link|all` is a proposed filter; final spelling remains a design
choice. It describes the structural Resource kind, not a nominal Type.
`--bead-type` implies Bead kind and `--link-type` implies Link kind; either
can be used without `--kind`. Reject the two Type selectors together and reject
an explicitly incompatible kind. The mixed `--kind all` form accepts common
Resource filters only; it must not silently reinterpret a Bead-only or
Link-only filter. Mixed collection support is recommended for Resource inventory
and cross-kind filtering; it is additional implementation, not present in preview2.

The default stays the existing Bead view, including its existing Issue-filter
behavior; adding Links must not unexpectedly change existing list scripts.
Keep `--all`'s existing visibility/limit semantics separate from `--kind all`;
it is not a Resource-kind selector. This distinction needs prominent help.
Legacy Issue-specific flags such as `--status` or `--type task` cannot silently
filter out Links from a requested Link or mixed collection; reject incompatible
combinations. Type-definition Beads, if introduced, can be selected through
`--bead-type` just like any other nominal Type.

On a Link collection, `--source` and `--target` are independent exact endpoint
filters combined with AND. `--incident-to` means source OR target equals that
Bead; initially reject combining it with source/target filters to avoid unclear
intent. Require explicit Link selection (`--kind link` or `--link-type`) for
these Link-only filters; reject them for Bead or mixed listing. Existing
`bd links BEAD --direction ...` can translate to the appropriate filters.

Unanchored Link listing requires a proper bounded/paginated collection query.
It must not loop over Beads' incident Links, duplicate results, or imply that
preview2 already implements it. Keep canonical JSON record shapes independent
of filters or row count, include ID/kind/Type, and disclose truncation. A mixed
human view should show common fields with optional Type-specific summaries,
without assuming every Resource has an Issue title or status.

#### Unanchored, mixed and filtered listing

`bd list --kind link` without source, target or incident filters selects the
workspace's entire current Link collection. It does not require an anchor,
choose an implicit last-touched Bead, or traverse from anything. An empty
collection returns an empty result. Normal ordering, row limits and explicit
continuation/truncation apply. This excludes removed Links and historical
versions; retained state remains accessible through version operations.

`bd list --kind all` selects both Beads and Links. It is useful for auditing
metadata conventions, inspecting all Resources in a workspace, and generic
client tooling. It selects kinds, not unlimited output or historical records;
existing visibility and limit controls remain separate. A full inventory must
apply the desired visibility options and consume every page or explicitly
request no row limit. Mixed rows need common ID, kind, nominal Type, revision
and a Type-aware summary. Use a deterministic common sort (for example ID),
not an Issue-only priority/title sort. Do not present a mixed flat list as a
connected graph or promise that every Link endpoint is in a filtered page.

Filter layers should be shared by generic list and any query shortcut:

| Layer | Proposed use |
| --- | --- |
| Resource selection | kind, nominal Bead/Link Type, ID |
| Link structure | source, target, incident-to |
| Common metadata | metadata equality and key existence |
| Type-defined properties | property equality/existence and typed comparisons |
| Type-specific shortcuts | existing Issue status, assignee, priority, label, dates, etc., where applicable |
| Presentation | sort, limit, continuation, output format; these do not define match semantics |

Current upstream v1.3.1 already registers repeatable `--metadata-field key=value`
and `--has-metadata-key KEY` on ordinary list. Its query evaluator also supports
`metadata.KEY=value` and `has_metadata_key=KEY`; metadata comparison currently
supports equality rather than arbitrary numeric/range operators. Preview2 graph
list's flag allowlist excludes these metadata filters, and it does not provide
arbitrary Type-property filtering. Those are implementation gaps, not reasons
to exclude them from the generic design.
[Release list flags](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/list.go#L517),
[query metadata evaluation](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/internal/query/evaluator.go#L499),
[preview graph list admission](https://github.com/donnabox/beads/blob/ff25e824d80d4c1a52370c539ff8c767f1410de6/cmd/bd/graph_preview_list.go#L85).

Recommended extensions, not current capabilities:

```sh
bd list --kind all --metadata-field team=compiler
bd list --kind link --has-metadata-key reviewed_by
bd list --bead-type TYPE --property-field severity=high
bd list --link-type TYPE --has-property-key note
bd list --kind all --where 'metadata.team = "compiler" AND properties.risk >= 3'
```

`--property-field` and `--has-property-key` mirror the existing metadata forms.
`--where` is illustrative syntax for a shared predicate parser extended from
`bd query`; settle its precise grammar and nested-key addressing before
implementation. Keep compatibility with existing metadata-key interpretation.
Do not implement separate query semantics for flag filters and expressions.
Metadata and Type-defined properties are separate objects and must never be
merged into one ambiguous field namespace.

Independent filters combine with AND; an expression provides explicit OR/NOT.
Generic property paths can be applied across Types: a missing value fails a
positive equality/range match, while explicit existence predicates distinguish
absence from JSON null. Comparisons must preserve JSON value types; do not
silently compare string "3" as numeric 3. A property absent from one Resource
in a heterogeneous result is not a command error. Malformed syntax, conflicting
kind selectors, or a field/operator combination invalid for a selected known
schema is an error. Type-specific shorthand flags still require compatible
selection rather than silently hiding other kinds.

#### Graph algorithms have a graph operation family

CRUD applies to individual Resources and collections. Traversal, tree rendering
and cycle analysis operate on topology and deserve explicit graph operations:

```text
bd graph traverse BEAD --direction out --depth N [edge filters]
bd graph tree BEAD --direction out [edge filters]
bd graph cycles [--root BEAD] --edge-set all|scheduling [edge filters]
```

These are proposed canonical spellings. `bd graph BEAD` can remain a shorthand
for traversal, and `bd dep tree/cycles` remain compatibility adapters. The noun
`graph` names an operation domain, not another stored Resource kind or Type.
It does not require noun-scoped CRUD for every Type.

Reuse Link Type, metadata and property filter semantics to select traversable
edges, with explicit graph input scope. On graph commands, unqualified
`--link-type` and metadata/property filters apply to edges; any future Bead
filters must be separately scoped. Output filtering must not silently prune
intermediate traversal nodes. Tree is a projection of a graph: mark revisits,
parallel Links and cycles instead of claiming the underlying structure is a
tree or looping forever.

Generic cycle discovery defaults to all selected edge types and reports cycles
as graph facts. Informational cycles can be valid. `--edge-set scheduling`
selects the scheduling cycle predicate; the dep-cycles adapter supplies that
selection and preserves the legacy result/exit contract. Reuse current domain
semantics rather than equating every readiness-affecting edge with membership
in the static cycle set. No root means workspace-wide cycle discovery; a root
selects a declared reachable subgraph. Bounds that prevent a complete scan must
report incompleteness, never assert that the graph is acyclic.

Bidirectional relate/unrelate, parent changes and Issue duplicate/supersede
remain conveniences or domain workflows over shared operations. Neither a
rename nor graph query placement removes their existing semantic differences.

#### Identity selects reads, edits, deletion and history

```text
bd show beads/ID
bd show links/ID
bd update beads/ID --properties JSON [guards]
bd update links/ID --properties JSON [guards]
bd delete beads/ID [guards]
bd delete links/ID [guards]
bd versions RESOURCE_ID
bd compare RESOURCE_ID --from TOKEN --to TOKEN
```

An ID determines kind; the stored Resource determines Type. No extra selector
is needed, and namespace selection must not depend on probing for existence.
Bare IDs can retain the existing Bead shorthand. Generic Link deletion needs
new dispatch; graph show/update/versions/compare already admit Links. Preserve
revision, ownership, immutable endpoint/Type and history contracts. Whether
canonical delete should uniformly preview or immediately apply remains a
separate decision: current Bead delete and unlink differ there.

For ordinary dependencies without a public canonical Link ID, keep endpoint
pair selection inside the existing adapters initially. A canonical ID-oriented
show/update/delete surface in ordinary mode needs an explicit identity mapping
before it can be advertised. Do not force legacy endpoint-pair flags into the
new Resource API simply to claim name-level parity.

#### Types and shortcuts

If Type definitions become Beads, the same generic operations can create,
filter, inspect and edit them through their own metatype and Bead IDs. This
proposal does not settle that metatype, relocate existing Type identities, or
relax Type evolution constraints.

Examples (proposed, using preview Type names for illustration):

```sh
bd create --link-type types/example-cites --source policy --target work --id policy-work
bd list --link-type types/example-cites --source policy
bd show links/policy-work --json
bd update links/policy-work --metadata '{"reviewed":true}' --unconditional
bd delete links/policy-work --unconditional
```

Existing `dep`, `links`, `unlink`, and the old two-endpoint `link` spelling may
remain compatibility adapters. There are no canonical `bd dep` operations and
no canonical `bd link` subcommand family. Type-specific shortcuts can be added
later as default Type/property/endpoint selectors over the same services.
Aliases do not establish separate Type registries, mutation rules, or storage
implementations. Domain operations such as closing an Issue remain more than
simple CRUD aliases when their semantics require it.

**Strengths:** One Resource CRUD surface, reuse of preview2's generic reads,
updates and history, extensibility to user-defined Types, and a direct path to
Type definitions being Beads. Command interpretation follows Type selection
on creation, filters on listing, and identity on existing-resource operations.

**Costs:** List requires careful Resource-kind and filter rules, plus an actual
Link collection query. Ordinary dependency identity and canonical delete policy
need explicit decisions. This is a verb-first design rather than the initial
Gas City noun-before-verb shape; compatibility aliases preserve familiar entry
points without making them canonical.

### Compatibility mappings in both designs

These are semantic adapter mappings, not a claim that Cobra's simple Aliases
field can rename and rearrange every command automatically.

| Existing entry point | Variation 1 canonical operation | Variation 2 operation service |
| --- | --- | --- |
| `bd dep add A B` | `bd link create A B` | Generic Link create; adapter supplies blocking Type/defaults |
| `bd dep A --blocks B` | `bd link create B A --type blocks` | Generic Link create with source B, target A |
| `bd dep list A` | `bd link list A` | Link collection filtered by source A; legacy result projection |
| `bd links A` | `bd link list A` | Link collection filtered by either endpoint A |
| `bd dep remove A B` | `bd link delete A B` | Resolve ordinary ordered pair, then existing dependency deletion service |
| `bd unlink ID` | `bd link delete ID` | `bd delete links/ID`, preserving legacy guard/apply behavior |
| `bd dep tree A` | `bd link tree A` | `bd graph tree A`, preserving legacy direction and rendering |
| `bd dep cycles` | `bd link cycles` | `bd graph cycles --edge-set scheduling` |
| `bd link A B` | Link parent compatibility handler | Generic Link creation adapter |

The old dep-list adapter preserves outgoing-by-default behavior and its legacy
JSON projection. Canonical Resource listing returns the selected collection's
stable records. Do not silently map ordinary custom type strings onto graph
nominal Types: that requires the eventual Type installation/compatibility
mapping, not just changing command registration.

## Boundary with other commands

Retain these as domain operations rather than disguising them as Link CRUD:

- `create --parent`, `create --deps`, `create --waits-for`, and graph-plan
  creation are Issue/graph construction that includes edge writes. Preserve
  atomicity; do not replace them with a create-then-link sequence that can fail
  halfway through. `update --parent` is reparenting, including clearing/replacing
  prior parent edges, not just adding an edge.
- `duplicate --of` and `supersede --with` create an edge **and close an Issue**.
  `duplicates --auto-merge` additionally moves children. These remain Issue
  workflows (potentially `bd issue ...` in the broader normalization).
- `children`, issue list/show relationship views, ready/blocked work, gate
  creation, molecule bonding/cooking, import/export and sync are domain views
  or compound operations. Their internal edge access does not create another
  public standalone relationship API.
- Type definitions use the common verbs described above. Existing `bd types`
  can alias Type listing. Type authoring, evolution and installation remain
  explicit implementation/design work; a namespace change does not supply
  them, and their future representation as Beads remains open.

Sources: [creation flags](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/create.go#L919),
[duplicate and supersede](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/duplicate.go#L13),
[children](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/children.go#L9),
[molecule bonding](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/mol_bond.go#L18).

## Migration and implementation decisions

1. Choose the namespace separately from changing relationship semantics.
   Both variations must share generic Resource operations that support
   user-defined Types and can support Type definitions represented as Beads.
   Variation 1 retains the original noun/verb alternative; the current working
   direction is variation 2, exposing generic Resource verbs directly with
   Type-selected creation, filter-selected listing and ID-selected operations. `dep` is compatibility-only in both;
   retaining it is not an argument for either canonical design.
2. Resolve the `link`/`dep add` custom-type divergence explicitly. Recommended
   canonical ordinary create permits the existing open link vocabulary, maps
   recognized dependency aliases to their actual scheduling type, and documents
   unknown types as informational. This is an intentional policy change for
   old dep add, so its compatibility adapter must keep its old validation
   unless a separate behavior change is approved. Graph creation still requires an installed Type.
3. Retain all current scheduling safeguards, local/remote constraints and bulk
   atomicity. Do not imply ordinary external references, JSONL bulk, tracks or
   parent-child are available in graph workspaces before their writers exist.
4. Use shared command factories and operation handlers. Do not attach the same
   Cobra Command instance to two parents. Register each path explicitly; graph
   admission currently compares command identities, so new commands must be
   admitted intentionally and cannot fall through to legacy storage.
5. Preserve legacy stdout/JSON/defaults on compatibility adapters in both
   variations. Advertise a stable edge result on new canonical list/show routes. Keep graph Resource
   tokens opaque and retain ownership guards and capability errors.
6. Update `bd batch` separately: it parses its own `dep add/remove` line grammar
   instead of dispatching through Cobra. Its current type validation also differs
   from dep add. Add the chosen grammar, preserve the old grammar during migration,
   and do not lose transaction semantics. Audit completion, docs, agent prompts,
   graph help/admission, JSON consumers and Gas City provider scripts together.
   [batch parser](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/batch.go#L20).
7. Before implementation is called compatible, verify old and new routing,
   verb/ID collisions, edge direction, each backend/workspace mode, type
   validation and aliases, bulk behavior, result shapes, bidirectional removal,
   multi-Link ambiguity, stale Link/source guards, no-ops and retained history.
   Namespace normalization alone is not evidence of semantic parity.

No application code, repository branch, database state, or normative spec was
changed by this research. This document is the reviewable design artifact.
