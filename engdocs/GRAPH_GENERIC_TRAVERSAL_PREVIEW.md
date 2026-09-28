# Current local generic traversal preview

`bd graph --view generic` explores the current local Issue/Memory graph in a
workspace initialized with `bd init --graph-mode link`. It reads one existing
checked snapshot and prints summaries of selected Beads and independently
identified Links. No writer, SQL traversal, per-node fetch, history clock or
remote request is introduced.

```sh
bd graph beads/work --view generic --json
bd graph beads/plan --view generic --direction out --depth 2
bd graph beads/plan --view generic --direction in --depth 0 --readonly --json
```

This is a private, reversible preview of part of Proposal #6703, not adoption of
its durable output shape, external stubs or continuation policy. Ordinary
non-graph `bd graph`, its renderers, and `graph check` retain their existing
behavior. Graph workspaces require explicit `--view generic` for this route.

The root must be one live local Issue or Memory, selected by `beads/PATH` or its
exact canonical Scope URL. Link and Type roots, foreign roots, version-qualified
URLs and aliases are unavailable. A canonical but absent root returns
`not_found` (exit 3). The defaults are direction `both`, depth `1`, at most `100`
Beads and `200` Links. Depth accepts 0..1000; `--max-nodes` and `--max-links`
accept 1..1000. Invalid values/selectors refuse before graph storage opens.
The existing Cobra excess-positional-argument error is preserved; it is not a
new typed graph diagnostic.

Every changed legacy graph flag, including `--all=false`, `--compact=false`,
`--box=false`, `--dot=false`, `--html=false`, `--open=false` and `--max-rows=0`,
refuses this view with `capability_unavailable` (exit 5). Ordinary workspaces
refuse explicitly supplied generic flags, including default/empty values, before
opening legacy storage. A positive `BEADS_MAX_ROWS` refuses; unset it or use zero.
Malformed/negative environment caps return `invalid_properties` (exit 2).

## What direction, depth and complete mean

Direction follows Link orientation: `out` follows source→target, `in` reverses
it, and `both` follows either. For a blocking Dependency, `out` from the dependent
leads to its prerequisite. This is independent of native graph execution-layout
layers and does not change readiness or blocking policy.

Traversal is iterative breadth-first search. Root has distance zero; only nodes
whose shortest distance is less than the requested depth are expanded. Each
Bead is emitted once by identity and each encountered Link once by its own
identity. Parallel Links remain distinct. Self and cycle Links do not cause
repeated node expansion. Embedded `owned` Link copies are never another edge
source; the checked snapshot's standalone Link records supply adjacency.

After expansion, an included node is in `frontier` if it has a direction-eligible
incident Link that no expansion emitted. The other endpoint might already be
included: a diamond's boundary edge or a boundary self Link still matters.
No extra endpoint is emitted just to name a frontier. Depth zero emits root
only; it is incomplete whenever root has an eligible Link. A cycle whose Links
are all represented can be complete even at the depth boundary.

`complete` means frontier is empty relative to this root, direction and admitted
local graph. It does not mean the entire workspace, external graph, full Memory
model or History is implemented. Canonical percent-encoded ID spelling supplies
stable node, Link and frontier ordering, independent of source-record order.

## Summary output and refusal bounds

The existing JSON preview envelope contains a result with exactly these fields:
`projection` (`"summary"`), `scope`, `root`, `direction`, `depth`, `maxNodes`,
`maxLinks`, `nodes`, `links`, `frontier`, and `complete`.

Each node contains only `id`, `type`, `title`, `version`, and `attribution`.
Each Link contains only `id`, `type`, `source`, `target`, `version`, and
`attribution`. Current title/version/attribution values are copied, including
empty titles and unknown attribution; no timestamps are invented. No Memory
body/excerpt, Link note, metadata value, native Issue property object or Issue
long text is emitted. Reading complete content remains an explicit separate
`show` or `recall` action. Arrays are always arrays, including when empty.

Human text quotes canonical IDs and titles, identifies Links separately, and
ends with a visible `Complete: true|false; frontier: N` line followed by quoted
frontier IDs. `--quiet` suppresses human output; `--quiet --json` still emits the
structured result. Output uses the command's writer and propagates write errors,
including short writes.

Node and Link caps count distinct emitted identities; root consumes one node.
Inspecting omitted frontier Links consumes neither output allowance. Crossing a
cap refuses the entire result with `capability_unavailable` (exit 5), rather
than silently truncating or returning a continuation. Serialized output has a
separate 1MiB bound for either human or JSON mode and is fully prepared before
emission. These bounds are not exact heap guarantees. Physical writer failures
can occur after bytes have been written; such errors propagate rather than
claiming successful complete output.

The existing current-inventory limits still apply to the whole workspace:
1,000 live Bead/Link records (plus the four installed Type descriptors) and a
16MiB acquisition budget, including unrelated content.
Summary projection is an output choice, not a new authorization boundary or a
claim that full bodies were never read internally. Invalid or unsupported
records/endpoints refuse; no external references are fetched or silently dropped.
`status --graph` advertises `genericTraversal` and the operational
`genericTraversalOutputBytes`, `genericTraversalDepth`, `genericTraversalNodes`,
`genericTraversalLinks`, and `genericTraversalInventoryResources` limits. It does
not enable full Issue workflows, Memory, public Write or History profiles.

The pure tests retain diamond-edge regression intent from contributor #5283 and
direction/permutation context from #6148 without importing their held algorithms.
The early graph command dispatch is separate from #5624's native renderer work.
Public external-stub behavior, partial continuation policy, arbitrary Types,
History traversal and durable presentation decisions remain outside this slice.
Source tests and an unqualified base are not delivery evidence: installed
both-engine receipts, review, final-source baseline/lint and Linux qualification
must be recorded separately before candidate promotion.
