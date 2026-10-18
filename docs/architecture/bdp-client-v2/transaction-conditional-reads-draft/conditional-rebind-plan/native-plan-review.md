# Native review of precise four-file conditional rebind plan

Approve bounded authoring of the plan as written. No blocking or actionable plan findings. This approval covers only the existing four owned files and the explicitly partial Transactional requirement; it is not code approval, execution evidence, profile qualification or permission to expand parser capabilities.

Reviewed plan SHA256 574e19154137da4c142629cde00653a3244ca33838b336d011bcc5f6f6f51d53, proposed-measurements.json, actual frozen four-file draft and relevant supporting source at HEAD 1ba65f2f192cc84e44eb09ef89b9693ec6cdaf05. Also read claude-code-review/root-dispositions.md and claude-code-review/native-findings-audit.md. No tests, product imports, source edits, external publication or delegation were performed.

## Routing and lifecycle

The proposed source of scenario identity is sound. apps/bdptest/test-support/transactional-execution-binding.ts:113 calls the local launch callback with the runner-selected scenario ID and actual reopen flag. Its existing binding checks still validate ready metadata, served configuration, runtime identity and active session before successful exchange. ExactHttpExchangeRequest supplies no scenario field, so capturing actual launch context locally avoids fabricating one or deriving identity from request contents.

The plan explicitly records context only after successful launch, snapshots that frozen context at ordinary dispatch, and requires reference identity unchanged after the awaited real target operation before response observation. This preserves dispatch/response pairing even if a session transition were unexpectedly interleaved. Only conditional scenario exchanges enter the new wire oracle; every other ordinary exchange retains the principal/cross/minimum/member chain and its original ordering. Client-helper exchanges remain on their existing separate path. No callback is allowed to replace a response or grant a target capability.

The proposed single conditional session, reopen:false, matching session ordinal, exact14/4 dispatch-response cardinality and action-order report join are appropriate independent checks. Full cohorts additionally require all16 parent first presentations in their original order, retaining legitimate reopen sessions and existing old observer completion checks. Selected runs use the same context adapter without assuming the full-cohort ordinal. This removes the obsolete prefix count assumption; keeping the prefix or passing a count-inferred context would not implement this approved plan.

The actual runner preserves catalog order for scenarioSelection. Its mustStopMatrixAfter does not stop the matrix on ordinary assertion failure, so cross-carrier controls should reach the new tail scenario. Their existing checks already require all other planned scenarios to pass; retain those and explicitly finish the conditional wire at the tail. Transport, preparation, context or cleanup failure must remain a failure, never substitute for completion.

## Placement, first-page fixture and preservation

The existing Transactional conditional catalog row follows every parent planned row. Appending it at planned index16 restores principal isolation at index0 and receipt pagination at index11. The eight existing first-page authority cases in runner.test.ts spread index0: restoring its original read-update row restores the accepted POST/body profile without touching the runner. Validate those unchanged cases as planned; this does not cure their pre-existing positional coupling generally.

The new row and controls must be accessed by stable ID. Placement assertions may check index16 but must not become row selectors. Tail removal of exactly the newly added object's comma/newline/indentation is a plausible byte-exact parent reconstruction; final authoring must demonstrate the unchanged complete parent hashes and all32 old raw-object hashes. Historical source/evidence is retained, and no original coverage row is rebound. Read conditional remains separately required and unplanned; only the unexecuted draft changes attribution.

## Portable assertions and evidence scope

Actual current row has14 actions and198 assertions. Replacing two304 ETag-presence checks with existing equalsResponse:initial-get is count-neutral and supported by runner.ts:2086–2102. Removing the four named412 ETag/media-presence assertions yields194 and preserves native bodylessness/current context checks. response-metadata-equals accepts both omitted metadata values, so retaining HEAD parity does not secretly reintroduce mandatory412 ETag. Reference-service412 ETag and media observations remain app-specific. The fixture's subject binding is exactly https://reference.invalid/acme/beads/a; replacing only that existing target spelling changes no wire URL.

The single-header control still receives a genuine200 with full Resource body and unchanged initial validator at request4. Thus its new ETag equality should pass while exactly status and body-absent fail. The independent app oracle must continue proving prerequisites and rejecting extra failures/requests, rather than counting any failure as success. Captures remain empty; full Resource-to-receipt-postimage equality belongs to the app oracle, not portable manifest capture. Whole-contract false, claimEligible false and the listed omissions make the narrowed TX attribution honest.

## Validation boundary

No production/parser/bundle/schema/catalog/runner changes are needed for this plan. Refresh route ordering/configuration identity, row/assertion pins, actual42 derivative measurements and selector names after authoring. Do not copy provisional digests or automatically widen fences. The seven-cohort arithmetic712 expected HTTP/727 diagnostic ceiling/83 children is internally consistent with the enumerated prospective cohorts; it is not observed evidence. Actual execution, finite workload inventory for existing first-page/binding regressions, source review and installed payload provenance remain root-controlled gates. All existing runtime and callback budgets remain unchanged.

## Frozen source SHA256
- packages/conformance/test-support/transactional/manifest-v1.json: 479ec242a91668d5b54abd5711e75658b3ea9e57a20db49ac3eba3be1f3f63f2
- packages/conformance/test-support/transactional/coverage-v1.json: 2d73a595ae02fef6e86b1c06bdef9e3549e004a16943e1b0afc76ff23e137e15
- packages/conformance/src/transactional-cumulative.test.ts: 60acebc7f3bacc2b405a42b27b4d3a2d38473ba31fbbe4ec13b97a763d5bfde9
- apps/bdptest/src/transactional-matrix.test.ts: 50470a99ecf9091d6cc4599fcf359bbe8085ab289968039f7900e25fa660a719

All four hashes and the exact plan hash were rechecked before this guarded report write.
