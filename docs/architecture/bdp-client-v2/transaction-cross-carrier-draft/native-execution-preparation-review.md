# Native execution-preparation review

**No remaining blocker in the corrected preparation. Two concrete findings identified during review are closed by the verified v3/v4 preparation corrections.** This is a read-only review of unexecuted external orchestration, not execution release, syntax/type/build verification, package qualification or a claim about runtime success. Root owns release; `execution-gate.json` is absent.

## Findings and closure

**P1 — Medium, closed — cumulative source workload incorrectly recorded as zero HTTP.** The original reviewed `commands.json` recorded zero HTTP for the full39-case cumulative selection; its report described in-memory-only executors. Actual `packages/conformance/src/transactional-cumulative.test.ts:494–653` has six fault kinds flattened over batch/update-where, creating12 loopback listener instances. Every case issues one actual POST; the two retained cases also issue a GET. Expected workload is14 HTTP, maximum24 (two action wires per case),12 listeners and zero installed children. Each case's finally closes all connections and awaits server.close; there is no separately authored listener-cleanup deadline.

Root explicitly approved accounting for this already selected work without another invocation or a timeout change. The final command ledger/report now record it separately from the nine installed cohorts. The exact full39-case argv is unchanged. Expected combined wire work is1655, outer maximum1752; the12 source listener instances remain distinct from189 installed children. The source sequence also selects the same two app static cases, which add no HTTP or installed children. No attempt was executed to establish these figures; they are static expectations and bounds.

**P2 — Medium, closed — final directory snapshot was excluded from runtime-leak classification.** The reviewed supervisor at lines78–83 retained `tempEntriesAfter`, but classified only paths previously seen by the50ms sampler. A runtime directory created after the last sample could remain at exit, appear in the final snapshot and still be omitted from `runtimeDirectoriesStillPresent` and its failure condition. The final supervisor forms the union of sampled still-present paths and the actual final listing before classification. Both raw inventories remain recorded; known `bdp-tx-installed-` leftovers fail the invocation, and other entries remain unclassified for review without deletion. This is the smallest correction and adds no sampling, runtime bound, workload or cleanup framework.

Read the exact v2→final diffs. Only the P1 metadata/report correction, matching command-ledger template hash, P2 three-line classification correction and correction-history/report hashes changed. All command argv arrays are byte-for-byte unchanged as parsed data. Gate/package/adjudication/source-observation/amendment files remain identical. All v2 originals retain their recorded hashes; v3 is separately preserved. No new source change is part of these corrections.

## Final reviewed identity

E: `/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-cross-carrier-draft/execution-preparation`.

| Artifact | SHA256 |
| --- | --- |
| preparation-hashes.json |a647c4ae60f2d8166b44bb8c6416e29ccab44dee953a6f7e25365f40c7406010|
| commands.json |7d67c2e75364ea5d93277c95f42ee13ac613fae21555fb2a45beccc5271420ea|
| supervise.py |8977f22c315e031f0693ce70e9f54963a17e76593422be0e8c60427bd1e378a9|
| packages.py |2ff9848c7e8e64e54aedc8f7fe05b7428006b3ebf8ca76b65d83cd8e5eabc694|
| gate.py |23fb9feeb44cd3015e366bd2b4a1401ca282ded177d28c5bab33e961ad079249|
| report.md |f0d63d7b52bc34d6da71c91ff94a9169e6cf7704f70109658860dab421f0985d|
| execution-gate.template.json |f8aea8c4e39d64ac5bc08e63763388abf68b9224adb048a734338954bcba1a77|

Independently recomputed every listed preparation file hash/size, reviewed the scripts and complete command ledger, and compared their product-facing assumptions with actual source. The earlier Path import correction and accepted N2 output variable/prefix are present. No prepared script was imported, syntax-run or executed. Data-only file/hash/diff/regex/count inspection was used. No tests, typecheck, build, install, pack, service, network, source edit or coordination write was performed; this report is the only artifact this reviewer wrote for this assignment.

## Command, selector and count review

Cumulative selection expands to39 cases:17 literal `it` cases plus parameterized3+12+4+3. App-static selects exactly the two existing route/deletion guard titles; cross-carrier static checks are folded into the latter. The nine installed selectors match the exact authored positive, two cross-carrier, three no-op and three deletion expanded titles. They are fully anchored, uniquely distinguish those nine titles, use the explicit app file and maxWorkers1, and each expects one selected pass. No full app/workspace suite, smoke command, retry loop or parallel cohort launch is scheduled.

The nine expected HTTP counts192/183/184/181/181/186/176/181/177 sum1641; nine outer192 ceilings sum1728. Every cohort expects/maxes21 children, total189. The ledger preserves the tighter source assertions rather than treating192 as a replacement count. Failed attempts grant no replacement. Separate static workload accounting is now truthful. All commands retain120s supervisor limits; existing inner request5s, cohort90s, positive63s, prepare/cleanup10s and smaller tests remain unchanged.

The new report option is in the closed environment allowlist. Both control commands use the accepted `run/cross-carrier` prefix and expect distinct batch-semantic-conflict/sequence-fresh-key JSON files. The supervisor checks output destinations remain under its evidence root and removes unrelated BDP environment variables. Final source application/hash verification and `newControlReportsResolved` remain explicit future gates; a missing report after an earlier assertion failure is not converted into a synthetic successful result.

## Provenance and cleanup review

The gate binds root acceptance bytes, the final four source hashes and command-ledger hash. Frozen commands additionally require explicit non-parent final SHA/tree, a clean checkout, ancestor relationship to eff1718, exactly the four owned changed paths and root acceptance of source checks. This does not independently prove an ordinary single-parent commit: root's stated final source-freeze review still supplies that ancestry check. There is no fallback that relabels the current parent as installed source.

Package preparation checks the build receipt against accepted source hashes, compares every actual tar member to source/emitted bytes, rejects extra/missing members, confines the documented root LICENSE mapping and server workspace dependency normalization, and verifies the same tar bytes in the consumer. No historical tar/member/payload value is substituted. Consumer configuration uses the accepted offline local protocol override and fast-uri3.1.5. The package closure resolves manifests from real paths, checks exact expected package versions/dependency edges, and reproduces the actual sorted path-sensitive payload hashing algorithm in `packages/server/src/installed-payload.ts`, including repeated dependency visits, file-extension filters and internal-symlink rejection. The installed entry is resolved under the fresh consumer and tied to retained provenance. Actual archives/install/digests still need execution and independent reconciliation; these are source checks of a proposed verifier.

The supervisor starts one new process group per invocation, records PID/PPID/PGID/RSS snapshots and progress, preserves the original command exit, and uses bounded TERM/drain/KILL only for its own group. Timeout, forced cleanup, errors, remaining group members or known leftover runtime roots cause failure. The P2 fix ensures final filesystem evidence participates. It retains sampled PID presence and unknown temp entries for independent review; sampling is explicitly not a complete process/syscall audit. Root must still review those fields and actual reports before accepting adjudication. The next cohort requires accepted prior adjudications; an exit-zero value alone is insufficient.

The ledger remains a manually stepped, root-owned procedure. It does not automatically adjudicate all report facts or execute every command. This is consistent with the bounded task and needs no larger gating framework. Final binding, exact source/static results, cleanup, packed observations and exact-head CI remain required evidence after release, not conclusions of this preparation review.
