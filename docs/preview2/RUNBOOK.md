# Beads Graph Preview 2 runbook

Use this guide with the Preview 2 `bd` program in a new workspace. The examples build on one another, so run them in order. This is experimental software: use practice data that you can recreate. See **Known Limitations** before using it for other work.

Some examples use `jq` to copy an ID or revision from a command result. Install `jq` if needed. If a value-capturing command prints nothing, stop and check the previous step before continuing.

## Start a workspace

Start in a new, empty directory outside any existing Beads workspace. The `--graph-mode link` flag selects the Preview 2 graph model. For now, `bd init` also requires a unique Scope URL even when you only use the CLI. **That is a temporary Preview 2 requirement.** The URL identifies your workspace; it does not contact or start a web service.

```sh
mkdir beads-preview2-test
cd beads-preview2-test
bd init --graph-mode link --scope-url "https://example.invalid/preview2/$(date +%s)/"
bd status --graph
bd types
```

Run the rest of the CLI examples in this directory. The optional BDP Read exercise near the end has its own setup, shown as terminal commands.

## Beads, Links, Types, and kinds

The system stores a graph: **Beads** are its nodes, and **Links** connect two Beads.

Every Bead or Link has a Type that describes its expected properties. A Link Type also says which Beads it can connect.

The installed Type catalog is fixed for this preview:

| Bead Types | Link Types |
| --- | --- |
| `preview-issue` | `preview-dependency` |
| `preview-memory` | `preview-related` |
| | `example-cites` |
| | `example-follows` |

Run `bd types` to see these in your workspace, or `bd types --details` to inspect their fields. `--bead-type` selects a Bead Type when creating an Issue or Memory. `--link-type` selects a Link Type when connecting Beads.

The CLI checks these Type rules when it creates or updates a record. For example, `preview-dependency` connects Issues, while `preview-related` can connect Issues or Memories.

The separate `--type` flag sets an **Issue kind**, such as `task` or `epic`, or a **Dependency kind**, such as `blocks` or `parent-child`. Both kinds are text values on a single installed Type. A custom Dependency kind may be entered, but only well-known kinds have built-in behavior; an unknown kind does not block work.

These examples show how the two Type selectors and the separate kind field interact:

| `--bead-type` | `--type` | Result |
| --- | --- | --- |
| Omitted | `task` (or omitted) | An Issue; the default kind is `task`. |
| `preview-issue` | `epic` | An Issue with `issue_type=epic`. |
| `preview-memory` | Omitted | A Memory. Issue kinds do not apply. |
| `preview-issue` | `random-kind` | Refused unless that Issue kind has been configured in `types.custom`. |
| `random-type` | Omitted | Refused because that Bead Type is not installed. |

| `--link-type` | `--type` | Result |
| --- | --- | --- |
| `preview-dependency` | `blocks` (or omitted) | A Dependency with `dep_type=blocks`. |
| `preview-dependency` | `parent-child` | A Dependency with `dep_type=parent-child`. |
| `preview-dependency` | `tracks` | A Dependency with a custom kind; it has no built-in blocking behavior. |
| `preview-related` | Omitted | A Related Link; this is a placeholder Type for Preview 2. |
| `random-type` | Omitted | Refused because that Link Type is not installed. |

An unconfigured random Issue kind is refused; an arbitrary nonempty Dependency kind of up to 32 bytes is accepted. This is a current CLI difference, not a difference between the installed Issue and Dependency Types.

Beads and Links both have IDs, properties, metadata, and saved versions. Properties are checked against the record's Type. Metadata holds extra information that is not specific to its Type.


The main commands operate on Beads, and a parallel `bd link` set operates on Links:

| Task | Beads | Links |
| --- | --- | --- |
| Create | `bd create` | `bd link create` |
| List | `bd list` | `bd link list` |
| Show one | `bd show` | `bd link show` |
| Update | `bd update` | `bd link update` |
| Delete | `bd delete` | `bd link delete` |
| List versions | `bd versions` | `bd link versions` |
| Compare versions | `bd compare` | `bd link compare` |
| Count | `bd count` (Issues) | `bd link count` |

