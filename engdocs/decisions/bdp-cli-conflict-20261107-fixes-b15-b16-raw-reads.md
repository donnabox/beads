# Decision record: raw reads and the listing shape over bd connect (native fixes, rows B15 and B16)

Date: 2026-11-07
Decided by: Donna Box
Status: decided; both `trish/fixes-d-gaps-20261107` and
`trish/fixes-b-connect-20261107` carry these commits, because each needs them

Scope: what the HTTP client answers when a command reads through a raw storage
method, on v0 and on BDP alike. These two rows are in group B, and groups B and D
both stand on them, so they sit at the bottom of both branches as the same commits.
They in turn stand on row E2 (see below).

| Row | Reads | Before | After |
|---|---|---|---|
| B15 | `GetIssuesByIDs`, `GetDependents`, `GetDependencies`, `GetDependencyRecordsForIssues`, `GetDependencyCounts`, `DetectCycles`, `GetStaleIssues`, `GetEpicsEligibleForClosure` | Each refuses at the store: "operation X not supported by the http backend". `bd stale`, `bd epic status`, `bd mol stale`, `bd delete --dry-run`, `bd delete`'s no-flag preview and `bd import --dry-run` exit 1 | Each is composed from existing operations, several requests per read. Those commands answer |
| B16 | `SearchIssues` with a filter naming neither ids nor a parent | Exit 1: "a raw SearchIssues call matching neither expressible shape refuses". `bd lint`, `bd todo`, `bd todo list`, `bd duplicates`, `bd find-duplicates`, `bd orphans`, `bd ready --gated`, `bd mol ready`, `bd graph --all`, `bd human list`, `bd swarm list`, `bd label list-all`, `bd rename-prefix --dry-run` exit 1 | One walk of `listIssues` with every default suppression lifted. A member the listing cannot state refuses by ledger row (`S-IssueFilter.*`) |

`bd rename-prefix` without `--dry-run` still refuses, nonzero and before any write:
`UpdateIssueID` has no operation. Only its dry run answers.

## What comes with them, whether or not group D is taken

A command that read through one of these and swallowed the refusal now gets an
answer, and sends the requests to get it. So taking these rows alone already moves
some answering commands, on v0 too:

- `bd list` text shows the gate marker (row D1).
- `bd ready` text shows the parent epic (row D2).
- `bd graph <id>` shows dependents and dependencies (the read half of row D3).
- `bd ready --explain` shows blockers and counts (the read half of row D4).
- `bd info --json` carries `issue_count` (row D5).

Those rows have no code of their own to change this, and no gate holds it back:
the v0 gates in the source history go away on any yes, because a yes applies to both
protocols. `scripts/bdp-call-parity/wire_proof.py` holds a v0 connection to a `bd`
built before the change, so it reports these as differences. It is meant to.

## Why E2 sits underneath

With the listing shape, `bd gc` no longer refuses on its first read. It would run and
print "not supported" per phase with exit 0, which reads as a sweep that ran. Row E2
refuses first, so the E2 commit is at the bottom of this stack.

## CLI surface

No flag or help text moves. The `--json` output of the commands above is new where
they refused before.

## Ledger

`engdocs/design/http-divergence-ledger.md` retires `E-IssueFilter.noShape`,
`F-ready-gated`, `F-mol-ready` and `F-ready-explain`, adds the `S-IssueFilter.*`
population under its own section, and says what now holds of the search bridge's
shapes. `F-ready-explain` said the command refused; it never did.

## What pins it

`TestServedRawReadsAgreeWithTheReference`, `TestServedStaleExemptsALiveHeartbeat`,
`TestServedSearchIssuesListingShape` (dual runs against the reference store),
`TestARawFilterIsTheListingShape`, `TestE2E_RawReadsServeCommandsOverConnect` (the
real `bd`, both protocols).
