# H1a private Transactional carrier groundwork

Frozen source **87423d5ba52d7a3167c43dabc574cbf40c4c928d**, direct successor of **d907a1e1576c5795c17edb5f15528f47edfd7d6b**. Isolated checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-harness-carrier`, branch `codex/vickie-transaction-harness-carrier`, clean status. Origin verified/set to `https://github.com/gastownhall/bdp.git`; no fetch/push/PR/bus mutation. Full source diff: `transaction-h1a-carrier.patch` beside this report. Eight files,868 insertions/10 deletions; most additions are the dedicated tests and formatted source-only fixture/manifest/coverage JSON.

## Implemented boundary

- New `createCumulativeTransactionalArtifactBundle` requires the existing canonical factory-loaded baseline TX/no-History selection. It reuses source copying, parsers, fixture binding and canonical exact-file catalog digest semantics; separate WeakSet brand identifies private TX bundles. TX requires v2/transactional-v1 and exact selected row/profile agreement. Structural/shrunken selections, wrong factories, History, retired/misprofiled rows and v1 are refused.
- `runCumulativeTransactionalMatrix` reaches the same execution core through an internal mode argument, not a public option. It requires the private TX brand and declared TX profile. Ordinary `runConformanceMatrix` rejects that brand regardless of declared profile: Read, Read+Update and Transactional all fail before fixture preparation or any executor call. This closes revised Claude F7's genuine-bundle/misdeclared-RU path.
- Exact Scope/profile configuration remains required. **The existing exact executor still rejects Transactional**, so this slice cannot run a target and has no synthetic fallback. Tests prove absent exact mode and a hostile cast of a TX exact mode fail without preparation/dialing. Actual matching TX exact-mode execution and discriminating canonical-Scope mismatches remain future H1b controls, after its separately reviewed transport admission.
- The planner gains only the baseline TX/v2 carrier branch plus an explicit structural-readiness caveat. Its existing TX1MiB limit was not widened. The factory now grants4MiB only to cumulative RU, leaving TX at1MiB; exact boundary/one-byte-over tests exercise both factory and planner. Fixture limits remain1MiB. Public/root exports, ordinary Read and diagnostic RU admission, public matrices/catalog inventories, and HTTP route/credential rules are unchanged.

## Source-only authored material

`test-support/transactional/{manifest,fixture,coverage}-v1.json` remains outside public matrices/catalogs and published package files. Manifest11033bytes; fixture549; coverage2465. Four catalog-ordered partial scaffolds cover atomic-commit, rollback, receipt.readable and failed-retained. They contain only initial authored carriers/assertions; the receipt client action is a declared future H1b operation, not an implemented adapter. These do **not** implement the complete four-row observation plan. Coverage says authored-unexecuted, transport pending, claimEligible:false and wholeContract:false for every row, with explicit missing raw-wire/schema/poststate/negative controls. H1b must complete those row-local observations before execution/credit; no happy-path status assertion is being presented as a whole contract.

The one canonical selection/planner computes **250 required,4 structurally ready-unexecuted,246 unplanned,250 unobserved,0 whole/qualified**. Coverage JSON records authored bindings only rather than mirroring a second250-ID authority; the test checks authored ordering/bindings and the complete planner denominator. Fixture capabilities are declared setup requirements, not observed target capabilities. Test-only configuration digest strings are shape/negative-control inputs, never emitted target evidence.

Artifact SHA-256 values:
- manifest07ba57c264a069c557027a73894947ddd4544048d9ac4f2bc44e7c8c946befed
- fixture9577eb72d5b916a4d4e281d869451bc9c7fd1bd2fc6123494553331c08c4a4db
- coverage8aa3267475e153de98e075841c4222588fa8e6772fc63a7a75d0d47e4aec9e7d

## Plan disposition and ownership

Read revised plan SHA5cc7106 and revised Claude review. Per root authorization, changed only the plan's H1a basis/ready wording and appended an H1a disposition; M1/H1b clauses were preserved. Resulting plan SHA-25661d837b8fad1d0e9c1d4e61de6a0e0193d29a1dcc87bb797fae047ab9e826f88. Accepted F7 with genuine-TX/RU-declared rejection and unchanged-RU4MiB control; F1 structural-ready wording explicitly does not claim transport admission. Existing TX planner cap needed no edit. Root owns other revised findings/dispositions. No M1/runtime/store/HTTP executor/schema/public catalog/public matrix edits.

## Verification

Node24.16.0/pnpm11.20.0. Frozen install PASS. Initial nine new tests PASS; the first clean pre-build typecheck then found a test-only executor-spy variance mismatch. Corrected the mock's type rather than casting the positive exact execution path; preserved red log `h1a-build.log`. Final **clean→pre-build typecheck→build PASS**, `h1a-final-build.log`.

**806tests/25files PASS8.08s** for all `packages/conformance/src` with maxWorkers1, verbose plus JSON (`h1a-conformance-tests.log/.json`). Includes nine new H1a tests plus existing ordinary runner, Read/RU artifact and inventory guards, canonical selection/planning and transport regressions. No runtime/installed TX execution is claimed.

Dependency boundaries, full format check, lint (45 warnings/1 info, no errors), unchanged89-row Read evidence verification and git diff whitespace PASS (`h1a-static.log`). No full repository suite or packed TX observation was run; no evidence artifact/constant resealed. All command sessions completed. Frozen checkout remains clean.

## Next blockers

M1 must supply reviewed actual consumed runtime metadata, per-credential permission identity and configured canonical Host support on root's chosen integration. H1b needs revised Claude F1's narrowly reviewed exact TX transport-profile admission, F2's actual private dynamic-route observation recorder, and finished row-local wire/schema/independent-poststate/replay/HEAD/reopen controls. Its dynamic receipt digest must remain separate from the base route digest, with original raw observations and credential redaction. Collection paging/grammar and emitted deep-import inventory refinements belong there. No Snapshot, cold discovery, retention advertisement, public H2 admission or profile qualification follows from H1a.
