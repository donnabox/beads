# Current Type and incident-Link minimum reads — bounded implementation plan

Supporting plan under the sole `bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md`. Planning only: no implementation, tests, builds, network, branch changes or outward publication performed. This removes an actual current-read refusal while discovery advertisement and public profile admission remain closed.

## Exact basis and prerequisite

Inspected clean `vickie-discovery-limits-integration` commit **9b95df5dd866dbaf84150404c4db86ac8054b724**, tree **d6749c7da26a5eeb2debce8336af91acc279249a**. Read `transaction-discovery-serving-plan-v2.md` and `transaction-compatible-missing-minimum-plan-v2.md` (SHA256 `08d9e0728bcc87c20f1d25f73293a21a01c94cf272bc11149121d7db759798f1`). The latter is a pending implementation prerequisite, not an implemented property of9b95.

Root must assign an ordinary integrated, frozen, reviewed missing-minimum successor before authoring. Reconcile its exact callback, ordinary-facet narrowing, installed helper label/counter and test changes first. Preserve its compatible404 behavior and explicit incompatible-missing gaps. Do not copy a proposed implementation into this slice or edit its active checkout. No dependency on completing discovery activation or the unrelated Event rulings.

## Settled contract and actual mechanism

Normative citations below are `docs/specs/bdp.md` on9b95; re-anchor by heading on the assigned successor.

| Surface / law | Existing source and precise increment |
|---|---|
| Canonical in-Scope Type descriptor GET/HEAD | `Types`1090–1136 describes the Type inventory and descriptor representation. `read-request.ts199–221` parses only a canonical `types/…` point, without view parameters. `authority-read.ts790–795` already selects the configured descriptor under the owner entry and returns its observation. Allow that existing Resource request through minimum admission. External Type hosts and Type properties are not part of this route. |
| Initial Type inventory GET/HEAD | `Types`1090–1111 and collection rules1510–1583 require exact id/name/describes summaries, canonical default ordering and ordinary page semantics. `read-request.ts145–164` already accepts limit/cursor and excludes selectors. `authority-read.ts809–814` already produces the configured authorized summaries. Admit only requests with no continuation. |
| Initial incident-Link view GET/HEAD | `Incident Link reads`1468–1504 defines inbound/outbound/both (default both), complete Link records, next and the in-Scope Bead anchor. `read-request.ts236–253` already distinguishes initial requests from cursor continuations. `authority-read.ts816–844` checks the anchor, reads actual incident rows, applies direction/visibility and deduplicates. Admit initial requests; preserve the same implementation and its filters. |
| Same-entry current minimum, conditional and HEAD | `Transactional HTTP consistency`5758–5787, `Transactional Resource read context`5793–5799 and `Conditional read projection`5801–5811 require an eligible projected prefix, context and eligibility before conditional shortcuts. Type data can remain unchanged while the Scope position advances: a constant descriptor is not permission to ignore a minimum. |
| Missing anchor versus current context | `Reads after deletion`1445–1450 explicitly includes a Bead's view=links. The pending missing repair checks an observed404's actual current context and returns it only when compatible/satisfied. Its narrow interpretation of a Problem representation applies here too; it does not choose precedence for missing+behind/foreign. Unknown Type uses the existing observed404 source path, not an invented Type lifecycle rule. |

At9b95, `isInitialTransactionCollection` (`transaction-read.ts26–32`) admits only initial beads/links. Both `minimumFor.perform`112–125 and `transaction-http.ts525–538` repeat the remaining bead/link point/properties predicate. They reject all three proposed current surfaces before the existing mechanism can check them.

The mechanism already suffices. `authority-read.ts977–998` materializes data and captures observation in one owner entry, then invokes `checkSourceCondition` before finishing/allocating a first page. `transaction-service.ts289–308` wraps each `minimumFor.perform` attempt in the existing service lifetime and finite wait; each retry obtains a fresh entry and no entry spans the wait. HTTP passes the Request signal and attaches the actual observation's current context. No service, owner, store, pagination or minimum algorithm change is planned.

## Small production delta and exact boundaries

Replace the two copied admission expressions with one source-private predicate in `transaction-read.ts`, used by its facet and the HTTP gate. Prefer a plainly named `isCurrentTransactionMinimumRead(request)` that recognizes the complete existing-plus-new subset:

- `kind:resource`, resource bead/link/type;
- `kind:properties`, resource bead/link only;
- `kind:collection`, collection beads/links/types, with continuation absent;
- `kind:bead-links`, with continuation absent.

The existing Read parser/authority validation remains authoritative for canonical IDs, controls, scope, direction, limits and request shape. This predicate is not another validator. Keep alias resolution explicitly rejected, as well as every collection/incident continuation. Do not admit Scope/descriptor/directory, external URLs, hypothetical Type properties, retained Snapshots, or malformed/partial minimum tuples. Preserve materialized invalid-parameter/limit-exceeded refusal ordering and existing HTTP early400 behavior. Do not make a wildcard `resource`/`collection` acceptance that silently admits future kinds.

