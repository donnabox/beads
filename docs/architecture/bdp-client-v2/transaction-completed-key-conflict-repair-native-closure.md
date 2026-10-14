# PR110 repair native closure

No actionable findings remain in the accepted repair scope at **84e6b67b2d322362f1d8e44e76873584d025071c**, tree **a5ded8cfb5fecf571f5f481866469071c194bab2**. This closes the original native P3 and external findings 1–5 according to the root dispositions. It is a bounded repair review, not whole-profile qualification or permission to publish/merge.

## Exact source and scope

Reviewed the entire one-file b130550 → **b3b9327c262ba16eb6b8f994303f3a08f098c7a6** delta (191 added lines), then the entire direct b3 → **84e6b67** three-file delta (+27/−5). Both checkouts were clean at inspection; Git confirms the stated direct ancestry. Final checkout: `/Users/dbox/repos/wt/vickie/vickie-completed-key-conflict-external-repair`.

Read the original external code review, original native report and updated root dispositions; inspected cumulative-test guard construction, the actual manifest and coverage, the app controls, both repair reports, frozen test results, source/mapping audits, summary, packed log records and tarball hashes. No tests, builds, services, network calls, source edits or outward actions were performed for this closure. Only this external report was written.

## Dispositions

| Finding | Closure evidence |
| --- | --- |
| Native P3 / external F1: response oracles were unprotected | `transactional-cumulative.test.ts:1063–1233` independently constructs complete assertion arrays for all twelve actions, including all 99 assertions and exact capture arrays. Expected values are authored literals, not extracted from the tested manifest. Full-record and retained-body/root equality, both 409 shapes, complete Resource/Event tuples and context presence are pinned. Removing either retained-record equality produces its named action failure. Existing request/action-ID pins remain. |
| F2: negotiation precedence overclaim | The sole coverage edit says the changed request receives 409 even when Accept admits no successful response media. The final report explicitly supersedes the historical inference and acknowledges the absent companion 406 discriminator. It does not claim a tested negotiation stage. Wire observations and partial scope remain unchanged. |
| F3: mandatory context presence | Exact arrays pin all three context headers on both inputs. The successor removes Scope-position presence separately from `c-commit`, `c-events` and `c-events-reopened`, requiring each named guard failure. Keeping existing assertion order is consistent with the accepted disposition: all assertions are evaluated and must pass. No requirement to reorder unrelated rows is introduced. |
| F4: pagination drift control | Both raw-span controls keep the pagination ID and exact byte length, change a value, and require the named pagination assertion failure. They therefore reach the SHA comparison instead of missing-key lookup or length failure. |
| F5: eligibility observation | Each actual installed control asserts `observed.result.claimEligible` is false and logs that value. The five fresh records all report false; no hardcoded substitute remains. |

## Independently checked preservation and evidence

Using read-only Git/Python inspection, recomputed all nine original manifest and nine original coverage raw spans against parent2172092: every span remains byte-identical. The final manifest equals b3 byte-for-byte. Parsed coverage differs from b3 only in the new row reason. Source-file hashes and all five summary-listed report hashes match their actual files; fresh protocol/server tarball hashes match the summary. No runtime/schema/fixture/route/canonical clause, assertion binding, whole flag, control count or deadline changed.

The manifest remains407227bytes/10891nodes, coverage60069bytes,151actions, ten partial rows and zero whole rows. New row remains12actions/99assertions. Authoring headroom2990bytes/391nodes for manifest and2998coveragebytes remains positive. Required unimplemented rows remain visible; no claims were promoted.

Inspected author executions, not independently rerun:

- Frozen cumulative29/29 pass; typecheck/build/format/lint/boundaries/Read seal/diff check exit0. Earlier b3 grammar/preparation failures are preserved in its report, not erased.
- Fresh packed positive1pass/48unselected;151actions,17sessions,18children launched and closed,3636.530ms. Ten partial rows pass; observed eligibility is false.
- Five fresh controls5pass/44unselected, exact failed assertions respectively `problem-code`, `http-status`, `ordered-state`, `retained-body`, `ordered-facts`; actual statuses409,409,200,200,200; complete request writes. Each control's row fails for the intended observation. Children close17,18,17,18,18 respectively. All records identify source84e6 and report observed eligibility false.
- Packed observation/mapping artifacts agree on the new row passing, with eleven wire exchanges and three private client actions. Fresh payload digest6945e90b33b51b1fc26cb92866399de58d36dee2da4fc09ebe126fe5ccba59ef; contracts c67d12264b97e16c76bf68f850eaf394f61896bb72491e06603ef1d6debb335b. Same payload bytes as an earlier run are not used to relabel that earlier execution: this successor has fresh packing/install/run records.

Final author report SHA256 **67882bd66756f5f8200b993f509377fa692fb4e9f4aa04368046499e9405b910**; summary SHA256 **5168921a70e9db94685a982e380013457e2cc0f2747e7432a6c7fc57b2f79b66**. Both independently rehashed. Report directory: `transaction-completed-key-conflict-external-repair/` beside this file.

## Cleanup and limits

Author cleanup records show all installed children closed, supervisors exited and no remaining processes; process-inventory file contains only a newline. This review launched no server/test child and owns no cleanup work. No broad suite was rerun. Closure covers the accepted test/coverage repair only; the row remains bounded to its configured fixed principal/view/epoch and completed-key retained-detail example. Complete profile conformance, general retention lifetime, negotiation discrimination and public serving remain unproved/unclaimed. Root owns PR110 integration and readiness.
