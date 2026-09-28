# A release plan you can follow, change and revisit

**Prepared, unexecuted successor recording.** Root must choose and qualify the exact runtime commit, supply its clean-source/install evidence and verify the resulting captures before this packet is presented as working evidence. Neither this guide nor the recorder qualifies the product.

This is separate from PR49's verified older `b7bf5040…` recording. Keep that packet unchanged. The new story adds due-date selection, blocked inspection, ordered Memory/Link editing and current generic navigation to one small workspace.

[Recorder](../scripts/graph-current-workflow-demo.py) · [Independent client](../scripts/graph-current-workflow-demo-client.mjs) · [Implementation map and preparation review](GRAPH_CURRENT_IMPLEMENTATION_MAP.md)

## What the crew will see

A release Issue waits on a verification Issue. A Memory explains the release plan; another explains its rationale. The Issue points to the plan, and the plan points to the rationale. These are explicit stored Links, not relationships guessed from the text.

We follow that chain without automatically printing the Memory bodies. Then we deliberately read the plan, edit its body and its rationale Link, and retrieve the previous exact state. Completing verification makes the release ready. In the server recording an independent BDP client sees the same current records.

The intended presentation is five to eight minutes. The source currently schedules **32 fresh-process CLI calls per backend: 31 successes and one expected stale-revision refusal**. Server mode also owns a `bd serve` process and a Node client; ten HTTP requests are expected. Counts remain source estimates until successful receipts establish them.

All command blocks below are representative, not a complete replay script. The recorder's receipt index is authoritative. The abbreviated blocks omit `version --json`, `status --graph`, the deliberate stale patch and unchanged reread, and the final `show` calls for the release, current rationale Link, context Link and rationale Memory. Every domain write in the recorder explicitly supplies `--actor demo`; the displayed examples use that same actor. All recorded body bytes and revision tokens come from the recorder and actual preceding results.

## 1. Create work with a plan and a reason

Normal initialization creates the workspace. Embedded storage is the default; the separate server recording supplies an existing disposable Dolt server and a unique database. Both modes run sequentially.

```sh
bd init --graph-mode link --scope-url "$scope" --prefix demo --non-interactive --skip-hooks --skip-agents
bd create 'Release to a small group' --id beads/work --due '2030-01-02T12:00:00Z' --actor demo --json
bd create 'Verify the deployment' --id beads/verify --actor demo --json
bd dep add beads/work beads/verify --actor demo --json
bd remember 'Ship after verification. Keep the rollout reversible.' --id beads/plan --title 'Deployment plan' --actor demo --json
bd remember 'Small releases make rollback practical.' --id beads/rationale --title 'Why start small?' --actor demo --json
bd link beads/work beads/plan --resource-type "${scope}types/preview-related-v2" --id links/context --properties '{"note":"Read the plan before release"}' --actor demo --json
bd link beads/plan beads/rationale --resource-type "${scope}types/preview-related-v2" --id links/rationale --properties '{"note":"Original rationale"}' --if-source-revision "$plan_revision" --actor demo --json
```

These are illustrative command spellings; the recorder captures output and obtains each revision from the actual preceding response. Its Memory bodies deliberately include one terminal newline and exact recall verifies those bytes. No schema script, SQL fixture or hidden bootstrap is involved. The installed Type belongs to this persisted Scope; the story does not install custom Types.

## 2. Find the work and follow its context

```sh
bd list --format records-json --due-before '2030-01-03T00:00:00Z'
bd blocked --json
bd ready --json
bd graph beads/work --view generic --direction out --depth 1 --json
bd graph beads/work --view generic --direction out --depth 2 --json
bd recall beads/plan --readonly
```

Due date answers when; the Dependency answers what must finish first. Verification is ready, while release is blocked. Depth 1 reaches verification and the plan, and exposes the plan as frontier because its rationale Link remains unexplored. Depth 2 includes rationale and completes this outgoing local component. It does not claim a complete external graph or History.

Generic results contain node titles, identities, saved-version tokens and explicit Link identities. Both nodes and Links also carry recorded attribution: actor, status and observed wall-clock `recordedAt`. They omit bodies and notes; this content summary is not an anonymity or authorization boundary. `recall` is the separate, deliberate request for content.

## 3. Change the explanation without losing the earlier one

```sh
bd update beads/plan --if-revision "$plan_old" --actor demo --patch '[{"op":"remove","path":"/body"},{"op":"add","path":"/body","value":"Verification first; then ship to a small group.\n"}]' --json
bd update links/rationale --if-revision "$link_old" --if-source-revision "$plan_body_revision" --actor demo --patch '[{"op":"remove","path":"/note"},{"op":"add","path":"/note","value":"Verification supports a small reversible rollout"}]' --json
bd show beads/plan --version "$plan_old" --readonly --json
bd recall beads/plan --version "$plan_old" --readonly
bd compare beads/plan --from "$plan_old" --to "$plan_new" --json
bd show links/rationale --version "$link_old" --readonly --json
```

