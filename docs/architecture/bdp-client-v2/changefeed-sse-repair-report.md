# G2a bounded review repair

Candidate: `b14fb2875447a2bfd9de9fa725add412ba28ad5c`, direct parent `b64d7e46652739c6436259a8c8ed4c5e8049a077`.
Checkout: `/Users/dbox/repos/wt/vickie/vickie-changefeed-sse-repair`.
Branch: `codex/vickie-changefeed-sse-repair`; origin is `https://github.com/gastownhall/bdp.git`.
No outward writes. Frozen b64 and its evidence remain unchanged. G2b remains plan-only.

## Findings disposition

1. **Timer relationship:** pump inputs must be positive safe integer Node timer values, at most2,147,483,647. Configuration checks the installed default heartbeat strictly precedes socket idle expiry, with no invented factor-of-two policy. Current defaults are unchanged. Root-approved short transport test seams preserve finite/HEAD deadline observations but refuse selected live streams before headers or registration with the existing private development-gap response convention. Such seams already suppress configuration metadata. Tests pin equal/too-short refusal, minimally longer acceptance, invalid pump values, finite preservation and live refusal.
2. **Private fault retention:** retain only the first unexpected post-header cause, before finally destroys the response and creates a disconnect signal. Runtime close aggregates that original cause after draining requests and closing service/owner. No error Event, second status or credential-bearing logging. Actual injected owner failure proves identity-preserving aggregation and one wire status; ordinary disconnect/shutdown tests remain clean.
3. **Test dependency direction:** raw recorder moved unchanged from app support to server support, re-exported only by the existing source-only `@bdp/server/testing` mapping. App imports that test edge; server uses its own relative support path. Original and moved bytes both SHA256 `09ff8343ac1a5dbad77eaf24fc2745cc28ea870ada6b05dc4ea1644c0ad5ef81`. No installed export or new production output.
4. **Pre-abort guard:** outer runtime failure handling tests the captured request signal before finite500 fallback. The regression deliberately aborts between handler selection and stream registration through existing module mocks; the raw connection receives zero response bytes and close remains clean.
5. **Serialization oracle:** source and packed installed tests additionally compare SSE data text directly to actual finite response text; structural/order assertions remain. Source fixture includes numeric-like keys, UTF8/newline data and the existing admitted1.0e2→100 value witness. No parse/reserialize-only byte claim.
6. **Docblocks:** restored module docblocks ahead of imports in runtime/configuration. `index.ts` actually has no leading module docblock; its injected import was regrouped with existing imports rather than inventing a disclaimer.
7. **Deferred reconnect composition:** progress documentation explicitly records that `start=now` plus Last-Event-ID prevents automatic EventSource reconnect after first delivery. Its private refusal is unchanged. No new protocol status or browser support claim.

## Source scope

Nine files changed: runtime, runtime configuration, pump, pump tests, index import placement, source-only testing barrel, recorder rename, actual installed SSE test and implementation progress note. No store/evaluator/HTTP handler/schema/catalog/public profile edits. Pump validator is source-private behind the existing denied module boundary.

## Validation

Pinned Node24.16.0 / pnpm11.20.0; frozen-lock install.

- PASS clean build, typecheck, dependency boundaries, format check, lint (161 warnings/8 infos, no error).
- PASS focused SSE/configuration47 tests.
- PASS all server source tests:5,406 tests in59 files;20.88s. This is not a full-workspace8k claim.
- PASS installed-package smoke and development smoke; public profile admission remains fail-closed.
- PASS7 E2E readiness preflight tests.
- PASS existing Read evidence verify, unchanged seal; no relabeling.
- PASS fresh exact-head packed installed SSE:1 test,15,189.106791ms idle,2 heartbeat comments,2 change-group frames,2,970 wire bytes.
- PASS fresh exact-head packed finite Transactional/Snapshot:30 tests in2 files.
- PASS fresh exact-head packed installed RU:151 tests in4 files;82.78s. All seven executor stages exit0.

Packed source/install/logs are under `packed-b14fb287/`; `packed-evidence.py` is an external copy of the prior executor with corrected checkout/log paths. Build/pack/install/test stages have explicit exit codes and bounded subprocess deadlines. `sse-observation.json` and `sse-response.http` retain actual wire evidence. SSE wire SHA256 `d47607165fc2385b3f0a39e978bf0bc636cbd43dc8797a7c7c49388b2e895ebf`; installed entry SHA256 `bcf153170f953f9e28c6c89f1153be907f64df289e64f3ad929cad4a4a54c47c`.

Actual payload digest `8be5c8a42240841a85b2233620cd5ee1c1c660f10fa03ae92f895be9063535be`; runtime configuration identity `952c08b6189abb4836c2e8d9c665062e33b1e276ea39e30f2a4e4ede38198875`. The configuration digest includes fresh credential associations, so it is observation-specific. `claimEligible` remains false.

## Remaining boundaries

Event Source SSE, erasure/mutable-view fences, set HTTP admission, remaining discovery/admission and deferred start-now reconnect combination remain separate obligations. This repair does not widen G2a. Root owns independent closure review and integration/publication.


## Final cleanup and freeze

The exact source remains clean after all packed tests. No owned test/server/review
children remain; no run timed out or was interrupted. No hook was bypassed.
`final-processes.json` records the final process inspection. Parent reviews and
unrelated services were untouched.

Tarball hashes: protocol `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`;
server `50a947387e97fd99c250f1fcd1fe3e514dd719b99f5530a3c366361d1f979db4`.
Source and all exact pack identities are also recorded in `packed-b14fb287/source.json`.
