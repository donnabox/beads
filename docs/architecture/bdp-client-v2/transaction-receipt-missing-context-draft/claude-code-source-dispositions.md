# Receipt missing-context: Claude source dispositions

One concrete required test correction was found under F1. F2–F8 do not establish another mandatory source change. This is a read-only source audit, not a new test run, approval of installed evidence, or a replacement for exact-head full CI.

Reviewed complete Claude stdout SHA256 96959814652b5360d014a6f67b162a1ff774aec8730cefe0da642371ddafb094 against the eight-file draft source binding edc5c98c8e5c82f2073e1bf35aa84098efe9552c7dbc74537784de75788cca5a, full patch 837bacdc192bcf0401171ecc9a8312ae034beabed7b69620b6a3beaef0cb9de7, base 015cac1aa4fa827cc4a2c647dff08454c6c8e6f8. Anchors below refer to that frozen workspace. No product edit or execution was performed in this audit.

## F1 — required: reconcile one remaining stale header assertion

At packages/server/src/transaction-http.test.ts:1762–1766, the existing callback “uses actual epoch/view decisions and an exhausted wait before conditional or timeout responses” reads a receipt as otherToken with foreign epoch/ahead minimum, correctly expects404, then incorrectly expects absence of bdp-scope-position. Current transaction-service.ts:206–220 and495–511 forward the genuine missing observation; transaction-http.ts:540–551 carries it on404. Nondisclosure precedes semantic minimum refusal. This fixture has only its initial commit, so the header must be pos-1.

Smallest correction: replace the absence assertion with exact header value pos-1 (a complete tuple comparison is optional, not needed to repair this stale assertion). Select this exact existing callback once after correction. Its source-derived workload is14HTTP: one seed POST, two targets each with one timeout and two epoch/view refusals (six), one hidden receipt, five unsupported routes and one malformed-position request. One fixture owns one owner, listener and temporary store; zero installed product children. Existing fixture cleanup applies. No new timing allowance is justified.

The original14 selected callbacks and typecheck/build are real but do not cover this callback. A repository search of service result producers, receipt/page calls, resource-not-found objects and literal/looped false context-header assertions found this concrete stale expectation. TransactionServiceReceiptResult is defined and produced in transaction-service.ts only; the old service hidden-object assertion at service.test.ts:237–242 is already reconciled. Other examined false-header assertions concern405/401, unsupported visibility, aliases or excluded sequence paths (HTTP1073/1089/1907/1920/2994/3026/3118/3123; service2020; SSE340). Page namespace defense equality at receipt-pages.test.ts:281–288 remains intentional. This bounded inventory is not a claim of full regression coverage. Exact-head full CI including downstream gates remains required; running three entire files additionally is not necessary merely to duplicate that gate once the concrete omission is repaired and its callback verified.

## F2 — preserve historical checkpoint

Progress document lines1865–1882 explicitly describe “This draft” and “At this draft checkpoint”. That is the authored pre-execution checkpoint, not a live statement that later external validation never happened. Replacing it with selected positives while control/installed are moving would relabel historical evidence and become stale again. Preserve it; current run/source/control/installed records retain their own source/time attribution. An eventual separately scoped current-status update is not this code correction.

## F3 — already bound before dispatch

apps/bdptest/test-support/transactional-execution-binding.ts:50–73 builds all12 exact target names, including updateWhere and createBead, and deep-compares ready.targets against the entire operationTargets object before returning. The new installed test calls installedTransactionBinding before constructing or dispatching routes (installed test:155 onward) and checks ready.scope against the fixed fixture scope. Thus the two literal operation URLs are already validated against emitted metadata; no guessed route can reach the first dispatch without the existing inventory check passing. Additional two assertions or deriving URLs would be redundant.

## F4 — excerpt gap, HEAD handling exists

packages/conformance/src/http-executor.ts:725 adds Connection: close for exact requests. isFramedResponseComplete at1091 deliberately does not infer HEAD completion from advertised GET content-length; socket completion leads to parseRawResponse at1191, which treats actual HEAD wire octets directly rather than decoding the advertised length. This also preserves illegal body octets for an oracle to reject. http-executor.test.ts:294 explicitly tests HEAD with Content-Length123 and no body, expecting empty text/zero octets. This audit does not rerun that test or convert its existence into fresh execution proof. Prior Event installed HEAD evidence is separate historical evidence. No throwaway exchange or extra installed run is required to resolve the source question.

## F5 — premise is false for /pages

transaction-http.ts:526 chooses receiptPage only if the suffix contains “/pages/”. A terminal “/pages” has no such substring and takes the root read at540–551, yielding current-context404. It cannot exercise the private marker<0 page defense at store2034–2035. The outside-namespace page guard at2038 is likewise a trusted private API defense; ordinary receipt routing supplies the Scope receipt namespace. Pinning headers on the existing terminal-/pages request is optional specificity, not coverage of a wire-visible contextless page defense.

## F6 — optional table style

The existing tables are finite and explicit: HTTP330–344 maps anonymous401, foreign404, owner410;2037–2045 maps owner500 versus foreign404;2050–2058 maps401/404;2158–2176 maps owner500/foreign404. Their status/principal predicates are correct for those exact rows, and status is independently asserted before presence. Expanding future rows warrants revisiting expectations, but there is no current silent mismatch in these four tables. Explicit expected-presence columns are optional cleanup and not required for this repair.

## F7 — keep distinct error channels; actual result adjudication exists

HTTP4053–4086 names the field primaryOrCleanupFailureCount, not total assertion failures. Soft tuples intentionally remain Vitest failures outside that local primary/cleanup array. tupleChecks explicitly counts attempts. The completed control has eight named errors in genuine Vitest JSON, all eight missing tuple values in the safe13-response record, and primaryOrCleanupFailureCount0 with both cleanup callbacks completed. Those are consistent statements. P/actual-audit/audit.json and root-control-acceptance.json retain the separate evidence, including the failed original wrapper graph postcheck; they do not infer a pass from the local summary. Adding matched/mismatched counters would be optional duplicated adjudication. Preserve actual reporter errors as the control verdict authority.

## F8 — impossible collision under current source

The new source test constructs the sibling literal “unknown” at HTTP3992. The actual owner mints admitted receipt IDs as scope + receipts/ + randomUUID() at transaction-store.ts:1106; “unknown” is not a UUID output. No configured mint seam or imported receipt can supply that literal in this fresh fixture. An explicit inequality is optional defense against a future mint policy change; there is no current random-collision probability with this literal.

## Verified-note boundaries

The reviewed same-entry context returns occur after existing integrity checks and before minimum checks; private namespace defenses remain unchanged. Context does not disclose receipt membership: missing/foreign uniform Problems and current authorized Scope context stay aligned. The metadata error/fence distinction,32-inline/one-tail installed fixture arithmetic and declared route ordering remain source-grounded. Source13-response/eight-tuple control is actual observed evidence; installed14 remains separately gated and cannot be claimed from source or static arithmetic. The required stale assertion repair changes no production behavior, fixture operation, runtime limit or public eligibility.
