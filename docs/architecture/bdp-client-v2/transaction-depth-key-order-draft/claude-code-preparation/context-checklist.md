# Context checklist — preparation only

The source is being authored. No current draft was frozen, no final code prompt assembled and no review launched. Read baseline through `git show 0d63cf59d447eca2e7ff8b2e0e6a5f31513b9a3a:<path>` if needed during preparation. Final excerpts must come from the root-frozen successor and be hashed with it.

Use the complete eight-path patch/new file once, not whole repeated store/test files. Include final imports and the complete relevant function or discriminated branch where surrounding code is needed:

- `transaction-batch.ts`: prepared branding/constructors, direct mixed/set/singleton/sequence preparation, depth classification and final object ownership, normalization/assertPrepared. Include `json-admission.ts` iterative depth check and metadata parser contract only as needed.
- `transaction-store.ts`: receiving/admission result types, row expiry, complete `admitKey`, public `admit`, controls snapshot/context, plus all internal no-controls callers in cancellation/sequence turn branches and reservation retraction behavior.
- `transaction-comparison.ts`: complete small comparison loop and derived union; dependency normalization/alias equality and iterative semantic identity encoding.
- `transaction-submission.ts`: complete small receiving function and actual imports; relevant service submit dispatch and HTTP refusal renderer/direct-vs-sequence branch. Include accepted receiving-order spec lines, not the whole spec.
- Source tests: complete changed callbacks and helper setup/cleanup referenced by them, including strict singleton parity, 20k retained metadata specimen construction, pending reservation cases, exact 13-request HTTP oracle. New installed file is complete in patch; supplement only imported launcher/serializer helper contracts needed for provenance, deadline and cleanup.
- Progress document: complete changed prose, identify claims as implementation or observed evidence accurately.

Freeze handoff fields root must supply:

1. Actual HEAD/tree, exact eight changed paths and full source hashes; complete root source manifest path/hash and full patch path/hash (including new test).
2. Exact `git status --porcelain=v1 --untracked-files=all`, `git diff --binary HEAD --` SHA256 and exact untracked file list. The tracked diff alone excludes the new test; separately hash every owned path and include the complete new file in the full review patch. No clean-HEAD assertion is used.
3. Final context file hashes, v2 plan/ledger/disposition identities, actual completed static-check artifacts and any remaining pending gates. Never relabel the original parent as the dirty source.
4. Assemble `prompt.txt` with all placeholders replaced and verify compact full-diff coverage; create `final-launch-binding.json` from the closed example, `status=frozen-code-review`, populated sourceHashes/plans and dirty-state fields. Pin the source manifest and full patch in plans as well.
5. Root alone supplies an explicit released `root-release.json` binding exact prompt, runner, binding and HEAD. No live binding or release file exists in this preparation.

At authorized launch, invoke the existing guarded-action.py with `python3 <this directory>/run.py`; do not invoke an example. CLI argv is exactly `/opt/homebrew/bin/claude --safe-mode -p --tools '' --no-session-persistence`. No model override, product command or reviewer tools. One attempt, 895-second TERM / 900-second bound with KILL fallback; preserve silence/failure as actual outcome, no automatic retry. Capture full stdout/stderr, exit, before/after source and artifact hashes, process group/sampled PID absence and temporary-path inventory. Sampling is not proof against detached/short-lived descendants. Source can stay dirty while ignored setup/build outputs appear; tracked/untracked review source and bound artifacts must stay exact.
