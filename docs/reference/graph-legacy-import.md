---
title: "Import ordinary Beads data into a graph workspace"
description: "Load a supported ordinary export into a fresh graph preview workspace."
---

This independent import candidate is pending integration and release qualification.
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
The same command works in embedded mode and an ordinary shared-server graph
workspace initialized with `--server --external`.

The destination must be empty and fresh: existing Beads, Links, native rows or
reserved/deleted identities refuse import. Repeating a successful import
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
| Issue IDs | Required and preserved as the underlying Issue ID. The graph identity is `SCOPE/beads/ID`; the result maps each source ID to its graph URL. All Issue IDs must have one nonempty ordinary prefix and form valid local Bead paths. Import atomically sets the fresh workspace's Issue prefix to that source prefix, so later creates and dependencies continue to work; dry-run rolls that change back. |
| Current Issue content and workflow fields | Preserved using the ordinary Issue writer. Missing status/type default to open/task; missing timestamps are filled at import. Omitted priority remains P0, matching ordinary import. Invalid statuses, types or values refuse. Fields that storage cannot represent exactly refuse the whole import. |
| Issue metadata | Preserved as common graph metadata. It must meet graph JSON-object validation; array/scalar metadata and duplicate keys refuse. JSON formatting and object key order are canonicalized. |
| Labels | Preserved as a set. Storage representation that loses values refuses. |
| Dependencies | Only same-prefix local `blocks` edges between Issues in the same input are supported. Direction, creator and supplied timestamp are preserved. Each source/target pair gets a canonical Link; the result maps pairs to Link URLs. Native surrogate IDs are recomputed from the pair and become the Link path. Source dependency IDs are not portable graph identity. |
| Dependency metadata and threads | Only absent/empty-object metadata and empty thread IDs are accepted. Other dependency types, missing targets, cycles, self-edges and duplicate pairs refuse. No edge is silently skipped. |
| Comments | IDs, Issue ownership, authors, text and supplied timestamps are preserved in the separate current comment feed. Older int64 numeric IDs become their decimal string spelling. Missing IDs are derived by the ordinary writer. Comments do not become retained Issue versions. Duplicate IDs/content or unrepresentable values refuse rather than disappear. |
| Legacy memories | A key/value record becomes a Memory at `SCOPE/beads/KEY`, with title `KEY`, body `VALUE` and empty metadata. The key must form a valid local Bead path and cannot collide with another imported identity. The result maps keys to URLs. |
| Counts, content hashes and readiness | Exported dependency/dependent/comment counts and readiness are projections; recomputed from accepted data. |
| Source history | Ordinary export contains no retained revisions, events or Dolt history. Each imported Bead and Link starts new retained graph history. Initial Issue/Memory writes name the importing actor; Link attribution carries the source dependency creator and timestamp. Earlier source history is neither reconstructed nor advertised as preserved. |
| Compound lineage | Nonempty `bonded_from`, `source_formula` and `source_location` refuse because ordinary storage does not persist them. |
| Ephemeral, no-history, live leases and explicit storage-class markers | Refused (including an explicit `versioned` marker that the ordinary writer normalizes away). Import does not recreate active claims or change storage planes. |
| Tombstones, graph records and unknown fields | Refused. This stage imports the existing format only. Graph-format import follows graph export. |

Admission rejects input above 16 MiB, JSON deeper than 128 levels, and more than
1,000 total Beads and Links. These are rejection ceilings, not guaranteed accepted
sizes: persisted payloads, relations and retained snapshots must also fit the
16 MiB current-read budget. That check can refuse a smaller input after staging
it; the entire transaction rolls back. Byte/resource ceilings use
`capability_unavailable`; malformed JSON, including excessive nesting, uses
`invalid_properties`. Import does not split a large input into partially
committed chunks. Engine errors and timeout roll back before commit;
an uncertain commit outcome is reported as `outcome_unknown`, without automatic
replay. Inspect the destination before deciding what to do after that error.

This is a fresh-workspace interchange operation. It does not migrate an
existing workspace in place or carry full backup/restore history.
