# Comparison timeout: final narrow preflight closure

**Closed: N1 and both exact-optional typing blockers are corrected. No remaining blocker identified in this bounded successor review.** Root may now decide whether to release the enumerated validation. This is static preflight closure, not a compiler/test pass, execution authorization, frozen code review or public qualification.

Reviewed exact successor identities:

- preflight.md SHA25679aae00ec518107c72bc904abdd148c842036d3cdfdeafb640fdac859fc80252.
- preflight-bounds.json SHA25674200eb1f9d125a8a7856c2c1568ce4553ab198241b46402dac5fa7970e93054.
- source.diff SHA25697b5bd7840fb1101622fbedde0d037d7e40479b98316ddf5f2544967fb607b0e.

Read typing-correction.json and independently verified all before/after identities, eight preserved source snapshots, current eight source hashes and byte-equivalence of live diff with recorded source.diff. The sole source change relative to the held pretyping draft is the explicitly undefined-capable source-private observeDeadline option in transaction-comparison.ts. The other seven source files, including all tests, are byte-identical. Current eight-file diff is552 additions/35 deletions against c555fe01df5cd20454fb0d663431df816c1270aa.

The annotation now admits the explicit undefined value supplied by submission under exactOptionalPropertyTypes. Submission's inherited Omit still excludes receiving and observeDeadline; its final actual override after spreading caller options remains intact. Unobserved calls still discard an untyped callback; observed calls still bind only the owner method. No service/runtime/public configuration or behavior is added.

The isolated submission-boundary mutant now writes `context: undefined as unknown as TransactionReadContext`, using the existing imported type. This addresses its static assignability contradiction while preserving an enumerable runtime undefined property. That property reaches the existing selected strict submission assertion and differs from the expected absent property, so the intended `legacy comparison result exact` oracle remains applicable. No production context type is weakened. The generator delta is only this derivative cast.

Reconstructed all five successor mutants in memory and verified patch/mutant hashes. Comparison-related patch context/derived hashes adjust to the annotation; no additional mutant behavior changed. The preflight accurately distinguishes submission-boundary sensitivity from strict direct-comparison undefined-field coverage and records the typing successor separately from its earlier correction history.

All measured selection/request/body/deadline keys are unchanged:18 selected source cases,16HTTP/two listeners; five controls expected12/max36HTTP across three listeners; installed direct14 plus sequence13,27/two children. Request maximum277B/total629B and all12 ledger entries remain identical. No new test, network path, child, installed timeout scenario, retry or broad suite is introduced. Event-page substitution and AggregateError reporting conclusions from native-preflight remain applicable.

Initial artifacts remain under initial-preflight; immediately preceding artifacts and all source snapshots remain under pretyping-preflight. Held control-native-closure.md remains unchanged at SHA256295a59818692dcff608fe3c433eadd027229b806a17caf978ca8a67ee16e4f07; this file supersedes its execution hold only for the reviewed corrections.

No tests, builds, compiler, installs, services, product execution, network or source writes occurred during this review. Actual compilation, positive observations, named derivative failures, counts and cleanup must still be established by the separately authorized execution and frozen code/evidence council.
