# Beads graph delivery plan — revised proposal v3

**2026-09-23. Status: implementation plan for review; bounded implementation start authorized.** This replaces v2 as the active plan proposal. It incorporates the revised CLI proposal and Donna’s ten inline comments. It does not convert open CLI questions into approvals, claim a new council review, or certify previously completed code.

The [CLI proposal has a separate upstream Proposal issue, #6703](https://github.com/gastownhall/beads/issues/6703). Its initial review source is commit `310bc2d6d071ae4574efb5c1d5d9ebbec7b099bd`. Review CLI shapes and compatibility there; review feasibility, sequencing, reuse, estimates and acceptance here. Implementation does not decide a disputed public contract by getting code there first.

## 1. Recommendation and execution authority

**Simplify substantially, then deliver narrow vertical slices that converge on the complete feature.** Preserve the full agreed destination: Issues as generic Beads, Dependencies as Links between Issue-typed Beads, canonical Memory with its required shared History, familiar Issue workflows, and usable CLI/BDP surfaces. Do not require the whole replication or arbitrary-Type roadmap before the first installed CLI demonstration. Do not sacrifice identity, concurrency, integrity or workflow correctness to produce a convincing transcript.

Donna has now explicitly requested that implementation begin while the two proposals go out for review. Start the previously approved **three-working-day first-checkpoint attempt** with reversible work in an isolated branch. This authorizes work toward that bounded checkpoint, not merging unresolved semantics, deploying irreplaceable data, or an unlimited complete-A budget. The first implementation edit was made on September 23 at 07:01 PDT (14:01 UTC), on the isolated `codex/beads-graph-start-20260923` branch from the qualified base below. That starts the three-working-day attempt; documentation preparation does not consume an invented earlier start date. The initial edit adds workspace-format discovery/admission and does not deliver C0.

Initial implementation base: qualified `99f7f7d6f81f66447556f9ed53e358b508b4ce37`. Preserve the existing graph/P0 ancestry and useful tests. Prepared successor `b610946f6d72920894f718c289e4c3cdc259fa78` is an optional reuse candidate; its complete integration gates have not run. The implementation plan is maintained and reviewed only in the `donnabox/beads` fork. Its review branch starts from the fork’s main and contains only this document. The CLI proposal is reviewed independently in the upstream issue; its full Markdown source stays in the fork. The old design PR gastownhall/beads#6154 and P0 PR gastownhall/beads#6422 remain intact; this proposal requests targeted amendments, not silent replacement or closure.

## 2. What complete A means

- One authoritative representation in an opted-in workspace. Issues are Issue-typed Beads; Dependencies are Links with Issue-domain semantics and Issue endpoints. Generic and familiar commands use the same authority, domain rules and transaction effects. Specialized backing tables are allowed; permanent dual authoritative Issue/graph state is not.
- Existing workspaces remain legacy until explicit migration. Preserve all documented Issue/Dependency workflows and CLI outputs by complete A, including batches, labels, event history, ephemeral/wisp/no-history behavior and maintenance/integration writers. Preview refusals are allowed, but an unresolved required workflow prevents declaring A complete. Legacy-mode maintenance is part of the cost; no unapproved sunset is assumed.
- CLI first, BDP next. Both embedded Dolt and an ordinary released shared Dolt server; macOS and Linux. One authority-owning workspace with multiple concurrent CLI processes. Windows, arbitrary engines, independent offline multi-writer reconciliation and full BDP Transactional are outside A.
- Canonical Memory follows Chris’s gastownhall/beads#5877 API plus Donna’s amendments: `remember` writes, `recall` reads complete selected content, `memories` discovers, `forget` removes current state while preserving History. These are adapters over the same graph. No substitute Note feature and no second canonical key/value store. Canonical keyed-memory conversion is the preferred A scope; price any blocker explicitly.
- Complete A includes Memory-required shared History: stable immutable version addresses, historical recall/as-of behavior or its typed refusal, comparison, restoration and change/removal observations. This supersedes v1’s blanket History deferral. Broader generic History features that Memory does not require remain later work. An early partial preview is not full gastownhall/beads#5877 conformance.
- Stable, nonreusable canonical identity within a Scope; reopening/restoring preserves identity. Mutable aliases are locators, not a third key or part of a content version. Cross-Scope references do not imply enforceable remote integrity. Unresolved external blocking Dependencies never count as satisfied.
- Supported SQL reads remain possible. Arbitrary direct SQL mutation compatibility is excluded; CLI/domain/BDP and supported adapters are controlled mutation entry points. Privileged engine administration remains outside that API, with explicit continuity/recovery obligations.
- Fresh-machine recovery of a documented backup must prove Scope continuity or require a new Scope/refuse. Old/incomplete backups must not reopen allocation for identity reuse. Migration retains the source database: rollback before graph writes, forward recovery afterward; no promise of lossless downgrade after new writes.

## 3. Corrections incorporated from the CLI review

| CLI decision or correction | Implementation consequence | Status / latest gate |
|---|---|---|
| Initialize graph storage once; no routine mode switching | Persist and validate workspace format before selecting any writer. CLI flags cannot redirect a graph workspace into a legacy authority | Storage persistence is agreed; revised `--graph-mode` assertion grammar is proposed and must be settled before public CLI freeze |
| Preserve old `--type`; propose generic `--resource-type` | Parser/dispatch tests must demonstrate old Issue classification and Dependency spellings unchanged; no per-mode reinterpretation | Open CLI shape; do not publish it as settled because an internal branch contains a parser |
| Bead Type is immutable too; Task/Bug nominal types desired | Single immutable Issue Type plus mutable `issue_type` is only a provisional early-slice model, not a final ruling | Major model decision before descriptor freeze, migration cutover and complete-A compatibility claims |
| Common metadata on Beads and Links | Top-level metadata is the CLI recommendation, not existing BDP wire. Nested placement would change schema/patch mapping | Placement decision before durable public metadata ABI or remote writes; optional metadata need not block a minimal title/body checkpoint |
| Owned Links and pins already exist in public BDP | Adopt `ownsOutgoing`, wildcard ownership, inline `ownedLinks`, existing Reference union and source-version effects | Protocol law settled at inspected public pins; Beads realization and evidence remain work |
| Link-property update is first-class | `update links/ID` needs property/metadata validation, Link guard, owning-source effects and output. Endpoints, Type and pin remain immutable | Essential M1 mutation slice; not postponed behind arbitrary custom Types |
| Multiple generic Links may share endpoints | Stable Link IDs, ID unlink and ambiguous-pair refusal; distinguish intentional duplicates from request retries | Essential M1 semantics; reconcile legacy Dependency pair uniqueness before adoption |
| Generic operations suffice for generic CLI | Use generic records and CLI summary rendering; no mandatory new remote Memory-summary service | Remove that previously invented prerequisite; measure remote search/projection needs separately |
| Optional atomic delete-all-incident-Links | If accepted, one authority transaction including surviving owners’ revisions; no client unlink/delete loop | Open convenience/policy; not on first-checkpoint critical path. Remote Read+Update alone cannot emulate atomicity |
| Explicit behavior-change inventory | Turn CLI §4.1/§14 into a leaf-command/write-path census with source references and validation | Required before compatibility completion; Memory guard/default-output changes remain labeled breaking proposals |
| Current revision, local ordinal and portable version differ | Reuse Jim’s writer, but never expose its local ordinal as stable historical identity | Exact identity/address and continuity tests before Memory compatibility advertisement |
| Corrected BDP source | Public gastownhall/bdp#17 and gastownhall/bdp#30 and adopted Read contracts are authority; old internal checkout `650f70aa` is not a newer public baseline | Remove false “ownership/pins not specified” gates; retain integration/capability gates |

The editable CLI draft is not copied into this plan. Its independently reviewed public contract controls user-visible behavior after decisions are accepted. Cost or code sunk into one grammar is not a reason to reject a better reviewed shape.

## 4. Existing evidence and its limits

At the starting pins, existing Issue/Dependency commands and legacy keyed Memory exist. Static inspection found no production root `--graph-mode`, complete graph bootstrap/catalog/Scope issuance, storage accessor/adapter chain or graph CLI routing. Therefore adding the flag alone cannot make normal `init → create → show` work.

Previously inspected evidence, not rerun for this plan:

| Evidence | What it supports | What it does not support |
|---|---|---|
| Parent full-suite receipt, exit 0, approximately 43m22s, 105 passing packages; reported 7,191 top-level passes and 1,622 skips; raw stdout hash checked | Regression evidence for that exact parent and environment | Generic CLI usability, new changes, all storage topologies or unskipped coverage |
| Embedded fixture receipt: 258 plans including 140 incident plans; separate seed/read processes had the same digest | Private reader/fixture behavior and persistence at that boundary | Normal initialization, production authority, stable cross-request snapshots |
| Managed fixture: two generations observed three Beads/three Links and matching digest | Candidate managed read composition | Normal shared-server deployment; prepared EXPLAIN was omitted |
| b610 preparation: component tests/lint pass, optional depguard skip; integration gates not executed | Reuse candidate for private read-attempt composition | Qualification of the whole successor |
| P0 Dolt 2.1.8 / embedded driver 2.2.0 experiments | Historical ordinary-write trigger/session observations | Merge/pull trigger behavior, current distribution or cross-process admission proof |
| E1 merge exclusion/drain experiment | Named risk and proposed observation | Engine-runtime qualification; still absent |
| Jim gastownhall/beads#6661/#6358 inspected at `64becbc335cd754efa1f68effcbd5ca3f33873db`; runbook pin `32d4f966e…` | Real `versions` command, store config wiring, same-transaction Issue snapshots, mutation census | Generic Memory, complete CAS/portable History, concurrent writer safety or integrated graph feature |

Jim’s inventory reports 21 of 34 paths versioned, 27 seam calls and 46 must-mint entry points: different denominators, not one completeness percentage. His direct/UOW paths can accept different version counts; unchanged import can mint; lifecycle gaps remain. The writer uses `MAX(revision)+1`, has a one-writer restriction, and truncates timestamps to seconds. The runbook’s precise round-up explanation does not match the inspected implementation. Reported demonstrations and author test results are not independently replayed receipts.

The old plan’s first council covered v1. The API, Scope amendments, required Memory History and September 23 CLI changes arrived later; do not call that council approval of v3. Team review is requested now. No new automated review council is presumed by this publication.

## 5. Start now versus wait for a ruling

| Work | Can begin now? | Boundary |
|---|---|---|
| Read-only workspace discovery, format admission, wrong-workspace refusal, isolated build/install path | Yes | Preserve old behavior and reject unsupported graph routes before migration/effects; no false successful init |
| Counted storage/layout comparison and actual command/writer census | Yes | Concrete files/queries/call sites, not another generic framework project |
| Ordinary-engine transaction candidate and bounded concurrency/cancellation observation | Yes | One owning workspace; no production database experiments or new engine fork |
| Production bootstrap/read/write plumbing behind internal interfaces | Yes, reversible | No public metadata ABI, nominal subtype freeze or mandatory command spelling settled by implementation |
| Minimal experimental Memory title/body creation with retained-state obligations | After viable bootstrap/transaction choice | Experimental disposable Scope, explicit incomplete capability; cannot stand in for complete Memory |
| Nominal Task/Bug descriptor identity, retyping, common metadata wire | Wait for the corresponding ruling | No permanent public descriptor/contracts that make the choice irreversible |
| Guarded Memory compatibility writes and changed default output | Prepare adapters/tests; freeze behavior after review | Existing scripts’ breaking changes must be explicit |
| Linked-target deletion policy and `--unlink-incident` | Prepare atomicity design only | No implemented cascade passed off as approval; no spin/retry deletion loop |
| Migration cutover, graph sync/merge/restore, advertised remote writes | Not yet | Require their identity, route, compatibility and recovery gates |

The first bounded code slice is admission/discovery and its real call-path guard. It is useful only as part of the immediate bootstrap/create/show work; a marker, parser, interface or passing refusal test is **not** the accepted checkpoint. If foundation work consumes the checkpoint without the command working, report a missed commitment and reassess.

## 6. Reuse and physical-layout decision

Keep existing Issue domain behavior, projections, tests, useful graphops laws/read bodies and bdpwire contracts. Keep source changes behind the storage/driver boundary. No Beads-side flock, engine-introspection workaround or bespoke retry/recovery subsystem merely to get a demo working. If a needed operation exceeds the driver interface, make that limitation explicit and price the smallest legitimate extension.

Compare these two layouts before choosing durable schema:

| Candidate | Potential saving | Required counted evidence |
|---|---|---|
| Universal generic payload storage | Uniform enumeration/Type validation for custom Beads/Links | Actual changes to Issue ready/list/search, labels, batches, dependency effects, UOW/direct paths, indexing, migration, SQL read projections and History |
| Specialized Issue/Dependency backing under one graph identity/revision/transaction model, generic storage for other Types | Reuse existing Issue query/domain paths and Jim’s recording seam | Exact authoritative owner per field/Link, complete generic enumeration and mutation, identity allocation, revision coupling, canonical Link multiedges, foreign-key and retention consequences |

**Starting hypothesis, not a selected schema:** specialized backing may be the smaller change because preserving existing Issue workflows is mandatory. It qualifies only if generic and domain mutations really share authoritative state and effects. An eventually consistent observer or reconstructed second graph is not that design.

Concrete reuse constraints at the qualified pin: `internal/storage/dolt/transaction.go` normally splits durable and ignored-table writes across SQL transactions and commits them in sequence; its events-journal mode already collapses them into one transaction. Evaluate that existing path before introducing another runner, and prove its failure behavior rather than treating the `RunInTransaction` name as an atomicity guarantee. Embedded SQL commit and later Dolt history commit are also distinct; durable content versions must be recorded inside the authoritative SQL transaction, with later-history failures classified honestly.

The existing Dependency schema uses a deterministic ID and unique source/target keys without Type in the identity (`0050_dependencies_deterministic_id.up.sql`). Generic multiedges therefore cannot simply reuse that table unchanged. Count an explicit specialization/identity mapping against a canonical Link-table alternative, including mixed Issue/Memory enumeration; do not double-write two authoritative Links. These are inspected source constraints, not newly executed backend proofs.

The comparison’s required output is one table of affected source files/interfaces and query families, an allocation/transaction sketch for both backends, an estimated migration delta, and the selected smaller coherent layout with rejected alternatives. Give it a half-day investigation budget inside the checkpoint. Count the actual surfaces; this plan does not fabricate counts that have not been measured.

Reuse Jim’s writer/census before writing a second Issue-history system. His snapshots are Issue-specific, so extension cost is explicit. The qualified starting tree also contains a pinned MySQL client fork; record whether the selected path needs it. An ordinary released server does not by itself prove a stock-client distribution. A required maintained **engine** fork triggers architectural reassessment; inherited driver patches must be justified, qualified and counted, not silently shipped.

## 7. Delivery order and work packages

### C0 — Three-day installed checkpoint, part of M1

The complete checkpoint is a real installed CLI on both embedded and ordinary shared-server Dolt:

```sh
bd init --graph-mode link --scope-url <operator-owned-scope-url>
bd remember "First durable graph record." --id beads/plan --title Plan
bd show beads/plan
# Exit the process, reopen the same workspace:
bd show beads/plan
```

The init and command spellings are the CLI proposal and may change through its independent review. The invariant is normal initialization, one non-Issue Bead at a stable canonical identity, complete exact read, and persistence after a new process. Also accept the original explicit `bd --graph-mode link show beads/plan` assertion spelling if retained. No manual schema, seed fixture, mock storage or background-server magic may be required.

C0 also needs an accepted bounded ordinary-engine authority/transaction observation: allocation, current revision, required retained state and supported domain effects commit atomically; two CLI processes serialize or conflict; failed/cancelled/uncertain operations do not permit an unaccounted writer to overlap. Wrong workspace identity and missing metadata in an existing-store operation refuse before migration or manufacturing a new empty database. Fresh init intentionally provisioning its requested database is distinct.

Start installation at the beginning. Record source/binary, backend versions, build dependencies, command outputs/exit statuses, wall time and engineering effort. All four OS/backend combinations are required by M1; lack of a Linux host is a visible qualification gap, not a waived requirement. C0 is not a full Memory conformance claim.

### M1 — Mixed graph CLI

| Package | Deliverable | Observable exit |
|---|---|---|
| W1: workspace admission and bootstrap | Persisted format, supported-client fence, catalog, Scope/allocation state, real storage role consumption and exact read/write route | C0 transcript works; unsupported routes refuse before effects; bootstrap interruption leaves a recoverable state, never false success |
| W2: mixed domain mutation | Two Issues and two canonical Memory Beads, generic and familiar Issue writes, local blocking Dependencies | Both API directions see the same identities/state; close blocker changes ready; Memory cannot be a workflow endpoint |
| W3: informational Link lifecycle | Issue→Memory and Memory→Memory, Link IDs, property update, ID unlink, multi-Link ambiguity | Link properties change without retargeting; owned-source revision/history changes atomically, target unchanged; retry versus second Link distinguished |
| W4: failure/concurrency and distribution | Actual process failures/guards, no-op behavior, nonreuse, installed artifact on both OS/backend combinations | Reproducible evidence for admitted paths and named unsupported paths; no hidden engine patch |

Canonical Memory recording must preserve the complete state that its accepted versions cover, including owned Links. An optional Issue history switch cannot turn it off. A partial preview can defer rich History read UX, but cannot record knowingly incomplete “versions” and retrofit missing history later. A semantic no-op does not mint; each actual owned-Link mutation follows the public source-version law. Broader restore/deletion semantics do not have to be invented to accept the initial live create/read/update slice.

The mixed transcript must exercise a descriptor-admitted Link **property** update as well as common metadata after its placement is selected. Do not demonstrate only metadata and call Link properties finished. Use a reviewed built-in property contract or a deliberately scoped experimental descriptor; arbitrary custom Type administration need not be completed first.

### Early adoption-feasibility gate — before M2 breadth

Begin this during M1, before committing to collections/federation breadth. Produce (1) representative read-only migration preflight with named exceptions, (2) external blocking-Dependency representation consistent with local Issue constraints and opaque external references, and (3) backup/continuity classification identifying evidence that must survive outside the database.

This pulls deployability forward as Donna requested. It is a feasibility result, not a requirement to finish migration/recovery before C0. A blocker that makes existing workspace adoption implausible changes the plan immediately, even if the four-Bead demo works.

### M2 — BDP Read of the same records

W5 exposes exact reads, installed Types, generic collections and incident Links through the selected public Read contract. Use the existing owned-Link/pin/History definitions; do not redesign missing contracts that already exist. Prove one interoperable client, typed absence/refusals and honest continuation invalidation or stable snapshot lifetime.

Generic CLI summaries may render generic records. Efficient remote Memory search/projection is measured separately, not an invented universal API prerequisite. A private `afterPath` is not a stable cross-request continuation. Local embedded snapshot consistency and shared-server remote cursor lifetime are separate observations. Embedded support does not add an embedded HTTP server.

### M3 — Deployable adoption and Memory-required History

W6 closes the complete Issue/Dependency command and writer census, implements explicit migration with old-writer exclusion and retained source database, admits single-authority synchronization where qualified, and proves backup recovery on a fresh machine.

W7 completes canonical keyed Memory conversion and its required History surfaces: exact version addressing, historical recall, bounded discovery with citations, comparison, guarded restoration, changes/removals, retention/erasure disclosure and exposure defaults. Price/address Jim’s direct/UOW and lifecycle differences. This work starts with M1’s recording prerequisites and proceeds alongside adoption; it is not first discovered at the end of M3.

Historical source snapshots must contain the whole owned set. Restoring one cannot silently reallocate a deleted Link identity or retarget an immutable Link. Its restoration/mapping rule is a concrete decision gate. Deletion preserves retained versions; erasure is distinct. Open local-integrity policy must be settled before admitting affected forget/delete workflows. If optional incident cleanup is accepted, it requires one authoritative transaction and all surviving-source effects; no separate client calls advertised as atomic.

Legacy Memory conversion preserves empty bodies and exact keys under the canonical-path/alias namespace contract. No invented history, actor or mutable key field. `memories --json` retains its body-map compatibility including a real `schema_version` entry; canonical records use a separate representation. `prime` does not inject Memory bodies. Automatic export adds/refreshes no Memory records; complete backup includes them.

Before adoption is accepted, someone other than the implementer installs and completes the documented workflow without live coaching. Donna reports Steph is working on this validation; no completion is assumed and no new assignment is made by this plan.

If multi-clone workflows cannot submit mutations to the authority until M4, M3 adoption is explicitly restricted to qualified solo/single-authority workflows. Record that limit and remaining cost; do not let replicas originate independent graph changes as a shortcut. Complete A still requires the agreed compatible workflow coverage.

### M4 — BDP Read+Update

W8 exposes the reviewed single-Resource mutation contract and required Memory adapters over the **same** domain transaction paths. Qualify current revisions, no-op rules, immutable members, local endpoint validation, external opaque references, owned-source results, conditional writes and uncertain outcomes.

CLI-only extra guards or metadata cannot be sent as invented base BDP fields. Adopt the selected protocol/profile amendments or refuse unsupported operations explicitly. A remote Read+Update-only service cannot implement `--unlink-incident` by a loop; that convenience is either supported by an advertised atomic equivalent/Transactional path or unavailable on that route. Full BDP Transactional remains outside A.

### M5 — Custom Types and complete-A qualification

W9 provides installed descriptor closure validation, known Type identity, custom non-Issue Beads/Links, endpoint conformance and open-vocabulary ownership limits. Unknown Type creation refuses. Identical descriptor reinstall is a no-op; changed contract under the same Type ID conflicts.

W10 performs final command-level, installed-artifact and independent workflow qualification across the required matrix. Explicitly resolve Issue nominal typing/reclassification before claiming complete compatibility. Complete A requires every preceding required exit, including Memory History and legacy workflows, not merely the custom-Type demonstration. Broader History, Transactions, independent offline writers and Windows remain a separately priced roadmap, not abandoned functionality disguised as complete A.

## 8. Decision deadlines and architectural amendments

| Decision / amendment | Needed before | Default while open |
|---|---|---|
| Init-bound format and generic command spelling | Publishing a stable CLI/help contract | Work in isolated preview code; do not claim accepted grammar |
| Nominal Task/Bug versus mutable property or explicit reclassification | Public Issue descriptor freeze and migration cutover | Provisional Issue property only; no nominal-type claim or silent break of `update --type` |
| Top-level versus nested common metadata | Durable public metadata ABI, migration and remote write exposure | Avoid unnecessary metadata dependency in C0; examples remain explicitly proposed |
| Local deletion integrity: live/historical, latest/pinned, incoming/outgoing | Affected delete/forget/restore acceptance | Restrictive refusal; no implicit cascade or revision rewriting |
| Optional atomic incident cleanup | Implementing that flag | Defer it; keep default restrictive deletion |
| New generic/Memory guard requirements and legacy-key upsert changes | CLI compatibility freeze | Label breaking proposals, preserve existing Issue output/guards |
| Issue ownership and restored owned-Link identity | History schema/restore freeze | Preserve recorded Issue/Dependency semantics; refuse unsupported restoration rather than partial success |
| External blocking Dependencies | Migration acceptance | Preserve data and report unresolved; never silently satisfy/drop/retype it |
| Authority and recovery continuity | Admitting graph writes and later lifecycle routes respectively | C0 supports only the proven initial topology; replication/takeover/restore remain refused |

Amend the conflicting passages in `BDP_BEAD_GRAPH_PLAN.md`, `BDP_GRAPH_CLI_AND_STORAGE_SPEC.md`, `BDP_GRAPH_ARCHITECTURE.md` and the replication ADR after review accepts the new direction. Important deltas: persistent workspace authority; one graph rather than separate authoritative planes; only necessary storage roles for consumed operations; a mutation role; Issue-specific commands still meaningful; complete init rather than invisible extra mint; separate content restore versus Scope recovery; and later replication proof rather than blocking every fresh disposable workspace on every topology.

Keep the project charter’s storage boundary. The generic/Memory expansion is explicitly requested by Donna and gastownhall/beads#5877; required charter changes should be reviewed rather than hidden. CLI proposals such as `scope recover` also need scrutiny against the repository’s preference for `doctor --fix`; publication is a request for that feedback, not a waiver.

## 9. Validation and evidence discipline

Use existing receipts first. For each new package, name the user-visible failure risk and smallest meaningful test boundary. Run focused tests while coding; use real initialization/processes/backend transactions where mocks cannot establish the property. Run the applicable repository lint/docs checks and required final qualification before presenting code as ready. Do not rerun a 43-minute parent suite merely to decorate a Markdown review.

Evidence rows must say: exact source/binary/backend, command, observed output/exit, accepted/refused scope, skip/N/A reason, and remaining uncertainty. Commit a reproducible transcript or link a durable receipt. A interface/test count, a green fixture or a reviewer’s approval is not a runtime feature. Full-suite counts with thousands of skips are not “everything tested.”

Workload progression: authored four-Bead mixed graph first; frozen public Beads GitHub issues for realistic content/state; existing 10K/20K generators for capacity; deterministic mutation/concurrency sequences for known outcomes. Label synthetic edges; GitHub mentions are informational, not inferred blockers. Preserve source/date/license. Jim’s reported 5,800-Issue clone is a candidate if a permitted sanitized fixture becomes available; it is neither obtained nor a prerequisite to C0. This is surrogate evidence, not production adoption.

## 10. Estimate and economic decision

Time to a usable feature is primary; agent spend and operator attention also count. The following is an **initial planning hypothesis for review**, in active engineering days including the named package’s focused validation. It is not a measured velocity forecast, a commitment to spend, or a claim that parallel agents make elapsed time equal the sum divided by headcount.

| Work | Preliminary range | Main uncertainty |
|---|---:|---|
| C0 attempt | 1–3 days, hard stop at 3 | Normal install/bootstrap and ordinary-engine transaction viability |
| Remaining M1, W2–W4 | 4–8 days | Issue adapter fan-out, owned Links and process concurrency |
| Early adoption-feasibility work | 1–3 days | Existing data exceptions and external Dependency/recovery model |
| M2 Read, W5 | 3–6 days | Snapshot/continuation and serving-route realization |
| M3 compatibility/migration/recovery, W6 | 7–14 days | Full writer census, maintenance, old-client fence, fresh-machine continuity |
| Memory adapters and required History, W7 | 6–12 days | Jim reuse, portable identity, lifecycle, feed/restore coverage |
| M4 Read+Update, W8 | 3–6 days | Protocol amendment/provider mapping and remote uncertain outcomes |
| M5 custom Types, W9 | 2–4 days | Descriptor closure, conformance and ownership limits |
| Cross-matrix independent qualification, W10 | 2–4 days | Linux/backend availability, install reproducibility and discovered regressions |
| **Subtotal** | **29–60 active days** | Hypothesis before counted layout/census; work packages overlap in calendar time |
| Explicit rework/review reserve | **7–15 days** | Open CLI/model decisions and integration conflicts |
| **Initial complete-A planning range** | **36–75 active days** | Low confidence; re-estimate after C0; not an approved budget |

This range is large enough that success at C0 is necessary but insufficient evidence of economy. Reviewers should challenge both scope and estimates. A stable-identity retyping design, an external authority service or a maintained engine fork could move the range materially; these are not silently funded by the reserve. No numerical latency/memory ceiling or total approved spend ceiling exists yet.

At C0 reassessment, replace this hypothesis with measured effort, the counted representation/writer deltas, top unresolved risks and a narrower remaining range. Agree a cumulative ceiling **before** committing beyond the first bounded attempt. Do not convert uncertainty into a new prerequisite project with no stop date.

## 11. Daily checkpoints and viability rules

Each working day, report yesterday’s specific commitment, demonstrated CLI/BDP behavior and reproducible evidence, risk resolved or falsified, prerequisites added/removed, active effort and spend, next concrete commitment, and effect on the complete-A estimate. Documentation and admission plumbing are progress, but they do not count as delivering create/show. No unattended daily automation is installed by this document.

Initial commitment: finish the representation/admission decision, produce a real install/build path, and connect bootstrap to the first create/show path. By the end of the three-day attempt, accept C0 only with its full command and authority evidence; otherwise report partial/failure and replan. Do not keep extending the same attempt under a new label.

Two consecutive missed daily commitments trigger replan before further implementation. One critical architectural falsification can trigger it immediately. Replan must identify the false assumption, remove or replace machinery, preserve the complete destination, and account for revised cost. Do not make commitments trivial just to avoid the rule.

Judge viability on three axes: visible user value; decreasing deployment/authority/adoption risk; and convergence of remaining effort within the agreed ceiling. Stop or change approach if stock-engine support fails, lawful unification cannot preserve required workflows, or remaining cost is unacceptable—even if old work was expensive or a narrow demo looks good.

## 12. Review and publication protocol

Keep this implementation plan and its review in the `donnabox/beads` fork. Do not propose or post the plan upstream. Publish the CLI contract as an upstream Proposal issue linked from Memory proposal gastownhall/beads#5877, with generic graph/BDP context and no implementation schedule or cost rationale. Open the separate fork plan review after the CLI issue; link to the CLI review one way so plan reviewers know the target. Do not cross-post plan arguments into the CLI discussion or claim implementation start resolves CLI questions.

Ask plan reviewers to identify missing mandatory work, overgeneralization, unsafe ordering, underpriced compatibility/History, insufficient acceptance evidence, and conflicts with existing contributor work. Request severity, affected package/section, reason and smallest correction. Keep dissenting model choices visible; agreement is not runtime evidence.

Fold accepted feedback into the exact affected document, increment its review revision and summarize substantive changes in its own discussion. Do not close gastownhall/beads#6154/#6422, merge plans, request team members by guessed handle, or launch a council automatically. The implementation start remains bounded and reversible while feedback is collected.

## 13. CLI-to-delivery coverage

This mapping prevents small but essential parts of the CLI contract from disappearing between the vertical demonstrations. A listed package owns the acceptance evidence, not just a parser. Open semantics retain the decision gates above.

| CLI proposal surface | Delivery package and required evidence |
|---|---|
| §3 selection, canonical selectors, explicit mode assertions; §15.1 init | W1: persisted format, unambiguous selection and fail-before-effects mismatches; normal initialization/reopen. Namespace rules must precede allocation. |
| §§4–5 create/show/update/edit, JSON/file/stdin input, previews | W2–W4: the editor uses the same guarded update path; reject immutable members and double stdin consumption; preview does not allocate or reserve. Existing Issue parser/output meanings remain intact. |
| §6 link/unlink, property update, multiedges and ownership | W3–W4: complete create/update/unlink by identity, source/target guard behavior, owned-source History and retry-versus-new-intent evidence. |
| §7 local integrity, external references and version pins | W2–W4 for live mixed graph and pinned references; adoption gate/W6 for external Dependencies; W7 for retained historical behavior. No implied external liveness or implicit pin rewriting. |
| §8 uncertainty; §4 request IDs and `request show` | W4: atomically associate supported mutation receipts with effects, replay without duplication, distinguish conflicting token reuse and unknown outcomes, and expose honest request-status capability. This is part of retry safety, not a second workflow engine. |
| §9 delete and optional incident cleanup | W6–W7 after policy ruling: guards, complete authorized commit set, surviving owned-source effects and retained/deleted/erased distinctions. Optional atomic cleanup stays deferred if not accepted; no silent loop fallback. |
| §10 Memory authoring/recall/discovery | W2 starts canonical live content; W5 adds bounded generic discovery; W7 closes keys, search, citations and History-dependent behavior. Preserve empty bodies and complete explicit recall. |
| §11 versions/compare/restore/changes | Recording begins W2; W7 completes required Memory History and guarded restoration. Keep Issue local ordinals distinct from revisions and portable version addresses. Broader generic History outside Memory remains separately scoped. |
| §12.1 lists, incident collections, bounded generic graph view | W3 supplies initial CLI reads; W5 completes paging/filters/traversal before exposing them through BDP. Test cycles, parallel Links, labeled frontier and invalidated continuations. Historical completeness requires W7 capability; no present-day incoming set relabeled as historical. |
| §12.2 built-ins, custom descriptor installation and aliases | W1 installs built-ins. W5 adds conditional, live in-Scope Bead alias operations without chains or content-version changes; W7 verifies legacy-key conversion. W9 closes custom Type installation and atomic pinned descriptor closure. |
| §12.3 export/import/backup | W6–W7: exposure defaults, retained manual Memory entries, connected-reference omissions/refusals and complete backup. Connected interchange stays explicitly capability-gated until its shared contract is implemented; parsing JSONL is insufficient. |
| §13 text/JSON/error/limits and `status --graph` | W1 starts truthful capability reporting; W2–W5 test each admitted command's stdout/stderr, error/exit contract, completeness and measured size limits. W6/W10 close compatibility census. Do not defer automation-facing contracts until the final polish pass. |
| §14 all existing command families | W6 command/writer census, with incremental coverage in W2–W5; W10 final qualification. An unhandled existing writer is an adoption blocker, not a documented escape hatch. |
| §15.2 client routes, serving, credentials and administration | W5 Read and W8 Update share domain paths, preserve binding/auth rules and never fall back to stale local writes. W6 recovery proves backup continuity. Promotion/steal and ledger administration beyond the qualified initial topology remain later, explicitly unavailable work; a proposed grammar is not a capability claim. |
| §16 acceptance matrix | W10 consolidates receipts already produced by the owning packages and the independent installer. Missing evidence remains a gap, not a pass inferred from another backend or related operation. |

These items are included in the provisional package estimates where required by complete A; the table does not provide free additional scope. If the counted census or accepted CLI feedback expands them beyond the estimates, revise the cost range at the daily checkpoint and C0 reassessment.
