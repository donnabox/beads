# Independent graph export milestone

Status: discovery in progress; interchange format and CLI contract are proposals
until coordinated and reviewed. Governing plan: donnabox/agent-coordination#6.
Supporting export issue: donnabox/agent-coordination#10.

Donna authorized export to begin in parallel with legacy import on October 10,
2026. Legacy import remains independently eligible for Preview 3. Export has its
own PR, tests, council and release decision, and may land in Preview 4. Neither
export nor graph-format round trips are prerequisites for releasing legacy
import. Graph-format import starts after the corresponding export exists.

## Isolation

The export branch starts from qualified integration commit
`76fde9c9cb0865d430bf12d76fc473490f51128a`, without PR86 ancestry. Newer staged
integration commits are not assumed qualified. Janet selects the integration
pin; Donna decides the cutline. No mutation of Janet's staging, core worktree,
runner or release candidate is authorized. Engine-backed qualification uses
fresh disposable stores and the coordinated engine slot.

## Inventory and contract work

- Inventory classic `cmd/bd/export.go` and its filters, flags, read transaction,
  file publication and stdout behavior. Existing `cmd/bd/graph_export.go` renders
  DOT/Mermaid diagrams; it is not a graph interchange format.
- Inventory graph snapshot, retained Issue fields, metadata, comments, Memory
  keys/values, Links, Types and configuration. Identify which can be captured in
  a single consistent read through the existing storage boundary.
- Propose a versioned format with exact identity/property preservation and
  deterministic serialization. Choose explicit inclusion, exclusion or refusal
  for every field family. Document comment identity, unsupported relationships,
  leases and history; current state alone cannot claim to be a history backup.
- Route format versioning, Type/config ownership, bounds and CLI flag semantics
  through the agent bus to Janet and the relevant BDP contract owner. An
  interchange envelope must not silently redefine the BDP wire contract.
- Review the format/implementation plan with the requested council before
  treating it as settled. Gemini remains excluded by Donna's instruction.

## Export acceptance, independent of import

1. Export the documented supported subset consistently through a read-only
   storage operation. Refuse unsupported states or disclose documented omissions;
   never emit a deceptively complete archive or silently truncate a bound.
   Use a controlled concurrent-writer test to prove records, Links, comments and
   included configuration all belong to one snapshot, rather than individually
   valid results from different transactions.
2. Preserve exact IDs, dependency endpoints, metadata, Memory key/value content
   and admitted comments. Test deterministic output and empty, mixed and bounded
   graphs; disclose any identity remapping explicitly.
3. Validate and serialize before publication. File output uses atomic publication
   so validation/read/serialization failures preserve an existing destination.
   Test writer errors and cancellation; document that a stdout I/O failure can
   leave bytes already delivered to the pipe.
   Complete the read transaction and connection cleanup before publishing any
   output. Inject rollback/close failures and preserve the existing destination;
   an error after replacing the file must not look like a clean export failure.
4. Prove source state unchanged after success and failure, including comments,
   retained state and configuration. Qualify the canonical installed binary in
   embedded and ordinary shared-server modes with no required skips.
5. Complete documentation, manifests, required checks and the requested council;
   preserve exact source/build/engine identities and owned cleanup evidence.

Graph re-import and semantic round trips are a subsequent acceptance milestone
after export establishes the format. They do not gate the legacy importer or
the independently qualified exporter. Export fixtures can be decoded and
compared directly with the source snapshot before graph import exists.

## Contributor preflight

The upstream graph-export search returned broad matches. Reviewed relevant
preflights and diffs: PR7363 changes terminal title rendering, PR7408 changes
proxied ready/graph/search row caps, and PR6394 restores test-global JSON state.
These are not interchange exporters. No contributor branch is being replaced,
merged or closed. Preserve their behavior/tests if later touching shared areas;
re-run preflight when the concrete implementation scope is known.

## First implementation slices

Begin with a pure format serializer and disposable synthetic fixtures after the
contract review. Add a consistent storage reader only if the inventory proves
the existing boundary insufficient, keeping it separate from CLI wiring. Add
CLI routing and atomic publication next, followed by installed qualification.
Keep every slice free of PR86 implementation dependencies.

## Initial storage findings

`internal/storage/graphstore/collection_read.go` provides `CurrentSnapshot` and
`currentSnapshotInTx`: current records and installed persisted Types are read
in one transaction with 1,000 live-resource and 16 MiB acquisition guards.
Comments, Issue prefix/configuration, deleted allocations and past versions
are outside that snapshot. Calling `IssueComments` or configuration accessors
afterward does not extend its consistency guarantee. A candidate export reader
should collect the approved current-state components in one transaction through
the storage boundary and finish cleanup before returning a publishable result.
Its aggregate bounds must include comments/configuration/Types as applicable;
the existing record budget alone does not prove whole-export boundedness.

| Data | Source seam | Export implication |
| --- | --- | --- |
| Persisted Types | `types_read.go`, `informational_types.go` | Preserve actual installed descriptors/fingerprints; do not reconstruct the binary's known Type set. |
| Issue identity | `types.go`, `issues.go` | Preserve both canonical graph identity and native `properties.id`; graph path need not be `beads/` plus native ID. |
| Comments | `issue_comments.go`, `internal/storage/issueops/comments.go` | Acquire current comment IDs, authors, text and timestamps inside the export transaction. |
| History | `version_history.go`, `history_exact.go` | Per-subject APIs are separately transactional and do not establish a complete consistent history archive. |
| Public vocabulary | `cmd/bd/graph_preview_record_json.go` | Public shaping removes the internal `version` alias and projects attribution `status` to `basis`; choose export vocabulary deliberately without altering caller properties/metadata. |
| Publication | `internal/atomicfile/atomicfile.go` | Reuse temp-file/fsync/rename after successful reads, cleanup and serialization. |
| Opening | `backend.go`, `OpenExisting` | Existing-store open performs no DDL/migration; do not initialize a store merely to export it. |

Owned-Link snapshots and authoritative top-level Links need exact consistency
checks, endpoint closure and Type closure. Output bounds must account for JSON
escaping and duplicated owned snapshots. Live lease fields must be explicitly
preserved as observation, omitted with disclosure, or refused; source Scope and
revision tokens are provenance and do not authorize future import to recreate
the source authority. Memory title/body must remain exact; graph Memories need
not possess a legacy key. Decide destination file permissions explicitly rather
than inherit ordinary export's `0644` without review.

## Questions routed to contract owners

1. Single versioned JSON document or distinct graph JSONL vocabulary; graph-mode
   default versus an explicit format flag; separate capability/profile name?
2. Public revision/attribution vocabulary, source Scope/generation and approved
   provenance; actual installed Types as observations for later remapping?
3. Memories included by default for a complete graph; if filtered, Link closure
   and explicit omissions; approved config subset, chiefly `issue_prefix`?
4. Live leases and deleted/history state: explicit current-only omissions versus
   refusal; no full-history claim from per-resource readers?
5. Aggregate resource/comment/encoded-byte ceilings, timeout/error contract and
   output file permissions?
