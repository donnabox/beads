# Native review — frozen six source-test paths

Three concrete blockers are recorded below: one already self-reported invalid Property Change operation, one invalid compiled fixture, and one missed production depth dependency with two call sites. No source, test, build, import, service or reviewer CLI was executed. Source remained frozen; this report does not authorize a repair or validation.

## N1 — compiled Type identities must be Scope-local

The new transaction-submission.test.ts1767–1807 callback compiles descriptors using the file-global https://example.test/bead and https://example.test/link, with Scope https://example.test/submission/. transaction-store.ts278–280 explicitly rejects any compiled descriptor outside `${scope}types/`. The witness therefore throws at its first open, before any deep operation.

Use callback-local Scope-local bead/link URIs for both descriptors and every request/expected value. The final after-expiry body currently invokes global create(), so it must also use the local Type explicitly. Keep the global fixtures unchanged. Recompute affected body measurements; this needs no new case, owner, HTTP or runtime limit. Root accepted this source finding.

## N2 — unsupported public copy operation (known author finding)

transaction-submission.test.ts1888 uses Property Change op copy. The accepted grammar is add/replace/remove (protocol read-update-values.ts57–59 and schema propertyChange). The evaluator's internal copying helper is not a wire operation. Preserve known-correction.md and the failed-to-be-executable frozen proposal.

Replace only that first change with add /copy and the independently constructed existing deep literal, retaining the following /deep/- append and full expected postimage. Its body remains bounded; refresh body/hash/measurement facts. This was already found by the author; it is not a newly independent discovery.

## N3 — two recursive receipt freezes still prevent the required deep serving path

transaction-store.ts894–900 observeReceipt recursively visits every Object.values child. The separate readReceiptPage freeze at2255–2260 is also recursive. Both are within the already-owned store path but unchanged by the frozen production patch.

The new source callback reads an actual issued continuation at submission1883–1885 whose result contains the depth12000 Resource. That page reaches the recursive freeze even after N1/N2 are corrected. More broadly, executeObserved calls observeReceipt (store1398/2159, submission206), so the new installed request1 deep seed reaches the same recursion in its first observed response. This is a reachable implementation dependency, not a proposal for a new depth limit or an extra witness.

Replace both local walkers with a single depth-safe iterative full-value freeze seam, preserving projection, prefix/page-plan validation, contents and nested immutability. Do not weaken the page test, reduce depth, skip frozen values as a substitute for proof, change retained detail, or broaden JSON semantics. Keep the genuine issued page and the installed full deep receipt assertions. Root was notified immediately. My prior production review missed these walkers; the prior installed report's oracle correctness remains useful, but it must not be read as closure of this production dependency. No runtime stack-overflow claim is made here; this is the same directly reachable unbounded-recursion defect being addressed by the slice.

## Remaining review

