# Fixed Memory contract foundation (candidate)

This package has **no runtime callers**. It is F0 groundwork for Preview 3, not
installed Memory support, BDP conformance, or adoption of the Memory proposal.
There is no storage, network access, schema loader, catalog installation, or
public type ID/selector mapping. It adds no Go dependencies.

The input is Melinda's candidate at public `gastownhall/bdp` commit
`38791e4cc6e16cf661fa839303d4ff315d49ab3d`. Copied test inputs and byte hashes are
listed in `testdata/candidate/provenance.json`. The schema fixtures and definition
literals are nonnormative inputs to tests; production code does not read them.
Reconcile fixture, predicate, and test changes explicitly after proposal review.
`Association` is the proposed replacement name for this candidate's
`memory-related`; existing Dependency `related`, `relates-to`, and `supersedes`
are neither represented nor altered here.

## Contract

- `ValidateProperties` checks a **complete** JSON object for one explicitly
  selected Memory family kind. Child checks always include Memory. It neither
  merges patches nor infers affiliation from fields. Unknown root values are
  allowed without rewriting or dropping them. Nested inception/applicability
  and exact authored reference values are closed as specified. No `kind` field
  is permitted. `Violation` gives a deterministic JSON Pointer and expectation.
- Format checks follow the pinned Python schema checker's asserted `date-time`
  and `uri` behavior. No timezone or observation time is invented. Applicability
  additionally requires `validFrom < validUntil`, normalized across offsets with
  arbitrary fractional precision. The pinned checker rejects leap-second values
  and year zero; the implementation follows that candidate behavior. This is an
  explicit format compatibility edge to revisit with the spec owner, not a claim
  that RFC3339 forbids leap seconds.
- `ConformsTo` is a fixed membership predicate over internal symbols. It does
  not examine schemas, chase types, or treat successive contract revisions as
  automatically compatible. An integration adapter must map the exact applicable
  catalog contract to these symbols. These validators do not select revisions.
- `ValidateLinkProperties` applies the common open informational properties
  contract. These six relationship symbols are not workflow Dependencies.
- `ValidateEndpoints` applies each whole source/target pair, identity-based
  self-link rejection, and explicit target pins for Supports, Challenges, and
  MemorySupersedes. It has no arity, pair-uniqueness, or cycle engine. Parallel
  links and longer cycles are not rejected by these predicates.

## Caller obligations before runtime integration

An endpoint's `CanonicalIdentity` excludes revision selection and must already
be normalized. Local facts must describe the exact selected current or retained
state and its proven contract membership. `Pinned` means an explicit nonempty
revision was supplied, not a revision silently chosen by this helper. The caller
must prove local liveness/retention, visibility, canonical identity, and the
correct contract revision. Passing these predicates does not prove those facts.

Cross-scope Beads and other external resources are distinct opaque categories.
They never acquire verified Memory/Issue membership. Candidate external pair arms
check only category; roles in the prose remain expectations. No fetch, import,
availability, or immutability claim is made. The caller still owns supported
reference syntax and external-carriage policy.

Storage integration must preserve complete payloads and historical affiliation,
validate atomic post-state, enforce workflow separation, and provide mutation,
retyping, ownership, history, and provenance behavior. None is implemented here.
Memory-family retyping is the selected Preview 3 cutline; cross-family conversion
remains deferred. This package does not implement either mutation path.

## Checks

```sh
./scripts/test.sh ./internal/memorytypes
./scripts/test.sh -race ./internal/memorytypes
uv run --no-project --with 'jsonschema[format]==4.25.1' python internal/memorytypes/check_schema_oracle.py
```

Go tests cover all 72 pinned property cases, inherited requirements, nonmutation,
diagnostics, extra format/interval cases, and all 600 combinations of the ten
endpoint classifications over six candidate pair sets. The independent Python
oracle asserts Draft 2020-12 formats and forbids remote schema lookup. Its
`schemaValid` overrides separate interval ordering from schema shape. Dependency
acquisition by `uv` may use the network; validation itself is offline.

This is pure helper evidence. It does not qualify a CLI, database, installed
catalog, shared server, BDP profile, or the complete Memory workflow.
