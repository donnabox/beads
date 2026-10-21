# Preview 3 CLI compatibility selection

Donna selected the first compatibility slice on October 21, 2026: basic
Bead/Link CRUD and the `ready` invariant. The selected implementation PRs are
[#151](https://github.com/donnabox/beads/pull/151),
[#152](https://github.com/donnabox/beads/pull/152), and
[#156](https://github.com/donnabox/beads/pull/156); the selected `ready`
regression is [#175](https://github.com/donnabox/beads/pull/175). The existing
graph Bead/Link create, list, show, update, delete, and versions routes also
belong to this slice and still require end-to-end qualification. Selection
does not mean a PR is merged or that every shipping flag is supported. The
other linked parity PRs remain candidates for a later pass.

Donna subsequently asked for shipping v1.3.1 feature parity across CRUD+
and a [complete 43-command assessment](preview3-cli-parity.md). The focused
selection below is a tested starting slice, not a declaration that CRUD+
parity is complete. The full assessment records command-by-command gaps and
release-note wording before additional work is staged.

A disposable selected-only merge of those four PRs passed 18 focused
installed CLI workflows on embedded and real shared-server Dolt, with no
skips in the server-enabled run. [The audit](graph-mode-shipping-compatibility-audit.md)
records the exact heads and test scope. Hosted CI and review decisions remain
separate gates for each unmerged PR.

This table records that decision; it is not a replacement for the [graph CLI
guide](../../docs/reference/graph-cli.md) or the [detailed compatibility
audit](graph-mode-shipping-compatibility-audit.md). It compares the shipping
Issue workspace (`--graph-mode dependency`) with the selected Preview 3 link
workspace at integration commit `b6d0493ba`. A linked PR is **unmerged work**,
whether selected or a later candidate. “Supported” means the named form works; it does not
imply every shipping flag works. “Refused” means an explicit rejection rather
than a silent, possibly wrong answer.

Canonical JSON currently identifies the vendored BDP wire input
`de99030d13a57e77f4f2660b91ea3c39b82823cb` in `bdp_spec`. Donna's
provisional Preview 3 Read target `205a26f5e89e7b14516f45a3087dff22d5b780f1`
still needs record-level alignment and qualification; the table does not claim
that current JSON conforms to that target.

| Command/form | `--graph-mode dependency` status | `--graph-mode link` status |
| --- | --- | --- |
| `bd create` Issue, including ordinary inline fields | Supported through the shipping Issue route. | Supported for fresh graph Issues. File/stdin description input is selected [#152](https://github.com/donnabox/beads/pull/152); bare native ID/prefix handling [#180](https://github.com/donnabox/beads/pull/180), initial `--status` [#181](https://github.com/donnabox/beads/pull/181), `--silent` [#159](https://github.com/donnabox/beads/pull/159), atomic local `--deps` [#197](https://github.com/donnabox/beads/pull/197)–[#198](https://github.com/donnabox/beads/pull/198), combined `--parent --deps` [#199](https://github.com/donnabox/beads/pull/199)–[#200](https://github.com/donnabox/beads/pull/200), and existing-target Markdown dependencies [#201](https://github.com/donnabox/beads/pull/201)–[#202](https://github.com/donnabox/beads/pull/202) are candidates. Donna selected explicit refusal of `external:` targets. Partial dependency IDs, same-batch target references and graph-JSON batch relations remain gaps. |
| `bd create --type types/memory` | Refused; this mode has no Memory Beads. | Supported with the installed Memory Type and its admitted properties. |
| `bd show ISSUE` / `bd show MEMORY` | Shipping Issue details supported; Memory not applicable. | Current Issue read works; familiar human/multi-ID/current/refs/children/watch behavior is selected [#151](https://github.com/donnabox/beads/pull/151). Memory read uses the graph record. |
| `bd list` Issues / all Beads | Shipping Issue list and filters supported. | Bounded current Bead listing and Issue-only JSON/filter route supported. `--format=json`/`--tree=false` selected [#156](https://github.com/donnabox/beads/pull/156); sort/long [#158](https://github.com/donnabox/beads/pull/158); no-parent [#166](https://github.com/donnabox/beads/pull/166); parent selection/tree [#169](https://github.com/donnabox/beads/pull/169)–[#172](https://github.com/donnabox/beads/pull/172); ID filtering [#182](https://github.com/donnabox/beads/pull/182); blocker-aware ready listing [#189](https://github.com/donnabox/beads/pull/189)–[#190](https://github.com/donnabox/beads/pull/190) remain candidates. Other shipping filters can explicitly refuse. |
| `bd update ISSUE` / `bd update MEMORY` | Shipping Issue edits supported. | Admitted current Issue and Memory edits supported. Description file/stdin is selected [#152](https://github.com/donnabox/beads/pull/152); forced assignee replacement is candidate [#164](https://github.com/donnabox/beads/pull/164). Issue `--status` and close-policy force have stacked storage/CLI candidates [#183](https://github.com/donnabox/beads/pull/183)–[#184](https://github.com/donnabox/beads/pull/184); atomic `--if-status` and `--if-assignee` guards are stacked [#185](https://github.com/donnabox/beads/pull/185)–[#188](https://github.com/donnabox/beads/pull/188). Native Issue `--type` is a separate unmerged pair [#191](https://github.com/donnabox/beads/pull/191)–[#192](https://github.com/donnabox/beads/pull/192), preserving Memory Type retyping. Checked native label updates and shipping add/remove/set flags are stacked candidates [#195](https://github.com/donnabox/beads/pull/195)–[#196](https://github.com/donnabox/beads/pull/196); they are not selected or integrated. |
| `bd delete ISSUE` / `bd delete MEMORY` | Shipping Issue deletion and its directed dependency cascade apply. | Current unreferenced Bead deletion supported; linked deletion refuses. Single-Issue dry-run is candidate [#163](https://github.com/donnabox/beads/pull/163). Generic incident-Link cleanup and surviving-owner validation are approved directions requiring an implementation assessment; they are not current behavior. |
| `bd versions ID` / retained `show` | Existing CLI history/version route, where applicable; verify specific shipping form before promising it in the runbook. | Bead and Link retained versions and exact reads supported in the graph route. |
| `bd link BLOCKED BLOCKER [--type KIND]` | Shipping positional Link route; exact/unique partial Issue IDs resolve and native kind is stored. | Positional dependency route supported. Familiar feedback is selected [#151](https://github.com/donnabox/beads/pull/151); literal custom kind preservation #161, exact kind reads #174, and transaction-scoped bare partial-ID resolution [#193](https://github.com/donnabox/beads/pull/193)–[#194](https://github.com/donnabox/beads/pull/194) remain candidates. Explicit graph paths stay exact. |
| `bd link create/list/show/update/delete/versions` | Parallel Link subcommand namespace is not a shipping Issue-workspace promise. | Core Link namespace is supported for admitted Types and guards. Canonical collection/output rules apply separately; unsupported Link mutations refuse. Legacy list/count custom-kind selection is candidate [#174](https://github.com/donnabox/beads/pull/174). |
| `bd dep add/list/remove` | Shipping dependency facade and quirks preserved. | Native blocking add/remove and bounded human views supported. All-kind human list is candidate [#165](https://github.com/donnabox/beads/pull/165); shipping JSON list shapes and some flags explicitly refuse. |
| `bd ready` / `bd ready --claim` | Shipping dependency readiness and claims supported. | Native dependency readiness and atomic claim supported, within current filters/limits. Installed two-engine informational-Link regression is selected [#175](https://github.com/donnabox/beads/pull/175); read-only `--parent` is candidate [#160](https://github.com/donnabox/beads/pull/160). `--claim --parent` refuses. |
| `bd blocked --parent` | Shipping parent-scoped blocked view supported. | Baseline refuses; candidate [#150](https://github.com/donnabox/beads/pull/150) restores it with a checked selector. |
| `bd graph` | Shipping dependency graph and existing output options. | Current graph view is limited to its admitted dependency presentation. Melinda owns mixed-Type graph traversal and rendering; retain explicit refusals until that candidate is qualified. |

For graph `bd create --deps`, Donna selected atomic dependencies on local
Issues and an explicit refusal for shipping `external:` targets. The current
BDP Link contract does not admit external endpoints. Unmerged #197–#202
implement local subsets; this is a selected behavior rule, not a claim about
the pinned integration build or selection of every candidate PR.

`bd links` is absent from shipping v1.3.1 and unregistered in link mode. Donna
selected cutting its dead implementation in unmerged
[#177](https://github.com/donnabox/beads/pull/177), which passed 121
current-head checks with eight intentional skips; `bd link list` is the
supported spelling. Generic `bd query` redesign, broad output
propagation, and filtered generic cascade are outside Preview 3's selected
compatibility pass.

The selected slice includes the forms needed to write, inspect, edit and list
a connected Memory/Issue graph, plus the `ready` non-dependency regression.
Additional familiar entry points will be considered by their runbook value
and actual PR dependencies. Green checks make a candidate reviewable; they do
not select it for this release.
