# Experimental blocking Dependency workflow

These notes describe the earlier bounded slice. For the current successor
(schema 4), see [mixed-Link operations and limits](GRAPH_MIXED_LINK_PREVIEW.md).

This extends [PR #20](https://github.com/donnabox/beads/pull/20) at
`0cabfa252fa136ac036e0d4e04036b88f10f9fbd`, within W2 of the
[fork delivery plan](https://github.com/donnabox/beads/pull/18).
The original attempt began September 23, 2026 at 07:01 PDT; this is not a
new checkpoint clock. Neither this slice nor create/read completes W2 or M1.

Donna explicitly agreed on September 25 that Issues own their outgoing
blocking Dependencies: each actual change advances the source Issue's version
and retains its relationships. This does not settle ownership for every Type,
deletion, restoration identity, nominal Task/Bug Types, aliases, or metadata.

## Storage decision

Issues and blocking Dependencies join one graph while retaining their existing
specialized authoritative tables. Memory uses generic payload storage in the
same Dolt database. The common catalog owns canonical allocation and current
graph revision. It does not store a second editable Issue or Dependency payload.
There is no planned prerequisite to move all Issue content into generic JSON.
The [layout comparison](GRAPH_C0_LAYOUT_COMPARISON.md) explains the retained
query/index and compatibility benefits, and the obligations of this choice.

The existing Issue writer and Jim Wordelman's retained-version writer remain
the domain path. Narrow new-edge and close hooks build on his
[PR #6661](https://github.com/gastownhall/beads/pull/6661) at `64becbc`.
`issue_versions` is the sole Issue payload snapshot store. A graph version
mapping also retains the full canonical outgoing Link state for that Issue
version. Historical membership must never be reconstructed from live Links.

A Dependency's deterministic legacy backing key is not its canonical Link ID.
It receives a separate `links/PATH` allocation. Current Link reads use the
Dependency row; retained Link snapshots supply history and consistency checks.
All admitted domain, catalog, Link and source-version effects share one SQL
transaction and the existing graph writer coordination cell.

## Installed demonstration

Use a fresh disposable directory and this branch's installed binary:

```sh
bd init --graph-mode link --scope-url https://example.invalid/workflow --prefix demo
bd create 'Release deployment' --id beads/release
bd create 'Verify deployment' --id beads/prerequisite
bd remember 'Deploy after verification.' --id beads/plan --title Plan
bd remember 'Verification catches regressions.' --id beads/rationale --title Rationale
bd ready --json
bd dep add beads/release beads/prerequisite --json
bd show beads/release --json
# Use the returned canonical Link ID with bd show.
bd ready --json
bd dep add beads/release beads/prerequisite --json
bd close beads/prerequisite --json
bd ready --json
bd close beads/prerequisite --json
```

Each invocation is a fresh process. The release is initially ready, becomes
blocked after the Dependency, then becomes ready after the prerequisite closes.
The repeated Dependency assertion and repeated close report `changed: false`
without minting content versions. Adding the Dependency changes the source
version but not the target version. Closing the target changes its version but
not the source's recorded content or owned Link: readiness is a derived view.

`bd link SOURCE TARGET` reaches the same operation. The experimental generic
form selects the installed Type explicitly:

```sh
bd link beads/release beads/prerequisite \
  --resource-type https://example.invalid/workflow/types/preview-blocks-v1 \
  --unconditional-source --json
```

`--type blocks`, `blocked-by` and `depends-on` preserve the same direction:
the second argument blocks the first. Other Dependency Types refuse in this
slice. Generic owned-Link creation requires an observed `--if-source-revision REV`
or explicit `--unconditional-source`. A stale source revision refuses even when
the pair already exists; serialization alone is not a stale-observation guard.
Familiar Dependency assertions retain their existing unguarded semantics.
`--type` and `--resource-type` cannot be combined. Selectors accept
canonical local paths or their exact full Scope URLs; no alias or remote
resolution is implemented. Shared-server init additionally takes
`--server --external --server-host 127.0.0.1 --server-port PORT`.

## Explicit preview limits

- Schema/marker version 3 and Issue descriptor `types/preview-issue-v2` are
  disposable. Earlier preview workspaces refuse; no migration is supplied.
- Only local durable Issue endpoints, a registered experimental blocking Type,
  latest endpoints and empty Link properties are admitted. Memory endpoints,
  cycles and closing a blocked Issue refuse. No force-close bypass is exposed.
- One blocking Dependency per ordered Issue pair preserves the existing domain
  assertion. This does not decide generic multiedge policy: informational Links
  with equal endpoints remain required future work.
- The provisional source ownership budget is 1,000 Links. An addition beyond
  that limit refuses before domain changes; this is not a final product limit.
- The v2 generic Issue projection omits nested legacy `dependencies`; canonical
  relationships appear in `owned`. Jim's internal Issue snapshot retains the
  authoritative legacy aggregate. The legacy `properties.id` remains exposed
  provisionally and is not a settled alias contract.
- `ready` supports the unfiltered Issue view. Filter/claim/limit flags and
  multi-Issue close refuse. A nonzero `BEADS_MAX_ROWS` also refuses rather than
  silently ignoring the configured cap. Familiar legacy-mode outputs and routes remain
  unchanged; graph-mode results use the explicit preview envelope.
- This does not add Memory updates, informational Links, Link-property updates,
  historical CLI reads, restore, migration, BDP serving or full Issue workflow
  compatibility. Recorded snapshots alone are not complete History delivery.

## Verification boundary

`scripts/graph-c0-smoke.py --dependency-workflow` exercises the installed CLI
and preserves raw command receipts and observed version transitions. Run it on
embedded and ordinary shared-server Dolt. Separate real-engine tests must prove
all-or-none rollback across domain/Link/catalog/retained writes and the
Dependency-versus-close race. A smoke transcript alone does not prove these.
The existing lost-COMMIT and cancelled-writer controls remain applicable to the
shared transaction runner; report which controls actually ran at the tested head.

Contributor intake also considered upstream [#6719](https://github.com/gastownhall/beads/pull/6719)
(post-commit blocked-state repair), [#6565](https://github.com/gastownhall/beads/pull/6565)
(manual blocked-status drift), and [#5648](https://github.com/gastownhall/beads/pull/5648)
(Dependency Type aliases). This preview shares the existing alias normalizer;
its coordinated write transactions must prevent the overlapping write anomaly
without adding an out-of-transaction repair. It does not replace or close those
contributor PRs or expand the existing doctor gate.

Installed testing exposed an additional retained-writer defect: advancing
`current_revision` fired the Issue table's automatic `updated_at` after capturing
its snapshot. Ordinal bookkeeping now explicitly preserves `updated_at`. A
deterministic real-engine test pins an older timestamp and verifies that minting
retains exactly that state; the live-versus-retained equality check is unchanged.
