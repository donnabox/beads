# Native closure: sequence-alias plan v2

No remaining actionable plan findings. The revised plan closes all seven external findings and native P2. This is planning closure; the compatibility and runtime assertions below still require implementation and execution.

Exact reviewed artifacts:

- `transaction-sequence-alias-plan-v2.md`: SHA256 `575a2f104f3ece1412645099a7b6839ffa278c3a0f877f5d1b528d37314b492d`.
- `transaction-sequence-alias-plan-dispositions.md`: SHA256 `0f1b6686742a479b16859062bf6358d1b513a325178b07f4857197428a257d36`.
- Original external review `transaction-sequence-alias-claude-plan-review.txt` and native `transaction-sequence-alias-native-plan-review.md`.

## Finding closure

| Finding | Verified resolution |
|---|---|
| External F1: preparation/admission/turn predicate divergence | A publishes one immutable prepared deferral fact. B requires an eligible owned row with non-null dependency to normalize, checks row/preparation agreement, and preserves transient-creator precedence before key consultation. Alias-only and mixed tests explicitly guard the execute precondition. |
| External F2: alias-only comparer accidentally returns equal | C.5 seeds unresolvedness from original alias operands and permits only conflict or pending while the template persists. Required tests cover alias-only direct singleton and one-operation-batch waiting, no pending-receipt delivery, eventual equality or conflict. The disposition correctly rejects the external review's prediction that reusing the same original locator necessarily conflicts after a repoint. |
| External F3 and native P2: live resolution of original candidate locators and false correlation conflict | C.1–4 treats every eligible candidate occurrence of an original unresolved canonical locator as the same correlated symbol before lookup, including equivalent spellings and different slots. Only genuinely new candidate locators use live comparison lookup. Fixed facts, symbol domains, pins and immutable fields remain constrained. The exact work/work versus work/B A→B counterexample, swapped slots, two-original-locator swap and fixed A/B contradiction are required. This closes the previously identified false-conflict path without introducing a general solver or new protocol rule. |
| External F4: compatibility decision | D chooses unchanged format, identityPolicy and terminal codec, confines the tag to pending dependency JSON, and requires integrity rejection of malformed/unknown forms. Required maintenance, actual old-reader process, new-reader recovery, terminal replay and exclusive-owner proofs go beyond merely asserting startup safety. Failure returns to root; no fallback bump is delegated. |
| External F5: omitted operative law and exact slot set | The plan now cites bdp.md3062–3080 and lists updateBead.bead, deleteBead.bead, createLink.source/target including pinned URIs, preserving Link-subject and putAlias-target refusals. Source aliasFields and the cited normative paragraph agree. |
| External F6: asynchronous fallback expanding shared execution | Removed. B requires the existing synchronous fenced path with no await/fallback, retains executeClaim's unresolved-state precondition, and requires the existing post-normalization fault/recovery test. |
| External F7: incomplete source spans | Preparation/refusal/normalizer ranges are corrected; the complete nextSequence tail and alias-bearing Link creator-fact writeback are explicitly included. Inspected tail still handles writeFact and attempt cleanup as described. |

## Root's explicit staging amendment

The new root instruction supersedes the plan/disposition prose that waited for C1 before all implementation. No concrete dependency prevents starting only `transaction-batch.ts`, `transaction-dependencies.ts`, `transaction-store.ts` and their separate core tests at verified main581 in isolated `vickie-tx-sequence-alias`, while C1 owns HTTP/read files. These modules already expose the preparation, reservation, stepper, comparison and disposable store seams needed for the core and recovery proofs. No HTTP source change is needed to implement those mechanisms.

Ordinary integration of frozen C1 must precede alias HTTP-file/test changes and final route qualification. One relevant staging limit: removing the preparation refusal already makes valid alias sequences pass the existing HTTP caller; the obsolete HTTP catch is not an independent feature gate (`transaction-http.ts:323–336`). Consequently the core-only branch is not evidence that HTTP remains disabled, and its inherited HTTP gap-refusal test can become stale until the explicitly deferred HTTP phase. This does not block isolated core work; it prevents claiming completed HTTP validation from that intermediate stage. Preserve C1's route/method/admission repairs when applying the final alias HTTP delta.

## Source checks and limits

Independently verified checkout `/Users/dbox/repos/wt/vickie/vickie-schema-property-change` at clean `61ed530ac6a460598e0e312f5d28ec0e6fa2a33d`; its tree `22e448fe0f517c0d6300b732de55494bd70c7938` equals local main-basis `581bad0de6f5eb098b920147f6a7e09809c553ea` exactly. Read-only checks covered the amended documents, existing HTTP preparation call, aliasFields, operative alias contract, pending sweep restriction and nextSequence creator/writeback tail, using prior inspected admission/normalizer/comparison/startup evidence. Hashes were independently recomputed. No tests, builds, services, network, agents, credentials or source edits. No new external v2 review, runtime compatibility proof, public activation or whole-contract credit is claimed.
