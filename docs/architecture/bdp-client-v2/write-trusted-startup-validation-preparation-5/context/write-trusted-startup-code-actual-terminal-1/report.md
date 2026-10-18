# Independent startup CODE terminal audit

The single authorized attempt produced one complete actual Claude result with ten findings, then **failed supervised finalization**. The preserved outer classification is `process_ownership_or_cleanup_failure`, wrapper exit `-15`, elapsed 642.010743 seconds, `ownershipUnknown: true`, `finalRoleVerified: false`, and error `reviewer_identity:RuntimeError`. This is usable complete substantive evidence for root disposition, conditional on root accepting it as such; it is not a clean review run or approval. No retry occurred.

## Actual provenance and failure boundary

Reviewer PID/PGID 51716, wrapper51688 and supervisor51679 correspond to the original process record beginning 2026-10-18T04:31:58.000396Z. Live observation independently saw these processes. Default initialization was `claude-opus-5[1m]`, Read/Glob/Grep only, empty MCP; actual assistant tools were 20 Read and 17 Grep calls. No model override, subagent tool, execution tool or permission denial was observed. CLI usage also reports an auxiliary Haiku accounting row; that is preserved separately and does not imply a reviewer model override.

The 1,451,853-byte raw stream SHA256 `959b47c76c8f32846eb7a53bbc4e3828ed1973e838aa58a31dab85db68418768` consists of 2,519 valid newline-terminated JSON records. Exactly one terminal result is subtype success, is_error false, end_turn, 38 turns, duration638980ms. The complete 9,782-byte result is extracted verbatim into `reviewed-result.txt` (SHA256 `40ae1279e4747e6c245d8302df6d7febc99a9fe4dca25b01f9fd0d4e0c665502`); `terminal-result.json` preserves its complete structured record. These auditor artifacts do not reconstruct missing runner outputs.

Runner progress received that terminal at639.814403s. Outer progress still showed the live reviewer at639.821111s, then signaled wrapper51688 TERM at640.840758s; later STOP/KILL passes signaled no groups. The last retained reviewer row is zombie, with exactly the prior authenticated PID/group/start/argv identity. The supervisor's generic RuntimeError omits the failing require message and triggering intermediate inventory. The last zombie row therefore does **not** establish the cause of the earlier identity failure. No PermissionError or unanchored known-group observation was recorded. Preserve this unresolved supervisor diagnostic limitation; do not infer that the earlier zombie defect necessarily recurred.

Runner `exit.json`, `source-after.json`, `final-review.md`, and `terminal-results.json` are absent. Wrapper stdout/stderr and reviewer stderr are empty. All42 periodic outer role checks exited0, but finalRoleVerified remains false. Root separately observed execsession38477 close exit1; this auditor did not poll the root-owned session or infer CLI exit0 from the success payload.

## Inputs and process closure

Independent terminal rehash matched all682 current source files, 98 contexts and9 artifacts against both binding and actual prelaunch source-before. All679 source identities in the original frozen checkout remain unchanged. Both HEAD refs still resolve, by direct metadata read, to a56254d682cf4623cab22a552e3d2f695f691795. No Git command or postterminal live-status recomputation was performed: recorded prelaunch status and index-transition evidence remain the status provenance, and independent hashes are explicitly attributed to this auditor rather than a missing runner postcheck.

The actual source patch remains `00fa952644ab1725e4bfb8fc8c642683a39f60dd45f6a08de8881d590150a0b4`. The 124-entry preparation index and106-entry frozen packet index retain their closed meanings. Only binding.json and launch-supervision/control.json differ from their indexed bytes: each is the exact single literal false→true rootLaunchReleased replacement of the release archive. Root release SHA256 is `132aca59eff6ca15a3a598b16dc3ed9a53575611d083fb51bba75a168d64f3e1`. All125 mirrored preparation files match, with exactly the same two authorized release exceptions; all other frozen inputs match. Runtime outputs are separately inventoried, not inserted into closed indices.

A fresh successful full OS inventory checked all106 observed PID numbers, including supervisor, and45 observed groups; none was present. This is actual postterminal absence at the audit timestamp, not timer inference. It cannot repair historical ownershipUnknown or prove absence of never-observed descendants. Current remote role claim remains exact session2026-10-18T02:16:24Z-549f at audit boundaries. No supersession was observed.

