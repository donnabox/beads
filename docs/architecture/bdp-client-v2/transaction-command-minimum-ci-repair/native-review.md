# Native delta review — PR123 spec boundary repair

No actionable findings in the frozen spec-only repair. The delta preserves the operator's command-minimum ruling and corrects its normative placement without relaxing either boundary guard. The proposed narrow local validation can proceed under root's release; no test pass or CI readiness is claimed here.

## Actual patch and semantics

Actual W diff exactly equals repair.patch SHA256 ececd735849ae0033817273276c673adf4f72bbea9df0865f36b516c688356fd. The only dirty path is docs/specs/bdp.md, after SHA256287a0a64bea0c1c3be51b5bb50d528283185ef7886a9c7402b4d4b5a7d5a0941, before c2bee88f7a29349f7e420f711d10f3acff1fb0559d012b0d111fd7ac21bd175d. All six other reviewed owned files retain their final source hashes. Parent remains f2ac3f980fc8d35a3b485db04b813ad1b569d94d/tree74327f052cef6026b3ea10e186523ace9c87a128; this uncommitted spec delta is not that tested tree.

Read+Update's existing Mutation command preconditions now contains the operative MUST-ignore rule, all three literal field names, presence/absence/completeness/value coverage, no minimum barrier/no identity effect, and unchanged ordinary transport/header/body/auth/idempotency/carrier/expectedRevision constraints. I compared the operative core through the expectedRevision lead-in byte-for-byte: it is moved unchanged. The applicability tail explicitly covers this profile and cumulative profiles, preserving Read+Update plus Transactional scope without introducing a lower-profile receipt obligation. Read semantics are expressly unchanged.

Transactional HTTP consistency now links backward to the already-existing #mutation-command-preconditions heading and preserves the explicit subsequent Mutation Receipt GET/HEAD minimum-controls exception. Its preceding read minimum and exact cited `It never reports success with an older position.` sentence remain unchanged. No new heading, protocol field, outcome, schema or catalog requirement is introduced by the repair.

## Boundary and selection audit

Read the complete profile-catalog-boundaries.test.ts. Its headingOwners assigns Part III rank2 and Part IV rank3. assertProfileLinks rejects the former rank2-to-rank3 link; assertPartLinks independently rejects a later-Part dependency. The repair removes that forward dependency and uses a permitted PartIV-to-PartIII reference; no alias, shared-anchor allowlist or test exception is added. Existing injected later-profile and later-Part negative cases are intact. These tests correctly exposed a missed review/selection dependency; the earlier citation-only Q pass did not establish all Markdown-link ownership boundaries.

The frozen validation-proposal.json (SHA25663cef877ccbc176b7e012878ca42be2fc0f569a6920ddfcdad6ed7ba3ee77c24) selects the entire boundary file: fourteen fixed cases plus six actual catalog JSON files plus one development catalog =21 expanded cases. The second command selects exactly the existing fully qualified Transactional citation Q title =1. Thus this artifact proposes22 CPU/filesystem-only cases, zero HTTP/owner/listener/service/installed-child work, maxWorkers1 and unchanged120s supervisor. This is a suitable focused selection for the concrete boundary failure and retained citation text.

Root separately mentioned broader catalog/contract-plan suite validation. That may be bound in a separate root command packet; the author's frozen two-command proposal does not claim those full suites, and this review does not count them as executed or included in22. profile-catalog-boundaries.test.ts, transactional-catalog.test.ts and profile-contract-plan.test.ts remain byte-identical to HEAD. Full exact-head CI remains required after committing the repair.

## Limits

Reviewed source/diff/text and selection statically only. No product imports, tests, build, installation, services or source edits by this reviewer. Original council, runtime/installed evidence and failed CI are preserved historically, not relabeled as runs against the repaired spec tree. Selector v3 remains pending; its eventual reanchor must use the final verified merged tree and refreshed line anchors. No authoring or profile qualification follows this delta review.
