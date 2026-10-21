# Command-minimum execution preparation — draft only

No execution is released or performed. This directory is outside the product checkout. The product author remains the sole writer of its seven files. No execution-gate.json or run directory is supplied. execution-gate.example.json is deliberately closed and cannot authorize a command.

Inputs read: reviewed command plan v2 f8824197e765e8228679044ab32fd7a54ebd26ec3a892614fdd9b5a15cf07c82; authoring-only acceptance5815c7b2f7a79e1824b4922187f8c719df1c0e033351bf7305feb372e7457564; current package scripts; HTTP/service fixtures; emitted RU launcher/request helper; installed collection helper. Adaptation basis is the reviewed cross-carrier supervise.py/gate.py and maintenance supervise-v4.py. Product code was not imported or executed.

## Files and default-closed gate

- commands.json pins Node24.16.0 and pnpm11.20.0, checkout, full argv, literal anchored selectors, prospective workload and owned temp prefixes.
- gate.py permits only source-phase IDs explicitly approved by a future root measured-execution gate. It verifies that acceptance document's hash, plan hash, the preparation scripts/ledger hashes, exact main122800511ee6306e4105e1f73bf3658f35c90a14/tree088290751951c353797b984f690b7d9fd96deb80, seven measured draft hashes, binary diff hash and no unowned tracked/untracked changes. The authoring-only acceptance is not used as execution authorization.
- supervise.py executes at most one exact ledger command per invocation after the gate. It does not accept arbitrary argv/cwd. It refuses any prior attempt artifacts for that ID. A partial or unsuccessful earlier attempt blocks later source commands, with no automatic retry, cleanup waiver or mutation of original results.
- execution-gate.example.json documents the minimal future root fields, with released:false, empty approved IDs and null measured bindings. Do not rename it into a release without root's independent review.
- preparation-hashes.json records the prepared file bytes; it does not certify execution.

The source draft is still being authored. No guessed hashes, false final source commit or borrowed predecessor payload is inserted. gate.py checks the actual parent and dirty draft hashes; the future installed phase instead needs a separately accepted clean ordinary successor and package provenance.

## Exact source commands

Eleven commands are ready as argv proposals, each with a120-second supervisor cap:

1. offline-setup: pinned pnpm install --offline --frozen-lockfile, only if root approves setup.
2. typecheck: pnpm typecheck.
3. build: pnpm build (the actual package script performs clean then tsc -b).
4. format-check: pnpm format:check, never the write-mode format script.
5. lint: pnpm lint, no fixes.
6. check-boundaries: pnpm check:boundaries.
7. evidence-verify: pnpm evidence:verify, unchanged Read evidence verification.
8. http-positive: exactly module-scope A–E in transaction-http.test.ts.
9. service-exclusions: exactly the nested six-row S case in transaction-service.test.ts.
10. catalog-citations: exactly the nested Q case in transactional-catalog.test.ts.
11. ru-emitted: exactly the new module-scope RU case in installed-reference.test.ts.

All four test commands use --maxWorkers=1, verbose+JSON reporters and unique outside-checkout result files. S/Q include their describe prefixes; A–E use an anchored alternation of the five complete literal titles. Test commands require successful typecheck/build exit artifacts for the same bound source, because emitted imports must exist and match the draft. Root may approve only the next command rather than the whole list. Exit zero is not a replacement for root review of selected case counts, assertions, reports and cleanup.

Positive source workload is8 cases,86 planned HTTP,7 store directories,13 owner opens,13 listener lifetimes,one RU child and6 read-only SQLite handles. S adds6 direct handler calls but no TCP requests or listener. RU includes one port-probe listener plus one serving listener. Adding the two held controls gives reviewed10 cases,89 expected/119 maximum HTTP,9 stores,15 owners/15 listeners andone RU child. Counts are prospective until the actual draft and observations are bound; no request instrumentation or extra call is added by this harness.

## Explicitly held later work

control-command-gap and control-receipt-bypass have no runnable argv/cwd. They require separately accepted disposable derivative paths, exact patch/file/tree hashes, emitted rebuild commands, positive-source baseline and named expected-failure adjudication before a dedicated control release. Expected first failures remain old command500 at A request1 and receipt200 against E's503 at request2; max27/6 HTTP. This source gate cannot treat a control exit1 as a pass or allow it to fall through to later execution.

pack-and-consumer is a non-runnable placeholder. No packages.py or package/install command is supplied now. Its later preparation must bind the final clean source/build, protocol/server archives and all member bytes, a fresh offline consumer, full runtime dependency graph/payload, actual entry and configuration/artifact/authentication identities. The old cross-carrier source, archive counts/digests and192/21 constraints are not reused.

fresh-tx has the exact proposed18-request test argv for review but phase:held/runnable:false. Even placing its ID in approvedCommandIds does not enable it. It requires separately accepted package provenance and only then the two real BDP_TRANSACTIONAL_INSTALLED_ENTRY/BDP_TRANSACTIONAL_SOURCE_COMMIT overrides. No output environment variable or cross-carrier observation hook is needed: the existing helper prints safe console JSON with observation, outcome, source declaration, entry hash, binding, child PID/exit and18 response fingerprints. It does not retain full wire bodies or raw credentials. The fresh cohort is one child/store/owner/listener, nine URLs, eight POST/nine GET/one HEAD, separate from all source/control allowances.

Later phases need a narrowly reviewed successor to this preparation; filling the source gate cannot enable them.

## Supervision, failure evidence and cleanup

The supervisor clears all inherited BDP_* variables plus NODE_OPTIONS/NODE_PATH, pins Node's PATH prefix and CI=true, assigns a unique attempt TMPDIR and supplies no BDP override for source commands. It never prints an environment dump. Started metadata records actual parent/dirty state and measured source hashes, not a fabricated commit for emitted output. Process samples retain executable identity through ps comm, excluding process argument strings; the task's stdout/stderr still go to its retained log, so product test assertions must continue their existing credential-safe reporting discipline.

Each attempt gets exclusive started/log/processes/results/temp/exit paths. A retained started marker blocks reuse even if setup/spawn fails. The supervisor emits progress every30 seconds, samples owned process-group membership and temp entries at50 ms, limits command time to120 seconds, and uses SIGTERM then at most5 seconds before SIGKILL on only its own new process group. Reaping/cleanup grace is separate from the120-second command budget. SIGINT/SIGTERM trigger retained cleanup evidence; forced cleanup, inventory errors, remaining sampled PIDs/groups, owned runtime roots or changed source make the supervisor fail. There is no broad process-name kill or temp deletion.

Fixture roots are classified per selected command under only that attempt TMPDIR: bdp-tx-http- for A–E/controls, bdp-tx-service- for S, bdp-installed- for RU; the held TX scenario records bdp-collection-minimum-installed-. Classification uses the union of sampled still-present entries and the actual final directory listing. Other entries, including Node caches, remain untouched and visible for root review. Source hashes/diff/head/untracked inventory are checked again at completion.

An ordinary test failure remains its actual nonzero exit, with logs, JSON result if produced, original attempt and lifecycle evidence preserved. No automatically generated report claims an expected-control failure was valid or a helper closed flag proves process/directory absence. Group sampling can miss brief or detached descendants; actual helper launch/close evidence and independent final PID/directory checks remain separate.

Future invocation, only after root release, is the session-guarded Python command ending in supervise.py COMMAND_ID. No invocation of this harness, tests, setup, build, installation, packaging, service, subprocess.Popen or product code was performed during preparation. Script syntax was checked by parsing Python source text only; JSON and hashes were produced by the external guarded file writer. The required role-session guard read is the only network action.
