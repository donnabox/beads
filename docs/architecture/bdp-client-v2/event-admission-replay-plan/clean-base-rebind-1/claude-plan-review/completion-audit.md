# Actual Event PLAN terminal provenance audit

Completed one actual review: raw exit0, review_complete, 1091.769s. Independently parsed 4409 records/2406613 bytes; exactly one successful non-error terminal result maps byte-for-byte to final-review.txt. Raw SHA256 `ecd9959dd2664372430c005cb378f9cd51964e1b4803292787a5cb0bc6582810`; final SHA256 `62e71bf378ef230ce32181bd02f219bf7869c8e41e0c6790064ff1109daf260c`. Empty stderr, zero stream parse errors and no wrapper error. Completion is not approval; findings are undispositioned here.

Verified all29 prepared entries with the original indexed binding mapped to binding.closed.json. Root release authorizes only rootLaunchReleased false→true; released hash and native/root mappings match. All679 current source files,19 copied contexts and11 artifact/tool pins match; source-before equals source-after exactly. HEAD9f89271e remains clean. No bound file or source was changed by this audit.

Actual tool invocations: {"Glob": 1, "Grep": 27, "Read": 40}. Error-marked tool results: 0; unmatched calls: 0. Successful Read paths: 15; copied context Read paths: 0. The reviewer explicitly disclosed copied contexts were not read, using inline plan/selection and actual repository/spec instead. There are19 bound copied contexts plus two packet metadata paths listed in the prompt; its prose reference to “20” contexts is not an exhaustive access accounting. This audit verifies their provenance independently, not Claude endorsement. Tool-access/range details are in the JSON; no full-file read claim is inferred from partial requests.

Successful actual Read paths:
- /Users/dbox/repos/wt/vickie/vickie-current-observation-minimum/apps/bdptest/src/transactional-event-sse-installed.test.ts
- /Users/dbox/repos/wt/vickie/vickie-current-observation-minimum/docs/specs/bdp.md
- /Users/dbox/repos/wt/vickie/vickie-current-observation-minimum/packages/server/src/read-http.ts
- /Users/dbox/repos/wt/vickie/vickie-current-observation-minimum/packages/server/src/transaction-consistency.ts
- /Users/dbox/repos/wt/vickie/vickie-current-observation-minimum/packages/server/src/transaction-event-sse.test.ts
- /Users/dbox/repos/wt/vickie/vickie-current-observation-minimum/packages/server/src/transaction-events.test.ts
- /Users/dbox/repos/wt/vickie/vickie-current-observation-minimum/packages/server/src/transaction-events.ts
- /Users/dbox/repos/wt/vickie/vickie-current-observation-minimum/packages/server/src/transaction-http.test.ts
- /Users/dbox/repos/wt/vickie/vickie-current-observation-minimum/packages/server/src/transaction-http.ts
- /Users/dbox/repos/wt/vickie/vickie-current-observation-minimum/packages/server/src/transaction-runtime.ts
- /Users/dbox/repos/wt/vickie/vickie-current-observation-minimum/packages/server/src/transaction-service.test.ts
- /Users/dbox/repos/wt/vickie/vickie-current-observation-minimum/packages/server/src/transaction-service.ts
- /Users/dbox/repos/wt/vickie/vickie-current-observation-minimum/packages/server/src/transaction-store.ts
- /Users/dbox/repos/wt/vickie/vickie-current-observation-minimum/packages/server/test-support/raw-sse-recorder.ts
- /Users/dbox/repos/wt/vickie/vickie-current-observation-minimum/vitest.config.ts

PID/PGID92109 and every sampled PID independently absent. Runner reported no cleanup intervention or survivors; current process snapshot agrees. Same1800s/TERM1790 bounds/default-model arguments, only Read/Glob/Grep, no retry. Root owns the numbered findings and next action; no authoring/runtime release is implied.
