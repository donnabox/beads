---
title: "CLI selection, projection and output"
description: "The canonical contract for generic Bead and Link selectors, field projection, rendering and completion."
---

## Status and scope

This is the October 19, 2026 design specification for the generic Beads CLI.
It consolidates Donna's accepted output decisions and Melinda's query/list
reconciliation. Examples show intended syntax, not currently qualified commands.
The scope is the common selector/output interface, not the complete query language.

The contract applies only to generic command paths that explicitly adopt it.
`bd dep` and Issue/Dependency-specific commands retain their shipping compatibility
contracts. An implementation must not attach these flags indiscriminately to the
root command and thereby change every command's output.

BDP defines Resource and operation payloads. This specification defines their
CLI selection, projection and presentation. Melinda's query design defines
population, predicates, semantic projections, traversal, aggregation and read
consistency. Iris's interchange design defines transferable export packages.
This document does not add a BDP capability, assign a Type identity, or imply that
an arbitrary selected graph is a complete export.

The common behavior is settled. Sections explicitly labeled **candidate detail**
are editorial resolutions for schema/adapter review, not additional decisions
already approved by Donna. In particular, final BDP contract identifiers and
command payload adapters must be pinned before implementation claims conformance.

## Processing model

Process an invocation in this order:

1. Validate command/flag combinations and pointer syntax.
2. Resolve candidate population, predicates, semantic result, order and bounds.
3. Execute against the command's declared read context.
4. Apply the representation mask to the selected result.
5. Render using the selected output format and report completion.

Steps 4–5 cannot change candidate membership, selected graph edges, aggregation,
order, revision selection or the unit being counted. Changing the requested page
mode can change how many selected records are delivered; changing field masks
cannot. Output mode has explicit paging defaults described below.

Resources are complete by default in JSON. Preserve BDP field names, values and
the distinction between absent and null. Do not reshape arbitrary Types into
Issues. Do not chase Links for display. Inline `ownedLinks` belongs in a complete
Resource when the selected BDP contract includes it; it is not an instruction to
fetch more Links or endpoint Resources.

## Selector contract

### Population and identity

The query model admits `--beads`, `--links` and their union. Neither flag means
Beads; both mean the union of Beads and Links. These flags select query candidates,
not fields, and do not make root writes polymorphic. Registration on list wrappers
must follow the shared query/list design rather than be inferred from this page.

Root `show`, `history`, `versions` and `compare` may inspect either Resource kind.
Existing explicit Link identities disambiguate root reads; `bd link` readers
resolve bare IDs as Links. Root `create/update/delete` operate on Beads;
`bd link create/update/delete` operate on Links.

Existing Type and endpoint conveniences such as `--type`, `--source`, `--target`
and `--incident-to` lower to the same semantic request as equivalent queries.
This specification does not redefine their operand grammar or declare all these
flags valid on every command. Source/target filtering does not itself traverse.

### Exact and family Type selection

An exact Type test compares declared Type identity. A family test admits the
selected Type and its transitive refinement/conformance family. Definition revision
selection is separate. Ambiguous friendly names fail rather than choosing a Type.

The shared query design distinguishes `type = T` from `type is T`. Its predicate
grammar and value rules are maintained with the query specification. These tests
do not imply traversal or change the candidate population. The BDP `conformsTo`
wire spelling is unaffected by CLI terminology.

### Link selection for graph operations

| Flag | Meaning |
| --- | --- |
| `--include-link-type TYPE` | Admit Links with exactly this declared Type. |
| `--exclude-link-type TYPE` | Exclude Links with exactly this declared Type. |
| `--include-link-conforms-to TYPE` | Admit Links in this Type's family. |
| `--exclude-link-conforms-to TYPE` | Exclude Links in this Type's family. |

Each flag is repeatable with one Type identity per occurrence. Do not split
commas, which can be part of an identity. Positive selectors form a union;
exclusions win. With no positive selectors, retain the command's documented
selection default. Generic graph commands must not silently choose a Memory or
work-domain family; a domain view may explicitly document its family.

