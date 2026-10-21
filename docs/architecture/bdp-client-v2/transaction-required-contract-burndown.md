# Transactional required-contract burn-down

Read-only planning, 2026-10-13. Runtime basis is frozen owned candidate `01e2d5470941a49240fab7820a6dea1b6deb291e`; integration reference is `b54392340de34c6c3e8aa07d959d7b6458338261`. The latter has not incorporated the frozen owned candidate yet, so owned behavior below is candidate evidence, not an integration claim. Spec, TX/RU catalogs and TX HTTP source are byte-identical between these inspected checkouts. No implementation, catalog edits, evidence relabeling, new semantics or public profile claim follows. No new test suite was run for this read-only inventory; quoted run counts belong to the frozen author reports. The sole fork design remains the plan of record; this memo is its completion inventory.

Integration update, 2026-10-13: current inspected root **c0d67388cd849410b38ec7ecb3a328771431dbfa** incorporates owned candidate **01e2d5470941a49240fab7820a6dea1b6deb291e**. The b543 non-integration statement above describes the original inventory baseline only; it is no longer the current integration status. Owned runtime is integrated candidate evidence, still not installed Transactional qualification. The remaining “review/integrate01e2” task in the historical ledger is superseded by this update. Event serving and materialization remain separately owned active lanes; G1 now has its own bounded finite-first plan in `transaction-changefeed-plan.md`.

## What completion still means

The current canonical selection factory returns **250 normative obligations:43 Read,74 Read+Update (including companions),133 Transactional** with optional History disabled. Twelve RU rows are retired by TX; Read's closed discovery row is also retired transitively. These are requirements, not250 execution results. `transactional-v1.json` by itself is not a cumulative claim. Diagnostics, source unit-test counts and protocol-schema examples cannot replace observations against the same installed service.

Current runtime evidence is substantial: one durable owner handles singleton/batch/sequence mutation, canonical dependencies, direct-carrier compare/admit, alias singletons/direct witnesses, receipts and durable pages, exact installed contract retention, same-owner Read context, bounded point/receipt minima, and the owned projection in01e2. The last owned run passed753 tests/18 files plus checks/smokes. It does **not** imply753 contracts or a completed Transactional profile.

The largest unconditional runtime holes are **replication Snapshot/changefeed serving, complete set targets/admission, remaining valid sequence alias forms, complete ordinary Read consistency coverage, and truthful cold discovery/directory**. Event serving and materialization are active separate lanes; this report does not re-plan or assign their files. The largest qualification hole is separate: **there is no installed cumulative TX execution carrier/manifest/fixture/evidence lane**, even for behavior already tested internally.

## Exact authority and evidence anchors

- `docs/specs/bdp.md`: scope history/authorization3652–3706; discovery3717–3833; transactions3835–4140; set mutation4215–4297; twelve operation targets4305–4381; batch admission4382–4625; receipts4945–5273; Events5350–5641; groups5642–5733; strict reads5734–5824; snapshots5825–5958; changefeed5959–6129; Event replay6130–6288; erasure6289–6524; validators/tokens6536–6586; normative TX index6800ff.
- `packages/conformance/catalog/transactional-v1.json`:133 exact normative rows. `src/profile-contract-selection.ts` is the cumulative selector; `src/transactional-catalog.test.ts` expressly tests catalog-only status and absence of executable TX manifest. `matrices/` has read-v1 and read-update-v1 only. `src/artifact-bundle.ts:109` admits the cumulative factory for RU only; this is an intentional gate, not a reason to trim TX selection.
- TX runtime: `transaction-batch.ts`, `transaction-store.ts`, `transaction-submission.ts`, `transaction-comparison.ts`, `transaction-sequence.ts`, `transaction-read.ts`, `transaction-consistency.ts`, `transaction-service.ts`, `transaction-http.ts`, `transaction-runtime.ts`, and corresponding transaction tests. These implement private emitted runtime evidence; public package deep-import guards stay intact.
- Installed reference: `installed-reference-configuration.ts:173` fixes profile read-update. `installed-reference.ts` binds actual payload/configuration and remains nonattesting. Existing installed RU observations cannot be renamed TX observations.
- Spec SHA256 `8c4ca0c0f2bdef1fea70465e84695c2b7975fa40b7ee9aec7d87329688bb345f`; TX catalog `d824e73b05a37ba6c7a66b3d97769dcb6b27fe0204310b0155b5c18a11e232d0`. Canonical selection was executed read-only from01e2's built module; no test/evidence artifact was generated.

