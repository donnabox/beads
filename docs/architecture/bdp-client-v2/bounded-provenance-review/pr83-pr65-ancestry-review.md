# PR83 ancestry merge freeze

Clean source head: `18415bcdcbf9546b5ab3f7fc0cfd78f8b8f6eb3d`.
Checkout: `/Users/dbox/repos/wt/vickie/bdp-provenance-runtime-integration`.
Branch: `codex/vickie-bdp-provenance-runtime-integration`.
Ordinary parents: `4496fb044e6d0212ba818d5f74f72b0be1b23b8c` and root PR65 `a68db90f995c1d3317c19acb0b6dcdf63cd21b0e`.

Only four files differ from4496: imported Read evidence artifact and capability file; independently reviewed projection test pin bd7da0fc3f8082e0bfb6f4969c04409f2a610e12408dbc6a0bd14b4b351127cc; graph-policy review-key ordering.

The sole conflict was graph-policy JSON key order: both diagnostic and nonempty-output justifications and every hash were identical. Took incoming ordering, losing no content. All other files merged automatically. Verified evidence/capability files byte-identical to incoming a68, with constant e425bb5c8256e306b251800ba1ba738a4f82a1ed. This is coherent PR64 historical evidence, not observed PR83 or PR65 common-provenance qualification. No artifacts were relabeled or hand-edited.

Verified byte-identical preservation from4496 of four-part spec, private RU/TX numeric formatter, Transactional evaluator/store, global recovery-store and development installation/lineage fences. The automatic Transactional catalog merge produced no net difference.

Build/typecheck/boundaries/format/diff checks pass. Focused pin/catalog/graph-policy suite: **46 passed /5files /1.98s**, including all seven catalog ownership audits. Logs and JSON: pr83-pr65-ancestry-*.

Committed evidence:verify fails as expected: imported historical projection26db67b2 versus currentbd7da0fc, reseal required. This is not a failed runtime observation and is not waived. No fresh PR83 cohort or full-suite repeat occurred. No outward writes. Source is frozen pending root’s final PR65 evidence ancestry and separately authorized PR83 qualification.
