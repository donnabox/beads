# Schema PropertyChange evidence — frozen author report

Completed 2026-10-14T02:58:20.562745+00:00. This is bounded, nonattesting evidence. R remains `wholeContract:false`; public admission is unchanged.

## Source and scope

- Head `c57f5a7a804b1b7ababf2d9195d7fdccbd37dee8`; tree `b7ffb7fb34ea5014ba824be945026a353eb94a70`.
- Direct ordinary parent `627b560c0daf1a76b292615718c86670255ac9af`; branch `codex/vickie-schema-property-change`.
- Checkout `/Users/dbox/repos/wt/vickie/vickie-schema-property-change`, clean, no pushes or external mutations.
- Accepted plan SHA-256 `76fa65b2586ab0f5baa09523744e5dd27f4a3c22947d6364e8502aaad9f958ce`. The twelve RU basis files were byte-equal to e086 before editing (`basis.json`).
- Exactly three files changed: continuation manifest, RU coverage, and `apps/bdptest/src/read-update-matrix.test.ts`. No runtime, schema, catalog, fixture, runner, TX, or public-admission edits.

## Observations and independent controls

For both Bead and Link, the existing schema Resource's corrected create now captures its URL; the existing GET independently checks its complete declared state and captures that actual validated preimage. Eight new actions exercise schema-invalid `propertiesChange` replacing `/n` with -1, full unchanged GET equality, a fresh-key correction replacing only -1 with 1, and independently expected current-state GETs. Both carry the companion metadata patch, so the refusal proves it did not commit.

The refusal requires the complete configured Problem, including schemaLocation `/properties/n/minimum` and instanceLocation `/n`. Corrected results preserve all four absence checks; the final GET independently pins n=1 and metadata.phase=updated. No weaker Problem or comparator oracle was introduced.

Four actual-transport assertion-sensitivity controls passed: Bead schemaLocation corruption, Link refused-state metadata corruption, Bead corrected-state n corruption, and Link corrected-state metadata corruption. Each applies exactly once and fails exactly its selected assertion after an actual expected response; cleanup remains required. Existing raw/client control mechanisms were reused without executor changes.

## Preservation and budget audit

`source-audit.json` independently checks all four other continuation scenario blocks, 84 untouched R action blocks, the 39 original raw carriers outside the two authorized GET-to-client conversions, every old assertion ID, and all eight existing control-table initializer source spans. The original baseline manifest, fixture, index and catalogs remain byte-identical by Git diff and the basis audit. Coverage changes only R's reason and ordered binding list; all former bindings remain.

- Continuation: **899,400 bytes / 23,502 nodes**, under unchanged 901,895 / 23,546 ceilings (2,495 bytes / 44 nodes remaining).
- Coverage: **712,792 bytes**; only accepted +2,560 / +2,048 guard increments were applied.
- R: **96 actions / 687 assertions**, including43 raw exchanges,53 client actions and36 actual raw POSTs.
- N remains52 actions/373 assertions; combined **148/1,060**, under the accepted148 action and unchanged1,300 assertion limits.
- Eight inserted actions add66 bindings; two converted GETs add2 write-state bindings. Existing16-action binding set now100; the old98 remain.
- Baseline route digest `3224c31c83988dc0606cb76ee1f4ba8e6394159bdf7daaccc301461316cc2477` and continuation route digest `272c31e6bdf474d3889042e3bc636f4f144b58ea1e87a8f8176c772a1a026110` remain pinned and unchanged.
- Inventory remains54 authored /129 required /33 reviewed-whole /96 uncovered /75 omitted. R stays partial; no whole-list promotion.

## Verification at frozen source

Pinned Node24.16.0 / pnpm11.20.0; frozen-lockfile install, typecheck, clean build, boundary checks, format, lint, both executable smokes, and historical Read evidence verification all passed. `stages.json` records every final command exit and elapsed duration.

- Static selection:22 passed (193 filtered), three files.
- Final focused catalog/cumulative tests:20 passed, three files.
- Fresh packed installed cohort:23 passed (179 filtered), including genuine both-shard54-row/129-required aggregate, R/N positives, nine inherited numeric controls, three inherited R controls, four new PropertyChange controls and related static guards.
- Installed selected cohort:29.844s. Genuine aggregate cumulative elapsed18.452s, within unchanged90s aggregate /120s outer budget; per-exchange5s and focused30s remain unchanged.
- Fresh packed payload `1b12629fe54c39c8f7864dbb13bdba16bcbe4298b0a5bd420fa19daaa4fb7ec2`, contracts `48667497d6e13064316d0364cf8f49dfc247ab9189306985a357d7b558333dfe`. Both are actual captured target identities; sourceCommit is separately caller-bound to the frozen commit.
- `independent-report-audit.json` verifies all129 mappings resolve to genuine report rows and both canonical result digests plus exact written-file hashes/lengths match. `summary.json` retains actual per-shard manifest/route/config/source/report identities.
- R positive actual43 raw /53 client /36 POST; N21 raw /31 client /16 POST. Full reports are `numeric-admission.round-trip.json` and `.nested.json`.
- `cleanup.json`: all reported cleanup completed; zero remaining installed child processes. No caffeinate was used. No owned supervisor remains.

## Failed preparation attempts (retained, not relabeled)

1. Initial typecheck rejected assignment into a readonly test clone; the negative fixture now uses a fresh mutable array. `typecheck-initial.log` and its explicit exit preserve this preparation failure. No artifact semantics changed.
2. Initial static run:21 pass/1 fail. The new84-action preservation assertion compared factory-normalized key order against raw-source JSON.stringify order. Parent and current raw preimages both hash `393f9c3cab0cbfd501e863a75a243a28fa40806c648993eaf2068916cf488c20`; normalized preimage hashes `72a33538573525f628f76735c1e08c342115b45c55efe9e97c6ec5c6238986e7`. The decoded JSON values are equal; the factory changes property order. Both exact preimages are retained. The guard now deliberately hashes raw source, keeping the independently established parent hash. `static-initial.*` retain the failed attempt, and temporary extraction instrumentation was restored before commit.
3. The external preservation audit initially attempted to compare the two explicitly converted raw GET carriers as unchanged; it was corrected to name those authorized conversions and check the remaining39 raw carriers. This was audit bookkeeping, not a source change.
4. A preliminary emitted focused run passed16 selected tests before freezing. Its generic focused filenames were subsequently reused by the final20-test catalog run; this report relies on the durable final23-test fresh-installed evidence instead of presenting the earlier run as a separately archived result.

## Limits / next

These are finite configured schema partitions, not complete schema-language or profile conformance. Raw corrected response-to-GET opaque revision linkage remains unobserved, as the coverage reason says. No claim that all possible schema failures have been exercised. The narrow44-node continuation headroom is explicit; further growth needs separately accepted artifact budgeting/segmentation. Council and CI decisions remain root-owned. Source is frozen and available for independent review.
