# PR120 W1b: one diagnostic execution

## Outcome and next decision

The single authorized local diagnostic **passed**. It did not reproduce CI's counter0: action36 `owned-metadata` completed real HTTP, the corruption hook ran once, and the returned matrix row failed only assertion `metadata`, exactly as the unchanged outer test requires. There were no preceding failed assertions or execution errors in this observation. **The CI cause remains unexplained. This is not a flake diagnosis, a green full suite, merge clearance, or permission to retry CI.** No second diagnostic execution occurred.

The smallest useful follow-up is to retain bounded, credential-safe returned-row diagnostics on the original counter/assertion failures in CI. That would expose the first prerequisite failure if it recurs. Root separately owns any such source change and future CI decision. This run cannot establish an environmental cause or eliminate a rare serving/transport/harness failure.

## Source and instrumentation

Isolated checkout `/Users/dbox/repos/wt/vickie/vickie-pr120-w1b-diagnostic`, branch `codex/vickie-pr120-w1b-diagnostic`, ordinary base/HEAD `70ac666002708202d2f93195eed90dfe8541300d`. Original frozen70ac checkout was not edited. Only `apps/bdptest/src/read-update-matrix.test.ts` has local observational modifications (42 insertions,2 deletions): lifecycle/report retention, raw/client executor entry logging, safe error retention. Existing test input, mutation, three assertions, selection and all deadlines are unchanged. No commit or outward push was made; the actual report correctly says `sourceDirty:true`. This is instrumented-source diagnostic evidence, not unmodified-source conformance.

Final instrumentation patch SHA `d2a453cd5fe5800d79d4368b1aad5a678b19a01db51daf1584d9cb259207a9c8`; test SHA `73a1e46ce836d4141226341057c850710810e8910329868939c727445e117d07`; preflight SHA `fc5faba773afb6ceb59a34d79974df6c91cd4e61fc9e759624977e197dd5e032`. Guarded-action session check was used before worktree creation, source mutation and execution, requiring session `2026-10-08T20:26:35Z-81e5`. No AGENTS.md was present in this checkout or searched ancestor paths; repository CONTEXT.md was read. Observable-test-runs discipline was applied.

Root caught a pre-execution instrumentation TDZ: the directory record preceded its const declaration. It was moved immediately after mkdtemp. The original patch/preflight/script are preserved under `initial-preflight/`; no test ran on that draft. The corrected instrumentation passed typecheck. `request.url` is the actual request type member and matches existing adjacent assertions. This is a preparation correction, not a serving failure.

## Exactly one case; prerequisite work separately counted

Pinned Node24.16.0/pnpm11.20.0. `BDP_CUMULATIVE_INSTALLED_ENTRY` was absent and explicitly cleared by the supervisor; default fresh emitted `packages/server/dist/installed-reference-process.js` was used. No package archive, installation cohort, network dependency install, broad suite or load simulation occurred.

| Command | Result | Elapsed |
|---|---|---:|
| `pnpm install --offline --frozen-lockfile` | exit0 |1.058s|
| `pnpm build` | exit0 |2.102s|
| `pnpm typecheck` (instrumentation prerequisite) | exit0 |1.050s|
| `pnpm exec vitest run apps/bdptest/src/read-update-matrix.test.ts -t '^detects W1b link-metadata assertion-input corruption after actual HTTP$' --maxWorkers=1 --reporter=verbose --reporter=json` | exit0;1pass,201filtered |3.155s supervisor|

The selected test phase took698ms; Vitest total1.17s. Actual interval2026-10-14T18:49:44.475Z–18:49:47.630Z. Supervisor120s, test30000ms, requests5000ms and existing startup/close10000ms remained unchanged. There was no timeout or forced termination. `.exit.json` files retain exact commands/timestamps; the verbose/JSON logs retain all actual outcomes.

## Returned evidence and workload

`observations.jsonl` retains the genuine report before outer cleanup (`cleanupComplete:false`) and after outer cleanup (`true`), complete selected row/action outcomes, safe synthetic client observations, raw report projection, baseline/configuration/session identities, actual counter and lifecycle events. Known fixture credentials are redacted before persistence. No raw authorization values are retained. The journal is256369B, below the preflight8MiB/record and32MiB aggregate observational envelope.

-6 raw executor entries and30 client executor entries;30 real client returns plus6 raw report exchanges, all writeState complete: **36 HTTP observations**. Client executor entry alone is not proof of socket dispatch; in this actual run every entry has the corresponding completed observation.
-0 fixture setup HTTP,0 native-read HTTP,1 scenario session plus bootstrap.
-Last action `owned-metadata`; counter1. Its status/written/schema/outcome/source/revision/Resource/provenance assertions all pass, with only `metadata` false. Earlier reached assertions all pass.
-The real client observation is captured before assertion-input corruption. `owned-metadata-real-observation.json` preserves the actual written request, status, returned body and synthetic transport observation. This is not a wire-corruption experiment.
-Highest executed client request body361B. Static preflight records all63 possible action inputs and dynamic binding names; opaque revisions/URLs were not invented to premeasure resolved bodies. Actual observed body sizes and routes are in summary.json.
-2 child launches, sequential bootstrap then scenario, under the63HTTP/two-child/two-store ceilings. No retries/replacements.

Runtime ready payload identity `65f7f2f18afa2dc0a760e05162c96d0df08b6785b1530fcc89bec9727d618de6`. Exact configuration, artifact/policy/authentication identities and report manifest/fixture/configuration bindings are in summary.json; source attribution is the supervisor's checkout plus instrumentation hashes, not child-attested git provenance. `emitted-files.json` hashes308 protocol/server emitted members and the actual entry from the explicit fresh build. No tar/packed identity is claimed.

## Cleanup

Children1923 and1924 both exited0, no signals. Independent `kill(pid,0)` checks found both absent; process group1907 has no remaining members. Temporary root `/var/folders/cc/hh74y1f16qj9gfbv8x4x2n6m0000gn/T/bdp-cumulative-ru-SfFoj6`, bootstrap store and scenario store are independently absent. Pre-/post-cleanup snapshots and independent checks are separate evidence; cleanup was not inferred solely from the outer pass. No owned test/build/child process remains.

## Durable local artifacts

- `diagnostic.log`, `diagnostic-vitest.json`, `diagnostic.exit.json`: exact selected execution.
- `observations.jsonl` SHA `0a43c68254b9f930f2c0bcc1df5b0ae13264c3a7c56d4c09ce2e6266bc3db900`: actual retained reports/lifecycle and client observations.
- `summary.json` SHA `83b477a57332e56187fd90831e296915bd6ef0c8fecfca1cc0654964e8d614b9`: source/entry/configuration/payload, result, routes/body sizes and cleanup.
- `cleanup.json` SHA `fce18a5bbcd2901e6620717441ad66112aabfa5f707cfd216f0dd4dd07647b6c`: independent PID/group/store absence.
- `instrumentation.patch`, `preflight.json`, `selected-row.json`, `emitted-files.json`, prerequisite logs and exit records.
- `supervise.py` records process groups every second, enforces120s, and would terminate only its own process group on timeout. No timeout cleanup was needed.

The original failing CI evidence and frozen70ac source remain untouched. Stop condition fulfilled: one observation, retained evidence, actual cleanup, no automatic retry.
