The conformance parser and exact HTTP serializer incorrectly classified `bdp-authorization-view` as a credential. Admit this exact protocol context name as authored input while continuing to reject it as a credential-handle field; retain credential protection and report redaction. Update the two installed test helpers to author the field directly.

Add a bounded, partial Transactional command-minimum scenario and coverage mapping, with actual-response checks and selected negative controls. Preserve inherited Read/Read+Update authority, historical scenarios, runtime limits, and incomplete-profile status. This does not change the server's command-minimum semantics or claim public profile qualification.

Validation on the exact committed bytes:
- 30 selected static cases passed; 601 skipped.
- Typecheck, fresh build, format check, lint, dependency boundaries, and the 89-row historical Read seal passed.
- Lint reported 765 warnings and 48 infos; format check reported one existing oversized-file warning. Neither applied fixes.
- Native and actual Claude source reviews completed; findings dispositioned.

Fresh installed validation and exact-head CI are pending. Planned installed validation separately covers two helper regressions and three matrix cohorts; no installed-pass or whole-profile claim is made here.
