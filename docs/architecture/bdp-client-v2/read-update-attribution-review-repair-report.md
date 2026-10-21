# PR94 attribution/numeric review repair

Frozen source `9c7b643d22562ebabd8d9b51a24157b66f3e5c89`, direct ordinary parent `2bc2a94214954e3e03261d5e8b0dd2baac2a7268`.
Checkout: `/Users/dbox/repos/wt/vickie/vickie-read-update-attribution-review-repair`. Branch `codex/vickie-read-update-attribution-review-repair`.
The original 28a and integration 2bc sources remain untouched. No outward writes.

## Dispositions

1. Three fresh-key guarded admissions now follow the Task, unowned Link and owned Link no-op GETs. Each GET first asserts exact complete Resource equality against its independently pinned pre-no-op Resource, captures the observed revision, then the next fresh-key operation uses precisely that binding for expectedRevision. Each admitted command asserts its entire postimage equals the old expected Resource, and a following real GET repeats that equality. Owned Link also preserves the independently observed source revision and existing owner/target afterstate checks. No response becomes its own expected postimage. Static tests pin ordering, capture source, fresh key and guard binding. The new post-wire owned-guard status409 control applies once and must fail exactly `status`.
2. Numeric refusal-success control now checks selected clientAction.input deep-equals execution.input before capturing the independently observed successful Resource. Existing exact failure list and actual-wire observation remain.
3. Raw control cases have a named array; control budget uses both actual array lengths. Thirteen client controls plus three raw controls =16. No control was removed.
4. Reproduced exact CI38447602878 catalog failure before fixing: numeric-model-catalog.test.ts 4 pass/1 fail, exit1 (`catalog-red.log`, `catalog-red.exit`). The minimal test-only fix explicitly maps read-v1 to the existing Read numeric companion and reference-continuation to same-value-no-op. Every other catalog row/artifact pair remains denied; numeric-model is still forbidden as an executable catalog identity. Existing profile selection, Read catalog separation and normative citation tests are unchanged. Searched analogous catalog tests; Read+Update already explicitly recognizes the two approved shards and tests unapproved filename/tag denial. Owned-wildcard and Transactional remain correctly unbound for their mutation rows. Included all four catalog suites in the focused regressions; no catalog/schema/runtime changes.

## Scope and budgets

Four changed files: continuation manifest, attribution coverage bindings, app matrix test, numeric-model catalog test. Fixture, baseline manifest, artifact-set index, all old49 scenarios, restored C block, complete numeric B block and runtime bytes are unchanged. The byte suffix beginning at C was compared to parent2bc; original numeric lexemes are preserved. The initial formatting pass detected generator reformatting; it was repaired before any executable evidence run. `format-first.log` records that failed check, final checks are separately successful.

A grows69→75 actions and630→663 assertions; B remains73/533. Combined148<=150, each<=75. C remains73/358. Both new wholeContract flags remain false; required129/authored52/whole30/uncovered99/omitted77 unchanged.
Continuation:535,052 bytes/14,286 JSON nodes (+13,131/+341 versus2bc), below existing source caps. Growth from pre-A/B continuation114,182/3,158 is420,870bytes/11,128nodes, below650KiB/24k. Coverage648,157bytes/13,494nodes; growth57,310bytes from590,847 is below65KiB. Fixture29,055bytes/750nodes unchanged. Exact digest ledger: source-audit.json. Route set unchanged (35 continuation targets); no new resource or route.

## Completed checks

Pinned Node24.16.0 / pnpm11.20.0, frozen install exit0. Typecheck, build, format, lint, dependency boundaries, both executable smokes, unchanged Read evidence verification all exit0; explicit command/exit/time records alongside this report. Regression run283pass/10files, including exact catalog red→green and closed-world sibling tests. Initial focused attribution/numeric run18pass/155filtered; after adding the sixteenth control, targeted guard-control/static run2pass/172filtered, exit0. No full workspace test claim.

## Fresh installed verification

Fresh protocol/server packs installed into a disposable directory with npm --ignore-scripts; all pack/install/matrix commands exit0. Exact clean frozen source: `9c7b643d22562ebabd8d9b51a24157b66f3e5c89`. Five-file installed cohort242pass/0fail/0skip, wall127.36s under240s process supervisor. This is the total five-file cohort, not the individual120s outer matrix budget. Both artifacts complete in20.735s under90s aggregate (69.265s remaining); all existing time bounds unchanged. Sampler peak1,309,296KiB is the entire supervised concurrent process tree, not an owner-memory or universal heap bound. No installed child processes remain (`cleanup-processes.txt`), source remains clean.

A actual observations:6 raw discovery/schema +69 client exchanges;663 assertions; max actual client body452bytes. B:15 raw +58 client exchanges;533 assertions; max actual client body333bytes/raw body526bytes. All10 B raw mutations use real serializer observations with complete write state and withheld report body digest. Genuine child wrapper byte hashes and lengths checked against set summary. Reports remain nonattesting, claimEligible:false.

Payload `3ad235102ca01c8e28acebc2bf892b107c65d40502f97a1cfc0dc9a135c8440d`; contracts `48667497d6e13064316d0364cf8f49dfc247ab9189306985a357d7b558333dfe`; policy `dbe8dbd5430f26e19e754d38167cd8084a32d702c695550e2ee0063dc2912430`; authentication `66941efc779f4f2431d65fdd5b2387b2af67bd7c1559e09dac032f82259ea16e`. Fixture `07bbe646a5ce8436e8a746fadd22a9906013f0c28e28f608dcd9a0a84aa7b234`. Scope `https://reference.invalid/acme/`, profile read-update, history:false. Baseline route digest3224c31c83988dc0606cb76ee1f4ba8e6394159bdf7daaccc301461316cc2477; continuation35-route digest6713ab2c41b09869b664f372de1ef1390c00cfd24b154ea717f7ee6d62c35059 unchanged. Full source/config/artifact/report identities and timing are in `packed-summary.json`.

No runtime, protocol, schema, index, fixture, public admission or capability change. Both new rows remain partial; no whole-contract promotion. Prior catalog failure and formatting failure retained honestly; final installed run has no failure. No outward publication performed.
