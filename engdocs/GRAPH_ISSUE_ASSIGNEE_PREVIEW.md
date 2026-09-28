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

`--claim`, `--force`, `--if-assignee`, `--if-status` and graph status editing remain
unavailable. This increment does not expose assignee-filtered graph listing.
Complete current/exact record reads retain assignment along with other Issue
properties and owned Dependencies. It does not advertise public ordered History.

The `issueAssigneeUpdate` capability and graph admission of this existing flag are
provisional preview choices. `issueWorkflows` and complete Memory remain false.
The draft starts at qualified `deff78f6`, independently of the held label-writer
and graph-label drafts. It changes no ordinary writer, lease, schema or ownership
policy. Integration must reconcile future retained-recorder ownership once, as
with the earlier Issue adapters. Jim's exact target remains pending.

Contributor boundaries: upstream PR5349 and PR6501 own claim-plus-assignee override
corrections; PR4697 and PR4715 own broader ownership fencing and tokens. This
non-claim adapter reuses existing writer behavior and does not replace those
contributions. The existing ordinary revision-guard work (PR6030) and recorder
integration (PR6650) remain separate.

The installed harness uses normal CLI initialization and authoring on embedded
and ordinary shared-server Dolt. Active-holder, pool and lease controls require
real-store tests because graph CLI claim/status authoring is not exposed. Neither
component tests nor a CLI recording qualifies this draft without complete
exact-source Linux evidence. Dolt 2.1.8 different-database provisioning remains
serialized.