Several familiar commands provide narrower, convenient forms of these operations:

| Command | Use |
| --- | --- |
| `bd remember`, `bd recall`, `bd forget`, `bd memories` | Create or update, read, delete, and list Memories. `recall` returns the body rather than the whole record. |
| `bd dep add`, `bd dep remove` | Create a Dependency between two Issues, or remove a blocking Dependency by its Issue pair. The pair-removal limits are listed below. |
| `bd history` | List saved versions of one record, like `bd versions`. |
| `bd new`, `bd q` | Short ways to create an Issue. |

`bd link query` is not available in this preview.


## 1. Beads, Issues and Memories


A **Memory** is a Bead that stores reusable information. These four examples all hold policies. These examples move from the general `create` form to shorter field flags and then the Memory-specific `remember` command.

### Create

Create the first Memory by supplying its Type-checked properties and separate metadata as JSON:

```sh
bd create --bead-type preview-memory --id policy-a \
  --properties '{"title":"Policy A","body":"Review before publishing."}' \
  --metadata '{"origin":"manual test"}'
```

Create the second with individual field flags:

```sh
bd create --bead-type preview-memory --id policy-b --title 'Policy B' --body 'Retain the review evidence.'
```

The positional argument can also supply the title:

```sh
bd create 'Policy C' --bead-type preview-memory --id policy-c --body 'Review policy changes before release.'
```

Finally, `remember` chooses the Memory Type for you and takes the body as its positional argument:

```sh
bd remember 'Keep the release checklist.' --id policy-d --title 'Policy D'
```

Create two kinds of Issue, `task` and `epic`. The first pair will demonstrate a blocking Dependency; the child and parent will demonstrate a parent-child Dependency later.

```sh
bd create 'Blocking prerequisite' --id blocker --type task
bd create 'Work waiting on prerequisite' --id blocked-work --type task
bd create 'Relationship child' --id rel-child --type task
bd create 'Relationship parent' --id rel-parent --type epic
```

### List Beads

The general `list` command can include all installed Bead Types. Narrow it to Memories with `--bead-type`, or use the Memory-specific `memories` command:

```sh
bd list --format records-json --all
bd list --format records-json --bead-type preview-memory --all
bd memories --all
```

The ordinary `--json` output of `bd list` shows Issues only for compatibility, even without `--bead-type`. Use `--format records-json` when you want both Issues and Memories.

```sh
bd list --json --all
```

### Read a Bead

`show` returns the full record. The Memory-specific `recall` command returns only its body:

```sh
bd show policy-a --format graph-json
bd recall policy-a
```

### Edit and compare versions

`update` changes only the named properties. `--replace-properties` replaces the entire writable properties object, so it must include every required field. `--patch` applies a list of changes. `remember --update` is another way to edit a Memory.

```sh
bd update policy-a --properties '{"body":"Review and qualify before publishing."}'
bd update policy-a --replace-properties '{"title":"Policy A","body":"Reviewed policy."}'
bd update policy-a --patch '[{"op":"replace","path":"/body","value":"Qualified policy."}]'
```

```sh
bd remember 'Review before publishing and after any major change.' --update policy-a
bd update policy-a --set-metadata team=docs
bd update policy-a --set-metadata reviewed=true
bd update policy-a --unset-metadata team
```

Every real edit saves a version. `versions` lists them; `history` is an alias for that list in this preview. A `revision` is an opaque value from a read that can guard the next write. If someone changed the record after your read, `--if-revision` rejects your edit.

```sh
bd versions policy-a --json
bd history policy-a
old_policy_revision="$(bd show policy-a --format graph-json | jq -er '.result.revision')"
bd update policy-a --set-metadata review=complete --if-revision "$old_policy_revision"
new_policy_revision="$(bd show policy-a --format graph-json | jq -er '.result.revision')"
bd show policy-a --version "$old_policy_revision" --format graph-json
bd compare policy-a --from "$old_policy_revision" --to "$new_policy_revision" --json
```

### Delete

