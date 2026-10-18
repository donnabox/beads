# Event missing-context Claude findings — native source dispositions

**One concrete scope correction is required before authoring: add the existing Event SSE regression test file and amend one stale context-presence flag.** The original plan already reconciles its three owner shape assertions, so that part of F1 is not a new defect. The requested repo-wide sweep nevertheless found a fourth affected expectation outside the five-path scope. This corrects my earlier native plan review's incomplete fanout conclusion; preserve that report as historical.

F2's proposed installed mutant is unnecessary and would increase the installed-child budget. Retain the emitted-only derivative and its source HTTP consumer. F3/F4/F6 are closed by actual source or the frozen ledger. F5 merits an explicit corruption/fault-boundary acknowledgement, without a new production mechanism.

## Binding and method

Fully read actual Claude output `transaction-event-missing-context-plan-claude/stdout.txt`, SHA256 **7ae49581171b42222aa1c476214e8a584a15a90667592514a42a30e8dbe98bb3**, including verified notes. It explicitly limits itself to supplied excerpts.

Reviewed frozen plan **cfeb672740e393b71376b4d7cf39989d7fd8ea75f097195279e2525844db8c9f** and ledger **457c0c33d0558cb92c4bbdd0ae060f81446656c6562ef59c5eb0e394b7f6ce18**, against `/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence` merged88c9408e908590ab1a4c64c9b9dee1a8b705999c/treeb2d9f2c18502e0778c51123be8b4564af052d803. Source references below are relative to that checkout.

Read-only rg/source inspection and artifact hashing; no test, build, product import, runtime, install, source/plan edit or Claude launch. Report write is session-guarded. Recommendations do not authorize a sixth path or execution.

## Findings F1–F6

### F1 — partially already addressed; genuine omitted test fanout

The frozen plan explicitly says “Preserve/reconcile the three existing cases” and “Their exact missing shapes gain independently expected context.” Thus the claim that no amendment is listed is false. The three proposed precise replacements remain right:
- transaction-events.test.ts221: exact missing plus actual configured epoch/view andpos-2 after two effective commits.
-283–285: same atpos-1 with the existing config.
-893–895: same with current.context, preserving semantic minimum and reconnect assertions.

**Additional required path:** `packages/server/src/transaction-event-sse.test.ts`.

Its existing callback at343, **“Event reconnect syntax, minimum ordering and exact context codes are preserved over HTTP”**, contains a table row at383–390:
- canonical unknown `beads/missing?view=events`;
- repeated Last-Event-ID values;
- expected404/resource-not-found;
- **observed:false**.

The GET/HEAD loop at412–423 interprets false as `bdp-scope-position` absent. With the proposed repair, this unknown identity is observed first and legitimately returns all three context fields, so both methods fail the stale expectation.

**Smallest source correction:** change only this row's observed flag to true. The existing true branch verifies all context fields andpos-1. Keep the following malformed URL `?view=events&after=bad` row false: URL grammar rejects it before owner observation and it remains400. Keep anonymous HEAD context absence and all other rows unchanged. No new test registration or helper refactor.

**Smallest targeted validation:** select that existing callback once alongside the original eight positive source callbacks. It contributes:
- seed POST1;
- finite known Scope Event page1 (one created Event; page capacity2);
- initial successful HEAD1;
- ten table rows × GET/HEAD=20;
- three override HEADs=3;
- one JSON negotiation read1;
- anonymous HEAD1.
Total **28 HTTP**, one in-process owner/directory/listener, zero installed children. These are26 raw recorder requests plus2 fetch requests. Successful SSE GET streams are not opened in this callback; the successful SSE requests are HEAD, while its successful JSON requests complete. Two behind-minimum requests retain the fixture's25ms wait budget. runtimeFixture194–280 registers every recorder close plus runtime close/directory removal; afterEach35–38 drains cleanup in reverse. This is an existing cleanup mechanism, not a claim of executed cleanup.

If root accepts this selection, revised prospective totals are **9 positive source callbacks/39 HTTP**, then the unchanged11-request emitted control and11-request installed witness: **11 case executions,61 expected/max HTTP,11 owner lifetimes/directories,4 listeners,1 installed child**. This adds no new full-CI registration. The count change must be recorded explicitly; do not silently keep33.

**Search scope:** repo-wide source/script rg for readEvents, beginEventScan, beginEvents, missing-result shapes and view=events, followed by exact unknown/missing Resource Event URLs and nearby assertions. Inspected the additional transaction-event-sse, transaction-sse, transaction-service, transaction-changefeed, generic HTTP, Snapshot HTTP, protocol wire and installed Event SSE callers. The only additional affected exact presence expectation found is the row above. Receipt-page `{kind:"missing"}` assertions concern a separate unchanged API. The finite Event HTTP unknown case at transaction-events.test.ts459–463 asserts status only and is compatible. Generic Read/Read+Update and Snapshot view-query refusals remain distinct pre-owner routes.

### F2 — reject; emitted derivative has an existing source-test consumer

