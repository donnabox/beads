# Preview 3 graph-mode CLI qualification

Status: active execution record, October 20, 2026. The tested starting point is
fork PR #141 at `e8c3b9516` (based on Preview 3 integration `dd6b9359b`).
The shipping comparison is tag `v1.3.1` (`c1c4b642a`). Recheck the candidate
head and affected tests when PR #141 changes or lands.

## Acceptance boundary

The first gate covers `bd create`, `update`, `delete`, `show`, `list`, `versions`;
their `bd link` subcommands; and the shipping two-operand `bd link FROM SOURCE`
form in a `--graph-mode link` workspace. Test both canonical Resource forms and
admitted Issue compatibility forms. The latter must preserve shipping behavior
unless the graph specification deliberately changes it. A flag that refuses
explicitly is still a parity gap. Do not turn a passing graph-specific test into
a claim that the corresponding shipping form works.

Use existing tests as regression evidence, and port relevant shipping command
cases to installed graph workspaces where they exercise a distinct process,
storage or output boundary. The ordinary test suite cannot be run wholesale
with the graph flag: many fixtures initialize dependency-mode storage and some
assert legacy-only policy. The acceptance result must name the cases ported,
cases intentionally different, and cases still unavailable. Run embedded and
shared-server variants where a command touches storage, then the relevant CI
lanes. A skipped server subtest is not a pass for that engine.

| First-gate route | Current candidate evidence | Known work to settle |
| --- | --- | --- |
| `create` | Existing installed tests cover Issue and Memory creation, parent-child creation, Markdown batch and nominal Types. | Shipping Issue `--body-file`/stdin and `--deps` are refused; the latter needs atomic Issue-and-Link behavior. Check other shipping input/validation flags. |
| `update` | Existing tests cover Issue, Memory and Link guarded edits, properties/metadata, ordered patches and historical base edits. | Shipping Issue file/stdin text, label/status and multi-ID forms are refused. Separate input, native policy and batch semantics. |
| `delete` | Existing tests cover single Issue/Memory and Link deletion, incident-Link refusal, preview/apply and retained versions. | Shipping multi-ID and `--cascade --dry-run` refuse. Choose graph deletion policy before changing destructive behavior. |
| `show` / `link show` | Canonical show and legacy current-Issue `--json` have installed tests. This branch restores familiar human Issue details, `--short`, `--long`, and comment-tail rendering; selected cases pass on embedded and shared-server Dolt. | Check shipping multi-ID, current, refs, children and watch forms separately; keep canonical JSON and retained-state rules separate. |
| `list` / `link list` | Existing tests cover bounded Bead/Issue/Link selections and legacy Issue JSON. | Common shipping Issue filters and tree/pretty presentation still refuse. Pagination implementation is deferred; preserve current product limit behavior. |
| `versions` / `link versions` | Existing tests cover Issue, Memory, live Link and removed-Link histories, local ordering and JSON member sets. | Recheck after PR #141 integration and with both real engines; do not infer HTTP History from CLI History. |
| `link FROM SOURCE` | Existing installed tests exercise informational Links, source guards and endpoint pins. This branch additionally passes embedded-process checks for default `blocks`, first-operand orientation, duplicate/no-op, explicit `tracks` and refusal to layer a different kind on the same Issue pair. | Check shipping and graph JSON/human feedback and qualify the same cases on shared-server Dolt. The starting candidate's help incorrectly said exact pins refuse; this branch corrects it. |

The second gate is `bd dep`, `bd ready`, and logical aliases to the core verbs
(including `new`, `view`, `q`, `forget`, `unlink` and graph `history`, where
registered). The remaining commands from Donna's October 20 inventory form the
third gate: assignment, child and lifecycle operations, comments, edit/form,
gate/state/labels, todo and merge-slot, reports, and dependency/structure
views. Generic query execution waits for Melinda's design. No real pagination
implementation is planned for Preview 3.

## Verification so far

On the starting PR #141 candidate, focused `go test -tags gms_pure_go
./cmd/bd` runs passed the embedded graph cases for Markdown create, Issue
delete, hidden-Issue list, versions, canonical show and compatibility defaults.
Their server subtests skipped because `BEADS_GRAPH_TEST_SERVER_PORT` was not
provided. This is inner-loop evidence only: local Bazel is unavailable in this
environment, and PR #141's extensive CI was still running at this checkpoint.
No first-gate parity or full CI success is claimed yet.

After adding the positional-Link cases, focused
`TestGraphPreviewCompatibilityDefaultsWorkflow` and
`TestGraphPreviewHelpDocumentsPlaytestCommands` passed locally. The former's
shared-server subtest still skipped. Generated CLI docs passed their freshness
check. An initial new test incorrectly assumed a pair could carry `blocks` and
`tracks` simultaneously; both the native Issue writer and graph writer reject
that, so the final test checks the type conflict and creates `tracks` on a
different pair.

`TestGraphPreviewShowLegacyIssueJSON`, `TestCanonicalShowWorkflow`, and
`TestGraphPreviewCompatibilityDefaultsWorkflow` passed locally after the
familiar human Issue renderer was added. On a local Dolt SQL server, the
selected server subtests for those three plus Markdown create, Issue delete,
hidden-Issue list, and removed-Link versions also passed. The full familiar
show test, including `--short`, `--long`, `--comments-tail`, legacy JSON and
canonical JSON, passed on embedded and shared-server Dolt. The complete
graph-mode test run and remote CI remain separate qualification gates.
