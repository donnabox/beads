# PR87 council repairs — frozen source

- Commit: **28b5da2092e57cac53fd67fab38ab9e95b64df28**
- Parent: **9ea408958e08be38cc4e586a4846caf3d6ef6874**
- Checkout: `/Users/dbox/repos/wt/vickie/vickie-profile-contract-council-repairs`
- Branch: `codex/vickie-profile-contract-council-repairs`
- Verified origin: `https://github.com/gastownhall/bdp.git`
- Clean working tree. Original 9ea reviewer checkout unchanged. No push/PR/bus or other outward writes.

## Council dispositions

1. **Preserved native closure.** Own-descriptor snapshot, proxy/accessor/boxed-value rejection, explicit primitive identities and immutable detached output remain. The source-byte input now replaces direct capability-array input; nested fixture/source/manifest getters and proxies are rejection-tested without invocation.
2. **Repaired normative declaration.** RU conformance index now declares `read-update.discovery.document` retiring `read.discovery.document`; TX prose explicitly inherits that earlier retirement alongside twelve direct TX→RU edges. RU metadata/declaration mirror guard added. Existing closed discovery laws unchanged.
3. **Preserved surviving clauses.** RU and TX replacement discovery rows both bind omitted order→canonical-uri and optional limits/maximumEndpointMultiplicity/order/History contracts. Counterfactual tests remove each shared clause at each replacement profile and detect its loss. Deliberately contradicted closed-shape citations are not reintroduced.
4. **Honest carrier status.** Current ordinary normative artifact/runner support is limited to read-v1 on Read; normative RU/TX and unknown catalog carriers report `carrier-unsupported`. Manifest bytes are parsed, so supported versions remain governed by the existing parser. No public artifact or runner admission behavior was enabled.
5. **Prerequisites.** Planner traverses actual parsed prerequisites; missing/retired/unready/late-in-catalog/cyclic prerequisites block readiness. Current Read realizations contain nine dependencies on the retired discovery row at RU, demonstrated from the real manifest after selecting inherited normative rows. Unsupported carrier takes display precedence while blocked prerequisite IDs remain visible. No predicate is counted as an observation.
6. **Canonical loader.** Ordinary `loadProfileContractSelection` has no injected readSource parameter. It reads actual on-disk catalog files. Separately named `inspectProfileContractSourcesForTest` permits malformed/shrunken negative controls, but never brands its selection; such a selection cannot enter `planProfileExecution`.
7. **Directory completeness.** Every JSON catalog file must be explicitly classified core/companion/optional; missing, duplicate and newly unclassified files fail. Tested against actual directory inventory and mutations.
8. **Captured fixture source.** Configuration input now accepts exact fixture source bytes. Existing bounded source and fixture parser derive fixtureDigest/fixtureId/capabilities; independently supplied claims are rejected. Only factory-bound configurations can be compared/planned. Fixture oracle pointers are checked using the existing bundle helper.
9. **Shared canonical checks.** Planner reuses runner's canonical Scope predicate and existing fixture identifier/capability validation. Reproduced encoded-slash, duplicate-slash, needless-unreserved-encoding and whitespace-ID failures reject.
10. **Exact bounded sources.** Manifest and fixture use existing intrinsic Uint8Array copy, bounded JSON parsing and fixture validation. Source wrappers are captured as own data before helper calls. Array-like/nonbyte/Proxy inputs, oversized/deep/invalid-UTF8 input and custom-iterator coercion counterfactuals covered; copied buffers remain unaffected by caller mutation.
11. **Audit lockstep.** The 81-row table is parsed and compared against exact catalog ID order, inherited/retired treatment and replacement IDs. Targeted discovery clause survival is guarded as above. No additional TX retirement was invented.

Helper reuse adds non-root exports in artifact-bundle.ts and runner.ts; neither package entry point nor package exports map changes. Existing helper bodies and public bundle guards remain unchanged. The bounded plan, exact input/output binding schema and next slice are documented in `docs/design/profile-contract-planning.md`.

## Verification

Pinned Node 24.16.0 / pnpm 11.20.0; frozen install passed.

- PASS build and typecheck.
- PASS dependency boundaries, 11 packages.
- PASS full conformance package plus bdptest Read matrix: **816 tests / 25 files**, 15.93s, serial maxWorkers 1. Structured result `profile-contract-council-conformance.json`; verbose log alongside.
- PASS 53 planner tests in focused run, including all repaired native controls.
- PASS formatting, focused lint and git diff --check.
- PASS unchanged sealed evidence verification: 89 rows; constant 674982be2123b9b7aae630bf79c3f774122900cd; no artifact or constant edits/generation.

During authoring, a new prerequisite negative-control fixture initially violated existing manifest rules by converting requests to actions without a family and then making every action prerequisite-owned. Corrected the fixture to valid HTTP prerequisite plus scenario-owned actions; all final runs above passed. No server behavior was loosened to make the test pass.

No full workspace suite, actual RU/TX served qualification or new evidence claim. All planning remains claimEligible:false. Catalog inventories remain 44 normative Read / 129 cumulative RU / 250 cumulative TX; those counts are explanatory, not selection authority.

## Next bounded delivery slice

After review closure, build the actual cumulative RU v2 manifest and fixture, repairing inherited discovery schema/prerequisites and preserving omitted-order behavior. Add a separately reviewed profile-specific artifact path while preserving Read's existing guard. Execute credentialed HTTP and after-mutation reads through the installed P1 factory with exact observed fixture/configuration bindings. Do not bind candidate evidence to provisional 46d8809 while U2 A1–A5 repairs are underway. No further planner expansion is proposed.
