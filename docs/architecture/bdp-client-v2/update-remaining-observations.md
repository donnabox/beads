# Remaining Update diagnostic observations

Read-only audit at frozen **59915ea916a210f5ce36611eb6d799f7c3124056**, `/Users/dbox/repos/wt/vickie/bdp-alias-qualified-integration`. Current evidence is **81 normative rows: 48 bounded-observation bindings and 33 omissions**, not 48 passing normative contracts. Existing matrix: 23 scenarios / 221 actions / 705 assertions. No code changed, tests run, new mappings made, or YAGNI proposal treated as law during this audit.

## Recommended next slice

Add **two deterministic HTTP/process diagnostic plans** to the existing runner: pending/dependent retries and interleaving. Reuse the emitted development runtime, existing scheduler barrier, real SDK Fetch transport, existing assertion/capture engine and disposable store cleanup. Do not add a runtime clock, fault API, authorization fiction, schema fixture or alternate execution engine.

The existing source-only child parks **dispatch 2 before its callback**, after the first member committed. Other dispatches continue normally. This is sufficient for new HTTP requests to observe an admitted but unexecuted creator/tail and to interleave work. The fixture already has bounded IPC waits and cleanup on assertion failure/abort. The new work is observation/action wiring, not a new runtime mechanism.

**Plan A: pending creator and dependent retry.** Submit a four-member original: create fixed `beads/prefix`; create named allocated Bead (parked at dispatch 2); create Link using that creator; create independent `beads/tail`. Await IPC `parked`, then prove prefix exists and tail does not.

1. Submit the original carrier concurrently while parked. Assert prefix replays; pending creator, dependent and reserved tail report `idempotency-in-progress`/409/`after-delay` in the correct positions. Re-read graph to prove no duplicate execution or premature tail.
2. Submit a singleton against the pending tail key, including a statically different request. Read+Update must still return in-progress, not conflict; no side effect. This is Read+Update precedence, not Transactional pending-receipt joining.
3. Submit another short sequence referencing the pending creator under its original key, but with a **fresh dependent key** and fresh independent-tail key. Assert transient creator/dependent and successful independent tail. The later dependent must not poison either the original dependent key or its own fresh key.
4. Release the original once. Await the original request's completed response, assert all intended original work succeeds once, and compare its exact-key retry with its retained result. A singleton using the fresh dependent key and now-known canonical creator identity must execute, proving that key was not bound to the transient failure. A second retry must replay that fresh success.
5. Assert graph after-state, allocated identity equality and key conflict after a retained success. This distinguishes a real transient release from returning the expected Problem while improperly retaining it.

Potential **bounded** new bindings: `sequence.transient-predecessor`, `idempotency.in-progress`, `idempotency.key-reservation`, `idempotency.concurrent-dependent-retry`, `idempotency.transient-not-retained`. They cover pending-based transience only, not rate limiting or service-unavailable injection. No claim that HTTP observation can prove the absence of an internal lookup. It proves its required externally observable non-poisoning/no-effect behavior.

**Plan B: real interleaving.** Original sequence creates `beads/prefix`, then parks an update of that Bead before execution, then performs an independent tail. While parked, send an unrelated keyed singleton update adding a distinct property to prefix. Assert that request finishes before release and its read-back revision/state is visible. Release the original; its patch must observe/preserve the intervening property and add its own property. Assert both values and one retained outcome per key. A deletion variant can be a later extension: delete prefix between members, then observe later not-found while independent tail still succeeds.

Potential new binding: `sequence.interleaving`. This is one authority process, not distributed writer qualification.

**Helper constraints:** `development-process-actions.ts` currently hardcodes `release` to wait for `turn-completed` index 4, and `submit` stores only whether delivery occurred. Competing requests change dispatch numbering, so index 4 no longer proves original completion. Extend source-only observation wiring to await the original request result and feed status/body to the existing assertion runner. Keep the scheduler's existing dispatch-2 barrier unchanged. Do not add sleeps, fabricate a completion event, or infer completion from an arbitrary later dispatch. Existing `development.request` actions can send SDK requests against the active process Scope. Keep all targets bound to that disposable loopback Scope and existing fixture credential.

**Acceptance:** authored assertion references map only observed obligations; controlled child never remains parked after failure/abort; no unhandled original-request rejection; every child is killed/awaited before store removal; no arbitrary target/credential controls; both positive after-state and exact-key recovery assertions; build completes before process tests; run matrix/inventory, original process/HTTP tests, applicable checks and both smokes on pinned Node/pnpm. Existing 48 bindings remain historical partial evidence. At most six additional rows would gain bounded observations; actual review determines mappings, not a target count.

