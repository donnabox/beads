# Receipt-control binding successor — closed v4

Only command output/script paths move from J-v3 to J-v4 and the closed example's bypass patch filename/hash/expected output hash changes. All six Python scripts, prospective binding, selectors, requests, deadlines, cleanup and package logic remain byte-identical. Exact changes and line/hash inventories are in successor.patch and changes.json.

Actual source remains final-draft manifest ff9c63417fbb348bf8ad9863483096ccc0329beb5c05b532333cf75cd0726a38 / diff891a3dd1d17d9f754dc81ef6223bdb61de4c330e15bc5ab7760559ca5c563106. All seven actual source hashes match. No W write occurs.

The corrected external patch is copied exclusively to the original DRAFT directory as bypass-receipt-root-minimum-v2.patch, preserving the old patch and satisfying the existing controls.py parent guard. SHA dca316bf79495061823fb12ed9ffd0c369e51ac2ca32858451c7301b8a691fff; expected negative transaction-http.ts SHA9da034938dc1a48dd48b961a0ec9fa518dc954d1bddcb2ab38f321a3da13a131. Standalone text reconstruction confirms the existing exact single-hunk parser can reproduce it. No derivative checkout or runtime was created.

The J-v3 restore-command named500 result and clean evidence remain in J-v3. Do not repeat restore-command in v4. The J-v3 receipt derivative failed its build with TS2367; no receipt test ran. That failure and its original derivative remain preserved, not reclassified as a sensitivity result.

Root's next release is only the v4 bypass prepare/setup/build/test/review sequence, with the new actual snapshot hash supplied after prepare. Incremental runtime budget is one case, expected2/max6HTTP, one owner/listener/runtime store, zero installed children, unchanged120-second command ceilings. The ledger retains equivalent command definitions, including restore, for structural continuity; approved IDs are empty and root must not release restore again. Cumulative controls acceptance must explicitly cite restored-command evidence from v3 and corrected receipt evidence from v4. Package/consumer/TX remain held until that acceptance and the actual clean final source/provenance bindings exist. No evidence is copied into v4.

No execution-gate.json or run directory is supplied. No tests/build/install/services/pack/publication or source edits occurred. Only guarded external copy/write, Python AST parsing and byte/text checks were performed; prepared scripts were not imported or executed.
