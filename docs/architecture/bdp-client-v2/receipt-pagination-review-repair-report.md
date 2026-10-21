# Receipt pagination council repair

Frozen clean source `0359f1cae3f36f600b1b375ea276ec7d60bf689b`, tree `6a5337cf9548cfe340593a3858213593e4aca1d9`, ordinary direct child of PR101 source `c872f035ed16222711ef1736c031024cc5c3d309`. Checkout `/Users/dbox/repos/wt/vickie/vickie-receipt-page-review-repair`, branch `codex/vickie-receipt-page-review-repair`. Seven bounded TX conformance/helper/test files change; no store/runtime/schema/catalog/fixture/RU edits. Both prior literal32 pins remain.

## Findings disposition

Medium: a completed actual mutation whose captured `/next` is null now raises `TransactionalReceiptFailure`, classified as target `fail`, instead of `TransactionalHarnessError`. Absent `/next` remains rejected by the normative receipt schema as target `fail`. Actual raw-HTTP runner cases prove both, record a completed200 source exchange and no page request. Restoring only the old error class makes the null-next regression fail specifically because it receives `harness-error`; `missing-page-before.log` preserves that discriminator. The final focused suite passes with the repair.

The new `p-update:next-page` assertion declares `/next` presence. Presence alone accepts null; this is explicitly not treated as a non-null oracle. Existing canonical same-parent page capture validates the actual URL and rejects null. Added `p-root-read:next-page` independently compares actual `/body/next` with the previously validated `p-page` binding, while helper/root whole-body equality ties that GET to the original POST. Both assertions have exact ordered coverage bindings. The sixth installed control replaces that typed root GET equality with literal null (deleting equalsBinding through the existing control seam), then requires exactly the named assertion to fail on actual200 wire. No new assertion grammar or capture API.

Low boundary: static authoring test independently pins matched-count32, the ordered matched-plus31-updated result tuple array/projections, absent inline results32, single-page absence at results1 and the four literal p31 Resource facts. It also pins both new continuation assertions. Earlier f5 independent checks of both complete32 current-state tuple arrays remain intact.

Low ownership: clauseMap now pins both `#mutation-receipt-responses` anchors for `transactional.receipt.pagination`. No normative source or catalog changes.

## Budgets and preservation

Before tests, standard formatter produced manifest377449B, growth138756B,10082 authoring nodes; coverage54875B, growth13063B. All remain inside unchanged136KiB growth/370KiB total/10200 node/+16KiB coverage fences; growth headroom508B. Eighteen new/139 total actions and unchanged execution/raw/wire/timeout caps. Manifest SHA `5e0e4c688363a8d530ec2d5347522d0567755faf292a7ca313c8d5e5426f9571`; coverage SHA `1336c8e83a775856594955f7ad399582c4cd37422b95e7a26d803d3ce4ec27d7`. All eight predecessor raw scenario/coverage objects and fixture bytes remain unchanged, enforced by passing static tests. Old df51 evidence and all failed attempts remain untouched.

## Exact validation

Pinned Node24.16.0/pnpm11.20.0. Clean build, typecheck, format, lint, boundaries passed. Focused helper/runner/cumulative suite:337 tests pass. New null/absent-next tests use actual raw HTTP. Before-fix discriminator is intentionally red; restored final source is green. Original Read evidence verification passed.

Fresh packed source0359: seven selected installed tests passed (nine-row positive plus six pagination controls),37 unrelated app tests intentionally filtered. Positive `3372.577`ms,139 actions,15 sessions plus bootstrap,16/16 children closed;12 reopened receipt observations inclHEAD remain checked by the existing app oracle. New row maximum raw5104B and response8985B, unchanged. All nine partial rows pass,246 missing-plan rows remain, claimEligible=false. Full runtime/profile qualification is not claimed. RU/full44 app cohorts were not repeated per root's bounded-repair instruction; their source is untouched and predecessor evidence is not relabeled.

`safe-summary.json`, `source.json`, `complete.json`, packed tarballs and installed logs provide exact source/config/payload/hash/timing binding. Actual runtime configuration identity `e3f28f343b43dc938f95515ef96769e47aeca7dbbe1097f43028096d6627d999`; installed payload `1b12629fe54c39c8f7864dbb13bdba16bcbe4298b0a5bd420fa19daaa4fb7ec2`. No timeout, forced termination or owned child remains. No push/PR/main/bus writes. Native/root council acceptance remains external to this author report.
