# scripts/scenarios: Memory Beads scenario driver

A black-box driver for Memory Beads on the fork's graph mode. It runs JSON
scenario files against a real `bd` binary, in two fresh hermetic workspaces per
scenario, and checks two things:

1. the scenario's **expectations**, on workspace A's raw values. They are
   declared from the cited spec text (#5877, #5898), never recorded from what
   the CLI happens to do. Where the CLI differs from the spec text the scenario
   is `xfail` and the difference is written down in a finding bead;
2. **determinism**: the normalized transcripts of A and B must be byte-identical.

The scenarios are data. The driver knows nothing about any particular scenario.

## Running it

```sh
go build -tags gms_pure_go -o /var/tmp/bd ./cmd/bd     # a bd built at the commit under test
go run ./scripts/scenarios --list                       # one scenario id per line
go run ./scripts/scenarios --bd /var/tmp/bd --out /var/tmp/scenario-out
```

| flag | meaning |
|---|---|
| `--bd PATH` | the `bd` under test. Always an explicit file, never looked up on `PATH`. Required unless `--list`. |
| `--out DIR` | transcripts and `receipts.json`. Must not exist or be empty. Required unless `--list`. |
| `--scenarios DIR` | where the scenario files are read from, at run time. Default `scripts/scenarios/scenarios` under the repo root. |
| `--scenario ID` | run only this scenario; repeatable. `discovered` in the receipts is unaffected. |
| `--engine embedded` | the only engine in this version. |
| `--list` | print the scenario ids, one per line, sorted, and exit. The CI lane builds its matrix from it. |
| `--keep-workspaces` | keep the temporary workspaces (their paths are printed on stderr). |

## Adding a scenario

Add one file, `scripts/scenarios/scenarios/<id>.json`. That is all: **no Go
change and no `BUILD.bazel` change.** The directory is read at run time and the
scenarios are deliberately not `go:embed`ed or listed anywhere, because gazelle
enumerates `embedsrcs` and every new scenario would then edit the generated
`BUILD.bazel`, so scenario changes could no longer land independently.

* `id` must equal the file stem, must match `[A-Za-z0-9][A-Za-z0-9._-]{0,63}`,
  and is unique case-insensitively (transcript directories are named after it).
* An empty or missing scenarios directory is exit 3, never a silent zero.
* Cut a scenario, never weaken it. If it needs something outside the assertion
  vocabulary below it goes back to the pm to cut or to the architect to extend
  the vocabulary. Nobody loosens an assertion to make it fit.
* Prove it loads: `--list` must print your id.

```json
{
  "id": "R2",
  "title": "one line",
  "spec": ["#5877 S7", "be-awvi5 step 1"],
  "state": "pass",
  "steps": [
    {
      "name": "remember-m1",
      "argv": ["remember", "--stdin", "--id", "beads/r2/m1", "--title", "M1", "--json"],
      "stdin": {"name": "B1", "text": "body\n"},
      "capture": {"REV1": {"path": "$.result.revision", "kind": "token"}},
      "expect": [{"op": "exit", "equals": 0}]
    }
  ]
}
```

* `spec` cites the spec text the expectations come from. It is required.
* `argv` is passed to `bd` after the binary. `${NAME}` is replaced by a capture
  of an **earlier** step. Anything else that looks like `${...}` is a manifest error.
* `actor` (`^[a-z][a-z0-9-]{0,31}$`) appends `--actor <name>` to that step's argv,
  which shows in the transcript and so is identical across A and B. `--actor`
  beats the pinned `BEADS_ACTOR=scenario-actor` (cmd/bd/main.go). Steps without
  an actor run as `scenario-actor`.
* `stdin` is `{"text": ...}`, `{"b64": ...}` or `{"generate": {"repeat": "ab", "bytes": N}}`
  (`repeat` cycled and cut at exactly N bytes: deterministic, for bodies over
  1 MiB). Give it a `name` to refer to those bytes from `stdout equals_input`.
* `capture` names a value of the step's result: `{"path": "$.a.b"}`, `{"stdout": true}`
  (the raw stdout) or `{"stderr_code": true}`. A capture is per workspace. Mark
  values that are random per write with `"kind": "token"` (version and revision
  tokens) or `"kind": "id"` (a returned id): the normalizer gives those stable
  ordinals. Captures without a kind are stable and are recorded in
  `receipts.json`.
* Every step needs exactly one `exit` expectation.
* Steps run in order and stop at the first step whose expectation fails. B
  replays exactly the steps A ran.
* Always pass explicit `--id beads/...` (Memory) and `--id links/...` so no
  random ids appear. Blocking links and deps take no `--id`: capture the id
  they return with `"kind": "id"`.

## The assertion vocabulary (closed)

Each entry of `expect` is an object with an `op`. **The vocabulary is closed**:
an unknown `op` or field is a manifest error (exit 3), so a scenario can never
carry code. Every assertion is evaluated on workspace A's raw values.

| `op` | fields | holds when |
|---|---|---|
| `exit` | `equals: N` (0..255) or `nonzero: true` | the exit code is N / is not 0 |
| `stderr_code` | `equals: "X"` or `not_equals: "X"` | the refusal's typed code is / is not X |
| `stdout` | `equals_input: NAME`, `equals_capture: NAME` or `empty: true` | stdout is byte-for-byte the named stdin / capture, or has no bytes (raw, no normalization) |
| `contains`, `not_contains` | `marker` and one of `in: "stdout"` or `path: "$.a.b"` | the marker is / is not in stdout or in that JSON string value |
| `json_path` | `path`, `equals_capture` | the value at the path equals the captured value |
| `id_set` | `path`, `equals: ["a", "b"]` | the ids at the path (an array, or a `[*]` projection) are exactly that set, order ignored |
| `compare` | `left`, `right`, `relation: "equal"` or `"differ"` | two steps' fields are equal / differ. A side is `{"step": S, "field": "stdout" \| "stderr" \| "stderr_code" \| "exit"}` or `{"step": S, "path": "$.a"}`; S is this step or an earlier one |
| `stderr_contains` | `capture` | stderr contains the captured token |

The stderr code is the `code` of a JSON `{code,message,retryable}` refusal, or
else the leading `code:` token of the first line of the CLI's plain-text form
(`gone: graph resource was deleted`). `recall` refuses `--json`, so its
refusals only ever come in the plain-text form.

### JSON path

A deliberately tiny subset, so no scenario can lean on an implementation-specific
JSONPath feature: `$.a.b[0].c` (fields and non-negative indexes) and the array
projection `$.items[*].id` (at most one `[*]`). Anything else, including `$`
alone, `..`, filters, slices and negative indexes, is a manifest error. A path
that does not resolve when the step runs is an assertion failure.

## States

| `state` | required | meaning |
|---|---|---|
| `pass` | nothing else | every expectation must hold |
| `xfail` | `finding`: a bead id, `^[a-z][a-z0-9]*-[a-z0-9]+(\.[a-z0-9]+)*$` | the CLI differs from the spec text and the finding bead says exactly how. The scenario must **fail**; if it passes that is an unexpected pass (`xpass`) and exits 1 |
| `skip` | `gap`: an item number from be-tatfn, `be-tatfn#N` | a declared gap. Not run, but recorded in the receipts with its gap and counted as executed, so it is visible |

`finding` is only valid with `xfail` and `gap` only with `skip`.

## Normalization

The two workspaces get different tokens, timestamps and paths; the transcripts
are compared after normalization, which is structure-preserving, not a set of
blanket masks:

* JSON stdout starts from `protocol.CanonicalizeJSON`: sorted keys, id-sorted
  object arrays, RFC3339 timestamps to `<TS>` (which is how `attribution.recordedAt`
  is masked), `commit` dropped.
* Version and revision tokens (16 random bytes, hex, new on every write) and
  captured `token` / `id` values become **ordinals by first appearance across the
  whole transcript**: `<TOKEN#1>`, `<TOKEN#2>`, `<ID#1>`. Equal values keep equal
  placeholders and distinct values stay distinct, so a bug that collapses two
  versions into one still shows. This also holds for tokens that appear in
  argv and inside stderr messages (a `revision_conflict` message names the
  current revision).
* `CanonicalizeJSON` masks the bare keys `revision` and `version` whatever they
  hold, which would merge distinct tokens. The normalizer hides those two keys
  from it for the call and restores them, then assigns the ordinals itself.
* The workspace root, in either spelling of a symlinked temp dir, becomes `<WS>`.
* Stdout that is not a JSON object (such as `recall`'s exact body bytes) is kept
  byte for byte.
* `attribution.recordedAt` is masked for the comparison, but the order of the
  states each workspace wrote is still asserted: a create's `result.attribution`
  and a changed update's `result.memory.attribution` must be non-decreasing.

## Receipts (`--out DIR`)

```
transcripts/<scenario>/<A|B>/NN-<step>/{argv,exit,stdout,stderr,sha256}
receipts.json
```

The transcript files hold the normalized values. `receipts.json` records
`driver_version`; `source` (`commit`, `dirty`); `bd` (`path`, `sha256`, `version`);
the `env` whitelist; the `scenarios_dir`; and per scenario `id`, `state`,
`outcome`, `engine`, `steps`, the scenario `file` and its `file_sha256`,
`transcript_sha256_A` and `_B`, `equal`, `duration_s`, the `finding` or `gap`,
and the stable `captures`.

`discovered` is the number of `*.json` files in the directory, whatever
`--scenario` selected. `executed` is the number processed. With no `--scenario`
filter they must be equal, the same exhaustive-receipts rule as
`scripts/ci/graph-c0-qualify.py`.

## Exit codes

| code | meaning |
|---|---|
| 0 | every selected scenario is `pass`, `xfail` or a declared `skip`; discovered == executed; the workspaces agreed |
| 1 | a scenario failed, an `xfail` passed (`xpass`), or the two workspaces diverged after normalization |
| 2 | harness error or a guard tripped: `bd` missing, graph mode not active, unsafe workspace, timeout, unsupported engine, bad usage, non-empty `--out` |
| 3 | manifest error: bad or duplicate id, id not equal to the file stem, unknown state, `xfail` without a finding, `skip` without a gap, unknown op or field, unsupported JSON path, bad actor name, an undefined `${var}` or capture, empty or missing scenarios directory |

The manifest is validated before `bd` is touched, so a bad scenario file never
costs a workspace.

## Guards

Each is exit 2:

* **Workspace**: a fresh directory from `os.MkdirTemp` holding `work/`, `home/` and
  a driver-written marker naming the root. Before *every* `bd` call the root
  must be under the OS temp dir and the marker must exist and name that root.
* **Environment**: built from nothing, never from the caller's. Exactly:
  `HOME` and `XDG_{CONFIG,CACHE,DATA,STATE}_HOME` under `root/home`, a minimal
  `PATH`, `TZ=UTC`, `NO_COLOR=1`, `GIT_CONFIG_NOSYSTEM=1`, `GIT_CONFIG_GLOBAL`
  (an empty file), `BEADS_DOLT_AUTO_START=0`, `BD_DISABLE_METRICS=1`,
  `BD_DISABLE_EVENT_FLUSH=1`, `DOLT_METRICS_DISABLED=1`, `BEADS_ACTOR=scenario-actor`,
  `BEADS_TEST_IGNORE_REPO_CONFIG=1`. `BEADS_DIR`, `BEADS_DB` and any other `BD_*`
  cannot reach `bd`; that is asserted before every call.
* **Graph mode**: `bd init --graph-mode link --scope-url https://example.invalid/scenarios/
  --prefix scn --skip-hooks --skip-agents --non-interactive`, then `bd status --graph`
  must succeed and `.beads/graph-preview-format` must exist. Otherwise the run
  fails closed.
* **Every `bd` call**: the explicit binary from `--bd`, cwd = `work/`, capped at
  90 seconds, no network (`example.invalid` only).

## Not in scope

The server engine and a cross-engine diff, a Dolt-native oracle over the graph
tables, erase, memory classes, rebuild and refuse-or-preserve (the be-tatfn
gaps), and any CI workflow file. The scenarios themselves come later, as data.
