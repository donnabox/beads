# W2: first installed sequence conformance slice

Read-only inventory against frozen W1b `f2ce17bafd1c3d218cf404b815e12e288983fe05` in `vickie-read-update-w1b`. This is a bounded execution proposal under the sole `read-update-next-execution-plan.md`, not a competing product plan. Wait for the reviewed W1 low-finding successor as implementation base; do not modify either frozen source. No tests, source edits, qualification generation, or outward actions were performed for this inventory.

## Recommendation and scope

Implement **member-problem-shape, local-bindings, and carrier-rejection**, with the ordinary order/partial-commit case expressly partial. These are contiguous sequence carrier/member behaviors available in the installed service today. Award each whole flag only after its listed clauses and negative controls actually pass and are reviewed; no predicted quota. Defer scheduler, failure injection, retention-clock, custom schema and new Type work.

The baseline has 129 required contracts and W1b records 36 bounded / 28 whole / 101 uncovered, subject to the separately assigned inherited W1 coverage correction. Recompute from that successor, not these historical totals. Unfiltered missing rows must remain visible and claimEligible remains false.

## Existing installed support and real seams

| Surface | Current source evidence | Consequence |
|---|---|---|
| Installed sequence route | `development-runtime.ts:450–464` rejects a carrier Idempotency-Key, then invokes the same principal authority; `development-authority.ts:309–316` prepares and submits a sequence to its actual lifecycle. | No new runtime route or bypass is needed. Operation Directory already discovers sequence. |
| Limits | `development-authority.ts:40–46` binds sequenceOperations 32 and bodyBytes 65536 in the same immutable owner tuple; installed configuration hashes/consumes it. `sequence-executor.ts:589` refuses a longer prepared carrier before admission. | Observe discovery's exact 32, submit 33 small members below the byte cap, expect 413 limit-exceeded. Do not confuse body-size refusal with member-count proof. |
| Static local references | `read-update-values.ts:307–350` rejects duplicate keys/names and forward/unknown/wrong-kind names; `read-update-preflight.ts` does pure Scope syntax before owner admission. | Wrong-kind @name is a carrier 400, not a member resource-not-found. Names are allowed on Resource creates only. |
| Ordered durable execution | `sequence-executor.ts:480–519` obtains each terminal turn/creator fact before scheduling the next. Resource evaluator reports identity-taken (409), and the same lifecycle retains dispositions. | Existing runtime supports the proposal. Source inspection is not installed schedule evidence. |
| Existing actual smoke | `scripts/installed-reference-observation.mjs:261–280` already sends sequence success / missing subject / independent success and observes prefix/tail GETs. | Useful request starting point, not coverage credit. |
| Exact client | `apps/bdptest/test-support/read-update-exact-client.ts` currently requires idempotencyKey on every body-bearing POST, and matrix's explicit client POST routes omit sequence. Body bindings only patch top-level fields. | If used for positive sequence requests, narrowly permit explicit omission, add the exact sequence route, and add the unchanged normative sequenceResponse wrapper to installed-body-schema.ts. Do not add nested templating. |
| Existing raw lane | `executable-manifest.ts:62–81` already admits literal UTF-8/base64 source bodies and exact header lines; runner.ts:1024–1033 sends those bytes; matrix builds exact routes from native actions. | Reuse native raw actions for literal carrier negatives, including duplicate decoded JSON members. No raw-body DSL, protocol/API extension, or client callback into the owner is necessary. |

Prefer literal sequence bodies with fixed fixture IDs and literal Scope-local Type URLs whose identity is first checked against the discovered/captured Type inventory. For allocated creator IDs, subsequent positive sequence members use literal @name; capture the resulting identity and use the existing canonical dynamic Bead GET. This needs no nested body substitution. Current native assertions can use the normative sequenceResponse directly. If retaining the client-wrapper lane simplifies its existing counterfactual controls, the small optional-header and wrapper additions above are sufficient; choose one lane consistently for positive cases rather than building a new adapter framework.

## W2a cases and independent oracles

