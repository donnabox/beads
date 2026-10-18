# H1b: private installed Transactional wire witnesses

Plan only, 2026-10-13. Source inspection is pinned to `d907a1e1576c5795c17edb5f15528f47edfd7d6b` in `vickie-replication-integration-validation`. Root's current integration `31aab4d` adds schema-provenance test/prose closure with the same runtime/transport semantics; its full exact identity must be recorded and the touched-file equality checked before authoring. This does not relabel d907's packed evidence as 31aab. H1b starts only after ordinary integration of frozen, reviewed H1a and M1. No source edits, tests, outward writes or admission changes in this planning task.

Authority: revised `transaction-installed-harness-plan.md`, separate M1 `transaction-runtime-configuration-plan.md`, and `transaction-installed-harness-revised-claude-review.txt`. This document makes the previously unresolved H1b mechanics mandatory. It does not supersede the sole fork design authority or add protocol semantics.

## Deliverable and predecessor gates

Run four bounded partial Transactional observations through real HTTP against a fresh packed/installed private candidate. Retain its actual bytes/configuration, source, process and response provenance. Every report remains `claimEligible:false`; all four rows remain `wholeContract:false`. Required canonical selection and missing rows come from H1a, not a hand-built shortened conformance catalog.

Prerequisites:

1. H1a's private v2-only bundle/runner path, canonical 250-row selection and public-run refusal at both declared TX and RU profiles are reviewed and integrated. H1a currently owns `runner.ts`; no simultaneous edit or assumption that this older source contains its gate repairs.
2. M1 supplies actual `runtimeConfigurationIdentity`, unchanged compiled `artifactIdentity`, policy/authentication digests including credential-level `canMutate`, explicit canonical Host mode and actual loopback port0 binding. H1b never manufactures these from expected data. Store compatibility JSON is not a runtime digest.
3. H1b's own exact-transport and dynamic recorder changes below receive review. A plan marked ready-unexecuted is not an executable TX transport until these changes exist.

## Required trusted exact-transport admission

At d907, `ExactHttpModeOptions.profile` in `packages/conformance/src/http-executor.ts:57` and `snapshotExactHttpMode` at `:443` admit only Read/RU. Extend those exact two profile guards/types to accept `transactional`, then run the existing snapshot path. Do not clone the executor or introduce a second firewall.

Preserve byte-for-byte or behaviorally unchanged `snapshotExactRoutes`, reserved header rules, credential handle/secret header validation, Scope containment, method/URL matching, no Host override, request-write observations, limits and cancellation. Every TX route must still be frozen and exactly matched before credential resolution or dialing. Trusted low-level TX transport acceptance is **not** public Transactional runner/profile admission: the ordinary `runConformanceMatrix` retains H1a's unconditional private-TX-bundle refusal regardless of declared profile. Read/RU exact configuration and runner behavior remain unchanged.

Countercontrols: actual snapshotted TX profile succeeds only with valid exact routes; wrong method/URL, extra fields, foreign Scope, reserved Host and malformed credential handles fail before credential-bearing dial. Existing Read/RU suites run unchanged. Public TX and RU-declared/private-TX invocations still refuse.

## Real installed factory, process and metadata binding

New private emitted modules:

- `packages/server/src/transaction-installed-configuration.ts`: closed bounded plain administrative inputs only; stable Scope, directory/create intent, loopback port0, credentials and exact contract artifacts. Require M1's default compiled path and canonical authority mode. No scheduler, clock, retention shortening, operation/page overrides, fake outcomes, generic policy or test controls.
- `transaction-installed.ts`: starts the one M1 runtime; obtains actual immutable returned metadata; binds actual installed payload using the shared binder; no second owner or Read plane.
- `transaction-installed-process.ts`: inert import and inherited-IPC-only launch; one configuration then close-only messages. Bounded ready only after real startup; no late ready on disconnect/close; actual drain/reopen cleanup and cause preservation.
- Extract existing installed payload hashing unchanged into private `installed-payload.ts`, used by RU and TX factories. Preserve exact traversal/order/dependency/symlink laws; new package bytes naturally yield a new payload digest. Do not relabel prior artifacts or invent executable dependency claims for erased type imports.

The parent captures expected plain configuration and validates child-reported actual metadata, then constructs a **private TX execution binding** containing (a) existing ServedProfileConfiguration, (b) actual `runtimeConfigurationIdentity`, and (c) current physical listenUrl/closed targets. Do not add a field to the shared closed ServedProfileConfiguration just to avoid this explicit pairing. The compiled identity is passed unchanged and decoded with the existing `installedContractsDigest`; no new digest format.

Validate expected versus actual at bootstrap, each new scenario's startup **before setup or first HTTP request**, and each reopen. A reused process keeps its immutable startup binding plus live process identity; no fake refresh IPC. Reopen requires a fresh actual ready observation and a newly validated dial endpoint. Logical runtime identity stays stable for the same configuration; an allocated port change is separately rebound. Bind sourceCommit from verified pack provenance, not child assertion.

