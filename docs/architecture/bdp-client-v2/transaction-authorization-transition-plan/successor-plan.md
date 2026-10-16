# Private Scope authorization transition — executable successor plan

Plan only; root authoring release is still required. Original plan1a923751, actual Claude outputbcfbfc45, native29c91ccf and full source dispositions40781140 remain immutable. This successor incorporates their source-confirmed corrections and root's explicit Scope-wide/quiescent-publication decisions. Basis actual merged67a2842e3f1918e70978f745587cf11eaf0df5a8/tree596abe870220fb61ce606f8382a9da52d554d063. Proposed isolated author checkout /Users/dbox/repos/wt/vickie/vickie-authorization-transition, branchcodex/vickie-authorization-transition, starts clean at that commit. No product edits/tests/build/services were performed for this plan.

## One capability and its boundary

Implement a private trusted service operation that changes this Scope's immutable authorization snapshot from all-visible/mutation-enabled to empty/mutation-disabled, durably and without altering Scope epoch, Resource revisions, transaction positions or retained execution facts. The operation affects every authenticated principal in the Scope. A principal-aware capture interface remains, but no principal registry, wildcard fallback, per-principal policy map or public administration endpoint is introduced.

The service first stops NEW mutation admission, drains already-admitted finite mutation work under its established grant, then commits and publishes the policy. Only completed publication is revocation. Existing execution deadlines bound the drain; administrative caller cancellation or failed drain leaves the old policy and reopens admission unless the owner is independently faulted/closing. Do not extend those deadlines, roll back a committed prefix, fabricate cancellation receipts, or claim immediate revocation of in-flight writes. During the temporary admission pause use existing temporary-unavailable refusal without ownership/effects, not403 pretending revocation already occurred. After successful publication every new authenticated mutation, including same-key replay, is403; credential-level denial remains403 and unauthenticated behavior remains401. Ordinary request/Accept/key/control precedence otherwise remains unchanged.

The direct trusted owner replacement refuses while pending claims or active sequence attempts exist. Maintain exact bounded active-attempt accounting rather than an unbounded history or enumerable replacement for existing WeakMaps. A direct trusted mutation entry after successful replacement cannot bypass the owner policy. No wire request can supply a snapshot, policy name, permission override or administration capability.

## Private durable representation

Use new store format literal **bdp-private-transaction-snapshot-11**, existing identityPolicy **transaction-revision-provenance-4**, and numeric **policyStorageRevision:1**. Keep the installation configuration immutable: retain its existing field sequence and values except format; rename authorizationView to initialAuthorizationView (value development-all-visible), insert policyStorageRevision immediately after it. Use the existing transactionStorageText serializer, not JSON reformatting. Exact preimages depend on supplied Scope/Types/contracts/bounds and must be recorded in the measured draft.

Add one mandatory metadata record named **authorizationPolicy**, bounded to4096 UTF-8 bytes before loading/parsing. Closed own-data shape:
- revision:1
- bindingKind:"scope"
- generation:0 or1
- authorizationView: development-all-visible at generation0; one newly generated UUID wire token distinct from it at generation1
- readPolicy:"all-visible" at0, "empty" at1
- canMutate:true at0, false at1.

These two supported states are the first built-in policies. Replacing with the identical current semantic policy is a no-op preserving bytes/view/generation; empty→all-visible is unsupported by this private operation, not an invented protocol refusal. No regrant or arbitrary partial ACL claim. Generation/view must not be a credential digest or content-only reusable token. A later storage revision can add bindings/policies without rewriting retained graph/receipt/group facts.

New creation writes configuration and authorizationPolicy atomically with the existing epoch/position/capture records. New-format open requires a valid complete policy record; missing/oversized/unknown fields/inconsistent state faults, never defaults all-visible.

Migration accepts only exact format10 expected configuration bytes reconstructed with the ORIGINAL field order and serializer from the same supplied installation options. Validate identity codec, exact bounded contractCapture and installation preimage before migration. Atomically install initial policy record and format11 configuration before existing recovery/maintenance. Reject format9, foreign old view, changed bounds/Types/contracts/codec or incompatible capture. Preserve graph/group/receipt/page/Snapshot bytes, epoch and position across the migration itself; ordinary documented startup recovery/maintenance remains separately attributable. The old binary's exact configuration comparison rejects format11. Do not claim property lookup depends on JSON key position or run a shared production database.

