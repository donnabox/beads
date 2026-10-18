# Read+Update exact installed-artifact retention: revised next slice

Planning only. Original audit basis was `8d64fcf7fdaa34642fadb3c2c0b8f4729c7778df`; this revision inspected the unchanged RU seams at TX successor `cbecfeac9b2279e7424d2499d558e976ff3f9dcf` in `vickie-transaction-contracts`. That successor adds the shared immutable capture. Implementation will use a NEW isolated successor of cbecfeac only after root confirms its external accessor review has closed. No source edits, tests, evidence regeneration, or outward actions in this planning turn.

## Contract and exact gap

`docs/specs/bdp.md:2532–2576,2578–2582` (`#installed-type-contracts`, `#descriptor-installation-for-mutation`) requires exact pinned descriptor/schema artifacts and an integrity fingerprint, immutable validation closure, no same-ID refresh, and offline validation/retry/replay/recovery. Effective contracts intersect inherited schemas (`#effective-type-contracts`,562–605). Exact artifacts survive while live/retained representations refer to them, with Type ID/fingerprint surviving for the logical Scope lifetime.

Current safeguards already prevent configuration drift: `development-contracts.ts:80–165` compiles solely from detached captured bytes, brands the registry, and binds artifact digest plus all compiler limits into identity. `development-authority.ts:143–154` derives installation/lineage from that identity and aggregate policy. Changed schema bytes or artifact limits therefore already refuse on reopen. That proves identity mismatch refusal, not retained byte availability.

`installed-reference-configuration.ts:123–159` captures raw inputs and separately keeps schema hosting bytes in memory; `:160–186` binds aggregate/authentication/policy/profile/owner limits into the served candidate. `recovery-store.ts:557–561` stores canonical descriptor text only, not raw descriptor or transitive schema bytes.

The ordering gap is specific: store open deletes abandoned claims at `recovery-store.ts:613` and commits at617; exact installed-Type inventory validation occurs later at `development-authority.ts:193–198`. Move **capture and installed-inventory** checks inside locked open. Resource-Type consistency (`development-authority.ts:199–204`) and runtime-limit compatibility (`:205–207`) remain existing post-commit refusal classes in this slice. This is not a universal startup atomicity repair.

## Fixed implementation boundary

### Capture and metadata

Consume the reviewed `DevelopmentContracts.capture` from cbecfeac. It contains exactly detached artifact records and `artifactLimits`, `graphLimits`, `evaluatorLimits`. Keep the existing compiler identity algorithm. “Policy” in the earlier draft meant compiler-policy witnesses only: persist NO live access-policy, authorization, authentication, credential, or request-principal identity. Aggregate identity is already covered by installation metadata and needs no second capture field.

Use **one row in the existing metadata table only**. It is unreachable by MemberTransaction writers; `policy` and `installed_types` are excluded. No new table or expectedTables change. The row always exists under the new format, with explicit text `null` for uncaptured/schema-free paths; missing row is corruption, not absence.

The authority (or a single small private serializer) serializes the branded capture into bounded opaque text before store opening. Preserve exact artifact records, including original base64 bytes, URI, length and digest, and all three limit sets. Use the finite conservative TX-equivalent ceiling `8*totalBytes + 1024*submittedArtifacts + 4096` =142,610,432 UTF8 bytes; it covers separately bounded URI escaping, base64 expansion and entry overhead. Do not impose an unrelated1MiB ceiling. Never put capture bytes inside installationId, lineageId or Read configuration identity: compact identity is a hard invariant.

The storage interface receives **opaque pre-serialized bounded expected text** plus an optional detached frozen `Record<id, bodyJson>` of expected canonical Types. It does not import the compiler, parse capture JSON, hash saved bytes, run arbitrary startup callbacks, or interpret schema structure. Validate option primitives and snapshot the Type map before touching files; reject proxy/accessor/nonstring new option data rather than invoking hooks. The expected capture is already bounded by the authority. Inside the lock, query `length(CAST(value AS BLOB))` before materializing the saved text; compare with the expected UTF8 byte length (an exact bound, which also rejects oversize), then compare text. No capture JSON CHECK is added, so quick_check does not parse capture payloads. Existing other-table integrity checks retain their behavior.

Freeze the built-in `types` map at `development-authority.ts:107`. Configured `contracts.types` is already frozen. The authority always supplies the complete expected inventory, including built-in mode. Compare all rows and canonical bytes, including missing/extra/ID mismatch, inside locked startup before pending cleanup. Remove the redundant later inventory check while preserving the later Resource-Type and runtime-limit scans and their explicit residual ordering.

