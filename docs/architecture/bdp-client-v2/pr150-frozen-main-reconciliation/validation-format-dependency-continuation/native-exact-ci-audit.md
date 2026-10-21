# Independent PR159 exact CI audit

PASS. Push run39030688462 executed exact candidate d7197e97a29bff8f075028419afad05388e2e655 and passed all27 steps. PR run39030713743 also passed all27, but its actual checkout is GitHub synthetic merge835e75b78c83eee252cb513d142cc51e0efd5261, not the candidate commit despite metadata headSha=d719. Read-only GitHub commit verification confirms synthetic parents [exact8d6671,d719] and identical tree630a94d0ca1545f898cb8a8b614d4b7d2dd6d454. The push run provides exact-head execution; the PR run adds same-tree merge validation.

Both complete raw job logs independently show9375 tests passed,1 skipped,187 files passed, and7 additional preflight tests passed. The required isolated pinned bd CLI matrix passes (33501ms push;34035ms PR), as do all five selected SSE/restart/fresh-depth/current-minimum/retained-sequence callbacks. Source/version/build/test/smoke/evidence steps all conclude success. No test rerun occurred during this audit.

The original gh-formatted logs (~1.84MB) omit the terminal full-suite summary and remain preserved as partial evidence. Root-authorized raw job endpoint capture recovered complete logs without rerunning CI:

- Push job117390356808:3565380bytes, SHA256261275630e50057b161700190659a08e72b395097cde12ed38ed4e6d04f50a0c. Full-suite summary lines11282–11283; preflight11306–11307.
- PR job117390430291:3567600bytes, SHA256862d51dc21fb8606af7523ca6d8d338def952278340206b3f64eeb16b7a5e377. Full-suite summary lines11294–11295; preflight11318–11319.

Actual smoke records in both runs retain claimEligible=false, allNineTargets=true and packed payload39e78bd02dfb98de71e03e71faf2af37b56ff3d724f4d2ea930bf3cecbaf2bd9, identical to independently verified local packed bytes. CLI help/version/diagnostics, clean lifecycle, public Read admission/discovery and fail-closed unsupported profile checks complete. Full-suite child payload73455528487c841c80c4bc0243e49d8284b2441807b027fc03633d5bdfa5d11f is workspace identity, kept distinct. Restart reports development-local in these full-suite runs. Parseable fresh-depth/sequence reports show13/8 responses and successful cleanup assertions; minimum-current JSON is broken by GitHub masking in both logs, so its passing callback is verified without inventing a repaired complete report or independently recovered32-response array. Remote cleanup fields remain reported assertions; no remote process-absence claim is made.

Both historical Read evidence events retain constant674982be2123b9b7aae630bf79c3f774122900cd, runHeade5f7d0c43c151426350ceededdc77fe7e372db9b, evidenceCommitdbaf63a24219e1ea3f80394746853598b0d4464a and47definitions/46reachable/89rows. They are historical evidence, not a new candidate identity.

At audit, remote PR159 head is exactd719 and base is exact8d6671ed19b0c0406428ccc2d2161768d01e8b9b. Local checkout is clean, all701 source hashes match, and its ordinary parents remain d4f3f4be then8d6671. No newer main was adopted. No product edits, ready/publication operation, release or main merge occurred in this audit. The separate local95-case Type example pass is not attributed to CI; these outcomes do not qualify a complete write profile.

Companion JSON SHA256 `7da38b55de2877e2c8cb5ae043f2c1a50f70de50a7a4c6eff4ead1386df171ce`.
