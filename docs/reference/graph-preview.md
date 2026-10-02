# Mixed Memory and Issue graph preview

For setup and task-oriented commands, start with the
[graph CLI guide](/reference/graph-cli). This page is the living technical
reference for the preview's exact command matrix, contracts, bounds and
unsupported operations. The blog post is the publication narrative; this
reference and the CLI guide continue to change with the implementation.

This integration checkpoint supports a bounded Memory/Issue workflow in a
**fresh, explicitly selected graph workspace**. Existing ordinary Issue
workspaces continue using their existing commands and storage. The generic
model, Type names and result shapes remain an experimental preview.

Start in a new directory with no `.beads` directory:

```sh
git init
bd init --graph-mode link --scope-url https://example.org/team/ \
  --non-interactive
bd remember 'The release uses the integration branch.' \
  --id plan --title 'Release plan' --json
bd show plan --json
# A separate invocation reopens the same stored Memory.
bd show https://example.org/team/beads/plan --json
bd status --graph --json
```

The Scope URL establishes local identity; initialization does not publish a
web server at that address. Every Bead is canonically under `beads/`; every
Link is under `links/`. CLI `plan` is shorthand for `beads/plan`, including in
commands that also accept Links (`show`, `compare`, and `update`). Select a Link
there with explicit `links/PATH`. Local canonical paths and their exact Scope
URLs remain accepted; aliases and foreign Scope URLs are unavailable. This
shorthand only changes CLI input, never stored identity or output.
Creation accepts an optional bare `--id ID` or canonical `--id beads/PATH`;
omitting it generates a random canonical ID. An explicit ID is used at its
canonical Bead path and duplicates fail;
an allocated identity cannot be reused for a different record.

For an ordinary external Dolt SQL server, add these options to `bd init`:

```sh
--server --external --server-host 127.0.0.1 --server-port 3306 --server-user root
```

Use the normal connection configuration appropriate to your server. Embedded
and server modes use the existing driver and standard schema initialization;
there is no manual schema seeding step. Graph initialization records and checks
workspace identity and refuses existing workspaces, mismatched bindings and
incomplete initialization. This preview does not adopt an existing Issue
database. Different-database provisioning on Dolt 2.1.8 must be serialized.

## Instructions for agents

