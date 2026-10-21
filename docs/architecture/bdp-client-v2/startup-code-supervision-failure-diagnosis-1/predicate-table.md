# Exact reviewer identity checks and replay limits

All line references below are the frozen helper4 in inputs/startup/supervision/supervise.py, SHA e61b9008610b7349b06e8c322e240d707dbdedb9124c611890132f2cf81fb32b. require raises RuntimeError at22-23. The reviewer catch at143-149 preserves only its class, sets ownershipUnknown and reviewer_identity_failure, losing the condition message and failing row. The observed error is compatible with these explicit checks; a RuntimeError thrown by another operation within the same try cannot be excluded from class alone.

| Check | Frozen helper line | Required condition | Saved-data result |
| --- | --- | --- | --- |
| record PID/PGID | 119 | pid has exact int type and record.pgid equals pid | final record passes; actual first-failure record not logged |
| argv | 120 | record.argv exactly equals control.reviewerArgv | final record passes; no changed argv snapshot recorded |
| freshness | 122 | parsed record.started at least outer epoch minus1 second | final record passes |
| live PID/group/start | 126-127 | row.pgid equals pid and absolute lstart-to-record delta less than2s | every saved periodic row passes final record with operator timezone; original environment TZ not separately recorded |
| zombie continuity | 135-136 | previously live-authenticated exact tuple equals current tuple | latest startup and clean CODE aggregates pass; actual first-failure inventory unknown |
| live executable | 138 | Path(row.comm).name equals claude | all613 startup outer reviewer rows pass; all22 runner rows have matching basename, but omit state/lstart |

JSONDecodeError is explicitly ignored as a concurrent incomplete write (140-142). Other errors in decode/key/date/start/Path operations are caught in the same block, so the future proposal labels fixed stages rather than guessing a message. Current diagnostic evidence does not identify the first failing predicate or its actual row.

collect runs roughly every50ms (main loop343/358) while periodic outer process_sample rows are emitted about once per second (344-346). Additional collect calls in role-check waits and cleanup are unsampled. observedOwned is a mutable latest-per-PID aggregate (175-179), not the failure-triggering census; later cleanup can overwrite a row. The runner samples a different ps projection about every30s (run.py79-80), omitting state and lstart, and its elapsed clock begins at a different point. Those datasets cannot reconstruct the exact missing census or align subsecond result-versus-TERM chronology.

The saved arrays contain owned/group rows only, not every row of every system-wide ps inventory. Absence of a PID from one sample is not authenticated process absence, PGID reuse evidence or socket census. Unknown ownership remains sticky even if remainingOwned later becomes empty (408/433-434). Normal completion additionally requires no owned/unanchored groups, zero wrapper exit, final role check, and runner exit classification review_complete (361-372); none can be skipped because a raw reviewer result exists.
