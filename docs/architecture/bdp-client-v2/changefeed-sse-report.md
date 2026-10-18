# G2a private changefeed SSE — frozen candidate

Source: `b64d7e46652739c6436259a8c8ed4c5e8049a077`, ordinary child of root-assigned `94ef8da934694b6ef12653851ccee0716130a8aa`.
Checkout: `/Users/dbox/repos/wt/vickie/vickie-changefeed-sse`.
Branch: `codex/vickie-changefeed-sse`; origin is GitHub gastownhall/bdp. Source is clean and immutable. No pushes, PRs, bus or launch writes. Remote session guard matched `2026-10-08T20:26:35Z-81e5`; captured in `remote-launch.md`.

Plan: `transaction-sse-next-plan-revised.md`, SHA256 `4e000306d9b20dc9a47565a489a7dc47edbfeb80f5b26425ed12d6e672bef2dd`. No store, evaluator, schema, catalog, public export, Type or ordinary H1b runner/finite recorder changes. `claimEligible` remains false.

## Implementation

- `read-http.ts`: additive source-private representation selector and conditional helper, reusing the existing media/tag parser. Existing Read signatures and JSON/204 behavior remain intact.
- `transaction-http.ts`: discriminated finite/changefeed-stream result, raw reconnect-header list, initial authoritative effective-cursor/minimum observation, SSE HEAD/304/412, exact context/cache/marker/Vary headers. JSON ignores Last-Event-ID. URL grammar still requires a start intent; selected SSE `start=now` plus Last-Event-ID remains explicit private 500 gap.
- `transaction-service.ts`: separate cancellable stream registry, capacity before callback, abort on service close, drain of finite accepted work before owner closure. No await/callback introduced inside SQLite.
- `transaction-sse.ts`: one complete frame per write, no owner pull or heartbeat while blocked on drain, abort/error/close listener cleanup, monotonic polling/heartbeat scheduling. Frame serialization uses admitted JSON values, identical to finite group encoding. It does not preserve lexical `1.0e2` spelling.
- `node-http-disconnect.ts` + `index.ts`: extract the existing private Node disconnect bridge without changing its Read behavior. TX runtime uses that signal for the already signal-aware changes/Events minimum paths and stream lifetime. Point/receipt signal behavior is unchanged.
- `transaction-runtime.ts` and configuration: actual Node transport delivery, raw duplicate Last-Event-ID capture, post-header close instead of a second response, one live stream / 100ms polling / 5s comment defaults included in actual configuration identity. Existing four server timers remain unchanged. Closed own-data transport **test seams** suppress attestation and are not installed launch inputs.
- New bounded raw SSE recorder and installed observation test; existing finite exact recorder is untouched. The old independent whole-group replica oracle was moved into a test-only helper and reused, preserving its owned-Link/source agreement checks.
- Smoke inventory/denied bare and dist imports cover both new emitted private modules. Progress document appended; historical evidence unchanged.

## Acceptance evidence

| Plan predicates | Executable evidence |
|---|---|
| 1 replay/live/no-op, 2 start/exclusive | Real raw socket replay compares exact frames with finite groups; two initial groups, later mutation, no-op silence, exclusive reconnect and reopen. Actual `1.0e2` body literal normalizes to100 in finite/SSE. Minimum wait commits a later group after start=now capture and does not skip it. |
| 3 effective reconnect | Genesis, empty/repeated/header-only, stale epoch, foreign view, never-issued position, overridden stale query, finite malformed/repeated header ignorance, explicit deferred combination. Media quality/specificity tested independently plus actual JSON/SSE requests. |
| 4 no replay/live gap | Initial page is used directly; following reads start from last published complete checkpoint. Live write after initial replay, minimum-wait race, duplicate-safe replica oracle. |
| 5 partial publication | An actual TCP proxy forwards an exact response prefix withholding the final frame LF/chunk terminator, client closes with zero complete frames, then real reconnect from genesis replays the group. Separate replica failure-before-publication leaves graph/checkpoint unchanged. |
| 6 owned agreement | Actual owned-Link group, injected inconsistent nested source revision rejected before publication, correct replay and duplicate apply accepted. |
| 7 errors/headers | Actual auth, cursor errors and preconditions; same-entry initial context, no finite Content-Length/ETag/Last-Modified for SSE; no renegotiation of cursor errors. |
| 8 HEAD/304/412 | Actual responses while another live slot is occupied; all terminate with no frames or stream slot. |
| 9 backpressure | Bounded sink returns false: one pending frame, zero read-ahead, no accumulated comments or duplicate drain listeners over many heartbeat periods; actual mutation in the same fixture still commits. This is deterministic injected backpressure, not an RSS/stress claim. |
| 10 lifecycle/faults | Actual capacity/disconnect/reopen; service pre-abort/stale registration; close aborts stream and drains already admitted scheduled sequence, whose two groups survive reopen. Actual minimum disconnect stops remaining polling. Pump preserves injected owner/transport failure identity; real post-header owner fault closes with only one HTTP status and no invented error Event. |
| 11 real timers | Installed raw observation survives15.18s idle under requestTimeout10000 / headersTimeout10000 / timeout10000 / keepAliveTimeout1000, two comment frames without IDs, then later mutation. Targeted source-only shortened-timer tests separately observe incomplete header/body and finite keepalive closure. No inference about which Node internal timer is disabled is claimed. |
| 12 installed payload | Fresh protocol/server tarballs, npm installation, canonical Host with physical port0, generated CSPRNG credential, real child process and validated runtime identity. Dedicated raw recorder does not label deliberate close a complete finite response. |
| 13 bounds | Existing finite bound186,334,744 +512 =186,335,256, checked below Node MAX_STRING_LENGTH without giant allocation. Fixed M32/E256 admission remains unchanged. |
| 14 regressions | See validation below. No public conformance/profile claim. |

