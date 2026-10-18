# Next TX evidence: one durable receipt continuation

Read-only supporting proposal from frozen00fa81281291e83bac16f917250a7f5615c744ae, tree c9a4795509134bb5cc44bddf95d5f1be0b722f85, `/Users/dbox/repos/wt/vickie/vickie-set-effects-evidence`. Sole plan of record remains the fork design document. No tests, builds, services, source edits, network or outward actions were performed for this plan. Existing author evidence was inspected only. This proposal requires council and root scope/base assignment before implementation. The frozen source and prior report remain unchanged.

## Recommendation and actual blocker

Observe one new partial canonical row, `transactional.receipt.pagination`, using32 ordinary Task Beads and a singleton updateWhere that yields33 result entries: one matched32 plus32 updated postimages. Inline results must be32 entries (matched plus p00…p30); one observed continuation must contain only p31 and next:null. Use existing Task schema/configuration, no Links/owned Types. Later mutate p31 again to distinguish immutable receipt history from current afterstate; repeat the same receipt/page reads across one actual reopen.

Runtime support exists; the blocker is the deliberately inline-only private conformance helper. Do not start authoring under the prior four-file evidence-only scope. The smallest useful change is a bounded root-plus-one-continuation extension of the existing private `transactional.receipt-read` path, not a generic URL follower, automatic page walker, new public action grammar or new runtime behavior. If council declines that extension, stop at this seam; neither fabricated URLs nor a standalone test relabeled as a canonical pass is a substitute.

## Exact catalog clauses and boundaries

`packages/conformance/catalog/transactional-v1.json:420–434`, id `transactional.receipt.pagination`, has two selectedText requirements, both anchored `docs/specs/bdp.md#mutation-receipt-responses`:

1. “Otherwise the response contains the first page and an absolute `next` URL for another immutable page of the same receipt.”
2. “`page.maximumItems`, when advertised, bounds the entries in the receipt's inline `results` and in each page, every entry counting as one whatever its outcome”.

Spec5068–5073 describes atomic recorded results/no recomputation.5167–5176 defines inline prefix, absolute next URL, page `{receipt,results,next}`, entry counting/no entry splitting/projection.5179–5187 preserves issued URLs and recorded boundaries through same-epoch ordinary restart;5189–5193 requires absent validators and private,no-store. Schema `schemas/bdp-v0.schema.json:4094` requires page.receipt/results/next, nonempty results, absolute parent and nullable absolute next.

