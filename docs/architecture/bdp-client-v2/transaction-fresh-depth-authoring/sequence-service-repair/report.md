# Bounded sequence service freeze repair

Only transaction-service.ts and transaction-service.test.ts changed. All18 prior authored path hashes remain exact. No commit/push, external package rebuild or installed rerun by this author.

Regression first: one real fixture/owner/service, literal12000-level array, fresh sequenceObserved then retained same-key replay. The original recursive freeze produced the exact RangeError at transaction-service.ts646: red1failed/67skipped in0.78s. Original production was unchanged during this red run.

Production fix replaces only the method-local recursive freeze with the same explicit-worklist traversal already used by private store freezeReceiptValue. Already-frozen parents are traversed; no API/policy/admission/limit changes or shared abstraction. The test independently compares every JSON leaf/key/container, checks every observed object frozen including descendants, asserts exact created Resource shape/type/deep leaf/revision, same serving position and full replay, and identical resources/groups/identities/position/Events after replay. The full body never enters native stringify or a recursive matcher. Focused case uses0HTTP/1owner/1directory, standard awaited fixture afterEach closes service/owner and removes directory. New callback timeout20s; each external run deadline120s,TERM115s.

Green focused1pass/67skipped in1.73s; complete service file68/68pass in2.62s. Format passed. Additional typecheck initially failed TS2339 because projected entries infer {} before .resource access. That failed log remains. Narrow test typing correction binds the already length-checked first entry as Record<string,unknown>, then reads the same resource; no assertion/request/control changes. Successor typecheck and targeted format both pass. The68-runtime pass belongs to the pre-type-annotation bytes; no runtime rerun was represented as occurring after this trivial typing bridge.

All seven run snapshots report source unchanged during their own execution, no forced termination, no surviving process group and no remaining test store directories. validation-summary.json contains exact unique sampled PID set and independent final absence check. Source-only suite uses repository Vitest source aliases, not external installed qualification. JSON results and raw logs retain actual failures/successes separately. No unselected or package-wide pass is claimed.

Root owns code council, final commit, build/package and original installed13 rerun. Source bindings and full two-file correction.patch are frozen here for review.
