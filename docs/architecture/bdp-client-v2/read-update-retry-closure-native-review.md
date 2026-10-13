# Native closure review: retained retry Problems

**Accepted: the prior P2 is closed. No new actionable finding.**

Reviewed frozen `965f27cd1268c390d7b7c9e6302b0e1f8a76e4eb` against its parent `29e715162691e6d0a4bcc1ce699f33e243ebecdd`. Author checkout `/Users/dbox/repos/wt/vickie/vickie-read-update-retry-closure` remains clean and untouched. Tests and review-only probes ran in separate `/Users/dbox/repos/wt/vickie/vickie-retry-closure-native-review`. Read the author report/packed summary and prior native finding; no external implementation review used.

## Closure evidence

- Parsed manifest comparison confirms exactly13 additive whole-object assertions in the two retry rows. Every pre-existing action/input/capture/assertion is unchanged, and all other43 scenarios are unchanged.
- Comparisons cover initial sequence failure at index1 and reordered failure at index2; both singleton-to-sequence and sequence-to-singleton initial/replayed failures; final singleton retries; and supporting absence/conflict Problems. They assert the complete deterministic fixture object, including current index where present and absence of any other member implicitly through full equality. The repeated reordered response already compares its complete results array against the first reordered capture.
- Re-ran all four original native probes: added cross-carrier title, added singleton instance, added arbitrary extension, and initial sequence title lost on subsequent replay. Each mutation applies once and now fails its specific `complete-problem-*` assertion. **The same action's normative schema assertion passes in each probe**, confirming the repair checks retained fixture equality without narrowing valid protocol Problems.
- All five new permanent controls fail the immediate intended whole-object assertion. The reordered extension control is now caught on that action, rather than relying on later repeated-array disagreement. All13 predecessor controls still pass their sensitivity tests.
- Runtime, runner, executor, normative schemas, fixture, routes and route-digest pins are unchanged. Test comments and coverage reasons explicitly distinguish exact fixture equality from the extensible protocol schema.
- Coverage remains129 required /45 authored /26 whole /103 uncovered, with both retry rows partial and claims false. The closure adds no operation-kind, authorization, expiry, disconnect or recovery qualification.

## Independent checks

Pinned Node24.16.0 / pnpm11.20.0, frozen dependency install in a clean isolated clone:

- PASS pre-build typecheck.
- PASS build (including clean).
- PASS23 tests /1 file,13.16s: positive installed retry test for both rows,18 permanent post-wire controls, four original independent probes now expecting closure.77 unrelated tests filtered, not counted as passing.
- PASS6 cumulative artifact tests /1 file,0.285s: source digests, canonical coverage/bindings/order, profile carrier and bounds. Total29 independent test passes across the two runs.
- Author frozen HEAD verified and worktree clean after review; no review-owned installed child remains. Review clone retains only four review-only test probes, no source repair or commit.

Independent logs and exact probe patch: `/Users/dbox/repos/wt/vickie/bdp-integration-logs/read-update-retry-closure-native/` (`install.log`, `typecheck.log`, `build.log`, `focus.log/json`, `artifact.log/json`, `probe.patch`).

Inspected author packed `progress.json`: source965f27c clean; all13 added assertions passed;45 actual sessions /781 client observations; public claims false. Recomputed source manifest SHA256 `f5ef88ba8e2032fdd04bd8ecc5b3d2705db2cf95447df87a52c0d2c0e93fcf94` equals packed report binding. Author's131 packed tests and full45 cohort are author evidence, not independently rerun here. Packed result includes134 catalog entries, distinct from129 selected required obligations and45 authored executions; no extra qualification inferred.

Full repository regression/static checks and pack/install were not repeated for this bounded oracle closure. No outward action or public promotion.
