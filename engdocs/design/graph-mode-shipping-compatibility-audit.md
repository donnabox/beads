# Graph-mode shipping CLI compatibility audit

Status: working audit, October 20, 2026. Compare the shipping `v1.3.1` tag
(`c1c4b642a`) with Preview 3 integration `dd6b9359b`. Recheck the baseline
when the shipping release changes. This document records observed behavior;
the graph CLI specification remains the contract for intentional differences.

## Objective and method

An existing Issue command that is admitted by `--graph-mode link` should keep
its ordinary argument, flag, output and failure behavior unless the graph
contract explicitly changes it. Compare actual command execution on disposable
ordinary and graph workspaces, in separate `bd` processes. Test representative
combinations of filters, mutations, JSON and human output, rather than only
the registered help. Run embedded and shared-server Dolt where the behavior
depends on storage. A route that refuses an unsupported option explicitly is
safer than one that silently ignores it, but still counts as a parity gap.

Classify every discrepancy as preserved, intentional graph-specific behavior,
missing capability, or regression. A missing capability needs a documented
owner and test before it is treated as release-ready. Preserve the ordinary
`bd dep` facade as a compatibility route; canonical Link work stays under
`bd link`. The future generic query language is outside this audit until its
design is settled.

## First-pass findings

| Surface | Shipping behavior | Preview 3 behavior at baseline | Disposition |
| --- | --- | --- | --- |
| `blocked --parent ID` | Filters to a parent's children using native parent-child edges and dotted-ID fallback. | Refuses `--parent` despite using the same native blocked predicate. | First repair in this audit: resolve the graph parent in the blocked read transaction and pass its native ID to the predicate. Test explicit paths, bare IDs and canonical URLs. |
| `ready --parent ID` | Parent-scoped readiness. | Not in graph-mode flag admission. | Missing capability; audit alongside claim atomicity before changing. |
| `q --parent ID` | Quick capture under a parent. | Explicitly refuses; `create --parent` is the current graph path. | Missing convenience behavior; check Issue identity and title defaults before routing. |
| `update ISSUE` text sources | Ordinary command accepts stdin/file sources for supported text edits. | Graph route currently accepts inline text only. | Missing input behavior; preserve source precedence and mixed-field atomicity. |
| `create --file` Markdown dependencies | Ordinary Markdown import can create Issue relationships. | Graph route refuses templates requiring mapped Links. | Missing capability; requires atomic Issue/Link mapping, not a CLI-only flag change. |
| `list --limit` and `query --limit` | A page of Issues, default 50 unless context changes it; explicit zero means unlimited. `--offset` is proxied-server-only. JSON is a bare Issue array. | Legacy graph Issue JSON keeps the array shape. Canonical collection output has a separate page contract and refuses a truncated selection until retained continuation exists. | Preserve legacy output for the admitted Issue route. Keep canonical collection paging as separate provider work. |

The first `blocked --parent` repair passed installed-process tests for an
embedded graph workspace, including a graph/native ID mismatch and an
ambiguous bare ID; the pre-existing blocked/query journey also passed.
Focused graphstore blocked lifecycle and admission tests passed. The broad
macOS CLI suite reported an unrelated failure in
`TestRunTestsAndSweepReapsOrphanedServer`; that run was stopped before the
long graphstore suite completed. These results do not qualify the shared-server
route or the remaining rows above.

The Preview graph CLI specification has an older pinned-commit implementation
ledger and a larger target inventory. Its target rows are not evidence that
the current binary already supports an option. Continue this audit against
the actual built commands and refresh the ledger after each qualified slice.

## Next comparison batches

1. Read-only Issue discovery: `list`, `show`, `ready`, `blocked`, `search`,
   `count`, `query`, `children`, `graph`, including default filters, order,
   limits and both output modes. Do not redesign the query language in this
   batch.
2. Single-Issue writes: `create`, `q`, `update`, `close`, `reopen`, `defer`,
   `undefer`, `delete`, `assign`, `priority`, `note`, with omitted and explicit
   flags, guard behavior and failure atomicity.
3. Existing Dependency and workflow facades: `dep`, `links`, `ready --claim`,
   `todo`, `state`, `merge-slot`, and import/export where admitted. Keep Link
   canonical behavior and ordinary Issue JSON compatibility distinct.

For each batch, add an installed-command regression test only where it covers
a real process, persistence or output boundary that existing tests miss. Record
the exact Preview commit and actual validation result before closing a row.
