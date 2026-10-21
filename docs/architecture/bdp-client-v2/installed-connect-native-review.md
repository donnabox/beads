# Native review: installed CONNECT correction

Reviewed frozen `b50c5e8d91336b087966cd3374de7e72513d6906` against `d779ea1a80f2d606c0db3e3981608ac2bd2dc274` in `vickie-installed-profiles-integration`. Three files, 27 insertions/3 deletions. No tracked source modifications; clean at entry and exit.

## Conclusion

No actionable finding in this bounded correction. All three findings in `installed-marker-complete-claude-review.txt` are closed. This is not a cumulative profile qualification or broader transport audit.

1. **Detached CONNECT errors and lifetime:** `packages/server/src/development-runtime.ts:488–494` installs an error listener before calling `socket.end`, then destroys the detached socket in the completion callback. The listener prevents unhandled socket error termination; the callback avoids depending on the client's FIN or `server.closeAllConnections` (which does not own upgraded/detached CONNECT sockets). This matches the established public Read transport at `packages/server/src/index.ts:530–535`. Existing timeouts/normal request ownership are unchanged.
2. **Exact refusal values:** `scripts/installed-reference-observation.mjs:320–321` checks exact `Allow: GET, HEAD` and `X-Bdp-Development: non-attesting` bytes. The literal header-name casing is stricter than HTTP requires but appropriate to this exact runtime-produced response; lowercase method tokens or changed marker capitalization now fail. The status, empty-body delimiter and zero length checks remain.
3. **Scope precision:** `docs/design/update-implementation-progress.md:589–603` now explicitly enumerates the union across ordinary and installed configurations and bounds reset-peer observations. It does not claim every kernel write-error timing was injected. Node-generated errors remain outside the enumeration.

## Executable observations

Pinned Node 24.16.0. Existing emitted bytes contain the exact reviewed listener/error/destroy lines; no rebuild was performed in the frozen checkout.

- PASS: installed-reference `cold-navigates` test: **1 passed, 12 filtered/skipped**, 958 ms total; logs `installed-connect-native-tests.log` / `.json`. This executes the updated bounded raw-socket refusal, four reset peers, subsequent real discovery, normal close, empty child stderr, and reopen observations. The reset operations are genuine socket activity; no deterministic ECONNRESET branch coverage is inferred.
- PASS: separate external `installed-connect-native-probe.mjs`: deliberately holds the client writable side open (`allowHalfOpen: true`, no client FIN), receives exact bodyless marked 405, independently GETs discovery 200, then observes clean child termination and empty stderr while that client side is still open. Output `installed-connect-native-probe.log`. This supplements the committed test, which sends its CONNECT request with `socket.end`.

The same observation helper is called by packed smoke at `scripts/smoke-executables.mjs:986`. Root's 33-test run and packed smokes are not represented as independently rerun here. No full suite, pack rebuild or fresh evidence generation was needed for this bounded review. Probe temporary store removed, owned child exited; frozen repository stayed clean.
