# Frozen transaction recovery repair

Commit **270a1216cb918a4194f3b195db6c8fcbecf0c040**, direct child of root **7a768c160eecd6c698c934a47492a5943359c50d**. Checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-recovery-repair`, branch `codex/vickie-transaction-recovery-repair`, origin `https://github.com/gastownhall/bdp.git`. Clean and frozen; no further source edits planned. Local signed-off commit only, no outward writes.

I authored this repair and am not its independent reviewer. Root/external review is required. The prior native paging review did not exercise a hot rollback journal; its postcommit and injected-failure cases were insufficient for this crash shape.

## Changes and external F1–F5 dispositions

**F1 High — repaired.** Removed the redundant read-only preflight entirely. SQLite can perform ordinary rollback-journal recovery on the writable connection before the owner reads installation metadata. The existing locked, transactional identity-policy and configuration checks remain before pending retraction, creator-fact cleanup and retention sweeping. No catch-all fallback, sidecar decoding, migration, reseed, format bump or policy change.

The compatibility guarantee is precise: a normal incompatible-format/bound file is still refused byte-identically. An incompatible file with a hot journal may first be restored by SQLite to its committed state, then refused before any BDP recovery/maintenance. Byte identity of the interrupted physical database image is not promised in that case. The progress document now makes this distinction.

**F2 Low–Medium — repaired.** Root reads now share a disclosure-only lookup before any read-triggered retention. An unsupported minimum returns a typed internal `unsupported` result after principal/epoch nondisclosure but before expiry or representation. The existing supported and no-control `readReceipt` signatures remain available through overloads; the HTTP path consumes the typed result. No new public code or retention pin. Unknown/foreign roots remain uniform 404; supported minimum handling and independent maintenance retain their behavior.

**F3 Low — redundant path removed.** Open errors now use the existing writable owner's metadata/installation/fencing path; there is no separate `String(undefined)` preflight or duplicated metadata parser. This does not invent broader error normalization for corrupt SQLite files.

**F4 Low — repaired.** Removed the two tautological request/parent recomposition checks. The store now independently requires the configured Scope's receipt namespace; stored principal/epoch/parent/token and whole-chain validation remain authoritative. Service checks retain nonempty/max-sized slices and next-parent prefix/length checks. Corrected the comment to say the selected page entries are frozen locally rather than claiming every retained fact was frozen by prefix projection.

**F5 Low — evidence added.** New hot-journal, pagination-max mismatch, concurrent page-wait/drain and context-withholding controls described below. No new cancellation plumbing; actual HTTP disconnect still drains the finite remaining budget.

## Red/green evidence

Pinned Node 24.16.0 / pnpm 11.20.0; build completed before emitted/runtime tests.

- `hot-journal-red.log/.exit`: two expected failures against the parent implementation, both SQLite `ERR_SQLITE_ERROR`, errcode 776, “attempt to write a readonly database”. Each child commits 60 three-result receipts (180 Resources), then receives real SIGKILL inside `before-maintenance-commit` or `after-terminal`. The test asserts the nonzero SQLite rollback-journal magic before reopen, distinguishing a genuinely hot journal from harmless postcommit exit. Maintenance uses the ordinary cache and over 3 MiB of receipt detail. The terminal case uses a small real SQLite cache to force dirty-page spill. After repair both reopen successfully, preserve all 60 committed receipt/page URLs and full selected postimages, retain 60 groups/180 Resources, and retract the terminal case's one uncommitted pending admission.
- `root-controls-red.log/.exit`: expected 500 private gap instead returned 404 for an expired failed root. After repair, actual GET/HEAD on completed/failed expiry-eligible roots returns the gap for their principal, 404 for a foreign principal, no serving context, and no gap body. Direct inspection after close (without maintenance-triggering reopen) proves both full retained receipt documents unchanged.
- New concurrent service test parks two real page minimum reads, commits another real mutation, starts close while both reads remain parked, then releases them. Both return their original page facts with actual `pos-2` context; close drains, then further access refuses. This is bounded interleaving/lifecycle evidence, not production-load qualification.
- The independent `pagination.maxPageItems` mismatch branch is now tested. Unsupported page gaps explicitly assert all three serving-context fields absent.

## Verification

- PASS frozen install and final build (`install.log`, `build-final.log/.exit`).
- PASS **523 tests / 15 files**, no skips, **5.11 seconds**, maxWorkers 2 (`regressions.log/.json/.exit`). Includes all Transactional suites plus authority Read, lifecycle and both pagination suites. Preliminary narrow run: 81 tests / 4 files passed (`targeted-green.log/.exit`).
- PASS typecheck, format, boundaries and lint (matching `.log/.exit` files). Lint retains three warnings and one info in untouched files; no new changed-file warning.
- PASS installed `smoke` and `smoke:development`; both retain private/nonattesting admission bounds (`smoke.log/.exit`, `smoke-development.log/.exit`).
- PASS unchanged 89-row Read evidence (`evidence-verify.log/.exit`). No evidence or capability constant edited or regenerated.
- PASS actual predecessor-file and mid-page-insert probes (`old-format-and-partial-insert.mjs/.log/.exit`). The d325 emitted owner creates an actual prior-format database containing a completed receipt and pending admission; repaired open at an expiry-advanced clock refuses byte-identically, preserves both rows and creates no page table. A throw on the second page-range insert still rolls back every effect/page/receipt, and the same key succeeds as 2/2/1 after reopen.
- PASS diff whitespace check and clean postcommit status. All spawned crash children exited; disposable fixture directories are removed. No test run was interrupted.

No full workspace suite was repeated; root owns broader integration. SIGKILL and injected SQL failures are not claims of physical disk or power-loss validation.

## Scope

Seven files: store/service/HTTP implementation, service/HTTP tests, new `transaction-recovery.test.ts`, progress documentation. No schema, normative spec, Types, shared Update writer, private storage format, semantic identity, public export/admission, Read observer or sealed evidence change. The separate set implementation can ordinarily merge this frozen repair before resuming overlapping owner/service/HTTP work.
