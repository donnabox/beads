# Fresh Snapshot creation minimum — revised implementation plan v2

Plan only under the sole fork design.md and accepted cold-client critical path (C2). No implementation authority until root accepts this revision. Original reviewed plan is preserved byte-for-byte in [reviewed-v1](transaction-snapshot-creation-minimum-plan-reviewed-v1.md); [dispositions](transaction-snapshot-creation-minimum-plan-dispositions.md) account for both reviews. Supporting consistency audit SHA256 `aa25a080d1cba17a9e757a5476ea4283d36573c0c6305251bbfba99bc497fc42` remains historical, qualified below where it overstates retained-target uncertainty.

## Outcome, source and authority

Support complete canonical minimum tuples on the canonical fresh Snapshot creation target (no query), for GET and HEAD, using existing protocol semantics. Retained manifest/page minimum support, discovery assembly, public admission and profile credit remain outside this slice. Private installed targets are not cold-discovery qualification.

Inspected main `581bad0de6f5eb098b920147f6a7e09809c553ea` and frozen C1 successor `20d61119b5faf7251a5d285c5c34a1049866ffe5`, tree `3de0c4b5589423d01eb0dcb7df7ad150d184273d`. Before authoring, root supplies the reviewed ordinary C1 + alias integration successor. Record that exact base and re-anchor every symbol below; do not overwrite either lane or assume their HTTP changes absent.

Normative basis: `docs/specs/bdp.md`, **Snapshots and strict reads** (5758–5764), **Transactional HTTP consistency** (5770–5787), **Transactional conditional reads** (5801–5811), **Scope snapshots** (5829–5843 and creation response row 5915). These establish strict current prefix, minimum eligibility or explicit consistency failure, minimum before conditional shortcuts, and safe creation GET/HEAD. Fresh creation has no old immutable anchor to reinterpret. The unsettled retained issue is specifically which position to compare (immutable anchor versus current serving observation). Retained epoch/view refusal can potentially be implemented separately without deciding that issue; neither is implemented here.

## Production changes

1. **Owner — `observeSnapshot`, transaction-store.ts:1426–1466.** Accept an optional minimum for create/preflight only. Revalidate/snapshot the tuple on every owner attempt using `snapshotTransactionMinimum`; explicitly reject a minimum paired with retained manifest/page kinds. In the synchronous owner entry capture one actual `readContext`; call `checkTransactionMinimum` before expiry, preflight, copy or allocation. Then retain the existing atomic Snapshot work and same-entry response context. A `TransactionMinimumFailure` escapes the SnapshotFailure-only catch. No new table, storage format, capacity, token, clock or representation member. No claim that successful reads forbid all maintenance.
2. **Service — `snapshot`, transaction-service.ts:252–256.** Capture/freeze the minimum and a create/preflight `{kind}` copy synchronously at service entry, before `run` or any await. Do not retain the caller's mutable request or tuple across retries. Preserve no-minimum behavior. Both minimum create and minimum preflight execute through `run` and the existing `waitForTransactionMinimum`, forwarding the actual AbortSignal. Each attempt calls the actual owner and revalidates there. The existing helper takes one finite monotonic deadline (default1000ms), retries only behind, checks abort before each attempt and converts exhaustion into catch-up-timeout with the last actual context. No owner/SQLite entry crosses a wait. Close/drain owns the actual pending promise. Private malformed/retained misuse fails explicitly, never silently drops controls.
3. **HTTP — Snapshot route and creation branch, transaction-http.ts:388–430.** Preserve authentication/method/closed-query parsing. Complete canonical controls become eligible only on fresh creation; malformed, partial or unsupported spelling remains the existing private gap. Complete controls on retained manifest/page targets remain a gap. Pass the captured tuple and signal to BOTH actual capture and authoritative nonallocating preflight. Never call an owner minimum attempt directly from HTTP. An intermediate `behind` must never escape the finite wait to the HTTP catch (565–591 excludes behind); do not disguise an unwaited behind as catch-up-timeout or broaden the outer catch to do so.

