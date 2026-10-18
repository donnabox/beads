# D4 helper3 independent native mechanics review

The three requested helper3 corrections are present. **One Low failed-state conformance finding remains.** The packet stays CLOSED; this review authorizes no helper, synthetic or product execution.

This is a full static review of the other author's helper3. I authored helper2, but did not use that prior authorship as acceptance: I read the complete current seven Python sources and MJS, full 913-line selected test, actual semantic data/schema, clock/ownership/proof controls, 21 stages, 48 prospective controls, prior independent rejection and root disposition. Reviewed modules were never imported or evaluated. AST parsing, JSON reads, hashing and own pure-data diff replay were the only analysis operations.

## Finding D4-P3-N1 — Low: impossible failed observation states

[`audit.py:90`](/Users/dbox/repos/wt/vickie/bdp-integration-logs/write-profile-D4-validation-preparation-3/audit.py:90) checks `validated ⇒ received && settled`. Its unreceived branch at lines100–102 now correctly rejects a settled transport rejection with null category. The complementary source states remain insufficiently constrained:

- `settled=false, received=true, validated=false` is not rejected by the current row predicates.
- `settled=false, received=false, validated=false` can carry a nonnull error category or write state, although the source's pending row has both null.

These are source-path contradictions, not new wire semantics. The actual test initializes the row at424–440 and has its only exchange await at461. Once that await resumes, received assignments/assertions at462–485, catch486–493 and finally494–498 run synchronously without another await. Thus a report cannot observe received true before settled becomes true. A pending row observed while the await has not resumed still has null category/write state. A synchronous execute refusal also passes through catch/finally and becomes a settled rejection.

| Settled / received / validated | Source observable | Existing relevant audit behavior |
| --- | --- | --- |
| 000 | Pending; response facts/category/writeState null | Allows pending, but fails to constrain category/writeState |
| 001 | Impossible | Rejected by existing validated implication |
| 010 | Impossible: no suspension before finally | May pass current row guards |
| 011 | Impossible | Rejected by existing validated implication |
| 100 | Settled transport rejection; actual nonnull allowed category and nullable captured writeState | Helper3 correctly enforces category |
| 101 | Impossible | Rejected by existing validated implication |
| 110 | Response retained, validation failed, finally settled | Allowed with actual facts and null category |
| 111 | Response validated and finally settled | Allowed only with existing wire/context/schema predicates |

The impact is limited to `CONFORMING_FAILED_PARTIAL_NO_CREDIT`: a fabricated failed record may receive that classification despite impossible source facts. This does **not** admit a passing callback or grant private/public/cumulative qualification. Other whole-report constraints still apply. No synthetic report was executed to demonstrate acceptance; this conclusion follows the full source control flow and exact current predicates.

The smallest correction is `received ⇒ settled` plus null category/writeState for an unreceived, unsettled row. Preserve the existing settled-rejection guard, genuine nullable captured writeState on a settled rejection, all nine current categories, closed schema, counters, finite source flow, raw retention and no-credit boundary. Do not reconstruct missing response facts.

[`source-state-truth-table.json`](source-state-truth-table.json) records all eight rows and source anchors. [`prospective-D03-subcases.json`](prospective-D03-subcases.json) proposes twenty finite subcases inside existing D03: eight flag tuples, nine pending-category mutations and three pending-write-state mutations. Existing D03 cases remain; no additional top-level ID, product call, execution attempt, cap or deadline is proposed. These are descriptions only, not a prepared or approved executable harness.

Serialization timing is distinct from process timing. `console.log(JSON.stringify(report))` at909 serializes the object synchronously; callback-end at910 follows that call's enqueue return. Later RPC/stdio arrival and parent receipts do not expose intermediate states inside the synchronous response/catch/finally job and do not prove flush or owner closure.

## The requested three corrections

| Prior finding | Current implementation | Static disposition |
| --- | --- | --- |
| D4-P2-N1, flat index | `gate.py:107` iterates the actual `read(idx).items()` flat SHA-string mapping; checks all520 members with unchanged exact six exclusions | Resolved |
| D4-P2-N2, late acceptance | `run.py:215` and226 both require the product flush fact and actual current time strictly before `min(flush+10, begin+term)`; ordinary branch unchanged | Resolved |
| D4-P2-N3, settled rejection | `audit.py:102` requires one existing nonnull category when unreceived and settled | Resolved; distinct complementary state issue above |

