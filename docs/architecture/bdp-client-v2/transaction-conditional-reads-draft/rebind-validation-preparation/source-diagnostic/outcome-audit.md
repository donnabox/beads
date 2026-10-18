# Diagnostic source outcome audit

All eight released stages passed once, in order. All56 selected source cases passed; independently read actual Vitest JSON and matched complete selected names/counts against commands.json. This is frozen dirty-source diagnostic evidence with native review accepted and external review pending, **not full council acceptance, integrated-tree validation or installed qualification**.

- typecheck: PASS, 2.459s.
- format-check: PASS, 1.088s.
- lint: PASS, 0.987s.
- boundaries: PASS, 1.069s.
- cumulative-full: PASS, 2.257s, 42 exact selected cases; 0 unselected/skipped cases.
- binding-source: PASS, 0.869s, 3 exact selected cases; 0 unselected/skipped cases.
- receipt-first-page: PASS, 1.186s, 8 exact selected cases; 247 unselected/skipped cases.
- app-static: PASS, 1.186s, 3 exact selected cases; 64 unselected/skipped cases.

Static command caveats: format checked454files without fixes and reported one size warning for the3.5MiB Read+Update matrix above its configured1MiB limit. Lint checked454files without fixes and reported1020warnings/83infos (1083diagnostics suppressed by its display limit); exit0 does not mean warning-free. Typecheck's pnpm prelude reported lockfile up-to-date/already up-to-date. Boundaries confirmed11workspace packages. No fixes, retries, builds or installed phases ran.

Cleanup: every supervisor reports no timeout, forced cleanup, errors, surviving process group/PIDs or runtime directory. Independent final ps check found none of the sampled PIDs present. Four static stages retain only node-compile-cache entries under their dedicated TMPDIRs; those were preserved, not deleted or mislabeled as runtime leaks. The source-derived47synthetic attempt/58diagnostic ceiling and24listener inventory is not an independently counted HTTP completion result; deliberately canceled requests remain attempts.

Final source audit: all tracked input hashes equal input-binding.json; HEAD remains1ba65f2f192cc84e44eb09ef89b9693ec6cdaf05; exact four-file dirty status and patchbc7e620ea97a37933901ee4bf2e2d818685178bd974a262e88b025e513e2a60a unchanged. Concurrent reviewer source inputs were not modified.

Invocation detail: the last app-static supervisor was invoked via ambient python3 without the additional outer guarded-action wrapper used for the first seven. It resolved to the exact pinned /opt/homebrew/opt/python@3.14/bin/python3.14; the supervisor still ran the actual command through guarded-action.py, as its start/result guardedArgv records show. This did not bypass the diagnostic gate or the command's session check. No duplicate run was performed to conceal the invocation difference.

Full machine-readable stage exits, logs, process samples and four result JSON files remain under run/. Root can disposition these diagnostic results independently while actual Claude review continues. Optional ordinary Update hardening remains separate.
