# H1b installed harness closure — frozen

Source `ee947ae919472d937625d196e2b5e5b67aa809d1`, direct parent `dfcf1a95950fd213a9c0b98c266ee9869918da06`.
Checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-installed-h1b-closure`; branch `codex/vickie-transaction-installed-h1b-closure`; origin GitHub; clean and frozen. No outward writes. Twelve files changed, 752 insertions/127 deletions. Original source/evidence unchanged.

## Result

Both reproduced native defects are repaired: actual cooperative cohort cancellation owns bootstrap, session, raw requests, reopen and child cleanup; ordinary Resource /id captures work without acquiring receipt authority. Accepted external F1/F2/F3/F5/F6/F7 are implemented, and F4's unreachable tamper claim is removed without an injection hook. Full finding dispositions are in `dispositions.md`.

The cohort has a 90-second cooperative budget with 10-second preparation/reopen and cleanup, 5-second wire bounds, and unchanged 120-second outer guard. Real installed short-budget abort cannot publish success after cleanup; pre-abort starts no child; startup abort releases its store for actual reopen. Dynamic pending traffic shares the session abort and pending drain.

Safe private ledger/session/input errors retain their reason and classify as harness errors. Actual HTTP401 and assertion failures remain target failures. Failed dynamic exchanges retain configuration/source/current-session fields at recording time. Recorded receipt authority is route-qualified, with original bytes/schema/identity checks; ordinary bindings cannot authorize receipt reads. Explicit presence assertions supplement all three context-header equality checks.

## Validation

Node24.16.0, pnpm11.20.0, maxWorkers2; logs outside checkout. Clean build and typecheck passed. Full suite: **8,492 passed, one skipped, 152 files**, 112.10s. Focused closure: **302 passed, five files**, 17.47s. Boundaries, format, lint, installed package smoke, development smoke, seven preflight tests and historical Read evidence verification all passed. `git diff --check` passed.

Fresh protocol/server tarballs were packed from this clean commit and installed into a new directory with npm scripts disabled. Packed TX: **21 passed**, 19.23s. Packed RU regressions: **112 passed, four files**, 55.33s. Both exited0; executor completed normally. No timeout termination or lingering owned Node child/process. `cleanup.txt`, exact package/entry hashes in `packed/source.json`, and evidence hashes in `hashes.json` record the checks.

Positive installed evidence: four partial rows, 35 actual HTTP exchanges, ten dynamic receipt exchanges, five physical sessions. Canonical catalog still has250 required/five optional IDs;251 untouched rows remain not-implemented harness errors. Whole-contract flags remainfalse and `claimEligible:false`. Positive evidence is `packed/transactional-observation.json`; summary is `packed/summary.json`.

Installed payload `09768af5b56fe2db91ff53bb4254e08b4983734cc547b1e549de34385008f464` (unchanged runtime payload; this is a harness-only closure). Actual fresh logical runtime identity `795917202d184f9909277d3134223c547dd99f139edb92aebdb17e857bdf0804` includes this run's generated credential associations. Source/configuration binding is exact to this observation; no old evidence is relabeled.

## Limits and next

No Snapshot/store/service/runtime owner changes, public profile admission, schema/catalog expansion or new protocol semantics. Synthetic header/lifecycle controls are distinguished from actual installed-wire witnesses. General paging, dependency/alias/catch-up/Snapshot/changefeed/owned qualification remain outside these four rows. Legacy report-v3 exclusions and package guards are preserved. Source stays frozen for independent closure review. SSE plan revision/implementation is not started; parent will authorize after review and S1b integration.

## Earlier diagnostic runs

The new ordinary-ID probe reproduced the native failure before its repair (`id-before`). Intermediate focused runs caught a test expecting runtime capture rebinding although the manifest already rejects it, then an empty-assertion negative test fixture; corrected tests respect the existing parser boundary. A temporary typecheck found two test-only typing errors, corrected before final validation. Earlier failure logs remain available; only final green runs above support the frozen result.
