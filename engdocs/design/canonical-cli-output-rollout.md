# Canonical CLI output implementation plan

Status: first read-adapter wave merged into `codex/preview3-integration` on
October 20, 2026. The target contract is
[selection-and-output.md](../../docs/reference/bdp/cli/selection-and-output.md).
This checklist separates delivered functionality from remaining work; it is not
a declaration that the complete output contract or Preview 3 release is finished.

## Delivery ledger

The five implementation PRs are merged at integration commit
`a40452caefe2f1d57d2b6a980445b880ad24c98c`.

| PR | Delivered | Remaining boundary |
| --- | --- | --- |
| [#123](https://github.com/donnabox/beads/pull/123) | Shared Resource projection, schema-2 envelopes, detail/JSON/raw output, structured failures and protected context. | Source-side projection remains BDP work; this renderer does no fetching. |
| [#126](https://github.com/donnabox/beads/pull/126) | Collection descriptors, tables, finite JSONL and delivery-boundary accounting. | An engine capable of consuming pages does not supply provider cursors. |
| [#127](https://github.com/donnabox/beads/pull/127) | Current and retained `show`/`link show`, BDP-shaped payloads, masks, errors and public corpus. | No JSONL show adapter or hidden traversal; nested masks are limited to properties/metadata. |
| [#128](https://github.com/donnabox/beads/pull/128) | `versions`/`history`/`link versions` and root/Link comparisons. Subject context, exact operands and coverage survive framing removal. | Version JSONL and retained paging refuse; comparison data masks refuse. |
| [#129](https://github.com/donnabox/beads/pull/129) | `list`/`link list`, structured exact-Type/family/endpoint selection, tables and framed/raw JSONL. | Local provider returns complete bounded selections; it cannot resume or safely truncate them. |

### Qualification already recorded

The PR bodies retain final source identities and actual Linux run receipts:
[#123](https://github.com/donnabox/beads/actions/runs/39448828081),
[#127](https://github.com/donnabox/beads/actions/runs/39452980710) (attempt 2),
[#128](https://github.com/donnabox/beads/actions/runs/39456238030), and
[#129](https://github.com/donnabox/beads/actions/runs/39465085335).
Their selected lanes passed: 217 test targets, 175 integration targets with nogo,
10 Dolt-server targets, two command-server targets, pure-Go/js/wasm and all eight
configured release cross-compilation targets. Optional skipped tiers are not
credited. #126 has its own focused Bazel race/nogo evidence and byte-identical
engine files in the fully qualified #128 stack; no standalone full #126 workflow
is claimed.

Native and Claude review dispositions are recorded; Gemini supplied no substantive
review and is not counted. #127's first run had an unchanged Notion test failure;
its failed integration lane passed on retry without source edits. The original
failure and bounded diagnosis remain in the PR. #129's help-expectation failure
was repaired and its final source passed a fresh run.

Combined qualification with selected Types, historical recovery, endpoint pins
and export belongs to [#141](https://github.com/donnabox/beads/pull/141). Individual
PR receipts do not qualify that combined tree. The merge retains the shared
current/retained projector while adding selected owned-Link families and pinned
References; its tests must exercise those combinations.

## Remaining-work checklist

Order reflects the next useful slices, not a blanket dependency on one owner.
All implementation owners retain their own validation and publication authority.

- [x] Common projection/framing engine and current/retained identity reads.
- [x] Complete bounded list/Link-list output and retained-state inspection.
- [ ] **1. Provider pagination:** establish portable continuation and read-context
  guarantees with Melinda and Trish; implement resumable delivery for list/Link
  list. Cover short pages, total caps, format/projection changes under continuation,
  late failure and no skipped or duplicate records. Until then retain explicit
  refusal when a requested page/limit is smaller than the complete selection.
- [ ] **2. Retained JSONL and paging:** define a delivery unit that preserves the
  subject and kind even with `--omit envelope`; then admit versions/history JSONL
  and real retained continuation. A version row alone is insufficient. Comparisons
  remain compound results; they need no row stream merely to match collections.
- [ ] **3. Generic mutation output:** adapt Bead and Link create/update/delete with
  actual applied/no-op/unknown outcomes, revision/source effects and stale-guard
  failures. Rendering failure must not imply rollback or safe retry. No dependency
  on the query parser; coordinate actual overlap with historical update work.
- [ ] **4. Exact generic count and query/search:** share the normalized selection
  model; exact count may precede textual predicates when a qualified provider
  supports it. Melinda owns the language and lowering. Never report a page count
  as a total or wrap Issue-shaped results as generic Resources.
- [ ] **5. Generic graph and admitted Type results:** add declared result adapters,
  Link inclusion/exclusion, topology-preserving table/tree output and protected
  exact revisions. No new facade or Issue workflow migration is implied.
- [ ] **6. Release acceptance:** close the per-command compatibility matrix and
  applicable O01–O17 coverage, publish corpus examples, and qualify the actual
  combined release tree. Document every remaining capability refusal. A changed
  BDP pin requires payload/schema/conformance adoption together.

Completion means every admitted command has an explicit supported-format,
projection, selection and failure contract backed by command-level evidence.
Excluded domain workflows remain unchanged. Deferred protocol extensions and
facades are not prerequisites for declaring the bounded read wave delivered.

## Scope

Build the clean interface around generic Bead/Link operations. Preserve shipping
Issue/Dependency workflows and route users who need arbitrary Types, field
projection or the new framing to those generic operations. Do not spend Preview 3
retrofit effort on `bd dep`.

| Treatment | Command families |
| --- | --- |
| Adopt first | `bd show`, `bd link show`; then generic `bd list`, `bd query`, `bd link list`. |
| Adopt with dedicated result adapters | `versions`, `history`, `compare`, their explicit Link readers, and the generic graph view. |
| Adopt for mutation results | Generic Bead `create/update/delete` and `bd link create/update/delete`. Kind-specific mutation dispatch remains unchanged. |
| Adopt when their generic semantics are ready | Search hits, aggregates/count, Type Resource inspection, and a Link query wrapper if the query design admits it. |
| Excluded from this rollout | `remember`, `recall`, `memories`, `forget`, and future Link/Memory facades. Future adoption requires a separate scope decision; export retains its separate interchange contract. |
| Leave on shipping contracts | `dep` and all its children; `ready`, `blocked`, Issue lifecycle, assignment/priority, comments, work hierarchy, epic/gate/state workflows, labels, formula/molecule/workflow conveniences. |
| No normalization project here | Init/setup/doctor, serving, maintenance, sync, integrations, administrative reports, raw SQL, and other operational commands. |

Inclusion is by semantic command path, not a root-command name alone. A generic
Resource may have a task Type and still use the new contract. Conversely, a root
command that returns an Issue-specific report does not become generic because it
can print JSON. `graph`'s generic view can adopt the contract without altering its
work-policy projection or checks.

Existing domain commands may later use common internal utilities, but that must
not change their public defaults, flags, errors or output. Do not offer a subset
of the new flags on `dep` that encourages users to treat it as the generic API.
Use help/documentation examples to show the corresponding generic route.

## Command-by-command adoption matrix

This is an explicit allowlist for the normalization effort. Every command path
not admitted below stays outside this rollout unless separately approved. An
entry describes its generic operation, not permission to rewrite a native Issue
route with the same spelling. Output-only means selection and mutation semantics
stay unchanged; it does not mean a new JSON wrapper can conceal an Issue-shaped
payload.

Source cross-check: Melinda's command inventory and registrations at
`116a4ea6a459ef823ea714c1f418ad4c39dd81af`; the older output-design checkout remains
documentation-only. Use the selected integration baseline for new implementation work.
Melinda reviewed the command dependencies on October 19. Her key distinction is
the shared structured selection model versus its textual predicate frontend:
list and simple exact count can adopt the former before the latter exists.
Query/search need their new predicate semantics; Resource identity reads and
mutation output do not. The table below incorporates that recommendation.

| Command path | Planned change | Query-language dependency | Batch |
| --- | --- | --- | --- |
| `bd show` (generic route) | Resource descriptor, include/omit, detail/JSON/raw output and errors; preserve exact/current selection. | None: identity lookup. | A: first vertical slice. |
| `bd view` | Existing alias inherits only the adopting `show` route; no separate implementation. | Same as `show`. | A. |
| `bd link show` | Same output service with Link identity resolution and endpoint protection. | None: identity lookup. | A. |
| `bd list` (generic route) | Normalize simple selectors, ordering, masks, paging and JSONL/table output together. | Structured selector model first; full predicate parser not required. | B, then C for richer predicates. |
| `bd link list` | Same collection contract for Links, including Type/endpoint filters. | Same structured model as list; no independent Link filter engine. | B, then C. |
| `bd count` (generic operation) | Exact count of the same selected population, with aggregate output. | Current root count uses Issue filters; its generic route needs the normalized selector/count operation. No output-only rewrapping. Full textual parser is not required for simple selectors. | B where exact generic count is supported; predicate/grouped counts follow their C operators. |
| `bd link count` | Exact count matching Link list selection; never a displayed-page count. | Same as `link list`, plus exact count capability. | B if supported. |
| `bd query` | New predicate/query frontend plus the common result descriptors and output. | Direct dependency on Melinda's selected grammar and normalized request. | C: language slice; not an output-only rewrite of legacy expressions. |
| `bd search` (generic operation) | Text-search convenience over shared predicates, declared hit shape and output. | Text semantics, case behavior and hit/result contract; existing Issue search is a compatibility path. | With the relevant C text-predicate slice. |
| `bd versions` | Version-item descriptor, projection, framing and supported pagination; no full-snapshot fetching to fill summaries. | No current-population query parser; actual History capability required. | Independent follow-on after A/B. |
| `bd link versions` | Same version service for Links. | Same as versions. | Paired with versions. |
| `bd history` (generic route) | Map to the actual graph History/version result contract; share output with versions where it is an alias. | History semantics, not ordinary current-Resource predicates. | With retained-state readers. |
| `bd compare` | Comparison descriptor; preserve operands, exact revisions and differences; share framing and meaningful projection. | Exact retained reads and comparison service, not the new predicate parser. | After its operand/result descriptor is fixed. |
| `bd link compare` | Same comparison contract for Links. | Same as compare. | Paired with compare. |
| `bd create` (generic Bead route) | Canonical mutation result and errors; no change to creation or Type policy. | None. | Independent write-output follow-on after A. |
| `bd update` (generic Bead route) | Same framing with applied/no-op/unknown and revision/source effects intact. | None; coordinate file/order overlap with historical-base and symmetric-edit work. | Write-output follow-on. |
| `bd delete` (generic Bead route) | Same framing with actual deletion/preview effects preserved. | None; lifecycle/cascade redesign is not part of output work. | Write-output follow-on. |
| `bd link create` | Link mutation result, including owning-source effects. | None. | Paired with generic create output. |
| `bd link update` | Link mutation result with Link/source guards and no-op/unknown preserved. | None. | Paired with generic update output. |
| `bd link delete` | Link deletion result and source effects. | None. | Paired with generic delete output. |
| `bd graph --view generic` | Existing summary can receive honest result framing; full Link selection/topology/table/tree output needs the new graph operation. Do not call current summary nodes complete Resources. | Summary framing has no parser dependency. New traversal/family selectors need shared graph operations; authored predicates additionally need C. | Summary adapter after A if useful; full graph D. |

The lettered batches describe dependencies, not a mandatory single-file queue.
After A, the structured-selector work for B, mutation output, retained-state
descriptors and the query frontend may progress independently when owners and
contracts permit. C should enrich list/query/search through one model, not cause
a second round of command-local filter implementations.

### Conditional and deferred adapters

| Surface | Disposition and reason |
| --- | --- |
| `bd link query` | Future facade, explicitly excluded from this rollout. Generic query can select Links through the shared population model; no new wrapper or separate language here. |
| `bd link search`, `bd link history` | Not registered in the inspected Link command set. No new verbs in this output rollout; existing generic query/reader routes are the design inputs. |
| `bd types` | Optional installed-catalog output adapter after A, using its actual descriptor shape. Type-as-Bead behavior waits for Trish's contract; configuration/aliases must not be labeled Resource rows. No new `link types` command. |
| `bd memories` | Explicitly excluded from this rollout; preserve its current contract. |
| `bd recall` | Explicitly excluded. Preserve its current body-only contract; do not prepend an envelope to raw body output. |
| `bd remember`, `bd forget` | Explicitly excluded; preserve their existing mutation and output contracts. |
| New assessment/coverage/explanation commands | Future Memory facades are explicitly excluded. The common output contract records context-preservation obligations for any separately approved future adapter; it does not authorize new commands. |
| `bd export` | Iris coordinates selection reuse and packaging; generic JSONL is not a replacement for the transfer/legacy export format. |
| `bd import` | No generic rendering rewrite here; preserve interchange/admission semantics. |
| `bd batch` | Separate operation/transaction/result contract; not automatically changed by adopting output on individual mutations. |
| `show/list --watch` | Separate refresh/subscription contract; not the finite JSONL header/items/terminal contract. Defer canonical admission. |
| `show --children/--refs/--thread/--include-comments`, legacy list tree/work filters | Preserve existing route behavior; do not quietly pull expansion/workflow semantics into a complete no-chasing Resource read. Generic equivalents require explicit query/graph mappings. |

### Explicit exclusions

| Commands / family | What remains unchanged |
| --- | --- |
| `bd dep` and every child (`add`, `remove`, `list`, `tree`, `cycles`, `relate`, `unrelate`, aliases) | Shipping direction, defaults, work policy, flags, JSON and errors. |
| `bd link SOURCE TARGET` shorthand and existing compatibility aliases | No incidental modernization; their shipping semantics remain compatibility inputs. Do not resurrect removed preview aliases. |
| `ready`, `blocked`, `children`, `epic`, Issue `graph`/`graph check` | Domain selection, blocking/readiness, hierarchy and work checks. |
| `close`, `reopen`, `defer`, `undefer`, `assign`, `priority`, `note`, `edit`, labels/tags, comments | Issue mutation/editor/feed behavior. Generic Resource writes receive the new contract separately. |
| `todo`, `human`, `gate`, `state`, `statuses`, `status`/`stats`, `stale`, `orphans` | Named work views and their extra policies/external facts. |
| `mol`, `wisp`, `swarm`, formula workflows, `merge-slot`, claims/leases/heartbeats | Workflow, coordination and compound mutation semantics. |
| `duplicates`, `find-duplicates`, `lint`, repository `diff`, native commit `history` | Analysis, similarity, validation and Git/commit meanings; not generic Resource outputs. |
| `prime`, native key/value commands, setup/configuration, serving, doctor, maintenance, sync and integrations | Context/control-plane/operational contracts. |

Excluded commands may remain useful examples in Melinda's expressibility corpus.
Proving a generic query could express their reads is different from migrating
their public CLI. Their unchanged output must be a regression check when shared
utilities are refactored.

## Architecture

One shared pipeline should own:

1. Output-option validation, pointer parsing and incompatible-option errors.
2. Command result descriptors: BDP pin, result kind/schema, projection roots,
   protected context, delivery unit, formats and paging capabilities.
3. Representation projection over original values, including recursive identity
   protection and original-index array handling.
4. Page consumption and boundary accounting, using the query/provider continuation
   contract rather than locally invented cursors.
5. JSON/JSONL framing, buffered ordinary JSON, terminal outcomes and raw mode.
6. Detail/table/tree presentation and stable diagnostics.

The query layer owns candidates, bindings, semantic projection, order, aggregation
and read context. Renderers receive selected typed results; they do not perform
store reads, follow endpoints or reinterpret Issue fields. Mutation services keep
their existing atomicity, guards, ownership and outcome semantics.

The CLI adapter maps a command result into the common pipeline. It should not
copy a giant output switch into every Cobra command, encode arbitrary domain
knowledge into the Resource projector, or pipe one CLI command through another.
Do not require a universal BDP wire payload for CLI-derived results BDP does not
define. Their descriptor identifies their contract explicitly.

## Acceptance criteria by work area

The delivery ledger and remaining-work checklist above determine current status.
The criteria below remain obligations for each applicable adapter.

### 0. Reconcile contracts and isolate compatibility

For each newly admitted adapter:

- Pin the adopted BDP commit/spec identifier and corresponding payload schemas.
- Use the accepted error/completion fields, result-kind registry and indexed
  inclusion rules; declare the adapter's supported subset explicitly.
- Adopt Melinda's separation between semantic query projection and representation
  masks, including mandatory assessment context and strict continuation.
- Map existing flags/output for each overlapping root path, with examples from
  shipping v1.3.1 and the selected Preview 3 integration commit.
- Check for actual shared-file conflicts with other owners and Janet. Work in
  isolated checkouts and record any agreed collision order under
  [MEM-POL-003](https://github.com/gastownhall/beads/issues/7656#mem-pol-003-independent-ownership-and-green-ci-code-flow);
  hypothetical overlap is not a publication or merge hold.

The unresolved `--allow-partial` extension does not need to block strict `show`
and ordinary collection output. Do not silently admit it before its completion
and exit semantics are specified.

### 1. Complete one read vertically: show

Use `bd show` and `bd link show` to implement and verify the Resource descriptor,
field masks, protected identity/references, schema-2 envelope, detail output,
raw output and structured failures together. Exercise a task, a non-Issue Bead,
a Link and an owning Bead with inline owned Links.

This proves the shared abstraction before introducing traversal or paging. It
also catches the current temptation to convert Memory into an Issue-shaped DTO.
Root and explicit Link readers must use the same service and contract, with
their documented identity resolution.

Completion gate: the same selected Resource produces matching BDP-shaped JSON
and explicit projection across detail/JSON; malformed/conflicting options fail
before any service call; no renderer performs extra fetches.

### 2. Collections, paging and JSONL

Integrate generic `list` and `link list` with the common selected-result stream,
plus generic `count` and `link count` where exact counts under that same structured
request are supported. Root `query/search/count` currently use Issue filters,
even in graph mode; none becomes generic by receiving a new envelope. Simple
generic count needs the new selector/count operation, but not the textual parser.
Generic query and search land with their corresponding predicate-language work.

Do not make the textual query parser a prerequisite for useful list output.
Start with the shared structured selector request and operations admitted by its
provider. Later predicate grammar lowers into the same model. Do not claim full
query parity from the first adapter.

Add page size versus total cap, continuation, JSONL header/items/terminal,
empty selections, original-index array masks and default tables. Preserve
query/result membership across formats and do not lose undisplayed items at a
limit boundary. Force failures on later pages and after emitted JSONL items.

Completion gate: strict query semantics, output framing and provider continuation
agree; no hidden 50-item cap in unpaged JSONL, no false collection exhaustion, and
no incomplete aggregate/negative predicate presented as a completed answer.

### 3. Generic writes and retained-state inspection

Retained-version output, pure mutation output and query-language execution can
be bounded parallel follow-ons after the shared read foundation. They need not
wait for each other when their file ownership and contracts are independent.

Mutation adapters can follow the Resource output foundation without depending on
the full query grammar. Add complete mutation result framing, applied versus
no-op, predecessor/source effects and unknown outcomes. Root writes stay
Bead-only; Link writes stay under `bd link`. A failed printer must not suggest a
committed mutation was rolled back or is safely repeatable.

Add `versions`, `history` and `compare` with their actual selected contracts.
Version lists, event History and comparison results are not interchangeable
Resource arrays. Protect exact revisions, operand order and effect information.
Coordinate with Janet's historical-base update work rather than implementing it
inside an output change.

Completion gate: successful output and failures preserve existing revision,
ownership, attribution and exactly-once semantics. No storage feature is inferred
from a new format flag.

### 4. Generic graph and derived results

Add Link selectors, graph table/tree output, cycle/shared-node references and
parallel Links. Use complete compound graph JSON items initially; do not invent
a graph event stream merely because JSONL exists.

Integrate generic search/count and Type inspection as their result adapters are
defined. Preserve selected exact revisions; do not coalesce them into current
identity. Memory facade commands and future Link facades are outside this rollout.
The shared pipeline's assessment-context obligations remain available to separately
approved future adapters; implementing those facades is not a completion gate here.

Completion gate: non-Issue examples work without fabricated fields and no
projection erases the meaning of a represented graph edge. Future assessment
adapters must separately qualify their collective-context obligations.

### 5. Separate follow-ons

- Source-side BDP projection is a protocol amendment, not a flag-only optimization.
- `--allow-partial` needs its shared query/result/exit contract.
- Export adapts the shared values without replacing its own package/integrity
  protocol; legacy raw export stays compatible until explicitly versioned.
- Disk spooling is optional later optimization, not the first output milestone.
- Numeric exit-code normalization is a later compatibility decision.

## Compatibility contract

The clean engine has one contract. Compatibility belongs at the command boundary.
For existing `show/list/query` scripts, a schema bump alone is not enough to
authorize replacing bare arrays or legacy detail objects.

The delivered adapters use explicit `--output-format` as the canonical-output
opt-in in graph mode. Historical `--json` and `--format` retain their existing
behavior without that opt-in. Mixing either with `--output-format` refuses,
including explicit `--json=false`. Projection flags require canonical opt-in.
This transition does not change the pure model where `--json` aliases canonical
JSON; there is no scheduled legacy cutover or removal in this wave.

Canonical `list --all` refuses because it is an Issue-status flag. Existing
legacy `--all` behavior remains unchanged. Canonical list selection has no
implicit Issue-status filter. Paged human/JSON defaults to 50; JSONL defaults to
unpaged. The provider bounds a complete selection to 1,000 Resources and 16 MiB.
A selection exceeding the requested page/limit refuses, with no subset presented
as exhausted and no fabricated continuation. `--unpaged` removes the presentation
page boundary, not the provider's acquisition bounds.

Before release, decide the exact cutover and retirement mechanism with Donna.
Do not add a new permanent top-level namespace solely for migration, silently
repurpose hidden global `--format`, or maintain two separate implementations of
field projection. Preserve existing `--all`, total `--limit`, pager behavior and
domain-specific output flags until individually mapped.

The adoption matrix should name command path/context, old shape, new shape,
default, opt-in, warning channel and scheduled removal if any. Machine stdout
must never carry migration hints mixed with result data.

## Validation and ownership

Implement O01–O17 from the spec through shared-contract tests and a small set of
real command journeys. A flag-help snapshot is not runtime evidence.

Required journeys include complete/non-Issue/owned-Link Resource reads; missing
and null properties; array masks; raw and framed outputs; empty/limited/resumed
collections; a later-page failure; a midstream error; mutation no-op/stale guards/
unknown result; graph cycles/parallel edges; collective assessment protection;
and unchanged domain-command behavior.

Avoid testing each renderer by duplicating its implementation. Use semantic
invariants, golden public payloads where stability matters, and injected provider
failures. Verify embedded/shared-server paths where admitted; report unsupported
capabilities explicitly.

Each implementation owner owns formatting, build metadata, tests, review and PR
preparation for their isolated slice. Cleo owns this output contract and the
`internal/clioutput/` foundation. Janet coordinates shared command files,
database/service use, release integration and installed-product qualification;
that coordination does not gate independent package validation or PR preparation.
Melinda owns query semantics; Trish owns shared BDP/Types contracts; Iris owns
interchange. [MEM-POL-003](https://github.com/gastownhall/beads/issues/7656#mem-pol-003-independent-ownership-and-green-ci-code-flow)
supersedes earlier blanket authoring, validation and publication holds, including
agent-only ordering holds. A real collision requires both owners' and Janet's
agreement and an entry in the policy's visible hold register; explicit
Donna-requested exceptions and repository protections remain in force.

This document and the specification are documentation-only changes. Delivered
implementation and review evidence are recorded above; shared-file coordination
does not impose a new approval or qualification gate on follow-on authors.

Trish's [bounded BDP review](https://github.com/donnabox/agent-coordination/blob/7fbe9e41413198b3af9ed04ddae6a82adb0e7be1/context/trish/output-contract-review-20261019.md)
confirms that client-side projection and framing need no new protocol envelope.
It also records the existing wave 3 discrepancy between the consumed BDP wire
schema at `de99030d13a57e77f4f2660b91ea3c39b82823cb` and the intended
`revisionProvenance` contract. Preserve that single coordinated adoption task;
do not repair it by renaming output fields or citing an unconsumed integration
head. Collection, predicate and source-side projection dependencies remain
separate from the first identity-read adapter.

## Publication

The common contract is published through
[PR122](https://github.com/donnabox/beads/pull/122), with its proposal mirror and
work tracker at [#7718](https://github.com/gastownhall/beads/issues/7718).
The older graph draft links to the common contract; the CLI navigation entry
is included in this documentation PR alongside the existing Memory/reference
pages. No separate navigation assembly or Janet approval is required.

Melinda's [query-language proposal](https://github.com/donnabox/beads/pull/125)
contains the portable language, CLI conveniences and acceptance corpus. Its
adapters reuse this contract. Trish owns BDP payload/continuation reconciliation,
Iris owns interchange, and Janet owns the combined integration qualification.
Cleo owns the output specification and remaining-work checklist. Track actual
implementation gaps here without relabeling approved design as undecided or
claiming pending adapters are delivered.
