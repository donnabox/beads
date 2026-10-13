# Native review: installed follow-up

Frozen head `39a76748329fa88eee6cd1ca31b9501c5ef7f395`; base `dea6f9e7a95a3e3eba9c667de39360346f25bdf4`. Read-only source review in independent detached checkout `/Users/dbox/repos/wt/vickie/bdp-update-installed-followup-native-review`. No author checkout, shared dependency, evidence or outward changes.

## Required repair: P2 payload identity silently excludes executed symlink targets

`packages/server/src/installed-reference.ts:31–35` traverses directories and hashes regular files, but silently ignores internal symbolic links. A declared runtime dependency can consequently execute changed bytes without changing its reported payload identity. This is a concrete residual of external finding7, not a claim that the unmodified pinned packages currently contain such a link or that public profile admission is enabled.

Independent isolated reproduction copies the actual server and fast-uri package, keeps the real protocol package read-only, moves `fast-uri/lib/utils.js` to an owned external `actual-utils.cjs`, and replaces the original file with a symlink. Both launches reach actual IPC readiness. Between launches, the target gets a statement writing `NATIVE_CHANGED_UTILS_EXECUTED` to stderr. Second launch executes that new statement, yet both payload identities equal `01aea107e7701baecb7fd7fd7a9c5449e93bb9c6a48133c192b2bd3aef855509`. The temporary copy/store are removed and both child processes are closed.

Exact probe: `/Users/dbox/repos/wt/vickie/bdp-integration-logs/update-installed-followup-native-symlink-repro.mjs`; result `.log` and `.exit` alongside. Exit0 means the probe reproduced the defect, not qualification success. No production dependency was changed.

Smallest bounded repair: refuse symbolic links encountered inside traversed package payload trees. Keep Node-resolved package-root links working (normal pnpm layout); no need for a generalized realpath traversal/attestation framework. Add a fail-closed regression for the reproduced file and an internal directory symlink, while retaining regular-file identity sensitivity. No independent case-sensitivity or root-cycle defect is asserted or audited here.

## Closed findings and confirmed boundaries

- External4: canonical fixture imports the emitted `development-composition.js`; installed factory uses that same implementation. No duplicate route body or implementation dependency on test support remains. Existing discovery/matrix tests now exercise the shipped source-private composition.
- External5: concurrent close callers share an attempt; failed listener cleanup can retry. Captured owner completion is terminal and never re-invoked or erased. Maintained tests prove owner lock release despite listener failure, subsequent listener cleanup, and retention of the original owner-close failure after retries. Original startup gate/inert import/owner limits remain unchanged.
- External6: schema branch sets non-attesting marker before owner observation, method or negotiation. Actual GET/HEAD/OPTIONS regressions preserve exact bytes, HEAD emptiness and method handling.
- External7: ordinary regular-file fast-uri mutation is now bound and tested; transitive declared dependencies and manifests are traversed. The internal-symlink omission above keeps this finding partially open. Node, TLS and environment-injected code remain explicitly outside candidate identity.
- External8: reference closure output recursively checks URI-looking keys and values; foreign schema identities, nested fields, neighboring scope paths and normalized traversal are rejected. This is a default-reference closure guard, not new custom Type semantics.
- External9: real HTTP directory/schema observations are gated by the current owner before static results. Controlled discover refusal returns503 Problem; restored owner observation returns200. The injected seam does not replace the store or HTTP path.
- Qualification metadata: actual capture computes policy identity and credential/principal digest, binds them plus profile/history into configuration identity, returns them from the factory and sends them over private IPC. Tests verify capture-to-ready equality, independent policy/credential changes, no raw token, and refusal of caller-supplied identity/profile/history fields. These are available for the next conformance integration; this review does not claim a cumulative runner already consumes/qualifies them.
- Package export/bin/CLI boundaries remain unchanged; new composition is private, inventory/deep-import denial checks added. Public Read admission and sealed evidence are untouched; candidate remains claimEligible:false.

## Independent validation

Pinned Node24.16.0 / pnpm11.20.0. Frozen install and build pass before tests. **51 tests passed across6 files in6.34s, maxWorkers2**: installed-reference, installed-reference-startup, installed-reference-import, installed-reference-closure, development-discovery, development-matrix. Logs `update-installed-followup-native-{install,build,focused}.log` and `.exit`; structured result `update-installed-followup-native-focused.json`. Independent symlink reproduction executed twice, final retained log includes executed-code proof.

No timeout/interrupted run; no source edits or leftover test files. Author and independent checkouts clean. Broader503-test/check/smoke/evidence claims remain author results; this review did not repeat them. Full profile eligibility, installed crash qualification and public startup admission remain future work. Root received the required repair and retains frozen39a for the external review.


## Successor closure

Payload symlink P2 is closed at `4cf6acd03774fd90e629ba63516ff1365865b00e`. Independent build and13 focused real-process tests passed; original actual-code symlink probe now refuses readiness and leaves the store absent. File/directory/manifest links fail closed; ordinary resolved package-root links and regular-file identity discrimination still work. This is identity/readiness admission protection, not prevention of pre-factory module initialization. Full closure/evidence: `update-payload-symlink-native-review.md`.
