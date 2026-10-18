# F0 candidate reconciliation and review dispositions

This is implementation review evidence, not a normative specification or runtime
adoption claim. No storage, CLI, installed catalog, provenance representation,
retyping operation, or protocol capability is added.

## Historical review at implementation 499c227fa

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

The pin for that implementation is public `gastownhall/bdp`
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
| 5. Compare class plus canonical identity | Rejected. Classification and locality must not create identity namespaces or let self-links bypass the rule. Clarified the common canonical identity contract. The original cross-class test observed external-category refusal, not the identity branch; the final closure section corrects that evidence claim. |
| 6. Coupled family oracle | Replaced production `IsMemory` use in the fixture oracle with an independent explicit test inventory. Candidate conformance declarations remain separately checked. |
| 7. Empty additional cases silently pass | Both Go and Python checks now reject empty additional cases. No brittle exact local-case count is imposed. |
| 8. Common properties exercise only Association | Every Link property case now runs against all six Link kinds. |
| 9. All-external pair invariant | Added an explicit behavioral test rejecting all cross-scope/resource combinations for every Link kind, separate from descriptor equivalence. |
| 10. Shared enum-range guard | Deferred optional polish. The bounds intentionally differ: `OtherBead` is an endpoint classification but not a conformance contract. A shared maximum would erase that distinction. Unknown and OtherBead conformance behavior is tested. |
| 11. Reference diagnostic | Invalid authority branch shapes now report `a URI string or {uri, revision} reference`; tests assert the structured expectation. Inception-method diagnostic concerns are obsolete after removal. |

## Historical verification at implementation 499c227fa

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

## Historical body-primary reconciliation at implementation d22db0cd

This normal follow-up to implementation `499c227fa5d0c208db0bf8d63c27e6fae9927483`
pins public `gastownhall/bdp` candidate
`50b13ee08f79a00bf8b135495b3b4d432f811992`. Inputs were first compared with
body-primary successor `72023840672b1fb5407ae8b2729a9408992f392a`, then fetched and
compared again at `50b13ee`. All nine copied inputs and hashes now bind to the latter.
Earlier reviewed commits and the evidence above retain their original identity.

Accepted product changes:

- All six Memory Types require only common title/body. Empty body remains valid.
  Optional observedAt, alternatives, authority and applicability.description retain
  their contracts. The former statement/choice/steps/rule/takeaway requirements are
  removed; old root fields can survive as uninterpreted extensions.
- Structured applicability intervals are deferred. Comparison code is removed;
  the still-closed applicability object rejects validFrom/validUntil. Timestamp
  format checking remains for Observation.observedAt. Obsolete interval tests were
  retired rather than relabelled as passes; unknown fields and format edge cases
  remain covered where applicable.
- Explicit Type-only change is semantic and revision-worthy. The caller contract
  now says to compare affiliation as well as properties for no-op detection. F0
  does not implement mutation, revision allocation, or prose-quality validation.
- Association is directly authorable and is the common Link Type. Five supplied
  specializations conform to it and inherit optional nonblank note. A pure fixed
  Link-family predicate distinguishes nominal equality from family membership;
  future dynamic conformance is not specified as a permanently closed union.
- Only Supports and Challenges require exact targets. Supersedes accepts current
  or exact targets without implying redirects, deletion, incoming-link rewriting,
  or automatic repair.

The new main document is a product proposal; it does not prescribe a BDP
Type-definition representation. The retained exploratory definitions still have
broader wildcard/external endpoint arms. The worker raised that discrepancy and
parent explicitly selected the release product table: only local Issue/Memory
pairs are accepted. OtherBead and external categories refuse. Pair tests now use
the product §4.2 table independently; exploratory definitions supply only a subset
cross-check, not an authority for additional admission. Self-links still compare
canonical identity, and each allowed arm remains a correlated source/target pair.

The exploratory graph now contains ten declarations (five Memory children plus
five Association children). Tests check those supplied declarations without making
that representation a product prerequisite. No Inception/provenance representation,
Type installation, generic query engine, CLI or storage surface was added.

Current exact checks:

- `internal/memorytypes/check.sh -race -cover`: passed, 94.0% Go statement coverage,
  no race report. Offline oracle passed 137 property cases (64 source + 73 local).
  All six Link kinds receive common property cases; 600 endpoint combinations
  exercise the release product table, plus explicit pin/refusal/family checks.
- Exact upstream `scripts/check_memory_examples.py`, default mode, executed from
  `/tmp/trish-f0-family-source-check` with its own pinned artifacts and the package's
  pinned Python environment: passed seven schemas, 64 cases, nine Bead and seven
  local Link excerpts, without shared BDP inputs. This validates examples/properties,
  not graph mutation or runtime behavior. Optional `--with-bdp-exploration` mode was
  not run and is not counted as release qualification.
- `make ci-pr-lint`: passed zero issues on native, Windows non-CGO and Darwin
  non-CGO; formatting passed.
- `python3 tools/bazel/go_srcs.py --check`, `bash -n internal/memorytypes/check.sh`,
  and `git diff --check`: passed. Existing BUILD source list remains unchanged.

No Bazel execution, full-repository suite, engine, installed binary or shared-server
qualification is claimed. Parent retains final review/publication authority. The
package still has no runtime callers.

## Historical Memory V1 reconciliation at implementation 54be1474f

This normal follow-up preserves implementation
`d22db0cd661196f4bf85450556f469de220ebebd` and pins public `gastownhall/bdp`
`c064cc33274d1a88532ef6d135d7446b9b64b039`. All copied bytes were fetched from
that exact public commit and compared before implementation. There are now ten
pinned files: eight property schemas, cases and the nonnormative definition graph.
The new Requirement schema and every SHA256 are in the provenance manifest.

The product is now titled Memory V1. This does not claim that F0 is shipped or
adopt a Type-definition representation. The product endpoint/pin table was
independently checked and is unchanged: local Issue/Memory family pairs only,
exact targets required only for Supports/Challenges, Supersedes current or exact.
The exploratory graph has thirteen definitions and eleven declarations; the new
Requirement-to-Memory declaration is cross-checked as candidate evidence.

Selected changes implemented:

- Memory and Association now each declare optional nonblank open-string status,
  inherited by their supplied specializations. All seven Bead kinds and six Link
  kinds accept arbitrary nonblank values verbatim and refuse null/blank/non-text
  values. No enum, default, normalization or Issue-workflow admission is used.
  There are no runtime callers or workflow effects implemented by this package.
- Requirement is the seventh Memory kind, with the common title/body/status
  contract and no new required primary field. Membership and endpoint admission
  now use explicit family predicates rather than incidental enum ranges.
- Applicability validation is removed entirely. Any JSON applicability value is
  uninterpreted open-root extension data; tests preserve scalars, nulls, arrays,
  objects, and former invalid interval/description shapes. No field is silently
  promoted into a metadata schema or provenance representation.
- Existing Issue decision and Memory Decision remain separate classifications.
  A kind/status property cannot select or infer either contract. Type-only
  semantic revision behavior remains the caller's explicit obligation.

The old closed-applicability cases were retired or moved to still-relevant
Policy.authority diagnostics. Previous format-edge fixes remain for observedAt
and authority. The schema oracle includes Requirement, and every common Link
property case still executes against all six Link kinds.

Current verification:

- `internal/memorytypes/check.sh -race -cover`: passed, 94.3% Go statement coverage,
  no race report. Offline oracle passed 279 cases (102 source + 177 local), including
  open status positive/negative cases across every supplied Bead and Link kind.
- Product pair tests cover 726 combinations (eleven endpoint classifications over
  six Link kinds), independently checking the product table and only subset
  compatibility with the broader nonnormative BDP exploration.
- Exact source default checker run under the package's pinned Python environment
  from `/tmp/trish-f0-memory-v1-source-check`: passed eight schemas, 102 cases,
  ten Bead and eight local Link excerpts. No shared BDP inputs were required.
  Optional exploration mode was not run or claimed as product qualification.
- Final `make ci-pr-lint`: passed zero issues on native, Windows non-CGO and Darwin
  non-CGO. An initial unused-argument finding after applicability removal was fixed
  by removing the now-constant property path argument from the private helper.