Confirmed rollback before commit preserves old snapshot and leases. An uncertain COMMIT/unclean rollback/storage failure uses existing owner fencing and stops all delivery until reopen. A postcommit publication fault also fences/stops, never compensates to assumed old bytes. Add narrow private fault labels before-policy-commit and after-policy-commit to the existing fault-injection option for deterministic clean-rollback and committed-publication-failure tests; do not pretend these simulate every SQLite COMMIT failure.

## Reusable capture, projection and delivery seams

1. **One same-entry observation.** Immutable installation facts remain stable. Capture current policy+epoch+position inside the guarded owner entry; create a trusted principal-aware internal facet/reader capture that supplies matching sourceContext and read authorization. No reentrant owner call from authority-read.captureReadAuthorization. Reuse authority-read's existing per-entry principal/capture interface and keep epoch lifetime validation. Do not add a lifetime view-equality assertion or rebuild the Read plane on transition, which would lose issued cursor state. Scope policy resolves identically for both trusted principals; HTTP must stop passing anonymous placeholder principals after authentication.

2. **Complete private caller fanout.** Thread trusted identity/current snapshot through Resource/collection/alias reads; Snapshot create/preflight/manifest/page; Event finite/initial/next; Changes initial/subsequent pulls; receipt/root/page and mutation/sequence presentation; schema/discovery/refusal/deadline context-bearing observations. Bind principal identity on scan/lease creation; callers cannot switch it later. Keep authentication source unchanged. Effective mutation permission is credential.canMutate AND owner policy; a snapshot must never elevate a read-only credential.

3. **Retained authority facts versus current rendering.** Initial/legacy groups and receipts retain their original execution view. Under this monotonic policy all authority mutations execute before publication, so retained writer view must equal the recognized initial development-all-visible authority view. Continue validating epoch/position/ordinal/checkpoint/framing/endpoints; do not trust arbitrary stored group.view or strip integrity checks. Render current-view finite Events as empty. Render historical Changes groups as the existing normative projectionAdvance shape: current epoch/view, proper position/previousPosition/checkpoint, no transaction, zero eventCount and empty changes/erasures/events. Add a PRIVATE rendered-group union; normative schema already supports this and freezeChangefeedPage is generic. No schema/catalog/API expansion.

4. **Genesis precision.** Keep spec-reserved genesis for position0. Existing positive-position checkpoints already encode epoch/view and must retain that behavior for empty projection advances. The concrete old-checkpoint409 witness uses the actual positive-position Snapshot checkpoint after two seeds. Do not claim the untagged reserved genesis identifies a previous view. A CPU historical scan fromgenesis under the current empty view must reveal only current-view identifier-free advances. No new genesis spelling or broad token-policy decision.

5. **Snapshot integrity versus serving fence.** Validate retained dispositions/commitments under recognized stored context independently of the request's current view, including expire/liveLedger/capacity accounting. Validate requested id and real page-token legitimacy before semantic disclosure. A legitimate unexpired old-view handle returns409 foreign-view before body materialization; unknown/bad/undisclosable stays404; actual expiry under the same view stays410. Preserve existing expiry order where needed and keep simultaneous expired+foreign/minimum/hidden collisions outside witnesses. Add private SnapshotFailure foreign-view mapping to existing409 Problem. Fresh empty Snapshot inventory is empty, not raw SQL Resources with a changed token; retain old snapshots/capacity until actual expiry. Revocation is not fabricated expiration.

6. **Receipts.** Introduce one projection primitive used by root, pages, duplicate/synchronous presentation and sequence-member presentation. Preserve retained execution-body view/epoch/disposition/required/effect positions; HTTP context reflects the serving observation. Created/updated entries and allocated identities become withheld under empty policy. Nonrecord matched/deleted/erased/alias facts and failed Problem remain as retained. A completed receipt whose every result is withheld uses detail:withheld and omitsresults/next/allocated as specified. Already-issued available receipt pages remain200 with their original entry boundaries and projected withheld entries, not a Snapshot-style409; expiry410/nondisclosure404 remain. Never discard the projection result and then slice original records. Internal dependency evaluation may retain trusted execution facts; no such raw compatibility accessor becomes an HTTP exit.

