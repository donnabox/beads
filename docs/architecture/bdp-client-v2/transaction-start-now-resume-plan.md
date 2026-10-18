# Scope start=now reconnect: bounded implementation plan

Plan only; native and Claude review, then root acceptance, precede implementation. The operator has explicitly ruled: for an otherwise valid Scope `changes/?start=now` SSE request with a saved `Last-Event-ID`, resume **exclusively after the saved checkpoint**; the header overrides `start=now`. This settles the old assessment’s single question. It does not change Event Source query grammar or create another product decision.

## Exact basis and ownership

Inspected `/Users/dbox/repos/wt/vickie/vickie-noop-conformance`: HEAD8d66855ec4e26d3af0b3d7d22d621e1f37005294; local origin/main5995f6b08ebd045e4343059c2da1cffd38b09b80; their trees are identical (`a1ae75b822e1d7fadd739d818d09a25fb788e380`). Author from root-assigned5995 in a new isolated checkout after approval. Read `transaction-event-start-now-assessment.md` but rechecked current source: its HTTP and SSE-test locations remain valid; service changes handling is now366–390, not old257–280. No applicable ancestor/repository AGENTS.md was found in the inspected path. No tests/build/network or source writes were performed for this plan.

The deletion evidence author owns manifest, coverage, cumulative test and Transactional matrix test. This plan touches none of those four files. Proposed ownership is eight files:

|File|Bounded change|
|---|---|
|`packages/server/src/transaction-http.ts:428–456`|Remove the two-line `start-now-last-event-id` development refusal. Preserve URL parsing, media selection, raw repeated-header rejection and existing effective-header mapping. No service/store/pump change expected.|
|`packages/server/src/transaction-sse.test.ts`|Replace obsolete500 assertion, add focused same-URL replay proof and direct new-combination boundary checks using existing fixture/recorder. Preserve existing lifecycle/finite/backpressure tests.|
|`apps/bdptest/src/transactional-sse-installed.test.ts`|Add one separate installed same-URL reconnect scenario using existing launcher, exact finite executor and raw recorder. Keep existing15-second idle scenario and configuration checks intact. Test-local fact extraction/reporting is permitted; no shared framework or recorder change.|
|`docs/specs/bdp.md#scope-changefeed`|Minimal approved precedence wording; preserve durable application obligation, explicit invalid/expired/foreign refusal and finite JSON contract.|
|`packages/conformance/catalog/transactional-v1.json`|Update only first selectedText in existing `transactional.changefeed.sse-reconnect` to the exact revised normative sentence. Preserve ID/order/other requirements and coverage status.|
|`docs/design/w1-transactional-packet.md:6325`|Align that maintained catalog row’s first selectedText; no unrelated packet rewrite.|
|`packages/conformance/src/transactional-catalog.test.ts:68–98`|Add the existing reconnect ID to the packet-mirror check’s affectedIds, guarding exact canonical/mirrored row equality; no catalog-count change.|
|`docs/design/transaction-implementation-progress.md:1337–1343`|Replace only the now-obsolete start/header gap statement with actual effective-checkpoint behavior; retain no-header anchoring and finite-header ignorance. No broad historical cleanup.|

The catalog mirror is additional minimal alignment discovered by searching the current repository, not a new conformance artifact or runner feature. No schemas, public exports, discovery admission, claim flags, sealed evidence, fixture configuration, Event Source paths, Type/policy/metadata semantics or pending minimum-header rulings are in scope. If a required production change extends beyond the HTTP guard, stop for root disposition rather than implicitly expanding ownership.

## Existing mechanism and normative edit

Current `transaction-http.ts:428–445` validates query grammar first, selects media, preserves raw header multiplicity, then rejects start/header with500 before the ordinary mapping. The retained mapping already selects `{after: fields[0]}` for exactly one selected-SSE header. `transaction-service.ts:366–390` freezes that intent and reuses its checkpoint through minimum retries; only headerless `start:now` captures a new owner head. `transaction-changefeed.ts:54–105` handles checkpoint syntax/authority and starting-query grammar. Store read applies exclusive replay. Removing the guard therefore reuses the established checkpoint path.

Proposed exact replacement sentence for spec and both catalog occurrences:

> On automatic reconnect, `Last-Event-ID` overrides the original `after` value or `start=now` intent; replay begins exclusively after that checkpoint.

Near the initial intent list, clarify that `start=now` without a selected-SSE reconnect checkpoint observes future groups; it does not invalidate the existing exactly-one-starting-query rule. A header alone still cannot rescue `changes/`, malformed query, simultaneous after/start, or unknown/repeated query members. Do not broaden the clause to finite JSON: it continues to ignore this header, including malformed/repeated values. Existing durable-application sentences remain unchanged; server receipt of a transport ID does not prove client durable application. Event Source `events/` and Resource `view=events` continue supporting after only.