## Contract and implementation evidence

All paths below are relative to the frozen checkout unless stated otherwise.

| Evidence | Exact location | Relevant rule/behavior |
| --- | --- | --- |
| Sequence behavior | `docs/specs/bdp.md:2685–2739` | declaration order, individually atomic members, partial commit, independent tail and no sequence-wide lock |
| Transient creator | `docs/specs/bdp.md:2799–2850` | transient dependency consults no key state, executes/retains nothing, releases its own claim; disconnect does not cancel |
| Pending/claims | `docs/specs/bdp.md:3217–3299` | all unknown claims at admission; all-or-none admission; in-progress precedes replay; retry retrieves retained outcome |
| Retention | `docs/specs/bdp.md:3301–3336` | completed disposition detail may expire into permanent tombstone; failures may be forgotten; allocated identity retained |
| Recovery | `docs/specs/bdp.md:3338–3374` | effects/disposition atomic; abandoned claims cleared; no automatic sequence resume |
| Process barrier | `apps/bdptest/test-support/development-process.ts:17–56` | real emitted runtime, scheduler-only control, dispatch 2 parked; later requests can run |
| Existing lifecycle actions | `apps/bdptest/test-support/development-process-actions.ts:45–113` | current submit/release/drain/replay handling; original completion observation needs narrow wiring extension |
| Fixed authority composition | `packages/server/src/development-authority.ts:76–130` | closed Types, startup `expire(Date.now())`, 24h minimum, lifecycle clock `Date.now`, scheduler already injectable privately |
| Existing expiration implementation | `packages/server/src/recovery-store.ts:908–924` | failed outcomes deleted and successful outcomes compacted at supplied expiry time |
| Maintenance | `packages/server/src/sequence-executor.ts:419–451` | scheduled expiry uses configured member clock; altering scheduler alone cannot move that clock |
| Existing lower-level expiry tests | `packages/server/src/recovery-store.test.ts:590` | success/no-op tombstones, failed forgetting, resolutions and turn reclaim already exercised below HTTP |

## Expiry/compaction is implemented, but not a short HTTP binding today

Read+Update has **retained dispositions and compact tombstones**, not Transactional receipt URLs or `detail: expired` receipt documents. Do not use Transactional store tests to claim Update HTTP evidence.

The production composition hardwires a 24-hour minimum and `Date.now()` for terminal retention and maintenance. The process fixture controls scheduling, not time. Running its maintenance callback early leaves real current time unchanged. Waiting over 24 hours would be real evidence but is not a bounded practical next slice. Backdating SQLite fields, inserting manufactured retained outcomes, shortening the mandatory retention, or monkey-patching Date globally would bypass the behavior this qualification is intended to establish.

`openRecoveryStore.expire(now)` and the configurable member/lifecycle clock already support lower-level deterministic tests. Root could separately choose to bind those **component-level** observations into the diagnostic runner, clearly labeled as such, or approve a narrow authority clock-composition seam for future HTTP observations. Neither is included in this proposal. A component test using genuine evaluator-created outcomes followed by explicit expiry is preferable to seeded synthetic outcomes, but still does not establish the actual HTTP composition's clock/maintenance behavior.

Future expiry acceptance, if that fixture is authorized: observe success and failure before the minimum; expire completed create/update/no-op/delete into non-reexecuting same-key expired outcomes; preserve different-request conflict; forget failures so the old key executes against changed live state; bind a fresh dependent through an expired creator's original allocated identity; preserve alias recorded-target/miss identity and non-renewal across retries; verify restart with compacted keys. Permanent lifetime behavior is a persistence invariant, not a duration a finite test can literally exhaust. No public advertised retention guarantee can be established while development discovery remains absent.

## Every omitted row: disposition

The following table accounts for all **33** omitted rows. IDs omit the `read-update.` prefix. “Bind now” means a bounded observation, not complete conformance.