### Defaults, lifetime and compatibility

Direct store callers default to expected uncaptured text `null`; omitted inventory skips only inventory equality. The inventory gate is opt-in and always enabled by development-authority. A captured store opened without its expectation refuses because present capture differs from `null`; the default is not a bypass. Existing direct tests that seed Types and reopen without them retain their documented low-level store behavior.

At explicit new-store creation, write capture metadata atomically with installation. At reopen, preserve writable SQLite journal recovery, then check ordinary format/Scope/installation/lineage and new capture/inventory under BEGIN EXCLUSIVE before claims deletion, creator/allocation changes or startup commit. Capture/inventory refusal leaves ordinary files byte-identical and claimed rows intact. A hot rollback journal may necessarily alter physical bytes before metadata becomes readable; after recovery, no BDP cleanup runs on refused capture/inventory. No readonly preflight, sidecar, migration, reseed or unsafe SQLite parsing.

Retain full capture for installation/store lifetime with no GC, including after all live Resources are deleted and receipts expire. Root has approved this conservative lifetime choice; no Type lifecycle decision is pending. Administrative configuration remains required on reopen. A test reconstructs from saved bytes through existing ingestion/compiler offline; there is no new auto-loading installer.

Bump recovery-store format3 to4 for mandatory capture metadata. This intentionally rejects existing on-disk stores of the **public, nonattesting development facade** as well as private installed fixtures. It is an unmigrated format break within the authorized0.0/private deployment work; do not describe it as fixture-only. No retained outcome/identity codec or installation string changes are needed merely for the format bump. Compiler ceiling/policy changes also deliberately fail closed with no migration; all three persisted witnesses permit attribution.

### Hosting and conformance limits

Existing hosting already uses the same accepted input bytes; leave that path unchanged. Exact accepted capture/configuration equality ensures they correspond to the pinned installation. Regression-check existing hosting and policy/auth reopens; do not rewire serving or add a schema route.

Built-in descriptors originate as JS literals, so compiled-capture absence plus canonical descriptor inventory does **not** earn whole installed-Type exact-artifact conformance. Keep this non-claim explicit. No public profile activation, new Type lifecycle, schema/evaluator change, transport vocabulary, owned-Type expansion or new product semantics is proposed.

## Actual caller inventory

`rg openRecoveryStore` on cbecfeac finds one non-test direct caller, `packages/server/src/development-authority.ts:180`, plus callers across eight test files:

| Caller | Relevant default/reopen behavior | Required handling |
|---|---|---|
| recovery-store.test.ts | Bare helper; many create/reopen calls; emitted child at766; staged crash harness898ff | Keep default absence; add metadata/option/inventory/oldformat/hotjournal tests. |
| sequence-executor.test.ts | Seeded Types then reopen; emitted child1902 | Default absence preserves direct semantics; rerun actual subprocess/recovery paths. |
| member-executor.test.ts | Seeded resources/Types at111, bare reopen226/348 | No artificial inventory requirement or fixture rewrite. |
| alias-evaluator.test.ts | Seed resources75; bare reopen166 | Preserve absence default. |
| resource-evaluator.test.ts | Multiple seeded stores/bare reopens1448ff and child variants | Preserve absence default; relevant regression suite. |
| read-sequence-lifecycle.test.ts | Seed Types139, bare reopen1021 | Preserve absence default; exercise lifecycle recovery. |
| authority-read.test.ts | Seeded store91 | Preserve absence default and Read behavior. |
| retained-disclosure.test.ts | Direct store40 | Preserve absence default; no policy retention changes. |
| development-authority.ts | Configured and built-in production composition | Always pass frozen complete inventory and capture/absence expectation. |

The scan also includes recovery-store.ts's own function declaration. Public `development.ts` → development-runtime → authority reaches the built-in path. Installed private factory reaches the configured path. No other production caller was found. Existing `apps/bdptest/src/installed-reference.test.ts:484–503` explicitly permits policy and credential changes on same-store reopen; preserve it.

## Bounded tests and done predicates

