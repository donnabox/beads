# Independent PR132 CI audit

No actionable CI-evidence findings. Saved run38778497973 / job116572249950, attempt1, completed success on exact head `82f60a25da8799334381f6ebd810f32a3574cd2d` / base `18104a03b850553641f24fe83e7318493fa102ca`. All27 recorded job steps are completed/success; none skipped. This report does not approve landing while actual code council is pending.

Checkout log lines431–434 proves tested synthetic `6bb108b532ebda037ba63725f35b3a925048f2c6`. Saved synthetic Git commit records ordered parents `[18104a03…,82f60a25…]` and tree `3cd1da66ba3a6cbd0bbdaa2442d10002939d826b`, equal candidate tree. This is tested-tree equality, not a claim that main has merged. No remote query or mutation was performed by this audit.

Full suite lines11124–11127:178 files passed,9309 tests passed/1 skipped (9310total),753.53s. Job17:34:43–17:50:47 is964s; observed236s under the1200s job allowance, not a future guarantee. Downstream preflight lines11148–11151 is7passed/1file/1.46s; smoke lines11161–11178 records installed diagnostics/lifecycle, public Read serving/clean stop and fail-closed profile paths. Read evidence line11187 verifies unchanged89-row historical artifact with its own historical runHead/evidenceCommit; no current-run relabeling.

Actual witness summaries parsed directly from timestamp-stripped raw JSON, without masking repair:

- Positive, line2321:220 runner actions/22 sessions;23 launched/23 closed children; canonical manifesta11a79f6…; claimEligible=false. Passed once at line2323,21020ms (summary elapsed21017.282862ms). Actions are not HTTP counts; CI summary does not independently report213HTTP or the complete in-memory sequence facts. Those remain supported by the passing callback's source assertions and separate full local evidence.
- New selected successful-middle-delete negative, line11120:1 exchange,2 launched/2 closed; exactly seven expected failures `serving-position, middle-type, middle-code, middle-status, middle-retry, middle-no-outcome, middle-no-deleted`. Supplemental observedIds is onlys-sequence, with no later comparisons/receipt facts. Its enclosing sensitivity callback passed once at line11122,1170ms. This is intended mismatch evidence, not a failed CI case.
- Cross batch, line11104:204exchanges, actual409/complete write, exact13 declared assertion failures,23 launched/23 closed. Callback passed once,13551ms.
- Cross sequence, line11108:205exchanges, actual200/complete write, exactlypresentation+created-record failures,23 launched/23 closed. Callback passed once,13543ms.

All four summaries attribute source to actual tested synthetic6bb108b532ebda037ba63725f35b3a925048f2c6 and remain claimEligible=false. Reported closure totals71/71 for these cohorts. Saved CI stdout does not give independent per-PID/temp-directory snapshots for these summaries; no remote PID absence or deleted-directory proof is invented. Local installed evidence/package audits remain separate and stronger on cleanup/provenance. The CI workspace payload digest differs from the separately packed local-consumer payload, so they are not conflated.

All27 recorded steps (durations from saved job timestamps):
- 1. Set up job: success (2s)
- 2. Run actions/checkout@11d5960a326750d5838078e36cf38b85af677262: success (2s)
- 3. Run pnpm/action-setup@b906affcce14559ad1aafd4ab0e942779e9f58b1: success (1s)
- 4. Run actions/setup-node@49933ea5288caeca8642d1e84afbd3f7d6820020: success (6s)
- 5. Run actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16: success (8s)
- 6. Run test "$(node --version)" = "v24.16.0": success (0s)
- 7. Run test "$(pnpm --version)" = "11.20.0": success (1s)
- 8. Run sudo apt-get update: success (4s)
- 9. Run sudo apt-get install --yes libicu-dev: success (1s)
- 10. Run mkdir -p "$RUNNER_TEMP/bdp-tools": success (0s)
- 11. Run go install github.com/steveyegge/beads/cmd/bd@v1.0.5: success (156s)
- 12. Verify the Go-installed source-build bd identity: success (0s)
- 13. Run pnpm install --frozen-lockfile: success (2s)
- 14. Run pnpm format:check: success (1s)
- 15. Run pnpm lint: success (2s)
- 16. Run pnpm check:boundaries: success (2s)
- 17. Run pnpm typecheck: success (3s)
- 18. Run pnpm build: success (3s)
- 19. Run pnpm test: success (754s)
- 20. Run pnpm e2e:ready:preflight: success (2s)
- 21. Run pnpm smoke: success (9s)
- 22. Read-cohort evidence gate: success (0s)
- 41. Post Run actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16: success (0s)
- 42. Post Run actions/setup-node@49933ea5288caeca8642d1e84afbd3f7d6820020: success (0s)
- 43. Post Run pnpm/action-setup@b906affcce14559ad1aafd4ab0e942779e9f58b1: success (1s)
- 44. Post Run actions/checkout@11d5960a326750d5838078e36cf38b85af677262: success (0s)
- 45. Complete job: success (0s)

Read method: parsed all raw-log lines, selected actual checkout/summary/witness/downstream sections, checked fatal GitHub error/failed-suite markers, reconciled terminal run/job/target/synthetic objects. Expected negative-fixture error vocabulary is not treated as CI failure. Existing warning output (including deprecations) is not claimed absent. No product imports/tests/builds/services, reruns, network queries, source/gate edits or publication occurred.

Exact immutable evidence hashes:

- ci-target.json: `b75319b7845a9fe77a6a5b1e604f4a84c320e48c9d8f4ca63b8ccf17391c8ba1`
- ci-terminal.json: `747d370e298ef08bb7b977b34ae51ad67979e273cda7f415e5a9a7d629198657`
- ci-job-terminal.json: `abff27024a8766336cb20843181397870382d5d14cf9d825b51852bb65b41228`
- synthetic-initial.json: `934e563f71cfcba9f7088e1ac966fff09e0135e6aca345c08e6e0f661340554a`
- ci-full.log: `ad90468d26c6bb079f1a7600996df3ac6296e10720082334023c8721a328869f`

Raw log retained byte-for-byte:3383503bytes. This external report alone was written through the same-session guard.
