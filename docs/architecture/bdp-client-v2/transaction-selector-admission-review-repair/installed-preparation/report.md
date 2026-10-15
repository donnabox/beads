# Selector25 fresh packaging preparation

Status: closed preparation only. No prepared command, package, install, test, process launch, or service was executed by this preparation. Python files were parsed as text with ast.parse; JSON files were decoded. Root owns release, execution, and evidence acceptance.

Source: ordinary commit 22777eccc347b9c926de1da92409de9a924ba692, tree fe1867a19de19a824b3569a3b901269dddec221f, parent 757aedc8008386745d59d5005498a9256d6ea76e. The gate checks the clean exact commit, twelve-file footprint and bytes, final-source.json, and historical tested source-binding.json 03a861851a195d5da505f311291bee6d3cd98f63866d4b6df756e2c3ae289afd. The historical source record remains unchanged.

## Review and release

There is no execution-gate.json and no run directory. execution-gate.example.json is closed, has no approved command IDs, and leaves actual root/source/control acceptances and future consumer-lock/provenance bindings unset. Root must fill only genuine reviewed records and approve exact command IDs. All launch commands are `python3 supervise.py COMMAND_ID` under the usual session guard.

The supervisor is reused byte-for-byte through a hash-checked wrapper from command-minimum-later-preparation-v4. packages.py is that proven helper with a small retained packages-from-v4.patch: use the actual Selector static-build record shape, and rename the external consumer. No derivative/control framework is included.

Order: node-version; package-directory; pack-protocol and pack-server; verify-tar; prepare-consumer; consumer-install; verify-installed; installed-selector; installed-selector-review. Dependencies require earlier clean exits. Each command has a separate 120-second group supervisor; an existing attempt marker prevents automatic repetition. Root should release packaging first, inspect the actual offline consumer lock and bind its SHA before verify-installed, then inspect/bind installed-provenance.json before the single installed test. Failures remain retained for disposition, not overwritten.

## Build and package provenance

No fresh build is proposed. Root's existing static release identifies the tested draft; static-results.json retains the actual successful build row and log hash, and static-exit.json records all seven successful static commands. The gate reads that actual schema directly, rather than inventing old-style start/exit records. build-output-inventory.json is a read-only snapshot of current dist/schema/package/LICENSE files; it needs root acceptance linking that inventory to the successful build and unchanged source bytes. Every later command rechecks it.

Protocol and server packs use the existing pinned Node24.16.0/pnpm11.20.0 with --filter and explicit destination. Neither package has lifecycle scripts. Archive inventory must exactly equal source dist plus protocol schemas, package metadata and LICENSE; only the known server workspace dependency publication normalization is allowed. Consumer install is offline, production-only, ignores scripts, and uses local tar dependencies plus protocol-local and fast-uri3.1.5 overrides. This adds explicit --prod/--no-frozen-lockfile to the existing fresh-consumer proposal; a fresh lock is inspected and pinned before validation.

Installed member inventory and every byte must equal the verified archives/source. The fixed eight-package versions and actual dependency resolution are checked, confined to the fresh consumer. Full recursive path-sensitive runtime payload hashing follows the existing helper and is compared with the child's actual served binding. No previous archive hashes, payload, entry hash, or fixed archive member count are reused as new evidence.

## Single installed workload

One exact anchored title: “installed Selector grammar and fresh capacity preserve keys receipts and complete effects”. One Vitest case, one installed child/listener/store, exactly 25 expected and maximum HTTP exchanges: 19 POST and 6 GET; seven fixed method/URL grants plus observed receipt and changes continuation grants, nine total. Body total35,865 bytes, maximum20,054; request limit65,536/header16,384; response limit1,048,576 per exchange, 26,214,400 aggregate body envelope. Existing request timeout5s, cohort90s, test/supervisor120s remain unchanged.

The existing test retains all request bodies, response bodies/headers, raw wire bytes/digests, source/entry attribution, configuration/payload binding, and child cleanup in its installed-selector-admission console JSON. No new output hook, token logging, route, or runtime seam. The postprocessor checks the selected pass, genuine full report, all25 ledger rows/body hashes/status/context, observed routes, replay/receipt equality, finite two-group/two-Event consistency, payload/configuration and cleanup. Full schema/Selector/graph assertions remain in the frozen test; this candidate postprocessing does not replace root review or independent code council.

## Lifecycle and limits

BDP_* and Node injection/debug variables are cleared; only actual installed entry and exact commit attribution are supplied. Each command gets its own TMPDIR. Existing supervisor records actual argv, progress, exits, sampled process group and temporary entries, then TERM/KILL only its group if required. Only bdp-selector-installed- roots are classified as owned runtime leftovers; caches and unknown entries are retained separately. The test itself closes its one child, removes its runtime root and checks PID absence. Postprocessing independently checks the reported PID is absent; root still inspects process/temp inventories because short-lived/detached processes can evade sampling. Consumer and tar directories remain evidence.

All source attribution is supervisor/build/archive verification, never child-attested Git identity. New payload/configuration evidence is private diagnostic evidence only; no public Transactional qualification, original-source rerun, or deployment-wide claim follows.

Preparation correction: v1 is preserved. Root/source/control acceptance records are validated as exact document bytes, permitting Markdown or JSON; JSON parsing remains required only for structured evidence. No execution or workload change.
