# Next serving recommendation: initial finite Link Event minimum

Recommend the smallest remaining independently settled slice: canonical initial finite Link `view=events` GET/HEAD with a complete lexical unknown minimum and the actual current epoch/view should reuse the existing bounded wait and contextual503 `catch-up-timeout`. This is a concrete private500 removal, not an evidence-only exercise. It extends the accepted reference choice already delivered for Scope and Bead Events; it does not establish a new Event-specific protocol MUST.

This supporting audit is subordinate to the sole design.md workstream. It is not source-authoring, execution, public-admission or integration permission. Basis: clean published PR155 commit5bcd277b2a61461211ffc0c2d2f8fd701ef0a385, tree64fa36f4196da7e9eb48a9d248725d5f38c46282, soleparentf6204255d53ef65d4aea56ff5ccc86d72e8cc7b0. Parent reports all21 local stages and final native audit passed; exact-head CI39001388411 remains root-owned, with no new CI polling or integration here.

## Comparison and decision boundary

| Candidate | Concrete evidence | Recommendation |
| --- | --- | --- |
| Initial finite Link Event, matched-context opaque minimum | Current seeded Link GET/HEAD pair explicitly expects private500; all three selected-minimum seams exclude Link | Next bounded implementation-plan preparation; no S1/G2 ruling needed |
| Opaque minimum plus foreign epoch/view | checkCurrentReadMinimum rejects it as Unsupported before any known-token comparison; Scope/Bead foreign controls preserve private500 | Keep excluded. Choosing a foreign-context outcome is outside the accepted matched-context ruling; do not generalize known-token behavior or silently decide policy |
| Cold Scope/descriptor/operation discovery | transaction-http.ts603–611 deliberately refuses these routes; transaction-discovery.ts25–73 already assembles real internal descriptor/directory data | Genuine delivery blocker, but public activation remains explicitly closed by the sole plan pending complete same-target functionality and qualification. Do not expose an unsupported cumulative profile merely because an internal descriptor exists |
| Sequence-alias / ordinary Update | Existing workstream evidence supplies no new concrete defect in this bounded read | No evidence-expansion substitute for the demonstrated Link serving gap |

## Existing behavior and precise production seam

`packages/server/src/transaction-http.ts503–530` selects lexical current-read minimum handling only for Bead or canonical Scope Event sources, finite JSON, no parsed or explicit after, no normalized/raw Last-Event-ID. Link goes through strict transactionMinimumHeaders; literal `opaque` becomes the unsupported sentinel. For an existing visible Link, readEvents returns unsupported and HTTP produces bodyless private `minimum-position`500.

The discriminating witness is `packages/server/src/transaction-events.test.ts1816`: the Bead HTTP callback creates real Beads a/b/gone plus Link l in its seed at1723–1728, then calls the existing Link GET/HEAD pair with opaque and expects500. This is not a missing-resource fixture or hypothetical path. The same file1620 has a direct strict-capture TypeError for canonical links/l, but that owner fixture seeds only Bead a: it is an intentionally missing Link, not a second existing-Link500 witness. Widening capture must deliberately reclassify that old direct oracle to contextual missing, while a new seeded-Link owner oracle proves Unproved. Do not silently seed this historical fixture or describe its prior error as observed HTTP behavior.

Three production changes are sufficient in principle:

1. HTTP: add parsed Link kind alongside Bead in the existing initial finite selected-source predicate. Preserve every current finite/no-after/Last-Event-ID/media fence, transport/auth/lexical receiving order and existing parser. No URL grammar or new parser.
2. Service (`transaction-service.ts469–512`): permit canonical parsed Link alongside Bead inside the first unchanged-waiter attempt. Require exact parsed source.url equality and no after; preserve policy → initial clock → abort → classification/capture order, single capture and immutable retry tuple. Catch only existing URL/parser classification errors; keep snapshotCurrentReadMinimum outside that catch. Failed classification must reach the original owner error path.
3. Owner (`transaction-store.ts1639–1693`): widen initialMinimumSource to Link with after undefined. Preserve retained identity/authorization check before semantic comparison, sentinel behavior, one owner entry, current-context check before any group reader/projection/continuation allocation, and all subsequent Event projection/cursor code. Update only comments naming the selected source kinds.