### 1. `read-update.sequence.member-problem-shape` — whole candidate

One actual 200 response contains (a) successful named Bead creation, (b) named creation refused because its explicit ID is already committed, (c) unnamed update of a missing Bead, and (d) later independent unnamed success. Seed collision independently and capture its full record. Assert the complete normative sequenceResponse, exact results length and zero-based operationIndex, and operationName if and only if the request declared a name. Names cannot be fabricated on updates/alias members just to increase coverage.

Check static status/type/code/retry for identity-taken (409) and resource-not-found (404), not just generic schema validity or a reused expected object. Independent GETs verify both successes, missing subject, and unchanged collision record; compare returned created records with their independent GETs while still current. Cold requests use the discovered operation URL and actual HTTP; fixture state is never supplied as a synthetic result.

Counterfactuals after real HTTP: wrong index, missing/extra/wrong name, wrong would-be status/code, and a missing or swapped member. Require exact intended assertion failures and one application. Do not count top-level 200 as success of failed members.

### 2. `read-update.sequence.local-bindings` — whole candidate; retained-failure separately partial

Use an allocated named Task create, followed by an update of @fresh that changes a literal property, then an independent tail. Assert the dependent's canonical Resource ID equals the creator's allocated ID; GET that identity and verify the static final properties and tail. The creator's retained creation postimage legitimately predates the dependent update: never require that old creation response to equal the current GET after later mutations.

Exercise a failed named creator plus dependent and independent tail. The simple identity-taken setup supports the immediate failure branch. To make the retained-failure branch discriminating without schemas, prefer the already-supported live-alias collision: put alias/x, then named create with id beads/x (alias-path-taken), dependent update @blocked (binding-unavailable), independent tail. Delete alias/x through real HTTP; now the same create would succeed under a fresh key. Replay original member keys and verify both failures stay retained, no beads/x appears, and the successful tail's complete prior record is unchanged. A fresh independent creation proves the cleared path is usable; do this only after the retained replay observation.

Expected dependent problem is the closed binding-unavailable shape (400/never). Assert the failed dependent did not mutate the alias target or create any Resource. Counterfactuals: wrong binding identity, dependent spuriously succeeding, altered retained failure, changed tail state. Use static expected afterstate plus independent reads, not replay-body equality alone.

This observes fresh and retained creator/dependent bindings; expired and transient creators remain separate contracts. It does not by itself close retained-failure's re-keyed creator/dependent semantic-identity clause. A later W2b may show fresh creator key + old dependent key conflicts, then both fresh keys succeed; do not silently count that unexecuted branch.

### 3. `read-update.sequence.carrier-rejection` — whole candidate only with the complete refusal set

Use bounded independent literal variants, each with at least one otherwise valid uniquely named prefix create and valid key whenever meaningful:

- 33 small members versus advertised maximum 32; also successful exactly-32 boundary. Verify JSON body remains below 65536 bytes and all keys are distinct.
- duplicate member keys; duplicate creator names; invalid key syntax and invalid name syntax; a name on a noncreation member.
- forward, unknown, self and wrong-kind @name references, including Bead name used as a Link subject and Link name used as a Bead subject/endpoint.
- forbidden top-level Idempotency-Key HTTP field on sequence; do not confuse that field with a JSON property.
- closed-envelope/member errors: empty/missing operations, unknown root/member field or operation, malformed JSON; duplicate decoded JSON member names and unpaired-surrogate escapes use the existing raw UTF-8/base64 body facility. These support the actual sequence-envelope antecedent, not just the abbreviated catalog excerpt.

Every refusal must validate the complete direct Problem Details response (400 malformed-request, or 413 limit-exceeded for the count case), absence of a results envelope, and no effect. GET every otherwise admissible newly requested identity and verify preexisting endpoints/aliases unchanged where relevant. Then submit corrected changed semantics under the original valid member keys and prove success: unchanged retry alone could conceal accidental key retention. Invalid keys themselves cannot be reused; valid sentinel keys prove pre-admission behavior for those carriers. For 33-member overflow, verify all 33 absent, then correct with 32 plus the final singleton using the same keys, changed marker properties, and independent GETs. Do not use only the first member's absence as proof of no tail effects.

