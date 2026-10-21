# Independent completed-prefix audit

Verdict: seven completed successor stages accepted; no prefix blocker. This audit deliberately excludes active packaging, consumer and installed stages and makes no claim about their outcomes.

Fresh typecheck, owned-three-file format and lint passed. Lint is not warning-free:47 warnings and5 informational diagnostics were reported, with no fixes applied and exit0. Fresh build log explicitly records pnpm clean followed by tsc-b; clean itself invokes tsc clean and clean-build.mjs. Build binding preceded HTTP suite collection. Independently recomputed476 emitted/artifact file hashes, checked the exact build source identity and pinned build start/exit/log bytes; all match the current candidate and recorded build inventory.

Exact focused selection:7 passed and79 skipped in the one HTTP suite, no unexpected callback executed. The actual names are:

- preserves concrete query syntax failures and accepts the reserved genesis minimum
- serves canonical current Type and incident minimums with truthful context and parser boundaries
- rejects malformed minimum tuples across read routes without entering the actual owner
- waits for current Resources and initial collections with unknown minima and preserves residual refusals
- rejects collection parser errors before unknown minimum support checks
- preserves parser refusal parity on noncollection and excluded query routes
- preserves receiving and HTTP error precedence for parser refusals

Observed success supports assertions of four direct source-handler callbacks and three emitted-runtime fixture callbacks against one freshly bound candidate.100 handler calls (67 new+33 existing),91 HTTP exchanges and7 stores remain source-derived workload, not new wire telemetry. No counts are profile qualification or completion percentages. The review-focused-source artifact correctly records this distinction.

Every completed stage has actual exit0, accepted=true, matching start/end successor source identity, unchanged source, matching log hash, no forced cleanup, no errors, no recorded surviving processes or owned runtime directories. Independently checked all prefix temporary directories for surviving bdp-prefixed stores and current ps for104 sampled PIDs; none remained. This check did not inspect or affect active packaging processes. Every sampled prefix duration remained below120 seconds; no retry or interruption was used.

Source identity1a3caa86808b72da3488c1a9fe34f05ca7b7ce244b75d7d537137b37fe5594ae matches all679 tracked files at base d4f3f4be88c7837b71d6646518ca3589d464ba3d and the frozen three-file patch. No untracked source was present. Full source hash map, actual HEAD and binary diff hash were independently verified.

Bootstrap attribution remains correct: original offline frozen-lock/scripts-disabled install and tool capture belong to source identity f48ed2633f62f4722f3bb9fe99ddc0ab750159b029afdb5df12a26dd931b262d. The original tool record retains that identity; it is not mislabeled as successor execution. Original record hashes, current tool realpaths/bytes, identical lock/manifests/settings and exact two-test-file transition to the successor were independently checked. Fresh successor static passes above are new executions; original successful typecheck/failed format evidence is preserved.

Stage elapsed seconds and immutable outcome/log hashes:

- typecheck: 5.869395s; exit record `350dd53c3231776f182e0263e81093ecdf3354c8e3d4d4f743834c04375f3ccd`; log `91d54ffbee73a3ac72294d090e58d77745f8b93edbf94b862807698273a6505a`.
- format-check: 4.822350s; exit record `d35d050f69e24f50420590ebedb10da87926a211011758708c90ae26445e5b80`; log `234b80584535ecf3447bab9d5b896282b4dc7105a864eb82cc425183c152cb20`.
- lint: 4.985686s; exit record `fdfb65305214a4ef834671702efe599a947a84e00a4bcc374248a269b186e6f4`; log `594753149cde709b7f5810f78fffa7378478ed535c767ea1a451c740203c87fd`.
- fresh-build: 5.927175s; exit record `c00cabfcd1c242b2c3822b0c5d436b66b9808dffbeedc7ba85640d9142039a5f`; log `232dac5e840e45b3bb06e6e04249b5241f2e1b4f6c797a22c046d916302ac208`.
- bind-fresh-build: 10.159238s; exit record `a563f165af1df70deeed1c2c457d1ebc18429046cf670b99e0c5387dc3729789`; log `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.
- focused-source: 6.192820s; exit record `c833c3192e61ef4b641e9b4193093637365bc215251d9cef9f5958d4bd3c65c4`; log `1fd339cabbb22d36bff8858ba358eed81fbe9efb7e71e9f82347f7b0c0920f42`.
- review-focused-source: 10.014253s; exit record `a6468f3107cd00340bc4795282da1e76e66fc9c41461c80205094d5439f708bf`; log `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.

Other bound artifacts:

- `source-binding.json` `5fecd704d54e08b0de899505decf8da1ae80f6ad64fdcbc90608e8a516f715db`
- `bootstrap-reuse-binding.json` `eca0ce4463b8ead4d4777841cc881f53a74a7d8a3804a6026bd04c1200b51017`
- `build-artifacts.json` `806c426e70c1fb092a7937486476705e5fd7d2826b715a07307658819e8b707a`
- `focused-source.results.json` `96102517ec242bbc7631c5f5ee498506510adb52d7a94e6cd7a9483206fd4dff`
- `focused-source.accepted-observation.json` `800f0614aabce22fd26eff252be725abb05ac4e14d482dd264e13dc038a393ce`

Only this report was written under the remote-session guard. No test, import, build, install, service, helper/source modification or process interruption was performed by this audit. An initial report-data assertion expected the runner’s unselected status to be pending; actual JSON says skipped. That check failed before writing, was corrected to observed skipped, and did not rerun any stage. Await separate completion and release for installed evidence.