Do not change transaction-consistency.ts, Event projection/membership, discovery, schemas, exports, mutation semantics or ordinary Update. The current three test files can hold the new regressions: transaction-events.test.ts, transaction-service.test.ts and apps/bdptest/src/transactional-event-missing-context-installed.test.ts. Proposed scope remains six files; a final reviewed plan must freeze exact callbacks/counts before authoring.

## Normative basis and exclusions

- bdp.md5841–5843 permits the remaining reads to serve an eligible representation, wait/route, or return normative consistency Problems; never stale success. The adjacent explicit unknown-minimum MUST at5844–5849 is restricted to current Resources/properties. This proposal is the existing allowed reference waiting choice extended to Link Events, not an inference that Resource wording universally governs Events. Timeout semantics are5329–5332.
- bdp.md5583–5585 gives Link Event Sources their own created/updated/deleted facts. The source is already supported by transaction-events.ts144–178, and projection at238–272 includes those Link-subject facts. The G2 guard at261–267 is specifically a Bead-source incident-Link metadata case. No membership choice is needed to gate an initial Link source before projection. Keep the G2 callback and refusal untouched; make no combined G2+opaque observation claim.
- bdp.md6258–6263 defines omitted after as earliest retained Event. Restricting this change to an initial read avoids S1 retained cursor/Snapshot precedence. Do not touch beginEvents/nextEvents, continuation or SSE; no Link incident-list view=links changes.
- bdp.md5665–5676 preserves authorized retained Event Sources after Resource deletion and distinguishes unknown/hidden sources. transaction-store.ts693–710 uses retained identities, not live Resource existence. Preserve deleted Link history, current deleted-Link Resource404, and hidden/unknown Event404 with actual current observation context. This is not a new universal precedence rule for all retained reads.
- Receiving lexical rules5819–5832, conditional precedence1986–1989/5876–5878 and finite validator omission6612–6616 remain unchanged. Exact HEAD bodylessness and corresponding GET content-length are distinct assertions.

Matched-context initial finite Link eligibility is not blocked by S1/G2. Unknown-plus-foreign, after/Last-Event-ID, SSE/incompatible media, retention-expiry policy, Type/History/provenance work and public profile activation remain outside scope. If implementation requires changing these boundaries, stop for root review; no fresh operator question is proposed here.

## Smallest useful regression families

Reuse proven fixtures and lifecycle helpers without rewriting historical evidence. A final plan must enumerate actual operation/request/owner/probe counts, including setup, before any source release.

- Owner: seed two real Beads plus Link in one real batch before installing the scan spy; read Link Event setup for actual context/cursor. Prove matched opaque Unproved and recognized behind before retained-group SQL; current known/no-minimum yields the correct Link-subject page as sensitivity controls. Delete Link and prove retained created/deleted history plus opaque Unproved; replace authorization policy and prove hidden missing/current view with no group scan. Check unknown Link missing, foreign still unsupported, after still strict. Count reads separately from seed/delete/policy calls. Capture the actual prepare implementation once; do not restore/re-spy around deletion or use owner.inspect as a no-scan probe.
- Service: use a real seeded Link owner with delegating readEvents counters. Establish canonical classification, tuple capture against caller mutation, original fixed deadline, owner.withRead probe availability during waits, cancellation and drain. Preserve exact preflight error ordering and classify malformed/noncanonical candidates through the original path. Existing common preflight controls may be reused; add only Link-specific discrimination necessary to prove the new branch rather than duplicate every historical case.
- Emitted HTTP: seed real Link endpoints and Link, obtain its actual Event context/cursor, then opaque GET/HEAD yields contextual503/after-delay with no private marker/validator and no conditional304/412 substitution. Include known/no-minimum positives, malformed/parser refusals, unknown/hidden or retained-deleted identity discrimination, plus explicitly unchanged foreign/after/SSE/Last-Event-ID/media fences. Keep the5s callback bound; if exact selected workload approaches it, stop and review a split, not a timeout increase.
- Supersede only the direct missing-Link TypeError at1620 and seeded Link500 pair at1816 after bounded inventory confirms no further applicable old oracle. The historical Bead HTTP callback keeps its37 requests: its Link pair becomes503 and adds two80ms waits, so timing must be measured. Its prior PR155 evidence remains attributed to5bcd; it is not relabeled as successor proof.
- Installed: append one independent four-request witness: real batch POST containing two Beads+Link, canonical Link Event setup GET, then opaque GET and HEAD using returned context. Exactly four HTTP requests means three seed operations inside one actual request, not a prepopulated store or synthetic discovery. Reuse actual child/package/credential provenance and allowlisted exact routes; preserve5s request/45s owned/60s callback bounds, actual PID/exit/directory observations and removal only after verified clean close. Historical installed callbacks remain byte-preserved.