Apply the Link mask before traversing edges. Excluded Links cannot be crossed.
Initial roots remain; a Bead can remain reachable through another admitted path.
Preserve parallel Link identities. These flags restrict the graph operation already
requested; supplying one to a command without graph semantics must not create
an implicit walk. Seed filtering, traversal pruning and post-walk filtering remain
distinct query operations.

```text
bd graph ID --include-link-type TYPE_A --include-link-type TYPE_B \
  --exclude-link-conforms-to TYPE_C --output-format tree
```

Graph selection is identical for table, tree and machine renderings.

## Representation projection

### Flags and defaults

| Flag | Meaning |
| --- | --- |
| `--include POINTER` | Establish an explicit field selection; repeat to union fields. |
| `--omit POINTER` | Subtract optional fields; repeat to union omissions. |
| `--omit envelope` | Emit raw payloads without CLI framing. |

With no projection flags, JSON contains complete Resources and tables use useful
command-specific columns. An explicit include replaces the default optional
field selection; it does not merely add a column. Always retain the result
contract's mandatory context. Omissions win over optional includes; attempts to
omit mandatory context refuse.

The same pointers describe fields in JSON, detail and table output. Do not
introduce an independent `--columns` language. An included object or array remains
that value; the CLI must not invent new flattened columns or summarize a body
into a fabricated display property.

### JSON Pointer rules

Use RFC 6901 string-form JSON Pointers. `~1` represents `/` within a key and `~0`
represents `~`. Shell quoting is separate. No JSONPath, wildcard expansion, implicit
array flattening, dot-path shorthand or URI-fragment pointer syntax is implied.

```text
bd show ID --output-format json --omit /metadata \
  --omit /properties/largeBody
bd list --include /properties/title --include /metadata/team
bd show ID --omit /properties/a~1b
```

In the intended adopted Resource contract, whole optional sections include
`/properties`, `/metadata`, `/revisionProvenance` and `/ownedLinks`. Field names
always come from the adapter's actual BDP payload contract; `/revisionProvenance`
is a target-contract example, not a promise that the currently consumed wire
DTO already has that field (see the adoption dependency below). Nested
selection applies within properties and metadata. Other optional sections are
selectable as complete sections unless their result contract explicitly supports
finer projection. Do not use field masks to remove graph members.

Missing optional paths are no-ops, including when traversing a value of the wrong
shape in a heterogeneous collection. They do not create nulls or missing-value
sentinels. Invalid pointer syntax is an error. Duplicate paths and parent/child
overlap are harmless. Omitting the root would remove required identity and refuses.
An included root selects the complete addressable value.

Array omission evaluates every pointer against the original result, then removes
the selected original elements once, retaining survivor order and array shape.
It does not preserve original cardinality or final index numbers. Never reevaluate
a later pointer against an already compacted array.

**Candidate detail — indexed inclusion:** use the same original-index rule for
includes. Keep selected elements in source order and compact once; retain ancestor
objects and arrays, including singleton arrays. Do not pad holes with invented
nulls. Pointer metadata refers to the original payload. Includes and omissions
are evaluated against that same original value before materializing a projection.

### Projection roots and protected context

The result descriptor declares addressable units before execution. A pointer's
meaning must not depend on which fields happen to be present.

| Result | Addressable unit | Context that field projection cannot erase |
| --- | --- | --- |
| Resource / Resource collection | Each primary Resource. | Identity, kind-defining fields, Type, revision and Link endpoints. |
| Version collection | Each declared version item, using the provider's actual shape. | Identity and exact revision; versions must remain distinct. |
| Search hits / computed rows | Each declared hit or row, using its actual field names. | Row discriminator and identity/reference context required by that result contract. |
| Graph | Each represented Resource; graph membership is outside this mask. | Roots, edges, reference identity and read/completeness context. |
| Comparison | Represented operand Resources where present. | Operand identity/revisions/order and comparison meaning; no rewriting change paths. |
| Mutation | Represented Resource postimages where present. | Applied/no-op/unknown effects, guards/predecessors and affected-source information. |
| Assessment-bearing result | Represented Resources under the assessment contract. | Evaluation identity, full collective exact subject group, claim identities/Types/endpoints, scope, completeness and unresolved context. |

