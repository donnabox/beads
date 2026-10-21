# Serving audit after minimum and the pending sequence correction

Read-only source audit at clean minimum candidate **2e1676caa0f5861d16c5741fbd412f6c3ef7c269**, tree **d909acfdbc65e111855233ff86a3db6edfedd49e**. This is PR141's candidate, **not merged main**. The sequence draft on df80 is only a future assumption: its two frozen production changes address retained-member depth ordering, but have not been integrated or runtime-qualified by this audit. The sole design's active goal/T2/T3/T4/P predicates remain authoritative (design.md:7–24).

## Conclusion and one useful next slice

**No additional small, settled serving repair was established in this bounded inspection.** This does not mean full Transactional functionality is complete. The source still has real private refusals and activation constraints described below; neither absence of a new defect nor successful standalone witnesses permits advertisement.

The smallest independently implementable next slice is **cumulative current-Resource minimum-read evidence**, using the already implemented behavior, not another implementation of it. The current Transactional manifest and coverage bind read-update.http.command-minimum-fields, whose coverage reason explicitly omits receipt/read minimum enforcement. Neither transactional.changefeed.consistency-fields nor transactional.changefeed.catch-up-timeout has a scenario or coverage row in these current artifacts. Their exact catalog obligations are respectively no older-position success and normative failure when the bounded wait cannot establish eligibility.

A coherent bounded witness can bind a real seed/current context, missing/hidden refusal before semantic minimum, and authorized existing Resource unknown-minimum catch-up-timeout, with complete body/schema/current-context assertions and a named discriminating control. Use the accepted PR141 scope and actual observations; do not extend it to retained handles or incident-Link lists. Known-ahead success/timeout and conditional ordering can share the same current-read context if included deliberately. Dynamic epoch/view values must come from actual serving observations through a supported harness seam, not invented literals or assumed raw-body capture. This report does not specify counts, release a plan, or assert that the existing manifest vocabulary alone can express every dynamic request. The next design must bind the exact actual input/output interfaces and preserve partial attribution.

This advances T4 toward qualifying the actual served configuration. It does not close all clauses of those catalog rows, retroactively promote standalone32+2+2 results, or solve cold discovery. Ordinary Update stays independent: PR138's cold directory and PR140's command-minimum evidence are already delivered; copying Transactional context or wait semantics into Update would be incorrect.

## Remaining actual serving/activation boundaries

1. **Cold Transactional entry remains deliberately unavailable.** transaction-http.ts:212,244–245,595–602 returns the private transactional-discovery refusal for Scope/discovery/directory navigation. transaction-service.ts:266–282 and transaction-discovery.ts:24–66 already assemble real same-entry context, descriptor and12-target directory. The descriptor explicitly says profile transactional. Wiring that builder to HTTP now would be an advertisement, not a neutral route repair; a non-attesting header cannot waive the sole plan's required behavior/evidence gate. No duplicate inactive builder, alternate endpoint or bypass flag is recommended.

2. **S1 and G2 remain actual unresolved boundaries.** Retained Snapshot controls still take the private minimum gap in transaction-http.ts:425–430, with owner/service retained-minimum restrictions; the original-versus-current retained context question is not settled here. transaction-events.ts:260–266 explicitly refuses metadata-only incident-Link updates projected into endpoint Bead feeds. That is G2, not a new question and not permission to choose membership. Existing ordinary Resource/initial missing-anchor rules in spec5833–5851 do not settle either issue.

3. **Fresh depth/composition refusal remains a full-delivery limitation even after sequence repair.** transaction-set.ts:52–63 is a private depth32 execution guard. The pending sequence change preserves its fresh-key refusal while allowing retained comparison; it does not implement arbitrary-depth fresh execution. Spec1048–1055 says omitted advertised limits do not permit a non-normative failure, and the command status/receipt rules remain operative. Thus the private500 cannot become final qualification by relabeling it. A safe capacity/serialization/failed-outcome slice would require tracing all execution and projection paths and choosing the already-specified correct phase; this narrow audit does not establish that simply deleting the guard or changing500 to413 is sound. I recommend the concrete existing-behavior evidence slice above rather than asserting an unreviewed depth redesign.

4. **Other unknown-token combinations remain explicitly outside PR141's ruling.** transaction-consistency.ts:131–147 permits unknown waiting only for waitable current Resources with matching actual epoch/view; unsupported routes/context combinations still throw the private gap. HTTP receipt/Event/changefeed paths retain their existing strict minimum handling. The operator's narrow current-Resource rule must not be silently expanded to arbitrary retained contexts, successful incident-Link lists or unknown-plus-foreign collisions. No new human question is raised by this report.

No broader ACL map/regrant feature, optional History, Types work, provenance redesign or erasure administration is invented as a prerequisite. The installed reference configuration still needs truthful applicability and required-contract evidence; public activation, packaging/API selection and exact-source admission remain root-owned. This is not a claim that independent progress is exhausted: the named cumulative slice is available without new semantics, while S1/G2 genuinely remain pending for their affected behavior.

## Evidence and limits

Inspected actual routing, same-entry consistency checks/waits, Event projection, discovery assembly, current-read capture/classification, current catalog/manifest/coverage and sole-plan constraints. Existing fulfilled prerequisites were not reopened. I did not run tests, imports, builds, servers, reviewer processes, source edits, Git mutation or outward publication. Only this report was written through the current session guard. Source pins below independently match the clean candidate's committed bytes.

- docs/specs/bdp.md: 67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8
- packages/server/src/transaction-http.ts: 90489b22858be99579f7ce8753bab0d6204d4caae215f7a293c1537d09477636
- packages/server/src/transaction-service.ts: c700d5719918b60c3239a336eef02c9dff6717faca698777aba01181ebdcd278
- packages/server/src/transaction-consistency.ts: ad51311d6098db32fd47849e0b78d3118162cdf1e80ee9ef4d46324bfa9f436f
- packages/server/src/transaction-read.ts: 7e2cc29b668bb80a3ba2948e90511518a98a7ff798b1dbfb8c6f51f0c0bd2a6f
- packages/server/src/transaction-discovery.ts: bb8055e2089f640a5d2071e389ac2011d08c038427265b50ab3d03c02047cfef
- packages/server/src/transaction-events.ts: 6c5ab4482f0b00110a71a96b16d44ca5fe713cb5f397ff78778cc6646f632a3a
- packages/server/src/transaction-set.ts: 03b1ca9427c6963f5b6ce712a7f70540501e081ed635c742015d2bc49cc9524a
- packages/conformance/catalog/transactional-v1.json: c9125f914129c71f01cefc5c54b32db80637d37b57e321571c96e81936c31eb4
- packages/conformance/test-support/transactional/manifest-v1.json: cd90592e76cace65b115dbfedf5f604c9c4011ed2e1a0fad954c1edb12aca6cd
- packages/conformance/test-support/transactional/coverage-v1.json: 0c991feb128a50625397e337464badcb17d22a0c31e1d57a6b66fa1032a92a51

Sole-plan snapshot SHA256: 47948cb6b5c10717eece9a25910ad49b5036e5042a294f9c5bbf1cb7e163e9cc.
Pending sequence binding SHA256: 987f61764ca09df676f77ab5b99a8886809f4ce7908342bc8c27fd462fddcc81; production transaction-batch.ts 95f3ce3957503cba240f18e911fe3d2c8027b4accbba8f1128ea47e3d2b0c362; transaction-http.ts 0a616e27da60c746be738821008e7dab6511e31325e9442418b4bbece9be40f8. These are future assumptions only; not substituted into the minimum candidate.
