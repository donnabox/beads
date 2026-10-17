# F0 candidate reconciliation and review dispositions

This is implementation review evidence, not a normative specification or runtime
adoption claim. No storage, CLI, installed catalog, provenance representation,
retyping operation, or protocol capability is added.

## Inputs and scope

Historical implementation commit `1935d820ba1f8bdd42bb1897438f04e34b486627`
was reviewed against public `gastownhall/bdp` candidate
`38791e4cc6e16cf661fa839303d4ff315d49ab3d`. Native Codex reported no defects;
Claude supplied the eleven findings below. This follow-up preserves that commit.

The reviewed candidate successor was
`0b151ac9f5cadf4d92f76d24134d48edaf1f0e0e`. Before editing, all nine JSON inputs
were fetched from the public repository and compared with the old pinned bytes.
All six Link pair sets and all five conformance declarations were unchanged.
The five child property schemas were byte-identical. The changes were:

- `association` replaces the candidate `memory-related` identity; existing
  Dependency identities and semantics are untouched.
- Root `kind` is permitted as uninterpreted extension data, never dispatch.
- Inception is removed from Memory's property contract. Bespoke validation is
  removed; no shared provenance validation or new open-root prohibition replaces it.
- There are 60 active source cases: 72 minus 14 retired Inception cases plus two
  dispatch cases. Seven former kind refusals are now positive cases. Retired cases
  are not reported as passed.

The final pin is public `gastownhall/bdp`
`d6337ab96e93d1b9a7b93beb4aea6d48bd179311`. All nine copied JSON inputs were
fetched again and verified byte-identical to `0b151ac9`; their SHA256s are recorded
in `testdata/candidate/provenance.json`. The later prose-only CLI direction does
not expand this package's scope. Memory remains a proposed specification.

## Claude findings

| # | Disposition |
|---|---|
| 1. Trailing-newline checker divergence | Fixed documentation/tests. Added timestamp and authority URI cases with `valid: false`, `schemaValid: true`. Keep the stricter Go rejection; do not reproduce Python's regex newline bug. |
| 2. Oracle is only a remembered command | Added explicit `check.sh` acceptance entrypoint for package tests plus the offline oracle. It fails on missing cached dependencies; ordinary Go tests do not invoke `uv`, fetch dependencies, or skip the gate. CI wiring is an explicit later integration task, not claimed here. |
| 3. Unpinned format dependencies | `oracle-requirements.txt` pins jsonschema 4.25.1, rfc3987 1.3.8, and rfc3339-validator 0.1.4; the checker asserts all three installed versions. |
| 4. Link identities as endpoints | Caller contract now explicitly requires proving Bead categories refer to Beads, never Links. This pure helper does not resolve resource identity. |
| 5. Compare class plus canonical identity | Rejected. Classification and locality must not create identity namespaces or let self-links bypass the rule. Clarified the common canonical identity contract and tested same-identity cross-class refusals. |
| 6. Coupled family oracle | Replaced production `IsMemory` use in the fixture oracle with an independent explicit test inventory. Candidate conformance declarations remain separately checked. |
| 7. Empty additional cases silently pass | Both Go and Python checks now reject empty additional cases. No brittle exact local-case count is imposed. |
| 8. Common properties exercise only Association | Every Link property case now runs against all six Link kinds. |
| 9. All-external pair invariant | Added an explicit behavioral test rejecting all cross-scope/resource combinations for every Link kind, separate from descriptor equivalence. |
| 10. Shared enum-range guard | Deferred optional polish. The bounds intentionally differ: `OtherBead` is an endpoint classification but not a conformance contract. A shared maximum would erase that distinction. Unknown and OtherBead conformance behavior is tested. |
| 11. Reference diagnostic | Invalid authority branch shapes now report `a URI string or {uri, revision} reference`; tests assert the structured expectation. Inception-method diagnostic concerns are obsolete after removal. |

## Verification

Executed in the isolated Beads worktree on final Preview 2 base
`703bb13e91aa2cab2092443319139006d74242c8`:

- `internal/memorytypes/check.sh -race -cover`: passed; Go package coverage 93.9%,
  no race report. Offline oracle passed 124 property cases (60 source + 64 local).
  Link property cases exercise all six kinds. Endpoint tests cover 600 candidate
  combinations, exact-family declarations, pins, self-identity, correlated arms,
  opaque categories, and all-external refusal.
- `make ci-pr-lint`: passed, zero issues for native, Windows non-CGO, and Darwin
  non-CGO; repository formatting check passed.
- `python3 tools/bazel/go_srcs.py --check`: passed; no Go files/imports were added
  or removed in this follow-up, so existing BUILD source lists remain unchanged.
- `bash -n internal/memorytypes/check.sh` and `git diff --check`: passed.

The original `make bazel-sync` attempt remains unavailable because Bazel is not
installed; no Bazel execution is claimed. No full-repository, engine, installed
binary, real database, or shared-server qualification was run. The source owner's
full artifact checker (which also needs shared Types artifacts) is distinct from
this package's property oracle and was not run here. Parent owns final review,
publication, and release admission.
