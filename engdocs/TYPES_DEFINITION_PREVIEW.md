# Types as ordinary Beads: increment 1 candidate

This is an unreleased implementation candidate, awaiting combined build and
public CLI qualification. The command sequence below is the acceptance surface;
this document does not certify a particular binary or main branch. Published
release documentation remains pinned by `docs/cli-docs.pin`.

Use a fresh disposable workspace with the explicitly selected embedded Types
profile. This profile uses schema14/generation13. Default schema11 and existing
workspaces are unchanged; opening an older workspace does not migrate or reseed
it. Keep existing test workspaces and their pinned binaries.

```sh
bd init --graph-mode link --graph-profile types --prefix bd \
  --scope-url https://example.invalid/my-types/
bd show type
bd show bead-type
bd show link-type
bd show refines
bd list
bd count
bd list --type type
bd list --type bead-type
bd list --type link-type
bd create --id bd-333 --type task "Hello world"
bd create --id note --type memory --body "A persisted note" "Note"
bd show bd-333
bd show note
```

ALL installed Type definitions are real, locally owned Beads. That includes the
four intrinsic definitions and the already supported Memory, Issue and Dependency
Types, including concrete classifications such as task. They participate in the
ordinary Show, List, Count, Graph and generic AST Query paths alongside domain
Beads. Link definitions are Beads; Link instances remain Links. Normal native
creation retains its validation and writes real persisted records. The existing
Issue ID prefix rule remains: `bd-333` is valid in a workspace configured with
prefix `bd`; bare `333` refuses without allocating a record.

Bootstrap writes the installed records. Checked compiler/index interpretations
refer to those same records, rather than a second mutable definition registry.
Additional configured native classifications become installed through their
existing native writer; a reader does not invent definitions from name spelling
or automatically enroll unused configuration.

The fresh intrinsic local IDs are `type`, `bead-type`, `link-type` and `refines`.
Their protocol identities remain common across workspaces, while locally owned
record IDs and revisions belong to the workspace. `type-for-` was a development
placeholder and is not part of these fresh IDs. The human field label is Type.
Full Resource output preserves actual Type pins, properties, metadata and owned
Link captures. Closing and reopening must preserve the complete persisted record.
A filtered Type family is not an inventory of only four built-ins; do not derive
its expected population from a hardcoded name list.

A missing `propertiesSchema` adds no local constraints. All ancestor schemas and
other constraints still apply. Supplying a schema adds constraints; null or a
malformed schema refuses. Existing exact publication1 records retain their
original explicit schema; publication2 may genuinely omit the local schema.
Neither reading nor a new binary rewrites the older edition.

Generic Query uses one complete checked capture for ordinary Beads, Type Beads
and Links. Unpinned request seeds can follow stored exact CURRENT endpoint pins
when the same native capture supplies explicit current-edition evidence; matching
body or private writer token text alone is insufficient. Explicit pinned request
references still refuse during admission, before storage is read. Retained, missing
or wrong-kind stored editions remain unsupported; no historical body is acquired.
No pin is stripped or silently replaced with current state. Legacy expression Query
can retain its existing Issue/Dependency-only boundary; use generic AST Query for
the common Resource surface.

The delivery order is deliberately incremental:

1. Qualify all installed definitions on ordinary reads and existing native
   creation, including the public CLI, generic AST and close/reopen sequence.
2. Add real `bd type create --type bead-type` and `--type link-type` definitions.
3. Add real `bd link create --type refines BASE DERIVED`, reconciling the convenient
   CLI operand order with the normative stored derived-to-base direction.
4. Add persisted custom `bd create/update --type MY_NEW_TYPE`.
5. Verify the complete sequence in batch.

Steps2–5 are pending increments, not supported-command claims in this candidate.
An unsaved instance prototype is not successful creation and cannot qualify any
of these increments. This document replaces the earlier definition-only demo
instructions; it does not authorize a store reset, reinit, migration or alias.
Packs, namespace design, general lifecycle redesign, server-profile adoption,
default promotion, release and global installation remain separate work.

For contributors, representation carriers validate shape and preserve bytes;
they do not grant authority from ID spelling, body shape or revision text.
Use the real installed/native population and independent complete-record/state
witnesses in tests. Exact source, normal build membership, affected tests and a
pinned executable public CLI pass must precede readiness. Follow
[TESTING.md](TESTING.md), retaining all failed and unexecuted evidence literally.
