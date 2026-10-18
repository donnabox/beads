# Root execution handoff — existing validation packet

This is a handoff, not an execution release. No indexed file, source, gate, candidate or product process was changed. Current preparation index is `be10d7342a150dda502bba130eb7e999fb82f6c2850657e4f8c1770adb535c4f`; commands are the sole argv/cwd authority. Source review is still of the staged tree, not a committed candidate: frozen tree `630a94d0ca1545f898cb8a8b614d4b7d2dd6d454`, 701 tracked files from `I/source-freeze.json` (SHA `e31f11ce1b35c126e83a2277cd3db00db4142e85686ff874e7a5e8bb4e818ea6`). If code review requires changes, rebind the changed source and review disposition before execution; do not reuse this tree assertion.

Here `I=/Users/dbox/repos/wt/vickie/bdp-integration-logs/pr150-frozen-main-reconciliation`, `V=I/validation`, `E=V/run`, and `W=/Users/dbox/repos/wt/vickie/vickie-pr150-main-reconciliation`. All outputs/consumer paths are outside W. Keep this one packet; root measurements/release records can sit beside it without adding them to its frozen preparation index.

## Before dependencies

After native/actual Claude CODE disposition, root creates the ordinary local merge commit. Verify actual HEAD, tree, two parents **in order**, clean `git status --porcelain`, and all 701 committed/current file hashes against the reviewed freeze. Required parents are d4f3f4be88c7837b71d6646518ca3589d464ba3d then8d6671ed19b0c0406428ccc2d2161768d01e8b9b. The eventual commit is not yet known; do not substitute staged base HEADd4. Keep branch `codex/vickie-pr150-main-reconciliation`.

Write a measured `candidate-source.json` with exactly the gate-required fields (additional audit fields permitted):

```json
{
  "worktree": "/Users/dbox/repos/wt/vickie/vickie-pr150-main-reconciliation",
  "branch": "codex/vickie-pr150-main-reconciliation",
  "head": "ACTUAL_MERGE_COMMIT",
  "tree": "ACTUAL_VERIFIED_TREE",
  "trackedSourceFiles": {"each tracked relative path": "sha256 of actual bytes"}
}
```

Root updates only `execution-gate.json`: `sourceBinding={"path":absoluteCandidateSourcePath,"sha256":actualDocumentSha}`, actual `finalHead`, `finalTree`, `sourceLockSha256`; `preparationIndexSha256="be10d7342a150dda502bba130eb7e999fb82f6c2850657e4f8c1770adb535c4f"`; `expectedParentEvidenceExit=0`. Set `rootExecutionReleased=true` and initially `approvedStages=["dependencies"]` only after acceptance. All document references use absolute path + SHA256 of complete file bytes. Gate verifies the branch/clean source/parents/map on every dispatch and source bytes after each stage. Source identity is SHA256 of Python `json.dumps(trackedSourceFiles, sort_keys=True).encode()`; let existing gate compute it, do not use a differently serialized map hash.

Measured isolated tools are already bound in tools.json: bd1.0.5dev SHA9556f847…, Go1.26.2, Node24.16, pnpm11.20, supervisor Python3.14 and isolated Python venv/dependencies. Provisioning is completed separately; no repeated bootstrap. Root reviews preserved successful tool identities and initial failed Go attempt without relabeling either.

## Single-stage invocation and common acceptance

After the corresponding root release, invoke exactly one stage at a time:

```sh
/opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/guarded-action.py /opt/homebrew/opt/python@3.14/bin/python3.14 /Users/dbox/repos/wt/vickie/bdp-integration-logs/pr150-frozen-main-reconciliation/validation/run.py final STAGE
```

Replace STAGE with the exact identifier below; no shell chaining, retries or altered arguments. The helper executes the concrete indexed argv. For a finished stage inspect `E/STAGE.started.json`, `.log`, `.processes.jsonl`, `.exit.json`; raw exit must0, `accepted=true`, `sourceUnchanged=true`, errors/process survivors/runtime directories empty and `forcedCleanup=false`. Rehash the log against `logSha256`; inspect actual test-owned process/store cleanup, since process-group sampling cannot prove absence of never-observed detached descendants. Runtime/JSON results require semantic review as well as process exit. A cap failure is budget-limited/inconclusive pending diagnosis, not permission to retry/widen or proof of product failure. Existing test/CI deadlines remain unchanged.

The gate enforces all earlier stages in the indexed order. Root extends `approvedStages` only through the next accepted boundary; do not erase prior attempts. The unchanged predecessor packet and native report are under `V/prior-preparation`.

## Release boundaries and measured documents

