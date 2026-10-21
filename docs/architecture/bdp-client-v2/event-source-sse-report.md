# G2b Event Source SSE — frozen implementation report

Source: `33ebad3ec539852ff46990a3647ca8195276a9f8`, clean branch `codex/vickie-event-source-sse`, checkout `/Users/dbox/repos/wt/vickie/vickie-event-source-sse`. Ordinary ancestry is assigned `15c2641053609081b2f2ff63cd734770b64f7859` → owner checkpoint `e6b44a0038371ef72a40c919a75c8f0d46af4d37` → serving commit `33ebad3`. Root owns publication and independent review. No outward changes performed.

Accepted plan SHA256: `e390c9416b0b307a99fb270188936fa513fc1ac0584ca112036928871aee2364`. This is supporting implementation evidence, not a new plan of record. `claimEligible: false`.

## Implemented boundary

- Owner-local immutable WeakMap continuations with source/context/head/within-group ordinal. One committed group per entry; no group statement construction or body GET at caught-up idle. Constant per-entry metadata/identity work remains. Multiple packets can reparse one large group, never rewalk unrelated historical groups while idle.
- Exact finite collector byte reservations, order, membership and lookahead preserved. Small per-group projection and candidate-size helpers are shared; no general scanner framework, durable log, cursor/format or write-path change.
- One shared write/drain pump serves groups and Events. The pump adopts only delivered successors, drops prior packet references, yields between historical packets and checks heartbeat deadlines even for empty backlog. Frame/comment backpressure forbids another read. Initial HTTP result relinquishes its packet.
- Scope, Bead and Link Event SSE uses existing routes/media negotiation/auth/minimum/conditional/lifecycle behavior. One capacity slot across changefeed and Event streams. A single Last-Event-ID overrides after; JSON ignores it. Known-source400 and capacity503 carry the original captured owner context, demonstrated across an intervening commit; no later inspect.
- Runtime settings bind actual one-group/4MiB Event mechanics. Existing transport idle10s, request10s, headers10s, keep-alive1s, poll100ms and heartbeat5s defaults remain unchanged. Finite-only short-timer test seams still cannot register live SSE.
- Retained deletion Sources, metadata-only incident-Link projection refusal, genuine post-header cause preservation and ordinary disconnect cleanup remain explicit. No new error Event, second status, wire cursor, public export or profile advertisement.

## Owned files

13-file diff from assigned15c: transaction-store/events/service/sse/http/runtime/runtime-configuration and their focused tests; new server transaction-event-sse.test.ts; new private app transactional-event-sse-installed.test.ts; transaction-implementation-progress.md. No protocol, schema, catalog, evaluator, mutation, Type, History, selector, public export, installed launcher, raw recorder or conformance runner edit. No new emitted production module or smoke inventory change.

## Acceptance evidence

| Predicate | Executable evidence and limits |
|---|---|
| Gap-free Scope/Bead/Link parity | Actual HTTP full Events and raw finite data substring; live update exactly once and no-op no Event. Normalized numeric value from1e2, numeric-like property names, UTF8/newline source witness. Installed witness is separate and uses simpler compiled numeric fields. |
| Cursor/error ordering | GET/HEAD malformed/empty/repeated/genesis/wrong-source/unissued/stale/foreign matrix; valid header overrides stale/foreign/unissued query, query grammar remains pre-owner; no-auth/unknown nondisclosure; future minimum beats invalid header; JSON ignores header; actual mixed Accept tie/winner matrix. |
| Ordinal and byte boundary | Self-Link endpoint facts and every reconnect suffix; small item pages;32 updates under a long canonical Scope exceed4MiB and pin identical finite/stream boundary and absolute continuation, preserving same-group suffix. Original finite count/lookahead tests remain passing. |
| Deleted Sources | Real canonical GET404 while retained Bead/Link SSE matches finite including deletion. |
| Capability integrity | Frozen continuation, fabricated/other-owner/reopened/closed refusal; regressed head and injected commit fence; old genuine handle replay is deliberately allowed. Fixed owner epoch/view cannot change at runtime: stale/foreign cursor tests do not simulate grant transitions. |
| Sparse work | Actual SQLite body GET instrumentation:13 historical groups once, zero idle historical body GET/prepares, new unrelated group once, later relevant mutation once; reader preparation bounded per newly inspected group. Existing finite two-preparation witness retained. |
| Context, conditions and lifecycle | Known-source400 and capacity503 retain earlier position after an intervening committed mutation; HEAD304412/406 without slot; cross-kind capacity in both directions. Event pre-abort, minimum disconnect and idle close; shared G2a lifecycle/reopen/fault regression suite still passes. |
| Backpressure/backlog | Blocked Event prohibits pulls/comments while owner mutation commits; exact successor on drain. Scaled backlog exceeds heartbeat interval; separate blocked-comment witness prohibits further pulls/writes and removes listeners on abort. |
| Partial delivery | Real TCP proxy forwards only an exact prefix of actual Event response, withholding final frame LF; recorder accepts no partial frame, reconnect replays from last fully accepted application position (none before first frame). Same-group between-ordinal reconnect is independently proven at owner boundary. |
| Explicit projection gap | Metadata-only incident-Link update causes initial bodyless private500 for Bead, while Scope/Link still stream; later gap closes after200 without Event/status fabrication, shutdown retains cause; repeated reconnect deterministically reaches same gap. |
| Bounds | Existing group bound plus exact4MiB finite reservation; Event frame arithmetic4,194,816 bytes below pinned Node string ceiling. Memory/liveness claims remain bounded-materialization and cooperative scheduling, not a maximum synchronous parse latency promise. |
| Installed proof | Fresh exact tarball results are appended below. Observer uses existing source-only raw recorder, target loads installed child payload. Canonical Host/physical port0, generated CSPRNG credential, metadata/configuration identity, raw wire and hashes retained. |

