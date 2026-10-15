Read requests with incomplete, repeated or malformed minimum-position fields currently reach a development-only500 path. This change applies the approved all-three-fields rule: return400 malformed-request after normal transport/method/auth checks and before owner reads or existence lookup. It preserves bodyless HEAD metadata, finite SSE errors, mutation POST behavior and receipt nondisclosure for valid tuples.

Targeted source checks pass, including raw repeated headers and an actual-owner no-entry witness. Removing only the new gate from emitted code produces the expected named regression failure. Actual code review, fresh installed validation and CI are pending; this draft does not claim complete Transactional profile support.

Agent-Signature: Vickie; Codex; unknown-model; unknown-reasoning
