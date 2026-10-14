# Comparison timeout: stage-0 parent re-anchor

No blocking mismatch found in the accepted eight-file design. This is read-only preparation, not source authoring or execution release. Frozen inspected source is c555fe01df5cd20454fb0d663431df816c1270aa, tree062a3ab7e24236417bd915b373a4b44ba9908647, clean `vickie-local-schema-serving`. No tests/builds/services/installs/network or source edits occurred. Only this supporting report was written. Root must assign the isolated authoring parent, then review draft measurements before execution.

Authority remains accepted plan941f350fb4634a77000f3bc0108f3ea80c94ee5701aef94bdc456c66a28ad4e6, dispositions994e5fa3195a2515db8fcb9b8bfd6f7ad30f8ba5557909230408e29a0a4155db, native closureb2a637e8f4d951a6fb0166198822ee8fe55a3a7687ca6690b3c059da7ab17d2c and `transaction-comparison-timeout-root-queued-acceptance.md`. This is subordinate to the sole design plan; it adds no product law.

## Current seams and exact ownership

1. `packages/server/src/transaction-comparison.ts:10–54`: current result type has bare temporarily-unavailable; options24–31 have only receiving/budget/now/wait. At44–46 the epoch check short-circuits before deadline. Split these conditions without changing that order; invoke a bound synchronous observation only in the healthy observed deadline branch. Keep single deadline38–42 and await53 outside owner entry.
2. `packages/server/src/transaction-store.ts`: `enter`341–352 checks faults/closed/reentrancy and supplies JS exclusion; `atomic`362 onward uses BEGIN IMMEDIATE/COMMIT/rollback, so the new observation must explicitly nest in both. `metadata`407; immutable epoch initialized553; `readContext`583–592 reads current persisted position with owner-lifetime epoch/view. Add the method next to `comparisonEpoch`1724. Final persisted epoch validation and tuple creation occur in one atomic callback; no receipt/key lookup, maintenance, admission, reservation change or async work.
3. `packages/server/src/transaction-submission.ts:103–119`: observed admission is an explicit selected facade, unobserved passes full owner. Introduce the timeout opt-in here only. Current result union29–56 and mapping146–152 need the observed discriminator. Crucially, the comparison argument type currently derives from compare options at66: omit the new observation member along with `receiving`, or the callback would accidentally become a service/runtime test-config option. This is a required local typing correction within the accepted file, not an additional service/runtime edit.
4. `packages/server/src/transaction-http.ts:340–390`: `service.submit` result is consumed here; selected 409/413/406 wrappers remain. Add a branch before generic fallback386 for the observed healthy deadline, require context and retain the exact canonical503 Problem/retry/Retry-After1. Owner-stopped remains distinct and contextless. Existing outer error handling continues to sanitize missing-context/internal faults. Do not use conditional context decoration with silent fallback.
5. `packages/server/src/transaction-store.test.ts`: existing real owner fixture76 onward, SQLite prototype/fault facilities and readonly inventory patterns permit atomic/fault tests without new production seam.
6. `packages/server/src/transaction-submission.test.ts`: exact legacy timeout assertion332–376, controls-before-negotiation302, retraction379 onward, single-deadline/epoch-double449–485 are present. Add observed-path tests here while retaining all existing unobserved shapes.
7. `packages/server/src/transaction-http.test.ts`: emitted runtime import13–15, fixture82–167, final readonly `retainedReceiptInventory`108–121; earlier basic timeout818–865 remains. Replace the obsolete selected exclusion test2621–2680 with healthy current-context proof; preserve neighboring malformed/auth exclusions2600–2619 and new sequence tests2682 onward. These helpers use a real listener and the existing private scheduler/comparison options.
8. `docs/design/transaction-implementation-progress.md`: append a bounded timeout section after current schema section1742 onward; preserve sequence1720 onward and schema history. State default installed timeout sensitivity and masked maintenance fault remain unobserved/unrepaired, respectively.

Comparison/store/submission/store-test/submission-test bytes are unchanged from the accepted ed359 basis. HTTP now contains reviewed sequence refusal observation and the private `hasMinimumControls` export, and HTTP tests contain their additional regressions. These do not repair this timeout and must be preserved. Current service `submit`499–517 still propagates every nonreceipt result under its existing `run`; the new private observed branch can flow without a service edit. Local schema adapter and schemaObserved need no edit. No package allowlist/export addition is necessary.

## Proposed private types and owner method

Suggested narrow spelling (not authored):