Counterfactuals: fake successful prefix GET, wrong error/status, retained key conflict on corrected retry, and count-limit request accidentally below the boundary. These are assertion-input sensitivity controls after actual wire observations, not additional service observations.

### 4. `read-update.sequence.order-and-partial-commit` — partial

Give this row its own bound assertions or deliberately reference the exact checked prefix/failure/tail observations according to the current one-row binding mechanism. Use a successful create/update before a failed member and a later successful independent create; GET both committed states. This witnesses retained earlier work and continuation. Final response order and dependency success do **not** establish that all members never execute in parallel or that commits were externally visible immediately between members. Keep wholeContract false; do not add a scheduler hook to the installed service for this slice.

## Explicitly uncovered or partial after W2a

- sequence.transient-predecessor, expired-creator-binding, retry-after-hint: require actual controlled pending/temporary or expiry conditions; ordinary successful requests do not create them. No installed test-only clock, eviction, scheduler, or synthetic problem injection is approved here.
- sequence.retained-failure: the optional alias-clearing observation above is useful, but the dependent re-key identity rule still needs its own observation.
- sequence.idempotent-retry: same-position replay alone leaves present-position/name redecoration and recovery semantics incomplete. A later finite replay can prepend a new independent member, rename creator and matching @name references consistently, and assert historical disposition payloads with new index/name. Do not accidentally change creator keys when testing mere name/index changes.
- sequence.contextual-validation: W2's static @name refusal is not the entire contract (durable kind/Scope, ownership, alias resolution and request/result correspondence remain). W1b owner results do not automatically close this sequence row.
- sequence.disconnect, internal-fault, interleaving and idempotency.crash-mid-sequence: require actual lifecycle or fault probes beyond W2a; no coverage from the existing source-only diagnostics.
- alias.sequence-binding: may be a later bounded follow-on; using an alias as setup does not cover an alias sequence member.
- schema/update-bead/document-patch-validation remain W3's deliberately captured schema configuration; no Type or schema policy change is needed in W2a.

No unavoidable new product ruling or observed runtime defect was identified by this source inventory. The missing scheduling/retention/fault observations are qualification gaps, not inferred protocol gaps. Any actual installed mismatch must be reported with request, response and afterstate before changing runtime or weakening the contract.

## Minimal implementation and configuration

Required artifact/test files: matrices/read-update-v1.json, fixtures/read-update-reference-v1.json, test-support/read-update-coverage-v1.json, conformance/src/cumulative-artifact.test.ts, apps/bdptest/src/read-update-matrix.test.ts, and docs/design/read-update-executable-progress.md. Paths under packages/conformance unless prefixed otherwise.

Only if the existing client lane is used for positive sequences: apps/bdptest/test-support/read-update-exact-client.ts plus its focused test, and installed-body-schema.ts for a cloned normative sequenceResponse wrapper. Preserve singleton missing-key refusal by the actual authority; helper omission must never insert a fabricated key. Keep exact Scope, credential handles, discovered operation agreement and body/header bounds; no wildcard routes or relaxed destination guards.

The fixture can retain the same installed factory, policy/authentication, stable Scope, Type contracts, lifecycle and advertised limits. New route/fixture/manifest bytes must change their own bound identities and reviewed exact-configuration pin. Do not relabel previous observations or change installed policy to make tests pass. Native raw actions already permit finite source bytes; no new generic raw API is warranted.

## Done when

