# Shared label writer prerequisite

This branch carries the complete contribution from [upstream PR5793](https://github.com/gastownhall/beads/pull/5793), head `638e87da1cc240fba896d6ea717734c0e8598a79`, into the graph candidate's pinned Issue writer. Rongjun GENG's fix makes actual label mutations advance the Issue's ordinary `updated_at` and row-lock token. Without it, label-only changes can be missed by timestamp filters and optimistic row guards.

The complete nine-file contribution is retained, including its domain repository parity fix, create-time exception and tests. Labels supplied during creation preserve the accepted create/import timestamp. Repeating an existing label addition or removing an absent label now leaves both state and audit events unchanged. That last point is an intentional ordinary Issue behavior change, reflected in the contributor's conformance test; it requires review along with the timestamp fix.

The upstream branch has intervening main merges. This fork uses the verified final delta instead of importing those unrelated ancestors. It does not replace, close or merge the contributor's upstream PR. Source provenance and attribution remain in the commit and fork review.

PR44 remains held for explicit review of ordinary Issue behavior. A real label
addition or removal remints `RowVersion`, so a guarded close with the token from
before that edit now refuses. Duplicate additions and absent removals preserve
the token. The comments and regression tests describe this provisional behavior;
they do not approve widening the ordinary guard contract. `RowVersion` still has
partial mutation coverage and is not the graph record revision.

The selective staging repair carries actual `RowsAffected` results through the
shared writer to its callers. A real label mutation stages the routed Issue,
label and event tables together. A no-op stages nothing and preserves dirty marks
from earlier real operations in the same transaction. This uses the existing
**table-level** staging model: a real label mutation can include other pending
Issue-table edits. It does not provide row-level commit isolation. For example,
a real `bd label add` on the same branch as a deferred import can publish that
import's pending Issue rows under the label command's commit. This newly widened
Issue-table staging needs explicit acceptance before promotion; passing tests
does not settle it. Conversely, the direct no-op path no longer stages unrelated
pending label rows either. Audit events remain Dolt-ignored and unversioned.

Hook behavior is a separate, unresolved limitation. The legacy store and
transaction hook decorators still fire or queue `on_update` after a successful
label no-op because their error-only interfaces carry no changed-result signal.
Thus no-op preservation here covers stored state, audit rows and commits; it
does not promise hook suppression. Fixing that interface is deferred for review,
not folded into this staging correction.

No graph label-editing command is enabled by this prerequisite. The next proposed slice is guarded complete-set replacement through the existing `IssuePatch.Labels` writer, with exact prior-record recall and filtering. Add/remove shorthand and label vocabulary retain their existing contributor ownership. The CLI shape and capability spelling remain provisional.

Tests cover the reused helpers, real native mutation behavior on embedded and ordinary shared-server Dolt, and the existing installed graph workflows. Committed-state regressions separately check the affected public update, transaction and direct-server paths through `AS OF HEAD`, including no-ops with unrelated dirty Issue rows. Direct native label tests do not demonstrate graph revision synchronization or graph label commands. Native History atomic commit-stamp binding remains unresolved: ordinary Issue timestamps are not such a guarantee.

Future integration with upstream PR6650 must reconcile version recording ownership. The current pinned graph adapter records its retained Issue once explicitly; the newer upstream `ExecuteUpdate` records internally. Calling both would duplicate recording. This prerequisite changes neither recorder.

The Linux qualification job explicitly enables the existing domain repository
suite on its owned disposable server using `BEADS_GRAPH_TEST_SERVER_PORT`.
Setup, cleanup and independent journal writer connections all use that port.
Without this opt-in the normal Docker fixture and local skip policy remain
unchanged. The job requires the domain suite to pass; a green run that silently
skips it does not qualify this contribution. Database provisioning remains
serialized across packages. This is test-fixture coverage, not a replacement for
installed CLI evidence or permission to change the ordinary Issue contracts.
