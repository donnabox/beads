The conformance parser and exact HTTP serializer incorrectly classified `bdp-authorization-view` as a credential. Admit this exact protocol context name as authored input while continuing to reject it as a credential-handle field; retain credential protection and report redaction. Update the two installed test helpers to author the field directly.

Add a bounded, partial Transactional command-minimum scenario and coverage mapping, with actual-response checks and selected negative controls. Preserve inherited Read/Read+Update authority, historical scenarios, runtime limits, and incomplete-profile status. This does not change the server's command-minimum semantics or claim public profile qualification.

Validation on the exact committed bytes:
- 30 selected static cases passed; 601 skipped.
- Typecheck, fresh build, format check, lint, dependency boundaries, and the 89-row historical Read seal passed.
- Lint reported 765 warnings and 48 infos; format check reported one existing oversized-file warning. Neither applied fixes.
- Native and actual Claude source reviews completed; findings dispositioned.

Fresh external installation passed both helper regressions and all three matrix cohorts: 220 actual HTTP responses / 28 service launches, including the intended batch-conflict 409 and read-only-principal 403 controls. All 81 sampled validation PIDs are absent; no runtime directories remain. The new archives match 313 source members and the pinned eight-package dependency closure. [Installed evidence](https://github.com/donnabox/beads/blob/63b3ac50212cd0ad3ed21e290830de8c34bcee5a/docs/architecture/bdp-client-v2/transaction-command-minimum-evidence-validation-v3/root-installed-summary.json).

The conformance host runs from the workspace; the service runs from the fresh external installation. Evidence is partial, with raw equality asserted during execution and projected/fingerprinted reports retained. The saved-outcome audit is complete. Exact-head CI remains pending; this remains a draft with no whole-profile claim.
