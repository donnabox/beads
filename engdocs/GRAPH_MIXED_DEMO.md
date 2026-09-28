# A release, a plan, and the reason behind it

This is a five-to-eight-minute walkthrough for people who have not been following
the implementation. Use the verified recording first; the same script can create
a new recording from a fresh disposable database.

The separately delivered **`crew-demo.zip`** packet contains
`embedded-reviewed/recording.md` and `shared-server-reviewed/recording.md`, with
their adjacent receipts. Start with the embedded story, then open the server
recording for the BDP chapter. The packet accompanies this guide; it is not
checked into this repository. If you only have the repository, use “Record it
again” below to make your own recordings. Each recording header identifies its
runtime pin, binary hash, backend and `qualification: false`; retain the packet
rather than sharing an unlabeled excerpt.

The story uses the qualified Beads candidate **b7bf5040**. Claim is a separately qualified successor in draft PR47; notes append remains
pending in draft PR48. Both are deliberately absent from this pinned recording.
The [implementation map](GRAPH_IMPLEMENTATION_MAP.md) lists the wider supported
surface, where the data lives, and what remains unfinished.

## Set the scene — recorded chapter 1

“We have work to do, a plan for doing it, and a reason for the plan. The work is an
Issue. The plan and rationale are Memories. All three are Beads, so we can link
them and inspect the relationships without pretending a document is a task.”

The recording creates two Issues, `beads/release` and `beads/verification`, and
two Memories, `beads/plan` and `beads/rationale`. It uses normal `bd init`, `create`
and `remember` calls. Every CLI call is a new process. There is no SQL seed or
hidden bootstrap operation.

## 1. Follow the context — also recorded chapter 1

The scene-setting and this section both belong to the recording's first chapter,
“Give the project a plan and a reason.” `binary-version` precedes the chapters;
`graph-capabilities` records the runtime's own incomplete-capability disclosure
after initialization.

Open the recorded `issue-context`, `memory-rationale`, `show-linked-plan` and
`list-plan-links` commands.

“The release points at the plan. The plan points at its rationale. A Link has its
own identity and properties: this one says ‘Original rationale.’ The plan owns
its outgoing Link, so its saved state includes the relationship as well as the
text. Adding the relationship did not rewrite the rationale.”

The informational Link uses an installed experimental Type. It is not an
arbitrary unregistered label. This demo does not install a new Type.

## 2. Change a relationship

Open `edit-link-note`, `show-plan-after-link-edit` and `show-original-linked-plan`.

“We learned something: verification found a real regression. We can put that fact
on the relationship. The Link gets a new version, and so does the Memory that
owns it. Looking at the old plan still shows the old relationship.”

This demonstrates Link-property editing and complete owned-state retention.
Issue-to-Memory informational Links are provisionally unowned; that broader
ownership policy is still under review.

## 3. Change the plan and read what it used to say

Open `edit-memory`, `show-current-plan`, the stale-guard refusal and unchanged
reread, `recall-saved-body` and `compare-plans`.

“The plan now requires a rollback rehearsal too. The update checks the revision
we read. Reusing the old revision is refused, and the following read confirms
that the rejected write changed nothing. A fresh
process reads the new plan. We can also retrieve the exact old body and compare
the two saved versions. The rationale Link is still there.”

Be precise about History: this is exact saved-version retrieval and comparison.
It is not ordered history browsing, an as-of query, restoration, or public BDP
History. The presenter supplies tokens saved from earlier results; the demo does
not discover a chronological version list.

## 4. Show the Issue workflow still works

Open `add-blocking-dependency`, `ready-before-verification`, `close-verification`
and `ready-after-verification`.

“The release is blocked on verification. When we close verification, the release
becomes ready. This is the existing Issue workflow, using the existing Issue and
Dependency storage, exposed as part of the same graph.”

The rationale Link supplies context; the blocking Dependency controls readiness.
They are different relationship Types with different behavior.

## 5. Read it through BDP

Use the **ordinary-server** recording's public-client section.

“Now a separately built client reads the same graph over HTTP using BDP. It reads
the revised Memory, the Issue and their owned Links, follows incident Links, and
pages through the Beads. The client is from the BDP repository; it is not calling
our Go storage API.”

