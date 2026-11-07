# Query language: relational lowering review examples

These are **unexecuted design sketches**, October 19, 2026. They explain how the
[language draft](../../docs/reference/bdp/cli/query-language.md) could map to
PostgreSQL and SQLite. Neither backend is being added, qualified or promised by
this document. The [semantic corpus](query-language-cases.json) is authored
expected behavior, not a passing parser/provider test suite.

## Logical input boundary

Assume a provider has prepared these authorized relations for **one coherent read
context**, with all requested retained editions and adjacency:

```text
bead_edition(bead_key INTEGER_KEY PRIMARY KEY, identity TEXT, revision TEXT,
             type_id TEXT, properties JSON_VALUE, metadata JSON_VALUE)
current_bead(bead_key INTEGER_KEY PRIMARY KEY)
link_edition(link_key INTEGER_KEY PRIMARY KEY, identity TEXT, revision TEXT, type_id TEXT)
adjacency(source_key INTEGER_KEY, link_key INTEGER_KEY, target_key INTEGER_KEY NULL)
edge AS SELECT l.link_key AS edge_key, l.identity, l.revision, l.type_id,
               a.source_key, a.target_key
        FROM adjacency a JOIN link_edition l USING (link_key)
direct_refinement(child_type TEXT, parent_type TEXT)
known_type(type_id TEXT)
type_name(name TEXT, type_id TEXT)
declared_reference(link_key INTEGER_KEY, endpoint TEXT, identity TEXT, revision TEXT NULL)
selected_requires_family(child_type TEXT)
```

`INTEGER_KEY` means PostgreSQL BIGINT or SQLite INTEGER, with validated positive
64-bit surrogate values. `JSON_VALUE` means PostgreSQL JSONB or SQLite validated
JSON text; these are logical relation signatures, not executable common DDL.
The provider supplies `selected_requires_family` from the applicable catalog
closure for the bound Requires identity. It is a plan input, not an undefined
SQL table or a parameterized database view. `type_name` resolves a friendly name
only when exactly one catalog identity matches.

`bead_key` is a local surrogate for an identity+revision pair, never public
identity. `link_key` similarly distinguishes Link editions, including parallel
Links. `adjacency` records selected membership separately: the same unchanged
Link edition may belong to A@a1 and A@a2. Its two membership rows do not invent
two Link-edition identities. The explanatory `edge` view joins those relations;
`edge_key` is not unique across membership rows. In that view, `source_key` identifies the source
edition whose outgoing state applies; `target_key` is the edition selected by an
exact pin or a floating Reference in this context. Stored References and their
pin state must remain available separately for `identity(ref)`/`revision(ref)`;
these resolved keys cannot answer those declared-reference functions.
`declared_reference` retains source/target identity and explicit revision or NULL
for a floating Reference. Inspection reads that authorized record without joining
the target body; for example, select its `revision` where `endpoint = 'target'`.
An external/nonfetchable target may have a NULL adjacency target key; it remains
usable for Reference-only tests, not as an invented neighbor Resource.

This relation is not a recipe for joining every historical Link with every Bead
sharing its URI. A provider must obtain correct edition-owned adjacency first.
If B@b1 is unavailable, A's pin to B@b1 fails the requested traversal rather than
joining B@b2. Different Type-definition revisions likewise need the applicable
resolved family relation; a current-only catalog does not qualify historical
family queries. The sample uses one synthetic, immutable Type catalog.

No cross-request snapshot, inaccessible Resource disclosure or BDP graph API is
invented by the SQL. Missing capabilities fail before a misleading complete
answer. Nonfetchable external endpoints can still participate in declared
Reference predicates, but do not become invented neighbor rows.

Before a binding operation publishes any answer, the provider must prove
complete authorized adjacency for that requested scope and validate every
required selected source/neighbor edition. A detection pre-pass over the edges
needed by the plan can use LEFT JOINs:

