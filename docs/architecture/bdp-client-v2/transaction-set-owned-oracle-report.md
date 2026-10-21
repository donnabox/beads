# Installed owned-set oracle repair

Frozen source: `c9345e51da225d11a801c0d86c024d7cb3a5fe30` (tree `cb1cf161639341f622a7663fd1ecff1405852d60`), ordinary direct child of `2f7a1641c613680c5b4dfe056b933bfa6af869dd`.
Checkout: `/Users/dbox/repos/wt/vickie/vickie-transaction-set-owned-oracle`; branch: `codex/vickie-transaction-set-owned-oracle`; origin: `https://github.com/gastownhall/bdp.git`.

## Scope and disposition

Native P2 was an evidence gap, not a runtime defect: mutually consistent stale receipt/live Resource observations could satisfy the original oracle. Only `apps/bdptest/src/transactional-set-installed.test.ts` changed (46 additions, 4 deletions). Frozen parent remains untouched. No runtime, fixture, protocol, configuration, budget, schema, catalog, public admission or qualification changes.

The test now captures actual pre-update owner and Link GET bodies and the creation receipt. It requires literal Link `n: 1` after requesting the update, both owner and Link revisions different from prestate, a nonempty string effect position different from the preceding required position, and equality of the new effect/required positions. It retains result/live Link and sourceRevision/live owner equality, verifies the owner's embedded Link, and checks unchanged metadata and source properties. Existing subsequent no-op, compiled-schema rollback, retained success/failure and restart checks remain.

A permanent stale-output discriminator supplies the actual prestate and mutually consistent stale receipt and requires failure at the labeled requested-properties assertion. This adds a fifth negative oracle; it does not treat an arbitrary exception as evidence.

## Discriminating runtime counterprobe

`mutant.py` applies the native review's exact ignored emitted-runtime mutation: for the selected owned Link only, substitute requested n1 with n0. The repaired test fails (exit 1) at **requested owned Link properties**, expected `{n:1}`, observed `{n:0}`. The script restores the emitted file in `finally`, verifies byte-identical restoration, then the test passes (exit 0). This intentional mutant run is not qualification evidence.

- Original/restored emitted transaction-batch.js SHA-256: `10edbb4cf0e196cd000c4a19900875e46832fbb9a2b39bd4797f15a88375a1ff`.
- Mutant SHA-256: `de93ffcfd9e96db43c8f96176fc00c95d565c3107296d29146f1799ea91b14a8`.
- Evidence: `mutant-proof.json`, `native-mutant-red.log`, `restored-green.log`; initial normal-runtime baseline also passes.

## Validation and exact installed evidence

Pinned Node 24.16.0 and pnpm 11.20.0; bounded tests use maxWorkers=2. Frozen-lockfile installation, clean build, typecheck, format check, lint and git diff check all exit 0. No full suite was repeated for this one-file test repair. The baseline/counterprobe runs preceded the commit and are explicitly separate from clean frozen installed evidence.

`pack-executor.py` verifies exact frozen HEAD and a clean working tree, rebuilds and packs protocol/server, installs into a disposable directory with scripts disabled, and runs the affected installed test: **1 file / 1 test passed**. All build/pack/install/test exits are 0. The test records `workspaceDirty:false` and `claimEligible:false`. Actual child close/reopen uses the same captured runtime identity. Logs and full wire observations are under `packed-c9345e51/`; `observation.json` retains actual captured responses and `observation-summary.json` their source/config binding.

- Protocol tarball SHA-256: `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`.
- Server tarball SHA-256: `c89ff8bc08a262a4609ef126e169ee7bda9ff5ae7901e33c57a3683a97b30428`.
- Both package hashes exactly match the parent 2f7 packages, as expected for a test-only change.
- Installed payload: `312ffe5038aefa399096870568a26013fc9c2dfa55168fff700af2a2c9ff5ce6`.
- Captured contracts: `95b4ef854865acafc04caa438285dd52b7955fa99e6b241397660d6f262b4935`.
- Actual runtime configuration: `b39183d9ca061836dae53da7ad02d88372668f828653ebe1763e7bce63a2acc6`.
- 32 recorded wire observations; 33 receipt entries across pages [32,1]; 96 Events across [32,32,32]; one group; five negative oracles.

## Limits and cleanup

This closes the owned-transition oracle gap only; it does not change the runtime or establish whole-profile conformance. Existing Selector questions/private gaps, bounded owner implementation limits, and closed public profile admission remain. Parent 2f7's broader tests and sealed Read evidence are not relabeled. No pushes, PR changes or coordination writes performed. Source is clean and frozen. The exact affected runner and installed child processes exited; final process inspection found no remaining owned children. No run timed out or required killing.
