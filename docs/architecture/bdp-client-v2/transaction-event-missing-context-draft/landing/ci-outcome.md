# PR127 exact-head CI outcome

**Failed; landing gate is not satisfied.** Run38728209051/job116399188531 completed2026-10-15T11:49:45Z with failure. The sole test failure was a5000ms per-test timeout in `apps/bdptest/src/development-contracts.test.ts:628`, “keeps an actual evaluator work refusal local and permits correction using the same unclaimed key.” It was not the20minute job deadline. No retry, cancellation, budget change, source edit or outward mutation was performed.

## Exact identity and retained evidence

Terminal metadata and a fresh local read confirm clean head `3df82ac480f50c8f486b4e456336366e7acbdf87`, tree `d5f02c936cf080666a8e73bb5f253e52086705f4`, with the sole ordinary parent `88c9408e908590ab1a4c64c9b9dee1a8b705999c`. All six source hashes still equal the accepted redirect-repair binding acf190e742dfb669fe5607085a3ac60c5ecb66f4f86b1f05025424a7873615b0; the parent diff contains exactly those six paths. PR127 remained draft on that exact head/base at capture.

Synthetic commit `33580067aa88dbc1f0b5f15e2c2b9af739d9a200` has ordered parents[88c9408e,3df82ac] and the exact candidate tree d5f02c936cf080666a8e73bb5f253e52086705f4. The full log records its fetch and checkout at lines94/401/429. Run metadata is bound to3df; CI's actual checkout and source labels are the synthetic commit, not silently relabeled as the ordinary candidate.

Files retained here:
- `ci-initial.json`: SHA-256 `0aff960df38e44515d0457cd2096d328a59ecbe1482989c3e924e9977db83421`.
- `ci-terminal.json`: actual run/job/PR/synthetic/annotation metadata, SHA-256 `fc02b17d1bcbded783025eb3bb1e923d11049f580b8e71b31a2e31fdf2ce4e54`.
- `ci-full.log`: complete raw `actions/jobs/116399188531/logs` response,3,308,540bytes, SHA-256 `f64b71325085156fc9c4d2079c5b9bb07873a2e7d536a57565b8c697301b363b`. No truncated `gh run view --log` output was substituted.

The terminal failure annotation says process exit1. No exceeded-job-deadline annotation is present. Other annotations concern action Node20 deprecation and the prospective ubuntu-latest image migration; they do not establish the timeout's cause.

## Actual suite and gates

Full Vitest summary at raw lines11098–11101:174files (173passed/1failed),9297tests (9295passed/1failed/1skipped), duration755.84s. The failure is first reported at5518–5520 as5029ms, then at11084–11094 as “Test timed out in5000ms,” pointing to the callback declaration at628. The log does not identify the awaited operation that failed to settle, establish a CPU cause, or establish a flaky test. That test file is unchanged by this candidate.

The one skipped callback is `scripts/demo-read.test.mjs > runnable Read walkthrough > matches real bd and preserves the established seed command recipe` (line10434). It remains skipped; passing surrounding cases do not make it executed.

All six newly registered Event case expansions passed:
- finite Bead/Link immutable missing-context cases:8ms each (lines4110–4111).
- SSE Bead/Link immutable missing-context cases:8ms/7ms (4112–4113).
- contextual missing Event HTTP case:486ms (4117).
- installed missing Event context case:1116ms (7146).

The source and installed console records at4115/7144 are present, but Actions redaction corrupts the WWW-Authenticate string into invalid JSON around `***`. They must not be treated as successfully parsed complete CI response records. The immutable raw log is preserved without repair. The passing Vitest case rows are valid test outcome evidence; the separately audited local fresh-consumer report remains its own valid JSON/provenance evidence. The CI case uses the built workspace entry and synthetic source; it is not a second independently packaged-consumer witness.

Setup, frozen dependency install, format, lint, boundaries, typecheck and build succeeded. Format reported one warning; lint reported765warnings and48infos (raw902/1300–1301). Ready preflight, installed smoke and Read-cohort evidence gate were all skipped after the test failure. No successful Read seal or smoke is attributed to this run. Setup-go/setup-node post steps were skipped; pnpm/checkout post cleanup and Complete job steps succeeded.

## Timings and next gate

Job11:33:45–11:49:45 =960s (16minutes). Setup to test-step start11:37:03 took198s, including158s Go install. Test step11:37:03–11:49:40 took757s, consistent with the755.84s Vitest duration. Metadata records format1s, lint2s, boundaries2s, typecheck3s, build3s; post-failure settling through job completion took5s. The failure ended240s before the configured20minute maximum, but skipped downstream gates mean this is not a demonstrated full-green headroom measurement.

Compared only as aggregate observations, this755.84s suite is15.89s shorter than the prior PR126 successor771.73s suite; it proves neither per-case throughput nor a cause for this isolated5000ms timeout. The first actionable evidence is the exact failing callback, not a reason to change workflow/runtime bounds or call the failure resolved.

Root must disposition/diagnose this failure and obtain the required exact-head successful gates before landing. Existing local installed success and passing new Event cases do not override the failed CI gate. Historical source-selector/control failures and local evidence remain untouched; no merge or public qualification is claimed.
