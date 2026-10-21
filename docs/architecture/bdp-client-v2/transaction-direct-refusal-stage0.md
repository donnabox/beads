# Direct-refusal context: read-only stage 0

Ready for final-parent/ownership release, **not authoring yet**. No additional production seam or plan change was identified. Root's accepted substance and independent plan closure remain distinct from PR116 CI/merge and implementation authorization.

## Frozen basis and reviewed authority

Inspected clean `/Users/dbox/repos/wt/vickie/vickie-start-now-resume` at `8c78cba5d2be5d5612c6a9324ee36ed690fdeb19`, tree `f2bc6fd3750bb7b453984f9927f75e64554c777a`. No source writes, tests, builds, installs or network operations. Only this report was written.

Read complete plan `transaction-direct-conflict-context-plan-v2.md` SHA256 `16917c36972aefa6601f8af602b15da2c2ce24b1ec72e01022fe7ffda37714f3`, complete dispositions SHA256 `6fc8bfb3e2324ca2a3ea09e8ffe1123ed9667f399a65ddbc372afaccc76e8d7e`, and independent v2 closure SHA256 `1d74b6a9a69d94895d71d8a8bc839df79fd636cb075dd2b17633da2e57b63a68`.

## Re-anchored production boundaries at 8c78

1. **Owner decision/capture — transaction-store.ts.** Existing admission variants are at121–138; readContext583–592 captures a frozen actual epoch/view/position tuple. The ordinary and unresolved-template conflict branches remain1065–1079. Fresh count413 and fresh406 selection remain1082–1084. `admit`1375ff calls `admitKey` inside `atomic`1387ff; retained equal406 exits at1392; ordinary result returns at1395. Capture only selected observed conflict/operation-limit/not-acceptable results, including both406 exits, before leaving this atomic callback. Keep the unobserved return object exact and do not add reads to owned admissions. `admitObserved` is exposed at1716; `comparisonEpoch`1713 separately reads epoch and is not a position-observation defect.
2. **Propagation — transaction-submission.ts.** Optional context belongs only on existing selected variants (receipt already has it; current not-acceptable at39, shared refused40–55). Observed comparison wrapper102–118 selects owner.admitObserved. Conflict and operation-limit translations120–130 currently discard everything except the refusal fields; not-acceptable is returned directly. Preserve optional tuple and require it only on the selected final observed result. Existing receipt check139ff is the pattern. The comparison module49–53 returns the owner result unchanged; service499–517 passes all non-receipts unchanged. No edits to either derived layer are needed.
3. **HTTP projection — transaction-http.ts.** Selected sink346–364 remains receipt success, bodyless406, development gap, idempotency-conflict, then shared fallback. Wrap only selected406, idempotency-conflict409, and limit-exceeded with limit transaction.operations. Do not branch on status413 alone: request-too-large shares that status but occurs before owner observation. Existing context114ff sets only the three BDP fields; empty79ff and JSON response retain existing cache behavior. POST Vary:Accept remains at589. No fallback read or new context requirement on syntax/auth/deadline refusals.

The accepted law at spec5766–5774 and admission4490–4530 still applies to these known authenticated observed cases. No mutation minimum barrier or global error/disclosure precedence is introduced.

## Exact seven-file write inventory after release

1. `packages/server/src/transaction-store.ts`
2. `packages/server/src/transaction-submission.ts`
3. `packages/server/src/transaction-http.ts`
4. `packages/server/src/transaction-store.test.ts`
5. `packages/server/src/transaction-submission.test.ts`
6. `packages/server/src/transaction-http.test.ts`
7. New `apps/bdptest/src/transactional-direct-refusal-installed.test.ts`

Dependency/alias/set/owned test files below are run **unchanged**, not additional write ownership. No service/comparison, schema/spec/catalog, conformance artifact, fixture, launcher, runtime configuration, SSE or shared authority edit. Source outside these seven requires root arbitration.

## Existing exact-shape tripwires

All names/anchors below were read from8c78, not copied as unverified39c line references. Preserve their assertion bodies.

