# Preview 3 endpoint-pin storage slice

This implementation stores a local informational Link's source and target
identity independently of each optional exact endpoint revision. It is an
internal writer/read slice; CLI endpoint pins remain refused until the adapter
slice lands. Export closure adaptation is separate work owned by its consumer.
It does not adopt a new public BDP contract or enable external endpoints.

## Carrier and writer API

`LinkRecord.Source` and `Target` remain canonical URI strings for identity-based
traversal. `SourceRevision` and `TargetRevision` carry opaque exact tokens.
`SourceReference()` and `TargetReference()` expose the complete references.
JSON encodes floating endpoints as strings and pinned endpoints as closed
`{uri, revision}` objects, including retained and owned Link snapshots and
HTTP graph-read output. No suffix syntax enters stored identity strings.

Create accepts `SourcePin` and `TargetPin` (`EndpointPin`). On update a nil
pointer preserves the existing pin; a zero value explicitly floats it; `Current`
resolves an exact revision inside the writer transaction. `Revision` selects an
exact retained Bead state. Unknown tokens refuse and terminal states do not
qualify as live endpoint states. Both Current and Revision together refuse.

Pin-only update preserves Link identity, Type, endpoint URI identities,
properties and metadata. Link and current-source guards remain independent of
endpoint selection and run before no-op determination. Equal complete
references do not mint a revision. Effectful owned-target re-pins atomically
revise the Link and its current owner's snapshot; old retained bytes remain
unchanged. Ordinary owner revision recording keeps its existing attribution
behavior. Native Dependency writes remain in their existing domain writer.

Donna selected owned-source pins as deferred for the first slice. Both current
and historically selected owning Types refuse pinned-source writes, including
a historical Evaluation after its current identity has been retyped. No
pre-write or post-write source observation point is implied.

The installed Type sidecar validates endpoint facts from selected historical
states and applies its existing correlated alternatives, ownership, counts and
cycle rules. There is no second Type-validation engine. Current Link reads
return stored references without implicitly reading endpoint content. Writer
admission and the complete Type-validation pass reserve bounded historical
payload bytes before hydration, charging repeated references per Link.

Revision comparison reports complete source/target reference changes in the
`endpoints` area. Owned comparison keeps URI identity checks while comparing
the complete pinned snapshot bytes.

## Fresh workspace boundary

The new durable pin columns require graph schema version **8**. Existing
version-7 workspaces are refused; this slice does not migrate or reinterpret
them. The separate installed Type marker is unchanged.

## Focused acceptance

- Both embedded and server Dolt: persistence/reopen, retained old Link bytes,
  pin/unpin/omission/no-op, missing exact revision, stale independent guards.
- Owned-target re-pin: old owner bytes retained and rollback at all six writer
  stages, selected floating-only SupersededBy refusal.
- Historical payload census refuses an over-budget population without partial
  commit; source-owned pin deferral also survives current retyping.
- HTTP graph-read transports both complete references independently.
- Comparison reports pin changes and accepts owned pinned snapshots while
  refusing endpoint identity retargeting.

See the named `EndpointPins`, `OwnedTargetRepin`, `OwnedSourcePins`,
`PinnedEndpointValidationBudget`, and `VersionComparisonEndpointRepin` tests in
[graphstore](../internal/storage/graphstore/endpoint_pins_test.go) and the
[HTTP test](../internal/httpapi/graph_endpoint_pins_test.go).
