# Formatting-only successor plan — CLOSED

Actual fresh bootstrap/install, workspace tool binding and typecheck passed on sourceIdentity377e7fc06d9fc96128dde57c2a9b870608b003db17215eff2b1c1bb79b671ee7. Format-check exited1/acceptedfalse in4.527s with exactly two diagnostics; no fixes applied. All four recorded attempts preserve unchanged source, matching raw-log hashes, no forced cleanup/errors/survivors/stores. Independent ps census found all47 sampled PIDs absent. Lint did not start. The actual Claude code review remains independently active: do not alter source until its terminal outcome and root disposition.

Format log SHA721818f2b3ebaac0360a710f91ff044fba055941055b3fecb6d46ef0efb4a332; failed exit f33bd1162650309f6e4ad857677fb569583892960d56b0c547851ee97368aac0. This is a formatting failure, not runtime evidence. Prior native semantic review is not a formatter pass.

## Only proposed edits

1. transaction-events.test.ts1896–1897: combine the final Bead HTTP `if (failures.length > 1)` and AggregateError throw on one line, exactly as printed.
2. transaction-service.test.ts2894–2973: collapse the final new Bead it.each invocation opening to one line with its unchanged test-name string and async callback; deindent every callback-body line by two spaces; close as `});` instead of the multiline `},` / `);`. This removes only an optional trailing call-argument comma. No callback, identifier, literal, expression, branch, await, assertion or statement order changes. The log truncates its final ten added lines: those suffix lines are a deterministic continuation of the displayed reindent and known call closure, explicitly not a claim that the log printed them all.

Exact old/new text is in proposed-replacements.json and exact hypothetical diff in proposed.patch. Both proposed transformations reverse to original bytes. These files are proposals only; no source snapshot was changed. The remaining677 tracked files, all production, installed historical prefix, source historical callbacks, real seeds, parser/tuple precedence,25+3 selection,20 observations,D37/E4 traffic, all bounds and lifecycle controls remain unchanged. Neither a formatter execution nor any broader auto-fix is authorized by this plan.

## Release and evidence discipline

Wait for actual Claude terminal review and root disposition; if semantic findings require further changes, do not treat this formatting-only proposal as authority for those edits. Root must release the concrete successor. Revalidate session and original679 source pins immediately before any future edit. Apply only these exact replacements, save new successor binding/full patch/correction delta, independently reverse-check both files, retain original author/validation/review packets unchanged. Re-freeze for narrow native equivalence closure and root acceptance.

Successor validation must live in a fresh directory with its own source/selection/helper/index/gate binding; no overwrite/retry of failed format or its attempt files. Preserve all original logs/start/exit/process/tool records with original source identity. All successor execution flags initially false; reviewers/root control release.

Only the two successful dependency prerequisites may be proposed for reuse: original workspace-install and bind-workspace-tools. Explicitly map the old source identity to the final successor, verify that the source delta is only these two tests, and independently rehash unchanged pnpm-lock, workspace/package/dependency/config inputs and all installed tool realpaths/bytes. Original tools record remains attributed to original source; do not simply rewrite its sourceIdentity. Gate acceptance must reference both successful original prerequisites and an explicit root-accepted reuse mapping before fresh static checks. If any dependency/tool/input mismatch exists, stop for root disposition instead of reinstalling or laundering evidence.

Run fresh typecheck, format-check and lint for successor bytes (do not reuse old typecheck as successor validation). Runtime stays closed until code-review dispositions and exact source acceptance. Retain existing fresh build/bind BEFORE source collection,25 source callbacks/immediate audit, archives/consumer/actual lock acceptance,3 sequential installed callbacks with immediate audits/cleanup, unchanged120s/no-retry supervision. Commit only after all source and installed validation; preserve truthful dirty base f620 attribution and prove later clean-commit byte equivalence.

No source edits, formatter/helper/test execution, installs, builds, services or review reruns occurred while preparing this proposal. Recorded hashes are proposals and evidence pins, not pass claims for hypothetical successor bytes.
