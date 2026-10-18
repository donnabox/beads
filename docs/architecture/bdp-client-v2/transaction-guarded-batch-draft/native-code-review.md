# Native code review — guarded-batch draft

No actionable correctness, oracle, privacy or lifecycle finding in the frozen two-path draft. This is an independent source review, not an execution verdict or permission to publish.

## Exact target

- Workspace: /Users/dbox/repos/wt/vickie/vickie-guarded-batch.
- Actual HEAD: 9bf601bfb570adae97ce8cdc99caf2cc7a2b34c1; tree 592c5e6e5c0c790b6e2690495b94ac48705b9069.
- Source binding: 1ec4f7a37f765d086be85deb47441ec9c6ca82400507e78c0f1ce0202592cf87.
- Complete patch: 88eeab067e096b565c6cd372c44e1a56a23acabd1fb813a8e1bc234939665e6c.
- New test, all 1,095 lines reviewed: f716b51eb31226beafaba80a049063d758ed803388e0ac0439ee0298e0d5019b (43,176 bytes).
- Progress document: 90c370e47d9f2dbff040f757c6b8bb89f6a6eda7b540961ea377f161d5a09152.
- Report contract: 33a9a7183cb2d309eec3a57257e634cdb744829de14400d163f5ba8788d1482c; selection: 6cfa676d0a7ce7c3da779e8e9de3403a0e3e34ed4d06dd6d3e2aa4f6d907be1a.
- Accepted proposed-successor-v2 plan c2c4642c4fb8d5f7266c00a3b3e2cb4af198e34c93c468165d3ccbdc865a0415 and ledger 5aa8f75096d1f9eb3b1dc8ab3244dd012d64bf5e843f7738e820dc621f77a1e0, including complete supplement dispositions, were read.

Final read-only hash checks match both owned files. Git status contains only the document modification and new test. Independently reconstructing both complete current files from the retained patch and HEAD originals reproduced their exact hashes; the untracked test was not omitted by relying on ordinary git diff. No product file was changed by this reviewer.

## Substantive checks

1. **Receipt and failed-transaction semantics.** Test lines 611–718 schema-check the complete receipt before extracting values, require HTTP200, own key, available detail, current tuple and exact full member set. The failed branch pins the five-member revision-mismatch Problem, required pos-2 and absence of completed-result/effect fields. This agrees with resource-evaluator.ts:139–159, transaction-batch.ts:749–757 and transaction-store.ts:1262–1307. HTTP receiptResponse at transaction-http.ts:176–190 changes only pending mutation receipts to202. The test does not mistake the embedded409 for an HTTP409.

2. **Real guard and replay inputs.** The closed construction and independently pinned reference literals differ only by the two expectedRevision omissions (test:41–62,307–334). The actual seed revision comes from its validated full Resource; the reference UUID is never sent. The 128-byte JSON-string bound precedes contender dispatch (725–744). Original losing key/body/hash are retained and checked before replay (806–824); replay and root recovery require full receipt equality, including expiry (825–834). No extra dispatch or retry loop exists.

3. **Rollback and Event completeness.** The successful Resources are exact five-field objects, including metadata, and the new seed revision differs from the seed. This matches materializeVersion in resource-evaluator.ts:1125–1169 for this fixture. Complete collection equality (839–845) preserves canonical order from authority-read.ts:887 rather than weakening to identity sets. Event equality (850–906) includes all three complete objects and next:null; the updated Event correctly uses previousRevision/revision/change, matching transaction-batch.ts:790–809. Created data includes metadata. IDs use the private renderer's tagged hashes with each group's actual position/ordinal (transaction-events.ts:76–99), explicitly without claiming normative public ID spelling. No losing marker/Event fits the complete expected results.

4. **Control sensitivity and prerequisites.** Both guards are omitted, so either arrival order gives two effective updates with distinct titles. Before the named failure, test:782–803 requires two full schema-valid successful receipts, distinct receipt/transaction identities, distinct updated revisions, unordered effects2/3, bounded current positions, exactly three completed responses/connections, and dispatch prerequisites. Only then does the one-retained-revision-failure assertion require actual0 to equal1. The control cannot continue to requests4–7. A generic exit1 or early schema/transport/cleanup error remains unacceptable under the report contract.

5. **Transport and lifetime.** send checks sequential index and mode-specific request caps, grants only fixed routes or the validated issued receipt root, dials the helper-validated loopback using canonical Host, and uses agent:false with Connection:close. There is no redirect following. Request/header checks precede dispatch; response body overflow rejects before retention (521–590); fatal decoding follows complete nonempty response. Both pair promises receive immediate rejection handling. Closure promises are awaited before dependent sends. Dispatch-before-settlement and distinct sockets are asserted; connection/write-before-response observations are recorded without inventing server admission overlap. Work/exchange/drain/cleanup/test bounds remain45/5/5/10/60 seconds.

6. **Failure reporting and cleanup.** Finally aborts and destroys remaining owned work, observes pending request/close failures, races cleanup against the retained hard bound, and makes timeout, residual sockets, unknown pre-ready stop or unclean child exit nonpassing (919–979). Unknown pre-ready ownership preserves the store. The helper memoizes close and uses graceful IPC before bounded fallback; a local timeout is not proof of process absence. Both report overflow branches add a failure; report-write failures also survive to the final throw (1057–1095). Report serialization allowlists source/binding identities, safe response fingerprints and validated receipt facts, not token/configuration/arbitrary errors. The optional file snapshot can precede a later report failure, so external adjudication must continue to require actual exit, final stdout and report flags.

## Scope and remaining evidence

Exactly one callback has a seven-request default positive and separate three-request control: 10 HTTP maximum and two scheduled product children across local cohorts. Positive request bytes are at most1,585, control818, combined2,403; retained-response sums are not heap claims. The doc is explicitly a historical unexecuted draft checkpoint and adds no catalog or public-profile credit.

Root separately reported eight static/build stages completed while this review ran. This report does not independently adjudicate those executions, claim a positive/control run, or substitute for actual Claude council, fresh source/build/archive/consumer provenance, runtime prerequisite adjudication and independent PID/store absence. No test, build, product import, service, installed invocation or review CLI was run by this reviewer.
