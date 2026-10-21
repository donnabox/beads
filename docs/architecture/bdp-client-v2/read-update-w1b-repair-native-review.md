# W1b repair independent native review

**Verdict: bounded pass; W1b findings F1–F5 closed as scoped. No new blocker.** Exact frozen `c1a17442ff2f5936616e7644e827dfa28dcd5ab7`, direct parent `3e2b8784daefd5893f990785cfe91f1d872a2ffe`. Branch/checkout name says W2, but the reviewed eight-file diff contains W1b repairs and private harness capacity only; no sequence implementation.

Read-only review of `vickie-read-update-w2`, W1b Claude report and the final author repair report. Independent clone `bdp-read-update-w1b-repair-native` remains clean at that head. No author logs overwritten or source/evidence/publication changes.

## Closure and claims

- **F1 Medium closed.** Each of the three owned-Link scenarios now tests accepted current Link guards and rejected superseded guards after properties and metadata mutations. The guard is taken from an independent current GET; the stale guard is the immediately preceding version's captured revision. Each request has its own key, expected 200 or 409 revision-mismatch, canonical response schema, and a following full independent Link GET equal to the expected unchanged record. The no-op row adds the same current/stale contrast after retained-envelope behavior; the unowned update carries its own row-local pair. This fills the original native review's blind spot: source revision advancement was previously supported, while the Link's own advancement was not independently discriminated. The new stale-problem corruption substitutes schema-valid alias-path-taken and fails exactly revision-mismatch.
- **Catalog restoration valid.** `catalog/read-update-v1.json:253,274` restores the existing matrix title about the Link's own revision and cites the existing fresh-revision sentence. Spec itself and normative semantics are unchanged; citation/profile tests pass. No scenario IDs or profile tags changed.
- **F2 closed without inflated credit.** Attribution stays omitted/wholeContract false, with an accurate reason for existing cross-row minting/no-op observations and missing row-local attribution, absent-envelope/actor and wider obligations.
- **F3 closed.** The post-deletion control reinserts a real schema-valid Link into the empty explicit ownership plane; it fails exactly empty-owned-plane.
- **F4 closed.** A source-only unowned result fails both body-schema and no-source, directly exercising the dependentRequired biconditional rather than only absent-both/present-both.
- **F5 closed.** Every owned row pins the entire served ownsOutgoing member: wildcard max8, explicit relates max1, explicit cites label/max8. Prose distinguishes explicit precedence from unqualified wildcard behavior.

Structural comparison confirms only the three owned scenarios changed, adding guard/GET pairs and replacing the ownership assertion. Other scenarios were untouched. The frozen inventory remains **36 bounded, 27 proposed whole, 102 uncovered of 129**; this repair grants no new row credit. The inherited b6 and other explicitly deferred low findings remain separate and are not new W1b defects. Claim eligibility stays false.

## Private 4 MiB capacity boundary

`artifact-bundle.ts:128–139` raises the manifest copy budget only when the private cumulative factory supplies a canonical branded Read+Update, no-History selection. Catalog/fixture copies keep their default1MiB. Public Read and development factories call the ordinary default path. The matching source-private planner admits up to4MiB only for that baseline selection and re-applies1MiB to other catalog carriers. A noncumulative planner source may therefore be parsed under the finite4MiB staging bound before rejection by catalog ID; it is never admitted with the larger budget. Shared JSON structural bounds stay unchanged at depth128, nodes100000, container entries10000.

`copySourceBytes`' optional bound and the constant are internal module exports, not new public package/index exports. Package exports and public admission remain unchanged. The source capacity increase is neither a service limit nor a profile promise.

Committed boundary tests independently passed acceptance at1MiB+1 and exactly4MiB, refusal at4MiB+1, ordinary/development manifest refusal, fixture refusal and ordinary planner refusal. An additional external native probe exercised the actual emitted private cumulative factory with excessive depth/width/nodes, default catalog overflow, a forged selection, and checked public index absence of the new bound/helper: all refused/absent as expected. Probe: `read-update-w1b-repair-native-bounds.mjs`, `.log`, `.exit` beside this report.

## Independent validation

Pinned Node24.16.0 / pnpm11.20.0; frozen install and completed build before emitted tests.

**89 tests /5 files passed, 26.59s, maxWorkers2:**

- apps/bdptest/src/read-update-matrix.test.ts (actual installed IPC/factory and HTTP, all new and inherited controls)
- packages/conformance/src/cumulative-artifact.test.ts
- packages/conformance/src/artifact-bundle.test.ts
- packages/conformance/src/read-update-catalog.test.ts
- packages/conformance/src/profile-catalog-boundaries.test.ts

Build, tests and external bounds probe exited0. Logs: `read-update-w1b-repair-native-{clone,install,build,tests,bounds}.log`, build/tests/bounds exit files, tests JSON. Separate clean observation report: `read-update-w1b-repair-native-progress.json` (exact c1a1744, sourceDirty false, claimEligible false, counterfactualApplications0).

Bindings match the author's fresh report:

- catalog e26d3fc758a34ebb0b5028be15d28e3e09db87dc7984bba7bad6a00b26329aa0
- manifest db1a06300146d843ed2855be3f9330502078903ca047b1fb0c110a7777ce2620
- fixture 3a60db8714da2490ad67f9bf4f281f50da293393b31c6e1d5bec1ece12fd955c
- exact configuration f04e98d80971e813b73d1b57d225c0a6bfb0669647d6bbf9b026a908bbf37f26

No independent packed-tarball rerun, broad workspace suite, smoke, or evidence generation was performed. The author's packed57-test result was read and is distinct from this independent emitted-target89-test result. This review does not approve every inherited whole-row claim or public profile qualification. Source and historical evidence remain untouched.
