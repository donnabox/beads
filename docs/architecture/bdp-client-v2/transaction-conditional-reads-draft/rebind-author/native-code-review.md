# Independent native code review — conditional rebind successor

**Hold code approval and installed execution: one high-severity blocker.** No source edits were made. The four-file frozen draft and patch remained byte-identical throughout this review. This report does not approve execution, altered measurement fences or profile qualification.

## Finding N1 — principal oracle now selects conditional action IDs (High, blocking)

Location: apps/bdptest/src/transactional-matrix.test.ts:4066, inside assertPrincipalPositive.

The changed expression is `declared.actions!.find((a) => a.id === conditionalActionIds[i])!.assertions`. `declared` is explicitly the principal-isolation row. Its first action is `pk-a-create`; its complete actions are pk-a-create, pk-b-create, pk-a-rotated, pk-b-replay, pk-a-conflict, pk-current and pk-events. The conditional list starts with `seed`, which does not exist in that row. On the first iteration the lookup therefore returns undefined and dereferencing `.assertions` throws, even after every genuine principal HTTP observation and portable assertion passed. The TypeScript non-null assertion cannot prevent the runtime error.

This breaks the existing selected principal positive callback and both full/cross-carrier paths that invoke assertPrincipalPositive. In the latter paths it throws before the new conditional finish/tail checks, defeating the plan's explicit preservation of old observer completion and preventing the affected installed cohorts from validating. It is unrelated to the new scenario's correct ID-based selection at line4542.

Bounded correction: restore the original `declared.actions![i]!.assertions` expression, whose order is already checked immediately above, or use the principal-specific `principalRequests[i]!.action` in the stable-ID lookup. Do not weaken or remove the old oracle. Preserve this review packet; author a one-line successor after other active readers finish, refresh its binding and review the exact correction before running the planned selected-principal/full/cross regressions. This finding was communicated promptly to root; no tests were run to rediscover the deterministic failure.

## Reviewed behavior without additional findings

- Actual scenario/session routing is correctly sourced from successful launch arguments. A frozen scenario/reopen/sessionOrdinal object is passed to dispatch, retained across the awaited real target operation and checked by reference before response observation. The callback does not invent an ExactHttpExchangeRequest scenario property or infer identity from URL, response or count. Existing execution/configuration checks remain in the unmodified binding helper.
- conditionalScenarioWire routes only the actual conditional scenario to the new oracle. Other ordinary exchanges retain the principal dispatch and principal→cross/minimum/member response chain. Client-helper execution remains separate. Selected conditional runs reject unexpected nonconditional traffic. The above N1 is the concrete preservation regression; no other observer-routing regression was identified.
- conditionalReadWire requires the same frozen non-reopen session for all dispatches/responses and exact cardinality. finish joins the genuine sessions record by ordinal, checks one conditional session, checks action order and portable assertion IDs by scenario/action ID, and for full cohorts checks all16 parent first presentations in independent order with the conditional session last. Legitimate parent reopens remain in the session sequence. Runner catalog ordering and mustStopMatrixAfter establish that ordinary cross-control assertion failures continue to this tail; exact-observation, cleanup and other stop failures cannot satisfy finish.
- The one-header control remains a literal If-None-Match substitution at conditional-get. The original initial GET independently establishes a distinct actual validator; the altered fourth exchange must genuinely return200 with the full original receipt Resource and unchanged ETag. finish requires exactly status/body-absent failures, four completed exchanges and no transport/harness/prerequisite/cleanup failures. All other assertions, including the new ETag equalsResponse check, must pass. Missing execution, deadline or an arbitrary failure cannot substitute for discrimination.
- Standard-library JSON inspection confirms14 actions and exactly194 portable assertion IDs, matching the194 coverage bindings in order. Both304 ETag checks use equalsResponse:initial-get. Neither mismatch412 action contains the removed ETag-presence or no-media assertion; the remaining HEAD metadata equality accepts both omitted ETags in the existing runner. Reference-only412 ETag/Content-Type observations remain inside the app oracle and are not promoted to a universal protocol rule.
- The complete parent manifest/coverage are reconstructed byte-exactly by removing the new tail object. All32 prior raw objects match HEAD exactly. Parent hashes are cd90592e76cace65b115dbfedf5f604c9c4011ed2e1a0fad954c1edb12aca6cd and 0c991feb128a50625397e337464badcb17d22a0c31e1d57a6b66fa1032a92a51. Principal isolation is restored at index0, with requiredProfile read-update, so the existing eight first-page fixtures spreading that row no longer inherit the proposed Read-only POST restriction. Those runtime regressions remain required validation, not claimed passes.
- Existing-A targets use the fixture subject binding with the same URL; missing targets remain scope-relative. No parser, runner, catalog, fixture, schema or production file was changed. The new row uses the existing Transactional requirement, with wholeContract:false; Read conditional remains explicitly unplanned. Coverage names its omitted minima, epoch/view, erasure/checkpoint, optional-no-validator, matching-specific-tag, If-Match:* and date cases. claimEligible remains false. Counts and measurement fences are not completion percentages or observed execution evidence.
- App reports retain authored credential handles, safe response headers/body digests and nonsecret session identity. No effectiveRequest, bearer value or complete runtime credential configuration is newly retained. Schema/Resource validation and receipt-postimage comparison occur in the independent app observer, while captures remain empty; coverage describes that distinction accurately.

