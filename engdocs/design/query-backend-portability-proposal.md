# Query language and relational backend portability

**2026-10-19 — research and recommendation for discussion.** No implementation, backend qualification, grammar adoption, output contract or BDP wire change is claimed. This extends the [language survey](graph-query-language-survey.md) for the [scope proposal](memory-query-scope.md) and [wrapper inventory](query-operation-inventory.md).

**Recommendation:** define one typed query model that compiles to relational operations, and retain the accepted simple Beads predicates with a small **Cypher-style relationship-pattern fragment**. Borrow a documented fragment rather than inventing a complete language or claiming Cypher conformance. Do not adopt full SPARQL for the initial common operation. SPARQL is technically viable over PostgreSQL; the reason for this recommendation is the fit to independently identified, property-bearing Links and the already accepted CLI predicates—not a relational-backend disadvantage.

**Direction accepted by Donna on 2026-10-19.** Complete grammar and the detailed result semantics below remain under review. The isolated SQL exercise is design evidence, not backend or product qualification.

## SPARQL over SQL is a real option

SPARQL's graph-pattern joins, filters, unions, projection and aggregation align with relational operations. But solution mappings, unbound variables, expression errors, RDF terms and duplicate multiplicities need careful translation; similar keywords do not establish identical semantics. [W3C SPARQL algebra](https://www.w3.org/TR/sparql11-query/#sparqlAlgebra).