These are adapter obligations, not newly invented BDP payload fields. An adapter
must declare its roots and protected context before admitting projection flags.
Unsupported projection refuses rather than silently changing pointer roots.
A version item may be a reference or summary rather than a full Resource
snapshot. Projection never fetches a full snapshot to fill fields absent from
that item. An embedded Resource, when actually returned, uses recursive Resource
protection; the version item remains its declared result shape.

Protection applies recursively to represented Resources and References. It is
not equivalent to preserving every Type-required property: a large required
body can still be omitted from an explicitly projected Resource. Ordinary objects
that happen to contain an `id` property do not become Resources.
An optional embedded section, such as the entire `ownedLinks` plane, may be
omitted when its parent result contract permits that omission. Recursive
protection applies to values that remain represented; it does not force every
optional embedded Resource to remain in the result. Keeping a projected embedded
Link while removing its identity or endpoints still refuses.

Assessment protection applies to assessment-bearing contracts. Generic Resource
projection must not hard-code a Memory Type name or treat every `subjects` property
as special. An assessment adapter preserves its complete subject group, including
each exact revision, in result data. `--omit envelope` therefore cannot erase the
context and turn a collective claim into a claim about one subject.

An absent target body does not permit fetching it to satisfy an include. A
Reference stays a Reference; a search hit stays a hit. Re-rendering and projection
changes under continuation cannot broaden the read or silently change its context.

### Projection metadata

When explicitly projecting, include a `projection` object in the envelope/header.
Its masks describe the requested original-value paths, not shifted array indexes.
Omit it when no explicit representation projection was requested.

**Candidate schema detail:**

```json
{
  "projection": {
    "include": ["/properties/title", "/metadata/team"],
    "omit": ["/metadata/team/private"]
  }
}
```

Omit an absent mask member. Protected fields retained by the result contract are
implicit, not user-requested includes. `envelope` is a framing control and is not
a Resource pointer in this metadata. Deterministic deduplication may normalize
mask arrays without changing semantics.

Projected payloads are explicitly incomplete representations; they need not
validate as complete stored Resources and must not be treated as complete write
replacements or transfer packages. Raw output loses this projection declaration.
Future protocol projection should use the same semantics, subject to BDP review;
client-side projection alone does not save upstream transfer bytes.

## Formats and human output

```text
--output-format default|detail|table|tree|json|jsonl
```

Omission means `default`, the command's documented presentation. It does not switch
because stdout is redirected or an agent environment variable is set. `show`
defaults to detail; generic collections and generic graph views default to tables.
A domain command may explicitly choose tree. Commands admit only the formats
appropriate to their result contract.

Retain `--json` as shorthand. `--json --output-format json` is equivalent;
`--json --output-format table` refuses. Conflicting repeated scalar flags refuse
rather than selecting the last value. Explicitly repeatable selectors/masks
accumulate. Existing `--format`, `--pretty`, brief and body-only contracts require
an explicit compatibility mapping, not silent reinterpretation.

Detail shows labeled fields/sections without silently truncating selected values.
Tables may visibly abbreviate long cells; full detail remains available through
`show`. Defaults should be useful without requiring the caller to enumerate columns.
Use a Type-declared display property, a conventional title only where defined, or
identity as fallback. Do not infer labels by acquiring or summarizing large bodies.
The actual Type display declaration remains part of the Types contract.

Tree expands each selected Resource node once and shows later occurrences as
identity references. Preserve distinct revisions when the graph distinguishes
them. Mark cycle-closing edges, Link Types and directions, and retain parallel
Links. A tree is a rendering of the selected graph, not permission to discard
edges or disconnected selected components. Terminal color/pager settings do not
alter selection or machine payloads.

## Pagination and continuation

| Flag | Meaning |
| --- | --- |
| `--paged` | Return one result page. |
| `--unpaged` | Follow continuation to exhaustion or the total cap. |
| `--page-size N` | Set the per-fetch page size. |
| `--limit N` | Cap total emitted items; `0` means unlimited. |
| `--continue VALUE` | Resume the protocol-supplied opaque continuation. |

