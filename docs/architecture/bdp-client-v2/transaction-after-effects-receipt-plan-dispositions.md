# Receipt continuation plan review dispositions — native and external

Original plan SHA2560c53cdccde7b232f778d05d8f216637996bfbb96ea68798740f15e4ddc02c9a2 is unchanged. Native-reviewed plan4bb05e656f0214f8137e59406955783040339d1707fae5e03db0ea938a198118 and dispositions02c92b31d8260be8d222d5460c25b79945d823e8cf19c448752da0239f8ef433 are preserved in separate native-reviewed snapshots. The current revised plan incorporates all ten external findings under root's decisions. No implementation authorization follows from this revision; root/native closure and exact base assignment remain pending. No tests, services, source changes, network or outward writes were performed.

## Preserved native dispositions

- Numeric32 comes from the exact no-override source chain, not a nonexistent live numeric metadata field. Record actual installed payload/configuration identities and observe32+1. No runtime metadata extension; advertised-limit branch stays partial.
-18required new actions /139actual total;20new /141ceiling requires explicit justification for either spare action.15sessions plus bootstrap16children.
- All12 after-reopen dynamic exchanges (old9rootGET + newrootGET/pageGET/pageHEAD) require advanced serving generation, retained original source/configuration and current context. HEAD cannot initialize/replace page JSON baseline.
- Proposed replacement fences are anchored to00fa238693bytes/6699nodes and coverage41812, not the older baseline. Original eight scenario/coverage objects and716-byte fixture remain unchanged.
- Existing private-field report exclusion is variant-independent. Eight-file proposal remains; no extra report/runtime/configuration file authorized.

## External findings

1. **Medium — accepted, canonical insertion.** Explicit position8of9 in manifest AND coverage, between receipt.readable and idempotency.failed-retained. Old object content stays byte-pinned. Appending is not allowed.
2. **Medium — already corrected, retained explicitly.** Exactly12 post-reopen dynamic exchanges, including the first post-reopen HEAD. All receive the generation/provenance/context witness;11 refers only to GET count.
3. **Medium — accepted, entry-counting partiality.** New coverage reason must state only matched+updated outcomes are observed; no created/deleted/withheld/other boundary outcomes. Matched consumes a slot and is precisely why32updates spill into33entries. Advertised-limit branch remains separately unproved.
4. **Medium — accepted with root's narrower response cap.** New row alone gets16KiB raw seed /32KiB decoded responses. Preserve existing set mutation/zero-match2KiBPOST/32KiBresponse guards unchanged, with a distinct pagination filter. Remove64KiB response proposal; measure and stop if exceeded, never silently raise it.
5. **Low — accepted, both root-next sites named.** Source registration transactional-observation.ts:180–181 AND runner.ts:957 retained root GET must accept only null or a validated same-parent first page; preserve existing inline behavior.
6. **Low — accepted, positive URL whitelist.** Exact `${source.receipt}/pages/${token}` with canonical validated root and token `/^[A-Za-z0-9_-]{1,256}$/`;<=8192characters, parsed href equality, exact prefix, no credentials/query/fragment. Empty/oversize/illegal tokens, wrong parent, normalization and URL-component negatives are explicit.
7. **Low — accepted, frozen citations corrected.** receiptUrl56–77;register129–201;runner876–978;service320–360;store1937–2019. Source references refer to frozen00fa, not moving branches.
8. **Low — accepted, concrete oracles and terminology.** json-pointer `/headers/cache-control` equals `private, no-store`; `/headers/etag` and `/headers/last-modified` use `exists:false`. Applied to actual page GET/root+page HEAD observations with bodyless HEAD pins. Missing SchemaValidator is a separate harness defect, not an HTTP-validator absence.
9. **Low — concern accepted, proposed API change rejected.** Existing per-capture register accepts `/id` OR `/next` from canonical mutation POST, independently validates the complete SAME actual body each invocation and derives root/page. Each grants ONLY its requested distinct binding. No implicit second binding inside `/id`, no need for that call to know the page binding name, no API expansion or capture-order dependence. Test both orders, duplicate binding refusal, ordinary reuse/revocation and source/target mismatch. This meets the invariant with the existing capture seam.
10. **Low — accepted, controlling bound and headroom explicit.** transaction-batch.ts:62–73 fixes operation32/matched32/request64KiB;:865 uses strict greater-than refusal. Selection32 is legal with zero match headroom. The source-defined receipt32 default chain remains documented. A changed bound invalidates the experiment; no overrides or weakened oracle.

The external review is complete, replacing the earlier pending marker. Root/native closure is the remaining plan gate; source authoring has not begun.
