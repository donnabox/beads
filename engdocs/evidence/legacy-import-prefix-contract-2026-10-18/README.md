# Missing-prefix oracle follow-on to frozen import2b

See [diagnosis.md](diagnosis.md), [frozen-failure.md](frozen-failure.md) and [receipt-manifest.json](receipt-manifest.json). The preserved owner attempt is RED. [lower-failure.log.gz](lower-failure.log.gz) is a lossless copy of its synthetic lower-suite log; compressed bytes are not a new execution.

A clean private child changes only the real-engine test oracle and adds a no-write-stage assertion. Production and shared error contracts are unchanged. Source-only formatting/changed-unit whitespace and retained-receipt hashes are checked; all successor builds/tests/lint/generators/check modes/CLI/engine/DB work remain UNRUN and Janet-owned. Fresh review and exact-source rerun are pending.
