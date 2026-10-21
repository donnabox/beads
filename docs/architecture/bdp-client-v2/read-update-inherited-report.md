# Inherited Read coverage on the installed Read+Update candidate

Source checkout `/Users/dbox/repos/wt/vickie/vickie-read-update-read-coverage`, branch `codex/vickie-read-update-read-coverage`, origin GitHub gastownhall/bdp. Successor of native-repair `68db54ac5d2971aaf02d10b829e2757d37349311`; both68 and310 reviewer checkouts remain immutable. Frozen clean final commit `c5627933dc932040a90aea463eb179dad00c8afa`. No outward actions.

## Observations and reviewed completeness

The same configured installed factory now executes 17 additional inherited Read scenarios, for 24 authored rows among 134 applicable rows. All 17 use the existing canonical selection/private bundle/shared assertions; no runner, server, protocol or public admission semantics changed. The required set stays 129. The accepted whole-contract list is pinned explicitly in `cumulative-artifact.test.ts`; source coverage inventory and real-run assertion require 112 uncovered. Passing a probe remains distinct from completing its cited clauses.

| Inherited row | Actual observations / assessment |
|---|---|
| scope-service-desc | Real successful Scope request and parsed registered service-desc link. Whole. |
| canonical-navigation | Discovered beads/links/types exact Scope-confined canonical paths. Whole. |
| types.inventory | Actual 13 shipped local Types; closed summary schema, exact inventory and canonical order. Whole. |
| types.descriptor | Real GET of all 13 hosted descriptors, closed normative schema and independent identity expectations. Whole. |
| collections.invalid-query | Unknown parameter and repeated limit/selector requests; actual 400 problem family/code/retry and cache checks. Whole. |
| http.method-405 | Unsupported method on Scope Read endpoint; exact Allow and body absence. Does not reject supported mutation routes. Whole. |
| resource.bead-record | Complete schema, independent open properties and metadata. Incomplete: cited preservation of untouched properties after an effectful update is not exercised here. |
| resource.link-record | Complete schema, independent identity/type/source/target/properties/metadata. Whole. |
| resource.properties-view | Both complete open properties objects independently checked, no record envelope. Whole. |
| resource.default-bounded | Non-owning Bead with real incident Links has no embedded incident-Link plane. Whole. |
| collections.beads | Complete Resource schema, exact independent contents, next:null at complete page, ascending canonical IDs despite reversed create order. Whole. |
| collections.links | Complete Resource schema; independent identity/type/endpoints/properties/metadata tuples, next:null and canonical order. Revisions are validated as actual records, never assigned fixture constants. Whole. |
| http.get-head | Actual Scope, discovery, collections, records, properties/link views, Type inventory and bead/Link Type descriptor GET/HEAD. Whole. |
| http.cache-cors | Authorization-dependent representations and missing Resource cache fields, actual Origin and OPTIONS observations on fixed CORS-disabled target. Whole for this configuration. |
| http.accept-negotiation | Real media weights/specificity/exclusions and missing/query precedence. Incomplete: authorization/non-disclosure precedence not observed with this single-principal fixture. |
| http.conditional-reads | Real observed validators, strong/weak/list/star matching,304/412 empty outcomes/cache and missing/query precedence. Incomplete: authorization/cursor failures remain unobserved; History absent. |
| http.head-conditional-parity | Actual paired GET/HEAD success, conditional, negotiation, missing/query responses. Incomplete: authorization/cursor refusal parity absent. |

These 13 new complete Read contracts plus the four previously reviewed RU contracts (create-bead, delete-bead, expected-revision, idempotency-conflict) make 17; seven other authored rows remain incomplete. The original Read matrix, diagnostic RU matrix and sealed Read evidence are untouched.26 inherited Read IDs remain entirely unbound, and four authored inherited rows have incomplete clauses. All missing rows still appear in the unfiltered result, never hidden or auto-N/A.

## Fixture and execution boundaries

Six static cases need no graph. Eleven seeded cases use the same captured fixture's two Beads/two Links, created in reverse canonical order only through actual credentialed create HTTP. Preparation checks actual identity/type/endpoints/properties/metadata against declared input and records each exchange; failed or missing setup prevents the scenario. Store files are neither inspected nor modified. Native observations independently compare with static fixture oracles; no server-derived expected revisions are invented. The fixture digest changes intentionally while installed contracts/policy/auth/profile remain fixed.

Every inherited discovery request validates RU discovery locally rather than referencing the later RU catalog row. The retired closed Read discovery row is absent, canonical order is unchanged and valid earlier Scope prerequisites remain. The new test pins this ordering and local schema behavior.

Finite native routes are generated once from authored static targets and fixture bindings, then deduplicated; no wildcard transport. Root approved the narrow adapter for shared-runner-resolved ETag header values: it admits only named Read fields, rejects unresolved/nonstring/accessor/unsafe/colliding fields, and delegates to the same raw exact executor/credential handle. Literal headers stay on native raw requests. Response/status/ETag/HEAD assertions remain in the shared runner. Safe actual serialized header observations are recorded; the private adapter never supplies expected outcomes. A real socket test verifies transmitted strong/weak conditional fields and HEAD response behavior; unsafe controls fail before execution.

## External310 A–H disposition

