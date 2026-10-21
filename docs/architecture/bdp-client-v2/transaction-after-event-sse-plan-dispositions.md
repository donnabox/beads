# S2a set-serving plan: eight review dispositions

Read-only revision, 2026-10-13. Review: `transaction-after-event-sse-claude-plan-review.txt`. Original supporting plan SHA-256: `19624131f3f6a16b748d3e6a6a8fcb2453a1e7b981d9ef50ec96ca3c8fa4205e`, preserved unchanged. Source inspection used frozen310871d49c4bb9adb1177a318e06fde29b29cf96; its set/spec files are unchanged from the original33ebad3 review. Root subsequently reports merged, remotely tree-verified PR92 commit96b00e0be63994b79025fc84ba82a36a55da0bb7 with the same reviewed tree. That is the prospective implementation base, pending revised-plan acceptance. No source changes, test runs or outward actions were performed for this revision.

## F1 — Accept, explicitly an observable correction

Current transaction-submission.ts:77–92 returns a development limit before principal checks; the revision must not pretend otherwise. Proposed private subset handling completes unambiguous closed-shape/label/key syntax, checks principal, then returns the unsupported-Selector gap without key admission. It intentionally changes the unsupported unauthorized response to401/403 and removes its development-limit header. Body/key grammar remains before principal; retained-key comparison follows principal. Add singleton and mixed-batch unauthenticated/read-only cases, later malformed-member and malformed-key precedence, actual no-key/no-receipt/no-effect inspection and corrected-key reuse. This does not settle Q1/Q2's eventual normative Selector failure code or phase.

## F2 — Accept phase separation; reject the1e30 numeric counterexample

The actual implementation, protocol/json-admission.ts exactDecimal/isAdmissibleJsonNumber, compares the input's exact decimal with JSON.stringify(Number(literal))'s exact decimal. For1e30 the latter spelling is1e+30 and both normalize to1e30. There is no binary-fraction exactness or safe-integer requirement. Thus1e30 is admissible; use min:1e30 for cardinality409 on a small population, max:1e30 for numeric success and1e+30 for equal identity. This conclusion follows the actual function and normative number rule, not a new runtime execution or test claim.

9007199254740993 passes nonnegative-integer shape but its shortest round-trip decimal is9007199254740992: retained receipt200/embedded422, not pre-key400. Negative/fractional/empty cardinality remains shape400/no key. min5/max2 passes shape/numeric checks and fails cardinality409. Add discriminating tests for every phase. Correct the892–902 anchor to include the number model and cite JSON mutation admission2518–2521. Numeric diagnostics and original-request problem pointers are separate obligations (F6).

## F3 — Accept the exact stored-value nonmatch and rollback proof; avoid an overbroad limitation

Selection868–878 explicitly compares stored values. An endpoint Pinned Reference object is not equal to a URI string. The revised plan pins this through component selection and actual HTTP cascade failure: deleteWhere may tentatively delete bare matching Links while missing a pinned incident Link; the guarded Bead deletion then fails incident-links-exist and rolls everything back, with no committed group/Events. Do not invent partial matched receipt entries for the failed transaction.

Dedicated Read endpoint filters are not new set parameters. However the language permits singular nested candidate paths, so the broad claim that set selection cannot express any pinned endpoint predicate is unnecessary. The documented limitation is absence of implicit pin-transparent comparison; do not rewrite literals, strip pins or silently widen the selected population. Progress notes must describe that distinction.

## F4 — Accept wrapper-equivalent depth

The Resource singleton already checks the original operation content at depth2; batch root/operations adds those two levels. Set singleton preparation must reuse that constraint before synthetic wrapping, preserving original singleton field validation. Add identical accepted boundary and one-level-deeper refused singleton/batch cases. This remains a private development bound, not a new normative error ruling. No second parser or forged brand.

## F5 — Accept and specify a complete installed page oracle

