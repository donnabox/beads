# Issue text and priority editing in the graph preview

A normally initialized graph workspace accepts four familiar Issue text flags:
`--title`, `--description`, `--design` and `--acceptance`, plus `--priority`. Use the
canonical `beads/PATH` or its exact local Scope URL, and supply either the
observed graph revision or explicit `--unconditional`.

```sh
bd show beads/task --json
bd update beads/task --title 'Revised task' --description 'The revised plan' \
  --if-revision OBSERVED_VERSION --json
bd show beads/task --json
bd show beads/task --version OBSERVED_VERSION --json
```

This is a reversible preview adapter for familiar Issue fields, not generic
JSON patch support or a settled CLI contract. The authoring proposal remains
[upstream #6703](https://github.com/gastownhall/beads/issues/6703).

Each supplied field replaces that field; omitted fields remain unchanged.
Explicit empty description, design and acceptance values clear those
fields. Titles retain the Issue CLI's
trimming and nonempty, 500-byte validation. Description aliases `--body` and
`--message` work; multiple aliases must agree. Other text is preserved verbatim
and must be UTF-8. A matching no-op preserves the entire Resource and creates
no retained version. Stale guards refuse even when the proposed fields match
the current values.

`--design=-` remains literal text, as in the existing Issue CLI.

Priority uses the same parser as ordinary Issue updates: values 0–4 or P0–P4
(case-insensitive P prefix), including that parser's permissive integer-prefix
behavior. This adapter does not introduce stricter lexical validation. An omitted
priority preserves the old value; explicit P0 sets zero. Priority can be combined
atomically with the supported text fields, under the same complete graph guard.
Reprioritizing blocked work does not remove its Dependencies or make it ready.

```sh
bd list --format records-json --all --limit 0 --sort priority
# Use the revision returned by the selected Issue record.
bd update beads/task --priority P0 --title 'Urgent task' \
  --if-revision OBSERVED_VERSION --json
bd list --format records-json --priority 0 --all --limit 0
bd show beads/task --version OBSERVED_VERSION --json
```

Only inline text is admitted here. File input, stdin (including description `-`), notes, append-notes, status, classification, labels, assignment,
parent changes and metadata flags refuse explicitly. Mixing these Issue flags
with generic `--properties` or source guards refuses. Legacy IDs and fuzzy or
alias selectors are unavailable in this graph route. Notes editing is intentionally deferred to avoid a competing contract with
[the existing contributor safeguard PR #5946](https://github.com/gastownhall/beads/pull/5946).
Existing legacy Issue
workspaces and their update behavior remain unchanged.

The adapter uses the existing authoritative Issue writer and Jim's retained
Issue recorder in one checked transaction with the graph mapping. It does not
add an Issue payload table, schema, recorder or separate database. An actual
edit records exactly one new Issue snapshot and opaque graph version containing
the complete owned blocking Dependencies. Dependencies, their targets and
informational Links remain unchanged. Editing a closed Issue's text or priority does not
reopen it or clear its close time. Readonly and migration-freeze policy still
refuse writes before storage is opened.

Saved versions remain available through exact `show --version` and experimental
`compare`. This slice does not provide ordered History, native atomic change
context, public BDP History, generic Issue property replacement, or the complete
Issue workflow. Existing attribution timestamps are not claimed as qualified
native History commit timestamps. The outstanding storage-owner binding gate
remains open.

`status --graph --json` advertises `issueTextUpdate: true`,
`issuePriorityUpdate: true`,
`issueTextFileInput: false`, `issueTextStdinInput: false`, and retains
`issueWorkflows: false`.

## Installed-command proof

Install with `make install-force INSTALL_DIR=/absolute/disposable/bin`. Start
ordinary Dolt 2.1.8 with its own disposable data directory, then run:

```sh
python3 scripts/graph-issue-edit-smoke.py \
  --bd /absolute/disposable/bin/bd --backend both --server-port PORT \
  --output-dir /absolute/new/evidence-directory
```

The harness uses normal initialization and installed CLI writes, then reads
the results from fresh processes on embedded and shared-server Dolt. No manual
schema seeding, SQL fixture or mock is used. It preserves command receipts and
failed evidence; the caller owns the server. Provision test databases serially
on Dolt 2.1.8. Separate real-store tests exercise transactional rollback,
authority checks and concurrent writers. The harness alone is not full graph
delivery or public History qualification.

The separate `scripts/graph-bdp-read-smoke.py` proof also edits an Issue using
the installed CLI while the ordinary-server Read service is running. The
unchanged public BDP client verifies the new Resource revision, all four fields
through Resource/properties/inventory reads, unchanged owned Dependencies and
incident Links, and the unchanged target Issue. A prior ETag yields the fresh
record. This exercises CLI writing followed by BDP Read; HTTP writes remain
unavailable.

The priority workflow has a separate installed harness:

```sh
python3 scripts/graph-issue-priority-smoke.py \
  --bd /absolute/disposable/bin/bd --backend both --server-port PORT \
  --output-dir /absolute/new/priority-evidence
```

It selects an Issue from a complete list record, edits priority (alone and with
text), observes filtering/order, and reopens the exact retained predecessor.
It also checks blocked/closed state, complete owned Dependencies and targets,
no-op/stale guards and refusal paths. This is a command recipe; execution status
and exact artifacts are reported separately in the fork delivery plan.

Integration note: this pinned base's `ExecuteUpdate` does not mint its final
retained version, so the graph adapter invokes the existing recorder once.
Upstream PR6650 moved that responsibility into `ExecuteUpdate`. Integrating a
base that contains it requires reconciling these calls and requalifying exactly
one complete version per accepted edit. The priority adapter does not change
recorder ownership or import an upstream stack speculatively.