- Owner `observeComparisonDeadline(expectedEpoch)` returns a frozen `{kind:"observed", context:TransactionReadContext}` or `{kind:"owner-stopped"}`. Healthy path is `enter(() => atomic(...))`; compare persisted epoch against the original epoch before calling readContext. A normal already-closed owner returns stopped; sticky owner/maintenance errors that reach this method still propagate (do not catch and convert). Reentrant entry throws. A defensive final mismatch returns stopped without a mixed tuple.
- Comparison retains every legacy union member. With explicit bound callback, healthy deadline yields `{kind:"comparison-deadline", context}`; epoch/normal-owner stop yields `{kind:"comparison-owner-stopped"}`. Without opt-in either stop remains exactly `{kind:"temporarily-unavailable"}` even when the supplied full store has the new method. No capability detection.
- Submission maps the two observed forms to a source-private `kind:"comparison-refusal"` union with `reason:"deadline"` plus required context, or `reason:"owner-stopped"` without context; both retain status503/code temporarily-unavailable/retry after-delay. Legacy unobserved still maps to its exact `{kind:"refused",status:503,code:"temporarily-unavailable",retry:"after-delay"}`. A dedicated kind avoids accidental overlap with other generic refused outcomes.
- HTTP handles this private kind explicitly. For reason deadline it runtime-checks context (despite the required TypeScript field) then emits existing Problem plus tuple; missing context throws. For reason owner-stopped it preserves the bare transient response. Neither discriminator is serialized into the Problem. `TransactionServiceSubmissionResult` already excludes only receipt from the submission union, so it requires no edit.

These are suggested implementation names; final draft can choose equivalent closed discriminants. The load-bearing constraints are explicit opt-in, synchronous deadline capture, missing-tuple failure, no unobserved fields, and no callback in inherited service/runtime options. No caller-supplied tuple or generic callback framework is proposed.

The existing `comparisonEpoch` closed shortcut1724 still precedes the new operation; maintenance failure1340–1357 closes the owner, so its original masking behavior remains. This slice must not claim that all maintenance faults now reach500. A changed-epoch double/final-check control is synthetic: exclusive owner provisioning520 and epoch initialization553 provide no supported live epoch rotation.

## Proposed deterministic source proof and cost

Propose 12 targeted semantic cases for the draft: six new store cases, three new submission cases, two retained submission regression cases, and one replaced HTTP case. Thus nine new cases plus one rewritten existing case; no new app scenario. These are planned counts, not test results. Full selected store/submission/HTTP/dependency/alias regression counts must be statically inventoried in draft preflight, including parameter expansion, before release; do not pretend this12 replaces all accepted regressions.

Store six: (1) one atomic metadata-only entry, frozen tuple and later-progress immutability; (2) defensive final epoch mismatch yields stopped; (3) normal closed yields stopped; (4) reentrant entry throws; (5) fenced owner original error propagates; (6) injected metadata/SQLite failure propagates with current rollback/fencing behavior. Existing test-only SQLite/fault machinery is sufficient. Instrument preparation/exec only within synchronous method call; no receipt/key statement and exactly BEGIN/COMMIT on healthy path. Do not write a production metadata mutation API.

Submission three: (1) observed unresolved deadline gets post-independent-commit context and no extra admission; (2) defensive epoch-stop exact observed shape without context/admission in new epoch; (3) unobserved full-store exact shape plus spy proving observation is never called. Retain two explicitly selected existing cases: exact bare timeout at332 and repeated replacement/original deadline plus epoch double449. Existing negotiation/retraction/alias dependencies remain separately targeted regressions.

HTTP replacement planned12 request dispatches, one listener, no installed child:

1. GET changes at head0 to capture actual pre-wait tuple; exact empty page/no continuation.
2. POST two-member creator/dependent sequence; hold its first scheduled turn.
3. POST direct dependent comparison (key dependent), kept pending at its first genuine comparison wait.
4. In the deterministic comparison wait, POST independent create of beads/b; require completed receipt/currentpos1; then advance only the private comparison clock to its original deadline.
5–7. After503, GET complete beads, links and changes pages (limit32, expected cardinality1/0/1, nextnull). Require only exact independently described b and its one group; no a or Link/effect. Timeout response must have canonical Problem, retry after-delay, Retry-After1, no Location/receipt leak and epoch/view matching pre-wait with positionpos1 (different frompos0).
8. GET beads/a404 as an explicit retained no-prefix oracle.
9. After releasing and awaiting the original two-member sequence, POST the same direct candidate; require equality with its original dependent receipt and correct currentpos3, not another execution.
10–12. GET complete beads/links/changes pages (2/1/3, exact identities/content and nextnull) to prove no duplicate replay effects. Close and inspect real readonly receipt rows: exactly creator/dependent/independent, not a timeout receipt. Inventory is DB evidence, not a hidden HTTP call and not inferred from inspect().

