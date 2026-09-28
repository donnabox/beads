# A release, a plan, and the reason behind it

This is a five-to-eight-minute walkthrough for people who have not been following
the implementation. Use the verified recording first; the same script can create
a new recording from a fresh disposable database.

The story uses the qualified Beads candidate **b7bf5040**. Claim and notes append
are separate pending drafts, PR47 and PR48, and are deliberately absent here.
The [implementation map](GRAPH_IMPLEMENTATION_MAP.md) lists the wider supported
surface, where the data lives, and what remains unfinished.

## Set the scene

“We have work to do, a plan for doing it, and a reason for the plan. The work is an
Issue. The plan and rationale are Memories. All three are Beads, so we can link
them and inspect the relationships without pretending a document is a task.”

The recording creates two Issues, `beads/release` and `beads/verification`, and
two Memories, `beads/plan` and `beads/rationale`. It uses normal `bd init`, `create`
and `remember` calls. Every CLI call is a new process. There is no SQL seed or
hidden bootstrap operation.

## 1. Follow the context

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

Open `edit-memory`, `show-current-plan`, `recall-saved-body` and `compare-plans`.

“The plan now requires a rollback rehearsal too. The update checks the revision
we read, so a stale writer cannot silently overwrite a later change. A fresh
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

The script verifies current properties and owned-Link identities, endpoints,
revisions and properties against the CLI results. It does not compare the BDP
attribution projection. It retains actual request URLs, status codes and response
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

Build the runtime from the **exact qualified commit** in its own checkout:

```sh
make install-force INSTALL_DIR=/absolute/demo-bin
```

Run the demo script from this documentation branch. It requires the installed
binary's `b7bf5040b` build label and records its SHA-256; retain the installer/source
receipt too. It refuses an existing output directory and retains failures.

```sh
python3 scripts/graph-mixed-demo.py --bd /absolute/demo-bin/bd \
  --backend embedded --output-dir /absolute/new-embedded-recording
```

For the BDP chapter, start a disposable ordinary Dolt 2.1.8 server with its own
empty data directory. Build the public BDP client at
`53bdbd03136875f952af184fce7b3c7af8f74e96` using Node 24.16.0 and pnpm 11.20.0,
as described in the [HTTP proof guide](GRAPH_BDP_READ_HTTP.md). Supply the same
validated archive build and adjacent `client-build.json` manifest used by that
proof; this small presentation script intentionally does not create or download
client builds. The manifest records the commit, successful build and file hashes.

```sh
dolt sql-server --host 127.0.0.1 --port 3307 --data-dir /absolute/disposable-dolt
# In another terminal, with that server reserved for this run:
python3 scripts/graph-mixed-demo.py --bd /absolute/demo-bin/bd \
  --backend server --server-port 3307 \
  --bdp-checkout /absolute/client-53bdbd0 --node /absolute/node \
  --output-dir /absolute/new-server-recording
```

The harness chooses a free HTTP port before normal initialization and persists
that Scope identity. It stops its HTTP server and client; the caller stops Dolt.
Provision different databases serially. Open `recording.md` in the new output
directory: each chapter has the real command and expandable recorded output.
`summary.json`, `commands.json`, per-command receipts, `client-results.json` and
`client-provenance.json` preserve the checks and provenance. These are textual
recordings, not a fabricated terminal video.

The implementation plan and review remain in [fork PR18](https://github.com/donnabox/beads/pull/18).
CLI proposal feedback belongs in [upstream #6703](https://github.com/gastownhall/beads/issues/6703).
