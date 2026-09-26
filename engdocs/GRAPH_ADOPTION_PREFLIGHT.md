# Read-only graph adoption preflight

This is a bounded inventory for the delivery plan's early adoption-feasibility gate.
It reports named exceptions from frozen legacy table exports. It does not open a
Beads workspace, run migrations, allocate canonical IDs, change schema, establish
a Scope, or certify adoption. `adoptionReady` is always `false`.

```sh
python3 scripts/graph-adoption-preflight.py /absolute/path/frozen-export.json \
  > /absolute/path/new-preflight-report.json
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts \
  -p test_graph_adoption_preflight.py -v
```

Exit 0 means a report was produced, including its blocking exceptions. Exit 2
means invalid or oversized input; stderr contains `invalid_export` and stdout
contains no partial report. The tool reads one file and writes only stdout/stderr.
Use a new report destination, never the input path.

## Input contract

The input is a UTF-8 JSON object, at most 32 MiB and 100,000 exported rows total.
Duplicate object members, non-JSON numbers, numbers outside the finite float
range, non-UTF-8 byte encodings, malformed table arrays, and excess size refuse. No filenames, URLs, commands, or database routes inside the bundle
are followed or executed.

```json
{
  "schemaVersion": 1,
  "complete": true,
  "provenance": {
    "kind": "release-authored-surrogate",
    "release": "v0.62.0",
    "frozen": true,
    "head": "observed Dolt commit identity",
    "workingSetSha256": "separately recorded frozen working-set fingerprint"
  },
  "tableInventory": ["issues", "dependencies", "config"],
  "tables": {
    "issues": [{"id": "old-1", "issue_type": "task", "status": "open"}],
    "dependencies": [],
    "config": [{"key": "kv.memory.plan", "value": ""}]
  },
  "metadata": {"backend": "dolt"},
  "artifacts": {
    "database-backup": {"present": false}
  }
}
```

This small example intentionally lacks history and backup evidence and receives
exceptions. Production callers must explicitly label provenance as `synthetic`,
`release-authored-surrogate`, or `permitted-workspace`. These labels and the
`frozen`/`complete` assertions are exporter declarations, not independently
verified facts. Even `permitted-workspace` does not make this report a permission
check or an adoption qualification.

The exporter supplies every table name observed, and complete row arrays for
exported tables. Omit an unavailable table's array and keep its name in the
inventory; the report names `TABLE_NOT_EXPORTED` instead of treating it as empty.
Missing legacy tables and unknown tables are separately reported. Counts for
unexported tables are `null`, not zero. `parsedDependencies` counts only rows
whose endpoint shape could be interpreted; malformed rows remain in the raw
Dependency row counts and named exceptions. There is no
truncated-success or pagination contract.

Useful exports include `issues`, `wisps`, `dependencies`, `wisp_dependencies`,
`config`, `labels`, `comments`, `issue_versions`, `store_epoch`, schema metadata,
`dolt_ignore`, `dolt_status`, and any `graph_preview_*` tables. The exact installed
schema determines which exist. Include system/ignored-table evidence separately
when it is not returned by ordinary `SHOW TABLES`. Preserve the raw exports and
schema inventory alongside the bundle. The tool does not decode SQL binary
encodings or reinterpret arbitrary historic schemas.

`metadata` is a sanitized inline copy of relevant `.beads/metadata.json` fields.
`artifacts` entries have `present` and an optional lowercase SHA-256 `sha256`.
The required inventory names are `metadata.json`, `config.yaml`, `database-backup`,
`working-set`, and `table-inventory`. A fingerprint is an inventory assertion;
the report does not access, validate, or restore that artifact. Do not put
credentials in the bundle. Output omits Memory bodies, Issue content/metadata
values, and configuration values; it includes exact identities, external
references, and observed graph bindings, so it may still contain sensitive data.

## Interpretation and decisions left open

