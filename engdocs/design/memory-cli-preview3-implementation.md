# Memory CLI: Preview 3 implementation slices

## Status and authority

Donna accepted the CLI semantic target on October 21, 2026. This is the implementation plan and authored acceptance inventory, not qualification evidence. The CLI contract belongs in the [Memory facade](../../docs/reference/bdp/cli/types/memory/memory.md); shared visual behavior belongs in the [graph CLI draft](../../docs/reference/graph-cli-specification-draft.md). The [semantic cases](memory-cli-preview3-cases.json) are logical fixtures, not a wire representation, public query language, canonical import format or executed test suite.

This docs branch starts from `donnabox/beads:codex/preview3-integration` at `b6d0493baf7c0a52f68df5a943158290e3b75878`. Publication changes no source/runtime pin and inherits no candidate qualification.

## Existing work to build on

- [PR173](https://github.com/donnabox/beads/pull/173) adds the private `internal/memoryquery` Requires operation and `ReadMemoryRequires` store adapter. Its read context is an exact-root, outward Requires-family context, not a generic incident-Link interface. It does not establish incoming traversal, native Issue roots, multi-root reads, Subject To stages or a public CLI. Follow its actual successor head and hosted qualification before using it as an integration dependency.
- [PR176](https://github.com/donnabox/beads/pull/176) extends the existing generic graph projection with Link-family selection and induced result Links. Its current acquisition is globally bounded `CurrentSnapshot`, and its text view is the flat candidate renderer. It does not implement the accepted compact relationship tree or scoped multi-root acquisition. Preserve its actual tests and qualification history while adapting its projection/presentation work.
- Trish owns selected Type validation and the default-off assessment gate. Keep definitions, recognition, validation, history and ordinary reads/cleanup; refuse gated assessment authoring. Gate qualification is separate from a prose scope statement.
- Cleo owns shared ID allocation, canonical selection/output infrastructure and remaining generic CRUD output. Donna/Melinda own graph/Memory layout. Reuse the shared writer and output contract.
- Janet owns actual integration conflict coordination and the release ledger. Authors own their implementation, tests, review and direct draft publication. No automatic main merge/release or runtime/spec pin movement follows this plan.

## Slice 1: bounded query reads for References and succession

Extract/reuse selection and traversal behavior behind an internal operation interface for the fixed CLI questions. Do not shell out to another CLI command or create another storage graph. Read actual canonical Issue/Memory/Link records through one authorized, consistent checked store view. Use the installed Type engine for refinement membership; don't rebuild a string-name ancestry table.

Add bounded current incident acquisition in both directions, including native Issue endpoints, without changing the native dependency-mode evaluator. Build on the scoped adapter's validation/SQL-side acquisition discipline; don't pretend its Requires-only reader already supplies generic incidence. Preserve canonical records, authored pins, ownership and exact-source bindings. Document supported revision routes and refuse unsupported historical traversal rather than unpin, mix views or reinterpret old records.

Traversal maintains separate reached-node and emitted-Link sets. Depth chooses the per-root population; include every matching actual Link between that root's reached Beads, not just discovery edges. Retain cross-links, reverse Links, cycles and independent parallel Links. Excluded Types cannot act as hidden bridges. A zero-match root remains a result node.

For multiple roots, compute each root's graph with independent depth/frontier accounting in the same view, then aggregate those graphs and their memberships. Do not re-induce extra edges solely across the union of independently reached node sets. A shared target does not transfer another root's obligations or depth budget. Preserve per-root reasons even when a Bead or exact witness is shared.

Acquisition work/bytes, selected nodes/Links and serialized output must be bounded. Depth-limited frontiers are different from failed acquisition/result budgets. Budget exhaustion, cancellation, unavailable context and validation failure return failure without successful partial output. An unrelated large population must not be silently loaded and then advertised as a scoped query.

## Slice 2: shared compact relationship presenter

Adapt the graph presentation at a seam that accepts validated operation results; don't manufacture Issue objects for Memory. Use actual Type labels and authored arrow direction. Expand records finitely. A repeated reference earns a cycle label only if it
closes an actual directed cycle; incoming/both-direction traversal or ancestor
indentation alone is insufficient. Use already-shown references otherwise,
including parallel same-direction edges and mixed-direction walks in a DAG. Every selected parallel/cross Link remains represented. Every requested root remains identifiable even if it was already reached from another root.

Elide local `beads/`, `links/` and `types/` display prefixes only. Exact Link-ID display controls/alternate view flag combinations are deferred. Preserve IDs in authoring confirmations and canonical JSON. Counts distinguish unique records from actual selected Links. Do not invent readiness, Issue priority or satisfaction labels for Memory, and don't change dependency-mode presentation or its compatibility quirks.

Keep the canonical complete-record output and selected graph edges distinct from embedded `owned` state. Multi-root membership and per-root frontiers require reconciliation with the common output framing before exposing them; they are internal operation data until then, not a private new JSON schema/version. Re-rendering cannot fetch additional resources behind the projection's back.

## Slice 3: Requires and Defines reads

Expose `requirements`, `required-by`, `definition`, and `defined-by` through the same bounded reads and presenter. Forward reads use the accepted depth-1000 bound; reverse defaults are one hop. Requires includes Defines, while definition selects the Defines family alone. Keep actual edge Types visible; proposed human ordering puts defining children before other requirements at each expanded parent.

Respect Type structural restrictions at shared authoring, not through query-time deletion of invalid evidence. Requires may cycle; Defines has at most one defining parent and forbids its own cycles. Closure is not satisfaction. The subject of a query remains its explicit root; the result does not establish new applicability.

## Slice 4: Subject To stages and authoring facades

`constraints` takes each original subject's explicit Subject To roots, then follows Requires/Defines. Never recursively inherit another Constraint's Subject To declaration. Result edges must respect this staged path selection; an arbitrary union-induced graph would admit unrelated Subject To Links. `subjects` initially reads direct incoming Subject To declarations targeted
to each named Constraint only. Exclude secondary declarations between reached
subjects; this operation does not repeatedly traverse Subject To. Its edge
scope is a direct incidence selection, unlike generic induced graph output. No executable implicit-applicability predicate or reverse-derived subject mode is introduced.

Expose the five accepted authoring facades through common Link create/property/metadata/guard validation. Reject duplicate property inputs. Plain References repeated creation yields independent Links. Supersede is create-only and refuses an existing outgoing successor, including an already-identical target; no hidden retarget or bespoke no-op/replay contract. An explicit `--id` conflict follows the actual shared authoring contract, not an invented idempotency guarantee. Source ownership and atomic authoring failures remain common store responsibilities.

No Assessment/Evaluation/Conformance command tree is advertised. Existing remember/recall/memories/forget keep their narrow existing behavior.

## Qualification and handoff

Turn the authored cases into meaningful pure operation tests, real store reads/writes and installed CLI scenarios. A fixture/reference integrity check on the JSON is not an execution of those queries. Test actual mixed Issue-to-Memory-to-Issue-to-Memory paths, Type refinements, induced cross-links, cycles, parallel Links, roots with no matches, per-root frontiers, one-view aggregation and staged applicability exclusion.

Test real provider acquisition limits, cancellation and malformed/read-inconsistent context; preserve atomicity and no successful partial answer on refusal. Exercise pin/revision routes actually claimed by each operation. Regression-check native dependency-mode behavior, including existing quirks, without routing it through the new generic engine.

Run the repository's required affected Bazel/nogo, formatting, docs and installed/provider qualification, then actual source review. Report missing/failed lanes honestly; exact CI and assembled-source checks are required before qualification. Use review evidence and exact source links, not authored expectation files, as execution receipts.

No broad public query-language rewrite, BDP/BTS language/IR restart, new backend, assessment evaluator, namespace migration, general Type registration CLI, cleanup of others' branches or release action is authorized by this implementation plan.