The script verifies complete current properties and the public members of owned
Links—identity, Type, endpoints, revision and properties—against the CLI results.
It does not compare the BDP attribution projection or the CLI's private saved
version aliases. It retains actual request URLs, status codes and response
hashes; raw HTTP bodies are not retained for offline rehashing.
Writes in this demonstration use the CLI. BDP here is Read-only. Embedded CLI
works, but HTTP serving requires ordinary shared-server Dolt in this preview.

## What this exercises, and what it does not

| Exercised in this story | Outside this story or unfinished |
|---|---|
| Normal graph initialization; Issue and Memory creation | Existing-database adoption, migration and backup continuity |
| Mixed links, incident listing and Link-property updates | Custom Type installation, remote endpoints, target pins |
| Guarded Memory update; complete old-state recall and comparison | Ordered/native/public History, as-of, restoration, erasure |
| Blocking Dependency and close/ready workflow | Full Issue CLI parity; claim/append drafts; held label edits |
| Independent BDP discovery, resource reads, incident links and pagination | HTTP writes and BDP Update/History |
| Same ordinary Dolt database and existing Issue machinery | Common metadata, full Memory Inception/derivation and aliases |

The wider qualification suite separately covers concurrency, rollback, authority
failures and uncertain COMMIT outcomes. A successful short recording is not a
replacement for those checks or human review of provisional contracts.

## Record it again

Use separate checkouts for the qualified runtime and this demo's scripts. Build
Beads from the exact qualified commit, with no source changes, and retain the
source status and installation log. Substitute fresh absolute paths below:

```bash
set -euo pipefail
demo_beads=/absolute/clean-beads-checkout
demo_evidence=/absolute/new-runtime-build-evidence
demo_bin=/absolute/demo-bin
demo_base=b7bf5040beae863b0d4fafa8d4b838dd04b3cd20
mkdir "$demo_evidence"
cd "$demo_beads"
test "$(git rev-parse HEAD)" = "$demo_base"
git status --porcelain > "$demo_evidence/source-status.txt"
test ! -s "$demo_evidence/source-status.txt"
git diff --exit-code HEAD > "$demo_evidence/source-diff.txt"
git rev-parse HEAD > "$demo_evidence/source-commit.txt"
make install-force INSTALL_DIR="$demo_bin" > "$demo_evidence/install.log" 2>&1
"$demo_bin/bd" version --json > "$demo_evidence/binary-version.json"
python3 - "$demo_bin/bd" "$demo_evidence/binary-sha256.txt" <<'PYHASH'
import hashlib, pathlib, sys
pathlib.Path(sys.argv[2]).write_text(hashlib.sha256(pathlib.Path(sys.argv[1]).read_bytes()).hexdigest() + "\n")
PYHASH
```

Run the script from this demo branch. It accepts a 7–40-character hexadecimal
build prefix of the required commit; if the version JSON supplies a `commit`
field, it must match too. The recorded installed binary has no `commit` field:
its build label alone is **not** full-commit runtime attestation. Keep the clean
source status, exact source commit, successful install log and binary hash above
alongside the recordings. The script refuses an existing output directory and
retains failures.

```sh
python3 scripts/graph-mixed-demo.py --bd /absolute/demo-bin/bd \
  --backend embedded --output-dir /absolute/new-embedded-recording
```

For the BDP chapter, make a clean, explicit client build at the public pin. This
recipe creates the **sibling** `client-build.json` that the presentation harness
requires; merely following the separate HTTP guide does not create that file.
Use Node **24.16.0** and pnpm **11.20.0** on PATH. The clone and dependency download
are visible steps here; the harness does not download or build the client.