Validation sequencing must retain fresh build before any source HTTP dynamic dist import or installed test module collection; selected source checks with immediate observation audit, packaging/consumer binding, new and existing installed callbacks with immediate audits, and final unchanged-source/cleanup review. Exact current production bytes must map to any later clean ordinary commit. No commit/publication until required validation and actual native+Claude code disposition; root owns all release and exact-head CI/integration decisions. This audit performs none of those steps.

## Public-support effect

An authenticated client already reaching the private Transactional reference listener gains a real normative timeout response for one previously unsupported Link Event request class. This closes a required serving prerequisite, not usable public Transactional delivery. Scope/bdp.json/operations discovery, cumulative profile advertisement/admission and release/Preview2 remain closed; full-profile claims remain unjustified. The recommendation does not fabricate cold navigation or count local diagnostic assertions as completion percentages.

## Read checkpoint

No tests, imports, builds, service probes, review CLI, source edits, branches, main fetches, commits, pushes or coordination posts were performed. Only this supporting audit is written through the session guard. The following hashes identify inspected source and supporting plan bytes; the live sole-plan hash is a read checkpoint, not a future mutable review dependency.

- `5bcd277b2a61461211ffc0c2d2f8fd701ef0a385:docs/specs/bdp.md` — `67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8`
- `5bcd277b2a61461211ffc0c2d2f8fd701ef0a385:packages/server/src/transaction-http.ts` — `37b8ed7a63c25d310159edba83d7ecc6f0292724831c2b59643e7a3788835589`
- `5bcd277b2a61461211ffc0c2d2f8fd701ef0a385:packages/server/src/transaction-service.ts` — `cede2090af8b4cb4470334ff27e6f5038ab4cb186972ed2bd03a300ae19a6ba6`
- `5bcd277b2a61461211ffc0c2d2f8fd701ef0a385:packages/server/src/transaction-store.ts` — `3e7b9b4609602ff5c207e697bdf0d70a5c4f6e94515130b926e83eb6a8c093b7`
- `5bcd277b2a61461211ffc0c2d2f8fd701ef0a385:packages/server/src/transaction-consistency.ts` — `ad51311d6098db32fd47849e0b78d3118162cdf1e80ee9ef4d46324bfa9f436f`
- `5bcd277b2a61461211ffc0c2d2f8fd701ef0a385:packages/server/src/transaction-events.ts` — `6c5ab4482f0b00110a71a96b16d44ca5fe713cb5f397ff78778cc6646f632a3a`
- `5bcd277b2a61461211ffc0c2d2f8fd701ef0a385:packages/server/src/transaction-events.test.ts` — `c627c3103b49af98df4c9d32e40456989da5deb7d5a3eec4ade539de7b24e44a`
- `5bcd277b2a61461211ffc0c2d2f8fd701ef0a385:packages/server/src/transaction-service.test.ts` — `45f2444071c5cbbd3651449f878feb1e70d9e51c3c57cdcc243395b7735b25a7`
- `5bcd277b2a61461211ffc0c2d2f8fd701ef0a385:packages/server/src/transaction-discovery.ts` — `bb8055e2089f640a5d2071e389ac2011d08c038427265b50ab3d03c02047cfef`
- `5bcd277b2a61461211ffc0c2d2f8fd701ef0a385:apps/bdptest/src/transactional-event-missing-context-installed.test.ts` — `e4510bf2dfdaacac981bf9583a2f6be4c536de5ceca5309c754f93fd56b6f652`
- `/Users/dbox/repos/wt/vickie/bdp-integration-logs/finite-event-next-serving-audit/plan-v3.md` — `57bf427256d4531ee255d7b1e99663158f10057512bad27bf5bf472f7e7e9761`
- `/Users/dbox/repos/wt/vickie/bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md` — `77678091f77a551e1377ad22aff7785ecdb9c69ba16f163a2bfe0a4c9f19db04`
