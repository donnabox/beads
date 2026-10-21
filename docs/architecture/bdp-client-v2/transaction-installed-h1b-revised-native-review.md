# H1b revised plan: native independent assessment

Reviewed plan50cea58fe1987755841207cb9d843525ce01f1af3e35da5b0ee56893978e9b5d against exact8e45 runner/http-executor/report/raw-target code and frozen M1b2d6 runtime interfaces. Read-only assessment; no test execution.

No remaining architectural blocker found in the revised bounded plan, conditional on reviewed M1 integration. The previous proposal incorrectly assumed an available dynamic transport/session path. The revision explicitly owns the missing private capability and intercepts the existing client operation before programmable execution, without adding a manifest family or allowing a caller-supplied response to count as wire evidence.

Inspected seams: runCumulativeTransactionalMatrix calls runMatrix through the private brand path; snapshotRunOptions closes public options; runScenarioAction has the exchange accumulator but runProgrammaticAction does not; raw target preparation owns and hides session.dialRoute; snapshotExactHttpMode currently refuses TX before credentials/dial; report serialization explicitly rejects exact fields in v3. Revised ownership and required controls match those actual constraints. M1 returns metadata only after owner/service startup and actual bound listen URL, so prepare-before-actions is the achievable validation boundary.

Implementation acceptance reminders, already substantially in the plan:
- Update LegacyObservedExchange's omitted/never keys as well as runtime v3 serialization guards for newly added dynamic fields. Runtime-only v3 refusal does not prevent accidental TypeScript widening.
- Both ordinary exact requests and dynamic receipt requests must use the same validated session generation after reopen. Neither retaining the original base execute closure nor rewriting the old receipt provenance is sufficient.
- The private holder must revoke before cleanup and drain dynamic requests through existing pending-operation settlement. A cleanup callback that merely closes the base wrapper misses dynamically created executors.
- Whole-contract flags remain false and canonical250 denominator unchanged. Structured availability markers and source scaffolds must retire together only with actual observations; they are not public admission.

No normative semantics, public profile admission, Snapshot rules or additional carrier features are authorized by this review. Four partial observations are the bounded deliverable.
