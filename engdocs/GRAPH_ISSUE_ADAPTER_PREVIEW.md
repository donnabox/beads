# Experimental Issue create/read adapter

The original create/read slice is described below. The [Dependency workflow extension](GRAPH_DEPENDENCY_WORKFLOW_PREVIEW.md) advances the disposable schema to version 3 and the Issue descriptor to v2, adding owned blocking Dependencies, close and ready. Statements about v1/empty ownership below describe the original PR #20 boundary.

This is a reversible extension of [C0 PR #19](https://github.com/donnabox/beads/pull/19), within the original September 23–25 attempt. It begins W2 of [the fork delivery plan](https://github.com/donnabox/beads/pull/18); it does not complete W1, W2 or M1. No production contract or migration is accepted by this branch.

## Observable slice

Use a fresh disposable directory and the binary built from this branch:

```sh
bd init --graph-mode link --scope-url https://example.invalid/issue-demo --prefix demo
bd remember 'Check the deployment notes.' --id beads/plan --title Plan
bd create 'Fix deployment' --id beads/work --description 'Preserve the existing Issue machinery.' --type bug --priority 1 --labels demo
bd show beads/work --json
bd show beads/work --json
bd show beads/plan --json
```

Each command is a new process. Shared-server initialization additionally takes `--server --external --server-host 127.0.0.1 --server-port PORT`. The server must be an operator-owned disposable instance. Normal init creates the standard schema, both experimental descriptors and Issue prefix configuration. There is no seed script or hidden bootstrap.

`--id beads/work` is the canonical graph path. The returned `properties.id` is the internally allocated legacy Issue backing ID, generated through the existing prefix/ID machinery. This deliberate preview exposure is not an alias contract or a recommendation for the final generic representation. `--type` remains the Issue classification convenience. The descriptor is Scope-local `types/preview-issue-v1`; Task/Bug nominal Types, metadata placement and legacy alias presentation remain open.

Only durable Issue creation with title, description, classification, priority and labels is exposed. Unsupported flags refuse. Dependencies, custom routing, ephemeral/no-history records, metadata, Issue updates/close/ready, generic Links and historical reads remain unavailable. Existing dependency-mode workspaces retain their original command routes. This is not a claim that all Issue workflows work in graph mode.

## Storage and contributor reuse

The adapter invokes `issueops.ExecuteCreate` on the same SQL transaction that validates the graph binding and changes the common writer coordination cell. Existing `issues`, labels, events and related tables remain authoritative. Graph identity allocation has a unique specialized backing key. The graph mapping ties an opaque graph version to an Issue history row; it stores no second editable Issue payload and no second Issue snapshot.

The retained writer, deferred creation minting, canonicalization tests and attribution/LONGBLOB migration are reused narrowly from [Jim Wordelman's PR #6661 at 64becbc](https://github.com/gastownhall/beads/pull/6661), building on [his writer PR #6358](https://github.com/gastownhall/beads/pull/6358). `ScopeVersionedHistoryTransaction` activates the existing writer only for the admitted graph transaction. Creation retains exactly one complete Issue snapshot after creation-time effects. No adapter-level second mint is added. Other Issue mutation hooks and the versions CLI are not imported.

Jim's local revision ordinal stays internal. The graph API does not expose it as a portable version address. His second-resolution change timestamp is attribution data, not proof of a precise acceptance ordering. Memory versions continue using the C0 recorder; neither recorder is advertised as complete public BDP History. Generic owned-Link history remains future work.

The schema/marker advance to disposable preview version 2. Version 1 stores refuse; no conversion is supplied. SQL migrations use the normal schema machinery, including the direct-DDL CLI substitution for older Dolt. The inherited MySQL client pin remains, and no new engine fork is introduced.

## Validation contract

The installed harness extends the C0 transcript with Issue create/read/fresh-process read, exact labels/classification/body, shared Issue/Memory path collisions, and unsupported-effect refusals. Storage integration controls must separately verify exact authoritative hydration, one retained Issue snapshot, zero generic payload copies, rollback after Issue and catalog writes, and competition between Issue and Memory writers on embedded and ordinary shared-server Dolt. A passing smoke transcript alone does not establish transaction atomicity or full W2 delivery.

Remaining W1 work includes full catalog and recoverable bootstrap beyond disposal. Remaining W2 work includes shared graph/domain mutation paths, Dependencies and readiness; W3 adds mixed informational Links and Link-property updates. Open contracts must be reviewed before this experiment becomes durable.
