# Native audit of actual principal-code supplement

**No new blocking source defect is established. The substantive correction is to our earlier F1 claim and the planned proof boundary:** in the actual pinned default Vitest worker path, test process.argv does not receive the CLI --outputFile arguments. The new argv normalization is harmless defensive code, but neither the original ordinary-run failure nor execution of its repaired branches has been demonstrated. A relative-output positive run can establish actual reporter compatibility and artifact placement; it cannot establish branch coverage of that test-local scan.

## Scope and pins

Read all five actual supplement findings and its checked-clean notes, the complete relevant source, original preserved files and integration chronology, pinned Vitest4.1.10 config/fork/worker/reporter code, and EP2 supervisor/adapter. No test, service, product import, compiler probe, review CLI or source edit was run. This report is the sole artifact written.

Original frozen target and accepted repair remain independently recorded. Supplement target test is 0f941001445e001f96f8c432a7099a5eb29dc4b68bc4458d6d1dfae472a3ce20, progress fd69135a88ee2802eb9ce55d2c2f3771d4a16228c0e306202c9a8778f0bdcb3f; repaired binding fce5c0dfb33221f9cc29fd6420eab607722d88d0d9b6076b64a57f763f5f9406. Actual hashes of the inspected supplement/tool files are appended below. Source HEAD chronology is separate from immutable test bytes.

## All five findings

### F1 — historical wording is supported; no mandatory correction

integration-source.json records an actual ordinary fast-forward at18:17:42Z from18104a03b850553641f24fe83e7318493fa102ca to f5511718371a6935614eae6756e796cef71570c2, preserving test e6af6263 and doc f0171ecc exactly. review-repair/before contains those same original bytes. The review repairs happened afterward. Thus “The unchanged draft was integrated onto f551…” accurately describes that earlier event; it does not assert that the final repaired test is byte-identical to the original draft. A dirty worktree can be carried across an ordinary fast-forward without committing its disjoint changes. “Integrated” here is not a claim that the draft was committed or published.

Adding “then review-repaired” would be optional chronological clarity, not repair of falsified evidence. Do not rewrite the historical event into a claim that repairs preceded the fast-forward. Current final-candidate/publication identity remains externally gated.

### F2 — historical checkpoint, not current execution ledger

The paragraph expressly says “at this draft checkpoint”; original authoring was unexecuted, and before/ plus original report preserve that state. Actual earlier/later static records remain in EP/EP2 with their true source pins and timestamps. Those records supersede the checkpoint operationally without making the historical sentence false. Inventory is not packaging, so the review must not imply packs/consumer work occurred merely because build-inventory passed.

Root may add a later evidence checkpoint when actual runs finish, but should not relabel earlier authoring evidence. No source change is required solely to turn the historical note into a live status report.

### F3 — tuple closure sufficient; literal-key enhancement optional

The accepted as-const tuple makes the pair shape explicit for Object.fromEntries. The keys remain a string index signature and checked(text,schema:string) retains a runtime unbound-schema guard. This is adequate for the fixed five literal callers; no mistyped key was found. Root typecheck uses tsconfig.check.json including the test and noEmit; package build config exclusions are irrelevant to that gate. A const schema-name union is optional. We still do not claim an empirical measurement of the original inferred any type.

### F4 — lexical containment limitation acknowledged, not an accepted-run defect

outsideStore compares lexical absolute paths; it does not prove symlink-resolved containment for arbitrary aliases. That is a real limitation of the helper, not newly introduced by the repair. In accepted EP2, TMPDIR is an explicit absolute /Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-principal-key-draft/execution-preparation-v2/run/<cohort>.tmp; the owned randomized store is its child. The result file and custom observation output are fixed siblings outside that store, not attacker-selected alternate /var or /private/var spellings. EP2 sets TMPDIR itself and independently checks store path/absence. The macOS default /var alias does not describe this planned environment.

Do not promote the helper to an adversarial filesystem sandbox. No realpath redesign is required for these fixed accepted paths; the remaining alias limitation must remain explicit.

### F5 — confirmed for actual configuration; correct earlier F1 assessment

The concrete installed dependency source proves the following chain:

1. Repository vitest.config.ts73–84 sets include/setup/coverage only, with no pool or argv override. EP2 invokes the pinned Node/Vitest CLI with no --pool/custom worker options; it clears NODE_OPTIONS/NODE_PATH and inherited BDP overrides.
2. Vitest dist/chunks/coverage.DM_a_rWm.js180 sets resolved.pool ??= "forks". The later threads default in that file is a different resolution context, not an override of this resolved pool.
3. dist/chunks/cli-api.BK8pd4xc.js3536 selects new ForksPoolWorker. Constructor3072–3073 points it at dist/workers/forks.js. start3084 calls Node child_process.fork(this.entrypoint, [], {...}); the positional argument array is explicitly empty. execArgv is Node runtime options (resolveOptions3731ff and project config), not parent CLI process.argv.
4. dist/workers/forks.js initializes runBaseTests/setupBaseEnvironment through workerInit. A bounded search of Vitest dist for process.argv/argv assignments and the repository setup source found no restoration of parent CLI arguments into that worker. The argv matches belong to the CLI parser, not test-worker setup. The reviewed test and setup do not mutate it.
5. The actual implementation is Vitest's own ForksPoolWorker. It imports node:child_process directly; tinypool appears as a historical comment link at3109, not the selected pool implementation. The conclusion does not require assuming a different tinypool release.
6. Pinned JSON reporter index.UpGiHP7g.js3619–3627 receives resolved outputFile configuration in the parent and resolves it against ctx.config.root. That path continues to work independently of the worker's argv scan.

