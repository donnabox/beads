# PR126 first CI: actual failure and bounded diagnosis

Verdict: failed assertion, not job timeout. Do not retry unchanged or infer green from the9289 passing cases. Root owns repair/publication.

Exact run38702665285/job116314371322 ran candidate a912c23fbd80e70ead382a27e700a70123bf3f21. Initial synthetic merge410744a17cb78abc42aee1a7a84fab0e0f099ea2 has parents453b6dab386746bb03aac2c0a293572b8a71c969 and candidate, tree501791c6372c3d69b50ef38f777e79de5807f10a. The complete raw log confirms checkout of that synthetic merge. No source change or retry by reviewer.

Full raw job log retrieved through actions/jobs/116314371322/logs,3291140bytes, SHA2564a00cff967662d5d21d630feb217439d58719b1f8642adf4b1f885aba83212ca. Terminal metadata/annotations ci-terminal.json SHA256c3d0b6e034e2205acd4160ea54b514d7d716ac20698584ae11d8d3a40480aec5. Neither is truncated gh run view output.

## Failure and repair boundary

Raw log2464 first records the failed test at08:04:48.866Z; lines11075–11095 contain its sole failure:
packages/conformance/src/transactional-cumulative.test.ts:1011,
“pins the one-page matched-plus-updated witness without advertised-limit or other-outcome credit”
expected9, received10 in manifest.scenarios.indexOf(row).

Actual manifest inserts read-update.http.command-minimum-fields at index0 and preserves the existing scenario objects/order. transactional.receipt.pagination is now index10, while the older absolute index oracle remains9. The18-action shape and later content checks are not implicated by this failing comparison; their execution in this callback was interrupted by the index assertion, so do not claim those later checks passed.

Minimal proposed source repair: change only this index expectation9→10. Preserve the row-order oracle,18-action and all content assertions. No production, manifest, coverage, historical hash, runtime, timeout or test-input change is justified by this failure.

Read-only fanout audit searched every cumulative scenario index/at/slice/splice/count/history reference. This is the sole absolute indexOf(row) pin. Other scenarios[0] uses inspect frozen objects, a local RU fixture, or intentionally target the newly inserted first row. ID-based row lookups remain valid. Historical reconstruction explicitly removes minimum row first; exact old13-object/byte reconstruction and prior cross-carrier preservation callbacks passed in this CI run. This supports the narrow repair, not a prediction that future CI must pass.

Proposed exact focused callbacks, all in transactional-cumulative.test.ts:
- pins the one-page matched-plus-updated witness without advertised-limit or other-outcome credit
- binds the canonical cumulative TX selection and exact source digests without qualification
- pins the inherited command-minimum manifest facts, read-header absence and all thirteen prior objects
- preserves every original TX scenario and coverage object while adding one bounded partial cross-carrier row

These four callbacks perform static JSON/schema/bundle assertions with0HTTP/listeners/store/product-child workload; ordinary module setup still occurs. No focused run performed by this reviewer.

## Actual gates and timings

Job started08:00:59Z, completed08:17:13Z:974s. Pnpm test step08:04:20–08:17:09:769s metadata; Vitest duration768.20s. Summary:1failed/9289passed/1skipped (9291total),1failed/172passed files (173total). The official summary is authoritative; simplistic pass-line counting misses differently formatted entries.

Setup through start of tests consumed201s, including pinned bd Go install159s (08:01:27–08:04:06). Format, lint, boundaries, typecheck and fresh build succeeded. Test failure exits1. E2E ready preflight, standalone installed smoke and Read-cohort evidence gate are skipped, not passed. Local earlier Read seal/local installed results are separate evidence and do not fill these CI step outcomes.

Workflow's current job limit is20minutes (1200s). At terminal job completion nominal remaining budget was226s, but downstream preflight/smoke/Read gates did not execute. Thus226s is unused failed-run headroom, not proven successful-run margin or permission to grow future workload. Relative to root's prior750.80s suite observation, current768.20s is17.40s longer; no CPU/runner-cause attribution or per-case prediction follows.

Slow tail remains transactional-matrix repeated full observations:08:15:20–08:17:06 contains semantic conflict/deletion/cross-carrier controls around13s each, ending with the two new selected minimum controls1.299s and1.219s. The new full14-row positive passed20.913s; RU full partial matrix passed75.168s. The sole failing static test itself took11ms and was observed early; its failure did not shorten the rest of the suite. Future sequence evidence planning should preserve existing command/test limits and explicitly budget any repeated matrix work rather than assuming the failed-run remainder is guaranteed margin.

CI also passed the two helper callbacks (collection10 case1.079s, Snapshot2 case1.740s) and new command controls. These are actual CI emitted-workspace observations on synthetic merge, not substitutes for independently bound local fresh-consumer provenance. The skipped standalone smoke remains skipped.

## Landing verifier review

Read verify-landing.py without execution. Cached actual job.check_run_url confirms check-run ID116314371322. Its strict premerge/postmerge checks are coherent for the specified ordinary two-parent merge, fail closed on moved base/tree/head or non-success checks, and preserve exclusive verification outputs. This failed run correctly cannot satisfy its green-CI prerequisite.
