# Fixed Memory contract foundation (candidate)

This package has **no runtime callers**. It is F0 groundwork for Preview 3, not
installed Memory support, BDP conformance, or adoption of the Memory proposal.
There is no storage, network access, schema loader, catalog installation, or
public type ID/selector mapping. It adds no Go dependencies.

The input is Melinda's candidate at public `gastownhall/bdp` commit
`d6337ab96e93d1b9a7b93beb4aea6d48bd179311`. Its copied JSON inputs are byte-identical
to the reviewed successor `0b151ac9f5cadf4d92f76d24134d48edaf1f0e0e`.
The historical implementation at `1935d820ba1f8bdd42bb1897438f04e34b486627`
used candidate `38791e4cc6e16cf661fa839303d4ff315d49ab3d`.
See [REVIEW.md](REVIEW.md) for the explicit reconciliation. Copied test inputs and byte hashes are
listed in `testdata/candidate/provenance.json`. The schema fixtures and definition
literals are nonnormative inputs to tests; production code does not read them.
Reconcile fixture, predicate, and test changes explicitly after proposal review.
`Association` now corresponds to the candidate's `association` identity;
existing Dependency `related`, `relates-to`, and `supersedes`
are neither represented nor altered here.

## Contract

- `ValidateProperties` checks a **complete** JSON object for one explicitly
  selected Memory family kind. Child checks always include Memory. It neither
  merges patches nor infers affiliation from fields. Unknown root values are
  allowed without rewriting or dropping them, including uninterpreted `kind`
  and `inception` extensions. Neither selects affiliation nor replaces required
  fields. Inception is no longer a Memory property contract; this package supplies
  no shared provenance validation. Nested applicability and exact authored
  authority references remain closed as specified. `Violation` gives a
  deterministic JSON Pointer and expectation.
- Format checks target the pinned Python schema checker's asserted `date-time`
  and `uri` behavior, with explicit exceptions below. No timezone or observation time is invented. Applicability
  additionally requires `validFrom < validUntil`, normalized across offsets with
  arbitrary fractional precision. The pinned checker rejects leap-second values
  and year zero; the implementation follows that candidate behavior. This is an
  explicit format compatibility edge to revisit with the spec owner, not a claim
  that RFC3339 forbids leap seconds. Go deliberately rejects timestamps and URIs
  containing a trailing newline that the Python format checkers mistakenly accept.
  Named regression cases mark this stricter result with `schemaValid: true`;
  the implementation does not reproduce that checker bug.
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
be normalized in one identity namespace across local and external classes.
Changing classification/locality cannot create a new identity or bypass self-link
rejection. The caller must prove Bead endpoints identify actual Beads, never Links
classified as `OtherBead` or opaque cross-scope Beads. Local facts must describe the exact selected current or retained
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

The combined, fail-closed package acceptance entrypoint runs Go tests and the
independent oracle with no oracle network access:

```sh
internal/memorytypes/check.sh
internal/memorytypes/check.sh -race -cover
```

It requires `uv` and cached dependencies. To acquire them explicitly before running
the offline gate (this preparation command may use the network):

```sh
uv run --no-project --with-requirements internal/memorytypes/oracle-requirements.txt python internal/memorytypes/check_schema_oracle.py
```

Ordinary Go tests do not invoke Python, `uv`, or dependency downloads. Missing
oracle dependencies cause the combined entrypoint to fail; they never produce a
silent qualification skip. Wiring this entrypoint into repository CI remains an
explicit runtime-adoption follow-up outside F0's package-only scope.

Go tests cover all 60 active pinned property cases, inherited requirements,
nonmutation, diagnostics, extra format/interval cases, and all 600 combinations of
the ten endpoint classifications over six candidate pair sets. Every common Link
property case exercises all six Link kinds. The independent Python oracle asserts
Draft 2020-12 formats and forbids remote schema lookup. Exact versions of
`jsonschema`, `rfc3987`, and `rfc3339-validator` are pinned and checked. Its
`schemaValid` overrides identify interval ordering and stricter trailing-newline
refusals separately from schema shape.

This is pure helper evidence. It does not qualify a CLI, database, installed
catalog, shared server, BDP profile, or the complete Memory workflow.
