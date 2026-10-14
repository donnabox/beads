# Current alias minimum reads — bounded feasibility and implementation plan

Supporting plan under the sole `bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md`. Read-only preparation, not implementation authorization. A small settled-mechanism slice is feasible for the existing fixed all-visible authority. It need not choose a new alias-versus-target authorization rule, resolve incompatible missing precedence, or wait for discovery activation.

## Basis and exact remaining gap

Inspected frozen9b95df5dd866dbaf84150404c4db86ac8054b724/tree d6749c7da26a5eeb2debce8336af91acc279249a and its direct compatible-missing successor7791aed7394e30fef2d2c1ed4fed3f43504e42ba/tree902349503578491c8d5bfda757e1600d19811765 in `vickie-compatible-missing-minimum`. Root must assign the final reviewed ordinary integrated base before authoring;7791 is context for this plan, not a waiver of its pending council/integration gates. Preserve the current Type/incident plan as separate work. Coordinate overlapping read/HTTP/test files through root and ordinary ancestry, never simultaneous edits.

Three explicit refusals/omissions currently prevent a complete supported tuple on an otherwise successful alias read:

1. `transaction-read.ts`, `minimumFor.resolveAlias`, rejects every alias instead of forwarding a source condition to the existing authority facet.
2. `transaction-service.ts289–308`, `readAtMinimum.resolveAlias`, also rejects instead of using the existing run/wait lifetime.
3. `transaction-http.ts196–213` always resolves aliases through ordinary readFor and does not forward Request.signal; its general minimum gate around525–538 rejects any alias. The existing redirect already attaches `transactionReadContexts(result.observation).current` and emits307/Location/private,no-store/content-length0.

The underlying source needs no new alias reader. `authority-read.ts769–787` reads alias mapping, validates its canonical in-Scope Bead target, looks up that target and returns target plus the captured observation (or observed404) within the guarded entry. `observe`977–998 runs sourceCondition against this observation before releasing the entry. The generic alias delivery path already accepts signal/sourceCondition. No second target GET is needed to assemble a redirect and no owner entry should span a finite wait.

## Existing law and what this slice does not infer

All citations are `docs/specs/bdp.md` at the frozen basis; use headings/symbols when reanchoring.

- **Aliases254–269:** aliases are locators, not Resources; they have no Resource revision and changing an alias changes no Bead revision. Resolution is one step to an in-Scope canonical Bead.
- **Alias resolution1422–1439:** GET/HEAD returns temporary307, absolute canonical Location and no body. Unknown/query/fragment aliases use uniform404 under the same authorization projection. The redirect target is authorized on follow; resolving an alias itself asserts nothing about target readability. This slice must not reinterpret that last sentence as permission to remove existing visibility checks or expose targets under a different policy.
- **Transactional alias mutations4188–4208:** alias mutations occupy no Scope position, Event, change group or snapshot; a successful alias effect can omit effectPosition. Its requiredPosition is an ordinary serving-view barrier and does not certify alias-effect replication. Thus a minimum does not pin a historical alias mapping or act as an alias revision. Repointing can change Location while the position stays identical. A catch-up test needs a real Resource mutation, not an alias put pretending to advance the feed.
- **Strict reads/Transactional HTTP consistency5758–5787** and **Resource read context5793–5799:** a current Scope-bounded alias observation can report its actual current prefix and require an eligible view/position. Existing alias routing already attaches that context. Apply the same owner-entry eligibility/wait mechanism before releasing Location; no target fetched outside that entry supplies context.
- **Conditional reads5801–5811:** eligibility precedes any conditional shortcut. Existing shared `applyReadHttpSemantics` only applies JSON/204 conditional negotiation to its success representation paths.307 is redirect-only, so this slice preserves307 for If-None-Match/If-Match and does not invent an alias ETag,304,412 or JSON Accept406. A future minimum must still yield503, never leak307/Location. HEAD is bodyless and no mutation.

**Authorization/absence limits:** the private owner uses `canRead:()=>true` and no distinct alias-policy API. Generic authority currently checks target existence/visibility before returning a target. Keep that behavior; this plan does not qualify hidden-target policies or adjudicate whether a different implementation should disclose a redirect to an unreadable target. For a syntactically valid missing/deleted alias or dangling target, reuse7791's existing callback: compatible/satisfied observed404 includes actual current context, while missing plus behind/foreign remains the explicit visibility gap. No wait, foreign Problem or success is substituted for that collision. The existing malformed alias classifier runs earlier and returns its context-free404, even with minimum fields; keep that local route ordering, not a new global precedence rule. Raw canonicality remains the listener's job.

No unavoidable product decision blocks this fixed-policy current success slice. Hidden alias/target authorization and incompatible missing arbitration remain unqualified boundaries. If the assigned base changes those rules, stop and request root disposition rather than infer new law.

