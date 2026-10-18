# Selector grammar and fresh-capacity admission implementation plan

## Authority, status and dependency

Plan only; no authoring or execution release. Direct authority is `transaction-selector-command-operator-rulings.md`, SHA2565f7c78447789da4df73c616c174a23832d4b72b4c00238a3c7c26240993d2611: invalid embedded Selector grammar returns400 before key lookup, without a receipt; valid over-limit fresh requests return413 without consuming the key; equal pending/retained outcomes and409 conflict precede fresh capacity. The same record assigns mutation minimum fields to a separate earlier change. Prior audit `transaction-next-serving-blocker-audit.md`, SHA2563ba9c0a96cd29e5f7261570031206ea0083c60a37dc1aa8027d27830ed3eae48, supplies the diagnosed serving gap; its pending-decision wording is superseded by that ruling.

Inspected source is clean c02ccc3706f2e258d380805162a8c47dbb73f4e1, tree088290751951c353797b984f690b7d9fd96deb80, at `/Users/dbox/repos/wt/vickie/vickie-cross-carrier-draft`. Root has now verified PR122 merged as actual main122800511ee6306e4105e1f73bf3658f35c90a14 with that same tree088290751951c353797b984f690b7d9fd96deb80. Thus the inspected source tree equals verified merged main. The command-minimum predecessor remains not authored; root serializes that HTTP/spec change first. **Required before authoring:** ordinary reanchor to the actual verified merged tree including that future predecessor, inspect shared HTTP/spec/test differences, renew exact source and test inventory and obtain native Codex plus Claude council closure/root release. Do not assume that future predecessor merged or reuse tested evidence under a different tree.

The sequence member plan remains preserved/deferred, untouched. This slice adds serving behavior for rejected set requests; it does not add another cumulative diagnostic row, advertise either write profile, promote Preview2, change Read controls or answer metadata Link Event / retained Snapshot questions.

## Current fault and narrow design

`transaction-set.ts:132–143` converts all ReadSelectorError categories into a private development-limit. `transaction-batch.ts:168–171` invokes that gate after the private whole-input depth check. `transaction-submission.ts:79–116` delays the gap until after authentication, then returns it before key lookup. `transaction-http.ts:365–366` emits empty development500. Consequently `$..*` does not receive the new required400, and valid Selector capacity never reaches unknown-key-only admission.

`read-selector.ts` already tokenizes and parses with explicit token/operator/value stacks (`:200–423`); evaluation is iterative too. Reuse those algorithms. A second grammar, recursive-descent rewrite, larger operational limits or Infinity/MAX_SAFE_INTEGER pseudo-limit is unnecessary. The existing Read entry point must retain byte-first/early-AST-limit behavior and exact accepted grammar/evaluation results.

Introduce one **source-private complete-grammar analysis** entry point in read-selector.ts, not a package export. It accepts a primitive decoded Selector plus an explicit enclosing source-byte safety budget; Transactional passes the unchanged65536-byte request body bound. It reuses the same tokenizer/parser and returns a frozen `{bytes,depth,nodes}` summary only. It must parse the complete grammar, including trailing tokens, before reporting success; operational4096/32/256 limits are not applied during this analysis. An input beyond the caller's already-enforced enclosing byte budget is a programmer/precondition error, never an alternative public Selector capacity response. No unlimited new exported parser API.

Factor the parser's existing AST-limit check into a private policy callback/strategy passed at the same five construction points (path,group,not,logical,comparison). Read supplies exactly its existing check. Grammar analysis supplies no operational rejection and obtains exact final-root depth/nodes from the constructed tree. Tokenization, precedence, syntax/type checks and complete-input termination remain identical. AST/tokens stay local and are released after summarization. Keep all traversals iterative; do not recursively walk/freeze/stringify a deep AST. Memory is bounded by the existing body envelope, not by the larger selector operational envelope. A malformed suffix cannot be masked by an early depth/node error. No evaluation or Resource scan occurs during preparation.

## Preparation, identity and diagnostics

Keep prepareSetShape as the closed-field/cardinality/change/provenance shape validator. After **every** carrier shape/reference/label check is complete, analyze each set Selector in declaration order. Treat both `syntax` and `unsupported-feature` as invalid bounded-profile grammar. Wrap only those known parser errors in a source-private `TransactionSelectorSyntaxError` carrying the request pointer, not a complete wire response; unexpected exceptions propagate. This typed error is not a subclass caught as an unrelated private capacity failure.