Therefore the selected ordinary forked test will inspect worker arguments without --outputFile; its argv loop performs no outputFile normalization/containment for the actual CLI flag. This is source-established reachability, not a newly measured process.argv capture. No runtime observation or extra test is needed to recognize the limitation.

**Erratum:** my original native-findings-audit called F1 a “required compatibility correction” and a “concrete ordinary invocation defect”; native-repair-review and EP2 preparation text also described relative positive output as exercising F1. Those statements were too strong because they skipped worker argument propagation. Preserve those artifacts as historical, but interpret them with this correction: the original scan would reject relative paths **if those values were in test process.argv**, which this actual configured worker path does not supply. The accepted change improves that dormant/alternate-embedding branch and does not create a demonstrated regression.

The relative positive command may remain without adding requests: it verifies that the real parent reporter writes the expected external JSON file and that EP2 resolves it correctly. It cannot be credited as validation of missing-value throws, containment branches or path.resolve inside the test. Fixed external paths and the live custom BDP output check still provide the accepted evidence placement boundary. Do not broaden this into a new argv instrumentation run, custom pool or transport API merely to validate defensive code.

## Checked-clean statements and disposition

The exact narrow repair preserves the custom output's absolute precondition, request/role/body/schema oracles, startup/cleanup handling,7positive/expected2negative HTTP,14maximum and two product children. Empty and separated-option values are handled consistently if the scan is invoked with them; ++i consumes one value. These are source facts, not observed branch coverage. No new runtime or public-qualification claim follows from static success.

Recommend retaining the accepted minimal repair and proceeding under root's actual code-council/package/installed gates, with this explicit proof correction. F1/F2 supplementary documentation concerns do not justify historical rewriting; F3/F4 are bounded optional limitations. No additional code change or run is requested by this audit.

## Inspected input hashes

- /Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-principal-key-draft/claude-code-supplement/stdout.txt: 9fb12dcd3ed820680d3ee3d8bdc7da334f916005fa367f5c927d9611827722cc
- /Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-principal-key-draft/claude-code-supplement/exit.json: eeb6f513b10e13bcb2e3e0bb8ccbc911f3de44fb287514bb3db22901aedf315c
- /Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-principal-key-draft/integration-source.json: 65081d918c480eb22b1a636bf3ffc8e738087bf96d3fc611962e3ef0712fc671
- /Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-principal-key-draft/review-repair/source-binding.json: fce5c0dfb33221f9cc29fd6420eab607722d88d0d9b6076b64a57f763f5f9406
- /Users/dbox/repos/wt/vickie/vickie-principal-key-isolation/vitest.config.ts: a84575ea9ef0b6e75b6484024a627fa89422aea2fdbeb26127e09714dba0641f
- /Users/dbox/repos/wt/vickie/vickie-principal-key-isolation/packages/server/test-support/read-conformance-mock.ts: dbfb86ac2325c684567ffce40f17105c6b12a63edfa54cc08f0601df8dc613f8
- /Users/dbox/repos/wt/vickie/vickie-principal-key-isolation/node_modules/.pnpm/vitest@4.1.10_@types+node@24.13.3_vite@8.2.1_@types+node@24.13.3_/node_modules/vitest/package.json: caa41f04799bd42f3cfd100c4282e630d77ffc7deb1e7e4927794fcb7f137f34
- /Users/dbox/repos/wt/vickie/vickie-principal-key-isolation/node_modules/.pnpm/vitest@4.1.10_@types+node@24.13.3_vite@8.2.1_@types+node@24.13.3_/node_modules/vitest/dist/chunks/coverage.DM_a_rWm.js: e509f3cc1bd81ae6426265253fa926cbde02be5cebfe35b1df422017bffdca31
- /Users/dbox/repos/wt/vickie/vickie-principal-key-isolation/node_modules/.pnpm/vitest@4.1.10_@types+node@24.13.3_vite@8.2.1_@types+node@24.13.3_/node_modules/vitest/dist/chunks/cli-api.BK8pd4xc.js: a2d629ec07fbaf58055bca9858076c6c13763042d97196272c45ccfdb727f498
- /Users/dbox/repos/wt/vickie/vickie-principal-key-isolation/node_modules/.pnpm/vitest@4.1.10_@types+node@24.13.3_vite@8.2.1_@types+node@24.13.3_/node_modules/vitest/dist/workers/forks.js: b461149b7a4488ff72f90af16d900ad9c8c410e1b8ddb2c443d49f6cd1aa7206
- /Users/dbox/repos/wt/vickie/vickie-principal-key-isolation/node_modules/.pnpm/vitest@4.1.10_@types+node@24.13.3_vite@8.2.1_@types+node@24.13.3_/node_modules/vitest/dist/chunks/index.UpGiHP7g.js: cdafaf8e508042dddba72f8c7bbae84e2836ddb6b06c7026520c35d91af4e798
- /Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-principal-key-draft/execution-preparation-v2/supervise.py: da74971257c3cc804b404dbc2d72759e301431e04d2474c692a658ed43b18f9a
- /Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-principal-key-draft/execution-preparation-v2/review-cohort.py: 87f7895e3458ec082cecd342b785a1093c5caa8466dcdc4938992bb5acd6f46b

Context HEAD: f5511718371a6935614eae6756e796cef71570c2
