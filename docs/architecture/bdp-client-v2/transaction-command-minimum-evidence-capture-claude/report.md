# Actual Claude focused capture-correction review

Actual exit 0 in 462.714521124959 seconds, within the authorized 900-second limit. No timeout, forced termination, retry, tests, builds, installs, services or source edits. This was a supplied-context review of the correction only, not a full v4 rereview or implementation approval.

## Binding

- CLI: /opt/homebrew/bin/claude --safe-mode -p --tools '' --no-session-persistence. No model override.
- Source: /Users/dbox/repos/wt/vickie/vickie-command-minimum-evidence, HEAD453b6dab386746bb03aac2c0a293572b8a71c969. The two existing draft JSON modifications were retained. All tracked source bytes, status/diff, target files and prompt stayed unchanged during review.
- Proposal: transaction-command-minimum-evidence-draft/capture-correction-proposal.md; exact digest in binding.json.
- Frozen addendum SHA25682f415e5d95e727bc919dd8a7c7d09ce73f19fb983fcdf9d3c0d28807cab637f.
- Native qualifier SHA2568074c328c51ec1555b82aaf7df167fdbf9868146eb34f9a1013395e40a47df1f.
- Prompt49,478 bytes, SHA2568bcd22c883cfe6edebbb3d5169293a3fcd39ae0201d7c06ffdb746ac34cdea6d.
- Actual stdout SHA256fdba275aa0a4aa5fa09e16a00c4d425a1b8efd4f37b1f7453cbc02d17bce2c25. Stderr empty.
- Before/after snapshot SHA256781f73b964fee49d108f5a5d45fe58e66eae9e8686b01d1879a09f608e15bd34.

## Returned review

Claude judged the core correction sound. It confirmed URL-only raw captures, the distinction between /id original-body equality and /next first-GET initialization, hook-before-manifest ordering, parsed-JSON equality, and the dynamic receipt transport bypass of the outer app hook. It returned these twelve findings; their original severity and qualifications remain in stdout.txt:

1. High: require activation, per-mode performed-comparison counts, and postrun no transport/harness errors so a supplemental hook cannot silently miss the row or lose an assertion failure.
2. Medium: scope removal to raw captures; programmatic JSON captures exist. App-only use may remain deliberate rather than claiming such captures cannot carry A facts.
3. Medium: distinguish internally enforced route/session/credential properties from report pins whose projection must be confirmed.
4. Medium: establish available in-hook schema validation, or describe a feasible pre-baseline semantic gate plus postrun schema conjunction.
5. Medium: label actual-status/request/body receipt assertions as runner-internal; derive header absence from known empty authored headers and the assertion semantics, never filtered reports.
6. Medium: make the replay ledger drive comparisons and require performed counts; a declaration-only guard does not establish execution.
7. Low: avoid relying on one thrown receipt failure message without checking the callback catch/projection path.
8. Low: distinguish logical configuration identity from same generation/listener identity.
9. Low: state receipt-authority eligibility; do not assume sequence /id grants authority or retain an unused B root.
10. Low: do not claim a hook method field without verifying its real interface.
11. Low: clarify existing static registration amendment versus adding a registration.
12. Low: verify position/body expectations against the actual fixture; identify fixture dependencies.

Several findings explicitly depend on absent source context, including callback catch behavior, report-field projection/redactUrl path handling, validator availability, empty authored-header semantics, and credential/control behavior. Those are questions for actual source inspection, not demonstrated production defects. This operator report does not accept Claude's speculative premises or propose expanding the action count. Root owns dispositions.

The complete stdout is authoritative. Claude's optional programmatic-capture alternative is not adopted here; the approved direction remains a bounded existing-interface correction unless root changes it.

## Cleanup and proof limits

Runner reaped PID/process group92228. Final group and sampled-PID presence sets are empty; sampled/final owned temporary-entry lists are empty. The empty evidence-owned tmp directory remains. One-second sampling can miss short-lived or detached descendants; this is not exhaustive detached-process proof.

Claude had no tools and received exact supplied excerpts plus the three complete correction documents. It did not inspect broader source or execute any proposed proof. No actual implementation or qualification credit follows from this review. No second attempt was launched.