| Row | Next disposition |
| --- | --- |
| `discovery.document` | Real public profile admission/discovery work; not a diagnostic shortcut |
| `discovery.aliases` | Same; development alias routes are not an advertised BDP claim |
| `discovery.limits` | Same; fixed local limits are not advertised limits |
| `discovery.operation-directory` | Operation Directory is absent; requires implementation/admission decision |
| `singleton.incident-links-nondisclosure` | Hidden-Link fixture requires actual configurable policy composition; all-visible fixture cannot establish it |
| `singleton.owned-link-source-revision` | Real owned-Type installation/composition needed; generic evaluator support alone is insufficient |
| `singleton.owned-link-no-op` | Same |
| `singleton.hidden-subject` | Actual hidden authorization fixture needed |
| `alias.authorization` | Actual alias read/write policy fixture needed |
| `alias.forbidden` | Same; Origin refusal is not operation-local alias policy |
| `alias.retained-without-reauthorization` | Actual changing alias policy needed |
| `sequence.transient-predecessor` | Bind now, Plan A; pending creator variant only |
| `sequence.expired-creator-binding` | Implemented lower-level behavior; clock/retention limitation described above |
| `sequence.retry-after-hint` | Pending plan can observe allowed omission. No nonnegative hint is currently emitted by pending member paths; do not add runtime output solely to turn a row green. Positive hint needs actual emitting condition |
| `sequence.contextual-validation` | Easy separate HTTP plan: schema-shaped out-of-Scope/wrong-kind subjects and invalid forward/unknown/wrong-kind labels; retain ownership/schema limits |
| `sequence.internal-fault` | Current helper cannot deterministically induce a genuine execution/storage fault; scheduler misbehavior or SIGKILL is not a bodyless HTTP500 observation |
| `sequence.interleaving` | Bind now, Plan B |
| `idempotency.in-progress` | Bind now, Plan A |
| `idempotency.key-reservation` | Bind now, Plan A; pending tail cannot run ahead |
| `idempotency.claim-atomicity` | Plan A supplies a limited overlapping-admission observation, but no observation inside admission or injected atomic failure. Keep omitted initially rather than conflating reservation with full linearizability |
| `idempotency.concurrent-dependent-retry` | Bind now, Plan A |
| `idempotency.transient-not-retained` | Bind now, Plan A; pending-based variant only, rate-limit/service variants unqualified |
| `idempotency.principal-isolation` | Multiple actual authenticated principals needed; carried attribution is not an identity switch |
| `idempotency.authorization-view` | Actual policy/view changes needed |
| `idempotency.forbidden-retained` | Real operation-level denial then grant needed |
| `idempotency.retention-minimum` | Clock control plus public advertisement needed for full row; no short HTTP binding today |
| `idempotency.expired` | Implemented lower-level; needs controlled real-outcome expiry fixture for bounded HTTP evidence |
| `idempotency.expired-failure` | Same; failure forgetting must be distinguished from committed tombstones |
| `idempotency.restore` | Restore that loses recovery state/new logical Scope is not ordinary process reopen; no restore workflow exposed by helper |
| `validation.type-contract` | Current development Types have no property schemas; real installed schema/compiler fixture needed |
| `validation.diagnostics` | Actual schema validation/truncation fixture needed for full row; numeric/alias diagnostic fragments alone do not substitute |
| `validation.aggregate-constraint` | Actual multiplicity/configured contract fixture needed |
| `http.cors-idempotency-key` | Cross-origin capability deliberately disabled; no enabled-CORS observation possible |

Counts: 9 scheduling, 8 policy/principal, 5 retention/recovery, 4 admission, 3 ownership/aggregate, 2 schema, 1 contextual validation and 1 disabled CORS. These describe the current fixture limitation, not an assertion that every underlying evaluator lacks the behavior.

## Incomplete variants within the 48 bounded rows

Empty `limits` does not mean exhaustive qualification. Per-binding phrases such as “pins not yet bound” can describe an earlier plan even when a later binding in the same row supplies a pin observation; read the row's union. Do not count the same variant as both newly covered and missing.

The next table groups **all 48** existing rows and records actionable additions or limits. IDs again omit `read-update.`. No row is upgraded by this audit.

