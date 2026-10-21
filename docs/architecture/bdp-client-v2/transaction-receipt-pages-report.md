# Frozen durable receipt paging slice

Commit `0ff180a531b3c9ae594e0a104321e80744de38ed`, direct child of frozen
`d3252087a31529cf97b5bfcfe2c402f02c694b91`.

Checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-receipt-pages`;
branch `codex/vickie-transaction-receipt-pages`; origin
`https://github.com/gastownhall/bdp.git`. Clean after commit. Local only, no
outward writes. All predecessor/review checkouts preserved.

## Result

Existing explicit mutations now produce durable receipt pages through the private
same-owner HTTP service. Five entries at maximum 2 yield 2/2/1, while maximum 5 keeps
all five inline. The default remains 32. One immutable validated maximum supplies
store partitioning, service page.maximumItems and Read pagination.maxPageItems;
defaultItems is min(2, maximum). No separate smaller bound hides under maximum 32.

Full results remain in the receipt row. Raw execute/admit/receipt and sequence
internals continue using those complete facts. Only observed wire projections
return a prefix. The owner creates a parent plan row (receipt/principal/key/count/
maximum) and opaque page range rows (token PRIMARY KEY; UNIQUE(receipt,start)) in
the terminal SQLite transaction alongside effects/group/receipt. No stored next
link duplicates ordering; next URLs derive from recorded contiguous boundaries.
No page stores a second postimage copy. Retention compacts all creation allocations,
including tail pages, while retaining compact issued page identity for the epoch.

Root/page projection validates recorded parent, count, maximum, nonempty contiguous
ranges, no overlaps, complete coverage, token uniqueness and current integrity.
Missing/inconsistent chains cause a service fault, never re-partitioning, silent
truncation, oversized inline fallback or false expiry. Only the fixed all-visible
projection is supported, as before. Internal complete facts are never replaced by
a prefix or re-read from current Resources.

Page lookup establishes authenticated principal/epoch-bound identity, then supported
minimum context/wait, then expiry and representation in the same owner entry. An
unsupported tuple yields the explicit private gap after nondisclosure but before
expiry. Known expired pages return 410; hidden/unknown/prior-epoch pages remain 404.
Root/page current serving context is independent of immutable execution facts.
Normal GET/HEAD/conditional behavior and validator omission remain in the existing
HTTP projection. No SQLite transaction survives an await.

Private format is now `bdp-private-transaction-receipt-pages-6`. Read-only
installation compatibility preflight runs before writable owner opening; the
locked owner checks again before recovery. Prior format/mismatched bound refuses
byte-identically with pending and expiry-eligible records present. Compatible
reopen retracts pending work and performs normal expiry. No migration or reseed.

## Files

- transaction-store.ts: durable page metadata/ranges, projection, same-entry page
  reader, read-only compatibility preflight, retained-root minimum ordering.
- transaction-service.ts: exact bound compatibility, bounded root/page invariants,
  tracked page reader and finite minimum wait; preserves raw/sync interfaces.
- transaction-http.ts: page dispatch, unsupported control ordering and parser 400
  preservation. Page detection is relative to the receipt namespace, so a canonical
  Scope containing `/pages/` does not misroute ordinary root receipts.
- transaction-runtime.ts: one detached private test-seam page maximum fed to all
  consumers before owner open. No public option or profile flag.
- transaction-consistency.ts: reserved genesis minimum maps to initial position.
- transaction-receipt-pages.test.ts (new), transaction-http.test.ts,
  transaction-service.test.ts, transaction-runtime.test.ts: acceptance evidence.
- smoke-executables.mjs: additional root-negative export assertion for the private
  bound helper; no new emitted runtime module or export-map change.
- transaction-implementation-progress.md: bounded paging status, corrected HTTP
  disconnect disclosure and polling qualification.

## Paging plan review dispositions

F1: exact persisted/advertised/Read maximum equality, default 32; new fixtures use
maxima 1/2/5 and valid defaults. Maximum 5 inlines five entries; maximum 2 pages them.
The review's suggestion to derive receipt size from defaultItems is not adopted:
root/native arbitration follows the explicit maximum-bound inline rule. No new
normative ruling or unadvertised smaller partition value.

F2: page reader has an explicit private unsupported-control result after identity
nondisclosure but before expiry. Actual expired-page tests distinguish authorized
500 gap, foreign 404, complete-tuple timeout 503 and eventual expiry 410.

F3: replaced next:null-only service guard with bounded nonempty prefix and same-root
navigation checks. Store validates next against recorded full coverage. Page
counterpart checks nonempty length, exact maximum and parent/next relationship.

