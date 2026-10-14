# Sequence refusal context — measured draft preflight

**Unexecuted draft; awaiting root measured execution release.** Accepted plan v2 SHAea8f0a3b9decc2613c45a8966fc49db8f94cdbdbae782357073676583cbc5b36, dispositions10327a99 and native closure7720d010; root authority is `transaction-sequence-refusal-root-plan-acceptance.md`. Supporting work remains under the sole fork design plan. No claim that source/tests compile or pass.

Checkout `/Users/dbox/repos/wt/vickie/vickie-sequence-refusal-context`, branch `codex/vickie-sequence-refusal-context`, uncommitted ordinary descendant draft at verified assigned main `d50f554cf6418ba036072ebf6795cb272cd6cbc1`, tree `7544b53aa8aa6d7edae94bd6bf0a4d157ee55985`. Created with local Git objects; original PR117 worktree untouched. Six owned files only. No source commit, push, PR/main/coordination, runtime/config/export/discovery/schema/parser/store/evaluator changes.

## Draft behavior and authority boundaries

`sequenceRefusalObserved()` registers existing finite service work through run(), enters owner.withRead synchronously once, returns the same frozen context inside a frozen object, and retains entry refusal causes. No members execute, no keys are admitted, no timer/minimum/wait semantics are introduced. HTTP invokes it only after existing authorization at the sequence over-count413 or supported-carrier unacceptable-media406 branches. Count still wins over unacceptable Accept; unsupported controls/composition remain in their old places.

Service tests draft real owner identity/entry-count/immutability, unchanged state and no sequence/direct admission, later position and closed/reentrant failures. HTTP adapter tests draft zero observations and absent fields for carrier-key400, malformed400,401,403,415 and the existing minimum development500. An unhealthy owner may still produce sanitized bodyless500; no stale context fallback is introduced. Static inspection preserves runtime closing/not-ready503 before handler dispatch (`transaction-runtime.ts:115–118`), without claiming that prevents all live owner faults. The runtime file is unchanged.

New HTTP tests seed genuine commits and test both refusals, repeated after a third commit. The first status is parameterized413/406 so each omission/staleness control can reach its intended first assertion separately. The source fixture adds a **test-local final** read-only principal/key/receipt inventory; it closes runtime only after all live graph/Event/root observations. Exact expected seed receipts and rejected-key absence are separate from graph inspection. A separate fixture tests different-identity reuse and only the subsequently admitted receipt rows. No production inspection seam was added.

Mandatory progress prose scopes the old HEAD/304/412/406 sentence to fresh Snapshot creation, distinguishes successful sequence post-runner context from nonexecuted refusals and discloses observation fault500. Its added section explicitly states draft/unvalidated status. No unresolved minimum/Event/product question is resolved by this change.

## Actual draft footprint

828 added/6 removed lines across six files, including the previously untracked installed test. `git diff --stat` alone excludes that file; `source.diff` and the machine bounds include it.

|File|Bytes|Delta|SHA256|
|---|---:|---|---|
|`packages/server/src/transaction-service.ts`|24395|+12/−0|`6e1213f5546f8e910f61d9573a4b4486418f754014663aa067339715da9a214b`|
|`packages/server/src/transaction-service.test.ts`|78316|+133/−0|`9174642351432941fe941cd6f9843c92722f359921b5430f62d5ff175555ddd7`|
|`packages/server/src/transaction-http.ts`|27010|+11/−3|`0859e2291ce335b482e9120b57f81387b1e8e7c1c7c4ca933eae00b54d2b09b1`|
|`packages/server/src/transaction-http.test.ts`|110825|+150/−0|`0b9886ff0c7b5f411a59886ddc14b912b1ce4ed9700f538d454270c54e0bedc0`|
|`apps/bdptest/src/transactional-sequence-refusal-installed.test.ts`|19327|+501/−0|`7861e55113e3e3bbf6621e526f2836a638e2bbcffafd2b321cc98e40846f1ba2`|
|`docs/design/transaction-implementation-progress.md`|116687|+21/−3|`dadcde5f49535704fdf34dc946701682d09cbd8e10cdc107c29eca848181dde0`|

Entire service prefix containing discoveryObserved remains byte-identical; the existing direct-submission HTTP suffix is byte-identical. Existing direct installed test hash is `c817729792a8bd97473b4b7ebe06474d4a0c27bcd5122e58268b997aa03408b0`. Service tests append only; HTTP tests append cases and add the permitted test-local inventory method, retaining original test bodies. No other source paths changed. Schema definition names used by the new installed proof were statically checked present, not executed/validated through AJV yet.

## Actual proposed installed ledger

