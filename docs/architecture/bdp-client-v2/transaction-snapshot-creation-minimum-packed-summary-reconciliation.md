# Snapshot packed-summary reconciliation

The native review observed summary SHA256 `f3237923d2b8ff57db8adec2e85d03c7e9a1a9623420ab18b984247e874a78ed`. The final summary is SHA256 `367fc47fc59725749a59233620fcbfc83b6d8b98c8a7d6a152be87621e548fcd`.

Exactly three additions explain the difference:

- `exits.frozen-evidence-verify`: the final frozen-source verifier command, exit 0 (0.321 seconds).
- `remoteVerifiedRef`: `refs/heads/vickie-tx-snapshot-minimum`.
- `remoteVerifiedHead`: `587a7eeaa9663ab3d516a45375b3dbf0e9fedca9`.

Removing these additions reproduces the earlier bytes and exact earlier digest. Those reconstructed bytes are preserved as `packed-summary-before-final-verification.json`; this is explicitly a reconstruction verified by the independently recorded historical digest, not a claim that a separately saved contemporaneous copy existed.

A structural comparison confirms all remaining fields are identical, including source/tree, package and installed payload digests, test outcomes, derivative provenance, requests, and child cleanup. No result was rerun or relabeled. Frozen source and historical native report remain unchanged. This is metadata completion, not a substantive evidence discrepancy.
