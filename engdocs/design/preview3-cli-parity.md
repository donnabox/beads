# Preview 3 CLI parity against shipping v1.3.1

This is the complete command inventory Donna requested for the three prioritized
CLI families. It compares shipping tag `v1.3.1` (`c1c4b642a`) with the **as-built**
Preview 3 integration branch at `b6d0493ba` (October 21, 2026). An open PR is
listed as a possible repair, never as behavior in that integration commit.
The detailed, source-backed reports cover all 43 top-level commands and expand
namespaces and graph-only Link verbs where their behavior differs:

- [Working With Issues, A–M](preview3-cli-parity/issue-a-m.md): 13 commands plus
  positional `bd link` and eight `bd link` subcommands.
- [Working With Issues, N–Z](preview3-cli-parity/issue-n-z.md): 14 commands and
  21 expanded routes.
- [Views & Reports; Dependencies & Structure](preview3-cli-parity/reports-structure.md):
  16 commands and 30 expanded rows.

The [root-cause map](preview3-cli-root-cause-map.md) groups exceptions that
may share transaction composition, selector, relationship-read, batch-policy
or legacy-output work. It is a discussion draft, not a new release gate.

The auditors built both exact commits and used disposable embedded and real
shared-server Dolt workspaces. Every top-level command received a live probe;
the reports classify every shipping command-local flag and invocation form,
distinguishing executed behavior from source-backed admission or refusal.
This is a **complete command and registered-flag inventory**, not proof of
every flag combination, error path, output mode, or high-volume boundary.
“Tested core” below means only the forms stated in its detailed report passed.
No current integration command should be called fully v1.3.1-compatible
solely from this table.

**Work-status column (editorial, removable for release notes).** `As-built`
means the pinned integration commit has the stated behavior, not that the
whole shipping command is compatible. `Selected/open` means Donna selected
an unmerged PR for the first CRUD/ready slice. `Candidate/open` means an
unmerged PR exists but has not been selected. `Gap` means no repair PR is
currently listed. `Separate owner` identifies work outside this CLI parity
stack. `Merged/integration` will be used only after a PR lands on the
integration branch and the affected form is re-probed; no command-specific
parity PR has that label in this snapshot. `CI needs triage` reports a failing
current-head check without attributing it to product code. A PR can have
focused test evidence and still be open, unreviewed, or
failing a hosted check; the validation details are recorded below rather than
collapsed into the lifecycle label. These labels were checked against PR state
on October 21, 2026. On PR creation, focused validation, review change, or
merge, update the affected command row and the exact-head evidence below.
Only after a merge into `codex/preview3-integration` and a fresh comparison
should its behavior move into the pinned as-built column. This column is
tracking scaffolding, not proposed release-note text.

**Execution order.** Continue the root-cause and command-local tracks in
parallel. For cross-resource operations, first prove one checked transaction
maintains native Issues/edges, canonical Links, retained versions, and History;
local `create --deps` has an unmerged proof in #197–#202, while linked delete
is the next useful canary. Reuse transaction-scoped Issue selection and one
blocking/relationship classification rather than cloning partial-ID and
ready rules per command. Pin multi-ID outcome policy separately from atomic
commit. In the command-local track, qualify already-admitted CRUD/ready and
high-use commands, then wire forms that only need existing operations, such
as label/tag facades after the checked label writer, and independent read or
help/error fixes. Once the underlying operations work, compare shipping
arguments, defaults, human/JSON output, stderr and exit codes on both Dolt
engines. The legacy output adapter is later work unless it reduces the cost
of capability implementation or prevents a misleading success. No item in
this plan is a new release gate by itself.

