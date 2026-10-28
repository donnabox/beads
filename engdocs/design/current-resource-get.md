# Checked current Resource Get

The public `resourceget` role exposes one complete current Bead or Link read. It
reuses the checked acquisition and projection already used by canonical Show and
the ordinary point HTTP reader. CLI Reference syntax and output formats stay in
their adapters; a library caller receives native Resource data directly.

This increment preserves the implementation's existing BDP pin
`bad780f19761e5714873aa3a85adeb75bc885928`, schema `bdp-v0.schema.json` and version
`0`. Publishing a public role does not adopt a newer protocol, implement an entire
profile or supply a hosted provider. The receiving work is tracked in
[issue7926](https://github.com/gastownhall/beads/issues/7926).

## Operation and values

A `resourceget.Reader` accepts a Scope-local shipping selector or a same-Scope
canonical URL. Local selection tries exact public identity first, then preserves
the source's existing Issue abbreviation and old suffix compatibility rules;
ambiguous selections refuse. The result carries the actual canonical identity.
The reader's checked source supplies its Scope and authority.
An optional kind assertion verifies the result; path spelling does not establish
kind. Request identity data grants no authorization.

The result is an immutable, contract-tagged native document. It exposes actual
kind, canonical ID, revision, Scope and copied native JSON. Its JSON encoding is
the native Resource, without a CLI result envelope. Complete properties,
metadata, captured owned state and whole Link References retain their existing
meaning. Neither input nor returned byte slices can mutate the document.

The public parser validates the admitted current carrier against the explicit
old binding. That validation does not establish source authority, permissions,
Type-policy enforcement or the correctness of a stored revision. The supported
local adapter obtains those facts through the existing checked store/projector;
callers cannot grant themselves access by constructing a JSON document.
Parsing is byte-in-only and requires explicit Scope and kind; generic JSON
unmarshaling refuses. Callers must bound untrusted input. NativeJSON preserves
input bytes, while standard JSON encoding may compact or escape them. The
published SchemaID identifies the wire representation; internal DTOs stay
private. The public role names the operation and carrier, and the internal
projector implements it without exporting its storage representation.

## Existing local source

`resourceget/local` opens an existing, explicitly identified graph workspace.
Callers supply the real canonical workspace path, exact Scope and authority,
database and selected embedded or shared-server connection. It retains the
existing binding, installed generation, realpath and native-state checks. Opening
the reader does not initialize, migrate, repair or relocate a workspace.
This graph source supports only its `main` branch, matching canonical Show.

The handle owns its connection until the caller closes it. Reads use caller
cancellation; Close coordinates with active reads and subsequent reads refuse.
Callers must check read and Close errors before reporting successful delivery.
No renderer receives the store, transaction or connection handle.

A shared-server connection in this local factory is still a checked local
workspace configuration. It is not a hosted tenant/project opener. A service
embedder must authorize its actual principal/project, preserve applicable policy
and credential wrappers, and provide a separately supported source. This change
neither adds a required method to every backend nor unwraps one to raw SQL.

## Compatibility and limits

Only the full-current operation is lifted into this first role. Existing exact
retained reads keep their own path; an exact selector must never be silently
reinterpreted as current content. The role does not parse `@` suffixes, local
revision ordinals or shell quoting. CLI adapters resolve their admitted Reference
syntax before dispatch.

Properties views, incident aggregates, Type-definition lookup, native masks,
History, Query, List and Event capabilities do not follow from supporting this
role. Existing canonical output masks remain renderer operations under the old
contract. A future native projection or selected protocol port needs its own
representation, admission and qualification work.

Existing Issue/detail, watch/many/related and retained Show routes retain their
shipping contracts. Unsupported providers and invalid binding, identity, kind
or native payload refuse honestly. Local errors are not fabricated BDP Problems,
and an output failure does not turn an incomplete delivery into success.

Current canonical Show and HTTP point reads pass the projection through the
checked current carrier. Invalid or incomplete projected native state therefore
refuses before delivery. The adapter retains existing source read budgets and
error classifications; success means complete captured state within those
budgets, rather than partial data at a limit.

## Qualification

Meaningful receiving evidence includes a real out-of-module consumer using only
the public packages, actual current Bead and Link reads from a disposable native
workspace, parity with canonical CLI and HTTP point output, complete edition
checks, malformed/cross-Scope/kind controls, immutable copies and handle cleanup.
A mock, interface assertion, old binary or library build alone is insufficient.

Each implementation report records exact source and executable identities,
executed cases, backend and skipped or unsupported variants. This document is an
API boundary description, not a claim that a particular build or service has
passed that qualification.
