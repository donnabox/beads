# Independent native review: U1 and T1 served-profile preparation

## Finding

1. **Medium — T1 unintentionally exposes its source-context seam through the public pagination API.**
   - **Locations:** `packages/server/src/read-pagination.ts:44` adds `sourceContext` to `ReadFirstPageInput`; `:120–122` adds the context accessor to `ReadPagination`; `:312–315` implements it on the returned engine. Existing `packages/server/src/index.ts:176–181` exports `createReadPagination()` and returns that same object, and `:1565` exports its type.
   - **Issue and impact:** Although the new comments and implementation report call this source-private and say public exports were untouched, existing public exports transitively expose both the new input and accessor. An ordinary consumer of the package root can supply context and read it back. This creates an unplanned supported API surface and a compatibility obligation for a seam intended for private TX composition. It is **not evidence of Transactional token leakage, unauthorized cursor traversal, or public profile advertisement**.
   - **Reproduction:** In the private build of exact T1, import `createReadPagination` from the emitted public `packages/server/dist/index.js`, construct an ordinary engine, then call `firstPage({items: [], authorizationView: 'v', scopeEpoch: 'e', projection: 'p', continuationUrl: 'https://example.test/beads/', sourceContext: {source: 'reachable-from-public-export'}})`. `typeof engine.sourceContext` is `function`, and calling it with the page returns the supplied context. Actual output is preserved in `profile-service-native/t1-public-context-probe.log`.
   - **Recommended repair:** Keep one pagination implementation, but separate the existing public facade/input type from the internal context operation and page-context lookup. Internal module helpers or a private capability backed by WeakMaps can pass context into the same snapshot/capacity logic and retrieve paired metadata. The public factory should return only its previous surface and should not consume an extra `sourceContext` field. Add a public-import type/runtime regression plus retain the internal budget/expiry tests. An explicit public generic API addition would instead require an intentional scope decision and public contract/tests; the smaller repair is to keep this slice internal.
   - **Disposition:** Resolve before calling T1 a private-only API change. No source changes made by this reviewer. Root was notified and confirmed the boundary discrepancy should be treated as real.

No other actionable findings were identified in these bounded deltas. U1 has no new blocker. The single T1 finding above concerns API exposure, not the correctness of captured context, pagination snapshots, or wire output.

## Exact scope

- Baseline for **both complete deltas**: `caf5ca75a4e9b9c7865e61e97f849ab736c21e64`.
- **U1:** `3edba2140cd98878abfe36eb5bf8ab961eec708c`, `/Users/dbox/repos/wt/vickie/vickie-update-served`; all seven changed files reviewed, including preserved earlier work, shared routing extraction, private runtime composition, source-only descriptor/directory fixture, real HTTP tests, and smoke inventory.
- **T1:** `20fbf447e712c85beb32c4371dd49f2d085d6455`, `/Users/dbox/repos/wt/vickie/vickie-transaction-served`; all eleven changed files reviewed, including preserved reader-facade work, same-entry context capture, generic authority observation and retained pagination context, tests, progress note, and smoke inventory.
- Both author trees were clean and remained unchanged. Tests ran in private exact-commit archives under `profile-service-native/u1` and `profile-service-native/t1`, with private built workspace packages and reused external dependencies. Node was explicitly pinned to 24.16.0. No install or lockfile change was needed; existing pnpm-managed dependencies were reused. Build completed before emitted-runtime/process tests.

## Verified behavior and limits

### U1

The extraction preserves the prior public handler's method/origin/probe/alias ordering, static alias identity gate, branded-server admission, local-error conversion and single application of HTTP semantics. Type-only references from the helper back to the index do not create an executable module cycle. The mutable source fixture uses the real authority, receiver, fixed operation table and lifecycle. Canonical `/acme/` navigation reaches a schema-parsed service descriptor and directory; the public development facade does not accept the composition argument. Construction failure closes both listener and created owner.

Properties-view ETags use the Resource revision captured with the value. This matches the existing normative properties-view requirement at `docs/specs/bdp.md:1368–1370`; it does not fabricate a tag with a later read. The existing public static adapter's lack of a captured properties revision is inherited and explicitly documented by the author, so it is not classified as a regression or expanded into this repair.

Observed tests cover all eight singleton targets plus sequence, cold SDK navigation, actual afterstate/alias reads, retained failure and cross-carrier identity behavior, body/key/auth/raw-target/Origin/media gates, real socket disconnect and graceful drain/reopen, HEAD/304/412, bounds and stable continuation after mutations. Optional aggregate, policy/schema configuration, complete catalog qualification and public profile activation remain outside this source fixture.

### T1

`withRead` takes the persisted epoch, fixed current view and committed metadata position under the same synchronous owner guard as the expiring reader. The row facade cannot mutate, reenter or survive the callback. Fenced/closed owners refuse entry. Captured descriptors/configuration are immutable and align with mutation installation facts.

The authority snapshots generic source context via own data descriptors before calling authorization/configuration capture. Proxy/accessor fields in that context are rejected without execution. The TX interpreter validates exact own epoch/view/position fields and prevents future snapshot positions. The current serving observation remains distinct from a retained page's original rows/revisions/context; it is not inferred from a receipt or overwritten by a newer Scope position.

Context is included in retained byte/node budgets and released with snapshot capacity. WeakMap values do not reference pages/snapshots, so externally held pages can retain metadata without keeping a cursor traversal alive. Basic Read/RU observations remain unchanged when no context is supplied. Normative basis reviewed: stable selected sets/projections/revisions at `docs/specs/bdp.md:1571–1581` and strict TX prefix/context rules at `:5758–5805`. This slice does not implement or claim HTTP context fields, minimum-position barriers, receipts, replication Snapshots, changefeeds, or complete Transactional conformance.

