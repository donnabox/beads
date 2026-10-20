# Query language review and publication handoff

October 19–20, 2026. Documentation-only design checkpoint; not runtime acceptance.

## Review evidence

Native Codex and actual Claude independently reviewed the same four-file bundle.
Claude ran through `claude -p`, completed with exit **0**, and returned findings.
The native reviewer returned three findings; Claude returned 26. No model identity
is inferred from CLI defaults. **Gemini was reported down and was not run**, so
this is reduced two-provider council coverage. Both reviews were read-only:
no product tests, builds, generation or SQL execution.

The reviewed source snapshot preceded corrections. Its SHA256 identities are:

| Reviewed input | SHA256 |
| --- | --- |
| `docs/reference/bdp/cli/query-language.md` | `dfcea2354199f9d36793a4295d6446b5eb9d2bae588877db03102b39662d0ac5` |
| `engdocs/design/query-language-cases.json` | `64db3d9379e57b7e77ed55d2cdb419dd7740f8a3751b8bd9f03c72dcb8f1bcdc` |
| `engdocs/design/query-language-wrapper-map.json` | `6d0ed3b4af91972cdf1c5dbe7186a2afdbd7acdd607c22a5f0d601e30d329a7b` |
| `engdocs/design/query-language-lowering.md` | `d8218ff4b610c6ea5785d4fd03ae419cf1629045b8e05e0ef48c752737543d42` |

The raw Claude stdout is preserved in the local review handoff, SHA256
`4a4c40571be0d75ee8cf5f6906918fc53d31683b82967c8420a0d27ab882044c`.
This note records every actionable finding and disposition without publishing
private coordination notes. Corpus expectations were checked by reviewers; that
is source review, not executed query tests.

## High-confidence agreement

Both reviewers identified missing ordering/continuation acceptance coverage.
The corrected corpus now includes identity/revision ordering, explicit-sort
identity ties, pages covering the complete selection once, mask-only changes,
semantic-request mismatch and lost read-context refusals. These remain authored
expectations until the actual implementation executes them.

## Native Codex findings

| Finding | Severity | Disposition |
| --- | --- | --- |
| N1: Link-edition identity was conflated with source-edition adjacency membership. | High | Corrected: separate link_edition and adjacency relations; add two source editions sharing one unchanged Link. |
| N2: A witness and a dynamic evaluation error inside unordered EXISTS could produce order-dependent outcomes. | Medium | Added an explicit candidate strict match-error policy and paired witness-first/error-first cases for EXISTS and NOT EXISTS. The October 20 pass uses incompatible ordering instead of conversion. Product acceptance remains open; no optimization may silently choose semantics. |
| N3: Ordering/continuation could not be falsified by the corpus. | Medium | Corrected as above; independently corroborated by Claude finding15. |

## Claude findings and dispositions

