# Beads Graph Preview 2 CLI reference

This reference describes the frozen manual-test candidate at the
[`graph-preview2-rc.3` tag](https://github.com/donnabox/beads/tree/graph-preview2-rc.3),
commit `97446d1f0894417457f9e9b8f16569baf73d68f4`. The [runbook](/preview2/RUNBOOK)
has copyable examples and a command-by-command **Known Limitations** section.
The broader [`Graph CLI Specification (Draft)`](/reference/graph-cli-specification-draft)
contains proposals and older examples; it is not the contract for this candidate.

## Start safely

- Use a new, disposable workspace. Preview data may be lost, and there is no
  supported upgrade between previews.
- `bd init --graph-mode link` currently requires a unique `--scope-url`, even
  when using only the CLI. The URL names the workspace; it need not serve HTTP.
- The embedded walkthrough needs no separate Dolt server. The optional
  shared-server and BDP Read exercise uses Dolt 2.1.8 in a separate data directory.
- Build the tagged source with `make install`. Its `go.mod` selects Go 1.26.7;
  the Makefile enables CGO and the `gms_pure_go` build tag. The fork has not
  published a Preview 2 binary package.

## Types and relationships

The installed nominal Bead Types are `preview-issue` and `preview-memory`.
The installed nominal Link Types are `preview-dependency`,
`preview-related-memory`, `example-cites`, and `example-follows`.
`bd types` shows the **CLI choices** users can pass to `--type`, including
`task`, `bug`, and the other registered Issue choices, plus `blocks`,
`parent-child`, and the other built-in Dependency choices. The nominal
`preview-issue` and `preview-dependency` names are not extra choices in that
human-readable list. Users cannot install or edit graph Type descriptors in
this preview.

`bd create --type task` makes an Issue, while `bd create --type preview-memory`
makes a Memory. Omitting `--type` from `bd create` makes a `task` Issue.
`bd link create SOURCE TARGET --type blocks` makes a Dependency. It also
accepts `--source SOURCE --target TARGET`; it requires `--type`. The shorter
`bd link SOURCE TARGET` makes a blocking Dependency. An unrecognized `--type`
is refused by `bd create` and `bd link create`. The legacy `bd dep add --type`
still accepts an arbitrary nonempty Dependency label up to 32 bytes; only
built-in choices have their defined workflow behavior.

A Dependency connects two local Issues in one Scope. Only one native
Dependency may exist for an ordered source–target pair. `blocks` affects
readiness. `parent-child` is directed from child to parent and also defines
the hierarchy. Other built-in Dependency choices have their own meanings;
creating a nonblocking choice does not by itself block work. Native
Dependency Links have no custom properties or metadata in this candidate,
and their chosen kind cannot be updated. Informational Links can carry
Type-checked properties and metadata.

## CLI namespaces

| Operation | Bead command | Link command |
| --- | --- | --- |
| Create | `bd create` | `bd link create` |
| List | `bd list` | `bd link list` |
| Show | `bd show` | `bd link show` |
| Update | `bd update` | `bd link update` |
| Delete | `bd delete` | `bd link delete` |
| List versions | `bd versions` | `bd link versions` |
| Compare versions | `bd compare` | `bd link compare` |
| Count | `bd count` (Issues) | `bd link count` |

Use `bd link` for all installed Link Types, including Dependencies. The
`bd dep` convenience commands address Issue Dependencies; their list, tree,
cycles, and pair-removal forms cover blocking Dependencies only. They are not
general Link commands. `bd link query` is unavailable.

`bd remember`, `bd recall`, `bd memories`, and `bd forget` are Memory-focused
conveniences. `bd history` lists saved graph record versions rather than the
ordinary Issue event log.

## Write and read boundaries

- Bead deletion never cascades. Remove each live incident Link first. Plain
  `bd delete ID` previews one deletion; `--force` applies it. The optional
  `--if-revision TOKEN` rejects a stale write. `bd link delete` also needs
  `--force` to apply. `bd forget ID` immediately deletes an unreferenced
  Memory. Its help text still mentions a removed `--unconditional` flag;
  that flag is not accepted.
- `bd link update` supports informational Link properties and metadata.
  Native Dependency Link property, metadata, and kind updates are refused.
- `bd create --deps`, batch Bead deletion or update, remote or version-pinned
  Link endpoints, and cross-Scope Dependencies are unavailable.
- `bd list` lists Beads, not Links. Its ordinary `--json` output is Issue-only;
  use `--format records-json` for mixed Bead records and `bd link list` for
  Links. Collections are bounded; there is no CLI cursor paging.
- `bd show` reads one resource per call. Use `bd versions` and `bd compare`
  with saved opaque version tokens; do not treat ordinal numbers as portable
  version addresses.
- The optional BDP HTTP service supports Read only. HTTP writes, HTTP
  History, import/export migration, and general backup/recovery are outside
  this candidate.

The runbook's [Known Limitations](/preview2/RUNBOOK#known-limitations) separates
commands believed complete for its examples from partially supported and
unavailable commands. A refusal is not evidence that a command is supported.
Please include the exact source commit, platform, command, and output in
manual-test feedback on the documentation PR.
