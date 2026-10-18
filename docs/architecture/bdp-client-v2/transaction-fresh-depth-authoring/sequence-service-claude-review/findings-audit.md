# Native dispositions — actual sequence-service Claude review

Actual review completed normally: exit0,535.189146s, no timeout, empty stderr and remaining process groups. stdout SHA256 `3245b6d7c1e7cf22ab50e3302f92b07355488cccbeaf799b3b9c18486ff22450`; exit record `2e762df268e3c2263342a0a0253e8a0bfc18e2fa994cf7fe1167dae14ba5e1a4`. Before/after source/input snapshots are identical. Independently checked sampled PIDs7044/9029 absent. This is a substantive completed review, unlike a timeout/empty result.

**No further mandatory code or test repair established.** The root-applied rationale comment is an appropriate narrow response to F1. Fresh packaged end-to-end validation remains required; no source-only result closes that gate.

1. **F1 — genuine limitation of the regression's discrimination, not an incorrect production traversal.** The test proves deep graph immutability and fails against the recursive implementation. It does not separately discriminate a hypothetical Object.isFrozen skip: fresh Resource descendants are already frozen upstream by evaluator.freeze (resource-evaluator.ts:411–421), and retained projected receipts by store.freezeReceiptValue (transaction-store.ts:85–95, receipt observation). The reviewer is correct to distinguish those facts. A third-file exported helper/unit seam is unnecessary scope growth for this private method's acyclic projected JSON. Retain complete traversal and document its layer-owned guarantee; do not claim an independently exercised skip mutant. Root's comment does precisely that.

2. **F2 — optional oracle cost/diagnostic refinement, no demonstrated timeout defect.** The per-node assertions do incur real work, but the measured focused run passed in1.733s and complete68-case service file in2.625s. There is no measured failure or timeout justifying an immediate oracle rewrite or increased bound. Keep the20-second callback cap and current independent complete comparison/freeze checks. A future efficient first-divergence oracle could be useful; naive accumulated full path strings on a12000-level chain risk quadratic allocation and are not automatically an improvement. The current failures retain labels and safe scalar/key diagnostics without recursively formatting the whole deep graph.

3. **F3 — missing-context claim incorrect; installed completion correctly still pending.** The committed transactional-fresh-depth-installed.test.ts already builds literal depth12000 at136 and sends the actual dependent-sequence request7 at724–746 with full200/context/result assertions. The preserved installed run actually reached that request and exposed the exact500, so the end-to-end regression exists. Do not add another handler request or deepen an unrelated40-level HTTP case merely to duplicate it. The required fresh packaged rerun must still pass; its prior failure is not coverage success.

   The serializer explanation in F3 is also incorrect. transaction-http imports jsonResponse from read-http-routing.ts; that function at44–52 only packages body/status/headers. transaction-runtime.ts imports stringifyJsonValue from @bdp/protocol at12 and uses it in deliver at91. Protocol json-values.ts:101–145 uses explicit tasks; native JSON.stringify is applied only to primitive values/individual keys. Storage similarly aliases transactionStorageText as json (store:67), which delegates to stringifyJsonValue (batch:1065–1066). The green source run is not evidence that native JSON.stringify accepts a12000-level graph.

## Exact post-review comment bridge

The reviewed production hash was `4faf6b2187cacd2c50fe8437819e60b4c216f3e8485fff3251f5ed9058f7645d`. Current hash is `b6f9584d67992ddd3a08304a6b76563318252c316b09afd991e49a3304d3af05`. Removing exactly the single added line:

`// Traverse frozen parents too: this observation guarantees every descendant is frozen.`

reproduces the reviewed hash byte for byte. Test hash remains `32010b6fb1107076534c544f2bbb93e62cc39c89cd0fbc3c482f0736cc28cfe2`. This is a verified comment-only bridge after terminal review, not an undisclosed alteration to active inputs. No execution or source edit by this reviewer; original review and runtime failures remain intact.