Retain compact immutable complexity summaries without editing `op.input`. Recommended representation: a source-private WeakMap keyed by the already frozen PreparedTransactionSet object, populated by its grammar validation, plus a source-private capacity accessor. Missing analysis for an otherwise branded prepared set is an internal error, not success. The existing PreparedTransaction brand remains the trust boundary. No caller-supplied diagnostic or callback can choose capacity outcomes inside SQLite.

Exact source spelling of the decoded Selector remains in semantic identity. No whitespace normalization, URL decoding, JSONPath canonicalization, or conversion of embedded ID-like string literals to URLs. JSON escapes in the outer request are decoded once by the existing lossless JSON admission; Selector byte count is UTF-8 of the resulting string. A literal `%24` stays percent text. Selector string escapes remain part of the Selector source byte count and are interpreted by its grammar only. Existing numeric lexemes and all non-Selector preparation/normalization stay untouched.

Move the existing `assertTransactionInputDepth(root)` **after complete Selector grammar analysis**, so a private depth rejection cannot hide a known malformed Selector elsewhere. Its threshold32, root-depth convention and development refusal remain unchanged for otherwise valid carriers. Deep Selector parentheses are characters in a string, not deep request JSON. A valid carrier independently exceeding the private whole-input depth still takes the existing private gap; this plan does not choose a new public status or settle that separate limitation. Document that exclusion rather than suggesting this slice qualifies every request.

Pointers are `/selector` for the singleton body and `/operations/N/selector` for the batch, using original operation index including intervening Resource operations. Error offset may remain internal, but never substitute a UTF-16 character offset for the RFC6901 pointer. For multiple invalid shapes, preserve existing whole-carrier shape validation precedence; for otherwise valid shapes with multiple bad Selectors choose the first invalid Selector in declaration order. An earlier valid over-limit Selector cannot hide a later malformed Selector or other known malformed carrier shape.

## Authoritative key/capacity order

Only the real owner decides whether a key is new. In `transaction-store.ts:1050–1084`, leave all prior dependency/normalization/equality/conflict paths ahead of capacity. Under the existing short atomic owner entry, after the existing `transaction.operations` check and before Accept negotiation or receipt insertion, inspect the prepared set summaries against **unchanged** `TRANSACTION_SELECTOR_LIMITS={bytes:4096,depth:32,nodes:256}`. Return a typed `selector-limit` outcome with dotted limit name and original pointer. No receipt, claim, reservation, effect or key write is allowed on this return.

For a fresh carrier with several violations, deterministic implementation selection is: existing operation-count cap first; then first over-limit set in declaration order; within it bytes, then depth, then nodes. This is a stable diagnostic selection, not new execution semantics or a requirement that all implementations choose the same simultaneous-violation report. Every Selector was grammar-validated before reaching this point. A valid over-limit second operation causes direct413 even if operation0 would fail during execution; operation0 is never evaluated. No Selector cap is deferred to the receipt evaluator.

Extend TransactionReceivingAdmission with the narrow new outcome and include it in observed-context capture alongside operation-limit/conflict/not-acceptable (`store.ts:1381–1405`). Direct source-private store calls that omit receiving controls retain the existing operation-limit-style developer-error behavior for a fresh capacity refusal rather than widening TransactionAdmission or pretending to return a receipt. Normal HTTP/submission supplies receiving controls and receives the typed direct refusal. Do not add configurable Selector limits to those controls.

`compareAndAdmitTransaction` already propagates non-comparison-pending outcomes through its union; no loop, deadline, identity or wait behavior change is expected. Following an unresolved comparison's retraction, the normal loop re-enters the authoritative owner and then pays the fresh cap if the key is now unknown. No separate key pre-read or cached new-key result.

In transaction-submission.ts catch the typed grammar error and return400 malformed-request/pointer before principal/key lookup, preserving existing body and key-shape handling. Extend limit/pointer result types only as needed; map selector-limit to413 limit-exceeded with `selector.bytes|selector.depth|selector.nodes`, pointer and the same owner-entry context. Include the new outcome in the observed-context invariant. Grammar errors have no invented attestation tuple; capacity errors have the actual owner observation, as existing transaction.operations413 does.

