---
title: "Import ordinary Beads data into a graph workspace"
description: "Load a supported ordinary export into a fresh graph preview workspace."
---

This independent import candidate targets Preview 3 and is pending integration and release qualification.
Use its build of `bd import` to copy an ordinary Beads export into a new graph workspace.
The complete import either commits or rolls back. Validate the file with
`--dry-run` before importing it.

```bash
# Run in a newly created disposable directory.
bd init --graph-mode link --scope-url https://example.invalid/new-project/ \
  --skip-hooks --skip-agents --non-interactive
bd import old-issues.jsonl --dry-run --json
bd import old-issues.jsonl --json
bd show legacy-b --json
bd comments legacy-b --json
```

`bd import FILE`, `bd import -i FILE` and `bd import -` use the same format and
validation. With no source argument, the configured ordinary import path is
used inside this workspace. Redirected stdin requires `-`. Use `--json` for
counts, source-to-graph identity maps and an explicit history disposition.
Generated Link IDs in a dry-run are provisional and may differ on apply;
`generated_ids_provisional` identifies that case. Issue and Memory source mappings
remain exact. Supplied Memory rows need no import inclusion flag. If the source IDs are ambiguous,
use the original configured prefix on both dry-run and apply:

```bash
bd import old-issues.jsonl --source-prefix web-app --dry-run --json
bd import old-issues.jsonl --source-prefix web-app --json
```


The same command works in embedded mode and an ordinary shared-server graph
workspace initialized with `--server --external`.

The destination must be empty and fresh: existing Beads, Links, native rows or
reserved/deleted identities refuse import. This includes orphan graph payloads
and Links, native labels, event/provenance feeds, journals, leases, snapshots,
previous identity/sequence allocations (including retained minted epoch addresses),
interaction records, counters and wisp auxiliary rows, and classic KV Memories even
when their values are empty. Import preserves these rows by refusing the
destination; it does not repair or adopt them. Ordinary infrastructure
configuration is allowed, including repository/clone bookkeeping, routing and
workflow settings. These settings are not proof that a workspace has never been
used; the gate checks the retained data and identity planes listed above.
The store epoch may be absent or initialized to epoch 1 without bump markers.
A changed epoch or any recorded bump marker refuses import even if all resources
were deleted. Repeating a successful import
refuses without updating or duplicating anything. `--allow-stale` and `--dedup`
are unavailable in graph mode. Read-only invocations and migration freezes
refuse both apply and dry-run. Dry-run validates by staging the same operations
in a transaction that is always rolled back; it does not reserve IDs.

## What the existing format carries

Ordinary `bd export` writes one JSON object per line. Current producers mark
Issue lines with `_type: "issue"`; earlier Issue lines without that marker are
also accepted. With `--include-memories`, exports additionally contain
`_type: "memory"`, `key` and `value` records. A standalone historical `_schema`
header with `_schema: "beads-jsonl/1"` is accepted; optional `_dolt_branch`,
`_dolt_commit`, `_project_id` and `_sort` provenance strings are read as a header
and do not recreate source history. Blank lines are ignored. Other schema versions/provenance keys refuse; graph import deliberately admits a narrower subset than ordinary import. Empty or header-only sources refuse.