| Finding | Disposition | Change or remaining boundary |
| --- | --- | --- |
| C1 | Partly corrected; path retained | Removed personal coordination names from the language page. It remains an explicitly provisional draft at the path reserved by integration. The common-contract fork is public (verified); immutable commit URLs remain useful before in-repo reconciliation. No site publication/merge is claimed. |
| C2 | Corrected | Declared selected_requires_family as a provider-prepared input, explained closure derivation, and removed the execution/portability claim. |
| C3 | Corrected | Added an isolated pin-vs-floating context: two reachable J editions, Reference inspection and distinct-neighbor count = 2. |
| C4 | Corrected as candidate detail | Added Kelvin sign, locale-independent I, full-fold expansion and no-normalization cases. Full C+F folding is explicitly proposed; the Unicode version remains open. |
| C5 | Corrected | Moved EXISTS and quantifiers into primary expressions under unary negation; candidate grammar can represent not exists. |
| C6 | Corrected | Added non-disclosure/authorized-Reference rules and cases for denied versus absent editions, visible-only matching and unavailable required target bodies. No new disclosure permission is granted. |
| C7 | Corrected design obligation; execution pending | Added a plan-work-set LEFT JOIN validation sketch plus explicit adjacency-coverage and budget-failure obligations before publishing answers. SQL joins are not a completeness proof. |
| C8 | Corrected | Logical type aliases now specify PostgreSQL JSONB/BIGINT versus SQLite JSON text/INTEGER. Physical presence uses jsonb_exists. |
| C9 | Corrected explicit adaptation | Optional dynamic reads have nullable JSON typing. Known non-null coalescing and dynamic admissibility remain candidate checker rules. Conversion functions were subsequently deferred in the October 20 markup pass. |
| C10 | Corrected | Contains joins the relational/membership tier; insensitive is a modifier token. |
| C11 | Corrected as candidate policy | Null/absent ordering and contains error behavior is explicit; primitive guard and missing/present comparison cases demonstrate short-circuit guards. |
| C12 | Corrected | Path SELECTs are labeled internal surrogate state, with a mapping relation/join preserving ordinal Link identity+revision before public output. |
| C13 | Corrected notation coverage | Added a case-notation JSON Schema, diagnostic enumeration and comparison rules for rows, arrays, sets, pages and collective context. This is not a public query IR. Full JSON Schema validation remains unrun because the validator library is unavailable; structural checks ran. |
| C14 | Corrected | Added a disposition legend, current-rollout coverage obligation, unproven mapping status and illustrative_case_ids. Excluded rows do not claim command equivalence or approved language exceptions. |
| C15 | Corrected; corroborated by native Codex | Added ordered Resource ties, explicit-sort ties, complete page concatenation, mask-only continuation changes, semantic mismatch and context-loss refusals. |
| C16 | Corrected owner/root | Assessment cases are marked common-output-contract requirements with the addressable subject_group root. Their pointer is /subject_group/1; they authorize no facade implementation. |
| C17 | Corrected as candidate policy | Added adjacent large integer comparison and unsupported-precision refusal. Decimal/domain selection remains provisional. |
| C18 | Original limitation resolved | The query page now has a navigation entry. The documentation-link suite and freshness checks pass; full live site preview remains unrun. Author owns the change under MEM-POL-003. |
| C19 | Corrected sketches; execution pending | Separated seeded lookup from head-only population, typed recursive seed keys, aligned logical integer types, supplied provider-specific binary collations and a single recursive family term. |
| C20 | Corrected input boundary | Added declared Reference identity/pin records, catalog names, known Types and nullable nonfetchable adjacency targets; explained inspection without target body lookup. |
| C21 | Mostly corrected; explicit deferrals | Added exact Link Type, Reference inspection and code-point length cases. string(value) signatures and complete parser/result forms remain named review gates, not implied support. |
| C22 | Corrected | Pinned-history and current-outgoing cases now explain the selected-edition invariant they protect. |
| C23 | Corrected | Removed assessment terminology from coalescing; explicit unordered value-set comparison rules preserve ordered array values. |
| C24 | Corrected | Synthetic Evaluation subjects now carry exact uri/revision objects; collective report requirements remain externally owned and do not validate a Memory catalog. |
| C25 | Clarified; historical rollback declined | Reconciliation uses actual common-contract head39a5d7297; issue7718 preserves its explicitly earlier aeb39e764 snapshot. Historical scope notes are not overwritten or represented as the latest source. |
| C26 | Corrected presentation; complete grammar open | Pipeline illustration is labeled partial, omitted stages named, and predicate tiers clarified. Full lexer/pipeline productions remain gate 1. |

## Remaining decisions and disagreements

The review did not ratify proposed grammar or move the release boundary. Keep
these choices visible for operator and language review:

1. Complete syntax/checker rules: dynamic JSON nullability/access and comparison
   admission; unordered-match error policy versus
   ordered quantifier short-circuiting. `??` remains null-only, not error recovery.
2. Numeric domain and Unicode version. The proposed no-rounding decimal model
   and full default Unicode folding require explicit acceptance or a narrower
   portable subset; silent host-dependent approximation is unacceptable.
3. Graph witness/post-filter composition, row/path ordering and resource budgets.
   Reachability, paths and graph membership remain different operations.
4. Partial-answer status, compatibility cutover and derived-result schema versions
   stay with their existing owners. This draft enables none of them implicitly.

Claude recommended moving the page to engdocs and citing the older issue pin.
We retained the explicitly reserved documentation review path and actual latest
common-contract reconciliation pin, removed internal actor wording, and disclosed
the earlier mirror snapshot. The public fork and immutable source references were
verified. This is a draft review publication, not release/site qualification.

A bounded native Codex follow-up inspected the corrected adjacency, match-error
gate, ordering/continuations, pin/floating context, authorization cases and
notation/schema changes. It reported no remaining actionable issues in that
follow-up. Claude reviewed the initial snapshot, not the corrected revision.
No runtime, SQL or full JSON Schema validation occurred in either pass.

## Actual checks and limits

Static JSON parsing, unique case identifiers, fixture endpoint/Reference consistency,
wrapper source/case references, disposition enumeration and Markdown fence/whitespace
checks were performed. The schema's operation/key/diagnostic enumerations were
checked structurally. A full JSON Schema validator was unavailable in both local
Python runtimes; no full schema-validation result is claimed.

The revised artifacts contain **129 authored cases** and **31 wrapper families**.
No query grammar/parser, SQL sketch, product CLI, provider, Go test, formatter,
Bazel build, generation, installed command, docs site or navigation gate ran.
The previous small isolated SQLite experiment is separate historical evidence;
it does not qualify this revised draft or its 129 cases.

## Integration handoff

Original review base: `5b8cdefe055b8f08cf9eb385e946396c1b5799d7`.
The original signed author commits remain preserved. A private-metadata recovery
clone preserves them and prepares a documentation-only successor from fork main;
no contaminated checkout or remote fixture branch was reset or force-pushed.
The new query page is registered in Reference navigation.

