# Transactional discovery HTTP serving — bounded plan v2

Supporting implementation plan under `bdp-client-design/checkout/docs/architecture/bdp-client-v2/design.md`, not a second plan of record. Read-only preparation; no implementation authorization, tests or source changes. This replaces only the activation/31-exchange outline in `transaction-cold-discovery-serving-next-plan.md`; the separately approved limits and missing-Resource work retain their own ownership.

## Target, basis and gating

Deliver a useful cold path: Scope → actual service-desc → descriptor → directory/Type inventory → create → current read/list → receipt/retry → genuine reopen/rediscovery. Public server profile selection, claimEligible and cumulative qualification remain closed. A31-request subset proves its exercised path, not the entire Transactional profile.

Inspected exact limits candidate `fd07bf7ac177d5e53cabcd7666d9f2c77fba6e60`, tree `3b689aa3ef601307b5ee35b5de3c72f64048fcc7`, direct repaired-C3 parent3f486c2. Root identifies separate C3 test/policy repaira53f; its ordinary integration and exact final reviewed base must be verified before dispatch. Do not author against a mutable sibling or treat this inspection as acceptance of that uninspected delta.

**Wire activation is not yet ready.** Spec `Conformance profiles and reading guide` (bdp.md73–91) and accepted cold-client-v2 C4/C5 explicitly make a served `profile:transactional` a cumulative behavioral assertion, including on a private listener. No partial profile exists. A development header or false qualification flag does not make misleading advertisement lawful. The plan below defines the actual serving increment once the blockers below are closed/dispositioned. Settled internal service/minimum preparation and direct tests may be staged behind the unchanged existing HTTP refusal if root dispatches them separately; that is not cold-wire evidence. Do not add a bypass flag, test-only endpoint, synthetic descriptor or alternative profile to get an early green cold walk.

## Precise remaining gates

These are source-observed mandatory gaps, not row-count targets or a claim to have audited every inherited law. Recheck each on the dispatched base; completed independent repairs should leave this table rather than be duplicated.

| Source-observed gap | Required closure / decision boundary |
|---|---|
| Discovery current minimum and response context: service.discoveryObserved currently takes no minimum; HTTP533–556 refuses all three surfaces | This serving increment implements same-entry eligibility, finite wait and coherent headers/body. Settled existing law, no product question. |
| Current Type inventory/Type point, initial incident-Link reads, successful alias resolution reject complete minima: transaction-read26–32,112–125; HTTP533–547 | Separate settled consistency repair, enumerating current routes only. Type properties are **not** an existing route (`read-request.ts219`); do not invent one. Alias307 needs its own no-disclosure/context tests. Continuations remain separate. |
| Compatible/satisfied missing/deleted Resource minima still private-fail at transaction-read sourceCondition | Existing separately owned compatible-missing plan; integrate its reviewed implementation, do not implement it here. Missing/hidden combined with behind/foreign context remains a narrower precedence decision, not a reason to delay the compatible repair. |
| Retained collection/Snapshot minimum interpretation is refused | Existing unresolved anchor-versus-current-serving comparison. Preserve immutable original data/anchor/expiry; do not recapture or guess. Receipt current-serving law does not settle other retained surfaces. |
| Partial/repeated/malformed tuples and unrecognized position spelling are refused; commands with minimum fields also private-fail | Obtain exact receiving/applicability disposition from existing rulings, otherwise retain as narrow questions. Do not infer a mutation barrier from read-minimum law. Valid future pos-N waits already work. |
| Invalid set Selector/static complexity becomes development gap (`transaction-set.ts132–143`) | Resolve existing set-selector admission/retention phase decision; valid-set success is not evidence for these required failure classes. No new expression semantics here. |
| Incident Bead Event projection of metadata-only Link update throws gap (`transaction-events.ts259–264`) | Existing operator membership ruling and bounded implementation/finite+SSE proof. Does not block unrelated implementation, but unresolved required behavior blocks complete advertisement. |
| Scope changes SSE start=now plus Last-Event-ID returns gap (`transaction-http.ts435–450`) | Existing operator precedence ruling and bounded reconnect proof. Do not borrow Resource EventSource override semantics silently. |

