# Fresh Snapshot minimum repair — native closure

## Conclusion

**No actionable findings. All four accepted Low repairs are closed.** The exact delta changes two tests, restores a production comment without executable logic changes, and clarifies current-context prose. Optional external items4–7/9 remain deferred as root directed; this closure does not reopen retained Snapshot semantics, add cancellation behavior, or claim public readiness.

## Exact source and findings

Reviewed clean frozen HEAD `0fafcf19b0fdc6476f50a5c5fccb176a7e6af22d`, tree `b2f89df7b7ad007ad3f6a938b313d0c0a58700c4`, direct parent `587a7eeaa9663ab3d516a45375b3dbf0e9fedca9`, at `/Users/dbox/repos/wt/vickie/vickie-tx-snapshot-minimum`. Read completed external code review, prior native report, exact four-file diff and repair report/summary/test outputs. Frozen source remained untouched and clean.

1. **GET branch spy:** `transaction-service.test.ts:1039–1073` now includes ordinary GET with expected `{kind:"create"}`, alongside HEAD/304/412/406 preflight cases. Both satisfiable genesis and fixed-head future bound assert exactly one service call, correct first request kind, actual status and pos-0. This discriminates inserting a preflight before direct capture without altering the production path.
2. **Cancellation limitation:** `transaction-service.ts:257–259` explicitly states the unchanged no-minimum synchronous owner path does not observe the signal or promise interruption of retained reads/copying. No new pre-abort check or behavior was introduced.
3. **Retry-After:** `transaction-http.test.ts:2073` checks exact `Retry-After: 1` on every future-bound row of the focused matrix, including HEAD and conditional/media candidates. Existing exact catch-up-timeout Problem, current-context and bodylessness assertions remain.
4. **Context wording:** `transaction-implementation-progress.md:677` includes406 and minimum refusals in the fresh complete-tuple context statement. Adjacent text continues to identify retained/malformed/partial forms as explicit private gaps and disclaims synchronous interruption/public qualification.

## Independent checks

Archived exact source into `transaction-snapshot-creation-minimum-repair-native/source`; copied server emitted output from the repair's retained installation and reused existing dependencies. No author build/output writes or package install.

Pinned Node24.16.0 command from that disposable archive:

```text
node node_modules/vitest/vitest.mjs run packages/server/src/transaction-service.test.ts packages/server/src/transaction-http.test.ts -t 'uses one service call for each fresh HTTP branch|serves complete fresh Snapshot minima' --maxWorkers=1 --reporter=verbose --reporter=json --outputFile.json=/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-snapshot-creation-minimum-repair-native/focused.json
```

**Exit0:2 passed,0 failed,97 unselected/skipped across2 files.** Full output retained in `focused.log` and `focused.json`. This checks the two changed cases directly; it is not an independent rerun of99 tests or the installed cohort.

Independently compared actual old/new installed `dist/transaction-service.js`. SHA256s are `c36ec829d9853b6308cb73cfea01253aac0f537bea21b211a4515a871a8c691a` and `775928ae0a4c40a1b8f26ed7915908d3b18f63855b07f246019c59db9de1c41a`. Removing exactly the two added comment lines reproduces the old module byte-for-byte. Thus the changed emitted/payload bytes are real, but do not represent a changed executable branch; this conclusion does not rest solely on passing tests.

## Fresh author evidence inspected

`transaction-snapshot-creation-minimum-review-repair/packed-summary.json` SHA256 `e4fb08405b2603c39559e66fa7bc48078c41ed3235845388e31066c592f0badb` binds exact0faf/source tree and clean status. Author focused JSON records99 passed/0 failed. Fresh installed creation JSON records1 passed/0 failed,9 filtered cases;31 actual completed exchanges/two child sessions. This is newly bound0faf evidence, not relabeled587 results.

Independently recomputed package hashes against retained tarballs:

- protocol `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`;
- server `84c3c73fff5d8156fa7c181a0dd68ae60066c2c499dec3c3bcdf3e5374b1be34`.

Observed installed payload is `13a24e64340be538f74de8c2dbe77f7b8ee20bffb264cb52bb09f9537072c172`; its difference from587 is honestly recorded, with the exact comment-only module comparison above. The report retains actual config/contracts/fixture/route bindings. Child PIDs49249/49255 exited0 without signals and were absent when independently checked. No review-owned process remains.

The broader retained-page89, capacity39 and derivative2 exchange observations remain historical at587/the distinct55d9 mutant. They were not rerun or copied to0faf. No broad suite, duplicate packed run, source edits, outward writes, merge or public-profile claim was made by this reviewer. Root owns ordinary alias/main integration and final validation.
