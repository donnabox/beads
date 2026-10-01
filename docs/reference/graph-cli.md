# Graph CLI guide

This page describes the experimental CLI in a workspace initialized with
`bd init --graph-mode link`. It is a standalone command guide; the
[graph preview reference](/reference/graph-preview) has the complete capability bounds
and storage rules. Ordinary Beads workspaces keep their existing Issue and
key/value-memory commands.

In a graph workspace, a Bead is an Issue or a Memory. Its canonical identity
is under `beads/`. In CLI arguments, `policy` means `beads/policy`; a Link
still needs an explicit `links/ID` where a command accepts either kind of
resource. The shorthand does not change stored IDs or HTTP URLs. Omit `--id`
when creating a Bead to allocate an ID, or supply one to use it exactly;
creating a second Bead at that ID fails. The examples below use explicit IDs
only so later commands are easy to follow.

## Create Beads

`bd remember` creates a Memory from text. It preserves the full body. Without
`--title`, it makes a title from the first nonempty body line (collapsed
whitespace, at most 80 Unicode characters including an ellipsis). An explicit
`--title` is used as supplied. It can also read a body from `--body-file PATH`
or `--stdin` instead of the positional text.

```sh
bd remember 'Code flow policy: changes land on integration.' --id policy
bd remember 'Keep review branches until their changes land.' --title 'Review branch policy'
```

`bd create` creates an Issue by default. Use `--bead-type` to select an
installed Bead Type; the currently supported graph writers are Issue and
Memory. `--type task` or `--type bug` is an **Issue classification**, not a
Bead Type. Graph Memory creation through `bd create` accepts an optional
positional title or `--title` and an inline `--body` (also spelled
`--description` or `--message`). It derives a title from the body if none is
given. This form does not read a body from a file or stdin.

```sh
bd create 'Move the release branch' --id work --type task
bd create --bead-type types/preview-memory-v2 \
  --body 'Code flow policy: keep the old policy as a versioned Memory.'
```

Use `bd types` in the selected workspace to discover its installed
`types/NAME` IDs. Do not assume a Type shown in another workspace is installed
in this one.

## Find and read Beads

| Command | Graph workspace behavior |
| --- | --- |
| `bd memories [SEARCH]` | List current Memory title/body summaries. Use `--all` for all matches within the preview's bounds, `--details` for saved version and attribution, or `--format records-json` for structured summaries. |
| `bd recall ID` | Print **one** Memory's exact body bytes. It does not enumerate Memories or add a newline. |
| `bd show ID --json` | Read one current Issue or Memory record; use `links/ID` for a Link. |
| `bd list --flat` or `bd list --format records-json` | List **Issues only** today. Existing status, Issue classification, assignee and due-date filters apply. This is not yet an all-Bead inventory. |

```sh
bd memories --all
bd memories 'code flow' --details
bd recall policy
bd show work --json
bd list --flat --all
```

There is not yet one CLI command that enumerates every Bead regardless of
Type. The BDP HTTP `beads/` collection already does so in an ordinary
shared-server graph workspace, with pagination. Follow every response's
`next` URL until it is `null`, or use the
[public Python read example](https://github.com/versioned-beads/beads/blob/integration/examples/bdp-read/read_beads.py), which
follows those pages. Graph `bd list` becoming an all-Bead inventory is a
proposed CLI change, not current behavior.

## Update and delete Beads

For a Memory, `bd remember --update ID` changes only the fields you supply.
It requires `--update` so an existing Memory is never silently overwritten by
a creation command. Omitted title or body stays unchanged; `--update` accepts
the current revision by default. A title-only update needs no body argument.

```sh
bd remember 'Changes now land on the release branch.' --update policy
bd remember --update policy --title 'Current code flow policy'
```

`bd update` edits an Issue with its Issue flags, or replaces a Memory's whole
properties document with `--properties`. Unlike `bd remember --update`, these
routes require an explicit write choice; the examples use `--unconditional`
to accept the current state. A Memory properties replacement supplies both
`title` and `body` strings. See [Versioning and History](#versioning-and-history)
when you need stale-write protection.

```sh
bd update work --title 'Move the release branch after review' --unconditional
bd update policy --properties '{"title":"Code flow policy","body":"Land reviewed changes on integration."}' --unconditional
```

Memory deletion applies only to an **unreferenced** Memory. `bd delete ID`
previews the result without changing storage; `--force` applies it. `bd forget
ID` applies the same deletion directly. Applying either command requires an
explicit write choice, shown here with `--unconditional`. Any live incoming,
outgoing or self-Link makes deletion refuse; `--force` does not cascade.
Issue deletion is not available in this graph preview.

```sh
bd delete policy
bd delete policy --force --unconditional
# Or, for another unreferenced Memory:
bd forget scratch --unconditional
```

Deletion removes current Memory state but reserves its ID and retains prior
snapshots. It does not create a deletion version or promise erasure or restore.

## Versioning and History

Current graph records carry opaque revision/version tokens. Save a token from
a record or `bd memories --details` to read that exact retained state later.
Version reads do not depend on the record still being current. `bd compare`
compares two **chosen** retained versions; the token order you give it sets
the comparison direction. It does not establish chronological order.

```sh
bd memories 'code flow' --details
bd recall policy --version SAVED_TOKEN
bd show policy --version SAVED_TOKEN --json
bd compare policy --from FIRST_TOKEN --to SECOND_TOKEN --json
```

| Flag | Current use |
| --- | --- |
| `--version TOKEN` | Select one exact retained state for `bd show`, or one retained Memory body for `bd recall`. |
| `--from TOKEN --to TOKEN` | Select the two complete states for `bd compare`. |
| `--if-revision TOKEN` | On a supported write, refuse if the current record no longer has that revision. |
| `--unconditional` | Where a write requires an explicit choice, accept the current record without an expected revision. |
| `--if-source-revision TOKEN` | On a Memory-owned Link write, optionally require the source Memory's observed revision; otherwise that source defaults to unconditional acceptance. |

`bd remember --update` defaults to unconditional acceptance. Memory deletion,
`bd update`, and Link edits/removal still have their command-specific guard
requirements; consult the [graph preview reference](/reference/graph-preview) before
automating them. A semantic no-op retains the existing revision.

**An ordered graph History command is not available in this checkpoint.**
The ordinary Issue `bd history` and `bd versions` commands must not be taken
as a graph Memory/Link History API. Exact retained reads and comparisons are
available locally, but they do not provide a complete timeline, timestamps,
deletion events, restoration, or BDP HTTP History. `bd status --graph`
advertises the current capability set; `historyExact` is false here. This
section should change only after a History implementation is proved and
landed in the integration branch.

## Discover installed Types

`bd types` reads the descriptors **installed in this workspace** and lists
Bead Types separately from Link Types. Use the printed `types/NAME` ID with
`--bead-type` or `--link-type`; a full local Type URL also works.
`--details` shows each complete stored descriptor, and `--json` returns the
descriptors as structured data. An older workspace may have fewer Types than
a new one; reading it does not install new example Types.

```sh
bd types
bd types --details
bd types --json
```

The supported graph `bd create` Bead Types are the installed Issue and Memory
descriptors. `bd link` uses installed Link Types. The ordinary `bd types`
command lists Issue classifications instead; select a graph workspace to see
the graph Type catalog.
