# Actual static-prefix audit

Confirmed: bootstrap installation, tool binding and typecheck passed; format-check failed. No lint, build, source callback, package or installed result exists in this attempt. No semantic correction is established by this failure.

| Stage | Raw exit / accepted | Elapsed seconds |
|---|---|---:|
| workspace-install | 0 / true | 4.099605 |
| bind-workspace-tools | 0 / true | 8.696362 |
| typecheck | 0 / true | 5.233991 |
| format-check | 1 / false | 4.290834 |

All four actual log hashes independently match their exit records; all report unchanged source, no forced cleanup/errors, no remaining processes or runtime directories. Fresh read-only process census found all 45 unique sampled PIDs absent. Independently rehashed all 679 source files against original author cb403c8a binding and all 11 actual workspace tool hashes/realpaths. Actual workspace-tools binding13d53c8ed5eb6f3caf3f247dbc798a60c266debf88ed4dcf26e53b0bde2c1831 matches the root gate. Typecheck log is the successful `tsc -p tsconfig.check.json` command; this is not runtime coverage.

Format log: SHA2562aea74c1948a420af5b218a16b92070de5c82e50230a2d17fcfb83c89f3d3f34, 33285 bytes. Exit record000ef2e8f6a1bc20cb8375e57177d3c878c0f1085dd4e55ed9964f5cf6c314f6. It checked six files, applied no fixes and reported exactly five file-level formatting diagnostics:

- apps/bdptest/src/transactional-event-sse-installed.test.ts (log1)
- packages/server/src/transaction-event-sse.test.ts (log167)
- packages/server/src/transaction-events.test.ts (log327)
- packages/server/src/transaction-service.ts (log406)
- packages/server/src/transaction-store.ts (log449)

transaction-http.ts was checked but not flagged. The first two diagnostic diffs explicitly truncate 20 and 830 additional lines, respectively. Five diagnostics are not five edit sites; this log cannot supply a complete site inventory. The earlier Claude formatting prediction is now superseded by an actual measured failure, without claiming its predicted site inventory was exact.

Recommend a separately root-authorized single bounded pinned Biome `format --write` targeting exactly the five listed files, with the existing configuration/Node24.16/tool identities and no automatic retry. Do not include transaction-http.ts or run repository-wide formatting. Preserve this entire failed attempt. Freeze a distinct formatting successor with complete before/after snapshots, exact correction and full patch, all679 hashes, complete syntax/token equivalence review (including any formatter punctuation), and historical-prefix/preservation accounting. The full resulting diff, not truncated diagnostic output, must establish layout-only scope. Production service/store formatting must receive the same semantic-equivalence scrutiny as tests. No completed substantive CODE verdict may be relabeled as review of an unexamined semantic change.

Only the actual unchanged dependency/bootstrap installation and tool-capture origins may be proposed for explicit reuse. Original successful typecheck remains evidence for its original source, not a pass for reformatted successor bytes; all 19 later stages and four required checks must be freshly bound/executed under root release. No deadline, workload, assertion or failure requirement should change.

At inspection the gate still enumerates only the five authorized static stages and has no runtime stages; actual tool pin is populated, consumer lock/proof/inventory remain null. The on-disk static release remains true and includes lint, but no lint attempt occurred after the failed format gate. This audit grants no continuation: root should close/rebind the failed attempt and separately authorize the exact source correction and successor gates. No helper/product execution or source mutation was performed for this audit.
