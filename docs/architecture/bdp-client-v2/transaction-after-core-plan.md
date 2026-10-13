# Next bounded Transactional authority slice

Read-only proposal against frozen `2fa2d0291d0a4dec95bc8b4e864c8f30c00b74b8`
(timer-only child of fccb951). Sole fork design T2/T3 and required-contract matrix
remain authority. No new implementation is authorized by this memo.

## Recommendation

Bridge **the existing compiled installed-contract authority into the private
Transactional owner**, first for schema-backed unowned Types under the existing
fixed all-visible policy. Exercise actual singleton/batch/sequence HTTP writes and
same-owner Type/Resource reads. This closes the schema-free-only limitation using
already implemented contract semantics; it does not introduce Type lifecycle work,
new schemas/protocol, hidden policy, or profile advertisement.

Do not enable set HTTP while Selector refusal semantics remain unresolved. The
private valid-set core can nevertheless test that an invalid selected postimage
rolls back its whole transaction using the same compiled registry.

This is a better next boundary than adding a second Read engine or broad barrier
coverage. Current transaction-read.ts already composes AuthorityReadPlane with
same-entry context and retained original/current pagination observations. The old
c4fce0b source in bdp-transaction-read and its surviving log
`transaction-read/read-composition-plan.md` describe those earlier seams; current
fccb/2fa source wins. The specifically named transaction-read-pause-checkpoint.md
was not found in the supplied log location; no result from it is assumed.

## Concrete current seams

- `development-contracts.ts:54–145` exposes branded immutable DevelopmentContracts:
  identity, descriptors, installed Type bytes, and ResourceContractRegistry.
  Compilation consumes a captured offline artifact closure and existing evaluator;
  no authority request performs network resolution. Its registry checks exact bytes.
- `development-authority.ts:126–160,223–240` already derives installation identity,
  installed descriptors/bytes and Resource evaluation from that same contract object.
  `apps/bdptest/src/development-contracts.test.ts` supplies real transitive-schema,
  diagnostics, immutability, restart and unavailable-contract fixtures for reuse.
- `installed-reference-configuration.ts:130–183` captures artifacts, compiles them,
  preserves hosted schema bytes and constructs the pinned installed candidate. Its
  current profile is explicitly Read+Update and its shipped default contains owned
  Types. Do not route that whole configuration into TX by changing a profile string.
- `transaction-store.ts` still rejects every propertiesSchema at open, constructs
  descriptor-only contracts inline in evaluateTransaction, and fingerprints only
  its current installed Type bytes/configuration. readConfiguration already returns
  the actual owner's installed descriptors/config identity to transaction-read.
- `resource-evaluator.ts` already evaluates complete effective contracts via the
  registry, including inherited constraints and bounded final diagnostics. Reuse it;
  do not fork its validator or alter shared Type semantics.

Normative anchors at this source: effective Type contracts562–605; installed pinned
closure2532–2575; validation/results and diagnostics in the following Read+Update
section; Transactional whole-batch atomicity3860ff and receipt error closure5297–5332.
Read Type inventory/probing must represent the same installation used by mutations.
These are existing required contracts, not new Type proposals.

## Proposed smallest implementation sequence

1. **Installation bridge, before store open.** Accept only the existing branded
   DevelopmentContracts through a source-private TX composition seam. Derive
   descriptors and installed bytes from that authority, not an independent registry.
   Existing schema-free fixtures remain supported. If preserving the old `types`
   argument alongside an optional compiled contract is least disruptive, require
   byte-exact agreement before opening the owner; neither input may silently win.
   Reject forged registry objects, mismatched descriptor bytes and any owned Type
   until the separate owned-Link slice is implemented.
2. **Pinned retention and compatibility first.** Include the existing contract
   closure identity in the durable installation check, before BDP pending recovery,
   cleanup and maintenance. Preserve the hot-journal fix: SQLite recovery may precede
   that check. Reopening must use the exact pinned artifacts and compiler policy;
   changed/missing closure fails closed rather than refreshing an ID.
3. **One evaluation registry.** Pass the compiled registry into the existing owner
   evaluation options; remove only the schema-free rejection where this validated
   installation path exists. Keep source-private schema-free fallback honest.
   Same-owner Read configuration uses exactly the chosen descriptors and identity.
   No changes should be needed to AuthorityReadPlane, Resource evaluator, selector,
   receipt lifecycle or alias identity.
4. **Actual serving evidence.** Thread the bridge through the existing private
   transaction runtime fixture and test live create/update/delete, batch and sequence
   receipts; invalid writes; Type/Resource reads and schema references; restart/replay.
   This is an actual served slice behind private URLs, not an installed public profile
   selection. Do not advertise a Transactional Directory lacking replication routes.

