---
title: "Generic query language — review draft"
description: "A proposed common selection and relationship language for Beads and Links, with explicit values, bounds and result semantics."
---

## Status and purpose

Select Beads and Links with one query model, whether the request comes from
`bd query`, list/search flags or a typed library caller. A wrapper constructs a
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
GQL, GraphQL or SPARQL compatibility. PostgreSQL and SQLite are lowering targets,
not newly implemented providers. Regular expressions, arbitrary user functions,
federation, mutation syntax and a general constraint solver are outside this draft.

## Accepted direction and proposed details

| Area | Accepted direction | Candidate detail in this draft |
| --- | --- | --- |
| Population | Beads by default; explicit Beads, Links or their union | Flag registration on each admitted wrapper |
| Types | Exact affiliation versus transitive refinement-family membership | Type-literal grammar and resolution diagnostics |
| Values | Bare properties, reserved Resource fields, explicit metadata; no implicit coercion | Operator admissibility, numerical precision and function signatures |
| Absence | Optional reads coalesce absent/null; physical presence stays testable; null-only `??` | Exact safe/required access productions |
| Bindings | `$this` plus named `$` bindings; lexical scope; no accidental shadowing | Pattern and pipeline spelling |
| Relationships | Correlated existence, expansion and bounded walks | Advanced walk/graph composition and budgets |
| Results | Candidate filters do not multiply rows; preserve Link identities; distinct reachable editions; explicit paths | Row/aggregate/graph construction syntax |
| Execution | Stable ordering, aggregates before paging, explicit errors and truthful completeness | Capability admission and partial-answer contract |

The default reference for value expressions is Lumen. Beads may report expression
errors; Lumen's exception-free functional-expression rule is not inherited.
Beads-specific differences, especially JSON numerical precision, require explicit
review rather than silent import of host-language behavior.

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

The caller supplies an authorized read context and supported Resource/revision
population. Current collection candidates are the provider's selected heads.
Exact-edition reads and graph endpoints retain their declared revisions. Two
editions of one identity remain distinct when the requested operation encounters
both. Deduplication uses Resource identity **and selected revision**, never body
equality or automatic selection of the newest revision.

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

## Null, presence and access

Optional root property/metadata lookup yields null for an absent member or a
present-null member. Resource serialization still preserves the original object.

```text
status = null
status != "ratified"
(status ?? "unrecorded") = "unrecorded"
"status" in properties
```

The first predicate matches either absence or present null. The second includes
them because null is not the string `"ratified"`. The last tests physical key
presence and distinguishes those stored forms. Grouping by the read value combines
absent/null; grouping by presence and value can separate them. There is no public
`undefined`, missing-value sentinel or missing-field three-valued Boolean model.

`??` evaluates its right operand only when the left is null. False, zero and empty
strings are retained. It is right-associative and does not catch an error or unwrap
a `Result`. Boolean operators require actual Booleans and short-circuit;
no truthiness conversion is performed. Following Lumen, the checker requires a
coalescing left operand whose type admits null. Optional root reads are nullable
JSON values, not an unrestricted host-language `any` escape hatch. A known
non-null left operand is rejected before execution; short-circuiting does not
hide syntax/type errors in an unexecuted branch.
Typing optional dynamic reads as nullable JSON is a Beads-specific adaptation;
it enables this idiom without a declared property schema. Lumen itself rejects
`??` on unconstrained `any`. Admissibility on other dynamic expressions remains
part of the candidate checker contract.

**Candidate access detail:** immediate lookup under `properties`/`metadata`, or
its bare-property shorthand, is optional. Subsequent dot/bracket accesses follow
Lumen's required/safe distinction: an absent required key, invalid index or null
receiver fails; `?.` and `?.[index]` provide explicit safe access. Safe access does
not make a non-object receiver into a record. A statically known impossible
access fails before execution. Quoted bracket keys are literal; dots mean nested
objects. Array positions are zero-based and never count backward.

```text
metadata.team.name = "platform"
metadata["team.name"] = "ui"
metadata.team?.name ?? "unassigned"
tags?.[0] = "security"
```

The first two are different paths. No path implicitly flattens arrays. Candidate
details for safe access on heterogeneous dynamic values must be pinned with the
checker; output-mask missing-path no-ops are a separate representation rule.

## Comparisons, text and explicit conversion

Equality compares whole values; it never changes into substring matching for
title/body/notes. Values are type-sensitive: strings/numbers/Booleans do not
implicitly convert; numeric 3 and 3.0 are equal. Arrays are ordered and objects
ignore key order. The special absence rule applies to optional field reads, not
arbitrary raw JSON objects: `{}` and `{"x":null}` retain different field sets.