The genuine code delta adds exactly two pure `min` calls to the existing acceptance sites, changes the one gate receiver, and adds one audit predicate. No process query/signal/reap/ownership/deadline/retry sequence changes. Five helpers are byte-identical to helper2: worker, scanner, packages, MJS and current role guard. Source binding, ownership/clock policy, schema, audit expectations and stdio mechanics are unchanged. All 21 commands are identical after the exact namespace inverse; all48 IDs and other case fields remain unchanged.

## Full integration assessment

[`full-mechanics-assessment.json`](full-mechanics-assessment.json) provides fourteen concrete assessed areas. The important limits and preserved controls are:

- Registered inherited worker-group containment exists from product spawn, with actual creation acknowledgement. Product/schema processes remain phase charged. Quiet excludes exactly the authenticated worker PID, not its whole PGID. Other members, including zombies, remain evidence of membership. Parent direct-worker reaping is confined to the fresh census/no-other-member path, with a subsequent group-absence census before success. Fresh birth/PGID/live-comm identity and terminal-only comm rules are preserved; no guessed PID/PGID signaling or zombie-as-absence.
- One parent clock covers setup, source/tool/proof work, product, scanner/schema and final checks. Existing limits remain start10; end min(start+120,130); flush min(end+5,135); drain10 and strict TERM145/KILL149/outer150. The responsive loop has20ms slices; slow work is in owned processes. Both tentative and final acceptance now recheck the smaller drain/absolute deadline. Actual root outer receipt remains mandatory. No hard real-time guarantee is asserted.
- Scanner retains strict UTF8/duplicate-key/finite-number and pre-tree structure limits. Limits remain report64MiB/results16MiB/raw256MiB/marker4096, depth128/nodes1e6/container100000. Overflow is a resource refusal, never evidence that every valid product response fits these caps; unfinished hash stays unknown.
- The two five-key markers require a fresh nonnull32hex attempt and order. Ordinary null diagnostic markers cannot pass validation. Pinned Vitest console RPC plus the existing verbose/JSON pair can deliver the report after the end marker; the actual complete report must still flush within the phase deadline. No custom reporter or product body cap was introduced.
- Positive evidence independently requires all32 HTTP requests (11POST/21GET), both child generations,33receipt entries,96 deletion Events, the original99seed inventory, five oracle controls, full wire/body/header/base64/hash agreement and same-store replay/cross-row laws. GET requestBody is omitted, and auth/idempotency request headers are not fabricated as report fields. Failed source paths follow actual next links, each loop≤12 and total≤32; they are not forced into a success-table prefix. Unsupported failed boundaries retain raw/no credit.
- Real child exit0 **and** close0stdio, cached close settlement/success, pending counts and sticky ownership unknown remain required. Directory removal and snapshots cannot establish SQLite/owner closure. Marker emission or finally is not guaranteed under termination.
- Fresh build/archive/external-consumer/lock/payload/entry/proof bindings and explicit root barriers remain unresolved. Source metadata or a digest-shaped value alone is not external installed proof. The exact four-Type D4 configuration remains separate from other fixtures. Later historical Read checks do not supply Write admission, missing administrative obligations do not become N/A, and claim/qualification flags remain false.

## Frozen identity and verification

The reviewed preparation index is `8629a07c6299b0f36572c2f8217dbf9c108a9f178f60d9ca6753d864353536b3`; binding `f67df42053f97d7dadeca6240e8c4e5e8fff200682716da6ca6d1ed44d73601b`; implementation index `dec44766de20e4279d5ea4bdbd5bedced198dd5798e63dad7eb172cf70f0ab36`.

Own static checks rehashed all526 physical/525 indexed files,520 flat implementation entries/exact6 exclusions,472 origin pairs,1557 reference pins plus186 semantic pins, all679 current source files, and all15 genuine forward/inverse pairs by recomputation and pure-data replay. The mirror is exact. The full rejected helper2's296 files and nested original helper1's130 remain exact. Current source is one owned `4f4f942f…` file and678 byte-equal base files. HEAD/tree/binary Git observations are attributed to pinned root receipts; this reviewer invoked no Git command. [`static-verification.json`](static-verification.json) contains the exact counts and identities; full input copies and their origins are indexed.

Two reviewer data-inspection mistakes are retained in [`review-process-notes.json`](review-process-notes.json): an incorrect JSON key and a mistyped interpreter path. Neither invoked reviewed code or changed source/control bytes; later own inspections used the correct key/path. They are not omitted or presented as helper/runtime failures.

The result is **review complete with one Low open**, not execution acceptance. Both gates are false, approved stages and bootstrap reuse empty, and no run directory exists. Root owns disposition and any separately named correction/review/release. A concrete zero-product synthetic harness is still UNPREPARED/UNAPPROVED. Neither the48 descriptions nor this static review is synthetic or product validation.
