# M1 runtime configuration: independent native code review

**Verdict: no actionable correctness or regression finding in the bounded M1 delta.** Ready for root's ordinary integration subject to the separate council review. This is not installed Transactional profile qualification.

## Pin and scope

Reviewed exact frozen `b2d6eefd263bde18ee2430fbfb4b901a2d866cf6` against `31aab4db0f61e3d8b1139a7b78ff1ff10f890836`. Author checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-runtime-configuration` remained untouched and clean. Six-file diff agrees with author report: private capture module/tests, runtime wiring, existing runtime test assertion, additive owner constant export and private package inventory/denials. Read author report, approved plan and A1–A9 dispositions.

Independent validation used a separate clean clone, `/Users/dbox/repos/wt/vickie/vickie-runtime-config-native-review`, at the same exact head. No source edits, commits or outward actions. Only external review logs/report were written, plus ignored build/dependency outputs in the reviewer clone.

## Findings

None requiring a code fix. In particular, I did not find a bypass allowing arbitrary test behavior to acquire default runtime metadata, a new authority-routing escape, a principal/permission association collapse, or a change to persisted store compatibility/retention semantics.

## Load-bearing checks

- `transaction-runtime-configuration.ts:31–78,128–175,182–267`: own-data validation rejects proxies/accessors/unknowns; administrative values and nested seam records are synchronously detached. A supplied undefined seam key still suppresses metadata. Functions are confined to the intentionally ineligible seam path. Token buffers remain closed over by authentication rather than exposed on the returned metadata. New credential/principal/directory/Type input restrictions are explicitly documented as private administrative bounds.
- `:204–231,275–285,325–341`: authentication consumes the same captured principal/canMutate values represented by sorted token-digest associations. Both authentication and policy identities retain per-token permission; credentials with the same principal are not collapsed. Actual permission-swap HTTP checks exercise allow/refuse behavior. Sorting makes input ordering irrelevant without altering dispatch semantics. Policy/authentication digests are intentionally coupled under the fixed view.
- `:235–265,307–342`: exact artifact capture passes through the existing compiler; the same branded closure goes to the owner. The existing owner still verifies compiled descriptor agreement before creating the database. Metadata checks actual descriptor closure/view/empty aggregate rather than forwarding compatibility JSON as a hash. Its runtimeConfigurationIdentity is distinctly named and contains actual captured settings, contracts, mode, policy and authentication. Existing service construction checks receipt maximum agreement; omission of that duplicate check in metadata is intentional and safe on the runtime path.
- `:85–125,287–304` and `transaction-runtime.ts:165–190,212–232`: selector, pagination, transport, maintenance and wait settings are used by their real consumers. Imported mutation limits/target limit remain their original shared sources. The receipt-detail alias in `transaction-store.ts:159` exports the actual existing day value and does not change its uses, format or compatibility preimage. Clock samples/random tokens are correctly excluded from logical configuration identity. The pure leaf-hash test alone would not prove consumer wiring; the wiring and actual HTTP cases were inspected separately.
- `transaction-runtime.ts:103–118`: canonical Host mode is explicit; default listener mode is preserved. Physical loopback address and canonical Scope remain distinct. Wrong/foreign/default-port Host variants fail before admission, while canonical Host succeeds at the actual port-zero listener. Existing raw target/method/authentication/Origin ordering remains unchanged. The loopback HTTPS-identity mapping is private transport configuration, not a TLS or full inherited Read claim.
- `transaction-runtime.ts:229–256`: metadata is prepared from the successfully opened owner, but no runtime object is returned until handler construction also succeeds. The existing handler gate keeps requests unavailable before setup; startup failures close listener and owner/service using the prior aggregate cleanup paths. No new owner or Read plane is created. Normal close retains its memoized drain behavior.
- The added package artifacts are explicitly inventoried and both package-import spellings are denied. No public export map, launcher switch, admission guard or Read evidence artifact changed. The adjacent reopen test removes an invalid assumption about OS port allocation; the new simultaneous-listener test provides a real distinct-address witness instead.

## Independent executable evidence

Pinned Node24.16.0 and pnpm11.20.0. External logs: `/Users/dbox/repos/wt/vickie/bdp-integration-logs/runtime-config-native/`.

| Check | Result |
| --- | --- |
| Frozen dependency install | exit0 |
| Clean project build before emitted-runtime tests | exit0 |
| `transaction-runtime-configuration.test.ts`, `transaction-runtime.test.ts`, `transaction-http.test.ts`, maxWorkers2, verbose | **60 passed, three files, zero failures/skips**, 2.11s Vitest |
| Typecheck | exit0 |
| Dependency boundaries | exit0 |
| Packed executable smoke, including new private-module import denials and unchanged public admission boundary | exit0 |

Each command has its own log and explicit `.exit` record. These are independently rerun existing tests, not newly authored probes. The runtime tests exercise actual emitted code and abrupt-process/reopen paths; new configuration tests exercise raw sockets with independent dial/Host/target, input mutation, real per-token permissions, canonical/legacy modes and metadata suppression. Packed smoke's installed mutable observation is the existing **Read+Update** factory; it does not count as a fresh packed Transactional cohort.

No full repository suite, separate development smoke or new Transactional installed cohort was run in this review. Author's255-test/both-smoke report remains separately attributed. No timeout, stopped run or forced cleanup occurred. Reviewer and author source trees were clean afterward; no reviewer test child remained in the final process check.

## Disclosed boundaries, not findings against this slice

H1b still must bind a fresh packed TX factory/process, actual payload and physical endpoint, require this metadata, admit the private exact Transactional executor and record dynamically captured receipt routes without weakening route/credential confinement. M1 does not implement before-first-request binding enforcement. Snapshot settings are absent at this source and must be included when Snapshot capability is integrated; this does not invalidate the current fixed configuration. Retention is bound but unadvertised/unqualified. Existing private Origin/absolute-target refusals, missing public Transactional discovery and the remaining required profile evidence are not cured or newly claimed here. Metadata remains claimEligible:false.