| Data | Graph import behavior |
| --- | --- |
| Issue IDs | Required, canonical Scope-relative identities. Without a root, `ID` becomes `SCOPE/ID`; literal `beads/`, `links/` or `types/` source segments are preserved, subject to exact identity reservations below. The result maps each original source ID to its accepted URL. All source Issues must share one ordinary prefix; ambiguous IDs require `--source-prefix PREFIX`. `source_issue_prefix` records that source configuration and `issue_prefix` records the effective destination configuration. Unrooted import adopts the source prefix atomically; a rooted import preserves the destination's future-create prefix. Mounted native Issue IDs must fit the existing 255-byte backing limit. |
| Reserved identities | The Scope discovery ID `bdp.json` and installed Type IDs are reserved, including concrete native Issue and Dependency Types needed by this package. A collision with any such identity refuses the whole import, even if the Type is needed only by a later record. An explicit `--id-root` can move an otherwise valid source ID away from a reservation; a root that creates a reserved identity also refuses. |
| Current Issue content and workflow fields | Preserved using the ordinary Issue writer. Missing status/type default to open/task; missing timestamps are filled at import. Omitted priority remains P0, matching ordinary import. Custom Issue classes/statuses are not installed from the source export; this converter admits its supported built-in vocabulary. Invalid statuses, types or values refuse. Fields that storage cannot represent exactly refuse the whole import. |
| Issue metadata | Preserved as common graph metadata. It must meet graph JSON-object validation; array/scalar metadata and duplicate keys refuse. JSON formatting and object key order are canonicalized. |
| Labels | Preserved as a set. Storage representation that loses values refuses. |
| Dependencies | Same-prefix local edges between Issues in the same source package preserve every native-valid kind, including `blocks`, `parent-child`, `relates-to` and custom kinds (nonempty, at most 32 bytes). The native adapter selects the concrete Dependency Type and preserves `dep_type`, direction, creator and supplied timestamp. Each pair gets a new graph Link identity from the shared short-ID allocator. `dependency_ids` maps original source pairs to those URLs. Native surrogate IDs remain internal and deterministic from the mounted pair; they are not graph Link identities. Dry-run Link IDs are provisional. |
| Dependency metadata and threads | Only absent/empty-object metadata and empty thread IDs are accepted. Invalid dependency kinds, missing targets, parent/child hierarchy conflicts, self-edges and duplicate pairs refuse. Cycles in `blocks`, `conditional-blocks` and `parent-child` refuse; cycles in other kinds, including `relates-to` and `waits-for`, are accepted, matching native dependency validation. No edge is silently skipped. |
| Comments | IDs, Issue ownership, authors, text and supplied timestamps are preserved in the separate current comment feed. Older int64 numeric IDs become their decimal string spelling. Missing IDs are derived by the ordinary writer. Comments do not become retained Issue versions. Duplicate IDs/content or unrepresentable values refuse rather than disappear. |
| Legacy memories | Every supplied Memory row is processed by default. An admitted key `KEY` becomes the exact local public ID (beneath an explicit root, if supplied); `VALUE` becomes the exact body, including an empty string. The independent title uses the shared body summary: first trimmed line, collapsed spacing, at most 80 Unicode characters including an ellipsis; an empty summary uses `Untitled memory`. Metadata is empty. Blank, canonically incompatible, duplicate or colliding keys refuse the package. There is no automatic encoding, digest name, suffix, title lookup or separate key alias. `memory_ids` maps original keys to accepted URLs. |
| Counts, content hashes and readiness | Exported dependency/dependent/comment counts and readiness are projections; recomputed from accepted data. |
| Source history | Ordinary export contains no retained revisions, events or Dolt history. Each imported Bead and Link starts new retained graph history. Initial Issue/Memory writes name the importing actor; Link attribution carries the source dependency creator and timestamp when supplied; otherwise it uses the importing actor and import time. Earlier source history is neither reconstructed nor advertised as preserved. |
| Compound lineage | Nonempty `bonded_from`, `source_formula` and `source_location` refuse because ordinary storage does not persist them. |
| Ephemeral, no-history, live leases and explicit storage-class markers | Refused (including an explicit `versioned` marker that the ordinary writer normalizes away). Import does not recreate active claims or change storage planes. |
| Tombstones and graph records | Refused. This conversion imports the existing ordinary format only. Graph-format import is separate. |
| Unrecognized fields | Refused by default. Explicit `--ignore-unknown-fields` can discard only unrecognized fields, with bounded field paths and counts on stderr. Known unsupported data and malformed input still refuse. |

Admission rejects input above 16 MiB, JSON deeper than 128 levels, and more than
1,000 total Beads and Links or 1,000 total comments. The comment ceiling is separate from the Bead/Link count. These are rejection ceilings, not guaranteed accepted
sizes: persisted payloads, relations and retained snapshots must also fit the
16 MiB current-read budget. That check can refuse a smaller input after staging
it; the entire transaction rolls back. Byte/resource ceilings use
`capability_unavailable`; malformed JSON, including excessive nesting, uses
`invalid_properties`. Import does not split a large input into partially
committed chunks. Definite engine failures and dry-runs roll back before
commit. If commit or rollback cleanup is uncertain, import conservatively reports
`outcome_unknown` (exit 6) without automatic replay. An unsuccessful rollback,
including an already-ended driver transaction, does not provide a verified
rollback acknowledgment. Inspect the destination after that error before
deciding what to do next.

