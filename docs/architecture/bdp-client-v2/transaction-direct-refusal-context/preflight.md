# Direct-refusal context — source-only preflight

Draft complete within the accepted seven-file scope; STOPPED before execution pending root's measured gate. No tests, typecheck, build, dependency installation, service, pack or installed run occurred. This is unexecuted source, not passing evidence.

## Source and scope

- Checkout `/Users/dbox/repos/wt/vickie/vickie-direct-refusal-context`, branch `codex/vickie-direct-refusal-context`.
- HEAD/base `8c78cba5d2be5d5612c6a9324ee36ed690fdeb19`, base tree `f2bc6fd3750bb7b453984f9927f75e64554c777a`.
- Six tracked modified files and one new untracked installed test; no commit yet. Existing frozen SSE checkout was not edited. No network, pushes, PR/main/bus/launch actions.
- Draft is captured by [source.diff](source.diff), SHA256 `6658fc0c4d14828c52b609560f2f9d0758dd9247fb9cba75caf164010c73b399`.
- [preflight-bounds.json](preflight-bounds.json), SHA256 `09592556d54b6cac6455e831959f4fed8f540922229f0ba88e1c39bebd0ef4c6`, records every file's old/new byte count, SHA, lines, exact request JSON bytes/hash, routes and four unexecuted source-control patches.

Accepted plan16917c36, dispositions6fc8bfb3, native closure1d74b6a9 and root draft release remain the authority. Root reported PR116 attempt1 full tests passed but the whole-job deadline cancelled later gates; an unchanged retry was authorized. This report does not verify CI or claim main has merged. Final ordinary ancestry/publication remains root-owned.

## Authored source

|File|Added/removed lines|Purpose|
|---|---:|---|
|server/src/transaction-store.ts|15/4|Optional context on existing selected variants; capture in the existing atomic admission callback only for observed conflict, operation-limit and both fresh/retained406 paths.|
|server/src/transaction-submission.ts|17/2|Preserve captured tuple through existing translations; selected observed refusal guard. Other refused variants unchanged.|
|server/src/transaction-http.ts|24/9|Attach tuple to exact conflict409, transaction.operations413 and empty406; existing status/body/cache/Vary remain. No general status413/context fallback.|
|server/src/transaction-store.test.ts|144/0|Observed tuple, freeze/same-entry SQL, ordinary/template conflicts, fresh/retained/pending406, no extra owned/unobserved position read, stable old tuple and refused-key reuse.|
|server/src/transaction-submission.test.ts|137/0|Propagation/order/reuse, final context after actual dependency wait plus unrelated commit, observed immediate provable dependency conflict.|
|server/src/transaction-http.test.ts|78/0|Actual HTTP contexts atpos2 versus old receiptpos1, both carriers/406 paths/413, retained receipt/current graph/Events and key reuse.|
|apps/bdptest/src/transactional-direct-refusal-installed.test.ts|463/0|One dedicated real installed14-request observation with explicit closed routes, actual wire/body hashes, config/payload binding, safe result summaries and child cleanup.|

Totals878 added/15 removed lines. Exact old contents of all three modified existing test files remain byte-identical prefixes; tests were appended without relaxing a single old assertion. Unchanged dependency, sequence-alias, set and owned test files remain named regression gates. No additional source/module/config/schema/fixture/catalog/DSL scope was needed.

The installed scaffold follows existing raw-executor/launch/binding conventions without exporting a helper or modifying shared instrumentation. It does not run the cumulative matrix. Its source-size cost is17741 bytes; generalization into a framework is outside this repair.

## Exact installed ledger and limits

Bodies are ASCII JSON emitted by the literal source builders; `measure-draft.py` independently calculates the same insertion order/compact JSON bytes without importing or executing target/test modules. This is static arithmetic, not observed wire. All14 request bodies and SHA256 values are in the machine ledger.

|#|Label|Route|Body bytes|Expected response|
|---:|---|---|---:|---|
|1|seed-a|POST operations/create-bead|94|completed,pos1|
|2|seed-b|POST operations/batch|136|completed,pos2|
|3|singleton-conflict|POST operations/create-bead|106|409,currentpos2|
|4|batch-conflict|POST operations/batch|148|409 despite unacceptable Accept,currentpos2|
|5|retained-negotiation|POST operations/create-bead|94|empty406,currentpos2|
|6|fresh-negotiation|POST operations/create-bead|94|empty406,currentpos2|
|7|fresh-capacity|POST operations/batch|4616|33 valid creates,413/transaction.operations,currentpos2|
|8|retained-root|GET exact observed seed receipt URL|0|same complete seed receipt,currentpos2|
|9|unchanged-beads|GET beads/?limit=32|0|exact a,b,next:null|
|10|unchanged-events|GET events/|0|two exact creation facts,next:null|
|11|retry-negotiation-key|POST operations/create-bead|94|same formerly refused key,completed c,pos3|
|12|retry-capacity-key|POST operations/batch|136|same formerly refused key,completed d,pos4|
|13|final-beads|GET beads/?limit=32|0|exact a,b,c,d,next:null|
|14|final-events|GET events/|0|four exact creation facts,next:null|

Largest request4616 bytes, aggregate authored bodies5518 bytes. Existing request-body ceiling65536 unchanged. Four literal method/URL pairs plus one dynamic GET bound to the actual schema-validated seed receipt id; exact observed URL set must equal these five routes. The static route preimage/digest is in the ledger and is explicitly not a substituted installed configuration identity. The actual installed binding will separately supply payload/contracts/policy/auth/runtime configuration/source identities.

