# Proposed title

Wait for opaque minima on initial Scope Event reads

# Proposed PR body

Initial finite canonical Scope Event GET/HEAD requests currently return the private minimum-position500 for an unrecognized minimum token. With a complete tuple matching the current epoch and authorization view, they now wait within the existing bound and return contextual503 `catch-up-timeout` with `retry: after-delay` when the minimum remains unproved.

This applies only to JSON negotiation with no `after` or `Last-Event-ID`. The service captures the tuple once after the waiter's policy/clock/abort preflight; retries retain that tuple and the original deadline. The owner checks eligibility before scanning retained Events. Known minima, ordinary pages, conditional precedence and cancellation/drain behavior are preserved. Unknown minima on Resource Event views, continuation/SSE requests, incompatible media and foreign contexts retain their existing private unsupported gaps.

Stacked on `codex/vickie-parser-refusal` at `b76c3b7df77ea9b33611bc9027db651833e20abb`. Six files changed; the parser region, historical callbacks and frozen PR150 are preserved. This bounded serving change does not claim complete write-profile support or enable public profile admission.

Validation:

- Fresh typecheck, format, lint and build passed on successor-3.
- All13 selected source tests passed, including new owner/service/HTTP cases, five preflight-precedence controls and four existing regressions. Immediate results/cleanup audit passed with unchanged source and no remaining sampled processes or runtime directories.
- External installed-consumer validation: **PENDING — root to replace with actual results before publication.**
- Exact final commit/head CI: **PENDING — root to replace with actual commit and CI result.**

Full six-file native and Claude code review completed on the predecessor. Subsequent formatting and the equivalent local-catch-to-direct-error-collection change received separate native/root disposition; predecessor review and failed static attempts retain their original identities. The fresh source checks above apply to the corrected draft.

---

Preparation note (not part of PR body): proposed text only; no commit, push or PR mutation. Verified source evidence at finite-event-minimum-validation/successor-2/run: focused-source.results.json reports13 passed,0 failed; focused-source.accepted-observation.json lists each selected callback; review-focused-source.exit.json acceptedtrue. Static3/build/bind exits all acceptedtrue. External results and CI intentionally unresolved here. Root owns final ancestry/scope/commit/tree/publication checks and replacement of pending placeholders.
