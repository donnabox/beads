# Source council disposition and follow-on

Reviewed source f870c97f527f7a763d8d3352e3da28c2b7a87bd1/tree63c3dbac02f92498f4d35469e4e1b10ac48540f4 is preserved in history; it is NOT cleared. Native and actual direct reviews completed. Actual completed09:41:40UTC/exit0/empty stderr/complete final/source unchanged; one High and three Low entries, plus five missing-source questions. Native found one Medium. Third reviewer excluded. Full final reports/receipt retained; local raw stream is not published.

| Finding | Disposition |
| --- | --- |
| Native Medium, orphan wisp auxiliary FKs | Accepted/repaired. A test-only helper pins one owned session, requires FK checks initially enabled, configures conservative transaction settings/main branch, disables checks, and executes fixture/export in one rollback-only transaction. Rollback runs before bounded uncanceled restoration/readback and connection release; restoration failure discards the session and fails the test. Existing exact plane-name, zero-state and whole logical source-state assertions retained. No production FK/session behavior changes. |
| Actual High, Dependency fixture Actor | Accepted/repaired. Explicit test-author supplied to the lifecycle AddDependency call; production validation unchanged. |
| Actual Low2, Type-count oracle | Accepted/repaired. Successful cleanup acquisition independently pins the current supported six-Type cardinality; driver COUNT still reflects fixture descriptors. |
| Actual Low3a, wisp_dependencies refusal coverage | Accepted/authored. Add orphan dependency fixture using current split target schema with an explicit external target and FK bypass helper. No parent wisp masks the specific probe. |
| Actual Low3b, invalid UTF-8 prefix/deleted-count scan coverage | Accepted coverage limitation for this bounded current-core adaptation, no claimed dynamic coverage. Both paths return zero/error by inspection; no new production defect established. These cases are not silently counted as exercised. |
| Actual Low3c, empty-catalog cleanup connector | Accepted documented harness scope. It is deliberately empty controlled workspace; engine-backed lifecycle covers nonempty acquisition separately. Descriptor lookup unspaced SQL would need extension if this mock is later expanded. Do not claim nonempty driver-cleanup fault proof or count it as current coverage. |
| Actual Low4, publication boundary absent from design | Accepted/repaired. Update status/owner-selected source-only base and describe detached bytes, caller Store cleanup, cancellation/publication/stdout boundaries and atomicfile limits. No public contract/default selected. |

## Missing-source questions resolved

1. GetIssueCommentsInTx (internal/storage/issueops/comments.go:18–28) orders created_at ASC,id ASC; equal timestamp oracle is deterministic by source, runtime unrun.
2. Graph AddIssueCommentWithIssue validates UTF-8 and calls ExecuteAddComment directly; that function calls AddIssueCommentInTx, which forwards text unchanged to InsertDerivedComment. ValidateAddCommentRequest rejects blank text for its callers but is not called on this graph fixture route. No new product comment-admission choice is made here.
3. ReadConfigPrefix (internal/storage/issueops/helpers.go:615) uses SELECT value FROM config WHERE `key` = ?; normalized cleanup-driver prefix matches. No runtime claim.
4. graphops/types.go defines ResourceKind, TypeDescriptor.Describes, Owns and IsZero at44/1254/1295/1334. APIs exist with expected receiver/returns; no compile proof.
5. IssueDeleteRequest/DeleteIssue (issue_delete.go:16/28) accept the fields/call shape used. workflowState (workflow_test.go:322) has the assumed signature and returns logical row strings for documented tables. exportSourceState additionally compares comment/config/Type/label/wisp tables. No compile/whole-config or filesystem immutability claim.

Fresh final council must inspect the follow-on source; neither reviewer result qualifies changed source. Only static formatting/whitespace/source inspection ran. An intermediate mechanical edit placed a case row inside the struct declaration; gofmt caught the parse error before commit, it was corrected, then gofmt/changed-source whitespace passed. No build/test/lint/generator/check-mode/CLI/engine/DB ran; Janet owns all runtime. Import2b is untouched and never waits for export.
