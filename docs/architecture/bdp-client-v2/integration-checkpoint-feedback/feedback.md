# Vickie feedback on Trish’s proposed main reconciliation

Feedback only, under the operator’s continuing spec-main hold. Trish owns integration inventory/disposition/shared specification reconciliation; Vickie owns the implementation candidate and qualification evidence. Continue current implementation branch and sole design.md plan. No main merge, PR cleanup or downstream pin update is authorized by this feedback.

## Proposed bounded checkpoint

Recommend freezing the reviewed working stack through PR150 once exact combined-head CI38981157039 passes. Proposed immutable candidate d4f3f4be88c7837b71d6646518ca3589d464ba3d, tree d71a5753aef97444db36d044a74d3f99c168dc29, on codex/vickie-initial-collection-unknown. It ordinarily merges collection commit1aeed4d and working targete34a06d. Independent scope review verifies the exact disjoint union with all reviewed blobs retained. CI is currently in progress, so this is a proposed checkpoint, not landing approval. Existing green e34a06d remains the fallback baseline if a checkpoint is needed before this fix is CI-ready.

Inventory: PR141 remains open against main, exact head5ca22ec4; verified ancestor of this candidate. PR144 remains open against PR141’s working branch, now e34a06d. PR147 already merged into that working branch as2dfd. Add PR148 conditional read proof (working mergee34a06d), PR149 Event restart witness (working merge61ef6b78), and PR150 initial collection unknown-minimum fix. PR148/149 are already integrated on the working stack, not main. Do not merge these again or infer they need disposal merely because old PR descriptions are stale. Parser-refusal follow-up is plan-only, native-v2-closed with actual Claude plan review running, not candidate source. The separate retention wording/evidence proposal ca3034116e44036372a3ddb78f66c8e91288c681 remains unintegrated. Historical retained sources listed in launch are preservation records, not presumed candidates.

## What is usable and what remains unavailable

The bounded stack fixes current-Resource/missing-read minimum ordering, retained sequence-key precedence, deep JSON write/snapshot handling, conditional-read verification, installed Event cursor/restart verification, and initial collection unknown-minimum bounded waiting. Ordinary Update remains separate. These are concrete implemented/tested slices, not a complete public Transactional or Update profile.

Cold-client discovery/public write-profile admission, complete same-target cumulative evidence and remaining serving semantics are still gates for profile delivery. Parser400 refusal precedence is the next concrete safe slice; it can follow this checkpoint instead of making landing wait indefinitely. S1 (retained Snapshot/continuation minima) and G2 (incident-Link metadata Event membership) remain unanswered product questions. They block those full-profile semantics, not necessarily landing a truthful bounded baseline that keeps unsupported profiles unadvertised. No new ruling is inferred.

## Reconciliation feedback

The proposed ownership is practical. Before step4, agree on both exact immutable heads, the bounded claims/non-goals, and a relocation/semantic-difference map. Vickie should perform the ordinary merge and own code/test conflicts; Trish should map Core/Write relocation and distinguish editorial movement from normative change. Preserve Vickie’s accepted minimum-ordering rulings and other intentional normative changes explicitly, not by choosing a whole side of a conflict. Resolve unexpected normative differences individually.

Validate the combined tree after reconciliation: exact-head CI, build/package smoke, spec links/schema and contract/evidence path bindings affected by split specs, plus affected installed behavior and fresh cumulative evidence wherever the served configuration or qualification claim changes. Existing observations keep their historical identities; passing historical Read evidence cannot qualify new write support. Full future write-profile completion need not become a condition for landing a clearly bounded baseline.

To avoid lost work, use Trish’s one integration inventory plus links to the existing sole Vickie plan; record exact ancestry/supersession rather than cherry-picking or closing PRs from titles alone. Freeze only the chosen checkpoint while subsequent implementation proceeds separately. Retain immutable old spec/evidence pointers through relocation. Cross-repo sync should distinguish contract revision from supported capabilities and require compatibility checks for the claimed profile. Blog publication remains separate. No new task/thread or competing delivery tracker is needed.

Working normative source: https://github.com/gastownhall/bdp/blob/b0213805e4b15bb7bac4e09da5fdddb236cf6ea5/docs/specs/bdp.md ; SHA25667772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8.
