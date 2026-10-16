Byte-bounded deeply nested JSON could be rejected by a private nesting cap, fail receipt maintenance or Snapshot comparison, or commit a dependent sequence and then return HTTP 500 while recursively freezing its response.

This change removes the private JSON-depth refusal while retaining existing byte, Resource, Selector and work limits. It shares the iterative JSON-value comparator with Snapshot integrity checks, filters retained receipt detail through the JavaScript parser, and freezes deep receipt/sequence observations iteratively. It preserves retry identity, no-op behavior, atomic rollback and nested immutability.

Validation at `b0213805e4b15bb7bac4e09da5fdddb236cf6ea5`:
- 43 selected source cases, typecheck, build, formatting, lint, module boundaries and unchanged historical Read evidence verification passed.
- The copied Snapshot comparator control failed with exactly the expected nested RangeError and passed independent adjudication.
- Fresh protocol/server archives were verified against source and installed into an external consumer with eight pinned production dependencies. All original 34 HTTP exchanges passed: direct13, sequence8 and fresh-depth13, including the previously failing dependent sequence. Three children exited and their stores were removed.
- Native and Claude reviews completed; findings and the comment-only follow-up were dispositioned. The original failed installed run is retained separately.

Successor `ddfb5f43a058fe0f16ff9e04e528b4e7bfc03d87` changes only two tests. It corrects the obsolete deep-carrier refusal and removes per-node matcher overhead from the independent deep-value oracle without changing depth, semantic checks or deadlines. Native and Claude reviews completed; a message-specific negative-control follow-up was verified independently. Typecheck, formatting, lint, all three focused cases and all105 cases in the two affected files pass on the final source. The earlier format-only follow-up failure is preserved separately. Production and installed-payload bytes are unchanged from b021; no new installed run is claimed.

Exact successor CI38950645644 is in progress. The prior b021 CI failure remains historical evidence; integration is held until successor CI is green.

This PR targets the existing sequence/minimum working branch, not main. Public write-profile admission, complete cumulative qualification, retained-Snapshot minimum semantics and incident-Link metadata Event membership remain outside this slice. No release or Preview2 promotion.

Plan of record: https://github.com/donnabox/beads/blob/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/design.md
Evidence: https://github.com/donnabox/beads/tree/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/transaction-fresh-depth-authoring/validation-service-successor