```sql
SELECT e.edge_key
FROM needed_edge e
LEFT JOIN bead_edition s ON s.bead_key = e.source_key
LEFT JOIN bead_edition t ON t.bead_key = e.target_key
WHERE s.bead_key IS NULL OR (e.needs_neighbor_body AND t.bead_key IS NULL);
```

`needed_edge` is the plan's work set with its explicit body-acquisition requirement,
not every historical edge in the store. Any row above prevents success; map the
failure using authorized, non-disclosing diagnostics. Reference-only reads need
no target-body check. Acquire/prove adjacency coverage separately: SQL absence
cannot distinguish a complete empty neighborhood from incomplete provider input.
If acquisition or execution exhausts a budget, raise failure rather than truncate
the prepared relations. No inner join below is permission to discard a required
unresolved edition. These provider obligations are not implemented by the sketch.

## Type families and correlated existence

Both databases support recursive CTEs. A reflexive closure plus `UNION` can
deduplicate Type pairs and terminate even if an invalid catalog contains a cycle;
catalog validation remains a separate responsibility.

```sql
WITH RECURSIVE initial_family(child_type, ancestor_type) AS (
  SELECT type_id, type_id FROM known_type
  UNION
  SELECT child_type, parent_type FROM direct_refinement
), family(child_type, ancestor_type) AS (
  SELECT child_type, ancestor_type FROM initial_family
  UNION
  SELECT f.child_type, d.parent_type
  FROM family f JOIN direct_refinement d ON d.child_type = f.ancestor_type
)
SELECT b.identity, b.revision
FROM bead_edition b
WHERE b.bead_key = :seed
  AND EXISTS (
    SELECT 1
    FROM edge e JOIN bead_edition p ON p.bead_key = e.target_key
    WHERE e.source_key = b.bead_key
      AND EXISTS (SELECT 1 FROM family f
                  WHERE f.child_type = e.type_id AND f.ancestor_type = :requires)
      AND EXISTS (SELECT 1 FROM family f
                  WHERE f.child_type = p.type_id AND f.ancestor_type = :policy)
  )
ORDER BY b.identity, b.revision;
```

Parameters above are explanatory notation; drivers bind their native placeholders
(`$1` in PostgreSQL, named/positional bindings in SQLite). Default collection queries separately join `current_bead`; an explicit retained
seed is not filtered through that head-only relation. The two family tests
share `e` and `p`: case `correlated-same-witness` selects no A. Two separate EXISTS
blocks can select A through different Links (`independent-exists`). Incoming
matching switches the endpoint join and `e.target_key = b.bead_key`; it does not
reverse the stored Link. `NOT EXISTS` is valid only over complete selected
adjacency. SQL absence cannot establish completeness of partially fetched inputs.

An exact Type test compares `type_id` directly. Family lookup should be materialized
or indexed appropriately, not recursively recomputed for every comparison.

## Expansion, counts and pagination

Reuse the same selected outgoing edge relation and Requires-family filter:

```sql
SELECT e.edge_key, e.identity, e.revision, p.bead_key
FROM edge e JOIN bead_edition p ON p.bead_key = e.target_key
WHERE e.source_key = :seed
  AND e.type_id IN (SELECT child_type FROM selected_requires_family);
```

A Link-only count with no neighbor predicate reads the selected Link records;
it does not add a target-body join merely to count them. For A@a1, the matching rows are ab, ab2 and ac. Count rows for three matches; count
distinct `p.bead_key` for two neighbor editions. Across A@a1 and A@a2, the
corpus has four binding rows but three distinct `edge_key` values. Use an
explicit distinct-Link-edition count when that is the question. Never replace edge count with
`count(DISTINCT target identity)`. Aggregate the entire selected input before
pagination. Order the final Resource result by canonical identity and revision
using the contract's explicit ordering, then page; an index's incidental order
does not qualify stable continuation. Locale-dependent collation must not change
canonical ordering. PostgreSQL uses `ORDER BY b.identity COLLATE "C",
b.revision COLLATE "C"`; SQLite uses `ORDER BY b.identity COLLATE BINARY,
b.revision COLLATE BINARY`. The generic ORDER BY sketches require that deliberate
provider substitution (and compatible validated Unicode inputs). A cursor also binds the authorized provider context, not just
the last identity string. No sequence of independent OFFSET queries makes an
unstable source stable.

