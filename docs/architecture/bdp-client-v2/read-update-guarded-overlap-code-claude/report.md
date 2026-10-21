# Actual Claude RU code-review outcome

One authorized external supplied-context review completed with CLI exit0 after759.966s; no timeout, forced termination or retry. Command was /opt/homebrew/bin/claude --safe-mode -p --tools '' --no-session-persistence, no model override. claude-review skill applied. No product tests/imports/build/services/source edits were performed.

Actual stdout SHA256 cdf542a216b67795420bd4df66921488293b7cee64ad5586c33129061d47b9d8; stderr is empty (e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855). Prompt109609B SHAc22e90b72ed534a80489801cd1d3a38f39f88fae482a254802c4e4a9ff8e7777. Full stdout is retained unchanged, not replaced by this summary.

Claude returned16 numbered findings:2High,5Medium,9Low. High items concern inherited SIGTERM exit assumptions (explicitly conditional on unsupplied process entry) and inherited BDP override enforcement. Medium items concern startup-error classification, source binding/context gaps, header/media source evidence, Resource projection evidence and request teardown. Low items cover request-shape hardening, unreachable overflow precision, tautological checks, overlap proof limits, optional explicit pair statuses, report/privacy scope, required evidence output, source/build attribution and native citation drift. These are Claude's findings, not yet source-adjudicated defects. It explicitly reports no constructed broken-guard false-pass path and verifies10positive/6negative requests, body arithmetic and both postimage schemas before the named control oracle.

All15 bound RU/reference/context files and all8 frozen external target/plan/native artifacts plus prompt are byte-identical before/after. Source-before/after SHA9c4fc001f17955261f400d8208287f41445892a7ccf712d1e84ef614c304e6dc. Separate checkout chronology changed from3968081f8231566908e69045424b067c3c36dc40 with the known unrelated matrix test repair dirty to1a37f7209e3ff6f409f43491d670c881259406ba clean; no same-HEAD claim. The preview remains external and unmodified. Actual receipt merge remains a separate root gate.

Owned Claude PID/group24219 reaped; wrapper final group/sample-PID sets empty, and a subsequent independent ps lookup returned1/no row. Temporary inventory empty. One-second group sampling can miss very short or detached descendants; no stronger claim is made. cleanup SHA9973e42f2108d886c3444756a400399067005e0ea97ad81d9af30c87a019e0cb.

Root requested subsequent read-only source dispositions of all16 findings. That audit is separate from this actual review outcome and authorizes no preview edits or rerun.
