# Direct Transactional refusal context: bounded serving repair plan

## Recommendation and status

**The missing response context on an authenticated, syntactically valid direct idempotency conflict is a settled serving defect.** Prioritize repairing it over authoring the proposed cross-carrier evidence row. The smallest coherent repair captures context at the existing observed admission decision and carries it to HTTP. The same seam can cover operation-limit413 and bodyless406 without another observation or a new precedence rule; recommend including those three final dispositions in one bounded repair.

This is a read-only supporting plan for root/council acceptance, not authorization to edit. The cross-carrier proposal remains unaccepted. The earlier native review’s suggested evidence-only narrowing does **not** discharge the normative obligation; this follow-up supersedes that recommendation as the proposed next work. No new normative text or interpretation is needed for the authenticated cases below.

Basis: clean `/Users/dbox/repos/wt/vickie/vickie-deletion-safety-schema-repair`, commit `39c160d0d35294003f400cef8d5ab5c6ec9aec9c`, tree `75e8fb97979e304bbe6fa24b7a570d9fa01efbbb`. All source references below are at that pin. PR115 is not assumed merged. Root must supply the actual ordinary parent and coordinate shared HTTP files with the active SSE repair before authoring.

## Exact law

- `docs/specs/bdp.md:5766–5774`, Transactional HTTP consistency: “Transactional Scope-bounded responses carry” BDP-Scope-Epoch, BDP-Authorization-View and BDP-Scope-Position. The sentence is not restricted to successful responses or receipt envelopes.
- `bdp.md:4490–4530`, admission order: syntax and principal checks precede key comparison. A different normalized request under the bound key produces direct409/idempotency-conflict and leaves the earlier outcome unchanged. Conflict precedes response negotiation; equal retained requests negotiate before fresh limits. An unknown over-limit request produces413/limit-exceeded with limit transaction.operations, and unacceptable response negotiation yields bodyless406 before admission. Refusals create no receipt and bind no key.
- `bdp.md:3930–3936` independently requires refusal before admission for a differing normalized request while the key remains bound.
- `bdp.md:5233–5244` explicitly distinguishes receipt execution facts from the serving observation. That receipt-specific rule is corroborating context, not the sole legal basis for direct Problem headers. A conflict after unrelated commits must not synthesize its current position from the retained receipt’s old effectPosition.

The tested cases have an authorized principal, a known configured Scope and an ordinary valid carrier with no minimum fields. They do not require resolving mutation minimum-field applicability, invalid minimum tuples, nondisclosure/error precedence, or a view-rotation policy.

## Actual source and smallest seam

1. `transaction-service.ts:499–517` invokes submitTransaction with observed=true and passes non-receipt results through. There is no need for a new read service or transport callback.
2. `transaction-submission.ts:101–118` selects owner.admitObserved for that path; compareAndAdmitTransaction returns the final admission result unchanged (`transaction-comparison.ts:47–54`).
3. `transaction-store.ts:1384–1396` already enters the owner and executes admitKey inside atomic(). `admitKey` decides both ordinary and dependency-template conflicts (`:1065–1079`), fresh operation-limit and fresh not-acceptable (`:1081–1086`). The outer atomic callback converts an existing receipt to not-acceptable at1392. Only existing receipt success currently receives observeReceipt/readContext.
4. `readContext` (`transaction-store.ts:583–594`) reads actual metadata position and returns a frozen current epoch/view/position tuple. Capture it **inside that same atomic callback**, after the final refusal kind is selected and before COMMIT, for observed conflict, operation-limit and not-acceptable. Handle the existing-receipt406 branch too; adding a wrapper only around the final return would miss its early return. Do not add context to ordinary unobserved store APIs or sequence-turn results as incidental behavior.
5. `transaction-submission.ts:120–130` currently drops any prospective conflict/limit context while translating their discriminants. Preserve the captured tuple in those two translations; not-acceptable already returns its object. Require context on these final dispositions when observed=true, so a missing propagation cannot silently become a headerless response. Keep unobserved callers’ result shapes intact.
6. `transaction-http.ts:346–364` must apply existing context(response, capturedContext) to these three selected direct responses, preserving their existing status, Problem JSON/media, limit, retry, cache/Vary and bodyless406 semantics. Require the observed tuple for these known paths. Do not attach a blanket fallback context to all errors.