`--paged` and `--unpaged` are mutually exclusive. Generic human collections and
JSON default to one page of 50; JSONL defaults to unpaged with no implicit total
cap. A user-supplied limit always applies. Domain defaults stay documented during
compatibility work. Reject negative limits and nonpositive page sizes.

Page size is not a total limit and may be bounded by the provider's contract.
Do not manufacture exhaustion when a provider returns a short page with a
continuation. Preserve `--all` as the shipping selection control, not pagination.
These flags do not control a terminal pager.

Continuation fixes selection, semantic projection, order and the protocol's read
context. Only representation projection and rendering may change. If the requested
representation needs unavailable fields under that context, refuse rather than
fetching from a new context. Do not invent snapshot consistency or reconstruct
pieces of an opaque continuation.

A consumed page must not lose items because a smaller CLI limit was applied.
An adapter must obtain or retain a continuation at the delivered boundary before
claiming resumability; if it cannot, disclose unavailable continuation rather than
return the next-page token and silently skip undisplayed items. Exact mechanisms
belong to the provider/query contract.

Count and aggregation operate on their defined selected input, not just the page
being printed. A graph or compound result is not arbitrarily split into Resource
rows to make it streamable. Each adapter declares its delivery unit.

## JSON envelope

```json
{
  "schema_version": 2,
  "bdp_spec": "<immutable BDP contract identifier>",
  "result_kind": "resource",
  "outcome": "success",
  "data": {}
}
```

This is a schematic envelope, not a valid Resource example. `schema_version` is
numeric `2`, advancing shipping Beads' numeric `1`. `bdp_spec` identifies the exact
selected BDP contract, never just the executable version. A CLI-derived graph or
computed result must not claim a BDP operation schema that BDP does not define;
the identifier governs the BDP values it contains, while `result_kind` identifies
the CLI result contract.

Resource payloads and BDP operation structures retain their wire meanings.
CLI framing sits outside them. Preserve a collection's native continuation and
item structure rather than flattening it into an unrelated array in ordinary JSON.

**Candidate result-kind registry:** `resource`, `resource_collection`,
`version_collection`, `history`, `comparison`, `graph`, `rows`, `aggregate`,
`mutation`. Each admitted command binds one kind and exact payload schema in its
descriptor. This is not a claim that all these result kinds are already implemented.

Where an outcome applies, use `success`, `failure` or `unknown`. Unknown is an
indeterminate mutation outcome, not null, a query nonmatch or an incomplete read.
Applied versus no-op stays in the underlying mutation result. Do not invent a
new revision just to distinguish output.

### Completion and errors

Ordinary JSON assembles and serializes the requested result before emitting its
success envelope, even when unpaged. A handled retrieval/encoding failure emits
a complete failure envelope. Start with in-memory buffering; disk spooling is
not a required feature or new flag. Output buffering does not make mutations
transactional or undo effects that occurred before a rendering failure.

**Candidate envelope additions:** `error` carries `{ "origin": "bdp" | "cli",
"problem": ... }`; preserve an available BDP problem unchanged. Local problems
have a stable code and message and are explicitly CLI-originated. Failure may
retain an available operation result in `data` when needed to disclose actual
effects; it must not replace an indeterminate write with an apparently retry-safe
error. Exact local error-code registry is an implementation prerequisite.

For collections, an outer `completion` reports `returned`, `exhausted` and
`stop_reason` (`exhausted`, `page`, `limit`, `error`). When a failure prevents
determining exhaustion, encode it as unknown rather than false certainty.
Only report emitted item counts, not a fabricated total matching count.
Protocol continuation remains in its native JSON payload location.
These exact `completion` field names are candidate schema details.

Successful no-match queries have empty results and zero status. Finding cycles
is data; an explicit validation/check command may fail on violations. Preserve
shipping numeric exit codes for Preview 3; use stable structured error codes to
improve machine handling. A later compatibility pass owns numeric normalization.

