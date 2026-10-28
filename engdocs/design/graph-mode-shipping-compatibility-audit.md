# Graph-mode shipping CLI compatibility audit

Status: working audit, October 21, 2026. Compare the shipping `v1.3.1` tag
(`c1c4b642a`) with Preview 3 integration `b6d0493ba`. Recheck the baseline
when the shipping release changes. This document records observed behavior;
the graph CLI specification remains the contract for intentional differences.

The [complete 43-command inventory](preview3-cli-parity.md) compares every
command in Donna's three prioritized families against the same pinned builds.
Its three detailed reports distinguish executed forms, explicit refusals,
silent wrong answers, changed contracts, candidate PRs, and untested edges.
This working audit tracks implementation and qualification progress.

Donna selected basic Bead/Link CRUD and the `ready` invariant as the first
Preview 3 compatibility slice on October 21. [The command/form table](preview3-cli-compatibility-selection.md)
marks #151, #152, #156 and the #175 readiness regression as selected unmerged
work, while the other parity PRs remain candidates. Selection does not imply
that the entire first slice is qualified or integrated.

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
| `blocked --parent ID` | Filters to a parent's children using native parent-child edges and dotted-ID fallback. | Baseline refused `--parent` despite using the same native blocked predicate. | [#150](https://github.com/donnabox/beads/pull/150) resolves the graph parent in the blocked read transaction and passes its native ID to the predicate; current-head C0 qualified both engines. Unmerged. |
| `ready --parent ID` | Parent-scoped readiness. | Baseline refused `--parent`. | [#160](https://github.com/donnabox/beads/pull/160) restores the read-only selector; atomic `--claim --parent` still refuses pending same-transaction selector resolution. |
| `q --parent ID` | Quick capture under a parent. | Explicitly refuses; `create --parent` is the current graph path. | Missing convenience behavior; check Issue identity and title defaults before routing. |
| `update ISSUE` text sources | Ordinary command accepts stdin/file sources for supported text edits. | Baseline accepted inline text only. | [#152](https://github.com/donnabox/beads/pull/152) restores description file/stdin sources on create and update, with both-engine installed tests. Other text-field forms need separate checks. |
| `create --file` Markdown dependencies | Ordinary Markdown import can create Issue relationships. | Pinned graph route refuses templates requiring mapped Links. | Unmerged #201/#202 map dependencies on existing local Issues in the batch transaction and pass focused tests on embedded and shared-server Dolt. External targets, partial IDs and references to another new heading still refuse. |
| `list --limit` and `query --limit` | A page of Issues, default 50 unless context changes it; explicit zero means unlimited. `--offset` is proxied-server-only. JSON is a bare Issue array. | Legacy graph Issue JSON keeps the array shape. Canonical collection output has a separate page contract and refuses a truncated selection until retained continuation exists. | Preserve the shipping limit and output behavior for admitted Issue routes. Canonical continuation work is deferred for Preview 3; do not invent a cursor. Generic query design awaits Melinda. |

The #150 implementation head `34c28e9d9` passed 121 hosted checks with eight
intentional skips, including the complete Graph C0 real-engine qualifier.
Its installed `blocked --parent` workflow covers embedded and shared-server
Dolt, graph/native ID mismatch, explicit paths, canonical URLs and ambiguous
bare IDs. The first macOS race attempt failed before tests because the runner
could not resolve `proxy.golang.org` while downloading `github.com/creack/pty`;
rerunning only that job passed, and the macOS coverage gate passed. This
qualifies that implementation head, not this later audit-only commit or the
other unmerged parity branches.

## First-gate implementation progress

The first implementation gate is `create`, `update`, `delete`, `show`, `list`,
`versions`, their `bd link` variants, and shipping `bd link FROM SOURCE` in
graph mode. As of October 21, the following fork PRs are open and unmerged.
Their tests qualify their stated slices, not the whole gate:

| PR | Implemented slice | Current evidence |
| --- | --- | --- |
| [#150](https://github.com/donnabox/beads/pull/150) | `blocked --parent`; also records this working audit. | Implementation head `34c28e9d9` and audit-only heads through `a265b40ee`: each passed 121 hosted checks, eight intentional skips; complete Graph C0 on real engines and macOS race coverage passed. Review remains undecided. The PR reports checks for its latest head. |
| [#151](https://github.com/donnabox/beads/pull/151) | Familiar Issue `show`, including multi-ID, `--current`, refs, children and watch; shipping positional Link feedback. | Six checks passed; selected installed workflows passed embedded and shared-server Dolt. |
| [#152](https://github.com/donnabox/beads/pull/152) | Issue create/update description file and stdin sources. | Seven checks passed, two skipped; the new workflow passed both engines. |
| [#156](https://github.com/donnabox/beads/pull/156) | `list --format=json` and `--tree=false` aliases. | Four checks passed; focused installed workflows passed both engines. |
| [#158](https://github.com/donnabox/beads/pull/158) | Issue list sort by visible ID, assignee or closed time; detailed `--long` output. | Installed workflow passed both engines at `089b595ac`; six current-head checks passed. Stacked on #156. |
| [#159](https://github.com/donnabox/beads/pull/159) | Shipping Issue `create --silent` ID output with `--json` precedence. | Shipping v1.3.1 behavior verified; installed create workflows passed both engines; six current-head checks passed. Targets Preview 3 integration. |
| [#160](https://github.com/donnabox/beads/pull/160) | Read-only `ready --parent` with checked Issue selector resolution. | Installed ready/claim workflows passed both engines; six current-head checks passed. `--claim --parent` still refuses pending atomic selector resolution. Targets Preview 3 integration. |
| [#161](https://github.com/donnabox/beads/pull/161) | Shipping custom native Dependency kinds in positional `bd link`, including literal `depends-on`/`blocked-by`. | Installed positional Link and neighboring workflows passed both engines; six current-head checks passed. Stacked on #151. |
| [#163](https://github.com/donnabox/beads/pull/163) | Shipping single-ID `delete --dry-run` preview. | Installed workflow passed embedded and shared-server Dolt; six current-head checks passed. Multi-ID and cascade behavior remain open. Stacked on #151. |
| [#164](https://github.com/donnabox/beads/pull/164) | Forced assignee replacement on Issue update. | Embedded/shared-server installed workflow and explicit `--force=false` append-notes regression passed; six current-head checks passed. Targets Preview 3 integration. |
| [#165](https://github.com/donnabox/beads/pull/165) | All native Dependency kinds in human `dep list`. | Embedded/shared-server workflows and filter-before-incident-cap regression passed; six current-head checks passed. Shipping JSON shapes remain refused. Targets Preview 3 integration. |
| [#166](https://github.com/donnabox/beads/pull/166) | `list --no-parent` selection, including exact human-row regression. | Embedded/shared-server installed workflow passed; six current-head checks passed. Stacked on #158. |
| [#169](https://github.com/donnabox/beads/pull/169) | Checked `list --parent` storage selection for parent-child edges and edge-less dotted graph IDs, plus the counted legacy projection. | Targeted storage tests and installed CLI workflow passed embedded and real shared-server Dolt; four current-head checks passed. Stacked on #166. |
| [#170](https://github.com/donnabox/beads/pull/170) | `list --parent`/`--filter-parent` for flat, `--tree=false`, JSON, long, and records output; explicitly refuses the unfinished default recursive tree. | Broad graph list tests, both-engine installed workflow, generated CLI docs, Markdown links, and six current-head checks passed. Stacked on #169. |
| [#171](https://github.com/donnabox/beads/pull/171) | Checked, same-transaction graph Issue hierarchy for recursive `list --parent`, with whole-subtree cap and shipping dotted-orphan rules. | Focused storage and installed CLI workflows passed embedded and real shared-server Dolt with no skips; four current-head checks passed, review pending. Stacked on #170. Credits Peter Chanthamynavong's shipping-tree prior art in [#7081](https://github.com/gastownhall/beads/pull/7081). |
| [#172](https://github.com/donnabox/beads/pull/172) | Default and explicit `--tree=true` human rendering of the complete graph Issue parent tree. | The prior head passed 60 focused test/subtest results on embedded and real shared-server Dolt with no skips; broad embedded list tests passed 181 results with 16 server-only skips. A later renderer fix bounds repeated diamond paths before they can expand the buffer; its focused non-CGO test and four hosted checks passed on current head `a4251effe`. Review pending. Stacked on #171. |
| [#174](https://github.com/donnabox/beads/pull/174) | Exact native Dependency kind filtering in legacy `bd link list/count`, including custom and literal `depends-on`/`blocked-by` kinds; corrects a stale `bd links` guide example. | Focused installed workflow passed embedded and real shared-server Dolt with no skips; doc sync, generated CLI docs, diff checks, and six current-head hosted checks passed. Review pending on `a6c276bd5`. Stacked on #161. |
| [#175](https://github.com/donnabox/beads/pull/175) | Installed regression that `ready` and atomic `ready --claim` ignore informational `types/references` Links in both directions between an Issue and Memory Bead. | New workflow passed embedded and real shared-server Dolt without skipped subtests at `c7363ec73`; hosted Bazel CI is pending. Directly targets Preview 3 integration. |
| [#197](https://github.com/donnabox/beads/pull/197)–[#198](https://github.com/donnabox/beads/pull/198) | Atomic local `create --deps` storage and CLI; external targets explicitly refused. | Forward/reverse Link and rollback storage tests plus installed CLI workflow passed embedded and real shared-server Dolt; hosted historical checks passed. Unmerged candidates. |
| [#199](https://github.com/donnabox/beads/pull/199)–[#200](https://github.com/donnabox/beads/pull/200) | Atomic native child creation with `--parent --deps`. | Focused child/dependency/rollback storage and installed CLI workflows passed both engines; #199 four and #200 six current-head hosted checks passed. Unmerged candidates. |
| [#201](https://github.com/donnabox/beads/pull/201)–[#202](https://github.com/donnabox/beads/pull/202) | Markdown batch dependencies on existing local Issues, with native edges, canonical Links and source-owned History. | Fault-injected batch storage and installed Markdown CLI workflow passed embedded and real shared-server Dolt; a local combined checkout passed seven selected create/storage workflow roots on both engines. Its corrected broad embedded `TestGraphPreview` run passed 274 top-level tests with one server-only skip (1,265 pass and 81 skip events including subtests); this is not full shared-server qualification. #201 has four and #202's current head `de012f5bb` has six hosted checks green; reviews pending. Parser changes in upstream contributor [#7512](https://github.com/gastownhall/beads/pull/7512) remain separate. |
| [#203](https://github.com/donnabox/beads/pull/203) | Shipping `bd list --ready --parent` transitive descendant selection, resolved and queried in one checked graph read; ready output remains flat. | Focused graphstore and installed CLI tests passed embedded and real shared-server Dolt, including a ready grandchild through a closed intermediary, legacy JSON, records, alias and human output. Docs checks passed; hosted CI and review pending. Stacked on #190, which still refuses the pair on its own. |

The shared base [#141](https://github.com/donnabox/beads/pull/141) merged into
`codex/preview3-integration` at `b6d0493ba` on October 21. Its head
`8baf79d43` includes the twenty-group C0 verifier correction
directly; redundant #162 was closed. Its next hosted C0 run reached Memory
deletion and exposed a stale three-field assertion: the product correctly
returned the exact final live predecessor plus a distinct terminal deletion
revision. The shared smoke/verifier fix was prepared as #167, fast-forwarded
onto #141 and combined with stronger negative controls; GitHub automatically
marked #167 merged. All 57 local verifier tests pass on the combined head.
Current-head full real-engine C0 qualification and both race tiers passed:
119 hosted checks succeeded, with ten intentional skips. Further first-gate gaps at the pinned baseline include atomic `create --deps`,
update label/status and multi-ID forms, multi-ID and cascading deletion,
shipping list filters and other tree/pretty output, and partial-ID behavior in
the positional Link form. Those remain inventory items until an implementation
and both-engine evidence close them. The default recursive `list --parent`
view is implemented in unmerged #171/#172 with current-head hosted checks
passing. The next
gate is `dep`, `ready`, and logical
aliases; the remaining command inventory follows that. Generic query design
waits for Melinda, and pagination implementation is deferred for Preview 3.

The three parallel audit shards identified `bd dep list` as the largest
remaining second-gate gap. #165 now covers all native kinds in human output,
but graph mode still refuses the shipping single-anchor and batch JSON shapes.
Read-only `ready --parent` is covered by #160; parent-scoped atomic claim still
needs a same-transaction selector.

A disposable integration checkout at `6bee01c` combined #150, #151, #152,
#156, #158–#161 and #163–#166 on the earlier #141 head `afe5115fc` without
pushing or merging any PR. Focused installed CLI workflows and the graphstore
incident-cap regression passed on embedded and real shared-server Dolt with no
skips; two textual merge conflicts were resolved to retain both intents.
That local result does not qualify #141's newer head or the full first gate.

The same disposable checkout was refreshed to `3fad828cf`, combining #141's
qualified head `8baf79d43` and stacked #169/#170 with the earlier parity
slices. Focused graph Issue list storage and installed CLI tests passed in two
packages on embedded Dolt (13 server subtests skipped without a server).
With a disposable real shared-server Dolt, the installed no-parent/parent and
sort/cap workflows plus the counted legacy projection passed both engines with
no skips. This is selected integration evidence, not a full first-gate or
whole-repository test run; the checkout was neither pushed nor merged.

The checkout advanced locally to `bb4917613` by cherry-picking #171/#172 on
top of those parity slices. Their commits applied without conflicts. Focused
installed parent-tree, sort/cap and graph-store hierarchy tests passed with a
disposable real shared-server Dolt: 17 test/subtest results, no failures or
skips. The server and temporary database were removed afterward. This selected
combined run still does not qualify the full first gate; no integration PR was
pushed or merged.

For Donna's selected first slice, a separate disposable checkout at
`bc7a9d9ba6` cleanly merged the qualified integration base `b6d0493ba` with
the unmerged heads of #151 (`e07944fab`), #152 (`f0e2c6cbf`), #156
(`e155c3ac5`) and the test-only #175 (`c7363ec73`). Eighteen focused
installed CLI workflows passed with `go test -tags gms_pure_go ./cmd/bd`
on embedded Dolt. The same selection then passed with a disposable real
shared-server Dolt: all 18 top-level workflows passed, with no failures or
skips. The selection includes canonical show/collection, Issue and Memory
authoring/deletion, Issue show/watch/list, Link create/update/show/list/delete
and retained versions, and `ready`/atomic claim with informational Memory
Links. The server and its temporary data were removed afterward. This is
selected-slice integration evidence, not a whole-repository or full first-gate
qualification; no selected PR was merged by this local test.

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
| P1 | `show ID` human output | Baseline graph smoke began with a raw JSON object; shipping renders an Issue detail document. Current-Issue `show --json` retains its legacy array shape. | #151 restores familiar human details, multi-ID, current, refs, children and watch, with both-engine installed tests. Other Issue show forms remain to audit. |
| P1 | `ready --parent ID` | Shipping admits parent-scoped readiness; baseline graph smoke refused `--parent` as `capability_unavailable`. | #160 restores read-only parent selection with both-engine tests. Atomic `--claim --parent` remains a separate gap. |
| P1 | `dep list` | Shipping lists all dependency kinds and has distinct single-anchor and batch `--json` shapes. Baseline graph `dep list` was blocks-only and refused JSON. | #165 restores all native kinds in human output. The facade's single-anchor and batch JSON shapes remain refused; do not invent them. |
| P1 | Positional `bd link BLOCKED BLOCKER` partial IDs | Shipping resolves exact IDs, configured-prefix spellings, then a unique leading hash abbreviation. Graph mode currently maps each bare operand to a literal `beads/` path; #161 preserves native kinds but does not resolve partial IDs. | Resolve both selectors inside the Link write transaction before mutation, preserving explicit graph paths, exact-match precedence, ambiguity refusal, and source revision guards. A prior CLI read followed by a path write would introduce a race. |
| P1/P2 | `create --deps` and `create --file` relationships | Shipping can create dependencies with the Issue; pinned graph mode refuses the combined form. Separate `create` and `dep add` do not preserve one-operation failure behavior. | Unmerged #197/#198 create a new Issue, local native edges, canonical Links, and source-owned History in one transaction; #199/#200 extend the same transaction to `--parent --deps`; #201/#202 admit Markdown dependencies on existing local Issues, preserving batch rollback. Both engines passed focused child/dependency/Markdown tests. Donna selected explicit `external:` refusal. Partial target IDs, same-batch target references and graph-JSON batch relations remain gaps. |
| P2 | `q --parent`, `create --body-file`; `update` text/label/status flags | Baseline graph smoke refused shipping forms including `q --parent`, `create --body-file`, `update --description=-`, `--add-label`, `--if-status`, and `--status`. | #152 restores description file/stdin on create/update; #164 covers forced assignee replacement. `q --parent`, label/status forms and other text fields remain. |
| P2 | `list`, `children`, `graph` presentation | Baseline graph refused shipping `list --parent/--no-parent/--ready/--pretty/--tree`, `children --pretty`, and several filters. `graph ISSUE` has a simpler blocking view and refuses shipping render switches; the latter is documented as an intentional graph change. | #156 restores format/tree aliases, #158 sort/long, #166 no-parent, #169/#170 flat/JSON parent selection, #171/#172 the default recursive parent tree, #189/#190 blocker-aware ready lists, and #203 transitive ready-parent selection. All are unmerged. Parent-tree `--quiet`, `--deps`, and `--watch`, other list filters, and tree/pretty output remain. The graph tree follows the subtree cap, cycle and dotted-orphan behavior established by upstream contributor [#7081](https://github.com/gastownhall/beads/pull/7081), which remains open for the shipping route. |
| P2 | `dep add --file`, `dep tree`, `dep cycles` | Graph refuses shipping bulk import and options such as tree status/Mermaid output and cycle inclusion of tracks. Source and installed smoke confirm representative refusals. | Preserve one ordinary facade at a time; test nonblocking relationships where supported. |
| P2 | `delete --cascade --dry-run` | Shipping previews a cascade; graph explicitly refuses it. A disposable shipping workspace reported two Issues and one dependency for a representative cascade. | #163 restores single-ID `--dry-run`; cascading and multi-ID preview still need a policy and shape before implementation. |

`bd link SOURCE TARGET` and `bd dep rm` retain shipping shorthand. Neither the
shipping tag nor the tested Preview 3 binary registers a top-level `bd links`
command; the graph code contains an unregistered candidate, which is not CLI
surface. #174 corrects a stale `bd links` guide example to the registered
`bd link list --incident-to` form. `todo` add/list/done, state read,
merge-slot verbs and legacy import are admitted. Shipping `set-state` and
`export` are refused in graph mode.
These are inventory findings, not release qualification for those routes.

## Next comparison batches

1. Read-only Issue discovery: `list`, `show`, `search`, `count`, `children`,
   `graph`, including default filters, order, limits and both output modes.
   Inventory the existing `query` route where it affects parity, but defer its
   generic language and execution changes until Melinda's design lands.
2. Single-Issue writes: `create`, `q`, `update`, `close`, `reopen`, `defer`,
   `undefer`, `delete`, `assign`, `priority`, `note`, with omitted and explicit
   flags, guard behavior and failure atomicity.
3. Existing Dependency and workflow facades: `dep`, `ready`,
   `blocked`, `todo`, `state`, `merge-slot`, and import/export where admitted.
   Keep Link canonical behavior and ordinary Issue JSON compatibility distinct.

For each batch, add an installed-command regression test only where it covers
a real process, persistence or output boundary that existing tests miss. Record
the exact Preview commit and actual validation result before closing a row.