7. **No stale stream registration.** Acquire a lease with the expected immutable snapshot under the same owner serialization used by transition publication. If handler selected an old packet before transition and registers afterward, lease acquisition rejects before headers/frames. Registered old leases abort synchronously after confirmed commit, before owner release. Existing pump checks the signal immediately before everywrite and nextpull; preserve backpressure/listener cleanup. Abort explains expected stream termination to runtime, avoiding firstStreamFault. Do not broadly swallow all foreign-view/errors. Bytes already accepted by Node/socket are not recalled; no instantaneous-client-silence claim.

8. **Metadata.** Separate launch-time installed configuration/payload binding from mutable policy. Current metadata capture must describe actual snapshot on reopen and reject unsupported/malformed state. Existing returned ready/configurationMetadata is a launch observation, not a perpetual current authorization attestation. This slice is source-runtime only and does not qualify an installed mutable-policy administration mechanism.

## Exact finite transport witness

Main callback planned title: **narrows current Transactional visibility and fences issued handles after policy publication**. One real owner/directory/listener, authenticated operator,17HTTP:
1 keyed batch32small Beads; save first receipt/position.
2 distinct keyed batch1Bead; save second actual position.
3 visible BeadGET200.
4 collectionGETlimit1; save actual next.
5 fresh SnapshotGET; save actual manifest id, checkpoint, beads.next; assert first32+next, no guessed URL.
6 finite Scope EventsGET; save actual emitted lastid.
7 Changes SSE start=now; await actual headers and registered lease.
Trusted administrative transition, zeroHTTP; require committed generation/view change and lease abort.
8 sameBead404.
9 collection200empty.
10 fresh Snapshot200empty/currentview.
11 old manifest409.
12 old actual Snapshot next409.
13 old collection next409.
14 Changes start=now plus Last-Event-ID from request5 actual checkpoint409.
15 Scope Events after actual request6id409.
16 firstseed receipt200withheld; body original required/effect position from firstseed, response current head from secondseed.
17 exact firstseed POST replay403; no new effect/group/position.

Firstseed has32inline results/no receipt page at default receiptMaximumItems32; secondseed has1. Snapshotpage32 is independent and really has a next at33Resources. Validate complete normative bodies/status/context before safe fingerprints. Bodies, keys, header ledger, schemas and byte ceilings must be materialized in authored preflight; no claim those bytes are already measured.

Second callback: **cancels an active Event stream after policy publication**, one owner/directory/listener,4HTTP: seed2; finite Events; live Event SSE withoutafter (eligible initial Events); after publication fresh finite Events200empty. This observes real stream registration/cancellation; deterministic queued-frame proof comes from CPU casebelow, not elapsed silence.

One control: suppress ONLY coherent successful policy publication in an isolated emitted/source derivative; do not mutate retained graph or falsely claim a committed policy state. Run main callback; require named **authorization-transition:generation-and-lease** failure after first7HTTP, max17, actual unchanged generation and uncancelled lease. It proves sensitivity to missing publication as a unit, not independent completeness of both branches. Same120s supervisor; no extrainstalled child/control cohort. Root must read exact derived mutation and execution graph before release.

## Planned CPU/source cases and owner accounting

The following are proposed single callbacks, not expanded parameter tables. Freeze actual titles/selectors/call counts from authored source before execution. All use real owner data, no fabricated read results.

| Case | Purpose | Successful owner opens | Dirs | Listeners | HTTP |
|---|---|---:|---:|---:|---:|
| T1 main |17-request schedule above|1|1|1|17|
| T2 activeEvent |4-request schedule above|1|1|1|4|
| T3 queuedFrame |Real2-Event packet, firstwrite(false), publish, drain; no secondwrite/ownerpull; abort/listeners clean|1|1|0|0|
| T4 staleLease |Select actual packet, publish, acquire with old snapshot; no headers/write/pull; new lease valid|1|1|0|0|
| T5 migrationNoopReopen |Exact legacy migration plus same-policy no-op/reopen identity; retained tables/epoch/position unchanged|3|1|0|0|
| T6 rollback |before-policy-commit fault; old policy/lease remain, no metadata mutation|1|1|0|0|
| T7 publicationFault |after-policy-commit fault; owner/leases fenced; reopen sees empty durable policy|2|1|0|0|
| T8 snapshotFences |Known unexpired oldmanifest/page409, unknown/badtoken404, separate same-view actualexpiry410; freshempty and maintenance valid|1|1|0|0|
| T9 receiptProjection |Actual mixed receipt inclrecord and nonrecord facts, real issued page via existing small page fixture, failedreceipt; empty projection preserves facts/boundaries/bodyhistory|1|1|0|0|
| T10 sequenceDrain |Actual scheduled sequence active when admin starts; newadmission paused, admittedwork drains/committedprefix facts preserved; publication only afterward; later receipt/member projection current|1|1|0|0|
| T11 scopeAndHistory |Two trusted principals share Scope snapshot; read-only credential not elevated; direct new mutation blocked; currentempty historicalChanges fromgenesis only advances|1|1|0|0|
| C1 control |Main sensitivity prefix|max1|1|1|7expected/17max|