### Integration seams

U1's new `authority.perform` remains compatible with T1's narrower Resource/configuration reader and optional observation additions; it uses the already-captured `resourceRevision`, not TX context. U1 does not edit authority-read or pagination. T1 does not edit the routing extraction or development runtime. Both smoke inventory edits are additive and must be preserved during integration. No protocol schema, conformance artifact, Read observer, evidence pin or advertised installed profile changes occur in either delta. This review does not relabel old observations as evidence for these new sources.

## Independent validation

Logs are in `/Users/dbox/repos/wt/vickie/bdp-integration-logs/profile-service-native/`.

- U1 build: PASS (`u1-build.log`, exit 0).
- U1 actual mutable discovery/HTTP plus shared HTTP semantics: **128 tests / 3 matched files PASS** (`u1-tests.log`). The invocation included a nonexistent `node-http.test.ts` selector; it contributed no tests. The actual listener file was run separately below.
- U1 public index/listener regressions: **87 tests / 2 files PASS** (`u1-public-tests.log`). Total U1 coverage here is 215 tests across five distinct files.
- T1 build: PASS (`t1-build.log`, exit 0).
- T1 transaction-read, transaction-store, authority-read, read-pagination and read-sequence-lifecycle: **376 tests / 5 files PASS** (`t1-tests.log`). Includes actual persisted rows, aliases, reentry/lifetime, uncertain-commit fencing/reopen, source-context timing, continuation replay after update/delete, expiry/budget behavior and proxy/accessor negatives.
- T1 dependency boundaries: PASS for all 11 workspace packages (`t1-boundaries.log`).
- Public-context reachability probe: reproduced finding 1 (`t1-public-context-probe.log`).

No full suite, fresh Read evidence cohort, package installation smoke, or combined U1+T1 tree was run by this reviewer. Author-reported checks remain separately attributable to their reports; they are not presented as independent verification here. No interrupted run or unverified result is counted as a pass. No outward writes or author-source edits were made.

## Native closure addendum — T1 context privacy repair

**Finding 1 is CLOSED at `1949f25cbd0e1c639a26e482f4a82763acf8aa20`, direct child of reviewed `20fbf447e712c85beb32c4371dd49f2d085d6455`. No new actionable findings.** The full eight-file repair was inspected read-only in `/Users/dbox/repos/wt/vickie/vickie-transaction-context-private`; its working tree remained clean.

`ReadFirstPageInput` no longer declares context, and public `ReadPagination` again exposes only `limits`, `validateLimit`, `firstPage`, `continuePage`, `cleanupExpired`, and `close`. The public factory returns a frozen facade with exactly those fields. Its one-argument `firstPage` forwards only the ordinary input; it neither reads a `sourceContext` getter nor forwards a supplied extra argument. Thus the earlier emitted-public-root reproduction no longer reaches context capability.

The internal factory's `firstPageWithContext(input, context)` and paired context lookup remain behind the unexported package module. Authority Read composition explicitly uses that private factory. Both paths share the existing engine: snapshot copying, context byte/node charging, cursor admission, expiry, and cleanup have not been duplicated or weakened. Package-root export maps remain unchanged, and package deep-import denial is tested. This repairs the API boundary without adding wire fields, public profiles, store/write semantics or evidence claims.

Independent exact-commit archive: `profile-service-native/t1-repair`. Explicit Node 24.16.0; private workspace builds and reused external dependencies. Build completed before the emitted package-root subprocess regression.

- PASS build (`t1-repair-build.log`, exit 0).
- PASS workspace TypeScript check (`t1-repair-types.log`, exit 0), including emitted public declaration negative assertions for no context member, no private factory export and exactly one firstPage argument.
- PASS **250 tests / 5 files** (`t1-repair-tests.log`): public pagination regression, pagination engine, transaction-read, authority-read and read-sequence-lifecycle. The emitted `@bdp/server` import test checks exact public object shape, inaccessible context methods/factory, ignored malicious getter/extra-argument proxy, unchanged tight-budget continuation, and rejected package deep imports. Internal tests preserve same-entry current context, original page context, budget/expiry accounting and cleanup behavior.

The updated installed-package smoke assertions were inspected but not independently rerun; the author's report records its own successful packed smoke. No full suite, combined integration tree, or fresh evidence cohort was run here. No author-source edits or outward actions were taken. This verdict closes the bounded T1 API issue and does not extend the earlier U1/T1 qualification scope.

### Clean-checkout type-test import follow-up

Inspected root's uncommitted three-line test-only diff in `/Users/dbox/repos/wt/vickie/vickie-profile-services-integration`: both type-query imports in `read-pagination-public.test.ts` change from generated `../dist/index.js` to source public entry `./index.js`, and the test label now accurately says public types rather than emitted declarations. No executable package-root probe, negative API assertion, getter/extra-argument case, or deep-import denial changes.

The change fixes the actual CI ordering (`pnpm typecheck` precedes `pnpm build`): source type checking no longer requires generated server declarations. Independently copied only this test delta into the private 1949f25 archive, ran `tsc -b --clean` and then `tsc -p tsconfig.check.json` with Node24.16.0; both passed with empty diagnostic logs. Confirmed `packages/server/dist/index.js` and `index.d.ts` were absent during/after that clean source check. Logs: `profile-service-native/t1-clean-import-clean.log` and `t1-clean-import-typecheck.log`.

No new finding; prior API-privacy closure stands. Type assertions now verify the public source signature, while actual emitted/package-root runtime coverage remains separately present. This bounded check does not claim a rerun of the combined integration suite or CI job.
