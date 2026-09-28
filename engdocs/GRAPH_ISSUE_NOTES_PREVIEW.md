# Append-only Issue notes in the graph preview

A normally initialized graph workspace accepts the existing `--append-notes`
flag on `bd update` for canonical Issue Beads. Supply the observed graph revision
or explicitly choose `--unconditional`:

```sh
bd show beads/task --json
bd update beads/task --append-notes 'Confirmed the reproduction on both backends.' \
  --if-revision OBSERVED_VERSION --json
bd show beads/task --json
bd show beads/task --version OBSERVED_VERSION --json
```

Notes remain an Issue text property. They are not separate Memory Beads,
comments, or ordered History entries. The authoring proposal remains
[upstream #6703](https://github.com/gastownhall/beads/issues/6703); this adapter
is a reversible preview of the existing Issue operation.

## What append means

The shared Issue writer appends the supplied text inside the checked transaction.
When notes already contain text, it inserts one newline before the new text.
Unicode, whitespace, CRLF and literal `-` are preserved. There is no file or
stdin interpretation or new per-fragment byte cap. A changed request that
supplies `--append-notes` must leave its entire workspace within the existing `currentReadBytes` acquisition budget
(currently 16 MiB). The check runs after the mutation and retained mapping inside
the same transaction; exceeding the budget rolls back notes, sibling edits,
events and retention. It counts more than notes alone, including current saved
state and other Resources. This reversible preview restriction prevents append
from making the workspace unreadable while notes replacement/clear is unavailable.
It does not change the other writers' existing admission policies or promise an
unlimited record size. An empty append supplied alongside a changing scalar
field still applies this postcondition; omit the append flag for a purely scalar
repair. A complete no-op returns before this check and changes nothing.

An empty append to empty notes is a no-op. An empty append to nonempty notes
**adds a newline**, matching the current native writer. Repeating an append
repeats its text. This is not an idempotent request/replay mechanism. Some earlier
issue prose described empty appends as universally no-op; the source behavior
above is what this preview preserves.

Omitting `--append-notes` preserves notes. It can be combined atomically with
admitted title, description, design, acceptance, priority and assignee edits.
A changed operation returns one complete Issue and mints exactly one retained
version, including owned blocking Dependencies. No-op returns the existing
complete Issue, revision and attribution. A stale guard refuses before no-op
planning. `--unconditional` chooses the current transactional predecessor; it
does not bypass ordinary assignment ownership policy.

Appending notes alone to a claimed Issue preserves its lease, holder and start
time. A mixed assignee edit still follows the existing transfer fence and lease
rules. Appending to a closed Issue preserves its status, close time and reason.
Dependencies, unrelated Memory Beads and informational Links are unchanged.
Existing authority, readonly and migration-freeze checks still apply.

The JSON result remains `{ "issue": <complete record>, "changed": <boolean> }`
inside the graph preview envelope. Human output is `Updated ID` or `Unchanged ID`;
quiet suppresses that human line.

## Boundaries and contributor work

`--notes`, clearing/replacement overrides, general status mutation, mixed claim
and append, and generic property replacement remain unavailable. The existing
notes-safety proposals [#5946](https://github.com/gastownhall/beads/pull/5946)
(transaction-level overwrite protection) and
[#6583](https://github.com/gastownhall/beads/pull/6583) (CLI replacement admission)
both preserve append. This adapter does not choose their disputed replacement
policy or supersede either contribution. It passes original `AppendNotes` intent
to the shared writer instead of constructing a replacement in the CLI.

Two overlapping writers may yield one accepted append and one known conflict.
There is no automatic application retry. After a known rollback, the caller can
read again and explicitly choose another guarded append. An unknown commit
outcome must not be blindly replayed: the text may already have been appended.

This pinned writer does not record its final retained version itself; the graph
adapter invokes the existing recorder once. Integrating upstream #6650 requires
reconciling recorder ownership and proving exactly-once and no-op behavior again.
The claim preview's five-minute nonrenewing lease and external-lease-drift
limitation remain unchanged. Full Issue workflows, Memory-required History,
atomic native History stamps and public HTTP writes remain incomplete or gated.

`bd status --graph --json` advertises `issueNotesAppend: true`,
`issueNotesReplace: false`, and keeps `issueWorkflows: false`.

## Reproduce the installed workflow

```sh
make install-force INSTALL_DIR=/absolute/disposable/bin
python3 scripts/graph-issue-append-notes-smoke.py \
  --bd /absolute/disposable/bin/bd --backend both --server-port PORT \
  --output-dir /absolute/new/notes-evidence
```

The caller owns a disposable ordinary Dolt 2.1.8 server. The harness runs normal
initialization and CLI authoring sequentially on both backends, preserving exact
command receipts, current/retained records, comparisons and failures. Different
databases must be provisioned serially on Dolt 2.1.8. Internal event/version
counts, forced writer overlap and lost-commit handling require the separate
real-store tests. A recipe alone is not qualification; the fork delivery plan
records exact executed source, artifacts and remaining limits.
