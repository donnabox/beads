# Retained Problem equality closure

Frozen clean **965f27cd1268c390d7b7c9e6302b0e1f8a76e4eb**, ordinary successor of **29e715162691e6d0a4bcc1ce699f33e243ebecdd**. Checkout `/Users/dbox/repos/wt/vickie/vickie-read-update-retry-closure`, branch `codex/vickie-read-update-retry-closure`, GitHub origin. Original29e7151 checkout/report remain unchanged. No outward writes.

## Finding and repair

The parent tests enumerated selected Problem fields and asserted detail absent, but the normative Problem schema permits additional standard and extension fields. Thus title/instance/arbitrary-extension corruption could pass while the tests described complete retained-disposition equality. This was a test evidence defect, not a demonstrated server behavior defect.

Three-file repair adds **13 complete deterministic Problem-object comparisons** across the two new retry rows: initial failures, retained replays, failed singleton/direct responses, sequence members with their current operationIndex and the supporting fixed absence/conflict Problems. These literal expected objects reflect the known fixture response, including its four non-positional fields. They complement the field assertions; they do not narrow the normative schema. A generally valid BDP Problem may have title/instance/extensions, but these original fixture failures have none, so an asymmetric addition or loss must not pass retained equality.

Five new post-wire controls check:

- cross-carrier retained failure gains title;
- cross-carrier retained failure gains arbitrary extension;
- reverse-carrier retained failure gains instance;
- reordered sequence failure gains arbitrary extension;
- initial sequence failure gains title before replay.

All use the existing assertion-input corruption layer after actual wire recording, apply once and fail the named immediate whole-object assertion. Initial and replay objects are both checked, closing the asymmetric initial-only case. No runner/executor/runtime/schema changes or new control framework. Fixture file, route allowlist and both route-digest pins are unchanged. Coverage binding lists now include the complete-object assertions; both rows still partial. **129 required /45 authored /26 whole /103 uncovered**, claimEligible:false.

Changed files: `apps/bdptest/src/read-update-matrix.test.ts`, `packages/conformance/matrices/read-update-v1.json`, `packages/conformance/test-support/read-update-coverage-v1.json`. Delta218 insertions/two deletions. Existing original limits and excluded disconnect/expiry/auth/other-kind variants remain unchanged.

## Red/green evidence

Pinned Node24.16.0/pnpm11.20.0, frozen install and build before emitted tests. Logs and explicit exit records are in this directory.

- Original manifest, new controls: **five tests fail** (`red-all-five.log`, exit1). Four initial/replay field additions leave their row passing; the reordered extension was caught only by a later repeated-array equality, so it lacked the intended immediate failed-disposition oracle. The earlier four-control red is also retained. Original and fixed manifest bytes are saved separately; the fixed bytes were restored and verified before continuing.
- Fixed focused run: **20 tests pass, two files**, comprising both positive retry rows, all18 sensitivity controls and cumulative coverage/order bindings (`focus-final.log`, exit0). No skipped test is counted as a pass.
- **Fresh packed131 tests pass, four files**, exact clean965f27c (`packed/matrix.json`, exit0). All45 scenario sessions pass;781 wire observations,44 setup and21 native Read observations. Packed aggregate12.379s, well within unchanged90s/120s; full packed test process65.530s. `packed-summary.json` preserves actual source, payload, contracts, policy/authentication, fixture and exact-route bindings.
- Typecheck, boundaries, formatting (including explicit4MiB manifest check) and lint exit0. Lint62 warnings/one info, no autofix; includes fixed-row non-null assertions inherited/added in the control table. No warning-free claim.

The packed runtime payload remains `cac59fd75a1e406211f4c39fe77e70ed1cfb9c1c4e5d625e1f93e0bc015511c5`; contracts remain `48667497d6e13064316d0364cf8f49dfc247ab9189306985a357d7b558333dfe`; fixture and exact route digest are unchanged from29e7151. Manifest source bytes and source commit changed and are freshly recorded. No old observation is relabeled.

Full repository suite and additional smokes were not repeated for this narrow assertion-only repair; parent smoke/evidence results remain historical to29e7151. Fresh packing/install and all installed observation tests were rerun here. No timeout, stopped run, forced cleanup or surviving test child. Final source tree clean; disposable test stores closed through existing cleanup.

This closes the identified native P2 oracle gap and its initial-only variant. Root owns independent review and integration. No qualification/public profile claim is made.