In transaction-http.ts carry the validated pointer to the existing direct Problem representation and generalize the observed limit response branch to the narrow Selector union.400 uses request/malformed-request/never;413 uses size/limit-exceeded/never. Keep existing content-type, cache headers, context behavior and ordinary malformed response policy. Do not add receipt fields, operationIndex, transaction identity, Location or Retry-After to these direct errors. `schemas/bdp-v0.schema.json:4179–4230` already permits pointer and dotted limit; no schema/codegen/schema-ID change is expected. Verify against actual validators before release.

The old mixed-set execution path `transaction-batch.ts:829–940` continues selecting once from staged visible Resources, checking cardinality and applying mutations. Its existing strict selectReadResources call is unchanged. Fresh requests admitted within caps cannot trip those same static caps later. Data-dependent cardinality/validation/authorization/matched-resource/Event/duration failures still produce their existing failed receipt and rollback behavior. This slice must not turn those into direct errors or unused keys.

## Normative amendment and exact file ownership

Amend docs/specs/bdp.md narrowly under Advertised limits (currently1065), set selection (4246 onward), and admission steps2/5 (4493 onward): decoded query Selector bytes follow existing percent-decoding rules; JSON-body Selector bytes count the decoded JSON string's UTF-8, with no URL-decoding. Assign complete embedded bounded grammar to syntax step2 and fixed structural capacity for unknown keys to step5, before execution/negotiation, retaining duplicate precedence and no-key/no-receipt refusal. Distinguish these preadmission static limits from postadmission transaction execution limits (5246/5303 onward) without changing those outcomes. Do not repeat or overwrite the separate command-minimum amendment. Numerical limits remain implementation values, not universal protocol constants.

Expected owned production/doc files (7): read-selector.ts, transaction-set.ts, transaction-batch.ts, transaction-store.ts, transaction-submission.ts, transaction-http.ts (all under packages/server/src), and docs/specs/bdp.md. Expected tests (5): packages/server/src/read-selector.test.ts, transaction-set.test.ts, transaction-submission.test.ts, transaction-http.test.ts, and one new apps/bdptest/src/transactional-selector-admission-installed.test.ts. Twelve paths total. The new installed file isolates a small24-request cohort instead of enlarging the existing heavy set-effects fixture or every cumulative matrix invocation. It may use existing installed launch/binding/raw executor helpers, not factor unrelated infrastructure.

No transaction-comparison.ts edit is expected because the result union is imported; if a concrete exhaustive switch or architecture boundary requires another path, report the exact dependency before adding scope. Runtime configuration/discovery remains4096/32/256 and byte-identical; no public exports, catalog/manifest/coverage, lockfile or package metadata changes. No changes to other agents' worktrees.

## Exact Selector constructions for tests

Use named test-local literal constructors; measure their decoded UTF-8 and summary against independent arithmetic, not expected values read from the production constant alone.

- `B(n)`: valid `$[?@.missing]` plus ASCII spaces before final `]` to give exactly n decoded UTF-8 bytes. Root `missing` is a valid singular path but absent from the canonical candidate projection, so it matches no Resources. B4096 is admissible; B4097 is valid over-byte. Invalid-suffix byte specimen is B4097 with trailing ` true` (tokenizable but invalid trailing grammar), still below body65536.
- `D(g)`: `$[?` + g opening parentheses + `@.missing` + g closing parentheses + `]`. One-segment path has depth2/nodes2; g30 gives32/32 and g31 gives33/33. The latter plus trailing ` true` must be400, not413. No recursive construction is needed.
- `T0=@.missing`; `T(k+1)=(` + Tk + `||` + Tk + `)`. At k6 there are64 path leaves and63 logical/group pairs: nodes254,depth14. Add2 enclosing groups =>256 nodes/depth16; add3 =>257/depth17. The independently calculated source lengths are836 and838 UTF-8 bytes respectively; pin them rather than reusing516/518 from the shorter @.id preliminary example. Both remain well below4096. N257 plus ` true` is malformed after a fully tokenizable over-node prefix.
- Deep safety specimens:10000 parenthesis pairs around @.missing (valid, below65536 body bytes) and the same with a trailing malformed token. Grammar analysis completes without RangeError/stack overflow; valid returns its large summary and fresh submission gets selector.depth413, invalid returns400. No candidate evaluation under that depth is permitted.
- UTF-8: a valid literal comparison containing `é` and an astral scalar, padded to4096/4097 bytes. Assert outer JSON raw `é` versus `\u00e9` decoding produces the same Selector and semantic identity; JSONPath source containing a literal backslash-u escape counts those source characters, not just its interpreted scalar. `%24%5B...` as an entire body Selector is invalid and is not decoded into `$[...]`. A quoted `%41` remains the literal `%41`.
- Unsupported `$..*`, malformed `$[?@.id ==]`, and syntax-invalid suffix after each capacity dimension must share400 behavior. Retain the existing Read excluded-feature corpus, JSON string/surrogate rules, paths, precedence and byte-first limit tests unchanged.