Conservative expected response estimates: <=8KiB per small receipt, <=4KiB Problem, empty406 exactly0, <=16KiB four-Resource collection, <=32KiB four-Event page. These estimates are not proven universal protocol bounds or actual observations. Raw executor response cap131072B, request headers16384B, request body65536B, per exchange5s, cohort90s and outer120s remain unchanged. The static input caps leave substantial request headroom; actual response bytes are recorded before assertion evaluation during execution.

One positive14requests/one child; identical new proof against separately identified original8c payload expected to stop at request3 after the two seeds pass. Conservative original ceiling14/one, total28/two. No test-only controls enter installed configuration. All four source controls are separate unexecuted patch artifacts, not additional installed cohorts. The original target must be attributed to its actual source/payload, never the repaired draft.

## New exact selectors and source controls

Authored new tests:

- owner: `captures selected observed refusals inside the owner transaction without changing unobserved admissions`
- submission: `propagates observed direct refusal context without changing admission precedence or key reuse`
- submission: `uses the final observed refusal entry after a dependency wait and an unrelated commit`
- submission: `keeps observed provable dependency conflicts ahead of negotiation`
- HTTP: `carries current owner context on direct conflict, operation-limit and both negotiation refusals`
- installed: `observes current context on installed direct refusals within14requests`

Four patch files under `controls/<name>/source.patch` record base/mutant/patch hashes in the ledger. They have **not** been applied to source or executed:

1. `missing-http-context`: remove only the selected conflict's wrapping. HTTP selector must fail `direct refusal context presence: bdp-scope-epoch` (or the first independently named required header), with status409 intact.
2. `stale-retained-position`: retain conflict headers but substitute pos1. Same selector must fail `direct refusal current context: bdp-scope-position` after presence passes.
3. `capture-outside-owner-transaction`: move selected capture after atomic, preserving observed gating. Owner selector must fail `refusal position read must remain inside owner transaction`, not an unrelated comparisonEpoch read.
4. `unobserved-context-leak`: remove observed gates at both retained406 and the fresh/refusal selection. The unchanged exact-shape submission tests `compares conflict before negotiation and negotiates exact pending/terminal replay without changing it` and `checks fresh capacity before negotiation and never captures a refused key` must fail on added context. These are the retained and fresh406 discriminators; dependency/set exact-shape selectors remain additional regression checks.

Controls are source counterfactuals only. A compile error, unrelated assertion failure, missing selection or setup error will not count as sensitivity. Isolated derivative builds must match any emitted HTTP code; source test edits never instrument a reported positive package.

## Proposed commands after root's measured release

All commands require explicit implementation cwd and observable external logs; no command below has run. Pinned PATH `/Users/dbox/.nvm/versions/node/v24.16.0/bin`, PNPM `/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs`. Record actual versions before execution, install frozen, then:

```sh
node "$PNPM" typecheck
node "$PNPM" build
node "$PNPM" exec vitest run packages/server/src/transaction-store.test.ts packages/server/src/transaction-submission.test.ts packages/server/src/transaction-http.test.ts packages/server/src/transaction-dependencies.test.ts packages/server/src/transaction-sequence-alias.test.ts packages/server/src/transaction-set.test.ts packages/server/src/transaction-owned.test.ts --maxWorkers=2 --reporter=verbose
node "$PNPM" format:check
node "$PNPM" lint
node "$PNPM" check:boundaries
node "$PNPM" evidence:verify
```

Each source derivative selects only its specified owner/HTTP/submission test(s), maxWorkers1; build separately when the selector imports emitted runtime. Preserve derivative source and its exact failure log, then return to unchanged positive source. Freeze a clean signed ordinary commit before any packed evidence. Pack/install the real packages with source/tar hashes using the existing workflow; then select only:

```sh
BDP_TRANSACTIONAL_INSTALLED_ENTRY="$ACTUAL_ENTRY" BDP_TRANSACTIONAL_SOURCE_COMMIT="$ACTUAL_SOURCE" node "$PNPM" exec vitest run apps/bdptest/src/transactional-direct-refusal-installed.test.ts -t 'observes current context on installed direct refusals within14requests' --maxWorkers=1 --reporter=verbose
```

Repeat that selected proof once against the original payload with original source identity as the bounded negative. **No unfiltered transactional-matrix, whole-workspace test, extra installed regression cohort or deadline increase is proposed.** Root reconciles actual main via ordinary ancestry independently of this draft.

## Preparation outcomes and remaining gate

Allowed source formatting completed: Biome format --write on exactly seven paths (four formatted), then the additional owner-test edit formatted cleanly. Text/JSON arithmetic and git diff --check completed. No test result is claimed. An initial source-read shell command mistakenly used zsh's special `path` variable and consequently could not locate ls/head/sed in that one shell; a fresh read-only command succeeded. No persistent environment setting or source was affected, and no execution attempt is hidden by this preparation correction.

During static review, corrected draft-only camelCase HTTP target names, removed guessed Event payload/position fields in favor of actual Event.data and transaction linkage, and used readConfiguration's real context fields instead of an unsupported inspect().receipts field. None of those unexecuted intermediate drafts was labeled passing or packed.

Root preview requested cleanup hardening within the new test. Directory-removal errors now join the failure list rather than replacing the original assertion/transport error; directory absence is separately observed with existsSync and logged even after a removal error. Entry hash is captured before the child starts. The emitted report explicitly says it retains response fingerprints and context only, not complete wire/body bytes; actual in-memory responses still drive schema/content assertions. Future execution reports must not describe these fingerprints as retained raw wire or full projection artifacts. This limitation is deliberate and visible, not a changed oracle or extra request.

No owned test/build/service/caffeinate process was launched. Seven-file draft is held unchanged at this source fingerprint until the measured execution gate. Principal typing, native SQLite spy compatibility, emitted HTTP assertions and installed observation are intentionally unverified until authorized checks; no product/scope obstacle has been found by static inspection.
