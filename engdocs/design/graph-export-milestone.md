# Independent graph export milestone

Status: internal reader and detached publication implemented; current-core source
review and runtime qualification pending.
Interchange format and CLI contract are proposals
until coordinated and reviewed. Governing plan: donnabox/agent-coordination#6.
Supporting export issue: donnabox/agent-coordination#10.

Donna authorized export to begin in parallel with legacy import on October 10,
2026. Legacy import remains independently eligible for Preview 3. Export has its
own PR, tests, council and release decision, and may land in Preview 4. Neither
export nor graph-format round trips are prerequisites for releasing legacy
import. Graph-format import starts after the corresponding export exists.

## Isolation

The original export branch started from qualified integration commit
`76fde9c9cb0865d430bf12d76fc473490f51128a`, without PR86 ancestry. Janet selected
`c67855b4f7c1e4fb24409eb22d142a52d19ab92f` on October 18 for a separate
source-only reader/publication adaptation. Other integration commits are not
assumed selected or qualified for this lane. Janet selects the integration
pin; Donna decides the cutline. No mutation of Janet's staging, core worktree,
runner or release candidate is authorized. Engine-backed qualification uses
fresh disposable stores and the coordinated engine slot.

## Inventory and contract work

- Inventory classic `cmd/bd/export.go` and its filters, flags, read transaction,
  file publication and stdout behavior. Existing `cmd/bd/graph_export.go` renders
  DOT/HTML diagrams; it is not a graph interchange format.
- Inventory graph snapshot, retained Issue fields, metadata, comments, graph
  Memory title/body, Links, Types and configuration. Identify which can be captured in
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
2. Preserve exact canonical Scope-qualified graph IDs and native Issue IDs,
   dependency endpoints, metadata, graph Memory title/body content
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

Acquire current records/Types through `currentSnapshotInTx`; a dedicated same-
transaction export reader is required for comments and exact persisted prefix.
Implement this internal acquisition slice separately from file-format and CLI
decisions. Add a pure serializer and detached synthetic fixtures after the
format review, then CLI routing/atomic publication and installed qualification.
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
Its aggregate bounds include comments/configuration/Types and deleted Issue final
heads decoded during existing mapping validation. The existing record budget
alone does not prove whole-export boundedness. This accounts for persisted
acquisition bytes, not exact Go heap use or encoded output size.

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

## Council clarifications and proposed field dispositions

The native and external council reviews are retained under
`engdocs/evidence/graph-export-plan/`. The discovery/acceptance plan is repaired;
interchange vocabulary and CLI decisions remain proposals on the agent bus.
The isolated internal acquisition reader does not establish a public format.

Other complete-read projections are `graphstore.ListBeads` in `bead_list.go`
(and CLI `graphProjectCompleteRecords`) and `graphread.Inventory` in
`internal/httpapi/graphread/inventory.go`. ListBeads answers Bead listing policy,
not a full Bead-and-Link interchange archive. Inventory consumes CurrentSnapshot
and projects the pinned BDP Read shape; its documentation prohibits serializing
its internal structure as a BDP response. Neither replaces the chosen complete
current snapshot reader or supplies transactional comments/configuration.

| Member/family | Proposed export disposition; contract coordination still open |
| --- | --- |
| Canonical graph ID, Type, revision | Preserve absolute Scope-qualified observations; revisions do not authorize destination revision minting. |
| Internal `version` alias | Deliberately distinguish alias from public `revision`; do not silently substitute a BDP projection. |
| Memory properties | Preserve path, title, body exactly; include common metadata, attribution and owned state. |
| Native Issue properties | Preserve exact native `properties.id`, workflow fields and dates as represented; lease fields need the explicit runtime disposition below. |
| Link source/target/properties | Preserve exact endpoints and properties, validate endpoint/Type closure. |
| Common metadata | Preserve strict canonical JSON without lossy typed decoding or changing caller keys such as `version`/`attribution`. |
| Attribution | Preserve actor, status/basis choice and observed recordedAt explicitly; attribution is distinct from caller metadata. |
| Owned Links | Preserve complete owned snapshots and top-level Links, with identical authoritative values/revisions. |
| Installed Types | Preserve persisted canonical descriptors and fingerprints, including actual installation generation. |
| Comments | Preserve native IDs, native ownership, canonical Issue subject, author, text and supplied timestamp from the same transaction. |
| Source binding | Include approved Scope/authority/schema-generation provenance; omit canonical filesystem WorkspaceID and connection settings. Destination authority remapping is future importer work. |
| `config.issue_prefix` | Acquire the exact persisted value, without a helper trimming a trailing hyphen; include approved source configuration rather than infer it from IDs. |
| Classic `kv.memory.*` configuration and wisp live planes | Not graph Memory/Issue inventory; initial reader refuses present rows by plane name. Inclusion requires a distinct approved acquisition contract. |
| Live leases | Acquire faithfully as current observation; final format must preserve as non-restorable observation, explicitly omit, or refuse. Never clear source claims during export. |
| Deleted/reserved allocations and tombstones | Current-only profile discloses omission/count; absence does not certify that a source-authority path is unallocated. Never claim source never-reuse state was preserved. |
| Prior graph/Issue versions, journal, Dolt branches/history | Direct exact versions selected by live top-level Link endpoints are retained separately in the internal inventory. Other versions, journals and branches remain omitted; no archive or recursive preservation closure is claimed. |
| WriterToken/local ordinals | Internal consistency inputs, not public revision, cursor, Scope epoch or durable snapshot handle. |

