# G2a final bounded cancellation repair

Frozen source: `bfafced4853a7813f7cfd49574ceadaf3cc63390`.
Direct parent: `b14fb2875447a2bfd9de9fa725add412ba28ad5c`.
Checkout: `/Users/dbox/repos/wt/vickie/vickie-changefeed-sse-final-repair`.
Branch: `codex/vickie-changefeed-sse-final-repair`; GitHub origin retained.

No push, PR, coordination, normative or G2b source actions. b14 and root's bcfa integration remain unchanged. Native review may use this clean frozen source while final packed evidence completes.

## Dispositions against changefeed-sse-repair-claude-review.txt

1. **Medium cancellation race closed by reproduced minimal guard.** First unexpected post-header fault is captured only while response is still live (`!response.destroyed`), before the existing finally destroys it. The deterministic test uses an existing dist-module spy to destroy the actual Node ServerResponse after headers, asserts destroyed=true and signal.aborted=false, crosses an explicit Promise microtask, asserts the abort is still pending, then runs the real pump. No production seam or timing-only stress. On unmodified b14 production, test fails because runtime.close rejects with AggregateError containing the pump's `stream transport closed`; after guard it resolves. The separate actual owner-failure test still retains the identical original error in shutdown aggregation. This intentionally shares the prior cancellation-first tradeoff when a real fault races transport destruction; it is not a durable diagnostic ledger.
2. **Low effective timer validation addressed without breaking finite seams.** Capture now validates the effective pump timer domain before serving; the default heartbeat/idle relationship assertion remains. `liveStreamsAvailable` still compares effective heartbeat<effective socket timeout. Short finite-only seams remain valid and unattestable, with live refusal; no generalized tuning seam, timeout widening or unconditional effective-idle startup rejection. Boundary tests pin timeout100/5000 unavailable and5001 available, with unchanged frozen effective timer settings.
3. **Low byte-oracle indexing clarified.** Explicit `[1,1]` finite group counts precede raw-frame/page comparisons. A changed grouping boundary now fails as a paging assumption rather than a misleading serialization assertion.
4. **Informational index docblock:** prior index edit was import regrouping only; no module docblock existed. No additional index edit.
5. **Informational installed key order:** numeric-like keys are in the source witness only. Installed witness is exact raw containment for its simpler `{n:100}` fixture, not independent numeric-key-order evidence.
6. **Informational observer graph:** app recorder comes through source-only server testing barrel, so the observer holds a source server graph. The actual target remains a separate child loading the packed installed entry with recorded identity; this is not a source-only target claim.

Five files: two production guards, two existing test files, implementation progress note. No new module/export/store/schema/catalog/public profile behavior.

## Before/after evidence

- `regression-before.log`: exit1;1 failed/25 deselected. The actual runtime close rejects with AggregateError whose cause is real pump `stream transport closed`. The failure is intentional red evidence, not an inconclusive run. Fixture cleanup still closes/removes its disposable owner; repeated assertion failure in cleanup exposes the same sticky promise.
- `regression-after.log`: exit0;2 passed/24 deselected. Includes destroyed-before-notification cancellation and unchanged genuine owner-failure retention.
- `focused-after.log`: exit0;161 tests/6 files: SSE, runtime configuration, service, changefeed, Events and Transactional HTTP.
- First typecheck found an untyped dynamic-import mock signature; fixed with the existing source module type annotation. `typecheck.log` preserves that failure; `typecheck-final.log` passes. No runtime correction was hidden by this type-only edit.
- PASS clean build; final typecheck/boundaries/format/lint (161 warnings/8 infos, no errors); installed-package/development smokes; unchanged Read evidence verification.
- No full-workspace rerun: root's broader bcfa validation covers unchanged surrounding code; this candidate has bounded affected regression and exact installed evidence. No hooks bypassed.

## Exact installed evidence

Executor `packed-evidence.py` derives source from clean HEAD, rebuilds/packs protocol/server, installs tarballs into a disposable isolated target and launches actual children. It retains explicit stage exit codes and source/tarball hashes. Only exact installed SSE and finite Transactional/Snapshot regressions are rerun; no relabeling of old b14 or sealed Read evidence.

PASS exact clean packed SSE:1 test;15164.183583ms idle,2 comments,2 change-group frames,2,970 raw wire bytes. PASS exact packed finite Transactional/Snapshot:30 tests/2 files. Every executor stage exited0; no timeout or interrupted run. Source matches `bfafced4853a7813f7cfd49574ceadaf3cc63390` and is clean.

Evidence: `packed-bfafced4/source.json`, `stages.json`, `sse-observation.json`, `sse-response.http` and per-stage logs.
Wire SHA256 `babf07f5f2dfa5b5f22b3afe035fa916a25cc6c373774e9ca997cb6b8ffadc87`.
Target entry SHA256 `bcf153170f953f9e28c6c89f1153be907f64df289e64f3ad929cad4a4a54c47c`.
Payload digest `c3fe7cf4cc30154d57686de8b042886193b419a48b3b2af8fb8139480024fde8`.
Runtime configuration identity `201b79ccf912b3988ad13566308d364044462072a5510345334500d153b50924` (fresh credential-bound observation).
Protocol tarball SHA256 `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`.
Server tarball SHA256 `965cbc5779489634355fb804b3c8b2ebbfe599041e421a4dd82fec8deed22be7`.

`final-processes.json` confirms no owned processes remain and source is clean. All actual child cleanup completed; unrelated services/reviews were untouched.

## Boundaries

No Event Source SSE, erasure/mutable-view fencing, set HTTP admission, complete discovery/profile claim, automatic EventSource reconnect for deferred start-now/header combination, RSS/latency guarantee or installed numeric-key-order qualification. `claimEligible` remains false. Root owns publication and final native closure.