An RDF storage engine is unnecessary. W3C R2RML defines relational-to-RDF mappings and explicitly permits virtual access. Ontop implements this approach: source data stays relational while SPARQL is translated into SQL, and PostgreSQL is a documented backend. Its intermediate representation combines SPARQL and relational algebra. This is concrete feasibility evidence, not a benchmark or a recommendation to depend on Ontop. [R2RML](https://www.w3.org/TR/r2rml/#overview), [Ontop introduction](https://ontop-vkg.org/guide/), [Ontop intermediate queries](https://ontop-vkg.org/dev/internals/iq).

SPARQL's visibility in our circles is a meaningful reason to consider it: recognizable joins, variables and negation, plus an existing standards ecosystem. A bounded SPARQL profile deserves a fair comparison, rather than rejection based on its association with RDF databases.

## Compare the actual work

Effort assessments below are design inferences, not measured implementation estimates. All options need explicit JSON value semantics, authorization, revision selection, refinement-family resolution and bounded execution.

| Choice | Relational translation | Additional commitment | Assessment |
| --- | --- | --- | --- |
| Bounded Cypher-style patterns plus accepted predicates | Filters → `WHERE`; patterns → joins; existence → correlated `EXISTS`; reachability → recursive CTE or bounded executor | Specify supported syntax, scope, duplicates and path rules; map BDP values rather than inherit every Cypher behavior | Best initial fit to first-class Link variables and concise CLI filters |
| Bounded read-only SPARQL profile | Triple-pattern mappings unfold into joins over Bead/Link views; filter/exists/grouping map to relational operators | Define RDF vocabulary/term mapping, allowed forms, JSON representation and unsupported-feature errors; bounded witnessed paths require additional design | Credible if ecosystem familiarity outweighs the mapping/semantic surface |
| Full SPARQL query language | Same foundation, with a substantially broader compiler/execution contract | Dataset/named-graph behavior, RDF literals/datatypes, optional/unbound/error rules, subqueries, property paths and query forms; federation policy if exposed | Larger commitment than these wrapper requirements justify today |
| Typed query representation underneath one frontend | Backend adapters share selections, bindings, typed comparisons, existence, walks and aggregation | Design the semantic representation and capability/error rules | Recommended architecture; it does **not** by itself choose a user language or require multiple parsers |

Neither PostgreSQL nor SQLite is assumed to provide native SQL/PGQ merely because both support SQL. The proposal uses ordinary joins and recursive CTEs; extensions or alternate engines need separate evaluation.

## Preserve Link identity in either language

RDF 1.1 graphs are sets of triples. Encoding every Requires Link only as `subject :requires target` would collapse parallel Links between that pair and lose their independent identity, properties, ownership and revisions. Representing each Link as its own resource avoids that: illustrative `:link123 :source :specA; :target :policyB; :type :Requires; :evidence "…"`. This is a possible mapping vocabulary, not adopted BDP encoding. Standard reification is another modeling tool; it should not be confused with automatically asserting the reified triple. [RDF concepts](https://www.w3.org/TR/rdf11-concepts/#section-rdf-graph), [RDF reification](https://www.w3.org/TR/rdf-schema/#ch_reificationvocab).

Then a SPARQL relationship filter remains straightforward:

```sparql
# Illustrative mapped vocabulary; prefixes omitted.
SELECT DISTINCT ?s WHERE {
  ?s :type :Specification .
  FILTER EXISTS {
    ?link :source ?s ; :target ?p ; :type :Requires .
    ?p :type :Policy .
  }
}
```

A Cypher-style spelling exposes the same Link as an edge variable:

```cypher
MATCH (s:Specification)
WHERE EXISTS { MATCH (s)-[link:Requires]->(p:Policy) }
RETURN s
```

These are **exact-label toy mappings**, not our refinement-aware `type is` operation or adopted grammar. Both lower to the same correlated Link–target join. Neither automatically implements BDP refinement families. [Cypher correlated existence](https://neo4j.com/docs/cypher-manual/current/subqueries/existential/).

For recursive traversal, standard SPARQL property paths bind endpoints, not witness Link instances along the path. In the resource mapping above, `(^:source/:target)+` would also traverse every Link family, not only Requires; typed intermediate-Link filtering needs a suitable mapped relation or additional query machinery. Bounded hop ranges and witness-edge results must be explicitly designed. Ontop's published compliance table, marked for version 5.4.0, omits several repeated-path operators; that is an implementation-specific limitation, not proof SPARQL recursion cannot compile to SQL or a tested statement about 5.5.0. No Ontop SQLite support is established here. [SPARQL paths](https://www.w3.org/TR/sparql11-query/#propertypaths), [Ontop compliance](https://ontop-vkg.org/guide/compliance.html).

## What PostgreSQL and SQLite can share

Both support recursive CTEs. A conceptual depth-bounded reachability query can use this shape; placeholders are adapter parameters, and `requires_links` means the already authorized, revision-selected Requires-family relation:

```sql
WITH RECURSIVE reach(id, depth) AS (
  SELECT :root, 0
  UNION
  SELECT l.target_id, r.depth + 1
  FROM reach r JOIN requires_links l ON l.source_id = r.id
  WHERE r.depth < :max_depth
)
SELECT DISTINCT id FROM reach WHERE depth > 0;
```

This returns reachable identities, **not paths or witness Links**. The adapter must distinguish those result semantics. A positive-length cycle can return the root; excluding zero-hop rows does not exclude that case. Adding depth means `UNION` no longer removes revisits at different depths; the explicit depth predicate makes this example finite. PostgreSQL offers `SEARCH`/`CYCLE`, but those are not the shared SQLite contract. Outer `LIMIT` is not a portable traversal-work guarantee, especially with sorting or joins. [PostgreSQL recursive queries](https://www.postgresql.org/docs/current/queries-with.html#QUERIES-WITH-RECURSIVE).

SQLite permits a recursive-table row limit and documents queue order controls, but recursive terms cannot use aggregate/window functions. Its limit bounds admitted recursive rows, not every candidate scanned or generated. Aggregate after closure where appropriate. Define depth, result size, examined work, cancellation and incomplete results separately. [SQLite recursive CTEs](https://www.sqlite.org/lang_with.html#recursive_common_table_expressions).

Two portability traps need shared semantics rather than accidental backend defaults:

| Concern | Required design |
| --- | --- |
| Missing/null/type | PostgreSQL extraction can return SQL NULL for missing data; text extraction can erase JSON type. SQLite scalar extraction returns SQL values, including NULL for JSON null; `json_type` distinguishes absent/null/boolean/number. Preserve presence and value type before comparison; decide missing and error behavior explicitly. [PostgreSQL JSON](https://www.postgresql.org/docs/current/functions-json.html), [SQLite JSON](https://www.sqlite.org/json1.html). |
| Literal substring and case | PostgreSQL `strpos` and SQLite `instr` provide literal occurrence operations. Avoid unescaped SQL `LIKE`, whose `%`/`_` are patterns. PostgreSQL `ILIKE` uses locale; SQLite default case folding covers ASCII. Define common Unicode/case behavior or report a capability limit; do not silently equate those defaults. [PostgreSQL strings](https://www.postgresql.org/docs/current/functions-string.html), [SQLite functions](https://www.sqlite.org/lang_corefunc.html), [PostgreSQL matching](https://www.postgresql.org/docs/current/functions-matching.html), [SQLite LIKE](https://www.sqlite.org/lang_expr.html#like). |

Counts also need care: a join can yield one row per parallel Link while an existence filter selects each candidate once. Specify whether counting Beads, Links, endpoints, paths or distinct values; do not repair every query with an unexplained `DISTINCT`.

## A concrete next step

**Choose borrowing, not full-language adoption, for this tranche.** Keep the accepted Beads predicate shorthand; specify directional, correlated existence patterns and bounded reachability using a deliberately small Cypher-style fragment. The fragment must document its divergences and must not be advertised as accepting arbitrary Cypher/GQL. Keep full SPARQL as a deferred facade option, not a required second frontend.

Next, write a small semantic model covering population, typed fields, bindings, correlated existence/negation, one-hop expansion, bounded reachability, projection and aggregation. Use it to map a representative wrapper corpus into **both PostgreSQL and SQLite SQL sketches**, with an explicit SPARQL alternative beside the relationship cases. This tests the frontend recommendation before grammar freezes; it is design work, not a new runtime assignment.

The comparison must preserve parallel Link identities, shared/cyclic graphs, exact versus refinement-family Types, missing versus null, literal text matching, counts and incomplete/unauthorized reads. Unsupported language features should fail explicitly. Wrappers, CLI syntax and any later SPARQL facade must share this model, rather than becoming independent semantic engines.

Our accepted `type is T`, bare properties, literal `contains`, population flags and bounded witness needs are already Beads-specific choices. A SPARQL-inspired spelling with these extensions should be named a **Beads query language/profile**, not represented as unqualified SPARQL compatibility. Neither SPARQL syntax nor Cypher syntax supplies ontology inference, domain satisfaction, cross-call consistency or evidence acceptance automatically.

Finally, BDP's current bounded JSONPath selector cannot silently acquire graph joins or traversal. A portable shared mechanism requires deliberate contract work with Trish. Output framing remains Cleo's lane and export manifestation Iris's; no backend syntax choice changes those boundaries.

## First relationship cases

Donna accepted these result distinctions on 2026-10-19; additional syntax remains open. They extend the existing [wrapper coverage obligations](query-operation-inventory.md#expressibility-and-equivalence-acceptance); they are not a second normative contract. Accepted CLI semantics must eventually land in the Beads query specification, with portable mechanisms reconciled separately with BDP.

Use this small, entirely visible graph. A is a Specification; B, C and E are Requirements; D is a Policy. References is outside the Requires family; Defines is inside it.

```text
A --ab:requires----> B --bd:requires--> D
A --ab2:requires---> B                  |
A --ac:requires---> C --cd:defines----> D
                    B <--db:requires---D
A --ae:references-> E
```

The two A→B Links have different identities. B and C share D. D→B forms a cycle. Labels identify individual Links, not separate invented Link Types.

Accepted distinctions:

| Operation | Question | Expected result |
| --- | --- | --- |
| Filter candidates with correlated existence | Which Specifications directly require a Requirement? | A once, despite three matching Links |
| Expand into Links | Which Requires-family Links leave A? | ab, ab2, ac; retain both parallel Links |
| Expand into neighboring Beads | Which distinct Beads does A directly require? | B and C |
| Negate existence | Which Beads have no outgoing Requires-family Link? | E, within this complete fixture |
| Walk to depth three | Which distinct Beads are reachable from A through Requires-family Links in 1–3 hops? | B, C, D |
| Return the traversed graph | Which Links were traversable from A within that bound? | ab, ab2, ac, bd, cd, db; all retain their concrete Types |
| Filter results after walking | Which reachable Beads are Policies? | D; traversal passes through B/C |

An outgoing Requires-family count for A is three Links or two distinct neighbors. The query must identify the counting unit. An existence test does not multiply candidates. These are different operations, not duplicates to eliminate indiscriminately.

Three semantic building blocks are sufficient to describe these cases: **filter** tests a candidate without expanding the outer result; **expand** binds a Link and its actual neighbor; **walk** repeats an expansion under explicit bounds. These are explanatory names for the structured model, not selected CLI keywords. Incoming direction reverses the endpoint binding, not the Link's stored source/target or ownership.

Reachable-Bead sets, traversed graphs and individual paths remain different results. The accepted flat requirement-list view returns reachable Beads; the explanatory tree view uses a graph preserving Link identities. A renderer may show shared-node/cycle references, but cannot turn the graph into an ownership tree. Individual path enumeration needs an explicit request, depth/work/result bounds and a selected rule for repeated vertices/Links; that rule remains open.

Keep seed selection, traversal pruning and post-walk filtering separate. Filtering the output to Policy finds D. Requiring every intermediate target to be a Policy prunes B and C and cannot reach D. These questions must not share an ambiguous filter slot.

### Common SQL shapes and evidence

For this exercise only, `beads` and `links` represent already authorized, revision-selected records; `type_family` represents resolved exact/family membership. This is not a physical schema proposal. The fixture intentionally has one selected edition per identity; it does not test pinned endpoint resolution or historical graphs.

Both cases start with this conceptual relation:

```sql
WITH RECURSIVE selected_links AS (
  SELECT l.*
  FROM links l JOIN type_family f ON f.type = l.type
  WHERE f.family = 'Requires'
)
```

Append this query for correlated existence:

```sql
SELECT b.id FROM beads b
WHERE b.type = 'Specification' AND EXISTS (
  SELECT 1 FROM selected_links l JOIN beads p ON p.id = l.target_id
  WHERE l.source_id = b.id AND p.type = 'Requirement'
)
ORDER BY b.id;
```

Or extend the CTE for bounded reachability:

```sql
, reach(id, depth) AS (
  SELECT 'A', 0
  UNION
  SELECT l.target_id, r.depth + 1
  FROM reach r JOIN selected_links l ON l.source_id = r.id
  WHERE r.depth < 3
)
SELECT DISTINCT id FROM reach WHERE depth > 0 ORDER BY id;
```

The examples use the same ordinary SQL shapes intended for PostgreSQL and SQLite. Exact Bead Types in the existence example deliberately keep the fixture small; production `type is` must use resolved family membership for Beads too. Projecting a graph requires the actual traversed Links; a join between every pair of returned Beads would incorrectly add Links that the walk did not traverse.

**Evidence:** eight expected-result cases passed in Python's in-memory SQLite on 2026-10-19: existence, outgoing Link identities, distinct neighbors, negative existence, bounded reachability, first-hop depth, post-walk filtering and traversed Link identities. The scratch fixture and machine-readable results are in the local session evidence directory, outside the repository. No database persisted and no installed Beads command ran. PostgreSQL was not executed; its SQL compatibility remains a documented design target rather than a passing result. JSON/text semantics, authorization, exact revisions, partial reads, parser behavior and wrapper equivalence were not tested by this fixture.

**Next open decisions:** the [batch for Donna](memory-query-scope.md#batched-design-questions-for-donna) covers binding/scope syntax, paths/cycles, zero hops, nested fields, missing/null, typed comparisons, arrays, Unicode, incomplete reads, revisions, ordering/pagination and aggregates. Output representation remains coordinated with Cleo.
