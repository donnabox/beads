---
title: Memory relationships in Preview 3
---

Memory relationships help an agent find useful context and follow knowledge
that has been replaced. These commands require a workspace initialized with
`--graph-mode link`. They read the current graph; they do not assess whether a
policy or requirement has been satisfied.

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