`delete` first previews a deletion. `forget` performs a guarded Memory deletion. It is the Memory-specific form of `delete --force --if-revision`. A Memory with a live Link cannot be deleted; the Link must be removed first. Keep `policy-a` and `policy-b` for the next section.

```sh
bd delete policy-c
policy_c_revision="$(bd show policy-c --format graph-json | jq -er '.result.revision')"
bd forget policy-c --if-revision "$policy_c_revision"
```

## 2. Connect Beads with Links

The examples here use Related Links and Dependency Links. `preview-related` is a placeholder Link Type for this preview. The `bd link` commands select Links, so the examples use short Link IDs.

### Related Links

The first example connects `policy-a` to `policy-b` with a Related Link. Its optional `note` property is checked against the Link Type; metadata is separate. Editing this Link advances its source Memory's saved state but does not change the target Memory.

```sh
bd link create --source policy-a --target policy-b --link-type preview-related --id policy-related-b --properties '{"note":"supporting policy"}' --metadata '{"origin":"draft"}'
```

Read the Link and the Links leaving `policy-a`:

```sh
bd link show policy-related-b --format graph-json
bd link list --source policy-a
bd link count --source policy-a
bd link versions policy-related-b --json
```

Update its properties and metadata:

```sh
initial_related_revision="$(bd link show policy-related-b --format graph-json | jq -er '.result.revision')"
bd link update policy-related-b --properties '{"note":"reviewed relationship"}'
bd link update policy-related-b --replace-properties '{"note":"final relationship"}'
bd link update policy-related-b --patch '[{"op":"replace","path":"/note","value":"qualified relationship"}]'
```

```sh
bd link update policy-related-b --set-metadata origin=reviewed
bd link show policy-related-b --format graph-json
current_related_revision="$(bd link show policy-related-b --format graph-json | jq -er '.result.revision')"
bd link compare policy-related-b --from "$initial_related_revision" --to "$current_related_revision" --json
bd graph policy-a --view generic --direction out --depth 2 --json
```

To remove the Link, copy its current revision. The source revision check is optional. `bd delete` removes an unlinked Bead; `bd link delete --force` removes a Link.

```sh
related_revision="$(bd link show policy-related-b --format graph-json | jq -er '.result.revision')"
bd link delete policy-related-b
bd link delete policy-related-b --force --if-revision "$related_revision"
```

### Blocking dependencies

A **Dependency** is a Link between two Issues. In `dep add BLOCKED BLOCKER`, the first Issue waits for the second. The default kind is `blocks`; the examples below use `--type` to name other relationships.
The general `bd link create` form makes the relationship directly. The `bd dep add` form is a shorter way to create a Dependency. An ordered Issue pair can have only one Dependency, regardless of kind, so remove the first one before trying the shorter form.

```sh
bd link create --source blocked-work --target blocker \
  --link-type preview-dependency --type blocks --id typed-blocker --unconditional-source
bd link show typed-blocker --format graph-json
bd dep list blocked-work
bd dep tree blocked-work --direction both
bd dep cycles
bd graph blocked-work
bd graph check --json
```

Remove that Link by ID with its current revision, then recreate the same relationship with `bd dep add`. Here `blocked-work` depends on `blocker`. `bd dep remove` selects a blocking Dependency by its Issue pair.

```sh
typed_blocker_revision="$(bd link show typed-blocker --format graph-json | jq -er '.result.revision')"
bd link delete typed-blocker --force --if-revision "$typed_blocker_revision" --unconditional-source
bd dep add blocked-work blocker
bd dep list blocked-work
bd dep remove blocked-work blocker
```

### Other Issue relationships

The next examples connect Issues with other relationship kinds. A parent-child Link points from child to parent. These Links describe relationships but do not themselves close Issues.

```sh
bd dep add rel-child rel-parent --type parent-child --id links/rel-parent-child
bd link show links/rel-parent-child --format graph-json
```

```sh
bd create 'Duplicate source' --id rel-duplicate-source --type task
bd create 'Duplicate target' --id rel-duplicate-target --type task
bd dep add rel-duplicate-source rel-duplicate-target --type duplicates --id links/rel-duplicates
bd link show links/rel-duplicates --format graph-json
```

