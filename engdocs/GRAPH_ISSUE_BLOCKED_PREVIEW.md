# Dependency-blocked Issues in the graph preview

In a disposable workspace initialized with `bd init --graph-mode link`,
`bd blocked` shows Issues whose work is blocked by Dependencies. This read-only
adapter uses the existing Issue engine query. It makes the current graph
workflow inspectable: list work, identify its prerequisites, close a prerequisite
or unlink a Dependency, then inspect `bd blocked` and `bd ready` again.

Dependency-blocked is **not the complement of ready**. A manually chosen
`blocked` status, deferred work, and a claim lease do not by themselves create
Dependency Links. The adapter preserves native query membership instead of
inventing additional scheduling rules. Informational Links and Memory Beads do
not make work dependency-blocked.

```sh
bd blocked
bd blocked --json
bd blocked --readonly --json
```

The human output uses quoted canonical Issue IDs, priority, title, and quoted
canonical blocker IDs. Quotes escape embedded newlines. The empty view prints
`Dependency-blocked Issues (0; graph preview)`. `--quiet` suppresses human output;
`--quiet --json` still returns the structured result.

`--json` uses the existing experimental graph envelope:

```json
{
  "schemaVersion": 1,
  "preview": true,
  "result": [
    {
      "issue": "complete existing IssueRecord object",
      "blockedBy": ["https://example.test/project/beads/prerequisite"]
    }
  ]
}
```

The string above stands in for the complete object to keep the illustration
short. `issue` is the same checked current record returned by graph `show`,
including its complete owned Dependencies and saved version token. Its inherited
`properties.id` remains the native Issue ID, as it does in existing graph
show/list output. The wrapper's `blockedBy` and human presentation use canonical
graph IDs. This slice does not rewrite the Issue property schema. Empty results
are `[]`, never `null`. Rows and blocker IDs sort by their canonical
percent-encoded ID spelling (using the shared code-unit comparator). This
presentation order makes no priority or history ordering claim.

The command accepts no positional selectors or filtering options. Explicit
`--parent`, `--label`, `--label-any`, and `--exclude-label` refuse with
`capability_unavailable`, including empty values. A positive `BEADS_MAX_ROWS`
also refuses; unset it or use zero for a complete view. Malformed/negative values
refuse with `invalid_properties`. Ordinary non-graph `bd blocked` is unchanged.

Admission, complete current records, and the unchanged native
`GetBlockedIssuesInTx` query share one read transaction. The adapter first uses
the existing current-inventory checks for workspace binding, installed Types,
current/retained authority, complete mapping, and local Dependency endpoints.
Unsupported wisp authority refuses, even if the native query would omit it;
absent optional native wisp tables follow the existing missing-table rule.
Unmapped or invalid records refuse the entire result through that admission,
with defensive native-result invariant checks afterward. The preview admits
unique local blocks-only pairs; duplicate native blockers refuse rather than
being silently deduplicated into an apparently supported result. Invalid-store
and unsupported wisp-authority errors retain the existing `graph_not_initialized`
code (exit 5); there is no new dedicated corruption or wisp error code. Authority
admission precedes wisp inspection. This command never repairs
blocking state, creates a version, changes a lease, or records an event.

The current inventory is limited to 1,000 live Resources and the existing 16MiB
current-read acquisition budget. These limits apply across the workspace,
including unrelated Memory content, even if no Issues are blocked. Output has a
separate 16MiB complete-or-refuse bound. These operational preview limits are not
heap guarantees, pagination, or public BDP contracts. `status --graph` advertises
`issueBlocked`, `issueBlockedInventoryResources`, and `issueBlockedOutputBytes`.
The inventory limit names the effective shared current-snapshot limit for this
command, not a separate allocation allowance. `issueWorkflows`, full Memory,
and full History remain false.

Source tests cover the two-prerequisite lifecycle, reopen/unlink, complete records
and owned Links, empty views, verified manual status/deferred/lease distinctions,
Dependency-bearing in-progress/closed/pinned subject membership, stable
ordering, corruption/refusal without repair, cancellation, and acquisition limits.
The ordinary-server concurrency case forces a prerequisite close to commit after
a reader establishes its snapshot; the read must remain wholly before that
commit. Embedded concurrent callers retain the production one-connection pool
and may observe either complete state. Source tests are not an installed CLI
receipt or a qualification claim. This preparation is based on the fixed Link
successor; complete qualification on its final source is required before it
can advance the integration candidate. A draft PR alone does not satisfy that gate.

Parent-child inheritance and other native dependency kinds remain outside the
admitted blocks-only graph boundary; their authoring would require review before
the native helper's parent fallback could become graph behavior.

Public HTTP serving, authentication, user-installed Type policy, labels writer
adoption, full Memory representation, aliases and full History are outside this
adapter. Contributor preflight identified relevant native blocking and repair
work; this adapter imports none of those repairs and leaves their ownership and
semantics unchanged.

Adjacent `bd ready` still has its existing unquoted human layout and no matching
output-size gate. Aligning that presentation is a separate follow-up; this slice
does not change its behavior.
