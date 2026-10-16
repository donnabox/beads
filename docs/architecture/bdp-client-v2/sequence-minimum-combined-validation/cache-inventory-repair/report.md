# Narrow generated-cache inventory correction

Confirmed the review finding: Vitest writes `node_modules/.vite/vitest/.../results.json`; the old mirror builder pinned it, so a genuine control run (and later positive run) could fail an otherwise valid byte check. The original packet and native source audit are preserved under `before/`.

Only `control.py` enumeration changes: skip real directories named `.vite` or `.vite-temp` immediately under `node_modules`. Keep directories of those names elsewhere and keep all symlinks pinned. The identical filtered file map supplies both positive and mirror hashes. No product, dependency, executable, emitted mutation, request, selector, limit, or cleanup oracle changes. No source/worktree edits or execution occurred. Only the example's control hash and preparation inventory were rebound; live gate remains absent.

`correction.patch` contains the full executable delta. Python syntax was parsed as data. Independent reviewer closure remains pending; this report does not release execution.
