# Memory CLI Preview 3 — candidate qualification

Tracking: [7759](https://github.com/gastownhall/beads/issues/7759).
Read composition: [draft PR217](https://github.com/donnabox/beads/pull/217).
Accepted design: [PR206](https://github.com/donnabox/beads/pull/206) at
`4c352eb1da0aba9de0e6feee5ba712374cfbb0a5`.

This records candidate evidence, not installed-release support or permission
to merge main. Janet owns release and parent integration qualification.

## Read composition

The scoped reader/presenter stack is composed through ordinary merges:

| Slice | Exact candidate |
| --- | --- |
| Current-family reader, PR208 | `ed1e9e34e770921f96820cffebff764f9e9a086d` |
| Ordinary presentation bridge, PR211 | `71099cdbf027e1b42448d9f6c73ff9bc51208a90` |
| Requires/Defines/staged reader, PR213 | `b2e37a6a012be2053a01d206822be8317446f80c` |
| Staged/direct presenter, PR215 | `539f5f3785393d6af23e4e5392e872c94d20bd21` |
| Common mutation framing, PR214 | `06a727738550e83a73cce2b7c2e0bbbcf6d4b929` |
| Ten-read CLI adapter and discovery corrections | `104f0c143`, `53e60d6fabf2b2aa98d1fac04bb2ab656ee05b17` |
| Five Link-create facades, [PR221](https://github.com/donnabox/beads/pull/221) | `3c5b839c23091ebe0ab4b290a56d20c09ab985db` |
| Final composed runtime | `19af7b56a50476928a5427be69a630630bf806fd` |

The full evidence directory is local to the authoring host:
`/Users/dbox/.local/state/melinda/memory-cli-runtime-20261021/`.
Each provider receipt binds its source hashes. Logs are ordinary test records,
not cryptographically signed attestations. They retain failed attempts as well
as corrected passes.

| Check | Result and scope |
| --- | --- |
| Installed CLI, embedded and released Dolt2.2 server | PASS: all22 accepted read scenarios, staged recurrence/exclusions, direct subjects and original four-read workflow; race enabled. The final full composition before the writer-only guard/diagnostic successor passed92 test/subtest nodes with no skips in153.25s, including all four installed workflow families on both providers. Its exact source is `096ec1fc41537ae619ee1ad24557c04a6978a9b2` in `final-cli-both-engines-r2.json`. Owned server/workspace cleanup verified. All read/common/metadata source hashes are unchanged in the final runtime; `final-read-source-continuity.json` records the comparison. |
| Final read affected Bazel | PASS: the92-pass composition also recorded121 affected command XML cases:117 passed and4 optional server skips; the separate real-server run covers all four skipped provider workflows. Native nogo passed. `final-cli-bazel-r2.log` and its cases index retain the exact selection. Earlier nonverbose receipts are not assigned case counts. |
| Final writer delta on composed runtime | PASS:19 tests/subtests, embedded and released Dolt2.2 server, race enabled, no skips; 79.29s. `final-writer-delta.json` binds exact runtime `19af7b56a50476928a5427be69a630630bf806fd`; cleanup and source hashes verified. |
| Final affected Bazel and platform lint | PASS: 122 recorded command cases, 118 passed, 4 optional server skips,0 failed. Native/Windows/Darwin nogo passed84s. `final-cli-bazel-r3.log`, its case index, and `final-platform-lint-r3.log`. |
| Actual formatting guard | PASS: `//scripts/repochecks:fmt_test`. Formatting-library tests are not this guard. |
| Dependency-query compatibility | PASS: four targeted cases preserve existing duplicate-title overwrite behavior. This is narrower than whole-product compatibility. |
| Documentation | PASS: pinned generated CLI reference, freshness, and docsync after moving unreleased candidate guidance out of the published tree. No navigation/release-pin edits. |
| Normal author hooks | PASS: eight-read and staged-read code commits. Source-only initial pushes explicitly skipped inherited broad pre-push qualification; see below. |

## Reviews and corrections

Actual Claude first-four review returned14 findings/context items. Supported
corrections included failure framing, output budget accounting, JSON encoding,
quiet-mode handling, help/flag accuracy and snapshot binding. Native follow-up
found no additional concrete defects. Native staged/direct review was clean;
actual staged Claude findings were processed, including shared quiet/machine scope validation and phase membership/metadata controls.

The five-facade actual Claude review and narrow follow-up were processed,
including canonical predecessor projection, invalid-selector framing, named
property UTF-8, bounded UTF-8 source guards and single-publication diagnostics.
Native final delta audits found no supported remaining defect. The staged
presenter also has its own processed actual Claude review. These slice reviews
do not establish a fresh Claude approval of the complete combined CLI. Gemini returned no substantive review (timeout or permission
denial in the bounded attempts); council coverage is reduced.

Installed checks initially caught direction serialization, invalid-mask framing,
a footer expectation after presenter integration, and the missing mandatory
Discovery line. Corrected reruns are listed separately in the local evidence
index. A local Bazel output-symlink/runfiles failure was corrected by moving only
this worktree's convenience links outside recursive source globs; no tracked
build policy or validation guard was weakened.

## Remaining integration gates

All ten reads and five creates are composed. The final writer-only guard and
diagnostic successor passed its composed checks, preserving the unchanged
read-source bindings from the full both-provider run. No full-suite,
hosted-green or release-readiness claim is made.
Parent Mac-capacity and leak-test investigation remain Janet's lane.

The source-only input branches and PR217 source publications used an explicit
`--no-verify` pre-push exception because branch pushes with Go/Bazel changes select the entire
Bazel CI suite, including inherited parent qualification already owned by Janet.
The initial source-only branches compared their whole ancestry against main. Normal author hooks and the
targeted checks above passed. That exception does not replace full assembled
qualification. See7759 and the PR body for the standing gate.

BDP/BTS language work stays held. Assessment/conformance operations, implicit
applicability, historical traversal, remember/recall/memories/forget, native
dependency behavior and release/spec pin movement remain outside this candidate.
