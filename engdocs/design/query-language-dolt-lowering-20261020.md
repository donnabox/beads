# Query language: Dolt lowering sketch

October 20, 2026. This supplements the [PostgreSQL/SQLite sketches](query-language-lowering.md).
It is an **unexecuted design example**, not SQL against today's physical Beads
schema or evidence that a Dolt provider passes the query corpus. It is a new
sidecar; the six-file publication checkpoint remains unchanged during recovery.

## Example: transitive requirements

Candidate query:

```text
id = "https://query.example.invalid/beads/A"
  | walk out via $r to $p where $r.type is Requires hops 1..8
  | return distinct $p
```

Use the same provider-prepared logical relations as the other sketches:

| Relation | Required meaning |
| --- | --- |
| `bead_edition` | One row per authorized identity+revision, addressed by a unique positive signed BIGINT `bead_key`. |
| `edge` | Selected source-edition/Link-edition/target-edition adjacency, with `source_key`, `target_key` and concrete `type_id`. Link identity remains separate from membership. |
| `selected_requires_family` | Distinct `child_type` identities in the selected Requires refinement closure, including Requires and Defines. This is prepared input, not an assumed installed table. |

The adapter resolves A to its selected edition key first. Prepared statement
bindings below are **seed key, maximum hops, minimum hops, maximum hops**:

```sql
WITH RECURSIVE walk(node_key, depth) AS (
  SELECT CAST(? AS SIGNED), CAST(0 AS SIGNED)

  UNION DISTINCT

  SELECT e.target_key, w.depth + 1
  FROM walk AS w
  JOIN edge AS e ON e.source_key = w.node_key
  JOIN selected_requires_family AS f ON BINARY f.child_type = BINARY e.type_id
  WHERE w.depth < ?
)
SELECT b.identity, b.revision
FROM bead_edition AS b
JOIN (
  SELECT DISTINCT node_key
  FROM walk
  WHERE depth BETWEEN ? AND ?
) AS reached ON reached.node_key = b.bead_key
ORDER BY BINARY b.identity, BINARY b.revision;
```

For the existing corpus fixture with A@a1 and range 1..3, the expected set is
**B@b1, C@c1, D@d1**. It does not switch B@b1 to current B@b2 or follow B@b2's
outgoing Link to E. A separate pin/floating fixture reaches both J@j1 and J@j2;
their distinct surrogate keys preserve both editions even though their identities
are equal.

`UNION DISTINCT` deduplicates `(node_key, depth)`, not identity alone. This avoids
duplicating the frontier for parallel Links while retaining later returns through
cycles. Depth zero participates internally; the requested hop range determines
whether it is returned. A positive cycle can still return the seed at depth two.
The final distinct-node step returns each selected edition once.

This is a reachable-set operation. Returning actual Links uses the selected
adjacency rows and preserves parallel Link editions. Enumerating paths requires
a separate path operator; the query above does not construct them.

