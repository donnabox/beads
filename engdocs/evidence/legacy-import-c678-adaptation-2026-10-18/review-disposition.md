# c6 activation council: source reconciliation and repair routing

Frozen source c6e94fd5037eff7e855d542b034745d1285463f1 remains unchanged.
Actual direct self-contained review completed exit0, empty stderr, 07:09:22 UTC.
The initial900-second timeout remains a failed attempt, not a verdict. The actual
report raised twelve findings, often explicitly dependent on omitted source.
Native source reconciliation overturns the reported High blockers and identifies
one genuine Medium runfiles defect. No build/test/DB has run.

## Source-resolved findings

1. Prefix: utils.ExtractIssuePrefix deliberately returns web for all-letter
   web-app-abcdef and web-app for a1b/b2c/d4e. The resolver chooses the longest
   candidate for the full batch, while the isolated all-letter case fails its
   confidence check. Both existing expectations are compatible.
2. Comments: Store.IssueComments explicitly sets IssueID=path, so the graph
   feed correctly expects beads/web-app-b2c. Native storage retains web-app-b2c.
   Comment has exactly the five expected fields; outputJSON returns a bare array
   under the fixed subprocess environment. The empty-list equivalence does not
   bypass a nonempty source comment.
3. Null: strictDecode calls exactFields before ordinary decoding; a scalar string
   null is refused with `null is unsupported for legacy string`. Empty string is
   independently valid and present in the corpus.
5. Exact-field comparison: this corpus already supplies status/type/UTC stamps
   and closed-at. SetDefaults preserves P0; PrepareIssueForInsert adds no carried
   defaults here, only validates/normalizes already-UTC dates and internal hash.
   commonMetadata uses the same graph.CanonicalizeJSON and empty-object rule.
   Keep the independent preservation oracle rather than calling the importer
   under test to construct expected output.
9. Runfiles: root docsync_files DOES glob engdocs/**, but the actual docsync
   go_test consumes repo_doc_files, not docsync_files. repo_doc_files supplies
   Markdown, not the linked JSON/JSONL/manifest files. Add //:docsync_files to the
   actual test data. This is a confirmed source-wiring defect; no Bazel run claims.
10. Producer create uses future --defer to select deferred status. Empty Memory
    value was genuinely exported after config set kv.memory.empty-value empty;
    remember itself rejects empty values. Document that distinction explicitly.
11. The prepared patch SHA44ea61ceec23a3e7fb1d0e99f8e71291669d708d5dd1cd021a0b9e5f8f23b40f
    is the reviewed patch and exactly matches activation hunks/tail. Fixture bytes
    end with LF. These previously omitted facts do not require source repairs.

## Proposed test-only repairs

The outside-checkout proposed.patch changes only the import workflow and actual
Bazel docsync test data. It pins diagnostic reason/exit/code/retryability/stdout,
adds explicit JSONL record separators, fixes repeat calls to specify source
prefix and require empty-target refusal, compares complete reordered previews,
checks changed exported count projections are ignored, fails nil Issue properties
clearly, checks exact twenty-Bead identity inventory with no truncation and samples
Issue/Memory/Link version stability on repeat. No producer bytes, SDK or product
behavior changes. Proposed Go has only source-format checking, not compilation
or execution. Native draft audit completed with no actionable findings and verified all proposed-file hashes; actual repair/adaptation review remains required.

Full database/config/history rollback including dry-run prefix non-adoption is
already authored at the lower store seam in legacyImportState/AtomicityAndPersistence/
FreshDataPlanes, both-engine UNRUN; the CLI has no graph-mode config read route.
Do not add a product feature merely to expose prefix for a test. The current CLI
inventory is a Bead projection, not Link/orphan/catalog completeness. Proposed
comments/docs narrow that claim and retain mandatory lower-matrix qualification.
Server t.Skip remains the existing source behavior; owner qualification must
require both named modes and zero required skips. Missing server proof is no PASS.

## Routing needed from Janet

Janet6246770795 says c6 stays immutable and a separate c678 child follows clear c6
review. A real runfiles repair and useful test-oracle changes are needed, so c6
cannot itself meet the no-unresolved-actionable finding gate without violating
immutability. Please explicitly allow these test-only repairs within a new clean
private c678 adaptation, preserving c6. The adapted exact source/tree/delta then
gets the requested council BEFORE Janet-owned installed qualification. No Iris
build/test/engine grant, PR86 retarget, shared Git/core/staging/production write,
main/default/release action is requested. Until routing, no adaptation checkout.
