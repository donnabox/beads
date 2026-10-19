# Memory query and CLI scope proposal

**Status: informative proposal for Donna's review.** Start with read-only views that explain a subject's declared constraints, necessary obligations, and recorded assessment claims. These views should share selection and traversal behavior with future `bd query` work. This proposal selects neither a query language nor a release allocation.

## Ownership and dependencies

The [active BDP/Beads policy](https://github.com/gastownhall/beads/issues/7656#issuecomment-6253701616) establishes the split below. This proposal applies that policy; it is not another policy or status tracker.

| Concern | Normative home |
| --- | --- |
| Constraint, Subject To, Requires/Defines, and claim meanings | BDP's Memory typedef Specification, currently [010be646](https://github.com/gastownhall/bdp/blob/010be6463c54041dc9ec266c385788ec1a3fd88b/specs/types/memory/memory.md). Reference these rules rather than redefine them in CLI code or prose. |
| Portable resource selection, Type-family matching, read consistency, and any future portable traversal mechanism | The applicable BDP shared contracts. A missing protocol mechanism needs a separate amendment; a Type definition cannot silently add one. |
| Commands, selectors, flags, defaults, output, diagnostics, compatibility, and exit behavior | Beads. The eventual Memory facade belongs at `docs/reference/bdp/cli/types/memory/memory.md` and Requires the whole Memory typedef Specification. This does not implicitly require the Types profile or a new BDP protocol. |
| Evidence acceptance and reconciliation for a release, compliance, or closing gate | The consuming application, under an explicit policy. A query does not manufacture a verified verdict. |

Trish owns Memory-Type implementation and Type/metatype design; Melinda owns the Memory Specification and this design coordination; Janet owns integration and qualification. The facade does not require a separate mutation model or a new command for generic CRUD.

## Four questions to make useful

Names below are sketches already raised by the Memory draft, not adopted command grammar. `SUBJECT` and `CONSTRAINT` are illustrative selectors.

| Candidate view | Question | Result to preserve |
| --- | --- | --- |
| `bd memory constraints SUBJECT` | Which declared Constraints govern this subject, and why? | Subject To roots, their Requires-family closure, and explanation paths. State whether implicit applicability was examined. |
| `bd memory requires CONSTRAINT` | What is necessary for satisfying this Constraint? | Requires closure, with Defines edges identified as the defining structure; shared targets and cycle markers remain visible. |
| `bd memory coverage SUBJECT` | What claims exist about these obligations? | Recorded Conformance-family claims bound to the selected subject scope, actual Link Types, evaluation identity where present, evidence/provenance, assessed revisions/context, conflicts, and obligations with no visible relevant assessment. |
| `bd memory explain SUBJECT CONSTRAINT` | Why is this an obligation, and what supports the assessment? | An applicability witness, dependency edges, relevant claims, and reasons a stronger conclusion or composition is unavailable. |

The same query foundation should support generic resource/Link selection, exact Type versus refinement-family matching, and reusable bounded walks. A Memory view applies the typedef's domain rules to those primitives. Extending the existing Issue-filter grammar is a separate decision; these examples require no new `bd query` syntax yet.

## One example and its expected answers

Use these labels as an illustrative graph, not assigned Bead identities:

```text
Release --subject-to--> Export Specification
Export Specification --defines--> Export selected rows
Export Specification --requires--> Protect customer data
Protect customer data --defines--> Redact logs
Export Specification --subject-to--> Review Specification documents
```

`constraints Release` reports the Specification, row Requirement, data Policy, and redaction subpolicy. It does not report the document-review Policy: Subject To is nontransitive. `requires Export Specification` shows its necessary dependencies and distinguishes Defines from ordinary Requires.

Suppose Release has a Satisfies claim for the row Requirement and an Addresses claim for the Specification. These direct subject-to-Constraint claims illustrate the published 010be646 baseline. The row has a complete-satisfaction **claim**; the Specification has a contribution claim; the Policy and subpolicy have no visible relevant assessment. None of this establishes verified compliance or changes Issue readiness. A Violates claim concerning the same row, subject, revisions, and context must coexist with the Satisfies claim as a contradiction. A claim concerning another revision must be shown separately, not merged into that conflict.

## Smallest proposed slice

Recommend **declared-root inspection, Requires closure, direct claims, and explanations** first. Keep the selected question explicit: this slice cannot claim to discover all applicability while predicate evaluation or governing-root authority is unavailable. An unsupported request must be reported rather than silently answered with a narrower query.

Do not initially evaluate prose or applicability predicates, run tests, author claims, reconcile evidence, or derive indirect Satisfies claims without established revision/context compatibility. The portable Type rules remain intact; unavailable capabilities are explicit. Donna has accepted the direction that a Bead representing a conformance evaluation owns its claims, and that it can assess multiple subjects or one compound subject. **Assessment** is a working name, not an existing or newly selected Type name. The published 010be646 baseline still uses direct subject → Constraint claims; its coordinated successor needs the evaluation/subject representation defined. Query results must distinguish the evaluation owning a claim from the subjects being assessed and must preserve their selected revisions. A recorded claim is “direct” here because it was authored, rather than derived through a Satisfies chain; that does not require the assessed subject to be the Link's source. Do not silently distribute a group claim to its individual members. A claim about several subjects applies to them together; it does not assert the same conclusion independently for each member. Individual conformance needs an explicit claim about that individual subject. Per-claim scope representation, Type name/reuse, subject binding, and lifecycle remain to settle.

Retain the four existing memory entry points—`remember`, `recall`, `memories`, and `forget`—as compatibility inputs to the eventual facade design. Their canonical-group mapping and precise behavior still need specification. Type-specific creation shortcuts and tuned create/update/show/delete operations are outside this first slice. Packs and a broad query-language rewrite are separate work.

## Truthful results and read context

A result must distinguish **recorded findings**, **assessment compatibility**, and **query completeness**. None is a stored Memory status or a new claim-kind enum.

- Preserve selected subject/Constraint identities and revisions, evaluation identity and full assessed subject scope where applicable, actual Link identities/Types, evidence, provenance, and supporting edges. Missing evidence is not verification; missing claims are not violations.
- Bound depth, visited resources, claims, and explanation size. Deduplicate repeated visits; represent shared nodes and cycles. A finite witness/edge graph avoids pretending to enumerate infinitely many walks. Distinguish an incomplete obligation set from abbreviated explanations.
- Define completeness relative to the selected question, authorized view, and read context. A visible result cannot certify the absence of hidden obligations or assessments. Do not reveal hidden endpoints, counts, or reasons.
- Report exhausted budgets, unsupported capabilities, unresolved revision/context compatibility, and unestablished read consistency. Never present a bounded or inconsistent read as an exhaustive “nothing unmet” answer.

Recommend a structured envelope for the **new** Memory views so completion and limits travel with the rows. Keep query observations separate from any consumer's pass/fail decision. Do not silently replace existing `bd query --json` arrays with a different shape. Defaults, exact fields, pagination, and exit behavior remain facade decisions.

## Existing mechanisms and the actual gap

At the [selected BDP shared source](https://github.com/gastownhall/bdp/blob/b12665d64c2aaf6172bd1cdab409cef31ef28f44/specs/profiles/core/core.md#selection), Selector filters individual complete Resources; it expressly excludes traversal, joins, projection, and aggregation. Collection reads already expose exact Type, family, and endpoint filters plus pagination. Their historical `conformsTo` wire spelling is not renamed by Memory's Refines terminology.

A client can compose those reads into an exploratory walk. However, a cursor stabilizes one collection query, not every independent request in a graph walk. Exhausting each page proves neither a common graph snapshot nor global closure. A stable multi-query read context or a portable server traversal would require an explicit existing guarantee or separate shared design. The initial facade should expose unestablished consistency rather than silently acquire a Snapshot-profile dependency or promise an authoritative gate result.

The [inspected Beads Preview 3 source](https://github.com/donnabox/beads/blob/5b8cdefe055b8f08cf9eb385e946396c1b5799d7/cmd/bd/graph_preview_discovery.go#L500) implements an Issue-oriented graph query route. Its [query frontend](https://github.com/donnabox/beads/blob/5b8cdefe055b8f08cf9eb385e946396c1b5799d7/cmd/bd/query.go) and generated help are behavior/reference inputs, not evidence that the new Memory traversal exists. Native CLI behavior and BDP HTTP support require separate capability claims.

## Decisions for the first review

1. **Initial scope:** accept the declared-root/direct-claim slice above, with unavailable predicate and indirect-assessment behavior explicit? Recommended: yes; it yields useful inspectors without inventing verification.
2. **Consistency:** allow clearly marked exploratory results when no common read context is available, or require a coherent read context before offering these views? Recommended: permit exploratory inspection, with consistency and bounds attached; it cannot supply a definitive gate verdict.
### Accepted multi-subject assessment direction

An evaluation-owned claim about a Web app and an API concerns the combination: they satisfy the Requirement together. It does not establish that either component independently satisfies the whole Requirement. A query for the Web app can show its participation in that collective assessment, but must preserve the full assessed scope rather than label the Web app individually satisfied. Conflicting-claim and composition reasoning must respect the assessed group and its selected revisions; overlapping groups are not automatically equivalent scope.

The concrete representation, Type name/reuse, subject binding, and cleanup contract still need design. This proposal selects neither a join nor a generic third-party owner field.

Before implementation, turn the example into acceptance cases covering nontransitive Subject To, an installed Link specialization included through Requires refinement, shared children/cycles, conflicting versus different-context claims, an incomplete walk, and an unestablished multi-request read context. Use an exact contract source for each case. These are design examples, not passing tests or a Preview 3 qualification claim.

Existing contributor work remains relevant input, not an adopted architecture: [GraphQL query, Peter Chanthamynavong](https://github.com/gastownhall/beads/pull/6780), [parent/mol_type predicates](https://github.com/gastownhall/beads/pull/7101), [no_history filtering](https://github.com/gastownhall/beads/pull/7405), [UTF-8 lexer handling](https://github.com/gastownhall/beads/pull/7544), and [truncation signals](https://github.com/gastownhall/beads/pull/5974). Review and preserve applicable work when a concrete implementation is commissioned.