Suggested production ownership is only transaction-store.ts, transaction-submission.ts and transaction-http.ts. Their result types need the narrow context members; service/comparison already pass results through and should not need behavior changes. If implementation needs another source file, report the specific reason before expanding ownership. No schema, spec, catalog, runtime configuration, launcher, installed controls, conformance manifest or coverage changes.

## Same-seam scope and remaining boundaries

| Disposition | Same authoritative decision? | Proposed coverage |
|---|---|---|
| Direct409 idempotency-conflict, ordinary or provably different unresolved-template key | Yes, admitKey within observed atomic admission | Required; both singleton and batch HTTP carriers, preserving conflict-before-negotiation |
|413 limit-exceeded, limit transaction.operations, unknown direct key | Yes, same admitKey invocation | Include; preserve capacity-before-negotiation and no key binding |
|406 unacceptable Accept on an otherwise eligible fresh request | Yes, same admitKey invocation | Include; no body or new receipt |
|406 unacceptable Accept for an equal pending/terminal retained request | Yes, outer admission callback decides it before release | Include; retained receipt unchanged |
|503 comparison deadline or epoch/lifecycle exit | No; comparison loop returns before its next owner admission at transaction-comparison.ts:43–46 | Separate follow-up; no reuse of an old comparison-pending tuple, no fabricated last-known context |

Earlier syntax/media/bounds/authentication failures and sequence-envelope refusals do not traverse this decision seam. This bounded plan makes no claim to repair every Scope-bounded response. In particular, do not infer anonymous/forbidden view disclosure behavior, unsupported minimum error behavior, or global post-error owner reads from this repair. Broad normative response coverage remains unfinished; exclusions are not waivers.

## Discriminating source tests

Use existing test files and fixtures; no new production seam.

- **Owner transaction:** extend the pattern at transaction-store.test.ts:3031ff. Seed a retained transaction at pos-1 and an unrelated real commit at pos-2; exercise observed conflict, operation-limit, fresh406 and retained406. Assert all three context fields, position pos-2, frozen tuple, no receipt/effect changes, and that the position metadata read occurs while DatabaseSync.isTransaction is true. Cover ordinary normalized and provable dependency-template conflicts. Preserve the legacy unobserved result shapes. After a later real commit, previously returned tuples must remain pos-2.
- **Submission propagation/order:** extend transaction-submission.test.ts:154–203 and the existing dependency comparison cases. Use observed=true to assert captured context survives translations. Prove conflict beats unacceptable Accept, operation-limit beats fresh negotiation, equal replay still bypasses fresh capacity, and both fresh refusal keys remain usable. Keep pending/timeout behavior and the original finite deadline intact. For the existing wait seam, allow an unrelated real commit before the final comparison; assert the final decision’s context, not a preceding attempt’s head.
- **Actual HTTP:** extend transaction-http.test.ts:248–274 and adjacent context checks. Assert presence independently before tuple equality; use a retained key from pos-1 after an unrelated pos-2 commit so copied receipt context fails. Cover singleton and batch409, fresh/retained406 and413. Validate both direct Problems against transactionalProblem;406 remains empty. Read retained receipt/current collection/finite Events after refusals to prove no changed outcome or extra effects. No mutation minimum headers.
- **Sensitivity:** dropping only the HTTP context application must fail named header-presence assertions; substituting old retained execution position must fail explicit pos-2 assertions; moving the context read outside the SQLite decision must fail the owner transaction assertion. These controls should use isolated source derivatives/local test instrumentation and never alter the positive source or introduce target response rewriting. Do not require additional installed cohorts for every source control.

