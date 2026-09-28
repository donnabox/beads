# Verified evidence checkpoint — September 28, 2026

The current packet is recorded against qualified runtime **9c86d6d1559ffcfcd9770b64e17a7f1f654690b3**. Both engines passed 32 CLI calls, including one exact exit-4/empty-stdout stale refusal each; server mode passed 10 real HTTP requests through the independent pinned client and conditional GET. Offline verification checked 64 CLI receipts, two process receipts, complete predecessor-derived state, 13 saved complete records per engine, exact saved reads, current graph results, raw HTTP bytes and zero children.

The recorder scripts are unchanged from reviewed PR59 commit `ec11dbe7d5d17d5748cca71e951141fa199ea64b`. The four-file source snapshot used by the verifier is retained separately from these updated presenter documents. The runtime's [complete Linux qualification](https://github.com/donnabox/beads/actions/runs/36419928551) passed 6,017 command receipts, 1,062 graphstore gates, 144 compiled roots, nine invocations across seven packages, 145 HTTP observations and exactly 22 inherited non-graph exclusions. The candidate advanced to that exact runtime; the failed standalone PR57 run remains failed. PR60's later workflow-only isolation change is outside this runtime and remains unqualified.

The packet's `recording-verification.json` and `runtime-linux-verification.json` keep presentation evidence separate from runtime qualification. Full Memory, native/public History, complete authenticated HTTP Write, adoption and human review gates remain open.

---

PR49's older `b7bf5040…` map and recording remain unchanged. This packet is a separate verified recording of the later runtime, with the boundaries below.

[Human guide](GRAPH_CURRENT_WORKFLOW_DEMO.md) · [Recorder](../scripts/graph-current-workflow-demo.py) · [Independent client](../scripts/graph-current-workflow-demo-client.mjs)

| What the new story directly exercises | Actual existing implementation seam | What it does not establish |
|---|---|---|
| Normal fresh initialization and separate-process current reads | Existing graph admission and persisted Scope/authority checks; embedded and ordinary-server Dolt | Existing-workspace adoption, restore/cutover or Jim integration |
| Due Issue creation and selection | `graph_preview_issue.go`, due helper, checked list adapter and shared native Issue writer | Defer/lazy wake, arbitrary status, scheduling, full Issue workflow compatibility |
| Blocking prerequisite and close→ready | Native Dependency writer plus graph mapping/owned retention; existing close, ready and blocked adapters | Repair, all Dependency Types, labels-held work or external blockers |
| Memory create, selected content read and retained body | Current Memory writer, `recall`, checked current/exact readers | Complete Memory representation, structured full recall, metadata/Inception/derivation |
| Issue→Memory context and owned Memory→Memory rationale | Existing informational Link writer with canonical Link identity and source ownership | Alias resolution, external/pinned endpoint traversal or arbitrary Type installation |
| Ordered Memory/Link patch and stale refusal | `internal/graphpatch` plus existing single Memory/Link transactions; dual guard for owned Link | HTTP Write, unsupported RFC operations, unbounded patches or arbitrary final properties |
| Current graph navigation | `cmd/bd/graph_preview_generic.go` admission/rendering and `cmd/bd/graph_preview_generic_projection.go` pure projection over one checked `CurrentSnapshot` | Durable public traversal DTO, remote traversal, historical incoming completeness or public graph-query endpoint |
| Exact previous state and comparison | Existing retained version reader and complete properties/owned comparator | Native commit instant, chronological History, restore or complete BDP History |
| Independent server BDP Read | Existing Read router and unchanged pinned public client; real GET/HEAD transport | Embedded HTTP serving, HTTP mutation, deployed authentication or durable write outcomes |

The generic view is a private summary of identity, title, version and Link endpoints, with per-node and per-Link attribution (actor, status and observed wall-clock `recordedAt`). It has no body, excerpt, Link note, metadata value or Issue long text. Summary does not mean anonymous. Depth exposes unexpanded frontier; node/Link/output caps refuse the result. The current snapshot still reads within its whole-workspace 1,000 live Bead/Link records plus the four installed descriptors, and 16MiB acquisition restrictions, and output has a separate 1MiB bound. It is not a claim that the engine never acquired Memory bodies internally.

Patch final shapes remain the existing preview contracts: Memory title/body and informational Link optional string note. The pure evaluator's ordered add/remove/replace behavior is broader than those final shapes, but the actual writer revalidates before acceptance. Source ownership, guards and retained versions are real; they are not replaced by a pre-read/compose/write demonstration.

## Evidence organization and final acceptance

Verified capture count is 32 CLI calls per engine: 31 successes plus one stale guard refusal. Server mode adds its long-lived `bd serve` and Node processes, with an exact final ten-HTTP-request gate and a separate in-flight safety cap of 16. These counts are established by the successful saved captures and independent verification; future failed attempts must retain their available receipts. The recording must provide complete expected records and visible source pin/binary hash, not merely pretty terminal output.