```sh
bd create 'Supersede source' --id rel-supersede-source --type task
bd create 'Supersede target' --id rel-supersede-target --type task
bd dep add rel-supersede-source rel-supersede-target --type supersedes --id links/rel-supersedes
bd link show links/rel-supersedes --format graph-json
```

You can also give a Dependency a custom kind. It is still a `preview-dependency` Link, but `tracks` has no built-in blocking or parent-child behavior:

```sh
bd create 'Tracking source' --id rel-tracks-source --type task
bd create 'Tracking target' --id rel-tracks-target --type task
bd link create --source rel-tracks-source --target rel-tracks-target \
  --link-type preview-dependency --type tracks --id links/rel-tracks --unconditional-source
bd link show links/rel-tracks --format graph-json
```

The higher-level `duplicate` and `supersede` commands are unavailable. To remove one of these relationship Links, read its current revision and use guarded `bd link delete --force` with `--unconditional-source`.

## 3. Work with Issues

### Create and find Issues

`bd show ISSUE --json` returns an array of Issue details. `--format graph-json` instead returns the graph record, including its revision. `bd list --json` lists Issues only; `--format records-json` can list both Issues and Memories. `q` creates an Issue and prints its ID. You can also create several Issues from a Markdown file, as shown below.

```sh
bd create 'Review the release' --id beads/review --type task --label release
bd new 'Alias-created task' --id beads/alias-task --type task
bd create 'Track the whole release' --id beads/release-work --type task
bd q 'Quick follow-up'
bd todo add 'Document a test' --description 'Record the expected result'
bd todo list --all
bd statuses --json
bd status
bd stats --json
bd list --json --all
bd show review --json
bd show review --format graph-json
bd search 'release' --json
bd count --by-status --json
bd query 'status=open AND priority<=2' --json
```

Save the IDs printed by `q` and `todo add` if you want to use those Issues later. The next block makes a temporary Markdown file with two Issue headings and passes it to `create --file`. Initial labels can be set during creation, but the `label` and `tag` management commands are unavailable. `status` shows Issue statistics; `status --graph` describes graph capabilities; `statuses` lists possible Issue statuses.

```sh
batch_dir="$(mktemp -d "${TMPDIR:-/tmp}/preview2-batch.XXXXXX")"
batch_file="$batch_dir/issues.md"
cat > "$batch_file" <<'MARKDOWN'
## Inspect the release notes

### Description
Read the published examples.

### Priority
1

## Verify the binary

### Description
Record its version and SHA-256.
MARKDOWN
bd create --file "$batch_file" --json
rm "$batch_file"
rmdir "$batch_dir"
```

### Edit fields and comments

```sh
bd priority review 1
bd assign review alice
bd note review 'Checklist started.'
bd update review --append-notes 'Second note.'
bd comment review 'Ready for review'
bd comments add review 'Reviewer acknowledged' --author alice
bd comments review --json
bd versions review --json
```

Read the Issue's current revision before and after an edit. For one Issue, `--if-revision` can reject an edit if the Issue changed since your read. `--append-notes` adds text, `--notes ... --force` replaces nonempty notes, and `--clear-notes` erases them. `bd edit review` opens your configured text editor; saving without a change should add no version. Comments have their own feed and do not add an Issue version.

```sh
bd show review --format graph-json
review_revision="$(bd show review --format graph-json | jq -er '.result.revision')"
bd update review --notes 'Final review note.' --force --if-revision "$review_revision"
bd show review --format graph-json
review_revision="$(bd show review --format graph-json | jq -er '.result.revision')"
bd update review --clear-notes --if-revision "$review_revision"
bd show review --format graph-json
```

### Lifecycle, readiness, and coordination

Add a blocking dependency and inspect `blocked`. After you close the prerequisite, `blocked-work` may appear in `ready`. Read the Issue again after each lifecycle change.