1. After **dependencies**, inspect frozen-lock install log and unchanged source/lock. Capture actual installed workspace tools into `workspace-tools.json` shaped `{"tools":{"absolute tool path":{"realpath":"actual resolved path","sha256":"actual byte hash"}}}`. Include actual pnpm implementation, Vitest entry/implementation manifests, Biome shim/platform binary, TypeScript entry/platform executable and relevant package manifests used by these commands; an empty map technically passes the helper but is not acceptable root evidence. Use actual resolved files after this install, not prior-worktree identities. Bind `workspaceTools` to this document's absolute path/hash. Every later stage requires it. No old typecheck/build/runtime evidence is reused.
2. Release static commands then **build**, **bind-build**, **full-test** in order. Existing build script cleans then compiles. `bind-build` creates `E/build-artifacts.json` with head/tree/sourceIdentity, actual artifact map and start/exit/log pins. All subsequent stages verify these bytes; no build/archive hash is prefilled in the gate.
3. After **full-test**, inspect actual `E/full-test.json`, raw log and structured observations: zero failures, expected complete test file set, required bd matrix execution, the five affected callbacks and relevant spec-layout/catalog/schema/cumulative cases really passed, no unexplained missing/skipped selected case. Fresh234 private workspace exchanges must be distinguished from inventory234 and external-consumer evidence. Existing spec-layout cases prove the148anchor/relative-link mapping; supplement with reviewed old-path classification, not another checker. Only then append `"full-test"` to `rootResultsAccepted` and release preflight/smoke/evidence/type examples/package prefix.
4. **evidence-verify** must return raw0 and emit the expected verified claim record. Parent baseline is pinned in parent-evidence-baseline.json. A0→2 change is NO_CLAIM and stops this reconciliation even though generic CI permits2; do not reseal/generate evidence. Inspect preflight/smoke actual success/cleanup and Type example result separately; no profile/Types adoption follows.
5. Release package stages through **consumer-install** only. Source cwd for pack, packet cwd for helper modes, and **E/consumer** cwd for install are already fixed. Archives and `E/tar-source-provenance.json` are measured by existing helper. Do not substitute a workspace entry. After install inspect actual manifest/settings/lock and graph provenance prerequisites; set `consumerLockSha256` to actual `E/consumer/pnpm-lock.yaml` SHA. Then release **verify-installed-consumer** alone.
6. After that verification passes, inspect `E/installed-provenance.json` and `E/installed-consumer-inventory.json`. Require sourceCommit/tree/sourceIdentity match actual candidate, both contained entry hashes, exact8-package production graph, archive/member/build correspondence, measured payload identity and lock. Bind `installedProvenance` and `consumerInventory` with absolute path/hash documents. Only then release **packed-affected-1**.
7. After each packed callback, inspect its one passed JSON fullName, raw observations, actual wire/status/context/cursor/receipt behavior, declared sourceCommit, workspace cleanliness, installed entry/payload proof, child closes and store cleanup. Append its identifier to `rootResultsAccepted` before releasing the next callback. Expected fixed counts13/32/8 apply to callbacks3/4/5; derive the two legacy Event observations from their actual unchanged tests, not an invented count. After callback5 record its acceptance as well. External private child plus workspace schema read is the scope; callback1 is a legacy SSE fence, not the repaired schema-read witness. No package export change or duplicate234 packed run.

No final publish/CI/merge authority is implied by local accepted stages. Root separately audits final exact tree/full results and subsequently requests exact-head CI. No main landing, PR cleanup or downstream pin change.

## Indexed stage order and supervisor bounds

| Stage | Outer seconds | TERM at seconds |
|---|---:|---:|
| `dependencies` | 300 | 295 |
| `format-check` | 120 | 115 |
| `lint` | 120 | 115 |
| `check-boundaries` | 120 | 115 |
| `typecheck` | 120 | 115 |
| `build` | 180 | 175 |
| `bind-build` | 120 | 115 |
| `full-test` | 1800 | 1795 |
| `e2e-ready-preflight` | 120 | 115 |
| `smoke` | 120 | 115 |
| `evidence-verify` | 120 | 115 |
| `type-examples` | 120 | 115 |
| `node-version` | 120 | 115 |
| `prepare-package-directory` | 120 | 115 |
| `pack-protocol` | 120 | 115 |
| `pack-server` | 120 | 115 |
| `verify-archives` | 120 | 115 |
| `prepare-consumer` | 120 | 115 |
| `consumer-install` | 120 | 115 |
| `verify-installed-consumer` | 120 | 115 |
| `packed-affected-1` | 120 | 115 |
| `packed-affected-2` | 120 | 115 |
| `packed-affected-3` | 120 | 115 |
| `packed-affected-4` | 120 | 115 |
| `packed-affected-5` | 120 | 115 |

All configured commands have automaticRetries0. The full-test direct argv is node + pinned pnpm exec vitest run --reporter=verbose --reporter=json --outputFile.json=absolute; all five packed output options use equals form. Callback4 uniquely receives transaction-installed.js through its unchanged private wrapper; callbacks1/2/3/5 receive transaction-installed-process.js. The source commit placeholder is resolved by the gate only from the measured merge commit.
