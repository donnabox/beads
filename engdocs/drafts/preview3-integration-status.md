# Preview 3 planning and execution

Last checked: October 15, 2026, 22:57 PDT. This is a milestone snapshot, updated when the state changes; it is not a live CI dashboard.

This tracker covers the release plan and its execution. The active first phase is establishing a qualified Preview 3 integration branch. The integration-base milestones below are only what we need to begin producing Preview 3 builds; they are not the complete release plan. The selected Memory target below is a requirement commitment, not a qualification claim; other workstreams enter the cutline only by explicit decision.

## Why we are releasing previews

Each preview should make progress toward one or both goals:

1. Satisfy the requirements of [Chris's Memory proposal](https://github.com/gastownhall/beads/issues/5877). This has somewhat more time pressure.
2. Develop the Beads Data Protocol (BDP) specification and reference implementation toward a stable, interoperable protocol.

Memory functionality is intended to merge into fork `main` when qualified. BDP remains experimental until its protocol is stable, but preview builds expose the BDP functionality they include directly, without an opt-in switch. Experimental describes its evolving contract, not an availability gate. The selected Memory target is below; this policy does not choose additional Memory features or BDP profiles for Preview 3.

## Compatibility policy

Strict compatibility with Preview 1 or Preview 2 is **not a Preview 3 goal**. Those previews can inform the design, but they do not freeze their CLI, data or protocol choices. Record intentional changes and their user impact in the Preview 3 release material.

Compatibility with the shipping `bd` product **is a goal**. Inventory and validate existing behavior before adapting it. When preserving a shipping behavior is unusually difficult or would make the resulting product awkward, make a deliberate, reviewed exception and document the changed behavior rather than silently treating compatibility as either absolute or irrelevant.

Data migration from the shipping product or earlier previews into Preview 3 is **not a release goal**. Shipping-product behavioral compatibility does not imply an upgrade path for existing data. Iris's import/export work could supply a migration path if it is ready and explicitly accepted into the cutline; otherwise, plan it as a Preview 4 candidate. Document the actual supported path when the Preview 3 release is frozen.

Preview 2 already demonstrates end-to-end Memory workflows. After reviewing a source-level Preview 2 report-card draft against the [Memory proposal](https://github.com/gastownhall/beads/issues/5877), Donna selected the following **Preview 3 Memory target**: intentional Memory Bead/Link Type definitions, a complete Memory record and identity slice, a useful shared History/Provenance V2 slice, and reliable Link Reference/lifecycle semantics. The requirements and acceptance outcomes are recorded below; qualification is still pending. Pre-installed Memory Types are acceptable for Preview 3; user authoring of Type definitions is a separate decision. Trish owns the Type-system foundation and existing Issue/Dependency definitions, Donna works directly with Melinda on Memory Types and record semantics, Polly owns the History/Provenance slice, and Trish owns Link constraints and Reference/lifecycle design. This selection does not add the rest of the Memory proposal, the full Type revamp, or all BDP profiles to the Preview 3 cutline.

## Selected Preview 3 Memory requirements

These are target outcomes, not claims about current code. The owners should turn each row into concrete product/spec changes and acceptance cases, identify cross-workstream dependencies, and bring unresolved semantics to Donna. Janet integrates only qualified exact-source results.

| Slice | Owner and dependencies | Required observable outcome |
| --- | --- | --- |
| Memory Bead and Link Types | Melinda with Donna; depends on Trish's reviewed Type foundation and constraint model. | Define intentional, distinct nominal Memory Bead and informational Link Types with schemas and relationship rules. A pre-installed catalog is sufficient for this preview. Memory Links do not change task readiness; Type meanings and constraints are visible to the product and BDP. Do not substitute an unconstrained domain `kind` field for the distinct Types under design. |
| Memory record and identity (#1) | Melinda with Donna; align the canonical record with the Type and shared History contracts before implementation is assigned. | A Memory has a canonical identity independent of human keys, a Project ID, complete title/body/metadata and derivation/Inception state, and a structured read that exposes those fields and Change Attribution. Keys can be assigned, looked up, repointed, released, and checked for uniqueness without changing canonical identity. Search, recall, update, and listing distinguish empty content from absence and preserve honest output bounds. |
| Shared History and Provenance V2 (#2) | Polly; coordinate with the existing Jim History lane and the pending Provenance V2 design. Janet alone owns build/Dolt/release qualification. | Accepted Memory creation, update, deletion, and restoration have revision Change Attribution; no-op requests create no version. A caller can read and compare exact retained versions, restore a deleted Memory under the same identity, and consume a project-wide change feed. Guards and unguarded predecessor disclosure remain truthful. A removed historical state returns an explicit retention/erasure reason instead of silently resolving elsewhere. Preserve the shared History contract rather than inventing a Memory-only one. |
| Link References and lifecycle (#3) | Trish; extend her current Link-constraint work and coordinate with Melinda's Memory Link Types and Polly's retained-version rules. | Link endpoints represent `latest` or an exact pinned Bead version. In-Scope targets are validated; cross-Scope References are carried without pretending to resolve them. Reads distinguish expected kind, observed kind, unavailable Scope, gone target, and invalid version. Deleting a Bead removes Links following `latest`, retains valid pinned Links, and reports removed Links and newly versioned sources. A Task's structured record exposes its informational Memory Links without loading target bodies or affecting readiness. Updating a Link preserves its ID; removal addresses its ID, and both version the source where required. |

Acceptance must cover the corresponding [Memory proposal](https://github.com/gastownhall/beads/issues/5877) requirements and whole scenarios, including S1–S3, S5, S7–S10, and S15–S16, on the exact installed candidate in embedded and ordinary shared-server modes. Any scenario still excluded at the Preview 3 cutline needs an explicit Donna-approved deferral; a partial journey does not count as a pass.

Polly completed a **planning handoff** for the shared History/Provenance slice, covering Memory R12–R16 and S3/S7/S15–S16 against frozen Preview 2 source and Jim's current History work. Its provisional source order is: (1) attributable accepted revisions and complete exact reads/versions/compare; (2) a shared ordered project History producer, feed and truthful removal evidence; (3) guarded same-ID restore using that ordered producer. The first step alone does not complete the selected slice. Donna selected terminal History revisions for ordinary Bead and Link deletion and a shared `revisionProvenance` envelope; these product directions still require BDP schemas, examples, receipts, Events and conformance work. Preview 1/2 adapters are not a compatibility requirement. The packet also requires an accepted cross-kind writer/output coverage matrix, canonical Memory comparison fields with Melinda/Trish, time-to-position and cursor rules, current admission checks on restore, Jim's native mint/participation fences, and real removal producers. These are entry gates, not additional completed features. Codex and Claude reviewed the plan and their findings were folded in; Gemini review was unavailable due quota. No source PR, test result, or protocol adoption is claimed from this planning review.

## Parallel workstreams to assess

These are planning lanes, not promises that all their changes will ship in Preview 3. Each lane needs a concrete proposal, acceptance scenarios, dependencies, an owner, and a qualified source PR before it enters the cutline.

| Lane | Owner and next gate | Preview 3 inclusion |
| --- | --- | --- |
| Type system, metatype, Issue/Dependency definitions | Trish owns design and BDP/CLI contract; her separate implementation agent is working on the runtime. Review the implementable Type draft and qualify the foundation Memory Types require. | Required foundation for selected Memory Types; broader revamp boundary to decide |
| Memory Bead/Link Type definitions and record identity | Melinda works directly with Donna; coordinate with Trish's Type foundation and the selected requirements above. | Selected target; design and qualification pending |
| Memory History and Provenance V2 | Polly's reviewed planning packet names three provisional source partitions and cross-owner entry gates. Apply Donna's terminal-revision and shared-envelope rulings; reconcile Jim's History mint/participation seams, Melinda's record shape, Trish's Link rules and Vickie's BDP contract before source work. | Selected target; planning handoff complete, implementation and qualification pending |
| Link References and lifecycle | Trish extends her Link-constraint work to the selected Reference, deletion, and structured-record outcomes, with Melinda/Polly dependencies explicit. | Selected target; design and qualification pending |
| Memory proposal gap audit | Janet has completed a source-level draft covering all 32 requirements and 16 scenarios against exact final Preview 2. Donna and Janet will edit it together before commenting on the proposal; the assessment is not a new scenario test pass or Preview 3 cutline. | Draft complete; edit pass pending |
| Additional Provenance V2 beyond selected Memory revision behavior | Finish the broader design separately; activity provenance and unrelated domains are separate scope choices. | To decide |
| Import/export | Iris owns data movement, initially to avoid in-place upgrades; evaluate installable Bead/Link/Type packages as a second use. | Migration is not a Preview 3 release goal; include only by an explicit qualified cutline decision, otherwise consider for Preview 4. |
| BDP spec and reference implementation | Vickie is finishing the Write profile implementation and validation suite, then completing BDP spec/reference cleanup. This work proceeds independently of the release train and gates what `bd` can implement. Any qualified BDP functionality included in a preview is directly available. | Experimental contract; exact `bd` slice to decide |
| Bead/Link CLI surface | A read-only exact-release inventory is complete. It maps namespace behavior, Dependency Link limits, and black-box harness gaps; settle semantics and acceptance cases with Donna before implementation. | Inventory complete; inclusion to decide |
| Remaining CLI surface | A separate read-only exact-release inventory is complete. It maps graph-admitted commands, refused families, and existing compatibility harnesses; resolve semantics before implementation. | Inventory complete; inclusion to decide |
| Integration and vendoring | Janet owns the fork/upstream merge, exact-source qualification and a reproducible `bd`/BDP dependency path. | Required infrastructure |

## Current state

The final Preview 2 source is frozen at [`graph-preview2`](https://github.com/donnabox/beads/releases/tag/graph-preview2), commit `703bb13e91aa2cab2092443319139006d74242c8`. Fork `main` is still `40b3232456dfcbe621ea66ee55d635ac56634a1e`; live upstream `main` has advanced to `b303d494dfb436784d5022cfb82e5503f3769f18`. The isolated successor below remains pinned to prior upstream `773b537c75c3fdf7c38371d5da5cd3d98b615e71` and has not been silently refreshed. This draft PR tracks planning and execution; there is no published Preview 3 integration branch or source catch-up PR yet.

The earlier local merge rehearsal combines final Preview 2 with upstream `b911ed28cd47da47e167484f32db586e7ad8a279`. Its focused CI-policy compile failed and the full attempt remains preserved. An uncommitted successor combines final Preview 2 with pinned upstream `773b537c75c3fdf7c38371d5da5cd3d98b615e71`. Its 31 textual conflicts have provisional resolutions and no unmerged files remain. The scripts package compiles. Its first focused CI/Bazel-policy run failed in 30 top-level test groups; after isolated source-list, tag-taxonomy and workflow-caller permission fixes, a complete rerun remains **RED in 27 groups**, with zero skips. Each attempt is frozen separately. The remaining failures expose a real mismatch between upstream's newer Bazel workflow/test model and legacy fork lanes that still supply required gates. No merged-source qualification passed, and the permission change remains unpublished pending whole-workflow review.

The source-level Memory report card draft now assesses all 32 requirements and 16 scenarios, with no new scenario execution claimed. Its initial cross-cutting findings include partial local History but no restore or project-wide change feed, fixed rather than user-defined informational Link Types, unpinned local Link endpoints, refusal to delete a Memory with any incident Link, and no connected interchange path. Task records also do not embed informational Task→Memory Links; those can be found through the separate Link collection. Donna and Janet will edit the draft before posting a comment on the Memory proposal. Only the explicitly selected slices above are Preview 3 requirements; other findings remain candidates or deferrals.

## Integration-base milestones

- [x] Freeze Preview 2 source and release tag.
- [x] Preserve a local merge rehearsal and resolve its recorded conflicts provisionally.
- [ ] Audit the combined CI/Bazel configuration and conflict resolutions.
- [x] Start an isolated successor from final Preview 2 and pinned upstream `773b537`; keep it unpublished until qualified.
- [ ] Pass complete local build, CLI, storage, and installed embedded/shared-server Dolt qualification on the exact successor.
- [ ] Open a draft fork PR; pass whole exact-merge hosted gates and independent evidence review.
- [ ] Obtain Donna's approval to update fork `main` with the qualified merge.
- [ ] Create the Preview 3 integration branch from the verified fork `main` commit.

## Release planning and delivery

- [ ] Review and land the Type-system foundation and an implementable BDP Types draft; qualify the product implementation separately.
- [ ] Melinda and Donna define the selected Memory Bead/Link Types and complete record/identity contract on the reviewed Type foundation; qualify the resulting implementation.
- [ ] Settle Polly's reviewed History/Provenance entry gates with Donna and the named owners, then implement and qualify the provisional accepted-revision → ordered-feed/removal → guarded-restore partitions. The planning packet is complete; no source partition has landed.
- [ ] Trish reconciles Link constraints, Memory Types, and retained-version rules, then lands and qualifies the selected Link Reference/lifecycle outcomes.
- [ ] Janet and Donna audit each Memory proposal requirement and scenario against verified behavior, then name Preview 3 candidates and explicit deferrals.
- [ ] Convert the selected Memory requirements into exact CLI/BDP acceptance cases and an explicit dependency order; record any Donna-approved deferrals.
- [ ] Inventory Bead/Link CLI commands and harness coverage, then settle semantics before assigning implementation; repeat separately for the remaining CLI surface.
- [ ] Choose the Preview 3 target date/window.
- [ ] Freeze the full Preview 3 cutline around the selected Memory target; record any additional included, experimental, and deferred work explicitly.
- [ ] Work back from the target date to set feature freeze, integration, RC, manual-test, documentation, and release decision checkpoints.
- [ ] Integrate and qualify each accepted change on the Preview 3 branch.
- [ ] Prepare a candidate runbook and manual playground for review.
- [ ] Freeze a verified release candidate, complete review, and publish the final release after Donna's sign-off.

## Current focus

Choose one coherent CI lane model for the pinned successor, reconcile the remaining 27 failing policy groups and Bazel target partition, then repeat the complete focused static check. Incorporate the newer upstream `b303d49` commit in a separately preserved successor before whole-source qualification and publication. The previous failed attempts remain separate. No Janet database test is running as of the timestamp above. Until qualification completes, the successor is not ready for submissions. Independent work can branch from the immutable `graph-preview2` tag now.
