# Native actual bootstrap/static-prefix audit

All five authorized stages passed once against the frozen Link draft. No source change or runtime execution is indicated. No additional run, helper invocation, formatter or source edit was performed by this audit.

| Stage | Raw exit / accepted | Actual supervised seconds |
|---|---|---:|
| workspace-install |0 / true|4.570667|
| bind-workspace-tools |0 / true|8.512783|
| typecheck |0 / true|5.290259|
| format-check |0 / true|4.430586|
| lint |0 / true|4.262739|

Install log records offline reuse of61 packages, no downloads, frozen lock resolution skipped, and the configured scripts-disabled command matches its start record. Tool capture produced the actual11 executable/wrapper/manifest identities; all realpaths and byte hashes still match. The actual measured tool document SHA745323190a5534a6714e17f8e8e3e57d507b59890750ccb81c7a192068972a8a binds the correct source identity, lock, HEAD and install exit/log hashes. This is fresh checkout evidence, not reused Bead evidence.

Typecheck ran tsc -p tsconfig.check.json with no diagnostic output beyond the command line. Format checked the six owned files and applied no fixes. Lint likewise applied no fixes and passed with **35 warnings and1 info**, not zero warnings. The logger retained its diagnostics-limit message:16 diagnostics were not displayed. This audit does not infer their individual content or request a warning sweep/rerun.

## Independent verification

Checked every terminal exit record, start argv/cwd/bounds, process-sample JSONL and recorded log SHA. All five records use source identity9ec1d3433641f3ff643b9297a77220a2bbe3b95d862300c76b64e3e34d322af3, sourceUnchanged=true, no errors/forced cleanup/survivors/runtime directories. All44 distinct sampled PIDs are currently absent by independent ps check. Each stage-owned tmp tree contains no bdp-* runtime directory. No exit record exists outside the authorized five-stage prefix. This checks sampled processes, not an invented census of every ephemeral subprocess.

All679 source hashes match binding97629a7fac3bc7b2e122b3446b1151497e60e8585e3ba8ebb9fa8c6d6503a2c9; HEAD remains5bcd277b2a61461211ffc0c2d2f8fd701ef0a385 and exactly the same six tracked files are dirty. The source was rehashed again immediately before this guarded write.

## Exact release transition

The previously reviewed index63f87881d050300db9b4978bcef2edb85297eb0bec88a98030ca4075b1e57297 and closed commands/gate are preserved under root-closed-preparation. Current index3826b2faefa766e238c7afdb5dd4861d0d8eafdaf8708b5a897b7b3d98bb4911 still has90 files; all hashes match, and commands.json is its only changed file hash. Parsed command comparison shows the sole change is configurationFinalized:false→true. Every helper, selector, argv, cwd, bound and stage order remains unchanged.

Current gate SHA4512bc6b13da3774b9fe3dd082490f20db8b26bd372ae10d686248f51ba38b68 is the static stages' recorded start gate. Setting only workspaceTools back to null reproduces exact bootstrap start-gate SHAab6c648bf7e48b02db548266db0ded092e498770bac79f8c3c862b946af40c75. Thus the between-stage transition is precisely pinning the newly measured tools; the five-stage approval was already explicit. Relative to the preserved closed gate, changes are release/plan/source/static/native flags, the authorized five-stage prefix, implementation index, root disposition and measured tool pin. Source/base-tool pins and null consumer provenance fields remain unchanged.

Root disposition1541b29e15ab795ad39abfc9a6174b2bb9ba5e175d46ee3683b8a7f0a184ff45 explicitly authorizes only dependency preparation and static checks. actualClaudeReviewDisposed=false and runtimeValidationAuthorized=false remain; fresh-build is not in approvedStages. This prefix supplies no serving/installed/elapsed-margin result. Runtime must await the actual code-council disposition and a separate root release.

Exit-record SHA256 values:

- workspace-install: `abd9e0ee7e65c82503d255fcd8ec5c4c1752715201183d1803774b645def1d9b`
- bind-workspace-tools: `208be05f3900a7721dadee466e5ab88ebe0824f6a639b532a59b4d48450c14ce`
- typecheck: `cb08162c105615736cd5b02817d3f58f9f8dc50716aa4560873bc436b5400347`
- format-check: `73dd8234406a4ff28f29cf9c793ebc9e86744aa94c6c65570978376daef0747f`
- lint: `dd0082228a434b0b001428aeea56039ee2a4e807a845cd028a41d6d8293de35b`

Remote session guard passed before this sole supporting report write. Prior evidence and all source/harness/gate bytes were preserved.
