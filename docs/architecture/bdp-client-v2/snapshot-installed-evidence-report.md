# Installed Snapshot evidence — frozen author report

Source: `abd474c6a178e37234d209228ec2d42217eace2e`, direct parent `2124adb31ff3406052231fa67502d83793bb649d`.
Checkout: `/Users/dbox/repos/wt/vickie/vickie-snapshot-installed-evidence`.
Branch: `codex/vickie-snapshot-installed-evidence`. Clean, local only; source immutable for review.

## Scope and unchanged boundaries

Only two new files: `apps/bdptest/src/transactional-snapshot-installed.test.ts` (438 lines) and `docs/design/transaction-snapshot-installed-observations.md` (62 lines).
No changes to runner, launcher, execution binding, shared matrix, runtime/server, package surface, schema, catalog, coverage flags or sealed evidence. No helper widening was needed.
The existing source-only launcher drives the emitted or actual installed private process. `installedTransactionBinding` validates metadata and physical dial shape; the existing raw HTTP executor serializes canonical Host and credentials on each exact route. These are independent source-private observations, not matrix row credit or public admission.

## Observations

Two tests use actual command HTTP requests and completed framed responses, default compiler/artifact configuration/default clock, canonical `https://reference.invalid/acme/` Host and loopback ephemeral physical port. Credentials remain transient. The JSON summary omits request headers and bodies.

1. Seed 34 Beads and 34 owned Links in four real batches. Both streams split at32 and have literal continuations. Check expected complete ID lists and owned-Link postimages; inspect Type descriptors against configured JSON (normalized member order is permitted). Capture nonzero checkpoint, exact retained manifest/pages, UTF-8 HEAD length, real wrong-stream token404. Mutate a Bead, restart actual child, assert same configuration binding and byte-identical retained bodies under a separate authorized reader/current context. Read finite changefeed from literal captured checkpoint, apply real changes, compare all68 independently enumerated current point reads. Omitting the real group or one expected Link fails the same oracle.
2. Fill seven handles; HEAD/304/412/406/auth/method/Origin/malformed-query/Q1 unsupported branches leave enough capacity for exactly one more capture. Ninth is503 with retryAfter; all eight retained bodies still work. Actual restart retains capacity refusal and prior manifest. This proves a bounded capacity inference, not private ledger inspection.

Completed fresh packed observations: first test86 exchanges (85x200,1x404); second29 (19x200; 304/400/401/403/405/406/412/500 once each;503 twice). Total115 completed raw requests/responses. No injected successful results, database inspection, test clock or process-only activity substituted for wire completion.

## Validation

Pinned Node24.16.0 and pnpm11.20.0.

- Frozen committed rebuild passed (`committed-build.log`).
- Frozen typecheck and format passed; dependency boundaries passed, lint exit0 (133 preexisting/worktree warnings,8infos; new test targeted lint clean). Formatting retains existing large-file warning.
- Focused final emitted validation17/17 across3files,0skip,2.11s: new2, binding3, SnapshotHTTP12 (`final-focused-green.{log,json}`).
- Actual new protocol/server tarballs built and installed in isolated external log directory. Committed exactsource test2/2,0skip,1.44s (`packed-snapshot.{log,json}`, `pack-status.json`, reproducible `pack.py`). No fallback to workspace process: entry points inside installed/node_modules/@bdp/server/dist.
- Historical Read verifier89 rows passed (`evidence.log`); original e5f7/dbaf observations preserved. No fresh Read run or relabeling.
- git diff check clean; no source changes after freeze; no owned test processes remain. Each service closed/reaped through existing launcher; fixture timers cleared and disposable store directories removed. Installed package/log directory intentionally retained for reproduction.
- Full suite and broad smokes not duplicated for this test/doc-only successor; root2124 integration had those checks. This report does not transfer their exact-head status to abd474.

## Preserved failed development observations

`focused.log`: first test passed, second failed because authored header was mistakenly `bdp-min-scope-position` and correctly did not activate the supported control name. Corrected to `bdp-minimum-scope-position`; original response200 retained in log.
`final-focused.log`: unsupported assertion that Type-hosted JSON bytes preserve input member order; changed to structural content comparison.
`final-focused-corrected.log`: attempted extra JSON Schema retrieval returned404. Removed this out-of-scope hosting assertion instead of changing runtime; no JSON Schema hosting qualification claimed. Type descriptor hosting and actual compiled metadata remain distinct observations.
Final17 and freshpacked2 pass after these test-only corrections. No deadlines, runtime checks or protocol semantics weakened.

## Explicit limits

No complete Snapshot/Transactional conformance, whole-row flag or public profile admission; claimEligible=false throughout. Q1 supplied consistency controls remain unsupported (500 development marker). No five-minute expiry sleep, clocked behavior, crash/fault recovery, corrupted store, ceiling-sized response or issued-ledger/heap qualification. Actual raw HTTP shutdown does not prove socket-disconnect AbortSignal propagation. All-visible fixed view only. Client response limit1MiB is local to this small fixture, not a server maximum claim. Descriptor structural equivalence is not exact original artifact byte hosting. Source commit label is supplied by the test launcher; actual loaded payload is separately bound by its measured digest.

## Packed identities

Complete per-test bindings, including runtime/authentication/policy differences from fresh local credentials: `packed-bindings.json`.

Test1: payload `40fdc98fe78dfc8a8d9916c41f5b02457e19ceea9b1d919cb4ac914f203ad520`; contracts `1d788f65b366fcac24fbbc90dd76a744883bd484130526de128416946d100564`; runtime `1038b80d47f1b0ccd3f57646639e26ea9f1f34559cf1582239fa37616567c8ee`.

Test2: payload `40fdc98fe78dfc8a8d9916c41f5b02457e19ceea9b1d919cb4ac914f203ad520`; contracts `1d788f65b366fcac24fbbc90dd76a744883bd484130526de128416946d100564`; runtime `3110c2fe126216ae16257dbc2a4f5248fb6a93cb34dfc118c40bf3d45f506190`.

protocol.tgz SHA256 `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`.

server.tgz SHA256 `432fda4fcad281a9bab225b93e14bb1e2c7aeb78d42053c905fa660cea962bad`.

## Root native follow-ups recorded after freeze

Root identified two bounded successor repairs; do not treat this author report as council acceptance. First, `withInstalled` cleanup failure can mask the original observation failure, and no final AbortSignal assertion prevents a late success summary. Preserve both causes and check the deadline before successful completion in a reviewed successor, without changing the shared launcher. Second, identity lists are independently seeded but owned-Link postimage equality compares two Snapshot views, followed by comparison with current point reads. The frozen tests do not independently assert every seeded type/property/source/target/metadata field; a consistent cross-surface wrong postimage could pass. The committed prose's independent-postimage implication should be narrowed or backed by explicit seeded facts. Root will obtain council on immutable abd474; these observations do not invalidate the recorded runs but limit their oracle strength.
