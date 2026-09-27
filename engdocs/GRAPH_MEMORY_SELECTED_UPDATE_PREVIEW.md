# Selected Memory editing preview

A graph workspace can revise an existing Memory with `remember`, preserving its
canonical identity, omitted title/body and complete owned Links. Creation uses
`--id`; selected editing uses `--update`. A missing selected Memory is an error,
not a request to create one.

This is a provisional subset of [Memory R4](https://github.com/gastownhall/beads/issues/5877)
and [CLI §10.1](https://github.com/gastownhall/beads/issues/6703#issuecomment-5796542827).
Command shape and guard ergonomics require review before becoming durable
contracts. Naming/upsert, metadata placement and title derivation remain open.

## Examples

Start with ordinary initialization and explicit creation in a disposable directory:

```sh
bd init --graph-mode link --scope-url https://example.invalid/notes/ \
  --prefix demo --non-interactive --skip-hooks --skip-agents
bd remember 'The original body' --id beads/plan --title 'The plan'
bd show beads/plan --json
```

Copy the returned `result.revision` into `REV`, then provide content explicitly:

```sh
bd remember --update beads/plan --if-revision "$REV" --body-file revised.md --json
bd recall beads/plan
```

To change only the title, use the latest observed revision:

```sh
bd remember --update beads/plan --if-revision "$REV" --title 'Revised plan' --json
```

To explicitly edit the current state without an observed revision guard:

```sh
bd remember --update beads/plan --unconditional --body-file revised.md --json
```

`--update` accepts a canonical local Bead path or URL in this Scope. Exactly one
write policy is required: `--if-revision REV` or `--unconditional`. Both together,
an empty revision, or `--unconditional=false` refuse before input acquisition.
Discovery summaries supply a selected version; preview revisions and retained
versions are currently equal. Recall that exact observed body with
`recall ID --version TOKEN` before editing it.

## Supplied fields and input

A selected edit must supply `--title`, one explicit body source, or both. Omitting
a field preserves it from the actual checked predecessor inside the write
transaction. An explicit empty string clears that field. An empty patch refuses.

Body sources are positional text, `--body-file FILE`, or `--stdin`. Multiple sources
refuse. An empty positional argument or empty file/stdin is a supplied empty body.
Valid UTF-8 bytes, including Unicode, frontmatter, CRLF and whitespace, are preserved
up to 1 MiB. Text beginning with a dash is easiest to supply through a file or
explicit stdin; normal command-line option parsing still applies.

A title-only edit never reads stdin. Attached pipe content is ignored unless
`--stdin` is explicit; a request with only an unmarked pipe and no title refuses.
Updating body never derives a new title. Creation still requires its explicit
nonempty title and exactly one explicit body source, and rejects either write guard.

## One transaction and retained state

Selected editing now calls `PatchMemory`, which shares the existing Memory mutation
writer with complete `update --properties` replacement. It resolves omitted fields
from the checked current Memory inside that same transaction. It does not construct
an unconditional replacement from a separate read, refresh a guard, or retry.

An intervening body, title or owned-Link mutation invalidates an observed graph
revision. Guard validation precedes no-op detection, so a stale no-op still fails.
An identical accepted edit returns `changed: false` without minting a version,
changing attribution or returning replacement disclosure, even with another actor.

Changed edits retain the complete Memory and owned set under a new version. Current
Links and targets remain unchanged. Earlier complete states remain readable with
`show --version` and `recall --version`. Changed unconditional edits return the actual
replaced Memory ID, version and recorded attribution under `replaced`; guarded edits
omit it. This reuses the existing provisional R14 result shape, without inventing a
native commit timestamp. Human output names the identity and quoted title, and
includes predecessor disclosure when applicable. `--quiet` is silent; JSON retains
the existing experimental `memory`, `changed` and optional `replaced` envelope.

A changed guarded edit has this result inside the normal CLI envelope (complete
records are abbreviated for readability):

```json
{
  "memory": "<complete accepted Memory record>",
  "changed": true
}
```

A changed unconditional edit adds the actual predecessor's context:

```json
{
  "memory": "<complete accepted Memory record>",
  "changed": true,
  "replaced": {
    "id": "https://example.invalid/notes/beads/plan",
    "version": "<actual previous opaque version>",
    "attribution": "<exact recorded attribution from that predecessor>"
  }
}
```

The `memory` and `attribution` placeholders represent objects in real output. No-op results
have `changed: false` and omit `replaced` under either write policy.

Unconditional editing can still conflict with an overlapping transaction. It is
an explicit choice to replace supplied fields, not permission to ignore authority
or concurrency failures. No successful result is emitted before commit and store
cleanup. Unknown COMMIT outcomes return no success or predecessor disclosure and
are never automatically replayed.

## Boundaries

- No preliminary public current read or whole-workspace acquisition is required
  by this writer. It still checks the selected record, authority, Type, retained
  integrity, guard, and complete owned state. Existing read/collection budgets
  remain unchanged; this does not remove limits from `show`, discovery or BDP.
- `--id` and `--update` are mutually exclusive. Graph-only flags refuse on the
  legacy route before body acquisition or legacy storage opening. Legacy keyed
  remember and existing complete `update --properties` semantics are unchanged.
- Existing readonly and migration-freeze policy applies before inputs are read.
  Wrong-kind Issues retain the existing `capability_unavailable` classification;
  malformed/foreign or Link selectors are `invalid_selector`.
- Keys/aliases, automatic identity/title generation, Inception, derivation, common
  metadata, full structured Memory recall, deletion/restoration and native/public
  History remain open. No schema, BDP wire or durable contract is added.
- `memorySelectedUpdate` and `memorySelectedUpdateUnconditional` are true. Full
  `memory` and `historyExact` remain false. Exact saved states are not native History.
  This preview deliberately widens `memorySelectedUpdate` to accept omitted body
  and title-only edits without adding another capability key. In this increment,
  `memorySelectedUpdateUnconditional: true` accompanies that wider behavior;
  clients must not infer title-only support from `memorySelectedUpdate` alone on
  older builds. Capability naming remains part of the CLI review gate.

## Qualification

The installed `graph-memory-selected-update-smoke.py` retains guarded body-edit
coverage. `graph-memory-atomic-patch-smoke.py` adds title-only and unconditional
omission preservation, exact predecessor disclosure, no-op/stale/refusal/output and
legacy controls on embedded and caller-supplied ordinary shared-server Dolt. Both
use normal initialization and CLI authoring, fresh processes, complete expected
records and captured output hashes. No seeded schema or hidden bootstrap.

`TestMemoryPatch*` uses real store APIs and the existing transaction failure/overlap
controls to exercise the shared writer. The older CLI read/compose interleaving
helper/test was removed with that implementation; its historical receipts remain.
Atomic storage checks now qualify omission preservation at the actual write boundary.
Different-database provisioning stays serialized on Dolt 2.1.8. No application lock,
engine fork, repair loop or fabricated History context is introduced.
