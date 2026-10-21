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

The auditors built both exact commits and used disposable embedded and real
shared-server Dolt workspaces. Every top-level command received a live probe;
the reports classify every shipping command-local flag and invocation form,
distinguishing executed behavior from source-backed admission or refusal.
This is a **complete command and registered-flag inventory**, not proof of
every flag combination, error path, output mode, or high-volume boundary.
“Tested core” below means only the forms stated in its detailed report passed.
No current integration command should be called fully v1.3.1-compatible
solely from this table.

The singular `--type` is the shipping Issue type flag; `bd create` without
it creates a `task` in both versions. Plural `--types` is unknown in both.
`bd links` does not exist in shipping v1.3.1 and is unregistered in the
integration build. [#177](https://github.com/donnabox/beads/pull/177) removes
its dead preview code and stale examples; `bd link list --incident-to ID` is
the incident-Link listing form.

## Release-blocking findings

1. **P1: explicit Issue identity and children.** `bd create Parent --id
   audit-parent` succeeds in graph mode with canonical path
   `beads/audit-parent`, but may store a generated native Issue ID such as
   `audit-y9x`. `bd create Child --parent audit-parent` then creates a child
   named `beads/audit-y9x.1`, not shipping's `audit-parent.1`; dependency
   output mixes the two identities. This reproduced on embedded and real
   shared-server Dolt. The graph model deliberately permits independent
   canonical and native IDs, so this needs an explicit compatibility decision
   before a fix or release claim. See the [A–M evidence](preview3-cli-parity/issue-a-m.md).
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
   feature-parity certification. See the [selection record](preview3-cli-compatibility-selection.md).

Other misleading answers found during flag qualification: `bd dep tree ID
--max-depth 1` reports “has no dependencies” for a root with a direct blocking
edge, and `bd status --graph --json` advertises pair unlink as unavailable
although `bd dep remove` works. Pending [#179](https://github.com/donnabox/beads/pull/179)
and [#178](https://github.com/donnabox/beads/pull/178) respectively correct
these answers; neither fix is present in the pinned integration commit.

## Working With Issues: command-by-command release-note draft

| Command | Current link-mode status at `b6d0493ba` | What users cannot rely on yet |
| --- | --- | --- |
| `assign` | Tested core works. | Graph IDs may be `beads/ID`; live-claim takeover was not requalified in this audit. Graph help describes the wrong revision-token shape and stale-guard exit code. |
| `children` | Immediate children work. | `--pretty` refuses; explicit-ID parents can yield surprising child and dependency IDs (P1). |
| `close` | Single close and no-op work. | A batch may close earlier IDs then fail on a later one; JSON differs; `--continue` and `--no-auto` refuse. |
| `comment` | Issue comment write works. | Comments on non-Issue Beads are unavailable; graph `issue_id` uses a canonical path. |
| `comments` | Issue feed and `comments add` work. | Non-Issue comments are unavailable; custom-author/file combinations were not fully retested. |
| `create` | Omitted `--type` creates a task; admitted inline Issue and Memory forms work. | Explicit-ID child identity is P1; initial status, atomic `--deps`, dry run and several file/batch/routing forms refuse. #152 and #159 are unmerged subsets. |
| `create-form` | Explicit refusal. | Interactive Issue creation is unavailable. |
| `delete` | One unreferenced Bead can be previewed and force-deleted with a retained tombstone. | Linked Beads refuse; multi-ID, cascade, `--from-file`, and `--dry-run` refuse at this commit. #163 is an unmerged single-ID dry-run alias. |
| `edit` | Tested one-field editor no-op works. | Both shipping and graph require an ID despite `[id]` help; combining field flags enters shipping's editor but refuses in graph mode. Actual editor mutation was not requalified; Memory editing refuses. |
| `gate` | Explicit refusal. | Gate create/check/list/show/resolve/waiter/discovery are unavailable. |
| `label` | Explicit refusal. | Label add/remove/list/propagate are unavailable, despite create-time labels and list filters. |
| `link` | Positional native dependency creation works for built-in kinds; graph-only `bd link` CRUD exists. | Positional custom kinds and shipping partial-ID lookup refuse; familiar JSON differs. #151/#161/#174 are unmerged subsets. Canonical Link list requires full Type URIs and refuses truncated pages without continuation; see its eight detailed subrows. |
| `list` | Issue-only `--json` and bounded all-Bead listing work for admitted filters. | Several shipping selectors/presentations refuse: parent/no-parent/ready/ID, sort id, long, tree, `--format=json` at this commit. #156/#158/#166/#169–#172 are unmerged subsets. |
| `merge-slot` | Create/check/acquire/release work in tested flows. | Wrong-holder release refuses with misleading `graph_not_initialized` instead of a holder-mismatch category. |
| `note` | Positional, file and stdin Issue-note append work. | Non-Issue notes and multi-ID note writes are unavailable. |
| `priority` | P0–P4 Issue update works. | Invalid-value exit code differs from shipping. |
| `promote` | Explicit refusal. | Wisp promotion is unavailable. |
| `q` | Standalone quick capture works, including `feat` type normalization. | `--parent` refuses; use `create --parent`. |
| `query` | Admitted Issue predicates work. | Generic Bead/Link query is unavailable; parent OR/NOT forms can refuse. Melinda owns the query design. |
| `reopen` | Single and batch Issue reopen work; batches can partly succeed. | JSON shape differs; registered `--if-revision` refuses in graph mode. |
| `search` | Issue search works for admitted filters. | `--type feat` silently yields no results despite current help claiming the alias; direct shipping v1.3.1 also did so, while the newer ordinary Preview route normalizes it. |
| `set-state` | Explicit refusal. | Atomic state event and label write is unavailable. |
| `show` | One current or retained Resource can be read. | Familiar human Issue detail, multi-ID, current/short/long/refs/children/watch and other display forms refuse at this commit; #151 is selected but unmerged. |
| `state` | Read-only value and list forms work. | `set-state` does not; JSON may expose native rather than canonical Issue ID. |
| `tag` | Explicit refusal. | Label shortcut and equivalent graph `update --add-label` are unavailable. |
| `todo` | Bare/list/add/done work in tested flows. | Multi-ID done can partly succeed while exiting zero, matching shipping; graph list is bounded. |
| `update` | Admitted single-Issue, Memory and informational-Link edits work. | Multi-ID, Issue type/status/labels/parent/conditional-status, description file/stdin and forced assignee transfer refuse at this commit; JSON shape differs. #152/#164 repair subsets. |

## Views & Reports: command-by-command release-note draft

| Command | Current link-mode status at `b6d0493ba` | What users cannot rely on yet |
| --- | --- | --- |
| `count` | Durable-Issue totals/groups and admitted filters work. | Infra/ephemeral inclusion and some filters refuse; large inventories refuse rather than return partial counts. |
| `diff` | Explicit refusal. | Dolt commit/branch diff is unavailable; retained `compare` addresses one Resource instead. |
| `find-duplicates` | Explicit refusal. | Similarity/AI duplicate discovery is unavailable. |
| `history` | Changed contract: alias of retained `versions ID`. | No Dolt commit/audit history; `--limit` and `--events` refuse; JSON shape differs. |
| `lint` | Explicit refusal. | Issue-content linting is unavailable. |
| `stale` | Explicit refusal. | Age-based stale-Issue report is unavailable. |
| `status` | Durable-Issue summary works; `--graph` adds limits/capabilities. | Activity/assigned/all controls refuse. `--graph --json` falsely reports pair unlink unavailable even though `dep remove` works; #178 is unmerged. |
| `statuses` | Built-in status listing works. | Configured custom-status path was not live-requalified. |
| `types` | Changed contract: inventories installed Bead/Link Types and classifications. | Shipping `--sections` refuses; JSON schema differs from Issue-only `bd types`. |

## Dependencies & Structure: command-by-command release-note draft

| Command | Current link-mode status at `b6d0493ba` | What users cannot rely on yet |
| --- | --- | --- |
| `dep` | Single local add/remove and blocking views work. | Default list silently omits nonblocking kinds (P1); tree/graph are blocking-only; depth-1 tree falsely says there are no dependencies (P2); list/tree/cycles JSON, file bulk add, relate/unrelate and other forms refuse. #165 fixes the default human list and #179 fixes the depth-1 answer; both are unmerged. |
| `duplicate` | Explicit refusal. | Atomic duplicate edge and close is unavailable. |
| `duplicates` | Explicit refusal. | Exact-content grouping and optional merge are unavailable. |
| `epic` | Status, close-eligible dry run and actual close work in tested flows. | Multiple-epic partial failures were not live-probed. |
| `graph` | Default is a human blocking-only view; explicit `--view generic` previews bounded mixed traversal; `graph check` reads blocking cycles. | Shipping default formats, JSON and `--all` refuse; Melinda's #176 is draft and unmerged. Clean `graph check` JSON has an empty-array/null difference. |
| `supersede` | Explicit refusal. | Atomic supersedes edge and close is unavailable. |
| `swarm` | Namespace help remains visible. | Create/list/status/validate all explicitly refuse in link workspaces. |

## Staging consequences

- **Already prepared, still unmerged.** The first CRUD slice is #151
  (familiar show and positional Link feedback), #152 (description file/stdin),
  #156 (list format aliases), and #175 (ready regression). Other directly
  relevant candidates are #159 (`create --silent`), #158 (list sort/long),
  #163 (single-ID `delete --dry-run`), #164 (forced assignee update), #161
  (positional custom Dependency kinds), #174 (native-kind Link collection
  filters), and the stacked #166/#169–#172 list parent/tree chain. Their
  passing individual checks do not validate the assembled branch or cover
  every shipping form.
- **No complete repair PR yet.** Issue `create --deps`, initial status and
  full batch input; `delete` multi-ID/cascade/from-file; Issue `update`
  multi-ID/type/status/labels/parent and conditional guards; `list --ready`
  and `--id`; and positional Link partial-ID resolution remain work to scope,
  implement and qualify. Some need a graph transaction primitive, not just a
  CLI flag. Their explicit refusals are safer than a false success, but they
  are still shipping parity gaps.
- **Decisions before implementation.** The native/canonical Issue-ID split
  affects explicit IDs, dotted children and dependency projection. Mixed
  `close` batch failure differs from shipping direct-Dolt behavior. Legacy
  Issue JSON shapes and graph-record output differ on several commands.
  Resolve those contracts with Donna rather than silently encoding a guess.
- **CRUD+ is not complete.** The selected combined test validates a useful
  basic path, while the create-ID defect, missing familiar flags and output
  differences above still block an unqualified “v1.3.1 parity” claim. Promote
  fixes by actual command value and dependency, then rerun the exact shipping
  forms against the assembled branch on embedded and shared-server Dolt.
- **Fix silent wrong answers before broadening the surface.** Resolve the
  explicit-ID child contract and integrate/qualify an all-kind `dep list` fix
  before advertising compatibility. Correct the status capability bit so
  automation receives truthful feature discovery.
- **Keep explicit boundaries visible.** Legacy `dep` and domain commands can
  remain familiar facades where their semantics are implemented. Generic
  query awaits Melinda; pagination is deferred. Neither unmerged candidate
  behavior nor a passing focused suite should be written as shipping release
  behavior.
