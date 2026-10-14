# Scope start=now resume v2 — independent native closure

No actionable findings. All eight external findings are addressed without weakening the accepted scope, budgets, or oracles. This is a plan closure, not implementation approval or executable qualification.

## Reviewed identities

- Revised plan: `transaction-start-now-resume-plan-v2.md`, SHA256 `462b96b97c041e4b1a4f66c9d3a51e7c5de35617cb66f16eae58eca996179a21`.
- Dispositions: `transaction-start-now-resume-plan-dispositions.md`, SHA256 `2b5558c5c0eec6f4525028f6942933628271fae64aaf375284ece355a79b9c63`.
- Preserved original: `transaction-start-now-resume-plan.md`, SHA256 `02cdb76e8f0f7282e8fa8887ffc161533b5672139cee282c9d374a17d02b2e11`.
- Read the full external review and original native review, and compared the original and revised plans.
- Source inspected: `/Users/dbox/repos/wt/vickie/vickie-noop-conformance`, clean HEAD `8d66855ec4e26d3af0b3d7d22d621e1f37005294`, tree `a1ae75b822e1d7fadd739d818d09a25fb788e380`, the specified frozen 5995-equivalent tree.

## Finding closure

| External finding | Independent disposition |
| --- | --- |
| 1. Second stale progress paragraph | Closed. V2 explicitly updates both progress ranges 1337–1343 and 1407–1410, preserving the browser-interoperability caveat and leaving the separate Event paragraph unchanged. Both stale refusal descriptions exist in the inspected source. |
| 2. Explicit-start citation continuity | Closed. The clarification attaches to the start=now bullet at spec line 5972. V2 explicitly preserves the contiguous, wrapped explicit-start text at 5974–5975 and its catalog/packet citation. Only the reconnect selected text is amended. |
| 3. Incorrect universal old-guard failure expectation | Closed. Newly admitted selected-SSE start/header cases discriminate the old refusal. The malformed/repeated-header finite-JSON case is explicitly a behavior-preservation control, expected to pass both implementations. This matches the media-gated refusal in `transaction-http.ts:427–455` and the existing JSON regression in `transaction-changefeed.test.ts:622–623`. |
| 4. Drifted source anchors | Closed. Corrected HTTP and service ranges match the actual parser, override and single-anchor paths; effective-cursor and truncated-frame test anchors are corrected. Named tests remain the durable locator. |
| 5. Catalog mirror title/docstring | Closed. The existing catalog-test ownership now explicitly includes title/docstring and affected-ID changes for the reconnect mirror, without implying executable conformance or widening the catalog contract. |
| 6. Event grammar ambiguity | Closed. V2 explicitly says Event queries remain after-only and their existing Last-Event-ID behavior is unchanged. No Event route/parser is authorized. |
| 7. Installed helper scope | Closed. The proposed module-local factory captures exactly address/token/controller and a fresh key counter, matching the existing nested helpers at `transactional-sse-installed.test.ts:110–157`. It preserves the old request sequence, 15-second idle wait, assertions, executor configuration, recorder and cleanup. No shared framework/export change is introduced. |
| 8. Toolchain evidence before build | Closed. V2 requires explicit absolute Node24.16/pnpm11.20 version checks and comparison with repository pins before build/pack, with pinned PATH propagated. Later process-version output is correctly supplementary. |

## Mechanism and bounded evidence

The two-line selected-SSE `start-now-last-event-id` refusal is the identified production change. The existing HTTP path first validates query grammar, then validates repeated selected-SSE header fields and substitutes the saved header as the effective `after` request before calling the service. Removing the refusal therefore does not make a header repair absent, conflicting, repeated or unknown query parameters. `transaction-service.ts:366–389` captures the effective intent once; header-derived `after` is retained through minimum waits, while genuine headerless start=now still captures its initial anchor once. No service/store rewrite is needed by the inspected plan.

The revised spec text explicitly applies the settled exclusive-after checkpoint ruling to both original query forms while preserving durable-application and invalid/expired/foreign cursor obligations. The finite-JSON control does not purport to be a red old-code discriminator.

Eight-file ownership is unchanged. The installed positive observes offline G2 before creating G3, so ignoring the saved header while returning 200 cannot pass by merely observing later live traffic. Full group/checkpoint and finite-response byte comparisons remain; the derivative intentionally removes only the start=now header override. Its named missing-G2 failure remains required. The preserved idle scenario is separate evidence.

The ledger remains eight requests/one child for the new positive, four requests/one child for the existing idle scenario, and at most eight requests/one child for the derivative: at most 20 requests and three children. Existing per-request, recorder, cohort, outer and cleanup bounds remain unchanged; the 150-second driver ceiling is not a per-test timeout increase. No installed extra requests, deletion files, schema/public admission changes, or weakened poststate/checkpoint assertions were added.

This review ran no tests, builds or network operations and changed no source. Only this closure report was written. Root acceptance and later exact-source implementation review remain separate gates; no profile, browser interoperability, or replica qualification is claimed.