### Minimal durable capture and shared ownership (root-approved plan revision)

The current installed-reference factory captures/compiler-memory artifacts, returns
artifactIdentity and separately hashes package payload; it DOES NOT persist the
complete descriptor/schema artifact bundle. There is no existing durable installer
to claim or reuse. Read+Update's retention gap is a separate follow-up; this slice
will not edit its runtime.

Root approved a minimal immutable captured-artifact view on the existing branded
DevelopmentContracts in development-contracts.ts. Reuse ingestContractArtifacts'
exact bytesBase64/retrievalUri/limits records, captured defensively with the existing
compiler. Keep its identity algorithm unchanged. This is neither a second compiler
nor a new installer/public API. Record all bounded artifacts needed to reconstruct
that identical compiler object offline, including compiler limits/policy identity.

At explicit TX creation, persist this exact capture alongside the closure identity
in existing locked installation metadata, atomically with owner creation. On reopen,
compare the supplied administrative configuration and exact persisted capture before
BDP pending retraction, creator cleanup or maintenance, after required SQLite journal
recovery. Reconstructing the branded object from the saved exact artifacts must work
offline; no ambient file refresh, URI fetching or mutable caller registry is permitted.
Administrative configuration may remain required on reopen. Equality of digest alone
is not the capture-retention witness.

Use the existing private format compatibility path if the metadata representation
requires a bump: refuse prior formats without migration/reseed; clean incompatible
files remain byte-identical and hot journals may be recovered before refusal. Keep
schema-free fallback stable, recording explicit absence of compiled capture; keep
owned Types refused. Final format marker will be selected only in the authorized
implementation, with actual predecessor-file refusal evidence.

Root has approved shared development-contracts.ts scope for the capture accessor;
main implementation awaits council. TX owns transaction-store.ts,
transaction-runtime.ts, transaction-installed-contracts.test.ts, targeted Read/HTTP
tests and progress note; a small private bridge only if it reduces coupling. The
shared module's existing tests must verify immutable capture and unchanged identity.
Smoke inventory changes only for a new private module. Installed-reference factories,
RU runtime, compiler, Resource evaluator, public exports/profile gate and evidence
remain untouched. This is an implementation choice under existing contracts, not an
operator product decision.

## Done-when evidence

1. A real unowned descriptor with inherited constraints and transitively referenced
   schema resources is compiled offline using the existing factory. Same-owner Type
   lookup/probe/list and mutation validation agree on its exact ID/content. Unknown
   Type fails as type-not-installed, with no lazy install or network work.
2. Valid create and update return ordinary Transactional receipts and can be read
   through actual HTTP with consistent revision/provenance/context. An invalid
   inherited/nested property yields schema-valid retained validation-failed with
   existing deterministic diagnostics/bounds; metadata remains schema-independent.
3. A later invalid batch operation rolls back prior creations/updates, identities,
   groups, Events and page plans while retaining the failed receipt. A new valid key
   succeeds. Equal retry retains the same failure even after other graph changes.
4. Sequence preserves independent earlier success and a failed member; it does not
   become atomic. Its result projection/retry/restart laws remain unchanged.
5. A private valid set selects multiple Resources and encounters an invalid final
   postimage; all prior selected mutations roll back. No served-set claim or negative
   Selector policy is inferred from that core test.
6. Same exact retained installation reopens and preserves receipt identities, pages
   and current Type reads. Changed schema bytes at the same URI, missing transitive
   artifact, changed contract identity, forged object or mismatched inventory refuses
   without BDP maintenance/key changes. Ordinary incompatible files remain byte-
   identical; valid hot journals still recover. Do not promise byte-identity for a
   journal rollback that necessarily restores committed database pages.
7. Caller mutation/accessors and later changes to source files cannot change the
   captured running installation. Requests, retries and recovery never perform remote
   resolution. Exact artifact retention is proved via the new locked installation metadata capture,
   not asserted from the digest alone.
8. Existing all-visible unowned behavior remains explicit. Owned Types continue to
   fail before owner use, and no receipt redaction/nonwritable-selected branch is
   claimed by this fixture. No schema/Type lifecycle or optional History work.
9. Build first, pinned Node24.16/pnpm11.20; affected TX and existing compiler/installed
   authority tests, same-owner Read regressions, type/format/lint/boundaries, both
   installed smokes and unchanged evidence. Root reviews and runs broader combined
   checks; new targeted diagnostic observations remain non-attesting.

## Why not the other candidates yet

