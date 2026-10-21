# Receipt missing-context pending merge reanchor notes

Planning only; no product authoring or execution authorized by this note.

Current clean checkout: `/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence`, branch `codex/vickie-event-missing-context`. Current candidate HEAD `db5a4d6a41814e59704ebef60fda02a2d57e9ad2`, tree `2e16329e195eb8214836de20340a5887236d78f0`, parent `3df82ac480f50c8f486b4e456336366e7acbdf87`. Actual main at the last root verification remains `88c9408e908590ab1a4c64c9b9dee1a8b705999c`; this note makes no remote merge claim. Root must verify PR127 merge/actual main ancestry and tree, switch the checkout, then explicitly release authoring.

Exact future branch: `codex/vickie-receipt-missing-context` (not created here). V3 plan/ledger remain frozen. The seven existing owned paths match their v3 hashes; the eighth, `apps/bdptest/src/transactional-receipt-missing-context-installed.test.ts`, remains absent.

| Existing owned path | SHA256 |
|---|---|
| `packages/server/src/transaction-store.ts` | `b2a444f7997ba4e2b34bd57ed2a8119f0999e6146305b5aae4fc8735ff78bc7c` |
| `packages/server/src/transaction-service.ts` | `191767f53b7f95503e39d7579a16b4843b64d90b4c887257a22352c446058409` |
| `packages/server/src/transaction-http.ts` | `9db33008a4c3a66214e2bf523b0b90f79c216c1f3f304c6df0491a0d1904439a` |
| `packages/server/src/transaction-receipt-pages.test.ts` | `e7bfb2a9d0a5f39a0033a13d59bfaee0d55263749b2cd3a4991237156b9da559` |
| `packages/server/src/transaction-service.test.ts` | `5e8dc095dcbe9c8fdfb28a5dea4db024ddc65e7e4f51175edeae5a90ea7125be` |
| `packages/server/src/transaction-http.test.ts` | `45ba9da34887b4daac1ed34cd9a623d7468f2448b5e8a744f52bebe23f33ef2a` |
| `docs/design/transaction-implementation-progress.md` | `c31d464b9d7928639c54acf827844d3ae62691d026ee9557ed6f7b921643b278` |

Carry forward root's native closure checkpoint into the eventual draft: capture the cleanup-array boundary before source HTTP fixture startup; in local finally splice and drain only newly registered callbacks in reverse order, catching each cleanup error separately and aggregating with the primary error. This covers startup and close failures without editing the shared fixture or relying solely on its global hook. Both advertised and pagination read limits must equal the owner's receiptMaximumItems2. Counts and selectors remain unchanged.

No product files, branch, tests, build outputs, services or installed artifacts were changed or executed.