Current execution follows [MEM-POL-003](https://github.com/gastownhall/beads/issues/7656#mem-pol-003-independent-ownership-and-green-ci-code-flow):
the author owns formatting/builds/tests/review and issue/PR publication. Earlier
notes reserving all runs, navigation and publication to integration are superseded.
Actual failed checks and repository protections remain applicable. The initial
publication attempt failed and its Git metadata changed during hook-launched
fixtures; its original receipts are preserved rather than relabeled as passes.
Recovery uses an isolated clone, with safe hook/test routing tracked separately.

The Dolt lowering and immutable primary-source notes are a later documentation
supplement, not part of the original four-file council snapshot. No Dolt SQL
execution or product query qualification is implied. Current documentation checks
and exact source are recorded in the PR, separately from original review evidence.
No main merge, release, pin change or runtime query implementation is authorized
by this documentation publication.

The common output contract remains [PR122](https://github.com/donnabox/beads/pull/122)
and [issue7718](https://github.com/gastownhall/beads/issues/7718).
Before a query implementation slice starts, reserve its writer and files, freeze
its subset/capabilities/compatibility contract, and execute the relevant cases
through its parser/evaluator and admitted providers. This review supplies no
unreserved runtime implementation assignment.

## Contributor protection preflight

The topic search found [Peterkc's upstream GraphQL PR6780](https://github.com/gastownhall/beads/pull/6780)
at `f937eeae812ff50407d422cebc7361435b60e554`. Its description, discussion and
relevant command/API/schema/resolver diff were inspected before this publication.
Read-role reuse, explicit optional acquisition and budgets are relevant input.
It is an Issue-specific proposed facade with separate GraphQL framing; this
query design publication copies no implementation, replaces no contributor PR
and changes none of its tests. Merge/close preflight remains blocked for that
separate PR; no such action is attempted here.

Editable artifacts: [language](../../docs/reference/bdp/cli/query-language.md),
[cases](query-language-cases.json), [case notation schema](query-language-cases.schema.json),
[wrapper map](query-language-wrapper-map.json), and [lowering](query-language-lowering.md).

## October 20 operator markup pass

The editable review copy's seven `///` comments were processed. The specification
now defines value/access rules directly, lists Dolt first, describes the ordinary
current-Scope read without inventing a historical CLI flag, and defers all
conversion functions. Existing error-order cases use incompatible dynamic values
instead of conversion. Three conversion examples are rejection cases; the corpus
still has 129 authored, unexecuted scenarios. Historical review rows describe the
original snapshot; they are not new validation of this pass.

Predicate-foundation review remains open. Current BDP main at
`3ab9cdf5d45a2ea49a84ab191eac77c9aeaa19d0` uses bounded RFC 9535 JSONPath strings
for reads and set mutation. Missing/null, failed ordering/access, negative indices
and numeric admission need explicit reconciliation before calling the candidate
a compatible extension. The current CLI parser has comparison/AND/OR/NOT nodes;
it is a different, Issue-oriented implementation with field-specific behavior.
No parser, runtime, SQL engine, publication issue or implementation contract was
changed by this markup pass.

## Conditional JSONPath missing-field ruling

Donna accepted JSONPath missing-field behavior if JSONPath is used. The current
spec's access section records that ruling; the previous absent/null coalescing
and required/safe-access proposals are suspended. The case corpus explicitly
lists 16 suspended historical candidate cases; their expectations must not be
reported as current acceptance. Graph function-shaped special forms versus sugar
remain a proposal, not an accepted spelling or an implemented macro system.

## Accepted BDP/Beads surface split

Donna accepted two surfaces over one semantic model on October 20: portable BDP
JSONPath plus graph extensions, and human-friendly `bd` syntax with exact
translation. BDP specification placement is coordinated with Trish; Core selector
semantics, profile activation and runtime/spec inputs do not move by publishing
a draft. The current exercises establish true RFC extension boundaries and CLI
sugaring, including paired semantic examples and explicit unresolved choices.
The conditional missing-field ruling now applies to the selected foundation.

## Two-surface exercise deliverables

The new Beads facade is
`docs/reference/bdp/cli/profiles/query/query.md`; the previous combined candidate
is retained as authoring history and points to it. The proposed portable source
is [BDP Query at 3e470bb](https://github.com/gastownhall/bdp/blob/3e470bb35c20e075aac55afb8821f48b82c595d6/specs/profiles/query/query.md).
Its isolated library experiment is executed mechanism evidence, separate from
the 129 authored historical/current scenarios here (16 explicitly suspended).
There is no CLI normalizer or complete paired-wire equivalence result yet.

Independent cross-review corrected the graph function's result domain, a repeated
edition fixture, the experiment's failure-policy caveat, simple-path repeat unit
and the Type-Reference projection caveat. The facade's direct Resource/alias
examples and the portable draft's edition-wrapper examples are explicitly
semantic pairings while input/capture schemas remain unresolved. Full graph
composition over JSONPath is an author recommendation for Donna, not an adopted
protocol amendment. Existing Core selectors, pins and output framing remain
unchanged.
