# C3 external repair — native closure

**No remaining actionable finding in this bounded repair.** Both accepted external findings are closed, with the related oracle-label clarification. This is independent read-only source/evidence review, not a new test run, installed observation or profile admission finding. Root may proceed to its final-head CI/integration gates; the limits-projection lane remains separate.

## Exact reviewed source

- Frozen `a53f4bbf573c64fe900de373b3fef75170002f7c`, tree `4a6f4cc5ad198c7beb2d0f855583d0f33e75a037`, direct parent `3f486c245bfd31f41fd42f5acc8e0f13683bad42`.
- Checkout `/Users/dbox/repos/wt/vickie/vickie-tx-discovery-repair`; HEAD and clean status independently checked.
- Exactly four files differ: transaction-discovery.test.ts, transaction-read.test.ts, transaction-service.test.ts and schema-graph-policy.json. No production source, schema, interpreter, config, store, export surface or admission changes in this successor.
- Read original `transaction-cold-discovery-claude-code-review.txt`, root accepted `transaction-cold-discovery-root-code-dispositions.md`, full successor diff, existing policy/version guard, parser initializer and routing/fixture seams.

## Dispositions

1. **High: stale interpretation-policy index pin — closed.** Manually reviewed the original index delta: two parser and two type reexports from transaction-discovery-values.ts. The new parser owns a private lazy AJV instance; validators initialize only when those parsers are called. It does not replace the installed schema graph interpreter or its admission, graph shape, URI/reference resolution, built-ins, error or budget logic. The unchanged policy's version rule permits an explicit reviewed behavior-preserving refresh. The successor adds that justification and changes only the index hash to `92e547e807ec2d23223c47dbc4e04867d55db89ac9aab5fbd55f33349f92bdc6`. Independently recomputed all14 pinned file hashes and compared old/new manifests: sole changed pin is index.ts; policy remains schema-resource-index-2. No row was removed and the test still requires all14 owners. This closes the specific full-suite tripwire failure; it does not assert a complete CI run occurred.

2. **Medium: receipt route only had a literal oracle — closed.** `transaction-service.test.ts:1199–1210` now binds the actual first completed receipt ID to `observed.discovery.receipts`, then invokes the real source handler at that receipt URL. The submission and lookup use the same principal `p` (batch helper authorization and access constant), and the fixture has actually committed two transactions. Status200, complete body equality to the first receipt, and current response position pos-2 distinguish a retained receipt fetch from an unknown route or a stale body-context shortcut. It is a source handler observation, not a new installed HTTP claim. The root-prefix binding plus real URL fetch meets the accepted two-part witness; no string-only assertion substitutes for the fetch.

3. **Medium: Snapshot route only had a literal oracle — closed.** `transaction-read.test.ts:836–853` requests the actual discovered snapshot URL through the real handler, not a reconstructed request target. It confirms status200 and a manifest-shaped body with independently known empty fixture scope/context, pos-0, genesis checkpoint and empty complete Bead/Link streams, plus actual current response position. This distinguishes genuine Snapshot creation from unknown-resource404 or the inherited discovery500. The added pathname equality is supplementary; the actual successful handler result supplies the missing route witness. Existing cleanup closes the owning service/store; no new fixture or live child was added.

4. **Related evidence wording — closed.** `transaction-discovery.test.ts` renames the substitution test to “literal projection oracle.” The literal mapping remains useful and unchanged, while live owner/context and receipt/Snapshot route bindings are now correctly demonstrated in the separate source tests. No widened evidence claim or new protocol meaning is introduced.

The accepted3f pagination receiver repair remains byte-for-byte in production. These test additions do not modify its getter-once, original receiver, invalid callback or scalar-capture behavior.

## Executed author evidence inspected, not rerun

`transaction-cold-discovery-external-repair/affected.json` and log report **84 passed,0 failed,4 files**, Vitest1.06s, supervisor exit0/1.428s. Selected files are installed-schema-graph.policy, transaction-discovery, transaction-read and transaction-service. `checks.json` records typecheck, owned formatting/lint and boundaries all exit0. The original `policy-red.exit` is1, preserved rather than relabeled.

The author also preserved an initial archive-to-source comparison failure: package.json is normalized by pnpm pack, so direct source equality was an invalid audit method. The correct retained packed-archive comparison below supersedes that method; it is not a runtime/test failure or a fresh installed test.

## Independent package identity verification

I opened both candidate tarballs and the exact retained3f archives, compared every regular member name and byte, and independently hashed the archive bytes:

| Package | Members | Exact member-path/byte equality | Exact archive equality / SHA256 |
|---|---:|---|---|
| protocol |63|yes|yes — `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227`|
| server |246|yes|yes — `92ab7c6036a6c45009c99bd6045f772f27067839207177492661d1b2af75b1dc`|

This independently agrees with package-identity.json. The 3f installed12/1 and targeted RU evidence remains historical evidence at3f with an identical package payload; it is **not relabeled as a new installed run at a53**. No separate fresh package execution is required to distinguish behavior for this test/policy-only delta, though root still owns final source CI and merge readiness.

## Review actions and limits

Read-only `git rev-parse/status/diff`, source/log inspection and Python standard-library hashing/JSON/tar-member comparison all completed exit0. No tests, builds, package writes, repository edits, network, push or child servers were run for this review. Author report.md arrived after initial inspection and was then read in full. It agrees with the independently inspected logs, JSON and archives, preserves the initial audit-method failure and correctly keeps installed evidence source-bound to3f. No new finding or evidence reattribution results. Only this review document was written. No process remains owned by this review.
