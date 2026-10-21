# Native Codex repair review — PR64

Reviewed commit `0172c16fea5f29b51628d25d43f3b014cbe04a63` in `/Users/dbox/repos/wt/vickie/bdp-pr64-freeze-repair` against `b5a5b4153130dd91fb3ddb30f5aafc8f57311d2e`. All three changed files were inspected. The checkout remained clean and unchanged by this reviewer.

**Both findings in `native-review.md` are closed by this delta. No new defect was found in the three-file repair.**

1. **P2 closed:** Both metadata projection helpers now freeze the synthesized empty object before returning it. Existing supplied metadata continues through the existing deep snapshot. The new regression checks direct Bead and Link parsing, both History wrappers and both mutation-result wrappers, verifies failed `Reflect.set`, unchanged empty metadata and revision, and unmodified caller inputs. It checks the actual reported invariant rather than just matching implementation structure.
2. **P3 closed:** The advertised-limit paragraph now defines one combined count for `propertiesChange` plus `metadataChange`, and byte/depth bounds for each path. Applying patch limits to each set-update member is consistent with the existing one-patch rule and the separate transaction/member cardinality limits. This wording does not add `metadataChange` to `UpdateWhere`, change its existing `change` grammar, or aggregate patch count across selected Resources. No set-operation semantic expansion was identified.

Independent dynamic check: copied the exact repaired protocol sources outside the checkout, transformed them with Node 24.16.0, and loaded the pinned protocol dependencies. All seven tested paths (direct Bead/Link, historical Bead/Link, Bead/Link mutation results and the collection control) returned frozen metadata, rejected mutation and retained revision `r1`. Results are saved in `native-repair-probe/probe-result.json`. No test cache or output was written to the source checkout.

Inspected the author's existing logs: `freeze-red.log` shows the new regression failing on the old implementation; `freeze-green.log` shows 201 tests passing after repair. These are author executions, not independent suite reruns.

## Remaining verification gate

During this review, `fix-full-tests.log` completed with **7,583 passed, one skipped and one failed** across 108 files. The failure is `packages/server/src/installed-schema-graph.policy.test.ts:38`, the deliberate source-pin review tripwire: `read-values.ts` changed from SHA-256 `07a908343f550f9400ed659be4afad7b134c0684dc2973f68558ddbbdc32ec22` to `6ff2cd2456a5487ded968bac6651164ed4a82ad665b8896986accfd662688482`.

This is a remaining landing verification task, not evidence that either original repair is incorrect. The schema-graph policy record explicitly requires reviewed justification before refreshing a behavior-preserving source pin. Inspection supports retaining `schema-resource-index-2`: the repair changes only the mutability of synthesized Resource metadata; it does not change schema graph admission, URI/reference interpretation, descriptor validation, graph shape, budgets or refusal behavior. The installed graph implementation uses protocol error/descriptor/JSON helpers, not these Resource projection helpers. Update the pin with that explicit review justification, run the pin check and appropriate final checks, and preserve honest source/evidence qualification. Do not mark the full suite green from the earlier 201-test subset.

Full Read requalification/evidence handling remains root-owned. This addendum approves the bounded repair logic; it does not claim completion of those remaining gates or of later provenance/editorial changes.
