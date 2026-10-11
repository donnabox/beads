# Publication council dispositions

Actual Claude [initial review](claude-initial.txt) completed exit0. Native review
and repaired-source confirmation are retained. Gemini excluded. The initial
frozen manifest remains immutable; the repaired packet has a separate manifest.

- F1 High: confirmed missing Windows permission disclosure. Document Unix mode
  bits versus Windows read-only/parent ACL semantics. No DACL/security guarantee
  or public permission default introduced; explicit ACL handling belongs to a
  later agreed caller contract.
- F2 Medium: rejected as demonstrated production defect. Factory/interface are
  private; PublishFile constructs only atomicfile.Writer. Arbitrary caller
  WriteStream writers are caller-owned. Panic recovery is speculative hardening.
- F3 Medium: rejected. Typed-nil creation failures return before dereference;
  testing.Fatal terminates its goroutine rather than returning a nil,nil writer.
  Invalid caller/factory implementations are outside the contract.
- F4 Medium: confirmed umask-sensitive test. Compare post-failure mode with the
  captured actual pre-publication mode; repaired full suite passes under077.
- F5 Medium: document private interface and inherited failed-Close/Abort limits.
  Fake close/abort and real rename failure claims remain explicitly distinct;
  no chmod/fsync/remove fault or successful physical-cleanup claim is made.
- F6 Medium: deferred inherited-regression test suggestion. atomicfile.Create
  currently explicitly uses the destination parent; this slice does not alter
  it. No new demonstrated cross-filesystem defect.
- F7 Medium: accepted empty path validation and before-IO test.
- F8 Medium: accepted caller-input immutability assertions on multi-chunk writes.
- F9 Low: rejected; context-first admission is intentional deterministic
  precedence, with no caller requiring the proposed contrary precedence.
- F10 Low: deferred error presentation to later CLI caller; errors.Is causes
  remain intact and this internal layer chooses no user-facing output contract.
- F11 Low: no new portability scope for unsupported Plan9/wasm targets. Required
  profiles cover native/Linux/Windows/Darwin; Windows runtime remains unrun.
- F12 Low: inherited read-only Windows replacement can fail and is documented.
  No promised ability to replace protected targets or ACL mutation is introduced.
- F13 Low: deferred speculative future-special-case coverage. Entry deadline
  and during-write cancellation use the same unchanged ctx.Err path.
- F14 Low: Bazel runtime unavailable, source wiring check passed and limitation
  documented. Visibility and timeout follow existing package conventions; no
  agreed serializer consumer exists to justify inventing a narrower target.

Native Low non-Unix atomicity wording is repaired. Inherited I1-I10 are retained
in raw review; directory fsync, power loss, suppressed cleanup causes, symlinks,
ACL/ownership and read-only limits are disclosed. Core atomicfile is unchanged.
The [actual repaired-source successor audit](claude-final.txt) completed exit0
and confirms no functional regression. Its two Low documentation nits are
repaired: name temporary-file descriptor Chmod before rename and document the
input sentinel precedence. Its Low symlink regression-test suggestion is
deferred to future publication/CLI qualification: inherited atomicfile is
unchanged and the documented rename behavior is statically confirmed. No
required runtime/source-preservation gate is omitted by that deferral. Neither this slice
nor its predecessor settles format/CLI or full exporter readiness.