```sh
bd dep add blocked-work blocker
bd blocked --json
bd ready --type task --limit 10 --json
bd close blocker --reason 'Prerequisite complete'
bd ready --type task --limit 10 --json
bd reopen blocker --reason 'Recheck prerequisite'
bd defer alias-task --until tomorrow
bd undefer alias-task
```

`close --suggest-next` and `--claim-next` apply to one Issue at a time; use a separate pair of Issues if you try them. Replace `ID` in `bd todo done ID` with an actual ID returned by `todo add`. Closing a prerequisite can unblock other Issues. Closing an already closed Issue again makes no change. A deferred Issue can become ready when its date passes.

```sh
bd create 'Claim-only walkthrough' --id beads/ready-claim-example --type task --label claim-demo
bd ready --claim --label claim-demo --actor demo --json
bd show ready-claim-example --json
bd todo list --all
todo_id="$(bd todo list --all --json | jq -er '[.[] | select(.title=="Document a test") | .id] | first')"
bd todo done "$todo_id" --reason 'Documented and verified'
```

```sh
bd create 'Lease example' --id beads/lease-example --type task
bd update lease-example --claim --actor demo
bd heartbeat lease-example --actor demo
bd hb lease-example --actor demo
bd unclaim lease-example --actor demo
bd reclaim --older-than 0s --id lease-example --json
bd state review patrol --json
bd state list review --json
bd merge-slot create
bd merge-slot check --json
bd merge-slot acquire --holder demo
bd merge-slot check --json
bd merge-slot release --holder demo
```

The `reclaim` command above should find nothing because the lease was just released. To test real recovery, you need a separate Issue with an **expired** lease. `ready --claim` claims one ready Issue in a single operation. `heartbeat` renews a lease without making an Issue version; `reclaim` changes an expired lease. `state` reads existing state labels and returns `null` when a requested state is unset. `set-state` is unavailable. Create the merge slot only once in a workspace.

### Parent and child; epic progress


You can make a parent-child relationship in two ways: create both Issues and connect them with `dep add --type parent-child`, or create a child directly with `create --parent`. The first example uses two separate Issues. `q --parent` is unavailable.

```sh
bd create 'Release epic' --id beads/epic-a --type epic
bd create 'Publish package' --id beads/epic-child-a --type task
bd create 'Similar dotted ID, no relationship' --id beads/epic-a.lookalike --type task
bd dep add epic-child-a epic-a --type parent-child
bd children epic-a
bd children epic-a --json
bd epic status
bd epic status --eligible-only
bd epic close-eligible --dry-run --reason 'All child work complete' --actor demo
```

`children` should show the actual child, even after it is closed. It should not treat the Issue with a similar dotted ID as a child. Close the real child, then inspect `children` and `epic status` again. Use `epic close-eligible --dry-run` to preview which epics can close. If `epic-a` is eligible, run the real command and read it again. An assigned, pinned, or blocked epic may still refuse to close. The command handles eligible epics one by one, so if any fail, check which ones succeeded before retrying.

```sh
bd close epic-child-a --reason 'Package published'
bd children epic-a --json
bd epic status --eligible-only
bd epic close-eligible --dry-run --reason 'All child work complete' --actor demo
bd epic close-eligible --reason 'All child work complete' --actor demo
bd show epic-a --json
bd versions epic-a --json
```

Now create a new epic and use `create --parent` to make its child. The command chooses a dotted child ID, copies the parent's labels unless told otherwise, and creates the parent-child Link in the same operation. Save the child and Link IDs returned in JSON; use those IDs for later `show`, `versions`, `history`, and `bd link list` commands. If the parent does not exist, the command should fail without creating either record.

```sh
bd create 'Parent for atomic child' --id beads/epic-b --type epic --labels inherited --json
bd create 'Atomic child' --parent epic-b --type task --labels own --json
bd children epic-b --json
bd link list --incident-to epic-b --direction in
bd show epic-b --format graph-json
bd versions epic-b --json
```

The parent-child Link points **from the child to the parent**, so `bd link list --incident-to epic-b --direction in` finds it at the parent. Use the returned child ID to see the outgoing Link. A second child made with `--no-inherit-labels` keeps only the labels you give it. `create --parent` cannot be combined with an explicit `--id`. To delete an Issue, first remove every live Link connected to it with `bd link delete --force`, then preview `bd delete ISSUE`; the real Bead deletion needs `--force` and a current revision or an explicit unconditional guard.

