# Preview 3 human Resource output

In graph-mode link, successful Bead and Link creation and ordinary collection
rows use the same order: `ID [Type]`, followed by native Issue status/priority
and title, or a Link's `from → to` endpoints. The five Memory relationship
creation commands use the generic Link row. They print JSON only when requested.

Human output removes only this workspace's exact Scope URL prefix. Literal
`beads/`, `links/`, and `types/` ID segments remain intact. Foreign references
remain absolute. Native Issue and Dependency Types use their authoritative
classification labels (for example `task` and `blocks`), not abstract base Types.

`bd show` and `bd link show` use the same document layout for current and exact
retained records. The selected revision, stored body/properties, metadata and
Link endpoint pins come from that record. A historical document never fills its
gaps with current comments or relationships. Empty sections are omitted.
`--long` includes the complete Type identity, exact revision token and recorded
administrative details. Optional ordinal presentation can fall back to the full
revision token when its separate bounded History lookup is unavailable; it cannot
turn a successful content read into failure.

Pinned endpoints display `ID@=TOKEN`; floating endpoints display just their ID.
The human versions timeline is headed `SEQ`, includes actual terminal deletion
events after recovery and never renumbers them. A deletion event is not a readable
live snapshot. `ID@N` is a store-local convenience; exact tokens remain the
portable references and all structured Resource identities/tokens stay complete.
Use `versions --long` (or `history --long` / `link versions --long`) to include
the exact token lines in the human timeline.

These human projections do not change legacy JSON, canonical projection/masking,
BDP records, mutation postimages, selection or filtering behavior.
