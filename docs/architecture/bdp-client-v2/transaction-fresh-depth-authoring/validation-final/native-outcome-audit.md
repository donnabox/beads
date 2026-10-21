# Native final-validation audit — failed fresh-installed suffix

**Overall validation is not complete.** On clean candidate `1ba65f2f192cc84e44eb09ef89b9693ec6cdaf05`, tree `773614c5ebc6c97d90c4c629c7c7a02ed51bc2ef`, the source/control/package prefix and two installed cohorts passed. The fresh installed cohort failed on its seventh exchange. Actual installed traffic is **28 HTTP (13+8+7), three children**, not the planned 34. No fresh-installed acceptance/adjudication was produced.

This audit read terminal files only, checked bytes and process/store absence, and executed no product, tests, imports or adjudicator. I previously reviewed the implementation/scaffold and proposed several corrections; this is an independent check of actual data, not an additional independent code council.

## Completed source and control

All fourteen ordered source stages have exit/supervisor 0 and sourceUnchanged true. The six result files contain exactly the 42 named callbacks in source-selection.json, with no missing or extra active case: set3, submission7, sequence13, HTTP6, Snapshot12, evaluator1. All pass. There are 381 unselected/pending callbacks across those files, not executed passes. The twelve observed source identity/body measurements are retained separately; their maximum observed identity length is 288492 bytes. Source HTTP workload is the code-bound selected fixture inventory, not an independent packet capture.

Static warnings are real: format checked454 files with one oversized Read+Update JSON warning; lint checked454 files with975 warnings/81 infos, no fixes; boundaries verified11 workspace packages. Read seal verified its existing89-row artifact, not a new Transactional qualification.

The three control stages pass adjudication while preserving **direct Vitest exit1**, adapter/guard0 and supervisor0. Exactly one active Snapshot callback failed at `fresh-depth:snapshot-capture`; verbose output carries the adjacent `Caused by: RangeError: Maximum call stack size exceeded`. The selected callback had already committed/asserted its deep owned Link. No setup error is being accepted. Result hash `84ca7d19a4f429cfee603e4febafabbd0b67bab9b19d46721bf0bcccada9f877`; adjudication hash `83b29d085e7a73a433dfe138ccd07f3f5c6d0d5af7be853343af6cacb553d947`. Zero HTTP is source-derived for this selected owner-only control.

Independently rehashed all3083 positive files and3083 mirror files; all match their manifest. All180 mirror symlinks match. The only positive/copy content difference is emitted `packages/server/dist/transaction-snapshot.js`, replacing the specified owned-Link jsonValueEqual call with isDeepStrictEqual. The positive source/build was not mutated.

## Source/build/archive/consumer binding

Git reports the exact head/tree above, sole parent a17af9bc, and clean status. Independently rehashed:
- all18 frozen source files;
- all679 input files, matching both input-binding and build inventory;
- all320 unique package source/output files and448 emitted files;
- all321 regular archive members (protocol63/server258), against their workspace source and installed consumer copies;
- all870 consumer inventory files and17 symlink spellings/resolved targets.

The321 archive members versus320 unique package inputs is not an inconsistency: the inventory deduplicates source paths while two archives can carry shared source material.

Protocol archive SHA256 `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227`; server `a06f735998ab0f9872b75585b2ea32a18ce6ed9e76fbeb6550ba29eaa71b1cc6`. Eight dependency manifests/versions and actual nearest-node_modules resolutions match: protocol/server0.0.0, ajv8.17.1, ajv-formats3.0.1, fast-deep-equal3.1.3, fast-uri3.1.5, json-schema-traverse1.0.0, require-from-string2.0.2.

A separate data-only reproduction of the documented path-sensitive payload hash produced `71855458afe7fc92901ba7c9f70f0eaccb01ad484799e79b55d297e53282f800` (475 file visits/14 recursive package visits). Each child's served payload matches. The external consumer entry hash is `bcf153170f953f9e28c6c89f1153be907f64df289e64f3ad929cad4a4a54c47c`. This binds actual packaged bytes; child-declared commit alone is not source attestation.

The original verify-installed attempt remains failed: `run/verify-installed-attempt1` preserves exit1 and `Consumer lock not accepted`. Root then accepted actual consumer lock `7a3288d4e94fbb4b7f9239db3b3d256994fbcf1d4b17abd06bed4c7e892ac622` and explicitly repeated only read-only verification, which passed. No product test was repeated by that recovery. Its preservation record distinguishes the attempts.

## Actual installed outcomes

Direct13 and sequence8 each have one exact passed callback, exit0, successful strict adjudication and complete observed request fingerprints matching installed-oracles.json. All21 status/context positions match. Direct request bodies total3270 bytes; sequence3179. Direct positions are1,1,1,2,2,2,3,3,3,4,4,4,4; sequence positions1,1,1,1,2,2,2,2, with its final two responses406.

Fresh installed records exactly seven complete attempted exchanges. Requests1–6 match every pinned request fingerprint/status/context (200/pos1,200/pos2,200/pos2,200/pos2,409/pos2,200/pos3). Request7's method/URL/keylessness/48363-byte body/hash still match the intended dependent sequence, but its actual status is500 instead of200, with zero body bytes and no scope epoch/view/position. It fails before any remaining six requests. Fresh attempted body bytes total169389; observed response bytes total221220/max48948. Only the genuine-deep-owned-link-created prerequisite was completed; the final receipt/Event/Snapshot/rollback prerequisites were not established by this attempt.

Fresh exit/supervisor1, one failed callback, failureCount1, no deadline or cleanup overrun. Relevant immutable hashes:
- fresh exit: `7a56118442324d1228435226249a92dea0921e2807afa1cbae822b9051b86add`;
- fresh log: `bda0dcbbd11021a12c868747c3530809875f87b3a71d412b2a45e9e2b9dfb776`;
- fresh results: `fdda0eaa9906ddfaffc9a0550d4a1b2b6713147648fc8a648ce3076590e2dd45`.

## Cleanup and consequence

All three installed observations report startup completed, known stop, child exit0/no signal, closed and directory absent. Independently checked child PIDs94706/95681/96727 absent and all three recorded stores absent. All259 sampled PIDs across current terminal stages were absent; all supervisors report empty remaining groups, no forced cleanup and no remaining runtime roots. Nine Node compile-cache directories remain as classified tooling artifacts; they are not owner stores and have not been silently deleted. Sampling cannot prove every short-lived/detached descendant was observed.

The precise remaining serving dependency is `transaction-service.ts:643–650`: sequenceObserved still recursively freezes the projected response after executing members. Root's separately authorized diagnostic confirms depth41 succeeds, depth12000 prepares then throws RangeError at emitted freeze472–473 with position1 already committed. That diagnostic is separate two-owner/zero-HTTP evidence, not part of the28 installed exchanges; its initial setup failure remains preserved.

A narrow iterative replacement at that local projection freeze plus one full deep fresh/replay service regression is appropriate. Do not skip already-frozen parents, change projection/capture timing, roll back committed members, or relax the installed oracle. The original full thirteen-request packaged fresh cohort must pass on a fresh corresponding payload before claiming this validation complete. Neither successful prefix, prior code council, nor the expected comparator control excuses this real500. No public/profile qualification follows from these private observations.
