Deeply nested JSON that fits the advertised byte limits could encounter a private depth refusal, fail receipt or Snapshot processing, or commit a dependent sequence and then return HTTP 500 while freezing its response.

The change removes the private depth refusal while retaining byte, Resource, Selector and work limits. An iterative JSON comparator supports Snapshot integrity; receipt maintenance filters through the JavaScript parser; receipt and sequence observations are frozen iteratively. Retry identity, no-op behavior, atomic rollback and nested immutability remain covered. The package smoke inventory includes the internal comparator’s four emitted files while retaining exact membership and private export checks. Deep-value test oracles retain their semantics and deadlines without per-node matcher overhead.

Validation:

- Native and actual Claude reviews completed for production, test and package-inventory changes, with findings dispositioned.
- Production source b0213805 passed 43 selected source cases and static/build/Read evidence checks. A strict Snapshot near-miss control demonstrated the intended failure.
- Fresh archives and an external consumer verified 34 real HTTP exchanges, including the previously failing dependent sequence; three children exited and stores were removed. This remains evidence for the unchanged production payload, not a new run on later test/script commits.
- Test-only successor ddfb5f43 passed all 105 affected tests locally, then the full CI test and readiness steps. Its later package smoke failure was the explicit missing four-file inventory update, preserved in the evidence.
- Final inventory repair 48067830 passed format, lint, a fresh clean build, full package smoke and the Read evidence gate (raw 0, verified). No process survivors or runtime stores remained. Exact-head CI 38954914073 is running; integration remains held until green.

This targets the existing sequence/minimum working branch. Public write-profile admission, full cumulative qualification, unresolved retained-Snapshot minimum semantics and incident-Link metadata Event membership remain outside this slice. No main merge, release or Preview2 promotion.

Plan: https://github.com/donnabox/beads/blob/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/design.md
Installed serving evidence: https://github.com/donnabox/beads/tree/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/transaction-fresh-depth-authoring/validation-service-successor
Smoke repair evidence: https://github.com/donnabox/beads/tree/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/pr147-ci-smoke-failure
