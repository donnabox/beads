# Set core / recovery integration and external plan dispositions

Frozen clean head `fccb951ce0a470526a22db9762236935f78ec5e8`, branch
`codex/vickie-transaction-set-integration`, checkout
`/Users/dbox/repos/wt/vickie/vickie-transaction-set-integration`.

History: ordinary merge `f67f2bc` preserves core `ca603d8` and recovery repair
`270a1216cb918a4194f3b195db6c8fcbecf0c040`, followed by bounded test commit
`fccb951`. No rewritten parents, pushes, PR/bus/launch writes or profile claims.
Old core and paging review checkouts remain frozen.

## Merge inspection

Only conflict: additive paragraphs at the end of the transaction progress note.
Both intents were retained, with the old "repair pending" sentence corrected.
The repair's compatibility wording now allows SQLite hot-journal recovery before
the locked configuration check, while that check still precedes BDP pending
retraction, creator cleanup and retention maintenance.

The store auto-merge was inspected. Relative to exact 270a121, its diff consists
only of approved format `set-core-7` and the lifetime-checked staged collection
facade. No receipt ordering, crash recovery, fencing or cause preservation edit
was lost. The read-only preflight was not reintroduced. Service/HTTP/runtime
behavior comes unchanged from the reviewed repair ancestry.

## External plan review disposition

Read the complete `transaction-after-paging-claude-plan-review.txt`.

- **F1:** covered without redundant operationCount. The implemented private carrier
  holds the FULL ordered union, not just the RU Resource subset. Its operations.length
  is exactly the outer count; the subset exists only temporarily for Resource syntax
  validation. New test covers one Resource plus two sets against max 2 and 33 set
  members against max 32; actual owner refuses before admission. Closed-database
  receipt table remains empty, and corrected reuse of the same key is admitted.
  HTTP sets remain closed, so this is authoritative owner evidence, not a served 413
  claim. Existing comparison/admission tests preserve duplicate-before-capacity law.
- **F2:** accepted as unresolved normative boundary: set Selector size/depth bounds
  at 4275–4278 versus the five enumerated post-admission transaction limits at
 5317–5319. No new receipt code or limit interpretation has been encoded. The
  existing Read mapping is implementation behavior, not a ruling for set bodies.
- **F3:** rejected per root. No pre-key all-Selector-error-to-400 shortcut. The normal
  prepareTransaction and HTTP paths still refuse sets. Experimental core preparation
  accepts known valid bounded selectors; any grammar/complexity error is a private
  TransactionDevelopmentLimitError, not a protocol Problem. Reparse uses the same
  fixed bounds at the actual staged selection turn. No artificial large bound or
  new public selector API was introduced.
- **F4:** exact cardinality shape is checked before the syntax-only RU projection,
  using actual JsonNumberLiteral coefficient/fraction/exponent structure. Zero
  placeholders never validate cardinality. Negative tiny decimals, fractional
  decimals and exponent spellings retain their real semantics; no invented max-safe
  bound. Added canonical bundle cross-check corpus for both set operation roots,
  including min>max, omitted bounds and invalid cardinality shapes. Existing precise
  literal tests cover values whose native Number would underflow/round, and numeric
  admission diagnostics use original /change paths. No HTTP mapping claim.
- **F5:** all normalization, witness, alias and evaluation arrays use the full outer
  union. Set positions persist exact-kind empty alias witnesses. Resource parser
  synthetic keys are syntax-only, neither stored nor identity-bearing; no carrier-
  index-to-outer-index remapping is used at execution. Existing tests cover original
  creation indices after sets and label renaming. New test places a set before an
  alias Resource mutation, repoints the alias after admission, verifies captured
  target and outer operationIndex on a tail page, then reopens/replays both original
  alias and equal direct canonical locator. Existing failures retain outer indices.
- **F6:** format 7 is implemented and approved, with no migration/reseed. The review's
  read-only-preflight recommendation is superseded by the actual hot-journal defect
  and 270a121. A genuine format 6 writer probe with pending receipt was repeated against
  the integrated emitted owner: refusal preserves file bytes. Valid hot journals
  recover before BDP compatibility/cleanup as established by crash tests.
- **F7:** explicit scope exclusion. Current provisioning rejects owned Types and the
  owner policy is all-visible/all-writable. Unowned provenance tests do not qualify
  owned-source fanout or selected-but-nonwritable authorization. These branches remain
  open for later installed configuration/policy integration.
- **F8:** conditional follow-up only. TransactionSubmissionResult.refused.limit stays
  unchanged because HTTP sets remain closed. Any approved direct Selector-limit
  policy will require coordinated preparation/submission/HTTP type and evidence work.
- **F9:** normative schemas are the bundle's updateWhereOperation/deleteWhereOperation
  (with their Members/Request definitions), not the Markdown preview sketch. New
  corpus tests compile those exact bundle roots. Normative prose 4904–4918 remains
  the selection/visibility anchor. No schema changes.
- **F10:** implemented as approved private 4096 module constant, evaluator-only lowering
  seam, no persisted/public installation option. Total Event failures omit a fabricated
  operationIndex. Test crosses the bound with combined ordinary and set facts and
  observes rollback plus retained failure/replay. No new setting/tuning policy.
- **F11:** deferred with route activation. No HTTP operation table was changed in this
  core. JSON object order is not a new normative requirement. Future Directory work
  can use familiar documented order without advertising absent routes now.

## Verification

Pinned Node24.16.0 / pnpm11.20.0, frozen install; build before process tests.

- PASS **545 tests / 14 files**, no skip, in `regressions-verified.log`: all transaction
  test files plus Read Selector and Resource evaluator. Includes 23 set-core tests,
  and both actual process-death hot-journal maintenance/terminal recovery tests.
- PASS build, typecheck, dependency boundaries, format, lint. Lint retains only the
  existing 3 warnings / 1 info in untouched files; no new warnings.
- PASS installed package smoke including private module import denial, development
  smoke (`claimEligible:false`), and unchanged 89-row Read evidence verification.
- PASS actual prior writer probe `actual-format 6-probe.mjs/.log`, using emitted
  frozen 0ff writer and emitted integrated format 7 opener with byte equality.
- Initial integrated regression run:544 pass / 1 test-fixture failure because inspect()
  does not expose receipts. Corrected the assertion to query the closed database
  directly before same-key reuse. No runtime behavior changed for that failure.
- No interrupted runs, killed test processes, hooks bypassed or outward writes.
  Full workspace suite was not rerun by this worker; root owns broader integration.

Logs are in this report's directory. The source is clean and frozen for independent
review. This remains a private valid-set core: HTTP set admission, negative Selector
semantics, installed owned/schema/policy behavior, remaining consistency, replication
and full Transactional discovery/qualification are not claimed complete.
