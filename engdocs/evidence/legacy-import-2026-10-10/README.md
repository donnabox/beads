# Iris stage 1 legacy import qualification

Recommendation: **Preview 3**. Import's required focused cases pass in both
installed modes, but the broad repository gate is not green and exact-head CI /
Janet integration are outstanding. Donna decides the cutline. This report is
an independent candidate, not Preview 2 acceptance.

## Scope and inventory

Stage 1 imports ordinary `bd export` JSONL into a fresh graph workspace. It does
not implement graph-format import or export, in-place migration, history
restoration, release promotion or production-data conversion.

The [clean producer receipt](../../../cmd/bd/testdata/legacy-import/producer.json)
pins `76fde9c9cb0865d430bf12d76fc473490f51128a`, tree and binary hashes and every
ordinary CLI command. [The fixture](../../../cmd/bd/testdata/legacy-import/ordinary.jsonl)
comes exclusively from a disposable synthetic workspace using two Issues,
one local blocks edge, object metadata, unsorted label input, Unicode/multiline
text, a closed P0 prerequisite, one comment and one key/value Memory. No source
history appears in the ordinary dump. The [generator](../../../scripts/legacy-import-fixture.py)
requires a clean, version-matching canonical producer build and removes its
synthetic workspace. Parser/storage tests supply historical and refusal cases;
they do not draw from production stores.

The [user guide](../../../docs/reference/graph-legacy-import.md) is the precise
preserved/remapped/refused contract. Current Issue IDs survive and map to graph
URLs; Link IDs are newly allocated from the dependency pair; comment IDs,
authors/text/timestamps remain in the current feed; old numeric IDs become exact
decimal strings. Current metadata is canonicalized under graph JSON rules.
New retained versions record accepted current state; absent source history is
never claimed as preserved. The sole Issue prefix is adopted atomically, allowing
subsequent normal creates/dependencies. Mixed-prefix batches refuse.

## Executable evidence

[qualification.json](qualification.json) pins tested source `ba235baae`, canonical
installed binary and qualified engine. [installed-qualification.jsonl](installed-qualification.jsonl)
contains **95 passing tests/subtests, 0 failures, 0 skips**, all three packages
passing. Both embedded and ordinary shared-server modes executed. Server data was
newly provisioned on Iris's own port 18343, separate from other agents' stores.

The installed workflow uses independent processes for initialization, malformed
source, missing targets, unsupported hierarchy/history, cycles, empty input,
mixed prefixes, refused flags/selector combinations, oversized input, readonly,
dry-run, successful file apply, current/retained reads, comments, repeat apply /
dry-run and post-import creates/dependencies in both directions.

Lower-layer fault injection covers coordination, prefix adoption, native rows,
Issue and Link catalogs, source mappings/retained state, Memory allocation /
payload / retained state and the final import stage. Fingerprints include native
Issues/dependencies/comments/labels/child counters, journals/events, leases,
retained versions/mappings/catalogs/payloads, writer token, epoch, scope binding
and Issue prefix. All faults and dry-runs restore the same state. Successful
apply persists across reopening; repeat import cannot add versions or overwrite
rows. A 9 MiB input below the parser ceiling exceeds the persisted read budget
and rolls back completely on both engines. Resource-count overflow also refuses.

Additional both-mode field families exercise assignment/references, scheduling,
closing/session attribution, pinned/template, compaction, molecule/work/wisp
classification, sender, gate/timeouts/waiters and event payloads. The importer
compares native hydration and validates retained graph state before committing.
Lineage, explicit storage-class markers and unsupported state refuse explicitly.

## Repository checks and review

- [Lint](lint.log): `make ci-pr-lint`, zero findings on native and Windows/Darwin non-cgo profiles.
- [Docs](docs.log): `make check-docs`, passed. Live Mintlify preview inspected title, navigation, code block and preservation table on Node22.23.3; default host Node25 is unsupported by Mintlify.
- `make bazel-sync`: unavailable (`bazel` missing). BUILD sources/dependencies updated and sorted manually; `tools/bazel/go_srcs.py` passed. CI/Gazelle remains required.
- Full `make test`: initial run includes the repaired capability guard failure and a 25-minute CLI package timeout. It began before final repair, so is not an immutable final-source result. The run finished with exit2: CLI, file-descriptor and script-harness packages failed; graphstore passed (1,216.6 seconds). The script assertion inherited this run's `TEST_VERBOSE` setting and saw an unexpected `-test.v`; [rerunning that assertion with the setting cleared](scripts-contract-rerun.jsonl) passed. [Package outcomes](initial-broad-packages.txt) preserve the broad result. A full CLI rerun at `ba235baae` uses the runner's documented `TEST_TIMEOUT=40m` override and is still running.
- Contributor preflight: reviewed upstream #6633/#5131 doctor changes, #6355 dependency floors, #5642 GitHub-sync framing and #5635 imported audit-writer changes. None implements fresh graph import. #5635 remains complementary: this adapter uses native writers and refuses unrepresentable data rather than replacing the contributor's UTC normalization work. No contributor PR is closed or superseded.
- File-descriptor failure: `TestMarkInheritedCloexec_ChildDoesNotInherit` also fails on the untouched base. [The baseline receipt](baseline-fdhygiene.log) preserves that reproduction; no unrelated file-descriptor code was changed.
- Existing dependency durable-workflow and close-conflict regression tests passed in embedded mode; optional server cases skipped after owned server cleanup. [Receipt](dependency-regression.jsonl). These supplemental skips are distinct from the required 95-case import qualification, which ran both modes without skips.
- [Two-seat council](council.md): native Codex plus actual Claude, Gemini explicitly excluded by the operator. Confirmed blockers repaired, including uncertain-commit error precedence. Claude's bounded repair review and native follow-up both confirm no remaining blocker.

## Original failures and remaining gates

Earlier focused failures identified test harness output flags, nil-vs-empty
metadata normalization, label order, retained Memory actor and null JSON scalar
coercion. They were fixed before final qualification. The [original broad failure excerpts](initial-broad-failures.txt) and the
original council findings remain recorded; no failing case was skipped to get
the 95-case result.

Exact-head CI, BUILD synchronization, a green broad baseline and Janet's
integration/release qualification remain open. The governing tracker is
[donnabox/agent-coordination#6](https://github.com/donnabox/agent-coordination/issues/6);
[stage-1 issue #9](https://github.com/donnabox/agent-coordination/issues/9) supports
this PR without closing the subsequent export workstream. Graph-format import
follows its corresponding export in a later milestone.

[Cleanup receipt](cleanup.json): owned shared server exited0, its exclusively synthetic data was removed; Mintlify stopped after live QA; clean detached producer checkout removed only after verifying its base is published. Other agents' processes and data were untouched. Final broad/review dispositions will be added before handoff.