Storage JSON contains `version` and attribution actor/status/recordedAt. CLI
record shaping removes the alias and projects status to basis. BDP Read shaping
uses principal/basis and drops recordedAt. Export must pin a deliberate
projection with a field-fidelity test rather than inherit either lossy public
shape accidentally. A distinct `formatVersion` is proposed; CLI schemaVersion,
storage schema generation and BDP envelope version are separate concepts.

Determinism means unchanged source state produces identical bytes, including
observed attribution timestamps; it does not promise identical bytes for two
separately created stores. Use strict canonical JSON, then strict decode and
structural comparison with the detached acquired data before publication. Do
not use a generic decoder that rounds numbers. Do not fabricate export-time
attribution or history timestamps.

The storage reader is not gated on BDP Read serving. Installed qualification
means embedded mode and `dolt.mode=server` using a separately provisioned ordinary
SQL server with static credentials. `dolt.shared-server` and proxied-server
workspaces retain their current refusal. Current store validity, including
validated deleted mappings, is a precondition. The current acquisition has a
1,000-live-resource/16 MiB preview bound; additional comment/orphan acquisition
and encoded output must be bounded before allocation/publication. Breaches
refuse the entire export, naming the bound; no listing pagination or truncation.

CLI wiring must add early graph dispatch together with graph command admission,
validate an explicit graph flag allowlist, preserve the proxy registry and
refuse inherited filters/scrub flags that have no implemented graph semantics.
Graph mode/generation gates isolate execution, but shared CLI files can still
regress ordinary behavior: run ordinary-export regression tests. Avoid changes
to `graphstore/records.go` and `dependencies.go`, which PR86 also touches.

Proposed complete graph output includes Memory bodies and uses `0600` for files;
this differs deliberately from classic opt-in Memory export. Coordinate the
inclusion policy before exposing flags. Publish a fully validated buffer with
atomicfile after all storage cleanup. Atomic rename protects concurrent readers
from partial files; it does not promise directory-fsync/power-loss durability.
Ordinary failures clean owned temp files; crashes may leave temp residue. The
current graph wrapper's 30-second context covers reads; the final CLI design
must define cancellation through serialization and the publication boundary.

## Initial reader boundary and pending evidence

`ExportCurrentSnapshot` is an internal detached result, with no interchange
schema or CLI contract. It uses the existing transaction boundary and current
record/Type validation, adds native comments keyed by canonical relative Issue
path, exact persisted prefix and a deleted-allocation count, and returns zero
state on any acquisition or transaction/connection cleanup error. The caller
still owns final Store closure and must check it before publication.

Preflight charges all current comment columns, including orphan rows, in the
same 16 MiB budget as the graph acquisition; its comment count ceiling derives
from the existing 256-byte row charge. Deleted Issue final retained heads also
share that budget because mapping validation reads them. Original UTF-8 is
validated before a future encoder can replace bytes. Current wisp payloads,
relations, comments, labels and legacy KV Memories refuse by plane name.
Unselected history, journals, counters and unrelated configuration remain
outside the result. Direct selected exact endpoint values are a separate
internal inventory, as described below. Source-preservation tests compare logical rows; engine filesystem
bookkeeping is not claimed unchanged.

Pure tests and compile checks do not qualify an engine. Authored embedded/server
lifecycle, rollback, bounds and controlled-interleaving cases require source-bound execution. Callback stage faults do not stand in for actual
driver rollback/connection/rows-close fault injection. Authored test-owned
connector cases exercise those real database/sql cleanup paths on an empty
controlled workspace with the complete six-Type installation. They are neither
an engine nor proof of cleanup with records present; current-source execution
of these cases remains unrun. Installed acquisition, serializer/publication
checks and the final output contract also remain gates before exporter readiness.