## Bounded reachability, distinct editions and cycles

This form is intended to be portable across PostgreSQL and SQLite against the
prepared relations. It has not been executed:

```sql
WITH RECURSIVE walk(node_key, depth) AS (
  SELECT CAST(:seed AS BIGINT), 0
  UNION
  SELECT e.target_key, w.depth + 1
  FROM walk w JOIN edge e ON e.source_key = w.node_key
  WHERE w.depth < :max_hops
    AND e.type_id IN (SELECT child_type FROM selected_requires_family)
)
SELECT DISTINCT b.identity, b.revision
FROM walk w JOIN bead_edition b ON b.bead_key = w.node_key
WHERE w.depth BETWEEN :min_hops AND :max_hops
ORDER BY b.identity, b.revision;
```

`UNION` deduplicates each `(node,depth)` state; a positive cycle can still return
the root at a later depth. A globally visited-node optimization must preserve
minimum-hop semantics: merely visiting a node at depth zero cannot suppress a
qualifying return at depth two. Parallel Links are not visible in this reachable
set but remain distinct in expansion/graph results. With A and hops 1..3, the
answer is B@b1, C@c1, D@d1. B@b2 and E@e1 are not reached. With D and hops 1..2,
the answer includes D through the positive cycle.

An endpoint Type filter belongs after recursion (`post-filter-not-prune`). A
per-hop target restriction belongs inside recursion (`prune-is-different`). For
an unfiltered graph result, separately collect each selected edge expanded at a
permitted source depth, including cycle-closing edges; do not derive edges by
inducing every possible Link among returned nodes. A post-filtered witness graph
needs the unresolved witness policy before an implementation can claim support.

Finite depth does not bound work tightly. Production execution also needs an
examined-work budget/cancellation strategy. A top-level LIMIT on answer rows is
not a reliable work budget and cannot silently turn budget exhaustion into a
complete answer. This sketch supplies no production complexity guarantee.

## Simple paths are a separate operator

PostgreSQL can carry an array of local selected-edition keys and an array of
edge keys. The path row stores Link-edition surrogate keys, distinguishing parallel paths:

```sql
WITH RECURSIVE paths(node_key, seen, edges, depth) AS (
  SELECT :seed::bigint, ARRAY[:seed::bigint], ARRAY[]::bigint[], 0
  UNION ALL
  SELECT e.target_key, p.seen || e.target_key, p.edges || e.edge_key, p.depth + 1
  FROM paths p JOIN edge e ON e.source_key = p.node_key
  WHERE p.depth < :max_hops
    AND e.type_id IN (SELECT child_type FROM selected_requires_family)
    AND NOT (e.target_key = ANY(p.seen))
)
SELECT node_key, edges FROM paths WHERE depth BETWEEN :min_hops AND :max_hops;
```

These SELECTs yield internal path state, not public path payloads. Map each edge
surrogate through `link_edition`, preserving its sequence position, before
constructing an answer. A batched intermediate `path_edge(path_key, ordinal, link_key)`
can be mapped by `SELECT p.path_key, p.ordinal, l.identity, l.revision FROM
path_edge p JOIN link_edition l USING (link_key) ORDER BY p.path_key, p.ordinal`.
This ordinal order preserves each path, while final ordering among paths follows
the declared public result order rather than local surrogate values.

For SQLite, fixed positive integer surrogate keys permit delimiter-safe visited
membership. This is an internal encoding, not a substring test over public URIs:

```sql
WITH RECURSIVE paths(node_key, seen, edges, depth) AS (
  SELECT :seed, printf('/%d/', :seed), json_array(), 0
  UNION ALL
  SELECT e.target_key,
         p.seen || printf('%d/', e.target_key),
         json_insert(p.edges, '$[#]', e.edge_key), p.depth + 1
  FROM paths p JOIN edge e ON e.source_key = p.node_key
  WHERE p.depth < :max_hops
    AND e.type_id IN (SELECT child_type FROM selected_requires_family)
    AND instr(p.seen, printf('/%d/', e.target_key)) = 0
)
SELECT node_key, edges FROM paths WHERE depth BETWEEN :min_hops AND :max_hops;
```

