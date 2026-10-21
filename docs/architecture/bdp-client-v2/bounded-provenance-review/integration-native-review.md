# Native Codex integration review — PR83

**Disposition: no actionable findings or merge blockers found in the bounded scope below.** This is a development implementation review, not public Read+Update/Transactional profile qualification.

## Exact scope

Reviewed frozen commit `8c78d3e198561c242286a185cd8cb0424f9b5b83` from `/Users/dbox/repos/wt/vickie/bdp-provenance-runtime-integration`, against PR83 baseline `cd22ac3`, with already-reviewed common provenance source/evidence `f1e83dc` and reconciled four-part editorial source `5a67b75` as the other inputs. Examined integration merge resolution and the PR83-specific migration surfaces rather than repeating the entire earlier runtime review.

Also independently reviewed the previously uncovered post-`b7eefc9`, pre-`cd22ac3` supplements:

- `088f04d963283ddc13390f69eb17cb38aee050e8`: private transaction submission/admission composition, operation-limit placement, comparison controls and failure tests; seven files, 828 additions/40 deletions.
- `9e11bbbe315bd4896abd02c5859bf3cc7dff658c`: expired-creator diagnostic README, catalog, coverage and manifest only; four files, 735 additions/7 deletions. This is an observation-plan addition, not a diagnostics implementation change.

Author checkout remained untouched. Tests ran in an external `git archive` of the exact frozen head at `native-source/`, with local workspace package links and independently generated build outputs. The author subsequently advanced to `55ed77e3035647f7ab32030978eba6b01a0dc993`; comparison showed only `STATUS.md` changed. That later status commit is outside this review.

## Findings

None. No inherited protocol concern was promoted into a new landing requirement. In particular, unsupported private set/owned-Link transaction execution and future socket/receipt HTTP work remain explicit bounded implementation limits; this review does not require completing a public write profile.

## Source checks

1. **Merge preservation.** `docs/specs/bdp.md` is byte-identical to `5a67b75`. Shared production protocol implementations, schema/adapters, `resource-evaluator.ts`, `member-identity.ts` and `recovery-store.ts` retain the reviewed PR65 source. The protocol test differences from PR65 move problem-table assertions to the owning Read, Read+Update and Transactional sections while preserving the closed inherited code set and fixture checks.
2. **Transaction provenance.** `packages/server/src/transaction-batch.ts:307–309` normalizes empty `revisionProvenance` to omission while retaining explicit actor/context states in identity. Lines 557 and 567 copy the complete created/updated version envelope into Events. No-op outcomes still bypass Events, and no deletion or alias provenance is introduced. Focused tests cover receipt/postimage/Event agreement, Link updates across both mutable planes, restart, cross-carrier identity and empty/omitted normalization.
3. **Store gates.** `packages/server/src/transaction-store.ts:112–115` advances both format and normalization-policy identity. Reopen checks at lines 385–391 precede pending receipt deletion, creator-fact deletion, expiry sweep and timer startup. New byte-preservation tests cover stale format/policy with retained, expired and pending state. `packages/server/src/development-authority.ts:36` advances the installation/lineage identity; the RU format-3 gate is preserved. The development authority opens/validates the store before its explicit expiry pass; compatibility tests reject old installations without rewriting or reseeding.
4. **Private admission supplement.** `transaction-submission.ts` enforces body bounds, carrier/key syntax and trusted principal decision before consulting key state. `transaction-store.ts` compares existing identity before fresh operation limits and response negotiation; exact duplicate negotiation does not replace the receipt. Unknown-key operation/negotiation refusal does not bind a key. `transaction-comparison.ts` snapshots receiving controls and preserves one deadline across unresolved owner replacement; epoch changes terminate comparison. The helper distinguishes unsupported private capabilities from protocol malformed input and propagates fenced/internal failures rather than fabricating permanent receipts. Raw HTTP receiving obligations remain explicitly deferred.
5. **Diagnostic supplement.** The new expired-creator plan captures the original allocated Bead identity, expires/reopens, binds a renamed label through the retained creator identity, reads the original identity, distinguishes a retained dependent postimage from newer live state, and repeats after reopen. Coverage references resolve to actual assertions; the bounded-observation limitations and non-attesting catalog guard remain intact. The actual development matrix passed in the private snapshot.
6. **Four-part guards.** `profile-catalog-boundaries.test.ts` retains cumulative profile ownership and historical anchor aliases, separately checks part ordering so Data Model cannot acquire a Read dependency, checks duplicate/ambiguous ownership, and audits diagnostic citations without admitting their IDs as normative production artifacts. Relocated problem-table tests passed.
7. **Claims/evidence.** No public write-profile advertisement or executable admission was added by these integration changes. Historical Read evidence is not a fresh observation of this integrated head. Fresh exact-source cohort sealing remains the parent workstream's separate gate; this report does not claim to verify that future seal.

## Independent validation

Runtime: Node `v24.16.0`. Private build: `node node_modules/typescript/bin/tsc -b`, exit 0 (`native-build.log`).

- `native-focused-built-tests.log`: **198 tests passed, 7 files** — transaction submission/store/dependencies/sequence, development-server, profile-catalog-boundaries, development-read-update inventory.
- `native-diagnostic-tests.log`: **313 tests passed, 4 files** — actual development-matrix HTTP/lifecycle diagnostics, transactional-wire, read-update-wire, schema-read-projection.
- Total final focused validation: **511 tests passed in 11 files**. No full-suite, fresh Read cohort, or public profile qualification is claimed.

The initial unbuilt archive run (`native-focused-tests.log`) passed 185 tests and failed 13 crash subprocess tests solely because `packages/server/dist/transaction-store.js` did not exist. After the private build, the same seven-file suite passed all 198 tests. The failed preliminary log is retained for evidence integrity.

No source edits, pushes, PR/bus messages or profile claims were made by this reviewer.