- A: exact reviewed17 whole-contract IDs pinned in the artifact test; required 129/uncovered112 pinned in actual-run test. This is a reviewed assessment, not machine-derived semantic proof. Every complete/incomplete flag has the rationale above and remains in the canonical inventory.
- B/C: inherited68 repair retained: update-bead remains incomplete for metadata/effective-schema clauses; direct no-op returned postimage equality and response-only counterfactual retained.
- D: renamed `guard-wire` to `guard-binding-agreement` everywhere and documented its request-object echo scope. Actual serializer digest/octet observations and real stale-guard behavior remain separate evidence; no claim that the echo proves bytes.
- E: documentation expressly states bootstrap self-consistency uses the installed module, payload/source association is reviewer/launcher-checked provenance, not independent target attestation. Per-scenario consistency checks remain substantive.
- F: planner readiness remains structural and conditional; no new readiness/admission framework. Actual exact runner and harness continue gating execution.
- G: added missing planner cells (read-v1/v2/Read; RU-v1/v1/RU; RU-v1/v2/Read), version in carrier table, and v1 non-Read refusal comment.
- H: capture planner return; assert its manifest digest equals bundle/report manifest digest. Pin actual existing exact configuration digest in the main run rather than duplicate identity calculation. Document that this existing projection excludes credentials; authentication identity is bound separately in private configuration.

## Validation and remaining work

Negative controls exercise wrong Type identity/navigation/query expectation, missing public-HTTP seed Resource, wrong independent properties/collection oracle, wrong conditional expectation, plus all prior stale-guard/no-op-response/credentials/configuration tests. Full conformance and related bdptest regressions, clean pre-build typecheck/build, boundaries, lint/format, existing Read evidence and a fresh packed clean-source run are the completion gates. Logs use `read-update-inherited-*` beside this report.

Remaining work is the canonical uncovered inventory: richer structural/direction/selector/continuation/authorization fixtures, controlled lifecycle obligations without fake hooks, remaining Update contracts and later Transactional execution. This slice does not enable public startup or generate cumulative evidence; `claimEligible:false` throughout.

## Final qualification of this bounded slice

- PASS: Node 24.16.0 / pnpm 11.20.0 frozen install; clean → pre-build typecheck → build on final source.
- PASS: 886 targeted tests / 34 files in 38.85 seconds (full conformance plus installed/reference/development/Read/RU bdptest suites), `read-update-inherited-final-tests.log` and `.json`.
- PASS: focused format/lint, dependency boundaries, git diff check and unchanged sealed Read evidence verification. Full workspace test suite was not run. Root's separate testing.ts format fix was not edited here; ordinary integration preserves it.
- PASS: freshly packed+installed protocol/server at exact clean c562793, 81 tests / 5 files in 12.59 seconds, including real conditional-header socket controls and installed counterfactuals. `read-update-installed-c562793/{pack.log,install.log,matrix.log,progress.json}`.
- Final actual unfiltered run: 134 applicable rows, 24 pass and 110 harness-error/not-implemented;129 normative requirements, 24 authored, 17 reviewed whole-contract and112 uncovered. SourceDirty:false and claimEligible:false.
- 24 scenario sessions; 44 existing actual client exchanges; 44 recorded public setup mutations; conditional semantic-to-exact exchanges separately recorded in nativeReadObservations; all remaining native exchanges are in the shared run report.

The final packed record is reviewer/launcher-checked provenance for the selected installed target, not independent source/payload attestation. Existing historical310/68 records remain intact. No public admission, evidence generation or outward writes occurred. The checkout is frozen for independent citation/oracle and A–H closure review.

Final identity binding:

```json
{
  "baselineReceipt": {
    "event": "qualification.ready",
    "scope": "https://reference.invalid/acme/",
    "claimEligible": false,
    "configurationIdentity": "0d292d7e7fd502ffd141986d47793c194d861de72ca5379e4eca630c3b33a198",
    "artifactIdentity": "development-schema5-e1dff401b5caafb04a4d935775d67eeefeae7bcf306d5d100b979a64eef148e7",
    "payloadIdentity": "6503d708974149ece46a9e3ef5d676cdd13ef74393b2bff1f417937392e5889c",
    "policyIdentity": "dbe8dbd5430f26e19e754d38167cd8084a32d702c695550e2ee0063dc2912430",
    "authenticationIdentity": "66941efc779f4f2431d65fdd5b2387b2af67bd7c1559e09dac032f82259ea16e",
    "profile": "read-update",
    "history": false
  },
  "reportBinding": {
    "artifacts": {
      "catalogDigest": "03ca39f24a988a16497d80a52c1fffad7e2c5dbd27a2a521679cc0183672b184",
      "manifestDigest": "bc3fd575e6038dc6807ad5cd323079bc33e288ec20f3a119e63701d8220b50f1",
      "fixtureDigest": "819d64520e40fa2f9ef6fed76096edd9d68c5181272f9db213bcd1f754a9ba6a"
    },
    "exactConfigurationDigest": "ae2f29b8cfe45a180057e387bfef03aa1f31957be0b32760f5d7217ad81fca13"
  },
  "nativeReadObservations": 21
}
```

## Subsequent native review correction (2026-10-12)

The frozen c562 source and packed progress above are historical observations. Native response-mutation probes identified missing whole-contract oracles in `read.collections.beads` (stored properties, metadata, Type) and `read.types.inventory` (name, describes). The 17-whole/112-uncovered completeness conclusion is withdrawn pending a successor repair and review; observed test/run counts are unchanged. No frozen source or historical progress JSON was rewritten.
