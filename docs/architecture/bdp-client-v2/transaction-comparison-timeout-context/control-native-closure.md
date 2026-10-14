# Comparison preflight N1 correction: native static closure

**Not yet closed for execution: the control now reaches the intended boundary, but exact optional-property typing introduces a derivative-build blocker; the same check also exposes a production callback-option typing blocker.** No test/compiler/build was run. Root should hold execution for the two narrow corrections below and regenerated hashes.

## Verified correction and unchanged scope

Read control-correction.json and full original/current preflight and generator deltas. Verified initial-preflight preserves every recorded original hash, including the original source.diff and all five patches. All eight current source hashes match the prior bounds; live git diff remains byte-identical to source.diff27fc1610da057443639bbd49012e73ec2e9f951f4fdfc20d8801435922809cbe. The only bounds delta is the leak-unobserved-fields control path and patch/mutant hashes. Selection18, source16HTTP/two listeners, controls expected12/max36HTTP, installed27/two and all body/deadline counts remain unchanged. The other four control patches are byte-identical.

Reviewed successor preflight SHA aeefef2b264571a162b24809192734939e9b7b36399c4b95da4e93e76c904785; bounds SHA68afc70753899701e4d570a65b4db5f527c1918f114bc084c40ae3c9c3b8d595. New control patch SHA829165a4c17bcb58aff2492f6510da6726ebf9c7264b2316e20bf6f755b39fbe; independently reconstructed mutant SHA4c852cc5ce218cf561244f7c0078f6b7f2d81d7ca868de6f95cd7545a6de792c.

The mutant now changes transaction-submission.ts's legacy temporary-unavailable return. Its enumerable context property would survive to the selected submission test's `toStrictEqual` at transaction-submission.test.ts:1124–1135 and fail `legacy comparison result exact`. Prose accurately limits this to submission-boundary sensitivity; it does not claim strict undefined-field detection by direct comparison callers. No new test/cohort is needed for this correction.

## Remaining N1 blocker — derivative typing before the intended assertion

`tsconfig.base.json:6` enables exactOptionalPropertyTypes and noEmitOnError is enabled. The submission refused union declares `context?: TransactionReadContext`, which allows absence but not an explicitly undefined value. Adding `context: undefined` to that typed return therefore makes the proposed derivative build fail before the named runtime assertion. This cannot be accepted as the required sensitivity failure.

Smallest control-only fix: write `context: undefined as unknown as TransactionReadContext` in the isolated mutant. TransactionReadContext is already imported in this file. The cast preserves an actually enumerable undefined runtime property, allows the intentionally invalid runtime result through static checking, and makes the existing strict assertion the selected oracle. Update generator/patch/mutant/bounds/preflight identities; do not change production context typing to admit invalid deadline tuples. Actual successful derivative build and named assertion failure remain later execution gates, not established by this review.

## Additional Medium production typing blocker

`transaction-submission.ts:127–133` always supplies `observeDeadline: observed ? store.observeComparisonDeadline : undefined`. `transaction-comparison.ts:35–36` declares the optional property as the owner method type without `| undefined`. With exactOptionalPropertyTypes, the former value is not assignable to the latter optional property when explicitly present. The prior general preflight missed this static mismatch; this closure corrects that assessment rather than implying compilation was checked.

Smallest safe production fix: permit explicit undefined on this **source-private comparison option only**, e.g. `readonly observeDeadline?: ReturnType<typeof openTransactionStore>["observeComparisonDeadline"] | undefined`. Keep submission's inherited `Omit<..., "receiving" | "observeDeadline">` unchanged, so neither service nor runtime configuration gains this callback. Keep actual submission's final override after spreading comparison so an untyped caller callback remains discarded. An alternative conditional construction must first remove the untyped field; merely conditionally spreading the selected method after the caller spread would leave an injection path on unobserved calls.

This requires a narrow source hash/diff update but no new semantic scope, request, test selection or owned file. The method remains observed-only, full-owner unobserved callers retain their bare shapes, and the final-epoch control still references expectedEpoch and has no analogous unused-parameter issue.

## Review limits

No tests, typecheck, builds, installs, services, network or source changes occurred. Only read-only source/artifact comparisons, in-memory patch reconstruction and this report write were performed. Event substitution, reporting/cleanup behavior and other findings from native-preflight remain unchanged. Preserve this intermediate correction record; obtain narrow successor closure before root releases enumerated execution.
