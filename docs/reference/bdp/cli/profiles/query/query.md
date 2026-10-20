---
title: "Query CLI syntax and portable mapping — design exercise"
description: "Readable bd query syntax over the proposed portable BDP JSONPath and graph contract."
---

Status: design exploration, October 20, 2026. This note does not amend BDP,
freeze CLI grammar, publish a protocol schema or document installed behavior.
Donna selected the separation: portable BDP predicates/graph operations and
readable Beads CLI syntax with one meaning. JSONPath is the selected predicate foundation, including its missing-field
semantics. The CLI must preserve the admitted JSONPath comparison behavior;
the earlier independent value/error rules do not carry forward automatically.

## Scope, authority and release boundary

This exercise concerns the generic query design review, not a release commitment.
It proposes CLI spelling and records its required translation. The eventual BDP
Query specification should own portable semantics and its operation/extension
contract. The Beads CLI specification should own sugar and its exact mapping.
The [common selection/output contract](https://github.com/donnabox/beads/pull/122)
continues to own representation, framing and field masks. Its current reference
snapshot is [3fd5e4356](https://github.com/donnabox/beads/blob/3fd5e4356ed9f9096ac22e7356809f317c885287/docs/reference/bdp/cli/selection-and-output.md),
including CLI/result versioning, error framing, completion and legacy/canonical
output flags. This citation does not move a runtime BDP contract pin. No second output schema or runtime implementation is introduced.
The earlier combined language draft is retained as authoring history; this
facade exercise is the current CLI surface and mapping proposal.

The normative destinations are therefore two coordinated specifications, with
paired semantic acceptance cases. This working note is not another authority.
The portable proposal is [BDP Query](https://github.com/gastownhall/bdp/blob/3e470bb35c20e075aac55afb8821f48b82c595d6/specs/profiles/query/query.md),
a proposed design exercise under `specs/profiles/query/query.md`. All graph plans
below remain explanatory notation until the portable carrier is selected.

## Recommendation

Keep the approachable predicate spelling plus the original arrow/pipeline style.
Do not carry forward the earlier draft's entire expression language automatically.
The CLI parser produces typed query operations and JSONPath predicates. The
provider consumes that model directly or receives its portable BDP serialization.
Do not implement a second value evaluator for the CLI.

The first admitted slice should cover:

1. Resource selection: property/root access, literals, comparisons, Boolean
   combinations, presence and explicit literal substring tests.
2. Exact Type and refinement-family tests after unambiguous catalog resolution.
3. Correlated one-hop `exists`/`not exists`, incoming and outgoing.
4. Expansion, explicit Link versus neighbor results, distinct selected editions
   and exact counts of the selected complete input.
5. Bounded reachable-set walks, with separate traversal guards and endpoint filters.

Nested correlation belongs in the acceptance model from the start even if a
provider refuses that capability initially. Graph/path results, grouping and
array expansion are separately admitted slices. Their accepted requirements
must survive the redesign; their syntax need not be part of the first parser.
No syntax implies a provider supports an operation.

## Sugar rules

| CLI construct | Translation obligation |
| --- | --- |
| `status` | Resource-relative `@.properties.status` |
| `metadata.team` | Resource-relative `@.metadata.team` |
| `properties["type"]` | Exact property lookup, unaffected by reserved `type` |
| `id`, `type`, `properties`, `metadata`, `source`, `target` | Reserved Resource roots, never presence-dependent fallback |
| `=` / `and` / `or` / `not` | JSONPath `==` / `&&` / `||` / `!`, with explicit syntax-tree translation |
| `exists(review)` | Physical node existence, not Boolean truthiness |
| `type = Policy` | Resolve catalog operand and compare declared Type identity under the adopted Type/Reference contract |
| `type is Constraint` | Resolve catalog operand and perform refinement-family membership |
| `title contains "abc"` | A specified literal substring function extension; no regex or SQL wildcard interpretation |
| `title contains insensitive "abc"` | Separately specified insensitive literal operation |
| `$p.status` | Selected Resource binding `p`, followed by its `properties.status` |
| `exists { (source)-[link]->(neighbor) where ... }` | Correlated existence over one actual Link/neighbor pair |
| `expand` / `walk` / `return` | Explicit query-plan operations, not output masks |

All ordinary string operands require quotes. There is no field-dependent switch
from equality to substring matching, unquoted enum values, implicit coercion,
relative-date magic, automatic newest-revision substitution or hidden Issue
population filter. Catalog Type operands are a narrow syntactic exception to
ordinary name lookup. Unknown/ambiguous names fail during resolution; quoted
absolute Type URIs are exact catalog operands. A stored property string is not
silently resolved as a Type name.

The compiler must map operator precedence structurally. It must not concatenate
user text into a generated JSONPath, SQL or shell command. Diagnostics should
retain source spans so an error in the normalized predicate points back to the
friendly expression.

Presence is a predicate context: `exists(review)` includes present null and false.
`review = null` selects present null, not absence. Missing paths and incomparable
ordering must ultimately follow the selected BDP JSONPath profile; the earlier
candidate's dynamic type-error behavior is not preserved by implication. An
implementation can diagnose an invalid expression, unsupported feature, graph
read failure or exhausted budget without redefining JSONPath comparison truth.

## Correlation without a second predicate grammar

RFC 9535 allows `@` to refer only to the immediately enclosing filter's current
node. It does not supply arbitrary named outer-variable references. The previous
`outgoing_any(@, predicate)` rebinding illustration is therefore not an ordinary
function-extension contract. Do not call it standards-conforming merely because
its punctuation looks like a function.

One candidate bridge is for the graph plan to expose a JSON binding frame to an
ordinary predicate evaluator. A frame can retain the seed and every lexically
visible binding:

```json
{"this": {"properties": {"team": "platform"}},
 "r": {"properties": {"status": "active"}},
 "p": {"properties": {"team": "platform"}}}
```

A predicate over a one-element collection of these frames is standard JSONPath:

```text
$[?@.r.properties.status == "active" &&
    @.p.properties.team == @.this.properties.team]
```

This frame is an illustration, not a proposed public Resource shape or final
BDP wire form. Resources keep their original fields; the frame adds no stored
properties. A provider can avoid materializing the frame when its lowering is
equivalent. The graph exercise must select whether such a frame, explicit
captures, supplied query descriptors or another standards-compatible mechanism
is the portable contract.

In CLI syntax, bare fields always address `$this`; introducing `$p` never
changes their referent. Nested scopes retain existing bindings, reject accidental
shadowing and discard new names on block exit. A traversal step has separate
bindings for the original seed and the current hop; hiding the latter distinction
in `@` would lose capability. Walk syntax needs one explicit current-source
binding if predicates compare each edge's source with its target.

## Twelve paired examples

All CLI invocations below are hypothetical. `Filter`, `Exists`, `Expand`, `Walk`,
`Project` and `Count` are readable plan notation, not adopted BDP keywords/JSON.
Each JSONPath is a complete filter over its indicated Resource/frame collection.
These are semantic pairings, not proven wire-equivalent encodings. This exercise
uses direct Resource/alias frames such as `@.p.properties`; the BDP exercise also
illustrates edition wrappers such as `@.neighbor.resource.properties` and a
`$.bindings` envelope. The input/capture schema must be settled and reconciled
before either becomes a public serialization or a compiler equivalence claim.
`bdp_contains` and `bdp_contains_ci` are candidate extension names, not registered
RFC functions. Type tests remain plan operations until the Type extension and
actual serialized affiliation path are fixed; assuming every affiliation is a
plain string would be unsafe.

### 1. Ordinary properties

```sh
bd query 'status = "open" and priority <= 2'
```

Resource predicate: `$[?@.properties.status == "open" && @.properties.priority <= 2]`.
Plan: select default Beads, filter, return selected Resources. No implicit closed
filter and no Issue-only classification.

### 2. Reserved-name collision and metadata

```sh
bd query 'properties["type"] = "internal" and metadata.team = "platform"'
```

Resource predicate: `$[?@.properties["type"] == "internal" && @.metadata.team == "platform"]`.
The property called `type` is not the Resource's Type affiliation.

### 3. Presence versus explicit null

```sh
bd query 'exists(review) and review = null'
bd query 'not exists(review)'
```

Resource predicates: `$[?@.properties.review && @.properties.review == null]`
and `$[?!@.properties.review]`. The first includes only present null; the second
includes only absence. No new stored missing value is introduced.

### 4. Literal substring

```sh
bd query 'title contains "100%_safe*"'
```

Resource predicate: `$[?bdp_contains(@.properties.title, "100%_safe*")]`.
Plan: ordinary filter with a declared function-extension capability. `%`, `_`
and `*` are literal characters. Default is sensitive. The insensitive spelling
maps to `bdp_contains_ci`, whose Unicode/version rules remain a contract gate.
A missing/non-string argument behavior must be specified in that extension;
the old draft's error rule is not silently imported.

### 5. Exact and family Type tests

```sh
bd query 'type = Policy'
bd query --links 'type is Requires'
```

Plan: `NominalTypeEquals(this, resolved Policy identity)` versus
`TypeFamily(this, resolved Requires identity)`. The portable predicate/extension
mapping remains pending; both use the same selected Types contract and revision
rules. The CLI catalog convenience is resolved before provider execution.

### 6. Correlated outgoing relationship

```sh
bd query 'exists {
  ($this)-[$r]->($p)
  where $r.type is Requires and $p.status = "ratified"
}'
```

Plan: `Exists(outgoing(this), bind r,p, TypeFamily(r,Requires) AND predicate)`.
Frame predicate for the ordinary property clause:
`$[?@.p.properties.status == "ratified"]`.
The same actual pair satisfies both clauses. Multiple witnesses do not multiply
outer candidate Resources.

### 7. Incoming relation and outer-value capture

```sh
bd query 'exists {
  ($this)<-[$r]-($p)
  where $r.status = "active" and $p.team = $this.team
}'
```

Plan: `Exists(incoming(this), bind r,p, predicate)`.
Frame predicate: `$[?@.r.properties.status == "active" && @.p.properties.team == @.this.properties.team]`.
Direction changes the bound neighbor, not ownership or stored Link direction.

### 8. Nested correlation

```sh
bd query 'exists {
  ($this)-[$r]->($p)
  where exists {
    ($p)-[$s]->($q)
    where $q.team = $this.team and $q.owner = $p.owner
  }
}'
```

Plan: nested `Exists`, with `{this,r,p}` captured into `{this,r,p,s,q}`.
Inner frame predicate:
`$[?@.q.properties.team == @.this.properties.team && @.q.properties.owner == @.p.properties.owner]`.
This demonstrates retained outer bindings using ordinary singular JSONPaths.
The provider cannot replace it with independent uncorrelated existence tests.

### 9. Preserve Links or deduplicate neighbors

```sh
bd query 'id = "https://example.org/beads/A"
 | expand ($this)-[$r]->($p) where $r.type is Requires
 | return $r'
```

Seed predicate: `$[?@.id == "https://example.org/beads/A"]`.
Plan: `Filter -> Expand(outgoing,Requires) -> Project(r)`.
Parallel Link identities remain distinct. Replacing the final stage with
`return distinct $p` projects/deduplicates neighboring editions by identity and
revision. Masks cannot turn one operation into the other.

### 10. Bounded reachability, then endpoint filtering

```sh
bd query 'id = "https://example.org/beads/A"
 | walk out via $r to $p where $r.type is Requires hops 1..8
 | where $p.status = "ratified"
 | return distinct $p'
```

Seed predicate is the preceding example. Plan:
`Walk(outgoing, Requires edge guard, min=1,max=8) -> Filter(endpoint predicate) -> DistinctEdition(p)`.
Endpoint frame predicate: `$[?@.p.properties.status == "ratified"]`.
The endpoint predicate is not an every-hop pruning rule. Unratified intermediates
can lead to ratified endpoints. `$r` is hop-local, unavailable after the walk.

### 11. Exact counts with explicit units

```sh
bd query 'id = "https://example.org/beads/A"
 | expand ($this)-[$r]->($p) where $r.type is Requires
 | return count($r)'
```

Plan: prior expansion followed by `Count(binding occurrences of r)`.
`count(distinct $r)` instead counts unique Link editions, while
`count(distinct $p)` counts unique neighboring editions. These can differ.
Aggregation precedes pagination and requires complete selected input. RFC
JSONPath `count()` counts a nodelist; it does not by itself define these graph
units or complete distributed execution.

### 12. Graph versus paths

```sh
bd query 'id = "https://example.org/beads/A"
 | walk out via $r to $p where $r.type is Requires hops 1..8
 | return graph'
```

Plan: `Walk -> GraphResult(selected roots, admitted witness edges/endpoints)`.
Changing only the final stage to `return paths` requests bounded individual
edge-sequence witnesses, not a different field mask. The graph preserves parallel
and cycle-closing Links. Default simple paths do not repeat one selected Bead
edition; parallel edges create distinct paths. Exact graph witness retention
following endpoint filters remains unresolved, so this example has no post-filter.

## Full requirement inventory and admission gates

| Requirement | Translation obligation / status |
| --- | --- |
| Beads, Links, union population | Query-level selection; neither flag means Beads; no fabricated stored kind property |
| Ordinary values and nested paths | JSONPath predicates, including selected missing/null and incompatible-comparison behavior |
| Physical key presence | Existence of selected node, not value truthiness |
| Bare roots and collisions | Fixed syntactic mapping; explicit property access always available |
| Exact Type, refinement family, Type URI resolution | Typed operation or declared extension; no partial-ID/string rewriting |
| Literal text and insensitive mode | Typed declared extension; regex remains deferred |
| Array membership, any/all, deep equality | Need explicit standard predicate/extension mapping; no old syntax carried forward by accident |
| Object/array literals in comparisons | Earlier broad literal grammar is not automatically an RFC JSONPath facility; inventory and decide required cases |
| Coalescing/safe access | Earlier spellings suspended; absence/null convenience needs a specified translation if retained |
| No conversions | Deferred; no hidden casts or host-language coercion |
| Correlated one-hop, incoming/outgoing, negation | Graph operation plus predicate over same binding; no synthetic neighbor from URI spelling |
| Nested outer binding and local scopes | Retained frame/capture contract; reject accidental shadowing; no scope leakage |
| Walk seed, hop source, edge, neighbor, endpoint | Distinct named roles; guard versus post-filter distinction explicit |
| Bounds, cycles, zero hops | Query-plan contract, separate work/result/path budgets; no false completeness after exhaustion |
| Exact endpoint pins and historical source state | Read context/provider resolution; identical across surfaces, no new pin syntax selected |
| Parallel Links and shared Link edition across source editions | Preserve binding occurrences and membership separately from Link identity |
| Distinct Resource/Link identity | Identity plus selected revision, not body equality or forced head substitution |
| Resource, binding, graph, path results | Semantic result construction, separate from renderer/field mask |
| Counts, groups, overlapping summaries | Units explicit; groups/unnest need later precise mapping and grammar |
| Stable order, continuation, aggregate-before-page | Shared contract and query identity; masks cannot alter semantic continuation |
| Authorization and unavailable bodies/history | Provider non-disclosure rules; unavailable binding is not an invented negative |
| Errors | JSONPath expression semantics preserved; syntax/admission/provider completeness failures remain explicit |
| Transactional mutations | Reuse portable per-Resource predicate core only; graph mutation targeting needs its own admission/atomicity contract |
| Pure reads and ephemeral evaluation graph | No mutation, persistence, allocated live identity or accepted evidence implied |

## Compatibility exceptions to record separately

Inspection source: the current `internal/query/parser.go`, `lexer.go` and
`evaluator.go` in the private publication source tree. Its AST has comparison,
AND, OR and NOT nodes; its evaluator targets Issue fields/IssueFilter and
field-specific predicate functions. This is an implementation comparison,
not an executed product qualification.

| Current behavior | Treatment in the new model |
| --- | --- |
| `status=open` unquoted values | Explicit legacy input translation or rejection; no parse-dependent guessing |
| Field names case-insensitive except metadata suffix | Explicit legacy exception; generic JSON member names remain exact |
| `title=export`, description/notes equality as insensitive substring | Legacy equality maps intentionally to literal-insensitive operation; pure `=` stays equality |
| `none`, `null`, empty string interpreted specially for assignee/labels/etc. | Explicit field-specific legacy mapping, not generic missing/null semantics |
| Boolean strings `yes`, `1`, `no`, `0` | Legacy parsing exception; no generic coercion |
| `id=prefix*`, `spec=prefix*` | Legacy prefix convenience; no generic wildcard equality |
| `updated>7d` and day-based date equality | Legacy temporal lowering with frozen reference time and explicit comparison contract; not JSONPath magic |
| Known Issue aliases (`desc`, `label`, `created_at`, etc.) | Wrapper/legacy alias map; no claim they are arbitrary Resource properties |
| Label equality means membership | Legacy operation mapping; generic array equality stays whole-value equality |
| Metadata scalar textual comparison | Legacy adapter behavior; no automatic JSON value/string equivalence |
| Existing Issue/default closed filter | Wrapper-specific disclosed population policy; pure query has none |

No predicate-language compatibility mode/default/cutover is selected here.
Output-format transition follows the common contract cited above. Existing
remember/recall/memories/forget and Issue/Dependency-specific commands remain
outside this rollout. Future Link/Memory facades will conform when implemented;
this note does not authorize implementation of them.

## Validation path and remaining decisions

Required paired cases should run the same selected graph through (a) friendly
parse/normalize and (b) portable form/normalize, compare semantic plans where
canonicalization is defined, and compare actual results, multiplicity, revision
identity, error/completeness and ordering against one common evaluator. Test
missing versus explicit null; false-valued presence; incomparable ordering;
reserved-name collision; malformed/ambiguous Type operands; same-witness versus
independent witnesses; outer captures; zero-hop/cycle; pinned old source state;
parallel Links; and count-before-pagination. Also reject unsupported syntax
rather than accept a partial approximation.

The parser/AST, binding checker, normalization, provider admission/lowering,
completion and diagnostic source maps, CLI help, editor support and legacy
translation cases all need corresponding implementation work. These examples
have not been parsed or executed. No Go/build/test/runtime files changed in this
exercise.

Selected direction: two surfaces, one query model, and JSONPath predicate
semantics for the admitted foundation. Proposed sugar: retain the short predicates and
arrow/pipeline graph forms, while pruning independent value-language additions.

Open decisions: the actual standards-compatible graph extension/wire contract;
complete RFC/BDP comparison and numeric policy; Type extension/reference mapping;
substring extension argument behavior and Unicode version; scope/binding frame
representation; current-hop source spelling; graph witnesses/limits; aggregate
syntax and predicate-language legacy cutover. These must be settled in their authoritative specs,
not treated as implicit choices made by this note.

Primary reference: [RFC 9535 filter scope](https://www.rfc-editor.org/rfc/rfc9535.html#section-2.3.5.1)
and [function extension types](https://www.rfc-editor.org/rfc/rfc9535.html#section-2.4).