The request order above describes dispatch; comparison response completes after request4. Final sequence content is validated against its normative schema and independently expected two entries. Releasing held callbacks and awaiting sequence in finally must precede drain even when a sensitivity assertion fails. Preserve primary and cleanup errors. Existing old selected case has four requests (sequence,timeout,a404,replay), so proposed replacement adds eight source requests. Earlier basic timeout test remains untouched.

Private comparison clock: one deadline, one unresolved admission, exactly one comparison wait; independent operation uses another key and must not itself wait. Current service loop212–225 awaits nextTurn once before each step and returns immediately on step.done. For the two-member sequence, plan exactly two scheduler deliveries (one held initial delivery, one post-release delivery), no third completion turn; the draft must assert and measure this against the actual stepper results. Existing local deadlines remain; no host-load/sleep timeout forcing.

Five proposed source derivative controls, isolated: remove observed HTTP wrapper (named context presence); substitute pre-wait position (named currency); remove required deadline tuple (named expected503/context assertion observes sanitized500); leak observed fields into unobserved result (exact-shape/no-invocation test); bypass final epoch check (synthetic store mismatch discriminator). First three select only the rewritten HTTP case, at most36 dispatched source requests if each completes all12; last two have no HTTP. Actual early-stop counts, callbacks and cleanup are measured in draft execution ledger before root release. No generic exception qualifies as a sensitivity pass.

## Installed regressions and constraints

At c555 the two established installed fixtures exist unchanged: `apps/bdptest/src/transactional-direct-refusal-installed.test.ts:107`, title `observes current context on installed direct refusals within14requests`; `transactional-sequence-refusal-installed.test.ts:173`, title `observes current context on installed sequence refusals within13requests`. They retain5s/exchange,90s scenario,120s outer and existing10s close conventions. The sequence file also has a separate confinement test125 with no HTTP. Proposal: fresh frozen-source installed regressions14/1 +13/1 =27 requests/two children, only after root measured release. No new timeout installed test, no timeout predecessor package run, no unfiltered transactional matrix.

These27 ordinary regression requests do not discriminate comparison timeout. Deterministic default installed timeout remains unobserved: current installed configuration exposes no comparison/scheduler control. Current runtime default comparison budget1000ms is not changed or forced by load, database edits, process stopping or new launch flags. Body65536/header16384/response131072/target8192 caps stay fixed. Source/body/response sizes and exact selectors require prospective draft measurement; no numeric size claim is inferred at stage0.

## Decision and release boundary

No new product decision or eight-path expansion is needed. Root must still assign the actual isolated parent and authoring release, then accept precise draft counts/bytes/control patches and source regression selection before execution. In particular, inherited options typing must exclude the new callback, and the obsolete HTTP exclusion expectation must be replaced. If the union cannot pass the existing service type without editing service/runtime, return to root rather than widening ownership. Preserve PR119 source/council and unresolved minimum/Event/public activation work.

## Frozen source fingerprints

| Path | SHA256 |
| --- | --- |
| `packages/server/src/transaction-comparison.ts` | `c54a351ce6aa61e25f56851a57acf1ea97781c1ad655aa7a1bdb99f874faff0c` |
| `packages/server/src/transaction-store.ts` | `92f06ec2844a5c343a8e3d8233682929fdb6eccadda3bc603b94689cf8f23636` |
| `packages/server/src/transaction-submission.ts` | `68f4bf7c8288ad0eb1d743bab07889c94ad39f283983924762ec93192e522c1f` |
| `packages/server/src/transaction-http.ts` | `e771744e51440c6043dbe35fe490058bd8102ff1ad92fba3efa4738929f1429c` |
| `packages/server/src/transaction-store.test.ts` | `d198a7dde70ef1179230c871faf4f61d88722a9d13dc1570df52cc08b762222a` |
| `packages/server/src/transaction-submission.test.ts` | `61178770397281640dc07f97326683d939c15b9ee072ae4e6d31c9b92387eb41` |
| `packages/server/src/transaction-http.test.ts` | `0b9886ff0c7b5f411a59886ddc14b912b1ce4ed9700f538d454270c54e0bedc0` |
