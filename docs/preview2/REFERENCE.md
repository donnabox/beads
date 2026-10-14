# Beads Graph Preview 2 CLI reference

This reference describes the **manual-test candidate** at the protected
[`graph-preview2-rc.1` tag](https://github.com/donnabox/beads/tree/graph-preview2-rc.1),
which resolves to commit `5879467f017295147387144392398b2d3cd360a4`.
It records what testers can exercise now. The broader
[`Graph CLI Specification (Draft)`](../reference/graph-cli-specification-draft.md)
contains proposals and older examples; it is **not** the support contract for
this candidate. Follow the [runbook](RUNBOOK.md) for copyable examples and its
**Known Limitations** section for the detailed command-by-command list.

## Start safely

- Use a **new, empty workspace**. Preview data may be lost, and there is no
  supported upgrade from an ordinary Beads workspace or another preview.
- `bd init --graph-mode link` currently requires a unique `--scope-url`, even
  when using only the CLI. The URL names the workspace; it need not serve HTTP.
- The embedded CLI walkthrough needs no separate Dolt server. The optional
  shared-server and BDP Read exercise uses Dolt **2.1.8** in a separate new
  data directory.
- Build the tagged source with `make build`. Its `go.mod` selects Go 1.26.7;
  the Makefile enables CGO and the `gms_pure_go` build tag. The fork has not
  published a Preview 2 binary package.

## Types, kinds, and relationships

| Bead Types | Link Types |
| --- | --- |
| `preview-issue` | `preview-dependency` |
| `preview-memory` | `preview-related` |
| | `example-cites` |
| | `example-follows` |

`--bead-type` and `--link-type` select installed Types. Users cannot install or
edit Types in this preview. An Issue's `--type` is its `issue_type` value;
unconfigured Issue kinds are refused. A Dependency's `--type` is its
`dep_type` value. It accepts a nonempty custom kind up to 32 bytes, but only
well-known kinds have built-in workflow behavior. Omitting it means `blocks`.

A native Dependency connects two **local Issues** in one Scope. There can be
only one native Dependency for an ordered source–target Issue pair. `blocks`
affects readiness. `parent-child` is also a Dependency kind, directed **from
child to parent**. Creating a custom kind such as `tracks` does not make it
blocking. Native Dependency Links have no custom properties or metadata in
this candidate; their kind is chosen at creation and cannot be updated.
Informational Links can carry Type-checked properties and metadata.

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

Use `bd link` for **all** installed Link Types, including Dependencies. The
`bd dep` convenience commands address Issue Dependencies; their list, tree,
cycles, and pair-removal forms cover **blocking** Dependencies only. They are
not general Link commands. `bd link query` is unavailable.

`bd remember`, `bd recall`, `bd memories`, and `bd forget` are Memory-focused
conveniences. `bd history` lists saved graph record versions rather than the
ordinary Issue event log. The older `bd links` and `bd unlink` aliases may be
present, but this reference and runbook use `bd link list` and
`bd link delete`.

## Write and read boundaries

- Bead deletion never cascades. Remove each incident Link first. Applying
  `bd delete` requires `--force` plus a current revision token or explicit
  `--unconditional`; a plain delete previews the effect.
- Native Dependency Link property, metadata, and kind updates are refused.
  `bd link update` supports informational Link properties and metadata.
- `bd create --deps`, batch Bead deletion or update, remote or version-pinned
  Link endpoints, and cross-Scope Dependencies are unavailable.
- `bd list` lists Beads, not Links. Its ordinary `--json` output is Issue-only;
  use `--format records-json` for mixed Bead records and `bd link list` for
  Links. Collections are bounded; there is no CLI cursor paging.
- `bd show` reads one resource per call. Use `bd versions` and `bd compare`
  with saved opaque version tokens; do not treat ordinal numbers as portable
  version addresses.
- The optional BDP HTTP service supports **Read only**. HTTP writes, HTTP
  History, import/export migration, and general backup/recovery are outside
  this candidate.

The runbook's [Known Limitations](RUNBOOK.md#known-limitations) separates
commands believed complete for its examples from partially supported and
unavailable commands. A refusal is not evidence that a command is supported.
Please include the exact source commit, platform, command, and output in
manual-test feedback on the documentation PR.
