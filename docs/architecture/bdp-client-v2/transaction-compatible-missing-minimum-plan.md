# Compatible missing-Resource minimum: bounded serving repair

Supporting implementation plan under the sole Vickie design plan; no implementation authorization, profile promotion, normative change or source modification. Prepared by native source inspection, 2026-10-14. Root/council accepts the plan and assigns a frozen ordinary successor before work starts.

## Recommendation and exact basis

Repair the current development500 for a missing canonical Bead or Link, including its `view=properties`, when the complete supplied minimum is compatible with and already satisfied by the actual same-entry observation. Return the existing observed404 and current Scope context. Preserve the development gap for missing targets combined with behind/foreign contexts, and preserve the receiving gap for unsupported/unrecognized tuples. No wait, retry, token grammar, hidden-resource policy or error-precedence redesign is needed.

Basis is main `aebf656506d187ed226302b68b526d34088537c6`, available as the parent of frozen C3 `b20c92914b48e66b33b66e9e7434dbc3467fad10` in `/Users/dbox/repos/wt/vickie/vickie-tx-discovery-assembly`. Actual `transaction-read.ts` at both pins has SHA256 `27e57473295d494e172107a5ea0f939e39429d66cbb6196f065d107b0dde7e1e`. C3's read-test additions and HTTP/service changes are separate preserved work; this plan does not assume its council/packed completion. Before authoring, use root's accepted main/projection integration and verify the cited callback, HTTP propagation and test helpers still agree. Do not replace or rewrite C3's source.

Relevant existing decisions: `transaction-consistency-rulings-audit.md`, row “Unknown/deleted/hidden canonical Resource or alias with minimum,” explicitly separates compatible/satisfied404 from simultaneous incompatible/nondisclosure cases; the parent accepted this distinction. `transaction-cold-discovery-serving-next-plan.md` identifies this as an independent serving follow-on while discovery disclosure remains off. Neither document chooses the broader precedence rule.

## Settled contract and narrow exclusions

All normative citations below refer to `docs/specs/bdp.md` at aebf:

- **Reads after deletion, lines1445–1450:** ordinary GET and `view=properties` of a deleted Bead/Link return the same `404 resource-not-found` as an unknown or non-visible identity. The explicit optional retained-history410 exception is not used by this private fixed configuration.
- **Transactional HTTP consistency, lines5758–5787:** a complete epoch/view/minimum tuple permits only an eligible observation, wait/route, or specified consistency failure. For a compatible, already-satisfied current observation there is no remaining minimum failure to arbitrate.
- **Transactional Resource read context, lines5793–5799:** the Scope-bounded read reports the actual projected-prefix context. Existing materialized missing responses already have this observation.
- **Transactional conditional reads, lines5801–5811**, plus inherited conditional/HEAD behavior: no conditional shortcut skips a fence. A missing404 remains404 under If-None-Match or If-Match; HEAD has the same status/context and no body. The helper applies conditional transformations only to successful representations.

There is **no remaining normative ambiguity for this narrow compatible case**. The unknown/hidden versus foreign epoch/view or behind minimum collision is deliberately not resolved. An all-visible installed policy cannot demonstrate hidden-target authorization; this slice preserves the existing shared missing response path and makes no hidden-policy qualification claim. Alias307/alias absence, incident-Link collections, Type reads, continuations, retained Snapshots, partial/repeated/malformed/unissued minimum controls and command-field applicability remain unchanged. No new mutation barrier, polling behavior or Resource resurrection rule is introduced.

## Source trace: why one production callback is enough

Paths in this section are relative to `packages/server/src/`.