## Runtime versus harness ledger

| Area | Actual runtime state at basis | Missing runtime/contract work | Qualification work even if runtime already works |
| --- | --- | --- | --- |
| Atomic Resource mutations/owned transitions | Durable compare/admit; same-transaction graph/group/receipt/pages; evaluator staging, guards, provenance/no-op; owned candidate two-Resource/four-Event seam | Review/integrate01e2; materialization lane's operational bound and compatibility work; nonempty aggregate-policy applicability not supplied by current fixture | Real installed direct/batch/sequence cases, concurrent same/different-key controls, rollback/owned after-state, exact ordinary/transaction-level failures |
| Sets | Private valid-set core selects staged visible canonical Resources, cardinality then count, reuses evaluator; same receipt/group owner | Normal prepareTransaction rejects sets; no updateWhere/deleteWhere HTTP target. Invalid Selector/complexity admission timing still unresolved; do not convert every failure to pre-key400 | Actual singleton/batch set runs, results/paging, lossless cardinality/identity, no-key-capture and retained-failure controls |
| Sequences/identity | Eight operation forms, canonical @name reservations/binding facts, pending/failure/retraction, direct unresolved comparison and shared keys | Resource-alias operands refused whole-carrier before effects (`transaction-batch.ts:195–235`); alias turn-time resolution versus admission identity needs the separately identified clarification | Actual HTTP competing presentations, pending joins, durable reservations/restart and alias/local-reference combinations; no fake pending receipt disclosure |
| Receipts/pages | Atomic paired outcome/current context, raw facts versus served prefix, durable page chain, expiry, replay, principal nondisclosure, min ordering | Mutable-view reauthorization, wholly withheld and erased detail unsupported; byte materialization remains active lane | Fresh packed HTTP root/page/HEAD/conditional/restart and same-key paths. Private raw full-results tests alone do not prove wire pagination |
| Read/consistency | Truthful same-owner facade; canonical Reads/list/aliases; original retained query-page context versus current serving context; complete canonical minimum tuple for visible Bead/Link/property points and receipts/pages | Minimum controls on list/continuation/Type/alias/discovery/other surfaces refused; malformed/partial/unknown-position categories and absent-point ordering not generally resolved. HTTP does not forward point disconnect signal; it drains a finite wait | Installed mutable Read selection, cold navigation, after-write state, point/receipt wait/conditional/context, future route-specific minima, HEAD parity |
| Event facts/source delivery | Owned/unowned facts persisted atomically, intermediate Events distinct from final changes | Active Event-serving lane supplies Scope/Resource finite+SSE, source-local IDs and historical membership; current scope/view/erasure cases remain explicit | Source/fact semantic assertions plus actual socket replay, reconnect/backpressure and erased/withheld branches when supported; Event SSE is not changefeed SSE |
| Changefeed | Durable groups contain final changes, ordered facts, predecessor/position and empty erasures | No finite/SSE serving, fenced checkpoint codec/validation, explicit starting intent, replay navigation or projected advances | Whole-group application client, token-crossing counterfactuals, duplicate reconnect, current context; show exact replica state rather than count Events |
| Scope Snapshot | Read query pagination exists; owner can observe coherent current rows | No first-class immutable manifest/two stream handle, checkpoint rendezvous, refetch/expiry promises or erasure ledger serving. Query cursors are not a substitute | Stage both streams, verify closure/inline agreement, apply suffix, publication only after complete bootstrap; HEAD creates no usable handle |
| Authorization projection | Actual TX owner fixed all-visible, empty aggregate policy; evaluator supports policy and owned closure | No installed mutable grants/view rotation, retained-copy filtering or identifier-free invisible-group projection. Decide supported fixture applicability explicitly before claiming these branches | Config-bound view changes across every delivery path, including sequence, receipt pages, Event and replication; token possession is not authorization |
| Erasure/restore | Empty erasures in groups; stable persisted epoch on normal restart; new store provisions new epoch | No supported erasure authority path/ledger/scrubbing/fence; no demonstrated restore/replacement control preserving erasure knowledge and fencing all token classes | Applicability review plus real supported trigger/recovery paths. Erasure is independent of optional History; unsupported triggers do not automatically earn passes/N/A |
| Discovery/Directory | Probe and private targets supplied to tests; actual bdp.json/operations return development gap; ten private mutation routes | Required TX discovery fields/navigation and full twelve-entry directory only when underlying mandatory surfaces are real; limits must match effective locked installation | Cold client starts only with Scope+credentials, learns profile/targets/limits and completes bootstrap/read/mutate/retry without path construction |
| Conformance/evidence |250-row selection and structural planning; existing RU installed binding/runner patterns; TX catalog metadata only | A source-private cumulative TX carrier/factory and actual installed target binding, then profile admission after all applicable obligations | Author observations/counterfactuals for each selected row, explain conditional applicability, persist fresh payload/source/config-bound evidence; preserve Read seal |