transaction-events.test.ts380–383 dynamically imports **new URL("../dist/transaction-runtime.js", import.meta.url).href** and calls emitted.startTransactionRuntime. The existing Event SSE file independently uses the same top-level emitted import at191–193. Being a Vitest .ts file does not force its explicit dynamic runtime import to resolve source TypeScript.

The planned new source HTTP callback can follow that existing mechanism. Its derivative must run with the copied source test location/import base resolving to the copied dist graph; the plan already requires proving this and binding every copied file/two-hunk mutation before execution. Merely copying dist without arranging that relative consumer binding would be insufficient, but that is a concrete preparation check, not an impossible plan.

**Disposition:** retain emitted-only control, one in-process listener, zero product children. Do not replace it with an installed mutant: that would add a product child and alter provenance/budget.

### F3 — closed by exact type declaration

transaction-http.ts78 declares **type Service = ReturnType<typeof createTransactionService>**;153–155 uses service:Service. transaction-service.ts390–423 has inferred events/beginEvents return types, forwarding the two owner results through generic waitForTransactionMinimum. The explicit owner unions are in the already-owned transaction-store.ts file.

**Disposition:** no extra service path. Record this declaration as source evidence; retain the plan's stop rule for a genuinely discovered typed dependency.

### F4 — closed; canonical unknown Link Event views are admitted

transaction-events.ts143–177 recognizes both `beads/` and `links/`, assigns bead/link kind, validates the canonical local Resource id, and returns source with subject and canonical `?view=events` URL. Unknown canonical links therefore reach eventSourceKnown rather than a different routing404.

transaction-store.ts621–627 checks that subject in identities. Successful Resource writes use putResource1212–1218, which inserts the identity for both Resource kinds; resource-evaluator.ts1093 passes staged writes through that method. Existing Event tests exercise Link sources, including installed Event SSE at210.

**Disposition:** retain Link axis and all four Link GET/HEAD/media tuples. No replacement source form or extra mutation is needed; the unknown-Link witness deliberately has no existing Link identity.

### F5 — accept narrow fault-boundary wording; initial position is source-pinned

Fresh store creation at transaction-store.ts530–536 inserts position "0". The new planned pre-commitpos-0 checks therefore have a real source basis and should stay.

Adding the context read does introduce a new integrity read on this previously contextless branch. metadata414–418 throws for absent/malformed metadata; readContext590–598 rejects malformed position. Thus an otherwise unknown identity with corrupt Scope position will now propagate an internal fault instead of returning a contextless404. Existing entry failures still precede the branch, unchanged.

**Recommended wording:** “An inability to read valid current Scope metadata is an internal integrity fault, propagated through the existing failure path; the missing404 must not fabricate context or silently omit required fields. Fresh create:true metadata starts at0.” No catch/fallback, synthetic tuple, new fault seam or extra test invocation is proposed. HTTP's existing catch emits empty500 for unexpected errors (transaction-http.ts610), consistent with the shared unexpected-internal-fault rule at docs/specs/bdp.md1959 and the Event endpoint summary6593. This is not a claim that malformed API inputs now reach the owner.

### F6 — already explicit in frozen ledger

The final ledger lists per-request Accept: application/json for seed, known page and four finite tuples; text/event-stream for four SSE tuples and anonymous SSE GET. The plan describes this order, explicitly references the request ledger and requires exact method/URL/Accept checks. Its “Content-Type and Accept are application/json” sentence is scoped to the seed.

**Disposition:** retain the ledger and exact authored/wire request checks. The eight missing-header failures alone indeed cannot prove branch selection, which is why request media binding remains independent and required. No new request or control is needed.

## Verified notes from Claude

- Same-entry placement, frozen context, missing-before-semantic-minimum/reconnect and unchanged success capture are supported. URL/input validation remains earlier.
- HTTP changes are two context arguments to the existing projector.
-404 retry-after and WWW-Authenticate absence follow the current Problem builder;401 remains before Event observation.
- HEAD length attribution is correctly clarified by native N1: read-http41–46 computes length before body removal; runtime preserves it.
- Unknown/missing context is Scope/view/head data. The current fixed-view implementation supports this case; do not claim new dynamic authorization-policy behavior from the same tuple.
- Original33-request arithmetic was internally consistent for the original scope, but the newly found regression selection requires the explicit61-request successor accounting above.
- The152-byte absolute-Type seed and its digest are already fixed in the ledger; do not substitute relative Type text.
-45s aggregate may expire before eleven5s per-exchange maxima are individually exhausted. That is intentional nested bounding, not an allowance increase.

## Recommended root disposition

Accept F1's new repo-wide fanout result and F5's precise integrity acknowledgement. Close the already-planned three shape amendments and F3/F4/F6 by source/ledger evidence. Reject F2's installed-mutant substitution. Amend plan/ledger to six paths and the one existing28-request regression callback only if root accepts that bounded scope. Preserve all earlier review artifacts and original33 model as historical. No broader redesign, new semantic choice, public qualification or execution release follows from this audit.
