# Proposed native Issue historical Update crosswalk

Implementation subject: the existing graph `UpdateIssueRequest` and `issueops.ExecuteUpdate` setter path, based on published Type and endpoint work plus the Memory historical-update stack. This is a bounded native adapter engineering proposal for review. It does not change shipping `bd restore`, native lifecycle semantics, public BDP profiles, or fields admitted by ordinary graph Update.

The historical source must be an exact retained live native Issue of the same graph subject. The actual current graph revision is checked independently before source acquisition. A deleted native Issue remains explicitly unsupported until its terminal/lifecycle successor is implemented; Memory recovery is implemented separately and remains under review, not integrated.

| Source native member | Historical composition | Explicit edit / ordinary authority |
|---|---|---|
| `title`, `description`, `design`, `acceptance_criteria` | Copy the historical editable value, including ordinary empty values | Existing string setters and native validators |
| `priority` | Copy the historical value, including P0 | Existing native priority validator |
| `estimated_minutes` | Copy the exact nullable value | Existing set/clear representation and native/storage bounds |
| `assignee` | Copy historical value, including unassignment | Existing active-assignment transfer fence; a graph revision guard grants no override |
| `notes` | Historical value supplies the notes base | Existing overwrite fence remains; append applies after that source base through native composition, preserving ordinary append intent when the source base equals current notes |
| `external_ref`, `spec_id`, `due_at` | Copy the exact supported nullable/scalar values | Existing column, null/empty and whole-second representation rules |
| top-level common `metadata` | Start with source metadata, then apply explicit common-metadata edits | Existing canonical I-JSON admission and metadata writer |
| `status`, `issue_type`, `owner`, scheduling/gate/molecule/message/event members outside the graph Update setter list | No automatic copying. A changed non-admitted editable member refuses historical mode with an explicit native-domain-mapping diagnostic | Separate ordinary native domain operations keep authority; no generic snapshot overwrite |
| Labels, parent, Dependencies, ID, graph Type, definition/endpoints and owned Links | Retain actual current identity/structure | Historical mode does not retype or recreate Links |
| native creation facts (`created_at`, `created_by`, source tracing/system, storage class) | Retain current initialization facts | No reinitialization from the older content source |
| native operation/derived facts (`updated_at`, started/closed context, node-local lease liveness, readiness and compaction observations) | Remain under the ordinary current native writer/provider | Never copied blindly from retained content; native write actor/time describe the new operation |
| future/unclassified serialized native member | Refuse when source and current differ | Extend the reviewed adapter table before claiming faithful support |

Ordered native property changes evaluate the existing closed add/replace/remove operations against the selected historical editable setter projection inside the same transaction. Replacement retains the ordinary required title/priority rule and clears omitted optional setter fields. No lifecycle or structural member is added to that projection.

Effectful live updates mint one ordinary native retained revision and one graph mapping. Live equality returns the entire existing record without a revision, event or provenance-only write. Source admission, guards, ordinary domain validation and acquisition bounds still run for a no-op.

Proposed persistence: nullable `revision_context` on `graph_preview_issue_versions`, storing only optional Inception and historical source/predecessor context. Actor/status/time remain authoritative in the native `issue_versions` row; they are not duplicated in the context column or the native Issue payload. Ordinary minting preserves current Inception presence/value and drops the previous operation's historical context; no-op preserves the existing context. Exact/current readers merge the validated optional context with native actor/time. This stacks on endpoint PR133's fresh-only schema8 without a second bump, migration, or change to Type generation.

Required qualification before native integration: exact-source current/historical equality, nullable fields, metadata/ordered edits, append and overwrite/assignment fences, non-admitted-member refusal, stale/foreign/unknown source, context retention and ordinary mint reset, cumulative acquisition budget, rollback, races/replay/unknown responses, installed embedded/server journeys, independent source review and actual hosted CI. The authored live native slice now wires persistent context and the CLI source-only/`--properties`/ordered-property-change/metadata route; deleted native recovery still refuses explicitly. Qualification and review receipts belong to its PR; publication does not itself mean integration or release qualification.
