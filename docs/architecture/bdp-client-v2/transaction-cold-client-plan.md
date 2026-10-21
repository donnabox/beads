# Transactional cold-client critical path and bounded next serving slice

Status: proposal for council/root review; no implementation authorized by this document. Supporting plan under the sole fork design.md, Autonomous profile completion T2/T3. No public profile claim, new role claim, schema change, or new protocol semantics.

## Exact inspected basis

Main `581bad0de6f5eb098b920147f6a7e09809c553ea` (PR102), tree `22e448fe0f517c0d6300b732de55494bd70c7938`. Read from clean `/Users/dbox/repos/wt/vickie/vickie-schema-property-change` HEAD `61ed530ac6a460598e0e312f5d28ec0e6fa2a33d`, whose tree equals main. Main627→581 changes only three RU evidence files; the inspected TX runtime, normative spec/schema and TX observations are unchanged. References below are repository-relative source paths at this exact tree; they are not claims about uninspected future main.

No code, tests, builds, servers, network or outward writes were performed for this plan. Latest receipt proof is the separately source-bound 0359/PR101 evidence, not a new execution at581. Its root closure records 337 focused tests and seven installed cases, nine partial positive rows/139 actions; all whole-contract claim flags remain false. Earlier installed Snapshot and SSE witnesses remain bound to their own sources.

## Recommendation

Do not publish a reduced descriptor or an Operation Directory as a substitute for a working cold client. Implement the small, settled serving prerequisite **complete minimum tuples on initial current Bead/Link collection GET/HEAD** next. Then complete the explicit remaining serving decisions/implementations below, compose canonical discovery internally, and publish discovery only when the required behavior is truthful. A directory-only patch would expose targets that work for many requests but still refuse required categories with private development gaps; that is not this milestone.

This first slice closes a real read-after-write failure: commit a mutation, take its current epoch/view/effect position, then GET an initial Bead or Link collection with that minimum. Today `transaction-http.ts:530–539` refuses the collection before the real owner read, despite the existing same-owner condition machinery. The slice does not settle retained-page/Snapshot minima, alias lookup ordering, absent-point ordering, or partial/unknown minimum tuples.

## What a cold client must discover