Required diagnostic projection is already implemented in fd07; C1 collections, fresh Snapshot minima and sequence aliases are already delivered prerequisites. Optional History, Type lifecycle, optional multiplicity declarations, optional extra limit/retention advertising, administrative erasure/grants APIs, a full SDK, and a different backing store are **not** prerequisites. The five captured page/Selector values plus actual validation16/524288 stay unchanged. Optional omission never excuses incorrect runtime failure semantics.

**Recommended next concrete implementation while activation is held:** after the separately owned compatible-missing repair freezes, extend the existing current-minimum predicate/facet to present canonical Type point, initial Type inventory and initial `bead-links` requests, with same-entry context, eligible/behind/foreign/HEAD/conditional/abort tests. Keep alias resolution and every continuation out of that predicate. This closes a real current read-after-write gap without any new protocol decision; prepare its own small source/test plan rather than growing this discovery patch. Existing Event/receiving questions should be resolved in parallel by the root/operator stream.

## Exact serving behavior

Normative anchors: Scope discovery909–956; inherited methods/media/caching1937–1970; Transactional discovery3717–3776 and Operation Directory4305; Transactional HTTP consistency/conditional reads5758–5811. Line numbers are fd07 references; re-anchor by heading/symbol at dispatch.

| Surface | Required behavior / existing seam |
|---|---|
| Canonical Scope GET/HEAD |204, no body or invented JSON content type; Link `<actual descriptor URL>; rel="service-desc"; type="application/json"`; explicit `Cache-Control: private, no-store` plus current three context fields. `context()` only attaches context, so Scope must set its own cache field. Reuse existing probe path convention (`read-http-routing.ts`: matching Scope pathname, including its existing query handling); do not invent a new query grammar. |
| Descriptor GET/HEAD | Parsed C3 candidate:16 required fields+limits, actual seven scalar values, captured epoch/view/head matching **that response's** context, floorpos-0. No extra optional capability or target. JSON200; HEAD has GET metadata/body length, no body. |
| Operation Directory GET/HEAD | Parsed shared twelve-map directory, JSON200 and fresh owner context even though mapping is static. Relative values resolve against actual directory URL. Never serve it before auth or substitute a receipt collection. |
| Authorization/methods | Existing methodResponse precedes auth and retains405/Allow GET, HEAD without Problem; exact command targets stay Allow POST. CORS remains disabled, OPTIONS405. Authenticated read-only principal may discover. Anonymous GET/HEAD401 has no Link/profile/targets/context. Preserve raw Origin/auth/canonical authority handling. |
| Accept/conditions/cache | Shared applyReadHttpSemantics exactly once at handler tail. Descriptor/directory unacceptable JSON406 bodyless; Scope204 has no media type to negotiate. Existing If-None-Match:*304 and nonmatching If-Match412 apply only after real observation and eligibility. No headPosition-derived ETag or fake modification date. Preserve context/cache on304/412 and ordinary error behavior. |
| Minimum/lifetime | Complete canonical tuple checked against the same entry that supplies context; one finite wait, actual signal, no renewed deadline or owner entry spanning await. Foreign epoch410, foreign view409, future503 catch-up-timeout with exact current context and Retry-After1. Missing/unsupported tuple forms retain their existing private refusal until separately settled. No earlier Scope probe authorizes a later descriptor read. |
| Faults/immutability | Captured configuration stays immutable; descriptor head advances only from fresh owner observation, directory and all nested values remain immutable. Closed/reentrant/fenced owner or parse fault cannot produce discovery; arbitrary faults become sanitized bodyless500, never invented BDP Problems. Existing cancellation behavior is preserved outside new finite-minimum paths. |

## Minimal production work

