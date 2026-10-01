package compare

// The tables a view covers. The comparator needs a decision about every column
// of each of them, and the drift tests hold these lists to the schema of a
// freshly initialized store: a migration that adds a column fails the build
// until someone decides whether it is compared or a stamp.
const (
	tableIssues       = "issues"
	tableDependencies = "dependencies"
)

// comparedColumns are the columns of each replayed table that decide whether a
// replay matched. The rule is to compare every column except the stamps, so a
// column the lists have never heard of is compared too; these lists exist so the
// drift test can make a person look at each new column once.
var comparedColumns = map[string][]string{
	tableIssues: {
		"id", "title", "description", "design", "acceptance_criteria", "notes",
		"status", "priority", "issue_type", "assignee", "estimated_minutes",
		"created_by", "owner", "external_ref", "spec_id", "compaction_level",
		"original_size", "sender", "ephemeral", "wisp_type", "pinned",
		"is_template", "mol_type", "work_type", "source_system", "metadata",
		"source_repo", "close_reason", "event_kind", "actor", "target", "payload",
		"await_type", "await_id", "timeout_ns", "waiters", "hook_bead",
		"role_bead", "agent_state", "role_type", "rig",
		// Set by the user, and replayable exactly, so they are compared.
		"due_at", "defer_until",
		"no_history", "storage_class",
	},
	tableDependencies: {
		"issue_id", "type", "metadata", "thread_id",
		"depends_on_issue_id", "depends_on_wisp_id", "depends_on_external",
	},
}

// stampColumns are the columns that are never compared: values bd fills from the
// wall clock, derives, or takes from the writer's environment rather than from
// the input a replay is given. Each carries the reason, because a stamp is a
// decision not to look and should say why.
var stampColumns = map[string]map[string]string{
	tableIssues: {
		"created_at":          "wall clock at insert",
		"updated_at":          "wall clock at every write",
		"closed_at":           "wall clock at close",
		"started_at":          "wall clock when work started",
		"last_activity":       "wall clock of the last agent activity",
		"compacted_at":        "wall clock at compaction",
		"compacted_at_commit": "a commit hash from the compacting clone's own history, which differs between clones by construction",
		"content_hash":        "derived from the other columns on every write",
		"row_lock":            "a write counter the store advances on every write",
		"is_blocked":          "derived from the edge set, which the view compares",
		"closed_by_session":   "the session that wrote the close, not a property of the issue",
		"current_revision":    "a local ordinal the store advances when it records a version, not a property of the issue",
	},
	tableDependencies: {
		"id":         "a surrogate key generated when the edge is written; the edge is identified by its issue, its target and its type",
		"created_at": "wall clock at insert",
		"created_by": "the identity that wrote the edge, not a property of the edge",
	},
}

// jsonColumns are the JSON-typed columns of each replayed table. Only these are
// parsed as documents; every other column is compared as its exact text. A drift
// test holds the set to the schema, because a JSON column compared as text would
// depend on how each Dolt release spells its numbers.
var jsonColumns = map[string][]string{
	tableIssues:       {"metadata"},
	tableDependencies: {"metadata"},
}

func isStamp(table, column string) bool {
	_, ok := stampColumns[table][column]
	return ok
}

func isJSONColumn(table, column string) bool {
	for _, c := range jsonColumns[table] {
		if c == column {
			return true
		}
	}
	return false
}
