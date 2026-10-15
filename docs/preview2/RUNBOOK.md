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

This uses embedded Dolt storage. To try BDP, use the separate shared-server workspace in [Optional BDP Read exercise](#5-optional-bdp-read-exercise) *instead* of running the commands above. BDP requires a shared-server workspace.

## Beads, Links, and Types

The graph contains **Beads** (nodes) and **Links** (connections between Beads). Each has a Type. A Type identifies the sort of record and defines which properties and endpoints are allowed. Use `--type` to choose a Type when you create, search for, or filter a Bead or Link.

Run `bd types` to see the names accepted by `--type`. In a new Preview 2 workspace, its default display is:

Without `--type`, `bd create` makes a `task` Issue. `bd link create` requires `--type`; the shorter `bd link SOURCE TARGET` makes a `blocks` Dependency.

```text
Bead Types (14):
  preview-memory            Memory Preview
  task                      Task Issue
  bug                       Bug Issue
  feature                   Feature Issue
  chore                     Chore Issue
  epic                      Epic Issue
  decision                  Decision Issue
  spike                     Spike Issue
  story                     Story Issue
  milestone                 Milestone Issue
  message                   Message Issue
  molecule                  Molecule Issue
  gate                      Gate Issue
  event                     Event Issue

Link Types (22):
  preview-related-memory    Related Memory Preview
  example-cites             Example Cites
  example-follows           Example Follows
  blocks                    Blocking Dependency
  parent-child              Parent/Child Dependency
  conditional-blocks        Conditional Blocks Dependency
  waits-for                 Waits For Dependency
  related                   Related Dependency
  discovered-from           Discovered From Dependency
  replies-to                Replies To Dependency
  relates-to                Relates To Dependency
  duplicates                Duplicates Dependency
  supersedes                Supersedes Dependency
  authored-by               Authored By Dependency
  assigned-to               Assigned To Dependency
  approved-by               Approved By Dependency
  attests                   Attests Dependency
  tracks                    Tracks Dependency
  until                     Until Dependency
  caused-by                 Caused By Dependency
  validates                 Validates Dependency
  delegated-from            Delegated From Dependency
```
Bead types already require a registered type when you create a new bead.  Links don't have that requirement in the shipping product. We will soon close that hold and require both bead AND link types to be registered.

We recently changed the design for the type interface and this preview has only implemented some of the changes.  If you see the names `preview-issue` or `preview-dependency` you found one of the seams. Preview 3 is focused on correctness, consistency, and completeness of the CLI, so those SHOULD disappear shortly.

Look for a future blog post going over all the things related to bead and link types. For now, this should be all we need to know to use the product.

## Working with Beads and Links

Beads and Links have much in common:

- Each has a unique ID. IDs can contain slashes, so you can organize them into paths.
- Each has properties checked against its Type. Beads and Links also have separate metadata. Dependency Link metadata cannot be set in this preview; `preview-related-memory` and the example Link Types support it.
- Each has saved versions that you can inspect and compare.

The CLI uses top-level commands for Beads and the `bd link` namespace for Links:

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

| Shortcut | Closest general command | Difference |
| --- | --- | --- |
| `bd remember` | `bd create --type preview-memory` or `bd update` | `remember` takes the Memory body as a positional argument. |
| `bd recall` | `bd show` | `recall` prints the body rather than the full record. |
| `bd forget` | `bd delete --force` | `forget` deletes a Memory immediately. |
| `bd memories` | `bd list --type preview-memory` | `memories` also searches Memory titles and bodies. |
| `bd dep add` | `bd link create` | `dep add` accepts an Issue pair and a Dependency kind. |
| `bd dep remove` | `bd link delete --force` | `dep remove` selects a blocking Dependency by Issue pair; `link delete` selects a Link ID. |
| `bd history` | `bd versions` | Both list saved versions. |
| `bd new`, `bd q` | `bd create` | Short Issue creation forms. |

`bd link query` is not available in this preview.


## 1. Beads, Memories and Issues

A **Memory** is a Bead that stores reusable information. Its properties are a title and a body; the body is Markdown text.

The next four examples create Memories about release policies. They show the general `create` form, shorter field flags, and the Memory-specific `remember` command.

### Create

Create the first Memory by supplying its Type-checked properties and separate metadata as JSON:

```sh
bd create --type preview-memory --id policy-a \
  --properties '{"title":"Policy A","body":"Review *before* publishing."}' \
  --metadata '{"origin":"manual test"}'
```

Create the second with individual field flags:

```sh
bd create --type preview-memory --id policy-b --title 'Policy B' --body 'Retain the review evidence.'
```

The positional argument can also supply the title:

```sh
bd create 'Policy C' --type preview-memory --id policy-c --body 'Review policy changes before release.'
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

The general `list` command can include Issues and Memories. Narrow it to Memories with `--type`, or use the Memory-specific `memories` command:

```sh
bd list --format records-json
bd list --format records-json --type preview-memory
bd memories --all
```

`bd list --json` shows Issues only. Use `--format records-json` for both Bead Types. On `bd list`, `--all` includes closed and pinned Issues and removes the default row limit within Preview 2's bounds; it does **not** add Links. Use `bd link list` for Links.

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

`update --properties` changes only the named properties. `--replace-properties` replaces the entire writable properties object, so it must include every required field. `--patch` applies an ordered JSON Patch-style list of `add`, `replace`, or `remove` operations to properties.

Here are all three ways to update properties:

```sh
bd update policy-a --properties '{"body":"Review and qualify before publishing."}'
bd update policy-a --replace-properties '{"title":"Policy A","body":"Reviewed policy."}'
bd update policy-a --patch '[{"op":"replace","path":"/body","value":"Qualified policy."}]'
```

`remember --update ID` is a Memory-specific shortcut. A positional argument changes the body; `--title` changes the title. You can supply either or both. Omitted fields stay unchanged.

```sh
bd remember 'Review before publishing and after any major change.' --update policy-a
bd remember --update policy-a --title 'Policy A, revised'
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

`delete` first previews a deletion. `forget` performs a guarded Memory deletion. With `--if-revision`, it is the Memory-specific form of `delete --force --if-revision`. A Memory with a live Link cannot be deleted; the Link must be removed first. Keep `policy-a` and `policy-b` for the next section.

```sh
bd delete policy-c
policy_c_revision="$(bd show policy-c --format graph-json | jq -er '.result.revision')"
bd forget policy-c --if-revision "$policy_c_revision"
```

## 2. Connect Beads with Links

The examples here create and manage Links. `bd link` is the general namespace for Links. `bd dep` is a shortcut for Dependencies between Issues.

### Related Links

The first example connects `policy-a` to `policy-b` with a `preview-related-memory` Link. This is different from the `related` Dependency between Issues. Its optional `note` property is checked against the Link Type; metadata is separate. Editing this Link advances its source Memory's saved state but does not change the target Memory.

```sh
bd link create --source policy-a --target policy-b --type preview-related-memory --id policy-related-b --properties '{"note":"supporting policy"}' --metadata '{"origin":"draft"}'
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

A **Dependency** is a Link between two Issues. In `bd dep add BLOCKED BLOCKER`, the first Issue waits for the second when the Type is `blocks`. `bd link create` is the general way to make the same Link; `bd dep add` is a shorter form. An ordered Issue pair can have only one Dependency, regardless of Type, so remove the first one before trying the shorter form.

The built-in Dependency choices shown by `bd types` are `blocks`, `parent-child`, `conditional-blocks`, `waits-for`, `related`, `discovered-from`, `replies-to`, `relates-to`, `duplicates`, `supersedes`, `authored-by`, `assigned-to`, `approved-by`, `attests`, `tracks`, `until`, `caused-by`, `validates`, and `delegated-from`. They all connect Issues, but have different meanings. `blocks`, `parent-child`, `conditional-blocks`, and `waits-for` can affect readiness; `parent-child` also defines the hierarchy. The other choices do not block work merely by existing.

```sh
bd link create --source blocked-work --target blocker \
  --type blocks --id typed-blocker --unconditional-source
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

The higher-level `duplicate` and `supersede` commands are unavailable. To remove one of these relationship Links, read its current revision and use guarded `bd link delete --force` with `--unconditional-source`.

## 3. More ways to work with Issues

This section follows the Link examples because its readiness exercise uses the `blocked-work` and `blocker` Issues from above.

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

Run `bd types` for usable CLI choices. Installing or editing graph Type descriptors is unavailable in this preview. Other unfinished areas include:

- Import and export for migration, plus general backup and recovery.
- HTTP writes and BDP History. The BDP service above provides Read only.

A stale revision or invalid Type causes a refusal. After a refusal, check that no data changed before continuing. This preview focuses on Issues, Memories, Links, Dependencies, and the commands used with them. Other Beads commands may refuse.

### Commands believed complete for the uses in this guide

This list covers only the forms described above. It does not promise every option from ordinary Beads.


| Area | Commands covered here |
| --- | --- |
| Issue work | `bd assign`, `bd close`, `bd defer`, `bd undefer`, `bd edit`, `bd note`, `bd priority`, `bd reopen`, `bd comment`, `bd count`, `bd query`, `bd search`, `bd statuses`, `bd todo` |
| Coordination | `bd heartbeat`, `bd reclaim`, `bd unclaim`, `bd merge-slot` |
| Graph setup and reads | `bd init` in a new workspace, `bd types`, `bd versions`, `bd compare` |
| Link operations | `bd link create`, `list`, `show`, `count`, `versions`, `compare`, and guarded `delete` as shown above |

### Commands with known limitations

| Command | What does not work in Preview 2 |
| --- | --- |
| `create` | `--deps` cannot add ordinary dependencies during creation, and dependency sections in a Markdown input file refuse. |
| `create` | In `bd create`, Issue description and Memory body text cannot come from stdin or `--body-file`; use inline text. The Memory-specific `bd remember` does accept `--body-file` or `--stdin`. `create --dry-run` is unavailable; `create --file` accepts a Markdown file of Issues. |
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
| `link create` | Use `--type` with a displayed Link choice. For Dependencies, `--type blocks` and other built-in Dependency choices work; an unrecognized choice refuses. Without `--type`, the `bd link SOURCE TARGET` shortcut creates a blocking Dependency, but `bd link create` requires `--type`. Dependency Links require two local Issues, allow only one Link per ordered Issue pair, require an explicit source guard on `link create`, and accept no custom properties or metadata. |
| `dep add` | Unlike `bd link create`, its legacy `--type` option still accepts arbitrary nonempty Dependency labels up to 32 bytes. Only built-in choices have their defined behavior. This guide uses built-in choices. |
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
