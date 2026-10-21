# Local schema serving: Read+Update and private Transactional audit

## Recommendation

**Preserve Update's existing serving path; no new Update production repair is justified by this bounded inspection.** The installed Read+Update candidate already hosts exact captured local schema retrieval documents, including any transitive documents present in its accepted schema bundle. PR119 adds a separate private Transactional path; it does not need an Update rewrite or shared-router abstraction.

The smallest existing targeted preservation selection is the exact-byte/schema-write test in `apps/bdptest/src/installed-reference.test.ts:197` plus the static-route owner-gate test in `installed-reference-startup.test.ts:218`. The existing schema-cohort capture test provides a cheap additional identity/manifest check when that fixture changes. None was run during this audit. Root can schedule those selections against the final combined source if desired; they are not automatically part of the PR119 evidence reported here.

A real cross-profile cold-traversal residual remains where a schema's canonical `$id` or embedded resource ID differs from its captured retrieval identity. Both implementations can validate such an installed closure without serving every URL a cold client might derive. That is already explicitly deferred by PR119's accepted subset, not a reason to reject previously accepted configurations or silently fabricate alias routes. A future narrowly reviewed retrieval-identity/alias design could address it; no new semantics or serving grant is selected by this audit.

## Basis and evidence limits

Read clean frozen `/Users/dbox/repos/wt/vickie/vickie-local-schema-serving`, HEAD `c555fe01df5cd20454fb0d663431df816c1270aa`, tree `062a3ab7e24236417bd915b373a4b44ba9908647`, parent `4aa2d0be950f157f3f0617278dd37c40f41ec035`. Inspected the sole design.md and its current Transactional-first goal, accepted schema plan/dispositions, actual Update/TX source and targeted tests, and `transaction-cold-schema-serving/report.md`. The nine-file PR119 delta contains no Update runtime/configuration/composition changes.

No tests, builds, services, installations, network actions, source edits or outward writes occurred. Only this supporting report is written. PR119 author evidence is reported evidence, not independently re-executed here; its native/external code council and public gates remain root-owned.

## How Update actually works

1. **Capture and authority.** `packages/server/src/installed-reference-configuration.ts:123–159` descriptor-safely copies input artifact strings into a bounded bundle and compiles it with `compileDevelopmentContracts`. Compiled contracts retain exact immutable captured bytes, fingerprints and offline reference closure (`development-contracts.ts:56–96`). The same locally copied bundle yields the serving manifest; it is not a later network fetch or mutable caller-object reference. The manifest uses retrieval URI plus original UTF8 string, not reconstructed schema JSON. Capture/configuration identities bind contracts and policy/authentication (`installed-reference-configuration.ts:165–202`).

2. **Deployment restrictions already exist.** Update's installed candidate requires descriptor IDs under `${scope}types/` and every supplied schema retrieval URI under `${scope}schemas/`, rejecting nonempty URL search/hash (`installed-reference-configuration.ts:149–159`). It therefore rejects external or outside-namespace **retrieval artifacts** rather than silently allowing them as this candidate's hosted closure. This is an existing private deployment restriction, not universal BDP law. It does not require schema `$id === retrievalUri`; canonical schema IDs may still differ, including externally based IDs in locally retrieved documents. Do not confuse the two identity classes.

3. **Installed factory to runtime.** `installed-reference.ts:8–31` captures configuration, creates the private cumulative Read+Update composition, and passes both the same compiled authority configuration and the closed schema manifest into `startDevelopmentReferenceRuntime`. The returned result is explicitly `claimEligible:false`. The authority installs `configured.contracts.types`/descriptors into its Read plane (`development-authority.ts:146–147,197–203,266`). The composition delegates Type inventory/descriptors to those ordinary reads (`development-composition.ts:44–81`); returned descriptors retain propertiesSchema.

4. **Lookup and authorization.** `development-runtime.ts:191–220` captures/freeze-copies the private transport manifest and losslessly decodes a representation for HTTP selection. At305–333 the request must pass runtime readiness, trusted credential, principal view, raw target/Host handling, and the runtime's applicable Origin behavior. Exact `entry.uri === url.href` selects a schema; it is not a prefix grant. At334–343 `principalAuthority.discover()` gates the static representation through the actual owner/read lifecycle before GET/HEAD method handling. `development-authority.ts:295–302` shows this is a principal-scoped read operation, not a fabricated permission boolean. Unknown schema URLs continue through ordinary routing and become404.

5. **Exact response bytes.** `development-runtime.ts:344–365` applies the shared Accept/conditional helper, preserves private/no-store and non-attesting marker, then sets successful Content-Length from original UTF8 byte length and emits that original string only for non-HEAD200. HEAD has the GET representation length with no body;304/412/406 remain bodyless. It does not fetch or reserialize a schema to serve it. `read-http.ts:10–55` supplies no-ETag wildcard/specific preconditions and preserves selection metadata. No Transactional epoch/view/position is added to this nontransactional profile.

6. **Root versus transitive documents.** Configuration maps **all** supplied schemas into the manifest, not just descriptors' root propertiesSchema documents. Thus a root and its captured transitive schema at eligible exact retrieval identities both have the same route by source inspection. Offline compilation already requires the complete referenced closure. The route is not an online `$ref` resolver: it serves documents, and the client resolves reference bases/fragments. A document's `$id` alias or embedded-resource ID is not automatically another HTTP route.

