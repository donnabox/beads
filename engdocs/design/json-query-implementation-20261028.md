# JSON AST Query implementation

Donna selected the private JSON AST as the immediate CLI and protocol request
format on October 28. Human syntax is deferred. Denise owns the separate
language and storage comparison.

## Scope and integration

This additive implementation starts at fork main
`6dd6a81f5c733a0b13f92f787e5e16c91affbb1d`. Janet owns receiving integration
and qualification; Sybil owns the separate canonical CRUD and staged List work.
Query must not change the old Issue expression path, dependency-mode behavior,
existing GET representations, retained wire pins, or published discovery.

The selected request discriminator is `bdp-query-design-1`. Protocol draft
[BDP PR317](https://github.com/gastownhall/bdp/pull/317) records the selected
semantics and bootstrap restrictions. This is an explicitly configured Query
operation, not a claim of full BDP v0.1 service conformance.

See the [candidate guide](query-ast-candidate-guide.md) for request examples,
response shapes, and supported boundaries.

## Implementation layers

1. `internal/resourcequery` owns closed AST decoding, static admission, values,
   lexical bindings, traversal and result formation. It opens no store or
   network connection.
2. `internal/storage/graphstore/query_read.go` captures one bounded coherent
   current graph inventory and authoritative Type families in a checked
   transaction. Failed acquisition, rollback and cancellation return no data.
3. `internal/httpapi/graphread/query.go` projects the captured records into the
   selected Resource shape and calls the shared evaluator.
4. `bd query --ast` reads inline JSON, `@FILE`, or stdin and uses the common
   output renderer. The optional HTTP adapter accepts the same AST inside
   `{"query": AST}` through the existing server's authority/security chain.

No SQL emitter is wired into this path. The first implementation validates and
loads a complete bounded current inventory, then evaluates it in memory. Earlier
SQL-lowering experiments are separate evidence, not performance results for
this Go implementation.

## Contract boundaries

Supported selection forms are scan, explicit floating references, filters,
endpoint navigation, and outgoing/incoming/connected walks. Predicates include
field comparisons, presence, literal substring matching, membership, exact or
family Type tests, Boolean composition and nested any/all. A through predicate
controls expansion; an outer filter only selects results.

Results are complete Resource collections, exact counts, or traversal graphs.
Graphs preserve explicit roots, selected Link editions and actual endpoint
edition mappings. Traversal does not implicitly add induced-only Links.

The initial context is current-only. Explicit pinned acquisition refuses even
if its token equals the current revision. Stored pins remain data and constrain
incidence. No unavailable historical or foreign content is replaced with current
content. Type-family tests require complete authoritative Type facts; there is
no exact-name fallback.

Request, AST, work, allocation and representation budgets are enforced. A limit
never becomes a successful shortened graph, false any, vacuous all, exact count,
or partial page. Unsupported sorting, masks and paging controls refuse rather
than being ignored. Discovery remains unchanged until its common Core and Events
dependencies are implemented.

## Verification completed

- A normal `make build` produced candidate SHA256
  `4fd2cb928050fe79b36161a4667fcc9543f639cd497857d74ba0362c68a5288f`.
- A fresh embedded-Dolt CLI smoke created an Issue, Memory and Reference Link.
  Scan, title selection, count2, outgoing one-hop and separate-process repeat
  reads passed. All10 commands exited0 with empty stderr. Exact receipts are
  published in coordination commit
  `deb5815abeae83d29f4b588a26f5ad1a46155d3b`.
- A normal Bazel run passed49 selected top-level pure/fake tests and138 named
  cases across the evaluator, store, representation, HTTP and CLI packages.
  Both actual Go run/pass events and XML agree; no cases failed or skipped.
  This does not execute the native engine fixtures.
- The managed BUILD metadata generator reached a clean fixed point after two
  BUILD-only repairs. A subsequent documentation runfiles fix exports the
  schema linked by the candidate guide; that successor needs its own check.
- Historical native Codex and actual Claude findings were processed. A new
  shipping review examines the current source; reduced reviewer coverage will
  be stated if a backend cannot complete.

Receipts and exact source freezes live in
`/Users/dbox/.local/state/melinda/query-implementation-20261028/`. Historical
sandbox/metadata experiments remain archived there and are not prerequisites
for the ordinary product build or publication path.

The later hardening successor passed37 affected pure/fake roots and120 named
cases plus formatting. The real embedded HTTP fixture passed its root and six
semantic cases. Actual embedded CLI tests passed the mixed graph root and four
subcases, and a fresh dependency-mode control passed. Raw Go events and XML
censuses match, with no selected failure or skip. Exact commands, source hashes
and captured undeclared-output HTTP receipts are retained in the hardening
subdirectory. The first HTTP attempt failed before engine acquisition because
its receipt path was outside the Bazel sandbox; the corrected fixture uses the
standard undeclared-output directory, with no permission expansion.

## Remaining readiness work

- [x] Close the native shipping finding and test its admission correction.
- [ ] Complete the remaining external review and resolve supported findings.
- [x] Qualify expanded Issue/Dependency/Memory and configured HTTP scenarios on
      a fresh owned graph in Janet's reserved engine window.
- [x] Preserve a dependency-mode legacy-query compatibility proof.
- [ ] Run affected documentation/format checks and hosted Linux repository CI
      against the actual publication commits.
- [ ] Publish the layered implementation PRs with exact evidence and restrictions.
- [ ] Reconcile Query with the receiving integration tree under Janet; no main
      merge, wire pin movement or release is implied by a passing candidate.

PostgreSQL/SQLite product adapters, service-side SQL pushdown, retained Query
acquisition, complete v0.1 discovery and deferred delivery controls remain
separate work. They must not be advertised as installed capabilities.
