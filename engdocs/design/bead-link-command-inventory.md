# Bead and Link command inventory

Research snapshot: 2026-10-13. This is an inventory and recommendation for the
namespace discussion, not an amendment to `canonical-cli.md` or an implemented CLI
contract.

## Sources and scope

- Stable: `gastownhall/beads` v1.3.1, commit
  `c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c`.
- Preview: `donnabox/beads` `release/preview2-integration`, commit
  `63733f4d005611c1d6d19c8707f651ac16e441cc`.

The inventory examines the root command registrations, the graph admission gate,
and the resource dispatches. Registration and help alone do not establish Link
support: many commands remain Issue-only or are refused in graph workspaces.
The references below are pinned to the inspected code, not moving branch heads.

“Polymorphic” here means a resource command that accepts both a Bead ID and a Link
ID. It does not mean a command that accepts multiple Bead Types (Issue, Memory,
or potentially Type-definition Beads), nor a workspace operation that naturally
contains both Beads and Links.

## Recommended parallel resource surface

Use the same ten core verbs in the Bead and Link command spaces:

| Operation | Bead form | Link form | Present Preview support and proposed work |
|---|---|---|---|
| Create | `bd create` | `bd link create` | Currently separate: root create creates Beads; root link creates Links. Add the Link subcommand without removing `bd link SOURCE TARGET`. |
| List | `bd list` | `bd link list` | Root list selects Beads. Existing `bd links BEAD` lists incident Links; an unanchored Link collection listing is additional work. |
| Show | `bd show ID` | `bd link show ID` | Root show already reads both, including exact versions. Add scoped Link spelling. |
| Update | `bd update ID` | `bd link update ID` | Root update already dispatches by kind. Add scoped Link spelling with the same properties/metadata and ownership rules. |
| Delete | `bd delete ID` | `bd link delete ID` | Root delete currently accepts Beads; Links use unlink. Preserve unlink as a compatibility adapter. |
| Versions | `bd versions ID` | `bd link versions ID` | Root versions already accepts both. |
| Compare | `bd compare ID --from V1 --to V2` | `bd link compare ID --from V1 --to V2` | Root compare already accepts both; compares retained versions of one Resource. |
| Search | `bd search TEXT` | `bd link search TEXT` | Current graph search is Issue-only. Link search needs a defined searchable field set; do not assume every Link has a title or description. |
| Query | `bd query EXPR` | `bd link query EXPR` | Current graph query is Issue-only. A common expression language can serve two explicitly scoped collections. |
| Count | `bd count` | `bd link count` | Current graph count is Issue-only. Link count should share Link list/query predicates. |

Sources: [stable link shorthand][stable-link], [Preview creation][create],
[Preview list][list], [Preview show][show], [Preview update dispatch][update],
[Preview deletion][delete], [Link lifecycle commands][lifecycle],
[versions][versions], [compare][compare], [Issue discovery][discovery].

`history` is an eleventh spelling, but not an eleventh independent Preview
operation: `bd history ID` dispatches to the same version listing as `versions`.
It should have a `bd link history ID` compatibility alias if the naming scheme is
fully parallel. Stable `history` has different semantics: Dolt commit history and
an optional database-event view. Do not silently collapse that stable behavior
into exact Resource versions. Preview's versions route also rejects unsupported
history flags. [History dispatch][history], [version flag admission][versions].

`edit` is the strongest next candidate: `bd link edit ID` could edit properties
and metadata through an editor, using the same validation and revision checks as
update. Today's editor is explicitly an Issue-field editor; it rejects Links and
non-Issue Beads. Symmetry here requires a new generic editor contract, not merely
registration of an alias. [Preview edit][edit].

### Existing polymorphism and compatibility

At this snapshot the exact top-level spellings accepting a resource selector of
either kind are **show, update, versions, history, compare**. This excludes
collection operations and commands whose input is a Bead but whose output
contains Links. Root create, list, and delete do not currently provide Link CRUD.
The difference matters when estimating migration work. [Admission gate][admission],
[show][show], [update][update], [versions][versions], [history][history],
[compare][compare], [list][list], [delete][delete].

There are two distinct promises available:

1. **No canonical polymorphic resource commands:** new documentation and help put
   Link operations under `bd link`; the five old root forms continue as
   compatibility routes when passed an explicit Link ID.
