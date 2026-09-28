# A plan, its context, and the work it describes

This walkthrough puts Issues, Memories and Links in one disposable workspace.
Use the installed preview `bd` and `jq`. The new claim slice is still awaiting
its exact-source Linux qualification; this is a demo script, not a completion
claim for the full graph or Memory proposals.

## 1. Start with an ordinary initialization

Run in a shell where `bd` selects the preview binary. Embedded storage is the
default. To include the HTTP section, choose the shared-server option **before**
initializing; it requires an already-running disposable Dolt server. Do not use
a production database or initialize multiple databases concurrently.

```sh
work=$(mktemp -d)
cd "$work"
scope='http://127.0.0.1:8765/demo/'
init_backend=()
# For HTTP, replace the empty array above with this (adjust the Dolt port):
# init_backend=(--server --external --server-host 127.0.0.1 --server-port 54550 --server-user root --database "demo_$(date +%s)")
bd init --graph-mode link --scope-url "$scope" --prefix demo \
  --non-interactive --skip-hooks --skip-agents "${init_backend[@]}"
```

There is no schema script or hidden seed step. The remaining commands each open
the workspace in a new process.

## 2. Connect the plan to real work

```sh
bd create 'Ship the change' --id beads/work --priority 2
bd create 'Review the change' --id beads/review --priority 1
bd dep add beads/work beads/review --json
bd remember 'Ship only after the review is complete.' --id beads/plan --title 'Release plan'
bd remember 'The reviewer needs the compatibility notes.' --id beads/notes --title 'Review context'
plan_revision=$(bd show beads/plan --json | jq -r '.result.revision')
bd link beads/plan beads/work --resource-type "${scope}types/preview-related-v2" \
  --id links/work-context --properties '{"note":"This plan explains the work"}' \
  --if-source-revision "$plan_revision" --json
plan_revision=$(bd show beads/plan --json | jq -r '.result.revision')
bd link beads/plan beads/notes --resource-type "${scope}types/preview-related-v2" \
  --id links/review-context --properties '{"note":"Read this context too"}' \
  --if-source-revision "$plan_revision" --json
bd show beads/plan --json
bd show beads/work --json
bd ready --json
```

The plan owns two informational Links: one to an Issue and one to another Memory.
The work Issue owns its blocking Dependency. `show` displays those complete
records; `ready` lists the review task because it blocks shipping. These are
stored relationships, not links inferred from words in a Memory body.

## 3. Claim the work, then finish its prerequisite

```sh
before_claim=$(bd show beads/work --json | jq -r '.result.version')
bd update beads/work --claim --actor alice --json
claimed=$(bd show beads/work --json | jq -r '.result.version')
bd list --format records-json --assignee alice --all
bd update beads/work --claim --actor alice --json
bd close beads/review --reason 'Review complete' --actor reviewer --json
bd close beads/work --reason 'Shipped' --actor alice --json
closed=$(bd show beads/work --json | jq -r '.result.version')
bd reopen beads/work --reason 'Follow-up needed' --actor alice --json
bd show beads/work --json
```

Claiming sets Alice as the assignee and moves work to `in_progress`. It does not
remove blockers or require the work to be ready. The repeated claim reports
`changed:false`. Claim grants the existing five-minute lease; repeating it does
**not** renew that lease. Graph heartbeat/reclaim and mixed claim-plus-edit
commands are not exposed. Closing the prerequisite allows shipping to close;
reopening makes that Issue open again while preserving its Dependency.

## 4. Read exactly what we saved

```sh
bd show beads/work --version "$before_claim" --json
bd show beads/work --version "$claimed" --json
bd show beads/work --version "$closed" --json
bd compare beads/work --from "$claimed" --to "$closed" --json
bd show beads/plan --json
bd recall beads/notes
```

Current reads show the reopened Issue. Saved-version reads show the earlier
complete Issue and owned-Link state; comparison explains the differences.
The Memory plan and linked notes are still available. This demonstrates exact
retained versions, not a full public History timeline or traversal API.

## 5. Read the same graph over BDP (shared-server setup only)

In the workspace, start the read-only HTTP service in a separate terminal:

```sh
bd serve --readonly --graph-mode link --addr 127.0.0.1:8765
```

Then read discovery, the plan, and its incident Links:

```sh
curl --fail http://127.0.0.1:8765/demo/bdp.json | jq
curl --fail http://127.0.0.1:8765/demo/beads/plan | jq
curl --fail 'http://127.0.0.1:8765/demo/beads/plan?view=links&direction=both' | jq
```

These are BDP Read responses from the same workspace. CLI commands authored the
data; BDP HTTP writes and embedded HTTP serving are unavailable. The independent
public-client acceptance harness exercises discovery and reads beyond these
simple `curl` examples. Stop the demo listener with Ctrl-C when finished.