Planned total12callbacks,15successful owner opens over12directories,3listeners,28expected/38maxHTTP, zero product children. Failed open attempts are separate from successful owners: T5 may perform exactly4 rejected opens using closed disposable copies for unsupportedformat9, changedlegacyoptions, missingnewpolicy and corruptnewpolicy; each has its own temporary copy directory, closed onfailure and deleted only after confirmed release. Thus reserve4 additional tempdirs/failedopen attempts, not4additional runningowners. Keep these exact; any necessary extra case/count must be reported before execution.

T5 constructs a closed legacyfixture from genuine retained facts and exactformat10preimage using data-only fixture setup, not an oldbinary launch or live database downgrade. Verify migration bytes separately from ordinary startup sweep/retraction. T7 exercises committed-publication failure, not an invented SQLitecommit-failure result. Existing atomic uncertain-commit logic is preserved and reviewed. T10 uses the existing scheduler seam and actual sequence results; it must not invent a wire cancellation contract. A drain failure/cancellation branch can be exercised within that samecase on a second admission attempt if deterministic, without a new owner/HTTP; exact callbacks and work remain in the measured sourceledger.

## Bounded implementation ownership

Production changes should remain confined to private Transactional owner/read/projector/Snapshot/service/runtime/HTTP/configuration modules and existing private submission/sequence error plumbing only if necessary. No public root/development exports, authority-read rewrite, protocol schema/catalog, Update harness or new general storage/policy framework. Expected corepaths:
transaction-store.ts, transaction-read.ts, transaction-events.ts, transaction-changefeed.ts, transaction-snapshot.ts, transaction-service.ts, transaction-runtime.ts, transaction-runtime-configuration.ts, transaction-http.ts.
Source tests use matching existing store/read/Snapshot/Event/SSE/service/runtime-configuration test modules or one focused transaction-authorization.test.ts if it reduces duplicated lifecycle fixtures. Progress documentation only; no normative spec edits. Before authoring release root must freeze the final path allowlist from this caller inventory; any necessary extra propagation path is reported explicitly rather than silently broadened.

Bounded private interfaces: capture current snapshot for a trusted principal within an existing owner entry; project retained receipt/member/page under that captured snapshot; acquire/release expected-snapshot delivery lease; replace Scope policy only when owner quiescent; service administrative replace that pauses/drains finite mutation admission and delegates replacement. Snapshot objects are opaque/frozen and owner-produced, never wire-deserialized authorization capabilities. Do not add a per-principal plane cache. Reuse existing AbortSignals, turn scheduler, transaction atomic helper, schemas and bounded raw transport.

## Validation and release

No source/test/build/install/service/Claude execution is authorized by this document. After root path release, author produces measured frozen diff, exact finite selectors/commands/body/header ledger, all owner/open/temp/HTTP counts and derivative graph. Existing pinned offline static/type/build/format/lint/boundary/Read checks precede source/runtime validation. Each command stays under120s; preserve existing request/stream/execution caps. Retain primary+cleanup failures, actual observations before cleanup, actual negative error and independently verify process/store/listener release. No auto retry, speculative time increase or hidden installedadmin channel.

Source-runtime proof is deliberately not fresh installed qualification. PublicRead-only admission stays held; S1/G2/unknown-position and collision choices remain untouched. No extra plan council loop is proposed: root adjudicates this concrete successor and path/count ledger, then independent native and actualClaude CODE review follow the frozen implementation.