Permission swapping across two tokens of one principal must change real authentication/policy/runtime identities and actual POST authorization. Changing only runtimeConfigurationIdentity with all other fields held equal is explicitly an injected mismatch control, not a real independently variable captured configuration. Similar identity/dial tampering controls fail before the first request. No expected-vs-expected comparison is called observation evidence.

## Fixed request grammar and finite fixture

Use explicit Resource IDs and at most eight operations per request. Predeclare all operation POSTs and Resource/Event GETs in the base exact route set. For each collection GET, predeclare a literal `?limit=32` (within actual maximum32), keep the fixture inventory below that count, and require `next:null`. A non-null next is a failing fixture assumption, not permission to fetch an unlisted cursor or silently omit records. Point GETs can provide additional independently authored checks. There is no collection cursor adapter in this slice.

All ordinary reads use the authorized operator credential; an explicit anonymous handle probes401. Same-view permission/identity rules follow actual runtime behavior, never carried actor metadata. Canonical URLs are derived from stable Scope, while the raw executor's trusted physical dial map points only to the validated actual loopback endpoint. Host derives from the canonical URL and is accepted by M1 canonical mode; do not relax reserved Host controls.

No receipt page adapter: production receipt maximum32 and direct maximum32, with this fixture ≤8 operations and HTTP sets closed, make a receipt continuation unreachable. Readable/restart pagination variants remain missing, never skipped or tested through lowered limits. No discovery claim: privately configured targets come from the actual handler/runtime inventory, not a schema-invalid Transactional Service Description.

## Mandatory dynamic receipt action and private recorder

A returned receipt URL cannot be present in the base frozen route list before execution. Required implementation, not an optional future extension:

1. Add a source-private action such as `family:"transactional.receipt-read"` to the private TX bundle/parser/runner dispatch only; it is **not** `family:"http"` and is refused outside the private TX path. It names a prior action's validated receipt capture and method GET/HEAD, not an arbitrary URL. No public manifest/normative schema action is added.
2. Wrap actual base execution with a bounded immutable per-scenario observation ledger. Admit a receipt target only from `id` of a schema-valid actual completed/failed POST receipt response or explicitly linked prior receipt response. Record source action/response identity, source JSON pointer, exact canonical URL, immutable TX execution binding and originating credential/view association. A fixture binding or caller-supplied string alone confers no follow authority.
3. Validate receipt URL against the actual d907/31aab grammar: same canonical Scope/origin, exact receipt path with one valid identity segment, no userinfo/query/fragment/encoded traversal. Reject `?view=events`, `?alias=...`, other alias-shaped queries, operation/Resource substitution and foreign targets before dialing. Current HTTP code classifies changefeed first and alias-shaped queries before receipts; do not accidentally test an Event/alias route as a receipt lookup. URL must equal the actual observed capture and belong to this scenario and allowed credential. No arbitrary credential switch.
4. Snapshot a **separate single-route exact configuration** for that captured URL/method, with the same validated physical binding and fixed caps. Execute through the existing raw executor; keep the base route list and its `exactConfigurationDigest` unchanged. The dynamic action cannot bypass that base restriction by pretending its URL was predeclared.
5. Extend the private runner recorder now. Proposed closed per-action evidence fields: `dynamicExactConfiguration`, `dynamicExactConfigurationDigest`, `sourceReceiptObservation` (action/response/digest/pointer linkage), and immutable served-runtime/physical binding reference. Names distinguish this route from the base exactConfigurationDigest. Reuse the existing exact observation validator against the dynamic configuration, not the base configuration that necessarily lacks this URL. Exported report shape is private and bounded; public report/manifest semantics remain unchanged.
6. Preserve original actual status, header lines, body octets/text and request-write observation. Parse the received JSON for assertions without constructing a replacement successful response. Existing secret header redaction/fingerprints apply to export; raw authorization must never reach report/log output. Assertions consume a projection linked to the original exchange, not a reserialized synthetic HTTP reply. Do not use synthetic-manifest approval on positive runs.
7. Replay/GET/HEAD assertions and dynamic provenance are recorded by the same private runner result, with expected exact configuration and source linkage checked before assertions count. Invalid dynamic ledger/action data fails as harness error, never an accepted server response or silently ignored row.

Counterfactuals include unobserved URL; same-Scope wrong identity/path; wrong source pointer/status/body; query-bearing receipt; cross-scenario capture; altered method/credential; changed runtime or physical binding; absent/incorrect exact request observation; body digest mismatch; substituted prior response; and modified dynamic config digest. Each fails at its intended pre-dial/evidence assertion. Positive evidence is never generated from those mutations. Verify the base frozen config digest remains unchanged after dynamic reads and that every dynamic exchange names its separate digest.

## Four partial observations and honest gaps