- `python3 tools/bazel/go_srcs.py --check`, `bash -n internal/memorytypes/check.sh`,
  formatting and `git diff --check`: passed. The new schema is testdata, so no Go
  source/BUILD inventory change is required.

No new Go dependency, engine call, shared database access, full-repository suite,
Bazel execution, CLI/storage integration, publication or release action is included.
Final review and publication remain with the parent.

## Final review closure following implementation 54be1474f

Native final review reported no findings. Claude's final review of
`54be1474f66c5f0a010a063208e133e2054a10e9` found no production-verdict defects;
it identified regression gaps and diagnostic/evidence limitations. Its proposed
mutation-survival scenarios were analytical rather than executed mutations, and
its Go-side format probing was source reasoning rather than differential fuzzing.
The new cases below were actually executed against Go and the independent oracle.

This closure changes tests and documentation only. Production validators and
candidate pin `c064cc33274d1a88532ef6d135d7446b9b64b039` remain unchanged.
Parent reports later candidate `aeda0b9cb21eb1ade18eceb8e389359da58b129d` changes
three Markdown files only, with JSON/checker inputs unchanged; its CLI/query follow-up
is outside F0. No input pin was advanced merely to follow prose edits.

| Finding | Closure |
|---|---|
| 1. Cross-class test did not reach identity comparison | Renamed/reframed it as local self-identity plus independent external refusal. Assert exact diagnostic paths/expectations for each. External categories still refuse first; production diagnostics were not reordered. README and the older review evidence now explicitly distinguish the caller's canonical-identity invariant from the branch exercised. |
| 2. Matrix omitted floating targets | All 1452 combinations now cover both pin modes. Only Supports/Challenges require exact targets, including when both endpoint roles otherwise match. AppliesTo current-target acceptance is covered. |
| 3. Authority-object URI validation untested | Added missing, non-string and relative uri cases with `/properties/authority/uri` diagnostic assertions. |
| 4. URI query/userinfo paths untested | Added ordinary valid query/userinfo cases and malformed opening/closing bracket negatives, checked by Go and the pinned schema oracle. |
| 5. Year/month/day zero guards untested | Added all three invalid observedAt cases with field-path assertions. |
| 6. Malformed Link payloads untested | The malformed JSON/UTF-8/null/array cases now exercise every Link property entrypoint as well as Memory. |
| 7. Error text untested | Added one exact `Violation.Error()` assertion for a nonblank title refusal. No nil-receiver behavior is promised. |
| 8. Missing versus wrong-type diagnostic | Deferred nonblocking polish. Both already produce a typed refusal at the correct field; no missing-field code or distinct expectation is promised. Avoid changing production diagnostics solely for this closure. |
| 9. Manifest could omit an input | Added the complete expected ten-file set, unique-entry checks, exact source-path checks, and directory/manifest set equality, retaining every hash check. Missing, duplicate and unlisted copied inputs now fail. |
| 10. Requirement/negative diagnostic gaps | Added Requirement null-body refusal and filled missing Requirement, authority, and whitespace diagnostic paths. |
| 11. Unpaired surrogate acceptance | Recorded as a schema-level limitation for the parent/spec owner. Both current schema checking and Go decoding accept the escape; no stricter policy, normalization or new generic JSON validation was invented. |

Executed closure checks:

- `internal/memorytypes/check.sh -race -cover`: passed; 98.4% Go statement coverage,
  no race report. Offline oracle passed 290 property cases (102 source + 188 local).
  The product endpoint matrix covers 1452 combinations. This is direct execution,
  not an assertion that hypothetical reviewer mutations were applied.
- `make ci-pr-lint`: passed zero issues on native, Windows non-CGO and Darwin
  non-CGO; formatting passed.
- `python3 tools/bazel/go_srcs.py --check`, `bash -n internal/memorytypes/check.sh`,
  and `git diff --check`: passed. Go source inventory and BUILD metadata are unchanged.

The upstream default-example check remains evidence from the earlier exact-pin run
above; it was not rerun or relabelled as new closure evidence. No full-repository,
engine, shared database, installed runtime or Bazel qualification was performed.
No publishing or runtime activation is included.