`--omit envelope` emits `data` alone, including native BDP operation structure.
For a failed Resource read with no data it emits no stdout. Diagnostics go to
stderr and status is nonzero. Raw output does not independently identify schema,
projection or completion. Never inject a rogue error Resource.

## JSONL

JSONL uses UTF-8 JSON values separated by newlines. Framed output has exactly one
header, zero or more items and one terminal. Empty selections still have both
control records. Use `record_type` to distinguish them from Resource kind.

**Candidate serialization of the accepted framing:**

```json
{"record_type":"header","schema_version":2,"bdp_spec":"<immutable contract>","result_kind":"resource_collection"}
{"record_type":"item","data":{"id":"<Resource identity>"}}
{"record_type":"terminal","outcome":"success","completion":{"returned":1,"exhausted":true,"stop_reason":"exhausted"}}
```

The item is schematic, not a complete BDP Resource. A real item contains the
declared delivery unit, projected only as requested. Header metadata uses the
same names as JSON and includes projection when applicable. It never claims an
outcome before execution completes. Terminal errors use the same `error` contract
as ordinary JSON; continuation, when available, is carried losslessly as a
protocol value in a terminal `continuation` member.

For Resource/version collections, emit one declared collection item per record.
For a single Resource, emit one Resource item. Compound graphs and comparisons
remain complete compound items unless an explicit streaming contract is designed;
do not lose relationships or operand context by silently flattening them.

A handled midstream error retains emitted items and writes a failure terminal,
diagnostics on stderr, and nonzero exit. A missing terminal means incomplete or
unverified delivery. A requested page or limit can complete successfully without
exhausting the collection. Failure is never disguised as an ordinary successful
page boundary. A write outcome may remain unknown.

`--omit envelope` removes header, item wrappers and terminal, leaving raw payloads
only. Midstream failure does not add an exceptional data row. The caller must
observe stderr/status; saved raw bytes cannot prove successful completion.

A broken pipe, terminated process or unwritable output can prevent even a failure
terminal from being delivered. Do not claim guaranteed completion delivery.

Generic JSONL is not automatically an export format. Export manifests, History
closure, source-key mapping, digests and transfer completion belong to interchange.
Do not prepend this header to legacy raw exports without explicit versioning.

## Partial results and semantic completeness

Pagination delivers portions of a defined selection. It is not evidence that
unfinished negative predicates, walks or aggregates have been fully evaluated.
Representation projection is another separate axis: a projected Resource can be
part of a successfully completed query.

Ordinary commands fail when they cannot answer their requested bounded question.
JSONL's incremental delivery does not implicitly authorize approximate or
semantically unfinished results. Never report a partial count as an exact total
or an unfinished absence test as proven absence.

Melinda's `--allow-partial` direction is a shared query extension. Its admissible
findings, required completeness metadata, failure/partial outcome and exit status
must be specified before that flag is admitted. This base contract neither
rejects that direction nor assigns it an invented successful status. Until its
adapter is specified, unsupported partial mode refuses explicitly.

## Applicability and compatibility

The generic readers, collectors and writers are the intended adopters:

| Surface | Treatment |
| --- | --- |
| `show` and `link show` | First complete Resource adapters; detail, JSON, projection. |
| `list`, `query`, `link list` | Shared selectors, projection, pagination, JSON/JSONL and table output as the query model lands. |
| `versions`, `history`, `compare` and explicit Link variants | Preserve version/event/comparison contracts; share framing and supported projections. |
| Generic graph view | Shared Link selectors and topology-preserving table/tree/machine output. |
| Generic create/update/delete and Link variants | Shared envelope and meaningful result projection; preserve effect/no-op/unknown semantics. Do not add meaningless paging/traversal flags. |
| Generic search/count/Types inspection when admitted | Reuse the common machinery with their own hit/aggregate/Resource contracts. No declaration that today's domain forms are already generic. |
| Memory facades and export | Later explicit adapters; retain assessment and interchange-specific context. |
| `dep`, Issue/Dependency workflows and operational commands | No retrofit of this flag family or new JSON contract. Preserve supported compatibility behavior. |

