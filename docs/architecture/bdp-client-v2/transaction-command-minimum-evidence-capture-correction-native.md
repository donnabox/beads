# Native capture-correction disposition

Recommend the unchanged **11-action / five-file** correction. An early twelfth GET is unnecessary for complete original-A/retained-A receipt-body equality: the existing private receipt authority already performs that exact comparison internally. Do not claim that client output or response bodies can be recovered from the returned report.

Read-only basis: runner.ts and transactional-observation.ts in `/Users/dbox/repos/wt/vickie/vickie-command-minimum-evidence`, ordinary parent453b6dab/tree40782d97. The author-owned manifest/coverage are an unfrozen draft and are not approved by this report. No source changes, imports, tests, build, services or new review process were performed.

## Actual interface limits

1. `packages/conformance/src/runner.ts:2468–2496` captureValue requires an absolute HTTP(S) string confined to scope. Raw response object, revision string, epoch/view string captures fail. Keep only legitimate URL captures, principally A `/id` for receipt authority; ordinary URL captures remain ordinary and do not imply authority. Remove every dangling equalsBinding that relied on rejected raw capture forms.
2. `runner.ts:189–198,1099–1101` exposes only programmable action output `{kind:"json"}`. `:1374–1383` reports response metadata/body kind/length, not bodyText; `:3306–3316` redacts/filters headers. `:1790–1810` also redacts raw header values and elides nonallowlisted headers. Therefore postrun comparisons cannot use report values as an unredacted client body, epoch/view or empty-header oracle. `/body` is available **inside** the private client action's manifest assertions, not in ObservedAction output.
3. `runner.ts:1126–1133` permits programmatic JSON captures after assertions, but capturing action9 does not make its original raw A/B/C bodies available or expose captures to the app. No new capture-return seam is needed here.

## Existing complete retained-root comparison

`transactional-observation.ts:164–245` grants authority only for supported mutation POSTs, pointer `/id` or `/next`, actual200/raw wire, schema-valid receipt, matching wire/body and confined receipt identity. At registration it saves `body: JSON.parse(response.bodyText)` privately with source action, pointer, receipt, credential, source wire digest and session identity.

For A's `/id` root, `:249–273` resolves that exact binding/target/session; `:274–305` validates real receipt wire and requires `isDeepStrictEqual(body, entry.body)` on GET, otherwise throws `retained receipt differs from original wire facts`. This compares the **entire parsed JSON body**, including extra fields, not only selected known facts. It is not a byte-for-byte serialization comparison. The `/next` first-page baseline has different behavior, so pin source pointer `/id` and never substitute page authority.

`runner.ts:945–980` invokes this validate callback before projection completion, then independently validates receipt schema/id. A mismatch throws before a successful client action. Keep action9's known body/schema/root/key/position/header assertions; replace its impossible raw-capture `retained-body` equalsBinding with an honest body-presence/known-fact check as appropriate, with full retention explicitly attributed to this existing private validator. Postrun require:

- action9 completed with output kind json, no executionError and every client assertion passing;
- its actual exchange passed actual-status/actual-request/actual-body with complete write state;
- sourceReceiptObservation.action exactly m-create-a, pointer `/id`, receipt equal the genuine A root, original source wire digest equal the raw A hook's digest, and expected credential/session/configuration binding;
- dynamic route is the exact observed A root GET, correct actual served session, and the entire new row passes.

The raw A hook may retain a safe root and original response-wire digest for these comparisons. Together these facts prove original-A→retained-A parsed-body equality without recovering the client body. Do not attribute this validator behavior to a newly authored manifest equality assertion, and do not count it as a newly added assertion ID.

## Raw replay, state and revision comparisons

Use the already permitted app afterExchange closure's actual response.bodyText/headers/exactRequest/wireResponseBytes for the ten raw actions. Keep exact row-specific activation, ordered route/body/key matching, duplicate/extra-traffic checks and original header sets. Independently validate complete schemas and literal semantic facts **before retaining** each baseline receipt/Resource/sequence result/revision. The outer hook runs before the runner's manifest assertions (app wrapper awaits hook before returning response), so merely assuming those later assertions have passed is invalid. Require their actual final pass too.

Compare A replay to original A full parsed receipt, B replay to original B full receipt, C replay to original C full sequence JSON, final collection items to the validated original A/B/C Resource objects, and Event subjects/revisions/ordinals to independently validated creates. Pin exact item/Event cardinalities and no stale-update effect. In negative control mode validate the actual intended409/403 shape and retain evidence without throwing the success expectation inside the hook; the runner must reach and report the named failing manifest assertions. Success-path full-object replay equality is separately an app assertion, not returned runner credit.

## Context and proof-boundary corrections

Raw-hook context values are visible and can be compared across all ten raw actions. Dynamic action9 has real unredacted headers only within its client projection assertions: retain epoch/view presence and literal current position pos-3, root body required/effect pos-1 and ordinary read response facts. The existing receipt validator proves body retention and logical session/configuration continuity; it does **not** compare dynamic response epoch/view values to raw A's response header values. Do not say raw-to-client epoch/view header equality is proved, or infer values from redacted report fields. Record that specific comparison as outside this bounded row if the earlier broad context-equality wording implied it. No extra GET alone is necessary for the accepted core retention witness.

Coverage bindings must mirror only real manifest assertion IDs. Full raw replay/resource/revision/state comparisons performed in the app are supplemental installed-test evidence; they must not be presented as runner assertion outcomes or synthetic report fields. Static guard/weakening descriptions must distinguish the manifest portion from app-owned equality checks: removing replay equality must be rejected by the actual app/independent source-ledger guard, not an equalsBinding that cannot execute. Keep wholeContract:false and existing missing partitions. The row's source-bound report must separately identify manifest checks, private receipt validator proof and app-hook checks. This is a truthful correction of evidence attribution, not permission to weaken complete equality in the app.

## Counts and recommendation

Retain11actions, positive203HTTP/22children, selected batch4/2, selected readonly1/2:208expected/225maximumHTTP and26children. CI increments remain454expected/473maximumHTTP,50maximumchildren; global invalid receipt mode still fails new action9, not2. The proposed12action arithmetic is internally plausible but unnecessary and should not be adopted. No runner/schema/API change, added request, export, capture interception or extra test registration. Root should record this source-backed plan correction, then review the actual five-file implementation/measurement changes before execution.

The deferred transaction-sequence-member-plan.md also assumes raw full-body/revision captures and must explicitly recheck this restriction on resumption; it remains deferred and was not re-audited here.
