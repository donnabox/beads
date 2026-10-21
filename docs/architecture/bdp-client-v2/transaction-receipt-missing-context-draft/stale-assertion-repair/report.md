# Receipt stale assertion repair and bounded validation

The single stale assertion now reads get('bdp-scope-position').toBe('pos-1'). No production, fixture, runtime limit, operation input, timeout or other assertion changed. Original before.ts and original-source-binding.json preserve the prior eight-file draft; correction.patch is exactly one removed/one added line.

Successor binding: e96f012c8502285454eb3867f8b7b35c82c483ae3ceb3f7fc21eaa31127d0cf3.
Full eight-file patch: d0afee38fdab6f78bbebe6b131066204a58580c7bf0bbfce3cead43e36a73809.
HTTP test SHA256: 37084d256def9c1f1cec33fc2c6a0d6e793dba3d29699a57bf105bda0cde68db.

Commands were recorded before execution in commands.json and run exactly once each under the existing bounded supervisor adapted to three stages, with source bindings, exclusive attempt markers, BDP overrides cleared, process-group and temporary-directory sampling. Parent's explicit authorization is recorded separately. No original14-case rerun, rebuild, installed run, commit or push occurred.

## Actual results

- Typecheck: exit0/supervisor0, 2.655s.
- Format check: exit0/supervisor0, 1.774s; one file checked, no fixes.
- Exact callback “uses actual epoch/view decisions and an exhausted wait before conditional or timeout responses”: one passed,76 skipped; callback72.01266700000002ms, supervisor1.673s. Genuine verbose log and JSON retained.

The callback's static workload is14HTTP,one owner/listener/store,zero installed product children. This is source-derived arithmetic, not an independently measured network count. No additional request instrumentation was introduced. The test's ordinary hard assertions and existing fixture cleanup completed; process and temporary-directory observations are separate evidence.

## Preservation and cleanup

All 1105 other source/emitted manifest members, including all 440 emitted files, still match their pre-repair hashes. The seven other owned draft paths are unchanged. This is byte preservation of previously built outputs, not a new build or new installed attribution. Original control/positive artifacts remain unchanged and attributed to the original draft.

All 14 sampled process IDs and three process groups were independently absent in a fresh process snapshot. No bdp-tx-http- runtime directory remained. Other retained cache entries are listed verbatim in cleanup.json and not deleted or misclassified. No timeout, forced cleanup or retry occurred. Sampling can miss very short-lived/detached processes; this is not a claim of complete OS tracing.

Full exact-head CI remains required. This repair closes the identified local stale assertion; it does not claim whole-suite success or public qualification.
