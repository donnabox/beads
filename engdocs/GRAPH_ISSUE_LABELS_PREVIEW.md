# Guarded Issue label replacement preview

This draft is stacked on PR44's shared label writer correction. It remains held
for that PR's ordinary RowVersion and whole-Issue-table staging decisions. The
qualified integration candidate is unchanged. Reusing the existing `--set-labels`
flag and advertising `issueLabelReplacement` are provisional preview choices,
not settled upstream CLI contracts.

After normal graph initialization, create an Issue, read its revision, and use:

```sh
bd update beads/work --set-labels 'review,shipping' --if-revision '<revision>' --json
bd show beads/work --json
bd list --format records-json --label shipping --all --limit 0
bd show beads/work --version '<previous version>' --json
```

Use the actual revision/version returned by a read. `--set-labels=` explicitly
clears the set; omitting the flag preserves it. The existing repeated/CSV flag
parser and label normalization trim surrounding whitespace, drop empty entries
and remove duplicates. Case and accent distinctions are preserved. A quoted CSV
field can contain a comma. Labels retain the existing 255-character bound.
Complete replacement can accompany inline text and priority edits in one checked
transaction and one retained graph version. Explicit `--unconditional` is also
supported by the existing Issue update policy; it does not promise the Memory
predecessor-disclosure result shape.

The revision guard covers the complete Issue record and its owned Dependencies.
A stale guard refuses even if the requested set equals the current labels.
Equivalent sets preserve the current record, attribution, events, coordination
state and retained-version count. Clearing an already empty set is also a no-op.
Failed mixed edits roll back the Issue row and authoritative label rows together.
This slice preserves the Issue's other properties, owned Dependencies, unrelated
Memory and informational Links, and existing readiness behavior.

The graph adapter delegates replacement to `IssuePatch.Labels.Replace` and the
existing `ExecuteUpdate` writer. A shared label-set planner keeps early no-op
admission consistent with the ordinary writer; actual affected-row results still
determine whether a mutation happened. Input is copied before transaction work.
The pinned retained recorder is invoked once for the complete accepted edit.
Future integration with upstream PR6650 must reconcile recorder ownership.

Add/remove shorthand, label vocabulary and exclusive namespace policy remain
outside this slice. Upstream PR5793's complete contribution remains in the base;
PR4757 and PR6008 keep their existing ownership. This draft does not merge or
replace those contributions. Full Issue workflows, public ordered History,
HTTP writes, aliases and recovery remain unavailable. Native History atomic
commit-stamp binding is still unresolved.

`scripts/graph-issue-labels-smoke.py` exercises the installed CLI from normal
initialization on embedded and ordinary shared-server Dolt. It authors all Beads
and Links through commands, captures each subprocess and checks current records,
exact prior records, comparisons, filters and refusal outcomes. No-op internals,
rollback and overlapping writers are covered separately by real-engine Go tests;
CLI record equality alone does not establish those properties. Different-database
provisioning on Dolt2.1.8 remains serialized. Local results, exact-source Linux
qualification and substantive human approval are distinct gates.
