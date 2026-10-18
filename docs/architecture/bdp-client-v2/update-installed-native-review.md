# Native implementation review: P1a + U2 repairs

Frozen reviewed source `4078ca709b55d363c147a5bdc0d5141ede9ec247`, base `2c5fe9e04250ae036e681f46c60d3592e5f1cdf1`. Author `/Users/dbox/repos/wt/vickie/vickie-update-installed` unchanged. Independent build/tests/probes in detached `/Users/dbox/repos/wt/vickie/bdp-update-installed-native-review`. Reviewed author update-installed/review.md, original U2 council and completed P1a plan council.

## Remaining actionable findings

### P2 — no whole-startup serving gate (P1a H3 partially repaired)

`development-runtime.ts:300–313` checks closing/authority and per-principal cache only. Startup fills principalViews/composedReads incrementally at522–542. Once the first principal is registered, it can execute requests while a later principal startup discover remains unresolved. There is no flag set only after the complete loop. The cache fix prevents the old uninitialized-principal legacy fallback, so do not repeat that exact obsolete symptom.

Independent controlled probe held the second principal's discover unresolved while using the real listener/store/first principal. GET returned200 and POST create-bead returned200 with a real canonical Resource/revision. Releasing the second check then rejected startup. Source `update-installed-native-startup-repro.test.ts`, log `update-installed-native-startup-repro.log` alongside this report; run source from apps/bdptest/src in isolated checkout. The sole injected seam wraps the actual authority's second-principal discover; effects/receiver/store are real.

Scope of evidence: current compiled discovery resolves via synchronous store work/promises, so this is a deterministic initialization-invariant failure under controlled delayed validation, not proof that ordinary unmodified startup offers a schedulable remote exploit window. The explicitly requested fail-closed startup invariant is absent nevertheless. Add a serving=false gate before authentication/routing, flip only after all principal/route startup checks succeed, and test GET/mutation503 with no key/effects until full success. A later failure must leave no early mutation.

### P2 — private process entry is not import-inert and has no entrypoint check (P1a M4 open)

`installed-reference-process.ts:7–8,61–66` invokes main unconditionally; it checks process.send/argv length but never whether this module is the entrypoint. Independent bare import without IPC prints `installed reference qualification failed`, sets process.exitCode2 and leaves the importing process to exit2. Under an IPC parent, importing from another entry script can instead register configuration/signal/lifetime handlers and launch the service. Package export denial does not prevent file-path import of emitted bytes.

Repro log `update-installed-native-import.log/.exit`: Node24 command `node --input-type=module -e 'await import("./packages/server/dist/installed-reference-process.js"); console.log("import finished", process.exitCode, process.listenerCount("SIGTERM"));'` from isolated clone reports `import finished 2 0`, exits2. Require exact process-entry detection before main; bare import must leave signals/listeners/exitCode/stdio untouched. Direct execution without inherited IPC should still refuse. Add separate import-inert and IPC-capable non-entry import tests. This is a private module lifecycle defect, not a public profile admission bypass.

### P2 — actual owner limit tuple is still only partly bound (P1a M2 partially repaired)

`installed-reference-configuration.ts:158–169` hashes DEVELOPMENT_LIMITS, DEVELOPMENT_READ_LIMITS and retention day. `development-authority.ts:186,191–194,235–246` still supplies inline maintenance interval, patch operation/path limits, representation bound, and pagination retention/capacity/token fields; even page/cursor settings are duplicated rather than read from the captured object. Thus the stated configuration identity does not include the complete actual fixed tuple and the owner does not structurally consume that tuple.

Use one frozen private owner-limit snapshot for these values, consumed by owner/pagination and configuration hash. Preserve existing numbers and optional advertisement semantics; no configurability framework needed. Tests should pin that changing a supported captured tuple would change config identity while defaults remain unchanged. Payload hash currently changes with source bytes, so this does not by itself collapse the full payload+configuration pair; it is a completeness/drift defect against the explicit configuration-binding acceptance requirement.

## Existing finding dispositions

