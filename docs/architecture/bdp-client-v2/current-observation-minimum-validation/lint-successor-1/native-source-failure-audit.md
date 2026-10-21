# Native actual source failure audit

Confirmed first failure: focused-source exited raw1/accepted false after 11.708703s, with 96 passing selected cases and one failing selected case. The 348 other enumerated tests were skipped; they are not passes. No immediate focused-source audit or packaging stage started. No installed or consumer result is established.

The failed callback is “serves current observation minimums with actual issued pages and HTTP parity”, transaction-http.test.ts:6339. Its Type inventory expectation is Task,Relation, while the actual HTTP result is Relation,Task. This is a test-oracle defect. The previous completed CODE F2 recommendation and my acceptance were incorrect: authority-read.ts:808–813 preserves descriptor input order at its internal page seam, but HTTP index.ts:1266–1269 calls sortCollectionResult; index.ts:1365–1370 applies inCanonicalUriOrder, whose read-request.ts:371–379 contract sorts canonical IDs. relation precedes task. Runtime input order is not wire order. No production change is indicated.

Minimal proposed correction, subject to root source release: swap only the two explicit E expected inventory objects back to Relation,Task. Keep exact equality, all descriptor fields, all requests and installed expectation unchanged. No sorting of actual output, weakened set comparison or new traffic is needed. Preserve this failed attempt and the historical erroneous F2 disposition; do not rewrite either as successful evidence. Native exact-delta closure can establish this one-block correction; fresh validation is still required.

Independently rehashed all679 current source files against normalized binding fc9206f9 and all six completed prefix log hashes. Typecheck4.895002s, format4.187468s, lint4.579739s, fresh-build4.914446s and bind-fresh-build9.011643s were raw0/accepted. Lint still has165 warnings and9 infos, not a zero-warning result. All six exit records report unchanged source cfed12d8, no forced cleanup, errors, remaining processes or runtime directories. A fresh ps census found all60 sampled prefix PIDs absent. Original two bootstrap prerequisites remain separately attributed; no previous static result was reused.

The source log contains nine new and28 legacy observation records, but record presence is not a successful audit. E is explicitly partial:11 attempted/completed responses,2 waits,12 clock calls, failureCount1, closed true and directoryAbsent true. Its remaining26 planned requests did not execute. A reports25 operations/35 attempts/10 waits/45 clocks/2 mutations; B12 operations/12 attempts/2 waits/14 clocks; each of five C surfaces reports2 attempts/1 wait/3 clocks and successful close/probe/removal; D reports8 calls,2 opens,2 offline inspections, unchanged payload and one post-close probe. These passing callbacks and their records cannot turn the failed97-case selection into a passing whole. No97-pass, full37-request E, final validation or profile qualification claim is supported.

Evidence pins:
- focused-source.exit.json: c2768a7218549df4c6b8c533b391eb47ed43a1dafbfac7627f69f922699c0752
- focused-source.results.json: 678748a84e6818d2cfffe83321ddbd924f1fac832a966ee24931f337dfb3e9fa
- focused-source.log: 1bbb32e47f201e181c375a984896e73c5a9ca0b07ec4268ad272d7d52ed551a4

Read-only evidence/source inspection only; no helper imports, reruns, source edits or gate mutations.
