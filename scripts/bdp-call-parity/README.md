# BDP Callables parity harness

Measures stage 1: every Callable on `bd serve`, and every native command behind it
over `bd connect`, with both asking for BDP: `bd serve --protocol both` and
`bd connect --protocol bdp`. Neither is the default. It reports anomalies. It changes no
product code and does not fail on what it finds, with these exceptions: a request from
`bd connect` that is not
`POST /bdp/call/NAME`, has `actor` or `author` in its arguments, or omits a parameter
with a service default fails the run. Otherwise it exits non-zero only when the harness
itself breaks.

```bash
./scripts/bdp-call-parity/run.sh --out /some/dir
```

`run.sh` builds `bd` and the Dolt that `go.mod` pins into `$BDP_PARITY_WORK`
(default `$TMPDIR/bdp-call-parity-work`), or takes `BD_BIN` and `DOLT_BIN`. `--part a`
or `--part b` runs one half; `--only REGEX` limits Part B to matching commands; `--keep`
keeps the workspaces for a postmortem. It writes
`results.json` (every case, with the bodies of each failure and, for Part B, the v0
operations each command dialed) and `summary.md`. Between thirty and sixty minutes once built:
Part A takes a minute, Part B the rest.

## Part A: the Callable against direct v0

- **Reads** run both ways against one server, back to back, and compare everything
  but `request_id`.
- **Writes** run in lockstep on two servers holding the same state: v0 on one, the
  Callable on the other. Timestamps, revision tokens and minted ids are masked
  (`normalize.py`). After each write the full state of both servers is compared, and
  `changed` is checked against whether the state moved.
- **Refusals** map v0's code through the spec's section 5, and check the code, status,
  `retry`, content type, and that v0's body is carried verbatim.
- `callable_cases.py` is the case table. Its `ROUTES`, `ACTOR_MEMBER` and `DEFAULTS` are
  written from the spec, not from `callables_table.go`, so a mistake shows as a difference.
- **A case's arguments are v0's.** `for_callable` moves its `actor`, or `author` on
  `issue/comment`, to `revisionProvenance`. `with_defaults` fills in, on the v0 side, each
  parameter the case omits that has a service default, because a direct v0 caller has to
  say it. An invocation naming no actor where one is required is refused before v0 runs,
  and the case checks the refusal names `revisionProvenance.actor`.

## Part B: the native command over `bd connect`

Two workspaces in lockstep: one using its Dolt server directly, one attached with
`bd connect --protocol bdp` to a `bd serve --protocol both` on another Dolt server. Each case runs with `--json` on
both, and again without `--json` because the text path often reads through different
calls, then compares exit code, normalized output, and what the command moved, read
back with `bd export` in direct mode on both databases. Verdicts: `ok`, `refused`,
`refused-exit-zero`, `output-differs`, `order-differs`, `state-differs`, `both-fail`, `both-fail-differently`,
`accepted-only-over-connect`, `hung`, `seed-failed`.

`native_cases.py` is the case table. `ops` on each result is what the serve log says
the command dialed, which is the evidence for which Callable a command corresponds to.

## What the client sends

Part B's client reaches `bd serve` through `recorder.py`, a recording proxy. For each
request it records the headers and argument names, flags any argument its Callable does
not declare, and fails the run if the request is not an invocation at all, carries
`actor` or `author` in `arguments`, or leaves a defaulted parameter to the service. Each read is
also asked of v0 at once and compared with what the client got, so `client_wire` in
`results.json` is the live evidence that the mount and v0 still agree on real traffic.

## The server table against the spec

`check_spec.py --spec PATH_TO_BDP_CHECKOUT` compares `internal/httpapi/callables_table.go`'s
name, kind, `cli` and `defaults` with the declarations in `specs/types/issue/callables.md`,
both ways, fails if a declaration on either side has `actor` or `author`, and exits 1 on
any difference. It is the server's counterpart of the reference
client's `callables:check`, so a spec change shows up as a failure. It starts no stack.

## The default against the base commit

`wire_proof.py` is the check that the defaults (`bd serve` and `bd connect` with no
`--protocol`) are the code that was there before BDP. It takes a `bd` built at the base
commit and the `bd` under test, and it starts no `--protocol` on either.

```bash
python3 scripts/bdp-call-parity/wire_proof.py --base BASE_BD --new NEW_BD --dolt DOLT --out DIR
```

- **Client half.** Part B's whole command table runs through each binary against its own
  `bd serve`, through a proxy that records every byte of every request and response. Each
  command's requests are compared in order: method, raw path and query, every header in
  order, and the body.
- **Server half.** Each server is asked the same raw requests in lockstep: Part A's v0 reads
  and writes, then fixed probes of `/`, `/bdp`, `/bdp/call`, Resource addresses and unrouted
  paths, with and without an auth token file. Status, headers and body are compared.
- **Masked, and nothing else:** ports and the temporary root, timestamps, UUIDs (the project
  id), revision tokens, minted ids numbered by first appearance, cursors, `request_id`,
  `arrival_seq` (it counts every create so far, the allowed ones too), `Date`, `Content-Length`
  (it moves with a masked value), the build hash in `bd version`, and the time in `bd ping`.
- **Verdicts.** `same`; `allowed` (the base refused the command and the new bd serves it, or
  gets further); `query-order` (one repeated query value in another order, which is map order
  in the client and differs base against base too); `state` (a command whose own wire matches, with an answer that moved because
  an earlier `allowed` command wrote something on one side only); `output` (same wire, different
  words on stdout or stderr); `unexplained` (anything else). `allowed-differences.md` lists every
  `allowed` command with the base's refusal and what the new bd now sends. It exits 1 on any
  `unexplained`.

## Cleanup

`stack.Root.reap` stops what it started, then anything left whose command line or
working directory is under its temporary root (`bd db-proxy-child` outlives its
parent). It kills nothing else.