## Recommended executable order after the active lanes

**First finish/integrate Event serving and materialization without broadening them. Then prioritize G1→S1→D1, while a separately owned H1 harness lane and bounded M1/C1 planning can proceed.** G1/S1 yield the first real cold replica; this closes a larger foundational gap than more isolated mutation unit cases. Their publication still waits for all other applicable requirements.

### G1 — complete changefeed vertical slice

Dependencies: reviewed Event transport/lifecycle and projection ID policy; owned01e2 (its review is still underway); accepted materialization bounds. Reuse the authoritative groups table and owner entry, not inspect(), receipt reconstruction, Event last-ID or a second log. A checkpoint names an entire committed group and binds epoch/view; source Event IDs do not become durable replication cursors. Initial fixed all-visible/no-pruning implementation is a legitimate bounded configuration, with its unsupported projection/erasure variants named.

Likely files: new private transaction-changefeed module/tests, bounded transaction-store observation seam, service lifecycle, existing TX HTTP/runtime route/stream branch; shared Event helpers only through root's writer coordination. No public Read response/export change. A format/refusal decision is needed only if persisted representation actually changes, not merely for serving a derived checkpoint.

Done when actual HTTP finite and SSE deliver complete normative groups for genesis, exclusive after and start=now; omitted/conflicting start intent refuses; foreign/expired/unavailable checkpoints never skip; current context/minimum/HEAD/conditional decisions precede bodies; owned postimage agreement and no-op/alias absence are asserted; restart/reconnect after a durable applied checkpoint preserves group order; no replay/live boundary loss; slow/disconnected clients do not retain unbounded queues or block writes/shutdown. An independent test replica applies a whole group then advances its durable checkpoint. No claim that transport Last-Event-ID proves application.

### S1 — first-class Snapshot plus bootstrap/rendezvous

Dependency: G1 checkpoint/replay contract and owner read entry; Event/materialization work establishes reusable bounded serving, not Snapshot semantics. Reuse the existing graph validation/projection and bounded pagination engine where sound, but give Scope snapshots a distinct retained handle and two typed streams. Capture manifest identity, epoch/view/position/checkpoint/expiry, initial pages and projected erasure ledger at one prefix. Required empty erasures is valid only while no erasures exist; it is not proof of erasure support.

Likely files: new private transaction-snapshot module/tests; owner read/retained-handle seam; service/HTTP/runtime. Reuse authoritative installed descriptors. Do not combine with optional revision History or change collection cursor contracts. Bounded handles may be retained or reconstructable; uphold expiresAt across the supported restart/failover model, rather than diagnosing lost in-memory state as normal expiry before the promise.

Done when create GET provides a normative manifest; creation HEAD returns decisions/metadata with no usable handle and no renewal; id GET/HEAD reproduces same anchor/pages/expiry; independent Bead/Link continuations remain at that anchor despite concurrent writes; closed graph and inline/first-class records match; complete staged snapshot followed by G1 suffix reproduces current visible graph exactly; handle unknown/undisclosable404 versus authorized expired410 versus pre-expiry inability/service-failure is tested; checkpoint replay remains legal through promised snapshot expiry; same-entry context and conditional ordering hold. No silent page recomputation to newer rows.

### M1 — finish required set receiving and alias-bearing sequence forms

Two separable bounded slices, each preserving the current atomic owner. Set serving can use the existing valid core once receiving error/timing is settled; alias sequences should use per-member turn-time locator witnesses pinned to its actual attempt without weakening existing canonical dependency reservations.

