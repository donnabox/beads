# Cross-carrier review successor

The two test files now contain the root-authorized successor. Before application, all four original hashes matched binding.json. The exact staged guard-and-report-amendment.patch (83cf3edd53e4c584afe159ccb43f96c836e3bcbd7a56b080fa8070528bac5122) was applied before the four accepted review fixes. Original evidence remains preserved. Manifest and coverage hashes remain byte-identical, so all canonical and derivative measurements, 24 raw object pins, route digest and control oracle lists retain their measured basis.

Changes beyond the accepted guard/report amendment:

- Claude4: the existing assertCrossCarrierControls helper loads both transformed manifests through the cumulative bundle, resolves every derivative json-schema assertion using existing createJsonSchemaValidator(schema).resolve, requires each execution digest to differ from canonical and the two to differ from each other. Bundle admission alone is not represented as schema-reference resolution. No new test case, env var or output was added for this finding.
- Claude6: the existing context-presence-before-equality case includes the cross-carrier row; its title now names set and cross-carrier rows.
- Claude8: the positive cross-carrier row requires POST bodyOctets <=2048 and decoded response bytes <=8192, matching completed-conflict guards.
- Claude9: the completed hook forbids all three new operation route URLs as well as both carrier keys.

The previously staged amendment adds the optional BDP_TRANSACTIONAL_CROSS_CARRIER_CONTROL_OUTPUT-<control>.json full returned-report writer, with the existing sourceCommit/entry/...observed pattern. It also applies all exact accepted absolute/delta guard values and explicit canonical stringify/coverage normal fences. Historical measurement comments remain.

Claude3 structural absence checks are retained, not counted as independent behavioral discrimination: nested member problem absence and no receipt/root checks are structural; the sequence negative is distinguished by presentation and created-record failures plus its full positive member Problem oracle. Claude5 redundant literal ID pins are deferred because the independent complete row hash and exact failure/positive sets already bind them. Claude7's existing named presentation weakening is retained; no fourth weakening or new registered case is added.

The afterExchange hook is an intentional hard gate: a thrown hook assertion aborts the cohort rather than reporting an ordinary row assertion failure. Future execution reports must preserve that distinction.

Manifest/coverage, production, exports, runtime limits, request/response payloads, control transformations, exact expected assertion sets, catalog, fixture, test registration count and deadlines are unchanged. The new static bundle/schema/digest checks run inside the already registered preflight case. No tests/build/typecheck/lint/install/pack/service/commit/push ran. The only source-preparation command was the existing Biome formatter at the path in source-binding.json, applied solely to the two owned TS paths (2 files formatted, 1 fixed).

amendment-against-original.patch isolates all successor changes against the reviewed original. final.patch is the complete four-file diff against eff171839fe4f37638284a4c3346428979d26b5d. source-binding.json pins final bytes. Independent native closure and root execution release remain required; no runtime pass or public profile qualification is claimed.

## Native successor correction v2

The first repair is preserved under v1/ with its source binding, complete/amendment patches, report and exact app source. Native found the derivative's compact execution digest was compared with the original pretty-disk bundle digest, so a serialization-only difference could satisfy the changed-content check. The sole v2 source correction computes canonicalExecutionDigest from createHash("sha256").update(JSON.stringify(canonical)).digest("hex") and compares both derivatives with that same compact-form canonical digest. The distinct-two-derivative check and every other source behavior remain unchanged. canonical-digest-correction.patch isolates this fix. The app file alone was formatted with the same available CLI; no tests/build/typecheck/install/service execution occurred.
