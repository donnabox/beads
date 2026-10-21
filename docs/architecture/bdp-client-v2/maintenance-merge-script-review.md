# Read-only PR121 merge-script review

Reviewed `merge-pr121.py` SHA256 `d8177b6e34983719688d1198233af6c5e48eada8671807f9afe2326b5d195429`; did not execute or modify it.

No accidental PR120 head/tree/run/count/output substitution found. PR121121, head31e74290, tree5c14, run38625916119, expected9261passed/1skip/172files plus7preflight, four required successful downstream steps, and maintenance output paths are correct. The older comparison/W1b/headroom review pins are legitimate inherited dependency gates, alongside the five maintenance pins. Independently verified all15 pinned review artifact hashes against current files. Final ordinary merge tree/parents check is explicit.

Narrow precondition corrections recommended before use:

1. Line5 currently extracts only `main` from the PR120 artifact. Validate the entire expected prerequisite record first: pr.state MERGED, pr.headRefOid40806dd3, pr.mergeCommit.oid==main, treee870145f, candidateTreeMatches true, and exact ordinary parents [oldmain7e46,408]. The later408-ancestor/current-remote checks constrain content but do not themselves establish that this is the actual intended PR120 merge record.
2. Lines55–57 should require exactly two synthetic CI parents and exact ordered equality [acceptedCIBase,31e], rather than indexing two entries without checking count. AcceptedCIBase remains old7e46 or the actual verified PR120 main. This is a narrow exact-shape strengthening, not a new CI requirement.
3. Refresh PR head/base/state/check rollup and remote main immediately before ready/merge, after log/API inspection. Current observations precede that work. `--match-head-commit` protects the PR head only, not the base. Continue serialized root ownership and the post-merge tree/parent verification; a final recheck reduces the window but cannot make separate GitHub API operations atomic. Stop if base/checks moved. No generalized automation redesign is required.

The prerequisite artifact is not present yet, so script currently stops before any outward action; that is appropriate until PR120 actually merges. Do not run merely because these static checks pass: actual exact-head CI and prerequisite merge gates remain necessary. No test/build/install/service/network or source mutation occurred during review.

## Root correction closure

Closed the three bounded recommendations by inspection of revised scripts. PR121 SHA `b73a698027bada5dd9a6b49e8783e426be3d4dcfa752a50bfad5633f1daab328` now checks the prerequisite PR120 MERGED/head408/treee870/exact ordinary parents/mergeCommit-to-main identity, requires exactly two CI parents, and refreshes open/head/base/nonempty successful check rollup plus remote main immediately before readiness/merge. PR120 SHA `2d3d01a8faf32ca55c37fb55e8b633e1c042ec42d181276b5b495f6166fbe2ab` has the same final refresh and already requires the exact two CI parents. The decisive prerequisite tree/parent/merge fields establish the intended record without relying on its descriptive boolean. No new copied-pin or scope error found.

No scripts were executed by this reviewer. Remaining separate-call race is bounded by root's serialized ownership and final actual-tree/parent verification, not claimed atomic. Closure is conditional on the scripts' actual CI/prerequisite gates succeeding when root later runs them.
