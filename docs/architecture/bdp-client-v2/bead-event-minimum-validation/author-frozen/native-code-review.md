# Native code review — frozen Bead Event minimum draft

## Findings

No actionable blocking defect found by this read-only review. The exact six-file draft implements the accepted plan-v3 and bounded dispositions. This is a native review finding set, not an execution result or an authoring/integration/validation release. Static checks, actual Claude code disposition, fresh build and source/installed validation remain required; formatter/typecheck/lint success is not inferred from inspection.

## Frozen identity and preservation

Checkout /Users/dbox/repos/wt/vickie/vickie-bead-event-minimum remains based on f6204255d53ef65d4aea56ff5ccc86d72e8cc7b0. Independently verified source-binding SHA25665cc7b8f38c03d43eabb8c76d3087a8c74b30ca37f8037989c90454acd7778a5 and full.patch33c7e00400c1923c4a16dadc03040cf5ae3f773c926a86396defd305d45c4ece; actual git diff --binary HEAD matches the frozen patch byte-for-byte. All679 current source hashes match the after map; exactly the six owned files differ, leaving673 unchanged. Six before snapshots match Git HEAD and six after snapshots match source. All ten context inputs rehash correctly, including plan57bf4272, selectionc08e9a2a and root release.

The installed file preserves all25284 historical bytes exactly. Removing the single new error import restores the entire old service test prefix. Reversing only the accepted four Event-test substitutions (comment, direct Bead oracle, Scope title, Bead HTTP oracle) reconstructs the exact old Event-test prefix. Production changes are confined to the HTTP minimum predicate, service classifier/import and owner initial-minimum predicate/comments. Parser, runtime, waiter, schema, catalog and preserved G2 implementation/tests are unchanged.

Final source SHA256:
- packages/server/src/transaction-http.ts:37b8ed7a63c25d310159edba83d7ecc6f0292724831c2b59643e7a3788835589
- packages/server/src/transaction-service.ts:cede2090af8b4cb4470334ff27e6f5038ab4cb186972ed2bd03a300ae19a6ba6
- packages/server/src/transaction-store.ts:3e7b9b4609602ff5c207e697bdf0d70a5c4f6e94515130b926e83eb6a8c093b7
- packages/server/src/transaction-events.test.ts:d6dd08a07355bca304c50909bce61b1744ae22cec8b1bba56c2638442439f36f
- packages/server/src/transaction-service.test.ts:1296229ecaad736390882d64b120b433fc8d84d40f022f46803a8574c99ac828
- apps/bdptest/src/transactional-event-missing-context-installed.test.ts:e4510bf2dfdaacac981bf9583a2f6be4c536de5ceca5309c754f93fd56b6f652

## Production review

HTTP503 selection remains limited to parsed initial Scope/Bead JSON routes without URL after, parsed after or normalized/raw Last-Event-ID. Other media, Link, after and SSE keep their previous path. Valid HTTP-selected Bead sources are also owner-selected, preventing a strict opaque-token parser failure between those layers.

Service classification runs only in the first existing waiter attempt, after policy/clock/abort preflight. It uses the existing Event parser, requires Bead kind/no parsed after/exact canonical source equality, and catches only parser/URL classification failures to retain original owner error order. Lexical tuple capture is outside the catch, preventing swallowed malformed-tuples. The captured flag and detached tuple persist across retries; owner access does not survive an await. No duplicated waiter policy or new public interface is introduced.

Owner selection is initial Scope-or-Bead only. Lexical capture remains before entry; unknown/hidden retained identity resolution remains before sentinel and semantic minimum. Current context and minimum comparison remain in the same owner entry before retained group reads/projection. Deleted retained identities remain known; replacing authorization policy hides the source before semantic foreign-view interpretation. Existing foreign/malformed/excluded behavior and S1/G2 scope remain explicit.

## Test and lifecycle review

A has12 direct Event reads and one separate final policy replacement, two real transactions, actual current/behind/foreign/cursor/Link/unknown/deleted/hidden controls. Its single delegating SQL spy keeps the original prepare implementation, has positive scan controls, and restores in finally. The last hidden read uses the old opaque tuple and asserts the replacement view/current context with zero group reads.

B/C use actual completed service seed receipts before scan observation, unchanged real-owner probes and deterministic10ms schedules. Caller mutation cannot change the captured tuple. Cancellation releases its gate before awaiting; drain begins afterward, rejects new calls synchronously, releases both remaining waits, and observes close. The preserved finish helper releases gates, observes pending promises and closes before restoring scans. Expected4/5 attempts,2/3 waits/probes and separate post-close refusal match the code.

Seven F rows retain isolated fixtures/service cleanup and discriminate aborted, invalid policy, throwing/nonfinite clock, selected malformed tuple, noncanonical query and relative URL. Exact error identity/class/message and clock counts1/0/1/1/1/1/1 are asserted; only the two source errors reach owner forwarding, and no waits occur.

D performs37 HTTP requests, including two distinct-key real mutations/five operations, known Link exclusion and deleted-resource404 followed by retained Event200/503. The37 sum is2 setup +30 paired requests +3 delete/resource/retained setup +2 retained refusal requests. Expected status totals are200×8,503×8,500×14,400×4,404×3. Selected503 GET bodies/schema/media/context/retry/cache/validators and paired empty HEAD metadata are checked; private gaps stay bodyless. Import precedes directory allocation, runtime close precedes removal, failures aggregate, and actual elapsed/cleanup are reported. Existing default5s and80ms minimum budget stay unchanged; no success is inferred if a future supervised attempt times out.

E appends one child/store with four actual requests POST200/GET200/GET503/HEAD503, explicit route/method bounds, real credentials, exact serialized seed-body/header evidence, completed receipt, schema-valid created Event and actual pos-1 context. Refusal body/media/cache/retry/context/HEAD-length assertions are present. Callback-local binding and module provenance retain actual source HEAD/workspaceDirty and entry digest. External archive/consumer identity is a validation-gate responsibility, not established by this callback alone. Shared45s deadline,5s requests/60s callback, late-rejection observation, known exit0/null-signal close, conditional directory removal and truthful unverified ownership/failure reports preserve the accepted failure semantics.

Authored totals remain25 source/3 sequential installed callbacks,128 emitted/18 installed HTTP,28 stores/three listeners/three children. The old Scope pair change adds waits without new requests. These are inspected prospective counts, not observed passes. No dedicated G2+opaque result is claimed; no profile qualification/public admission, discovery, main integration, release or promotion follows from this draft.

No tests, formatter, builds, installs, runtime imports, services or source changes were performed for this review.