### One terminal preflight branch and existing negotiation behavior

The current pure candidate invokes `applyReadHttpSemantics` on status200, application/json and no ETag. Preflight `eligible` supplies the same decision inputs plus real context/cache headers. Therefore, for the same immutable HTTP request, a non200 candidate has the same non200 decision after eligible preflight; HEAD is terminal regardless. Snapshot problems are also terminal. Pin this actual invariant with focused tests and make the preflight branch return its decision without falling through into a second capture/service wait. It must not create a second deadline, allocate a handle, or use preflight as stale authority for later capture. A normal GET capture still observes/checks in its own single owner entry.

Preserve real capacity/integrity preflight for HEAD, conditional and unacceptable-Accept creation decisions. Today even a no-minimum406 observes the owner first. Do not introduce a406-first bypass. For a compatible minimum plus `Accept: text/plain`, eligible preflight yields406 with actual context and no handle. A future minimum plus that Accept waits and yields503 catch-up-timeout; a Snapshot owner problem remains authoritative. This is the chosen compatible extension of existing implementation ordering, **not a new universal normative minimum-versus-Accept precedence law**. The specification explicitly orders minimum before conditional shortcuts but does not establish a universal negotiation ordering in the cited clauses. Include both discriminators. Preflight is not a full Snapshot copy. Preserve global `read-http.ts:10–48` semantics and no-minimum cases. If a contrary normative clause is found at the implementation base, report that exact clause before changing this choice.

## Ownership and exclusions

Production: `packages/server/src/transaction-store.ts`, `transaction-service.ts`, `transaction-http.ts`. Focused tests: their existing natural-owner Snapshot/service/HTTP test files. Installed: `apps/bdptest/src/transactional-snapshot-installed.test.ts`, including its narrow credential adapter; a separate narrow test file is allowed only if it reuses the existing launcher/lifecycle. Documentation: only current Snapshot support/gap wording in `docs/design/transaction-implementation-progress.md` (fresh creation, finite wait/cancellation, retained/malformed exclusions).

No changes to transaction-consistency.ts absent a concrete defect and root arbitration. No runtime/config/public exports/admission, protocol/schema/spec, catalog/whole flags, RU matrix, retention/lifetime, Types, Events, Selector, mutation-minimum or retained continuation semantics. No installed fake clock/wait/fault hook. No independent competing plan or broad progress-doc rewrite.

## Direct evidence and independent oracles

- Current complete tuple on fresh create: schema-valid manifest, known literal Resource values, actual scope/epoch/view, anchor/checkpoint and context from the same entry. Empty genesis included. HEAD/preflight creates no usable handle.
- Behind then catch-up: existing controlled wait commits a real mutation after the owner entry unwinds; retry captures its independently expected afterstate and actual new anchor/context. Bypassing the check fails state/anchor assertions, not just a wait counter.
- Fixed-head future minimum through BOTH create and preflight: finite exhaustion, exact error and actual fixed context. HTTP GET error is503 with `code: catch-up-timeout`, `type: https://github.com/gastownhall/bdp/problems/unavailable`, `status:503`, `retry:after-delay` and valid Retry-After. Do not accept neighbouring503 temporarily-unavailable. HEAD has identical status/context and no body; direct service evidence checks its error reason. At a known fixed head assert the exact position and that it is below the supplied future bound.
- Wrong epoch:410 cursor-expired, type family `/problems/gone`; wrong view:409 foreign-view, `/problems/conflict`; exact actual current context, not echo of foreign controls. Preserve existing Problem fields/retry values and HEAD bodylessness.
- Inspect handle/page/record row counts through existing bounded introspection or read-only inspection after close. Failed minimum attempts allocate zero. The finite wait can execute many owner attempts, so a misplaced post-capture check could otherwise exhaust8 handles and mimic503. Require exact catch-up-timeout and later eligible capture200 with a fresh id. Do not pin a timing-dependent retry count.
- On compatible GET/HEAD/304/412/406, assert actual epoch/view and exact independently known fixed-head position; compatible positions meet the minimum using only this private fixture's canonical genesis/pos-N encoding. HEAD/304/412 have no Resource body oracle, so their exact context assertions are mandatory. Direct HTTP spy/count tests show one service call and no capture after preflight for each branch. Behind conditional/media cases cannot become304/412/406 prematurely.
- Mutate the caller's request.kind and tuple immediately after calling service while a controlled wait is armed; the original frozen create/preflight and bound remain in force. Owner-side malformed/retained misuse is checked per attempt. Pre-abort gives zero attempts. Real HTTP disconnect while waiting gives no later owner attempt after observed cancellation, clean drain and no leaked reader/timer. No instantaneous interruption promise for synchronous copying or an already pending≤5ms sleep.
- Preserve owner fencing/rollback/cleanup causes and no-minimum Snapshot behavior. Add NEW complete-tuple retained manifest and actual issued page gap assertions:500 plus `x-bdp-development-limit: minimum-position`. The existing partial create-tuple installed control stays unchanged. Preserve no-minimum retained bytes/page/reopen.