The missing-repair callback has no request-kind argument. Once these routes are admitted it will correctly apply its common observed404 rule without a new callback or flag: compatible absent Type/Bead anchor returns existing404 plus actual context; missing+behind, missing+foreign epoch/view remain explicit `minimum-position-visibility` gaps with no context. Unknown position spelling remains the receiving gap. Integrity/context failures propagate as sanitized500 without the visibility marker. These are not hidden-policy witnesses: the installed policy is all-visible. No unresolved precedence is chosen. An initial Type inventory has no individual subject anchor; its empty authorized set, if configured, is a collection, not a missing Type. Continuation-anchor-versus-current minimum remains unresolved and is not admitted.

Ordinary `readFor` must retain the predecessor's signal-only forwarding; caller `sourceCondition` must not provide a back door around this subset. Existing aliases and non-minimum continuations keep their existing behavior. Do not change the Type descriptor body, fabricate a Resource revision or derive an ETag from Scope position.

## Six owned files, no other production edits

1. `packages/server/src/transaction-read.ts`: shared private predicate and facet use; reconcile header comment, preserve predecessor callback/narrowing.
2. `packages/server/src/transaction-http.ts`: import/use that predicate in the existing minimum gate only.
3. `packages/server/src/transaction-read.test.ts`: actual-owner/source and service-lifetime tests using existing fixtures/imports.
4. `packages/server/src/transaction-http.test.ts`: actual handler projection/refusal/conditional matrix and boundary tests.
5. `apps/bdptest/src/transactional-collection-minimum-installed.test.ts`: separate labelled installed scenario using the predecessor's helper; existing scenarios unchanged.
6. `docs/design/transaction-implementation-progress.md`: exact implemented subset, exclusions and source-bound evidence.

No changes to runtime, service, shared authority/parser/HTTP semantics, installed launcher/binding, configuration, schema/catalogs, Type lifecycle, public exports/admission, coverage flags or existing test budgets. If an actual defect requires another production file, stop and report it before broadening.

## Source and HTTP evidence

Use the existing real owner fixtures, controlled minimum wait hooks and handler. Do not test only the predicate or mock a successful projection.

1. For all three surfaces, eligible equal and earlier minima return the expected literal descriptor, exact ordered summary tuples, or complete incident Link tuples plus next. Bind returned current context to the same owner observation. Make a real subsequent Bead commit and prove Type descriptor/inventory bytes stay equal while their reported current position advances; actual behind checks still reject. For incident reads, independently assert endpoints, Type, properties/metadata and opaque revision relations, not just counts. Include inbound, outbound, both/default and a self-Link deduplication source case.
2. Present empty incident result still checks the minimum. Source inventory fixture may also use an empty installed set if the existing fixture/configuration permits it without new contracts; do not invent that fixture solely for a quota. A configured nonempty inventory is sufficient for the main Type witness.
3. Actual behind attempts reject before page token allocation for Type inventory and incident reads. Use limit1 with more than one item, existing token counter, repeated attempts and a subsequent genuine commit; then eligible first page can allocate. Follow its real continuation ordinarily after another commit and retain original rows/anchor versus new serving context. The same next URL with a minimum remains refused. This checks both preservation and boundary, not continuation-minimum support.
4. Service finite wait: for each surface a real independent mutation advances to the required position, and the successful response uses a fresh entry. Wrong epoch/view produce the existing normative context-bearing errors; future bound times out without conditional304/412. No owner lock survives a wait. Preserve exact retry classification/Retry-After and GET versus HEAD body projection. Use controlled hooks for source timing rather than sleep-based success.
5. Actual Request pre-abort must reject before any owner SQL for each new surface, extending the existing DatabaseSync spy pattern. Abort while a behind attempt is waiting must reach the existing service signal, end without a later success/page allocation, and leave the owner reusable; preserve close/drain behavior. Test using the existing wait seam and AbortController, no new runtime cancellation API. This does not promise interruption inside synchronous SQLite work.
6. HTTP matrix: GET/HEAD, successful If-None-Match:*304 and mismatching If-Match412 only after eligibility; unacceptable Accept behavior and malformed/repeated/unsupported controls retain shared semantics. Unauthorized requests remain401 without context/body disclosure. Source foreign minimum tests cover each surface and both foreign types. Preserve all three discovery refusals and alias/continuation/Type-properties exclusions.
7. Missing reconciliation tests: absent Type and absent/deleted incident anchor at compatible bound produce observed404/current context; behind/foreign retain the predecessor's gap without polling. Deleted anchor is proven by actual prior create/delete. No hidden-policy claim or new missing-order rule. Keep the predecessor's malformed-context (`pos-02` reaching callback) integrity discriminator.

Bounded independent sensitivity checks: ignoring the minimum only for Type/current incident reads must fail behind/wait tests; skipping incident direction filtering must fail exact tuple tests; replacing current context with an earlier captured context must fail the Type unchanged-body/advanced-position test. Use disposable source mutants only if needed for evidence; keep real red/green outputs separate from installed success. Local descriptor/summary/Link corruption controls must invoke the same oracle and fail for changed name/describes, omitted Link metadata or wrong endpoint, not merely invalid JSON. No generic harness feature or extra installed mutant required.