## Validation

Pinned Node24.16.0 / pnpm11.20.0, maxWorkers2. All logs are external in this directory.

- `server-regressions.log`: **5,544 tests /61 files passed**, covering the entire server directory (Read/RU/TX/finite Events/feed/Snapshot), narrated transactional HTTP contracts and installed SSE source-child observation. This run preceded the final test-only raw-prefix witness and literal-input refinement; production code was unchanged afterward.
- `final-sse.log`: **22 SSE tests passed**, including final raw-prefix and wire-literal tests. Earlier `raw-prefix.log` also22; focused six-file run254 tests passed. These are overlapping counts, not additive coverage totals.
- Clean build passed (`build-final.log`, then fresh exact-head packed build). Typecheck passed including final clean head (`frozen-typecheck.log`). Boundaries, format, lint, public installed smoke, development smoke, preflight and unchanged Read evidence verify passed. Lint has existing/nonfatal style warnings; format notes the pre-existing oversized matrix file. No zero-warning claim.
- Initial development run had two fixture failures (invalid empty patch, waiting on keepalive closure); corrected, with clean final passes. No unresolved test failure.
- No full entire-workspace8k run is claimed here; root will validate the integrated result.

### Exact clean-head packed evidence

Directory `packed-b64d7e46/`; executor `packed-evidence.py`, stage transcript `packed-executor.log`.

- SSE **1 installed test passed**. `sse-observation.json` records `workspaceDirty:false`, exact source, actual imported installed settings, matching runtime configuration digest, payload identity, CSPRNG credential association digests and actual wire bytes/hash. `sse-response.http` is the response transcript (no request credentials).
- Actual idle15,180.095709ms; stream15,192.802375ms;2 comment frames;2 complete group frames;2,970 response bytes.
- Installed payload digest: `772299f215f09a5f8102ee3f1c4e0078cdb846c7908d786cf7c359c669249196`.
- Runtime configuration identity: `9dab5005d8071d25203920be0f72851a1bfa8b5391626215ee4ef173ba6df691`.
- Response SHA256: `472ddc251194d1cac70a5377f5fd7aa01ff85cdc546e2c1a04022129d805d411`.
- Finite TX + Snapshot **30 installed tests /2 files passed**.
- Packed RU: **151 tests /4 files passed** in88.11s; all pack/install/test stages exited0. Final tarball hashes and clean source are recorded in `packed-b64d7e46/source.json`.

Preliminary logs `sse-lifetimes.log` and `server-regressions.log` explicitly record dirty workspace and predecessor HEAD; they are not relabeled as tarball evidence. Actual final package proof is the separately recorded clean b64 head.

## Remaining boundaries

Event Source SSE, its sparse-source internal scan-progress issue, erasure/revocation fencing, set HTTP admission, remaining consistency/discovery and whole-profile qualification are follow-ons. No G2b implementation included. `start=now` plus Last-Event-ID still needs its separate disposition. The private all-visible fixed-view policy remains unchanged; no browser EventSource/CORS claim. Polling is bounded loopback reference behavior, not a general distributed notification system.

## Process cleanup

All owned validation processes exited normally. No timeout, interrupt, deadline widening or forced cleanup was needed in final runs. `final-processes.json` records zero matching owned Node/Python processes after the executor finished; source is clean at the frozen head. Test fixtures closed child services/sockets and removed disposable stores.
