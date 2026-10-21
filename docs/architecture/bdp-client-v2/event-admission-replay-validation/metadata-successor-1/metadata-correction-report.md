# Closed metadata correction

Corrected commands.sourceBindingSha256 to the actual unchanged Event source binding. Corrected consumer-name prose to describe the namespace-only rename. Rebound live packet paths and dependent hashes to separate sibling packets. Main helpers are byte-identical; required gate changes only the main packet path. Commands differ only in output/helper packet paths and the corrected attribution field; source, selection, schemas, workload, assertions, deadlines and execution semantics are unchanged.

The predecessor packets, native blocker and root release remain immutable. metadata-correction.patch is a genuine complete-file forward diff, with exact before/after hashes in rebind-map.json. Historical initial skeletons and original adaptation patches are retained as historical evidence, not current execution paths. All gates remain closed, approved stages empty and actual tools/build/consumer acceptance unresolved. No helper or product execution occurred.
