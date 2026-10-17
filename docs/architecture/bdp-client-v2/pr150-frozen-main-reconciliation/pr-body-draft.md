# Draft PR title

Reconcile PR150 Transactional baseline with the Core/Write specification split

# Draft PR body — validation pending

PR150's Transactional implementation and tests still referenced the monolithic specification and schema paths moved by #145. This ordinary merge preserves the implementation baseline, adopts the split Core/Write layout, and repairs the four candidate-only schema readers so they can execute against the reconciled tree.

Candidate `d7197e97a29bff8f075028419afad05388e2e655` has ordered parents PR150 `d4f3f4be` and frozen main `8d6671ed`; tree `630a94d0`. The old specification path remains #145's navigation page. The accepted d4 minimum-read block moves into Write: missing/deleted/hidden current targets retain uniform404 before semantic comparison, while existing authorized current Resource/properties reads with a matching opaque minimum use bounded waiting and503. That behavior is intentional; it is not part of the editorial relocation.

The reviewed editorial-equivalence comparison covers the common-base normative body after documented relocation normalization, excluding the changed preface. It does not assert that every inherited document is textually or semantically unchanged. Stable schema identities, sealed historical citations/evidence and bounded manifest additions are preserved.

Native and actual Claude CODE reviews are dispositioned. Local static checks and build passed. The first full suite failed in required bd initialization because a temporary workspace beneath `/Users/dbox` discovered an ancestor `.beads` database. An isolated reproduction succeeded under `/private/tmp`; a fresh full validation sequence using owned private temporary directories is prepared but pending. This is not a full-suite pass. Fresh packed checks, exact-candidate CI and publication remain pending.

Public Update/Transactional capabilities remain incomplete; these private observations do not qualify or advertise a profile. S1 retained Snapshot/continuation minima and G2 incident-Link metadata Event membership remain unresolved and separate. Later write-profile branches, retention work and the alias draft are excluded. This PR does not authorize landing main.

Preparation evidence: [sole reconciliation plan](/Users/dbox/repos/wt/vickie/bdp-integration-logs/pr150-frozen-main-reconciliation/reconciliation-plan.md), [review dispositions](/Users/dbox/repos/wt/vickie/bdp-integration-logs/pr150-frozen-main-reconciliation/root-code-dispositions.json), and [pending isolation correction](/Users/dbox/repos/wt/vickie/bdp-integration-logs/pr150-frozen-main-reconciliation/validation-private-tmp-successor/plan.md). These local preparation links must be made reviewer-accessible before publication.
