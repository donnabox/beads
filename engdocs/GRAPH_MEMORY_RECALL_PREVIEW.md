# Memory body recall preview

A disposable graph workspace can recall a Memory's current body or a saved
version's body through the installed CLI:

```sh
bd recall beads/plan
bd recall beads/plan --version SAVED_TOKEN
bd recall https://example.invalid/project/beads/plan --version SAVED_TOKEN
```

Use the actual canonical local ID and opaque version token returned by `remember`,
`show` or `update`. Foreign Scope URLs, aliases, ordinal and as-of selection are
not supported. Normal graph initialization and Memory authoring are described in
the [editing guide](GRAPH_MEMORY_UPDATE_PREVIEW.md).

The command writes complete UTF-8 body bytes to stdout without a header or an
added newline. An empty body succeeds with zero bytes. Whitespace, CRLF, Markdown
and terminal control characters remain content; recall does not render or sanitize
them. `--quiet` suppresses no content, and `--readonly` permits recall. Diagnostics
go to stderr. The selected read and storage cleanup finish before body output
starts; an output-device error can still leave a partial stream and returns failure.

Both current and historical reads validate the workspace authority and the stored
record. Historical recall never substitutes current properties or owned Links.
A missing Memory is `not_found`; an unavailable token is `revision_unknown`;
corrupt retained data refuses. Healthy Issue and Link records are unsupported
Memory selections, not empty successful bodies. Read acquisition and token limits
are the existing [exact-version limits](GRAPH_HISTORY_EXACT_PREVIEW.md); exceeding
one refuses rather than returning truncated content.

`--json` explicitly refuses with `capability_unavailable`: the proposed complete
Memory JSON form requires Inception, derivation and common metadata that this
preview does not provide. `bd show beads/PATH --json` still exposes the existing
experimental record. Status separates `memoryBodyRecall: true` from
`memoryJSONRecall: false`, complete Memory and public History.

Legacy keyed-memory recall keeps its existing behavior outside graph workspaces.
Using the graph-only `--version` there refuses before opening the legacy store.
This increment adds no schema, stored field, writer, ordering, public HTTP History,
compare or restore behavior. It implements the body-stream portion of
[CLI proposal §10.2](https://github.com/gastownhall/beads/issues/6703), not the whole
Memory proposal.

Run the installed proof against a caller-owned disposable ordinary Dolt server:

```sh
python3 scripts/graph-memory-recall-smoke.py --bd /absolute/path/to/bd \
  --backend both --server-port 15440 --output-dir /absolute/path/to/receipts
```

The harness uses normal CLI authoring and fresh processes, checks raw output bytes,
and runs storage modes sequentially. Provision different databases serially on
Dolt 2.1.8. Review and qualification remain in [fork plan18](https://github.com/donnabox/beads/pull/18).
