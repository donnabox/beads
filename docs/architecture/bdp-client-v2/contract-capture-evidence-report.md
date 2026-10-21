# Capture evidence LOW follow-up

Frozen clean head **8da9a30ffa586d4a12362295758645118b2ed7e5**, direct parent **4d13f4104764d2a4757909c7ee48bdba97789374**.
Checkout `/Users/dbox/repos/wt/vickie/vickie-contract-capture-evidence`; branch `codex/vickie-contract-capture-evidence`; GitHub origin. Local only, no outward writes.

External three-frozen-followups review F10/F11 closed with two test-file changes:

- Both the shared compiler and Transactional installation capture assertions zero every supplied schema and descriptor buffer, not just index0. Existing exact capture/freeze and offline reconstruction assertions remain.
- Actual emitted-runtime Type test installs descriptor artifacts in reverse order, explicitly sorts expected summaries by canonical id, asserts input and expected order differ, and compares the exact inventory before and after ordinary restart. Full Type GET/HEAD and durable mutation replay assertions remain. This tests ordering rather than deriving an accidentally correct expectation from input order.

No fixture bundle function, URL constants, historical identity literal, predecessor probe or frozen checkout changed. Compared the shared compiler test's bundle body to exact frozen fc75 source: byte-identical. Reversing one runtime test's supplied descriptor array does not alter the historical compiler-identity fixture. No runtime/service/store/schema/format/export/evidence changes.

Pinned Node24.16.0/pnpm11.20.0 frozen install and build succeeded before emitted tests. **28/28 tests,2 files,1.69s,maxWorkers2** passed (`tests.json`, `tests.log`). Typecheck, boundaries, format and lint pass; lint retains34 warnings/one info and no errors. No broad suite or package smoke rerun for assertions-only changes. No stopped/hung runs or remaining test processes.

G1 remains plan-only awaiting reviewed predecessor/council. This follow-up makes no profile claim.
