# Native review: private point/receipt minimum reads

Frozen `d3252087a31529cf97b5bfcfe2c402f02c694b91` versus parent `7bfba525bbb66e64647110ac336f76b164b41942`. Independent detached checkout `/Users/dbox/repos/wt/vickie/bdp-transaction-point-minimum-native-review`. Read author report, next-slice plan and external14-findings dispositions. Source read-only; no outward/evidence changes.

## Verdict

No blocking implementation defect found in the approved visible canonical Bead/Link Resource/properties and receipt-instance complete-tuple subset. One documentation accuracy limit is confirmed below and accepted by root as clarification, not a requirement to add transport cancellation plumbing. Missing/hidden Resource ordering is the explicit private visibility gap, not a newly inferred normative rule.

## Low — distinguish facet cancellation from socket disconnect

`transaction-http.ts:167` calls `facet.perform(parsed)` without request.signal; `transaction-runtime.ts:149` creates its Request without a disconnect-linked AbortController. Consequently **both point and receipt HTTP waits** drain the remaining finite minimum budget after the socket disconnects. The service point facet separately accepts and checks an explicitly supplied AbortSignal. The author report singles out receipt disconnect, which can imply stronger point-HTTP cancellation than exists.

Independent actual socket probe starts a real minimum read, parks its first wait through the approved deterministic seam, destroys the client request, and begins runtime.close(). Close remains pending while that wait is parked. After release, two bounded wait turns exhaust the original3ms simulated budget and close completes, for both point and receipt. This observes real socket/owner/lifecycle behavior with a controlled clock/wait seam; it is not production wall-clock timing evidence. Exact probe: `transaction-point-minimum-native-disconnect.test.ts`; log/exit alongside. Both cases pass.

Minimal fix: clarify report/design prose that facet AbortSignal cancellation is supported, HTTP disconnect propagation is not wired, and both kinds may delay shutdown by the remaining finite budget. Root explicitly accepted this disposition. No normative immediate disconnect-cancellation obligation was located for this private slice.

## Confirmed behavior and boundaries

- Tuple capture rejects proxies/accessors, partial/duplicate/malformed/unknown-position forms, freezes primitive epoch/view/minimum data, and compares supported safe-integer pos-N tokens only inside the authority. Clients are not taught lexical ordering. Unknown opaque/partial forms remain named private gaps.
- Visible point Resource and serving context are captured in the existing same exclusive synchronous owner entry. Condition hook executes after authorization/materialization and before delivery. It adds no SQLite BEGIN to point reads. No public Read protocol concept/export is added; ordinary calls omit the private data slot.
- Receipt retention, principal/epoch nondisclosure, serving context and minimum check share the actual existing owner transaction. Each retry performs a fresh lookup; disappeared/expired failed receipts cannot leak an earlier cached body.401/404 withhold context; typed409/410/503 carry actual final-attempt context.
- Foreign epoch/view precede minimum comparison, then actual behind observations alone trigger waits. Finite monotonic original deadline is not renewed; default waits are bounded5ms sleeps outside the owner. Real concurrent mutations satisfy point/receipt waits. Ordinary owner errors are rethrown unchanged by the retry helper, not converted to timeout.
- HTTP preclassification prevents ignored controls on aliases/bootstrap/discovery/collections/unsupported routes. Authenticated current visible points and receipts use the same representation path for GET/HEAD; HTTP conditional semantics apply after minimum selection. Canonical transactionalProblem tests validate409/410/503 bodies; HEAD suppresses final bodies while preserving context/Retry-After.
- Missing Resource with controls returns explicit bodyless minimum-position-visibility gap without context. This is intentionally not404-first or barrier-first qualification. Fixed all-visible owner is the only projection tested; dynamic policy/view rotation remains outside the claim.
- Service tracks asynchronous waits and drains them on close; accepted mutation drain behavior unchanged. Authentication/permission on unsupported command controls precedes gap response without key admission, with disposable receipt-table/no-effect witnesses.
- No store format/write identity, evaluator, schema/spec, public profile admission, sealed evidence or Type lifecycle change. Continuation anchor/minimum combinations, unknown tokens, malformed tuples, replication/snapshots/events and cold Transactional discovery remain explicit omissions.

## Independent verification

Pinned Node24.16.0 / pnpm11.20.0 frozen install and build pass before emitted tests. **227 tests/5 files pass,1.61s,maxWorkers2**: transaction-consistency, transaction-http, transaction-read, transaction-service, authority-read. Separately **2 native actual-socket disconnect probes pass** in429ms. External logs `transaction-point-minimum-native-{install,build,focused,disconnect}.log/.exit`, structured focused.json; probe source retained externally and temporary clone copy removed. No interrupted tests.

Broader498-test regression, checks/smokes and unchanged Read evidence remain author results rather than independently rerun. This bounded review establishes no full Transactional qualification, power-loss, changing-policy or production deployment claim. Author and independent checkouts clean at exactd325; root owns subsequent documentation correction and outward decisions.
