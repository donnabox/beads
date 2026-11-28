# Query and filtering: Preview 3 and Preview 4 decisions

Recorded October 22, 2026 by Melinda from Donna's interactive design discussion.
This is a product decision and disposition record, not a new query-language
specification, implementation receipt, BDP profile claim or merge authorization.

## Relationship to the original query assessment

This is the release-scoped follow-up to the
[query operation inventory](query-operation-inventory.md) and
[delivery proposal](selection-output-query-delivery-plan.md). Their earlier
recommendation that every read become a wrapper around a general `bd query`
is historical design exploration, not the current Memory release mandate.

The October 20 backend experiment established that canonical graph storage could
support the required Memory operations, but widening the Issue evaluator was not
enough: scoped coherent acquisition was required. Memory operations now have a
separate bounded reader over the existing graph store; they do not invoke the
legacy Issue expression evaluator. The backend experiment is preserved at:

`/Users/dbox/.local/state/melinda/memv1-backend-scenarios-20261020/README.md`

Its assessment examples and proposed release work are historical. They do not
reintroduce later-deferred capabilities into Preview 3. The resulting runtime
candidate and source-bound qualification are recorded in
[PR217](https://github.com/donnabox/beads/pull/217) and
[implementation ledger 7759](https://github.com/gastownhall/beads/issues/7759).
Neither that historical experiment nor this note establishes current integration
or installed-release qualification.

## Accepted release boundaries

| Surface | What it does / inspected boundary | Decision or disposition |
| --- | --- | --- |
| `bd query` | Existing expressions over Issue candidates; generic Memory operations do not use this evaluator. | Preserve compatibility. No requirement to expand beyond Issues for MemV1; Donna is comfortable leaving it Issue-only indefinitely. |
| `bd sql` | Executes backend SQL. | Preserve existing behavior. No generic graph abstraction requirement. |
| BDP collection `selector` | Bounded JSONPath predicate on generic Bead/Link HTTP collections served by graph-mode `bd serve` with shared-server Dolt. Generic CLI list paths reuse collection selection without exposing arbitrary expressions. | Leave existing preview capability alone. No selector expansion or conformant BDP profile is a MemV1 gate. Broader BDP/BTS language/IR assessment remains on hold. |
| Generic `bd list` | Current Bead collection with Type selection and output controls. Existing routes have different population/default/JSON behavior; the new default below is a deliberate change, not already installed behavior. | Preview 3 needs no new property/metadata/text predicates. Adopt shared Type-filter direction with count, preserving Issue compatibility. |
| `bd count` | Existing graph-mode count is Issue-only. | Count must parrot list selection. Type-only generic counting is the agreed design direction; Cleo and Sybil support a bounded Preview 3 increment, subject to integration and qualification. Count the complete selection before paging/display limits; fail incomplete acquisition. |
| `bd search` | Searches Issue title OR ID, with special handling for ID-like input/external reference, existing text matching and all-status default. Structured filters further narrow the result. | Preserve Issue-only behavior for Preview 3. No generic search implementation now. A future list search shorthand is a possible factoring, not an accepted new flag. |
| `bd ready` | Issue/Dependency readiness plus existing Issue filters and optional atomic claim. | Semantically entirely within the Issue/Dependency graph. Generalization would be wrong; preserve working behavior. |
| Regex/glob filters | Existing domain filters, principally labels. Graph ready admits them; graph list does not currently admit every legacy flag. | No generic pattern filtering for MemV1. Preserve existing semantics; missing Issue compatibility is a separate concrete gap. |
| Formula conditions | Workflow recipe conditions and work progression/construction, not generic Memory selection. Molecules are arrangements of work Issues and Dependencies; wisps are ephemeral work. | Preserve the workflow-domain boundary. Donna agrees with that principle. Wisp exclusion from generic graph mode is explicitly deferred for later discussion, not accepted as permanent policy. |
| Output templates / paths | Legacy Go templates and generic JSON Pointer representation masks. | Output is distinct from filtering. Cleo owns the output contract. Detailed disposition of this inventory row remains unreviewed here. |
| Batch grammar | Existing narrow line-oriented write grammar; Iris has the newer batch design. | Preserve current scope pending separate review. No new generic batch or Memory operation is authorized by this inventory discussion. |

A graph-store route and generic-Type semantics are different capabilities.
Existing Issue adapters can work on the graph store without becoming arbitrary
Bead operations. `--graph-mode dependency` behavior remains a compatibility
boundary, including its existing quirks.

## Accepted Type-filter interface for list and count

| Invocation | Selected population |
| --- | --- |
| No Type selector | Issue and its transitive refinements, effectively `--type issue`. |
| `--type X` | X and all transitive refinements. |
| `--exact-type X` | Exactly declared Type X; no refinements. |
| `--type any` | Every Bead Type; no Type restriction. |
| `--exact-type any` | Invalid. |

`--type` and `--exact-type` are mutually exclusive. `any` replaces the proposed
shell-sensitive `*`; Donna explicitly reserves `any`: it cannot be a legal Type name.
The same rules apply to list and count. Exact declared Type identity does not
mean an exact Type-definition revision.

Compatibility requirements:

- Preserve shipping `--type bug`, `task`, etc. as Issue-classification filters;
  these values are not currently distinct declared Bead Types.
- Default Issue-family selection must preserve existing Issue visibility and
  status/pinned policies, not merely exclude Memory records.
- Count the selected population before delivery limits. A bounded incomplete
  read must not produce a successful partial count.
- Output format must not silently change selection or drop non-Issue members.
  Legacy Issue-shaped JSON compatibility requires deliberate admission rules.
- Do not invent generic meanings for Issue grouping/filter flags in the Type-only
  count increment.

## Preview 4 generic filtering direction

Donna expects the following bounded set to be needed for Preview 4:

- Metadata equality/presence: existing `--metadata-field` and its presence
  equivalent; the existing product spelling is `--has-metadata-key`.
- Property equality/presence: `--property` and `--has-property` counterparts.
- Convenient title, status and body filters.

**Invent zero semantic variation.** Reuse existing matching, value parsing,
case sensitivity and flag conventions. This discussion does not adopt the earlier
assistant suggestions for new presence/null, top-level-key, typed-literal or text
matching rules. Resolve exact flag spellings with the CLI owner instead of
silently renaming existing flags. This is selection work, not output work, and
it does not require a new expression language.

## Ownership and remaining alignment

Donna explicitly requested direct coordination with Cleo and Sybil; both agreed
with the Type-filter landing and a bounded Type-only generic count for Preview 3.
Their proposed ownership split is:

| Owner | Work |
| --- | --- |
| Cleo | Shared list/count CLI admission, help, legacy compatibility and both-provider workflow checks, reusing her existing list stack. |
| Trish / Sybil | Installed Type resolution and applicable transitive refinement semantics. |
| Melinda | Memory-family examples, fixtures and semantics; keep the query/BDP work separate. |
| Janet | Exact integration base, composition, qualification and release checkpoint. |
| Iris | Batch/interchange design where applicable; no reassignment in this discussion. |

Cleo's read-only overlap and compatibility matrix is preserved at:

`/Users/dbox/.local/state/cleo/type-filter-compatibility-matrix-20261022.md`

The existing list descendants edit the same CLI route. Reuse that stack; this
record creates no duplicate runtime lane or automatic merge permission.

Owner alignment still needed:

1. Resolve short legacy Issue classifications versus declared Type names and
   the stable `issue` CLI alias, preserving compatibility.
2. Reconcile Donna's `any` Type-name reservation with Sybil's proposed distinction
   between bare `any` and a full Type URI ending in `/any`. The latter is owner
   feedback, not a separate human ruling.
3. Pin the legacy Issue JSON versus generic records admission behavior and the
   applicable visibility policy for explicit mixed selections.
4. Choose the checked shared selection seam, current stack/base and exact-source
   qualification. Peer agreement is not evidence that this code has landed.
5. Complete the output-template and batch rows separately if needed. Revisit
   wisp exclusion later; no change is required tonight.

## Evidence and authority

Human decisions come from Donna's Melinda chat on October 22. Cleo and Sybil's
replies provide implementation/ownership recommendations, not additional human
product rulings. The source audit inspected current Preview 3 paths and the
frozen PR217 source; Cleo/Sybil record exact inspected integration/list heads in
their compatibility notes. Recheck current heads before implementation.

No runtime files, Type catalogs, BDP pins, shared navigation or release status
are changed by this record. Earlier language surveys, experiments and proposals
remain preserved as historical inputs.

_Agent-Signature: Codex-unknown-model-unknown-reasoning on behalf of Donna Box_
