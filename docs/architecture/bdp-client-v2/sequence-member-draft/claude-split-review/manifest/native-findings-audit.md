# Independent manifest-scope findings audit

No required source correctness repair established. Read all six findings and Verified/Omitted-context sections against frozen source and actual evidence. Root retains disposition; the separate wire review is not adjudicated by this report.

1. **Not a defect: ordinal is per authority change group.** `docs/specs/bdp.md:6092–6093` says zero-based authority-group ordinal, and`:6230–6233` calls it stable change-group ordinal. `transaction-batch.ts:692–706` takes ordinal from the current evaluator's events.length; each one-operation sequence transaction gets a fresh evaluator. Store`:1234–1255,1280–1309` binds each core.transaction into its own evaluation/group/receipt. `[0,0]` is correct for separate single-created-Event groups. App`:3099–3133` also strictly compares both full Events, ordinal0, distinct checkpoint-specific IDs and receipt transaction links. Do not change second ordinal to1 or weaken the oracle. Coverage need not claim a special atomic-control proof: distinct receipts/transactions/effect positions and success–failure–success already discriminate the accepted narrow behavior; no atomic mutant was run.

The review's “never observed because review timeouts” inference is false. Actual local positive report has the scenario pass, all10 exchanges/all209 assertions passed, including s-events:ordered-facts; sequence wire supplement completed seven raw responses and full replay comparison. Source eight-case acceptance and CI132 were independently recorded. Review CLI timeouts are inconclusive reviewer outcomes, unrelated to product executions. The redacted report does not persist raw Event bodies; passed assertions plus frozen helper give the actual oracle evidence.

2. **Cross-scope gap closed.** Actual `apps/bdptest/src/transactional-matrix.test.ts:1736–1737,2147–2148` has both derivative fences812575/21176. These are separate from canonical-normal806457; no derivative is checked against the smaller canonical bound. Accepted measurement maxima808479/21112 plus4096/64 exactly justify them. The manifest-only excerpt understandably omitted the other owned file. Numeric scope is complete; comment freshness is a separate Low item in wire review.

3. **Optional static-helper sensitivity addition, not missing runtime coverage.** Cumulative`:1022–1036` checks both indices0/7;`:1058–1060` independently pins entire row48e52459…; controls`:1065–1077` intentionally demonstrate initial middle-status and retained failure-index properties. Current replay assertion and full-body wire comparison are present and actually passed. A third static removal at index7 would improve helper demonstration but is not needed to repair functionality or a weakened row; defer rather than expand the completed selection without a concrete failure. Changing helper loop alone does not remove the independently pinned manifest replay facts.

4. **Accurate Low stale title; optional editorial cleanup.** Cumulative`:1063` still mentions one cross-carrier row although callback now includes sequence/minimum preservation. This does not change asserted behavior; exact selection/acceptance records deliberately use that historical title and passed. Renaming requires updating selectors/records, so retain during this bounded landing unless root explicitly wants that fanout. Do not describe title change as a product fix.

5. **Pre-existing arithmetic redundancy intentionally retained.** 488280+305825=794105;13019+8001=21020;70662+48064=118726. Delta fences record historical bases while absolute fences enforce current maxima; no independent extra protection is claimed. No removal/refactor needed in this evidence slice.

6. **Intentional fixed preservation layers, no mismatch found.** Cumulative`:837–1001` pins28 raw parent objects and then full reconstructed parent bytes; nested withoutMinimumRow retains the older13-row basis. Individual pins locate drift while whole-parent hash pins preambles/format too. They are deliberately independent constants, not derived from the newer mutable table. Root-approved four-path/28-object preservation requires retaining them; no deduplication needed. Whole-parent and all28 pins actually passed in the selected source callback and fullCI.

Omitted context closure: profile spec`:6937` already names transactional.sequence.member-transactions; cumulative selection uses existing251 required IDs/256 total and new row is partial, without adding clauseMap/public credit. Exact action counts/209 assertion mirror/28 raw objects are covered by actual selected source validation and whole-row/parent hashes. Other schema dependencies are handled in the separate wire audit; no new schema assumptions introduced here.

Evidence, with historical attribution preserved:

- claude-split-review/manifest/stdout.txt: `797e7cb90f042e68b7610951db1c1cb12ecfcc29765d0bcae0f6d1df682e6c98`
- source-binding.json: `a6eb425d2616279fed28132c2b59258b7503b186224ba538179704b6364c9e2a`
- root-source-acceptance.json: `15e91b6ab40beda200189a569d9475eaedc6ffc09cb7d0154c4ca9bececbdf39`
- installed-preparation-v2/run/positive-fifteen.observation.json: `e7c2d0195b6c7c14d05b4ef2bdb9bd8cfb9ca515974c74d80f28e620a1a7fd12`
- installed-preparation-v2/run/positive-fifteen.review-candidate.json: `ee76d6b00e17f35f47431aea7a9ceba307cb0fb0d3529c5c6a24550a5a8011eb`
- landing-preparation/native-ci-audit.md: `147a58d81f115795d124b54328b3942e9d5f0175cf7916e43129cd27201318bf`

Read-only source/JSON/log inspection; no tests, product imports, builds, services, network queries, plan/source changes or reviewer reruns. Only this report written under guard.