Four exact static method/URL targets: singleton create, sequence, current Beads `?limit=32`, Scope Events. One additional GET target is the actual validated seed receipt root; it is never guessed. The local mirror requires canonical exact href, exact Scope receipts/ prefix, a nonempty single [A-Za-z0-9_-]+ suffix, no query/fragment/userinfo, and at most8192 characters/UTF-8 bytes. Thus descendants or query variants cannot gain route authority from an absolute-URL receipt schema alone. Route preimage and digest are in `preflight-bounds.json`; digest `f7a455537dfef750506fcbdaec8b858a4d1484f37a56693639ced38f5fb24d1c`. Largest static URL is53 bytes. Dynamic receipt target remains subject to actual transport/runtime bounds.

|#|Exchange|Expected status|Request body bytes|
|---|---|---:|---:|
|1|seed-a|200|94|
|2|seed-b|200|94|
|3|initial-capacity|413|5497|
|4|initial-negotiation|406|176|
|5|retained-root|200|0|
|6|unchanged-beads|200|0|
|7|unchanged-events|200|0|
|8|reuse-negotiation-key|200|104|
|9|reuse-capacity-key|200|94|
|10|repeat-capacity|413|5728|
|11|repeat-negotiation|406|183|
|12|final-beads|200|0|
|13|final-events|200|0|

Both33-member carriers are valid keyed operations; the larger repeated-key form is **5728 bytes** and the positive request bodies total **11970 bytes**. Python arithmetic reproduces ASCII JSON.stringify order/spacing from the authored builders; each body/preimage/hash is retained. Bodies omit the carrier-level key for sequences. The first negative refusal verifies actual413, schema, code=limit-exceeded and limit=sequence.operations **before** the named context-presence assertion.

Seeds a/b pin pos-1/2. First refusals reportpos-2; receipt body remains the seed while headers are current. Graph/Event pages show only a/b. Reuse of the406 key changes c's title from original-c to corrected-c; reuse of the first413 key creates distinct d. Both repeated fresh-key refusals must reportpos-4 before the final exact a/b/c/d Resource and Event observations.

Proposed author execution: new positive13/one child + unchanged direct-refusal positive14/one + separately packed assigned-parent d50 negative expected3/one. **30 expected requests/three children, maximum40/three**, no replacement without disposition. Ordinary whole-suite collection would add only the new positive13/one; the predecessor negative is separately selected. No tests/services or installed children have run in this draft stage.

Unchanged transport bounds: request65536, header16384, response131072 bytes; runtime target8192 bytes is separate. Five-second exchange,90-second cohort,120-second outer, existing10-second child-close bound. No cap/deadline changes. Response planning estimates: receipt8192, Problem4096, complete four-Resource collection16384, four-Event page32768 bytes; these are estimates, not observed sizes.

## Isolated source controls (unexecuted)

- `omitted-context` (`8c25fc2d4f2043bd0be74bcb427a0259bb4a1101fa8e1cf9d5ad4e37ceeff0c9`): `observes sequence refusal context with 413 first`, `observes sequence refusal context with 406 first`; expected `sequence refusal context presence: 413 bdp-scope-epoch`, `sequence refusal context presence: 406 bdp-scope-epoch`.
- `stale-context` (`1ad41f0c743e78786a6dfad9175e395c2f896f143829bb6da0fccb1d781a8e74`): `observes sequence refusal context with 413 first`, `observes sequence refusal context with 406 first`; expected `sequence refusal current context: 413 bdp-scope-position`, `sequence refusal current context: 406 bdp-scope-position`.
- `over-wrapped-exclusion` (`5ad8571498859a178e33702a5a60d297965bbef804d49984fd14ba5e1bf16b4b`): `keeps excluded sequence responses unobserved and without refusal context`; expected `excluded sequence refusal malformed-request context: bdp-scope-epoch`.

Each patch changes only the derivative HTTP production file in an isolated checkout. Omission retains the actual observation call while omitting its wrapper. Staleness substitutes only the position in the actual tuple. The exclusion mutant adds header fields to refused sequence responses; the named absent-header assertion precedes zero-observation assertion and must be its discriminator. No fifth installed cohort or substitute implementation is used. Five individually selected test runs across these three derivatives; no reused failing earlier assertion may conceal the second refusal path.

## Exact proposed commands after release

All commands run with explicit implementation or derivative cwd, PATH prefix `/Users/dbox/.nvm/versions/node/v24.16.0/bin`, and pnpm `/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs`. An external supervisor retains command/start/end/exit, bounded process-group cleanup and structured Vitest output; no unfiltered transactional-matrix/full-workspace test command.