1. Extend **existing** `discoveryObserved(minimum?, signal?)` in service. Snapshot supplied tuple synchronously; under one existing run lifetime, each attempt performs exactly one `owner.withRead`. Check minimum in that callback, copy/freeze three context values, return; assemble/parse **after** entry releases. Reuse `snapshotTransactionMinimum`, `checkTransactionMinimum`, `waitForTransactionMinimum` and existing finite policy. No store/config API, second builder, new barrier or timer. Without minimum preserve present synchronous observation path and document its interruption limit. Do not turn an unwaited behind result into timeout or relabel owner/validation faults.
2. At the existing private HTTP discovery-refusal location, identify the existing Scope pathname, exact descriptor and exact directory URLs after actual method/auth handling. Parse supported controls once; malformed/partial controls never default to absence. Route these three before the general Resource-only minimum guard. One `discoveryObserved` supplies response body and context. Scope explicitly sets Link/cache; other bodies use jsonResponse; all go through the normal tail once. Do not invoke createReadHttpRoute.probe for a second owner observation or use Resource-discovery projection from the Read facet.
3. Preserve canonical service/handler Scope equality, exact command-URL classification, all remaining unsupported-route refusals and Read/RU behavior. Replace only obsolete three-route refusal assertions/comments after activation gate passes. Keep public index admission, installed claim flags, IPC metadata, exports/schema/catalogs untouched.

If internal preparation is staged first, leave the three-route HTTP refusal executable and limit proof to direct service tests. Do not claim the31-wire ledger passed on that staged source. Root dispatches one explicit phase/base; do not let an implementation silently cross the wire gate.

## Owned files and verification

At most8 existing files, after ordinary integration of reviewed C3+limits+test/policy repairs:

- transaction-service.ts/.test.ts: optional tuple/lifetime and direct owner/spy tests.
- transaction-http.ts/.test.ts: three-route projection and actual HTTP matrix/disconnect tests.
- transaction-installed-contracts.test.ts: only reconcile the existing emitted discovery-boundary assertions and preserve actual receipt/Snapshot witnesses added by C3 repair.
- apps/bdptest/src/transactional-discovery-installed.test.ts: replace12/1 refusal fixture **only on activation** with the independently reviewed31/2 cold walk below; test-local navigation helpers, no executor/public API change.
- docs/design/transaction-implementation-progress.md: exact implemented/private/qualified distinctions.
- transaction-runtime.ts: at most the stale header comment saying no incomplete discovery claim; no runtime logic/configuration change.

Do not edit transaction-discovery/parser/schema/limits, missing-Resource code, shared read-http/routing or main Read/RU sources. Reconcile any newly located refusal test at the final base explicitly before adding a file. Source tests cover all routes GET/HEAD, POST/OPTIONS Allow, read-only/anonymous/Origin, JSON Accept and Scope distinction, same-entry context and immutable response, satisfied/behind/foreign minimum, behind conditional, single service call, synchronous tuple capture, independent commit catch-up, pre-abort/no attempts, close/drain and actual disconnect. Assert exact Problem/type/status/retry/Retry-After; HEAD bodyless with inherited metadata. Existing callback receiver/limit/getter tests remain intact. Controlled waits use existing test seams, not new runtime knobs.

Four local route/oracle controls: foreign-origin service-desc, canonical same-origin outside Scope path, corrupt directory name/target, descriptor head differing from its own context. Reject before credentialed execution; spy records zero executor calls. They add no installed requests and are labelled assertion sensitivity, not observed target failures. No extra mutant installation is planned; a real scope expansion needs a separate budget.

## Cold routine and bounded installed ledger

The routine receives only canonical Scope, credential handles/resolver, intended application payload and a bounded exact transport callback. Launcher-owned loopback dialing/ready hashes stay outside it; no ready.targets, hardcoded descriptor/directory URLs or configured Type ID enters navigation. Parse actual Link with existing parseLinkHeader, choose one unambiguous applicable service-desc relation, resolve against response URL, validate canonical credential-free HTTP URL and same Scope origin/path before granting an exact request. Validate parsed descriptor.scope equals initial Scope and all followed roots stay confined. Parse actual directory and resolve its values against its response URL. Do not follow redirects automatically.

Actual Type collection returns **Type summaries**, not full descriptors (`TypeSummary` id/name/describes). For this known payload choose the sole returned bead summary, independently assert the one-item inventory and next:null, and use its observed id; no schema/Type-descriptor fetch is needed for this fixture and no Type lifecycle claim follows. Author fixtures may know expected literals; the cold routine must not use them to select URLs. Returned Resource/receipt IDs get canonical kind/root/Scope checks before following. After reopen discard navigation grants and rediscover, retaining only genuine receipt/Resource identities for durability assertions.

