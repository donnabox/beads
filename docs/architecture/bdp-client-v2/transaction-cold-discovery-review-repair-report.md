# C3 callback receiver repair

## Frozen source

- Commit **3f486c245bfd31f41fd42f5acc8e0f13683bad42**, tree **2a68550d01df1cf829b9fd76183fa879e5f4acbe**, ordinary parent **b20c92914b48e66b33b66e9e7434dbc3467fad10**.
- Checkout `/Users/dbox/repos/wt/vickie/vickie-tx-discovery-repair`; branch `codex/vickie-tx-discovery-repair`; clean. Agent-Signature trailer included.
- Exactly two changed files: `packages/server/src/transaction-service.ts` and `packages/server/src/transaction-service.test.ts`. Per-file bindings in `source.json`.
- Original b20 checkout/source remains untouched. Its execution is historical, not relabeled. No C4 limits/activation, schema/runtime/store/export/public-admission change.

## Finding and repair

Accepted native P2 from `../transaction-cold-discovery-native-code-review.md` and the differential `../transaction-cold-discovery-native-probes/pagination-receiver-probe.mjs`: C3's frozen pagination copy changed the callback receiver. A lawful stateful clock could return503 rather than a collection because `this.ticks++` targeted a frozen clone. The token callback had the same receiver risk.

The repair captures the original pagination object and spreads its selected values once. It reads each captured callback from that plain copy once and, for callable values, binds it to the original pagination receiver before freezing the copy. Invalid callback values remain untouched for the existing authority validation and its error. Frozen scalar/nested page/Selector/TTL captures remain intact; no callback getter is re-read.

The permanent real-owner test uses getter-returned stateful clock/token methods, checks original receiver identity and state advancement, follows an actual continuation, and compares all returned Resources with the independently created receipt Resources. Getter counts remain one across construction and pagination. Two invalid-callback cases retain the `pagination callbacks required` failure and one getter read. Existing nested/outer limit mutation tests stay unchanged and pass.

## Red/green and checks

Pinned Node24.16.0 / pnpm11.20.0. Full commands/exits/times in `exits.json`; verbose and JSON logs retained.

- RED `receiver-red`: b20 production with new test, exit1 as expected; **1 failed,2 passed,41 intentionally filtered**. Stateful pagination returned a problem instead of success. Invalid callback validation cases already passed. This red result is preserved and is not a product pass.
- GREEN `affected`: **202/202**, five files,2.30s Vitest duration. Includes actual service, Read, HTTP, discovery builder and protocol parser tests. State receiver and existing immutable-limit cases pass.
- Frozen clean build, typecheck, dependency boundaries, format, lint and unchanged Read evidence seal: exit0. Lint exit0 is not a claim of warning-free repository code.
- No full workspace run or redundant broad smoke repetition. Prior b20 smoke evidence stays explicitly historical.

## Fresh exact-source installed observations

`pack.py` asserts clean exact HEAD before/after clean build, packs protocol/server, installs in a new retained offline install, and runs the unchanged installed fixture with explicit entry and source metadata.

- TX **1 passing test**,12 actual exchanges/one child,0.522s runner and0.219s case. Six authenticated discovery refusals500; three anonymous refusals401; batch200, independent afterstateGET200, retained retry200. Discovery remains closed.45s cohort/60s outer/5s exchange/10s launcher; ceiling16/1 unchanged.
- RU **2 passing tests**,200 intentionally unselected,2.33s runner. Actual installed retained-retry and reference/category positive scenarios on the same changed payload.
- New installed payload `5e8ce6bf6d12267a5ea999e4dee84fde15c1c820b683e88a9e712093db0c351f`.
- Contracts `184330942591336c6ad84ef3d32e2704f67006b8ff64882a300488ab4698608a`; runtime configuration `95037bb10187ada959b24f03f1677070780c28108e8d94d5c6d091cd0e5ea977`.
- Authentication `90ddf6d71ca6d60ab09ecb97e9217c8fc3011ab036df035d53487a14547df736`; policy `48630a72476a2daf88a28f5bc6a8a4d2634b8f60b7300e961460ca8b195765eb`; fixture `3a7899ced771d4880066d98b0000c53f801ece53458502cc4f3bc641146c9a85`.
- Exact tarball SHA256s, per-response wire/body hashes, route mappings, ready-derived binding and source audit are in `summary.json`. Source is launcher-declared metadata independently checked against frozen clean HEAD, not target attestation. All claimEligible flags remainfalse.

Packed PID61802 exited0 and was independently absent after return. Fixture removed its disposable store; retained install/tarballs are evidence only. No active owned child/build/test remains. `cleanup.json` records the check.

This closes the author repair for native P2, pending root's independent closure and external C3 review dispositions. No claim that C3 or the profile is now ready to expose publicly. Recorded 2026-10-14T09:35:31.460001+00:00.

Guarded repair-branch push and independent exact remote ref verification completed; see `session-before-push.json` and `remote-verification.json`. No PR/main/bus/launch changes.