**Owned Links:** the shared evaluator understands them, but transaction-batch still
hard-rejects sourceRevision and only records the direct subject in its changes map.
A correct slice must handle every source revision/postimage, aggregate constraints,
owned projection, source+Link Event agreement, receipt authorization dependencies
and set fanout. It is a coherent next slice after the installation bridge, not a
one-line removal of provisioning guards. No new Type rules are needed.

**Remaining Read barriers:** visible point and receipt/page complete tuples exist.
Absent/hidden point ordering, query continuation minimum semantics and unsupported
opaque/partial tuple behavior remain unresolved. Preserve current explicit gaps;
source-private installed contracts do not need those rulings.

**Snapshot/changefeed:** groups and current same-entry context exist, but first-class
Snapshot needs joint Bead/Link capture, immutable typed streams/manifest, expiry and
an exact exclusive replay checkpoint. Changefeed needs explicit start intent,
epoch/view-bound checkpoint validation, finite/SSE readers and replay retention.
Spec5740ff ties Snapshot expiry to availability of all subsequent projected groups;
a diagnostic inspect or ordinary query cursor cannot satisfy it. First plan the
shared durable checkpoint/replay-floor authority, then implement the serving pair.
Required empty erasures framing is distinct from optional History. No such machinery
is proposed in this smaller installed-contract slice.

No unresolved operator decision is required for the proposed schema-backed unowned
semantics. The minimal shared captured-artifact view is root-approved; main implementation
awaits council. RU durable retention remains a separate follow-up. HTTP set error policy and broader profile activation remain separate.

## Authorized implementation revision after both council reviews

Implementation base: 3e2b8784daefd5893f990785cfe91f1d872a2ffe. Recovery-low5e remains
separate for root integration. Both native and Claude plan reviews are incorporated.

- Capture is a SEPARATE bounded metadata row, never embedded in configurationIdentity.
  Existing configuration text stays bounded by the existing inventory plus small
  closure digest/runtime-limits fields; no AuthorityReadPlane changes. TX capture
  serialization has an explicit finite byte cap derived conservatively from artifact
  ceilings, including entry URI/metadata overhead; reject before opening any owner.
- Private runtime accepts JSON-transportable bounded plain artifacts (URI + UTF8 text
  and artifact limits), requires explicit stable Scope, then compiles inside its own
  module instance. Store accepts only the existing branded compiled object. Require
  all compiled Type IDs under Scope types/ using existing installed-factory policy.
  Preserve old schema-free fixture path. Exact descriptor agreement is checked.
- Separate capture contains detached ingest artifact records/limits and actual graph/
  evaluator constant witnesses. Identity algorithm unchanged; no exact-field validator
  is applied to the branded object. Persist/check exact bytes under locked installation
  metadata. Runtime properties/diagnostic limits are included in configuration before
  maintenance. Changed compiler constants fail closed; no overrides/reseed/migration.
- Existing identity supplies changed-artifact refusal; new capture supplies exact
  retention and offline reconstruction evidence. Recompile saved exact bytes through
  the SAME existing compiler with current policy; saved constants are comparison
  witnesses, never controls. Persist full capture for installation lifetime, with no
  artifact GC even after delete/expiry/retraction. Detect tampered capture despite an
  unchanged claimed fingerprint before maintenance.
- Explicit exclusions: aggregate policy remains empty, owned Types refused, no schema
  GET/HEAD hosting (external synthetic schema URI is offline-only), no public profile.
  Small inline server fixture duplicates only necessary transitive schema structure;
  no server-to-app imports. Existing shared compiler tests regress the accessor.
- Native A-D accepted as lifetime retention, tamper/limits/branding/refusal, honest
  schema-reference-only HTTP evidence and actual predecessor-format/crash tests.
  Claude two HIGHs accepted via separate row and compile-in-runtime input seam; all
  MEDIUMs accepted via scope-locality, limit pinning, identity attribution and build
  constant refusal. LOWs accepted via explicit aggregate/hosting limits, small fixture
  budget, corrected citations (installed factory123-186; authority143-154/231-237),
  and branded accessor without exact-object-field introspection.

Required focused positives/negatives remain the nine predicates above, with actual
child/dist restart/HTTP tests, offline saved-capture reconstruction, input bounds,
caller mutation, corrupted metadata and changed limits, retained capture through
normal maintenance, late batch/set rollback and sequence prefix commits. Root owns
outward actions and broader integration. Main implementation now authorized.


Implementation clarification: stringifyJsonValue preserves ECMAScript property
ordering; it is not a canonical identity serializer. Identity behavior is unchanged.
Root accepted runtime omission of redundant types when plain artifacts are supplied:
the child-side compiled descriptors supply the store inventory. When both are given,
exact installed-byte agreement is still required. Schema-free input requires types.
