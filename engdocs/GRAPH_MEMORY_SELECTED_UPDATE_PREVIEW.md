# Selected Memory body editing preview

A graph workspace can revise an existing Memory with `remember`, preserving its
canonical identity, omitted title and complete owned Links. Creation still uses
`--id`; selected editing uses `--update`. A missing selected Memory is an error,
not a request to create one.

This is a provisional subset of [Memory R4](https://github.com/gastownhall/beads/issues/5877)
and [CLI §10.1](https://github.com/gastownhall/beads/issues/6703#issuecomment-5796542827).
The command shape and guard ergonomics require review before becoming durable
contracts. It does not settle naming/upsert, metadata placement or title derivation.

## Example

Start with ordinary initialization and explicit creation in a disposable directory:

```sh
bd init --graph-mode link --scope-url https://example.invalid/notes/ \
  --prefix demo --non-interactive --skip-hooks --skip-agents
bd remember 'The original body' --id beads/plan --title 'The plan'
bd show beads/plan --json
```

Copy the returned `result.revision` into `REV`. Then provide new content explicitly:

```sh
bd remember --update beads/plan --if-revision "$REV" --body-file revised.md --json
bd recall beads/plan
```

`--update` accepts a canonical local Bead path or a canonical URL in this Scope.
The revision must be the caller's observation. Discovery summaries also supply a
version; preview revisions and retained versions are currently equal. A caller
can recall that exact observed body with `recall ID --version TOKEN` before editing.

Exactly one explicit body source is required: positional text, `--body-file FILE`,
or `--stdin`. Empty positional text or an empty file/stdin clears the body;
a missing body source is an error. Input preserves valid UTF-8 bytes, including
Unicode, frontmatter, CRLF and whitespace, up to 1 MiB. Supplying `--title`
replaces the title, including an explicit empty string. Omitting it preserves the
title; updating the body never silently derives a different title. Creation's
existing requirement for an explicit nonempty title is unchanged.

## Guard and retained state

The command reads the selected Memory to compose the supported title/body payload,
then calls the existing guarded Memory writer with the unchanged caller revision.
The writer checks current authority and the complete graph revision within its
transaction. An intervening body, title or owned-Link mutation causes a conflict.
The adapter does not refresh the guard or retry the edit.

An identical edit returns `changed: false` without minting a version or changing
attribution, but a stale guard still fails even for identical content. A changed
edit returns the complete Memory with a new revision/version and the caller's
attribution. Current targets and Links remain unchanged; earlier complete Memory
states remain readable with `show --version` and `recall --version`.

JSON uses the existing experimental mutation envelope:

```json
{"schemaVersion":1,"preview":true,"result":{"memory":{"id":"...","type":"...","version":"...","revision":"...","properties":{"title":"The plan","body":"..."},"owned":[],"attribution":{"actor":"...","status":"claimed","recordedAt":"..."}},"changed":true}}
```

Guarded updates have no `replaced` disclosure. Human output identifies the canonical
ID and quoted title and distinguishes updated from unchanged; `--quiet` is silent.
No successful result is printed before transaction completion and store cleanup.
An uncertain COMMIT response remains an unknown outcome, with no automatic retry.
Recorded attribution times are observations, not native commit timestamps.

## Deliberate limits

- This body-edit preview requires `--if-revision`; `--unconditional` is unsupported.
  Reading a title and then unconditionally replacing it could overwrite a concurrent
  title change. Existing complete `update --properties ... --unconditional` and its
  predecessor disclosure remain available. A safe partial unconditional write is
  separate work requiring an in-transaction omission-preserving operation.
- Title-only selected edits with omitted body are unavailable. Use existing complete
  title/body replacement if needed. `--id` and `--update` are mutually exclusive;
  creation does not accept a revision guard. Legacy keyed remember is unchanged.
- Keys/aliases, automatic identity/title generation, Inception, derivation, common
  metadata, complete Memory JSON recall, deletion/restoration and native/public
  History remain open. No new schema, field placement or BDP wire contract is added.
- The checked current-record reader applies its existing 16 MiB workspace
  acquisition budget before composing an edit. Oversized unrelated records can
  therefore make this convenience command refuse even when the selected Memory
  is small. The 1 MiB body limit is an input limit, not a guarantee of read capacity.
- Unregistered flags such as `--unconditional` produce ordinary Cobra diagnostics,
  before the graph route runs; they do not promise a typed graph JSON error.
- Existing readonly and migration-freeze policies apply before reading body inputs.
  Graph-only flags refuse in legacy workspaces before body acquisition or legacy
  storage opening. Foreign selectors, missing Resources and wrong kinds refuse.
- `memorySelectedUpdate` is true; `memorySelectedUpdateUnconditional`, full `memory`
  and `historyExact` remain false. Exact retained reads are not native History.

## Validation

`scripts/graph-memory-selected-update-smoke.py` exercises normal installed CLI
initialization, creation, owned-Link authoring, editing, discovery and exact/current
reads in fresh processes on embedded and caller-supplied ordinary shared-server Dolt.
It records commands, output hashes and expected complete records. It never seeds a
schema or writes SQL fixtures. `TestGraphRememberSelectedUpdateInterveningWrites`
separately composes an edit, accepts a competing title or Link mutation through real
store APIs, then proves that the original edit refuses and leaves the winner intact.

Existing Memory writer tests cover transaction rollback, forced concurrent writer
overlap and lost COMMIT responses. This adapter introduces no storage writer or new
locking mechanism. Dolt 2.1.8 different-database provisioning remains serialized.
The ordinary test database server is disposable and caller owned.
