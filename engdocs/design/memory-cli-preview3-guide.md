# Memory relationships in Preview 3 — candidate guide

This guide describes the implementation candidate tracked in
[7759](https://github.com/gastownhall/beads/issues/7759). It is staged for the
Preview 3 release documentation; it does not describe the currently pinned
published release. Promotion into `docs/` belongs to release integration.

Memory relationships help an agent find useful context and follow knowledge
that has been replaced. These commands require a workspace initialized with
`--graph-mode link`. They read the current graph; they do not assess whether a
policy or requirement has been satisfied.

## Keep identity separate from content

A Memory's ID names it. Its title summarizes its content; changing title or body
does not change an explicitly selected ID. The compatibility flag `--key` means
`--id`, and the two flags cannot be combined.

```sh
bd remember "Use the checked dependency writer." --key dependency-policy
bd remember "Use the checked dependency writer and retain history." --id dependency-policy
bd recall dependency-policy
bd memories dependency-policy
```

The second write updates the same Memory and preserves its omitted title. On
creation, `--title` supplies the title; otherwise the first body line is trimmed,
whitespace is collapsed, and the summary is limited to 80 characters including
an ellipsis. The stored body remains exact. `memories` displays ID, title and a
body summary, and searches all three.

Without either identity flag, `remember` uses the existing content-derived key
algorithm and overwrite behavior. Bare existing key-like content recalls rather
than overwrites; use an explicit identity flag to write such content. Ordinary
`create --type memory` remains create-only and uses the common allocator
when no ID was supplied.

IDs preserve case, spaces, Unicode and punctuation. They must be nonblank UTF-8.
Private locator layout does not rename the public ID. A public ID remains
reserved after deletion; restore that identity explicitly with `update
--from-revision`, rather than creating it again. An ID already allocated to an
Issue or Link is a conflict, including after deletion.

Bare selectors that could mean both an exact public ID and a different canonical
locator refuse to guess. The existing `show --id ID` form selects literal ID data.
Canonical graph record JSON continues to carry its canonical URI and established
BDP fields; the local `memories --format records-json` summary displays the public
ID. This change adds no JSON format or protocol field.

## Find related knowledge

```sh
bd memory references design
bd memory referenced-by policy
bd memory references design --depth 3
```

`references` follows outgoing References Links. `referenced-by` follows the
same relationship in reverse. Both start with one hop and include Link Types
that refine References. The path can cross between Issues and Memories.

## Follow replacement knowledge

```sh
bd memory successor original
bd memory supersedes revised
bd memory supersedes revised --depth 3
```

`successor` follows the replacement chain, including the starting Memory.
`supersedes` shows the Memories directly replaced by the input; increase its
depth to look further back. A Memory with no replacement remains in the result.

## Find necessary constraints

```sh
bd memory requirements privacy
bd memory required-by authorize
bd memory definition privacy
bd memory defined-by authorize
```

`requirements` follows Requires Links, including Defines children, from a
Constraint. `definition` follows only Defines Links. Both follow the closure
up to 1000 hops by default. Their reverse forms start with one hop. Use
`--depth` to change the bound. Cycles in Requires remain visible and terminate;
a Defines tree has one defining parent per child and cannot contain a cycle.

The starting Bead must be a Constraint or a Type that refines it, such as a
Policy, Decision, Procedure, Specification, Requirement or Scenario. A Requires
relationship records necessity; this query does not claim satisfaction.

## Find a subject’s explicit constraints

```sh
bd memory constraints service
bd memory constraints web api
bd memory subjects privacy
```

`constraints` starts with each supplied subject’s explicit Subject To Links,
then follows the Requires and Defines relationships of those constraints.
It follows Subject To only from the original subjects. A Subject To Link
found later in the graph does not introduce another set of constraints.

`subjects` shows the Beads with direct Subject To Links to the supplied
Constraint. It does not include subjects of dependent constraints. These two
commands have fixed traversal rules and do not accept `--depth`.

The human output includes a Discovery line describing this scope. JSON records
it as `traversal.discovery`; each `constraints` root also has a `stages` object
separating its initial declarations from the Requires closure. Its
`stages.requires.depth` records that phase’s internal bound; the traversal
direction is `staged`. Direct and staged operations omit ordinary root depth
metadata and expose no depth flag. These commands
report recorded relationships. They do not discover implicit applicability or
evaluate whether a constraint is satisfied.

## Record relationships

```sh
bd memory reference design --to policy --title "Relevant policy"
bd memory supersede original --to revised
bd memory requires privacy --to authorize
bd memory defines privacy --to retention
bd memory subject-to service --to privacy
```

All five authoring commands also accept `SOURCE TARGET` or `--from SOURCE
--to TARGET`. Supplying the same endpoint both positionally and by flag is
an error.

Each command uses the checked Link writer. The Link Type determines ownership
and source revision effects: a bare References Link does not revise its source.
`supersede` retires the old Memory by creating its owned Superseded By Link to
the replacement. It refuses a
second successor, including an identical one; it never retargets an existing
Link or changes either Memory’s status automatically. The result includes the
new Link identity.

Use `--title`, `--body` and `--status` for common Link properties. Additional
properties go in `--properties`; informational annotations go in `--metadata`.
Supplying a property in both a named flag and `--properties` is an error.

```sh
bd memory reference design --to policy --status reviewed \
  --metadata '{"reviewer":"team"}'
```

The commands reuse the shared Link identity allocator. `--id` supplies an
explicit Link identity; `--if-source-revision` guards the owning source against
concurrent changes where the Link Type declares source ownership. Relationship validation still applies: Requires cycles
are allowed, while Defines forbids a second defining parent and all-Defines
cycles.

## Read the result

The default output is a relationship tree. Link arrows keep their authored
direction, including in reverse queries. Parallel Links remain separate. A
repeated Bead points to its other occurrence rather than expanding forever.
Cycles are marked explicitly. Bead and Link Types are shown using their actual
names; display abbreviations do not change stored identities.

Supply several roots to get a separate tree and depth budget for each:

```sh
bd memory references design policy --depth 2
```

The footer counts unique Bead and Link identities across all root views. Each
root reports whether more relationships lie beyond its selected depth. The
selected graph includes matching Links between the Beads reached within that
root, even when a Link points opposite to the traversal direction.

Use `--json` or `--output-format=json` for the canonical CLI graph document.
Its Bead and Link records retain complete properties, canonical identities,
revisions, endpoint references and ownership. Per-root traversal context records
membership and frontiers; an owned Link inside a Bead is not automatically a
selected traversal edge. `--omit=envelope` returns the graph data directly.
Resource field masks are not yet supported by these graph commands.

The default bounds are 100 unique Beads and 200 unique Links across the query.
Use `--max-nodes` and `--max-links` to raise either bound, up to 1000. Exceeding
a bound refuses the query rather than returning a successful partial graph.
Selected endpoint pins must match the captured current revision; a historical
pin is refused rather than silently replaced by today's record.

Existing `remember`, `recall`, `memories` and `forget` retain their contracts.
The native dependency-mode commands keep their existing behavior.