|File and test title|Exact assertion / purpose|
|---|---|
|submission154: `compares conflict before negotiation and negotiates exact pending/terminal replay without changing it`|Lines162 and165 require exactly `{kind:"not-acceptable"}` for retained pending and terminal receipts: retained406 exit.|
|submission169: `checks fresh capacity before negotiation and never captures a refused key`|Line183ff exact413 object;193 exact `{kind:"not-acceptable"}` for fresh406, plus refused-key success/no effects.|
|submission229: `preserves failed and expired duplicate dispositions and does not renew detail`|243/256 exact retained406 for failed/expired outcomes.|
|submission302: `keeps unresolved comparison ahead of negotiation, then snapshots immutable controls`|323 exact not-acceptable after actual wait.|
|submission331: `returns a provable unresolved conflict immediately and times out possible equality without disclosing a receipt`|Exact503 at361ff remains without context; no wrapping comparison timeout.|
|submission377: `rechecks fresh negotiation after unresolved retraction without capturing a replacement key`|394 exact fresh406 after retraction.|
|dependencies435: `checks pins, repeated-creator equalities and supplied identities before direct waiting`|463/473 exact `{kind:"conflict"}`.|
|dependencies682: `uses partial durable creator facts for immediate comparison without waiting for the other creator`|707 exact conflict.|
|sequence-alias433: `rejects fixed input conflicts while alias identity is unresolved`|446 exact conflict.|
|sequence-alias698: `waits for a fixed A candidate then conflicts with the owner's actual B witness`|719/721 exact conflict.|
|set504: `counts all outer set members before new-key admission without a filtered Resource count`|513 exact operation-limit.|
|set792: `accepts 32 real no-op matches at the private default and durably refuses overflow`|832 exact conflict.|
|owned897: `observes 128/129 real Events under a lowered evaluator cap within the 32-match owner`|948 exact conflict.|
|store197: `rolls back earlier Resource and identity allocation on final operation failure, retaining only failed receipt`|215 exact conflict.|
|store231: `excludes names, object order and equivalent numeric/local spellings from identity but includes metadata and operation order`|266 exact conflict.|

Existing store3031 `captures exact replay and execute observations inside their owner transactions` provides the direct-owner SQL spy pattern; its expectation at3041 uses DatabaseSync.isTransaction. New tests must keep the spy local to direct owner admission and distinguish position reads, not assert every higher-layer metadata(epoch) lookup is inside atomic. HTTP248 `preserves conflict-before-negotiation and replay-before-capacity without binding refused keys` is the receiving-order regression to extend beside new current-position assertions.

## Source-control selectors to establish while authoring

These four controls are already accepted plan scope. Newly proposed exact test titles below do **not** exist yet; record final titles/assertion locations before execution and verify nonzero selection. Use isolated disposable derivatives; source controls never become altered positive responses or new installed cohorts.

- `missing-http-context`: omit wrapping at selected HTTP sink only. Select proposed HTTP test `carries current owner context on direct conflict, operation-limit and both negotiation refusals`; require named independent BDP header-presence failures, not setup or status failures.
- `stale-retained-position`: replace the selected tuple position with pos-1 in the derivative after the positive fixture has committed unrelated pos-2. Same HTTP test must fail explicit current-pos-2 equality after presence passes. The retained receipt still has its real old execution facts; no historical evidence relabel.
- `capture-outside-owner-transaction`: move only new refusal capture after atomic in derivative. Select proposed owner test `captures selected observed refusals inside the owner transaction without changing unobserved admissions`; require the position-read isTransaction assertion to fail. A higher-layer epoch lookup is not an acceptable failure.
- `unobserved-context-leak`: remove the observed gate on selected refusals, including the retained406 early branch and fresh406 branch. Select the unchanged submission154/169 tests above, plus dependencies435/set504. Both fresh and retained406 exact-shape failures are required. A mutant touching only one406 branch is insufficient proof for the other. Record actual failing tests/diffs; do not accept compile failure as discrimination.

The new direct-owner test also pins all three tuple values/frozenness, selected-position-read counts, unchanged owned/unobserved paths, and previous tuple immutability after a later commit. Higher-level observed=true tests retain conflict-before-negotiation/capacity order and test an unrelated actual commit during comparison wait; the final observed decision, not a previous attempt, supplies context.