For each bound test avoid crossing another operational cap accidentally; independent measured bytes/depth/nodes must prove isolation. For simultaneous-cap tests explicitly pin the chosen bytes→depth→nodes precedence. Whole request bodies must remain<=65536 except the separately retained request-body-bound regression.

## Source test plan and fanout

New registrations are bounded:4 parser analysis cases,1 preparation case,3 submission cases,2 HTTP cases and1 installed case =11 additions. Revise existing gap expectations in place, preserving surrounding guarantees. Before authoring release, inventory all uses of validateSetSelector, selector-limit result unions, development-gap tests and readonly/configuration assertions; after draft, list actual expanded titles rather than assuming lexical it.each counts.

Parser titles:

1. `analyzes complete Selector grammar with exact uncapped structural metrics` — B/D/N boundary pairs and fixed arithmetic.
2. `does not hide invalid Selector suffixes behind operational capacity` — the three tokenizable invalid suffixes plus unsupported syntax.
3. `analyzes body-bounded deep Selectors without recursive traversal` — exactly two10000-group specimens.
4. `counts decoded Selector UTF-8 without URI decoding or identity rewriting` — Unicode/JSON escape/percent specimens.

Existing Read Selector file has15 plain cases and four tables (2,6,14,26) =63 expanded cases at inspected source; these remain, so67 total if authoring follows this plan. Existing Read entry point still rejects oversized source before grammar and retains the current exact depth/node metrics. No network/store in this file.

One new preparation test (`preserves literal Selector identity and original mixed-carrier diagnostic pointers`) checks complete singleton/batch grammar, typed pointers with Resource/set index spacing, no input mutation, no AST/summary in normalized identity, and the unchanged private JSON-depth error after otherwise complete grammar. Extend existing transaction-set.ts deep-input regression only where necessary; do not remove it.

Three new submission cases, each using one real owner fixture and bounded requests:

- `refuses invalid Selector grammar before every key state without owner admission`: singleton/batch and both operations; known completed key, genuinely owned pending key and fresh key; invalid body with noauth/read-only also400. Use the existing real-owner spread delegating admit spy (typed wrapper pattern) and require zero admission calls. No fabrication of receipts or replacing the owner with a stub. Complete later grammar/shape failure defeats earlier over-cap Selector; correct original pointers.
- `applies fixed Selector capacity only to fresh valid carriers before execution and negotiation`: all three caps, operation-count precedence, later over-cap Selector before an earlier potential execution failure, otherwise-valid unauthorized401/403, malformed key400, new unacceptable Accept plus over-cap413, within-cap unacceptable Accept406. Assert no execution/receipt/effect/key capture via existing inspect/real-admit behavior and successful subsequent different valid request under each refused key. At most24 submission calls in this case.
- `preserves genuine pending completed and failed receipts before Selector capacity`: create genuine within-cap owner claims/outcomes; delegation-spy the production capacity accessor, requiring zero calls on equal pending/terminal receipt replay and conflicts, and a call on a fresh valid request. Equal unacceptable Accept still406; different valid over-limit request under existing key409. Retained body, identity, state and group inventory unchanged. This proves branch precedence without inventing an over-limit historical receipt, changing constants, injecting a new limit seam or claiming an installed limit-reduction experiment. At most20 submission/admission operations in this case.

The first new case has at most24 submission/admission operations. Existing `keeps unsupported Selectors and excessive private depth distinct...` changes its Selector half to400 and preserves private-depth refusal. Existing two expanded `checks complete shape/key and principal before %s Selector gap` cases become grammar-before-principal tests with the original key-reuse/no-effect checks; use valid over-cap inputs when checking401/403, not invalid syntax. Their exact request counts and legitimate source outcome changes must be documented.

Retain these exact ten existing source cases, each once, through an escaped title selector (or saved explicit title list verified before dispatch):