2. **No accepted polymorphic resource commands:** root commands reject Link IDs.
   That breaks the five already-supported Preview forms and requires an explicit
   compatibility boundary or deprecation period.

The first satisfies the usability goal while preserving current surface area.
It still has polymorphism in compatibility handling. Calling it “none anywhere”
would conceal a real exception.

A Link subcommand should interpret bare IDs as Link IDs and accept `links/PATH`;
a Bead command should interpret bare IDs as Bead IDs. Explicitly wrong-kind IDs
should fail in the new scoped interface rather than silently switch command
spaces. Compatibility routing must be clearly documented separately. This is a
proposal, not current parser behavior.

## Conditional candidates: do not mirror by name alone

These are the remaining resource-facing families worth an explicit decision.
They are not all recommendations to add commands now.

| Existing family | Current meaning | Link recommendation |
|---|---|---|
| `label`, `tag` | Issue labels; tag is a convenience for adding one. Label also has collection-wide rename/list-all and parent-child propagation. | Add only if labels become a common Resource capability. Mirror add/remove/list and tag then; define propagation independently. Arbitrary Link metadata alone does not establish a label contract. |
| `comment`, `comments` | Append/list Issue comments. Preview routes still require Issues. | Useful if discussion attaches to any Resource; add `link comment` / `link comments` only with Link comment storage, attribution, and history semantics. |
| `note` | Append Issue notes, a property-level operation. | Optional convenience only if a Link Type defines a compatible notes field. Do not require all Link Types to have notes. |
| `provenance record/log` | Append-only external artifact bindings keyed by Issue ID. | A credible future common Resource capability; if adopted, mirror under link or offer an explicitly resource-oriented provenance namespace. It is not current Link history. |
| `lint` | Required description/template sections for Issue classifications. | A future Link Type/schema linter is meaningful, but it is new semantics. Avoid pretending today's template linter validates arbitrary Links. |
| `rename` | Change an Issue ID and rewrite references. | Do not mirror automatically. First decide immutable graph identity versus human aliases and redirection. This is not a title update. |
| `restore` | Recover pre-compaction Issue content; optional apply. | Do not advertise Link undelete or version rollback under this spelling without a separate recovery design. Current restore does neither. |
| `stale` | Find Issues not updated recently, with Issue lifecycle selection. | Could be a Link-list convenience for an age predicate. Define Link staleness first; old relationships can remain valid indefinitely. |
| `duplicates`, `find-duplicates` | Detect duplicate Issues (and, for duplicates, optionally merge). | Only with a Link-equivalence definition: endpoint pair alone is insufficient when Types/properties/multiple edges differ. Prefer a filter or report before a mutation command. |
| `duplicate`, `supersede` | Issue workflows, including relationship and lifecycle effects. | Do not equate with copying/replacing a Link. Require explicit semantics for Link history, identity, owner changes, and deletion before adding parallel forms. |
| Pinning | Issue flags/status and list filters, not a `bd pin`/`bd unpin` root command. | If common pin metadata is adopted, expose it through update/list first; there is no root pin command to mirror. |

Sources: [labels][label], [tag][tag], [comments][comments], [comment][comment],
[note][note], [provenance][provenance], [lint][lint], [rename][rename],
[restore][restore], [stale][stale], [duplicates][duplicates],
[find-duplicates][find-duplicates], [duplicate/supersede][duplicate],
[Issue list filters][list]. The graph admission gate explicitly refuses unsupported
routes, and refuses duplicate/supersede until their graph workflows are complete.
[Admission gate][admission].

There is no independently registered top-level `metadata`, `properties`, `patch`,
`pin`, `unpin`, `get`, or `validate` command in this source inventory. Metadata,
properties, patches, and pinning belong to existing commands/flags. They should
not be counted as extra verbs merely because the data model supports them.

## Mixed operations that should not be mistaken for ID polymorphism

