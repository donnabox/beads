# Actual prefix audit — author successor-3

Read-only independent audit at 2026-10-17T02:28:06.013421+00:00. PASS for the seven completed stages through `review-focused-source`; no execution performed by this audit. Source remains `502d8fb13e471353de934ca6f11ad22aa76486aa1f3f6481f964222b67ca5458` on `b76c3b7df77ea9b33611bc9027db651833e20abb`. All 679 tracked files, exact full dirty patch and current HEAD independently rehashed; no untracked files.

- typecheck: raw 0, accepted; 4.443s; exit SHA256 `6a5dac887424d223184de5e86fa531b5f6ce763a9e201d2cc50a04c18f561bb7`.
- format-check: raw 0, accepted; 3.606s; exit SHA256 `b8c8c6bb78abb2693acdf7334d9014a9a688ed354513d8a4e621f8dac3240d71`.
- lint: raw 0, accepted; 3.639s; exit SHA256 `1998c45c466c3baf5db5d65b3ce59a2c7cb1745775e55506b102697edfce3032`.
- fresh-build: raw 0, accepted; 4.448s; exit SHA256 `c3b21c946b92317f9b940e17823f27ce35677479ab879e9258d7d37d2107b276`.
- bind-fresh-build: raw 0, accepted; 7.707s; exit SHA256 `8d5b73f1ab1e0ee5ae7992ed5b6de758ef773d1938c8e8687fb9c3f762ed87e8`.
- focused-source: raw 0, accepted; 5.849s; exit SHA256 `4ba20bdb7578e598b67710085300177b011498aaa870c91462a43a5d9f48d445`.
- review-focused-source: raw 0, accepted; 7.648s; exit SHA256 `22845282dfccd940e3a6571dbb41d7e2f7fa6ac4db8f25ab681103e5c89aa838`.

Lint passed with **19 warnings and 1 info**, retained as warnings rather than a clean-diagnostics claim. No fixes applied. Previous failed format/lint records remain separate and unchanged.

Fresh build log explicitly records `pnpm clean && tsc -b` and `tsc -b --clean && node scripts/clean-build.mjs`. Independently rehashed all 476 emitted artifact files and the build start/exit/log links. Build manifest SHA256 `9058fbca00b58a056cadfe5dd794ef25a48aa4f12467b9e87752ae647eddfa83`. Fresh build and binding preceded focused-source collection; no stale emitted-build evidence is reused.

Exactly 13 selected callbacks passed; 97 unselected callbacks skipped. Exact full names compared with the pinned command selection; result SHA256 `d9bad3ec71cd9dc6e54fd0e908359211779bb3d11546e394f58ec35b90a130af`. Accepted observation SHA256 `e3f65da9d13ed7e0e6a02f3e2b5a567c3980a8f95f96e31c91556740278b0f42`. Nine actual JSON observation records match the audit artifact: owner 8 reads/1 seed; capture 4 reads/2 service calls/2 waits/2 probes with attempts [0,0,5,10]; cancel/drain 5 reads/4 service calls/3 waits/3 probes with attempts [0,0,5,10,15] and one post-close direct refusal; both service observations report zero retained group reads. Five preflight rows each report one service call, zero owner reads/waits, and four clocks total. New HTTP callback reports **32 calls**, zero product children, removed directory and zero failures. The historical HTTP callback passed, but its **59 calls remain source-derived workload**, not a separately observed wire census; consequently 91 HTTP is prospective combined accounting, not a measured total.

All 85 distinct sampled PIDs from these seven stages independently absent at audit. Every terminal record reports no forced cleanup, errors, survivors or remaining runtime directories. This prefix does not launch installed product children. Source callbacks' cleanup assertions and supervisor records are the available lifecycle evidence; passing exact callbacks supplies semantic evidence, not profile qualification.

Bootstrap remains explicitly reused from the original implementation attempt, source identity `6d547d6c57fcc0f1017770cc97c2b0b977b859162aeb78538c61289be312fb7c`. It is not relabeled as an install on successor-3. Reverified all 17 dependency inputs and 11 actual tool path/realpath/hash identities; the bound old-to-new delta remains three test files. No reinstall or prior static-result reuse claimed.

Packaging/external stages are outside this audit, including any currently active files. No default-mode installed callback or complete write-profile claim is made. Root owns remaining releases and final acceptance.
