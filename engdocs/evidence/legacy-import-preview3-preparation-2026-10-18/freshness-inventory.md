# Freshness inventory at preparation source 89ca1e312

Read-only schema/native ownership sweep: 40 final application tables (31 native, seven graph, two migration cursors), 29 guarded planes and 11 permitted infrastructure tables. Migration scratch tables are renamed/dropped, not additional retained planes. Tests contain 32 occupied cases over the 29 planes plus pristine/config controls; engine matrix is authored UNRUN.

| Disposition | Tables/state | Reason |
|---|---|---|
| Must be empty, graph | graph_preview_catalog, graph_preview_payloads, graph_preview_links, graph_preview_versions, graph_preview_issue_versions | Allocations/deleted identities, payload, Links and retained history |
| Must be empty, native | issues, dependencies, labels, comments, events, provenance_events, child_counters, issue_snapshots, compaction_snapshots | Native ownership inventory in internal/storage/issueops/wisp_routing.go; retained/unmapped rows can collide or attach to new IDs |
| Must be empty, wisp | wisps, wisp_labels, wisp_events, wisp_comments, wisp_child_counters, wisp_dependencies | Ephemeral payload/relations/history cannot be silently discarded |
| Must be empty, remaining | issue_versions, bd_events_journal, leases, issue_counter, epoch_minted_addresses, interactions | History, coordination, consumed identities and unowned retained user/tool payloads |
| Seed-aware guard | bd_events_seq, store_epoch | Permit sequence(0,0) and absent/pristine epoch(1,1,NULL,NULL); refuse consumed sequence or changed epoch/markers |
| Prefix-aware guard | config | Refuse kv.memory. including empty values; permit ordinary infrastructure settings |
| Permitted infrastructure | graph_preview_scope, graph_preview_types | Required binding/descriptor configuration, independently validated |
| Permitted cursors | schema_migrations, ignored_schema_migrations | Required initialization bookkeeping, not resource history |
| Permitted configuration/cache | custom_statuses, custom_types, routes, federation_peers, repo_mtimes, metadata, local_metadata | Workflow/routing/cache and known bookkeeping/coordination configuration; blanket row refusal would reject initialized fresh targets |

Newly discovered interactions schema (0014) retains IDs, actor, issue/parent IDs, prompts/responses/tool outcomes and no seeded rows/FK; dormant SQL writer status does not make retained rows safe to ignore. Epoch registry (0071) deliberately retains old addresses after bump/deletion and has no Issue FK. store_epoch singleton is lazily initialized; row presence alone is normal. No implemented expected_revision_records/retention/alias table exists in this selected source: references in migration0071 are historical slice descriptions, not actual tables.

The inventory distinguishes known infrastructure uses, not arbitrary writes hidden in metadata. Missing/unreadable guarded tables fail closed with plane and cause; import performs no schema repair. Admission and full-state rollback tests must be run in real embedded/shared-server engines before qualification.
