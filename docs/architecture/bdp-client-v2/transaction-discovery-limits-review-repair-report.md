# PR107 limits review repair — frozen source report

Source **26d6cd1b947db8c00e24015771506d26098a293e**, tree **375c1910e99ada9f5a43e332203c8778ab8122dc**, direct ordinary parent **9b95df5dd866dbaf84150404c4db86ac8054b724**. Checkout `/Users/dbox/repos/wt/vickie/vickie-discovery-limits-review-repair`, branch `codex/vickie-discovery-limits-review-repair`. Four authorized files only, 46 insertions/23 deletions including test formatting. Original fd07 and9b95 are preserved. Ordinary commit has the inherited Agent-Signature trailer; no cryptographic signing/configuration change.

## All five accepted findings

1. **Single capture:** assembler reads Selector bytes/depth/nodes once, validates those exact locals against actual set enforcement, and projects those locals. Accessor regression flips every scalar after its first read, asserts one read each, complete expected4096/32/256 output and frozen result. It deliberately accepts the coherent first capture; it does not claim a second value must be rejected. No configuration or enforcement change.
2. **Shipped defaults:** existing runtime configuration test imports actual TRANSACTION_SELECTOR_LIMITS and asserts TRANSACTION_RUNTIME_DEFAULTS.selector equality, alongside existing shared captured/enforced/advertised identity assertions. This is the root-approved additional test file; no runtime configuration production edit.
3. **Partial validation groups:** both diagnostics-only16 and diagnosticBytes-only524288 candidates are positively parsed by the unchanged permissive protocol schema before the literal owner oracle rejects them. Existing absent/changed cases remain.
4. **Optional enumeration:** progress prose now explicitly lists sequence and patch among omitted optional groups.
5. **Wrap:** the existing no-serving/no-public-admission sentence is wrapped without changing its claim.

## Executed verification

Pinned Node24.16.0/pnpm11.20.0. `checks.py`, `pack.py`, `validate.py` preserve exact commands; `exits.json` records all16 stages, each exit0. No failed attempts or interrupted runs in this repair. No hook bypass.

- Frozen-lockfile install; owned TypeScript format; clean → prebuild typecheck → build: PASS.
- Focused discovery/runtime-configuration/service/Read/HTTP/protocol-discovery/installed-contract tests: **250 passed,7 files**,0 failed/skipped (`affected.json`). Elapsed stage3.844s.
- Full format check, lint, dependency boundaries and existing Read-seal verification: PASS. Existing large Read+Update JSON formatting skip is unchanged; it is not evidence that the skipped file was reformatted.
- Freeze clean source, rebuild, actual protocol/server pack and offline install: PASS.
- Exact frozen installed discovery-boundary test: **1 passed,12 completed HTTP exchanges,1 child**,5 distinct URLs. Six authenticated discovery refusals500, three anonymous401, real batch/afterstate/retained retry200. Stage0.858s; child PID72898 exited0 and is absent.
- Two targeted installed RU tests (retained retry/current state and reference syntax/category failures): **2 passed**,200 tests deliberately filtered,0 failures. Stage2.739s. Same freshly installed server package; retained reference report in `packed-ru-reference-report.json`.

This is focused validation, not a full suite. No additional mutant run was required: the actual accessor-count regression and schema-valid partial-group substitutions discriminate the repaired behaviors. Installed proof preserves HTTP refusal/mutation boundaries; discoveryObserved remains unreachable on those wire routes, so the12/1 cohort does **not** execute or qualify internal projection. The direct source tests do exercise it.

## Exact identities and evidence

- Protocol tarball SHA256: `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227` (unchanged bytes from preceding package).
- Server tarball SHA256: `502bb785fe3f54503440a7a3c9298c4c607a086470b1f5e5b00f73235ecdb222`.
- Installed payload digest: `6945e90b33b51b1fc26cb92866399de58d36dee2da4fc09ebe126fe5ccba59ef`.
- TX configuration identity: `b7d436111e9b4ae5d3790a6e4bdc0b88b33e1fa6d86c5380836ba273db5beb5d`.
- Entry SHA256: `bcf153170f953f9e28c6c89f1153be907f64df289e64f3ad929cad4a4a54c47c`.

`source.json` records exact changed-file hashes and clean/tree/parent identities. `summary.json` retains real safe completed responses, wire/body hashes, actual configuration/contracts/policy/auth/fixture identities and all stage exits. SourceCommit is launcher-declared against checked git state, not independent target attestation. Production byte changes explain a fresh server payload; historical packed evidence is not relabelled.

Cleanup: installed child exited0; path-bound process inventory found no remaining installed children. `cleanup.json` records this and the clean worktree. The external log/pack/install artifacts remain available for review; no shared DB or unrelated process touched. No activation, public admission, schema/spec/catalog change or profile credit. Root owns PR advance, further council and merge.