## Independent finding crosscheck

**F1 — Medium, reference-deployment.ts:40,150; config/index.ts:570–573.** Confirmed: real config.bdptest contains deploymentFile, but exact nested allowlist accepts fixture only. A structurally assignable real section is refused before open. Current loader is still an isolated future-consumer seam; this is a concrete integration hazard, not evidence that an existing listener ran. Accept deploymentFile in nested type/allowlist, preserve fixture conflict; witness deploymentFile-only success and fixture+deploymentFile refusal. If equality is added, make the chosen policy explicit.

**F2 — Low–Medium caveat, reference-deployment.ts:269,292–298; installed-reference-configuration.ts:170–172.** Confirmed public SHA identities commit to secret-containing content. Existing accepted provenance permits them; absence of literal tokens in JSON does not imply secrecy. Accept documentation saying digests commit to contents and must not be diagnostic-emitted. Do not introduce a salting/hash-identity change under this bounded correction.

**F3 — Low, reference-deployment.ts:154; transaction-runtime-configuration.ts:118.** Confirmed TX loopback is a duplicate literal, currently equal to the authoritative default. Reuse TRANSACTION_RUNTIME_DEFAULTS.transport.host; preserve existing accepted host behavior.

**F4 — Low, reference-deployment.ts:160–173,321–324; transaction-runtime-configuration.ts:334–340.** Confirmed duplicate page formula. Reviewer proposed relying solely on the postcapture check, which would move unsupported-limit refusal after file access and conflicts with the accepted pre-file refusal order. Retain the explicit pre-file unsupported tuple refusal and authoritative captured readOptions comparison. Defer abstraction unless both timing and current behavior stay explicit.

**F5 — Low, reference-deployment.ts:317–320; transaction-runtime-configuration.ts:307–308.** Confirmed heartbeat check duplicates liveStreamsAvailable and seams is empty at this call site. Harmless future-refactor tripwires. Retain/defer or explain briefly; no correctness blocker.

**F6 — Low, reference-deployment.ts:182–198.** Confirmed unconditional 32 MiB+1 allocation (33,554,433 bytes), bounded and intentional; stat is not the real growth bound. Retain/defer. Optional future allocation optimization must preserve one-handle reads, stale-stat growth ceiling and close discipline; no bound change now.

**F7 — Low witness gap, reference-deployment.test.ts:340–355,601–620.** Confirmed calls.length is only used as a no-reread baseline and close spy uses mockImplementationOnce; neither proves exactly one open. Accept an explicit exactly-one-open assertion, resetting/isolating counts appropriately.

**F8 — Low witness gaps, reference-deployment.test.ts:340–355,591–600.** Confirmed mutation targets page.maximumItems but assertions omit preflightReadLimits; MAX+1 refusal exists, exact-MAX acceptance and empty syntax are not in these witnesses. Accept preflightReadLimits snapshot assertion and a small empty-file syntax witness. Exact-MAX accept is optional; if added use a valid envelope plus trailing whitespace to isolate the outer byte boundary.

**F9 — Low, config/index.ts:1080–1094; reference-deployment.ts:99–110.** Confirmed equivalent path grammar across the package boundary. Server does not depend on config; duplicate validation is structurally justified. Retain/defer or cross-reference comments; do not add a new runtime/package dependency.

**F10 — Informational, reference-deployment.ts:213,251,262–301.** Confirmed fixed non-echo code policy: explicit internal refuses remain internal; arbitrary capture defects under contract phase are masked as contract. No demonstrated boundary breach. Retain/defer; optional error policy prose, no new semantics needed.

The reviewer found no boundary breach of accepted items1–6 and no N1 lossless depth/node containment defect. Its review was static: no compilation, tests, builds, package installation, emitted-declaration inspection or service/runtime observation. The deeper compiler bodies and some existing test regions were not fully inspected. The auditor performed no product execution, source correction, flag mutation, reviewer launch or release. Root owns substantive disposition and all subsequent authoring/validation.

The auditor initially assumed mirror flags remained closed; that local audit assertion failed because root had mirrored the same two releases. `auditor-preparation-failure-1.json` preserves the error. The corrected comparison checks precisely those two exceptions and byte equality to their released originals; no review input changed.