Set files: transaction-set/batch/submission/HTTP/tests, operation table and receipts unchanged in principle. Done when both singleton set targets and batch sets use identical semantic identity and staged selection, all operation/member limits charge correctly, unsupported syntax/grammar/complexity categories have justified current protocol disposition and key timing, retained duplicates bypass fresh controls correctly, mixed ordinary/set/owned rollback and durable result paging pass over actual HTTP. Before implementation, resolve the precise Selector question: the request schema's nonempty string does not say whether grammar-invalid or too-complex Selector is pre-admission carrier400, retained operation failure, or another existing bound category. Do not invent a new code/limit or use Read's invalid-parameter by analogy.

Sequence files: batch/comparison/dependency/store/sequence tests, coordinated sole store writer. Done when a prior putAlias/repoint/delete in the same sequence affects a later fresh member at its turn; previous presentations preserve original target/miss witnesses and direct cross-carrier equality; competing owners and pending creators never leak/retract each other; alias change between durable reservation and execution follows the clarified rule; exact replay survives failed creator forgetting/reuse/restart. Canonical @name support already exists and is not to be rewritten. The current alias timing tension is a real normative clarification item, not permission for a new identity algorithm.

### C1 — complete settled Read/HTTP consistency surfaces

Dependency: G1/S1 add their own barriers; normal list/Type/alias/discovery controls can be a separate small seam. Preserve one original finite deadline, current final-attempt context, principal nondisclosure, no SQLite await, and original retained query-page state. Do not use later inspect or mutable last-position variables.

Files: transaction-read/consistency/store read methods and HTTP route tests; shared authority-read/pagination changes require explicit root scope. Done when every supported Scope-bounded read/HEAD/conditional route has tested context and canonical minimum handling, including list/Type/alias/discovery plus new feed/snapshot targets; unknown/foreign/expired controls have current normative disposition; 304/412 cannot bypass fences or invent checkpoint application. Keep these outstanding exact questions visible until authority settles them: partial/malformed minimum tuple general handling, unknown-but-well-formed opaque minimum, absent/hidden point versus minimum ordering, and a continuation whose original anchor precedes a new requested minimum. Do not adopt a before-materialization or continuation-refresh policy by inference. Existing finite HTTP wait drain is a limitation to report, not a new cancellation subsystem requirement.

### D1 — truthful installed cold discovery/directory

Dependencies: actual mandatory routes including G1/S1, sets and Event serving; locked materialization/retention/selector/page configuration. Return one canonical schema-valid discovery from the same owner entry with required epoch/view/head/minimumReplayPosition and live receipt/snapshot/changes/events navigation. Directory lists all twelve normed targets from the same operation map used for dispatch. No synthetic second Read listener or privately constructed target map as completion evidence.

Files: private TX runtime/service/discovery composition and installed TX fixture adapter, coordinated with root; no profile admission yet. Done when an actual installed client given only stable Scope+trusted credentials follows probe/discovery/directory, validates exact TX schemas, reads own writes, handles receipts/pages and bootstraps/replays through discovered URLs under a new listener/restart. All advertised limits reflect effective configuration; TX omits RU retention.idempotency. Discovery must not conceal still-missing mandatory behavior behind a complete-looking document.

### V1/E1 — explicit supported projection, erasure and restore boundary

Do not silently select these away and do not require a new public administrative API. First record which policy/erasure/restore triggers the intended installed fixture supports. Current all-visible immutable configuration permits an honest initial G1/S1; it does not provide mutable-view, hidden projection, erased-version or restore evidence. Catalog applicability unknown is still a gap.

If supported, partition implementation: (V1) one captured policy/view authority shared by reads, mutations and every retained receipt/sequence/page/Event/changefeed projection; then (E1) version erasure ledger and atomic publication/scrubbing, followed by restore preservation/fences. Reuse existing access policy/Resource closure after a concrete owner design; do not clone grants in each route. Erasure must account for all retained copies, including provenance, intermediate owned source versions and shared receipt result/page facts; digest before destruction with current normed sha-256-jcs. Source preimages captured transiently by01e2 are not a new durable version archive.

Done when changed policy rotates view and fences old handles/checkpoints while actual current delivery removes unauthorized details; invisible groups emit identifier-free projectionAdvance; every receipt delivery path and sequence binding obeys withholding. For supported erasure, correction/tombstone + ledger commit atomically, content disappears from all current/retained copies, source/Link cascades and publication fences work, Events do not invent an erasure verb, pre-erasure anchors expire without rotating epoch, and restore cannot resurrect known erased content. For supported replacement/restore, preserve canonical URLs but fence every history-token class and use fresh key namespace. Optional History remains independently off unless selected. These are larger audited slices after cold replication, not one omnibus patch.

