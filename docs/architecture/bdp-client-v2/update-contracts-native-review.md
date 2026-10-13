# Native review — U2 installed contracts and trusted principals

**No actionable findings or blockers identified in this bounded increment.**

Reviewed all six changed files in frozen `2c5fe9e04250ae036e681f46c60d3592e5f1cdf1` against its sole parent `3edba2140cd98878abfe36eb5bf8ab961eec708c`, read-only checkout `/Users/dbox/repos/wt/vickie/vickie-update-contracts`. This covers the full new compiler bridge, authority/runtime configuration, emitted-runtime HTTP tests, smoke inventory and implementation-progress note. It does not repeat or expand the prior U1 whole-delta review. Author tree remained clean.

## Review conclusions

- **Actual installed schema closure:** the bridge captures artifact bytes with the existing ingester, reconstructs compiler input only from captured bytes, uses the guarded offline compiler, and exposes contracts only for the exact installed serialized descriptor bytes. Artifact identity plus captured aggregate configuration determines installation/lineage identity; exact reopen works and changed closure or aggregate refuses. Caller byte mutation cannot replace the live compiled schema. Compilation refusal and incomplete evaluation remain local errors; the bridge does not invent complete standardized diagnostics.
- **Diagnostic contract:** a complete invalid evaluator result emits real keyword locations, instance pointers and bounded messages; the Resource evaluator attaches the failing effective Type and applies the shared diagnostic budget. The bridge checks completeness before any emission. Existing evaluator applicability/ceiling limitations remain explicit; this review does not qualify arbitrary schemas or resource-use guarantees.
- **Owned and aggregate behavior:** the actual HTTP observations exercise creation/update/deletion of owned Links, source revision/provenance fanout, source guard failures, no-op preservation, target aggregate refusal and source write permission. Rejections preserve source and Link state; endpoint revision pins remain carried data rather than being treated as optimistic guards.
- **Trust and policy:** credential mappings are captured as bounded administrative data and determine stable principals independently of client actor attribution. Read and mutation capture current immutable policy snapshots. Cursor view identity includes principal and policy snapshot; retained Resource disclosure rechecks current visibility without replacing retained outcomes or rerunning execution. Hidden owned Links withhold the complete source record and retained postimage. Principal key namespaces remain isolated.
- **Lifecycle and API:** the private fifth runtime argument cannot be passed through the public one-argument development wrapper. New contract/policy factories remain in a denied package subpath. The default runtime still uses its existing installation and non-attesting facade. Configured routes are principal-bound; startup failure cleans up the existing listener/owner path. U1's shared route and mutable owner are reused; no shared evaluator/store/schema/public-export implementation changes occur in this increment.

Relevant existing normative boundaries were checked against the candidate spec: authenticated principal versus actor (`:743`, `:804`), per-principal retained keys (`:3183–3184`), current disclosure authorization (`:3307` onward), and diagnostic Type attribution (`:3446–3474`). This slice does not establish installed Read+Update profile qualification or widen those obligations.

## Independent verification

A private exact-commit archive was built under `/Users/dbox/repos/wt/vickie/bdp-integration-logs/u2-numeric-native/u2`, using explicit Node24.16.0 and private workspace package builds with existing external dependencies. Build completed before emitted-runtime HTTP tests.

- PASS build (`u2-build.log`).
- PASS **152 tests / 4 files** (`u2-tests.log`): actual development contracts, U1 discovery, default development server and installed schema evaluator.
- PASS **10 tests / 1 file** (`u2-native-probes.log`): eight original contract observations plus two independent additions written only in the private archive. The eight repeats are not additive unique coverage.

The additional probes establish:

1. A real compiled schema with 1,025 failing `allOf` branches returns a bodyless HTTP 500 rather than an invented complete `validation-failed` result; no Resource is created. A corrected request satisfying all branches under the **same key** succeeds, proving the failed incomplete evaluation left no retained refusal or captured key preventing correction. This extends the author's direct validator completeness test through the actual emitted runtime, receiver and durable owner.
2. Bob cannot continue Alice's page even when both principals have equal access. Mutating the caller's original credential token/principal mapping after startup cannot change Alice's selected authority; the captured credential still performs the authorized write while Bob is denied.

The first version of the extra completeness probe used `not.toHaveProperty` on the intentionally absent body and hit a test-harness TypeError. The preserved `u2-probe-harness-error.log` is not counted as product evidence. The corrected probe asserts the actual empty response text and completes the same-key correction check successfully.

## Limits and integration notes

No independent full suite, packed install smoke, fresh conformance cohort, or combined U1/U2/T1 integration run was performed. Author-reported broader checks are separately attributable to `update-contracts/review.md`. No evidence artifacts/pins or public profile advertisement changed; historical Read evidence is not relabeled as U2 observations.

The separate numeric-contract supplement is not included in this U2 frozen head and has its own review report/finding. Shared numeric behavior should be integrated only with that supplement's corrected regression expectations. Root owns integration with the context-privacy repair and later TX service work. No author-source edits or outward actions were made.
