# G2a ordinary main integration — validated baseline, follow-up pending

Clean frozen **bcfa43f04235ee04e0ca26902dc63c243d8495d0**, branch `codex/vickie-changefeed-sse-integration`, isolated checkout `/Users/dbox/repos/wt/vickie/vickie-changefeed-sse-integration`. No push, bus, PR, main or frozen-parent mutation. No G2b implementation or readiness claim.

Ordinary parents are reviewed repair `b14fb2875447a2bfd9de9fa725add412ba28ad5c` and remote main `1f9d7dbc401f6a265cf1ce0bdfea0fe493cd79a5`. Main was freshly verified with `git ls-remote` and fetched before merge. Remote role guard passed before work.

## Merge scope

Merge base94ef8da934694b6ef12653851ccee0716130a8aa. No conflicts or manual source changes. Exactly six incoming evidence/Snapshot-polish files differ from b14: RU matrix app test, Snapshot installed app test/progress note, RU manifest, cumulative artifact test and RU coverage inventory. Each is byte-identical to main. Server, protocol, client and scripts trees are **exactly b14**. `merge-audit.json` records exact parents/tree and assertions. No unexpected overlap or semantic repair was inferred. The ordinary merge was committed with sign-off; clone has no active hooks, none bypassed.

## Exact-source validation

Pinned Node24.16.0/pnpm11.20.0. `validate.py` records every command, exit and duration, with verbose/JSON test outputs and600s supervised process-group bounds. All stages exited0, naturally; no timeout/interruption.

- Frozen install, **pre-build typecheck**, clean build: pass.
- Full workspace: **8,656 passed /157 files /1 skipped**,143.57s Vitest wall time (144.67s supervisor). Skip is the optional actual-bd Read walkthrough; it is not counted as qualified.
- Dependency boundaries, format, lint: pass. Lint retains161 warnings/8 infos; no warning-free claim.
- Installed-package smoke, development smoke, seven E2E preflight tests, unchanged Read evidence verification: pass. No seal regeneration or historical relabeling.
- Fresh protocol/server tarballs installed into `packed/install/`; actual packed process entries used thereafter.
- Packed G2a SSE: **1 passed**,16.09s stage. Actual unchanged-timer idle15,188.44ms, two heartbeat comments, two complete change-group frames,2,970 wire bytes. SSE data text/finite response text and structural assertions both passed.
- Packed finite TX/Snapshot: **30 passed /2 files** (22 Transactional harness tests,8 Snapshot observations/controls),18.91s stage.
- Packed RU: **152 passed /4 files**,89.62s Vitest wall time (90.02s supervisor).47 authored scenario sessions pass;87 not-implemented harness-error results remain explicit. Required129, whole28, uncovered101, claimEligiblefalse. Reference A/B are not on this main-side source; their separate integration is not borrowed here.

No full suite or packed result from another head substitutes for these observations. Full source and packed tests remain diagnostic support for the covered configuration; public gates stay closed.

## Exact observed artifacts

`summary.json` binds source/parents, source audit, stage exits, test counts and safe installed metadata. Full genuine child reports remain separate under `packed/`.

| Artifact | Value |
| --- | --- |
| Source | bcfa43f04235ee04e0ca26902dc63c243d8495d0 |
| protocol.tgz SHA256 | 2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c |
| server.tgz SHA256 | 50a947387e97fd99c250f1fcd1fe3e514dd719b99f5530a3c366361d1f979db4 |
| Installed payload | 8be5c8a42240841a85b2233620cd5ee1c1c660f10fa03ae92f895be9063535be |
| SSE runtime configuration identity | c8230e0a2658d7519ecd7805ae8c7db7547e2957462720b7cc6724a40d7a828e |
| SSE wire SHA256 | 118f4077440e7fdad28ac82eccc44c74b6cbd7d8d3584dead69e872f7529fc1d |
| RU exact execution projection | a8d5687519da4040d2f23989f86f9bc804d30ac6c91a85074397eaf7c92de712 |
| RU manifest | 0c6e1c9e9853e82e714e919ba23fa33dbf267c8a36ae36aa9457c1a9d3efc64a |
| RU fixture | 64814924334103e9375c043de472f8bf15c8b71a267a176e06cd0443b8249751 |

Fresh tarball hashes/payload equal b14 because emitted runtime is unchanged. The fresh SSE runtime identity includes generated credential associations and is observation-specific; it must not be equated to RU's different configured cohort. Both observations bind this exact source. `packed/sse-observation.json` and decoded `packed/sse-response.http` preserve actual wire evidence; its hash was independently checked. `packed/transaction-observation.json` and `packed/ru-progress.json` retain genuine finite reports. No raw credentials are introduced into the whitelist summary.

## Remaining gate and cleanup

During validation, external b14 review completed and root assigned a separate successor for F1 destroyed-before-close/abort-capture race, F2 effective timer validation and F3 explicit finite page counts. **These repairs are not in bcfa**, and passing tests do not dismiss them. This is an exact baseline for ordinary integration of that bounded accepted repair, followed by final affected validation. Root owns council closure/publication; G2b remains held.

Final source is clean and unchanged after tests/packing. Process snapshots before/after plus final inspection found no run-owned installed child. All commands completed naturally; no unrelated process was killed. Reports/scripts live outside the checkout. Frozen b14 and its checkout/branch were untouched.
