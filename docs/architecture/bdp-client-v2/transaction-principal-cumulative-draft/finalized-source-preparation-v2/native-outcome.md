# Principal cumulative source outcome audit

No discrepancy found in the completed source-only evidence. I authored the implementation, format-only repair and validation preparation; this is a separate audit of actual retained results, **not independent code approval**. No rerun, import, build, service or product edit was performed for this audit.

The nine raw exits and supervisor exits are all zero. Each recorded argv matches the frozen commands; source hashes are unchanged before/after each stage. The current checkout remains the four-file dirty draft at 5d7fa3f049fa0b4db3aa8dae20e20d8e47fb3854, binding 0dbae7cb6139263cbfd3931e80d8dec2c628d7e982deecdd6c8242573a35c6e2, full diff 471e354a504a45cdd785e3c3be7ebbab99c657122b616d00d14b14bfc954bdb4. Gate, preparation and root-release hashes match.

| Stage | Raw / supervisor exit | Seconds |
|---|---|---|
| offline-setup | 0 / 0 | 0.784 |
| typecheck | 0 / 0 | 1.602 |
| build | 0 / 0 | 1.515 |
| format-check | 0 / 0 | 0.981 |
| lint | 0 / 0 | 1.090 |
| boundaries | 0 / 0 | 1.251 |
| read-seal | 0 / 0 | 1.028 |
| static-cases | 0 / 0 | 1.546 |
| build-inventory | 0 / 0 | 0.564 |

Total recorded stage duration: 10.360 seconds. Vitest raw output and JSON independently agree on **9 passed, 96 skipped, 2 files passed**; exact file/full-name pairs match all nine approved selections, with no extra executed callback or reported runtime error. The selected callbacks are CPU-only; this is not installed or network qualification.

Format-check succeeds with one existing oversized Read+Update JSON warning. Lint succeeds with **897 warnings and 56 infos**, with 933 diagnostics omitted by its display limit; this is not a warning-free claim. Boundaries reports 11 packages. The Read evidence verifier reports 89 rows.

I independently rehashed and checked the complete path sets for **672 tracked build inputs, 312 protocol/server package output candidates, and 440 emitted files**, as well as the bound build start/exit/log records. Build input bytes still match the actual dirty-build start. This does not relabel the build as clean-candidate packaging, archives or consumer evidence.

All **39 distinct sampled PIDs** are absent in a fresh process listing. Every stage records no forced cleanup, timeout, residual group member or runtime-root leak. Current attempt directories retain only Node compile-cache entries; static-cases and inventory temp directories are empty. The per-stage empty/runtime classifications and raw process samples remain available. Sampling cannot establish absence of every hypothetical unsampled detached process.

The predecessor format-check failure remains intact with its original nonzero raw/supervisor exits and original app hash 6318eccb; its report was not overwritten or reclassified. V2 proves the corrected source selection only. Actual Claude council, fresh installed cohorts and eventual exact-head CI remain separate gates.

Evidence pins:
- run/static-cases.results.json: e13c073ac8011fa493abb2b8699f46a3899ba6c69975b567fddc2ac8e25d212e
- run/static-cases.log: fa2f02dfc9526d9819298dd8dc83b86e93e420bec43a3836c2a7c3434d91f113
- run/build-output-inventory.json: e6d7d949416d0d8c540d5675d134ad91109a467997d7453316cf5464599aeffb
- run/lint.log: 89312d3fb7618ef5a2a8c26caa5901b23e98d970bac44cf70c31f750b3f68a35
- run/format-check.log: 1bd9e208a36baec9eef70a9a6fde8d026ae38555c38e6d74dc111731ab713f37
- root-source-validation-release.json: 2a01a6d3a1960da136546aa481fa39f1c43bc63875586d04195251cbcfbe68ea

Audit timestamp: 2026-10-15T22:27:07.838852+00:00.
