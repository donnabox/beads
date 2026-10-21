# PR126 successor CI outcome

Verdict: exact successor CI succeeded. Run38705226698/job116322877081, head01a247b2283f70890c8c35ec65a43aa3c3ab0542, completed2026-10-15T08:39:14Z. All job steps succeeded, including full suite, ready preflight, installed smoke and Read evidence gate. No retry, source edit or outward mutation by reviewer.

## Source and merge binding

Verified clean ordinary chain453b6dab386746bb03aac2c0a293572b8a71c969 → a912c23fbd80e70ead382a27e700a70123bf3f21 → 01a247b2283f70890c8c35ec65a43aa3c3ab0542. All12 source hashes match the accepted successor; parent diff remains the exact one-line fdd1944a044a091ebe0bf0c81a88376e7c0dfc56b163064a9f6999333164d647 repair.

Synthetic merge d8431542eacbd815ee785fcaa22e0b3e1534cb63 has parents[453b6dab386746bb03aac2c0a293572b8a71c969,01a247b2283f70890c8c35ec65a43aa3c3ab0542], treeb2d9f2c18502e0778c51123be8b4564af052d803 identical to candidate. Raw log records fetching that merge. Initial capture ci-initial.json SHAa0541f4978a834d1dbbfa9fab5187da82798bfc7a5e26385875e1961f865237a. Actual merge/publication remains root-owned; no merged-state claim here.

## Complete evidence retained

Raw actions/jobs/116322877081/logs response saved as ci-full.log,3294728bytes, SHA256f187f3c3f2388608b0133b6dfd3f48bcb34b285ce706b0b303d0d348f9e20fdb. This is complete raw job output, not truncated gh run view text. Terminal run/job/annotations retained in ci-terminal.json SHA2561ae0972429ec07231365eafa24ab54b2e68a80cb35ac6775bc9f739a30c7bd12. Original failed run remains untouched in ../landing.

Full suite:9290passed,1skipped,0failed,173files passed. Vitest duration771.73s; repaired pagination-order callback passed9ms (raw line2456). Ready preflight:7passed/1file,1.44s. Installed smoke actually ran: installed help/version/diagnostics and clean lifecycle for bdp/bdptest/bdpbd; configuration round-trips; fail-closed serve/profile checks; public Read discovery; installed.reference.observed reports allNineTargets:true and claimEligible:false. This smoke is not public Transactional qualification.

Read-cohort evidence gate emitted evidence.verified,89rows; historical runHead e5f7d0c43c151426350ceededdc77fe7e372db9b and evidenceCommit dbaf63a24219e1ea3f80394746853598b0d4464a remain historical, not replaced with current head. Format and lint succeeded with disclosed one oversized-file format warning,765lintwarnings and48infos.

## Actual timings and limits

Job08:22:47–08:39:14 =987s (16m27s), leaving213s under the existing20minute limit on this actual run. Setup through test start08:26:08 consumed201s; pinned bd Go install08:23:12–08:25:54 consumed162s. Typecheck4s/build3s metadata. Full-test step08:26:08–08:39:00 =772s; preflight2s, smoke9s, Read gate within the same second; cleanup completed thereafter.

Successful full suite771.73s is3.53s longer than original failed768.20s. This small comparison does not establish per-case performance bounds or runner-cause attribution.213s is observed whole-job margin on this run, not a guaranteed allowance for future matrix amplification. Existing runtime/test/workflow limits unchanged.

The successor verify-landing.py was read without execution and correctly distinguishes ordinary candidate parent a912 from actual merge parents[453b,01a], requires this exact successful run/job, and rejects base/head/tree drift. Root still performs final premerge/postmerge identity and local-installed evidence checks. CI success alone does not grant whole-profile/public qualification.