- **Identity.** Issues, wisps, and exact `kv.memory.` keys are inventoried.
  Empty Memory bodies and a real `schema_version` key remain data. Conservative
  `beads/<identity>` candidates expose collisions without allocating paths or
  approving the alias contract. Unusual keys remain exact and require review.
  Legacy Dependency identity is its source/target pair, excluding Type. The
  report preserves existing IDs and identifies duplicate pairs; rows lacking an
  ID use a report row locator, not a fabricated Link ID. An eventual importer
  must reuse `internal/storage/depid.New` and `graphops` validation rather than
  adopting this report's candidate strings or reimplementing ID derivation.
- **External blockers.** Both historical `depends_on_id` and the typed target
  columns are read. `external:` values, URLs, explicit opaque external targets,
  missing local targets, and wisp endpoints remain distinct exceptions. A
  differently prefixed missing target is unresolved, not presumed external or
  satisfied. The current blocking descriptor admits local Issues only. A future
  domain adapter needs an explicitly reviewed external-reference representation
  that preserves opaque locators and keeps unresolved blockers unsatisfied.
  Creating placeholder local Issues would violate that distinction. This tool
  proposes no durable Type change and performs no remote lookup.
- **Fields and history.** Custom classifications/statuses, populated unsupported
  Issue/Dependency fields, and aggregate labels/comments require adoption
  adapters. Retained-state presence, JSON shape, removal, epoch, and attribution
  gaps are named. A default `current_revision = 1` does not imply historical
  backfill. This is not a full retained-content equality check, writer census,
  historical traversal, or restoration proof. No history or actor is invented.
- **Backup and continuity.** SQL data alone cannot prove graph authority.
  Scope/authority/workspace bindings, allocation and tombstones, Type descriptors,
  retained history, working-set state, and ignored/clone-local tables need a
  coherent backup. `_project_id` does not establish graph Scope continuity.
  Existing `runBackupExport` uses a commit watermark; its inclusion of graph
  working-set writes still needs qualification. Artifact presence never replaces
  a fresh-machine restore test or exclusion of old writers.

## Verified release-authored surrogate

On September 25, 2026, the mechanics were exercised against a disposable source
created by the actual v0.62.0 release binary. The archive was checked against
`scripts/migration-test/lib/versions.sh`'s pinned Darwin/arm64 SHA-256:

`4f441c616240c7063a2257ac26138d06eae060a961f029c81855928eae6ae1c0`

The source-generation pattern came from
`scripts/migration-test/historical-dolt-upgrade-test.sh:create_historical_fixture`:
normal initialization on ordinary Dolt 2.1.8, three Issues, one blocking
Dependency, a label, a comment, one closed Issue, and one legacy Memory. The
historical upgrade writer was never run, and the candidate CLI never opened the
source. The test server was stopped and reaped. A separate quiescent database
copy supplied SELECT exports; strict before/after source file fingerprints
matched. The report itself read the resulting frozen JSON only.

The observed report contained 3 Issues, 1 Dependency, 1 legacy Memory, and no
`issue_versions` table among 23 exported tables (`retainedRows: null`; retained
history was unavailable). It identified all three
current-history gaps, missing epoch/graph binding, legacy Memory conversion,
aggregate import, unclassified table, populated field, and working-set backup
exceptions. No backup restore was performed. External reference encodings, wisps,
collisions, malformed history, and input failures were covered by synthetic
focused tests, not by this historical source.

Local reproducibility evidence is retained under
`/Users/dbox/.local/state/janet/bdp-read-20260925/adoption/`: release provenance,
source-generation and export scripts, per-command receipts, the frozen export,
source fingerprints, `report.json`, and `python-tests.log`. These are local
receipts, not a published distributable dataset.

**The representative adoption gate remains incomplete.** No permitted existing
production workspace was supplied or inspected. This small release-authored
surrogate proves report mechanics on an authentic historical layout. It does
not measure real workspace exception rates, establish an import/rollback route,
resolve external blocking semantics, or qualify backup continuity. The next
bounded step is a permitted sanitized frozen workspace export using this
contract, followed by explicit disposition of its named exceptions.
