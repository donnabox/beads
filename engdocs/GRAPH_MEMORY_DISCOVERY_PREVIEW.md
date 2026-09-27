# Compact Memory discovery preview

This disposable graph preview adds a discovery-to-recall loop on embedded and
ordinary shared-server Dolt. It is a partial implementation of Memory proposal
[R6/R8](https://github.com/gastownhall/beads/issues/5877), using the search shape
under review in [CLI proposal §10.2](https://github.com/gastownhall/beads/issues/6703).
It does not establish the complete Memory record or a final compatibility contract.

```sh
bd memories 'deployment'
bd memories 'deployment' --details
bd memories 'deployment' --format records-json
bd recall 'https://example.invalid/project/beads/plan' --version 'TOKEN_FROM_RESULT'
```

Search is literal Unicode case-folded substring matching against current Memory
titles and bodies. It does not trim, normalize, stem, rank or interpret regular
expressions. Each Memory appears once, ordered by canonical ID using UTF-16 code
units. Results identify the matching fields. Empty or omitted search lists current
Memories. Keys are not installed in this workspace generation and are not searched.

Ordinary output and `--details` are summaries. They include canonical ID, title,
saved version and exactly recorded attribution. A body match adds an excerpt
around the first match, limited to 160 Unicode code points including any ellipsis,
with its field and shortening status. A matching body of at most 160 code points
appears verbatim in `excerpt.text`. Title-only and queryless results omit body
excerpts. Details add the outgoing owned-Link count, never target bodies or Link
payloads. The human result gives a shell-quoted exact recall command. Recorded
attribution time remains an observed wall-clock value, not native commit order.

The ID and version identify the saved state that matched. After another process
edits the Memory or its owned Links, that selector still recalls the old saved
body. Discovery itself does not create a version or perform recall automatically.

## Complete bounded results

The default accepts at most 50 matching Memories. More matches cause a `capability_unavailable` error (exit 5)
before any stdout, with guidance to narrow the search or use `--all`. Explicit
`--all` returns all matching summaries within the preview bounds; it does not
expand bodies. A successful empty result means no matching Memory was found in
the complete admitted snapshot.

This complete-or-refuse behavior is provisional. It does **not** implement the
proposal's paged default or continuation tokens. `--limit`, `--after` and full
record output are unavailable. No process-local HTTP cursor is presented as a
cross-process CLI cursor.

The existing current reader admits at most 1000 live Resources and conservatively
bounds acquisition to 16 MiB for the entire workspace, before filtering Memories.
Unrelated Issues or Links can therefore cause even a narrow query to refuse.
Authority and corruption checks are unchanged. Search input is bounded to 4096
UTF-8 bytes. The complete rendered response is bounded to 1 MiB; excess refuses
atomically instead of cutting a record, title, identity or JSON document. A very
large title can exceed that output bound even with `--all`. Narrow the search to
exclude it, or read that Memory directly with `show` or `recall`. Acquisition and
output limit refusals use `capability_unavailable` (exit 5).

## Structured summaries and compatibility

`--format records-json` returns the existing experimental wrapper:

```json
{
  "schemaVersion": 1,
  "preview": true,
  "result": {
    "projection": "summary",
    "scope": "https://example.invalid/project/",
    "items": [],
    "next": null,
    "complete": true
  }
}
```

Each item contains `id`, `type`, `title`, `version`, `attribution`, and
`matchedFields`. Body matches add `excerpt: {field, text, truncated}`. `--details`
adds `details: {ownedLinkCount}`. Queryless items have an empty `matchedFields`
array. There is no body field, owned record payload, metadata value, Project
mapping, key, fabricated Inception or derivation record. This is neither full
structured recall nor connected interchange/export.

Selecting `--format records-json` also selects typed JSON diagnostics on stderr,
including admission failures; refusals produce no partial stdout.

Bare `--json` and `--format legacy-json` are refused in graph workspaces while
Memory's complete-record requirement and the CLI's legacy map compatibility are
reconciled. The existing `--format json` alias (case insensitive) has the same
restriction in graph workspaces. An explicit `--format table` or
`--format records-json` overrides a configured `json: true`; an explicit `--json`
still refuses, including when combined with either format. With no explicit
format, configured JSON output also refuses.

Existing legacy `memories --json` and `memories --format json` keep their
key-to-body map. The new `--all`, `--details` and summary-format options refuse on
legacy workspaces before opening storage. There is no stored option that silently enables body expansion.

`status --graph --json` reports `memoryDiscovery: true` and
`memoryDiscoveryPagination: false`, plus the limits above. Full Memory, structured
Memory recall and public History remain unavailable. HTTP Read is unchanged;
this command does not add a public BDP search profile. Prime/onboarding guidance
and the complete generic exposure audit remain separate work.

## Qualification

The installed harness uses normal initialization and CLI writes; it never seeds
schema or fixtures:

```sh
python3 scripts/graph-memory-discovery-smoke.py --bd /absolute/path/to/bd \
  --backend both --server-port 3307 --output-dir /absolute/new-receipts
```

It exercises compact search, Unicode matching, exact saved-state recall after
edits, summary-only details, empty bodies, refusal and complete-result bounds,
readonly use, and preserved legacy output. The server leg adds bounded concurrent
process reads/writes on one already provisioned database. Different-database
provisioning remains serialized for Dolt 2.1.8. Existing real snapshot tests supply
forced transaction-overlap, authority, corruption and acquisition-limit evidence.
Actual qualification receipts and limitations are recorded in fork PR18.
