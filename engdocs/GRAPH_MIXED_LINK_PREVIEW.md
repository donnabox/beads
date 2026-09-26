# Experimental mixed Links and property updates

For the schema 5 successor, see [incident listing and guarded unlink](GRAPH_LINK_LIFECYCLE_PREVIEW.md).

This bounded W3 slice builds on the qualified Dependency workflow at
`cd22fcbf4060f5f7c941d826bd570a738f748042` ([PR #21](https://github.com/donnabox/beads/pull/21)).
The [delivery plan](https://github.com/donnabox/beads/pull/18) and
[CLI proposal](https://github.com/gastownhall/beads/issues/6703) remain the review
sources. The original attempt began September 23, 2026 at 07:01 PDT.
This is further implementation within that attempt, not a restarted clock.

## What the installed CLI can exercise

Use a fresh disposable workspace and the binary from this branch:

```sh
bd init --graph-mode link --scope-url https://example.invalid/mixed --prefix demo
bd create 'Release deployment' --id beads/release
bd remember 'Deploy after verification.' --id beads/plan --title Plan
bd remember 'Verification catches regressions.' --id beads/rationale --title Rationale
bd link beads/release beads/plan \
  --resource-type https://example.invalid/mixed/types/preview-related-v2 \
  --id links/release-plan --properties '{"note":"Implementation context"}' --json
bd show links/release-plan --json
bd link beads/plan beads/rationale \
  --resource-type https://example.invalid/mixed/types/preview-related-v2 \
  --id links/plan-rationale --properties '{"note":"Original rationale"}' \
  --unconditional-source --json
bd update links/plan-rationale --properties '{"note":"Revised rationale"}' \
  --unconditional --unconditional-source --json
bd show beads/plan --json
bd show beads/rationale --json
```

Each invocation is a new process. The final plan contains the revised Link in
its complete `owned` array. Its version advances on Link creation and on the
property edit. The target rationale stays unchanged. The Issue→Memory Link
also has an independent identity and editable properties, but changes neither
endpoint version under the provisional unowned-Issue policy.

Use `--if-revision REV` and, for owned Links, `--if-source-revision REV` to
protect observed state instead of the explicit unconditional options in this
short example. A stale guard refuses the whole operation, including an
otherwise semantic no-op. A matching no-op preserves both Link and source
versions. `--unconditional=false` is not an unconditional authorization.
For an unowned Issue source, an optional source guard is checked when supplied.

`--properties` replaces the **entire** properties object. It accepts literal
JSON, `@file`, or `@-` for stdin. Inputs must be objects, with no duplicate
member names or invalid Unicode, and fit the disclosed 1 MiB preview input
budget. The private descriptor admits only optional string `note`; `{}`
removes that property. Unknown fields and non-string `note` refuse.
Identity, Type, source and target are immutable. The preview does not silently
interpret properties as a whole Resource or metadata.

Different Link IDs with the same endpoints create independent Links. Omitting
`--id` allocates a new Link. Reusing an allocated ID refuses; the preview has
no request replay protocol and must not automatically retry an uncertain write.
This differs intentionally from the existing blocking Dependency's pair-based
assertion/no-op rule. Informational Links do not affect `bd ready`.

## Storage and provisional contracts

All records share the ordinary embedded or shared-server Dolt database. Issues
and blocking Dependencies remain authoritative in their specialized tables.
Informational Links have their own authoritative endpoint/property rows under
the same canonical catalog, revision allocation and transaction boundary.
There is no second editable Dependency copy.

Memory v2 owns all outgoing Links, with the existing wildcard ownership rule
and a provisional complete-set limit of 1,000. Its current properties and
complete owned-Link set are retained atomically for each accepted source
version. Reads compare live state with its retained snapshot; old snapshots
must not be reconstructed by joining current Links. A distinct target is not
rewritten as a side effect of an incoming relationship. A Memory selflink
advances that one Bead as its owning source, with no second target write.

Issue v2 continues to own only blocking Dependencies. Whether Issues should
own other informational Types remains open. The private
`types/preview-related-v2` name and its `note` property are deliberately
experimental; neither is a settled production Type contract. Metadata
placement, nominal Issue Types, aliases and deletion/restore policy remain
open review questions.

Schema/marker version 4 and `types/preview-memory-v2` are disposable. Earlier
preview workspaces refuse rather than being migrated. Shared-server init adds
`--server --external --server-host 127.0.0.1 --server-port PORT`; no manual
schema seed or private engine build is required.

## Qualification and remaining work

Run `scripts/graph-c0-smoke.py --bd /absolute/path/to/bd --output-dir NEW_DIR
--mixed-links` for the installed-process transcript, adding `--server-port PORT
--server-root DIR` for ordinary shared-server Dolt. The expanded transcript
also runs the preceding C0 and Dependency workflows. It preserves exact
commands, exits, JSON results and version transitions. Separate real-engine
storage tests cover rollback and competing writers.

This slice does not complete W3 or M1. ID unlink, ambiguous-pair refusal,
incident listing, target-version pins, request replay, common metadata,
Memory updates, public historical reads/traversal, migration/backup continuity,
full Issue compatibility and BDP serving remain work. Retaining snapshots is
necessary for History but does not deliver its public interface. The plan
places BDP Read after the mixed CLI milestone; it does not wait for every
History or migration feature.
