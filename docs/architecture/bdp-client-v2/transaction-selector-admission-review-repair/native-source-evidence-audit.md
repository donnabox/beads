# Independent actual source evidence audit

1. **Medium — missing spec/profile boundary execution, not a demonstrated source defect.** The successful `check:boundaries` invokes `scripts/check-dependency-boundaries.mjs`; it checks package imports, not Markdown profile/Part dependencies or catalog citation excerpts. None of the four source cohorts invokes `profile-catalog-boundaries.test.ts`. The smallest existing selection covering those three invariants is **nine expanded cases in that one file**: the two actual-spec callbacks at lines 161 and 177, plus the seven catalog callbacks at line 275. Run these before treating the changed spec as locally validated. No source change or relaxed boundary guard is needed.

## Exact missing selection (proposed only)

Working directory: `/Users/dbox/repos/wt/vickie/vickie-selector-review-repair`.

```text
/Users/dbox/.nvm/versions/node/v24.16.0/bin/node
/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs
exec vitest run packages/conformance/src/profile-catalog-boundaries.test.ts
--maxWorkers=1
-t ' (resolves internal links without lower-profile forward dependencies|rejects model-to-Read dependencies despite their shared profile rank|audits every requirement in .+)$'
--reporter=verbose
```

The prefix shared by all nine full titles is `catalog citations respect cumulative profile boundaries `. Suffixes are:

- `resolves internal links without lower-profile forward dependencies`
- `rejects model-to-Read dependencies despite their shared profile rank`
- `audits every requirement in packages/conformance/catalog/history-v1.json`
- `audits every requirement in packages/conformance/catalog/numeric-model-v1.json`
- `audits every requirement in packages/conformance/catalog/owned-wildcard-v1.json`
- `audits every requirement in packages/conformance/catalog/read-update-v1.json`
- `audits every requirement in packages/conformance/catalog/read-v1.json`
- `audits every requirement in packages/conformance/catalog/transactional-v1.json`
- `audits every requirement in packages/conformance/development/read-update-catalog-v1.json`

Keep the existing 120-second supervision; no increased test timeout is proposed. Add the existing JSON reporter/output discipline when root prepares the invocation. This is a selection recommendation, not execution authorization.

Static footprint: normal Vitest config reads workspace manifests and loads its existing Read mock setup. The selected module imports conformance/Markdown utilities, reads the entire spec, constructs heading ownership, and enumerates the six canonical catalog filenames plus the explicit diagnostic catalog. The first two callbacks inspect all actual internal spec links, including profile and Part direction; their four existing string-replacement counterfactuals remain. The seven table expansions read their JSON catalogs, validate citations/excerpts, and assert profile ownership and citation distinctness. These callbacks do not launch a service, issue HTTP, open SQLite owners, or launch product children. The runner itself is a normal process/worker. The other twelve callbacks remain unselected. No catalog or contract-plan source changed in this slice, so the broader contract-plan runtime/readiness suite is not needed solely to close this specific boundary omission.

## Actual source evidence closure

I independently parsed all four actual Vitest JSON files and compared each passed full title with both the frozen selection inventory and the supervisor's expected titles. The results are exactly 67 parser, 6 set preparation/execution, 13 submission, and 4 HTTP cases: **90 passed**, zero failed, zero selected skips, four exit-0 commands and no timeout. Unselected skips are 0/49/31/72 respectively; they are not counted as passes.

The saved JSON and log hashes match every result record. All twelve current source files match the frozen binding, and every command's recorded before/after source map matches it. All command process groups were empty after command completion and cleanup; no cleanup signals were needed, and no temporary fixture/cache/unknown entries remained. I inspected the process snapshots and independently found supervisor child PIDs 46703, 46712, 46734 and 46741 absent; all four owned temporary roots are currently empty. The source-bound supervisor preserves unexplained residuals rather than deleting them to manufacture clean results.

I also verified all seven static command logs against their recorded hashes: offline frozen setup plus typecheck, build, format, lint, package boundaries and unchanged Read evidence gate all exited 0. These static outcomes do not substitute for the missing nine boundary cases above.

The 94 HTTP / 39 owner counts remain **static workload accounting**, not instrumented actual counters; the supervisor explicitly says so. The evidence establishes actual selected-case results and cleanup, not installed-service behavior, counterfactual discrimination, or complete profile qualification. The emitted controls and installed 25-request cohort remain separate gates.

I read the actual Claude finalization review and root-positive release. Its R1/R2 repairs are the already independently closed test-only delta; F1/F2 are recorded low no-change dispositions. Nothing in the observed positive outcomes contradicts those dispositions. Claude's static review is not relabeled as runtime evidence.

## Exact basis and limits

- Source binding: `03a861851a195d5da505f311291bee6d3cd98f63866d4b6df756e2c3ae289afd`.
- Source patch: `4837959bceec2e16358c46c11a6a2676d113bbaad49ba8b3d184ebd34a366b24`.
- Actual positive summary: `9d0afb8d9f03b1c44231ba13f788139212ea0af84cd561de5adb4fbaecce0cfa`.
- Static results: `901381995cd5c73c17afb7737a46654442055447fc51d9cbe5653f8fde661ea6`.
- Boundary test source: `3c6e825732b94607a626787ad8199d74b169d8ed823f3463e9bf1d8e8bf23de0`.
- Parent: `757aedc8008386745d59d5005498a9256d6ea76e`.

This reviewer ran only read-only file/hash/process inspection, no product import, test, build, install, service, or source edit. The nine boundary cases are not claimed to have passed. Root owns their release and the remaining control/installed gates.
