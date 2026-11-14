# Graph query languages: references for the Beads query design

**Research date: 2026-10-19. Status: informative design survey.** No grammar, runtime, output format, or BDP amendment is selected here. Read with the [scope proposal](memory-query-scope.md) and [operation inventory](query-operation-inventory.md).

The subsequent [backend portability proposal](query-backend-portability-proposal.md) develops the recommendation with a closer SPARQL comparison and PostgreSQL/SQLite translation considerations.

**Recommendation for discussion:** use Cypher/openCypher and the newer ISO GQL direction as the main references for relationship patterns; borrow explicit traversal controls from Gremlin. Keep the simple predicate form for ordinary CLI filters. This is a design inference, not a recommendation to implement an entire existing language.

## What “commonly used” means here

The established families worth comparing are Cypher/openCypher and Gremlin for property graphs, and SPARQL for RDF. There is concrete deployed support: Amazon Neptune offers all three, and Azure Cosmos DB offers Gremlin. That establishes real product adoption, not a defensible global popularity ranking. [Neptune languages](https://docs.aws.amazon.com/neptune/latest/userguide/access-graph-queries.html), [Cosmos DB Gremlin](https://learn.microsoft.com/en-us/azure/cosmos-db/gremlin/).

ISO GQL and SQL/PGQ belong in this survey because of their standardization and design relevance. Publication of a standard is not evidence of widespread deployment. GQL is **ISO/IEC 39075:2024**; SQL/PGQ is **ISO/IEC 9075-16:2023**. [ISO GQL](https://www.iso.org/standard/76120.html), [ISO SQL/PGQ](https://www.iso.org/standard/79473.html).

## Comparison

The CLI-fit column is our assessment; capability descriptions are supported by the linked manuals below.

| Family | Relationship selection | Repeated traversal | Bindings and results | Fit for our CLI |
| --- | --- | --- | --- | --- |
| Cypher/openCypher | Visual directed patterns and node/edge predicates; pattern existence, with correlated `EXISTS` subqueries in modern Cypher | Variable-length patterns; newer Cypher has quantified patterns | Named nodes, edges and paths; separate `RETURN` | Strong reference: readable topology and familiar Boolean filters |
| Gremlin | Traversal filters, `where`, `not`; declarative `match` also exists | `repeat`, `times`, `until`, `emit`; explicit path controls | Traverser stream; `as`, `select`, `project`, `path` | Useful semantic vocabulary; long pipelines demand more procedural understanding |
| SPARQL 1.1 | Joined triple patterns, `FILTER EXISTS` / `NOT EXISTS` | Property paths, inverse and sequence; `*`, `+`, `?` | RDF-term bindings; `SELECT`, `ASK`, `CONSTRUCT` | Good reference for correlation and negation; RDF mapping adds machinery for our first-class Links |
| ISO GQL | Declarative property-graph matching | Quantified path-pattern design | Graph-element bindings and tabular results | Long-term reference; choose a documented subset rather than claim full conformance |
| SQL/PGQ | Graph matching within SQL `GRAPH_TABLE` | Graph path patterns; verify implementation coverage | Named vertices/edges, `COLUMNS`, outer SQL | Valuable integration reference; full SQL scaffolding is verbose for routine CLI use |
| GraphQL | Nested selections through schema-defined fields and arguments | General reachability needs API-specific fields/resolvers | Typed response selection and aliases | Valuable API/projection reference; not by itself a generic graph-pattern language |

## Cypher/openCypher: the closest starting point

Cypher makes the correlation visible: `(a)-[r]->(b)` binds one particular source, relationship and target together. A correlated existence test filters the outer candidate without multiplying it into one result per matching neighbor. Negation is `NOT EXISTS`. Here is a small adapted modern Cypher example, **not proposed `bd` syntax or an openCypher 9 compatibility claim**:

```cypher
MATCH (p:Person)
WHERE EXISTS {
  MATCH (p)-[r:HAS_DOG]->(d:Dog)
  WHERE r.since > 2015 AND d.name = p.name
}
RETURN p
```

Outer variables are available inside the existence expression; inner variables stay local. This directly addresses our need to test a Link and its actual neighbor together. [Cypher EXISTS](https://neo4j.com/docs/cypher-manual/current/subqueries/existential/), [MATCH and negation](https://neo4j.com/docs/cypher-manual/current/clauses/match/).

Selection and output are separate: `RETURN p`, `RETURN r`, or `RETURN p.name` chooses a node, relationship, or value. This is useful even while Cleo owns result framing and rendering. [Cypher RETURN](https://neo4j.com/docs/cypher-manual/current/clauses/return/).

Modern Cypher can bound a repeating segment with `((l)-[r:NEXT]->(m)){1,3}`. Variables within that segment become groups outside it. Older variable-length relationship syntax differs from the newer GQL-aligned form. A hop bound still does not prevent a large number of alternative paths. [Quantified paths](https://neo4j.com/docs/cypher-manual/current/patterns/variable-length-paths/).

**Dialect caveat:** current Neo4j Cypher is not identical to openCypher 9 as deployed in Neptune. The openCypher project is evolving toward GQL. Borrowing this notation would not make Beads compatible with every Cypher implementation. [Neptune dialect warning](https://docs.aws.amazon.com/neptune/latest/userguide/access-graph-opencypher.html), [openCypher direction](https://opencypher.org/).

## Gremlin: make traversal choices explicit

Gremlin composes steps over a stream. `outE` visits outgoing edges, `inV` their targets; `has` filters, `as` names bindings, and `select` projects them. Adapted examples:

```groovy
g.V().hasLabel('person').as('p').
  outE('knows').has('weight', gt(0.5)).as('r').
  inV().has('name', 'josh').as('q').select('p','r','q')

g.V(1).repeat(out()).times(2).emit().path()
```

`where` can filter by a traversal; `not` rejects candidates when its child traversal succeeds. `match` supplies declarative patterns too, so “exclusively imperative” would be misleading. `simplePath` excludes repeated objects on a traverser's path. Placement of `emit` changes whether the starting vertex is included. These are useful reminders to define zero-hop behavior, cycle rules and path output explicitly. [TinkerPop reference](https://tinkerpop.apache.org/docs/current/reference/).

Provider support varies: Cosmos DB documents unsupported `match` and other behavior differences. Reusing a language name does not establish implementation portability. [Cosmos compatibility](https://learn.microsoft.com/en-us/azure/cosmos-db/gremlin/support).

## SPARQL: strong correlation, different data model

SPARQL joins RDF subject–predicate–object patterns. A predicate variable denotes a relation IRI, not an independently identified edge instance with its own properties. Beads Links therefore need a deliberate RDF representation rather than assuming triples are interchangeable with Links.

```sparql
PREFIX foaf: <http://xmlns.com/foaf/0.1/>
SELECT ?person WHERE {
  ?person a foaf:Person .
  FILTER NOT EXISTS { ?person foaf:name ?name }
}
```

This selects people lacking a matching name pattern in the queried data. `NOT EXISTS` and `MINUS` are not interchangeable; shared bindings matter. Property paths support inverse, sequence, alternatives and repetition, but SPARQL 1.1 has no `{m,n}` repetition operator. Finite alternatives or expanded patterns can express fixed bounded cases. Property paths bind endpoints, not every traversed edge or a path object. `SELECT`, Boolean `ASK`, and graph-producing `CONSTRUCT` separate different result needs. [SPARQL 1.1 specification, especially §§8–9 and 16](https://www.w3.org/TR/sparql11-query/).

## ISO GQL and SQL/PGQ: direction, not an adoption shortcut

GQL is the standalone property-graph language standard; SQL/PGQ extends SQL for property-graph queries. Public ISO catalog pages establish status and scope; the complete paid ISO texts were **not** inspected for this survey. Syntax detail here is grounded in public implementation documentation, not a claim to exhaustively summarize the standards.

Microsoft documents a GQL implementation using `MATCH`, `WHERE` and `RETURN`; Oracle documents SQL/PGQ-style `GRAPH_TABLE`, such as this adapted example:

```sql
SELECT * FROM GRAPH_TABLE(students_graph
  MATCH (a IS person)-[e IS friends]->(b IS person)
  WHERE a.name = 'John'
  COLUMNS (a.name AS person_a, b.name AS person_b)
);
```

These reinforce named node/edge patterns, predicates, and separate projections. Exact existence and quantified-path support must be checked against the chosen standard features and implementation; it should not be inferred merely from similar arrows. For example, Microsoft's current page explicitly lists `EXISTS` and quantified paths as unsupported in that implementation. [Microsoft GQL](https://learn.microsoft.com/en-us/kusto/query/graph-query-language?view=microsoft-fabric), [Oracle GRAPH_TABLE](https://docs.oracle.com/en/database/oracle/oracle-database/26/sqlrf/graph_table-operator.html).

## GraphQL and the contributor proposal

GraphQL specifies selections over API-defined fields and arguments. Nested fields can expose relationships, but the API must define the filtering, reachability and limits; GraphQL supplies no universal `Requires` traversal operator. Its typed selections and aliases are relevant to a facade and projection design. [GraphQL introduction](https://graphql.org/learn/introduction/), [queries](https://graphql.org/learn/queries/).

The existing contributor [Beads PR #6780](https://github.com/gastownhall/beads/pull/6780) deserves review for reusable selection, projection and tests before implementation. This language survey does not assess its diff or supersede its contribution. “GQL” (ISO) and “GraphQL” are distinct names and contracts.

## Another useful comparison: recursive rules

Datalog expresses transitive closure with base and recursive rules. Soufflé's tutorial is a useful concrete example. This offers a way to distinguish stored Requires edges from derived reachability; it does not select a constraint solver or make reachability imply satisfaction. It is a conceptual comparison here, not a claim of broad product popularity. [Soufflé transitive-closure tutorial](https://souffle-lang.github.io/tutorial#transitive-closure).

## Recommendation and the next design test

**Inference:** evaluate a defined Cypher/GQL-compatible subset against our wrapper corpus before deciding between adopting that subset and extending the accepted Beads predicate language. Use Gremlin's explicit traversal choices as a checklist. Keep `bd query` as the common operation with Bead/Link population flags; wrappers lower to the same structured request. No separate wrapper-specific relationship languages.

Before spelling is fixed, test these semantic cases against the operation inventory:

1. Correlated outgoing/incoming existence and its negation, with properties on both Link and neighbor. Parallel Links must remain distinct.
2. Reachable nodes versus matching paths versus the union of traversed graph elements. Deduplicating nodes must not erase shared-parent or assessment evidence context.
3. One-to-N versus zero-to-N hops; cycles and revisits; depth, work and result bounds separately. A result `LIMIT` is not a traversal budget.
4. Select Beads because of matching Links versus return those Links. Selection and projection must remain separate, with output representation coordinated with Cleo.

Our `type is T` means refinement-family membership. Cypher labels, relationship labels/types and RDF `rdf:type` do **not** automatically carry that meaning. Preserve explicit exact-versus-family tests instead of importing a misleading shorthand.

Finally, syntax alone supplies no BDP-wide snapshot, authorization, revision, completeness or evidence-acceptance guarantee. In particular, “no matching Link found” in a bounded or partially visible read is not proof that no such Link exists. Those guarantees need their own shared contracts, independently of the chosen notation. Iris owns export preservation/packaging; this survey changes neither that boundary nor existing compatibility obligations.
