# Native measured-draft preflight: comparison maintenance-fault guard

## Verdict

**No actionable findings.** The stable five-file draft implements the accepted minimal guard and planned source observations without changing protocol, fixtures, runtime seams, admission policy or deadlines. This is independent static preflight closure, not an execution result or authoring/publication waiver. Root must release execution; the dependent427cfef parent remains subject to its own CI/integration gates.

Reviewed checkout `/Users/dbox/repos/wt/vickie/vickie-comparison-fault-guard`, parent `427cfef6decbb422b393024e36218affc26007fa`, tree `9075897867d88344a8477e1f4a8e1a00a50f1228`. Exact draft patch SHA256 `69c3c4fd2e60ff1ec069755efaac37920b55d51402d89f620f4eeb1faf17dd6c`;272insertions/4deletions across exactly the five accepted paths. All owned/read-only source hashes in source-binding.json were independently recomputed. Parent owned files match70ac. Author confirmed the draft/preflight stable before this closure. No reviewer tests, builds, installs, service starts, network or source edits.

## Production and source oracles

The only production change is `comparisonEpoch`'s early return: benign closed owners still return undefined; retained ownerFailure/maintenanceFailure invoke existing fault-first `enter`, which throws before attempting metadata access on the closed database. This matches the adjacent final-deadline guard and preserves observed/unobserved healthy shapes. No failure is swallowed or context manufactured.

All original deadline-observation checks remain in the three owner cases. Added epoch assertions cover benign close, live/closed fencing and actual timer-maintenance failure with original cause. New submission variants use the real owner's closure-backed fields plus typed vi.fn delegates on a spread object, avoiding mutation of frozen owner properties. They create no dependency reservation and use no cancelSequence teardown. Both admission delegates, final observation and wait must remain uncalled. Fault variants assert wrapper/cause; normal-close variants strictly distinguish legacy refused from observed comparison-refusal/owner-stopped, with no context.

The HTTP test uses the existing emitted runtime fixture with only existing clock/maintenance interval options. Selective interval timers, seed200/schema/independent Resource facts, one injected clock failure and cancelled interval establish the intended actual callback path. Real timers are restored before the second POST and in finally. The supplied clock explicitly replaces fixture time; no inert setTime call is used. `$defs/mutationReceipt` exists in the unchanged schema and its properties agree with the seed assertions.

The second response must be the existing handler's500/empty/length0/private,no-store/non-attesting/varyAccept, with no content-type, retry, location, limit, three context fields or sentinel in body/header values. This cannot pass by accepting an ordinary503 or a different fallback missing vary. The actual seed receipt is retained for exact sole-row inventory equality, while its identity/type/properties/metadata/status/position/result facts are independently asserted first.

Primary HTTP failures are collected; finally restores timers, attempts runtime close and exactly one readonly inventory, records actual counters/status/close/inventory, and aggregates rather than replacing the primary cause. Thus the old-guard control can fail at the named500 assertion and still report cleanup. Fixture teardown remains unchanged and closes/removes its registered resources. Submission fault teardown similarly preserves primary and close errors and restores timers. Successful execution and derivative adjudication must still verify the logged facts; static review does not assert that cleanup has run.

## Exact selection, control and budgets

selection.json has14 distinct parameter-expanded titles. Its suffix-anchored escaped selector matches each intended title and excludes an unrelated title; four false/true variants are counted separately. There are3owner +6submission +1service +1dependency +1alias +2HTTP cases. The existing selected cases and their fixture/reopen patterns were inspected. Any runner-selected count other than14 is a stop, not permission to run additional tests.

The old-guard derivative is specified as exactly one reversal of the five-line production guard to the original unconditional closed shortcut, built separately before selecting only the new HTTP test. Required sensitivity is actual second503 plus nested named `maintenance failure is internal500`, with requests2/faultCalls1/intervalCancelled/closed/inventoryChecked all proving the intended path. A generic AggregateError, setup failure, timer incompatibility or timeout does not satisfy it. No derivative was executed during this review.

Static workload matches the accepted v2 ledger:14positive cases use14directories/15owner opens/1readonly handle/5HTTP/3listeners; one derivative adds1directory/1owner/1readonly handle/2HTTP/1listener. Combined **15directories,16owner opens,2readonly inventory connections,7HTTP,4listener lifetimes,zero installed children**. Existing fencing's single reopen explains the extra owner/listener without an extra directory. The12 non-HTTP cases each open once. Inventory handles are separate from owner opens.

I independently verified all7 source-request bodies/lengths/hashes in requests.json. New positive/control bodies are150/164B; existing fencing POSTs120B each and its GET body empty. No dynamic Host/port is fabricated. Selective interval/fetch compatibility remains a named unproved go/no-go inside the seed request; it adds no extra probe, retry or timeout.

## Separate healthy installed proposal

The unchanged installed direct-refusal test has14 explicit bounded exchanges and one child, matching all14 request ledger entries (largest33-operation body4616B). It uses four static routes plus only the actual captured seed receipt URL, request5s/cohort90s/test120s and existing size/cleanup bounds. I read its request sequence, binding, report and finally cleanup. It remains healthy regression, not installed maintenance-fault sensitivity.

Preflight correctly requires an externally verified clean frozen source and actual fresh protocol/server tar-to-consumer provenance; the helper's declared source label and reported workspaceDirty value do not themselves prove that. It also correctly limits retained evidence to existing fingerprints/context rather than full wire bytes. The proposed combined source/control/healthy workload is21HTTP,one installed child,16directories/17owner opens/2inventory handles/5listeners; root release still governs that later packaging/execution.

## Stable reviewed artifacts and limits

- preflight.md: `56056217cf2257dbebc89f2fc2ebe5732ad171bc712ed1d727c733a2734eb4de`
- source-binding.json: `9980b6701d2a144979e7f76c6452f1483fe79cbec65535d638294a522f1e446f`
- selection.json: `4bf4631e67436f30f62bcd985931033f831be5b83a307ba9b68c3396722576f4`
- requests.json: `117449bd402a40533e995c224108c0c506b50c463bc87fb979b89a55ff5cf52a`

The progress note preserves30035/70ac residual history and labels this successor unexecuted. Source/static/Read gates, actual14pass selection, named503 control plus cleanup, fresh-source package regression if accepted, and independently absent owned processes all remain future evidence requirements. No complete profile, production failure recovery, CI cause or merge-readiness claim is established by this review.
