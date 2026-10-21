# Transactional matching no-op evidence — native code review

## Finding

**1. P2 — Installed positive confuses the client-action ledger with the complete exchange ledger.** `apps/bdptest/src/transactional-matrix.test.ts:440–448` requires fifteen `noop.actions`, pairs those entries by numeric index with `noop.exchanges`, and expects the final action to be raw `n-changed-groups`. The runner sends HTTP actions directly to `runRequest` and returns (`packages/conformance/src/runner.ts:854–881`); only programmatic actions enter `actions` through `runProgrammaticAction` (`:1013–1030`). This scenario therefore has fifteen exchanges but only three recorded client actions: `n-seed-facts`, `n-noop-facts`, `n-changed-facts`. Even correcting only the length would leave the index comparison and final action expectation wrong.

**Impact:** the valid installed positive necessarily fails its outer bookkeeping test, preventing the required frozen-source evidence. The author's actual `transaction-noop-conformance/packed-positive.log` confirms exit1 at line440, expected15/actual3, after the row pass checks and fifteen-exchange assertion. This is a test-ledger bug, not an observed runtime/semantic failure.

**Bounded fix:** independently pin the three client action IDs and their successful projection assertions; independently pin all fifteen exchange IDs/order and successful exchange assertions, matching each client to its exchange by ID rather than shared numeric index. Require the final raw exchange `n-changed-groups` plus the overall row pass: successful row completion includes its post-assertion scoped URL capture validation. Preserve fifteen actions in the authored manifest,159 complete-cohort exchanges and all existing limits. Keep this failed attempt as failed history; fresh successor positive/control evidence must not be attributed to32711.

No other actionable finding found in the complete four-file delta. Repair1 is required before accepting installed proof; this report does not declare readiness or qualification.

## Exact scope

Reviewed clean source `/Users/dbox/repos/wt/vickie/vickie-noop-conformance` HEAD `32711dc30598c3b30947fbb54dd229e110eb4f82`, tree `c8938a9a872b19537b89d611541a0ffc80969296`, direct base `ea584103f0dc90ae365970bb672f24bedf318e42`. HEAD and clean status were checked at both start and end. Diff is exactly manifest, coverage, cumulative conformance test and app Transactional matrix test:3418 insertions/20 deletions. Runtime, fixture, schema, catalog, runner/DSL, public gates and sealed evidence are unchanged.

Reviewed accepted v3 plan, measured amendment/root acceptance, source-artifact audit, full test diffs and artifact content. Artifacts independently match the previously audited complete candidate except the two required lowercase assertion-ID corrections (`a/l-metadataChange` to `a/l-metadata-change`). No assertion semantics changed in that correction.

## Checks of implementation

- **Law and honest scope:** exact two catalog selected clauses are pinned in `transactional-cumulative.test.ts:1668–1690`. The new row stays partial and discloses unobserved failed/no-effect categories, owned effects, guards, policy and retention cases.239 required plus five optional missing rows remain explicit. No whole-contract or profile promotion.
- **Actual independent oracles:** seed and changed records have literal identities/Types/properties/metadata/endpoints before capture. No-op POST and retained receipt compare complete prior records and omit effectPosition; current Bead collection and Link GETs independently compare full records. Event pages require five then seven facts, incidence endpoint payloads, changed/new revisions, transaction and exact property/metadata patches. Group reads require one seed group, previous pos-0, eventCount5, exact three upserts/full records, seed transaction and current head. The final real effectful comparator requires advanced head plus a non-null scoped continuation through the required unused capture. These match the accepted unconditional, unowned partition.
- **Capture/context:** raw captures remain URLs; complete Resource/revision captures use validated receipt clients. All referenced captures are prior-established. Context presence precedes metadata equality; comparisons refer to same-status200 raw responses. Existing `captureValue` executes after assertions and rejects null/non-string/out-of-Scope URLs; GET capture cannot grant receipt authority. The positive's intended row-pass check is the right completion proof once finding1's ledger bookkeeping is fixed.
- **Static protection:** `assertMatchingNoopObservation` independently constructs the complete expected assertion arrays and capture lists; it does not derive expected values from candidate assertions. It also pins request semantics, keys, literal GET targets and credentials. Four weakening controls remove full-record/no-effect/context evidence or change the group tuple and require the matching expected-content failure. Coverage mirrors the exhaustive223 IDs. Prior ten manifest and ten coverage raw spans are guarded.
- **Actual counterfactual design:** three controls use the existing manifest seam, changing only Bead property input, Link metadata input, or the two group/head expectations. They require actual complete wire,200, valid response schema, named intended failures and stop at the intended last exchange. The property/metadata controls permit additional legitimate effectful failures; the group control requires exactly its two expected assertion failures. No synthetic target response or parser refusal is used. Controls remain unexecuted in the available logs at review time.
- **Routes and budgets:** the new finite route is independently pinned seventh/JS6, between beads/new and beads/b. The canonical preimage independently hashes to `906199051adc043408d64c0879a3d3f2cce32b06faa08b9458f4a270ce31a6ce`, matching the fixed declaration. The app's live2048/32768 wire-bound loop now includes this row, including receipt exchanges. Measured artifact guards and exact223/1539 assertions,15/166 actions are implemented as accepted. No parser/time/child limits increased.

Independent raw-span parsing confirmed all twenty old objects and all surrounding original bytes: removing only the inserted new object reconstructs each entire parent artifact exactly. Source artifact measurements:

|Artifact|Bytes|SHA256|
|---|---:|---|
|Manifest|488280|`0358cbf7416321bb911e990259236f89bc2720cbc29d037581f542080d9f32ab`|
|Coverage|70662|`8a770b4371689a3d9d5f84ecfc7a1d9b72715c81e617cf0ea5ae04e55c66098e`|

The two-byte ID growth leaves4094 manifest bytes/1022 coverage bytes under measured guards; nodes remain13019 with64 headroom. The source audit records unchanged716-byte fixture/hash. These measurements do not imply successful execution.

## Observed validation and limits

I ran no tests, builds, installs, servers or network calls and changed no source. Review used local Git/text inspection and independent JSON/raw-span/hash arithmetic only; those commands exited0. The sole written output is this report; no owned processes.

Author artifacts actually read: corrected cumulative JSON/log reports34 passes; app-static JSON/log reports1 pass/52 skipped; exits.json records successful type/build/format/lint/boundaries/Read-seal and frozen clean/type/build, along with the earlier failing cumulative attempt (invalid uppercase IDs) and the current failing packed positive. The first packed positive reports0 pass/1 failure/52 skipped and must remain failed. Although it reached row-pass assertions, it is not a completed passing cohort. No success claim is made for pending controls, final package cleanup report, external council, CI or merged-tree verification. Full code council remains required after the bounded successor.
