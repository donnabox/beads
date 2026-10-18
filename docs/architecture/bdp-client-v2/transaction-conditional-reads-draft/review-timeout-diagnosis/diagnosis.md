# Read-only diagnosis: actual Claude reviews without a verdict

## Conclusion

The immediate reason for the three missing verdicts is known: each still-running CLI was terminated by its own configured TERM threshold at approximately895 seconds and returned143 around896 seconds. The reason the CLI had not completed is **not established**. Neither empty stdout nor one sampled PID proves it was idle, hung, denied tools, blocked on authentication, or doing useful review. No review result exists to disposition as approval.

The available evidence does identify an observability defect in the review harness: it retains only final plain-text output, PID membership and completion metadata. It captures no intermediate API/tool progress, last successful action, read volume, per-turn duration or CPU samples. Therefore the current packets cannot discriminate slow reasoning/reading from a provider/network stall. A third identical attempt or a larger deadline would repeat this ambiguity.

## Observed comparison

- Original conditional rebind: started18:55:14Z; prompt108431B;4source/22context/84artifact pins; timeout896.629s; stdout0B/stderr0B.
- Reduced-context conditional: started19:17:56Z; prompt7676B;4source/14context/11artifact pins; timeout896.344s; stdout0B/stderr0B.
- Event CODE review: started19:20:40.220Z; prompt3122B;1source/13context/4artifact pins; timeout896.673s; stdout0B/stderr0B.
- Smoke CODE review: started19:20:40.229Z (nine milliseconds after Event); prompt900B; one four-line source delta and2artifact pins; completed159.602s with4246B review output and empty stderr while the other reviews were alive.
- Smoke plan review completed149.004s with5218B output. The earlier Event plan review completed562.578s with13580B output despite its33017B prompt.
- All six used exactly the same CLI argv: safe-mode, print mode, Read/Glob/Grep tools/allowlist, disabled slash commands, strict empty MCP configuration and no-session-persistence. No model override.
- All completed/failed packets report unchanged pinned inputs, unchanged review packets and no observed surviving processes. Each sampled only its CLI PID; Read/Glob/Grep/network work may happen in that process, so this does not measure tool activity.

Pin counts and on-disk sizes describe available review context, **not the number of files actually read or tokens consumed**. The focused conditional review still asks for full review of a109750B four-file diff with2642insertions/50deletions, including matrix composition and194assertions. Reducing its prompt reduced presentation overhead without reducing that semantic workload. Event failure with a3122B prompt and earlier Event-plan success with33017B mean prompt length alone is not a sufficient explanation.

## Runner inspection

The smoke and focused-conditional runners differ only in workspace and the descriptive dirty-scope assertion. They pass the prompt as an already-open regular file on stdin and direct stdout/stderr to regular files, so there is no shell quotation, pipe backpressure, wrapper capture queue or absent stdin EOF evident in the code. The timer starts immediately before Popen. At895seconds the wrapper sends SIGTERM to its own session/process group, allows5seconds, then would SIGKILL if necessary. Observed143 means termination, not a Claude-produced error verdict. The900second naming includes cleanup allowance; it does not promise900seconds of inference.

The successful simultaneous smoke invocation argues against a universal CLI/authentication outage and against an inherent deadlock in this shared runner. It does **not** exclude request-specific auth/rate/service behavior, concurrency effects or tool/read problems; those require observations this harness lacks. No package/runtime changes or reviewed-source mutations are implicated by these logs.

## Safe diagnostic inspection

Checked metadata for ~/.claude/debug (absent), relevant ~/.claude/projects directories (no persisted conditional session trace; fresh-depth directory contained only a session folder), and the four known CLI PID session metadata paths (absent at final inspection). No per-attempt debug destination appears in argv. The existing no-session-persistence option is consistent with absent transcripts, but does not independently prove why every diagnostic is absent.

~/Library/Logs/Claude exists and contains desktop app/MCP/cowork logs. Their names identify a different surface; contents were not inspected because no evidence links them to these CLI PIDs. Credential/config/key files, authentication payloads and unrelated transcripts were not read. No secret values were emitted or copied. No further observability can be recovered from these specific packets using the bounded available evidence.

## Recommended next action for root

1. Preserve all three attempts and keep council/validation gates closed. Do not repeat either full review unchanged or silently replace Claude with native approval.
2. Prepare an explicitly observable reviewer harness before any new code-review attempt. Root should confirm supported options from the installed CLI help, then use its documented structured streaming output mode if available. Persist a bounded, sanitized progress record with timestamps, event category, tool name, completion/error status, byte/token counts when provided, and final result/exit status. Do not record credentials, request headers, auth bodies or full tool payloads in diagnostic summaries. Keep final review text separate from progress metadata. No need to enable broad debug logging or session persistence merely to get progress.
3. Use one bounded minimal read-only diagnostic target first, with one known accessible file and an objective short answer requiring a real Read. This distinguishes CLI startup/API completion from tool access; a greeting-only probe would not. Keep the normal model/auth configuration and Read-only permissions. Define success as observed read/tool completion plus final answer; define failure by a bounded explicit stage and last progress record. This is a new designed diagnostic, not a retry of either missing verdict. Do not infer approval from it.
4. If that diagnostic works, split conditional review into bounded semantic scopes on the SAME frozen source: (a) row/coverage/assertions/raw preservation/fences; (b) session routing, old observers and14/4controls; then require an explicit cross-scope integration review/disposition so splitting does not lose composition concerns. Each scope must include relevant full functions and necessary runtime/spec context, not just a tiny hunk. Event lifecycle/oracle review likewise needs a focused observable packet. Stop on a clearly observed infrastructure failure; do not manufacture a product issue or ask a product decision.
5. If progress proves productive reading/reasoning reaches the deadline, redesign scope based on the last observed stage. If it proves tool denial, repeated provider errors or no first response, address that concrete failure first. Consider a different duration only after there is measured progress justifying it; increasing budgets now would conceal the same unknown.

No next attempt has been prepared or launched by this diagnosis. Root owns exact instrumentation design, execution release and coordination. This report is evidence inspection, not a reproduced/fixed underlying Claude failure. A live red-capable diagnostic was deliberately not run because this assignment prohibits Claude execution, tests and services; no new approval question was asked.
