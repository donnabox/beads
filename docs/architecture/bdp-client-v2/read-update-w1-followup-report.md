# Bounded W1 external-review follow-up

Frozen source: `b6e671195ce2897e53f25f0b4f8dbb8c3b8b7c00`, direct child of W1b `f2ce17bafd1c3d218cf404b815e12e288983fe05`. Checkout `/Users/dbox/repos/wt/vickie/vickie-read-update-w1-followup`, branch `codex/vickie-read-update-w1-followup`, origin `https://github.com/gastownhall/bdp.git`. Both predecessors remain immutable. No outward writes.

Six changed files: matrix harness, Read response oracle tests, cumulative manifest, private coverage inventory, inventory test, progress design document. No runtime/evaluator/protocol/runner/public schema/admission/sealed artifact changes. The private record remains ineligible for claims.

## Dispositions of read-update-w1-claude-review.txt

1. **Row-local alias controls:** actual alias PUT followed by assertion-input substitution of a fabricated `revision` fails `put-alias.body-schema` and `.alias-result` in `alias.no-version-minted`. Separately, the first independent target GET after the actual alias PUT receives a fabricated revision and fails only `put-alias-a.postimage`. Each control requires exactly one substitution, a failing row and the exact named failed assertions. Original observations are retained; these controls do not claim corrupted authority wire behavior. `alias.put-creates` now accurately names sensitivity demonstrated in the byte-identical put-repoints workflow rather than claiming its own control.
2. **Omitted reasons:** generic text now says no row-local scenario; it does not assert that no relevant executable observation exists. Discovery-aliases explicitly names observed directory/resolution in other scenarios. Patch-validation names nonexistent-path/atomic rejection and the new metadata non-object refusal but leaves effective-schema and other row-local obligations uncovered. W1b has since given update-link and owned-link-source-revision their own bounded scenarios, so their current inventory reasons remain specific to that work.
3. **Pin equality:** `create-link` is explicitly incomplete (`wholeContract:false`). Store/echo/no interpretation remain observed, but no discriminating equality-comparison observation is added. The exact whole-ID audit removes only this row. Proposed whole count is now **27**, with **102 of129 required IDs uncovered**, across36 authored scenarios. This is an honest scope boundary, not a normative interpretation or capability change.
4. **Metadata non-object:** update-bead sends a real root `metadataChange` replacement with numeric value5. It requires422 `validation-failed` and the closed problem schema, then independently GETs the complete unchanged preceding record. Its whole flag remains false for effective-Type-schema rejection. No Type schema/config changes.
5. **Alias workflow trimming deferred** as authorized. The current repeated workflows discriminate the cited clauses; reducing duplication would cause avoidable evidence churn. No inventory or runner contract redesign.
6. **Cleanup order:** oracle port allocation precedes `mkdtempSync`; configuration creation and child launch run inside cleanup. Matrix harness also allocates its port before creating its temporary directory. A port allocation failure therefore cannot leak either new directory.
7. **Fixed client GET allowlist:** routes now derive from actual client `installed.exact-request` GET actions and their known fixture target bindings. No blanket fixture expansion or extra operation-directory GET is needed. Dynamically captured Bead identities retain the existing restricted per-request route gate. A negative control adds an unused Link fixture binding and checks that its digest stays identical to the baseline. No wildcard/public transport change.

## Verification

Pinned Node24.16.0 and pnpm11.20.0, frozen install.

- Clean → pre-build typecheck → build: PASS (`read-update-w1-followup-final-build.log`).
- All conformance src plus matrix/oracle/fixture/exact-client tests: **847 tests /28 files PASS in33.77s** (`read-update-w1-followup-final-tests.log/.json`).
- Changed-file format/lint, dependency boundaries for11 packages, diff check: PASS (`read-update-w1-followup-checks.log`). Full format is not claimed; root separately owns the unchanged testing.ts format correction.
- Existing89-row Read evidence verifies unchanged. No new sealed evidence generated.
- First focused run had46 pass and2 stale expected-digest failures, because the narrower exact route set changes its identity. Both normal and unused-binding runs independently returned `f04e98d80971e813b73d1b57d225c0a6bfb0669647d6bbf9b026a908bbf37f26`; updated only these expectations, then the full focused regression passed. No behavior/schema assertions were weakened.

Fresh packed observations are recorded below once complete. No full workspace suite or cumulative profile qualification is claimed. W1b whole flags remain subject to independent council review.

## Frozen packed result

Fresh protocol/server tarballs installed together outside the checkout: **53 tests /4 files PASS in27.29s**. Exact logs, packages and progress are under `read-update-installed-b6e6711/`. The unfiltered run records134 scenarios:36pass/98harness-error (absent scenarios),129required/36authored/102uncovered,36sessions,337client exchanges,44setup exchanges and21native Read observations. Source is clean at the frozen commit; claim eligibility remains false.

```json
{
  "sourceCommit": "b6e671195ce2897e53f25f0b4f8dbb8c3b8b7c00",
  "sourceDirty": false,
  "claimEligible": false,
  "configuration": {
    "scope": "https://reference.invalid/acme/",
    "profile": "read-update",
    "history": false,
    "sourceCommit": "b6e671195ce2897e53f25f0b4f8dbb8c3b8b7c00",
    "installedPayloadDigest": "6503d708974149ece46a9e3ef5d676cdd13ef74393b2bff1f417937392e5889c",
    "installedContractsDigest": "e1dff401b5caafb04a4d935775d67eeefeae7bcf306d5d100b979a64eef148e7",
    "policyDigest": "dbe8dbd5430f26e19e754d38167cd8084a32d702c695550e2ee0063dc2912430",
    "authenticationDigest": "66941efc779f4f2431d65fdd5b2387b2af67bd7c1559e09dac032f82259ea16e",
    "fixtureDigest": "3a60db8714da2490ad67f9bf4f281f50da293393b31c6e1d5bec1ece12fd955c",
    "fixtureId": "reference-read-update-v1",
    "capabilities": [
      "installed-read-update-http"
    ]
  },
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
      "manifestDigest": "fcc9f68af721d9ea6725a83e29c24d4cc375fc9dbd653180c2e0b924e608a7b6",
      "fixtureDigest": "3a60db8714da2490ad67f9bf4f281f50da293393b31c6e1d5bec1ece12fd955c"
    },
    "exactConfigurationDigest": "f04e98d80971e813b73d1b57d225c0a6bfb0669647d6bbf9b026a908bbf37f26"
  }
}
```

Actual source association and payload selection remain launcher/reviewer checked, not independent target attestation. All child-backed test cases completed and cleanup paths returned. The source checkout remains unchanged after freeze.
