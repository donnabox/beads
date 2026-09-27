# Issue reopen preview

The disposable graph workspace can use the existing Issue reopening operation
through one explicit canonical selector:

```sh
bd reopen beads/plan --reason 'Need to revisit the acceptance criteria' --json
```

The same-Scope canonical URL returned by `show` also works. This preview accepts
one Issue per invocation. It does not resolve short IDs, aliases, other Scopes,
implicit last-touched IDs or batches. Ordinary legacy workspaces retain their
existing reopen route.

Reopening moves a done-category Issue to open, clears its closure fields and
recomputes dependency readiness through the existing Issue writer. That writer
also clears `defer_until`; this preview cannot set that field, so its nonempty
clearing behavior is inherited and not exercised by the installed story. The JSON
preview result is `{issue, changed}` inside the existing experimental wrapper.
The returned Issue includes its complete owned blocking Dependencies. An Issue
outside the done category is unchanged: no new saved version, event or graph
revision is created. Human output reports the unchanged status accurately.

A changed reopen saves exactly one new retained Issue version and binds it to
one graph version in the same transaction as the domain write. Existing exact
`show ID --version TOKEN` reads continue to resolve earlier closed and open
states, including their owned Links. The reason is domain event/audit text;
it is not a new Issue property or a structured comment. This slice does not
expose those events through a graph History command.

A useful sequence is to create a blocking Dependency from a dependent Issue to
a prerequisite, close the prerequisite, and observe the dependent in `ready`.
Reopen the prerequisite and the dependent is blocked again. Closing it with a
new reason then records a new closed state without rewriting the previous
closed state. This also follows the amendment path discussed in upstream
[first-close-wins work](https://github.com/gastownhall/beads/pull/6296); it does
not incorporate that PR's differing-reason refusal into graph close.

## Boundaries and integration

The provisional guard policy matches the existing graph close command: reopen
has no new mandatory revision flag. Broader guard and claim-fence contracts
remain under review. Readonly invocations, migration freezes, mismatched
workspace authority, unsupported options and wrong Resource kinds refuse.
No success is emitted before commit. A lost commit response reports the existing
unknown-outcome error with no success record; the adapter does not retry it.
As with graph close, a no-op still completes the existing transaction, so a
lost COMMIT reply can report unknown outcome even when no state was written.
Integrity failures currently share the `graph_not_initialized` diagnostic code
with initialization failures; inspect the accompanying message before deciding
how to recover. Reopen reasons have no additional preview byte budget and
inherit the existing domain storage limits and atomic failure behavior.
These are shared workflow limitations, not new final error/input contracts.

This route reuses `issueops.ExecuteReopen`. At this pinned base that writer
journals changes but does not call the retained Issue recorder, so the adapter
calls that recorder once. At Jim's integration target, recheck recorder ownership
and remove the adapter call if the writer has assumed it. Keep the mapping and
exactly-one-version tests. Jim's exact target is still pending.

[Plane-consistent transactions](https://github.com/gastownhall/beads/pull/6015)
and [claim-fence work](https://github.com/gastownhall/beads/pull/4697) remain
separately owned integration work. This preview stays within the admitted durable
Issue plane and does not introduce a second Issue writer, fence column or schema.
Memory restore, wider Issue workflows, native atomic commit stamps and complete
public History remain unavailable.

## Evidence

The installed harness initializes fresh workspaces normally and authors all
Issues and Links through the CLI:

```sh
python3 scripts/graph-issue-reopen-smoke.py --bd /absolute/path/to/bd \
  --backend both --server-port 3307 --output-dir /absolute/new-receipts
```

Qualification must include both embedded and ordinary shared-server Dolt, exact
old-state reads, unchanged owned relationships and failure controls. Storage
tests additionally cover rollback, no-op/version accounting, custom status
categories, real server transaction overlap and a dropped COMMIT response.
Different-database provisioning remains serialized on Dolt 2.1.8. The existence
of these tests or this harness is not itself a qualification receipt; current
results and limits belong in the fork's delivery plan.

The ordinary preview leaves the optional journal mode disabled. Tests fingerprint
journal/lease tables for rollback and no-op preservation, but do not claim a
positive journal-feed or lease workflow. Custom done/non-done categories are
covered through storage configuration tests; the installed CLI has no new
status-configuration route.
