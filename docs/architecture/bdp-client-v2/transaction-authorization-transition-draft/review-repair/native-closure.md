# Native exact-delta closure

The two authorized repairs close N1 and N2. One additional mandatory test-oracle mismatch was found by the author during context assembly and independently confirmed below; this packet is not yet fully clear for the selected runtime cohort.

N1 closed: `transaction-authorization.ts:59` now requires an actual string before regex evaluation. The existing migration callback's fifth rejected-open copy stores the genuine generation-1 view as a singleton array (test lines269–274); the helper must observe open rejection and closes unexpected owners before directory removal. It uses the actual format-11 closed database after publication, so this is sensitive to the original coercion bug.

N2 closed: the real three-operation receipt is now create a/delete a/create b with prefix size2. Lines385–390 require the third result to be a created Resource and the actual issued page to equal that entire retained result. Lines404–412 require exactly its withheld replacement while preserving all page identity/boundary fields and comparing current context. Root non-record equality, failed receipt equality, compact allocations0/2 and empty-results behavior remain intact. No new request or successful owner is added; rejected-open copies increase4→5 explicitly.

## N3 — Medium, newly confirmed: stale stream lease expects the wrong error code

`transaction-service.test.ts:2346` still expects `forbidden`. `transaction-service.ts:306` synchronously calls `owner.acquireAuthorizationLease(context)` and does not map its error. `transaction-store.ts:1877–1881` rejects the old selected epoch/view with `TransactionAuthorizationFailure("foreign-view", readContext())`. The test catches that genuine exception, so its expectation is guaranteed to fail and prevents its subsequent fresh-lease branch from being exercised. This is the stale-context error, distinct from mutation permission refusal. Change only the expectation to `foreign-view`; retain all zero callback/write/pull assertions and the fresh-lease/cleanup assertions.

This third finding was identified by the implementation author after my original full review and then independently traced here. I missed it in the original report; that report remains immutable. No runtime test has been executed in this review lane, and no failure is claimed as observed runtime evidence.

## Binding and preservation

Exact two-file delta independently reconstructed from preserved original snapshots and current files; it equals correction `e1f0dedf9c04c608a35516435d59cb3d3cfd3a8dcc3630359032ac911b6b7eaa`. All fifteen current files match successor binding `c3471ed8963747e9a873f28308ecc067b2f4578e36ec49fd99177e8b69870e6a`; the other thirteen hashes are identical to original `2928d83d55f2496bb3687ac0c3e87b1a536bea45fa511cd0a7ff4f2aceb5bcac`. Full patch `beba7a15a94b70d4efc42f369fc851ef233f59032a59af66b0f0db5004d8515f` verified. Original review `71b12ad72efc08fef4256db2079ebada58141d1b026da8c4ff7c8243beaf6394` is retained.

Only read-only source/hash/diff inspections and guarded report writing were performed. Author-reported static results are separate evidence. No source edits, imports, tests, build, services or external review were run. Scope remains private authorization transition; no publication or whole-council approval is asserted.
