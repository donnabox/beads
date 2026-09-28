# Issue scalar editing in the graph preview

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

Only inline text is admitted here. File input, stdin (including description `-`),
notes replacement, status, classification, labels, parent changes and metadata
flags refuse explicitly. Mixing these Issue flags with generic `--properties`
or source guards refuses. Legacy IDs and fuzzy or alias selectors are unavailable.
[Assignee editing](GRAPH_ISSUE_ASSIGNEE_PREVIEW.md) and
[append-only notes](GRAPH_ISSUE_NOTES_PREVIEW.md) extend this same guarded writer.
Notes replacement/clear remain held for the existing contributor safeguards in
[PR #5946](https://github.com/gastownhall/beads/pull/5946) and
[PR #6583](https://github.com/gastownhall/beads/pull/6583).
Existing legacy Issue workspaces and their update behavior remain unchanged.

The adapter uses the existing authoritative Issue writer and Jim's retained
Issue recorder in one checked transaction with the graph mapping. It does not
add an Issue payload table, schema, recorder or separate database. An actual
edit records exactly one new Issue snapshot and opaque graph version containing
the complete owned blocking Dependencies. Dependencies, their targets and
informational Links remain unchanged. Editing a closed Issue's admitted scalar fields does not
reopen it or clear its close time. Readonly and migration-freeze policy still
refuse writes before storage is opened.

Saved versions remain available through exact `show --version` and experimental
`compare`. This slice does not provide ordered History, native atomic change
context, public BDP History, generic Issue property replacement, or the complete
Issue workflow. Existing attribution timestamps are not claimed as qualified
native History commit timestamps. The outstanding storage-owner binding gate
remains open.

`status --graph --json` advertises `issueTextUpdate: true`,
`issuePriorityUpdate: true`, `issueEstimateUpdate: true`, `issueReferenceUpdate: true`,
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

## Estimate minutes

The existing `--estimate` (`-e`) flag works with `update` on canonical graph
Issues. Graph `create` also accepts it for the initial Issue version:

```bash
bd update beads/task -e 0 --if-revision "$revision" --json
bd update beads/task --estimate 45 --title 'Sized task' --unconditional --json
```

Omitting the flag preserves the estimate, including an absent estimate. An
explicit zero is stored and read back as zero; it does not clear the field to
null. Negative values are rejected. Parsing and SQL integer storage reuse the existing
Issue machinery. The adapter also rejects values above the current signed SQL INT
maximum of 2,147,483,647 before opening a transaction, so strict and coercing
engines produce the same validation refusal. If storage coerces the estimate to another
value, the graph transaction refuses and rolls back all accompanying edits. The
check does not alter ordinary Issue commands or widen the schema. This is an
estimate in minutes, with no scheduling effect. The preview does not add a null-clear option.

Estimates can share one transaction and retained version with the other admitted
scalar edits and notes append. The graph revision guard covers owned Dependencies
and runs before no-op detection. A same-value estimate is unchanged, while a
stale same-value request refuses. Estimate-only edits preserve status, assignment,
claim lease and owned Links. They do not renew a claim or repair external drift.
`status --graph` reports the provisional `issueEstimateUpdate` capability;
`issueWorkflows` remains false.


Reproduce the estimate-specific installed sequence against a disposable server:

```sh
python3 scripts/graph-issue-estimate-smoke.py \
  --bd /absolute/disposable/bin/bd --backend both --server-port PORT \
  --server-root /absolute/disposable/dolt-data \
  --output-dir /absolute/new/estimate-evidence --total-timeout 300
```

The real-store `TestIssueEstimateLifecycle` also checks the signed INT maximum;
both engine cases are required by the exact-source Linux workflow and verifier.

## External and spec references

The existing `update --external-ref` and `--spec-id` flags accept literal values
on canonical graph Issues, under the same observed revision or explicit
unconditional policy:

```bash
bd update beads/task --external-ref 'tracker/42' --spec-id 'docs/design' --if-revision "$revision" --json
bd update beads/task --external-ref '' --spec-id '' --unconditional --json
```

Omission preserves the field. An empty external reference becomes SQL NULL;
an empty spec ID becomes an empty string. Both cleared properties are omitted
from the existing Issue JSON representation. Values are not trimmed, normalized,
interpreted as file input, fetched, or converted into graph identity. Multiple
Issues may have the same reference. The existing VARCHAR columns permit up to
255 Unicode code points for the external reference and 1,024 for the spec ID.
Invalid UTF-8 and values outside those storage bounds refuse before SQL; an
exact writeback check prevents accepted values from being silently changed.

Reference pairs and other admitted scalar edits share one transaction and
retained revision. Same-value and repeated-clear requests are no-ops after guard
validation; a stale guard still refuses. Closed state, assignment, claim lease,
canonical identity and owned Links are preserved. Reference editing does not
renew claims, synchronize trackers or enable mixed claim edits. Graph create
also accepts these fields in the initial version. `status --graph` reports provisional
`issueReferenceUpdate`; full `issueWorkflows` remains false.

The existing current-read budget is unchanged. References can be shortened or
cleared through the checked writer if the workspace exceeds that budget; a
caller still needs its saved accepted revision or the existing explicit
unconditional option. Mixing references with notes append retains append's
existing post-write read-budget check. This adds no general repair interface.

Reproduce the installed sequence against a disposable ordinary server:

```sh
python3 scripts/graph-issue-references-smoke.py \
  --bd /absolute/disposable/bin/bd --backend both --server-port PORT \
  --server-root /absolute/disposable/dolt-data \
  --output-dir /absolute/new/reference-evidence --total-timeout 300
```

It records normal initialization, separate and combined set/replace/clear,
complete current and exact saved records, and refusals. Storage tests additionally
check SQL NULL and the exact ASCII/multibyte column boundaries on both engines.
The independent public BDP client observes CLI-authored assignment and clearing,
preserving complete properties and ownership with changed ETags. HTTP writes
and public Memory History remain unavailable. Execution receipts and qualification
status belong to the fork delivery plan; this recipe is not a qualification claim.


## Complete initial fields

A single graph `create` can set the existing inline design, acceptance,
assignee, estimate and external/spec reference fields in its first retained
version. For example, in a normally initialized disposable graph workspace:

```sh
bd create 'Deliver the plan' --id beads/task \
  --design 'Reuse the current storage writer' \
  --acceptance 'Fresh-process reads agree' \
  --assignee donna --estimate 45 \
  --external-ref tracker/task --spec-id specs/plan --json
bd show beads/task --json
```

This reuses one ordinary Issue create transaction and its initial recorder.
There is no follow-up update. Assignment leaves the Issue open and grants no
claim lease. Design and acceptance text preserve whitespace, Unicode and the
literal `-`; file sources remain unavailable. Omitting the estimate leaves it
absent; an explicit zero is present. Empty CLI external references become SQL
NULL, while an empty spec ID remains an empty string. The internal Go
CreateRequest preserves its existing nil versus nonnil-empty external-reference
pointer distinction. References remain literal properties, without tracker I/O.

All six fields obey their existing representation limits. The adapter checks
exact hydrated values before publishing the graph mapping, rolling back all
create effects if storage changes an accepted value. Later guarded edits retain
the complete first version. Creation labels still use their existing events;
one retained version does not mean one total event for a labeled Issue.

`status --graph` reports provisional `issueCreateFields: true`;
`issueWorkflows` remains false. Status defaults, initial notes, due/defer,
composite relationships, metadata, aliases and file inputs are outside this
slice. Native Owner and CreatedBy properties remain unset as in the earlier graph
creator; the create actor still supplies retained attribution. This is partial
ordinary-create compatibility. Global preview read limits are unchanged; very
large text may exceed them even when its existing LONGTEXT column accepts it.

Reproduce both installed backends with
`scripts/graph-issue-create-fields-smoke.py`, using the same `--bd`, `--backend`,
`--server-port`, `--server-root` and fresh `--output-dir` options as the reference
recipe above. Storage tests separately verify a single initial retained snapshot,
SQL presence, native bounds, failure rollback and linked-state preservation.
The independent public BDP client observes all six fields after one CLI create;
HTTP writes and public Memory History remain unavailable.