1. `authority-read.ts:977–998` enters the actual owner, captures context, materializes the requested result and invokes `checkSourceCondition` with **the same observation** and the materialized refusal. The callback is not called for a pre-context authorization refusal. `observed()` at:509–515 retains the observation on a materialized Problem; `finish()` at:852ff passes ordinary refusals through. There is no need to fabricate a second observation or issue an extra read.
2. `transaction-read.ts:56–68` preserves materialized invalid-parameter/limit-exceeded controls, then currently throws `TransactionMinimumVisibilityGap` for every resource-not-found before checking the tuple. Its normal check uses `snapshotTransactionMinimum(condition)` and `transactionReadContexts(observation).current`. `transactionReadContexts` at:135–171 rejects missing, invalid, mismatched or ahead source context rather than trusting it.
3. `minimumFor()` at:105–125 is already limited to initial Bead/Link collections and canonical Bead/Link Resource/properties requests. It passes a captured tuple as `sourceCondition`. The proposed repair does **not** widen this route predicate.
4. `transaction-service.ts` symbol `readAtMinimum` uses the existing lifecycle and `waitForTransactionMinimum`. That helper retries **only** `TransactionMinimumFailure` with reason `behind` (`transaction-consistency.ts:106–126`). Returning observed404 does not wait. Preserving `TransactionMinimumVisibilityGap` for missing-target behind cases ensures this patch does not accidentally introduce polling or change their eventual failure.
5. `transaction-http.ts` symbol `readRoute.perform` already converts a Problem and then attaches `transactionReadContexts(result.observation).current`. Its catch maps `TransactionMinimumVisibilityGap` to the existing bodyless500 marker `minimum-position-visibility`; generic errors remain sanitized500 without that marker. The general minimum receiving/subset guards remain intact. No HTTP production change is necessary.

## Precise implementation boundary

One production edit in `transaction-read.ts` plus import of the existing `TransactionMinimumFailure` class:

- Keep the early invalid-parameter/limit-exceeded handling unchanged.
- Validate/snapshot the supplied tuple and parse the current observation **outside the special missing-result catch**. Do not hide malformed context or callback/provider failures as visibility decisions.
- For materialized `resource-not-found`, call the existing `checkTransactionMinimum` with those values. If it returns, allow the existing observed Problem to flow through unchanged.
- Catch only an actual `TransactionMinimumFailure` whose reason is one of `behind`, `foreign-view`, `cursor-expired`; translate that known consistency collision to the existing `TransactionMinimumVisibilityGap`. Do not catch all Error, duck-type a `reason` property, absorb invalid owner context, or include `catch-up-timeout` merely because it belongs to the class union. The synchronous checker emits only the first three; a future unexpected reason must propagate.
- For all non-missing materialized results, retain the normal checker behavior. Known present-target behind reads still wait; foreign-view/epoch still use their existing normative outcomes.

No new helper abstraction, public option, policy seam, token codec, database mutation, schema, source-condition interface or HTTP error mapping. Do not parse `pos-N` again or compare strings independently; reuse the existing validated checker. Do not attach context to an error by taking a new observation after leaving the original entry.

## Five-file ownership ceiling

1. `packages/server/src/transaction-read.ts` — sole production change.
2. `packages/server/src/transaction-read.test.ts` — source-plane matrix, actual context/integrity/owner reuse and narrow failure discrimination.
3. `packages/server/src/transaction-http.test.ts` — revise the existing “does not invent absent Resource versus minimum ordering” test and add the compatible/deleted/property/HEAD/conditional controls.
4. `apps/bdptest/src/transactional-collection-minimum-installed.test.ts` — one separately identified ten-exchange/one-child scenario reusing its exact raw transport, fixture, cleanup and reporting helpers. Preserve the existing collection scenario and its source assertions/ledger; never splice its results into the new count.
5. `docs/design/transaction-implementation-progress.md` — exact repaired surface, preserved exclusions, links to source-bound evidence.

No edits to authority-read, service, HTTP production, runtime, parser, conformance framework or shared fixture. Root owns projection/C3 integration and other lanes. If existing test helpers cannot support the scenario without a sixth file or new runtime seam, stop and report the specific dependency rather than expanding scope.

## Focused executable oracles

### Actual Read callback

Use the existing real owner fixture, committed mutations and `minimumFor` for the full small product: absent versus created-then-deleted; Bead versus Link; full Resource versus properties; equal versus earlier satisfied minimum. Each result must be `kind:problem`, `phase:observed`, exact canonical resource-not-found code/status and valid same-entry context. Use real create/delete commits and independently expected commit positions; do not create a synthetic Problem and call that runtime coverage.

