# Preview 3 CLI compatibility selection (working draft)

This is a decision aid for Donna, not a replacement for the [graph CLI
guide](../../docs/reference/graph-cli.md) or the [detailed compatibility
audit](graph-mode-shipping-compatibility-audit.md). It compares the shipping
Issue workspace (`--graph-mode dependency`) with the selected Preview 3 link
workspace at integration commit `b6d0493ba`. A linked PR is **candidate work**,
not integrated behavior. “Supported” means the named form works; it does not
imply every shipping flag works. “Refused” means an explicit rejection rather
than a silent, possibly wrong answer.

Canonical JSON currently identifies the vendored BDP wire input
`de99030d13a57e77f4f2660b91ea3c39b82823cb` in `bdp_spec`. Donna's
provisional Preview 3 Read target `205a26f5e89e7b14516f45a3087dff22d5b780f1`
still needs record-level alignment and qualification; the table does not claim
that current JSON conforms to that target.

| Command/form | `--graph-mode dependency` status | `--graph-mode link` status |
| --- | --- | --- |
| `bd create` Issue, including ordinary inline fields | Supported through the shipping Issue route. | Supported for fresh graph Issues. File/stdin description input is candidate [#152](https://github.com/donnabox/beads/pull/152); `--silent` is candidate [#159](https://github.com/donnabox/beads/pull/159). Atomic `--deps`/Markdown relationships are refused. |
| `bd create --type types/memory` | Refused; this mode has no Memory Beads. | Supported with the installed Memory Type and its admitted properties. |
| `bd show ISSUE` / `bd show MEMORY` | Shipping Issue details supported; Memory not applicable. | Current Issue read works, but familiar human/multi-ID/current/refs/children/watch behavior is candidate [#151](https://github.com/donnabox/beads/pull/151). Memory read uses the graph record. |
| `bd list` Issues / all Beads | Shipping Issue list and filters supported. | Bounded current Bead listing and Issue-only JSON/filter route supported. `--format=json`/`--tree=false` candidate [#156](https://github.com/donnabox/beads/pull/156); sort/long [#158](https://github.com/donnabox/beads/pull/158); no-parent [#166](https://github.com/donnabox/beads/pull/166); parent selection/tree [#169](https://github.com/donnabox/beads/pull/169)–[#172](https://github.com/donnabox/beads/pull/172). Other shipping filters can explicitly refuse. |
| `bd update ISSUE` / `bd update MEMORY` | Shipping Issue edits supported. | Admitted current Issue and Memory edits supported. Description file/stdin is candidate #152; forced assignee replacement is candidate [#164](https://github.com/donnabox/beads/pull/164). Other Issue label/status forms need selection and qualification. |
| `bd delete ISSUE` / `bd delete MEMORY` | Shipping Issue deletion and its directed dependency cascade apply. | Current unreferenced Bead deletion supported; linked deletion refuses. Single-Issue dry-run is candidate [#163](https://github.com/donnabox/beads/pull/163). Generic incident-Link cleanup and surviving-owner validation are approved directions requiring an implementation assessment; they are not current behavior. |
| `bd versions ID` / retained `show` | Existing CLI history/version route, where applicable; verify specific shipping form before promising it in the runbook. | Bead and Link retained versions and exact reads supported in the graph route. |
| `bd link BLOCKED BLOCKER [--type KIND]` | Shipping positional Link route; exact/unique partial Issue IDs resolve and native kind is stored. | Positional dependency route supported. Familiar feedback is candidate #151; literal custom kind preservation #161; exact kind reads #174. Bare partial-ID resolution remains a gap; explicit graph paths are exact. |
| `bd link create/list/show/update/delete/versions` | Parallel Link subcommand namespace is not a shipping Issue-workspace promise. | Core Link namespace is supported for admitted Types and guards. Canonical collection/output rules apply separately; unsupported Link mutations refuse. Legacy list/count custom-kind selection is candidate [#174](https://github.com/donnabox/beads/pull/174). |
| `bd dep add/list/remove` | Shipping dependency facade and quirks preserved. | Native blocking add/remove and bounded human views supported. All-kind human list is candidate [#165](https://github.com/donnabox/beads/pull/165); shipping JSON list shapes and some flags explicitly refuse. |
| `bd ready` / `bd ready --claim` | Shipping dependency readiness and claims supported. | Native dependency readiness and atomic claim supported, within current filters/limits. Installed two-engine informational-Link regression is candidate [#175](https://github.com/donnabox/beads/pull/175); read-only `--parent` is candidate [#160](https://github.com/donnabox/beads/pull/160). `--claim --parent` refuses. |
| `bd blocked --parent` | Shipping parent-scoped blocked view supported. | Baseline refuses; candidate [#150](https://github.com/donnabox/beads/pull/150) restores it with a checked selector. |
| `bd graph` | Shipping dependency graph and existing output options. | Current graph view is limited to its admitted dependency presentation. Melinda owns mixed-Type graph traversal and rendering; retain explicit refusals until that candidate is qualified. |

`bd links` is absent from shipping v1.3.1 and unregistered in link mode. It is
not a compatibility requirement. Generic `bd query` redesign, broad output
propagation, and filtered generic cascade are outside Preview 3's selected
compatibility pass.

For the first selection pass, I recommend the forms needed to write, inspect,
edit and list a connected Memory/Issue graph, plus the `ready` non-dependency
regression. Then choose additional familiar entry points by their runbook
value and actual PR dependencies. Green checks make a candidate reviewable;
they do not choose it for this release.
