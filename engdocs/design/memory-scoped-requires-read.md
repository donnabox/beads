# Scoped Memory Requires read

`internal/storage/graphstore.ReadMemoryRequires` is an internal, read-only operation for one selected local Constraint and its complete bounded Requires-family closure. It is not a CLI command or a BDP result schema. The BDP/BTS query language assessment is separate and on hold.

The reader uses the selected Memory projection's Type bindings and existing catalog affiliation/ownership operations. It checks that Requires-family definitions are Constraint-owned before relying on an owner's retained Links. Defines participates through its Requires affiliation. No new Type engine, schema or legacy Issue evaluator is introduced.

One existing rollback-only transaction supplies current and retained facts, floating-reference resolution and the writer token. Exact references never fall back to latest. Historical owners contribute their own retained Links; current Links are not joined onto old owners. Results retain the authored endpoint pins beside exact resolved witness endpoints, full canonical Bead/Link records, parallel Links, shared descendants and cycles. No satisfaction inference is performed.

## Scope and refusals

This initial operation supports live generic Memory identities whose selected revision affiliates with Constraint, including retained revisions of those live identities. It refuses a missing identity, deleted identity, unavailable revision, native Issue root, plain Memory root, malformed current/retained state, unknown Type binding, cancellation and exceeded bounds. Refusals return no partial result. The caller also checks Store.Close before rendering output.

Generic graph connectivity still includes native Issues and Memory Beads. This operation's Constraint endpoints follow Requires semantics; they are not a restriction on generic mixed paths. Subject To discovery, Evaluation joins, arbitrary scoped traversal and CLI wiring are separate operations.

## Acquisition accounting

SQL-side byte and row checks precede payload hydration. Only the selected subjects and their source-local Link validation feeds are acquired; the globally bounded `CurrentSnapshot` and its limits are unchanged. An unrelated population does not spend this scoped read's byte/node budget.

The byte budget conservatively charges persisted values with copy factors: four for selected catalog rows, two for retained snapshots/current payloads, and three for current Link validation feeds, including their endpoint catalog facts. It is not an exact Go heap or SQL engine memory limit. The fixed workspace/Type binding is checked by existing helpers and excluded from this operation's variable acquisition budget; repeated binding checks remain intentional reuse of those helpers.

The pure traversal's MaxInspected bounds candidates returned by its reader. The storage adapter separately applies that same requested ceiling to all acquired/validated source-local Links, including omitted nonmembers. Thus an optimized candidate set cannot bypass total storage acquisition accounting. MaxNodes and MaxLinks bound the complete exact-reference result. At the store boundary every limit refusal matches graphstore.ErrLimitExceeded; the pure traversal's underlying sentinel is retained as a cause where relevant.

## Compatibility and verification

No existing runtime Go source, dependency-mode mode routing, parser/evaluator, storage driver, default or output writer changes. The added `TestGraphModeCLIMemoryDependencyQueryCompatibility` is unconditionally selectable in the existing C0 graph CLI phase, and pins explicit/default dependency-mode results including the shipping repeated-title overwrite behavior.

Focused tests use real disposable embedded Dolt. An optional released-server variant uses BEADS_GRAPH_TEST_SERVER_PORT; a missing server is a skip, not successful server qualification. The pure core is covered by its own Bazel target. See [the test authority](../TESTING.md) for the full build/lint/provider gates. No local focused pass implies installed Memory command support or a qualified release.
