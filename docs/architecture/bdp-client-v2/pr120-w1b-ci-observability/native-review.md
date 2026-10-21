# Native review: W1b CI assertion diagnostics

## Verdict

**No actionable findings.** Exact source `427cfef6decbb422b393024e36218affc26007fa`, tree `9075897867d88344a8477e1f4a8e1a00a50f1228`, ordinary parent `70ac666002708202d2f93195eed90dfe8541300d`, clean at review. Reviewed the complete one-file4+/2− patch, unchanged diagnostic helper/security unit, final author report and raw retained verification/cleanup records. Author confirmed final artifact stability. No reviewer tests, builds, services, network or source edits.

This closes the narrow observability change only. The original CI counter0 failure remains unexplained; neither the earlier instrumented diagnostic nor these passing two cases establishes a flake, clears merge, or authorizes a CI retry. External council/exact-head CI/outward actions remain root gates.

## Source correctness and credential safety

The patch computes `retainedRetryDiagnostic(report,{scenario:id,action:actionId})` only after the existing `run` returns and passes that string to all three existing `expect` calls. The counter still must equal1, the selected row still must fail, and the complete failed-assertion list still must equal the literal expected IDs. Input/control, matcher arguments/expected values, async flow and30000ms timeout are unchanged. There are no added HTTP actions or hidden setup calls.

The existing helper projects expected scenario/action, observed counts/IDs, row state/category, exchange status/failure IDs/error categories/writeState, action failure IDs/category, last client route/status and cleanup flags. It excludes raw bodies, authorization/header values, arbitrary error messages and child output. Known fixture credentials are redacted even from route strings. The unchanged unit independently checks exact projection, omitted secret material, both fixture token redactions, missing-row behavior, harness-versus-transport fields and malformed optional last-observation shapes. I verified the complete helper-plus-security-unit byte slice is identical to70ac (SHA256 `9ce1f55a182e52a3ea41efae1994a34863d7e35546000638c20560212ca5864d`). Reusing it here does not introduce an arbitrary error serializer or weaken any oracle.

Git object/mode comparison independently confirms all662 other tracked files unchanged. The one changed test SHA256 is `25559baca1589865de4cde257de028b80385a502da543e754de19b899b7c2941`; diff SHA256 `cee92317f11bfdf242255251826a2404dc2614f1966739685aca0a031c6fc0fa`. Production/package/configuration/fixtures/manifests are unchanged. The commit includes Agent-Signature attribution.

## Evidence reviewed

The JSON reporter confirms exactly two passed cases and200 selection skips, no failed case: W1b link-metadata709.812ms and existing diagnostic security unit0.781ms. The selected command's anchored two-title regex and maxWorkers1 match the preflight. Actual supervisor interval18:53:52.900Z–18:53:55.216Z,2.316s; Vitest1.22s. Six command exit records show offline frozen setup, build, typecheck, touched-file format/lint and selected tests completed0 with no timeout. Lint's226warnings/8infos are disclosed rather than called a clean warning-free result. Diff whitespace check is separately recorded.

Tests ran on the same reviewed source bytes before the commit; this is accurately stated in source.json. The newly built local emitted entry is identified by a single-file hash, explicitly not a packed/full-payload identity. No fresh installed-distribution claim is made.

The author did not add per-request instrumentation: expected36/max63 HTTP remain static workload figures, not claimed actual retained counts. This correctly keeps the separate measured36/two diagnostic under `pr120-w1b-diagnostic` distinct. Passing W1b matchers establish the counter/application and expected failed metadata oracle, not an independent full HTTP transcript.

External100ms process sampling observed server PIDs4967/4971 under group4904 and one temporary root. Cleanup.json records all27 sampled PIDs absent, no remaining owned groups and the observed temporary root absent. The passing unchanged close assertions additionally require child exit0 orSIGTERM. Sampling can miss short-lived processes; the report explicitly states that limit and does not invent detailed child exit records or fresh wire counts. No reviewer re-execution was needed.

## Stable reviewed artifacts

- report.md: `5090144b45ee8638dc034cdf82870df791a8920b6f5c52b5d33676df9de7efa8`
- source.json: `32dfaec927dbdb2b95ced67115e5468aaec38dcada20c1189d6c59d98f8d0d41`
- summary.json: `56a157636bda51dd945a15cfb969bd6c7a89d09a3a504cf391b08451f4d41fe1`
- cleanup.json: `fecd9d686ebca4a11e0b75e225a2f8e143349711486c5c3de759ef95c55fc6d8`
- selected.json: `a998dcdd5d06e8e191362c0fc85745b17112d2c3e879c53bcd1451d9d37dcaf1`

The intended result is more useful diagnostic context if the same CI assertion fails again. It is not a proposed fix to an identified serving defect.