### H1/H2 — cumulative installed execution and final admission

H1 can begin as read-only mapping/manifest authoring against the250 selected rows while runtime owners work. Use the existing canonical selection and exact served-configuration binding. Implement a reviewed source-private TX carrier/installed factory, not an unguarded profile parameter to the RU runner. Preserve Read/public guards, profile retirements and same-target identity. Versioned evidence changes require separate review; this memo introduces none.

H1 done when a fresh packed target reports actual payload/source/Scope/profile/contracts/policy/authentication/fixture bindings; plans retain every missing selected row; executable positive/negative actions observe real target responses and after-state; pure metadata/schema tests cannot emit pass. Real callbacks/faults may arrange service state but may not fabricate replies/positions. Port the applicable RU observations to receipt/shared-sequence forms rather than replaying incompatible RU inline-output assertions.

H2 done when all applicable250 obligations have reviewed observations or explicit justified conditional applicability; missing harness and missing runtime are both zero for the accepted configuration; source/payload/configuration are exact, all counterfactuals fail for the intended reason, restart/pending/receipt/replication/owned tests are from the same target; fresh evidence passes the unchanged Read seal constraints and the separately reviewed TX admission gate. Until then retain nonattesting markers. No percentage from test counts, no row retirement to accommodate a gap, no automatic claim from a cold smoke alone.

## Decisions and nondecisions

The settled next implementation work is G1/S1; it needs implementation review, not a new protocol ruling. M1 Selector/sequence-alias timing and C1 ambiguous minimum combinations require exact clause reconciliation/ruling before encoding those cases; they do not block settled replication work. Supported policy/erasure/restore applicability is a configuration/scope decision with explicit evidence consequences, not a reason to pretend those rows passed. Optional History, optional CORS and optional finite validators need no implementation solely because TX is selected; if enabled, their contracts apply. Finite Event/receipt validation omission is itself tested behavior.

No generic mutation refactor, new expression language, Type lifecycle, alternate store, command-line architecture or new log format belongs in this burn-down. Active materialization work owns the operational tuple and count/byte proof. Current owned assumptions are at most two Resource effects and four Events per local owned creation/deletion; no source preimage body is emitted. The initial high-fanout4,096 test was inconclusive and does not establish a diagnosed runtime performance defect; use only the final small-body count observation for that boundary.

## Exact TX-row routing appendix

The following is an obligation-to-work-lane inventory, **not coverage status**. Every row still needs H1/H2 evidence. `P` means bounded current private behavior exists; `M` means missing serving/admission despite private core; `G/S/D` means missing replication/discovery runtime; `V/E/R` means projection/erasure/restore work or explicit supported-trigger applicability; `F` means active Event delivery, with current facts separately tested; `C` means cross-route consistency/HTTP coverage is incomplete. A P row may still have unsupported branches. No row is marked passed or N/A here.

