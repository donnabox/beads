# Native assessment — successor PR123 landing checker

The old checker cannot authorize this head unchanged. Its exact old-head, direct-parent, source-schema and installed-attribution checks must be split as below. No product blocker was identified in this reanchor inspection, but completed Claude delta review plus root disposition/acceptance remains a hard gate. The account-switched process record alone is not review success.

## Required successor changes

1. Preserve the old checker and artifacts. Bind current H=`727e919c2893bd34d5e72ea4659302a999bc4a6f`, T=`734bf9daf40fd17d0fe2bf86b085598589122f2f`, original base B=`122800511ee6306e4105e1f73bf3658f35c90a14`, intermediate P=`f2ac3f980fc8d35a3b485db04b813ad1b569d94d`, original runtime tree PT=`74327f052cef6026b3ea10e186523ace9c87a128`. Replace direct H→B with exact ordinary H→P and P→B checks. I verified both actual single-parent lists locally. B remains the PR main base, not P.

2. Keep the full B→H footprint exactly the original seven paths. Additionally require P→H to change only docs/specs/bdp.md, with before c2bee88f… and after287a0a64…; bind the full repair.patch hash below or exact equivalent full diff. Verify all seven actual current file hashes against repair/final-source.json; six must also equal the original ff9c binding. Require clean exact H/tree. Current repair final-source has `parent=P` and `mainBase=B`; it lacks the original `sourceBindingSha256` field. Do not copy the old schema assertion blindly: bind both original source-binding and separate repair source-binding by hash, checking their before/after fields explicitly.

3. Preserve historical source/control/installed attribution. Original final-source.json remains headP/treePT and original source-binding ff9c6341…; original source tests and controls bind those exact draft/runtime bytes. Original root-installed-summary sourceCommit/sourceTree must still equal P/PT, not H/T. Keep its actual18 responses, payload digest, child exit/absence, claimEligible=false checks. Assert the repair is spec-only and runtime files remain identical; report `freshRuntimeRerun=false`. New full CI is evidence for H/T; old installed evidence is not relabeled as an H installation.

4. Replace exact CI constants with run38648263438/job116140423947 and bind successful complete log b823305c… below. Retain all required individual successful steps, PR head/branch/base/check readbacks and immediate race checks. Saved successful CI metadata identifies synthetic checkout ac6652f2d235915d87db04b710f4afcff45a6080 with parents[B,H] and treeT. Reverify actual remote metadata at readiness rather than trusting this saved summary alone. Synthetic CI and eventual ordinary no-ff merge must still have parents[B,H], not[P,H]. Update the misleading old “H direct child of B” comment: B is the exact grandparent along the verified ordinary chain; both remain ancestor relationships and the tested merge tree must equal T.

5. Extend completed-review/evidence hashes without replacing original council reports. Add native spec delta review1c86875d…, final repair source, repair patch/binding, actual95-case static results/root validation, CI summary/log/tree verification, and eventual completed Claude delta output+successful exit/source-unchanged evidence and root findings disposition. Require a newly explicit root landing acceptance for H/T/B, new CI IDs and the final reviewed hash maps. Preserve failed OAuth attempts and first failing CI as historical failures. Do not count a running process, empty review, exit1 authentication error, or the old original-code Claude pass as repaired-spec closure.

6. Use unambiguous artifact roots: original F and repair R. If the successor resides in R, the old `E=script.parent` references otherwise silently point old final-source/control/installed lookups at the wrong directory. Give original and repair bindings distinct names/relative paths. Write new readiness/merge reports without overwriting prior evidence. Keep no automatic ready/merge action, no release or public qualification, and both modes' exact main/base/head checks. A moved base still stops for explicit integration review.

## Actual evidence and pending state

The repair final-source88a7bd32… correctly records no fresh runtime rerun; its Claude-blocked note and ci-tree-verification's blocker are historical snapshots. Do not mutate them to imply earlier success. Once Claude really completes, bind a separate updated acceptance/disposition record. At this audit the account-switched process.json records PID17055 and source287a…; no completed review is assumed. The95-case static result supersedes only the earlier22-case proposed selection:21 boundaries+58 contract-plan+7 Read+Update catalog+9 Transactional catalog, not a runtime rerun.

Exact local hashes inspected:

- old landing-check.py:8d63c1e06731124d1b6af8238527a1452b2c534cfc042a6fda9bb5d8d93a2bc1
- repair final-source.json:88a7bd322e05d397011e06bc74ef2f2e5e1cfe5f5803bb1f4aea648abd56bbaa
- repair source-binding.json:cc3f9bb7c37e1120ec88d7fb320d8188e227712c1a22be9ac4386992cd00e83e
- repair.patch:ececd735849ae0033817273276c673adf4f72bbea9df0865f36b516c688356fd
- native-review.md:1c86875d3c78af1772cc156ac5fb807be870ff22bb6977ee2a268c6c9bd94395
- root-validation-summary.json:534763a4cdfab4be5f5b0d66abc652341ae626c0b56e9ecaf1e007f3b2807140
- ci-tree-verification.json:82d1a1f445c044409bc9433341d1e67253c27cf84e211b641daca1936a6d1ac0
- ci-success.log:b823305ca58f91dff21fbe0e508902a8ac24845bca04b0c381cd0a09544213ad

Read-only local script/artifact/Git inspection and this guarded report write only. No checker execution, new remote CI query, test, build, product source edit, push, merge or coordination mutation. This assessment specifies the successor requirements; it does not certify a successor script not yet reviewed.