| Existing bounded rows | Feasible additional observation / remaining limit |
| --- | --- |
| `singleton.create-bead`, `alias.allocation-avoids-aliases` | Allocation/read-back already observed. Cannot force private random candidate collision through existing HTTP controls; collision skipping remains unqualified |
| `singleton.update-bead`, `singleton.document-patch-merge-replace`, `singleton.document-patch-atomicity` | Easy guarded combined properties/metadata no-op with revision and attribution unchanged; invalid second document proves neither changed. Array append/remove and null-vs-remove can be additional cases |
| `singleton.document-patch-validation` | Nonobject failure observed; effective property schema failure needs real schema fixture |
| `singleton.delete-bead`, `singleton.update-link`, `singleton.attribution` | Hidden incident Link, owned-source revision/no-op and attribution fanout require actual policy/owned contracts; unowned Link no-op can be bound now |
| `singleton.create-link` | Ordinary, alias and pinned endpoints now observed; neither historical pin resolution nor external fetch is promised by pin storage. Additional allocated-Link read-back is easy |
| `singleton.delete-link`, `singleton.expected-revision`, `singleton.expected-revision-race` | Add after-state checks/Link guarded no-op if useful; no distributed-writer claim from one reference process |
| `singleton.idempotency-key-required`, `singleton.unauthenticated`, `singleton.unsupported-media-type`, `http.request-too-large` | Existing raw cases cover many refusals. Existing source-only raw helper can bind hostile raw target/Host/duplicate Host/Origin tests with corrected same-key requests; this does not add auth-policy capability |
| `singleton.idempotent-retry`, `singleton.idempotency-conflict`, `idempotency.semantic-identity` | Easy defaults/object-order/exact numeric representation equality and changed guard/attribution/array-order conflicts. Persist original after-state, not only Problem code. Expiry/non-renewal still needs time fixture |
| `singleton.method-405`, `singleton.result-headers`, `http.cache-no-store` | More supported mutation methods and success/Problem headers can be bound. Directory and enabled CORS remain absent. Raw helper/Fetch method limitations should stay explicit |
| `singleton.no-local-bindings` | Pinned singleton `@name` refusal with no key capture is easy |
| `alias.put-creates`, `alias.put-repoints`, `alias.delete-releases`, `alias.no-version-minted` | No new omission inferred from empty limits; no-op put and revision-stability variants are optional small additions, not full-row qualification |
| `alias.uniqueness-invariant`, `alias.link-coexistence` | Explicit alias put at a previously deleted Bead segment; Link/alias same-segment coexistence and state checks are easy. Random allocation collision is separate |
| `alias.non-canonical-target`, `alias.unknown-subject` | Current alias/Link/external target rejection and unknown subjects observed. Hidden subjects need policy; fixed diagnostic message is not portable mandated wording |
| `alias.carrier-syntax`, `sequence.carrier-rejection`, `sequence.member-problem-shape` | Add duplicate keys/names, unknown/forward/wrong-kind labels and over32 carrier, with absent effects and corrected same-key retry. Broad Problem/ownership variants remain partial |
| `alias.sequence-binding`, `sequence.local-bindings` | More real slots (Bead update/delete, Link update/delete, alias put target) and transitive creators are easy. Pending variants are Plan A; expired bindings require time fixture |
| `alias.reference-resolution`, `alias.reference-idempotency` | Alias Bead update/delete subjects are not yet mapped; add turn-time repoint/miss then bind behavior using existing interleaving barrier. Recorded/fresh alias retries already observed for Link endpoints. Expiry, non-renewal and authorization remain missing |
| `sequence.order-and-partial-commit`, `sequence.retained-failure` | Existing partial commit/independent tail/retained failure observations; interleaving and transient tail additions are Plans A/B, not proof of transactional batch behavior |
| `sequence.idempotent-retry`, `idempotency.cross-carrier` | Reordered independent members with current indices and preserved dispositions, sequence→singleton reverse direction, and replay with renamed labels are easy. Do not reorder dependent edges into invalid syntax |
| `sequence.disconnect`, `idempotency.durable-boundary`, `idempotency.crash-mid-sequence` | Existing mid-execution disconnect/drain/between-member SIGKILL are genuine observations. No power-loss, storage fault, inside-commit crash, failover or full state/disposition atomicity qualification; new Plan A does not close these |
| `problem.table`, `validation.identity-taken` | Bind currently observable revision mismatch, binding-unavailable, type-not-installed, incident-links-exist, identity-taken and limits with status/family/retry/content-type; table still includes policy/schema/fault behavior not configured |

## Smaller follow-on if process scope is deferred

A pure HTTP contextual/retry plan can bind `sequence.contextual-validation` partially and fill existing-row gaps with no new action family: invalid label edges and duplicate keys/names before any effects/key claims, reverse cross-carrier replay, reordered independent members, defaults/object order and altered guards/attribution/array order. Keep set of cases bounded and report assertion references rather than claiming normative completion. Host/raw-target bindings can reuse the established source-only raw helper; do not weaken the portable executor's credential/Host/framing guard.

No discovery or full-profile advertisement should change for any of these diagnostic additions. The speculative simplicity catalog is not the contract for this work.