| Family | Boundary and recommendation |
|---|---|
| `graph`, `dep tree`, `dep cycles`, `children` | Operations on connectivity, often rooted at a Bead, involving both kinds. Keep a graph/domain namespace. `bd link tree` is possible vocabulary, but duplicating graph algorithms under each resource kind adds little. Generic traversal currently requires a Bead root. |
| `types` | Type discovery, not “show the Type of either kind of Resource.” A Link-scoped type list is a useful convenience. If Type definitions become Beads, CRUD on those definitions remains Bead CRUD even when they describe Links. |
| `schema` | Prints output schemas from code, without opening a store. A global schema document naturally includes both Issue and dependency schemas. Link-schema selection need not become a second top-level schema operation. |
| `status` / `stats`, `info` | Workspace summaries/capabilities. An aggregate containing both kinds is not a polymorphic ID command. |
| `statuses` | Issue lifecycle vocabulary. No automatic Link counterpart. |
| `events` | Durable workspace journal; dependency mutations are included among Issue-related writes. Cross-kind event order is valuable. Keep a workspace feed; add filters only when graph event support exists. |
| `audit` | Explicit agent/tool interactions in an append-only JSONL sidecar, not Resource audit history. No automatic `link audit`. |
| `diff` | Dolt commit/branch comparison, currently Issue-oriented output. It differs from exact versions of one Resource (`compare`). Keep workspace scope even if Link changes are later included. |
| `export`, `import` | Data movement can include connected Beads and Links. A coherent graph transfer should remain possible; selection by kind is a filter, not necessarily another command namespace. Current support must be evaluated separately from future transfer design. |
| `batch` | Transaction/work orchestration, not a resource kind. Namespace choice alone cannot settle its grammar or atomicity; this discussion leaves the earlier batch design deferral in place. |
| `sql`, `vc`, `dolt`, `branch`, `conflicts`, `backup`, `sync`, `federation`, `repo` | Query/storage/version-control/workspace operations. Do not duplicate just because stored data includes both kinds. |
| `purge`, `prune`, `compact`, `gc`, `flatten` | Existing retention/compaction/storage operations, often with Issue lifecycle rules. Deleting associated edges does not make them per-Link CRUD. A future graph-retention policy needs a separate design. |
| `rules` | Audits/compacts `.claude/rules` Markdown, not Link Type rules. No Link counterpart. |

Sources: [graph traversal][graph], [dependencies][dep], [types][types],
[schema][schema], [status][status], [statuses][statuses], [events][events],
[audit][audit], [diff][diff], [export][export], [import][import], [batch][batch],
[purge][purge], [prune][prune], [compact][compact], [gc][gc], [flatten][flatten],
[rules][rules]. These recommendations do not imply that workspace commands are
currently admitted in graph mode; consult the [admission gate][admission].

## Complete root-registration triage

The catalog below covers the static root registrations in the inspected Preview
source, including hidden compatibility and internal commands. Aliases such as
`stats` do not represent separate operations. Cobra-generated `help` and
`completion` are CLI tooling, not Resource operations. Nested commands such as
`wisp` under another namespace are not additional roots. This is a source
inventory, not a claim that every command is shown in normal help or admitted in
every workspace mode.

| Category | Root command spellings |
|---|---|
| Core resource operations (10 proposed parallel verbs) | `compare`, `count`, `create`, `delete`, `list`, `query`, `search`, `show`, `update`, `versions` |
| History compatibility alias | `history` |
| Conditional resource extensions (see decisions above) | `comment`, `comments`, `duplicate`, `duplicates`, `edit`, `find-duplicates`, `label`, `lint`, `note`, `provenance`, `rename`, `restore`, `stale`, `supersede`, `tag` |
| Relationship and topology commands | `children`, `dep`, `graph`, `link`, `links`, `unlink` |
| Issue/domain workflows; no generic Link counterpart | `assign`, `blocked`, `close`, `create-form`, `defer`, `epic`, `gate`, `heartbeat`, `human`, `mail`, `merge-slot`, `mol`, `orphans`, `priority`, `promote`, `q`, `ready`, `reclaim`, `recompute-blocked`, `reopen`, `set-state`, `ship`, `state`, `swarm`, `todo`, `unclaim`, `undefer` |
| Memory workflows; Bead-side convenience commands | `forget`, `memories`, `recall`, `remember` |
| Workspace/catalog/schema/event/transfer/retention operations | `audit`, `backup`, `batch`, `branch`, `compact`, `conflicts`, `diff`, `dolt`, `events`, `export`, `federation`, `flatten`, `gc`, `import`, `info`, `prune`, `purge`, `repo`, `rules`, `schema`, `sql`, `status`, `statuses`, `sync`, `types`, `vc` |
| Setup, integrations, configuration, diagnostics, developer tooling, internal and compatibility utilities | `admin`, `ado`, `bootstrap`, `claude-hook`, `codex-hook`, `config`, `context`, `cook`, `cursor-hook`, `db-proxy-child`, `doctor`, `formula`, `github`, `gitlab`, `hooks`, `init`, `init-safety`, `jira`, `kv`, `linear`, `metrics`, `migrate`, `migrate-issues`, `migrate-personal`, `notion`, `onboard`, `ping`, `preflight`, `prime`, `quickstart`, `rename-prefix`, `send-metrics`, `serve`, `setup`, `upgrade`, `version`, `where`, `worktree` |

