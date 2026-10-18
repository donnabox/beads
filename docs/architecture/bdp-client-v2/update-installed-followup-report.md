# Installed composition and qualification follow-up

Frozen head: `39a76748329fa88eee6cd1ca31b9501c5ef7f395`
Base: `dea6f9e7a95a3e3eba9c667de39360346f25bdf4`
Checkout: `/Users/dbox/repos/wt/vickie/vickie-update-installed-followup`
Branch: `codex/vickie-update-installed-followup`
Origin: `https://github.com/gastownhall/bdp.git`

No outward writes. Both predecessor checkouts remain frozen. No authority-read, store, evaluator, schema or public admission edits.

## External 4078 finding dispositions

1–3 remain addressed by predecessor dea6f9e (global startup readiness, inert imports, owner limit tuple). Existing six predecessor regressions still pass.

4. One source-private `development-composition.ts` now owns canonical Read+Update discovery and Operation Directory routing. The installed factory and source fixture import the same emitted implementation. No implementation imports test support. Packed inventory/deep-import denial includes the new private module. All existing discovery/contracts/matrix tests now exercise the shipped composition bytes.

5. Shutdown retries listener cleanup without repeating owner cleanup. Concurrent callers share the attempt; the owner terminal completion is captured once across retries. A closed listener is not closed again. Tests inject a first listener failure that leaves the actual listener open, prove the real owner has released its lock by reopening under the same Scope, finish listener cleanup, and assert exactly one original owner-close invocation. A separate regression injects failure after actual owner closure and proves retries preserve that terminal error rather than erasing it or invoking owner close again. No shared lifecycle semantics changed.

6. Hosted schema routes set the non-attesting marker before negotiation/method/owner checks. Actual GET, HEAD and OPTIONS header assertions pass; captured schema bytes remain exact.

7. Payload identity now hashes the resolved declared runtime package closure. In addition to server/protocol emitted JS and protocol schema assets, it includes manifests and dependency JS/JSON/CJS/MJS/native/WASM files. Runtime dependencies resolve relative to the actual installed package; package identity paths are independent of absolute installation location. The protocol's Ajv/ajv-formats transitive dependencies are covered as well as fast-uri. Workspace-only server/protocol source/config files are excluded; installed manifests are included and can legitimately differ from workspace manifests. Node itself, TLS/deployment and undeclared/environment-injected executable code are outside this candidate identity. Dependency cycles/depth32 refuse before opening a listener/store. This is private candidate binding, not a general package attestation framework.

   An isolated copied server/fast-uri payload uses the actual read-only protocol dependency installation. Appending a comment to the copy's real `fast-uri/lib/utils.js` changes the next process's payloadIdentity while configurationIdentity and artifactIdentity remain equal. The shared installed package is never modified. Full packed smoke passes with the new binding.

8. A generic recursive output guard checks every key/value URI in the default reference closure against configured Scope, including normalized path traversal. Tests add future fields containing foreign propertiesSchema, nested URI values, URI keys, neighboring-path identities and dot-segment escapes; all refuse. Local future values pass. This guards the shipped reference remapping, not arbitrary custom administrative schema semantics or a Type lifecycle redesign.

9. A controlled discover refusal after real startup proves actual HTTP Operation Directory and hosted schema requests return503 Problem before static data is emitted. Restoring the same real owner's observation restores200. Only the discover observation is intercepted; requests, routing, owner and store are real. No production policy replacement hook is added.

## Actual qualification binding metadata (root F4 request)

Private capture → actual installed factory → IPC readiness now emits `policyIdentity`, `authenticationIdentity`, `profile: "read-update"`, `history: false`.

Policy identity comes directly from the captured administrative policy. Authentication identity hashes the exact ordered captured `{principal, tokenDigest}` mapping, with the existing canonical digest procedure; no raw token is emitted. Configuration identity binds those same values. Tests compare actual IPC ready against capture, observe policy-only and credential-only identity changes independently, assert no raw token in readiness, and reject caller-supplied binding/profile/history fields. Source commit is not emitted: it remains launcher-verified metadata.

Configuration and payload identity preimages change intentionally; historical identities/evidence are not relabeled. Persisted Type/aggregate installation identity is unchanged. Public Read admission/evidence remains separate; claimEligible remains false and public Write startup still refuses.

## Validation

Pinned Node24.16.0 / pnpm11.20.0. Frozen install and build PASS before process tests.

- Initial focused run: two fixture failures (new-port reopen changed development Scope; isolated copy lacked protocol Ajv dependencies). Fixed by using the real installed canonical Scope and preserving the actual read-only protocol package. Initial log retained as `focused.log`; no runtime failure hidden.
- Corrected focused run:19 tests /4 files PASS (`focused-2.log`).
- Broader regression:502 tests /17 files PASS, maxWorkers2,9.49s (`regression.log`, `regression.json`). Includes all497 predecessor tests plus five new observations.
- Final test-only addition: startup/cleanup suite6 tests PASS (`final-cleanup.log`), including one additional terminal-owner-error observation;503 distinct tests across completed runs. Runtime source unchanged after broader suite/smokes; no needless full repeat.
- Typecheck PASS after final test; boundaries/format/lint PASS (existing3 warnings/1 info). Final changed test formatted and git diff --check PASS.
- Packed smoke and development smoke PASS. Packed observation payload `06aae1bba99b1e20e9454a61f04815e34b7906d9c21211a07425f2ea8cc11456`, all9 targets, stable Scope, claimEligible:false; private module inventory/denial and public Read/write-refusal checks PASS.
- Existing public Read evidence verification PASS, unchanged89 rows and original run/evidence identifiers.

No interrupted/timed-out run or skipped hook. Clean committed tree. No full repository suite rerun; no complete write-profile eligibility, installed between-member crash, mid-commit/power-loss or production deployment claim. Full cumulative executor/applicability and independent profile-specific evidence/public gate remain subsequent work.
