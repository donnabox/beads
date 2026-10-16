# Native CI expectation repair review

No actionable defect found. The one-callback repair matches the current handler; the original CI500 expectation and the superseded unexecuted no-observation premise were both incorrect for this negotiation branch.

Reviewed exact dirty source atop77b4cf980ec91f66c52a554c3dd8c9de5aee859f:
- transaction-service.test.ts SHA2568f5e3875f09522d3e0678a5cb55eb495d0a6aa2256fd1ef655c7b5edcdb506f3.
- Exact HEAD-to-file patch29faf0265b6e4c08d79cb72286561cebcda6a568c46119d4185609310657158a.
- Working-tree status contains only this modified path; patch bytes match the packet. Production is unchanged.

transaction-http.ts332–334 explicitly calls sequenceRefusalObserved for unacceptable JSON negotiation after valid sequence preparation, then attaches its current context to empty406. service619–628 performs a real owner.withRead observation, without member admission/execution. Therefore the sixth row correctly requires exactly one observation, actual owner epoch, fixture's development-all-visible view, pos-0, absent private-limit/content-type, undefined body and Content-Length0. The status code field “not-acceptable” labels the test row; it is not asserted as a nonexistent Problem body.

The original first five rows remain unchanged: outer key400, malformed syntax400, unauthenticated401, forbidden403 and media415. Their body code/status checks and absence of all three context headers remain; after each, the cumulative observation spy must still be untouched. The sixth and final row is the sole observed refusal, making the cumulative count exactly1 meaningful.

The added sequenceObserved forwarding spy is asserted untouched after every row. Together with final ownerposition0, this retains the intended nonexecution proof rather than using zero position as a substitute for a method-boundary assertion. The real service observer remains delegated, not mocked to a fabricated context. Existing fixture cleanup is unchanged. This is six in-process handler calls, not six real HTTP/socket exchanges; no new child/listener workload is introduced.

Read existing actual validation artifacts: format0/no fixes, typecheck0, selected callback0 with1passed/66notselected. These are root-run outcomes, not execution by this reviewer. The original push CI9348pass/1fail/1skip remains failed; the focused pass does not establish a new full CI result.

No source edits, test/build/import execution, reviewer launch, main query or Git mutation were performed. This report is native review of the frozen callback, separate from the concurrently running actual Claude review.