```bash
set -euo pipefail
demo_client_root=/absolute/new-demo-client-build
demo_client="$demo_client_root/client-53bdbd0"
demo_client_pin=53bdbd03136875f952af184fce7b3c7af8f74e96
mkdir "$demo_client_root"
test "$(node --version)" = v24.16.0
test "$(pnpm --version)" = 11.20.0
git clone https://github.com/gastownhall/bdp.git "$demo_client"
cd "$demo_client"
git checkout --detach "$demo_client_pin"
test "$(git rev-parse HEAD)" = "$demo_client_pin"
git status --porcelain > "$demo_client_root/source-status-before.txt"
test ! -s "$demo_client_root/source-status-before.txt"
git diff --exit-code HEAD > "$demo_client_root/source-diff-before.txt"
pnpm install --frozen-lockfile --ignore-scripts > "$demo_client_root/install.log" 2>&1
pnpm exec tsc -b packages/client > "$demo_client_root/build.log" 2>&1
# Dependencies and generated dist files are expected; tracked source must not change.
test "$(git rev-parse HEAD)" = "$demo_client_pin"
git diff --exit-code HEAD > "$demo_client_root/source-diff-after.txt"
test -z "$(git status --porcelain --untracked-files=no)"
python3 - "$demo_client" "$demo_client_pin" <<'PYMANIFEST'
import hashlib, json, pathlib, sys
checkout = pathlib.Path(sys.argv[1])
files = [
    "pnpm-lock.yaml",
    "packages/client/src/index.ts",
    "packages/client/dist/index.js",
    "packages/protocol/src/index.ts",
    "packages/protocol/dist/index.js",
    "schemas/bdp-v0.schema.json",
]
manifest = {
    "source": "gastownhall/bdp", "commit": sys.argv[2],
    "source_method": "fresh clone, exact detached pin, clean tracked source",
    "node": "24.16.0", "pnpm": "11.20.0",
    "install": "pnpm install --frozen-lockfile --ignore-scripts",
    "build": "pnpm exec tsc -b packages/client",
    "passed": True, "http_interop": False,
    "files": {name: hashlib.sha256((checkout / name).read_bytes()).hexdigest()
              for name in files},
}
with (checkout.parent / "client-build.json").open("x") as output:
    json.dump(manifest, output, indent=2)
    output.write("\n")
PYMANIFEST
```

`passed: true` is written only after those build and source checks succeed. It
means the client build succeeded, not that HTTP interoperability has passed.
The manifest contains hashes of exactly the six named files. Retain it and the
build/source logs; the harness validates the files and saves the provenance.

Start a disposable ordinary Dolt **2.1.8** server with its own empty data
directory. In another terminal, reserve that server for this run:

```sh
dolt sql-server --host 127.0.0.1 --port 3307 --data-dir /absolute/disposable-dolt
# In another terminal:
python3 scripts/graph-mixed-demo.py --bd /absolute/demo-bin/bd \
  --backend server --server-port 3307 \
  --bdp-checkout /absolute/new-demo-client-build/client-53bdbd0 \
  --node /absolute/path/to/node \
  --output-dir /absolute/new-server-recording
```

The defaults are 60 seconds per CLI command and 300 seconds overall. They were
ample for the recorded runs; slower machines can explicitly increase
`--command-timeout` and `--total-timeout`. The receipts retain the chosen bounds.
Each completed backend story records **27 CLI calls: 26 successes and one
expected stale-revision refusal**. The server story additionally records the
independent public client's HTTP requests.

The harness chooses a free HTTP port before normal initialization and persists
that Scope identity. It stops its HTTP server and client; the caller stops Dolt.
Provision different databases serially. Disposable `/private/tmp/bd-c0-*`
(or `/tmp/bd-c0-*`) workspaces and the server's `demo_bd_c0_*` databases remain for
inspection; the caller cleans up these leftovers after preserving the evidence.
A zero active-child count does not mean the database or workspace was deleted.

Open `recording.md` in the new output directory for actual commands and expandable
recorded output. Preserve these adjacent artifacts:

- `summary.json` and `commands.json`, plus each command's receipt/stdout/stderr.
- `provenance.json`: actual binary hash, isolated environment and execution bounds;
  its `harness_sha256` names the shared `graph-c0-smoke.py` capture helper.
- `demo-source.json`: hashes of the demo scripts and their shared Python helpers.
- `expected.json`: CLI records used for comparison with the public client.
- For the server story, `client-results.json`, `client-provenance.json` and the
  `public-client/` and `serve/` process receipts and logs.

These are textual recordings, not terminal videos. Their headers disclose the
preview status and source/binary/backend identity. The separately delivered
`crew-demo.zip` preserves both reviewed recordings and the evidence next to them.

The implementation plan and review remain in [fork PR18](https://github.com/donnabox/beads/pull/18).
CLI proposal feedback belongs in [upstream #6703](https://github.com/gastownhall/beads/issues/6703).
