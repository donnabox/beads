# Memory properties patch preview

In a disposable `graph_mode: link` workspace, `bd update` can apply an ordered
properties patch to an existing Memory. This uses the checked Memory writer in
one transaction. It does not add an HTTP mutation endpoint, a public BDP Write
profile, full Memory compatibility, or native History.

```sh
bd show beads/plan --json
bd update beads/plan --patch '[{"op":"replace","path":"/body","value":"Revised plan — 雪"}]' --if-revision REVISION_FROM_SHOW --json
bd update beads/plan --patch @changes.json --unconditional --json
printf '%s' '[{"op":"replace","path":"/title","value":"Next steps"}]' |
  bd update beads/plan --patch @- --if-revision REVISION_FROM_SHOW --json
```

These are command examples, not a recorded qualification run. Use the exact
opaque revision returned by the current Memory read. Literal JSON, `@file`, and
explicit `@-` are the three input forms; a pipe is not consumed implicitly.
Unicode, whitespace and CRLF inside string values retain their meaning.

## Accepted operations and final representation

The patch is a nonempty JSON array of `add`, `replace` and `remove` operations.
Paths are JSON Pointers relative to the properties object. The shared evaluator
handles escaped member names, existing-parent/target requirements, ordered array
insertion/removal and `-` append semantics. It rejects other operations, unknown
operation members, duplicate JSON keys, malformed Unicode and inadmissible
numbers before evaluation. It does not silently repair or round JSON values.

The implemented operation semantics and nonempty array requirement follow the
selected BDP reference at `53bdbd03136875f952af184fce7b3c7af8f74e96`;
see the evaluator's source attribution.
The pinned `propertyChange` schema requires `minItems: 1`, so the CLI rejects
`[]` with `invalid_properties`. A nonempty same-value or reversing patch is valid syntax
and can be a transactional no-op.

The final properties object must still contain exactly `title` and `body`, both
UTF-8 strings. That is the existing experimental Memory representation, not a
new public BDP Type restriction. Empty strings are valid. Removing `body` and
adding it again in the same patch can succeed; leaving it missing or null
refuses the entire operation. Temporary extra members, arrays and intermediate
values are evaluated in order, then removed or replaced before final validation.
Leaving extra members, a nonobject root or nonstring title/body refuses atomically.
No schema, persisted Type, common metadata or alias contract changes here.

## Guards, dispatch and output

Supply exactly one of `--if-revision TOKEN` or `--unconditional`. A stale guard
refuses even when the patch would be a no-op. The writer evaluates against the
actual accepted predecessor within its transaction, preserving untouched fields
and the complete owned Link set. The CLI does not pre-read a Memory and replace
it from a stale local copy.

On the Memory route, `--patch` is mutually exclusive with `--properties` and Issue/Link update flags,
including explicitly false claim or source-guard flags. Read-only and migration
freeze checks, selector admission, unsupported flags and guard validation happen
before patch files/stdin are consumed or the graph store is opened. Ordinary
workspaces refuse this graph-only flag before legacy storage opening. Canonical
Link selectors select the separate [informational Link patch route](GRAPH_LINK_PROPERTIES_PATCH_PREVIEW.md),
which admits independent Link and source guards. Mixing `--patch` with
`--properties` still refuses before input on either route. A canonical `beads/PATH` does not reveal its
Type: a healthy Issue at that path is refused by the checked writer after input
syntax admission, without mutation.

JSON uses the existing preview envelope with a complete `MemoryMutationResult`.
Human output says `Updated` or `Unchanged` and the canonical ID. Existing quiet
behavior and unconditional predecessor disclosure are preserved. Only a changed
unconditional operation carries `replaced`; no-op, refusal and unknown outcome
do not manufacture a replacement receipt. Output begins after the operation
and ordinary store cleanup succeed.

There is no automatic application retry, request-status service or durable
idempotency key in this preview. An uncertain COMMIT remains an unknown outcome;
do not replay it automatically. Exact saved reads and comparisons remain the
existing preview features, not an ordered History traversal or immutable native
change-context promise.

## Explicit preview limits

`status --graph` reports `memoryPropertiesPatch: true` and these private bounds:

| Limit field | Maximum |
|---|---:|
| `memoryPatchInputBytes` | 1,048,576 raw input bytes |
| `memoryPatchOperations` | 256 operations |
| `memoryPatchPointerBytes` | 4,096 UTF-8 bytes per Pointer |
| `memoryPatchPointerSegments` | 64 decoded segments per Pointer |
| `memoryPatchDepth` | 64 levels of container nesting |
| `memoryPatchDocumentBytes` | 1,048,576 canonical working-property bytes |
| `memoryPatchEvaluationBytes` | 16,777,216 cumulative charged bytes |

The working-document bound applies to the predecessor, intermediate operations
and final document. Acquisition reads at most the input limit plus one byte to
detect excess. No input or result is truncated. Before each operation, evaluation
charges the current canonical document bytes plus the incoming value bytes
(zero incoming value bytes for removal). These charges accumulate across the
patch and must fit `memoryPatchEvaluationBytes`. This bounds repeated work on
large documents even when every intermediate document fits its individual limit.
It is a work-charge proxy, not an exact heap, allocation or elapsed-time bound.

For a changed patch, the writer also requires the resulting workspace to fit the
existing 16 MiB current-read acquisition budget. That whole-workspace
postcondition is separate from the per-document and cumulative evaluation bounds.
A no-op writes nothing and does not run this additional post-write check; its
guard and patch parsing/evaluation limits still apply. Existing full replacement
and selected title/body update behavior stays intact. These private, reversible
limits do not establish negotiated public BDP limits or modify storage columns.

### Error categories

The CLI preserves the existing input-admission and checked-writer error
boundaries; there is no universal exit code for every patch limit.

| Failure boundary | CLI code | Exit |
|---|---|---:|
| Input acquisition, invalid patch syntax, or Parse limits: raw/canonical patch bytes, operation count, Pointer bytes/segments and supplied-value depth | `invalid_properties` | 2 |
| Invalid operation against the predecessor, or invalid completed Memory title/body representation | `invalid_properties` | 2 |
| Writer limits: predecessor/intermediate document bytes or depth, cumulative evaluation charge, or changed-result workspace acquisition budget | `capability_unavailable` | 5 |
| Invalid selector or guard admission | `invalid_selector` | 2 |

For example, an input one byte over the raw patch limit is refused at input
admission with exit 2. A valid small patch against existing Memory properties
larger than the working-document limit is refused by the writer with exit 5.
Depth similarly has an input-value check and an applied-document check. Both
refusals preserve state; the distinction identifies the boundary that refused
rather than promising an alternate route can always perform the edit.

The CLI parses before opening storage so malformed input cannot acquire a
store. The raw-byte storage API independently parses and captures operations for
its own callers. This deliberate validation at both boundaries keeps the storage
API safe without introducing a second constructor that trusts CLI admission.

Qualification must include the installed CLI on embedded and ordinary shared-
server Dolt, normal initialization and fresh-process reads, exact retained owned
state, transactional no-ops, competing guarded/unconditional writers, rollback,
cancellation and lost-COMMIT evidence. Component parsing tests alone do not prove
that delivery. Existing full Memory, native History, HTTP Write, production
adoption and integration review gates remain open.