| Finding | Native result |
|---|---|
| U2 A1 diagnostics | Closed: confirmed invalid prefix is emitted until receiver stops, yielding incomplete/truncated rejection; compiler truncation before receiver stop or non-evaluated work exhaustion still throws. Actual1025-allOf422/replay and4000-uniqueItems500/no-key tests pass. No synthetic validation or evaluator/store change. |
| A2 positive aggregate bound | Closed before store provisioning; max0 refused and corrected create tested. |
| A3 canonical aggregate identity | Closed for exact duplicates/reordering via sorted canonical set; distinct redundant numerical maxima intentionally not coalesced. |
| A4 pre-provision trusted configuration | Brands, aggregates and credential policy membership checked before listener/store; actual discover success checked. Whole-startup serving gate remains separately open above. |
| A5 compiler bounds identity | Closed: artifact receipt limits, graph and evaluator ceilings included in contract identity. |
| A6 global/per-principal policy | Frozen captured config and documented distinction; acceptable. |
| A7 deny-list Read limitation | Explicit retained fixture limitation; no whole-Scope read-denial claim. |
| A8 canonical Resource paths | Closed by protocol canonical path validation; actual hidden/read-only HTTP tests pass. |
| A9 implicit configured principal | Closed: configured authority exposes explicit facets only; default unconfigured facet retained. |
| A10 allocation/membership | Cached principal facets/routes and private membership sets; aliases of one principal share cache. No remaining leak claim. |
| P1a H1 cleanup | Closed: startup allSettled preserves original plus cleanup causes; normal close also attempts owner cleanup after listener failure. Injected actual startup cleanup/reopen test passed. |
| P1a H2 transport identity | Closed: validated target path/query rebased onto configured canonical origin; Host syntax checked but installed transport does not bind identity to Host/listener. Canonical Scope survives changed listener. |
| P1a M1 Origin split | Closed: explicit private installed transport branch accepts Origin without grants; old development403 behavior unchanged and tests pass. |
| P1a M3 deployment shape | Closed: transport is separate sixth parameter; existing strict deployment data shape preserved. |
| P1a L1/L2 routing/gates | Extraction uses normalized URL for probe/descriptor/directory/operations and per-request discover gate for directory/schema paths. Type/schema routes remain current-owner-backed. |

Reference Type identity remapping covers the full shipped closure, including ownsOutgoing keys/endpoint/parents; static schema hosting is a captured local manifest. No network installation, Type/CLI product extension, public export/bin/activation flag, public Read brand weakening, or evidence reseal. IPC readiness remains explicitly claimEligible:false and not a future admission grant. Installed payload bytes are observed separately from actual configuration; cumulative runner/evidence/public activation remain pending.

## Independent validation and limits

Pinned Node24.16.0/pnpm11.20.0. Isolated build PASS. Targeted14-file suite491 passed in8.93 seconds, maxWorkers2; structured `update-installed-native-focused.json` plus verbose/exit companions. Includes both new suites, development HTTP/discovery/disconnect/matrix, public Read composition/HTTP/listener, read lifecycle, recovery store, compiler and artifact ingestion.

Separate startup probe1 test passed by asserting faulty observed200/200 acceptance; it is a defect reproducer, not desired-behavior qualification. Bare-import process probe exited2, confirming the import side effect. External probe source/logs preserved; scratch source removed. No full duplicate suite or new packaged smoke was rerun beyond installed process tests; author's packed smoke remains author evidence. No source/outward edits or claim of full write-profile eligibility. Both checkouts clean at completion.

Verdict: hold P1a council closure for the three bounded refinements above. U2 fixes need not be reworked; no new operator product decision is needed.


## Successor closure at dea6f9e

All three native findings H3/M4/M2 are closed at frozen `dea6f9e7a95a3e3eba9c667de39360346f25bdf4`. Independent isolated build and15 tests/4 files passed, including the original controlled startup reproduction (now503/503), original bare import (nowexit0/no state change), IPC import inertness, actual entry lifecycle, and owner-limit identity checks. H3 remains a controlled initialization invariant; M2 is identity coverage, not an asserted behavioral exploit. Full closure and exact evidence: `update-installed-repair-native-review.md`. No public/full-profile qualification or outward action.