## Detached publication boundary

`internal/graphoutput` accepts already validated detached bytes and caller-selected
file permissions. It does not acquire a Store, encode a format or authorize
content exposure. The caller must complete acquisition, validation and checked
Store cleanup before publication. Cancellation and write failures before final
publication preserve an existing destination; successful publication remains
success if cancellation arrives afterward. Stdout writes can leave partial
output, which cannot be retracted.

The file path uses the existing atomicfile temporary-write/fsync/rename boundary.
It does not promise directory-fsync or power-loss durability, nor recovery of
physical cleanup errors suppressed by that underlying API. These internal
mechanics do not activate public export flags, select file permissions or settle
Memory inclusion policy. Source-bound historical pure results and fresh source
council remain separate from author-owned current-source runtime qualification.


## Endpoint carrier adaptation, October 20

The internal reader consumes the shared endpoint carrier from PR133
(`202bbf215ce8cb7995bd237855a0669892c958e4`). Endpoint identity and optional
revision remain separate. Floating references remain floating. Complete
canonical Link bytes, including both pins, must agree with source-owned state;
ownership comes from `Snapshot.OwnedLinkTypes`, captured in the same transaction
by the published Type bridge `716af68f7e531108e4b1f09ae47cacefa18ee93a`. Missing
metadata refuses. Installed inventory Type descriptors deliberately omit some
selected constraints and are not an ownership fallback.

Export validates distinct exact endpoint addresses using the existing retained
reader inside the same rollback-only transaction. Shared retained-availability
checks refuse removed native Issue editions before decoding. The aggregate
export preflight charges each distinct historical endpoint address together
with current records and comments. A source/target UNION selects only referenced
addresses before joining retained payloads; repeated References decode and are
charged once. The acquisition retains the selected endpoint payloads in a
separate detached inventory and keeps their selected Type for historical
ownership checks. No new transaction or current-head fallback is used.
Owned-source pins remain deferred and refuse, including when several Links
share an already acquired source address.

`Snapshot.Records` still contains live current records only.
`ExportCurrentState.PinnedEndpoints` separately contains distinct exact Bead
endpoint versions selected by those current top-level Links. It is nonnil and
empty on successful pin-free acquisition; every failed acquisition or cleanup
returns the literal zero state. URI membership establishes current identity
coverage, not portable retained-version closure. References inside retained
owned Links are not traversed by this slice.
This internal reader therefore does not establish roundtrippable public export,
a wire format, CLI support or new-format import. Detached byte publication is
unchanged. Earlier qualification applies only to its recorded source; fresh
adapter evidence must identify the combined source and engines separately.


## Retained endpoint inventory boundary

This internal follow-on uses the existing `ReadVersion` projection as its value
oracle, not raw database rows. Native Issue dependencies are omitted and
metadata is exposed in the common record position by that existing reader;
retained owned bytes and native revision context remain those of the selected
version. Current comments are not added to a retained Issue version.

Each exact endpoint is hydrated independently, including a pin at the current
revision. A current native Issue can carry live lease/heartbeat observations
that changed without minting another retained revision, while its exact reader
returns the retained observations. Thus equal `(ID, Revision)` across
`Snapshot.Records` and `PinnedEndpoints` does not prove equal full values.
These inventories must not be concatenated into a supposedly unique current
snapshot or normalized to one another. This preserves two existing internal
read views; it does not select a public live-observation export policy.

Distinct selections deduplicate across both endpoint roles. Ordering uses the
shared code-unit comparator on canonical URI and then opaque revision. Floating
endpoints add no retained value; separate exact revisions remain separate.
Returned nested properties, metadata, owned bytes and attribution are detached
from current records and subsequent acquisitions.

The persisted-byte preflight and native-context charge remain aggregate.
Retaining all selected payloads increases simultaneous result retention even
though the same payloads were already read for validation; the 16 MiB
persisted-byte guard is not a Go heap bound. Selection is bounded by the
existing live-Link inventory. No all-history scan is introduced.

The follow-on starts from Janet's explicitly selected PR141 assembly
`e8c3b9516c60c5191d34408e8861917522bc7e35`. Its runtime/review evidence must
identify its own changed source and dependencies; earlier validation-only
reader proof remains historical. Public serialization, CLI output, recursive
closure, import and round-trip qualification remain separate work.