1. **Store option/default boundary.** New text/map input is detached and primitive-only; accessor/proxy/nonstring refuses before file creation. Absence row always exists. Bare direct reopen works, captured store without expectation refuses, captured versus uncaptured mismatch refuses. Oversized saved text is refused by byte-length query before materialization; missing row differs from `null`. Prove UTF8 byte accounting rather than character count.
2. **Configured capture corruption before cleanup.** Seed actual compiled inherited/transitive contracts, successful/failed outcomes, live Resource and claimed key. Alter a byte while keeping claimed fingerprint; delete capture; change each limit witness; separately change/remove/add installed-Type rows. Refused open preserves claimed raw row, outcomes/counters and ordinary whole-file bytes. Restore only disposable test corruption, then valid reopen reports exact `runtime.recoveredClaims` removal count. Do not say successful recovery preserves the claimed row.
3. **Built-in absence/inventory path.** With literal installation ID unchanged and `null` capture, corrupt/remove/add an installed-Type row; each refusal preserves pending key before cleanup. Missing/replaced absence marker also refuses. Valid reopen reports recoveredClaims. No whole installed-Type conformance credit for built-in literal descriptors.
4. **Exact bytes and lifetime.** Use raw descriptor whitespace plus transitive schema, persist actual complete capture, reconstruct offline with identical compiler identity/types, and confirm original bytes differ from canonical installed_types as expected. Valid/invalid mutations and replay after graph changes remain correct. Delete live Resources, perform normal expiry and reopen; capture remains exact. Existing caller-mutation tests plus new detached expectation controls preserve snapshot semantics.
5. **Compatibility and real recovery.** Produce a real format3 store with frozen predecessor opener, then new format4 refuses byte-identically with pending key intact and no new metadata. Reuse staged actual hot-journal harness: compatible reopen succeeds; mismatched expected capture after SQLite recovery refuses before BDP cleanup; later correct open reports the pending recovery count and retained facts. No power-loss guarantee or expectation of physical byte identity across SQLite recovery.
6. **Authority/installed regressions.** Existing Type/Resource reads, schema GET/HEAD, validation diagnostics and replay still use accepted artifacts; existing policy replacement and changed policy/credential reopen remain allowed. Capture/inventory refusal must not emit ready or leave listener/store open. Resource-Type and limits scans remain postcommit residuals; test/report does not claim they moved.
7. **Validation/evidence.** Pinned Node24.16/pnpm11.20; build before emitted tests; targeted new store/authority tests, lifecycle/compiler/installed startup tests and relevant cumulative matrix. Run type/boundaries/format/lint, both smokes and evidence verifier; root handles combined broader checks. At final accepted clean source, renew actual packed RU observations with new payload/config binding and explicit retention test evidence. Preserve old observations and public Read seal; report any sealed binding change rather than relabel/reseal unasked. Existing evidence verification alone does not prove retention.

## Council dispositions

| Claude finding | Root disposition incorporated |
|---|---|
| H1 policy ambiguity | Compiler witnesses only; no access/auth identities. |
| H2 writable table ambiguity | Existing metadata row only. |
| H3 trust/parameter gap | Opaque bounded pre-serialized text plus frozen Type map; authority serializer, no store parser/compiler. |
| H4 universal ordering claim | Only capture/inventory gate moves; Resource-Type/limit postcommit refusals explicitly residual. |
| M5 saved read bounds | UTF8 BLOB length before text materialization, no capture CHECK/parser. |
| M6 defaults/caller churn | Explicit `null` default, optional inventory; authority always opts in; captured-without-expectation refuses. |
| M7 limits | Persist artifact/graph/evaluator witnesses, deliberate fail-closed constant change. |
| M8 facade compatibility | Name intentional unmigrated public nonattesting development format3→4 break. |
| M9 recovery wording | Refusal preserves claimed row; valid reopen reports recoveredClaims deletion. |
| M10 built-in coverage | Separate absence/inventory corruption controls required. |
| M11 hosting | Existing serving copy stays unchanged; equality follows accepted artifacts. |
| L12 citations | Corrected spans and immutable-closure2532–2582 included. |
| L13 mutable builtin map | Freeze the expected built-in map. |
| L14 table list | No new table; expectedTables unchanged. |

Semantic questions are resolved within root scope: built-in exact-artifact conformance is not claimed; whole-capture lifetime/noGC is approved; compact installation/lineage identity is mandatory. No human ruling is needed for these implementation choices.

## Ownership and implementation hold

This RU lane owns recovery-store, development-authority, a minimal private serializer only if it reduces coupling, and bounded tests/progress documentation. It consumes the reviewed shared accessor without editing TX store/runtime or duplicating the compiler/capture. Installed hosting code, normative spec, public API/profile gate, schema/evaluator, Types lifecycle, Janet CLI/Preview assets and shared databases are out of scope. Root's explicit shared-accessor closure is the remaining start gate. No source implementation begins before that confirmation.