Dolt documents CTE support; its current SQL-engine implementation has both
distinct and all recursive iterators. These source checks establish a plausible
lowering, not qualification of the version embedded in Beads. See the
[Dolt CTE support documentation](https://www.dolthub.com/docs/sql-reference/sql-support/miscellaneous/)
and the accompanying [source notes](query-language-dolt-source-notes-20261020.md).

## Type families and counts

An exact Link Type test compares the declared Type identity; Requires-family
selection joins the prepared closure and includes Defines. Canonical Type
identity comparisons must use exact/binary semantics rather than a default
case-insensitive SQL collation.

For one source edition, counts can use the same selected edge relation:

```sql
SELECT COUNT(*) AS matching_bindings,
       COUNT(DISTINCT e.edge_key) AS distinct_link_editions,
       COUNT(DISTINCT e.target_key) AS distinct_neighbor_editions
FROM edge AS e
JOIN selected_requires_family AS f ON BINARY f.child_type = BINARY e.type_id
WHERE e.source_key = ?;
```

A@a1 yields **3 bindings, 3 Link editions, 2 neighbor editions**. Across A@a1
and A@a2, an unchanged shared Link means binding count can exceed distinct Link
count. Neighbor-edition count requires complete, non-null resolution for the
requested neighbors; SQL COUNT must not silently discard unresolved targets.
A Link-only count does not require fetching target bodies.

## Optional properties, JSON types and text

Dolt lists the relevant JSON functions and binary/string operations in its
[function support reference](https://www.dolthub.com/docs/sql-reference/sql-support/expressions-functions-operators/).
These expressions assume validated JSON-object property containers:

| Query operation | Candidate Dolt expression |
| --- | --- |
| Physical key presence | `JSON_CONTAINS_PATH(properties, 'one', '$.review') = 1` |
| Missing or present JSON null | `JSON_EXTRACT(properties, '$.review') IS NULL OR JSON_TYPE(JSON_EXTRACT(properties, '$.review')) = 'NULL'` |
| Inspect value type before text comparison | `JSON_TYPE(JSON_EXTRACT(properties, '$.status')) = 'STRING'` |

For example, generic `status != "ratified"` can be expressed without losing
missing/null values to SQL's null logic:

```sql
CASE
  WHEN JSON_TYPE(JSON_EXTRACT(properties, '$.status')) = 'STRING'
    THEN BINARY JSON_UNQUOTE(JSON_EXTRACT(properties, '$.status'))
         <> BINARY ?
  ELSE TRUE
END
```

Bind `ratified`. A missing, null or differently typed value differs from that
string under the proposed dynamic equality policy. Case and trailing-space
behavior must remain exact; relying on the database's default collation is
insufficient. Do not implement no-coercion equality using an untyped SQL compare,
or introduce a conversion through permissive CAST. Value conversions are deferred.

Literal substring and full Unicode insensitive matching require additional
care. The inspected SQL-engine `LOCATE` implementation lowercases both arguments;
do **not** assume `LOCATE(BINARY needle, BINARY text)` supplies case-sensitive
contains. After a string-type check, a source-backed candidate is
`INSTR(BINARY text, BINARY ?) > 0`. Binary operands also avoid a wrapped-string
branch defect in the inspected INSTR implementation. This must still be
qualified on the actual product version; otherwise use the shared typed evaluator. Full Unicode folding also needs the selected common folding version,
not merely SQL LOWER. The source notes identify the inspected engine revision;
this is not an assertion about every installed Dolt version.

If a predicate remains in the shared evaluator, apply it before exact counts,
ordering and semantic pagination. Fetching a SQL page first and filtering it
later does not answer the same query.

## Database snapshots versus Resource revisions

Dolt's `AS OF` selects committed database state. A commit-qualified database
connection is another way to select a committed revision; a branch-qualified
connection can include working-set changes. See [Dolt history queries](https://www.dolthub.com/docs/sql-reference/version-control/querying-history/).

An immutable database commit can provide the provider's storage read context,
but it is not automatically a Bead revision token. The adapter still resolves
pinned/floating References, retained Bead editions and their owned-Link state.
One historical database snapshot does not by itself materialize every requested
Resource revision or prove access to it. The Reference inspection functions
read declared pins, not the resolved database commit.

Historical views also need care: Dolt documents that `view AS OF commit` uses
the current view definition with historical underlying rows. Fix the applicable
adapter/view contract as well as the data revision, or use the appropriately
commit-qualified schema. Do not use a branch name as an immutable cursor context
without resolving and retaining the selected state.

## Completeness and qualification

Before publishing an answer, prove the required authorized adjacency is complete
and every required selected edition exists in prepared input. Otherwise the
JOINs above could hide an unresolved target and return a smaller answer.
Preserve provider non-disclosure rules when reporting failure.

The hop predicate is a semantic bound, not a complete work budget. Work, memory,
deadline and cancellation limits remain adapter obligations; exhaustion fails
the strict request. An engine recursion limit is not evidence of query
completeness and must not silently truncate results.

No Dolt process, SQL statement, product test, build or schema migration was run
for this example. Next qualification would bind the actual product engine
revision and physical mapping, then execute the relevant authored cases under
the assigned implementation/validation lane.
