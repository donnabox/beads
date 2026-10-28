# Preview 3: inspect a connected Bead graph

An Issue can point to a Memory, which can point to another Issue. In a link
workspace, `bd graph` follows these relationships in one graph. Select the
relationships relevant to your question:

```sh
bd graph work                       # Dependency family, both directions
bd graph work --link-type Dependency # Exactly the same selection and defaults
bd graph work --link-type References --direction out --depth 3
bd graph policy --link-type Requires --link-type Subject-To --depth 2
bd graph work --link-type any --dot > graph.dot
bd graph work --exact-link-type References
bd graph --all --link-type References --html > graph.html
```

This is a Preview 3 implementation candidate. It must be integrated and
qualified before it becomes a release capability. The dependency workspace
implementation (`--graph-mode dependency`) retains its existing behavior;
the selection flags below apply to link workspaces.

## Choose relationships

Repeat `--link-type NAME` to select a union: a Link matches if its Type refines
any selected Type. The selector also accepts `types/NAME` or a canonical Type
URL in this workspace. Bare installed names are case-insensitive; use the
exact canonical path for path/URL forms. `Requires` includes `Defines`.
Use `bd types` to inspect the actual installed catalog.

Omitting the selector selects the Dependency family. An explicit selection
replaces this default. `--exact-link-type NAME` selects that declared Type
without refinements. Repeat either flag for an OR union; `--link-type any`
selects every installed public Link Type and can join that union. The `any`
keyword is lowercase; a full Type URL selects its literal identity. An excluded Link
cannot act as a hidden bridge to another Bead.

One `--direction in|out|both` and one `--depth` apply to the whole selection.
The defaults are `both` and 1000 hops for every selector. `out` follows a
Link's authored source to its target; `in` follows the reverse. Displayed
arrows retain the authored source-to-target direction regardless of which way
you traversed the Link. They do not imply blocking or execution order.

Depth chooses the reached Beads, starting with the named root at depth zero.
The result includes **every matching Link between reached Beads**, even when
that Link was unnecessary to discover a Bead: cross-links, parallel Links,
and cycles remain visible. Cycles terminate through visited Bead identities.
A named root remains visible when no selected Link matches.

`--all` takes no root. With only the Dependency-family selector, implicit or
explicit, it retains the familiar open/in-progress/blocked native Issue
population, including isolated Issues. Other selections, including
`--exact-link-type Dependency`, include matching Links and their endpoint Beads, including
disconnected components, without unrelated isolated Beads. `--all` already
selects the entire eligible population; explicitly combining it with
`--direction` or `--depth` refuses instead of ignoring those settings.

## Read the result

Use the ordinary output options: default text, `--compact`, `--box`, `--dot`,
`--html`, and `--json`. Default text and `--compact` share a relationship tree
with actual Types, authored arrows and individually identified Links. Every
selected Link appears once, including parallel and cross-Links. A Bead expands
once, under its canonical shortest-hop parent in the admitted direction;
non-tree Links cannot change that parent. References retain identity and Type,
including for a canonical appearance later in the view. All non-tree references
say `shown elsewhere`; `cycle` additionally marks a selected directed loop.
Counts distinguish unique Beads from actual Links. At a depth limit, the view
names the reached Beads whose further relationships are outside this view,
without routine completion/frontier bookkeeping. Human views keep
literal path segments and shorten only the local Scope URL; pins and JSON
remain unchanged. Beads and relationships use `[Type]` labels. `--box` retains its flat Bead/Link view. Generic views do not
assign Memory priority, readiness or workflow status. HTML uses the existing
D3 browser visualization and requires its D3 CDN to be reachable.

`--json` uses the common graph envelope and canonical complete Bead/Link
records, with separate traversal metadata. The top-level `links` array is the
selected induced graph. A complete Bead's `owned` state is retained even when
it describes a Link outside that selection; those embedded owned-state copies
are not additional traversal edges. `--quiet` suppresses human output and
preserves JSON output. Existing format precedence remains JSON, DOT, HTML,
box, then default/compact text. No alternate-view flags are added.

`--open` remains explicitly unavailable in link mode. The native dependency
view can synthesize transitive display edges through closed Issues; generic
views preserve actual Link identities instead. Omit `--open` for this preview. `bd graph check` retains its
existing blocking-dependency consistency report.

## Limits and revision meaning

The result comes from one checked **current** inventory. Authored endpoint
revision pins stay on the Link; the adjacent Bead record remains its current
record. This is not an evaluation of an arbitrary historical graph. Use exact
[revision-aware reads](../docs/reference/bdp/cli/references.md) to inspect a pinned Bead revision.

A depth boundary can return a useful bounded graph with `complete: false` and
a `frontier` identifying reached Beads with eligible directional neighbors
beyond the boundary. `complete` is relative to the selected direction, Types
and population; it does not claim that no other relationships exist.

The default output bounds are 1000 Beads and 1000 Links; `--max-nodes` and
`--max-links` can narrow either bound. Serialized output is capped at 16 MiB. Exceeding a bound refuses the operation
without successful partial output. Acquisition still uses the provider's
complete current inventory, capped at 1000 Resources and 16 MiB, before
selection. An unrelated large population can therefore refuse this preview
operation. These flags do not promise a scalable scoped acquisition.

A positive `--max-rows` or `BEADS_MAX_ROWS` refuses because it cannot bound
that complete acquisition. `--max-rows=0` explicitly overrides the environment
cap. `--view generic` remains an optional deprecated spelling; it no longer
changes the selector or defaults. No general query language is introduced.
