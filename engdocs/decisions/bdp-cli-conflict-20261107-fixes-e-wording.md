# Decision record: bd connect refusals say why (native fixes, group E)

Date: 2026-11-07
Decided by: Donna Box
Status: decided; branch `trish/fixes-e-wording-20261107` carries the change

Scope: the words of four refusals on a `bd connect` workspace, on v0 and on BDP
alike. Every row refused nonzero before and refuses nonzero after. Direct and
proxied workspaces are not touched, and no flag, help text or `--json` success
shape moves. The `--json` envelope of each refusal carries the new words in
place of the old, which is the only output a CLI-surface comparison can see.

The rows come from the list of native fixes kept out of the BDP landing
candidate so that it would stay identical to the base.

| Row | Command | Before | After |
|---|---|---|---|
| E1 | `bd sql` | `'bd sql' is not yet supported in embedded mode`, which is false: the workspace is not embedded | `'bd sql' is not available over bd connect: bd serve publishes no SQL surface; run it on the server's own workspace` |
| E2 | `bd gc` | `searching closed issues: listIssues: a raw SearchIssues call ... refuses` | `bd gc is not available over bd connect: it maintains the server's own database; run it on the server's workspace` |
| E3 | `bd sync` | `storage backend does not support is_blocked recompute` | `bd sync is not available over bd connect: bd serve is the sync boundary and syncs its own database; run it on the server's workspace` |
| E4 | `bd show --as-of` | `Error fetching X as of HEAD: operation AsOf not supported`, then a `no issues found` envelope in JSON, which is false: the ids may exist | `bd show --as-of is not available over bd connect: bd serve publishes no point-in-time read; run it on the server's own workspace` |

## Why the words matter

An agent reads these messages to decide what to do next. "Embedded mode" sends it
looking for a mode switch, and "no issues found" says an issue that exists does
not.

## Connection to the listing shape of SearchIssues

E2 is also needed by row B16. Once a raw `SearchIssues` has a listing shape, `bd gc`
would no longer refuse on its first read: it would run and print "not supported"
per phase, with exit 0. So `trish/fixes-d-gaps-20261107` and
`trish/fixes-b-connect-20261107` carry the E2 commit in their history.

## What pins it

`TestE2E_ConnectRefusalsSayWhy` in `backend/http/e2e_refuse_over_connect_test.go`
runs each row through the real `bd` on both protocols. The corpus allowlist reads
E1's words for `gc-sql-ready-projection` and `gc-sql-release-fallback`.
