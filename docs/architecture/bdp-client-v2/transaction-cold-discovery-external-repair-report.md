# C3 external-review repair

Frozen ordinary successor `a53f4bbf573c64fe900de373b3fef75170002f7c`, tree `4a6f4cc5ad198c7beb2d0f855583d0f33e75a037`, parent `3f486c245bfd31f41fd42f5acc8e0f13683bad42`. Clean checkout `/Users/dbox/repos/wt/vickie/vickie-tx-discovery-repair`; branch `codex/vickie-tx-discovery-repair`. Commit includes Agent-Signature attribution. Guarded push checked remote launch session `2026-10-08T20:26:35Z-81e5`; exact remote branch head verified. No PR mutation, main merge or profile activation.

## Findings closed

1. **Schema interpretation policy pin.** Reproduced the stale protocol-index pin failure before repair (`policy-red.log`, exit1). Manually reviewed the additive parser/type reexports and their implementation: schema validators are lazy, private AJV state; installed schema graph indexing does not invoke these new discovery parsers. Existing descriptor admission, graph shape/resolution, built-ins, errors and budgets are unchanged. Retain `schema-resource-index-2` with explicit `transactionalDiscoveryReview`; update only the index pin to `92e547e807ec2d23223c47dbc4e04867d55db89ac9aab5fbd55f33349f92bdc6`. All other thirteen pins were individually verified unchanged and matching. A changed pinned module must include the policy tripwire in affected-test selection.
2. **Actual navigation proofs.** The existing service receipt/replay test now checks its real receipt ID against the observed discovery prefix, fetches that receipt through the same-principal source HTTP handler and asserts200, complete receipt equality and actual pos-2 context. The existing non-root Scope test GETs the discovered Snapshot URL through the real handler and asserts200, actual owner context, genesis checkpoint and empty real Bead/Link page bodies. These assertions connect discovery candidates to existing routing/owner behavior without making discovery public.
3. **Oracle wording.** Rename the hand-authored candidate substitution oracle from “independent owner oracle” to “literal projection oracle.” Actual owner binding remains in service tests.

Exactly four files changed: private policy JSON and the three authorized test files. No production, schema, manifest, package export, runtime, public admission or C4 change. Original assembly b20 and packed3f evidence remain unchanged.

## Verification

Pinned Node24.16.0 and pnpm11.20.0. All command/exit/timing evidence is in `checks.json` and corresponding logs.

- Initial stale policy test: expected failure,1 test, preserved `policy-red.log`/`.exit`.
- Four affected files:84 tests passed (`affected.json`/`.log`),1.428s.
- Typecheck, focused Biome format/lint, dependency boundaries and `git diff --check`: passed.
- Fresh protocol/server pack commands: exit0. This is package identity verification, not a fresh installed run.
- No full workspace, build or installed cohort repeated: the changed tests are excluded from emitted server output and private test-support policy is outside package `files:[dist]`.

## Exact package/evidence identity

`package-identity.json` compares every member path and file byte of fresh candidate archives against the retained source3f archives: protocol63 files and server246 files, all exactly equal, including pnpm-normalized package manifests. Initial direct archive-to-source comparison failed on protocol package.json because pnpm transforms that file during packing; `identity-initial-failure.txt` preserves this audit-method correction. The successful comparison uses packed-to-packed bytes, not inferred source equivalence.

Historical installed evidence remains **source3f only**, at `../transaction-cold-discovery-review-repair/report.md` and `summary.json`: actual TX12 exchanges/one child plus targeted RU observations; payload `5e8ce6bf6d12267a5ea999e4dee84fde15c1c820b683e88a9e712093db0c351f`. No new source binding or fresh installed claim is assigned to those runs. The successor has new test evidence and unchanged package bytes. `source.json` and `summary.json` record exact current source/file bindings and package hashes.

## Cleanup and limits

All focused commands completed normally. Final process inspection found no test/child process for this checkout or report lane; only the inspection shell/rg itself. Worktree clean. No shared databases or unrelated checkouts changed. These repairs close the accepted policy/route-evidence findings; they do not activate Transactional discovery or establish profile conformance.