Existing readable row (`catalog:407`, spec#mutation-receipts) remains preserved; its selectedText covers independently readable receipts for lost responses, pending duplicates or paginated results. A paging witness supplements it but does not replace its object or grant broader credit. Do not add a separately credited page-restart row: its selectedText at1973 also requires current authorization/erasure projection and preserved expiry/forgetting/retraction/prior-epoch nondisclosure, none proved by this fixed-view normal-restart case. Page-expiry438 and validator-omission2017 likewise remain uncredited. Presence/cache/schema checks needed to validate this experiment do not automatically earn those rows.

The current private fixture has a32-entry bound in actual normalized service configuration. Public schema-valid Transactional discovery remains unadvertised; do not present a private configured bound as proof of the advertised-limit branch of requirement2. This row stays partial pending selected-clause audit. Denominator250required+5optional remains; proposed9 planned partial rows,241unplanned required/246actual missing-plan results, no whole flags.

## Known runtime path and inspected context

- `transaction-http.ts:497–522`: receipt namespace routes `/pages/` to service.receiptPage, preserving auth/minimum paths and current-context decoration.
- `transaction-service.ts:320–363`: same-owner page reads, maximum/nonempty/next boundary integrity and current context. Existing service constructor locks advertised Read page maximum to persisted receipt maximum; no new setting is needed.
- `transaction-store.ts:1937–2021`: principal-bound durable receipt_pages plan, parent/epoch checks, recorded token/boundary selection, minimum before retention, immutable selected entries and recorded successor URL. This is the actual path to exercise, never inspect() or another owner.
- `transaction-runtime-configuration.ts:338`: normalized Read page maximum comes from receiptMaximumItems. Actual owner settings must be recorded from installed metadata, not invented from the number in the test.
- `apps/bdptest/src/transactional-set-installed.test.ts:239–268`: existing custom installed oracle observes33 results as32+1 by deleting32 Links, validates page schema and ordered identities. That fixture has different descriptors/owned operations and a custom sender. It is feasibility context only; do not borrow its pass, payload identity, configuration or response bodies into a new canonical row.

## Why the current private grammar cannot execute the row

1. `transactional-observation.ts:180–181` explicitly refuses completed source receipts with next!=null (`source receipt exceeds inline fixture`).
2. receiptUrl:59–81 accepts only a canonical root checkpoint token after receipts/, rejecting `/pages/` paths.
3. register:132–205 grants authority only to a canonical mutation POST capture at `/id`; `/next` is treated as an ordinary capture and revokes prior authority.
4. `runner.ts:884–982` resolves a receipt binding, builds the exact one-request route from source.receipt, forces mutationReceipt schema/body.id agreement and next:null, then returns a closed client projection.
5. Dynamic client receipt-read deliberately has runRequest.captures=[], so its later programmatic captures cannot secretly grant new HTTP authority. Ordinary exact routes are predeclared before execution. Merely capturing a page URL or constructing one from a key cannot work safely.

Thus changing only manifest/coverage/tests is insufficient. This is a small explicit private observation seam requiring root approval, not a new protocol question.

## Proposed minimal private seam (bounded to one continuation)

Reuse `transactional.receipt-read` input `{binding,method:GET|HEAD}` unchanged. Extend the existing observation ledger with a discriminated root versus first-page entry. Register root `/id` and first continuation `/next` from the **same already validated actual canonical mutation POST**; independent capture order must not alter authority. Retain root receipt identity, actual original action/session/principal credential and raw response digest. A page target is authorized only by that actual body's absolute next, after source mutationReceipt schema/terminal/detail and wire checks. Do not accept an ordinary Resource /next, a fixture string, transformed binding, inferred token, foreign origin/scope, query/hash, nested unobserved page or malformed path.

Preserve the existing root provenance shape for old rows. A page provenance variant explicitly says pointer `/next`, parent receipt ID and observed page URL; do not call the page URL a receipt ID or claim it came from `/id`. A readonly resolved target URL can drive the existing exact transport while the root remains available for page.receipt validation. New metadata should stay inside the existing private source-observation field, with type/discrimination tests; no new public report field is needed. Existing v3 serializer/legacy exclusions must continue rejecting the private field regardless of variant.

Allow the source/root representation's next only when it is the validated same-parent first page URL. Keep detail:available and original retained root JSON equality. For page GET, use mutationReceiptPage schema, require page.receipt===original root, nonempty results and next:null for this bounded two-part fixture. A further continuation is an explicit private unsupported boundary, not a normative invalidity ruling. The initial page body is **not present in the POST**: validate its own actual raw response, schema and parent before retaining its JSON as the comparison baseline. Subsequent GETs and reopen GETs compare with that first observed page body, not a fabricated expected body or the root body. Manifest independently asserts literal p31/content/order, so a self-consistent wrong initial page still fails its oracle. HEAD has no body and cannot establish a first page-body baseline; it remains real wire observation with current context.

Use the same validated active session, abort signal, pending registration, response limits and exact recorder for every request. Preserve original grant/session provenance while current serving generations advance. Register no wildcard routes; the dynamic page request gets exactly its observed URL. Ordinary capture reuse revokes ledger authority as today. Ledger entries remain bounded by authored action/capture limits and close with the scenario. No page chain registry, recursive walker, background pump or new credential mechanism.

## Concrete scenario and action ledger

One fresh scenario,32 ordinary Beads only. Literal IDs p00…p31, existing Task Type, stable metadata, original title `before`. Seed32 creates in reverse canonical order via one batch. Maximum32 ordinary operation slots/32Resource mutations/32Events fit existing configured limits; selection32 produces33receipt entries without exceeding the32matched cap. No Link or incidental closure. Seed batch itself has exactly32results and next:null, a useful every-entry boundary contrast.

All expected IDs are explicitly zero-padded literals in the artifact, never generated from runtime results. Source-generation tooling may write literals, but the installed expected tuples must not be derived from observed order.

| # | Action and required assertions |
|---:|---|
|1|Seed POST batch32; schema, completed, exactly32 created entries in submitted reverse order, literal ID/type/properties/metadata and next:null; capture actual root ID.|
|2|Seed root GET through existing helper; full retained facts/schema/context; no generated URL.|
|3|GET beads/?limit=32; schema/next:null and independent32-ID before-state inventory.|
|4|updateWhere POST selecting Task Type; replace/title→changed, capture actual root `/id` and first page `/next`. Root schema/completed, exactly32inline entries, matched32 first then updated p00…p30, each index0/full Type/properties/metadata, effect/required position present. Absolute next must belong to this root; don't infer its token.|
|5|Root GET; schema and original-prefix/body equality; capture complete root body and original transaction/checkpoint facts.|
|6|First page GET using only captured next authority; page schema, parent equality, exactly one updated p31/index0/full expected postimage, next:null. Capture actual page body and p31 record after validation.|
|7|Same page HEAD; actual200/body absent/headers and private,no-store/no validators.|
|8|Root HEAD; corresponding no-body/cache/context checks.|
|9|GET beads/?limit=32; all32 literal changed states; complete p31 record agrees with observed page record.|
|10|Fresh-key singleton updateWhere selecting only p31; title→later, own completed inline receipt with matched1/updated1; distinguish current state from old receipt detail.|
|11|GET collection; exactly p00…p30 changed and p31 later, stable metadata.|
|12|Original root GET again; original identity/prefix/body unchanged, current response context is captured independently.|
|13|Original page GET again; exact first observed page JSON still contains changed, not later; no recomputation.|
|14|Actual reopen of same directory/create:false through existing lifecycle.|
|15|Same root URL GET, retained body and next unchanged; current generation advanced, original provenance retained.|
|16|Same already issued page URL GET, exact original page/boundary unchanged, parent equality/current context/provenance and generation checked.|
|17|Same page HEAD after reopen, no body/context/cache checks.|
|18|Final collection exact later state; no mutation replay or reseeding.|

Require all three context fields on actual root/page responses and their client projections. Equality checks must have presence on both sides; do not compare current serving context to historical required/effect positions as if they were the same concept. Check immutable root/page JSON while allowing current response headers to change. Do not hard-code opaque revisions/positions/tokens. The successful late mutation guarantees state/position advances without changing old recorded page content.

Proposed caps18required/new<=20 (extra slots require specific seam reason), total<=141 versus current121;32liveResources maximum;3mutationPOSTs (seed/update32/update1); exactly one reopen. Current13sessions+2=15scenario/reopen sessions plus bootstrap16children. Existing nine retained root reopen IDs stay exact; add one root and one page named reopen witness, discriminated so assertions do not blindly expect page target===root ID. One page chain with32+1 entries only; no second page, expiry/failover/policy change/stress claimed.

No new fixture binding or ordinary exact route is necessary: existing batch/update-where and beads/?limit=32 routes suffice. Exact ten-route configuration inventory/digest should stay unchanged. Fixture716bytes/3a7899ce remains byte-identical. Both complete old eight scenario and coverage raw objects stay pinned.

## Files and authority required before work

Evidence files: TXmanifest/coverage, transactional-cumulative.test.ts, app transactional-matrix.test.ts. Small helper extension: transactional-observation.ts/test and runner.ts plus targeted runner tests for selected schema/routing/provenance. Eight likely files; no fixture/catalog/runtime/schema/Type/installed configuration changes. Existing report.test.ts is a ninth file only if current field-level exclusion tests cannot prove the page variant; inspect/reuse existing exclusions first and request scope if needed. No testing export or new helper module/deep-import allowance is anticipated.

The app current positive asserts all dynamic source references and nine reopened roots. Update it explicitly for root/page targets and the two new reopen witnesses; preserve the existing actual denied-credential/root controls. Do not weaken a field to unknown/any to fit pages. The ordinary public runner must still refuse private actions before preparation. Existing module/serialization/claim closure tests remain mandatory.

## Controls, budgets and done predicates

Five installed expectation controls, each against real successful wire and one uniquely named failure: wrong final inline member p31 instead of p30; wrong continuation member p30 instead of p31; wrong page parent expectation; wrong recorded p31 title `later` after state advances; wrong terminal next expectation. Exact one-entry array length/order and first-page prefix assertions independently prevent omission/duplication; count matched as one inline entry. Do not pretend valid receipt/page cross-wiring is unauthorized when it is instead the wrong expected result.

Helper/runner security tests additionally reject unobserved/foreign/query/hash/nested page targets, ordinary capture reuse, parent mismatch, schema-invalid page, first-page nonterminal unsupported shape, changed retained body, source-wire mismatch, missing validator and unvalidated physical session transition. Use existing seams and classify target-observation failures separately from harness defects. No fake source-wire injection in the actual installed positive. Verify aggregate abort/close still drains dynamic page traffic with unchanged bounds.

Planning fences proposed for council: max16KiB raw seed request (32small creates; comfortably below existing64KiB), max64KiB finite response per root/page/collection, unchanged1MiB transport observation and2MiB wire bound. New manifest<=80KiB growth and total<=320KiB (238693+81920=320613<327680); <=10000JSONvalue nodes versus current6699; coverage<=16KiB growth. These are proposed authoring fences, not runtime changes or automatic authority. Literal arrays of32IDs/records must be measured; stop/report if they exceed them rather than minify/drop assertions. No Event traversal is needed for this receipt row.

Keep90s cohort/120s outer/5s wire/10s preparation+cleanup; actual positive<=63s. Existing00fa eight-row3.118s result is historical context, not a promise for this extension. Build before tests; frozen source before final packing; actual new-source installed nine-row cohort and appropriate genuine RU aggregate. Record source/payload/contracts/config/fixture/route/report identities, max bytes, all action counts, parent/link wire digests, first page baseline, later/current sessions, exact budgets and all children closed. Preserve failed attempts. No borrowing custom set oracle observations.

Done only after exact old8objects/fixture preservation, canonical9-row/250required accounting,18-action installed positive, five discriminating controls, root/page shared lifecycle/security/schema tests, actual reuse across later mutation and restart, final static/Read seal/packed cross-profile evidence and native+Claude council. All whole flags remain false pending clause audit. No protocol question blocks the ordinary fixed-view two-part runtime behavior; the **private harness authority/type extension is an explicit scope decision** awaiting root/council, not implied by current grammar. Authoring remains paused until that decision and exact base assignment.
