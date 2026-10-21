# PR121 bounded merge-gate preparation

No new source discrepancy found. Candidate remains clean `31e74290fb43974ce6f7b15ed1c421badf180e4c`, tree `5c14bcc20ec7a72118bc07d2ed5896703211a660`, ordered ordinary parents `fc47fbfb5f1055766e2f680df40a1d729d39638d` and `40806dd3ac3e74ff94c07330c89ee47dc730080c`. This report is read-only preparation, not a CI or merge verdict. PR120 must land before PR121.

## Exact expected full-suite delta

Compared actual408→31e source: no new test files, no removed/renamed existing cases, no changed discovery/skip/configuration. Three existing store cases gain assertions. Submission adds two `it.each([false,true])` groups (maintenance-fault propagation and normal closed shape), totaling four cases. HTTP adds one maintenance-fault case. Therefore the full suite gains exactly five cases and zero files. The selected local14 consists of five new and nine existing cases; its separately executed old-guard derivative is external and does not add a CI case.

The preserved completed70ac full-run summary in `transaction-comparison-timeout-repair-ci-first-full.log` at2026-10-14T18:39:25Z is1 failed +9255 passed +1 skipped =9257 across172 files.427 only adds assertion messages;408 only changes CI timeout. Thus408's expected all-pass inventory is9256 passed +1 existing skip across172 files. PR121's expected all-pass inventory is **9261 passed +1 existing skip =9262 cases across172 files**. The deliberately separate downstream preflight remains seven passing cases and must not be added to the full-suite total. Full-suite assertions still need actual terminal success; row counts or a cancelled job are insufficient. The older30035 failure log has a different9255-case inventory and is not this baseline.

Workflow, package.json and Vitest configuration are byte-identical408→31e. All commands, action pins, required check identity, inner bounds and final gates remain unchanged. Expect full suite, explicit preflight, smoke and Read evidence gate to finish successfully under the reviewed20-minute job. Existing Read exit2 means no claim/fail-closed under its unchanged gate; other error exits are not successful verification. Investigate any inventory difference, test failure, cancellation or skipped final gate rather than silently adjusting expectations.

## Review and disposition chain

The external Claude code review172f87e covers original69c3 draft only. Root dispositions preserve that boundary and explicitly acknowledge both subsequent typing failures. Native final closure covers the actual wrapper/stable-local corrections, factual doc update, fourteen passing source cases, intended old-guard503-versus500 discriminator and cleanup at fc47. The separate installed native closure covers exact fc47 fresh packages, eight-package closure, one14-request child and setup-failure distinctions. It is ordinary healthy regression, not installed fault qualification. Integration preserves every fc47 production/test/doc entry and adds only the reviewed408 CI blob; no evidence is relabeled as execution at31e.

Verified current artifact hashes:

- `transaction-comparison-fault-guard-claude-code-review.txt` — `172f87e16713899b3d686a603fc6ceb4ceb04ec5204c4172c5b4d2b67d15a49f`
- `transaction-comparison-fault-guard-code-dispositions.md` — `d7711a290648f1984c83784ebcfb4f9f2eee7fbaee0b4f4e30a489977480086f`
- `transaction-comparison-fault-typing-repair/native-final-closure.md` — `8f6a154dc312a1d7c7878054e9422699234d6598be4577bfe2774a3ba030889a`
- `transaction-comparison-fault-typing-repair/validation-report.md` — `856abff27cc2d715c47ffa06b80b7abcfaca3cd9f93bdddba98362f953931248`
- `transaction-comparison-fault-typing-repair/fresh-healthy-installed/report.md` — `c22f49fb2fc4794353b8a9bd8d5ef92378e6c9d1850907742ddc398798ef44b8`
- `transaction-comparison-fault-typing-repair/fresh-healthy-installed/native-installed-closure.md` — `32f1e98e77abf375b79b26fb5c4d02710f4b8cb4e50df2d13ad6d33e2b9d532f`
- `maintenance-integration-report.json` — `08ad9d672d7dee996624660d1e79ce3cbca11e030d3cb04804d17ad811a06a0a`
- `pr120-ci-headroom-native-code-closure.md` — `4f6eab72509c3fab57e6c95269232d1b1f8cf6ff6bee9f6731c473ccae859e7d`

## Publication/merge checks remaining for root

1. Verify PR120's exact head408 successful required CI and all final gates, then merge and verify actual main ancestry/tree. Do not infer success from this report or from PR121 running concurrently.
2. Confirm PR121 still points to exact31e and retains both parents/clean reviewed tree. Its expected combined tree is5c14. Read exact CI head, tested merge/base commit/tree, conclusion, terminal inventory and final gates; CI38625916119 was supplied as running, not queried here.
3. Once PR120 lands, an ordinary main merge can preserve PR121's existing31e head:31e already contains408. If actual main's reviewed content and Git merge result give the same5c14 tree, no content repair, rebase, artificial integration commit or redundant rerun is justified merely to make main an explicit parent of the PR branch. Preserve ordinary ancestry and let root verify the actual tested combined tree and required-check applicability. If main adds any unexpected content or the tested tree differs, stop and reconcile; this conditional statement is not a waiver of platform-required checks.
4. Merge PR121 only after those gates, then verify the resulting main tree and ancestry. Preserve historical failed CI/setup attempts and the unresolved earlier W1b cause; a later successful run does not retroactively explain them.

No tests/builds/installs/services/network/pushes/source changes were performed. Only local source/diff and retained evidence were inspected and this report written.
