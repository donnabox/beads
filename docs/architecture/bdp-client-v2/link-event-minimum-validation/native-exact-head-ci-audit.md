# PR156 independent exact-head CI audit

No blocker or discrepancy found in the bounded audit. This is exact-head CI and serving-slice evidence, not whole-profile qualification or a release recommendation.

## Identity and scope

Live read-only GitHub query confirms open PR156 `codex/vickie-link-event-minimum` targets `codex/vickie-bead-event-minimum`, head `0f215c05678f507dd8a39ee08e288f0901710135`. Local checkout is clean at that head, tree `03371e5c2e1ea756e1501e0e26cccb7aa2f5d04e`, sole parent `5bcd277b2a61461211ffc0c2d2f8fd701ef0a385`. Independently hashed all 679 committed blobs against the frozen validation map: exact equality. Parent diff contains only the six reviewed files (three production transaction modules, two server test files, and the installed Event test). This is the ordinary reviewed-parent extension; no newer-main integration was introduced.

## CI and observed behavior

Run 39007135471 / job 117321782596 completed successfully at 2026-10-17T05:55:53Z. All 27 recorded steps succeeded, including pinned tool versions, formatting, lint, boundaries, typecheck, build, full tests, preflight, smoke and historical Read gate. Raw log independently rehashed: 3,610,054 bytes, SHA256 `dda59c0f057df9e36de15970797e54d1ad5ec4359e1962b20e3ec874dfce660e`.

The full test summary is 9,396 passed, one skipped, 186 passed files. The separate readiness preflight passed seven tests. Every one of the 33 selected source callback names and four selected installed callback names appears exactly once as passed in the full-suite verbose output.

All 28 source observation records were located. Counter/preflight/waiter/owner records match the local accepted observations. Bead D reports 37 ordered responses (200×8, 503×10, 500×12, 400×4, 404×3), cleanup-inclusive 915 ms. Link D reports 35 (200×8, 503×8, 500×12, 400×4, 404×3), 749 ms. Both are below 4,000 ms, report zero children/failures and directory absence. Compared full ordered method/status/suffix arrays to local evidence, normalizing only the freshly generated `after` cursor; no historical relabeling.

The three parseable minimum reports independently confirm the actual workspace-backed installed observations:
- finite: 3 exchanges; statuses [200, 503, 503]; reported child 19173 closed with exit 0, null signal, no deadline/failures, directory absent.
- bead: 4 exchanges; statuses [200, 200, 503, 503]; reported child 19272 closed with exit 0, null signal, no deadline/failures, directory absent.
- link: 4 exchanges; statuses [200, 200, 503, 503]; reported child 19395 closed with exit 0, null signal, no deadline/failures, directory absent.

For each minimum report, the final GET is 503 with the expected 127-byte problem digest, and HEAD is 503 with zero bytes and matching content-length 127/problem media type/current context/retry-after 1/private-no-store/non-attesting headers. Bead/Link setup reports actual pos-1; Scope pos-0. Link's four ordered routes are real batch seed, setup `links/known?view=events`, opaque GET and HEAD; Bead analogously uses `beads/known`.

The historical missing-Event callback also passes. Its retained raw line contains 11 response status fields (200,200, eight 404s,401) and visible closed/exit-0/null-signal/directory-absent/failure-0 fields. GitHub masking damages the WWW-Authenticate string, so that line cannot be parsed as intact JSON. I did not reconstruct/unmask it, infer unretained response bodies, or treat it as independently parseable wire proof.

## Provenance and limits

The full-suite minimum callbacks declare clean commit 0f215 and the workspace `packages/server/dist/transaction-installed-process.js`, entry SHA256 `bcf153170f953f9e28c6c89f1153be907f64df289e64f3ad929cad4a4a54c47c`. Their workspace payload is `077a3998a9b86fede376933eb9cce37cc1d80d2c008b49776cb3c5bdc274a852`.

Packed smoke separately reports payload `81e1ee6779a2e0db7fece6427ab0fa10538c14a0c24902789b6572ad33550344`, matching the previously audited local packed-consumer payload, allNineTargets true and claimEligible false. These are different execution/provenance identities; full-suite observations are not relabeled as packed-consumer observations.

The Read gate verifies its historical artifact `docs/design/evidence/read-cohort/read-v1.json`: constant `674982be2123b9b7aae630bf79c3f774122900cd`, runHead `e5f7d0c43c151426350ceededdc77fe7e372db9b`, evidenceCommit `dbaf63a24219e1ea3f80394746853598b0d4464a`, projection `bd7da0fc3f8082e0bfb6f4969c04409f2a610e12408dbc6a0bd14b4b351127cc`, 47 definitions/46 reachable/89 rows. Its identity remains historical.

Remote lifecycle statements above are recorded child-exit/store observations, not an independent remote process census. No remote PID-absence claim is made. Counts do not establish public discovery, S1/G2, foreign-context semantics, complete profiles, main integration or release. No tests, builds, product helpers, source edits or runtime retries were performed for this audit.