Each backend packet contains `recording.md`, `commands.json`, raw per-command stdout/stderr and receipts, `expected-records.json`, public-client expected state, `demo-source.json`, `artifact-hashes.json`, runtime-evidence copies and `summary.json`. Server mode adds `client-provenance.json`, `client-results.json`, raw HTTP body artifacts and listener/client process receipts. The recorder uses relative artifact links; retain whole directories when sharing. Absolute runtime/workspace paths are honest provenance, not a requirement for following those relative receipt links.

The runtime evidence is caller-supplied clean-source/install evidence with checked hashes, not an independent source attestation. `bd version --json` supplies a build label and optional commit; those must agree with the supplied full commit and installed binary. Root should not fabricate a hash-like Build label merely to satisfy this recorder: install from the selected final source with the ordinary explicit version/Build/Commit stamping and retain that command/result.

A successful recording remains `qualification:false`. Final-source baseline/lint, actual review, complete required Linux evidence and independent receipt verification are separate gates. Source or local-only success must not promote the candidate by itself.

## Bounded source review and corrective disposition

The original four-file target is preserved under the local `demo-reviewed-original` evidence directory. Actual Claude review and root's source review were read before these corrections. This section preserves the source-review dispositions from before capture; the verified execution is reported above. Automated review does not replace substantive human approval.

The earlier native findings are closed in source: operator interruption stops the backend sequence; predecessor-derived whole-record transition checks cover Memory body, Link note, owning Memory and the explicit close reason/native closure fields; all domain writes specify `--actor demo`; and the stale attempt independently requires `revision_conflict`, exit 4 and exactly zero stdout bytes. The previous map's statement that the stale assertion remained pending was stale itself and is removed.

Corrected recorder SHA-256: `1ef382119833572f2857fc2cfe30accf1e474ec1d634a8b8bf130875344aa7dc`. The final packet source manifest and retained receipts, rather than a previous review's digest, identify what was actually run. No cancellation, database or HTTP execution has occurred during this corrective source pass.

Claude finding dispositions:

1. **Stale review status/hash:** corrected above; no remaining pending stale-refusal claim.
2. **Helper admission:** `--helpers-dir` is required, absolute and validated for both helpers before loading either.
3. **Deadline:** retain the finite 300-second per-backend default. Explicit pre-start/pre-wait deadline checks and timeout diagnostics name the remaining budget; the guide documents overrides. No unmeasured increase to 900 seconds.
4. **Scope port:** recheck the persisted port immediately before serve; diagnose collision without choosing a new Scope. The final bind race remains an honest limitation.
5. **HTTP chapter:** process chapter metadata has process names, not an invented 33rd CLI receipt. One recorded title supplies the displayed heading.
6. **Failed HTTP evidence:** listener and client receipts, stdout and stderr are linked/displayed even without `client-results.json`.
7. **Streams:** stdout and stderr have separate labelled fences and byte counts, including zero-byte stdout on the expected refusal.
8. **Python timestamps:** RFC3339Nano-shaped UTC values are checked with a 0–9-digit-fraction expression; no version-dependent `fromisoformat` fractional parsing.
9. **Count gate:** retain the exact ten-request final assertion. A modest 16-request in-flight cap preserves bounded evidence when extra requests occur. This does not quietly broaden the accepted proof.
10. **Request input:** URL, method and headers normalize both Request objects and plain inputs; missing ETag baseline reports the actual observed URLs separately.
11. **Client provenance diagnostics:** manifest and artifact file-presence checks precede reads/hashing. The successful pinned sibling build-manifest requirement stays in force.
12. **Guide completeness:** all displayed command blocks are explicitly representative; omitted verification calls are named once above the chapters. Actual receipts supply all 32 commands.
13. **Projection disclosure/location:** the pure projection file is identified, and both documents state that summaries carry attribution.

Root's embedded-provenance correction is also applied: embedded child arguments clear caller server port/root before Capture. Both modes still use separate disposable workspaces, and the server remains caller-owned. The docs' intended locations are `engdocs/`; their script links target `../scripts/` and the command recipe starts at repository root.

Source-checked boundaries remain: records-json list uses the graph preview envelope while version JSON is ordinary; exact recall reads stdout bytes; public Read maps claimed actor to principal/status and omits unknown attribution; Issue ownership includes the explicit empty Dependency group while Memory groups its populated wildcard Types. Two-step discovery plus eight subsequent requests gives the expected 10. All writes remain installed CLI; the final conditional request uses real HTTP directly.

No deployed-authentication, arbitrary corpus, cancellation-execution or complete failure-injection proof was run for this packet. Existing product qualification owns forced concurrency, unknown commit, schema corruption, allocation bounds, complete selector/pagination and writer-error gates. Link that exact-source evidence separately; this compact presentation cannot substitute for it.

## What should guide implementation after this evidence step

The next meaningful decisions remain complete Memory representation/immutable-Type rollout, stable deployment principal and durable write outcomes, representative permitted adoption input, Jim's named target/recorder reconciliation, and held PR44/45 review. Approved existence of metadata, registered Link Types and the BDP two-identity design are not being reopened. Native History's missing atomic stamp primitive is a completed stop; this recording supplies no new primitive and does not justify another investigation loop.
