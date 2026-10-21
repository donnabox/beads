# Frozen Read+Update artifact retention

Frozen clean **8e3781ca78bc69571f6baa0e27d8d1f593986694**, direct child of reviewed TX capture **cbecfeac9b2279e7424d2499d558e976ff3f9dcf**. Checkout `/Users/dbox/repos/wt/vickie/vickie-read-update-artifact-retention`, branch `codex/vickie-read-update-artifact-retention`, origin gastownhall/bdp. No outward writes. Root lifted the authoring hold, then confirmed the shared accessor external review found no correctness defect; no dependency source was changed here.

Six files: recovery-store.ts, development-authority.ts, recovery-store.test.ts, new development-retention.test.ts, update-implementation-progress.md, and the two current format references/explanation in STATUS.md. No shared development-contracts, TX, compiler, hosting, schema, public profile, or sealed evidence changes.

## Implementation

Format4 stores one contractCapture row in metadata, unreachable from member policy writes. Default opaque text is `null`; missing metadata is not absence. Development authority serializes its existing branded capture under the finite142,610,432-byte ceiling and always supplies the complete frozen Type inventory (including built-in mode). Installation/lineage remains compact and unchanged. Capture contains raw artifact records and all three compiler-limit witnesses; no access-policy/authentication identity enters it.

Storage snapshots primitive-only own expectation fields and the Type map, refusing new accessor/proxy/nonstring inputs without executing their hooks. It interprets no capture structure. Under existing locked open, after ordinary format/identity checks, `length(CAST(value AS BLOB))` must equal expected UTF8 byte length before stored text is fetched; exact capture and complete Type equality precede claimed-key deletion/startup commit. No capture JSON CHECK/parser is added. Direct callers without expected inventory retain their low-level behavior, but captured stores refuse an omitted capture expectation.

The redundant authority post-open inventory check moved into that gate. Resource-Type consistency and retained runtime-limit checks remain post-open and may refuse after claim cleanup/expiry; no universal ordering fix is claimed. Writable SQLite hot-journal restoration remains before metadata readability. Ordinary refusal preserves bytes; journal recovery may necessarily change interrupted bytes before BDP startup refuses.

Full capture is retained for installation lifetime, without GC, refresh, migration or reseed. Administrative input remains required on reopen; saved raw bytes reconstruct through the same compiler offline. Shared capture and serialized expectation retain additional bounded in-memory representations; hosting is unchanged. Public nonattesting development-facade stores also encounter the intentional format3→4 break. Built-in literal Types do not gain a whole exact-artifact conformance claim.

## Tests and controls

- Configured and built-in authorities seed actual successful/failed dispositions and a live Resource; a real store admission leaves an abandoned pending claim. Fourteen corruption cases cover missing/replaced capture, raw artifact bytes with unchanged installation identity, each compiler limit witness, and changed/missing/extra installed Types. Refused opens preserve ordinary file bytes and pending row at an advanced clock. Restoring only disposable test corruption then valid reopen reports exact recoveredClaims; authority can reopen again.
- Opaque text tests distinguish missing expectation from captured value, use equal character counts/different UTF8 byte counts, and instrument StatementSync to prove the value fetch is not reached after length mismatch. Invalid expectation accessors/proxies/nonstrings cause no hook calls or new file. Built-in map is frozen; store inventory snapshot is detached.
- Noncanonical raw descriptor whitespace and transitive schema are retained exactly, distinct from canonical installed_types. Saved capture reconstructs equal compiler identity. Real valid create/update, retained invalid schema failure, changed graph, restart replay, deletion and expiry leave capture unchanged.
- Existing real child crash harness gained one isolated hot mode: tiny SQLite cache plus staged bulk writes produces the asserted nonzero rollback-journal magic. SIGKILL, incompatible expected capture, and valid reopen prove SQLite recovery happens without BDP claim sweep on refusal; committed saved disposition survives, tentative resources vanish, and valid open reports two recovered claims.
- External `predecessor.mjs` uses the actual frozen predecessor emitted format3 opener and current emitted format4 opener in both directions. Both refuse with main-file bytes identical and pending key intact. No fabricated old version marker or migration is used.

## Validation

Pinned Node24.16.0/pnpm11.20.0. Frozen install, typecheck and build PASS before emitted tests. **763 tests /12 files PASS in33.61seconds**, maxWorkers2: recovery store, new retention, all direct-caller evaluator/lifecycle/Read/disclosure suites, existing compiled development contracts, installed-reference tests and cumulative RU matrix. Machine-readable `regressions.json` plus complete log/exit in this directory.

Final test-helper cleanup removed all new lint warnings and reran typecheck plus all3 retention tests (14 corruption subcases + lifetime). Production emitted bytes did not change after the successful build/regressions. Boundaries, touched-source format/lint, git diff check, unchanged89-row evidence verifier and both smokes PASS. No hook was skipped; this isolated clone had no installed pre-commit hook.

Initial focused run was85pass/1failed because the new test expected an internal retained wrapper rather than the authority's actual disposition object. Corrected the oracle; focused87tests then passed. Initial TypeScript errors were spy overload/this annotations and a temporary unused test import. No production semantics were relaxed to cure tests.

### Fresh packed observations at the frozen source

Packed actual protocol/server tarballs were installed outside the source checkout into `packed/install`. **35 tests /2 files PASS in24.68seconds** against that installed entry: exact client and cumulative RU matrix. `packed/progress.json` records source8e3781c, sourceDirty:false, claimEligible:false,36 pass/98 missing-harness outcomes,36 authored IDs and102 uncovered contracts. This is the inherited cbecfeac corpus; later W2 observations are on a separate branch and are not claimed here.

Payload digest `4c9a4bb1885af3671bef8796b9b7fbbdb6853e0701ff0685bda770261aeb6f7a`; contract digest `e1dff401b5caafb04a4d935775d67eeefeae7bcf306d5d100b979a64eef148e7`. Artifact/config bindings remain the unchanged inherited corpus; source/payload are fresh observations, not relabeled old evidence.

A separate **packed retention probe** (`packed/retention-probe.mjs/.log/.exit`) launches actual IPC-installed HTTP with additional whitespace-preserving descriptor/transitive schemas, creates a valid Resource and retains an invalid update, checks hosted schema GET/HEAD, closes and reads exact persisted capture, reconstructs the same compiler from saved bytes, then restarts on a different listener under stable Scope and replays both dispositions/read state/hosted bytes unchanged. Its own custom artifact identity is `development-schema5-0fc00f4fb3d2ce9a39608bc1070a04271ab6f838e38d86264a91d1322311e276`, separate from baseline matrix configuration; payload is identical. This is explicit retention evidence, not new whole-contract credit or public qualification.

No full workspace suite or new public Read cohort was run. Existing Read seal/artifact/constant are unchanged. Test processes completed; final process check found no remaining packed installed-reference child. Disposable stores were removed; frozen tarballs/logs remain outside the tracked tree for review.

## Remaining scope

Resource-Type/runtime-limit postcommit refusals, broader Type lifecycle/GC, automatic installation loading/migration, public profile activation, materialization memory optimization, new HTTP/schema semantics and other authors' TX/W2 integration remain outside this change. Root owns independent council, ordinary integration, final combined checks and publication. Source is frozen; no further edits planned.
