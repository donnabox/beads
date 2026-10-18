# C3 code-review dispositions

Native review confirmed a P2 compatibility regression in service pagination capture: freezing a shallow clone changes the receiver to which authority-read binds clock and token callbacks. The differential probe uses an actual owner/service and stateful clock method: prior service returns the collection and increments original ticks; b20 returns temporarily-unavailable503 and leaves original state unchanged. Root inspected original authority-read callback binding at395–397 and C3 capture delta.

Accepted repair: capture the supplied pagination object and callbacks once; preserve original callback receivers while freezing scalar configuration. Retain ordinary validation and getter-capture semantics. Add actual stateful clock and token pagination tests plus current nested-limit mutation regressions. No C4 or callback policy redesign. Fresh frozen successor evidence is required because production emitted bytes change.

Original author repairs isolated codex/vickie-tx-discovery-repair at /Users/dbox/repos/wt/vickie/vickie-tx-discovery-repair. Original b20 worktree stays frozen for external review, preserving source-bound evidence. Root owns PR106 advancement after reports and findings disposition. Do not infer readiness from b20 CI or historical installed12/1. External review remains active; additional findings will be dispositioned before integration.


## Native repair closure at3f486c2

Root read the complete two-file repair and independent closure.3f486c245bfd31f41fd42f5acc8e0f13683bad42/tree2a68550d01df1cf829b9fd76183fa879e5f4acbe preserves original callback receivers, one getter capture and prior invalid-callback validation while freezing copied scalars. Author red/green test and native three-variant real-owner probe establish defect removal.202 affected tests and fresh12/one-child plus two targeted Update scenarios pass at new payload5e8ce6bf; originalb20 proof is preserved. No remaining native finding.

Root advanced the PR106 remote head by ordinary fast-forward to3f486 without changing the originalb20 review worktree. This improves CI throughput while keeping external review's exact source immutable. It supersedes the earlier sequencing note to wait for external completion before advancing the remote head; readiness still requires its findings disposition and exact3f or later final-head CI. No local review source, evidence attribution, or merge gate is weakened. Repair branch remains separately preserved.


## External review disposition — b20 source, 3f production repair preserved

External review completed successfully and independently verified original source and packed provenance. Both additional findings are accepted. They do not require a product ruling or expand public serving.

1. High: the protocol index export surface changed but its installed-schema interpretation policy pin did not. Root read the policy version rule and tripwire. The two additive Transactional parser/type reexports do not change graph admission, shape, URI/reference resolution, built-in closure, errors or budgets. Their private lazy validator does not replace the graph interpreter. Retain schema-resource-index-2 with explicit reviewed justification and refresh only index.ts from c14d73d16d998ac18b1532e6603c36916493995a342340893761d64286735291 to 92e547e807ec2d23223c47dbc4e04867d55db89ac9aab5fbd55f33349f92bdc6. Other pins must remain unchanged. The policy file is an accepted fifteenth file in the original C3 inventory. Future affected-test selection must include installed-schema-graph.policy.test.ts whenever a pinned input changes.
2. Medium: receipt and Snapshot navigation currently have literal oracles without actual route witnesses. Add receipt lookup through the source handler using the existing real receipt and observed discovery URL, and Snapshot creation through the observed discovery URL. Rename the literal projection test accurately. Existing fixtures suffice; no transport or product behavior change.

update_implementation owns a four-file test/policy successor in the existing repair checkout. Production bytes must remain identical to3f; preserve its fresh packed evidence at its original source rather than relabeling it. Run the policy and affected source tests, compare production identities, and obtain independent native closure. Root will advance PR106 ordinarily and require final-head green CI before merge. The known stale-pin3f head is not merge-ready. The independent limits lane may continue on3f and incorporate this successor ordinarily before promotion.


## Final external repair closure at a53f4bb

Root read the entire four-file successor and independent native closure. Both external findings are closed; all fourteen pins independently match, only index changed, actual same-principal receipt and Snapshot source-handler witnesses pass. The84-test focused run and static checks passed. Native independently compared63protocol/246server member bytes and whole archives equal3f. No production change and no installed-proof relabel. PR106 advanced ordinarily to a53f4bbf573c64fe900de373b3fef75170002f7c; exact CI38545418962 remains required. Council findings are dispositioned; readiness awaits CI and final remote scope/ancestry check.
