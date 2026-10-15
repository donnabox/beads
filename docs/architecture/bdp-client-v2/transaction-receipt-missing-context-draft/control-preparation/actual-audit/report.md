# Receipt omission control: post-run evidence audit

The one authorized control produced the intended eight named context-tuple failures in one selected case, with 13 actual HTTP responses and completed cleanup. The original wrapper nevertheless exited 1 with CONTROL_NOT_ACCEPTED: “Shim changed beyond bound path spelling.” Its sourceUnchanged:false field reflects that graph-check exception, not an observed source change. Original exit, adjudication, log and result files are preserved unchanged. Root acceptance is separately recorded in ../../root-control-acceptance.json. No rerun or restoration occurred.

This audit was performed by the preparation/wrapper author; it is an evidence reconciliation, not another independent code council.

## Exact attribution and stability

audit.json SHA256 67425cc07a53cdbe831ed24b24025c4bebf5303250dd9eed2b5cb1f16f8095e9 binds the complete details and original evidence hashes. All 1106 positive source/emitted files and 1106 copied counterparts match the prepared manifest, including the declared two-site copied HTTP mutant. Eight draft source hashes and full patch 837bacdc192bcf0401171ecc9a8312ae034beabed7b69620b6a3beaef0cb9de7 remain exact. Dependency inventory retains 2066 members.

Fourteen physical generated .bin shell shims changed: nine shared-store members and five additional local/root shims. Each retained before.sh reconstructs the exact pre-run SHA and byte length. Four required the corresponding existing sibling checkout bytes with its root replaced by the positive workspace root; the other ten reconstruct by replacing the copied root in current bytes. The hypothesis of an old Event-copy prefix was not the explanation for those four.

Four shared-store shims additionally rewrote executable target spellings through a package-local symlink to a versioned pnpm-store path: rolldown, two vite shims, and why-is-node-running. Their five executable branches each resolve to the same actual files and hashes. All fourteen match after normalizing only actual NODE_PATH, target-comment and executable-path spellings; other shell lines match. This is effective target equality, not dependency byte equality, and is consistent with tooling shim regeneration without attributing the writer from unobserved system calls.

There are 38 logical symlink-text changes (35 physical links), all with identical before/after effective realpaths. Every changed shim, reconstruction provenance, exact before/after hash, target and patch is retained beside audit.json. The original wrapper deliberately did not allow the four executable-path rewrites, so its failed verdict is not edited or retroactively called a pass.

## Observed control and cleanup

The actual Vitest process exited 1; supervisor exited 1 after 2.077155 seconds. One selected case failed with exactly eight expected labels, 76 cases were skipped, and no unhandled errors were reported. Response statuses were 200,200,404,404,404,404,404,404,200,404,404,401,401. The seed context was pos-1 and advancement pos-2; all eight selected 404 context tuples were absent under the mutant. Both anonymous responses remained contextless. GET/HEAD representation metadata and empty HEAD bodies reached their hard assertions.

The safe retained response record reports 13 attempts/responses, eight tuple checks, two registered and two completed cleanups, primaryOrCleanupFailureCount=0 and deadlineExceeded=false. Full bodies were checked in memory; the retained trace contains selected safe headers and body lengths, not full-body artifacts.

Sampled PIDs 8882,8884,8888,8893,8898,8908 and process group 8882 were independently absent. The owned runtime directory and transient entry were gone; only node-compile-cache remained. No timeout or forced cleanup occurred. Two failed read-only audit assembly attempts are disclosed in audit-preparation-note.txt and were not extra control executions.

The evidence supports root's separate acceptance of the intended control. Installed validation remains separate. A future preparation may invoke pinned Node directly on the Vitest CLI to avoid pnpm exec shim regeneration; no current script or dependency was changed for that suggestion.
