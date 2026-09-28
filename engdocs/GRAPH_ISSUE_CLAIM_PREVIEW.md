# Experimental graph Issue claim

Claim one canonical graph Issue using the existing standalone command:

```sh
bd update beads/work --claim --actor alice --json
bd show beads/work --json
bd list --format records-json --assignee alice --all
```

A successful claim sets the assignee to the actor, moves the Issue to
`in_progress`, and grants the ordinary node-local lease. It sets StartedAt only
when that value is absent. The checked Issue, complete owned Dependencies and
retained graph version remain coherent in the same transaction. Target Issues
and informational Links are unchanged. The JSON result uses the graph preview
envelope with `result.issue` and `result.changed`; it is not legacy update JSON.
Human output is `Claimed CANONICAL_ID` or `Unchanged CANONICAL_ID`, followed by a
newline. Quiet mode suppresses human success output.

An Issue can be claimed from `open`, or a configured custom status in the active
category, when unassigned, assigned to the same actor, or assigned to a literal
configured claim pool. Other holders and nonclaimable statuses return
`constraint_violation` (exit 4). Actor equivalence and exact pool membership reuse
the ordinary claim rules; this introduces no authentication or new identity
model. The existing actor source and 255-code-point field bound apply.

Claim does not check readiness. An open Issue with unfinished blocking
Dependencies can be claimed; those blockers remain present and `bd ready` still
applies its usual rules. This command does not claim the prerequisites.

Repeating a claim as the same actor while already `in_progress` returns
`changed:false`. It preserves the stored actor spelling, timestamps, version,
owned state, attribution and lease. It does not create another retained version
or claim event. The adapter checks the native row token and complete projection before accepting
a reported no-op, and rolls back inconsistent writer behavior. It is not a
general audit of arbitrary future writer side effects: adopting a new writer
still requires proving its no-write behavior. Concurrent changes still use the existing transaction conflict
behavior; a no-op is not a guarantee that another process cannot subsequently
change the Issue.

## Bounded command admission

Only the standalone `--claim` or `--claim=true` form is admitted. Exactly one
canonical `beads/PATH` or canonical Scope URL is required. `--claim=false` refuses.
Even with field edits, `--claim=false` returns `invalid_properties`; omit the
flag to request a field edit. The ordinary actor, JSON, quiet and workspace
controls retain their meanings.
Read-only and migration-freeze checks run before storage is opened.

Every mixed edit is unavailable, including assignee, status, text, priority,
labels, metadata and parent changes. Graph revision guards, unconditional mode,
ordinary assignee/status guards, force, and TTL options are unavailable. Explicit
false or empty values do not bypass the changed-flag refusal. Unknown flags retain
normal parser errors. Ordinary non-graph `bd update --claim` remains unchanged,
including its broader combinations, batches and routing.

This operation reuses claim's atomic ownership check. It does not translate an
opaque graph revision into a native row token, and does not implement claim by
setting assignee/status fields through the scalar update adapter. Wrong resource
kinds, foreign authority, missing resources and transaction failures retain the
existing graph refusal rules. An uncertain commit returns `outcome_unknown`;
inspect the canonical resource rather than automatically replaying the write.

## Lease and History limits

The ordinary default lease lasts five minutes. Repeating a claim does **not**
renew it. This preview does not expose graph heartbeat, reclaim, unclaim or TTL
controls and does not provide a complete long-running lease lifecycle. Ordinary
writer commands are refused in a graph workspace. Bypassing that controlled
writer boundary can invalidate current graph reads: hydrated lease fields are
checked against the retained Issue snapshot. If an external heartbeat changes or
removes that lease row, current reads and mutations refuse. There is no supported
repair path for that Issue in this preview; it offers no recovery command. Saved exact versions remain
readable as the observations recorded at that time. Passive time expiry alone
does not change the stored lease. Node provenance can be empty under the existing
unnamed-node policy. Clone/adoption, replication recovery and the choice of live
lease fields versus retained properties require separate review; this slice does
not redefine those public representations.

`issueClaim` is a provisional capability; `issueWorkflows` and public History
remain unavailable. Existing exact retained reads and comparisons are available,
without a public History ordering or native commit-time claim.

At this integration base, `ExecuteClaim` records its ordinary events but does not
own the retained Issue recorder. The graph adapter invokes the recorder exactly
once. Upstream PR6650 has merged recorder changes elsewhere; integration must
reconcile recorder ownership against the actual imported writer, removing the
extra call if that writer assumes responsibility. Do not double-record versions.
This isolated adapter does not absorb or supersede contributor work on mixed
claim-plus-assignee behavior or broader ownership/lease policies.

Qualification requires real installed CLI evidence from normal initialization on
embedded and ordinary shared-server Dolt, retained/current complete-state checks,
no-op and refusal controls, rollback and concurrent-claim checks. Different-
database provisioning on Dolt 2.1.8 remains serialized. This document specifies
acceptance scope; it is not a claim that a pending draft has passed qualification.
