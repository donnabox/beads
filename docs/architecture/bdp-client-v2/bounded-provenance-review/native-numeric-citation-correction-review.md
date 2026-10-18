# Native review — unapplied numeric update locations and citation correction

**No actionable findings or blockers in either frozen correction.** The formerly deferred numeric-update instance-location issue is closed for the reviewed inline and shared/private runtime paths. No source edits or outward actions were performed by this reviewer.

## Exact scope

- PR64 inline correction: `4594d526ebb3b1388c288733369afc00703e1eff` against parent `3deafd650958fe0845efcdfe37c868f98b7ec652`, from `/Users/dbox/repos/wt/vickie/bdp-pr64-numeric-update-location`. Five files, 116 additions/33 deletions.
- PR83 combined correction: `e5f7d0c43c151426350ceededdc77fe7e372db9b` against `b0a6bca90c6d0578e545ec4b4a13c999da019884`, from `/Users/dbox/repos/wt/vickie/bdp-pr83-council-repairs`. Thirteen files, 302 additions/26 deletions. Includes `6e8d10c7bdaf1f86663c53e594905238994cc36a` plus the ordinary merge of the PR64 correction. Parent confirmed this clean head was frozen.

## Dispositions

1. **Inline/shared formatter parity — closed.** Both implementations remove caller-supplied `instanceLocation` for unapplied properties numeric patches before generic numeric admission snapshots and budgets the final diagnostic. Metadata updates continue to omit it. A request-body path is no longer presented as a location in an unavailable postimage. Properties retain caller message and paired Type/schema fields while refusing caller document relabeling; metadata retains its mandatory discriminator and metadata-prefixed caller message and strips Type/schema fields. Neither mutates the callback's supplied object.
2. **Actual private default diagnostics — closed.** Development authority and Transactional batch defaults label update paths as `request location` in message text. The formatter preserves this distinction while omitting the structural instance location. Create defaults are unchanged, and the shared helper still derives document-relative create pointers from the actual lossless properties/metadata occurrence. No request/schema member, error code, or identity normalization rule changes.
3. **Budget correctness — closed.** Formatting happens before byte accounting, so omitted location bytes cannot cause premature truncation and retained message-path bytes still count. Independent tests cover exact-fit and one-byte-too-small callback budgets, a long discarded caller pointer, preserved metadata discriminator, count truncation, and a real Transactional update with exactly 65536 bytes of message diagnostics followed by an over-bound variant that retains fewer complete entries. No field is appended after accounting or dropped solely to fit.
4. **Atomicity and replay — closed.** Both Resource kinds' inline numeric update refusals preserve state and replay after reopen without invoking the formatter; an equal inadmissible numeric spelling replays while a different exact value conflicts. Shared Transactional tests cover properties and metadata across create/update on both Resource kinds, rollback of an earlier effectful prefix, skipped tail effects, retained receipt schema, and restart replay. The update message-budget test also verifies unchanged store state and retained replay. Actual development HTTP tests exercise both document planes and observe the complete diagnostic and unchanged Resource.
5. **Spec/catalog agreement — closed.** The operative diagnostic paragraph and numeric conformance row distinguish request pointers from document-relative positions and permit omission for unapplied patches in either plane. The PR83 merge correctly uses the Read+Update problem-details anchor in the four-part spec; PR64 uses its existing combined problem-details anchor. The change does not weaken numeric admission or invent a postimage-location algorithm.
6. **Read citation loss — closed.** The existing pin-preservation citation is retained and four mistakenly duplicated replacements are restored to their distinct endpoint shape, live in-Scope Bead, at-least-one-local-endpoint, and owned-Link representation obligations. Each resolves against its source. The regression checks distinct normalized `(source, anchor, selectedText)` tuples per scenario, so the previously repeated valid excerpt no longer passes merely because it resolves. This is traceability repair, not a new wire rule.
7. **Supporting documentation — coherent.** Historical digest vectors are explicitly labeled historical without rewriting bytes/digests. Projection documentation correctly distinguishes the direct-root reduction from 20 to 16 (no definition text change) from nonempty output provenance (sealed Read definition change), and the write-only diagnostic discriminator from the Read projection.

## Independent validation

Tests ran with Node v24.16.0 in separate exact-commit archives `native64-source/` and `native83-source/`, outside all author and root qualification checkouts. PR83 package links resolve to the private snapshot; its crash tests use privately built outputs.

- PR64 `native64-tests.log`: **218 passed, 3 files** — member identity, member executor and numeric-model catalog.
- PR83 `native83-build.log`: private `tsc -b`, exit 0.
- PR83 `native83-typecheck.log`: required `tsc -p tsconfig.check.json`, exit 0.
- PR83 `native83-tests.log`: **368 passed, 6 files** — member identity, member executor, transaction store, actual development server, profile/citation boundaries and numeric-model catalog.
- Total selected test executions: **586 passed across 9 file executions** (some suites intentionally run against both distinct implementations).

No full suites or fresh cohorts were rerun here. The Read catalog bytes changed in PR83, so earlier seals must not be relabeled as observations of this correction; root's actual fresh qualification remains separate. This review does not claim those in-progress qualifications finished or authorize public write-profile advertisement.

## Final ancestry/evidence-only audit

**No findings.** Audited PR65 `d4d78fc85f9645210564f428d785145d45400d6d` and PR83 `4a6c020296aac7e7331af8f52e527365786d83cd` without repeating runtime review or tests.

PR65 includes final PR64 `a57991f` as an ancestor through ordinary merge `cca16bbe67faebddc9be6644b6002eee70be4895`. Its actual observed source is that merge, sealed at `c197096c475af70a631e39e026af71c1e60c8ec2`; the run-head-to-seal diff is exactly artifact plus shipped constant. Its final `d4d78fc` child changes only STATUS.

PR83 `4a6c020` has `d4d78fc` as an ordinary merge parent/ancestor. Crucially, the complete diff from its actual integrated seal `dbaf63a24219e1ea3f80394746853598b0d4464a` to `4a6c020` is **STATUS only**. Runtime, wire/schema, tests, observer, catalog, artifact and constant bytes are unchanged. The resolved integrated artifact still names actual source `e5f7d0c43c151426350ceededdc77fe7e372db9b`; the standalone PR65 observations are identified separately as preserved historical ancestry, not substituted for integrated observations.

Independently invoked each current checkout's exported evidence-verifier `main()` with Node v24.16.0. Both exit 0:

- `native-final65-evidence.log`: source `cca16bbe…`, seal `c197096…`, constant `1497814237d74f3495ade9e80c06b0bd709d719c`.
- `native-final83-evidence.log`: source `e5f7d0c…`, seal `dbaf63a…`, constant `674982be2123b9b7aae630bf79c3f774122900cd`.
- Each verifies 89 applicable observations and projection `bd7da0fc3f8082e0bfb6f4969c04409f2a610e12408dbc6a0bd14b4b351127cc`, with 47 sealed definitions and 46 reachable.

STATUS claims match the preserved generation/full-run logs: PR65 162.25-second cohort and **7624 passed/1 skipped across 108 files**; PR83 203.18-second cohort and **7868 passed/1 skipped across 117 files**. The 38 packaged/11 self-certified counts are case definitions, with the established per-target observation distinctions retained. These full suites were inspected, not rerun. No fresh cohort is necessary for the final STATUS-only reconciliation, and no old observed source was relabeled.

No remote write or publication verification was performed by this reviewer. All native findings remain closed for these exact source/evidence scopes; this is not public write-profile qualification.
