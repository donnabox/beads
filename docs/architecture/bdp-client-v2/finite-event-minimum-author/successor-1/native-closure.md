# Native closure: first-attempt Event capture successor

The single P2 from the original native full-code review is closed. No new actionable finding in this narrow two-file correction. Actual Claude full-code review and separate validation remain required; no execution pass is claimed.

Production now holds minimum/captured as call-local state inside run and snapshots only inside the unchanged waiter's first attempt. The waiter therefore retains policy→initial clock/deadline→abort precedence before inspecting the tuple. run.check still throws synchronously for closing. Selected capture executes before the first await, sets the captured flag once, and owner retries use the frozen tuple; excluded arguments remain unchanged. Capture errors are non-retryable. No preflight logic was copied and no new async public wrapper, error conversion or timeout was introduced.

Five parameterized cases exactly implement the approved controls. Invalid budget0 expects the policy TypeError and zero clock calls; aborted expects ScopeServerOperationAbortedError after one clock call; throwing clock checks sentinel object identity; NaN clock checks its exact TypeError; valid preflight with malformed epoch checks the capture TypeError. Every case confirms returned Promise, observes rejection immediately, and asserts zero owner calls/waits. Each owns an independent fixture/service with registered cleanup. Added workload is5 callbacks/stores/service calls,4 clock calls,zero HTTP/owner reads/waits. Existing B still discriminates capture stability across retries; C preserves cancellation/drain evidence.

Independently verified all679 current tracked hashes; only transaction-service.ts and transaction-service.test.ts differ from the reviewed original draft. Regenerated the exact original→successor correction diff and compared the full git diff to the successor patch. The entire original service-test file is a byte prefix of the successor, preserving B/C and historical callbacks. Other four originally changed files are byte-identical, so A/D/E, original installed prefix, HTTP predicate/parser preservation, store checks and all deadlines remain unchanged. The first five selection entries are exactly the original A–E entries; five expanded parameter names are added separately. Overall new35 HTTP and combined91 emitted+14 installed requests are unchanged.

Current HEAD remains b76c3b7df77ea9b33611bc9027db651833e20abb. Frozen hashes:

- source-binding.json: c6dbf36769f443f86b34a8e0f800e6c4fdae843727712fff037004fc0457790f
- correction.patch: b30f680743db5c1669ed28e7afb35c68f04d2b5b203310222514ba55a6c5227b
- full draft.patch: 6f35130f0fe0a8c91e89e48585e138bad92d83be2d9d8844bb08f0b5492f5320
- selection.json: 2465fcdc4c318d43a792f603be827b4ac7f01e4087ab4fd4342a2bd7768cba9b
- transaction-service.ts: a3b0984a61603300024e22afc1d6826b89296bfc27d13465c1e1ede0b1791b8c
- transaction-service.test.ts: 4b61f377aaa13c1fac5a41650cf4809239b4e71a2d34c9060c2d9c7e7fdfd032

No source edits, tests, product imports, formatter, builds, installs, services, commits or pushes. Only this guarded report was written; predecessor artifacts remain immutable.