receiptMaximumItems32 also bounds finite Event pages. The revised installed fixture uses32 small unowned Links with distinct source and target Beads in Scope: deleteWhere produces33 receipt entries,96 Scope Event facts and one changefeed group. Seed in legal <=32-operation requests, capture the pre-command Event cursor after seeding, follow actual receipt and Event next URLs to completion, and assert two receipt pages versus three Event pages of32 for this bounded no-concurrency fixture. The independent Event omission/reorder oracle runs on the complete suffix. No source cursor or continuation URL is fabricated; other Resource-specific Event lists are assessed against their own expected facts. A whole-group feed limit is not a Resource/Event page limit.

## F6 — Accept the actual conditional normative obligation, not the suggested deferral

Spec5258–5262 says that when the authority can locate the cause in the request, the problem carries a pointer into the body as submitted. The schema's optional field allows causes that cannot be located; it does not make locatable causes optional. Numeric occurrences already provide exact operation-relative pointers. Cardinality checks also know the failing bound. A path embedded in diagnostics.message is not problem.pointer; diagnostic instanceLocation, when present, is governed by its own properties/document semantics and must not be repurposed blindly.

Add bounded set failure mapping in transaction-batch.ts: numeric first offending occurrence and failing cardinality bound use the original carrier prefix; inspect other known patch locations and translate change rather than leak the synthetic propertiesChange spelling. Do not invent precise input locations from a final Resource schema path. If wider evaluator machinery is necessary, request scope instead of waiving the requirement. Original singleton `/cardinality/min` and original batch `/operations/2/cardinality/min` are executable cases. Cross-carrier retained retries and reopen preserve original receipt bytes/meaning and never regenerate the pointer from the retry. Keep carrier spelling out of semantic identity. This is a real existing-contract repair, not merely a documentation quality issue.

## F7 — Accept concrete reservation and keep ordinary runner authority separate

Reserve apps/bdptest/test-support/transactional-execution-binding.ts:50–71 and apps/bdptest/src/transactional-binding.test.ts:59–70, including exact inventory/rejection assertions. Server transaction-http.test.ts's10-target assertion is also changed to the exact twelve-target contract. Keep equality checks, not arbitrary subsets. Runtime metadata derives the handler inventory.

The nine mutation-source names in conformance/transactional-observation.ts:144–156 deliberately do not grant set receipt capture authority today. S2a's private raw installed proof does not alter that file, grant ordinary-runner admission, or earn catalog credit. Root/conformance lane must independently own that future expansion and its negative controls. This is a named follow-up, not silently completed by changing two URL inventories.

## F8 — Accept staged-selection constraint; qualify physical-scan irreducibility

The current store enumeration is SELECT IDs plus requiredResource for each row, per set operation, followed by candidate visibility and selection. Up to32 set operations can repeat that materialization. A transaction-wide frozen selection snapshot is incorrect because later selections must see preceding staged writes.

The normative requirement is a fresh logical selection against current staged state, not an immutable requirement to do32 physical full scans. A correct future index/incremental staged representation could reduce work. The separately proposed streaming memory seam preserves full selection/cardinality-before-budget order, bounded retained matches and staged visibility, but alone provides no candidate-count, CPU, repeated-scan or recursive-visibility bound. It is not part of S2a without additional shared-file approval; no new examined-work limit or Selector semantics are introduced.

## Other verified review assertions and preserved gaps

selectReadResources parses/bounds before iterating candidates; preparation's empty-candidate probe is therefore sufficient to distinguish the current supported grammar/complexity subset without depending on graph contents, given unchanged immutable input and limits. This is implementation evidence, not a protocol admission ruling. Q1/Q2 remain unresolved. Valid sets still use existing atomic Resource/owned evaluator effects, cardinality-before-matched accounting, durable receipt paging and shared singleton/batch identity. Public discovery/profile advertisement/catalog qualification, general candidate-work bounds, sequence aliases, erasure and metadata-only Event membership gaps remain unchanged.

Root acceptance of the revised plan and exact-base assignment is required before implementation. No repository source, frozen evidence, public schema or catalog was modified for these dispositions.
