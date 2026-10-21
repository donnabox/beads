# Native final five-file code review

No unresolved actionable finding in the final frozen five-file draft. The one Medium finding N1 is closed by the exact local report-retention repair. This is static code review, not test, installed evidence, qualification or execution authorization.

## Exact source

Base453b6dab386746bb03aac2c0a293572b8a71c969; worktree /Users/dbox/repos/wt/vickie/vickie-command-minimum-evidence. Exactly the five authorized paths are modified. Hashes verified at final read and again at report creation:

- packages/conformance/test-support/transactional/manifest-v1.json — a48612a0935e7fe557cd3f75af281cc87522c339c3f09c0fa6aff6ccd1671446
- packages/conformance/test-support/transactional/coverage-v1.json — edfa30e2bc58b7b067a3b941da896d0ed991dcf3173467e348407904f054162a
- packages/conformance/src/transactional-cumulative.test.ts — f755f427ce59b59f43e42d3ec6ddd09a69985b98f59929a524324890d6c0af18
- packages/conformance/src/profile-contract-plan.test.ts — e5be4864e456431db74bd325ec64d785bf3c5301a6be5b230bfcc7351ade4fc4
- apps/bdptest/src/transactional-matrix.test.ts — 7bcb736e933c26d603ca40244e08eabef62f53a3894f988a2235751441796fe9

Full diff SHA256 16c0465a581a9fd3504e7667488481f94cde49be17c3a0aa9ed2ac609363246e.

## N1 closure

At app:93–104 the existing actual observation shape is named locally and an optional synchronous onObserved callback is added. At:314 it is called immediately after the real runner result is assigned, inside the original try and before finally. Both positive/control writers now serialize the same safe sourceCommit/entry/result/sessions/baseline/supplemental object from that callback. A write failure enters the existing primary-error catch; target.close, timer clearing and owned directory removal remain in finally. Cleanup errors continue to aggregate with the primary error. No report is fabricated when runCumulativeTransactionalMatrix itself never returns a result. This resolves the demonstrated loss of an already-completed report on subsequent outer cleanup failure. The report captured at that point is genuine pre-outer-cleanup evidence; later cleanup failure must still be reported separately, never presented as success.

## App review

Reviewed the complete app diff plus unchanged observe, runner, private receipt authority, schema and relevant store/Event context. The app helper is a test-only supplement; no transport/runtime/export/API seam changes.

- scenarioSelection is optional and passed to the existing runner option. Both new controls select exactly the inherited row; full route authority remains the same fifteen routes.
- The hook activates on row-specific create URL/key, checks exact ordered URL/key/header/body/write-state, schema and independent literals/absence facts before cloning a baseline. It avoids inventing a method on HttpExchangeResponse; postrun actual request IDs/methods are pinned.
- Replay comparisons are driven by the validated independent ledger. Successful comparison records are added only after full parsed JSON equality. finish pins activation,10/4/1 raw observations and3/1/0 replay pairs for positive/batch/readonly modes. A missing pair or skipped hook cannot silently pass.
- Full collection items equal the three validated original Resources; Event cardinality/identity/content/revisions and available original transaction identities are checked. The C sequence has no exposed receipt transaction baseline, so no nonexistent C receipt field is inferred.
- Positive and selected controls require no transport/harness errors, intended complete request observations, exact action/exchange inventory and no cleanup/prerequisite failure. Controls explicitly validate actual409/403 bodies/context in the hook, preserve original success manifest assertions, and require the exact failing-ID arrays with earlier exchanges passing. request-authored-body evaluates the actual materialized derivative request, correctly remaining unchanged.
- The dynamic GET bypasses the outer hook. Its actual source action/root/pointer/source-wire digest/credential/configuration/session fields are verified against safe raw-A facts; helper original-body equality is correctly attributed. Internal actual-status/request/body outcomes are checked separately from authored coverage. There is no redacted-body inference or raw-to-client epoch/view equality claim.
- New full-report output preserves genuine runner structures and separately labeled safe supplemental root/digest/IDs/pairs; bearer headers and raw body JSON are not added to the report.
- Existing observe cleanup/90s cohort/120s test/5s exchange/10s preparation-cleanup limits remain. Historical controls are retargeted explicitly to atomic-commit where needed, with mechanical +11exchange/+1child bounds; existing hook isolation and full route membership remain intact.

## Data/static review and measured fences

The earlier native-four-path-review.md remains the full audit of the unchanged JSON/mapping bytes and original cumulative code additions. Final cumulative changes from that snapshot are only the approved guard literals/comments. Exact parent reconstruction preserves all26 original objects and both complete parent files; new row11actions/192unique authored assertions/oneAroot; totals14rows/210actions/2294assertions. Coverage mirrors exactly authored IDs, keeps all wholeContract flags false, and describes app/helper supplements separately. RU130coverage/97partial/76omitted and omitted/false/empty command binding remain; Read exclusion and inherited RU ownership remain.

Independent route calculation using actual fixture.bindings gives15 unchanged grants, reordered1228-byte preimage SHA02dd7d9f35e7482c07ce1550a6572232cc48ba6858cd66f8dfc8f5384f3628be. An initial scratch calculation omitted fixture bindings and was discarded; the corrected calculation includes them. New row maximum authored body173bytes.

Canonical measurements independently recomputed738783disk/750589normal/19575nodes/108091coverage. Recounted all40 supplied derivative JSON specimens: maximum756707normal bytes/19731nodes. Root growth acceptance SHA8a6640bea4052028007ab0f5853e19bd11214a1b70e7c1b1aa6c5dc4699ad8b7 is applied exactly: canonical742879/754685/19639/109115 and both app derivative760803/19795. Original488280/13019/70662 bases,4096/64/1024 margins and historical comments remain. Product1MiB/100000node capacities and all runtime/wire limits are unchanged. This verifies specimen arithmetic, not execution of the derivative transformations.

## Boundaries

No tests, imports of product modules, builds, services, packages or actual Claude process were run by this review. Python performed read-only text/JSON/hash/count inspection only. Source/header/body/runtime outcomes remain to be measured by the separately authorized seven static callbacks and three installed cohorts, expected208/max225HTTP and26children. Existing52 plus2 new observe registrations remain distinct from that local selection; no broad suite success is claimed.

The full supplied-context Claude CODE prompt is prepared separately with exact final five-file diff and relevant unchanged source. Root owns final launch and execution release. This native review cannot substitute for actual controls, installed provenance, cleanup observations or external review.
