# After no-op deletion-safety plan — native review

## Finding

**1. P3 — Make the two successful-order control oracles executable and explicit.** Plan lines115–116 require the operation-swap controls to demonstrate “coherent completed results/currentpos-2” as well as failure of the expected-failed disposition. The ordinary failing-receipt assertions intentionally expect no results/effectPosition and pos-1; their failures cannot establish which completed postimage or position was actually returned. The existing runner report contains status, redacted headers, body classification/length and assertion outcomes, not the full response body (`packages/conformance/src/runner.ts:143–176,1368–1377`). A200/schema pass plus a failed `status == failed` assertion also does not by itself establish the required ordered completed result content.

**Impact:** an implementer could claim the positive discriminator from assertion failures or report fields that do not expose its required values. This is a bounded evidence-design ambiguity, not a runtime or deletion-law defect.

**Fix:** in each derivative control manifest, retain the intentionally failing expected-failed disposition assertion and add named positive assertions for completed disposition, required/effect/current pos-2, and the actual four-result order (updated b, deleted first Link, deleted remaining Link, deleted a), with literal identities/kinds and the updated b fields. Require those positive assertion IDs to pass in the actual returned exchange while the intended failure remains present. The existing manifest seam evaluates all assertions in that exchange before aborting the row; no new action, response interceptor, DSL or helper is needed. Alternatively specify an already-existing equally independent body-observation seam, with exact provenance. Include any control-only assertion additions in the measured derivative ledger, distinguishing them from canonical authored-row counts. Do not require later GETs after the intentionally failed action.

No blocking product-semantic ambiguity or other actionable defect found in the bounded plan. Recommend acceptance after this clarification and normal root/council disposition; implementation remains unauthorized by this review.

## Basis

Reviewed plan `transaction-after-noop-next-plan.md`, independently verified SHA256 `f99e40c70d621b6ce950f2210efb1f47b35be8812c6cc00b1a14e859bcab1b76`. Source basis is frozen bec97a7a00b8df663a62004e1ee640303d8527fa/treef173c70f0806b72fa75075b1cdb1ce35d0c14bf0 in `/Users/dbox/repos/wt/vickie/vickie-noop-conformance`. At final inspection HEAD had advanced to8d66855ec4e26d3af0b3d7d22d621e1f37005294 through the announced two-comment correction: diff against bec97 is solely six comment lines in the cumulative test. Runtime, manifest/coverage and app matrix bytes used for this review are unchanged. No fetch or external review was used.

## Verified feasibility and boundaries

### Operation-time refusal and atomicity

The catalog's deletion-safety row selects the single exact operation-time incidence sentence (`transactional-v1.json:215–227`; `bdp.md:2303–2306`). `resource-evaluator.ts:1023–1029` checks incident Links against staged state at deletion, and both endpoint directions are represented by the current incident query. The two proposed batches each leave one opposite Link when delete a is reached, so both fail at index2. Later suffix deletion cannot repair an already failed operation. The reordered controls remove both Links first and can legitimately complete without changing Types or policy.

`transaction-store.ts:1236–1253` rolls the effects savepoint back and constructs a failed receipt at the prior position without results or effectPosition. `transaction-batch.ts:700–707` retains operationIndex; evaluator failure table/function at140–158 gives the specified incident-links-exist/conflict/409/after-state-change tuple. Receipt failure has outer HTTP200, distinct from its would-be409. The plan correctly avoids new failure precedence claims.

The complete seed-record comparisons after each failed prefix establish restoration of the deleted Link and unchanged metadata/properties/revision on b, with independent literal tuple inventory protecting against self-consistent wrong seed/afterstate expectations. The final successful l,r,a deletion matches the current deleted-result structure (`resource-evaluator.ts:1072–1085`): `deleted.resourceKind` plus final `deleted.resource` id/type/revision. Revisions from the seed remain valid because prefix updates and Link deletions rolled back and the configured Links are unowned.

### Events, groups and captures

`transaction-batch.ts:738–771` emits created/linked and deleted/unlinked facts in operation/endpoint order. Four creation operations yield the planned eight Events; l,r,a deletion yields seven additional Events, for fifteen in the final Scope page. Updates in failed prefixes do not escape the savepoint. The unowned all-visible Scope stream avoids the pending incident-subject metadata question. Existing store tests625–696 support deletion ordering/tombstone and prefix rollback, while the HTTP test1771–1808 proves an incomplete cascade refuses atomically; these precedents are not claimed as installed evidence for the new row.

The store's finite changes endpoint returns one group per page (`transaction-store.ts:1683–1710`). The repeated seed-group oracle is valid; after success it must retain seed group/eventCount8 with observed headpos-2 and a scoped non-null continuation. The plan explicitly declines a deletion-tombstone payload claim because it never fetches the second group. Existing unused raw `/next` capture validates a scoped absolute URL after assertions and grants no receipt authority for GET; client receipt capture handles Resource/revision/transaction values. No body-capture or continuation-route extension is needed.

### Ordering, route and request budgets

Catalog deletion-safety precedes no-op, so insertion after rollback/before no-op is correct. The new row's changes GET occurs before its Link collection GET. The independent twelve-route inventory therefore correctly retains changes as one-based7/JS6 and inserts links/?limit=32 at one-based8/JS7, before beads/b. Previous route relative order remains intact; canonical digest must change and be derived anew.

The twenty-one-action ledger reconciles: nineteen raw (four POST/fifteen GET), two receipt GETs, no reopen. Parent166+21=187 actions;159+21=180 exchanges. Twelve partial rows,19 scenario/reopen sessions and20 children including bootstrap are consistent; four complete cohorts bound to80children/720exchanges. The client ledger remains two entries, not twenty-one. No action needs more than four batch operations or fifteen finite Events, below the existing32 item bounds. Actual byte sizes remain future measured evidence.

### Bounds and preservation

The proposed initial stops are explicitly planning estimates, not committed measured guards:488280+162106=650386 manifest bytes;13019+4256=17275 nodes;70662+21186=91848 coverage bytes. New446 assertion maximum is stated explicitly. Root must accept actual M/N/C/A and exact counts before installed tests; final measured margins are capped at4096/64/1024 and cannot exceed the absolute stops. All loader/transport/deadline limits remain unchanged. The plan correctly carries forward both old conjunctive guard replacements rather than overlooking a smaller live bound.

The four owned files, eleven prior raw-object pairs, fixture hash and original no-op223 assertions/controls remain preservation constraints. The final assigned parent must be reconciled after council/merge before authoring. No assertion/route/pin refresh is authorized merely because the parent changed. The new row and predecessor flags remain false;238 required plus five optional missing obligations are not represented as profile qualification.

## Review limits

Read-only Git/text inspection and small hash/arithmetic calculations only; no tests, builds, installs, network, servers, agents or source writes. One search initially named the wrong schema filename and returned exit2; subsequent file inventory located `packages/protocol/schemas/bdp-v0.schema.json`. That search error is not test evidence or a product defect. The only written artifact is this report, and no owned process remains. Installed execution, actual sizes, performance, final static pins and implementation council remain future gates.
