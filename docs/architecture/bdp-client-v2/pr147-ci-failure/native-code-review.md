# Native code review: PR147 two-test CI repair

No actionable findings. Reviewed diagnosis.md, native-plan-review.md, root-plan-dispositions.md, the actual complete two-file diff and surrounding test context against HEAD b0213805e4b15bb7bac4e09da5fdddb236cf6ea5. This is independent source review only: no tests, product imports, source edits, external publication or additional delegation.

Only transaction-set.test.ts and transaction-submission.test.ts are modified in the checkout. No production source or dependency changes appear. The exact diff SHA256 is b655e721ed00449cdbf739d969b38a31443d9afac5ba72a2fb0589a802122b8e.

The stale depth refusal is replaced with actual successful preparation, two carrier operations, the expected updateWhere operation and literal Selector identity, plus successful branded Selector-capacity lookup. Invalid Selector syntax still rejects first, and the following forged capability still rejects. Existing literal normalization/identity and mixed-carrier diagnostic-pointer assertions are untouched.

The independent equality helper preserves its prior finite acyclic JSON comparison domain: identical values short-circuit via === (including signed zero), primitive/null/type mismatches reject, array kind must match, complete sorted own-key vectors must match, and every paired child is pushed onto the iterative stack. The new direct array/key checks are equivalent to the old matcher predicates in this domain. Old matchers remain on mismatch paths for diagnostics, followed by explicit throws; successful nodes no longer invoke matchers. It does not import a production comparator, recurse, omit nested values, memoize comparisons away, or impose a new depth limit. Object prototypes remain intentionally irrelevant.

The new top-level helper callback has independently parsed equal depth-12000 inputs and a differing primitive at the same depth; array/object, replaced equal-count key, missing/extra key in both directions, null/object and permuted-array controls reject. Key insertion order, signed zero and null-prototype/plain-object controls pass. These controls operate outside the original store-owner work budget and introduce no HTTP, child process or store lifecycle.

Byte comparison against HEAD confirms the entire durable deep-submission callback is unchanged, including all 18 helper callsites and receipt/resource/group/no-op/replay/rollback/expiry/page/reopen/frozen-value assertions, depth12000, request-byte checks, store cleanup and 20-second work / 5-second cleanup / 30-second callback limits. Test-helper improvement does not itself establish CI timing headroom; root's focused and complete affected-file validation plus exact successor CI remain required. Prior CI failure evidence remains valid history.

Reviewed source SHA256:
- packages/server/src/transaction-set.test.ts: 86642f122eb0ab8d0ba1b5c454d4bd76fb31a7b1c485e6bbe05284c4ccd9b85f
- packages/server/src/transaction-submission.test.ts: a88e229297744c6915db8a370a9c504ce1f9bbd1b2187c6498133fdccd76f921

Source hashes were rechecked immediately before writing this report under the session guard.