Do not silently switch existing scripts to envelope 2. The rollout must map
legacy root `--json`, `--format`, tree/brief/body-only forms and environment
controls before each overlapping generic path becomes canonical. Root names
such as `show` and `list` have shipping users even though they are generic
destinations. Compatibility choice does not belong in the provider or projection
engine. Point users to generic commands in help/documentation without adding
advisory chatter to machine stdout.

## Conformance cases

These are required implementation checks, not tests already executed.

| ID | Obligation |
| --- | --- |
| O01 | Complete Resource JSON uses BDP names and contains no fetched Link/target additions. |
| O02 | Exact versus family Type selection, union includes and winning exclusions agree across renderers. |
| O03 | Explicit field includes replace defaults; protected fields remain and prohibited omissions refuse. |
| O04 | Missing paths, escaped keys, null, heterogeneous shapes and overlapping masks behave deterministically. |
| O05 | Array omissions use original indexes; indexed includes retain containers without holes or singleton promotion. |
| O06 | Embedded Resource/Reference protection is recursive; Type-required large bodies can still be projected out. |
| O07 | Assessment collective subjects and exact revisions survive projection and envelope removal. |
| O08 | Query semantic projection and representation masks remain distinct under continuation. |
| O09 | Page-size, total limit, short provider pages and continuation neither skip nor duplicate undisplayed items. |
| O10 | JSON late failure emits no partial success document; JSONL late failure emits an error terminal when writable. |
| O11 | Empty JSONL, page completion, limit completion, exhaustion, unknown writes and absent terminals remain distinguishable. |
| O12 | Unwrapped JSONL contains only data; process failure never masquerades as an extra Resource. |
| O13 | Conflicting scalar/format flags fail before effects; equivalent values and repeatable masks compose. |
| O14 | Tree cycles, parallel Links, shared descendants, disconnected components and distinct revisions retain graph meaning. |
| O15 | Legacy `dep` and domain-command snapshots remain unchanged; generic command JSON compatibility is explicitly exercised. |
| O16 | Every admitted adapter pins its BDP contract, result schema, projection roots, mandatory context and delivery unit. |

## Beads and BDP implementation responsibilities

The specification lives in the Beads repository because these flags and the
envelope are a product CLI contract. Shared protocol semantics belong in BDP.
Do not implement a new portable read capability only in a CLI renderer or claim
BDP conformance from a local convenience implementation.

| Capability | Beads responsibility | BDP responsibility / required change |
| --- | --- | --- |
| Complete Resource/Reference and operation payloads | Adapt the selected provider results without Issue reshaping or field renaming. Pin the exact contract and qualify local/shared-server paths separately. | Existing selected schemas are authoritative. Any missing operation payload/effect information needs a shared-contract change before the CLI invents it. A new envelope alone requires no BDP amendment. |
| JSON Pointer include/omit | Parse and validate flags, apply representation masks, protect descriptor context, record projection and render. Initial client-side projection may fetch the full authorized Resource. | Source-side field projection is future protocol work: define request encoding, protected context, original-index array rules, projected-response schema and continuation interaction. It is not a prerequisite for the first client-rendered show slice. |
| Type, population and endpoint selection | Lower convenience flags and predicates into a normalized request; preserve exact versus family identity and explicit kind. | Reuse supported exact-Type/family/endpoint operations at the adopted pin. Do not rename `conformsTo` merely to match CLI terms. Unsupported shared operators require a BDP change or explicit refusal. |
| New property/text predicates and query semantics | Own language syntax, checking and lowering; reuse shared semantic operations instead of creating command-local filter dialects. | Review portable operator/value semantics and provider capabilities with the BDP owner. New substring/text operations are not implied by existing Selector support. Amend schema/spec/conformance and implement the shared evaluator/provider surface where the chosen execution path requires them. |
| Correlated graph queries and traversal | Own CLI graph requests, result adapters and rendering; do not perform hidden traversal in a formatter. | Correlation, joins and portable graph traversal are not automatically existing Selector features. The shared operation and its authorization/completeness/read-context guarantees require BDP design and implementation if exposed through BDP. Local bounded composition must disclose its weaker context and cannot claim a coherent snapshot. |
| Paging, continuation and total caps | Follow supplied continuation, count emitted units and report request completion separately from exhaustion. Avoid skipping undisplayed items at cap boundaries. | The provider defines continuation encoding and read-context guarantees. A resumable delivered boundary, snapshot stability or projection-changing continuation cannot be fabricated by the CLI; any gap must be resolved in the shared contract/provider or reported unsupported. |
| Exact counts and aggregates | Use the same selection as listing and preserve the counting unit. Never count one displayed page as the total. | Reuse an existing qualified exact operation where available. New portable aggregation/count/read-consistency guarantees need a shared operation and evidence; do not infer them from pagination. |
| Assessment context and Type display declarations | Result descriptors protect collective assessment meaning; render declared labels without fetching hidden data. | Type/Memory contracts define subjects, References, claims and any display declaration. CLI output cannot invent a metatype property or new Memory semantics. Missing shared declarations go through their BDP owners. |
| Errors, revision effects and unknown writes | Keep diagnostics separate, preserve BDP problems/effects, and identify local failures. Do not infer rollback or safe retry from output failure. | Existing Write/History/ownership contracts remain authoritative. Amend BDP only for missing shared outcome semantics, not to add CLI exit codes or `record_type`. |
| Envelope, JSONL, table/detail/tree, scalar-flag conflicts | Entirely Beads: schema 2, formats, process status and finite-stream framing. | No BDP wire envelope change is required. BDP Resource schemas must not acquire CLI headers, table fields or terminal records. |
| Output-schema discovery | Future Beads `--json-schema` introspection identifies command envelopes and their payload contracts. | Reuse published BDP schema identifiers. New protocol discovery is separate work if a concrete need is selected. |

