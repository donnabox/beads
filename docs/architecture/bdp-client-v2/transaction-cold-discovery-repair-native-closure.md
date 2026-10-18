# C3 callback repair — native closure

No remaining actionable defect in the bounded repair. The original P2 callback-receiver regression is closed at **3f486c245bfd31f41fd42f5acc8e0f13683bad42**, tree **2a68550d01df1cf829b9fd76183fa879e5f4acbe**, direct parent **b20c92914b48e66b33b66e9e7434dbc3467fad10**. Reviewed both changed files in full; author source remains clean and untouched. This closes the native repair finding only; external council/root readiness and C4 activation remain separate.

## Code and regression coverage

`transaction-service.ts` now captures the original pagination receiver and spreads its properties once. It obtains clock/token callbacks from that captured plain copy, binds callable values to the original receiver, then freezes the copied configuration. The later AuthorityRead `.bind(suppliedPagination)` cannot replace an already bound function's receiver. Nested page/Selector/TTL capture and validation are unchanged. Mutable scalar limits still cannot become authority after construction.

Invalid non-function callback values are not wrapped into callable functions; they pass unchanged to existing `pagination callbacks required` validation. Getter evaluation remains once because the later accesses are to the plain copied properties, not the caller object. The permanent test proves both original receiver identities, advancing clock/token state, actual first-page/continuation completion and all Resources equal to independently created receipt Resources. Separate clock/token null-getter controls preserve one read and the exact existing error. Existing mutable nested/outer configuration tests are unchanged.

## Independently executed proof

Command: `/Users/dbox/.nvm/versions/node/v24.16.0/bin/node /Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-cold-discovery-native-probes/pagination-receiver-closure.mjs` — **exit0**.

The retained original real-owner differential now runs three variants with the same stateful `clock() { return this.ticks++; }` and same empty collection request:

-  Exact-parent service variant: successful collection, original ticks advance1000→1001.
-  Original b20 C3 variant: observed503 `temporarily-unavailable`, original ticks remain1000.
-  Exact repaired installed service variant: successful collection, original ticks advance1000→1001.

The script asserts these outcomes. Parent service source equivalence to the retained compiled0faf module was established in the original review; other modules in that differential are held constant. Repaired variant uses the author's actual new retained installed module tree. This is bounded service behavior proof, not another installed HTTP cohort. All three SQLite owners close and temporary directories are removed in finally; no child/server/network is created. Logs and disposable module copies remain outside source checkouts under `transaction-cold-discovery-native-probes/`.

## Evidence independently inspected and hashed

Verified clean exact repair HEAD, the two source file byte/hash bindings, and both retained package hashes. Inspected actual author JSON reports and exits rather than relying only on report prose:

- 202/202 affected tests, no filtered cases.
- Actual frozen packed TX:1 pass,12 completed exchanges/one child; actual route statuses500×6,401×3,200×3; claimEligiblefalse.
- Actual frozen packed RU:2 pass,200 deliberately filtered.
- Typecheck/build/boundaries/format/lint/Read evidence verification exit0; original expected receiver red is retained separately.

No duplicate packed cohort or broad suite was run by this reviewer. Author child PID61802 exited0 and was independently checked absent. Retained installed payload is `5e8ce6bf6d12267a5ea999e4dee84fde15c1c820b683e88a9e712093db0c351f`; source metadata remains launcher-declared plus external frozen-source binding, not target attestation.

SHA256s:

- Author report `b366f2c8cd235969f69c36ca79e77138a6d4b3296b9661169580de519cb98a8a`.
- source.json `2cd5a2eb8d42992b3a2f3b626b3b843cdc5ef2b915a94d26de348cd1dcd36d7a`.
- summary.json `1c410541bac64d4e1af948120c57492a52e99a4accf110b7a592651d7fefbfed`.
- exits.json `028ce18eff4ae5cccf51f754faf303a7971a3e63e5f42ca6a24e64fcfccd2e0b`.
- cleanup.json `9e6317b7a5bd14b8dac2234b735f6e09e9a7abbd9efcfd32bdaab50b634352cb`.
- protocol.tgz `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227`.
- server.tgz `92ab7c6036a6c45009c99bd6045f772f27067839207177492661d1b2af75b1dc`.

Original b20 source/report/probes remain preserved. No source writes, production changes, outward actions or owned processes from this review. C3 still assembles an internal candidate only; no public or full-profile qualification follows.
