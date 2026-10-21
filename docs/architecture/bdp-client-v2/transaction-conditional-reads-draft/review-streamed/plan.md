# Closed observable full-scope Claude successor

Root explicitly chose a streaming full-scope successor after two conditional timeouts and the Event timeout. This supersedes the diagnostic report's optional probe/splitting recommendation; no tiny probe is scheduled. Successful concurrent smoke already shows the general CLI path worked. No provider/auth/model fix or increased deadline is proposed. Prior packets remain untouched and neither timeout supplies a verdict.

Keep identical frozen four-file source and complete109750B HEAD diff. A concise prompt supplies current code/spec ranges and permits broader necessary reads without historical review/validation paperwork. Native/actual Claude/root qualification gates remain separate.

Verified installed /opt/homebrew/bin/claude --help supports --output-format stream-json, --verbose and --include-partial-messages; exact help/argv evidence is pinned. Existing safe-mode/read-only tools/noMCP/no-session-persistence/default-model arguments remain. Root launch gate is false.900second bound, TERM895/5second graceful cleanup, no retry.

The child writes directly to a private regular-file stream, avoiding pipe backpressure. A separate reader tails it every250ms and flushes structural progress metadata per record. Process samples every5seconds add CPU time/percent, group membership, byte/record counts and last-event timestamp. Raw stream and stderr are retained at0600 for deliberate diagnostic inspection; progress summaries omit text, tool inputs/results, headers, credentials and error bodies. Prompt forbids reading credentials/auth configuration/unrelated transcripts. No broad debug logs or persistent conversation sessions.

Parser accepts exactly one terminal type=result with subtype=success, is_error=false, nonempty result and no permission denials, plus exit0, no timeout, validJSONrecords, unchanged pins and no cleanup/survivors. It separately classifies timeout, max-turns, explicit result/CLI errors, missing/multiple/unknown result, binding change and cleanup failure. Unknown schema fails closed. Assistant/token/tool partials never create final-review.txt. Only a complete terminal result is copied there; root must read/disposition it, and reviewComplete never means approval. terminal-results.json retains all actual terminal objects even on failure.

No review, tests, product imports, builds or services were executed during preparation. Python syntax was parsed without executing runner/parser. Local CLI --help is the sole CLI invocation. Root owns release and execution.
