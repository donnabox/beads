# Guarded Memory editing preview

The installed graph CLI can replace a Memory's complete title/body properties:

```sh
bd show beads/plan --json
bd update beads/plan --properties @replacement.json --if-revision '<observed revision>' --json
bd show beads/plan --json
bd show beads/plan --version '<earlier version>' --json
```

`replacement.json` must contain exactly two string members:

```json
{"title":"Deployment plan","body":"Updated Markdown knowledge.\n"}
```

Inline JSON and `--properties @-` (stdin) use the same validation and 1 MiB input
budget. Missing members, nulls, extra keys, duplicate keys and invalid Unicode
refuse. Strings are preserved verbatim. Explicit empty title/body strings are
valid under the existing experimental Type; omission does not clear a field.
The `remember` creation convenience still requires a nonempty title. These are
preview rules, not a settled production title-derivation contract.

Use exactly one `--if-revision TOKEN` or explicit `--unconditional`. A stale guard
refuses even if the replacement equals current properties. A semantic no-op
keeps the revision, attribution and retained state; changing only the actor does
not mint a version. Source-guard flags apply to owned Link operations, not to
this Memory edit: the Memory revision already includes its owned set.

A successful edit keeps the same canonical ID and Type, advances one Memory
revision/version, and atomically saves its complete properties and owned Links.
It does not version or recreate those Links, their targets, or incoming sources.
Changing an owned Link still advances its owner independently; subsequent body
edits preserve that newly observed owned set. The JSON result is the normal
preview envelope containing `memory` and `changed`.

This uses schema v5 in the same ordinary embedded/shared-server Dolt database.
It reuses the existing owned-Memory snapshot writer, with no Type/schema rewrite,
new store, migration, automatic retries or Issue recorder changes. Healthy Issue
selectors refuse as unsupported; existing informational Link editing is unchanged.
Read-only and migration-freeze policy still apply before writes.

## Qualification and boundaries

The installed demonstration uses normal initialization/authoring and fresh
processes on both engines. It checks current and old bodies, retained owned sets,
no-op/stale guard behavior, unrelated identities, empty strings, file/stdin input,
and explicit failures. Storage tests separately cover rollback, cancellation,
unchanged writer/version counts, two guarded writers, and a content edit racing
an owned Link edit. Embedded calls use the production one-session pool; their competing guards
serialize there. Ordinary-server tests force overlapping independent transactions.
An exploratory embedded two-session pool returned an uncertain commit outcome
from its driver; it is not a qualified production topology. No text-parsed error
or automatic retry is used to turn that outcome into a known rollback.

```sh
make install-force INSTALL_DIR=/absolute/memory-bin
python3 scripts/graph-memory-update-smoke.py \
  --bd /absolute/memory-bin/bd --backend both --server-port 3307 \
  --output-dir /absolute/new-memory-receipts
```

The supplied ordinary Dolt server must be disposable and caller-owned. Provision
different databases serially on Dolt 2.1.8. The harness preserves receipts and
reaps command processes; it retains disposable data for inspection.

The [public-client Read harness](GRAPH_BDP_READ_HTTP.md) also edits Memory through
the CLI and reads the changed body through the unchanged public BDP client. It
checks that owned Links stay unchanged. That demonstrates Read after a local
write, not an HTTP mutation endpoint.

`status --graph` reports `memoryPropertiesUpdate: true` and the Memory input
budget. Complete Memory and public History remain false. This slice does not
add patch operations, common metadata, aliases, generic deletion, ordered
versions, as-of selection, comparison, restoration, public HTTP Update/History,
or production adoption. Existing attribution is not a native History commit
stamp. Ordering, immutable change context and v5 History admission still require
explicit design review; no timestamps or token spelling are used to invent them.

The properties input bound, 1,000-owned-Link descriptor bound and current Read's
16 MiB acquisition budget are separate. An accepted aggregate is not guaranteed
to fit every Read budget; no stored record or snapshot is silently truncated.

This is a draft increment under [fork plan #18](https://github.com/donnabox/beads/pull/18).
CLI replacement syntax is proposed in [upstream #6703](https://github.com/gastownhall/beads/issues/6703).
Exact commits, receipts and remaining qualification are recorded in the review PR.
