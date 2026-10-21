# Closed targeted evidence successor

Reuses the reviewed source-validation-t11 supervisor. Four ordered stages only: typecheck, formatter CHECK on exactly two files, T1 alone, T6/T9/T10/T11 together. No build, other six callbacks, control or installed execution. No live execution-gate.json/run exists. Root must write acceptance/release explicitly.

Changes to supervisor: new exact command/prior-evidence hashes and four-stage order; allow pinned Biome for format-check; two changed-test exception in prior source bridge; verify all other positive copy-manifest members (including468emitted) before AND after each stage. This compares positiveSha256, not mutated copySha256. All process-group115TERM/120KILL, immediate failure/no retry, immutable logs, selected JSON names/status, TMPDIR/PID/cleanup and session checks remain unchanged. Actual old failed evidence is pinned without asserting its failed owner stage passed.

Source binding derives evidence-repair by adding tree=baseTree only. Existing source/head/fifteenpath/diff/scope gates remain. At preparation1141 other product/build members including468emitted matched old positive graph; two revised tests are the only exclusions. Source itself is fully checked against new manifest. Preparatory Python AST parse of runner only; no product execution.

Expected targeted5cases18HTTP6successfulopens5dirs1listener0children; complete positive successor remains11cases22HTTP15opens11dirs+5failedcopydirs2listeners. T2 is not selected even though its local helper supports HEAD now. C1 previous7request expected-failure remains historical; exact unchanged prefix/gate proof is pinned in repair measurements. Extra HEAD lies after that failure. No rerun implicit.
