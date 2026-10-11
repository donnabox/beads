# Graph preview markdown Issue batch (source review; unqualified)

`bd create --file issues.md` accepts a markdown file with one `## Title` per
Issue. Each heading may contain ordinary `### Priority`, `### Type`,
`### Description`, `### Design`, `### Acceptance Criteria`, `### Assignee`,
and `### Labels` sections. The ordinary parser, template lint and native Issue
batch writer are reused. All headings and graph Issue identities land in one
transaction, so a refused item leaves none of the file's Issues behind.
`--json` returns the ordinary bare array of Issue objects with Scope-relative
IDs, preserving the familiar fields rather than using a graph envelope.

This bounded route requires durable Issues with native History. It refuses
`### Dependencies`, ephemeral/no-history storage, infrastructure Issue types,
and any single-Issue flag mixed with `--file`. The ordinary `--graph` JSON plan
remains unavailable. Graph plans can include arbitrary native Dependency kinds,
parentage, metadata references, events and wisps; exposing those without an
atomic canonical Link/owned-History mapping would create native state that graph
reads cannot see. `--file` dependency sections have the same Link-mapping
obstacle. Neither should be implemented as a loop of single-Issue creates,
because ordinary batch create is atomic.

The implementation and test in this source slice have **not** been built or
run. Before release, qualify the exact combined source in installed embedded
and ordinary shared-server modes with no skipped workflow tests, then the
required hosted gates and review. Keep `bd status --graph` and the CLI spec's
integration ledger tied to their actually qualified build until then.
