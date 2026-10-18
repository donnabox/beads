# Scope start-now resume: native code review

## Finding

1. **P3 — Preserve actual finite request observations promised by the plan.** `apps/bdptest/src/transactional-sse-installed.test.ts:104–110,349–371,430–444`: the finite executor checks actual write completion, but the helper discards `response.exactRequest` and finite wire-response bytes. The new report records authored method/target/category entries pushed before execution plus three decoded finite body strings. These are useful and the live assertions are meaningful, but they are not the accepted plan's retained “exact request observations.” A later reviewer cannot inspect the actual finite request projection from this report, and failed attempts are represented by the same authored request list. Preserve a bounded, credential-redacted observation of the six actual finite exchanges for the new test (method/target, completed write state, request body length/hash and safe actual projection, status and response length/hash); keep the old idle scenario unchanged. Alternatively root can explicitly narrow this retained-evidence promise and label the current list as authored attempts. This is an evidence-retention discrepancy, not an identified replay/runtime defect or false successful assertion.

No other actionable finding in the frozen eight-file delta. Fresh installed positives and the separately packaged ignored-header derivative remain independent pending gates; their completion is not inferred here.

## Exact basis and checks

Source `/Users/dbox/repos/wt/vickie/vickie-start-now-resume`, HEAD `d53450a4f337bcd01aa5963980b6e85eb50a987e`, tree `6712045caed63dc0ad36f2876b3d8ae8cfa85480`, direct parent `5995f6b08ebd045e4343059c2da1cffd38b09b80`. Clean status and unchanged HEAD verified at start/end. Exactly eight approved files changed, 450 insertions/83 deletions. Accepted plan v2 SHA256 `462b96b97c041e4b1a4f66c9d3a51e7c5de35617cb66f16eae58eca996179a21` and root acceptance read in full.

Read complete diff, relevant current HTTP/service/recorder paths, author report/source/exits/preservation and focused result JSON. All eight actual source hashes match source.json. Local Git/text/JSON/hash calculations only: no tests, builds, network, source edits or services. Only this requested report was written. Author recorded 163 focused tests across six files with zero failures; listed typecheck/build/format/lint/boundaries/Read verifier/diff exits are 0. These are inspected author results, not reviewer reruns. The author accurately labels fresh installed proof pending in the report consumed here.

## Production and contract

The entire production change is removal of the two-line development refusal at `transaction-http.ts:438`. Query parsing and minimum parsing still precede effective header selection. Only selected SSE consults the raw Last-Event-ID fields, repeated fields still reject, and one field maps to the existing `{after}` intent. Finite JSON continues to ignore malformed/repeated headers. Authentication, Event Source routes and their distinct starting-query grammar are untouched.

`transaction-service.ts:366–389` freezes the effective intent and reuses the selected checkpoint across minimum retries. The no-header start path alone captures admission head. Existing owner checkpoint validation and exclusive replay are reused; no cursor, store, retention, pump, policy or lifecycle algorithm is added. The removed gate therefore exposes the operator-approved behavior without synthesizing a new interpretation of checkpoint authority.

The normative change is limited to the approved start-now exception and exclusive reconnect sentence. Durable applied-checkpoint obligations and adjacent explicit-start text remain. Canonical catalog and maintained packet change only the same existing reconnect selectedText; the mirror test adds that ID. No schema, admission flag, coverage assertion or historical evidence is rewritten. Both stale progress statements are reconciled, with no browser interoperability claim retained explicitly.

## Source tests and installed-test design

Source tests replace the old 500 with actual retained first-group delivery, reject empty/malformed/repeated headers with invalid-parameter and no groups, reject an actual Event ID as the wrong token domain, and cover effective epoch/view/future checkpoint refusals before conditional shortcuts. Selected start-now cases accompany HEAD/304/412 slot checks. The finite JSON malformed/repeated-header case retains current-head/empty-groups behavior and is correctly a regression, not a discriminator of this repair.

The new controlled minimum test commits the intervening group inside the existing wait seam and requires that group at pos-2 after the saved pos-1 checkpoint, with current response position. The same-URL source test commits offline before reconnect and requires pos-2 followed by pos-3, excluding inclusive replay and skipped offline history. Existing finite, backpressure, cancellation, reopen and Event tests remain regression coverage.

The installed scenario's sequencing is meaningful: start empty, create/observe/finite-compare G1, apply its independently checked upsert before saving the checkpoint, close the first socket, commit G2 offline, reconnect at the identical URL, and require G2 before issuing G3. Independent create IDs/types/properties, exact position chain, event type/subject/ordinal and single-upsert counts prevent mere agreement between two wrong representations. Finite groups equal the complete SSE object and finite text contains the exact SSE data encoding; final frames are exactly G2,G3. The saved checkpoint differs from G2's checkpoint. Actual receipt completion and position are checked separately.

The ignored-header derivative would reach the named offline-G2-before-live-G3 wait with 200, then fail before G3; the catch labels that specific failure and finally closes both recorder sockets and child. Its actual frozen derivative bytes/package identity and observed failure still require the separate pending author artifact; source inspection alone does not establish that execution.

The new positive has three POSTs, three finite GETs and two SSE GETs, one child. Finite 5s, recorder 25s, cohort 45s, outer 60s and cleanup defaults are unchanged. Both recorders close in finally; the shared deadline reaches launch/exchanges/recorders. No abandoning Promise.race or parallel installed cohort is introduced. The report distinguishes deliberately closed nonterminal streams from finite responses.

## Preservation and limits

Independently extracted and whitespace-normalized the old nested helper versus the new local factory: identical bodies. Replacing only that declaration block with the factory invocation makes the entire old idle test whitespace-normalized identical. Its four requests, fifteen-second wait, assertions, report fields and cleanup are preserved. The per-factory key counter resets independently. This confirms preservation beyond trusting preservation.json.

The current test's raw SSE response bytes/hash/frames are retained. Finite response body text is retained for all three parity observations; the specific request-projection limitation is finding 1 above. No general durable replica, whole-catalog execution, public discovery or profile qualification follows from these checks. Root retains external council and final installed/counterfactual evidence disposition.
