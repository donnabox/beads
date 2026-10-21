# Native closure — PR93 catalog inventory repair

**One low-severity test-maintenance finding; current exact-head behavior passes.** Frozen `87abe223dd27519d900fa015845ee3d90c97819a`, direct parent `4819068d0201fb3d1fe32dd6adaf8625ef9ba666`, changes only `packages/conformance/src/read-update-catalog.test.ts`.

## Finding — first-row-dependent rejection oracle

**Low**, `packages/conformance/src/read-update-catalog.test.ts:119–122`: the forged Read-tag negative control passes the entire evolving continuation but demands an error naming `read-update.alias.unknown-subject`. The assertion helper visits catalog IDs in order. The already accepted future attribution row is index16 versus alias index35 in the RU catalog; when that row is added to the continuation, the correct denial will instead name attribution and this test fails. This is test/CI fragility, not weakened admission or a production defect. Root identified the integration case during this review; independent source inspection confirms the order and exact regex coupling.

**Fix:** select the actual `read-update.alias.unknown-subject` scenario from the continuation, assert it exists, and pass only that scenario to the forged Read-tag control. Keep the exact error assertion on the isolated row. Alternatively match the precise closed-world denial gate without relying on which valid RU row is encountered first. Do not remove the control or weaken the file inventory checks.

## Otherwise reviewed

Reviewed the complete diff and author report. The explicit filename allowlist contains exactly the baseline and reviewed reference-continuation shard. It is not inferred from untrusted manifest catalog/profile tags. The new exact index assertion pins version, both shard IDs, both sources and order; both files must exist. Each approved entry still requires RU catalog ID, manifest version2, partial cumulative coverage and membership only in the cumulative required-ID selection. Every other JSON matrix still rejects the RU catalog tag and any RU-required row. The two negative controls exercise both an unapproved filename with the actual RU tag and a forged Read tag with a contained RU row. The sealed Read ID-collision test remains unchanged.

No production admission path, public export, sealed Read source/evidence, catalog, fixture, matrix, index, runtime or validator changes. This repairs the obsolete test inventory and adds closed-world checks; it does not authorize arbitrary extra shards or make the previous failed CI retrospectively successful.

Independent verification in disposable detached `/Users/dbox/repos/wt/vickie/vickie-artifact-set-catalog-inventory-native`:

- Remote guard and exact-source checks: exit0; author checkout clean and untouched.
- Pinned Node24.16.0 / pnpm11.20.0 `install --frozen-lockfile`: exit0.
- `pnpm exec vitest run packages/conformance/src/read-update-catalog.test.ts --maxWorkers=2 --reporter=verbose`: exit0, **7 tests /1 file**, 673ms. Both denial controls passed.
- Isolated tracked tree clean after execution. No source edits, installed cohorts or broad suite repeated.

Logs: `bdp-integration-logs/artifact-set-catalog-inventory-native/{install.log,focused.log}`. Exact source CI remains a separate root gate; author424-test/static evidence was read, not claimed as independently rerun.

## Exact successor closure

Reviewed direct successor **2ec7eb5c22ebf00c045154cccbcf0f7a425a0ae2** against87. Its entire +9/-1 delta is confined to the same test file. It filters the actual continuation scenarios by the C-row ID, requires exactly one match, and uses only that real row in the forged Read fixture while preserving the exact denial assertion. Adding a different earlier RU row can no longer change which denial this isolated control expects. The inventory/index checks and first unknown-filename control are unchanged.

**Finding closed; no remaining actionable finding in this bounded repair.** Independently fetched the local frozen successor into the disposable review clone and reran the exact catalog command: exit0, **7 tests /1 file**, 663ms (`artifact-set-catalog-inventory-native/successor-focused.log`). Diff-check passes and tracked tree is clean. No broad or installed tests repeated; root owns exact-head CI. Original87 results remain recorded above and have not been relabeled as successor evidence.
