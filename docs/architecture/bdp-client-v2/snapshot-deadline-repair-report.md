# Snapshot aggregate deadline repair

Frozen clean source `ffe188a3405881e9488b4e7c997f557d39829fc5`, direct parent `6413cbc2842d898e523869d56c8286e87f358804`.
Checkout `/Users/dbox/repos/wt/vickie/vickie-snapshot-deadline-repair`, branch `codex/vickie-snapshot-deadline-repair`. Only the installed Snapshot test and its dedicated observation document changed. No runtime, settings, store, serving, public, shared launcher/binding/runner, H1b or SSE changes. No outward writes.

Remote read guard before work confirmed `origin/codex/vickie-snapshot-serving-final` exactly6413cbc; maincaf5ca75. Original6413 source and qualification remain immutable.

## Defect and bounded repair

Native reviewer demonstrated that two sequential fixture-local45s budgets inside one60s test could permit the outer timeout first. Exact original probe and logs remain in `snapshot-serving-final-native/{native-probes.patch,probes.log,probes.json}`; its controlled31s delays produced an active child at60s and eventual cleanup at62.89s.

The calibration and subsequent observation now share one45s controller through a local `withDeadline` wrapper. Both fixtures consume that same budget for startup, HTTP exchanges, reopen, final deadline checks and outcome reporting; no per-fixture timer resets it. An already-aborted signal is refused before creating a directory or launching a child. Single-fixture callers retain their original45s budget. Per-request5s, launcher cleanup10s and outer60s bounds are unchanged. Cleanup preserves original/cleanup causes and the source/provenance ledger behavior from6413.

Two scaled deterministic lifecycle controls use20ms of virtual time only in local helper tests: expiry between fixtures prevents the second fixture callback/startup path; two12ms observations share one budget, clean both observations and produce [passed,failed] rather than late success. Timers are cleared. These are controls over the local lifecycle, not injected server clocks or emulated target responses.

The shared signal is cooperative. It bounds the authored callbacks' actual signal-aware startup/HTTP/reopen operations; it does not forcibly interrupt an arbitrary user callback that ignores AbortSignal (such as an uncooperative31s timer). No Promise.race abandons pending work. A prior fixture's passed ledger describes that fixture only; aggregate test success still requires the whole test to finish. Failed-route attempt detail remains root-deferred; the ledger still honestly records completed responses plus failed outcome.

## Exact verification

- Build and typecheck pass; boundaries, format, targeted lint and git diff check pass.
- Focused8/8 pass (`focused.{log,json}`).
- Fresh clean committed rebuild, protocol/server tarballs and actual isolated installation all exit0 (`pack-status.json`, `pack.py`). Fresh packed test8/8 pass (`packed-snapshot.{log,json}`): two actual service tests across3fixtures, four inherited lifecycle controls and two new aggregate controls.
- Three packed service records report passed, aborted=false, workspaceDirty=false and exact declared ffe188a source;127 completed HTTP exchanges. Full identities in `packed-bindings.json`.
- Historical Read verifier89 passes; no fresh Read run, evidence edit or relabeling.
- Full8k suite and broad smokes intentionally not repeated for this two-file test/doc delta;6413 exact full8596/155+skip remains separate prior evidence and root is validating its integration.
- Frozen source stays clean; no owned child processes remain after final process inspection. Test stores cleaned; external logs/installed payload retained for reproducibility.

## Initial work and limits

An edit script initially stopped on an overly broad textual-match count before writing any source; corrected targeted replacement used. No test failure, timeout, deadline widening or production change occurred in this successor. Existing Q1/profile/expiry/heap/disconnect limits remain unchanged. Independent native closure and root publication remain separate.

Packed payload `78a80950acd6c1981d6829fc4d59c05eaafc65e0d68cf20d3d84f61f377acb99`; measured realpath entry `bcf153170f953f9e28c6c89f1153be907f64df289e64f3ad929cad4a4a54c47c`.