F4: page tokens are primary keys; receipt/start is unique. Parent metadata retains
principal/key for direct parent-row access. There is no redundant next-token field.
Current bounded reader validates the whole recorded chain; this is not a large-set
paging performance claim.

F5: default 32 preserves prior fixtures. Raw full-fact methods/sequence/creatorFact/
expiry remain distinct from observed wire-prefix methods. Full TX regressions pass.

F6: format and maximum both participate in the installation fingerprint. Refusal
has the existing installation-mismatch diagnostic; distinct input fixtures prove
both branches byte-identically. No new public error or migration behavior. The
recorded boundaries ensure identity stability; the fingerprint deliberately
rejects changed-bound installations even though an eventual migration could exist.

F7: lower caps apply only to newly configured private fixtures and their Read query
pages. Existing default remains 32, no shared router change or Read evidence reseal.
Removing the old >=32 guard is correctness cleanup now that pages actually exist.

Smaller items: actual process witness uses configured stable Scope and different
listener addresses (including a Scope path containing pages); existing after-terminal
fault stage tests page/effect rollback without a new stage; uncertain COMMIT uses
the existing DatabaseSync injection seam. Intermediate `/receipts/id/pages` is 404.
Page path syntax is implementation-defined navigation; illustrative query fixtures
are not normative URL spelling. Spec support includes 5160–5166 set expansion,
5167–5187 paging and 5189–5244 projection/current-context clauses. expireRow's full
result/next:null invariant is preserved, not loosened.

## d325 code review dispositions

F1: minimum now precedes read-triggered expiry/projection for a currently stored,
disclosed receipt. A failed receipt that becomes disposal-eligible during a wait
can still produce actual 503 when the minimum never arrives; real HTTP and service
tests cover this. If independent maintenance actually forgets it, the next attempt
returns 404 before minimum. No receipt pin or invented retention extension.

F2: a concrete parser 400 remains 400 under valid minimum controls. Unsupported
executable collection/alias/discovery paths retain the private gap; absent-point
ordering is not changed. HTTP tests cover unknown query and invalid view syntax.

F3: documentation now explicitly distinguishes the private facet AbortSignal from
actual HTTP: neither point nor receipt/page disconnects propagate an abort signal.
They drain finite remaining waits, which can delay close. No new cancellation system.

F4: reserved genesis is recognized as initial position; real HTTP test succeeds.
Unknown opaque positions remain a separate unresolved gap.

F5: bounded 5 ms polling remains, including repeated materialization/SQLite work and
no production concurrency bound. Documented as a private loopback efficiency limit,
not cured by introducing a notification/concurrency subsystem.

## Verification

Pinned Node 24.16.0 / pnpm 11.20.0. Frozen install passed. Build preceded emitted and
real-process tests. Logs are alongside this report.

- PASS final 519 tests / 14 files, no skips: `regressions-accepted.log`.
  Includes all Transactional suites, authority Read, lifecycle and pagination.
- PASS build: `build-verified.log` (final implementation; later edit was test-only).
- PASS typecheck/format: `typecheck-accepted.log`, `format-accepted.log`.
- PASS boundaries: `boundaries.log`, 11 workspace packages.
- PASS lint: `lint.log`; three existing warnings and one info in untouched files
  (numeric-model, read-update-transport, installed-schema-evaluator,
  read-sequence-lifecycle). No new changed-file diagnostic.
- PASS final installed smoke: `smoke-accepted.log`, private export denial preserved.
- PASS development smoke: `smoke-development.log`, claimEligible:false.
- PASS unchanged 89-row Read evidence: `evidence.log`; no relabel or reseal.
- PASS git diff --check and clean after commit.

Initial typecheck identified two unused imports after removing the inline-only
restriction; corrected. Earlier final-matrix attempts reached 518 passes / 1 failure due
to a new test choosing an alias path colliding with the existing Resource identity,
then using an invalid empty patch for its no-op. Corrected test uses a distinct
alias and a valid semantic no-op replacement. No implementation weakened. Final 519
pass includes those cases. Full workspace suite was not run by this worker; root
owns broader integration. Injected SQL failure and abrupt process-exit witnesses
are not claims of real disk or power-loss testing. No runs were terminated.

## Remaining limits

This is explicit-mutation receipt paging, not executable large-set expansion.
Schemas/catalog/evidence/public profile admission are unchanged. Fixed all-visible,
schema-free/unowned configuration remains explicit; installed contracts, variable
views/erasure and Types are not expanded. Set operations, full consistency surface,
Snapshot/checkpoint/replay, Event/changefeed serving and truthful cold Transactional
discovery remain open. All unresolved absent-point, query-continuation and unknown
minimum-form questions remain outside this slice. No shared database or external
asset was opened; all owner/process witnesses used disposable directories.
