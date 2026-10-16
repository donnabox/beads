# Preview 3 planning and execution

Last checked: October 15, 2026, 20:44 PDT. This is a milestone snapshot, updated when the state changes; it is not a live CI dashboard.

This tracker covers the release plan and its execution. The active first phase is establishing a qualified Preview 3 integration branch. The integration-base milestones below are only what we need to begin producing Preview 3 builds; they are not the complete release plan. Proposed workstreams will be added to the release cutline only after their scope and qualification are decided.

## Why we are releasing previews

Each preview should make progress toward one or both goals:

1. Satisfy the requirements of [Chris's Memory proposal](https://github.com/gastownhall/beads/issues/5877). This has somewhat more time pressure.
2. Develop the Beads Data Protocol (BDP) specification toward a version that could be submitted to a consortium. AAIF is a possible home, not a committed destination.

Memory functionality is intended to merge into fork `main` when qualified. BDP should remain explicitly experimental until there is a stable protocol to support: a Preview 3 build may expose it behind an opt-in switch or omit it. This policy does not yet decide which Memory features or BDP profiles are in the Preview 3 cutline.

Preview 2 already demonstrates end-to-end Memory workflows. Donna identified two likely gaps for the next feature cutline: intentional Memory Bead and Link Type definitions, likely requiring a bounded part of the new [Types proposal](https://github.com/gastownhall/beads/issues/7590), and at least the revision-provenance part of [Provenance V2](https://github.com/gastownhall/beads/issues/7510). Donna considers pre-installed Memory Types an acceptable path for Preview 3; user authoring of Type definitions is a separate product decision. Trish is working on the Type-system design, its BDP specification, the Memory Type specifications, and descriptions of the existing Issue and Dependency Types. The Types and Provenance proposals are not completed implementation. Exact Preview 3 slices and a fresh gap audit against the [Memory proposal](https://github.com/gastownhall/beads/issues/5877) remain to decide.

## Parallel workstreams to assess

These are planning lanes, not promises that all their changes will ship in Preview 3. Each lane needs a concrete proposal, acceptance scenarios, dependencies, an owner, and a qualified source PR before it enters the cutline.

| Lane | Direction | Preview 3 inclusion |
| --- | --- | --- |
| Memory beads and links | Audit the Memory proposal against Preview 2 and define the remaining Memory behavior and relationships. | To decide |
| Issue and Dependency model | Trish is specifying the existing Issue and Dependency Types for completeness. | To decide |
| Types | Trish is designing the Type system and BDP Types profile plus the Memory Type definitions; choose the bounded runtime slice and its migration/CLI effects. User Type authoring is a separate decision. | Likely required slice; boundary to decide |
| Provenance V2 | Land at least revision provenance, subject to design decisions and qualification; activity provenance is a separate scope choice. | Likely required slice; boundary to decide |
| BDP core and profiles | Advance the read protocol and optional profiles without claiming a stable protocol. | Experimental or deferred |
| Import/export | Resume data movement work after Preview 2. | To decide |
| CLI completeness | Extend the bead/link command surface and usability, while preserving the qualified release line. | To decide |
| Integration and vendoring | Keep bd and BDP dependencies, qualification, and the fork/upstream merge reproducible. | Required infrastructure |

## Current state

The final Preview 2 source is frozen at [`graph-preview2`](https://github.com/donnabox/beads/releases/tag/graph-preview2), commit `703bb13e91aa2cab2092443319139006d74242c8`. Fork `main` is still `40b3232456dfcbe621ea66ee55d635ac56634a1e`; upstream `main` is `773b537c75c3fdf7c38371d5da5cd3d98b615e71`. This draft PR tracks planning and execution; there is no published Preview 3 integration branch or source catch-up PR yet.

The earlier local merge rehearsal combines final Preview 2 with upstream `b911ed28cd47da47e167484f32db586e7ad8a279`. Its focused CI-policy compile failed and the full attempt remains preserved. A fresh, uncommitted successor combines final Preview 2 with pinned current upstream `773b537c75c3fdf7c38371d5da5cd3d98b615e71`. Its 31 textual conflicts have provisional resolutions and no unmerged files remain. After duplicate test helpers were reconciled, a focused compile still failed: fork policy tests refer to legacy CI lanes that upstream's newer test model retired, while the merged fork workflows still contain those lanes. This is an active workflow/test reconciliation; no merged-source qualification passed.

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

- [ ] Define the Memory Types and revision-provenance acceptance slice; audit the rest of the Memory proposal for Preview 3 candidates or deferrals.
- [ ] Choose the Preview 3 target date/window.
- [ ] Agree the Preview 3 cutline; record included, experimental, and deferred work explicitly.
- [ ] Work back from the target date to set feature freeze, integration, RC, manual-test, documentation, and release decision checkpoints.
- [ ] Integrate and qualify each accepted change on the Preview 3 branch.
- [ ] Prepare a candidate runbook and manual playground for review.
- [ ] Freeze a verified release candidate, complete review, and publish the final release after Donna's sign-off.

## Current focus

Choose one coherent CI lane model for the successor, reconcile its policy tests and Bazel target partition, then repeat the focused static checks. The previous failed attempts remain separate. No database test is running as of the timestamp above. Until qualification completes, the successor is not ready for submissions. Independent work can branch from the immutable `graph-preview2` tag now.
