# External T2b review repairs, before minimum-read implementation

Frozen clean7bfba525bbb66e64647110ac336f76b164b41942, direct child8844fe7.
Checkout:/Users/dbox/repos/wt/vickie/vickie-transaction-minimum
Branch:codex/vickie-transaction-minimum
Origin:https://github.com/gastownhall/bdp.git
Despite branch name, this commit contains no barrier implementation.
Frozen b1/8844 checkouts untouched; no outward writes.

F-A closed: every command carrier refuses supplied minimum-context fields with
bodyless500 + private minimum-position marker before any key/effect. Actual HTTP
matrix covers singleton/batch/sequence × each of3 fields, checks absent Resource,
then uses the exact key for a different successful request to prove no capture.
This is an honest unsupported development category, not new command semantics.

F-B closed: existing exact canonical operation URL policy is now explicit in the
method gate. Unknown/unimplemented set operation targets, encoded spellings and
query-bearing command URLs are unregistered and404 for all methods. They no
longer acquire a Read Allow table. GET/POST/DELETE matrix proves15 cases; later
same-key command succeeds. No normalization or new query parameter semantics.

F-C closed: sequence depth TransactionDevelopmentLimitError uses unsupported-
transaction-composition; only the actual alias restriction uses sequence-resource-
alias. HTTP depth/no-effect/key-reuse test added.

F-D inherited8844 method repair, no additional parser changes.
F-E documented: close aborts sockets then drains durable work. It does not promise
receipt delivery to still-connected clients; key replay after reopen recovers it.
F-F explicitly remains a private diagnosability gap: HTTP catches discard cause
visibility although owner/service preserve causes. No cause/identifier is emitted
on wire. No diagnostic callback/API silently introduced in this bounded repair.
F-G private maximumOperations documented as transaction.operations only; sequence
keeps32. Actual test proves2 sequence members succeed with direct cap1, while
2-operation batch returns413 limit=transaction.operations. No limit policy changed.
F-H documented: sequence final context is post-runner; failure loses delivery,
not committed member outcomes. Final-turn capture remains separate future work.

Changed4 files:transaction-http.ts/test, one runtime option comment, progressdoc.
Node24.16.0/pnpm11.20:
PASS frozen install/build before emitted tests.
PASS49 tests/3 files HTTP+runtime+service (focused-final.log).
PASS typecheck/boundaries/format/lint (existing warnings), both smokes, unchanged
Read evidence verification. All commands have separate external logs here.
PASS git diff --check; committed worktree clean. No interrupted tests.

Root may independently review/integrate. Approved canonical complete minimum-read
subset will start in a new isolated successor, preserving this checkpoint.
