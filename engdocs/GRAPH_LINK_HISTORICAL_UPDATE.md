# Informational Link historical Update adapter

This engineering crosswalk describes the checked preview adapter. It adds no public BDP profile or general snapshot overwrite. Native Dependencies use the Issue-domain writer, with properties and metadata authoritative in `dependencies`. Explicit current-head operation attribution is stored in `graph_preview_versions.native_attribution`; retained snapshot JSON never substitutes for native current payload authority.

| Link member | Historical composition | Explicit edit / authority |
|---|---|---|
| `properties` | Source editable object supplies the base | Existing merge, replacement or ordered patch; final current Type validation |
| `metadata` | Source common metadata supplies the base | Existing canonical metadata edits apply afterward |
| `id`, `type`, `source`, `target` | Preserve actual current identity/structure | Historical mode has no retype or endpoint-identity editor |
| `source` / `target` stored revision pins | Preserve current pins; last-live baseline on recovery | Only explicit endpoint pin flags change them, under current pin admission rules |
| Source ownership and owned membership | Current Type/owning Bead decide | Current source guard and owned-source writer; never copy historical membership |
| `revision`, `version`, `attribution` | Never copy source envelope | Effectful operation records new actor/time, selected source and actual current/D predecessor; live no-op preserves the whole envelope |

The writer selects current state or true terminal D inside one Type-validation transaction. For deletion it separately validates the terminal's `previousVersion` live snapshot as the protected baseline. A selected historical source cannot supply D or replace protected baseline facts. Source acquisition is exact and same-subject; old terminal rows remain terminal after recovery. Properties/metadata composition and source selection precede final equality evaluation; guards and admission still run for a no-op.

Recovery restores the existing allocation and its informational or native payload, appends a fresh retained Link version, and updates an owned source once through the existing completion writer. The terminal and every previous live version remain unchanged. Both endpoints must be currently admitted. Informational Links receive staged-graph Type and pin validation. Native Links revalidate their retained envelope and require exact current/retained agreement through owned-set capture; native integrity remains in the Issue-domain writer. Missing/corrupt baseline, invalid endpoints, authoring restrictions or any write failure roll back the complete transaction.

Historical acquisition reserves the distinct D, last-live baseline and selected-source roles together with the complete current working set before hydration. Explicit pin acquisition adds to the same budget before decoding each additional input. Current source and selected snapshot tokens remain independent from both Link and owning-source revision guards.

The shared terminal History carrier includes Type only for Links. Bead decoders explicitly reject that member. `currentState` describes only the actual current D; `terminal_versions` also retains old D rows after recovery. JSONL's existing refusal for unrepresentable terminal context remains in force rather than silently dropping D.

Native writes restore `dep_type` and common metadata, then apply explicit edits. Live kind and metadata changes update the native row in place after validation of the final native graph, preserving creation facts. The journal records `dep_add` as an upsert rather than remove/add reconstruction. Readiness is recomputed before the one source version. Native creation actor/time remain creation facts during live edits. Operation actor/time/context project consistently through current, retained, owned, and export surfaces. Deleted recovery refuses an occupied pair rather than adopting its new canonical Link ID.

The attribution column requires fresh schema 9. Schema-8 workspaces are explicitly refused; this implementation performs no in-place migration. NULL attribution is the stored initial native creation projection, with exact current/retained agreement still required. Only explicit creation callers may project the interval before the first retained row exists; ordinary current readers refuse a missing head mapping. Malformed or oversized explicit context refuses.

Effectful native writes require a native actor; a complete no-op may return its unchanged envelope without authoring. Admitted graph writes serialize on the graph writer token. This adapter does not provide the plain native store's post-commit blocked recheck across mixed writers; direct native writes bypassing graph mappings remain outside controlled graph authority. The wider cross-writer recheck gap is pre-existing in graph close/delete/unlink as well.
