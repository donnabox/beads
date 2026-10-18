# Command minimum actual package and installed evidence audit

No blocking discrepancy found in the completed artifacts.

This is an actual-evidence audit by the author of the execution preparation, not an additional independent source-code council review. I read the frozen source and retained logs/results, compared archive/source/installed bytes, independently recomputed the dependency payload, and checked actual process/directory absence. I did not import product code, invoke a prepared script, or run tests, builds, installs, services or packaging.

## Exact source and package attribution

Clean ordinary commit `f2ac3f980fc8d35a3b485db04b813ad1b569d94d`, tree `74327f052cef6026b3ea10e186523ace9c87a128`, has the expected parent `122800511ee6306e4105e1f73bf3658f35c90a14`. Git reports exactly the seven authorized changed paths; all seven current bytes match the final-draft manifest ff9c6341. The entry is the actual consumer-resolved transaction-installed-process.js, SHA `bcf153170f953f9e28c6c89f1153be907f64df289e64f3ad929cad4a4a54c47c`.

Protocol archive `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227` contains 63 regular members; server archive `ecfe5b6997d6c653b42070addbbd97e9034d9cdf521a45fbdce0fe163aed28e4` contains 250. All 313 members match their installed counterparts and current built-source bytes, with only the expected package.json publish normalization of server's protocol dependency from workspace:* to 0.0.0. Inventories include the complete published dist/schema/LICENSE/package.json sets, with no missing or extra installed members.

The actual installed graph is exactly eight packages: protocol/server 0.0.0, ajv 8.17.1, ajv-formats 3.0.1, fast-deep-equal 3.1.3, fast-uri 3.1.5, json-schema-traverse 1.0.0 and require-from-string 2.0.2. I resolved each recorded dependency edge against consumer paths and compared actual manifest hashes/versions. Source lock, consumer lock, consumer manifest/settings and entry hashes match provenance. An independent byte traversal following installed-payload.ts's path-sensitive rules gives `53d3e65061c706d51b666d16e770db7bb0a8147decd15383d42851a16e0f5c68`, with 14 recursive package visits and 473 runtime-file visits. This equals both retained provenance and the actual served binding. Entry identity alone is not used as whole-payload proof.

## Actual run and oracles

All ten final package/consumer/verification/TX stages have actual exit0/supervisorExit0, unchanged source, no timeout/forced cleanup/inventory error or remaining owned runtime roots. The TX command took 3.138 seconds under the unchanged 120-second supervisor ceiling. Results select exactly one passing case: “ignores command read controls while enforcing a later installed receipt minimum”.

The genuine helper JSON parsed from fresh-tx.log equals the postprocessor's retained observation. It records 18 completed responses: 8 POST, 9 GET and 1 HEAD, across exactly nine URLs, with every ordered route/method/status matching the frozen case. The strict observed receipt-root route is canonical and within Scope. Maximum retained response-body length is 803 bytes; aggregate 7095 bytes. Replay body and wire fingerprints match for singleton, batch and sequence; the unchanged resource and later compatible receipt body fingerprints match their earlier values.

The inspected case asserts receipt/sequence schemas, exact idempotency keys/effect/required positions, resource revisions and representations, preserved replay/key/revision behavior, current context, receipt future-minimum503 and foreign-view409, compatible receipt200, HEAD503 with empty body, and exact final three-item graph. The source's final call and URL cardinality assertions agree with the retained observations. The log contains fingerprints and safe bindings rather than full response bodies/headers; full assertion values are evidenced by the selected passing source case, not independently reconstructed from hashes.

Configuration is the private canonical `https://reference.invalid/acme/` Transactional fixture with history false and claimEligible false. Actual source declaration/installed entry/payload match the launch metadata and provenance. The retained operation target inventory and runtime/artifact/policy/authentication/fixture identities come from the real installed child through installedTransactionBinding, whose source validates the loopback endpoint, authority/profile/history and exact target inventory. Authentication includes ephemeral random credentials: this audit does not invent or reconstruct missing raw credentials/configuration, and does not upgrade digest presence to independent attestation of deployment identity.

## Cleanup and claim limits

The helper reports one launched child PID99866, exit0, closed true, aborted false. Independent current process inventory confirms PIDs99835/99839/99847/99866 and process group99835 absent. The sampled runtime root `bdp-collection-minimum-installed-xf4Mrb` is absent; no matching runtime roots remain under the actual attempt TMPDIR. Only node-compile-cache remains. These are actual absence checks, separate from helper closed/exit fields and sampling limitations.

The earlier v3 restore-command control evidence and failed v3 receipt build remain historical, separate from corrected v4 receipt evidence and this fresh installed run. This audit does not relabel the failed build as an executed sensitivity test, rerun controls, or replace root's combined control acceptance. No public Scope discovery, complete profile qualification, pending minimum/Event decisions or broader installed coverage is claimed.

Audit output is retained in native-actual-evidence-audit.json. Input anchors: tar provenance `a078a81a944442e7cc76c5db768c68cb4c01615a669acaef610cabab254caccb`; installed provenance `c8b26871b4ddfc899065eae5553965892434e30af6d11f911837a7585093d0b9`; TX result `b4f675121ff757af5ab3417f534957d4450ecf2a701bde14c4c603edf438be7c`; TX log `c051a36c3f8ef1c46ed4cb9caf4340a027f53c9f8419554c6dae3f283adb378a`.