For a representative request in each Resource kind, preserve behind, foreign epoch and foreign view as **TransactionMinimumVisibilityGap**, not a returned404 and not an exposed TransactionMinimumFailure. Existing service/HTTP tests must demonstrate no minimum-wait callback is entered for these missing cases. Use an existing injectable wait counter in the HTTP fixture, not a wall-clock threshold. Unsupported position spelling/partial tuples stay parser/HTTP development gaps and must not reach this compatible branch.

Reuse the fixture's existing owner-entry interception pattern to assert missing-record lookup, source context and checking happen inside the same guarded entry; owner reentrant access fails there and a later ordinary mutation/read succeeds after delivery. Do not pin unrelated global SQL query totals. A test-only owner wrapper may corrupt sourceContext while delegating the actual owner entry: missing/mismatched/malformed context must throw its existing integrity error, **not** TransactionMinimumVisibilityGap and never404. No injected source failure may be counted as a target refusal. If a checker sentinel is used to verify rethrow, keep it in the test seam and assert exact identity; do not add a production injection option.

### HTTP contract and retained exclusions

The existing test at `transaction-http.test.ts:1206ff` includes `{}` alongside foreign epoch and pos-99 and currently expects500 for all. Split that deliberately: the compatible case becomes404 with actual current epoch/view/position; the incompatible cases retain status500, `x-bdp-development-limit: minimum-position-visibility`, no Scope-context headers and empty body. Add foreign view and an unsupported position spelling as distinct controls (the latter keeps `minimum-position`, not visibility). Preserve the rest of the inherited exclusions.

For absent/deleted Bead+Link Resource/properties, table-drive GET and HEAD. GET validates the real Problem's required fields and media type; do not close RFC9457 against extensions. HEAD validates404, no body and actual context, with existing body-length behavior. Apply If-None-Match:* and a mismatching If-Match to missing targets: neither yields304 or412. Check private,no-store and no fabricated Resource ETag. Anonymous access retains401 without context/Resource existence disclosure; this is authentication coverage, not hidden-policy coverage. Present compatible and present behind/foreign controls remain their existing outcomes.

Meaningful sensitivity controls: removal of only the compatible repair must make the actual compatible404 test fail with500; blanket return404 must fail the behind/foreign controls; overly broad catch must fail the invalid-source-context test. These can be focused disposable source mutants under the evidence directory or narrowly scoped test spies, preserving the author checkout. They do not require a second installed child or a new shared test framework. Actual commands/exits must be retained; an unexecuted mutation is not claimed as proof.

## Installed ledger: exactly10 exchanges /1 child

Reuse the current installed fixture's Task and Relationship artifacts, including existing ownership. Use two fresh terminal batch keys. Seed/delete bodies are ordinary actual HTTP batch requests, not direct DB writes. Existing `write()` accepts batches and `exchange()` gives exact per-request routes and observes bytes. The two transactions are <=3 and<=2 members, within unchanged32; maximum live resources2 Beads+1 Link. At most9 distinct request URLs (one batch target plus eight Resource/property targets). No reopen or extra seed request.

| # | Actual request | Independent oracle |
|---|---|---|
|1|Batch creates canonical Beads a and b, then Link l from a to b, with valid independent literal titles.|Completed receipt; exact three operations/allocated identities and literal Resource results, no unexpected result; effectPosition, requiredPosition and actual response position agree with the fixture's first committed mutation at pos-1. Capture actual epoch/view/position for later lower-bound requests; no fake revision.|
|2|Fresh-key batch deletes l, then a (order satisfies no-live-incident-Link deletion).|Completed receipt with exact two deleted identities; effectPosition/requiredPosition/current position identify the second commit pos-2. Link ownership may update source before its deletion inside the transaction; still one Scope position. b remains live by authored operations, not claimed independently GET-verified in this ten-request slice.|
|3|GET absent Bead record at satisfied earlier minimum from#1.|404 canonical Problem, current context from#2.|
|4|HEAD absent Bead properties at equal minimum from#2.|404/bodyless/current context.|
|5|GET absent Link record, If-None-Match:*, equal minimum.|404 (not304), canonical Problem/current context.|
|6|HEAD absent Link properties at earlier minimum.|404/bodyless/current context.|
|7|GET deleted Bead a record, equal minimum.|404 canonical Problem/current context.|
|8|GET deleted Bead a properties at earlier minimum.|404 canonical Problem/current context.|
|9|GET deleted Link l record, mismatching If-Match, equal minimum.|404 (not412), canonical Problem/current context.|
|10|GET deleted Link l properties at earlier minimum.|404 canonical Problem/current context.|

