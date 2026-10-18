# Read reseal feasibility at 29eb76fd

Assessed read-only against `/Users/dbox/repos/wt/vickie/bdp-profile-write-candidate`, exact HEAD `29eb76fd8d6a841e69df54c0a88bd22477ae9465`. No generator, baseline verification, CLI invocation, evidence mutation, staging operation, or database creation was performed for this assessment. The checkout remains clean.

## Recommendation and authority

A fresh two-target Read reseal is technically supported entirely through disposable BDP-owned fixtures, with the existing CLI compatibility pin. No Janet checkout, CLI branch, staging deployment, pinned playground or shared database is required. Under the user's authorization for BDP validation/conformance, running those isolated fixtures is within Vickie's assigned BDP scope (interpretation of the user's ownership instruction). Root should own the frozen-head decision, evidence commit, review and publication. No new operator product ruling appears necessary if all existing pins and evidence policy remain unchanged.

This certifies the BDP reference target and adapter against the existing **bd 1.0.5** baseline. It does not qualify Janet's current Preview 2 CLI or establish CLI/BDP parity. Changing the CLI pin, accepted portable build identity, observations, coverage exclusions, schema seal, or evidence policy is outside routine resealing: coordinate qualified CLI identity/behavior with Janet, and escalate product/policy changes to Donna through the existing bus. Do not recapture a drifting baseline to make the gate pass.

## Why isolated runs are supported

- `scripts/generate-read-cohort.mts:373` checks a clean tree and records HEAD, rebuilds both packaged payloads, then performs the pinned baseline gate at line393. It verifies the tree and HEAD again before writing at line690.
- Baseline `observe()` at `scripts/bd-baseline.mjs:376` creates separate temporary HOME/workspace, pins git identity, uses an allowlisted environment, and cleans up. Only PATH/TMPDIR pass through. Baseline verification compares observations without rewriting them (`:493`); `capture` rewrites baseline evidence and must not be used (`:480`).
- Generator line427 creates a private `bdp-read-cohort-*` root; lines439–450 rerun the two reviewed matrix suites and require the real bd lane. Lines458–508 create private HOME, two private bd workspaces, and seed them through the CLI. Packaged targets listen on loopback and receive explicit private workspace configuration (`:526–563`). Fixture Type descriptors use a controlled local publisher (`:244`). Finally removes the private root (`:729`).
- The real bd matrix similarly creates its own private root, HOME, workspaces and restored copy; it allows only pinned environment fields (`apps/bdpbd/src/read-matrix.test.ts:350–450`). Its lifecycle cleanup removes that root. An explicit `BDP_BD_MATRIX_EXECUTABLE` prevents ambient CI or another executable override from silently selecting a different binary (`:157–166`).
- `packages/conformance/test-support/bd-workspace.ts:65–84` supplies an explicit working directory and exact environment for each bd command, with timeout/output bounds. Mutations seed disposable fixtures; this does not edit or implement CLI behavior.

## CLI dependency checked without running it

Existing pin (`scripts/bd-baseline.mjs:58`): version `1.0.5`, schema_version `1`, observations_digest `d39ced7abdaec37513b93bcafa34ddfbaf62f935cc3a2aa7c6eb1f0b0ccca38b`. Portable identity is the Homebrew tuple, including branch `v1.0.5`, build `Homebrew`, and text `bd version 1.0.5 (Homebrew)`; see `docs/design/evidence/bd-baseline/README.md:23`.

Read-only filesystem check found `/opt/homebrew/Cellar/beads/1.0.5/bin/bd` symlinked to `beads`; SHA256 of that executable is `2fff6a2dc534cf8b56ec1f7a0c98b1e8509a0a5408905c420769a7213ee34707`, exactly the documented capture-host hash. Thus the required binary is present. This is strong availability evidence, not a substitute for fresh `baseline:verify`.

CI's source-built bd 1.0.5 `(dev)` is accepted by the matrix fixture, but is not a replacement for the generator's Homebrew baseline gate. `.github/workflows/ci.yml:36–52` installs/verifies that distinct tuple; the evidence law explicitly distinguishes it (`packages/conformance/matrices/README.md:581`). Do not upgrade, rebuild or replace the installed CLI for this reseal.

## Exact proposed workflow — not executed

Use a new isolated BDP clone/branch from the chosen final implementation head. For this assessment that is `29eb76fd8d6a841e69df54c0a88bd22477ae9465`; if more workers integrate first, regenerate from their reviewed combined head instead. No concurrent source edits during generation. Keep all logs outside that clone.

```sh
export PATH="/Users/dbox/.nvm/versions/node/v24.16.0/bin:/opt/homebrew/Cellar/beads/1.0.5/bin:$PATH"
export BDP_BD_MATRIX_EXECUTABLE=/opt/homebrew/Cellar/beads/1.0.5/bin/bd
export BDP_REQUIRE_BD_MATRIX=1
pnpm --version # must be 11.20.0; node --version must be v24.16.0
pnpm install --frozen-lockfile
pnpm build
pnpm format:check
pnpm lint
pnpm check:boundaries
pnpm typecheck
pnpm exec vitest run --maxWorkers=2
pnpm e2e:ready:preflight
pnpm smoke
pnpm smoke:development
```

Before generation the known three current-seal positive tests and `pnpm evidence:verify` remain red for the documented catalog mismatch; do not describe that pre-generation state as green or alter those assertions. Require every unrelated check to pass. A separate `pnpm baseline:verify` can fail early, but generator repeats it; it is safe only in the sense that it uses temporary databases, not a read-only command.

Then, with the implementation committed and tree clean:

```sh
pnpm evidence:generate --maxWorkers=2
```

This executes `vitest run --config vitest.evidence.config.ts`, rebuilds, verifies baseline, reruns both matrix suites, launches packaged processes per applicable scenario, and writes only `docs/design/evidence/read-cohort/read-v1.json`. It prints the new constant and exact run head (`scripts/generate-read-cohort.mts:701–725`). The script has no output-path/dry-run flag. To experiment without replacing an existing checkout's evidence, run in the dedicated clone.

Record exactly the printed constant in `packages/server/src/read-conformance-capability.ts`; inspect artifact runHead, provenance, required/not-applicable/uncovered rows and changed bindings. Commit **only those two files**, directly atop the observed implementation head. A precommit production `evidence:verify` cannot pass: its input cleanliness gate intentionally rejects dirty evidence (`scripts/read-cohort-evidence.mjs:130`). After the evidence-only commit, rebuild and run the complete checks above plus:

```sh
pnpm evidence:verify
```

Require exit 0, then exact-head CI, independent whole-branch review and tree-equality proof per `packages/conformance/matrices/README.md:631–638`. Preserve the run-head→evidence-commit two-file delta and ancestry when integrating: ordinary merge of the seal branch, not squash/rebase over unrelated implementation changes. The verifier checks the latest evidence-path commit and that confined delta (`scripts/read-cohort-evidence.mjs:123–173`).

The existing non-null shipping record already permits the generator's packaged listeners (`packages/server/src/read-conformance-capability.ts:49–60`); no bootstrap constant or test grant should be added. Current qualification is still checked separately by evidence:verify; the old record is not proof for current catalog bytes.

## Cost, blockers, limits

- Expected minutes, not a guaranteed duration. The prior full run on this exact candidate took 62.11 seconds; real bd matrix took 56.54 seconds, reference matrix 0.362 seconds. Generator additionally rebuilds, runs a baseline (13 named creations with 12 × 1.1 seconds intentional spacing plus 45 fillers), reruns both matrices, and launches a fresh packaged server for each selected row/target. Its configured test budget is 30 minutes (`vitest.evidence.config.ts:18`); this is a timeout ceiling, not an estimate.
- Baseline verify is the first meaningful external dependency and has not been run in this assessment. On output/identity drift, stop and report to root/Janet; do not use `baseline:capture` or loosen the pin.
- The package generator is ready without Janet's assets. Qualification against a newer Preview 2 CLI would be a different workstream and needs Janet's qualified pin and reviewed baseline/policy changes.
- `pnpm e2e:ready` is **not** part of the reseal generator. It requires `BDP_E2E_BD_WORKSPACE` and uses fixed ports 18280/18281 (`scripts/e2e-ready.mjs:373–388`). Do not let ambient values route it to a pinned/shared workspace. CI runs `e2e:ready:preflight`, not that standalone full E2E command. If a separate readiness demonstration is requested, prepare a separately reviewed disposable fixture and explicit environment.
- Normal cleanup removes disposable roots; abrupt signals can leave temporary baseline directories (documented in `docs/design/evidence/bd-baseline/README.md:98`). Do not clean arbitrary shared temp paths. A supervised run should retain logs and clean only its identified children/roots.
- Any later change to exact Read catalog/manifest/fixture bytes invalidates this new seal again. Finish pending integration first if avoiding duplicate work matters. Runtime-only later changes may not trip every byte gate; final scope/review should still identify what head was actually observed.

## Options

1. Recommended: wait for a reviewed combined head, then root delegates one disposable two-target reseal preserving all existing pins. No new human decision presently identified.
2. Generate from 29eb76fd now in a disposable clone as a separately reviewable successor seal; expect possible repeated qualification after upcoming integration. Root owns adoption.
3. If the exact baseline cannot be reproduced, remain explicitly unqualified while Janet supplies a reviewed CLI/baseline path or Donna rules on a pin/policy change. Do not reinterpret historical evidence or weaken current-input checks.
