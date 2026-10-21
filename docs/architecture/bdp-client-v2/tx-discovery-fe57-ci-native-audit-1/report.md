# Exact-head CI native audit

Accepted for the retained completed CI run only. No actionable CI evidence discrepancy was found. Run **39149875661**, job **117741137845**, completed successfully at head **fe57ea2942615d5dbe770b9229d0da25f51bf1a7**. All **27** recorded steps succeeded. The original archive independently reconciles **9,505 passing tests, one skipped test, 186 passing files**, followed by seven passing readiness-preflight callbacks, smoke observations, and the historical Read evidence gate.

The original root acceptance is `78748ece4ed786f6f09d451cd2ab586253d81ddf161fcb990794a86b2c43c4f0`. This audit makes no source, readiness-status, admission, merge, or execution change.

## Exact source and publication attribution

The original checkout step records `fe57ea2942615d5dbe770b9229d0da25f51bf1a7` at member lines 416 and 428, including actual `git log -1 --format=%H` output. The retained commit receipt records tree `6813c562dc4287bda697c4b7c1d6d94bec342ba4`, sole parent `cce884ccea1db97cbc6fd7bff876d4d6091f9081`, a clean worktree, and the exact same six reviewed files. Independently reading and hashing all **679 current known source paths** reproduces `a61cd3fbf9fac68e3aad857a24a2456dd16b83048067eef85348d7091c0aa08e` and matches the earlier local source map exactly.

The earlier 52-stage local validation remains attributed to dirty `cce884ccea1db97cbc6fd7bff876d4d6091f9081` plus `eed91e5f1abc48ef92f31d92690772c8a6489a48dd9e5e07af66de9c392712ab`. Its native closure `92f6154c65b81549aa5270acc7c7754ba9a7cdcd6af5879cb0cb4417a6de9981` and root acceptance `9ee5f9ddb070c4bf83e5912002d0e150cc0b1645b623416ad01d2255af4577b0` are preserved as prior-run evidence. This CI run supplies new exact-commit execution evidence; it does not change the attribution of that local run.

The retained PR snapshot identifies **PR170**, open and draft, head `fe57ea2942615d5dbe770b9229d0da25f51bf1a7`, base `codex/vickie-current-changefeed-minimum`. It is a stacked PR; the snapshot does not show a merge into main. Root acceptance permits readiness only after its durable publication. This native audit does not perform that action or assert the later PR state.

## Original logs and counts

The **973,410-byte** original Actions ZIP has **29 members**, passes a full CRC read, and has SHA256 `fb277e175020e59c557238786cb4d7a7b77c32310357a12d96637d44003f27d9`. Every member was fully read and hashed. The retained **4,060,297-byte** raw test log is byte-for-byte identical to ZIP member `verify/19_Run pnpm test.txt`, SHA256 `5aa7853048f9be97d3dda12f902ebe5c7dfd66c651cddaa7b38b887e761a450e`.

The terminal main-suite summary appears at original member lines **10493–10496**: `186 passed (186)`, `9505 passed | 1 skipped (9506)`, duration **874.63s**. A separate finite scan counts all 9,505 actual verbose passing rows and the single skipped row, across 186 distinct passing file names. One Actions chunk begins with an internal UTF-8 BOM at line 5486; allowing that retained delimiter in the comparison reconciles the exact pass count. No log bytes were rewritten. `test-file-counts.json` retains the per-file counts.

The single skip is original line **9896**: `scripts/demo-read.test.mjs > runnable Read walkthrough > matches real bd and preserves the established seed command recipe`. Source line **106** selects that optional demo using `BDP_DEMO_BD_EXECUTABLE`; the workflow's required bd-matrix variables are separate. The skipped callback is not counted as passing or as fresh demo proof.

The first retained `gh run view --log` output is incomplete for the main test step and contains no main 9,505/186 summary. Its original **2,569,158 bytes** remain unchanged. The later original ZIP supplies the complete original run output; no tests were rerun. Flat step20/21/22 files are ANSI-stripped views of their original ZIP members, not byte-identical raw members; comparison-only removal of ANSI escapes makes each pair exact. Both representations are retained and separately hashed.

## CI and qualification boundaries

The full source workflow matches the recorded ordered steps: pinned actions and Node/pnpm checks, Go-built bd identity, frozen install, format, lint, boundaries, typecheck, build, main tests, preflight, smoke, Read gate, and post steps. Each of the 27 metadata steps is `completed/success`; `ci-step-assessment.json` retains their full timestamps and names. No independent numeric process exit status or fresh CI process census is invented from the GitHub success labels. The original complete-job log records GitHub's cleanup step; local P10 process absence belongs to the earlier local run.

Preflight has seven actual passing verbose rows and a seven-pass terminal summary. Smoke explicitly records `installed.reference.observed`, **claimEligible=false**, all nine targets, configuration identity `4216aa83dff44fe5f2430b0cde44c140b61e5bff902dc1c2725766189fdaf51c`, payload `729b744b17806cf5b66ae4834a8172420e0f56839aac1bf5a3351388def90cac`, artifact `development-schema5-e1dff401b5caafb04a4d935775d67eeefeae7bcf306d5d100b979a64eef148e7`. Smoke also preserves fail-closed public profile admission observations. It is not whole Write qualification.

The Read gate's original `evidence.verified` record has **89 rows**, seal `674982be2123b9b7aae630bf79c3f774122900cd`, historical run head `e5f7d0c43c151426350ceededdc77fe7e372db9b`, and evidence commit `dbaf63a24219e1ea3f80394746853598b0d4464a`. The gate command allows exit 2 for an unclaimed cohort, but this actual retained output specifically contains verification. Its historical identities remain historical; these 89 rows do not become new Write evidence at fe57.

## Integrity and residual limits

All **12 indexed CI files plus their self index**, **six source-publication files plus self index**, the local acceptance, and **431 indexed prior-native files** were rehashed and matched their exact durable documentation mirrors. Only bounded directly relevant local receipts are copied; the complete prior-native tree remains in place under its verified index. Original archive, test log, formatted log, root acceptance, source commit receipts, full current map, workflow and skip source are retained in this new packet with exact origins.

This is an offline audit of retained GitHub metadata and original logs, not independent re-execution or a later live PR-state query. GitHub-hosted process identities and full remote installed filesystem/member/cleanup proofs are not reconstructed. Same 679 source bytes do not combine unrelated profile configurations or runs into a final cumulative Write claim. Final combined-source/configuration evidence, missing law owners, administrative erasure/restore proof, public Write admission, and main landing retain their existing separate gates. Existing warnings in original logs remain retained; step success does not claim warning-free output.

No test, compiler, build, install, subject helper, reviewer, Git command, network query, runtime, source edit, gate or flag mutation was performed. Root owns readiness/publication decisions after this audit is independently consumed.
