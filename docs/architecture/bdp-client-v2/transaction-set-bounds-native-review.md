# Native review: set bounds and required failure fields

**Verdict: bounded pass; no blocking defect identified** at frozen `064db8a99629122f130a91ee8b5189432ffe97e9` versus parent `8d64fcf7fdaa34642fadb3c2c0b8f4729c7778df`. Reviewed author report, prior set-repair Claude findings and exact three-file diff. Author checkout remained unchanged. Independent clone `/Users/dbox/repos/wt/vickie/bdp-transaction-set-bounds-native` is clean at the frozen head. No outward actions.

## Independent verification

Pinned Node24.16/pnpm11.20 frozen install, build and typecheck PASS. Build completed before emitted tests. **225 tests /3 files PASS in2.73seconds**, maxWorkers2: transaction-set, transaction-store and alias-evaluator. This includes the complete41-test set suite, actual default4096 boundary, guard-before-cap rollback/restart, and malformed evaluator-failure recovery controls. Logs `native-{install,build,typecheck,tests}.log`, relevant `.exit`, and `native-tests.json` are alongside this report. Typecheck includes the expectTypeOf negative cases; their verification is not attributed to Vitest's runtime no-op.

No broad workspace rerun, packed smoke, large receipt measurement, or evidence regeneration was performed independently. The author's554-test sweep/smokes/8.39MB measurement remain separately identified author evidence. All independent processes completed normally.

## Findings closed and scope

**F1 required fields:** `transaction-batch.ts:453–480` replaces index-signature-erasing Omit with explicit required type/status/retry plus optional typed diagnostics. Both operation-local and whole-limit branches retain required fields. Runtime receiptProblem checks presence/types, integer HTTP status range and closed retry values before constructing resource/alias failure Problems. It throws on malformed internal output; it neither defaults missing fields nor persists an invalid terminal failure. Valid ordinary evaluator objects retain their existing members and values, with type/status re-assigned to the same values; no wire renaming, default, schema or codec change. Existing failure/receipt regression tests passed. The three malformed-output tests use real evaluator failures, remove a required field, observe rollback/no graph effects, restore the evaluator, reopen to retract the pending claim, and reuse the key successfully for the valid retained failure. These are deliberately internal-fault probes, not new client error semantics.

The five compiler negatives independently require type/status/retry, reject invalid retry and string status, while the positive shape remains accepted. The local type still preserves RFC extension fields and does not seal arbitrary extension names or completely validate Problem semantics. In particular, this is not a new exhaustive restriction of every limit/code/index combination; report does not claim that.

**F2 guard precedence:** `transaction-batch.ts:707–727` checks the client's cardinality range before charging this operation's matched count. This preserves the operation-local409/cardinality-violated/after-state-change/index while a passing guard still allows the transaction-wide413/limit-exceeded/never cap refusal. The new real owner test gives a prior tentative create followed by selection of three with max1/cap2, verifies full rollback and retained409 across reopen, then max3/cap2 yields413 with no fabricated index. Shrinking actual selected state permits a fresh key, while original failed key retains its409. No operational cap was raised and the selection itself is unchanged.

**F7 real default:** the permanent test uses128 actual Beads ×32 no-op sets, without lowering the evaluator cap or reading an exported constant, and observes4096 matches/4128 result entries, no effectPosition, and unchanged graph state. Adding the129th makes the same carrier exceed the count budget and retain atomic413. Both successful and failed receipts replay after reopen. This genuinely pins inclusive default behavior from below and above.

**F5 prose:** deletion provenance is accurately inert for both currently supported collections because Beads/unowned Links do not mint deleted-Resource revisions, and owned transitions remain unavailable. No ownership semantics were changed.

**F3/F4/F6:** explicitly deferred. The documented materialization example distinguishes measured8,392,367-byte receipt from projected1,073,741,824 postimage bytes; it does not say a GiB persistence was attempted. Count bounds are not byte/scan/lock-duration bounds, and paging does not avoid complete receipt materialization. Shared visibility predicate is not full Read checkedGraph validation; repeated endpoint reads and complete scans remain. No byte condition is mislabeled as a count limit and no new failure vocabulary is introduced.

## Limits

HTTP set admission remains closed and unchanged; no public profile/configuration, schema, evaluator, store-format, compiler or receipt-codec edits. This is a repair to the private set core and its tests, not complete transactional qualification. Materializable receipt delivery still requires a separate design. Root owns any ordinary integration with compiled-contract and other reviewed successors and combined validation.