## Proposed actual installed proof

Reuse existing launcher, raw HTTP executor and installedTransactionBinding conventions from adjacent installed tests. A small dedicated `apps/bdptest/src/transactional-direct-refusal-installed.test.ts` can own the proof without changing their helpers or the cumulative matrix. Use one Task Type and the real default32-operation limit; do not add a configurable installed maximum or use source-only test seams.

One child,14actual requests, no reopen:

1. Singleton create beads/a under retained key; validate receipt/Resource and current pos-1.
2. Batch create beads/b under a fresh key; validate receipt and pos-2.
3. Singleton changed create under retained key:409/idempotency-conflict, all three headers at pos-2.
4. Equivalent changed one-operation batch under retained key with unacceptable Accept: same409 and context, proving key precedence.
5. Exact retained singleton with unacceptable Accept: bodyless406 and pos-2.
6. Fresh singleton create beads/c under key fresh-negotiation with unacceptable Accept: bodyless406 and pos-2.
7. Fresh batch of33valid explicit-ID ordinary creates under key fresh-capacity with unacceptable Accept:413/limit-exceeded/transaction.operations and pos-2. Keep body below normal installed limits; no operation executes.
8. Actual GET of the seed receipt root: complete receipt equals its validated initial body; current serving tuple remains pos-2.
9. Finite beads collection: exactly a,b with authored literal facts and next:null.
10. Finite Events: exactly the two creation facts, ordered expected subjects/ordinals/transaction relationships and next:null; response pos-2.
11. Accepted create c under the same fresh-negotiation key: completed,pos-3, proving the earlier406 did not bind it.
12. Accepted one-operation create d under fresh-capacity key: completed,pos-4, proving the earlier413 did not bind it.
13. Finite beads collection: exactly a,b,c,d and next:null.
14. Finite Events: exactly four creation facts and next:null, response pos-4.

The original frozen target is a useful installed negative control: run the identical selected proof against a separately identified original payload; it must stop at request3’s exact missing-context assertions, after the two seed assertions passed. Keep old and repaired source/payload identities separate. No blanket expected-failure acceptance or relabeling of old payload evidence.

Proposed ceilings pending root acceptance:14requests/1child per full cohort; positive plus original-source negative at most28requests/2children, with expected negative early-stop3 and actual counts recorded. Existing5s per-exchange,90s cohort and120s outer limits remain. Before any execution, measure the exact route inventory,14request bodies, largest request/response estimates, focused test additions and source/file footprint; root accepts those actual bounds. A33-operation valid batch is about a few KiB, but its exact authored byte count must be reported rather than estimated as evidence. Require clean shutdown on success/failure and retain failed attempts.

## Gates and known questions

- Root decides whether to dispatch the recommended three-disposition seam or conflict-only first; neither choice needs new product semantics. All three share the source boundary, while failure/negotiation preservation determines the focused tests.
- Root must release/reconcile transaction-http.ts and transaction-http.test.ts ownership with SSE and pin the final ordinary parent. No work in this report assumes PR115 or SSE merged.
- After accepted authoring, freeze source, build before emitted tests, run focused owner/submission/HTTP tests plus applicable type/format/lint/boundary/Read-seal checks. Fresh exact-source pack/install for the bounded observation; no broad historical matrix repetition unless shared-source evidence reveals a regression needing it.
- Publish exact commit/tree/parent, changed files, actual counts/timings, canonical routes/configuration/policy/auth/fixture identities, pack and installed payload hashes, complete result artifacts, negative assertion IDs and child cleanup. Source provenance remains launcher/reviewer binding, not target self-attestation.
- Stop for new normative questions, accidental changes to admission order, synthetic or unrelated context reads, new runtime seams, source ownership conflict, or budget overflow. No full-profile/public-discovery qualification follows.

Read-only audit completed without code changes, tests, builds, installations or network operations. This plan is the only new artifact for the follow-up; the original cross-carrier plan and review remain preserved.
