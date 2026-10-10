# BDP client platform and protocol design

Status: exploratory design, owned by Vickie with product rulings by Donna. Last updated: 2026-10-10.

This Markdown document on `donnabox/beads:codex/vickie-bdp-client-design` is the single plan of record at Donna's request. It replaces the upstream issue and intentionally overrides the fleet's issue-shaped tracker convention. It is a coordination and design record, not a normative BDP amendment or release commitment. Earlier client/journal proposals remain on this fork branch; the BDP editorial and implementation work authorized below has separate upstream review PRs.

## Current priority: BDP profile structure and implementation

BDP specification and executable profiles take priority over further V2 architecture design. Work is authorized in the BDP repository on normative prose, schemas, examples, validation, reference implementations and conformance. Donna retains product rulings; Vickie owns BDP integration and normative changes. Janet owns Beads CLI implementation and Preview 2 release integration; Trish owns Types for Preview 3. No merge is authorized. PR78 remains the earlier client/journal design discussion; profile implementation belongs in separate reviewable BDP PRs.

Keep one specification file, organized into three cumulative, substantially self-contained parts: **Read → Read+Update → Transactional**. Each part defines the minimum concepts needed for its own semantics and may depend only on earlier parts. Read uses Resource revisions and ordinary read/pagination semantics; it must not depend on transaction identity, Scope epochs, commit positions, Events or replication snapshots. Introduce a concept in a later part only where its semantics require it. Capability-specific material such as History must have an explicit profile location instead of leaking later-profile requirements into basic reads.

| Work | Owner / boundary | First deliverable | Done when |
|---|---|---|---|
| Editorial | Bounded editorial worker; Vickie integrates normative changes | Section allocation, concept-dependency map, old-to-new clause/anchor map and contradiction register | Every existing normative requirement is located or explicitly dispositioned; no normative forward dependency; profile examples and schemas agree; unresolved semantics are visible rather than silently changed |
| Implementation | Bounded implementation worker, isolated files/worktree | Current runtime/conformance inventory, then one complete Read+Update operation through the installed reference endpoint | Request validation, persistence, guards, errors, response and observed resulting state pass positive/negative executable cases; no unqualified profile advertisement |
| CLI crosswalk | Vickie, with independent review as useful; Janet supplies qualified CLI evidence | create/delete/list/show/update matrix | Each row pins CLI and BDP evidence and covers IDs/references, guards, merge/replace, JSON, errors, atomicity and Events; classified equivalent, legitimate shape difference, BDP gap, CLI gap or unresolved ruling; Preview 2 relevance explicit |
| Integration | Vickie | Small separate PRs for editorial movement, semantic rulings and runtime slices | Each PR aligns affected prose/schema/examples/conformance; exact tested head and limits recorded; no CLI parity claim without checking both sides |

Implementation proceeds cumulatively: preserve the Read baseline; make Read+Update executable and qualified; then implement Transactional atomic batch, rollback/concurrency, durable outcomes and ordered Events/changefeed/snapshot recovery. Existing evaluators and fixtures are reusable evidence, not proof that either write profile is served. Apply each profile's required contract without importing a later profile's wire concepts for convenience. The existing BDP implementation stack is the starting point; this does not initiate the later Rust V2 rewrite.

Editorial moves and heading/link repairs should be separable from behavior changes. Establish stable clause identities and an anchor migration map early so code/conformance work can proceed while prose moves. Contradictions go into a decision register with evidence and alternatives; only genuine design choices come back to Donna. Proposed Event enrichments and query changes remain proposals until resolved.

Preview 2 is not the delivery target for the profile reorganization or broad write implementation. Record isolated actual compatibility blockers for Janet; release inclusion remains a separate decision. Use current qualified CLI behavior rather than the tip of an unqualified branch. At planning inspection, release tracker #7170 pins CLI staging at `76fde9c9cb0865d430bf12d76fc473490f51128a`; verify this with Janet before parity qualification. BDP main is `182f1fcf8a01d896976bff3c9e3fb87c596c6ca6`; pending PR64 common metadata and PR65 attribution basis must be tracked separately, with exact accepted source chosen before integration. Trish's PR60 is design evidence, not an adopted wire contract.

Use asynchronous bounded workers with disjoint write scopes under this interactive session. Vickie retains design decisions, the integration queue and the agent bus. No separate interactive role is needed to begin the editorial and implementation inventories. Concrete implementation assignments follow those inventories and the agreed profile boundaries.

First-round review surfaces:

