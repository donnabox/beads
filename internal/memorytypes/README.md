# Fixed Memory contract foundation (candidate)

This package has **no runtime callers**. It is F0 groundwork for Preview 3, not
installed Memory support, BDP conformance, or adoption of the Memory proposal.
There is no storage, network access, schema loader, catalog installation, or
public type ID/selector mapping. It adds no Go dependencies.

The input is Melinda's product proposal at public `gastownhall/bdp` commit
`50b13ee08f79a00bf8b135495b3b4d432f811992`. Earlier reviewed implementations
`1935d820ba1f8bdd42bb1897438f04e34b486627` and
`499c227fa5d0c208db0bf8d63c27e6fae9927483` remain historical evidence against their
own exact candidate pins; see [REVIEW.md](REVIEW.md). Copied test inputs and byte
hashes are listed in `testdata/candidate/provenance.json`.

The product proposal does not require BDP or a Type-definition representation.
Property schemas are review aids. The retained `definitions.json` is a separate
nonnormative BDP exploration, not an authoritative required product format. None
of these files is read by production code. Reconcile fixture, predicate, and test
changes explicitly after proposal review.
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
- Required `title` and `body` supply common content for all six kinds. Empty body
  is valid with a title, including on specializations. `statement`, `choice`,
  `steps`, `rule`, and `takeaway` are no longer declared primary fields; as open-root
  extensions they receive no special interpretation or validation. Optional
  `observedAt`, `alternatives`, and `authority` remain type-specific. Applicability
  has only a required description when present; its closed schema rejects the
  deferred `validFrom`/`validUntil` fields. No interval-ordering engine is supplied.
  Shape validation does not establish the meaning or quality of prose.
- Format checks target the pinned Python schema checker's asserted `date-time`
  and `uri` behavior, with explicit exceptions below. No timezone or observation
  time is invented. The pinned checker rejects leap-second values
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
- `LinkKind.ConformsTo` distinguishes concrete affiliation from the Association
  family: all five supplied specializations conform to directly authorable
  Association. This finite table serves the fixed release catalog, not a
  permanently closed definition of future dynamic conformance.
- `ValidateLinkProperties` applies Association's open property contract, including
  its optional nonblank `note`, to Association and every supplied specialization.
  These six relationship symbols are not workflow Dependencies.
- `ValidateEndpoints` applies each whole source/target pair, identity-based
  self-link rejection, and explicit target pins for Supports and Challenges.
  MemorySupersedes accepts current or explicitly pinned targets; it does not
  redirect, delete, repair, or rewrite incoming links. Accepted endpoints are
  same-scope Issue/Memory family Beads only, according to product §4.2. Other local
  Beads and all external categories refuse. It has no arity, pair-uniqueness, or
  cycle engine. Parallel
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

Cross-scope Beads and other external resources remain distinct categories for
explicit refusal. No fetch, import, availability, or immutability claim is made.
The broader endpoint arms in the exploratory definition graph are not implemented
release behavior. Tests take their acceptance table from the product prose and
cross-check only that accepted product pairs lie within that separate experiment.
External URI values in Policy.authority remain authored property data, not graph
endpoint admission.

Storage integration must preserve complete payloads and historical affiliation,
validate atomic post-state, enforce workflow separation, and provide mutation,
retyping, ownership, history, and provenance behavior. None is implemented here.
Memory-family retyping is the selected Preview 3 cutline; cross-family conversion
remains deferred. An explicit Type-only change with unchanged properties is a
semantic change and must create a revision; repeating the same Type and unchanged
properties is a no-op. The caller must compare affiliation as well as properties.
Successful validation of the same payload under two Types does not make a Type
change a no-op. This package implements neither mutation nor revision allocation.

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

Go tests cover all 64 active pinned property cases, inherited requirements,
nonmutation, diagnostics, extra format and deferred-field cases, and all 600 combinations of
the ten endpoint classifications over six product pair sets. Every common Link
property case exercises all six Link kinds. The independent Python oracle asserts
Draft 2020-12 formats and forbids remote schema lookup. Exact versions of
`jsonschema`, `rfc3987`, and `rfc3339-validator` are pinned and checked. Its
`schemaValid` overrides identify stricter trailing-newline refusals separately
from schema shape.

This is pure helper evidence. It does not qualify a CLI, database, installed
catalog, shared server, BDP profile, or the complete Memory workflow.