The corpus's seven simple paths from A differ from reachability: ab/bd/db would
repeat B@b1 and is excluded. A repeated-walk mode would require an explicit finite
contract and different predicate. Never enumerate all simple paths to implement
a reachable-set request. Path/result budgets are independent of hop bounds.

## JSON values, presence and errors

PostgreSQL `jsonb` distinguishes absent key from JSON null until extraction;
SQLite `json_type` yields SQL NULL for absence and text `null` for JSON null.
These sketches intentionally demonstrate optional root reads only:

| Semantic operation | PostgreSQL | SQLite |
| --- | --- | --- |
| Physical root key | `jsonb_exists(properties, 'review')` | `json_type(properties, '$.review') IS NOT NULL` |
| Optional read is null | `properties->'review' IS NULL OR properties->'review' = 'null'::jsonb` | `json_type(properties, '$.review') IS NULL OR json_type(properties, '$.review') = 'null'` |
| String value comparison | check `jsonb_typeof(properties->'status') = 'string'`, then compare extracted text | check `json_type(properties, '$.status') = 'text'`, then compare extracted text |

Do not lower pure `status != "ratified"` directly to SQL `<>`: SQL's null
unknown result would wrongly drop missing/null matches. Lower tagged optional
values explicitly and produce a Boolean. Likewise do not use SQL COALESCE on
untyped extraction for arbitrary JSON values: SQLite false and zero, and JSON
null versus SQL NULL, need deliberate translation. Object/array equality needs
structural value semantics, not JSON source-text equality or database-specific
ordering. Object key order is irrelevant; absent field versus null remains
different in whole-object equality.

Safe nested access, required access and dynamic type errors require a typed
evaluator or guarded provider functions. SQL optimizers may reorder predicates;
`guard AND dangerous_cast(value)` is not generally a safe implementation of
language short-circuit/error semantics. Validate and use controlled evaluation
forms rather than rely on SQL WHERE order. Static errors must be detected
independently of whether a data set happens to be empty.

Literal substring lowering can use `strpos(text, needle) > 0` in PostgreSQL and
`instr(text, needle) > 0` in SQLite after type validation. `LIKE` wildcard behavior
is not literal contains. Case-insensitive matching requires the same explicit
Unicode folding/version on both sides: locale-sensitive lower()/ILIKE and
SQLite's default ASCII NOCASE are not automatically equivalent. A shared
evaluator/function or compatible precomputed index may be needed. Unsupported
matching is a capability error, not an approximate answer.

Numeric handling remains a product decision. PostgreSQL numeric and SQLite
JSON extraction/REAL do not offer identical domains or wire-lexeme preservation.
Do not parse into binary64 or let SQLite affinity coerce `"3"` into numeric 3
to claim the proposed precision rule passed. A selected portable domain or typed
decimal implementation and range diagnostics need their own acceptance cases.
Strict `number("3x")` must fail, whereas SQLite CAST could yield 3; direct CAST
therefore does not implement the proposed conversion.

## Review and implementation limits

SQL execution, query plans, provider capability negotiation, coherence acquisition,
continuations and resource budgets have **not** been qualified here. No function
registration, schema migration, SQL backend or runtime dependency is installed.
Before implementation, freeze a subset and execute its corpus through the actual
parser/evaluator and each admitted provider under Janet's qualification plan.

Primary syntax references: [PostgreSQL recursive queries](https://www.postgresql.org/docs/current/queries-with.html),
[PostgreSQL JSON functions](https://www.postgresql.org/docs/current/functions-json.html),
[SQLite WITH](https://www.sqlite.org/lang_with.html), and
[SQLite JSON functions](https://www.sqlite.org/json1.html). These references describe
database facilities, not Beads support or proof these sketches execute unchanged.
