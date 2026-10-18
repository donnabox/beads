# Receipt projection successor

The source-v3 predecessor remains a failed execution: T1 reached request16 and failed mutationReceipt schema validation because its whole-withheld receipt retained expiresAt; T2 passed4 requests. Actual20 HTTP, not planned21. Both fixture cleanup records show closure/removal; supervisor reports source unchanged and no remaining sampled PIDs/runtime paths. Original results and logs are untouched.

Only three files differ from final-code-v2: private authorization projection, owner assertions, and one transport expected-field deletion. For a completed receipt, nonempty all-withheld results OR compact allocations now share the whole-withheld form, deleting results, next, allocated, expiresAt. Existing schema bdp-v0.schema.json:4060–4090 prohibits expiresAt; spec bdp.md:5248–5250 requires whole withholding. Failed/nonrecord facts and empty available/expired collections remain preserved. T9 checks actual retained expiry and adds pure empty/failed projection equality, with no new owner or wire work.

T7/T10 exact expected receipts and T1 exact expected receipt omit expiresAt. T9 expired allocated expectation is intentionally corrected to whole withheld; original operation indices0/2 still derive from create/delete/create and are no longer disclosed in that representation.

Source binding: a4f2da6e6f209eb79b25f14f213dd9cf40d79a8c6777f8955a8bc042c6d1878b
Full patch: 84d8dfca808f7bb2f0083019750b83486454f79ea9cb9fe798b315925b5020c8
Correction: 89a01bd9fc67f5a3685fd8a8d231f6533e1f8879bb94ac46040e64a3613dc579
Twelve other files unchanged. Planned11 callbacks/21 HTTP/14 successful opens/11 ordinary directories+5 rejected copies/2 listeners/0 product children unchanged. C1 remains separately gated7 expected/17 max HTTP. No schema/spec/cap modifications.

Only targeted formatting was run (exit0, one file formatted); output is retained under receipt-projection-repair. No build, typecheck, tests or service invocation by this repair author. Previous production build attribution is historical and does not cover the changed projection. Fresh build/validation and native closure remain root-owned. Claude packets remain closed. This report is authored by the implementation author, not independent code approval.
