---
title: "Revision-aware CLI references"
description: "Graph preview input selectors, exact retained reads, and safe refusal of unsupported endpoint pins."
---

Use a saved revision to read the exact state that informed a decision. Revision
selectors identify content; write guards independently protect the current
state against stale edits.

This speclet defines the Preview 3 graph CLI adapter for `bd show`, `bd link
show`, Link creation, and the mutation targets listed below. It requires the
[BDP Core Reference model](https://github.com/gastownhall/bdp/blob/b12665d64c2aaf6172bd1cdab409cef31ef28f44/specs/profiles/core/core.md#references)
for the distinction between floating and pinned References. It introduces no
BDP operation, JSON encoding, or implicit Types or History profile dependency.
The available retained reads are the existing preview store contract; they do
not advertise BDP HTTP History support. This candidate must be qualified and
integrated before it becomes a release capability.

## Input syntax

| Input | Meaning in the selected reference positions |
| --- | --- |
| `ID`, `beads/PATH`, or `links/PATH` | Identity without an explicit revision; resource kind remains command-driven. |
| `ID@TOKEN` | Local identity plus an exact, opaque retained revision. The first raw `@` separates the identity and token. |
| `ID@current` | Select current state on a read. On endpoint creation this requests an exact pin resolved at admission, which this candidate refuses. |
| `ID@unpinned` | Explicitly select current state on a read or a floating endpoint at creation. |
| Unwrapped locator containing `://` | Literal URL candidate, including any `@`. Existing canonical Scope URL and authority checks still apply. |
| `'[URL]@TOKEN'` | Canonical HTTP(S) Scope URL plus exact revision. Brackets wrap the entire URL; nested IPv6 host brackets are allowed. |

An existing local ID containing raw `@` must be addressed by its complete
canonical Scope URI in these positions. For example,
`https://example.org/team/beads/plan@draft` selects that literal identity;
`plan@draft` selects revision `draft` of `plan`. Do not resolve this ambiguity
by looking up both possibilities. The existing CLI URL boundary is lexical:
unwrapped inputs containing `://` remain literal URL candidates; colon-only
spellings use local syntax. Thus `plan:v1@r7` selects revision `r7` of
`plan:v1`, and `urn:x@y` selects revision `y` of the local ID `urn:x`. A
literal ID `plan:v1@r7` is addressed by
`https://example.org/team/beads/plan:v1@r7`. No stored identity changes.

A revision is nonempty UTF-8 of at most 4,096 bytes; its bytes convey no time
or ordering. `current` and `unpinned` are reserved suffix words. On exact reads,
use `--version current` or `--version unpinned` to select a literal token equal
to either word. There is no additional suffix escape in this candidate.
Malformed brackets, empty selectors/tokens, invalid UTF-8, and a suffix
combined with its alternative flag must refuse with `invalid_selector`
(exit 2). URI brackets outside the host-literal syntax must be percent-encoded.
Bracketed operands must be canonical HTTP(S) Scope URL candidates with an
authority. Opaque URI spellings such as `[urn:x@y]@r7`, no-authority URLs,
and malformed URLs refuse; they must never lower to local IDs. Scope
membership remains a separate admission check. Quote bracketed operands to
protect them from shell expansion.

## Selected command behavior

| Command | Required behavior |
| --- | --- |
| `bd show RESOURCE`, `bd link show LINK` | Bare identity reads current. An exact suffix reads that retained state, even after a later update or deletion. `@current` and `@unpinned` read current. `--version TOKEN` remains an exact-token alternative. Refuse a suffix together with `--version`, even when their values match. |
| `bd link SOURCE TARGET`, `bd link create SOURCE TARGET` | Resolve each endpoint selector separately. Bare and `@unpinned` endpoints remain floating. `--source-revision TOKEN` and `--target-revision TOKEN` are alternatives to suffixes, with the same reserved words. A suffix and its flag on the same endpoint conflict. |
| Endpoint exact pins or `@current` | Refuse with `capability_unavailable` (exit 5), before mutation. Do not silently drop the pin, resolve it in a separate client read, or imply that path-only storage preserves it. |
| `bd update RESOURCE`, `bd link update LINK`, `bd delete BEAD`, `bd forget BEAD`, `bd link delete LINK` | Require bare identities or literal canonical URIs. Refuse every revision suffix, including `@current` and `@unpinned`, with `invalid_selector` (exit 2). A suffix never supplies a write guard. |

`--if-revision` remains a comparison against the current resource revision.
`--if-source-revision` independently compares the current owning source;
`--source-revision` is an endpoint selector and cannot substitute for it.
Omitting a guard retains the selected command's current-state default; no
`--unconditional` flag is introduced. Existing stale-guard refusals remain
`revision_conflict` (exit 4). An exact
retained token absent from the preview store refuses with `revision_unknown`
(exit 3); it must not fall back to current state. Refusals emit no success
stdout and do not change the source or its Links.

```sh
bd show policy@SAVED_TOKEN --json
bd show '[https://example.org/team/beads/policy]@SAVED_TOKEN' --json
bd show https://example.org/team/beads/plan@draft --json
bd link plan@unpinned target --type types/preview-related --target-revision unpinned
bd update policy --if-revision OBSERVED_TOKEN --properties '{"body":"Revised"}'
```

A locator URL outside this workspace remains literal during parsing. Existing
local-only admission still refuses it; this syntax does not authorize external
endpoint storage or dereferencing.

## Results and boundaries

The CLI must keep identity and revision separate in JSON and persisted records.
It must not store `ID@TOKEN` as an identity or invent a pinned Reference in
output. Property/metadata edits through `bd update` and `bd link update`,
existing-Memory edits through `bd remember`, and Link creation disclose the
writer-returned current revision separately from any replaced predecessor.
Owned-Link creation and covered informational Link writes additionally disclose
the writer-returned current source revision for a subsequent guarded write.
Other creation and Issue lifecycle summaries retain their existing behavior;
use `bd show` when a summary does not supply the current revision. Current human references use `'[URI]@TOKEN'` syntax; a removed Link has
no new current Link revision to advertise.

This stage does not deliver persistent endpoint pins, same-ID re-pin, Type
revision selection, external Link admission, or a new History contract. `bd
recall` retains its existing `--version` interface. Other unselected command
positions keep their existing selector rules. In particular, ordinary Issue
mode, native `bd dep add`/`bd dep remove` operands, and legacy `--deps`/`--parent`
values retain literal Issue-ID semantics; this stage does not parse revision
suffixes there. The graph `bd link` form and native `bd dep add` perform the
same underlying dependency operation for bare Issue IDs, but their selected
input syntaxes differ. `bd unlink` is not registered in this candidate; remove
a Link by its ID with `bd link delete`.

For a walkthrough of the existing graph commands, see the
[graph CLI guide](/reference/graph-cli).