| File | Exact title | Owner opens |
|---|---|---:|
|transaction-submission.test.ts|compares conflict before negotiation and negotiates exact pending/terminal replay without changing it|1|
|transaction-submission.test.ts|joins overlapping calls through one pending receipt without second execution|1|
|transaction-submission.test.ts|counts every outer set before negotiation while retained sets bypass fresh operation capacity|1|
|transaction-submission.test.ts|separates cardinality syntax, numeric admission and guarded failure, preserving original pointers on replay/reopen|5 (one initial plus four reopens)|
|transaction-submission.test.ts|maps known numeric failure at the original outer index and rolls back staged work|1|
|transaction-submission.test.ts|uses the same depth boundary for singleton and wrapped set content|1|
|transaction-set.test.ts|selects staged canonical Resources once, retains indices and expands in canonical order|1|
|transaction-set.test.ts|cardinality is inclusive, rolls back the whole batch and retains original failed disposition|1|
|transaction-set.test.ts|late selected patch failure rolls back earlier selected mutations and batch creation|1|
|transaction-set.test.ts|serializes byte-bounded deep syntax iteratively without changing number lexemes or key order|0|

Those ten cases use13 owner opens across9 directories, no HTTP/listener/readonly inventory. With the exact proposed new/changed selection the source positive is87 expanded cases:67 parser +1 preparation +3 new submission +3 revised submission +10 regressions +3 HTTP. It uses37 owner opens across18 directories,18 listeners,17 readonly inventory opens and89 HTTP before controls; two controls add2 owner/listener opens,2 directories and2 HTTP, yielding39 owners/20 listeners/17 readonly opens/91 HTTP. These are the inspected structural model, subject to mandatory final-draft recount, not executed observations. No new schema-evaluator or pending-owner concurrency seam. Existing full CI remains the broad regression gate; do not add arbitrary repetitions.

HTTP: rewrite the existing `serves both set targets with method/auth/shape precedence and no unsupported-key capture` expectation precisely: unsupported grammar is400 even for absent/forbidden principal, correct pointer, no development header; valid over-cap input supplies the401/403 checks. Preserve wrong-method405/Allow and later-malformed shape checks. The existing body executes37HTTP calls, invokes receiptRows15 times, and therefore opens16 runtime/listener sessions on one directory; **do not count that as one listener**. If preserving both new invalid/auth and old valid/auth branches adds calls, measure and seek root acceptance before execution rather than hide the growth.

New HTTP case H1 `serves exact fixed Selector boundaries and leaves refused keys unused` uses one fixture and36requests:2targets ×2carriers ×3caps × (over-limit413, valid no-match reuse200, independent exact-boundary200). Its very first request is a valid over-byte singleton delete; this supports the capacity counterfactual below. No seeded Resources or mutation effects are needed. Use one terminal retainedReceiptInventory read (closes without reopen), not repeated receiptRows. Exactly24 retained zero-match receipts and no direct-refusal receipts; no resource/group effects. Correct selector limit/pointer and same-entry tuples on every413, no receipt-shaped fields.

New H2 `serves Selector syntax and prior-key precedence over capacity` uses one fixture and16requests in a fixed authored ledger: first authorized invalid singleton400; valid seed create200; invalid same-key400; valid over-byte same-key409; equal seed200; equal seed with unacceptable Accept406; invalid over-depth suffix fresh400; invalid over-node suffix batch400; invalid over-byte suffix batch400; valid over-byte unauthenticated401; valid over-byte read-only403; valid over-byte plus unacceptable Accept413; within-cap plus unacceptable Accept406; valid same-key reuse of that406 key200; batch earlier potentially failing delete plus later over-node Selector413; valid reuse of that413 key200. Pin all original pointers, old receipt fullbody and current context; one terminal inventory read checks only genuine successes' keys. No same-key ambiguity between requests; authoring table must give literal unique keys. No global fixture/clock/service change.

The prospective focused HTTP ordinary total is37+36+16=89 requests,18 listener/owner opens across3 directories,17 readonly inventory opens. The37 legacy calls must be independently recounted after precedence edits; its original15 restart inspections are part of real cost. Proposed conservative preflight stops are100 ordinary sourceHTTP,24 listeners,24 readonly opens across these three cases; these are ceilings, not permission for extra test scenarios. Exact final counts replace estimates before release. New direct-submission fixtures and selected historical source tests have separately listed owner/readonly counts, with a proposed maximum40 owner opens across the entire selected source cohort; if the actual necessary unchanged regressions exceed it, return a scoped measurement report rather than silently enlarge it.