```sh
node "$PNPM" typecheck
node "$PNPM" build
node "$PNPM" exec vitest run packages/server/src/transaction-service.test.ts packages/server/src/transaction-http.test.ts packages/server/src/transaction-sequence.test.ts packages/server/src/transaction-submission.test.ts packages/server/src/transaction-store.test.ts packages/server/src/transaction-dependencies.test.ts packages/server/src/transaction-sequence-alias.test.ts --maxWorkers=2 --reporter=verbose --reporter=json --outputFile.json="$LOG/source.json"
node "$PNPM" format:check
node "$PNPM" lint
node "$PNPM" check:boundaries
node "$PNPM" evidence:verify
git diff --check
```

In each isolated derivative, build before the emitted HTTP consumer and run its exact file/selector from the machine control table with maxWorkers1, verbose+JSON output. Expected nonzero must match the intended assertion, not a compiler/setup/deadline failure. Restore/preserve derivative evidence separately; production candidate remains unchanged.

After successful source gates, freeze a clean ordinary Agent-Signature commit and notify root for council. Freshly build/pack protocol+server and install into a disposable directory with offline/no-scripts; pin archives, full file inventories, actual installed entry, source/tree and actual ready payload/contracts/config/policy/auth bindings. Expected source is launcher metadata, never inferred solely from child claims or an unchanged entry module. Fresh parent negative packages come from separate clean **d50** source, not stale8c, under the repaired harness's explicit installed-entry/source override.

```sh
node "$PNPM" exec vitest run apps/bdptest/src/transactional-sequence-refusal-installed.test.ts --maxWorkers=1 --reporter=verbose --reporter=json --outputFile.json="$LOG/sequence-positive.json"
node "$PNPM" exec vitest run apps/bdptest/src/transactional-direct-refusal-installed.test.ts --maxWorkers=1 --reporter=verbose --reporter=json --outputFile.json="$LOG/direct-regression.json"
# Same new test with separately packed d50 entry/source: expected named failure after3 requests.
node "$PNPM" exec vitest run apps/bdptest/src/transactional-sequence-refusal-installed.test.ts --maxWorkers=1 --reporter=verbose --reporter=json --outputFile.json="$LOG/sequence-parent-negative.json"
```

Do not execute these commands until root accepts this measured preflight. Every attempt, including setup failures, consumes its applicable child/request allocation. An unexpected installed failure stops for disposition, not an automatic replacement. Complete wire/body assertions occur in memory; persisted observations explicitly retain fingerprints/context only, not full bytes. Cleanup preserves the original failure, records cleanup failures separately, confirms child exit/PID absence and directory removal, and logs the observation even after cleanup failure.

## Draft work actually performed

Local worktree creation, source/Markdown edits, offline frozen dependency setup and Biome formatting only. No tests, typecheck, build, services, package install cohort, source-control execution or network/outward actions. Static source inspection corrected draft handler-body and inspection-position expectations before execution; no failed-test evidence is invented. No applicable AGENTS.md was found in this checkout or inspected repo ancestors. Existing observable-test discipline will apply when execution is released.

Stage1 stops here, with uncommitted source preserved. Root must review actual source/preflight hashes and authorize execution. This is not code-council closure, CI readiness, installed evidence or profile qualification.


## Receipt-root preflight correction disposition

Root/native identified that the original prefix+origin check admitted receipt descendants or query variants. Accepted correction locally mirrors `packages/conformance/src/transactional-observation.ts:58–78` without importing/exporting a shared helper or widening owned files. The byte bound supplements the existing length convention; all accepted suffix characters and the configured prefix are ASCII, so the valid grammar has identical character/byte length. The raw executor still uses the single observed, validated root as exact GET authority.

A source-only fixture test in the new installed-test file authors two valid roots (including exact8192 bound) and20 malformed inputs: nonstrings, invalid URL, empty suffix, descendant/trailing slash, query and fragment including empty delimiters, encoded/dotted/Unicode suffix, dot-segment/whitespace normalization, userinfo, foreign Scope/host, noncanonical host spelling and8193 length. These cases are **authored but unexecuted**; no installed calls were added. The future file run will include this static fixture plus its one installed scenario. No execution release is inferred.

The original preflight/bounds/diff, measurement script and installed-test draft are preserved under `stage1-original/`. The successor changes only the new installed test; five other file hashes, carrier bodies, request counts, route preimage, all control patches and deadlines remain unchanged. Total successor828 additions/6 removals. Largest carrier5728 bytes and positive aggregate11970 remain unchanged; proposed30 expected/max40 requests across three children is still awaiting measured release. No tests/build/services/controls were executed to make this correction.
