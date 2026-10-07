# Graph CLI guide

> **New workspaces only.** The graph preview cannot be enabled on an existing
> `.beads` project or used to migrate its Issue database. Start in a new
> directory without `.beads`; `bd init --graph-mode link` refuses an existing
> or incomplete workspace rather than replacing its data. Keep ordinary
> projects on their existing format. Do not copy a `.beads` directory into a
> graph workspace as an upgrade path.

Graph workspaces are not migrated between graph schema versions. If a
workspace's `graph_schema_version` in `.beads/metadata.json` is not this
build's `6`, `bd` refuses it with `graph_not_initialized` and does not open
its database. Preserve its data and create a fresh workspace. A schema-6
workspace created before the two example Link Types existed opens and works
normally without them; see [Discover installed Types](#discover-installed-types).

Build `bd` from the [integration branch](https://github.com/versioned-beads/beads/tree/integration),
not a released binary, and initialize a new project explicitly:

```sh
mkdir graph-demo && cd graph-demo
git init
bd init --graph-mode link --scope-url https://example.org/team/ \
  --skip-hooks --skip-agents --non-interactive
bd status --graph
```

The Scope URL establishes canonical identity; this command does not start an
HTTP server at that address. The example skips agent-file and hook setup to
keep the CLI exercise isolated. See the
[Graph CLI Specification (Draft)](/reference/graph-cli-specification-draft) for supported
embedded and external Dolt modes, agent setup, limits, and refusals.

This page is the evolving **task-oriented CLI guide**. The graph preview
technical reference is the evolving, detailed command matrix and contract;
neither page is a frozen release note. The blog post explains the model and
links to these pages for commands that may change after publication. Ordinary
Beads workspaces keep their existing Issue and key/value-memory commands.
The technical reference also carries the [proposed next-milestone command contract](/reference/graph-cli-specification-draft#proposed-command-contract-for-the-next-milestone);
its **NYI** rows describe proposed behavior, not commands available in this
build. Use the current command matrix there when trying the preview.

In a graph workspace, a Bead is an Issue or a Memory. Its canonical identity
is under `beads/`. In CLI arguments, `policy` means `beads/policy`; a Link
still needs an explicit `links/ID` where a command accepts either kind of
resource. The shorthand does not change stored IDs or HTTP URLs. Omit `--id`
when creating a Bead to allocate an ID, or supply one to use it exactly.
`bd create` refuses a duplicate ID; `bd remember --id ID` updates an existing
Memory at that ID unless `--create-only` is supplied. The examples below use
explicit IDs only so later commands are easy to follow.

## Create Beads

`bd remember` stores a Memory from text. It preserves the full body. Without
`--title`, it makes a title from the first nonempty body line (collapsed
whitespace, at most 80 Unicode characters including an ellipsis) on creation.
An explicit `--title` is used as supplied. It can also read a body from
`--body-file PATH` or `--stdin` instead of the positional text. With `--id`,
an unused ID creates and an existing Memory updates in place, matching
ordinary `bd remember --key` behavior. `--create-only` with `--id` refuses
any previously allocated ID, including one that has been deleted.

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
| `bd list` or `bd list --format records-json` | Without an Issue filter, list every current Memory and Issue the ordinary `bd list` would show, newest recorded change first; closed and pinned Issues need `--all`. Use `--bead-type types/NAME` to narrow by nominal Type. An Issue filter, or a matching directory label, switches to the Issue-only query described below and says so. |

```sh
bd memories --all
bd memories 'code flow' --details
bd recall policy
bd show work --json
bd list --all
bd list --format records-json --bead-type types/preview-memory-v2 --all
```

Without an Issue filter, `bd list` reads one bounded snapshot of every
current Bead. Each human row shows the Bead's local ID (such as
`beads/work`) and its kind (`Memory` or `Issue`); an Issue row also shows its
status and priority, then the title. Closed and pinned Issues are hidden
unless you pass `--all`, which also removes the row limit. `--limit` sets a
visible prefix and `hasMore` indicates that the prefix omitted matches.
It is not a continuation cursor. A positive `BEADS_MAX_ROWS` refuses a page of
more Beads than the cap. The BDP HTTP `beads/` collection provides
pagination in an ordinary shared-server graph workspace. Follow every
response's `next` URL until it is `null`, or use the
[public Python read example](https://github.com/versioned-beads/beads/blob/integration/examples/bdp-read/read_beads.py), which
follows those pages.

The Issue filters are `--status` (or `--state`), `--type`, `--title`,
`--title-contains`, `--priority`, `--priority-min`, `--priority-max`,
`--assignee`, `--no-assignee`, `--label`, `--label-any`, `--exclude-label`,
`--pinned`, `--no-pinned`, `--due-before`, `--due-after`, `--overdue`,
`--sort` and `--reverse`. Supplying any of them switches `bd list` to the
existing Issue query, even when an empty value such as `--assignee=` adds no
restriction. A configured `directory.labels` entry that matches the current
directory switches it too. That query lists Issues only, never Memories. It
omits closed and pinned Issues unless `--all` or a filter selects them. A line
under the header says Memories are not listed and names the option that
selected the query. A directory label alone never refuses a `--bead-type`; it
is not applied to a Memory. Its human rows add status and priority, for example
`"https://example.org/team/beads/work" "open" P2 "Move the release branch"`.
An Issue filter cannot be combined with a non-Issue `--bead-type`.

## Update and delete Beads

For a Memory, `bd remember --id ID` changes only the fields you supply if
that ID already exists. Omitted title or body stays unchanged, and the current
revision is accepted by default. `--update ID` remains an existing-only
spelling; it refuses a missing Memory. A title-only update needs no body
argument. Add `--if-revision TOKEN` to either spelling to reject a stale edit.

```sh
bd remember 'Changes now land on the release branch.' --id policy
bd remember --update policy --title 'Current code flow policy'
```

`bd update` edits an Issue with its Issue flags, or replaces a Memory's whole
properties document with `--properties`. Unlike `bd remember` with an existing
ID, these
routes require an explicit write choice; the examples use `--unconditional`
to accept the current state. A Memory properties replacement supplies both
`title` and `body` strings. See [Versioning and History](#versioning-and-history)
when you need stale-write protection.

```sh
bd update work --title 'Move the release branch after review' --unconditional
bd update policy --properties '{"title":"Code flow policy","body":"Land reviewed changes on integration."}' --unconditional
```

`bd close ID...` and `bd reopen ID...` accept one or more local Issues. A
single `--reason` applies to every close target, repeated reasons map to IDs
in input order, and `--reason-file PATH` preserves the file's literal text.
The `bd done ID MESSAGE` alias accepts a trailing reason. A batch keeps
successful changes when another target fails: successful records go to stdout,
per-target failures go to stderr, and the command exits nonzero. Reopen
reports already-open Issues without adding a version. `--force` overrides the
native pinned, holder, blocker and open-child close policies; it does not make
an already-closed Issue change again. `--session ID` records the closing
session, with `CLAUDE_SESSION_ID` as its fallback. Interactive last-touched,
gate evaluation, post-close workflow flags and remote routing remain pending.

```sh
bd close work review --reason 'Finished the work' --reason 'Review complete'
bd close blocked-work --force --session review-session --reason 'Override approved'
bd reopen work review --reason 'Follow-up needed'
```

`bd delete ID` previews deletion of one unreferenced Memory or Issue without
changing storage; `--force` applies it. `bd forget ID` applies Memory deletion
directly and does not accept Issues. Applying either command requires an
explicit write choice, shown here with `--unconditional`. Any live incoming,
outgoing or self-Link makes deletion refuse; `--force` does not cascade.
Memory refusal names the first blocking Link ID and the total count. Issue
refusal lists the incident Link IDs; unlink them explicitly before retrying
with a fresh revision.

```sh
bd remember 'Temporary note' --id scratch
bd delete scratch
bd delete scratch --force --unconditional
# Or, for another unreferenced Memory:
bd remember 'Another temporary note' --id other-scratch
bd forget other-scratch --unconditional
# An unreferenced Issue follows the same preview and guarded apply shape:
bd create 'Temporary task' --id temp-work
bd delete temp-work
bd delete temp-work --force --unconditional
```

Deletion removes current Bead state but reserves its ID and retains prior
snapshots. It does not create a graph deletion version or promise erasure or
restore. Native Issue deletion records its ordinary delete journal event.

Use `bd defer ID...` to set Issues aside and `bd undefer ID...` to return
deferred Issues to open. An undated defer stays in the icebox until undeferred.
`--until` accepts the same date and relative-time forms as ordinary `bd`;
the next `bd ready` after that time wakes the Issue, records a new version,
and clears its defer date. `--reason` appends a line to the Issue's notes.
Assigned Issues retain their assignee through defer and undefer. Neither
command requires a revision flag, but `--if-revision` checks the observed
revision for a single Issue. A repeated dateless defer or undefer is a no-op
when no date or reason changes. `bd undefer` also clears a stale defer date
without changing a non-deferred status.

```sh
bd defer work --until tomorrow --reason 'Waiting on review'
bd ready                       # Wakes work once its defer date has passed
bd undefer work                 # Or restore it explicitly
bd defer work another-work      # Set multiple Issues aside indefinitely
bd show work --json
bd undefer work --if-revision REVISION_FROM_SHOW
```

The graph preview can narrow `bd ready` by priority, Issue classification,
assignee, label, or defer state and can sort or limit the result. `bd blocked`
accepts the ordinary label, label-any, and exclude-label filters. Both use
the native blocker-aware predicate, then return canonical graph Issue IDs.
An explicitly supplied blank label filter is an error, so a typo cannot turn
an intended narrow query into the whole work queue.

```sh
bd ready --priority 2 --type task --limit 10 --json
bd ready --label release --sort priority
bd ready --claim --label release --actor rig.agent --json
bd blocked --label release --json
```

`ready --claim --json` returns a one-element array of complete graph Issue
records, or an empty array when no matching ready Issue exists. An empty claim
does not create a claim lease or revision. As with ordinary `ready`, the lazy
defer-wake before selection may separately version an expired deferred Issue.
The remaining ordinary ready modes and blocked `--parent` remain in
the [draft CLI specification](/reference/graph-cli-specification-draft)
and refuse in this graph slice until their supporting graph behavior is ready.

## Edit Issue notes

An Issue's notes can carry progress across sessions. `--append-notes` adds a
line without discarding earlier notes. `--notes` sets a nonempty value; if it
would replace different, nonempty notes, add `--force` to make that intent
explicit. To erase notes, use `--clear-notes`: `--notes=` is refused, even with
`--force`, because an accidentally empty shell value could erase the record.

```sh
bd show work --json                     # Read the current revision.
bd update work --append-notes 'Review started.' --if-revision REVISION_FROM_SHOW
bd update work --notes 'Revised handoff' --force --if-revision NEW_REVISION
bd update work --clear-notes --if-revision LATEST_REVISION
```

Each graph Issue edit needs `--if-revision` or `--unconditional`. A stale guard
refuses even with `--force`; force only authorizes the notes overwrite. An
identical value or a clear of already-empty notes leaves the revision alone.
Use `bd versions work` and `bd show work --version TOKEN` to inspect prior
snapshots after an accepted change.

## Create, inspect, edit and remove Links

Use `bd types` to find installed Link Types. An informational Type such as
`types/preview-related-v2` can connect a Memory to a Memory or Issue. An
explicit `--id ID` is shorthand for `--id links/ID`; omit it to allocate an ID.
The Link's Type and endpoints do not change during a properties edit. For the
installed preview informational Types, the optional property is a string
`note`. A Memory owns its outgoing informational Links, so changing one also
changes the source Memory's version. A target does not change merely because
it is linked.

```sh
bd link policy work --link-type types/preview-related-v2 \
  --id policy-work --properties '{"note":"work follows this policy"}'
bd links policy
bd show links/policy-work --json
bd update links/policy-work --properties '{"note":"reviewed policy"}' --unconditional
bd unlink policy-work --unconditional
```

`bd links ID` lists current incident Links. A Memory-owned informational Link
defaults to accepting the current source state; add `--if-source-revision
TOKEN` to reject a stale source. Link updates and unlink still require their
own `--if-revision TOKEN` or `--unconditional` choice. Unlink removes the
current Link but retains its identity and prior snapshots. The blocking
`types/preview-blocks-v1` Type is only for live Issues; it refuses a Memory
endpoint. See the [Graph CLI Specification (Draft)](/reference/graph-cli-specification-draft) for the
separate blocking Dependency unlink rules and Link Type bounds.

## Claim and release Issue work

`bd update ID --claim` atomically claims a live Issue for the current actor.
`bd unclaim ID...` releases assigned open or in-progress Issues. By default
the current actor must hold each claim. Each successful release clears its
assignee, lease, and started time, returns it to open, and retains one native
Issue version and one graph version. A second unclaim refuses because no claim
remains; it does not create another version. A batch attempts each ID and exits
nonzero if any release fails.

```sh
bd update work --claim --actor rig.agent
bd unclaim work --actor rig.agent --reason 'Handing this back'
bd comments work              # The reason is a native Issue comment
bd update work --claim --actor another.agent
bd unclaim work --if-assignee another.agent --actor supervisor
```

`--if-assignee HOLDER` releases only while that holder remains assigned; a
mismatch leaves the Issue and its lease untouched. `--force` bypasses the
holder check for an abandoned claim, but still uses the native row
compare-and-swap. The two flags cannot be combined. `--reason TEXT` appends a
native Issue comment after the release; if that separate append fails, the
release remains committed and the CLI warns. Comments are a separate feed,
outside retained Issue state, and do not mint another Issue version. Other
comment writes are not yet exposed in graph mode.

## Versioning and History

**Revision** names the current state of one Bead or Link. A complete graph
record has one opaque `revision` field. Use it with `--if-revision` when a
write must apply only to the state you read. A different current revision
means the state changed and the guarded write refuses. **Version** names a
retained state of that Resource. `bd versions` rows and discovery summaries
call its exact-read address `version`; use that token with `--version`, or as
an operand to `bd compare`, to retrieve or compare the state later. A current
record's `revision` is also its retained-state address. These names express
two uses of one opaque token, not separate counters. The token encodes no
time or order and is not a Dolt commit ID.

Save a token from a record or `bd memories --details` to read that exact
retained state later. Version reads do not depend on the record still being
current. `bd compare` compares two **chosen** retained versions; the token
order you give it sets the comparison direction. Compare does not discover
their chronological order.

```sh
bd memories 'code flow' --details
bd recall policy --version SAVED_TOKEN
bd show policy --version SAVED_TOKEN --json
bd compare policy --from FIRST_TOKEN --to SECOND_TOKEN --json
bd versions policy
bd history policy  # same listing in a graph workspace
```

| Flag | Current use |
| --- | --- |
| `--version TOKEN` | Select one exact retained state for `bd show`, or one retained Memory body for `bd recall`; use a saved `version` token. |
| `--from TOKEN --to TOKEN` | Select the two complete states for `bd compare`. |
| `--if-revision TOKEN` | On a supported write, refuse if the current record no longer has the saved `revision`. In an ordinary workspace, `bd update` and `bd delete` instead read it as the legacy compare-and-swap on a decimal bead `revision`. |
| `--unconditional` | Where a write requires an explicit choice, accept the current record without an expected revision. |
| `--if-source-revision TOKEN` | On a Memory-owned Link write, optionally require the source Memory's observed revision; otherwise that source defaults to unconditional acceptance. |

`bd remember` with an existing `--id` or `--update` defaults to unconditional
acceptance. Memory deletion,
`bd update`, and Link edits/removal still have their command-specific guard
requirements; consult the [Graph CLI Specification (Draft)](/reference/graph-cli-specification-draft) before
automating them. A semantic no-op retains the existing revision.

`bd versions ID` lists a Memory, Issue or Link's versions newest first in a
graph workspace. `bd history ID` is an alias there; in an ordinary workspace,
`bd history` retains its Dolt-commit meaning. Each graph row includes an
opaque `version` token, a store-local `local_revision`, a display `change_at`
time, and attribution. Use the **token** for `show --version` or `compare`, never
the local revision number. Human output calls the number `REV`, matching
ordinary `bd versions`; graph JSON uses `local_revision`. It orders versions
within this store and is not a stable cross-clone address. `change_at` is not
the ordering authority. A removed Link's deletion marker is listed with
`removed: true` but is not a readable Link version. Memory deletion adds no
deletion version. There is still no
BDP HTTP History, as-of selection or restoration. `bd status --graph` reports
`versionList: true`; `historyExact: false` refers to the unavailable HTTP
History profile. The [technical reference](/reference/graph-cli-specification-draft#list-a-resources-versions)
details ordering and refusal behavior.

## Discover installed Types

`bd types` reads the descriptors **installed in this workspace** and lists
Bead Types separately from Link Types. Use the printed `types/NAME` ID with
`--bead-type` or `--link-type`; a full local Type URL also works.
`--details` shows each complete stored descriptor, and `--json` returns the
descriptors as structured data.

`bd init` installs six Types: the Issue and Memory Bead Types and four Link
Types (`types/preview-blocks-v1`, `types/preview-related-v2`,
`types/example-follows` and `types/example-cites`). A schema-6 workspace
created before the two example Link Types existed has only the other four.
It opens and works normally, and `bd types` lists those four. No command
installs the examples into it; selecting one there refuses with
`capability_unavailable`.

```sh
bd types
bd types --details
bd types --json
```

The supported graph `bd create` Bead Types are the installed Issue and Memory
descriptors. `bd link` uses installed Link Types. The ordinary `bd types`
command lists Issue classifications instead; select a graph workspace to see
the graph Type catalog.
