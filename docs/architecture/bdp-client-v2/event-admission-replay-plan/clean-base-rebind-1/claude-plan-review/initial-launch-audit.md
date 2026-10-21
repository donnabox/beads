# Initial actual launch audit

Actual Claude PID/PGID 92109 started 2026-10-17T22:18:27.018424+00:00 under root session3062. The owned process group was present at this snapshot. Args match the prepared read-only/default-model invocation; one1800s bound, TERM1790 and no retry remain unchanged.

Verified all29 prepared entries using binding.closed.json for the original indexed binding. Released binding differs only in rootLaunchReleased and matches the root release. All679 current source files,19 contexts and11 artifact/tool pins match; clean HEAD9f89271e and source-before match the released inputs.

Observed 226 complete stream records, 0 parse errors and 0 error-marked tool results at this snapshot. Terminal results: 0; exit record present: False. The live stream snapshot is not completion evidence, and tool use does not by itself establish successful read coverage. No approval inferred.

Terminal audit remains pending: independently map one terminal result to final output, verify final stream and before/after/current pins with the authorized binding exception, classify actual tool success/errors and checked context paths, and establish sampled-PID/process-group absence after exit. No review retry or product/source execution is authorized.
