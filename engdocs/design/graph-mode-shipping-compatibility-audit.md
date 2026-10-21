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
| `list --limit` and `query --limit` | A page of Issues, default 50 unless context changes it; explicit zero means unlimited. `--offset` is proxied-server-only. JSON is a bare Issue array. | Legacy graph Issue JSON keeps the array shape. Canonical collection output has a separate page contract and refuses a truncated selection until retained continuation exists. | Preserve the shipping limit and output behavior for admitted Issue routes. Canonical continuation work is deferred for Preview 3; do not invent a cursor. Generic query design awaits Melinda. |

The first `blocked --parent` repair passed installed-process tests for an
embedded graph workspace, including a graph/native ID mismatch and an
ambiguous bare ID; the pre-existing blocked/query journey also passed.
Focused graphstore blocked lifecycle and admission tests passed. The broad
macOS CLI suite reported an unrelated failure in
`TestRunTestsAndSweepReapsOrphanedServer`; that run was stopped before the
long graphstore suite completed. These results do not qualify the shared-server
route or the remaining rows above.

## First-gate implementation progress

The first implementation gate is `create`, `update`, `delete`, `show`, `list`,
`versions`, their `bd link` variants, and shipping `bd link FROM SOURCE` in
graph mode. As of October 20, the following fork PRs are open and unmerged.
Their tests qualify their stated slices, not the whole gate:

| PR | Implemented slice | Current evidence |
| --- | --- | --- |
| [#150](https://github.com/donnabox/beads/pull/150) | `blocked --parent`; also records this working audit. | Six checks passed; shared-server route remains unqualified. |
| [#151](https://github.com/donnabox/beads/pull/151) | Familiar Issue `show`, including multi-ID, `--current`, refs, children and watch; shipping positional Link feedback. | Six checks passed; selected installed workflows passed embedded and shared-server Dolt. |
| [#152](https://github.com/donnabox/beads/pull/152) | Issue create/update description file and stdin sources. | Seven checks passed, two skipped; the new workflow passed both engines. |
| [#156](https://github.com/donnabox/beads/pull/156) | `list --format=json` and `--tree=false` aliases. | Four checks passed; focused installed workflows passed both engines. |
| [#158](https://github.com/donnabox/beads/pull/158) | Issue list sort by visible ID, assignee or closed time; detailed `--long` output. | Installed workflow passed both engines at `089b595ac`; six current-head checks passed. Stacked on #156. |
| [#159](https://github.com/donnabox/beads/pull/159) | Shipping Issue `create --silent` ID output with `--json` precedence. | Shipping v1.3.1 behavior verified; installed create workflows passed both engines; CI running. Stacked on #141. |
| [#160](https://github.com/donnabox/beads/pull/160) | Read-only `ready --parent` with checked Issue selector resolution. | Installed ready/claim workflows passed both engines; CI running. `--claim --parent` still refuses pending atomic selector resolution. Stacked on #141. |
| [#161](https://github.com/donnabox/beads/pull/161) | Shipping custom native Dependency kinds in positional `bd link`, including literal `depends-on`/`blocked-by`. | Installed positional Link and neighboring workflows passed both engines; CI running. Stacked on #151. |

The shared base [#141](https://github.com/donnabox/beads/pull/141) remains a
draft. Its current head includes the C0 test-sharding repair from #157; broad
CI found one verifier failure: the test still expected three CLI groups after
the runner changed to twenty. [#162](https://github.com/donnabox/beads/pull/162)
updates that test, with all 53 local verifier tests passing. The real-engine
C0 qualifier has not passed on the current base head. Further first-gate gaps
include atomic `create --deps`,
update label/status and multi-ID forms, multi-ID and cascading deletion,
shipping list filters and tree/pretty output, and partial-ID behavior in the
positional Link form. Those remain inventory items until an implementation and
both-engine evidence close them. The next gate is `dep`, `ready`, and logical
aliases; the remaining command inventory follows that. Generic query design
waits for Melinda, and pagination implementation is deferred for Preview 3.

The three parallel audit shards identified `bd dep list` as the largest
remaining second-gate gap: graph mode lists only blocking Dependencies and
refuses JSON, while shipping lists other native kinds and has different
single-anchor and batch JSON shapes. Read-only `ready --parent` is covered by
#160; parent-scoped atomic claim still needs a same-transaction selector.

The Preview graph CLI specification has an older pinned-commit implementation
ledger and a larger target inventory. Its target rows are not evidence that
the current binary already supports an option. Continue this audit against
the actual built commands and refresh the ledger after each qualified slice.

## Parallel parity sweep (partial)

The following findings compare shipping `v1.3.1` with Preview 3 integration
`dd6b9359b` and, where noted, an installed binary built from this PR at
`0f32273e7`. The discovery, write and relationship sweeps stopped before an
exhaustive flag matrix or final regression suite. Code-backed observations are
distinguished from disposable-workspace execution below.

| Priority | Route | Verified discrepancy | Next bounded decision or repair |
| --- | --- | --- | --- |
| P1 | `defer`/`undefer ID...` | Embedded graph `defer second missing --json` changed `second`, reported the missing ID on stderr, and exited 0; shipping direct-Dolt `defer` exited 1 without changing `second`. Graph source and an existing workflow test explicitly select the ordinary **proxied** route's partial-success behavior, which also exits 0. | Decide which shipping backend contract graph mode should follow before editing. In either case, a warning plus exit 0 must be clearly documented and tested for both verbs. Do not silently treat this as an accidental implementation bug. |
| P1 | `search --type feat` | Installed graph smoke returned zero while `--type feature` returned one. Shipping normalizes the `feat` alias. Graph `runGraphPreviewSearch` passes the raw Type; this silently hides a matching Issue. | Normalize the Issue Type using the shipping helper and test both spellings. |
| P1 | `show ID` human output | Installed graph smoke began with a raw JSON object; shipping renders an Issue detail document. Current-Issue `show --json` retains its legacy array shape in an existing graph test. | Restore Issue detail presentation for human output while retaining the canonical and legacy JSON contracts. |
| P1 | `ready --parent ID` | Shipping admits parent-scoped readiness; installed graph smoke refused `--parent` as `capability_unavailable`. This PR restores the same parent selector for `blocked`. | Resolve the parent in the ready read/claim transaction; test eligibility and atomic claim under the parent. |
| P1 | `dep list` | Shipping lists all dependency kinds and has distinct single-anchor and batch `--json` shapes. Graph `dep list` is blocks-only and refuses JSON; installed smoke showed a `tracks` Link under `link list --incident-to`, while `dep list` said no blocking Dependencies. | Restore the facade's kinds and output shapes without changing canonical `bd link`; test single and batch forms. |
| P1/P2 | `create --deps` and `create --file` relationships | Shipping can create dependencies with the Issue; graph refuses the combined form. Separate `create` and `dep add` are available but do not preserve one-operation failure behavior. | Design an atomic Issue-plus-Link creation route and test rollback; keep this separate from quick flag-admission repairs. |
| P2 | `q --parent`, `create --body-file`; `update` text/label/status flags | Installed graph smoke refused shipping forms including `q --parent`, `create --body-file`, `update --description=-`, `--add-label`, `--if-status`, and `--status`. | Split into small input/selector and mutation-policy slices, preserving guard and source precedence. |
| P2 | `list`, `children`, `graph` presentation | Graph explicitly refuses shipping `list --parent/--no-parent/--ready/--pretty/--tree`, `children --pretty`, and several filters. `graph ISSUE` has a simpler blocking view and refuses shipping render switches; the latter is documented as an intentional graph change. | Prioritize common Issue filters and tree usability only after confirming which output changes are intentional. |
| P2 | `dep add --file`, `dep tree`, `dep cycles` | Graph refuses shipping bulk import and options such as tree status/Mermaid output and cycle inclusion of tracks. Source and installed smoke confirm representative refusals. | Preserve one ordinary facade at a time; test nonblocking relationships where supported. |
| P2 | `delete --cascade --dry-run` | Shipping previews a cascade; graph explicitly refuses it. A disposable shipping workspace reported two Issues and one dependency for a representative cascade. | Decide graph deletion policy and preview shape before implementing any destructive cascade. |

`bd link SOURCE TARGET` and `bd dep rm` retain shipping shorthand. Neither the
shipping tag nor the tested Preview 3 binary registers a top-level `bd links`
command; the graph code contains an unregistered candidate, which is not CLI
surface. `todo` add/list/done, state read, merge-slot verbs and legacy import
are admitted. Shipping `set-state` and `export` are refused in graph mode.
These are inventory findings, not release qualification for those routes.

## Next comparison batches

1. Read-only Issue discovery: `list`, `show`, `search`, `count`, `children`,
   `graph`, including default filters, order, limits and both output modes.
   Inventory the existing `query` route where it affects parity, but defer its
   generic language and execution changes until Melinda's design lands.
2. Single-Issue writes: `create`, `q`, `update`, `close`, `reopen`, `defer`,
   `undefer`, `delete`, `assign`, `priority`, `note`, with omitted and explicit
   flags, guard behavior and failure atomicity.
3. Existing Dependency and workflow facades: `dep`, `links`, `ready`,
   `blocked`, `todo`, `state`, `merge-slot`, and import/export where admitted.
   Keep Link canonical behavior and ordinary Issue JSON compatibility distinct.

For each batch, add an installed-command regression test only where it covers
a real process, persistence or output boundary that existing tests miss. Record
the exact Preview commit and actual validation result before closing a row.
