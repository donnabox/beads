# Independent bootstrap/static prefix audit

Completed read-only audit of terminal records only. PASS: fresh workspace-install, bind-workspace-tools and typecheck. FAIL: format-check raw1/acceptedfalse; no lint or build attempt exists. This is a stopped failed prefix, not successful static validation. No test/install/build/import/formatter/helper/gate was executed by this audit; only files, Git objects and process state were read, followed by guarded report writes.

- workspace-install: raw0, accepted=true, 5.135661s.
- bind-workspace-tools: raw0, accepted=true, 9.454817s.
- typecheck: raw0, accepted=true, 5.949475s.
- format-check: raw1, accepted=false, 5.057139s.

All four start/exit records match the concrete commands, base HEAD `b76c3b7df77ea9b33611bc9027db651833e20abb`, tree `4d9f8ac00702d32a45c3eb38cf50fac55ac96f2d` and source identity `6d547d6c57fcc0f1017770cc97c2b0b977b859162aeb78538c61289be312fb7c`. Full six-file patch `6f35130f0fe0a8c91e89e48585e138bad92d83be2d9d8844bb08f0b5492f5320` and all679 tracked source hashes match before/after. No untracked source. The source remains frozen for the active Claude review.

Fresh bootstrap is honestly attributed to this successor source: install log reports offline frozen-lock install across12 workspace projects,61 reused packages,0 downloads, scripts disabled in the recorded command. Workspace tools JSON `c59801cdde0c1e16242840e1c2a9aa29040d44e3dbc07c55d63c3b183b6b75f8` records Biome2.5.7, Vitest4.1.10 and TypeScript7.0.2; all11 recorded tool paths/realpaths/content hashes independently match, including actual native Biome/TypeScript binaries. Install log/exit linkage and source lock `4150d8821c96dbc2cfec0b2762ff76b58318079f65fa4f5c8858ef2dd61f8d6d` match. There is no inherited predecessor install claim.

The pinned implementation index remains unchanged and matches every indexed file. Completed logs all match their recorded hashes. Each stage reports sourceUnchanged=true, no forced cleanup/errors/process/store survivors; independently all50 sampled PIDs are currently absent and stage TMPDIRs contain no bdp-prefixed directories. No runtime behavior has been validated.

Format diagnosis: exactly three test files have formatting diagnostics, across eight displayed formatting regions. There is no production file diagnostic. The log contains complete before/after formatting text; it says Checked6files/No fixes applied/Found3errors. Raw failure/log preserved. A separate CLOSED formatter-only successor proposal is `formatting-successor-plan.md`; no source change is authorized by this audit.

## Evidence fingerprints

- `run/workspace-install.started.json` — `3257934cb67e2ef13539064ad36f9374e5078f4e16be80a587bedb2d8cf3f733`
- `run/workspace-install.exit.json` — `bd232d1467fe659f1857cb44f91eaf608e1191f5651698530eae4d439d8dbf54`
- `run/workspace-install.log` — `1a579e3771868420bb3eb10221fb9c354db058f4cf8ba2dd89122107308ceb79`
- `run/workspace-install.processes.jsonl` — `351f56ccb1bf7c97855fcbee504eccd9f84b0d8b66a8c6c5b03bae119ec4997c`
- `run/bind-workspace-tools.started.json` — `9fc5b223769f14eab272163c373c5f395997fe8cba75da5d0e5767e6ebf1cf2f`
- `run/bind-workspace-tools.exit.json` — `6a09435b5be2ba196fbb41ea9d0912196466659238965afe18f16c5ca119cd18`
- `run/bind-workspace-tools.log` — `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
- `run/bind-workspace-tools.processes.jsonl` — `33362debba407e4bebfe27f30c5d18442d14f550c95807f9e00b49826efb4bae`
- `run/typecheck.started.json` — `c829a74f03807cbb0204e71f2da981757bc924664669e41a080526455e8c3a40`
- `run/typecheck.exit.json` — `4cd752b13bc024edd8fa549a39c55a65456779566c20e93c414c34ebfbf88a36`
- `run/typecheck.log` — `91d54ffbee73a3ac72294d090e58d77745f8b93edbf94b862807698273a6505a`
- `run/typecheck.processes.jsonl` — `57f7cf62d9f41ae2d5df6cfbf4778af940cfb39fa17ce5a5bb31357f5d063519`
- `run/format-check.started.json` — `fdf43996e480b4240d74b43707aeaaa7656385e3fa2c11aa00f0e343e920fbe4`
- `run/format-check.exit.json` — `7dc5ffca21348b7d0038d26cf6919fe786abe9652393fda153faf2a6689bb45a`
- `run/format-check.log` — `09a4531ec83f8b2de1428405a8530c0eeca8bcc2f20f8f6e6a38ffd6b2ad6ff4`
- `run/format-check.processes.jsonl` — `931491509eddfe9e0cf4c113c732f675610407b68565e80afe9c7ec2b0689a81`
