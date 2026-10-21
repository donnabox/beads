# PR90 latest-main integration validation

Frozen source: `419756cd01dc8b4798b900369c510513026896c7` on `codex/vickie-changefeed-sse-publication`, checkout `/Users/dbox/repos/wt/vickie/vickie-changefeed-sse-publication`. Clean ordinary merge, with parents `15c2641053609081b2f2ff63cd734770b64f7859` and remote-verified main `47d7580ae6b42be2fb31f5435d408aba615021d6`. No conflicts, semantic edits, outward writes, or source changes after freezing.

## Preservation

The five incoming evidence files equal main exactly. All 49 authored scenarios, including reference A/B, equal main; the earlier 47 raw scenario blocks remain byte-identical to the runtime parent. Server, protocol, client, and scripts trees equal `15c2641` exactly. The merge audit is in [merge-audit.json](merge-audit.json). Main tree was `f14ca53d69c7d1ec9d5a44cc8bcf7be4cf1da5b3`; merged tree is `63f3fc91c0bc0b833940e3aad61a8f34e22d30be`.

## Exact-source validation

Node 24.16.0 and pnpm 11.20.0 were pinned. Frozen dependency installation, pre-build typecheck, clean build, boundaries, formatting, lint, both smokes, seven preflight tests, and unchanged Read evidence verification all exited 0. No evidence seal was regenerated.

| Execution at this source | Passed | Files | Failed/skipped |
| --- | ---: | ---: | ---: |
| Cumulative artifact checks | 6 | 1 | 0/0 |
| Affected RU matrix | 138 | 1 | 0/0 |
| Fresh installed RU checks | 173 | 4 | 0/0 |

The affected matrix took 97.98 seconds under supervision; the fresh installed run took 97.02 seconds. All stages completed naturally with exit 0. Stage commands, elapsed times, JSON results, and logs are retained beside this report. Final source status was clean and no owned test/server child remained.

The full 8,656-test suite plus one skip was executed at `bcfa43f04235ee04e0ca26902dc63c243d8495d0`, not at this source. The 164 affected server tests, fresh installed SSE test, and 30 finite TX/Snapshot tests were executed at `15c2641053609081b2f2ff63cd734770b64f7859`; their exact report is [the runtime-parent report](../changefeed-sse-final-integration/report.md). Those results are retained as ancestor evidence supported by runtime tree equality, not relabeled as executions at this merge. The requested evidence-only integration did not repeat the full server suite.

## Fresh packed evidence and limits

Fresh protocol/server archives were installed in an isolated consumer. SHA-256:

- Protocol archive: `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`.
- Server archive: `965cbc5779489634355fb804b3c8b2ebbfe599041e421a4dd82fec8deed22be7`.
- Observed installed payload: `c3fe7cf4cc30154d57686de8b042886193b419a48b3b2af8fb8139480024fde8`.
- Manifest: `c86b33535f25a1d117ce021a3bb70164f5635bdaee50758e3428c130ccc3227f`.
- Fixture: `f5f38235dfd28cf06e760688e2054d3e29d8d64b08c4c729baa94f380f1f7a63`.
- Exact executor configuration: `3224c31c83988dc0606cb76ee1f4ba8e6394159bdf7daaccc301461316cc2477`.

[summary.json](summary.json) records complete safe source/configuration/artifact bindings. [The actual packed progress record](packed/ru-progress.json) preserves the 49 physical sessions and observed requests/responses: 985 client observations, 21 native Read observations, and 44 setup observations. SourceCommit is launcher-verified source provenance; actual bootstrap identities bind installed payload, policy, authentication, and contracts.

Required denominator remains 129, with 49 authored passing scenarios, 30 whole contracts and 99 uncovered obligations. The report's 85 harness-error entries represent unimplemented scenarios; they remain failures/uncovered, not fabricated passes or not-applicable outcomes. These counts measure different sets and are not additive. `claimEligible` remains false. No public profile admission, promotion, schema, catalog, or protocol semantics changed. Root owns publication and any subsequent exact CI decision.
