# Explicit Issue listing in graph workspaces

The graph preview supports `bd list --flat` and `bd list --format records-json`.
It returns Issues only, including open Issues blocked by Dependencies. Use `ready`
for scheduling eligibility; listing does not claim that displayed work is ready.

```sh
bd list --flat
bd list --format records-json --limit 5
bd list --format records-json --all --sort title
bd list --flat --status closed --type task
```

This is an explicit preview subset of the existing Issue workflow, pending CLI
review. Bare tree output, `--json`, and `--format json` refuse in a graph workspace.
The existing Issue JSON contract is not reinterpreted based on workspace format.
Legacy workspaces retain their normal list path; `records-json` is graph-only.
The generic `list --kind bead|link` proposal is a separate, unfinished surface.

## Selection and limits

The adapter reuses `gatherListInput`, `workapi` configuration/filter/page policy,
and the existing Issue SQL query. Supported filters are status (or state), type,
title/title-contains, priority and priority range, label/label-any/exclude-label,
and pinned/no-pinned. It supports all, limit, reverse, and sorting by priority,
created, updated, title, status or type. Other changed flags refuse, including
explicit false/empty settings for unsupported features.

Omitted filters retain existing policy: configured done/frozen status categories,
pinned defaults, templates/gates/infra visibility, and directory labels. `--all`
retains its Issue-list meaning; it is not generic BDP page draining. Explicit
limit wins over all/configuration; otherwise all, configured list.limit, piped
unlimited, terminal agent default of 20, then the normal default of 50 apply in that order.
A deliberate page limit returns `hasMore`, never an invented total. Repeating a
query is a new read, not a continuation of the previous snapshot.

Ordering comes from the existing Issue query and page policy, including its
backing-ID tie break. It is not generic canonical Resource ordering. Explicit
ID sorting, IDs/aliases, readiness, parent walks, tree/pretty/dependency renderers,
watching, ephemeral/infra widening, skipped fields/count projections, routing,
offsets and keyset pagination remain unavailable here. Comma-separated status
sets retain shared semantics; repeated status/state/type flags and simultaneous
status/state refuse before gathering. Supplied-empty labels also refuse before
directory defaults can erase the distinction. These graph admission restrictions
do not implement or supersede separately owned ordinary parser fixes.

## Output

Flat output quotes canonical ID, status and title, shows priority, and reports
whether more matching Issues exist. It deliberately has no dependency tree,
blocker annotation, aggregate counts or pager. Quiet suppresses flat output.

Records JSON uses the existing graph preview envelope around this result:

```json
{
  "items": ["<complete canonical Issue records>"],
  "hasMore": true
}
```

The item placeholder represents a complete object, including canonical identity,
Type, version/revision, Issue properties and owned blocking Links. The backing
Issue ID remains a property; it is not rewritten to imitate a legacy JSON row.
`items` is an empty array when nothing matches. `--format records-json` selects
structured output even under ambient JSON configuration or alongside `--flat`;
the explicit format takes precedence over the human flat rendering. Explicit `--json` still
refuses. Flat under ambient JSON also refuses until legacy JSON compatibility is
implemented. The provisional page shape and eventual default rendering require
review before promotion to durable contracts.

## Authority and failure boundary

Configuration resolution, filtering, probe-row checking and canonical-record
mapping occur in one checked read transaction. No legacy store opens, expired
defers are not awakened, and no graph coordination/version/event write occurs.
Returned records and the over-fetched probe are validated before trimming; a bad
probe cannot be concealed by `hasMore`. Output begins after successful operation
and store cleanup. Failures return no partial page.

This preview retains the whole-workspace 16 MiB current-read acquisition and 1000
live-Resource bounds, plus DB configuration acquisition bounded at 64 KiB and 256 rows.
The resolved configuration, including YAML fallbacks, is bounded at 64 KiB and
256 entries. Database query, scan and iteration failures refuse instead of
selecting fallback policy. A malformed nonempty stored `status.custom` value
retains the existing shared parse behavior: no custom statuses, rather than
refusal or YAML fallback. YAML comes from the already-initialized CLI workspace
configuration; it is not part of the database transaction or historical configuration. The resolved
bound does not cap the frontend's earlier YAML file parsing and allocation.
Direct multi-workspace server configuration is not supplied by this adapter. These are
operational admission limits: unrelated Memory/Link growth can refuse a small
Issue query. Lightweight catalog/Type checks do not hydrate every unreturned
Memory body; complete retained-record checking applies to selected/probe Issues.
The serialized output cap is 16 MiB. Exceeding a bound refuses rather than silently
truncating or advertising a complete result. Existing `BEADS_MAX_ROWS` policy also
applies through the shared query. Its inherited malformed-value warning/ignore
behavior is unchanged.

## Verification scope

The installed harness uses normal initialization and CLI authoring on embedded
and ordinary shared-server Dolt, mixed Memory/Issue/Links, lifecycle changes,
filters, deliberate page limits, exact retained reads and admission failures.
Real-store tests exercise config policy, authority/corruption, cancellation,
concurrent snapshot behavior and no-write assertions. Some custom-status/infra
configuration cannot yet be authored through graph CLI commands; those policy
checks use existing configuration APIs and are not claimed as installed setup.
Template and expired-defer authoring are unavailable in the preview; their policy
has shared-builder/source evidence, not an installed graph workflow claim.
Different-database provisioning remains serialized on Dolt 2.1.8.

This slice does not complete Issue workflows, generic BDP CLI listing, full
Memory, native/public History, or aliases. Existing contributor tree, ordinary
JSON truncation and repeated-filter fixes keep their ownership and provenance.