## Scope and evidence

Read the complete four-file diff, complete rebind plan and native plan closure, prior Claude root dispositions, new scenario/coverage objects, relevant actual app observer and installed binding code, runner ordering/stopping/header assertion logic and indexed receipt fixture. Read-only standard-library JSON/hash/raw-object analysis was used; no product imports, tests, builds, services, source mutation, main inspection, network publication or delegation. Root owns independent measurement-fence acceptance and actual Claude review. Existing authoring/runtime/deadline limits were not executed or qualified here.

Source binding SHA256: bf626872ba83bfafab4b957632247684d1169e0b9e3fbf6c627f8c4a644413d6.
Patch SHA256: 245c1247153d38875a9a9c716c323e3ca058b1a985a1eac68473d1d8356e47af.
HEAD: 1ba65f2f192cc84e44eb09ef89b9693ec6cdaf05.
Before and after hashes were checked against that binding; current git diff also exactly matches the pinned patch.

## Four-file SHA256, unchanged before/after

- `packages/conformance/test-support/transactional/manifest-v1.json`: `60c527f6a5673161eacf022c0015d09bea2743d153b8b0d0c86160b76fafeb68`
- `packages/conformance/test-support/transactional/coverage-v1.json`: `33032fe07eb06d2a8968fcd8e5a3cf93cb4637c551868c22da9ea0c4edd587a3`
- `packages/conformance/src/transactional-cumulative.test.ts`: `9f70f8a24ea358b2cd613fd392be68057abd7b1552ec8cc6c4bc1a3241de69f6`
- `apps/bdptest/src/transactional-matrix.test.ts`: `a0a7462aaface5f1442e038e59f9a905b3c2b7c505e9bd57e283737708911569`

## Additional read inputs (SHA256 at report)

- `/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-conditional-reads-draft/conditional-rebind-plan/plan.md`: `574e19154137da4c142629cde00653a3244ca33838b336d011bcc5f6f6f51d53`
- `/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-conditional-reads-draft/conditional-rebind-plan/native-plan-review.md`: `bbe1dbf2e1e7ce6a638b0d9cdaabc1fafc50bc36380dcdf6b5a679a9f16dc792`
- `/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-conditional-reads-draft/claude-code-review/root-dispositions.md`: `ad571142d3c33b6d3c82b90668e0cc554dad2bce317b62d80a5b7120bad8901e`
- `/Users/dbox/repos/wt/vickie/vickie-transaction-conditional-reads/apps/bdptest/test-support/transactional-execution-binding.ts`: `cde506c72a8710eafe58522900f6db70ed3701cf333560252dc8e7c5e706cf89`
- `/Users/dbox/repos/wt/vickie/vickie-transaction-conditional-reads/packages/conformance/src/runner.ts`: `ac20c7fbfe15a5e42a6559e89cd3467562efe9bc74d462478414414ed3efede1`
- `/Users/dbox/repos/wt/vickie/vickie-transaction-conditional-reads/packages/conformance/src/runner.test.ts`: `ad7e756bad1f70ecc6285e19830bc001ebec01899ed868af3fb4b08e888aba86`
- `/Users/dbox/repos/wt/vickie/vickie-transaction-conditional-reads/packages/conformance/test-support/transactional/fixture-v1.json`: `3a7899ced771d4880066d98b0000c53f801ece53458502cc4f3bc641146c9a85`
