# Native closure — receipt-root negative-control correction

No actionable findings in the proposed corrected negative fixture. The narrow correction is suitable for root's separate bounded execution release; this is not evidence that its build or expected failure has run successfully.

I read the actual one-hunk patch, before/after snapshots, report and summary, the service return type and receipt method, and the selected receipt-minimum test. Recomputed unified diff exactly equals the proposed patch. The before snapshot equals the still-frozen production file. The only change beyond the previously approved direct service.receipt dispatch is removal of the impossible unsupported branch and its adjacent comment. Refused authentication/foreign-identity behavior, receiptResponse, receipt pages, minimum parsing and all other routing remain byte-identical.

TransactionServiceReceiptResult at transaction-service.ts:50 contains receipt and refused only. The direct receipt method at line570 returns that union; receiptAtMinimum at line470 additionally permits unsupported. Consequently after handling refused, receiptResponse receives the receipt alternative. Removing the unsupported check resolves the specific TS2367 defect without casts or broadening the mutant. This is a static feasibility conclusion, not a compiler pass claim.

The original build failure is preserved: bypass-receipt-build.log reports TS2367 at transaction-http.ts:538. No runtime sensitivity result may be credited to that attempt. The prior review's type-feasibility assessment missed this impossible comparison; this correction explicitly supersedes it.

The unchanged selected test is `keeps receipt minimum enforcement after commands ignore copied controls`. Its first POST creates a receipt; the following future-minimum receipt GET should return the actual receipt with 200 under this mutant, failing the named `receipt minimum remains enforced` status oracle expecting503. Exact intended traffic remains two HTTP exchanges with the existing six-request ceiling, unchanged cleanup and deadlines. Only that named failure with full actual receipt/context adjudication counts; a build failure or generic exit1 does not.

Bindings:

- Product diff: `891a3dd1d17d9f754dc81ef6223bdb61de4c330e15bc5ab7760559ca5c563106`.
- Before/current production: `1339f32e0818aee5a44fc01a2034366a73e72c40027d6cde7041f26525aba366`.
- Corrected negative fixture: `9da034938dc1a48dd48b961a0ec9fa518dc954d1bddcb2ab38f321a3da13a131`.
- Corrected patch: `dca316bf79495061823fb12ed9ffd0c369e51ac2ca32858451c7301b8a691fff`.
- Report: `51d254dc1dc1ea3268b1ac74059e745d26e0ca6ac0e48fd4693680b0ed518b89`.
- Summary: `6b64eb0bbf2857a7a74369f6119d6122e5f9d37ef2e8fe9c9532fba8098ccd42`.

No source edit, product import, build, test, installation or service execution by this reviewer. The proposed mutant remains separate from production; root owns applying it to an isolated derivative and interpreting actual execution.
