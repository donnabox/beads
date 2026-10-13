# S2a external review closure

Frozen head: `3964b18dafcbbc3598c5ab21c550b73ccb0746cf`.
Tree: `89a8f91c41b9bd3b5fd771a79c9382393a384300`.
Ordinary parent: `eb4939b6c3e9496a450724953142e47c7933f69e` (integrated c934 oracle and main).
Checkout: `/Users/dbox/repos/wt/vickie/vickie-transaction-set-review-repair`.
Branch: `codex/vickie-transaction-set-review-repair`; origin GitHub gastownhall/bdp.

Source is clean and frozen. No outward writes. Original 2f7/c934/main integration checkouts remain unchanged.

## Five findings and dispositions

1. **Depth before determinate syntax: fixed.** Moving only the outer check would leave the same bug in synthetic Resource batch and set/RU serialization. The existing private lossless serializer is now iterative and does not apply execution depth. It preserves numeric lexemes and decoded object key ordering. The private depth check is a separate traversal with the same >32 convention, applied after complete known carrier shapes, member discriminators, reference/label syntax and Property Change syntax. The singleton input is still at equivalent batch depth 2. Actual input remains limited to 64KiB before any decoding; no new public or protocol limit is introduced. Unsupported Selector validation remains a later private gap, not a newly inferred malformed rule. No graph candidate scan occurs during syntax preparation.

   Actual HTTP counterexamples cover invalid root, invalid Resource member field, forbidden member idempotencyKey, malformed set patch and a malformed later member following a deep set. Before the repair, all five authorized requests returned 500 instead of expected 400. The earlier unauthenticated reproduction returned 401 instead of 400. After repair both authorization cases return malformed-request 400 with no development marker, no key rows and no effects; the refused key remains usable. The reproduction script temporarily builds the exact inherited two production files in the isolated successor, records intentional failures, then restores the repaired source byte-for-byte in finally and rebuilds. It never modifies a frozen checkout. See `ordering-repro.json`, `authorized-depth-before-red.log`, `depth-before-red.log`, `reproduce-ordering.py`.

2. **Remaining non-set pointer gap: disclosed, not expanded.** The progress note explicitly says ordinary Resource-operation receipts still omit problem.pointer for some locatable missing-path and inadmissible-number failures. That shared RU/evaluator mapping requires separate scope. The conditional pointer row is not claimed fully qualified.

3. **Duplicate singleton set validation: removed.** The singleton first rejects a raw operation field so the body cannot replace the route discriminator, then the mixed path performs the one authoritative set shape/RU syntax validation. That pass also rejects raw name, idempotencyKey and unknown fields. There is one serializer implementation, no parallel parser or changed negative Selector policy. No quantitative speedup is claimed. Tests preserve lossless deep numeric syntax, raw closed fields and route protection, actual 64KiB vs 64KiB+1 handling, and singleton/batch depth parity at the exact existing boundary.

4. **Sets forbidden in sequence: regression added.** Real HTTP updateWhere and deleteWhere sequence requests include an earlier valid creation. Both return malformed-request 400; neither member key/receipt nor the earlier Resource exists afterward, and the earlier key can then be used by a valid request. No sequence behavior was changed.

5. **Historical set-refusal/ordering claims: explicitly superseded.** The S2a closure names the earlier direct admission, first HTTP and private set core statements as historical. Valid supported sets are served privately; unsupported Selector policy, general scan-work bounds and complete qualification remain open. Historical tests, budgets and evidence are retained.

## Files and boundaries

Six files changed (313 additions, 19 deletions): private `transaction-batch.ts` and `transaction-set.ts`; `transaction-http.test.ts`, `transaction-submission.test.ts`, `transaction-set.test.ts`; and `docs/design/transaction-implementation-progress.md`. No owner/store/evaluator/Selector, protocol/schema/catalog, public exports, budgets, installed fixture or conformance-authority changes. The c934 installed owned oracle is inherited unchanged.

The deep-input tests use received input under 64KiB with 20,000 array levels: syntax serialization preserves `1e+30`, `9007199254740993`, `-0`, keys and escaped strings without a recursive serializer stack failure. Complete known malformed syntax still yields 400; a valid but too-deep request remains a private gap after key/auth checks, with owner state unchanged. This is finite input traversal, not a new scan-work guarantee or permission to execute deep payloads.

## Verification

Pinned Node 24.16.0 / pnpm 11.20.0, frozen lockfile install, tests maxWorkers=2; logs outside source.

- Before-fix discriminator: **5 intentional failures**, authorized requests observed 500 where 400 was expected; source restored byte-identically and rebuild exit 0.
- Focused final suites: **136 tests / 3 files passed** (includes actual emitted HTTP receiver).
- Broader Transactional suites: **625 tests / 23 files passed**, 13.41s.
- Clean build, typecheck, dependency boundaries, format check, lint, installed package smoke, development smoke and unchanged Read evidence verification all exit 0 (`check-exits.json`). Lint reports 323 warnings/15 infos and no errors; no warning-free claim. Full repository tests were not repeated in this bounded worker slice.
- Fresh frozen pack: verifies exact HEAD and clean source, clean rebuild, packs protocol/server, installs with scripts disabled into disposable directory, then runs installed set test **1/1 passed** plus existing installed Transactional regressions **35 tests / 5 files passed**. The latter includes actual-default 15-second SSE observations; no timeout was widened. All pack/build/install/test exits 0.

Fresh pack logs: `packed-3964b18d/`; script `pack-executor.py`. Full actual set wire record is `observation.json`; compact source/config binding is `observation-summary.json`. Records say source 3964b18, workspaceDirty false and claimEligible false.

- Protocol tarball SHA256: `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`.
- Server tarball SHA256: `80f0ab8fa061480be11428b72b110b0f2f8185d467f561f6d30912f2e268562c`.
- Installed payload digest: `1b12629fe54c39c8f7864dbb13bdba16bcbe4298b0a5bd420fa19daaa4fb7ec2`.
- Captured contract digest: `95b4ef854865acafc04caa438285dd52b7955fa99e6b241397660d6f262b4935`.
- Actual set cohort runtime configuration: `c7f5bc9e714168e8cad2f0dffa8e87e898f07b19def656e80aab574eec544dd1`.
- Set cohort: 32 wire observations, 33 result entries across pages [32,1], 96 Events across pages [32,32,32], one group, five negative oracles; actual owned literal/revision/effect-position checks and replay/reopen remain passing.

## Cleanup and limits

No timed-out or interrupted test run; no owned child processes remain after final process inspection. Source remains clean. No push/PR/bus/main/other-lane edits. Parent owns integration and publication. These results close the bounded review findings; they do not settle Selector Q1/Q2, provide a general candidate-scan bound, repair ordinary Resource pointers or qualify/advertise the full Transactional profile. Sealed historical evidence is unchanged and not relabeled.
