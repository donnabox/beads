# Due dates in the graph preview

This preview connects the ordinary Issue deadline fields to graph create,
guarded update, and explicit Issue listing. It is a bounded adapter awaiting
its own real-engine and installed-command qualification; the presence of the
flags alone is not that evidence. `issueWorkflows` remains false.

```sh
bd create 'Review the plan' --id beads/review --due '2030-01-02T09:00:00-08:00' --json
bd list --format records-json --due-before '2030-01-03' --all --limit 0
# Copy the selected Issue's version from the record above.
bd update beads/review --due '+6h' --if-revision OBSERVED_VERSION --json
bd show beads/review --version OBSERVED_VERSION --json
bd list --flat --overdue
bd update beads/review --due '' --if-revision CURRENT_VERSION --json
```

An omitted due date leaves the existing value unchanged on update. An explicit
empty update clears it; empty or omitted create leaves the date absent. Due
edits may be combined with the other admitted Issue fields under the same graph
revision guard or explicit `--unconditional`. A stale guard refuses even a
no-op. These edits do not close, reopen, claim, wake, or reschedule other work.

Input uses the existing `ParseRelativeTime` parser: compact offsets, natural
language, date-only, and RFC3339 forms. Dates without an offset use the frontend's
local timezone, and parsing yields UTC. Existing calendar arithmetic is retained.
The CLI passes the typed instant to storage without its own rounding. The
existing database column stores whole seconds. The storage adapter converts to
UTC and rounds to the nearest second before equality and writing, with exact
half-second ties rounding upward; the resulting year must be 1–9999. Years below 1000 are verified Dolt behavior,
not a claim about every MySQL implementation; the adapter checks each stored
value before publication. Native
engine controls pin this representation, while the complete graph workflow
still requires its own qualification. Fractional-second preservation is not
promised. This preview does
not introduce a schema change or a new date language.

`--due-before` and `--due-after` use the ordinary strict-before/strict-after
predicates. The adapter normalizes API-supplied bounds to UTC; CLI parsing
already performs that conversion. Native query bounds are formatted at whole-second precision
(fractional digits are dropped, not rounded as on writes). An empty filter adds
no predicate. `--overdue` uses the native
current-time query and excludes closed Issues, together with the normal list
status/type/limit policy. All filters intersect. These are current reads, not
historical clock queries, and they do not mutate deadlines or priority.
`--sort=due`, tree output and legacy Issue JSON remain unavailable; choose
`--flat` or `--format records-json`. The flat summary remains the existing
compact view; complete records and `show` carry the due date.

Invalid due values on create/update produce graph validation errors. Invalid
list-date syntax retains the ordinary list parser's stderr diagnostic and exit
1; this slice does not promise a new stable error envelope for it. Graph date
text must be valid UTF-8 and at most 4096 bytes. Readonly and migration-freeze
policy still precede write input and store opening.

The advertised preview capabilities are `issueDueDate` and `issueDueFilter`.
Saved exact versions and comparison cover the retained Issue and its owned
Dependencies. This does not deliver full public History or BDP writes. The
existing BDP Read service can expose the Issue's stored properties after a CLI
write; it is not a remote deadline-writing endpoint.

Contributor boundaries remain explicit: #5718 owns shared error-hint improvements;
#5771 fixes bug #5765 through shared optional-timestamp UTC normalization. Required deadlines and
reasoned clearing (#6531), recurrence/calendar-policy changes (#6530), due sweeps
and escalation (#6528), and backfill (#6532) are separate proposals. Defer and
its lazy wake path remain unavailable in the graph adapter. Reconciliation of
shared retained-recorder ownership (#6650) remains an integration gate.