For a BDP amendment, update the normative contract, schemas, shared implementation
and conformance evidence together, then pin downstream adoption in Beads.
Trish owns shared-contract reconciliation; the existing BDP implementation owners
and Janet coordinate realization and downstream qualification. This table is a
dependency disposition, not a claim that every possible extension is required
before the first CLI read adapter or already implemented.

The independently approved empty-update-array change supports symmetric mutation
editing and is tracked separately. It is not an output/projection prerequisite
and must not be bundled into this spec's implementation without its existing
review and integration gates.

### Current wire-adoption dependency

Trish's October 19 review identified an existing adoption gap: the inspected
Beads wire schema records BDP commit
`de99030d13a57e77f4f2660b91ea3c39b82823cb`, and its current DTOs still expose
`attribution` and `changeContext`. The intended `/revisionProvenance` examples
above must not be implemented by relabeling those older fields. This is the
existing wave 3 schema/DTO/projection/client adoption dependency, not a new
protocol change or permission to move a pin independently.

Each adapter must identify the contract it actually emits. An integration-head
commit in `bdp_spec` is inaccurate when the payload still uses an older schema.
Janet owns the coordinated adoption and qualification; the first output slice
must freeze its actual payload contract after that handoff.

## Related sources and remaining dependencies

The [Preview 3 design](/reference/graph-cli-specification-draft) remains the
implementation-baseline ledger. This page is the common output contract.
The local query/list reconciliation and its acceptance cases are in
`engdocs/design/memory-query-scope.md` in Melinda's design worktree; publication of
that draft must provide the final durable query-spec link before this design ships.

Implementation prerequisites are the selected BDP contract, result-specific
adapters, applicable Type display declarations and explicit legacy-mode mapping.
The query extension contract for opt-in partial answers is required only for
adapters admitting that capability; it does not block strict Resource reads or
ordinary collections. None of these prerequisites is satisfied merely by adding
flag declarations or changing a JSON printer.

`--json-schema` discovery is Memory V1 design scope, outside Preview 3. It must
distinguish CLI result schemas from Type property schemas and account for
projection, raw output and JSONL.

Shipping comparison: [Beads schema version 1](https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/output.go#L13);
[Gas City buffering and failure handling](https://github.com/gastownhall/gascity/blob/f66474617d2bb7fbe51a3d746da5f3ceab2b8517/cmd/gc/main.go#L210).
