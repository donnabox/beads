# PR83 bounded council repair handoff

Frozen clean source: `e5f7d0c43c151426350ceededdc77fe7e372db9b` on `codex/vickie-bdp-pr83-council-repairs` in `/Users/dbox/repos/wt/vickie/bdp-pr83-council-repairs`.

Base: `b0a6bca90c6d0578e545ec4b4a13c999da019884`. Own repair commit `6e8d10c`; ordinary merge parents are that commit and PR64 numeric repair `4594d526ebb3b1388c288733369afc00703e1eff`. No outward writes, artifact/constant changes, schema changes, identity changes, or private format changes.

## Repairs

- Restored the four distinct external-endpoint requirements from cd22ac3, keeping the current first requirement. Catalog audit now rejects duplicate source/anchor/normalized-excerpt triples within each scenario. Actual pre-fix red run: 1 failed, 20 passed, with expected 5 versus observed 1 distinct requirements.
- Shared numeric formatter omits an unapplied properties patch instanceLocation, just as metadata already did. Caller-supplied document mislabels are stripped; message, type and schemaLocation remain. Create document-relative locations are unchanged.
- Updated only the development authority and transaction-batch reference callbacks to preserve the offending update pointer in the message, explicitly labelled request location. Arbitrary callback messages remain untouched. Actual Transactional tests prove longer messages are charged against the 65536-byte bound, truncation behavior, failed-state rollback, retained replay and reopening. Real development HTTP regression verifies both patch document planes and unchanged Resource state.
- Merged PR64's symmetric normative omission and nested-refusal row plus independent member-identity/executor replay/identity controls.
- Corrected the seal narrative: derived root count 20→16 does not itself alter the projection, minProperties:1 does, and validationDiagnostic.document is outside the Read projection. Marked W1 section 5.4 vectors historical without changing recorded hashes.
- Deferred normalization count optimization as directed.

## Merge resolution

Three conflicts: member-identity.ts retained the shared formatter call, whose behavior includes the incoming inline fix; member-identity.test.ts retained both test sets; bdp.md retained the four-part/common-envelope document and applied only the two exact normative edits from 4594. The added numeric catalog citation uses the owning #readupdate-problem-details anchor. Automatically merged executor tests reviewed. No stale inline formatter reintroduced.

## Verification

Pinned Node 24.16.0 / pnpm 11.20.0, maxWorkers=2. Build completed before tests.

- PASS: 387 tests in 8 files, 3.35 seconds. Development HTTP, member identity/executor, Transactional owner, catalog boundaries, Transactional catalog, numeric catalog, Read projection. Structured log: `pr83-council-merged-tests.json`; verbose log and exit file share the prefix.
- PASS: build, typecheck, dependency boundaries, format, lint; `pr83-council-merged-*.log` and `.exit`.
- PASS: git diff --check; clean frozen checkout.
- EXPECTED BLOCKER: evidence:verify exits 1, reporting current catalog bytes drift for bdptest and re-seal required. Exact log: `pr83-council-merged-evidence-verify.log`. Restoring actual Read obligations changes the catalog bytes. Existing evidence and capability token are untouched; the old seal is not claimed to cover this source.
- Full suite, smokes and fresh Read cohort were not repeated for this bounded repair; root owns final reseal and qualification after source review. No timed-out or interrupted run is counted as passing.

All logs are under `/Users/dbox/repos/wt/vickie/bdp-integration-logs/`.
