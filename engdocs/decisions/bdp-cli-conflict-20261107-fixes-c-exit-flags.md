# Decision record: bd defer's exit code and comments add --author (native fixes, group C)

Date: 2026-11-07
Decided by: Donna Box
Status: decided; branch `trish/fixes-c-exit-flags-20261107` carries the change

Scope: two behaviors that change a working command, on the direct, proxied and
`bd connect` routes alike, on v0 and on BDP alike. A script that read the old exit
code, or passed both flags, will see a difference. That is the point of each.

The rows come from the list of native fixes kept out of the BDP landing candidate
so that it would stay identical to the base.

| Row | Command | Before | After |
|---|---|---|---|
| C1 | `bd defer`, `bd undefer`, a failure for one id. Includes `bd undefer` on an issue that is not deferred | Prints the error; **exit 0** | Prints the error; the other ids still run; **exit 1**. Whatever the command did for the other ids is still reported |
| C2 | `bd comments add --author X` | `--author` names the comment's author only. `--actor` governs the embedded commit, and the two can differ | `--author` is an alias of `--actor`: it sets the actor for the whole write, embedded commit included. `--actor sam --author kim` refuses before any write. The same value twice is one value. `BD_ACTOR` and `BEADS_ACTOR` are defaults, which `--author` overrides, as before |

## CLI surface

C2 changes the help tree: `cmd/bd/testdata/command_tree/tree.golden` gains an example
under `bd comments` and `bd comments add`, and the usage of `--author` becomes
"Alias of --actor; two different values are refused". No flag is added or removed.
The generated CLI reference under `docs/` is pinned to the released version
(`docs/cli-docs.pin`), so it changes when a release carries this, not here.

C1 changes no help text and no `--json` success shape. A failed id's exit code moves;
a JSON caller still gets the array of the ids that did land.

## A conflict to expect with row B8

Row B8 (`trish/fixes-b-connect-20261107`) moves the same two commands onto the
Lifecycle role. Taking both means one hand-merged hunk in `cmd/bd/defer.go` and
`cmd/bd/undefer.go`: keep B8's write and C1's `hasError`.

## What pins it

- C1: `TestDeferAndUndeferExitNonzeroOnAPerIDFailure` (direct, stubbed store),
  `TestProxiedLookupReportsMissingIssue` and `TestProxiedLookupReportsBackendFailure`
  (proxied; `exitsZero` was `true` for both commands), `TestEmbeddedUndefer`
  (embedded), and a subtest of `TestE2E_ACommandAnswersTheSameOnBothProtocols`
  (connect, both protocols).
- C2: `TestCheckCommentAuthorAlias`, `TestEmbeddedCommentsCLI`'s
  `comments_add_author_is_an_alias_of_actor`, `testdata/comments.txt`, `tree.golden`,
  and a subtest of `TestE2E_ACommandAnswersTheSameOnBothProtocols`.