1. All proposed requests reach the same actually installed service/configuration through existing exact HTTP; no owner calls, storage reads/edits or lifecycle test controls substitute for observations.
2. Reviewer confirms the three potential whole rows against every cited clause and the complete envelopes; partial/missing rows stay visible with named reasons. Recompute totals from accepted W1 successor; pin exact IDs, not just count.
3. No-effect/key-reuse, independent final-state and replay discriminators pass. Each negative control fails exactly its intended assertion without corrupting the underlying service. Cleanup completes on failure and abort.
4. Clean/typecheck/build precedes emitted tests; focused matrix/client/cumulative/installed regressions pass with pinned Node24.16/pnpm11.20. Existing timeout budgets remain bounded; adjust only a justified aggregate test budget if expanded actual work requires it, never inner HTTP timeouts or assertions.
5. Run a fresh packed installed observation at the final clean source, with exact payload/config/fixture/catalog bindings; report omissions and claimEligible:false. Preserve public Read admission and sealed evidence unchanged. Root owns review, integration and outward publication.

## Approved execution update after native and external plan review

Working successor: `vickie-read-update-sequence`, based on frozen W1b repair `c1a17442ff2f5936616e7644e827dfa28dcd5ab7` (combined3e2 ancestry). Baseline129required,36bounded/93omitted,27proposedwhole/102uncovered, false claim eligibility. The four new IDs each have their own same-ID scenario and assertions; there is no cross-scenario binding credit. `local-bindings`, `order-and-partial-commit` and `carrier-rejection` stay bounded-observation/wholeContract:false. Only member-problem-shape remains a whole candidate after observed assertions and controls; no quota or anticipated qualification.

External H1–H4: each row owns its HTTP; positive sequence cases use the existing client lane with actual Directory `/body/sequence` capture and explicit exact sequence POST route. Counterfactuals dispatch on sequence target/method, current scenario and sequence-call position, never a fabricated carrier key. Byte-level carrier errors use production `family:http`+`raw` actions for the first body-bearing POSTs in this private cumulative manifest. No development adapter or diagnostic inventory is used. Literal bodies are bound manifest bytes; no nested body template is added.

External M1–M4: normative sequenceResponse wrapper remains unchanged apart from private body wrapping; ordered operation-index projection witnesses count/order, per-index name presence/value mirrors declarations, failed entries forbid outcome and successful entries forbid problem code, and no fifth result is allowed. Named identity collision uses409/never; alias-path collision uses409/after-state-change and a supplied, never-committed literal path. Missing Resource uses404/after-state-change (the inherited Read problem table). Fixture reset isolates those IDs. Exact whole-ID list stays canonical, and actual totals are recomputed rather than reusing historical28/101.

Native retained-failure discriminator: after clearing the alias and replaying original failed creator/dependent keys, a changed valid update under the old dependent key must conflict and preserve the known record; a fresh key must perform that update. Fresh creator key also proves the cleared alias path usable. This does not establish commit-before-binding timing or close the separate creator-rekey semantic identity obligation.

External M5/M6 and native carrier scope: retain5000ms inner deadlines and measure actual runs before any timeout change. Carrier refusal includes actual401,415,body-size413 and invalid-pointer400 in addition to static syntax and count32/33. Body-size uses request-too-large; count uses limit-exceeded. Harness capacity131072 is separately bound, only to send above the unchanged installed65536 cap. Native/root's conservative interpretation controls over external M6: rate/unavailable branches are not silently excluded or credited via unrelated rows. No artificial rate limiter, scheduler, expiry clock or fault seam.

External L1–L6: the corrected first32 of the rejected33 are the exact32 success case, avoiding identity collisions; every valid original key is distinct where required and retry semantics change the marker. Raw duplicate/surrogate spelling is authored literally. Name-on-update refusal is schema closure. Only test-support/read-update-coverage-v1.json changes, never development inventory. Same-row discovery pins `/limits/sequence/operations`32; body413 and count413 assert their distinct codes. Absent/malformed envelopes with no valid member keys can prove carrier refusal, but do not claim a key-retention discriminator for nonexistent keys.

W1b repair remains immutable; the working successor incorporates only approved cheap b6 disclosure/duplicate-route cleanups and explicitly defers the broad temporary-store lifecycle refactor and alias scenario trimming. Existing private cumulative source cap is4MiB; public/diagnostic fixture/catalog/source restrictions remain as frozen in c1. Final evidence must retain actual same-target configuration, raw observations, independent afterstate, explicit uncovered IDs and false eligibility.