## Installed observation:20 requests,1 child,7 canonical URLs

Reuse the installed collection fixture's Task and Relationship contracts, actual guarded raw transport and label-aware lifecycle. Task owns outgoing Relationship Links. Seed using one real five-member batch: create Beads a,b,c and two opposite Links l-ab/l-ba between a and b. Literal authored IDs, types, properties/endpoints and per-operation receipt results independently establish state. c has no incident Links. The single effective commit must have effect/required/current position **pos-1**, despite five members and owned-source revision changes. Capture actual epoch/view and legitimate opaque Link revisions; do not expect a's create revision to equal its final owned-source revision. No new resource schema or Type lifecycle action.

|Requests|Observation|
|---|---|
|1|Real seed batch, exactly5 per-operation outcomes, literal IDs/endpoints/payloads, real receipt and independent pos-1.|
|2–7|GET and HEAD for canonical Task Type, initial Type inventory, a?view=links. GET descriptor equals authored fixture descriptor; inventory exactly Relationship/Task summaries in canonical URI order, next:null; incident page exactly both complete Links in canonical order, next:null. HEAD bodyless with corresponding successful metadata and exact pos-1.|
|8–10|Type GET If-None-Match:* →304; Type inventory HEAD mismatching If-Match →412; incident GET If-None-Match:* →304. Actual current context retained, no invented ETag.|
|11–13|Recognized future pos-2: Type conditional GET, Type inventory HEAD, incident conditional GET each503 with current pos-1, exact Retry-After1; GET canonical catch-up-timeout Problem/retry, HEAD no body. Type's constant bytes do not bypass eligibility.|
|14–16|Type foreign epoch GET410; Type inventory foreign view GET409; incident foreign epoch HEAD410. Exact existing error/context/body projection; broader combinations tested at source.|
|17|c?view=links at satisfied minimum: exactly empty items,next:null, pos-1.|
|18–19|a?view=links&direction=inbound then outbound: exactly l-ba then l-ab full records, next:null, pos-1.|
|20|Anonymous Type inventory GET with minimum401, no context or Type data.|

Seven URL identities: batch target, Task descriptor, types collection, a default Link view, c default Link view, a inbound and a outbound views. GET/HEAD remain separate exact route tuples. No discovery probes, hidden polling, reopening, schema fetches, extra GETs or DB seeding. Use the actual captured source/configuration binding; per-request route grants are exact and do not imply one identical executor-route digest. Preserve the declared credential resolver channel for Authorization-View and credential redaction.

The new scenario's local counter includes the POST, fails before exceeding20, and requires exactly20 on success; exactly1 actual child. Use a new explicit label while preserving default/old labels and reports. Original collection and pending missing scenarios, if rerun, are separately counted, not incorporated into20. Maximum live3 Beads+2 Links, one five-member batch; first pages have<=2 items under existing default2, so the wire ledger does not retain extra pagination state. Source tests exercise allocation boundaries.

Preserve existing **90s cooperative aggregate,120s outer,5s exchange,10s lifecycle**,32768 response-body,16384 header,65536 request-body and general80-response safety bounds. Three real1s future waits are included. These are inherited limits, not promised performance; record actual final timings and stop on budget failure rather than raising them. Existing helper signal/cleanup and genuine completedResponses reporting remain unchanged; failed/aborted runs must remain failed with actual observations only.

## Done-when and handoff

- Root accepts this plan and supplies the exact reviewed ordinary missing-repair base. Only the six files above differ; existing scenarios and all public/discovery refusal gates remain preserved.
- Predicate/parser, same-entry context, eligibility, wait, conditions, abort and missing-collision cases pass meaningful focused tests. Retained minima and aliases remain excluded. Fixed all-visible coverage is labelled accurately.
- Pinned Node24.16/pnpm11.20 clean prebuild typecheck → build → affected Read/HTTP/service/consistency tests, format/lint/boundaries and existing Read evidence verification pass. Avoid an automatic redundant full suite; run any further checks justified by actual changes/failures.
- Freeze a clean source. Retain actual protocol/server build/pack/offline install and run the new20/1 observation with installed entry plus exact source variables. Record source/tree/dirty, tarball/payload/config/fixture/policy/auth identities, request/response hashes, counts, commands/exits/timing, child exit and no remaining owned processes. Launcher-declared source is not target attestation. Old packed evidence remains pinned to its original source.
- Native and external council examine production and oracles; root decides acceptance/integration. This slice closes three current serving gaps but supplies no whole-profile/whole-row credit, discovery activation or public admission.

No new product ruling is needed for these present current reads or applying the accepted compatible404 callback. Missing anchor combined with behind/foreign context, retained continuation comparison and unsupported minimum receiving forms retain their explicit gaps. Stop for any attempted choice among those unresolved semantics, additional production files, fixture contract changes, budget increases or claimed hidden visibility evidence.
