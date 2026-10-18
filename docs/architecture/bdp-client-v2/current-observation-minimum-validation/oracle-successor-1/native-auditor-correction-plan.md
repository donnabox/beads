# Native actual auditor failure and bounded correction plan

The product source selection passed97/0, with348 unselected/skipped tests, but its immediate auditor failed raw1. This is a concrete auditor schema defect, not a product failure or a passing completed validation. No packaging/external stage ran.

At audit.py:120, scalar checking demands numeric refusedOwnerEntries=0. The historical fixture at transaction-http.test.ts:5598–5601 explicitly asserts entries equals[], then serializes refusedOwnerEntries:entries at5674. All three actual parser-refusal-source records contain an empty list, with calls37/13/17. setupCalls1, refusedWaits0, refusedTokenAllocations0, owners1 and httpRequests0 all match. The erroneous numeric expectation was already in the original validation implementation; my earlier preparation review missed the shape mismatch. Preserve that history rather than modifying the source/report to fit the helper.

Proposed sole behavioral helper correction in a separate CLOSED validation/helper-successor-1 packet:

```python
for o in by['parser-refusal-source']:
 scalars(o,{'setupCalls':1,'refusedWaits':0,'refusedTokenAllocations':0,'owners':1,'httpRequests':0})
 rows(o,'refusedOwnerEntries',0)
```

The existing rows helper requires an actual list and exact length0. It rejects a numeric0, missing field, non-list or nonempty owner-entry list. Retain every scalar, all three required report instances and all other predicates. Do not coerce the field or relax to truthiness. Product source, selection97, schemas/workload, response assertions and deadlines remain unchanged.

Data-only inspection of the remaining source predicates after the failing line found no additional mismatch in this actual run: capture rows pos-01/unknown and missing rows absent/dangling/deleted/hidden match; retained baseline equals after; capture arrays have5 contexts/5 attempts/4 unproved and required counters; missing rows have3 entries/3 mappings and writes2 or3 as specified; incident has2 perform calls/3 entries/0 alias lookups/1 wait/4 clocks/2 writes. Snapshot status histogram is200:5,503:8,500:4,400:4,401:2,409:2,410:2. Receipt status sequence is200,503,503,503,410,409,400,404,404,401. These observations diagnose readiness only; no corrected helper was executed and they do not convert the failed audit into a pass.

Actual prefix evidence: typecheck4.944262s, format4.221617s, lint4.698594s, build5.227806s, binding8.924604s and focused-source11.615579s passed raw0; review-focused-source9.489538s failed raw1. All seven source/log bindings checked, all679 source files match normalized f133bf6d/current identity2ce269a0, no forced cleanup/errors/remaining processes/directories reported. Fresh ps census found all107 sampled prefix PIDs absent. Lint warnings remain165 plus9 infos. Original bootstrap2 remains separately attributed.

Actual E now has37 responses,10 waits/52 clocks, failureCount0, closed/directoryAbsent true and its owned directory is absent. Its methodsPOST3/GET21/HEAD13 and statuses200:7/503:10/410:8/409:8/404:2/400:1/401:1 match the planned matrix. Nine new and28 legacy records are present. This is distinct from the original failed E11 run and still not installed evidence.

Evidence pins:
- focused-source exit: 023055e638b11e21789a1c5793a1f3ec9d6990778be3743a9a0943f9aa7e52ea
- focused-source results: 801cb0d37c101234c3f740b7da8d10915befc0de525441513c5fed4f84623b8b
- focused-source log: 0d72d35607b64e8f4dc96144a513521c65ea44c99d1af479aa7a77e935b8b32d
- failed review exit: d8e0b84c8933f1ecf95aac0f35d5ad8bcce5843df756b955e16170156e66845e
- failed review log: 8a076bfad9a580c6d2187d2516213e73aa073d8747eee03029b4ca28d03b4ba7

Preserve the entire current packet and actual failed attempt. Conditional should author only the helper delta plus dependent rebind/index and sibling required-check paths; native exact-delta review and root release precede execution. Reuse only original correction-successor-2 bootstrap2, with unchanged17dependencies/11tools; all19 later stages and4 required checks remain fresh. No current successful static/build/source record is relabeled as successor validation. Completed product CODE/native source closures remain applicable because all679 product bytes are unchanged. No extra product source review or new test surface is needed for this schema correction. No imports, helper execution, tests, source edits or gate changes occurred in this audit.