- **Atomic commit:** explicit-ID Bead create with name, backward local-name update, second Bead and Link. Verify receipt indices/names/intermediate postimages, actual revisions and independent literal final graph through point/limited collection GETs. Forward/pinned/alias references and concurrent isolation remain unobserved.
- **Rollback:** seed over HTTP, capture original full Resource; update/create prefix followed by a real stale guard. Failed receipt has proper index/problem/no committed result, original full afterstate/revision remains, new identity absent, finite Scope Events unchanged. Fresh follow-up write succeeds. Final-state checks do not prove transient invisibility.
- **Receipt readable:** follow actual completed/failed receipt IDs through the dynamic action; GET body matches retained facts, HEAD is bodyless with required context. Terminal close/reopen preserves receipts and actual final graph; current headers may advance after writes. No pending/fenced recovery or pagination claim.
- **Failed retained:** first carrier fails on missing Resource; create it separately; exact original retry remains failed with no new effect; fresh key succeeds. Changed carrier under old key conflicts. Also distinguish completed retry's old postimage from separately advanced current point read. Expiry/duration/principal transfer/cross-carrier/pending variants remain missing.

Finite Event observations assert actual types/order/source/ordinal/position relations. Finite changefeed exists in current private runtime but is **not observed or advertised by this H1b cohort**; do not conflate it with Events. Snapshot/SSE/cold discovery/Directory, remaining minimum controls, sets/sequence variants, mutable policy and erasure, general HTTP/CORS/absolute-target gaps and all unexecuted canonical rows remain explicit. Existing historical Read seal remains unchanged.

## File ownership after ordinary predecessor integration

H1b new server modules above plus `transaction-installed.test.ts`; `apps/bdptest/src/transactional-matrix.test.ts`; source-only `apps/bdptest/test-support/transactional-exact-client.ts` and `transactional-execution-binding.ts`; `scripts/transaction-installed-observation.mjs`. Shared server edits only `installed-reference.ts` for unchanged payload helper reuse and smoke inventory. M1 runtime/configuration and owner fields are consumed, not reimplemented here.

After H1a freezes/merges, root assigns the narrow `runner.ts` dynamic action/recorder and private parser/type-test support; ordinary integration preserves H1a's branding/public gate fixes. `http-executor.ts` edits are limited to profile type/snapshot acceptance plus dedicated firewall regression tests. No broad route-rule rewrite. Update only private source-only TX fixture/manifest/report validation for this action, not published matrices/catalogs or schema worker's RU work.

Every new private emitted server module—`transaction-installed`, `transaction-installed-configuration`, `transaction-installed-process`, `installed-payload`—gets `.js`, `.js.map`, `.d.ts`, `.d.ts.map` in the exact output inventory and both `@bdp/server/<name>` and `@bdp/server/dist/<name>.js` denied-import controls. Do not rely solely on the transaction-prefix auto-filter, which tests only dist spelling and misses installed-payload. M1 runtime-configuration inventory is preserved from its predecessor. No package exports/bin/public startup added.

## Budgets and completion predicates

Exact mode: maximumRequestHeaderBytes **16,384**, maximumRequestBodyBytes **65,536**, raw-wire evidence **2,097,152**. These are explicit preimage inputs, not RU's larger copied header setting. Clarify Node's header accounting versus executor serialized-head accounting in tests rather than asserting every byte-bound implementation is identical. Initial bodies <1KiB properties and ≤8 operations; source-only TX manifest≤1MiB; existing artifact/compiler and cohort limits unchanged. Child ready/output≤64KiB, no credentials.

Startup/close≤10s, each request≤5s, cohort work≤120s; separate bounded cleanup and forced termination reported as failure. These harness deadlines are not protocol timing claims. Node24.16.0/pnpm11.20.0, build before emitted tests, maxWorkers2, verbose/JSON logs outside source.

Done when: actual packed protocol/server installed in a fresh directory; exact installed process/closure path and tarball/source/payload/configuration hashes recorded; all four positive partial witnesses execute (no silent skip) and negative controls fail for the intended reason; no HTTP/setup before binding validation; reopen rebinds physical endpoint while retaining actual logical identity and terminal facts; source/public/deep-import gates remain closed; missing inventory remains visible and claimEligible false. Startup refusal, parent disconnect, abort and assertion failure all close and await the child, preserve cleanup causes and release the owner lock. No test-owned children remain. Build/typecheck/boundaries/format/lint/both smokes/public gates and unchanged evidence verification pass. Freeze exact source/report for independent review; do not promote profile or reseal historical Read evidence.

## Review disposition summary

F1 accepted: explicit exact TX type/snapshot admission, no public run admission. F2 accepted: private dynamic action/recorder mandatory with separately bound config/digest. F8 accepted: predeclared limit32 collections require next:null; no collection cursor adapter. F9 accepted: all four private modules have complete emitted inventory and both import spellings denied. F10 accepted: current integrated routing grammar, actual feed existence versus unobserved cohort, and exact source repin. F11 accepted: pin16KiB headers/64KiB body/2MiB wire cap. M1 F3–F6 are enforced here through actual runtime metadata and binding checks, with injected mismatches distinguished from actual changed captures. No optional wording remains for these known required seams.