The operations apply in order through one writer transaction. The Link edit guards both the Link and its owning Memory. It advances the Memory's saved linked state without changing the body again.

The recording also attempts a stale plan edit. That must refuse with revision conflict and no success output; a new process then reads the unchanged accepted state. The comparison reports body and owned-Link changes. Old reads return the old body and complete old Link. These are exact saved-state reads and comparison, not chronological History, restore or a public History profile.

## 4. Finish verification; keep the relationships

```sh
bd close beads/verify --reason 'Verification passed' --actor demo --json
bd blocked --json
bd ready --json
bd graph beads/work --view generic --direction out --depth 2 --json
```

Now no Issue is dependency-blocked and the release is ready. The Dependency remains a graph relationship. Closing a prerequisite changes workflow status; it does not delete the plan or its rationale. Final explicit reads verify that the release, current Link, context Link and rationale did not acquire unrelated changes.

## 5. Independently read the same graph through BDP

Server mode starts `bd serve --readonly --addr 127.0.0.1:PORT` against the same workspace. The persisted Scope uses that listener's selected loopback port. The unchanged independently built public BDP client at `53bdbd03136875f952af184fce7b3c7af8f74e96` discovers Read, retrieves all four Beads, reads Bead and Link collections and plan incident Links, and checks an unchanged plan ETag yields 304.

The client compares complete public properties, revisions, attribution and owned Links with the saved CLI records. The expected ten-request shape is two discovery requests, four Bead reads, two collections, incident Links and one conditional GET. The final conditional probe uses real HTTP directly; the other nine requests run through the public client's discovery/Read API. The client has an in-flight safety cap of 16 requests and requires exactly 10 at completion; a changed request shape fails qualification of the recording while preserving available evidence. There are no HTTP writes or generic HTTP traversal endpoint in this story.

## Capturing and sharing

From the repository root, use Python 3.9 or newer. `--helpers-dir` is required and must contain both existing capture helpers. These files are intended for `scripts/`; this guide and the map are intended for `engdocs/`. Root must supply an installed executable from the chosen qualified commit, a clean-source/install manifest with checked artifact hashes, existing helper scripts, and—for server mode—the already built pinned client and caller-owned disposable server. No client build or download happens inside the recorder. Produce the hash-stamped runtime with the repository's normal `make install` or `make build` procedure and retain the actual install/build receipt; a default `dev` build label is refused rather than fabricated.

```sh
python3 scripts/graph-current-workflow-demo.py \
  --bd /absolute/qualified/bd --runtime-commit FULL_40_CHARACTER_SHA \
  --runtime-evidence /absolute/runtime-evidence/manifest.json \
  --helpers-dir /absolute/qualified/source/scripts \
  --backend both --server-port PORT --server-root /absolute/disposable/server \
  --bdp-checkout /absolute/pinned/client-checkout --node /absolute/node \
  --output-dir /absolute/new/recording
```

This command is a preparation recipe, not a reported run. The caller-supplied runtime manifest requires `commit` (the full selected SHA), `source_clean:true`, `install_passed:true`, `binary_sha256`, and an `artifacts` map of relative paths to SHA-256 values. That map must include `git-head.txt` (the same full SHA), `git-status.txt` (exactly empty from the clean source check at build), and nonempty `install.log`. The recorder copies and verifies these artifacts; it does not independently attest the original build or certify the manifest's asserted success.

The default wall budget remains 300 seconds **per backend**, with a 60-second individual CLI bound and a public-client wait capped at 90 seconds or the remaining backend budget. Use explicit `--total-timeout`/`--command-timeout` overrides only when the measured environment warrants them. The recorder reports exhausted remaining time distinctly; it does not restart work or hide a missing client result. Keep the chosen loopback Scope port free between initialization and serving. The recorder rechecks that exact port before launching the listener and refuses a collision; the unavoidable final bind race can still fail, and it never rewrites the persisted Scope to retry another port.

Each backend creates its own disposable workspace/database. The caller owns the server and later cleanup of retained data; the recorder must stop and reap its CLI/listener/client children even on interruption.

The recording labels stdout and stderr separately, including empty streams. Its server chapter links and displays both listener and public-client receipts/streams even when the client fails before producing a result JSON. Embedded provenance carries no caller-server port/root.

After execution and independent verification, share the entire output directory. Start at `embedded/recording.md` and `server/recording.md`; their relative links lead to exact argv/exit/stdout/stderr, complete saved records, source hashes and raw HTTP evidence. A `passed` recording still says `qualification:false`: its runtime qualification belongs to the separate exact-source evidence. Do not remove that distinction when presenting it.
