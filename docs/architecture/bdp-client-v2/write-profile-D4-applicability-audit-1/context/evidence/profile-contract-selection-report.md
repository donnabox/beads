# Cumulative profile contract groundwork — frozen report

- Checkout: `/Users/dbox/repos/wt/vickie/vickie-profile-contract-selection`
- Branch: `codex/vickie-profile-contract-selection`
- Base: `caf5ca75a4e9b9c7865e61e97f849ab736c21e64`
- Frozen clean commit: **`abdd1d4af45cd02ad1238b311462a1b154e44b79`**
- No push, PR, bus, runtime, public admission, public artifact/schema or sealed evidence changes.

## Delivered

The source-only `packages/conformance/test-support/profile-contract-plan.ts` reuses current catalog parsing, profile/retirement selection, canonical JSON and executable manifest parsing. It derives required obligations from all current core and companion catalog files, deduplicates exact companion Read copies, rejects conflicting copies/duplicate retirement owners/invalid edges, isolates optional History and excludes diagnostics. Factory-loaded immutable selections prevent replacing a complete selection with a caller's delta/subset. Exact source bytes are hashed; the loader remains a trusted repository boundary, not an attestation of arbitrary replacement catalog text.

The execution planning API consumes actual manifest bytes and a closed served-configuration binding. It preserves every missing inherited or write obligation and separately reports unplanned, unavailable setup, unresolved applicability and ready-but-unexecuted rows. All remain unobserved and claimEligible:false. It never turns missing fixture capability into an automatic N/A or pass.

The binding names Scope, selected profile/History, exact source commit, installed payload, installed contracts/schema closure, policy, trusted authentication configuration and exact fixture digest/ID/capabilities. Mismatches in any field refuse combination. It validates shape/equality, not truth; launchers must capture those non-secret identities from the actual installed target. There is no second action language or persisted evidence format.

## Catalog correction and inheritance audit

Confirmed Claude F2: `read-update.discovery.document` now retires `read.discovery.document`, because the latter fixes the closed Read-only discovery shape. Existing TX selection correctly carries this retirement forward even while retiring the RU document. No spec change required.

`read.discovery.optional-limits` remains inherited: its actual citation is the unchanged law about omitted bounds and refusal/truncation. The old Read executable row hardcodes `readDiscovery`; later cumulative executable realizations must use profile-specific schemas rather than copying that assertion.

Audited all 81 RU rows for F12 in `docs/design/profile-contract-planning.md`. The spec expressly preserves sequence200/shared member envelope and semantics at TX, while singleton/batch use receipt200/202. The existing twelve TX retirements remain sufficient for these catalog obligations; each row's treatment is documented. Removed one misleading descriptive citation (“Read+Update has no durable Mutation Receipts and no BDP Events”) from the inherited owned-Link no-op row; retained all governing no-op/revision/source/provenance citations. TX's own no-op Event obligation remains separate.

Correct derived normative inventories: **44 Read, 129 RU, 250 TX** at this commit. The Read catalog contains49 rows but five are diagnostics. Companion catalogs add five distinct Update obligations. There is one RU→Read retirement and twelve TX→RU retirements. These are inventory snapshots, never coverage scores or hardcoded selection authority. Earlier native report arithmetic was corrected with an explicit frozen-source-versus-later-repair note.

## Validation

Pinned Node24.16.0 / pnpm11.20.0; frozen install passed.

- PASS build (`pnpm build`).
- PASS typecheck (`pnpm typecheck`).
- PASS dependency boundaries (11 workspace packages).
- PASS focused serial Vitest: **375 tests / 11 files**, 7.72s. Files: profile-contract-plan, catalog, read-update-catalog, transactional-catalog, artifact-bundle, numeric-model-catalog, owned-wildcard-catalog, history-catalog, profile-catalog-boundaries, development-read-update, runner.
- PASS Biome check on all changed TypeScript/JSON and `git diff --check`.
- PASS existing evidence verification: unchanged constant `674982be2123b9b7aae630bf79c3f774122900cd`,89 rows, existing source/evidence pair e5f7d0c/dbaf63a. No regeneration.
- No full workspace suite or served HTTP qualification executed; this slice changes only metadata, source-private planning, tests and documentation.

Initial focused run caught a test selecting a nonduplicated companion row instead of the copied Read row; corrected the test target, then the complete focused suite above passed. Biome initially requested import sorting; safe import-only fixes applied and check passed. No test-owned services or pending sessions remain.

Logs: `profile-contract-selection-{focused,typecheck,build,boundaries,evidence,format}.log` alongside this report.

## Next integration

U1's actual installed mutable fixture should provide the binding and a cumulative executable manifest using its profile-specific wire forms. Feed those bytes to `planProfileExecution`; execute ready IDs against the actual same-owner target and retain all unplanned/applicability/setup gaps. U2 supplies real installed schema/ownership/policy/auth fixture variants with separate bindings. T2/T3 supply the same inputs from the Transactional owner. A later reviewed runner/evidence/admission change must connect actual observations; this planner grants none.

No unavoidable product decision identified. Root should review the small catalog retirement/citation repair and the explicit config capture contract before merging. The existing Read/diagnostic artifact guards and source graphs remain intact.