Never infer “missing” solely from the observed404: absent IDs are never authored in either transaction; deleted IDs are proven first created then deleted by successful exact mutation results. The fixture is a fresh isolated store and no other writer runs. Literal pos-1/pos-2 are independent owner-implementation expectations for these two effects; requests use captured actual epoch/view and positions, not guessed opaque values. Each response position is compared to the independently verified second receipt position rather than to its own request bound. No read may fabricate a newer position. Source tests separately check unchanged owner state after reads; this installed ledger does not claim a separate post-read head exchange that was not made.

The full GET/HEAD and foreign/behind matrix runs in source HTTP tests, **not** secretly as additional installed calls. Installed uses its fixed all-visible policy and claims no hidden-resource behavior. A local report helper can give this new scenario a distinct label and report expected/actual10, one child and its exact completed-response list; keep the original collection test's label/output intact. The existing transport helper's80-response defensive ceiling need not change, but this scenario must enforce its own <=10 and exact final10 count. Failed runs report observed completed responses and failure, never fill the missing entries.

Bounds unchanged:5s/request, current90s cooperative aggregate/120s outer, existing10s lifecycle and cleanup limits; actual exact executor body/header caps and installed target limits remain unchanged. No intentionally delayed minimum wait in this installed scenario; behind behavior is tested through bounded source seams. Measure actual timing rather than promise it. Reports retain exact source/tree/clean state, external packed install entry/source env, tarball/payload/config/fixture identity, actual request/response-byte provenance and safe completed-response ledger; no credentials, stderr or raw exceptions published. A source-only passing run does not replace final frozen installed proof. Preserve historical C1/C3 evidence and include this new scenario separately in summary counts.

## Validation, integration and executable done-when

1. Root assigns an ordinary successor from accepted current main/projection; all unrelated ancestor files preserved. Verify only the five owned files differ and all old scenario/test assertions outside the deliberate compatible-case change survive.
2. Focused source Read/HTTP tests cover the full small matrix, preserved collisions and arbitrary/integrity failure discrimination; owner remains reusable and no missing-behind poll is introduced. Run existing relevant service/consistency tests without editing them as regression coverage.
3. Pinned Node24.16/pnpm11.20, clean prebuild typecheck then build and focused emitted tests; applicable format/lint/boundary checks. Save command/stdout/exit/timing using observable-test-runs. No automatic redundant broad full suite unless changed results justify it or root requests integration validation.
4. Freeze clean source, then genuine clean pack/install and the new ten-exchange scenario at that exact head. Existing collection installed scenario may run as regression in the same file, but record its separate established count/children and do not pretend combined execution fits10/1. Root may authorize targeted test selection for this bounded proof; no silent skipped selected scenario. All owned children exit and temporary stores are cleaned on failure as on success.
5. Positive404 must carry the actual same-entry current context. Behind/foreign/unsupported requests retain their previous gaps. No arbitrary exception is relabeled as visibility; no HTTP/public admission/claims/catalog/coverage flags change.
6. Publish a source-bound local report with exact changed scope, actual test/mutant/packed exits and limits. This closes only compatible missing-resource minimum serving; it does not activate transactional discovery or qualify a whole profile.

Stop for a sixth-file dependency, an unexpected changed ancestor, lost context on observed404, failure of source-condition exception propagation, any inability to prove the two batches committed as authored, extra installed requests/children needed, or a bound overrun. Do not relax predicates or invent error precedence to finish. Current inspection found no normative decision needed for the specified compatible case. Planning performed no tests, build, server, network, repository edits or outward publication.
