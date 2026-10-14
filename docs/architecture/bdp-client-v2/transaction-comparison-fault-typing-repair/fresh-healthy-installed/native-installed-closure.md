# Native healthy installed evidence closure

No actionable findings. Read-only closure of the single healthy installed14/1 regression at clean source `fc47fbfb5f1055766e2f680df40a1d729d39638d`, tree `11641e37dc87560c4e92928fed86c88b2b314412`. This supplements, without replacing, the source/control closure. It does not qualify installed maintenance-fault behavior or clear CI/merge gates.

## Package closure and identity

Read the final report, summary, provenance, root packed audit, raw outcome/cleanup and setup-failure records. Independently verified all313 installed member fingerprints against the provenance inventory and all eight actual package manifests/versions, including every declared dependency edge resolved from real package roots and confined to consumer-v4. The closure is @bdp/server0.0.0, @bdp/protocol0.0.0, ajv8.17.1, ajv-formats3.0.1, fast-deep-equal3.1.3, fast-uri3.1.5, json-schema-traverse1.0.0, require-from-string2.0.2. Consumer manifest/settings/lock and frozen source-lock digests match the recorded bindings. Root's separate313-member tar/source/installed audit explicitly accounts for root LICENSE and the server workspace dependency rewrite.

Independently reproduced installed-payload.ts's read-only algorithm: sorted file traversal, supported extension filtering, package.json/dist/schemas handling for BDP packages, sorted recursive dependency identity paths, internal-symlink rejection and no deduplication of shared dependencies across paths. Eight distinct packages produce fourteen identity-path visits and473 hashed file visits including repeats. Computed digest exactly equals child-derived `2e982099f4c4e3688c47734256e2cc6b1cb461948a265bcc5fe1945fdfdda5d0`. No product module was executed for this reproduction. Node/deployment remain outside this payload digest as specified by the actual implementation.

The actual entry resolves into the fresh consumer and its hash matches the record. Narrow supervisor override preserves that installed entry plus exact source label; workspace fallback is excluded by the recorded provenance checks. Attribution is external clean-source/build/package evidence, not child source attestation. The report correctly discloses that complete raw IPC-ready data was not separately saved: the retained binding is the existing harness's validated projection. Response evidence retains fingerprints/context rather than complete bytes.

## Execution, failures and bounds

The sole selected test passes once, fourteen attempted/completed exchanges, one child27563 exiting0/no signal. Actual statuses are seed200/200, conflicts409/409, negotiation406/406, capacity413, then seven successful receipt/state/retry observations200. Existing5s exchange/90s cohort/120s case and supervisor/byte limits remain unchanged; maximum authored body4616bytes. No extra probe, mutant, restart or test retry is represented.

Three offline consumer attempts exited1 before product execution: missing local protocol resolution, ineffective package.json override, then uncached fast-uri3.1.8. The separately authorized v4 workspace override uses the same protocol tar and fast-uri3.1.5 matching frozen dependencies. No archive was edited/repacked or registry fallback used. The initial symlink-facing dependency-inspection failure is separately retained and correctly classified as a verifier failure, with zero product runs; corrected realpath traversal precedes the sole test. These failures are not converted into passing evidence.

Cleanup reports directory absence and no surviving command groups. Independently rechecked child27563 and every recorded supervised group absent, and source HEAD/clean state unchanged. The source/control and installed runs remain separately attributed; aggregate21HTTP/one installed child is workload accounting only.

## Verified stable report hashes

- report.md: `c22f49fb2fc4794353b8a9bd8d5ef92378e6c9d1850907742ddc398798ef44b8`
- summary.json: `058469f3461cefb35539ece4566c47a6cd34f2bcc906fc04d333f5606a20073f`
- installed-provenance.json: `99b15d16c66b39c0ffe05e13634d9eb2b98dd082ce9786e19a85c83a0f870193`
- cleanup.json: `28fddfa248b9e7f9ac4e83119eebb1eb7c2d327865926cdc67537596f2ff7298`

Reviewer ran only read/hash/process inspection and wrote this separate report. No tests, builds, installs, services, source edits, network or PR actions.
