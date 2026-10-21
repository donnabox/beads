# Original PR135 CI outcome — success, original source only

Run38807257217 attempt1 / job116671188007 completed **success** for original candidate9e5f040edab88be364f683098ead72f595f0ab55. All27 actual job steps succeeded, including post steps. This result does not cover local repair successor8833c9242b7788e843e9eb97a2c99b2277fe75ef and is not merge authorization.

## Identity and scope

The complete raw checkout log prints synthetic2d6ba4dd5e65e7f049acbbe22aef30b430f122d4. Its independently fetched immutable Git object has ordered parents[9bf601bfb570adae97ce8cdc99caf2cc7a2b34c1,9e5f040edab88be364f683098ead72f595f0ab55] and tree81c6786a3075f1869f54460880cac1234c27fb77, exactly the original candidate tree. The fetched candidate object has sole parent9bf601. Run and job head_sha both equal9e5.

Committed base-to-candidate scope contains exactly two files: installed test SHA f716b51eb31226beafaba80a049063d758ed803388e0ac0439ee0298e0d5019b and progress document SHA90c370e47d9f2dbff040f757c6b8bb89f6a6eda7b540961ea377f161d5a09152. Raw binary parent-to-candidate patch SHA baa7be39ac7d4fb6635287126194362633b555d921be7a821550e36fbe506d71 is retained. Current local HEAD8833 is recorded only as surrounding context; all source checks above use immutable9e5 objects. The captured PR remained open/draft with original head9e5/base9bf.

## Actual results

- Main suite:181 files passed;9312 tests passed,1 skipped,9313 total. Duration524.22s (reported aggregate tests1194.75s reflects parallel test time, not wall clock). No failed test/file summary or error annotation.
- Sole skip: scripts/demo-read.test.mjs, runnable Read walkthrough > matches real bd and preserves the established seed command recipe.
- New guarded callback passed884ms. Its genuine CI stdout observation reports7 attempts/7 completed responses, passed, zero primary/cleanup/total failures, childExitCode0, closedtrue and directoryAbsenttrue. It labels execution workspace-emitted and sourceCommit as the actual synthetic2d6 commit. This is not fresh external-consumer evidence and not independent remote PID attestation.
- E2E-ready preflight:1 file,7 tests passed;998ms.
- Smoke succeeded; actual installed.reference.observed records allNineTargets:true, stableScope https://reference.invalid/acme/, claimEligible:false; role isolation check also completed.
- Read evidence step succeeded with actual evidence.verified,89 rows,47 sealed definitions/46 reachable. This is existing Read evidence, not Transactional qualification.
- Formatting succeeded with1 warning; lint succeeded with851 warnings and54 infos. A separate final Actions annotation warns that pinned actions target deprecated Node20 but were forced to run on Node24. One warning annotation, zero error annotations; no claim of warning-free CI.

Job metadata spans21:06:09–21:17:46 UTC (697s), leaving503s relative to its20-minute cap; no future throughput guarantee. pnpm test step spans21:08:48–21:17:33 (525s). Go tool installation124s is the largest setup step. Preflight2s, smoke5s, and Read gate/post steps all completed.

## Retained capture and limits

full.log was obtained directly with gh api repos/gastownhall/bdp/actions/jobs/116671188007/logs after terminal metadata;3,419,719 bytes, SHA2ec602219519bd816b1b98720c9ae059c61c9cb971e8e98390030df25348fa22. It is the complete raw job output, not gh run view truncation. capture.json retains exact argv/exit/output hashes, with run/job/check/PR/candidate/synthetic/base metadata and stderr files. scope-verification.json records immutable identity checks.

Only read-only API/git/file inspection and guarded external artifact writes were performed. No fetch, repository mutation, product execution, retry, push, readiness or merge action. Original successful CI remains distinct from the F1/F2 repair's required exact-head CI and root council.
