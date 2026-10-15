# Native installed/publication audit — PR127

No actionable discrepancy found in the candidate identity, fresh installed package, retained eleven-response observation, or final cleanup. This is a read-only artifact audit; it does not approve merging, assert CI success, or qualify the Transactional profile. The reviewer previously prepared the emitted-control copy and audited its outcome; this report independently recomputes the fresh package and installed evidence without executing product code.

## Exact source and publication

The clean checkout is `3df82ac480f50c8f486b4e456336366e7acbdf87`, tree `d5f02c936cf080666a8e73bb5f253e52086705f4`, with exactly one parent `88c9408e908590ab1a4c64c9b9dee1a8b705999c`. Its parent diff has exactly the six accepted paths. Each actual file equals the preserved redirect-repair snapshot and its SHA-256:

| Path | SHA-256 |
|---|---|
| packages/server/src/transaction-store.ts | b2a444f7997ba4e2b34bd57ed2a8119f0999e6146305b5aae4fc8735ff78bc7c |
| packages/server/src/transaction-http.ts | 9db33008a4c3a66214e2bf523b0b90f79c216c1f3f304c6df0491a0d1904439a |
| packages/server/src/transaction-events.test.ts | 0152561cdf096a48b0b01ec0e0d8bc81208a905bb5bf71dad411694387b58b27 |
| packages/server/src/transaction-event-sse.test.ts | 69f58153b93947f028c37aebce2f57b0a25f725aacb9a1891c7b276b1c9e44c0 |
| apps/bdptest/src/transactional-event-missing-context-installed.test.ts | 828c05f358ae5863c992f93194da15a31df9a200e073e5e3525531ba39fe7d0d |
| docs/design/transaction-implementation-progress.md | c31d464b9d7928639c54acf827844d3ae62691d026ee9557ed6f7b921643b278 |

The historical binding `acf190e742dfb669fe5607085a3ac60c5ecb66f4f86b1f05025424a7873615b0` correctly remains a dirty draft atop 88c; final-source.json separately attributes the unchanged bytes to the clean commit. Root council and control acceptances remain separate records.

One remote snapshot, saved as `native-publication-ci-snapshot.json` (SHA-256 `f86727ba1138208a11d4e0a775b07d15ccdb4e493187156acf8b7551efe8c9b2`), confirms draft PR127 head 3df and base 88c. Run 38728209051 and job 116399188531 were in progress with null conclusions. Synthetic commit `33580067aa88dbc1f0b5f15e2c2b9af739d9a200` has ordered parents 88c and 3df, and the exact candidate tree d5f. No further CI polling or terminal-success inference was made.

## Fresh packages and installed identity

Independently reread and hashed all 665 build input entries and 312 retained build/output entries against the checkout, plus the pinned successful build start/exit records. All matched. Independently opened both archives without extraction or imports:

- Protocol: 63 members, archive SHA-256 `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227`.
- Server: 250 members, archive SHA-256 `7ec576ab3e2b1f175868e3d85260fb442eb0bc5712b1646a012a8e3a42a509a0`.

All 313 members are unique regular package-relative files. Exact member sets match source dist/schema/license/package inventories and installed package inventories. Every ordinary member matches its source and installed bytes. Package JSON comparison allows only JSON serialization and the declared server dependency publication normalization `@bdp/protocol: workspace:* → 0.0.0`.

The actual consumer contains exactly the eight pinned production packages: protocol/server 0.0.0, ajv 8.17.1, ajv-formats 3.0.1, fast-deep-equal 3.1.3, fast-uri 3.1.5, json-schema-traverse 1.0.0, require-from-string 2.0.2. Independently resolved every declared dependency inside the consumer and checked manifest hashes/versions. Independently recomputed the path-sensitive recursive payload hash over 14 dependency visits and 473 file visits:

`0caa48ec6e4a74ad3e3c322150cd29d3ec172048b82e5cf62ae55bd917141b48`.

This equals both provenance and the actual observed served binding. The installed entry resolves inside that consumer and has SHA-256 `bcf153170f953f9e28c6c89f1153be907f64df289e64f3ad929cad4a4a54c47c`; that entry digest is not presented as the whole payload digest. Source attribution is established by external source/build/archive checks, not child attestation.

## Actual installed outcome

All ten preparation/execution/review stage records have command and supervisor exits zero, no timeout, no forced cleanup, no errors, and unchanged source. The selected Vitest result is exactly one passing case, zero failed/skipped cases:

`observes genuine installed missing Event context across Bead and Link JSON and SSE GET and HEAD`.

The genuine console JSON equals installed-observation.json byte-content after JSON parsing; no response was fabricated. That observation has SHA-256 `3a79fd2ddd9a333927e17ecce7e61d5a7035a51c8d67b42684e4574064a52cf6`; the actual result JSON has SHA-256 `5156a58463b3573a55800c329ee7c8e4471b1ea653834adce48e4e80c519bef1`.

Independently checked all eleven method/URL/Accept/status rows: batch POST200, finite Scope Events GET200, eight Bead/Link missing Event GET/HEAD requests with JSON/SSE Accept returning404, and anonymous Bead SSE GET401. The eight404s carry the same actual epoch/view/pos-1 tuple as the finite comparator; the401 carries none of those context fields. GET Problems have134 bytes; paired HEAD responses have zero body bytes and representation content-length134. The retained metadata includes problem media type, private/no-store and the non-attesting marker, without retry/challenge/development-limit on404.

The source/config/entry/payload bindings are consistent with the private transactional fixture, canonical authority, and 336 artifact bytes. Runtime configuration identity is `31156f0ecb8147fc7e3cd124eb29d0c8771a446e365ff0c3e17a8a934949d8bb`. Full schemas/bodies and serializer assertions ran inside the passing test; the retained report contains selected headers and fingerprints, not recoverable full bodies or an independently observed transport-method field.

The actual child PID89610 exited zero without a signal and was closed. A fresh read-only process listing found all37 sampled PIDs and all10 recorded process groups absent. The owned runtime/store directory is absent. Current attempt temp directories contain only four retained node-compile-cache directories; other attempt temp directories are empty. Process sampling does not prove no short-lived unsampled process ever existed.

## Preserved limitations and failures

Original source-selection supervisor failure (four passed/four intended cases missed), successful targeted follow-on cases, and original emitted-control dependency-postcheck failure remain unchanged. The separately accepted control disposition documents regenerated command shims and equivalent-target link spelling changes; this audit does not rewrite that run as a byte-identical dependency graph or rerun it. The first empty Claude timeout remains inconclusive; the completed focused review and accepted ten source dispositions are separate.

This audit added no tests, builds, installations, imports, services, source edits, publication changes, or CI retries. Local installed evidence is complete for this eleven-request scenario. CI was pending at the single snapshot; public discovery/profile eligibility and merge remain unclaimed.
