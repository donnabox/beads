# Final-draft binding successor — closed v3

Author confirmed the final draft frozen. Actual source manifest /Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-command-minimum-final-draft/source-binding.json has SHA ff9c63417fbb348bf8ad9863483096ccc0329beb5c05b532333cf75cd0726a38; diff 891a3dd1d17d9f754dc81ef6223bdb61de4c330e15bc5ab7760559ca5c563106. All seven actual source hashes were checked. Relative to v2, only the expected progress-document hash/bytes change; three formatted tests and other three files are unchanged. Expected binding is 455372e6767b204095b3884c0c25cc3baa182ec9368b668d5e336c9816cd4dad.

The two v2 preparations and their original predecessors/evidence remain preserved. Exact v2-to-v3 changes are in successor.patch and changes.json: output/script/evidence paths move to v3; both gates pin the reanchored expected binding; closed examples name the genuine actual manifest/diff and corrected progress hash. No gate logic, command selection, dependencies, workload, runtime limit, oracle or cleanup behavior changes. Supervisors/control/package/adjudication scripts remain byte-identical.

The actual-manifest gate remains mandatory. Examples have released:false and no approved IDs; no execution-gate.json or run directory is supplied. Final commit/tree, acceptance, derivative snapshot, consumer lock and installed provenance release points remain held as before. Original successful and failed attempts are not transferred or relabelled as v3 validation.

Guarded external artifact writes, source-byte hashes and Python AST parsing only. No product edits, tests, build, install, services, package creation, publication or prepared-script execution. This is a binding preparation, not execution authorization.