`docs/specs/bdp.md:909` (#scope-discovery-and-human-documentation) requires successful Scope GET with a registered service-desc Link. `:3717` (#transactional-discovery), `:3753` (#discovery-profile-membership), and `schemas/bdp-v0.schema.json:4986` ($defs.transactionalDiscovery) require a closed document with bdpVersion=0, profile=transactional, scope, scopeEpoch, authorizationView, headPosition, minimumReplayPosition, beads, links, types, operations, aliases, receipts, snapshot, changes, events. No private/partial profile or reduced Transactional descriptor exists. Optional limits cannot falsely advertise unmeasured semantics; retention.idempotency is prohibited (:3801, #transactional-limits).

The Operation Directory is a separate JSON Resource, with all twelve exact names and relative targets, not a list of transaction receipts: spec:4305 (#transactional-operation-directory-and-singletons), schema:3465 ($defs.transactionalOperationDirectory). GET/HEAD and wrong-method bodyless405/Allow behavior come from :2064 and :4600. A client follows Scope Link→descriptor→directory/collections; it must not need our hard-coded fixture target URLs.

Required canonical discovery rows are transactional.discovery.document, transactional.discovery.operation-directory, transactional.discovery.limits and transactional.discovery.no-idempotency-retention (`packages/conformance/catalog/transactional-v1.json:5–83`). Schema success alone does not attest that the referenced operations meet their contracts.

## Twelve-target readiness ledger

All twelve names are already in `transaction-http.ts`'s one authoritative operation mapping and have real dispatch. None needs a second transaction owner.

| Targets | Real current behavior | Known limitations that publication must not conceal |
|---|---|---|
| createBead, updateBead, deleteBead | Singleton Mutation Receipts, guards, shared key namespace, atomic effects, replay/restart, installed contracts | Complete profile input/error coverage remains unqualified; ordinary non-set diagnostic pointer precision has a separately recorded gap. Supplied command minimum controls use a private gap, not a documented protocol resolution. |
| createLink, updateLink, deleteLink | Canonical/local references, owned transitions/source revisions, committed Events/changes | Same generic qualifications; metadata-only Link updates can make incident Bead Event projection fail. A successful command cannot by itself prove its required Event surfaces are usable. |
| putAlias, deleteAlias | Real same-owner singleton receipt/key semantics and alias storage | Alias state is intentionally not replicated Resource history. Do not infer sequence alias-operand completeness from singleton support. |
| sequence | Ordered independent member commits, canonical @name dependencies, retained/pending receipt projection, cross-carrier keys | Resource-alias operands explicitly refused pending turn-time resolution/identity reconciliation (`transaction-http.ts:324–332`). Sets are correctly forbidden in sequence, not a missing feature. |
| updateWhere, deleteWhere | Actual valid supported Selector requests, staged selection, cardinality, noops/rollback, owned effects, receipt paging | Invalid/unsupported Selector and complexity/depth admission questions remain explicit private gaps; no protocol error/phase may be invented. Candidate enumeration before match cap is not a bounded scan claim. |
| batch | Actual ordered atomic operations, local labels, shared durable outcome, late-failure rollback | Composition inherits the above gaps and bounded materialization limits; it is not merely a loop of independent commits. |

Current `transaction-http.ts:541–548` deliberately returns a development gap at Scope, bdp.json and operations/. `read-http-routing.ts` already contains inherited canonical Scope/Link/GET/HEAD machinery. Internal compilation and canonical URL construction exist; directory publication is currently withheld, not an absent dispatch implementation.

## Required surfaces beyond mutation dispatch

Already real, under fixed all-visible owner configuration: same-owner current reads, receipt root/page GET/HEAD and durable page identities; finite/SSE Scope changefeed; finite/SSE Scope/Resource Events; durable Snapshot manifest/streams; installed compiled Type inventory. Actual standalone installed tests exist in `apps/bdptest/src/transactional-{set,snapshot,sse,event-sse}-installed.test.ts`. The cumulative TX runner does not yet qualify all their contracts.

Concrete readiness work remains:

1. **Consistency coverage:** initial collections (recommended next slice); continuation and Snapshot minimum semantics; aliases/Types/discovery minimum and required response context; absent/undisclosable points; partial/malformed/unknown tokens. Do not conflate an ambiguity with a known missing implementation. `transaction-read.ts:43–49,100–123`, `transaction-http.ts:391,529–539`, spec:5758–5808 show the current boundary. Established complete point/receipt/feed/Event minima and conditional ordering must remain intact.
2. **Sequence Resource aliases:** existing error is explicit. Reconcile member-turn resolution with admission/key identity and retained original witnesses before enabling it. No new alias API or Type lifecycle work is needed, but the previously recorded normative tension must not be silently chosen.
3. **Event completeness:** metadata-only Link updates currently trigger `TransactionEventProjectionGap` in `transaction-events.ts:257–263` for incident Bead projection. Resolve the existing Event law, then close the gap across finite/SSE/reconnect. Also retain the documented `start=now` plus Last-Event-ID combination gap; do not infer its handling from after+header precedence.
4. **Authorization/erasure domain:** fixed all-visible policy and fixed epoch do not themselves require implementing an arbitrary policy editor. CORS is conditional and can remain disabled. Optional History stays excluded. But an authority that processes an erasure must implement its mandated receipt/Snapshot/stream/replay fences (spec:5188–5238,6330ff). Audit which administrative acts this installed service actually permits before calling dormant behaviors unconditional blockers or claiming they are covered. Current empty erasures/projectionAdvance=false is only evidence for its present domain. No new administrative protocol is proposed.
5. **Discovery truth:** head/view/epoch must come from the same authoritative entry as the served document, with a genuine replay-floor policy. Current no-pruning storage may justify a genesis floor after an explicit proof, not by guessing from the last receipt, Snapshot anchor, or current head. Snapshot advertised lifetime promises constrain replay retention. Optional advertised bounds derive from actual locked configuration; retain persisted==advertised receipt page maximum and no retention.idempotency.
6. **Qualification:** actual installed tests and partial selected clauses do not discharge cumulative required-contract execution. This is a gate after implementation, not a reason to defer the implementation itself.

Questions already requiring root/product interpretation: Selector error/complexity phase; retained capture versus current-serving minimum rules where still unresolved; missing-point nondisclosure/minimum ordering; sequence alias timing/identity; metadata-only incident-Link Event projection. Root should confirm which already has an authoritative ruling before commissioning new design. No new ruling is needed for the fresh, existing all-visible collection slice below.

## Next executable slice C1: initial collection minimums

### Exact contract and seam

Spec:5758–5766 (#snapshots-and-strict-reads) requires each ordinary read to observe a transaction-consistent prefix linearizable during the request. :5770–5786 (#transactional-http-consistency) permits a minimum, requires representation at/after it or foreign-view/cursor-expired/catch-up-timeout, and forbids success with older position. :5801–5808 (#transactional-conditional-reads) puts checks before conditional shortcuts and preserves current context on304/412. Catalog rows transactional.changefeed.consistency-fields and transactional.changefeed.catch-up-timeout apply beyond the feed endpoint; this slice supplies partial witnesses, not whole-row passes.

Accept only the existing parser's complete canonical epoch/view/position tuple on an **initial**, syntactically valid current Bead/Link collection, with no continuation. Reuse the normal collection parser, Selector semantics and materializer. Empty current collection is a successful collection, not the unresolved absent-point case. Preserve the existing handling of syntactically invalid requests and every unsupported tuple/route outside this slice.

`transaction-read.ts.minimumFor` currently restricts to canonical Resource/properties. Extend its private allowlist to the bounded initial collection case. Pass existing sourceCondition into the same AuthorityReadPlane. `authority-read.ts:970–1017` captures/materializes inside one synchronous owner entry, checks sourceCondition there, and calls finish only afterwards. That ordering permits waiting attempts to fail before cursor retention. Do not move minimum evaluation to a later inspect or mutable last-position field. Demonstrate this fact rather than changing shared Read code unnecessarily.

`transaction-service.ts.readAtMinimum` already owns the finite monotonic wait; reuse it with its current budget/backoff/lifetime. `transaction-http.ts` route allowlist must match the same initial-collection predicate. Existing context projection and Read conditional response logic remain authoritative. No store schema/format, evaluator, public API or new timeout.

### Owned files and arbitration

Expected production files: `packages/server/src/transaction-read.ts`, `transaction-http.ts`; service only if a demonstrated existing seam needs a small adjustment. Tests: corresponding private Read/HTTP tests, plus a new source-private installed collection-consistency test in apps/bdptest, using existing installed launch/metadata helpers. Narrow progress doc update allowed only after code approval. No catalog, matrix, fixture, runtime settings, RU evidence or conformance runner changes. Shared AuthorityReadPlane, store, installed configuration and runtime transport remain read-only unless root approves an identified necessity. Root must serialize any future discovery owner access with this lane; do not opportunistically patch sibling files.

### Done-when and discriminators

1. Actual installed canonical Host/physical port0: commit through a real operation; initial Bead and Link collection GET using its current epoch/view/position succeeds with expected literal records, current revision and all three response headers.
2. Successful empty collection at genesis/known current minimum is supported without fabricating a404.
3. Behind minimum waits; a real later commit can satisfy it. Each retry re-enters owner after releasing prior entry. Observed final records and context come from the satisfying entry, never stale attempt data.
4. No progress reaches the existing finite deadline and returns normative503 catch-up-timeout, exact schema/code/retry and final attempt context. Wrong view409 and epoch410 use actual observation; unauthenticated request stays nondisclosing with no context leak.
5. GET and HEAD apply identical minimum and conditional ordering. Matching/absent validator conditions cannot bypass a behind or incompatible minimum; successful304/412 carry current context; HEAD has no body.
6. An initial response that creates an ordinary continuation retains rows/revisions/context from the successful capture. Waiting/refused attempts do not consume retention slots. Following that continuation **without minimum controls** preserves inherited behavior; continuation-with-minimum remains explicitly outside scope.
7. Accepted Selector and pagination-limit behavior remains unchanged. Unsupported tuples/routes do not accidentally become successful because the collection allowlist widened.
8. Direct same-owner tests prove no owner callback escapes, reentrancy/fencing unchanged and no SQLite await. Existing point/receipt/feed/Event/snapshot and RU Read regressions pass as appropriate.
9. Installed negative control that removes the collection source condition must fail on final state/context or premature success; a fake header-only implementation must not pass. Use controlled actual owner progression for waiting test, not timing-only races.
10. Fresh clean-source packed proof records source/tree, payload and runtime configuration identities, exact targets/headers/body observations and child cleanup. Keep existing request/cohort caps; no broad timing expansion, no new partial profile.

Feasibility: at most a few small Beads/Links, no 32-result boundary or new matrix actions. Use current collection page size and bounds; exact test/action/time ledger should be measured in implementation proposal before packed execution. Repeated materialization while behind costs work; this proposal does not claim bounded total scan by page size, nor introduce notification/cache infrastructure. If existing seam cannot prove pre-retention condition ordering, stop and request narrow shared-file scope rather than accept leaked cursor capacity.

## From C1 to a genuinely usable cold client

**C2, in parallel only where disjoint:** resolve/rule the named interpretation gaps and author bounded required serving repairs. This is the critical path, not additional isolated selected-clause rows.

**C3, internal assembly first:** use one owner entry to assemble immutable canonical discovery values and exact twelve-target directory from the actual route table. Parse with existing protocol/schema authority; do not construct a shadow descriptor inventory. Owner epoch/view/head/floor and configuration must be captured together. Positive/negative assembly tests can run without publishing a profile. Test missing member, invented target, stale context, wrong profile, forbidden retention.idempotency and mismatched scope. The internal shape is the normative shape, not a private wire protocol. Do not serve this descriptor while known mandatory behavior remains unsupported.

**C4, actual cold private candidate once serving readiness is established:** Scope GET service-desc Link→descriptor→directory→all twelve operations and collections, receipts/paging, Snapshot→feed and Event sources. The installed client receives only Scope+credentials and follows actual URLs. No direct fixture target lookup in the cold oracle; independently compare returned targets against subsequent observed successful routes and schema. Commit then list at minimum and verify retained receipt/replay/reopen. Scope/descriptor/directory GET/HEAD/method/auth/cache/context behavior must be tested, including current position advancing between independent requests (each response coherent, not required to have identical positions). This may need a separately reviewed installed harness seam to record discovered target authority safely, never unchecked arbitrary URL execution.

**C5, evidence/public admission:** run fresh canonical cumulative requirements against the exact installed cold candidate. `packages/conformance/src/profile-contract-selection.ts` loads the complete catalogs; ordinary runner and source-private branded TX runner remain distinct (`runner.ts:320–349`). `packages/server/src/index.ts:288–306` currently rejects any public profile other than Read. Change that admission only under a separate reviewed evidence-binding plan after required-contract completion. The current private runner already executes candidate TX without public admission, so no circular gate requires opening the public server to qualify it. Conversely, a successful schema parse or a handful of cold operations is not sufficient to open admission. Never relabel historical Read or partial TX evidence.

A candidate assembly test is not advertisement; a served document carrying profile=transactional is an assertion about the service even if the listener is called private. Therefore C3 tests can precede readiness, C4 wire publication follows truthful required behavior, and C5 public deployment admission follows fresh evidence. If product wants a publicly usable subset earlier, that is a separate product/protocol decision, not permission to invent a private profile in this implementation.

## Limits and next decision

No new product decision blocks C1. Root/council should approve its small production scope and actual installed proof. The complete cold milestone remains blocked by the explicitly enumerated required-contract decisions and repairs, not by the absence of an Operation Directory JSON helper. No estimates of complete-profile readiness or row-count completion are made here.
