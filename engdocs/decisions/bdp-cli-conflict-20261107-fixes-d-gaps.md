# Decision record: commands that answered with a gap over bd connect (native fixes, group D)

Date: 2026-11-07
Decided by: Donna Box
Status: decided; branch `trish/fixes-d-gaps-20261107` carries the change

Scope: commands that answered with exit 0 and a gap on a `bd connect` workspace.
Each row applies to v0 and to BDP alike, so a v0 connection stops matching the base
for these on purpose, and rows D3 and D4 also change the direct route.

This branch stands on rows E2, B15 and B16
([bdp-cli-conflict-20261107-fixes-b15-b16-raw-reads.md](bdp-cli-conflict-20261107-fixes-b15-b16-raw-reads.md)),
because nothing here can read without them.

| Row | Command | Before | After |
|---|---|---|---|
| D1 | `bd list` text | Exit 0; no row renders as gated over connect, and the refused read is swallowed | Reads the gate targets and renders the marker. New requests |
| D2 | `bd ready` text | Exit 0; no parent epic shown over connect | Reads edges and epics and shows it. New requests |
| D3 | `bd graph <id>` | Exit 0 over connect with only the root and its own edges, because a refused read hit `continue`. Direct swallows a failed read the same way | Reads dependents and dependencies. A failed read is the command's failure, **direct included** |
| D4 | `bd ready --explain` | Exit 0 over connect with blank blockers and zero counts; failed reads go to a debug log | Reads counts, edges, cycles and blockers. A failed read is the command's failure, **direct included**. The rows' embedded edges are cleared, because the explanation reports edges from its own read |
| D5 | `bd info --json` `issue_count` | Absent over connect, exit 0 | Present: the listing is sent |
| D6 | `bd show --refs`, `bd show --children` | Exit 0 over connect with shallow rows: JSON has zeroed timestamps and blank fields | Rows are hydrated through `batchGetIssues`; a row it cannot supply refuses |
| D7 | `bd update --claim` with nothing else | Exit 0, sent as `PATCH` `updateIssue` | Sends `claimIssue`, with the template guard client-side and a read-back. Same result, different request. A wisp, or a claim beside any other edit, stays on `updateIssue` |
| D8 | `bd kv get`, `list`, `set`, `clear` | `kv get` of any key exits 1 "(not set)" even when it exists; `kv list` exits 0 "No key-value pairs set". Both are silently wrong, because `bd serve` hides `kv.*` | Every `bd kv` subcommand refuses up front, nonzero, naming why |

D1, D2 and D5 have no code of their own: `cmd/bd` at the base already does what they
want once the reads answer. This branch carries their tests.

## CLI surface

No flag or help text moves. D5 adds a key to `bd info --json` over connect. D6
changes `--json` rows from shallow to whole. D8 turns an exit-0 `kv list` into a
refusal.

## Ledger

`engdocs/design/http-divergence-ledger.md`: `F-show-refs` and `F-show-children` retire
(they said the commands refused, which they never did). `F-ready-explain` says a
failed read is the command's failure.

## What pins it

`TestE2E_ConnectGapsAreClosed` (the real `bd`, both protocols, every row),
`TestGraphAFailedReadIsTheCommandsFailure`, `TestReadyExplainAFailedReadIsTheCommandsFailure`,
`TestHydrateRemoteDependentsRefusesARowItCannotSupply` (stubbed store, for the failure
paths), and `TestE2E_ACommandAnswersTheSameOnBothProtocols`.
