# PR91 current-main integration

Frozen source **`b919ab37326e259ff30304ab0dd642562a21ea40`**, branch `codex/vickie-artifact-set-main-integration`, checkout `/Users/dbox/repos/wt/vickie/vickie-artifact-set-main-integration`.

Ordinary conflict-free merge of accepted publication `fe5b49fb78ce8bc01fb6bdb1d7a416b03ce3aed4` and remote-verified main `4fcf36a2f6c4144532e90688eca7f056bf67170b`. Merged tree `b15512f20e3294f0b49ea815b9f94d7ac31d0051`.

Exactly three files differ from main, each byte-identical to accepted fe5:

- `packages/conformance/src/read-update-artifact-set.test.ts`
- `packages/conformance/test-support/profile-contract-plan.ts`
- `packages/conformance/test-support/read-update-artifact-set.ts`

All other files, including runtime, application evidence, catalogs, schemas, public APIs, and historical Read evidence, equal main. No semantic edits or conflict resolution were needed. Source remained clean and immutable during validation.

Node 24.16.0 and pnpm 11.20.0 were pinned. Frozen dependency install, pre-build typecheck, clean build, boundaries, format, lint, and Read evidence verification all exited 0. Lint reported 195 warnings and 8 informational diagnostics; no fixes were applied.

The six conformance files passed **417 tests, zero failures, zero skips**, in 2.81 seconds under supervision: artifact set, artifact bundle, cumulative artifact, profile-contract planner, runner, and transactional cumulative tests. [regressions.json](regressions.json) retains machine-readable results; [results.json](results.json) records every stage and actual exit. [run.py](run.py) retains exact commands and a bounded process-group supervisor.

The unchanged Read seal verified 89 rows, projection `bd7da0fc3f8082e0bfb6f4969c04409f2a610e12408dbc6a0bd14b4b351127cc`, historical run head `e5f7d0c43c151426350ceededdc77fe7e372db9b`. It was not regenerated or presented as a new profile qualification.

No full suite or installed execution was repeated for this three-file private conformance delta, as requested. The previous fe5 CI outcome remains historical and is not relabeled as validation of this merge. No public claims or admission changes were made by this integration. No owned child processes remain, and no outward writes occurred. Root owns draft publication and exact-source CI.

Exact parents, changed-file preservation, and validation summary are in [source.json](source.json).
