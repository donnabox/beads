# Native final source and evidence audit — PR65/PR83 landing candidates

**No new findings or native-review blockers. All prior native findings remain closed.** This is an independent bounded source/evidence audit, not a claim that other reviewers have completed or that public write profiles are qualified.

## Exact final scope

- PR65 documentation follow-up: `074a6e32781834062d80f42df71e069f553b0ae9` in `bdp-revision-provenance`.
- PR83 candidate: `2c8b63c293b0342c8bf819ad49a69c8ff8a2916e` in `bdp-provenance-runtime-integration`.
- PR83 fresh seal: `fccaa537c478b8f70046eb5b685326147714ad48`, whose observed parent is `919baf57897359287c5cbd4b7e8c0eae9dd03537`.
- Subsequent PR83 merge `5b48a5800add26f950018c53a961650eb09efba1` incorporates PR65 `074a6e3` ancestry, including the already-reviewed anti-relabel test. Its final child `2c8b63c` changes only STATUS.

Both author checkouts were clean when inspected. This reviewer made no source edits or outward changes and did not repeat full suites.

## Final source checks

`074a6e3` changes four files only: STATUS, descriptive prose in two illustrative fixture JSON files, and the prose review note in the private graph-policy JSON. Fixture request/response bodies, schema definitions, source pins, runtime code and public API are unchanged. The fixture narration now correctly names preservation of the entire prior revisionProvenance envelope. The graph-policy note distinguishes a private source-hash refresh from the separate mandatory Read projection recomputation and fresh cohort. No wire or compatibility rule changes are introduced.

The PR83 run-head-to-seal delta contains exactly the evidence artifact and shipped constant. The seal-to-final delta is limited to STATUS, the two fixture narratives, the anti-relabel test and the graph-policy review note. No schema, runtime, request identity, retained record, observer, catalog or evidence bytes changed after the sealed run. Ordinary Git ancestry includes the final PR65 follow-up; no observation was relabeled to a later source.

PR83 STATUS now cites its actual integrated cohort rather than the standalone PR65 predecessor and labels earlier runs as historical. Its source-matched claims are accurate:

- RU recovery store format 3.
- Private Transactional format `bdp-private-transaction-provenance-5` and identity policy `transaction-revision-provenance-4`, checked before recovery maintenance.
- Development incarnation `bdptest-development-2-revision-provenance`.
- Component tests and private development observations are distinguished from unfinished Transactional HTTP/full-profile evidence; development runs remain `claimEligible:false`.

The repaired type, diagnostics, strict Read-wire checks, current-schema roots, metadata selector and owned-Link provenance citations remain the implementations reviewed in the preceding addenda. The meaningful anti-relabel assertion survives the final merge. Deferred inherited properties-update numeric-location work remains a before-public-Update-advertisement condition, not a newly invented full-profile landing prerequisite.

## Independent seal verification

Invoked exported `main()` of `scripts/read-cohort-evidence.mjs` with Node v24.16.0 in both final checkouts; both return exit 0 with `evidence.verified`.

PR83 `native-final-seal-verify.log` confirms:

- run head `919baf57897359287c5cbd4b7e8c0eae9dd03537`;
- evidence commit `fccaa537c478b8f70046eb5b685326147714ad48`;
- constant `debd39622ae93c855eba015e179c7124c62c0c1b`;
- Read projection `bd7da0fc3f8082e0bfb6f4969c04409f2a610e12408dbc6a0bd14b4b351127cc`;
- 47 sealed definitions, 46 reachable, 89 passing applicable target-row observations.

The generation log `../pr83-provenance-reseal-generate.log` records that same head/constant and 234.36-second run. STATUS's 38 packaged and 11 self-certified **case definitions** are not misrepresented as 89 black-box rows; the artifact preserves per-target packaged/in-process distinctions and N/A cases.

Final PR65 verification is preserved in `../pr65-review-repairs/native-docs-followup-evidence.log`; the narrative-only follow-up leaves its valid `c8c8c2d` seal over `a68db90` intact.

## Validation evidence inspected, not rerun

- PR65 `../pr65-review-repairs/final-tests-serial.log`: **7620 passed, 1 skipped, 108 files**, duration 143.31 seconds. This closes the pending-full-suite status of the earlier native closure report. The original two ENOENT PID-fixture failures and 138-test focused successful rerun remain preserved; no causal explanation is asserted and no test was weakened.
- PR83 `final-tests.log`: **7860 passed, 1 skipped, 117 files**, duration 130.12 seconds.
- PR83 post-merge `final-docs-tests.log`: **275 passed, 4 files**. Follow-up typecheck and formatting pass.
- PR83 recorded build, typecheck, boundaries, formatting, normal executable smoke, development executable smoke and seven E2E preflight checks pass. Lint completes with five warnings and one informational notice; this is not called warning-free.

The evidence claim is Read-only. Passing component tests and development smoke do not create Read+Update, Transactional or History capability admission. Native review is complete for the exact scopes above; other reviewers' pending results remain the integration owner's separate gate.

## Documentation-only follow-up `bf53b5fb4ec0a30083cac9577c4f5c289a718c73`

Independently inspected the exact two-file PR65 child of `074a6e3` (4 additions/2 deletions). No findings. `recovery-store.ts` changes only its function documentation from “Format 2” to “Format 3”, matching the unchanged runtime format constant. `STATUS.md` narrows its description of the evidence gate: catalog, manifest, fixture and Read-projection bindings are re-derived and ancestry checked; remaining recorded bindings are format-checked provenance, not hashes re-derived from today's tree. Checked this against `deriveReadInputBindings`, projection derivation and verification input assembly.

No executable expression, schema, parser, current Read assertion, observer, fixture body, request identity, retained-state policy, artifact or capability constant changes. This follow-up does not warrant a new observed source or fresh cohort; the existing actual run head and seal must remain intact. Build/evidence/format checks were underway in the parent workstream and are not claimed as completed by this addendum. Ordinary integration into PR83 should preserve its distinct `919baf5`/`fccaa537` cohort narration while adopting only this verification-scope clarification and comment correction.
