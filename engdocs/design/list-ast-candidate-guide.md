# Candidate-local List predicates

This source candidate adds `--ast` to `bd list` and `bd link list` in an existing
graph-mode `link` workspace. It reuses Query's JSON AST and predicate evaluator.
It is not yet a qualified installation or a new release. The already installed
Query developer build does not contain these new List flags.

Use a List request that returns Resources. For example, save this as
`memories.json`:

```json
{
  "ir": "bdp-query-design-1",
  "profile": "list",
  "selection": {
    "op": "filter",
    "input": {"op": "scan", "collection": "beads"},
    "bind": "item",
    "predicate": {
      "op": "contains",
      "text": {
        "op": "field",
        "of": {"op": "binding", "name": "item"},
        "path": ["properties", "body"]
      },
      "substring": {"op": "literal", "value": "release rails"}
    }
  },
  "result": {"form": "resources"}
}
```

```sh
bd list --type memory --ast @memories.json
bd list --type memory --ast - < memories.json
bd list --type memory --ast @memories.json --output-format json --include /properties/title
```

One inline JSON document is also accepted. Repeated `--ast`, trailing JSON,
unknown members and unsupported operations refuse; no last-value-wins behavior
or ignored controls. Literal field paths are arrays of exact keys and array
indexes, so dotted keys do not become accidental nested paths.

## Population and composition

`bd list` scans Beads; `bd link list` scans Links. The AST must use that same
population and the `list` profile. Predicates inspect candidate fields only.
They support the admitted comparisons, presence, literal substring search,
membership, Boolean groups and exact or family Type tests. Link source/target
references are candidate data; dereferencing their Beads, inspecting incident
Links and traversal require `bd query --ast`.

The AST predicate and existing canonical population selectors are ANDed:

- List still defaults to the Issue family. Use `--type memory` for Memories or
  `--type any` for all Beads. An AST does not silently broaden that default.
- `--type` includes refinements; `--exact-type` selects the declared Type only.
- Link List retains its existing endpoint, direction and pin applicability
  rules. These also intersect the AST selection.
- Type authority, current records and predicate input come from one checked
  snapshot. No fabricated revision or second current inventory is introduced.

For Link List, change the AST scan to `"collection": "links"`, then use, for
example, `bd link list --from i1 --ast @links.json`.

## Results, boundaries and compatibility

`--ast` opts into the existing canonical collection renderer. It keeps the
normal human default and supports its existing JSON/JSONL, `--include`, `--omit`
and delivery flags. Legacy `--json`/`--format` and Issue-specific filter flags
are not silently combined into this mode. Without `--ast`, existing List
behavior is unchanged. Dependency-mode AST execution refuses rather than
falling through to the older Issue List.

Filtering happens before delivery boundaries and projection. Results retain
complete public editions and provenance until the existing renderer applies
the requested masks. Ordering remains canonical URI code-unit order. This
candidate adds no sort or group option.

The local provider still cannot retain continuations: a complete result that
exceeds the requested page or total delivery bound refuses without emitting
items. Use `--unpaged --limit 0` or narrow the selection. The existing work,
result and mapped-input bounds also refuse; they never return a shortened list
as complete. These bounds limit execution independently of delivery flags.

The List command returns a Resource collection, not an aggregate. An AST count
request refuses here; use `bd query --ast` with `result.form = "count"` for now.
That Query request must carry its intended Type predicate explicitly because
Query does not add List's default Issue-family restriction. AST projection is
also refused; public List uses its existing `--include`/`--omit` controls.
Unknown properties retain the shared evaluator's established presence and
comparison behavior; no additional coercion, case folding or substring syntax
is introduced.

The retained List output codec and Query's newer provenance representation are
different. Until they share a representation, List AST refuses field accesses
rooted at `revisionProvenance`, `attribution`, `changeContext` or `ownedLinks`,
even in a branch that would short-circuit. Use Query for its normalized
provenance view. This does not block properties or metadata with those literal
names: `["properties", "ownedLinks"]` is an ordinary property path. Other
unknown fields keep normal missing-field semantics. List still returns its
existing provenance and owned-Link representation; this limitation applies to
predicate access, not to the existing output masks.

Execution currently uses the local graph service and embedded or server Dolt.
This does not add a PostgreSQL/SQLite adapter, remote CLI transport, SQL pushdown,
new BDP binding, pin upgrade or complete service-profile advertisement. The
[Query candidate guide](query-ast-candidate-guide.md) documents the shared AST
and the broader Query execution boundary.
