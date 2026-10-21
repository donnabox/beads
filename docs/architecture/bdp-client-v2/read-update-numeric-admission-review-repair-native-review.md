# Native closure: numeric admission review repair

## Findings

No actionable findings. Both accepted LOW findings are closed by the frozen two-test-file delta. No artifact, runtime, schema, coverage, fixture, budget or admission semantics changed.

## Exact review target

- Head: `06b0e04a921c8501e36a8a68480cee72dc4c54b3`.
- Direct parent: `bbab34f78864e7b3d2210ecc4527c9143e5e53ef`.
- Tree: `6eb97d923c18b6cb84c641220fa4cdb8f4851e5e`.
- Frozen author checkout: `/Users/dbox/repos/wt/vickie/vickie-numeric-admission-review-repair`.
- Independent detached review checkout: `/Users/dbox/repos/wt/vickie/vickie-numeric-admission-review-repair-native`.

Read the entire actual diff (two files, 92 insertions / 5 deletions), original Claude131 report, root dispositions, pause report, source audit, exact guard-removal scripts/patch records and resulting JSON failures. Source remained clean; the author checkout was not edited. Review-local Git metadata flattened inherited alternate object paths for checkout; this changed no tracked source.

## Dispositions

1. **Independent nested corpus pins — closed.** `apps/bdptest/src/read-update-matrix.test.ts:4088–4137` checks both Bead and Link N rows. The expected `deep/d0…d13/a~1b~0c` chain is constructed from independently authored constants, not parsed or derived from the observed body; shallow/deep/patch integer, metadata underflow and metadata precision are checked as lossless raw decimal strings. This preserves depth 16 and all accepted lexical partitions. The admitted `large:1e300` spelling also has an explicit pin at line 4084.

   The permanent controls at lines 4139–4182 establish that the genuine corpus passes, shorten only the deep chain, and perturb each accepted lexical class independently. They require their named guard failures. The shortened replacement is explicitly checked to differ from the original; the other replacements necessarily differ after the positive guard has established their exact needles. These controls ran independently in this review.

2. **Distinct catalog/version refusals — closed.** `packages/conformance/src/numeric-model-catalog.test.ts:79–80,120–125` now uses unique `numeric continuation catalogId guard` and `numeric continuation manifestVersion guard` labels. Their negative tests match those labels rather than the generic entry filename that also labels ownership checks. Existing ownership allowlists and wrong-owner/Read denial remain intact. Removing either named guard cannot pass by throwing an unrelated ownership error.

3. **Stale title and admitted spelling — closed.** The earlier test is accurately named `pins attribution/numeric action ledgers, raw literals and prior scenario bytes` at app line 3643. Budget checks remain in the separate numeric-admission source-growth test. The explicit `1e300` pin is present; no unrelated ledger/capture cleanup was introduced.

## Independent checks

Pinned Node 24.16.0 and pnpm 11.20.0, detached review checkout:

```sh
node /Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs install --frozen-lockfile
node /Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs build
node /Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs exec vitest run apps/bdptest/src/read-update-matrix.test.ts packages/conformance/src/numeric-model-catalog.test.ts -t 'pins numeric admission|shortened or drifted|draft numeric-model catalog|pins attribution/numeric action' --maxWorkers=2 --reporter=verbose
```

All exited 0. **11 tests passed / 2 files; 184 tests filtered**, 957 ms focused execution. These are static corpus/catalog tests; no installed authority or HTTP cohort was started by the reviewer. Build and focused logs are under `numeric-admission-repair-native/`. `git diff --check HEAD^ HEAD` passed; review and author tracked source stayed clean.

Independently rehashed and byte-compared the baseline matrix, continuation matrix, fixture, coverage and artifact-set index with the exact parent: all five unchanged and matching the author's source-audit hashes. Recorded in `numeric-admission-repair-native/artifact-audit.json`. The actual Git delta itself confirms no other tracked path changed.

## Separately inspected author sensitivity/evidence

The three authored source-counterfactual runs are retained in `read-update-numeric-admission-review-repair/`:

- `deep-guard-removed`: exit 1, exactly the Bead and Link `rejects shortened or drifted...` tests fail because the intended function no longer throws.
- `catalog-guard-removed`: exit 1, exactly the two `rejects unapproved artifact ownership...` tests fail because the intended catalog check no longer throws.
- `version-guard-removed`: exit 1, the same two ownership controls fail at the distinct version check.

Read their exact removed-line records and JSON failure messages; the script restores original bytes in `finally`, and the frozen source contains all restored guards. These are author-run sensitivity probes, not native reruns or current artifact mutations. The independent green run above executes the permanent body-mutant and precise denial controls in the final source.

The preserved pause report honestly records an initial pre-build import failure and distinguishes the subsequent green dirty-source work from unfinished frozen evidence. It is not a current qualification report.

During this review the author frozen packed run completed: `packed/source.json` identifies clean exact 06b0e04 and its installed package entry; `packed/matrix.exit.json` records exit 0 in 26.958 seconds; `packed/matrix.json` records **22 passed, 0 failed, 173 filtered/pending**. The selected command covers the aggregate partial cumulative RU observation, numeric admission rows/controls and catalog/lexical checks. This is inspected author evidence, not an independently repacked run. Its archive hashes are protocol `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c` and server `80f0ab8fa061480be11428b72b110b0f2f8185d467f561f6d30912f2e268562c`.

The completed author `report.md` and `summary.json` were also inspected after root notification. They retain the earlier pause/failure records and identify both genuine shards completing in 18.790 seconds (71.210 seconds remaining), payload `1b12629fe54c39c8f7864dbb13bdba16bcbe4298b0a5bd420fa19daaa4fb7ec2`, unchanged 129 required / 54 authored / 30 whole / 99 uncovered / 75 omitted inventory, clean source and drained children. The two written child-report byte hashes/lengths were independently recomputed against the summary. No author observation was relabeled as native execution.

## Limits

No broad workspace suite, independent installed cohort, source-counterfactual rerun, new qualification or public claim was made. Native closure is limited to the two accepted test defects and exact unchanged artifact boundary. Root retains publication, integration and qualification decisions. No outward writes or author source edits were performed.