This static scan finds 127 explicitly registered root spellings, before
Cobra-generated help/completion and declared aliases. It includes internal and
hidden roots and should not be compared directly with a visible-help count.

## Tradeoffs and proposed decision

A consistent scoped resource surface has the simplest rule: Bead verbs at the
root, Link verbs under `link`. The Link namespace can have precise endpoint and
Link-Type help without teaching users when a shared command changes meaning.
It also prevents incomplete symmetry from turning some Link operations into
surprising exceptions.

The cost is an extra token and explicit kind dispatch in scripts that receive a
mixture of IDs. Shared implementations, output envelopes, predicates, and error
contracts can keep the two surfaces aligned; syntactic duplication need not
be implementation duplication. Retaining root compatibility dispatch for the
five existing polymorphic spellings reduces migration pain, provided help and
examples consistently teach the new form.

A split policy—CRUD scoped but versions/compare polymorphic—has a defensible
principle (“mutations and collection discovery are scoped; universal inspection
is shared”), but show is also universal inspection and history is an alias.
Users then need to memorize exceptions. Either articulate and test that rule
carefully or use the parallel ten-verb surface plus the history alias.

The recommended boundary is therefore **kind-scoped canonical resource
operations, mixed workspace/graph operations where appropriate, and explicit
compatibility routes for already-supported root Link selectors**. Preserve
multiple Bead Types at the root. There is no need to turn Type-definition Beads
into a third kind or put their CRUD under `link` merely because they describe a
Link Type.

## Source references

[stable-link]: https://github.com/gastownhall/beads/blob/c1c4b642ac1c08d8c828007a1c2f96e47e43ef7c/cmd/bd/link.go#L12
[create]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/graph_preview_issue.go#L23
[list]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/graph_preview_list.go#L26
[show]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/graph_preview.go#L669
[update]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/graph_preview_memory.go#L12
[delete]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/graph_preview_memory_delete.go#L15
[lifecycle]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/graph_preview_lifecycle.go#L13
[versions]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/graph_preview_versions.go#L80
[compare]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/graph_preview_compare.go#L12
[discovery]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/graph_preview_discovery.go#L27
[history]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/history.go#L22
[edit]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/graph_preview_edit.go#L19
[admission]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/graph_preview.go#L367
[graph]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/graph_preview_generic.go#L36
[label]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/label.go#L1
[tag]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/tag.go#L1
[comments]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/comments.go#L1
[comment]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/comment.go#L1
[note]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/note.go#L1
[provenance]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/provenance.go#L1
[lint]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/lint.go#L1
[rename]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/rename.go#L1
[restore]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/restore.go#L1
[stale]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/stale.go#L1
[duplicates]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/duplicates.go#L1
[find-duplicates]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/find_duplicates.go#L1
[duplicate]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/duplicate.go#L1
[dep]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/dep.go#L1
[types]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/types.go#L1
[schema]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/schema.go#L1
[status]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/status.go#L1
[statuses]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/statuses.go#L1
[events]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/events.go#L1
[audit]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/audit.go#L1
[diff]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/diff.go#L1
[export]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/export.go#L1
[import]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/import.go#L1
[batch]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/batch.go#L1
[purge]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/purge.go#L1
[prune]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/prune.go#L1
[compact]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/compact_dolt.go#L1
[gc]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/gc.go#L1
[flatten]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/flatten.go#L1
[rules]: https://github.com/donnabox/beads/blob/63733f4d005611c1d6d19c8707f651ac16e441cc/cmd/bd/rules.go#L1
