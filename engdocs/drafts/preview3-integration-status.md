# Preview 3 planning and execution

Last checked: October 15, 2026, 20:15 PDT. This is a milestone snapshot, updated when the state changes; it is not a live CI dashboard.

This tracker covers the release plan and its execution. The active first phase is establishing a qualified Preview 3 integration branch. Proposed workstreams will be added to the release cutline only after their scope and qualification are decided; listing a workstream here will not itself commit it to Preview 3.

## Current state

The final Preview 2 source is frozen at [`graph-preview2`](https://github.com/donnabox/beads/releases/tag/graph-preview2), commit `703bb13e91aa2cab2092443319139006d74242c8`. Fork `main` is still `40b3232456dfcbe621ea66ee55d635ac56634a1e`; upstream `main` is `773b537c75c3fdf7c38371d5da5cd3d98b615e71`. This draft PR tracks planning and execution; there is no published Preview 3 integration branch or source catch-up PR yet.

A local, uncommitted merge rehearsal combines final Preview 2 with earlier upstream commit `b911ed28cd47da47e167484f32db586e7ad8a279`. All 30 recorded conflicts have provisional resolutions, but the merged source has not passed build or test qualification. Upstream has since advanced by one commit affecting 15 files, which the successor must include.

## Integration-base milestones

- [x] Freeze Preview 2 source and release tag.
- [x] Preserve a local merge rehearsal and resolve its recorded conflicts provisionally.
- [ ] Audit the combined CI/Bazel configuration and conflict resolutions.
- [ ] Refresh a successor from final Preview 2 and current pinned upstream.
- [ ] Pass complete local build, CLI, storage, and installed embedded/shared-server Dolt qualification on the exact successor.
- [ ] Open a draft fork PR; pass whole exact-merge hosted gates and independent evidence review.
- [ ] Obtain Donna's approval to update fork `main` with the qualified merge.
- [ ] Create the Preview 3 integration branch from the verified fork `main` commit.

## Release planning and delivery

- [ ] Agree the Preview 3 cutline and record which parallel workstreams are included or deferred.
- [ ] Integrate and qualify each accepted change on the Preview 3 branch.
- [ ] Prepare a candidate runbook and manual playground for review.
- [ ] Freeze a verified release candidate, complete review, and publish the final release after Donna's sign-off.

## Current focus

Review the CI/Bazel reconciliation before making the successor merge. No build or database test is running as of the timestamp above. Until qualification completes, neither the local rehearsal nor a new Preview 3 branch should be treated as ready for submissions. Independent work can branch from the immutable `graph-preview2` tag now.