## Focused source regressions and discrimination

Build first: `transaction-sse.test.ts` imports emitted runtime, so a source-only test run against stale dist is not evidence. Keep raw duplicate-header inputs as raw pairs; do not normalize them through Headers before testing multiplicity.

|Case|Necessary new evidence versus existing regression|
|---|---|
|Saved checkpoint + same start=now URL|New core case: initial start-now stream, fully validate/applyG1, disconnect, commitG2 offline, reconnect with savedG1, receiveG2 first; then liveG3. Exact full groups/IDs/order and finite parity exclude dropped or repeated groups.|
|Headergenesis + startnow|Replace obsolete500 fixture with200 plus actual retained first-group observation, not status alone.|
|No-header initial anchoring|Existing source test351–379 (controlled minimum wait) and initial phase of new case prove unchanged behavior. Do not add another idle/head-capture matrix.|
|Explicit after + header; reopen; partial frame|Existing212–222, effective-cursor583–612 and raw truncated-response620ff remain regression gates. They already prove retained replay/override/reopen and partial-frame recovery; no new installed restart or proxy needed.|
|Raw header rejection on newly admitted intent|Add startnow counterparts for empty, malformed and repeated header cases; require400 invalid-parameter and no stream frames. Existing after cases remain. A genuine application Event ID is outside the checkpoint domain; use an existing group’s actual Event ID as one malformed-header comparator, not a fabricated new token rule.|
|Effective authority errors|Extend existing epoch/view/future header test583–612 to the startnow target: same real-checkpoint-derived inputs and expected410cursor-expired/409foreign-view/400invalid-parameter, with existing context rules. With If-None-Match:* still refuse rather than304. No new authority precedence.|
|URL grammar, auth and media|Rerun current missing-query, raw repeated/header, selected-after and401-before-invalid-input cases. Add one finite JSON startnow request with malformed/repeated header:200, captured current anchor, empty groups; confirms SSE-only precedence. Existing changefeed parser tests cover both/unknown/repeated query grammar; do not multiply every grammar case by every header.|
|Minimum retry and shortcuts|Add one controlled-clock startnow+savedheader case where a later commit satisfies a known future minimum: saved checkpoint remains effective and intervening group is delivered. Retain existing no-header anchor and cancellation tests. Extend existing HEAD/304/412 cases to valid startnow+header without opening a stream slot; expected context corresponds to observed head, empty body. No new minimum semantics.|
|Lifecycle/backpressure|Rerun unchanged entire SSE file: stalled sink, capacity, first-fault/cancellation, shutdown and finite work tests. The tiny HTTP repair changes no pump/registry algorithm; do not invent new pressure/fault experiments.|

Use existing fixture and explicit sequential waits for actual readiness/frame completion; source tests needing minimum retry use controlled now/wait seams. Keep existing finite/recorder deadlines. Check real Problem code and status where affected, not a new guessed error ordering. All newly passing start/header cases must fail on the old guard; the central replay check must also discriminate a200 implementation that ignores the header.

## Fresh installed proof: exact8-request reconnect scenario

Add a second test to the existing installed SSE file, in a fresh directory and one child, with unchanged Task descriptor/configuration and canonical Host. No prior seed Resource is needed: the source suite already covers nonempty-head no-header anchoring. Test starts with empty Scope. Only one live stream at a time; close and await the first before offline mutation/reconnect. Preserve exact fixture/source/entry/configuration binding and authentication handles from the existing harness.

|#|Actual request|Required observation|
|---:|---|---|
|1|SSE GET `/acme/changes/?start=now`, no header|Ready200 at pos0, actual SSE/cache/Vary/development/context headers; no preexisting group.|
|2|POST batch createG1 Bead|Completed receipt and expected canonical Resource/property; receive completeG1 frame on stream1.|
|3|Finite GET `changes/?after=genesis`|ExactlyG1; finite JSON group bytes equal SSE data bytes and full parsed group. Independently validate create identity/type/properties and complete group framing before saving its checkpoint.|
|—|Close stream1 and await closure|Apply the validatedG1 to a small test-local state and save its checkpoint only after application; do not claim a general durable-replica implementation. No request.|
|4|POST batch createG2 while disconnected|Completed result at next position; no active stream.|
|5|SSE GET **same** `/acme/changes/?start=now`, Last-Event-ID savedG1|Ready200; first complete group must beG2, previousPositionG1, ID its checkpoint; neverG1. Failure to produce offlineG2 fails before another live write can hide the loss.|
|6|Finite GET `changes/?after=G1`|ExactlyG2, full group and encoding parity with first reconnect frame; independently expectedG2 mutation.|
|7|POST batch createG3|Completed result; next reconnect frame exactlyG3, previousPositionG2; no duplicate or skipped group.|
|8|Finite GET `changes/?after=G2`|ExactlyG3, full group/encoding parity. Close stream2; final recorded group list exactly[G2,G3].|

