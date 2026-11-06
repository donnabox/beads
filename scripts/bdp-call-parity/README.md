# BDP Callables parity harness

Measures stage 1: every Callable on `bd serve`, and every native command behind it
over `bd connect`. It reports anomalies. It changes no product code and does not fail
on what it finds, with one exception: a request from `bd connect` that is not
`POST /bdp/call/NAME` fails the run. Otherwise it exits non-zero only when the harness
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
- `callable_cases.py` is the case table. Its `ROUTES` is written from the spec, not
  from `callables_table.go`, so a routing mistake shows as a difference.

## Part B: the native command over `bd connect`

Two workspaces in lockstep: one using its Dolt server directly, one attached with
`bd connect` to a `bd serve` on another Dolt server. Each case runs with `--json` on
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
not declare, and fails the run if the request is not an invocation at all. Each read is
also asked of v0 at once and compared with what the client got, so `client_wire` in
`results.json` is the live evidence that the mount and v0 still agree on real traffic.

## The server table against the spec

`check_spec.py --spec PATH_TO_BDP_CHECKOUT` compares `internal/httpapi/callables_table.go`'s
name, kind and `cli` with the declarations in `specs/types/issue/callables.md`, both
ways, and exits 1 on any difference. It is the server's counterpart of the reference
client's `callables:check`, so a spec change shows up as a failure. It starts no stack.

## Cleanup

`stack.Root.reap` stops what it started, then anything left whose command line or
working directory is under its temporary root (`bd db-proxy-child` outlives its
parent). It kills nothing else.