Omit `--skip-agents` when initializing a graph workspace to install a managed
graph instruction block in `AGENTS.md` (or the configured agents filename).
The block incorporates Stephanie Jarmak's durable-memory guidance from
[integration PR43](https://github.com/versioned-beads/beads/pull/43), with
commands supported by this graph preview:

```sh
bd remember "Use UTC for timestamps" --id beads/time-policy --title "Timestamp policy"
bd recall beads/time-policy
bd memories timestamps --format records-json
```

The block also points to `bd status --graph` for the workspace's supported
capabilities.

Omit `--id` to allocate a distinct canonical ID for each new fact, or choose an
explicit ID as above. These commands demonstrate storage and retrieval across invocations; installing instructions does not
guarantee that an agent will decide to save a fact.

Graph initialization preserves surrounding user-authored text and an existing
`CLAUDE.md` import of `@AGENTS.md`. An existing minimal managed block is replaced
with graph instructions; pass `--skip-agents` to preserve that block unchanged.
By default, initialization also registers Stephanie's existing Claude Stop
reminder in project-local `.claude/settings.json` and adds an active import in
`CLAUDE.md` when needed. `--skip-hooks` omits the Stop registration and the
`CLAUDE.md` import; `--skip-agents` omits guidance, the import and the Stop
registration. Existing Claude settings, plugins and instructions are checked
for conflicts before a graph database is created; init refuses without
rewriting conflicting files, names the file to change, and suggests
`--skip-hooks`. If the Stop hook cannot be written after the database is
created (for example, an unwritable `.claude` directory), init prints a
warning and completes the workspace without it. Use `bd setup claude` to
install the Stop hook later in a workspace initialized with `--skip-hooks` or
after such a warning:

```sh
bd setup claude
bd setup claude --check
# Remove only the project Stop registration when no longer wanted:
bd setup claude --remove
```

This graph adapter adds only `bd claude-hook stop` to project-local
`.claude/settings.json`, preserving user settings values (including large integer
values), hook siblings, existing file permissions, and all
instructions (including `ProfileGraphPreview`). A changed settings file is
reformatted as sorted, two-space-indented JSON; original key order/whitespace
is not retained. A stale graph-profile hash refuses unchanged and requires
operator reconciliation of that managed block before setup. It creates or appends an active
`@AGENTS.md` import in `CLAUDE.md` (using the configured agents filename); an
existing active import and user text remain unchanged. Unsafe paths, unclosed
fences or stale managed Beads blocks refuse before settings/instructions are
changed. Removal retains the import and all instructions. `--project` is optional;
`--check` and `--remove` are mutually exclusive. It never installs `bd prime`
SessionStart/PreCompact hooks. Existing Beads plugins or prime hooks in project,
legacy-local or global settings cause installation/check to refuse unchanged;
reconcile those configurations deliberately first. Removal touches only the
managed Stop command in project `settings.json`; it does not remove legacy-local,
global or plugin registrations.

Global/stealth setup, other recipes, custom output/template flags and setup
`--json` are unsupported. The ordinary `bd prime` remains unavailable. The
Stop command accepts its native JSON stdin/stdout protocol without `--json`,
runs without opening storage, and follows Steph's transcript and reentrancy
rules: it reminds at a session's first Stop and again at a later Stop when
tools were used since the previous one, never while Claude Code reports
`stop_hook_active`. It does not write a Memory itself. After the reminder,
`bd remember "a fact"` stores the agent's chosen content and `bd recall ID`
reads it. Integration tests prove this delivery/execution sequence, not that
agents reliably choose what to remember.

Existing full or unknown managed profiles, malformed or duplicate managed
blocks, and symlink or nonregular targets refuse before database initialization.
Keep those files and pass `--skip-agents` to initialize without changing them.
If guidance publication to the agents file fails after database
initialization, the workspace remains incomplete and fenced; this preview
provides no repair command.

The ordinary minimal profile also includes Stephanie's exact durable-memory
contribution, which changes its managed content hash. Existing refresh policy
is unchanged; this addition does not independently rewrite instruction files.

On Memory creation, omitted `--title` uses the first nonempty body line with
whitespace collapsed, capped at 80 Unicode characters including an ellipsis.
The full body is preserved. A whitespace-only body yields an empty title. An
explicit creation title must remain nonempty; updates preserve omitted fields.

## Supported graph commands

| Command | Admitted scope and flags |
|---|---|
| `types [--details]` | List the Bead and Link Type IDs installed in this workspace. `--details` prints each complete persisted descriptor; `--json` returns the descriptors as structured data. An older four-Type workspace does not claim the two example Types. Legacy `--sections` is unavailable. |
| `remember BODY [--id ID] [--title TITLE]` | Memory creation. Bare IDs resolve under `beads/`; an explicit `--body-file PATH` or `--stdin` replaces the positional body source. These sources are mutually exclusive; empty text is present content. |
| `remember --update ID` | Change only supplied `--title` and/or one explicit body source, preserving omitted fields inside the transaction. Defaults to unconditional; optional `--if-revision TOKEN` rejects stale edits. Explicit `--unconditional` remains accepted. |
| `memories [SEARCH]` | Complete bounded Memory title/body search summaries. Supports `--all`, `--details` and `--format table\|records-json`; legacy `--json` refuses. |
| `recall BEAD` | Stream one Memory's exact body bytes. Optional `--version TOKEN` selects a retained body. `--quiet` does not suppress content; `--json` refuses. |
| `update BEAD --properties JSON` | Replace a Memory's complete properties with exactly the `title` and `body` strings. Requires `--if-revision TOKEN` or `--unconditional`. |
| `update RESOURCE --patch JSON` | Apply ordered `add`, `replace`, and `remove` property operations to one Memory or informational Link. Accepts literal JSON, `@file`, or explicit `@-` stdin. Requires a Resource guard; the Memory source guard is optional. |
| `delete BEAD` | Read-only preview of deleting one unreferenced Memory. `--force` applies and requires `--if-revision TOKEN` or `--unconditional`. A preview needs no guard but checks any supplied guard. |
| `forget BEAD` | Apply the same unreferenced Memory deletion immediately, with `--if-revision TOKEN` or `--unconditional`. |
| `create TITLE [--id ID]` | Create an Issue. Bare IDs resolve under `beads/`. Allows `--title`, inline `--description`/`--body`/`--message`, `--type`, `--priority`, `--labels`/`--label`, inline `--design`, `--acceptance`, `--assignee`, `--estimate`, `--external-ref`, `--spec-id`, `--due` and initial `--notes`. Existing classification rules apply. Initial status is open. Ordinary creator identity and git-email Owner defaults are included in the Issue data. |
| `update BEAD` with Issue scalar flags | Inline `--title`, `--description`/`--body`/`--message`, `--design`, `--acceptance`, `--priority`, non-claim `--assignee`, `--estimate`, `--external-ref`, `--spec-id`, `--due` and literal `--append-notes`. Requires `--if-revision TOKEN` or `--unconditional`. Description aliases must agree. Files/stdin and other Issue fields are unavailable. |
| `update BEAD --claim` | Atomically claim one Issue for the current actor using the native writer. Standalone `--claim=true` only; no other edits or revision/force guard. Repeating the same actor is a no-op and does not renew its five-minute lease. |
| `show RESOURCE` | Current Memory, Issue or Link; optional `--version TOKEN` selects an exact retained record. Use `versions` to list a Resource's versions in order. |
| `versions RESOURCE` | List one Memory, Issue or Link's retained versions newest first, each with its ordinal, version token, change time, actor and a `removed` marker. In a graph workspace `history RESOURCE` is an alias with the same output; ordinary workspaces keep the Dolt-commit `history`. |
| `compare RESOURCE --from TOKEN --to TOKEN` | Compare two complete retained preview versions of one Memory, Issue or Link. Explicit tokens determine direction, not chronology. |
| `link SOURCE TARGET --link-type TYPE` | Use an installed Link Type as `types/NAME` or its full local URL. Informational Links permit `--id links/PATH`, `--properties JSON` and source guards. Memory sources own informational Links; Issue sources do not. |
| `dep add SOURCE TARGET` or `link SOURCE TARGET` | A local blocking Dependency between Issues, using the ordinary default `blocks` type. No bulk, remote, routing or bypass flags. |
| `update LINK --properties JSON` | Replace all informational Link properties. Requires a Link guard; the Memory source guard is optional. Blocking Dependency properties are not editable here. |
| `links BEAD` | Complete bounded current incident Links, with optional `--direction in\|out\|both` and `--link-type TYPE` filter. No pagination. |
| `unlink LINK` | Remove one informational Link or blocking Dependency by canonical ID. Requires a Link guard; the source guard is optional for Memory-owned Links, required for blocking Dependencies. |
| `unlink SOURCE TARGET --link-type TYPE` | Remove an unambiguous informational Link with the same guards. Multiple matches refuse and report candidate IDs. Blocking Dependency pair removal is unavailable. |
| `close BEAD` | Close one Issue through the existing Issue policy, optionally with ordinary reason aliases. No force or batch operations. |
| `reopen BEAD` | Reopen one Issue, optionally with `--reason`. |
| `ready` | Unfiltered current ready Issues through ordinary readiness rules. No list filters, output limit or configured positive `BEADS_MAX_ROWS`. |
| `list` or `list --format records-json` | Without an Issue filter, one bounded current snapshot of every Memory and every Issue the ordinary `bd list` would show: closed and pinned Issues are hidden unless `--all`, which also lifts the row limit. Newest recorded change first, ties by canonical Bead ID; human rows show local ID and kind, and for an Issue its status and priority, then the title. `--bead-type types/NAME` narrows by nominal Bead Type. A positive `BEADS_MAX_ROWS` refuses a page of more Beads than the cap. Any Issue filter (status/state, type, title/title-contains, priority and range, assignee/no-assignee, label/label-any/exclude-label, pinned/no-pinned, due-before/due-after/overdue, sort or reverse, or a matching configured directory label) selects the native Issue-only query, which lists Issues only, says so under the header in human output, and omits closed and pinned Issues unless `--all` or a filter selects them. See [All-Bead listing](#all-bead-listing). `hasMore` reports whether a row limit omitted matches; tree and legacy JSON remain unavailable. |
| `blocked` | Complete native dependency-blocked Issue view with canonical blocker IDs. No filters or positive `BEADS_MAX_ROWS`. |
| `graph BEAD --view generic` | Current local summary traversal with `--direction in\|out\|both`, `--depth`, `--max-nodes` and `--max-links`. |
| `status --graph` | Report the capabilities and bounds admitted by this checkpoint, including initial Issue fields/notes, append-only notes, estimate/reference edits and due-date authoring/filtering. |
| `serve --readonly --addr HOST:PORT` | BDP Read over HTTP for an ordinary shared-server graph workspace. Existing token-file authentication, Host controls and non-loopback opt-in apply. Embedded serving is refused. |
| `setup claude [--project] [--check\|--remove]` | Add, check or remove only the project-local Claude Stop hook (`bd claude-hook stop`); adding also ensures `CLAUDE.md` imports the graph guidance. Adding and `--check` require current graph-preview guidance. Global/stealth setup, other recipes and `--json` are unavailable. |
| `claude-hook stop` | Steph's Stop reminder, which Claude Code runs with its JSON hook input on stdin; it does not open storage. If graph admission refuses it (for example, `BD_BACKEND` is set or the workspace was moved), it prints one warning line and exits 1, which Claude Code does not treat as blocking. |

Commands accept the common graph controls `--json`, `--graph-mode`, `--actor`,
`--quiet`, `--no-color`, `--directory` and `--readonly` where applicable. A
read-only invocation cannot write. Unsupported command options refuse rather
than falling through to ordinary storage operations.

`create` defaults to an Issue. `--bead-type types/preview-memory-v2` selects a
Memory; `types/preview-issue-v2` explicitly selects an Issue. Full local Type URLs
also work. `--type` remains Issue classification (such as `task` or `bug`). Memory
creation accepts an optional ID, positional title or `--title`, and inline
`--description`/`--body`/`--message`. Without a title, body text supplies the summary.
Issue-only fields are refused on Memory creation. For example:

```sh
bd create --bead-type types/preview-memory-v2 --body 'Code flow policy: target integration.'
```

Run `bd types` in the selected graph workspace to see the Type IDs usable with
`--bead-type` and `--link-type`; `bd types --details` displays the full stored
descriptors. This graph Type catalog is distinct from the ordinary Issue
classifications selected by `bd create --type`.

New workspaces install three informational Link Types: `types/preview-related-v2`,
`types/example-follows` (the source follows a policy described by the target),
and `types/example-cites` (the source cites the target as context). All accept
Memory or Issue endpoints and optional `note` properties; none affects scheduling.
The blocking Type is `types/preview-blocks-v1` and requires Issue endpoints.
Use scope-relative `types/NAME` or the full local Type URL. Existing four-Type
workspaces remain readable and writable with their original Types; reads do not
install the two examples. Arbitrary Type installation is unavailable.
`--resource-type` remains a hidden compatibility alias for `--link-type`; do not
supply both.
For example:

```sh
bd remember 'Context for the plan.' --id beads/context --title Context
bd create 'Ship the release' --id beads/work --priority 1
bd link beads/plan beads/context \
  --link-type types/preview-related-v2 \
  --id links/context --properties '{"note":"background"}'
bd link beads/plan beads/work \
  --link-type types/preview-related-v2 \
  --id links/work
bd links beads/plan --json
bd update links/context --properties '{"note":"revised background"}' \
  --unconditional
bd unlink links/context --unconditional
```

`--unconditional` explicitly accepts the current record; use an observed
`--if-revision TOKEN` to reject a stale write. `remember --update` defaults to
unconditional acceptance when neither flag is supplied. Property replacement,
property patches, Memory deletion and Issue edits still require an explicit
revision or unconditional choice. Source guards are
`--if-source-revision TOKEN` or `--unconditional-source`. Memory-owned Link
writes default to unconditional source acceptance; `--if-source-revision` opts
into stale-source refusal. Explicit `--unconditional-source` remains accepted.
The Link guard itself is still required for edits/removal, and blocking
Dependency removal still requires its Issue-source guard.
An unconditional changed Memory write, including an owned-Link change,
discloses the actual replaced version and attribution. Guarded writes and
semantic no-ops do not claim an unconditional overwrite. Targets do not acquire
new versions merely because another Bead links to them.

Unlink removes current Link state while reserving its identity and retaining
private prior snapshots. It does not erase a Memory or promise irreversible
erasure, restoration or identifier reuse.

## Ordered property changes

`--patch` applies a nonempty ordered array to the actual guarded predecessor
inside the existing write transaction. It cannot be combined with
`--properties`, selected Memory fields or Issue update flags. For example:

```sh
bd update beads/plan --if-revision OBSERVED_MEMORY_REVISION --patch \
  '[{"op":"replace","path":"/title","value":"Release plan"},{"op":"replace","path":"/body","value":"Updated context."}]' --json
bd update links/context --if-revision OBSERVED_LINK_REVISION \
  --if-source-revision OBSERVED_MEMORY_REVISION \
  --patch '[{"op":"add","path":"/note","value":"Updated background"}]' --json
bd update beads/plan --if-revision OBSERVED_MEMORY_REVISION --patch @changes.json
```

Use fresh observed tokens for each changed write. `--unconditional` accepts
the current Resource; for a Memory-owned Link, the default `--unconditional-source` is a
separate decision. An Issue-source informational Link does not require a
source guard, but any supplied source guard is checked. Patching that Link
does not change its Issue source. Issue properties and blocking Dependency
properties cannot be patched through this route.

Operation objects contain only `op`, `path`, and, for add/replace, `value`.
Pointers use JSON Pointer escapes (`~0` for `~`, `~1` for `/`); an empty
pointer selects the root. Add inserts array items or replaces an object
member, and `-` appends to an array. Replace/remove require an existing
target; missing parents are never synthesized. Operations run in order and
may temporarily introduce arrays or other JSON values. The final Memory
properties must contain exactly the `title` and `body` strings; final Link
properties permit only an optional string `note`. Empty strings are present
values, distinct from absent members or null.

All operations succeed together or leave the stored state unchanged. A
same-value or reversing patch is a semantic no-op: the existing version,
revision and attribution remain unchanged, with no replacement disclosure.
Stale guards still refuse before no-op evaluation. A changed unconditional
write discloses the actual replaced Memory state, including when the change
is to a Memory-owned Link. Prior complete states remain available through
the exact retained reads described below. Deleted Memories and removed Links
cannot be revived by patching.

Input and each working properties document are limited to 1 MiB, with at
most 256 operations, 4,096 pointer bytes, and 64 pointer segments/nesting
levels. A cumulative 16 MiB evaluation-byte limit bounds repeated document
work; it is not an exact heap or timing bound. Duplicate members, invalid
Unicode, inexact unsupported numbers, extra operation members and unsupported
operations refuse. Files/stdin are read only when explicitly selected and
after readonly/freeze, selector and local guard admission. Input/parse errors
report `invalid_properties`; stale guards report `revision_conflict`.

Changed patches must also fit the complete current-read acquisition budget,
including retained final heads of deleted Memories. Failure rolls back the
mutation and retained records. Existing replacement/selected-field writers
keep their established admission policy. A no-op writes nothing and does not
repair an oversized workspace; a properties document already above the
patch limit cannot use this route to shrink itself. Runtime bound refusals
report `capability_unavailable`. These bounded CLI changes do not enable a
public BDP Write profile, a public History contract or durable request outcomes.

## Find and inspect retained Memory

`memories` searches current Memory titles and bodies using literal Unicode case
folding. It reads one checked current snapshot before filtering. The default
must match at most 50 Memories; a larger result refuses instead of returning a
partial page. Narrow the search or use `--all` for all matching summaries within
the same acquisition and output bounds. Default human output shows local IDs and
titles, plus labeled excerpts for body matches. `--details` adds exact versions,
attribution, owned-Link counts and recall commands. Structured `records-json`
output retains its full existing summary fields.

```sh
bd memories release
bd memories release --details --format records-json
# Use --details to copy an exact version for a retained read.
bd memories release --details
bd recall beads/plan --version SAVED_TOKEN
bd show beads/plan --version SAVED_TOKEN --json
bd compare beads/plan --from EARLIER_SELECTED_TOKEN --to OTHER_SELECTED_TOKEN --json
```

Structured summaries include identity, title, saved version, attribution, matched fields
and a body excerpt when the body matches. Search input is limited to 4,096 UTF-8
bytes, excerpts to 160 code points and final output to 1 MiB. A short matching
body can appear in full as its excerpt. Results are sorted by canonical ID,
marked complete and have no continuation. The existing 1,000 live Resource and
16 MiB acquisition bounds apply before filtering, including retained deleted
Memory heads counted by the current reader.

Use `--format records-json` for experimental summaries; `--json` has no settled
graph Memory-list mapping and refuses. Explicit `--format table` or
`--format records-json` takes precedence over ambient JSON configuration.
Ordinary KV-memory workspaces retain their existing output and `--format json`
alias. BDP Read remains the protocol interface for scripts.

`recall` writes the selected body's exact bytes with no envelope, added newline
or quiet suppression. Empty content succeeds with zero output bytes. Without
`--version`, it reads current Memory; with a token from a saved result it reads
that retained body even after edits or deletion. Graph JSON recall remains
unavailable because the complete Memory representation is unresolved.

`show --version` also supports retained Issue and Link records, including their
saved owned Links. Tokens are opaque nonempty UTF-8 strings of at most 4,096
bytes. `--version` does not interpret local ordinals, timestamps, full version
URLs or as-of selection; it takes only a token. An unknown or other-resource token returns `revision_unknown`;
a missing subject returns `not_found`. A removed Link's prior live versions
remain readable; its private removal token returns `gone`. A deleted Memory
retains its final live head without inventing a deletion version.

`compare` reads both operands in one authority-checked transaction. Equal tokens
still require a valid retained record. Either unavailable or corrupt operand
fails the whole operation. Differences include complete property values and
owned Links matched by canonical identity, with explicit presence so absence,
null and empty content remain distinct. Arrays retain order; object member
order is ignored; JSON numbers are compared without float rounding. Reversing
the tokens reverses the comparison. The selected versions and attribution are
context rather than property differences. Immutable identity or endpoint
mismatches refuse.

Comparison output is provisional JSON, indented for human output. Its
`compared` and `unsupported` fields describe the projection: common metadata and
Memory inception/derivation are unavailable. Issue comparison uses the native
retained durable projection; current row-lock and content-hash fields are not
reconstructed. Each retained snapshot acquisition has a 16 MiB input budget. A deleted
Memory also validates its final live snapshot before reading an older selected
snapshot, so a distinct pair can acquire up to 64 MiB in total (up to 32 MiB
for equal older tokens, which are resolved once). This excludes decoding/output
overhead and is not a total-heap or rendered-output cap. These read commands work under `--readonly` and migration freeze. They
do not order versions themselves (see `versions` below), and they do not enable
HTTP History, restoration or a stable public comparison contract;
`historyExact` remains false.

## List a Resource's versions

`versions` lists the retained versions of one Memory, Issue or Link, newest
first. In a graph workspace, `history` is an alias that produces the same
output. In an ordinary workspace `history` is unchanged and still reports
Dolt commits.

```sh
bd versions beads/plan
bd history beads/plan     # same output, graph workspaces only
# Feed a listed token to an exact read or a comparison.
bd show beads/plan --version LISTED_TOKEN
```

Each row carries `ordinal`, `version`, `change_at`, `actor`, `attribution`
and `removed`; `--json` reports them under those names, inside a result that
also names the `resource` and its `kind`. `version` is the
opaque token and the only citable address for a version. `ordinal` is an
ordering key, not an address: it is local to one Resource in one store and
cannot be passed to `show --version` or `compare`. The field is deliberately
not named `revision`, which already means the opaque token in graph records
and the row-lock token on native Issues. `change_at` is for display and is
never used to order rows, so two versions written within the same second
still list in write order. Memory and Link ordinals are allocated per
Resource when each version is written. Issue rows come from the native Issue
version record, use its revision as the ordinal and populate `attribution`
with its attribution status; Memory and Link rows leave `attribution` empty.
A deleted Memory's list ends at its final live head; deletion adds no version
to it.

`removed` is false on every row except one: a removed Link's deletion marker.
That row is listed, as the Link's newest version, because the removal is part
of its history. It is not citable: `show --version` refuses its token with
`gone`, as it always has. `removed: true` means "listed but not citable".
Every other listed token is one `show --version` and `compare` accept.

The command has two answers:

- An ordered list, newest first. A Resource's creation is its version 1,
  written in the same transaction that allocates it, so the list always has at
  least one row. Deleted Memories and removed Links still list their history.
- `not_found` when nothing was ever allocated at that path.

An allocated Resource with no versions is corruption, not an empty history,
and the command refuses rather than printing an empty list. Every plane in a
schema version 6 workspace can order its versions. A history too large to
return (more than 1,000 versions, or over the 16 MiB read budget) still
refuses with `capability_unavailable`, because limit refusals use that code;
there is no pagination.

Memory and Link ordinals are allocated as `MAX(ordinal)+1` per Resource inside
the writing transaction. Every graph write also updates one store-wide writer
fence, so two concurrent writers in one store always contend: the loser fails
at commit with `revision_conflict` (exit 4) and nothing is retained for it. A
unique `(path, ordinal)` key backs this up. Issue ordinals are native revisions
and are allocated by the native Issue writer.

This is a CLI listing only: no
HTTP History route or public History contract is added, `serve` publishes no
History, and the list does not support as-of selection or restoration.
`status --graph` reports `versionList: true` for this command;
`historyExact: false` continues to describe the HTTP profile, where exact
History remains unavailable.

## Unreferenced Memory deletion

`delete` previews one current Memory without changing it, including under
`--readonly`. Apply with `delete --force` or `forget`; both require the existing
revision guard or an explicit unconditional choice. The `result.memory` in
both preview and apply is the checked final live Memory, including its existing
revision and attribution. `result.preview` distinguishes the read-only preview;
`result.deleted` is true only after successful apply. The outer `preview: true`
continues to identify the experimental graph format, including applied writes.

```sh
bd remember '' --id beads/scratch --title Scratch --json
bd delete beads/scratch --readonly --json
# Supply the revision observed in result.memory.revision from the preview.
bd delete beads/scratch --force --if-revision OBSERVED_REVISION --json
# Or apply to another unreferenced Memory with explicit unconditional intent:
bd remember 'Temporary note' --id beads/another-scratch --title Scratch
bd forget beads/another-scratch --unconditional --json
```

An empty body is present content and can be deleted. Successful deletion removes
current Memory state while preserving the canonical ID reservation and immutable
prior snapshots. A new `show` reports `gone`; recreating the same ID reports
`identity_reserved`. Repeated deletion refuses. Deletion does not create a new
live revision or a deletion version, and the result does not claim a deletion
timestamp or actor event. Internal snapshot retention does not expose public
History or restoration.

Every live incident Link causes both preview and apply to refuse with
`deletion_policy_unresolved`, including incoming, outgoing and self-Links.
Explicitly unlinking each Link with its normal guards allows a later deletion;
these are separate transactions, and a concurrently added Link can still cause
deletion to refuse. `--force` does not cascade. Batch selectors, `--from-file`,
`--cascade`, `--dry-run`, erasure and Issue deletion are unavailable in this
graph slice. Ordinary Issue deletion and key/value `forget` remain unchanged.
`status --graph` advertises `memoryUnreferencedDelete`; general `memoryDelete`
remains false because linked Memory deletion policy is unresolved.

## All-Bead listing

`bd list` has two modes, selected by its flags.

Without an Issue filter (bare `bd list`, or only `--flat`,
`--format records-json`, `--bead-type`, `--limit` and `--all`), the command
reads one checked current snapshot and lists its Beads: every current Memory
and every current Issue that the ordinary `bd list` would show. Closed and
pinned Issues are hidden, exactly as there: an Issue with status `closed` or
`pinned`, a set pinned flag, or a custom status in the done or frozen
category. `--all` shows them and also removes the row limit. A Memory has no
status and is never hidden.

Beads are ordered by when their current version was recorded, newest first,
and by canonical Bead ID where two share an instant. The order is the same
for every kind of Bead. It is a display order: a recorded time is a
wall-clock reading, not an acceptance-order token. Human output prints one
unquoted row per Bead: the local ID and kind, then for an Issue its status
and priority, then the title. For a workspace holding the Memory
`beads/plan` and the Issue `beads/work`, with `work` recorded last:

```text
Beads (2; more: false; graph preview)
  beads/work  Issue   open  P1  Ship the release
  beads/plan  Memory  Release plan
```

`--format records-json` returns the same Beads, in the same order, as
complete Memory and Issue records in `result.items`, with `result.hasMore`.
An optional `--bead-type types/NAME` keeps one installed Bead Type; Link
Types and uninstalled Types refuse with `capability_unavailable`, and a
selector that is neither `types/NAME` nor a full local Type URL refuses with
`invalid_selector`. `--bead-type types/preview-issue-v2` on its own stays in
this mode, so it hides closed and pinned Issues too unless `--all` is given.
`--limit N` returns the first N Beads and a truthful `hasMore`, not a
continuation cursor. A positive `BEADS_MAX_ROWS` refuses a page of more
Beads than that: it counts the Beads the command would print, after hiding
and after `--limit`, so a limit at or under the cap never trips it. The
snapshot inherits the 1,000-live-Resource/16 MiB acquisition bounds,
including Links that are not printed. For paginated all-Bead enumeration,
use the BDP HTTP `beads/` collection.

Any Issue filter selects the existing native Issue query instead. The Issue
filters are `--status` (or `--state`), `--type`, `--title`,
`--title-contains`, `--priority`, `--priority-min`, `--priority-max`,
`--assignee`, `--no-assignee`, `--label`, `--label-any`, `--exclude-label`,
`--pinned`, `--no-pinned`, `--due-before`, `--due-after`, `--overdue`,
`--sort` and `--reverse`. A supplied flag counts even when an empty value
such as `--assignee=` or `--sort=` adds no restriction. A configured
`directory.labels` entry that matches the current directory supplies an
implicit label filter, so it also selects this query, unless `--bead-type`
names a non-Issue Type: a label does not apply to a Memory, so the entry is
not applied there and nothing is refused. The query returns Issues only,
never Memories, and omits closed and pinned Issues unless `--all` or a filter
selects them. Human output says so on the line under the header:

```text
Issues (1; more: false; graph preview)
Memories are not listed (Issue query selected by: --status).
"https://example.org/team/beads/work" "open" P1 "Ship the release"
```

With a configured label the line reads
`Memories are not listed (Issue query selected by: directory.labels "frontend").`
Structured output carries no such line. A typed Issue filter cannot be
combined with a non-Issue `--bead-type`; that refuses with
`capability_unavailable`.

Both modes apply the ordinary `bd list` row limit. An explicit `--limit N`
wins, with `0` meaning no limit; otherwise `--all` means no limit; otherwise a
configured `list.limit` applies; otherwise there is no limit when output is
piped, 20 rows in agent mode at a terminal and 50 rows at other terminals.

## Read-only Issue queries and traversal

```sh
bd list --format records-json --status open --limit 1
bd list --flat --all --sort title
bd blocked --readonly --json
bd graph beads/plan --view generic --direction out --depth 2 --json
```

When an Issue filter (listed under [All-Bead listing](#all-bead-listing)) is
supplied, listing reuses the existing native Issue query, configuration and
limit policy.
It accepts status (or state), type, title/title-contains, priority and priority
range, assignee/no-assignee, label/label-any/exclude-label, pinned/no-pinned and
due-before/due-after/overdue filters. Sorting accepts
priority, created, updated, title, status or type; reverse is supported. Explicit
limit wins over `--all` and configured limits. The result contains complete
canonical Issue records in `items` and a truthful `hasMore` boolean. It is a new
read each time, not a snapshot cursor or BDP continuation. Returned records and
the extra probe record are validated before trimming the page.

Human output is flat by default; `--flat` is accepted and changes nothing.
The Issue query prints one row per Issue: its quoted canonical ID and status,
its priority and its quoted title, for example
`"https://example.org/team/beads/work" "open" P1 "Ship the release"`.
A line under the header says that Memories are not listed and names the
option that selected the query; structured output carries no such line.
`--format records-json` returns the experimental graph envelope. Tree output
(`--tree`, `--pretty` or `--flat=false`), `--json`, `--format json`, watch,
readiness, parent/ID/routing/offset selectors, repeated
status/state/type/assignee/bead-type filters and supplied-empty labels
refuse. Defer filters remain unavailable; due selection
does not schedule or wake work.
This does not change ordinary list parsing or implement contributor-owned filter
unions. Records-json selects structured errors and overrides ambient human
format selection; explicit `--json` still refuses. BDP Read remains the documented
script interface below; these CLI result shapes are experimental.

`blocked` follows the existing native dependency-blocked query. It is not the
complement of ready, nor a query for every manually blocked or deferred Issue.
The result is an array of objects with complete `issue` records and canonical
`blockedBy` IDs, sorted by canonical ID. Empty is `[]`. It accepts no positional
selectors or filter flags, even explicit empty values. A positive
`BEADS_MAX_ROWS` refuses; zero or unset permits the complete bounded view.
No query wakes deferred work, repairs blocked state, creates versions or opens
the ordinary store. Readonly and migration freeze permit these reads.

Generic traversal reads one checked current snapshot. Nodes expose only ID,
Type, title, version and attribution; Links expose ID, Type, source, target,
version and attribution. Memory bodies, Issue long text and Link properties
are omitted. Directions are in/out/both; depth defaults to 1 and accepts 0..1000.
The root counts as one node. Default node/Link caps are 100/200, each accepting
1..1000. Distinct Link identities survive repeated-node visits, including diamonds
and self-Links. `frontier` identifies reached nodes with eligible unexpanded
Links; `complete` says whether that frontier is empty. Exceeding a node/Link cap
refuses rather than silently truncating. Foreign roots, external endpoints,
custom Types, legacy graph options and positive `BEADS_MAX_ROWS` refuse.

All three views retain the shared whole-workspace 1000-live-Resource and 16 MiB
acquisition bounds, including unrelated content and retained final heads of
deleted Memories. List config acquisition/resolved policy is additionally bounded
to 64 KiB and 256 rows/entries; database failures refuse instead of silently
choosing YAML fallback. List and blocked output have a 16 MiB bound; traversal
summary output has a 1 MiB bound. These are operational bounds, not heap guarantees.
Output is prepared before emission; physical output failures can still occur
after some bytes have been written.

A valid unreferenced Memory deletion does not invalidate an Issue query. Fresh
views exclude it, while its identity and final live retained state remain reserved.
Traversal cannot emit a deleted endpoint; a removed root is absent from its live
snapshot. Corrupt deleted allocations still refuse, including an invalid retained
head or an inconsistent current payload/Link. No new tombstone representation,
public History surface or external traversal contract is introduced.

## Priority and assignment

```sh
bd show beads/work --json
bd update beads/work --priority P0 --assignee alice --if-revision OBSERVED_REVISION --json
bd list --assignee alice --format records-json --all
bd update beads/work --assignee= --if-revision NEW_REVISION --json
bd list --no-assignee --format records-json --all
```

Use the opaque revision from the preceding `show` or accepted mutation. The
revision covers the Issue and its owned blocking Dependencies. Priority uses
ordinary P0–P4 meanings; omission preserves it, and zero explicitly sets P0.
Assignment preserves the literal value, including whitespace; an explicit empty
value clears it. Text, priority and assignment can be combined in one guarded
operation and one retained version. A matching same-value edit is a no-op;
a stale revision still refuses, even if every requested value is already current.

Assignment does not claim work, start a lease or change status. The ordinary
active-holder fence still applies to guarded and unconditional transfers. An
authorized transfer or clear removes the prior lease; unrelated edits preserve
it. Clearing an in-progress assignment does not reopen the Issue. Graph claim is
available only in the standalone form below; heartbeat and reclaim remain unavailable. Changes made through an external lease
writer have no supported graph-preview repair path.

Assignee listing reuses native SQL comparison/collation, not actor identity
normalization. `--no-assignee` includes native empty/NULL assignments. An empty
`--assignee=` supplies no assignee restriction but still selects the Issue
query; combining a nonempty assignee
with `--no-assignee` intersects the filters and returns no matches. Listing does
not modify state or claim work. Its output remains the experimental CLI envelope;
BDP HTTP Read is the script interface and exposes the current Issue properties.

## Append-only Issue progress

```sh
bd show beads/work --json
bd update beads/work --append-notes 'Finished the first pass.' --if-revision OBSERVED_REVISION --json
```

The native transaction appends to the current notes. Omission preserves notes;
empty on empty is a no-op, while empty on nonempty appends one newline. Text is
literal, including whitespace, Unicode, CRLF, `-` and `@file` markers; this flag
does not read files or stdin. Repeating nonempty text appends it again. A stale
revision refuses even when the append would otherwise be a no-op. An unknown
commit outcome must be inspected; never automatically replay an append.

Append can accompany supported scalar edits in one guarded update and one
retained version. Notes-only edits preserve owned Dependencies, claim lease
fields and closed state. Assignment mixed with append retains the native holder
fence and lease rules. A changed request supplying append intent must leave the
workspace within its existing read acquisition budget, including its new retained
snapshot, or every effect rolls back. This conservative postcondition also applies
to an empty append combined with a changing scalar. Scalar-only repair paths and
true no-ops keep their existing behavior. No new per-input notes limit is added.

Replacement/clear, force overwrite, and combined claim-plus-append remain
unavailable. Contributor notes replacement safeguards are a separate review gate.
CLI current/exact reads and BDP Read expose the accepted notes; no HTTP write or
public History contract is added.

## Standalone Issue claim

```sh
bd update beads/work --claim --actor alice --json
```

The native claim writer decides eligibility, including configured active statuses
and literal claim-pool aliases. A successful claim sets the assignee and
`in_progress` status, preserves an existing start time, and records one native
version and one complete graph version. It does not require the Issue to be ready.
A foreign holder or non-claimable status is refused without mutation.

The native lease lasts five minutes. Repeating the claim as the same actor
(including equivalent actor spelling) is a no-op, including the lease timestamps.
This preview does not provide heartbeat, renewal, reclaim, unclaim or lease repair.
An external writer that changes the lease bypasses the saved graph projection;
current graph reads and writes then refuse that mismatch. Previously retained
snapshots remain readable. Do not use external lease commands on these preview
Issues expecting the graph to repair them.

`--claim=false`, combined claim plus scalar/property edits, revision guards and
unconditional/force claim flags are refused. A lost commit reply reports an
unknown outcome; callers must inspect rather than automatically retry the write.

## Issue due dates

```sh
bd create 'Prepare the review' --id beads/review --due '2030-01-02T10:00:00Z' --json
bd update beads/review --due '+30min' --if-revision OBSERVED_REVISION --json
bd list --format records-json --due-after '2030-01-01' --due-before '2030-01-03'
bd list --format records-json --overdue --all
bd update beads/review --due= --if-revision OBSERVED_REVISION --json
```

Create and guarded update reuse the ordinary date parser, including `+30min`,
`+6h`, date-only and offset-bearing timestamps. `m` means calendar months;
`min` means minutes. Timezone-less input uses the invoking process's local zone.
Omission preserves the existing due date; explicit empty update clears it.
An empty create value leaves the date absent. Due edits may accompany other
admitted scalar edits or notes append in the same native transaction/version;
standalone claim still cannot be combined with an edit.

This private preview retains the existing whole-second SQL representation:
writes normalize to UTC and round to the nearest second, with half-second ties
forward. Filter cutoffs normalize to UTC and truncate fractions before native
strict `<` / `>` comparisons. A due time of `12:00:00Z` therefore does not match
a before-cutoff of `12:00:00.900Z`. Dates must remain within years 1..9999 after
normalization. These are provisional compatibility rules, not a durable BDP
precision contract. The CLI's 4096-byte UTF-8 date-input bound is likewise a
private preview admission limit.

The ordinary filters intersect, including existing status/type/limit policy.
Overdue means a non-null due date before the native UTC query clock, excluding
closed Issues even with `--all`; it neither changes status nor schedules work.
After normalization, an unchanged due value or repeated clear is a no-op, while
a stale revision still refuses. Current and exact retained reads preserve the
accepted instant; due writes retain existing links, lease and lifecycle state.
No defer, recurrence, mandatory deadline or scheduler behavior is added.

## Limits and remaining work

Linked Memory deletion and Issue deletion,
defer/scheduling, notes replacement/clear, label mutation,
and full Memory remain unavailable. Issue creation can set initial
labels; that does not adopt a label-editing contract. These restrictions apply
to graph workspaces; ordinary Issue workspaces keep their existing behavior.

BDP HTTP Read is available on ordinary shared-server Dolt as described below.
CLI JSON is a separate command output format. No public History route is
enabled by the exact retained CLI reads or experimental comparison.

Writes retain their complete accepted state within the existing transaction.
Those snapshots are not a claim of complete native History, Dolt HEAD/sync
durability, public BDP Write or recovery support. The current read acquisition
budget can refuse an oversized workspace; this checkpoint does not promise
that every successful create preserves that aggregate budget. `status --graph`
reports the current bounds.

## BDP Read from scripts

A graph workspace on ordinary shared-server Dolt can expose its current records
through BDP HTTP. The protocol and independent public client are pinned to
`gastownhall/bdp@53bdbd03136875f952af184fce7b3c7af8f74e96`. This serves the Read
profile; it does not enable HTTP writes or History.

Choose a stable, reachable HTTP address **before** initializing a new workspace.
The persisted Scope URL supplies record identities and the HTTP path. It does
not change when an allowed Host alias or reverse proxy reaches the listener.
For a disposable local example, with an ordinary Dolt server already listening
on port 3306, use a new directory:

```sh
git init
bd init --graph-mode link --scope-url http://127.0.0.1:8765/demo/ \
  --server --external --server-host 127.0.0.1 --server-port 3306 \
  --server-user root --skip-hooks --skip-agents --non-interactive
bd remember 'Why we chose this design.' --id beads/plan --title Plan
bd serve --readonly --addr 127.0.0.1:8765
```

From another terminal, read actual BDP representations:

```sh
curl --fail-with-body -i http://127.0.0.1:8765/demo/
curl --fail-with-body http://127.0.0.1:8765/demo/bdp.json
curl --fail-with-body http://127.0.0.1:8765/demo/beads/plan
curl --fail-with-body 'http://127.0.0.1:8765/demo/beads/plan?view=properties'
curl --fail-with-body 'http://127.0.0.1:8765/demo/beads/?limit=1'
curl --fail-with-body http://127.0.0.1:8765/demo/types/preview-memory-v2
curl --fail-with-body 'http://127.0.0.1:8765/demo/beads/plan?view=links&direction=both'
```

The Scope root returns a `service-desc` Link to `bdp.json`. Collections expose
`items` and `next`; enumerate by following each complete `next` URL until it is
null. A first page alone is not the full inventory. Bead, Link and Type
collections are available at `beads/`, `links/` and `types/`. Bead/Link
collections support `type`, `conformsTo` and bounded `selector` filters; Link
collections also support `source`, `target` and `endpoint`. Incident HTTP
directions are `inbound`, `outbound` and `both`. `include=links` returns a Bead
and its first incident Link page from one snapshot.

GET and HEAD use canonical current identities. Resource/properties responses
carry revision ETags. Authentication, authority, query and storage validation
precede conditional responses: an invalid cursor or missing Resource cannot
become a successful 304. No Last-Modified timestamp is invented. Historical
version queries, aliases, HTTP mutation, legacy `/v0` and `/healthz` routes are
unavailable. Use an authenticated Scope or discovery request for readiness.

A collection continuation retains its original records and owned Link state
even if a later CLI write changes fresh reads. Cursors expire after five
minutes and do not survive service restart. This is retained pagination, not
chronological History. Stop the service before restoring storage; same-binding
hot restore and out-of-band SQL are unsupported.

The [standard-library Python example](https://github.com/versioned-beads/beads/tree/integration/examples/bdp-read)
follows every continuation and can also read one canonical Bead. From this
repository's root, with the example listener running:

```sh
python3 examples/bdp-read/read_beads.py --scope http://127.0.0.1:8765/demo/ --limit 1
python3 examples/bdp-read/read_beads.py --scope http://127.0.0.1:8765/demo/ \
  --id http://127.0.0.1:8765/demo/beads/plan
```

For an authenticated listener, supply its token through the `BDP_TOKEN`
environment variable. The example reads BDP HTTP directly and prints results
only after the requested read or full enumeration succeeds.

### HTTP access and limits

The existing `serve` controls apply: `--auth-token-file`, `--allowed-host`,
`--allow-non-loopback` and explicit `--insecure-no-auth`. The default listener
is loopback; its existing local trust policy permits no-token access there.
For a token-protected listener, provide a bearer token on every request,
including continuations. Token-file rotation keeps the existing last-good-file
reload policy. Every accepted token sees the complete workspace; separate
per-user authorization views are not implemented. Responses use
`Cache-Control: private, no-store`. The server provides no TLS or CORS support.

The store stays open until HTTP requests have drained. Serving opens only an
already initialized, matching ordinary shared-server graph authority; it does
not create a workspace or enable legacy writers. Embedded CLI operations remain
supported, but embedded HTTP serving refuses before opening storage or binding
a listener.

Page size defaults to 100 and is capped at 1,000. The current inventory is
limited to 1,000 live Resources and a conservative 16 MiB persisted-byte
acquisition budget before filtering. That budget can refuse a small or exact
read because other current data is oversized. HTTP representations are limited
to 8 MiB and retained snapshots to 7 MiB; at most 32 snapshots, 4,096
continuation positions and 32 MiB total are retained. Request targets are capped
at 32 KiB, Selectors at 16 KiB, depth 256 and 2,048 nodes. Capacity pressure
refuses new snapshots without evicting valid continuations. These are preview
bounds, not a guarantee about Go heap usage or production migration.

## Regression coverage

The required Graph C0 gate checks source/CLI provenance, storage tests on both
engines, graph configuration and CLI admission with no required skips. Its
installed initialization and fresh-process read checks remain in place. The
mixed workflow adds real CLI coverage for Memory edits, all three informational
endpoint combinations, Link replacement/unlink, blocking readiness and Issue
close/reopen/text editing. Storage tests verify retention, guards, no-ops,
concurrency, rollback and uncertain commit outcomes separately. Ordinary Issue
regression suites remain required. Qualification of any new combined commit
requires the actual destination CI results, not this document alone.

BDP coverage includes storage-to-wire projection on both real engines, bounded
selectors and retained pages, plus an ordinary-server HTTP listener exercising
authentication, token rotation, Host identity, authority loss, conditionals and
concurrent aggregate reads. The independent public-client capture uses the
installed CLI as its writer and real HTTP as its reader. Aggregate/replay
checks use real fetch plus the pinned public parsers where the public client
has no corresponding API. Combined-source qualification must include these
roots and captures without required skips; internal projection tests alone
are not installed HTTP proof.

The unreferenced deletion workflow adds normal installed CLI initialization on
both engines, empty-body preview/apply, stale and missing guard refusals,
read-only admission, fresh-process absence, ID nonreuse, incident-Link refusal
and explicit unlink before deletion. These tests do not implement or qualify
the unresolved linked-deletion policy.

## Issue authoring fields

```sh
bd create 'Implement the plan' --id beads/work --design 'Approach' --acceptance 'Done when verified' --assignee alice --estimate 45 --external-ref 'tracker #42' --spec-id 'spec/section' --notes 'Initial context' --json
bd update beads/work --estimate 0 --external-ref= --spec-id= --if-revision OBSERVED_REVISION --json
bd update beads/work --append-notes 'First pass complete.' --if-revision NEW_REVISION --json
```

Initial fields are part of one native create and retained record. Initial assignment
is not a claim and does not create a lease. Existing Owner and CreatedBy defaults
still apply, including the ordinary git-email Owner data already disclosed above.
Omitted estimates remain absent; explicit zero is present. Estimates use the
existing nonnegative signed SQL INT range; clearing an estimate to NULL is not
admitted. External and spec references are literal strings, not fetched URLs or
alternate Bead identities. The existing columns allow 255 and 1024 Unicode
codepoints respectively. Empty external reference on update clears it to NULL;
empty spec ID clears it to an empty string. CLI create omits an empty external
reference; the Go API preserves an explicitly supplied empty pointer.

Design, acceptance and initial notes are literal UTF-8, including whitespace and
line endings. Changed initial fields must survive storage hydration exactly.
Nonempty initial notes also require the completed current and retained state to
fit the existing workspace read budget before commit. Later notes changes use
append only; replacement and clear remain unavailable. Guarded scalar edits keep
owned Links and unrelated fields unchanged; an identical scalar edit records no
new version. Current CLI records, exact retained records and BDP Read carry these
fields without a new wire representation or HTTP write operation.
