# Snapshot integration ancestry gate

The earlier checkpoint said to merge PR104's actual main merge commit into Snapshot before readiness. That refresh is unnecessary if PR104 lands exactly the verified e925 tree: Snapshot candidate3c9 already has e925 as its second ordinary parent, including every PR104 source change. A further main merge would add no source changes but would reset CI on a new head.

The gate is instead checked directly: PR104 actual main parents/tree must match recorded verified result; e925 must be an ancestor of3c9; no other main changes may intervene unnoticed. Snapshot may merge by an ordinary GitHub merge after its exact3c9 CI and completed council gates. Actual merged Snapshot parents must be the observed current main merge and3c9; actual tree must equal tested3c9. This preserves all main ancestry in the resulting merge without rebase, squash, skipped CI or relabeled evidence. If main content changes, reassess/integrate and validate rather than assume compatibility.

This is an execution-efficiency correction to root's sequencing, not a change to protocol, user merge authorization or review policy.
