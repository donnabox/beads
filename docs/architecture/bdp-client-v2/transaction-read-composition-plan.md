# Private Transactional Read composition

Base: cd22ac3e388be3baffaf1d3a4caac7c4bd288502. This bounded implementation plan supplements the sole fork design plan; it introduces no protocol ruling or profile claim.

## Slice 1 (authorized)

Narrow AuthorityReadPlane's private reader to resource/resources/incidentLinks/alias/policy. Keep the complete RecoveryStore/SequenceLifecycle RU reader and its key semantics. Add an expiring synchronous reader to the Transactional owner using its existing exclusive connection/entry guard, and a private transaction-read composition over the existing Read plane. Configuration and installed descriptors come from the actual fixed owner installation, not an independent inventory. No diagnostic inspect or fake RU key adapter.

Rows/collections are immutable. Every facade method expires after callback return or failure. Reject async callbacks and reentrant read/write/close. Respect sticky fencing and preserve causes. Canonical reads, properties/revisions, lists/selectors, incident Links, installed Types, aliases and retained pagination use existing Read semantics. Schema-free/unowned Types and fixed all-visible policy remain explicit private restrictions.

Files: authority-read.ts and its tests; read-sequence-lifecycle.test.ts annotation adjustments; transaction-store.ts/tests; new private transaction-read.ts/tests; transaction implementation progress note; installed-package smoke output allowlist. No public exports, schema/spec/format change, HTTP listener, shared RU runtime edits, or profile advertisement.

Acceptance: atomic old/new graph visibility; no failed-batch partial state; canonical bodies/endpoints/revisions; live/dangling/repointed/deleted alias behavior; owner-installed descriptor/configuration truth; lifetime/async/reentry failures leave a healthy owner usable; fences preserve causes; valid reads after restart recovery; retained pagination rows/revisions across mutation and explicit invalidation after close/restart; existing Read and RU regressions preserved. Run pinned Node24.16/pnpm11.20 focused tests, build/typecheck/boundaries/format/lint, both package smokes and evidence verification. Do not reseal.

## Slice 2 (separately scoped before HTTP)

Capture source context inside the same owner read entry. Keep generic internal source context separate from Transactional serving fields; Read profiles gain no required Scope-history concepts. Do not derive context from a later inspect call, a receipt's effect position, or shared mutable last-position state.

Plan context propagation through ordinary results and refusals, plus retained collection snapshots. Keep the snapshot's original rows/revisions/context distinct from current continuation validation context. Existing pagination retains view/epoch but no source position. Derive an explicit acceptance matrix from spec collection snapshot semantics (1551ff), strict TX reads (5697ff), and TX context/minimum-position/conditional ordering (5742ff) before claiming serving semantics. Receipt reads, HTTP fields/minimum barriers, replication Snapshot handles and changefeeds remain outside slice 1.

## Coordination

Update worker owns server/index.ts, new read-http-routing.ts, and development-authority/runtime. Its route/Directory work does not need the narrowed reader seam. Both branches may add separate private output entries to scripts/smoke-executables.mjs; integrate additively. Root owns outward publication and review.