- Read the complete six-path source diff in owned.patch, the added submission and Snapshot callbacks and their fixture/caller context, preparation/sequence modifications, HTTP callbacks and response schemas, selection/measurement/known-correction records, and relevant receipt/evaluator/sequence internals.
- The obsolete depth-only expectations are replaced by successful effects or existing negotiation/conflict outcomes. Invalid Selector/carrier/key grammar, clone-brand rejection, normalized identity, frozen carrier/arrays, independent pending ownership, real atomic-admission fault cases, completed prefixes and after-expiry key behavior remain covered. The direct legacy owned claim is now executed rather than abandoned.
- HTTP changes retain receiving401/403/400 precedence and full Problems, exact deep accepted results, retained replay, no-effect refusals, current context and final Resource checks. The two-member capacity test legitimately continues after a conflicting first member; its second independent member advances the position, while direct maximumOperations1 remains413. The source direct13 deliberately retains fresh deep unacceptable Accept406 and then submits the same deep value with acceptable negotiation; this differs honestly from the installed direct13 accepted fixture change.
- Deep source assertions use explicit worklists and independently constructed raw literal expected values; shallow matcher inputs are keys, tags and scalar metadata. The supported-input correction must retain that property. Snapshot positive uses BOTH emitted owner and emitted preparation, preserving brands; its actual control therefore needs a fresh bound emitted build, copied source/config/setup closure and a sole comparator mutation, not a stale dist or generic failure.
- Snapshot negative corrupts one leaf in the real owned inline Link after a genuine committed Link, leaving its first-class representation/receipt intact. It asserts the specific integrity error and zero issued Snapshot rows. Snapshot positive checks manifest schema, full deep first-class and owned Link equality, ordinary handle and receipt after reopen. It makes no S1 minimum claim.
- Source receipt recovery preserves actual issued continuation and deep available receipts across maintain/reopen, then completed allocation compaction, failed forgetting and retained key. Those existing assertions should run after the fixture/recursive-freeze fixes; no fabricated receipt/page is needed.
- The three new callbacks register bounded local error collection and close/remove cleanup around owner use, including offline SQLite handles in finally. Their20s work/5s cleanup checks are elapsed assertions, not preemptive interruption of synchronous code; the30s test and120s outer process supervision remain necessary. No extra product child is introduced.
- The parsed selection has42 expanded cases. Its six HTTP callbacks are4+11+6+13+5+8=47 planned exchanges. Summed declared successful lifecycle inventory is53 owner opens,37 directories,9 listener lifetimes,15 offline handles and0 product children; new callbacks contribute6 opens/3 directories/2 offline handles/0HTTP. These are source-derived planned accounting, not observed success. Installed34HTTP/3children and the separate0HTTP comparator control remain separate. No selected runtime was run here.
- Identity numeric output is an execution-time observation still pending. Do not report its future values, runtime outcome or full profile eligibility from this review.

## Binding and minimal next step

HEAD stayed77b4cf980ec91f66c52a554c3dd8c9de5aee859f. The closing check matched all six frozen hashes below. owned.patch SHA2ea02de69f576ecdacd46e944eec67d64b371a95d6a32e3ee5db613a9927173e; source-tests-binding.json dc8c54893cbfb0eb8b75e3ed6a57583981250950990080880a69e3058a713877; selection.json21146a1af68e0015728c78966bc4f1c9a398140093ce4c7f7ed40b92e8db2fc4.

| Source test | SHA256 |
|---|---|
| transaction-set.test.ts | aec0ce27196f6ec4916cb2f35b8a3de7ee879ab35f14aab657e11416254860b8 |
| transaction-submission.test.ts | 40541de34a881acdfefde923cf02a8e0d72687488fd079fb79ca395987d5ae29 |
| transaction-sequence.test.ts | 5526903a6e094602612bac8448fe1bee5e8f8c6e675458360abc43e43eda6d91 |
| transaction-http.test.ts | 6b6ea02c8a6fa7bf2c59c6abd24e190a577a43858260443ad8e09b413b09cbb5 |
| transaction-snapshot.test.ts | 01c8e65a6dfe19bd13c41088446ac47c2dd2b06c9915b63480528c4495056cb4 |
| resource-evaluator.test.ts | b7f557f57e6980a6c514afe6b76b0b6b69938ecc4265c02246994ab7ed8e5510 |

Preserve the active actual Claude inputs until terminal. A successor can combine the two submission-fixture corrections and store freeze correction, plus the separately reported installed startup-retention repair. Rebind exact source/body measurements and preserve the original frozen packet. The already-planned source and installed witnesses exercise these corrections; no count or capacity expansion is required. Actual execution and final review disposition remain root-owned.

## Root-accepted bounded repair design

Root confirmed both N3 sites and accepted reopening the earlier eight-production council while preserving its reports. The proposed implementation stays in transaction-store.ts: one private, unexported module-level freezeReceiptValue(value):void with an explicit pending stack, a for-of over Object.values and Object.freeze for every encountered object. It must descend through already-shallow-frozen parents; no Object.isFrozen early skip. Domain remains parsed dense acyclic JSON. Both observeReceipt(projected) and readReceiptPage(page) call it at their existing boundaries, preserving observation/projection and transaction timing.

Within the existing submission callback, add an independent stack walk asserting Object.isFrozen for every object in actual returned root receipt, issued page and reopened page. Preserve complete-value assertions, actual URLs and the existing workload; no added HTTP, owner, directory, case or limit. This is a proposed successor, not applied here. Final native and focused actual Claude correction review must cover the store helper/call sites, repaired fixture/body and immutability assertions after the active immutable reviews finish.
