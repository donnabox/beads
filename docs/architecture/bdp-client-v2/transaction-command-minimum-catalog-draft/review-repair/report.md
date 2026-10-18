# Catalog review repair — unexecuted successor draft

Workspace /Users/dbox/repos/wt/vickie/vickie-command-minimum-catalog; unchanged HEAD bd2178dd6f4b9f7c2d239624d7edaa3d961c3e26, tree 062c32778217aeb702b2f7798dc82b0d985e9c67. Original eleven-file draft binding755192feb51c3fb063a48f1d5189ee8194fac32a0a4ee5e3f49a556844d6c010 and every original draft/review artifact remain untouched. This is an uncommitted successor; no validation is claimed.

Only two files differ from that draft:
- profile-contract-plan.test.ts: the existing new mapping callback now asserts actual loaded RU required count130, coverage rows130, rows with !wholeContract97, status omitted76. Same callback title; no extra registration.
- read-update-matrix.test.ts: exactly14336→14579 and9728→9778. Append rationale for the401B new omitted row and respective243B/50B excess; preserve698614/703415 historical bases and both earlier schema-amendment comments. Other numeric/residual fences unchanged. Zero slack is intentional, not an accidental missing margin.

Source inspection supporting F4: artifact-set.ts217–244 requires complete canonical coverage with the selected required count/order, and absent plans must be omitted/false/empty;249–270 requires planned rows bounded-observation. read-update-matrix.ts955–959 derives uncovered IDs by required ID lookup with !wholeContract; apps/bdptest/test-support/read-update-artifact-execution.ts329 derives omitted IDs by owner===null. Thus the added static counts agree with the actual derivations; they do not execute installed report construction. Pure JSON inspection independently found130/97/76, not an execution of the new assertions.

F1/F2/F5/F6/F7 broader proof gaps are root-dispositioned through separate native full-source audit; this repair does not claim additional source/runtime coverage for them. F3 exact zero residual is the explicit root-approved guard decision. Optional editorial findings are not broadened into this two-file repair.

JSON/operative documentation unchanged from original draft. Coverage remains713193B (parent712792 +401), SHA256 f8f0b001928f8488292403b57c973bd918cba224ef9e00fe0691482bf5cfac0c; original catalog/development/coverage insertion bytes and all preservation proofs still apply. Amended arithmetic:698614+14579=713193 and703415+9778=713193. Numeric ceiling714717 and residual714369 unchanged. No fixture/manifest/runtime/parser/wire/time allowance changes.

Original selection.json SHA58e834e4aa5e9d53af0c940b7412c1e5ef9b5425ab076af902d2f475e02af162 remains unchanged:12 prospective selected cases, one new callback, zero HTTP/listeners/stores/product children. No tests/build/typecheck/format/lint/install/package/services were run. Formatting/type success remains unverified pending release.

Excluded installed callbacks, including their edited counts, are NOT executed by the local12 selection:
- RU “executes the partial cumulative RU matrix against actual installed factory HTTP and afterstate”: uncovered97 in its report and progress assertions; omittedRequired76.
- RU “observes retained retry dispositions across actual installed carriers and current positions”: uncovered97.
- RU “observes key and media admission against installed HTTP with paired afterstate”: uncovered97.
- RU “observes reference syntax and retained category failures against installed HTTP”: uncovered97.
- TX “observes thirteen partial Transactional contracts through packed-capable actual installed HTTP”: required251 and report256.
Exact-head full CI tests these later. The local static closure supports count arithmetic only, not installed-runtime evidence or profile qualification. README/planning prose is inspected metadata, not a runtime result.

Artifacts: full draft.patch SHAb9abce1c55b865ba389ed53d7c7b86f9bc854287cfec373158d54b9440dea489, original→successor correction.patch SHA97fab9b4c8708b7e604eea0b75166de62bafc064b544aca5126f9f84ca0ea953, complete11-path source-binding.json SHAe6a3147f6a76028c5ca98726f5ba0bce62f431955bf359c90eaed7bffdba3270. source/ contains exact successor snapshots. All other nine original draft path hashes match. No commit/push/PR/coordination write; guard preceded the two source edits and this artifact batch. Stop here for root/native inspection and measured validation release.
