# v3 delta

Actual Claude reviewed v2; its packet/output and all v2 files are preserved. Ten findings accepted (1–9,12); optional10/11 deferred by root. Plan remains closed.

- A correct baseline is1 Snapshot/1 record/2 pages, measured immediately after real capture and compared synchronously after gated calls. No maintenance override. One additional matched-context pos-01 create makes13→14 owner calls; two offline inspections reuse the same store.
- D claims only observable HEAD metadata/empty body/retained bytes, not internal nonallocation or unchanged expiry.
- E uses one callback-local exact executor with a local pre-dispatch maximum3/final3 counter and explicit wire invariants; compiles mutationReceipt locally and reuses withDeadline/finishObservation unchanged with owned clean-close-before-remove.
- Strict helper-only TypeError catch, separate lexical catch/rethrow-original, kind outside catches and synchronous recapture are explicit. HTTP undefined-minimum fence remains. C post-close probe targets the real owner.
- No known/no-minimum kind-capture refactor or added known disconnect rows. Scope7files,22source callbacks,45source network+15handler requests, installed31+3network/3children/2stores,24totalstores unchanged. All deadlines unchanged. Source/build before installed; commit after all accepted audits remains unchanged.

Native findings audit3d1fe945 confirms these corrections. Earlier native clean v2 critique missed A page-row baseline and remains historical. Deferred finding11 would add4 network requests including both seed POSTs, not2; accepted selection remains unchanged.