## Minimal seven-file implementation

1. **packages/server/src/transaction-read.ts:** implement minimumFor.resolveAlias using the already captured minimum, underlying facet.resolveAlias(url,{...options,sourceCondition:minimum}), observed promise handling and existing checked(result). Preserve ordinary readFor's signal-only forwarding and the7791 callback exactly. Do not add an alias-specific context source or modify the Resource/collection subset.
2. **packages/server/src/transaction-service.ts:** implement readAtMinimum.resolveAlias under the same `run` + `waitForTransactionMinimum` policy/signal as perform. Snapshot the tuple before asynchronous attempts as the enclosing method already does. No renewed deadlines, callbacks under SQLite locks, new configuration or generic runtime abstraction.
3. **packages/server/src/transaction-http.ts:** after unchanged method/auth/alias-shape/tuple validation, permit a classified well-formed alias through the minimum route gate without attempting to reinterpret it as a canonical ReadRequest. Preserve the existing Resource gate for nonaliases, including any independently accepted Type/incident extension. In the alias callback select readAtMinimum when the complete tuple is present, otherwise readFor, and pass Request.signal. Reuse the current307/Problem/context projection and final shared HTTP semantics exactly once. The additional ordinary-alias signal forwarding is an intentional narrow cancellation improvement; do not claim that in-process/wire abort behavior is wholly unchanged.
4. **packages/server/src/transaction-read.test.ts:** real-owner conditioned alias tests and service-lifetime cases using existing fixtures/imports. Replace only now-obsolete minimumFor alias rejection assertion; preserve ordinary sourceCondition-ignore controls and other unsupported subset tests.
5. **packages/server/src/transaction-http.test.ts:** exact redirect/context and refusal tests; update only obsolete alias-minimum gap expectations after locating them on the assigned base.
6. **apps/bdptest/src/transactional-collection-minimum-installed.test.ts:** new separately labelled18/1 scenario using the accepted existing raw minimum-header/credential/lifecycle helper. Do not modify the older alias sequence installed test's fixed36/2 accounting or its unrelated recovery oracles. Existing collection/missing/Type scenarios remain preserved and separately counted.
7. **docs/design/transaction-implementation-progress.md:** reconcile alias minimum exclusion, locator-only position semantics and exact executed subset.

No shared authority/parser/HTTP-semantics, store, runtime/launcher, schema/catalog/coverage, discovery projection, public export/admission, Type lifecycle or optional mechanism changes. Additional files require root approval. Production implementation is three narrow forwarding/gate edits; tests establish the semantics already supplied by existing modules.

## Meaningful source and actual HTTP checks

- Real put to a, resolve with equal/earlier eligible minimum; exact canonical target and same-entry epoch/view/current position. Repoint to b without advancing position, resolve b at the same minimum; neither target Bead revision changes because of the alias put. Establish these through actual owner/receipt observations, not invented revisions.
- Real finite catch-up: begin alias read requiring the next Resource position. During the existing controlled wait hook repoint a→b and perform a real effectful Bead update to reach the bound; result must be b with the fresh eligible context. This defeats returning a cached target captured before waiting. Separately show alias-only progress does not satisfy the future Resource bound. Guard against holding an owner entry while awaiting by doing the independent real write successfully.
- Foreign epoch/view on a live alias yield the existing410/409 context-bearing errors; recognized future bound eventually503 catch-up-timeout with exact retry/Retry-After and final attempt context. No Location on any refusal. Unknown spelling/partial tuples remain the old receiving gap.
- Compatible absent/deleted alias and dangling target use existing observed404/context; incompatible cases retain VisibilityGap/no context/no wait. Malformed query/fragment classifier tests retain context-free404 before minimum; do not call a URL fragment stripped by an HTTP client an actual wire fragment witness. Existing raw-target tests remain the authority for canonicality. Corrupt stored target/context/provider faults remain integrity failures, not a borrowed404.
- GET and HEAD success exactly307, absolute Location, content-length0/private,no-store, no body, Resource ETag or fabricated target representation. Conditional and unacceptable-JSON Accept headers do not change a valid redirect. Future conditions cannot return307. Anonymous401 has no Location or Scope context. Read-only credentials may resolve, without exercising mutations. Methods retain405/Allow GET, HEAD.
- Pre-aborted actual Request must reject before alias/target owner SQL. Abort during finite wait terminates with the existing abort classification and no later Location; ordinary alias requests now pass that signal too. Existing close/drain lifetime must cover resolveAlias just as perform; no new deadline or unsupported interruption of synchronous SQLite work. Verify owner reusable after failure.
- Bounded sensitivity: ignoring the alias condition must fail the future/foreign tests; caching first-attempt target must fail the repoint-during-wait case; substituting Location with the old target or context with an older prefix must fail the same response oracles. Use existing controlled fixtures and disposable source mutants only if needed, no runtime test hooks. Local corrupted response controls prove exact Location/context oracles without extra installed requests.