## 4. Optional Claude Code setup

If you use Claude Code with your own project, install its hook and check it. The hook's `stop` command is called by Claude with transcript input; do not invoke it as a standalone manual test.

```sh
bd setup claude --project
bd setup claude --project --check
```

## 5. Optional BDP Read exercise

BDP serves graph records over HTTP. It needs a Dolt shared server, so use a **separate new workspace** for this exercise. The earlier CLI workspace uses embedded Dolt and cannot serve BDP. You need `dolt` 2.1.8 and `curl` installed. Keep each service running in its own terminal. The commands below use local ports 13395 and 18795; choose unused ports if either is occupied.

In **terminal A**, create a fresh directory and start Dolt:

```sh
mkdir "$HOME/beads-preview2-bdp"
mkdir "$HOME/beads-preview2-bdp/server-data" "$HOME/beads-preview2-bdp/workspace"
cd "$HOME/beads-preview2-bdp"
dolt sql-server -H 127.0.0.1 -P 13395 --data-dir "$PWD/server-data"
```

In **terminal B**, enter that workspace, initialize it against the running Dolt server, and create one Memory. Start the read-only BDP service in this terminal:

```sh
cd "$HOME/beads-preview2-bdp/workspace"
bd init --graph-mode link --scope-url 'http://127.0.0.1:18795/graph/' \
  --server --external --server-host 127.0.0.1 --server-port 13395 --server-user root
bd remember 'Review before publishing.' --id policy-bdp --title 'BDP policy'
bd serve --readonly --addr 127.0.0.1:18795
```

In **terminal C**, request the service description and the Memory. The service description lists its Read capability and collection URLs:

```sh
curl --fail-with-body http://127.0.0.1:18795/graph/bdp.json
curl --fail-with-body http://127.0.0.1:18795/graph/beads/policy-bdp
curl --fail-with-body 'http://127.0.0.1:18795/graph/beads/?limit=1'
```

A Bead collection response has `items` and possibly a `next` URL. Follow each `next` URL to read the whole collection; the first page may not contain everything. BDP in this preview is read-only. It does not offer HTTP writes or History.
When you finish, press Ctrl-C in terminals B and A to stop the BDP and Dolt services.

## Known Limitations

**Preview 2 is experimental: data in a preview workspace may be lost, and preview-to-preview compatibility is not supported. There is currently no supported way to upgrade a workspace from one preview to the next. Use data you can recreate, and do not rely on this preview as your only copy of important work.** Start each preview in a new workspace.

The Type catalog is fixed in this preview. Run `bd types` to see the installed Bead and Link Types; installing or editing Types is unavailable. Other unfinished areas include:

- Import and export for migration, plus general backup and recovery.
- HTTP writes and BDP History. The BDP service above provides Read only.

A stale revision or invalid Type causes a refusal. After a refusal, check that no data changed before continuing. This preview focuses on Issues, Memories, Links, Dependencies, and the commands used with them. Other Beads commands may refuse.

### Commands believed complete for the uses in this guide

This list covers only the forms described above. It does not promise every option from ordinary Beads.

- Issue work: `assign`, `close`, `defer`, `undefer`, `edit`, `note`, `priority`, `reopen`, `comment`, `count`, `query`, `search`, `statuses`, and `todo`.
- Coordination: `heartbeat`, `reclaim`, `unclaim`, and `merge-slot`.
- Graph setup and reads: `init` for a new workspace, `compare`, `bd link create/list/show/count/versions/compare/delete` for the examples above, `types`, and `versions`.

### Commands with known limitations