## Installed transport, budget and exact planned ledger

Reuse the actual packed installation, canonical Scope/Host transport and unchanged finite Snapshot harness: **5s wire request,45s aggregate per installed observation,60s outer test,10s lifecycle cleanup/start discipline**. The1000ms service wait is separate. Fixed-head waits are approximately1s per request, not a hard SLA or a new≥900ms/2s timing assertion. Report actual durations; enforce existing ceilings. Do not widen them if slow; preserve failure and ask for a bounded adjustment.

Reuse C1's safe view-header adapter (`transactional-collection-minimum-installed.test.ts:249–287`): remove only `bdp-authorization-view` from raw.headerLines; declare its exact header name on the selected credential handle and supply it through the resolver, including anonymous. Preserve the executor's credential firewall. Pin the transmitted view via the actual exact-request evidence without logging credentials. No public executor change.

The new small positive cohort uses2 Beads and1 Link, one seed batch and one later update. The planned ledger is31 requests, plus2 mutant requests, **33 total with a hard maximum40 and3 sequential physical children** (positive initial, positive reopen, distinct mutant). Unused headroom is not authorization for unrelated cases. All HTTP calls, including probes, count; IPC readiness is not a hidden HTTP request.

| Requests | Observation and exact fixed head |
|---|---|
| 1 | No-minimum empty create; obtain actual epoch/view, genesis manifest and pos-0 response context. |
| 2 | Anonymous complete tuple refuses401 without context. |
| 3 | One real batch seeds2 Beads +1 Link, pos-1. |
| 4–8 | Compatible create GET200, HEAD200, GET304 (`If-None-Match:*`), GET412 (`If-Match` absent tag), GET406 (`Accept:text/plain`), all actual pos-1. |
| 9–13 | Future bound pos-2 at fixed pos-1: GET, HEAD, conditional304 candidate, conditional412 candidate, unacceptable-Accept candidate; each503 timeout with pos-1, GETs exact Problem code/type. |
| 14–17 | Foreign epoch GET/HEAD then foreign view GET/HEAD; actual pos-1 context. |
| 18 | Later eligible create200 pos-1 with a new id, proving capacity remains usable. |
| 19–21 | Malformed tuple, partial tuple, and misspelled required field while remaining canonical control fields are present; explicit existing gap. |
| 22 | NEW complete-tuple retained manifest refusal500 + minimum-position marker. |
| 23–24 | Same retained manifest no-minimum GET/HEAD; exact original bytes/bodylessness and pos-1. |
| 25 | Real update to known literal afterstate, pos-2. |
| 26 | Compatible create GET pos-2 with literal updated state/anchor. |
| 27 | Earlier retained manifest no-minimum GET retains bytes, now actual serving pos-2. |
| 28–29 | After genuine reopen, earlier manifest GET/HEAD preserves bytes/context pos-2. |
| 30 | Fixed pos-2 future pos-3 GET timeout, exact Problem and context. |
| 31 | Eligible creation200 after reopen, literal current state and pos-2. |
| 32–33 | Separate derivative mutant: empty no-minimum GET obtains context, then future-minimum GET must fail the same positive timeout/status/anchor oracle by returning premature success. |