During freshness and prefix inspection, cancellation or the import's five-minute
storage deadline uses `capability_unavailable` (exit 5), with the inspected plane
in the message. This classification is specific to the legacy-import route.
These interrupted inspections do not diagnose a damaged or
nonempty workspace. Shared binding/schema admission precedes these checks and
currently hides some underlying error causes: interruption there can still
report `invalid_store`. This inherited limitation is not a schema-repair verdict
established by the importer. A definite rollback acknowledgment permits retrying
the operation; if the error also reports `outcome_unknown`, inspect the
destination first as described above. A definitely missing or unreadable
prerequisite also uses `invalid_store`; import does not repair it.

This is a fresh-workspace interchange operation. It does not migrate an
existing workspace in place or carry full backup/restore history.

## Optional placement beneath an ID root

Use `--id-root` to place this package beneath an explicit Scope-relative path:

```bash
bd import old-issues.jsonl --id-root restored/team --dry-run --json
bd import old-issues.jsonl --id-root restored/team --json
```

For Scope `https://example.invalid/project/`, source Issue `old-a` becomes
`https://example.invalid/project/restored/team/old-a`. A source Memory key
`beads/note` becomes `https://example.invalid/project/restored/team/beads/note`;
the literal `beads/` is preserved. Newly generated Link IDs receive the same root
once. Internal endpoints, comment ownership, owned Links and initial retained
state use the accepted identities; registered Type identities stay shared.

The root must already be a canonical relative path. Empty, absolute, escaping
or otherwise incompatible roots refuse; nothing is silently cleaned or encoded.
A root that would classify native local dependencies as external also refuses.
The root affects only this import, preserves the destination's configured
future-create prefix, and does not enable import into an occupied workspace.
Results include `id_root`, `source_issue_prefix` and the effective `issue_prefix`.

A successful dry-run reserves nothing. Only source-derived Issue and Memory
mapping is stable across dry-run and apply; new Link IDs and revision tokens
remain provisional. A collision or compatibility refusal leaves the whole
package unapplied.

## Explicitly discarding unknown fields

Use this option only when you have reviewed the fields a newer ordinary producer
added and accept losing them:

```bash
bd import old-issues.jsonl --ignore-unknown-fields --dry-run --json
bd import old-issues.jsonl --ignore-unknown-fields --json
```

The default remains strict. This option discards unrecognized fields in ordinary
Issue, dependency, comment and key/value Memory records. Arbitrary Issue metadata
keys and their values are preserved. Known fields are still validated and known
unsupported data still refuses the whole package, including live leases,
ephemeral state, lineage, unsupported relationship payloads and identity
collisions. Malformed known values always refuse. Invalid JSON, UTF-8,
duplicate keys and excessive nesting refuse even inside a field that would
otherwise be discarded.
Schema headers and record kinds remain strict; unknown header provenance,
reserved graph Resource/package fields and case aliases are never discarded.
This option cannot convert a graph package into legacy Issues. Detail/list
completeness markers such as `comments_omitted` refuse even when `false`; those
projections are not the ordinary export record family.

Warnings go to stderr even with `--json`; stdout keeps its normal result format.
Warnings name paths and occurrence counts without source values or record IDs.
Array elements share a path. An unknown object counts once for its containing
field. Up to 32 distinct paths are displayed in sorted order; each path component
is limited to 64 UTF-8 bytes and encoded with ASCII JSON escapes, so terminal
controls cannot alter the warning. Truncation notices include a stable short
full-path digest to distinguish names sharing the same prefix. The warning
reports the original distinct-path total and occurrence counts, including omitted
paths. Warnings describe
parsing only: they do not mean the destination admitted or committed anything.
The complete source must parse before warnings are emitted. If a warning cannot
be written, the command reports `output_error` (exit 5), stops before storage
admission, and imports nothing.

The flag belongs only to legacy conversion into a graph workspace. Supplying it
outside graph mode refuses before opening ordinary storage, including explicit
`--ignore-unknown-fields=false`. In graph conversion, explicit `false` keeps the
strict default. Ordinary dependency-workspace import behavior is unchanged.