| Command | What does not work in Preview 2 |
| --- | --- |
| `create` | `--deps` cannot add ordinary dependencies during creation, and dependency sections in a Markdown input file refuse. |
| `create` | Issue description and Memory body text cannot come from stdin or `--body-file`; use inline text. `--dry-run` is unavailable. `create --file` does accept a Markdown file of Issues. |
| `create --parent` | An explicit `--id` cannot be combined with `--parent`. The parent and new child must be Issues; the parent need not be an epic. |
| `update` | Multi-record updates are unavailable. |
| `update` | A native Dependency Link has no editable properties in this preview; its kind is fixed when created, and metadata editing is unavailable. Informational Links can be updated with `bd link update links/ID`, including metadata. |
| `update` | Issue labels, status, and category cannot be edited through `update`. Use the supported lifecycle commands for status changes. Issue text fields cannot be read from file/stdin flags; use inline text. JSON property input has its own `@FILE` and `@-` forms. |
| `delete` | A Bead with any live Link cannot be deleted. Use `bd link delete --force` on each Link first. Deletion does not cascade to Links. |
| `delete` | Batch deletion and `--from-file` are unavailable. To actually delete one Bead, use `--force` with either `--if-revision TOKEN` or `--unconditional`; plain `bd delete ID` only previews. |
| `show` | Multiple Beads or Links cannot be read in one call. Current Issue `--json` uses the ordinary Issue array; Memory, Link, and older-version reads use a graph record. |
| `list` | There is no Link/Bead/both selector: `list` covers Beads, not Links. `--json` covers Issues only; use `--format records-json` for multiple Bead Types and `bd link list --incident-to BEAD` for incident Links. |
| `list` | Tree output, cursor paging, and some ordinary filters and sort orders are unavailable. Output is flat and bounded. |
| `dep` | Pair removal and `list`, `tree`, and `cycles` views cover blocking Dependencies only; they do not cover parent-child, duplicates, or supersedes relationships. |
| `dep` | `list/tree/cycles --json`, cross-scope Dependencies, and bulk operations are unavailable. |
| `link` | Remote and version-pinned endpoints are unavailable. Both endpoints must be unpinned Beads in the same Scope. |
| `link create` | Only installed Link Types work. With `--link-type preview-dependency`, `--type KIND` sets the Dependency kind; omitting it means `blocks`. Dependency Links require two local Issues, one Link per ordered Issue pair, an explicit source guard, and no custom properties or metadata. |
| `link update` | Informational Link properties and metadata can be edited. Native Dependency Links refuse property and metadata updates; their kind cannot be changed. |
| `link list`, `link count` | The current Link collection is bounded and has no cursor paging. A workspace above the read limit is refused rather than partially counted or listed. |
| `link` | Creating a duplicate or supersede Link does not close an Issue. |
| `graph` | `graph ISSUE --json` is unavailable; use `graph --view generic --json` for bounded machine-readable traversal. |
| `children` | Arbitrary descendants and `--pretty` output are unavailable; it lists immediate children. |
| `epic` | Other ordinary epic operations are unavailable; this preview supports `status` and `close-eligible`. |
| `q` | `q --parent` is unavailable. |
| `comments` | Existing comments cannot be edited or deleted; reading and adding are available. |
| `history` | The ordinary Issue event log is unavailable here; `history` lists saved record versions. |
| `state` | Setting state labels is unavailable; `state` reads them. |
| `ready`, `blocked`, `status` | Some ordinary options are unavailable. In particular, `blocked --parent` is unavailable. `status --graph` describes graph capabilities, not Issue statuses. |
| `memories` | `memories --json` is unavailable. |
| `forget` | A Memory with a live Link cannot be deleted until the Link is removed. |

### Commands outside this preview

| Area | Unavailable commands or forms |
| --- | --- |
| High-level Issue and state operations | `duplicate`, `supersede`, `gate`, `set-state`, `label`, `tag` |
| Other graph and analysis commands | `create-form`, `promote`, `provenance`, `diff`, `find-duplicates`, `duplicates`, `lint`, `stale`, `swarm` |
| Link query | `bd link query` is unavailable. |
| Optional version display inside a graph workspace | `bd version` refuses inside a graph workspace. It is not needed to use the commands in this guide. |

`new`, `hb`, and `stats` are aliases for `create`, `heartbeat`, and `status`. Each alias has the same limits as its main command.
