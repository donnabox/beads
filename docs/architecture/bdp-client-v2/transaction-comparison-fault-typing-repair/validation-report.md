# Maintenance-fault comparison guard: bounded validation report

Completed exact source validation and one counterfactual. Clean frozen HEAD `fc47fbfb5f1055766e2f680df40a1d729d39638d`, tree `11641e37dc87560c4e92928fed86c88b2b314412`, ordinary parent `427cfef6decbb422b393024e36218affc26007fa`. Commit includes the repository-style Agent-Signature trailer; Git reports N (not a cryptographic signature). Nothing pushed. Parent integration/CI and final council remain root-owned; no merged-main claim.

## Results

- Offline frozen setup passed in the positive repair checkout. Typecheck initially failed TS18048 after the earlier frozen draft TS2352 failure; both attempts remain preserved. Root authorized the stable local owner-reference correction. Corrected typecheck then passed, followed by a fresh build, format, lint, dependency boundaries and unchanged Read evidence verification.
- One positive invocation selected exactly14 cases:14 passed,363 skipped,zero failures across6 files. Full per-case JSON, nested reporter output and verbose log are retained. The run took 3.177s under the120s command limit; no test/hook timeout changed.
- New emitted HTTP case: seed200 under selective interval/fetch timers; one real callback clock failure; timer cancelled; fresh request returned500 with required empty body/header/no-detail assertions. Finally logged requests2, status500, faultCalls1, intervalCancelled/closed/inventoryChecked=true. The actual seed200 receipt was the sole inventory row.
- One isolated derivative restored only the old comparisonEpoch guard, received its own offline frozen setup and fresh build, and ran only the new HTTP case. Exactly1 expected failure/70 skipped, actual503 versus required500 at named assertion `maintenance failure is internal500`. The JSON nested reporter exposes the actual AssertionError with expected500/actual503; there are no extra setup or cleanup failures. Control took 1.811s. Its finally logged requests2, faultCalls1, cancellation/close/inventory=true. This is the planned sensitivity result, not an unexpected failure or retry.
- The service regression logged drained=true, closed=true, pendingCallbacks0. Every supervised process group is absent. Both test temp roots contain no bdp store directories; retained node-compile caches are separate artifacts.

## Attribution and scope

`final-source-binding.json` and `final-source/` bind exact source and emitted files; `final.patch` is the clean commit change. `stable-owner-draft.patch` binds tested pre-documentation source, and `post-validation-docs.patch` separately records the factual outcome-only edit made after tests. No runtime or test code changed after the positive/control evidence. The original external-Claude checkout remains frozen.

Completed selected fixture paths account for15 directories,16 owner opens,2 readonly inventory handles,7 HTTP requests and4 listener lifetimes, zero installed children. Sampling verifies cleanup but cannot independently count every short-lived directory; the aggregate ledger derives from exact selected source paths and completed cases, while the new HTTP request/fault/cleanup counters are directly observed. No full suite was run.

Earlier reports remain source-specific. No fresh consumer packaging, healthy installed14/1, default-installed fault injection, profile qualification, release or Preview2 change occurred. Fresh healthy installed14/1 still needs separate release and must be described as ordinary regression, not installed maintenance-fault sensitivity.

## Evidence

- `validation-summary.json`: all command arguments, cwd, timestamps, elapsed/exit/process groups, exact positive titles and control AssertionError.
- `positive.results.json`, `positive.cases.jsonl`, `positive.log`; corresponding control files retain full results and cleanup lines.
- `cleanup.json`: final process and store-directory absence plus exact runtime cleanup records.
- `source-binding.json`, `stable-owner-source-binding.json`, `final-source-binding.json`: review, typing correction and final clean-source boundaries.
- `oldguard-control-binding.json`, `oldguard-control.patch`: sole derivative identity.
- Original TS2352 report ../transaction-comparison-fault-guard/first-attempt-stop.json and TS18048 report typing-attempt-stop.json remain intact; no silent retries.
