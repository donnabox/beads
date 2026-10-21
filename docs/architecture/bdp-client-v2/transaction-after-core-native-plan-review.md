# Native plan review: compiled contracts in the private TX owner

**Verdict: proceed with the bounded plan, incorporating the concrete acceptance clarifications below. No new product ruling or blocking architectural gap identified.** This is a plan review, not an implementation pass or qualification claim.

Reviewed `transaction-after-core-plan.md` against clean integration `3e2b8784daefd5893f990785cfe91f1d872a2ffe` in `vickie-installed-profiles-integration`. The plan header still names 2fa2d029; refresh that documentary basis when dispatching the successor. Current source, including the repaired writable SQLite startup, was used here. No source changes, test runs, or outward writes.

## Confirmed fit with actual source

1. `development-contracts.ts:54–145` already provides a branded immutable contract authority and guards registry lookup by exact canonical installed Type bytes. Its identity binds artifact digest, artifact limits, schema-graph limits and evaluator limits under the existing development-schema5 policy. Captured artifacts are currently internal; neither that object nor the installed-reference factory persists the complete descriptor/schema bytes. The revised plan correctly retracts any assumption of an existing durable installer.
2. `installed-contract-artifact.ts` already captures detached immutable entries (kind, retrievalUri, byteLength, sha256, bytesBase64), sorts/binds their exact URI and byte data into the artifact digest, and freezes limits. A minimal accessor on the branded authority can expose those existing records. Do not persist mutable Buffers, evaluator functions, Maps, or caller registry objects. Reconstruction can reuse ingestContractArtifacts and compileDevelopmentContracts without introducing a second compiler or parser.
3. `transaction-store.ts:179–198` currently parses descriptor-only configuration and rejects propertiesSchema/ownsOutgoing. Its inline registry at 978–983 has no property validator. Replacing only that registry with the validated branded registry is the appropriate seam; `resource-evaluator.ts:690–711,812` already traverses effective Type closure and applies every validator. Keep the schema-free fallback and reject all owned descriptors in the complete installed closure before file creation/open effects.
4. `transaction-store.ts:427–451` creates configuration metadata atomically and checks stored installation before pending receipt retraction, creator_facts deletion and retention sweep. Extend that same locked check/capture, rather than adding an unlocked or read-only preflight. SQLite hot-journal recovery must remain allowed before reading configuration. Normal incompatible-file byte identity and hot-journal restoration are correctly distinguished in the plan.
5. `transaction-store.ts:460–469` and `transaction-read.ts:37–43` already supply the same owner's installed descriptors/configuration to the Read plane. Derive those from the same accepted contract object and keep the owner as the entry gate. No second Read engine or mutable per-request installation is needed.

## Concrete acceptance clarifications

### A. Retain the exact capture for the installation/store lifetime

State that this slice has **no artifact garbage collection**. Keep the full immutable capture even after live Resources are deleted, receipt detail expires, or startup retracts pending work. The existing spec requires exact artifacts while live or retained historical representations refer to them, and Type ID/fingerprint for the logical Scope lifetime (`#installed-type-contracts`). Conservatively retaining the entire small bounded installation for that lifetime satisfies both without analyzing references from retained receipts/groups or inventing Type lifecycle.

Test this with actual create/delete plus retained success/failure receipts, restart and normal maintenance. Assert the saved artifact capture remains exact. Deleting or expiring all currently live Resources must not silently remove the capture. No migration, automatic reseed, artifact refresh or new GC authority is warranted.

### B. Preserve all current compiler identity inputs, not just artifact bytes

Expose/persist a closed, frozen reconstruction record containing the existing captured artifact entries and artifact limits, plus the actual schema graph/evaluator policy and limits already consumed by compilation. The identity algorithm itself should remain unchanged. Do not let saved data become an override for arbitrary compiler internals: unsupported saved policy/limit combinations must fail before BDP maintenance. Compare the supplied configuration and persisted capture in full; a matching digest alone is not evidence of retained bytes.

Required negative witnesses: one changed schema byte at the same URI, missing transitive schema, descriptor/capture mismatch, different bounded artifact limits, forged branded authority, and corrupted saved capture with an unchanged claimed fingerprint. The first five should refuse before open/maintenance as appropriate; persisted corruption must be validated under the locked installation check before any retraction/sweep. Capture-only accessor tests should prove unchanged prior identity and detached immutability (including nested arrays/limits), not merely Object.freeze on the outer view.

The plan already promises compiler-limit binding; this makes the test target precise. It does not require compiler changes or a new public configuration system.

### C. Distinguish schema references from schema HTTP hosting

The private TX runtime does not currently serve a captured schema-byte manifest like installed-reference-configuration does. Same-owner Type GET/probe/list can prove the descriptor's exact propertiesSchema URI and Resource validation can prove the installed closure is used. This does **not** establish an HTTP GET/HEAD schema-hosting claim.

Keep the proposed slice at exact persisted/offline artifacts, Type/Resource HTTP reads, and correct schema references. An externally named synthetic retrieval URI in the offline bundle is sufficient; requests must never dereference it. If the author chooses Scope-local schema references, document the absence of schema-hosting coverage rather than accidentally borrowing the RU factory's schema route. Do not add a second installer or new hosting surface under vague “schema references” wording.

### D. Freeze compatibility and atomicity evidence before broad claims

A private format bump is appropriate if capture metadata changes the required on-disk representation. Validate an actual independently produced predecessor-format file (not a fabricated version string alone), normal incompatible files unchanged, and existing valid hot-journal recovery. Refuse compiled-vs-schema-free installation mismatch and supplied descriptor disagreement before BDP maintenance; do not silently prefer one input.

For writes, retain the proposed late invalid batch and valid-set core rollback tests, including absence of tentative Resource IDs, group/position changes and receipt-page facts while preserving the failed terminal receipt. Keep sequence prefix commits and later continuation independently observed. Equal-key failed replay should be observed after a real graph change that could otherwise affect execution. Use independent Type/Resource reads after valid writes/restart; returned receipt equality alone cannot establish same-owner contract installation.

## Scope and unresolved work

The fixed all-visible, unowned authority is a sound bound. No Type lifecycle, hidden-policy closure, owned-Link source fanout, Snapshot/changefeed, HTTP set refusal rule or public profile activation is needed to complete it. Existing compilation/evaluation bounds remain explicit and this work should not broaden accepted schemas to make tests pass. If the reused compiler refuses a valid required schema or fails an admitted evaluation, report the actual failure to root; this plan does not authorize inventing a new wire error or mapping incomplete evaluation to false validation success.

Schema-free compatibility means the old administrative descriptor path remains available for newly provisioned stores under the selected current format. It does not mean older private files are silently upgraded. The separate recovery low findings remain owned by their author and are not introduced by this plan.

The minimal shared capture accessor plus TX installation metadata/registry/read composition is sufficient. The plan's runtime/HTTP tests, affected existing compiler tests, build/checks/smokes and evidence verification are appropriate; no fresh public conformance or evidence eligibility follows from them.
