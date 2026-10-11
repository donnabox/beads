# Detached graph export publication slice

Supporting [issue12](https://github.com/donnabox/agent-coordination/issues/12)
under governing [tracker6](https://github.com/donnabox/agent-coordination/issues/6)
and full [export10](https://github.com/donnabox/agent-coordination/issues/10).
This package publishes caller-validated detached bytes; it decides no format,
serializer, CLI, source scope, permission default or BDP contract. The caller
must complete acquisition and checked storage cleanup before publication.

File writes use the existing atomicfile implementation, bounded cancellation
checks, and joined operation/returned-abort causes. Stdout/pipe writers remain
caller-owned and may retain partial bytes after failure. Successful final writes
or file finalization remain successful despite cancellation during or afterward.
The inherited filesystem boundary lacks directory fsync/power-loss proof,
non-Unix rename atomicity, and cleanup causes suppressed by atomicfile. No core
atomicfile behavior is changed. A native Low doc-qualification finding remains
pending actual Claude completion before source wording is updated.

[Canonical pure run](pure.log):9 roots/27 cases,0failures/0skips. Temporary
files exercise replacement, permission preservation on Unix, cancellation,
short writes and real rename failure. Fake close/abort injection proves the
wrapper's error handling, not actual fsync/remove faults. Windows permission
assertions are guarded while content/cleanup assertions remain. [Lint](lint.log)
passes native/windows/darwin profiles. BUILD source check passes; local Bazel
and Windows runtime were not run. [Manifest](manifest.txt), [prompt](prompt.txt)
and [native review](native.txt) identify the frozen packet; actual Claude is
running. No engine work occurs in this slice.

Contributor preflight found no exact publication topic match. Broader export
atomic search found external gastownhall/beads#6875 (rjgeng); its full three-file
diff concerns classic relation loader errors, no publication mechanics. No code
adopted, contributor PR modified, closed, superseded or merged. Its contribution
is preserved. This slice is draft progress and does not close full export10.
