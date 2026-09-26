# Exact graph comparison preview

`bd compare RESOURCE --from TOKEN --to TOKEN` compares two complete retained
preview versions of one Memory, Issue or Link. Both tokens are explicit opaque
values obtained from earlier command results; their spelling does not establish
order or time. `bd diff` retains its separate Dolt-ref behavior.

```sh
bd compare beads/plan --from "$before" --to "$after" --json
bd compare links/context --from "$old_link" --to "$new_link"
```

The pair is acquired through the existing retained decoders in one
transaction with the workspace authority checked. No current payloads or target
bodies are substituted. Same-token comparison still resolves that version;
unknown versus unknown fails. Either unavailable or invalid operand fails the
whole operation before result output. Output-device failures return failure too,
although a partially written stream cannot be recalled. A deleted Link's retained live versions
remain comparable; its private deletion marker is not a Resource version.

The result compares every serialized property and the complete recorded owned
set. Property differences carry complete old/new values, including explicit
presence so absence differs from null or empty content. Arrays preserve order;
object member order is not a difference. JSON numbers are compared without
float64 rounding. Property members and owned-Link IDs use deterministic UTF-16
code-unit ordering. Issues use their retained durable projection and Dependency
owned set, preserving the existing Issue recorder.

Owned Links match by canonical identity, including two Links with equal endpoints.
Changed owned versions remain a difference even when their properties changed
away and back. The selected Resource's own token and attribution are context,
not themselves property differences. Immutable ID, Type or Link endpoint
mismatches refuse rather than presenting a supported retarget or retype operation.

## Proposed output, awaiting review

This is a reversible experimental format, not a stable public contract. The
existing `--json` preview envelope contains the result below; human output is
that same complete result as indented JSON, with control characters escaped.
No unified-diff algorithm or executable patch semantics are implied. `--quiet`
suppresses human output; `--quiet --json` retains structured output.

A body change contains an entry like:

```json
{"area":"properties","member":"body","from":{"present":true,"value":"Before\n"},"to":{"present":true,"value":"After\n"}}
```

An owned Link removal contains `area: "owned"`, its canonical `id`, the complete
old Link in `from: {present: true, value: ...}`, and `to: {present: false}`. An
addition reverses presence. A changed Link carries both complete records. An
empty difference is `changes: []`, with both selected versions and attribution
still present. Reversing the tokens reverses the comparison direction.

The result also identifies the Resource, its Type (and Link endpoints), the
compared domains, and unsupported fields. Common metadata is not implemented;
Memory Inception and derivation are also unavailable. Existing Issue
`properties.metadata` remains an Issue property, not the absent common field.
These limitations prevent this preview from satisfying the full proposed
Memory comparison or public History contract. The final envelope, presence
encoding and version-only owned-difference presentation need review before
becoming durable contracts.

## Bounds and evidence

Each exact operand has the existing 16 MiB retained-input acquisition budget;
the pair reads at most two such inputs (32 MiB), or one for equal tokens. This
bounds retained bytes before decoding, not total heap or rendered output size.
Each token is nonempty UTF-8 at most 4096 bytes. No accepted comparison is
silently truncated. Read-only flags work; comparison adds no writer, schema,
History order, lineage or native immutable commit-time context.

The installed demonstration uses normal disposable initialization and authoring:

```sh
python3 scripts/graph-history-compare-smoke.py --bd /absolute/path/to/bd \
  --backend both --server-port 15440 --output-dir /absolute/path/to/receipts
```

The supplied ordinary Dolt server is caller-owned. Provision different databases
serially on Dolt 2.1.8. Exact commits, test receipts and qualification limits belong
in the review PR under [fork plan #18](https://github.com/donnabox/beads/pull/18).
CLI shape feedback belongs in [upstream §11.1](https://github.com/gastownhall/beads/issues/6703).
Aliases, full version-address selection, ordered enumeration, as-of, restoration,
common metadata, erasure and public HTTP History remain unfinished.
