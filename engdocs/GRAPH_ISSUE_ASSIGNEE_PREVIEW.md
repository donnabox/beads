# Experimental graph Issue assignment

Canonical graph Issues accept the existing `bd update --assignee` option:

```sh
bd update beads/work --assignee alice --if-revision "$revision" --json
bd update beads/work --assignee '' --if-revision "$revision" --json
bd update beads/work --assignee alice --title 'Assigned work' --priority P1 --if-revision "$revision" --json
```

Use the complete graph revision returned by `bd show`. Assignment can share one
transaction and one retained version with admitted text and priority edits.
Omitting `--assignee` preserves it; an explicit empty value clears it. Values
retain their literal spelling, including spaces, under the existing 255-code-point
field bound. An exact no-op keeps the revision, after checking the supplied guard.

This is ordinary assignment, not an atomic claim. It does not move an Issue into
`in_progress` or arm a lease. The existing Issue writer still refuses taking
another actor's active assignment, even with an exact graph revision or explicit
`--unconditional`. That refusal is reported as `constraint_violation`, exit 4.
Existing holder identity and configured-pool rules still apply. Accepted transfers
clear an existing lease; exact no-ops preserve it. Actor-equivalent spellings may
pass the ownership fence while changing the stored bytes and clearing a lease.
Clearing an active assignment leaves its status `in_progress`, with no assignee
or lease. This preview does not expose status editing or unclaim; do not assume
that clearing the assignee returns the Issue to `open`.

Mixed `--claim` plus assignment, `--force`, `--if-assignee`, `--if-status` and graph
status editing remain unavailable. The separate [claim preview](GRAPH_ISSUE_CLAIM_PREVIEW.md)
admits standalone `--claim`. Graph listing accepts the existing `--assignee` (`-a`) and `--no-assignee`
filters with explicit `--flat` or `--format records-json` output.
Complete current/exact record reads retain assignment along with other Issue
properties and owned Dependencies. It does not advertise public ordered History.

```sh
bd list --format records-json --assignee alice --all
bd list --format records-json --no-assignee --all
```

These reuse the ordinary list query. Empty `--assignee` and explicit
`--no-assignee=false` add no predicate. Nonempty `--assignee` together with true
`--no-assignee` is an intersection, normally empty; it is not a union or a flag
conflict. Input is not trimmed or interpreted as an actor alias. SQL equality
inherits the Issue column's collation, verified here on pinned Dolt 2.1.8;
no cross-provider byte-equality guarantee is added. Repeating `--assignee`,
including mixed short/long spellings or empty values, refuses rather than
silently retaining the last value. The graph-only occurrence check leaves
ordinary flag registration and contributor parser work unchanged.

The `issueAssigneeUpdate` and `issueAssigneeFilter` capabilities and graph admission
of these existing flags are provisional preview choices. `issueWorkflows` and
complete Memory remain false.
The draft starts at qualified `deff78f6`, independently of the held label-writer
and graph-label drafts. It changes no ordinary writer, lease, schema or ownership
policy. Integration must reconcile future retained-recorder ownership once, as
with the earlier Issue adapters. Jim's exact target remains pending.

Contributor boundaries: upstream PR5349 and PR6501 own claim-plus-assignee override
corrections; PR4697 and PR4715 own broader ownership fencing and tokens. PR5739 owns
ordinary repeated-assignee validation; PR6548 owns the broader filter sweep. This
non-claim adapter reuses existing writer behavior and does not replace those
contributions. The existing ordinary revision-guard work (PR6030) and recorder
integration (PR6650) remain separate.

The installed harness uses normal CLI initialization and authoring on embedded
and ordinary shared-server Dolt. Pool policy and forced concurrency require real-store tests. The separate
claim harness now authors an active holder through the installed CLI; arbitrary
status editing remains unavailable. Neither
component tests nor a CLI recording qualifies this draft without complete
exact-source Linux evidence. Dolt 2.1.8 different-database provisioning remains
serialized.