## Two isolated counterfactuals, no new test registrations

Only after positive source validation and root approval of exact emitted-artifact patches, run two isolated built-runtime controls, each selecting one existing new HTTP case. Preserve the original build and all positive evidence.

C1 changes only the new grammar-error mapping in an isolated emitted submission module back to the former authorized development-limit result. H2's first invalid singleton must fail its named expected400 assertion with actual development500; row/case stops after1HTTP. C2 omits only the new unknown-key Selector capacity branch in a separate emitted store copy. H1's first valid over-byte request reaches the existing strict evaluator and returns actual bodyless500 instead of expected413; require the named first-status assertion and preserve actual cleanup/failure. If actual behavior differs, stop and diagnose rather than reinterpret success. These deliberately broken derivatives are never packed, published or relabeled as implementation.

Each control has one runtime/listener/owner, one directory and1HTTP; terminal inventory is not reached. Two controls therefore2HTTP/two listeners/two owners, no installed child. Source snippets, exact before/after hashes, entry module-resolution binding, fresh process/PID, selected title and full failure chain must be retained. This validates both new grammar classification and fresh-capacity placement with real HTTP, not synthetic counts. It is an execution proposal only; source authoring does not authorize it.

## Fresh installed evidence: one24-request cohort

New isolated installed file reuses existing launchInstalledTransaction, installedTransactionBinding and raw exact HTTP executor conventions from transactional-set-installed.test.ts. Fixed fixture is one Task/Bead Type with existing captured local schema `{n:number}`, retrievalUri equal to $id; no schema alias gap or Link metadata case. One child, one private database/directory, no reopen, one90s cohort deadline,5s requests, existing10s launch/cleanup,120s command/test ceiling, maxWorkers1. Zero runtime/test-only seams. Capture source commit/tree, entry realpath/hash, runtime/configuration identity and actual readiness; keep attribution limits honest.

Literal selector constructions above use @.missing for all no-match boundary operations, so they neither delete the seed nor advance state. `updateWhere` change is replace `/n` with1. Seed Bead is explicit beads/a, n0. Recovery no-match bodies are legal singleton/batch counterparts with fresh exact keys reused only where indicated.

| # | Request | Expected |
|---:|---|---|
|1|batch create a,n0,key seed|200 completed,pos-1; capture original receipt|
|2|singleton deleteWhere `$..*`,key bad-single|400 /selector|
|3|valid no-match delete under bad-single|200; proves no key capture|
|4|batch create late then malformed updateWhere,key bad-batch|400 /operations/1/selector|
|5|GET beads/late|404; proves prefix did not execute|
|6|valid no-match batch under bad-batch|200|
|7|singleton delete B4097,key bytes|413 selector.bytes /selector|
|8|valid no-match delete under bytes|200|
|9|singleton update D31,key depth|413 selector.depth /selector|
|10|valid no-match update under depth|200|
|11|batch missing delete then deleteWhere N257,key nodes|413 selector.nodes /operations/1/selector, not failed receipt|
|12|valid no-match batch under nodes|200|
|13|singleton delete B4096,key exact-bytes|200 matched0|
|14|singleton update D30,key exact-depth|200 matched0|
|15|batch deleteWhere N256,key exact-nodes|200 matched0|
|16|invalid deep suffix under seed|400, retained seed unaffected|
|17|valid over-byte different request under seed|409, not413|
|18|exact original seed retry|200 full original receipt|
|19|valid updateWhere matching absolute ID a,key write|200 completed,n1,pos-2|
|20|exact write retry|200 same original receipt; no reselection/effects|
|21|GET actual write receipt root|200 fullbody equality,current context|
|22|GET beads/a|200 n1 and expected retained revision|
|23|GET changes/?after=genesis|exact2 groups for seed/write, no refusal/no-match/replay groups, observedHead pos-2,next null|
|24|GET events/|exact2 created/updated Events, corresponding transaction/revision identities,next null|

No client receipt pointer is guessed. The existing actual runtime targets are used; capture dynamic root only from the genuine successful write. Ten admitted unique keys are expected: seed,bad-single,bad-batch,bytes,depth,nodes,exact-bytes,exact-depth,exact-nodes,write. Freeze that explicit set as the inventory oracle. There is no direct database read through the installed child; key reuse plus actual effects/group inventory establishes the wire claims. Do not assert unseen receipts merely from the count. Final report must list24 exact exchanges, one child exit, actual contexts, valid Problem/receipt/Event/change schemas, raw body/header digests and cleanup. No fresh/pending capacity-reduction claim: pending precedence remains a real-owner source test, while installed checks fixed configured limits and conflict/equal replay.

