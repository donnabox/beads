# Native code review: finite initial Scope Event minimum

## Finding requiring disposition before validation

**P2 — Preserve waiter preflight and cancellation precedence before tuple capture (transaction-service.ts:474–486).** The new snapshotCurrentReadMinimum runs before waitForTransactionMinimum. On an already-aborted direct canonical initial events() call with a malformed supplied tuple, this now rejects TypeError; the previous code first entered the waiter, validated the wait policy and clock, then rejected ScopeServerOperationAbortedError before invoking owner.readEvents or inspecting that tuple. Similarly, invalid wait policy or throwing/nonfinite initial now() previously failed before owner tuple capture; the new code can instead reject the tuple error before ever consulting that policy/clock. Even a valid tuple is now inspected before initial now(), reversing the prior observable ordering for a custom clock. No HTTP malformed-receiving rule is implicated; this is the existing internal finite service contract. Closing refusal remains synchronous and correctly precedes capture because run.check still executes first.

This is a consequence of the accepted plan's capture-before-wait wording, not author drift. Amend that narrow plan detail rather than duplicate waiter logic. Minimal preservation proposal: keep events non-async and keep run as the entry; inside run declare call-local captured state, then invoke the unchanged waiter with an attempt callback that snapshots the selected lexical tuple ONCE on its first invocation and calls the owner using that frozen result on all attempts. Leave excluded source/after and undefined/unsupported cases unchanged. The first attempt runs synchronously before the waiter's first await, so the existing caller-mutation B oracle still proves capture before deferred waiting; policy/clock/abort preflight retains its previous order. A capture error exits without retry. Do not hand-copy budget/clock checks or introduce a second cancellation policy. Root must release this plan adjustment and any bounded discriminating direct-service checks separately; no correction was made here.

No other actionable implementation finding was established by this read-only review. Do not interpret that as a typecheck/formatter/runtime pass: none was executed, and static validation remains necessary.

## Reviewed behavior and evidence

- HTTP predicate exactly requires canonical Scope source, absent after control, JSON selection, and no normalized or raw Last-Event-ID before using lexical tuple capture. The old finite/SSE dispatch, receiving/authentication/parser ordering and unsupported fallback remain intact. Store narrowing is only initial Scope; source existence and sentinel precedence remain ahead of comparison inside the entry, and comparison still precedes group reader, envelope and projection. Known tokens delegate to existing comparison; opaque foreign context remains visibly unsupported.
- A's setup precedes its SQL spy. Entered-owner Unproved/behind/foreign refusals have zero group prepares; seeded eligible read proves the counter can observe a scan; after/Resource TypeErrors are accurately labeled pre-entry exclusions. Eight calls reconcile.
- B/C use a real-owner delegating facade, scan-free withRead probes, fixed mutable clocks, early rejection observation and deferred gate release in finally. B counts4 reads at0/0/5/10,2 waits/probes. C cancels after the first wait and releases its gate before awaiting rejection; drain starts at absolute5 and reaches15, giving elapsed0/5/10. C counts5 reads,3 waits/probes and4 service calls; closing refusal adds no owner read; post-close refusal is synchronous and separate. Cleanup observes all admitted promises and closes the service after gates are released.
- D contains32 real emitted HTTP requests with the reviewed seed/setup/control split. GET bodies use Event/Problem schemas, selected503 is exact reference unavailable/catch-up-timeout/after-delay, HEAD compares GET media/byte length without parsing a body, and selected context/retry/cache/validator checks are explicit. The private500 exclusions correctly remain empty development-marker responses, not invented Problems. The missing-source404 has actual context; malformed/parser400 has no context. The runtime is closed before directory removal, and errors aggregate. Its existing5s callback and80ms minimum budgets are unchanged; emitted code must be freshly built before runtime execution.
- E independently constructs its exact GET/HEAD routes, <=3 dispatch counter, fingerprints and actual child binding. Empty-page setup captures real context, then opaque GET/HEAD distinguish503 from old500; reference Problem shape, retry/cache/context/absence of validators and HEAD byte length are checked. Raw request completion and supplied minimum headers are inspected. The launcher uses an actual generated credential. Default development execution is still not external installed provenance; the separately required archive/consumer gate owns that attribution.
- E abort initiates the launcher's idempotent close; the explicit close call awaits that same operation under the original shared45s deadline. Removal occurs only after known code0/no-signal exit within the deadline. Failed/unknown/tardy close retains the directory and reports measured nullable metadata; late rejection is observed and cannot restore success. A missing returned launcher handle is not invented as a PID or exit. Unknown cleanup still needs the outer owned-process audit, as planned. Historical cleanup semantics are preserved separately rather than relabeled.
- All5 new callback counts match selection.json. Existing regression inventory is separate from new evidence. No callback/runtime result or conformance qualification is claimed here. No production expansion into S1/G2, Types, discovery, foreign-context policy, retained continuations or main was found.

## Exact frozen binding and preservation

HEAD b76c3b7df77ea9b33611bc9027db651833e20abb. Independently verified all679 current tracked hashes, exact six-file git diff against the frozen patch, and673 unchanged files. The installed historical prefix is exactly15956 bytes, SHA256828c05f358ae5863c992f93194da15a31df9a200e073e5e3525531ba39fe7d0d; service-test baseline prefix also remains exact. The pinned canonical parser region occurs unchanged in HTTP, and the entire HTTP test file is unchanged. Existing Event tests change only added imports/validator and appended callbacks.

Artifact hashes:

- source-binding.json: e439fb06b06dabf70445688412f1a30388298bc0880cbad59a4ccd547ed4c764
- draft.patch: ee6d34ccc8d253f845e1e7d07c469220883698382f788f69566291b6623fbf2d
- selection.json: 41a76a1ff42c1cb8d179b39dfad407be17b25c8c2c31c2d3ded66a281822a5e1

Actual source SHA256 values:

- packages/server/src/transaction-http.ts: 8b6e0be9f9df4c718a9e26d5dc321d7517780edef53b9580dae314328d6a3eca
- packages/server/src/transaction-service.ts: f1b9dd65cb59ded5d340a60c33b539799fa59e99d7e329257b358b8e3beccec9
- packages/server/src/transaction-store.ts: dd5de25d1f73faa8c14c08d94d91a4b160a4abee991f4934495260ead93a9ab0
- packages/server/src/transaction-events.test.ts: a9f893ec8bd79b40b5a3e34b002781ec765bfbec863f508bf9f793e3be5702f6
- packages/server/src/transaction-service.test.ts: 2dff5855620c5d37ee4f6e32fb8bdb7b17fc02d708525209f4a971c7914e2631
- apps/bdptest/src/transactional-event-missing-context-installed.test.ts: 83c3662385cf7049fb9450f66e5ae80c3abf21e994e10eeae4f5495108b49b7a

No tests, builds, formatter, product imports, installs, services, commits or pushes occurred. No product source or existing packet was changed; only this report was written after session guard verification.