- [BDP PR66](https://github.com/gastownhall/bdp/pull/66), Janet PR64/65 integration at `a9bcf9ed1d869aaba2cb49fe9cca10e3b8d52a94`. Corrected 28 later fixture members to the integrated attribution vocabulary. Build/typecheck/boundaries/format and the full bounded suite pass: 108 files, 7,583 tests, one optional skip. The prior sealed Read cohort remains historical evidence.
- [BDP PR67](https://github.com/gastownhall/bdp/pull/67), cumulative profile structure at `c493faf12460fee63e313f9ee83135acee5dbe92`. All 61 original blocks are mapped, all 84 old anchors survive, and 183 links resolve without lower-profile links into later definitions. Migrated 94 citations, preserving case semantics and assertions. The citation integration passed 1,629 conformance/protocol tests; final review repairs passed 970 affected tests. Independent review found no lost protocol guarantee. The clause map records the remaining design questions.
- [BDP PR68](https://github.com/gastownhall/bdp/pull/68), initial durable mutable reference HTTP surface at `ec9df671157a0f4c7a7a484160222254c7777170`. Six Resource singleton mutations use the existing SQLite recovery and shared live reads. Full validation passes: 109 files, 7,588 tests and one optional skip; build, typecheck, boundaries, process restart/replay and installed-package smoke pass. Independent review found no blocker. Development-only, with no full Read+Update or Transactional advertisement.
- [Initial CLI operation crosswalk](preview2-operation-crosswalk.md): five operations with fixed-source evidence, interface translations, conditional guards, metadata/properties semantics, atomicity, JSON and Event boundaries. Qualified CLI pin confirmation and executable cross-repo parity remain outstanding; neither blocks delivery of this source-review artifact.

The first editorial round and initial implementation slice are delivered as drafts, separate from Preview 2. No PR is merged, no release promotion is made, and full write profiles are not qualified. Three-provider council review remains a readiness gate before marking these PRs ready. Remaining implementation work includes aliases, sequences, complete HTTP admission/recovery and qualification; the Transactional slice must atomically commit Resource changes, durable outcome and change group rather than loop over independently committing members. Exact gaps and the next boundary are recorded in PR68's progress document.

The crosswalk identifies two further cold-client decisions: portable efficient recent-first ordering and optional owning-source guards. Accepted main also differs from qualified records on metadata and attribution; PR66 integrates the proposed remedy. Janet owns qualified CLI evidence and release impact. Do not treat these future writer/ordering decisions as new Preview 2 Read blockers.


## Agreed direction

BDP is the provider interface. A provider writes and owns its HTTP endpoint over its chosen store. Our Rust CLI/client is a surface over BDP; commands need not correspond one-to-one with requests. Any provider meeting the required generic BDP capabilities must run the Issues pack without Issue-specific provider code.

Packs build on Gas City core concepts and carry Bead/Link definitions, skills and extensible CLI commands. Issue commands belong in an Issues pack. A shared provider library is possible, but does not take over the endpoint. Brian's team plans the Postgres store; our team designs the client. Current `bd` must coexist side-by-side. Mode/URL/Scope configuration is a candidate, not settled syntax.

Trish owns Type lifecycle design. Janet owns current Beads/Preview 2 integration, CLI/BDP alignment and qualification. Do not edit their branches, PRs, specs, tests, playgrounds or shared databases. Coordinate dependencies through the existing bus. BDP profile implementation is authorized under the current priority above; V2 client implementation and merges remain out of scope.

## Accepted decisions

### D01 — A usable cold client

Accepted by Donna during the design interview on 2026-10-08: **we want a usable cold client.** Common CLI commands must work efficiently over remote BDP without first downloading, indexing or synchronizing the whole Scope. Local caches and replicas are optional optimizations, not prerequisites for ordinary use.

Evaluate the query and transaction proposals against that requirement. The exact command coverage, supported query shapes, scale envelopes and latency/request/byte budgets remain open; this decision does not promise constant latency for arbitrary queries or accept the proposed wire syntax.

### D02 — Required relationship-aware queries

Accepted by Donna during the design interview on 2026-10-08: every provider claiming Issues-pack compatibility must support generic relationship predicates, filtering and ordering sufficient for queries such as “return ten open tasks with no unfinished blockers.” The Issues pack supplies the domain predicates; providers evaluate the generic operations efficiently. Full-Scope download is not an acceptable prerequisite for these ordinary cold-client commands.

Donna wants to keep the expression language bounded. Treat that as a design goal to validate against the command audit; the operator set, nesting/recursion rules and execution budgets are not yet accepted. A bounded expression shape alone does not bound the amount of data examined.

### D03 — Zero domain-specific event types

Accepted by Donna on 2026-10-09: the generic stream must contain **zero domain-specific event types**. Closing, commenting, assignment and other application actions must be expressed through generic resource/state transitions; no special `close`, `comment` or Issue-only event kind is admitted. This does not settle the exact generic vocabulary or comment data model, and is not authorization to change release code. Compatibility with existing consumers requires an explicit migration decision. The [worked five-verb model](five-verb-event-model.md) maps current actions and shows comment, Link, transaction and recovery behavior; its inline-comment choice remains provisional. A proposed [stream-only identity bridge](five-verb-event-model.md#stream-only-bridge-for-legacy-dependency-identity) gives legacy dependency lifetimes opaque Link references while preserving existing CLI/pair uniqueness; it requires durable mapping and sound baseline/continuity.

Comments need not become Beads to satisfy D03. An inline collection of comment objects could be part of the parent Bead's canonical state and change through a generic update. That choice remains proposed, including owner revisions, efficient append and pagination. Current comments already have string IDs, author, text and creation time; current dependencies have mutable JSON metadata. [Source-grounded clarification](beads-1.3-journal-vs-bdp.md#clarification-zero-domain-events-comments-and-link-updates)

## Evidence

The [CLI audit](audit-20261008/README.md) classifies 274 main registry paths and five Preview additions, with source/test contracts and cost analysis. 219 existing BDP tests and 20 additional bounded probes passed. This is parser/schema/fixture/model evidence, not end-to-end Transactional provider qualification. [Verification](audit-20261008/verification.json) pins source commits and checks 287 citation anchors and 160 source files. The [earlier source map](baseline/runtime-boundary-map.md) records 13 runtime interfaces and five paths.

The historical [journal audit](audit-20261008/event-log-audit.md) distinguished Events and change groups. The current review must sharpen that distinction: missing original command intent is not itself an obstacle to forward state replication. Neither before-images nor original selectors are intrinsically necessary to apply complete committed effects. The question is whether today's Event Source contract carries every state/control transition and an adequate atomic replay boundary.

## Beads 1.3 journal comparison

Information exposed by working 1.3.1 should be considered for generic BDP support when it improves efficiency or consumer convenience. The comparison presents [recommended changes to BDP](normalized-1.3-vs-current-bdp.md#recommended-changes-to-bdp) and [recommended changes to 1.3](normalized-1.3-vs-current-bdp.md#recommended-changes-to-13), including the three working 1.3.1 changes: generic close updates, comments in parent state, and durable Link lifetime identities. Specific payload choices and release scope remain open.

The [detailed side-by-side comparison](normalized-1.3-vs-current-bdp.md) holds the proposed close/comment normalization and Link lifetime identity changes constant, then compares modified 1.3 with both current BDP Events and Scope change groups. It separates the now-aligned vocabulary from remaining payload, transaction, snapshot, epoch, coverage, projection and erasure obligations. These assumptions are a comparison scenario, not implementation or release approval.

Donna requested a comparison of the shipped Beads 1.3 journal with BDP before choosing possible 1.4 changes. The [comparison and recommendation](beads-1.3-journal-vs-bdp.md), backed by an [independent source audit](beads-1.3-journal-source-audit.md), finds substantial overlap with BDP's **changefeed**, especially its postimages. The existing journal's transactional capture is reusable; its flat records lack transaction framing, history epochs and snapshot/checkpoint rendezvous, and some writers/state changes can bypass it.

Recommendation, not an accepted release scope: converge the committed-effects capture, preserve the legacy projection, and add the stronger replication guarantees before claiming BDP compatibility. Preserve comment payloads despite contradictory six-op projection guidance. Coordinate with existing journal PRs 7211/7213/7144. No implementation or release artifacts were changed. D01/D02/D03 are accepted; the bounded-expression-language question remains open while this research is discussed.

## Current questions

1. Propose the smallest coherent generic query and transaction additions that make current Issue workflows efficient, with examples beyond Issues, provider costs, limits, authorization semantics and failure behavior.
2. Identify precisely what an Event-only consumer lacks for an authorized current-state replica. Separate this from historical reconstruction, command replay and full provider backup. Compare enriching an event stream with retaining current change groups; avoid redundant independent sources of truth.

## Decision plan

| Step | Deliverable and done-when predicate | State |
|---|---|---|
| C1 — client/provider contract | Interface/responsibility table names invariant and error ownership, excludes direct store coupling and distinguishes accepted/open choices. | Open; direction agreed |
| C2 — CLI over BDP | Every in-scope behavior has a request plan or explicit gap; representative cases state atomicity, requests, bytes and retry/failure behavior. No proposed behavior is presented as implemented. | Bounded audit delivered; choices and runtime qualification remain open |
| C3 — packs | Inspected Gas City provenance, reuse/change list, dispatch/dependencies/version/trust interface and worked Issue command over generic BDP. | Open; Type lifecycle remains Trish's |
| C4 — Rust/coexistence | Library responsibilities and phased migration cover executable selection, configuration/auth/capabilities, compatibility and recovery, with entry/exit conditions. | Open |
| C5 — review | Exact reviewed revision and findings recorded here; Donna's acceptance and Janet's coexistence acknowledgment recorded with provenance; dependencies have owners. | Open; two human acceptance records remain |

The proposals below are ready for discussion. They remain proposals until Donna rules; their names and syntax are illustrative and do not change the current schema.


## Proposal 1: one query model reused for reads, checks and writes

I recommend one bounded, structured query algebra, with three uses. This keeps the provider interface generic while moving selection work close to the indexes. The Issues pack supplies domain predicates; the provider only evaluates the generic algebra.

| Use | Proposed capability | Why it matters |
|---|---|---|
| Find data | `query` with filtering, ordering, selected fields, counts and anchored relationship predicates | Return ten useful records instead of shipping the candidate corpus to the CLI. |
| Protect a decision | `assertQuery` inside an atomic batch | Check conditions such as absence, exact membership or unchanged ordered results at serialization, including newly inserted matching records. |
| Choose and change | `queryApply` inside an atomic batch | Select a bounded ordered set and apply generic mutation templates to it without a read/network/write race. |

### Read semantics

The initial algebra should include scalar predicates, exact array membership, explicit null/missing rules, deterministic ordering with an ID tie-break, selected property paths, exact existence/count and `existsLink` with direction/type/Link predicates plus a predicate on the opposite endpoint. Negation supplies “no related record satisfies this condition.” The same mechanisms work for appointments, inventory reservations, documents awaiting approval and equipment with expired certificates.

Text search must state its semantics. For current CLI compatibility we need literal substring search, with a protocol-defined Unicode case-folding mode where requested. Database-default collation or a ranked full-text query is not an equivalent substitute. A provider may choose its indexes, but must implement the same meaning. No arbitrary SQL, custom server function or provider-loaded Issue code is needed.

Ordering and projection reduce wire cost; they do not guarantee cheap server execution. Broad relationship predicates and substring queries can still scan. Required query shapes need indexed, measured scale envelopes, with explicit budget failures. Exact counts may be expensive; an estimate must never masquerade as an exact count.

Related queries can share an expiring immutable **read handle** covering both collections at one Scope position and Authorization View. Return it with the first query, rather than requiring a separate round trip. This does not create a client-controlled open write transaction. Providers choose how to retain the view; it need not pin a PostgreSQL connection per client. Full snapshots remain the replication bootstrap; projected query rows are visibly distinct from full canonical records.

Bounded reachability is an additional generic capability for descendants and transitive dependencies. Depth/visited-edge/result limits must be explicit. If a complete negative result cannot be proved within the budget, report incomplete/limit failure instead of returning false. It may be optional for generic BDP implementations, but must be required by any Issues-pack compatibility profile that promises those commands.

### Transaction semantics

`assertQuery` is a read assertion, so a reader does not need mutation permission on every dependency simply to guard a decision. It can assert zero matches, expected cardinality, or a prior observation token representing the ordered result identities and revisions. The provider rechecks or proves equivalence at serialization. A token compresses the request; it does not eliminate validation work. A phantom, changed ordering or changed relevant revision must invalidate the observation.

Assertions see preceding staged effects. They are not blindly checked against final post-state: claiming an unassigned task intentionally changes the state that passed its precondition. The enclosing batch must nevertheless serialize the checks and mutations against concurrent writers.

`queryApply` freezes an ordered selection when reached, then applies a bounded template of existing generic mutations. A take-one selection can update the chosen task and create a lease Link in the same transaction. Bindings initially cover selected identity/revision, named creations and narrowly specified scalar assignments. This is a finite operation plan, not arbitrary server code. The complete mutation remains subject to operation, resource, byte and induced-event limits.

Keep **match cardinality** distinct from **take**: “exactly one exporter exists” differs from “choose one available item among many.” Idempotent replay must return the original selected identities and outcome. Strict queue order cannot silently become `SKIP LOCKED` opportunistic selection; relaxed selection would need a separate explicit contract.

Partial-success multi-close is the one larger control-flow question. A possible optional capability is a flat guarded group: false business precondition skips that group's mutations; true executes them; all applied groups still commit together. Structural/auth/limit/execution errors abort the batch. This is not general exception catching and does not promise parity with every legacy per-item failure until tested. Start with query/assertion/selection, and standardize guarded groups only against the precise compatibility cases. A full Issues profile must require the capability or declare an approved semantic change.

### Concrete CLI journeys

| Journey | Proposed request plan | Target, not a measured result |
|---|---|---|
| `ready --limit 10` | Wake expired dated deferrals, then query scalar + relationship predicates with projection; share one handle for the recent/older hybrid-order buckets. | Usually 2–3 application requests after discovery and O(returned rows) wire payload; bounded graph evaluation stays at provider. Separate handle creation or receipts can add requests. |
| `ready --claim` | One batch performs any required wake and ordered `queryApply(take:1)`, patching holder/status and creating lease state. | One mutation submission for simple ordering; no whole candidate list. Hybrid ordering needs conditional fallback or a guarded probe/retry. Receipt polling and contention retries remain possible. |
| `close A B --claim-next` | Guard each target's policy, apply permitted closes, then choose next work against the staged state. | One atomic unit if guarded groups are selected; otherwise explicitly documented preflight/retry or changed semantics. |
| Search or small graph view | Exact text/relationship query against a consistent handle; project only display data. | No mandatory full-Scope bootstrap for an ordinary cold command. Index/query-budget qualification is still required. |

The [detailed generic proposal](generic-bdp-proposals.md) defines the operator semantics, limits, PostgreSQL implementation options, visibility rules and three worked journeys. It also explains where a local index remains useful, notably fuzzy pairwise comparison and repeated graph analysis.

### Limits of this proposal

All reads and joins operate within the authorized view. A hidden blocker cannot silently become evidence of “no blocker.” A command that promises authoritative global policy needs a principal with the required complete view, or a separately defined conservative completeness contract. Refusal must not itself reveal whether a particular hidden record exists.

These mechanisms enforce a submitted predicate; they do not force every client to submit an Issues-pack policy. Decide separately whether those policies are cooperative-client guarantees or generic authority-enforced constraints. Do not add private provider Issue logic to bridge that distinction.

Clock/lease semantics and no-history operational state remain separate decisions. Query constants can initially use an explicit client cutoff, but strict authoritative lease timing would need a generic, precisely defined transaction-time value rather than casually using server wall-clock functions. No-history heartbeats cannot be claimed solved merely by placing them on a separate versioned Resource. Identity rename, full history restore and native database administration also retain their explicit migration questions.

## Proposal 2: a replica-capable event stream

**The prior answer was too broad. Events can reconstruct a replica.** Missing original commands, selectors, guards and before-images is not the problem. For ordinary forward replication, create inserts a record, update applies a committed delta to the known predecessor, and delete removes it. A deleted record's previous body is needed for undo or retained history, not to remove it from current state.

BDP already provides most of those payloads, including owned-Link source deltas. The earlier fixture probe demonstrates event-folded state matching the change-group postimages for its sample. That is positive evidence, though only for the fixture.

The precise question is whether the **current Event Source delivery contract** carries every transition and enough framing to maintain an atomic, resumable, authorized current-state copy. It presently does not promise all of that.

### The smallest state counterexample

Start with a replica containing Bead A. The current specification permits A to leave a projection while remaining alive in the authority and while the view token stays stable. A change-group tombstone communicates that removal. But there is no actual `deleted` lifecycle fact, and an `updated` fact is exposed only when visible in both pre-state and post-state. An Event-only consumer can therefore retain A incorrectly.

On entry, a previously hidden existing Bead needs a full visible image: it was not just created, and the consumer does not possess a base for an ordinary patch. This needs explicit projection-entry/exit transitions, or a rule requiring resnapshot on every such membership change. The latter is simpler but potentially expensive. Grants/revocations already rotate the view and are a separate reset case. [Current projection and visibility rules](https://github.com/gastownhall/bdp/blob/182f1fcf8a01d896976bff3c9e3fb87c596c6ca6/docs/specs/bdp.md#L6037-L6140).

### The remaining replica contract

| Requirement | Why today's Event Source is insufficient or awkward | Minimal event-stream solution |
|---|---|---|
| Snapshot handoff | Snapshot returns a Scope checkpoint; Events accept source-local Event IDs, with no defined rendezvous between them. | Snapshot returns/accepts the exact exclusive stream checkpoint. |
| Atomic transaction delivery | Individual SSE Events share a transaction ID but lack an explicit completion boundary; ordinal gaps can be legitimate. | Deliver a complete transaction frame or explicit completion marker. Persist state and cursor atomically. |
| Canonical initialization | Creation Events omit Type-derived empty owned-Link keys. | Complete creation/entry image, or stable metadata resolution. Full images on every update are unnecessary. |
| Recovery and freshness | Source-local delivery progress is not durable applied state or a Scope-position guarantee. | Epoch/view fences, durable checkpoint, continuity and explicit replay-expiry/reset semantics. |
| Erasure | An archive may retain erased prior content while lifecycle Events show no live-state change. | Erasure records and a bootstrap ledger in the same stream contract. Historical erasure alone is not a counterexample for a copy that retains only unaffected current state. |

Atomic reconstruction from individual Events is not mathematically impossible: additional finite reads reaching the current source end could supply completion information. The framed stream makes that guarantee direct and efficient, especially when the final transaction is followed by silence.

### Recommended shape

Keep one committed transaction envelope with Scope epoch, Authorization View, previous position, position and checkpoint. Make its ordered entries sufficient to reconstruct the projected state:

- Complete creation/projection-entry records.
- Committed property/owned-Link deltas with predecessor and successor revisions.
- Deletion/projection-exit identities, with the meanings kept distinct.
- Erasure and progress/reset control records where needed.

Activity/resource feeds become filtered views of that stream. Final postimages can remain a derived transfer mode for simpler clients and efficient catch-up; they need not be an independently authored truth or mandatory duplication for every small patch.

Current BDP already atomically commits Events and final-state changes together. It does **not** require two independent logs. The proposed change is to make the canonical transition stream reconstruction-complete and make the relationship between its views explicit. Providers remain free to use relational transactions and an outbox/change table, event sourcing, or another compliant implementation. We should require event-fold/read equivalence, not dictate the storage engine's internals.

A command request, a committed transition and a retained historical record remain different concepts. Converting Events back to mutation requests is unnecessary and can duplicate derived effects; directly reducing those facts into replica state is the appropriate operation. Full provider backup adds administrative and historical state beyond this current-state replica contract.

The [source-grounded reassessment](replica-reassessment.md) records the exact counterexamples and qualifications. My recommendation is to evaluate this event-first wire design against the existing delta-plus-postimage group design, preserving snapshot, atomicity, projection and erasure guarantees in both.

## Validation and next decisions

No runtime protocol changes were made and no new runtime conformance was claimed. The original audit tests remain evidence for the baseline only. This revision is source-grounded design review; its [source anchors and migrated files are checked mechanically](design-verification.json).

Before accepting the proposals, require:

1. Differential query and command results on two providers, including a non-Issue domain using the same operators.
2. Phantom/queue races, view changes, strict ordering, expired handles/tokens, limits and ambiguous-retry tests. Publish cold/warm byte/request/latency budgets rather than assuming that expressiveness implies efficiency.
3. Snapshot plus event-fold equivalence to canonical reads through ordinary and owned mutations, projected entry/exit, multiple updates in a transaction, reconnect/crash, quiet-stream completion, erasure and epoch/view reset.
4. Byte/CPU/storage comparison of event deltas versus final postimages on large records, small patches and graph batches. Keep provider internal architecture free.

The first decision I recommend discussing is the shared query algebra and transaction selection/assertion contract. For the log question, first agree that the target is an authorized, atomic current-state replica; that lets us resolve the actual stream gaps without dragging command audit or full backup into the requirement.

## Workspace and migration receipt

This work moved from the former upstream issue into this fork branch after Donna requested a quieter design workspace. The full previous issue and its comments are preserved in [the historical archive](legacy-tracker-archive.md); old links inside evidence are not live authority.

The branch and archive were verified before attempting issue removal. GitHub denied `DeleteIssue` because the current account lacks permission. The former issue is now closed as not planned, its current title/body replaced with a withdrawal, and all four author-written comments deleted. One automatic triage comment remains; actual issue deletion requires an upstream administrator. No replacement issue or PR was created. The fork itself is public.
