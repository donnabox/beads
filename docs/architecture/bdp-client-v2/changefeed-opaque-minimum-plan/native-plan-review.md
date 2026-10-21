# Native finite start-now changefeed plan review

Production seam and normative basis are sound. One small test-lifecycle precision must be closed before authoring; no new semantic ruling, broader workload or implementation layer is needed. The proposed five-file boundary,17 source callbacks and separate historical10/new4 installed cohorts are otherwise acceptable for the planned actual-Claude review.

## N1 — Explicit failure-safe latch cleanup for both capture rows

Plan A holds a deferred wait while the test mutates controls, commits b and asserts counters. Unlike B, it does not explicitly require unconditional latch release and observed-operation/close drainage if any assertion or controlled commit fails before release. The existing service helper registers instance.close in cleanup; service.close waits for active operations, so an unreleased latch could turn an informative assertion failure into a hung cleanup.

Minimal correction: for each A row, immediately attach an outcome observer to the returned operation, await a race between wait-entered and premature operation settlement, and put the held-wait schedule inside try/finally that releases the latch on every path and awaits the observed operation and service close. Preserve primary and cleanup errors. Register the latch release before any cleanup that may await service.close, or keep the same obligation callback-local. Apply the same entered-versus-settled guard to B. These are failure-path controls only: no added service calls, HTTP requests, stores, waits, clocks, callbacks or deadlines on the success path.

## Precision notes (no extra tests required)

- B's immediate “close has not settled” observation cannot independently prove deferred close: a .then assignment cannot become true synchronously. Keep the assertion if useful, but attribute drain evidence to the specified held-wait/abort/release schedule, the observed aborted rejection, awaited close and raw closed-owner refusal. Do not add a timer/microtask-only proof or infer all cancellation schedules.
- In D and installed timeout metadata, write the intended literal assertions explicitly: content-type application/problem+json and Retry-After 1 for the selected503 responses. This tightens the already-listed metadata assertions without dispatch or scope growth and avoids repeating the prior parity-only attribution ambiguity. Preserve GET/HEAD representation-length parity and actual context.

## Source-backed assessment

Read the recommendation, full frozen plan/selection/workload, actual service changes/run lifecycle, owner capture/readChanges, HTTP media/effective-intent path, strict/lexical consistency helpers and waiter, selected existing callbacks and cited specification passages. General remaining-read law permits this bounded-wait choice; the current-Resource/properties MUST is not extended into a new changefeed mandate. No public activation, fabricated descriptor, retained-policy ruling or foreign-opaque conformance claim is introduced.

The explicit default-strict service fourth/owner third argument preserves fresh provenance after start=now becomes a checkpoint. Existing direct start-now/after and SSE callers remain strict by default. Requiring both marker and copied start exactly now at the service avoids accidentally promoting explicit after. Owner marker is correctly described as a private caller assertion, not an inaccessible capability. No checkpoint spelling is used to infer freshness.

The ordering is correct and intentionally differs from the Event seam: synchronous intent copy then tuple snapshot before run/check; registered one-time captureChangesStart before waiter policy/clock/abort preflight; each retry uses the same anchor/frozen tuple but a fresh actual context. Owner after-string and tuple validation remain before enter; unsupported remains before context; checkpoint validity/future/retained-start corruption remain before minimum comparison and group projection. No owner entry survives a wait. C's expected clock counts1/0/1/1 follow actual waiter source. Malformed selected tuples still throw synchronously before lifecycle checks; admitted capture/policy/clock failures are registered Promise rejections.

HTTP selection media !== sse includes undefined media intentionally: matching unproved minimum refuses before a406 or conditional shortcut. Finite JSON ignores Last-Event-ID, including malformed scalar and repeated raw fields. The new scalar observation does not claim an extra repeated-header wire test. SSE reconnect precedence/effective after and strict parser remain unchanged. An owner-free Accept shortcut would violate the plan and is not proposed.

A's two opaque attempts/one held wait/three opaque clock reads plus one positive clock read are feasible; exactly one service anchor capture is distinct from its baseline capture. The separately successful known-pos2 explicit-after call proves the original checkpoint includes b. B contributes one live read plus one post-close probe; C contributes four captures/no reads; X's four invocations include its two pre-entry/post-close probes. The eight delegated callbacks sum to10 commits,11 captures,14 readChanges invocations,3 waits and15 clocks as stated. Those totals exclude HTTP-inferred owner work.

The eight existing literal source titles exist and the anchored selection matches all17 distinct names. Existing HTTP counts15+7+12 plus new10 yield44; installed10+4 yields58 total. Store/owner totals distinguish two existing reopens. Historical source callbacks/shared helpers and the14,201-byte installed callback are pinned for preservation. Fresh emitted build before test collection is necessary and correctly required. New installed cleanup has one total envelope with explicit unknown/failure preservation, actual child exit+close and no socket-census claim; historical lifecycle is not upgraded by attribution.

## Binding and review limits

Verified all679 current source hashes, clean exact5ceb213b632e17f28714684efc5e35a62b282bf6, historical installed callback byte range/hash, and these frozen artifacts:

- plan.md c38a5d4da53e0b52e90781ac47d7270ee46fdaf203ad6186d670ba8c9d826708.
- binding.json e1a3e4c2d85079e828c1136906f19a945f17939c5920489f13bb81d6e6dd0974.
- selection.json 6a04227516db1753f5ec3f8fb1d59ce8e5e0294c0b12340778c23bca6b2fc687.
- workload.json f36a0fe9561b78301abbc4f31a045e94f7ec9a2f71d44847c1ba731ca07e496a.

Reference hash changes at report time: [].

No product/helper execution, imports, tests, formatter, source edits, branch changes or actual-Claude launch occurred. Plan authority remains root-owned and closed. Counters are prospective schedules; no validation outcome or completion percentage is asserted.