The wire subset has no cumulative matrix/catalog change and no new public qualification. It adds one24-request/one-child case to recurring CI, not a new row multiplied across53 matrix invocations.

## Authoring preflight, validation release and CI

Before source authoring: native+Claude plan review, disposition findings, root verifies merged command-minimum parent and ownership. Before execution: freeze normal-format draft with exact file hashes/diff; all new/changed expanded titles and body byte/metric tables; actual source calls, owner/listener/readonly inventory; exact installed24request ledger; capacity and pointer branch list; counterfactual patch plans; no unexpected path/config/guard changes. No tests/build/services in measurement-only drafting. Root explicitly releases each stage.

After release, one pinned offline frozen setup if needed, fresh build/type/format/lint/boundaries and unchanged Read gate. Run the67 Read Selector cases, one preparation case, three new and three changed submission cases, the exact ten historical cases listed above, and three affected HTTP cases once (87 source-positive cases in the proposed selection). Freeze exact selections/counts before command dispatch; proposed source selection stop100 expanded cases. Existing suites carry their current timeouts; outer commands120s, maxWorkers1, no automatic retries. New source tests use the repository's existing default timeouts, not a blanket increase. Run both isolated controls only under their separate exact-patch gate, retaining expected failure evidence.

After source/council closure, freeze an ordinary clean child, build fresh and pack/install the same existing two-package dependency closure with the established offline consumer method. Verify manifests and every emitted/tar/installed payload member plus pinned transitive dependency graph; exact installed entry realpath and allowlisted environment overrides are bound before launching the sole cohort. Do not inherit historical package attribution, reuse stale dist or treat copied sourceCommit environment as child self-attestation. Separate fresh-installed release is required. Stop on provenance mismatch or test/service failure; preserve all logs and cleanup, no silent retry.

Run installed24/1 once. The two source controls add2HTTP/2listeners, not2installed children. Existing cumulative/source/full workspace regression is exact-head CI after all findings are disposed, within unchanged current1200s job limit. Static67 parser cases are CPU-only; new source owner/HTTP cases add modest workload but the author must record actual per-case durations and estimate full-job increment against the actual merged predecessor, not PR120/121 timing alone. Planning estimate is20–60s incremental full CI, dominated by3 added owner cases, two new source HTTP runtimes/52requests and one installed24/1; measure it before outward integration. Existing changed legacy case37HTTP/16opens is not counted as a new full cohort. No time or capacity guard increase is authorized if the actual job lacks headroom.

Root owns final ordinary integration, exact-head green CI, native+Claude findings disposition and merged-tree verification. No release/profile promotion follows. No changes to Janet staging/playgrounds/shared databases or Trish Type work.

## Risks and explicit stop conditions

- Complete grammar analysis must not accidentally change Read's existing early-limit precedence or permit malformed suffixes. All five AST construction sites, lexer errors, trailing token checks and deep structures need review; grammar-only mode is not a larger operational parser allowance.
- The65536-byte envelope bounds parser work, but existing string scanning/tokenization can still be CPU-sensitive. Keep the maximum deep specimen finite and measure source-test time; do not add a generic parser redesign here.
- Whole-input private depth remains a gap. Moving it after known complete grammar is the only ordering adjustment here; its threshold/status is not redefined.
- Capacity summary must never enter identity or be caller-forged. Fixed constants/config/discovery and strict execution checks must agree. Missing summaries are internal errors; do not classify arbitrary exceptions as400.
- Same-entry owner context is required for the new capacity result. Missing context, execution before capacity, a retained failure for a static cap, or duplicate cap consultation is a blocker.
- Exact equality beyond current fixed caps cannot be truthfully demonstrated by inventing historical rows. The delegated real-helper call-count proof and real retained outcomes are deliberately limited; no new limits seam is proposed.
- Shared HTTP/spec predecessor conflicts require explicit ordinary reanchor, not parallel edits. A new required package/schema/runtime path or any measured workload/size excess returns to root before expansion.

Only local read-only inspection, arithmetic and this guarded report write were performed. No source/test/build/typecheck/install/pack/service/remote publication. All proposed outcomes and counts are planning expectations, not newly observed execution evidence.
