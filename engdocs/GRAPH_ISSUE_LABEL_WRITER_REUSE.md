# Shared label writer prerequisite

This branch carries the complete contribution from [upstream PR5793](https://github.com/gastownhall/beads/pull/5793), head `638e87da1cc240fba896d6ea717734c0e8598a79`, into the graph candidate's pinned Issue writer. Rongjun GENG's fix makes actual label mutations advance the Issue's ordinary `updated_at` and row-lock token. Without it, label-only changes can be missed by timestamp filters and optimistic row guards.

The complete nine-file contribution is retained, including its domain repository parity fix, create-time exception and tests. Labels supplied during creation preserve the accepted create/import timestamp. Repeating an existing label addition or removing an absent label now leaves both state and audit events unchanged. That last point is an intentional ordinary Issue behavior change, reflected in the contributor's conformance test; it requires review along with the timestamp fix.

The upstream branch has intervening main merges. This fork uses the verified final delta instead of importing those unrelated ancestors. It does not replace, close or merge the contributor's upstream PR. Source provenance and attribution remain in the commit and fork review.

No graph label-editing command is enabled by this prerequisite. The next proposed slice is guarded complete-set replacement through the existing `IssuePatch.Labels` writer, with exact prior-record recall and filtering. Add/remove shorthand and label vocabulary retain their existing contributor ownership. The CLI shape and capability spelling remain provisional.

Tests cover the reused helpers, real native mutation behavior on embedded and ordinary shared-server Dolt, and the existing installed graph workflows. Direct native label tests do not demonstrate graph revision synchronization or graph label commands. Native History atomic commit-stamp binding remains unresolved: ordinary Issue timestamps are not such a guarantee.

Future integration with upstream PR6650 must reconcile version recording ownership. The current pinned graph adapter records its retained Issue once explicitly; the newer upstream `ExecuteUpdate` records internally. Calling both would duplicate recording. This prerequisite changes neither recorder.
