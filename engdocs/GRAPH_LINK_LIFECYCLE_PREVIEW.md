# Experimental incident Links and guarded unlink

This slice starts from qualified [PR #22](https://github.com/donnabox/beads/pull/22)
at `8b742c511bffff8813732d2fde56c5a487f5d115`, continuing the
[fork delivery plan](https://github.com/donnabox/beads/pull/18) and
[CLI proposal](https://github.com/gastownhall/beads/issues/6703).
The original September 23, 2026 07:01 PDT attempt clock continues.

## Commands

Create a disposable workspace with the [mixed-Link demo](GRAPH_MIXED_LINK_PREVIEW.md),
then use this branch's installed binary:

```sh
bd links beads/plan --json
bd links beads/plan --direction out --json
bd links beads/rationale --direction in \
  --resource-type https://example.invalid/mixed/types/preview-related-v2 --json
bd unlink links/plan-rationale --unconditional --unconditional-source --json
bd links beads/plan --json
bd show beads/plan --json
bd show links/plan-rationale --json  # gone; the identity remains reserved
```

For edits based on observed state, use `--if-revision REV` and
`--if-source-revision SOURCE_REV` instead of the explicit unconditional choices.
The Link guard is mandatory; a Memory-owned Link also requires its source guard.
Issue informational Links remain provisionally unowned. Optional Issue-source
guards are checked when supplied. Stale guards refuse the entire operation.

Unlink returns a preview tombstone plus the affected source record. It removes
only the selected live informational Link, retains its earlier snapshots and
permanently reserves its canonical ID. A Memory source receives one new version
with the complete remaining owned set, while a distinct target stays unchanged.
A selflink advances the one Bead as its source, without a second target write.
Repeated unlink reports `gone`; it is not request replay or an invented success.
A later create using the deleted ID refuses with `identity_reserved`.

Pair selection is available when the identity is not already known:

```sh
bd unlink beads/plan beads/rationale \
  --resource-type https://example.invalid/mixed/types/preview-related-v2 \
  --unconditional --unconditional-source --json
```

Zero matching live Links reports `not_found`. Multiple matches report
`ambiguous_link` with sorted canonical `candidateIDs` and change nothing. Select
one candidate by ID; there is no implicit remove-all. Selection, guards, Link
removal and source retention share one authority transaction, not a CLI
list-then-delete loop. Conditional pair unlink checks the selected Link's revision.

## Incident listing

`links` reads both authoritative specialized blocking Dependencies and generic
informational Links in one read transaction. Direction is relative to the
selected Bead (`both` by default); a selflink appears once. Results contain
complete current Link records in canonical ID order. Deleted Links are excluded.
An optional exact installed Type URL filters the result. Missing or inconsistent
authority mappings refuse explicitly instead of disappearing behind joins or
Type filters. This disposable store assumes its controlled graph writer; a
damaged mapping anywhere in that store can block collection inspection. Endpoint content is not
expanded and no remote endpoint is fetched.

This disposable implementation has no pagination: it returns the complete
matching set up to the disclosed 1,000-Link budget or refuses without partial
results. It does not silently honor `--limit`, `--after`, `--all`, `--version`,
or `--resolve-targets`; those remain unsupported. This bounded interface is
not the final collection or historical-incident contract.

## Storage and limits

Schema/marker 5 is disposable; previous preview workspaces refuse. The private
`preview-related-v2` descriptor corrects an inherited endpoint declaration:
`conformsTo` is an intersection, not an either/or list. It requires no nominal
endpoint Type and prohibits external endpoints. The controlled preview writer
admits its two installed Bead Types (Issue and Memory). Future custom Type
admission must revisit this restriction explicitly. Existing
Issue and Dependency tables retain sole authority for their current payloads.
Informational Link deletion changes its catalog entry to deleted and records a
new retained tombstone while removing its current generic payload. Complete
source-owned history is retained in the same ordinary Dolt transaction.
The prior Link versions remain immutable; no historical query joins today's
owned membership in place of the retained version.

Blocking Dependency **listing** works. Blocking Dependency **unlink** refuses
until its domain removal and source-Issue history adapter is connected. This
preview must never erase a Dependency through the informational writer.

Bead deletion/cascade, restoration and erasure, request replay, common metadata,
target pins, public History, BDP serving and existing-workspace adoption remain
unfinished. The private tombstone and envelope are explicitly experimental,
not BDP wire. Issue informational ownership and other disputed contracts remain
open. This increment does not by itself establish full W3/M1 delivery.

## Verification

`scripts/graph-c0-smoke.py --bd /absolute/path/to/bd --output-dir NEW_DIR
--link-lifecycle` runs normal initialization and the previous C0, Dependency and
mixed-Link transcripts before these commands. Add `--server-port PORT
--server-root DIR` for ordinary shared-server Dolt. Every command is a fresh
process; raw outputs and version transitions are preserved. Separate storage
tests prove rollback, retention and competing-write behavior.