Only requests1/4/18/26/31 allocate positive handles (5 total, below8). Positive waits9–13/30 total approximately6s; startup/reopen and wire work must still finish within the unchanged45s aggregate. Run the2-request mutant as a distinct bounded observation with its own unchanged45s/60s harness ceilings and include its child/requests in the new-cohort total. No installed behind→scheduled-catch-up claim; that scheduling evidence is direct controlled-wait evidence.

### Retained page regression is a distinct existing larger cohort

The small fixture cannot issue continuation at actual pageItems32. Preserve the existing34-Bead/34-Link Snapshot/changefeed/reopen cohort and add exactly1 complete-minimum refusal using its **actual returned beads.next** URL, captured epoch/view and pos-4; assert500 plus the gap marker. Its existing no-minimum page collection, retained bytes/reopen and68 point-read convergence remain intact. Do not fabricate a page URL or reduce page size.

Source-counted existing ledger is88 HTTP requests:2 hosted Type reads + denied write/missing point/empty creation3 +4 seed writes +1 populated capture +2 continuation GETs + retained manifest GET/HEAD2 + wrong-stream1 + update1 +3 retained refetches after reopen +1 feed +68 point GETs. With the new page control it is89, with2 sequential children (initial/reopen), under its existing45s aggregate/60s outer and5s wire/10s lifecycle bounds. This is a separately enumerated existing regression cohort, **not hidden inside the small40-request/3-child cap**. Reconcile counts at the assigned base before authoring; any extra traffic or resource growth needs explicit ledger adjustment.

Also preserve/run the existing no-minimum capacity cohort:39 requests and3 sequential children (10 in the first fixture,29 across the second fixture/reopen), unchanged deadlines and source assertions. It proves eligible no-minimum HEAD/conditional/media decisions still consume no handle. The planned installed total across these named cohorts is161 requests and8 sequential child launches (33+89+39;3+2+3). The cap on newly authored small evidence remains40/3; total ceiling allowing its7 spare calls is168 requests, not a global40 claim. Local cleanup/error-helper tests launch no children. No parallel listener quota expansion.

## Source binding and counterfactual provenance

Freeze clean positive source before packing. Report source commit/tree, package bytes, entry/payload/config/artifact/fixture bindings and actual exchanges/durations/child cleanup. Preserve historical parent evidence.

In a separate disposable derivative checkout, bypass only the new owner check and pack that derivative. Prefer a committed throwaway branch and its actual commit/tree; alternatively record an explicit derivative binding (clean parent + exact patch/hash + derivative tree/build provenance) rather than reporting parent as its clean source. Never mutate the frozen source or silently reuse its `workspaceDirty:false` claim for the derivative. Assert derivative installed payload digest differs from positive. Record the expected failed oracle and process exit, retain raw logs, and clean the derivative child. A mocked direct countercontrol is not a substitute for these expressly budgeted2 installed mutant exchanges. No permanent fault hook.

## Validation and done-when

After root accepts and supplies the base: re-anchor symbols and ledgers, implement only owned changes, run affected owner/service/HTTP checks including C1/alias integration regressions and appropriate unchanged executable smokes, build/typecheck/lint/format/boundaries/Read evidence verification, then freeze source and run the bounded fresh packed observations above. No automatic fullworkspace repetition. Reports separate direct, actual installed positive and installed derivative negative evidence.

Done requires both fresh paths using one finite wait, no raw behind reaching HTTP, immutable entry controls, same-entry representation/context, exact refusal/context and no-allocation discriminators, terminal preflight, preservation of current negotiation behavior, genuine retained page gap/parity/reopen, accurate progress wording, and no owned processes/handles left. No public profile/discovery/whole-contract claim. Unforeseen seam or normative conflict returns to root before scope expansion. Retained position interpretation and malformed-control protocol design remain unresolved outside this patch; retained epoch/view is separately possible work, not claimed blocked by that position question.