| Exact catalog ID | Lane / actual remaining distinction | Normative anchor |
| --- | --- | --- |
| `transactional.discovery.document` | D1: absent cold discovery/directory; align materialization/retention limits | `#discovery-profile-membership` |
| `transactional.discovery.operation-directory` | D1: absent cold discovery/directory; align materialization/retention limits | `#operation-directory-and-singleton-targets` |
| `transactional.discovery.limits` | D1: absent cold discovery/directory; align materialization/retention limits | `#transactional-limits` |
| `transactional.discovery.no-idempotency-retention` | D1: absent cold discovery/directory; align materialization/retention limits | `#transactional-limits` |
| `transactional.batch.atomic-commit` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-transactions` |
| `transactional.batch.local-references` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#batch-operation-target` |
| `transactional.batch.rollback` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.batch.label-errors` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#batch-operation-target` |
| `transactional.batch.envelope-errors` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#batch-operation-target` |
| `transactional.batch.pre-admission-precedence` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#batch-operation-target` |
| `transactional.batch.transaction-limits` | P + active materialization/H1: supported counts, no general byte proof yet | `#set-mutation` |
| `transactional.batch.duration-limit` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.batch.deletion-safety` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#explicit-bead-operations` |
| `transactional.batch.owned-link-transitions` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#owned-links` |
| `transactional.batch.no-op-entries` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#change-groups-and-replication` |
| `transactional.batch.revision-guard-race` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-revisions-and-guards` |
| `transactional.batch.aggregate-constraint-race` | P/V1 + H1: serial owner; supported nonempty aggregate fixture still needed | `#aggregate-constraint-enforcement` |
| `transactional.set.mutation` | M1: private valid core; required HTTP targets/admission missing | `#set-mutation-objects` |
| `transactional.set.zero-match` | M1: private valid core; required HTTP targets/admission missing | `#events-and-event-sources` |
| `transactional.set.cardinality` | M1: private valid core; required HTTP targets/admission missing | `#set-mutation` |
| `transactional.set.singleton-targets` | M1: private valid core; required HTTP targets/admission missing | `#batch-operation-target` |
| `transactional.set.attribution-fanout` | M1: private valid core; required HTTP targets/admission missing | `#set-mutation` |
| `transactional.singleton.receipt` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#transactional-operation-directory-and-singletons` |
| `transactional.singleton.semantics` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#transactional-operation-directory-and-singletons` |
| `transactional.receipt.readable` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipts` |
| `transactional.receipt.pagination` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.receipt.page-expiry` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.receipt.nondisclosure` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipts` |
| `transactional.receipt.root` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.receipt.headers` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.receipt.reauthorization` | V1/E1: fixed all-visible receipts only | `#mutation-receipt-responses` |
| `transactional.receipt.reauthorization-paths` | V1/E1: fixed all-visible receipts only | `#mutation-receipt-responses` |
| `transactional.receipt.non-record-entries` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.receipt.whole-withheld` | V1/E1: fixed all-visible receipts only | `#mutation-receipt-responses` |
| `transactional.receipt.pending-202` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#batch-operation-target` |
| `transactional.receipt.deleted-identity` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.receipt.owned-link-source` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.receipt.problem-shape` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.receipt.transaction-level-failure` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.receipt.validation-diagnostics` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.receipt.no-transient-failure` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#transactional-problem-details` |
| `transactional.receipt.expiry-allocated` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipts` |
| `transactional.receipt.retention-minimum` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.idempotency.concurrent-join` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-transactions` |
| `transactional.idempotency.conflict` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-transactions` |
| `transactional.idempotency.semantic-identity` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-transactions` |
| `transactional.idempotency.cross-carrier` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-transactions` |
| `transactional.idempotency.key-grammar` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-transactions` |
| `transactional.idempotency.expired-detail` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.idempotency.transient-abort` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#batch-operation-target` |
| `transactional.idempotency.failed-retained` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-transactions` |
| `transactional.idempotency.durable-admission` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-transactions` |
| `transactional.idempotency.ownership-fencing` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-transactions` |
| `transactional.idempotency.pending-recovery` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-transactions` |
| `transactional.idempotency.one-key-state` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-transactions` |
| `transactional.sequence.member-transactions` | P + M1/H1: canonical dependency core; remaining alias forms/installed evidence | `#mutation-transactions` |
| `transactional.sequence.completed-projection` | P + M1/H1: canonical dependency core; remaining alias forms/installed evidence | `#mutation-transactions` |
| `transactional.sequence.in-flight-projection` | P + M1/H1: canonical dependency core; remaining alias forms/installed evidence | `#mutation-transactions` |
| `transactional.sequence.expired-projection` | P + M1/H1: canonical dependency core; remaining alias forms/installed evidence | `#mutation-transactions` |
| `transactional.sequence.pending-creator` | P + M1/H1: canonical dependency core; remaining alias forms/installed evidence | `#mutation-transactions` |
| `transactional.sequence.failed-creator` | P + M1/H1: canonical dependency core; remaining alias forms/installed evidence | `#mutation-transactions` |
| `transactional.sequence.member-reauthorization` | V1/E1 + M1: withheld/erased projection absent; alias forms incomplete | `#mutation-transactions` |
| `transactional.event.owned-link-delta` | F: Event delivery active; stored facts are not serving evidence | `#events-and-event-sources` |
| `transactional.event.owned-link-update-delta` | F: Event delivery active; stored facts are not serving evidence | `#events-and-event-sources` |
| `transactional.event.owned-link-deleted-identity` | F: Event delivery active; stored facts are not serving evidence | `#events-and-event-sources` |
| `transactional.event.fact-order` | F: Event delivery active; stored facts are not serving evidence | `#events-and-event-sources` |
| `transactional.event.no-op-owned-update` | F: Event delivery active; stored facts are not serving evidence | `#events-and-event-sources` |
| `transactional.event.graph-facts` | F: Event delivery active; stored facts are not serving evidence | `#events-and-event-sources` |
| `transactional.event.self-link` | F: Event delivery active; stored facts are not serving evidence | `#events-and-event-sources` |
| `transactional.event.set-expansion-order` | F: Event delivery active; stored facts are not serving evidence | `#events-and-event-sources` |
| `transactional.event.history-expired` | F + E1/V1: Event delivery active; stored facts are not serving evidence | `#events-after-resource-deletion` |
| `transactional.event.cursor-after-withheld` | F + E1/V1: Event delivery active; stored facts are not serving evidence | `#version-erasure` |
| `transactional.changefeed.group` | G1: durable group core, absent feed serving | `#scope-changefeed` |
| `transactional.changefeed.wire-form` | G1: durable group core, absent feed serving | `#scope-changefeed` |
| `transactional.changefeed.sse` | G1: durable group core, absent feed serving | `#scope-changefeed` |
| `transactional.changefeed.start-intent` | G1: durable group core, absent feed serving | `#scope-changefeed` |
| `transactional.changefeed.owned-link-agreement` | G1: durable group core, absent feed serving | `#change-groups-and-replication` |
| `transactional.changefeed.consistency-fields` | G1: durable group core, absent feed serving | `#transactional-http-consistency` |
| `transactional.changefeed.catch-up-timeout` | G1: durable group core, absent feed serving | `#transactional-problem-details` |
| `transactional.changefeed.replay-window` | G1: durable group core, absent feed serving | `#snapshots-and-strict-reads` |
| `transactional.changefeed.sse-reconnect` | G1: durable group core, absent feed serving | `#scope-changefeed` |
| `transactional.changefeed.projection-advance` | G1 + V1: durable group core, absent feed serving | `#change-groups-and-replication` |
| `transactional.changefeed.owned-closure` | G1 + V1: durable group core, absent feed serving | `#owned-links` |
| `transactional.snapshot.rendezvous` | S1: absent first-class Snapshot serving | `#scope-snapshots` |
| `transactional.snapshot.closed-projection` | S1: absent first-class Snapshot serving | `#scope-snapshots` |
| `transactional.snapshot.expiry` | S1: absent first-class Snapshot serving | `#scope-snapshots` |
| `transactional.snapshot.erasure-ledger` | S1 + E1: absent first-class Snapshot serving | `#version-erasure` |
| `transactional.erasure.record-digest` | E1: absent authority path; supported-trigger applicability unresolved | `#version-erasure` |
| `transactional.erasure.digest-domain` | E1: absent authority path; supported-trigger applicability unresolved | `#bdp-json-and-http-protocol` |
| `transactional.erasure.historical-version` | E1: absent authority path; supported-trigger applicability unresolved | `#version-erasure` |
| `transactional.erasure.live-successor` | E1: absent authority path; supported-trigger applicability unresolved | `#version-erasure` |
| `transactional.erasure.live-tombstone` | E1: absent authority path; supported-trigger applicability unresolved | `#version-erasure` |
| `transactional.erasure.owned-link-cascade` | E1: absent authority path; supported-trigger applicability unresolved | `#version-erasure` |
| `transactional.erasure.event-withholding` | E1: absent authority path; supported-trigger applicability unresolved | `#version-erasure` |
| `transactional.erasure.graph-facts` | E1: absent authority path; supported-trigger applicability unresolved | `#version-erasure` |
| `transactional.erasure.receipts` | E1: absent authority path; supported-trigger applicability unresolved | `#mutation-receipt-responses` |
| `transactional.erasure.replication-invalidation` | E1: absent authority path; supported-trigger applicability unresolved | `#version-erasure` |
| `transactional.erasure.retention-non-propagation` | E1: absent authority path; supported-trigger applicability unresolved | `#version-erasure` |
| `transactional.erasure.restore` | E1: absent authority path; supported-trigger applicability unresolved | `#version-erasure` |
| `transactional.erasure.unestablishable-content` | E1: absent authority path; supported-trigger applicability unresolved | `#version-erasure` |
| `transactional.erasure.view-projection` | E1: absent authority path; supported-trigger applicability unresolved | `#version-erasure` |
| `transactional.erasure.reads` | E1: absent authority path; supported-trigger applicability unresolved | `#reads-after-deletion` |
| `transactional.erasure.epoch-unrotated` | E1: absent authority path; supported-trigger applicability unresolved | `#version-erasure` |
| `transactional.http.status-matrix` | C1 + H1: existing routes tested; cross-surface obligation still incomplete | `#batch-operation-target` |
| `transactional.http.method-405` | C1 + H1: existing routes tested; cross-surface obligation still incomplete | `#problem-details` |
| `transactional.http.problem-table` | C1 + H1: existing routes tested; cross-surface obligation still incomplete | `#transactional-problem-details` |
| `transactional.http.problem-contexts` | C1 + H1: existing routes tested; cross-surface obligation still incomplete | `#transactional-problem-details` |
| `transactional.http.internal-fault` | C1 + H1: existing routes tested; cross-surface obligation still incomplete | `#batch-operation-target` |
| `transactional.http.disconnect-admitted` | C1 + H1: existing routes tested; cross-surface obligation still incomplete | `#mutation-transactions` |
| `transactional.http.timestamps` | C1 + H1: existing routes tested; cross-surface obligation still incomplete | `#bdp-json-and-http-protocol` |
| `transactional.http.ijson-strings-objects` | C1 + H1: existing routes tested; cross-surface obligation still incomplete | `#json-mutation-admission` |
| `transactional.http.token-profile` | C1 + H1: existing routes tested; cross-surface obligation still incomplete | `#event-id-and-checkpoint-character-profile` |
| `transactional.restore.epoch-fence` | R / E1: restart exists; replacement/restore trigger evidence missing | `#scope-history` |
| `transactional.restore.key-namespace` | R / E1: restart exists; replacement/restore trigger evidence missing | `#mutation-transactions` |
| `transactional.alias.receipt` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#transactional-alias-mutations` |
| `transactional.alias.locator-only` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#transactional-alias-mutations` |
| `transactional.alias.no-batch` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#transactional-alias-mutations` |
| `transactional.sequence.withheld-allocation` | V1/E1 + M1: withheld/erased projection absent; alias forms incomplete | `#mutation-transactions` |
| `transactional.sequence.withheld-binding` | V1/E1 + M1: withheld/erased projection absent; alias forms incomplete | `#mutation-transactions` |
| `transactional.erasure.live-publication` | E1: absent authority path; supported-trigger applicability unresolved | `#version-erasure` |
| `transactional.erasure.disconnect-race` | E1: absent authority path; supported-trigger applicability unresolved | `#version-erasure` |
| `transactional.idempotency.unresolved-admission` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-transactions` |
| `transactional.idempotency.unresolved-comparison` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-transactions` |
| `transactional.idempotency.retraction-retry` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-transactions` |
| `transactional.erasure.persistent-event-consumer` | E1: absent authority path; supported-trigger applicability unresolved | `#event-replay-and-live-observation` |
| `transactional.snapshot.manifest-refetch` | S1: absent first-class Snapshot serving | `#scope-snapshots` |
| `transactional.snapshot.handle-refusal` | S1: absent first-class Snapshot serving | `#scope-snapshots` |
| `transactional.receipt.page-restart` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.http.accept-refusal` | C1 + H1: existing routes tested; cross-surface obligation still incomplete | `#problem-details` |
| `transactional.receipt.validator-omission` | P + H1: bounded private owner/HTTP evidence; cumulative installed row unobserved | `#mutation-receipt-responses` |
| `transactional.http.finite-validator-omission` | C1 + H1: existing routes tested; cross-surface obligation still incomplete | `#receipt-and-finite-feed-http-validators` |
| `transactional.http.conditional-reads` | C1 + H1: existing routes tested; cross-surface obligation still incomplete | `#conditional-reads-and-head` |
| `transactional.http.head-parity` | C1 + H1: existing routes tested; cross-surface obligation still incomplete | `#conditional-reads-and-head` |

Appendix integrity:133 distinct IDs copied from the exact TX catalog; no omission, invented row or retirement. The117 inherited selected obligations remain in H1/H2 and the Read/Read+Update source catalogs. Their required forms must be observed on this TX target; the canonical selector—not this appendix—owns their identities and retirements.
