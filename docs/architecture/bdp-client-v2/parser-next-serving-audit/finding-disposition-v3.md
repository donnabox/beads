# Actual Claude plan findings disposition — v3

Verified against actual clean b76c3b7 (all679 predecessor tracked bytes still match). This is plan correction only; no authoring or execution release. Actual review final-review.txt is preserved unchanged. Accepted four blockers and eleven precision items with the corrections below.

| Item | Disposition and independently checked basis |
| --- | --- |
| BL1 | Accept synchronous toThrow for while-closing service and post-close direct owner. service139–154 calls check before promise creation; store396–407 enter throws; existing service test1036/1039 confirms shape. Only admitted calls use rejection observation. |
| BL2 | Accept byte-scoped parser preservation. HTTP Event branch502–536 changes; canonical gate574–591 and entire transaction-http.test.ts do not. Whole-file preservation would be impossible. Freeze the actual region hash and compare bytes after relocation. |
| BL3 | Accept conditional removal and preserve on failed/unknown/deadline cleanup. Existing installed finally356–370 unconditionally removes; launcher48–60 can reject cleanup. CORRECT the suggested hard-coded directoryAbsent:false: measure actual existsSync and report nullable actual child metadata. Preserved directory is evidence, not proof of live survivor; late exit never rescues success. |
| BL4 | Accept new GET+HEAD exact route declarations, own <=3 counter/final3 and response array. Executor560–567 rejects undeclared pairs before wire. CORRECT suggestion to reuse module-level binding: binding is callback-local; create a new installedTransactionBinding from the new child. Shared constants/helpers/validators can be reused. |
| O1 | Accept exact conditional-ordering citations5876–5878 and1986–1989. |
| O2 | Accept count timing: C3 attempts at close invocation,5 total after close settles. |
| O3 | Accept mutable clock read now:()=>tick; no tick++ generator. B and C move clock only at gate release. |
| O4 | Accept capture inside run start callback before waiter; this preserves check precedence and admitted malformed-input rejected promise shape. |
| O5 | Accept source parse/tuple capture outside entry; existence/sentinel precede comparison inside entry (store1649–1667). |
| O6 | Accept setup before SQL spy; meaningful no-scan assertions cover entered owner cases, excluded TypeError cases are separate pre-entry controls. |
| O7 | Accept corrected build timing: Event test396 import executes inside callback; installed realpath at module evaluation. Practical build-before-both stage stays. |
| O8 | Accept new transactionalProblem validator via existing Ajv/bundle. Event test122 only compiles validPage; schema4260–4283 fixes unavailable/503/after-delay. |
| O9 | Accept explicit known preserved private500 gaps for excluded Event requests. No blanket compliance claim. |
| O10 | Accept explicit unchanged D5s default/80ms waiter; E5s exchange/45s owned/60s test bounds verified. No prophylactic widening or budget lowering. |
| O11 | Accept no catalog selection/coverage claim from D/E; catalog row1186 does not turn observations into qualification. |

No other erroneous substantive finding was established. BL2's old freeze wording also reflected a then-active parser validation, now complete; v3 records actual clean commit chronology without rewriting earlier packets. The cleanup warning is applied even though v2 already intended stronger ownership discipline: v3 now specifies the actual conditional removal/reporting implementation. Six-file scope and all request/owner/probe counts remain unchanged.