Counters:3 POST +3 finite GET +2 deliberately nonterminal SSE GET =8 actual requests, one installed child, three authored single-operation commits, stream1 one complete group and stream2 two. SSE requests are not completed finite exchanges. Reports retain each stream’s raw bytes/hash/headers/frames plus finite response bytes and exact request observations; validate id equals each data checkpoint and complete groups contain independently expected mutation effects, not just mutually agreeing representations. Canonical dynamic checkpoint URLs are built from actual validated group checkpoints within the test’s existing exact executor route binding. No conformance raw-capture DSL is involved.

The existing idle test remains a separate scenario:2 POST +1 finite GET +1 SSE GET =4 requests/one child, at least15 seconds idle. Together fresh positive installed evidence is12requests/two children, not a replay of historical G2a evidence. Existing5-second finite,25-second per-recorder,45-second per-test cohort,60-second test outer and10-second cleanup bounds stay unchanged. Stream recorder remains2MiB/16KiB-header; finite executor remains64KiB body; runtime transport/default timing unchanged. Run the two installed tests serially to avoid competing15-second live observations; each retains its own original deadline, with a150-second driver bound for the two positive tests plus setup overhead reported separately.

## One bounded counterfactual

In a separate disposable derivative of the frozen source, change only effective selection for startnow to ignore its one header while leaving after-header selection intact; rebuild and package that derivative. Run only the new8-request test. Original guard restoration is a useful red check but is weaker than this counterfactual because it yields500 instead of200.

The derivative must reach reconnect request5 with200 and no offlineG2 frame; assert/report the named offline-G2 oracle failure (recorder deadline/closure if no frame), not a generic setup error. It must not sendG3 to manufacture evidence after missingG2. Up to8requests/one child is a conservative ceiling; expected stop is request5. Retain an explicit source patch/commit, distinct tar/payload digests and failing oracle/log, with every child/socket cleaned. No modified package is relabelled as clean-source positive evidence. Combined installed ceiling:20requests/three children (12 positive +≤8 derivative), no broad historical cohorts. Successful full ordered comparison in the canonical run independently rejects inclusive replay; existing foreign-token source tests reject silently starting fresh. No additional production mutants or malformed-token installed matrix required.

## Validation, evidence and done-when

After council/root acceptance, before editing inspect the newly assigned source and current repository instructions. Use Node24.16.0/pnpm11.20.0, isolated dependencies, no edits in deletion author’s checkout. Run meaningful affected checks: typecheck/build; transaction-sse, transaction-changefeed, transaction-http and transaction-event-sse source tests; transactional-catalog citation/mirror and profile-catalog-boundary tests; format/lint/boundaries and existing Read evidence verifier. If a test filename differs, inventory rather than guessing. No automatic full workspace suite. Catalog clause changes are intentional; any evidence-binding failure is reported for disposition, never repaired by relabelling/resealing historical evidence.

Freeze clean source before fresh installed proof. Adapt the existing `changefeed-sse/packed-evidence.py` packaging pattern after inspection, but do not rerun its obsolete broad finite/RU sweeps: build/pack protocol+server, install real tarballs in disposable directory, use explicit BDP_TRANSACTIONAL_INSTALLED_ENTRY and SOURCE_COMMIT, then run the two installed SSE tests and one separately packaged derivative. Record exact commands/exits, source/tree/dirty status, archive and installed payload/entry hashes, fixture/artifact/configuration identities, routes, observed request/process counts, raw evidence and cleanup. Keep credentials local; remove disposable stores, stop recorder sockets and wait child exits. Do not add counters by altering production or testing exports.

Done when approved Scope wording/catalog/packet match, the old500 gap is removed through the existing path only, all focused checks pass, fresh installed same-URL replay proves offlineG2 and liveG3 with exact finite parity, ignored-header derivative fails the intended oracle, no owned process remains, and native/Claude/root findings are resolved. No profile admission, whole-clause executable flag, public server advertisement or general replication qualification is implied. The completed assessments remain historical; this report contains a plan and inspected feasibility, not test results.
