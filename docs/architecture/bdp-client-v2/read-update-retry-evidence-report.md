# Installed Read+Update retained retry evidence — freeze

Clean local commit **29e715162691e6d0a4bcc1ce699f33e243ebecdd**, ordinary parent **8e45ef4d2c9674296f217ac0cf75bc706f49ff09**. Checkout `/Users/dbox/repos/wt/vickie/vickie-read-update-retry-evidence`, branch `codex/vickie-read-update-retry-evidence`, origin `https://github.com/gastownhall/bdp.git`. No pushes/PRs, merges to main, coordination/launch edits, public admission changes or runtime changes.

## Delivered bounded evidence

Two formerly omitted IDs now have row-local actual installed HTTP observations:

- `read-update.idempotency.cross-carrier`: singleton↔sequence retained successes and failures in both directions, fresh state deliberately diverging from retained postimages, changed-input conflict and original-key recovery. Missing-subject failures remain retained after the subject becomes live; otherwise identical fresh-key updates positively prove the operation can now execute.
- `read-update.sequence.idempotent-retry`: allocated creator and other retained success/failure members are reordered, one successful creator gains a name and another loses it, and a fresh independent member is inserted. Index-only ordered tuples plus per-member name/outcome/resource/Problem assertions preserve the actual assertion grammar. Independent GETs show newer live state while retries return old images; repeating the reordered request recovers the same full result array.

Both remain **partial**. Inventory: **129 required / 45 authored / 26 whole / 103 uncovered**;19 partial and84 omitted. `claimEligible:false`. Retention/expiry, in-flight behavior, disconnection, crash/restart, changed authorization, other principals/Scopes, full operation-kind/alias/pin/local-binding identity variants are not qualified by this slice. A deliberate completed-response retry is not lost-response recovery proof. No other required row is credited from neighboring observations.

Four files only: RU matrix manifest, RU coverage, RU fixture bindings and app matrix tests. Added60 actions/491 assertions including row-local discovery preambles; seven fixed Bead GET bindings. All43 prior manifest rows are semantically unchanged (independently compared against parent JSON); canonical row order preserved. Neither normative catalog nor runner/executor/H1/M1/runtime/schema was edited. Existing schema wrappers remain unchanged.

Thirteen controls alter assertion input **after** real wire capture: refreshed instead of retained image, changed retained revision, old failure replaced by success, forbidden singleton sequence field, regressed GET postimage, lost conflict, overwritten retained result, old position, unwanted/incorrect name, added failure detail, refreshed sequence image and regressed sequence GET. Every control applies exactly once and causes its intended named assertion to fail. The test requires that failure; raw recorded observations remain untouched. Known properties/metadata/IDs/Types and direct afterstate GETs complement opaque captured Resource/revision equality.

The two exact route-digest pins now match actual `092b9ac9abf64dcd02f592059bc6333b0a6dfc69efe9d50332086335474c8f3c`. This changes only the closed allowlist's seven observed fixture GET routes; no wildcard/credential rule changed. The unused-fixture-binding denial remains tested. Fixture digest changes because those named paths are part of its bytes; runtime contracts/configuration/payload are unchanged from parent.

## Council dispositions

Original plan is preserved at `read-update-next-evidence-plan.reviewed-original.md` with SHA256 `e2236d26337ef1deaaa426806e4f3805b2eb47b286d507a6006ee6b09bb0dcfb`. F1–F7 dispositions are appended to the active external plan, not silently substituted. Fixed-path client GETs use the approved fourth fixture file, optional names and Problems use decomposed assertions, existing cross-carrier conflict evidence is acknowledged, exact spec lines corrected, both digest pins changed and canonical insertion order retained.

One authoring correction was separately accepted by root: updateBead cannot have a local name. The missing-subject update therefore remains unnamed; successful creates supply the name addition/removal witnesses. No new protocol ruling was needed.

## Validation and exact artifact provenance

Pinned Node24.16.0/pnpm11.20.0. All logs/exits are under this report's directory. Frozen install and clean build preceded emitted tests. Baseline unchanged43-row aggregate passed in **11.840s** (Vitest12.32s). Expanded45-row aggregate passed in **12.329s** locally and **12.229s** freshly packed. Existing cooperative90s/outer120s budgets remain unchanged.

- **97 tests, two files passed**: entire app RU matrix, including14 new tests (two-row positive test plus13 controls), and cumulative artifact invariants.
- **Fresh packed126 tests, four files passed** at exact clean29e7151: RU matrix, exact client, development server and installed Read oracles. Pack/install/matrix have explicit exit records; matrix process wall63.983s. All45 actual scenario sessions pass, with781 client wire observations,44 setup observations and21 native Read observations. Full packed report is `packed/progress.json`; bindings summary is `packed-summary.json`; test results are `packed/matrix.json`.
- Final build, typecheck, dependency boundaries, formatting, lint, both smokes and historical Read evidence verification exit0. Large manifest was explicitly formatted and checked with command-local `--files-max-size=4194304`; repo configuration unchanged. Normal formatter skips that preexisting >1MiB file, so the explicit check is load-bearing.
- Lint exits0 with57 warnings/one info, no autofixes. This includes non-null assertions for the fixed two-element retry row list in the new control table; no hidden lint failure. No claim of a warning-free repository.
- Full repository suite was not run for this four-file evidence-only slice. Historical Read verification retains its original observed source/seal and is not resealed as this commit.

Fresh packed bindings:

```
sourceCommit: 29e715162691e6d0a4bcc1ce699f33e243ebecdd
sourceDirty: false
installedPayloadDigest: cac59fd75a1e406211f4c39fe77e70ed1cfb9c1c4e5d625e1f93e0bc015511c5
installedContractsDigest: 48667497d6e13064316d0364cf8f49dfc247ab9189306985a357d7b558333dfe
policyDigest: dbe8dbd5430f26e19e754d38167cd8084a32d702c695550e2ee0063dc2912430
authenticationDigest: 66941efc779f4f2431d65fdd5b2387b2af67bd7c1559e09dac032f82259ea16e
fixtureDigest: a200afeb240f4c0c14e7bab43605d753e1640de8f17920c7dc5c174a901fb450
```

Initial unsuccessful runs are retained: `initial.log` exposed invalid slash-containing assertion IDs in the authoring script; `controls-initial.log` exposed the test's action-report lookup using action.id instead of action.action.id. Corrected controls pass in controls.log, the full matrix and packed run. These were authoring/test defects, not observed runtime defects. No timed-out/stopped run or forced process cleanup occurred. All test-owned authorities closed through existing finally blocks; no owned children remained at final process inspection. Source tree remains clean. External generator and pack script are retained for reproducibility; neither is a new repository runtime or test framework.

Ready for independent review; root owns integration/publication. Small remaining work should be selected against actual unmet clauses, not partial-row count. This slice makes no Transactional/H1 claim and does not resolve Selector complexity or metadata-only incident Link questions.