## Verification

Pinned Node24.16.0 / pnpm11.20.0. External log directory is this report's parent.

- PASS clean build (`final-build.log`); fresh packed run also rebuilds frozen source.
- PASS server regressions:5,442 tests /61 files (`server-regression.log`,21.69s). This run preceded the final three test-only mixed-Accept cases; all production files were already final. The final Event file then passed18 tests (`final-event-tests.log`) including those three cases. No claim of a full workspace run.
- PASS typecheck, dependency boundaries, format, lint, public installed smoke, development smoke,7-case preflight, unchanged sealed Read evidence (`typecheck.log`, `check-boundaries.log`, `format-check.log`, `lint.log`, `smoke.log`, `smoke-development.log`, `preflight.log`, `evidence-verify.log`).
- Preliminary dirty-source installed-target test passed15s idle (`phase3-tests.log`,40tests/3files). This is not the frozen tarball proof.
- Preserved development failures were test/scaffolding fixes: oversized initial owner fixture exceeded existing properties budget; omitted EventStream return union caused an intermediate type/build failure; duplicate default Accept in rejection fixture selected SSE correctly; Error subclass name assertion assumed a name the class does not set; SQLite overloaded get mock needed explicit unknown comparison. Their original logs remain. No timed-out run is counted as success.

## Remaining boundary

Metadata-only incident-Link membership remains an unresolved deterministic projection gap; erasure/revocation/pruning publication fences remain separate. The scan fast path assumes append-only committed history and does not rehash already traversed rows for out-of-owner modification. Browser EventSource/CORS interoperability and mutable authorization are not qualified. G2a start=now plus Last-Event-ID gap remains unchanged. Set HTTP admission, remaining required consistency/discovery and full Transactional profile qualification remain open. No Read sealed evidence was regenerated/relabelled.

## Fresh packed results and cleanup

All stages exit0 against frozen33ebad3; `packed-33ebad3e/stages.json` and `source.json` record execution and pack hashes.

- PASS fresh packed installed Event SSE:1 test,15.79s total. Sparse Bead idle15,157.53ms;2 heartbeat comments, then1 relevant Event. Four actual wire recordings cover Scope/Bead/Link initial replay and sparse Bead continuation. `transaction-event-sse-observation.json` and `transaction-event-sse-{0,1,2,3}.http` retain complete headers/frame observations/raw bytes/hashes. `workspaceDirty:false`.
- PASS fresh packed installed G2a SSE:1 test,15.77s total; idle15,160.83ms (`transaction-sse-observation.json`, `transaction-sse.http`).
- PASS packed finite Transactional/Snapshot:30 tests/2files,19.64s (`transaction-finite.log`).
- PASS packed cumulative Read+Update/exact-client/installed Read/development regressions:152 tests/4files,83.48s (`read-update.log`, `ru-progress.json`). Earlier progress used the predecessor's151 count; this frozen cohort actually contains152.

Tarball SHA256: protocol `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`; server `83c9cb7a02a512cc6e108766b5d7775d9331b21185067b95a123a5885e021bc8`.

Installed Event observation payload digest `9add08f913ba0bb5960926988b181987a6ddc150a4840d16008ec7b036a6b16a`, captured contracts digest `826a867731bf28d606110d97443772081cbcba3986c1a5b6ced66082c6bfb094`, runtime configuration identity `ece24118bb242b1245afd23a456e37f7dd69ada30eb70f9a8c1f34f39478d401`. Credential/policy association hashes, exact timer settings and fresh epoch/view values are retained in observation JSON; no token is logged. Scoped responses use canonical Host with ephemeral physical port.

All owned commands completed normally, no timeout/interrupted result counted, no test-owned Node children remain (`process-cleanup.json`). Frozen checkout is clean. No process required forced cleanup and no source was changed during packed execution.