## Proposed explicit validation commands — not executed

Run only after root releases a separate implementation checkout/final parent. Every shell operation must specify that checkout as cwd. Log commands/exits/elapsed/cleanup externally under the assigned implementation report directory. No command below runs `pnpm test`, the whole workspace, or unfiltered `transactional-matrix.test.ts`.

Set pinned toolchain in the bounded command environment:

```sh
export PATH=/Users/dbox/.nvm/versions/node/v24.16.0/bin:$PATH
PNPM=/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs
/Users/dbox/.nvm/versions/node/v24.16.0/bin/node --version
node "$PNPM" --version
```

Read .nvmrc/.node-version/package.json pins before build; inspected8c78 pins are24.16.0 and11.20.0. Then, in the assigned cwd with observable supervisor/external logs:

```sh
node "$PNPM" install --frozen-lockfile
node "$PNPM" typecheck
node "$PNPM" build
node "$PNPM" exec vitest run packages/server/src/transaction-store.test.ts packages/server/src/transaction-submission.test.ts packages/server/src/transaction-http.test.ts packages/server/src/transaction-dependencies.test.ts packages/server/src/transaction-sequence-alias.test.ts packages/server/src/transaction-set.test.ts packages/server/src/transaction-owned.test.ts --maxWorkers=2 --reporter=verbose
node "$PNPM" format:check
node "$PNPM" lint
node "$PNPM" check:boundaries
node "$PNPM" evidence:verify
git diff --check
```

A derivative's bounded selector can use the following template, replacing FILE/TITLE with the actual named test above and recording exact count. Do not invoke it before the corresponding test exists:

```sh
node "$PNPM" exec vitest run "$FILE" -t "$TITLE" --maxWorkers=1 --reporter=verbose
```

Unobserved leak selector over the existing submission tests (regex is a union of the two exact name prefixes):

```sh
node "$PNPM" exec vitest run packages/server/src/transaction-submission.test.ts -t 'compares conflict before negotiation and negotiates exact pending/terminal replay without changing it|checks fresh capacity before negotiation and never captures a refused key' --maxWorkers=1 --reporter=verbose
```

For packed validation, build/pack/install exact frozen source using existing repository package workflow; capture actual tar hashes and installed entry before the command. Select **only** the new dedicated test file, whose one test must implement the accepted14-request ledger. Suggested title `observes current context on installed direct refusals within14requests`; this is a proposed title pending authoring, not current evidence.

```sh
BDP_TRANSACTIONAL_INSTALLED_ENTRY="$ACTUAL_INSTALLED_ENTRY" BDP_TRANSACTIONAL_SOURCE_COMMIT="$FROZEN_SOURCE_SHA" node "$PNPM" exec vitest run apps/bdptest/src/transactional-direct-refusal-installed.test.ts -t 'observes current context on installed direct refusals within14requests' --maxWorkers=1 --reporter=verbose
```

Run that same selected authored proof separately against the preserved original frozen payload with its actual original source SHA, not the repaired SHA. Require request3 missing-context failure after both seed assertions pass; keep the actual payload/source identities distinct. These environment fields follow existing installed conventions; the new test must explicitly validate override pairing. Never replace the default entire matrix with a narrower report and call it full qualification.

## Execution gate still pending

Read-only stage0 does not produce measured request bodies or author source footprint. During authoring, before any installed execution, publish the actual14-request ledger/routes/body sizes, largest33-operation carrier byte count under65536, exact seven-file diff and new test names. Default maximumOperations is32 at transaction-batch.ts63, confirmed unchanged. Positive is14requests/one child; original negative expected3/one child with conservative aggregate28/two, unchanged5s/90s/120s exchange/cohort/outer bounds. Four source controls add no installed cohort.

Freeze clean source before pack. Retain prior/source-specific evidence, original negative and exact result/cleanup artifacts. PR116 CI/merge and final ownership release are explicitly unresolved here; no readiness or network verification was attempted. No source authoring was started.
