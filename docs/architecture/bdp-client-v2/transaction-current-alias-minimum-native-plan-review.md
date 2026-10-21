# Native review — current alias minimum plan

**No actionable findings.** The proposed fixed-policy alias minimum slice is implementable through the three existing forwarding/gate points. No unavoidable product ruling blocks its selected scope. Hidden-target authorization and incompatible missing/minimum arbitration remain explicit limitations, not decisions made by this forwarding change.

## Exact basis

Reviewed plan SHA256 `3ee5fc951c2278488f5153692956416b1b94e54bcb5c1832a144c2f2cc69d7dc` reproduced. Inspected frozen integration9b95 and compatible-missing successor `7791aed7394e30fef2d2c1ed4fed3f43504e42ba`; the latter's callback and ordinary-facet narrowing are the required predecessor semantics, not an assumed waiver of its review. No source edits, builds, tests or network activity were performed.

## Law and fixed-policy scope

- `bdp.md:1424–1439` supplies temporary307, absolute canonical Location, no representation/body, uniform unknown/query/fragment404 and authorization projection. The generic authority's existing target existence/visibility check (`authority-read.ts:769–787`) is narrower than treating a redirect as an unconditional locator disclosure. The plan correctly retains that implementation and confines evidence to all-visible policy; it neither removes the check nor claims to resolve hidden-target semantics.
- `bdp.md:4189–4213` explicitly makes alias state authority-local: singleton mutation receipts protect effects, but aliases have no Resource revision, position, Event, change group or snapshot. requiredPosition is a serving-view barrier, not alias replication proof; replicas resolve through the authority and cannot resolve offline. Therefore same-position repointing may legitimately change Location, and future-position catch-up must include a real Resource effect. The plan's seed/repoint/update design follows this law.
- Applying current same-entry view/position eligibility does not turn an alias into a historical Resource. Scope context remains the projected prefix accompanying current locator state. The plan does not infer a mapping pinned by the minimum or an atomic create-plus-alias batch; aliases remain singleton/sequence operations.

## Actual forwarding, lifecycle and HTTP behavior

`minimumFor.resolveAlias` currently rejects at transaction-read.ts146 (7791); the underlying facet already accepts delivery options, captures/materializes the alias and invokes sourceCondition inside the same guarded entry (`authority-read.ts:977–998,1127–1138`). Overriding supplied sourceCondition with the captured minimum and preserving checked-result handling is sufficient. The ordinary signal-only facet must remain unchanged.

`transaction-service.ts:289–303` already snapshots the minimum and wraps perform in run/waitForTransactionMinimum. The proposed resolveAlias wrapper reuses this lifecycle, finite deadline and signal semantics. Each retry must call the underlying alias resolution anew, not retain a first-attempt target. Alias repoint plus a real independent Resource write during a paused wait is a meaningful discriminator and also proves the owner is not held across await. Existing close drains the registered operation; no separate lifetime is needed.

The HTTP change must bypass ReadRequest classification only for a well-formed classified alias, retaining the nonalias subset gate. Its earlier method/authentication/classifier/tuple precedence can stay intact. The alias callback at transaction-http.ts196–212 already projects307 or Problem and attaches actual observation context; selecting the minimum facet and forwarding Request.signal supplies the missing mechanism. Final shared semantics are applied once at:596–600. Invalid query/fragment classification at read-request.ts935–950 remains before the later minimum gate, as claimed.

`read-http.ts:10–48` applies conditional/Accept transformations only to JSON200 or204. A307 therefore remains307 with Location, no ETag and zero body even with the specified conditional or Accept fields; HEAD needs no synthetic representation. Foreign/timeout exceptions use the existing context-bearing mapping in transaction-http.ts559–574, while the inherited missing collision uses the explicit visibility marker. The plan correctly requires no Location on any refusal and no wait for missing+behind. Ordinary alias signal forwarding is an intentional bounded improvement, not falsely described as unchanged cancellation behavior.

## Verified receipt shapes and installed feasibility

The plan's alias receipt assumptions match the actual implementation; they must not be replaced with Resource-result assumptions during authoring. `transaction-batch.ts:972–985` invokes the alias evaluator, appends `{operationIndex,...result.outcome}`, and creates no Resource change record. Existing executable source oracle `transaction-store.test.ts:1271–1302` validates the full mutationReceipt schema and exact available results:

- First put: `[{ operationIndex: 0, outcome: "created", alias: "<scope>alias/work", target: "<scope>beads/a" }]`.
- Repoint: same one-member shape with `outcome:"updated"` and target b.
- Delete: `[{ operationIndex: 0, outcome: "deleted", alias: "<scope>alias/work" }]`, without target.

Each has completed status, requiredPosition pos-1 and no effectPosition; no Resource/deleted Resource/operationName member. The schema's receiptResult alias branch and `bdp.md:4193–4198` agree. The canonical alias URL and target appear in **receipt results**, not as a top-level alias result. Use JSON mutation requests with `alias:"alias/work"`, not bare `work`, and fresh keys. This is a confirmed implementation detail, not a new requirement or additional API.

The installed helper in transactional-collection-minimum-installed.test.ts supplies exact raw Method/URL routes, no auto-follow, credentials including reader/anonymous and the existing authority-view resolver channel. It can call all three mutation targets through exchange; the local18 counter must include these writes, as the plan requires. Its fixture owns outgoing Relationship Links, but this ledger creates no Link, so no owned-Link source revision changes arise. Bead b's independently fetched Resource can retain its creation revision across alias puts.

The ledger sums to18 requests/one child and exactly seven URLs: three operation URLs, work/never/query alias URLs and canonical Bead b. No receipt-page request, extra schema fetch, hidden follow or restart is implied. Row9 is the explicit ordinary follow. The single future timeout fits the existing5s exchange under the default finite1s policy without asserting a hard latency guarantee. The existing90s aggregate/120s outer/10s lifecycle and wire caps need not change. Row18 legitimately tests the classifier's context-free404 at a future minimum and must not be mistaken for a malformed wire-fragment test.

## Test/evidence boundaries and ownership

The seven-file inventory is sufficient. Although service-lifetime tests are assigned to transaction-read.test.ts rather than service.test.ts, that file already imports createTransactionService and can reuse a real owner with page limits adjusted consistently, as its existing service tests do. No additional production/store/helper seam is required. The paused catch-up, alias-only nonprogress, pre-abort, close/drain, malformed-context and no-Location controls test distinct obligations; local Location/context corruption controls add no installed calls.

Freeze only after preserving the accepted7791 missing callback, ordinary sourceCondition-ignore behavior, existing C3 service changes and old installed scenario bodies. Any Type/incident route extension arriving first must be preserved through the exact root-assigned ordinary base. No shared concurrent writer or implicit source rebasing is justified by this review.

Actual source tests plus fresh exact-source packed18/1 proof are appropriate here because these are real served alias routes, unlike the separate private discovery accessor. The new reports must keep source/configuration/fixture/payload identities, failed-run counts and cleanup distinct from the old collection/missing/alias-sequence proofs. No old evidence relabel, no discovery/profile activation, no whole-row promotion and no claim to resolve the excluded policy questions.
