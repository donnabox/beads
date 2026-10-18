# Frozen owned-Link Transactional projection

Clean frozen HEAD **01e2d5470941a49240fab7820a6dea1b6deb291e** at `/Users/dbox/repos/wt/vickie/vickie-transaction-owned`, branch `codex/vickie-transaction-owned`, origin `https://github.com/gastownhall/bdp.git`. No outward writes. Base is exact root71549f2610d61e6562b2c703d96e4c25d1093492.

History:293fab8 adds owned runtime/projection and tests; ordinary merged guard27bad2badc765fc02aeec9456e1fc5f9dd49ea5c as d3008c7, preserving both parents;01e2d54 adds final count/recovery evidence and progress notes. The only merge conflict was additive progress-document text; both intents remain. Batch code merged automatically, retaining receipt guards and cardinality-before-cap. No hooks were bypassed.

## Runtime scope

Owned author changes are transaction-batch.ts's applyResource/fact/change projection, transaction-store.ts's owning-descriptor refusal, new transaction-owned.test.ts, the narrow installed-contract negative/positive fixture split, and the progress document. The incoming guard merge additionally carries its existing transaction-set.test.ts/Problem helper changes. No shared evaluator/compiler/Read/HTTP/service/format/schema/profile or count/materialization budget change.

The evaluator still decides ownership, validation, source-write policy, CAS, versions, no-ops and provenance. A synchronous local wrapper captures the existing same-turn Resource row immediately before forwarding each evaluator putResource. This obtains the owning source's actual previous revision without exposing sourceRecipe, rereading later owner state, minting revisions or duplicating writes. Result source/sourceRevision survives unchanged, including no-ops.

All changed IDs are looked up as local staged paths, parsed once, and entered into the changes Map by canonical absolute Resource ID. Deleted subject facts use the direct pre-evaluation record and verify its canonical id/type/revision against the evaluator outcome. Repeated operations replace only the final Map entry; intermediate receipt results and Event chains remain. An owned mutation changes at most two Resources: Link and source Bead. Ordinary target remains unchanged; self-Link source changes once.

Existing lifecycle/incidence facts stay ordered. Owning source updated is appended last, so local creation/deletion yields four Events and update yields two. Created nested Link is the complete postimage; updated nested delta is field-built id/type/revision pair, actual admitted patch(es) and new provenance; deleted nested value is identity only. Source Event carries source revision/provenance and excludes direct patches. The existing scopeEvent cap accounts for this extra fact. No-op exits before changes/facts while preserving sourceRevision.

Owned installed descriptors now use the same compiled or schema-free registry already driving the owner. The negative installation test still proves no provisioning on its remaining refusals; a separate complete owning descriptor bundle opens/closes its own file. Format8 is unchanged: no durable shape/identity change, and old accepted installation configurations could not contain owning descriptors.

## Executable evidence

Pinned Node24.16.0/pnpm11.20.0. Frozen install and build PASS before process tests. Final **753 tests /18 files PASS in9.52s**, maxWorkers2: all transaction test files plus Resource evaluator, AuthorityRead, Read-sequence lifecycle and shared compiler. Full log `transaction-owned-regressions.log`, machine JSON `transaction-owned-regressions.json`. New owned file has15 tests including:

- Create/property/metadata/combined update/delete; exact inline/first-class agreement, source/target version behavior, provenance removal, normative receipt/group schemas and semantic Event payload/order assertions.
- Self-Link source/target incidence ordering, explicit versus wildcard empty buckets, ownership maximum failure, two Links sharing a source, interleaved direct source update and reversal chains/final normalization.
- No-op/mixed transitions, local creation labels through Link deletion and final source tombstone, addressed-Link and source-Bead stale guards, incident deletion safety, exact atomic rollback.
- Staged canonical set updates/deletes of owned Links, per-member sourceRevision, durable receipt paging, original replay after later graph mutation and reopen, same singleton/one-op batch namespace.
- Existing closure predicate at batch/set seam, including transitive cycle with hidden target; selected readable-but-source-nonwritable failures are not filtered out. This is policy-interface evidence, not installed variable-grant qualification.
- Inherited compiled Link schema rejects a later mutation over actual emitted HTTP; owning Bead/Link reads agree, failure rolls both back, changed graph plus reopen retains exact failed receipt and same-owner current reads.
- Real child SIGKILL at after-effects and after-terminal leaves neither partial Link/source nor group/identity effects after reopen; re-admission can complete the original caller-assigned Link identity. No power-loss claim.
- Lowered Event cap3 fails on fourth source fact atomically. Actual default count test:1,024 small self-Links on separate source Beads delete to4,096 Events/2,048 final Resource changes;1,025 selection fails with transaction-wide inducedEvents limit and retained failure replays after reopen. Isolated test2.14s. No byte or arbitrary high-fanout qualification.

Typecheck, dependency boundaries, format, lint, git diff check, both installed package smokes and unchanged89-row Read evidence verifier PASS. Lint reports32 inherited warnings and1 info; no new owned warnings or errors. Format reports the inherited oversized1.6MiB matrix warning. Relevant logs `transaction-owned-{build,typecheck,boundaries,format-check,lint,smoke,smoke-development,evidence}.log`. Public profile fail-closed smoke remains intact. No evidence artifact regenerated or relabeled; no full workspace suite/new public cohort claim.

## Debugging and cleanup disclosure

Initial paging test used a key instead of receipt URL; fixed the oracle. Initial HTTP fixture omitted required administrative artifact limits and had an untyped response access; fixed fixture/types. One new lint error was an unsafe optional-chain cast in a test; fixed before final verification. Production behavior was not weakened to cure these tests.

An initial full-cap experiment with32 self-Links per owner remained CPU-active beyond80seconds and was terminated; a short accidental repeat during a failed editing command was also stopped. Both exact process trees and their named disposable stores were cleaned. That run is INCONCLUSIVE, not a timeout pass or correctness failure. Its synchronous work prevented the test runner timer from firing. The final bounded fixture uses one Link per source and avoids repeatedly schema-validating every previously checked group in its helper. Because both graph shape and test-helper work changed, the slow attempt does not isolate runtime versus test-harness cost and must not be cited as a diagnosed owner performance defect. Partial first log is `transaction-owned-boundary.log`; final successful default test is `transaction-owned-default-events.log`.

All final runs completed, no test-owned child remained, empty failed HTTP fixture directory removed, source clean. No shared/pinned database touched.

## Remaining scope

Root's receipt/group byte-materialization design and mutated-Resource count accounting are separate; no cap or byte/count reinterpretation here. The actual owner remains fixed all-visible with empty aggregate policy. Mutable grants/retained redaction, HTTP set admission, Snapshot/changefeed serving, schema hosting and incomplete minimum controls remain unqualified. The4,096 count test proves the small per-source workload only. Two-Resource/four-Event assumptions describe the currently normed owned transition, not a general future fanout mechanism. Root owns independent review, integration and publication.