## Actual installed witness:18 requests,1 child,7 URLs

Reuse the collection fixture's two existing Type contracts and transport; create only2 Beads and0 Links. Maximum batch size2. Count every POST/read locally with ceiling18/final18 and7 exact URL identities. New scenario label is distinct; old scenarios remain byte-preserved except separately reviewed necessary expectation changes. No hidden discovery, polling, extra schema/Type fetches, reopen or redirect following.

|#|Request and independent observation|
|---|---|
|1|Batch creates literal Beads a,b; exact two receipt results/full authored fields, independent effect/required/current pos-1 and valid opaque revisions.|
|2|putAlias(work→a), fresh key; exact alias result, required/current pos-1, absent effectPosition. No Resource postimage.|
|3–4|Alias GET then HEAD with actual minimum from#1:307/Location a, no body, exact current pos-1/context.|
|5|Alias GET at same minimum plus If-None-Match:* and mismatching If-Match: still307 a, no alias ETag/conditional shortcut.|
|6|Alias HEAD with Accept:text/plain at satisfied minimum:307 a, not fabricated JSON406.|
|7|Fresh-key putAlias(work→b): exact repoint result, required/current pos-1 and no effectPosition.|
|8|Alias GET with same minimum:307 Location b. Same prefix does not freeze locator state.|
|9|Independent canonical b GET with same minimum: complete authored Resource and its original creation revision, demonstrating an ordinary follow as a separate read. No claim about hidden-policy authorization.|
|10|Alias GET with future pos-2 and conditional headers:503 canonical catch-up-timeout/retry/Retry-After1, actual current pos-1, no Location.|
|11–12|Live alias HEAD foreign view409 and GET foreign epoch410: exact context/body projection, no Location.|
|13|Unknown alias GET at satisfied bound:404 Problem/current pos-1, no Location.|
|14|Unknown alias HEAD future pos-2: bodyless visibility500, no context/Location.|
|15|Fresh-key deleteAlias(work): exact alias deletion result, required/current pos-1, absent effectPosition.|
|16|Deleted work alias GET at satisfied bound:404/current pos-1, no Location.|
|17|Anonymous work alias GET with minimum:401, no Location/context.|
|18|work?invalid=1 GET with recognized future bound: existing classifier404, no context/Location; does not enter the wait.|

URLs: batch, put-alias, delete-alias targets; alias/work, alias/never, alias/work?invalid=1; canonical beads/b. Actual Method/URL pairs receive exact route grants. No client auto-follow; #9 explicitly follows independently checked canonical Location. Epoch/view/positions are captured from real responses; literal pos-1 independently represents the one effective Resource batch. Alias commands do not change that position or prove replication of locator state.

Use inherited90s aggregate/120s outer/5s exchange/10s lifecycle,32768 response-body/16384 header/65536 request-body limits and existing general80 guard unchanged. One real1s future timeout is included. Preserve cooperative signal through startup/exchange/close, genuine failed/completed-response reporting and cleanup. Source catch-up/abort controls add no installed child or requests. Report measured headroom, not a performance promise; stop rather than raise bounds.

## Done-when, evidence and acceptance

Root/council accepts this plan and assigns an exact frozen ordinary base including accepted missing behavior and any earlier Type/incident edits. Verify seven-file inventory and preserved existing scenarios/route exclusions before authoring. No production change to the alias lookup or policy itself.

Pinned Node24.16/pnpm11.20 clean prebuild typecheck, build, focused Read/service/HTTP/consistency regressions, format/lint/boundaries and existing Read-seal verification pass. Real source assertions distinguish all planned positive/refusal/cancellation/cached-target cases; inspection alone is not acceptance. No automatic full suite unless failures or overlap justify it.

Freeze clean source, actual pack/offline install, set explicit installed entry/source variables, execute new18/1 plus narrowly selected unchanged regressions separately. Retain commands/exits/timing, exact source/tree/dirty and package/payload/configuration/fixture/policy/auth identities, safe request/response hashes, actual counts, child exit and no remaining process. A launcher source variable is not target attestation. Preserve failed attempts and all older evidence at its original source.

Root owns council closure, publication and integration. No new whole-row flags, profile claim, discovery advertisement or optional alias policy mechanism. The successful fixed-policy minimum path can be implemented independently; missing+incompatible context and hidden-target authorization remain explicit unqualified boundaries rather than invented precedence.

Planning performed no tests/builds/network/servers or source changes. Only this supporting plan was written.