## Comparison with frozen PR119

| Boundary | Existing installed Update | PR119 private Transactional |
|---|---|---|
| Accepted artifact configuration | Requires local descriptor/schema retrieval deployment as above | Preserves broader accepted offline captures; filters a supported local retrieval subset |
| Manifest source | Copied input bundle that also feeds compiled capture | `contracts.capture.artifacts`, exact captured bytes |
| Schema routing | Existing exact manifest lookup in development runtime | New source-private `transaction-schema-http.ts:9–69` exact manifest map |
| Owner/lifetime gate | Principal-scoped discover/read gate before static serving | Narrow nonexecuting schemaObserved current-owner observation |
| Context | Read+Update revisions/HTTP semantics; no TX tuple | Actually observed current epoch/view/position on selected local schema response |
| Minimum controls | No inherited Transactional minimum obligation | Shared presence guard delegates every minimum-bearing request to existing refusal handling |
| Delivery | Existing schema raw-string branch | Existing private deliver extended with optional raw string (`transaction-runtime.ts:87–94,168–171`) |
| Discovery/admission | Private diagnostic cumulative composition exists; public admission still closed | Discovery still explicitly withheld; public admission closed |

The implementations intentionally retain their own existing method/auth/Origin ordering. For example, Update's installed schema path authenticates and owner-gates before its method refusal; TX retains its earlier method gate and Origin refusal before adapter dispatch. This audit does not assert identical precedence or recommend transplanting one router into the other.

Both share the unchanged captured compiler and HTTP selection helper. Neither needs to import Transactional stores, receipts, epochs, comparison loops or Event machinery into Update. Public `admitReadServerProfile` at `packages/server/src/index.ts:288–311` still rejects profiles other than Read; the private candidate factory and diagnostic discovery do not bypass that boundary. Existing schema serving is not complete cumulative qualification.

## What the existing tests actually prove

**Minimal installed-child preservation test:** `installed-reference.test.ts`, exact name:

`hosts exact captured local schema bytes and validates real HTTP writes with inherited pinned contracts`

Its source at197–254 has one local root schema, explicit preserved whitespace, GET bytes/marker, HEAD bodylessness/byte length, unacceptable Accept406, OPTIONS405, wrong credential401, unknown schema404, invalid write422 plus absent-record GET, and valid write200. Static path count is nine HTTP requests and one child; this count is inspection, not a newly recorded run. Despite the title, this fixture has `conformsTo:[]`, no `$ref`, and starts from an authored schema URL. It does **not** itself establish a descriptor-to-transitive-schema cold traversal, inherited schema conjunction, all conditional cases or `$id` alias hosting.

**Minimal source owner-gate regression:** `installed-reference-startup.test.ts`, exact name:

`checks the actual owner before serving a static directory or hosted schema`

At218–261 it starts the actual candidate factory, requests operation directory and schema, injects a principal discover refusal, requires503 rather than static data, then restores and observes200. The mock wraps the actual owner and targets refusal propagation. Six requests over one in-process listener, no installed-child claim. This protects against a static schema route bypassing owner state; it is not a natural owner-corruption experiment.

**Optional capture-only fixture preservation:** `installed-schema-cohort.test.ts:18–67`, exact name:

`pins the complete schema cohort closure, identical schema URIs and canonical 17-Type fixture oracles`

It verifies four added schemas, explicit `$id`/propertiesSchema/retrieval equality, the unchanged shipped Type closure, and complete manifest equality. It does not start HTTP or prove transitive `$ref` retrieval; these fixtures model Type inheritance with separately named schemas.

A future selected run must build before the tests' emitted imports and use exact test-name filters in these existing files. No broad matrix, package override, new fixture or new test framework is required just to preserve the existing path. Ordinary selected tests use their existing emitted factory/child entries; do not label them fresh installed-package evidence without a genuinely source-bound pack/install invocation.

## Remaining work and safest next step

For the supported retrieval subset, no concrete Update serving failure was established. Existing source routes every accepted captured schema entry, including transitive documents. The inspected tests are narrower than that implemented behavior. If root later wants additional cold-client confidence, the smallest evidence follow-up is to extend or separately review a bounded Update root-plus-relative-ref traversal using existing helpers, independent captured bytes and explicit identity alignment; that would be evidence work, not a known production repair, and should not displace active Transactional serving work.

The shared canonical/embedded-ID reachability residual is real, but fixing it requires a separately reviewed alias/representation mapping boundary: serving a whole captured document at an embedded resource identity is not automatically the right representation, and new startup refusals would break accepted configurations. This audit selects neither behavior. Update's pre-existing external/out-of-namespace retrieval rejection and TX's accepted-but-unserved subset differ deliberately; PR119 must not silently impose Update's restrictions on TX or loosen Update's candidate configuration.

Schema retrieval requires no pending incident-Link Event membership decision. Transactional schema minimum handling remains an explicit separate gap; Update does not acquire that machinery merely because it shares HTTP helpers. Preserve public admission gates and existing unresolved semantics. The immediate safe action is targeted preservation of the two existing tests when root schedules combined validation, with **no Update production change** proposed here.
