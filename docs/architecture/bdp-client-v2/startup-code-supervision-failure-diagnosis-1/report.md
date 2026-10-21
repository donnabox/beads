# Startup CODE supervision failure diagnosis

The exact triggering identity predicate is UNPROVEN. The startup attempt produced a complete reviewer result, then failed supervised finalization. The final saved zombie row does not establish the cause: the same helper explicitly accepts its already-authenticated tuple, and the clean D1 CODE run ended with the same kind of accepted zombie row.

## Evidence and comparative replay

| Run | Helper SHA prefix | Raw records / successful results | Saved outer reviewer rows | Terminal outcome |
| --- | --- | --- | --- | --- |
| Startup CODE | e61b9008 (helper4) | 2519 /1 | 613 | failure642.010743s, wrapper -15, ownershipUnknown:true |
| D1 CODE | e61b9008 (byte identical) | 6613 /1 | 1724 | complete1778.759540s, wrapper0, final role verified |
| Original D1 PLAN | 44a8c194 (helper2) | 4500 /1 | 1064 | failure1133.902927s, wrapper -15, ownershipUnknown:true |

The exact startup helper SHA is e61b9008610b7349b06e8c322e240d707dbdedb9124c611890132f2cf81fb32b. The diagnosis binds52 immutable origins and copies, including actual outer-progress/exit, process record, raw stream, runner progress and historical helpers/terminal dispositions. The bounded pure-data replay inspected7651 recorded ps rows across all three runs; these are saved owned rows, not full system-wide inventories. It compared3401 outer reviewer rows with the exact recorded process identities and predicates. All saved periodic rows pass the applicable helper checks against final record bytes. The operator timezone America/Los_Angeles supplies the otherwise unrecorded lstart timezone; original environment TZ was not independently logged.

All22 startup runner reviewer samples show /opt/homebrew/bin/claude, as do all60 clean D1 CODE samples and38 original D1 PLAN samples. Runner samples omit state and lstart, so they cannot fully replay the outer identity checks. The pure-data replay was run once and exited0 with REPLAY_COMPLETE_CAUSE_UNPROVEN; replay-execution.json gives its exact invocation/output. It imports no supervisor, runner or guard as product code. Normal role guard was used before every diagnostic action. No actual review, process census, product or CLI runtime was launched.

## What is known

Startup outer-exit records only reviewer_identity:RuntimeError. The identity catch sets sticky ownershipUnknown and reviewer_identity_failure; final classification becomes process_ownership_or_cleanup_failure even when remainingOwned is empty. All42 recorded startup role checks exited0; finalRoleVerified:false means the final check was not confirmed, not that a recorded role check failed. There is no recorded supersession or timeout cause.

The supervisor sent TERM to wrapper group51688 at outer elapsed640.840757s, followed by STOP/KILL passes with no signaled groups, then terminal output at642.011088s. Deadline TERM/KILL flags and scheduled pass counters arefalse/zero, and642s is below1790/1798/1800. This was failure cleanup, not the deadline watchdog firing. The wrapper's recorded exit is -15. No PermissionError is recorded in startup; the three signal:PermissionError entries belong to the earlier D1 PLAN failure and do not establish a startup permission fault.

The preserved startup raw stream has exactly2519 valid records and one non-error success result. Result text SHA40ae1279e4747e6c245d8302df6d7febc99a9fe4dca25b01f9fd0d4e0c665502 matches independently accepted substantive evidence. That does not establish reviewer CLI exit0 or clean supervision. Runner exit.json, source-after.json, terminal-results.json and actual final-review.txt remain missing; none was reconstructed. The parent review directory's separate source-after.json is not the missing runner output. The historical audit's final-review.md basename erratum stays preserved.

All613 saved startup outer reviewer rows have the same PID51716/PGID51716/start tuple and executable basename claude. The latest aggregate row is Z/<defunct> with the same tuple. Terminal authenticatedReviewers contains exactly that tuple, so the helper4 zombie branch passes for this saved snapshot. Clean D1 CODE likewise preserves authenticated PID73969 as Z/<defunct>, with no identity error, forced cleanup or unknown ownership. Identical helper bytes and that successful case falsify the claim that helper4 always rejects an authenticated terminal zombie.

## What remains unknown

The first failing inventory, exact predicate/exception message, process-record snapshot at failure and failure timestamp were not logged. collect runs at roughly50ms cadence and during waits/cleanup, but process_sample is periodic about1s. Its later observedOwned map retains only the latest row per PID and can overwrite the triggering row during cleanup. Thus all saved normal rows plus one later zombie cannot identify an unsampled transient state, executable name, group/start tuple or changed record as the historical cause.

The generic RuntimeError is compatible with six explicit require checks in the reviewer try block (predicate-table.md), but does not identify which one; a RuntimeError from another caught operation also cannot be excluded by class alone. A transient non-zombie executable spelling, PID/PGID/start mismatch, zombie binding mismatch or record identity/argv/freshness issue are hypotheses only. None is confirmed by the saved trigger evidence. No identity guard should be relaxed on that basis.

The runner's last result progress elapsed639.814403s and outer last periodic sample639.821111s use different monotonic origins. They cannot be directly subtracted to place result drain versus first error or TERM to subsecond accuracy. The complete result's presence is established; the wrapper's exact finalization line when interrupted is not. Whether it had completed reviewer wait/reaping, residual-group inspection, source snapshot or final role guard is unknown. Empty final remainingOwned also does not erase earlier ownershipUnknown or prove a system-wide socket/descendant absence beyond the existing contract.

## Helper history and minimal next proposal

Original D1 PLAN used helper2, which lacked state and required executable basename claude even for a defunct reviewer. Its latest aggregate <defunct> row fails that older executable predicate in snapshot replay. This demonstrates an older-code incompatibility with that saved row, not the exact first historical trigger. Helper3 added state, prior-live authenticated zombie continuity and complete anchored group membership; helper4 added unanchored-known-group census/refusal. Genuine2→3 and3→4 source deltas are preserved. Those historical safeguards remain necessary and unchanged; their addition is not proof of the new failure's cause.

Recommend only the inert proposed-observability.patch and genuine inverse. It stores a fixed stage before reviewer identity operations and retains the FIRST exception's stage/class/elapsed time and already-fetched no-arguments row in the existing terminal JSON. It adds no file write or process query and never emits arbitrary exception/record/argv/prompt/credential contents. It preserves all22 require expressions, process/role/signal/poll calls and exclusive no-retry marker; source-order AST checks document this limited static evidence. Current helper and every original packet/control/guard/gate remain byte unchanged. The proposal was neither applied nor executed.

This proposal captures a future first failure rather than guessing a permissive fix. It cannot recover this missing historical trigger, guarantee a future failure reproduces, or supply diagnostics if the outer terminal output itself is absent. It changes no PID/start/ancestry/authentication/known-group rule, safe direct-child escalation, cleanup/reaping lock, sticky unknown classification, session guard, tool policy or1800/1790/1798 deadlines. No framework rewrite or new attempt is warranted by this record.

Root/A/B should independently inspect the exact proposal before any future helper authoring or D3 packet. Any later released implementation needs bounded offline stage/refusal/ownership counterfactual checks without supervisor or Claude execution. The old startup attempt is never retried and no second startup CODE review is requested. This bounded diagnosis is complete; the causal conclusion remains unknown and all future authoring/execution gates remain closed.
