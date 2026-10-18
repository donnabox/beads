# Current-core test compatibility follow-up

Native preliminary source review of e923b31528bfb1887ed08baa751f9217b22621a3 found two concrete test assumptions incompatible with Janet-selected c67855b4f7c1e4fb24409eb22d142a52d19ab92f. Both are repaired before fresh final source council; production export code is unchanged.

- Cleanup connector fixture: install the complete current Type catalog, report its actual cardinality, and use that cardinality in the successful acquisition oracle. The selected core requires all six current descriptors. Preserve acquisition-stage, rollback/error, transaction/connection and zero-state controls; do not weaken the core guard to accept four-Type stores.
- Engine-backed lifecycle oracle: `IssueComments` exposes a graph path as IssueID, whereas the export comment feed preserves the persisted native Issue ID. Normalize only the copied expected comments to the independently obtained native Issue ID. Preserve exact comparisons of all remaining comment fields, cardinality, snapshot state, read-only persistence and reopen behavior.

Static formatting and changed-source whitespace checks only. Builds, tests, generators (including check mode), installed workflows and engine/database activity remain UNRUN and Janet-owned. These repairs are source inspection, not a claim that tests pass. Historical 68-case evidence remains source-bound to its old commits. Fresh council must inspect this repaired source.

Original 70-path transfer now has 66 byte-identical paths, four adapted original paths (two BUILD files and two test files), plus two newly touched selected-core BUILD files. This follow-up adds two evidence paths. No import ancestry, public serializer/CLI activation, format adoption, integration/release/default change or import2b mutation.
