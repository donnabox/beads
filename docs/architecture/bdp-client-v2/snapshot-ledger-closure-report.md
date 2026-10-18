# Snapshot ledger review closure

Frozen clean source: `f72566bed6f9a9cd7c9ca2ae05759c5a3f26d4e6`, direct successor of combined `dc15bc92bc83253da1d3d1fc7f47df6ef3b4924b`.
Checkout `/Users/dbox/repos/wt/vickie/vickie-snapshot-ledger-closure`, branch `codex/vickie-snapshot-ledger-closure`. Six files, 168 insertions/18 deletions. No outward writes. This is the author report, not independent approval.

## Verified findings and dispositions

The findings from `snapshot-core-efficiency-claude-review.txt` were checked against actual source before implementation.

- **F1, accepted:** classification now calls `json_type` once per ledger row using `ifnull(...,'?') NOT IN ('true','false')`. Ledger count/size still comes first, then classification, then detail integrity; malformed JSON fails before live selection. Compact ledger scans remain on Snapshot entries and maintenance. This is not O(1) work or a latency guarantee.
- **F2, accepted:** separate detail-driven guards reject missing or expired parents for each of `snapshot_records` and `snapshot_pages`. The inner `SELECT DISTINCT snapshot` uses existing covering indexes; each distinct handle causes one indexed parent lookup per table rather than one per body/page. This avoids probing empty detail for every expired historical handle. Actual `EXPLAIN QUERY PLAN` output is saved in `query-plans.log` with its reproducible script. The tradeoff is scanning retained-detail index entries to collapse duplicate handles; it is not free or a new independent durable counter. Separate orphan-record and orphan-page tests delete the parent, preserve pending claims, and assert startup refusal plus byte-identical ordinary file contents and retained detail counts. Existing expired-detail, malformed expiry and cold-corruption tests still pass.
- **F3, accepted as an internal precondition, not a present caller bug:** capture requires every flagged-live disposition to remain unexpired at its captured instant. The actual owner still commits expiry first using that same instant. A deliberate internal-helper bypass now fails integrity before allocating any handle; ordinary owner capture expires then succeeds. There is no silent filtering of retained detail from live budget accounting. An explicit nine-live corruption fixture preserves the existing integrity refusal rather than weakening it to an ordinary capacity response.
- **F4, accepted:** epoch is read once in the synchronous provisioning/open transaction after configuration/capture validation and before pending retraction/first sweep. Store identity, readConfiguration and shared readContext use that value. Dynamic position remains canonical-validated inside every guarded entry. Receipt and point-read tests now precisely assert the original two transactional metadata/receipt observations and two current-position reads respectively, including reentrancy rejection. No dynamic epoch policy or administration semantics are introduced.
- **F5, accepted:** the real 65,536-disposition expiry loop is now under a SQL prepare spy. All non-metadata statement preparation must be unique; the disposition UPDATE and both detail DELETE statements must each be prepared exactly once. This exercises the hot loop itself, unlike the earlier zero-expiry capture witness. The full fixture still expires, permanent-cap refusal still occurs, and subsequent refusal performs no JSON parsing.
- **F6, documented limitation:** committed closure note now explicitly states that maximum population/byte fixtures use empty owned-Link planes. Smaller owned topology tests establish consistency; no combined ceiling-owned workload or universal heap qualification is claimed. No new large-capacity exercise was added.

The a0 successor-token reserve, 9→10/99→100 page tests, prior format/view pre-maintenance refusal, fixed finite Snapshot limits, cold expired commitment lookup validation, and all prior durable evidence remain intact. Format stays 10. No new wire protocol, schema, public exports, rate/forgetting policy, Q1 decision, runtime settings/readConfiguration tuple, package inventory or evidence file changed.

## Exact validation

Pinned Node24.16.0/pnpm11.20.0; frozen install and build passed. Build preceded all emitted/child-process tests.

- **306 tests passed, 0 failed/skipped across eight files**, 8.34 seconds: Snapshot44, store132, point-read20, receipt pages13, Events21, changefeed20, owned17, HTTP39. `focused.log` and `focused.json` preserve actual outcomes. No timeouts were widened.
- `typecheck.log`, `check-boundaries.log`: exit0.
- `format-check.log`: exit0 with the existing 2.1 MiB matrix size warning.
- `lint.log`: exit0, 93 warnings/2 infos; not a warning-free claim.
- `smoke.log` and `smoke-development.log`: both exit0. Unlike the prior a0 repair base, this combined ancestor contains root's Snapshot package inventory. Installed smoke observed private configuration identity `bc807295939d88df15cc027808d4fc807992008538efaccc58df4ca53e3d2967`, payload identity `78ed45b267061ad47ce2ba91edd97de34879c44ced73639908c8d379b9e7a3be`, artifact identity `development-schema5-e1dff401b5caafb04a4d935775d67eeefeae7bcf306d5d100b979a64eef148e7`; `claimEligible:false`.
- `evidence-verify.log`: exit0, existing 89 Read rows verified with their original source/evidence identities. No fresh Snapshot cohort or evidence relabeling.
- `git diff --check`: passed before commit. Clean tree after commit. No full suite or fresh packed conformance cohort was repeated; root owns combined qualification and independent review.

Initial typecheck caught TypeScript's inability to infer assignment through the synchronous startup callback. The final declaration explicitly states definite assignment, with assignment before first context use; final typecheck and build pass. No source edits after the frozen commit.
