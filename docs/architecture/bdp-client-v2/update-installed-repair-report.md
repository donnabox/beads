# Private installed boundary repairs

Frozen head: `dea6f9e7a95a3e3eba9c667de39360346f25bdf4`
Base: `4078ca709b55d363c147a5bdc0d5141ede9ec247`
Branch: `codex/vickie-update-installed-repair`
Checkout: `/Users/dbox/repos/wt/vickie/vickie-update-installed-repair`
Origin: `https://github.com/gastownhall/bdp.git`

The original `vickie-update-installed` checkout remains frozen. No push, PR, coordination, launch or public activation action was taken.

## Three native review findings

- H3: `development-runtime.ts` uses a single serving=false admission gate before authentication/routing. It becomes true only after every principal Read/discovery route is initialized; close disables it. Deterministic emitted-runtime tests hold the real second-principal discover, observe GET503 and POST503, then exercise success and later startup failure. Both paths prove Resource404 and original-key create200 after readiness/reopen. The injected seam delays or rejects discover only; actual HTTP, owner and durable state are real. This tests the initialization invariant, not a claimed exploit window in ordinary synchronous startup.
- M4: `installed-reference-process.ts` invokes main only when the real process entry path matches the module path. Separate child tests prove import without IPC and import with IPC leave listener counts, exitCode and stderr unchanged. Direct execution without IPC still exits2 with the fixed redacted diagnostic. No public export/bin/flag was added.
- M2: `development-authority.ts` supplies one nested-frozen `DEVELOPMENT_OWNER_LIMITS` tuple to the actual owner, receiver/member bounds, store minimum retention and pagination. Installed config identity hashes that same tuple. Existing numerical defaults and advertisements are unchanged. The tuple now covers patch/path/representation, lifecycle retention/minimum/maintenance, page/cursor/selector, pagination capacities/token settings and random token bytes. Test-only imported tuple substitutions verify hash sensitivity for six previously absent fields across member/lifecycle/pagination groups and restore the baseline; they are not production configuration overrides or an assertion that the altered mock tuple runs an owner.

## Identity and claims

The qualification configuration identity necessarily changes because its captured tuple is more complete. Persisted contract/aggregate installation identity is unchanged by this repair, so the same configured store can reopen. Historical configuration/evidence hashes are not relabeled. Packed smoke observed payload identity `a48d8ef04784895374cc3b10406d412a16e15644ce70ea66ed07dbcd30e0a7d0`, claimEligible:false, all nine targets and stable Scope. That hash is observed payload metadata, not attested write-profile evidence. No generic config, store, evaluator, schema, public admission or Type lifecycle change.

## Verification

Pinned Node24.16.0 / pnpm11.20.0; frozen install and build PASS. Full logs in this directory.

- Focused three files: 14 tests PASS (`focused.log`).
- Broader16-file suite: 497 tests PASS, maxWorkers2, 8.61s (`regression.log`, structured `regression.json`). This includes all491 native-review baseline tests plus six new tests. Files: installed-reference/startup/import, development-contracts/discovery/server/disconnect/matrix, server-composition, read-sequence-lifecycle, node-http-listener, http-handler, read-http, recovery-store, installed-contract-artifact and installed-schema-evaluator.
- Typecheck, boundaries, format check PASS. Lint PASS with existing3 warnings/1 info; no fixes applied.
- Packed installed smoke PASS (`smoke.log`), including private entry inventory/deep-import denial, actual installed factory observations and ordinary public Read/Write-refusal checks.
- Development smoke PASS (`smoke-development.log`).
- Existing public Read evidence verification PASS (`evidence-verify.log`); artifact constant674982be2123b9b7aae630bf79c3f774122900cd, original runHeade5f7d0c43c151426350ceededdc77fe7e372db9b, evidenceCommitdbaf63a24219e1ea3f80394746853598b0d4464a,89rows unchanged.
- git diff --check PASS; commit complete and working tree clean.

No interrupted/timed-out test run or skipped hook. No full repository suite rerun and no new full-profile, power-loss, installed between-member interruption or production deployment claim. Cumulative executable qualification and independent profile-specific evidence/admission remain subsequent work; external review may identify additional bounded repairs.
