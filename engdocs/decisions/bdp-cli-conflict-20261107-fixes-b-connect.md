# Decision record: commands that refused over bd connect and could work (native fixes, group B)

Date: 2026-11-07
Decided by: Donna Box
Status: decided; branch `trish/fixes-b-connect-20261107` carries the change

Scope: commands that refused over `bd connect`, on v0 and on BDP alike. Rows B1 to B8
also move the direct route onto the same role, so a direct workspace takes a different
code path to the same write.

This branch stands on rows E2, B15 and B16
([bdp-cli-conflict-20261107-fixes-b15-b16-raw-reads.md](bdp-cli-conflict-20261107-fixes-b15-b16-raw-reads.md)).

| Row | Command | Before | After |
|---|---|---|---|
| B1 | `bd priority` | Over connect exit 1: `UpdateIssue` not supported | Lifecycle role on every route |
| B2 | `bd tag` | Over connect exit 1: `AddLabel` not supported | Lifecycle role patch on every route |
| B3 | `bd note` | Over connect exit 1: `UpdateIssue` not supported | Lifecycle role append-notes on every route |
| B4 | `bd q`, with or without `--parent` | Over connect exit 1: `CreateIssue` or `GetNextChildID` not supported | Lifecycle create; the role reserves the child id and inherits labels in one write, on every route |
| B5 | `bd todo add` | Over connect exit 1: `CreateIssue` not supported | Lifecycle create on every route |
| B6 | `bd duplicate`, `bd supersede` | Over connect exit 1: `AddDependency` not supported | Dependency editor on every route |
| B7 | `bd dep relate`, `bd dep unrelate` | Over connect exit 1 | Dependency editor; relate writes both edges in one call. One error's words move from `a -> b` to `a <-> b` |
| B8 | `bd defer`, `bd undefer` | Over connect prints "Error deferring ...: `UpdateIssue` not supported", **exits 0**, writes nothing | Lifecycle role on every route, so the write lands. `--reason` is an append operation, not a read-merge-write, direct too. The exit code on a failed id is row C1's, not this row's |
| B9 | `bd unclaim`, all flags | Over connect exit 1 | Releaser role; the `--reason` comment goes through the Commenter role |
| B10 | `bd close --claim-next` | Over connect exit 1, nothing closes | Closes, then claims as a **second request**. A failed claim prints an error and exits nonzero with the closes already committed; the one-transaction promise is lost |
| B11 | `bd batch` | Over connect exit 1: `RunInTransaction` not supported | One `applyBatch` request. A `dep remove` line, which has no item kind, refuses by line number before anything is sent |
| B12 | `bd import` | Over connect exit 1 | `applyBatch` or `batchCreateIssues` by chunk. Creates only, never upserts. A row with comments or stored timestamps refuses, naming the row; skipping conflicting rows refuses |
| B13 | `bd create --file` | Over connect exit 1 | Drops the caller's history label on a remote store, tries `batchCreateIssues`, falls back to `applyBatch` when Owner is set, then reads back the rows (several requests, one transaction) |
| B14 | `bd events tail`, `bd events export`, `bd events prune` | Over connect exit 1, and the local "journal disabled" note, which says nothing true about the server | Reads `listEvents`, without the note. Carries the journal row's `actor`, which `internal/httpclient/journal.go` dropped, for library callers too. `bd events prune` refuses in words |

## CLI surface

No flag or help text moves. B7 changes one error's words. Commands that refused now
print what their `--json` always printed on a local workspace.

## Conflicts to expect

B8 and row C1 edit the same lines of `cmd/bd/defer.go` and `cmd/bd/undefer.go`. Taking
both means keeping B8's write and C1's `hasError`.

## Ledger and docs

`engdocs/design/http-divergence-ledger.md`: `F-close` says what `--claim-next` does now.
`README.md`'s "one refusal to know up front" paragraph was about the same flag, and is
rewritten.

## What pins it

`TestE2E_ConnectCommandParity` (the real `bd`, both protocols, every row, including a
claim that fails after its close), `TestServedJournalCarriesTheActor`, and the existing
direct-route tests for each command, which now run through the role.
