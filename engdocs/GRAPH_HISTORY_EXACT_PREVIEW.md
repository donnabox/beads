# Exact retained-version preview

This draft increment implements the proposed `show --version` spelling from
[CLI proposal #6703](https://github.com/gastownhall/beads/issues/6703). It is local
CLI access to existing retained records, not the public BDP History capability.
It does not change the v5 schema, migrate old workspaces, or manufacture missing
version order or change context. The broader delivery plan stays in
[fork PR18](https://github.com/donnabox/beads/pull/18).

## Use

Initialize and author a disposable graph workspace using the
[C0 guide](GRAPH_C0_PREVIEW.md). Capture the `version` returned by a current
`show --json`. Then read that exact retained version, even after later writes:

```sh
bd show beads/plan --json
bd show beads/plan --version '<version from the earlier read>' --json
bd show links/context --version '<earlier Link version>' --json
```

The provisional selector accepts the opaque token returned for that Resource.
It does not interpret local ordinals, timestamps, `@v3`, or full version URLs.
Canonical Resource selectors retain their existing path/URL rules. Tokens must
be nonempty UTF-8 and at most 4,096 bytes. The command uses the same preview
record envelope as current reads, with its original revision/version, properties,
attribution and complete owned set. Human output prints the complete record too.

A historical Memory retains its original owned Links after those Links change
or disappear. A deleted Link's old Resource versions remain readable; its private
deletion token is not a Resource version. Historical Issues reuse the existing
Issue recorder's durable state and the graph revision's saved owned mapping.
No body or owned state is filled in from today's Issue or Link tables.

Both embedded and ordinary shared-server Dolt use the same exact reader. Each
request checks current workspace/authority binding in the read transaction.
The reader bounds selected retained bytes before loading them; it may refuse an
oversized record with `capability_unavailable`. This is an acquisition refusal,
not a retention policy or evidence that a version was removed.

## Outcomes and limits

- Unknown tokens, including a token belonging to another Resource, return
  `revision_unknown`. Missing subjects return `not_found`.
- A private Link deletion token returns `gone`; it never becomes a fabricated
  historical Link record. Ordinary current reads of that Link still return `gone`.
- Positive malformed or missing required state refuses as `graph_not_initialized`.
  This internal-store failure is not a public BDP History disposition.
- `status --graph` distinguishes `exactVersionRead: true` from `historyExact: false`.
  The latter remains false because the full public capability is not implemented.
- No HTTP History advertisement, ordered `versions`, as-of reads, comparisons,
  restoration, generic Memory update/delete, aliases, erasure or import admission
  is delivered here. Existing BDP Read remains unchanged.

Current v5 snapshots lack lawful generic authority order. A future ordering and
context recorder needs explicit format/admission review; token spelling and
claimed timestamps cannot be used to invent chronology. Existing immutable Type
descriptors and old stored records are not rewritten by this slice. Jim's Issue
recorder and storage remain authoritative.

## Reproduce the installed demonstration

Use the canonical installer, never a source-tree binary built by an alternate
command. Choose an absolute output path for the installed binary and receipts:

```sh
make install-force INSTALL_DIR=/absolute/history-bin
python3 scripts/graph-history-exact-smoke.py \
  --bd /absolute/history-bin/bd \
  --backend both --server-port 3307 \
  --output-dir /absolute/new-history-receipts
```

The supplied ordinary Dolt server must be disposable and caller-owned. The
harness provisions fresh workspaces sequentially, uses normal CLI authoring, and
checks exact old records from new processes after Link update/unlink and Issue
changes. It captures binary/harness hashes, every command, typed failures and
cleanup. It does not seed SQL or a schema. Different-database provisioning on
Dolt 2.1.8 must remain serialized; this harness is not an engine fix.

This document describes the intended increment. Qualification receipts and exact
commits are recorded in the review PR; a component test alone is not an installed
demonstration or complete Memory delivery.
