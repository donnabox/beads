# Transactional no-op conformance plan v3 — native closure

## Findings and disposition

**No remaining actionable finding in this bounded revision.** All six external findings and both retained native corrections are closed at plan level. Recommend root acceptance for bounded authoring, with the separate measured-candidate gate before installed execution preserved. This is not implementation approval, executed evidence, or profile qualification.

## Exact reviewed basis

- Plan: `transaction-next-conformance-plan-v3.md`, SHA256 `8d5b9baab8c98067729a297d6bdf10ab03b1ae9ce9717b810d4b24b7df29f495`.
- Dispositions: `transaction-next-conformance-v3-dispositions.md`, SHA256 `e52aa0e3527c1938759fd0b9ffb3b9e2d640ba599a140e232bfee0e820626b9a`.
- Completed external review: `transaction-next-conformance-claude-plan-review.txt`, SHA256 `b25abbc1f9c365bfdbb112133709325bc1e31170b825f0c06c5cbd95badc9c9f`.
- Independently inspected clean checkout: `/Users/dbox/repos/wt/vickie/vickie-current-minimum-receiving`, HEAD `c534fcab840ef2e3753ab179feae10d3ce6431e5`, tree `48537438975a4c56ee1a5bb6056013bbddd38e03`. Root reports merged main `ea584103f0dc90ae365970bb672f24bedf318e42` has this identical tree; no remote verification was performed in this review.
- Original plan hash remains `1ce661b78ee37a2c22a27239af6ea318c1942149dbd828d0b9d64de6dcce4127`; v2 remains `38d0b851848e5d0006ec972d4ea7c806c64ef857b2e6843314793e02cc602ea8`. Both were rehashed, not modified.

## Closure checks

| Review item | Source-grounded closure |
|---|---|
| External F1: actual changed-state changefeed discriminator | Action15 now reads the same genesis route after the real effectful comparator. It requires current/head pos-2, the retained seed group at pos-1 with previous pos-0, five seed Events/eventCount5, and a validated non-null continuation. `transaction-store.ts:1683–1710` serves one group and constructs the scoped absolute `changes/?after=<checkpoint>` URL only when the returned group precedes head. Thus the proposed one-group/advanced-head/non-null combination is feasible without a continuation fetch or second route. Controls1/2 explicitly stop at their failing no-op POST and no longer claim later observations. |
| External F2: measured budget rather than broad allowance | V3 separates candidate-construction stops from final measured guards. Root must accept actual M/N/C/A, per-action assertions/captures and mirror, plus declared margins before installed tests. The hard220 new-assertion ceiling remains; no240 permission is inferred. Exact counts and margins become committed guards, bounded by the unchanged fixed preflight ceilings. Fit remains unmeasured and is an explicit stop condition. |
| External F3: route placement | Current app route order at `transactional-matrix.test.ts:166–195` has beads/new before beads/b. Inserting the new row after rollback makes the new route seventh, one-based7/JS6, between those two entries. V3 preserves previous relative order and requires an independently authored eleven-route inventory and recomputed full canonical exact-configuration preimage. |
| External F4: live positive/wire-bound pins | The fourth owned file now explicitly changes pass count10→11 and extends the M/Z-only loop at `transactional-matrix.test.ts:406–421` to the new row (or an equally strict separate loop). All15 exchanges, including three receipt-client GETs, receive response bounds; all three raw POSTs receive the authored-byte bound. |
| External F5: redundant guard disclosure | Measured candidate growth plus declared small margin and fixed absolute preflight maximum have separate stated roles. Possible redundancy is disclosed. Margins must shrink if they would exceed the fixed ceiling; obsolete prior conjuncts must be visibly reconciled. No parser, transport or deadline cap increases. |
| External F6: citations and precedent | Read the corrected `transaction-store.test.ts:507–538`, `transaction-changefeed.ts:87–106`, and `transaction-batch.ts:738–771`. The former really guards only its Bead update; v3 describes the new entirely unconditional partition accurately. The seed emits five Scope Events and three Resource upserts; the later updates do not enter the linked/unlinked branch. |
| Native control correction | The third control still changes both the expected group-position tuple and observed-head expectation and requires both named failures. Its cardinality sensitivity limitation remains explicit. The separate successful action15 supplies the real advanced-state comparator. No fourth control is added. |
| Native eventCount correction | Actions3,9,15 explicitly require eventCount5 in addition to the exact five-fact array. Action15 still returns the seed group despite the Scope having seven Events after the comparator; the plan does not substitute7 into that group's count. Independent static expected-content pins are required. |

### Unused `/next` capture is an existing executable check

I inspected the actual capture path rather than treating `exists:true` as a non-null test. `runner.ts:1439–1502` evaluates assertions, rejects failed outcomes, checks exact request/wire agreement, then iterates every declared capture regardless of whether a later action references it. `captureValue` at2467–2499 requires a string absolute HTTP(S) URL and Scope confinement; null, missing, non-string and foreign URL values cannot complete this action successfully. The successful scenario result occurs only after the full action loop (`runner.ts:637–668`). A capture failure can leave successful assertion outcomes in the exchange, so the plan correctly requires overall action/scenario completion and does not invent a failed assertion ID for capture rejection.

`executable-manifest.ts:795–808,1510–1562` and runner binding validation at3090–3166 reject duplicate/shadowed or unknown references but impose no must-use rule. The exact unique unused `n-changed-next` capture therefore needs no DSL extension. `transactional-observation.ts:174–197` requires an approved POST mutation source before granting receipt authority; this GET takes the ordinary-capture branch, deleting any prior authority for that binding and returning. It creates neither a receipt/page grant nor an extra request. Static pins for the exact capture remain essential, as v3 requires.

## Independently reconciled ledgers and gates

Read-only JSON measurement confirmed the parent manifest is407227 bytes, SHA256 `7f55695401b815534f17fc69dc924ee0c0920fb6c9175529709a8e60afda5786`, with10 scenarios,151 actions and1316 assertions. Its families are101 HTTP and50 client actions; the latter include43 receipt reads and seven reopen actions as recorded in the inspected parent. The new ledger is12 raw HTTP (three POST/nine GET) plus three receipt GETs:15 actions/exchanges, no reopen. Therefore total166 actions,159 exchanges; new assertions A and total1316+A are correctly distinguished from action counts. At most19 children per cohort and four cohorts yield76 children/636 exchanges. Actual early-stop counts must be recorded rather than asserted to equal the maxima.

Fixed authoring stops remain manifest growth65536 bytes,2200 value nodes,16384 coverage bytes and220 assertions; derived absolute ceilings472763/13091/76453 are arithmetically correct. Actual measured consumption plus margins (at most8192 bytes/128 nodes/2048 coverage bytes) needs root acceptance before installed tests and must also satisfy those absolute ceilings. Nothing in this closure establishes that the fully authored row will fit. The63/90/120-second positive/cohort/outer and5/10-second exchange/lifecycle bounds remain unchanged.

The four-file scope, ten parent raw manifest/coverage object preservation, unchanged fixture/runtime/schema/catalog/DSL and false whole/public flags remain explicit. Eleven partial required rows leave239 required obligations unimplemented, with five optional rows separate; this is diagnostic coverage, not cumulative qualification.

## Work performed and limits

Only local read-only Git inspection, text reads, SHA256 checks and Python JSON/count arithmetic were performed; all completed with exit0. No tests, builds, installs, services, network requests, credentials, source edits, Git mutations or agent launches. The sole written artifact is this report. No owned background process remains. This review closes the specified planning corrections; actual candidate measurement, authoring correctness, final-source installed observations and independent implementation council remain future gates.
