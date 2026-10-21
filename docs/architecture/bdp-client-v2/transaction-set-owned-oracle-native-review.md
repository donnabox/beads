# Native closure review: installed owned-set oracle

## Disposition

**P2 closed; no new actionable findings.** Frozen `c9345e51da225d11a801c0d86c024d7cb3a5fe30`, direct child of `2f7a1641c613680c5b4dfe056b933bfa6af869dd`, repairs the exact installed-oracle defect from `transaction-set-http-native-review.md`.

I inspected the full one-file delta: `apps/bdptest/src/transactional-set-installed.test.ts`,46 additions/4 deletions. It captures actual pre-update owner/Link state, asserts the literal requested Link properties, unchanged metadata/source properties, different Link and owning Bead revisions, effectPosition presence and advancement, required/effect position equality, receipt/live identity and owner embedded-Link identity. Existing no-op, rollback and reopen assertions remain. The permanent stale-output control uses the captured prestate and requires the named literal assertion to reject it; the negative-oracle count correctly rises from4 to5. No runtime, schema, fixture, budget or public-admission change is present.

## Independent reproduction

Isolated detached checkout: `/Users/dbox/repos/wt/vickie/vickie-transaction-set-owned-oracle-native`. Author checkout and predecessor were untouched. Remote guard succeeded before review. Node24.16.0/pnpm11.20.0, maxWorkers2.

- Frozen install: exit0; clean build: exit0.
- Unmodified emitted-source installed test: **1 test passes**, exit0.
- Exact original runtime counterprobe: **1 test fails**, exit1, at `transactional-set-installed.test.ts:336`, `requested owned Link properties`, expected `{n:1}`, observed `{n:0}`. It fails in the positive observation, before the permanent negative control, rather than at an unrelated exception. This is the same precise one-expression mutant that passed under parent2f7.
- Byte-restored runtime installed test: **1 test passes**, exit0.
- `git diff --check` for the sole repair delta: exit0.

Exact command for each test run:

```text
/Users/dbox/.nvm/versions/node/v24.16.0/bin/node /Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs exec vitest run apps/bdptest/src/transactional-set-installed.test.ts --maxWorkers=2 --reporter=verbose
```

Reproduction script and logs: `transaction-set-owned-oracle-native/probe.py`, `probe-result.json`, `baseline.log`, `exact-mutant.log`, `restored.log`, `install.log`, `build.log`. The script alters only ignored emitted `transaction-batch.js` in the isolated clone and restores it in `finally`. Original/restored SHA256 `10edbb4cf0e196cd000c4a19900875e46832fbb9a2b39bd4797f15a88375a1ff`; exact mutant SHA256 `de93ffcfd9e96db43c8f96176fc00c95d565c3107296d29146f1799ea91b14a8`. The probe verifies both digests and the named expected/observed values. This intentional countermutation is not qualification evidence.

## Scope and cleanup

Both source checkouts remain clean at exactc9345e51. No test-owned installed child remains; failure and success paths completed cleanup. No source changes, pushes, PR or coordination writes. This targeted closure uses actual emitted-server HTTP observations; it does not relabel the author's packed results as independent runs or repeat the full subsystem/repository suite. The previous full-diff native review, external full2f7 review and broader qualification remain separate. Public profile claims and recorded Selector/candidate-enumeration gaps are unchanged.
