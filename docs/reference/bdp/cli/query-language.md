---
title: "Generic query language — review draft"
description: "A proposed common predicate and graph query language for Beads and Links, with explicit values, bounds and result semantics."
---

## Status and purpose

> **Current design exercise:** the [CLI surface and portable mapping](/reference/bdp/cli/profiles/query/query) follows the accepted two-surface architecture. The combined candidate below is authoring history, including superseded value/access proposals; it is not a second current grammar.

This language expresses predicates over Beads and Links, together with the
relationship operations needed to query their graph. `bd query`, list/search
flags and typed library callers use one query model. A wrapper constructs a
request; it does not invoke another CLI command or invent another evaluator.

This October 19, 2026 **design review draft** records accepted semantic direction
and proposes concrete grammar. It is not documentation of installed syntax.
Sections labeled **candidate** need review; examples use that candidate syntax.
Publishing this draft does not amend BDP, adopt a language version or qualify a
backend. Proposed diagnostic names below describe error classes, not allocated
CLI/BDP error codes or numeric exit statuses.

The [common selection/output contract](https://github.com/donnabox/beads/blob/39a5d729761fef52d1bded1311b5f103e5a0fa76/docs/reference/bdp/cli/selection-and-output.md)
owns JSON Pointer field masks, rendering, envelopes, result descriptors and delivery
framing. Its [review issue](https://github.com/gastownhall/beads/issues/7718) and
[PR122](https://github.com/donnabox/beads/pull/122) remain separate from this language.
`return` below changes the semantic result; `--include`/`--omit` only change its
representation. A field include never causes a relationship or body fetch.
This boundary uses common-contract commit `39a5d729761fef52d1bded1311b5f103e5a0fa76`;
issue7718 preserves an earlier `aeb39e76456c452742cdb0b9423a30887288dff9`
review snapshot. Neither link claims installed behavior.

The model is informed by property-graph patterns but does not claim Cypher,
GQL, GraphQL or SPARQL compatibility. Dolt, PostgreSQL and SQLite are lowering targets,
not newly implemented providers. Regular expressions, arbitrary user functions,
federation, mutation syntax and a general constraint solver are outside this draft.

## Accepted direction and proposed details

| Area | Accepted direction | Candidate detail in this draft |
| --- | --- | --- |
| Population | Beads by default; explicit Beads, Links or their union | Flag registration on each admitted wrapper |
| Types | Exact affiliation versus transitive refinement-family membership | Type-literal grammar and resolution diagnostics |
| Values | Bare properties, reserved Resource fields, explicit metadata; no implicit coercion | Operator admissibility, numerical precision and function signatures |
| Absence | If JSONPath is selected, preserve its distinction between missing and present null | Any convenience access/defaulting syntax requires an explicit mapping |
| Bindings | `$this` plus named `$` bindings; lexical scope; no accidental shadowing | Pattern and pipeline spelling |
| Relationships | Correlated existence, expansion and bounded walks | Advanced walk/graph composition and budgets |
| Results | Candidate filters do not multiply rows; preserve Link identities; distinct reachable editions; explicit paths | Row/aggregate/graph construction syntax |
| Execution | Stable ordering, aggregates before paging, explicit errors and truthful completeness | Capability admission and partial-answer contract |

This specification defines its value, presence and error rules directly. Query
expressions can report explicit evaluation errors; no host language supplies
additional syntax, conversions or exception behavior.

**Accepted architecture — October 20.** BDP owns the portable query contract,
using an explicitly bounded RFC 9535 JSONPath predicate foundation and reviewed
graph extensions. Beads owns a human-readable CLI surface that translates into
that same contract. Both frontends use the same semantic model; CLI sugar does
not introduce different missing-value, comparison, graph or failure meanings.
The existing BDP bounded Selector remains the deployed-contract input until a
reviewed successor is deliberately adopted.

The current work has two exercises: establish the graph requirements that fit
JSONPath's actual extensibility rules and document any required BDP query
composition outside them; then specify the CLI sugar and its precise lowering.
BDP prose belongs under that repository's `specs/` tree, with exact placement
coordinated with its owner. This Beads draft owns CLI spelling and the mapping,
not an independent portable graph contract. Exact graph extension names, request
shapes and complete grammar remain proposals, not installed syntax. The earlier
function-shaped rebinding macro is an illustrative dialect, not standard RFC
function semantics.

## Population and read context

Working population spelling:

```text
bd query 'type is Constraint'
bd query --beads 'type is Constraint'
bd query --links 'type is Conformance'
bd query --beads --links 'metadata.reviewed = true'
```

Supplying neither population flag selects Beads. Both select the union. Population
is independent of domain status/default filters; the pure query adds no implicit
Issue classification, readiness or closed-state exclusion. Wrappers disclose their
additional defaults in the normalized request. The union adds no stored kind
property and does not imply a universal root Type.

An ordinary query starts with the current Beads or Links in the configured
Scope, visible under the caller's existing credentials. For example,
`bd query 'type is Constraint'` filters the current visible Beads; it needs no
extra read-context expression. The provider records the read state needed to
execute that request and continue its pages coherently.

The predicate does not choose credentials, a database snapshot or a historical
revision. This draft has not selected a new CLI flag for those choices. A typed
caller can supply an explicitly supported exact-edition input, and a graph walk
can encounter a pinned endpoint. These operations retain the selected revision.
Two editions of one identity remain distinct if both are encountered;
deduplication uses Resource identity **and selected revision**, never body equality
or automatic substitution of the newest revision.

Existing endpoint pins are honored. A floating endpoint resolves under the read
context; an exact endpoint is not silently retargeted to a head. Exact traversal
must use the outgoing/owned state applicable to the selected source edition,
not today's Links attached to an old Bead body. An unchanged Link edition can
appear in more than one selected source edition's outgoing state. Preserve each
source/Link/neighbor binding without inventing another Link identity; graph
membership and Link-edition identity are separate. Missing History or a coherent
read capability is an explicit limitation, not permission to mix editions.

BDP's existing collection continuations do not establish a snapshot spanning
independent collections or requests. A provider must supply the consistency the
specific operation requires or refuse it. In particular, combining a Bead page
and Link page does not by itself prove a coherent graph or complete union.

Authorization precedes query interpretation of disclosed data. Reference
inspection can inspect only values already disclosed by an authorized Resource
response; it grants no permission to fetch a target body. Honor the provider's
non-disclosure rules: an unauthorized Resource must remain indistinguishable
from an absent one where the contract requires that. History/capability errors
must not become an existence oracle. Detailed retention diagnoses require the
corresponding disclosure permission. A match needing an unavailable neighbor
body cannot silently treat that missing binding as a negative match.

## Fields and Type tests

```text
type = Policy
type is Constraint
status = "ratified"
metadata.team = "platform"
properties.type = "internal"
```

A bare ordinary name is an optional property lookup. `status` and
`properties.status` address the same value. Reserved Resource roots are `id`,
`type`, `properties`, `metadata`, and Link-only `source`/`target`. `$this` is the
candidate. These names are resolved syntactically, without schema- or
presence-dependent fallback. Other Resource members require an explicit future
language decision; a stored property with their spelling is not reserved by
accident. Use `properties["type"]` for a colliding property.

`type = T` compares declared nominal Type identity; `type is T` includes T and
the transitive family refining T. These tests apply to Beads and Links. Type
definition revision selection is separate, as defined by the adopted Types
contract. History/family matching cannot silently switch a pinned affiliation
to the current Type definition.

**Candidate Type operands:** an unquoted identifier in the Type-test position
resolves through the declared catalog; a quoted absolute Type URI selects that
identity. For example, `type is Policy` and `type is "https://example.invalid/types/Policy"`.
Names must resolve uniquely; unknown/ambiguous names fail. A query against a known
abstract Type can validly have no exact instances. Do not rewrite arbitrary
property strings or use partial-ID matching to resolve equality operands.

This special operand rule applies only to reserved Resource Type fields, including
`$r.type`. `properties.type = "internal"` is ordinary value equality.
For ordinary expressions a bare name denotes a property, not an enum/string
literal. Strings must be quoted.

Source/target are References, not fetched Resource objects. **Candidate functions**
`identity(ref)` and `revision(ref)` inspect the declared reference without chasing
it. `identity` returns its canonical identity; `revision` returns the explicit pin
or null for a floating reference, not an invented resolved head. A bound neighbor
from a relationship match is the selected Resource and has its actual `id` and
edition. Full Reference equality, identity matching and exact-edition matching are
different tests; their mapping must follow the adopted BDP Reference contract.
Using Link-only reserved fields on a Bead is a type error; guard a mixed-population
request explicitly or split the selection.

## Missing fields, null and access

**October 20 ruling:** use JSONPath's missing-field behavior for the selected
JSONPath predicate foundation. This supersedes the earlier candidate rule that ordinary optional
reads automatically combine absence and present null. A missing path selects no
node; a present JSON null is a selected node whose value is null. The internal
JSONPath result `Nothing` is not a stored JSON value or a new literal.

Canonical JSONPath filter examples are:

```text
$[?@.properties.review]
$[?@.properties.review == null]
$[?!@.properties.review]
$[?!@.properties.review || @.properties.review == null]
```

The first tests presence, including a present false or null. The second matches
present null only. The third tests absence. The fourth explicitly selects either
absence or present null. No stored Resource fields are added or removed.

A missing member, a member step on a non-object, or an out-of-range index selects
nothing under JSONPath. The previous required/safe dot-access distinction and
`??` spelling are not part of that standard. Any retained convenience forms must
have an explicit translation; their earlier candidate rules are suspended while
the predicate foundation is settled. This ruling does not settle grouping output
for absent keys or adopt all other JSONPath features.

The rest of the textual grammar remains a candidate pending this reconciliation.
In particular, the existing field shorthand and graph binding examples are not
claims that the complete candidate already parses as standard JSONPath.

## Comparisons, text and arrays

Equality compares whole values; it never changes into substring matching for
title/body/notes. Values are type-sensitive: strings/numbers/Booleans do not
implicitly convert; numeric 3 and 3.0 are equal. Arrays are ordered and objects
ignore key order. Missing-path behavior follows the selected predicate basis. For stored JSON
objects, `{}` and `{"x":null}` retain different field sets.

**Candidate checker policy:** reject statically disjoint literal comparisons;
when an expression can contain different JSON kinds, ordinary equality of
different present kinds is false. Null is admitted in optional equality.
Ordering requires comparable non-null numbers or strings; incompatible actual
values and invalid function arguments fail explicitly. The exact static/dynamic
admissibility matrix remains a review gate, not an excuse to coerce or silently
skip malformed values. Use an explicit type guard to skip incompatible values.
Under this candidate policy, null/absent operands of ordering or `contains`
fail rather than becoming false. A wrapper can request
`priority is number and priority > 2`, or
`title is string and title contains insensitive "export"`, to skip values of
other types deliberately. String ordering follows Unicode code-point order,
not host locale. The wrapper must choose these guards explicitly.

**Candidate precision policy:** finite JSON decimal values compare without silent
rounding. A backend unable to preserve a requested comparison reports a capability
or numeric-range limitation. The current BDP draft specifies exact-decimal
equality with binary64 round-trip admission. Aligning this language's literal
and stored-value domain with that contract is an open reconciliation item;
this candidate does not expand the BDP number domain. The accepted no-coercion
and 3=3.0 rules alone do not settle admission. Never round an inadmissible large
integer into an admitted neighbor before checking it.

```text
title contains "100%_safe*"
title contains insensitive "export"
"security" in (tags ?? [])
any($x in checks where $x.passed = true)
all($x in (checks ?? []) where $x.passed = true)
```

`contains` means literal substring, case-sensitive by default; `%`, `_`, `*`
have no pattern meaning. Candidate empty-needle behavior follows literal
substring convention: every string contains the empty string. The explicit insensitive mode uses locale-independent
Unicode case folding with a shared pinned Unicode version. The candidate uses
full default folding (C+F mappings), rather than locale-sensitive lowercasing or
simple folding; the Unicode version remains a review gate. It does not remove
accents or add Unicode normalization. `bd search` chooses insensitive mode by
default when its generic path is admitted. **`contains insensitive` is candidate
spelling.** Regex is deferred. Array membership uses `in`; array `contains` is not
selected. Arrays use deep equality; object membership tests physical keys.

`any`/`all` bind one array element locally. On an empty array, `any` is false and
`all` is true. Require nonempty separately if intended. Missing/null can be made
an empty array explicitly with `?? []`. A non-array argument fails; no implicit
unwrapping or expansion occurs. `length` counts array elements or string code
points, following the selected value rules.

**Conversions are deferred.** This draft admits no `number(...)`, `bool(...)`,
`string(...)`, cast or tolerant-parser operation. Incompatible values do not
implicitly convert. Predicates can guard the expected type; normalizing stored
values is outside predicate evaluation. Any later conversion facility needs its
own signatures, failure rules and provider qualification.

## Bindings and correlated relationships

```text
type is Specification and exists {
  ($this)-[$r]->($p)
  where $r.type is Requires and $p.type is Policy
}
```

An existence match binds one actual Link and its actual neighbor. Both predicates
must hold for that same pair. The filter returns each outer candidate once even
if several Links match. Link properties and concrete Types remain available on
`$r`; a neighbor is not synthesized from a target URI's spelling.

Use `($this)<-[$r]-($p)` for incoming matches. Direction reverses which endpoint
binds the neighbor, not stored Link direction or ownership. `not exists` tests
absence in the declared authorized context; it is not proof about hidden state.
An unfinished read cannot make absence true. A provider may stop a single positive
existence check after a conclusive witness only when doing so preserves the
admitted error policy; it must still complete the requested outer selection.

**Candidate match-error policy:** for an unordered match set, an error in any
otherwise admitted match fails the strict existence operation, even if another
match is a valid witness. Apply the same rule to `exists` and `not exists`;
provider enumeration order must not choose success versus error. Early witness
exit is safe when the remaining predicate is proven total, or equivalent
validation has ruled out errors. This is a review gate, not an accepted ruling.
The corpus labels witness/error cases as candidate expectations. Array `any`/`all`
have an ordered input; their short-circuit evaluation order must likewise be
frozen before admission. SQL predicate order alone cannot supply this contract.

Bare fields inside a block continue to refer to `$this`, not the neighbor.
Use `$p.status` for neighbor status. Outer bindings are visible in a nested block;
new inner bindings do not leak. Reusing an existing binding name is an error;
explicit reuse of `$this` as the pattern source is a reference, not a declaration.

Two separate existence blocks need not share a witness. This expression is
deliberately weaker than the correlated example:

```text
exists { ($this)-[$r]->($p) where $r.type is Requires }
and exists { ($this)-[$s]->($q) where $q.type is Policy }
```

A Requires Link to a Requirement plus a different Reference Link to a Policy
satisfies the second expression, not the first. Preserve that distinction in
provider lowering and acceptance cases.

## Candidate composition and grammar

The following is a **candidate query grammar**, not full Cypher. Wrappers can
construct the corresponding operations before the textual parser is implemented.
This pipeline illustration is partial: walk, group, unnest, aggregate, graph and
path stages are discussed below and do not yet have complete productions here.

```text
predicate
  | expand ($this)-[$r]->($p) where predicate
  | where predicate
  | return distinct $p
  | order by $p.id asc
```

`expand` produces bindings for each matching Link/neighbor, retaining parallel
Links as separate rows. `return $r` returns those Links; `return distinct $p`
deduplicates neighboring Resources by selected edition. Output field masks do
neither operation. With no semantic `return`, return the distinct candidate
Resources. Candidate column/row projections must preserve their declared result
schema and the references needed to interpret rows.

Candidate predicate grammar, with lexer details omitted deliberately:

```text
predicate  := coalesce-expression
coalesce-expression := disjunction ["??" coalesce-expression]
disjunction := conjunction ("or" conjunction)*
conjunction := equality ("and" equality)*
equality   := relational [("=" | "!=") relational]
            | resource-type-field ("=" | "!=") type-operand
relational := unary [("<" | "<=" | ">" | ">=") unary]
            | unary "contains" ["insensitive"] unary
            | unary "in" unary
            | unary "is" primitive-type
            | resource-type-field "is" type-operand
unary      := "not" unary | primary
primary    := access-or-call | "(" predicate ")"
            | "exists" "{" pattern ["where" predicate] "}"
            | ("any" | "all") "(" binding "in" predicate "where" predicate ")"
pattern    := "(" outer-binding ")" "-[" new-binding "]->(" new-binding ")"
            | "(" outer-binding ")" "<-[" new-binding "]-(" new-binding ")"
```

Candidate precedence for the admitted operators is: access/call, unary `not`,
relational/membership/Type tests (including `contains`), equality, `and`, `or`, then `??` from tightest
to loosest. Type-operand alternatives are contextual to the reserved Resource
Type field, rather than competing ordinary property/string parses. A parser
must reject chained comparisons. `insensitive` modifies `contains`; it is not
a separate operator.
Use parentheses around coalescing inside comparisons, as in
`(status ?? "unrecorded") = "unrecorded"`, and around negated comparisons.
No arithmetic, object construction or user calls are implied by the expression sketch. JSON scalar,
array and object literals are admitted as constants, not executable object members.
Primitive `is` tests (`number`, `string`, `bool`, `null`, `array`, `object`) are
distinct from the reserved Resource `type is T` family test. Exact lexer rules,
reserved words, operator alternatives and complete pipeline productions remain
parser-design work before implementation; unsupported forms must fail rather
than be silently approximated.

## Walks, paths and graph results

The accepted model has three building blocks: filter, expand and bounded walk.
A repeated query walk does not make an arbitrary Link semantically transitive.
Requires implication, Subject To behavior and Conformance meaning remain in
their Type contracts; reachability alone does not establish satisfaction.

**Candidate walk spelling:**

```text
id = "https://example.invalid/beads/specification"
  | walk out via $r to $p where $r.type is Requires hops 1..8
  | where $p.type is Policy
  | return distinct $p
```

`$this` remains the seed. Inside the walk's guard, `$r` is each encountered Link
and `$p` its neighbor; the guard applies at every hop. After the walk, `$p` is
the selected endpoint; `$r` is local and unavailable rather than misleadingly
denoting only the last edge. The post-walk `where` filters endpoints after
traversal, so it can reach Policies through non-Policy intermediate Constraints.
Filtering every intermediate target to Policy is a different question.

Depth counts Links from the seed. Default minimum is one; zero-hop inclusion is
explicit. A positive-length cycle may return the seed even when zero-hop results
were excluded. Distinct reachable Resources use identity+revision; a graph retains
parallel and cycle-closing Links and the editions actually traversed. An admitted
Link mask is traversal pruning; display masks cannot erase graph membership.

**Candidate result composition:** `return graph` requests the selected roots,
matching endpoints and their required witness Links/intermediates, not every Link
between the returned nodes. A root-only empty-expansion graph is distinguishable
from no selected seeds. Whether post-filtered graph results retain all traversed
edges or only matching witnesses must be settled before this form is implemented;
this draft recommends matching witnesses while retaining selected roots. Graph
completeness is relative to the requested roots/guards/hop range, not the whole
connected component.

`return paths` is a different semantic operation. Default individual paths do not
repeat a selected Bead edition; parallel Links yield distinct edge sequences.
Thus a reachability query can find a root via a cycle while simple-path enumeration
excludes that repeated-root sequence. Repeated-vertex walks require an explicit
mode plus finite bounds; its spelling is deferred. Path enumeration must never
be an implicit cost of asking for a reachable set or a graph.

Specify hop, examined-work, resulting-Resource/Link, path and output limits
separately. A finite hop bound can still yield many paths. Default numerical
budgets are open; no hard-coded demo bound becomes a public guarantee. Unsupported
traversal or an exhausted execution budget fails the strict request instead of
returning a complete-looking prefix.

## Counting, grouping, order and continuation

**Candidate result operators** make the counting unit explicit:

```text
type is Constraint | return count($this)
id = "https://example.invalid/beads/A"
  | expand ($this)-[$r]->($p) where $r.type is Requires
  | return count($r)
id = "https://example.invalid/beads/A"
  | expand ($this)-[$r]->($p) where $r.type is Requires
  | return count(distinct $p)
```

The latter two can count three Links versus two distinct neighbors. Candidate
`count($r)` counts matching binding occurrences; `count(distinct $r)` counts
unique Link editions. Multiple selected source editions may share an unchanged
Link edition, so those units can differ even without parallel Links. Aggregation
and distinct values operate on the selected input before output pagination.
`count` of empty input is zero; a partial or inaccessible input cannot be called
an exact total. Null/presence grouping follows the value rules above.

Candidate grouping uses `| group by expression as name | return ...`.
Array-valued keys remain arrays; they do not create overlapping groups unless an
explicit `unnest` operation is requested. An overlapping label-summary wrapper
therefore requests expansion intentionally. Exact unnest/group/aggregate grammar,
empty aggregate behavior other than count, and scalar key ordering are open;
only operators with a frozen contract may be admitted by an implementation.

Resource results default to deterministic identity order, with selected revision
as a tie-breaker when necessary. Explicit sorts add stable identity tie-breakers.
Revision tokens have no chronological ordering meaning. Binding/path/row results
need their own total order descriptor; candidate paths use the sequence of exact
Link-edition keys after the seed key. No SQL execution order is a public order.

The common output contract owns page flags and delivery. A continuation fixes
population, predicates, semantic result, ordering, bounds and provider read
context. Changing representation masks does not change that query. A provider
unable to preserve a continuation's context refuses stable continuation rather
than silently starting a new read. Requesting another page cannot recompute a
count from only that page or lose undisplayed items at a smaller total cap.

## Execution, failures and capability admission

Candidate operation stages are validate/resolve, select/bind, semantic result,
aggregate/order, provider continuation, and representation/render. The normalized
request carries all semantic inputs; this is a library model, not an adopted BDP
request JSON format. Filters and renderer implementations must not reconstruct
each other's rules from printed query text.

Validation rejects unknown syntax, illegal bindings, ambiguous Type operands,
unsupported operators and known-invalid argument types. Evaluation can fail on
bad dynamic values or missing required access.
Binding/syntax/checker failures, unsupported consistency/history/precision,
exhausted work budgets, unavailable continuation contexts and protected-result
context violations are distinct failure categories. The corpus names illustrative
classes for these categories; it does not allocate protocol or numeric statuses.
Capability errors distinguish unsupported history/traversal/precision from empty
results. A read that cannot answer the requested bounded question fails explicitly.
Only a complete visible authorized search establishes `not exists` within that
scope; it says nothing about hidden Resources.

JSONL delivery is not permission to return unresolved semantic results as success.
The proposed `--allow-partial` capability needs a separate admissible-findings,
completeness and outcome/exit contract. This draft does not admit the flag or choose
its exit code. It never converts an unfinished negative test into absence or a
partial count into an exact total. Raw output has the common contract's explicitly
weaker completion evidence; no alternate error row is invented here.

No evaluation performs implicit mutation. Reading Resources, collecting a graph
or returning ephemeral assessment data does not persist it, allocate live identity,
establish evidence acceptance or create an importable transfer. Export seed
selection may share this model; preservation/admission is a separate contract.

## Compatibility and adoption

Start with a shared typed request and simple selectors. Generic list/Link-list and
supported exact counts can use it before the parser. Add predicates, then graph
operators in bounded slices. Query/search must acquire their selected semantics,
not merely a new envelope around the old Issue evaluator.

Legacy query text equality can mean substring matching, and legacy search has
other text/ID conveniences. Those are explicit compatibility work, not pure
operators. No local-only legacy mode, cutover flag or default is silently selected
here. Keep excluded Issue/Dependency paths on their existing contracts.

This rollout excludes `bd dep`, Issue/Dependency-specific operations,
`remember`/`recall`/`memories`/`forget`, and future Link/Memory facades such as
`bd link query` and `bd memory`. Their requirements may be coverage inputs;
that does not authorize command implementation. Existing generic Link readers
and writes remain in the common rollout, with their independent contracts.

[BDP's bounded JSONPath Selector](https://github.com/gastownhall/bdp/blob/3ab9cdf5d45a2ea49a84ab191eac77c9aeaa19d0/specs/profiles/core/core.md#selection)
is a string expression over each candidate Resource, not a JSON expression tree.
It admits singular paths, literals, existence, comparisons and Boolean operators;
it excludes functions, graph traversal, projection and aggregation. The same
Selector is used by collection reads and Transactional `updateWhere`/`deleteWhere`.

A common predicate basis is selected; complete extension and lowering contracts
remain design work.
The earlier candidate combined missing and null and proposed required-access
faults and nonnegative-only indices. The access section above supersedes those
proposals for the selected JSONPath foundation. Invalid dynamic ordering still needs review:
standard JSONPath yields false for incomparable ordering. Type-family tests, literal substring
matching and named graph bindings also need an explicit mapping or extension.
Do not change BDP or call these forms compatible without resolving those choices.
The current BDP number-admission rule is a further shared-contract input.

BDP does not define this graph language, a client-derived population union,
aggregation or a new result schema. Portable extensions require shared-contract
review; a local evaluator cannot advertise an HTTP capability by accepting a CLI
expression. Publication does not choose a runtime/spec pin.

## Graph extension design under discussion

The proposed direction is a small set of built-in graph operations sharing the
JSONPath predicate core. Plain functions can test already available values.
Graph special forms may use function-shaped syntax, capture predicate syntax,
introduce a documented Link/neighbor binding and lower to the shared graph plan.
They are not ordinary eager function calls or an implicit adoption of a general
macro system. Exact names and syntax are not selected.

Illustrative function-shaped special form, not adopted syntax:

```text
$[?outgoing_any(@,
  @.link.properties.status == "active" &&
  @.neighbor.properties.status == "ratified"
)]
```

The first argument is evaluated in the outer candidate context. The second is
captured as predicate syntax and evaluated for each actual outgoing binding;
inside it, `@` denotes a temporary `{link, neighbor}` record. The two tests must
hold for the same binding. A compiler lowers this to correlated graph existence
in the shared plan; it must not first evaluate the predicate on the outer Bead.
This rebinding makes the form a Beads special form, not an ordinary RFC 9535
function extension. Any additional surface sugar must lower to the same plan.
Names, nesting/outer-reference rules and graph read errors remain review items.

A per-candidate graph existence form preserves its outer Resource and tests one
correlated Link/neighbor pair. Expansion and bounded walks can instead produce
new Resources or bindings; graphs and paths have separate result contracts.
These result-producing operations need an explicit query-composition contract,
because an ordinary JSONPath filter continues to return its selected candidates.
Keep coherent read context, revision selection, completeness and bounds visible
in the graph plan. Alternative arrow/pipeline sugar is a later usability choice.

## Review gates and acceptance artifacts

Before parser/runtime implementation, settle these candidate details:

1. Predicate foundation, complete lexer/pipeline grammar and precedence, including
   Type operands versus primitive tests and reserved-word escaping.
2. Static/dynamic comparison admissibility, required/safe dynamic access and
   deterministic error behavior within unordered matches and ordered quantifiers.
3. Numeric admission/domain alignment with BDP and the Unicode version.
   Conversion functions remain deferred.
4. Post-filter graph witness semantics, per-result total orders and numerical
   budgets; repeated-walk mode remains deferred.
5. Provider capability/consistency mapping and the separate partial-answer contract.

The repository design corpus supplies synthetic selected-edition inputs and
expected semantic outcomes. It is not BDP wire JSON, an installed Memory catalog,
a product parser or proof that a backend passes. Wrapper mappings identify
supported plans versus unresolved gaps and explicit domain/external boundaries;
unimplemented expressions are not automatically approved exceptions.

Publication includes independent review findings and their dispositions. Full
live site and product-query qualification remain pending. Static navigation/link
checks are recorded in the PR.
The first runtime handoff must name the exact language subset, input/result
schemas, provider capabilities, positive/negative cases and compatibility policy;
a grammar document alone is not implementation readiness.

Historical review artifacts at the published design checkpoint (these precede
the October 20 markup pass and are not the current local acceptance corpus):
[semantic cases](https://github.com/donnabox/beads/blob/71dfe6082a37b81ee2b2c5c4d5dac6a6f17e928f/engdocs/design/query-language-cases.json),
[case notation schema](https://github.com/donnabox/beads/blob/71dfe6082a37b81ee2b2c5c4d5dac6a6f17e928f/engdocs/design/query-language-cases.schema.json),
[wrapper map](https://github.com/donnabox/beads/blob/71dfe6082a37b81ee2b2c5c4d5dac6a6f17e928f/engdocs/design/query-language-wrapper-map.json),
[SQL lowering sketches](https://github.com/donnabox/beads/blob/71dfe6082a37b81ee2b2c5c4d5dac6a6f17e928f/engdocs/design/query-language-lowering.md),
and [review dispositions](https://github.com/donnabox/beads/blob/71dfe6082a37b81ee2b2c5c4d5dac6a6f17e928f/engdocs/design/query-language-review.md).

The [Dolt lowering supplement](https://github.com/donnabox/beads/blob/71dfe6082a37b81ee2b2c5c4d5dac6a6f17e928f/engdocs/design/query-language-dolt-lowering-20261020.md)
and its [immutable source checks](https://github.com/donnabox/beads/blob/71dfe6082a37b81ee2b2c5c4d5dac6a6f17e928f/engdocs/design/query-language-dolt-source-notes-20261020.md)
are later unexecuted examples, separate from the original council snapshot.
