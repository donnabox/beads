# Native closure review: artifact-set step 2

No actionable findings in this bounded successor. The seven accepted external findings and the native combined cancellation/I/O finding are closed by the inspected changes and targeted checks below. This is repair closure, not public profile qualification or publication approval.

Reviewed exact source **c91f45cc5c827f5d4684a62c73f7b296c1e65230**, direct parent **b569524dfe1826b0b43ac77afc1accf362d906db**. Author checkout remained unchanged. Independent clone: `/Users/dbox/repos/wt/vickie/vickie-artifact-set-review-closure-native`. The delta contains only the two app tests and private artifact-execution helper; no runtime, runner, manifest, fixture, coverage, schema, capability or budget changes.

## Closure dispositions

| Finding | Disposition and evidence |
|---|---|
| F1 independent continuation routes | Closed. Matrix lines 1150–1183 explicitly enumerate all 26 routes and pin `59c70c9f22238cd7711a19034139c404136564ad50c58b7f3ef795a385cc990e` independently of `artifactRoutes`. Existing baseline pins remain. This adds an independent expected inventory; association still checks each observed report's exact configuration and actual result digest. The full installed success run is the author's separate fresh-pack evidence, not an execution performed by this review. |
| F2 observations versus plans | Closed. Helper lines 115–181 derives required mappings, row indices, canonical result digest references and state partitions from actual returned owner results. Omitted requirements resolve to baseline missing-plan witnesses; absent observations remain null. `authoredIds` stays a separate static inventory; `observedIds` excludes not-run/not-implemented placeholders. Uncovered IDs come from validated set coverage. Actual failed/aborted process tests verify persisted mappings against their genuine partial result rows. |
| F3 cleanup | Closed. Matrix line 798 begins cleanup-incomplete. Lines 840–860 perform real close/removal, account for row cleanup errors, and only then set/record completion. Helper line 232 separately rejects every cleanupError row for success. Actual startup, abort and cleanup controls and the specific association discriminator passed. Evidence-write failure still throws; it cannot publish an association success. |
| F4 timing and resource disclosure | Closed at the requested scope. Matrix lines 967–1025 use one origin before timer and artifact construction, record per-stage remaining start budget/outcome, and finally clear the timer even when construction fails. The unchanged 90-second aggregate, 120-second outer, 30-second focused and 5-second exchange bounds remain. Source byte/node measurements are separate report evidence; process RSS is explicitly a supervisor child-tree sample, not a per-shard or universal memory guarantee. This review independently rechecked byte lengths/digests and parent equality; it did not measure another packed-run RSS peak. |
| F5 digest scope | Closed. `resultDigest`/`resultBytes` identify canonical genuine result serialization. Helper persist lines 448–459 computes `writtenReportDigest`/`writtenReportBytes` from the exact encoded wrapper bytes only after successful replacement; successful summary references them without self-hashing. Independent focused test rereads each file and recomputes both digest scopes. |
| F6 credential screening | Closed. Successful summary goes through the same safeJson boundary before publication (helper lines 545–569). The secret-bearing summary control passed and emitted no successful summary. Literal, base64, percent-encoded and key credential rejection controls also passed. This is known configured credential exclusion, not a claim of arbitrary secret detection. |
| F7 negative discrimination | Closed. All 22 inherited association controls now assert the intended mismatch label. Separate planned-manifest and row-cleanup controls supplement them. These controls ran in the 450-test regression suite, so an unrelated earlier rejection cannot silently satisfy the intended cases. |
| Native cancellation plus I/O | Closed. Matrix lines 862–877 preserve both original signal reason and distinct execution/cleanup errors; runShard lines 941–959 also preserves cancellation when evidence operations fail earlier. The permanent actual-process chmod test reproduces the prior real-response cancellation plus EACCES combination and passed independently in 1.059 seconds. It checks original Error identity, EACCES presence, unchanged retained baseline bytes, absence of successful summary and child termination. The chmod case was not skipped on this host. Raw causes remain in memory, not saved evidence. |

## Independent verification

Pinned Node 24.16.0 and pnpm 11.20.0. Commands are recorded in `artifact-set-review-closure-native/run.py` and `focused.py`; full logs, JSON test results and actual exits are beside them.

- Frozen-lockfile install: exit 0.
- Pre-build `pnpm typecheck`: exit 0.
- `pnpm build`: exit 0.
- Seven-file conformance/association regressions: **450 passed**, no failed/skipped, exit 0. Files: artifact-execution, artifact-set, artifact-bundle, cumulative-artifact, profile-contract-plan, runner and transactional-cumulative.
- Focused app matrix/writer run with `-t 'private attempt persistence|retains actual artifact|retains actual returned|preserves actual abort and IO'`: **15 passed**, 173 intentionally unselected, exit 0, 6.42 seconds runner duration. Includes seven real-process failure controls and eight writer controls.
- `pnpm evidence:verify`: exit 0, historical Read seal unchanged.
- Independent source audit: all five artifact input files have identical bytes to b569 and match the author's reported digests/byte counts. This preserves all 49 baseline blocks, partial C, 129 required IDs and existing 30 whole/99 uncovered diagnostic accounting; no credit promotion.
- Both frozen author source and disposable review source were clean at inspection. Process inspection after completion found no test-owned child process. No outward action.

No full suite or fresh installed cumulative success run was duplicated while the author packaged the same frozen source. The independently executed controls prove the changed failure/association paths; actual 50-row packed observation remains separately source-bound author evidence. Original b569 red probe and prior reports remain historical and unmodified.