**Candidate checker policy:** reject statically disjoint literal comparisons;
when an expression can contain different JSON kinds, ordinary equality of
different present kinds is false. Null is admitted in optional equality.
Ordering requires comparable non-null numbers or strings; incompatible actual
values and invalid function arguments fail explicitly. The exact static/dynamic
admissibility matrix remains a review gate, not an excuse to coerce or silently
skip malformed values. Use a type guard or explicit conversion where needed.
Under this candidate policy, null/absent operands of ordering or `contains`
fail rather than becoming false. A wrapper can request
`priority is number and priority > 2`, or
`title is string and title contains insensitive "export"`, to skip values of
other types deliberately. String ordering follows Unicode code-point order,
not host locale. The wrapper must choose these guards explicitly.

**Candidate precision policy:** finite JSON decimal values compare without silent
rounding. A backend unable to preserve a requested comparison reports a capability
or numeric-range limitation. Adopting exact decimal arithmetic instead of Lumen's
binary64, or choosing a smaller portable domain, needs product review. The accepted
no-coercion and 3=3.0 rules do not settle this precision choice. Do not use a binary64
parse to make distinct large wire integers accidentally equal.

```text
title contains "100%_safe*"
title contains insensitive "export"
"security" in (tags ?? [])
any($x in checks where $x.passed = true)
all($x in (checks ?? []) where $x.passed = true)
number("3") = 3
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

**Candidate conversions:** reuse `number(text)`, `bool(text)` and `string(value)`
names rather than invent a generic cast. Parse the entire JSON scalar of the
requested kind, allowing only JSON whitespace around it; never accept a numeric
prefix, truthiness or locale-specific number. Unlike Lumen's nullable parsers,
strict failure here is an evaluation error: `number("3x")` fails if evaluated.
A strict parser returns a non-null value on success, so
`number("3x") ?? 0` is rejected by the non-null-left checker rule rather than
catching that failure. A nullable property read can still coalesce to a strict
parser without evaluating the parser when the property is non-null.
`bool("false")` produces false. The signatures, supported `string` conversion
domain and any separately named tolerant parser remain review details. Error
recovery syntax is not added by this draft.

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

The following is a **candidate small language**, not full Cypher. Wrappers can
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

Follow Lumen's precedence for the admitted operators: access/call, unary `not`,
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
bad dynamic values, missing required access or invalid explicit conversion.
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

[Current BDP bounded JSONPath selection](https://github.com/gastownhall/bdp/blob/b12665d64c2aaf6172bd1cdab409cef31ef28f44/specs/profiles/core/core.md)
does not define this graph language,
client-derived union, aggregation or result schema. Portable extensions require
shared-contract review; a local evaluator cannot advertise an HTTP capability
by accepting a CLI expression. Publication does not choose a runtime/spec pin.

## Review gates and acceptance artifacts

Before parser/runtime implementation, settle these candidate details:

1. Complete lexer/pipeline grammar and Lumen-compatible precedence, including
   Type operands versus primitive tests and reserved-word escaping.
2. Static/dynamic comparison admissibility, required/safe dynamic access and
   deterministic error behavior within unordered matches and ordered quantifiers.
3. Numeric precision/domain, strict parser signatures and Unicode version.
4. Post-filter graph witness semantics, per-result total orders and numerical
   budgets; repeated-walk mode remains deferred.
5. Provider capability/consistency mapping and the separate partial-answer contract.

The repository design corpus supplies synthetic selected-edition inputs and
expected semantic outcomes. It is not BDP wire JSON, an installed Memory catalog,
a product parser or proof that a backend passes. Wrapper mappings identify
supported plans versus unresolved gaps and explicit domain/external boundaries;
unimplemented expressions are not automatically approved exceptions.

Publication includes independent review findings and their dispositions. Full
navigation/site qualification and product qualification remain pending.
The first runtime handoff must name the exact language subset, input/result
schemas, provider capabilities, positive/negative cases and compatibility policy;
a grammar document alone is not implementation readiness.

Review artifacts at the fixed design checkpoint:
[semantic cases](https://github.com/donnabox/beads/blob/fbd1e4c6bf1399691c2b2136ad3756c832ac0f4c/engdocs/design/query-language-cases.json),
[case notation schema](https://github.com/donnabox/beads/blob/fbd1e4c6bf1399691c2b2136ad3756c832ac0f4c/engdocs/design/query-language-cases.schema.json),
[wrapper map](https://github.com/donnabox/beads/blob/fbd1e4c6bf1399691c2b2136ad3756c832ac0f4c/engdocs/design/query-language-wrapper-map.json),
[SQL lowering sketches](https://github.com/donnabox/beads/blob/fbd1e4c6bf1399691c2b2136ad3756c832ac0f4c/engdocs/design/query-language-lowering.md),
and [review dispositions](https://github.com/donnabox/beads/blob/fbd1e4c6bf1399691c2b2136ad3756c832ac0f4c/engdocs/design/query-language-review.md).
