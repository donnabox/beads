# Preview 3 planning and execution

Last checked: October 15, 2026, 21:12 PDT. This is a milestone snapshot, updated when the state changes; it is not a live CI dashboard.

This tracker covers the release plan and its execution. The active first phase is establishing a qualified Preview 3 integration branch. The integration-base milestones below are only what we need to begin producing Preview 3 builds; they are not the complete release plan. Proposed workstreams will be added to the release cutline only after their scope and qualification are decided.

## Why we are releasing previews

Each preview should make progress toward one or both goals:

1. Satisfy the requirements of [Chris's Memory proposal](https://github.com/gastownhall/beads/issues/5877). This has somewhat more time pressure.
2. Develop the Beads Data Protocol (BDP) specification toward a version that could be submitted to a consortium. AAIF is a possible home, not a committed destination.

Memory functionality is intended to merge into fork `main` when qualified. BDP should remain explicitly experimental until there is a stable protocol to support: a Preview 3 build may expose it behind an opt-in switch or omit it. This policy does not yet decide which Memory features or BDP profiles are in the Preview 3 cutline.

Preview 2 already demonstrates end-to-end Memory workflows. Donna identified two likely gaps for the next feature cutline: intentional Memory Bead and Link Type definitions, likely requiring a bounded part of the new [Types proposal](https://github.com/gastownhall/beads/issues/7590), and at least the revision-provenance part of [Provenance V2](https://github.com/gastownhall/beads/issues/7510). Donna considers pre-installed Memory Types an acceptable path for Preview 3; user authoring of Type definitions is a separate product decision. Trish's type-system design is nearing completion, a separate agent is implementing it, and the BDP Types specification needs a few more design turns before it can be called a fully implementable draft. Melinda now has a separate Memory Type design task for read-only preparation. Once the Type foundation lands, Donna and Trish will make the intentional Memory Bead/Link Type decisions with Melinda driving that workstream. Trish is also describing the existing Issue and Dependency Types. Neither the Type implementation nor the Memory definitions are accepted into the Preview 3 cutline yet. Janet and Donna will audit unmet requirements and scenarios against the [Memory proposal](https://github.com/gastownhall/beads/issues/5877); no gap is silently assigned to Preview 3.

## Parallel workstreams to assess

These are planning lanes, not promises that all their changes will ship in Preview 3. Each lane needs a concrete proposal, acceptance scenarios, dependencies, an owner, and a qualified source PR before it enters the cutline.

| Lane | Owner and next gate | Preview 3 inclusion |
| --- | --- | --- |
| Type system, metatype, Issue/Dependency definitions | Trish owns design and BDP/CLI contract; her separate implementation agent is working on the runtime. Review the implementable Type draft and qualify an exact source slice. | Boundary to decide |
| Memory Bead/Link Type definitions | Melinda prepares a requirements/dependency map read-only. After the Type foundation is ready, Donna and Trish make the design decisions; Melinda drives the separate workstream. | To decide |
| Memory proposal gap audit | Janet and Donna compare each requirement and scenario with verified Preview 2 behavior, then classify gaps and choose a release cutline. | Audit first; inclusion to decide |
| Provenance V2 | Finish design and qualify at least the revision-provenance slice; activity provenance is a separate scope choice. | Likely required slice; boundary to decide |
| Import/export | Iris owns data movement, initially to avoid in-place upgrades; evaluate installable Bead/Link/Type packages as a second use. | To decide |
| BDP spec and reference implementation | Vickie is finishing the Write profile implementation and validation suite, then completing BDP spec/reference cleanup. This work proceeds independently of the release train and gates what `bd` can implement. | Experimental or deferred in `bd` |
| Bead/Link CLI surface | Unassigned. Inventory commands and existing harnesses, decide semantics and acceptance cases, then delegate implementation and run the broader harnesses. | To decide |
| Remaining CLI surface | Unassigned. Inventory separately after the Bead/Link cutline, resolve compatibility choices, then delegate implementation. | To decide |
| Integration and vendoring | Janet owns the fork/upstream merge, exact-source qualification and a reproducible `bd`/BDP dependency path. | Required infrastructure |

## Current state

The final Preview 2 source is frozen at [`graph-preview2`](https://github.com/donnabox/beads/releases/tag/graph-preview2), commit `703bb13e91aa2cab2092443319139006d74242c8`. Fork `main` is still `40b3232456dfcbe621ea66ee55d635ac56634a1e`; upstream `main` is `773b537c75c3fdf7c38371d5da5cd3d98b615e71`. This draft PR tracks planning and execution; there is no published Preview 3 integration branch or source catch-up PR yet.

The earlier local merge rehearsal combines final Preview 2 with upstream `b911ed28cd47da47e167484f32db586e7ad8a279`. Its focused CI-policy compile failed and the full attempt remains preserved. A fresh, uncommitted successor combines final Preview 2 with pinned current upstream `773b537c75c3fdf7c38371d5da5cd3d98b615e71`. Its 31 textual conflicts have provisional resolutions and no unmerged files remain. The scripts package now compiles, but the first focused CI/Bazel-policy run failed in 30 top-level test groups. The failures expose a real mismatch between upstream's newer Bazel workflow/test model and legacy fork lanes that still supply required gates. The complete failed attempt is frozen separately; no merged-source qualification passed.

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
- [ ] Melinda maps Memory Type requirements without editing Trish's work; after the Type foundation lands, Donna and Trish define the Memory Bead and Link Types.
- [ ] Janet and Donna audit each Memory proposal requirement and scenario against verified behavior, then name Preview 3 candidates and explicit deferrals.
- [ ] Set the revision-provenance acceptance slice and identify any dependent `bd`/BDP work.
- [ ] Inventory Bead/Link CLI commands and harness coverage, then settle semantics before assigning implementation; repeat separately for the remaining CLI surface.
- [ ] Choose the Preview 3 target date/window.
- [ ] Agree the Preview 3 cutline; record included, experimental, and deferred work explicitly.
- [ ] Work back from the target date to set feature freeze, integration, RC, manual-test, documentation, and release decision checkpoints.
- [ ] Integrate and qualify each accepted change on the Preview 3 branch.
- [ ] Prepare a candidate runbook and manual playground for review.
- [ ] Freeze a verified release candidate, complete review, and publish the final release after Donna's sign-off.

## Current focus

Choose one coherent CI lane model for the successor, reconcile its policy tests and Bazel target partition, then repeat the focused static checks. The previous failed attempts remain separate. No database test is running as of the timestamp above. Until qualification completes, the successor is not ready for submissions. Independent work can branch from the immutable `graph-preview2` tag now.