Reuse the existing raw executor and declared credential firewall; epoch/minimum are raw fields, while BDP-Authorization-View is supplied by the existing C1/Snapshot declared credential-resolver adapter and checked against actual transmitted headers. Do not bypass reserved-header checks or log tokens. Each exact route binding records its observed parent response/hash and route tuple; per-request configuration varies as actual URLs are learned, so do not claim one static route digest for the entire walk.

**31 completed exchanges,2 sequential children,1 Bead,0 Links; ceiling36/2.** This deliberately supersedes C3's12/1 refusal budget only on the activated successor.45s aggregate includes startup/reopen,60s outer,5s each exchange,10s lifecycle. Three real future waits use existing1s policy; measured headroom required, no promised timing SLA or silent cap increase.

| # | Exchange / oracle |
|---|---|
|1|Anonymous Scope GET401, no discovery/context leak.|
|2–5|Authenticated Scope204 → observed descriptor200 → directory200 → Type summaries200; schema/exact membership, own-response context, complete one-item inventory.|
|6–8|DescriptorHEAD200, directoryHEAD200, ScopeHEAD204; bodyless, exact current pos0.|
|9–13|Descriptor If-None-Match:*304; directory absent If-Match412; descriptor unacceptable Accept406; directoryOPTIONS405; descriptorPOST405. Exact headers/context where applicable, no position change.|
|14|Observed createBead target, observed Type ID, literal intended properties; real completed receipt, allocated Resource identity/revision, required/currentpos1.|
|15–18|GET actual Resource; GET observed Bead collection at actual minimum; GET actual receipt; exact-key/body retry on discovered target. Full state/equality, next:null and singleton count, retained receipt equality, head remains1.|
|19–21|DescriptorGET, directoryHEAD, ScopeHEAD with satisfied minimum. Descriptor head equals its own context; no cross-request snapshot assumption.|
|22–23|Anonymous descriptor/directoryGET401, no Link/context/target leak.|
|24–26|Close/reopen same actual store and rediscover Scope→descriptor→directory with fresh grants. Genuine identity/configuration and currentpos1.|
|27–28|GET retained actual receipt/Resource IDs; complete retained receipt equality and literal current state/revision.|
|29–31|Futurepos2 at fixedhead1: descriptorGET, directoryHEAD, Scope conditionalGET each503; exact Problem/type/retry/Retry-After for GET, bodyless HEAD, actualpos1, no conditional bypass.|

No hidden Type fetches, polls or readiness HTTP calls. Source/emitted matrices—not this ledger—cover read-only credentials, all foreign/media/method combinations, catch-up and disconnect. This subset does not walk all twelve mutations, receipt pages or replication rendezvous and earns no automatic catalog credit.

## Done-when and evidence

1. Root records exact ordinary integrated base and disposition/closure of each applicable mandatory gate above before emitting a profile document. Remaining optional capabilities do not hold activation. No unresolved semantics chosen in code.
2. Source/matrix assertions establish same-entry data/context, one deadline, correct actual HTTP projection and unchanged unsupported-route/public admission boundaries. Focused affected tests and build/type/format/lint/boundaries/Read-seal checks complete; build precedes emitted tests. No automatic full-workspace run.
3. Final clean frozen source receives actual retained protocol/server pack+offline install and31/2 wire walk with explicit BDP_TRANSACTIONAL_INSTALLED_ENTRY and exact source variable. Record source/tree/dirty, tarball/payload/contracts/config/policy/auth/fixture bindings, every captured URL parent/hash and actual response hashes, started/completed counts, exits and cleanup. Failed partial runs preserve failure and actual completed responses, never fabricated completion. Source variable is launcher-declared, not target attestation.
4. Independent native+Claude code council and exact-head CI close; no child remains. Historical C3 refusal proof stays pinned to old source. Public admission/qualification require separate cumulative target-bound evidence; a31-step cold subset alone does not satisfy the sole plan's finished-product predicate.

No new product decision is needed for the three discovery routes themselves. The retained/receiving/Selector/Event decisions above already exist and remain exact blockers to truthful complete advertising. Until those are resolved, recommend the enumerated current Type/incident minimum implementation after existing missing-Resource work, while preparing/reviewing this serving plan rather than creating fake discovery.