The singular `--type` is the shipping Issue type flag; `bd create` without
it creates a `task` in both versions. Plural `--types` is unknown in both.
`bd links` does not exist in shipping v1.3.1 and is unregistered in the
integration build. [#177](https://github.com/donnabox/beads/pull/177) removes
its dead preview code and stale examples; `bd link list --incident-to ID` is
the incident-Link listing form.

## Priority findings for discussion

1. **P1: explicit Issue identity and children.** `bd create Parent --id
   audit-parent` succeeds in graph mode with canonical path
   `beads/audit-parent`, but may store a generated native Issue ID such as
   `audit-y9x`. `bd create Child --parent audit-parent` then creates a child
   named `beads/audit-y9x.1`, not shipping's `audit-parent.1`; dependency
   output mixes the two identities. This reproduced on embedded and real
   shared-server Dolt. Donna selected shipping identity for a flat bare
   `--id`; [#180](https://github.com/donnabox/beads/pull/180) implements and
   tests that rule but remains unmerged. Explicit graph-qualified paths may
   retain independent IDs. Head `4d3559d16` now has 119 successful hosted
   checks and eight intentional skips after failed-job retries; review is
   still undecided and the PR is unmerged. See the
   [A–M evidence](preview3-cli-parity/issue-a-m.md).
2. **P1: `bd dep list` silently omits nonblocking kinds.** Graph `dep add
   --type tracks` succeeds, yet default `dep list` can say “No blocking
   Dependencies.” Shipping lists that edge. [#165](https://github.com/donnabox/beads/pull/165)
   is an unmerged candidate that fixes human all-kind reads; JSON remains an
   explicit refusal. See the [structure evidence](preview3-cli-parity/reports-structure.md).
3. **Core display and editing still require unmerged work.** Current graph
   `show ID` human output is a raw record rather than familiar Issue detail;
   many familiar `show` and `update` forms refuse. Selected
   [#151](https://github.com/donnabox/beads/pull/151) and
   [#152](https://github.com/donnabox/beads/pull/152) repair subsets, while
   other forms have no selected fix. A focused combined 18-workflow test of
   #151, #152, #156 and #175 passed both engines, but that is not a v1.3.1
   feature-parity certification. A newer disposable assembly at local-only
   `fbf03c401` includes the current #175 fixture correction: its broad
   embedded graph CLI sweep passed 258 top-level tests with one expected
   server-only skip and no failures
   (`/tmp/cleo-selected-slice-broad-current.jsonl`). Seven focused installed
   create, edit, delete, versions, list, Link and ready workflows passed on
   embedded and real shared-server Dolt (23 test/subtest pass events, no
   skips; `/tmp/cleo-selected-slice-server-current.jsonl`). This qualifies
   the selected local assembly, not every shipping form or a PR merge. See
   the [selection record](preview3-cli-compatibility-selection.md).

Other misleading answers found during flag qualification: `bd dep tree ID
--max-depth 1` reports “has no dependencies” for a root with a direct blocking
edge, and `bd status --graph --json` advertises pair unlink as unavailable
although `bd dep remove` works. Pending [#179](https://github.com/donnabox/beads/pull/179)
and [#178](https://github.com/donnabox/beads/pull/178) respectively correct
these answers; neither fix is present in the pinned integration commit. #178
advanced to `56dd61119` after its first hosted run exposed a stale C0
capability allowlist. The corrected policy and pair-removal tests passed
locally; current-head hosted CI remains pending.

## Working With Issues: command-by-command release-note draft

| Command | Current link-mode status at `b6d0493ba` | What users cannot rely on yet | Work status (living tracker) |
| --- | --- | --- | --- |
| `assign` | Tested core works. | Graph IDs may be `beads/ID`; live-claim takeover was not requalified in this audit. Graph help describes the wrong revision-token shape and stale-guard exit code. | As-built; no parity PR |
| `children` | Immediate children work. | `--pretty` refuses; explicit-ID parents can yield surprising child and dependency IDs (P1). | Candidate/open #180; current-head checks green, unmerged |
| `close` | Single close and no-op work. | A batch may close earlier IDs then fail on a later missing ID; JSON differs; `--continue` and `--no-auto` refuse. Donna selected shipping's no-change-on-missing-ID rule for Preview 3. Draft #226–#227 implement same-transaction selector preflight for ordinary multi-ID close and `--claim-next`; other per-item policy refusals can still partly succeed. | Candidates/open #226–#227; current integration unchanged |
| `comment` | Issue comment write works. | Comments on non-Issue Beads are unavailable; graph `issue_id` uses a canonical path. | As-built; no parity PR |
| `comments` | Issue feed and `comments add` work. | Non-Issue comments are unavailable; custom-author/file combinations were not fully retested. | As-built; no parity PR |
| `create` | Omitted `--type` creates a task; admitted inline Issue and Memory forms work. | Explicit-ID child identity is P1 at the pinned commit; #180 fixes bare-ID alignment but is unmerged. Initial status, atomic `--deps`, dry run and several file/batch/routing forms refuse at that commit. #181 fixes initial status, #197–#198 add local atomic `--deps`, #199–#200 combine it with `--parent`, and #201–#202 admit Markdown dependencies on existing local Issues as unmerged candidates; #152 and #159 are other subsets. Draft #231 proves a new-Issue symbolic batch with canonical Links in storage, but wires no CLI form. External targets, partial dependency IDs, same-batch Markdown targets and full graph-JSON batch relationships remain gaps. | Selected/open #152; candidates/open #159, #180–#181, #197–#202, #231; #231 storage proof only, current integration unchanged |
| `create-form` | Explicit refusal. | Interactive Issue creation is unavailable. | Gap; no repair PR |
| `delete` | One unreferenced Bead can be previewed and force-deleted with a retained tombstone. | Linked Beads refuse; multi-ID, cascade, `--from-file`, and `--dry-run` refuse at this commit. #163 is an unmerged single-ID dry-run alias. Draft #209 proves opt-in linked deletion in one transaction; #210 adds a checked unforced preview; stacked draft #212 wires single-Issue CLI preview and `--force`. It leaves surviving dependents live. Draft #228–#229 extend the storage proof to bounded named sets and cascade. Draft #230 admits the multi-ID, cascade, and file CLI forms over that proof and preserves #163's single-ID dry-run behavior. Dotted native IDs explicitly refuse set deletion pending upstream #7377. | Candidates/open #163, #209–#210, #212, #228–#230; current integration unchanged |
| `edit` | Tested one-field editor no-op works. | Both shipping and graph require an ID despite `[id]` help; combining field flags enters shipping's editor but refuses in graph mode. Actual editor mutation was not requalified; Memory editing refuses. | As-built; mutation unqualified |
| `gate` | Explicit refusal. | Gate create/check/list/show/resolve/waiter/discovery are unavailable. | Gap; no repair PR |
| `label` | Explicit refusal. | Label add/remove/list/propagate are unavailable, despite create-time labels and list filters. | Gap; no command PR |
| `link` | Positional native dependency creation works for built-in kinds; graph-only `bd link` CRUD exists. | Positional custom kinds and shipping partial-ID lookup refuse in the pinned build; familiar JSON differs. #151/#161/#174 and #193–#194 are unmerged subsets. Canonical Link list requires full Type URIs and refuses truncated pages without continuation; see its eight detailed subrows. | Selected/open #151; candidates/open #161, #174, #193–#194 |
| `list` | Issue-only `--json` and bounded all-Bead listing work for admitted filters. | Several shipping selectors/presentations refuse: parent/no-parent/ready/ID, sort id, long, tree, `--format=json` at this commit. #156/#158/#166/#169–#172, #182 (`--id`), #189–#190 (`--ready`), and #203 (`--ready --parent`) are unmerged subsets. Donna selected a future shared Type-filter contract with `count`: omitted Type means the Issue family, `--type X` includes refinements, `--exact-type X` does not, and `--type any` selects all Bead Types. None of that contract has landed in the pinned build. | Selected/open #156; candidates/open #158, #166, #169–#172, #182, #189–#190, #203; Type selection design accepted, implementation unqualified |
| `merge-slot` | Create/check/acquire/release work in tested flows. | Wrong-holder release refuses with misleading `graph_not_initialized` instead of a holder-mismatch category. | As-built; error-category gap open |
| `note` | Positional, file and stdin Issue-note append work. | Non-Issue notes and multi-ID note writes are unavailable. | As-built; no parity PR |
| `priority` | P0–P4 Issue update works. | Invalid-value exit code differs from shipping. | As-built; exit-code gap open |
| `promote` | Explicit refusal. | Wisp promotion is unavailable. | Gap; no repair PR |
| `q` | Standalone quick capture works, including `feat` type normalization. | `--parent` refuses at the pinned commit; draft #218 uses the checked native child writer and preserves one-ID output, with focused both-engine qualification but pending hosted CI/review. | Candidate/open #218; current integration unchanged |
| `query` | Admitted Issue predicates work. | Parent OR/NOT forms can refuse. Generic Bead/Link query remains unavailable by deliberate Preview 3 scope: Donna chose to preserve Issue-only `bd query` compatibility without requiring a generic evaluator for Memory. | Intentional Issue-only scope; Melinda owns query; no parity PR |
| `reopen` | Single and batch Issue reopen work; batches can partly succeed. | JSON shape differs; registered `--if-revision` refuses at the pinned commit. Draft #219–#220 add a checked single-Issue revision guard, including the already-open no-op; guarded multi-ID reopen still refuses. | Candidates/open #219–#220; current integration unchanged |
| `search` | Issue search works for admitted filters. | `--type feat` silently yields no results despite current help claiming the alias; direct shipping v1.3.1 also did so, while the newer ordinary Preview route normalizes it. | As-built; no parity PR |
| `set-state` | Explicit refusal. | Atomic state event and label write is unavailable. | Gap; no repair PR |
| `show` | One current or retained Resource can be read. | Familiar human Issue detail, multi-ID, current/short/long/refs/children/watch and other display forms refuse at this commit. Selected #151 implements those named current-Issue forms and passed focused both-engine workflows, but is unmerged; other display forms remain unqualified or refused. | Selected/open #151; current-head checks green, review pending |
| `state` | Read-only value and list forms work. | `set-state` does not; JSON may expose native rather than canonical Issue ID. | As-built; no parity PR |
| `tag` | Explicit refusal. | Label shortcut and equivalent graph `update --add-label` are unavailable at this commit. #195–#196 provide the checked label writer; stacked #216 admits the `tag` shorthand and passed a focused both-engine workflow, with hosted CI and review pending. | Candidate/open #216 on #196; current integration unchanged |
| `todo` | Bare/list/add/done work in tested flows. | Multi-ID done can partly succeed while exiting zero, matching shipping; graph list is bounded. | As-built; no parity PR |
| `update` | Admitted single-Issue, Memory and informational-Link edits work. | Multi-ID, Issue type/status/labels/parent/conditional-status, description file/stdin and forced assignee transfer refuse at this commit; JSON shape differs. #152/#164 and stacked #183–#188/#191–#192/#195–#196 repair subsets. Draft #222–#223 admit a checked parent-only move; stacked #224–#225 prove a combined parent-plus-scalar edit. Empty detach remains unresolved. None changes the pinned integration behavior. | Selected/open #152; candidates/open #164, #183–#188, #191–#192, #195–#196, #222–#225 |

## Views & Reports: command-by-command release-note draft

| Command | Current link-mode status at `b6d0493ba` | What users cannot rely on yet | Work status (living tracker) |
| --- | --- | --- | --- |
| `count` | Durable-Issue totals/groups and admitted filters work. | Infra/ephemeral inclusion and some filters refuse; large inventories refuse rather than return partial counts. Donna selected the same future Type population as `list`, counted before page/display limits, with an explicit refusal if bounded acquisition is incomplete; Type-only generic counting is not in the pinned build. | As-built; filter/cap gaps open; Type selection design accepted, implementation unqualified |
| `diff` | Explicit refusal. | Dolt commit/branch diff is unavailable; retained `compare` addresses one Resource instead. | Gap; no repair PR |
| `find-duplicates` | Explicit refusal. | Similarity/AI duplicate discovery is unavailable. | Gap; no repair PR |
| `history` | Changed contract: alias of retained `versions ID`. | No Dolt commit/audit history; `--limit` and `--events` refuse; JSON shape differs. | Changed contract; no parity PR |
| `lint` | Explicit refusal. | Issue-content linting is unavailable. | Gap; no repair PR |
| `stale` | Explicit refusal. | Age-based stale-Issue report is unavailable. | Gap; no repair PR |
| `status` | Durable-Issue summary works; `--graph` adds limits/capabilities. | Activity/assigned/all controls refuse. `--graph --json` falsely reports pair unlink unavailable even though `dep remove` works; #178 is unmerged. | Candidate/open #178; new-head CI pending |
| `statuses` | Built-in status listing works. | Configured custom-status path was not live-requalified. | As-built; custom path unqualified |
| `types` | Changed contract: inventories installed Bead/Link Types and classifications. | Shipping `--sections` refuses; JSON schema differs from Issue-only `bd types`. | Changed contract; no parity PR |

## Dependencies & Structure: command-by-command release-note draft

| Command | Current link-mode status at `b6d0493ba` | What users cannot rely on yet | Work status (living tracker) |
| --- | --- | --- | --- |
| `dep` | Single local add/remove and blocking views work. | Default list silently omits nonblocking kinds (P1); tree/graph are blocking-only; depth-1 tree falsely says there are no dependencies (P2); list/tree/cycles JSON, file bulk add, relate/unrelate and other forms refuse. #165 fixes the default human list and #179 fixes the depth-1 answer; both are unmerged. | Candidates/open #165, #179 |
| `duplicate` | Explicit refusal. | Atomic duplicate edge and close is unavailable. | Gap; no repair PR |
| `duplicates` | Explicit refusal. | Exact-content grouping and optional merge are unavailable. | Gap; no repair PR |
| `epic` | Status, close-eligible dry run and actual close work in tested flows. | Multiple-epic partial failures were not live-probed. | As-built; multi-epic path unqualified |
| `graph` | Default is a human blocking-only view; explicit `--view generic` previews bounded mixed traversal; `graph check` reads blocking cycles. | Shipping default formats, JSON and `--all` refuse; Melinda's #176 is draft and unmerged. Clean `graph check` JSON has an empty-array/null difference. | Separate owner: Melinda; #176 draft |
| `supersede` | Explicit refusal. | Atomic supersedes edge and close is unavailable. | Gap; no repair PR |
| `swarm` | Namespace help remains visible. | Create/list/status/validate all explicitly refuse in link workspaces. | Gap; no repair PR |

## Staging consequences

- **Already prepared, still unmerged.** The first CRUD slice is #151
  (familiar show and positional Link feedback), #152 (description file/stdin),
  #156 (list format aliases), and #175 (ready regression). Other directly
  relevant candidates are #159 (`create --silent`), #158 (list sort/long),
  #163 (single-ID `delete --dry-run`), #164 (forced assignee update), #161
  (positional custom Dependency kinds), #174 (native-kind Link collection
  filters), #180 (bare explicit Issue ID), #181 (initial create status), #182
  (list `--id`), #183–#184 (Issue update status), #185–#186 (atomic Issue
  `--if-status` guard), #187–#188 (atomic Issue `--if-assignee` guard), and the stacked
  #166/#169–#172 list parent/tree chain. Their
  passing individual checks do not validate the assembled branch or cover
  every shipping form.
  Current-head #181 passed 256 top-level `TestGraphPreview` tests locally,
  with one server-only skip. #182 passed 12 graph list CLI tests and 15
  graph Issue list storage tests on embedded and real shared-server Dolt,
  without skips. Both remain unmerged candidates; the pinned matrix does
  not include either change.
  A disposable checkout combining candidates through #182 with the selected
  ready regression passed six installed create/list/ready workflows on both
  engines (28 test and subtest passes, no skips); this is selected integration
  evidence, not complete CRUD+ qualification.
  Stacked #183–#184 add checked Issue `update --status` with close policy,
  `--force`, session attribution, and revision/no-op behavior. #183's storage
  test and #184's installed status workflow passed embedded and real
  shared-server Dolt. Current-head #184 passed 151 pure-Go top-level
  `TestGraphPreview` tests with no skips. A disposable checkout combining
  candidates through #184 passed six installed create/list/update/ready
  workflows on both engines (28 test and subtest passes, no skips); this is
  selected integration evidence, not complete CRUD+ qualification.
- Stacked [#185](https://github.com/donnabox/beads/pull/185)–[#186](https://github.com/donnabox/beads/pull/186)
  carry `bd update --if-status` through a checked transaction, including
  stale no-op refusal. The installed status workflow passed embedded and real
  shared-server Dolt; current-head #186 pure-Go graph tests had 921
  test/subtest passes without skips. Both are unmerged candidates.
- Stacked [#187](https://github.com/donnabox/beads/pull/187)–[#188](https://github.com/donnabox/beads/pull/188)
  carry `bd update --if-assignee` through the same checked transaction,
  including expected-unassigned and stale no-op cases. The installed
  assignment workflow passed embedded and real shared-server Dolt;
  current-head #188 pure-Go graph tests had 922 test/subtest passes without
  skips. Both are unmerged candidates.
  A disposable checkout combining candidates through #188 and selected #175
  passed eight installed create/list/update/ready workflows on embedded and
  real shared-server Dolt (32 test/subtest passes, no skips). This is focused
  integration evidence, not complete CRUD+ qualification or a merge.
- Stacked [#189](https://github.com/donnabox/beads/pull/189)–[#190](https://github.com/donnabox/beads/pull/190)
  query the blocker-aware ready set in the checked graph snapshot and wire
  `bd list --ready` for human, Issue JSON, and graph-record output. Focused
  storage tests passed embedded and real shared-server Dolt (85 test/subtest
  passes, no skips); the broader embedded graph list sweep had 160 passes and
  11 server-only skips. The installed CLI ready-list workflow passed both
  engines (three test/subtest passes, no skips). #189 passed four and #190 six
  current-head hosted checks; both remain unmerged with review pending.
  #190 still refuses `--ready --parent`; stacked
  [#203](https://github.com/donnabox/beads/pull/203) admits the shipping
  transitive parent selector in the same checked transaction and keeps its
  ready output flat. Focused storage and installed CLI tests passed embedded
  and real shared-server Dolt for legacy JSON, records, the alias, and human
  output; all six current-head hosted checks passed and review is pending.
  Other filters the ready
  query cannot honor still refuse.
  A disposable checkout combining candidates through #190 and selected #175
  passed five installed create/update/list/ready workflows on embedded and
  real shared-server Dolt (25 test/subtest passes, no skips). This is focused
  integration evidence, not complete CRUD+ qualification or a merge.
- Stacked [#191](https://github.com/donnabox/beads/pull/191)–[#192](https://github.com/donnabox/beads/pull/192)
  admit native Issue `update --type` inside the checked graph transaction while
  preserving nominal Memory retyping. A task-to-bug edit, repeat no-op, unknown
  type refusal and retained version passed on embedded and real shared-server
  Dolt (three storage test/subtest passes, no skips). The installed Issue type
  workflow and existing Memory Type workflow passed both engines (six
  test/subtest passes, no skips). Generated CLI docs and diff checks passed.
  #191 passed four and #192 six current-head hosted checks; both reviews
  remain pending. Neither PR is merged.
  A disposable checkout combining candidates through #192 and selected #175
  passed four installed create/list/Issue-type/Memory-type workflows on
  embedded and real shared-server Dolt (22 test/subtest passes, no skips).
  This is focused integration evidence, not complete CRUD+ qualification.
- Stacked [#193](https://github.com/donnabox/beads/pull/193)–[#194](https://github.com/donnabox/beads/pull/194)
  resolve the shipping positional `bd link BLOCKED BLOCKER` abbreviated
  native Issue IDs inside the dependency write transaction. Explicit
  `beads/PATH` selectors remain exact, and ambiguous bare IDs or path
  collisions refuse before a write. The familiar result uses the resolved
  target returned by that transaction. The storage regression and installed
  CLI workflow each passed embedded and real shared-server Dolt (three
  top-level/subtest passes apiece, no skips); docsync, generated CLI docs,
  freshness and diff checks passed for #194. #193 passed four and #194 passed
  six current-head hosted checks; both review decisions remain pending.
  Neither is merged.
  A disposable checkout combining candidates through #194 and selected #175
  passed the positional Link, ready-list, and Issue type-update workflows plus
  the storage abbreviation regression on embedded and real shared-server
  Dolt (12 test/subtest passes, no skips). This is a focused combined slice,
  not complete CRUD+ qualification or a PR merge.
- Stacked [#195](https://github.com/donnabox/beads/pull/195)–[#196](https://github.com/donnabox/beads/pull/196)
  route shipping `bd update --add-label`, `--remove-label` and `--set-labels`
  through the checked native Issue writer, including combined native type and
  label edits. The storage add/replace/remove/clear, stale and no-op regression
  and the installed CLI workflow each passed embedded and real shared-server
  Dolt (three test/subtest passes apiece, no skips). Focused parser tests,
  generated CLI docs and diff checks passed. A broader embedded
  `TestIssueUpdate` slice passed 13 tests/subtests, with seven server-only
  subtests skipped; it is separate from the focused two-engine label proof.
  #195 and #196 each passed four current-head hosted checks; both review
  decisions remain pending. Neither PR
  is merged or selected. This does not restore the
  separate `bd label` or `bd tag` aliases on those heads. In a disposable checkout combining
  candidates through #196 with selected #175, installed description-source,
  ready, positional Link partial-ID, Issue type and label workflows passed
  embedded and real shared-server Dolt (15 test/subtest passes, no skips).
  The two textual cherry-pick conflicts preserved both existing file/stdin
  description admission and new label flags. This remains a focused combined
  slice, not complete CRUD+ qualification or a PR merge.
- Stacked [#216](https://github.com/donnabox/beads/pull/216) at `fb3ac1499`
  routes `bd tag ID LABEL` through #196's checked native Issue label writer
  and preserves its Issue-shaped human and JSON feedback. The installed
  shortcut workflow passed on embedded and real shared-server Dolt without
  skips, including label normalization,
  duplicate no-op, native-ID and graph-path selection, ambiguity, Memory and
  read-only refusals. A temporary wrapper used the repository's isolated
  local-server helper with the pinned Dolt 2.2.0 binary, then was removed;
  the exact PR source commit is unchanged (`/tmp/cleo-tag-local-qualification-20261022.jsonl`).
  Generated CLI docs, Markdown links and diff checks passed; current-head
  hosted CI and review are pending. It is an unmerged,
  unselected command-local candidate, not integration behavior.
- Draft [#218](https://github.com/donnabox/beads/pull/218) at `2a5ad8af1`
  admits `bd q TITLE --parent PARENT` on a fresh graph workspace through the
  existing checked native child writer. The child receives a dotted native
  Issue ID, inherited labels, and a parent-child Link in the same transaction;
  quick capture still prints exactly one local ID, including with `--json`.
  The installed workflow passed embedded and isolated real shared-server Dolt
  without skips, covering the resulting Issue and Link plus empty/missing
  parent refusals (`/tmp/cleo-quick-parent-local-qualification-20261022.jsonl`).
  A temporary wrapper used the repository's isolated local-server helper with
  pinned Dolt 2.2.0, then was removed; the exact PR source is unchanged.
  Generated CLI docs, Markdown links and diff checks passed. Hosted CI and
  review are pending; this is an unmerged, unselected candidate.
- Draft [#219](https://github.com/donnabox/beads/pull/219) at `e8ab9b98b`
  adds a graphstore reopen guard checked inside the native lifecycle write
  transaction, before an already-open no-op. Stacked draft
  [#220](https://github.com/donnabox/beads/pull/220) at `99ae3cd66`
  admits `bd reopen ID --if-revision TOKEN` for one graph Issue while
  preserving unguarded batch behavior and refusing an empty token or guarded
  multi-ID request. The focused storage and installed CLI workflows each
  passed embedded and isolated real shared-server Dolt without skips
  (`/tmp/cleo-reopen-revision-storage-local-20261022.jsonl` and
  `/tmp/cleo-reopen-revision-cli-local-20261022.jsonl`). Adjacent reopen
  storage and batch CLI tests passed locally. Generated CLI docs, Markdown
  links and diff checks passed. Both PRs credit the native reopen guard
  contributor in [gastownhall/beads#7209](https://github.com/gastownhall/beads/pull/7209)
  and preserve earlier [#64](https://github.com/donnabox/beads/pull/64).
  Hosted CI and review are pending; neither candidate is merged or selected.
- Draft [#222](https://github.com/donnabox/beads/pull/222) at `19c074276`
  adds a checked graphstore nonempty Issue reparent operation. The native
  parent-edge replacement, old canonical Link tombstone, new canonical Link,
  and child-owned History commit in one transaction. Fault injection after
  each write stage, stale-revision refusal, same-parent no-op, cycle refusal,
  and native-ID selector tests passed embedded and isolated real shared-server
  Dolt without skips (`/tmp/cleo-reparent-store-qualification-20261022.jsonl`).
  Adjacent child-create and Issue-update storage tests passed locally. Its
  first hosted native-lint job failed before a code finding because Bazel
  timed out fetching Toybox for an unrelated test package. The isolated
  failed-job retry and required CI gate passed at current head. The first
  macOS graphstore race attempt then hit its 30-minute test timeout in
  `TestLinkPropertiesPatchDeletedBudget/embedded` after three preceding
  graphstore groups passed; no reparent assertion failed. Only that failed
  job was retried in workflow `39710093424`, and the retry is pending. It is
  storage-only: `bd update --parent` remains unavailable on the pinned
  integration commit and on this PR head. Empty-parent detach is deliberately
  excluded because dotted IDs still imply a parent in shipping selectors;
  [upstream contributor #7369](https://github.com/gastownhall/beads/pull/7369)
  documents that ambiguity and is credited without being replaced. Stacked
  draft [#223](https://github.com/donnabox/beads/pull/223) at `e7143c7c8`
  wires `bd update CHILD --parent PARENT` for one local Issue, with the
  ordinary graph Issue-update feedback and optional child revision guard.
  Its installed workflow passed embedded and isolated real shared-server Dolt
  without skips, covering stale refusal, native children, Link replacement,
  same-parent no-op, and explicit refusal of empty detach or combined edits
  (`/tmp/cleo-reparent-cli-qualification-20261022.jsonl`). Generated CLI
  docs, Markdown links and diff checks passed. A broader exact-head pure-Go
  `TestGraphPreview` run passed in `cmd/bd` and `internal/storage/graphstore`:
  150 top-level tests, 928 test/subtest passes, no skips or failures
  (`/tmp/cleo-reparent-broad-20261022.jsonl`). A disposable combined checkout
  at local-only `f2969678d` cleanly cherry-picked #222–#223 onto the earlier
  candidate integration checkout; its broad embedded graph run passed 151
  top-level tests and 931 test/subtest results without skips or failures.
  Selected installed reparent, linked-delete, Markdown-dependency and ready
  workflows then passed embedded and isolated real shared-server Dolt in the
  same checkout: 13 test/subtest results, no skips or failures
  (`/tmp/cleo-parity-through223-embedded-20261022.jsonl`,
  `/tmp/cleo-parity-through223-server-20261022.jsonl`). This is selected
  integration evidence, not complete CRUD+ qualification or a PR merge.
  Current-head hosted CI is running or queued and review pending for both
  PRs; neither is merged or selected.
- Draft [#224](https://github.com/donnabox/beads/pull/224) at `aad565b91`
  extends #222's graphstore reparent composer so one native Issue writer
  applies the parent edge and admitted scalar patch before the old/new
  canonical Links, retained Issue head and History publish in the same
  transaction. Shared field validation also remains the path for ordinary
  Issue updates. A same-parent scalar edit keeps its Link identity; a fully
  identical retry is a no-op. Mixed-edit stale revision and injected rollback
  controls passed embedded and isolated real shared-server Dolt: three
  test/subtest passes, no skips in the server-backed run
  (`/tmp/cleo-reparent-edit-server-20261022.jsonl`). The focused update and
  reparent test run passed six test/subtest results with three expected
  server-only skips in the separate no-server invocation. Stacked draft
  [#225](https://github.com/donnabox/beads/pull/225) at `a40fa4b66` wires
  combined `bd update CHILD --parent PARENT --title/--priority/...` through
  that composer. Its installed workflow passed embedded and isolated real
  shared-server Dolt (three test/subtest passes, no skips in the server-backed
  run; `/tmp/cleo-reparent-edit-cli-server-20261022.jsonl`). Generated CLI
  docs, Markdown links and diff checks passed. The exact-head broad graph CLI
  sweep passed 256 top-level tests and 1,220 test/subtest results, with one
  top-level and 63 total expected server-only skips, zero failures
  (`/tmp/cleo-reparent-edit-cli-broad-20261022.jsonl`). At their current
  heads, #224 and #225 each have four hosted successes; #225 also has two
  intentional documentation skips. Reviews remain undecided; both PRs are
  unmerged, unselected
  candidates, and empty-parent detach is still refused. A disposable combined
  checkout at local-only `beaec990e` cherry-picked #224–#225 over earlier
  status/type/label, ready and linked-delete candidates; its shared Issue
  update helper required a manual merge that retained those earlier fields.
  The combined branch explicitly refuses parent moves with status/type/label
  or conditional field edits until those policies are composed rather than
  silently dropping them. Focused embedded tests passed eight test/subtest
  results with four expected server-only skips; isolated real shared-server
  storage and installed CLI workflows passed six results without skips
  (`/tmp/cleo-parity-through225-focused-20261022.jsonl`,
  `/tmp/cleo-parity-through225-server-20261022.jsonl`). This is local assembly
  evidence, not a PR merge or full CRUD+ qualification. A later disposable
  assembly at local-only `8ce4e86f9` retained the newer update fields during
  the #224–#225 cherry-pick and kept the combined parent-plus-status/type/label
  and conditional edits as explicit refusals. Its reparent storage and installed
  CLI workflows passed on embedded and real shared-server Dolt without skips
  (`/tmp/cleo-integrated-reparent-server-test.jsonl`); the exact assembled
  branch's broad embedded graph CLI sweep passed 274 top-level tests with one
  expected server-only skip and no failures
  (`/tmp/cleo-integrated-reparent-broad.jsonl`). These results qualify the
  local composition only; they do not select or merge either PR.
- Stacked [#197](https://github.com/donnabox/beads/pull/197)–[#198](https://github.com/donnabox/beads/pull/198)
  admit local Issue `create --deps` through one transaction. #197 creates the
  native Issue and dependency edges, canonical Links and source-owned History;
  #198 reuses the shipping parser for forward `blocked-by`/`depends-on` and
  reversed `blocks`, while refusing `external:` targets explicitly. The
  storage forward/reverse and injected Link-retention rollback test and the
  installed CLI workflow each passed embedded and real shared-server Dolt
  (four test/package results apiece, no skips). Related create/child storage
  tests passed five results with two server-only skips; adjacent CLI create
  tests passed 67 results with three server-only skips. Generated CLI docs,
  Markdown links and diff checks passed. Both PRs are unmerged candidates;
  #197 has four and #198 has six passing current-head hosted checks, with no
  review decisions. Partial dependency
  IDs and dependency-bearing Markdown batches remain explicit gaps. Unmerged
  #199–#200 admit `--parent --deps` atomically in the child-create route.
  Their exact assembled source passed the existing child-create, new
  rollback, and installed CLI dependency workflows on embedded and real
  shared-server Dolt (11 test/package results, no skips), including injection
  after all Links were staged. Generated CLI docs, Markdown links, and diff
  checks passed. The broader `TestGraphPreviewCreate` selection on #200's
  assembled source passed 67 test/subtest results with three server-only
  skips. Hosted checks and review are pending. #198 head `ff68a860e`
  also asserts that omitted `--type`
  still creates a task; its focused embedded workflow and six hosted checks
  passed. A disposable checkout combining earlier candidates through #196
  with #197's cycle-error fix and current #198 at `684fda7d3` preserves
  description file/stdin, `--silent`, initial status and local dependency
  creation. Their selected installed workflows passed embedded and real
  shared-server Dolt (23 test/subtest results, no skips). The textual
  create-command conflict was resolved to retain both flag sets. Its assembled
  `TestGraphPreviewCreate` selection passed 69 test/subtest results with four
  server-only skips on embedded Dolt. The combined checkout is local-only,
  not a PR merge or full CRUD+ qualification.
- Stacked [#201](https://github.com/donnabox/beads/pull/201)–[#202](https://github.com/donnabox/beads/pull/202)
  admit Markdown `### Dependencies` on existing local Issues. #201 maps
  native edges, canonical Links and source-owned History in the batch
  transaction; #202 wires `bd create --file` and retains the ordinary bare
  Issue-array JSON response. A fault after Link retention and a missing target
  roll the whole file back; `external:` is explicitly refused. The storage
  batch tests and installed Markdown workflow each passed embedded and real
  shared-server Dolt on their exact source. The broader embedded
  `TestGraphPreviewCreate` selection passed with server-only subtests not run.
  Docsync, generated CLI docs, freshness and diff checks passed. #201's four
  current-head hosted checks and #202's six checks at `de012f5bb` passed;
  both reviews are pending. A disposable local checkout at `fb8e1d02c` combined earlier
  parity candidates, selected #175, and #197–#202 without pushing or merging
  PRs. Seven selected create/storage workflow roots passed on embedded and real
  shared-server Dolt in two Go packages. An assembled broad
  `TestGraphPreview` run first found one obsolete expectation: a local
  dependency targeting a Memory now fails Issue-selector validation rather
  than the old blanket capability refusal. #202 corrected the test; the
  local-only combined checkout at `cb610557b` then passed the full selected
  embedded run: 274 top-level tests passed, one server-only top-level skip,
  zero failures (1,265 pass and 81 skip events including subtests). This is
  broader embedded evidence, not complete CRUD+ or shared-server qualification.
  The PRs remain unmerged candidates.
  A separate local-only proof at `9f92ab4aa` assembles the #193–#194
  positional selector with #197–#202 create-dependency work and reuses the
  same transaction-scoped native Issue operand resolution for `create --deps`.
  Abbreviated local targets resolve before the create write; ambiguous IDs
  roll back, while explicit graph paths stay exact. Focused storage and
  installed CLI workflows passed on embedded and real shared-server Dolt
  without skips (`/tmp/cleo-create-deps-partial-server-tests.jsonl`). This
  proof has no PR or merge; the two candidate stacks still need a clean
  integration base before it can change the pinned command behavior.
- Draft [#209](https://github.com/donnabox/beads/pull/209)–[#210](https://github.com/donnabox/beads/pull/210) and stacked [#212](https://github.com/donnabox/beads/pull/212)
  provide a checked single-Issue linked-delete path: unforced preview refuses
  surviving native dependents; `--force` tombstones incident canonical Links
  and the Issue in one transaction, leaving surviving dependents live and
  unlinked. Their exact source passed focused embedded and real shared-server
  Dolt storage and installed CLI workflows without skips. A disposable local
  checkout also combined #163's `--dry-run` admission with these PRs and
  earlier CRUD candidates; the existing Issue and Memory delete workflows
  passed on embedded Dolt (two server subtests skipped). Its two textual
  cherry-pick conflicts retained both `--dry-run` and linked-Issue semantics.
  The broader assembled `TestGraphPreview` run at local-only `5b0552c61`
  then passed 273 top-level embedded tests, with one server-only top-level
  skip and no failures (1,263 pass and 80 skip events including subtests;
  `/tmp/cleo-combined-through212-graph-preview-20261022.jsonl`). This is not
  a shared-server qualification of the assembled checkout, and
  multi-ID/cascade/from-file remain gaps. #212's current head `806934bc3`
  corrects stale guide examples that used an unregistered `--unconditional`
  flag; its six current-head hosted checks passed, with review pending. No PR
  was merged by this test.
- Draft [#226](https://github.com/donnabox/beads/pull/226)–[#227](https://github.com/donnabox/beads/pull/227)
  implement Donna's selected multi-ID close rule. #226 adds an opt-in
  same-transaction all-selector preflight to the graph batch writer; #227
  routes ordinary multi-ID close through it and keeps the existing ordinary
  JSON result shape. A missing selector leaves every sibling Issue, claim,
  History entry and retained version unchanged. A valid selector with an
  assignee-policy refusal still has a per-item outcome. Focused graphstore
  and installed CLI workflows passed on embedded and real shared-server
  Dolt, with no skips; docs link, generated CLI docs, freshness and diff
  checks passed. Hosted current-head CI and review are pending. #64's
  Preview 2 partial-missing example is superseded only for this selected
  Preview 3 rule; these PRs do not replace its other batch behavior.
- Draft [#228](https://github.com/donnabox/beads/pull/228)–[#229](https://github.com/donnabox/beads/pull/229)
  extend the linked-delete storage proof from one Issue to a bounded named
  set or native cascade closure. The graph composer preflights every named
  selector, maps the full native set to live graph Issues, tombstones its
  canonical incident Links, deletes the native set, retains Issue terminals,
  and versions surviving rewritten citations in one transaction. Focused
  plan/apply tests passed embedded and real shared-server Dolt without skips,
  including force, preview, named pair, cascade and staged rollback. A broader
  embedded Issue-delete sweep passed, with server subtests intentionally
  skipped in that sweep. These PRs remain storage-only candidates; stacked
  draft [#230](https://github.com/donnabox/beads/pull/230) admits multi-ID,
  `--cascade`, and `--from-file` over the checked set writer. Its exact source
  passed installed set, single-Issue and Memory deletion workflows on embedded
  and real shared-server Dolt: 32 test/subtest passes, no skips. A broader
  embedded `TestGraphPreview` sweep passed 256 top-level tests, one expected
  server-only skip, and zero failures. It also preserves #163's single-ID
  `--dry-run` behavior, including `--force --dry-run --readonly`, and credits
  that prior work. Generated CLI docs, docsync, doc flags/freshness, managed
  Go source lists, and diff checks passed. #230 is an unmerged candidate with
  hosted CI and review pending. Dotted native Issue IDs still refuse until
  upstream contributor
  [#7377](https://github.com/gastownhall/beads/pull/7377) is integrated.
  Upstream contributor
  [#7512](https://github.com/gastownhall/beads/pull/7512) owns a separate
  Markdown alias/direction parser correction; these PRs leave that parser
  untouched and do not replace the contributor's work. Partial target IDs,
  references to another new heading in the same file, and full graph-JSON
  batch relationships remain explicit gaps. Draft
  [#231](https://github.com/donnabox/beads/pull/231) now proves the smaller
  storage case of two generated-ID Issues and a forward symbolic edge. Its
  fault after native edge insertion and before canonical Link retention left
  no partial state on embedded and real shared-server Dolt. It admits no CLI
  form, existing-Issue endpoint, metadata or update/close batch item; a
  broader embedded graphstore sweep passed 177 test/subtest results with 21
  server-only skips. Current-head hosted CI and review are pending.
- A broad local assembled `TestGraphPreview` run through #196 completed with
  two failing top-level tests: `--ready=false` expected a now-obsolete
  refusal, and the usability workflow expected `capability_unavailable` for
  a native Link kind with Memory endpoints rather than the now-correct
  `invalid_properties`. Their owning PRs #190 and #161 were updated, and the
  focused tests passed. The subsequent assembled `TestGraphPreview` rerun
  through #196 passed 1,263 test/subtest results with 80 server-only skips
  and no failures. It ran on embedded Dolt; the earlier failing run is not
  counted as passing, and neither run qualifies the shared-server tier.
- **No complete repair PR yet.** Issue `create --deps` has an unmerged local
  target subset in #197–#198, with parent combination in #199–#200 and
  existing-target Markdown batches in #201–#202. #231 adds an unmerged
  storage-only symbolic graph batch proof. Full graph-JSON batch relations,
  `delete` multi-ID/cascade/from-file and linked CLI policy (draft #230,
  unmerged), Issue `update` multi-ID/parent,
  remaining `list --ready` combinations, and broader Link selector forms
  remain work to scope, implement and qualify.
  Some need a graph transaction primitive, not just a CLI flag. Their explicit
  refusals are safer than a false success, but they are still shipping parity
  gaps. #193–#194 cover positional partial IDs only after they land.
- **`create --deps` has one model boundary.** Shipping's dependency parser
  preserves an `external:` target, but the current graph Reference contract
  explicitly does not admit external Link endpoints. Local Issue targets need
  one transaction that creates the Issue, native edges, canonical Links, and
  source-owned History together. Unmerged #197–#198 implement that local
  transaction, #199–#200 extend it to a child Issue and parent Link, and
  #201–#202 handle Markdown batches targeting existing Issues. Their
  focused embedded and real shared-server Dolt tests cover both directions
  and rollback. An `external:` target cannot silently become
  a native-only edge invisible to graph reads. Donna selected atomic local
  Issue targets with an explicit `external:` refusal for graph mode. Partial
  dependency IDs and references to a newly created heading in the same
  Markdown file remain unavailable in these candidates. See
  `cmd/bd/create_deps.go` and
  `docs/reference/bdp/cli/references.md`.
- **Decisions before implementation.** Donna selected native ID alignment
  for flat bare `create --id`; #180 carries it, while graph-qualified path
  semantics remain distinct. Mixed `close` batch failure differs from
  shipping direct-Dolt behavior. Legacy Issue JSON shapes and graph-record
  output differ on several commands. Resolve those contracts with Donna
  rather than silently encoding a guess.
- **CRUD+ is not complete.** The selected combined test validates a useful
  basic path, while the create-ID defect, missing familiar flags and output
  differences above still block an unqualified “v1.3.1 parity” claim. Promote
  fixes by actual command value and dependency, then rerun the exact shipping
  forms against the assembled branch on embedded and shared-server Dolt.
- **Fix silent wrong answers before broadening the surface.** Integrate and
  qualify the explicit-ID child fix and an all-kind `dep list` fix
  before advertising compatibility. Correct the status capability bit so
  automation receives truthful feature discovery.
- **Keep explicit boundaries visible.** Legacy `dep` and domain commands can
  remain familiar facades where their semantics are implemented. Generic
  query awaits Melinda; pagination is deferred. Neither unmerged candidate
  behavior nor a passing focused suite should be written as shipping release
  behavior.
