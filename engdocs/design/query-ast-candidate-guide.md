# JSON AST queries

This implementation candidate adds `bd query --ast` to an existing graph-mode
`link` workspace. It is separate from the older Issue query language. The AST
can select Beads or Links, filter their fields, follow endpoints and traverse
Links. It can return Resources, an exact count, or a traversal graph.

This is an unreleased contributor guide. A normal development build and a
bounded embedded-Dolt CLI smoke passed. Expanded native Dolt and focused
Bazel checks also passed; full hosted repository qualification remains open.
The pinned-release user docs and installed binary have not changed. This does
not claim complete BDP v0.1 service support.

## Supply one request

```sh
bd query --ast '{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"scan","collection":"beads"},"result":{"form":"count"}}'
bd query --ast @query.json
bd query --ast - < query.json
```

Each invocation accepts exactly one AST document. Do not combine it with a
positional Issue expression or the older selection and sorting flags. This
command reads an existing workspace; it does not initialize or migrate one.

For example, save this request as `query.json` to count current Beads whose
`status` property equals `draft`:

```json
{
  "ir": "bdp-query-design-1",
  "profile": "list",
  "selection": {
    "op": "filter",
    "input": {"op": "scan", "collection": "beads"},
    "bind": "b",
    "predicate": {
      "op": "compare",
      "operator": "eq",
      "left": {"op": "field", "of": {"op": "binding", "name": "b"}, "path": ["properties", "status"]},
      "right": {"op": "literal", "value": "draft"}
    }
  },
  "result": {"form": "count"}
}
```

There is no implicit Issue-only or open-status filter on this AST path. Without
`--ast`, `bd query` retains its existing Issue expression behavior.

## Read the response

JSON through the common CLI envelope is the default. To receive the native
`result` and `evaluation` object directly:

```sh
bd query --ast @query.json --omit envelope
```

The common `--output-format detail` and explicit `--output-format default`
select the existing human detail renderer. Omitting `--output-format` selects
JSON for this command. Raw envelope omission requires JSON.

`result.form` distinguishes `resources`, `count` and `graph`. `evaluation`
records completeness and traversal boundaries. A work limit is a refusal; it
never becomes a successful partial count, shortened graph or empty selection.
Resource results contain their full selected records and `next: null` in this
first atomic slice.

Field projection belongs in the AST. `--include` and other `--omit` selectors
are not an alternate projection language for this path. Masks, paging and
sorting explicitly refuse until their implementation is available.

Diagnostic and traversal JSON Pointers address the normalized request
`{"query": AST}`. For example, `/query/selection` addresses the submitted
selection after that envelope is added. This is the same envelope used by the
BDP POST operation; it is not a second CLI AST format.

## Initial support boundary

Execution uses one checked current graph context, including Memory, Issues and
dependency-backed Links. Exact revision acquisition and retained incidence are
unsupported. Stored pins remain queryable data and constrain current incidence;
the evaluator does not silently replace a pinned edition with current content.
Type-family tests require complete authoritative Type interpretation.

`bd status --graph --json` reports the local Query capability and its request,
AST, evaluation, allocation and mapped-input budgets. Mapped evaluator input is
limited to16MiB before defensive validation retains further copies. The CLI read
has a30second deadline. The configured HTTP handler admits one complete Query
execution at a time; excess concurrent requests receive503 temporarily-unavailable
with `Retry-After: 1`.
These are operational limits, independent of traversal depth.

The configured BDP operation accepts `{"query": AST}` at its Scope root using
`POST S?readVersion=1&operation=list` or `operation=query`, with
`application/vnd.bdp.v0.1+json`. The configured handler has passed real-store
qualification through its security chain and an HTTP recorder; this does not
qualify a socket/listener deployment. Generic v0.1 discovery remains unavailable
until the service's required Core capabilities are implemented.

The implementation's selected AST schema is
[query.schema.json](../../internal/resourcequery/query.schema.json). Exact source
revisions and qualification results belong in the implementation evidence;
this guide does not substitute for them.

## Recorded development proof

The normal `make build` route produced candidate SHA256
`4fd2cb928050fe79b36161a4667fcc9543f639cd497857d74ba0362c68a5288f`
from receiver `6dd6a81f5c733a0b13f92f787e5e16c91affbb1d` plus its exact working
source. Real CLI commands created one Issue, one Memory and their Reference Link
in a fresh private embedded-Dolt workspace. AST scan returned both Beads, a title
predicate returned only the Memory, count returned2, and depth1 outgoing traversal
returned the authored Link. Separate-process count/traversal repeats returned
identical JSON; all10commands exited0 with empty stderr and clean owned groups.

[Exact ASTs, commands, environment, raw JSON and custody evidence](https://github.com/donnabox/agent-coordination/tree/deb5815abeae83d29f4b588a26f5ad1a46155d3b/context/melinda/json-query-20261028/evidence/normal-build-and-cli-smoke-1)
include the initial manual validator's field-name error separately. The product
correctly emitted `resourceKind`; only the validator had used `resource_kind`.
The retained scan was checked offline and remaining reads completed without
replaying fixture writes or changing product code.

This proof is limited to the fresh current-read embedded CLI path. It does not
qualify configured POST, shared-server execution, Dependency or all Memory
scenarios, retained revisions, PostgreSQL/SQLite adapters, large-graph latency,
normal Bazel/lint/test gates or release readiness. The experimental SQL lowerer
and its measurements are separate from this in-memory product evaluator.

## Expanded hardening proof

The corrected source passed37 affected pure/fake top-level tests (120 named
cases), plus formatting, through normal Bazel. A separate embedded-Dolt handler
fixture passed all six semantic cases: mixed inventory and Type families, through
versus outer selection, traversal graphs and native Dependencies, unchanged legacy
GET/discovery, refusals/cancellation/reopen, and the configured authentication/Host/
admission/deadline chain. The handler proof uses a recorder; it does not qualify
a network listener.

Actual CLI subprocess tests also passed on a fresh embedded graph: an
Issue→Memory→Issue→Memory path, native Dependency reads, exact counts, graph
edition mappings, error delivery, and close/reopen. A separate fresh dependency-mode
control preserved the old no-AST Issue query and refused AST mode. No selected
cases failed or skipped. These are development qualification results, not a
release or PostgreSQL/SQLite support claim.

The later [shipping correction evidence](https://github.com/donnabox/agent-coordination/tree/ad0fc168940c9841d48c836cd47023cde5e51748/context/melinda/json-query-20261028/evidence/daytime-hardening-2)
qualifies additional embedded mixed-graph, configured HTTP, capability and legacy
compatibility cases. It records the exact runtime source separately from this
contributor-prose update. Hosted receiving-tree CI remains a landing gate.
