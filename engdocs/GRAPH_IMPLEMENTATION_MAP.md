# What the Beads graph preview can demonstrate

This map is pinned to **qualified candidate `b7bf5040beae863b0d4fafa8d4b838dd04b3cd20`**, on `donnabox/beads` branch `codex/janet-graph-integration`, as recorded in the September 27, 20:45 PDT handoff. This is a pinned capability inventory, not a claim about every later branch. The [short demo](GRAPH_MIXED_DEMO.md) supplies a separately recorded scenario on this runtime.

**The useful demo is already real:** create Issues and Memories through the installed CLI, connect them with two kinds of Links, edit them, see blocking/readiness change, retrieve saved versions, and read the same graph through an independent public BDP client. It is a disposable preview of selected workflows. It is not full Memory conformance, adoption of an existing Beads database, or complete BDP.

The source of truth for the current executable surface is [CLI admission and capabilities at the qualified commit](https://github.com/donnabox/beads/blob/b7bf5040beae863b0d4fafa8d4b838dd04b3cd20/cmd/bd/graph_preview.go). Some earlier slice documents retain historical “not yet implemented” lists. In particular, later slices added Dependency unlink, Memory edits, comparison, BDP serving and assignee editing; those older lists must not be mistaken for current limits.

## Which branch supplies each behavior?

| Source | Qualification and what it adds |
|---|---|
| **Qualified candidate: PR46, `b7bf5040…`** | All capabilities in the main map below. Exact-source Linux run `36367186429` was already verified: 4,185 command receipts, 553 required graphstore test names, 106 Issue CLI names, seven storage packages, 70 real HTTP requests, zero remaining children. C0 receipts cover embedded 360 commands/56 checks and ordinary-server 367/57, with no stated qualification gaps. |
| **Draft PR47: `3a827e91ed634f5dc4285227d606f6697059f592`** | Adds standalone `bd update beads/work --claim --actor alice --json`. It reuses the Issue claim/lease writer and complete graph retention. Five-minute lease, no graph heartbeat/renewal/unclaim or drift repair. Parent runtime remained unchanged in the latest correction; two obsolete negative harness expectations were fixed. Exact-source qualification remains pending in the referenced checkpoint. |
| **Draft PR48: `562d62be840deccf7e4cd1123ffb19e851d281ab`** | Adds `bd update beads/work --append-notes TEXT --if-revision REV --json` or explicit `--unconditional`, including mixed admitted scalar edits. Includes PR47's fixture correction. Corrected local proof covers 256 append commands, 242 older edit commands, 80 HTTP requests and 53 new storage test names. A real read-budget defect was reproduced and corrected; changed append requests roll back if the complete current workspace would exceed the existing read budget. Exact-source Linux qualification remains pending. |
| **Held PR44/45 label work** | Separate from this qualified candidate and the claim/append path. Technical qualification does not resolve the human holds on widening ordinary RowVersion behavior and staging the whole pending Issue table. Do not present graph label replacement as available on this candidate. |

Evidence: [qualified PR46](https://github.com/donnabox/beads/pull/46) and its [exact-source Linux run](https://github.com/donnabox/beads/actions/runs/36367186429). Drafts: [PR47](https://github.com/donnabox/beads/pull/47), [PR48](https://github.com/donnabox/beads/pull/48). Their status above is the September 27, 20:45 PDT checkpoint; consult those reviews for subsequent qualification.

## Qualified CLI capabilities

Commands below are supported spellings, not a newly executed script. Use a **fresh disposable workspace** and an installed binary built from the qualified commit. `REV`, `LINK_REV` and `SOURCE_REV` mean actual opaque tokens returned by prior reads. `SCOPE` means the persisted Scope URL, with its trailing slash; Type URLs must use that Scope.

| What a person can do | CLI spelling | What the operation means |
|---|---|---|
| Initialize a graph workspace | `bd init --graph-mode link --scope-url https://example.invalid/demo/ --prefix demo --non-interactive --skip-hooks --skip-agents` | Normal initialization creates the preview in ordinary Dolt. Embedded works for CLI use. Shared-server initialization adds `--server --external --server-host 127.0.0.1 --server-port PORT --server-user root --database NAME`. No manual SQL bootstrap. |
| Inspect support | `bd status --graph --json` | Reports admitted capabilities and bounds. Full `memory`, `historyExact`, `issueWorkflows`, `requestStatus` and `backupContinuity` remain false. |
| Create an Issue | `bd create 'Ship the change' --id beads/work --priority P1` | Creates an Issue-typed Bead with a canonical identity and retained complete state. This is a bounded durable-Issue route, not all legacy create options. |
| Read any admitted Bead or Link | `bd show beads/work --json`; `bd show links/context --json` | Complete experimental record, including opaque revision/version, properties, attribution and saved owned Links where applicable. Local canonical paths and exact same-Scope URLs work; legacy short IDs, fuzzy matching and aliases do not. |
| Edit an Issue | `bd update beads/work --title 'Ship safely' --description 'Plan' --design 'Design' --acceptance 'Checks' --priority P0 --assignee alice --if-revision REV --json` | Each supplied field changes; omitted fields stay unchanged. Text, priority and ordinary assignment can share one transaction and one retained version. `--unconditional` is an explicit alternative to the guard. Stale guards refuse even for otherwise equal values; no-ops keep the version. |
| Clear ordinary assignment | `bd update beads/work --assignee '' --if-revision REV --json` | Clears the assignee using existing domain rules. This is not claim/unclaim. Clearing an active assignment can leave `in_progress` with no assignee or lease; another actor's active assignment is fenced. |
| Find Issues | `bd list --flat`; `bd list --format records-json --all --sort priority`; `bd list --format records-json --assignee alice --all`; `bd list --format records-json --no-assignee --all` | Issue-only lists use existing Issue filters/query policy. Status/type, title, priority/ranges, labels and pinned filters are supported, plus assignee/unassigned. Explicit limited pages report `hasMore`; repeated calls are new reads, not snapshot continuation. Bare tree output and legacy `--json`/`--format json` refuse. |
| Close/reopen an Issue | `bd close beads/work --reason 'Done' --json`; `bd reopen beads/work --reason 'Needs another pass' --json` | Reuses domain transitions and retention. Repeated effective no-ops do not mint a version. These workflow commands do not introduce a mandatory revision guard. Reopening recomputes readiness and clears closure fields. |
| Find ready work | `bd ready --json` | Existing Issue scheduling query, not “all open Issues.” A Dependency can block an open Issue. |
| Create a Memory | `bd remember 'Why we chose this approach' --id beads/plan --title 'Plan' --json` | A non-Issue Bead. Explicit identity/title are required by this preview; bodies are UTF-8 and can be empty. |
| Read Memory content | `bd recall beads/plan`; `bd recall beads/plan --version TOKEN` | Emits body bytes without decoration or an added newline. Structured `recall --json` refuses because the full proposed Memory representation is incomplete. Use `show --json` for the experimental complete stored record. |
| Create/edit from a file or pipe | `bd remember --id beads/plan --title Plan --body-file plan.md`; `bd remember --update beads/plan --if-revision REV --stdin` | Explicit file/stdin input; no frontmatter parsing, inferred title or implicit upsert. `--body-file -` means a file literally named `-`. |
| Edit selected Memory fields | `bd remember --update beads/plan --title 'New title' --if-revision REV --json`; `bd remember 'New body' --update beads/plan --unconditional --json` | Supplied title/body change atomically; omitted fields and complete owned Links are preserved inside the checked transaction. A missing selected Memory does not create one. |
| Replace Memory properties | `bd update beads/plan --properties '{"title":"Plan","body":"New body"}' --if-revision REV --json` | Entire admitted properties object: exactly title/body strings. `@file` and `@-` also work. This is not arbitrary JSON Patch or full BDP Update. |
| Discover Memories | `bd memories 'deployment' --details`; `bd memories 'deployment' --format records-json --all` | Literal Unicode case-folded substring search of current title/body. Returns one summary per Memory, canonical ID, saved version and attribution, optional bounded excerpt and owned-Link count. No ranking, target-body expansion or paged CLI continuation. |

Source guides: [Issue editing](https://github.com/donnabox/beads/blob/b7bf5040beae863b0d4fafa8d4b838dd04b3cd20/engdocs/GRAPH_ISSUE_EDIT_PREVIEW.md), [assignment and filtering](https://github.com/donnabox/beads/blob/b7bf5040beae863b0d4fafa8d4b838dd04b3cd20/engdocs/GRAPH_ISSUE_ASSIGNEE_PREVIEW.md), [Issue listing](https://github.com/donnabox/beads/blob/b7bf5040beae863b0d4fafa8d4b838dd04b3cd20/engdocs/GRAPH_ISSUE_LIST_PREVIEW.md), [selected Memory edits](https://github.com/donnabox/beads/blob/b7bf5040beae863b0d4fafa8d4b838dd04b3cd20/engdocs/GRAPH_MEMORY_SELECTED_UPDATE_PREVIEW.md), [discovery](https://github.com/donnabox/beads/blob/b7bf5040beae863b0d4fafa8d4b838dd04b3cd20/engdocs/GRAPH_MEMORY_DISCOVERY_PREVIEW.md). Current capabilities override older “assignment unavailable” prose in the earlier Issue-edit slice.

## Two kinds of Links, one visible graph

**Blocking Dependencies connect Issue Beads.** `bd dep add beads/work beads/check --json` means “check blocks work.” The familiar assertion is unguarded and an already-present pair is a no-op. The generic spelling can select the installed `SCOPEtypes/preview-blocks-v1` and guard its owning source with `--if-source-revision`. The source Issue owns outgoing blocking Dependencies: adding/removing one versions the source's complete owned set. Closing its prerequisite changes readiness without rewriting the dependent's saved content.

**Informational Links can connect Memory↔Memory, Issue→Memory and Memory→Issue.** For example:

```sh
bd link beads/plan beads/work --resource-type "${SCOPE}types/preview-related-v2" \
  --id links/context --properties '{"note":"Explains the work"}' \
  --if-source-revision SOURCE_REV --json
bd links beads/plan --direction out --json
bd update links/context --properties '{"note":"Revised context"}' \
  --if-revision LINK_REV --if-source-revision SOURCE_REV --json
bd unlink links/context --if-revision LINK_REV --if-source-revision SOURCE_REV --json
```

The private informational Type accepts an optional string `note`; properties replacement is whole-object. Identity, Type and endpoints are immutable. Different IDs with the same endpoints remain separate Links. Memory owns all outgoing Links, so an owned-Link change versions that Memory; its distinct target is unchanged. Issue informational Links are provisionally **unowned**, so editing them does not version either endpoint. This ownership difference is visible behavior, not a settled universal policy.

`bd links BEAD --direction in|out|both` includes both specialized Dependencies and informational Links. An exact `--resource-type URL` can filter it. It returns the complete bounded set, not expanded endpoint content or a paged CLI traversal. Informational typed-pair unlink is supported but refuses ambiguity with candidate IDs. Blocking unlink is **canonical ID only**, requires both Link/source guards (or their explicit unconditional alternatives), and delegates to the Dependency writer; `bd dep remove` and blocking pair unlink remain unavailable. Deleted IDs stay reserved; current CLI reads return `gone`; old live versions remain readable.

Changed unconditional Memory edits report the actual prior Memory/version/attribution as `replaced`. Changed Memory-owned Link writes using `--unconditional-source` report `replacedSource`. Guarded writes, no-ops, failures and uncertain outcomes do not invent a successful overwrite receipt. See [unlink](https://github.com/donnabox/beads/blob/b7bf5040beae863b0d4fafa8d4b838dd04b3cd20/engdocs/GRAPH_DEPENDENCY_UNLINK_PREVIEW.md) and [overwrite disclosure](https://github.com/donnabox/beads/blob/b7bf5040beae863b0d4fafa8d4b838dd04b3cd20/engdocs/GRAPH_OVERWRITE_DISCLOSURE_PREVIEW.md).

## Saved state works; full History is still missing

`bd show RESOURCE --version TOKEN --json` retrieves a complete retained Memory, Issue or live Link version, including its saved owned set. `bd compare RESOURCE --from OLD_TOKEN --to NEW_TOKEN --json` compares all serialized properties and complete owned Links; it does not infer chronology from token spelling. Historical Memory recall can return the old body. These make an edit-and-recall demo useful today.

There is no public ordered version enumeration, as-of selection, restore, alias/version-URL resolution, erasure, generic Memory deletion or public HTTP History. A deleted Link's private tombstone token is not a Resource version. Current timestamps are recorded attribution data, **not proven native commit timestamps**. The bounded storage investigation found no supported exposed primitive that atomically binds an engine-observed commit stamp into retained context on both required backends. That prerequisite remains open; a Go clock, SQL NOW or postcommit repair is not a substitute. [Exact reads](https://github.com/donnabox/beads/blob/b7bf5040beae863b0d4fafa8d4b838dd04b3cd20/engdocs/GRAPH_HISTORY_EXACT_PREVIEW.md), [comparison](https://github.com/donnabox/beads/blob/b7bf5040beae863b0d4fafa8d4b838dd04b3cd20/engdocs/GRAPH_HISTORY_COMPARE_PREVIEW.md), [History review gate in fork PR18](https://github.com/donnabox/beads/pull/18).

## What BDP actually exercises

On **ordinary shared-server Dolt only**, initialize the Scope as the actual reachable URL, such as `http://127.0.0.1:8765/demo/`, then run `bd serve --readonly --addr 127.0.0.1:8765`. The persisted URL is canonical identity; choosing another free port later does not rename the graph. Embedded HTTP serving is refused; embedded CLI use is supported.

The real Read service provides discovery (`/demo/` and `/demo/bdp.json`), canonical Bead/Link/Type resources, properties views, inventories, structural/Selector filtering, incident Links and aggregates. GET/HEAD, revision ETags and conditional requests are implemented. Collection `next` URLs preserve the selected snapshot while later CLI writes change fresh reads. Those cursors are process-local, expire after five minutes and disappear on restart; they are not durable History.

The qualified proof uses the independently built public client pinned to **gastownhall/bdp `53bdbd03136875f952af184fce7b3c7af8f74e96`**. The Beads CLI performs writes; the client reads the changed graph and retained continuation state over actual HTTP. Aggregate/replay checks use real fetch plus public parsers because this client pin lacks an aggregate API. **There are no HTTP mutation endpoints or BDP Write/Update/History profiles delivered here.** This map makes no claim about unrelated BDP-repository implementation. [HTTP guide and harness](https://github.com/donnabox/beads/blob/b7bf5040beae863b0d4fafa8d4b838dd04b3cd20/engdocs/GRAPH_BDP_READ_HTTP.md).

## Where the data lives and why

All of this uses the ordinary selected **Beads Dolt database**, on the embedded engine or released ordinary server. There is no separate Memory database and no second editable Issue/Dependency copy.

| Data | Authoritative storage |
|---|---|
| Issue properties, labels and domain relationships | Existing specialized Issue machinery and tables, including `issues`, normalized label relations and `dependencies`. Existing indexes and domain validation are reused. |
| Generic Memory content | `graph_preview_payloads`. |
| Informational Link endpoints/properties | `graph_preview_links`. |
| Shared canonical identity/current revision/allocation | `graph_preview_catalog`; Scope/workspace binding and installed private Types use `graph_preview_scope` / `graph_preview_types`. |
| Retained generic records and Link snapshots | `graph_preview_versions`. |
| Retained Issue body | Existing contributor recorder's `issue_versions.durable_state`; `graph_preview_issue_versions` maps the graph token to that saved Issue revision and stores its complete owned-Link state. |

The catalog unifies identity and enumeration; it does not require moving all Issue fields into generic JSON. Admitted domain writes, graph mapping and retained versions commit together. The approach preserves mature Issue queries while integrating their results into one graph. Future Jim-writer integration must reconcile who calls the retained recorder so each actual mutation still records exactly once and no-op behavior stays correct. [Storage schema](https://github.com/donnabox/beads/blob/b7bf5040beae863b0d4fafa8d4b838dd04b3cd20/internal/storage/graphstore/schema.go), [Issue adapter](https://github.com/donnabox/beads/blob/b7bf5040beae863b0d4fafa8d4b838dd04b3cd20/internal/storage/graphstore/issues.go), [exact reader](https://github.com/donnabox/beads/blob/b7bf5040beae863b0d4fafa8d4b838dd04b3cd20/internal/storage/graphstore/history_exact.go).

## Limits to state before the demo

- Fresh schema-v5 disposable workspaces only. Older previews, existing legacy adoption, copied/moved authority, backup/restore and replication continuity are not delivered. Legacy behavior outside graph workspaces is distinct from full compatibility inside them.
- The common open `metadata` field, Memory Inception/derivation, complete Memory JSON, keys/aliases and mutable user Type installation are missing. Existing Issue `properties.metadata` is not the proposed common Bead/Link metadata. Private installed Type names are provisional.
- Current graph acquisition has a conservative **16 MiB workspace budget** and inventories admit at most **1,000 live Resources before filtering**. Unrelated data can make a narrow read refuse. Owned/incident Link bounds are also 1,000. Memory body/properties input is bounded to 1 MiB. These are complete-or-refuse limits, not silent truncation or universal write-size guarantees.
- Memory discovery allows 50 matches by default, explicit `--all` within workspace bounds, 1 MiB rendered output and 160-code-point excerpts. Issue list limits and `hasMore` are a separate existing Issue-query policy.
- Full Issue parity is absent: general status edits, notes replacement/clear, graph label editing, classification, parent changes, broader claim lifecycle, bulk workflows and generic Issue JSON patch are not supplied by the qualified candidate. Claim/append are separately labeled drafts above.
- Ordinary-server Dolt 2.1.8 provisioning of different databases stays serialized. Concurrency qualification concerns operations on admitted workspaces; it does not establish safe parallel provisioning. Uncertain commit acknowledgment is reported, never silently retried.
- Human review remains necessary for provisional CLI/result/Type contracts. PR44/45 remain held; Jim's fork is the chosen integration destination but its exact target branch is unknown. No main/protected/Jim-target merge is implied by green CI or this map.

The [fork delivery plan PR18](https://github.com/donnabox/beads/pull/18) remains the implementation authority; CLI shape feedback belongs in [upstream proposal #6703](https://github.com/gastownhall/beads/issues/6703), and the full Memory destination is [proposal #5877](https://github.com/gastownhall/beads/issues/5877). The original September 23 attempt and missed commitment have not been reset. The next demonstration should name its exact binary/source pin and show these implemented operations, then show the unresolved items explicitly.
