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

The current four signed DRAFT layers are [350 evaluator](https://github.com/donnabox/beads/pull/350),
[351 snapshot](https://github.com/donnabox/beads/pull/351),
[352 reader](https://github.com/donnabox/beads/pull/352), and
[353 CLI/HTTP](https://github.com/donnabox/beads/pull/353).

- A normal make-built predecessor passed ten actual embedded-Dolt CLI commands.
  Its private candidate and exact source are separately preserved; PATH was not
  replaced. The earlier canonical pure/fake cohort passed49roots/138named cases.
- Shipping Codex and actual Claude reviews were processed. The final source
  closure found no remaining actionable finding. Gemini could not access its
  headless tool, so substantive three-provider review is unavailable.
- A changed-source run exposed a real CLI initialization cycle. Its failure is
  preserved; the evaluator, reader and HTTP cases passed45roots/118named. The
  corrected CLI then passed7roots/36named cases plus formatting. The actual
  evaluator bottom passed its new Unicode test and14subcases before publication.
- Final native checks passed configured embedded HTTP7named, actual mixed CLI
  5named, the existing closed-world capability guard, and legacy dependency-mode
  Query. All131captured artifacts matched their hashes and raw Go/XML census;
  no selected case failed or skipped. The32HTTP JSON receipts include successful
  store closure. HTTP used the actual configured security chain and a recorder;
  socket/listener behavior is a separate proof.
- Janet independently accepted the final source/custody handback. All owned
  engine windows are released. Runtime evidence binds exact top
  `d43a9ff4b66fa2ae00133bddf5007c66fea08ee4`; this later verification-prose update
  changes no Go, BUILD, schema, imports or contract.

[Exact source, reviews, commands, raw results, failures and independent census](https://github.com/donnabox/agent-coordination/tree/ad0fc168940c9841d48c836cd47023cde5e51748/context/melinda/json-query-20261028/evidence/daytime-hardening-2)
are immutable. Earlier sandbox and metadata experiments remain historical,
separately source-bound evidence; no old failure has been relabelled as a pass.

## Remaining readiness work

- [x] Process supported shipping review findings and qualify their corrections.
- [x] Publish four ordinary signed layered DRAFT PRs.
- [x] Qualify mixed Issue/Dependency/Memory, configured HTTP, capability reporting
      and dependency-mode compatibility on a fresh owned embedded graph.
- [ ] Pass required hosted Linux CI on each actual receiving head/base. Upper
      branch-base PRs require sequential retargeting; absent checks are not green.
- [ ] Reconcile the canonical evaluator seam with Sybil's staged List work and
      validate the resulting integration tree under Janet before main landing.

PostgreSQL/SQLite product adapters, service-side SQL pushdown, retained Query
acquisition, complete v0.1 discovery and deferred delivery controls remain
separate work. They must not be advertised as installed capabilities. No main
merge, release or runtime pin movement follows from these bounded passes.
